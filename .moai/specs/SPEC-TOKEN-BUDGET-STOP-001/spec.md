---
id: SPEC-TOKEN-BUDGET-STOP-001
title: "Token Budget Graceful-Abort + /tmp Evidence Persistence"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/runtime/budget.go"
lifecycle: spec-anchored
era: V3R6
tags: "token-economy, budget, graceful-abort, handoff, file-redirect"
---

# SPEC-TOKEN-BUDGET-STOP-001 — Token Budget Graceful-Abort + /tmp Evidence Persistence

**Epic**: Token-Economy Epic (4-SPEC A→B→C→D). This SPEC is **D of 4** (the final SPEC).

- A = SPEC-TOKEN-ACCOUNTING-001 (closed, f88d0226f) — per-SPEC token accounting
- B = SPEC-TOKEN-ROUTING-001 (closed) — Tier×Phase declarative model/effort routing
- C = SPEC-TOKEN-VERIFY-DIET-001 (closed, 3bb890f75) — file-redirect contract for verification output (doctrine-primary)
- **D = this SPEC** — budget graceful-abort + paste-ready handoff + /tmp evidence persistence

Closing D completes the Token-Economy Epic (4/4 → 100%).

**Tier**: M (standard) — see plan.md §A.4 for evidence.

**era**: V3R6 (modern 3-phase close: plan→run→sync).

---

## User Story

**As a** MoAI orchestrator running a long-horizon agent whose per-agent token usage has reached the hard_clear_threshold,
**I want** the runtime to auto-generate a paste-ready resume message (conforming to session-handoff.md 6-block format) and recommend turn termination WITHOUT hard-failing or auto-invoking `/clear`,
**so that** the session ends gracefully with a recoverable handoff — preserving the BC-V3R3-006 warning-first policy and the `/clear` NEVER auto-invoked HARD constraint.

**And as a** MoAI orchestrator citing verification evidence in a Verification Matrix or §E self-verification block,
**I want** the file-redirect contract's cited evidence path to persist beyond `/tmp` clearance,
**so that** `verification-claim-integrity.md` §1.1 surface 1+2 (direct-observation obligation) is preserved at audit time even after the OS clears `/tmp`.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

Per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3, a defect claim must cite a measured source. The three gaps below each name the measured file, the measured line range, and the observed pattern.

### GAP-1 — IsAtHardLimit exists but no graceful-abort action is wired

- **Measured source**: `internal/runtime/budget.go` lines 122-134 (`IsAtHardLimit` method) + lines 66-91 (`RecordCall` — warning-only, no abort path).
- **Observed pattern**: `IsAtHardLimit(agentName) bool` returns true when an agent's cumulative usage ≥ `hard_clear_threshold` (default 0.90, config.go:29). The method exists and is unit-tested (budget_test.go:106-122, `TestHardLimitAt90Pct`), but NO downstream runtime path is wired to this signal — it is never called by `RecordCall` or any other method to trigger a graceful-abort, handoff generation, or turn-termination recommendation. The doc comment at budget.go:16 states: *"Hard-fail is deferred to Phase 5 (future SPEC)."* — D is that future SPEC, fulfilling the **graceful** path (not true hard-fail; see §Out of Scope — True hard-fail).
- **Implementation gap**: The signal exists; the action does not. D wires the signal to a graceful-abort + handoff-generation path.

### GAP-2 — /tmp evidence is non-persistent (EC-2 residual risk from VERIFY-DIET-C)

