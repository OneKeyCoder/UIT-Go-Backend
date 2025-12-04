resource "azurerm_container_app" "go-services" {
  name = var.name
  container_app_environment_id = var.container_app_environment_id
  resource_group_name = var.resource_group_name
  revision_mode = "Single"

  secret {
    name = ""
    key_vault_secret_id = ""
  }

  template {
    container {
      name = var.name
      cpu = "0.5"
      memory = "1Gi"
      image = "${var.acr_login_server}/${var.name}:${var.image_tag}"
      dynamic "env" {
        for_each = var.envs
        content {
          name = env.key
          value = env.value
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