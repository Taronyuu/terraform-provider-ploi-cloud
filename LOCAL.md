# Running Ploi Cloud Terraform Provider Locally

This guide covers how to run and test the Terraform provider locally during development.

## Quick Start

```bash
# 1. Build the provider
make build

# 2. Run local test script
./test-local.sh

# 3. Set up development environment
./setup-local-dev.sh
```

## Local Development Setup

### Step 1: Build the Provider

```bash
# Build the provider binary
make build

# Verify the build
./terraform-provider-ploicloud --help
```

### Step 2: Configure Terraform for Local Development

Create a Terraform development override file:

```bash
# Create ~/.terraformrc (Linux/macOS) or terraform.rc (Windows)
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "ploi/ploicloud" = "$(pwd)"
  }
  direct {}
}
EOF
```

### Step 3: Create Test Configuration

```bash
# Create test directory
mkdir -p test-local
cd test-local

# Create basic configuration
cat > main.tf << 'EOF'
terraform {
  required_providers {
    ploicloud = {
      source = "ploi/ploicloud"
    }
  }
}

provider "ploicloud" {
  api_token = "your-api-token-here"
  # For local testing with mock server
  # api_endpoint = "http://localhost:8080/api/v1"
}

resource "ploicloud_application" "test" {
  name = "local-test-app"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
}
EOF
```

## Local Testing Methods

### Method 1: Schema Validation Testing

```bash
# Test provider schema without API calls
terraform validate

# Expected output:
# ✅ Success! The configuration is valid.
```

### Method 2: Plan Testing (No Apply)

```bash
# Generate execution plan (will show API errors but validates schema)
terraform plan

# Expected: Provider loads successfully, shows planned resources
```

### Method 3: Mock API Server Testing

Create a simple mock server for complete local testing:

```bash
# Create mock-server.go
cat > mock-server.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type Application struct {
    ID              int64     `json:"id"`
    Name            string    `json:"name"`
    Type            string    `json:"type"`
    URL             string    `json:"url"`
    Status          string    `json:"status"`
    NeedsDeployment bool      `json:"needs_deployment"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type Service struct {
    ID            int64             `json:"id"`
    ApplicationID int64             `json:"application_id"`
    Type          string            `json:"type"`
    Version       string            `json:"version"`
    Status        string            `json:"status"`
    Settings      map[string]string `json:"settings"`
}

var applications = make(map[int64]*Application)
var services = make(map[int64]*Service)
var nextAppID int64 = 1
var nextServiceID int64 = 1

func main() {
    // Application endpoints
    http.HandleFunc("/api/v1/applications", handleApplications)
    http.HandleFunc("/api/v1/applications/", handleApplication)
    
    // Service endpoints
    http.HandleFunc("/api/v1/services", handleServices)
    http.HandleFunc("/api/v1/services/", handleService)
    
    fmt.Println("🚀 Mock Ploi Cloud API server running on http://localhost:8080")
    fmt.Println("📝 Test with: api_endpoint = \"http://localhost:8080/api/v1\"")
    log.Fatal(http.ListenAndServe(":8080", enableCORS(http.DefaultServeMux)))
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func handleApplications(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    switch r.Method {
    case "GET":
        var apps []*Application
        for _, app := range applications {
            apps = append(apps, app)
        }
        json.NewEncoder(w).Encode(apps)
        
    case "POST":
        var app Application
        if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        app.ID = nextAppID
        nextAppID++
        app.URL = fmt.Sprintf("https://%s.test.ploi.it", app.Name)
        app.Status = "running"
        app.NeedsDeployment = false
        app.CreatedAt = time.Now()
        app.UpdatedAt = time.Now()
        
        applications[app.ID] = &app
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(app)
    }
}

func handleApplication(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/applications/")
    if strings.Contains(idStr, "/") {
        parts := strings.Split(idStr, "/")
        idStr = parts[0]
    }
    
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    switch r.Method {
    case "GET":
        app, exists := applications[id]
        if !exists {
            http.Error(w, "Application not found", http.StatusNotFound)
            return
        }
        json.NewEncoder(w).Encode(app)
        
    case "PUT":
        var app Application
        if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        existing, exists := applications[id]
        if !exists {
            http.Error(w, "Application not found", http.StatusNotFound)
            return
        }
        
        app.ID = id
        app.CreatedAt = existing.CreatedAt
        app.UpdatedAt = time.Now()
        app.URL = fmt.Sprintf("https://%s.test.ploi.it", app.Name)
        app.Status = "running"
        
        applications[id] = &app
        json.NewEncoder(w).Encode(app)
        
    case "DELETE":
        delete(applications, id)
        w.WriteHeader(http.StatusNoContent)
    }
}

func handleServices(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    switch r.Method {
    case "POST":
        var service Service
        if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        service.ID = nextServiceID
        nextServiceID++
        service.Status = "running"
        
        services[service.ID] = &service
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(service)
    }
}

func handleService(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/services/")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    switch r.Method {
    case "GET":
        service, exists := services[id]
        if !exists {
            http.Error(w, "Service not found", http.StatusNotFound)
            return
        }
        json.NewEncoder(w).Encode(service)
        
    case "DELETE":
        delete(services, id)
        w.WriteHeader(http.StatusNoContent)
    }
}
EOF

# Run the mock server
go run mock-server.go
```

### Method 4: Real API Testing

```bash
# Set your actual API token
export PLOICLOUD_API_TOKEN="your-actual-api-token"

# Test with real API
terraform plan -var="ploi_api_token=$PLOICLOUD_API_TOKEN"

