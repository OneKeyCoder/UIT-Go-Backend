resource "azurerm_container_app_environment_storage" "configs" {
  name                         = "configs-share"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.configs_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadOnly"
}

resource "azurerm_container_app" "alloy" {
  name                         = "alloy"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 4317
    transport        = "http2"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    container {
      name   = "alloy"
      image  = "grafana/alloy:latest"
      cpu    = "0.5"
      memory = "1Gi"
      
      command = ["run", "/etc/alloy/config.alloy", "--server.http.listen-addr=0.0.0.0:12345"]

      volume_mounts {
        name = "config-volume"
        path = "/etc/alloy/config.alloy"
        sub_path = ""
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
  }
}

# resource "azurerm_container_app_environment_storage" "jaeger_data" {
#   name                         = "jaeger-data"
#   container_app_environment_id = module.aca-infra.env-id
#   account_name                 = module.files-mount.storage_account_name
#   share_name                   = module.files-mount.jaeger_share_name
#   access_key                   = module.files-mount.storage_account_key
#   access_mode                  = "ReadWrite"
# }

# resource "azurerm_container_app" "jaeger" {
#   name                         = "jaeger"
#   container_app_environment_id = module.aca-infra.env-id
#   resource_group_name          = local.rg_name
#   revision_mode                = "Single"

#   ingress {
#     external_enabled = false
#     target_port      = 16686
#     transport        = "auto"
#     traffic_weight {
#       percentage      = 100
#       latest_revision = true
#     }
#   }

#   template {
#     container {
#       name   = "jaeger"
#       image  = "jaegertracing/jaeger:2.2.0"
#       cpu    = "0.5"
#       memory = "1Gi"
      
#       args = ["--config", "/etc/jaeger/config-badger.yml"]

#       volume_mounts {
#         name = "config-volume"
#         path = "/etc/jaeger"
#       }
#       volume_mounts {
#         name = "data-volume"
#         path = "/badger"
#       }
#     }
    
#     volume {
#       name         = "config-volume"
#       storage_name = azurerm_container_app_environment_storage.configs.name
#       storage_type = "AzureFile"
#     }
#     volume {
#       name         = "data-volume"
#       storage_name = azurerm_container_app_environment_storage.jaeger_data.name
#       storage_type = "AzureFile"
#     }
#   }
# }

# resource "azurerm_container_app_environment_storage" "prometheus_data" {
#   name                         = "prometheus-data"
#   container_app_environment_id = module.aca-infra.env-id
#   account_name                 = module.files-mount.storage_account_name
#   share_name                   = module.files-mount.prometheus_share_name
#   access_key                   = module.files-mount.storage_account_key
#   access_mode                  = "ReadWrite"
# }

# resource "azurerm_container_app" "prometheus" {
#   name                         = "prometheus"
#   container_app_environment_id = module.aca-infra.env-id
#   resource_group_name          = local.rg_name
#   revision_mode                = "Single"

#   ingress {
#     external_enabled = false
#     target_port      = 9090
#     transport        = "auto"
#     traffic_weight {
#       percentage      = 100
#       latest_revision = true
#     }
#   }

#   template {
#     container {
#       name   = "prometheus"
#       image  = "prom/prometheus:v3.2.0"
#       cpu    = "0.5"
#       memory = "1Gi"
      
#       args = [
#         "--config.file=/etc/prometheus/prometheus.yml",
#         "--storage.tsdb.path=/prometheus",
#         "--storage.tsdb.retention.time=15d",
#         "--web.enable-lifecycle",
#         "--web.enable-remote-write-receiver"
#       ]

#       volume_mounts {
#         name = "config-volume"
#         path = "/etc/prometheus"
#       }
#       volume_mounts {
#         name = "data-volume"
#         path = "/prometheus"
#       }
#     }
    
#     volume {
#       name         = "config-volume"
#       storage_name = azurerm_container_app_environment_storage.configs.name
#       storage_type = "AzureFile"
#     }
#     volume {
#       name         = "data-volume"
#       storage_name = azurerm_container_app_environment_storage.prometheus_data.name
#       storage_type = "AzureFile"
#     }
#   }
# }

# resource "azurerm_container_app_environment_storage" "loki_data" {
#   name                         = "loki-data"
#   container_app_environment_id = module.aca-infra.env-id
#   account_name                 = module.files-mount.storage_account_name
#   share_name                   = module.files-mount.loki_share_name
#   access_key                   = module.files-mount.storage_account_key
#   access_mode                  = "ReadWrite"
# }

# resource "azurerm_container_app" "loki" {
#   name                         = "loki"
#   container_app_environment_id = module.aca-infra.env-id
#   resource_group_name          = local.rg_name
#   revision_mode                = "Single"

#   ingress {
#     external_enabled = false
#     target_port      = 3100
#     transport        = "auto"
#     traffic_weight {
#       percentage      = 100
#       latest_revision = true
#     }
#   }

#   template {
#     container {
#       name   = "loki"
#       image  = "grafana/loki:3.2.1"
#       cpu    = "0.5"
#       memory = "1Gi"
      
#       args = ["-config.file=/etc/loki/loki-config.yml"]

#       volume_mounts {
#         name = "config-volume"
#         path = "/etc/loki"
#       }
#       volume_mounts {
#         name = "data-volume"
#         path = "/loki"
#       }
#     }
    
#     volume {
#       name         = "config-volume"
#       storage_name = azurerm_container_app_environment_storage.configs.name
#       storage_type = "AzureFile"
#     }
#     volume {
#       name         = "data-volume"
#       storage_name = azurerm_container_app_environment_storage.loki_data.name
#       storage_type = "AzureFile"
#     }
#   }
# }

# resource "azurerm_container_app_environment_storage" "grafana_data" {
#   name                         = "grafana-data"
#   container_app_environment_id = module.aca-infra.env-id
#   account_name                 = module.files-mount.storage_account_name
#   share_name                   = module.files-mount.grafana_share_name
#   access_key                   = module.files-mount.storage_account_key
#   access_mode                  = "ReadWrite"
# }

# resource "azurerm_container_app" "grafana" {
#   name                         = "grafana"
#   container_app_environment_id = module.aca-infra.env-id
#   resource_group_name          = local.rg_name
#   revision_mode                = "Single"

#   ingress {
#     external_enabled = true
#     target_port      = 3000
#     transport        = "auto"
#     traffic_weight {
#       percentage      = 100
#       latest_revision = true
#     }
#   }

#   template {
#     container {
#       name   = "grafana"
#       image  = "grafana/grafana:12.3.0"
#       cpu    = "0.5"
#       memory = "1Gi"
      
#       env {
#         name  = "GF_SECURITY_ADMIN_USER"
#         value = "admin"
#       }
#       env {
#         name  = "GF_SECURITY_ADMIN_PASSWORD"
#         value = "admin"
#       }
#       env {
#         name  = "GF_USERS_ALLOW_SIGN_UP"
#         value = "false"
#       }
#       env {
#         name  = "GF_FEATURE_TOGGLES_ENABLE"
#         value = "traceToLogs,correlations"
#       }
#       env {
#         name  = "GF_PATHS_PROVISIONING"
#         value = "/etc/grafana/provisioning"
#       }
#       env {
#         name  = "GF_AUTH_ANONYMOUS_ENABLED"
#         value = "true"
#       }
#       env {
#         name  = "GF_AUTH_ANONYMOUS_ORG_ROLE"
#         value = "Viewer"
#       }

#       volume_mounts {
#         name = "data-volume"
#         path = "/var/lib/grafana"
#       }
#     }
    
#     volume {
#       name         = "data-volume"
#       storage_name = azurerm_container_app_environment_storage.grafana_data.name
#       storage_type = "AzureFile"
#     }
#   }
# }
