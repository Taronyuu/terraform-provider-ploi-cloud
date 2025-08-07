terraform {
  required_providers {
    ploicloud = {
      source  = "ploi/ploicloud"
      version = "~> 1.0"
    }
  }
}

variable "ploi_api_token" {
  description = "Ploi Cloud API token"
  type        = string
  sensitive   = true
}

variable "app_key" {
  description = "Laravel application key"
  type        = string
  sensitive   = true
}

provider "ploicloud" {
  api_token = var.ploi_api_token
}

resource "ploicloud_application" "api" {
  name                = "complete-laravel-api"
  type                = "laravel"
  application_version = "11.x"
  
  runtime {
    php_version    = "8.4"
    nodejs_version = "22"
  }
  
  build_commands = [
    "composer install --no-dev --optimize-autoloader",
    "npm ci",
    "npm run build"
  ]
  
  init_commands = [
    "php artisan migrate --force",
    "php artisan config:cache",
    "php artisan route:cache",
    "php artisan view:cache"
  ]
  
  settings {
    health_check_path  = "/api/health"
    scheduler_enabled  = true
    replicas          = 3
    cpu_request       = "1"
    memory_request    = "2Gi"
  }
  
  php_extensions = ["redis", "gd", "zip", "bcmath", "intl"]
  
  php_settings = [
    "memory_limit = 256M",
    "max_execution_time = 120",
    "upload_max_filesize = 32M"
  ]
}

resource "ploicloud_service" "mysql" {
  application_id = ploicloud_application.api.id
  type          = "mysql"
  version       = "8.0"
  
  settings = {
    database = "production"
    size     = "20Gi"
  }
}

resource "ploicloud_service" "redis" {
  application_id = ploicloud_application.api.id
  type          = "redis"
  version       = "7.0"
}

resource "ploicloud_service" "rabbitmq" {
  application_id = ploicloud_application.api.id
  type          = "rabbitmq"
  version       = "3.12"
  
  settings = {
    size = "10Gi"
  }
}

resource "ploicloud_worker" "default_queue" {
  application_id = ploicloud_application.api.id
  name          = "default-queue"
  command       = "php artisan queue:work --queue=default --sleep=3 --tries=3"
  replicas      = 2
}

resource "ploicloud_worker" "email_queue" {
  application_id = ploicloud_application.api.id
  name          = "email-queue"
  command       = "php artisan queue:work --queue=emails --sleep=1 --tries=5"
  replicas      = 1
}

resource "ploicloud_volume" "storage" {
  application_id = ploicloud_application.api.id
  name          = "app-storage"
  size          = 50
  mount_path    = "/var/www/html/storage/app"
}

resource "ploicloud_volume" "uploads" {
  application_id = ploicloud_application.api.id
  name          = "uploads"
  size          = 100
  mount_path    = "/var/www/html/public/uploads"
}

resource "ploicloud_domain" "api" {
  application_id = ploicloud_application.api.id
  domain        = "api.example.com"
}

resource "ploicloud_domain" "api_www" {
  application_id = ploicloud_application.api.id
  domain        = "www.api.example.com"
}

resource "ploicloud_secret" "app_key" {
  application_id = ploicloud_application.api.id
  key           = "APP_KEY"
  value         = var.app_key
}

resource "ploicloud_secret" "db_password" {
  application_id = ploicloud_application.api.id
  key           = "DB_PASSWORD"
  value         = "secure-password-123"
}

resource "ploicloud_secret" "mail_password" {
  application_id = ploicloud_application.api.id
  key           = "MAIL_PASSWORD"
  value         = "mail-service-password"
}

output "application_url" {
  description = "Primary application URL"
  value       = "https://${ploicloud_domain.api.domain}"
}

output "application_status" {
  description = "Current application status"
  value       = ploicloud_application.api.status
}

output "needs_deployment" {
  description = "Whether the application needs deployment"
  value       = ploicloud_application.api.needs_deployment
}