variable "resource_group_base_name" {
  description = "Base name of the resource group. Actual name will be {basename}-{env} if `environment` var is not set or null"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "environment" {
  description = "Environment identifier (dev/staging/prod/...)"
  type        = string
  nullable    = true
  default     = null
}
