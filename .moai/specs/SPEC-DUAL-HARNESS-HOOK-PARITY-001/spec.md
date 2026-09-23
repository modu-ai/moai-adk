---
id: SPEC-DUAL-HARNESS-HOOK-PARITY-001
title: "Dual harness M2 — hook chain, decision preservation, goal continuation, and obligation coverage parity between Claude Code and Codex"
version: "0.1.0"
status: draft
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/codexadapter, internal/codexwiring, internal/cli, internal/goal, internal/verify"
lifecycle: spec-anchored
tags: "dual-harness, codex, hook, stop-chain, decision-preservation, goal, receipt, obligation-coverage, card-t1099"
tier: L
card: t1099
related_specs: [SPEC-CODEX-HOOK-ADAPTER-001, SPEC-CODEX-EVENT-COVERAGE-001, SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-WIRING-001]
---

# SPEC-DUAL-HARNESS-HOOK-PARITY-001 — Dual harness M2: hook, approval, goal, and obligation parity

## HISTORY

- 2026-09-23 v0.1.0 — Initial draft (manager-spec, card t1099, Class C). Derived from the dual-harness full design (`reports/moai-dual-harness-full-design-20260922.md` §07–§09, §17 F2/F3, §18 M2, §19), the implementation status report (`reports/moai-dual-harness-implementation-status-20260923.md`, verdict FAIL), and the handoff (`reports/moai-dual-harness-handoff-20260923.md`). Code anchors re-measured at HEAD `530d8cc06` — see research.md §R1.

## §A Background and Motivation

The dual-harness design defines parity as *effect* parity: the same task under Claude Code and
Codex must reach the same required artifacts, the same approval scope, the same verification
conditions, and the same completion verdict (design §03). Two audit findings block that for the
hook layer, and this SPEC is the M2 milestone that closes them:

- **F2 — the Codex Stop registration is one generic handler, not the Claude chain.** Claude's
  Stop event runs eight handlers (re-measured, research.md §R1.2): turn state, sync-phase quality
  gate, goal evaluation, two security observers, two review gates, and a learning observer. The
  Codex `hooks.json` renderer registers exactly one handler per adapted event, so Codex Stop runs
  only `moai hook stop --harness codex`, whose handler returns `{}` (allow) on every path except a
  factory continuation (research.md §R1.3). The goal loop, the quality gate, and both review gates
  never run under Codex.
- **F3 — four of twelve Codex events are not adapted.** `PreCompact`, `PostCompact`,
  `PermissionRequest`, and `Interrupt` are recognized and refused (research.md §R1.1). Checkpoint,
  approval-decision, and cancellation semantics therefore have no Codex path.

A further hazard was found during re-measurement and is brought into scope because AC-HOOK-02
names it directly: the Codex output adapter drops a PreToolUse `ask` decision to the no-opinion
`{}` (research.md §R1.4). Whether Codex then prompts or proceeds depends on its approval policy —
a `needs_input` decision can silently become an allow. That is a hypothesis until measured; the
requirements below make the measurement mandatory and forbid the degradation.

Finally, the design requires that every required obligation be wired to both harnesses'
application path and check (AC-POL-01). No obligation registry exists in the tree today
(research.md §R1.6), so the coverage claim cannot currently be made or refuted.

## §B Scope

In scope — the four acceptance families of design §19 assigned to M2:

- **AC-HOOK-01** — effect equivalence of the Claude mandatory Stop chain on Codex, proven by
  event-input goldens and by real firing in both hosts.
- **AC-HOOK-02** — `deny` / `needs_input` never degrade to `allow`; `PreCompact`, `PostCompact`,
  `PermissionRequest`, and `Interrupt` fire and preserve decisions; inexpressible, timeout, and
  corrupt-output cases are injected.
- **AC-GOAL-01** — Codex goal continuation while unmet, termination when met, cancellation
  precedence, and budget termination; a repeated-block cap or a host override is never recorded
  as success.
- **AC-POL-01** — every required obligation in the registry is wired to both harnesses'
  application path and to a check; a variant with one obligation removed fails.

## §C Requirements (GEARS)

### C.1 Stop chain parity (AC-HOOK-01)

- **REQ-HPR-001** (Ubiquitous): The Stop-chain inventory shall enumerate every handler registered on the Claude `Stop` event in the distributed settings template and classify each as `required-gate`, `goal`, or `advisory`, with a declared Codex effect path or an explicit `UNSUPPORTED` record carrying evidence.
- **REQ-HPR-002** (Ubiquitous): For every `required-gate` and `goal` member, the Codex harness shall produce the same normalized decision (`allow` / `deny` / `needs_input` / `retryable_error` / `fatal_error`) and the same continuation reason class as the Claude harness on the same event-input golden and the same project configuration.
- **REQ-HPR-003** (Capability gate): **Where** the project is deployed with the `gpt` profile (no `.claude/` tree), every `required-gate` and `goal` Stop-chain member shall remain executable on Codex without depending on a file under `.claude/hooks/`.
- **REQ-HPR-004** (Event-driven): **When** an `advisory` Stop-chain member fails on either harness, the chain shall record the failure and continue, and shall never record that member as passed.
- **REQ-HPR-005** (Ubiquitous): The Stop chain shall honour each member's existing configuration switch (for example the sync gate's blocking opt-out) identically on both harnesses, so that the same configuration yields the same effect.

