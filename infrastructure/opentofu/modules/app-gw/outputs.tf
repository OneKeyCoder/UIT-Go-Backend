output "public_ip_address" {
  value = azurerm_public_ip.public_ip.ip_address
}

output "api_public_hostname" {
  value = local.api_hostname
}

output "monitor_public_hostname" {
  value = local.monitor_hostname
}