---
id: SPEC-FACTORY-DECISION-AUTO-001
title: "Factory decision automation: decision board, PASS-WITH-DEBT admission, audit-ceiling policy, audit kickoff decider, FOUNDER defaults, wake latency, messaging degradation (card t1481)"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/decision, internal/cli, internal/contract, internal/homestate, internal/hook, internal/config, .claude/rules/moai/workflow, .claude/agents/moai, .claude/skills/moai-lane-watchdog"
lifecycle: spec-anchored
tags: "factory, autonomy, decision-board, pass-with-debt, audit-ceiling, kickoff, founder-default, wake-latency, factory-messaging, t1481"
tier: L
related_specs: [SPEC-DECISION-AUTHORITY-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-AUTONOMY-BATCH-GATE-001, SPEC-FACTORY-STALE-RUN-HEAL-001, SPEC-SYNC-PARALLEL-DOCS-001]
---

# SPEC-FACTORY-DECISION-AUTO-001 — Codify the Leader's Recurring Rulings

## HISTORY

| Version | Date | Author | Description |
|---|---|---|---|
| 0.1.0 | 2026-10-03 | manager-spec | Initial Tier L authoring for card t1481 (operator directive 2026-10-03: survey the factory, remove the bottlenecks that make lanes wait on the leader). Eight concerns: decision board, authority-register extension, PASS-WITH-DEBT definition and admission, audit-ceiling policy, factory `audit` kickoff decider, FOUNDER default application, short wait recheck, factory-messaging bind cache and degraded-state surfacing, MCP staleness fallback. Re-measured at `d7112d005`. |

## §A Context and Problem

### A.1 Origin

Card t1481, issued under the operator directive of 2026-10-03 (factory autonomy item ④). The
leader's investigation of the current mission session (base `42d8474de`) observed lanes
repeatedly stopping to wait for a leader reply on judgments that are **not** keep-set operator
gates. Every such ruling travelled as a chat message, was hand-copied into reports, and the same
ruling ("PASS-family + blocking 0 → autonomous Kickoff") was re-sent per card.

### A.2 Observed waits (none is a keep-set gate)

| Wait class | Observed instances | Root cause (re-measured at `d7112d005`, research.md §2) |
|---|---|---|
| plan-audit ceiling hit | t1356 iter3/iter4, t1404 iter2 (score regression), t1409 ×3, t1458 (Tier M cap) | auditor text routes ceiling hits to the user channel; the leader's rulings were formulaic (one delta round, then hold or split) |
| PASS-WITH-DEBT Kickoff | t1377, t1409 F3, t1438 | doctrine blocks PASS-WITH-DEBT; three code sites admit it; the token has no emission rule |
| audit-debt disposition | t1409 D1, t1438 OD-1..OD-8 (all FOUNDER, blank) | the authority register excludes mission contracts and leader rulings, so every row falls to FOUNDER |
| factory card Kickoff | every factory card | `guardKickoffDecision` and `factory decide` accept only the human decider; doctrine §9 says AUTONOMOUS |
| wake latency | every leader reply | factory messages arrive only at turn boundaries; an idle lane wakes on the 20-minute cron |
| "factory messaging degraded: context deadline exceeded" | repeated | the bind pays a DB open and peer query on every prompt before the already-bound early return; degraded inbox states are discarded |
| stale lane MCP | rc.23 lane lacked `codex_review` | intake never compares the MCP server build with the installed CLI |

### A.3 The ladder step that does not exist

`auto-semantics.md` §6 step ② and §11 define a decision board — a HOME-surface, append-only,
harness-neutral store that the watchdog polls — and the watchdog skill already reads it. No code
implements it: no `moai decision` command, no store under the moai home (research.md P1). Step ②
therefore always reads "board empty" and the ladder falls through to the lead-chat step.

## §B Solution Shape

Eight capabilities, each converting a recurring leader ruling into a rule or a disk record that
a lane reads without waiting:

1. **Decision board** — `moai decision record|read`, the §11 SSOT made real; messages become nudges.
2. **Authority register extension** — standing board records and signed mission contracts become
   citable authority once pinned into the committed tree.
