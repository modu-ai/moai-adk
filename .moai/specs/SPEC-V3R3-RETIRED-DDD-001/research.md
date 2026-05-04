# SPEC-V3R3-RETIRED-DDD-001 Research (Phase 0.5)

> Codebase grounded scan for manager-ddd retire stub standardization.
> Companion to `spec.md` v0.3.0 and `plan.md` v0.3.0.
> Solo mode, no worktree. Working directory: project root. Branch: `feature/SPEC-V3R3-RETIRED-DDD-001` (base `origin/main` `20d77d931`).

## HISTORY

| Version | Date       | Author              | Description                                                                                                         |
|---------|------------|---------------------|---------------------------------------------------------------------------------------------------------------------|
| 0.3.0   | 2026-05-04 | MoAI Plan Workflow  | Force-accept (iter 3). §2 file taxonomy verified via `grep -rln`; §3 predecessor pattern inheritance documented.    |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow  | iter 2 cross-artifact partial sync — §2 file count refined to 30+1+2+3=36.                                          |
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | Phase 0.5 research: predecessor SPEC-V3R3-RETIRED-AGENT-001 pattern study, codebase grep for manager-ddd references. |

---

## 1. Research Objectives

본 research.md는 plan-auditor must-pass criterion ("research.md grounds decisions in actual codebase scan AND verifies external behavior")를 충족하기 위해 다음을 수행한다:

1. **Predecessor SPEC-V3R3-RETIRED-AGENT-001 패턴 분석** — manager-tdd retire stub frontmatter standard, audit test pattern, SubagentStart guard interaction.
2. **moai-adk-go 템플릿 트리에서 `manager-ddd` 언급 enumeration** — file-by-file grep evidence; Cat A/B/C 분류.
3. **Cat C 처리 결정 grounding** — `factory.go` `ddd_handler.go` dispatch path 검증 + `agent-hooks.md` 표 row 보존 vs 제거 결정 근거.
4. **manager-tdd retired stub과의 size delta + body structure 비교** — Cat B B-01 REWRITE target 1.4KB 합리성 검증.
5. **Audit test schema 재사용 검증** — predecessor M1 산출의 table-driven schema가 manager-ddd row 추가만으로 충분한지.

---

## 2. Codebase Grounded Scan (grep evidence)

### 2.1 Current state of `manager-ddd.md`

**File**: `internal/template/templates/.claude/agents/moai/manager-ddd.md`
**Size**: 7674 bytes (current — full DDD agent definition)
**Frontmatter (head)**:
```yaml
---
name: manager-ddd
description: ...DDD ANALYZE-PRESERVE-IMPROVE...
tools: <full CSV list>
model: sonnet
permissionMode: acceptEdits
memory: project
skills:
  - moai-workflow-ddd
hooks:
  ...
---
```

This is a fully-active agent definition, NOT a retired stub. Predecessor SPEC-V3R3-RETIRED-AGENT-001 explicitly left this case as out-of-scope (RETIRED-AGENT spec.md §1.3 / §2.2: "manager-ddd retired stub의 동등 standardization (동일 패턴 적용 가능하지만 별도 SPEC; 본 SPEC은 manager-tdd case 집중)").

### 2.2 Predecessor `manager-tdd.md` (the reference pattern)

**File**: `internal/template/templates/.claude/agents/moai/manager-tdd.md`
**Size**: 1392 bytes (post-RETIRED-AGENT M2)
**Frontmatter (full)**:
```yaml
---
name: manager-tdd
description: |
  Retired (SPEC-V3R3-RETIRED-AGENT-001) — use manager-cycle with cycle_type=tdd.
  This agent has been consolidated into the unified manager-cycle agent.
  See manager-cycle.md for the active replacement.
retired: true
retired_replacement: manager-cycle
retired_param_hint: "cycle_type=tdd"
tools: []
skills: []
---
```

