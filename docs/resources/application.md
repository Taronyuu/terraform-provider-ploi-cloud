# ploicloud_application Resource

Manages a Ploi Cloud application.

## Example Usage

### Basic Laravel Application

```terraform
resource "ploicloud_application" "main" {
  name = "my-laravel-app"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
  
  build_commands = [
    "composer install --no-dev --optimize-autoloader",
    "npm ci",
    "npm run build"
  ]
  
  init_commands = [
    "php artisan migrate --force",
    "php artisan config:cache"
  ]
}
```

### Laravel Application with Settings

```terraform
resource "ploicloud_application" "api" {
  name                = "api-backend"
  type                = "laravel"
  application_version = "11.x"
  
  start_command = "php artisan serve --host=0.0.0.0 --port=8000"
  
  additional_domains = [
    "api.example.com",
    "www.api.example.com"
  ]
  
  runtime {
    php_version    = "8.4"
    nodejs_version = "22"
  }
  
  settings {
    health_check_path  = "/api/health"
    scheduler_enabled  = true
    replicas          = 3
    memory_request    = "2Gi"
  }
  
  php_extensions = ["redis", "gd", "zip", "bcmath"]
  
  php_settings = [
    "memory_limit = 256M",
    "max_execution_time = 60"
  ]
}
```

### WordPress Application

```terraform
resource "ploicloud_application" "blog" {
  name = "my-blog"
  type = "wordpress"
  
  settings {
    replicas = 2
  }
}
```

## Schema

### Required

- `name` (String) - Application name
- `type` (String) - Application type. Valid values: `laravel`, `wordpress`, `statamic`, `craftcms`, `nodejs`

### Optional

- `application_version` (String) - Application version (e.g., 11.x for Laravel)
- `build_commands` (List of String) - Build commands to run during image build
- `init_commands` (List of String) - Initialization commands to run before starting the application
- `start_command` (String) - Custom command to start the application
- `additional_domains` (List of String) - Additional custom domains for the application
- `php_extensions` (List of String) - PHP extensions to install
- `php_settings` (List of String) - PHP ini settings
- `repository_url` (String) - Repository URL
- `repository_owner` (String) - Repository owner
- `repository_name` (String) - Repository name
- `default_branch` (String) - Default git branch. Defaults to `main`
- `social_account_id` (Number) - Social account ID for git integration
- `region` (String) - Region to deploy the application. Defaults to `default`
- `provider` (String) - Cloud provider. Defaults to `default`

### Nested Schema for `runtime`

- `php_version` (String) - PHP version. Valid values: `7.4`, `8.0`, `8.1`, `8.2`, `8.3`, `8.4`
- `nodejs_version` (String) - Node.js version. Valid values: `18`, `20`, `22`, `24`

### Nested Schema for `settings`

- `health_check_path` (String) - Health check path. Defaults to `/`
- `scheduler_enabled` (Boolean) - Enable Laravel scheduler. Defaults to `false`
- `replicas` (Number) - Number of replicas. Defaults to `1`
- `memory_request` (String) - Memory request. Defaults to `512Mi`

### Read-Only

- `id` (Number) - Application ID
- `url` (String) - Application URL
- `status` (String) - Application status
- `needs_deployment` (Boolean) - Whether the application needs deployment

## Import

Applications can be imported using their ID:

```bash
terraform import ploicloud_application.main 12345
```