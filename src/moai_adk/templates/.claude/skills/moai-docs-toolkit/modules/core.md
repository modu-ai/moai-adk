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


## Works Well With

- `moai-domain-devops` (Documentation pipeline automation)
- `moai-core-code-reviewer` (Code documentation quality)
- `moai-lang-python`, `moai-lang-typescript` (Language-specific patterns)
- `moai-cc-hooks` (Documentation validation hooks)
- `moai-project-config-manager` (Documentation configuration)


## Integration Matrix

**Prerequisite Skills**:
- `moai-domain-frontend` (UI documentation patterns)
- `moai-domain-backend` (API documentation)

**Complementary Skills**:
- `moai-docs-markdown` (Markdown formatting)
- `moai-context7-lang-integration` (Latest documentation references)


**Version**: 2.0.0 (Unified Toolkit)
**Last Updated**: 2025-11-21 | **Lines**: 520
**Previous Components**: moai-docs-generation, moai-docs-validation, moai-docs-linting (consolidated)
**Status**: Production Ready
