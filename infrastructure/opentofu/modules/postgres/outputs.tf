output "server_fqdn" {
  description = "FQDN of the PostgreSQL server"
  value       = azurerm_postgresql_flexible_server.postgres.fqdn
}

output "server_id" {
  description = "ID of the PostgreSQL server"
  value       = azurerm_postgresql_flexible_server.postgres.id
}

output "connection_string_secret_id" {
  description = "Key Vault secret ID (versionless) for PostgreSQL connection string"
  value       = azurerm_key_vault_secret.postgres_connection_string.versionless_id
}
