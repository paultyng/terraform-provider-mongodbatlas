package streamprivatelinkendpoint_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

var (
	resourceType         = "mongodbatlas_stream_privatelink_endpoint"
	resourceName         = "mongodbatlas_stream_privatelink_endpoint.test"
	dataSourceName       = "data.mongodbatlas_stream_privatelink_endpoint.test"
	dataSourcePluralName = "data.mongodbatlas_stream_privatelink_endpoints.test"
)

func TestAccStreamPrivatelinkEndpoint_basic(t *testing.T) {
	acc.SkipTestForCI(t) // needs confluent cloud resources
	tc := basicTestCase(t)
	// Tests include testing of plural data source and so cannot be run in parallel
	resource.Test(t, *tc)
}

func TestAccStreamPrivatelinkEndpoint_failedUpdate(t *testing.T) {
	acc.SkipTestForCI(t) // needs confluent cloud resources
	tc := failedUpdateTestCase(t)
	// Tests include testing of plural data source and so cannot be run in parallel
	resource.Test(t, *tc)
}

func TestAccStreamPrivatelinkEndpoint_missingRequiredFields(t *testing.T) {
	acc.SkipTestForCI(t) // needs confluent cloud resources
	tc := missingRequiredFieldsTestCase(t)
	// Tests include testing of plural data source and so cannot be run in parallel
	resource.Test(t, *tc)
}

func basicTestCase(t *testing.T) *resource.TestCase {
	t.Helper()

	var (
		projectID           = acc.ProjectIDExecution(t)
		provider            = "AWS"
		region              = "us-east-1"
		awsAccountID        = os.Getenv("AWS_ACCOUNT_ID")
		networkID           = os.Getenv("CONFLUENT_CLOUD_NETWORK_ID")
		privatelinkAccessID = os.Getenv("CONFLUENT_CLOUD_PRIVATELINK_ACCESS_ID")
		config              = acc.GetCompleteConfluentConfig(true, true, projectID, provider, region, vendor, awsAccountID, networkID, privatelinkAccessID)
	)

	return &resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t); acc.PreCheckConfluentAWSPrivatelink(t) },
		CheckDestroy:             checkDestroy,
		ExternalProviders:        acc.ExternalProvidersOnlyConfluent(),
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{

				Config: config,
				Check:  checksStreamPrivatelinkEndpoint(projectID, provider, region, vendor, false),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: importStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	}
}

func failedUpdateTestCase(t *testing.T) *resource.TestCase {
	t.Helper()

	var (
		projectID           = acc.ProjectIDExecution(t)
		provider            = "AWS"
		region              = "us-east-1"
		vendor              = "CONFLUENT"
		awsAccountID        = os.Getenv("AWS_ACCOUNT_ID")
		networkID           = os.Getenv("CONFLUENT_CLOUD_NETWORK_ID")
		privatelinkAccessID = os.Getenv("CONFLUENT_CLOUD_PRIVATELINK_ACCESS_ID")
	)

	return &resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		CheckDestroy:             checkDestroy,
		ExternalProviders:        acc.ExternalProvidersOnlyConfluent(),
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{
				Config: acc.GetCompleteConfluentConfig(true, true, projectID, provider, region, vendor, awsAccountID, networkID, privatelinkAccessID),
				Check:  checksStreamPrivatelinkEndpoint(projectID, provider, region, vendor, false),
			},
			{
				Config:      acc.GetCompleteConfluentConfig(true, false, projectID, provider, region, vendor, awsAccountID, networkID, privatelinkAccessID),
				ExpectError: regexp.MustCompile(`Operation not supported`),
			},
		},
	}
}

func missingRequiredFieldsTestCase(t *testing.T) *resource.TestCase {
	t.Helper()

	var (
		projectID           = acc.ProjectIDExecution(t)
		provider            = "AWS"
		vendor              = "CONFLUENT"
		networkID           = os.Getenv("CONFLUENT_CLOUD_NETWORK_ID")
		privatelinkAccessID = os.Getenv("CONFLUENT_CLOUD_PRIVATELINK_ACCESS_ID")
	)

	return &resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		CheckDestroy:             checkDestroy,
		ExternalProviders:        acc.ExternalProvidersOnlyConfluent(),
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{
				Config:      acc.ConfigDataConfluentDedicatedCluster(networkID, privatelinkAccessID) + missingRequiredFieldsConfig(projectID, provider, vendor),
				ExpectError: regexp.MustCompile(`(?s)^.*?service_endpoint_id is required for vendor CONFLUENT.*?dns_domain is required for vendor CONFLUENT.*?region is required for vendor CONFLUENT.*$`),
			},
		},
	}
}

func missingRequiredFieldsConfig(projectID, provider, vendor string) string {
	return fmt.Sprintf(`
	resource "mongodbatlas_stream_privatelink_endpoint" "test" {
		project_id          = %[1]q
		provider_name       = %[2]q
		vendor              = %[3]q
	}`, projectID, provider, vendor)
}

func checksStreamPrivatelinkEndpoint(projectID, provider, region, vendor string, dnsSubdomainsCheck bool) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{checkExists()}
	attrMap := map[string]string{
		"project_id":    projectID,
		"provider_name": provider,
		"region":        region,
		"vendor":        vendor,
	}
	pluralMap := map[string]string{
		"project_id": projectID,
		"results.#":  "1",
	}
	attrSet := []string{
		"id",
		"interface_endpoint_id",
		"state",
		"dns_domain",
		"service_endpoint_id",
	}
	if dnsSubdomainsCheck {
		attrSet = append(attrSet, "dns_sub_domain.0")
	}
	checks = acc.AddAttrChecks(dataSourcePluralName, checks, pluralMap)
	return acc.CheckRSAndDS(resourceName, &dataSourceName, &dataSourcePluralName, attrSet, attrMap, checks...)
}

func checkExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type == resourceType {
				projectID := rs.Primary.Attributes["project_id"]
				connectionID := rs.Primary.Attributes["id"]
				_, _, err := acc.ConnV2().StreamsApi.GetPrivateLinkConnection(context.Background(), projectID, connectionID).Execute()
				if err != nil {
					return fmt.Errorf("Privatelink Connection (%s:%s) not found", projectID, connectionID)
				}
			}
		}
		return nil
	}
}

func checkDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type == resourceType {
			projectID := rs.Primary.Attributes["project_id"]
			connectionID := rs.Primary.Attributes["id"]
			_, _, err := acc.ConnV2().StreamsApi.GetPrivateLinkConnection(context.Background(), projectID, connectionID).Execute()
			if err == nil {
				return fmt.Errorf("Privatelink Connection (%s:%s) still exists", projectID, id)
			}
		}
	}
	return nil
}

func importStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s-%s", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["id"]), nil
	}
}
