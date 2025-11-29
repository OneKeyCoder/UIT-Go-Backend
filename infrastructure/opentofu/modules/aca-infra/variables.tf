variable "resource_group_name" {
  description = "Name of the resource group to create vnet in"
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

variable "vnet_id" {
  type = string
}

variable "subnet_id" {
  type = string
}