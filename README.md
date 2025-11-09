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

## Instructions

To start developing (or running locally):

1. Copy `.env.example` file and name it `.env` in the same root folder. Tweak it if needed.

2. Run the Compose app with

```bash
docker compose up
```

3. If you make any changes and want to see the result, do a `docker compose down` then `docker compose up` again.

TODO finish this