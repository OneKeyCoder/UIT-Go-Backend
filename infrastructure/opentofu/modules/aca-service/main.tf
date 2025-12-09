locals {
  image_name = var.image_name != null && var.image_name != "" ? var.image_name : var.name
}

resource "azurerm_container_app" "go-services" {
  name = var.name
  container_app_environment_id = var.container_app_environment_id
  resource_group_name = var.resource_group_name
  revision_mode = "Single"

  ingress {
    external_enabled = var.is_external_ingress
    target_port = var.target_port_ingress
    traffic_weight {
      percentage = 100
      latest_revision = true
    }
  }

  dynamic "secret" {
    for_each = var.secrets
    content {
      name = secret.key
      key_vault_secret_id = secret.value
    }
  }

  template {
    container {
      name = var.name
      cpu = "0.5"
      memory = "1Gi"
      image = "${var.acr_login_server}/${locals.image_name}:${var.image_tag}"
      dynamic "liveness_probe" {
        for_each = var.liveness_probe != null ? [var.liveness_probe] : []
        iterator = "i"
        content {
          transport             = i.value.transport
          port                  = i.value.port
          path                  = i.value.path
          interval_seconds      = i.value.interval_seconds
          timeout               = i.value.timeout
          failure_count_threshold = i.value.failure_count_threshold
          initial_delay         = i.value.initial_delay
        }
      }
      dynamic "readiness_probe" {
        for_each = var.liveness_probe != null ? [var.liveness_probe] : []
        iterator = "i"
        content {
          transport             = i.value.transport
          port                  = i.value.port
          path                  = i.value.path
          interval_seconds      = i.value.interval_seconds
          timeout               = i.value.timeout
          failure_count_threshold = i.value.failure_count_threshold
          success_count_threshold = i.value.success_count_threshold
          initial_delay         = i.value.initial_delay
        }
      }
      dynamic "env" {
        for_each = var.envs
        content {
          name = env.key
          value = env.value
        }
      }
      dynamic "env" {
        for_each = var.secrets
        content {
          name = env.key
          secret_name = env.key
        }
      }
    }
  }

  registry {
    server = var.acr_login_server
    identity = var.acr_pull_identity_id
  }

  lifecycle {
    ignore_changes = [ secret ]
  }
}