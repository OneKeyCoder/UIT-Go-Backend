provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
  }
}

data "azurerm_client_config" "current" {}

# Role for ACA to access keyvaults
resource "azurerm_user_assigned_identity" "apps" {
  name                = "${var.resource_prefix}-aca-id"
  location            = var.location
  resource_group_name = var.resource_group_name
}

# random id
resource "random_id" "random_keyvault_suffix" {
  byte_length = 16
}

resource "azurerm_key_vault" "keyvault" {
  name     = "${var.resource_prefix}-kv-${random_id.random_keyvault_suffix.hex}"
  sku_name = "standard"

  location            = var.location
  resource_group_name = var.resource_group_name
  tenant_id           = data.azurerm_client_config.current.tenant_id

  public_network_access_enabled = false
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  rbac_authorization_enabled = true
}

resource "azurerm_private_endpoint" "name" {
  subnet_id = var.private_endpoint_subnet_id

  name = "kv-private-endpoint"
  location = var.location
  resource_group_name = var.resource_group_name

  private_service_connection {
    name = "kv-connection"
    private_connection_resource_id = azurerm_key_vault.keyvault.id
    subresource_names = ["vault"]
    is_manual_connection = false
  }

  private_dns_zone_group {
    name = "default"
    private_dns_zone_ids = [var.private_dns_zone_id]
  }
}

resource "azurerm_key_vault_key" "a" {
  
}