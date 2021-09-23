resource "mongodbatlas_project" "test" {
  name   = var.project_name
  org_id = var.org_id
}

resource "mongodbatlas_cluster" "test" {
  project_id                   = mongodbatlas_project.test.id
  name                         = var.cluster_name
  disk_size_gb                 = 100
  num_shards                   = 1
  replication_factor           = 3
  provider_backup_enabled      = true
  pit_enabled                  = true
  auto_scaling_disk_gb_enabled = true
  mongo_db_major_version       = var.mongodb_major_version

  # Provider Settings "block"
  provider_name               = "AWS"
  provider_disk_iops          = 300
  provider_instance_size_name = "M30"
  provider_region_name        = "EU_CENTRAL_1"
}

output "project_id" {
  value = mongodbatlas_project.test.id
}
output "cluster_name" {
  value = mongodbatlas_cluster.test.name
}
