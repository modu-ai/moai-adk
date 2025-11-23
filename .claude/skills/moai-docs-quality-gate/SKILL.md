---
name: moai-docs-quality-gate
description: Documentation quality assurance (validation, linting, link checking, formatting)
version: 1.0.0
modularized: true
tags:
  - documentation
  - quality
  - validation
  - linting
  - enterprise
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-docs-quality-gate
**Domain**: Documentation Quality Assurance & Validation
**Freedom Level**: high
**Target Users**: Technical writers, QA engineers, developers
**Invocation**: Skill("moai-docs-quality-gate")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed rules)
**Last Updated**: 2025-11-24
**Modularized**: true
**Replaces**: moai-docs-validation, moai-docs-linting

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Comprehensive documentation quality gate with automated validation, linting, and compliance checking.

**6-Phase Quality Gate**:
```
1. Markdown Syntax Validation
   ↓
2. Link Integrity Checking
   ↓
3. Code Block Validation
   ↓
4. Typography & Formatting
   ↓
5. Style Guide Compliance
   ↓
6. Quality Report Generation
```

**Core Validations**:
- **Markdown Syntax**: Headers, lists, tables, code blocks
- **Links**: Broken links, anchor validity, external URL checking
- **Code Blocks**: Language declaration, syntax highlighting
- **Typography**: Trailing spaces, encoding, full-width characters
- **Style Compliance**: Heading case, terminology, formatting standards
- **Diagrams**: Mermaid/PlantUML syntax validation

**Quality Metrics**:
- **Pass Threshold**: ≥95% validation success rate
- **Broken Links**: 0 broken internal links tolerated
- **Code Blocks**: 100% must have language declarations
- **Header Violations**: 0 H1 duplicates, no level skipping
- **Overall Score**: 90+ for production-ready documentation

**When to Use**:
- Pre-commit hooks for documentation changes
- CI/CD pipelines for documentation validation
- Pre-publication quality checks
- Multi-language documentation validation
- Documentation refactoring quality assurance

---

## 📚 Core Patterns (5-10 minutes each)

### Pattern 1: Markdown Syntax Validation

**Concept**: Enforce correct markdown syntax and structure.

**Header Validation Rules**:
```yaml
H1_Rules:
  - Exactly 1 H1 per document
  - No duplicate H1s
  - H1 must be first header in document

Header_Hierarchy:
  - No level skipping (# → ## → ### ✓, # → ### ✗)
  - Proper nesting (parent-child relationships)
  - No duplicate headers on same level

Header_Content:
  - No emojis in headers (MoAI-ADK standard)
  - No trailing punctuation (periods, colons)
  - Title case for H1-H2, sentence case for H3-H6
```

**Code Block Validation**:
```yaml
Code_Block_Rules:
  - Language declaration required (```python, ```typescript, etc.)
  - Matching delimiters (opening ``` must have closing ```)
  - No inline code blocks (use single backticks for inline)
  - Syntax highlighting validation

Valid Examples:
  ```python
  def example():
      pass
  ```

  ```typescript
  const example = (): void => {};
  ```

Invalid Examples:
  ❌ ```  (missing language)
  ❌ ``` python (space before language)
  ❌ ```python
      # No closing delimiter
```

**List Validation**:
```yaml
List_Rules:
  - Consistent markers within same list (all -, all *, or all +)
  - Proper indentation (2 or 4 spaces for nested items)
  - No mixing ordered and unordered in nested contexts

Valid:
  - Item 1
  - Item 2
    - Nested item
    - Another nested

Invalid:
  ❌ - Item 1
     * Item 2  (inconsistent markers)
```

---

### Pattern 2: Link Integrity Checking

**Concept**: Validate all links (internal, external, anchors) to prevent broken references.

**Link Types & Validation**:
```yaml
Internal_Links:
  - File existence check
  - Relative path validation
  - .md extension required for markdown files
  - Case-sensitive path matching

External_Links:
  - HTTP status check (200 OK)
  - HTTPS protocol preferred
  - Timeout: 5 seconds per link
  - Retry: 2 attempts for transient failures

