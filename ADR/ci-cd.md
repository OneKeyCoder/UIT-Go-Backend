# Automation - CI/CD

**GitHub Actions** is used as the CI/CD platform, simply because CI/CD itself is just a task runner that run a shell script when a webhook is triggered, usually through Git repo push actions to a branch. Since the project is already hosted on GitHub, GitHub Actions is easy to add. It allows custom runner too, so no dependent costs, and can migrate easily anytime since it's just CI/CD.

Each services (a folder in the `services` folder) that need to be deployed will have its own pipeline defined. If any files in and only in that folder changes, it will trigger a build run, which `docker build` that image, then pushes to ACR, then trigger updates in ACA to create a new revision with `azure-cli`.

A revision is like an app update, config update or similar that doesn't change underlying infrastructure. Usually it's used to update our apps, like in this case, or even do A/B testing or multiple API versions, etc.

Rolling updates will be automatically done on revision change by ACA (it's just K8s) so no service disruption.
