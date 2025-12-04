# main.tf – containing the resource blocks that define the resources to be created.

module "resource-group" {
  source = "./modules/resource-group"

  resource_group_base_name = var.resource_prefix
  environment = var.environment
  location = var.location
}

locals {
  rg_name = module.resource-group.name
  rg_location = module.resource-group.location
}

module "acr" {
  source = "./modules/acr"

  resource_group_name = local.rg_name
  location = local.rg_location
  resource_prefix = var.resource_prefix
}

# First time deploy will stop here. Run CD pipeline once then continue the deploy.

module "networking" {
  source = "./modules/networking"

  resource_group_name = local.rg_name
  location = local.rg_location
  resource_prefix = var.resource_prefix
}

module "postgres" {
  source = "./modules/postgres"

  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  virtual_network_id = module.networking.main-vnet-id
  subnet_id = module.networking.postgres-subnet-id

  # TODO: swap this out for a randomized per-db password.
  admin_username = var.postgres_admin_username
  admin_password = var.postgres_admin_password
}

module "key_vault" {
  source = "./modules/key-vault"

  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  allowed_subnet_ids = [
    module.networking.aca-subnet-id,
  ]
}

module "aca-infra" {
  source = "./modules/aca-infra"

  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  vnet_id = module.networking.main-vnet-id
  subnet_id = module.networking.aca-subnet-id
}

