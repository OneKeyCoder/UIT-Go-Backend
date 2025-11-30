# Data plane - Postgres

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

## Why no sharding? How about horizontal scaling?

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

### References

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
