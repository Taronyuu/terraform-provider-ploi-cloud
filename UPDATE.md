# Updating the Ploi Cloud Terraform Provider

This guide covers how to update and maintain the Terraform provider, including adding new features, fixing bugs, and managing versions.

## Update Types

1. **Patch Updates** (1.0.0 → 1.0.1) - Bug fixes, security patches
2. **Minor Updates** (1.0.0 → 1.1.0) - New features, backward compatible
3. **Major Updates** (1.0.0 → 2.0.0) - Breaking changes, API changes

## Development Workflow

### Setting Up for Updates

```bash
# Clone repository
git clone https://github.com/ploi/terraform-provider-ploicloud.git
cd terraform-provider-ploicloud

# Create feature branch
git checkout -b feature/new-resource

# Set up development environment
make dev-setup
```

### Making Changes

```bash
# 1. Make your changes
vim internal/provider/new_resource.go

# 2. Update tests
vim tests/new_resource_test.go

# 3. Build and test locally
make build
make test

# 4. Test with local development
./test-local.sh
```

## Adding New Resources

### Step 1: Create Resource File

```bash
# Create new resource file
cat > internal/provider/backup_resource.go << 'EOF'
package provider

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/ploi/terraform-provider-ploicloud/internal/client"
)

type BackupResource struct {
    client *client.Client
}

type BackupResourceModel struct {
    ID            types.Int64  `tfsdk:"id"`
    ApplicationID types.Int64  `tfsdk:"application_id"`
    Name          types.String `tfsdk:"name"`
    Schedule      types.String `tfsdk:"schedule"`
    Status        types.String `tfsdk:"status"`
}

func NewBackupResource() resource.Resource {
    return &BackupResource{}
}

func (r *BackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *BackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a Ploi Cloud backup configuration",
        
        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                Computed:            true,
                MarkdownDescription: "Backup ID",
            },
            "application_id": schema.Int64Attribute{
                Required:            true,
                MarkdownDescription: "Application ID",
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Backup name",
            },
            "schedule": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Backup schedule (cron format)",
            },
            "status": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Backup status",
            },
        },
    }
}

// Implement CRUD methods...
EOF
```

### Step 2: Add to Provider

```bash
# Add to provider.go
vim internal/provider/provider.go

# In the Resources() method, add:
# NewBackupResource,
```

### Step 3: Add API Client Methods

```bash
# Add to client.go
vim internal/client/client.go

# Add backup-related methods
# func (c *Client) CreateBackup(backup *Backup) (*Backup, error)
# func (c *Client) GetBackup(id int64) (*Backup, error)
# etc.
```

### Step 4: Add Tests

```bash
# Create test file
cat > tests/backup_resource_test.go << 'EOF'
package tests

import (
    "fmt"
    "testing"
    
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBackupResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccBackupResourceConfig("test-backup"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("ploicloud_backup.test", "name", "test-backup"),
                    resource.TestCheckResourceAttrSet("ploicloud_backup.test", "id"),
                ),
            },
        },
    })
}
EOF
```

### Step 5: Update Documentation

```bash
# Create documentation
mkdir -p docs/resources
cat > docs/resources/backup.md << 'EOF'
---
page_title: "ploicloud_backup Resource - terraform-provider-ploicloud"
subcategory: ""
description: |-
  Manages a Ploi Cloud backup configuration.
---

# ploicloud_backup (Resource)

Manages a Ploi Cloud backup configuration.

## Example Usage

```hcl
resource "ploicloud_backup" "example" {
  application_id = ploicloud_application.example.id
  name          = "daily-backup"
  schedule      = "0 2 * * *"
}
```

## Schema

### Required

- `application_id` (Number) Application ID
- `name` (String) Backup name
- `schedule` (String) Backup schedule in cron format

### Read-Only

- `id` (Number) Backup ID
- `status` (String) Backup status
EOF
```

## Updating Existing Resources

### Schema Changes

```bash
# For backward-compatible changes (minor version)
# Add new optional attributes
"new_field": schema.StringAttribute{
    Optional:            true,
    MarkdownDescription: "New optional field",
},

# For breaking changes (major version)
# Remove or change existing attributes (carefully!)
```

