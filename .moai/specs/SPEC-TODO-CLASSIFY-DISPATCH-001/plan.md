# SPEC-TODO-CLASSIFY-DISPATCH-001 — Implementation Plan

Tier M · card t1332 · plan-phase authored 2026-09-29 at HEAD `145c3d98c` (worktree
`.moai/worktrees/t1332`, branch `WT-card-autodispatch`). Milestones are ordered by
decision-reversibility: the data-model change (M1) is the least reversible and leads; mechanical
doc/regression edits (M5) close. Priority labels only — no time estimates.

## §A Context

The queue carries no judgment metadata: `BacklogItem` (`internal/kanban/backlog_store.go:83`) is
five frozen fields plus additive stamps, order is insertion order, and every machine selector —
`factory next`'s auto-promotion arm above all (`WT-factory-self-dispatch:internal/cli/factory_card.go`
arm (c)) — reads insertion order as priority. This SPEC adds creation-time classification
(priority / blocked / execution mode), a priority-sorted queue, mode-aware lane leases, and a
default-on auto-dispatch for `-f` lanes. The t1240 self-dispatch surface is the contract this plan
targets and is NOT on develop yet — its absorption is an entry precondition (§D.1).

## §B Known facts (measured in this tree; full detail: `.moai/reports/t1332/surface-notes.md`)

1. **Add path** — `newTodoAddCmd` (`todo.go:504`), `runTodoAddAppend` (`:534`) and
   `runTodoAddPick` (`:575`) both go through `appendAnalyzedCard` inside one locked `Mutate` —
   the classification write belongs inside that same closure (REQ-TCD-001).
2. **Additive per-item field precedent** — `Landing *LandingEvidence json:"landing,omitempty"`
   (`backlog_store.go:92-100`), `PickedAt`/`DroppedAt` (`:104-110`); SQLite side:
   `backlogLandingColumn` TEXT appended by `ensureLandingColumn`
   (`backlog_sqlite.go:108-129, 480-550`). The classification field follows exactly this shape
   (one nullable struct, one TEXT JSON column).
3. **No priority/mode today** — the only `Priority` hit in `internal/kanban` is a comment
   (`board_store.go:87`); queue order is slice order.
4. **t1240 surface** (branch `WT-factory-self-dispatch`, tip `d43e50bb3`, 28 commits ahead of
   develop at plan time; `9866ca25e` is an ancestor): `factory next` selection is
   (a) assigned-to-this-lane → (b) picked-no-owner → (b2) queue-picked-no-record → (c) first
   `queued` item promoted picked-then-recorded; `factoryNextNoCardExit = 3`; duplicate prevention
   is the version-checked claim (`ErrStaleVersion`/`ErrLeaseHolder` → race retry).
