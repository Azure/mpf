


terraform {}

provider "azurerm" {
  features {}
}

resource "random_id" "rg" {
  byte_length = 8
}

resource "random_integer" "rnd_num" {
  min = 10000
  max = 30000
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = var.location
}


resource "azurerm_container_group" "aci" {
  name                = "aci${random_integer.rnd_num.result}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_address_type = "Public"
  dns_name_label  = "aci${random_integer.rnd_num.result}"
  os_type         = "Linux"

  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "0.5"
    memory = "1.5"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }

  tags = {
    Environment = "Development"
  }
}
