resource "azurerm_network_security_group" "aca_nsg" {
  name = "aca-nsg"
  location = var.location
  resource_group_name = var.resource_group_name

  # allow traffic from dmz to gateway
  security_rule {
    name = "AllowDmzInbound"
    priority = 100
    direction = "Inbound"
    access = "Allow"
    protocol = "Tcp"
    source_port_range = "*"
    source_address_prefixes = azurerm_subnet.dmz.address_prefixes
    destination_port_ranges = ["80", "443"]
    destination_address_prefix = "*"
  }
}
resource "azurerm_subnet_network_security_group_association" "aca_nsg_ass" {
  subnet_id = azurerm_subnet.aca.id
  network_security_group_id = azurerm_network_security_group.aca_nsg.id
}