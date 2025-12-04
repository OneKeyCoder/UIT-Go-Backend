resource "azurerm_servicebus_namespace" "main" {
  name = "${var.resource_prefix}-bus"

  location = var.location
  resource_group_name = var.resource_group_name

  sku = "Standard"

  local_auth_enabled = true
  public_network_access_enabled = true

  network_rule_set {
    default_action = "Deny"
    network_rules {
      subnet_id = var.aca_subnet_id
    }
  }
}

# Create the logs queue for log ingestion
resource "azurerm_servicebus_queue" "logs" {
  name         = "logs"
  namespace_id = azurerm_servicebus_namespace.main.id

  partitioning_enabled = false
  
  # Message TTL: 7 days
  default_message_ttl = "P7D"
  
  # Lock duration for message processing
  lock_duration = "PT1M"
  
  # Max delivery count before dead-lettering
  max_delivery_count = 10
  
  # Enable dead lettering on message expiration
  dead_lettering_on_message_expiration = true
}

# Get the primary connection string
data "azurerm_servicebus_namespace_authorization_rule" "default" {
  name         = "RootManageSharedAccessKey"
  namespace_id = azurerm_servicebus_namespace.main.id
}

# Format: amqp://rule-name:key-value@hostname:5671/?sasl=plain
locals {
  key_name = urlencode(data.azurerm_servicebus_namespace_authorization_rule.default.name)
  key_value = urlencode(data.azurerm_servicebus_namespace_authorization_rule.default.primary_key)
  amqp_connection_string = "amqps://${local.key_name}:${local.key_value}@${azurerm_servicebus_namespace.main.name}.servicebus.windows.net:5671/?sasl=plain&verify=verify_none"
}

# Store AMQP Service Bus connection string in Key Vault
resource "azurerm_key_vault_secret" "servicebus_connection_string" {
  name         = "${var.resource_prefix}-servicebus-connection-string"
  key_vault_id = var.key_vault_id
  value        = local.amqp_connection_string
  
  depends_on = [azurerm_servicebus_namespace.main]

  lifecycle {
    ignore_changes = [ value ]
  }
}