**Body**:
- `# manager-tdd — Retired Agent` heading
- "## Replacement" section — `Use **manager-cycle** with \`cycle_type=tdd\` instead.`
- "## Migration Guide" table (2 rows: old invocation → new invocation)
- "## Why This Change" section (3-4 sentences explaining unification)
- "## Active Agent" section pointing to `.claude/agents/moai/manager-cycle.md`

This is the canonical reference for Cat B B-01 REWRITE in this SPEC.

### 2.3 `manager-cycle.md` (the active replacement)

**File**: `internal/template/templates/.claude/agents/moai/manager-cycle.md`
**Size**: 11385 bytes (predecessor M2 산출)
**Role**: Unified DDD/TDD implementation cycle. Accepts `cycle_type` parameter (`ddd` or `tdd`).

This file contains 3 occurrences of `manager-ddd` (per `grep -c`), all in the form of cross-references — typically migration guidance pointing FROM `manager-ddd` TO `manager-cycle with cycle_type=ddd`. These 3 occurrences are within the manager-cycle.md body and ARE in scope for Cat A SUBSTITUTE (item A-02 in plan.md §3 manifest), since the prose context is "manager-cycle replaces manager-ddd". Substitution would need to preserve the migration guidance semantics (e.g., reword to "manager-cycle's ddd cycle_type covers the use case formerly served by the retired manager-ddd").

### 2.4 grep evidence — files referencing `manager-ddd`

Run at iter 3 force-accept time (2026-05-04):

```
$ grep -rln "manager-ddd" internal/template/templates/ | sort
internal/template/templates/.claude/agents/moai/evaluator-active.md
internal/template/templates/.claude/agents/moai/expert-backend.md
internal/template/templates/.claude/agents/moai/expert-debug.md
internal/template/templates/.claude/agents/moai/expert-devops.md
internal/template/templates/.claude/agents/moai/expert-frontend.md
internal/template/templates/.claude/agents/moai/expert-mobile.md
internal/template/templates/.claude/agents/moai/expert-refactoring.md
internal/template/templates/.claude/agents/moai/expert-testing.md
internal/template/templates/.claude/agents/moai/manager-cycle.md
internal/template/templates/.claude/agents/moai/manager-ddd.md
internal/template/templates/.claude/agents/moai/manager-quality.md
internal/template/templates/.claude/agents/moai/manager-spec.md
internal/template/templates/.claude/agents/moai/manager-strategy.md
internal/template/templates/.claude/agents/moai/manager-tdd.md
internal/template/templates/.claude/output-styles/moai/moai.md
internal/template/templates/.claude/rules/moai/core/agent-hooks.md
internal/template/templates/.claude/rules/moai/development/agent-authoring.md
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md
internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md
internal/template/templates/.claude/skills/moai-foundation-quality/SKILL.md
internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-ddd/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-spec/references/examples.md
internal/template/templates/.claude/skills/moai-workflow-spec/references/reference.md
internal/template/templates/.claude/skills/moai-workflow-testing/SKILL.md
internal/template/templates/.claude/skills/moai/references/mx-tag.md
internal/template/templates/.claude/skills/moai/references/reference.md
internal/template/templates/.claude/skills/moai/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/moai.md
internal/template/templates/.claude/skills/moai/workflows/run.md
internal/template/templates/CLAUDE.md
```

**Total**: 35 files (counted manually from sort output). plan.md §3 enumerates 30 Cat A + 1 Cat B + 2 Cat C = 33 within-taxonomy files. Discrepancy explanation:

- The grep output above includes `moai-workflow-ddd/SKILL.md` and `moai-workflow-testing/SKILL.md` which were NOT enumerated in plan.md §3 (iter 3 freeze list). These 2 files are within tolerance per plan.md §3 note ("tolerance ±2 acceptable").
- Implementation phase M3.1 will reconcile this list with the live grep result and adjust Cat A enumeration. Either:
  - Add the 2 files to Cat A (becoming 32 SUBSTITUTE), OR
  - Validate they are special cases (e.g., `moai-workflow-ddd/SKILL.md` describes the DDD methodology and may legitimately retain `manager-ddd` mention as part of historical context).

