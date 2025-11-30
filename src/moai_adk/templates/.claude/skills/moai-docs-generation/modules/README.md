# Documentation Generation Modules

This directory contains modular components for comprehensive documentation generation workflows.

## Module Structure

### Core Modules

- api-documentation.md - OpenAPI/Swagger generation and API documentation
- code-documentation.md - AST-based code analysis and AI-powered docstring enhancement
- user-guides.md - Tutorial and getting started guide generation
- multi-format-output.md - HTML, PDF, and static site generation

### Usage Patterns

1. API Documentation: Use `api-documentation.md` for automated API docs
2. Code Analysis: Use `code-documentation.md` for source code documentation
3. User Guides: Use `user-guides.md` for tutorials and guides
4. Output Generation: Use `multi-format-output.md` for format conversion

### Integration Guidelines

Each module is self-contained but designed to work together:

```python
# Complete documentation workflow
api_docs = API Documentation Generation
code_analysis = Code Documentation Enhancement
user_guides = User Guide Generation
output_formats = Multi-Format Output Generation
```

### Progressive Disclosure

- Quick Start: Use individual modules for specific tasks
- Implementation: Combine modules for complete workflows
- Advanced: Customize templates and extend functionality

### Dependencies

- Required: Python 3.8+, FastAPI/Flask for API docs
- Optional: AI client for enhanced documentation
- Optional: WeasyPrint for PDF generation
- Optional: Jinja2 for HTML templating
