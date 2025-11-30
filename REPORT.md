# Final ~~duet~~ report

Heads-up: We uses Azure instead of AWS to:
1. explore our options
2. Azure provides 100$ of free credit for students
3. better looking portal

Aside from these, most of the concepts should be similar or have a competitor alternative between 2 service, like ACA=Fargate, EC2=VM,...

Keep in mind, generally Azure is more invested in their own MS SQL Server and Document DB services, while AWS is more invested in MySQL, PostgreSQL so hyperscale service offerings between 2 companies in these areas might differ.

## App diagram - dev's story

We have the following services in our app skeleton:

- API Gateway: Act as HTTP-to-gRPC translator, and route requests to the correct service, and can have more functionality added. This is separate to avoid lock-in, since we can move to another service easily, and devs can test new functionality locally with `compose`.
- Logger service: All other services if need to save data or analytics will produce to the message queue, then `logger service` will consume from it, process if needed, and save to MongoDB store for later use. This is a PoC for a BigData pipeline.
- User service: Handle users
- Trip service: Handle trip progress, trip status, etc.
- Location service: Handle location-related worloads, like driver location, user location tracking, find nearby drivers, etc. As it needs high data throughput, frequent updates, low latency (near-real-time), and doesnt need to persist data, it uses in-memory database (Redis).
- Authentication service: handle user credentials and authentication operations. Does NOT do authorization, as we use stateless JWT.

Cross services communication will use gRPC, with ProtoBuf as the serialization method to reduce data serialization overhead but might sacrifice dev experience.

## System architecture diagram

![architecture diagram](images/architecture.png)

![architecture diagram](images/architecture-dark.png)

Short outline:

- Custom Virtual networks (not fully-managed networking), with private ACA dedicated subnet, private Postgres subnet, and a public DMZ subnet to handle Internet traffic.
- Services will be deployed in **Azure Container Apps**
- RabbitMQ in local dev will be replaced by **Azure Service Bus**
- Redis will use **Azure Managed Redis**
- Secrets will be stored in **Azure Key Vault** and connected and mounted into ACA
- MongoDB will use **DocumentDB with MongoDB**
- Persistent volumes in ACA will mount from **Azure Files**.
- Public traffic will go through **Azure Application Gateway**, put inside DMZ subnet.
- Private DNS zones will be used to resolve names of managed services inside VNet.
- Location service will call to **HERE Maps SDK**, an external Map API.

## Compute

App services and all auxiliary containers will be deployed as an **Azure Container Apps** (ACA). This is a managed K8s solution, comparable to AWS Fargate, to automate and abstract away the control plane and config work. It also supports scale-to-zero and other features that K8s has, since it's K8s under the hood.

Each services deployed inside ACA environment (the entire cluster) have auto-scaling, auto-healing and load-balancing built in.

A service container is deployed as an ACA, in an ACA environment, running one or many replicas of that container, and have a load balancer to distribute load to these replicas.

