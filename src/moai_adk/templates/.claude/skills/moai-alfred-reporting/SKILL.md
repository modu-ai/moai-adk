---
name: moai-alfred-reporting
version: 1.0.0
created: 2025-11-02
updated: 2025-11-02
status: active
description: Guide report writing and output formatting for Alfred operations
keywords: ['reporting', 'output', 'formatting', 'documentation', 'communication']
allowed-tools:
  - Read
---

# Alfred Reporting Guide - Documentation & Communication

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-reporting |
| **Version** | 1.0.0 (2025-11-02) |
| **Status** | Active |
| **Tier** | Alfred |
| **Purpose** | Guide report writing and output formatting best practices |

---

## What It Does

Alfred uses consistent reporting patterns for task completion, documentation, and team communication.

**Key capabilities**:
- ✅ Screen output formatting (plain text, NO markdown syntax)
- ✅ Internal document formatting (markdown, structured sections)
- ✅ Sub-agent report templates
- ✅ Command completion reporting
- ✅ Language-aware output

---

## When to Use

**Automatic triggers**:
- Task completion → generate report summary
- Command execution done (*/alfred:*) → report results
- Documentation sync complete → report changes
- Quality verification passed → report metrics

**Manual reference**:
- Understanding report structure
- Formatting task completion summaries
- Learning best practices for documentation

---

## Output Format Rules

### Screen Output to User (Plain Text)

**When responding directly to user in chat/prompt:**

Use plain text format (NO markdown headers, tables, or special formatting):

```
Task Completion Summary:

Status: ✅ Complete
Duration: 2 hours
Files Modified: 5
Test Coverage: 95%

Key Achievements:
- Feature implemented
- Tests passing
- Documentation updated

Next Steps:
1. Review PR
2. Merge to main
3. Deploy
```

### Internal Documents (Markdown Format)

**When creating files in `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`:**

Use markdown format with proper structure:

```markdown
## 🎊 Task Completion Report

### Results
- ✅ Item 1 completed
- ✅ Item 2 completed

### Metrics
| Item | Result |
|------|--------|
| Coverage | 95% |
| Validation | ✅ Passed |

### Next Steps
1. Action item
2. Follow-up
```

---

## Standard Report Template

### Minimal Report (1-2 items)

```markdown
## ✅ [Task] Complete

Status: Done
Duration: X mins
Next: [Recommended action]
```

### Standard Report (3-5 items)

```markdown
## 🎊 [Task] Complete

### Results
- ✅ Item 1
- ✅ Item 2
- ✅ Item 3

### Metrics
| Metric | Value |
|--------|-------|
| Coverage | 95% |
| Speed | ✅ OK |

### Next Steps
1. Action 1
2. Action 2
```

### Comprehensive Report (6+ items)

```markdown
## 📊 [Task] Detailed Report

### Executive Summary
One-sentence overview

### Key Results
- ✅ Accomplished
- ✅ Accomplished
- ✅ Accomplished

### Quality Metrics
| Category | Result | Status |
|----------|--------|--------|
| Test Coverage | 95% | ✅ Excellent |
| Performance | Fast | ✅ Good |
| Validation | Passed | ✅ Complete |

### Files Modified
- `src/feature.py` - Core implementation
- `tests/test_feature.py` - Test suite
- `README.md` - Documentation

### @TAG Verification
- ✅ SPEC → CODE connection
- ✅ CODE → TEST connection
- ✅ TEST → DOC connection

### Blockers & Risks
- None identified

### Next Steps
1. Create PR
2. Request reviews
3. Merge when approved
```

---

## Sub-Agent Report Examples

### spec-builder Completion Report

```markdown
## 📋 SPEC Creation Complete

### Generated Artifacts
- ✅ `.moai/specs/SPEC-001/spec.md`
- ✅ `.moai/specs/SPEC-001/plan.md`
- ✅ `.moai/specs/SPEC-001/acceptance.md`

### Validation Results
- ✅ EARS format compliance: 100%
- ✅ @TAG chain created
- ✅ Requirement coverage: Complete

### Quality Gates
- ✅ Clarity: High
- ✅ Completeness: 100%
- ✅ Traceability: Set up

### Ready for Implementation
Next: Run `/alfred:2-run SPEC-001`
```

### tdd-implementer Completion Report

```markdown
## 🚀 Implementation Complete

### TDD Phases
| Phase | Status | Duration |
|-------|--------|----------|
| RED | ✅ Tests failing | 5 mins |
| GREEN | ✅ Implementation | 45 mins |
| REFACTOR | ✅ Code improvement | 15 mins |

### Implementation Files
- ✅ `src/feature.py` (234 LOC)
- ✅ `tests/test_feature.py` (156 LOC)

### Quality Metrics
- Test Coverage: 98%
- Linting: 0 issues
- Type Checking: 0 errors

### @TAG Verification
- ✅ SPEC → CODE links verified
- ✅ CODE → TEST links verified

### Ready for Sync
Next: Run `/alfred:3-sync`
```

### doc-syncer Completion Report

```markdown
## 📚 Documentation Sync Complete

### Updated Documents
- ✅ `README.md` - Usage examples added
- ✅ `.moai/docs/architecture.md` - Implementation details
- ✅ `CHANGELOG.md` - v0.8.0 entries
- ✅ `API.md` - Endpoint documentation

### @TAG Verification
- ✅ SPEC → CODE verified
- ✅ CODE → TEST verified
- ✅ TEST → DOC verified
- ✅ All chains complete

### Quality Checks
- ✅ No orphan TAGs
- ✅ All links valid
- ✅ Formatting consistent

### Ready to Merge
Next: Create PR and merge to main
```

---

## Key Principles

1. **Format Consistency**: Always use structured templates
2. **Language Awareness**: Match user's `conversation_language`
3. **Progress Visibility**: Include metrics and status indicators
4. **Actionable Output**: Always suggest next steps
5. **Traceability**: Include @TAG verification when relevant

---
