# This cluster is in GCP cloud-provider with VPC peering enabled

resource "mongodbatlas_cluster" "cluster" {
  project_id   = var.project_id
  name         = "cluster-test"
  cluster_type = "REPLICASET"
  replication_specs {
    num_shards = 1
    regions_config {
      region_name     = var.atlas_region
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  labels {
    key   = "environment"
    value = "prod"
  }
  provider_backup_enabled                 = true
  auto_scaling_disk_gb_enabled            = true
  mongo_db_major_version                  = "4.4"
  auto_scaling_compute_enabled            = true
  auto_scaling_compute_scale_down_enabled = true


  # Provider Settings "block"
  provider_name                                   = "GCP"
  provider_instance_size_name                     = "M10"
  provider_auto_scaling_compute_max_instance_size = "M20"
  provider_auto_scaling_compute_min_instance_size = "M10"
  disk_size_gb                                    = 40
  advanced_configuration {
    minimum_enabled_tls_protocol = "TLS1_2"
  }
  lifecycle {
    ignore_changes = [
      provider_instance_size_name
    ]
  }
}
# The connection strings available for the GCP MognoDB Atlas cluster
output "connection_string" {
  value = mongodbatlas_cluster.cluster.connection_strings
}
