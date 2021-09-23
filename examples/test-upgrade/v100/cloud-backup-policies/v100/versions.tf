terraform {
  required_providers {
    mongodbatlas = {
      source  = "mongodb/mongodbatlas"
      version = "1.0.0"
    }
  }
  required_version = ">= 0.13"
}
