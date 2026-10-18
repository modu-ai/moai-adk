# Plan: SPEC-CARD-DURABLE-IDENTITY-001

Implementation plan for `spec.md`. Sections A to F are ordered by decision reversibility: what is most likely to change comes first, mechanical routing last. No time estimates; ordering and priority only.

Base tree **B0** = `f4bcd4f2cc65342559b505f1e5a29df80f992894` (local `main` at the card worktree's creation; branch `WT-card-durable-identity`).

## A. Decision register (review order)

### A.1 Data model and identity (highest change risk)

- **One mutable row per card id, copies are sealed.** `card_current` keeps the legacy column names so the 15 statements that name `cards` change by table name only (§ I). Reversal cost: high after any deployment is cut over; the rollback verb is the exit.
- **Attempt = one lease grant or run engagement, sealed when it ends** (decision-index Q2, ruled: design principle 6). Every write that clears a holder, reassigns, moves run or re-picks a terminal card seals the pre-write record in its own transaction (REQ-CDI-023). Cost: one archive row per ended grant. Reversal cost: high once rows exist.
- **Project scope = empty `run_id` on a non-terminal record** (Q3). The alternative (force a destination at cutover) adds an input to every cutover.
- **Sealed legacy rows plus a retained `cards_pre_cutover`.** Two verbatim copies of each legacy row exist after the cutover by design: the immutable attempt (the history) and the retained table (the rollback and count reference).
- **Layout is the database's property** (Q1, Q4): marker on `meta.schema_version`, never a per-open read of the key.
- **Version and `attempt_no` never fall.** Receipts bind a version, and the compare-and-set on `version` is the only concurrency control the transitions have (`card_transition.go:882`).

### A.2 The seam and the published interface

The eight facts S-1 to S-8 of `spec.md` § E.2 are the interface; `design.md` § 8 lists what each lets stage A drop and what it must still carry. This plan makes three decisions about it:

1. The interface is **facts, not code**: this SPEC adds no exported symbol that stage A must call. The read functions of REQ-CDI-022 are named in design.md § 12 so that AC-CDI-022 can assert them, nothing more.
2. The epoch contract is tested **generically**: AC-CDI-013 creates a stand-in table shaped as the stage-A SPEC states its grant (card id primary key, `owner_epoch`), and also scans every table for the `owner_epoch` column, so the test does not depend on the name `card_grants`.
3. Stage A's re-plan is told, in `spec.md` § E.3 and `design.md` § 8.2, what the identity layout removes and what remains its own. Nothing in this SPEC edits stage A's artifacts.

### A.3 Operator surface and retirement semantics

- `moai factory identity report [--emit <path>]` (read-only), `identity cutover --resolutions <path>`, `identity rollback`; `moai factory reconcile [--card <id>|--all]`; `moai factory runs --retire <run> [--into <run>]`.
- Lane sessions are refused for `cutover` and `rollback` (the guard `factoryLaneRefusal` already refuses decisions in lanes).
- On a refused migration `relaunch --from-run` prints the left-active line and still launches (REQ-CDI-016). It passes its `--run` as the destination; its help text, which says it "does not migrate card leases or worktrees" (`factory_relaunch.go`, verb long description), is rewritten for the identity layout.
- Operator-facing text follows the neighbours (English, identifiers verbatim). No prompt is added; the verbs take everything from flags and files (the subagent boundary forbids prompts).
- Exit codes: `0` done; `1` refusal on a precondition (nothing written); `2` validation failure inside the transaction (rolled back). Output names the failing check by the ids V-1 to V-7 and R-1 to R-4 of design.md.

### A.4 The operation journal

- Four kinds, closed (S-8). Images are held in the journal, so the queue's table and the lock record do not change (Out of Scope).
- Decisions are a pure function of rows and readbacks, with no clock (REQ-CDI-019); the 14-cell table in design.md § 6.3 is the specification and the test list.
- Separable: the journal does not change the identity layout. If the plan-audit ceiling forces a split, the second part is M6 (cut line in § C).

### A.5 Rollout, rollback, compatibility

- This SPEC ships the key `false` in the repository and the template. No deployment is cut over by landing it.
- Operator sequence, documented in the verbs' help and the sync-phase CHANGELOG entry: set the key in the tree that runs the verb; stop every lane and leader process; `identity report --emit`; fill the blanks; `identity cutover`; restart every process that opened the factory before; `identity report` again.
- Rollback is `identity rollback`; a code revert across the cutover is not a rollback, and an older build cannot open the database (REQ-CDI-010).

## B. Run-entry precondition and re-verification

No prerequisite card has to land first. At run entry the run re-reads, on the tree it starts from, and records in `progress.md` § E.2:

- **B-1 coordinates.** `grep -n "PRIMARY KEY(run_id, card_id)" internal/homestate/factory.go` (line 77 at B0), `grep -rn "FROM cards\|INTO cards\|UPDATE cards\|JOIN cards" internal --include='*.go' --exclude='*_test.go'` (15 lines at B0, § I), `grep -n "const factorySchemaVersion" internal/homestate/factory.go` (6 at B0).
- **B-2 schema base.** If stage E's migration landed first the base is higher, and if it adds a `cards` column it follows design § 2.7 (the `cards` view cannot be altered; alter the three tables and rebuild the view and its triggers); every fixture is built from the run-entry tree's own schema (AC-CDI-006), and M1's helper keeps the marker beside whatever base the chain ends at.
- **B-3 the legacy retire guard.** If SPEC-FACTORY-RETIRE-GUARD-001 (card t1852) has landed, `RetireRunIfDead` carries its typed refusal; the identity branch is placed before it and the legacy branch keeps it. If it has not, the legacy branch is the B0 behavior. AC-CDI-002 holds in both orders.
- **B-4 stage A.** If a re-planned stage A has landed a grant table, AC-CDI-013 adds the real table as a second fixture.
- **B-5 the ceiling record.** `moai spec ceiling` counts rounds only from `plan-audit.md` and `plan-audit-iter<N>.md` in `.moai/reports/t1842/`; verdict files use those names.

## C. Milestones (execution order)

### M0 — Evidence first, in its own commit

The `TestCDI_*` family as failing tests, compiling against declarations that return `errNotImplemented` (no behavior), plus `.moai/reports/t1842/baseline.md`: the verbatim failing run of the family, the verbatim passing run of the existing named tests of AC-CDI-002, and the passing run of the control `TestCDI_LegacyRepickLosesStage_Control` (it characterizes today's defect: a re-pick in a new run starts at `picked`). The baseline file is force-added (the directory is ignored). Commit subject names `t1842`. Nothing else is in the commit. AC-CDI-024.

### M1 — Flag, layout detection, marker-aware open

`CardIdentityConfig`, the two `workflow.yaml` keys, the loader, `splitSchemaVersion` and the marker-preserving advance helper, the layout recorded on the handle. A legacy database is byte-identical after open (AC-CDI-002). Template-First: edit the template `workflow.yaml`, then `make build`, then the local copy.

### M2 — Report, resolution, cutover, validation, immutability, projection

`identity report`, the resolution input, `identity cutover` with C-1 to C-4, V-1 to V-7, the triggers, the `cards` view, the fault seams. AC-CDI-003 to AC-CDI-009 and the cutover half of AC-CDI-016.

### M3 — Identity-layout records

Statement routing, `ErrCardNotMember`, `ErrCardEngaged`, the seal writers of design § 4.4 (REQ-CDI-023, AC-CDI-025), join, re-pick, version and `attempt_no` monotonicity, cross-run readers. AC-CDI-011, AC-CDI-012, AC-CDI-013 (the join and seal halves), AC-CDI-016 (join half), AC-CDI-025.

### M4 — Retirement as migration

`retireRunWithMigration`, `--into`, `relaunch` destination, roster re-bind, dispatch re-point, refusals, the flag-set assertion. AC-CDI-014, AC-CDI-015.

### M5 — Rollback and the seam contract

`identity rollback` with R-1 to R-4, the generic epoch-carrier fixture across cutover, join, retire and rollback, the published read surface and the additive JSON fields. AC-CDI-010, AC-CDI-013, AC-CDI-022.

> **Cut line.** M0 to M5 deliver the identity layer and stand alone. M6 may land as a second part.

### M6 — Journal and reconciliation

`factory_operations`, the protocol at the four effect sites, `factory reconcile`, the pre-steps. AC-CDI-017 to AC-CDI-021.

### M7 — Close

The scope guards, the read-only `identity report` run on a copy of the live factory database whose counts the reviewer compares with the SQLite counts (RC-3), every regression guard green, the sync-phase hand-off (§ G). AC-CDI-023.

## D. Test plan

| Family | Package | Run command | Notes |
|--------|---------|-------------|-------|
| Every `TestCDI_*` name of `acceptance.md` except the three in the next row | `internal/homestate` | `go test -v -count=1 -run '^(NameA\|NameB\|…)$' ./internal/homestate/` | Anchored selectors; a name with no `--- PASS:` line is red (verification-completeness § 1.1). |
| `TestCDI_RunsFlagSet`, `TestCDI_IdentityVerbsLaneRefused`, `TestCDI_ReconcileNoLeaderImports` | `internal/cli` | `go test -v -count=1 -run '^(…)$' ./internal/cli/` | Verb-level tests. Heavy package: take the slot lease first (`moai slot acquire --resource go-test-heavy`), release after. |

Rules: each behavior test has a **control cell** (the same fixture under the contrasting condition) so it can fail; each criterion lists the **mutants** it closes (acceptance.md). No `go test ./...` locally; no skip; no timeout widening; cross-build check `GOOS=windows GOARCH=amd64 go build ./...` at M7 (the new code uses no platform-specific facility).

## E. Plan-artifact hash binding

The plan-artifact hash is the SHA-256 the runtime computes over the SPEC directory (`internal/runtime/audit_cache.go`): for each of `acceptance.md`, `decision-index.md`, `design.md`, `plan.md`, `research.md`, `spec.md`, `tasks.md`, in that order, present files only, write `<name>:<byte length>`, a zero byte, the file bytes, one zero byte, into one stream; lowercase hex. `progress.md` is not hashed, so the plan-phase record may be updated after the audit; any change to a hashed file after the audit invalidates it. `decision-index.md` is hashed unconditionally, so its `DEFAULT-APPLIED` stamps are written last, before the audit. `tasks.md` is absent.

## F. Risks and mitigations

See `spec.md` § F. The two with a mechanical answer: R-3 (driver behavior) is answered by AC-CDI-007 and AC-CDI-009 running under the project driver; R-8 (sealed attempts read as leases) is answered by AC-CDI-011's reader assertion and the inventory below.

## G. Sync-phase hand-off

- CHANGELOG entry and the operator sequence of § A.5 (manager-docs).
- A grep at B0 found `runs --retire` in no rule, skill, doc or `AGENTS.md`; five Go message sites tell an operator to retire (`internal/cli/factory.go:500`, `:545`, `:687`, `internal/cli/factory_run_adopt.go:103`, `internal/factory/factory_slots.go:142`). They stay unchanged, as SPEC-FACTORY-RETIRE-GUARD-001 Q5 also leaves them; on the identity layout the retirement they point at now migrates.
- MX tags for the new funnels (the seal and migration writers are high fan-in once routed).

## H. Cross-references

- `spec.md` § E (interface, defect disposition), `design.md` (layout, procedures, tables), `acceptance.md` (criteria, ledger), `research.md` (evidence), `decision-index.md` (rows Q1 to Q7).
- `.claude/rules/moai/development/verification-completeness.md` (RED-now cells, mutant probe), `.claude/rules/moai/core/verification-claim-integrity.md` § 2.3 (ordering attribution, AC-CDI-024), `.moai/docs/audit-artifact-convention.md` (verdict files).
- Code coordinates in `design.md`; stage-A material under worktree t1841 (`.moai/reports/t1841/`, `.moai/specs/SPEC-CARD-LIVENESS-OWNERSHIP-001/`).

## I. Statement inventory (mechanical routing; last)

The fifteen lines that name `cards` in non-test code at B0 (`grep -rn "FROM cards\|INTO cards\|UPDATE cards\|JOIN cards" internal --include='*.go' --exclude='*_test.go'`), and the routing each gets:

| # | Site | Kind | Routing on the identity layout |
|---|------|------|--------------------------------|
| 1 | `internal/homestate/card_worktree.go:77` | read (join with `runs`) | legacy only; finds nothing on the identity layout (design § 4.6) |
| 2 | `internal/homestate/card_transition.go:527` | write (renew) | `card_current` |
| 3 | `internal/homestate/card_transition.go:877` | write (`updateCardRow`) | `card_current` |
| 4 | `internal/homestate/card_record.go:181` | read (`loadCard`) | `card_current` plus membership check |
| 5 | `internal/homestate/card_record.go:197` | read (`ListCards`) | `card_current`, listed by membership |
| 6 | `internal/homestate/leader_approval.go:257` | read (count) | `card_current` or the view |
| 7 | `internal/homestate/leader_approval.go:433` | read (count) | `card_current` or the view |
| 8 | `internal/homestate/card_reassign.go:23` | read (cross-run live lease) | unreachable on the identity layout |
| 9 | `internal/homestate/factory_run_retire.go:331` | read (boot-proof activity) | `card_current` plus `card_attempts` timestamps |
| 10 | `internal/homestate/factory_run_adopt.go:34` | read (live records) | `card_current` by membership |
| 11 | `internal/homestate/card_picked.go:166` | write (insert) | the three cases of design § 4.3 |
| 12 | `internal/homestate/card_picked.go:275` | read (predecessor) | current record and sealed attempts |
| 13 | `internal/factory/gtd_operation.go:479` | read (count) | the view (reads keep working) |
| 14 | `internal/factory/gtd_operation.go:529` | read (count) | the view |
| 15 | `internal/factory/gtd_operation.go:558` | read (`run_id` and `card_id`) | the view; a card not in that run reads as zero rows, as a missing row does today |

Writers are rows 2, 3 and 11 plus the new seal, join and migration writers; every other row is a reader. A reader that must see sealed attempts reads `card_attempts` by name; none reads a sealed attempt as a lease (REQ-CDI-012, AC-CDI-011).

## J. Revision after plan audit iteration 1 (FAIL 0.74)

Source: `.moai/reports/t1842/plan-audit.md`. Probes for every engine claim below ran under the project driver (`modernc.org/sqlite` v1.60.1, SQLite 3.53.4) in a scratch program, not on the command line (research.md § 6). Lane rulings: Q2 follows design principle 6; Q7 is strict.

| Defect | Disposition | Where |
|--------|-------------|-------|
| D1 rollback keeps trigger names | Fixed: rollback drops and recreates triggers and indexes under `__rb<N>` names; AC extended (new tables refuse update, delete, replace after the second cutover; distinct names). Probe: six distinct trigger names, both tables protected. | design § 7, § 2.2; AC-CDI-010 |
| D2 assign committed, lease not | Fixed: a queue claim is one operation across both transitions with a `progress` marker set in the assign transaction; new cell `queue_claim__assigned_only` (fourteen cells); relation to the `BeforeCommit` ordering stated. | design § 6.1 to § 6.3; REQ-CDI-017, REQ-CDI-019; AC-CDI-019 |
| D3 conflict rule ignores owners and liveness | Fixed per ruling: strict rule on distinct owner labels plus unexpired lease or registered-worker holder on a row to be sealed; report lists owners and expiries; Q7 ruled. Fresh counts recorded. | REQ-CDI-003, REQ-CDI-005; design § 3.3 C-2; AC-CDI-003, AC-CDI-005; research § 4 |
| D4 "removed by construction" overclaim | Fixed: wording narrowed to duplicate-row classes; surviving single-row hazards are an explicit out-of-scope item with seam S-2. | spec § E.3, § G; design § 8.1, § 8.2 |
| D5 file-set byte identity | Fixed: invariant restated as `factory.db` plus table content, sidecars permitted; probe shows `mode=ro` creates `-wal`/`-shm`, `immutable=1` does not (not used, unsafe beside a writer). | REQ-CDI-003; design § 3.1; AC-CDI-003 |
| D6 `attempt_no` across rollback | Fixed: per-card floor kept in the `__rb<N>` archive, next cutover numbers above it; AC-011 mix gains reassign, rollback and re-cutover. | REQ-CDI-007, REQ-CDI-011, REQ-CDI-012; design § 1, § 7; AC-CDI-011 |
| D7 idempotence beyond in-flight rows | Fixed: intent lookup in every state with a defined return per terminal state; concurrent presentations resolve to one row. | REQ-CDI-018; design § 6.2; AC-CDI-018 |
| D8 REPLACE onto a terminal journal row | Fixed: `BEFORE UPDATE` guard on `op_id`, `kind`, `intent_key`, `card_id`, `attempt_no`. Probe: the replace-onto statement aborts, the terminal row survives. | REQ-CDI-021; design § 2.2; AC-CDI-021 |
| D9 sealed attempt is not a lease | Fixed: clause added to REQ-CDI-012 naming the readers. | REQ-CDI-012; AC-CDI-011 |
| D10 Q2 narrowing | Fixed per ruling: principle 6 followed. Feasible: every write that ends a grant seals the pre-write record in its own transaction (one row per ended grant, no second transaction); no obstacle found, so no open row. | REQ-CDI-023; design § 4.4; AC-CDI-025; decision-index Q2 |
| D11 column-adding migrations vs the view | Fixed: S-6 and D-5 scoped to the version string; rule for a column-adding migration stated and probed (alter the three tables, rebuild the view and triggers in one transaction). | spec D-5, S-6; design § 2.7; plan § B-2 |
| D12 REQ-015 vs AC-014 columns | Fixed: `attempt_no` added to REQ-CDI-015. | REQ-CDI-015 |
| D13 AC-023/024 trace | Fixed: D-8 added for AC-CDI-023; AC-CDI-024 is mapped to the plan's M0 ordering explicitly. | spec D-8, § H |
| D14 existence-only RED and file-pinned E20 | Fixed: E00 note states the behavioral red is `baseline.md`; E20 replaced by a pair over the `operation_*.go` glob plus an import-graph test. | acceptance E00, E20a, E20b |
| D15 `FactoryTablePresent` | Fixed: query `type IN ('table','view')` for `cards`; subtest added. | design § 9; AC-CDI-022 |
| D16 relaunch after a refused migration | Fixed: stated (prints the left-active line, still launches, exit status as for any refused retirement). | REQ-CDI-016 |
| D17 S-2 at cutover | Fixed: S-2 states the numbering at cutover. | spec S-2 |
