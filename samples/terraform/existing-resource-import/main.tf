resource "random_id" "rg" {
  byte_length = 8
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = "eastus2"
}

# create log analytics workspace
resource "azurerm_log_analytics_workspace" "this" {
  location            = azurerm_resource_group.rg.location
  name                = "law${random_id.rg.hex}"
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = "PerGB2018"
  retention_in_days   = var.retention_in_days
  tags                = var.tags
}

# resource "azurerm_application_insights" "this" {
#   application_type                      = "web"
#   location                              = azurerm_resource_group.rg.location
#   name                                  = "ai-${random_id.rg.hex}"
#   resource_group_name                   = azurerm_resource_group.rg.name
#   workspace_id                          = azurerm_log_analytics_workspace.this.id
# }

module "application_insights" {
  source = "Azure/avm-res-insights-component/azurerm"
  # source = "github.com/Azure/terraform-azurerm-avm-res-insights-component"
  version                       = "0.1.5"
  resource_group_name           = azurerm_resource_group.rg.name
  workspace_id                  = azurerm_log_analytics_workspace.this.id
  name                          = "ai-${random_id.rg.hex}"
  location                      = azurerm_resource_group.rg.location
  local_authentication_disabled = false
  internet_ingestion_enabled    = false
  internet_query_enabled        = false
  tags                          = var.tags
  enable_telemetry              = true
}

# resource "azapi_resource" "appinsights" {
#   type      = "Microsoft.Insights/components@2020-02-02"
#   name      = "ai-${random_id.rg.hex}"
#   parent_id = azurerm_resource_group.rg.id
#   location  = azurerm_resource_group.rg.location

#   body = {
#     kind = "web"
#     properties = {
#       Application_Type                = "web"
#       Flow_Type                       = "Bluefield"
#       Request_Source                  = "rest"
#       IngestionMode                   = "LogAnalytics"
#       WorkspaceResourceId             = azurerm_log_analytics_workspace.this.id
#     }
#   }

#   response_export_values = [
#     "id",
#     "properties.ConnectionString",
#   ]
# }

