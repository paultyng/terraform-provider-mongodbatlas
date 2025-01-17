terraform {
  required_providers {
    mongodbatlas = {
      source  = "mongodb/mongodbatlas"
      version = "~> 1.24"
    }
    confluent = {
      source  = "confluentinc/confluent"
      version = "2.12.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.0.0"
    }
  }
  required_version = ">= 1.0"
}
