output "connection_strings" {
  description = "All connection strings from MongoDB cluster"
  value       = azurerm_mongo_cluster.main.connection_strings
  sensitive   = true
}

output "primary_connection_string" {
  description = "Primary MongoDB connection string value"
  value       = length(azurerm_mongo_cluster.main.connection_strings) > 0 ? azurerm_mongo_cluster.main.connection_strings[0].value : ""
  sensitive   = true
}

output "connection_string_secret_id" {
  description = "Key Vault secret ID (versionless) for MongoDB connection string"
  value       = azurerm_key_vault_secret.mongo_connection_string.versionless_id
}

output "cluster_id" {
  description = "ID of the MongoDB cluster"
  value       = azurerm_mongo_cluster.main.id
}

output "cluster_name" {
  description = "Name of the MongoDB cluster"
  value       = azurerm_mongo_cluster.main.name
}