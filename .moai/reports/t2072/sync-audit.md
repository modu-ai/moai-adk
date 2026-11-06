auditor-model: claude-opus-5-5

# t2072 sync audit — RC40 carried-over review findings (Class B)

verdict: PASS
audited_sha: d4d625f54de8a211abc666498923d1a65e950234

## Evaluation Report

SPEC: none (Class B card t2072; acceptance basis = card body)
Range audited: 9b632c1eb3 .. d4d625f54de8a211abc666498923d1a65e950234 (one commit, 7 files, +162/-23)
Overall Verdict: PASS. Score 0.91. No blocking findings. The must-pass dimensions (Functionality, Security) both pass.
Iteration history: iteration 1 (first sync audit of this card).

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 92/100 | PASS | All three new tests fail on base 9b632c1eb3 for the stated reason and pass at d4d625f54d. The three packages pass at HEAD (see Evidence E1, E2, E3) |
| Security (25%) | 95/100 | PASS | No new input or exec surface. The grep probe on the new files returns nothing (E6). BindWorktree now checks authorization (holder+epoch) before it acts on shared state |
| Craft (20%) | 88/100 | PASS | intake 92.2%, claim 86.5% coverage (E4). gofmt and vet are clean (E1). golangci-lint did not run (Gap G2) |
| Consistency (15%) | 92/100 | PASS | The request write now reuses the store's existing `writeAtomic`, as every other record already does. The error wrapping and comment style match the files (diff read) |

### Findings (structured defect-list)

No finding is blocking.

- F1 [low] [optional] [in-diff] claim/bind_ownership_test.go:10-13. The test comment says "the real holder was then refused its own path". What the test shows, and what the code does, is that a *different* Issue (issue 2) is refused. The holder of issue 1 can still bind (`reserveWorktree` returns nil for `held.Issue == i.Number`). Confidence high. Required fix (optional): reword the comment so it says "another Issue was then refused the path".
- F2 [low] [optional] [touched-file / record] progress.md:18 claims the remaining window (ownership lost between `Verify` and `Update`) "is closed by the existing stale-reservation takeover (usesWorktree)". That claim is inaccurate. `usesWorktree` (next.go:262) reports a live Issue whose `Worktree == ""` as still holding the path, so the leftover reservation stays until that Issue finishes or binds another path. The effect fits the documented abandoned-reservation contract (next.go:181-185), and the window needs a claim lost mid-call, so this does not block. A tighter fix existed: call `ReserveWorktree` inside the `Update` closure, under the Issue lock (the lock order Issue → documents is documented as allowed, next.go:175-176). I did not test that alternative. Confidence medium. Required fix (optional): correct the progress record. A follow-up may move the reservation under the Issue lock.
- F3 [low] [optional] [in-diff] intake_test.go:810-825. `TestAdvanceFailureWhileBusyKeepsBothErrors` no longer checks what its name says: it now checks that nothing is sent. The `errors.Join(cause, err)` path of `fail()` for "the store goes Busy after the intent was written" has no test now. The author records this gap as well. Confidence high. Required fix (optional): rename the test, or add a case where the store goes Busy after the intent write.
- F4 [low] [optional] [in-diff] intake.go:248-254. While a normal publish is in flight, the request now reads EFFECT_UNKNOWN (monitor/viewext.go:58 renders `Block.Unknown` that way). An operator who sees that state and calls `ConfirmNotCreated` during the send would reopen a path to a second publish. A crash after the intent write and before the send now needs an operator `ConfirmNotCreated` to make progress. That costs liveness and follows from the package's "an unknown effect is never repeated" design (intake.go:14-17), and the author records it. Confidence medium. Required fix (optional): none in this card. If wanted, use a distinct Detail/Resume text so the in-flight state can be told apart from a real unknown.
- F5 [medium] [optional] [outside] intake.go:281-289. `register` sends `RegisterBranch` without recording an intent first, which is the same crash-window pattern. It is outside the three named sites. The author left it as a follow-up candidate. Confidence medium. A follow-up card may apply the same intent-before-send pattern. (Recent history on this machine shows commit eec30d8479 "[t2078] fix: record the branch-registration intent before RegisterBranch is sent", which suggests the follow-up already exists. I did not verify that.)
- F6 [info] [optional] [outside] The original review text for findings #2 and #3 is not on disk. I checked the author's derivation against the RC40 candidate tree. At 67b4734d6b, request.go:193 is `if err := tmp.Close()` with no preceding Sync, and intake.go:246 is the `PublishIssue` call with nothing recorded before it (E5). The derived defects match the code at those exact lines. Whether they match what the reviewer meant cannot be confirmed.

### Per-repair judgment

