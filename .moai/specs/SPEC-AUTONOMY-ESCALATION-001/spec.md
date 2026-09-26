---
id: SPEC-AUTONOMY-ESCALATION-001
title: "Contract-mode escalation detector: mechanical detection of the six escalate_on classes plus operational trips, reported as an escalation record without blocking or mutating the queue"
version: "0.4.3"
status: in-progress
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "autonomy,contract,escalation,detector,hook,needs-decision"
tier: L
related_specs: [SPEC-AUTONOMY-TIERS-001, SPEC-ACSNAPSHOT-COMMIT-GUARD-001]
---

# SPEC-AUTONOMY-ESCALATION-001 — Contract-mode escalation detector

## §A — History

- **2026-09-26** — v0.1.0 plan-phase draft (card t1235, contract-based autonomous harness
  track A2). Base tree `develop` at `ca1d5dc43`. Every "existing asset" premise in the
  card's design table was measured against this tree before any requirement was written;
  the measured table (asset → exists? → file:line) lives in `research.md` §B.
- **2026-09-26** — v0.1.1 contract-field alignment to the A1 schema draft at commit `8f77d9a33`.
  Dependency tag 「A1 plan-audit 통과본으로 재확인」 introduced.
- **2026-09-26** — v0.2.0 after plan-audit iteration 1 (FAIL 0.79): lane-owned defects D4-D16
  repaired; lead rulings 09-26 #1-#6 folded in (§H).
- **2026-09-26** — v0.2.1 added two A3 preconditions; **moved out at v0.3.0** (§K).
- **2026-09-26** — v0.3.0 after plan-audit iteration 2 (FAIL 0.82): lead rulings 09-26 (2) folded
  in; A3 preconditions and the mission-validator projection moved to card t1245. Requirements
  25 → 23, criteria 25 → 23.
- **2026-09-26** — v0.4.0 after plan-audit iteration 3 (FAIL 0.83). Lead rulings 09-26 (3) folded
  in (§H): the resolver matches a `card:` field in the contract instead of the queue `spec_id`
  (P1); one unified disarm rule writes exactly one record per card for every way an armed card
  becomes unarmed, and class 10 is renamed `contract-void` → `detection-disarmed` (P1, P2, P4);
  A1 verify runs at PreToolUse (P3, B5); the card state file leaves the `.moai/state/`
  exemption and gains tamper detection (P4); B1-B6 are applied, including the five acceptance
  reasons with `signature_acceptance_mismatch` (B2). §F was re-pinned first to `6d98ca466` and then,
  by a later lead instruction, to A1 v0.5.0 at `67a2f55cb` (single kickoff decider; §F, §I.1).
  Requirements stay 23; criteria 23 → 24. Mapping: acceptance.md §A.
