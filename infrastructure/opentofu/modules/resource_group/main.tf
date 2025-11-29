locals {
  # if var.env is not null or empty, append env to rg name
  rg_name = (
    var.environment != null && var.environment != ""
    ? "${var.resource_group_base_name}-${var.environment}"
    : var.resource_group_base_name
  )
}

resource "azurerm_resource_group" "project-rg" {
  name     = local.rg_name
  location = var.location
  tags = {
    "environment" = var.environment
  }
}