# Apply to create real resources (be careful!)
terraform apply -var="ploi_api_token=$PLOICLOUD_API_TOKEN"
```

## Development Workflows

### Iterative Development

```bash
# 1. Make code changes
vim internal/provider/application_resource.go

# 2. Rebuild provider
make build

# 3. Test changes (no need to restart Terraform)
terraform plan

# 4. Repeat
```

### Debug Mode Testing

```bash
# Terminal 1: Run provider in debug mode
./terraform-provider-ploicloud -debug

# Copy the TF_REATTACH_PROVIDERS value from output

# Terminal 2: Use the debug provider
export TF_REATTACH_PROVIDERS='{"ploi/ploicloud":{"Protocol":"grpc","Pid":12345,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin123"}}}'
terraform plan
```

### Schema Testing

```bash
# Test all resource schemas
terraform providers schema -json | jq '.provider_schemas["ploi/ploicloud"]'

# Validate specific resource
terraform show -json | jq '.configuration.provider_config.ploicloud'
```

## Local Testing Scripts

### setup-local-dev.sh

```bash
#!/bin/bash
# setup-local-dev.sh

set -e

echo "🔧 Setting up local development environment..."

# Build provider
echo "Building provider..."
make build

# Set up dev overrides
echo "Setting up Terraform dev overrides..."
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "ploi/ploicloud" = "$(pwd)"
  }
  direct {}
}
EOF

# Create test directory
echo "Creating test directory..."
mkdir -p test-local
cd test-local

# Create test configuration
echo "Creating test configuration..."
cat > main.tf << 'EOF'
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
  description = "Ploi Cloud API token"
  type        = string
  sensitive   = true
  default     = "test-token"
}

resource "ploicloud_application" "test" {
  name = "local-test-app"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
  
  settings {
    replicas = 1
  }
}

output "app_url" {
  value = ploicloud_application.test.url
}
EOF

# Test validation
echo "Testing configuration validation..."
terraform validate

echo "✅ Local development environment ready!"
echo ""
echo "Next steps:"
echo "1. Set API token: export PLOICLOUD_API_TOKEN='your-token'"
echo "2. Run: terraform plan"
echo "3. Or start mock server: go run ../mock-server.go"
```

### test-complete-workflow.sh

```bash
#!/bin/bash
# test-complete-workflow.sh

set -e

echo "🧪 Testing complete Terraform workflow locally..."

# Clean previous test
rm -rf test-workflow
mkdir test-workflow
cd test-workflow

# Create comprehensive test
cat > main.tf << 'EOF'
terraform {
  required_providers {
    ploicloud = {
      source = "ploi/ploicloud"
    }
  }
}

provider "ploicloud" {
  api_token    = "test-token"
  api_endpoint = "http://localhost:8080/api/v1"
}

resource "ploicloud_application" "test" {
  name = "workflow-test"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
}

resource "ploicloud_service" "db" {
  application_id = ploicloud_application.test.id
  type          = "mysql"
  version       = "8.0"
  
  settings = {
    database = "test"
    size     = "5Gi"
  }
}
EOF

echo "1. Validating configuration..."
terraform validate

echo "2. Starting mock server in background..."
(cd .. && go run mock-server.go) &
MOCK_PID=$!

sleep 2

echo "3. Testing plan..."
terraform plan

echo "4. Testing apply..."
terraform apply -auto-approve

echo "5. Testing state refresh..."
terraform refresh

echo "6. Testing destroy..."
terraform destroy -auto-approve

# Cleanup
kill $MOCK_PID 2>/dev/null || true
cd ..
rm -rf test-workflow

echo "✅ Complete workflow test passed!"
```

## Debugging Local Issues

### Common Issues

**Issue**: Provider not found
```bash
# Check dev overrides
cat ~/.terraformrc

# Verify binary exists
ls -la terraform-provider-ploicloud
```

**Issue**: Schema validation errors
```bash
# Check provider logs
export TF_LOG=DEBUG
terraform validate
```

**Issue**: API connection errors
```bash
# Test API endpoint directly
curl -H "Authorization: Bearer test-token" http://localhost:8080/api/v1/applications
```

### Debug Logging

```bash
# Enable debug logging
export TF_LOG=DEBUG
export TF_LOG_PROVIDER=DEBUG
export TF_LOG_PATH=./terraform.log

terraform plan

# View logs
tail -f terraform.log
```

## Local Performance Testing

### Resource Scale Testing

```bash
# Create configuration with many resources
cat > scale-test.tf << 'EOF'
resource "ploicloud_application" "test" {
  count = 10
  
  name = "scale-test-${count.index}"
  type = "laravel"
  
  runtime {
    php_version = "8.4"
  }
}
EOF

# Test plan performance
time terraform plan
```

### Memory Usage Testing

```bash
# Monitor provider memory usage
ps aux | grep terraform-provider-ploicloud

# Or use top/htop during terraform operations
```

## Local Testing Checklist

- [ ] Provider builds successfully
- [ ] Dev overrides configured correctly
- [ ] Schema validation passes
- [ ] Mock server testing works
- [ ] Real API testing works (optional)
- [ ] Debug mode functions
- [ ] All resource types can be planned
- [ ] Error handling works correctly
- [ ] Performance is acceptable

## Next Steps

After successful local testing:

1. **Production Testing**: See [TESTING.md](TESTING.md)
2. **Deployment**: See [DEPLOY.md](DEPLOY.md)
3. **Updates**: See [UPDATE.md](UPDATE.md)

## Local Development Tips

- Use `make build` for quick rebuilds during development
- Keep the mock server running for consistent testing
- Use debug mode for troubleshooting complex issues
- Test both success and error scenarios
- Validate schema changes thoroughly before committing