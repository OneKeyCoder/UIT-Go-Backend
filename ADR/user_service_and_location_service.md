# User Service
- Which database should use? Why?

We chose MongoDB for the User Service because of its high read and write speeds. Secondly, the database is not large, does not have too many rules, and only needs to store two tables: the user table and the vehicle table, so MongoDB provides the necessary speed. Secondly, we are familiar with MongoDB, allowing us to develop quickly and focus more on the overall system architecture.

# Location Service
- Why use Redis but not DynamoDb?

We understand the importance of scaling, however the most important is response speed because this is an extremely important factor to satisfy users (the main reason we chose Redis). Secondly, Redis has Geo Cache, which makes location search faster and more importantly, it reduces a lot of pressure on the CPU. Thirdly, we have experience from many projects working with Redis, so the deployment time is also faster.
