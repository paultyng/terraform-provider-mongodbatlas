# This is only for import stuff because it needs the resource names and set to
# avoid changes when terraform plan
resource "mongodbatlas_project" "test" {
  name   = var.project_name
  org_id = var.org_id
}
resource "mongodbatlas_privatelink_endpoint" "test" {
  project_id    = mongodbatlas_project.test.id
  provider_name = "AWS"
  region        = "us-east-1"
}

provider "aws" {
  region     = "us-east-1"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

resource "aws_vpc_endpoint" "ptfe_service" {
  vpc_id             = var.aws_vpc_id
  service_name       = mongodbatlas_privatelink_endpoint.test.endpoint_service_name
  vpc_endpoint_type  = "Interface"
  subnet_ids         = [var.aws_subnet_ids]
  security_group_ids = [var.aws_sg_ids]
}

resource "mongodbatlas_privatelink_endpoint_service" "test" {
  project_id          = mongodbatlas_privatelink_endpoint.test.project_id
  private_link_id     = mongodbatlas_privatelink_endpoint.test.private_link_id
  endpoint_service_id = aws_vpc_endpoint.ptfe_service.id
  provider_name       = "AWS"
}
