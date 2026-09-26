# plan.md — SPEC-FACTORY-RECORD-001 (card t1239)

## §A. Context

- Worktree `.claude/worktrees/t1239`, branch `WT-factory-record-state`, base develop `553e224f3`.
- Tier L: 25 requirements, 24 acceptance criteria (ceilings 25 / 25).
- Development mode: TDD (RED-GREEN-REFACTOR). Every behaviour AC has a RED-now cell in
  `acceptance.md` §C.1 measured on `553e224f3`.
- Card family: F1 (this) → F2 (self-dispatch lane verbs + launcher) → F3 (controller + Decider). A1
  (t1234) is independent of F1's schema; see research.md R11.
- Packages touched (expected): `internal/homestate` (schema, migration, transition API, lease),
  a git-evidence reader (new file or package — run-phase choice), `internal/cli` (`factory assign /
  status / decide`, dispatch wiring in `gtd.go` and `goal.go`). `internal/kanban` is read only
  (queue item state). No file under `internal/template/templates/` is expected to change; if one does,
  template neutrality applies (no SPEC IDs, card ids, dates, or SHAs in template content).

## §B. Known issues (from research.md)

- B-1. `homestate` cannot import `kanban` (`kanban` already imports `homestate`: `factory_slots.go:31`,
  `state_dir.go:12`, `todo_root.go:52`). The queue-state precondition of REQ-FR-022 is therefore
  enforced in the command layer, which already imports both.
- B-2. `moai factory <unknown>` prints help and exits 0 (research.md R9). CLI ACs assert output
  content and database content, never the exit code alone.
- B-3. `RecordCard` is an unconditional writer with no production caller (R3); see §C.6.
- B-4. `internal/cli` and `internal/hook` whole-package tests write the real `~/.moai` profile-lease
  database (R14). Run CLI tests only with `-run` filters under an isolated MoAI home.
- B-5. The integration branch name must come from configuration, not a literal `develop`; the
  distributed template must stay neutral.

## §C. Decisions for the operator / lead (recommended defaults)

Each item states the default this plan proceeds with unless the lead or operator overrides it before
Implementation Kickoff Approval.

1. **F1 / F2 / F3 split of launcher and lane verbs.** Default: F1 changes no launcher and adds no
   lane-facing verb. F2 owns `moai factory next / stage / complete / heartbeat`, removal of the
   still-live `-k N` factory shape (`internal/cli/kanban.go:124-131`), the stale `-f [N]` error text
   (`internal/cli/factory.go:202-203`), and making every `-f` launch self-dispatch. F3 owns
   `moai factory run`, the Decider (`human | llm | jev`, default `llm`), automatic local merge, CI
   evidence production, and notifications.
2. **Lease duration.** Default: 15 minutes, a single constant in the existing defaults file; no new
   YAML key in F1 (a new key would change `.moai/config` templates and `moai update` behaviour).
   F2 renews at a fraction of it; F3 may make it configurable.
3. **Queue runtime report vs factory record.** Default: the factory record is authoritative for card
   state; `todo_runtime_assignments` stays as a report and keeps being written by the existing paths
   (both dispatch callers read it back for idempotency, research.md R4). No reader is switched in F1.
4. **Machine-readable verdict lines.** Default: F1 defines two lines an audit verdict file must carry
   for REQ-FR-009 — `verdict: <PASS|PASS-WITH-DEBT|FAIL>` and `audited_sha: <sha>` — at the start of a
   line anywhere in the file. The audit-artifact convention (`.moai/docs/audit-artifact-convention.md`)
   and the plan-auditor / sync-auditor output instructions are updated in sync-phase by manager-docs to
   emit them. Until auditors emit them, audit-exit transitions are refused (fail closed) and the card
   waits in the audit state — which is the intended behaviour of "Go decides".
5. **Factory write failure on the existing dispatch paths (REQ-FR-025).** Default: fail-open — the
   dispatch completes as today, the factory write error is printed to stderr with a
   `FACTORY_RECORD_UNAVAILABLE` prefix, and nothing is retried. Rationale: "behaviour identical to
   today" is the requirement; a new failure mode on a working path is the larger risk. Alternative:
   fail the dispatch.
6. **`RecordCard` disposition.** Default: remove it (no production caller) so the transition API is
   the only writer of `cards`. The two test references use raw SQL and are unaffected.
