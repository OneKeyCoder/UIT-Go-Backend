resource "azurerm_application_gateway" "app_gw" {
  name = "app-gw"
  resource_group_name = var.resource_group_name
  location = var.location
  sku {
    name = "Standard_v2"
    tier = "Standard_v2"
  }
  autoscale_configuration {
    min_capacity = 0
    max_capacity = 50
  }
  backend_address_pool {

  }
  backend_http_settings {

  }
  frontend_ip_configuration {

  }
  frontend_port {

  }
  gateway_ip_configuration {

  }
  http_listener {

  }
  request_routing_rule {

  }
}