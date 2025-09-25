# moai_adk.core.docs.documentation_builder

@FEATURE:DOCS-001 MkDocs documentation builder

Handles MkDocs Material site initialization, building, and validation.
Follows TRUST 5 principles for maintainable documentation automation.

@REQ:DOCS-SITE-001 → @TASK:DOC-BUILDER-001

## Functions

### __init__

Initialize documentation builder

Args:
    project_root: Root directory of the project

```python
__init__(self, project_root)
```

### initialize_site

@TASK:DOC-BUILDER-002 Initialize MkDocs site structure

```python
initialize_site(self)
```

### _create_mkdocs_config

Create basic mkdocs.yml configuration

```python
_create_mkdocs_config(self)
```

### _create_essential_files

Create essential documentation files

```python
_create_essential_files(self)
```

### validate_config

@TASK:DOC-BUILDER-003 Validate MkDocs configuration

```python
validate_config(self)
```

### build_docs

@TASK:DOC-BUILDER-004 Build documentation site

```python
build_docs(self, incremental)
```

### get_build_status

Get current build status information

```python
get_build_status(self)
```

### validate_links

@TASK:DOC-BUILDER-005 Validate internal links

```python
validate_links(self)
```

### check_completeness

@TASK:DOC-BUILDER-006 Check documentation completeness

```python
check_completeness(self)
```

## Classes

### DocumentationBuilder

@FEATURE:DOCS-002 Main documentation builder class
