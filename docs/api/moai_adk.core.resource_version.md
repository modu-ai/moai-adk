# moai_adk.core.resource_version

@FEATURE:RESOURCE-VERSION-001 Resource version tracking utilities for MoAI-ADK projects.

## Functions

### __init__

```python
__init__(self, project_path)
```

### read

Return version metadata if available, otherwise defaults.

```python
read(self)
```

### write

Persist version metadata for the project.

```python
write(self, template_version, package_version)
```

### is_outdated

```python
is_outdated(self, expected_template_version)
```

## Classes

### ResourceVersionManager

@TASK:RESOURCE-VERSION-MANAGER-001 Read and write MoAI resource/template version metadata.
