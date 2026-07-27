output "resource_group_name" {
  value = azurerm_resource_group.rg.name
}

output "container_registry_name" {
  value = azurerm_container_registry.acr.name
}

output "login_server" {
  value = azurerm_container_registry.acr.login_server
}

output "storage_account_name" {
  value = azurerm_storage_account.storage.name
}

output "virtual_network_name" {
  value = azurerm_virtual_network.vnet.name
}

output "log_analytics_workspace_name" {
  value = azurerm_log_analytics_workspace.law.name
}
