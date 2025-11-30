# Security - Full managed vs Virtual Network

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
