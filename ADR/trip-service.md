# Trip Service

## Which database should use? Why?

We chose PostgreSQL for the Trip Service because of:
- ACID transaction support, which is critical for trip data where consistency is paramount - a trip can only have one driver, and financial data (fare, payments) must be accurate. 
- The team already has experience with PostgreSQL from the Authentication Service, allowing faster development. 
- PostgreSQL also offers mature tooling, JSON support for flexible fields, and can be deployed cost-effectively on Azure Database for PostgreSQL.

## Why use HERE Maps API for route calculation?

We chose HERE Maps API over alternatives like Google Maps because of:
- Its generous free tier (250,000 requests/month). 
- HERE provides accurate road distance calculation (not straight-line) with ETA that includes traffic data. 
- Most importantly, HERE has good coverage for Vietnam, unlike some competitors like Mapbox or Google Maps. The API is RESTful and easy to integrate with OAuth token authentication. While Google Maps is the industry standard, HERE offers the best balance between accuracy, cost, and regional coverage for our use case.
