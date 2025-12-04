variable "resource_group_name" {
  description = "Name of the resource group to create resource in"
  type        = string
}

variable "name" {
  description = "Unique, alphanumeric-only name for the resource"
  type        = string
}

variable "location" {
  description = "Azure region of resource group"
  type        = string
}

variable "sku" {
  description = "Service tier to use"
  type = string
  default = "Standard"
  validation {
    condition = contains(["Basic", "Premium", "Standard"], var.sku)
    error_message = "Invalid SKU, valid values is: Basic, Premium, Standard"
  }
}