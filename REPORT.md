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

## Core ADRs

### Security - Full managed vs Virtual Network

We can either:
- make a fully VNet-less infrastructure, where all components has a FQDN attached to it, accessible from either public Internet, private Azure, or both. Or,
- make a custom network, with subnets and proper ACLs to restrict access to resources.

We decided to go with proper custom virtual network, since exposing your databases and services directly to the Internet (even with firewalls and service endpoints) is still a pretty bad idea security-wise.

We deployed a DMZ tiered network scheme, where only a small subnet (DMZ) is actually exposed to the Internet through a Gateway with restricted ACLs, and all other services either live in a private subnet, or is managed by Azure elsewhere in the datacenter region, but connected to the virtual network through a service endpoint or a private endpoint through private link.

We use a basic 10.0.0.0/16 address space, and currently have these subnets:

- aca-subnet (10.0.0.0/22): This gives 10 bit of host address, 1019 hosts precisely, for our pods. This subnet is delegated to our computing solution (Azure Container Apps, see later).
- DMZ (10.0.4.0/24, **public**): The public subnet for services that is public.
- postgres-subnet (10.0.5.0/24): Subnet delegated to our PostgreSQL solution.
- endpoints-subnet (10.0.6.0/27): Subnet for private endpoints from Azure managed services, so that we have a private connection directly to the resource through an IP address. May or may not be used since private endpoints does have extra costs, will be touched on later.

### Compute - Azure Container Apps

App services and all auxiliary containers will be deployed as an **Azure Container Apps** (ACA).

This is a managed K8s solution, comparable to AWS Fargate, to automate and abstract away the control plane and config work. It also supports scale-to-zero and other features that K8s has, since it's K8s under the hood.

Each services deployed inside ACA environment (the entire cluster) have auto-scaling, auto-healing and load-balancing built in.

A service container is deployed as an ACA, in an ACA environment, running one or many replicas of that container, and have a load balancer to distribute load to these replicas.

### Persistent storage - Azure Files

### Data plane - Postgres

### Data plane - Redis Cache

### Data plane - MongoDB

### Message queue - Azure Service Bus

### Secret management - Azure Key Vault

This one is a simpler choice. With Azure Container Apps, you get an environment setup out of the box, like `docker-compose` env block, but the config is finicky, and spread out across the apps so it's harder to manage. We need a centralized keystore to quickly adjust, revoke and/or change the secrets in case of credentials leaks and similar.

Most keystore's feature sets are pretty close to each other, so picking any is fine functionality-wise. The only decision-makers lies in pricing and ease-of-use.

Azure Key Vault integrates directly into ACA so setup is very minimal, with no code changes to the underlying apps, so we get a very cloud-agnostic, no-lock-in solution. As such, there's not really a need for alternatives here.

### Azure Application Gateway

A gateway to public Internet. Provides load-balancing (with auto-scaling built in), traffic routing, and a WAF to filter out bad traffic.

Since we already have a dedicated API Gateway service, we only use traffic routing to split our domain into `api.*` for the API, and `monitor.*` for Grafana. 

However, WAF is an optional feature, as it provides higher security in exchange for ~50% of request performance, so we either need to double our price by scaling out, or accept the performance lost. As such, we don't use it.

### No NAT Gateway

Usually, a VNet would have a NAT Gateway to handle egress traffic (connections that started from within our infra) too, but since we don't need the static IP (HERE SDK does not require IP whitelisting) and the floor monthly cost is significantly higher than default ACA egress (around 2000$ for a study case we found), we decided NOT to include it.

## Modules

The project will primarily do Module D and E:
- Module D: Observability
- Module E: FinOps and Automation

## Module D: Observability

## Module E: FinOps and Automation

### Automation - CI/CD

**GitHub Actions** is used as the CI/CD platform, simply because CI/CD itself is just a task runner that run a shell script when a webhook is triggered, usually through Git repo push actions to a branch. Since the project is already hosted on GitHub, GitHub Actions is easy to add. It allows custom runner too, so no dependent costs, and can migrate easily anytime since it's just CI/CD.

