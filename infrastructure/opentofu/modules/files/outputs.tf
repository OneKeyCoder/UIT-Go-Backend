output "storage_account_name" {
  description = "Name of the observability storage account"
  value       = azurerm_storage_account.observability.name
}

output "storage_account_id" {
  description = "ID of the observability storage account"
  value       = azurerm_storage_account.observability.id
}

# Data volume share names
output "prometheus_share_name" {
  description = "Name of the Prometheus data share"
  value       = azurerm_storage_share.prometheus.name
}

output "alertmanager_share_name" {
  description = "Name of the alertmanager data share"
  value       = azurerm_storage_share.alertmanager.name
}

output "grafana_share_name" {
  description = "Name of the Grafana data share"
  value       = azurerm_storage_share.grafana.name
}

output "loki_share_name" {
  description = "Name of the Loki data share"
  value       = azurerm_storage_share.loki.name
}

output "tempo_share_name" {
  description = "Name of the Jaeger data share"
  value       = azurerm_storage_share.tempo.name
}

# Configuration share name
output "configs_share_name" {
  description = "Name of the observability configs share"
  value       = azurerm_storage_share.configs.name
}

# Storage key secret
output "storage_key_secret_id" {
  description = "Key Vault secret ID for storage account access key"
  value       = azurerm_key_vault_secret.storage_key.versionless_id
  sensitive   = true
}
output "storage_account_key" {
  description = "Primary access key for the storage account"
  value       = azurerm_storage_account.observability.primary_access_key
  sensitive   = true
}
