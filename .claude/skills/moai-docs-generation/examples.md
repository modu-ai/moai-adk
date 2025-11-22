# moai-docs-generation: Practical Examples

Documentation generation with AI-powered features and Context7 integration.

---

## Example 1: Basic README Generation

Generate a README from package metadata:

```python
from moai_docs.generators import ReadmeGenerator

generator = ReadmeGenerator(
    project_name="my-project",
    description="A sample project",
    author="John Doe"
)

# Generate README with template
readme = generator.generate()
readme.save("README.md")
```

**Output**:
```markdown
# my-project

A sample project

---

## Installation

```bash
npm install my-project
```

---

## Usage

See [examples](./docs/examples) for more.
```

---

## Example 2: API Documentation from JSDoc

Auto-generate API docs from TypeScript/JavaScript:

```typescript
/**
 * Calculate the sum of two numbers
 * @param a First number
 * @param b Second number
 * @returns Sum of a and b
 * @example
 * const result = add(2, 3);  // Returns 5
 */
export function add(a: number, b: number): number {
    return a + b;
}
```

Uses JSDoc parser to generate:

```markdown
### Function: add

Calculate the sum of two numbers

**Parameters**:
- `a` (number): First number
- `b` (number): Second number

**Returns**: number - Sum of a and b

**Example**:
```typescript
const result = add(2, 3);  // Returns 5
```
```

---

## Example 3: Python Docstring to API Docs

Convert Python docstrings to Sphinx/Markdown:

```python
def calculate_mean(numbers: List[float]) -> float:
    """
    Calculate arithmetic mean of numbers.

    Args:
        numbers: List of numerical values

    Returns:
        Arithmetic mean of the values

    Raises:
        ValueError: If list is empty

    Example:
        >>> calculate_mean([1, 2, 3])
        2.0
    """
    if not numbers:
        raise ValueError("Cannot calculate mean of empty list")
    return sum(numbers) / len(numbers)
```

Generates:

```markdown
### Function: calculate_mean

Calculate arithmetic mean of numbers.

**Parameters**:
- `numbers` (List[float]): List of numerical values

**Returns**: float - Arithmetic mean of the values

**Raises**:
- ValueError: If list is empty

**Example**:
```python
>>> calculate_mean([1, 2, 3])
2.0
```
```

---

## Example 4: Scaffold Documentation Structure

Generate complete docs folder hierarchy:

```python
from moai_docs.scaffold import DocumentationScaffold

scaffold = DocumentationScaffold("my-project")

# Create multilingual docs structure
scaffold.create_project_structure()

# Result:
# docs/
# ├── ko/
# │   ├── guides/
# │   ├── api/
# │   ├── examples/
# │   └── reference/
# ├── en/
# │   ├── guides/
# │   ├── api/
# │   ├── examples/
# │   └── reference/
# ├── ja/
# └── zh/
```

---

## Example 5: Generate Guide Templates

Create documentation guides with templates:

```python
from moai_docs.templates import GuideTemplate

guide = GuideTemplate(
    name="Getting Started",
    language="en"
)

# Use predefined template
content = guide.render(
    title="Getting Started with My Project",
    prerequisites=["Node.js 18+", "npm or yarn"],
    steps=[
        "Install the package",
        "Configure your settings",
        "Run your first example"
    ]
)

guide.save("docs/en/guides/getting-started.md")
```

**Output**:
```markdown
# Getting Started with My Project

## Prerequisites
- Node.js 18+
- npm or yarn

## Step 1: Install the package
[Details here]

## Step 2: Configure your settings
[Details here]

## Step 3: Run your first example
[Details here]

## What's Next
- See [examples](../examples)
- Read [API reference](../api)
```

---

## Example 6: Batch Generate All Docs

Generate documentation for entire project:

```python
from moai_docs.batch import BatchDocGenerator

generator = BatchDocGenerator(
    source_dir="src/",
    output_dir="docs/en/api"
)

# Generate API docs for all modules
results = generator.generate_all(
    format="markdown",
    include_examples=True,
    create_index=True
)

print(f"Generated {len(results)} documents")
```

---

## Example 7: SPEC-based Documentation

Generate documentation from SPEC files:

```python
from moai_docs.spec import SpecDocGenerator

generator = SpecDocGenerator()

# Read SPEC and generate docs
spec = generator.read_spec("SPEC-001.md")
docs = generator.generate_from_spec(spec)

# Creates:
# - Overview page
# - Implementation guide
# - API reference
# - Examples
```

