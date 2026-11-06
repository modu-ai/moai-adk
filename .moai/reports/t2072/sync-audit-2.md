auditor-model: claude-opus-5-5

# t2072 sync audit, iteration 2: merge-level re-audit of the absorbed HEAD

verdict: PASS
audited_sha: 37fb4650d73b6450d9f39c94acbe751238939c03

## Evaluation Report

- Card: t2072 (Class B, three carried-over RC40 review findings repaired; no SPEC)
- Scope: the no-ff merge 37fb4650d7 (first parent 08fd89b85e = card; second parent 04d8c27a7e = local main)
- Overall: PASS. Score 0.91. No blocking finding. Both must-pass dimensions (Functionality, Security) pass.
- Prior art: iteration 1 PASS 0.91 at d4d625f54d (`.moai/reports/t2072/sync-audit.md`). Read as context only. Each claim below was measured again in this run.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | Three packages pass at HEAD (`ok intake 19.918s`, `ok claim 173.017s`, `ok runner 76.625s`, exit=0). With the three production hunks reverted in a probe copy of HEAD, all four card tests fail as expected (E6). |
| Security (25%) | 92/100 | PASS | No new input surface. The changes tighten an authorization order (ownership is verified before the reservation), make a write durable (fsync of file and directory), and remove a duplicate side effect (intent recorded before the send). The main side touches no file in these packages (E3). |
| Craft (20%) | 88/100 | PASS | `gofmt -l internal/agenticfactory` printed nothing (exit 0). `go vet` on intake and claim printed nothing (exit 0). The four carried optional findings from iteration 1 are unchanged. Coverage was not measured (Gap). |
| Consistency (15%) | 92/100 | PASS | `writeRequestLocal` now uses the package's shared `writeAtomic` (durable.go:44). The intent row uses the same `Blocked{Unknown:true}` / Resume wording as `fail()` (intake.go:389-391). |

### Findings (structured defect list)

No new blocking finding. The merge introduced nothing. Iteration 1's F1 to F6 still apply word for word, because the card's code is byte-identical (E1). They are carried below as optional.

- F1 [low] [optional] [in-diff] claim/bind_ownership_test.go:10-13. The comment says "the real holder was then refused its own path". What is actually refused is another Issue. Carried from iteration 1. Fix: reword the comment.
- F2 [low] [optional] [record] progress record. The claim that the Verify-to-Update window is closed by the stale-reservation takeover is inaccurate. The window is narrow and matches the documented abandoned-reservation contract (next.go:181-185). Carried. Fix: correct the record. A follow-up may move ReserveWorktree under the Issue lock.
- F3 [low] [optional] [in-diff] intake_test.go:810-825. `TestAdvanceFailureWhileBusyKeepsBothErrors` no longer checks what its name says. Carried. Fix: rename the test, or add a case where the store goes Busy after the intent write.
- F4 [low] [optional] [in-diff] intake.go:248-254. A publish that is in flight now reads EFFECT_UNKNOWN. A crash after the intent write and before the send needs an operator `ConfirmNotCreated`. This is the accepted liveness cost. Carried.
- F5 [medium] [optional] [outside] intake.go `register` has the same crash window. The follow-up commit eec30d8479 (t2078) exists but is NOT an ancestor of HEAD (E8). It touches intake.go and intake_test.go, so landing it after this card is a textual-overlap risk. That risk is not a defect of this merge.
- F6 [info] [optional] [outside] The original reviewer text for findings #2 and #3 is still not on disk. Carried.
- F7 [low] [optional] [process] The audit's probe copy (77M) at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/probe` could not be deleted: `rm -rf` was refused by the permission system (Gaps G4). Fix: the lead deletes that directory.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

