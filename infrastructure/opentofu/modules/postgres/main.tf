# Private DNS Zone
resource "azurerm_private_dns_zone" "postgres" {
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = var.resource_group_name
}

resource "azurerm_private_dns_zone_virtual_network_link" "postgres_link" {
  name                  = "postgres-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.postgres.name
  virtual_network_id    = var.virtual_network_id
}

# Postgres server
resource "azurerm_postgresql_flexible_server" "postgres" {
  name = "${var.resource_prefix}-psql-primary"
  resource_group_name = var.resource_group_name
  location = var.location
  version = "16"

  delegated_subnet_id = var.subnet_id
  private_dns_zone_id = azurerm_private_dns_zone.postgres.id
  public_network_access_enabled = false

  administrator_login    = var.admin_username
  administrator_password_wo = var.admin_password
  administrator_password_wo_version = var.admin_password_version

  storage_mb = 131072 # 128GB
  storage_tier = "P30"
  auto_grow_enabled = true

  # Use "General Purpose" (D-series) for production consistency. 
  # Use "Burstable" (B-series) only if cost is the #1 priority.
  sku_name = var.sku_name
}

# Configuration: Enable PgBouncer
resource "azurerm_postgresql_flexible_server_configuration" "pgbouncer" {
  name      = "pgbouncer.enabled"
  server_id = azurerm_postgresql_flexible_server.postgres.id
  value     = "true"
}

resource "azurerm_postgresql_flexible_server_configuration" "pooling_mode" {
  name      = "pgbouncer.pooling_mode"
  server_id = azurerm_postgresql_flexible_server.postgres.id
  value     = "transaction" # 'transaction' is best for microservices
}

# Read Replica (Optional)
# If the Auth Service reads are slowing down the Trip Service, deploy this.
# resource "azurerm_postgresql_flexible_server" "replica" {
#   name                   = "${var.prefix}-psql-replica"
#   resource_group_name    = var.rg_name
#   location               = var.location # Can be same or different region
#   version                = "16"
#   create_mode            = "Replica"
#   source_server_id       = azurerm_postgresql_flexible_server.main.id
#   delegated_subnet_id    = var.subnet_id
#   private_dns_zone_id    = var.private_dns_zone_id
  
#   sku_name = "GP_Standard_D2ds_v5" # Usually same size as primary
# }

# Logical Databases
resource "azurerm_postgresql_flexible_server_database" "dbs" {
  for_each = var.db_names
  name      = each.key
  server_id = azurerm_postgresql_flexible_server.postgres.id
  collation = "en_US.utf8"
  charset   = "utf8"
}

# Store PostgreSQL connection string in Key Vault
resource "azurerm_key_vault_secret" "postgres_connection_string" {
  name         = "${var.resource_prefix}-postgres-connection-string"
  key_vault_id = var.key_vault_id
  value        = "host=${azurerm_postgresql_flexible_server.postgres.fqdn} port=5432 user=${var.admin_username} password=${var.admin_password} dbname=users sslmode=require timezone=UTC connect_timeout=5"
  
  depends_on = [azurerm_postgresql_flexible_server_database.dbs]
}
