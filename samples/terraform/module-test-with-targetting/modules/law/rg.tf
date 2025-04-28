resource "random_id" "rg" {
  byte_length = 8
}

resource "azurerm_resource_group" "this" {
  location = "East US2" # Location used just for the example
  name     = "rg-${random_id.rg.hex}"
}