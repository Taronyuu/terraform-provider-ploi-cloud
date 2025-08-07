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
  settings = {
    database = "production"
    size     = "20Gi"
  }
}

# Queue Worker
resource "ploicloud_worker" "queue" {
  application_id = ploicloud_application.api.id
  name          = "default-queue"
  command       = "php artisan queue:work"
  replicas      = 2
}
```

## 📚 Resources

| Resource | Description | Recent Enhancements |
|----------|-------------|-------------------|
| `ploicloud_application` | Applications (Laravel, WordPress, etc.) | ✨ `start_command` - Custom start commands |
| `ploicloud_service` | Databases, caches, message queues | ✨ `storage_size`, `extensions` (PostgreSQL) |
| `ploicloud_domain` | Custom domains with SSL | |
| `ploicloud_secret` | Environment variables | |
| `ploicloud_volume` | Persistent storage | ✨ `storage_class` - Storage class specification |
| `ploicloud_worker` | Background job workers | ✨ `type`, `memory_request`, `cpu_request` |

### 🆕 New in v1.1.0

**Enhanced Application Resources:**
- `start_command` - Override default application start commands for custom setups

**Enhanced Service Resources:**
- `storage_size` - Configure storage allocation for databases and caches (e.g., "10Gi", "500Mi")
- `extensions` - PostgreSQL extensions support (uuid-ossp, pgcrypto, citext, etc.)

**Enhanced Worker Resources:**
- `type` - Worker type specification (queue, scheduler, custom)
- `memory_request` - Memory allocation for workers (e.g., "512Mi", "1Gi")
- `cpu_request` - CPU allocation for workers (e.g., "250m", "1")

**Enhanced Volume Resources:**
- `storage_class` - Storage class for volumes (fast-ssd, standard, etc.)

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