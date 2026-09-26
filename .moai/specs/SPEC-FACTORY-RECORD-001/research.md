# research.md — SPEC-FACTORY-RECORD-001 (card t1239)

Every claim below was measured in this worktree at `553e224f3` (branch `WT-factory-record-state`,
base develop `553e224f3`) on 2026-09-26, unless it names another ref. Line numbers are coordinates
in that tree and will drift; the quoted identifiers are the durable anchors.

## R1. Where factory.db lives

- `internal/homestate/paths.go:229-235` `FactoryDBPath` = `FactoryDir(root)/factory.db`;
  `FactoryDir` (`:221-227`) = `ProjectDir(root)/factory`.
- `ProjectDir` (`paths.go:142-153`) resolves to `~/.moai/db/<project-key>/` (or
  `<root>/.moai/db/<key>/` under the root layout). The queue lives beside it:
  `BacklogDBPath` = `ProjectDir/todo/backlog.db` (`paths.go:213-219`).
- The factory messaging broker is a **separate** SQLite file per run:
  `internal/factorymsg/store.go:130-139` `BrokerPath` = `FactoryDir/messages/<run-id>/broker.db`.

## R2. factory.db schema today (version 3)

- `internal/homestate/factory.go:20` `const factorySchemaVersion = 3`.
- `cards` (`factory.go:44-52`): `run_id, card_id, owner_label, state, version INTEGER DEFAULT 1,
  evidence_path, updated_at`, `PRIMARY KEY(run_id, card_id)`. No lease, stage, decision, hint, SHA, or
  contract column. `state` has no CHECK constraint.
- `workers` (`factory.go:24-32`): `label PK, pid, backend, session_id, run_id, registered_at,
  heartbeat_at`.
- `runs` (`:33-43`), `events(seq, run_id, kind, payload_json, created_at)` (`:54-60`), plus
  dead_letters and handoff tables.
- Migration chain: `OpenFactoryPath` runs `factoryDDL`, reads `meta.schema_version`, then applies
  `migrateFactoryV1ToV2` and `migrateFactoryV2ToV3` in order (`factory.go:186-197`) and rejects any
  other version (`:198-200`). `migrateFactoryV2ToV3` (`:285-313`) is the pattern F1 follows: read
  existing columns via `PRAGMA table_info`, `ALTER TABLE ... ADD COLUMN` only the missing ones, update
  `meta` inside the same transaction. It exists because `CREATE TABLE IF NOT EXISTS` does not add
  columns to an existing table.
- Existing migration test to model: `TestMigrateFactoryV2ToV3PreservesRows`
  (`internal/homestate/factory_run_retire_test.go:236`).

## R3. The existing cards writer is unconditional and unused

- `internal/homestate/runtime.go:53-84` `RecordCard`: `INSERT ... ON CONFLICT DO UPDATE SET ...
  version=cards.version+1` (`:66-69`) — the version is incremented but never compared, so two writers
  never see each other. It appends a `card.updated` event in the same transaction (`:72-82`).
- Production callers: `grep -rn "\.RecordCard(" internal cmd --include='*.go' | grep -v _test` returns
  nothing. Readers of `cards`: `internal/homestate/factory_run_retire.go:259` (latest `updated_at` for
  run retirement) and the test insert at `factory_run_boot_proof_test.go:114`.
- Consequence: F1 may redefine the card row's write contract without breaking a live caller. The
  retirement reader only needs `updated_at` to keep being written.

## R4. Where per-card factory state is recorded today

- `internal/kanban/factory_runtime.go:58-71`: `RecordFactoryCardAssignment` writes state `picked`
  with event `card.assigned`; `RecordFactoryCardState` writes an arbitrary state. Both go to
  `backlog.db` via `recordRuntime` (`internal/kanban/todo_runtime.go:158`).
- Table: `todo_runtime_assignments(run_id, card_id, owner_label, reported_state, event_kind,
  provenance_json)` (`todo_runtime.go:53-57`), declared "the latest execution report, not card
  completion authority" (`todo_runtime.go:14`).
- Callers: `internal/cli/gtd.go:249` (queue dispatch), `internal/cli/goal.go:864` (auto-mission
  dispatch), `internal/cli/todo.go:1008` (`recordFactoryCardState`, e.g. `queued`/`card.unpicked` at
  `:989`). Both dispatch callers read the assignment back from the queue runtime report for their
  idempotency check (`gtd.go:237-247`, `goal.go:850-860`), so that report must keep being written
  (REQ-FR-025).

## R5. Queue items schema and its existing freeze guard

- `internal/kanban/backlog_sqlite.go:112-119`: `items(seq, id UNIQUE, text, added_at, spec_id, state
  CHECK (state IN ('queued','picked','dropped')))`. The comment above `backlogDDL` (`:100-106`)
  explains SQLite cannot ALTER a CHECK constraint, which is why the queue schema is frozen.
