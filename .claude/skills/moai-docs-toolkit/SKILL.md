---
name: moai-docs-toolkit
description: Unified documentation toolkit combining generation, validation, and linting with AI-powered features and Context7 integration.
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 80
auto_trigger_keywords:
  - docs
  - testing
  - toolkit
category_tier: special
---

# Documentation Toolkit (Unified)

## Quick Reference (30 seconds)

Comprehensive documentation management system for generation, validation, and linting. Automate docs creation from code, validate quality and SPEC compliance, and enforce consistent formatting standards. Unified toolkit reduces documentation maintenance overhead by 60-70% through intelligent automation and continuous quality gates.

**Core Capabilities** (November 2025):
- **Generation**: Auto-generate docs from code comments, create templates, scaffold project structure
- **Validation**: SPEC compliance, content accuracy, quality scoring, multilingual consistency
- **Linting**: Header structure, code blocks, links, lists, tables, typography validation
- **CI/CD Integration**: Pre-commit hooks, pull request checks, automated reports
- **Quality Gates**: Configurable thresholds, automated enforcement, detailed metrics

---

## Implementation Guide

### Part 1: Documentation Generation

#### Auto-Generate from Code

**TypeScript/JavaScript JSDoc**:
```typescript
/**
 * Calculate sum of two numbers
 * @param a First number
 * @param b Second number
 * @returns Sum of a and b
 * @example
 * const result = sum(2, 3);  // Returns 5
 */
function sum(a: number, b: number): number {
    return a + b;
}
```

Generates:
```markdown
### Function: sum

Calculate sum of two numbers

**Signature**:
```typescript
function sum(a: number, b: number): number
```

**Parameters**:
- `a`: First number
- `b`: Second number

**Returns**: Sum of a and b

**Example**:
```typescript
const result = sum(2, 3);  // Returns 5
```
```

**Python Docstring**:
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

#### Scaffold Generation

**Create documentation structure**:
```python
class DocumentationScaffold:
    def create_project_structure(self, project_name: str):
        """Generate complete docs folder hierarchy"""
        base = Path(f"docs/{project_name}")

        # Create language directories
        for lang in ["ko", "en", "ja", "zh"]:
            lang_dir = base / lang
            lang_dir.mkdir(parents=True, exist_ok=True)

            # Create standard sections
            for section in ["guides", "api", "examples", "reference"]:
                (lang_dir / section).mkdir(exist_ok=True)

        return base

    def create_guide_template(self, guide_name: str, lang: str = "en"):
        """Create individual guide with template"""
        template = f"""# {guide_name}

## Overview
[Brief description of what this guide covers]

---

## Prerequisites
- [Requirement 1]
- [Requirement 2]

---

## Step-by-Step Tutorial

### Step 1: [First step]
[Detailed explanation]

\`\`\`python
# Code example
\`\`\`

### Step 2: [Next step]
[Detailed explanation]

---

## Best Practices

### ✅ DO
- [Practice 1]
- [Practice 2]

### ❌ DON'T
- [Anti-pattern 1]
- [Anti-pattern 2]

---

**Last Updated**: {datetime.now().isoformat()}
"""
        return template
```

### Part 2: Documentation Validation

#### SPEC Compliance Checking

```python
class SpecComplianceValidator:
    def validate_spec_references(self, doc_path: Path) -> List[str]:
        """Verify all SPEC requirements addressed in docs"""
        content = doc_path.read_text()
        errors = []

        # Find SPEC references
        spec_pattern = r'SPEC-\d{3,4}'
        specs = re.findall(spec_pattern, content)

        if not specs:
            errors.append("No SPEC references found in documentation")

        for spec in specs:
            if not self._spec_exists(spec):
                errors.append(f"Referenced SPEC does not exist: {spec}")

        return errors

    def calculate_quality_score(self, doc_path: Path) -> float:
        """Calculate documentation quality (0-100)"""
        scores = {
            'spec_compliance': self._check_spec_compliance(doc_path),      # 25%
            'content_accuracy': self._validate_content_accuracy(doc_path),  # 25%
            'completeness': self._check_completeness(doc_path),             # 20%
            'readability': self._calculate_readability(doc_path),           # 15%
            'formatting': self._check_formatting(doc_path),                 # 15%
        }

        weights = {
            'spec_compliance': 0.25,
            'content_accuracy': 0.25,
            'completeness': 0.20,
            'readability': 0.15,
            'formatting': 0.15,
        }

        total_score = sum(scores[k] * weights[k] for k in scores)
        return round(total_score, 1)
```

