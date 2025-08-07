# Deploying Ploi Cloud Terraform Provider to Production

This guide covers how to deploy the Terraform provider to production environments and make it available to users.

## Deployment Overview

The Terraform provider can be deployed through several methods:

1. **Terraform Registry** (Recommended - Public)
2. **Private Registry** (Enterprise/Internal)
3. **Direct Binary Distribution**
4. **GitHub Releases**

## Method 1: Terraform Registry (Recommended)

### Prerequisites

- GitHub account with repository
- Terraform Registry account (free)
- GPG key for signing releases
- GoReleaser configuration

### Step 1: Prepare Repository

```bash
# Ensure repository is properly structured
terraform-provider-ploicloud/
├── main.go
├── go.mod
├── go.sum
├── internal/
├── examples/
├── docs/
├── .goreleaser.yml
└── .github/workflows/release.yml
```

### Step 2: Set up GPG Signing

```bash
# Generate GPG key (if you don't have one)
gpg --batch --full-generate-key <<EOF
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: Ploi Provider Bot
Name-Email: provider@ploi.io
Expire-Date: 0
%no-protection
%commit
EOF

# Export GPG key
gpg --armor --export-secret-keys provider@ploi.io > private-key.asc
gpg --armor --export provider@ploi.io > public-key.asc

# Get GPG fingerprint
gpg --with-colons --fingerprint provider@ploi.io | awk -F: '$1 == "fpr" {print $10; exit}'
```

### Step 3: Configure GitHub Repository

Add the following secrets to your GitHub repository:

```bash
# In GitHub repository settings > Secrets and variables > Actions
GPG_PRIVATE_KEY=<content of private-key.asc>
PASSPHRASE=<your GPG passphrase or leave empty if none>
```

### Step 4: Create Release Workflow

The `.github/workflows/release.yml` is already configured. It will:

1. Trigger on git tags (v1.0.0, v1.1.0, etc.)
2. Build binaries for all platforms
3. Sign releases with GPG
4. Create GitHub release
5. Automatically publish to Terraform Registry

### Step 5: Create and Push Release

```bash
# Tag and push release
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions will automatically:
# - Build all platform binaries
# - Sign the release
# - Create GitHub release
# - Publish to Terraform Registry
```

### Step 6: Register on Terraform Registry

