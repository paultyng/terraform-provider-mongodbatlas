output "privatelink_id_state" {
  value = mongodbatlas_privatelink_endpoint.test.id
}
output "private_link_id" {
  value = mongodbatlas_privatelink_endpoint.test.private_link_id
}
output "privatelink_endpoint_service_state" {
  value = mongodbatlas_privatelink_endpoint_service.test.id
}
output "endpoint_service_id" {
  value = mongodbatlas_privatelink_endpoint_service.test.endpoint_service_id
}
