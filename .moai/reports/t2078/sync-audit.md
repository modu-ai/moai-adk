auditor-model: claude-opus-5-5

# t2078 sync-phase audit — intake.register pre-send intent (Class B)

verdict: PASS
audited_sha: eec30d8479c712de6b2d29187b505eb0242a1368

Overall: PASS, score 0.93 (weighted 93.6/100). No blocking finding. Scope audited: d4d625f54de8a211abc666498923d1a65e950234..eec30d8479c712de6b2d29187b505eb0242a1368 (one commit, 3 files, +60/-0), tree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2078, branch WT-intake-register-register. Profile: built-in default (Class B, no SPEC, no evaluator_profile); no `Binding run conditions` heading in progress.md.

## Dimension Scores

| Dimension | Score | Verdict | Evidence (verbatim, full commands below) |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | RED on pre-fix code: `intake_register_crash_test.go:41: RegisterBranch ran 2 times for one request, want 1`; fixed tree: `ok .../intake 12.832s`; 4 auditor probes PASS |
| Security (25%) | 95/100 | PASS | diff adds only static Detail/Resume strings; no input, secret, or injection surface; `fail()` redaction path unchanged |
| Craft (20%) | 88/100 | PASS | `coverage: 91.8% of statements`; `register 90.0%`; `golangci-lint` `0 issues.`; two mutants survive (F1, F2) |
| Consistency (15%) | 95/100 | PASS | mirrors t2072 `publish` intent block line for line (same Resume text, same error return); `gofmt -l internal/agenticfactory/` empty |

Must-pass firewall (Functionality, Security): both PASS.

## Findings (structured defect-list)

