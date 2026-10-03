---
id: SPEC-FACTORY-DECISION-AUTO-001
title: "Factory decision automation: decision board, PASS-WITH-DEBT admission, audit-ceiling policy, audit kickoff decider, FOUNDER defaults, wake latency, messaging degradation (card t1481)"
version: "0.4.0"
status: in-progress
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/decision, internal/cli, internal/contract, internal/homestate, internal/hook, internal/config, .claude/rules/moai/workflow, .claude/agents/moai, .claude/workflows, .claude/skills/moai-lane-watchdog"
lifecycle: spec-anchored
tags: "factory, autonomy, decision-board, pass-with-debt, audit-ceiling, kickoff, founder-default, wake-latency, factory-messaging, t1481"
tier: L
related_specs: [SPEC-DECISION-AUTHORITY-001, SPEC-FACTORY-RECORD-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-AUTONOMY-BATCH-GATE-001, SPEC-FACTORY-STALE-RUN-HEAL-001, SPEC-SYNC-PARALLEL-DOCS-001]
---

# SPEC-FACTORY-DECISION-AUTO-001 — Codify the Leader's Recurring Rulings

## HISTORY

| Version | Date | Author | Description |
|---|---|---|---|
| 0.1.0 | 2026-10-03 | manager-spec | Initial Tier L authoring for card t1481 (operator directive 2026-10-03: survey the factory, remove the bottlenecks that make lanes wait on the leader). Eight concerns: decision board, authority-register extension, PASS-WITH-DEBT definition and admission, audit-ceiling policy, factory `audit` kickoff decider, FOUNDER default application, short wait recheck, factory-messaging bind cache and degraded-state surfacing, MCP staleness fallback. Re-measured at `d7112d005`. |
| 0.1.1 | 2026-10-03 | manager-spec | Leader decisions Q1-Q7 recorded (mission contract 07d28c4b; operator 2026-10-03: implement now, include in v3.2.0). Q1 board writes leader-only. Q2 manager-spec no-recommendation clause narrowed to judgment calls. Q3 product-level defined. Q4 one-shot recheck, 5 min floor. Q5 no degraded notice on an already-bound session; run M0 measures. Q6 REQ-SD-016 narrowed to the single `--decider audit` own-card shape. Q7 hold + split only. Observed one-delta-round rulings added as evidence. Plan gains M0. |
| 0.2.0 | 2026-10-03 | manager-spec | Plan-audit iter1 FAIL 0.74 (`.moai/reports/t1481/plan-audit-iter1.md`) revised in the auditor's order with leader decisions D2-D10 (decision-index Q8-Q15): D4 audit decider re-checks audit-ready, blocker/hold, open product-level FOUNDER rows and the `audited_sha` binding (D11); D5 bind cache never skips the run-state probe, retirement invalidates (new REQ); D2/D3 phase-scoped predicate with explicit fields incl. a must-pass field; D6 mechanical delta eligibility by `fix_scope` diff and identical REQ/AC id sets; D7 policy binds every session, final hit = ceiling + `auto_delta_rounds`; D8 binding run conditions re-read by sync-audit-4dim and sync-auditor, undisposed = must-pass FAIL; D10 waits resolve only by a `resolves` reference. New: the release-blocking AC-wording exception observed on t1458 (three conditions, board record). D12 recast REQ-FDA-024; D13 removed "may". RED-now ledger re-measured with command, verbatim stdout and exit code at `ba2033d22`. REQs renumbered (merges: board verbs+lane refusal; watchdog read+record-before-message) so REQ-FDA-NNN ↔ AC-FDA-NNN; 25/25. |
| 0.3.0 | 2026-10-03 | manager-spec | Plan-audit iter2 FAIL 0.80 (`.moai/reports/t1481/plan-audit-iter2.md`) closed with leader decisions N1-N7 (decision-index Q17-Q23). N1 audit-decider authority = card record owner; T8a moves kickoff→run and leases to that owner atomically (baseline: kickoff holds no lease, `fr_lease_test.go:148-162`). N2 `decision-index.md` joins the plan-artifact hash input set and leaves REQ-FDA-011's exemption; DEFAULT-APPLIED therefore fills at plan close, before the audit. N3 decider refuses on any empty FOUNDER verdict. N4 release-blocking = listed in a `release-scope` standing board record or reachable through `depends`/`blocks` relation edges. N5 `defect_class`, `reread_hunks`, `blocking_count` (renamed from `blocking_findings`) and `scope` added to the verdict block; re-read = full `scope: reread` verdict admitted by the predicate; hunk limit diff-checked; hold released by `resolves`. N6 AC-FDA-019/020 release-blocking with probes of today's behavior; M0 values moved to measurement notes. N7 T13 label-only check unchanged. O1 `dispose_in`; O2 Class fail-closed fallback wording; O3 anchor hunk ranges; O4 split proposal informational outside a factory. |
| 0.4.0 | 2026-10-03 | manager-spec | Plan-audit iter3 FAIL 0.83 (`.moai/reports/t1481/plan-audit-iter3.md`); leader-ruled single delta round confined to the auditor's fix_scope (decision-index Q24-Q26). N8 DEFAULT-APPLIED passes the decider only on an implementation-level row carrying a `Default:` line. N9 SPEC-FACTORY-RECORD-001 added to related_specs; Amendments to REQ-FR-004 (T8a edge, AC-005 count 65/296 → 66/295) and REQ-FR-019 (decider `audit` on T8a only) obligated. N10 REQ-FDA-018 narrowed to verdicts recorded after the audited SHA. N11 §B item 5 aligned with REQ-FDA-015. O6 the widened run Phase 1 skip-cache key stated as intended (design §5). |

