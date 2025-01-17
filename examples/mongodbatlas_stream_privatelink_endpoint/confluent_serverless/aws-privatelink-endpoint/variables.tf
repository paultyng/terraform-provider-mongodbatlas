variable "vpc_id" {
  description = "The VPC ID to private link to Confluent Cloud"
  type        = string
}

variable "privatelink_service_name" {
  description = "The Service Name from Confluent Cloud to Private Link with (provided by Confluent)"
  type        = string
}

variable "dns_domain" {
  description = "The root DNS domain for the Private Link Attachment, for example, `pr123a.us-east-2.aws.confluent.cloud`"
  type        = string
}

variable "subnets_to_privatelink" {
  description = "A map of Zone ID to Subnet ID (ie: {\"use1-az1\" = \"subnet-abcdef0123456789a\", ...})"
  type        = map(string)
}