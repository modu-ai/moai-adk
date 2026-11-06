---
card: t2118
class: B
stage: sync
iteration: 1
auditor: sync-auditor
auditor_model: claude-opus-5-5
verdict: FAIL
audited_sha: cf616281365f5b60bb84a81918339c926580509c
audited_range: 04d8c27a7ed77afc09217f8623519806d48df3c4..cf616281365f5b60bb84a81918339c926580509c
score: 0.78
dimensions:
  functionality: 0.70 (FAIL, must-pass)
  security: 0.95 (PASS, must-pass)
  craft: 0.85 (PASS)
  consistency: 0.85 (PASS)
codex_receipts: [rcpt-c403c403d97e3c1d034d1dc7]
date: 2026-11-06
---

auditor-model: claude-opus-5-5

# t2118 sync audit, iteration 1: FAIL (0.78)

## Overall

The production fix is correct. `sqliteuri.File` produces a clean `file:///C:/…` DSN for every drive-path form, keeps the old bytes for POSIX absolute paths, encodes `#`/`%`/`?`/space, and the switched sites keep their query semantics. RED was committed before the fix. The security flag is a false positive.

The card still fails on Functionality (a must-pass dimension). Of the 474 CI rows, **3 are not fixed by this diff**. They come from a test helper the lane scoped out on the grounds that test-code builders were "not in the CI evidence", but the CI evidence does contain them. The progress record's cause table also gets the error-shape counts wrong (it says 473/1; the measured split is 469/5), so the Class B cause-establishing evidence contains a false statement.

## Findings