C1. The merge 37fb4650d7 did no conflict resolution of its own. Its content is exactly the union of its two parents.
C2. What HEAD adds on top of main is exactly the card's own change d4d625f54d plus the two report files the card committed (sync-audit.md, verdict.md). There are no extra files and no altered hunks.
C3. No commit on the main side between the merge base and 04d8c27a7e touches `internal/agenticfactory/`. The homestate change main brought in (t2116) touches no symbol these packages use.
C4. The three repairs are present at HEAD: BindWorktree runs Verify before ReserveWorktree; writeRequestLocal goes through writeAtomic; publish records an Unknown intent row before PublishIssue.
C5. The repairs are effective at HEAD: the three packages pass, and reverting only the three production hunks on a copy of HEAD makes the card's tests fail.
C6. Format and vet are clean on the changed packages.

### Evidence

E1 (C2). The stats agree, and the code patches are byte-identical:
```
$ git diff 04d8c27a7ed7 37fb4650d7... --stat
 .moai/reports/t2072/sync-audit.md                  | 150 +++++++++++++++++++++
 .moai/reports/t2072/verdict.md                     | 150 +++++++++++++++++++++
 .../agenticfactory/claim/bind_ownership_test.go    |  42 ++++++
 internal/agenticfactory/claim/next.go              |   6 +
 internal/agenticfactory/claim/request.go           |  19 +--
 .../agenticfactory/claim/request_durable_test.go   |  49 +++++++
 internal/agenticfactory/intake/intake.go           |  11 ++
 .../agenticfactory/intake/intake_crash_test.go     |  37 +++++
 internal/agenticfactory/intake/intake_test.go      |  21 ++-
 9 files changed, 462 insertions(+), 23 deletions(-)

$ git diff 9b632c1eb3 d4d625f54d... --stat
 (the same 7 internal/ files, the same per-file counts)
 7 files changed, 162 insertions(+), 23 deletions(-)

$ git diff 9b632c1eb3 d4d625f54d -- internal/ > card.diff
$ git diff 04d8c27a7ed7 37fb4650d7 -- internal/ > merged.diff
$ cmp card.diff merged.diff; echo cmp_exit=$?
cmp_exit=0
     269 card.diff
     269 merged.diff
```
The difference between the two stats is exactly the two report files, and those are byte-identical across the merge:
```
$ git rev-parse 08fd89b85e:.moai/reports/t2072/sync-audit.md 37fb4650d7:.moai/reports/t2072/sync-audit.md 08fd89b85e:.moai/reports/t2072/verdict.md 37fb4650d7:.moai/reports/t2072/verdict.md
2ee70016704b9b74c334862aa45db94667a6f3b7
2ee70016704b9b74c334862aa45db94667a6f3b7
2ee70016704b9b74c334862aa45db94667a6f3b7
2ee70016704b9b74c334862aa45db94667a6f3b7
```
(verdict.md and sync-audit.md are the same blob, which fits commit 08fd89b85e's message "copy of the independent sync-audit".)

E2 (C1). The merge base is the card's base, and the combined diff of the merge is empty:
```
$ git merge-base 08fd89b85e 04d8c27a7ed7
9b632c1eb3544cef5f275e502922a375e48d20aa
$ git show --cc --format='parents=%P' 37fb4650d7...
parents=08fd89b85ec87d021c1067982f952d914a26b78c 04d8c27a7ed77afc09217f8623519806d48df3c4
(no diff body follows)
```

E3 (C3). The main side does not touch the packages:
```
$ git log --format='%h %s' 9b632c1eb3..04d8c27a7ed7 -- internal/agenticfactory/
(no output)
$ git diff --stat 08fd89b85e 37fb4650d7   (what the merge brought from main)
 24 files changed, 1477 insertions(+), 42 deletions(-)
 (internal/cli/*, internal/config/*, internal/homestate/card_lease_expired_owner*.go, .moai/*, CHANGELOG.md; none under internal/agenticfactory/)
$ grep -nE 'agenticfactory/(claim|intake)' internal/cli/factory_card_rebind.go internal/cli/factory_release_route.go internal/cli/todo_landed.go internal/homestate/card_lease_expired_owner.go internal/config/types.go internal/config/loader_agentic_release_branch.go
grep_exit=1
$ grep -rhoE 'homestate\.[A-Za-z]+' internal/agenticfactory/claim internal/agenticfactory/intake internal/agenticfactory/runner | sort | uniq -c
   3 homestate.ClassifyOwnerWith
   3 homestate.OwnerClassification
   6 homestate.OwnerDead
   4 homestate.OwnerIndeterminate
   6 homestate.OwnerLive
   1 homestate.ProbeProcessIdentity
   7 homestate.ProcessIdentityDead
   3 homestate.ProcessIdentityIndeterminate
  22 homestate.ProcessIdentityLive
  11 homestate.ProcessIdentityState
$ grep -rln 'IsReacquirableExpiredLease\|ReacquireExpiredLease' internal/
internal/homestate/card_lease_expired_owner_staged_t2116_test.go
internal/homestate/card_lease_expired_owner.go
internal/homestate/card_lease_expired_owner_test.go
internal/cli/factory_card.go
```
The functions t2116 changed (IsReacquirableExpiredLease, ReacquireExpiredLease) are not used by claim, intake, or runner.

E4 (C4). Code read at HEAD:
- next.go:160-168: `if _, err := s.Verify(i, holder, epoch); err != nil { return Record{}, err }` comes before `s.ReserveWorktree(i, canon)`. `Update` (with its own verification under the Issue lock) follows.
- request.go:184: `return s.writeAtomic(s.requestsDir(), "req.tmp-*", s.requestPath(r.Key), b)`. durable.go:44-73 writeAtomic does CreateTemp → Write → `s.sync.syncFile` → Close → Rename → `s.sync.syncDir`, and every failure maps to `ErrStoreUnavailable`.
- intake.go:244-254: `in.Store.UpdateRequest(...)` sets `Blocked{Step: StageAccepted, Detail: "publish in flight", Unknown: true, ...}`. An error returns before `in.pub.PublishIssue` (intake.go:257). `fail()` (intake.go:389-396) and `confirmIssue` replace the row afterwards, as the comment states.

E5 (C5, GREEN at HEAD):
```
$ cd .../t2072 && timeout 500 go test -count=1 ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	19.918s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	173.017s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	76.625s
exit=0
```
No `No space left on device` appeared. `df` before the run showed 52Gi available on /System/Volumes/Data.

E6 (C5, RED on HEAD minus the repairs). The probe was `git archive 37fb4650d7 go.mod go.sum internal pkg | tar -x` into scratch, then `patch -R -p1` of only the next.go, request.go, and intake.go hunks of 9b632c1eb3..d4d625f54d:
```
patching file 'internal/agenticfactory/claim/next.go'
patching file 'internal/agenticfactory/claim/request.go'
patching file 'internal/agenticfactory/intake/intake.go'
--- FAIL: TestBindWorktreeRefusedForAForeignCallerLeavesNoReservation (0.16s)
    bind_ownership_test.go:36: a refused bind left a reservation behind: <nil>
    bind_ownership_test.go:40: a refused bind blocked an honest Issue: agenticfactory: the worktree is already used by another open Issue: .../owned is reserved by issue-1
--- FAIL: TestWriteRequestLocalSyncsFileAndDirectoryAndRefusesOnFailure (0.00s)
    request_durable_test.go:31: fsync of file=0 dir=0, want 1 and 1
    request_durable_test.go:38: a failed fsync was acknowledged: <nil>
    request_durable_test.go:41: a request whose fsync failed is on disk: <nil>
    request_durable_test.go:46: leftover file after a failed write: k-failed.json
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.558s
--- FAIL: TestAProcessDyingDuringPublishDoesNotCreateASecondIssue (0.01s)
    intake_crash_test.go:35: 2 issues exist for one request after a restart, want 1
--- FAIL: TestAdvanceFailureWhileBusyKeepsBothErrors (2.63s)
    intake_test.go:823: an intent that could not be recorded was sent anyway
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	3.019s
```
Each test fails for its own named defect and passes at HEAD (E5). So the repairs are still what makes them pass after main was absorbed.

E7 (C6):
```
$ gofmt -l internal/agenticfactory; echo gofmt_exit=$?
gofmt_exit=0
$ go vet ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/; echo vet_exit=$?
vet_exit=0
```

E8 (F5 context):
```
$ git merge-base --is-ancestor eec30d8479 37fb4650d7; echo ancestor_exit=$?
ancestor_exit=1
eec30d8479 [t2078] fix: record the branch-registration intent before RegisterBranch is sent
 internal/agenticfactory/intake/intake.go           | 10 +++++
 .../intake/intake_register_crash_test.go           | 46 ++++++++++++++++++++++
 internal/agenticfactory/intake/intake_test.go      |  4 ++
```

### Baseline-attribution

- Tree: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2072`. `git rev-parse HEAD` = 37fb4650d73b6450d9f39c94acbe751238939c03 at the start and again after all runs. `git status --short` was empty after the runs (the tree was unchanged).
- Toolchain: `go version go1.27.3 darwin/arm64` (an external toolchain). No project-built moai tool produced any measurement here, so the §2.2 build-provenance coordinate does not apply.
- All diffs were taken from shared git objects at the SHAs named above, in this run.

### Gaps

- G1. The cross-model leg was NOT run. The lead's hard constraint was Claude-only (stale MCP server). No audit_multi, codex, glm, or claude_audit call was made, and no receipt was issued. `moai verify audit-plan` was also not run, so the configured audit plan was not read. If the receipt guard refuses this PASS for lack of a receipt, that refusal stands. It was not worked around. Observed: the SubagentStop hook did refuse it with `AUDIT_RECEIPT_VIOLATION` (no receipt cited) and asked for a codex_audit or audit_multi call. That call was not made, because the lead's hard constraint forbids it in this run. The verdict reached here is unchanged. A receipt-backed corroboration is still outstanding and is the lead's decision.
- G2. Coverage was not measured (`go test -cover` not run), and neither was golangci-lint. Craft rests on gofmt, vet, and the tests only.
- G3. `-race` was not run. The card's tests are not concurrency tests, but the claim package has cross-process locking.
- G4. Probe cleanup was refused. `rm -rf .../scratchpad/probe` was denied by the permission system (twice: once in a compound form, once alone). The 77M probe copy remains at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/probe`. The three small .diff files in the same scratchpad also remain. No alternative deletion was attempted.
- G5. Only the three named packages were tested. The main-side packages (internal/cli, internal/config, internal/homestate) were not re-tested at the merged HEAD. They are outside this card's diff and were audited under t2116/t2086/t2098.

### Residual-risk

- R1. The ownership check (Verify) and the reservation are not under one lock. A claim lost between them can still leave a reservation (iteration-1 F2). Narrow, and documented.
- R2. A process crash after the intent write and before the send needs an operator `ConfirmNotCreated` (F4).
- R3. t2078 (eec30d8479) edits the same intake.go publish/register region and intake_test.go. When it lands after this card, a textual conflict or a semantic interaction with the publish-intent row is possible. Neither was tested here.

## Iteration history

| Iteration | SHA | Result | Scope |
|---|---|---|---|
| 1 | d4d625f54d | PASS 0.91 | the card's own change |
| 2 | 37fb4650d7 | PASS 0.91 | merge-level: union check, re-run of tests at the merged HEAD, RED probe on HEAD minus the repairs |

## Operational Notes (unverified)

- Delete the leftover probe directory named in G4.
- Land t2078 after this card and rebase it over the publish-intent hunk. Re-run `TestAdvanceFailureWhileBusyKeepsBothErrors` and both crash tests after that landing.
- If the landing verb needs a receipt-backed PASS, schedule one cross-model run once the MCP server is reconnected.