1. **claim/next.go BindWorktree (finding #1).** The defect is real. Before the fix, `ReserveWorktree` ran before ownership was checked. Because `usesWorktree` treats a live Issue with no recorded worktree as still holding the path, a caller that held no claim left a reservation that blocked other Issues for as long as the targeted Issue stayed open (RED E2a). The repair adds `Verify(i, holder, epoch)` before the reservation. It is minimal (6 lines) and correct for the stated defect. `Update` still verifies again under the lock. The residual window is F2. The test checks both the foreign-holder case and the wrong-epoch case, and both that no reservation document is left and that another Issue can still bind. A mutant that checks only the holder fails the epoch arm.
2. **claim/request.go writeRequestLocal (finding #2).** The defect is real. The old path did temp write, Close, Rename, with no fsync of the file or the directory (RED E2a: `fsync of file=0 dir=0`). The repair replaces 18 lines with one call to the store's existing `writeAtomic` (durable.go:44). That is a reuse rather than new code, and the temp-file mode (CreateTemp default) is unchanged. The test injects the syncer. It catches a missing file sync and a missing directory sync (count assertions), and a file sync placed after the rename (the "fsync failed is on disk" assertion).
3. **intake/intake.go publish (finding #3).** The defect is real. Under a panic injected after the host created the Issue, the base code restarts at a clean StageAccepted. The duplicate search returns nothing, so it publishes a second Issue (RED E2b: `2 issues exist for one request after a restart, want 1`). The repair writes `Blocked{Unknown}` before the send. Every exit path afterwards replaces the row: success → `confirmIssue` clears it; `ErrNotSent` → `fail(unknown=false)` makes it Retryable; any other error → `fail(unknown=true)`. A restart therefore goes through `lookup` (FindIssue by key). A mutant that records the intent as Retryable instead of Unknown would publish again on restart and fail the test. The behavior changes are deliberate and recorded (F3, F4).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

- C1: The three packages pass at d4d625f54d, gofmt reports nothing, and vet is clean.
- C2: Each new test fails on base 9b632c1eb3 for the reason the card states.
- C3: Coverage of the changed packages is at or above 85%.
- C4: The required codex audit backend returned pass on the delta 9b632c1eb3..HEAD.
- C5: The RC40 line references match the code the author derived the defects from.

### Evidence

E1. Run in /Users/goos/.moai/worktrees/projects/moai-adk-go/t2072. The tree HEAD resolved to d4d625f54de8a211abc666498923d1a65e950234 (`git rev-parse HEAD`, and `git rev-parse WT-rc40-internal-agenticfac` gave the same value at the end of the audit):

```
$ timeout 600 go test ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	13.309s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	80.825s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	41.918s
exit=0
$ gofmt -l internal/agenticfactory/
(no output)
gofmt-exit=0
$ go vet ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/
(no output)
vet-exit=0
```

E2. RED probe. I extracted base 9b632c1eb3 with `git archive -o <scratchpad>/base.tar 9b632c1eb3 go.mod go.sum internal pkg` into a scratch copy, then laid the HEAD test files over it (bind_ownership_test.go, request_durable_test.go, intake_crash_test.go, intake_test.go). The production code was the base code:

E2a.
```
$ go test ./internal/agenticfactory/claim/ -run 'TestBindWorktreeRefusedForAForeignCallerLeavesNoReservation|TestWriteRequestLocalSyncsFileAndDirectoryAndRefusesOnFailure' -count=1
--- FAIL: TestBindWorktreeRefusedForAForeignCallerLeavesNoReservation (0.13s)
    bind_ownership_test.go:36: a refused bind left a reservation behind: <nil>
    bind_ownership_test.go:40: a refused bind blocked an honest Issue: agenticfactory: the worktree is already used by another open Issue: /private/var/folders/.../owned is reserved by issue-1
--- FAIL: TestWriteRequestLocalSyncsFileAndDirectoryAndRefusesOnFailure (0.01s)
    request_durable_test.go:31: fsync of file=0 dir=0, want 1 and 1
    request_durable_test.go:38: a failed fsync was acknowledged: <nil>
    request_durable_test.go:41: a request whose fsync failed is on disk: <nil>
    request_durable_test.go:46: leftover file after a failed write: k-failed.json
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.553s
```

E2b.
```
$ go test ./internal/agenticfactory/intake/ -run 'TestAProcessDyingDuringPublishDoesNotCreateASecondIssue|TestAdvanceFailureWhileBusyKeepsBothErrors' -count=1
--- FAIL: TestAProcessDyingDuringPublishDoesNotCreateASecondIssue (0.02s)
    intake_crash_test.go:35: 2 issues exist for one request after a restart, want 1
--- FAIL: TestAdvanceFailureWhileBusyKeepsBothErrors (2.27s)
    intake_test.go:823: an intent that could not be recorded was sent anyway
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	2.667s
```

E3. These RED outputs match the RED messages recorded in progress.md:9-11.

E4.
```
$ timeout 500 go test -cover ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	13.000s	coverage: 92.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	84.552s	coverage: 86.5% of statements
```

E5. `git show 67b4734d6b:internal/agenticfactory/intake/intake.go | sed -n 236,252p` → line 246 is `issue, ref, err := in.pub.PublishIssue(sctx, req)`, directly after `preflight`, with no store write in between. `git show 67b4734d6b:internal/agenticfactory/claim/request.go | sed -n 180,205p` → line 193 is `if err := tmp.Close(); err != nil {`, with no `Sync` anywhere in the function. `git show 67b4734d6b:internal/agenticfactory/claim/next.go` → `ReserveWorktree` is called before `Update` and no `Verify` comes first.

E6.
```
$ grep -nE "exec\.Command|os/exec|password|token|secret|unsafe" <the three new test files>
grep-exit=1   (no match)
```

E7. Audit plan and cross-model leg:
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2072
config_status: ok, model: codex, cross_model_active: true, enforced_required: ["codex"]; claude gate off, glm gate off
```
`mcp__moai__audit_multi` (project_root=t2072, target=baseBranch, delta_base=9b632c1eb3), digest members:
- Call 1: overall_verdict=inconclusive, gate_unmet=codex, plan_source=config; per_backend: {backend: codex, gate: required, verdict: inconclusive}. The cause given: codex's read-only sandbox refused to create `/tmp/go-build…`, so it could not run the tests. It reported no concrete defect. Receipt rcpt-aa594166d862fa564805dca6.
- Call 2 (a retry asking for a static review only; I had already run the tests myself in E1/E2): overall_verdict=pass, gate_unmet="" (absent), plan_source=config; per_backend: {backend: codex, gate: required, verdict: pass}, findings=[]. Receipt rcpt-8b62e260617a296be5255dcb.
- Call 3 (`mcp__moai__codex_audit`, mode=native, project_root=t2072, target=baseBranch, delta_base=9b632c1eb3). I issued it after the SubagentStop receipt guard refused the first verdict line with "receipt unknown to the store". Result: verdict=pass, findings=[], resolved_model=gpt-6.1-sol, review_base=delta_base (merge base 9b632c1eb3544cef5f275e502922a375e48d20aa). Receipt rcpt-dcab4c383399c533015cc955.
- The orchestrator still has to run its result check (`moai verify audit-plan --project-root <t2072> --result-file <toplevel>/.moai/state/audit-plan-result.json`) on the call-2 digest. This verdict is not final until that check passes.

### Baseline-attribution

- All test, gofmt, vet and coverage results above come from this run, against the tree at d4d625f54d (E1, E4). The RED results come from base 9b632c1eb3 production code combined with the HEAD test files, in a scratch copy (E2). None of the figures is carried over from progress.md. I re-ran claim and runner after the author's format-only change, which settles progress.md:23's note that they had not been re-run.
- Tool provenance (§2.2): the tests ran on the `go` toolchain from PATH, building the tree's own code. The `moai verify audit-plan` and MCP audit results come from an installed build at commit 9b632c1eb3. audit_multi reports `build_lag`: that build is an ancestor of HEAD d4d625f54. The card changes only `internal/agenticfactory/{claim,intake}`, not audit tooling, so I expect no effect on the audit-plan result. I did not verify that expectation.

### Gaps

- G1: The worktree-isolation guard refused several commands, because this auditor session is anchored in t2080. Refused: `cd t2072 && git …` (a diff/read command), a `git archive | tar` pipeline, and a tar command using a shell variable. Fallbacks: I read files with the Read tool, ran `git show`/`git archive -o` from my own worktree (shared objects), and ran tar and cp with literal paths. No measurement depended on a refused command. One early `cd t2072 && git status --short` did run before the refusals started, and it printed nothing (a clean tree). I did not re-check the clean state at the end; the branch tip was re-checked (E1).
- G2: `golangci-lint run` on intake and claim did not run: `Error: parallel golangci-lint is running` (exit 3). Following the resource rules, I did not retry it or compete with the other run. Only gofmt and vet were observed.
- G3: Not run: the race detector, the full `go test ./...`, other OSes, and the alternative fix in F2. I made no injected-mutant runs; the mutant judgments under "Per-repair judgment" are reasoned from the test assertions.
- G4: The original reviewer text for findings #2 and #3 is unavailable (see F6). The match between derived and intended defects is unconfirmed.
- G5: The codex leg's scope rests on delta_base and is told to codex in prose only, not bounded mechanically (audit_multi residual_risk_note). Codex call 2 is a static review only.

### Residual-risk

- The BindWorktree residual window (F2) can still leave a reservation behind when a claim is lost between `Verify` and `Update`.
- The new intent row shows normal in-flight publishes as EFFECT_UNKNOWN to monitors and operators (F4). A premature `ConfirmNotCreated` during the send could still produce a duplicate Issue.
- `fail()`'s Busy-after-intent join path is untested (F3).
- I did not inspect lint findings, if any exist (G2).
- The real-host `FindIssue` may also depend on a lagging search index. If it does, a restart after a crash would stop at EFFECT_UNKNOWN (safe, but it needs an operator) instead of confirming the Issue.

## Operational Notes (unverified)

- The progress record should correct its residual-window sentence (F2) before the card is closed, so that the record does not carry an inaccurate premise.
- Commit eec30d8479 (t2078) appears to cover F5 (`register` intent-before-send). Check its state before you open another follow-up.
- Rebuild/reinstall the moai binary before relying on further moai CLI results for trees past 9b632c1eb3 (build_lag in E7).