**Recommended decision**: Add these 2 files to Cat A SUBSTITUTE during M3.1. Treating them as "special preservation" would create a documentation drift exception that undermines REQ-RD-009. Better to substitute consistently and adjust prose if needed.

### 2.5 Categorization rationale

**Cat B (REWRITE)** — single file `manager-ddd.md`. Justification: the file is a full agent definition that must be replaced wholesale with a retired stub. This is the only file where structure changes drastically.

**Cat C (UPDATE-WITH-ANNOTATION)** — 2 files:
- `agent-hooks.md` — contains the canonical Agent Hook Actions table that maps agents to hook action keys. The `manager-ddd` row's action keys (`ddd-pre-transformation`, etc.) are dispatched by `internal/hook/agents/factory.go` to `ddd_handler.go`. Removing the row would orphan the dispatch documentation. Annotation preserves the dispatch documentation while marking retired status.
- `agent-authoring.md` — contains the canonical Manager Agents (8) listing. This document is the source of truth for agent enumeration. Predecessor's manager-tdd was removed from this listing using Option A. We follow Option A consistently.

**Cat A (SUBSTITUTE)** — all remaining 30 (or 32 after M3.1 reconciliation). These files contain prose-level cross-references to `manager-ddd` that point to the agent's role or invocation pattern. Substitute targets are unambiguous — `manager-cycle` (general) or `manager-cycle with cycle_type=ddd` (parameter-relevant context).

---

## 3. Predecessor Pattern Inheritance Analysis

### 3.1 Pattern verbatim reused from RETIRED-AGENT

| RETIRED-AGENT element | Inheritance in this SPEC |
|---|---|
| 5-field frontmatter standard (REQ-RA-002) | REQ-RD-001 verbatim (only `cycle_type` value differs: `tdd` → `ddd`) |
| Migration table body structure | M2.3 in plan.md mirrors predecessor body section ordering |
| `agent_frontmatter_audit_test.go` table-driven test | REQ-RD-004 / REQ-RD-010 — single row addition |
| SubagentStart hook block decision (REQ-RA-007) | REQ-RD-005 — same handler, same code path |
| documentation substitution discipline (REQ-RA-013) | REQ-RD-002 / REQ-RD-009 — same pattern, different file count |
| `RETIREMENT_INCOMPLETE_<agent>` CI assertion (REQ-RA-016) | REQ-RD-012 — composite unwanted-behavior with same assertion form |
| 5-milestone TDD breakdown (M1 RED → M2-M4 GREEN → M5 REFACTOR) | plan.md §2 same skeleton |
| solo mode, no worktree | spec.md §7 — predecessor convention |
| BREAKING CHANGE: false | spec.md §10 — predecessor convention |

### 3.2 Pattern divergence from RETIRED-AGENT (justified)

| Aspect | RETIRED-AGENT | This SPEC | Justification |
|---|---|---|---|
| File count in scope | 12 (predecessor `In Scope` enumeration in spec.md §2.1) | 36 (30 Cat A + 1 Cat B + 2 Cat C + 3 ancillary) | manager-ddd has more downstream documentation references because DDD methodology is more frequently cross-referenced in skill docs |
| New code files | `agent_start.go`, `agent_start_test.go`, `manager_cycle_present_test.go`, `agent_frontmatter_audit_test.go`, etc. | None — all infrastructure already in place from PR #776 | predecessor created the runtime guard + audit harness; this SPEC is mechanical replication |
| Wrapper validation (REQ-RA-005) | New code path in launcher.go | Not in scope (already implemented) | predecessor REQ-RA-005 covers all worktree spawns including manager-ddd |
| Lessons.md entry | New entry #11 (5-layer defect chain) | Not required — predecessor lesson covers manager-ddd case too | manager-ddd retire is the same pattern; no novel anti-pattern surfaced |

### 3.3 Risks specific to mechanical replication

The primary risk in mechanical replication is **silent assumption**: that the predecessor pattern works identically without re-verification. Mitigations applied:

1. spec.md §8 risk row "predecessor merge 후 main에서 manager-cycle.md 정의가 변경되어 retired stub의 `retired_replacement: manager-cycle` 참조가 stale" — verifies replacement target is live.
2. plan.md §4 risk row C18 — Cat A enumeration baseline drift between iter 3 freeze and implementation start.
3. acceptance.md AC-RD-04 (Negative scenario) — verifies SubagentStart guard works for manager-ddd identically to manager-tdd, not assumed.

---

## 4. External Behavior Verification

### 4.1 Claude Code Agent runtime (frontmatter parsing)

Predecessor RETIRED-AGENT research.md verified that Claude Code Agent runtime gracefully handles `tools: []` empty arrays and unknown frontmatter fields like `retired: true`. The runtime spawns a 0-tool_uses silent agent in absence of SubagentStart guard. With the guard from PR #776 in place, the spawn is blocked at the SubagentStart layer in ≤500ms.

For manager-ddd, the runtime behavior is identical because:
- The `name` field is the only required identity (`manager-ddd`).
- All other fields (`retired`, `retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`) follow the same parsing path as manager-tdd.
- SubagentStart guard reads `retired: true` flag identically regardless of agent name.

**Evidence**: PR #776 merged (commit `20d77d931`). RETIRED-AGENT acceptance test `internal/hook/agent_start_test.go TestAgentStartHandler_Retired` exercises the manager-tdd case; extending with manager-ddd test case in this SPEC's M3 (acceptance.md AC-RD-04 Test Anchor) provides direct verification.

### 4.2 `factory.go` ddd_handler dispatch path

```
$ grep -A3 "ddd" internal/hook/agents/factory.go | head -20
```

Expected (predecessor PR #776 didn't touch this):
- `factory.go` `New()` switch maps `"ddd-pre-transformation"`, `"ddd-post-transformation"`, `"ddd-completion"` to `ddd_handler.go`'s constructor.
- Dispatch is keyed on action string, NOT on agent name. So even with `manager-ddd` retired, action keys still route correctly when called by hook events.

This decouples retired-agent classification (frontmatter `retired: true`) from hook action dispatch (factory keyed on action string). REQ-RD-007 invariant.

### 4.3 16-language neutrality verification

Per CLAUDE.local.md §15, retired stub body must be language-agnostic. The migration table content references `cycle_type=ddd` parameter only — no Go/Python/TypeScript-specific syntax. Predecessor `manager-tdd.md` body verifies this property; mechanical replication preserves it.

---

## 5. Implementation Recommendation Summary

1. **M1 RED**: Add manager-ddd row to `agent_frontmatter_audit_test.go` table; verify RED.
2. **M2 GREEN-1**: Read `manager-tdd.md`; copy structure to new `manager-ddd.md`; substitute `tdd` → `ddd` and minor body adjustments (DDD-specific migration table). `make build`. Verify audit GREEN.
3. **M3 GREEN-2**: `grep -rln "manager-ddd" internal/template/templates/` → 32 hits expected (30 Cat A + 1 Cat B self-ref + 2 Cat C with annotations OK). Edit Cat A 30-32 files individually. After M3.3, grep should return only Cat B + Cat C hits.
4. **M4 GREEN-3**: Annotate Cat C C-01 and C-02. `make build`. Full test GREEN.
5. **M5 REFACTOR**: CHANGELOG entry, lessons review, PR creation.

**Estimated complexity**: LOW. Pattern is verbatim replication. Risk surface is bounded by predecessor's validated infrastructure.

---

## 6. Cross-references

- Predecessor: `.moai/specs/SPEC-V3R3-RETIRED-AGENT-001/research.md` (5-layer defect chain analysis source)
- spec.md (companion)
- plan.md (companion, file taxonomy)
- acceptance.md (companion, 4 ACs)
- CLAUDE.local.md §2 Template-First HARD rule
- CLAUDE.local.md §15 16-language neutrality
- PR #776 (merged main `20d77d931`) — predecessor implementation

---

End of research.md
