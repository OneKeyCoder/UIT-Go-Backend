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

variable "private_endpoint_subnet_id" {
  description = "Subnet to install private endpoint to the key vault in"
  type = string
}

variable "private_dns_zone_id" {
  description = "Private DNS zone"
}