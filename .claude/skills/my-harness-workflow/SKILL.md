---
name: my-harness-workflow
description: >
  SPEC Workflow domain knowledge for moai-adk-go covering SPEC document structure, EARS format
  requirements, plan-run-sync pipeline, MX tag protocol, and acceptance criteria.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-14"
  modularized: "false"
  tags: "spec, workflow, plan, run, sync, EARS, MX tag, acceptance criteria"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["SPEC", "plan", "run", "sync", "EARS", "MX tag", "acceptance criteria", "milestone", "wave", "AC-", "REQ-"]
  agents:
    - "my-harness-workflow-specialist"
    - "manager-spec"
    - "manager-develop"
    - "manager-docs"
  phases:
    - "plan"
    - "run"
    - "sync"
---

# SPEC Workflow Domain Knowledge

Domain-specific knowledge for the SPEC-based workflow in moai-adk-go. Supplements existing workflow skills with project-specific patterns.

## Quick Reference

### Workflow Pipeline

```
/moai plan "description"  ->  manager-spec  ->  SPEC document with EARS requirements
/moai run SPEC-XXX        ->  manager-develop  ->  Implementation (TDD or DDD)
/moai sync SPEC-XXX       ->  manager-docs  ->  Documentation + PR creation
```

### SPEC Document Structure

```
.moai/specs/<SPEC-ID>/
  spec.md          # EARS requirements + acceptance criteria
  plan.md          # Implementation plan with milestones (M1, M2, ...)
  scenarios.md     # Test scenarios
  risks.md         # Risk assessment + mitigations
  progress.md      # Progress tracking (auto-generated during run)
```

### EARS Requirement Patterns

| Type | Pattern | Example |
|------|---------|---------|
| Ubiquitous | The system shall [action] | "The system shall validate all template paths" |
| Event-Driven | When [event], the system shall [action] | "When a hook fails, the system shall log the error" |
| Unwanted | If [unwanted condition], the system shall [action] | "If template rendering fails, the system shall report the error" |
| State-Driven | While [state], the system shall [action] | "While in plan phase, the system shall capture LSP baseline" |
| Optional | Where [feature] is enabled, the system shall [action] | "Where worktree isolation is enabled, the system shall create a worktree" |

### MX Tag Types

| Tag | Purpose | When to Apply |
|-----|---------|---------------|
| `@MX:NOTE` | Context and intent | New exported functions, non-obvious logic |
| `@MX:WARN` | Danger zone (requires `@MX:REASON`) | Goroutines, complexity >= 15, concurrency |
| `@MX:ANCHOR` | Invariant contract | High fan_in functions (>= 3 callers) |
| `@MX:TODO` | Incomplete work | Untested public functions, deferred work |

### Plan-in-Main Doctrine

SPEC plan PRs merge to main (not feature branches). Run phase uses worktree isolation. This ensures:
- Plan documents are always available on main
- Run worktrees start from a clean main checkout
- Sync PRs merge to main with full review history

## Implementation Guide

### SPEC Naming Convention

```
SPEC-V{major}R{minor}-{CATEGORY}-{NUMBER}
```

Categories: WF (workflow), ORC (orchestration), RT (runtime), SPC (spec-system),
HOOK (hooks), CI (CI/CD), CON (constitution), MX (MX tags), CLI (CLI), TUI (TUI),
BRAIN (brain), STATUSLINE (statusline), HYBRID (hybrid), MIG (migration), GLM (GLM)

### Milestone Decomposition

Milestones group related requirements and map to testable acceptance criteria:

```
M1: Foundation (data structures, interfaces)
M2: Core logic (main implementation)
M3: Integration (cross-module connections)
M4: Edge cases (error handling, corner cases)
M5: Polish (documentation, final tests)
```

### Wave Splitting

For SPECs with 30+ tasks, split into waves:
- Each wave = 1 PR with cohesive set of milestones
- Waves execute sequentially (plan -> run -> sync per wave)
- Progress tracked in `.moai/specs/<SPEC-ID>/progress.md`

### Acceptance Criteria Format

```
AC-{SPEC-SHORT}-{NUMBER}: {verifiable condition}
  - Verification: {how to verify}
  - Priority: {P0|P1|P2}
```

Example:
```
AC-HOOK-001: All 27 hook events have handler scripts
  - Verification: `ls .claude/hooks/moai/handle-*.sh | wc -l` returns 27
  - Priority: P0
```

### SPEC Status Lifecycle

```
draft -> in-progress -> implemented -> in-review -> completed
```

Status is tracked in `spec.md` frontmatter `status` field and synchronized via `moai status`.

## Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md`: Canonical SPEC workflow rules
- `moai-workflow-spec` skill: SPEC workflow orchestration
- `moai-foundation-core` skill: SPEC system and TRUST 5
- `moai-workflow-tdd` skill: TDD RED-GREEN-REFACTOR cycle
- `moai-workflow-ddd` skill: DDD ANALYZE-PRESERVE-IMPROVE cycle
- `.claude/skills/moai/references/mx-tag.md`: MX tag protocol details
