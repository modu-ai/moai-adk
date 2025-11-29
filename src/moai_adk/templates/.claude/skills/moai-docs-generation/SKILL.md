---
name: moai-docs-generation
description: Automated documentation generation specialist for technical specs, API docs, user guides, and knowledge bases with multi-format output
version: 1.0.0
category: workflow
tags:
  - documentation
  - generation
  - automation
  - markdown
  - html
  - api-docs
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# Documentation Generation Specialist

## Quick Reference (30 seconds)

**Automated Documentation Generation** - Comprehensive documentation automation covering technical specs, API documentation, user guides, and knowledge base creation with multi-format output capabilities.

**Core Capabilities**:
- 📚 **Technical Documentation**: API docs, architecture specs, code documentation
- 📖 **User Guides**: Tutorials, getting started guides, best practices
- 🔧 **API Documentation**: OpenAPI/Swagger generation, endpoint documentation
- 🌐 **Multi-Format Output**: Markdown, HTML, PDF, static sites
- 🤖 **AI-Powered Generation**: Context-aware content creation and enhancement
- 🔄 **Continuous Updates**: Auto-sync documentation with code changes

**When to Use**:
- Generating API documentation from code
- Creating technical specifications and architecture docs
- Building user guides and tutorials
- Automating knowledge base creation
- Maintaining up-to-date project documentation

---

## Implementation Guide

### API Documentation Generation

**OpenAPI/Swagger from Code**:
```python
from typing import Dict, List, Optional
import inspect
from fastapi import FastAPI

class APIDocGenerator:
    def __init__(self, app: FastAPI):
        self.app = app
        self.openapi_schema = app.openapi()

    def generate_openapi_spec(self) -> dict:
        """Generate complete OpenAPI specification."""
        return {
            "openapi": "3.1.0",
            "info": {
                "title": self.openapi_schema["info"]["title"],
                "version": self.openapi_schema["info"]["version"],
                "description": self.openapi_schema["info"].get("description", ""),
                "contact": {
                    "name": "API Support",
                    "email": "support@example.com"
                }
            },
            "paths": self.enrich_paths_with_examples(),
            "components": self.generate_components()
        }

    def enrich_paths_with_examples(self) -> dict:
        """Add examples and detailed descriptions to API paths."""
        enriched_paths = {}

        for path, path_item in self.openapi_schema["paths"].items():
            enriched_paths[path] = {}

            for method, operation in path_item.items():
                if method.upper() in ["GET", "POST", "PUT", "DELETE"]:
                    enriched_paths[path][method] = {
                        **operation,
                        "responses": self.enrich_responses(operation.get("responses", {})),
                        "examples": self.generate_examples(path, method, operation)
                    }

        return enriched_paths

    def generate_examples(self, path: str, method: str, operation: dict) -> dict:
        """Generate request/response examples."""
        examples = {}

        if method.upper() == "POST":
            examples["application/json"] = {
                "summary": "Example request",
                "value": self.generate_request_example(operation)
            }

        return examples

    def generate_markdown_docs(self) -> str:
        """Generate comprehensive Markdown documentation."""
        template = """
# {title}

{description}

## Base URL
```
{base_url}
```

## Authentication
{authentication}

## Endpoints

{endpoints}

## Data Models

{models}
        """

        endpoints = self.generate_endpoint_markdown()
        models = self.generate_model_markdown()

        return template.format(
            title=self.openapi_schema["info"]["title"],
            description=self.openapi_schema["info"].get("description", ""),
            base_url="https://api.example.com/v1",
            authentication=self.generate_auth_docs(),
            endpoints=endpoints,
            models=models
        )

# Usage example
def generate_api_docs():
    app = FastAPI(title="User Management API", version="1.0.0")

    doc_generator = APIDocGenerator(app)

    # Generate different formats
    openapi_spec = doc_generator.generate_openapi_spec()
    markdown_docs = doc_generator.generate_markdown_docs()
    html_docs = doc_generator.generate_html_docs()

    return {
        "openapi": openapi_spec,
        "markdown": markdown_docs,
        "html": html_docs
    }
```