- **2026-09-26** — v0.4.1 lead ruling 09-26 (4) folded in (§H): decider rules removed from this
  SPEC (the `decider` value follows the A1 schema; judgment rules are A3's), and request R10
  closed — the card state and the detector audit log move to the moai-owned store
  `$MOAI_HOME/db/<project-key>/contract/`.
- **2026-09-26** — v0.4.2 after plan-audit iteration 4 (FAIL 0.83), findings Q3-Q5 and m2 only
  (Q1 and Q2 await a lead design ruling): contract-store writes are judged before every class 3
  exemption; the §G residual risk states what an unkeyed chain can and cannot show; §F re-pinned
  to A1 v0.5.2 at `25283ebf8`, which closes R8, R9, and R10; the `spec.md` `status` read joins
  the C4 operation list. Lead ruling 09-26 (5) folded in (resolves Q1 and Q2): one audit log per
  card, and a card's own log entries — not the state file — decide whether it is armed. Criteria
  24 → 25 (two merged, three added; mapping in acceptance.md §A).
- **2026-09-26** — v0.4.3 after plan-audit iteration 5 (PASS-WITH-DEBT 0.87), closing R1, R2,
  n1, n2: a per-card exclusive lock serializes a card's state writes and log appends; a card log
  that does not show the card armed (absent, empty, or tail-truncated) is judged against
  unaccounted evidence of arming; the state-file digest is always compared; a broken chain is
  `state-tamper` whatever the arming status; §G states what the chain does and does not catch.
  Counts unchanged: 23 requirements, 25 criteria.

## §B — Problem

### B.1 What the card asks for

A contract-mode run (a SPEC carrying a `contract.yaml`, schema owned by card t1234 / A1)
lets an agent proceed without per-step human questions **as long as it stays inside the
contract**. Staying inside is only a guarantee if leaving is detected mechanically. The
contract names six `escalate_on` classes; the card adds three operational trips, and the lead
rulings add a fourth for an armed card that stops being armed:

| # | Class | Kind |
|---|---|---|
| 1 | acceptance-change | contract |
| 2 | invariant-violation | contract |
| 3 | ownership-move | contract |
| 4 | new-architecture-or-api | contract |
| 5 | contradictory-evidence | contract |
| 6 | irreversible-action | contract |
| 7 | budget-exceeded | operational |
| 8 | same-diagnostic-repeat (3 consecutive) | operational |
| 9 | audit-fail-at-retry-cap | operational |
| 10 | detection-disarmed | operational |

When a class trips, the card becomes *needs-decision* and an escalation record captures the
observation, the options, and the line that tripped. Escalation is a **report, not a
question**: the agent continues other work.

### B.2 Premises that did not hold as stated (measured, `research.md` §B)

1. **No needs-decision card state exists, and the store forbids adding one.** The queue
   states are exactly `queued` / `picked` / `dropped` (`internal/kanban/backlog_store.go:61-65`);
   `internal/kanban/backlog_schema_freeze_test.go:3-5,70` pins the three-state CHECK. The
   needs-decision marking is therefore a **record beside the queue** (REQ-AE-018).
2. **The AC-snapshot guard is not reusable as a detector.** `scripts/ac-baseline/check-staged.sh:1-4`
   declares itself a local-only dev tool comparing against a corpus snapshot. Class 1 consumes
   A1's own verify (§F).
3. **`graph_file_api` reads the working tree only.** `internal/graph/codequery.go:65-81` has no
   commit parameter; the before-side of a comparison comes from the base commit's blob
   (design.md §C.4).
4. **The queue `spec_id` cannot link a card to its SPEC.** The iteration-3 audit measured 0 of
   119 live cards carrying one (`.moai/reports/t1235/plan-audit-iter3.md` P1). The resolver
   therefore reads the card from the contract itself (REQ-AE-002). Keeping the queue `spec_id`
   filled is the factory record layer's (F1) concern; this SPEC does not rely on it.

Four further assets exist only as doctrine or partial coverage (frozen-file guard scoped to
the harness-learner identity only; `audit_multi` disagreement is a tri-state pointer; the
deny list blocks force-push but not a plain push to `main` or `git tag`; super-advisor E1 is
a doctrine row with no counter). Each is recorded with file:line in `research.md` §B.

## §C — Goal

Under `autonomy.mode: contract`, every class in §B.1 is detected by a mechanical predicate,
each trip yields exactly one attributable escalation record naming the line that tripped, and
no loss of detection on an armed card happens silently. Under the distributed default
`autonomy.mode: guided`, nothing observable changes.

## §D — Requirements (GEARS)

23 requirements, numbered in document order. Requirements whose predicate reads a
`contract.yaml` field, an A1 verify result, or an A1-owned configuration key carry the tag
「A1 plan-audit 통과본으로 재확인」; every such field is listed once, in §F, against the A1 final
at `25283ebf8` (v0.5.2).

Definitions used below. The **worktree root** is the top-level directory of the git worktree
containing the tool call's working directory. The **card id** is the base name of that worktree
directory (`kanban-dispatch.md`: the worktree directory keeps the card id). The **resolved
contract** is the one `contract.yaml` REQ-AE-002 selects. An **operation** is one write-capable
or Bash tool call observed at PostToolUse; a **turn** is one Stop hook event; an **audit retry**
is one audit verdict file for the card beyond the first. The **contract store** is the
moai-owned directory `$MOAI_HOME/db/<project-key>/contract/`, outside the repository and every
worktree, with `$MOAI_HOME` and `<project-key>` resolved exactly as the queue database resolves
them (research.md P17). The **card state file** is `<contract store>/escalation/<card-id>.json`
and the **card audit log** is `<contract store>/escalation/<card-id>.log.jsonl`, one per card; in
both names `<card-id>` is the worktree directory base name, byte for byte, with no case folding or
escaping (design.md §C.6, §C.11). The card audit log is the authoritative record of arming: a
card is **armed** exactly while its own log's most recent `armed` entry has no later `disarmed`
entry. The card state file is a cache of the arming snapshot and counters; its content never
arms or disarms a card by itself.

### D.1 — Activation, contract resolution, and effects

- **REQ-AE-001** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the escalation detector shall perform no detection, write no
  escalation record, and add no subprocess to any hook invocation, so that hook output is
  byte-identical to the pre-change behavior. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-002** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`,
  when a hook or checkpoint fires, the escalation detector shall first run the disarm check of
  REQ-AE-017 and then resolve the contract inside one resolver: take the card id from the
  worktree directory name, read every `.moai/specs/*/contract.yaml` under the worktree root, and
  keep those whose `card` field equals the card id and whose A1-derived `terminal` is false.
  Exactly one surviving contract is the candidate for arming (REQ-AE-023); zero appends one
  `not-armed` line, and two or more append one `not-armed` line plus one warning line naming
  every candidate. Neither the branch name nor the queue `spec_id` shall be an input.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-003** (Ubiquitous) — The escalation detector shall never deny, ask about, or
  alter a tool call; its only effects shall be the escalation records (which carry the
  needs-decision marking), the card audit log, and the card state file.
- **REQ-AE-004** (Event-detected) — When the escalation detector faults internally (a
  panic, an unreadable input, a timeout), the detector shall let the tool call proceed,
  shall append a `not-checked` line naming the class and the fault, and shall not write an
  escalation record for that fault.

### D.2 — The six contract classes

- **REQ-AE-005** (Event-driven) — When A1 verify of the resolved contract reports any of the five
  acceptance reasons — `acceptance_hash_mismatch`, `ac_count_mismatch`, `ac_count_ambiguous`,
  `acceptance_missing`, `signature_acceptance_mismatch` — the escalation detector shall trip class
  `acceptance-change`, citing the recorded `acceptance.sha256` / `acceptance.ac_count` beside the
  measured values. An `acceptance.md` edit after signing is a contract violation: the same
  observation also voids the signature, so REQ-AE-017 writes its disarm record as well. An edit
  made during plan phase, before the contract carries a signature, arms nothing and writes no
  record. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-006** (Event-driven) — When a Bash tool call whose command string equals a
  command-kind entry of the contract's `invariants` (an entry that is neither a
  `constitution:` glob nor `frozen-files`) completes with a non-zero exit status, the
  escalation detector shall trip class `invariant-violation` (sub-kind `command`).
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-007** (Capability gate + event) — Where the contract's `invariants` contains
  `frozen-files`, when a write-capable tool call targets a path matching any glob of the A1-derived
  `frozen_files` list (registry Frozen-zone target files ∪ `**/CLAUDE.md` / `**/CLAUDE.local.md` ∪
  `ownership.never`, obtained from verify at arming), the escalation detector shall trip class
  `invariant-violation` (sub-kind `frozen-file`) regardless of the calling agent's identity.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-008** (Event-driven) — When a write-capable tool call targets a path inside the
  worktree root that matches no `ownership.write` pattern, or matches any `ownership.never`
  pattern, and is not exempt under REQ-AE-013, the escalation detector shall trip class
  `ownership-move`, with a `never` match taking precedence in the record.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-009** (Capability gate + event) — Where `workflow.autonomy.escalation.new_api_detector`
  is `graph`, when the on-demand checkpoint compares the card base to HEAD and observes a new
  package directory, a new exported declaration, a new CLI verb, a new MCP tool name, or a new
  configuration key, the escalation detector shall trip class `new-architecture-or-api` and
  list every observed addition by kind and path. The **card base** is the merge-base of HEAD
  with the integration branch the project's git strategy configuration names (the development
  branch where one is configured, otherwise the default branch); where no such branch resolves
  or the merge-base is empty, class 4 is not-observed (REQ-AE-022). Sub-kind coverage is per
  language: new package and new exported declaration apply to every language the declaration
  extractor supports; new CLI verb, new MCP tool name, and new configuration key apply only to
  languages for which the detector carries a registration recognizer, and every other language
  reports those three sub-kinds as not-observed.
- **REQ-AE-010** (Event-driven) — When a recorded verdict pair for the card disagrees —
  an `audit_multi` result whose `disagreement_flag` is present and true, a verdict from the
  contract's `review.second_model` opposite to the first verdict, or a recorded CI failure for
  a head whose local verification was recorded as passing — the escalation detector shall trip
  class `contradictory-evidence`. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-011** (Event-driven) — When a Bash tool call is a push, a tag creation or tag push,
  a release creation, a force push, or a match for the existing destructive-command denylist,
  and it is not a push of `develop` authorized by the `push-develop` token in the contract's
  `actions`, the escalation detector shall trip class `irreversible-action` before the command
  executes, whether or not the existing denylist then denies it. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-012** (State-driven) — While the card is armed, when a write-capable tool call
  targets a path in the A1-derived `effective_never` list — which adds the SPEC's `contract.yaml`
  and `acceptance.md` once signed — the escalation detector shall trip class `ownership-move` even
  where `ownership.write` covers the path. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-013** (Ubiquitous) — The escalation detector shall judge a write-capable tool call
  targeting the contract store first, before any exemption below is applied: it trips
  `ownership-move`, is never exempt — not even when the store lies under the operating system's
  temporary directory, the session scratchpad, or another exemption root — and is never listed as
  not-observed, because the store path is always determined by its own resolution rule (and an
  unresolvable store is a REQ-AE-004 fault, design.md §C.6). The detector's own writes there are
  not tool calls and are never judged. Every other write is then exempt from class 3 when it
  falls under `.moai/reports/<card-id>/`, under `.moai/state/`, under the operating system's
  temporary directory, under the session scratchpad, under the auto-memory store, or under any
  glob in the A1-derived `scratch` list; a write outside the worktree root that none of these
  covers shall trip `ownership-move` when every exemption root was determined, and shall be
  listed as not-observed when one or more roots could not be determined (design.md §C.9). 「A1 plan-audit 통과본으로 재확인」

### D.3 — Operational trips

- **REQ-AE-014** (Event-driven) — When the card's observed turn count, operation count, or
  audit-retry count exceeds `budget.turns` / `budget.operations` / `budget.audit_retries` of the
  resolved contract — or, once the card has been disarmed, of the budget the card state file
  recorded at arming — or, where neither exists, the `workflow.autonomy.escalation.budget_default`
  value, the escalation detector shall trip class `budget-exceeded` naming the exceeded dimension.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-015** (Event-driven) — When the same failure fingerprint (normalized failing
  command plus diagnostic key) is observed on three consecutive failing attempts with no
  intervening success of that command, the escalation detector shall trip class
  `same-diagnostic-repeat`.
- **REQ-AE-016** (Event-driven) — When a plan-audit or a sync-audit verdict file for the card
  records FAIL at iteration `budget.audit_retries + 1` of that audit, the escalation detector
  shall trip class `audit-fail-at-retry-cap`; no separate configuration key supplies this
  ceiling. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-017** (Event-detected) — When the card's own audit log shows the card armed and a hook
  or checkpoint — before REQ-AE-002 resolution can report not-armed — observes any disarm reason,
  the escalation detector shall write exactly one `detection-disarmed` record for that arming,
  naming the reason, append a `disarmed` entry to that card's audit log, and keep classes 7-9
  active. A card counts as disarmed only once its log carries that `disarmed` entry; a card state
  file that has lost its arming value, or whose bytes differ from the digest in the log's latest
  state entry, while the log still shows the card armed is judged `state-tamper`, never read as a
  disarm. The disarm reasons are: `contract-absent` (the armed contract file is gone),
  `signature-invalid` (A1 verify no longer reports `signed-valid`, including a removed signature
  block and `signature_acceptance_mismatch`), `terminal-status` (the SPEC's A1-derived `terminal`
  became true, including a completion at sync), `card-mismatch` (the contract's `card` field no
  longer equals the card id, or resolution no longer yields exactly one contract), and
  `state-tamper` (the card state file is missing or lacks its arming value while the log shows the
  card armed). Independently of whether the log shows the card armed, the detector shall judge
  `state-tamper` when the card log's hash chain is broken, and whenever the card state file exists
  but does not match the digest in the log's latest state entry. When the card audit log does not
  show the card armed — it is absent, empty, or ends before an arming that other evidence records
  — the detector shall read the card as never armed unless evidence of arming exists that the log
  does not account for (design.md §C.6 step 3: a card state file naming an arming the log has no
  `armed` entry for, a `detection-disarmed` record whose fingerprint no `disarmed` entry carries,
  or a `contract`-kind record while the log holds no `armed` entry), in which case it shall write
  the `state-tamper` record and append a `disarmed` entry, starting a new log if none exists.
  The detector shall hold a per-card exclusive lock while it reads or writes that card's state
  file and log, so concurrent hooks of one card append in sequence; failing to take the lock is a
  REQ-AE-004 fault, never a tamper. Entries in another card's log shall never affect this card. A second disarm reason observed after the first
  increments that record's `occurrences` and writes nothing new.
  「A1 plan-audit 통과본으로 재확인」

### D.4 — Reporting and marking

- **REQ-AE-018** (Ubiquitous) — The escalation detector shall write each escalation record as
  one Markdown file at `<worktree root>/.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`
  whose YAML frontmatter conforms to the schema in §I, shall treat a card as needs-decision exactly
  when at least one of its records of kind `contract` or `operational` has `status: open`, and
  shall not add a state to, rewrite, reorder, or drop any queue item.
- **REQ-AE-019** (Event-driven) — When any class trips, the escalation record shall carry the
  class, the observation (event or command plus verbatim evidence), at least two options for the
  decider, the tripped line in the `contract_ref` form §I.1 defines for its class, the SPEC ID,
  the card id, and the HEAD commit at detection. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-020** (State-driven) — While an escalation record for the same class and fingerprint
  exists with `status: open`, the escalation detector shall not write a second record for that
  trip and shall increment the existing record's `occurrences`; when the existing record has
  `status: resolved`, a new trip shall be written as `<class>-<fingerprint>-<n>.md` with the next
  free ordinal, leaving the resolved record and its `decider` unchanged.
- **REQ-AE-021** (Ubiquitous) — The escalation detector shall present an escalation as a
  record, shall not route it through the user question channel, and shall not block, pause,
  or require acknowledgement before the triggering agent's next tool call.
- **REQ-AE-022** (Ubiquitous) — The escalation detector shall label every detection it could not
  complete (an unavailable CI verdict, an unsupported language for declaration extraction or
  registration recognition, an unresolvable card base, an unreadable contract field, an
  undetermined exemption root, a `constitution:` invariant, which has no mechanical violation
  signal, a command-kind invariant that no tool call executed since the previous checkpoint) as
  not-observed, and shall never render not-observed as agreement or as absence of a trip.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-023** (Event-driven) — When REQ-AE-002 yields exactly one candidate for a card that is
  not armed, the escalation detector shall run A1 verify on it on that same hook call — PreToolUse
  included — and arm the card only on `signed-valid`, recording in the card state file the SPEC
  ID, contract path, `signature.contract_sha256`, card id, the derived `frozen_files`,
  `effective_never` and `scratch`, and the budget; on `unsigned` it shall append a `not-armed`
  line, and on `signed-invalid` a `not-armed` line and a warning line carrying the verify reason
  codes, keeping operational classes 7-9 active against
  `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」

## §E — Constraints

- **C1 — Default preserves today.** `workflow.autonomy.mode` and
  `workflow.autonomy.escalation.budget_default` belong to A1's configuration (A1 § Configuration:
  default `guided`, invalid value → `guided` plus a warning); this SPEC reads them and adds only
  `workflow.autonomy.escalation.new_api_detector`. No behavior, file, or hook output changes under
  `guided` (REQ-AE-001). `MOAI_AUTONOMY_TIER` (`internal/config/envkeys.go:148`) keeps its meaning.
