# Unified Documentation Validation Examples

Practical examples for running documentation validation scripts and integrating them into workflows.

---

## Basic Script Usage

### Phase 1: Markdown Linting

```bash
# Lint all Korean documentation
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py --path docs/src/ko

# Lint specific file
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py --path docs/src/ko/guide/installation.md

# Output detailed errors
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py --path docs/src/ko --verbose

# Save report to file
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py --path docs/src/ko > .moai/reports/markdown-lint.txt
```

**Expected Output**:

```
=== Markdown Linting Report ===
File: docs/src/ko/guide/installation.md
  Line 15: Missing language identifier in code block
  Line 23: Broken link: ./quickstart.mdd (should be .md)
  Line 45: Header hierarchy skip (H2 -> H4)

File: docs/src/ko/api/functions.md
  Line 8: Unordered list indentation incorrect

Summary:
  Total Files: 12
  Files with Issues: 2
  Total Issues: 4
```

### Phase 2: Mermaid Diagram Validation

```bash
# Validate all Mermaid diagrams in documentation
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

# Validate specific directory
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py --path docs/src/ko/architecture

# Generate preview images (requires mermaid-cli)
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py --generate-previews

# Check only specific diagram types
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py --types flowchart,sequence
```

**Expected Output**:

```
=== Mermaid Diagram Validation ===
Diagram 1: docs/src/ko/architecture/system-overview.md (Line 45)
  Type: flowchart TD
  Nodes: 8
  Edges: 10
  Status: ✓ Valid

Diagram 2: docs/src/ko/api/auth-flow.md (Line 120)
  Type: sequenceDiagram
  Participants: 4
  Messages: 12
  Status: ✗ Invalid
  Error: Syntax error at line 3: Missing arrow operator

Summary:
  Total Diagrams: 45
  Valid: 44
  Invalid: 1
```

### Phase 2.5: Extract Mermaid Details

```bash
# Extract all Mermaid code blocks
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py

# Extract and generate preview files
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py --export-code --output-dir .moai/diagrams

# Generate statistics
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py --stats
```

**Expected Output**:

```
=== Mermaid Extraction Report ===
Extracted 45 diagrams:
  - Flowcharts: 18
  - Sequence Diagrams: 12
  - Class Diagrams: 8
  - State Diagrams: 5
  - ER Diagrams: 2

Exported to: .moai/diagrams/
```

### Phase 3: Korean Typography Validation

```bash
# Validate Korean typography rules
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

# Check specific file
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py --path docs/src/ko/guide/quickstart.md

# Fix common issues automatically (experimental)
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py --auto-fix
```

**Expected Output**:

```
=== Korean Typography Validation ===
File: docs/src/ko/guide/installation.md
  Line 12: Missing space after Korean punctuation: "설치방법은." → "설치 방법은."
  Line 18: Half-width parentheses should be full-width: (예시) → (예시)
  Line 25: Inconsistent quotation marks: "설정" → 「설정」

File: docs/src/ko/api/reference.md
  Line 34: English/Korean spacing issue: "API키를" → "API 키를"

Summary:
  Total Files: 12
  Files with Issues: 2
  Total Issues: 4
```

### Phase 4: Generate Comprehensive Report

```bash
# Generate final aggregated report
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

# Generate HTML report
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py --format html --output .moai/reports/report.html

# Generate JSON for programmatic access
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py --format json --output .moai/reports/report.json
```

**Expected Output**:

```
=== Documentation Quality Report ===
Generated: 2025-11-20 14:48:52

Overall Score: 92/100 (Excellent)

Critical Issues: 1
  - Broken link in installation.md (Line 23)

High Priority: 3
  - Invalid Mermaid diagram in auth-flow.md
  - Missing code language identifier (2 occurrences)

Low Priority: 8
  - Korean spacing improvements
  - Quotation mark consistency

Recommendations:
  1. Fix broken link in installation.md
  2. Fix Mermaid syntax error in auth-flow.md
  3. Add language identifiers to code blocks
  4. Review Korean typography in 2 files

Report saved to: .moai/reports/comprehensive-report.txt
```

---

## Complete Validation Pipeline

### Bash Script

