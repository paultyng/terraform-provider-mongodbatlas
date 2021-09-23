data "mongodbatlas_project" "test" {
  name = var.project_name
}

resource "mongodbatlas_network_container" "test" {
  project_id       = data.mongodbatlas_project.test.id
  atlas_cidr_block = "192.168.208.0/21"
  provider_name    = "AWS"
  region_name      = var.region_name
}

resource "mongodbatlas_network_peering" "test" {
  accepter_region_name   = lower(replace(var.region_name, "_", "-"))
  project_id             = data.mongodbatlas_project.test.id
  container_id           = mongodbatlas_network_container.test.id
  provider_name          = "AWS"
  route_table_cidr_block = var.route_table_cidr_block
  vpc_id                 = var.vpc_id
  aws_account_id         = var.aws_account_id
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_network_peering" "test" {
  project_id = data.mongodbatlas_project.test.id
  peering_id = mongodbatlas_network_peering.test.peer_id
}
