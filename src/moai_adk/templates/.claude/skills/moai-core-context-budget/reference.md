# Context Budget Management Guide

## Overview

- **Total context budget**: 200K tokens (Claude Sonnet 4.5)
- **Session start overhead**: ~20K tokens (core files: CLAUDE.md, agents, commands, skills, memory)
- **Available for work**: ~180K tokens
- **Warning threshold**: 180K tokens (90% usage)
- **Critical threshold**: 190K tokens (95% usage)

## Progressive Loading Strategy

### Load Priority Tiers

#### Tier 1: CRITICAL (Always Load) — ~10K tokens

**Must load at session start**:
- `.moai/config/config.json` (project configuration)
- `.moai/project/product.md` (business requirements)
- `.moai/project/structure.md` (architecture design)
- `.moai/project/tech.md` (technology stack)
- `CLAUDE.md` (project directives)

**Why**: These files define project scope, constraints, and workflow rules.

---

#### Tier 2: HIGH (Load When Needed) — ~15K tokens

**Load when working on specific SPEC**:
- `.moai/specs/SPEC-{ID}/spec.md` (requirements)
- `.moai/specs/SPEC-{ID}/plan.md` (implementation plan)
- `.moai/specs/SPEC-{ID}/acceptance.md` (acceptance criteria)

**Load when quality verification needed**:
- `.moai/memory/development-guide.md` (TRUST principles)
- `.moai/memory/spec-metadata.md` (SPEC standard)

**Why**: Context-specific guidance needed only for active work.

---

#### Tier 3: MEDIUM (Load On-Demand During Execution) — ~15K tokens

**Load selectively during TDD implementation**:
- Existing code references (src/ files related to current SPEC)
- Test files (tests/ files related to current SPEC)
- Related SPEC documents (dependencies, blockers)

**Load during document sync**:
- `.moai/reports/sync-report-latest.md` (previous sync state)
- Documentation files (README, architecture docs)

**Why**: Needed only when actively modifying or referencing.

---

#### Tier 4: LOW (Load If Context Allows) — ~10K tokens

**Optional context enhancements**:
- Historical context (git log, SPEC history)
- Examples and templates
- Related Skills (moai-foundation-*, moai-core-*)
- Agent instructions (spec-builder, tdd-implementer)

**Why**: Helpful but not essential for task completion.

---

## Agent-Specific Guidelines

### spec-builder

**Phase 1: Planning** (20-30K tokens)
- **Tier 1**: Always load (product.md, structure.md, tech.md)
- **Tier 2**: Load EARS standard, existing SPEC list
- **Tier 3**: JIT load related SPEC examples

**Phase 2: SPEC Creation** (30-40K tokens)
- **Tier 1**: Maintain project documents
- **Tier 2**: Load SPEC metadata standard
- **Tier 3**: Load existing SPEC documents for reference

---

### tdd-implementer

**Phase 1: Analysis** (25-35K tokens)
- **Tier 1**: Load SPEC documents (spec.md, plan.md, acceptance.md)
- **Tier 2**: Load TDD workflow, development guide
- **Tier 3**: JIT load existing test files

**Phase 2: TDD Cycle** (40-60K tokens)
- **Tier 1**: Maintain SPEC documents
- **Tier 2**: Load implementation files (src/)
- **Tier 3**: Load test files (tests/)
- **Tier 4**: Load refactoring guidelines when needed

---

### doc-syncer

**Phase 1: Analysis** (30-40K tokens)
- **Tier 1**: Load project documents
- **Tier 2**: Load TAG scanning results
- **Tier 3**: JIT load documentation files

**Phase 2: Synchronization** (50-70K tokens)
- **Tier 1**: Maintain project documents
- **Tier 2**: Load SPEC documents
- **Tier 3**: Load documentation files (README, architecture)
- **Tier 4**: Load previous sync report for comparison

---

### implementation-planner