```bash
#!/bin/bash
# validate-docs.sh
# Complete documentation validation pipeline

set -e  # Exit on error

PROJECT_ROOT=$(pwd)
REPORTS_DIR=".moai/reports"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")

echo "📚 Starting Documentation Validation Suite"
echo "Timestamp: $TIMESTAMP"
echo "Project: $PROJECT_ROOT"
echo ""

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Phase 1: Markdown Linting
echo "▶️  Phase 1: Markdown Linting"
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
  --path docs/src/ko \
  > "$REPORTS_DIR/markdown-lint-$TIMESTAMP.txt" 2>&1
echo "✓ Markdown linting complete"
echo ""

# Phase 2: Mermaid Validation
echo "▶️  Phase 2: Mermaid Diagram Validation"
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py \
  > "$REPORTS_DIR/mermaid-validation-$TIMESTAMP.txt" 2>&1
echo "✓ Mermaid validation complete"
echo ""

# Phase 2.5: Extract Mermaid Details
echo "▶️  Phase 2.5: Mermaid Extraction"
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py \
  --export-code \
  --output-dir ".moai/diagrams" \
  > "$REPORTS_DIR/mermaid-extraction-$TIMESTAMP.txt" 2>&1
echo "✓ Mermaid extraction complete"
echo ""

# Phase 3: Korean Typography
echo "▶️  Phase 3: Korean Typography Validation"
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py \
  > "$REPORTS_DIR/korean-typography-$TIMESTAMP.txt" 2>&1
echo "✓ Korean typography validation complete"
echo ""

# Phase 4: Generate Comprehensive Report
echo "▶️  Phase 4: Generating Comprehensive Report"
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py \
  --format html \
  --output "$REPORTS_DIR/comprehensive-report-$TIMESTAMP.html"
echo "✓ Comprehensive report generated"
echo ""

echo "✅ Documentation Validation Complete"
echo "Reports available in: $REPORTS_DIR"
echo ""
echo "View comprehensive report:"
echo "  open $REPORTS_DIR/comprehensive-report-$TIMESTAMP.html"
```

**Usage**:

```bash
# Make script executable
chmod +x validate-docs.sh

# Run validation
./validate-docs.sh
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/docs-validation.yml
name: Documentation Validation

on:
  pull_request:
    paths:
      - "docs/**"
      - "**.md"
  push:
    branches: [main, develop]
    paths:
      - "docs/**"
      - "**.md"

jobs:
  validate:
    name: Validate Documentation
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.11"
          cache: "pip"

      - name: Install uv
        run: |
          pip install uv
          uv --version

      - name: Create reports directory
        run: mkdir -p .moai/reports

      - name: Phase 1 - Markdown Linting
        id: markdown-lint
        run: |
          uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
            --path docs/src/ko \
            > .moai/reports/markdown-lint.txt || echo "markdown_issues=true" >> $GITHUB_OUTPUT
        continue-on-error: true

      - name: Phase 2 - Mermaid Validation
        id: mermaid-validation
        run: |
          uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py \
            > .moai/reports/mermaid-validation.txt || echo "mermaid_issues=true" >> $GITHUB_OUTPUT
        continue-on-error: true

      - name: Phase 3 - Korean Typography
        id: korean-typography
        run: |
          uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py \
            > .moai/reports/korean-typography.txt || echo "typography_issues=true" >> $GITHUB_OUTPUT
        continue-on-error: true

      - name: Phase 4 - Generate Report
        run: |
          uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py \
            --format html \
            --output .moai/reports/comprehensive-report.html

          uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py \
            --format json \
            --output .moai/reports/comprehensive-report.json

      - name: Upload Reports
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: documentation-reports
          path: .moai/reports/
          retention-days: 30

      - name: Parse JSON Report
        id: parse-report
        run: |
          # Extract critical and high priority issue counts
          CRITICAL=$(jq '.summary.critical_issues' .moai/reports/comprehensive-report.json)
          HIGH=$(jq '.summary.high_priority_issues' .moai/reports/comprehensive-report.json)
          SCORE=$(jq '.summary.overall_score' .moai/reports/comprehensive-report.json)

          echo "critical_issues=$CRITICAL" >> $GITHUB_OUTPUT
          echo "high_priority_issues=$HIGH" >> $GITHUB_OUTPUT
          echo "overall_score=$SCORE" >> $GITHUB_OUTPUT

      - name: Comment on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const reportPath = '.moai/reports/comprehensive-report.json';
            const report = JSON.parse(fs.readFileSync(reportPath, 'utf8'));

            const comment = `## 📚 Documentation Validation Report

            **Overall Score:** ${report.summary.overall_score}/100

            ### Summary
            - ❌ Critical Issues: ${report.summary.critical_issues}
            - ⚠️  High Priority: ${report.summary.high_priority_issues}
            - ℹ️  Low Priority: ${report.summary.low_priority_issues}

            ### Details
            ${report.recommendations.map((rec, i) => `${i + 1}. ${rec}`).join('\n')}

            [View Full Report](${process.env.GITHUB_SERVER_URL}/${process.env.GITHUB_REPOSITORY}/actions/runs/${process.env.GITHUB_RUN_ID})
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });

      - name: Fail if critical issues found
        if: steps.parse-report.outputs.critical_issues > 0
        run: |
          echo "❌ Critical issues found: ${{ steps.parse-report.outputs.critical_issues }}"
          exit 1
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
# Run documentation validation before commit