## §A Context and Problem

### A.1 Origin

Card t1481, issued under the operator directive of 2026-10-03 (factory autonomy item ④). The
leader's investigation of the current mission session (base `42d8474de`) observed lanes
repeatedly stopping to wait for a leader reply on judgments that are **not** keep-set operator
gates. Every such ruling travelled as a chat message, was hand-copied into reports, and the same
ruling ("PASS-family + blocking 0 → autonomous Kickoff") was re-sent per card.

### A.2 Observed waits (none is a keep-set gate)

| Wait class | Observed instances | Root cause (research.md §2) |
|---|---|---|
| plan-audit ceiling hit | t1356 iter3/iter4, t1404 iter2 (score regression), t1409 ×3, t1458 (Tier M cap) | auditor text routes ceiling hits to the user channel; the leader's rulings were formulaic (one delta round, then hold or split) |
| PASS-WITH-DEBT Kickoff | t1377, t1409 F3, t1438 | doctrine blocks PASS-WITH-DEBT; three code sites admit it on the label alone; the token has no emission rule |
| audit-debt disposition | t1409 D1, t1438 OD-1..OD-8 (all FOUNDER, blank) | the authority register excludes mission contracts and leader rulings, so every row falls to FOUNDER |
| factory card Kickoff | every factory card | `guardKickoffDecision` and `factory decide` accept only the human decider; doctrine §9 says AUTONOMOUS |
| wake latency | every leader reply | factory messages arrive only at turn boundaries; an idle lane wakes on the 20-minute cron |
| "factory messaging degraded: context deadline exceeded" | repeated | the bind pays a DB open and peer query on every prompt before the already-bound early return; degraded inbox states are discarded |
| stale lane MCP | rc.23 lane lacked `codex_review` | intake never compares the MCP server build with the installed CLI |

The ceiling rulings were one rule applied by hand: in the 2026-10-03 mission session the leader
applied "one delta round" four times (t1399, t1454, t1469, t1458) and "hold" twice (t1356, and one
earlier instance), and granted one exception on t1458 (a release-blocking card whose only remaining
blocking defect was an acceptance-criterion wording defect; hunk-limited fix plus re-read) —
leader-reported, research.md §1. REQ-FDA-011..013 codify exactly these rulings.

### A.3 The ladder step that does not exist

`auto-semantics.md` §6 step ② and §11 define a decision board — a HOME-surface, append-only,
harness-neutral store that the watchdog polls — and the watchdog skill already reads it. No code
implements it (research.md P1-P3). Step ② therefore always reads "board empty" and the ladder
falls through to the lead-chat step.

## §B Solution Shape

Eight capabilities, each converting a recurring leader ruling into a rule or a disk record that
a lane reads without waiting:

1. **Decision board** — `moai decision record|read`, the §11 SSOT made real; messages become nudges.
2. **Authority register extension** — standing board records and signed mission contracts become
   citable authority once pinned into the committed tree.
3. **PASS-WITH-DEBT** — defined once per phase, admitted by §9.1/§9.2, enforced identically by the
   code sites; enumerated debts become binding run conditions re-read by both sync verdict owners.
