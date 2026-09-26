# plan.md — SPEC-FACTORY-RECORD-001 (card t1239)

## §A. Context

- Worktree `.claude/worktrees/t1239`, branch `WT-factory-record-state`, base develop `553e224f3`.
- Tier L: 25 requirements, 25 acceptance criteria (ceilings 25 / 25). Revision 0.2.0 answers plan-audit
  iteration 1 (FAIL 0.71, `.moai/reports/t1239/plan-audit.md`).
- Development mode: TDD (RED-GREEN-REFACTOR). Every release-blocking AC has a RED-now cell in
  `acceptance.md` §C.1 measured on `553e224f3`.
- Card family: F1 (this) → F2 (self-dispatch lane verbs + launcher) → F3 (controller, Decider, CI
  verdict reader). A1 (t1234, read at `de8aee456`, v0.5.1) is independent of F1's schema; see
  research.md R11 and R15.
- Packages and files touched (expected): `internal/homestate` (schema, migration, transition API,
  lease), a git-evidence reader (new file or package — run-phase choice), `internal/cli` (`factory
  assign / status / decide`, dispatch wiring in `gtd.go` and `goal.go`), and the verdict-line producer
  (plan-auditor and sync-auditor definitions, local and template, emitted Codex definitions, and the
  audit-artifact convention, local and template). `internal/kanban` is read only (queue item state).
  Template edits must stay neutral (no SPEC IDs, card ids, dates, or SHAs in template content), and a
  template agent edit requires `make agents-emit`.

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
- B-6. Editing `internal/template/templates/.claude/agents/moai/*.md` without `make agents-emit` leaves
  the emitted `.toml` stale; `make build` runs the read-only `agents-emit-check` and fails on it.

## §C. Decisions for the operator / lead (recommended defaults)

Each item states the default this plan proceeds with unless the lead or operator overrides it before
Implementation Kickoff Approval.

1. **F1 / F2 / F3 split of launcher and lane verbs.** Default: F1 changes no launcher and adds no
   lane-facing verb. F2 owns `moai factory next / stage / complete / heartbeat`, removal of the
   still-live `-k N` factory shape (`internal/cli/kanban.go:124-131`), the stale `-f [N]` error text
   (`internal/cli/factory.go:202-203`), and making every `-f` launch self-dispatch. F3 owns
   `moai factory run`, the Decider (`human | llm | llm+jev`, per A1 v0.5.1), automatic local merge,
   the CI verdict reader behind `pushed → ci-green`, and notifications.
2. **Lease duration.** Default: 15 minutes, a single constant in the existing defaults file; no new
   YAML key in F1 (a new key would change `.moai/config` templates and `moai update` behaviour).
   Because `kickoff` and `needs-decision` hold no lease (REQ-FR-018), the duration bounds only worker
   liveness, never a human decision.
3. **Queue runtime report vs factory record.** Default: the factory record is authoritative for card
   state; `todo_runtime_assignments` stays as a report and keeps being written by the existing paths
   (both dispatch callers read it back for idempotency, research.md R4). No reader is switched in F1.
4. **Machine-readable verdict lines.** Decided in the SPEC (REQ-FR-011, milestone M3b): auditors write
   `verdict: <PASS|PASS-WITH-DEBT|FAIL>` and `audited_sha: <sha>` at the start of a line. Open for the
   lead only: whether the auditors also keep the prose verdict (default: yes — the lines are added, not
   substituted).
5. **Factory write failure on the existing dispatch paths (REQ-FR-025).** Default: fail-open — the
   dispatch completes as today, the factory write error is printed to stderr with a
   `FACTORY_RECORD_UNAVAILABLE` prefix, and nothing is retried. Alternative: fail the dispatch.
6. **`RecordCard` disposition.** Default: remove it (no production caller) so the transition API is
   the only writer of `cards`. The two test references use raw SQL and are unaffected.
7. **Non-human Kickoff decider evidence.** Default: out of F1; `decide` refuses any decider other than
   `human`. F3 decides whether an `llm` / `llm+jev` decision must cite a line of the A3 store.
8. **Predecessor with no factory record (`after:`).** Default: refuse with an unknown-predecessor
   error; the operator clears the hint with `assign --after ""`.
9. **Lease expiry while `merging`.** Default: `blocked` (REQ-FR-014), because a half-finished merge in
   the integration worktree needs a human look before anyone else takes the card.
