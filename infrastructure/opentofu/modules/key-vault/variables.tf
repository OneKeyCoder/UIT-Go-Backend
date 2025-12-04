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

variable "allowed_subnet_ids" {
  description = "List of Subnet IDs that are allowed to access the Key Vault public FDNS"
  type        = set(string)
}

variable "jwt_secret" {
  description = "JWT secret for authentication service"
  type        = string
  sensitive   = true
}

variable "here_id" {
  description = "HERE Maps API ID for trip service"
  type        = string
  sensitive   = true
}

variable "here_secret" {
  description = "HERE Maps API secret for trip service"
  type        = string
  sensitive   = true
}
