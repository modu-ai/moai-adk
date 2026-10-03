auditor-model: claude-sonnet-5-5[1m]

# Sync-phase audit — card t1452, SPEC-VERIFY-RUN-REUSE-001 (v0.2.2, Tier S)

verdict: FAIL
audited_sha: 710e0eed589d718f2eac48958792929948073665

Score: 62/100 — REVISED to FAIL after the codex receipt rcpt-9b6dddc66af83bcb42694803 and my own binary reproduction (see "REVISION — iteration 2" at the end). The initial PASS-WITH-DEBT 88 and "Blocking findings: none" in the body below are withdrawn; blocking finding F0.
Tree: `.claude/worktrees/t1452`, branch WT-merge-window-hold-time, base for diff 2b9e4a4d0. Tree clean before and after the audit (`git status --short` empty; both probed files restored, md5 identical to the pre-probe copies).

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | 8/8 REQ implemented; scoped tests PASS with counts matching acceptance.md (below); 4 of 5 mutation probes caught |
| Security (25%) | 88 | PASS | direct exec (no shell), env values SHA-256 hashed in ConfigDigest, fail-open degrades to re-execution; no secrets echoed |
| Craft (20%) | 84 | PASS | golangci-lint v2.1.6 0 issues, vet clean, gofmt clean, Windows build ok; recorded_at semantic untested; Ctrl-C debt |
| Consistency (15%) | 90 | PASS | AGENTS.md and template mirror added-lines identical; cobra extra-command pattern; scope clean |

## Findings

