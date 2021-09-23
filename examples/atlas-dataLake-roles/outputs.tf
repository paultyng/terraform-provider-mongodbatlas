output "project_id" {
  value = mongodbatlas_project.test.id
}
output "role_id" {
  value = mongodbatlas_cloud_provider_access_setup.setup_only.role_id
}
output "role_name" {
  value = aws_iam_role.test_role.name
}
output "policy_name" {
  value = aws_iam_role_policy.test_policy.name
}
output "data_lake_name" {
  value = mongodbatlas_data_lake.test.name
}
output "s3_bucket" {
  value = mongodbatlas_data_lake.test.aws[0].test_s3_bucket
}
