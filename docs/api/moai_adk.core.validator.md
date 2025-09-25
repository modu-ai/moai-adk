# moai_adk.core.validator

@FEATURE:VALIDATOR-001 🗿 MoAI-ADK Validation Utilities

Provides validation functions for environment, configuration, and project setup.

## Functions

### validate_python_version

@TASK:VALIDATE-PYTHON-001 Validate Python version meets minimum requirements

Args:
    min_version: Minimum required Python version as tuple

Returns:
    True if version is valid, False otherwise

```python
validate_python_version(min_version)
```

### validate_claude_code

Validate that Claude Code is available and properly configured.

Returns:
    True if Claude Code is available, False otherwise

```python
validate_claude_code()
```

### validate_git_repository

Validate that the path is within a git repository.

Args:
    path: Path to validate
    
Returns:
    True if valid git repository, False otherwise

```python
validate_git_repository(path)
```

### validate_project_structure

Validate MoAI-ADK project structure.

Args:
    project_path: Path to the project
    
Returns:
    Dictionary of validation results

```python
validate_project_structure(project_path)
```

### validate_environment

Comprehensive environment validation.

Returns:
    True if environment is valid, False otherwise

```python
validate_environment()
```

### validate_project_readiness

Validate project is ready for MoAI-ADK integration.

Args:
    project_path: Path to the project
    
Returns:
    True if project is ready, False otherwise

```python
validate_project_readiness(project_path)
```

### validate_moai_structure

Validate complete MoAI-ADK project structure.

Args:
    project_path: Path to the project
    
Returns:
    Dictionary of validation results for MoAI components

```python
validate_moai_structure(project_path)
```

### validate_trust_principles_compliance

Validate project compliance with MoAI 개발 가이드 5 principles.

Args:
    project_path: Path to the project
    
Returns:
    Dictionary with compliance results for each principle

```python
validate_trust_principles_compliance(project_path)
```

### _calculate_project_complexity

Calculate project complexity score (1-10).

```python
_calculate_project_complexity(project_path)
```

### _check_architectural_modularity

Check architectural modularity (0-100).

```python
_check_architectural_modularity(project_path)
```

### _check_testing_setup

Check testing infrastructure (0-100).

```python
_check_testing_setup(project_path)
```

### _check_observability_setup

Check observability setup (0-100).

```python
_check_observability_setup(project_path)
```

### _check_versioning_setup

Check versioning setup (0-100).

```python
_check_versioning_setup(project_path)
```

### run_full_validation

Run complete MoAI-ADK validation suite.

Args:
    project_path: Path to validate
    verbose: Whether to print detailed results
    
Returns:
    Complete validation results

```python
run_full_validation(project_path, verbose)
```

### validate_global_resources

Validate global MoAI-ADK resources and install if needed.

Returns:
    bool: True if global resources are available, False otherwise

```python
validate_global_resources()
```
