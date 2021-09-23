# Configure the MongoDB Atlas Provider and connect via a key
provider "mongodbatlas" {
  public_key  = var.public_key
  private_key = var.private_key
}
# Create the mongodb atlas Azure cluster
resource "mongodbatlas_cluster" "azure-cluster" {
  project_id   = var.project_id
  name         = var.name
  cluster_type = "REPLICASET"
  replication_specs {
    num_shards = 1
    regions_config {
      region_name     = var.provider_region_name
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  backup_enabled               = false
  auto_scaling_disk_gb_enabled = true
  mongo_db_major_version       = "4.2"

  # Provider settings block in this case it is Azure
  provider_name               = "AZURE"
  provider_disk_type_name     = var.provider_disk_type_name
  provider_instance_size_name = var.provider_instance_size_name
}

# Create the peering connection request
resource "mongodbatlas_network_peering" "test" {
  project_id            = var.project_id
  container_id          = mongodbatlas_cluster.azure-cluster.container_id
  provider_name         = "AZURE"
  azure_directory_id    = data.azurerm_client_config.current.tenant_id
  azure_subscription_id = data.azurerm_client_config.current.subscription_id
  resource_group_name   = var.resource_group_name
  vnet_name             = var.vnet_name
  atlas_cidr_block      = var.atlas_cidr_block
}
