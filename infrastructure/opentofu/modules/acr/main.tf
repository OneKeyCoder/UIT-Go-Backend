resource "azurerm_container_registry" "acr" {
  name = var.name
  location = var.location
  resource_group_name = var.resource_group_name
  sku = var.sku
  public_network_access_enabled = true
}

resource "azurerm_user_assigned_identity" "acr_pull" {
  name                = "${var.name}-acr-pull-identity"
  location            = var.location
  resource_group_name = var.resource_group_name
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.acr_pull.principal_id
}