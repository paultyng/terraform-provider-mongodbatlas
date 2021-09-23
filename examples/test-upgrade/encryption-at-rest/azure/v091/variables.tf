# azure
variable "client_id" {
  default = ""
}
variable "subscription_id" {
  default = ""
}
variable "resource_group_name" {
  default = ""
}
variable "client_secret" {
  default = ""
}
variable "tenant_id" {
  default = ""
}
variable "key_vault_name" {
  default = ""
}
variable "key_identifier" {
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
