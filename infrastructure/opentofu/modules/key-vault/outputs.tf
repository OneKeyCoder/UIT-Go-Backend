output "kv_fdns" {
  value = azurerm_key_vault.keyvault.vault_uri
}

output "id" {
  value = azurerm_key_vault.keyvault.id
}

output "access-identity" {
  value = azurerm_user_assigned_identity.apps.id
}