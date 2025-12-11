resource "random_string" "storage_account_name" {
  length = 6
  lower = true
  numeric = true
  special = false
  upper = false
}

resource "azurerm_storage_account" "observability" {
  name                     = "${random_string.storage_account_name.result}obssa"
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  
  network_rules {
    default_action = "Deny"
    bypass         = ["AzureServices"]
    virtual_network_subnet_ids = var.allowed_subnet_ids
  }
}

resource "azurerm_storage_share" "prometheus" {
  name                 = "prometheus-data"
  storage_account_id   = azurerm_storage_account.observability.id
  quota                = 10 # 10 GB
}

resource "azurerm_storage_share" "grafana" {
  name                 = "grafana-data"
  storage_account_id   = azurerm_storage_account.observability.id
  quota                = 5
}

resource "azurerm_storage_share" "loki" {
  name                 = "loki-data"
  storage_account_id   = azurerm_storage_account.observability.id
  quota                = 20
}

resource "azurerm_storage_share" "tempo" {
  name                 = "tempo-data"
  storage_account_id   = azurerm_storage_account.observability.id
  quota                = 10
}

# Store access key in Key Vault
resource "azurerm_key_vault_secret" "storage_key" {
  name         = "${var.resource_prefix}-storage-key"
  key_vault_id = var.key_vault_id
  value        = azurerm_storage_account.observability.primary_access_key
  
  lifecycle {
    ignore_changes = [value]
  }
}