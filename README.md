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

TODO finish this