# Code Documentation Patterns

## Overview

Generate documentation from source code using established tools: Sphinx for Python, TypeDoc for TypeScript, JSDoc for JavaScript, and mkdocstrings for MkDocs integration.

## Python Documentation with Sphinx

### Setup and Configuration

Install Sphinx and extensions:
- sphinx - Core documentation generator
- sphinx-autodoc-typehints - Type hint integration
- sphinx-rtd-theme - Read the Docs theme
- myst-parser - Markdown support in Sphinx

Initialize project with sphinx-quickstart in docs directory.

Configure conf.py:
- Add source directory to sys.path for autodoc
- Enable extensions: autodoc, napoleon, typehints, viewcode
- Set html_theme to sphinx_rtd_theme
- Configure autodoc options for member ordering

### Writing Docstrings

Use Google style or NumPy style docstrings:
- Include summary line (one sentence)
- Add detailed description if needed
- Document all parameters with types
- Document return values
- Include usage examples
- Document raised exceptions

### Generating Documentation

Run sphinx-apidoc to generate rst files from source:
- Specify source directory and output directory
- Use --separate for one file per module
- Use --force to overwrite existing files

Build HTML documentation with make html or sphinx-build command.

## Python Documentation with MkDocs

### mkdocstrings Integration

Install mkdocs-material and mkdocstrings-python.

Configure mkdocs.yml:
- Add mkdocstrings plugin
- Configure Python handler with options
- Set docstring style (google, numpy, sphinx)
- Enable signature rendering

### Usage in Markdown

Reference Python objects with ::: syntax:
- Use ::: module.Class to document a class
- Use ::: module.function to document a function
- Configure display options per reference

Organize documentation with nav structure in mkdocs.yml.

## TypeScript Documentation with TypeDoc

### Setup

Install typedoc as development dependency.

Create typedoc.json configuration:
- Set entryPoints to source files or directories
- Configure out for output directory
- Enable excludePrivate and excludeProtected as needed
- Set includeVersion for version display
- Configure theme (default or custom)

### Writing TSDoc Comments

Use TSDoc comment format:
- Start with summary paragraph
- Use @param for parameters
- Use @returns for return value
- Use @example for code examples
- Use @see for cross-references
- Use @deprecated for deprecated items

### Plugin Ecosystem

Enhance TypeDoc with plugins:
- typedoc-plugin-markdown for Markdown output
- typedoc-plugin-missing-exports for coverage
- Custom themes for branding

## JavaScript Documentation with JSDoc

### Setup

Install jsdoc as development dependency.

Create jsdoc.json configuration:
- Configure source paths and patterns
- Set output destination
- Enable plugins like markdown
- Configure templates

### Writing JSDoc Comments

Document functions with JSDoc tags:
- @param for parameters with type and description
- @returns for return value
- @throws for exceptions
- @example for usage examples
- @typedef for custom types
- @callback for callback definitions

Document modules with @module tag at file start.

## Documentation Coverage

### Python Coverage with interrogate

Install interrogate for docstring coverage analysis.

Configure in pyproject.toml:
- Set minimum coverage threshold
- Exclude specific patterns or files
- Configure output format

Run in CI to enforce documentation standards.

### TypeScript Coverage

Use typedoc with strict mode to identify undocumented items.

Configure TSDoc linting with eslint-plugin-tsdoc.

## Integration with Editors

VS Code Extensions:
- Python Docstring Generator
- Document This for TypeScript/JavaScript
- TSDoc Comment Tags

IntelliJ/WebStorm:
- Built-in documentation generation
- Quick documentation preview

---

Version: 2.0.0
Last Updated: 2025-12-30
