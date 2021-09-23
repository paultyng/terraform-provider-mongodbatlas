output "container_id_state" {
  value = mongodbatlas_network_container.test.id
}
output "container_id" {
  value = mongodbatlas_network_container.test.container_id
}
output "peering_id_state" {
  value = mongodbatlas_network_peering.test.id
}
output "peer_id" {
  value = mongodbatlas_network_peering.test.peer_id
}
