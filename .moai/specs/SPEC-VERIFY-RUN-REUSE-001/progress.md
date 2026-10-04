# SPEC-VERIFY-RUN-REUSE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: draft
plan_complete_at: 2026-10-03
plan_audit_iter1: FAIL 0.76 (.moai/reports/t1452/plan-audit-iter1.md) — D1-D4 blocking
revision: 0.2.0 applied D1-D4 + D6, D7, D8, D12, D13; open: D5 (tier call), D10 (heading renumber), D11 (symbol names in REQs)
_plan-audit iter2 pending (card t1452, Tier S, REQ 8 / AC 8)_

## §E.2 Run-phase Evidence

Tree: worktree `.claude/worktrees/t1452`, branch `WT-merge-window-hold-time`, base `20e447a46`. Judging toolchain: `go` as installed (go1.26.x per `golangci-lint version` build line), golangci-lint `v2.1.6` at `/Users/goos/go/bin/golangci-lint`; moai binary build not used as a measurement tool (tests run through `go test`, `go run ./cmd/moai spec lint` builds from this tree).

### Commits (each message ends with `(card t1452)`, trailer `Authored-By-Agent: manager-develop`, marker `🗿 MoAI`)

| SHA | Subject |
|---|---|
| see `git log 20e447a46..HEAD --format='%h %s'` | M1 decision functions + spec status draft -> in-progress; M2 verb; M3 doctrine sentence; this evidence commit |

### TDD RED (E8, observed before GREEN, this session, tree = base + tests only)

- `go test -v -run '^(TestDecideReuseHit|TestCanonicalCommand|TestDecideReuseKeyMismatch|TestDecideReuseCommandBytes|TestDecideReuseNonzeroExit|TestDecideReuseTTL|TestEnvDigest)$' -count=1 ./internal/verify/` -> compile failure `undefined: DecideReuse`, `undefined: CanonicalCommand`, ... `FAIL github.com/modu-ai/moai-adk/internal/verify [build failed]`; `--- PASS` lines: 0 (N=7 expected).
- `go test -v -run '^(TestVerifyRunHitExecutesZeroTimes|...|TestVerifyRunUsageErrors)$' -count=1 ./internal/cli/` (14 names) -> 14 x `--- FAIL`, 0 x `--- PASS`, `FAIL github.com/modu-ai/moai-adk/internal/cli 15.946s` (the `run` verb did not exist). Exit code not captured as a number (output piped through grep/head); the `FAIL` line is the observation.
- AC-VRR-007 RED-now equivalent was already recorded in acceptance.md (`grep -n -F -e "--env" AGENTS.md` -> exit 1); re-observed GREEN below.
- Mutant probe on AC-005 (GrandchildKilled): changing `syscall.Kill(-pid, ...)` to `syscall.Kill(pid, ...)` in `verify_run_unix.go` -> `--- FAIL: TestVerifyRunTimeoutAndNotFound/GrandchildKilled` ("the grandchild survived the timeout: its marker file exists"); restored, GREEN again.

### GREEN, per AC (tree after commit `c44f64459` + doctrine edit; commands run once on the final tree)

Command (CLI, 14 names + 3 existing verify tests): `go test -v -run '^(TestVerifyRunHitExecutesZeroTimes|TestVerifyRunMissOnTreeChange|TestVerifyRunMissOnCommandBytes|TestVerifyRunNeverReusesFailure|TestVerifyRunMissOnTTL|TestVerifyRunMissOnEnvChange|TestVerifyRunToolVersionBinding|TestVerifyRunToolVersionTimeout|TestVerifyRunExitCodePassthrough|TestVerifyRunTimeoutAndNotFound|TestVerifyRunTreeMovedNotRecorded|TestVerifyRunStoreFailOpen|TestVerifyRunIgnoresHandRecordedEntry|TestVerifyRunUsageErrors|TestVerifyHelpListsVerbs|TestVerifyRegisteredOnRoot|TestVerifyRecordThenCheckFresh)$' -count=1 ./internal/cli/`

