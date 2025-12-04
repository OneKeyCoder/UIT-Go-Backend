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

output "endpoints-subnet-id" {
  value = azurerm_subnet.endpoints.id
}

# dns zones
output "redis-dns-zone-id" {
  value = azurerm_private_dns_zone.redis.id
}