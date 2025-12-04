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

variable "subnet_id" {
  description = "Subnet ID to create resource in. Must not contain any other resource."
  type        = string
}
variable "base_hostname" {
  description = "Base host name to use for AppGW, where a subdomain will be prepended for each endpoint"
  type        = string
}

variable "api_aca_fqdn" {
  description = "FQDN of target API ACA (target app, not environment)"
  type        = string
}
variable "api_subdomain" {
  description = "Subdomain of the api endpoint."
  default     = "api"
  type        = string
}

variable "monitor_aca_fqdn" {
  description = "FQDN of target Grafana ACA (target app, not environment)"
  type        = string
}
variable "monitor_subdomain" {
  description = "Subdomain of the api endpoint."
  default     = "monitor"
  type        = string
}
variable "request_timeout" {
  type    = number
  default = 30
}

variable "pfx_ssl_filename" {
  description = "Name of the .pfx file containing SSL certificate for frontend in the `certs` folder. Should have all endpoint hostnames in CN."
  type        = string
}
variable "pfx_ssl_password" {
  description = "Password for the pfx SSL file"
  type        = string
  sensitive = true
}
