resource "azurerm_container_registry" "acr" {
  name = var.name
  location = var.location
  resource_group_name = var.resource_group_name
  sku = var.sku
  public_network_access_enabled = true
}