Anchor_Links:
  - Target heading existence
  - Anchor format validation (#heading-slug)
  - Case-sensitive anchor matching
```

**Execution Example**:
```bash
# Check all links in documentation
docs-validate-links --all

# Check specific file
docs-validate-links --file README.md

# External links only (with timeout)
docs-validate-links --external --timeout 10
```

**Output Format**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Link Validation Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Internal Links: 45/45 valid
❌ Broken Links Found:
   - README.md:15 → docs/missing-file.md (File not found)
   - GUIDE.md:42 → https://broken-link.com (404 Not Found)
   - API.md:78 → #nonexistent-anchor (Anchor not found)

⚠️  Warnings:
   - README.md:30 → http://example.com (HTTP instead of HTTPS)

Total Links: 50
Valid: 47 (94%)
Broken: 3 (6%)
```

---

### Pattern 3: Code Block Validation & Syntax Highlighting

**Concept**: Ensure all code examples are properly formatted with language declarations.

**Language Detection**:
```yaml
Required_Languages:
  - python, typescript, javascript, go, rust, java, kotlin
  - bash, shell, sh, zsh
  - yaml, json, toml, xml
  - sql, graphql
  - markdown, html, css

Language_Aliases:
  - js → javascript
  - ts → typescript
  - py → python
  - sh → bash
```

**Validation Pipeline**:
```bash
# Step 1: Detect code blocks without language
docs-validate-code --missing-language

# Step 2: Validate syntax highlighting
docs-validate-code --syntax-check

# Step 3: Auto-fix language declarations (if possible)
docs-validate-code --auto-fix
```

**Auto-Fix Example**:
```markdown
Before:
```
def example():
    pass
```

After (auto-fixed):
```python
def example():
    pass
```
```

---

### Pattern 4: Typography & Formatting Quality

**Concept**: Enforce consistent typography and formatting across documentation.

**Typography Rules**:
```yaml
Korean_Typography:
  - Proper spacing between Korean and English
  - Full-width punctuation for Korean text
  - UTF-8 encoding validation
  - No half-width Hangul

English_Typography:
  - No trailing whitespace
  - Consistent quotation marks (" vs ')
  - Oxford comma usage (optional)
  - No multiple consecutive spaces

Formatting:
  - Line length ≤ 100 characters (soft limit)
  - No tabs (use spaces)
  - Consistent indentation (2 or 4 spaces)
  - Single blank line between sections
```

**Validation Execution**:
```bash
# Korean typography validation
docs-validate-typography --lang ko

# English formatting
docs-validate-typography --lang en

# Fix common issues automatically
docs-validate-typography --auto-fix
```

---

### Pattern 5: Style Guide Compliance

**Concept**: Enforce project-specific documentation standards.

**MoAI-ADK Style Guide** (Example):
```yaml
Terminology:
  - Use "SPEC" (not "specification" or "spec")
  - Use "TDD" (not "test-driven development")
  - Use "Alfred" (not "alfred" or "Mr. Alfred")

Voice:
  - Active voice preferred
  - Second person ("you") for user-facing docs
  - Third person for API documentation

Prohibited_Terms:
  - "just" (minimizes complexity)
  - "simply" (condescending)
  - "obviously" (assumes knowledge)

Heading_Case:
  - Title Case for H1-H2
  - Sentence case for H3-H6
```

**Compliance Checking**:
```bash
# Check against style guide
docs-style-check --guide moai-adk

# Report violations
docs-style-check --report

# Auto-fix formatting issues
docs-style-check --fix
```

---

## 📖 Advanced Documentation

For detailed rules and validation strategies:

- **[modules/markdown-validation-rules.md](modules/markdown-validation-rules.md)** - Complete markdown syntax rules
- **[modules/link-checking-strategies.md](modules/link-checking-strategies.md)** - Link validation algorithms
- **[modules/typography-standards.md](modules/typography-standards.md)** - Multi-language typography rules
- **[modules/style-guide-enforcement.md](modules/style-guide-enforcement.md)** - Custom style guide configuration
- **[modules/quality-metrics-reference.md](modules/quality-metrics-reference.md)** - Quality scoring, CI/CD integration

---

## ✅ Best Practices

### DO
- ✅ Run validation before committing documentation changes
- ✅ Fix broken links immediately (0 tolerance)
- ✅ Declare language for all code blocks
- ✅ Enforce style guide compliance in CI/CD
- ✅ Generate quality reports for review
- ✅ Use auto-fix for common formatting issues
- ✅ Validate diagrams (Mermaid/PlantUML syntax)

### DON'T
- ❌ Skip validation steps
- ❌ Ignore broken links or warnings
- ❌ Mix heading case styles
- ❌ Use prohibited terminology
- ❌ Leave code blocks without language declarations
- ❌ Publish documentation below quality threshold

---

## 🔗 Works Well With

- `moai-docs-manager` - Documentation generation and management
- `moai-readme-expert` - Professional README generation with validation
- `moai-nextra-architecture` - Nextra documentation framework
- `moai-mermaid-diagram-expert` - Diagram syntax validation
- `moai-context7-integration` - Latest documentation best practices

---

## 📈 Integration Workflow

**Pre-Commit Hook**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run documentation validation
docs-validate --all --fail-on-error

# Check style compliance
docs-style-check --guide moai-adk
```

**CI/CD Pipeline**:
```yaml
name: Documentation Quality Gate
on: [push, pull_request]

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - name: Markdown Validation
        run: docs-validate --markdown
      - name: Link Checking
        run: docs-validate --links
      - name: Code Block Validation
        run: docs-validate --code-blocks
      - name: Typography Check
        run: docs-validate --typography
      - name: Style Compliance
        run: docs-style-check --fail-below 90
      - name: Generate Quality Report
        run: docs-validate --report --output quality-report.md
```

---

## 📊 Quality Metrics & Thresholds

**Production-Ready Thresholds**:
- Overall Quality Score: ≥90/100
- Broken Links: 0
- Code Blocks with Language: 100%
- Header Violations: 0
- Style Compliance: ≥95%
- Typography Issues: ≤5

**Scoring Formula**:
```
Quality Score = (
  Markdown Syntax (25%) +
  Link Integrity (30%) +
  Code Block Quality (20%) +
  Typography (15%) +
  Style Compliance (10%)
)
```

---

## 🔄 Changelog

- **v1.0.0** (2025-11-24): Initial unified skill combining docs-validation and docs-linting

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Modular Architecture**: SKILL.md + 5 modules
**Replaces**: moai-docs-validation (content validation), moai-docs-linting (syntax linting)
