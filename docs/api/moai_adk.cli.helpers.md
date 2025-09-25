# moai_adk.cli.helpers

@FEATURE:CLI-HELPERS-001 CLI Helper Functions for MoAI-ADK

@TASK:CLI-UTILS-001 Contains utility functions used by the CLI commands including backup,
conflict detection, environment validation, and project analysis.

## Functions

### create_installation_backup

@TASK:BACKUP-001 Create a backup of existing MoAI-ADK installation.

Args:
    project_path: Path to the project directory

Returns:
    bool: True if backup was created successfully, False otherwise

```python
create_installation_backup(project_path)
```

### detect_potential_conflicts

Detect potential conflicts with existing files/directories.

Args:
    project_path: Path to check for conflicts

Returns:
    list: List of potential conflict descriptions

```python
detect_potential_conflicts(project_path)
```

### analyze_existing_project

Analyze existing project structure and provide recommendations.

Args:
    project_path: Path to analyze

Returns:
    dict: Analysis results with recommendations

```python
analyze_existing_project(project_path)
```

### print_banner

Print MoAI-ADK banner with optional usage information.

```python
print_banner(show_usage)
```

### validate_environment

Validate the current environment for MoAI-ADK installation.

Returns:
    bool: True if environment is suitable

```python
validate_environment()
```

### format_project_status

Format project status information for display.

Args:
    project_path: Path to the project
    config_data: Optional configuration data

Returns:
    dict: Formatted status information

```python
format_project_status(project_path, config_data)
```