5. **Lane bootstrap** — the SessionStart notice instructs the lane to run `moai factory next`
   (`WT-factory-self-dispatch:internal/hook/session_start_factory_i18n.go:99-100`); admission via
   `MOAI_FACTORY_ROLE=lane` (`internal/config/envkeys.go:325-347`). The `-f` launcher already
   carries a lane-only stamped flag precedent — `--clear-policy`
   (`WT-factory-self-dispatch:internal/cli/factory.go:64-67, 104, 205-214`; the flag is t1240-only
   and absent from this tree's develop) — the auto-dispatch opt-out flag follows it.
6. **Decider vocabulary** — `DeciderLLM = "llm"`, `DeciderLLMJev = "llm+jev"`
   (`internal/contract/seal.go:55-56`); "jev" is a named configuration error, not a member
   (`internal/config/autonomy_contract.go:29-35`); R5 refusal in `contract decide`
   (`internal/contract/kickoff/decide.go:111-112`). No Go `type Decider` interface exists yet —
   M2 defines the classification seam; the identity strings reuse the contract constants.
7. **Status view extension point** — `factoryCardView` (`internal/cli/factory_card.go:162-205`)
   already renders owner/lease/hints; mode+priority extend it (REQ-TCD-010).
8. **t1306 inherits the sort for free** — `/moai:todo --auto` consumes "in queue order"
   (`todo_auto.go:24`); once the queue is sorted, the serial cycle is priority-ordered with zero
   changes (constraint C2).

## §C Decision records (OD-1/OD-3 = leader rulings 2026-09-29, folded; OD-2 open for run Kickoff)

- **OD-1 — serial semantics = serial-card MUTUAL exclusivity (operator ruling 2026-09-29;
  supersedes this plan's 0.1.0 pipeline-exclusive choice).** A serial card in flight blocks only
  OTHER SERIAL cards — serial cards run one at a time, served in priority order, and the lane
  finding only serial candidates blocks on the existing no-card exit. Parallelizable cards remain
  pickable by other lanes even while a serial card is in flight. Rejected alternative (the 0.1.0
  choice, recorded): pipeline exclusivity — a full `next` refusal costs half the throughput
  benefit that classification exists to deliver (leader rationale). Provenance: leader Jev
  doctrine-fallback ruling, noul below threshold (OD-1 0.31), dated 2026-09-29.
- **OD-2 — `Decider(llm)` transport (RECOMMENDATION, run-phase detail).** The seam is the
  interface; the product LLM implementation's transport — shelling to the signing backend vs.
  accepting an agent-supplied judgement file as the primary path with the deterministic fallback
  for unattended adds — is chosen at run Kickoff. The judgement-file input (REQ-TCD-004) is
  mandatory either way: it is the only product surface that keeps an LLM out of the CLI's
  critical write path while still recording its judgment.
- **OD-3 — failure default = `serial` (operator ruling 2026-09-29; supersedes this plan's 0.1.0
  parallelizable choice).** Fail-safe: a parallelizable failure default could run true-serial
  cards concurrently and violate the ordering the mode exists to protect; the serial default
  costs throughput only. The 1-line stderr notice stays. Provenance: leader Jev doctrine-fallback
  ruling, noul below threshold (OD-3 0.36), dated 2026-09-29.
- **FLAGGED — failure-default vs read-default tension (lead follow-up; NOT silently resolved).**
  REQ-TCD-003's decider-failure default is now `serial` (write-time record), while REQ-TCD-014's
  absent-field read default stays `parallelizable` (read-time derivation for legacy/unrecorded
  cards): a card admitted during a decider outage runs serial, a card with no classification
  field at all reads parallelizable. Whether the read default should follow the fail-safe
  direction is the lead's follow-up question; this SPEC records the interaction and defers — the
  run phase must not resolve it by silently picking one side.

## §D Constraints

### D.1 Absorption order — run-phase entry precondition (REQ-TCD-013)

Run entry waits until **local develop carries the t1240 surface**: SPEC-FACTORY-SELF-DISPATCH-001
merged into develop, with `factory next` and `complete` present in develop's `internal/cli`
(mechanical check: `git grep -l "newFactoryNextCommand" develop -- internal/cli` non-empty).
Absorption source: branch `WT-factory-self-dispatch` (28 commits ahead of develop at plan time,
tip `d43e50bb3`). This SPEC's branch never merges that branch; the lead performs the develop
absorption through the serial integration window. M3 (and anything downstream of it) does not
start before the check passes.

### D.2 Hard boundaries

- t1240 scope (self-lease verbs, lease mechanics) — consumed, not re-planned (spec.md §E).
- t1306 scope (`/moai:todo --auto`) — untouched; it inherits the sort (constraint C2).
- t1261 scope (kickoff decider layer) — vocabulary reuse only.

### D.3 Jev boundary and the local dogfood seam (REQ-TCD-012)

`scripts/jev/` is never referenced from `internal/`, `pkg/`, `cmd/`, or
`internal/template/templates/`. The **exact dogfood seam**: the judgement-file input of
REQ-TCD-004 (`todo add --classification-file <path|->`). A local operator runs
`scripts/jev/triage.sh`-style tooling OUTSIDE the product to produce the classification JSON, then
invokes `moai todo add "<text>" --classification-file <that file>` — the product records a
validated, decider-attributed judgment it did not compute, and Jev stays a pre-add step in the
operator's hands. No Decider implementation in the product tree knows Jev exists.

## §E Self-verification (plan-phase obligations the run inherits)

- E1. Every REQ-TCD-XXX above maps to exactly one AC-TCD-XXX in `acceptance.md` (14/14).
- E2. Additive-marshal proof: a pre-SPEC queue fixture round-trips byte-identically (AC-TCD-002).
- E3. The `scripts/jev` boundary grep carries a positive control (AC-TCD-013 discipline — a
  0-hit grep without one is unmeasured, per the repo's positive-control rule).
- E4. Serial mutual-exclusivity and duplicate-dispatch ACs run as concurrent (goroutine) tests
  against
  the version-checked store, not as sequential approximations.
- E5. No config key, no template mirror, no `internal/jev` consumer is added (grep-negative
  assertions in M5).

## §F Milestones

- **M1 — Classification data model (P1, least reversible).** `BacklogItem.Classification`
  (pointer, omitempty; priority/blocked/mode/decider/classified_at/reason), SQLite
  `classification` TEXT column via the guarded ADD COLUMN path, read-side positive default
  derivation (REQ-TCD-002, REQ-TCD-014). Value-set types defined once, in `internal/kanban`.
- **M2 — Decider seam + add-path wiring + queue sort (P1).** The classification `Decider`
  abstraction (default + llm implementations), the `--classification-file` validated input, the
  fallback + stderr notice, and the sort re-established inside the add's locked write; position
  output and list/web order follow (REQ-TCD-001, -003, -004, -005, -006, -012).
- **M3 — Mode-aware factory dispatch (P1; entry gate §D.1).** `factory next` priority-order
  auto-promotion, blocked cards never auto-selected, serial-card mutual exclusivity (serial
  blocks only serial, served in priority order) with a positively-enumerated terminal set,
  parallelizable concurrency pin including while a serial card is in flight, `factoryCardView`
  mode+priority extension (REQ-TCD-007, -008, -009, -010). Depends on M1+M2 and the t1240
  absorption.
- **M4 — Auto-dispatch default-on (P2).** `-f` lane launcher default + `--no-auto-dispatch`
  opt-out stamped into the lane bootstrap, following the t1240 lane-only `--clear-policy`
  stamping precedent (`WT-factory-self-dispatch:internal/cli/factory.go:64-67, 104, 205-214`;
  absent from this tree's develop) (REQ-TCD-011).
- **M5 — Regression sweep + doc parity (P3, mechanical).** t1306/t1308 interaction regression
  tests, the jev-boundary grep with positive control, `moai todo` skill/command doc parity for
  the new flag and the sorted-order contract.

## §G Anti-patterns (named for the run to refuse)

- G1. A classification write OUTSIDE the add's locked `Mutate` (a two-step add+classify window a
  concurrent `factory next` could observe).
- G2. Negative-enumeration terminal-state checks (`if state != done && state != ...`) — the
  HOLD-STATE form defect; REQ-TCD-008 requires positive enumeration.
- G3. A fourth priority value or a third mode smuggled in as `""` — absence is the typed default
  derivation (REQ-TCD-014), never a hidden member.
- G4. Any `scripts/jev` import, exec, or path constant in product code; any Jev-named decider
  identity accepted by the classification validator.
- G5. Re-sorting on read (list/status) instead of on the locked write — two orderings that can
  disagree is the defect this SPEC exists to close.

## §H Cross-references

- spec.md §B (REQ-TCD-001..014) · acceptance.md (AC-TCD-001..014) · `.moai/reports/t1332/surface-notes.md`.
