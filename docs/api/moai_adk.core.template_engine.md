# moai_adk.core.template_engine

@FEATURE:TEMPLATE-001 Template engine for dynamic file generation in MoAI-ADK

This module provides a simple template system for generating project files
from templates with variable substitution.

## Functions

### __init__

Initialize template engine.

Args:
    template_dir: Path to template directory containing _templates/

```python
__init__(self, template_dir)
```

### create_from_template

Create a file from template with variable substitution.

Args:
    template_name: Template file name (e.g., 'specs/spec.template.md')
    target_path: Target file path to create
    context: Variables for template substitution
    create_dirs: Whether to create parent directories

Returns:
    bool: True if successful

```python
create_from_template(self, template_name, target_path, context, create_dirs)
```

### create_spec_from_template

Create SPEC file from template.

```python
create_spec_from_template(self, spec_id, spec_name, description, target_path)
```

### create_steering_from_template

Create Steering document from template.

```python
create_steering_from_template(self, steering_type, project_name, context, target_path)
```

### create_constitution_from_template

Create 개발 가이드 document from template.

```python
create_constitution_from_template(self, project_name, project_type, target_path)
```

### should_copy_as_template

Determine if a file should be processed as a template.

Args:
    file_path: Path to check

Returns:
    bool: True if file should be templated

```python
should_copy_as_template(self, file_path)
```

### _render_template

Render template content with context variables including version info.

Args:
    template_content: Raw template content
    context: Variables for substitution

Returns:
    str: Rendered content

```python
_render_template(self, template_content, context)
```

### _enhance_context_with_version

Enhance context with automatic version information.

Args:
    context: Original context

Returns:
    Dict[str, Any]: Enhanced context with version info

```python
_enhance_context_with_version(self, context)
```

### _create_spec_context

Create context for SPEC template rendering.

```python
_create_spec_context(self, spec_id, spec_name, description)
```

## Classes

### TemplateEngine

@TASK:TEMPLATE-ENGINE-001 Simple template engine for MoAI-ADK project file generation

Handles dynamic creation of project files from templates with
variable substitution and context management.
