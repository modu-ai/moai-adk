---
name: moai-docs-unified
version: 4.0.0
updated: 2025-11-20
status: stable
tier: specialization
description: Unified documentation validation ecosystem with AI-powered features
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Unified Documentation Expert

**Comprehensive Documentation Validation & Management**

> **Focus**: Markdown Linting, Mermaid Validation, Korean Typography, CI/CD  
> **Tools**: Python, uv, GitHub Actions

---

## Overview

A complete ecosystem for maintaining high-quality documentation across multiple languages.

### Core Components

| Phase   | Script                                   | Purpose                                         |
| ------- | ---------------------------------------- | ----------------------------------------------- |
| **1**   | `lint_korean_docs.py`                    | Markdown syntax, structure, and link validation |
| **2**   | `validate_mermaid_diagrams.py`           | Mermaid syntax check and rendering verification |
| **2.5** | `extract_mermaid_details.py`             | Diagram code extraction and preview generation  |
| **3**   | `validate_korean_typography.py`          | Korean encoding, spacing, and punctuation check |
| **4**   | `generate_final_comprehensive_report.py` | Aggregated quality reporting and prioritization |

---

## Usage Patterns

### 1. Single Script Execution

Run individual validation phases using `uv`.

```bash
# Phase 1: Markdown Linting
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py --path docs/src/ko

# Phase 2: Mermaid Validation
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

# Phase 3: Korean Typography
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

# Phase 4: Generate Report
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py
```

### 2. Complete Validation Pipeline

Run all checks sequentially.

```bash
#!/bin/bash
echo "Running Documentation Validation Suite..."

# Run all phases
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

echo "Check reports in .moai/reports/"
```

### 3. CI/CD Integration (GitHub Actions)

Automate validation on Pull Requests.

```yaml
name: Docs Validation
on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: "3.11"

      - name: Install uv
        run: pip install uv

      - name: Run Validation
        run: |
          uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
          uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py
          uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py
          uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

      - name: Upload Reports
        uses: actions/upload-artifact@v3
        with:
          name: doc-reports
          path: .moai/reports/*.txt
```

---

## Script Details

### Markdown Linting (`lint_korean_docs.py`)

- **Checks**: Header hierarchy, code block syntax, link validity, list formatting.
- **Output**: Detailed error list by file and line number.

### Mermaid Validation (`validate_mermaid_diagrams.py`)

- **Supported Types**: Flowchart, Sequence, Class, State, ER, Gantt.
- **Checks**: Syntax validity, node relationships, complexity metrics.

### Korean Typography (`validate_korean_typography.py`)

- **Checks**: UTF-8 encoding, full-width characters, spacing rules.
- **Goal**: Ensure professional and consistent Korean text.

### Comprehensive Report (`generate_final_comprehensive_report.py`)

- **Features**: Aggregates all findings, assigns priority (Critical/High/Low), and generates an executive summary score.

---

## Validation Checklist

- [ ] **Structure**: Headers are hierarchical (H1 -> H2 -> H3).
- [ ] **Links**: All relative links resolve to existing files.
- [ ] **Code**: Code blocks have language identifiers (e.g., \`\`\`python).
- [ ] **Diagrams**: Mermaid syntax is valid and renders correctly.
- [ ] **Language**: Korean text follows typography standards.

---

## Related Skills

- `moai-mermaid-diagram-expert`: Creating valid diagrams
- `moai-project-documentation`: Documentation standards
- `moai-devops-ci`: CI/CD pipeline configuration

---

**Last Updated**: 2025-11-20
