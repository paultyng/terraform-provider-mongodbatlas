variable "project_id" {
  description = "Unique 24-hexadecimal digit string that identifies your project"
  type        = string
}

variable "public_key" {
  description = "Public API key to authenticate to Atlas"
  type        = string
}
variable "private_key" {
  description = "Private API key to authenticate to Atlas"
  type        = string
}

variable "confluent_cloud_api_key" {
  description = "Public API key to authenticate to Confluent Cloud"
  type        = string
}
variable "confluent_cloud_api_secret" {
  description = "Private API key to authenticate to Confleunt Cloud"
  type        = string
}

variable "subnets_to_privatelink" {
  description = "A map of Zone ID to Subnet ID (i.e.: {\"use1-az1\" = \"subnet-abcdef0123456789a\", ...})"
  type        = map(string)
}

variable "aws_region" {
  description = "The AWS Region"
  type        = string
}

variable "vpc_id" {
  description = "The ID of the VPC in which the endpoint will be used."
  type        = string
}