4. **Audit-ceiling policy** — automatic delta rounds under a mechanical eligibility test; the final
   hit becomes a hold record plus a split proposal, with one narrow release-blocking exception.
5. **Factory `audit` kickoff decider** — kickoff approve by verdict-file evidence, kickoff→run,
   leasing the card to its record owner, refused whenever a keep-set or product-level condition is open.
6. **FOUNDER defaults** — implementation-level rows carrying a rule-selected Default are recorded
   DEFAULT-APPLIED; only product-level verdicts block Kickoff.
7. **Wait recheck** — a one-shot recheck re-armed while a wait-on-leader record is open.
8. **Messaging and intake hygiene** — bind cache behind the run-state probe, degraded inbox states
   surfaced at warn with a rate-limited notice, MCP build comparison with a recorded CLI fallback.

The keep-set is unchanged and stays human (§C.9).

## §C Requirements (GEARS)

### C.1 Decision board

- **REQ-FDA-001** (Ubiquitous) — The `moai decision` command shall provide a `record` verb and a
  `read` verb over exactly one append-only, one-record-per-line board per project, stored under the
  moai home state directory keyed by the primary checkout's project key, shall never read or write a
  board file inside a working tree, and shall refuse `record` before any file I/O when invoked by a
  session for which lane refusal holds, while `read` stays available to every session.
- **REQ-FDA-002** (Ubiquitous) — Each board record shall carry a record id, a scope that is either
  `card:<card-id>` or `standing`, a kind from a closed enumeration, the three §10 fields
  (`decided_by`, `evidence_refs`, `ladder_path`), a ruling body, a UTC creation time, an optional
  `supersedes` record id, and an optional `resolves` wait id; a `standing` record's body shall state
  the predicate that selects the situations it governs, and a `standing` record of kind
  `release-scope` shall carry a release identifier and a card-id list as structured fields.
- **REQ-FDA-003** (Event-driven) — When `moai decision read` is invoked for a card scope, the command
  shall return that card's records together with every standing record, omit records superseded by a
  later record while keeping them in the file, and report an absent board, an empty board, and the
  count of unparseable lines as explicit statuses rather than as an empty success.
- **REQ-FDA-004** (Event-driven) — When the lane watchdog runs ladder step ② on any wake, the
  watchdog shall read the board for its card scope and the standing scope and follow a matching
  non-superseded ruling before falling through to step ③; when the leader resolves a judgment a lane
  is waiting on, the leader shall record the ruling on the board before sending any message, shall
  send only the record id as a nudge, and shall not re-send a ruling a standing record governs.

### C.2 Authority register

- **REQ-FDA-005** (Capability gate) — Where `interview.decision_gate` is `on`, the authority
  register shall additionally admit a standing board record and a signed mission contract, each cited
  by its identifier plus content digest and pinned by copying the cited line verbatim into the SPEC's
  committed `decision-index.md` row; a citation that cannot be pinned or whose digest does not match
  shall route the row to `FOUNDER` exactly as an unverifiable anchor does today.

### C.3 PASS-WITH-DEBT

- **REQ-FDA-006** (Ubiquitous) — The plan-auditor's verdict block shall carry machine fields for the
  verdict label, the aggregate score, the count of failed must-pass criteria (`must_pass_failed`), the
  count of blocking findings (`blocking_count`), the plan-artifact hash, `audited_sha`, a `scope` of
  `full`, `delta`, or `reread`, `fix_scope` (file plus anchor list), `defect_class` per blocking finding
  from a closed enumeration that includes `ac-wording` and `design`, `reread_hunks` (file plus anchor
  list), and — for `PASS-WITH-DEBT` — a `debts` list whose items each carry an id, a description, and
  `dispose_in` (`run` or `sync`); the auditor shall emit `PASS-WITH-DEBT` only when `must_pass_failed`
  is zero, the score is at or above the tier's plan threshold, `blocking_count` is zero, and `debts`
  holds at least one item.
- **REQ-FDA-007** (Event-driven) — When the plan→run Kickoff evaluates a plan-audit verdict admitted by
  the REQ-FDA-009 plan-phase predicate, with audit-ready status recorded, plan-artifact hashes
  unchanged since the verdict, and no open blocker, the Kickoff shall proceed autonomously under §9.1
  whether the label is `PASS` or `PASS-WITH-DEBT`, and the §9.2 batch summary shall classify the row
  approvable.