### C.2 Decision preservation and event adaptation (AC-HOOK-02)

- **REQ-HPR-006** (Ubiquitous): The hook layer shall normalize every decision-bearing handler result to one of `allow`, `deny`, `needs_input`, `retryable_error`, `fatal_error` before harness-specific translation, and the translation table shall be declared per event per harness.
- **REQ-HPR-007** (Unwanted): The Codex translation shall not map a `deny` or a `needs_input` decision to `allow`, nor to any output that the Codex host resolves as allow under any approval policy MoAI supports.
- **REQ-HPR-008** (Event-detected): **When** Codex cannot express a `needs_input` decision on an event, the adapter shall produce a fail-closed outcome (deny with a reason naming the required user input, or a pause routed to MoAI's explicit user-input path) and shall record the case as `UNSUPPORTED-native` with the measured host behaviour.
- **REQ-HPR-009** (Event-detected): **When** a decision-bearing hook times out, crashes, exits with a non-contract code, or emits unparseable output on a decision-bearing event (`PreToolUse`, `PermissionRequest`, `Stop` required-gate), the observed host outcome shall be measured on both harnesses, and where a host resolves the failure as allow, the case shall be recorded as `UNSUPPORTED` with a declared mitigation rather than as `PASS`.
- **REQ-HPR-010** (Event-driven): **When** Codex fires `PreCompact` or `PostCompact`, MoAI shall run the same checkpoint-save and checkpoint-restore behaviour it runs for Claude, and the restored content shall match the saved content.
- **REQ-HPR-011** (Event-driven): **When** Codex fires `PermissionRequest`, MoAI shall apply the same permission decision logic as for Claude, and a `deny` produced by that logic shall reach the host as a deny.
- **REQ-HPR-012** (Event-driven): **When** Codex fires `Interrupt`, MoAI shall record a user cancellation bound to the session and run, without registering a Claude-side dispatcher counterpart and without modifying the Claude-side event vocabulary in `internal/hook`.
- **REQ-HPR-013** (Event-detected): **When** a live trigger for an event cannot be achieved in a test run, the verdict for that event shall be `NOT_RUN` with the attempted trigger and observed output recorded, and shall never be counted as `PASS` or as "does not fire".

### C.3 Goal continuation (AC-GOAL-01)

- **REQ-HPR-014** (State-driven): **While** a goal is armed for the session, the Codex Stop path shall consult the existing goal evaluator (not a new goal engine), request turn continuation while any condition is unmet, and allow the stop with the goal state reading `satisfied` once every condition is satisfied.
- **REQ-HPR-015** (Event-driven): **When** a user cancellation (Codex `Interrupt`, Claude user interrupt, or `moai goal clear`) is recorded for the run, the cancellation shall take precedence over an unmet goal, the loop shall not resume automatically, and the goal state shall record a cancellation distinct from `satisfied`.
- **REQ-HPR-016** (Event-driven): **When** a goal bound (turn ceiling, wall-clock, stagnation) is reached, the loop shall terminate with a persisted verdict whose status is not `satisfied`.
- **REQ-HPR-017** (Unwanted): The goal state shall not be recorded as `satisfied` because the host stopped the session despite a block decision (a consecutive-block cap, `stop_hook_active`, or any other host override).

### C.4 Long-running checks and receipts

- **REQ-HPR-018** (Unwanted): A hook handler shall not start a check whose budget exceeds the hook timeout registered for that handler on the harness it runs under.
- **REQ-HPR-019** (Event-driven): **When** a required check cannot complete within the hook timeout, it shall run outside the hook and leave a receipt binding HEAD, working-tree digest, configuration digest, command, and tool version; the Stop chain shall accept the check only when every receipt field equals the current value, and shall treat an absent receipt or any differing field as not run, never as passed.

### C.5 Obligation coverage (AC-POL-01)

- **REQ-HPR-020** (Ubiquitous): The obligation registry shall list every required obligation with its identifier, its Claude application path, its Codex application path or an `UNSUPPORTED`/`blocked` marker, and the identifier of the check that verifies it.
- **REQ-HPR-021** (Event-detected): **When** a required obligation lacks an application path on either harness or lacks a check, the coverage check shall fail and name the obligation.
- **REQ-HPR-022** (Unwanted): The overall parity verdict shall not be `PASS` while any required obligation is `blocked`, `UNSUPPORTED`, or `unverified`.

### C.6 Evidence integrity

- **REQ-HPR-023** (Unwanted): No verdict in this SPEC shall count a skipped test, an empty run, a live `NOT_RUN`, or the mere existence of a hook configuration file as `PASS`.
- **REQ-HPR-024** (Ubiquitous): Every verdict shall be attributed to the commit, working-tree digest, Claude Code version, codex-cli version, and OS it was measured on.
- **REQ-HPR-025** (Unwanted): Live Codex runs shall not read from or write to the operator's real `~/.codex` state; they shall run under a temporary `CODEX_HOME`, and live Claude runs shall run in a scratch project outside the development checkout.

## §D Overlap with related SPECs

| Related SPEC | Overlap | Relation |
|---|---|---|
| SPEC-CODEX-HOOK-ADAPTER-001 (completed) | Event-name normalization (REQ-1), inert-key output mapping (REQ-2), no-silent-no-op (REQ-3), exit-2 semantics (REQ-4), package placement "nothing under internal/hook is modified" (REQ-7) | This SPEC extends the adapter; it keeps REQ-7 by placing new code in `internal/codexadapter`, `internal/cli`, and existing service packages. REQ-HPR-007 tightens the card-t590 PreToolUse drop branch that the adapter added after that SPEC closed. |
| SPEC-CODEX-EVENT-COVERAGE-001 (completed) | The 12-row event table, the Interrupt row (REQ-CEV-001..006), and the codex-cli 0.153.4 firing campaign (REQ-CEV-007..011) with the trigger-not-achieved vs not-fired distinction | This SPEC adapts the four rows that SPEC held back. REQ-CEV-005 ("M1 shall not register a new dispatcher subcommand") was scoped to that SPEC's M1; REQ-HPR-012 adds the Interrupt path in `internal/cli`/`internal/codexadapter`, keeping REQ-CEV-002 (no constant in `internal/hook`). REQ-HPR-013 inherits its trigger-not-achieved distinction. The campaign must be re-run: the installed codex-cli is 0.155.1, not 0.153.4 (research.md §R1.7). |
| SPEC-CODEX-DUAL-AGENTS-001 (completed) | Neutral agent source → `.codex/agents/*.toml`; per-agent Claude `hooks:` frontmatter documented as having no Codex per-agent equivalent; Codex-side permission enforcement excluded | No requirement overlap. The per-agent hook drop stays a documented drop; role-level permission enforcement belongs to sibling card t1100 (AC-AGENT-01). |
| SPEC-CODEX-WIRING-001 | `RenderHooks` merge model and `--harness codex` runtime flag | This SPEC changes what the table renders (adapted rows, Stop composition); the merge/preservation contract of that SPEC is unchanged. |

## §E Constraints

- Reuse before build: the goal evaluator (`internal/goal`), the verification snapshot key (`internal/verify` `Key` — HEAD + porcelain + diff + untracked digest), and the sync gate's existing outcome record are the starting points for REQ-HPR-014..021; a new engine or store needs a written justification in design.md.
- `internal/hook` (the Claude-side dispatcher vocabulary) is not modified, per SPEC-CODEX-HOOK-ADAPTER-001 REQ-7, unless design.md records an explicit amendment accepted at plan audit.
- Verification is scoped to affected packages locally; the full suite is judged by CI.
- No time estimates; milestones carry priority labels only.

## §F Exclusions (What NOT to Build)

The following are out of scope for this SPEC. Items marked t1100 belong to the sibling card.

### Out of Scope — M4 rollback, unwire, and interrupt recovery (t1100, AC-MIG-01)

- Safe uninstall / rollback / unwire of Codex wiring (`internal/codexwiring/wire.go` install-only path) and transaction-journal recovery.
- Recovery after an interrupted install or update. REQ-HPR-012/016 record a cancellation; resuming or repairing interrupted work is not built here.

### Out of Scope — M3 Codex kanban entry, role permissions, concurrent writers (t1100, AC-WT-01, AC-AGENT-01)

- `moai codex -k` entry and Codex participation in kanban lead/lane roles.
- Runtime permission enforcement for the 12 agent roles (sandbox, shell/MCP tool limits).
- Concurrent-writer protection and un-integrated-worktree deletion guards.

### Out of Scope — messaging and mixed factory (t1100, AC-MSG-01, AC-FACT-01)

- Message dedup, restart, lease/fencing, and late-reply rejection.
- The four mixed-factory combinations (Claude→Claude, Codex→Codex, Claude→Codex, Codex→Claude).

### Out of Scope — M5 command certification and MCP (t1100, AC-WF-01, AC-MCP-01)

- The 17-command × two-CLI certification matrix.
- MCP handshake, tool call, approval, and cancellation parity.

### Out of Scope — adjacent items not assigned to M2

- Policy delivery completeness for standing/scoped rules (M1, AC-TPL-01/02). The obligation registry built here records those obligations; building their Codex delivery path is M1 work.
- Doctor/status readiness display (AC-OBS-01), beyond the per-verdict attribution fields REQ-HPR-024 requires.
- Desktop, Web, and non-macOS runtime certification; verdicts from this SPEC are scoped to the OS they ran on.
- Claude events with no Codex counterpart (`TaskCompleted`, `TeammateIdle`, `StopFailure`, `PostToolUseFailure`, `Notification`) — parity for them is not claimed.
