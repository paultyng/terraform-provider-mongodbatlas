output "project_id" {
  value = mongodbatlas_project.test.id
}
output "private_endpoint_id" {
  value = mongodbatlas_private_endpoint.test.private_link_id
}
output "vpc_endpoint_id" {
  value = aws_vpc_endpoint.ptfe_service.id
}