**Phase 1: SPEC Analysis** (20-30K tokens)
- **Tier 1**: Load SPEC documents
- **Tier 2**: Load project documents (architecture constraints)
- **Tier 3**: JIT load existing code structure

**Phase 2: Plan Creation** (30-40K tokens)
- **Tier 1**: Maintain SPEC documents
- **Tier 2**: Load development guide, TAG system
- **Tier 3**: Load related implementation examples

---

## Context Monitoring

### Session Context Tracking

**File**: `.moai/memory/session-context.json`

```json
{
  "session_id": "2025-11-02-001",
  "started_at": "2025-11-02T10:00:00Z",
  "context_usage": {
    "current_tokens": 45000,
    "warning_threshold": 180000,
    "critical_threshold": 190000,
    "percent_used": 22.5
  },
  "loaded_files": [
    ".moai/config.json",
    ".moai/project/product.md",
    ".moai/specs/SPEC-AUTH-001/spec.md",
    "src/auth/service.py",
    "tests/auth/test_service.py"
  ],
  "current_command": "alfred:2-run",
  "current_spec": "SPEC-AUTH-001"
}
```

### Warning Indicators

**90% threshold (180K tokens)**:
```
⚠️ Context budget warning: 180K/200K tokens used (90%)
Recommendation: Start new session with /clear for optimal performance
```

**95% threshold (190K tokens)**:
```
🚨 Context budget critical: 190K/200K tokens used (95%)
Action required: Start new session immediately with /clear
```

---

## Optimization Tips

### 1. Use Skill() for Large Content

**Instead of loading entire files**:
```python
# ❌ BAD: Load entire 5K token file
Read(".moai/memory/development-guide.md")

# ✅ GOOD: Load specific Skill (~500 tokens)
Skill("moai-foundation-trust")
```

**Benefits**:
- Skill provides targeted guidance (~500 tokens)
- Full file includes examples, history, unused sections (~5K tokens)
- 10x efficiency improvement

---

### 2. Read Files Selectively

**Use grep/glob to get specific lines**:
```python
# ❌ BAD: Load entire codebase
Glob("src/**/*.py")

# ✅ GOOD: Load only related files
Glob("src/auth/**/*.py")

# ✅ BETTER: Search for specific patterns
```

---

### 3. Batch Related Operations

**Parallelize to avoid multiple passes**:
```python
# ❌ BAD: Sequential reads (3 separate responses)
Read(".moai/project/product.md")
Read(".moai/project/structure.md")
Read(".moai/project/tech.md")

# ✅ GOOD: Parallel reads (1 response)
# In a SINGLE response:
Read(".moai/project/product.md")
Read(".moai/project/structure.md")
Read(".moai/project/tech.md")
```

**Performance impact**: 1.5-2x faster execution

---

### 4. Clean Up Session Memory

**Remove stale context between phases**:
- After completing `/alfred:1-plan`, start new session with `/clear` before `/alfred:2-run`
- After completing `/alfred:2-run`, start new session with `/clear` before `/alfred:3-sync`

**Why**: Each phase has different context needs:
- Planning phase: Product documents, SPEC examples
- Implementation phase: SPEC documents, code files, tests
- Sync phase: Documentation files, TAG chains

Carrying over unused context wastes tokens.

---

### 5. Use Context Pointers

**Instead of loading full content, use pointers**:
```markdown
# ❌ BAD: Paste entire SPEC document (5K tokens)
Full SPEC content here...

# ✅ GOOD: Reference SPEC by ID (50 tokens)
See SPEC-AUTH-001 for detailed requirements.
```

---

## Command-Specific Context Budgets

### /alfred:0-project (Project Initialization)

**Expected usage**: 30-50K tokens
- Phase 1: Analysis (20K)
- Phase 2: Interview (30K)
- Phase 3: Document creation (40K)

**Strategy**: Load project documents progressively, avoid loading all Skills upfront.

---

### /alfred:1-plan (SPEC Planning)

