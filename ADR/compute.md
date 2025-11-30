# ADR: Compute

There is many choices to deploy the "compute" part of our app. Since we have a Dockerized project, we have 3 main ways to deploy the system:

- Raw Azure VM with *Scale set* for auto scaling.