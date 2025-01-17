resource "confluent_environment" "staging" {
  display_name = "Staging"
}

resource "confluent_network" "private_link" {
  display_name     = "terraform-test-private-link-network-manual"
  cloud            = "AWS"
  region           = var.aws_region
  connection_types = ["PRIVATELINK"]
  zones            = keys(var.subnets_to_privatelink)
  environment {
    id = confluent_environment.staging.id
  }
  dns_config {
    resolution = "PRIVATE"
  }
}

resource "confluent_private_link_access" "aws" {
  display_name = "example-private-link-access"
  aws {
    account = var.aws_account_id
  }
  environment {
    id = confluent_environment.staging.id
  }
  network {
    id = confluent_network.private_link.id
  }
}

resource "confluent_kafka_cluster" "dedicated" {
  display_name = "example-dedicated-cluster"
  availability = "MULTI_ZONE"
  cloud        = confluent_network.private_link.cloud
  region       = confluent_network.private_link.region
  dedicated {
    cku = 2
  }
  environment {
    id = confluent_environment.staging.id
  }
  network {
    id = confluent_network.private_link.id
  }
}

resource "mongodbatlas_stream_privatelink_endpoint" "test" {
  project_id          = var.project_id
  dns_domain          = confluent_network.private_link.dns_domain
  provider_name       = "AWS"
  region              = var.aws_region
  vendor              = "CONFLUENT"
  service_endpoint_id = confluent_network.private_link.aws[0].private_link_endpoint_service
  dns_sub_domain      = confluent_network.private_link.zonal_subdomains
}

data "mongodbatlas_stream_privatelink_endpoint" "singular_datasource" {
  project_id = var.project_id
  id         = mongodbatlas_stream_privatelink_endpoint.test.id
}

data "mongodbatlas_stream_privatelink_endpoints" "plural_datasource" {
  project_id = var.project_id
}

output "interface_endpoint_id" {
  value = data.mongodbatlas_stream_privatelink_endpoint.singular_datasource.interface_endpoint_id
}

output "interface_endpoint_ids" {
  value = data.mongodbatlas_stream_privatelink_endpoints.plural_datasource.results[*].interface_endpoint_id
}
