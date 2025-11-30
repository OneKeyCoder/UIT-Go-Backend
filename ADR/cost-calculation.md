# Cost calculation and optimization

We used the **Azure Pricing Calculator** to estimate our monthly costs.

## Calculation Assumptions (10k users):
- Peak concurrent users: 1,500-2,500 (15-25% of total users) - higher during peak hours
- Peak concurrent trips: 300-500 concurrent trips (3-5% of users are on an active trip)
- Request rate: 80-150 requests/user/day (higher due to location updates)
- Data transfer: 8-15GB/day.
- Region: Southeast Asia
- During peak hours, 10k users simultaneously, each user sending at least 1 request per second on average (for location updates). Peak hours last 2 hours in the afternoon and 2 hours in the morning. Thus, during each peak hour, the application must handle 10,000 requests per second.

## Networking

- **Application Gateway**: $195.64/month at Standard V2 tier, including 730 fixed gateway hours, 1 compute unit, 1000 persistent connections, 1 mb/s throughput, and 100 GB data transfer.  
  Note: Can be scaled.  
  Scale from 1 compute - 2,500 Persistent Connections - Throughput 2.22 mb/s  
  Scale to 10 compute - 25,000 Persistent Connections - Throughput 22.2 mb/s  
  Estimated Price: $219.00/month

- **Azure DNS (DNS ACA)**: $40.50/month, including 1 hosted DNS zone and 100 million DNS queries.

- **Azure DNS (DNS Postgres)**: $40.50/month, including 1 hosted DNS zone and 100 million DNS queries.

- **Load Balancer (Internal load balancer)**: Free at Basic tier.

- **Azure DNS (DNS Key Vault)**: $40.50/month, including 1 hosted DNS zone and 100 million DNS queries.

## Azure Container Apps

- **Location Service**: $52.35/month, Consumption plan, 0 million requests, 2 vCPUs, 4 GiB memory, 1 minimum replica.

- **Trip Service**: $105.12/month, Consumption plan, 0 million requests, 2 vCPUs, 8 GiB memory, 1 minimum replica.

- **API Gateway**: $642.24/month, Consumption plan, 1500 million requests, 4 vCPUs, 8 GiB memory, 1 minimum replica.

- **Auth Service**: $52.35/month, Consumption plan, 0 million requests, 2 vCPUs, 4 GiB memory, 1 minimum replica.

- **User Service**: $0.00/month, Consumption plan, 0 million requests, 2 vCPUs, 4 GiB memory.

- **Logger Service**: $0.00/month, Consumption plan, 0 million requests, 0.5 vCPU, 2 GiB memory.

- **Grafana**: $0.00/month, Consumption plan, 0 million requests, 0.5 vCPU, 1 GiB memory.

- **Observability stack**: $31.54/month, Consumption plan, 0 million requests, 1 vCPU, 2 GiB memory, 1 minimum replica.

Total for Azure Container Apps: $886.25/month.

Note: The calculations above are not entirely accurate as they are from the Azure calculator and do not account for internal task computations by the services (The current calculating assume that every service have been scaled to 0 due to the "pay as you go" plan). The formula for computing costs per service is:  
- vCPU cost: Number of vCPUs × active seconds × $0.00002822 per vCPU-s  
- Memory cost: Number of GiB memory × active seconds × $0.00000332 per GiB-s
- Due to the Resources: The first 180,000 vCPU-seconds each month are free.

### Recalculated Costs with Assumptions
Assuming each service has at least 1 active instance for 18 hours per day (64,800 seconds per day, or 1,944,000 seconds per month), recalculated using the formula above, accounting for free tiers (180,000 vCPU-seconds and 360,000 GiB-seconds per month per ACA).

Service vCPU allocations:  
- Location Service: 2 vCPUs  
- Trip Service: 2 vCPUs  
- API Gateway: 4 vCPUs  
- Auth Service: 2 vCPUs  
- User Service: 2 vCPUs  
- Logger Service: 0.5 vCPUs  
- Grafana: 0.5 vCPUs  
- Observability stack: 1 vCPU  

Total vCPUs across all services: 14.5  

Service memory allocations:  
- Location Service: 4 GiB  
- Trip Service: 8 GiB  
- API Gateway: 8 GiB  
- Auth Service: 4 GiB  
- User Service: 4 GiB  
- Logger Service: 2 GiB  
- Grafana: 1 GiB  
- Observability stack: 2 GiB  

Total memory across all services: 33 GiB  

Total vCPU-seconds per month: 14.5 × 1,944,000 = 28,188,000  
Total GiB-seconds per month: 33 × 1,944,000 = 64,152,000  

