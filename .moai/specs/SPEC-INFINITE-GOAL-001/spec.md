---
id: SPEC-INFINITE-GOAL-001
title: "Remove the /moai goal 1M cap — infinite auto-compact-driven autonomy with real bounds"
version: 0.1.0
status: draft
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: P0
phase: "v3.x target"
module: workflow-goal
lifecycle: spec-anchored
tags: "goal-engine, auto-compact, max-turns, block-cap, statusline, session-start, handoff-rearm, ac-converge, autonomy-epic"
tier: M
related_specs: [SPEC-AUDIT-SNAPSHOT-001, SPEC-STOPCHAIN-TRIM-001]
---

# SPEC-INFINITE-GOAL-001 — Remove the goal 1M cap (infinite auto-compact autonomy)

## HISTORY

- 2026-08-03 — Initial draft. Codifies §3.2 (v2) of the autonomy-workflow redesign report (`moai-autonomy-workflow-redesign-20260803.html`) + §1.2 (run-phase goal bottleneck) + the C2 finding that the 1M number is NOT a hard cap. Third and final P0 of the autonomy-workflow-epic; proceeds after `SPEC-STOPCHAIN-TRIM-001` completes. The HTML-dashboard aspect of §3.2 is split off to a separate P1 (`SPEC-GOAL-HTML-FLOW`); this SPEC focuses solely on the loop-continuation mechanics that make an armed goal run effectively unbounded.

## §A. User Story

