---
id: SPEC-PIPELINE-FANOUT-ACTIVATION-001
title: Activate the plan/run/sync pipeline fan-out sites and resolve the delta re-audit contradiction
version: 0.1.0
status: draft
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: High
phase: plan
module: workflow-harness
lifecycle: spec-anchored
tags: "fan-out, pipeline, plan-auditor, tier-budget, template-mirror"
tier: M
---

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-02 | Initial plan-phase authoring. M1 scope only: fan-out index tables, discretionary-to-conditional promotion of the ten fan-out sites, D-1 contradiction fix, Tier size budget. | manager-spec |

---

## 1. Context

A pipeline audit established that all ten parallel fan-out sites across the `plan` / `run` / `sync`
workflows are phrased discretionarily (`MAY`), and that eight of the ten live inside sub-skill bodies
that the orchestrator only Reads *after* it has already entered the phase serially. The two
script-backed fan-outs were invoked zero times across 761 session transcripts.

Separately, the `plan-auditor` agent contradicts itself on re-audit scope: two surfaces say a
confirming re-audit is scoped to the enumerated defect delta, while the Retry Loop Contract in the
same file says iteration 2+ performs a full audit.

Full evidence, including the re-anchored measurements this SPEC's acceptance criteria depend on,
is in `research.md`. The originating report is untracked, so `research.md` is the durable record.

---

## 2. Requirements (GEARS)

### Fan-out visibility

**REQ-PFA-001** — The `plan`, `run`, and `sync` router skills shall each carry a Fan-Out Index table
enumerating that phase's fan-out sites by Fan-Out ID, trigger condition, target sub-skill file, and
the unit of work parallelised.

**REQ-PFA-002** — Each of the ten fan-out sites shall carry its canonical Fan-Out ID
(`FO-PLAN-1`, `FO-PLAN-2`, `FO-RUN-1` through `FO-RUN-4`, `FO-SYNC-1` through `FO-SYNC-4`) as inline
text, so a router index entry and its site are mutually traceable.

### Fan-out activation

**REQ-PFA-003** — **While** a fan-out site's stated precondition holds, the orchestrator shall run
that fan-out rather than treating it as a discretionary option.

**REQ-PFA-004** — **Where** a fan-out's capability precondition is absent — the backing script is not
on disk, the runtime does not support dynamic workflows, or the concurrency ceiling is reached — the
site shall carry a fail-open fallback sentence directing execution to the serial path with no error
and no warning. Nine of the ten sites already carry such a sentence and shall retain it unchanged in
meaning; the sync MX-tag-scan sharding site (FO-SYNC-2) currently carries none, and shall gain one
as part of its promotion, because promoting a site to a conditional obligation without an escape
hatch is precisely the hazard this requirement exists to prevent.

**REQ-PFA-005** — A promoted fan-out clause shall not state an unconditional obligation: the
promoted form shall remain bound to its stated condition, and shall not read as a bare MUST or SHALL
detached from that condition.

### Re-audit scope contradiction (D-1)

**REQ-PFA-006** — The `plan-auditor` Retry Loop Contract shall describe an iteration-2-or-later
re-audit as scoped to the enumerated defect delta plus a regression check over prior-iteration
defects, in agreement with the defect-list clause in the same file and with the run-phase execution
sub-skill.

**REQ-PFA-007** — The `plan-auditor` shall retain sole authority over the binding PASS/FAIL verdict;
the delta scope shall reduce re-audit cost only, and shall not permit an orchestrator
self-assessment to substitute for an auditor verdict.

### SPEC size budget

**REQ-PFA-008** — The SPEC Complexity Tier section shall state a per-tier size budget of 8 (Tier S),
16 (Tier M), and 25 (Tier L), applied independently to the requirement count and to the acceptance-
criterion count.

### Template mirror integrity

**REQ-PFA-009** — **Where** a file changed by this SPEC has a template mirror, the mirror shall
receive the same semantic change while its existing neutralization divergences are preserved
unchanged.

**REQ-PFA-010** — Text newly written into a template mirror shall not contain SPEC identifiers,
requirement or acceptance-criterion tokens, internal working dates, commit hashes, or internal
source-code paths.

---

## 3. Constraints

- **Documentation-only.** No Go source, no test code, no configuration schema changes.
- **Fail-open is non-negotiable.** Every promotion preserves the existing fallback sentence verbatim
  in meaning; a fan-out that cannot run must still degrade silently to the serial path.
- **Ordering.** The Fan-Out Index tables (REQ-PFA-001/002) are applied before the promotion
  (REQ-PFA-003), because the index is what makes a conditional obligation discoverable at phase
  entry. Applying the promotion first would raise judgment load without providing the aid.
- **Verdict authority is untouched.** `plan-auditor` and `sync-auditor` remain the sole owners of
  their binding verdicts; nothing in this SPEC lets a fan-out's aggregated evidence stand in for a
  verdict.
- **Concurrency ceiling unchanged.** Promoted fan-outs stay within the existing 3-5 concurrent
  `Agent()` ceiling and remain read-only; the single-writer property of every applier step is
  preserved.
- **Anchoring.** Because file line numbers in this repository drift between measurements, every
  verification anchors on a stable content token.

---

## 4. Exclusions

This section states what is deliberately NOT built. Items here are out of scope for this SPEC and
must not appear in its diff.

### Out of Scope — refutation step (D-3)

- Adding an adversarial refutation stage to any of the ten fan-outs. All ten remain
  "gather evidence, then judge". This is owned by a later milestone.
- Introducing a `plan-audit-4dim` script or any new fan-out script.

### Out of Scope — verify-snapshot activation (V2)

- Any change to the `moai verify` snapshot key policy, its TTL, or its storage path.
- Adding `verify record` or `verify check` hooks to any phase.

### Out of Scope — audit-stream unification (V3)

- Adding a plan-artifact-hash field to the `plan-auditor` output format.
- Any change that would make a plan-phase audit populate the run-phase audit cache.

### Out of Scope — adjacent drift in a touched file

- The stale sync-phase quality-gate hook description in the local `sync.md` (it says the hook exits
  2 to block; the template mirror correctly says it exits 0 and signals blocking via stdout JSON).
  This is a real local-vs-template divergence in a file this SPEC edits, and it is deliberately left
  alone so it is not silently swept into this diff. It is recorded in `research.md` § E.7.

### Out of Scope — fan-out sites not in the inventory

- The codemaps extraction fan-out. It uses the same discretionary phrasing but is not one of the ten
  pipeline sites, and is not promoted here.
- Non-fan-out discretionary phrasing elsewhere in the touched files, notably the tier-judgment skip
  condition in the plan sub-skill.

### Out of Scope — behavioural verification

- Proving that the promoted phrasing actually increases fan-out activation. That requires a
  post-change transcript sweep and is not an acceptance criterion of this SPEC.

---

## 5. Acceptance

Acceptance criteria, their runnable judging commands, and the observed pre-change baseline for each
are in `acceptance.md`.
