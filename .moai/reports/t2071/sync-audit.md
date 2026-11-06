auditor-model: claude-opus-5-5

# Sync audit — card t2071 (af-flow local-main card close)

verdict: PASS-WITH-DEBT
audited_sha: 7131da04c1c0acfc9701d7b348675fed87b8a1fb

Overall: PASS-WITH-DEBT, score 0.90 (Functionality 92 / Security 95 / Craft 85 / Consistency 88).
Range audited: `9b632c1eb3..7131da04c1` (base = t2069 merge, HEAD = t2071 fix commit).
Iteration: 1 (first audit of this card; no prior audited SHA, so no repair-born findings).
Audit mode: Claude-only, in-session, as the leader's brief required. No codex / glm / audit_multi / claude_audit call was made.

- debt: D1 dispose_in=sync `.moai/docs/af-flow-landed-card-close.md` line 22 says "A lane cannot run steps 1, 3 and 4: `todo landed`, `factory approve` and `todo done`". In the numbered list that follows, those verbs are steps 3, 5 and 6. Step 1 is a read-only `git log`, and step 4 is `rebind-evidence`, which has no lane refusal in code. Change the numbers to "steps 3, 5 and 6".

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92 | PASS | Focused tests pass (verbatim below). The diff limits the product change to two refusal strings. Both hints use syntax the cobra definitions accept. |
| Security (25%) | 95 | PASS | No gate, receipt, or stored-data path changed. The card id in the new message has already passed `homestate.ValidCardID` (factory_card_rebind.go:158). No new input surface. |
| Craft (20%) | 85 | PASS | `go vet ./internal/cli` clean. The tests exercise the real verbs. Some mutants survive (F3), there is no in-test negative control (F4), and the test keeps a dead import (F7). |
| Consistency (15%) | 88 | PASS | The new code follows the file's existing `note` / `exitCodeError` idiom and carries card-tag comments. The doc has one numbering error (D1). |

Must-pass firewall (Functionality, Security): both pass.

## Findings

- F1 [low] [optional → recorded as debt D1] [in-diff] `.moai/docs/af-flow-landed-card-close.md:22`: the boundary sentence names the right verbs with the wrong step numbers ("1, 3 and 4" should be "3, 5 and 6"). Required fix: correct the numbers. Confidence: high.
- F2 [low] [optional] [in-diff] `internal/cli/factory_card_rebind.go:222-226`: `resolveRecordedLanding` returns `("", "")` both when there is no landing record and when the record has no `--sha` (line 121). In the second case the new hint says "no landing is recorded for the card", which is inaccurate: a record exists but carries no delivering commit. The prescribed action is still correct (re-record with `--sha`). Suggested wording: "no delivering commit is recorded". Confidence: high.
- F3 [low] [optional] [in-diff, tests] `internal/cli/factory_af_flow_close_test.go`: the tests pin only that the hint is present on the two intended paths. These mutants would still pass: (a) the rebind hint also appended on the verdict path (`phaseErr == nil`); (b) a hard-coded `--ref main` hint emitted for every ref, including a non-`origin/` ref or the existence-check failure; (c) the landed hint on the `unrunnable` branch. One mutant is already pinned: replacing a non-empty `note` with the hint fails `TestFactoryRebindEvidenceStageEmptyRow/still_refuses_a_recorded_landing_that_is_not_on_the_landed_ref`, which asserts "recorded landing commit refused". Confidence: high.
- F4 [low] [optional] [in-diff, tests] `TestAFFlowLocalMainCardClosesThroughFormalVerbs` has no in-test negative control showing that `todo done` refuses before `approve`. Gate engagement is shown only indirectly: by the explicit `VerifyApprovalReadonly` call and by the sibling `TestCompletionChainAcrossRunRows/an_unlanded_stage_empty_card_still_cannot_reach_the_gate`. If the gate silently treated the row as unlinked (`verifyForClose` returns nil when `!linked`), the `todo done` step would pass and only the explicit `VerifyApprovalReadonly` call would still witness the binding. Confidence: medium.
- F5 [info] [optional] [outside] Lane claim (2), "none named the missing step", is overstated. The pre-existing `factory approve` refusal (`internal/cli/leader_approval_gate.go:345`) already says "run `moai factory rebind-evidence ...` first (a squash-landed card takes its landing commit from `moai todo landed --sha`)". It lacks `--ref`, but it names the next step. This does not affect the verdict. Confidence: high.
- F6 [low] [optional] [in-diff] doc step 2 says "Check it is on local `main` and not yet on origin" but gives only the local-main command. The `origin/main` check appears only in the worked example. Confidence: high.
- F7 [info] [optional] [in-diff] `factory_af_flow_close_test.go:65` `_ = homestate.CardLeased` is the file's only use of `homestate`. It exists only to keep the import alive, and both the line and the import could be dropped. Confidence: high.