```
--- PASS: TestVerifyRunHitExecutesZeroTimes (2.27s)
--- PASS: TestVerifyRunMissOnTreeChange (3.21s)
--- PASS: TestVerifyRunMissOnCommandBytes (4.52s)
--- PASS: TestVerifyRunNeverReusesFailure (2.19s)
--- PASS: TestVerifyRunMissOnTTL (1.59s)
--- PASS: TestVerifyRunMissOnEnvChange (2.82s)
--- PASS: TestVerifyRunToolVersionBinding (8.04s)
--- PASS: TestVerifyRunToolVersionTimeout (1.55s)
--- PASS: TestVerifyRunExitCodePassthrough (3.31s)
--- PASS: TestVerifyRunTimeoutAndNotFound (8.95s)
    --- PASS: TestVerifyRunTimeoutAndNotFound/ExitCodes124And127 (3.50s)
    --- PASS: TestVerifyRunTimeoutAndNotFound/GrandchildKilled (5.45s)
--- PASS: TestVerifyRunTreeMovedNotRecorded (2.24s)
--- PASS: TestVerifyRunStoreFailOpen (1.35s)
--- PASS: TestVerifyRunIgnoresHandRecordedEntry (1.80s)
--- PASS: TestVerifyRunUsageErrors (0.69s)
--- PASS: TestVerifyHelpListsVerbs (0.00s)
--- PASS: TestVerifyRegisteredOnRoot (0.00s)
--- PASS: TestVerifyRecordThenCheckFresh (1.25s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	46.899s
```

Command (unit): `go test -v -run '^(TestDecideReuseHit|TestCanonicalCommand|TestDecideReuseKeyMismatch|TestDecideReuseCommandBytes|TestDecideReuseNonzeroExit|TestDecideReuseTTL|TestEnvDigest)$' -count=1 ./internal/verify/` -> 7 `--- PASS` lines (Hit, CanonicalCommand, KeyMismatch, CommandBytes, NonzeroExit, TTL, EnvDigest), `ok  	github.com/modu-ai/moai-adk/internal/verify	0.265s`, no `[no tests to run]`.

13-name existence: `go test ./internal/cli -list '^(...13 names...)$'` printed the 13 names (TestVerifyRunHitExecutesZeroTimes ... TestVerifyRunIgnoresHandRecordedEntry) then `ok  	github.com/modu-ai/moai-adk/internal/cli	0.764s`.

| AC | Result | Evidence |
|---|---|---|
| AC-VRR-001 | PASS | CLI `TestVerifyRunHitExecutesZeroTimes` (1 line), unit `TestDecideReuseHit` (1 line); counter file = 1 after 5 runs, runs 2-5 stderr carry `reuse key= recorded_at= duration_ms=`, empty stdout |
| AC-VRR-002 | PASS | unit 3 lines (CanonicalCommand, KeyMismatch, CommandBytes), CLI 2 lines (MissOnTreeChange, MissOnCommandBytes) |
| AC-VRR-003 | PASS | unit 2 lines (NonzeroExit, TTL), CLI 2 lines (NeverReusesFailure, MissOnTTL) |
| AC-VRR-004 | PASS | unit 1 line (EnvDigest), CLI 3 lines (MissOnEnvChange, ToolVersionBinding, ToolVersionTimeout) |
| AC-VRR-005 | PASS | 4 lines: ExitCodePassthrough, TimeoutAndNotFound, /ExitCodes124And127, /GrandchildKilled (ran, not skipped, on darwin) |
| AC-VRR-006 | PASS | 3 lines: TreeMovedNotRecorded, StoreFailOpen (regular file at the snapshot dir path, no permission bits), IgnoresHandRecordedEntry |
| AC-VRR-007 | PASS | five fixed-string greps below, plus byte ceiling test |
| AC-VRR-008 | PASS | `git diff --name-only 2b9e4a4d0 HEAD` below |

AC-VRR-007 (each `grep -n -F -e "<phrase>" <file>`, exit 0, exactly 1 matching line):

```
AGENTS.md:                                 "moai verify run" 188 | "output not re-observed" 193 | "recorded_at" 192 | "runs the command directly" 194 | "--env" 190
internal/template/templates/AGENTS.md.tmpl: "moai verify run" 184 | "output not re-observed" 189 | "recorded_at" 188 | "runs the command directly" 190 | "--env" 186
```

Sentence identity: `sed -n 188,194p AGENTS.md | md5` = `sed -n 184,190p internal/template/templates/AGENTS.md.tmpl | md5` = `6e04e796d33d46da64f3e29135c7d437`. Byte ceiling: `go test ./internal/config -run '^TestCodexContractByteCeiling$' -count=1 -v` -> `AGENTS.md = 22224 bytes (ceiling 24576, headroom 2352)`, `AGENTS.md.tmpl = 21936 bytes (ceiling 24576, headroom 2640)`, `--- PASS: TestCodexContractByteCeiling`. Also `go test ./internal/template -run 'Disclosure|Neutrality|AGENTS' -count=1` -> `ok`.

