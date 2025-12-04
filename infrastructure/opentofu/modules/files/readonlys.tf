# Configuration files share (read-only configs)
resource "azurerm_storage_share" "configs" {
  name                 = "observability-configs"
  storage_account_id   = azurerm_storage_account.observability.id
  quota                = 1 # Small quota for config files
}

# Upload Prometheus configuration files
resource "azurerm_storage_share_file" "prometheus_config" {
  name             = "prometheus.yml"
  storage_share_url = azurerm_storage_share.configs.url
  source           = "${path.root}/../../observability/prometheus.yml"
}

resource "azurerm_storage_share_file" "prometheus_alerts" {
  name             = "prometheus-alerts.yml"
  storage_share_url = azurerm_storage_share.configs.url
  source           = "${path.root}/../../observability/prometheus-alerts.yml"
}

# Upload Loki configuration
resource "azurerm_storage_share_file" "loki_config" {
  name             = "loki-config.yml"
  storage_share_url = azurerm_storage_share.configs.url
  source           = "${path.root}/../../observability/loki-config.yml"
}

# Upload Alloy configuration
resource "azurerm_storage_share_file" "alloy_config" {
  name             = "alloy-config.alloy"
  storage_share_url = azurerm_storage_share.configs.url
  source           = "${path.root}/../../observability/alloy-config.alloy"
}

# Upload Jaeger configuration
resource "azurerm_storage_share_file" "jaeger_config" {
  name             = "config-badger.yml"
  storage_share_url = azurerm_storage_share.configs.url
  source           = "${path.root}/../../observability/config-badger.yml"
}
