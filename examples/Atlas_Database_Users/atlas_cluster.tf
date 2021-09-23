resource "mongodbatlas_cluster" "cluster" {
  project_id             = mongodbatlas_project.project1.id
  name                   = "MongoDB_Atlas"
  mongo_db_major_version = "4.4"
  cluster_type           = "REPLICASET"
  replication_specs {
    num_shards = 1
    regions_config {
      region_name     = var.region
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  # Provider Settings "block"
  provider_backup_enabled      = true
  auto_scaling_disk_gb_enabled = true
  provider_name                = "AWS"
  disk_size_gb                 = 10
  provider_instance_size_name  = "M10"
}
output "atlasclusterstring" {
  value = mongodbatlas_cluster.cluster.connection_strings
}