AC-VRR-008 (`git diff --name-only 2b9e4a4d0 HEAD`, 17 lines at the time of measurement, before this evidence commit): `.moai/reports/t1452/{plan-audit-iter1,plan-audit-iter2,verdict}.md`, `.moai/specs/SPEC-VERIFY-RUN-REUSE-001/{acceptance,decision-index,plan,progress,spec}.md`, `AGENTS.md`, `internal/cli/{verify,verify_run,verify_run_test,verify_run_unix,verify_run_windows}.go`, `internal/template/templates/AGENTS.md.tmpl`, `internal/verify/{run,run_test}.go`. None under `internal/kanban/`, `internal/homestate/`, `internal/factorylane/`, none of `internal/cli/integration*|factory_complete*|factory_merge*|factory_card*`, no `AGENTS.local.md`, `gitflow-lane-protocol.md`, `kanban-dispatch-mechanics.md`. Deviation from plan.md §B table: two build-tag files `verify_run_unix.go` and `verify_run_windows.go` (plan M1.5 prescribes build-tag separation of platform code, B1).

### Quality batch (E2, E4, E5)

- `go vet ./internal/verify ./internal/cli` -> no output, exit 0 (`vet exit=0` echoed).
- `GOOS=windows GOARCH=amd64 go build ./...` -> no output, `winbuild exit=0`.
- `golangci-lint run --timeout=5m ./internal/verify/... ./internal/cli/` (v2.1.6, `/Users/goos/go/bin/golangci-lint`) -> first run 3 errcheck findings in `verify_run_test.go` (os.Stdout/Stderr.WriteString), fixed; re-run `0 issues.`
- `go run ./cmd/moai spec lint .moai/specs/SPEC-VERIFY-RUN-REUSE-001/spec.md` -> `0 error(s), 0 warning(s)` (one INFO OwnershipTransitionUnmeasured about the plan-phase commit's missing trailer, pre-existing).
- AskUserQuestion not used in the new CLI files (the verb never prompts; exit codes and stderr only).
- `go test ./internal/verify -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/verify	3.043s` (whole package).

### Gaps (not observed)

- Windows execution of the new tests (only `GOOS=windows go build ./...` was run; the Windows skip for GrandchildKilled, the no-op `verifyRunPrepare` and CI Windows matrix are unobserved). `gofmt -l internal/cli internal/verify` lists a pre-existing unformatted file `internal/cli/todo_classify_llm_test.go` that this card does not touch; the files this card adds are not listed.
- Exit codes of the RED runs were not captured as numbers (piped through grep/head); the build-failure and `--- FAIL` lines are the observation.
- The TDD ordering (tests first) is asserted here and in the M1 commit message; the commit graph does not witness it for M1/M2 because tests and implementation share a commit (verification-claim-integrity.md §2.3 permanent deviation, atomic authoring act).
- Full `internal/cli` package suite and the rest of the repository were not run (scoped verification per AGENTS.md §4; CI runs the full suite). `go test` was run once per tree state per command; the CLI selector run above is on the tree after the lint fix (test file edit, then the run).
- The installed `moai` binary was not rebuilt or used; `moai verify run` was exercised only through the in-process cobra command in tests, not as a built binary on a real `go test` invocation (no end-to-end observation of a real 30-minute suite).
- `moai verify check` interplay and consumers (sync Stop hook, sync-audit-4dim) with `verify run` entries were not exercised (SPEC §D-7 unchanged contract).

### Residual risk

- Spec §D items stand (unlisted env, toolchain `unversioned`, gitignored inputs, flaky tests, Windows grandchildren, shared store across same-key trees, other consumers, key-moving commands, reuse = prior observation, 16-hex digest).
- `Setpgid` puts the command in its own process group, so a Ctrl-C at the terminal reaches only the `moai` process, not the command's group; the command then outlives an interrupted verb until it ends or its `--timeout`. Not covered by the SPEC or tests.
- Test runtime: each helper spawn runs the cli package `TestMain` (about 0.3s); the 14 CLI tests take about 47s together.
- `exitCodeError` text `verify run: command exited with code N` is printed by the root error path on a non-zero command exit (extra stderr line after the command's own output).

### §E.2 revision — sync-audit F0 (registration) and F6 (tool-identity group kill), commit `a05ec62de`

Cause (F0): `verify.go` `init()` called `newVerifyCmd()` before `verify_run.go` `init()` appended to `verifyExtraCommands` (files init alphabetically), so the verb was never registered; every earlier test built the command directly. Fix: `cmd.AddCommand(newVerifyRunCmd(&projectRoot))` in `newVerifyCmd`, init append dropped. F6: `verifyRunPrepare(cmd)` now also applies to the `--tool-version-cmd` command. F1 (Ctrl-C) stays recorded debt, not implemented.

RED, observed before the fix (tree = HEAD `710e0eed5` + the 3 new tests):
`go test -v -run '^(TestVerifyRunRegisteredOnRoot|TestVerifyRunThroughRootCommand|TestVerifyRunToolVersionGrandchildKilled)$' -count=1 ./internal/cli/`
```
--- FAIL: TestVerifyRunRegisteredOnRoot (0.00s)        (rootCmd.Find(verify run) returned the `verify` command itself)
--- FAIL: TestVerifyRunThroughRootCommand (0.42s)      (first run printed the verify help text, not "counted")
--- FAIL: TestVerifyRunToolVersionGrandchildKilled (3.97s)  (the tool-version grandchild survived the timeout: its marker file exists)
FAIL	github.com/modu-ai/moai-adk/internal/cli	5.339s
```

GREEN on the final tree, env-scrubbed (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -v -run '^(TestVerifyRun[A-Za-z0-9]*|TestVerifyHelpListsVerbs|TestVerifyRegisteredOnRoot|TestVerifyRecordThenCheckFresh)$' -count=1 ./internal/cli/`): 20 top-level `--- PASS` lines (17 VerifyRun tests incl. the 3 new, plus the 3 existing verify tests) and the 2 `TimeoutAndNotFound` subtest lines, `--- SKIP: TestVerifyRunHelperProcess` (the helper), no `[no tests to run]`, `ok  	github.com/modu-ai/moai-adk/internal/cli	38.930s`. The acceptance.md PASS-line counts are unchanged (no AC selector name was added or removed; the 3 new tests are outside the AC selectors), so acceptance.md was not edited.

Real binary, built from this tree (`go build -o <scratch>/moai-t ./cmd/moai`, scratch repo with a committed `.gitignore` containing `.moai/`), env-scrubbed:
```
$ moai-t verify run --project-root <repo> -- echo hi
verify run: miss (no snapshot recorded for the current tree)
hi                                   exit=0
$ (same command again)
verify run: reuse key=4663871e1bbf204e700486b6385a209a3edccf43:6e340b9cffb37a98 recorded_at=2026-10-03T20:15:26+09:00 duration_ms=5     exit=0 (no "hi")
$ moai-t verify run --project-root <repo> -- false
verify run: miss (no entry recorded for this command)      exit=1
$ moai-t verify run --help | head -3   -> "Run <command...> in the project root, unless a passing result ..."
```
A first attempt in a scratch repo WITHOUT `.gitignore` for `.moai/` missed on every run (three snapshots under three keys): the store directory itself is an untracked path and moves the tree key. That is the Key() contract (the real project and the test fixtures ignore `.moai/`), not a defect of the verb, but a project that does not gitignore `.moai/` never gets a reuse.

Quality batch on the final tree: `go vet ./internal/verify ./internal/cli` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=5m ./internal/verify/... ./internal/cli/` (v2.1.6) `0 issues.`; `gofmt -l` on the three changed files empty.

Pre-existing, out of this card: `verify sync-gate` (verify_receipts.go) has the same init-order defect on this base and is NOT registered. Evidence from the built binary: `moai-t verify sync-gate --help` prints the parent `verify` help ("Shared diagnostic snapshot contract. ...") instead of the sync-gate help; its registration was deliberately left unchanged.

Gaps: Windows execution of the new tests unobserved (build only); the F6 test is Unix-only (skipped on Windows, residual risk §D-5); real-binary check was `echo`/`false`, not a real long test suite.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-03
run_commits: see §E.2 (M1 verify decision functions, M2 verb, M3 doctrine sentence, evidence)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete (3-phase close: in-progress -> implemented -> completed on the single sync commit)
sync_complete_at: 2026-10-03
sync_commit_sha: 45c1ee2fc (backfilled per D3 placeholder pattern)
sync_commit_subject: docs(SPEC-VERIFY-RUN-REUSE-001): sync-phase - 3-phase close (card t1452)
docs_changed: CHANGELOG.md (Unreleased/Added entry)
docs_gap: docs-site/** names `moai verify` only inside the 4-locale MCP-server guide table (docs-site/content/{en,ko,ja,zh}/guides/mcp-server.md); no CLI verb list exists. A new verb row there would need a 4-locale same-PR edit, so it was NOT made and is recorded as a Gap.

### Known debt

- Interrupt (Ctrl-C) leaves the child command running: Setpgid puts it in its own process group and no signal is forwarded.
- Windows execution is unobserved (only a GOOS=windows build was possible).
- TDD order (RED before GREEN) is not witnessed by the commit graph; RED was observed in-session only.
- Plan-audit closed as PASS-WITH-DEBT by leader decision (iter2 FAIL 0.88; N1 closed by measurement).
