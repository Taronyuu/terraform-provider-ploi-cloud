# Quick Start Guide

Get up and running with the Ploi Cloud Terraform Provider in 5 minutes.

## 1. Build the Provider

```bash
# Clone and build
git clone https://github.com/ploi/terraform-provider-ploicloud.git
cd terraform-provider-ploicloud
make build
```

## 2. Set Up Local Development

```bash
# Configure Terraform to use local provider
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "ploi/ploicloud" = "$(pwd)"
  }
  direct {}
}
EOF
```

## 3. Test Configuration

```bash
# Run the quick test
./test-local.sh

# Expected output: ✅ All tests pass
```

## 4. Create Your First Resource

```hcl
# main.tf
terraform {
  required_providers {
    ploicloud = {
      source = "ploi/ploicloud"
    }
  }
}

provider "ploicloud" {
  api_token = var.ploi_api_token
}

variable "ploi_api_token" {
  description = "Your Ploi Cloud API token"
  type        = string
  sensitive   = true
}

resource "ploicloud_application" "example" {
  name = "my-terraform-app"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
  
  settings {
    replicas = 2
  }
}

output "app_url" {
  value = ploicloud_application.example.url
}
```

## 5. Deploy

```bash
# Set your API token
export PLOICLOUD_API_TOKEN="your-api-token-here"

# Plan and apply
terraform plan -var="ploi_api_token=$PLOICLOUD_API_TOKEN"
terraform apply -var="ploi_api_token=$PLOICLOUD_API_TOKEN"
```

## 🎉 You're Done!

Your application is now managed by Terraform. 

### Next Steps:
- **Add Services**: [examples/complete/](examples/complete/) for full examples
- **Production**: See [DEPLOY.md](DEPLOY.md) for publishing
- **Updates**: See [UPDATE.md](UPDATE.md) for maintenance

### Need Help?
- **Building**: [BUILD.md](BUILD.md)
- **Local Testing**: [LOCAL.md](LOCAL.md) 
- **Full Testing**: [TESTING.md](TESTING.md)

### Resources Available:
- `ploicloud_application` - Laravel, WordPress, Node.js apps
- `ploicloud_service` - Databases, caches, queues
- `ploicloud_domain` - Custom domains with SSL
- `ploicloud_secret` - Environment variables
- `ploicloud_volume` - Persistent storage
- `ploicloud_worker` - Background job workers