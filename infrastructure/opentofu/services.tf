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
        for_each = {
          OTEL_EXPORTER: "otlp"
          OTEL_COLLECTOR_ENDPOINT: "alloy:4317"
          OTEL_INSECURE: "true"
        }
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
  name = "authentication-service"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name = local.rg_name
  revision_mode = "Single"

  template {
    container {
      name = "authentication-service"
      image = "${local.acr_login_server}/authentication-service:latest"
      cpu = "0.5"
      memory = "1Gi"
      dynamic "env" {
        for_each = {
          OTEL_EXPORTER: "otlp"
          OTEL_COLLECTOR_ENDPOINT: "alloy:4317"
          OTEL_INSECURE: "true"
        }
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

