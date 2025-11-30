# No NAT Gateway

Usually, a VNet would have a NAT Gateway to handle egress traffic (connections that started from within our infra) too, but since we don't need the static IP (HERE SDK does not require IP whitelisting) and the floor monthly cost is significantly higher than default ACA egress (around 2000$ for a study case we found), we decided NOT to include it.