- F1 [medium] [debt, non-blocking] internal/cli/verify_run_unix.go:13-17, verify_run.go:119 — Ctrl-C/SIGTERM on `moai verify run` leaves the child (e.g. a 30-60 min `go test`) running: the child is in its own process group (Setpgid) and the verb installs no signal handler, so nothing forwards the signal or fires cmd.Cancel. That is exactly the background load AGENTS.md §4 forbids. Severity judgment: real but bounded (child eventually ends; only an operator abort path; declared in progress.md §E.4 and CHANGELOG). Required fix (cheap, one function): derive ctx via `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` in verifyRun so cancellation reaches `cmd.Cancel` (group kill) and is not recorded (completed=false). Related, reasoned not measured: a Setpgid-only (non-foreground) group that reads the controlling terminal's stdin would receive SIGTTIN; not an issue for non-interactive test commands.
- F2 [low] [optional] internal/cli/verify_run.go:233 — the REQ-VRR-004 clause "recorded_at = completion time of the command" is not caught by any test. Mutation probe `RecordedAt: end` -> `start` left all 15 `VerifyRun` tests green. Effect of the defect would be a TTL measured from start (shorter validity, safe direction). Fix: a CLI test with a `sleep` helper longer than a `--ttl` that is shorter than sleep+epsilon but longer than epsilon after completion.
- F3 [low] [optional] AGENTS.md:188-194 and the .tmpl mirror — REQ-VRR-008 asks each key phrase on a single line; the phrase "prior observation" (named in acceptance.md prose, not in its five grep phrases) wraps across two lines (`grep -c -F "prior observation"` = 0 in both files). The five AC phrases each match exactly 1 line. No AC fails.
- F4 [low] [optional] internal/verify/run.go:67 — `FindCommand` returns one entry per command; if the same command was recorded under two `--check-id` values on one key, a stale pass entry could shadow a newer fail entry (reasoned from reading, not reproduced; requires two different check ids for one command, which the verb's default avoids).
- F5 [info] [optional] verify_run.go:205 — a signal-terminated child exits the verb with 1, not 128+n; documented in a code comment, matches the "no exit code to pass through" stance.

## Requirement / AC matrix (measured)

| REQ / AC | Code | Test evidence |
|---|---|---|
| REQ-001/002, AC-001 | DecideReuse + CheckReceipt in run.go:60-94; notice at verify_run.go:158 | TestVerifyRunHitExecutesZeroTimes, TestDecideReuseHit PASS |
| REQ-002/005, AC-002 | CanonicalCommand run.go:23; FindCommand byte match | TestCanonicalCommand, TestDecideReuseKeyMismatch, TestDecideReuseCommandBytes, TestVerifyRunMissOnTreeChange, TestVerifyRunMissOnCommandBytes PASS |
| REQ-002, AC-003 | exit/verdict leg run.go:91 | TestDecideReuseNonzeroExit, TestDecideReuseTTL, TestVerifyRunNeverReusesFailure, TestVerifyRunMissOnTTL PASS |
| REQ-002/006, AC-004 | EnvDigest run.go:46; tool identity verify_run.go:258 | TestEnvDigest, TestVerifyRunMissOnEnvChange, TestVerifyRunToolVersionBinding, TestVerifyRunToolVersionTimeout PASS |
| REQ-003, AC-005 | verifyRunExec verify_run.go:180; group kill unix.go | TestVerifyRunExitCodePassthrough, TestVerifyRunTimeoutAndNotFound + 2 subtests PASS (4 PASS lines) |
| REQ-004/007, AC-006 | verifyRunRecord :215; fail-open :138-163 | TestVerifyRunTreeMovedNotRecorded, TestVerifyRunStoreFailOpen, TestVerifyRunIgnoresHandRecordedEntry PASS |
| REQ-008, AC-007 | AGENTS.md §4 sentence + mirror | 5 phrases x 2 files = exactly 1 match each; TestCodexContractByteCeiling PASS |
| AC-008 | file scope | see scope check below |

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim 1: acceptance counts hold on this tree
Evidence (verbatim, filtered `--- ` lines):
```
$ go test -v -count=1 -run VerifyRun ./internal/cli/
--- SKIP: TestVerifyRunHelperProcess (0.00s)
--- PASS: TestVerifyRunHitExecutesZeroTimes (1.68s)
--- PASS: TestVerifyRunMissOnTreeChange (3.36s)
--- PASS: TestVerifyRunMissOnCommandBytes (4.67s)
--- PASS: TestVerifyRunNeverReusesFailure (1.34s)
--- PASS: TestVerifyRunMissOnTTL (1.45s)
--- PASS: TestVerifyRunMissOnEnvChange (4.12s)
--- PASS: TestVerifyRunToolVersionBinding (10.09s)
--- PASS: TestVerifyRunToolVersionTimeout (1.48s)
--- PASS: TestVerifyRunExitCodePassthrough (1.99s)
--- PASS: TestVerifyRunTimeoutAndNotFound (5.95s)
    --- PASS: TestVerifyRunTimeoutAndNotFound/ExitCodes124And127 (1.60s)
    --- PASS: TestVerifyRunTimeoutAndNotFound/GrandchildKilled (4.35s)
--- PASS: TestVerifyRunTreeMovedNotRecorded (1.49s)
--- PASS: TestVerifyRunStoreFailOpen (1.18s)
--- PASS: TestVerifyRunIgnoresHandRecordedEntry (1.26s)
--- PASS: TestVerifyRunUsageErrors (0.42s)
ok  	github.com/modu-ai/moai-adk/internal/cli	41.291s
$ go test -v -count=1 -run 'DecideReuse|CanonicalCommand|EnvDigest' ./internal/verify/
--- PASS: TestDecideReuseHit / TestCanonicalCommand / TestDecideReuseKeyMismatch / TestDecideReuseCommandBytes / TestDecideReuseNonzeroExit / TestDecideReuseTTL / TestEnvDigest  (7 PASS lines)
ok  	github.com/modu-ai/moai-adk/internal/verify	0.242s
```
13 acceptance-named CLI tests each 1 PASS line; 7 unit PASS lines (1+3+2+1); no `[no tests to run]`. The skipped HelperProcess is the designed non-helper-mode SKIP.
Baseline-attribution: run in this audit against tree HEAD 710e0eed5. The two-package scope was run as two invocations (the worktree guard refuses the single compound regex form `A|B|C` across two packages), results equal the requested union.

### Claim 2: mutation probes catch the named hazards
Method: in-place edit of the audited files, run the pinned tests, restore from a copy (md5 verified identical: verify_run.go bf83f1b2..., run.go 1fe16fe6...).
| Probe | Mutation | Result |
|---|---|---|
| P1 reuse of a failed prior exit | run.go `if false && (entry.ExitCode != 0 ...)` | CAUGHT: TestVerifyRunNeverReusesFailure FAIL, TestDecideReuseNonzeroExit FAIL |
| P2 record when tree key moved | verify_run.go `if false && keyAfter != keyBefore` | CAUGHT: TestVerifyRunTreeMovedNotRecorded FAIL |
| P3 reuse after env change | `EnvDigest(nil, ...)` (ignore --env) | CAUGHT: TestVerifyRunMissOnEnvChange FAIL |
| P4 recorded_at = start not completion | `RecordedAt: start` | SURVIVED (F2): all VerifyRun tests ok |
Note: P1 also failed the CLI test through the verb path, so the exit-code leg is guarded twice (CheckReceipt does not itself check exit code; DecideReuse leg is the only guard).

### Claim 3: toolchain gates
Evidence: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/verify ./internal/cli` exit 0; `gofmt -l` (verify, verify_run*.go) empty; `golangci-lint run ./internal/verify/... ./internal/cli/` -> `0 issues.` (v2.1.6, CI version); `moai spec lint spec.md` -> "No findings".
Judging build: the golangci-lint binary is the installed v2.1.6 (CI pin per memory), not rebuilt from the tree — it is a linter of the tree, not a tool under test; the `moai` spec-lint ran via `go run ./cmd/moai` (built from this tree).

### Claim 4: doctrine sentence and mirror
Evidence: `git diff 2b9e4a4d0 HEAD -- AGENTS.md internal/template/templates/AGENTS.md.tmpl` shows the same 8 added lines in both (byte-identical hunks, offsets differ only by line number); the five AC phrases each `grep -c -F` = 1 in both files; `TestCodexContractByteCeiling` PASS. Content satisfies AGENTS.md §1: states a reuse is a prior observation, the verdict cites key + recorded_at, lists "output not re-observed" under Gaps, "a reused result is a Gap, not a Claim", and a claim needing verbatim output runs directly. Satisfies §1.

### Claim 5: scope
`git diff --name-only 2b9e4a4d0 HEAD` = 18 paths (0 under internal/kanban, internal/homestate, internal/factorylane; none matching internal/cli/integration*, factory_complete*, factory_merge*, factory_card*; no AGENTS.local.md, gitflow-lane-protocol.md or kanban-dispatch-mechanics.md). Control: output non-empty (18 lines). Base is a pinned SHA, valid pre-absorption (acceptance AC-008 says so). Code files touched: verify.go (1 help line), verify_run.go, verify_run_unix.go, verify_run_windows.go (+tests), verify/run.go (+tests), AGENTS.md, the .tmpl, CHANGELOG.md.

### Claim 6: close contract
spec.md frontmatter `version: "0.2.2"`, `status: completed`; HISTORY row 0.2.2 records the 3-phase close. `sync_commit_sha: 45c1ee2fc` appears in progress.md §E.4 line 117 (backfilled by 710e0eed5); `git show --stat 45c1ee2fc` = spec.md, progress.md, CHANGELOG.md (single sync commit). The SPEC frontmatter itself carries no `sync_commit_sha` field — the repo's pattern keeps it in progress.md §E.4, so I treat this as consistent, not a defect. CHANGELOG entry states the three known debts.

### Correctness-hazard review (judgment from reading unless a probe is cited)
- Stale-reuse paths: five-leg conjunction (key, byte-identical command, bound config digest, tool identity, TTL) plus exit/verdict; hand-recorded entries carry no digest/tool_version so CheckReceipt reads them unbound (tested). `unversioned` marker is bound by default — absent flag cannot match an entry with tool identity. Residual: unbound inputs (§D) by design.
- TTL: measured by `now.Sub(RecordedAt) > ttl` in CheckReceipt; RecordedAt=`end` is correct in code (see F2 for the untested guard). A future-dated RecordedAt (clock skew) would never expire — not reachable by the verb's own writes.
- Exit codes: 0/N passthrough, signal-termination -> 1, start failure 127, timeout 124 checked before the error switch; neither 124 nor 127 is recorded (`completed=false`). A timeout check using `runCtx.Err()` could mislabel a command that finishes exactly at the deadline — negligible.
- Process-group kill: `Setpgid` + `Kill(-pid, SIGKILL)` + `WaitDelay` 2s; GrandchildKilled test PASS on darwin. Ctrl-C: F1.
- Fail-open store: tool-identity failure, store resolution, key, load errors all set `bound=false` -> execute without reuse or record; record errors are noted and non-fatal. No path fabricates a hit.
- Tool-version binding: executed every invocation, bounded 30s, stdout-only; failure unbinds.
- Simplicity: ~280 + 95 production lines for one verb with 8 REQ, thin over existing Key/CheckReceipt/RecordCheck; no new store/schema. @MX:ANCHOR with REASON and SPEC on DecideReuse; sub-line shape per protocol. Go conventions (context first param, error wrapping, build-tagged unix/windows files) followed.

## Gaps (not observed)
- Windows execution of the verb never run; only a Windows cross-build (progress.md §E.4 carries the same debt).
- Ctrl-C behaviour (F1) reasoned from Setpgid semantics, not exercised with a real SIGINT; SIGTTIN claim not measured.
- F4 (multi check-id shadowing) not reproduced.
- Cross-model audit (audit_multi / codex / glm) not run in this audit — the task asked for a skeptical single-auditor measurement; no `workflow.audit.gates.codex` receipt was issued or cited.
- Full `go test ./...` and the broader `internal/cli` suite not run (scope tests only, per instruction); regressions outside `-run VerifyRun|DecideReuse|CanonicalCommand|EnvDigest` and TestCodexContractByteCeiling are unobserved.
- progress.md §E claims were not trusted; the only §E item used is the sync_commit_sha line, cross-checked with `git show --stat`.

## Residual risk
Documented §D inputs (GOFLAGS, toolchain, gitignored data, flaky tests, cross-lane same-key reuse through the primary-checkout store) remain; the lane doctrine sentence is the only mitigation. F1 leaves an operator-abort background-load path until fixed.

## Recommendations
- Optional but cheap: fix F1 with `signal.NotifyContext` before wiring this verb into the integration gate (the SPEC defers that wiring to a later change).
- Add the F2 test when F1 is touched.

## REVISION — iteration 2 (supersedes the verdict above): FAIL

- F0 [P1] [BLOCKING] internal/cli/verify_run.go:41 + internal/cli/verify.go:265 — `moai verify run` is NOT registered in the real CLI. verify.go's `init()` (file order: verify.go before verify_run.go) calls `newVerifyCmd()`, which consumes `verifyExtraCommands` before verify_run.go's `init()` appends `newVerifyRunCmd`. Reproduced on a binary built from this tree: `moai verify run -- echo hi` printed the `verify` help text and exited 0 (no execution, no reuse); `moai verify --help` lists `run` only in the static Long text, not under COMMANDS (COMMANDS: record, check, audit-plan, codex-review). All tests passed because they build the command via `newVerifyRunCmd` directly, not through `rootCmd`. My earlier PASS-WITH-DEBT missed this: I never ran a built binary (a Gap I should have listed). Required fix: register order-independently (e.g. have verify.go's group read `verifyExtraCommands` lazily, or add the verb directly in `newVerifyCmd`) and add a test that executes `verify run` through the real root/`newVerifyCmd()`. Note: `verify sync-gate` (verify_receipts.go) shows the same ordering defect on this base — pre-existing, outside this card; report to leader.
- F1 raised to [P2] [debt]: codex reproduced Ctrl-C leaving the child writing after the parent exits.
- F6 [P2] [recommended] verify_run.go:264 — the `--tool-version-cmd` timeout kills only the direct child; a grandchild survives (codex isolated test, not reproduced by me). Apply `verifyRunPrepare` to that command.

Evidence: `go build -o <scratch>/moai-t ./cmd/moai` then `moai-t verify run -- echo hi` -> verify help text, `exit=0`. Codex native audit (project_root = this worktree, baseBranch): verdict fail, P1 + 2xP2, receipt `rcpt-9b6dddc66af83bcb42694803`.
Gaps added: F6 not reproduced by me; the registration defect was previously unobserved.

## Iteration history
Iteration 1 (superseded by the REVISION above): first sync-audit of the card at HEAD 710e0eed5. Plan-audit history (iter1 FAIL 0.76, iter2 FAIL 0.88 closed as PASS-WITH-DEBT by leader) is in `plan-audit-iter1.md`/`plan-audit-iter2.md`.
