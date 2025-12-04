variable "resource_group_name" {
  description = "Name of the resource group to create resource in"
  type        = string
}

variable "resource_prefix" {
  description = "Prefix for the name for all resources"
  type        = string
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
}

variable "aca_subnet_id" {
  type = string
}

variable "key_vault_id" {
  description = "ID of the Key Vault to store connection string"
  type        = string
}