output "name" {
  description = "Name of the Application Insights"
  # value       = azapi_resource.appinsights.name
  # value = azurerm_application_insights.this.name
  value = module.application_insights.name
}