- Existing guard: `TestBacklogDowngrade_PreChangeBinaryStillServes`
  (`internal/kanban/backlog_downgrade_test.go:96`) compares a verbatim frozen copy against the live
  `backlogDDL` (`:161`). F1's queue-boundary AC reuses it and adds a byte comparison of the live
  `sqlite_master` rows.
- `gtdDDL` (`internal/kanban/backlog_gtd_schema.go:20-94`) is additive in the same file and owns its
  own `gtd_meta` version; F1 touches neither.

## R6. Worker roster and heartbeat

- Writers of `workers`: `internal/kanban/factory_slots.go:96-104` (`SaveFactoryRegistry`, delete-all
  then insert), `:248` (`ClaimFactoryWorker`), `:225` (prune), and the legacy import
  (`homestate/factory.go:424`). Every insert sets `heartbeat_at = registered_at`.
- `grep -rn "UPDATE workers" internal --include='*.go'` returns nothing: no code ever refreshes
  `heartbeat_at`. REQ-FR-012 is the first writer of a live heartbeat.

## R7. The messaging dispatch record

- `internal/factorymsg/dispatch.go:15-23` states `assigned, delivered, started, result_recorded,
  integrated, abandoned`, with a fencing generation (`:43-45`). It lives in the per-run broker.db
  (R1), not factory.db.
- `CreateDispatch` (`:105`) and `ApplyResult` (`:237`) have no caller outside `internal/factorymsg`'s
  own file and tests (grep over `internal cmd`, `_test` excluded). The lane dispatch that runs today is
  the lead sending a `factory_msg_send` message (`internal/cli/mcp_server.go:588-595`) plus the
  queue-runtime write of R4. "Behaviour identical to today's messaging dispatch" (REQ-FR-025)
  therefore means: the message path and the R4 write are untouched, and the factory record is written
  alongside.

## R8. MCP surface

- `internal/cli/mcp_server.go:588-595` registers `factory_msg_send`, `factory_msg_list`,
  `factory_msg_body`, `factory_msg_receipt`, `factory_msg_status`. No card-state tool exists. F1 adds
  none (the lane-facing surface is F2).

## R9. `moai factory` command tree and the exit-code hazard

- `internal/cli/factory_handoff_recover.go:19-55`: `factory` has `handoff` (`recover-resume`,
  `abandon-lane`) and `runs`. No `status`, `assign`, `decide`.
- Measured with a binary built from this tree (`go build -o <scratch>/moai ./cmd/moai`, HEAD
  `553e224f3`): `moai factory status`, `moai factory decide`, and `moai factory assign` each print the
  `factory` help (COMMANDS: `handoff`, `runs`) and **exit 0**. An AC that checks only the exit code of
  these commands would pass on today's tree; the ACs therefore assert output content.

## R10. The `-f` / `-k` launcher state (input to the F2 split)

- `internal/cli/factory.go:67-73`: "The numeric count form was REMOVED"; `:237-240` bare `-f` uses
  `config.DefaultFactoryLeadWorkers`. A numeric value falls through to the usage error (`:159-165`).
- Still live: `-k N` selects Factory Mode with N workers (`internal/cli/kanban.go:52`, `:124-131`).
- Stale text: the `-k`/`-f` conflict error still says "use ... -f [N] for the factory"
  (`internal/cli/factory.go:202-203`).
- Both are launcher work, assigned to F2 (plan.md §C.1).

## R11. A1 (card t1234) — where its contract lives and how F1 shares the store

Re-read on 2026-09-26 at the branch tip, replacing the first reading of `6d98ca466` (v0.4.1, two
versions stale):

- `git rev-parse WT-contract-schema` → `de8aee45697d5e23c710fdcc9edc62c0489838ab` (short `de8aee456`).
  The last documentation commit on that branch is `65e0a9167` ("v0.5.1 decider llm|llm+jev, A1 schema /
  A3 rules split, card field, store path"); `086dfb2fd` and `de8aee456` are run-phase code commits.
  `SPEC-AUTONOMY-CONTRACT-001` reads `version: "0.5.1"`, `status: in-progress` (spec.md:4-5 at
  `de8aee456`). Not on develop: `.moai/specs` on this tree has no `SPEC-AUTONOMY-CONTRACT-001` directory.
- The contract is a file, `.moai/specs/<SPEC-ID>/contract.yaml` (A1 spec.md:51 at `de8aee456`). A1
  creates no database table.
- A1 spec.md:108-111 (`de8aee456`): "The factory redesign's `factory.db` card records are not
  implemented or modified here ... the storage itself is F1." §C.4 (`:166-171`): a factory card record
  carries a pointer only — SPEC ID, signed contract digest, signing time — never a copy.
- A1 design.md:539-553 (`de8aee456`) § F1 Reference Shape: `contract_ref { spec_id, contract_sha256,
  signed_at }`; "F1 decides storage; A1 guarantees the three fields are stable and available from
  `show --json`."
- New in v0.5.1 and relevant here (A1 spec.md:30, HISTORY 0.5.1, at `de8aee456`): the decider set is
  `human | llm | llm+jev` (Jev is never a sole decider); a required top-level `card` field sits inside
  the canonical digest; and lead decision R10 — "the moai-owned store path is
  `$MOAI_HOME/db/<project-key>/contract/` (A3-owned; A1 does not create it)" (also spec.md:96, :195,
  :204).
