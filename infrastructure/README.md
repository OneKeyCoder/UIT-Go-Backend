# Infrastructure

This should be a separate repo from the services code. But I digress.

## How-tos

I trust that you know how UNIX-like terminal works. Either terraform or opentofu should work, but this project assumes opentofu. Replace `tofu` in the commands with `terraform` if that's what you use.

### Prepare to deploy first time

We're gonna work inside `opentofu` folder, so make it your CWD by `cd`-ing into it. This will be our project folder.

Install terraform or opentofu, then azure-cli. Run `az login` and login to your Azure account.

Make sure you have a domain name, get a SSL certificate for the domain or subdomain you want to use. If provided cert is NOT in PFX format, like PEM, PKCS7,... convert to PFX with `openssl` command. Google is your bestfriend. **DO REMEMBER THE PASSWORD**, you're gonna need it later.

After you have the .pfx file, put it into the `certs` folder.

Now, create a `.tfvars` file inside the project folder.

Populate it with variables described in the `variables.tf` file in the project folder. You can skip ones with "default", everything else must be provided. Or, you can provide it on the CLI, again Google it if you want to do so.

Also, if you want a remote state file, open `provider.tf` file, uncomment `backend "azurerm"` block if commented out, and put your Azure Storage account credentials here, one that your logged in Azure account has access to. Otherwise, if you **do NOT need a remote state file**, just make sure to **comment out** that block.

I've decided against using a variable for this to checkout the remote state into git for better collaboration. But yeah, it should be a var. You can change these into vars if that's what you want.

### First-time deploy

Now, it's a bit convoluted, and requires multiple `tf deploy`s, but bare with me.

First, we need to deploy the resource group, and the ACR to keep the images. Run

```bash
tofu init
```