terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

# This sample is intentionally minimal and is used to exercise the
# Authorization_RequestDenied error path in MPF. Creating an Azure AD group
# requires Microsoft Graph application permissions (e.g. Group.Create) that
# require admin consent or Global Administrator role; these cannot be
# auto-discovered by MPF, so MPF should surface a clear guidance error.
resource "random_string" "rand" {
  length  = 8
  special = false
  numeric = false
  upper   = false
  lower   = true
}

resource "azuread_group" "mpf_test" {
  display_name     = "mpf-authreqdenied-${random_string.rand.result}"
  security_enabled = true
}
