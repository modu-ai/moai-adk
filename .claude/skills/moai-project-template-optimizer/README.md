# MoAI Project Template Optimizer

A comprehensive skill for intelligent template comparison and optimization workflows in MoAI-ADK projects.

## Overview

The Template Optimizer skill manages the complete lifecycle of template updates, ensuring that user customizations are preserved while maintaining the latest template structure and features.

## Key Features

- **🔍 Intelligent Backup Discovery**: Automatically finds and analyzes existing project backups
- **🔄 Smart Template Merging**: Preserves user customizations while updating template structure
- **📊 Comprehensive Reporting**: Detailed analysis and optimization reports
- **⚡ Performance Optimized**: Fast comparison algorithms and incremental updates
- **🛡️ Safe Operations**: Built-in validation, rollback, and recovery mechanisms
- **🔧 Flexible Configuration**: Multiple operation modes and customization options

## Quick Start

### Basic Template Optimization
```python
Skill("moai-project-template-optimizer")
```

### Analyze Without Changes
```python
Skill("moai-project-template-optimizer", mode="analyze_only")
```

### Restore from Backup
```python
Skill("moai-project-template-optimizer", 
       mode="restore", 
       backup="backup-2025-10-15-v0.16.0")
```

## Operation Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `optimize` | Complete optimization workflow (default) | Regular template updates |
| `analyze_only` | Analyze backups and differences | Pre-update assessment |
| `restore` | Restore from specific backup | Recovery from issues |
| `rollback` | Rollback to previous state | Undo optimization |

## File Structure

```
.claude/skills/moai-project-template-optimizer/
├── SKILL.md           # Main skill documentation
├── examples.md        # Usage examples and scenarios
├── reference.md       # Complete API reference
└── README.md          # This file - quick start guide
```

## Integration

The Template Optimizer integrates seamlessly with:

- **Alfred Commands**: `/alfred:0-project`, `/alfred:3-sync`
- **moai-adk Updates**: Automatic optimization after package updates
- **Other Skills**: Language initializer, auto-fixes, user questions

## Configuration

The skill updates `.moai/config.json` with optimization metadata:

```json
{
  "template_optimization": {
    "last_optimized": "2025-11-05T12:00:00Z",
    "template_version": "0.17.0",
    "customizations_preserved": ["language", "team_settings"],
    "optimization_flags": {
      "merge_applied": true,
      "conflicts_resolved": 0
    }
  }
}
```

## Performance

- **Fast Mode**: Optimized for large projects (parallel processing)
- **Standard Mode**: Balanced performance and accuracy
- **Incremental Updates**: Only process changed files
- **Hash Comparison**: Fast file difference detection

## Safety Features

- **Backup Validation**: Verify backup integrity before use
- **Rollback Capability**: Always able to undo changes
- **Conflict Detection**: Identify and resolve merge issues
- **Version Compatibility**: Ensure template version compatibility

## Error Handling

Common errors and their solutions:

- **No backups found**: Create initial backup or proceed with template-only optimization
- **Conflicts detected**: Choose resolution strategy or manual merge
- **Version incompatibility**: Use migration path or manual intervention
- **Permission errors**: Check file permissions and access rights

## Best Practices

1. **Regular Backups**: Always maintain recent backups before optimization
2. **Test Updates**: Use `analyze_only` mode to preview changes
3. **Review Reports**: Check optimization reports for any issues
4. **Monitor Performance**: Use appropriate performance mode for project size

## Troubleshooting

For detailed troubleshooting information, see the [reference documentation](reference.md#troubleshooting-reference).

Enable debug mode for detailed logging:
```python
Skill("moai-project-template-optimizer", debug_mode=True)
```

## Version History

- **v1.0.0** (2025-11-05): Initial release with complete optimization workflow
  - Backup discovery and analysis
  - Smart merge algorithms
  - Version management
  - Performance optimizations
  - Comprehensive error handling

## Support

For issues and questions:

1. Check the [examples](examples.md) for usage patterns
2. Review the [reference documentation](reference.md) for detailed API information
3. Enable debug mode to gather diagnostic information
4. Contact the MoAI-ADK team for additional support

## Contributing

When contributing to the Template Optimizer skill:

1. Follow MoAI-ADK skill structure standards
2. Maintain backward compatibility
3. Add comprehensive tests for new features
4. Update documentation with changes
5. Test with various project configurations

---

*This skill is part of the MoAI-ADK framework and requires Alfred integration for full functionality.*
