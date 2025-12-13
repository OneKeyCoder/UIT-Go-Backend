# Final ~~duet~~ report

This is the report for the course SE360.Q11's project.

## Members

- 23520923: Hồ Nguyên Minh
- 23520906: Hứa Văn Lý
- 23520950: Phan Đức Minh
- 23520949: Phan Đình Minh

## Project disclaimers

Heads-up: We uses Azure instead of AWS to:

1. explore our options
2. Azure provides 100$ of free credit for students
3. better looking portal

Aside from these, most of the concepts should be similar or have a competitor alternative between 2 service, like ACA=Fargate, EC2=VM,...

Keep in mind, generally Azure is more invested in their own MS SQL Server and Document DB services, while AWS is more invested in MySQL, PostgreSQL so hyperscale service offerings between 2 companies in these areas might differ.

And we also uses `opentofu` instead of `terraform` for IaC because licensing and Linux Foundation backing. However, there should be no difference running our tf project with both tools. Locally I symlinked `terraform` to `tofu` and everything works fine.

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

## Core system architecture ADRs

All ADRs not related to system (like app architecture) will be defined inside `ADR/` folder. This section only covers decisions for deployment and system related resources.

For example, we assume that "Postgres is needed for the app. How do we best deploy it" instead of questioning "Why Postgres". That part is handled by the dev team. This section is written under the viewpoint of the operation team, who's handed a list of services to deploy and requirements along with it.

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

This is a no-brainer. ACA provides only 3 volume mount options, 2 of which is ephemeral, so we have to use the third and only persistent option - Azure Files - to save data for services like Grafana, Prometheus,...

### Data plane - Postgres

Since data latency is a big requirement, we do NOT consider any options that has to have egress traffic (is outside of Azure network). This removes Neon Serverless Postgres, Render Postgres, and similar services along with AWS's offerings.

For Postgres, we have 3 main cloud-native options:

- Azure CosmosDB for PostgreSQL
- Azure Database for PostgreSQL - flexible servers
- Neon Managed Postgres in Azure

Azure CosmosDB is actually a rebrand of Managed Postgres with Citus plugin for horizontal scaling, and is deprecated, so we can't use it.

Neon Managed Postgres in Azure is dead after Jan 2026. So we also can't use it.

Our only option left is Azure Database for PostgreSQL - flexible servers. This has two options, single server option, and an `elastic cluster` option, which, like CosmosDB, is a cluster with Citus plugin.

For now, it will be deployed as single master server, with read-only replicas spread out in the same AZ, and has zone-redundant with a secondary fall-over server in another AZ. Or, if data is really important and have the budget to spare, in another region with geo-redundant. Geo-redundant only requires paying more for the egress cross-region transfer, server costs should be the same.

Also supports vertical scaling with near-zero downtime (~30 seconds to reboot, or a VM-swap zero-downtime if low traffic).

We will use the dedicated subnet hosting version to put it inside our VNet.

### Why no sharding? How about horizontal scaling?

The service itself supports scaling horizontally with sharding through Citus plugin, but I decided against using it because:

- Very expensive upfront cost (multiple nodes, multiple HA replicas)
- Use case does not warrant the cost (yet)
- Sharding requires data changes or more modifications to have proper sharding performance
- We do NOT have multi-tenant usecase, so sharding will be very hard to get right, or outright impossible
- We can always upgrade to it later when it's needed
- Do not have proper subnetting
- Cannot share DBs in the same server to preserve resources

Postgres cluster requires a sharding strategy, which is either schema-based - which we do not have a schema split to use and is not dynamically scalable - or row-based - which requires data and code changes, along with query considerations. Citus is designed more for multi-tenant applications and such, so we will (for now) use the single-node cluster or server mode, which still has vertical scaling both compute and storage, and horizontal scaling somewhat with read replicas, which technically is a way to scale horizontally, but only alleviates read workload and comes at a cost of slight delay from data replication, only for the Auth service as we do not have high writes there.

