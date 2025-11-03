---
name: moai-alfred-doc-management
description: Internal documentation placement rules, forbidden patterns, and sub-agent output guidelines
tier: alfred
freedom: low
tags: [documentation, file-locations, conventions, policies, management]
---

# Document Management Rules

**CRITICAL**: Alfred and all Sub-agents MUST follow these document placement rules.

## Allowed Document Locations

| Document Type           | Location              | Examples                             |
| ----------------------- | --------------------- | ------------------------------------ |
| **Internal Guides**     | `.moai/docs/`         | Implementation guides, strategy docs |
| **Exploration Reports** | `.moai/docs/`         | Analysis, investigation results      |
| **SPEC Documents**      | `.moai/specs/SPEC-*/` | spec.md, plan.md, acceptance.md      |
| **Sync Reports**        | `.moai/reports/`      | Sync analysis, tag validation        |
| **Technical Analysis**  | `.moai/analysis/`     | Architecture studies, optimization   |
| **Memory Files**        | `.moai/memory/`       | Session state only (runtime data)    |
| **Knowledge Base**      | `.claude/skills/moai-alfred-*` | Alfred workflow guidance (on-demand) |

## FORBIDDEN: Root Directory

**NEVER proactively create documentation in project root** unless explicitly requested by user:

- ❌ `IMPLEMENTATION_GUIDE.md`
- ❌ `EXPLORATION_REPORT.md`
- ❌ `*_ANALYSIS.md`
- ❌ `*_GUIDE.md`
- ❌ `*_REPORT.md`

**Exceptions** (ONLY these files allowed in root):

- ✅ `README.md` - Official user documentation
- ✅ `CHANGELOG.md` - Version history
- ✅ `CONTRIBUTING.md` - Contribution guidelines
- ✅ `LICENSE` - License file

## Decision Tree for Document Creation

```
Need to create a .md file?
    ↓
Is it user-facing official documentation?
    ├─ YES → Root (README.md, CHANGELOG.md only)
    └─ NO → Is it internal to Alfred/workflow?
             ├─ YES → Check type:
             │    ├─ SPEC-related → .moai/specs/SPEC-*/
             │    ├─ Sync report → .moai/reports/
             │    ├─ Analysis → .moai/analysis/
             │    └─ Guide/Strategy → .moai/docs/
             └─ NO → Ask user explicitly before creating
```

## Document Naming Convention

**Internal documents in `.moai/docs/`**:

- `implementation-{SPEC-ID}.md` - Implementation guides
- `exploration-{topic}.md` - Exploration/analysis reports
- `strategy-{topic}.md` - Strategic planning documents
- `guide-{topic}.md` - How-to guides for Alfred use

## Sub-agent Output Guidelines

| Sub-agent              | Default Output Location | Document Type            |
| ---------------------- | ----------------------- | ------------------------ |
| implementation-planner | `.moai/docs/`           | implementation-{SPEC}.md |
| Explore                | `.moai/docs/`           | exploration-{topic}.md   |
| Plan                   | `.moai/docs/`           | strategy-{topic}.md      |
| doc-syncer             | `.moai/reports/`        | sync-report-{type}.md    |
| tag-agent              | `.moai/reports/`        | tag-validation-{date}.md |

## Directory Structure

**Expected MoAI directory layout**:

```
.moai/
├── config.json              # Project configuration
├── docs/                    # Internal documentation
│   ├── implementation-*.md
│   ├── exploration-*.md
│   ├── strategy-*.md
│   └── guide-*.md
├── specs/                   # SPEC documents
│   ├── SPEC-ID-001/
│   │   ├── spec.md
│   │   ├── plan.md
│   │   └── acceptance.md
│   └── SPEC-ID-002/
├── reports/                 # Generated reports
│   ├── sync-report-*.md
│   └── tag-validation-*.md
├── analysis/                # Technical analysis
│   └── *-analysis.md
└── memory/                  # Session state (runtime only)
    └── session-state.json
```

## Enforcement Rules

### ❌ Violations to Prevent

1. **No root-level generated documents**
   - Creating `IMPLEMENTATION_REPORT.md`, `ANALYSIS.md` in project root
   - **Fix**: Place in `.moai/docs/`, `.moai/analysis/`, or `.moai/reports/` instead

2. **No random file locations**
   - Creating SPEC documents outside `.moai/specs/SPEC-*/`
   - **Fix**: Use `.moai/specs/SPEC-ID-XXX/spec.md` structure

3. **No sync/analysis reports in docs/**
   - Placing sync reports in `.moai/docs/` (wrong)
   - **Fix**: Use `.moai/reports/sync-report-*.md` instead

### ✅ Correct Patterns

**When to ask user**:
- "Would you like me to create a detailed implementation guide in `.moai/docs/`?"
- "I can generate a sync report in `.moai/reports/`. Proceed?"

**When to auto-create**:
- SPEC documents during `/alfred:1-plan` → `.moai/specs/SPEC-*/`
- Sync reports during `/alfred:3-sync` → `.moai/reports/`
- Implementation guides during `/alfred:2-run` → `.moai/docs/` (if documenting approach)

## Documentation Lifecycle

| Phase | Document Type | Location | Creator |
|-------|---------------|----------|---------|
| SPEC | spec.md, plan.md, acceptance.md | `.moai/specs/SPEC-*` | spec-builder |
| BUILD | implementation guide (optional) | `.moai/docs/` | tdd-implementer (if needed) |
| SYNC | sync-report-*.md | `.moai/reports/` | doc-syncer |
| ARCHIVE | README.md, CHANGELOG.md | Root (public) | Curated by project owner |
