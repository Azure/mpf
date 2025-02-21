


terraform {

}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
  skip_provider_registration = "true"
}


resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.location
}


# Generate random postfix to mitigate naming collisions
resource "random_id" "randomId" {
  keepers = {
    # Generate a new ID only when a new resource group is created
    resource_group = azurerm_resource_group.rg.name
  }

  byte_length = 8
}

# create vnet, subnet and aks in private subnet
resource "azurerm_virtual_network" "vnet" {
  name                = "vnet-${random_id.randomId.hex}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  address_space       = ["10.12.0.0/16"]
}

resource "azurerm_subnet" "subnet" {
  name                 = "subnet-${random_id.randomId.hex}"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = ["10.12.1.0/24"]
}

# create aks cluster in subnet
resource "azurerm_kubernetes_cluster" "aks" {
  name                = "aks-${random_id.randomId.hex}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  dns_prefix          = "aks-${random_id.randomId.hex}"


  identity {
    type = "SystemAssigned"
  }

  default_node_pool {
    name           = "default"
    node_count     = 1
    vm_size        = "Standard_B2s"
    vnet_subnet_id = azurerm_subnet.subnet.id
  }

  network_profile {
    network_plugin = "azure"
  }

}
