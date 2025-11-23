---
name: moai-docs-manager
description: Documentation management (auto-generation, tooling, consistency, validation)
version: 1.0.0
modularized: true
tags:
  - documentation
  - enterprise
  - generation
  - validation
  - toolkit
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-docs-manager
**Domain**: Documentation Management & Quality Assurance
**Freedom Level**: high
**Target Users**: Technical writers, developers, documentation teams
**Invocation**: Skill("moai-docs-manager")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed patterns)
**Last Updated**: 2025-11-24
**Modularized**: true
**Replaces**: moai-docs-generation, moai-docs-toolkit, moai-docs-unified

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Comprehensive documentation lifecycle management - from generation to validation and publishing.

**Core Capabilities**:
- **Auto-Generation**: Generate README, API docs, guides from code and templates
- **Validation**: Markdown linting, link checking, diagram syntax validation
- **Consistency**: Enforce style guides, typography rules, template compliance
- **Toolkit**: CLI tools, scripts, templates for documentation workflows
- **Multi-Language**: Support for Korean, English, Japanese, Chinese documentation

**5-Phase Documentation Workflow**:
```
1. Generation (from templates/code)
   ↓
2. Validation (markdown/links/diagrams)
   ↓
3. Consistency (style/typography)
   ↓
4. Review (automated checks)
   ↓
5. Publishing (deploy to GitHub Pages, Nextra, etc.)
```

**When to Use**:
- Generate documentation from code annotations
- Validate existing documentation for errors
- Enforce documentation standards across projects
- Automate documentation workflows
- Maintain multi-language documentation

---

## 📚 Core Patterns (5-10 minutes each)

### Pattern 1: Auto-Documentation Generation

**Concept**: Generate documentation automatically from code, templates, and SPEC files.

**Supported Generators**:
- **README.md**: Project overview from template
- **API Documentation**: From TypeScript/Python docstrings
- **User Guides**: From SPEC acceptance criteria
- **Changelog**: From git commits and PRs
- **Architecture Diagrams**: From codebase structure

**Example: README Generation**:
```bash
# Generate README from template
docs-generate --type readme --template enterprise

# Generate API docs from TypeScript code
docs-generate --type api --language typescript --output docs/api

# Generate guide from SPEC
docs-generate --type guide --spec SPEC-001 --output docs/guides/feature-001.md
```

**Template Structure**:
```markdown
# {project_name}

{project_description}

## Features
{features_list}

## Installation
{installation_steps}

## Usage
{usage_examples}

## API Reference
{api_reference}
```

---

### Pattern 2: Documentation Validation Pipeline

**Concept**: Multi-stage validation to catch errors before publication.

**5 Validation Phases**:
1. **Markdown Linting**: Syntax, structure, heading hierarchy
2. **Link Validation**: Internal links, external URLs, file existence
3. **Diagram Syntax**: Mermaid, PlantUML, diagrams.net validation
4. **Typography**: Language-specific rules (Korean spacing, etc.)
5. **Quality Report**: Aggregated findings with recommendations

**Execution**:
```bash
# Complete validation pipeline
docs-validate --all

# Individual phases
docs-validate --markdown
docs-validate --links
docs-validate --diagrams
docs-validate --typography --lang ko
```

**Example Validation Output**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Documentation Quality Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Markdown Linting: 45/50 files passed
❌ Link Validation: 3 broken links found
   - README.md:15 → docs/missing-file.md
   - GUIDE.md:42 → https://broken-link.com
⚠️  Diagram Syntax: 1 warning
   - ARCHITECTURE.md:78 → Invalid Mermaid syntax
✅ Typography: All Korean spacing rules passed

Overall Score: 88/100 (Good)
```

---

### Pattern 3: Consistency Enforcement

**Concept**: Enforce project-specific documentation standards.

**Consistency Checks**:
- **Style Guide Compliance**: Heading case, voice, terminology
- **Template Structure**: Required sections present
- **Code Block Languages**: Declared for all code examples
- **Formatting**: Consistent list markers, table alignment
- **Naming Conventions**: File naming patterns

**Configuration (`.moai/docs-standards.yml`)**:
```yaml
standards:
  heading_case: title  # title, sentence, lowercase
  list_marker: "-"     # -, *, +
  code_blocks_language_required: true
  max_line_length: 100
  required_sections:
    - Installation
    - Usage
    - API Reference
    - License
  prohibited_words:
    - "just"           # Weak language
    - "simply"         # Condescending
    - "obviously"      # Assumes knowledge
```

**Usage**:
```bash
# Check consistency against standards
docs-consistency --check

