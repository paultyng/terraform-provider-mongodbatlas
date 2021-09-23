# mongo
variable "project_id" {
  type = string
}
variable "cloud_provider_access_name" {
  type    = string
  default = "AWS"
}
variable "public_key" {
  type = string
}
variable "private_key" {
  type = string
}

# aws
variable "access_key" {
  type = string
}
variable "secret_key" {
  type = string
}
variable "aws_region" {
  type = string
}

# encryption at rest
variable "customer_master_key" {
  description = "The customer master secret key for AWS Account"
  default     = ""
}

variable "atlas_region" {
  default     = "US_EAST_1"
  description = "Atlas Region"
}

