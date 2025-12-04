output "namespace_id" {
  description = "ID of the Service Bus namespace"
  value       = azurerm_servicebus_namespace.main.id
}

output "namespace_name" {
  description = "Name of the Service Bus namespace"
  value       = azurerm_servicebus_namespace.main.name
}

output "logs_queue_name" {
  description = "Name of the logs queue"
  value       = azurerm_servicebus_queue.logs.name
}

output "connection_string_secret_id" {
  description = "Key Vault secret ID (versionless) for Service Bus connection string"
  value       = azurerm_key_vault_secret.servicebus_connection_string.versionless_id
}
