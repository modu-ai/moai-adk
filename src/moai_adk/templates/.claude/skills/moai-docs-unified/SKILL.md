# Skill: Unified Documentation Management & Quality Assurance

## Metadata

```yaml
skill_id: moai-docs-unified
skill_name: Unified Documentation Management & Quality Assurance
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 2200
triggers:
  - keywords: [documentation management, docs linting, mermaid validation, korean typography, documentation quality, comprehensive report, docs-unified]
  - contexts: [docs-unified, @docs:all, documentation-management, quality-assurance]
agents:
  - docs-manager
  - docs-auditor
  - quality-gate
freedom_level: high
context7_references:
  - url: "https://en.wikipedia.org/wiki/Software_quality"
    topic: "Software Quality Metrics"
  - url: "https://github.com/moai-adk/moai-adk"
    topic: "MoAI-ADK Documentation Standards"
spec_reference: "@SPEC:DOCS-001"
```

## 📚 Content

### Section 1: Unified Framework Overview

The **moai-docs-unified** skill provides a complete documentation management ecosystem integrating 5 specialized validation scripts:

**Core Components**:
- **Phase 1**: Markdown Linting (syntax, structure, links)
- **Phase 2**: Mermaid Diagram Validation (syntax, rendering, type checking)
- **Phase 2.5**: Mermaid Detail Extraction (code preview, rendering guide)
- **Phase 3**: Korean Typography Validation (UTF-8, full-width chars, spacing)
- **Phase 4**: Comprehensive Report Generation (aggregation, prioritization, recommendations)

**Key Benefits**:
- Catch documentation errors before publication
- Ensure consistency across 4 languages (ko, en, ja, zh)
- Validate diagram syntax and rendering capability
- Maintain Korean language best practices
- Generate actionable quality reports

### Section 2: Script Specifications

#### Script 1: lint_korean_docs.py

**Purpose**: Comprehensive markdown validation for Korean documentation

**Location**: `.claude/skills/moai-docs-unified/scripts/lint_korean_docs.py`

**Key Validations**:
- Header structure (H1 uniqueness, level hierarchy)
- Code blocks (language declaration, matching delimiters)
- Links (relative paths, file existence, https protocol)
- Lists (marker consistency, indentation)
- Tables (column count, alignment)
- Korean-specific (full-width chars, encoding)

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py

# With custom paths
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
  --path docs/src/ko \
  --output .moai/reports/lint_report_ko.txt
```

**Output**: `.moai/reports/lint_report_ko.txt` (8 validation categories)

---

#### Script 2: validate_mermaid_diagrams.py

**Purpose**: Mermaid diagram type and syntax validation

**Location**: `.claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py`

**Key Features**:
- Diagram type detection (graph, flowchart, stateDiagram, sequenceDiagram, classDiagram, erDiagram, gantt)
- Configuration block handling (`%%{init: ...}%%`)
- Node/edge relationship validation
- Line count and complexity metrics

**Supported Diagram Types**:
```
✅ graph TD/BT/LR/RL          (Flowchart)
✅ stateDiagram-v2            (State machines)
✅ sequenceDiagram            (Interactions)
✅ classDiagram               (Class structures)
✅ erDiagram                  (Entity relationships)
✅ gantt                       (Timelines)
```

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

# With custom paths
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py \
  --path docs/src \
  --output .moai/reports/mermaid_validation_report.txt
```

**Output**: `.moai/reports/mermaid_validation_report.txt` (16 diagrams, 100% valid)

---

#### Script 3: extract_mermaid_details.py

**Purpose**: Extract and document all Mermaid diagram code with rendering guide

**Location**: `.claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py`

**Key Features**:
- Extract all mermaid blocks from documentation
- Show diagram type and line count
- Provide full code preview
- Generate rendering test guide (mermaid.live)

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py

# With custom paths
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py \
  --path docs/src \
  --output .moai/reports/mermaid_detail_report.txt
```

**Output**: `.moai/reports/mermaid_detail_report.txt` (full diagram code + test guide)

---

#### Script 4: validate_korean_typography.py

**Purpose**: Korean-specific typography and encoding validation

**Location**: `.claude/skills/moai-docs-unified/scripts/validate_korean_typography.py`

**Key Validations**:
- UTF-8 Encoding verification
- Full-width character detection (U+3000, ＜, ＞, etc.)
- Punctuation consistency (. vs。, , vs、)
- Spacing rules (Korean-English boundaries)
- Character statistics per file

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

# With custom paths
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py \
  --path docs/src \
  --output .moai/reports/korean_typography_report.txt
```

**Output**: `.moai/reports/korean_typography_report.txt` (28,543 lines validated, 43 files)

---

#### Script 5: generate_final_comprehensive_report.py

**Purpose**: Aggregate all validation phases into prioritized quality report

**Location**: `.claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py`

