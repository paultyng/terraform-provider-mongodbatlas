# gcp
variable "service_account_key" {
  default = ""
}
variable "gcp_key_version_resource_id" {
  default = ""
}

# encryption at rest
variable "project_name" {
  description = "Atlas project name"
  default     = ""
}
variable "org_id" {
  description = "The organization ID"
  default     = ""
}
