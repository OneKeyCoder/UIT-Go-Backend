# Subnet:
# https://www.davidc.net/sites/default/subnets/subnets.html?network=192.168.0.0&mask=16&division=25.fb97000

resource "azurerm_virtual_network" "main" {
  name = "${var.resource_prefix}-vnet"
  resource_group_name = var.resource_group_name
  location = var.location
  address_space = ["10.0.0.0/16"] # 65k hosts
}

resource "azurerm_subnet" "aca" {
  name = "subnet-aca"
  resource_group_name = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  default_outbound_access_enabled = true
  address_prefixes = ["10.0.0.0/22"] # 1019 hosts (network and broadcast ip, and azure reserves additional 3 ips)

  delegation {
    name = "aca-delegation"
    service_delegation {
      name = "Microsoft.App/environments"
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
    }
  }
}

resource "azurerm_subnet" "dmz" {
  name = "subnet-dmz"
  resource_group_name = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes = ["10.0.4.0/24"] # 251
}

resource "azurerm_subnet" "postgres" {
  name = "subnet-postgres"
  resource_group_name = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes = ["10.0.5.0/24"] # 251

  delegation {
    name = "postgres-delegation"
    service_delegation {
      name = "Microsoft.DBforPostgreSQL/flexibleServers"
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
    }
  }
}

# special subnet for private endpoints
resource "azurerm_subnet" "endpoints" {
  name                 = "subnet-endpoints"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.6.0/27"]
  
  private_endpoint_network_policies = "Disabled" 
}
# Key Vault Private Links
resource "azurerm_private_dns_zone" "kv" {
  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = var.resource_group_name
}
resource "azurerm_private_dns_zone_virtual_network_link" "kv_link" {
  name                  = "kv-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.kv.name
  virtual_network_id    = azurerm_virtual_network.main.id
}

# acls
resource "azurerm_network_security_group" "aca_nsg" {
  name = "nsg-aca"
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