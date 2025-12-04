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