- **C2 — Name distinct from the harness escalation block.** `harness.escalation`
  (`internal/config/types.go:1199-1243`) names harness-level escalation; the keys read here live
  under `workflow.autonomy.escalation` and the two are never merged.
- **C3 — Detection only.** This SPEC adds no denial of any kind. No existing gate is rewired,
  weakened, or strengthened.
- **C4 — Hook budget.** Under `contract` mode the PreToolUse path performs: reading each
  `.moai/specs/*/contract.yaml` under the worktree root for its `card` field; reading the
  frontmatter `status` of each candidate SPEC's `spec.md` for A1's `terminal`; reading the card
  state file and checking its digest against the latest state entry in the card's own audit log
  (reading that card's log, which no other card writes, and appending to it under a per-card
  exclusive lock); path and glob
  matching against the
  cached derived lists; reading HEAD from the repository's HEAD and ref files. It calls A1 verify
  in-process — A1 § Non-Functional Constraints states verify completes without subprocesses or
  network access so that A2 can call it from a PreToolUse hook — at first arming (REQ-AE-023)
  and again whenever the contract file's content digest differs from the one cached in the card
  state file; otherwise it reuses the cached verify result. No subprocess runs on that path. The
  commit checkpoint inside a PostToolUse hook runs A1 verify afresh — so an `acceptance.md` change
  that leaves the contract bytes unchanged is seen there — and adds verdict-file reads, within the
  hook's configured timeout; a timeout is a REQ-AE-004 fault. Class 4 runs only at the on-demand checkpoint (open
  question Q2), never inside a hook.
