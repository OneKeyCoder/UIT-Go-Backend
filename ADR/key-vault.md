# Secret management - Azure Key Vault

This one is a simpler choice. With Azure Container Apps, you get an environment setup out of the box, like `docker-compose` env block, but the config is finicky, and spread out across the apps so it's harder to manage. We need a centralized keystore to quickly adjust, revoke and/or change the secrets in case of credentials leaks and similar.

Most keystore's feature sets are pretty close to each other, so picking any is fine functionality-wise. The only decision-makers lies in pricing and ease-of-use.

Azure Key Vault integrates directly into ACA so setup is very minimal, with no code changes to the underlying apps, so we get a very cloud-agnostic, no-lock-in solution. As such, there's not really a need for alternatives here. But if you insist, Bitnami Sealed Secrets and other Key Vault (from HashiCorp for example) can be used. We also avoid egress by using Azure services.
