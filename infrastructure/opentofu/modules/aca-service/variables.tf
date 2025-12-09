variable "resource_group_name" {
  description = "Name of the resource group to create resource in"
  type        = string
}
variable "name" {
  description = "Name for the service, will be used as internal domain"
  type        = string
}

variable "container_app_environment_id" {
  type = string
}

variable "acr_login_server" {
  type = string
}

variable "acr_pull_identity_id" {
  type = string
}

variable "image_name" {
  description = "image name to pull, defaults to the container app name"
  type = string
  default = null
}
variable "image_tag" {
  type = string
  default = "latest"
}

variable "envs" {
  type = map(string)
  default = {}
}
variable "secrets" {
  type = map(string)
  default = {}
}

variable "is_external_ingress" {
  type = bool
  default = false
}
variable "target_port_ingress" {
  type = number
  default = 80
}