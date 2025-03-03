


terraform {

}

provider "azurerm" {
  features {}
  skip_provider_registration = "true"
  storage_use_azuread        = true
}

resource "random_id" "rg" {
  byte_length = 8
}
resource "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = "uksouth"
}

resource "random_string" "rand" {
  length  = 8
  special = false
  numeric = false
  upper   = false
  lower   = true
}

data "azurerm_client_config" "current" {
}

variable "location" {
  type    = string
  default = "eastus"
}

resource "azurerm_storage_account" "st" {
  name                          = "saapermmismatch${random_string.rand.result}"
  resource_group_name           = azurerm_resource_group.rg.name
  location                      = azurerm_resource_group.rg.location
  account_kind                  = "Storage"
  account_tier                  = "Standard"
  account_replication_type      = "LRS"
  public_network_access_enabled = false
}