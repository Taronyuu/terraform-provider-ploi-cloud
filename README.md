# Terraform Provider for Ploi Cloud

Manage your Ploi Cloud infrastructure using Terraform's declarative configuration language.

## 🚀 Quick Start

```bash
# 1. Build provider
make build

# 2. Test locally  
./test-local.sh

# 3. Start using Terraform
terraform apply
```

**[👉 Full Quick Start Guide](QUICK_START.md)**

## 📖 Documentation

| Guide | Description |
|-------|-------------|
| **[🔨 BUILD.md](BUILD.md)** | How to build from source |
| **[💻 LOCAL.md](LOCAL.md)** | Local development and testing |
| **[🚀 DEPLOY.md](DEPLOY.md)** | Production deployment |
| **[🔄 UPDATE.md](UPDATE.md)** | Updating and maintenance |
| **[🧪 TESTING.md](TESTING.md)** | Comprehensive testing guide |

## ✨ Features

- **🏗️ Complete Infrastructure as Code** - Manage applications, services, domains, and more
- **🔧 All Application Types** - Laravel, WordPress, Node.js, Statamic, Craft CMS
- **🗄️ Database & Cache Services** - MySQL, PostgreSQL, Redis, MongoDB, RabbitMQ
- **📦 Background Workers** - Queue workers with auto-scaling
- **💾 Persistent Storage** - Volume management with resizing
- **🌐 Custom Domains** - SSL certificates included
- **🔐 Secret Management** - Secure environment variables
- **📊 Import Existing Resources** - Import your current setup

## 🏃‍♂️ Example Usage

```hcl
terraform {
  required_providers {
    ploicloud = {
      source  = "ploi/ploicloud"
      version = "~> 1.0"
    }
  }
}

provider "ploicloud" {
  api_token = var.ploi_api_token
}

# Laravel Application
resource "ploicloud_application" "api" {
  name = "production-api"
  type = "laravel"
  
  runtime {
    php_version    = "8.4"
    nodejs_version = "22"
  }
  
  start_command = "php artisan serve --host=0.0.0.0 --port=8000"
  
  additional_domains = [
    "api.example.com",
    "www.api.example.com"
  ]
  
  settings {
    replicas          = 3
    scheduler_enabled = true
  }
}

# MySQL Database
resource "ploicloud_service" "db" {
  application_id = ploicloud_application.api.id
  type          = "mysql"
  version       = "8.0"
  storage_size  = "20Gi"
  memory_request = "2Gi"
  settings = {
    database = "production"
  }
}

# Queue Worker (as service)
resource "ploicloud_service" "queue" {
  application_id = ploicloud_application.api.id
  service_name   = "default-queue"
  type          = "worker"
  replicas      = 2
  memory_request = "1Gi"
  settings = {
    command = "php artisan queue:work"
  }
}
```

## 📚 Resources

| Resource | Description | Status |
|----------|-------------|--------|
| `ploicloud_application` | Applications (Laravel, WordPress, etc.) | ✨ Enhanced logging and error handling |
| `ploicloud_service` | Databases, caches, message queues, **workers** | ✨ Worker support, enhanced validation |
| `ploicloud_domain` | Custom domains with SSL | ✨ Enhanced error messages |
| `ploicloud_secret` | Environment variables | ✨ Enhanced validation |
| `ploicloud_volume` | Persistent storage | ⚠️ **Read-only** - Use services with `storage_size` |
| `ploicloud_worker` | Background job workers | ⚠️ **Deprecated** - Use services with `type = "worker"` |

### 🆕 New in v1.2.0 - API Error Fixes & Enhanced Reliability

**🔧 Enhanced Error Handling & Logging:**
- Comprehensive request/response logging with debug support (`TF_LOG=DEBUG`, `PLOI_DEBUG=1`)
- Detailed 422 validation error parsing with field-specific suggestions
- Automatic retry logic for transient API failures (5xx errors)
- Sanitized logging to protect sensitive data (API tokens)

**🔄 Resource Strategy Updates:**
- **Worker Resources**: Deprecated in favor of services with `type = "worker"`
- **Volume Resources**: Read-only mode - volumes auto-created with storage services
- **Service Resources**: Enhanced with worker support and comprehensive validation

