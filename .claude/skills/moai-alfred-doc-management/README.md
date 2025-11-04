# Document Management Rules

Internal documentation placement rules, forbidden patterns, and sub-agent output guidelines for Alfred and all sub-agents.

## Purpose

Ensures consistent and organized documentation structure across all MoAI projects by enforcing strict placement rules and naming conventions.

## Core Principle

**CRITICAL**: Alfred and all Sub-agents MUST follow these document placement rules.

## Key Locations

| Document Type | Location | Examples |
|---------------|----------|----------|
| **Internal Guides** | `.moai/docs/` | Implementation guides, strategy docs |
| **SPEC Documents** | `.moai/specs/SPEC-*/` | spec.md, plan.md, acceptance.md |
| **Sync Reports** | `.moai/reports/` | Sync analysis, tag validation |
| **Technical Analysis** | `.moai/analysis/` | Architecture studies, optimization |

## Forbidden Root Files

**NEVER create in project root:**
- ❌ `IMPLEMENTATION_GUIDE.md`
- ❌ `*_ANALYSIS.md`
- ❌ `*_REPORT.md`
- ❌ `*_GUIDE.md`

**Allowed root files only:**
- ✅ `README.md`
- ✅ `CHANGELOG.md`
- ✅ `CONTRIBUTING.md`
- ✅ `LICENSE`

## Quick Decision Tree

```
Need to create .md file?
    ↓
User-facing official docs?
    ├─ YES → Root (README.md, CHANGELOG.md only)
    └─ NO → Internal to Alfred/workflow?
             ├─ YES → Check type:
             │    ├─ SPEC → .moai/specs/SPEC-*/
             │    ├─ Sync → .moai/reports/
             │    ├─ Analysis → .moai/analysis/
             │    └─ Guide/Strategy → .moai/docs/
             └─ NO → Ask user explicitly
```

## File Naming Patterns

- **Implementation**: `implementation-{SPEC-ID}.md`
- **Exploration**: `exploration-{topic}.md`
- **Strategy**: `strategy-{topic}.md`
- **Sync Report**: `sync-report-{YYYYMMDD}.md`
- **Analysis**: `{topic}-analysis.md`

## Usage

```bash
# When creating documents, always check:
Skill("moai-alfred-doc-management")

# This ensures compliance with:
# - Proper file locations
# - Correct naming conventions
# - Required user approvals
# - Integration with Alfred workflows
```

## Files Structure

- `SKILL.md` - Complete rules and guidelines
- `examples.md` - Practical implementation examples
- `reference.md` - Technical reference and matrices
- `README.md` - This quick overview

## Integration

Works with:
- All `/alfred:*` commands
- Sub-agents (implementation-planner, Explore, Plan, etc.)
- SPEC creation workflow
- Sync reporting system
