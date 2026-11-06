---
card: t2118
class: B
stage: sync
iteration: 2
auditor: sync-auditor
auditor_model: claude-opus-5-5
verdict: PASS
audited_sha: 734f8ac819f7a1706d3214405a583b2a9d4cf54f
audited_range: 04d8c27a7ed77afc09217f8623519806d48df3c4..734f8ac819f7a1706d3214405a583b2a9d4cf54f
delta_since_previous_audit: f7f536ade974c1edcb0e0d214d5536683cd0ada4..734f8ac819f7a1706d3214405a583b2a9d4cf54f
score: 0.92
dimensions:
  functionality: 0.95 (PASS, must-pass)
  security: 0.95 (PASS, must-pass)
  craft: 0.90 (PASS)
  consistency: 0.88 (PASS)
codex_receipts: [rcpt-e273b2e393454f11fea57e3e]
date: 2026-11-06
---

auditor-model: claude-opus-5-5

# t2118 sync audit, iteration 2: PASS (0.92)

## Overall

Both iteration-1 blocking findings are closed.

- **F1 (helper DSN)**: `qasRegistryState` now opens the registry through `qasReadOnlyDSN` → `sqliteuri.File(path, v)`. A drive path now yields `file:///C:/…`. The query is unchanged: `mode=ro` plus `_pragma=query_only(ON)`, with no busy_timeout before or after. The RED commit `dfefa51452` is an ancestor of the fix `95dfec79b5`, and at RED the helper still carried the old `url.URL{Scheme:"file", Path: filepath.ToSlash(path)}` builder.
- **F2 (cause counts)**: my own recount from `categorized.tsv` gives 469 rows of shape `C:%5C` and 5 of shape `C: (1)`. Every one of the 474 `invalid uri authority` rows maps to a producer the card has fixed: `ledger: ` (ledger.Open/ledgerDSN) 469, `factory census:` (runtime_census) 2, and `dump workers:` (qasRegistryState) 3. 469 + 2 + 3 = 474, and none of them is unfixed. The progress record §3 now matches this split.

The repair commits did not introduce a regression. The only code change is a behavior-preserving extraction plus a builder swap inside a `_test.go` file. The helper's consumers (AC018, AC021) still pass on darwin. The codex delta review (required gate) returned pass with no findings.

## Findings

No blocking findings.

- **F3 [MINOR] [optional] [in-diff, carried from iter-1] internal/sqliteuri/sqliteuri_test.go**: the guard regex still catches only the `"file:"+` concatenation form, not the `url.URL{Path: raw}` form that actually produced the CI errors. Not addressed in this repair, which the progress record §6 states as a deliberate scope choice. Confidence medium.
- **F4 [MINOR] [optional] [outside, carried from iter-1]**: seven prefix-slash `url.URL` builders remain: `internal/agenticfactory/ledger/readonly.go:20-24`, `internal/agenticfactory/ledger/importer/importer.go:170-174`, `internal/factorymsg/store.go:167-171`, `internal/homestate/profile_lease.go:52-56`, `internal/homestate/factory.go:279-283` and `:328-332`, `internal/cli/factory_quota_lanes.go:118-122`. I re-checked in this run that each one prepends `/`, so drive paths are safe. They do not absolutize relative paths. Consistency debt only.
- **F5 [MINOR] [optional] [in-diff, carried from iter-1] internal/sqliteuri/sqliteuri.go:23-48**: edge inputs (`C:rel.db`, `""`, `:memory:`). None is reachable from the switched callers. Unchanged since iter-1.
- **F6 [LOW] [optional] [outside] test helpers using `"file:"+filepath.ToSlash(path)`**, for example `internal/agenticfactory/ledger/importer/importer_test.go:24`, `internal/factorymsg/legacy_broker_test.go:50`, `internal/homestate/fr_schema_test.go:108`, `internal/discovery/factory_discovery_test.go:405`, and `internal/agenticfactory/ledger/failure_paths_test.go:388` (`"file:"+path`, no ToSlash). On Windows these yield `file:C:/…` or `file:C:\…`, which carry no `//` authority marker, so SQLite's URI parser should not see an authority. None appears in the 474 CI rows. However, some of them sit in tests that failed earlier at `ledger.Open`, so CI may not have reached them (masking). Confidence low that any fails, and only the Windows CI matrix can settle it. Not a card defect.

No finding is repair-born. The repair-stop rule does not apply.

## Iteration history

