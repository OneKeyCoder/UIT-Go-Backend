resource "azurerm_mongo_cluster" "main" {
  name                   = "${var.resource_prefix}-mongo"
  resource_group_name    = var.resource_group_name
  location               = var.location
  administrator_username = var.admin_username
  administrator_password = var.admin_password
  shard_count            = "1"
  compute_tier           = var.compute_tier
  high_availability_mode = "Disabled"
  storage_size_in_gb     = var.storage_size_in_gb
  version                = "7.0"
  public_network_access = "Enabled"
  tags = var.tags
}

# special IP range 0.0.0.0 to 0.0.0.0 is the Azure convention for allowing access
# from other Azure services within the same region or subscription.
resource "azurerm_mongo_cluster_firewall_rule" "azure_allow" {
  name = "AllowAzureServices"
  mongo_cluster_id = azurerm_mongo_cluster.main.id
  start_ip_address = "0.0.0.0"
  end_ip_address = "0.0.0.0"
}