- **REQ-FDA-008** (Event-driven) — When a run starts on a `PASS-WITH-DEBT` Kickoff, the run shall copy
  every enumerated debt item into the SPEC's `progress.md` under a `Binding run conditions` heading;
  the `sync-audit-4dim` workflow and the sync-auditor agent shall each re-read every binding run
  condition, and an undisposed condition shall be a failed must-pass criterion that caps the sync
  verdict at `FAIL`.
- **REQ-FDA-009** (Ubiquitous) — The contract verdict rule, the kickoff decision evaluator, and the
  card-transition verdict guard shall apply one shared admission predicate parameterized by phase: for
  the plan phase (T7 and the Kickoff sites) it shall check the label, the score against the tier's plan
  threshold, `must_pass_failed` = 0, `blocking_count` = 0, the `debts` schema for `PASS-WITH-DEBT`,
  and the plan-artifact hash; for the sync phase (T13) it shall keep today's label-only check unchanged
  (`PASS` and `PASS-WITH-DEBT` admitted), the only sync change being the REQ-FDA-008 binding-condition
  must-pass, which reaches T13 through the sync verdict label; `FAIL`, `INCONCLUSIVE`, `BYPASSED`, and
  an absent verdict shall be refused in both phases. Plain `PASS` at T7 thereby gains the score,
  must-pass, and blocking checks it does not have today.

### C.4 Audit-ceiling policy

- **REQ-FDA-010** (Ubiquitous) — The plan-audit ceiling policy shall be a configuration setting beside
  the tier ceilings, carrying the number of automatic delta rounds (`auto_delta_rounds`, default 1)
  and the final-hit disposition (`hold-and-split`, the only v1 value); the tier ceilings in that
  configuration shall be the only iteration caps any agent or rule text states, and the policy shall
  bind every session that runs a plan audit, lane or not.
- **REQ-FDA-011** (Event-driven) — When a plan-audit reaches its tier ceiling without an admitted
  verdict and the ceiling-hit verdict carries a non-empty `fix_scope` (a list of file plus anchor
  entries), the session shall run the next delta audit without asking anyone and write one decision
  record citing the policy setting; the delta round shall count as eligible only if the diff between the
  ceiling-hit `audited_sha` and the delta round's `audited_sha` touches nothing outside `fix_scope`
  except `progress.md` and `.moai/reports/**`, and the REQ id set and AC id set are identical across
  the two SHAs; an ineligible or `fix_scope`-less round shall go directly to the final-hit path. The
  hunk range of an anchor is its heading section, and for a REQ or AC id it is the requirement bullet
  through the next bullet or the matrix row plus that id's scenario section.
- **REQ-FDA-012** (Event-driven) — When the audit iteration count reaches the tier ceiling plus
  `auto_delta_rounds` without an admitted verdict, or a delta round is ineligible, or the auditor emits
  a score-regression STOP, the session shall stop plan iteration and write a hold wait record and a
  split proposal to the card's evidence path (outside a card, the SPEC's `progress.md`); a non-lane
  orchestrator shall then inform the user, and any question it asks shall only offer an override;
  inside a factory, card creation and queue mutation stay with the leader, and outside a factory the
  split proposal is informational for the user.
- **REQ-FDA-013** (Event-driven) — When the final hit occurs and all three conditions hold — the card
  is release-blocking, meaning its id is listed in a non-superseded standing board record of kind
  `release-scope`, or a listed card reaches it through todo relation edges of kind `depends` (listed
  card `depends` on it) or `blocks` (it `blocks` the listed card), followed transitively; the final
  verdict reports `blocking_count` = 1 with `defect_class: ac-wording`; and the verdict lists
  `reread_hunks` — the session shall, only after the leader writes a `card:<id>` board record naming
  this exception, change only the listed hunks, verified by diffing the final-hit `audited_sha` against
  the fix SHA (nothing outside `reread_hunks` except `progress.md` and `.moai/reports/**`), and shall
  obtain from the auditor a full verdict block with `scope: reread` at the fix SHA that the REQ-FDA-009
  plan-phase predicate admits; the hold shall be released only by a board record whose `resolves` names
  the hold's wait id; if any condition is absent the hold of REQ-FDA-012 stands.

### C.5 Factory audit kickoff decider

