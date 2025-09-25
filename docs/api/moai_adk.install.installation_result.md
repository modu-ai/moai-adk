# moai_adk.install.installation_result

@FEATURE:INSTALL-RESULT-001 Installation result data structure for MoAI-ADK.

@TASK:RESULT-001 Contains the result of project installation operations with success status,
created files, and next steps for the user.

## Functions

### __post_init__

Initialize optional fields if not provided.

```python
__post_init__(self)
```

### has_errors

Check if installation had any errors.

```python
has_errors(self)
```

### has_warnings

Check if installation had any warnings.

```python
has_warnings(self)
```

### add_error

Add an error message to the result.

```python
add_error(self, error)
```

### add_warning

Add a warning message to the result.

```python
add_warning(self, warning)
```

### get_summary

Generate a summary of the installation result.

```python
get_summary(self)
```

### get_file_count_by_type

Count files by their extensions.

```python
get_file_count_by_type(self)
```

### to_dict

Convert result to dictionary for serialization.

```python
to_dict(self)
```

### create_success

Create a successful installation result.

```python
create_success(cls, project_path, config, files_created, next_steps, git_initialized, backup_created)
```

### create_failure

Create a failed installation result.

```python
create_failure(cls, project_path, config, error, files_created)
```

## Classes

### InstallationResult

@TASK:RESULT-DATA-001 Result of project installation with comprehensive status information.
