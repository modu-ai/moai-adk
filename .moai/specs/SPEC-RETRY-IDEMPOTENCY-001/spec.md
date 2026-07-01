---
id: SPEC-RETRY-IDEMPOTENCY-001
title: "Retry Safety Asymmetry — Observe-Before-Retry Gate for Side-Effecting Tool Calls"
version: "0.1.0"
status: completed
created: 2026-07-01
updated: 2026-07-01
era: V3R6
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tags: "retry, idempotency, error-recovery, side-effects, rule-augmentation"
tier: S
---

# SPEC-RETRY-IDEMPOTENCY-001 — Retry Safety Asymmetry

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-01 | manager-spec | Initial draft — plan-phase artifacts (Tier S, 4 artifacts) |

## §A. Context & Motivation

### A.1 The confirmed gap

The MoAI error-recovery policy treats retry as a **uniform** operation that does not
distinguish a tool call's side-effect profile. Two surfaces express the current uniform
policy:

- `moai-constitution.md` § Error Handling Protocol — "Maximum 3 retries per operation".
- `agent-common-protocol.md` § Error Recovery Pattern — a 4-step list ending in
  "do not retry the identical call" (step 3) and "After 3 failures on the same operation,
  report the blocker" (step 4).

Neither surface distinguishes **read-only / idempotent** operations from
**side-effecting** operations when deciding whether a retry is safe. A full-repository
search for the terms `idempotent`, `side-effect`, `non-idempotent`, `retry safe`, and
`observe-before-retry` across both the deployed rule tree (`.claude/rules/`) and the
template mirror (`internal/template/templates/.claude/rules/`) returned zero matches.
The retry-safety-asymmetry principle is genuinely absent.

### A.2 Why the gap matters

Retry safety is **asymmetric with respect to a tool call's side effects**:

- An **idempotent / read-only** operation (re-reading a file, re-running a search,
  re-querying, re-running an initializer, fetching a URL) can be retried up to the
  existing 3-retry ceiling without consequence. A transient failure — a file lock, a
  network blip, a momentary resource contention — is legitimately recovered by a retry.
  Repeating the operation produces the same observable result as running it once.

- A **side-effecting** operation (creating/editing a file, committing, pushing, opening
  a pull request, deploying, mutating external-API state) carries a duplicate-effect
  hazard. When such a call fails **ambiguously** — the failure signal arrives, but
  whether the effect already landed is unknown — a blind retry risks applying the effect
  a second time: a duplicate commit, a duplicate pull request, a double deploy. The
  absence of a success signal is not evidence the effect did not land.

The missing principle is a gate: before retrying a side-effecting call whose success is
uncertain, the actor MUST first **observe** current state to determine whether the effect
already occurred, and only retry when the effect is confirmed absent.

### A.3 Relationship to existing rules (delta, not restatement)

This SPEC augments — it does not replace — the existing policy. It carves a precise delta
against three adjacent rules so it duplicates none of them:

