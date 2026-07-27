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
