output "kv_fdns" {
  value = azurerm_key_vault.keyvault.vault_uri
}

output "id" {
  value = azurerm_key_vault.keyvault.id
}

output "access_identity_id" {
  value = azurerm_user_assigned_identity.apps.id
}

output "jwt_secret_id" {
  description = "Key Vault secret ID (versionless) for JWT secret"
  value       = azurerm_key_vault_secret.jwt_secret.versionless_id
}

output "here_id_secret_id" {
  description = "Key Vault secret ID (versionless) for HERE Maps API ID"
  value       = azurerm_key_vault_secret.here_id.versionless_id
}

output "here_secret_secret_id" {
  description = "Key Vault secret ID (versionless) for HERE Maps API secret"
  value       = azurerm_key_vault_secret.here_secret.versionless_id
}