locals {
  go_service_names = [
    "api-gateway",
    "authentication-service",
    "location-service",
    "logger-service",
    "trip-service",
    "user-service",
  ]
}

locals {
  acr_login_server = module.acr.login_server_url
  acr_pull_identity_id = module.acr.acr_pull_identity_id
  otel_envs = {
    OTEL_EXPORTER: "otlp"
    OTEL_COLLECTOR_ENDPOINT: "alloy:4317"
    OTEL_INSECURE: "true"
  }
}

module "api_gateway" {
  source = "./modules/aca-service"
  name = "api-gateway"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id
  key_vault_access_identity_id = module.key_vault.access_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = true
  target_port_ingress = 80
  envs = local.otel_envs
  min_replica = 1
}

module "logger_service" {
  source = "./modules/aca-service"
  name = "logger-service"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id
  key_vault_access_identity_id = module.key_vault.access_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = false
  target_port_ingress = 50052
  envs = local.otel_envs

  secrets = {
    "MONGO_CONNECTION_STRING": module.documentdb.connection_string_secret_id
    "RABBITMQ_CONNECTION_STRING": module.service-bus.connection_string_secret_id
  }
}

module "authentication_service" {
  source = "./modules/aca-service"
  name = "authentication-service"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id
  key_vault_access_identity_id = module.key_vault.access_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = false
  target_port_ingress = 50051
  envs = merge(local.otel_envs, {
    "JWT_EXPIRY": "24h"
    "REFRESH_TOKEN_EXPIRY": "168h"
  })

  secrets = {
    "DSN": module.postgres.connection_string_secret_id
    "JWT_SECRET": module.key_vault.jwt_secret_id
    "RABBITMQ_CONNECTION_STRING": module.service-bus.connection_string_secret_id
  }
}

module "location_service" {
  source = "./modules/aca-service"
  name = "location-service"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id
  key_vault_access_identity_id = module.key_vault.access_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = false
  target_port_ingress = 50053
  envs = merge(local.otel_envs, {
    "REDIS_HOST": module.location-redis.hostname
    "REDIS_PORT": module.location-redis.port
    "REDIS_DB": 0
  })

  secrets = {
    "REDIS_PASSWORD": module.location-redis.kv_primary_access_key_id
    "MONGO_CONNECTION_STRING": module.documentdb.connection_string_secret_id
    "RABBITMQ_CONNECTION_STRING": module.service-bus.connection_string_secret_id
  }
}

module "trip_service" {
  source = "./modules/aca-service"
  name = "trip-service"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  key_vault_access_identity_id = module.key_vault.access_identity_id
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = false
  target_port_ingress = 50054
  envs = local.otel_envs

  secrets = {
    "DSN": module.postgres.connection_string_secret_id
    "HERE_ID": module.key_vault.here_id_secret_id
    "HERE_SECRET": module.key_vault.here_secret_secret_id
    "RABBITMQ_CONNECTION_STRING": module.service-bus.connection_string_secret_id
  }
}

module "user_service" {
  source = "./modules/aca-service"
  name = "user-service"

  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  key_vault_access_identity_id = module.key_vault.access_identity_id
  
  acr_login_server = local.acr_login_server
  acr_pull_identity_id = local.acr_pull_identity_id

  liveness_probe = {}
  readiness_probe = {}

  is_external_ingress = false
  target_port_ingress = 50055
  envs = local.otel_envs

  secrets = {
    "MONGO_CONNECTION_STRING": module.documentdb.connection_string_secret_id
  }
}