3. **PASS-WITH-DEBT** — defined once, admitted by §9.1/§9.2, enforced identically by the three code
   sites; enumerated debts become binding run conditions re-read by sync-audit.
4. **Audit-ceiling policy** — one automatic delta round under a mechanical eligibility test; a second
   hit becomes a hold record plus a split proposal. Card creation stays with the leader.
5. **Factory `audit` kickoff decider** — kickoff approve by verdict-file evidence, kickoff→run
   keeping the lease.
6. **FOUNDER defaults** — implementation-level rows carrying a Default marker are recorded
   DEFAULT-APPLIED; only product-level verdicts block Kickoff.
7. **Wait recheck** — a short recheck carrier while a wait-on-leader record is open.
8. **Messaging and intake hygiene** — bind cache, degraded inbox states surfaced at warn with a
   rate-limited notice, MCP build comparison with a recorded CLI fallback.

The keep-set is unchanged and stays human (§C.9).

## §C Requirements (GEARS)

### C.1 Decision board

- **REQ-FDA-001** (Ubiquitous) — The `moai decision` command shall provide a `record` verb and a
  `read` verb over exactly one append-only, one-record-per-line board per project, stored under the
  moai home state directory keyed by the primary checkout's project key; a board file inside any
  working tree shall never be read or written.
- **REQ-FDA-002** (Ubiquitous) — Each board record shall carry a record id, a scope that is either
  `card:<card-id>` or `standing`, a kind from a closed enumeration, the three §10 fields
  (`decided_by`, `evidence_refs`, `ladder_path`), a ruling body, a UTC creation time, and an
  optional `supersedes` reference; for a `standing` record the body shall state the predicate that
  selects the situations it governs.
- **REQ-FDA-003** (Event-driven) — When a session for which lane refusal holds invokes
  `moai decision record`, the command shall refuse before writing anything; `moai decision read`
  shall remain available to every session.
- **REQ-FDA-004** (Event-driven) — When `moai decision read` is invoked for a card scope, the command
  shall return that card's records together with every standing record, omit records superseded by a
  later record while keeping them in the file, and report an absent board, an empty board, and the
  count of unparseable lines as explicit statuses rather than as an empty success.
- **REQ-FDA-005** (Event-driven) — When the lane watchdog runs ladder step ② on any wake, the
  watchdog shall read the board for its card scope and the standing scope and follow a matching
  non-superseded ruling before falling through to step ③.
- **REQ-FDA-006** (Event-driven) — When the leader resolves a judgment a lane is waiting on, the
  leader shall record the ruling on the board before sending any message, and a message about that
  ruling shall carry only the record id as a nudge; the leader shall not re-send a ruling that a
  non-superseded standing record already governs.

### C.2 Authority register

- **REQ-FDA-007** (Capability gate) — Where `interview.decision_gate` is `on`, the authority
  register shall additionally admit a standing board record and a signed mission contract, each cited
  by its identifier plus content digest and pinned by copying the cited line verbatim into the SPEC's
  committed `decision-index.md` row; a citation that cannot be pinned or whose digest does not match
  shall route the row to `FOUNDER` exactly as an unverifiable anchor does today.

### C.3 PASS-WITH-DEBT

- **REQ-FDA-008** (Ubiquitous) — The plan-auditor shall emit `PASS-WITH-DEBT` only when every
  must-pass criterion passes, the aggregate score is at or above the SPEC tier's PASS threshold, the
  count of blocking findings is zero, and the verdict file enumerates at least one debt item with an
  identifier, a description, and the phase that must dispose of it.
- **REQ-FDA-009** (Event-driven) — When the plan→run Kickoff evaluates a `PASS-WITH-DEBT` verdict that
  satisfies REQ-FDA-008, with audit-ready status recorded, plan-artifact hashes unchanged since the
  verdict, and no open blocker, the Kickoff shall proceed autonomously under §9.1 and the §9.2 batch
  summary shall classify the row approvable.
- **REQ-FDA-010** (Event-driven) — When a run starts on a `PASS-WITH-DEBT` Kickoff, the run shall copy
  every enumerated debt item into the SPEC's `progress.md` as a binding run condition, and the sync
  auditor shall re-read each binding run condition and report an undisposed condition as a sync
  finding.
