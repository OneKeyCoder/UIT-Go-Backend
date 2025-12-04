output "postgres-subnet-id" {
  value = azurerm_subnet.postgres.id
}

output "aca-subnet-id" {
  value = azurerm_subnet.aca.id
}

output "dmz-subnet-id" {
  value = azurerm_subnet.dmz.id
}

output "main-vnet-id" {
  value = azurerm_virtual_network.main.id
}

output "endpoints-subnet" {
  value = azurerm_subnet.endpoints
}