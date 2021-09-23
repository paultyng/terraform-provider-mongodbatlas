resource "mongodbatlas_project" "test" {
  name   = var.project_name
  org_id = var.org_id
}


resource "mongodbatlas_cloud_provider_access_setup" "setup_only" {
  project_id    = mongodbatlas_project.test.id
  provider_name = "AWS"
}

resource "mongodbatlas_cloud_provider_access_authorization" "auth_role" {
  project_id = mongodbatlas_project.test.id
  role_id    = mongodbatlas_cloud_provider_access_setup.setup_only.role_id

  aws {
    iam_assumed_role_arn = aws_iam_role.test_role.arn
  }
}

resource "mongodbatlas_data_lake" "test" {
  project_id = mongodbatlas_project.test.id
  name       = var.data_lake_name
  aws {
    role_id        = mongodbatlas_cloud_provider_access_authorization.auth_role.role_id
    test_s3_bucket = var.test_s3_bucket
  }
}