### Code Documentation Enhancement

**AI-Powered Code Documentation**:
```python
import ast
import inspect
from typing import Dict, List, Any

class CodeDocGenerator:
    def __init__(self, ai_client=None):
        self.ai_client = ai_client

    def analyze_python_file(self, file_path: str) -> Dict[str, Any]:
        """Analyze Python file and extract documentation structure."""
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        tree = ast.parse(content)

        documentation = {
            "file": file_path,
            "modules": [],
            "classes": [],
            "functions": [],
            "imports": []
        }

        for node in ast.walk(tree):
            if isinstance(node, ast.ClassDef):
                class_info = self.analyze_class(node)
                documentation["classes"].append(class_info)

            elif isinstance(node, ast.FunctionDef):
                if not any(isinstance(parent, ast.ClassDef) for parent in ast.walk(tree)):
                    func_info = self.analyze_function(node)
                    documentation["functions"].append(func_info)

            elif isinstance(node, ast.Import):
                for alias in node.names:
                    documentation["imports"].append(alias.name)

        return documentation

    def analyze_class(self, class_node: ast.ClassDef) -> Dict[str, Any]:
        """Analyze class structure and generate documentation."""
        methods = []
        properties = []

        for node in class_node.body:
            if isinstance(node, ast.FunctionDef):
                method_info = self.analyze_function(node)
                methods.append(method_info)

            elif isinstance(node, ast.AnnAssign) and isinstance(node.target, ast.Name):
                prop_info = self.analyze_property(node)
                properties.append(prop_info)

        # Generate enhanced docstring with AI
        enhanced_docstring = None
        if self.ai_client:
            enhanced_docstring = self.enhance_docstring_with_ai(
                class_node.name,
                methods,
                properties,
                ast.get_docstring(class_node)
            )

        return {
            "name": class_node.name,
            "docstring": ast.get_docstring(class_node),
            "enhanced_docstring": enhanced_docstring,
            "methods": methods,
            "properties": properties,
            "inheritance": [base.id for base in class_node.bases if isinstance(base, ast.Name)],
            "decorators": [self.get_decorator_name(dec) for dec in class_node.decorator_list]
        }

    def enhance_docstring_with_ai(self, name: str, methods: List, properties: List, existing_doc: str = None) -> str:
        """Enhance documentation using AI analysis."""
        method_signatures = [method["signature"] for method in methods]

        prompt = f"""
        Enhance this Python class documentation:

        Class Name: {name}
        Existing Documentation: {existing_doc or 'No documentation'}

        Methods:
        {chr(10).join(method_signatures)}

        Properties:
        {', '.join([prop['name'] for prop in properties])}

        Please provide:
        1. Clear class description
        2. Usage examples
        3. Method descriptions
        4. Implementation notes
        5. Best practices for using this class

        Format as proper docstring with sections.
        """

        response = self.ai_client.generate_content(prompt)
        return response["content"]

    def generate_documentation_files(self, documentation: Dict, output_dir: str):
        """Generate various documentation files from analysis."""
        import os
        from pathlib import Path

        output_path = Path(output_dir)
        output_path.mkdir(exist_ok=True)

        # Generate README
        readme_content = self.generate_readme(documentation)
        (output_path / "README.md").write_text(readme_content, encoding='utf-8')

        # Generate API reference
        api_content = self.generate_api_reference(documentation)
        (output_path / "API.md").write_text(api_content, encoding='utf-8')

        # Generate examples
        examples_content = self.generate_examples(documentation)
        (output_path / "EXAMPLES.md").write_text(examples_content, encoding='utf-8')
```

### User Guide Generation