- F1 [Low] [optional] [in-diff] internal/agenticfactory/intake/intake.go:289-291 — the new intent-write error return (store Busy ⇒ nothing sent) is not pinned by a committed test; mutant M7 (`return claim.Request{}, err` → `_ = err`, i.e. send anyway) passes the whole intake suite. The code is correct (auditor probe `TestProbeBusyIntentSendsNothing` passes on it). Confidence high. Required fix (discretionary): add a register analogue of `TestAdvanceFailureWhileBusyKeepsBothErrors` (lock file present at StageIssueConfirmed ⇒ `ErrBusy`, `br.calls == 0`, no Blocked row).
- F2 [Low] [optional] [touched-file] internal/agenticfactory/intake/intake.go:296 — the `ErrNotSent` branch of `register` (which now must replace the in-flight Unknown row with a Retryable one) has no test; mutant M5 (`!errors.Is(err, ErrNotSent)` → `true`) passes the suite. Behaviour is correct today (`fail()` overwrites Blocked unconditionally; probe `TestProbeNotSentIsRetryable` passes). Confidence high. Required fix (discretionary): a brancher analogue of `TestAnErrorThatProvesNothingWasSentMayBeRetried`.
- F3 [Low] [optional] [in-diff] internal/agenticfactory/intake/intake.go:285-288 — the intent closure writes unconditionally; it does not re-check `r.Stage == StageIssueConfirmed` and absence of an Unknown block inside the update. Two concurrent `Advance` calls on one key could both pass `GetRequest` and both send `RegisterBranch`. Not a regression (pre-fix code double-sent in the same race, and t2072's `publish` has the identical shape); whether concurrent Advance on one key is reachable was not established. Confidence medium. Required fix (discretionary, would also apply to publish): compare-and-set inside the closure, returning `ErrTransitionInvalid` when the row moved.
- F4 [Info] [optional] [in-diff] intake_register_crash_test.go — the "restart" reuses the same `*Intake`/store instance rather than reopening the store from disk. Durability of the intent row across a real reopen was verified by the auditor probe `TestProbeCrashThenReopenedStore` (PASS), so this is test-strength only.
- F5 [Info] [optional] [in-diff] process, verification-claim-integrity §2.3 — "RED observed first" is asserted in the commit message; test and fix share commit eec30d8479, so git cannot witness the authoring order. The RED property itself was re-witnessed by this audit (test run against the base `intake.go` fails for the stated reason).

Claims in the author's records that were checked: RED message (`RegisterBranch ran 2 times`) — confirmed; tests ok ×3 — confirmed (timings differ, fresh run); gofmt empty — confirmed; the "died before send ⇒ ConfirmNotCreated" path the author listed as untested — exercised by this audit's probe and correct; card-review covered the combined t2072+t2078 diff — consistent with its own disclosure; this audit adds a t2078-only delta review (below).

## Claim

1. The new test fails on the pre-fix code for the stated reason (second RegisterBranch), and passes on the fix.
2. Every exit from `register` leaves a coherent row: success ⇒ `StageBranchRegistered`, Blocked nil; ErrNotSent ⇒ Retryable, not Unknown, retry registers once; unknown failure ⇒ Unknown (unchanged `fail`); Busy at intent ⇒ ErrBusy, nothing sent, no row; malformed receipt ⇒ unchanged `fail(...,true)` path, still caught by `TestBranchReceiptNeedsAFullBaseOIDAndRef`.
3. A crash after the intent but before the send leaves Unknown; lookup finds nothing ⇒ `ErrEffectUnknown`, no send; `ConfirmNotCreated` then permits exactly one registration.
4. The success path clears the block (`recordBranch` sets `r.Blocked = nil`), and the suite would catch a mutant that did not (M6 fails 3+ tests).
5. No regression in intake, claim, runner, e2e, monitor packages; gofmt, vet, golangci-lint clean.
6. The required codex gate is met on the t2078 delta.

## Evidence

E1 — prescribed test run (audited tree, HEAD eec30d8479):
```
$ cd /Users/goos/.moai/worktrees/projects/moai-adk-go/t2078 && timeout 600 go test ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	12.832s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	69.593s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	36.500s
$ gofmt -l internal/agenticfactory/
(no output)   gofmt_exit=0
```

E2 — wider regression + static checks:
```
$ timeout 600 go test ./internal/agenticfactory/e2e/ ./internal/agenticfactory/monitor/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/e2e	14.999s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/monitor	0.313s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/adapt	1.377s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/flow	0.380s
$ go vet ./internal/agenticfactory/intake/        -> no output, vet_exit=0
$ golangci-lint run ./internal/agenticfactory/intake/
0 issues.
```

E3 — coverage:
```
$ go test ./internal/agenticfactory/intake/ -count=1 -coverprofile=<scratch>/intake.cov
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	8.585s	coverage: 91.8% of statements
$ go tool cover -func=<scratch>/intake.cov
intake.go:281:	register		90.0%
intake.go:301:	recordBranch		100.0%
intake.go:313:	lookup			75.0%
intake.go:342:	ConfirmNotCreated	100.0%
intake.go:390:	fail			88.9%
total:		(statements)		91.8%
```

E4 — RED re-witness (scratch copy = `git archive eec30d8479` with `intake.go` replaced by `git show d4d625f54d:internal/agenticfactory/intake/intake.go`):
```
$ go test ./internal/agenticfactory/intake/ -count=1 -run 'TestAProcessDyingDuringRegisterDoesNotRegisterTwice'
--- FAIL: TestAProcessDyingDuringRegisterDoesNotRegisterTwice (0.25s)
    intake_register_crash_test.go:41: RegisterBranch ran 2 times for one request, want 1
FAIL
```

E5 — mutant probes (same scratch copy, fixed `intake.go` mutated one line at a time, full intake suite):
```
== M1 intent Unknown:false
--- FAIL: TestAProcessDyingDuringRegisterDoesNotRegisterTwice  intake_register_crash_test.go:41: RegisterBranch ran 2 times for one request, want 1
== M2 intent Step=StageAccepted
--- FAIL: TestAProcessDyingDuringRegisterDoesNotRegisterTwice  intake_register_crash_test.go:44: the restart should resolve through the lookup: stage ISSUE_CONFIRMED
== M5 register ErrNotSent treated as unknown
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	10.922s      <- survives (F2)
== M6 recordBranch does not clear Blocked
--- FAIL: TestFullPathThroughAllSixStagesWithEvidence  intake_test.go:361: advance must wait for an agent reservation: <nil>
--- FAIL: TestAttemptRefusedUntilReadyAckedByTheSameAgent / TestOnlyTheReservedAgentCanAck
== M7 register intent store error ignored (sends anyway)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	10.580s      <- survives (F1)
== M8 (intent removed) -> build failed in my perl edit; not counted (E4 covers the pre-fix shape)
restored
```

E6 — auditor probe tests on the fixed code (scratch-only file `zz_audit_probe_test.go`, not committed):
```
--- PASS: TestProbeDiedBeforeSendNeedsConfirmNotCreated (0.27s)
--- PASS: TestProbeBusyIntentSendsNothing (2.38s)
--- PASS: TestProbeNotSentIsRetryable (0.10s)
--- PASS: TestProbeCrashThenReopenedStore (0.11s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	3.249s
```

E7 — audit plan and cross-model legs:
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2078
"config_status": "ok", "model": "codex", "cross_model_active": true, "cross_model_required": true,
backends: claude gate=off, codex gate=required, glm gate=off; "enforced_required": ["codex"]
```
- `audit_multi` #1 (target baseBranch, delta_base d4d625f54d): overall_verdict=inconclusive, gate_unmet=codex, plan_source=config, per_backend [codex gate=required verdict=inconclusive] (codex could-not-run), receipt rcpt-7b83398251bfb271064cabe6.
- `codex_audit` native (baseBranch, delta_base d4d625f54d, review_base "delta_base (merge base d4d625f5...)"): verdict=pass, findings=[], resolved_model gpt-6.1-sol, receipt rcpt-e4cbbcc9238c3c126d2f9ea5.
- `audit_multi` #2 (same scope): overall_verdict=pass, gate_unmet="" (absent), plan_source=config, per_backend [codex gate=required verdict=pass], receipt rcpt-3360e4d8d7a1cfd599ed23d3.

## Baseline-attribution

- All test/lint/coverage outputs: this run (2026-11-06), against tree HEAD eec30d8479c712de6b2d29187b505eb0242a1368, `git status --short` clean before and after. `go test` compiles the tree itself, so no tool-build lag applies to E1-E6.
- RED baseline: base file content from commit d4d625f54de8a211abc666498923d1a65e950234 (t2072 head), placed into an archive export of eec30d8479; the only difference from the audited tree was that one file.
- moai tooling (audit plan, MCP codex legs): build commit 9b632c1eb3 (rc.41), a strict ancestor of the tree HEAD. Commits between 9b632c1eb3 and eec30d8479 touch only `internal/agenticfactory/{claim,intake}` (8 files, `git diff --stat 9b632c1eb3..eec30d8479`), none of which implements audit-plan resolution or the codex review transport, so the lag is stated and judged immaterial to those outputs.

## Gaps

- `audit_multi` #1 returned codex inconclusive (could-not-run, cause not reported); passed on direct native call and on the retry. The first result is recorded, not discarded.
- codex in `audit_multi` is scoped to the delta in prose only (tool note: "not mechanically bounded"); the native `codex_audit` leg honours delta_base exactly.
- Numeric exit codes for E1/E2 were not captured (zsh `PIPESTATUS` is empty); the verdict rests on the `ok` lines and the absence of FAIL lines.
- Not run: `go test ./...`, `-race`, other OSes, real GitHub/git brancher (only fakes exercised).
- Mutant M8 did not compile (my edit), so it proves nothing; E4 covers the same shape.
- `prepare` (PrepareWorktree) intent gap noted by the author was not examined (outside card).
- The audit_multi result is digest-reported here; the orchestrator's `moai verify audit-plan --result-file` convergence check was not run by this auditor. This verdict is not final until that check passes.
- No tool refusals affected a measurement. Two command forms were refused by the worktree guard (compound `git archive`/`git -C` lines from the t2080 session); each was re-issued as a plain split command and succeeded, so no fallback substitution occurred.

## Residual-risk

- Concurrent `Advance` on the same key can still double-send (F3); unchanged from before the fix and shared with `publish`.
- A crash between the intent row and the send now strands the request in EFFECT_UNKNOWN until an operator supplies `ConfirmNotCreated` evidence. That is the intended contract (same as publish), but it shifts a former silent duplicate into a manual step; the monitor shows such rows as EFFECT_UNKNOWN (`monitor/viewext.go:58`), including transiently during every normal registration.
- A real Brancher whose `FindBranch` cannot see a just-registered branch (lag) leaves the row Unknown rather than resolving it; covered only by fakes.
- The two untested branches (F1, F2) are correct now but unguarded against regression.

## Operational Notes (unverified)

- F1 and F2 are cheap to close in a follow-up card together with the `prepare` examination the author proposed; F3 would be a joint publish+register change.
- The moai binary/MCP server (9b632c1eb3) lags this tree; rebuild before relying on moai CLI output that concerns agenticfactory code.
- Receipts issued during this audit: rcpt-7b83398251bfb271064cabe6, rcpt-e4cbbcc9238c3c126d2f9ea5, rcpt-3360e4d8d7a1cfd599ed23d3.
- Iteration 2 (same SHA, no code change): the SubagentStop receipt guard refused the first citation ("receipt unknown to the store"). One more native `codex_audit` (baseBranch, delta_base d4d625f54d, project_root t2078) returned verdict=pass, findings=[], receipt rcpt-6ffb5c97d20b28e27aa3a6cb. A likely cause, not verified: the auditor session runs in worktree t2080 while the receipts were issued for tree t2078, so a guard that reads the session's own tree would not find them.
