# Data plane - Redis Cache

Azure Managed Redis was picked. We could, and wanted to, use Valkey instead, since it's completely open-source and not bound to an enterprise entity (backed by Linux Foundation), but since unlike AWS, Azure does not provide a Valkey Managed service, we use Redis instead.

Self hosting Redis instance inside ACA with Azure Disk persistent is also an option, but overall not worth the maintanance effort.
