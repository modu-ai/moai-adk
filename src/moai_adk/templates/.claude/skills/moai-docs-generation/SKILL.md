---
name: moai-docs-generation
description: Automated documentation generation specialist for technical specs, API docs, user guides, and knowledge bases with multi-format output
version: 1.0.0
category: workflow
allowed-tools: Read, Write, Edit, Grep, Glob
tags:
 - documentation
 - generation
 - automation
 - markdown
 - html
 - api-docs
updated: 2025-12-06
status: active
author: MoAI-ADK Team
---

# Documentation Generation Specialist

## Quick Reference (30 seconds)

Automated Documentation Generation - Comprehensive documentation automation covering technical specs, API documentation, user guides, and knowledge base creation with multi-format output capabilities.

Core Capabilities:
- Technical Documentation: API docs, architecture specs, code documentation
- User Guides: Tutorials, getting started guides, best practices
- API Documentation: OpenAPI/Swagger generation, endpoint documentation
- Multi-Format Output: Markdown, HTML, PDF, static sites
- AI-Powered Generation: Context-aware content creation and enhancement
- Continuous Updates: Auto-sync documentation with code changes

When to Use:
- Generating API documentation from code
- Creating technical specifications and architecture docs
- Building user guides and tutorials
- Automating knowledge base creation
- Maintaining up-to-date project documentation

---

## Implementation Guide (5 minutes)

### Quick Start Workflow

Basic Documentation Generation:
```python
from moai_docs_generation import DocumentationGenerator

# Initialize generator
doc_gen = DocumentationGenerator()

# Generate API documentation
api_docs = doc_gen.generate_api_docs("path/to/your/app.py")

# Create user guide
user_guide = doc_gen.generate_user_guide(project_info)

# Export to multiple formats
doc_gen.export_to_formats(api_docs, formats=["html", "pdf", "markdown"])
```

Single Command Documentation:
```bash
# Generate complete documentation
moai generate-docs --source ./src --output ./docs --formats html,pdf

# Update API docs from code
moai update-api-docs --app-file app.py --format openapi

# Create tutorial from feature
moai create-tutorial --feature authentication --output docs/tutorials/
```

### Core Components

1. API Documentation (`modules/api-documentation.md`)
- OpenAPI/Swagger specification generation
- Interactive HTML documentation
- Code example generation
- Request/response documentation

2. Code Documentation (`modules/code-documentation.md`)
- AST-based code analysis
- AI-powered docstring enhancement
- Automatic documentation structure extraction
- Multi-format documentation generation

3. User Guides (`modules/user-guides.md`)
- Getting started guides
- Feature tutorials
- Cookbook generation
- Step-by-step instructions

4. Multi-Format Output (`modules/multi-format-output.md`)
- HTML site generation
- PDF documentation
- Static site export
- Responsive design templates

---

## Advanced Patterns (10+ minutes)

### Continuous Documentation Integration

Git Hooks for Auto-Documentation:
```python
# .git/hooks/pre-commit
#!/bin/bash
# Auto-update documentation before commits

python -c "
from moai_docs_generation import DocumentationGenerator
doc_gen = DocumentationGenerator()
doc_gen.update_documentation_for_changed_files()
"
```

CI/CD Pipeline Integration:
```yaml
# .github/workflows/docs.yml
name: Generate Documentation
on:
 push:
 branches: [main]

jobs:
 docs:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v2
 - name: Generate Documentation
 run: |
 moai generate-docs --source ./src --output ./docs
 moai deploy-docs --platform github-pages
```

### AI-Enhanced Documentation

Smart Content Generation:
```python
# AI-powered example generation
ai_enhanced_docs = doc_gen.generate_with_ai(
 source_code="path/to/code",
 enhancement_level="comprehensive",
 include_examples=True,
 include_troubleshooting=True
)
```

### Documentation Quality Automation

Automated Quality Checks:
```python
# Validate documentation quality
quality_report = doc_gen.validate_documentation(
 completeness_threshold=0.9,
 include_example_validation=True,
 check_link_integrity=True
)
```

---

## Works Well With

Complementary Skills:
- `moai-foundation-core` - SPEC-first documentation approach
- `moai-workflow-project` - Project documentation integration
- `moai-lang-python` - Python code documentation patterns
- `moai-lang-typescript` - TypeScript code documentation patterns
- `moai-mcp-notion` - Notion documentation sync and knowledge management

Technology Integration:
- FastAPI/Flask applications
- Sphinx/MkDocs documentation
- GitHub/GitLab wikis
- Confluence knowledge bases
- Static site generators (Hugo, Jekyll)

---

## Usage Examples

### Command Line Interface
```bash
# Complete documentation suite
moai docs:generate --project ./my-project --output ./docs

# API documentation only
moai docs:api --source ./app.py --format openapi,html

# User guide creation
moai docs:guide --features auth,user-management --template getting-started

# Documentation updates
moai docs:update --sync-with-code --validate-links
```

### Python API
```python
from moai_docs_generation import DocumentationGenerator

# Complete workflow
generator = DocumentationGenerator()
docs = generator.generate_comprehensive_docs(
 source_directory="./src",
 include_api_docs=True,
 include_user_guides=True,
 output_formats=["html", "pdf", "markdown"]
)

# Individual components
api_gen = API Documentation Generation()
api_spec = api_gen.generate_openapi_spec(fastapi_app)

code_gen = Code Documentation Enhancement()
enhanced_docs = code_gen.analyze_and_enhance("./src/")
```

---

## Technology Stack

Core Technologies:
- Python 3.8+ (main implementation)
- FastAPI/Flask (API documentation)
- Jinja2 (HTML templating)
- Markdown (content formatting)
- WeasyPrint (PDF generation)

Optional Integrations:
- AI services (content enhancement)
- Git hooks (auto-updates)
- CI/CD platforms (continuous deployment)
- Static site generators (hosting)
- Documentation platforms (confluence)

Output Formats:
- HTML (responsive sites)
- PDF (print-ready documents)
- Markdown (version control friendly)
- OpenAPI/Swagger (API specifications)
- Static sites (hosting platforms)

---

*For detailed implementation patterns and advanced configurations, see the `modules/` directory.*
