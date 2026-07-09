---
id: SPEC-HANDOFF-ONEPASTE-001
title: "Session Handoff 1-Paste: auto-inject pipeline wiring + goal-first flow"
version: "0.1.2"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow + .moai/config/sections"
lifecycle: spec-anchored
tier: M
tags: "handoff, session, auto-resume, doctrine, config, goal-first"
related_specs: [SPEC-HANDOFF-AUTORESUME-001, SPEC-HANDOFF-GOALFIX-001, SPEC-HANDOFF-MSGMODE-001]
---

# SPEC-HANDOFF-ONEPASTE-001 — Session Handoff 1-Paste: auto-inject pipeline wiring + goal-first flow

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.2 | 2026-07-09 | manager-spec | Pre-run touch-up (auditor-sanctioned): R1 — doctrine sentence names the verbatim-persistence obligation WITHOUT the constitution registry ID (mirror leak-class; citation stays in spec/plan prose) + REQ-OP-011 forbidden classes extended with `CONST-*`; R2 — plan §E AC range 016→017 |
| 0.1.1 | 2026-07-09 | manager-spec | Plan-audit iter-2 remediation (D1-D11): byte-parity mirror-CI constraint (REQ-OP-011 rewrite + §B.2 env fact), flow-section /clear-only boundary + precondition clauses (REQ-OP-003), CONST-V3R2-152 temporal separation (REQ-OP-006), `tier: M` frontmatter, measurement anecdote softened |
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — Tier M plan-phase artifacts (doctrine + config, zero Go) |

## §A Context & Problem

The current post-`/clear` handoff requires up to **3 user interactions**: (1) paste the 6-block resume body, (2) approve Implementation Kickoff, (3) send the `/goal` line as a separate standalone message. The consumption infrastructure that collapses this to **1 user message** already exists and is verified — SPEC-HANDOFF-AUTORESUME-001 (19/19 AC, closed at origin `caf0146c4`) shipped:

- **Writer half**: `moai handoff save` / `moai handoff clear` (`internal/cli/handoff.go`; flags `--stdin --body --spec --phase --goal --ultrathink --ultracode --lang --session --project-dir`) writing `.moai/state/handoff/pending.json`.
- **Consumer half**: the SessionStart handler (`internal/hook/handoff_inject.go` + `handoff_inject_render.go`) which — in the single INJECT+CONSUME cell `source == clear ∧ handoff.mode == auto ∧ live pending record` — claim-renames `pending.json` to a `consumed/` audit-trail copy, then injects the saved body verbatim plus directive-restoration guidance as `additionalContext`. All paths are fail-open, race-safe (claim-then-inject atomic rename), and stale-TTL-guarded.

Two gaps keep emission and consumption **unwired**:

1. `.moai/config/sections/handoff.yaml` defaults to `mode: manual` — the injector is a verified pure no-op.
2. No doctrine obligates the orchestrator to run `moai handoff save` when it emits a paste-ready resume message. The message is persisted only to auto-memory and terminal scrollback; `pending.json` never exists.

This SPEC is **100% doctrine + config**: it wires emission to consumption at the policy layer (session-handoff.md SSOT + render/cross-ref surfaces) and flips the local config to `mode: auto`. **Zero new Go code.**

## §B Environment & Assumptions

### B.1 Officially confirmed platform constraints (pre-researched — cite, do not re-derive)

| # | Constraint | Source |
|---|-----------|--------|
| E1 | Slash commands parse only at the input start of a standalone message; queued messages containing slash commands are delivered as literal text, not executed. The 2-block split therefore cannot be collapsed by consecutive pastes. | GitHub anthropics/claude-code#18399; code.claude.com/docs/en/interactive-mode (no queued-slash guarantee) |
| E2 | SessionStart hook `additionalContext`: injected before the first prompt; 10,000-char cap; cannot start a turn; cannot claim effort/xhigh. | code.claude.com/docs/en/hooks |
| E3 | `/goal` is a user-only TUI command; `claude -p "/goal <condition>"` is officially documented; `/clear` removes an active goal; `--resume`/`--continue` restores a still-active goal. | code.claude.com/docs/en/goal |
| E4 | Anthropic long-running-agent guidance converges on disk-state-first: progress file + git log + task list read at fresh-session start; "compaction isn't sufficient". | anthropic.com/engineering/effective-harnesses-for-long-running-agents; code.claude.com/docs/en/best-practices |

