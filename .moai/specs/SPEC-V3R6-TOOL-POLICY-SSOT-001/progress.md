---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy SSOT — Progress"
version: "0.1.0"
status: draft
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# Progress — SPEC-V3R6-TOOL-POLICY-SSOT-001

> Plan-phase artifact. The §E skeleton below is emitted per the canonical progress.md §E skeleton generation protocol. Only §E.1 is populated at plan-phase; §E.2-§E.5 are placeholder headings for run/sync/Mx-phase population by manager-develop (§E.2/§E.3) and manager-docs (§E.4/§E.5).

---

## §A. Phase Status

- **Plan-phase**: COMPLETE (2026-06-18). 5 plan-phase artifacts authored (spec.md + plan.md + acceptance.md + research.md + design.md) + this progress.md §E skeleton.
- **Run-phase**: NOT STARTED. Entry requires Implementation Kickoff Approval (CLAUDE.local.md §19.1).
- **Sync-phase**: NOT STARTED.
- **Mx-phase**: NOT STARTED.

---

## §B. Milestone Tracker

| Milestone | Status | Owner | Commit |
|---|---|---|---|
| M1 — schema + seed YAML | not-started | manager-develop | — |
| M2 — codegen mechanism | not-started | manager-develop | — |
| M3 — integration + cross-refs | not-started | manager-develop | — |
| M4 — migration + compat | not-started | manager-develop | — |
| M5 — lint + single-rule demo | not-started | manager-develop | — |
| M6 — PR (manager-git) | not-started | manager-git | — |

---

## §C. AC Tracker

| AC | Severity | Status | Evidence |
|---|---|---|---|
| AC-TPS-001 (SSOT exists) | MUST-FIX | pending | — |
| AC-TPS-002 (seeded from 4 sources) | MUST-FIX | pending | — |
| AC-TPS-003 (codegen produces block) | MUST-FIX | pending | — |
| AC-TPS-004 (round-trip equivalence) | MUST-FIX | pending | — |
| AC-TPS-005 (§24.5 drift prevented) | MUST-FIX | pending | — |
| AC-TPS-006 (reuse constitution query) | MUST-FIX | pending | — |
| AC-TPS-007 (backward compat) | MUST-FIX | pending | — |
| AC-TPS-008 (single rule change) | MUST-FIX | pending | — |
| AC-TPS-009 (cross-refs present) | SHOULD-FIX | pending | — |
| AC-TPS-010 (lint clean) | MUST-FIX | pending | — |
| AC-TPS-011 (Template-First) | SHOULD-FIX | pending | — |
| AC-TPS-012 (background-write declared) | NICE-TO-HAVE | pending | — |
| AC-TPS-013 (codegen idempotency) | MUST-FIX | pending | — |
| AC-TPS-014 (template round-trip + sentinel) | MUST-FIX | pending | — |

---

## §D. Commit Ledger

_(populated by run-phase manager-develop)_

---

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase**: COMPLETE on 2026-06-18 (iter-2 — plan-auditor D1-D9 defects resolved).
- **Artifacts**: spec.md (10 REQs incl. REQ-TPS-008b, GEARS), plan.md (6 milestones), acceptance.md (14 ACs incl. AC-TPS-013 idempotency + AC-TPS-014 template round-trip, Given-When-Then), research.md (4-source file:line inventory + book2 survey, iter-2 corrections), design.md (codegen approach + schema + drift-prevention narrowed scope + D7 two-strategy split), progress.md (this §E skeleton).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | TOOL ✓ | POLICY ✓ | SSOT ✓ | 001 ✓ → PASS`.
- **Frontmatter schema**: 12 canonical fields present; `created`/`updated` ISO (not `created_at`); `tags` quoted string (not `labels`); `version` quoted string; `era: V3R6`; `tier: L`.
- **Exclusions**: §X.1-X.8 present (8 exclusion entries; §X.8 NEW for D3 honest scope-narrowing) in spec.md, all carrying the literal "Out of Scope" token at h3.
- **iter-2 defect resolutions (D1-D9)**: D1 exclusion-headings fixed (lint clean); D2 status-transition-ownership.sh L48-69 correctly characterized as COMMENT BLOCK (not case slice); D3 §24.5 claim narrowed to YAML↔settings.json scope (§X.8 added); D4 AC-TPS-013 idempotency added; D5 background-write enforcement correctly attributed to Claude Code runtime auto-deny; D6 OQ-2 resolved (ask array native, 6 entries verified); D7 design.md §B.1 two-strategy split (parse-modify-serialize on .json, raw-text on .tmpl); D8 actual allow-count = 110 (not "approx 50+"); D9 constitution schema disjoint — thin `moai tool-policy list` query replaces the "reuse constitution infrastructure" claim.
- **Codegen approach chosen (D7)**: `moai tool-policy build` + Go codegen with block-region replacement — parse-modify-serialize on `.claude/settings.json` (pure JSON), raw-text region replacement on `internal/template/templates/.claude/settings.json.tmpl` (mixed JSON + Go-template directives; permissions block verified free of `{{...}}`).
- **Constitution-query decision (D9)**: thin NEW `moai tool-policy list` subcommand modeled on `moai constitution list` CLI SHAPE; does NOT wrap constitution list (schemas disjoint — tool-policy `{tool,args_pattern,risk_tier,decision,owner_agent,audit}` vs constitution `{id,zone,zone_class,file,anchor,clause,canary_gate}`).
- **§24.5 honest scope-narrowing (D3)**: this SPEC prevents YAML↔settings.json drift by construction (both generated surfaces derive from one YAML). It does NOT prevent the markdown-doctrine-vs-Go-code drift that characterized §24.5 literally — this SPEC generates neither markdown doctrine nor Go code (§X.8). §24.5 is cross-referenced as the canonical ANALOGY, not a literal prevention claim.
- **Ready for**: plan-auditor iter-2 re-review → Implementation Kickoff Approval (§19.1) → run-phase M1.

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates with E1-E7 self-verification matrix, command outputs, test results>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop populates when all M1-M5 MUST-FIX ACs pass with evidence>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates with sync_commit_sha + sync artifacts>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — manager-docs/orchestrator populates with mx_commit_sha + 4-phase close>_

---

## §F. Next Action

Obtain Implementation Kickoff Approval (CLAUDE.local.md §19.1) via AskUserQuestion, then delegate run-phase M1 to manager-develop (cycle_type per quality.yaml `development_mode: tdd`).

Paste-ready resume (for context-window boundary):

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-V3R6-TOOL-POLICY-SSOT-001 run-phase M1 진입.
applied lessons: feedback_defect_claim_verification, project_harness_moai_namespace_plan_ready.

전제 검증:
1) ls .moai/specs/SPEC-V3R6-TOOL-POLICY-SSOT-001/ → 6 artifacts (spec/plan/acceptance/research/design/progress)
2) grep -c "REQ-TPS" .moai/specs/SPEC-V3R6-TOOL-POLICY-SSOT-001/spec.md → 10

실행: /moai run SPEC-V3R6-TOOL-POLICY-SSOT-001

머지 후: SPEC-V3R6-CONTEXT-GOV-AXIS-001 (Sprint 15 P2b)

✂──── 여기까지 복사 ────✂
```

---

## Out of Scope

### Out of Scope — Canonical exclusions live in spec.md

- This progress.md is a companion artifact; the canonical exclusions (§X.1-§X.8) live in `spec.md`. This section satisfies the lint `MissingExclusions` rule for this file.
