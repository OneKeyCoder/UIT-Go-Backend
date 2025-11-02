terraform {
  required_version = ">= 1.6.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.51.0"
    }
  }

  #! INFO: Configure remote state prior to first apply (Azure Storage backend recommended).
  # https://learn.microsoft.com/en-us/azure/developer/terraform/store-state-in-azure-storage?tabs=azure-cli#3-configure-terraform-backend-state
  backend "azurerm" {
    resource_group_name  = "tfstate-test"
    storage_account_name = "tfstatetest1254"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

locals {
  project_name = "uit-go"

  common_tags = merge({
    Project     = local.project_name
    Environment = var.environment
    ManagedBy   = "opentofu"
  }, var.additional_tags)

  app_name_prefix = substr(replace("${local.project_name}-${var.environment}", "_", "-"), 0, 32)
}

############################
# Core infrastructure      #
############################

data "azurerm_resource_group" "this" {
  name = var.resource_group_name
}

resource "azurerm_log_analytics_workspace" "this" {
  name                = substr("${local.project_name}-${var.environment}-law", 0, 63)
  location            = data.azurerm_resource_group.this.location
  resource_group_name = data.azurerm_resource_group.this.name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_analytics_retention_days
  tags                = local.common_tags
}

resource "azurerm_container_registry" "this" {
  name                = var.acr_name
  resource_group_name = data.azurerm_resource_group.this.name
  location            = data.azurerm_resource_group.this.location
  sku                 = var.acr_sku
  admin_enabled       = false
  tags                = local.common_tags
}

resource "azurerm_user_assigned_identity" "workload" {
  name                = substr("${local.project_name}-${var.environment}-uami", 0, 128)
  resource_group_name = data.azurerm_resource_group.this.name
  location            = data.azurerm_resource_group.this.location
  tags                = local.common_tags
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.this.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.workload.principal_id
}

resource "azurerm_container_app_environment" "this" {
  name                       = substr("${local.project_name}-${var.environment}-cae", 0, 63)
  location                   = data.azurerm_resource_group.this.location
  resource_group_name        = data.azurerm_resource_group.this.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
  tags                       = local.common_tags

  workload_profile {
    name                  = var.workload_profile_name
    workload_profile_type = var.workload_profile_type
  }
}

resource "azurerm_key_vault" "this" {
  name                        = var.key_vault_name
  location                    = data.azurerm_resource_group.this.location
  resource_group_name         = data.azurerm_resource_group.this.name
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  sku_name                    = "standard"
  purge_protection_enabled    = true
  soft_delete_retention_days  = 7
  enabled_for_deployment      = false
  enabled_for_disk_encryption = false
  rbac_authorization_enabled  = true
  tags                        = local.common_tags

  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
  }
}

resource "azurerm_role_assignment" "kv_workload" {
  scope                = azurerm_key_vault.this.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.workload.principal_id
}

resource "azurerm_role_assignment" "kv_admin" {
  scope                = azurerm_key_vault.this.id
  role_definition_name = "Key Vault Administrator"
  principal_id         = data.azurerm_client_config.current.object_id
}

############################
# Container apps           #
############################

resource "azurerm_container_app" "services" {
  for_each = var.container_apps

  name                         = substr(replace("${local.app_name_prefix}-${each.key}", "_", "-"), 0, 32)
  resource_group_name          = data.azurerm_resource_group.this.name
  container_app_environment_id = azurerm_container_app_environment.this.id
  revision_mode                = coalesce(each.value.revision_mode, "Single")
  tags                         = local.common_tags

  template {
    min_replicas = coalesce(each.value.min_replicas, 0)
    max_replicas = coalesce(each.value.max_replicas, 1)

    container {
      name   = each.key
      image  = "${azurerm_container_registry.this.login_server}/${each.value.image_repository}:${each.value.image_tag}"
      cpu    = each.value.cpu
      memory = each.value.memory

      dynamic "env" {
        for_each = coalesce(each.value.environment_variables, {})
        content {
          name  = env.key
          value = env.value
        }
      }

      dynamic "env" {
        for_each = coalesce(each.value.secret_environment_variables, [])
        content {
          name        = env.value.name
          secret_name = env.value.secret_name
        }
      }
    }
  }

  dynamic "ingress" {
    for_each = each.value.ingress == null ? [] : [each.value.ingress]
    content {
      external_enabled = ingress.value.external
      target_port      = ingress.value.target_port
      transport        = coalesce(ingress.value.transport, "auto")

      traffic_weight {
        latest_revision = true
        percentage      = 100
      }
    }
  }

  dynamic "secret" {
    for_each = coalesce(each.value.secrets, [])
    content {
      name                = secret.value.name
      key_vault_secret_id = secret.value.key_vault_secret_id
      identity            = coalesce(secret.value.identity_id, azurerm_user_assigned_identity.workload.id)
    }
  }

  registry {
    server   = azurerm_container_registry.this.login_server
    identity = azurerm_user_assigned_identity.workload.id
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.workload.id]
  }
}

############################
# Variables                #
############################

variable "resource_group_name" {
  description = "Name of the resource group to create or reuse."
  type        = string
}

variable "environment" {
  description = "Deployment environment identifier (e.g. dev, staging, prod)."
  type        = string
}

variable "acr_name" {
  description = "Globally unique Azure Container Registry name."
  type        = string
}

variable "acr_sku" {
  description = "SKU tier for the Azure Container Registry."
  type        = string
  default     = "Standard"
}

variable "workload_profile_name" {
  description = "Container Apps workload profile name."
  type        = string
  default     = "Consumption"
}

variable "workload_profile_type" {
  description = "Container Apps workload profile type (Consumption, Dedicated, or Premium)."
  type        = string
  default     = "Consumption"
}

variable "key_vault_name" {
  description = "Globally unique Key Vault name for shared secrets."
  type        = string
}

variable "log_analytics_retention_days" {
  description = "Retention period for Log Analytics data."
  type        = number
  default     = 30
}

variable "additional_tags" {
  description = "Optional map of additional tags to apply to all resources."
  type        = map(string)
  default     = {}
}

variable "container_apps" {
  description = "Microservice definitions to deploy as Azure Container Apps."
  type = map(object({
    image_repository      = string
    image_tag             = string
    cpu                   = number
    memory                = string
    revision_mode         = optional(string)
    min_replicas          = optional(number)
    max_replicas          = optional(number)
    environment_variables = optional(map(string))
    secret_environment_variables = optional(list(object({
      name        = string
      secret_name = string
    })))
    ingress = optional(object({
      external    = bool
      target_port = number
      transport   = optional(string)
    }))
    secrets = optional(list(object({
      name                = string
      key_vault_secret_id = string
      identity_id         = optional(string)
    })))
  }))
  default = {}
}

############################
# Outputs                 #
############################

output "resource_group_name" {
  value       = data.azurerm_resource_group.this.name
  description = "Deployed resource group name."
}

output "container_apps_default_hostname" {
  value       = { for k, app in azurerm_container_app.services : k => app.latest_revision_fqdn }
  description = "Public hostnames for the deployed container apps."
}
