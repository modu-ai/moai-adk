# progress.md — SPEC-FACTORY-RUN-RETIRE-001 (card t1107)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-23
- artifacts: spec.md, plan.md, acceptance.md, this skeleton (Tier M)
- baseline: worktree `.claude/worktrees/t1107`, branch `WT-factory-run-retire`, base local develop `176d8b658`
- evidence base: `.moai/reports/t1107/verdict.md` (reproduction, darwin, 2026-09-23)
- plan_audit: iter-1 FAIL 0.84 (`bb5b8f9d1`) → D1-D6 revision `c1ae8ff5e` → iter-2 FAIL 0.84
  (pinned to `c1ae8ff5e`; R6 PASS, R2 FAIL on D11/D14) → iter-3 revision (this commit). iter-3 is
  the last audit iteration.

### Process incident — two writers in an open audit window (2026-09-23)

Recorded at the lead's request, in the lead's words:

> While plan-audit iter-2 was reading `c1ae8ff5e`, commit `320cdeb90` landed on the same worktree,
> putting two writers in an open audit window. The fault was the lane's sequencing: the audit was
> opened without first telling the SPEC author to hold, and the author was working through an
> addendum the lane had itself sent. No content was damaged — the working tree stayed clean and the
> commits are linear. What was damaged was attribution, and it was repaired by pinning the iter-2
> verdict explicitly to `c1ae8ff5e` and recording `320cdeb90` as an unexamined successor rather
> than switching trees mid-audit. Corrective: the lane tells the author to hold before opening an
> audit window, and the author asks before committing when unsure whether one is open.

### v0.13.0 in-place amendment — plan phase (card t1169, 2026-09-25)

- baseline: worktree `.claude/worktrees/t1169`, branch `WT-retire-boot-proof-spec`, base develop
  `a0b78213d`, which descends from `372c1bb0b` (the t1168 boot-proof merge).
- why: card t1168 shipped a boot proof that retires identity-less legacy runs as `dead`; the
  literal REQ-006 required `indeterminate` for them. Evidence and the residual risks that motivate
  this card: `.moai/reports/t1168/verdict.md` (local evidence file, primary checkout), Residual-risk.
- amended: REQ-006 rewritten; REQ-006b (boot-proof premises), REQ-006c (boot-time reader: error
  cause, long `intr` line), REQ-010b (`basis` = `stamp` / `peer` / `boot` on `run.retired`) added;
  spec.md §C.3 and two §F items added; acceptance.md AC-018 (regression-guard), AC-019, AC-020 with
  RED-now cells R-09 / R-10 pinned at `a0b78213d`; plan.md M7.
- status: `completed → in-progress` (amendment), `amendment_of:` self, Amendments record citing the
  prior close `85414b3e6`.
- run phase owes: M7 in plan.md. The t1107 §E.2-§E.4 evidence below is the prior close's and is
  left untouched; the amendment's run evidence is manager-develop's to add.
- Implementation Kickoff Approval (run entry): granted by the operator in the lane window on
  2026-09-25. Operator's answer, verbatim: "run, sync 모두 진행". Plan-audit standing at kickoff:
  iter-2 PASS-WITH-DEBT (`.moai/reports/t1169/plan-audit-iter2.md`), debts N1 / N2 carried into the
  run delegation and resolved there — see §E.2.6.

## §E.2 Run-phase Evidence

Tree: branch `WT-factory-run-retire`, worktree `.claude/worktrees/t1107`, base `9b1805a67`.
Host: darwin/arm64, go1.26.8. cycle_type=tdd. All evidence below is from **this run, this tree**.
Raw captures live in `.moai/reports/t1107/run/`.

### E.2.0 Pre-flight (plan.md §C)

