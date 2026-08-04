---
id: SPEC-FOURDIM-PHANTOM-001
title: "sync-audit-4dim workflow phantom-mechanism guard — deterministic probe → FAIL verdict for claimed-but-absent mechanisms"
version: 0.1.0
status: completed
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P0
phase: "v3.x target"
module: ".claude/workflows/sync-audit-4dim.js + internal/template/templates mirror"
lifecycle: spec-anchored
tags: "sync-audit-4dim, phantom-mechanism, falsification, verdict-guard, verification-claim-integrity, dynamic-workflow, template-mirror, autonomy-epic"
related_specs: [SPEC-SYNC-AUDIT-FALSIFICATION-001, SPEC-AUDIT-SNAPSHOT-001, SPEC-INFINITE-GOAL-001]
tier: S
---

# SPEC-FOURDIM-PHANTOM-001 — sync-audit-4dim Phantom-Mechanism Guard

## §A Background

`sync-audit-4dim.js` is the MoAI-shipped dynamic-workflow script that performs the sync-phase 4-dimension quality read (Functionality / Security / Craft / Consistency). Its verdict pipeline is:

1. **Context** — one read-only `Explore` agent extracts the audit surface (SPEC id, acceptance criteria, changed files, test command).
2. **Judge** — four parallel read-only `Explore` judges (`effort: xhigh`, schema-forced), one per dimension, each scoring 0–1 with command+verbatim-output evidence.
3. **Verdict** — in-script JS aggregates the four scores via harmonic mean, gated by a null-judge guard (→ `INCOMPLETE`) and a zero-score guard (→ `FAIL`).

Per SPEC-AUDIT-SNAPSHOT-001 (A3 PROMOTED), the workflow's verdict is BINDING on the happy PASS path; the cold `sync-auditor` subagent is the FALLBACK for failure modes. Because a PASS bypasses the cold auditor entirely, the workflow is the sole gate on the happy path — which makes any dishonest PASS a load-bearing defect.

### Provenance — IMP-5

