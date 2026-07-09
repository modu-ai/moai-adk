---
id: SPEC-TOKEN-BUDGET-STOP-001
title: "Token Budget Graceful-Abort + /tmp Evidence Persistence — Implementation Plan"
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
tags: "token-economy, budget, graceful-abort, handoff, file-redirect, plan"
---

# SPEC-TOKEN-BUDGET-STOP-001 — Plan

> plan.md는 파생 실행 계획이다. WHAT/WHY의 SSOT는 spec.md. 본 문서는 HOW의 골격(마일스톤/제약/리스크)이며 함수명·시그니처 등 세부 설계는 run-phase 소관.

## §A Context

### §A.1 Problem summary

The Token Circuit Breaker (`internal/runtime/budget.go`) detects per-agent budget exhaustion via `IsAtHardLimit` (lines 122-134, returns true when usage ≥ 90%) but takes NO graceful action — no handoff is auto-generated, no turn-termination recommendation is emitted. The method exists and is tested, but the wiring to a graceful-abort + handoff-generation path is missing. Meanwhile, the file-redirect contract (authored in VERIFY-DIET-C) writes verbatim evidence to `/tmp/moai-verify/` — non-persistent; `/tmp` clearance makes cited paths unreachable, violating vci §1.1 surface 1+2 at audit time.

D resolves both gaps in a single integrated SPEC:
1. **Budget graceful-abort** — wire `IsAtHardLimit` to a new graceful-abort method that emits a WARN log + auto-generates a paste-ready resume message (conforming to session-handoff.md 6-block format) + recommends turn termination. MUST NOT hard-fail (no error return blocking next call) and MUST NOT auto-invoke `/clear`.
2. **/tmp evidence persistence** — extend the file-redirect contract to obligate evidence persistence under `.moai/state/verify/<session>/` (gitignored runtime state), resolving VERIFY-DIET EC-2 residual risk.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash, vci §2 attribution)

```
wc -l internal/runtime/budget.go            → 197 lines (EXTEND base — Tracker, IsAtHardLimit, RecordCall)
wc -l internal/runtime/config.go            → 168 lines (RuntimeConfig struct, defaults)
wc -l internal/runtime/budget_test.go       → 573 lines (existing test coverage baseline)
wc -l internal/runtime/persist.go           →   3293 bytes (PersistProgress method — reuse target)
wc -l internal/config/token_budget_guard.go → 168 lines (sibling — cross-ref ONLY, not target)
```

Section anchors (grep-verified):
- `grep -n "IsAtHardLimit" internal/runtime/budget.go` → line 122 (M1 EXTEND anchor — method exists, no action wired).
- `grep -n "RecordCall" internal/runtime/budget.go` → line 66 (M1 PRESERVE — warning-first signature).
- `grep -n "PersistProgress" internal/runtime/persist.go` → line 18 (M1 reuse target — already writes progress.md + returns resume message).
- `grep -n "File-redirect contract" .claude/rules/moai/core/agent-common-protocol.md` → line 317 (M2 doctrine anchor).
- `grep -n "evidence: /tmp/moai-verify" .claude/output-styles/moai/moai.md` → lines 413 + 613 (M2 banner evidence path anchor).
- `grep -n "will switch to hard-fail in P5" internal/runtime/config.go` → line 4 (M1 constraint — D fulfills graceful path, NOT true hard-fail).

