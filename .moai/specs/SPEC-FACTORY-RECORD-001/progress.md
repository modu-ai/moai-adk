# progress.md — SPEC-FACTORY-RECORD-001 (card t1239)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, this skeleton (Tier L)
- baseline: worktree `.claude/worktrees/t1239`, branch `WT-factory-record-state`, base develop `553e224f3`
- RED-now ledger: acceptance.md §C.1, pinned to `553e224f3`
- open decisions: plan.md §C items 1-10 (recommended defaults stated)
- plan_audit: iter-1 FAIL 0.71 on `4baab1d7a` (`.moai/reports/t1239/plan-audit.md`; Sonnet, Opus limit)
  → v0.2.0 revision (D1-D9 + lead updates: A1 re-read at `de8aee456`, contract store pointer R10);
  iter-2 PASS-WITH-DEBT 0.94 on `195e22697` (`.moai/reports/t1239/plan-audit-iter2.md`)
  → v0.2.1 debt notes (this commit)
- lead decisions (2026-09-26): all ten plan.md §C defaults adopted — §C marked DECIDED; condition on
  decision 5 folded into REQ-FR-025 and AC-024 / AC-025 (readable unavailable-record log reported by
  `status`, `record.drift` event on the next successful write); condition on decision 10 recorded with
  A3 `WT-contract-gate-rewire` at `710530d67`; D10 regression folded into AC-020 using
  `ParseVerdictLine`. A1 re-pinned to `WT-contract-schema` at `8a7cb0e22` (v0.5.2). REQ 25 / AC 25
  unchanged. Implementation Kickoff Approval still pending.
- tracked debt (from iter-2, all optional class):
  - D7 compound-requirement granularity (REQ-FR-013/018/019/020 and similar) — accepted, not split;
    splitting would exceed the Tier L 25-REQ ceiling
  - D10 `AUDIT-VERDICT:` chat-message convention unacknowledged — addressed by research.md R16, R13
    correction, and the verdict-file-only guardrail in design.md / plan.md M3b / AC-020; residual:
    the guardrail binds the M3b editor, verified at run phase by AC-020
  - D11 `blocked` excluded from T21 without rationale — addressed by one sentence in design.md
  - D12 `contract_event` computation unowned — addressed by a forward note in design.md naming A3/F3;
    residual: the hashed byte form is left to that SPEC

## §E.2 Run-phase Evidence

Tree: branch `WT-factory-record-state`, run start HEAD `04ca1a98c` (local develop `e62c3e183`
absorbed); every measurement below was taken in this worktree on 2026-09-26 after commit
`6e78d9a2d` plus the AC-005 log line (the M-final commit). research.md anchors re-verified at
`04ca1a98c` before M1 (`factorySchemaVersion = 3` at factory.go:20, `version=cards.version+1` at
runtime.go:68, both `RecordFactoryCardAssignment` callers at gtd.go:249 / goal.go:864,
`ParseVerdictLine` at store.go:325, the four `Cite your audit receipt` headings, the kanban guard at
backlog_downgrade_test.go:96) — all present as cited.

### Pre-flight (Section C)

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (at `04ca1a98c`).
- `golangci-lint run ./internal/homestate/... ./internal/cli/... ./internal/template/...` → `0 issues.`

### RED-first ordering predicate

Each milestone's RED commit touches only `_test.go` files and is an ancestor of its GREEN commit
(`git merge-base --is-ancestor <red> <green>` → exit 0 for all eight pairs). API-level REDs compile
against a test-only contract stub (`fr_contract_stub_test.go`, `fr_m4_stub_test.go`,
`factory_card_stub_test.go`, `factory_mirror_stub_test.go`) that the GREEN commit deletes, so each
RED fails on an assertion, never on the build.

