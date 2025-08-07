# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2025-08-07

### Added - Enhanced Resource Capabilities

#### Application Resources (`ploicloud_application`)
- **`start_command`** - Optional string field to override default application start commands
  - Allows custom startup behavior for specialized application configurations
  - Fully optional and backward compatible

#### Service Resources (`ploicloud_service`) 
- **`storage_size`** - Optional string field for configuring service storage allocation
  - Supports Kubernetes resource notation (e.g., "10Gi", "500Mi")
  - Computed field that preserves server-assigned defaults when not specified
- **`extensions`** - Optional list of strings for PostgreSQL service extensions
  - Enables PostgreSQL extensions like uuid-ossp, pgcrypto, citext, etc.
  - Only applicable to PostgreSQL services, ignored for other types

#### Worker Resources (`ploicloud_worker`)
- **`type`** - Optional string field for worker type specification
  - Defaults to "queue" for backward compatibility
  - Computed field that preserves server-assigned values
- **`memory_request`** - Optional string field for worker memory allocation
  - Supports Kubernetes resource notation (e.g., "512Mi", "1Gi")
  - Computed field for server-managed defaults
- **`cpu_request`** - Optional string field for worker CPU allocation  
  - Supports Kubernetes resource notation (e.g., "250m", "1")
  - Computed field for server-managed defaults

#### Volume Resources (`ploicloud_volume`)
- **`storage_class`** - Optional string field for storage class specification
  - Allows selection of storage performance tiers (e.g., "fast-ssd", "standard")
  - Computed field that preserves server defaults when not specified

### Enhanced API Models
- Updated all API client models to support new fields
- Maintained full backward compatibility with existing configurations
- Added proper JSON marshaling/unmarshaling for all new fields

### Developer Experience
- All new fields are optional and computed to ensure zero breaking changes
- Comprehensive example configurations demonstrating new capabilities
- Updated documentation with feature descriptions and usage examples

### Technical Details
- All resource schema definitions updated with appropriate validators
- Proper `toAPIModel` and `fromAPIModel` method implementations
- Complete test coverage for enhanced functionality
- Follows existing code patterns and conventions throughout

---

## [1.0.0] - Initial Release

### Added
- Initial Terraform provider for Ploi Cloud platform
- Support for applications, services, workers, volumes, secrets, and domains
- Complete resource lifecycle management (create, read, update, delete)
- Import capability for existing resources
- Comprehensive documentation and examples