Free vCPU-seconds per ACA per month: 180,000  
Free GiB-seconds per ACA per month: 360,000  

Per service costs (after subtracting free per ACA):  

- **Location Service** (2 vCPUs, 4 GiB): vCPU-s = 3,888,000; Billable vCPU-s = 3,888,000 - 180,000 = 3,708,000; vCPU cost = $104.67; GiB-s = 7,776,000; Billable GiB-s = 7,776,000 - 360,000 = 7,416,000; Memory cost = $24.60; Total: $129.27  
- **Trip Service** (2 vCPUs, 8 GiB): vCPU-s = 3,888,000; Billable = 3,708,000; vCPU cost = $104.67; GiB-s = 15,552,000; Billable = 15,192,000; Memory cost = $50.44; Total: $155.11  
- **API Gateway** (4 vCPUs, 8 GiB): vCPU-s = 7,776,000; Billable = 7,596,000; vCPU cost = $214.47; GiB-s = 15,552,000; Billable = 15,192,000; Memory cost = $50.44; Total: $264.91  
- **Auth Service** (2 vCPUs, 4 GiB): vCPU-s = 3,888,000; Billable = 3,708,000; vCPU cost = $104.67; GiB-s = 7,776,000; Billable = 7,416,000; Memory cost = $24.60; Total: $129.27  
- **User Service** (2 vCPUs, 4 GiB): vCPU-s = 3,888,000; Billable = 3,708,000; vCPU cost = $104.67; GiB-s = 7,776,000; Billable = 7,416,000; Memory cost = $24.60; Total: $129.27  
- **Logger Service** (0.5 vCPUs, 2 GiB): vCPU-s = 972,000; Billable = 972,000 - 180,000 = 792,000; vCPU cost = $22.37; GiB-s = 3,888,000; Billable = 3,888,000 - 360,000 = 3,528,000; Memory cost = $11.71; Total: $34.08  
- **Grafana** (0.5 vCPUs, 1 GiB): vCPU-s = 972,000; Billable = 792,000; vCPU cost = $22.37; GiB-s = 1,944,000; Billable = 1,944,000 - 360,000 = 1,584,000; Memory cost = $5.26; Total: $27.63  
- **Observability stack** (1 vCPU, 2 GiB): vCPU-s = 1,944,000; Billable = 1,944,000 - 180,000 = 1,764,000; vCPU cost = $49.79; GiB-s = 3,888,000; Billable = 3,528,000; Memory cost = $11.71; Total: $61.50  

Total recalculated active cost: $958.25/month (after free tiers).

## Database

- **Azure Database for PostgreSQL (db auth service)**: $538.31/month, Flexible Server, General Purpose, D4ds v6 (4 vCores), 128 GiB storage, 1000 IOPS, with High Availability.

- **Azure Database for PostgreSQL (db trip service)**: $538.31/month, Flexible Server, General Purpose, D4ds v6 (4 vCores), 128 GiB storage, 1000 IOPS, with High Availability.

- **Azure DocumentDB (MongoDB, db logger service)**: $107.74/month, M10 cluster, 3 Shards, 128 GB storage, without High Availability.

- **Azure DocumentDB (MongoDB, db user service)**: $271.12/month, M20 cluster, 3 Shards, 128 GB storage, without High Availability.

- **Azure Managed Redis (db location service)**: $794.24/month, Compute Optimized, 2 x X3 instances, with High Availability.

Total for Database: $2,249.72/month.

Note: We understand the importance of user data, so we always have data backup in case of bad situations. At the same time, the 2 databases for trip service and location service need to be always ready at all times to support users booking trips, so we chose High Availability (the other 2 databases do not have high availability).

## Additional Azure Services

- **Azure Service Bus**: $100.00/month, Basic tier, 2000 million messaging operations.

- **Azure Key Vault**: $6.00/month, Vault with 2,000,000 operations.

Total for Additional Azure Services: $106.00/month.

## Module E Total Cost

- **Networking**: $317.14/month (Application Gateway $195.64, Azure DNS $121.50, Load Balancer $0.00)
- **Azure Container Apps**: $958.25/month (18 hours per day)
- **Database**: $2,249.72/month
- **Additional Azure Services**: $106.00/month

**Total Estimated Monthly Cost**: $3,631.11/month

## Azure Pricing Calculator References
- Networking: https://azure.com/e/5689822e50e446cb87bd831f55edd0a1
- Azure Container Apps: https://azure.com/e/7cc01f411f674392b92729da5129515c
- Database: https://azure.com/e/3a1155332a6746c7b3d6fae464c0df67
- Additional Services: https://azure.com/e/c92b0f63e6564abb92a6a77b155f26b8
