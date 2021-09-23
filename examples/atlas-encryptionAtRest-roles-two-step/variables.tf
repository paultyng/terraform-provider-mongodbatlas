variable "public_key" {
  description = "The public API key for MongoDB Atlas"
  default     = ""
}
variable "private_key" {
  description = "The private API key for MongoDB Atlas"
  default     = ""
}
variable "project_id" {
  description = "Atlas project ID"
  default     = ""
}
variable "access_key" {
  description = "The access key for AWS Account"
  default     = ""
}
variable "secret_key" {
  description = "The secret key for AWS Account"
  default     = ""
}
variable "aws_region" {
  default     = "us-east-1"
  description = "AWS Region"
}