- **Sharing, concretely:** same project directory (`ProjectDir`, R1), different artifacts. F1 adds four
  text columns to `cards` for the pointer (A1's three fields plus the R15 store-event locator). There is
  no shared table and no migration-ordering dependency: A1 has no schema, so F1's factory.db migration
  can land before or after A1. F1 does not import A1's package and does not verify the digest; a
  stale-pointer check is a reader concern for A1/A3 (`moai contract verify`). F1 does not copy A1's new
  `card` field into the pointer: the pointer sits on the card row, which already names the card.

## R12. A3 (card t1236) signing-event store

- `git rev-parse WT-contract-gate-rewire` → `781ddc3554c4aa9cdf2f152eeab4b7b7eca5a03e`;
  `SPEC-AUTONOMY-GATE-REWIRE-001` `version: "0.3.0"`. spec.md:93 REQ-GR-012: an append-only
  hash-chained JSONL store at `~/.moai/db/<project-key>/contract/receipts.jsonl` recording every signing
  event (human, receipt, re-seal), decide receipt, and revoke; each line carries the previous line's
  hash. Its own citation of A1 still names `6d98ca466` — A3's text, not F1's to correct.
- That store is under `ProjectDir/contract` as a separate file. F1 never writes it. Open question for
  the F3 Decider (plan.md §C.7): whether a non-human Kickoff decision in the factory record must cite a
  line of that store.

## R13. Audit verdict files carry no machine-readable verdict line today

- `.moai/docs/audit-artifact-convention.md:35-44`: card verdicts live at
  `.moai/reports/<card-id>/plan-audit.md`, `plan-audit-iter<N>.md`, `sync-audit.md`.
- `:70-84` "What": the file carries "the verdict token and score" in prose; no fixed line format and no
  audited-SHA field is required. The document is mirrored at
  `internal/template/templates/.moai/docs/audit-artifact-convention.md`.
- Measured: `grep -c "audited_sha" .claude/agents/moai/plan-auditor.md` → `0` (exit 1);
  `grep -c "audited_sha" internal/template/templates/.claude/agents/moai/sync-auditor.md` → `0`
  (exit 1). Emitted Codex definitions exist at
  `internal/template/templates/.codex/agents/moai/{plan,sync}-auditor.toml` and are regenerated by
  `make agents-emit`.
- Consequence (plan-audit iteration 1, D6): without a producer, E-VERDICT refuses every audit exit
  forever. REQ-FR-011 therefore makes the producer part of this SPEC.

## R14. Test-environment constraint

- Lead instruction for this card: do not run whole-package `./internal/cli` or `./internal/hook` tests
  — they write the real `~/.moai` profile-leases database. `homestate` tests already use
  `t.TempDir()` project roots. CLI-level ACs must run with a `-run` filter and an isolated MoAI home.

## R15. The moai-owned contract store (lead decision R10)

- Lead decision R10 (2026-09-26, recorded in A1 HISTORY 0.5.1 at `de8aee456`): A2 armed state and A3
  signing events live under `$MOAI_HOME/db/<project-key>/contract/`.
- `ProjectDir` (`internal/homestate/paths.go:142-153`) resolves to `$MOAI_HOME/db/<project-key>/`
  under the default layout, so the store is `ProjectDir/contract`, the sibling of `factory/` and
  `todo/`. No helper returns that path on this tree: `grep -rn '"contract"' internal/homestate/paths.go`
  → no output, exit 1.
- F1's pointer names a store line by its SHA-256 (A3 lines are hash-chained, so a line hash is a stable
  locator) rather than by file name or line number, which A3 still owns.

## Reuse analysis

| Need | Reuse |
|------|-------|
| Migration shape | `migrateFactoryV2ToV3` + `factoryRunColumns` pattern (R2) |
| Transaction + busy retry | `OpenFactoryPath` pragmas (`_txlock=immediate`, busy_timeout) and `retryFactoryBusy` (`factory.go:140-151`, `:357`) |
| Event log | existing `events` table (R2); no new table |
| Worker roster | existing `workers` table (R6) |
| Active run resolution | `factorymsg.ResolveActiveRun` as used by `enterSelectedFactoryRun` (`internal/cli/factory.go:245-256`) |
| Queue item state read | existing kanban backlog store read path (`store.LoadPure`, `gtd.go:238`) |