10. **Contract-event locator format.** Default: the SHA-256 of the A3 store line that recorded the
    signing (64 hex), empty until the store exists. If A3 settles on a different stable line identifier,
    the column keeps its name and only the format check changes.

## §D. Constraints

- Queue schema (`backlogDDL`, `gtdDDL`, `todoRuntimeDDL`) byte-unchanged; no queue item row written.
- No `CHECK` constraint added to `cards.state` (research.md R5 rationale).
- Every migration step additive (`ADD COLUMN ... NOT NULL DEFAULT ''`), inside one transaction.
- Git evidence is read with git subprocesses in the card's worktree; F1 never runs `git fetch`,
  `checkout`, `merge`, `reset`, or any other ref-moving command.
- F1 never opens, creates, or writes `$MOAI_HOME/db/<project-key>/contract/`.
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
- `make agents-emit-check` exit 0 after M3b.

## §F. Milestones (ordered by decision reversibility — most change-prone first)

- **M1 — Data model and migration (Priority High).** Schema v4 columns (incl. `failure_reason` and the
  four contract-pointer columns), `migrateFactoryV3ToV4`, state set and groups, legacy-row
  classification. ACs: 001, 002. Highest reversal cost once rows exist in the field.
- **M2 — Transition table and concurrency (Priority High).** The 65-edge table with guards (T4a-f,
  T22a-e), reserved T19/T20, version compare, single-transaction event append, stale / illegal /
  reserved refusals. ACs: 003, 004, 005, 006, 019.
- **M3 — Evidence readers (Priority High).** E-ENTRY, E-VERDICT, E-MERGE, push ancestry, no-remote
  detection. ACs: 007, 008, 009, 018.
- **M3b — Verdict-line producer (Priority High).** plan-auditor / sync-auditor definitions (local +
  template), `make agents-emit`, audit-artifact convention (local + template). Must land in the same
  run as M3, or the E-VERDICT gate refuses every audit exit. AC: 020.
- **M4 — Lease, decision-pending, failure (Priority High).** Acquire, renew, expiry return,
  `merging → blocked`, lease release on `kickoff` / `needs-decision`, `failed`. ACs: 010, 011, 012,
  013, 017.
- **M5 — Operator commands (Priority Medium).** `assign` (hints, contract pointer, queue-picked
  precondition), `status` (read-only), `decide` (kickoff batch, choices, push gate). ACs: 014, 015,
  016, 021, 022, 023.
- **M6 — Wiring both existing dispatch paths (Priority Medium).** `gtd.go` and `goal.go` mirror writes
  with §C.5 fail-open; characterization tests of both paths written before the change. ACs: 024, 025.
- **M7 — Cleanup (Priority Low).** Remove `RecordCard` per §C.6; `@MX:ANCHOR` on the transition API
  (every writer funnels through it).

## §G. Anti-patterns to avoid

- Accepting a verdict, tree hash, CI result, or "passed" flag as an API argument (violates REQ-FR-010).
- Implementing a CI evidence reader for T19 in F1 (violates REQ-FR-006).
- Treating `kickoff` as lease-holding (violates REQ-FR-018).
- Checking only a CLI exit code (B-2).
- Letting `status` perform the lease-expiry return (violates REQ-FR-024).
- Mutating git state while reading evidence, or touching the contract store.
- Hand-editing emitted `.toml` agent files (B-6).
- Switching the dispatch idempotency readback from the queue report to the factory record in F1
  (that is F2's move, and it changes lane-visible behaviour).

## §H. Cross-references

- `spec.md` §B (requirements), `design.md` (transition table, edge count, readers, producer, commands),
  `acceptance.md` (RED-now ledger and AC matrix), `research.md` (R1-R15).
- Plan-audit iteration 1: `.moai/reports/t1239/plan-audit.md`.
- A1: branch `WT-contract-schema` at `de8aee456`, `SPEC-AUTONOMY-CONTRACT-001` v0.5.1 design.md § F1
  Reference Shape.
- A3: branch `WT-contract-gate-rewire` at `781ddc355`, `SPEC-AUTONOMY-GATE-REWIRE-001` REQ-GR-012.
- `SPEC-FACTORY-RUN-RETIRE-001` — owner of the v2 → v3 migration and the `cards.updated_at` reader.
