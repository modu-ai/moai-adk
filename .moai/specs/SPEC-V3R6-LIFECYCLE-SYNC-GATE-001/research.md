---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
artifact: research
version: "0.1.1"
created: 2026-05-25
updated: 2026-05-25
---

## HISTORY

### v0.1.1 (2026-05-25, manager-spec — iter-2 cross-reference sync)
- No research changes in this iteration; version bumped for consistency with spec.md v0.1.1 + acceptance.md v0.1.1 + design.md v0.1.1
- Verbatim 154-SPEC era × sync-signal cross-tab audit data unchanged (research origin remains stable)

### v0.1.0 (2026-05-25, manager-spec)
- Initial research.md authored capturing origin context + verbatim memory citations
- 154-SPEC era × sync-signal cross-tab data verbatim from orchestrator turn 2026-05-25
- L60 atomic backfill + L67 manager-docs scope-creep pattern verbatim citations
- Status Transition Ownership Matrix cross-reference + design alignment

---

## A. Origin / Motivation (Verbatim Findings)

### A.1 154-SPEC Era × Sync-Signal Cross-Tab Audit (2026-05-25)

The orchestrator session 2026-05-25 (post-Sprint 10 cohort 8/8 close) conducted a full enumeration of all 154 `status: implemented` SPECs across 4 era × sync-signal cross-tab dimensions. The verbatim breakdown:

```
era × sync-signal cross-tab (sync_section / sync_commit_sha / mx_commit_sha):
  80 N|N|N|N  ← progress.md 자체 없음 (V2.x / V3R2-R4 era-final)
  42 Y|N|N|N  ← progress.md 있으나 sync 흔적 없음 (구 era 표준)
  23 Y|N|N|Y  ← sync section만 있고 sync_commit_sha 없음 (V3R5 / 초기 V3R6)
   6 Y|Y|N|Y  ← sync done, Mx 누락 (V3R6 모던 표준 위반)
   3 Y|Y|Y|Y  ← 4-phase 완전 종료지만 spec.md status drift (L67 매니저-docs scope-creep)
```

The session resolved 4 violations (3 Y|Y|Y|Y + 1 LOCAL-NAMESPACE-CONSOLIDATION-001 status drift) via 4 orchestrator-direct atomic chore commits at HEAD `d74095e75`:
- `baaa1693e` SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001: spec.md status backfill — implemented → completed
- `d9ae06020` SPEC-V3R6-MULTI-SESSION-COORD-001: spec.md status backfill — implemented → completed
- `b8be7e44a` SPEC-V3R6-HARNESS-PROPOSAL-GEN-001: spec.md status backfill — implemented → completed
- `d74095e75` SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001: spec.md status backfill — implemented → completed

**5 modern-era violations remain** (Y|Y|N|Y category from above, minus 1 already-closed SPEC):
- SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (sync `11abb9a30`, mx missing)
- SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (sync `a853f2954`, mx missing)
- SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (sync `2d9871208b09e1ce647a4cc134b24267b713b42f`, mx=null literal)
- SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 (sync `009e68c5d`, mx missing)
- SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 (sync + mx both missing — broken state)

These 5 SPECs constitute the M6 dogfood verification target.

## B. Root Cause Analysis (Verbatim Memory Citations)

### B.1 L60 Pattern — Atomic Backfill Chicken-and-Egg (2026-05-23 onward)

