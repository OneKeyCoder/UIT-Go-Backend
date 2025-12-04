# provider.tf – containing the terraform block, s3 backend definition, provider configurations, and aliases.

terraform {
  required_version = ">= 1.11.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.54.0"
    }
    random = {
      source = "hashicorp/random"
      version = "3.7.2"
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
  subscription_id = var.subscription_id
  features {
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
  }
}
