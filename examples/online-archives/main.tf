resource "mongodbatlas_online_archive" "users_archive" {
  project_id   = var.project_id
  cluster_name = var.cluster_name
  coll_name    = var.collection_name
  db_name      = var.database_name

  criteria {
    type              = "DATE"
    date_field        = "created"
    date_format       = "ISODATE"
    expire_after_days = 2
  }

  partition_fields {
    field_name = var.partition_field_one
    order      = 0
  }

  partition_fields {
    field_name = var.partition_field_two
    order      = 1
  }
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_online_archive" "read_archive" {
  project_id   = mongodbatlas_online_archive.users_archive.project_id
  cluster_name = mongodbatlas_online_archive.users_archive.cluster_name
  archive_id   = mongodbatlas_online_archive.users_archive.archive_id
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_online_archives" "all" {
  project_id   = mongodbatlas_online_archive.users_archive.project_id
  cluster_name = mongodbatlas_online_archive.users_archive.cluster_name
}