If in the future the app grows to require sharding, we can just enable Citus. Or enable read replicas and pgPool to have load-distributed Postgres on a single FQDN.

#### References

Neon deprecated:

- https://neon.com/docs/manage/azure  
- https://learn.microsoft.com/en-us/azure/partner-solutions/neon/overview

> The Neon Azure Native Integration is deprecated and reaches end of life on January 31, 2026. After this date, Azure-managed organizations will no longer be available. Migrate your projects to a Neon-managed organization to continue using Neon.

CosmosDB for PostgreSQL deprecated:

- https://www.reddit.com/r/AZURE/comments/1o7gm7f/comment/njwxp0g/?context=3&share_id=ZS63-PBjspVNUQU0Km6N2&utm_medium=ios_app&utm_name=ioscss&utm_source=share&utm_term=1
- https://www.reddit.com/r/AZURE/comments/1ol63qe/weve_already_migrated_from_postgres_single_server/
- https://learn.microsoft.com/en-us/azure/cosmos-db/postgresql/introduction

> Azure Cosmos DB for PostgreSQL is no longer supported for new projects. Don't use this service for new projects. Instead, use one of these two services:
> 
>  - Use Azure Cosmos DB for NoSQL for a distributed database solution designed for high-scale scenarios with a 99.999% availability service level agreement (SLA), instant autoscale, and automatic failover across multiple regions.
> 
> - Use the Elastic Clusters feature of Azure Database For PostgreSQL for sharded PostgreSQL using the open-source Citus extension.

### Data plane - Redis Cache

Azure Managed Redis was picked. We could, and wanted to, use Valkey instead, since it's completely open-source and not bound to an enterprise entity (backed by Linux Foundation), but since unlike AWS, Azure does not provide a Valkey Managed service, we use Redis instead.

Self hosting Redis instance inside ACA with Azure Disk persistent is also an option, but overall not worth the maintanance effort.

### Data plane - MongoDB

Uses Azure DocumentDB for MongoDB, which is CosmosDB under the hood, a NoSQL open-source DB made my Microsoft that's compatible with MongoDB. Supports auto-scaling and other usual features.

Alternative to this that we considered is MongoDB Atlas in Azure, but since it's closed source, and requires licensing, along with the fact that it's almost deprecated, makes it not a good option.

### Message queue - Azure Service Bus

Originally, the app is developed to use RabbitMQ 3, with AMQP 0-9-1, which is the older standard. We could run a cluster inside ACA, but since configuring such a cluster would require tremendous work both in intial config and in management (around 3 docs pages, each really really long) to ensure a healthy cluster, we decided against it.

So, we migrated the app to use AMQP 1.0, the newer protocol, which both RabbitMQ 4+, and more importantly Azure Service Bus supports, so that we can use RabbitMQ 4 in local dev `docker-compose`, and **Azure Service Bus** when deploying. This also mitigate vendor-lockin since when moving away from Azure we can just drop in an AMQP1.0-compatible service bus, or RabbitMQ 4 itself.

### Secret management - Azure Key Vault

This one is a simpler choice. With Azure Container Apps, you get an environment setup out of the box, like `docker-compose` env block, but the config is finicky, and spread out across the apps so it's harder to manage. We need a centralized keystore to quickly adjust, revoke and/or change the secrets in case of credentials leaks and similar.

Most keystore's feature sets are pretty close to each other, so picking any is fine functionality-wise. The only decision-makers lies in pricing and ease-of-use.

Azure Key Vault integrates directly into ACA so setup is very minimal, with no code changes to the underlying apps, so we get a very cloud-agnostic, no-lock-in solution. As such, there's not really a need for alternatives here. But if you insist, Bitnami Sealed Secrets and other Key Vault (from HashiCorp for example) can be used. We also avoid egress by using Azure services.

### Azure Application Gateway

