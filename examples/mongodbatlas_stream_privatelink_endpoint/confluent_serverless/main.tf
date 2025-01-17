resource "confluent_environment" "staging" {
  display_name = "Staging"
}

resource "confluent_private_link_attachment" "this" {
  cloud        = "AWS"
  region       = var.aws_region
  display_name = "private-link-attachment"
  environment {
    id = confluent_environment.staging.id
  }
}

module "privatelink" {
  source                   = "./aws-privatelink-endpoint"
  vpc_id                   = var.vpc_id
  privatelink_service_name = confluent_private_link_attachment.this.aws[0].vpc_endpoint_service_name
  dns_domain               = confluent_private_link_attachment.this.dns_domain
  subnets_to_privatelink   = var.subnets_to_privatelink
}

resource "confluent_private_link_attachment_connection" "this" {
  display_name = "private-link-attachment-connection"
  environment {
    id = confluent_environment.staging.id
  }
  aws {
    vpc_endpoint_id = module.privatelink.vpc_endpoint_id
  }
  private_link_attachment {
    id = confluent_private_link_attachment.this.id
  }
}

resource "mongodbatlas_stream_privatelink_endpoint" "test" {
  project_id          = var.project_id
  dns_domain          = confluent_private_link_attachment.this.dns_domain
  provider_name       = "AWS"
  region              = var.aws_region
  vendor              = "CONFLUENT"
  service_endpoint_id = confluent_private_link_attachment.this.aws[0].vpc_endpoint_service_name
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