- **C5 — Template neutrality.** Any template content added carries no card id, SPEC id,
  internal date, or commit SHA and favors no programming language.
- **C6 — Heuristic honesty.** Class 4 is a heuristic; its misses are covered by sync-audit and
  second review, not by this SPEC.

## §F — Dependency on A1

This SPEC does not design the contract schema and does not copy it. **Source of every field
below:** SPEC-AUTONOMY-CONTRACT-001 v0.5.2 `design.md` § Contract Schema and `spec.md`. **Pin:
`25283ebf8`** — the single commit every A1 citation in this SPEC refers to (branch
`WT-contract-schema`, unpushed at the time of pinning; read with
`git show 25283ebf8:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md` and the same for
`spec.md`). The earlier pins `8f77d9a33`, `4208a3a3b`, `6d98ca466` (v0.4.1), `67a2f55cb` (v0.5.0),
and `65e0a9167` (v0.5.1) are superseded. Every row of §F.1 is present at `25283ebf8`; the closed
reason set there
adds `card_invalid`, which REQ-AE-017 absorbs as `signature-invalid`. This SPEC
defines no decider rule: the decider value follows the A1 schema, and judgment rules
(agreement, fallback, disagreement handling) are owned by A3 (card t1236), per lead ruling
09-26 (4) #1. Tags 「A1 plan-audit 통과본으로 재확인」 are
kept by lead instruction; every row of §F.1 is re-checked in run-phase pre-flight.

