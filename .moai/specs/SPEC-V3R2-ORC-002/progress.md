---
spec_id: SPEC-V3R2-ORC-002
plan_complete_at: 2026-05-10T00:00:00Z
plan_status: audit-ready
branch: feature/SPEC-V3R2-ORC-002-agent-lint
base_commit: "ab0fc4dda"
base_branch: origin/main
coverage_target: "85% on internal/cli/agent_lint*.go"
---

# Progress: SPEC-V3R2-ORC-002 — Plan Phase Complete

Plan-phase artefacts produced for SPEC-V3R2-ORC-002 (Agent Common Protocol
CI Lint — `moai agent lint`). Awaiting independent plan-auditor review
before proceeding to run phase.

---

## Status

- [x] Phase 0.5 — Codebase research (research.md)
- [x] Phase 1B — Implementation plan (plan.md)
- [x] Phase 1B — Acceptance criteria (acceptance.md)
- [x] Phase 1B — Tasks breakdown (tasks.md)
- [x] Phase 1B — Compact reference (spec-compact.md)
- [x] Phase 1B — Issue body (issue-body.md)
- [x] Plan complete; awaiting plan-auditor independent verification
- [ ] Plan-auditor approval received (pending)
- [ ] Run phase entry (deferred to next session per session-handoff
      protocol)

---

## Branch Information

- **Branch**: `feature/SPEC-V3R2-ORC-002-agent-lint`
- **Base**: `origin/main` HEAD `ab0fc4dda` (PR #814 — SPEC-V3R2-RT-004
  Typed Session State + Phase Checkpoint plan)
- **Worktree**: `/Users/goos/.moai/worktrees/moai-adk-go/orc-002-plan`
- **Repo**: `modu-ai/moai-adk`
- **Breaking**: `true` (BC-V3R2-004 declared in spec.md L23-24)
- **Lifecycle**: spec-anchored

---

## Plan Artefact Inventory

| File | Purpose | Status |
|------|---------|--------|
| `spec.md` | EARS requirements (17 REQs) + ACs preview (12) — preserved at v0.1.0 | Pre-existing (NOT modified by this plan run) |
| `research.md` | Phase 0.5 codebase research with 35+ file:line citations across 11 sections | NEW |
| `plan.md` | Phase 1B implementation plan (M1-M5 milestones with mx_plan, plan-audit-ready checklist) | NEW |
| `acceptance.md` | Hierarchical AC format (14 ACs, AC-V3R2-ORC-002-NN with .a/.b/.c sub-children) + REQ↔AC traceability matrix | NEW |
| `tasks.md` | 22 T-ORC002-NN tasks across M1-M5 with TDD owner roles + dependency graph | NEW |
| `progress.md` | This file | NEW |
| `spec-compact.md` | Compact reference (REQ + AC + Files-to-modify + Exclusions + Wave position) | NEW |
| `issue-body.md` | GitHub PR / issue body | NEW |

**Total artefacts**: 8 files (1 pre-existing + 7 new). Per task constraint,
`spec.md` was NOT modified by this plan run; spec-level amendments
(OQ-1..6 resolution HISTORY entry) are deferred to the run phase
milestone M5.2.

---

## Acceptance Criteria Status (PENDING)

All 14 ACs are PENDING until run phase implementation. Coverage target:
85% line coverage on `internal/cli/agent_lint*.go`.

| AC ID | REQ Coverage | Status | Verification handle |
|-------|--------------|:------:|---------------------|
| AC-V3R2-ORC-002-01 | REQ-001 | PENDING | `moai agent lint --help` exits 0 with all 4 flags |
| AC-V3R2-ORC-002-02 | REQ-002, REQ-003, REQ-006, REQ-007 | PENDING | Baseline tree → 9 LR-01, 4-5 LR-02, 1 LR-07 |
| AC-V3R2-ORC-002-03 | REQ-002, REQ-003 | PENDING | Post-cleanup tree → 0 errors |
| AC-V3R2-ORC-002-04 | REQ-010, REQ-013, REQ-014 | PENDING | JSON schema valid + version "1.0" + jq parity |
| AC-V3R2-ORC-002-04.a | REQ-014 | PENDING | `version == "1.0"` |
| AC-V3R2-ORC-002-04.b | REQ-013 | PENDING | Pre-commit hook YAML snippet documented |
| AC-V3R2-ORC-002-05 | REQ-006, REQ-011 | PENDING | New violation → CI Lint job exits 1 |
| AC-V3R2-ORC-002-06 | REQ-008 | PENDING | Dead-hook fixture → LR-04 |
| AC-V3R2-ORC-002-07 | REQ-009 | PENDING | Duplicate Skeptical → LR-07 |
| AC-V3R2-ORC-002-08 | REQ-004, REQ-012 | PENDING | Missing effort → warning only (exit 0) |
| AC-V3R2-ORC-002-09 | REQ-004, REQ-012 | PENDING | Missing effort + --strict → error (exit 1) |
| AC-V3R2-ORC-002-10 | REQ-006, REQ-015 | PENDING | Fenced-code AskUserQuestion → no LR-01 |
| AC-V3R2-ORC-002-11 | REQ-004, REQ-016 | PENDING | Malformed YAML → exit 2 |
| AC-V3R2-ORC-002-12 | REQ-005, REQ-009 | PENDING | Skeptical block exactly 1× in rule file |
| AC-V3R2-ORC-002-13 | REQ-006 carve-out (OQ-1) | PENDING | manager-brain → 0 LR-01 (orchestrator exempt) |
| AC-V3R2-ORC-002-14 | REQ-017 | PENDING | Two-tree drift → LINT_TREE_DRIFT warning |

