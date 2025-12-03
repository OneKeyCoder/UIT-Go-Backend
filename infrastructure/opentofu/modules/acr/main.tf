resource "azurerm_container_registry" "acr" {
  name = "${var.resource_prefix}-acr"
  location = var.location
  resource_group_name = var.resource_group_name
  sku = var.sku
  public_network_access_enabled = true
}