| Iter | Audited SHA | Verdict | Score | Blocking |
|---|---|---|---|---|
| 1 | cf616281365f5b60bb84a81918339c926580509c | FAIL | 0.78 | F1 (helper DSN, 3 CI rows), F2 (progress §3 counts 473/1) |
| 2 | 734f8ac819f7a1706d3214405a583b2a9d4cf54f | PASS | 0.92 | none (F1, F2 closed) |

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `--- PASS: TestQASReadOnlyDSN_DrivePathIsNotTheAuthority (0.00s)`; `--- PASS: TestQAS_AC018_LaneInventoryCandidates/read_only_main_file_and_rows_unchanged (0.76s)`; `--- PASS: TestQAS_AC021_SteeringChangesNothing/status_is_read_only (2.28s)`; producer recount 469 + 2 + 3 = 474, zero rows unmapped |
| Security (25%) | 95/100 | PASS | repair diff is test-only, with no new input boundary; secret grep over `git diff f7f536ade9..HEAD` gives exit 1 (no match); read-only flags `mode=ro` + `query_only(ON)` asserted by the new test |
| Craft (20%) | 90/100 | PASS | `ok github.com/modu-ai/moai-adk/internal/sqliteuri 3.358s coverage: 100.0% of statements`; `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` exit=0; gofmt -l empty (exit 0); RED before fix confirmed by ancestry |
| Consistency (15%) | 88/100 | PASS | reuses the existing `sqliteuri.File` instead of a new builder; commits carry the card id; F3/F4 consistency debt carried (optional) |

Flat weighted: 0.4·0.95 + 0.25·0.95 + 0.2·0.90 + 0.15·0.88 = 0.93. Harmonic mean = 0.92 (reported). Must-pass dimensions Functionality and Security both clear their thresholds.

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim

1. F1 closed: RED precedes the fix. The helper now yields `file:///C:/…` for both drive-path spellings, query flags keep their meaning, and busy_timeout behavior is unchanged (absent before and after).
2. F2 closed: all 474 `invalid uri authority` rows map to fixed producers (469 / 2 / 3), and progress §3 states 469/5 correctly.
3. No regression from the repair commits. The helper's consumers still pass, and the codex delta review found nothing.
4. Lane-local verification passed: the targeted cli test, the sqliteuri package, and the Windows cross-vet of internal/cli.

### Evidence (verbatim; all in this run, tree HEAD 734f8ac819f7a1706d3214405a583b2a9d4cf54f, darwin, GOCACHE=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-3604-44b1-a008-c14cce281ea1/scratchpad/gocache-lane28)

Tree identity:
```
$ git rev-parse --show-toplevel && git rev-parse HEAD && git status --short
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2118
734f8ac819f7a1706d3214405a583b2a9d4cf54f
(empty)
```

Repair diff (code part), `git diff f7f536ade9..HEAD -- internal/`:
```
+func qasReadOnlyDSN(path string) string {
+	v := url.Values{}
+	v.Add("mode", "ro")
+	v.Add("_pragma", "query_only(ON)")
+	return sqliteuri.File(path, v)
+}
...
-	v := url.Values{}
-	v.Add("mode", "ro")
-	v.Add("_pragma", "query_only(ON)")
-	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
+	db, err := sql.Open("sqlite", qasReadOnlyDSN(path))
```

RED ordering:
```
$ git log --format='%h %ci %s' dfefa51452 -1
dfefa51452 2026-11-06 22:50:51 +0900 test(cli): RED — qasRegistryState DSN puts a Windows drive path in the URI authority (t2118)
$ git log --format='%h %ci %s' 95dfec79b5 -1
95dfec79b5 2026-11-06 22:59:25 +0900 fix(cli): qasRegistryState opens the registry through sqliteuri.File (t2118)
$ git merge-base --is-ancestor dfefa51452 95dfec79b5 && echo red_is_ancestor_of_fix
red_is_ancestor_of_fix
$ git show dfefa51452:internal/cli/factory_quota_lanes_test.go | grep -n 'url.URL{Scheme: "file"'
102:	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String()
```
RED output on file (`.moai/reports/t2118/red-cli-qas.txt`): `qasReadOnlyDSN("C:\\Users\\RUNNER~1\\...\\factory.db") = "file://C:%5CUsers%5CRUNNER~1%5C...?_pragma=query_only%28ON%29&mode=ro", want the drive path after file:///`, then `FAIL github.com/modu-ai/moai-adk/internal/cli 1.700s`.

