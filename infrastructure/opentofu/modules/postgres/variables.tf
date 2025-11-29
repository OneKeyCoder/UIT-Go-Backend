variable "resource_group_name" {
  description = "Name of the resource group to create db in"
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

variable "virtual_network_id" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "db_names" {
  description = "List of db names to create"
  type    = set(string)
  default = ["auth_db", "trip_db"]
}

variable "admin_username" {
  type = string
}

variable "admin_password" {
  type = string
}

variable "sku_name" {
  description = "Compute size SKU"
  type = string
  default = "GP_Standard_D4ads_v5"
}