| Check | Command | Observed |
|---|---|---|
| sibling-card sets disjoint | `ls internal/cli/factory_lane_handoff* internal/factorymsg/handoff* internal/cli/mcp_codex.go internal/hook/factory_messages*` | `no matches found`, exit 1 — the t1082 / t1109 files do not exist in this tree (so this tree cannot establish non-overlap; the lead's t1082-worktree measurement stands as cited) |
| no retirement path exists | `grep -rn "UPDATE runs\|DELETE FROM runs" --include='*.go' . \| grep -v _test.go` | no output, exit 1 |
| instrument control | `grep -rn "status='active'" --include='*.go' . \| grep -v _test.go` | 2 hits (`internal/factorymsg/store.go:257`, `internal/homestate/runtime.go:25`), exit 0 — the search shape works, so the zero above is an absence |

### E.2.1 RED evidence (TDD invariant i — captured BEFORE GREEN)

| Layer | File | Verbatim RED |
|---|---|---|
| homestate | `.moai/reports/t1107/run/red-homestate.txt` | `undefined: OwnerClassification` / `unknown field LeadPID in struct literal of type FactoryRun` / `db.ReconcileActiveRuns undefined` … `FAIL github.com/modu-ai/moai-adk/internal/homestate [build failed]` |
| factorymsg | `.moai/reports/t1107/run/red-factorymsg.txt` | `--- FAIL: TestResolveActiveRunRetiresDeadOwnerAndJoins` (`resolve: AMBIGUOUS_FACTORY`), `--- FAIL: TestResolveActiveRunBothOwnersDead` (`err = AMBIGUOUS_FACTORY, want NO_ACTIVE_FACTORY`), `--- FAIL: TestResolveActiveRunAmbiguityNamesClassifications` (`error "AMBIGUOUS_FACTORY" does not contain "run-live"`), `--- FAIL: TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback` |
| cli (seam) | `.moai/reports/t1107/run/red-cli-owner.txt` | `undefined: stampFactoryRunOwner`, then behaviourally: all three `required/*` subtests of `TestRestampSeamIsCalledAtEveryNonReplaceCallSite` FAIL, and `--- FAIL: TestPaneDoorRefusalLeavesNoLauncherStampedRun` (`run row still carries the launching process's pid 97122`) |
| cli (operator) | `.moai/reports/t1107/run/red-cli-runs.txt` | `factory runs: unknown command "runs" for "factory"` |

### E.2.2 AC matrix

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-001 | PASS | `go test ./internal/homestate/ -run TestRecordRunStampsSessionOwnerIdentity -v` | `--- PASS: TestRecordRunStampsSessionOwnerIdentity (0.47s)` — owner stamp `(4242, "1700000000.000001")`, `schema_version = 3` |
| AC-002 | PASS | `go test ./internal/homestate/ -run TestRetireRunPreservesRowAndAppendsEvent -v` | `--- PASS` — row count 1 after retirement, `status='retired'`, one `run.retired` event |
| AC-003 | PASS | `go test ./internal/homestate/ -run TestClassifyOwnerUsesFingerprintNotBarePID -v` | `--- PASS` — matching fingerprint → live; live pid + differing fingerprint → dead; REQ-003b direction asserted explicitly |
| AC-004 | PASS (SPEC defect flagged — see E.2.5) | `sh .moai/reports/t1107/run/doors-run.sh worker-join -- cc -f worker-1` | `exit=0`; `tltbl4 retired` (dead owner 68247), `tltbl8 active` (live pane owner 68638). Unit cover: `--- PASS: TestResolveActiveRunRetiresDeadOwnerAndJoins`, `--- PASS: TestResolveActiveRunBothOwnersDead` |
| AC-005 | PASS | same worker-join run above + `go test ./internal/homestate/ -run TestReconcileRetiresDeadAndLeavesLive -v` | live owner held open across the measurement (tmux pane 68638, command sleeps 600s) stayed `active`; dead-owner run `retired`. `--- PASS` |
| AC-006 | PASS | `go test ./internal/homestate/ -run TestReconcileLeavesIndeterminateActive -v` | `--- PASS` — unprobeable owner stays `active`, reported `indeterminate` |
| AC-007 | PASS | `go test ./internal/factorymsg/ -run TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback -v` + `go test ./internal/homestate/ -run TestMigrateFactoryV2ToV3PreservesRows -v` | `--- PASS` both — a seeded schema-v2 DB with two unstamped `active` rows and dead `role='lead'` peers migrates to v3 and both rows retire through the peer fallback; the ambiguity is gone |
| AC-008 | PASS | `go test ./internal/factorymsg/ -run TestResolveActiveRunAmbiguityNamesClassifications -v` | `--- PASS` — error text contains `AMBIGUOUS_FACTORY`, both surviving run ids, and each classification |
| AC-009 | PASS (regression guard — green before and after) | `go test ./internal/factorymsg/ -run TestResolveActiveRunPreservesFailClosedSentinels -v` | `--- PASS` — zero active → `NO_ACTIVE_FACTORY`; two live-owner runs → `AMBIGUOUS_FACTORY`; neither returns a run id |
| AC-010 | PASS | `sh doors-run.sh factory-runs -- factory runs` and `… -- factory runs --retire tltbl8` | listing shows `tltbl4 retired dead`, `tltbl8 active live`, exit 0; `--retire tltbl8` → `Factory run owner is not dead: run tltbl8 owner classified live.`, `exit=1`, run left `active`. Indeterminate leg covered by `--- PASS: TestRetireRunIfDeadRefusesLiveAndIndeterminate` and `--- PASS: TestFactoryRunsCommandReportsAndRefuses` |
| AC-011 | PASS | four executed door invocations, `.moai/reports/t1107/run/doors-evidence.md` §1 | `moai cc -f` → `tltb75` row (91883 / 1790158577.527848) == lead peer; `moai glm -f` → `tltb7r` (95054 / 1790158599.696629) == peer; `moai codex -f` → `tltb98` (5309 / 1790158652.056659) == peer; `moai codex -f --spawn` → `tltbl8` (68638 / 1790159084.722334) == peer, and tmux reports `PANE %1 pane_pid=68638` |
| AC-012 | PASS-WITH-GAP | `go test ./internal/cli/ -run TestPaneDoorRefusalLeavesNoLauncherStampedRun -v` | `--- PASS` — the production `defaultCodexSpawnLaunch` refusal branch runs, the pane is cleaned up, and no `runs` row carries the launcher's pid. **Gap**: the resolver's real 2s deadline was not exhausted by a live tmux pane (`#{pane_pid}` names the pane's shell, alive before the exec'd command returns) — the seam was forced to its exhaustion return instead. `doors-evidence.md` §4 |
| AC-013 leg 1 | PASS | `go test -tags=integration -v ./test/integration/harness/ -run TestFactoryRunRetire` | `--- PASS: TestFactoryRunRetire (0.58s)` + `ok github.com/modu-ai/moai-adk/test/integration/harness 0.910s` — the `--- PASS:` line is present, so the selector actually selected the test |
| AC-013 leg 2 | **PENDING** (not a pass) | — | No CI run exists for a card branch before it merges: `ci.yml` triggers on `push: [main, develop]` and `pull_request: [main]`, and this project does not push `WT-` branches. The ubuntu / macos / windows conclusions land on the develop push after integration and are to be recorded by run id. |
| AC-014 | PASS | `find /private/tmp/t1107-doors/moaihome -name '*.db'` + `ls ~/.moai/db \| grep moai-adk-go` | all created state under `…/db/proj-c6f77cbe/factory/…`; this repository's key is `moai-adk-go-1bd3d038`, a different directory the exercise never opened. Every Go test uses a project root under `t.TempDir()` with `HOME`/`MOAI_HOME`/`MOAI_CLAUDE_BIN` scrubbed |
| AC-015a | PASS | mutant B (`retirable` → `true`), `.moai/reports/t1107/run/mutant-b-no-guard.txt` | 5 tests FAIL, including `--- FAIL: TestReconcileNeverRetiresLiveOrIndeterminate` and `--- FAIL: TestRetireRunIfDeadRefusesLiveAndIndeterminate` |
| AC-015b | PASS | mutant C (`retirable` → `false`), `.moai/reports/t1107/run/mutant-c-no-retire.txt` | 4 tests FAIL, including `--- FAIL: TestRetireRunPreservesRowAndAppendsEvent` and `--- FAIL: TestReconcileRetiresDeadAndLeavesLive` |
| AC-016 leg 1 | PASS | `go test ./internal/cli/ -run TestRunOwnerAndLeadPeerNameOneProcessOnEveryShape -v` | `--- PASS` on all three subtests (replace / spawn / pane), driven through the build-tag-free seam with fixture identities; plus the four executed doors under AC-011, where row and peer agree on every one |
| AC-016 leg 2 | PASS | `go test ./internal/cli/ -run TestRestampSeamIsCalledAtEveryNonReplaceCallSite -v`, plus three per-site mutants | All five subtests PASS. Per-site mutation, probed **individually**: mutant D (delete the call in `launch_exec_windows.go`) → only `required/launch_exec_windows.go` FAILs; mutant E (`codex_direct_windows.go`) → only that site FAILs; mutant F (`codex_launcher.go`) → only that site FAILs. Files: `mutant-d-*`, `mutant-e-*`, `mutant-f-*` |
| AC-017 | PASS | mutant A (reject-list `retirable`), `.moai/reports/t1107/run/mutant-a-rejectlist.txt` | The required divergence, observed: `--- PASS: TestReconcileRetiresDeadAndLeavesLive` (AC-005), `--- PASS: TestReconcileLeavesIndeterminateActive` (AC-006), `--- PASS: TestRetireRunIfDeadRefusesLiveAndIndeterminate` and `--- PASS: TestFactoryRunsCommandReportsAndRefuses` (AC-010) — while `--- FAIL: TestEveryRetirementPathDeclinesUnenumeratedClassification` fails on **all three** paths (reconciler / migration pass / operator retire). Restored to the positive form and re-verified green |

### E.2.3 Cross-cutting verification

| Item | Command | Observed |
|---|---|---|
| Cross-platform build | `go build ./...` then `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 (`both builds ok`) |
| Affected-package tests | `go test ./internal/homestate/... ./internal/factorymsg/...` | `ok …/internal/homestate 24.254s coverage: 67.1%` · `ok …/internal/factorymsg 11.920s coverage: 70.3%` |
| `internal/cli` suite | `go test ./internal/cli/... -timeout 30m` | `FAIL …/internal/cli 1801.359s` — exactly one test, `TestFactoryOperationalFixtureUsesProductionInit`, with `factory messaging degraded: context deadline exceeded`. Re-run in isolation on the same tree: `--- PASS: TestFactoryOperationalFixtureUsesProductionInit (6.81s)`. Every `internal/cli/*` sub-package `ok`. Load-sensitive, not attributed to this change; CI supplies the clean-environment verdict |
| Lint | `golangci-lint run --timeout=10m ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...` | `0 issues.` exit 0 |
| Scope | `git status --short` | only this SPEC's declared file set; no t1082 file (`factory_lane_handoff*`, `handoff*`, `mcp_codex.go`, `defaults.go`) and no t1109 file (`hook/factory_messages*`) touched |
| Sentinels preserved | `TestResolveActiveRunPreservesFailClosedSentinels` | `AMBIGUOUS_FACTORY` and `NO_ACTIVE_FACTORY` still literal substrings of the failure text |

### E.2.4 Files changed

Modified: `internal/homestate/factory.go`, `internal/homestate/runtime.go`,
`internal/factorymsg/store.go`, `internal/cli/factory.go`,
`internal/cli/factory_handoff_recover.go`, `internal/cli/launch_exec_windows.go`,
`internal/cli/codex_direct_windows.go`, `internal/cli/codex_launcher.go`,
`internal/homestate/handoff_lease_test.go`.

Added: `internal/homestate/factory_run_retire.go`, `internal/factorymsg/factory_run_retire.go`,
`internal/cli/factory_run_owner.go`, `test/integration/harness/it08_factory_run_retire_test.go`,
plus the four package test files.

One pre-existing test was **adjusted, not deleted**: `TestFactoryV1ClaimedRowsUpgradeToV2` asserted
the v1 chain terminates at schema version `"2"`; it now asserts `"3"`, because the chain is
v1→v2→v3. Its subject (the claimed-row upgrade) is unchanged.

### E.2.5 Two findings the run phase surfaced

**F1 — AC-004's Given contradicts its own Then (SPEC defect).** The criterion reads "two `active`
runs whose owners are **both dead** … Then the join succeeds, and the `runs` table afterwards shows
the run it joined as the sole `active` row with the other transitioned to `retired`." Those cannot
both hold: reconciliation retires **every** dead owner, so two dead owners leave zero active runs
and resolution correctly fails closed with `NO_ACTIVE_FACTORY`. The Then describes the shape the
reproduction actually produced — one stale run plus one live one. Both readings are implemented and
recorded rather than one being chosen silently: `TestResolveActiveRunRetiresDeadOwnerAndJoins`
covers the Then (and is the shape the executed worker-join demonstrates), and
`TestResolveActiveRunBothOwnersDead` covers the Given followed to its actual consequence. **No
requirement was changed**; this is reported for manager-spec to reconcile.

**F2 — a migration hazard found by an existing test, and fixed.** `factoryDDL` runs **before** the
schema-version check, so on a database whose `runs` table did not yet exist the DDL creates it
already carrying the v3 columns, and a blind `ALTER TABLE … ADD COLUMN` then fails with
`duplicate column name: lead_pid`. Caught by `TestFactoryV1ClaimedRowsUpgradeToV2`.
`migrateFactoryV2ToV3` now reads `PRAGMA table_info(runs)` and adds only the missing columns.

### E.2.6 v0.13.x amendment run (card t1169, plan.md M7)

Base: branch `WT-retire-boot-proof-spec` at `e5f020cc3` (its `internal/` tree equals `a0b78213d`'s).
Host: darwin. Every command below ran in this worktree against the tree at the commit named.

**Commits.** RED `09c99c071` (tests plus a signature-only `procStatBootTime` stub) → GREEN
`1610e0ee7` (implementation) → this evidence commit. The RED commit precedes the GREEN commit in the
graph, so the ordering is witnessed by git rather than asserted (verification-claim-integrity §2.3).

**RED (on `09c99c071`)** — `go test ./internal/homestate/ -run 'ProcStat|TestRetiredEventRecordsProofBasis|TestBootProofDeclinesWithoutEveryPremise' -count=1`, exit 1:

```
--- FAIL: TestProcStatBootTimeFindsBtimeAfterLongLine (0.00s)
    boot_time_procstat_test.go:20: procStatBootTime: procfs stat carries no btime record, want the btime after the 262149-byte intr line
--- FAIL: TestProcStatBootTimeReportsReadError (0.00s)
    boot_time_procstat_test.go:38: cause = procfs stat carries no btime record, want the read error procfs read failed mid-stream
--- FAIL: TestProcStatBootTimeRejectsMalformedBtime (0.00s)
--- FAIL: TestProcStatBootTimeAcceptsUnterminatedFinalLine (0.00s)
--- FAIL: TestRetiredEventRecordsProofBasis (0.01s)
        factory_run_boot_proof_test.go:268: run-stamp basis = "", want "stamp" (payload map[classification:dead])
        factory_run_boot_proof_test.go:268: run-peer basis = "", want "peer" (payload map[classification:dead])
        factory_run_boot_proof_test.go:268: run-boot basis = "", want "boot" (payload map[classification:dead])
        factory_run_boot_proof_test.go:282: payload = map[classification:dead], want classification dead and basis boot
FAIL	github.com/modu-ai/moai-adk/internal/homestate	0.479s
```

`TestProcStatBootTimeWithoutBtimeIsUnavailable` passed on the stub, which returns "no btime record":
leg (iii) discriminates only together with leg (ii), which requires the two causes to differ. The
AC-018 legs passed on RED by design — they are regression guards (`go test ./internal/factorymsg/
-run 'PartialStamp|LeadRecordAbsentFor' -count=1 -v` → four `--- PASS` lines, `ok … 7.712s`). Their
discriminating evidence is the mutant table below.

**GREEN (on `1610e0ee7`).**

| Command | Exit | Output tail |
|---|---|---|
| `go test ./internal/homestate/... ./internal/factorymsg/... -count=1` | 0 | `ok …/internal/homestate 17.806s` / `ok …/internal/factorymsg 50.592s` |
| `go test ./internal/cli/ -run 'FactoryRuns\|Abandon\|LaneHandoffRecover\|EnterSelectedFactoryRun\|ResolveActiveRun' -count=1` | 0 | `ok …/internal/cli 20.835s` — the selector matches 4 tests (`go test -list` with the same pattern prints 4), so this is not a zero-match green |
| `go vet ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/` | 0 | (no output) |
| `golangci-lint run ./internal/homestate/... ./internal/factorymsg/...` | 0 | `0 issues.` |
| `GOOS=linux go build ./internal/homestate/ ./internal/factorymsg/` | 0 | (no output) |
| `GOOS=windows go build ./internal/homestate/ ./internal/factorymsg/` | 0 | (no output) |
| `GOOS=linux go vet ./internal/homestate/` | 0 | (no output) |
| `GOOS=windows go vet ./internal/homestate/` | 0 | (no output) |

**RED-now cells re-run on `1610e0ee7`.** R-09 `grep -rl btime internal/homestate --include='*_test.go'`
→ `internal/homestate/boot_time_procstat_test.go`, exit 0 (was empty, exit 1). R-10
`grep -rl --exclude='*_test.go' '"basis"' internal/homestate` → `internal/homestate/factory_run_retire.go`,
exit 0 (was empty, exit 1).

**Mutant probes.** Each mutant changed one production site; then
`go test ./internal/homestate/ -count=1` and `go test ./internal/factorymsg/ -count=1 -run
'BootProof|PreBoot|PartialStamp|LeadRecordAbsentFor|ResolveActiveRun'` ran; then the original file
bytes were written back. Every mutant edit carried the marker `T1169-MUTANT`. After the run:
`grep -rn T1169-MUTANT internal/` → no output, exit 1; the worktree status was empty; and the suite
re-ran green (`ok homestate 14.698s`, `ok factorymsg 45.912s`). The generic `grep -rn MUTANT
internal/` count is 11 both before and after — pre-existing comments in other packages. Per-mutant
logs: `.moai/reports/t1169/mutant-M*.log` (local evidence, not committed); the failing test names are
copied here.

| Mutant | Edit | Failing tests |
|---|---|---|
| M1 proof always declines | `predatesBoot` returns `false` first | TestReconcileRetiresIdentitylessRunsThatPredateBoot, TestRetireRunIfDeadAcceptsBootProof, TestResolveActiveRunReapsPreBootIdentitylessRuns, TestResolveActiveRunPreBootRowsAloneFailClosedAsNoActive, TestPartialStampWithoutBrokerIsBootProven, TestRetiredEventRecordsProofBasis/{reconciler, operator_retire_of_a_boot-proven_run}, and — through their positive controls — TestPartialStampWithBrokerStaysIndeterminate, TestLeadRecordAbsentFor{TreatsStatErrorAsPossibleRecord, RejectsUnderivableBrokerPath}. **Stayed green:** TestBootProofNeverOverridesAnIdentity (the identity-precedence leg) |
| M2 lead-record premise dropped | `\|\| !opts.LeadRecordAbsent(runID)` → `\|\| false` | TestBootProofDeclinesWithoutEveryPremise/lead_record_may_exist, TestResolveActiveRunBootProofDeclinesWhenBrokerExists, TestPartialStampWithBrokerStaysIndeterminate, TestLeadRecordAbsentForTreatsStatErrorAsPossibleRecord, TestLeadRecordAbsentForRejectsUnderivableBrokerPath |
| M3 post-boot activity ignored | `perr != nil \|\| !at.Before(boot)` → `perr != nil` | TestBootProofDeclinesWithoutEveryPremise/{event_after_boot, worker_heartbeat_after_boot, card_updated_after_boot, run_row_touched_after_boot, timestamp_equal_to_boot}, TestResolveActiveRunAmbiguityNamesClassifications |
| M4 strictly-earlier → not-later | `!at.Before(boot)` → `at.After(boot)` | TestBootProofDeclinesWithoutEveryPremise/timestamp_equal_to_boot |
| M5a stat error as absence | `errors.Is(err, fs.ErrNotExist)` → `err != nil && !errors.Is(err, fs.ErrExist)` | TestLeadRecordAbsentForTreatsStatErrorAsPossibleRecord |
| M5b underivable path as absence | the `BrokerPath` error branch returns `true` | TestLeadRecordAbsentForRejectsUnderivableBrokerPath |
| M6 partial stamp routed to peer | fallback also taken when `start` is empty | TestPartialStampWithBrokerStaysIndeterminate |
| M19a read error discarded | the read-error branch returns `errProcStatNoBtime` | TestProcStatBootTimeReportsReadError |
| M19b default 64 KiB scanner | the seam rewritten over `bufio.Scanner` | TestProcStatBootTimeFindsBtimeAfterLongLine |
| M20a constant basis `stamp` | the payload writes `BasisStamp` | TestRetiredEventRecordsProofBasis/{reconciler, operator_retire_of_a_boot-proven_run} |
| M20b operator path constant `stamp` | `RetireRunIfDead` passes `BasisStamp` | TestRetiredEventRecordsProofBasis/operator_retire_of_a_boot-proven_run only — `/reconciler` stayed green |
| M20c `basis` key omitted | payload `{"classification":…}` only | TestRetiredEventRecordsProofBasis/{reconciler, operator_retire_of_a_boot-proven_run} |
| M20d `classification` key renamed | payload key `state` | TestRetiredEventRecordsProofBasis/{reconciler, operator_retire_of_a_boot-proven_run} |

The first forms of M3 and M5a did not compile (an unused variable; unused imports) and are not
counted. Both were re-expressed as compiling edits and re-run
(`.moai/reports/t1169/mutants-summary-rerun.log`); the table carries the re-run.

**AC standing for the amendment.**

| AC | Standing | Basis |
|---|---|---|
| AC-018 | regression-guard: green on the run tree; mutants 1-6 (5 as 5a and 5b) each observed red and restored | mutant table |
| AC-019 | PASS | RED on `09c99c071`, GREEN on `1610e0ee7`, M19a and M19b red |
| AC-020 | PASS | RED on `09c99c071`, GREEN on `1610e0ee7`, M20a-d red; M20b isolates the operator leg |

**Plan-audit debts carried into this run.** N1:
`TestLeadRecordAbsentFor{TreatsStatErrorAsPossibleRecord, RejectsUnderivableBrokerPath}` assert the
function's `false` AND the end-to-end reconcile outcome (`active` plus `OwnerIndeterminate`, read from
`Reconciliation.Remaining`), each beside a positive-control run that the same options retire `dead`.
N2: `TestPartialStampWithBrokerStaysIndeterminate` seeds the `role='lead'` peer with a dead PID and a
non-empty `process_start` (the test fails outright if that fingerprint is empty) and dates every run
timestamp before boot; its positive control retires, so premise 2 is the only declining premise, and
M6 turns the leg red. Residual-risk note: the `RecordRun` comment in `internal/homestate/runtime.go`
now names the REQ-006b boot proof instead of "indeterminate forever".

**Files changed** (`git diff --stat a0b78213d HEAD -- internal`): 7 files, +463 / −27 —
`internal/homestate/{boot_time_procstat.go (new), boot_time_procstat_test.go (new), boot_time_unix.go,
factory_run_boot_proof_test.go, factory_run_retire.go, runtime.go}` and
`internal/factorymsg/factory_run_boot_proof_legs_test.go (new)`. No `internal/cli` change. No
acceptance.md change, so the AC snapshot is not regenerated.

**Gaps.** The factorymsg boot-proof legs read the host's real boot time and skip where the host
reports none; on this darwin host they ran. The linux `platformBootTime` over a live `/proc/stat` and
the windows reader were cross-built and vetted, not executed — they land on the post-merge three-OS
CI run (spec.md §F). The full repository suite was not run locally, by design. No independent
reviewer has read this run yet.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-23
run_commit_sha: eaa3322a1   # backfilled by card t1146 (sync-audit S1); the run-phase commit, an ancestor of this SPEC's branch
run_status: implemented
ac_pass_count: 16          # AC-001..AC-012, AC-013 leg 1, AC-014, AC-015a/b, AC-016 (both legs), AC-017
ac_fail_count: 0
ac_pending_count: 1        # AC-013 leg 2 — post-merge, non-gating, recorded pending with its reason
ac_pass_with_gap_count: 1  # AC-012 — see §E.2.2
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (no push in this phase)
l44_post_push_fetch: n/a (no push in this phase)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64_cross_compile: pass
  linux: deferred to the develop-push CI run
total_run_phase_files: 17
m1_to_mN_commit_strategy: single run-phase commit covering M1-M6
```

### E.3.1 Run-phase Audit-Ready Signal — v0.13.x amendment (card t1169)

```yaml
run_complete_at: 2026-09-25
run_commit_sha: 1610e0ee7   # GREEN implementation commit; RED tests at 09c99c071
run_status: implemented
ac_pass_count: 2              # AC-019, AC-020
ac_regression_guard_count: 1  # AC-018 — mutants 1-6 observed red and restored, never a bare pass
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (no push in this phase)
l44_post_push_fetch: n/a (no push in this phase)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass (tests executed)
  linux_cross_build_and_vet: pass
  windows_cross_build_and_vet: pass
total_run_phase_files: 7
m1_to_mN_commit_strategy: M7 as a RED commit then a GREEN commit, then this evidence commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-24
sync_commit_sha: 85414b3e6   # a commit cannot cite its own hash; backfilled by card t1169
sync_status: audit-ready
changelog_entry_position: "CHANGELOG.md `## [Unreleased]` → `### Fixed`, first entry (the defect and the retirement mechanism) AND `### Added`, first entry (the `moai factory runs` operator surface)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed"    # merged into this single sync commit
  plan_md: "n/a - no frontmatter block"
  acceptance_md: "n/a - no frontmatter block"
  progress_md: "n/a - no frontmatter block"
  updated_field: "2026-09-23 -> 2026-09-24 (spec.md only)"
b12_self_test_a: "grep -c 'SPEC-FACTORY-RUN-RETIRE-001' CHANGELOG.md -> 0 before emission (clean; no duplicate from a parallel BATCH-SYNC session)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 17 (AC-001..AC-017), non-zero so not a vacuous match; matches the 17-row matrix in acceptance.md SS-C.2 and the 'all 17 criteria' statement in SS-D.2"
b12_self_test_c: "every path cited in the two CHANGELOG entries verified present via ls before commit - 11 paths: internal/homestate/{factory_run_retire.go,factory.go,runtime.go}, internal/factorymsg/{factory_run_retire.go,store.go}, internal/cli/{factory_run_owner.go,factory.go,factory_handoff_recover.go,codex_launcher.go,launch_exec_windows.go,codex_direct_windows.go}, test/integration/harness/it08_factory_run_retire_test.go"
canary_compliance_check: "n/a - this SPEC defines no forward-looking policy that its own sync would test"

docs_synchronised:
  changelog: "2 entries (Fixed + Added), both at the top of their section"
  readme_4_locale: "README.md / README.ko.md / README.ja.md / README.zh.md - one sentence appended to the factory-mode paragraph in each, recording the self-healing run record and the moai factory runs operator surface; no template mirror exists for README, so no Template-First rebuild is owed"
  docs_site: "not changed - measured, not assumed: grep -rn 'moai factory' docs-site/ returns 0 lines, so the command group has no page to synchronise. Authoring a new 4-locale reference page for an already-undocumented command group is a follow-up documentation card, not sync-phase drift repair (sync-audit S3)"
  codemaps: "no restamp owed - moai graph check: codemaps described-source-diff value=28 threshold=40 verdict=fresh; citations fresh; exit 0"

ac_final_standing:
  pass: 16                 # AC-001..AC-012, AC-013 leg 1, AC-014, AC-015a/b, AC-016 (both legs), AC-017
  fail: 0
  pending: 1               # AC-013 leg 2 - carried forward UNCHANGED, never upgraded to a pass
  pass_with_gap: 1         # AC-012 - carried forward UNCHANGED with its gap stated
  ac_013_leg2: "PENDING. No CI run exists for a card branch before it merges: ci.yml triggers on push [main, develop] and pull_request [main], and this project does not push WT- branches. The ubuntu / macos / windows conclusions land on the develop push that follows integration and are to be confirmed by run id."
  ac_012: "PASS-WITH-GAP. The production pane-door refusal branch executed and left no launcher-stamped run, but the resolver's real 2-second deadline was not exhausted by a live tmux pane - #{pane_pid} names the pane's shell, alive before the exec'd command returns - so the seam was forced to its exhaustion return instead."
  retained_debt: "D20 and D21 (spec.md SS-G) remain RETAINED with their evidence and re-opening conditions; neither is closed by this sync."

sync_phase_verification:
  tests: "go test ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/... (env-scrubbed, single compound invocation) -> ok homestate 27.400s, ok factorymsg 14.090s, FAIL internal/cli 1639.462s with exactly one failing test, every internal/cli/* sub-package ok. Capture: .moai/reports/t1107/sync/test-affected.txt"
  the_one_failure: "TestFactoryOperationalFixtureUsesProductionInit. Re-measured with go test ./internal/cli/ -run TestFactoryOperationalFixtureUsesProductionInit -count=3 -v -> three verbatim '--- PASS:' lines (5.65s / 6.34s / 7.37s), exit 0. Capture: .moai/reports/t1107/sync/test-flake-retry-v.txt. Load-sensitive: the assertion is governed by a 200ms deadline (factoryHookInspectionDeadline, internal/hook/factory_messages.go:20); it failed at host load average 24.89 with a peer session running go test ./internal/hook/, and passed 3/3 at 16.00-21.28. NOT established: a measurement on an idle host, or one at the merge base - see sync-audit SS-6."
  env_scrub_finding: "An unscrubbed run failed TestCodexSpawn_RealAssemblyThroughStubTmux because this session's MOAI_FACTORY_WORKER / MOAI_FACTORY_WORKERS / MOAI_KANBAN_BACKEND leaked into the assembled tmux command. With the scrub the failure is absent - an artifact of the measuring session, not a defect."
  lint: "golangci-lint run --timeout=15m ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/... -> '0 issues.', exit 0. Capture: .moai/reports/t1107/sync/lint.txt"
  spec_lint: "moai spec lint SPEC-FACTORY-RUN-RETIRE-001 -> 'No findings - all SPEC documents are valid', exit 0. Capture: .moai/reports/t1107/sync/spec-lint.txt"
  spec_audit: "mcp__moai__spec_audit(project_root=<this worktree>, filter_spec=SPEC-FACTORY-RUN-RETIRE-001) -> total_specs 1, modern_era_clean 1, one INFO EraAutoDetected finding; no drift"
  scope: "git diff --name-only 176d8b658 HEAD -> 23 paths (6 SPEC artifacts + 17 code/test files, matching SS-E.3 total_run_phase_files: 17). Zero hits for the t1082 set and zero for the t1109 set."
  full_suite: "NOT run locally by design (parallel lanes doing so drove this machine to load 413 on 2026-08-15). CI supplies the full-suite and cross-platform verdict."

sync_audit:
  report: ".moai/reports/t1107/sync/sync-audit.md"
  independence: "SELF-AUDIT by the sync agent, not an independent sync-auditor spawn - this agent carries no Agent tool and cannot spawn one. The report is evidence for the lead to read; the binding verdict is the lead's."
  proposed_verdict: "PASS-WITH-DEBT, harmonic mean 0.93 (Functionality 0.93 / Security 0.95 / Craft 0.94 / Consistency 0.90), zero blocking findings"
  findings:
    S1: "HAND-BACK. progress.md SS-E.3 still carries run_commit_sha: pending-backfill-run. The value is knowable (the run commit is eaa3322a1) but SS-E.3 is manager-develop's artifact, so it was left unchanged rather than edited across an ownership boundary."
    S2: "HAND-BACK, low severity. AC-007's Then ('the join no longer fails with AMBIGUOUS_FACTORY') is literally true but invites the same misreading AC-004 was repaired for: both peers dead means both rows retire, so zero active rows remain and the join fails closed with NO_ACTIVE_FACTORY rather than succeeding. Implementation and run-phase evidence are correct; only the wording is weak. spec.md / acceptance.md body text is manager-spec's."
    S3: "Follow-up card recommended. grep -rn 'moai factory' docs-site/ -> 0 lines; grep -c per README locale -> 0 (pre-edit). The command group has no page to synchronise, so a new 4-locale reference page is a documentation card rather than sync-phase drift repair. The narrower real gap - the user-visible behaviour change - was closed in all four READMEs."
    S4: "Observation only. retirable is called at exactly two non-test sites (factory_run_retire.go:175 and :207); AC-017's 'migration pass' leg reaches :175 as well, differing only by supplying a Fallback. Two of the three legs therefore share one call site. This is REQ-005's stated intent working, not a defect, but the SS-E.2.2 wording reads as three independent guards."
  ac_given_then_satisfiability: "The additional checklist item the lead asked for was applied to all 17 criteria: each Given's premises were executed to their consequence under the SPEC's own rules and checked against its own Then. 17 of 17 satisfiable. AC-004's repair at ae9443156 holds. One residue: AC-007 (S2 above). Rationale recorded in the sync-audit report SS-5 - form review (mutation direction, mechanical checkability, vacuous-green shapes) cannot see a well-formed criterion that contradicts itself, which is why AC-004 survived three Tier M iterations and a full Tier L audit."

not_done:
  - "No push. No PR. The branch is integrated by the lead; this worktree holds the only copy of the work."
  - "S1 and S2 are handed back rather than fixed - both require editing an artifact this agent does not own."
  - "No docs-site page authored (S3)."
```

### E.4.1 Sync-phase Audit-Ready Signal — v0.13.x amendment (card t1169)

```yaml
sync_complete_at: 2026-09-25
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash; the lead reads it off the sync commit
sync_status: audit-ready
changelog_entry_position: "CHANGELOG.md `## [Unreleased]` -> `### Fixed`, new entry (the REQ-006 boot-proof correction and the `basis` field), placed directly after the t1107 SPEC-FACTORY-RUN-RETIRE-001 Fixed entry it amends"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"    # merged into this single sync commit; version stays 0.13.1, no body text touched
  plan_md: "n/a - no frontmatter block"
  acceptance_md: "n/a - no frontmatter block"
  progress_md: "n/a - no frontmatter block"
  updated_field: "unchanged - spec.md `updated:` was already 2026-09-25 before this commit"
b12_self_test_a: "grep -c 'SPEC-FACTORY-RUN-RETIRE-001' CHANGELOG.md -> 2 before emission (the two t1107 entries at lines 27 and 84; a fresh grep after this commit finds 3, confirming the new entry is additive, not a duplicate of either existing one)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 20 (AC-001..AC-020), non-zero so not a vacuous match; this amendment's own entry cites AC-018/AC-019/AC-020 specifically, the three criteria acceptance.md SS-C.2 attributes to M7"
b12_self_test_c: "every path cited in the new CHANGELOG entry verified present via ls before commit - internal/homestate/{boot_time_procstat.go,boot_time_unix.go,factory_run_retire.go,runtime.go}, internal/factorymsg/factory_run_boot_proof_legs_test.go"
canary_compliance_check: "n/a - this SPEC defines no forward-looking policy that its own sync would test"

docs_synchronised:
  changelog: "1 new entry under Fixed, immediately after the t1107 entry it amends"
  readme_4_locale: "not changed - the README factory-mode paragraph already describes run retirement at the behavioural level (\"a run whose lead has died is retired automatically\"); it names no owner-classification mechanism (identity stamp vs REQ-006b boot proof) and no event-payload field, so the boot-proof correction and the `basis` key add nothing the README states"
  docs_site: "not changed - measured, not assumed: grep -rln 'run\\.retired|factory runs|basis' docs-site/ README*.md .moai/docs/ returns only the four README files (already assessed above); docs-site/ has no factory-runs page (confirmed at the t1107 sync, S3, unchanged since)"
  codemaps: "not restamped - the amendment's run phase touched only internal/homestate and internal/factorymsg (7 files, no internal/cli); graph freshness threshold unchanged from the t1107 sync assessment"

mx_tag_check:
  scope: "internal/homestate/{boot_time_procstat.go (new),boot_time_unix.go,factory_run_boot_proof_test.go,factory_run_retire.go,runtime.go}, internal/factorymsg/factory_run_boot_proof_legs_test.go (new)"
  finding: "grep -rn '@MX' internal/homestate/*.go internal/factorymsg/*.go -> 0 hits, both before and after this amendment's run phase (git show a0b78213d:internal/homestate/factory_run_retire.go | grep -c @MX -> 0). Neither package carries any @MX annotation anywhere, including the 20 pre-existing exported functions/types in factory_run_retire.go untouched by this amendment (e.g. ClassifyRuns, ReconcileActiveRuns, RetireRunIfDead). The new exported type `ProofBasis` (internal/homestate/factory_run_retire.go) and its three constants (BasisStamp/BasisPeer/BasisBoot) carry godoc comments but no @MX tag, matching the file's existing convention rather than diverging from it - adding tags here alone would be scope creep against an un-annotated package, not a repair of something this amendment broke"

sync_phase_verification:
  spec_lint: "moai spec lint SPEC-FACTORY-RUN-RETIRE-001 (pre-commit) -> 'No findings - all SPEC documents are valid', exit 0"
  spec_audit: "moai spec audit --json --filter-spec SPEC-FACTORY-RUN-RETIRE-001 (pre-commit) -> total_specs 1, modern_era_clean 1, one INFO EraAutoDetected finding (H-4), no drift - both re-run after this commit with an unchanged result"
  scope: "git diff --name-only a0b78213d HEAD -- .moai/specs/SPEC-FACTORY-RUN-RETIRE-001 CHANGELOG.md - spec.md (frontmatter only), progress.md (this section), CHANGELOG.md (1 new entry); no Go source touched by sync"
```

