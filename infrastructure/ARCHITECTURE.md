# dm cloud

MS docs are starting to be written by AI. Be careful.

## Vnet

https://learn.microsoft.com/en-us/azure/container-apps/custom-virtual-networks?tabs=workload-profiles-env

## compute

Run the services on Azure Container Apps. Basically managed Kubernetes. Give you rolling updates for apps and stuff. PaaS.

https://learn.microsoft.com/en-us/azure/container-apps/compare-options

### Plans

Use consumption profile because we have burst request pattern, and unpredictable workloads.  
https://learn.microsoft.com/en-us/azure/container-apps/workload-profiles-overview#profile-types

## MongoDB

I do not know why we picked MongoDB for logs, but *I couldnt care less.*

Use Azure DocumentDB (MongoDB). Separate network. Managed and serverless, pay as you go.  
Renamed from Azure CosmosDB for MongoDB.

Alternatives: MongoDB Atlas Managed, MongoDB Atlas on Azure,...

Ref:
- https://learn.microsoft.com/en-us/azure/documentdb/compare-mongodb-atlas
- https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-vnet-service-endpoint

## Postgres

Use Azure Database for PostgreSQL. 

Single master server, with read-only replicas spread out in the same AZ, and has zone-redundant with a secondary fall-over server in another AZ. Or, if data is really important and have the budget to spare, in another region with geo-redundant. Geo-redundant only requires paying more for the cross-region transfer, server costs should be the same.

Also supports vertical scaling with near-zero downtime (~30 seconds to reboot, or a VM-swap zero-downtime if low traffic).

Alternatives: Azure CosmosDB for PostgreSQL, Neon Managed PostgreSQL in Azure, Azure HorizonDB (Preview tho, basically Azure SQL hyperscale but postgres).  
All out-of-band services like Render, Neon,... are NOT considered because of high delay of networking and additional costs.

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

https://learn.microsoft.com/en-us/answers/questions/1067211/azure-cosmos-db-for-postgresql-vs-azure-database-f

### Why no sharding? How about horizontal scaling?

The service itself supports scaling horizontally with sharding through Citus plugin, but I decided against using it because:

- Very expensive upfront cost (multiple nodes, multiple HA replicas)
- Use case does not warrant the cost (yet)
- Sharding requires data changes or more modifications to have proper sharding performance
- We do NOT have multi-tenant usecase, so sharding will be very hard to get right, or outright impossible
- We can always upgrade to it later when it's needed
- Do not have proper subnetting
- Cannot share DBs in the same server to preserve resources

https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-elastic-clusters-limitations

However, we can still implement read-only replicas, which technically is a way to scale horizontally, but only alleviates read workload and comes at a cost of slight delay from data replication, only for the Auth service as we do not have high writes there.

## Secrets

Azure Key Vault.

Then connect to ACA to provide the secrets.

No need for alternatives. But if you insist, bitnami sealed secrets, and other key vaults.

## Grafana and friends

## 