### B.2 Existing verified infrastructure (SPEC-HANDOFF-AUTORESUME-001)

- Config keys `handoff.mode` (`manual`/`auto`) and `handoff.guide` already exist in `.moai/config/sections/handoff.yaml` — **no new config keys are needed**.
- `mode: manual` is a verified pure no-op in the injector (REQ-AUTORESUME-009): the manual branch never touches `pending.json`, even when stale. Reverting the config flip restores runtime baseline byte-identically.
- The claim-then-inject rename order (rename success before injection) prevents duplicate injection across two racing sessions (REQ-AUTORESUME-012/013); stale records are TTL-cleaned in auto mode (REQ-AUTORESUME-019).
- `session-handoff.md` is enrolled in **byte-parity mirror CI**: `internal/template/rule_template_mirror_test.go` `workflowOptMirroredPaths` (entry at line 51; sentinel `RULE_TEMPLATE_MIRROR_DRIFT`) requires the live file and its template mirror to be byte-identical, with both files staged in the same commit (the test failure message itself instructs staging both before commit). Verified at plan-phase: all three doctrine surfaces (session-handoff.md, goal-directive.md, moai.md) are currently byte-identical across trees; only session-handoff.md is CI-enrolled.

### B.3 Assumptions

- **A1** — The 6-block Diet Constraints (session-handoff.md § Diet Constraints) keep the saved body well below the 10,000-char `additionalContext` cap (E2). No cap-handling doctrine change is needed; the run-phase notes the cap in the auto-flow section for visibility.
- **A2** — Effort keywords placed inside a slash-command argument are NOT documented to fire (existing session-handoff.md § Goal-first bootstrap caveat (b)). The goal-first single-message flow inherits this caveat honestly; the doctrine does not claim `/goal ... ultrathink ...` restores xhigh effort.
- **A3** — The injected directive-restoration guidance text is rendered by existing Go (`handoff_inject_render.go`). The doctrine flow section MUST describe what that render actually emits (verify-before-author at run-phase), not an aspirational rendering.

## §C Requirements (GEARS)

### C.1 Emission obligation (SSOT: session-handoff.md)

**REQ-OP-001** [HARD] — **When** the orchestrator emits a paste-ready resume message (any of the 5 triggers in session-handoff.md § When To Generate), the orchestrator **shall** also persist the cut-line-bounded main block verbatim via `moai handoff save --stdin --spec <ID> --phase <phase> [--goal "<condition>"] [--ultrathink] [--ultracode] [--lang <conversation_language>] [--session <uuid>]` (body fed via stdin; `--goal` recorded only when the existing `/goal` emission condition holds — run-phase next SPEC with a machine-verifiable end-state).

**REQ-OP-002** [HARD] — **When** the `moai` CLI is absent from PATH or `moai handoff save` exits non-zero, the orchestrator **shall** emit the paste-ready surface unchanged: a save failure never blocks, delays, or alters handoff emission (fail-open; the manual paste path remains fully functional).

### C.2 mode=auto user-flow documentation

**REQ-OP-003** — The session-handoff.md SSOT **shall** carry a new section documenting the mode=auto flow: user runs `/clear` → the SessionStart handler injects the saved body plus directive-restoration guidance as `additionalContext` (claim-then-inject, consumed/ audit copy) → the user sends **ONE** message. **Where** the next SPEC is run-phase with a machine-verifiable end-state, that one message is the single `/goal <condition>` line (goal-first); otherwise it is a short approval message. The flow section **shall** also state: (a) the **/clear-only injection boundary** — startup/resume/compact session starts are notice-only (never consume; with `guide` default `false` the notice is silent), so a terminal restart or an L3 worktree Block 0 resume (new terminal ⇒ `source=startup`) falls OUTSIDE auto-inject and follows the manual paste path; and (b) the **precondition-verification obligation** — the injected Block 4 preconditions are verified at resumed-turn start, most acutely in the goal-first variant where `/goal` starts the turn immediately.