CI recount (`F=…/scratchpad/categorized.tsv`, 1335 lines, mtime Nov 6 21:13):
```
$ grep -c 'invalid uri authority' $F
474
== shape C:%5C by package
 416 internal/agenticfactory/ledger
  34 internal/agenticfactory/mergeq
  18 internal/cli
   1 internal/web
== shape C: (1) by package
   4 internal/cli
   1 internal/homestate
== neither shape
0
== context prefix
   1 "my dir": ledger: 
   3 dump workers: 
   2 error "ledger: 
   1 factory census: 
  52 ledger: 
   1 legacy resume unknown owner census: factory census: 
   1 open 1: ledger: 
   5 open the ledger: ledger: 
   1 open the store: ledger: 
 407 open: ledger: 
== C: (1) rows
cli1	internal/cli	TestFactoryRecoverResumeCommandRequeuesUnknownOwnerAtZeroCensus	    handoff_recover_test.go:152: legacy resume unknown owner census: factory census: SQL logic error: invalid uri authority: C: (1) 
cli3	internal/cli	TestQAS_AC018_LaneInventoryCandidates/read_only_main_file_and_rows_unchanged	    factory_quota_lanes_test.go:626: dump workers: SQL logic error: invalid uri authority: C: (1) 
cli3	internal/cli	TestQAS_AC021_SteeringChangesNothing/status_is_read_only	    factory_quota_lanes_test.go:832: dump workers: SQL logic error: invalid uri authority: C: (1) 
cli3	internal/cli	TestQAS_AC021_SteeringChangesNothing/auto_mutations_equal_pressure_off	    factory_quota_lanes_test.go:860: dump workers: SQL logic error: invalid uri authority: C: (1) 
rem	internal/homestate	TestRuntimeCensusCountsLiveAndIgnoresProvablyDead	    runtime_census_test.go:88: factory census: SQL logic error: invalid uri authority: C: (1) 
== C:%5C rows lacking ledger: prefix
0
$ grep 'invalid uri authority' $F | cut -f5 | grep -cE '^ *[A-Za-z0-9_]+_test\.go:[0-9]+: '
474
$ grep -c 'dump workers: SQL logic error: invalid uri authority' $F
3
$ grep -c 'factory census: SQL logic error: invalid uri authority' $F
2
$ grep -c 'ledger: SQL logic error: invalid uri authority: C:%5C' $F
469
```
Producer mapping in the source at HEAD:
- `ledger: ` wraps `sql.Open("sqlite", ledgerDSN(path))` at `internal/agenticfactory/ledger/schema.go` (Open; `ledgerDSN` returns `sqliteuri.File(path, q)`). The `%5C` (raw backslash) shape fits the old `url.URL{Path: path}` only. The other ledger reader `readonly.go:20-24` prepends `/`, so it cannot produce an authority. Fixed by 9deb412a48.
- `factory census: %w` at `internal/homestate/runtime_census.go:117/139` wraps `sql.Open("sqlite", runtimeCensusDSN(factoryPath))` (line 110), where `runtimeCensusDSN` = `sqliteuri.File(factoryPath, url.Values{"mode": {"ro"}})` (line 31). Fixed by 9deb412a48.
- `dump workers:` comes from `t.Fatalf("dump %s: %v", q.table, err)` inside `qasRegistryState` (`internal/cli/factory_quota_lanes_test.go:110-150`). The CI lines 626/832/860 map to today's 633/839/867 (the call sites `qasRegistryState(t, root)`); the shift of +7 matches the import line plus the helper extraction. Fixed by 95dfec79b5.

Lane-local verification:
```
$ export GOCACHE=…/gocache-lane28 && go test -count=1 ./internal/cli/ -run TestQASReadOnlyDSN -v
=== RUN   TestQASReadOnlyDSN_DrivePathIsNotTheAuthority
--- PASS: TestQASReadOnlyDSN_DrivePathIsNotTheAuthority (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.739s
exit=0

$ export GOCACHE=…/gocache-lane28 && go test -count=1 -cover ./internal/sqliteuri/
ok  	github.com/modu-ai/moai-adk/internal/sqliteuri	3.358s	coverage: 100.0% of statements
exit=0

$ export GOCACHE=…/gocache-lane28 && GOOS=windows GOARCH=amd64 go vet ./internal/cli/
exit=0
$ gofmt -l internal/cli/factory_quota_lanes_test.go internal/cli/factory_quota_lanes_dsn_test.go internal/sqliteuri/
(empty) gofmt_exit=0

$ export GOCACHE=…/gocache-lane28 && go test -count=1 ./internal/cli/ -run TestQAS_AC018 -v
exit=0
--- PASS: TestQAS_AC018_LaneInventoryCandidates (8.42s)
    --- PASS: TestQAS_AC018_LaneInventoryCandidates/read_only_main_file_and_rows_unchanged (0.76s)
ok  	github.com/modu-ai/moai-adk/internal/cli	9.944s

$ export GOCACHE=…/gocache-lane28 && go test -count=1 ./internal/cli/ -run TestQAS_AC021 -v
exit=0
--- PASS: TestQAS_AC021_SteeringChangesNothing (4.39s)
    --- PASS: TestQAS_AC021_SteeringChangesNothing/status_is_read_only (2.28s)
    --- PASS: TestQAS_AC021_SteeringChangesNothing/auto_mutations_equal_pressure_off (1.13s)
    --- PASS: TestQAS_AC021_SteeringChangesNothing/no_card_class_read (0.97s)
ok  	github.com/modu-ai/moai-adk/internal/cli	5.677s
```

