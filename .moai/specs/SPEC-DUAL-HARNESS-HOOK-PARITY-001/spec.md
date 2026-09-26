---
id: SPEC-DUAL-HARNESS-HOOK-PARITY-001
title: "Dual harness M2 — hook chain, decision preservation, goal continuation, and obligation coverage parity between Claude Code and Codex"
version: "0.6.0"
status: in-progress
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/codexadapter, internal/codexwiring, internal/cli, internal/goal, internal/verify"
lifecycle: spec-anchored
tags: "dual-harness, codex, hook, stop-chain, decision-preservation, goal, receipt, obligation-coverage, live-uncertified, card-t1099"
tier: L
card: t1099
related_specs: [SPEC-CODEX-HOOK-ADAPTER-001, SPEC-CODEX-EVENT-COVERAGE-001, SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-WIRING-001]
---

# SPEC-DUAL-HARNESS-HOOK-PARITY-001 — Dual harness M2: hook, approval, goal, and obligation parity

## HISTORY

- 2026-09-23 v0.1.0 — Initial draft (manager-spec, card t1099, Class C). Derived from the dual-harness full design (`reports/moai-dual-harness-full-design-20260922.md` §07–§09, §17 F2/F3, §18 M2, §19), the implementation status report (`reports/moai-dual-harness-implementation-status-20260923.md`, verdict FAIL), and the handoff (`reports/moai-dual-harness-handoff-20260923.md`). Code anchors re-measured at HEAD `530d8cc06` — see research.md §R1.
- 2026-09-23 v0.2.0 — Revision after plan-audit iter-1 (FAIL 0.82, `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-HOOK-PARITY-001-iter1.md`). Codex Stop timeout reframed as a design variable with a per-member budget table and a measurement AC (D2, D11); Claude user-interrupt cancellation source removed and recorded UNSUPPORTED (D3); live-test NOT_RUN hole closed (D4); live commands completed (D6); non-Stop chains excluded with counts (D5); mutation cells, partial legs, AC-OBS-01/AC-POL-01 mapping, template branch coverage, and pinned-test amendments addressed (D7, D8, D10, D12, D13); Q4 moved to run-phase M2d (D9). Clarifications resolved (D1): Q1 whole-catalog AC-POL-01, Q2 fail-closed visible deny, Q6 keep HOOK-ADAPTER REQ-7 (lead via Jev, standing operator delegation); Q3 new `cancelled` status and Q5 no live runs, closing as partial (live-uncertified) (operator, directly). Decision record: plan.md §C.
- 2026-09-23 v0.3.0 — Revision after plan-audit iter-2 (FAIL 0.87, `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-HOOK-PARITY-001-iter2.md`): review gates classed `fail-open-on-missing` on both harnesses to match Claude's source (N1); an in-hook timing leg added and the unmeasured "rule is met" claim withdrawn (N2); goal-status inventory extended to `internal/hook`, with the test's scan scope stated (N3); a unit leg added for the visible needs_input deny (N6); member 1's factory-continuation path exempted from the advisory cut-off (N4); durable live-uncertified marker required at sync close (N5); AC-HPR-004 mutation replaced (N8); adapted row count settled at 12 (N9).
- 2026-09-23 v0.4.0 — Revision after plan-audit iter-3 (FAIL 0.90, `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-HOOK-PARITY-001-iter3.md`), scoped to R1–R4 by operator decision (plan.md §C). R1: member 6 (codex review gate) leaves `fail-open-on-missing` and becomes a `required-gate` on the receipt method — codex installed with a reviewable change and no fresh passing receipt continues the turn, codex binary missing allows with a discard record, matching Claude; `fail-open-on-missing` now covers member 7 only. R2: the sync gate's own self-gate (last commit subject is a sync-phase commit, then language marker and code delta) is evaluated in-hook on Codex before any receipt compare, so a non-sync HEAD allows on both harnesses. R3: the intentional test-amendment list is completed and REQ-CEV-001/003 are recorded as reversed alongside REQ-CEV-004. R4: the durable partial-closure marker moves into plan phase — the `live-uncertified` tag and this HISTORY line are written now by manager-spec, leaving the sync close only manager-docs' own surfaces. Also A2 (factory 200 ms bound cited and counted) and A4 (HISTORY order).
- 2026-09-23 v0.4.0 (closure mode) — This SPEC closes as **partial (live-uncertified)** by operator decision Q5: no live Claude Code or Codex run is executed, every live leg records `NOT_RUN`, and the aggregate verdict is not-PASS for every obligation that needs a live leg. Live certification is deferred to a follow-up card issued by the lead; its id is recorded in progress.md §E.4 at sync close. This line and the `live-uncertified` tag are the plan-phase half of the durable marker (spec.md §E).
- 2026-09-23 v0.5.0 — Inline repair after plan-audit iter-4 (FAIL 0.91, `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-HOOK-PARITY-001-iter4.md`), scoped to S1–S3 by operator decision (inline repair + delta check). S1: `stop_hook_active: true` aligned to Claude — member 6 allows at step 2, named as the only bound on its step-7 continuation, with the field's presence cited from the SPEC-CODEX-HOOK-ADAPTER-001 Stop.json capture and its value on a continued turn recorded as live-unmeasured; the sync gate allows on a fresh `fail` receipt when the flag is set (Claude `sync-phase-quality-gate.sh:484–486`); the "no allow before a verdict" overstatement narrowed; AC-HPR-002 gains both goldens and both mutations. S2: the hook subcommand-count tests (43 → 44) and the `utilitySubcmds` reverse check join the M2e amendment list; the "complete" claim is scoped to the M2e subcommand set. S3: research.md R1.2 member-7 class marked superseded.
- 2026-09-23 v0.6.0 — Operator decision 09-23 (plan.md §C C1): Codex-only consecutive-`unmeasured` cap for the sync gate and Stop member 6 — after N consecutive `unmeasured` continuations for the same gate on the same HEAD + working-tree digest the stop is allowed and the gate is recorded `unverified` (discard record + reason text, never PASS); N proposed 3, finalized in M2d; counter in session-scoped `.moai/state/codex-stop-cap/<session-id>.json`, reset on a fresh receipt or a HEAD/digest change, never written by the receipt producer. Declared parity deviation (design.md §D3.8, §D3.4). REQ-HPR-002 and REQ-HPR-019 reworded; AC-HPR-002 gains four cap goldens and two mutations; AC-HPR-019 gains an `unverified` cap-record injection leg.

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
  event-input goldens and, in a follow-up card, by real firing in both hosts (live legs are built but not run here — §E).