**Tutorial and Guide Creation**:
```python
class UserGuideGenerator:
    def __init__(self, ai_client=None):
        self.ai_client = ai_client

    def generate_getting_started_guide(self, project_info: Dict) -> str:
        """Generate comprehensive getting started guide."""

        guide_structure = [
            "# Getting Started",
            "",
            "## Prerequisites",
            self.generate_prerequisites_section(project_info),
            "",
            "## Installation",
            self.generate_installation_section(project_info),
            "",
            "## Quick Start",
            self.generate_quick_start_section(project_info),
            "",
            "## Basic Usage",
            self.generate_basic_usage_section(project_info),
            "",
            "## Next Steps",
            self.generate_next_steps_section(project_info)
        ]

        return "\n".join(guide_structure)

    def generate_tutorial_series(self, features: List[Dict]) -> List[str]:
        """Generate a series of tutorials for different features."""
        tutorials = []

        for feature in features:
            tutorial = self.generate_feature_tutorial(feature)
            tutorials.append(tutorial)

        return tutorials

    def generate_feature_tutorial(self, feature: Dict) -> str:
        """Generate a single tutorial for a specific feature."""

        if self.ai_client:
            prompt = f"""
            Create a step-by-step tutorial for this feature:

            Feature Name: {feature['name']}
            Description: {feature['description']}
            Key Functions: {', '.join(feature.get('functions', []))}
            Example Usage: {feature.get('example', '')}

            Please include:
            1. Clear introduction explaining what the feature does
            2. Prerequisites and setup requirements
            3. Step-by-step implementation guide
            4. Complete code example
            5. Common use cases and variations
            6. Troubleshooting tips
            7. Related features and next steps

            Format as a comprehensive tutorial with clear sections and code blocks.
            """

            response = self.ai_client.generate_content(prompt)
            return response["content"]

        else:
            return self.generate_basic_tutorial(feature)

    def generate_cookbook(self, use_cases: List[Dict]) -> str:
        """Generate a cookbook of common patterns and solutions."""

        cookbook_content = ["# Cookbook", "", "## Common Patterns and Solutions"]

        for use_case in use_cases:
            recipe = self.generate_recipe(use_case)
            cookbook_content.extend(["", recipe])

        return "\n".join(cookbook_content)

    def generate_recipe(self, use_case: Dict) -> str:
        """Generate a single recipe for the cookbook."""

        return f"""
### {use_case['title']}

**Problem**: {use_case['problem']}

**Solution**: {use_case['solution']}

**Code Example**:
```python
{use_case['code_example']}
```

**Explanation**: {use_case['explanation']}

**Variations**: {use_case.get('variations', 'None')}

**See Also**: {use_case.get('related_features', '')}
        """
```

### Multi-Format Output Generation

