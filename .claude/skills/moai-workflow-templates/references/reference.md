# moai-workflow-templates Reference

Progressive Disclosure Level 2: Extended documentation for template management and optimization.

---

## API Reference

### Core Classes

TemplateManager:
- Purpose: Central template management and scaffolding
- Key Methods:
  - `load_template(template_name)` - Load template by name
  - `scaffold(name, features, customizations)` - Generate project from template
  - `list_templates(category)` - List available templates
  - `validate_template(template_path)` - Validate template structure

TemplateOptimizer:
- Purpose: Smart merge and optimization operations
- Key Methods:
  - `smart_merge(backup, template, current)` - Three-way intelligent merge
  - `extract_user_customizations(backup)` - Extract user modifications
  - `apply_customizations(customizations)` - Apply extracted customizations
  - `create_backup(backup_id)` - Create timestamped backup
  - `restore_from_backup(backup_id)` - Restore from specific backup

FeedbackTemplateGenerator:
- Purpose: GitHub issue template generation
- Key Methods:
  - `generate_issue(type, data)` - Generate issue from template
  - `get_template(issue_type)` - Get template by type
  - `validate_issue_data(data)` - Validate issue fields

### Template Types

Code Templates:
- `backend/fastapi` - FastAPI REST API project
- `backend/django` - Django web application
- `frontend/react` - React SPA application
- `frontend/nextjs` - Next.js full-stack application
- `frontend/vue` - Vue.js application
- `fullstack/fastapi-react` - Full-stack combination
- `infra/docker` - Docker containerization
- `infra/cicd` - CI/CD pipeline configuration

Feedback Templates:
- `bug-report` - Bug report with reproduction steps
- `feature-request` - Feature request with use cases
- `improvement` - Enhancement suggestion
- `refactor` - Code refactoring proposal
- `documentation` - Documentation update request
- `question` - Discussion and question template

---

## Configuration Options

### Template Variables Schema

```json
{
  "variables": {
    "PROJECT_NAME": "string",      // Project identifier
    "AUTHOR": "string",            // Author name
    "LICENSE": "string",           // License type (MIT, Apache-2.0, etc.)
    "PYTHON_VERSION": "string",    // Python version for backend
    "NODE_VERSION": "string",      // Node.js version for frontend
    "DATABASE": "string",          // Database type (postgresql, mysql, mongodb)
    "AUTH_TYPE": "string",         // Authentication type (jwt, oauth, session)
    "DEPLOYMENT": "string"         // Deployment target (docker, k8s, serverless)
  },
  "files": {
    "pattern": "action"            // substitute, copy, ignore
  },
  "hooks": {
    "post_generate": ["command"]   // Post-generation commands
  }
}
```

### Optimization Configuration

```yaml
template_optimization:
  last_optimized: "datetime"        # Last optimization timestamp
  backup_version: "string"          # Backup reference version
  template_version: "string"        # Current template version
  customizations_preserved:         # List of preserved customizations
    - "language"
    - "team_settings"
    - "domains"
  merge_strategy: "smart"           # smart, force, manual
  conflict_resolution: "preserve_user"  # preserve_user, use_template, manual
```

### Feedback Template Fields

Bug Report Fields:
- `title` (required): Brief description of the bug
- `environment` (required): OS, browser, versions
- `reproduction_steps` (required): Steps to reproduce
- `expected_behavior` (required): What should happen
- `actual_behavior` (required): What actually happens
- `screenshots` (optional): Visual evidence
- `additional_context` (optional): Extra information

Feature Request Fields:
- `title` (required): Feature name
- `problem_statement` (required): Problem being solved
- `proposed_solution` (required): How feature would work
- `use_cases` (required): List of use cases
- `alternatives_considered` (optional): Other approaches
- `additional_context` (optional): Extra information

---

## Integration Patterns

### Pattern 1: Project Scaffolding Workflow

```python
# Complete project scaffolding with customizations
from moai_workflow_templates import TemplateManager

manager = TemplateManager()

# Load and configure template
template = manager.load_template("backend/fastapi")

# Scaffold with features
project = template.scaffold(
    name="my-api",
    features=["auth", "database", "celery", "redis"],
    customizations={
        "database": "postgresql",
        "auth_type": "jwt",
        "python_version": "3.13"
    }
)

# Post-generation setup
project.run_hooks()  # Install deps, init git, etc.
```

### Pattern 2: Template Update with Smart Merge

```python
# Update templates while preserving customizations
from moai_workflow_templates import TemplateOptimizer

optimizer = TemplateOptimizer("/project/.moai")

# Step 1: Create backup
backup_id = optimizer.create_backup()

# Step 2: Extract customizations from current state
customizations = optimizer.extract_user_customizations(backup_id)

# Step 3: Get latest templates
new_templates = optimizer.get_latest_templates()

# Step 4: Smart merge
merged = optimizer.smart_merge(
    backup=customizations,
    template=new_templates,
    current=optimizer.current_templates
)

# Step 5: Apply merged templates
optimizer.apply_templates(merged)
```

