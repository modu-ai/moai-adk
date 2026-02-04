# moai-workflow-project Reference

Progressive Disclosure Level 2: Extended documentation for advanced users and integrators.

---

## API Reference

### Core Classes

MoaiMenuProject:
- Purpose: Unified interface for all project management operations
- Initialization: `MoaiMenuProject(project_path: str)`
- Primary Methods:
  - `initialize_complete_project(language, user_name, domains, project_type, optimization_enabled)` - Full project setup
  - `generate_documentation_from_spec(spec_data)` - SPEC-driven doc generation
  - `optimize_project_templates(options)` - Template optimization
  - `get_project_status()` - Comprehensive status report
  - `update_language_settings(settings)` - Language configuration update

DocumentationManager:
- Purpose: Template-based documentation generation
- Key Methods:
  - `generate_docs(project_type, language)` - Generate documentation set
  - `update_docs_from_spec(spec_data)` - Update from SPEC data
  - `export_docs(format, language)` - Multi-format export (md, html, pdf)
  - `detect_project_type()` - Auto-detect project type

LanguageInitializer:
- Purpose: Language detection and configuration
- Key Methods:
  - `detect_project_language()` - Analyze project for language
  - `create_multilingual_documentation_structure(language)` - Setup multilingual docs
  - `localize_agent_prompts(base_prompt, language)` - Prompt localization
  - `calculate_token_cost_impact(language)` - Token cost analysis

TemplateOptimizer:
- Purpose: Template analysis and optimization
- Key Methods:
  - `analyze_project_templates()` - Comprehensive template analysis
  - `create_optimized_templates(options)` - Apply optimizations
  - `create_backup()` - Create template backup
  - `restore_from_backup(backup_id)` - Restore from backup

---

## Configuration Options

### Project Configuration Schema

```yaml
project:
  name: "string"              # Project display name
  type: "string"              # web_application, mobile_app, cli_tool, library, ml_project
  initialized_at: "datetime"  # ISO 8601 timestamp
  version: "string"           # Semantic version

language:
  conversation_language: "string"   # en, ko, ja, zh, es, fr, de
  agent_prompt_language: "string"   # english (cost-optimized) or localized
  documentation_language: "string"  # Primary documentation language
  code_comments: "string"           # Code comment language

menu_system:
  version: "string"           # Menu system version
  fully_initialized: boolean  # Complete initialization status
  modules_enabled: []         # List of enabled modules
```

### Optimization Options Schema

```yaml
backup_first: boolean                    # Create backup before optimization
apply_size_optimizations: boolean        # Reduce template file sizes
apply_performance_optimizations: boolean # Improve template performance
apply_complexity_optimizations: boolean  # Reduce template complexity
preserve_functionality: boolean          # Maintain all existing features
max_complexity_score: number             # Maximum allowed complexity (1-10)
```

### Language Configuration Presets

Supported Languages with Token Impact:

| Language | Code | Locale       | Token Cost Impact |
|----------|------|--------------|-------------------|
| English  | en   | en_US.UTF-8  | 0% (baseline)     |
| Korean   | ko   | ko_KR.UTF-8  | +20%              |
| Japanese | ja   | ja_JP.UTF-8  | +25%              |
| Chinese  | zh   | zh_CN.UTF-8  | +15%              |
| Spanish  | es   | es_ES.UTF-8  | +5%               |
| French   | fr   | fr_FR.UTF-8  | +5%               |
| German   | de   | de_DE.UTF-8  | +5%               |

---

## Integration Patterns

### Pattern 1: SPEC-Driven Documentation Workflow

```python
# Integration with /moai:1-plan and /moai:3-sync
from moai_workflow_project import MoaiMenuProject

# Initialize project system
project = MoaiMenuProject("/path/to/project")

# After /moai:1-plan generates SPEC
spec_data = load_spec("SPEC-001")

# Generate documentation from SPEC
docs_result = project.generate_documentation_from_spec(spec_data)

# Automatically creates:
# - Feature documentation with requirements
# - API documentation with endpoint details
# - Architecture documentation
# - Multilingual versions if configured
```

### Pattern 2: CI/CD Integration

```python
# Integration with GitHub Actions or similar
from moai_workflow_project import MoaiMenuProject

def ci_documentation_check(project_path: str) -> dict:
    """Run documentation validation in CI pipeline."""
    project = MoaiMenuProject(project_path)

    # Get current documentation status
    status = project.get_project_status()

    # Validate completeness
    validation_result = {
        "docs_complete": status.documentation_completion >= 0.9,
        "language_configured": status.language_configured,
        "templates_optimized": status.templates_optimized,
        "warnings": status.warnings,
        "errors": status.errors
    }

    return validation_result
```

