# UIT-Go Microservices Backend

A non-modern, maybe-scalable microservices backend built with Go. Should support IaC, CI/CD and *technically* cloud-native.

## TODO

- [ ] Remove the `replace` directive from go.mod files and switch to go workspaces for local dev.

## Repo structure

Currently do NOT support per-service deployment.

Each service will have a folder inside `service` folder:

- authentication-service
- location-service
- logger-service
- trip-service
- api-gateway: what it sounds like

And also:

- common: common shared files between services
- proto

Apart from services, the repo also hosts infrastructure code at `infrastructure` folder. This should be separated into its own repo, but for easier access it's also here.

The `utils` folder hosts misc files for development or deployment, like init scripts for databases.

## Instructions for local development

To start developing (or running locally):

### Dependencies

- Podman (recommended) or Docker installed with Compose.
- That's it. Plug and play!

### Run

Use either `docker` or `podman` command here based on which you have installed and/or want to use.s

1. Copy `.env.example` file and name it `.env` in the same root folder. Tweak it if needed.

2. Run the Compose app with

```bash
docker compose build
docker compose up
```

3. If you make any changes and want to see the result, `docker compose down` then run the app again.

4. To "clean" the environment, run 
```bash
docker compose down -v
```
5. To clean everything, including built images, run: 
```bash
docker compose down -v --rmi all
```

Or, use Golang toolchain in the individual service folders if you want to run bare-metal for faster development.

## Notes

All volumes used are named volumes, not mounted volumes. This is to avoid permission errors when running on rootless or SELinux environment, etc. and for easier clean-up.

Development environment are managed through a `docker-compose.yml` file, and may drift from the production/staging environment deployed with `opentofu` and CD scripts. This is a trade-off for lightweight local environment with pure containers and Compose, and to utilize managed services (Postgres, Redis,...) provided by the cloud provider (in this case Azure).

Read more:
- https://www.redhat.com/en/blog/debug-rootless-podman-mounted-volumes
- https://stackoverflow.com/questions/79173758/podman-volume-mount-permissions-issues

## Footnote

Entire deploy flow is made without or with minimal LLM footprint. *All mistakes are my own.*

!TODO finish this