---

## Key Plan Decisions

1. **TDD methodology**: per `.moai/config/sections/quality.yaml`
   `development_mode: tdd`, M3 follows RED → GREEN → REFACTOR with 14
   AC-driven test functions in `internal/cli/agent_lint_test.go` and 9
   testdata fixtures.

2. **Carve-out via frontmatter (OQ-1, Option A)**: agents declaring
   `AskUserQuestion` in `tools:` are auto-exempted from LR-01. This is
   data-driven; new orchestrator-class agents self-assert.

3. **Inline-code IS flagged (OQ-2)**: REQ-015 strict reading. Only
   triple-backtick fences exempt.

4. **LR-08 warning-only (OQ-3)**: 50% peer-omission threshold, no
   `--strict` promotion in this SPEC. Calibrate after observation.

5. **Simple `\|` matcher split (OQ-4)**: complex regex matchers emit
   `LR-04-COMPLEX-MATCHER` warning instead of full AST analysis.

6. **LINT_TREE_DRIFT semantics (OQ-5)**: per-file violation-tuple diff +
   `LINT_TREE_FILE_MISMATCH` for presence-difference.

7. **Build-time CI metric for runtime (OQ-6)**: no in-binary self-timing.
   Observe via CI dashboard.

8. **Template-First mandate**: All edits begin in
   `internal/template/templates/`; `make build` mirrors. CLAUDE.local.md
   §2 strictly enforced.

9. **Deferred work**: ORC-003 (LR-03→error), ORC-004 (LR-05→error),
   ORC-001 cleanup of 8 remaining LR-01 + 4 LR-02 violations, MIG-001
   legacy SPEC rewriter, IDE plugin integrations are explicitly NOT in
   scope.

---

## Open Questions Captured

OQ-1 — manager-brain orchestrator carve-out mechanism (resolved: Option A
data-driven exemption via `tools:` frontmatter)
OQ-2 — LR-01 inline-code handling (resolved: NO exemption)
OQ-3 — LR-08 same-family threshold (resolved: 50% warning-only)
OQ-4 — LR-04 hook matcher regex (resolved: simple `\|` split + complex
warning)
OQ-5 — LINT_TREE_DRIFT semantics (resolved: per-file tuple diff +
file-mismatch separate)
OQ-6 — Lint runtime budget enforcement (resolved: CI-side metric)

All 6 OQs have **recommended resolutions** documented in research.md §9.
Plan-auditor may concur or override; if override, T-ORC002-03 and M3
implementation absorb the new resolution.

---

## Counts Summary

- REQs in spec.md: **17** (Ubiquitous 5, Event-Driven 5, State-Driven 2,
  Optional 2, Unwanted 3)
- ACs in acceptance.md: **14** (AC-V3R2-ORC-002-01 .. AC-V3R2-ORC-002-14
  + 2 sub-children .04.a / .04.b)
- Tasks in tasks.md: **22** (M1: 3, M2: 7, M3: 6, M4: 3, M5: 3)
- Cited file:line anchors in research.md: **35+** unique anchors across
  ~15 source files
- Milestones: **5** (M1: P0, M2: P0, M3: P0, M4: P1, M5: P0)
- Lint rules implemented: **8** (LR-01 .. LR-08)
- Testdata fixtures: **9** (8 rule-specific + 1 clean baseline)

---

## Next Action

Open PR with title `plan(spec): SPEC-V3R2-ORC-002 — Agent Common Protocol
CI Lint` requesting **plan-auditor independent verification (10
dimensions)**. plan-auditor should:

1. Validate REQ-AC traceability matrix completeness (17 REQs ↔ 14 ACs).
2. Verify EARS compliance on all 17 REQs in spec.md §5.
3. Verify file:line citations in research.md are reachable from base
   `ab0fc4dda`.
4. Inspect deferred-work boundary (ORC-001/003/004 + MIG-001
   declarations correct).
5. Confirm OQ-1..6 recommended resolutions for soundness.
6. Check TDD methodology consistency (tasks.md owner roles align with
   `development_mode: tdd`).
7. Verify Template-First discipline applied throughout (M2 file list).
8. Confirm no `spec.md` content changes by this plan run (HISTORY
   amendment deferred to run phase).
9. Verify plan-audit-ready checklist (plan.md §8) all 15 boxes ticked.
10. Confirm exclusions section comprehensive (spec.md §1.2 + §2.2 + plan
    §2.2).

If plan-auditor returns PASS, run phase begins in a separate session per
the session-handoff protocol (paste-ready resume message generated
post-merge).

---

## Blockers

**None at this time.** All inputs are present (existing spec.md, R5
audit, master design doc, source agent bodies, template tree, common
protocol rule file). No external dependencies block plan-auditor review.

The lint rule for `expert-mobile.md` LR-02 (Agent in tools) is a post-R5
finding surfaced by research.md §1.3; it is documented as known-residual
and will be cleaned up by a future ORC-001 follow-on or an ad-hoc fix
PR. It does not block this SPEC's plan acceptance.

---

End of progress.md.
