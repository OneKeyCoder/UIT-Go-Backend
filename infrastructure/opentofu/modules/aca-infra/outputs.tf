output "env-fdns" {
  value = azurerm_private_dns_zone.aca.name
}

output "env-id" {
  value = azurerm_container_app_environment.env.id
}