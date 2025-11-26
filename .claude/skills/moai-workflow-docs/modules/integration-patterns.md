---
name: moai-workflow-docs-integration
description: CI/CD integration patterns and automation workflows for documentation validation
---

## Integration Patterns & Automation

### Single Script Execution

```bash
# Phase 1: Markdown linting
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py

# Phase 2: Mermaid validation
uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py

# Phase 2.5: Mermaid code extraction
uv run .claude/skills/moai-workflow-docs/scripts/extract_mermaid_details.py

# Phase 3: Korean typography
uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py

# Phase 4: Comprehensive report
uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py
```

---

### Complete Validation Pipeline

```bash
#!/bin/bash
# Run all 5 phases sequentially

echo "Running Phase 1: Markdown Linting..."
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py

echo "Running Phase 2: Mermaid Validation..."
uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py

echo "Running Phase 2.5: Mermaid Detail Extraction..."
uv run .claude/skills/moai-workflow-docs/scripts/extract_mermaid_details.py

echo "Running Phase 3: Korean Typography..."
uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py

echo "Running Phase 4: Comprehensive Report..."
uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py

echo "All validation phases complete!"
echo "Check .moai/reports/ for generated files:"
ls -lh .moai/reports/*.txt
```

---

### CI/CD Integration

**GitHub Actions Integration**:

```yaml
# .github/workflows/docs-validation.yml
name: Documentation Validation

on:
  pull_request:
    paths:
      - 'docs/**'
  push:
    branches:
      - develop
      - main

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install uv
        run: pip install uv

      - name: Run Documentation Validation Suite
        run: |
          # Phase 1: Markdown Linting
          uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py

          # Phase 2: Mermaid Validation
          uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py

          # Phase 3: Korean Typography
          uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py

          # Phase 4: Comprehensive Report
          uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py

      - name: Upload Validation Reports
        uses: actions/upload-artifact@v3
        with:
          name: documentation-reports
          path: .moai/reports/*.txt
        if: always()

      - name: Check for Critical Issues
        run: |
          # Fail CI if critical issues found
          if grep -q "Priority 1" .moai/reports/korean_docs_comprehensive_review.txt; then
            echo "❌ Critical documentation issues found"
            exit 1
          fi
```

---

### Pre-Commit Hook Integration

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Run documentation validation on staged markdown files
STAGED_MD_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep "\.md$")

if [ -n "$STAGED_MD_FILES" ]; then
    echo "Running documentation validation on staged files..."
    
    # Run markdown linting
    uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
    
    if [ $? -ne 0 ]; then
        echo "❌ Markdown linting failed. Please fix issues before committing."
        exit 1
    fi
    
    echo "✅ Documentation validation passed"
fi

exit 0
```

---

### Makefile Integration

```makefile
# Makefile
.PHONY: docs-validate docs-validate-quick docs-reports

docs-validate:
	@echo "Running complete documentation validation suite..."
	uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
	uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py
	uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py
	uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py
	@echo "✅ Validation complete. Check .moai/reports/"

docs-validate-quick:
	@echo "Running quick validation (lint + mermaid only)..."
	uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
	uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py

docs-reports:
	@echo "Generating comprehensive validation report..."
	uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py
	cat .moai/reports/korean_docs_comprehensive_review.txt
```

---

### NPM Scripts Integration

```json
{
  "scripts": {
    "docs:validate": "bash -c 'uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py && uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py'",
    "docs:report": "uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py",
    "docs:full": "bash -c 'npm run docs:validate && npm run docs:report'"
  }
}
```

---

### Docker Integration

```dockerfile
# Dockerfile for documentation validation
FROM python:3.11-slim

WORKDIR /workspace

# Install uv
RUN pip install --no-cache-dir uv

# Copy validation scripts
COPY .claude/skills/moai-workflow-docs/scripts /scripts

# Run validation
CMD ["sh", "-c", "uv run /scripts/lint_korean_docs.py && uv run /scripts/validate_mermaid_diagrams.py && uv run /scripts/generate_final_comprehensive_report.py"]
```

**Usage**:
```bash
docker build -t docs-validator .
docker run -v $(pwd):/workspace docs-validator
```

---

### Best Practices

**Validation Frequency**:
- ✅ On every commit (pre-commit hook)
- ✅ On pull request (GitHub Actions)
- ✅ Daily scheduled runs (catch drift)
- ✅ Before releases (quality gate)

**Report Management**:
- ✅ Store reports in `.moai/reports/`
- ✅ Gitignore generated reports
- ✅ Archive reports from releases
- ✅ Track metrics over time

**Error Handling**:
- ✅ Fail CI on Priority 1 issues
- ✅ Warn on Priority 2 issues
- ✅ Log Priority 3 issues
- ✅ Provide actionable fix suggestions

---

**End of Module** | moai-workflow-docs-integration
