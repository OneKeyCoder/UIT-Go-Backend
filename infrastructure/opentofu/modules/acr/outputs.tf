output "id" {
  value = azurerm_container_registry.acr.id
}

output "name" {
  description = "Name of the Azure Container Registry"
  value       = azurerm_container_registry.acr.name
}

output "login_server_url" {
  description = "Login server URL for the ACR"
  value       = azurerm_container_registry.acr.login_server
}

output "acr_pull_identity_id" {
  description = "Resource ID of the managed identity with ACR pull permissions"
  value       = azurerm_user_assigned_identity.acr_pull.id
}

output "acr_pull_identity_client_id" {
  description = "Client ID of the managed identity with ACR pull permissions"
  value       = azurerm_user_assigned_identity.acr_pull.client_id
}

output "acr_pull_identity_principal_id" {
  description = "Principal ID of the managed identity with ACR pull permissions"
  value       = azurerm_user_assigned_identity.acr_pull.principal_id
}