| Milestone | RED commit (files) | GREEN | Observed RED failure (verbatim first line) |
|---|---|---|---|
| M1 | `dfaf3b114` (fr_schema_test.go) | `d662f02a0` | `fr_schema_test.go:218: schema_version = "3", want "4"` |
| M2 | `3db5167e0` (fr_contract_stub_test.go, fr_fixture_test.go, fr_transition_test.go) | `d3710780e` | `accepted 0 / refused 361, want 65 / 296`; AC-002/003/004/006/019 `F1 transition API not implemented` |
| M3 | `c82678533` (fr_evidence_test.go) | `e5d7fd538` | `missing sha: err = <nil>, want ErrEvidence`; `no verdict: err = <nil>`; `tree differs: err = <nil>`; `push before the ref contains the merge: err = <nil>` |
| M3b | `bcefc6e9b` (fr_producer_test.go) | `6b0392201` | `.claude/agents/moai/plan-auditor.md: audited_sha: count = 0, want ≥ 1` (all six files, both patterns, both .toml) |
| M4 | `ce3763a97` (fr_lease_test.go, fr_m4_stub_test.go) | `8ca9bc6bd` | `lease by ghost: err = <nil>, want ErrLeaseHolder`; `kickoff card holder="worker-1"`; `failed with an empty reason was accepted` |
| M5 | `02bd782a4` (factory_card_test.go, factory_card_stub_test.go) | `dae2987d8` | `unknown command "assign" for "factory"` / `unknown command "decide" for "factory"` |
| M6 | `8485c6742` characterization (factory_dispatch_test.go, PASS pre-change) → `98786a995` (factory_mirror_test.go, factory_mirror_stub_test.go) | `bfecfee9a` | `factory record for t1 = {…} (present=false), want assigned to worker-2` / `… worker-3` |

Two fixture corrections were made inside GREEN commits and are declared here: M2 GREEN set
`Version = 1` on two fixture cards (AC-006, AC-019) whose local copy carried version 0; the
assertions were unchanged.

### AC matrix

Commands: `go test ./internal/homestate -run 'TestFR_' -count=1 -race -v` (exit 0, swept 23) and
`go test ./internal/cli -run 'TestFR_AC0' -count=1 -v` (exit 0, swept 14).

| AC | Status | Deciding output |
|----|--------|-----------------|
| AC-001 | PASS | `--- PASS: TestFR_AC001_MigrationV3ToV4` |
| AC-002 | PASS | `--- PASS: TestFR_AC002_LegacyStateRefusesAllButAbandon`; `--- PASS: TestFR_AC002_LegacyAbandonViaDecide` |
| AC-003 | PASS | `--- PASS: TestFR_AC003_TransitionIsAtomic` |
| AC-004 | PASS | `--- PASS: TestFR_AC004_StaleVersionRefused` |
| AC-005 | PASS | `requested pairs: 65 accepted, 296 refused, 361 total; production table rows: 66`; `--- PASS: TestFR_AC005_TransitionTableEdgeCount` (66 rows = the 65 accepted + T18, refused with a remote) |
| AC-006 | PASS | `--- PASS: TestFR_AC006_RacingWritersExactlyOneWins` (50 races, `-race`) |
| AC-007 | PASS | `--- PASS: TestFR_AC007_AuditEntryEvidence` |
| AC-008 | PASS | `--- PASS: TestFR_AC008_AuditVerdictGate` |
| AC-009 | PASS | `--- PASS: TestFR_AC009_MergeEvidence` |
| AC-010 | PASS | `--- PASS: TestFR_AC010_LeaseAcquireAndRenew` |
| AC-011 | PASS | `--- PASS: TestFR_AC011_ExpiredLeaseReturnsToAssigned` |
| AC-012 | PASS | `--- PASS: TestFR_AC012_ExpiredMergingBlocks` |
| AC-013 | PASS | `--- PASS: TestFR_AC013_DecisionPendingHoldsNoLease`; `--- PASS: TestFR_AC013_DecideAfterClockAdvance` |
| AC-014 | PASS | `--- PASS: TestFR_AC014_AssignHints` |
| AC-015 | PASS | `--- PASS: TestFR_AC015_DecideKickoffBatch` |
| AC-016 | PASS | `--- PASS: TestFR_AC016_DecideQuestionChoices` |
| AC-017 | PASS | `--- PASS: TestFR_AC017_UnblockAndFailed`; `--- PASS: TestFR_AC017_DecideUnblock` |
| AC-018 | PASS | `--- PASS: TestFR_AC018_PushGateStore`; `--- PASS: TestFR_AC018_DecidePushGate` |
| AC-019 | PASS | `--- PASS: TestFR_AC019_ReservedCIEdgesRefused` |
| AC-020 | PASS | `--- PASS: TestFR_AC020_VerdictLineProducer`; the twelve `grep -c` runs each print 1 (below); `make agents-emit-check` exit 0; both `.toml` carry `audited_sha` (2 each); `TestTemplateNoInternalContentLeak` PASS; `go test ./internal/auditreceipt -run TestParseVerdictLine` → `--- PASS: TestParseVerdictLine` |
| AC-021 | PASS | `go test ./internal/kanban -run 'TestBacklogDowngrade_PreChangeBinaryStillServes'` → `--- PASS`; `--- PASS: TestFR_AC021_QueueSchemaUntouched` |
| AC-022 | PASS | `--- PASS: TestFR_AC022_AssignRequiresQueuePicked` |
| AC-023 | PASS | `--- PASS: TestFR_AC023_StatusIsReadOnly` |
| AC-024 | PASS | `--- PASS: TestFR_AC024_CharacterizeGTDDispatch`; `--- PASS: TestFR_AC024_GTDDispatchMirrorsFactoryRecord` |
| AC-025 | PASS | `--- PASS: TestFR_AC025_CharacterizeGoalDispatch`; `--- PASS: TestFR_AC025_GoalDispatchMirrorsFactoryRecord` |

