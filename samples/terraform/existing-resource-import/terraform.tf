

terraform {
  required_version = ">= 1.5.0"
  required_providers {

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.9.0"
    }
    modtm = {
      source  = "azure/modtm"
      version = "~> 0.3"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "~> 2.0"
    }
  }
}


provider "azurerm" {
  features {}
}
