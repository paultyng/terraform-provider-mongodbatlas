# Container example provided but not always required, 
# see network_container documentation for details. 
resource "mongodbatlas_network_container" "test" {
  project_id       = var.project_id
  atlas_cidr_block = "10.8.0.0/18"
  provider_name    = "GCP"
}

# Create the peering connection request
resource "mongodbatlas_network_peering" "test" {
  project_id     = var.project_id
  container_id   = mongodbatlas_network_container.test.container_id
  provider_name  = "GCP"
  gcp_project_id = var.gcpprojectid
  network_name   = "default"
}

# the following assumes a GCP provider is configured
data "google_compute_network" "default" {
  name = "default"
}

# Create the GCP peer
resource "google_compute_network_peering" "peering" {
  name = "peering-gcp-terraform-test"
  # The URI of the GCP VPC. self_link which is found by enabling the [Compute Engine API](https://console.cloud.google.com/apis/api/compute.googleapis.com)
  network = data.google_compute_network.default.self_link
  # The URI of the Atlas VPC
  peer_network = "https://www.googleapis.com/compute/v1/projects/${mongodbatlas_network_peering.test.atlas_gcp_project_id}/global/networks/${mongodbatlas_network_peering.test.atlas_vpc_name}"
}

# Create IP Access List for connection from GCP
# You will need to add the private IP ranges of the subnets in which your application is hosted to the IP access list in order to connect to your Atlas cluster. GCP networks generated in auto-mode use a CIDR range of 10.128.0.0/9
resource "mongodbatlas_project_ip_access_list" "test" {
  project_id = var.project_id
  cidr_block = var.gcp_cidr
  comment    = "cidr block for GCP VPC Whitelist"
}
