package acc

import (
	"fmt"
	"strings"
)

func ConfigDataConfluentDedicatedCluster(networkID, privatelinkAccessID string) string {
	return fmt.Sprintf(`
	%[1]s

	data "confluent_environment" "default_environment" {
  		display_name = "default"
	}

	data "confluent_network" "private-link" {
		id     = %[2]q
		environment {
			id = data.confluent_environment.default_environment.id
		}
	}

	data "confluent_private_link_access" "aws" {
		id = %[3]q
		environment {
			id = data.confluent_environment.default_environment.id
		}
	}`, ConfigConfluentProvider(), networkID, privatelinkAccessID)
}

func configBasic(projectID, provider, region, vendor string, withDNSSubdomains bool) string {
	dnsSubDomainConfig := ""
	if withDNSSubdomains {
		dnsSubDomainConfig = `dns_sub_domain = local.dns_sub_domain_entries`
	}

	return fmt.Sprintf(`
	locals {
		dns_sub_domain_entries = [
    		for zone in confluent_network.private-link.zones :
    		"${zone}.${confluent_network.private-link.dns_domain}"
  		]
	}	

	resource "mongodbatlas_stream_privatelink_endpoint" "test" {
		project_id          = %[1]q
		dns_domain          = confluent_network.private-link.dns_domain
		provider_name       = %[2]q
		region              = %[3]q
		vendor              = %[4]q
		service_endpoint_id = confluent_network.private-link.aws[0].private_link_endpoint_service
		%[5]s
		depends_on = [
			confluent_kafka_cluster.dedicated
		]
	}

	data "mongodbatlas_stream_privatelink_endpoint" "test" {
		project_id = %[1]q
		id         = mongodbatlas_stream_privatelink_endpoint.test.id
		depends_on = [
    		mongodbatlas_stream_privatelink_endpoint.test
  		]
	}

	data "mongodbatlas_stream_privatelink_endpoints" "test" {
		project_id = %[1]q
		depends_on = [
    		mongodbatlas_stream_privatelink_endpoint.test
  		]
	}`, projectID, provider, region, vendor, dnsSubDomainConfig)
}

func configNewConfluentDedicatedCluster(provider, region, awsAccountID string) string {
	return fmt.Sprintf(`
	%[1]s

	data "confluent_environment" "default_environment" {
  		display_name = "default"
	}

	resource "confluent_network" "private-link" {
		display_name     = "terraform-test-private-link-network"
		cloud            = %[2]q
		region           = %[3]q
		connection_types = ["PRIVATELINK"]
		zones            = ["use1-az1", "use1-az2", "use1-az4"]
		environment {
			id = data.confluent_environment.default_environment.id
		}
		dns_config {
			resolution = "PRIVATE"
		}
	}

	resource "confluent_private_link_access" "aws" {
		display_name = "terraform-test-aws-private-link-access"
		aws {
			account = %[4]q
		}
		environment {
			id = data.confluent_environment.default_environment.id
		}
		network {
			id = confluent_network.private-link.id
		}
	}

	resource "confluent_kafka_cluster" "dedicated" {
		display_name = "terraform-test-cluster"
		availability = "MULTI_ZONE"
		cloud        = confluent_network.private-link.cloud
		region       = confluent_network.private-link.region
		dedicated {
			cku = 2
		}
		environment {
			id = data.confluent_environment.default_environment.id
		}
		network {
			id = confluent_network.private-link.id
		}
	}`, ConfigConfluentProvider(), provider, region, awsAccountID)
}

func GetCompleteConfluentConfig(usesExistingConfluentCluster, withDNSSubdomains bool, projectID, provider, region, vendor, awsAccountID, networkID, privatelinkAccessID string) string {
	if usesExistingConfluentCluster {
		configBasicUsingDatasources := strings.ReplaceAll(configBasic(projectID, provider, region, vendor, withDNSSubdomains), "confluent_network.private-link", "data.confluent_network.private-link")
		configBasicUsingDatasourcesWithoutDependsOnCluster := strings.ReplaceAll(configBasicUsingDatasources, "confluent_kafka_cluster.dedicated", "")
		return ConfigDataConfluentDedicatedCluster(networkID, privatelinkAccessID) + configBasicUsingDatasourcesWithoutDependsOnCluster
	}
	return configNewConfluentDedicatedCluster(provider, region, awsAccountID) + configBasic(projectID, provider, region, vendor, true)
}