### Pattern 3: Feedback Command Integration

```python
# Integration with /moai:9-feedback command
from moai_workflow_templates import FeedbackTemplateGenerator

def handle_feedback_command(feedback_type: str, data: dict):
    """Generate GitHub issue from feedback."""
    generator = FeedbackTemplateGenerator()

    # Validate input
    validation = generator.validate_issue_data(data)
    if not validation.is_valid:
        return {"error": validation.errors}

    # Generate issue content
    issue_content = generator.generate_issue(
        type=feedback_type,
        data=data
    )

    # Create GitHub issue via gh CLI
    result = create_github_issue(
        title=issue_content.title,
        body=issue_content.body,
        labels=issue_content.labels
    )

    return {"issue_url": result.url}
```

### Pattern 4: Version-Controlled Template Management

```python
# Track template versions across projects
from moai_workflow_templates import VersionManager

version_manager = VersionManager()

# Get current version info
current = version_manager.get_current_version()
print(f"Template version: {current.version}")
print(f"Last updated: {current.last_updated}")

# Check for updates
updates = version_manager.check_for_updates()
if updates.available:
    print(f"New version: {updates.new_version}")
    print(f"Changes: {updates.changelog}")

    # Apply update with backup
    version_manager.apply_update(
        backup_first=True,
        preserve_customizations=True
    )
```

---

## Troubleshooting

### Common Issues

Issue: Template scaffolding fails with missing variables:
- Cause: Required variables not provided in customizations
- Solution: Check template.json for required variables and provide all values
- Prevention: Use `template.get_required_variables()` before scaffolding

Issue: Smart merge creates conflicts:
- Cause: Significant changes in both template and user customizations
- Solution: Review conflicts in `.moai/merge-conflicts/` and resolve manually
- Prevention: Regularly update templates to minimize drift

Issue: Post-generation hooks fail:
- Cause: Missing dependencies or incorrect environment
- Solution: Ensure prerequisites (npm, pip, docker) are installed
- Prevention: Run `template.check_prerequisites()` before scaffolding

Issue: Backup restoration incomplete:
- Cause: Corrupted backup or partial backup creation
- Solution: List available backups with `optimizer.list_backups()` and try older version
- Prevention: Verify backup with `optimizer.verify_backup(backup_id)` after creation

Issue: Feedback template missing fields:
- Cause: Template customizations removed required fields
- Solution: Reset to default template: `generator.reset_template(type)`
- Prevention: Use `validate_template_structure()` after customizations

### Diagnostic Commands

```python
# Template system diagnostics
from moai_workflow_templates import diagnose

# Run full diagnostics
report = diagnose.run_full_check()
print(f"Templates valid: {report.templates_valid}")
print(f"Backups available: {report.backup_count}")
print(f"Version status: {report.version_status}")
print(f"Issues found: {report.issues}")

# Individual checks
diagnose.check_template_integrity()
diagnose.verify_all_backups()
diagnose.validate_customizations()
```

### File Locations

- Template storage: `.claude/skills/moai-workflow-templates/templates/`
- Backups: `.moai/backups/templates/`
- Merge conflicts: `.moai/merge-conflicts/`
- Version history: `.moai/config/template-versions.json`
- Feedback templates: `.claude/skills/moai-workflow-templates/modules/feedback-templates.md`

---

## External Resources

### Official Documentation

- MoAI-ADK Templates Guide: See `modules/code-templates.md`
- Feedback Templates Reference: See `modules/feedback-templates.md`
- Template Optimizer Guide: See `modules/template-optimizer.md`

### Related Skills

- moai-workflow-project - Project initialization with templates
- moai-foundation-core - SPEC-driven template generation
- moai-docs-generation - Documentation template scaffolding
- moai-cc-configuration - Claude Code settings integration

### Template Development Resources

- Template Variable Syntax: `${VARIABLE_NAME}` for substitution
- Conditional Content: `{{#if CONDITION}}...{{/if}}`
- File Patterns: Glob syntax for file matching
- Hook Commands: Shell commands for post-generation

### Best Practices

Template Design:
- Keep templates focused on single purpose
- Use descriptive variable names
- Include comprehensive README in each template
- Test templates before publishing

Customization Management:
- Document all customizations
- Keep customizations minimal and focused
- Use version control for template modifications
- Regular backup before major changes

Smart Merge:
- Review merge results before applying
- Test merged templates in staging
- Keep customization history for rollback
- Document conflict resolution decisions

### Version History

| Version | Date       | Changes                                           |
|---------|------------|---------------------------------------------------|
| 3.0.0   | 2025-11-24 | Unified code, feedback, and optimizer modules     |
| 2.5.0   | 2025-11-20 | Added smart merge algorithm                       |
| 2.0.0   | 2025-11-15 | Introduced template version management            |
| 1.0.0   | 2025-11-10 | Initial release with code templates               |

---

Status: Reference Documentation Complete
Last Updated: 2025-12-06
Skill Version: 3.0.0