- **AC-HOOK-02** — `deny` / `needs_input` never degrade to `allow`; `PreCompact`, `PostCompact`,
  `PermissionRequest`, and `Interrupt` are adapted, and their decisions are preserved in goldens.
  Their live firing is `NOT_RUN` in this SPEC (§E). Inexpressible, timeout, and corrupt-output
  cases are injected.
- **AC-GOAL-01** — Codex goal continuation while unmet, termination when met, cancellation
  precedence, and budget termination; a repeated-block cap or a host override is never recorded
  as success.
- **AC-POL-01** — every required obligation in the registry is wired to both harnesses'
  application path and to a check; a variant with one obligation removed fails.

Mapping note. Design §17 maps AC-POL-01 to F1/M1, not to M2. This card pulls it forward by
operator scope because the obligation registry is the instrument M2's own verdicts are judged by.
Design §17 also maps F3 to AC-OBS-01 at M2. This SPEC covers the part of AC-OBS-01 that decides a
verdict: per-verdict attribution, and rejecting skips, empty runs, and stale receipts
(REQ-HPR-019, REQ-HPR-022..024). The doctor/status display part of AC-OBS-01 stays out of scope
(§F).

## §C Requirements (GEARS)

### C.1 Stop chain parity (AC-HOOK-01)

