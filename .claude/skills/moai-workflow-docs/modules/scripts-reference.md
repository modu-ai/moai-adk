---
name: moai-workflow-docs-scripts
description: Documentation validation scripts specifications and usage
---

## Documentation Validation Scripts

### Script 1: lint_korean_docs.py

**Purpose**: Comprehensive markdown validation for Korean documentation

**Location**: `.claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py`

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
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py

# With custom paths
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py \
  --path docs/src/ko \
  --output .moai/reports/lint_report_ko.txt
```

**Output**: `.moai/reports/lint_report_ko.txt` (8 validation categories)

---

### Script 2: validate_mermaid_diagrams.py

**Purpose**: Mermaid diagram type and syntax validation

**Location**: `.claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py`

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
uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py

# With custom paths
uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py \
  --path docs/src \
  --output .moai/reports/mermaid_validation_report.txt
```

**Output**: `.moai/reports/mermaid_validation_report.txt` (16 diagrams, 100% valid)

---

### Script 3: extract_mermaid_details.py

**Purpose**: Extract and document all Mermaid diagram code with rendering guide

**Location**: `.claude/skills/moai-workflow-docs/scripts/extract_mermaid_details.py`

**Key Features**:
- Extract all mermaid blocks from documentation
- Show diagram type and line count
- Provide full code preview
- Generate rendering test guide (mermaid.live)

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-workflow-docs/scripts/extract_mermaid_details.py

# With custom paths
uv run .claude/skills/moai-workflow-docs/scripts/extract_mermaid_details.py \
  --path docs/src \
  --output .moai/reports/mermaid_detail_report.txt
```

**Output**: `.moai/reports/mermaid_detail_report.txt` (full diagram code + test guide)

---

### Script 4: validate_korean_typography.py

**Purpose**: Korean-specific typography and encoding validation

**Location**: `.claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py`

**Key Validations**:
- UTF-8 Encoding verification
- Full-width character detection (U+3000, ＜, ＞, etc.)
- Punctuation consistency (. vs。, , vs、)
- Spacing rules (Korean-English boundaries)
- Character statistics per file

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py

# With custom paths
uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py \
  --path docs/src \
  --output .moai/reports/korean_typography_report.txt
```

**Output**: `.moai/reports/korean_typography_report.txt` (28,543 lines validated, 43 files)

---

### Script 5: generate_final_comprehensive_report.py

**Purpose**: Aggregate all validation phases into prioritized quality report

**Location**: `.claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py`

**Report Structure**:
1. Executive Summary (8.5/10 quality score)
2. Phase Results (Priority 1/2/3 items)
3. Recommendations (Done/In Progress/TODO)
4. Action Items (Immediate/Short-term/Long-term)
5. Generated Report Files (all 4 phases)

**Execution**:
```bash
# From project root
uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py

# With custom report directory
uv run .claude/skills/moai-workflow-docs/scripts/generate_final_comprehensive_report.py \
  --report-dir .moai/reports \
  --output .moai/reports/korean_docs_comprehensive_review.txt
```

**Output**: `.moai/reports/korean_docs_comprehensive_review.txt` (aggregated report)

---

### Project Root Auto-Detection

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

---

**End of Module** | moai-workflow-docs-scripts