There are no blocking findings. The verdict is PASS-WITH-DEBT rather than PASS only because the card's deliverable (4), the procedure doc, contains a factual mismatch (F1/D1).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1 — only the four named files changed; the product change is limited to error text

Evidence:

```
$ git -C <wt> diff 9b632c1eb3..HEAD --stat
 .moai/docs/af-flow-landed-card-close.md    | 53 ++++++++++++++++
 internal/cli/factory_af_flow_close_test.go | 97 ++++++++++++++++++++++++++++++
 internal/cli/factory_card_rebind.go        |  5 ++
 internal/cli/todo_landed.go                | 11 +++-
 4 files changed, 164 insertions(+), 2 deletions(-)
```

The full product diff (read in this audit):
- `factory_card_rebind.go`: inside `if phaseErr != nil {`, a new `if note == "" { note = fmt.Sprintf("; no landing is recorded for the card — record it first with `moai todo landed %s --sha <delivering commit> [--ref <branch>]` ...", cardID) }` is added before the unchanged `return fmt.Errorf("%w%s", phaseErr, note)`. The control flow is unchanged: the same branch returns the same wrapped `phaseErr`. Only the suffix differs, and only when `note` was empty.
- `todo_landed.go` `validateSuppliedSHA`: on the reachability failure only, `hint` is computed with `strings.CutPrefix(ref, "origin/")` and inserted into the existing message. The return value, exit code (1) and "nothing was written" behaviour are unchanged.

No gate, receipt, store or migration code is in the diff. No existing test file is in the diff, so no existing gate test was weakened.

Hint syntax against the cobra definitions:
- `todo landed`: `Use: "landed <n>"`, `Args: cobra.ExactArgs(1)`, flags `--sha`, `--ref`, `--clear` (todo_landed.go:54-68). The form `moai todo landed <card> --sha <sha> --ref main` is valid. The ref goes through `rev-parse --verify <ref>^{commit}` and `merge-base --is-ancestor <sha> <ref>`, so a local branch name `main` is accepted. The e2e test proves this.
- `factory rebind-evidence`: `Use: "rebind-evidence <card>"`, flags `--reason`, `--run` (factory_card_rebind.go:79-91).
- `factory approve`: `Use: "approve <card-id> --run <run-id>"`, flags `--run` (required at runtime, line 296) and `--issuer` (optional; defaults to the worker env var, then "leader", lines 347-352).

Baseline-attribution: tree at `7131da04c1c0acfc9701d7b348675fed87b8a1fb` (`git rev-parse HEAD` observed this value in this audit); diff base `9b632c1eb3`.

### Claim 2 — the focused tests pass

Evidence (verbatim result lines):

```
$ go test ./internal/cli -run 'TestAFFlow|TestFactoryRebindEvidenceStageEmptyRow|TestCompletionChainAcrossRunRows' -count=1 -v
--- PASS: TestAFFlowLocalMainCardClosesThroughFormalVerbs (3.08s)
--- PASS: TestAFFlowRefusalsNameTheLandingStep (4.61s)
    --- PASS: TestAFFlowRefusalsNameTheLandingStep/rebind_without_a_landing_record_names_todo_landed_with_ref (1.66s)
    --- PASS: TestAFFlowRefusalsNameTheLandingStep/landed_on_the_default_ref_names_the_local_branch_form (2.95s)
--- PASS: TestFactoryRebindEvidenceStageEmptyRow (5.65s)
    --- PASS: .../rebinds_from_the_recorded_landing_without_inferring_a_phase (2.19s)
    --- PASS: .../still_refuses_without_a_landing_record (1.64s)
    --- PASS: .../still_refuses_a_recorded_landing_that_is_not_on_the_landed_ref (1.81s)
--- PASS: TestCompletionChainAcrossRunRows (4.54s)
    --- PASS: .../rebind_and_approve_on_the_current_run_row_satisfy_the_done_gate (1.66s)
    --- PASS: .../a_receipt_for_an_older_run_is_still_refused_and_approve_says_so (1.42s)
    --- PASS: .../an_unlanded_stage_empty_card_still_cannot_reach_the_gate (1.46s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	19.889s
```

```
$ go vet ./internal/cli
(no output, exit 0)
```

Baseline-attribution: run in this audit against the worktree at HEAD `7131da04c1`, which `git status` reported clean at session start. The test binary was compiled from this tree by `go test`, so this is not an installed-binary measurement and §2.2 tool-provenance does not apply.

### Claim 3 — `TestAFFlowRefusalsNameTheLandingStep` would fail without the two product edits

Evidence (diff inspection, not executed): at base, the rebind stage-empty refusal is `fmt.Errorf("%w%s", phaseErr, "")` with an empty `note`, so the text lacks "moai todo landed t1 --sha". The base landed reachability message is "commit %s is not reachable from %s; nothing was written", with no "--ref main". Both subtests assert these substrings with `strings.Contains`, so both would fail at base. Judgment: RED is real for this test.