- **REQ-FDA-014** (Event-driven) — When the kickoff approve transition is requested with decider
  `audit`, the card state machine shall accept it only if the plan-audit verdict file is admitted by the
  REQ-FDA-009 plan-phase predicate, its `audited_sha` equals the card's evidence SHA, its plan-artifact
  hash equals the hash computed from the current plan artifacts — whose input set shall include
  `decision-index.md`, so any change to it after the audited SHA is a mismatch — audit-ready status is
  recorded, the card carries no open blocker and no operator hold, and the SPEC's `decision-index.md`
  holds no `FOUNDER` row of either class whose verdict is empty (only a recorded verdict or a
  `DEFAULT-APPLIED` verdict passes), and every row holding a `DEFAULT-APPLIED` verdict carries
  `Class: implementation-level` and a `Default:` line — a `DEFAULT-APPLIED` verdict on a
  `product-level` row, on a row with no `Class:` line, or on a row with no `Default:` line refuses;
  on any failed condition it shall refuse and leave the `human`
  Kickoff path as the only route, and it shall reject every decider value other than `human` and
  `audit`.
- **REQ-FDA-015** (Event-driven) — When a kickoff is approved by the `audit` decider, the card shall
  move from kickoff to the run stage and, in the same atomic transition, take a lease for its record
  owner (the owner label persisted on the card row through T7) exactly as the normal lease path binds
  one; the `human` decider path shall keep its present behavior. A kickoff card holds no lease today
  (`internal/homestate/fr_lease_test.go:148-162`), which this transition preserves up to the move.
  SPEC-FACTORY-RECORD-001 shall carry an Amendments row on REQ-FR-004 adding the T8a edge (kickoff →
  run) to its transition table, with its AC-005 requested-pair count moving from 65 accepted / 296
  refused to 66 accepted / 295 refused.
