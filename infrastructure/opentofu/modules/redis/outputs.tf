output "id" {
  value = azurerm_managed_redis.redis.id
}

output "hostname" {
  value = azurerm_managed_redis.redis.hostname
}

output "kv_primary_access_key_id" {
  value = azurerm_key_vault_secret.primary.resource_versionless_id
}

output "kv_secondary_access_key_id" {
  value = azurerm_key_vault_secret.secondary.resource_versionless_id
}

output "port" {
  value = azurerm_managed_redis.redis.default_database[0].port
}