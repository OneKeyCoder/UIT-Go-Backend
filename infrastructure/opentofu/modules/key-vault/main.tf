data "azurerm_client_config" "current" {}

# Role for ACA to access keyvaults
resource "azurerm_user_assigned_identity" "apps" {
  name                = "${var.resource_prefix}-aca-id"
  location            = var.location
  resource_group_name = var.resource_group_name
}

# random id
resource "random_id" "random_keyvault_suffix" {
  byte_length = 6
}

resource "azurerm_key_vault" "keyvault" {
  name     = "${var.resource_prefix}-kv-${random_id.random_keyvault_suffix.hex}"
  sku_name = "standard"

  location            = var.location
  resource_group_name = var.resource_group_name
  tenant_id           = data.azurerm_client_config.current.tenant_id

  public_network_access_enabled = true
  rbac_authorization_enabled = true

  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  # only allow service endpoints
  network_acls {
    bypass = "AzureServices"
    default_action = "Deny"
    virtual_network_subnet_ids = var.allowed_subnet_ids
  }
}

# Grant Terraform service principal permissions to manage secrets
resource "azurerm_role_assignment" "terraform_kv_admin" {
  scope                = azurerm_key_vault.keyvault.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = data.azurerm_client_config.current.object_id
}

# Grant ACA managed identity permission to read secrets
resource "azurerm_role_assignment" "aca_kv_reader" {
  scope                = azurerm_key_vault.keyvault.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.apps.principal_id
}