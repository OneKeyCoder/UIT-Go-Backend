# outputs.tf – containing the output that needs to be generated on successful completion of “apply” operation.

output "public_ip_address" {
  value = module.app-gw.public_ip_address
}

output "acr_name" {
  value = module.acr.name
}