- **REQ-HPR-001** (Ubiquitous): The Stop-chain inventory shall enumerate every handler registered on the Claude `Stop` event in the distributed settings template and classify each as `required-gate`, `fail-open-on-missing` (only a member whose Claude source allows the stop when its persisted result is missing), `goal`, or `advisory`, with a declared Codex effect path — in-hook, or receipt with its named receipt producer — or an explicit `UNSUPPORTED` record carrying evidence.
- **REQ-HPR-002** (Ubiquitous): For every `required-gate`, `fail-open-on-missing`, and `goal` member, the Codex harness shall evaluate the member's own applicability predicates (its self-gates, such as the sync gate's sync-phase-commit predicate and the codex review gate's reviewable-change and reviewer-installed predicates) before any receipt comparison, and shall produce the same normalized decision (`allow` / `deny` / `needs_input` / `retryable_error` / `fatal_error`) as the Claude harness on the same event-input golden, the same project configuration, and the same environment (including whether the codex binary is installed), and a continuation reason class that is either identical or paired with it in the declared reason-class mapping (design.md §D3.4). The one declared exception is the Codex-only consecutive-`unmeasured` cap (design.md §D3.8), which has no Claude counterpart; its goldens are Codex-only and are excluded from this equality.
- **REQ-HPR-003** (Capability gate): **Where** the project is deployed with the `gpt` profile (no `.claude/` tree), every `required-gate` and `goal` Stop-chain member shall remain executable on Codex without depending on a file under `.claude/hooks/`.
- **REQ-HPR-004** (Event-driven): **When** an `advisory` Stop-chain member fails on either harness, the chain shall record the failure and continue, and shall never record that member as passed.
- **REQ-HPR-005** (Ubiquitous): The Stop chain shall honour each member's existing configuration switch (for example the sync gate's blocking opt-out) identically on both harnesses, so that the same configuration yields the same effect.

### C.2 Decision preservation and event adaptation (AC-HOOK-02)

- **REQ-HPR-006** (Ubiquitous): The hook layer shall normalize every decision-bearing handler result to one of `allow`, `deny`, `needs_input`, `retryable_error`, `fatal_error` before harness-specific translation, and the translation table shall be declared per event per harness.
- **REQ-HPR-007** (Unwanted): The Codex translation shall not map a `deny` or a `needs_input` decision to `allow`, nor to any output that the Codex host resolves as allow under any approval policy MoAI supports.
- **REQ-HPR-008** (Event-detected): **When** Codex cannot express a `needs_input` decision on an event, the adapter shall produce a fail-closed deny whose reason names the required user input, shall surface the conversion visibly through the adapter's discard record, and shall record the case as `UNSUPPORTED-native` with the host behaviour (measured, or `NOT_RUN` where no live run exists).
- **REQ-HPR-009** (Event-detected): **When** a decision-bearing hook times out, crashes, exits with a non-contract code, or emits unparseable output on a decision-bearing event (`PreToolUse`, `PermissionRequest`, `Stop` required-gate), the observed host outcome shall be measured on both harnesses, and where a host resolves the failure as allow, the case shall be recorded as `UNSUPPORTED` with a declared mitigation rather than as `PASS`.
- **REQ-HPR-010** (Event-driven): **When** Codex fires `PreCompact` or `PostCompact`, MoAI shall run the same checkpoint-save and checkpoint-restore behaviour it runs for Claude, and the restored content shall match the saved content.
- **REQ-HPR-011** (Event-driven): **When** Codex fires `PermissionRequest`, MoAI shall apply the same permission decision logic as for Claude, and a `deny` produced by that logic shall reach the host as a deny.
- **REQ-HPR-012** (Event-driven): **When** Codex fires `Interrupt`, MoAI shall record a user cancellation bound to the session and run, without registering a Claude-side dispatcher counterpart and without modifying the Claude-side event vocabulary in `internal/hook`.
- **REQ-HPR-013** (Event-detected): **When** a live trigger for an event cannot be achieved in a test run, the verdict for that event shall be `NOT_RUN` with the attempted trigger and observed output recorded, and shall never be counted as `PASS` or as "does not fire".

### C.3 Goal continuation (AC-GOAL-01)

