# ploicloud_service Resource

Manages a service for a Ploi Cloud application (database, cache, etc.).

## Example Usage

### MySQL Database

```terraform
resource "ploicloud_service" "mysql" {
  application_id = ploicloud_application.main.id
  type          = "mysql"
  version       = "8.0"
  
  settings = {
    database = "production"
    size     = "10Gi"
  }
}
```

### Redis Cache

```terraform
resource "ploicloud_service" "redis" {
  application_id = ploicloud_application.main.id
  type          = "redis"
  version       = "7.0"
}
```

### PostgreSQL Database

```terraform
resource "ploicloud_service" "postgres" {
  application_id = ploicloud_application.main.id
  type          = "postgresql"
  version       = "16"
  
  settings = {
    database = "myapp"
    size     = "20Gi"
  }
}
```

### RabbitMQ Message Queue

```terraform
resource "ploicloud_service" "rabbitmq" {
  application_id = ploicloud_application.main.id
  type          = "rabbitmq"
  version       = "3.12"
  
  settings = {
    size = "5Gi"
  }
}
```

## Schema

### Required

- `application_id` (Number) - Application ID this service belongs to
- `type` (String) - Service type. Valid values: `mysql`, `postgresql`, `redis`, `valkey`, `rabbitmq`, `mongodb`, `minio`, `sftp`

### Optional

- `version` (String) - Service version
- `settings` (Map of String) - Service-specific settings (e.g., database name, storage size)

### Read-Only

- `id` (Number) - Service ID
- `status` (String) - Service status

## Import

Services can be imported using the format `application_id.service_id`:

```bash
terraform import ploicloud_service.mysql 12345.67890
```