---
name: moai-docs-unified
description: Enhanced docs unified with AI-powered features. Modular documentation validation framework
version: 1.0.0
modularized: true
allowed-tools:
  - Read
  - Bash
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - docs
  - unified
category_tier: special
---

## Quick Reference (30 seconds)

# moai-docs-unified

**Docs Unified**

> **Primary Agent**: doc-syncer  
> **Secondary Agents**: alfred  
> **Version**: 4.0.0  
> **Keywords**: docs, unified, cd, ci, test

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

#### Unified Framework Overview

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

---

#### Core Modules

This skill is modularized for optimal loading:

**Module 1: Validation Scripts** (`SKILL-scripts.md`)
- Script 1-5 detailed specifications
- Execution commands and parameters
- Output formats and locations
- Project root auto-detection

**Module 2: Integration Patterns** (`SKILL-integration.md`)
- Single script execution
- Complete validation pipeline
- CI/CD integration (GitHub Actions, pre-commit hooks)
- Makefile, NPM, Docker integration
- Best practices and error handling

---

### Level 2: Practical Implementation (Common Patterns)

#### Quick Usage

```bash
# Quick validation (2 phases)
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py

# Full validation (all 5 phases)
bash .claude/skills/moai-docs-unified/scripts/run_all_validations.sh

# Generate comprehensive report
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py
```

#### Validation Workflow

```
1. Phase 1: Markdown Linting
   └─ Check syntax, structure, links → lint_report_ko.txt

2. Phase 2: Mermaid Validation
   └─ Validate diagram types, syntax → mermaid_validation_report.txt

3. Phase 2.5: Mermaid Extraction
   └─ Extract full diagram code → mermaid_detail_report.txt

4. Phase 3: Korean Typography
   └─ Validate encoding, punctuation → korean_typography_report.txt

5. Phase 4: Comprehensive Report
   └─ Aggregate all phases → korean_docs_comprehensive_review.txt
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Best Practices Checklist

**Must-Have:**
- ✅ Run validation before every commit (pre-commit hook)
- ✅ Run on all pull requests (GitHub Actions)
- ✅ Store reports in `.moai/reports/`
- ✅ Fail CI on Priority 1 issues

**Recommended:**
- ✅ Daily scheduled validation runs
- ✅ Track quality metrics over time
- ✅ Archive reports from releases
- ✅ Provide actionable fix suggestions

**Security:**
- 🔒 Validate UTF-8 encoding (prevent injection)
- 🔒 Check external links (prevent broken refs)
- 🔒 Verify mermaid syntax (prevent XSS in rendered diagrams)

---

## 📚 Official References

**Primary Documentation:**
- [SKILL-scripts.md](/moai-docs-unified/SKILL-scripts.md) – All 5 validation scripts
- [SKILL-integration.md](/moai-docs-unified/SKILL-integration.md) – CI/CD & automation patterns

**Generated Reports** (`.moai/reports/`):
- `lint_report_ko.txt` – Markdown linting results
- `mermaid_validation_report.txt` – Diagram validation
- `mermaid_detail_report.txt` – Full diagram code extraction
- `korean_typography_report.txt` – Typography validation
- `korean_docs_comprehensive_review.txt` – Aggregated quality report

---

## 📈 Version History

**4.0.0** (2025-11-12)
- ✨ Modular structure with 2 sub-skills
- ✨ Enhanced Progressive Disclosure
- ✨ 5-phase validation framework
- ✨ CI/CD integration patterns
- ✨ Comprehensive automation examples

---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (doc-syncer)

---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-docs-generation") – Documentation generation

**Complementary Skills:**
- Skill("moai-project-documentation") – Project docs structure
- Skill("moai-git-flow") – Version control integration

**Next Steps:**
- After validation: Use Skill("moai-foundation-trust") for quality gates
- For publishing: Use Skill("moai-devops-docs-deployment")

---

**End of Skill** | Updated 2025-11-12