- **Measured source**: `.claude/rules/moai/core/agent-common-protocol.md` lines 290-311 (canonical 7-item batch — all commands redirect to `/tmp/moai-verify/<N>-<slug>.log`); `.claude/output-styles/moai/moai.md` line 413 (`└─ evidence: /tmp/moai-verify/<session>/`) + line 613 (`📎 Evidence: /tmp/moai-verify/<session>/`).
- **Observed pattern**: The file-redirect contract (authored in VERIFY-DIET-C) writes verbatim evidence to `/tmp/moai-verify/`. `/tmp` is OS-cleared periodically (macOS reboot, Linux tmpfs re-mount, systemd-tmpfiles). When `/tmp` is cleared, the cited path no longer resolves to a file → evidence unreachable at audit time → `verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) + surface 2 (manager §E self-verification) violated. VERIFY-DIET spec.md §Out of Scope (lines 158-161) explicitly deferred the persistence layer to a separate follow-up: *"A future hook that mechanically redirects Bash output to disk and validates the cited path resolves (file-redirect contract's enforcement layer). This SPEC authors the doctrine contract; the enforcement hook is a separate follow-up SPEC."* — D owns the **persistence obligation** (NOT the mechanical hook; see §Out of Scope — Mechanical enforcement hook script).
- **Persistence gap**: Evidence lives in `/tmp` (non-persistent). D extends the contract to obligate persistence under `.moai/state/verify/<session>/` (gitignored runtime state, same area as `context-usage.json` + `active-sessions.json`).

### GAP-3 — PersistProgress exists but is not triggered on hard-limit

- **Measured source**: `internal/runtime/persist.go` line 18 (`func (t *Tracker) PersistProgress(specID, roundLabel, approach, nextStep string) (string, error)`); `internal/runtime/budget.go` lines 122-134 (`IsAtHardLimit` — never calls `PersistProgress`).
- **Observed pattern**: `PersistProgress` already writes `progress.md` and returns a paste-ready resume message (test-verified at budget_test.go:199-239, `TestPersistProgressAt75Pct`). The `RuntimeConfig.ResumeMessageFormat` (config.go:146) carries a resume-message template. But `PersistProgress` is never automatically triggered when `IsAtHardLimit` returns true — the wiring gap between the hard-limit signal and the handoff-generation path is D's domain. The existing `ResumeMessageFormat` is a simplified template; D aligns the auto-generated handoff with the canonical 6-block format in `.claude/rules/moai/workflow/session-handoff.md`.
- **Wiring gap**: `IsAtHardLimit` → (missing wire) → `PersistProgress` / handoff generation. D connects them.

### Aggregate defect claim

The aggregate defect is: **the Token Circuit Breaker detects budget exhaustion (`IsAtHardLimit`) but takes no graceful action, and the file-redirect contract persists evidence to non-persistent `/tmp`**. Two VERIFY-DIET §Out of Scope items (budget hard-stop + persistence layer) are D's mandate. D resolves both in a single integrated SPEC, closing the Token-Economy Epic.

---

## Requirements (GEARS notation)

> **Subject convention**: GEARS (current notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format) generalizes the subject beyond "the system". The requirements below use "the runtime" / "the file-redirect contract" / "the evidence persistence layer" as appropriate generalized subjects. Each requirement is a discrete, testable assertion. No legacy `IF/THEN` modality is used.

### REQ-TBS-001 — Event-driven (When) — graceful-abort on hard-limit

**When** `Tracker.IsAtHardLimit(agentName)` returns true, the runtime SHALL emit a budget-exhaustion signal carrying a paste-ready resume message, WITHOUT returning an error that blocks the next call.

> **Test (AC-TBS-001)**: a new graceful-abort method on `Tracker` is invoked when `IsAtHardLimit` returns true; the method returns a signal + handoff message; `RecordCall` continues to return no error.

### REQ-TBS-002 — Event-driven (When) — handoff format conformance

**When** the runtime generates a paste-ready resume message on budget exhaustion, the message SHALL conform to the 6-block structure defined in `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (cut-line markers + ultrathink opener + applied lessons + preconditions + run command + follow-up).

> **Test (AC-TBS-002)**: the auto-generated resume message contains the canonical cut-line markers (`✂────`) and the 6-block structure tokens (`ultrathink`, `전제 검증` / `Preconditions`, `실행` / `Run`).

### REQ-TBS-003 — Unwanted behavior — /clear prohibition

The runtime SHALL NOT auto-invoke `/clear` or any mechanism that clears the conversation context automatically. Budget exhaustion generates a handoff recommendation; the user pastes the resume after a manual `/clear`.

> **Test (AC-TBS-003)**: grep verification — `internal/runtime/` contains no `os/exec` import for `/clear`, no shell invocation of `clear`, and the existing `TestNoAutoClearInvocation` (budget_test.go:328-341) continues to pass.

### REQ-TBS-004 — Ubiquitous — warning-first policy preservation

The runtime SHALL preserve the BC-V3R3-006 warning-first policy: `RecordCall` SHALL continue to return no error on budget exhaustion. Graceful-abort is a **recommendation + handoff generation**, NOT an execution block. The "will switch to hard-fail in P5" comment in `config.go:4` is NOT fulfilled by D — D fulfills the **graceful** path; true hard-fail (error return blocking the next call) remains out of scope.

> **Test (AC-TBS-004)**: `RecordCall` signature remains `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` (no error return added); calling `RecordCall` on an over-budget agent completes without error.

### REQ-TBS-005 — Ubiquitous — evidence persistence obligation

The file-redirect contract SHALL obligate the cited evidence path to remain reachable at audit time, including after `/tmp` directory clearance. Evidence persisted under `.moai/state/verify/<session>/` (gitignored runtime state, same area as `context-usage.json` + `active-sessions.json`) satisfies this obligation. The exact persist mechanism (direct write to `.moai/state/verify/` vs. `/tmp` write + copy step) is a run-phase design decision (plan.md §I).

