# UIT-Go Microservices Backend

A non-modern, maybe-scalable microservices backend built with Go. Should support IaC, CI/CD and *technically* cloud-native.

## Repo structure

Currently do NOT support per-service deployment.

Each service will have a folder inside `service` folder:

- authentication-service
- location-service
- logger-service
- trip-service
- api-gateway: what it sounds like
- common: common shared files between services

And also:

- project: this shit should be gone in around 1 day.
- proto

## Instructions for local development

To start developing (or running locally):

### Dependencies

- Podman (recommended) or Docker installed with Compose.
- That's it. Plug and play!

### Run

1. Copy `.env.example` file and name it `.env` in the same root folder. Tweak it if needed.

2. Run the Compose app with

```bash
docker compose up
```

3. If you make any changes and want to see the result, (optionally do a `docker compose down`) then `docker compose up` again.

4. To "clean" the environment, run 
```bash
docker compose down -v
```
5. To clean everything, including built images, run: 
```bash
docker compose down -v --rmi all
```

## Notes

All volumes used are named volumes, not mounted volumes. This is to avoid permission errors when running on rootless or SELinux environment, etc. and for easier clean-up.

Development environment are managed through a `docker-compose.yml` file, and may drift from the production/staging environment deployed with `opentofu` and CD scripts. This is a trade-off for lightweight local environment with pure containers and Compose, and to utilize managed services (Postgres, Redis,...) provided by the cloud provider (in this case Azure).

## Footnote

Entire deploy flow is made without or with minimal LLM footprint. *All mistakes are my own.*

!TODO finish this