### API Client Updates

```bash
# Update API models
vim internal/client/models.go

# Add new fields
type Application struct {
    // Existing fields...
    NewField string `json:"new_field,omitempty"`
}

# Update API methods to handle new fields
```

## Version Management

### Determining Version Bump

**Patch (1.0.0 → 1.0.1):**
- Bug fixes
- Security patches
- Documentation updates
- Performance improvements

**Minor (1.0.0 → 1.1.0):**
- New resources or data sources
- New optional attributes
- New provider configuration options
- Backward-compatible changes

**Major (1.0.0 → 2.0.0):**
- Removing resources or attributes
- Changing attribute types
- Changing default values
- Provider configuration changes
- Breaking API changes

### Updating Version

```bash
# Update version in main.go if needed
vim main.go

# Update CHANGELOG
vim CHANGELOG.md

# Add release notes
cat >> CHANGELOG.md << 'EOF'
## [1.1.0] - 2024-01-15

### Added
- New `ploicloud_backup` resource for managing backups
- Support for custom backup schedules

### Changed
- Improved error handling for API timeouts

### Fixed
- Fixed issue with volume resizing
EOF

# Commit changes
git add .
git commit -m "feat: add backup resource support"
```

## Testing Updates

### Comprehensive Testing

```bash
# Run all tests
make test

# Run acceptance tests
export PLOICLOUD_API_TOKEN="your-token"
make testacc

# Test specific resource
go test -v ./tests -run TestAccBackupResource

# Test backward compatibility
./test-backward-compatibility.sh
```

### Backward Compatibility Testing

```bash
#!/bin/bash
# test-backward-compatibility.sh

set -e

echo "Testing backward compatibility..."

# Test with old configuration format
mkdir -p test-compat
cd test-compat

cat > old-config.tf << 'EOF'
# Old configuration from v1.0.0
resource "ploicloud_application" "test" {
  name = "compat-test"
  type = "laravel"
}
EOF

# Test that old configs still work
terraform init
terraform validate
terraform plan

cd ..
rm -rf test-compat

echo "✅ Backward compatibility test passed"
```

## Release Process

### Preparing Release

```bash
# 1. Ensure all tests pass
make test
make testacc

# 2. Update version and changelog
vim CHANGELOG.md

# 3. Update documentation
make docs

# 4. Commit changes
git add .
git commit -m "chore: prepare v1.1.0 release"

# 5. Create and push tag
git tag v1.1.0
git push origin main
git push origin v1.1.0
```

### Automated Release

GitHub Actions will automatically:

1. Build binaries for all platforms
2. Run tests
3. Create GitHub release
4. Publish to Terraform Registry
5. Update documentation

### Manual Release (if needed)

```bash
# Build release manually
goreleaser release --clean

# Or create GitHub release
gh release create v1.1.0 \
  --title "Release v1.1.0" \
  --notes-file RELEASE_NOTES.md \
  --generate-notes
```

## Maintenance Tasks

### Regular Updates

```bash
# Update dependencies monthly
go get -u ./...
go mod tidy

# Check for security vulnerabilities
go list -json -m all | nancy sleuth

# Update GitHub Actions
# Check for newer versions of actions in .github/workflows/
```

### Security Updates

```bash
# Monitor for security advisories
# Update dependencies with security fixes immediately
# Tag emergency patch releases

# Example emergency patch
git checkout -b security/cve-fix
# Fix security issue
git commit -m "security: fix CVE-2024-xxxx"
git tag v1.0.1
git push origin v1.0.1
```

### Performance Optimization

```bash
# Profile provider performance
go build -o terraform-provider-ploicloud
./performance-test.sh

# Optimize API calls
# Add caching where appropriate
# Reduce unnecessary API requests
```

## Breaking Changes

### Planning Breaking Changes

1. **Announce early** - Give users advance notice
2. **Document migration** - Provide clear upgrade guides
3. **Support old and new** - Transition period when possible
4. **Version appropriately** - Use major version bump

### Migration Guide Example

