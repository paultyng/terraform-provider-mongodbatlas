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
	tc := basicTestCase(t)
	// Tests include testing of plural data source and so cannot be run in parallel
	resource.Test(t, *tc)
}

func TestAccStreamPrivatelinkEndpoint_failedUpdate(t *testing.T) {
	tc := failedUpdateTestCase(t)
	// Tests include testing of plural data source and so cannot be run in parallel
	resource.Test(t, *tc)
}

func basicTestCase(t *testing.T) *resource.TestCase {
	t.Helper()

	var (
		projectID         = acc.ProjectIDExecution(t)
		dnsDomain         = os.Getenv("MONGODB_ATLAS_STREAM_PRIVATELINK_DNS_DOMAIN")
		provider          = "AWS"
		region            = "us-east-1"
		serviceEndpointID = os.Getenv("MONGODB_ATLAS_STREAM_SERVICE_ENDPOINT_ID")
	)

	return &resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: configBasic(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, true),
				Check:  checksStreamPrivatelinkEndpoint(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, false),
			},
			{
				Config:            configBasic(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, true),
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
		projectID         = acc.ProjectIDExecution(t)
		dnsDomain         = os.Getenv("MONGODB_ATLAS_STREAM_PRIVATELINK_DNS_DOMAIN")
		provider          = "AWS"
		region            = "us-east-1"
		vendor            = "CONFLUENT"
		serviceEndpointID = os.Getenv("MONGODB_ATLAS_STREAM_SERVICE_ENDPOINT_ID")
	)

	return &resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: configBasic(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, false),
				Check:  checksStreamPrivatelinkEndpoint(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, false),
			},
			{
				Config:      configBasic(projectID, dnsDomain, provider, region, vendor, serviceEndpointID, true),
				ExpectError: regexp.MustCompile(`Operation not supported`),
			},
		},
	}
}

func configBasic(projectID, dnsDomain, provider, region, vendor, serviceEndpointID string, withDNSSubdomains bool) string {
	dnsSubDomainConfig := ""
	if withDNSSubdomains {
		dnsSubDomainConfig = fmt.Sprintf(`dns_sub_domain = [%[1]q]`, dnsDomain)
	}

	return fmt.Sprintf(`
	resource "mongodbatlas_stream_privatelink_endpoint" "test" {
		project_id          = %[1]q
		dns_domain          = %[2]q
		provider_name       = %[3]q
		region              = %[4]q
		vendor              = %[5]q
		service_endpoint_id = %[6]q
		%[7]s
	}

	data "mongodbatlas_stream_privatelink_endpoint" "singular-datasource-test" {
		project_id = %[1]q
		id         = mongodbatlas_stream_privatelink_endpoint.test.id
	}

	data "mongodbatlas_stream_privatelink_endpoints" "plural-datasource-test" {
		project_id = %[1]q
	}`, projectID, dnsDomain, provider, region, vendor, serviceEndpointID, dnsSubDomainConfig)
}

func checksStreamPrivatelinkEndpoint(projectID, dnsDomain, provider, region, vendor, serviceEndpointID string, dnsSubdomainsCheck bool) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{checkExists()}
	attrMap := map[string]string{
		"project_id":          projectID,
		"dns_domain":          dnsDomain,
		"provider_name":       provider,
		"region":              region,
		"vendor":              vendor,
		"service_endpoint_id": serviceEndpointID,
	}
	if dnsSubdomainsCheck {
		attrMap["dns_sub_domain.0"] = dnsDomain // this check might not work??? verify
	}
	pluralMap := map[string]string{
		"project_id": projectID,
		"results.#":  "1",
	}
	attrSet := []string{
		"id",
		"interface_endpoint_id",
		"state",
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
