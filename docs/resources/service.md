# ploicloud_service Resource

Manages a service for a Ploi Cloud application (database, cache, worker, etc.).

## Example Usage

### MySQL Database

```terraform
resource "ploicloud_service" "mysql" {
  application_id = ploicloud_application.main.id
  service_name   = "main-db"
  type          = "mysql"
  version       = "8.0"
  storage_size   = "10Gi"
  memory_request = "2Gi"
}
```

### Redis Cache

```terraform
resource "ploicloud_service" "redis" {
  application_id = ploicloud_application.main.id
  service_name   = "cache"
  type          = "redis"
  version       = "7.0"
  storage_size   = "1Gi"
  memory_request = "1Gi"
}
```

### PostgreSQL Database

```terraform
resource "ploicloud_service" "postgres" {
  application_id = ploicloud_application.main.id
  service_name   = "postgres-db"
  type          = "postgresql"
  version       = "16"
  storage_size   = "20Gi"
  memory_request = "2Gi"
  
  settings = {
    extensions = ["uuid-ossp", "pg_trgm"]
  }
}
```

### RabbitMQ Message Queue

```terraform
resource "ploicloud_service" "rabbitmq" {
  application_id = ploicloud_application.main.id
  service_name   = "message-queue"
  type          = "rabbitmq"
  version       = "3.12"
  storage_size   = "5Gi"
  memory_request = "1Gi"
}
```

### MongoDB Database

```terraform
resource "ploicloud_service" "mongodb" {
  application_id = ploicloud_application.main.id
  service_name   = "mongo-db"
  type          = "mongodb"
  version       = "7.0"
  storage_size   = "15Gi"
  memory_request = "2Gi"
}
```

### Valkey Cache

```terraform
resource "ploicloud_service" "valkey" {
  application_id = ploicloud_application.main.id
  service_name   = "valkey-cache"
  type          = "valkey"
  version       = "7.2"
  storage_size   = "2Gi"
  memory_request = "1Gi"
}
```

### MinIO Object Storage

```terraform
resource "ploicloud_service" "minio" {
  application_id = ploicloud_application.main.id
  service_name   = "object-storage"
  type          = "minio"
  version       = "latest"
  storage_size   = "50Gi"
  memory_request = "2Gi"
}
```

### Worker Service

```terraform
resource "ploicloud_service" "worker" {
  application_id = ploicloud_application.main.id
  service_name   = "queue-worker"
  type          = "worker"
  replicas      = 2
  memory_request = "1Gi"
  
  settings = {
    command = "php artisan queue:work"
  }
}
```

## Schema

### Required

- `application_id` (Number) - Application ID this service belongs to
- `type` (String) - Service type. Valid values: `mysql`, `postgresql`, `redis`, `valkey`, `rabbitmq`, `mongodb`, `minio`, `sftp`, `worker`

### Optional

- `service_name` (String) - Custom service name
- `version` (String) - Service version (required for database/cache services)
- `storage_size` (String) - Storage allocation (required for database/cache/storage services)
- `memory_request` (String) - Memory allocation (required for all services)
- `replicas` (Number) - Number of replicas (for worker services only). Defaults to `1`
- `settings` (Map of String) - Service-specific settings:
  - **PostgreSQL**: `extensions` (list of extensions to enable)
  - **Workers**: `command` (command to execute)

### Read-Only

- `id` (Number) - Service ID
- `status` (String) - Service status

## Import

Services can be imported using the format `application_id.service_id`:

```bash
terraform import ploicloud_service.mysql 12345.67890
```