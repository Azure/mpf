terraform {}

provider "azurerm" {
  features {}
}

resource "random_id" "rg" {
  byte_length = 8
}

resource "random_string" "acr" {
  length  = 12
  special = false
  upper   = false
  numeric = true
}

resource "random_string" "storage" {
  length  = 12
  special = false
  upper   = false
  numeric = true
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = var.location
}

# The container registry is created through a long running operation, the azurerm provider polls
# the returned operationStatuses URL until the registry is ready. Polling requires the
# Microsoft.ContainerRegistry/registries/operationStatuses/read permission.
resource "azurerm_container_registry" "acr" {
  name                = "acr${random_string.acr.result}"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false

  tags = {
    Environment = "Development"
  }
}

# The resources below belong to providers that do not expose a nested operationStatuses
# resource type. MPF still appends the candidate permission for each of their write actions,
# Azure rejects those candidates with InvalidActionOrNotAction, and MPF drops them from the
# result. They are part of this sample so that the discard path is exercised across several
# providers in a single deployment.
resource "azurerm_storage_account" "storage" {
  name                     = "st${random_string.storage.result}"
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "vnet-${random_id.rg.hex}"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "subnet" {
  name                 = "subnet-${random_id.rg.hex}"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_log_analytics_workspace" "law" {
  name                = "law-${random_id.rg.hex}"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
}
