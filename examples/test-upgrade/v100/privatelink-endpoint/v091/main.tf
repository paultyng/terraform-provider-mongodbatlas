# This config will create aws endpoint, private endpoint and private endpoint interface
# To verify that everything is working even after terraform plan and show not changes
# and then do the import stuff in the folder for v100

resource "mongodbatlas_project" "test" {
  name   = var.project_name
  org_id = var.org_id
}

provider "aws" {
  region     = "us-east-1"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

resource "mongodbatlas_private_endpoint" "test" {
  project_id    = mongodbatlas_project.test.id
  provider_name = "AWS"
  region        = "us-east-1"
}

resource "aws_vpc_endpoint" "ptfe_service" {
  vpc_id             = var.aws_vpc_id
  service_name       = mongodbatlas_private_endpoint.test.endpoint_service_name
  vpc_endpoint_type  = "Interface"
  subnet_ids         = [var.aws_subnet_ids]
  security_group_ids = [var.aws_sg_ids]
}
resource "mongodbatlas_private_endpoint_interface_link" "test" {
  project_id            = mongodbatlas_private_endpoint.test.project_id
  private_link_id       = mongodbatlas_private_endpoint.test.private_link_id
  interface_endpoint_id = aws_vpc_endpoint.ptfe_service.id
}