- **REQ-FDA-016** (Event-driven) — When a lane session whose label is the card's record owner invokes
  `factory decide` for that card with `--gate kickoff --choice approve --decider audit`, the command
  shall admit the call;
  every other lane invocation of `factory decide` shall remain refused, and REQ-SD-016 of
  SPEC-FACTORY-SELF-DISPATCH-001 shall carry an Amendments row naming this single exception;
  REQ-FR-019 of SPEC-FACTORY-RECORD-001 shall carry an Amendments row admitting decider `audit` on
  the T8a edge only (an `audit` approval moves the card to `run` with the owner's lease) while the
  `human` decider path to `assigned` stays unchanged.

### C.6 FOUNDER defaults

- **REQ-FDA-017** (Capability gate) — Where `interview.decision_gate` is `on`, every `FOUNDER` row shall
  carry a `Class:` line valued `product-level` or `implementation-level`, and a row whose options the
  published Default rule ranks shall carry a `Default:` line and an `Alternate:` line (first rule: the
  option that preserves current behavior); the REQ-FDA-018 closed list governs which Class a row is
  given, and a row without a `Class:` line shall be treated as `product-level` as a fail-closed
  fallback; the manager-spec decision-index clause that forbids an embedded recommendation shall
  state that it governs judgment calls only and that a rule-selected Default is a policy application,
  changed in the template source first.
- **REQ-FDA-018** (Event-driven) — When the plan phase closes, before the plan audit, on an
  `implementation-level` `FOUNDER` row that carries a `Default:` line and an empty operator verdict, the
  authoring session shall fill the verdict line with `DEFAULT-APPLIED`, the UTC time, and the deciding
  runner and role, so the audited hash covers the applied default; a `product-level` row, or an
  `implementation-level` row without a `Default:` line, shall stay empty, block the autonomous Kickoff,
  and be routed to the operator; a verdict recorded after the audited SHA reaches run only through the
  `human` Kickoff path, while a verdict recorded before the audited SHA is covered by the audited hash
  and passes the `audit` decider under REQ-FDA-014. A row is `product-level` exactly when its decision changes a shipped command's default
  user-visible behavior, removes a user-facing feature, or changes a template default; every other row
  is `implementation-level`.

### C.7 Wait recheck

- **REQ-FDA-019** (State-driven) — While the card's progress record holds a wait record with an id and
  `waiting_on` naming the leader, and no board record's `resolves` field names that wait id, the lane
  shall keep a one-shot recheck at the configured delay (default 5 minutes, never below 5) re-armed
  after each recheck that finds the wait still open, in addition to the standing 20-minute carrier;
  a board record for the same card that does not name the wait id shall leave the wait open.

### C.8 Messaging and intake hygiene

- **REQ-FDA-020** (Event-driven) — When a prompt-submit bind finds a cached binding whose session, run,
  owner PID, and process start all match the current invocation, the hook shall still run the
  run-state probe and shall skip opening the factory messaging database and the peer query only when
  the probe reports the same live run, so an already-bound session on a live run receives no degraded
  bind notice.
- **REQ-FDA-021** (Event-driven) — When the run-state probe reports the cached run retired or a
  different run, the hook shall invalidate the cache and take the full rebind path in the same
  invocation; when the probe itself fails on a cache hit, the hook shall report the degraded state
  rather than suppress it.
- **REQ-FDA-022** (Event-driven) — When an inbox claim returns a degraded state, the hook shall log it
  at warn level with the state string and surface a notice to the session at most once per configured
  interval per session.
- **REQ-FDA-023** (Event-driven) — When a lane takes a card, the lane shall compare the build of the
  running MCP server with the installed `moai version`, and on a mismatch or on a required tool absent
  from the server's tool list shall record `fallback=CLI` in the card's progress record and use the
  documented CLI equivalent for each affected tool.

### C.9 Keep-set preservation

- **REQ-FDA-024** (Ubiquitous) — The decision board, the ceiling policy, the `audit` decider, and the
  FOUNDER default shall refuse to open, approve, or record a decision for origin push, release or
  `main` integration, queue admission, contract signing, destructive disposal, a product-level
  verdict, or a final PASS/FAIL verdict.
- **REQ-FDA-025** (Ubiquitous) — Every shipped rule, agent, skill, workflow, and configuration file this
  SPEC changes shall carry the same change in its template mirror, and every changed agent definition
  shall have its emitted counterpart regenerated rather than hand-edited.

## §D Constraints

- Template-first: shipped rules/agents/skills/workflows/config change in
  `internal/template/templates/**` first; `make build`; `make agents-emit` for every changed
  `.claude/agents/moai/*.md` mirror.
- Template neutrality: no SPEC IDs, card ids, dates, or SHAs in template content.
- The board is a HOME-surface store; no `.moai/state/**` tree-local copy (doctrine §11 "ghost").
- Lane refusal (REQ-SD-015/016 of SPEC-FACTORY-SELF-DISPATCH-001) is relaxed only for the single
  REQ-FDA-016 call shape.
- Authority-gate invariant (§7): FAIL/INCONCLUSIVE never auto-proceed; availability failures never
  read as positive.
- The run-state probe of SPEC-FACTORY-STALE-RUN-HEAL-001 (REQ-SRH-004/005) runs on every bind,
  cache hit or not.
- Cache cost: every recheck fire after the cache window rewrites the prefix once; the one-shot recheck
  exists only while a wait-on-leader record is open.

## §E Boundary notes

### E.1 What the board is not

The board is not a queue (queue mutation stays `moai gtd`, leader-only), not an evidence store
(card evidence stays under `.moai/reports/<card-id>/`), and not an approval channel for keep-set
gates. It records leader rulings and standing rules so a lane can read them instead of waiting.

### E.2 Self-attestation

Decision records remain self-attested; the sync audit's re-read stays the compensating control
(§10). The `audit` decider narrows the gap with the `audited_sha` and hash bindings but does not
close it.

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

- `.moai/reports/t1481/plan-audit-iter1.md` (iter1 verdict, verbatim)
- `.claude/rules/moai/workflow/auto-semantics.md` §5.1, §6, §7, §9, §9.1, §9.2, §10, §11, §14
- `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract, verdict block
- `.claude/agents/moai/manager-spec.md` decision-index rows, authority register
- `.claude/agents/moai/sync-auditor.md`, `.claude/workflows/sync-audit-4dim.js`
- `.moai/config/sections/harness.yaml` (`plan_audit_tier_ceilings`)
- `internal/contract/rules.go`, `internal/contract/kickoff/decide.go`, `internal/homestate/card_transition.go` (T7/T13 share `guardVerdictPass`), `internal/homestate/card_evidence_readers.go` (`audited_sha` binding)
- `internal/cli/factory_card.go`, `internal/hook/factory_messages.go`, `internal/hook/user_prompt_submit.go`
- research.md (probe ledger), design.md (mechanism), acceptance.md (AC matrix), decision-index.md
