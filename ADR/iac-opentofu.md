# Automation - IaC OpenTofu modules

The IaC codebase is splitted into modules for easier re-use, and structured in industry best practice.

We have (at least) these modules:
- networking: vnets, acls, dns,...
- aca-infra: define ACR, ACA environments,...
- aca-service: create an ACA service that runs inside ACA environment created with `aca-infra`
- postgres
- redis
- resource-group
- azure-files
- service-bus
- key-vault
- documentdb
- app-gw

Then combined in the top-level module `main.tf`.

In case a new service is added, just add a new block of module `aca-service` into the top-level module with configs for it.

We do have the problem of chicken-and-egg when provisioning ACR and ACA, since ACR is newly-deployed, and have no images on it yet, but ACA needs a pullable image to provision, so this would fails.

https://www.mytechramblings.com/posts/how-to-push-a-container-image-into-acr-using-te

We picked the build-inline option, which involves using `null_resource` to trigger a shell command to use `az acr build` to build the images and pushes to ACR after creating the ACR but before creating ACA so that it has a valid image.

This is hacky, but the other option is to use a dummy image and update later with CI/CD, which can overwrite when we do `tf apply` so in my opinion even more hacky. This is the state of Terraform. This is not a good state. But what can we do?
