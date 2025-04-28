






module "law" {
  source                       = "./modules/law"
  log_analytics_workspace_name = var.log_analytics_workspace_name
  tags                         = var.tags
}

module "law2" {
  source                       = "./modules/law"
  log_analytics_workspace_name = "${var.log_analytics_workspace_name}2"
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