echo "Running documentation validation..."

# Only validate if markdown files are being committed
STAGED_MD_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.md$')

if [ -z "$STAGED_MD_FILES" ]; then
  echo "No markdown files to validate"
  exit 0
fi

# Run quick validation on staged files
TEMP_REPORT=$(mktemp)

for file in $STAGED_MD_FILES; do
  echo "Validating: $file"

  # Run markdown linting on individual file
  uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
    --path "$file" \
    >> "$TEMP_REPORT" 2>&1
done

# Check if there are critical issues
if grep -q "Error\|Critical" "$TEMP_REPORT"; then
  echo "❌ Documentation validation failed:"
  cat "$TEMP_REPORT"
  rm "$TEMP_REPORT"
  exit 1
fi

rm "$TEMP_REPORT"
echo "✓ Documentation validation passed"
exit 0
```

**Install pre-commit hook**:

```bash
# Make executable and install
chmod +x .git/hooks/pre-commit
```

---

## Custom Script Integration

### Python Script Example

```python
#!/usr/bin/env python3
"""
Custom documentation validation wrapper.
Integrates with existing validation scripts.
"""

import subprocess
import sys
from pathlib import Path
from datetime import datetime


class DocValidator:
    """Wrapper for documentation validation scripts."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.reports_dir = project_root / ".moai" / "reports"
        self.scripts_dir = project_root / ".claude" / "skills" / "moai-docs-unified" / "scripts"

        # Ensure reports directory exists
        self.reports_dir.mkdir(parents=True, exist_ok=True)

    def run_script(self, script_name: str, *args) -> tuple[int, str]:
        """Run a validation script and return exit code and output."""
        script_path = self.scripts_dir / script_name
        cmd = ["uv", "run", str(script_path), *args]

        print(f"Running: {' '.join(cmd)}")

        result = subprocess.run(
            cmd,
            cwd=self.project_root,
            capture_output=True,
            text=True
        )

        return result.returncode, result.stdout + result.stderr

    def validate_all(self) -> dict:
        """Run all validation phases."""
        results = {}

        # Phase 1: Markdown Linting
        print("\n📝 Phase 1: Markdown Linting")
        code, output = self.run_script("lint_korean_docs.py", "--path", "docs/src/ko")
        results["markdown"] = {"code": code, "output": output}

        # Phase 2: Mermaid Validation
        print("\n📊 Phase 2: Mermaid Validation")
        code, output = self.run_script("validate_mermaid_diagrams.py")
        results["mermaid"] = {"code": code, "output": output}

        # Phase 3: Korean Typography
        print("\n🇰🇷 Phase 3: Korean Typography")
        code, output = self.run_script("validate_korean_typography.py")
        results["typography"] = {"code": code, "output": output}

        # Phase 4: Generate Report
        print("\n📋 Phase 4: Comprehensive Report")
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        report_path = self.reports_dir / f"report-{timestamp}.html"

        code, output = self.run_script(
            "generate_final_comprehensive_report.py",
            "--format", "html",
            "--output", str(report_path)
        )
        results["report"] = {"code": code, "output": output, "path": report_path}

        return results

    def print_summary(self, results: dict):
        """Print validation summary."""
        print("\n" + "=" * 60)
        print("📊 VALIDATION SUMMARY")
        print("=" * 60)

        for phase, result in results.items():
            status = "✓" if result["code"] == 0 else "✗"
            print(f"{status} {phase.capitalize()}: {'PASSED' if result['code'] == 0 else 'FAILED'}")

        if "report" in results and "path" in results["report"]:
            print(f"\n📄 Full report: {results['report']['path']}")

        # Overall status
        all_passed = all(r["code"] == 0 for r in results.values())
        if all_passed:
            print("\n✅ All validations passed!")
            return 0
        else:
            print("\n❌ Some validations failed. Review reports for details.")
            return 1


def main():
    """Main entry point."""
    project_root = Path.cwd()
    validator = DocValidator(project_root)

    print("🚀 Starting Documentation Validation Suite")
    print(f"📁 Project: {project_root}")

    results = validator.validate_all()
    exit_code = validator.print_summary(results)

    sys.exit(exit_code)


if __name__ == "__main__":
    main()
```

**Usage**:

```bash
# Make executable
chmod +x custom-validate.py

# Run validation
./custom-validate.py
```

---

**See also**: [SKILL.md](./SKILL.md) for validation ecosystem overview and script details
