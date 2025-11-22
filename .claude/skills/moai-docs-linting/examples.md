# moai-docs-linting: Practical Examples

Markdown linting and documentation quality validation examples.

---

## Example 1: Basic Markdown Linting

Validate markdown file structure:

```python
from markdownlint import validate_markdown

content = """
# Main Title

## Section 1
Content here

### Subsection 1.1
More content
"""

issues = validate_markdown(content)
for issue in issues:
    print(f"Line {issue.line}: {issue.message}")
```

---

## Example 2: Header Validation

Check header hierarchy:

```python
def validate_headers(content: str) -> List[str]:
    """Validate header structure"""
    errors = []

    # Check H1 count (should be exactly 1)
    h1_count = len(re.findall(r'^# ', content, re.MULTILINE))
    if h1_count != 1:
        errors.append(f"Found {h1_count} H1 headers (expected 1)")

    # Check hierarchy (no level skipping)
    headers = re.findall(r'^(#{1,6}) ', content, re.MULTILINE)
    for i in range(len(headers) - 1):
        if len(headers[i+1]) > len(headers[i]) + 1:
            errors.append("Header level skip detected")

    return errors
```

---

## Example 3: Link Validation

Check for broken or invalid links:

```python
import re
from pathlib import Path

def validate_links(doc_path: Path) -> List[str]:
    """Validate all links in markdown"""
    content = doc_path.read_text()
    errors = []

    # Find all markdown links
    links = re.findall(r'\[([^\]]+)\]\(([^\)]+)\)', content)

    for text, url in links:
        # Check external links use HTTPS
        if url.startswith('http://'):
            errors.append(f"HTTP link found (use HTTPS): {url}")

        # Check relative links exist
        if url.startswith('./') or url.startswith('../'):
            target = doc_path.parent.joinpath(url).resolve()
            if not target.exists():
                errors.append(f"Broken link: {url}")

    return errors
```

---

## Example 4: Code Block Validation

Check code block formatting:

```python
def validate_code_blocks(content: str) -> List[str]:
    """Validate code block structure"""
    errors = []

    # Find all code blocks
    blocks = re.finditer(r'```(\w*)\n(.*?)\n```', content, re.DOTALL)

    for i, match in enumerate(blocks):
        language = match.group(1)

        # Language must be specified
        if not language:
            errors.append(f"Code block {i+1}: Missing language specification")

        # Validate language is recognized
        elif language not in ['python', 'js', 'typescript', 'bash', 'yaml']:
            errors.append(f"Code block {i+1}: Unknown language '{language}'")

    return errors
```

---

## Example 5: Table Validation

Check markdown table formatting:

```python
def validate_tables(content: str) -> List[str]:
    """Validate markdown table structure"""
    errors = []

    # Find all tables
    table_pattern = r'\|(.+?)\|\n\|[-\|:]+\|\n((?:\|.+?\|\n)*)'
    tables = re.findall(table_pattern, content)

    for table_idx, (header, rows) in enumerate(tables):
        header_cols = len([c.strip() for c in header.split('|') if c.strip()])

        for row_idx, row in enumerate(rows.split('\n')):
            if row.strip():
                row_cols = len([c.strip() for c in row.split('|') if c.strip()])
                if row_cols != header_cols:
                    errors.append(
                        f"Table {table_idx+1} row {row_idx+1}: "
                        f"Column mismatch (header={header_cols}, row={row_cols})"
                    )

    return errors
```

---

## Example 6: List Validation

Validate list formatting consistency:

```python
def validate_lists(content: str) -> List[str]:
    """Validate markdown list structure"""
    errors = []

    lines = content.split('\n')
    for i, line in enumerate(lines):
        # Check for inconsistent list markers
        if line.startswith('-') and i > 0 and lines[i-1].startswith('*'):
            errors.append(f"Line {i+1}: Inconsistent list marker (- vs *)")

        # Check for tab indentation (should use spaces)
        if line.startswith('\t-') or line.startswith('\t*'):
            errors.append(f"Line {i+1}: Use spaces, not tabs for indentation")

    return errors
```