### F.1 Fields consumed (confirmed present at `25283ebf8`)

| A1 field / surface | Consumed by |
|---|---|
| `.moai/specs/<SPEC-ID>/contract.yaml` (file location) | REQ-AE-002, REQ-AE-017 |
| `card` field (A1 `design.md:22`, `:124`, § Card Field `:146-163`) | REQ-AE-002, REQ-AE-017 |
| verify `state` and `reasons[]` (closed set incl. `signature_acceptance_mismatch`, `signature_seal_mismatch`, `signature_inconsistent`) | REQ-AE-005, REQ-AE-017, REQ-AE-023 |
| `acceptance.sha256` / `ac_count` with `acceptance.measured_sha256` / `measured_ac_count` | REQ-AE-005 |
| `signature.contract_sha256` | REQ-AE-023 |
| derived `terminal` (quoted or unquoted `status`) | REQ-AE-002, REQ-AE-017 |
| derived `frozen_files` (glob list) | REQ-AE-007 |
| derived `effective_never` | REQ-AE-012 |
| derived `scratch` (`ownership.scratch`) | REQ-AE-013 |
| `invariants[]` — kinds `constitution:<glob>`, `frozen-files`, command | REQ-AE-006, REQ-AE-007, REQ-AE-022 |
| `ownership.write[]`, `ownership.never[]` | REQ-AE-008 |
| `review.second_model` | REQ-AE-010 |
| `actions[]`; `push-develop` token | REQ-AE-011 |
| `budget.turns`, `budget.operations`, `budget.audit_retries` (shared by plan-audit and sync-audit) | REQ-AE-014, REQ-AE-016, REQ-AE-023 |
| `escalate_on[]` six tokens | REQ-AE-019 |
| `workflow.autonomy.mode`, `workflow.autonomy.escalation.budget_default` | REQ-AE-001, REQ-AE-014, REQ-AE-023 |
| verify callable from PreToolUse (A1 § Non-Functional Constraints) | REQ-AE-023, C4 |

### F.2 Open items and requests to A1

- **O1** `escalate_on` is fixed at six tokens. **O2** `constitution:<glob>` has no mechanical
  violation signal (not-observed). **O3** push authority reads `push-develop` only. **O4** `budget`
  is filled at sign time. **O5** `mode` and `budget_default` are A1's keys; this SPEC adds only
  `new_api_detector`. **O6** verify output carries no line numbers; `contract_ref` lines are mapped
  from the file text. **O7** class 1 consumes A1 verify's measured values. **O10** command-kind
  invariants are observed, never executed; unexecuted ones are not-observed.
- **R1-R4, R7** — satisfied at `25283ebf8`: `effective_never` (R1), `ownership.scratch` (R2), shared
  `audit_retries` (R3), `frozen_files` union with `**/` basename globs (R4), derived `terminal`
  excluded from resolution (R7). The registry Frozen list inside `frozen_files` is supplied by the
  caller through the constitution registry loader (A1 `design.md:200`, § Frozen Files item 1); the detector obtains it
  once at arming and caches it (design.md §C.6).
