# Ploi Cloud Provider

The Ploi Cloud provider allows you to manage your Ploi Cloud applications and services using Terraform.

## Example Usage

```terraform
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
```

## Authentication

The Ploi Cloud provider requires an API token for authentication. You can generate an API token from your Ploi Cloud dashboard.

### Environment Variables

You can provide your API token via the `PLOICLOUD_API_TOKEN` environment variable:

```bash
export PLOICLOUD_API_TOKEN="your-api-token"
```

## Schema

### Required

- `api_token` (String, Sensitive) - The API token for Ploi Cloud authentication.

### Optional

- `api_endpoint` (String) - The API endpoint for Ploi Cloud. Defaults to `https://cloud.ploi.io/api/v1`.