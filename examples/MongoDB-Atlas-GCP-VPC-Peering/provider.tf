provider "mongodbatlas" {
  public_key  = var.public_key
  private_key = var.private_key
}
provider "google" {
  # credentials = file("service-account.json")
  project = var.gcpprojectid
  region  = var.gcp_region
  # zone="us-central-1c"
}