- **R8 — Mission-projection owner label — CLOSED at `25283ebf8`.** A1 `spec.md:79` reads "Out of
  Scope — Mission-validator projection (A2b, card t1245)", `spec.md:179` reads "The
  mission-validator projection is deferred to A2b (t1245).", `design.md:241` names "A2b (t1245)"
  as the owner of the projection decision, and `design.md:470` reads "Forward Note — Mission
  Projection (for A2b, t1245)".
- **R9 — `card` field inside the signed digest — CLOSED at `25283ebf8`.** A1 `design.md:22`
  declares `card` REQUIRED and covered by the digest; `design.md:124` fixes the pattern
  `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$` with reason code `card_invalid`; § Card Field
  (`design.md:146-163`) states that the resolver arms only when `card` equals the worktree
  directory name and that editing `card` after signing changes the digest; A1 `spec.md:262-265`
  carries the same requirement.
- **R10 — Path of the moai-owned store — CLOSED (lead ruling 09-26 (4) #2; A1 `25283ebf8`).** The
  store is `$MOAI_HOME/db/<project-key>/contract/` (A1 `design.md:89`; `spec.md:95-97`, `:196`,
  `:205`), in the same home layout as the queue database. The card state file, the disarm state,
  and the per-card audit logs (each with its own hash chain) live there; A3's signing-event store uses the
  same place.

Also depends on A1 landing on `develop` before run-phase begins (card text: "plan 은 병행 가능,
run 은 A1 develop 병합 뒤").

## §G — Out of Scope

### Out of Scope — rewiring or adding gates

- Turning any escalation trip into a deny, an ask, or a block (that is track A3).
- Adding a plain-push-to-`main` or `git tag` deny rule; class 6 reports the attempt only.
- Push serialization and the contract-sign deny (moved to card t1245, §K).

### Out of Scope — the closure report

- Aggregating escalation records into a closure or completion report (track A4), including
  resolving the `terminal-status` disarm record that a normal sync completion writes.

### Out of Scope — the contract schema and revocation

- Defining, validating, signing, or versioning `contract.yaml` (track A1, card t1234); §F.2
  lists requests only.
- Detecting violation of a `constitution:<glob>` invariant; it is reported as not-observed.
- Implementing `moai contract revoke` or writing revoke records; §I only reserves their kind.

### Out of Scope — queue state and the F1 record layer

- Adding a `needs-decision` value to the queue state enum, or any queue column.
- Keeping the queue `spec_id` filled for picked cards (F1's concern); the resolver does not read it.
- Automatically resolving, dismissing, or acting on an escalation.
- Migrating escalation records into the factory worktree↔card record (F1 ingests the files later).

### Out of Scope — mission-validator projection

- Projecting the contract onto `mission.MissionContract` and reusing `ValidateMissionDecision`
  (moved to card t1245, §K).

### Out of Scope — tamper through Bash

- Local single-user: the goal is tamper evidence, not tamper prevention. Class 3 sees only
  write-capable tool calls, so a Bash edit is seen, if at all, by the digest check at the next hook
  (REQ-AE-017), never at the moment of the write.
- The hash chain is unkeyed and backward-linked, so it is evidence only against naive edits: it
  catches an edit to any line that has a successor; it does not catch truncation of trailing
  lines, an edit to the last line, or a Bash edit that rewrites the card state file and appends a
  recomputed chain line to the card audit log.
- Truncation and file removal are bypass routes. Deleting, emptying, or tail-truncating the card
  audit log is caught only while evidence of arming the log no longer accounts for survives
  (REQ-AE-017); removing or emptying the card state file, the card audit log, and any
  `contract`-kind or `detection-disarmed` record of the card via Bash leaves no trace: the card
  reads as never armed, and the next arming restarts the class 7 budget counters from zero.

### Out of Scope — reviving the ac-baseline guard as a product feature

- Mirroring `scripts/ac-baseline/` into the template or wiring it into this detector.

### Out of Scope — a CI poller

- Fetching CI results from a remote. Class 5 consumes a recorded CI verdict when one exists and
  reports not-observed otherwise.

## §H — Lead rulings

Each ruling is recorded with the reason it was needed; the reasons are the plan-audit findings it
closes (`.moai/reports/t1235/plan-audit-iter1.md`, `-iter2.md`, `-iter3.md`).

| Ruling | Decision | Reason | Carried by |
|---|---|---|---|
| lead ruling 09-26 #1a | After signing, `contract.yaml` and `acceptance.md` are implicitly `never`. | D2: editing the contract could turn detection off with no record. | REQ-AE-012 |
| lead ruling 09-26 #1b | Contract loss is never silently disarmed; one escalation through the operational path. | D2 / Q6. Generalized by (3) #2. | REQ-AE-017 |
| lead ruling 09-26 #2 | Resolve from the worktree root, never the branch name; log not-armed outcomes; one resolver function. | D1. Resolution input replaced by (3) #1. | REQ-AE-002 |
| lead ruling 09-26 #3 | Fixed default exemptions plus optional `ownership.scratch`. | D3. The (3) #4 carve-out is superseded by (4) #2 (state moved out of the worktree). | REQ-AE-013 |
| lead ruling 09-26 #4 | Needs-decision lives in an escalation record beside the queue. | Q1. Path and format superseded by (2) #6. | REQ-AE-018 |
| lead ruling 09-26 #5 | No new audit-ceiling key; `budget.audit_retries` for both audits. | Q4. | REQ-AE-016 |
| lead ruling 09-26 #6 | `frozen-files` = registry Frozen targets ∪ instruction files ∪ `ownership.never`. | Q7. Now read from A1's derived `frozen_files`. | REQ-AE-007 |
| lead ruling 09-26 (2) #1 | Resolver narrowed by card id via the queue `spec_id`, then by SPEC status. | N1, N3. **Superseded by (3) #1.** | — |
| lead ruling 09-26 (2) #3 | Split: push serializer, sign deny, their criteria, and the projection move to card t1245. | N5, N7, Tier L ceilings. | §K |
| lead ruling 09-26 (2) #4 | The disarm check runs before resolution, from the card state file. | N2. Kept; broadened by (3) #2. | REQ-AE-017 |
| lead ruling 09-26 (2) #6 | Record path `<class>-<fingerprint>.md` with YAML frontmatter; revoke kinds for A3. | Stable, parseable surface for A3 and F1. | REQ-AE-018, §I |
| lead ruling 09-26 (3) #1 | The worktree directory name is the card id; arm only when exactly one contract under the worktree has `card:` equal to it and a non-terminal SPEC; do not use the queue `spec_id`; ask A1 for a `card` field inside the signed digest (R9). | P1: the queue `spec_id` was filled on 0 of 119 live cards, so the (2) #1 resolver left nearly every card not-armed. `spec_id` hygiene belongs to F1. | REQ-AE-002, R9 |
| lead ruling 09-26 (3) #2 | One unified disarm rule: an armed card that becomes unarmed for any reason — contract deleted, signature invalid, SPEC status changed, card mismatch, state-file tamper — writes exactly one record per card. | P1, P2, P4: resolver transitions (terminal status, card change) and state-file edits disarmed with only a log line. | REQ-AE-017 |
| lead ruling 09-26 (3) #3 | Allow a read-only `moai contract verify` call at PreToolUse, consistent with A1 § Non-Functional Constraints. | P3: first observation needed verify but C4 forbade it on PreToolUse. | REQ-AE-023, C4 |
| lead ruling 09-26 (3) #4 | Move the prior-state file into the moai-owned store if A1/A3 names one; otherwise remove its path from the exemption list. | P4: the state file sat inside the `.moai/state/` exemption, so an agent could rewrite it unrecorded. No store path is named at `67a2f55cb`, so the path is carved out and R10 asks for the store. | REQ-AE-013, REQ-AE-017, R10 — superseded by (4) #2 |
| lead ruling 09-26 (3) #5 | Apply B1-B6: re-pin §F to `6d98ca466`; an `acceptance.md` edit during run is a contract violation, so the void is correct, distinct from plan-phase edits before signing; `terminal` per A1; projection owner request; verify at PreToolUse; `frozen_files` from A1. | B1-B6 of `plan-audit-iter3.md`. | REQ-AE-005, 007, 012, §F |
| lead instruction 09-26 (A1 re-pin) | Pin §F to A1 v0.5.0 at `67a2f55cb` instead of `6d98ca466`; re-check only the decider and agreement/disagreement surfaces. | A1 changed its kickoff decider section. The decider text this row added is withdrawn by (4) #1. | §F |
| lead ruling 09-26 (4) #1 | Operator's final decision: the decider value set is {`human`, `llm`, `llm+jev`}; judgment rules (agreement, fallback, disagreement handling) belong to A3 (card t1236), and A1 v0.5.1 is schema only. This SPEC defines no decider rule and says only that the decider value follows the A1 schema; the record's `decider` field is preserved verbatim and not validated against a value list. | Decider rules are A3's; restating any of them here would pin a rule this SPEC does not own. | §F, §I.1 `decider`, REQ-AE-010, and the criterion numbered AC-AE-012 since v0.4.2 |
| lead instruction 09-26 (plan-audit iteration 4) | Fix Q3, Q4, Q5 and m2: judge contract-store writes before every class 3 exemption; state the unkeyed-chain and deletion limits plainly; re-pin §F and close R8, R9, R10 citing A1 lines — first to v0.5.1 `65e0a9167`, then by a follow-up instruction to v0.5.2 `25283ebf8` (single pin), which also resolves the two stale A2 sentences; add the `spec.md` `status` read to C4. | `plan-audit-iter4.md` Q3 (store under a temp root was exempt), Q4 (residual risk overstated the chain), Q5 (A1 v0.5.1 landed after v0.4.1). | REQ-AE-013, §G, §F, C4, AC-AE-010 |
| lead instruction 09-26 (plan-audit iteration 5) | Close R1 and R2 before Kickoff, plus n1 and n2: a per-card exclusive lock around every state write and log append; the missing-log reading widened to "the log does not show armed" (absent, empty, tail-truncated) against unaccounted evidence; the state-file digest always compared; a broken chain is `state-tamper` whatever the arming status; §G corrected to what the chain catches; the card-id collision cases (L1/L2 parents, case-insensitive store) named as residual risk; (4) #2 marked superseded. R1 and R2 are derived from the design text; both are first observed in run milestone M5 (AC-AE-019 concurrent clause; AC-AE-018 (d), AC-AE-020 (d)). | `plan-audit-iter5.md` R1 (same-card concurrent appends could break the chain and read as tamper), R2 (an emptied or tail-truncated log disarmed silently), n1, n2. | REQ-AE-017, §G, C4, design.md §C.6/§C.11, AC-AE-018/019/020 |
| lead ruling 09-26 (5) | Option (a) for plan-audit iteration-4 Q1/Q2: one audit log per card inside the contract store, named `<card-id>.log.jsonl` from the worktree directory name byte for byte; the card's own log entries, not the state file, decide whether the card is armed. A disarm counts only when the log carries a disarm entry; a state file that lost its arming value while the log says armed is `state-tamper`; another card's appends never touch this card's chain; a missing log reads as never armed unless prior evidence of arming exists, in which case it is `state-tamper`. | Q1: the trigger read the state file, so emptying it disarmed silently instead of tripping `state-tamper`. Q2: one shared chain across cards and lanes let a legitimate append by another card break this card's check. | REQ-AE-017, definitions, design.md §C.6/§C.11, AC-AE-018/019/020 |
| lead ruling 09-26 (4) #2 — superseded by lead ruling 09-26 (5) and the iteration-4 lead instruction | R10 closed: the moai-owned store is `$MOAI_HOME/db/<project-key>/contract/`, outside the repository and every worktree, in the queue database's home layout. The card state file and the detector audit log with its hash chain move there; A3's signing-event store uses the same place. The store is not a project path: a tool-call write there is judged by the outside-root rule. | Supersedes the `.moai/state/escalation/` carve-out of (3) #4. | REQ-AE-013, REQ-AE-017, §F.2 R10, §G |

## §I — Escalation record format

One file per class and fingerprint:
`<worktree root>/.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`, with `-<n>`
appended for a re-trip after resolution (REQ-AE-020). `<fingerprint>` is the first 16 lowercase
hex characters of the SHA-256 of the class plus the class's normalized observation (design.md
§C.10). The directory is gitignored and lives in the card worktree, so the lead reads it — and F1
ingests it — before the worktree is disposed.

### I.1 Frontmatter (machine-readable, YAML)

| Field | Type | Rule |
|---|---|---|
| `schema_version` | int | `1` |
| `card` | string | card id (worktree directory name) |
| `spec` | string | the resolved or last-armed SPEC ID, or `""` when none |
| `kind` | string | `contract`, `operational`, or `revoke` |
| `class` | string | a §B.1 class name for `contract` / `operational`; a revoke class (I.2) for `revoke` |
| `fingerprint` | string | 16 lowercase hex characters, equal to the file name's fingerprint |
| `contract_ref` | string | contract classes: `contract.yaml:<line>`; class 7: `config:<key>` or the contract budget line; class 8: `rule:same-diagnostic-3`; class 9: `contract.yaml:<line>` of `budget.audit_retries`; class 10: `disarm:<reason>` with `<reason>` from REQ-AE-017; revoke records: `""` when no line |
| `escalate_on` | string | the matching token for contract classes, otherwise `""` |
| `status` | string | `open` or `resolved` |
| `decider` | string | `""` while open; set by whoever resolves the record. The value follows the A1 schema; judgment rules are owned by A3. The detector never writes it, preserves it verbatim (REQ-AE-020), and does not validate it against a value list |
| `occurrences` | int | ≥ 1 |
| `head_sha` | string | HEAD at first detection |
| `detected_at` | string | UTC RFC 3339, first detection |
| `updated_at` | string | UTC RFC 3339, last change |
| `not_observed` | string list | every detection this record could not complete (REQ-AE-022) |

The body carries three Markdown sections in order: `## Observation` (event or command plus
verbatim evidence), `## Options` (at least two numbered options), `## Not observed`.

### I.2 Revoke kinds (reserved for A3)

`kind: revoke` records document a `moai contract revoke` (track A3). This SPEC reserves the kind
and two class names and never writes them:

| Class | Meaning |
|---|---|
| `revoke-operator` | a person revoked the contract directly |
| `revoke-on-decision` | a decider revoked the contract while resolving an open escalation record |

A revoke record carries `status: resolved` and a non-empty `decider`; it never makes a card
needs-decision (REQ-AE-018). A3 may add revoke classes; this table is the reserved minimum.

A card is needs-decision exactly when one of its `contract` or `operational` records has
`status: open`.

## §K — Moved to A2b (card t1245)

Lead ruling 09-26 (2) #3 moved the following out of this SPEC to card **t1245** (A2b). They are
not requirements of this SPEC any more; numbers below are the v0.2.1 identifiers.

| What it was | Former id (v0.2.1) |
|---|---|
| Push serializer on the `moai slot` `push-develop` lease (A3 precondition) | REQ-AE-024 |
| PreToolUse deny on agent-invoked `moai contract sign`, with the receipt-path allowance (A3 precondition) | REQ-AE-025 |
| Push serializer criteria (second push denied / first unaffected; release and stale reclaim) | v0.2.1 AC-AE-022, AC-AE-023 (numbers since reused) |
| Contract-sign guard criteria (bypass shapes and controls; receipt path allowed) | v0.2.1 AC-AE-024, AC-AE-025 (numbers since reused) |
| Their mechanics, milestone, and bypass-shape table | design.md §G, plan.md M6, spec.md §J |
| Receipt path; which window `push_requires_window` means | A1 requests R5, R6 |
| Projecting the contract onto `mission.MissionContract` and reusing `ValidateMissionDecision` (`internal/mission/policy.go:200`) | O9 / N7 — mission-validator projection (A1 relabel requested as R8) |

Because of the split, plan-audit iteration-2 findings **N4** and **N8** no longer apply to this
SPEC; they travel with the sign deny to card t1245, as do N5, N6, and N7.