**📝 Migration Support:**
- Clear migration guidance for worker → service transitions
- Backward compatibility maintained for existing deployments
- Helpful error messages with actionable solutions

**✅ Previous Features (v1.1.0):**
- `start_command` - Custom application start commands
- `additional_domains` - Multiple custom domains per application
- `storage_size` - Storage allocation for databases and caches
- `memory_request` - Memory allocation for services
- `extensions` - PostgreSQL extensions support
- Enhanced worker resource specifications

## 🧑‍💻 Development

### Prerequisites
- Go 1.23+
- Terraform 1.5+
- Ploi Cloud API token

### Build & Test
```bash
# Build provider
make build

# Run tests
make test

# Test locally
./test-local.sh
```

### 🐞 Debugging & Troubleshooting

**Enable Enhanced Logging:**
```bash
# Enable comprehensive request/response logging
export TF_LOG=DEBUG

# Or use Ploi-specific debug logging
export PLOI_DEBUG=1

# Run your Terraform commands
terraform plan
terraform apply
```

**Common Issues & Solutions:**

| Error | Solution |
|-------|----------|
| **422 Unprocessable Entity** | Check service configuration - missing required fields or invalid resource specifications |
| **Worker resource not found** | Workers are now created as services with `type = "worker"` - see migration guide below |
| **Volume POST not supported** | Volumes are auto-created with storage services - use service resources with `storage_size` |

**Migration Guide:**

*Worker Resources → Service Resources:*
```hcl
# ❌ OLD (deprecated)
resource "ploicloud_worker" "queue" {
  application_id = ploicloud_application.app.id
  name          = "queue-worker"
  command       = "php artisan queue:work"
  replicas      = 2
}

# ✅ NEW (recommended)
resource "ploicloud_service" "queue" {
  application_id = ploicloud_application.app.id
  service_name   = "queue-worker"
  type          = "worker"
  replicas      = 2
  memory_request = "1Gi"
  settings = {
    command = "php artisan queue:work"
  }
}
```

*Volume Creation → Service Storage:*
```hcl
# ❌ OLD (not supported)
resource "ploicloud_volume" "uploads" {
  application_id = ploicloud_application.app.id
  size          = "10Gi"
}

# ✅ NEW (creates volume automatically)
resource "ploicloud_service" "database" {
  application_id = ploicloud_application.app.id
  type          = "mysql"
  storage_size   = "10Gi"  # Volume created automatically
  memory_request = "2Gi"   # Memory allocation
}
```

**Troubleshooting Steps:**
1. Enable debug logging with `export TF_LOG=DEBUG`
2. Check API response details in logs
3. Verify resource configurations match API requirements
4. For 422 errors, review field-specific validation messages
5. Check [API documentation](https://docs.ploi.io/cloud) for resource specifications

### Local Development Setup
```bash
# Set up dev overrides
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "ploi/ploicloud" = "$(pwd)"
  }
  direct {}
}
EOF
```

**[📖 Complete Development Guide](LOCAL.md)**

## 🚀 Production Deployment

### Terraform Registry (Recommended)
```bash
# Tag and release
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions automatically:
# ✅ Builds all platforms
# ✅ Signs binaries  
# ✅ Publishes to registry
```

### Direct Installation
```bash
# Download and install
curl -sSL https://raw.githubusercontent.com/ploi/terraform-provider-ploicloud/main/install.sh | bash
```

**[🚀 Complete Deployment Guide](DEPLOY.md)**

## 📈 Examples

| Example | Description |
|---------|-------------|
| **[Basic](examples/basic/)** | Simple Laravel application |
| **[Complete](examples/complete/)** | Full stack with services |
| **[WordPress](examples/wordpress/)** | WordPress with custom domain |
| **[Import](examples/import/)** | Import existing resources |

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/ploi/terraform-provider-ploicloud/issues)
- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/ploi/ploicloud)
- **Community**: [Ploi Discord](https://discord.gg/ploi)

---

**Built with ❤️ by the Ploi team**