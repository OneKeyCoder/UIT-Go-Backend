# main.tf – containing the resource blocks that define the resources to be created.

module "resource-group" {
  source = "./modules/resource-group"

  resource_group_base_name = var.resource_prefix
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
  name = var.acr_name
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
  admin_password_version = var.postgres_admin_password_version
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

  acr_pull_identity_id = module.acr.acr_pull_identity_id
}

locals {
  api_internal_hostname = "api-gateway"
  monitor_internal_hostname = "grafana"
}

module "app-gw" {
  source = "./modules/app-gw"
  
  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  subnet_id = module.networking.dmz-subnet-id
  base_hostname = var.base_hostname

  api_aca_fqdn = "${local.api_internal_hostname}.${module.aca-infra.env-fdns}"
  monitor_aca_fqdn = "${local.monitor_internal_hostname}.${module.aca-infra.env-fdns}"

  pfx_ssl_filename = var.pfx_ssl_filename
  pfx_ssl_password = var.pfx_ssl_password
}

module "location-redis" {
  source = "./modules/redis"

  resource_prefix = "${var.resource_prefix}-location"
  resource_group_name = local.rg_name
  location = local.rg_location

  endpoint_subnet_id = module.networking.endpoints-subnet-id
  endpoint_dns_zone_id = module.networking.redis-dns-zone-id

  key_vault_id = module.key_vault.id
  tags = {
    "service": "location"
  }
}

module "documentdb" {
  source = "./modules/documentdb"

  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  admin_username = var.documentdb_admin_username
  admin_password = var.documentdb_admin_password
  storage_size_in_gb = 32
}

module "files-mount" {
  source = "./modules/files"

  resource_prefix = var.resource_prefix
  resource_group_name = local.rg_name
  location = local.rg_location

  allowed_subnet_ids = [module.networking.aca-subnet-id]
  key_vault_id = module.key_vault.id
}