---

## Example 8: Multilingual Documentation

Generate docs in multiple languages:

```python
from moai_docs.i18n import MultilingualDocGenerator

generator = MultilingualDocGenerator()

# Generate English and Korean
generator.generate_multilingual(
    source_file="docs/en/guide.md",
    languages=["ko", "ja", "zh"],
    translator="ai"  # Uses Context7 for translation
)

# Creates:
# docs/
# ├── en/guide.md
# ├── ko/guide.md
# ├── ja/guide.md
# └── zh/guide.md
```

---

## Example 9: Code Example Generation

Generate executable code examples:

```python
from moai_docs.examples import ExampleGenerator

generator = ExampleGenerator(
    language="python",
    framework="FastAPI"
)

# Generate example with Context7 best practices
example = generator.generate(
    feature="user-authentication",
    complexity="intermediate"
)

example.validate()  # Check syntax
example.save("docs/examples/user-auth.py")
```

**Output**:
```python
from fastapi import FastAPI, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthCredentials

app = FastAPI()
security = HTTPBearer()

@app.post("/login")
async def login(username: str, password: str):
    # Validate credentials
    if not validate_credentials(username, password):
        raise HTTPException(status_code=401)

    # Generate token
    token = generate_jwt_token(username)
    return {"access_token": token}
```

---

## Example 10: Automated Changelog

Generate changelog from git commits:

```python
from moai_docs.changelog import ChangelogGenerator

generator = ChangelogGenerator(
    version="1.2.0",
    since_tag="v1.1.0"
)

# Parse commits and generate changelog
changelog = generator.generate()

changelog.save("CHANGELOG.md")
```

**Output**:
```markdown
# Changelog

## [1.2.0] - 2025-11-22

### Added
- New user authentication module
- Support for OAuth2 integration
- Enhanced error handling

### Fixed
- Fixed memory leak in connection pool
- Resolved race condition in cache

### Changed
- Improved API response times by 30%
- Updated dependencies

### Deprecated
- Legacy JWT authentication (use OAuth2)

### Removed
- Old authentication module
```

---

## Example 11: Documentation Quality Report

Generate quality metrics for docs:

```python
from moai_docs.quality import QualityAnalyzer

analyzer = QualityAnalyzer(docs_dir="docs/")

# Analyze documentation quality
report = analyzer.analyze()

print(f"Coverage: {report.coverage}%")
print(f"Code examples: {report.code_examples}")
print(f"Average readability: {report.readability_score}/100")

report.save("docs-quality-report.md")
```

---

## Example 12: Template Customization

Customize documentation templates:

```python
from moai_docs.templates import TemplateCustomizer

customizer = TemplateCustomizer()

# Define custom template
customizer.register(
    name="my-custom-api",
    template="""
# {title}

## Overview
{description}

## API Reference
{api_reference}

## Examples
{examples}

## FAQ
{faq}
"""
)

# Use custom template
docs = customizer.generate_with_template(
    template_name="my-custom-api",
    title="My API",
    description="Comprehensive API documentation"
)
```

---

## Example 13: Integration with CI/CD

Automate docs generation in GitHub Actions:

```yaml
# .github/workflows/generate-docs.yml
name: Generate Documentation

on:
  push:
    branches: [main, develop]
    paths: ['src/**', 'docs/**']

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Generate Documentation
        run: |
          python -m moai_docs.cli generate \
            --source src/ \
            --output docs/ \
            --format markdown \
            --create-index

      - name: Upload Artifacts
        uses: actions/upload-artifact@v3
        with:
          name: generated-docs
          path: docs/
```

---

## Example 14: Documentation Versioning

Manage documentation versions:

```python
from moai_docs.versioning import VersionManager

manager = VersionManager()

# Create version snapshot
manager.create_version("1.2.0", docs_dir="docs/")

# List all versions
versions = manager.list_versions()
print(f"Available versions: {versions}")

# Switch to specific version
manager.checkout_version("1.1.0")
```

---

## Example 15: Sync with GitHub Pages

Automatically deploy docs to GitHub Pages:

```python
from moai_docs.deployment import GitHubPagesDeployer

deployer = GitHubPagesDeployer(
    repo="username/repo",
    github_token="your-token"
)

# Generate and deploy
deployer.generate_and_deploy(
    source_dir="docs/",
    target_branch="gh-pages"
)
```

---

**Last Updated**: 2025-11-22
**Examples Count**: 15
**Code Blocks**: 25+
**Status**: Production Ready