- **REQ-FDA-011** (Ubiquitous) — The contract verdict rule, the kickoff decision evaluator, and the
  card-transition verdict guard shall apply one shared admission predicate, so a verdict file labelled
  `PASS-WITH-DEBT` that lacks the enumerated debt section or reports blocking findings above zero is
  rejected identically at all three sites, and `FAIL`, `INCONCLUSIVE`, `BYPASSED`, and an absent
  verdict remain hard blocks at all three sites.

### C.4 Audit-ceiling policy

- **REQ-FDA-012** (Ubiquitous) — The plan-audit ceiling policy shall be a configuration setting beside
  the tier ceilings, carrying the number of automatic delta rounds (default 1) and the second-hit
  disposition (default hold-and-split), and the tier ceilings in that configuration shall be the only
  iteration caps any agent or rule text states.
- **REQ-FDA-013** (Event-driven) — When a plan-audit reaches its tier ceiling without a PASS-family
  verdict and the auditor's final verdict declares the remaining blocking findings delta-eligible —
  every remaining blocking finding is confined to the auditor's own required-fix text, the fix changes
  no requirement count, acceptance-criterion count, or scope section, and no previously fixed finding
  regressed — the lane shall run the configured number of delta audit rounds without asking anyone and
  write one decision record citing the policy setting.
- **REQ-FDA-014** (Event-driven) — When a plan-audit hits the ceiling a second time, or hits it with
  the delta-eligibility declaration absent or false, or emits a score-regression STOP, the lane shall
  stop plan iteration, write a hold wait record and a split proposal to the card's evidence path, and
  leave card creation and queue mutation to the leader.

### C.5 Factory audit kickoff decider

- **REQ-FDA-015** (Event-driven) — When the kickoff approve transition is requested with decider
  `audit`, the card state machine shall accept it only if the plan-audit verdict file satisfies the
  REQ-FDA-011 predicate and the plan-artifact hash recorded in that verdict equals the hash computed
  from the current plan artifacts, and shall reject every decider value other than `human` and `audit`.
- **REQ-FDA-016** (Event-driven) — When a kickoff is approved by the `audit` decider, the card shall
  move from kickoff to the run stage while keeping its current lease and owner; the `human` decider
  path shall keep its present behavior.
- **REQ-FDA-017** (Event-driven) — When a lane session invokes the kickoff approve decision with
  decider `audit` for the card its own lease holds, `factory decide` shall admit the call; every other
  lane invocation of `factory decide` shall remain refused.

### C.6 FOUNDER defaults

- **REQ-FDA-018** (Capability gate) — Where `interview.decision_gate` is `on`, a `FOUNDER` row shall
  carry a `Class:` line valued `product-level` or `implementation-level` and may carry a `Default:`
  line and an `Alternate:` line, where the Default is selected by the stated reversibility rule rather
  than by preference; a row without a `Class:` line shall be treated as `product-level`.
- **REQ-FDA-019** (Event-driven) — When the Kickoff reaches an `implementation-level` `FOUNDER` row that
  carries a `Default:` line and an empty operator verdict, the Kickoff shall fill the verdict line with
  `DEFAULT-APPLIED`, the UTC time, and the deciding runner and role, and shall not block on that row;
  an empty verdict on a `product-level` row shall block the autonomous Kickoff and route the row to the
  operator.

### C.7 Wait recheck

- **REQ-FDA-020** (State-driven) — While the card's progress record holds a wait record whose
  `waiting_on` names the leader and no later board record for that card resolves it, the lane shall keep
  a short recheck carrier at the configured cadence (default 5 minutes) in addition to the standing
  20-minute carrier, and shall delete the short carrier when the wait resolves.

### C.8 Messaging and intake hygiene

- **REQ-FDA-021** (Event-driven) — When a prompt-submit bind finds a cached binding whose session,
  run, owner PID, and process start all match the current invocation, the hook shall return without
  opening the factory messaging database for the bind.
- **REQ-FDA-022** (Event-driven) — When an inbox claim returns a degraded state, the hook shall log it
  at warn level with the state string and surface a notice to the session at most once per configured
  interval per session.
