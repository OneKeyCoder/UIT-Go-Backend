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

variable "sku_name" {
  description = "SKU name to use for underlying nodes."
  type = string
  default = "ComputeOptimized_X3"
}

variable "high_availability_enabled" {
  type = bool
  default = true
}

variable "endpoint_subnet_id" {
  description = "Subnet id to install private endpoint into"
  type = string
}

variable "endpoint_dns_zone_id" {
  type = string
}

variable "eviction_policy" {
  description = "Key eviction policy. https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/managed_redis#eviction_policy-1"
  type = string
  default = "AllKeysLRU"
}

variable "key_vault_id" {
  type = string
}

variable "tags" {
  type = map(string)
  default = {}
}