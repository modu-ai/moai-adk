# MoAI Menu Project Configuration Management

Unified configuration management system for MoAI menu project operations, integrating 5 specialized modules into a single cohesive system.

## Overview

This skill replaces 5 separate configuration management skills with a unified system that provides:

- Centralized Configuration: Single source of truth for all menu project settings
- Legacy Migration: Automatic detection and migration from existing configurations
- Module Specialization: Dedicated managers for each functional area
- Validation Framework: JSON Schema validation and business rule checking
- Performance Optimization: Caching and efficient file handling

## Quick Start

```python
from moai_menu_project import create_config_manager, ConfigurationMigrator

# Initialize configuration manager
config_manager = create_config_manager("/path/to/config")

# Migrate legacy configurations
migrator = ConfigurationMigrator(config_manager)
result = migrator.detect_and_migrate(backup=True)

# Access module-specific configuration
from moai_menu_project import BatchQuestionsConfigManager
batch_manager = BatchQuestionsConfigManager(config_manager)
settings = batch_manager.get_config()
```

## Architecture

### Core Components

1. UnifiedConfigManager: Centralized configuration with JSON Schema validation
2. ConfigurationMigrator: Legacy configuration detection and migration
3. Module Managers: Specialized access for each functional area
4. Validation Framework: Schema and business rule validation
5. Backup/Recovery: Automatic backup creation and rollback

### Supported Modules

- Batch Questions: Batch processing configuration
- Documentation: Documentation generation settings 
- Language Config: Internationalization and localization
- Template Optimizer: Template processing optimization
- Project Initializer: Project setup and initialization

## Migration Guide

### From Legacy Configurations

The system automatically detects and migrates legacy configuration files:

1. Detection: Scans for legacy configuration patterns
2. Transformation: Converts to unified format with version-specific rules
3. Validation: Ensures migrated data meets schema requirements
4. Backup: Creates backup before applying changes
5. History: Records migration for audit and rollback

### Example Migration

```python
# Preview migration changes
preview = migrator.preview_migration()
print(f"Will migrate {preview['migrated']} files")

# Execute migration
result = migrator.detect_and_migrate(backup=True)
print(f"Successfully migrated {result['migrated']} configurations")
```

## Configuration Schema

The unified configuration follows this structure:

```json
{
 "version": "1.0.0",
 "metadata": {
 "created_at": "timestamp",
 "updated_at": "timestamp",
 "migration_history": [...]
 },
 "project_settings": {...},
 "batch_questions": {...},
 "documentation": {...},
 "language_config": {...},
 "template_optimizer": {...},
 "project_initializer": {...}
}
```

See `schemas/config-schema.json` for complete schema definition.

## File Structure

```
moai-menu-project/
 SKILL.md # Main skill documentation
 references/overview.md # Skill overview
 __init__.py # Package initialization
 modules/
 config_manager.py # Core configuration management
 migration_manager.py # Legacy migration system
 schemas/
 config-schema.json # JSON Schema for validation
 examples/
 config-migration-example.json # Migration example
```

## Integration

### With MoAI Ecosystem

- moai-cc-configuration: Claude Code configuration patterns
- moai-core-workflow: Workflow-based configuration management
- moai-quality-security: Security validation and compliance

### External Systems

- Environment variable overrides
- CLI argument integration
- API endpoint for dynamic updates
- Monitoring and observability integration

## Validation

The system provides comprehensive validation:

- JSON Schema: Structural validation with type checking
- Business Rules: Custom validation logic for consistency
- Security: Compliance checking and constraint enforcement
- Performance: Resource limit and efficiency validation

## Performance

Optimizations include:

- Caching: In-memory caching with file modification monitoring
- Lazy Loading: Load only required configuration sections
- Concurrent Access: Thread-safe operations with locking
- Resource Management: Efficient file handling and memory usage

## Security

Features for secure configuration management:

- Access Control: Role-based permissions for configuration access
- Encryption: Sensitive field encryption and secure key management
- Audit Logging: Comprehensive change tracking and audit trails
- Data Protection: Secure backup storage and transmission

## Troubleshooting

Common issues and solutions:

### Schema Validation Errors
- Check JSON structure against schema
- Verify required fields are present
- Ensure data types match schema definitions

### Migration Issues
- Review legacy file detection patterns
- Check transformation rules for specific versions
- Verify file permissions and access rights

### Performance Problems
- Enable caching for frequently accessed configurations
- Use lazy loading for large configuration files
- Monitor resource usage and optimize file access patterns

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review migration examples and schema documentation
3. Examine error messages for specific guidance
4. Validate configuration against schema

## Version History

- 1.0.0: Initial release with unified configuration management
- Automatic migration from 5 legacy skill configurations
- JSON Schema validation and business rule checking
- Performance optimization with caching and lazy loading
- Security features with access control and audit logging
