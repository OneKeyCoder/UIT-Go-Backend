# shared configs

resource "azurerm_container_app_environment_storage" "configs" {
  name                         = "configs-share"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.configs_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadOnly"
}

# loki

resource "azurerm_container_app_environment_storage" "loki-data" {
  name                         = "loki-data"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.loki_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_container_app" "loki" {
  name                         = "loki"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 3100
    transport        = "auto"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    container {
      name   = "loki"
      image  = "grafana/loki:3.2.1"
      cpu    = "0.5"
      memory = "1Gi"
      
      args = ["-config.file=/etc/loki/local-config.yml"]

      volume_mounts {
        name = "config-volume"
        path = "/etc/loki/local-config.yaml"
        sub_path = "loki-config.yml"
      }
      volume_mounts {
        name = "data-volume"
        path = "/loki"
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
    volume {
      name         = "data-volume"
      storage_name = azurerm_container_app_environment_storage.loki-data.name
      storage_type = "AzureFile"
      mount_options = "dir_mode=0777,file_mode=0777,mfsymlinks,cache=strict,nosharesock,nobrl"
    }
  }
}

# tempo

resource "azurerm_container_app_environment_storage" "tempo_data" {
  name                         = "tempo-data"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.tempo_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_container_app" "tempo" {
  name                         = "tempo"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 3200
    transport        = "auto"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    container {
      readiness_probe {
        port = 3200
        transport = "HTTP"
        path = "/ready"
        interval_seconds = 30
        timeout = 10
        failure_count_threshold = 3
        initial_delay = 30
      }
      name   = "tempo"
      image  = "grafana/tempo:latest"
      cpu    = "0.5"
      memory = "1Gi"
      
      args = ["-config.file=/etc/tempo/config.yml"]

      volume_mounts {
        name = "config-volume"
        path = "/etc/tempo/config.yaml"
        sub_path = "tempo-config.yml"
      }
      volume_mounts {
        name = "data-volume"
        path = "/var/tempo"
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
    volume {
      name         = "data-volume"
      storage_name = azurerm_container_app_environment_storage.tempo_data.name
      storage_type = "AzureFile"
      mount_options = "dir_mode=0777,file_mode=0777,mfsymlinks,cache=strict,nosharesock,nobrl"
    }
  }
}

# alloy

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
      
      args = ["run", "/etc/alloy/config.alloy", "--server.http.listen-addr=0.0.0.0:12345"]

      volume_mounts {
        name = "config-volume"
        path = "/etc/alloy/config.alloy"
        sub_path = "alloy-config.alloy"
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
  }

  depends_on = [ 
    azurerm_container_app.loki,
    azurerm_container_app.tempo
  ]
}

# alertmanager

resource "azurerm_container_app_environment_storage" "alertmanager_data" {
  name                         = "alertmanager-data"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.alertmanager_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_container_app" "alertmanager" {
  name                         = "alertmanager"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 9093
    transport        = "auto"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    container {
      name   = "alertmanager"
      image  = "prom/alertmanager:v0.28.1"
      cpu    = "0.5"
      memory = "1Gi"
      
      args = [
        "--config.file=/etc/alertmanager/alertmanager.yml",
        "--storage.path=/alertmanager",
        "--web.external-url=http://localhost:9093",
        "--cluster.advertise-address=0.0.0.0:9093",
      ]

      volume_mounts {
        name = "config-volume"
        path = "/etc/alertmanager/alertmanager.yml"
        sub_path = "alertmanager.yml"
      }
      volume_mounts {
        name = "data-volume"
        path = "/alertmanager"
      }
      liveness_probe {
        port = 9093
        transport = "HTTP"
        path = "/-/healthy"
        interval_seconds = 30
        timeout = 10
        failure_count_threshold = 3
        initial_delay = 10
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
    volume {
      name         = "data-volume"
      storage_name = azurerm_container_app_environment_storage.alertmanager_data.name
      storage_type = "AzureFile"
      mount_options = "dir_mode=0777,file_mode=0777,mfsymlinks,cache=strict,nosharesock,nobrl"
    }
  }
}

# prometheus

resource "azurerm_container_app_environment_storage" "prometheus_data" {
  name                         = "prometheus-data"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.prometheus_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_container_app" "prometheus" {
  name                         = "prometheus"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 9090
    transport        = "auto"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    max_replicas = 1
    container {
      name   = "prometheus"
      image  = "prom/prometheus:v3.2.0"
      cpu    = "0.5"
      memory = "1Gi"
      
      args = [
        "--config.file=/etc/prometheus/prometheus.yml",
        "--storage.tsdb.path=/prometheus",
        "--storage.tsdb.retention.time=15d",
        "--web.enable-lifecycle",
        "--web.enable-remote-write-receiver"
      ]

      volume_mounts {
        name = "config-volume"
        path = "/etc/prometheus/prometheus.yml"
        sub_path = "prometheus.yml"
      }
      volume_mounts {
        name = "config-volume"
        path = "/etc/prometheus/prometheus-alerts.yml"
        sub_path = "prometheus-alerts.yml"
      }
      volume_mounts {
        name = "data-volume"
        path = "/prometheus"
      }
    }
    
    volume {
      name         = "config-volume"
      storage_name = azurerm_container_app_environment_storage.configs.name
      storage_type = "AzureFile"
    }
    volume {
      name         = "data-volume"
      storage_name = azurerm_container_app_environment_storage.prometheus_data.name
      storage_type = "AzureFile"
      mount_options = "dir_mode=0777,file_mode=0777,mfsymlinks,cache=strict,nosharesock,nobrl"
    }
  }
}

# grafana - requires custom image

resource "azurerm_container_app_environment_storage" "grafana_data" {
  name                         = "grafana-data"
  container_app_environment_id = module.aca-infra.env-id
  account_name                 = module.files-mount.storage_account_name
  share_name                   = module.files-mount.grafana_share_name
  access_key                   = module.files-mount.storage_account_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_container_app" "grafana" {
  name                         = "grafana"
  container_app_environment_id = module.aca-infra.env-id
  resource_group_name          = local.rg_name
  revision_mode                = "Single"

  ingress {
    external_enabled = true
    target_port      = 3000
    transport        = "auto"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  identity {
    type = "SystemAssigned, UserAssigned"
    identity_ids = [module.acr.acr_pull_identity_id]
  }
  registry {
    server = module.acr.login_server_url
    identity = module.acr.acr_pull_identity_id
  }

  template {
    container {
      name   = "grafana"
      image  = "${module.acr.login_server_url}/grafana-custom:latest"
      cpu    = "1"
      memory = "2Gi"

      env {
        name  = "GF_USERS_ALLOW_SIGN_UP"
        value = "false"
      }
      env {
        name  = "GF_FEATURE_TOGGLES_ENABLE"
        value = "traceToLogs,correlations"
      }
      env {
        name  = "GF_PATHS_PROVISIONING"
        value = "/etc/grafana/provisioning"
      }
      env {
        name = "GF_SERVER_ROOT_URL"
        value = module.app-gw.monitor_public_hostname
      }

      volume_mounts {
        name = "data-volume"
        path = "/var/lib/grafana"
      }

      liveness_probe {
        port = 3000
        transport = "HTTP"
        path = "/api/health"
        interval_seconds = 30
        timeout = 10
        failure_count_threshold = 3
        initial_delay = 10
      }
    }

    volume {
      name         = "data-volume"
      storage_name = azurerm_container_app_environment_storage.grafana_data.name
      storage_type = "AzureFile"
      mount_options = "dir_mode=0777,file_mode=0777,mfsymlinks,cache=strict,nosharesock,nobrl"
    }
  }

  depends_on = [ 
    azurerm_container_app.prometheus,
    azurerm_container_app.loki,
    azurerm_container_app.tempo,
  ]
}