- **`agent-common-protocol.md` § Error Recovery Pattern step 3** ("do not retry the
  identical call") is the closest existing rule. The new principle **extends** it along
  the side-effect axis: step 3 already discourages a blind identical retry; the new
  principle adds the observe-before-retry gate specifically for side-effecting calls. The
  augmentation references step 3 rather than restating it.
- **`agent-common-protocol.md` § Ledger Closure** (the ledger-closure invariant: an
  aborted `Agent()` delegation MUST NOT leave a dangling `tool_use`) addresses a
  **different concern** — recovery from an *interrupt/abort* of a delegation, not the
  retry of an actor's own failed tool call. The two do not overlap.
- **`agent-common-protocol.md` § Pre-Spawn Sync Check** addresses a **different concern**
  — a *concurrency race* between parallel sessions committing to a shared tree, not the
  retry of a single actor's own uncertain call. The two do not overlap.

## §B. GEARS Requirements

### REQ-RI-001 — Retry-safety asymmetry principle (Ubiquitous)

The error-recovery rule shall classify each retryable tool call by its side-effect
profile — **idempotent / read-only** versus **side-effecting** — and shall state that
retry safety is asymmetric with respect to that profile.

### REQ-RI-002 — Idempotent retry allowance (Ubiquitous)

The error-recovery rule shall state that an idempotent or read-only operation MAY be
retried up to the existing 3-retry ceiling, because repeating it produces the same
observable result and legitimately recovers a transient failure.

### REQ-RI-003 — Observe-before-retry gate for side-effecting calls (Event-driven)

When a side-effecting tool call fails ambiguously — the failure signal is present but
whether the effect already landed is uncertain — the actor shall first observe the
current state to determine whether the effect already occurred, and shall retry the call
only when the effect is confirmed absent.

### REQ-RI-004 — Duplicate-effect hazard is named (Unwanted behavior)

The error-recovery rule shall state that retrying a side-effecting call without first
observing state is the duplicate-effect hazard, and the actor shall not blindly retry a
side-effecting call whose success is uncertain.

### REQ-RI-005 — Preserve the existing uniform 3-retry ceiling (Ubiquitous)

The augmentation shall preserve the existing "Maximum 3 retries per operation" ceiling and
the existing 4-step Error Recovery Pattern verbatim; it shall be a boundary refinement
layered on top of the existing steps, not a replacement of the 3-retry rule.

### REQ-RI-006 — Reference the closest existing rule (Ubiquitous)

The augmentation shall reference the existing step 3 ("do not retry the identical call")
as the rule it extends along the side-effect axis, so the new principle carves a delta
rather than duplicating adjacent policy.

### REQ-RI-007 — Template-mirror parity of the augmentation block (Ubiquitous)

Where `.claude/rules/` is deployed to user projects, the **augmentation block only** (the
§ Error Recovery Pattern insertion — NOT the whole file) shall appear byte-identically in
both the deployed file (`.claude/rules/moai/core/agent-common-protocol.md`) and the template
mirror (`internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`), and
`make build` shall be run so the embedded template reflects the mirror.

> **Whole-file byte-parity does NOT hold for this file — it is leak-test-covered, not
> byte-parity-covered.** `agent-common-protocol.md` is NOT enrolled in the byte-parity
> allowlist `workflowOptMirroredPaths` (`internal/template/rule_template_mirror_test.go`); it
> is covered by the internal-content leak test (`internal_content_leak_test.go`) instead. The
> deployed copy legitimately carries internal SPEC-ID / REQ / AC tokens (e.g. a
> hook-recovery-signal forward-link, ledger-closure REQ/AC tokens) that the template mirror
> deliberately strips per `CLAUDE.local.md` §25, so the two files diverge by roughly three
> dozen lines at baseline and CANNOT be made whole-file identical. Parity is therefore scoped
> to the augmentation block's § Error Recovery Pattern region, verified by a region-scoped
> diff — NOT a whole-file diff. A run-phase agent MUST NOT attempt whole-file byte-parity: it
> would either false-fail on the pre-existing divergence, or illegally copy internal SPEC-IDs
> into the neutral template and trip the leak guard.

### REQ-RI-008 — Deployed-rule neutrality (Unwanted behavior)

The deployed rule text shall not contain any moai-adk-internal development trace — no
SPEC-ID token, no REQ/AC token, no external-commentary citation, no author name, no
internal date, no commit SHA. The augmentation shall be generic prose describing the
mechanism only; all provenance shall live exclusively in the `.moai/specs/` artifacts
(which are not deployed).

## §C. Exclusions

### Out of Scope — completion-declaration sentence promotion

- A second, weaker gap exists: the "do not declare success before running the
  verification" judgment sentence lives only in the non-deployed `CLAUDE.local.md`, while
  the deployed rules carry only the principle form (the "no unobserved success claim"
  invariant in `verification-claim-integrity.md`). Promoting that judgment sentence into a
  deployed rule is **out of scope** for this SPEC. This SPEC reflects only the retry gap.

### Out of Scope — changing the retry ceiling

- This SPEC does not change the "Maximum 3 retries per operation" ceiling, the poll/retry
  interval, or any numeric retry parameter. It is a side-effect-axis refinement only; the
  count remains 3.

### Out of Scope — runtime / hook enforcement

- No Go code, PreToolUse hook, or mechanical detector is added to enforce the
  observe-before-retry gate. This is a doctrine-layer rule augmentation only. Mechanical
  enforcement, if ever desired, is deferred to a separate follow-up SPEC.

### Out of Scope — constitution.md edit

- The `moai-constitution.md` § Error Handling Protocol "Maximum 3 retries" line is left
  unchanged. The augmentation targets only `agent-common-protocol.md` § Error Recovery
  Pattern; touching a second rule file is unnecessary and would violate scope discipline.

## §D. Acceptance Criteria (summary)

Full Given-When-Then scenarios and testable ACs are enumerated in `acceptance.md`. In
summary, the run-phase is accepted when: the idempotency-asymmetry augmentation block is
present in both the deployed and template copies (region-scoped byte-identical — NOT
whole-file, per REQ-RI-007), the augmentation keywords appear in the regenerated
`embedded.go`, the existing 4-step pattern and 3-retry ceiling are unchanged, the required
keywords are present, and the CI internal-content leak guard plus the template-neutrality
guard both pass.

## §E. Lifecycle Signal (progress.md is the SSOT)

The plan/run/sync audit-ready signal skeleton lives in `progress.md` (§E.1 through §E.4).
This section is a pointer only; do not duplicate the signal tables here.
