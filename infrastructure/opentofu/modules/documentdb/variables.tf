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

variable "admin_username" {
  type = string
}

variable "admin_password" {
  type = string
  sensitive = true
}

variable "compute_tier" {
  type = string
  default = "M30"
}

variable "storage_size_in_gb" {
  type = number
  default = 128
}

variable "tags" {
  type = map(string)
  default = {}
}

variable "key_vault_id" {
  description = "ID of the Key Vault to store connection string"
  type        = string
}