provider "azurerm" {
  subscription_id = var.subscription_id
  client_id       = var.client_id
  client_secret   = var.client_secret
  tenant_id       = var.tenant_id
  features {
  }
}

data "azurerm_resource_group" "test" {
  name = var.resource_group_name
}

resource "azurerm_virtual_network" "test" {
  name                = "acceptanceTestVirtualNetwork1"
  address_space       = ["10.0.0.0/16"]
  location            = data.azurerm_resource_group.test.location
  resource_group_name = var.resource_group_name
}

resource "azurerm_subnet" "test" {
  name                                           = "testsubnet"
  resource_group_name                            = var.resource_group_name
  virtual_network_name                           = azurerm_virtual_network.test.name
  address_prefixes                               = ["10.0.1.0/24"]
  enforce_private_link_service_network_policies  = true
  enforce_private_link_endpoint_network_policies = true
}

resource "mongodbatlas_privatelink_endpoint" "test" {
  project_id    = var.project_id
  provider_name = "AZURE"
  region        = "eastus2"
}

resource "azurerm_private_endpoint" "test" {
  name                = "endpoint-test"
  location            = data.azurerm_resource_group.test.location
  resource_group_name = var.resource_group_name
  subnet_id           = azurerm_subnet.test.id
  private_service_connection {
    name                           = mongodbatlas_privatelink_endpoint.test.private_link_service_name
    private_connection_resource_id = mongodbatlas_privatelink_endpoint.test.private_link_service_resource_id
    is_manual_connection           = true
    request_message                = "Azure Private Link test"
  }

}

resource "mongodbatlas_privatelink_endpoint_service" "test" {
  project_id                  = mongodbatlas_privatelink_endpoint.test.project_id
  private_link_id             = mongodbatlas_privatelink_endpoint.test.private_link_id
  endpoint_service_id         = azurerm_private_endpoint.test.id
  private_endpoint_ip_address = azurerm_private_endpoint.test.private_service_connection[0].private_ip_address
  provider_name               = "AZURE"
}