1. Go to [registry.terraform.io](https://registry.terraform.io)
2. Sign in with GitHub
3. Click "Publish" > "Provider"
4. Select your GitHub repository
5. Verify GPG key
6. Publish the provider

After publishing, users can use:

```hcl
terraform {
  required_providers {
    ploicloud = {
      source  = "ploi/ploicloud"
      version = "~> 1.0"
    }
  }
}
```

## Method 2: Private Registry

### HashiCorp Terraform Cloud/Enterprise

```bash
# 1. Build and package provider
make build

# 2. Create provider package
tar -czf terraform-provider-ploicloud_1.0.0_linux_amd64.tar.gz terraform-provider-ploicloud

# 3. Upload to Terraform Cloud private registry
# via UI or API
```

### Private Registry Setup

```hcl
# terraform.tf
terraform {
  required_providers {
    ploicloud = {
      source  = "company.com/internal/ploicloud"
      version = "~> 1.0"
    }
  }
}
```

### Artifactory/Nexus Setup

```bash
# Build for multiple platforms
make build-all

# Upload to artifact repository
curl -u admin:password -T terraform-provider-ploicloud_1.0.0_linux_amd64.tar.gz \
  "https://artifactory.company.com/artifactory/terraform-providers/ploi/ploicloud/1.0.0/"
```

## Method 3: Direct Binary Distribution

### Build Release Binaries

```bash
# Build for all platforms
mkdir -p dist

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o dist/terraform-provider-ploicloud_1.0.0_linux_amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=1.0.0" -o dist/terraform-provider-ploicloud_1.0.0_linux_arm64

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o dist/terraform-provider-ploicloud_1.0.0_darwin_amd64

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=1.0.0" -o dist/terraform-provider-ploicloud_1.0.0_darwin_arm64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o dist/terraform-provider-ploicloud_1.0.0_windows_amd64.exe
```

### Create Distribution Package

```bash
# Create distribution script
cat > install.sh << 'EOF'
#!/bin/bash

set -e

VERSION="1.0.0"
PROVIDER="ploicloud"
NAMESPACE="ploi"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY_NAME="terraform-provider-${PROVIDER}_${VERSION}_${OS}_${ARCH}"
if [[ "$OS" == "windows" ]]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

# Download binary
echo "Downloading ${BINARY_NAME}..."
curl -LO "https://github.com/ploi/terraform-provider-ploicloud/releases/download/v${VERSION}/${BINARY_NAME}"

# Install binary
INSTALL_DIR="${HOME}/.terraform.d/plugins/${NAMESPACE}/${PROVIDER}/${VERSION}/${OS}_${ARCH}"
mkdir -p "$INSTALL_DIR"

mv "$BINARY_NAME" "$INSTALL_DIR/terraform-provider-${PROVIDER}"
chmod +x "$INSTALL_DIR/terraform-provider-${PROVIDER}"

echo "✅ Terraform provider ${NAMESPACE}/${PROVIDER} v${VERSION} installed successfully!"
echo ""
echo "Usage in your Terraform configuration:"
echo ""
echo "terraform {"
echo "  required_providers {"
echo "    ${PROVIDER} = {"
echo "      source  = \"${NAMESPACE}/${PROVIDER}\""
echo "      version = \"~> ${VERSION}\""
echo "    }"
echo "  }"
echo "}"
EOF

chmod +x install.sh
```

### User Installation

Users can install with:

```bash
# Download and run install script
curl -sSL https://raw.githubusercontent.com/ploi/terraform-provider-ploicloud/main/install.sh | bash

# Or manual installation
curl -LO https://github.com/ploi/terraform-provider-ploicloud/releases/download/v1.0.0/terraform-provider-ploicloud_1.0.0_linux_amd64
mkdir -p ~/.terraform.d/plugins/ploi/ploicloud/1.0.0/linux_amd64
mv terraform-provider-ploicloud_1.0.0_linux_amd64 ~/.terraform.d/plugins/ploi/ploicloud/1.0.0/linux_amd64/terraform-provider-ploicloud
chmod +x ~/.terraform.d/plugins/ploi/ploicloud/1.0.0/linux_amd64/terraform-provider-ploicloud
```

## Method 4: GitHub Releases Only

### Automated Release with GitHub Actions

The release workflow will create GitHub releases automatically when you push tags:

```bash
# Create and push tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions creates release with:
# - Compiled binaries for all platforms
# - SHA256 checksums
# - GPG signatures
# - Release notes
```

### Manual Release

```bash
# Build all binaries
make build-all

# Create release
gh release create v1.0.0 \
  --title "Release v1.0.0" \
  --notes "Initial release of Ploi Cloud Terraform Provider" \
  dist/*
```

## Deployment Verification

### Test Provider Installation

```bash
# Create test configuration
mkdir test-deployment
cd test-deployment

cat > main.tf << 'EOF'
terraform {
  required_providers {
    ploicloud = {
      source  = "ploi/ploicloud"
      version = "~> 1.0"
    }
  }
}

provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "deployment-test"
  type = "laravel"
}
EOF

# Test provider can be downloaded and loaded
terraform init
terraform validate
```

### Verify Registry Publication

```bash
# Check provider is available on registry
curl -s https://registry.terraform.io/v1/providers/ploi/ploicloud | jq .

# Test provider download
terraform providers lock -platform=linux_amd64 -platform=darwin_amd64 -platform=windows_amd64
```

## Deployment Environments

### Staging Deployment

```bash
# Deploy to staging registry first
terraform {
  required_providers {
    ploicloud = {
      source  = "staging.company.com/ploi/ploicloud"
      version = "~> 1.0.0-beta"
    }
  }
}
```

### Production Deployment

```bash
# Production release process
git checkout main
git tag v1.0.0
git push origin v1.0.0

# Monitor deployment
gh run watch

# Verify deployment
terraform providers schema -json | jq '.provider_schemas["ploi/ploicloud"]'
```

## Monitoring and Alerting

### Release Monitoring

```bash
# Monitor GitHub release status
gh api repos/ploi/terraform-provider-ploicloud/releases/latest

# Check registry status
curl -s https://registry.terraform.io/v1/providers/ploi/ploicloud/versions
```

### Usage Analytics

```bash
# Track provider downloads (if available)
curl -s https://registry.terraform.io/v1/providers/ploi/ploicloud/analytics
```

## Rollback Procedures

### Registry Rollback

If a release has critical issues:

```bash
# 1. Remove problematic version from registry (if possible)
# 2. Update version constraints in documentation
# 3. Notify users to pin to previous version

terraform {
  required_providers {
    ploicloud = {
      source  = "ploi/ploicloud"
      version = "= 0.9.0"  # Pin to known good version
    }
  }
}
```

### Emergency Hotfix

```bash
# Create hotfix branch
git checkout -b hotfix/v1.0.1
# Fix critical issue
git commit -m "Fix critical issue"
git tag v1.0.1
git push origin v1.0.1
```

## Security Considerations

### Signing and Verification

- All releases are GPG signed
- Users can verify signatures
- Checksums provided for integrity

```bash
# Verify release signature
gpg --verify terraform-provider-ploicloud_1.0.0_SHA256SUMS.sig terraform-provider-ploicloud_1.0.0_SHA256SUMS
```

### Supply Chain Security

- Builds run in isolated GitHub Actions environments
- Dependencies are pinned and verified
- No external dependencies during build

## CI/CD Pipeline

### Deployment Pipeline Stages

1. **Code Commit** → Triggers CI
2. **Tests** → Unit and integration tests
3. **Build** → Cross-platform binaries
4. **Security Scan** → Vulnerability checks
5. **Stage Deploy** → Deploy to staging
6. **Production Deploy** → Deploy to production
7. **Verification** → Post-deployment checks

### Pipeline Configuration

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    tags: ['v*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}
      
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

## Deployment Checklist

### Pre-Deployment
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Version bumped
- [ ] GPG key configured
- [ ] Release notes prepared

### Deployment
- [ ] Tag created and pushed
- [ ] GitHub Actions completed successfully
- [ ] Binaries built for all platforms
- [ ] Release created with assets
- [ ] Provider published to registry

### Post-Deployment
- [ ] Provider available on registry
- [ ] Test installation works
- [ ] Documentation updated
- [ ] Users notified of new release
- [ ] Monitor for issues

## Troubleshooting Deployment Issues

### Common Issues

**Issue**: GPG signing fails
```bash
# Check GPG key
gpg --list-keys
# Re-import private key
gpg --import private-key.asc
```

**Issue**: Registry publication fails
```bash
# Check provider manifest
# Verify GPG signature
# Contact Terraform Registry support
```

**Issue**: Binary compatibility issues
```bash
# Test on multiple platforms
# Check architecture support
# Verify build flags
```

## Next Steps

After successful deployment:

1. **User Documentation**: Update README and examples
2. **Community Engagement**: Announce on forums, social media
3. **Monitoring**: Set up alerting for issues
4. **Maintenance**: Plan regular updates and security patches

## Support and Maintenance

- Monitor GitHub issues for user feedback
- Respond to security vulnerabilities promptly
- Regular updates for Terraform compatibility
- Maintain backward compatibility when possible