7. **Non-human Kickoff decider evidence.** Default: out of F1; `decide` refuses any decider other than
   `human`. F3 decides whether an `llm`/`jev` decision must cite a line of A3's
   `~/.moai/db/<project-key>/contract/receipts.jsonl` (research.md R12).
8. **Predecessor with no factory record (`after:`).** Default: refuse with an unknown-predecessor
   error; the operator clears the hint with `assign --after ""`. Alternative: treat a queue item that
   is no longer `picked` as merged — rejected, because a dropped card would then unblock its successor.
9. **Lease expiry while `merging`.** Default: `blocked` (REQ-FR-015), because a half-finished merge
   in the integration worktree needs a human look before anyone else takes the card.

## §D. Constraints

- Queue schema (`backlogDDL`, `gtdDDL`, `todoRuntimeDDL`) byte-unchanged; no queue item row written.
- No `CHECK` constraint added to `cards.state` (research.md R5 rationale).
- Every migration step additive (`ADD COLUMN ... NOT NULL DEFAULT ''`), inside one transaction.
- Git evidence is read with git subprocesses in the card's worktree; F1 never runs `git fetch`,
  `checkout`, `merge`, `reset`, or any other ref-moving command.
- Tests use `t.TempDir()` project roots and fixture git repositories; no test touches the real
  `~/.moai` or the real `develop`.
- Code comments in English; errors wrapped with `fmt.Errorf("...: %w", err)`.
- No time estimates; milestones are ordered by priority and dependency only.

## §E. Self-verification (run-phase deliverables)

- E1 AC matrix with the verbatim `--- PASS:` line per AC test (`TestFR_AC0NN_*` naming convention).
- E2 `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`.
- E3 `go test -cover ./internal/homestate/...` (≥ 85%).
- E5 `golangci-lint run` on touched packages, NEW vs baseline separated.
- E8 verbatim RED output per milestone, captured before GREEN.

## §F. Milestones (ordered by decision reversibility — most change-prone first)

- **M1 — Data model and migration (Priority High).** Schema v4 columns, `migrateFactoryV3ToV4`,
  state set, legacy-row classification. ACs: 001, 002, 003. Highest reversal cost once rows exist in
  the field, so it lands and is reviewed first.
- **M2 — Transition API and concurrency (Priority High).** Transition table, version compare,
  single-transaction event append, stale / illegal refusals. ACs: 004, 005, 006, 007.
- **M3 — Evidence readers (Priority High).** E-ENTRY, E-VERDICT (with §C.4 line format),
  E-MERGE, push ancestry, no-remote detection. ACs: 008, 009, 010, 018, 019.
- **M4 — Lease and heartbeat (Priority High).** Acquire, renew, expiry return, `merging → blocked`.
  ACs: 011, 012, 013.
- **M5 — Operator commands (Priority Medium).** `assign` (hints, queue-picked precondition),
  `status` (read-only), `decide` (kickoff batch, needs-decision, push gate). User-facing surface;
  ACs: 014, 015, 016, 017, 020, 021, 022, 024.
- **M6 — Wiring existing dispatch paths (Priority Medium).** `gtd.go` and `goal.go` mirror writes
  with §C.5 fail-open. AC: 023.
- **M7 — Cleanup (Priority Low).** Remove `RecordCard` per §C.6; `@MX` annotations on the transition
  API (`@MX:ANCHOR` — every writer funnels through it).

## §G. Anti-patterns to avoid

- Accepting a verdict, tree hash, or "passed" flag as an API argument (violates REQ-FR-011).
- Checking only a CLI exit code (B-2).
- Letting `status` perform the lease-expiry return (violates REQ-FR-024).
- Mutating git state while reading evidence.
- Switching the dispatch idempotency readback from the queue report to the factory record in F1
  (that is F2's move, and it changes lane-visible behaviour).

## §H. Cross-references

- `spec.md` §B (requirements), `design.md` (transition table, readers, commands),
  `acceptance.md` (RED-now ledger and AC matrix), `research.md` (R1-R14).
- A1: branch `WT-contract-schema`, `SPEC-AUTONOMY-CONTRACT-001` design.md § F1 Reference Shape.
- A3: branch `WT-contract-gate-rewire`, `SPEC-AUTONOMY-GATE-REWIRE-001` REQ-GR-012.
- `SPEC-FACTORY-RUN-RETIRE-001` — owner of the v2 → v3 migration and the `cards.updated_at` reader.
