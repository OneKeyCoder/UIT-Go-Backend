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
}

# Grant Terraform service principal permissions to manage secrets
resource "azurerm_role_assignment" "terraform_kv_admin" {
  scope                = azurerm_key_vault.keyvault.id
  role_definition_name = "Key Vault Administrator"
  principal_id         = data.azurerm_client_config.current.object_id
  depends_on = [ azurerm_role_assignment.terraform_kv_officer ]
}
resource "azurerm_role_assignment" "terraform_kv_officer" {
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

# Application secrets - these should be rotated in production
resource "azurerm_key_vault_secret" "jwt_secret" {
  name         = "${var.resource_prefix}-jwt-secret"
  key_vault_id = azurerm_key_vault.keyvault.id
  value        = var.jwt_secret

  lifecycle {
    ignore_changes = [ 
      value
    ]
  }
  depends_on = [azurerm_role_assignment.terraform_kv_admin]
}

resource "azurerm_key_vault_secret" "here_id" {
  name         = "${var.resource_prefix}-here-id"
  key_vault_id = azurerm_key_vault.keyvault.id
  value        = var.here_id

  lifecycle {
    ignore_changes = [ 
      value
    ]
  }
  depends_on = [azurerm_role_assignment.terraform_kv_admin]
}

resource "azurerm_key_vault_secret" "here_secret" {
  name         = "${var.resource_prefix}-here-secret"
  key_vault_id = azurerm_key_vault.keyvault.id
  value        = var.here_secret

  lifecycle {
    ignore_changes = [ 
      value
    ]
  }
  depends_on = [azurerm_role_assignment.terraform_kv_admin]
}