AC-020 grep counts (`audited_sha:` / `verdict: <PASS|PASS-WITH-DEBT|FAIL>`): plan-auditor local 1/1,
sync-auditor local 1/1, plan-auditor template 1/1, sync-auditor template 1/1, convention local 1/1,
convention template 1/1. In plan-auditor the instruction sits in a new section before
`## MCP Audit Tools`, because the file's `## Output Format` follows the `Cite your audit receipt`
block and AC-020 clause (i) forbids the instruction after that block; the export mandate points to
it without the two literals.

### Other deliverables

- Builds: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Lint: `golangci-lint run ./internal/homestate/... ./internal/cli/... ./internal/config/... ./internal/template/...` → `0 issues.` (baseline 0).
- Template: `go test ./internal/template -count=1` → `ok` (catalog hashes regenerated for the two
  edited auditors via `gen-catalog-hashes.go --all`).
- Existing dispatch tests: `go test ./internal/cli -run '^(TestGTD|TestAutoMission|TestAuthoritativeDispatch|TestGoalMission)'` → 19 PASS, 0 FAIL.
- Coverage: `go test -cover ./internal/homestate/` → `coverage: 76.2% of statements` (below the
  85% DoD line). Baseline at `04ca1a98c`, measured on an extracted copy: `coverage: 69.0%` — the
  package was already below 85% before this SPEC; this SPEC raised it by 7.2 points.
- SQL string concatenation (security-guard advisory): every concatenated fragment is a compile-time
  constant (the column list `cardSelectColumns`, the constant `cardF1Columns` names, fixed table
  names) and every value goes through a `?` placeholder; each site carries a comment saying so —
  factory.go (ALTER TABLE, PRAGMA table_info), card_record.go (loadCard, ListCards), card_picked.go
  (INSERT), card_transition.go (updateCardRow), fr_fixture_test.go (frPlace), fr_schema_test.go
  (frTableColumns, frDump).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: <M-final commit; see git log>
run_status: implemented-pending-sync
ac_pass_count: 25
ac_fail_count: 0
preserve_list_post_run_count: n/a
l44_pre_commit_fetch: not run (no push in this lane)
l44_post_push_fetch: not run (no push in this lane)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: exit 0
  windows_amd64: exit 0
total_run_phase_files: 37  # net diff against 04ca1a98c, including spec.md and progress.md
m1_to_mN_commit_strategy: RED test-only commit then GREEN per milestone (M1-M6), M6 preceded by a characterization commit, M7 cleanup
gaps:
  - homestate package coverage 76.2% < 85% (pre-existing 69.0% baseline)
  - REQ-FR-025 log append vs. reconcile rewrite race is not guarded by a lock (residual)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