**REQ-OP-004** — **While** the injected `additionalContext` cannot claim effort/xhigh (E2), the mode=auto flow documentation **shall** keep recommending that the user include the `ultrathink` keyword in their first message, and **shall** state the A2 caveat for the goal-first variant (effort keywords inside a slash-command argument are not documented to fire).

**REQ-OP-005** — The session-handoff.md SSOT **shall** retain the existing two-step `/goal` follow-up block (§ Post-Paste /goal Follow-up Block) as the mode=manual fallback path, and the Paste-Time Activation Matrix **shall** remain accurate for both modes (no row deleted; auto-mode additions must not contradict class (d)).

### C.3 Auto-Memory Integration revision

**REQ-OP-006** — The session-handoff.md § Auto-Memory Integration **shall** direct that, on SPEC close, the consumed verbatim resume block inside the memory topic file (the `## 다음 세션 시작점` section) SHOULD be pruned to a one-line summary — verbatim preservation is owned by the `.moai/state/handoff/consumed/` audit trail. This stops double-storage growth (motivation: the auto-memory MEMORY.md index previously approached the official 25KB/200-line load cap and required an index diet). **Temporal separation (unchanged constitution obligation)**: generation-time verbatim persistence per CONST-V3R2-152 is UNTOUCHED — the resume message is still saved verbatim to memory when emitted; the pruning binds only later, at SPEC close. (The doctrine sentence authored at run-phase names this obligation WITHOUT the registry ID — see REQ-OP-011.) Forward-looking only; no retroactive rewrite of existing memory files is mandated.

### C.4 Render-surface parity

**REQ-OP-007** — The render surface `.claude/output-styles/moai/moai.md` §8 Session Handoff block **shall** carry the emission obligation (REQ-OP-001/002) and a compact pointer to the mode=auto flow at parity with the SSOT, and both mutual drift-mitigation sentinels (SSOT § Cross-references sentinel + moai.md §8 sentinel) **shall** be updated consistently in the same change.

### C.5 goal-directive cross-reference

**REQ-OP-008** — The `.claude/rules/moai/workflow/goal-directive.md` § MoAI Integration Notes **shall** cross-reference the mode=auto injected-context path (goal-first single-message flow) alongside the existing two-step manual mechanism, naming session-handoff.md as the SSOT for the flow.

### C.6 Config flip (LOCAL ONLY)

**REQ-OP-009** — The local config `.moai/config/sections/handoff.yaml` **shall** set `mode: auto`.

**REQ-OP-010** — The template mirror `internal/template/templates/.moai/config/sections/handoff.yaml` **shall** retain `mode: manual`: distributed-user auto-resume stays opt-in (same local-vs-template intent pattern as the dev-settings-intent doctrine, CLAUDE.local.md §22).

### C.7 Template mirror neutrality

**REQ-OP-011** — **While** `session-handoff.md` is enrolled in byte-parity mirror CI (§B.2), the run-phase **shall** author its doctrine text **neutrally ONCE** and keep the live file and its template mirror **byte-identical, staged in the same commit** — the doctrine text itself (in BOTH trees) **shall not** carry internal SPEC IDs, internal work dates, memory-measurement anecdotes, or constitution registry IDs (`CONST-*` — internal-content leak class per `internal_content_leak_test.go`) (per `.moai/docs/template-internal-isolation-doctrine.md` §25.1 content classes; motivation prose belongs in this spec.md, never in the doctrine text). goal-directive.md and moai.md are NOT CI-enrolled but are currently byte-identical across trees — the run-phase **shall** keep them byte-identical under the same neutral-once authoring guidance.

### C.8 Invariants