> **Test (AC-TBS-005)**: the file-redirect contract section in `agent-common-protocol.md` names `.moai/state/verify/<session>/` as the persistent evidence location; the contract states the reachability obligation survives `/tmp` clearance.

### REQ-TBS-006 — Ubiquitous — vci preservation

The evidence persistence layer SHALL preserve `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) and surface 2 (manager §E self-verification) — every claim row in a Verification Matrix or §E self-verification block MUST remain attributable to a directly-observed command whose verbatim output is reachable at the cited file path. **"Persist evidence" ≠ "drop evidence".** The contract is "verbatim evidence lives on disk with a citable path; context carries exit code + bounded tail" — NOT "drop the evidence".

> **Test (AC-TBS-006)**: the contract section explicitly names `verification-claim-integrity` and the §1.1 surface 1+2 preservation obligation, and explicitly rejects the "drop the evidence" interpretation; the cited path in a Verification Matrix row resolves to a file containing the command + full verbatim output.

### REQ-TBS-007 — Capability gate (Where) — template-first boundary

**Where** a doctrine file lives in both the LIVE tree (`.claude/rules/` / `.claude/output-styles/`) and the template tree (`internal/template/templates/`), the run-phase SHALL apply edits to the template source first and rebuild via `make build`, OR identically in both trees. Go code in `internal/runtime/` and `internal/config/` is dev/runtime tooling and is NOT templated (per CLAUDE.local.md §15/§25 language neutrality + template-internal-isolation doctrine).

> **Test (AC-TBS-007)**: `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` + `internal/template/templates/.claude/output-styles/moai/moai.md` carry the same persistence-obligation edits as their LIVE-tree counterparts; `internal/runtime/` Go files have no template-tree mirror (verified by `ls internal/template/templates/internal/runtime/` → not present).

---

## Constraints

1. **`/clear` auto-invocation prohibition (HARD)** — budget.go:18 + IncrementStallRetry:166. D's graceful-abort MUST NOT auto-invoke `/clear`. It generates a handoff and recommends termination; the user pastes the resume after a manual `/clear`. This HARD constraint is preserved from V3R3 (BC-V3R3-006). REQ-TBS-003 binds.
2. **Warning-first policy preservation (HARD)** — BC-V3R3-006. Graceful-abort is NOT a hard-fail. `RecordCall` continues to return no error; the abort is a *recommendation* + *handoff generation*, not an execution block. The "will switch to hard-fail in P5" comment in config.go:4 is NOT fulfilled by D — D fulfills the *graceful* path; true hard-fail remains out of scope. REQ-TBS-004 binds.
3. **vci §1.1 preservation (HARD)** — verification-claim-integrity.md surface 1+2. The /tmp evidence persistence MUST preserve evidence reachability. "Persist evidence" ≠ "drop evidence". The cited path in the Verification Matrix / §E block MUST resolve to a file containing the command + full verbatim output, whether under /tmp or .moai/state/verify/. REQ-TBS-006 binds.
4. **Template-First Rule** — agent-common-protocol.md + moai.md doctrine edits MUST mirror to `internal/template/templates/`. BUT budget.go + any Go code in `internal/runtime/` + `internal/config/` is dev/runtime tooling → NOT templated (per Epic memory §A + CLAUDE.local.md §15/§25). Only the `.claude/rules/` + `.claude/output-styles/` doctrine edits are templated. REQ-TBS-007 binds.
5. **GEARS notation** — requirements use current GEARS notation (Ubiquitous / When / While / Where + generalized subject). No legacy `IF/THEN` modality.
6. **era: V3R6**, status: draft, 12 canonical frontmatter fields (created/updated/tags — no snake_case aliases).

---

## Scope Decision — Why Integrated (Budget Graceful-Abort + Evidence Persistence)

D implements BOTH (a) budget graceful-abort + handoff generation AND (b) /tmp evidence persistence, in a single SPEC. Both are VERIFY-DIET §Out of Scope items explicitly deferred to D:

- VERIFY-DIET spec.md lines 139-142 (§Out of Scope — D): *"graceful abort + paste-ready handoff when runtime token budget is exhausted"* + *"Any Go code in `internal/runtime/budget.go` adding a hard-fail path — D's EXTEND base."*
- VERIFY-DIET spec.md lines 158-161 (§Out of Scope — Mechanical enforcement hook): *"A future hook that mechanically redirects Bash output to disk and validates the cited path resolves"* + *"Any change to `.claude/hooks/moai/*.sh` — out of scope."*

D owns the **persistence obligation** (evidence survives /tmp clear) AND the **budget graceful-abort**. The "mechanical enforcement hook" (a `.claude/hooks/moai/*.sh` that validates the cited path resolves at tool-call time) is a SEPARATE follow-up — D does NOT author a hook script. D's persistence layer is a runtime/Go + doctrine concern, not a hook. See §Out of Scope — Mechanical enforcement hook script for the boundary clarification.

Single SPEC closes Epic D (4/4 → 100%).

---

## Out of Scope

> Per `.claude/rules/moai/development/spec-frontmatter-schema.md` `OutOfScopeRule`, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — True hard-fail (error return blocking next call)

- Adding an error return to `RecordCall` that blocks the next call when `IsAtHardLimit` is true. D fulfills the **graceful** path (recommendation + handoff), NOT true hard-fail. The "will switch to hard-fail in P5" comment in `config.go:4` remains unfulfilled after D — a future SPEC may own true hard-fail if ever needed. REQ-TBS-004 binds this boundary.
- Any mechanism that prevents the next `RecordCall` or agent invocation from proceeding. Graceful-abort recommends termination; it does not enforce it.

### Out of Scope — Mechanical enforcement hook script

- A future hook script under `.claude/hooks/moai/*.sh` that mechanically redirects Bash output to disk and validates the cited path resolves at tool-call time (file-redirect contract's enforcement layer). D authors the **doctrine persistence obligation** (the contract states evidence MUST survive /tmp clearance) and MAY provide a Go persistence helper, but D does NOT author a `.claude/hooks/moai/*.sh` hook script. The mechanical enforcement hook is a separate follow-up SPEC.
- Any change to `.claude/hooks/moai/*.sh` — out of scope. D's persistence layer is a runtime/Go + doctrine concern.

### Out of Scope — Token accounting measurement

- Per-SPEC token spend measurement (`SPEC-TOKEN-ACCOUNTING-001`'s domain — A). D generates a handoff on budget exhaustion; it does not measure the cumulative spend. The `progress.md §I Token Accounting` section remains owned by the token-accounting mechanism at sync-close.

### Out of Scope — Model/effort routing

- Tier×Phase declarative model/effort routing matrix (`SPEC-TOKEN-ROUTING-001`'s domain — B). The graceful-abort trigger is orthogonal to which model runs the work.

### Out of Scope — Verification output diet (file-redirect contract representation)

- The file-redirect contract's **representation** (redirect + bounded tail vs inline verbatim) was authored in VERIFY-DIET-C. D extends the contract with a **persistence obligation** (evidence survives /tmp clearance); D does NOT re-author the representation contract itself.

### Out of Scope — session-handoff.md body modification

- D REUSES the 6-block paste-ready resume format from `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format. D does NOT modify session-handoff.md body content. The auto-generated handoff conforms to the existing format; the format itself is not changed.

---

## Cross-References

- **Preserved invariant (budget)**: `internal/runtime/budget.go` line 18 — `/clear is NEVER invoked automatically (HARD constraint)`. REQ-TBS-003 binds. Also `IncrementStallRetry` line 166 — *"/clear is NOT auto-triggered."*
- **Preserved invariant (warning-first)**: `internal/runtime/budget.go` lines 15-16 — BC-V3R3-006 warning-first policy. REQ-TBS-004 binds.
- **Preserved invariant (vci)**: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1+2. REQ-TBS-006 binds.
- **EXTEND base (Go)**: `internal/runtime/budget.go` (197 lines) — `IsAtHardLimit` (lines 122-134), `RecordCall` (lines 66-91), `Tracker` struct (lines 23-35). `internal/runtime/persist.go` line 18 — `PersistProgress`. `internal/runtime/config.go` (168 lines) — `RuntimeConfig` (lines 43-64), `DefaultRuntimeConfig` (lines 136-148).
- **EXTEND base (doctrine)**: `.claude/rules/moai/core/agent-common-protocol.md` § File-redirect contract (lines 317-323). `.claude/output-styles/moai/moai.md` §8 Verification Matrix (line 396+, evidence path at line 413) + Completion Report (line 605+, evidence path at line 613).
- **Reuse target (handoff format)**: `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (6-block structure + cut-line markers).
- **Sibling device (cross-ref only, NOT target)**: `internal/config/token_budget_guard.go` (always-loaded 75K tripwire, SPEC-TOKEN-EFFICIENCY-001). `internal/runtime/cache_control.go` (prompt cache_control injection).
- **Epic context**: Token-Economy Epic A→B→C→D. A = SPEC-TOKEN-ACCOUNTING-001 (closed). B = SPEC-TOKEN-ROUTING-001 (closed). C = SPEC-TOKEN-VERIFY-DIET-001 (closed, 3bb890f75).
- **Mandate source**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/spec.md` §Out of Scope (lines 139-161).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Token-Economy Epic D of 4 (final). Budget graceful-abort + /tmp evidence persistence. Tier M. |
