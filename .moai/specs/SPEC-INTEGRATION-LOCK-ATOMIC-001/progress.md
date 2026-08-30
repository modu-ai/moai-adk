---
id: SPEC-INTEGRATION-LOCK-ATOMIC-001
title: "Progress record — integration lock mutation atomicity (card t336)"
version: "0.1.1"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "integration-lock, atomicity, kanban, factory, t336"
tier: M
---

# Progress — SPEC-INTEGRATION-LOCK-ATOMIC-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t336. Tree `15453140a`, branch `WT-integration-lock-atomic`, worktree
  `.claude/worktrees/t336`.
- Artifacts authored at plan phase: `spec.md`, `plan.md`, `acceptance.md`, this file. Status
  `draft` across all four.
- Pre-flight consumed as-is: `.moai/reports/t336/preflight.md`. Its stated gap — no concurrent
  acquire was executed on this tree — is carried forward unchanged: the race is a code-path
  hypothesis until run-phase M1 observes it.
- Open design choice resolved at plan phase: option (b), a dedicated short-lived mutation lock
  reusing the existing platform substrate. Reasoning in `plan.md` §B.
- No `[NEEDS CLARIFICATION]` markers outstanding.
- Acceptance criteria: 9 (`AC-ILA-001`..`AC-ILA-009`); 6 MUST-PASS, 2 SHOULD-PASS, 1 CI-judged.
- No production or test Go code was written or modified at plan phase.

### plan-audit iteration 1 → repair (v0.1.1)

- Verdict read: `.moai/reports/t336/plan-audit.md` — FAIL, 0.803, six blocking findings (D1-D6),
  three optional (D7-D9). Must-pass firewall MP-1..MP-7 all passed; 14/14 citations resolved.
- All six blocking findings repaired; all three optional findings also applied (each was a
  one-line change that enlarged no scope).
- Two decisions taken during the repair, both recorded in the artifacts rather than left implicit:
  the deterministic interleaving hook was ADOPTED (D2's preferred fix) rather than declining it
  for a round ceiling; and the RED/GREEN pair was collapsed onto ONE shipped test function run in
  two tree states (D3+D4's first option), so no criterion names a test absent from the deliverable.
- Not changed, per the auditor's explicit "already sound" list: §A hypothesis framing, the 14
  verified citations, the traceability matrix and its §D.4 indirect-verification statement, the
  six-heading Exclusions block, the trailing-`kill` rejection, the choice of option (b), and M3's
  necessity.
- Scoped lint after the repair: `moai spec lint .moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/spec.md`
  → recorded in `.moai/reports/t336/spec-lint.txt`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