Template-tree mirror existence (Template-First Rule applicability check):
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` → exists (31550 bytes).
- `internal/template/templates/.claude/output-styles/moai/moai.md` → exists (59048 bytes).
- `internal/template/templates/internal/runtime/` → NOT present (Go code is dev/runtime tooling, NOT templated per CLAUDE.local.md §15/§25).

Both doctrine files are template-managed → Template-First Rule applies to M2 doctrine edits. Go code in M1 is NOT templated.

### §A.3 Approach — Integrated, Two Milestones

M1 (Go code): EXTEND `internal/runtime/budget.go` (or a new sibling file) with a graceful-abort method wired to `IsAtHardLimit`. Reuse `PersistProgress` (persist.go:18) for progress.md write + resume message generation, extended to conform to the session-handoff.md 6-block format. M1 does NOT add an error return to `RecordCall` (warning-first preserved). M1 does NOT auto-invoke `/clear` (HARD constraint preserved).

M2 (doctrine + persistence): Extend the file-redirect contract in `agent-common-protocol.md` to state the persistence obligation (evidence MUST survive `/tmp` clearance; `.moai/state/verify/<session>/` is the persistent location). Update `moai.md` §8 banner evidence path to reference `.moai/state/verify/<session>/`. Mirror both doctrine edits to the template tree. Optionally provide a Go persistence helper (run-phase discretion per §I open decision 1).

### §A.4 Tier Confirmation — M (standard)

**Tier M confirmed.** Rationale grounded in measured evidence:

- **Files affected**: ~5-7 files (budget.go EXTEND, budget_test.go EXTEND, persist.go EXTEND or new sibling, agent-common-protocol.md + template mirror, moai.md + template mirror). Within the 5-15 Tier M range. NOT Tier S (≠ isolated single-file edit). NOT Tier L (≠ 15+ files).
- **LOC**: ~50-100 Go LOC (graceful-abort method + tests) + doctrine prose extensions. Within the 300-1000 Tier M range (predominantly doctrine + small Go extension).
- **Runtime behavior change**: graceful-abort is a NEW behavior path, but it is bounded — it is a recommendation + handoff generation, NOT a hard-fail (no error return, no execution block). The existing `RecordCall` signature is PRESERVED. This is within Tier M scope ("runtime behavior change beyond graceful-abort" would trigger Tier L escalation, but D's runtime change IS graceful-abort, not beyond it).
- **Artifacts**: 3 (spec.md + plan.md + acceptance.md). Tier M standard.
- **plan-auditor threshold**: 0.80 (Tier M).

No escalation to Tier L needed. The integrated scope (budget graceful-abort + evidence persistence) fits Tier M bounds.

---

## §B Known Issues (measured)

- **KI-1 — Warning-first signature preservation**: `RecordCall` (budget.go:66-91) currently returns no error. D MUST NOT add an error return to `RecordCall`. The graceful-abort is a separate method (or a signal returned alongside the existing WARN log), NOT a `RecordCall` error. Failure = BC-V3R3-006 violation + `TestHardLimitWarning` (budget_test.go:126-144) regression.
- **KI-2 — /clear prohibition enforcement**: budget.go:18 + IncrementStallRetry:166 state `/clear` is NEVER auto-invoked. D MUST NOT import `os/exec` for `/clear` or invoke any shell-clearing mechanism. The existing `TestNoAutoClearInvocation` (budget_test.go:328-341) MUST continue to pass. Failure = HARD constraint violation.
- **KI-3 — PersistProgress format alignment**: The existing `ResumeMessageFormat` (config.go:146) is a simplified template (`"ultrathink. {round_label} 이어서 진행. SPEC-{spec_id}부터 ..."`) — it does NOT include the full 6-block structure (cut-line markers, preconditions, follow-up). D's graceful-abort MUST align the auto-generated handoff with session-handoff.md § Canonical Format. The exact alignment (extend `ResumeMessageFormat` vs. a new handoff-generation method) is a run-phase design decision (§I open decision 3).
- **KI-4 — Template-First Rule for doctrine edits**: `agent-common-protocol.md` + `moai.md` live in both LIVE and template trees (mirror existence verified §A.2). M2 MUST apply edits to template source first (`internal/template/templates/...`) and rebuild via `make build`, OR identically in both trees. Go code in `internal/runtime/` has NO template mirror (verified: `internal/template/templates/internal/runtime/` not present) → Go edits are LIVE-tree only. CLAUDE.local.md §2 [HARD] Template-First Rule.
- **KI-5 — moai.md §8 banner HARD rules**: moai.md lines 423-429 carry [HARD] rules governing the Verification Matrix banner (`✓`/`✗` symbols, `(cause: ...)` annotation, max 2 items per line, `📊 N/M PASS` line, evidence-path translation). M2's evidence-path update (`/tmp/moai-verify/` → `.moai/state/verify/`) MUST NOT violate these rules — it updates the path value, not the banner structure.
- **KI-6 — vci §1.1 double-burn preservation**: The file-redirect contract's purpose (from VERIFY-DIET-C) is to reduce double-burn (Bash inline + banner re-quote) while preserving evidence reachability. D's persistence extension MUST NOT regress the double-burn reduction — the cited path still replaces inline verbatim content; it just persists longer.

---

## §C Pre-flight — Devices Verified to Exist

All devices cited in spec.md §Cross-References verified to exist on disk (measured 2026-07-09):

- [x] `internal/runtime/budget.go` — 197 lines, 5998 bytes. `IsAtHardLimit` at line 122. `RecordCall` at line 66. `Tracker` struct at line 23. `@MX:ANCHOR` at line 20. `@MX:SPEC: SPEC-V3R3-ARCH-007` at line 22.
- [x] `internal/runtime/config.go` — 168 lines, 6732 bytes. `RuntimeConfig` struct at line 43. `DefaultRuntimeConfig` at line 136. "will switch to hard-fail in P5" at line 4.
- [x] `internal/runtime/budget_test.go` — 573 lines, 18320 bytes. `TestHardLimitAt90Pct` at line 106. `TestNoAutoClearInvocation` at line 328. `TestPerAgentBudgetOverWarning` at line 343.
- [x] `internal/runtime/persist.go` — 3293 bytes. `PersistProgress` at line 18.
- [x] `internal/config/token_budget_guard.go` — 168 lines (sibling, cross-ref only).
- [x] `.claude/rules/moai/core/agent-common-protocol.md` — § File-redirect contract at line 317. Canonical 7-item batch at lines 290-311.
- [x] `.claude/output-styles/moai/moai.md` — § Verification Matrix at line 396. Evidence path at line 413. § Completion Report at line 605. Evidence path at line 613.
- [x] `.claude/rules/moai/workflow/session-handoff.md` — § Canonical Format (6-block structure + cut-line markers). Reuse target, NOT modification target.
- [x] `.moai/state/context-usage.json` — exists, shape: `{schema_version, session_id, writer_pid, captured_at, context_window_size, tokens_used, raw_pct, stage, band}`. Orthogonal session-level budget signal (§I open decision 2).
- [x] `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/spec.md` — §Out of Scope lines 139-161 (D's mandate source).
- [x] Template mirrors: `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (31550 bytes) + `internal/template/templates/.claude/output-styles/moai/moai.md` (59048 bytes) — both exist.
- [x] `internal/template/templates/internal/runtime/` — NOT present (Go code not templated — correct per §15/§25).