**Expected usage**: 40-60K tokens
- Phase 1: Project analysis (30K)
- Phase 2: SPEC creation (50K)

**Strategy**: Load product/structure/tech documents first, then SPEC standards JIT.

---

### /alfred:2-run (TDD Implementation)

**Expected usage**: 60-90K tokens
- Phase 1: SPEC analysis (40K)
- Phase 2: TDD cycle (70K)
- Phase 3: Quality gate (80K)

**Strategy**: Load SPEC documents first, then code/tests JIT during TDD cycle.

---

### /alfred:3-sync (Document Synchronization)

**Expected usage**: 70-100K tokens
- Phase 1: TAG verification (50K)
- Phase 2: Document sync (80K)

**Strategy**: Load TAG inventory first, then documentation files JIT during sync.

---

## Session Transition Best Practices

### When to Start New Session

**Recommended transition points**:
1. ✅ After `/alfred:0-project` completion → `/clear` → `/alfred:1-plan`
2. ✅ After `/alfred:1-plan` completion → `/clear` → `/alfred:2-run`
3. ✅ After `/alfred:2-run` completion → `/clear` → `/alfred:3-sync`

**Why**: Each command has different context needs. Starting fresh ensures optimal performance.

### Session Handoff Pattern

**Before starting new session**:
1. Save session state to `.moai/memory/session-context.json`
2. Note current SPEC ID, phase, and progress
3. Run `/clear` to reset context
4. Load only essential files for next phase

**Example handoff**:
```markdown
✅ Implementation complete for SPEC-AUTH-001

Session transition recommended:
1. Run /clear to start fresh
2. Load SPEC-AUTH-001 documents only
3. Proceed with /alfred:3-sync for document sync

Context saved to: .moai/memory/session-context.json
```

---

## Performance Metrics

### Context Budget Impact on Response Time

| Context Usage | Response Time | Recommendation                      |
| ------------- | ------------- | ----------------------------------- |
| 0-50K tokens  | Fast (1-2s)   | Optimal, continue                   |
| 50-100K tokens| Normal (2-4s) | Good, monitor                       |
| 100-150K tokens| Slower (4-6s) | Acceptable, plan transition         |
| 150-180K tokens| Slow (6-8s)   | ⚠️ Warning, start new session soon |
| 180-200K tokens| Very slow     | 🚨 Critical, start new session now |

---

## Troubleshooting

### Problem: Context budget exceeding 90%

**Symptoms**:
- Slow response times (>6s)
- Warning messages about context usage
- Difficulty loading new files

**Solutions**:
1. Start new session with `/clear`
2. Load only Tier 1 files (CRITICAL)
3. Use Skill() instead of Read() for large files
4. Use Grep() instead of Glob() for file searches

---

### Problem: Running out of context mid-task

**Symptoms**:
- Unable to load required files
- "Context budget exceeded" errors
- Incomplete responses

**Solutions**:
1. Save current progress to `.moai/memory/session-context.json`
2. Start new session with `/clear`
3. Load minimal context (current SPEC + implementation files only)
4. Resume task from saved state

---

## Summary

**Golden Rules**:
1. ✅ Load progressively: CRITICAL → HIGH → MEDIUM → LOW
2. ✅ Use Skills over full files (10x efficiency)
3. ✅ Parallelize independent reads (2x faster)
4. ✅ Clean up between phases (start fresh with `/clear`)
5. ✅ Monitor usage: 90% = warning, 95% = critical

**Context Budget Formula**:
```
Available tokens = 200K total - 20K overhead = 180K
Safe usage = 80% of available = 144K tokens
```

**When to transition**:
- After each major command completion
- When approaching 90% threshold (180K tokens)
- When response times exceed 6 seconds

**Performance targets**:
- Context usage < 80% (144K tokens): Optimal
- Context usage 80-90% (144-180K tokens): Acceptable, plan transition
- Context usage > 90% (>180K tokens): Start new session immediately