- **F1 [BLOCKING] [outside the diff, with reproduction evidence] internal/cli/factory_quota_lanes_test.go:114**: the test helper `qasRegistryState` still builds `(&url.URL{Scheme: "file", Path: filepath.ToSlash(path), …}).String()`. On Windows that yields `file://C:/…?mode=ro&_pragma=query_only%28ON%29`, which is the same authority bug. The CI table has 3 rows from it, all ending in `dump workers: SQL logic error: invalid uri authority: C: (1)`: `TestQAS_AC018_LaneInventoryCandidates/read_only_main_file_and_rows_unchanged`, `TestQAS_AC021_SteeringChangesNothing/status_is_read_only`, and `TestQAS_AC021_SteeringChangesNothing/auto_mutations_equal_pressure_off`. They will still fail on the Windows race matrix after this card lands. **Required fix:** build that DSN with `sqliteuri.File(path, v)`, or reuse the production `factoryQuotaReadOnlyDSN` (internal/cli/factory_quota_lanes.go:113), whose query keys match apart from busy_timeout. Confidence high, severity High (the card's own defect stays open).
- **F2 [BLOCKING] [in-diff] .moai/reports/t2118/progress.md:28-30**: the cause table says shape (1) `C:%5C…` covers 473 rows ("ledger 416 + mergeq 34 + cli 22 + web 1") and shape (2) `C: (1)` covers 1 row (homestate). Measured: 469 rows of shape (1), all carrying the `ledger: ` wrap, and 5 of shape (2). Those 5 are homestate 1 + cli `TestFactoryRecoverResumeCommandRequeuesUnknownOwnerAtZeroCensus` 1, both through `factory census:`, which this card fixes, plus the cli 3 from F1, which it does not. **Required fix:** correct §3 to 469/5, attribute the 5 by producer, and remove the claim that "every test goes through ledger.Open" for the cli 22. Severity Medium.
- **F3 [MINOR] [optional] [in-diff] internal/sqliteuri/sqliteuri_test.go:447**: the guard regex `"file:"\s*\+` catches only the concatenation form. Both production sites that actually produced the CI errors (ledger `url.URL{Path: raw}`, census `url.URL{Path: filepath.ToSlash(p)}`) used the url.URL form without a leading slash, which the guard cannot see. So the guard does not stop the real regression class from coming back. Suggested: also flag `Scheme: "file", Path: filepath.ToSlash(` and `Path: path` in production code, or route the 7 remaining builders through `sqliteuri.File`. Confidence medium.
- **F4 [MINOR] [optional] [outside the diff]**: seven equivalent prefix-slash builders remain: `ledger/readonly.go:20-24`, `ledger/importer/importer.go:170-174`, `factorymsg/store.go:167-171`, `homestate/profile_lease.go:52-56`, `homestate/factory.go:279-283` and `:328-332`, `cli/factory_quota_lanes.go:118-122`. They are correct for drive paths (I checked each one has the `"/"+p` prefix), but they don't absolutize relative paths the way `File` does. That is a consistency debt, not a defect. The lane left them alone on purpose.
- **F5 [MINOR] [optional] [in-diff] internal/sqliteuri/sqliteuri.go:37-48**: edge inputs. A drive-relative `C:rel.db` becomes `file:///C:rel.db` (not absolutized). `""` becomes the cwd as a directory, where the old ledger form `file:?…` silently opened a private temp DB. `":memory:"` becomes a file literally named `:memory:` in the cwd. On darwin, a backslash UNC string is treated as relative. None of these is reachable from the switched callers: grep found no `:memory:` in the affected packages, the heartbeat guards `b.DB == ""`, and `afLedgerPath` always joins. For Windows UNC, the new output `file:////server/share/x.db` matches the old census/discovery output, while the old ledger form (`file://%5C%5Cserver…`) and the old concat form (`file://server/share…`) were already broken. No regression at any switched site.

## Answers to the lead's checklist

1. **Cause**: confirmed for 471 of 474. The probe reproduces old `ledgerDSN` → `file://C:%5CUsers%5CRUNNER~1%5C…%5Cledger.db?_pragma=busy_timeout(5000%29…`, the CI string shape. Old census → `file://C:/…?mode=ro`. The remaining 3 come from F1, which is not fixed.
2. **sqliteuri.File by input class**: correct for backslash and forward-slash drive paths, lowercase drives, POSIX absolute paths, relative paths, and special characters. UNC: no regression (F5). Query encoding: modernc v1.60.1 reads the query with `url.ParseQuery` (sqlite.go:286), so `%28`/`%29` decode to parens, and the full `file:` DSN goes to SQLite with `SQLITE_OPEN_URI` (conn.go:99-105). backlogDSN already relied on the encoded-paren form in production.
3. **Behavior preservation**: `mode=ro` is kept at census, migrate×2 and discovery; `_pragma=busy_timeout(100)` at factorymsg×5 and factory_card; the ledger pragma list and `_txlock=immediate` are unchanged in meaning. The concat sites now also encode `?`, which closes an injection of URI parameters through a path containing `?`.
4. **RED first**: `e267f5f4a6` (RED) comes before `9deb412a48` (fix). RED's `ledgerDSN` and `runtimeCensusDSN` are the old forms, failing with `invalid port ":%5CUsers…" after host`, the Go-side analogue of the authority error. The RED tests were not re-run by me (Gap).
5. **Security Guardian path-traversal at schema.go:85**: false positive. `path` comes from `afLedgerPath(--db)` or from the heartbeat binding file, both operator-owned. The same value went to the driver before the change. `File` adds no path resolution beyond `filepath.Abs` on relative input, and it encodes rather than interprets.
6. **Lane-local verification**: all green (Evidence below).

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim
C1 The production fix is correct and passes on darwin. C2 It cross-compiles and vets for windows/amd64. C3 3 of 474 CI rows remain unfixed (F1). C4 The progress-record shape counts are wrong (F2). C5 The encoded-paren query is accepted by the driver. C6 RED preceded GREEN.

### Evidence (verbatim; all in this run, tree HEAD cf61628136, go1.27.3 darwin/arm64, GOCACHE=…/scratchpad/gocache-lane28)

```
$ go test -count=1 -v ./internal/sqliteuri/   → exit=0
--- PASS: TestFile_WindowsDriveIsPathNotAuthority (0.00s)
--- PASS: TestFile_HostIsNeverTheDriveLetter (0.00s)
--- PASS: TestFile_RelativePathDoesNotBecomeAuthority (0.00s)
--- PASS: TestFile_OpensARealFileOnThisOS (0.01s)
--- PASS: TestNoHandBuiltFileDSNInProduction (1.24s)
ok  	github.com/modu-ai/moai-adk/internal/sqliteuri	1.564s
$ go test -count=1 ./internal/discovery/      → exit=0
ok  	github.com/modu-ai/moai-adk/internal/discovery	9.871s
$ go test -count=1 ./internal/factorymsg/     → exit=0
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	143.993s
$ GOOS=windows GOARCH=amd64 go vet ./internal/sqliteuri/ ./internal/agenticfactory/ledger/ ./internal/homestate/
exit=0   (no output)
$ go test -count=1 -run TestLedgerDSN_WindowsDrivePathIsNotTheAuthority -v ./internal/agenticfactory/ledger/
--- PASS: TestLedgerDSN_WindowsDrivePathIsNotTheAuthority (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	0.334s
$ go test -count=1 -run TestRuntimeCensus -v ./internal/homestate/
--- PASS: TestRuntimeCensusDSN_WindowsDrivePathIsNotTheAuthority (0.00s)
--- PASS: TestRuntimeCensusCountsLiveAndIgnoresProvablyDead (0.17s)
(+5 more PASS) ok  	github.com/modu-ai/moai-adk/internal/homestate	4.265s
$ go test -count=1 -cover ./internal/sqliteuri/
ok  	github.com/modu-ai/moai-adk/internal/sqliteuri	2.376s	coverage: 100.0% of statements
$ golangci-lint run ./internal/sqliteuri/
0 issues.
$ gofmt -l <10 changed .go files>   → gofmt-exit=0, no files listed
```

CI table recount (categorized.tsv from the #1969 analysis, CI head 9b632c1eb3, which is an ancestor of HEAD: `git merge-base --is-ancestor` → `ancestor=0`, i.e. true):
```
$ grep -c 'invalid uri authority' categorized.tsv
474
$ … | awk -F'\t' '{print $3}' | sort | uniq -c
 416 internal/agenticfactory/ledger
  34 internal/agenticfactory/mergeq
  22 internal/cli
   1 internal/homestate
   1 internal/web
$ … | grep -o 'invalid uri authority: C:[^?]\{0,4\}' | sed 's/%5C.*/%5C…/' | sort | uniq -c
   5 invalid uri authority: C: (1)
 469 invalid uri authority: C:%5C…
$ grep 'invalid uri authority: C:%5C' categorized.tsv | grep -c 'ledger: '
469
$ grep 'invalid uri authority: C: (1)' categorized.tsv | awk -F'\t' '{print $3, $4, $5}'
internal/cli TestFactoryRecoverResumeCommandRequeuesUnknownOwnerAtZeroCensus  handoff_recover_test.go:152: legacy resume unknown owner census: factory census: SQL logic error: invalid uri authority: C: (1)
internal/cli TestQAS_AC018_LaneInventoryCandidates/read_only_main_file_and_rows_unchanged  factory_quota_lanes_test.go:626: dump workers: SQL logic error: invalid uri authority: C: (1)
internal/cli TestQAS_AC021_SteeringChangesNothing/status_is_read_only  factory_quota_lanes_test.go:832: dump workers: SQL logic error: invalid uri authority: C: (1)
internal/cli TestQAS_AC021_SteeringChangesNothing/auto_mutations_equal_pressure_off  factory_quota_lanes_test.go:860: dump workers: SQL logic error: invalid uri authority: C: (1)
internal/homestate TestRuntimeCensusCountsLiveAndIgnoresProvablyDead  runtime_census_test.go:88: factory census: SQL logic error: invalid uri authority: C: (1)
```
Producer of "dump workers" at HEAD (`internal/cli/factory_quota_lanes_test.go:114`, inside `qasRegistryState`, `t.Fatalf("dump %s: %v", q.table, err)`):
```
db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
```
"factory census" producer: `internal/homestate/runtime_census.go:117/139` (fixed by this card).

DSN probe (scratchpad copy of `File`/`hasDrive`, verbatim from sqliteuri.go; Windows `ToSlash` simulated as `\`→`/`):
```
in="C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\x\\ledger.db"
  new       =file:///C:/Users/RUNNER~1/AppData/Local/Temp/x/ledger.db?mode=ro host=""
  oldLedger =file://C:%5CUsers%5CRUNNER~1%5CAppData%5CLocal%5CTemp%5Cx%5Cledger.db?_pragma=busy_timeout(5000%29&_pragma=journal_mode(WAL%29&_txlock=immediate
  oldCensusW=file://C:/Users/RUNNER~1/AppData/Local/Temp/x/ledger.db?mode=ro
in="//server/share/x.db"
  new       =file:////server/share/x.db?mode=ro host=""
  oldCensusW=file:////server/share/x.db?mode=ro
  oldConcatW=file://server/share/x.db?mode=ro
in="\\\\server\\share\\x.db"   oldLedger =file://%5C%5Cserver%5Cshare%5Cx.db?…
in="/tmp/a b/c#d%e?f.db"
  new       =file:///tmp/a%20b/c%23d%25e%3Ff.db?mode=ro host=""
  oldConcatW=file:/tmp/a b/c#d%e?f.db?mode=ro
in="rel/x.db"   oldLedger =file://rel/x.db?…   new=file:///<cwd>/rel/x.db?mode=ro
in=""           oldLedger =file:?…             new=file:///<cwd>?mode=ro
in="C:rel.db"   new=file:///C:rel.db?mode=ro
```
Driver query parsing: `modernc.org/sqlite@v1.60.1/sqlite.go:286  q, err := url.ParseQuery(query)`; `conn.go:99-105` opens the full dsn with `SQLITE_OPEN_URI`.

RED order: `git log --oneline 04d8c27a7e..HEAD` → `cf61628136 [t2118] docs…`, `9deb412a48 fix(sqliteuri)…`, `e267f5f4a6 test(…): RED…`. RED-commit `ledgerDSN` (`git show e267f5f4a6:…/schema.go`) = the old `url.URL{Scheme: "file", Path: path, RawQuery: strings.ReplaceAll(q.Encode(), "%28", "(")}`. The stub `File` = `return "file:" + path + "?" + query.Encode()`.

Remaining production builders: `grep -rn 'Scheme: *"file"' internal cmd pkg | grep -v _test.go` lists 7 non-sqliteuri SQLite builders, each preceded by a `"/" + p` prefix (F4).

Codex (standing rule d-20261106T080039Z-305c; 1 of 2 allowed calls used, tree unchanged; `git status --short` empty before and after, HEAD cf616281365f5b60bb84a81918339c926580509c):
```
mcp__moai__codex_audit(mode=native, target=baseBranch, delta_base=04d8c27a7e…)
→ {"verdict":"pass","findings":[],"audit_receipt":"rcpt-c403c403d97e3c1d034d1dc7",
resolved_model: gpt-6.1-sol, review_base: delta_base (merge base 04d8c27a7e…)
```
Codex reviewed the diff only and could not see F1, which is outside the diff and evidenced only by the CI log.

Audit plan: `moai verify audit-plan --project-root <tree>` → `config_status: ok`, `model: codex`, codex gate `required`, claude/glm `off`, `cross_model_active: true`.

### Baseline-attribution
All test, vet, lint, cover and probe outputs above were measured in this run against tree HEAD cf616281365f5b60bb84a81918339c926580509c (clean working tree). The CI rows come from the #1969 analysis table for CI head 9b632c1eb3 (an ancestor of this tree). Its line numbers (626/832/860) are from that head; the helper's current location is line 114 at HEAD.

### Gaps
- No real Windows execution. The Windows behaviour of `file:///C:/…` and `file:////server/share` is inferred from the SQLite URI rules and the existing Windows-tested backlogDSN precedent, not observed.
- I did not re-run the RED tests (that needs a checkout of e267f5f4a6, and the tree must stay unchanged). I relied on the lane's red-*.txt files plus my own probe reproducing the old DSN strings.
- Not run by me: full `internal/agenticfactory/ledger`, `mergeq`, the full `internal/homestate` package, `internal/cli`, `internal/web`. Their pass counts in progress.md §4 are the lane's.
- `audit_multi` was not used: the lead's standing rule forbids it, even though the plan reports `cross_model_active: true`. The required codex gate was met through `codex_audit` (receipt above). The MCP `codex_audit` provenance reports transport `codex app-server`, which is internal to the sanctioned MCP tool; no direct app-server call was made.
- The `build_lag` notice says the installed moai binary (04d8c27a7) is older than HEAD. The only moai CLI output cited here (audit-plan) reads config, which this diff does not change.

### Residual-risk
- After F1 is fixed, other Windows-only failures could surface in tests that previously failed early at `ledger.Open` and now run further, for example the other 32 hand-built test DSNs. The concat `file:C:/…` form opens on Windows, but paths with spaces or `#` would not. Only the next Windows race run can measure this.
- `internal/factorymsg/ensure_schema_bench_test.go:72` has the same authority-form bug, but it only runs under `-bench`, which the race matrix does not use.

## Iteration history
- iter1 (this file): FAIL 0.78 @ cf616281365f5b60bb84a81918339c926580509c. Blocking: F1, F2.
