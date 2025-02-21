

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY AN AZURE CONTAINER Instance
# This is an example of how to deploy an Azure Container Instance
# See test/terraform_azure_aci_example_test.go for how to write automated tests for this code.
# ---------------------------------------------------------------------------------------------------------------------

# ------------------------------------------------------------------------------
# CONFIGURE OUR AZURE CONNECTION
# ------------------------------------------------------------------------------

terraform {
  #   required_providers {
  #     azurerm = {
  #       version = "~>2.29.0"
  #       source  = "hashicorp/azurerm"
  #     }
  #   }

}

provider "azurerm" {
  features {}
  skip_provider_registration = "true"
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY A RESOURCE GROUP
# ---------------------------------------------------------------------------------------------------------------------

# data "azurerm_resource_group" "rg" {
#   name     = var.resource_group_name
# }



# invalid resource group name - extemely long
resourc "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = var.location
}
