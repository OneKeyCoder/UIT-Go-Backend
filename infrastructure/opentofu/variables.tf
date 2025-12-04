# variables.tf – containing the variable declarations used in the resource blocks.

variable "resource_prefix" {
  description = "Prefix for the name for all resources"
  type        = string
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "postgres_admin_username" {
  type = string
  sensitive = true
}

variable "postgres_admin_password" {
  type = string
  sensitive = true
}