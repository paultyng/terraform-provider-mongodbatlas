output "project_id" {
  value = mongodbatlas_project.test.id
}
output "entry" {
  value = mongodbatlas_project_ip_whitelist.test.ip_address
}
