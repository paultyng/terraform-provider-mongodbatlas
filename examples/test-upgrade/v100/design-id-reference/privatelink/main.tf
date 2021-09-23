data "mongodbatlas_project" "test" {
  name = var.project_name
}

provider "aws" {
  region     = "us-west-1"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

resource "mongodbatlas_privatelink_endpoint" "test" {
  project_id    = data.mongodbatlas_project.test.id
  provider_name = "AWS"
  region        = "us-west-1"
}

resource "aws_vpc_endpoint" "ptfe_service" {
  vpc_id             = var.aws_vpc_id
  service_name       = mongodbatlas_privatelink_endpoint.test.endpoint_service_name
  vpc_endpoint_type  = "Interface"
  subnet_ids         = [var.aws_subnet_ids]
  security_group_ids = [var.aws_sg_ids]
}

resource "mongodbatlas_privatelink_endpoint_service" "test" {
  project_id          = data.mongodbatlas_project.test.id
  private_link_id     = mongodbatlas_privatelink_endpoint.test.id
  endpoint_service_id = aws_vpc_endpoint.ptfe_service.id
  provider_name       = "AWS"
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_privatelink_endpoint" "test" {
  project_id      = data.mongodbatlas_project.test.id
  private_link_id = mongodbatlas_privatelink_endpoint.test.id
  provider_name   = "AWS"
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_privatelink_endpoint_service" "test" {
  project_id          = data.mongodbatlas_project.test.id
  private_link_id     = mongodbatlas_privatelink_endpoint_service.test.id
  endpoint_service_id = mongodbatlas_privatelink_endpoint_service.test.endpoint_service_id
  provider_name       = "AWS"
}
