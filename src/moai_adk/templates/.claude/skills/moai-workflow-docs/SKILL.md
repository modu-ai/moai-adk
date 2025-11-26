---
name: moai-workflow-docs
aliases: [moai-workflow-docs]
category: workflow
description: Enhanced documentation unified validation with AI-powered features
version: 2.0.0
modularized: true
tags:
  - workflow
  - enterprise
  - unified
  - development
updated: 2025-11-27
status: active
deprecated_names:
  moai-workflow-docs:
    deprecated_in: v0.32.0
    remove_in: v0.35.0
    message: "Use moai-workflow-docs instead"
---

## Quick Reference (30 seconds)

**Purpose**: Comprehensive documentation validation framework with 5 specialized phases.

**Core Phases**:
1. **Markdown Linting** - Syntax, structure, links validation
2. **Mermaid Diagrams** - Diagram syntax and type checking
3. **Mermaid Details** - Code extraction and rendering guide
4. **Korean Typography** - UTF-8, spacing, encoding validation
5. **Report Generation** - Aggregated quality report

**Key Benefits**:
- Catch errors before publication
- Multi-language support (Korean, English, Japanese, Chinese)
- Diagram syntax validation
- Typography consistency
- Actionable recommendations

---


## Implementation Guide (5 minutes)

### Features

- Unified documentation generation for technical projects
- README, API docs, architecture guides, and deployment docs
- CommonMark compliance with proper formatting standards
- Automated cross-referencing and navigation
- Multi-language documentation support

### When to Use

- Generating project documentation from code and specifications
- Creating API reference documentation automatically
- Building architecture decision records (ADRs)
- Producing deployment guides and runbooks
- Synchronizing documentation with code changes

### Core Patterns

**Pattern 1: Documentation Structure**
```
docs/
├── README.md (project overview)
├── API.md (API reference)
├── ARCHITECTURE.md (system design)
├── DEPLOYMENT.md (deployment guide)
└── CONTRIBUTING.md (contribution guide)
```

**Pattern 2: Auto-generated API Docs**
```python
# Extract from code comments and type hints
def generate_api_docs(source_files):
    1. Parse docstrings and annotations
    2. Generate markdown tables for parameters/returns
    3. Include code examples from tests
    4. Cross-reference related endpoints
    5. Validate all links and references
```

**Pattern 3: Documentation Sync**
1. Detect code changes via git diff
2. Identify affected documentation sections
3. Update docs automatically or prompt for review
4. Validate documentation completeness
5. Generate changelog entries

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Documentation Validation Pipeline

**Key Concept**: Run validation scripts in sequence to catch multiple error types

**Pipeline Flow**:
1. Run markdown linting on documentation files
2. Validate all Mermaid diagrams for syntax
3. Extract Mermaid diagrams for review
4. Check Korean typography (if applicable)
5. Generate comprehensive report

**Basic Execution**:
```bash
# Run complete validation
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py
uv run .claude/skills/moai-workflow-docs/scripts/validate_korean_typography.py
uv run .claude/skills/moai-workflow-docs/scripts/generate_quality_report.py
```

### Pattern 2: Markdown Structure Validation

**Key Concept**: Ensure consistent markdown structure and formatting

**Common Validations**:
- **Headers**: H1 unique, proper nesting (H1→H2→H3)
- **Code blocks**: Language declared, matching delimiters
- **Links**: Relative paths valid, files exist, HTTPS protocol
- **Lists**: Consistent markers (-, *, +), proper indentation
- **Tables**: Column count consistent, alignment markers

**Example Issues**:
```
❌ Missing language in code block: ```
✅ Correct syntax: ```python

❌ Invalid link: [text](../docs/file)
✅ Correct syntax: [text](../docs/file.md)

❌ Inconsistent list markers: - item1, * item2
✅ Consistent: - item1, - item2
```

### Pattern 3: Mermaid Diagram Validation

**Key Concept**: Validate diagram syntax and type compatibility

**Supported Types**:
- `graph TD/BT/LR/RL` - Flowcharts (top-down, bottom-up, left-right, right-left)
- `stateDiagram-v2` - State machines
- `sequenceDiagram` - Sequence diagrams
- `classDiagram` - Class structures
- `erDiagram` - Entity relationship diagrams
- `gantt` - Gantt charts (timelines)