```markdown
# Migrating from v1.x to v2.0

## Breaking Changes

### Resource: `ploicloud_application`

**Changed**: The `provider` attribute has been renamed to `cloud_provider`

#### Before (v1.x)
```hcl
resource "ploicloud_application" "example" {
  name     = "my-app"
  provider = "aws"
}
```

#### After (v2.0)
```hcl
resource "ploicloud_application" "example" {
  name           = "my-app"
  cloud_provider = "aws"
}
```

### Automated Migration

Use the provided migration script:

```bash
./scripts/migrate-v1-to-v2.sh
```
```

## Troubleshooting Updates

### Common Issues

**Issue**: Tests failing after update
```bash
# Check for breaking changes
git diff --name-only HEAD~1 HEAD

# Run specific failing tests
go test -v ./tests -run TestAccApplicationResource

# Fix test configurations
```

**Issue**: Registry publication fails
```bash
# Check GoReleaser configuration
goreleaser check

# Verify GPG signing
gpg --list-keys

# Check GitHub Actions logs
gh run list --workflow=release.yml
```

**Issue**: Backward compatibility broken
```bash
# Create compatibility shim
# Add deprecation warnings
# Provide migration path
```

### Debug Provider Issues

```bash
# Enable debug mode
export TF_LOG=DEBUG

# Test provider locally
./terraform-provider-ploicloud -debug

# Check provider schema
terraform providers schema -json
```

## Documentation Updates

### Generating Documentation

```bash
# Generate provider documentation
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

# Update examples
make examples

# Update README
vim README.md
```

### Documentation Structure

```
docs/
├── index.md              # Provider documentation
├── data-sources/
│   ├── application.md
│   └── team.md
└── resources/
    ├── application.md
    ├── service.md
    ├── domain.md
    ├── secret.md
    ├── volume.md
    ├── worker.md
    └── backup.md         # New resource
```

## Community Management

### Handling Issues

1. **Triage** - Label and prioritize issues
2. **Reproduce** - Create minimal reproduction cases  
3. **Fix** - Implement solutions
4. **Test** - Verify fixes work
5. **Release** - Include in next release

### Feature Requests

1. **Evaluate** - Assess feasibility and value
2. **Design** - Create implementation plan
3. **Implement** - Add new features
4. **Document** - Update documentation
5. **Release** - Include in minor version

## Monitoring and Analytics

### Usage Tracking

```bash
# Monitor provider downloads
curl -s https://registry.terraform.io/v1/providers/ploi/ploicloud/analytics

# Track GitHub issues and PRs
gh issue list --state=open
gh pr list --state=open
```

### Performance Monitoring

```bash
# Monitor provider performance
# Track API response times
# Monitor resource provisioning times
# Alert on errors or timeouts
```

## Update Checklist

### Pre-Update
- [ ] Identify type of update (patch/minor/major)
- [ ] Create feature branch
- [ ] Implement changes
- [ ] Add/update tests
- [ ] Update documentation

### Testing
- [ ] Unit tests pass
- [ ] Acceptance tests pass
- [ ] Backward compatibility verified
- [ ] Performance acceptable
- [ ] Security scan clean

### Release
- [ ] Version updated
- [ ] Changelog updated
- [ ] Documentation generated
- [ ] Tag created and pushed
- [ ] Release published
- [ ] Registry updated

### Post-Release
- [ ] Monitor for issues
- [ ] Update examples and tutorials
- [ ] Communicate changes to users
- [ ] Plan next iteration

## Automation Scripts

### Update Dependencies Script

```bash
#!/bin/bash
# scripts/update-deps.sh

set -e

echo "Updating Go dependencies..."
go get -u ./...
go mod tidy

echo "Running tests..."
make test

echo "Checking for vulnerabilities..."
go list -json -m all | nancy sleuth

echo "Creating PR for dependency updates..."
git checkout -b chore/update-dependencies
git add go.mod go.sum
git commit -m "chore: update dependencies"
git push origin chore/update-dependencies

echo "✅ Dependencies updated successfully"
```

This comprehensive update guide ensures maintainable, reliable updates to the Terraform provider while following best practices for version management and community engagement.