- **REQ-HPR-014** (State-driven): **While** a goal is armed for the session, the Codex Stop path shall consult the existing goal evaluator (not a new goal engine), request turn continuation while any condition is unmet, and allow the stop with the goal state reading `satisfied` once every condition is satisfied.
- **REQ-HPR-015** (Event-driven): **When** a user cancellation is recorded for the run by one of its two producers — the Codex `Interrupt` event handled by MoAI, or an explicit `moai goal clear` — the cancellation shall take precedence over an unmet goal and the loop shall not resume automatically; the Interrupt producer shall record the new status `cancelled`, `moai goal clear` shall leave no goal state, neither shall read as `satisfied`, and every non-test reader of goal status shall handle `cancelled` explicitly, with an unrecognised status surfacing a diagnostic rather than a silent block or `satisfied`.
- **REQ-HPR-016** (Event-driven): **When** a goal bound (turn ceiling, wall-clock, stagnation) is reached, the loop shall terminate with a persisted verdict whose status is not `satisfied`.
- **REQ-HPR-017** (Unwanted): The goal state shall not be recorded as `satisfied` because the host stopped the session despite a block decision (a consecutive-block cap, `stop_hook_active`, or any other host override).

### C.4 Long-running checks and receipts

- **REQ-HPR-018** (Unwanted): A hook handler shall not start a check whose budget exceeds the hook timeout registered for that handler on the harness it runs under.
- **REQ-HPR-019** (Event-driven): **When** a required check cannot complete within the hook timeout, it shall run outside the hook and leave a receipt binding HEAD, working-tree digest, configuration digest, command, and tool version; the Stop chain shall accept the check only when every receipt field equals the current value, and shall treat an absent receipt or any differing field as not run, never as passed. **When** the Codex Stop chain has continued the turn `unmeasured` for the sync gate or the codex review gate N consecutive times for the same gate, HEAD, and working-tree digest (N a declared constant, design.md §D3.8), it shall allow the stop, shall write a discard record and reason text naming the gate `unverified`, and shall reset the count when a fresh receipt is read or HEAD or the working-tree digest changes; no verdict or registry entry shall read that gate as passed.

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
| SPEC-CODEX-EVENT-COVERAGE-001 (completed) | The 12-row event table, the Interrupt row (REQ-CEV-001..006), and the codex-cli 0.153.4 firing campaign (REQ-CEV-007..011) with the trigger-not-achieved vs not-fired distinction | This SPEC adapts the four rows that SPEC held back. REQ-CEV-005 ("M1 shall not register a new dispatcher subcommand") was scoped to that SPEC's M1; REQ-HPR-012 adds the Interrupt path in `internal/cli`/`internal/codexadapter`, keeping REQ-CEV-002 (no constant in `internal/hook`). REQ-HPR-013 inherits its trigger-not-achieved distinction. Adapting the four held-back rows reverses three of that SPEC's requirements, each as an intentional amendment in M2e (plan.md §F, research.md §R1.13): REQ-CEV-001 (the Interrupt row has an empty `DispatcherArg` and `Adapted` false — pinned by `TestEventTableMapping`), REQ-CEV-003 (`Resolve("Interrupt")` refuses with the unadapted class — pinned by `TestResolveRecognizedButUnadapted` and `TestResolveInterruptNoCounterpart`), and REQ-CEV-004 (no Interrupt handler installed — pinned by `TestRenderHooks_InterruptNeverInstalled`). REQ-CEV-006's census comment in `events.go` is rewritten to the new 12-adapted truth, which keeps that requirement's intent (the comment states the truth after the change). The campaign must be re-run: the installed codex-cli is 0.155.1, not 0.153.4 (research.md §R1.7). |
| SPEC-CODEX-DUAL-AGENTS-001 (completed) | Neutral agent source → `.codex/agents/*.toml`; per-agent Claude `hooks:` frontmatter documented as having no Codex per-agent equivalent; Codex-side permission enforcement excluded | No requirement overlap. The per-agent hook drop stays a documented drop; role-level permission enforcement belongs to sibling card t1100 (AC-AGENT-01). |
| SPEC-CODEX-WIRING-001 | `RenderHooks` merge model and `--harness codex` runtime flag | This SPEC changes what the table renders (adapted rows, Stop composition); the merge/preservation contract of that SPEC is unchanged. |

## §E Completion condition — this SPEC closes as partial (live-uncertified)

