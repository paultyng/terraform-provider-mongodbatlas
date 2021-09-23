# This will Create a Project,  Cluster and Modify the 4 Default Policies Simultaneously

resource "mongodbatlas_project" "project_test" {
  name   = var.project_name
  org_id = var.org_id
}

resource "mongodbatlas_cluster" "cluster_test" {
  project_id   = mongodbatlas_project.project_test.id
  name         = var.cluster_name
  disk_size_gb = 5

  # Provider Settings "block"
  provider_name               = "AWS"
  provider_region_name        = "EU_CENTRAL_1"
  provider_instance_size_name = "M10"
  provider_backup_enabled     = true # enable cloud provider snapshots
}

resource "mongodbatlas_cloud_provider_snapshot_backup_policy" "test" {
  project_id   = mongodbatlas_cluster.cluster_test.project_id
  cluster_name = mongodbatlas_cluster.cluster_test.name

  reference_hour_of_day    = 3
  reference_minute_of_hour = 45
  restore_window_days      = 4


  policies {
    id = mongodbatlas_cluster.cluster_test.snapshot_backup_policy[0].policies[0].id

    policy_item {
      id                 = mongodbatlas_cluster.cluster_test.snapshot_backup_policy[0].policies[0].policy_item[0].id
      frequency_interval = 1
      frequency_type     = "hourly"
      retention_unit     = "days"
      retention_value    = 1
    }
    policy_item {
      id                 = mongodbatlas_cluster.cluster_test.snapshot_backup_policy[0].policies[0].policy_item[1].id
      frequency_interval = 1
      frequency_type     = "daily"
      retention_unit     = "days"
      retention_value    = 2
    }
    policy_item {
      id                 = mongodbatlas_cluster.cluster_test.snapshot_backup_policy[0].policies[0].policy_item[2].id
      frequency_interval = 4
      frequency_type     = "weekly"
      retention_unit     = "weeks"
      retention_value    = 3
    }
    policy_item {
      id                 = mongodbatlas_cluster.cluster_test.snapshot_backup_policy[0].policies[0].policy_item[3].id
      frequency_interval = 5
      frequency_type     = "monthly"
      retention_unit     = "months"
      retention_value    = 4
    }
  }
}

output "project_id" {
  value = mongodbatlas_project.project_test.id
}
output "cluster_name" {
  value = mongodbatlas_cluster.cluster_test.name
}