# Auto-fix formatting issues
docs-consistency --fix
```

---

### Pattern 4: Multi-Language Documentation

**Concept**: Manage documentation across multiple languages with consistency.

**Directory Structure**:
```
docs/
├── en/
│   ├── README.md
│   ├── guides/
│   └── api/
├── ko/
│   ├── README.md
│   ├── guides/
│   └── api/
├── ja/
│   └── README.md
└── zh/
    └── README.md
```

**Synchronization Workflow**:
```bash
# Generate initial translation structure
docs-i18n init --languages ko,ja,zh

# Check translation coverage
docs-i18n coverage
# Output: en: 100%, ko: 85%, ja: 60%, zh: 40%

# Identify missing translations
docs-i18n missing --lang ko
# Output: Missing: guides/advanced.md, api/webhooks.md
```

**Language-Specific Validation**:
```bash
# Korean typography validation
docs-validate --typography --lang ko

# Japanese validation (character encoding, proper particles)
docs-validate --typography --lang ja
```

---

### Pattern 5: Documentation Toolkit (CLI & Scripts)

**Concept**: Automation scripts for common documentation tasks.

**Available Tools**:
- `docs-generate`: Generate documentation from templates
- `docs-validate`: Run validation pipelines
- `docs-consistency`: Enforce style standards
- `docs-i18n`: Manage multi-language docs
- `docs-publish`: Deploy to GitHub Pages, Nextra, etc.

**Example Automation (GitHub Actions)**:
```yaml
name: Documentation CI
on: [push, pull_request]

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate Documentation
        run: |
          docs-validate --all
          docs-consistency --check
      - name: Generate Quality Report
        run: docs-validate --report --output docs-quality-report.md
```

---

## 📖 Advanced Documentation

For detailed patterns and implementation strategies:

- **[modules/generation-templates.md](modules/generation-templates.md)** - Template library, scaffolding, auto-generation
- **[modules/validation-pipeline.md](modules/validation-pipeline.md)** - Multi-stage validation, error detection
- **[modules/consistency-standards.md](modules/consistency-standards.md)** - Style guides, enforcement rules
- **[modules/multi-language-support.md](modules/multi-language-support.md)** - i18n strategies, translation workflows
- **[modules/toolkit-reference.md](modules/toolkit-reference.md)** - CLI reference, automation scripts

---

## ✅ Best Practices

### DO
- ✅ Generate documentation from templates for consistency
- ✅ Validate before publishing (run full pipeline)
- ✅ Enforce style standards across all documentation
- ✅ Use meaningful code block language declarations
- ✅ Test all links (internal and external)
- ✅ Automate documentation workflows (CI/CD)
- ✅ Keep translations synchronized

### DON'T
- ❌ Manually write boilerplate (use templates)
- ❌ Skip validation steps
- ❌ Ignore broken links or diagram errors
- ❌ Mix documentation styles within a project
- ❌ Forget to update documentation when code changes
- ❌ Leave translations incomplete

---

## 🔗 Works Well With

- `moai-docs-quality-gate` - Documentation validation and linting
- `moai-project-documentation` - Project-level documentation orchestration
- `moai-readme-expert` - Professional README generation
- `moai-context7-integration` - Latest documentation patterns
- `moai-nextra-architecture` - Nextra documentation framework

---

## 📈 Integration Workflow

**Complete Documentation Workflow**:
```
1. Generate documentation (Pattern 1)
   ↓
2. Validate markdown/links/diagrams (Pattern 2)
   ↓
3. Enforce consistency (Pattern 3)
   ↓
4. Synchronize translations (Pattern 4)
   ↓
5. Publish to platform (Pattern 5)
```

**CI/CD Integration**:
```
Pre-commit: Validate markdown syntax
PR Review: Check links, diagrams, consistency
Merge: Generate API docs, update changelogs
Deploy: Publish to GitHub Pages, Nextra
```

---

## 📊 Success Metrics

- **Generation Coverage**: 100% of code has auto-generated docs
- **Validation Pass Rate**: ≥95% files pass all checks
- **Broken Links**: 0 broken internal links
- **Translation Coverage**: ≥80% for primary languages
- **Style Compliance**: 100% adherence to standards
- **Documentation Freshness**: ≤24 hours lag from code changes

---

## 🔄 Changelog

- **v1.0.0** (2025-11-24): Initial unified skill combining docs-generation, docs-toolkit, and docs-unified

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Modular Architecture**: SKILL.md + 5 modules
**Replaces**: moai-docs-generation (generation), moai-docs-toolkit (tooling), moai-docs-unified (validation/consistency)