- **REQ-FDA-023** (Event-driven) — When a lane takes a card, the lane shall compare the build of the
  running MCP server with the installed `moai version`, and on a mismatch or on a required tool absent
  from the server's tool list shall record `fallback=CLI` in the card's progress record and use the
  documented CLI equivalent for each affected tool.

### C.9 Keep-set preservation

- **REQ-FDA-024** (Unwanted) — The decision board, the ceiling policy, the `audit` decider, and the
  FOUNDER default shall not open, approve, or record a decision for origin push, release or `main`
  integration, queue admission, contract signing, destructive disposal, a product-level verdict, or a
  final PASS/FAIL verdict.
- **REQ-FDA-025** (Ubiquitous) — Every shipped rule, agent, skill, and configuration file this SPEC
  changes shall carry the same change in its template mirror, and every changed agent definition shall
  have its emitted counterpart regenerated rather than hand-edited.

## §D Constraints

- Template-first: shipped rules/agents/skills/config change in `internal/template/templates/**`
  first; `make build`; `make agents-emit` for every changed `.claude/agents/moai/*.md` mirror.
- Template neutrality: no SPEC IDs, card ids, dates, or SHAs in template content.
- The board is a HOME-surface store; no `.moai/state/**` tree-local copy (doctrine §11 "ghost").
- Lane refusal (REQ-SD-015/016 of SPEC-FACTORY-SELF-DISPATCH-001) is relaxed only for the single
  REQ-FDA-017 call shape.
- Authority-gate invariant (§7): FAIL/INCONCLUSIVE never auto-proceed; availability failures never
  read as positive.
- Cache cost: every added recheck fire after the cache window rewrites the prefix once; the short
  carrier exists only while a wait-on-leader record is open.

## §E Boundary notes

### E.1 What the board is not

The board is not a queue (queue mutation stays `moai gtd`, leader-only), not an evidence store
(card evidence stays under `.moai/reports/<card-id>/`), and not an approval channel for keep-set
gates. It records leader rulings and standing rules so a lane can read them instead of waiting.

### E.2 Self-attestation

Decision records remain self-attested; the sync audit's re-read stays the compensating control
(§10). This SPEC adds re-read obligations (REQ-FDA-010) but adds no prevention mechanism.

## §F Exclusions

### Out of Scope — queue and leader powers

- Automatic card creation from a split proposal (the leader issues cards).
- Any change to queue admission, `done`, `drop`, or relation verbs.
- Final merge approval, CodeRabbit adjudication, and cross-session dispute coordination.

### Out of Scope — merge window

- Merge-window nomination automation (handled by card t1479).
- Changes to `moai integration acquire|release`.

### Out of Scope — messaging transport

- Idle-wake message delivery (the runtime delivers factory messages only at turn boundaries; this
  SPEC shortens the recheck, it does not add a push channel).
- Session-messaging broker changes for the codex runner.

### Out of Scope — audit content

- Changes to the plan-auditor's scoring rubric, must-pass criteria, or tier thresholds.
- Sync-audit 4-dimension scoring changes beyond the binding-run-condition re-read.

### Out of Scope — Jev

- Any use of Jev as an input to Kickoff, ceiling, or FOUNDER decisions.

## §G References

- `.claude/rules/moai/workflow/auto-semantics.md` §5.1, §6, §7, §9, §9.1, §9.2, §10, §11, §14
- `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract (`:693`, `:703`), verdict block `:203`
- `.claude/agents/moai/manager-spec.md:110-115` (decision-index rows, authority register)
- `.claude/rules/moai/workflow/spec-workflow.md:158` (plan-auditor escalation)
- `.moai/config/sections/harness.yaml:75-78` (`plan_audit_tier_ceilings`)
- `internal/contract/rules.go:25`, `internal/contract/kickoff/decide.go:393`, `internal/homestate/card_transition.go:109,470,480`
- `internal/cli/factory_card.go:1979-1984`, `internal/hook/factory_messages.go:145-153,210`, `internal/hook/user_prompt_submit.go:62,156-168`
- research.md (measured ledger), design.md (mechanism), acceptance.md (AC matrix)
