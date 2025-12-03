resource "azurerm_public_ip" "public_ip" {
  name = "${var.resource_prefix}-pip"
  resource_group_name = var.resource_group_name
  location = var.location
  allocation_method = "Static"
}

locals {
  frontend_ip_config = "fe-pip"
  https_port = "https-port"
  ssl_cert = "ssl-cert"
  api_hostname = (var.api_subdomain == "" || var.api_subdomain == null
    ? var.base_host_name
    : "${var.api_subdomain}.${var.base_host_name}")
  monitor_hostname = (var.grafana_subdomain == "" || var.grafana_subdomain == null
    ? var.base_host_name
    : "${var.grafana_subdomain}.${var.base_host_name}")
}

resource "azurerm_application_gateway" "app_gw" {
  name = "${var.resource_prefix}-app-gw"
  resource_group_name = var.resource_group_name
  location = var.location

  sku {
    name = "Standard_v2"
    tier = "Standard_v2"
  }
  autoscale_configuration {
    min_capacity = 0
    max_capacity = 10
  }

  gateway_ip_configuration {
    name = "gw-ip-config"
    subnet_id = var.subnet_id
  }
  frontend_ip_configuration {
    name = local.frontend_ip_config
    public_ip_address_id = azurerm_public_ip.public_ip.id
  }
  frontend_port {
    name = local.https_port
    port = 443
  }

  # Default backend http setting
  backend_http_settings {
    name = "empty-http-setting"
    protocol = "Http"
    port = 80
    cookie_based_affinity = "Disabled"
    dedicated_backend_connection_enabled = false
    request_timeout = var.request_timeout
  }

  # ssl certs for frontend https
  ssl_certificate {
    name = local.ssl_cert
    data = filebase64("${path.root}/certs/${var.pfx_ssl_filename}")
    password = var.pfx_ssl_password
  }

  # api endpoint
  backend_address_pool {
    name = "api-pool"
    fqdns = [var.api_aca_fqdn]
  }
  http_listener {
    name = "api-listener"
    frontend_ip_configuration_name = local.frontend_ip_config
    protocol = "Https"
    frontend_port_name = local.https_port
    ssl_certificate_name = local.ssl_cert
    host_name = local.api_hostname
  }
  request_routing_rule {
    name = "api-routing"
    rule_type = "Basic"
    http_listener_name = "api-listener"
  }

  # monitor endpoint
  backend_address_pool {
    name = "grafana-pool"
    fqdns = [var.grafana_aca_fqdn]
  }
  http_listener {
    name = "monitor-listener"
    frontend_ip_configuration_name = local.frontend_ip_config
    protocol = "Https"
    frontend_port_name = local.https_port
    ssl_certificate_name = local.ssl_cert
    host_name = local.monitor_hostname
  }
  request_routing_rule {
    name = "monitor-routing"
    rule_type = "Basic"
    http_listener_name = "monitor-listener"
  }
}