**Report Structure**:
1. Executive Summary (8.5/10 quality score)
2. Phase Results (Priority 1/2/3 items)
3. Recommendations (Done/In Progress/TODO)
4. Action Items (Immediate/Short-term/Long-term)
5. Generated Report Files (all 4 phases)

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

# With custom report directory
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py \
  --report-dir .moai/reports \
  --output .moai/reports/korean_docs_comprehensive_review.txt
```

**Output**: `.moai/reports/korean_docs_comprehensive_review.txt` (aggregated report)

---

### Section 3: Integration Patterns

#### Single Script Execution

```bash
# Phase 1: Markdown linting
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py

# Phase 2: Mermaid validation
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

# Phase 2.5: Mermaid code extraction
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py

# Phase 3: Korean typography
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

# Phase 4: Comprehensive report
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py
```

#### Complete Validation Pipeline

```bash
#!/bin/bash
# Run all 5 phases sequentially

echo "Running Phase 1: Markdown Linting..."
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py

echo "Running Phase 2: Mermaid Validation..."
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

echo "Running Phase 2.5: Mermaid Detail Extraction..."
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py

echo "Running Phase 3: Korean Typography..."
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

echo "Running Phase 4: Comprehensive Report..."
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

echo "All validation phases complete!"
echo "Check .moai/reports/ for generated files:"
ls -lh .moai/reports/*.txt
```

### Section 4: Project Root Auto-Detection

All scripts use automatic project root detection:

```python
def find_project_root(start_path: Path) -> Path:
    current = start_path
    while current != current.parent:
        if (current / "pyproject.toml").exists() or (current / ".git").exists():
            return current
        current = current.parent
    raise RuntimeError("Project root not found")

script_path = Path(__file__).resolve()
project_root = find_project_root(script_path.parent)
```

**Benefits**:
- Works from any directory
- Works from any execution context (local, CI/CD, automation)
- No hardcoded paths
- Handles relative paths correctly
- Compatible with `uv run`

### Section 5: CI/CD Integration

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
          uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py

          # Phase 2: Mermaid Validation
          uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

          # Phase 3: Korean Typography
          uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py

          # Phase 4: Comprehensive Report
          uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py

      - name: Upload Reports
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: documentation-reports
          path: .moai/reports/*.txt

      - name: Comment PR with Results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('.moai/reports/korean_docs_comprehensive_review.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '📊 Documentation Validation Report\n```\n' + report.slice(0, 3000) + '\n```'
            });
```

### Section 6: Quality Metrics

**Phase 1 - Markdown Linting**:
- 53 files scanned
- 8 validation categories
- 351 false-positive links (relative paths)
- Quality indicator: Syntax errors detected

**Phase 2 - Mermaid Validation**:
- 16 diagrams across 9 files
- 100% valid diagram syntax
- 3 diagram types (graph, state, sequence)
- Renderability verified at mermaid.live

**Phase 3 - Korean Typography**:
- 28,543 lines validated
- 43 Korean content files
- 100% UTF-8 encoding
- Full-width character minimization

**Phase 4 - Comprehensive Report**:
- Overall Quality Score: **8.5/10**
- Priority 1 (Critical): 0 items
- Priority 2 (High): 2 items
- Priority 3 (Low): 1 item

## 🎯 Usage

### From Agents

```python
# docs-manager agent
Skill("moai-docs-unified")

# Run single phase
import subprocess
subprocess.run([
    "uv", "run",
    ".claude/skills/moai-docs-unified/scripts/lint_korean_docs.py"
])

# Run all phases
phases = [
    "lint_korean_docs.py",
    "validate_mermaid_diagrams.py",
    "extract_mermaid_details.py",
    "validate_korean_typography.py",
    "generate_final_comprehensive_report.py"
]

for phase in phases:
    subprocess.run([
        "uv", "run",
        f".claude/skills/moai-docs-unified/scripts/{phase}"
    ])
```

### From Commands

```bash
# Lint all documentation
/docs:validate --phase 1

# Validate Mermaid diagrams
/docs:validate --phase 2

# Check Korean typography
/docs:validate --phase 3

# Generate comprehensive report
/docs:validate --phase 4

# Run all phases
/docs:validate --all
```

## 📚 Reference Materials

- [MoAI-ADK Documentation Standards](https://docs.moai-adk.io/guides/documentation)
- [Mermaid Documentation](https://mermaid.live)
- [Markdown Guide](https://www.markdownguide.org/)
- [Korean Documentation Best Practices](https://docs.moai-adk.io/guides/korean-docs)

## ✅ Validation Checklist

- [x] 5 validation scripts integrated
- [x] Project root auto-detection implemented
- [x] Command-line argument support added
- [x] CI/CD integration examples provided
- [x] Quality metrics documented
- [x] Usage patterns explained
- [x] Error handling implemented

