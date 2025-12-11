resource "azurerm_container_app_environment" "env" {
  name                           = "${var.resource_prefix}-aca-env"
  location                       = var.location
  resource_group_name            = var.resource_group_name

  infrastructure_subnet_id       = var.subnet_id
  internal_load_balancer_enabled = true
  zone_redundancy_enabled = true

  logs_destination = "azure-monitor"
  # logs_destination = "log-analytics"
  # log_analytics_workspace_id = TODO

  public_network_access = "Disabled"
  
  workload_profile {
    name                  = "Consumption"
    workload_profile_type = "Consumption"
  }

  identity {
    type = "SystemAssigned, UserAssigned"
    identity_ids = [var.acr_pull_identity_id]
  }

  lifecycle {
    ignore_changes = [
      infrastructure_resource_group_name,
    ]
  }
}

# Private DNS Zone for App gateway
resource "azurerm_private_dns_zone" "aca" {
  name                = azurerm_container_app_environment.env.default_domain
  resource_group_name = var.resource_group_name
}

resource "azurerm_private_dns_zone_virtual_network_link" "aca_link" {
  name                  = "aca-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.aca.name
  virtual_network_id    = var.vnet_id
}

resource "azurerm_private_dns_a_record" "wildcard" {
  name                = "*"
  zone_name           = azurerm_private_dns_zone.aca.name
  resource_group_name = var.resource_group_name
  ttl                 = 300
  records             = [azurerm_container_app_environment.env.static_ip_address]
}

resource "azurerm_private_dns_a_record" "apex" {
  name                = "@"
  zone_name           = azurerm_private_dns_zone.aca.name
  resource_group_name = var.resource_group_name
  ttl                 = 300
  records             = [azurerm_container_app_environment.env.static_ip_address]
}