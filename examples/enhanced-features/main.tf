terraform {
  required_providers {
    ploicloud = {
      source = "ploi/ploicloud"
    }
  }
}

# Configure the Ploi Cloud provider
provider "ploicloud" {
  # Configuration will be provided via environment variables
  # PLOI_API_KEY or PLOI_API_TOKEN
  # PLOI_API_BASE_URL (optional, defaults to production API)
}

# Example application with enhanced features
resource "ploicloud_application" "example" {
  name = "enhanced-app"
  type = "laravel"

  # Enhanced: Custom start command
  start_command = "php artisan serve --host=0.0.0.0 --port=8080"

  runtime {
    php_version = "8.3"
  }

  settings {
    replicas       = 2
    cpu_request    = "500m"
    memory_request = "1Gi"
  }

  build_commands = [
    "composer install --no-dev --optimize-autoloader",
    "npm ci",
    "npm run production"
  ]

  init_commands = [
    "php artisan config:cache",
    "php artisan route:cache",
    "php artisan view:cache"
  ]
}

# Enhanced PostgreSQL service with extensions and storage
resource "ploicloud_service" "postgres" {
  application_id = ploicloud_application.example.id
  type           = "postgresql"
  version        = "15"

  # Enhanced: PostgreSQL extensions
  extensions = [
    "uuid-ossp",
    "pgcrypto",
    "citext"
  ]

  # Enhanced: Custom storage size
  storage_size = "20Gi"

  # Enhanced: Resource requests
  memory_request = "2Gi"
  cpu_request    = "1000m"
}

# Enhanced Redis service with custom resources
resource "ploicloud_service" "redis" {
  application_id = ploicloud_application.example.id
  type           = "redis"
  version        = "7"

  # Enhanced: Resource configuration
  memory_request = "512Mi"
  cpu_request    = "250m"
  storage_size   = "5Gi"
}

# Enhanced worker with custom type and resources
resource "ploicloud_worker" "queue_worker" {
  application_id = ploicloud_application.example.id
  name           = "queue-processor"
  command        = "php artisan queue:work --verbose --tries=3 --timeout=60"

  # Enhanced: Worker type
  type = "queue"

  # Enhanced: Resource requests
  replicas       = 3
  memory_request = "512Mi"
  cpu_request    = "250m"
}

# Enhanced worker for scheduled tasks
resource "ploicloud_worker" "scheduler" {
  application_id = ploicloud_application.example.id
  name           = "task-scheduler"
  command        = "php artisan schedule:work"

  # Enhanced: Different worker type
  type = "scheduler"

  replicas       = 1
  memory_request = "256Mi"
  cpu_request    = "100m"
}

# Enhanced volume with storage class
resource "ploicloud_volume" "app_storage" {
  application_id = ploicloud_application.example.id
  name           = "app-data"
  size           = 50
  mount_path     = "/app/storage"

  # Enhanced: Storage class specification
  storage_class = "fast-ssd"
}

# Enhanced volume for uploads
resource "ploicloud_volume" "uploads" {
  application_id = ploicloud_application.example.id
  name           = "uploads"
  size           = 100
  mount_path     = "/app/public/uploads"

  # Enhanced: Different storage class
  storage_class = "standard"
}

# Application secrets
resource "ploicloud_secret" "app_key" {
  application_id = ploicloud_application.example.id
  key            = "APP_KEY"
  value          = "base64:${base64encode("your-secret-app-key-here")}"
}

resource "ploicloud_secret" "db_connection" {
  application_id = ploicloud_application.example.id
  key            = "DATABASE_URL"
  value          = "postgresql://user:password@${ploicloud_service.postgres.service_name}:5432/database"
}

# Outputs to show enhanced features
output "application_info" {
  value = {
    id            = ploicloud_application.example.id
    url           = ploicloud_application.example.url
    start_command = ploicloud_application.example.start_command
  }
}

output "postgres_service" {
  value = {
    id           = ploicloud_service.postgres.id
    extensions   = ploicloud_service.postgres.extensions
    storage_size = ploicloud_service.postgres.storage_size
    resources = {
      cpu    = ploicloud_service.postgres.cpu_request
      memory = ploicloud_service.postgres.memory_request
    }
  }
}

output "worker_details" {
  value = {
    queue_worker = {
      id        = ploicloud_worker.queue_worker.id
      type      = ploicloud_worker.queue_worker.type
      replicas  = ploicloud_worker.queue_worker.replicas
      resources = {
        cpu    = ploicloud_worker.queue_worker.cpu_request
        memory = ploicloud_worker.queue_worker.memory_request
      }
    }
    scheduler = {
      id        = ploicloud_worker.scheduler.id
      type      = ploicloud_worker.scheduler.type
      replicas  = ploicloud_worker.scheduler.replicas
      resources = {
        cpu    = ploicloud_worker.scheduler.cpu_request
        memory = ploicloud_worker.scheduler.memory_request
      }
    }
  }
}

output "volume_details" {
  value = {
    app_storage = {
      id            = ploicloud_volume.app_storage.id
      storage_class = ploicloud_volume.app_storage.storage_class
      size         = ploicloud_volume.app_storage.size
    }
    uploads = {
      id            = ploicloud_volume.uploads.id
      storage_class = ploicloud_volume.uploads.storage_class
      size         = ploicloud_volume.uploads.size
    }
  }
}