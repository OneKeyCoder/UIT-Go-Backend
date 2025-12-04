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

# I have decided against using modules, because the apis are stupid. Terraform syntax is horrible.
# Just declare the secrets manually I guess.

resource "azurerm_container_app" "api-gateway" {
  name = "api-gateway"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  revision_mode = "Single"

  template {
    container {
      name = "api-gateway"
      cpu = "0.5"
      memory = "1Gi"
      image = "${local.acr_login_server}/api-gateway:latest"
      dynamic "env" {
        for_each = local.otel_envs
        content {
          name = env.key
          value = env.value
        }
      }
    }
  }

  registry {
    server = local.acr_login_server
    identity = local.acr_pull_identity_id
  }
}

resource "azurerm_container_app" "logger-service" {
  name = "logger-service"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  revision_mode = "Single"

  secret {
    name = "mongo-connection-string"
    key_vault_secret_id = module.documentdb.connection_string_secret_id
    identity = local.acr_pull_identity_id
  }

  template {
    container {
      name = "logger-service"
      image = "${local.acr_login_server}/logger-service:latest"
      cpu = "0.5"
      memory = "1Gi"
      
      env {
        name = "MONGO_CONNECTION_STRING"
        secret_name = "mongo-connection-string"
      }
      
      dynamic "env" {
        for_each = local.otel_envs
        content {
          name = env.key
          value = env.value
        }
      }
    }
  }

  registry {
    server = local.acr_login_server
    identity = local.acr_pull_identity_id
  }
}

resource "azurerm_container_app" "authentication-service" {
  name                         = "authentication-service"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  secret {
    name                = "postgres-dsn"
    key_vault_secret_id = module.postgres.connection_string_secret_id
    identity            = local.acr_pull_identity_id
  }
  secret {
    name                = "jwt-secret"
    key_vault_secret_id = module.key_vault.jwt_secret_id
    identity            = local.acr_pull_identity_id
  }
  secret {
    name                = "rabbitmq-connection"
    key_vault_secret_id = module.service-bus.connection_string_secret_id
    identity            = local.acr_pull_identity_id
  }

  template {
    container {
      name   = "authentication-service"
      image  = "${local.acr_login_server}/authentication-service:latest"
      cpu    = "0.5"
      memory = "1Gi"

      env {
        name        = "DSN"
        secret_name = "postgres-dsn"
      }

      env {
        name        = "JWT_SECRET"
        secret_name = "jwt-secret"
      }

      env {
        name  = "JWT_EXPIRY"
        value = "24h"
      }

      env {
        name  = "REFRESH_TOKEN_EXPIRY"
        value = "168h"
      }

      env {
        name        = "RABBITMQ_CONNECTION_STRING"
        secret_name = "rabbitmq-connection"
      }

      dynamic "env" {
        for_each = local.otel_envs
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  registry {
    server   = local.acr_login_server
    identity = local.acr_pull_identity_id
  }
}

resource "azurerm_container_app" "location-service" {
  name                         = "location-service"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  secret {
    name                = "redis-password"
    key_vault_secret_id = module.location-redis.kv_primary_access_key_id
    identity            = local.acr_pull_identity_id
  }

  secret {
    name                = "redis-port"
    key_vault_secret_id = module.location-redis.kv_port_id
    identity            = local.acr_pull_identity_id
  }

  template {
    container {
      name   = "location-service"
      image  = "${local.acr_login_server}/location-service:latest"
      cpu    = "0.5"
      memory = "1Gi"

      env {
        name  = "REDIS_HOST"
        value = module.location-redis.hostname
      }

      env {
        name        = "REDIS_PORT"
        secret_name = "redis-port"
      }

      env {
        name        = "REDIS_PASSWORD"
        secret_name = "redis-password"
      }

      env {
        name  = "REDIS_DB"
        value = "0"
      }

      env {
        name  = "REDIS_TIME_TO_LIVE"
        value = "3600"
      }

      dynamic "env" {
        for_each = local.otel_envs
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  registry {
    server   = local.acr_login_server
    identity = local.acr_pull_identity_id
  }
}

resource "azurerm_container_app" "trip-service" {
  name                         = "trip-service"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  secret {
    name                = "postgres-dsn"
    key_vault_secret_id = module.postgres.connection_string_secret_id
    identity            = local.acr_pull_identity_id
  }

  secret {
    name                = "here-id"
    key_vault_secret_id = module.key_vault.here_id_secret_id
    identity            = local.acr_pull_identity_id
  }

  secret {
    name                = "here-secret"
    key_vault_secret_id = module.key_vault.here_secret_secret_id
    identity            = local.acr_pull_identity_id
  }

  secret {
    name                = "rabbitmq-connection"
    key_vault_secret_id = module.service-bus.connection_string_secret_id
    identity            = local.acr_pull_identity_id
  }

  template {
    container {
      name   = "trip-service"
      image  = "${local.acr_login_server}/trip-service:latest"
      cpu    = "0.5"
      memory = "1Gi"

      env {
        name        = "DSN"
        secret_name = "postgres-dsn"
      }

      env {
        name        = "HERE_ID"
        secret_name = "here-id"
      }

      env {
        name        = "HERE_SECRET"
        secret_name = "here-secret"
      }

      env {
        name        = "RABBITMQ_CONNECTION_STRING"
        secret_name = "rabbitmq-connection"
      }

      dynamic "env" {
        for_each = local.otel_envs
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  registry {
    server   = local.acr_login_server
    identity = local.acr_pull_identity_id
  }
}