**REQ-OP-012** [HARD] — The doctrine **shall** state explicitly that neither auto-injection nor a set goal pre-authorizes run-phase entry: the Implementation Kickoff Approval human gate is unchanged in both modes.

**REQ-OP-013** [HARD] — **When** `mode: manual` is restored in `.moai/config/sections/handoff.yaml`, runtime behavior **shall** revert to the current baseline byte-identically (the injector's manual branch is a verified pure no-op per B.2), and the doctrine's manual path **shall** remain complete and self-sufficient without the auto-mode section.

**REQ-OP-014** — The cut-line marker specification, the 6-block structure, and the localization tables **shall** remain unchanged (new sections may cite them; no edit inside those specification blocks).

**REQ-OP-015** [HARD] — The run-phase **shall** introduce zero new Go code, zero Go code modifications, and zero new config keys.

## §D Traceability

| REQ | Surface(s) | AC (acceptance.md) |
|-----|-----------|--------------------|
| REQ-OP-001 | session-handoff.md | AC-OP-001, AC-OP-014 |
| REQ-OP-002 | session-handoff.md | AC-OP-012 |
| REQ-OP-003 | session-handoff.md | AC-OP-007, AC-OP-014 |
| REQ-OP-004 | session-handoff.md | AC-OP-015 |
| REQ-OP-005 | session-handoff.md | AC-OP-008 |
| REQ-OP-006 | session-handoff.md | AC-OP-009 |
| REQ-OP-007 | moai.md §8 | AC-OP-002 |
| REQ-OP-008 | goal-directive.md | AC-OP-010 |
| REQ-OP-009 | handoff.yaml (local) | AC-OP-003 |
| REQ-OP-010 | handoff.yaml (template) | AC-OP-004 |
| REQ-OP-011 | template mirrors ×3 (session-handoff.md byte-parity CI-enrolled) | AC-OP-005, AC-OP-017 |
| REQ-OP-012 | session-handoff.md (SSOT) | AC-OP-007 |
| REQ-OP-013 | session-handoff.md + handoff.yaml | AC-OP-016 (scenario) |
| REQ-OP-014 | session-handoff.md | AC-OP-013 |
| REQ-OP-015 | repo-wide | AC-OP-011 |

## §E Exclusions

The following items are explicitly **out of scope** for this SPEC.

### Out of Scope — 16-language rollout of mode=auto default
- Flipping the template `handoff.yaml` default to `auto` for distributed user projects is NOT in scope. Auto-resume remains a per-project opt-in; only the local dev checkout flips (REQ-OP-009/010).

### Out of Scope — New CLI flags or Go changes
- No new `moai handoff` flags, no modifications to `internal/cli/handoff.go`, `internal/hook/handoff_inject*.go`, or any other Go file. The Go infrastructure is closed and verified under SPEC-HANDOFF-AUTORESUME-001.

### Out of Scope — Queued-message workarounds
- No attempt to collapse the 2-block split via consecutive pastes or queued messages. Officially impossible: queued slash commands are delivered as literal text (E1).

### Out of Scope — handoff.yaml schema changes
- No new keys; `mode` and `guide` already exist. `guide` stays `false` (its behavior is untouched).

### Out of Scope — Retroactive memory pruning
- REQ-OP-006 is forward-looking doctrine only; no batch rewrite of existing memory topic files.

## §F Cross-References

- `.claude/rules/moai/workflow/session-handoff.md` — SSOT being revised (M1)
- `.claude/output-styles/moai/moai.md` §8 — render surface (M1)
- `.claude/rules/moai/workflow/goal-directive.md` § MoAI Integration Notes — cross-ref surface (M1)
- `.moai/config/sections/handoff.yaml` — local config flip (M2)
- `internal/cli/handoff.go`, `internal/hook/handoff_inject.go` — existing verified infrastructure (read-only; environment)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — template mirror neutrality contract
- SPEC-HANDOFF-AUTORESUME-001 — consumption-side predecessor (closed, origin caf0146c4)
- SPEC-HANDOFF-GOALFIX-001 — two-step `/goal` follow-up mechanism (retained as manual fallback)
