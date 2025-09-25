# moai_adk.core.docs.api_generator

@FEATURE:API-GEN-001 API documentation generator

Automatic generation of API documentation from Python source code.
Extracts docstrings and creates structured markdown documentation.

@REQ:API-GEN-001 → @TASK:API-GEN-001

## Functions

### __init__

Initialize API generator

Args:
    project_root: Root directory of the project
    source_dir: Source code directory relative to project root

```python
__init__(self, project_root, source_dir)
```

### scan_modules

@TASK:API-GEN-002 Scan and find Python modules

```python
scan_modules(self)
```

### parse_module_docs

@TASK:API-GEN-003 Parse docstrings from a module

```python
parse_module_docs(self, module_path)
```

### generate_nav_structure

@TASK:API-GEN-004 Generate navigation structure for API docs

```python
generate_nav_structure(self)
```

### generate_api_docs

@TASK:API-GEN-005 Generate API documentation files

```python
generate_api_docs(self, docs_root)
```

### _generate_module_markdown

Generate markdown content for a module

```python
_generate_module_markdown(self, module_name, doc_info)
```

### update_mkdocs_nav

@TASK:API-GEN-006 Update mkdocs.yml with API navigation

```python
update_mkdocs_nav(self, mkdocs_config_path)
```

### generate_module_index

@TASK:API-GEN-007 Generate module index content

```python
generate_module_index(self)
```

## Classes

### ModuleInfo

Information about a Python module

### ApiGenerator

@FEATURE:API-GEN-002 Automatic API documentation generator
