resource "mongodbatlas_project" "test" {
  name   = var.project_name
  org_id = var.org_id
}
