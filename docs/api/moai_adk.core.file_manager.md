# moai_adk.core.file_manager

@FEATURE:FILE-001 File management utilities for MoAI-ADK

Handles file copying, template rendering, and file system operations
with security validation and error handling.

## Functions

### __init__

Initialize file manager.

Args:
    template_dir: Directory containing template files
    security_manager: Security manager instance for validation

```python
__init__(self, template_dir, security_manager)
```

### copy_template_files

Copy template files matching pattern with security validation.

Args:
    source_dir: Source directory containing templates
    target_dir: Target directory for copied files
    pattern: Glob pattern to match files
    preserve_permissions: Whether to preserve file permissions

Returns:
    List[Path]: List of successfully copied files

```python
copy_template_files(self, source_dir, target_dir, pattern, preserve_permissions)
```

### render_template_file

Render a template file with context variables.

Args:
    template_path: Path to template file
    context: Variables to substitute in template

Returns:
    str: Rendered template content

```python
render_template_file(self, template_path, context)
```

### copy_and_render_template

Copy and render a single template file.

Args:
    source_path: Source template file
    target_path: Target file path
    context: Template variables
    create_dirs: Whether to create parent directories

Returns:
    bool: True if successful

```python
copy_and_render_template(self, source_path, target_path, context, create_dirs)
```

### copy_hook_scripts

Copy MoAI Hook scripts to target directory.

Args:
    target_dir: Directory to copy hooks to

Returns:
    List[Path]: List of copied hook files

```python
copy_hook_scripts(self, target_dir)
```

### copy_verification_scripts

Copy MoAI verification scripts to target directory.

Args:
    target_dir: Directory to copy scripts to

Returns:
    List[Path]: List of copied script files

```python
copy_verification_scripts(self, target_dir)
```

### install_output_styles

Install MoAI-ADK output styles with template rendering.

Args:
    target_dir: Directory to install styles to
    context: Template context for rendering

Returns:
    List[Path]: List of installed style files

```python
install_output_styles(self, target_dir, context)
```

### create_gitignore

Create a comprehensive .gitignore file.

Args:
    gitignore_path: Path where .gitignore should be created

Returns:
    bool: True if successful

```python
create_gitignore(self, gitignore_path)
```

## Classes

### FileManager

@TASK:FILE-MANAGER-001 Manages file operations for MoAI-ADK installation.