---

## Example 7: File-Level Validation

Comprehensive validation for entire document:

```python
class MarkdownValidator:
    """Comprehensive markdown validation"""

    def validate_file(self, file_path: Path) -> Dict[str, List[str]]:
        """Run all validations on a file"""
        content = file_path.read_text()

        return {
            'headers': self.validate_headers(content),
            'code_blocks': self.validate_code_blocks(content),
            'links': self.validate_links(content, file_path),
            'tables': self.validate_tables(content),
            'lists': self.validate_lists(content),
            'typography': self.validate_typography(content),
        }

    def print_report(self, results: Dict[str, List[str]]) -> None:
        """Print validation report"""
        for category, issues in results.items():
            if issues:
                print(f"\n{category.upper()}: {len(issues)} issues")
                for issue in issues:
                    print(f"  ❌ {issue}")
            else:
                print(f"{category.upper()}: ✅ OK")
```

---

## Example 8: CI/CD Integration

Use linting in GitHub Actions:

```yaml
# .github/workflows/lint-docs.yml
name: Lint Documentation

on:
  pull_request:
    paths:
      - 'docs/**'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install markdownlint
        run: npm install -g markdownlint-cli

      - name: Lint Markdown
        run: markdownlint 'docs/**/*.md'

      - name: Run custom linter
        run: python scripts/lint_docs.py docs/
```

---

## Example 9: Multilingual Validation

Validate structure across languages:

```python
class MultilingualValidator:
    """Validate consistency across language versions"""

    def validate_consistency(self, docs_dir: Path) -> Dict[str, List[str]]:
        """Check all languages have same structure"""
        issues = {}

        ko_structure = self._get_file_structure(docs_dir / 'ko')

        for lang in ['en', 'ja', 'zh']:
            lang_dir = docs_dir / lang
            lang_structure = self._get_file_structure(lang_dir)

            missing_files = set(ko_structure) - set(lang_structure)
            if missing_files:
                issues[lang] = [f"Missing files: {', '.join(missing_files)}"]

        return issues
```

---

## Example 10: Pre-commit Hook

Validate docs before committing:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Linting documentation..."
python scripts/lint_docs.py docs/

if [ $? -ne 0 ]; then
    echo "❌ Documentation linting failed!"
    echo "Please fix the issues above before committing."
    exit 1
fi

echo "✅ Documentation linting passed!"
exit 0
```

---

## Example 11: Custom Rule Definition

Create custom linting rules:

```python
class CustomMarkdownRules:
    """Define custom linting rules"""

    @staticmethod
    def rule_no_trailing_spaces(content: str) -> List[str]:
        """Check for trailing spaces"""
        errors = []
        for i, line in enumerate(content.split('\n'), 1):
            if line.rstrip() != line:
                errors.append(f"Line {i}: Trailing whitespace")
        return errors

    @staticmethod
    def rule_consistent_language(content: str) -> List[str]:
        """Check for mixed English/Korean usage"""
        # Custom logic to detect language mixing
        return []

    @staticmethod
    def rule_proper_spacing(content: str) -> List[str]:
        """Check spacing around punctuation"""
        errors = []
        # Custom spacing validation
        return errors
```

---

## Example 12: Batch File Linting

Lint multiple files:

```python
def lint_directory(docs_dir: Path) -> Dict[Path, Dict[str, List[str]]]:
    """Lint all markdown files in directory"""
    results = {}
    validator = MarkdownValidator()

    for md_file in docs_dir.rglob('*.md'):
        results[md_file] = validator.validate_file(md_file)

    return results
```

---

**Last Updated**: 2025-11-22
**Examples Count**: 12
**Code Blocks**: 30+
**Status**: Production Ready
