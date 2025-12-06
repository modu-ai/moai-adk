# Documentation Generation Examples

Multishot prompting examples demonstrating practical usage patterns for automated documentation generation.

---

## Example 1: Basic API Documentation Generation

**Scenario**: Generate OpenAPI documentation from a FastAPI application.

**Input**:
```python
from moai_docs_generation import DocumentationGenerator

# Initialize generator
doc_gen = DocumentationGenerator()

# Generate API documentation from FastAPI app
api_docs = doc_gen.generate_api_docs(
    source_file="src/app.py",
    output_format="openapi",
    include_examples=True
)
```

**Output**:
```yaml
openapi: 3.1.0
info:
  title: User Management API
  version: 1.0.0
  description: API for managing user accounts and authentication

paths:
  /users:
    get:
      summary: List all users
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
              example:
                - id: 1
                  name: "John Doe"
                  email: "john@example.com"
```

**Explanation**: The generator analyzes FastAPI route decorators, Pydantic models, and docstrings to automatically produce OpenAPI 3.1 compliant documentation with examples.

---

## Example 2: User Guide Generation from SPEC

**Scenario**: Create a comprehensive user guide from a SPEC document.

**Input**:
```python
from moai_docs_generation import DocumentationGenerator

doc_gen = DocumentationGenerator()

# SPEC data for feature
spec_data = {
    "id": "SPEC-001",
    "title": "User Authentication System",
    "description": "Secure user authentication with JWT tokens",
    "requirements": [
        "User registration with email verification",
        "JWT token generation and refresh",
        "Password reset via email"
    ],
    "api_endpoints": [
        {"path": "/auth/login", "method": "POST"},
        {"path": "/auth/register", "method": "POST"},
        {"path": "/auth/refresh", "method": "POST"}
    ]
}

# Generate user guide
user_guide = doc_gen.generate_user_guide(
    spec_data=spec_data,
    template="getting-started",
    language="en"
)
```

**Output**:
```markdown
# User Authentication System Guide

## Getting Started

This guide walks you through the user authentication system,
including registration, login, and password management.

### Prerequisites
- API access credentials
- Valid email address

### Step 1: User Registration

To create a new account, send a POST request to `/auth/register`:

POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123",
  "name": "John Doe"
}

### Step 2: Email Verification
Check your inbox for a verification email and click the link.

### Step 3: Login
After verification, authenticate using `/auth/login`:

POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123"
}

Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600
}
```

**Explanation**: The generator transforms SPEC requirements into step-by-step user guides with practical examples and clear instructions.

---

## Example 3: Multi-Format Documentation Export

**Scenario**: Generate documentation in multiple formats for different audiences.

**Input**:
```python
from moai_docs_generation import DocumentationGenerator

doc_gen = DocumentationGenerator()

# Generate comprehensive documentation
docs = doc_gen.generate_comprehensive_docs(
    source_directory="./src",
    include_api_docs=True,
    include_user_guides=True,
    output_formats=["markdown", "html", "pdf"]
)

# Export to all formats
export_results = doc_gen.export_to_formats(
    documentation=docs,
    output_dir="./docs/output",
    formats=["markdown", "html", "pdf"],
    options={
        "html": {"theme": "modern", "responsive": True},
        "pdf": {"page_size": "A4", "include_toc": True}
    }
)
```

**Output**:
```python
{
    "status": "success",
    "exports": {
        "markdown": {
            "path": "./docs/output/markdown/",
            "files": ["api-reference.md", "user-guide.md", "architecture.md"],
            "size_kb": 145
        },
        "html": {
            "path": "./docs/output/html/",
            "files": ["index.html", "api-reference.html", "user-guide.html"],
            "size_kb": 892,
            "features": ["responsive", "search", "dark-mode"]
        },
        "pdf": {
            "path": "./docs/output/pdf/complete-documentation.pdf",
            "pages": 47,
            "size_kb": 2340,
            "includes_toc": True
        }
    }
}
```

**Explanation**: The generator creates documentation in multiple formats simultaneously, with format-specific optimizations like responsive HTML and PDF table of contents.

---

## Common Patterns

### Pattern 1: Continuous Documentation Integration

Automatically update documentation when code changes:

```python
# In pre-commit hook
from moai_docs_generation import DocumentationGenerator

def pre_commit_docs_update():
    doc_gen = DocumentationGenerator()

    # Get changed files
    changed_files = get_git_staged_files()

    # Update documentation for changed files only
    doc_gen.update_documentation_for_files(
        files=changed_files,
        validate_links=True,
        update_timestamps=True
    )

    # Stage updated documentation
    stage_updated_docs()
```

### Pattern 2: AI-Enhanced Documentation

Use AI to enhance documentation quality:

```python
from moai_docs_generation import DocumentationGenerator

doc_gen = DocumentationGenerator()

# Generate with AI enhancement
enhanced_docs = doc_gen.generate_with_ai(
    source_code="./src/",
    enhancement_level="comprehensive",
    options={
        "include_examples": True,
        "include_troubleshooting": True,
        "include_best_practices": True,
        "generate_diagrams": True
    }
)
```

### Pattern 3: Documentation Quality Validation

Validate documentation completeness and accuracy:

```python
from moai_docs_generation import DocumentationGenerator

doc_gen = DocumentationGenerator()

# Validate documentation
quality_report = doc_gen.validate_documentation(
    docs_path="./docs",
    completeness_threshold=0.9,
    include_example_validation=True,
    check_link_integrity=True,
    check_code_samples=True
)

# Report structure
# {
#     "completeness_score": 0.92,
#     "broken_links": [],
#     "missing_sections": ["troubleshooting"],
#     "invalid_examples": [],
#     "recommendations": ["Add error handling examples"]
# }
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Monolithic Documentation Generation

**Problem**: Generating all documentation in a single call without structure.

```python
# Incorrect approach
doc_gen.generate_all(source="./", output="./docs/all.md")
```

**Solution**: Use structured generation with clear separation of concerns.

```python
# Correct approach
api_docs = doc_gen.generate_api_docs(source="./src/api")
user_docs = doc_gen.generate_user_guide(source="./src/features")
arch_docs = doc_gen.generate_architecture_docs(source="./src")

doc_gen.compile_documentation(
    sections=[api_docs, user_docs, arch_docs],
    output_structure="hierarchical"
)
```

### Anti-Pattern 2: Ignoring Documentation Validation

**Problem**: Publishing documentation without validation leads to broken links and outdated examples.

```python
# Incorrect approach
doc_gen.generate_and_publish(source="./src", deploy=True)
```

**Solution**: Always validate before publishing.

```python
# Correct approach
docs = doc_gen.generate_documentation(source="./src")

validation = doc_gen.validate_documentation(docs)
if validation["completeness_score"] >= 0.9 and not validation["broken_links"]:
    doc_gen.publish(docs, platform="github-pages")
else:
    raise DocumentationQualityError(validation["issues"])
```

### Anti-Pattern 3: Manual Format Conversion

**Problem**: Manually converting documentation between formats introduces inconsistencies.

```python
# Incorrect approach
markdown_docs = doc_gen.generate_markdown(source="./src")
# Then manually converting to HTML using external tools
```

**Solution**: Use integrated multi-format export.

```python
# Correct approach
doc_gen.export_to_formats(
    documentation=docs,
    formats=["markdown", "html", "pdf"],
    maintain_consistency=True
)
```

---

## Integration Examples

### CI/CD Pipeline Integration

```yaml
# .github/workflows/docs.yml
name: Documentation Generation

on:
  push:
    branches: [main]
    paths:
      - 'src/**'
      - 'docs/**'

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate Documentation
        run: |
          moai docs:generate --source ./src --output ./docs
          moai docs:validate --path ./docs --threshold 0.9

      - name: Deploy to GitHub Pages
        if: success()
        run: |
          moai docs:deploy --platform github-pages --branch gh-pages
```

### Slack Notification on Documentation Updates

```python
from moai_docs_generation import DocumentationGenerator, notify_slack

doc_gen = DocumentationGenerator()

# Generate and validate
docs = doc_gen.generate_documentation(source="./src")
validation = doc_gen.validate_documentation(docs)

# Notify team
notify_slack(
    channel="#documentation",
    message=f"Documentation updated: {validation['completeness_score']*100:.1f}% complete",
    details=validation
)
```

---

*For additional patterns and advanced configurations, see the `modules/` directory.*
