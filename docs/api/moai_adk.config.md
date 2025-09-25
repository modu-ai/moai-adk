# moai_adk.config

@FEATURE:CONFIG-001 Configuration management for MoAI-ADK projects.

@TASK:CONFIG-MAIN-001 Handles project configuration, runtime settings, and validation.

## Functions

### __post_init__

Initialize computed fields.

```python
__post_init__(self)
```

### __init__

@TASK:CONFIG-INIT-001 Initialize Config with backward compatibility for project_path parameter.

```python
__init__(self, name)
```

### _validate

Validate configuration parameters.

```python
_validate(self)
```

### project_path

Get project path as Path object.

```python
project_path(self)
```

### project_type

Determine project type based on tech stack.

```python
project_type(self)
```

### get_template_context

Get template rendering context.

```python
get_template_context(self)
```

### to_dict

Convert config to dictionary.

```python
to_dict(self)
```

## Classes

### RuntimeConfig

@TASK:RUNTIME-CONFIG-001 Runtime configuration for the project.

### Config

@TASK:CONFIG-MAIN-001 Main configuration for MoAI-ADK projects.