**As a** MoAI maintainer running long unattended Tier-L implementations under an armed `/moai goal`,
**I want** the goal loop to keep running across auto-compact boundaries (the runtime's own context-reclamation mechanism) instead of being force-stopped by three vestigial caps that the 1M number disguises,
**so that** "arm a goal and walk away" actually works — the model keeps iterating until the goal condition holds or a REAL bound (wall-clock, cost, or genuine stagnation) fires.

**Outcome hypotheses (from §3.2 + §1.2 + C2 finding):**

- The 1M number (`internal/config/defaults.go:90` `Default1MContextTokens`) is NOT a hard cap — it is the auto-compact window. Auto-compact itself already preserves the goal (session-id is invariant across compaction, the goal state file survives, the evaluator re-runs each post-compact turn). The 3 caps that ACTUALLY break the loop are: (1) `Ceiling.MaxTurns=30` hardcoded in `NewGoal`, (2) the runtime `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=8` (silent, `min(30,8)=8`), (3) the `/clear` doctrine (50%/90% thresholds force `/clear`).
- After REQ-1 + REQ-2 + REQ-3 + REQ-5 + REQ-6: an armed goal at `--max-turns 0` continues past the 30-turn ceiling, past the runtime block cap (raised), past the auto-compact boundary (goal re-injected), and past a `/clear` (goal re-armed) — bounded only by wall-clock / cost (REQ-4) and the strengthened stagnation guard.

## §B. Scope

**In scope — the 3 real caps + real bounds + compact/clear continuity, all from §3.2 v2:**

- **REQ-1** — `--max-turns N` arm flag (0 = infinite) on `internal/cli/goal.go` arm verb, propagated to `NewGoal`. The `evaluate.go:142` `> 0` guard already disables the ceiling check when `MaxTurns == 0`; 0 IS the infinite entry point. No change to the guard — only to the value propagated.
- **REQ-2** — `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` handling: doctrine guide for raising it (user/runtime env), and an OPTIONAL MoAI-side inject path when a goal is armed.
- **REQ-3** — `internal/statusline/renderer.go:509-578` `computeHandoffStage`: when a goal is armed (`.moai/state/goal/<session>.json` exists + `Status=armed`), suppress the soft/hard `/clear` directive markers (informational-only); let auto-compact handle the pressure. The hard stage is already frequently preempted by auto-compact.
- **REQ-4** — REAL termination bounds: goal-level `max-duration` (wall-clock) and/or `cost cap` (max invocations / token-spend), because `MaxTurns=0` alone permits unbounded cost growth.
- **REQ-5** — `SessionStart(matcher: compact)` hook under `.claude/hooks/moai/`: on auto-compact, read `.moai/state/goal/<id>.json` + the active SPEC's `progress.md` tail and emit (stdout) the goal condition, SPEC-id, last-verified mechanical state, and the single next action — the official "Re-inject context after compaction" mechanism.
- **REQ-6** — `/clear` auto-rearm (Option A, low-risk): `moai handoff save` reads any live armed goal and embeds it in the pending record; `internal/hook/handoff_inject.go:58-108` writes a new goal file under the new session-id when `/clear` fires (`source=clear ∧ mode=auto`). Session-id keying (REQ-GLE-004) preserved. Option B (SPEC-id keying) is REJECTED — multi-session race risk.
- **REQ-7** — C2 finding correction: `ac_converge`'s "Max 20 turns" in `.claude/skills/moai/workflows/run.md:153` is NOT parsed by `parseCondition` (`internal/cli/goal.go:27` only matches `exits <N>`), so the goal actually runs at `MaxTurns=30`. Either add a `stop after N turns` regex to `parseCondition` (so the literal reflects into `Ceiling.MaxTurns` mechanically) or correct the documentation to state the actual 30-turn execution.

**Out of scope — sibling epic items:** the `moai goal render` HTML dashboard, the evaluate.go → HTML verdict rendering, the plan-phase HTML report, and the resume auto-rearm UI — these belong to `SPEC-GOAL-HTML-FLOW` (P1). The autonomy-tier mode token and Stop-chain trim belong to `SPEC-STOPCHAIN-TRIM-001` (this SPEC's REQ-2 / REQ-3 interact with tier-aware hooks but do not redefine them).

### Out of Scope — HTML dashboard surface

- The per-turn HTML dashboard rendering of `evaluate.go` Verdict (§3.2 ④) — sibling P1 `SPEC-GOAL-HTML-FLOW`.
- The `moai goal render` command surface and the plan-phase HTML report — sibling P1.
- The resume auto-rearm UI (the user-facing surface of `/clear` rearm) — sibling P1. REQ-6 here is the mechanical rearm path only, not its UI.

### Out of Scope — Autonomy tier + Stop-chain trim

- The `MOAI_AUTONOMY_TIER` token, the Stop-chain shell trim, and the per-edit hook consolidation — owned by `SPEC-STOPCHAIN-TRIM-001`. This SPEC's REQ-2 (block-cap raise) and REQ-3 (statusline `/clear` suppression) interact with tier-aware hooks but defer the tier-branching to the sibling SPEC.

### Out of Scope — auto-compact internals

- The runtime's auto-compact mechanism itself (Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact) — this SPEC CONSUMES it; it does not modify it. The thrashing guard (single large output → immediate re-inflation → halt) is a runtime invariant this SPEC respects, not changes.

## §C. Requirements (GEARS)

### REQ-INFINITE-GOAL-001 — `--max-turns` arm flag (0 = infinite)

**Where** the `/moai goal` arm verb is implemented in `internal/cli/goal.go`, the arm command SHALL accept a `--max-turns N` flag whose value is a non-negative integer; the value SHALL be propagated to `NewGoal` as `Ceiling.MaxTurns`.

**When** `--max-turns` is `0`, the goal SHALL be armed with `Ceiling.MaxTurns = 0`, and the evaluator's existing `if g.Ceiling.MaxTurns > 0 && g.TurnsUsed >= ...` guard at `internal/goal/evaluate.go:142` SHALL disable the ceiling check — i.e. `0` IS the infinite entry point. No change to the guard is required; only the propagated value changes.

**When** `--max-turns` is omitted on the arm verb, the default SHALL remain the current `30` (full backward compatibility, no regression).

### REQ-INFINITE-GOAL-002 — `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` handling

**Where** the runtime cap `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (default 8) is the silent terminator that pre-empts the goal loop before `MaxTurns` fires (effective bound `min(MaxTurns, 8)` today), the documentation SHALL surface this cap explicitly (in `goal-directive.md` and `workflows/goal.md`) so users understand that an armed goal at `--max-turns 0` requires raising this runtime env (e.g. to 200) for the infinite loop to actually persist.

**Where** MoAI controls session startup (e.g. `moai cc` / `moai cg` launchers), the launcher MAY inject `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=<raised>` into the session environment when a goal is armed; this inject path is OPTIONAL because the env is canonically user/runtime-owned. The doctrine guide is the load-bearing deliverable; the inject is the convenience.

### REQ-INFINITE-GOAL-003 — Suppress statusline `/clear` directive when goal armed

**Where** `internal/statusline/renderer.go:509-578` `computeHandoffStage` computes the soft/hard handoff stage and emits the `(⚠️/clear)` / `(🛑/clear!)` markers, the renderer SHALL first check whether a goal is armed for the current session (`.moai/state/goal/<session-id>.json` exists AND its `Status == "armed"`); **While** a goal is armed, the renderer SHALL suppress the `/clear` directive markers (the stage classification may still be computed for informational purposes, but the directive text is omitted), letting auto-compact handle context pressure instead of forcing a `/clear`.

**When** no goal is armed, the renderer SHALL retain the current behavior unchanged (full backward compatibility).

### REQ-INFINITE-GOAL-004 — Real termination bounds (wall-clock + cost)

**When** `Ceiling.MaxTurns == 0` (infinite turns), the goal SHALL be bounded by at least one of: (a) a goal-level `max-duration` (wall-clock seconds since arm time), OR (b) a goal-level `cost cap` (max model invocations OR max cumulative token spend), recorded in `Ceiling` alongside `MaxTurns`. The bound SHALL fire a verdict indistinguishable from the MaxTurns ceiling firing today (5-section Claim/Evidence/Baseline/Gaps/Residual-risk verdict at `evaluate.go`).

**The stagnation guard at `evaluate.go:84, 106-121, 153-161` (DefaultStagnationThreshold=3) SHALL be strengthened** so that "N consecutive turns with NO mechanical-condition change" halts the loop. The mechanical-condition change is defined as a diff in: test count, test pass/fail tally, OR the SHA of a bounded file set the goal's mechanical conditions target. "Same note 3 times" (today's heuristic) is replaced by this mechanical comparison.

### REQ-INFINITE-GOAL-005 — SessionStart(compact) re-inject

**Where** a SessionStart hook with `matcher: compact` is registered under `.claude/hooks/moai/`, the hook SHALL read the live armed goal (`.moai/state/goal/<session-id>.json`) and the active SPEC's `progress.md` tail, and SHALL emit to stdout a compact re-injection carrying: the goal condition text, the SPEC-id, the last-verified mechanical state (most recent failed-condition + observed output tail), and the single next action. This is the official "Re-inject context after compaction" mechanism — the post-compact model re-orients from this output.

**When** no goal is armed at compact time, the hook SHALL be a no-op.

### REQ-INFINITE-GOAL-006 — `/clear` auto-rearm (Option A, session-id keying)

**Where** `moai handoff save` runs (orchestrator-side, before `/clear`), the save command SHALL read any live armed goal for the current session and embed the goal's verbatim condition (plus the arm-time `Ceiling`: `MaxTurns`, `max-duration`, `cost cap`) into the pending handoff record at `.moai/state/handoff/pending.json`.

**Where** `internal/hook/handoff_inject.go:58-108` handles the `/clear` boundary (`source == "clear"` AND `mode == "auto"`), the handler SHALL — when the pending record carries an embedded goal — write a new goal state file under the NEW session-id (preserving session-id keying per REQ-GLE-004), re-arming the goal for the post-`/clear` session with the embedded `Ceiling`.

Option B (SPEC-id keying instead of session-id keying) is REJECTED because it introduces a multi-session race: two sessions on the same SPEC would both consume the same pending record and double-arm. Session-id keying preserves the one-session-one-goal invariant.

### REQ-INFINITE-GOAL-007 — `ac_converge` "Max 20 turns" parseCondition correction

**Where** `.claude/skills/moai/workflows/run.md:153` `ac_converge` block carries a "Max 20 turns" clause, and `internal/cli/goal.go:27` `parseCondition` only matches the `exits <N>` regex (so the clause is currently NOT parsed and the goal actually runs at `MaxTurns=30`), exactly ONE of the two corrections SHALL apply: (a) add a `stop after N turns` regex to `parseCondition` so the `Max 20 turns` clause reflects mechanically into `Ceiling.MaxTurns`, OR (b) correct the documentation to state the actual behavior (the `ac_converge` goal runs at `MaxTurns=30` because the `Max 20 turns` clause is not parseable by the current regex). The choice between (a) and (b) is an OQ-1 resolution.

## §D. Constraints

1. **auto-compact is the pressure-relief, not `/clear`** (from §3.2 + verification-claim-integrity): when a goal is armed, `/clear` doctrine is suppressed (REQ-3); auto-compact handles context reclamation. The `stop-goal` hook decides only block/continue; it does NOT decide auto-compact. "Yielding to auto-compact" is NOT a forced `/clear` halt (the doctrine change here is precisely that suppression).
2. **auto-compact thrashing guard respected**: each goal iteration MUST keep its context footprint small (sub-agent delegation, file-redirect bounded tail, targeted mechanical commands). A single large output that immediately re-inflates the context triggers the runtime's thrashing guard and halts the loop — this is a runtime invariant, not a bug.
3. **model conditions break under auto-compact** (verbatim evidence is summarized away): `ac_converge`'s model-condition shape is therefore migrated to mechanical conditions (test count, file hash, exit code). REQ-INFINITE-GOAL-004's stagnation guard encodes this.
4. **3 caps explicit**: `MaxTurns` (REQ-1), `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (REQ-2), `/clear` doctrine (REQ-3). The 1M number alone is NOT modified by this SPEC — modifying only `Default1MContextTokens` is meaningless and a mis-user hazard (a user would believe the cap was raised while the 3 real caps still terminate the loop).
5. **Backward compatibility**: `--max-turns` unspecified = `30` (today). `/clear` directive suppressed ONLY when a goal is armed. REQ-5 / REQ-6 hooks are no-ops when no goal is armed.
6. **MP-3 lesson**: `tags` is a comma-separated quoted string (not a YAML array). Self-verified via `moai spec lint` before returning.
7. **deny/ask invariance cross-reference**: an armed infinite goal does NOT weaken the deny/ask rules — those remain binding per `SPEC-STOPCHAIN-TRIM-001` REQ-007. The goal evaluator decides only continue/stop, never deny-bypass.

## §E. Assumptions

1. `internal/goal/schema.go:78,105` `Ceiling.MaxTurns` is a plain integer field; assigning `0` is type-valid and the `evaluate.go:142` `> 0` guard already handles `0` as "ceiling disabled". Confirmed by the C2 finding in the design report.
2. `internal/statusline/renderer.go:509-578` `computeHandoffStage` already computes the soft/hard stage and emits markers conditionally; adding a "goal armed?" precondition at the entry is a single-branch addition, not a renderer refactor.
3. The runtime's `SessionStart(matcher: compact)` hook contract already exists (Claude Code 2.x) and emits the compact event to a registered hook; MoAI's hook layer needs only to register a handler and emit stdout text.
4. `internal/hook/handoff_inject.go:58-108` already reads the pending handoff record on `/clear` (mode=auto path) and writes session-start context; embedding a goal record and writing a new goal file is an additive extension of that path, not a new injection surface.
5. `moai handoff save` is reachable from the orchestrator before `/clear` (the orchestrator-side emission-time save obligation in `session-handoff.md`), so the goal-embed step has a natural call site.

## §F. Open Questions (for plan-auditor)

- **OQ-1 (REQ-7)**: Should the `ac_converge` "Max 20 turns" clause be made mechanically parseable (add `stop after N turns` regex to `parseCondition`, option a) OR should the documentation be corrected to state the actual 30-turn execution (option b)? Option (a) is a behavior change (the goal would actually stop at 20 turns instead of 30); option (b) is a documentation-only change. Plan-auditor's read of `parseCondition`'s regex scope + the run.md intent should decide.
- **OQ-2 (REQ-4)**: Is the goal-level `max-duration` (wall-clock) the primary real bound, or is `cost cap` (max invocations / token-spend) primary? The report names both; the plan-auditor's read of which is more readily instrumented in `evaluate.go` (wall-clock needs only `time.Since(armTime)`; cost needs invocation/token accounting that may not yet exist) should decide which is the M-default and which is the opt-in.
- **OQ-3 (REQ-2)**: Does the project already inject `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` via `moai cc` / `moai cg` launchers, or is the launcher env-inject surface new territory? Determines whether REQ-2's inject path is a one-line addition or a new launcher contract.

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2 (v2 — 1M cap removal / infinite auto-compact persistence) + §1.2 (run-phase goal bottleneck) + C2 finding (1M number is auto-compact window, not hard cap).
- MaxTurns source: `internal/goal/schema.go:78,105` (`Ceiling.MaxTurns`, `NewGoal`), `internal/goal/evaluate.go:142` (`> 0` guard).
- arm verb: `internal/cli/goal.go` (arm command + `:27` `parseCondition` regex).
- Block cap: runtime `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (default 8); report §3.2 + §1.2 line 256.
- Statusline: `internal/statusline/renderer.go:509-578` `computeHandoffStage`.
- Stagnation guard: `internal/goal/evaluate.go:84, 106-121, 153-161` (`DefaultStagnationThreshold=3`).
- Handoff inject: `internal/hook/handoff_inject.go:58-108`; pending record at `.moai/state/handoff/pending.json`; save command `moai handoff save`.
- run.md ac_converge: `.claude/skills/moai/workflows/run.md:153-166`.
- Sibling P0s: `SPEC-AUDIT-SNAPSHOT-001`, `SPEC-STOPCHAIN-TRIM-001`. Sibling P1: `SPEC-GOAL-HTML-FLOW` (planned).
- Integrity invariants: `verification-claim-integrity.md` §1.1 (goal-evaluator verdict is a claim surface); `SPEC-STOPCHAIN-TRIM-001` REQ-007 (deny/ask invariance).

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-INFINITE-GOAL-001 (REQ-1): `--max-turns 0` arms a goal whose ceiling check is disabled.
- AC-INFINITE-GOAL-002 (REQ-1): `--max-turns` omitted = `30` (backward compat).
- AC-INFINITE-GOAL-003 (REQ-2): block-cap doctrine guide surfaces the silent cap.
- AC-INFINITE-GOAL-004 (REQ-3): armed goal → statusline `/clear` directive suppressed.
- AC-INFINITE-GOAL-005 (REQ-4): `--max-turns 0` bounded by wall-clock OR cost cap.
- AC-INFINITE-GOAL-006 (REQ-4): stagnation guard fires on N-turn mechanical-condition no-change.
- AC-INFINITE-GOAL-007 (REQ-5): post-compact SessionStart re-injects goal + SPEC + next action.
- AC-INFINITE-GOAL-008 (REQ-6): `/clear` re-arms the embedded goal under the new session-id.
- AC-INFINITE-GOAL-009 (REQ-7): `ac_converge` "Max 20" either parses or doc is corrected.
- AC-INFINITE-GOAL-010 (deny/ask cross-ref): armed infinite goal does NOT weaken deny/ask rules.