**HTML Documentation Site**:
```python
from jinja2 import Template
import markdown

class HTMLDocGenerator:
    def __init__(self):
        self.templates = self.load_templates()

    def generate_html_site(self, documentation: Dict, output_dir: str):
        """Generate a complete HTML documentation site."""

        # Generate index page
        index_html = self.generate_index_page(documentation)
        self.write_file(output_dir, "index.html", index_html)

        # Generate API reference pages
        self.generate_api_pages(documentation, output_dir)

        # Generate tutorial pages
        self.generate_tutorial_pages(documentation, output_dir)

        # Generate CSS and assets
        self.generate_assets(output_dir)

    def generate_index_page(self, documentation: Dict) -> str:
        """Generate the main index page."""

        template = Template("""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ title }}</title>
    <link rel="stylesheet" href="assets/style.css">
</head>
<body>
    <nav class="sidebar">
        <div class="logo">
            <h2>{{ title }}</h2>
        </div>
        <ul class="nav-menu">
            <li><a href="index.html" class="active">Home</a></li>
            <li><a href="api/index.html">API Reference</a></li>
            <li><a href="tutorials/index.html">Tutorials</a></li>
            <li><a href="cookbook.html">Cookbook</a></li>
        </ul>
    </nav>

    <main class="content">
        <section class="hero">
            <h1>{{ title }}</h1>
            <p>{{ description }}</p>
            <div class="cta-buttons">
                <a href="tutorials/getting-started.html" class="btn primary">Get Started</a>
                <a href="api/index.html" class="btn secondary">API Docs</a>
            </div>
        </section>

        <section class="features">
            <h2>Key Features</h2>
            <div class="feature-grid">
                {% for feature in features %}
                <div class="feature-card">
                    <h3>{{ feature.name }}</h3>
                    <p>{{ feature.description }}</p>
                </div>
                {% endfor %}
            </div>
        </section>

        <section class="quick-start">
            <h2>Quick Start</h2>
            <pre><code>{{ quick_start_example }}</code></pre>
        </section>
    </main>

    <script src="assets/script.js"></script>
</body>
</html>
        """)

        return template.render(
            title=documentation.get("title", "API Documentation"),
            description=documentation.get("description", ""),
            features=documentation.get("features", []),
            quick_start_example=documentation.get("quick_start", "")
        )

    def generate_pdf_documentation(self, markdown_content: str, output_path: str):
        """Generate PDF from Markdown content."""
        try:
            import weasyprint

            # Convert Markdown to HTML
            html_content = markdown.markdown(
                markdown_content,
                extensions=['tables', 'fenced_code', 'toc']
            )

            # Add CSS styling
            styled_html = f"""
            <!DOCTYPE html>
            <html>
            <head>
                <style>
                    body {{ font-family: Arial, sans-serif; line-height: 1.6; margin: 40px; }}
                    h1 {{ color: #333; border-bottom: 2px solid #333; }}
                    h2 {{ color: #666; }}
                    code {{ background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }}
                    pre {{ background: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }}
                    table {{ border-collapse: collapse; width: 100%; }}
                    th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
                </style>
            </head>
            <body>
                {html_content}
            </body>
            </html>
            """

            # Generate PDF
            weasyprint.HTML(string=styled_html).write_pdf(output_path)

        except ImportError:
            raise ImportError("weasyprint is required for PDF generation. Install with: pip install weasyprint")
```

---

## Advanced Patterns

### Continuous Documentation Updates

**Git Hooks for Auto-Documentation**:
```python
class DocumentationUpdater:
    def __init__(self, repo_path: str):
        self.repo_path = repo_path
        self.doc_generator = CodeDocGenerator()

    def setup_pre_commit_hook(self):
        """Setup git pre-commit hook for documentation updates."""
        hook_content = '''#!/bin/bash
# Auto-update documentation before commits

python -c "
import sys
sys.path.append('scripts')
from doc_updater import DocumentationUpdater
updater = DocumentationUpdater('.')
updater.update_documentation()
"

# Add updated documentation to commit
git add docs/ README.md API.md
        '''

        hook_path = Path(self.repo_path) / ".git" / "hooks" / "pre-commit"
        hook_path.write_text(hook_content)
        hook_path.chmod(0o755)

    def update_documentation(self):
        """Update documentation based on code changes."""

        # Analyze changed Python files
        changed_files = self.get_changed_python_files()

        for file_path in changed_files:
            documentation = self.doc_generator.analyze_python_file(file_path)

            # Update relevant documentation files
            self.update_api_docs(documentation)
            self.update_examples(documentation)

    def get_changed_python_files(self) -> List[str]:
        """Get list of changed Python files in current commit."""
        import subprocess

        result = subprocess.run(
            ["git", "diff", "--cached", "--name-only", "--diff-filter=ACM", "*.py"],
            cwd=self.repo_path,
            capture_output=True,
            text=True
        )

        return [line.strip() for line in result.stdout.splitlines() if line.strip()]
```

### Documentation Quality Validation

