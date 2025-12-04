# variables.tf – containing the variable declarations used in the resource blocks.

variable "subscription_id" {
  type = string
}

variable "acr_name" {
  type = string
}

variable "resource_prefix" {
  description = "Prefix for the name for all resources"
  type        = string
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
}

variable "postgres_admin_username" {
  type = string
}
variable "postgres_admin_password" {
  type = string
  sensitive = true
}
variable "postgres_admin_password_version" {
  type = number
}

variable "base_hostname" {
  type = string
}

variable "pfx_ssl_filename" {
  type = string
}

variable "pfx_ssl_password" {
  type = string
  sensitive = true
}