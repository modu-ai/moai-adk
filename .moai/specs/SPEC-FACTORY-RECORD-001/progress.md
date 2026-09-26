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

### Follow-ups ordered by the lead before sync (HEAD `c0624427a`)

**1. Unavailable-log append vs rewrite race (data-loss path).**

- `dc2566c89` — behavior-neutral seam: a nil `recordUnavailableRewriteHook` called between the
  rewrite's read of `record-unavailable.jsonl` and its rename (production code; the hook is nil, so
  no behavior change — `TestFR_RecordUnavailable*` still `ok`). Needed so the RED commit could be
  test-only and the race reproduced by ordering, not timing.
- `b679814e4` — RED, test-only (`internal/homestate/fr_unavailable_race_test.go`). Command
  `go test ./internal/homestate -run 'TestFR_UnavailableLog' -count=1 -v`, observed:
  `fr_unavailable_race_test.go:60: the entry appended during the rewrite was lost: [{ID:first ... Reconciled:true}]`
  / `--- FAIL: TestFR_UnavailableLogAppendSurvivesRewrite`. (The concurrent-appends companion passed
  on the unfixed code — it is probabilistic and serves as the `-race` run, not as the RED.)
- `14a23a6d4` — fix: appends and the whole read→temp→rename rewrite run under one exclusive file
  lock (`record-unavailable.jsonl.lock`, the existing admission-lock primitive: flock / LockFileEx).
  Command `go test ./internal/homestate -run 'TestFR_' -count=1 -race -v`, observed:
  `--- PASS: TestFR_UnavailableLogAppendSurvivesRewrite`, `--- PASS: TestFR_UnavailableLogConcurrentAppendsNoLoss`,
  `ok github.com/modu-ai/moai-adk/internal/homestate 19.695s`. `GOOS=windows GOARCH=amd64 go build ./internal/homestate` exit 0.
  AC-023/024/025 re-run: `--- PASS` ×5.
- Ordering: `git merge-base --is-ancestor dc2566c89 b679814e4` exit 0; `… b679814e4 14a23a6d4` exit 0.

**2. Push gate never runs `git fetch`.**

- `984773644` — `internal/cli/factory_push_nofetch_test.go` (`TestFR_AC018_PushGateNeverFetches`):
  a recording git wrapper first on PATH logs every call of `moai factory decide --gate push`; it fails
  on any `fetch` and, as a positive control, requires the log to contain the gate's `remote` and
  `merge-base --is-ancestor` reads. Current code passes: `--- PASS: TestFR_AC018_PushGateNeverFetches`.
- The current code already passed, so the RED step was mutant-based (neither mutant committed):
  (a) a `gitRead(ctx, dir, "fetch", "--quiet", remote)` inserted in `verifyPushed` →
  `the push gate ran git fetch: "-C …/repo fetch --quiet origin"` / `--- FAIL`;
  (b) a blind recorder (wrapper logs nothing) → `positive control: recorder did not see the push gate's git reads (remote=false ancestry=false)` / `--- FAIL`.
  Both reverted; the test passes again. POSIX only (skipped on windows: the wrapper is a shell script).

### Coverage debt (follow-up card candidate)

`go test -cover ./internal/homestate/ -count=1` → `coverage: 76.2% of statements` on this branch
(re-measured after the follow-ups at `984773644`); the same command on a copy of base
`04ca1a98c` extracted with `git archive` → `coverage: 69.0% of statements` (that copy also failed
`TestTempDiscriminantParity`, an environment premise tied to the copy living under a temp root). The
85% DoD line is not met and was not met before this SPEC. Candidate follow-up card: raise
`internal/homestate` coverage to 85% (largest untested areas pre-date F1).

### Design deviations (implemented differently from design.md; ACs govern)

1. **plan-auditor instruction placement** — the verdict-file lines sit in a new section before
   `## MCP Audit Tools`, not in the output-format section, because that section follows the
   receipt-citation block and AC-020 (i) forbids the instruction after it.
2. **Version compare first** — the stale-version check runs before the legacy, reserved-edge, and
   table checks (design pseudocode had it at the UPDATE), because AC-006 requires the losing writer
   to get a stale-version error, not an illegal-transition error.
3. **Abandon keeps every other column** — `abandon` changes only state, version, and `updated_at`
   (lease and decision fields stay, inert on a terminal card), per AC-016's "every other column
   unchanged".
4. **AC-025 failure scenario in a second project root** — the goal path's lane-ownership check
   refuses a second card for `worker-3` in the same queue, so the d4 failure/drift scenario runs in
   a fresh root.

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
  - (closed by 14a23a6d4) REQ-FR-025 log append vs. reconcile rewrite race — now serialized by a file lock
  - push-gate no-fetch test is POSIX-only (skipped on windows)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: pending-backfill-sync
sync_status: completed
b12_self_test_a: "grep -c SPEC-FACTORY-RECORD-001 CHANGELOG.md -> 0 before emission"
b12_self_test_b: "distinct AC ids in acceptance.md -> 25; CHANGELOG entry cites 25"
b12_self_test_c: "every path cited in the entry verified with ls"
changelog_entry_position: "[Unreleased] ### Added, first entry"
docs_site: "no page covers the moai factory subcommands; none created (factory-mode.md covers the launcher only)"
frontmatter_status_transitions:
  spec.md: in-progress -> completed
carried_debt:
  - internal/homestate coverage 76.2% < 85% (69.0% at base 04ca1a98c)
  - D7, D10, D11, D12 (see §E.2 / §E.3)
  - push-gate no-fetch test is skipped on Windows (POSIX-only wrapper)
  - record-unavailable.jsonl lock serializes only moai writers
sync_audit: "Tier L sync-audit deferred until the Opus limit resets (lead decision; model not lowered)"
```
