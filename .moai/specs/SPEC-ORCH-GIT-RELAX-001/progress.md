# progress.md — SPEC-ORCH-GIT-RELAX-001

> Tier L progress artifact. §E section skeleton populated at plan-phase close; §E.2–§E.4 are placeholder-only per the manager-spec §E skeleton generation protocol.

---

## §A. SPEC summary

- **ID**: SPEC-ORCH-GIT-RELAX-001
- **Title**: Orchestrator-direct Tier S/M git ops + state-sensitive worktree recovery (manager-git relaxation)
- **Tier**: L (6-artifact set: spec.md + plan.md + acceptance.md + research.md + design.md + progress.md)
- **Phase**: plan (artifacts authored; awaiting plan-auditor + Implementation Kickoff Approval)
- **Scope**: relax "always delegate push+PR to manager-git" — Tier S/M push+PR + state-sensitive git ops move to orchestrator-direct; manager-git RETAINED for Tier L / `--pr` / Late-Branch 4-Phase closure.
- **Change-surface**: 14 enumerated locations (13 edit + 1 verified no-change; evidence §4's original 12 + iter-2 additions: `spec-workflow.md` Route B edit + `CLAUDE.md:131` no-change).
- **Triggering incident**: PR #1338 handling (manager-git primary→main restore failure + concurrent-session worktree cross-swap).

---

## §B. Plan-phase artifact manifest

- [x] `spec.md` — 16 REQ-OGR requirements (GEARS notation), 10 AC tokens (iter-2: +AC-OGR-009/010), §D Table 1 (14 enumerated: 13 edit + 1 no-change), §F Out of Scope (6 H3 sub-headings).
- [x] `plan.md` — 4 milestones (M1 doctrine core / M2 agent def + delegation / M3 Go verification / M4 regression + full gate), ordered by decision-reversibility. 1 `[NEEDS CLARIFICATION]` marker carried to Implementation Kickoff Approval (iter-2: marker #2 resolved IN-SPEC).
- [x] `acceptance.md` — 10 AC-OGR tokens (all MUST severity, iter-2: +009/010), Given-When-Then scenarios, §D Traceability (16/16 REQ-mapped, iter-2 closed 5-gap deficit), §E Definition of Done, §F edge cases.
- [x] `research.md` — evidence §4 + §5 incorporation + iter-2 audit (locations #13/#14), 12 verified URLs, §10 open questions (iter-2: item #2 resolved IN-SPEC).
- [x] `design.md` — context-sensitivity inversion principle, 7-principle gate scorecard, 6 rejected alternatives.
- [x] `progress.md` — this file (§E skeleton).

---

## §C. Pre-plan-phase audit checks (run by manager-spec before §E.1 sign-off)

- [x] SPEC ID regex check: `SPEC-ORCH-GIT-RELAX-001` → `PASS` (executed via Bash, verbatim output cited).
- [x] ID uniqueness: no existing SPEC under `.moai/specs/` collides with `SPEC-ORCH-GIT-RELAX-001` (closest neighbors: `SPEC-V3R6-ORCH-IGGDA-001`, `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001`, `SPEC-WORKTREE-BRANCH-GUARD-001` — different domain tokens, no collision).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). `phase: "v14.4.0 target"` — release target, not a lifecycle token; passes the prohibited-phase-value check.
- [x] Requirements in GEARS notation (Ubiquitous / When / While / Where / shall not) — no residual IF/THEN modality.
- [x] Out of Scope section satisfies `OutOfScopeRule`: 6 `### Out of Scope — <topic>` H3 sub-headings, each with ≥1 `-` bullet.
- [x] Artifact set matches Tier L (6 artifacts: spec + plan + acceptance + research + design + progress).
- [x] spec.md carries no implementation detail (no function names, no API schemas — REQ tokens only).

---

## §D. Open items (resolved at Implementation Kickoff Approval)

- [ ] **branch-guard opt-in default state** — is `Workflow.BranchGuard.Enabled` on or off for the maintainer checkout? (plan.md §B, research.md §10 item 1)
- [x] **`MOAI_BRANCH_GUARD_EXEMPT=1` lifecycle** — RESOLVED IN-SPEC (iter-2, D4): per-invocation inline mandated per design.md §3.3 + plan.md §G.

One operational detail remains, carried to Implementation Kickoff Approval. Neither blocks plan-phase audit-readiness.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-04
spec_id: SPEC-ORCH-GIT-RELAX-001
tier: L
artifact_count: 6
req_count: 16
ac_count: 10   # iter-2: +AC-OGR-009 (catalog+frontmatter preservation), +AC-OGR-010 (foreign-session auto-isolation)
open_clarifications: 1   # iter-2: marker #2 (env-var lifecycle) resolved IN-SPEC per design.md §3.3; marker #1 (branch-guard opt-in) remains
evidence_base: .moai/reports/agent-skill-hook-redesign-evidence-20260804.md
notes: >
  Tier L redesign SPEC, first of a phased split ("순차 분할").
  manager-git relaxation only; skills/hooks cleanup in follow-up SPECs.
  Change-surface is 14 enumerated locations (13 edit + 1 verified no-change):
  evidence §4's original 12 + iter-2 additions (spec-workflow.md Route B edit;
  CLAUDE.md:131 no-change). Env exemption path expected to admit
  orchestrator-direct Tier S/M with no Go change.
  Iter-2 plan-auditor FAIL 0.79 → 5 defects D1-D5 addressed (no commit/push).
```

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop populates>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates>_

---

## §F. (reserved — Phase 4 Mode Selection is logged here by the orchestrator before the first run-phase Agent() spawn)

_<pending orchestrator — Phase 4 mode selection log>_

---

## §G. Risk register

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Env exemption path does NOT admit orchestrator-direct (M3 fails) | Low (code already in context confirms env branch fires first) | Medium — minimal Go change required | M3 verification + REQ-OGR-012 minimal-change clause |
| Doctrine cross-reference drifts post-edit (M4.2 grep audit fails) | Medium (12 locations, easy to miss one) | High — inconsistent doctrine is worse than the prior doctrine | M4.2 grep audit + AC-OGR-003 binary check |
| Late-Branch body accidentally edited (AC-OGR-007 fails) | Low (milestone explicitly excludes it) | Medium — regression in canonical Tier L closure | AC-OGR-007 byte-identity diff |
| Template-mirror forgotten for a template-mirrored file | Medium (4 mirrored files across M1+M2) | High — CI template-neutrality / mirror-parity fails | M2 verify block lists all 4 mirrors; AC-OGR-005 full gate |
| LOCAL-ONLY file accidentally mirrored | Low (explicit no-mirror list) | Medium — violates §24/§25 | AC-OGR-008 no-mirror audit |
| Over-relaxation (manager-git abolished instead of sculpted) | Low (design.md §6.1 rejects) | High — loses Tier L owner | REQ-OGR-008 + REQ-OGR-016 forbid abolition |

---

## §H. Recursive Self-Diagnosis Log

_<pending run-phase — manager-develop / orchestrator (DIAGNOSE-PATCH-VERIFY mechanical failures)>_

---

## §I. Token Accounting

_<pending sync-close — manager-docs invokes the token-accounting writer>_