The L60 atomic backfill pattern emerged during SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 sync-phase work (PR #1043 era). The pattern documents that:

> sync_commit_sha references its own future SHA (chicken-and-egg). Resolution: two-commit cadence where the second commit backfills the first commit's SHA into progress.md §E.2. The same pattern applies to mx_commit_sha — third commit backfills the second.

L60 occurrence count: 14x across SPECs (per memory `project_sprint10_test_refactor_4phase_close.md`).

**Design implication for this SPEC**: L60 design is operationally valid but introduces multi-commit complexity. The 5 modern-era violations all stem from one or more L60 steps being skipped during agent spawn-and-spawn cycles. `moai spec close` reduces this to 1 atomic commit via predictive/tag-after-commit SHA resolution (design.md §B.1 Option D).

### B.2 L67 Pattern — Manager-Docs Scope-Creep (NEW, 2026-05-25)

L67 is a NEW (proposed) lesson catalogued during the SPEC-V3R6-TEST-REFACTOR-001 4-phase close, captured verbatim from memory `project_sprint10_test_refactor_4phase_close.md`:

> manager-docs scope creep pattern: sync-phase spawn auto-absorbs Mx + 4-phase close concepts without explicit delegation, producing N commits with wrong commit message info (claimed counts/SHAs mismatch git state) + incomplete follow-through (claimed status: completed but actual implemented) + missing HISTORY entries despite frontmatter version bumps + partial §A backfills.
>
> Mitigation: delegation prompt MUST enumerate forbidden downstream work via L46-style + behavioral negation directives ("DO NOT execute X transition" patterns proved effective with manager-develop D1 mitigation); orchestrator independent verify MUST cross-check commit message claims against git state.

**Design implication for this SPEC**:
- The pre-commit hook (deliverable 3) MECHANIZES the L67 mitigation. When manager-docs commits with subject "4-phase close — atomic" but diff lacks spec.md status: completed, the hook blocks the commit at the git layer — eliminating the human/orchestrator vigilance burden.
- The spec-lint OwnershipTransitionRule (deliverable 4) detects the inverse pattern: manager-develop attempting `in-progress → implemented` transition that should be sync-phase responsibility. Both mechanisms together cover both directions of ownership drift.

### B.3 Status Transition Ownership Matrix (Existing SSOT)

Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix, the canonical owners are:

| Transition | Canonical Owner Agent |
|------------|----------------------|
| (none) → draft | manager-spec (plan-phase) |
| draft → planned | (CI/hook automatic on plan PR merge) |
| planned → in-progress | manager-develop (run-phase milestone 1 start) |
| in-progress → implemented | manager-docs (sync-phase) |
| implemented → completed | manager-docs (sync-phase) OR orchestrator-direct (4-phase close marker) |
| * → superseded | manager-spec (new SPEC supersession) |
| * → archived | manager-spec OR user explicit decision |
| * → rejected | user explicit decision |

This SPEC's REQ-LSG-004 + OwnershipTransitionRule mechanically enforces the matrix.

## C. Anthropic 2026 Alignment

### C.1 Sub-Agent Boundary (Verbatim from claude.com/docs/en/sub-agents)

> "Subagents cannot spawn other subagents."

**Design implication**: The pre-commit hook MUST NOT invoke AskUserQuestion (REQ-LSG-011 HARD). The hook is a sub-process, not a subagent, but the same boundary applies — user interaction belongs exclusively to the orchestrator. Hook emits exit 2 + JSON; orchestrator translates per agent-common-protocol §Hook Invocation Surface.

### C.2 Hook Vocabulary (Anthropic Audit 3 Finding A5)

> "Stop, PostToolUse, SubagentStop, TaskCompleted hook events are first-class observability surfaces."

**Design implication**: This SPEC adds a `PreCommit` hook to the canonical vocabulary (extension; not first-class per Anthropic but supported by Claude Code's Stop/PostToolUse infrastructure). The hook integrates with the 3-NEW-hook pattern established by SPEC-V3R6-AGENT-TEAM-REBUILD-001 (status-transition-ownership.sh / sync-phase-quality-gate.sh / team-ac-verify.sh). This SPEC adds a 4th NEW hook.

### C.3 Adaptive Thinking — `ultrathink` Keyword

Per CLAUDE.md §12, deep reasoning on Opus 4.7+ activates via the `ultrathink` keyword. The orchestrator session 2026-05-25 used `ultrathink` for the 154-SPEC cross-tab audit; the resulting analysis directly motivated this SPEC's 5-deliverable architecture.

## D. Predecessor SPEC Cross-Reference

### D.1 SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (Predecessor)

Status: implemented (Y|Y|N|Y modern-era violation per A.1)

This SPEC's Status Transition Ownership Matrix is the foundation REQ-LSG-004 + OwnershipTransitionRule enforce mechanically. The matrix authority remains with the predecessor SPEC; this SPEC merely adds mechanical enforcement.

M6 dogfood will close this SPEC via `moai spec close SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 --backfill-only`.

### D.2 SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (Predecessor)

Status: implemented (Y|Y|N|Y modern-era violation per A.1)

This SPEC canonicalized GEARS notation, which the present SPEC uses 100% (REQ-LSG-001..015 all GEARS-formatted, zero IF/THEN legacy patterns).

M6 dogfood will close this SPEC via `moai spec close SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 --backfill-only`.

### D.3 SPEC-V3R6-AGENT-TEAM-REBUILD-001 (Related)

Status: completed

Established the 8-agent retained catalog and the 3-NEW-hook pattern. This SPEC adds a 4th hook (`handle-pre-commit-spec-status.sh`) following the same integration pattern (exit 2 + JSON → orchestrator translates).

### D.4 SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (Related)

Status: completed (per Sprint 10 lane B 4/4 close)

Established CLAUDE.local.md §25 [HARD] template-internal-isolation rule. This SPEC's D.1.4 HARD constraint enforces the same isolation: `internal/template/templates/**` MUST NOT be modified by plan-phase artifacts (M3 milestone has the sole narrow exception for settings.json.tmpl hook registration).

## E. Decision Rationale — Why a NEW SPEC vs Inline Fix

### E.1 Considered Alternative: Inline Fix the 5 Violations Only

The 5 remaining violations COULD be resolved via 5 orchestrator-direct atomic chore commits (same pattern as the 4 fixed this session: `baaa1693e`, `d9ae06020`, `b8be7e44a`, `d74095e75`). Why escalate to a NEW SPEC?

**Reasons for NEW SPEC**:
1. **Recurrence prevention**: 9 violations occurred in 154 SPECs (~5.8% rate). Mechanical guardrails (hook + lint + atomic close) reduce future recurrence to near-zero.
2. **Operational simplification**: 5 commits per close → 1 commit per close reduces L60 chicken-and-egg complexity by 80%.
3. **Audit transparency**: `moai spec audit --json` makes drift detection deterministic and CI-friendly; no manual cross-tab analysis required for future audits.
4. **L67 mitigation**: The manager-docs scope-creep pattern is a recurring anti-pattern (catalogued just this session). Pre-commit hook mechanizes the mitigation.
5. **Status Transition Ownership Matrix enforcement**: REQ-ARR-002/003 of predecessor SPEC defined ownership; this SPEC mechanically enforces it.

**Reasons against inline fix**:
- Resolves symptom (5 SPECs), not cause (process complexity)
- Does not prevent the next occurrence
- Wastes the cross-tab audit's signal value

### E.2 Tier L Justification

Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier:
- **Tier S**: < 300 LOC, < 5 files → does not apply
- **Tier M**: 300-1000 LOC, 5-15 files → close, but file count (15) is the boundary
- **Tier L**: > 1000 LOC OR > 15 files → applies (estimated ~1650 LOC across 15 files per plan.md §F milestones)

**Files modified estimate**:
1. `internal/spec/closer.go` (NEW, ~300 LOC)
2. `internal/spec/audit.go` (NEW, ~200 LOC)
3. `internal/spec/lifecycle_close_test.go` (NEW, ~250 LOC)
4. `internal/spec/audit_test.go` (NEW, ~200 LOC)
5. `internal/cli/spec_close.go` (NEW, ~200 LOC)
6. `internal/cli/spec_audit.go` (NEW, ~150 LOC)
7. `internal/cli/spec_close_test.go` (NEW, ~150 LOC)
8. `internal/cli/spec_audit_test.go` (NEW, ~150 LOC)
9. `cmd/moai/main.go` (modified, ~20 LOC delta for subcommand registration)
10. `.claude/hooks/moai/handle-pre-commit-spec-status.sh` (NEW, ~80 LOC)
11. `.claude/hooks/moai/handle-pre-commit-spec-status_test.sh` (NEW, ~50 LOC)
12. `internal/template/templates/.claude/settings.json.tmpl` (modified narrow scope, ~10 LOC delta for hook entry)
13. `internal/spec/lint_ownership.go` (NEW, ~50 LOC)
14. `internal/spec/lint_ownership_test.go` (NEW, ~150 LOC)
15. `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (NEW, ~250 LOC)
16. `.claude/rules/moai/development/spec-frontmatter-schema.md` (modified, ~20 LOC delta for era field)
17. CHANGELOG.md (sync-phase, ~30 LOC delta)

Total: ~2260 LOC across 17 files → clearly Tier L.

## F. Open Research Questions (Deferred to Run-Phase)

1. **Predictive SHA mechanism feasibility**: Whether design.md §B.1 Option C (git commit-tree predictive SHA) is robust across git versions ≥ 2.34. M1 implementation MAY benchmark Option C; default to Option D if Option C proves fragile.

2. **Audit performance on 200-SPEC fixture**: NFR-LSG-001 bounds at 5s. Single-goroutine baseline timing unknown; M1 benchmark required.

3. **Pure-bash YAML parser limitations**: Hook deliverable uses pure-bash `awk`. If `awk` proves insufficient for nested YAML (e.g., `lint.skip` array), fall back to `yq` with documented dependency.

4. **Era heuristic false-positive rate**: H-1 through H-6 heuristics may misclassify edge-case SPECs. M1 implementation MUST add regression test against actual 154 historical SPECs to verify classification accuracy.

5. **mx_commit_sha self-reference UX**: When mx_commit_sha references the close commit's own SHA, downstream tooling (e.g., spec-lint, audit) MUST handle the recursion. M1 implementation MUST add recursion-detection guard.

## G. Verbatim Memory Citations Summary

| Memory File | Relevance |
|-------------|-----------|
| `project_sprint10_test_refactor_4phase_close.md` | L60 14x atomic backfill + L67 manager-docs scope-creep NEW |
| `project_sprint10_laneB_lnco001_4phase_close.md` | L60 12th instance demonstrating chicken-and-egg fragility |
| `project_atr001_4phase_close.md` | Predecessor SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 partial closure (Y|Y|N|Y) |
| `feedback_template_internal_content_isolation.md` | CLAUDE.local.md §25 HARD anchor for D.1.4 |
| `MEMORY.md` (Sprint 10/11 entries) | Cross-tab audit signal data |

All citations are verbatim from the memory snapshot at orchestrator turn 2026-05-25.
