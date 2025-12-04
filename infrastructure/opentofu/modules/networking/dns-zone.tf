# redis
resource "azurerm_private_dns_zone" "redis" {
  name                = "privatelink.redis.azure.net"
  resource_group_name = var.resource_group_name
}
resource "azurerm_private_dns_zone_virtual_network_link" "dns_link" {
  name                  = "vnet-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.redis.name
  virtual_network_id    = azurerm_virtual_network.main.id
}