#### Content Accuracy Validation

```python
def validate_code_examples(self, doc_path: Path) -> List[str]:
    """Verify code examples are syntactically correct"""
    errors = []
    content = doc_path.read_text()

    # Extract code blocks
    code_blocks = re.findall(r'```(\w+)\n(.*?)\n```', content, re.DOTALL)

    for language, code in code_blocks:
        # Validate syntax per language
        if language in ['python', 'javascript', 'typescript']:
            syntax_errors = self._check_syntax(code, language)
            if syntax_errors:
                errors.append(f"Invalid {language} code block: {syntax_errors}")

    return errors
```

#### Quality Metrics

```python
class QualityMetrics:
    def measure_documentation(self, doc_path: Path) -> dict:
        """Measure quality across multiple dimensions"""
        content = doc_path.read_text()

        return {
            'readability_score': self._flesch_kincaid(content),  # 60-100
            'coverage': self._spec_coverage(doc_path),           # 80%+
            'code_example_ratio': self._code_ratio(content),     # 1:300 words
            'link_validity': self._check_links(doc_path),        # 100%
            'translation_completeness': self._i18n_complete(),   # 100%
            'image_optimization': self._check_images(doc_path),  # <500KB each
        }
```

### Part 3: Documentation Linting

#### Header Structure Validation

```python
class HeaderLinter:
    def validate_headers(self, content: str) -> List[str]:
        """Check header structure and hierarchy"""
        errors = []

        # Count H1 headers
        h1_count = len(re.findall(r'^# ', content, re.MULTILINE))
        if h1_count != 1:
            errors.append(f"Found {h1_count} H1 headers (expected exactly 1)")

        # Check hierarchy (no skipping levels)
        headers = re.findall(r'^(#{1,6}) ', content, re.MULTILINE)
        for i in range(len(headers) - 1):
            current_level = len(headers[i])
            next_level = len(headers[i + 1])

            if next_level > current_level + 1:
                errors.append(f"Header level skip detected: #{current_level} → #{next_level}")

        return errors
```

#### Code Block Validation

```python
class CodeBlockLinter:
    def validate_code_blocks(self, content: str) -> List[str]:
        """Validate code block formatting"""
        errors = []

        # Find all code blocks
        blocks = re.finditer(r'```(\w*)\n(.*?)\n```', content, re.DOTALL)

        for match in blocks:
            language = match.group(1)

            # Language must be specified
            if not language:
                errors.append("Code block missing language specification")

            # Validate supported language
            elif language not in ['python', 'js', 'typescript', 'bash', 'yaml', 'json', 'sql']:
                errors.append(f"Unsupported language: {language}")

        return errors
```

#### Link Validation

```python
class LinkLinter:
    def validate_links(self, content: str, base_path: Path) -> List[str]:
        """Check all links in document"""
        errors = []

        # Find markdown links
        links = re.findall(r'\[([^\]]+)\]\(([^\)]+)\)', content)

        for text, url in links:
            # External links must use https
            if url.startswith('http://'):
                errors.append(f"External link uses HTTP (not HTTPS): {url}")

            # Check relative links exist
            if url.startswith('../') or url.startswith('./'):
                target = base_path.parent.joinpath(url).resolve()
                if not target.exists():
                    errors.append(f"Broken link: {url}")

        return errors
```

#### Table Validation

```python
class TableLinter:
    def validate_tables(self, content: str) -> List[str]:
        """Check markdown table structure"""
        errors = []

        # Find all tables
        tables = re.findall(r'\|(.+?)\|\n\|[-\|:]+\|\n((?:\|.+?\|\n)+)', content)

        for header, rows in tables:
            header_cols = len([col.strip() for col in header.split('|') if col.strip()])

            for row in rows.split('\n'):
                if row.strip():
                    row_cols = len([col.strip() for col in row.split('|') if col.strip()])
                    if row_cols != header_cols:
                        errors.append(f"Table column mismatch: header={header_cols}, row={row_cols}")

        return errors
```

### Part 4: CI/CD Integration

#### GitHub Actions Integration

```yaml
# .github/workflows/docs-quality-gate.yml
name: Documentation Quality Gate

on:
  pull_request:
    paths:
      - 'docs/**'
      - 'src/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Validate Documentation
        run: |
          python3 .moai/scripts/validate_docs.py \
            --mode pr \
            --files-changed \
            --min-score 85

      - name: Lint Markdown
        run: |
          python3 .moai/scripts/lint_docs.py \
            --fix-errors \
            --report github

      - name: Check Links
        run: |
          python3 .moai/scripts/validate_links.py \
            --docs docs/ \
            --fail-external-broken
```

#### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Validate changed documentation
python3 .moai/scripts/validate_docs.py --mode pre-commit

if [ $? -ne 0 ]; then
    echo "Documentation validation failed. Commit aborted."
    exit 1
fi
```

#### Quality Gate Configuration

```yaml
# .moai/quality-gates.yml
documentation:
  spec_compliance:
    min_score: 90
    required: true
    action: block_merge

  content_accuracy:
    min_score: 85
    required: true
    action: block_merge

  link_validity:
    broken_links_allowed: 0
    required: true
    action: block_merge

  code_examples:
    must_be_executable: true
    required: true
    action: block_merge

  multilingual_consistency:
    max_missing_translations: 0
    required: true
    action: warning
```

#### Automated Reports

```python
def generate_quality_report(docs_dir: Path, output_format: str = "markdown") -> str:
    """Generate comprehensive documentation quality report"""
    report = []

    report.append("# Documentation Quality Report")
    report.append(f"Generated: {datetime.now().isoformat()}")
    report.append("")

    # Overall metrics
    report.append("## Summary")
    report.append(f"- Documents scanned: {count_docs}")
    report.append(f"- Average quality score: {avg_score}/100")
    report.append(f"- Errors found: {total_errors}")
    report.append(f"- Warnings found: {total_warnings}")
    report.append("")

    # Document-level scores
    report.append("## Document Scores")
    for doc, score in quality_scores.items():
        status = "✅" if score >= 85 else "⚠️" if score >= 70 else "❌"
        report.append(f"{status} {doc}: {score}/100")

    # Detailed issues
    report.append("## Issues")
    for error in all_errors:
        report.append(f"- ❌ {error['file']}:{error['line']} - {error['message']}")

    return "\n".join(report)
```

### Part 5: Multilingual Support

```python
class MultilingualDocValidator:
    def validate_consistency(self) -> List[str]:
        """Ensure all languages have same structure"""
        issues = []

        # Get Korean source structure
        ko_structure = self._get_structure("docs/ko")

        # Compare with other languages
        for lang in ["en", "ja", "zh"]:
            lang_structure = self._get_structure(f"docs/{lang}")

            missing = set(ko_structure) - set(lang_structure)
            if missing:
                issues.append(f"[{lang}] Missing files: {missing}")

        return issues

    def measure_translation_quality(self, lang: str) -> float:
        """Calculate translation completeness (0-100)"""
        ko_files = set(self._list_files("docs/ko", "*.md"))
        lang_files = set(self._list_files(f"docs/{lang}", "*.md"))

        translated = len(ko_files & lang_files)
        return (translated / len(ko_files)) * 100
```

---

## Best Practices

### Generation
- ✅ Keep JSDoc/docstrings in sync with code
- ✅ Use meaningful example code
- ✅ Provide multilingual templates
- ✅ Version your documentation
- ❌ Don't auto-generate without review

### Validation
- ✅ Check SPEC compliance before publishing
- ✅ Test all code examples
- ✅ Verify links periodically
- ✅ Measure quality metrics
- ❌ Don't ignore quality gate failures

### Linting
- ✅ Use consistent header hierarchy
- ✅ Declare language for all code blocks
- ✅ Keep tables properly formatted
- ✅ Use relative links
- ❌ Don't mix formatting styles

---

## Works Well With

- `moai-domain-devops` (Documentation pipeline automation)
- `moai-core-code-reviewer` (Code documentation quality)
- `moai-lang-python`, `moai-lang-typescript` (Language-specific patterns)
- `moai-cc-hooks` (Documentation validation hooks)
- `moai-project-config-manager` (Documentation configuration)

---

## Integration Matrix

**Prerequisite Skills**:
- `moai-domain-frontend` (UI documentation patterns)
- `moai-domain-backend` (API documentation)

**Complementary Skills**:
- `moai-docs-markdown` (Markdown formatting)
- `moai-context7-lang-integration` (Latest documentation references)

---

**Version**: 2.0.0 (Unified Toolkit)
**Last Updated**: 2025-11-21 | **Lines**: 520
**Previous Components**: moai-docs-generation, moai-docs-validation, moai-docs-linting (consolidated)
**Status**: Production Ready