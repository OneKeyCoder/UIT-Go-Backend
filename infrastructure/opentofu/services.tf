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
    name = "MONGO_CONNECTION_STRING"
    key_vault_secret_id = module.documentdb.connection_string
  }

  template {
    container {
      name = "logger-service"
      image = "${local.acr_login_server}/logger-service:latest"
      cpu = "0.5"
      memory = "1Gi"
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