---

## §D Constraints

1. **`/clear` auto-invocation prohibition (HARD)** — budget.go:18. D's graceful-abort MUST NOT auto-invoke `/clear`. REQ-TBS-003. Test: AC-TBS-003 (grep + existing `TestNoAutoClearInvocation`).
2. **Warning-first policy preservation (HARD)** — BC-V3R3-006. `RecordCall` MUST continue to return no error. D fulfills the graceful path, NOT true hard-fail. REQ-TBS-004. Test: AC-TBS-004.
3. **vci §1.1 preservation (HARD)** — evidence persistence MUST preserve reachability. "Persist evidence" ≠ "drop evidence". REQ-TBS-006. Test: AC-TBS-006.
4. **Template-First Rule** — doctrine edits (agent-common-protocol.md + moai.md) MUST mirror to `internal/template/templates/`. Go code NOT templated. REQ-TBS-007. Test: AC-TBS-007.
5. **GEARS notation** — requirements use current GEARS notation. No legacy `IF/THEN`.
6. **era: V3R6**, status: draft, 12 canonical frontmatter fields.
7. **No mechanical hook script** — D does NOT author a `.claude/hooks/moai/*.sh` hook script. The persistence layer is runtime/Go + doctrine, not a hook. See spec.md §Out of Scope — Mechanical enforcement hook script.
8. **No session-handoff.md body modification** — D REUSES the 6-block format; it does NOT modify session-handoff.md. See spec.md §Out of Scope — session-handoff.md body modification.
9. **PRESERVE list (scope discipline)**:
   - `internal/config/token_budget_guard.go` — sibling device, NOT target.
   - `internal/runtime/cache_control.go` — cache_control injection, NOT target.
   - `.claude/rules/moai/workflow/session-handoff.md` — reuse target, NOT modification target.
   - `.claude/hooks/moai/*.sh` — out of scope (no hook script authored).
   - `internal/template/templates/internal/runtime/` — not present (Go code not templated; do NOT create).

