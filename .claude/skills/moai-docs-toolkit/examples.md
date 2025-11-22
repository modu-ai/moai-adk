# moai-docs-toolkit: Practical Examples

Unified documentation toolkit with generation, validation, and linting.

---

## Example 1: Complete Documentation Pipeline

End-to-end docs generation with validation:

```python
from moai_docs.toolkit import DocumentationPipeline

pipeline = DocumentationPipeline(
    source_dir="src/",
    output_dir="docs/"
)

# Generate, validate, and lint in one call
results = pipeline.execute(
    generate=True,
    validate=True,
    lint=True,
    report=True
)

print(f"Generated: {results.files_created}")
print(f"Errors: {results.errors}")
print(f"Quality Score: {results.quality_score}/100")
```

---

## Example 2: Documentation Generation

Auto-generate from source code:

```python
from moai_docs.generators import CodeDocGenerator

generator = CodeDocGenerator()

# Generate API docs from TypeScript
docs = generator.generate(
    source_dir="src/",
    language="typescript",
    format="markdown",
    include_examples=True
)

for doc in docs:
    doc.save(f"docs/{doc.module_name}.md")
```

---

## Example 3: Documentation Validation

Verify SPEC compliance and quality:

```python
from moai_docs.validators import SpecValidator

validator = SpecValidator()

# Validate against SPEC requirements
issues = validator.validate_against_spec(
    doc_dir="docs/",
    spec_file="SPEC-001.md"
)

print(f"Compliance: {validator.compliance_score}%")
for issue in issues:
    print(f"  - {issue}")
```

---

## Example 4: Markdown Linting

Check documentation format:

```python
from moai_docs.linters import MarkdownLinter

linter = MarkdownLinter()

# Lint all markdown files
results = linter.lint_directory("docs/")

for file_path, errors in results.items():
    if errors:
        print(f"{file_path}:")
        for error in errors:
            print(f"  - Line {error.line}: {error.message}")
```

---

## Example 5: Quality Report Generation

Generate comprehensive quality metrics:

```python
from moai_docs.reporting import QualityReporter

reporter = QualityReporter()

# Generate detailed quality report
report = reporter.generate_report("docs/")

print(f"Coverage: {report.coverage}%")
print(f"Readability: {report.readability}/100")
print(f"Code Examples: {report.code_examples}")
print(f"Broken Links: {report.broken_links}")

report.save_as_html("docs-quality-report.html")
```

---

## Example 6: CI/CD Integration

Automated docs validation in pipelines:

```yaml
# .github/workflows/docs-validation.yml
name: Documentation Quality

on:
  pull_request:
    paths:
      - 'src/**'
      - 'docs/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run Documentation Toolkit
        run: |
          python -m moai_docs.toolkit \
            --source src/ \
            --docs docs/ \
            --generate \
            --validate \
            --lint \
            --report
```

---

## Example 7: Multilingual Documentation

Generate docs in multiple languages:

```python
from moai_docs.toolkit import MultilingualDocToolkit

toolkit = MultilingualDocToolkit()

# Generate and validate for all languages
toolkit.process_multilingual(
    source_dir="src/",
    languages=["ko", "en", "ja"],
    validate=True,
    lint=True
)

# Verify consistency across languages
consistency = toolkit.check_consistency()
print(f"Language consistency: {consistency}%")
```

---

## Example 8: Documentation Versioning

Manage documentation versions:

```python
from moai_docs.toolkit import VersionedDocToolkit

toolkit = VersionedDocToolkit()

# Create version snapshot
toolkit.create_version("1.2.0", docs_dir="docs/")

# Generate version matrix
matrix = toolkit.generate_version_matrix(
    versions=["1.0.0", "1.1.0", "1.2.0"]
)

print(matrix)  # Shows compatibility across versions
```

---

## Example 9: Documentation Search

Build searchable documentation:

```python
from moai_docs.toolkit import SearchableDocToolkit

toolkit = SearchableDocToolkit()

# Build search index
toolkit.build_search_index("docs/")

# Search documentation
results = toolkit.search("authentication")

for result in results:
    print(f"{result.file}: {result.relevance}%")
```

---

## Example 10: Documentation Analytics

Analyze documentation usage:

```python
from moai_docs.toolkit import AnalyticsToolkit

toolkit = AnalyticsToolkit()

# Get documentation statistics
stats = toolkit.analyze(docs_dir="docs/")

print(f"Total documents: {stats.total_docs}")
print(f"Total lines: {stats.total_lines}")
print(f"Code examples: {stats.code_examples}")
print(f"Average readability: {stats.avg_readability}/100")
```

---

**Last Updated**: 2025-11-22
**Examples Count**: 10
**Status**: Production Ready