Operator decision Q5 (2026-09-23): no live Claude Code or Codex run is executed in this SPEC. The
live legs of AC-HPR-004, 007, 008, 009, 010, 011, 012, 020, and 021 are built as opt-in tests and
stay `NOT_RUN`. `NOT_RUN` is not PASS. The SPEC therefore closes as **partial (live-uncertified)**:
every unit and golden AC must be PASS under acceptance.md rule P, and the aggregate verdict is
reported as not-PASS for every obligation that needs a live leg. Live certification is deferred to
a separate follow-up card. No artifact of this SPEC may describe the hook, approval, or goal layer
as certified on either host.

Durable marker (plan-audit iter-2 N5; placement revised by iter-3 R4). The status enum has no
"partial" value, so the sync phase writes `status: completed`, and the partial state is carried in
four places a reader of the SPEC cannot miss. Each is written by the agent that owns its surface
(`.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings:
manager-docs may change only `status:` and `updated:` in spec.md, plan.md, and acceptance.md):

- **Plan phase, manager-spec (already written in v0.4.0):**
  - the `live-uncertified` token in the frontmatter `tags`;
  - the HISTORY line "v0.4.0 (closure mode)" stating that the SPEC closes partial
    (live-uncertified) and that live certification goes to a follow-up card.
- **Sync phase, manager-docs (its own surfaces):**
  - a progress.md §E.4 statement quoting the not-PASS aggregate verdict, listing the `NOT_RUN`
    live legs, and naming the follow-up card's id;
  - a CHANGELOG entry for this SPEC that says "partial (live-uncertified)" and names the same
    follow-up card.

Why the two frontmatter/HISTORY markers are written now and not at sync: the partial closure is
already decided (Q5), so their content does not depend on any run-phase result, and writing them
at sync would require manager-docs to edit spec.md body content, which it may not do. Everything
that does depend on run results — the aggregate verdict, the `NOT_RUN` list, the follow-up card id
— lives on surfaces manager-docs owns. A sync close missing either sync-phase marker is incomplete,
and a sync close that removes either plan-phase marker is a defect.

## §E.1 Constraints

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

### Out of Scope — non-Stop multi-handler chains

AC-HOOK-01 is narrowed to the Stop chain here. Design §08 lists other chains, and the Claude
template registers several handlers on some of those events. Codex renders one handler per event
(`internal/codexwiring/hooks.go:97–111`). Handler counts measured from
`internal/template/templates/.claude/settings.json.tmpl` at `530d8cc06`; some entries sit inside
`{{ if .HookOptIn.Enabled }}` branches:

- PreToolUse: 4 entries, all `handle-pre-tool.sh`.
- PostToolUse: 5 entries — `handle-post-tool.sh`, 3× `status-transition-ownership.sh`, `handle-harness-observe.sh`.
- SubagentStop: 3 entries — `handle-subagent-stop.sh`, `handle-harness-observe-subagent-stop.sh`, `chain-event.sh`.
- SessionStart: 2 entries — `handle-session-start.sh`, `handle-session-start-navigator.sh`.
- UserPromptSubmit: 2 entries — `handle-user-prompt-submit.sh`, `handle-harness-observe-user-prompt-submit.sh`.
- Owner: a follow-up card to be issued by the lead. The chains are registered in the obligation registry as `unverified` so the aggregate verdict cannot read PASS while they are unexamined.

### Out of Scope — adjacent items not assigned to M2

- Policy delivery completeness for standing/scoped rules (M1, AC-TPL-01/02). The obligation registry built here records those obligations; building their Codex delivery path is M1 work.
- The doctor/status readiness display part of AC-OBS-01. Its verdict-integrity part (attribution; rejecting skips, empty runs, and stale receipts) is in scope through REQ-HPR-019 and REQ-HPR-022..024.
- A Claude-side user-interrupt cancellation producer. The Claude Stop hook does not fire on user interrupt, and the only interrupt signal (`IsInterrupt`, `internal/hook/types.go:249`) belongs to PostToolUseFailure, which this SPEC excludes below. This source is recorded in the obligation registry as `UNSUPPORTED` with that evidence; it is not built here.
- Desktop, Web, and non-macOS runtime certification; verdicts from this SPEC are scoped to the OS they ran on.
- Claude events with no Codex counterpart (`TaskCompleted`, `TeammateIdle`, `StopFailure`, `PostToolUseFailure`, `Notification`) — parity for them is not claimed.