The predecessor **SPEC-SYNC-AUDIT-FALSIFICATION-001** (merged PR #1344, IMP-1/3/6) hardened the AGENT side (`sync-auditor.md` body) against falsified AC-mechanism claims: a sync-auditor that accepts a SPEC's *declared* mechanism as evidence of the mechanism's *operation* issues a falsified PASS. IMP-1/3/6 bound the agent to probe the actual write surface and treat any declared mechanism with zero on-disk evidence as a falsification.

**IMP-5** is the WORKFLOW-JS counterpart deferred to this SPEC. The `sync-audit-4dim.js` Runner has the same exposure: a SPEC under audit may declare a mechanism (`@MX:ANCHOR`, an input-validation guard, a migration call-site) in its plan/acceptance artifacts, and the four dimension judges — reading the declaration in the Context payload — may score on the assumption that the declared mechanism is real. A phantom mechanism (declared but absent from the diff / produced files) can ride a falsified PASS through the harmonic mean.

This SPEC adds a deterministic **phantom-mechanism guard** to the verdict pipeline. The guard sits between the zero-score guard and the harmonic mean; it executes a SPEC-declared probe against the actual write surface and forces a hard FAIL naming the phantom mechanism when the probe finds zero matches. The guard COMPOSES WITH the existing guards; it does not replace them and does not add a 5th dimension.

### Why this is workflow-side, not agent-side

The predecessor's IMP-1/3/6 bound the agent because the agent is the verdict owner on the failure path. This SPEC binds the workflow because the workflow is the verdict owner on the HAPPY path — and the happy path is exactly the path a phantom PASS exploits. Closing only the agent side leaves the binding-PASS surface unprotected.

### Applied lessons

- `feedback_ac_stated_mechanism_can_be_false` — a SPEC's declared mechanism is a claim, not evidence; the verifier MUST probe the actual write surface.
- `feedback_lsel_f1_actual_write_surface` — an allowlist auditing only the declared `target_surface` (statement of intent) misses the actual edit path; the probe MUST run against the diff.patch paths + produced files.
- `verification-claim-integrity.md` §1.1 surface 3 (defect / debt / drift claim) + §2 (baseline-integrity attribution) — the probe command IS the command, `actual_matches: 0` IS the observed output, named per-mechanism in the FAIL payload.

## §B Requirements (GEARS)

### REQ-FP-001 — Claimed-mechanism capture field
**Where** a SPEC under audit declares one or more defensive / structural mechanisms in its plan or acceptance artifacts, the Context payload SHALL carry an optional `claimed_mechanisms[]` array; each entry SHALL name the mechanism, the literal probe command, and the expected match substring.

### REQ-FP-002 — Probe execution at Context-extraction time
**When** the Context agent extracts the audit surface, the Context agent SHALL execute each declared `probe_command` against the actual write surface (the diff.patch paths plus the produced files in the working tree — NOT the declared `target_surface` intent alone) and SHALL return a `probe_results[]` array carrying `{name, probe_command, expected_match_substring, actual_matches}` per claimed mechanism.

### REQ-FP-003 — Phantom-mechanism guard in the verdict pipeline
**Where** the verdict phase runs, the phantom-mechanism guard SHALL execute only after the null-judge guard and the zero-score guard have both declined to return a verdict, and before the harmonic-mean computation.

### REQ-FP-004 — Hard FAIL on phantom mechanism
**When** any claimed mechanism's `actual_matches` equals zero, the verdict phase SHALL return verdict `FAIL` carrying the phantom mechanism(s) named in a `phantom_mechanisms[]` array; the verdict SHALL NOT fall through to the harmonic-mean computation.

### REQ-FP-005 — Deterministic verdict logic
The phantom-mechanism guard SHALL be implemented as in-script JS (the `actual_matches == 0` comparison is deterministic JS, not an LLM judgment); only probe EXECUTION is delegated to the Context agent because the script body has no shell/filesystem access per `dynamic-workflows.md` § How a Workflow Runs.

### REQ-FP-006 — Happy-path passthrough
**While** every claimed mechanism's `actual_matches` is greater than zero, the phantom-mechanism guard SHALL NOT alter the verdict; the verdict phase SHALL fall through to the harmonic-mean computation unchanged.

### REQ-FP-007 — Per-mechanism baseline attribution
The FAIL payload for a phantom-FAIL SHALL carry, per phantom mechanism, the `probe_command` (the command that was run) and `actual_matches: 0` (the observed output), satisfying `verification-claim-integrity.md` §2 baseline-integrity attribution.

### REQ-FP-008 — Composable, non-replacing guard
The phantom-mechanism guard SHALL NOT replace the null-judge guard, the zero-score guard, or the harmonic-mean computation; the four-dimension enum (Functionality / Security / Craft / Consistency) SHALL remain FROZEN at four — the phantom guard adds a verdict guard, NOT a fifth dimension.

### REQ-FP-009 — Template neutrality
The distributed JS's phantom-guard comments and prose SHALL be GENERIC — no SPEC IDs, no REQ tokens, no "IMP-5" references, no internal dates, no commit SHAs; the guard SHALL be named by what it does ("phantom-mechanism guard") per CLAUDE.local.md §25 Template Internal-Content Isolation.

### REQ-FP-010 — Template-First mirror parity
**When** the live `.claude/workflows/sync-audit-4dim.js` is edited, the template mirror `internal/template/templates/.claude/workflows/sync-audit-4dim.js` SHALL be edited byte-identically and `make build` SHALL be run to regenerate the embedded catalog per CLAUDE.local.md §2 [HARD] Template-First Rule.

## §C Acceptance Criteria (summary)

Acceptance criteria with Given-When-Then scenarios live in `acceptance.md`. Summary:

- **AC-FP-001** — Phantom-FAIL: a claimed mechanism absent from the diff yields verdict `FAIL` naming it in `phantom_mechanisms[]`.
- **AC-FP-002** — Happy-path passthrough: a claimed mechanism with non-zero probe count passes through to the harmonic-mean verdict unchanged.
- **AC-FP-003** — Precedence: null-judge (INCOMPLETE) and zero-score (FAIL) guards still fire before the phantom guard.
- **AC-FP-004** — Baseline attribution: the FAIL payload carries `probe_command` + `actual_matches: 0` per phantom mechanism.
- **AC-FP-005** — Template-First: live + mirror are byte-identical after `make build`.

## §D Constraints

- **Determinism** — the verdict logic is pure JS; the script body reads no wall-clock and draws no random value (resume-cache safe per `dynamic-workflows.md`).
- **Read-only** — the Context agent is `agentType: 'Explore'`; probe execution uses read-only Bash (grep / cat / git diff). No Write/Edit is granted.
- **No mid-run user input** — workflow agents cannot prompt the user; a missing probe result is an evidence_gap, never a question (`agent-common-protocol.md` § User Interaction Boundary).
- **No LLM arithmetic** — the `actual_matches == 0` comparison and the harmonic mean are both JS, deterministic and auditable.
- **Composable** — the phantom guard DOES NOT replace any existing guard or add a 5th dimension; the 4-dimension enum is FROZEN.
- **Trust boundary** — the phantom guard has a TIGHTER trust boundary than the harmonic mean (a literal command + deterministic count vs a subjective 0–1 score).
- **§25 neutrality** — distributed JS carries no SPEC IDs / REQ tokens / IMP-5 / dates / SHAs.

## §E Scope

In scope:
- `.claude/workflows/sync-audit-4dim.js` (live) — verdict-phase phantom-mechanism guard + Context-schema `claimed_mechanisms` capture field + Context-prompt probe-execution instruction.
- `internal/template/templates/.claude/workflows/sync-audit-4dim.js` (mirror) — byte-identical.
- `make build` to regenerate the embedded catalog.

### Out of Scope — Agent-side falsification

- `.claude/agents/moai/sync-auditor.md` body changes — OWNED by the predecessor SPEC-SYNC-AUDIT-FALSIFICATION-001 (IMP-1/3/6, merged PR #1344). This SPEC touches only the workflow JS.

### Out of Scope — 5th audit dimension

- Adding a 5th audit dimension (e.g. "Truthfulness") to the DIMENSIONS enum. The phantom guard is a VERDICT-PHASE guard, not a new judge. The 4-dimension enum is FROZEN at Functionality / Security / Craft / Consistency.

### Out of Scope — Probe-authoring UX

- Authoring ergonomics for `claimed_mechanisms[]` in the SPEC template (tooling, linting, autocomplete). The probe format is a deterministic specification consumed by the workflow; how a SPEC author comes up with the probe is a documentation concern, not this SPEC's deliverable.

### Out of Scope — Probe-fidelity escalation

- A dedicated 6th read-only probe agent (considered in plan.md § Design Tension, rejected as costlier without correctness gain). Revisit only if Context-agent probe fidelity becomes a measured concern.

## §F History

- 2026-08-05 — plan-phase artifacts authored (spec.md / plan.md / acceptance.md / progress.md). Provenance: SPEC-SYNC-AUDIT-FALSIFICATION-001 IMP-5 deferral + 3 applied lessons (declared-mechanism-can-be-false, actual-write-surface, verification-claim-integrity §1.1 surface 3 + §2).

## §G Cross-References

- Predecessor: `.moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/` (IMP-5 deferral, agent-side counterpart).
- Binding-verdict promotion: SPEC-AUDIT-SNAPSHOT-001 (A3 PROMOTED — happy-path verdict is binding, cold auditor is fallback).
- Falsification precedent: SPEC-INFINITE-GOAL-001 AC-011.
- Applied lessons: `feedback_ac_stated_mechanism_can_be_false`, `feedback_lsel_f1_actual_write_surface`, `feedback_full_test_suite_verification` (full-suite verification discipline).
- Verification doctrine: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2.
- Dynamic-workflow primitive: `.claude/rules/moai/workflow/dynamic-workflows.md` § How a Workflow Runs (script body has no FS / shell access).
- Template-First: `CLAUDE.local.md` §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation.
