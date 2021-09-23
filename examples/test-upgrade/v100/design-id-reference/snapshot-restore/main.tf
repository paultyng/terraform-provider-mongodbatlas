data "mongodbatlas_project" "test" {
  name = var.project_name
}

resource "mongodbatlas_cluster" "my_cluster" {
  project_id   = data.mongodbatlas_project.test.id
  name         = var.cluster_name
  disk_size_gb = 5

  # Provider Settings "block"
  provider_name               = "AWS"
  provider_region_name        = "US_EAST_1"
  provider_instance_size_name = "M10"
  provider_backup_enabled     = true # enable cloud provider snapshots
}

resource "mongodbatlas_cloud_provider_snapshot" "test" {
  project_id        = data.mongodbatlas_project.test.id
  cluster_name      = mongodbatlas_cluster.my_cluster.name
  description       = var.description
  retention_in_days = var.retention_in_days
}

resource "mongodbatlas_cloud_provider_snapshot_restore_job" "test" {
  project_id   = data.mongodbatlas_project.test.id
  cluster_name = mongodbatlas_cloud_provider_snapshot.test.cluster_name
  snapshot_id  = mongodbatlas_cloud_provider_snapshot.test.id
  delivery_type_config {
    download = true
  }
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_cloud_provider_snapshot_restore_job" "test" {
  project_id   = data.mongodbatlas_project.test.id
  cluster_name = mongodbatlas_cloud_provider_snapshot_restore_job.test.cluster_name
  job_id       = mongodbatlas_cloud_provider_snapshot_restore_job.test.id
}

output "snapshot_id_state" {
  value = mongodbatlas_cloud_provider_snapshot.test.id
}
output "snapshot_id" {
  value = mongodbatlas_cloud_provider_snapshot.test.snapshot_id
}
output "snapshot_restore_id_state" {
  value = mongodbatlas_cloud_provider_snapshot_restore_job.test.id
}
output "snapshot_restore_job_id" {
  value = mongodbatlas_cloud_provider_snapshot_restore_job.test.snapshot_restore_job_id
}