A gateway to public Internet. Provides load-balancing (with auto-scaling built in), traffic routing, and a WAF to filter out bad traffic.

Since we already have a dedicated API Gateway service, we only use traffic routing to split our domain into `api.*` for the API, and `monitor.*` for Grafana. 

However, WAF is an optional feature, as it provides higher security in exchange for ~50% of request performance, so we either need to double our price by scaling out, or accept the performance lost. As such, we don't use it. We only use it primarily as a reverse proxy and traffic router.

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

Each services (a folder in the `services` folder) that need to be deployed will have its own pipeline defined. If any files in and only in that folder changes, it will trigger a build run, which `docker build` that image, then pushes to ACR, then trigger updates in ACA to create a new revision with `azure-cli`.

A revision is like an app update, config update or similar that doesn't change underlying infrastructure. Usually it's used to update our apps, like in this case, or even do A/B testing or multiple API versions, etc.

Rolling updates will be automatically done on revision change by ACA (it's just K8s) so no service disruption.

### Automation - IaC OpenTofu modules

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

### Cost calculation and optimization

We used the **Azure Pricing Calculator** to estimate our monthly costs based on a realistic workload analysis.

#### System Overview (8,000 customers + 2,000 drivers):

The system operates with three distinct traffic patterns:
- **Peak Hours** (06:30-08:00 & 16:00-18:00): 4 hours daily, 120 hours/month
- **Off-Peak Hours** (Business hours): 13.5 hours daily, 405 hours/month  
- **Night Time** (22:00-05:00): 7 hours daily, 210 hours/month

#### Traffic Analysis:

**Peak Hours (~3,591 RPS)**
- 90% drivers online (1,800 drivers)
- 20% customers online (1,600 customers)
- Driver location pings: 1,800 RPS
- Customer location pings: 1,600 RPS
- Nearby driver queries: 160 RPS
- Booking requests: 20-30 RPS

**Off-Peak Hours (~977 RPS)**
- 70% drivers online (1,400 drivers)
- 2% customers online (160 customers)
- Driver location pings: 800 RPS
- Customer location pings: 160 RPS
- Nearby driver queries: 16 RPS
- Booking requests: 1-2 RPS

**Night Time (~245 RPS)**
- 10% drivers online (200 drivers)
- 0.5% customers online (40 customers)
- Driver location pings: 200 RPS
- Customer location pings: 40 RPS
- Nearby driver queries: 4 RPS

#### Networking ($137.58/month)

**Application Gateway (Standard V2): $133.68/month**

The gateway operates 24/7 with auto-scaling based on Compute Units (CU). Each CU provides:
- 50 connections per second
- 2.22 Mbps throughput  
- 2,500 persistent connections

For 3,600 concurrent requests at peak:
- CU for compute: 3,600 CPS ÷ 50 = 72 CUs
- CU for throughput: (3,600 RPS × 5 KB × 8 bits) ÷ 2.22 Mbps ≈ 65 CUs
- CU for persistent connections: (3,600 × 2s avg response time) ÷ 2,500 = 3 CUs
- **Final CU**: 72 (max of above)

Monthly breakdown:
- Peak hours (105h): 72 CUs × 8 instances
- Off-peak (405h): 20 CUs × 2 instances  
- Night time (210h): 5 CUs × 1 instance

**Azure DNS: $3.90/month**

3 DNS zones for ACA, PostgreSQL, and Key Vault:
- 3 hosted zones: $1.50
- 6 million queries/month: $2.40

DNS queries are minimal due to:
- PostgreSQL and Key Vault connections use connection pooling (query once per replica startup)
- Application Gateway caches DNS with 60s TTL
- Average 30s query interval during peak

#### Azure Container Apps ($1,183.63/month)

All services use **3-year savings plan with 17% discount**. Free tier provides 180,000 vCPU-seconds and 360,000 GiB-seconds per month per ACA.

**Pricing:**
- Active vCPU: $0.000024/vCPU-second
- Active Memory: $0.000003/GiB-second

**Scaling strategies:**
- **HTTP Traffic scaling**: API Gateway, Location Service, Auth Service (based on concurrent requests)
- **Resource scaling**: Trip Service, User Service (based on CPU/Memory)
- **Event scaling**: Logger Service (based on Service Bus queue depth)

##### Service-by-Service Breakdown:

**API Gateway (0.5 vCPU, 1 GiB):**
- Night time (210h): 1 replica → $61.42
- Off-peak (405h): 1 replica → $476.67
- Peak hours (105h): 4 replicas (1,436 concurrent ÷ 400 target) → $448.88
- **Subtotal: $986.97**

**Location Service (0.5 vCPU, 1 GiB):**
- Night time: 1 replica → $11.34
- Off-peak: 1 replica → $21.87
- Peak hours: 1 replica → $5.67
- **Subtotal: $38.88**

**Auth Service (0.25 vCPU, 0.5 GiB):**
- Night time: 1 replica → $6.80
- Off-peak: 1 replica → $13.12
- Peak hours: 1 replica → $3.40
- **Subtotal: $23.32**

**Trip Service (0.25-0.5 vCPU, 1 GiB):**
- Scales on CPU load (max 50 RPS per replica before CPU > 50%)
- Night time: 1 replica → $13.60
- Off-peak: 1 replica → $26.24
- Peak hours: 1 replica → $8.80
- **Subtotal: $48.64**

**User Service (0.25-0.5 vCPU, 1 GiB):**
- Scales on CPU load (max 100 RPS per replica before CPU > 50%)
- Night time: 1 replica → $13.60
- Off-peak: 1 replica → $26.24
- Peak hours: 2 replicas → $13.60
- **Subtotal: $53.44**

**Logger Service (Event-based, 0.25 vCPU, 0.5 GiB):**
- Scales on Service Bus queue depth (500 msg/s per replica)
- Night time: 1 replica (246 msg/s) → $11.34
- Off-peak: 2 replicas (996 msg/s) → $43.74
- Peak hours: 8 replicas (3,630 msg/s) → $45.36
- **Subtotal: $100.44**

**Observability Stack:**
- Alloy (0.5 vCPU, 1 GiB): $38.88
- Loki (0.5 vCPU, 1 GiB): $38.88
- Tempo (0.5 vCPU, 1 GiB): $38.88
- Prometheus (0.5 vCPU, 1 GiB): $38.88
- Grafana (0.25 vCPU, 0.5 GiB): $1.00 (scale-to-zero when not in use)
- **Subtotal: $156.52**

**Total before discount**: $1,426.46  
**After 17% savings plan discount**: **$1,183.63**

#### Database ($681.70/month)

**Azure Database for PostgreSQL - Flexible Server:**

We opted for **single server mode** instead of Citus-based horizontal scaling because:
- Multi-node clusters require expensive upfront costs
- Sharding requires data model changes and proper sharding strategies
- Our use case doesn't warrant horizontal scaling yet
- Can upgrade to Citus elastic cluster when needed
- Vertical scaling provides 99.9% SLA with near-zero downtime

Configuration:
- 2 instances (Auth Service, Trip Service): D4ds v6 (4 vCores), 128 GiB storage each
- High Availability enabled with zone-redundant standby
- **Cost**: $538.31 × 2 = **$1,076.62/month** (Not used initially due to budget - will deploy with single instance)

**Azure Managed Redis (Standard B1):**

- 1 GB storage (only 19.53 MB needed for 10k location records)
- 2 vCPUs, sufficient throughput for workload
- High Availability enabled
- **Cost: $57.20/month**

**Azure DocumentDB for MongoDB:**

Two databases without High Availability (buffer mechanism and Service Bus provide data durability):

*User Service MongoDB (M20 tier):*
- 32 GB storage (sufficient for 10k users + 6k vehicles = ~13 GB actual)
- Always-on service (730 hours/month)
- **Cost: $154.25/month**

*Logger Service MongoDB (M30 tier, first month):*
- Stores audit logs only (not location pings)
- Monthly audit logs: 98.1 million requests × 1 KB = 91.4 GiB
- 128 GiB storage tier
- **First month cost: $470.00**

**Storage growth projection:**

| Month | Data (GiB) | Storage Tier | Cluster | Monthly Cost |
|-------|-----------|--------------|---------|--------------|
| 1 | 91.4 | 128 GiB | M30 | $470 |
| 2 | 182.8 | 256 GiB | M30 | $502 |
| 3 | 274.2 | 512 GiB | M30 | $566 |
| 6 | 548.4 | 1 TiB | M40 | $1,205 |
| 12 | 1096.8 | 2 TiB | M50 | $2,118 |
| 28 | 2559.2 | 4 TiB | M60 | $4,309 |

**Total Database (first month)**: **$681.70** (Note: PostgreSQL cost reduced to $0 for initial deployment, actual with HA: $1,758.32)

#### Additional Azure Services ($10.30/month)

**Azure Service Bus (Standard tier): $10.00/month**

Base cost $10 + $0.80 per million operations beyond free 12.5 million/month.

Monthly operations breakdown:
- Trip lifecycle events (2,000 trips/day × 12 ops/trip): 720,000 ops
- User audit logs (10,000 users × 2 actions/day × 2 ops): 1,200,000 ops  
- Driver status changes (2,000 drivers × 4 toggles/day × 3 ops): 720,000 ops
- **Total: 2,640,000 ops/month** (well under 12.5M free tier)

**Azure Key Vault: $0.30/month**

Operations estimate:
- Secrets read on replica startup: 200/day
- Secrets write for updates: 10/day
- **Total: 6,300 ops/month** (under 10,000 ops = $0.30)

Key Vault is only accessed:
- Once per replica startup (connection pooling)
- On connection failures/restarts
- Manual secret rotation (rare)

#### Persistent Storage ($30.50/month)

**Azure Block Blob Storage (Standard tier)** for observability stack persistence:

*Loki & Tempo (2 instances):*
- 10 GB each (stores only errors/critical traces)
- 10k writes, 1k reads per month
- **Cost: $0.25 × 2 = $0.50**

*Grafana & Prometheus (2 instances):*
- 500 GB each (time-series metrics)
- 1M writes, 1k reads per month
- **Cost: $15.00 × 2 = $30.00**

Note: Block Blob includes free data writes and auto-scales capacity.

#### Cost Optimization Strategies

1. **Replica scaling**: Many small replicas > few large replicas
2. **Min replica always = 1** (except Grafana which scales to zero)
3. **3-year savings plan**: 17% discount on all ACA compute
4. **Connection pooling**: Minimizes DNS/Key Vault query costs
5. **Scale-to-zero**: Logger Service and Grafana during off-hours
6. **Standard tiers**: Balanced performance/cost (no Premium SKUs)
7. **No NAT Gateway**: HERE Maps SDK doesn't require static IP whitelisting (saves ~$2,000/month)
8. **HA selective**: Only enabled for critical databases (Redis, PostgreSQL)

#### Total Monthly Cost (First Month)

| Category | Cost |
|----------|------|
| Networking | $137.58 |
| Azure Container Apps | $1,183.63 |
| Database | $681.70 |
| Additional Services | $10.30 |
| Persistent Storage | $30.50 |
| **TOTAL** | **$2,043.71/month** |

**Notes:**
- Cost excludes PostgreSQL HA (add $1,076.62 for production-ready setup)
- Logger Service MongoDB scales linearly with audit log volume
- Application Gateway auto-scales based on traffic patterns
- All services in Southeast Asia region
- 24/7 monitoring and observability included