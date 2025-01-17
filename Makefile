
ifdef ACCTEST_PACKAGES
		# remove newlines and blanks coming from GH Actions
    ACCTEST_PACKAGES := $(strip $(subst $(newline),, $(ACCTEST_PACKAGES)))
else
    ACCTEST_PACKAGES := "./..."
endif

ACCTEST_TIMEOUT?=300m
PARALLEL_GO_TEST?=50

BINARY_NAME=terraform-provider-mongodbatlas
DESTINATION=./bin/$(BINARY_NAME)

GOFLAGS=-mod=vendor
GITTAG=$(shell git describe --always --tags)
VERSION=$(GITTAG:v%=%)
LINKER_FLAGS=-s -w -X 'github.com/mongodb/terraform-provider-mongodbatlas/version.ProviderVersion=${VERSION}'

GOLANGCI_VERSION=v1.62.2 # Also update golangci-lint GH action in code-health.yml when updating this version

export PATH := $(shell go env GOPATH)/bin:$(PATH)
export SHELL := env PATH=$(PATH) /bin/bash

default: build

.PHONY: help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' | sort

.PHONY: build
build: fmt fmtcheck ## Generate the binary in ./bin
	go build -ldflags "$(LINKER_FLAGS)" -o $(DESTINATION)

.PHONY: clean-atlas-org
clean-atlas-org: ## Run a test to clean all projects and pending resources in an Atlas org, supports export DRY_RUN=false (default=true)
	@$(eval export MONGODB_ATLAS_CLEAN_ORG?=true)
	@$(eval export DRY_RUN?=true)
	go test -count=1 'github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/clean' -timeout 3600s -parallel=250 -run 'TestCleanProjectAndClusters' -v -ldflags="$(LINKER_FLAGS)"

.PHONY: test
test: fmtcheck ## Run unit tests
	go test ./... -timeout=30s -parallel=4 -race

.PHONY: testmact
testmact: ## Run MacT tests (mocked acc tests)
	@$(eval ACCTEST_REGEX_RUN?=^TestAccMockable)
	@$(eval export HTTP_MOCKER_REPLAY?=true)
	@$(eval export HTTP_MOCKER_CAPTURE?=false)
	@$(eval export MONGODB_ATLAS_ORG_ID?=111111111111111111111111)
	@$(eval export MONGODB_ATLAS_PROJECT_ID?=111111111111111111111111)
	@$(eval export MONGODB_ATLAS_CLUSTER_NAME?=mocked-cluster)
	@$(eval export MONGODB_ATLAS_ADVANCED_CLUSTER_V2_SCHEMA?=true)
	@if [ "$(ACCTEST_PACKAGES)" = "./..." ]; then \
		echo "Error: ACCTEST_PACKAGES must be explicitly set for testmact target, './...' is not allowed"; \
		exit 1; \
	fi
	TF_ACC=1 go test $(ACCTEST_PACKAGES) -run '$(ACCTEST_REGEX_RUN)' -v -parallel $(PARALLEL_GO_TEST) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -ldflags="$(LINKER_FLAGS)"

.PHONY: testmact-capture
testmact-capture: ## Capture HTTP traffic for MacT tests
	@$(eval export ACCTEST_REGEX_RUN?=^TestAccMockable)
	@$(eval export HTTP_MOCKER_REPLAY?=false)
	@$(eval export HTTP_MOCKER_CAPTURE?=true)
	@if [ "$(ACCTEST_PACKAGES)" = "./..." ]; then \
		echo "Error: ACCTEST_PACKAGES must be explicitly set for testmact-capture target, './...' is not allowed"; \
		exit 1; \
	fi
	TF_ACC=1 go test $(ACCTEST_PACKAGES) -run '$(ACCTEST_REGEX_RUN)' -v -parallel $(PARALLEL_GO_TEST) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -ldflags="$(LINKER_FLAGS)"

.PHONY: testacc
testacc: fmtcheck ## Run acc & mig tests (acceptance & migration tests)
	@$(eval ACCTEST_REGEX_RUN?=^TestAcc)
	TF_ACC=1 go test $(ACCTEST_PACKAGES) -run '$(ACCTEST_REGEX_RUN)' -v -parallel $(PARALLEL_GO_TEST) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -ldflags="$(LINKER_FLAGS)"

.PHONY: testaccgov
testaccgov: fmtcheck ## Run Government cloud-provider acc & mig tests
	TF_ACC=1 go test ./... -run 'TestAccProjectRSGovProject_CreateWithProjectOwner' -v -parallel 1 "$(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -ldflags=$(LINKER_FLAGS) "

.PHONY: fmt
fmt: ## Format Go code
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w .

.PHONY: fmtcheck
fmtcheck: ## Currently required by tf-deploy compile
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: lint-fix
lint-fix: ## Fix Go linter issues
	@echo "==> Fixing linters errors..."
	fieldalignment -json -fix ./...
	golangci-lint run --fix

.PHONY: lint
lint:
	@echo "==> Checking source code against linters..."
	golangci-lint run

