resource "azurerm_postgresql_flexible_server" "auth-postgres" {
  location = var.location
  resource_group_name = var.resource_group_name
  name = "${var.resource_prefix}-"
}