Security probe:
```
$ git diff f7f536ade9..HEAD | grep -inE 'password|secret|token|api[_-]?key' ; echo "secret_grep_exit=$?"
secret_grep_exit=1
```

Audit plan and codex (required gate):
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2118
{"config_status":"ok","model":"codex", backends: claude off, codex required, glm off, "cross_model_active":true,"enforced_required":["codex"]}

mcp__moai__codex_audit(project_root=<worktree>, target=baseBranch, mode=native, delta_base=f7f536ade974c1edcb0e0d214d5536683cd0ada4)
→ {"verdict":"pass","findings":[],"audit_receipt":"rcpt-e273b2e393454f11fea57e3e","review_base":"delta_base (merge base f7f536ade974c1edcb0e0d214d5536683cd0ada4)","build_commit":"04d8c27a7e","provenance":{"backend":"codex","resolved_model":"gpt-6.1-sol","requested_effort":"high"}}
  summary (verbatim, Korean): "지정 기준 커밋 대비 변경에서 지적할 결함은 찾지 못했어. 읽기 전용 DSN 옵션을 유지하면서 기존 `sqliteuri.File`을 재사용한 수정이야. 정적 검토만 했고, 지침대로 시험·빌드·재현 명령은 실행하지 않았어."
```
The tree was unchanged during the call (`git status --short` empty before and after; no write by this auditor except this verdict file afterwards).

### Baseline-attribution

Everything above was measured in this run against worktree HEAD `734f8ac819f7a1706d3214405a583b2a9d4cf54f`, which is clean. The CI baseline is the classification file `categorized.tsv` (474 matching rows) supplied by the lead. I did not re-derive it from the raw CI logs. Go tests ran with the system Go toolchain on darwin.

Tool provenance (§2.2): the `moai` binary and MCP server that ran `verify audit-plan` and `codex_audit` were built from `04d8c27a7e`, a strict ancestor of HEAD (the server reported the lag). The card changes no audit-plan or codex-audit code, so the lag does not touch these measurements. The codex review read the git diff of this tree, not the binary.

### Gaps

- **Windows was not executed.** Every Windows claim is string-level (the DSN test) plus `GOOS=windows go vet`. The Windows race CI matrix is the deciding run.
- Full `internal/cli`, `internal/web`, `internal/agenticfactory/{ledger,mergeq}`, and `internal/homestate` suites were not re-run, per the lane-focus rule. golangci-lint was not run.
- **Refused commands (§3.1).** (a) The recount pipeline written with `awk '{…}'` was refused by the worktree-isolation guard. I redid it with `grep`/`cut`/`sed`, and the output above comes from that fallback. The numbers match the lane's awk-based counts. (b) `go test -run 'TestQAS_AC0[12][81]'` was refused as too complex to verify. I split it into separate `-run TestQAS_AC018` and `-run TestQAS_AC021` calls.
- **First codex invocation rejected.** I passed a mistyped `delta_base` SHA, and the tool rejected the input ("is not a commit") before any review ran, so no receipt was issued. The second invocation ran and is the one cited (rcpt-e273b2e393454f11fea57e3e). Together with the earlier rcpt-c403c403d97e3c1d034d1dc7, this stage has two codex reviews that actually ran, within the standing-rule cap.
- **audit_multi was not called** although `cross_model_active: true`. The lead's standing rule d-20261106T080039Z-305c forbids it for this stage. The only required gate (codex) was met through `codex_audit` with a receipt; claude and glm gates are `off`.
- The RED test at `dfefa51452` was not re-run by me. The RED output comes from the lane's file, and the old builder at RED was confirmed with `git show`.

### Residual-risk

- F6: some `"file:"+…` test helpers sit in tests that failed earlier at `ledger.Open` on Windows, so CI could not have reached them. Once ledger opens, a previously masked failure could surface on the Windows matrix. I judge this unlikely (no `//` authority marker), but it is not ruled out.
- The scratch outputs under `/private/tmp/claude-501/.../scratchpad/` (`qas018.txt`, `qas021.txt`, `categorized.tsv`) are machine-local. The deciding lines are copied above. Any other content is a known loss.
