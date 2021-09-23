# This is the same as version v091 but for mongodbatlas_cloud_backup_schedule but
# does nothing, it's only to confirm that it works for import from v091 and won't appear changes
# for cloud backup schedule

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
  cloud_backup                = true # enable cloud provider snapshots
  provider_disk_iops          = 1000
}

resource "mongodbatlas_cloud_backup_schedule" "test" {
  project_id   = mongodbatlas_cluster.cluster_test.project_id
  cluster_name = mongodbatlas_cluster.cluster_test.name

  reference_hour_of_day    = 3
  reference_minute_of_hour = 45
  restore_window_days      = 4

  policy_item_hourly {
    frequency_interval = 1
    retention_unit     = "days"
    retention_value    = 1
  }
  policy_item_daily {
    frequency_interval = 1
    retention_unit     = "days"
    retention_value    = 2
  }
  policy_item_weekly {
    frequency_interval = 4
    retention_unit     = "weeks"
    retention_value    = 3
  }
  policy_item_monthly {
    frequency_interval = 5
    retention_unit     = "months"
    retention_value    = 4
  }
}
