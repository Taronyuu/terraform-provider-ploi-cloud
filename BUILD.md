# Building the Ploi Cloud Terraform Provider

This guide covers how to build the Terraform provider from source.

## Prerequisites

- **Go 1.21+** installed
- **Terraform 1.5+** installed
- **Git** for version control
- **Make** for build automation (optional)

## Quick Start

```bash
# Clone the repository (if not already done)
git clone https://github.com/ploi/terraform-provider-ploicloud.git
cd terraform-provider-ploicloud

# Build the provider
make build

# Or build manually
go build -v -o terraform-provider-ploicloud
```

## Build Commands

### Using Makefile (Recommended)

```bash
# Build the provider
make build

# Install locally for testing
make install

# Run tests
make test

# Run acceptance tests (requires API token)
make testacc

# Lint code
make lint

# Clean build artifacts
make clean
```

### Manual Build Commands

```bash
# Basic build
go build -v -o terraform-provider-ploicloud

# Build with version info
go build -v -ldflags "-X main.version=1.0.0" -o terraform-provider-ploicloud

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -v -o terraform-provider-ploicloud

# Build for multiple platforms
GOOS=windows GOARCH=amd64 go build -v -o terraform-provider-ploicloud.exe
GOOS=darwin GOARCH=amd64 go build -v -o terraform-provider-ploicloud-darwin
GOOS=darwin GOARCH=arm64 go build -v -o terraform-provider-ploicloud-darwin-arm64
```

## Build Targets

### Development Build
```bash
# Fast build for development
go build -v
```

### Release Build
```bash
# Optimized build with version info
go build -v -ldflags "-s -w -X main.version=1.0.0" -o terraform-provider-ploicloud
```

### Cross-Platform Build
```bash
# Build for all supported platforms
for GOOS in darwin linux windows; do
  for GOARCH in amd64 arm64; do
    if [[ "$GOOS" == "windows" && "$GOARCH" == "arm64" ]]; then
      continue
    fi
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -v -ldflags "-s -w -X main.version=1.0.0" -o "dist/terraform-provider-ploicloud_${GOOS}_${GOARCH}"
  done
done
```

## Build Verification

### Test the Build
```bash
# Check if binary was created
ls -la terraform-provider-ploicloud

# Test provider can start
./terraform-provider-ploicloud --help

# Should output:
# Usage of ./terraform-provider-ploicloud:
#   -debug
#     	set to true to run the provider with support for debuggers
```

### Verify Dependencies
```bash
# Check Go module dependencies
go mod verify

# Update dependencies if needed
go mod tidy

# Check for vulnerabilities
go list -json -m all | go run golang.org/x/vuln/cmd/govulncheck@latest -mode=json
```

## Build Environment Setup

### Go Environment
```bash
# Check Go version
go version
# Should be 1.21 or higher

# Set Go environment variables
export GO111MODULE=on
export CGO_ENABLED=0

# Optional: Set GOPROXY for faster builds
export GOPROXY=https://proxy.golang.org,direct
```

### Development Dependencies
```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goreleaser/goreleaser@latest

# Install Terraform plugin development tools
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
```

## Build Configuration

### go.mod Configuration
The `go.mod` file defines the module and its dependencies:

```go
module github.com/ploi/terraform-provider-ploicloud

go 1.23.0

require (
    github.com/hashicorp/terraform-plugin-framework v1.14.1
    github.com/hashicorp/terraform-plugin-framework-validators v0.18.0
    github.com/hashicorp/terraform-plugin-go v0.26.0
    github.com/hashicorp/terraform-plugin-testing v1.5.1
)
```

### Build Tags
```bash
# Build with specific tags
go build -tags=integration -v

# Build without tests
go build -tags=!test -v
```

## Build Optimization

### Reduce Binary Size
```bash
# Strip debug info and symbol table
go build -ldflags "-s -w" -v

# Use UPX compression (if available)
upx --best terraform-provider-ploicloud
```

### Build Cache
```bash
# Clean build cache
go clean -cache

# Show build cache location
go env GOCACHE

# Build with cache disabled
GOCACHE=off go build -v
```

## Troubleshooting Build Issues

### Common Build Errors

**Error**: `package not found`
```bash
# Solution: Clean and rebuild modules
go clean -modcache
go mod download
go mod tidy
```

**Error**: `version conflict`
```bash
# Solution: Update dependencies
go get -u ./...
go mod tidy
```

**Error**: `CGO_ENABLED required`
```bash
# Solution: Disable CGO
CGO_ENABLED=0 go build -v
```

### Debug Build Issues
```bash
# Verbose build output
go build -v -x

# Show all dependencies
go list -m all

# Check module integrity
go mod verify
```

## Build Performance

### Parallel Builds
```bash
# Use all CPU cores for parallel builds
GOMAXPROCS=$(nproc) go build -v

# Limit parallel builds
GOMAXPROCS=4 go build -v
```

### Build Time Optimization
```bash
# Use build cache effectively
go build -v  # First build
go build -v  # Subsequent builds use cache

# Skip tests during build
go build -v -tags=!test
```

## Continuous Integration

### GitHub Actions Build
```yaml
# .github/workflows/build.yml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
          
      - name: Build
        run: make build
        
      - name: Test Build
        run: ./terraform-provider-ploicloud --help
```

### Local CI Script
```bash
#!/bin/bash
# ci-build.sh

set -e

echo "Starting CI build..."

# Clean previous builds
make clean

# Verify dependencies
go mod verify

# Run linting
make lint

# Build provider
make build

# Verify build
./terraform-provider-ploicloud --help

# Run tests
make test

echo "Build completed successfully!"
```

## Build Artifacts

After a successful build, you should have:

```
terraform-provider-ploicloud         # Main binary
terraform-provider-ploicloud.exe     # Windows binary (if built)
dist/                                # Cross-platform builds
├── terraform-provider-ploicloud_darwin_amd64
├── terraform-provider-ploicloud_darwin_arm64
├── terraform-provider-ploicloud_linux_amd64
├── terraform-provider-ploicloud_linux_arm64
└── terraform-provider-ploicloud_windows_amd64.exe
```

## Next Steps

After building successfully:

1. **Local Testing**: See [LOCAL.md](LOCAL.md) for running locally
2. **Deployment**: See [DEPLOY.md](DEPLOY.md) for production deployment
3. **Updates**: See [UPDATE.md](UPDATE.md) for updating the provider

## Build Checklist

- [ ] Go 1.21+ installed and working
- [ ] Dependencies downloaded (`go mod download`)
- [ ] Build completes without errors
- [ ] Binary can start and show help
- [ ] All tests pass (`make test`)
- [ ] Cross-platform builds work (if needed)
- [ ] Build artifacts are correct size and executable