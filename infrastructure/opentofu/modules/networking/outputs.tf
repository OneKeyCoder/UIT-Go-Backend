output "postgres-subnet" {
  value = azurerm_subnet.postgres-db
}

output "aca-subnet" {
  value = azurerm_subnet.aca
}

output "dmz-subnet" {
  value = azurerm_subnet.dmz
}

output "main-vnet" {
  value = azurerm_virtual_network.main
}

output "endpoints-subnet" {
  value = azurerm_subnet.endpoints
}