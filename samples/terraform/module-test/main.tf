


resource "random_id" "rg" {
  byte_length = 8
}

resource "random_integer" "rnd_num" {
  min = 10000
  max = 30000
}

resource "azurerm_resource_group" "this" {
  location = "East US2" # Location used just for the example
  name     = "rg-${random_id.rg.hex}"
}



module "law" {
  source                       = "./modules/law"
  location                     = azurerm_resource_group.this.location
  log_analytics_workspace_name = "lawtftest${random_integer.rnd_num.result}"
  resource_group_name          = azurerm_resource_group.this.name
  tags                         = var.tags
}

terraform {
  required_version = ">= 1.9.6, < 2.0.0"
  required_providers {

    azuread = {
      source  = "hashicorp/azuread"
      version = ">= 2.53, < 3.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.114.0, < 4.0.0"
    }
    # tflint-ignore: terraform_unused_required_providers
    modtm = {
      source  = "Azure/modtm"
      version = "~> 0.3"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }
}

provider "azurerm" {
  features {}
  skip_provider_registration = "true"
}