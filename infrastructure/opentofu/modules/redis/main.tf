# random id for unique name
resource "random_id" "random_keyvault_suffix" {
  byte_length = 6
}

resource "azurerm_managed_redis" "redis" {
  name = "${var.resource_prefix}-${random_id.random_keyvault_suffix.hex}-redis"
  resource_group_name = var.resource_group_name
  location = var.location
  sku_name = var.sku_name
  high_availability_enabled = var.high_availability_enabled
  public_network_access = "Disabled"
  default_database {
    eviction_policy = var.eviction_policy
    clustering_policy = "EnterpriseCluster"
    client_protocol = "Plaintext"
    access_keys_authentication_enabled = true
  }
  tags = var.tags
}

resource "azurerm_private_endpoint" "endpoint" {
  name = "${var.resource_prefix}-redis-endpoint"
  resource_group_name = var.resource_group_name
  location = var.location
  subnet_id = var.endpoint_subnet_id

  private_service_connection {
    name = "${var.resource_prefix}-con"
    private_connection_resource_id = azurerm_managed_redis.redis.id
    is_manual_connection = false
    subresource_names = ["redisEnterprise"]
  }
  private_dns_zone_group {
    name = "${var.resource_prefix}-dns-zone"
    private_dns_zone_ids = [var.endpoint_dns_zone_id]
  }
}

# add access keys into keyvault
#! WARNING: this will SAVE YOUR VALUES TO STATEFILE.
# But Terraform is too dumb to hash this before saving to state for later compare, so what can we do?
# why default_database[0]? idk ask whoever made this weird ass api. it's not the database's index btw.
resource "azurerm_key_vault_secret" "primary" {
  name = "${var.resource_prefix}-redis-primary-key"
  key_vault_id = var.key_vault_id
  value = azurerm_managed_redis.redis.default_database[0].primary_access_key
}
resource "azurerm_key_vault_secret" "secondary" {
  name = "${var.resource_prefix}-redis-secondary-key"
  key_vault_id = var.key_vault_id
  value = azurerm_managed_redis.redis.default_database[0].secondary_access_key
}
resource "azurerm_key_vault_secret" "port" {
  name = "${var.resource_prefix}-redis-port"
  key_vault_id = var.key_vault_id
  value = azurerm_managed_redis.redis.default_database[0].port
}