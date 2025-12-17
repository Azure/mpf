terraform {
  required_version = ">= 1.5.0"
  required_providers {

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.56"
    }
    modtm = {
      source  = "azure/modtm"
      version = "~> 0.3"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
    azapi = {
      source  = "azure/azapi"
      version = "~> 2.8"
    }
  }
}


provider "azurerm" {
  features {}
}