.PHONY: tools
tools:  ## Install the dev tools (dependencies)
	@echo "==> Installing dependencies..."
	go telemetry off # disable sending telemetry data, more info: https://go.dev/doc/telemetry
	go install github.com/icholy/gomajor@latest
	go install github.com/terraform-linters/tflint@v0.52.0
	go install github.com/rhysd/actionlint/cmd/actionlint@latest
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	go install github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@latest
	go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@latest
	go install github.com/hashicorp/go-changelog/cmd/changelog-build@latest
	go install github.com/hashicorp/go-changelog/cmd/changelog-entry@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_VERSION)

.PHONY: docs
docs: ## Give URL to test Terraform documentation
	@echo "Use this site to preview markdown rendering: https://registry.terraform.io/tools/doc-preview"

.PHONY: tflint
tflint: fmtcheck ## Linter for Terraform files
	tflint -f compact --recursive --minimum-failure-severity=warning

.PHONY: tf-validate
tf-validate: fmtcheck ## Validate Terraform files
	scripts/tf-validate.sh

.PHONY: link-git-hooks
link-git-hooks: ## Install Git hooks
	@echo "==> Installing all git hooks..."
	find .git/hooks -type l -exec rm {} \;
	find .githooks -type f -exec ln -sf ../../{} .git/hooks/ \;

.PHONY: update-atlas-sdk
update-atlas-sdk: ## Update the Atlas SDK dependency
	./scripts/update-sdk.sh

# e.g. run: make scaffold resource_name=streamInstance type=resource
# - type argument can have the values: `resource`, `data-source`, `plural-data-source`.
# details on usage can be found in contributing/development-best-practices.md under "Scaffolding initial Code and File Structure"
.PHONY: scaffold
scaffold: ## Create scaffolding for a new resource
	@go run ./tools/scaffold/*.go $(resource_name) $(type)
	@echo "Reminder: configure the new $(type) in provider.go"

# e.g. run: make scaffold-schemas resource_name=streamInstance
# details on usage can be found in contributing/development-best-practices.md under "Generating Schema and Model Definitions - Using schema generation HashiCorp tooling"
.PHONY: scaffold-schemas
scaffold-schemas: ## Create the schema scaffolding for a new resource
	@scripts/schema-scaffold.sh $(resource_name)

# e.g. run: make generate-schema resource_name=search_deployment
# resource_name is optional, if not provided all configured resources will be generated
# details on usage can be found in contributing/development-best-practices.md under "Generating Schema and Model Definitions - Using internal tool"
.PHONY: generate-schema
generate-schema: ## Generate the schema for a resource
	@go run ./tools/codegen/main.go $(resource_name)

.PHONY: generate-doc
# e.g. run: make generate-doc resource_name=search_deployment
# generate the resource documentation via tfplugindocs
generate-doc: ## Auto-generate the documentation for a resource
	@scripts/generate-doc.sh ${resource_name}

# generate the resource documentation via tfplugindocs for all resources that have templates
.PHONY: generate-docs-all
generate-docs-all: ## Auto-generate the documentation for all resources
	@scripts/generate-docs-all.sh

.PHONY: update-tf-compatibility-matrix
update-tf-compatibility-matrix: ## Update Terraform Compatibility Matrix documentation
	./scripts/update-tf-compatibility-matrix.sh

.PHONY: update-tf-version-in-repository
update-tf-version-in-repository: ## Update Terraform versions
	./scripts/update-tf-version-in-repository.sh

.PHONY: update-changelog-unreleased-section
update-changelog-unreleased-section: ## Update changelog unreleased section
	./scripts/update-changelog-unreleased-section.sh
  
.PHONY: generate-changelog-entry
generate-changelog-entry: ## Generate a changelog entry in a PR
	./scripts/generate-changelog-entry.sh

.PHONY: check-changelog-entry-file
check-changelog-entry-file: ## Check a changelog entry file in a PR
	go run ./tools/check-changelog-entry-file/*.go

.PHONY: jira-release-version
jira-release-version: ## Update Jira version in a release
	go run ./tools/jira-release-version/*.go

.PHONY: enable-advancedclustertpf
enable-advancedclustertpf: ## Enable Advanced Cluster V2 Schema
	make change-lines filename=./internal/config/advanced_cluster_v2_schema.go find="allowAdvancedClusterV2Schema = false" new="allowAdvancedClusterV2Schema = true"

.PHONY: delete-lines ${filename} ${delete}
delete-lines:
	rm -f file.tmp
	grep -v "${delete}" "${filename}" > file.tmp
	mv file.tmp ${filename}
	goimports -w ${filename}

.PHONY: add-lines ${filename} ${find} ${add}
add-lines:
	rm -f file.tmp
	sed 's/${find}/${find}${add}/' "${filename}" > "file.tmp"
	mv file.tmp ${filename}
	goimports -w ${filename}

.PHONY: change-lines ${filename} ${find} ${new}
change-lines:
	rm -f file.tmp
	sed 's/${find}/${new}/' "${filename}" > "file.tmp"
	mv file.tmp ${filename}
	goimports -w ${filename}