**Automated Documentation Quality Checks**:
```python
class DocumentationValidator:
    def __init__(self):
        self.quality_metrics = {
            'min_docstring_coverage': 0.8,
            'min_example_coverage': 0.6,
            'required_sections': ['description', 'parameters', 'returns', 'examples']
        }

    def validate_documentation(self, documentation: Dict) -> Dict:
        """Validate documentation quality and generate report."""

        validation_results = {
            'overall_score': 0,
            'issues': [],
            'recommendations': [],
            'metrics': {}
        }

        # Check docstring coverage
        docstring_coverage = self.calculate_docstring_coverage(documentation)
        validation_results['metrics']['docstring_coverage'] = docstring_coverage

        if docstring_coverage < self.quality_metrics['min_docstring_coverage']:
            validation_results['issues'].append(
                f"Low docstring coverage: {docstring_coverage:.1%} (target: {self.quality_metrics['min_docstring_coverage']:.1%})"
            )

        # Check example coverage
        example_coverage = self.calculate_example_coverage(documentation)
        validation_results['metrics']['example_coverage'] = example_coverage

        if example_coverage < self.quality_metrics['min_example_coverage']:
            validation_results['issues'].append(
                f"Low example coverage: {example_coverage:.1%} (target: {self.quality_metrics['min_example_coverage']:.1%})"
            )

        # Generate recommendations
        validation_results['recommendations'] = self.generate_recommendations(validation_results)

        # Calculate overall score
        validation_results['overall_score'] = self.calculate_overall_score(validation_results['metrics'])

        return validation_results

    def calculate_docstring_coverage(self, documentation: Dict) -> float:
        """Calculate percentage of documented items."""
        total_items = 0
        documented_items = 0

        for class_info in documentation.get('classes', []):
            total_items += 1
            if class_info.get('docstring'):
                documented_items += 1

            for method in class_info.get('methods', []):
                total_items += 1
                if method.get('docstring'):
                    documented_items += 1

        for function in documentation.get('functions', []):
            total_items += 1
            if function.get('docstring'):
                documented_items += 1

        return documented_items / total_items if total_items > 0 else 0
```

---

## Works Well With

- **moai-domain-backend** - Backend API documentation
- **moai-domain-frontend** - Component documentation and guides
- **moai-workflow-project** - Project documentation management
- **moai-integration-mcp** - MCP server documentation
- **moai-foundation-core** - Core architectural documentation

---

## Usage Examples

### Command Line Interface
```bash
# Generate API documentation
moai-docs generate-api --app-dir ./src --output-dir ./docs/api

# Analyze code and generate documentation
moai-docs analyze-code --src-dir ./src --output-dir ./docs/code

# Generate user guides
moai-docs generate-guides --project-config ./config/project.json

# Create HTML documentation site
moai-docs build-site --source-dir ./docs --output-dir ./public

# Validate documentation quality
moai-docs validate --docs-dir ./docs --threshold 0.8
```

### Python API
```python
from moai_docs_generation import DocumentationGenerator

# Initialize generator
generator = DocumentationGenerator()

# Generate API docs
api_docs = generator.generate_api_documentation(app)

# Generate user guides
guides = generator.generate_user_guides(project_info)

# Build complete documentation site
generator.build_documentation_site(
    source_dir="./docs",
    output_dir="./public",
    formats=["html", "pdf"]
)
```

---

## Technology Stack

**Core Libraries**:
- **ast**: Python code analysis and parsing
- **jinja2**: Template engine for HTML generation
- **markdown**: Markdown to HTML conversion
- **weasyprint**: PDF generation from HTML
- **pydoc**: Python documentation extraction

**Documentation Formats**:
- Markdown (GitHub, GitLab compatible)
- HTML with responsive design
- PDF for printable documentation
- OpenAPI/Swagger for API specs
- JSON for structured data

**Quality Tools**:
- PyLint for docstring validation
- Type hints for better documentation
- Sphinx for advanced documentation generation
- MkDocs for static site generation

---

**Status**: Production Ready
**Last Updated**: 2025-11-30
**Maintained by**: MoAI-ADK Documentation Team