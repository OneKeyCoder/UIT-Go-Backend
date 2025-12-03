variable "resource_group_name" {
  description = "Name of the resource group to create resource in"
  type        = string
}

variable "resource_prefix" {
  description = "Prefix of the name for the resources"
  type        = string
}

variable "location" {
  description = "Azure region of resource group"
  type        = string
}

variable "sku" {
  description = "Service tier to use"
  type = string
  validation {
    condition = contains(["Basic", "Premium", "Standard"], var.sku)
    error_message = "Invalid SKU, valid values is: Basic, Premium, Standard"
  }
}