**Validation Checks**:
- Diagram type recognized
- Configuration block valid
- Node/edge relationships valid
- Syntax errors detected
- Complexity metrics

### Pattern 4: Korean Typography Rules

**Key Concept**: Maintain Korean language best practices

**Validation Rules**:
- No full-width ASCII characters (ａ-ｚ should be a-z)
- Proper spacing around parentheses: `（한글）` vs `(한글)`
- UTF-8 encoding (no broken characters)
- Consistent punctuation (，vs, 、vs..)
- Proper use of Hangul vs Hanja (한글 vs 한漢字)

### Pattern 5: Quality Report Generation

**Key Concept**: Aggregate validation results with actionable recommendations

**Report Contents**:
- Summary statistics (files, issues, severity)
- Issue categorization (errors vs warnings)
- Priority ranking (critical, high, medium, low)
- Specific file locations and line numbers
- Recommended fixes

---

## Advanced Documentation

This Skill uses Progressive Disclosure. For detailed implementation:

- **[modules/validation-scripts.md](modules/validation-scripts.md)** - Complete script specifications
- **[modules/execution-guide.md](modules/execution-guide.md)** - How to run validations
- **[modules/troubleshooting.md](modules/troubleshooting.md)** - Common issues and fixes
- **[modules/reference.md](modules/reference.md)** - API reference and configuration
- **[modules/scripts-reference.md](modules/scripts-reference.md)** - Script API reference
- **[modules/integration-patterns.md](modules/integration-patterns.md)** - Integration patterns and examples

---

## 🔧 Common Use Cases

### Use Case 1: CI/CD Integration

Run validation on every commit:

```yaml
# .github/workflows/docs-validation.yml
on: [push, pull_request]
jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate documentation
        run: |
          uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
          uv run .claude/skills/moai-workflow-docs/scripts/validate_mermaid_diagrams.py
```

### Use Case 2: Pre-Commit Hook

Validate docs before committing:

```bash
#!/bin/bash
# .git/hooks/pre-commit
uv run .claude/skills/moai-workflow-docs/scripts/lint_korean_docs.py
if [ $? -ne 0 ]; then
  echo "Documentation validation failed"
  exit 1
fi
```

### Use Case 3: Documentation Review

Generate report for review team:

```bash
uv run .claude/skills/moai-workflow-docs/scripts/generate_quality_report.py \
  --path docs/src \
  --output .moai/reports/review_report.txt \
  --format detailed
```

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-docs-generation") - Generate documentation automatically
- Skill("moai-docs-toolkit") - Document manipulation and conversion
- Skill("moai-cc-claude-md") - Markdown formatting standards

**Typical Workflow**:
1. Use moai-docs-generation to create documentation
2. Use this Skill (moai-workflow-docs) to validate output
3. Use moai-docs-toolkit for additional processing

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure pattern
- 📚 Scripts moved to modules/ for clarity
- ✨ Core patterns highlighted in SKILL.md
- ✨ Added CI/CD and hook integration examples

**1.0.0** (2025-11-12)
- ✨ Markdown linting with 8 validation categories
- ✨ Mermaid diagram validation
- ✨ Korean typography validation
- ✨ Comprehensive quality reporting

---

## Works Well With

**Agents**:
- **workflow-docs** - Documentation generation workflow
- **core-quality** - Quality assurance and validation
- **workflow-spec** - Specification documentation

**Skills**:
- **moai-docs-generation** - Generate documentation automatically
- **moai-docs-toolkit** - Document manipulation and conversion
- **moai-cc-claude-md** - Markdown formatting standards
- **moai-library-mermaid** - Advanced diagram validation
- **moai-library-nextra** - Nextra-based documentation architecture

**Commands**:
- `/moai:3-sync` - Documentation synchronization
- `/moai:9-feedback` - Documentation improvement feedback

---

**Maintained by**: alfred
**Domain**: Documentation & Quality Assurance
**Generated with**: MoAI-ADK Skill Factory