Gap: not executed at base, because the brief forbids modifying tracked files and checking out.

### Claim 4 — the formal chain already closes such a card at base (the e2e test is a regression guard)

Evidence (diff inspection): the only changes on the success path are none. Both edits are inside refusal branches, so `TestAFFlowLocalMainCardClosesThroughFormalVerbs` exercises code identical to base. Judgment: the lane's claim (1) is consistent with the diff.

Gap: not executed at base (same reason as above).

### Claim 5 — the e2e test exercises the real verbs and the real done gate

Evidence (read in this audit):
- `runTodo` builds `newTodoCmd()` and calls `cmd.Execute()` (todo_test.go:46-58). `runFactory` and `rebindRun` drive the factory cobra commands the same way. `todo landed`, `factory rebind-evidence`, `factory approve` and `todo done` are therefore the production command paths.
- The fixture (`squashLandedFixture` → `fcFixture`) creates a real factory DB and records the dispatch binding (`RecordDispatchBinding`). `emptyStageRow` sets `state=leased, stage=''` by SQL to model the af-flow row shape. This is fixture setup, not a stub of the code under test.
- `todo done` reaches `holdDoneApprovalGate` → `verifyForClose` → `gate.Verify` (leader_approval_gate.go:63-103), which is not stubbed. The test additionally calls `db.VerifyApprovalReadonly` directly before `todo done`.
- The af-flow shape is real git: `afFlowLocalMainRoot` commits the merge on `main` after pinning `refs/remotes/origin/main` to the previous HEAD, so the merge is reachable from `main` and not from `origin/main`. The second subtest confirms this through the refusal "is not reachable from origin/main".

Gap: see F4. There is no in-test proof that `todo done` would refuse without `approve`.

### Claim 6 — doc commands and boundary statements

Evidence:
- The commands in steps 3-6 and in the worked example match the cobra definitions above. `moai factory status` exists (factory_card.go:4537, `Use: "status"`).
- Lane boundaries in code: `factory approve` refuses a lane via `factoryLaneRefusal()` (leader_approval_gate.go:290). `todo landed` and `todo done` refuse a lane via `todoRefuseLaneMutation` (todo.go:455-466), which allows only the read-only verbs list. `factory rebind-evidence` (`runFactoryRebindEvidence`) has no lane refusal. The doc correctly omits it from the lane-forbidden verbs but mis-numbers the steps (D1).
- The doc states that the worked example was not executed: "Worked example (read-only facts, not executed by the lane)" and "The real t2055 chain above was not executed by this lane."
- I checked the worked-example git facts in this audit:

```
$ git log -1 --format='%H %s' 7f1ab8d42e
7f1ab8d42eee7b3bee3bde3aa707539ad5537289 [t2055] merge: lane landing closes the card at the local main merge into the RC41 candidate (card t2055)
$ git merge-base --is-ancestor 7f1ab8d42e main          -> exit 0
$ git merge-base --is-ancestor 7f1ab8d42e origin/main   -> exit 1
```

Baseline-attribution: refs as they stood in this repository at audit time. `main` and `origin/main` are moving refs, so these results are the subject at audit time, not a pinned fact.

## Gaps

- Neither TestAFFlow test was executed at base `9b632c1eb3`. The RED and already-works-at-base judgments come from the diff only, because the brief forbade checkout and tracked-file edits.
- The worked example's run id `to0ck1` and the t2055 row state (`leased`, stage `-`) were not verified. I did not run `moai factory status`; the only checks were the git ones above.
- The real t2055 chain against the live queue was not executed, by the lane or by this audit.
- `moai verify audit-plan` was not run, and no cross-model MCP tool (codex / glm / audit_multi / claude_audit) was called, because the leader's brief mandated a Claude-only audit. If the receipt guard requires a codex receipt at stop, this audit does not supply one: receipts=none.
- No coverage figure was measured; the brief limited test runs to the one focused command. No golangci-lint run.
- The whole `internal/cli` suite was not run, under the lane rule.
- No tool refusals occurred during this audit. Two grep invocations failed on zsh glob expansion of `--include=*.go` and were re-run without that flag; this does not affect the measurements.

## Residual-risk

- The surviving mutants in F3 mean a later edit could add the hint to the verdict-path refusal, or emit `--ref main` for a non-origin ref, and no test would catch it.
- The landed hint fires for any commit unreachable from `origin/<branch>`, including one that is on no branch at all. The wording is conditional ("a card merged into local main ..."), and a wrong `--ref main` attempt would itself be refused by the same reachability check, so a wrong record cannot result. The only risk is a misleading suggestion.
- A landing record observed before the run's engagement start is refused by `resolveRecordedLanding` ("belongs to an earlier issuance"). The doc does not mention this refusal. A reissued rc.41 card that already carries an older landing record would need the record re-made, which the doc does not explain.
- The doc's moving-ref statements ("not yet on origin") are true at authoring time only.