### Pattern 3: Multi-Project Template Sharing

```python
# Share optimized templates across projects
from moai_workflow_project import TemplateOptimizer

# Optimize master templates
master_optimizer = TemplateOptimizer("/templates/master")
optimized = master_optimizer.create_optimized_templates({
    "backup_first": True,
    "apply_all_optimizations": True
})

# Deploy to multiple projects
for project_path in project_list:
    project = MoaiMenuProject(project_path)
    project.import_templates(optimized.templates)
```

### Pattern 4: Language-Aware Agent Delegation

```python
# Integrate with MoAI's delegation patterns
def delegate_with_language_context(task: str, language: str):
    """Delegate task with proper language context."""
    project = MoaiMenuProject(".")

    # Get localized prompt
    localized_task = project.language_initializer.localize_agent_prompts(
        base_prompt=task,
        language=language
    )

    # Token cost analysis
    cost_impact = project.language_initializer.calculate_token_cost_impact(language)

    return {
        "localized_task": localized_task,
        "estimated_token_overhead": cost_impact
    }
```

---

## Troubleshooting

### Common Issues

Issue: Documentation generation fails with template not found:
- Cause: Template directory missing or corrupted
- Solution: Run `project.template_optimizer.restore_from_backup()` or reinstall templates
- Prevention: Always enable `backup_first` option before optimization

Issue: Language detection returns incorrect language:
- Cause: Insufficient language indicators in project files
- Solution: Manually set language in configuration: `project.update_language_settings({"language.conversation_language": "ko"})`
- Prevention: Include language comments in main source files

Issue: Template optimization causes functionality loss:
- Cause: Aggressive optimization removed necessary content
- Solution: Restore from backup: `project.template_optimizer.restore_from_backup(backup_id)`
- Prevention: Set `preserve_functionality: True` in optimization options

Issue: Multilingual documentation structure incomplete:
- Cause: Partial initialization or interrupted process
- Solution: Re-run `project.language_initializer.create_multilingual_documentation_structure(language)`
- Prevention: Ensure stable connection during initialization

Issue: High token cost for non-English languages:
- Cause: Localized agent prompts increase token usage
- Solution: Use `agent_prompt_language: "english"` with `conversation_language: "ko"` for cost optimization
- Prevention: Configure language settings before heavy usage

### Diagnostic Commands

```python
# Full diagnostic report
status = project.get_project_status()
print(f"Initialization: {status.initialization_complete}")
print(f"Language: {status.language_configuration}")
print(f"Documentation: {status.documentation_completion}%")
print(f"Templates: {status.template_status}")
print(f"Errors: {status.errors}")
print(f"Warnings: {status.warnings}")
```

### Log Locations

- Project logs: `.moai/logs/project.log`
- Template optimization logs: `.moai/logs/template-optimizer.log`
- Language initialization logs: `.moai/logs/language-init.log`

---

## External Resources

### Official Documentation

- MoAI-ADK Documentation: https://github.com/moai-adk/docs
- Claude Code Skills Guide: https://docs.anthropic.com/claude-code/skills
- SPEC-First DDD Methodology: See `moai-foundation-core/modules/spec-first-ddd.md`

### Related Skills

- moai-foundation-core - Core execution patterns and SPEC workflow
- moai-foundation-claude - Claude Code integration patterns
- moai-workflow-docs - Unified documentation management
- moai-workflow-templates - Template optimization strategies
- moai-library-nextra - Advanced documentation architecture

### Template Resources

- Documentation Templates: `/templates/doc-templates/`
- Product Template: `/templates/doc-templates/product-template.md`
- Technical Template: `/templates/doc-templates/tech-template.md`
- Structure Template: `/templates/doc-templates/structure-template.md`

### Version History

| Version | Date       | Changes                                           |
|---------|------------|---------------------------------------------------|
| 2.0.0   | 2025-11-27 | Integrated modular architecture                   |
| 1.5.0   | 2025-11-20 | Added template optimization module                |
| 1.0.0   | 2025-11-15 | Initial release with documentation management     |

---

Status: Reference Documentation Complete
Last Updated: 2025-12-06
Skill Version: 2.0.0
