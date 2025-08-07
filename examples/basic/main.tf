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

provider "ploicloud" {
  api_token = var.ploi_api_token
}

resource "ploicloud_application" "main" {
  name = "basic-laravel-app"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
  
  build_commands = [
    "composer install --no-dev --optimize-autoloader"
  ]
  
  init_commands = [
    "php artisan migrate --force"
  ]
}

resource "ploicloud_service" "mysql" {
  application_id = ploicloud_application.main.id
  type          = "mysql"
  version       = "8.0"
  
  settings = {
    database = "production"
    size     = "5Gi"
  }
}

resource "ploicloud_secret" "app_key" {
  application_id = ploicloud_application.main.id
  key           = "APP_KEY"
  value         = "base64:your-app-key-here"
}

output "application_url" {
  value = ploicloud_application.main.url
}