---

## §E Plan-phase Self-Verification

- [x] All GEARS requirements use valid patterns (Ubiquitous / When / While / Where + generalized subject; no IF/THEN) — verified in spec.md §Requirements (7 REQs: REQ-TBS-001 When, REQ-TBS-002 When, REQ-TBS-003 shall-not, REQ-TBS-004 Ubiquitous, REQ-TBS-005 Ubiquitous, REQ-TBS-006 Ubiquitous, REQ-TBS-007 Where).
- [x] All gap claims cite measured file + measured line range (vci §2 attribution) — spec.md §Problem GAP-1 (budget.go:122-134, 66-91), GAP-2 (agent-common-protocol.md:290-311, moai.md:413/613), GAP-3 (persist.go:18, budget.go:122-134).
- [x] §Out of Scope has ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets — 6 H3 sub-headings (True hard-fail, Mechanical enforcement hook script, Token accounting, Model/effort routing, Verification output diet, session-handoff.md body modification).
- [x] 12 canonical frontmatter fields, no snake_case aliases (created/updated/tags, NOT created_at/updated_at/labels) — verified in all 4 artifacts.
- [x] SPEC ID matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` — Pre-Write Self-Check printed with → PASS marker (see agent response body).
- [x] progress.md §E.1 skeleton with literal `§E.1` heading (era.go parser load-bearing) — emitted in progress.md.
- [x] /clear prohibition + warning-first preservation stated as HARD constraints — spec.md §Constraints 1+2, plan.md §D 1+2.
- [x] Template-First boundary stated (Go code NOT templated; doctrine edits templated) — spec.md §Constraints 4, plan.md §D 4 + §A.2.
- [x] Cross-referenced devices verified to exist (§C Pre-flight — 11 devices verified).

---

## §F Milestones

### M1 — Budget Graceful-Abort + Handoff Generation (Go code)

**Goal**: Wire `IsAtHardLimit` to a graceful-abort method that emits a WARN log + auto-generates a paste-ready resume message (conforming to session-handoff.md 6-block format) + recommends turn termination. MUST NOT hard-fail. MUST NOT auto-invoke `/clear`.

**EXTEND targets**:
- `internal/runtime/budget.go` (197 lines → ~220-250 lines): add a graceful-abort method on `Tracker`. The method checks `IsAtHardLimit`, and if true, generates a handoff message (reusing/extending `PersistProgress` from persist.go:18). The method returns a signal + handoff message string, NOT an error. `RecordCall` signature unchanged.
- `internal/runtime/persist.go` (3293 bytes → extended): extend `PersistProgress` or add a sibling method that generates a resume message conforming to session-handoff.md 6-block format (cut-line markers + ultrathink opener + applied lessons + preconditions + run command + follow-up). The existing `ResumeMessageFormat` (config.go:146) may be extended or a new format string added.
- `internal/runtime/budget_test.go` (573 lines → extended): add tests for graceful-abort path, 6-block format conformance, /clear prohibition (grep + existing `TestNoAutoClearInvocation` continues to pass), warning-first preservation (`RecordCall` no error return).
- `internal/runtime/config.go` (168 lines → possibly extended): MAY add new config knobs (e.g., `HandoffAutoGenerate bool`, `EvidencePersistPath string`) per §I open decision 3. Run-phase discretion.

**PRESERVE**:
- `RecordCall` signature: `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` — NO error return added.
- `IsAtHardLimit` signature: `func (t *Tracker) IsAtHardLimit(agentName string) bool` — unchanged.
- `TestNoAutoClearInvocation` (budget_test.go:328-341) — continues to pass.
- `TestHardLimitWarning` (budget_test.go:126-144) — continues to pass.
- budget.go:18 `/clear` prohibition comment — preserved.
- budget.go:15-16 BC-V3R3-006 warning-first comment — preserved.
- `@MX:ANCHOR` (budget.go:20) + `@MX:SPEC` (budget.go:22) — preserved.

**Anti-patterns (M1)**:
- Adding an error return to `RecordCall` (BC-V3R3-006 violation).
- Importing `os/exec` for `/clear` invocation (HARD constraint violation).
- Modifying `session-handoff.md` body (out of scope — D reuses, not modifies).
- Creating `internal/template/templates/internal/runtime/` (Go code not templated).

### M2 — /tmp Evidence Persistence + Doctrine Extension

**Goal**: Extend the file-redirect contract to obligate evidence persistence under `.moai/state/verify/<session>/` (gitignored runtime state). Update moai.md §8 banner evidence path. Mirror doctrine edits to template tree.

**EXTEND targets (doctrine)**:
- `.claude/rules/moai/core/agent-common-protocol.md` § File-redirect contract (lines 317-323): extend the contract to state the persistence obligation. The contract SHALL name `.moai/state/verify/<session>/` as the persistent evidence location. The contract SHALL state that evidence MUST remain reachable at audit time, including after `/tmp` clearance. The exact persist mechanism (direct write vs. /tmp write + copy step) is a run-phase design decision (§I open decision 1).
- `.claude/output-styles/moai/moai.md` §8: update the evidence path reference from `/tmp/moai-verify/<session>/` to `.moai/state/verify/<session>/` (or dual-path: `/tmp/moai-verify/<session>/` → persisted to `.moai/state/verify/<session>/`). Lines 413 + 613.
- Template mirrors (Template-First Rule):
  - `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` — same persistence-obligation edit.
  - `internal/template/templates/.claude/output-styles/moai/moai.md` — same evidence-path edit.

**EXTEND targets (Go, optional — run-phase discretion)**:
- A persistence helper MAY be added to `internal/runtime/` (e.g., a function that copies `/tmp/moai-verify/<session>/` → `.moai/state/verify/<session>/`, or a writer that writes directly to `.moai/state/verify/<session>/`). The exact mechanism is §I open decision 1. If added, it is NOT templated (Go code in `internal/runtime/` is dev/runtime tooling).

**PRESERVE**:
- The file-redirect contract's representation obligation (redirect + bounded tail vs inline verbatim) — authored in VERIFY-DIET-C, NOT re-authored by D.
- The 7 verification keywords (`go test`, `coverprofile`, `grep`, `sentinel`, `cmd/moai`, `bench`, `lint`) in agent-common-protocol.md — grep-verified, not removed.
- The parallel-execution HARD obligation (single-turn multi-Bash) — unchanged.
- moai.md §8 [HARD] banner rules (lines 423-429) — not violated by the evidence-path value update.

**Anti-patterns (M2)**:
- Authoring a `.claude/hooks/moai/*.sh` hook script (out of scope — mechanical enforcement hook is a separate follow-up SPEC).
- Re-authoring the file-redirect contract's representation (redirect + bounded tail) — that was VERIFY-DIET-C's domain.
- Modifying session-handoff.md body (out of scope).
- Creating `internal/template/templates/internal/runtime/` (Go code not templated).

---

## §G Anti-Patterns

- **AP-1 — Hard-fail regression**: Adding an error return to `RecordCall` that blocks the next call. This violates BC-V3R3-006 warning-first policy. D fulfills the graceful path, NOT true hard-fail. See spec.md §Out of Scope — True hard-fail.
- **AP-2 — /clear auto-invocation**: Importing `os/exec` or invoking any shell-clearing mechanism. This violates budget.go:18 HARD constraint. See spec.md §Constraints 1.
- **AP-3 — Evidence drop**: Using the persistence extension as a license to drop evidence ("the path is persistent, so we don't need to cite it"). This violates vci §1.1 surface 1+2. "Persist evidence" ≠ "drop evidence". See spec.md §Constraints 3.
- **AP-4 — Hook script authoring**: Creating a `.claude/hooks/moai/*.sh` hook script. D's persistence layer is runtime/Go + doctrine, not a hook. The mechanical enforcement hook is a separate follow-up SPEC. See spec.md §Out of Scope — Mechanical enforcement hook script.
- **AP-5 — session-handoff.md modification**: Editing `.claude/rules/moai/workflow/session-handoff.md` body content. D REUSES the 6-block format; it does NOT modify the format definition. See spec.md §Out of Scope — session-handoff.md body modification.
- **AP-6 — Template-tree Go code**: Creating `internal/template/templates/internal/runtime/` or any Go-code mirror in the template tree. Go code in `internal/runtime/` + `internal/config/` is dev/runtime tooling, NOT templated. See spec.md §Constraints 4.
- **AP-7 — Scope creep to token accounting**: Measuring cumulative token spend in D. That is SPEC-TOKEN-ACCOUNTING-001's domain (A). D generates a handoff on budget exhaustion; it does not measure spend. See spec.md §Out of Scope — Token accounting measurement.

---

## §H Cross-References

- **spec.md (SSOT)**: `.moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` — WHAT/WHY.
- **Epic context**: Token-Economy Epic A→B→C→D. A = SPEC-TOKEN-ACCOUNTING-001 (closed, f88d0226f). B = SPEC-TOKEN-ROUTING-001 (closed). C = SPEC-TOKEN-VERIFY-DIET-001 (closed, 3bb890f75). D = this SPEC.
- **EXTEND base (Go)**: `internal/runtime/budget.go` (197 lines) + `internal/runtime/persist.go` + `internal/runtime/config.go` (168 lines).
- **EXTEND base (doctrine)**: `.claude/rules/moai/core/agent-common-protocol.md` § File-redirect contract + `.claude/output-styles/moai/moai.md` §8.
- **Reuse target**: `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (6-block structure).
- **Preserved invariants**: `internal/runtime/budget.go` line 18 (/clear prohibition) + lines 15-16 (BC-V3R3-006 warning-first) + `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1+2.
- **Mandate source**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/spec.md` §Out of Scope (lines 139-161).
- **Sibling SPECs (plan structure reference)**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/plan.md` (C — doctrine-primary plan structure; D's M2 mirrors). `.moai/specs/SPEC-TOKEN-ACCOUNTING-001/plan.md` (A — Go-code plan structure; D's M1 mirrors).
- **Status Transition Ownership**: manager-spec emits `status: draft` at plan-phase. `draft → in-progress` owned by manager-develop (M1 first run-phase commit). `in-progress → implemented → completed` owned by manager-docs (single sync commit). See `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix.

---

## §I Open Design Decisions (run-phase discretion)

### §I.1 — Persist mechanism (M2)

**Decision**: How does evidence get from the verification command to `.moai/state/verify/<session>/`?

- **Option (a)**: `/tmp` write + copy step. Commands write to `/tmp/moai-verify/<N>-<slug>.log` (fast, tmpfs), then a persistence helper copies the file to `.moai/state/verify/<session>/` after the batch completes. Trade-off: faster writes (tmpfs), but requires a copy step; if the session crashes before the copy, evidence is lost.
- **Option (b)**: Direct write to `.moai/state/verify/<session>/`. Commands write directly to the persistent location. Trade-off: simpler (no copy step), but may be slower on non-tmpfs filesystems; the `.moai/state/` directory is gitignored so no git pollution.
- **Option (c)**: Dual-path. Commands write to `/tmp` (immediate), and the cited path in the banner is `.moai/state/verify/<session>/` (persistent). A background persistence helper syncs the two. Trade-off: most resilient, but most complex.

**Run-phase decision owner**: manager-develop (M2). The doctrine layer (agent-common-protocol.md) states the persistence OBLIGATION without mandating the mechanism; the Go layer implements the chosen mechanism.

### §I.2 — Handoff trigger source (M1)

**Decision**: What signal triggers the graceful-abort?

- **Option (a)**: Per-agent `Tracker.IsAtHardLimit(agentName)` only. Precise per-agent; catches agent-level exhaustion. But misses session-level exhaustion (when the orchestrator's own context is high but no single agent is over budget).
- **Option (b)**: Session-level `context-usage.json` `stage` field only. Catches session-level exhaustion (the `stage` field is `soft` / `hard` per the 2-stage handoff marker). But less precise per-agent.
- **Option (c)**: Both. The graceful-abort triggers when EITHER `IsAtHardLimit` returns true OR `context-usage.json stage == "hard"`. Most comprehensive, but requires reading two signals.

**Run-phase decision owner**: manager-develop (M1). The spec.md REQ-TBS-001 uses `IsAtHardLimit` as the primary trigger; the session-level signal is an optional enhancement.

### §I.3 — Config knobs (M1)

**Decision**: Does `RuntimeConfig` need new fields?

- **Option (a)**: No new fields. The graceful-abort is always enabled; `HardClearThreshold` (already exists, config.go:47) controls the trigger; `ResumeMessageFormat` (config.go:146) is extended to the 6-block format. Simplest.
- **Option (b)**: Add `HandoffAutoGenerate bool` (default true). Allows disabling the auto-generated handoff if a caller wants manual control. More flexible, but adds a config surface.
- **Option (c)**: Add `HandoffAutoGenerate bool` + `EvidencePersistPath string` (default `.moai/state/verify/{SESSION_ID}/`). Full configurability for both M1 + M2.

**Run-phase decision owner**: manager-develop (M1, in coordination with M2). Prefer the simplest option that satisfies the REQs (Option (a) unless a caller needs the toggle).

### §I.4 — Graceful-abort method signature (M1)

**Decision**: What does the graceful-abort method return?

- **Option (a)**: `func (t *Tracker) CheckGracefulAbort(agentName, specID string) (handoff string, shouldAbort bool)` — returns the handoff message + a boolean recommendation. Caller checks `shouldAbort` and decides whether to terminate.
- **Option (b)**: `func (t *Tracker) CheckGracefulAbort(agentName, specID string) string` — returns the handoff message (empty if not at hard-limit). Caller checks for empty string. Simpler.
- **Option (c)**: A struct: `type GracefulAbortSignal struct { Handoff string; ShouldAbort bool; AgentName string; UsageRatio float64 }` — richest, but may be over-engineered for a Tier M SPEC.

**Run-phase decision owner**: manager-develop (M1). Prefer the simplest signature that satisfies REQ-TBS-001 (Option (a) or (b)).

---

## §J Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: M
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 7
acceptance_criteria: 9
out_of_scope_topics: 6
spec_id_self_check: PASS (SPEC-TOKEN-BUDGET-STOP-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts. Tier M, 2 milestones (M1 Go graceful-abort + M2 doctrine persistence). 4 open design decisions deferred to run-phase. |
