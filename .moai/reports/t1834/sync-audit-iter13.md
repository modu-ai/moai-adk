auditor-model: glm-5.3-flash

# Sync audit iter13 (post-repair confirmation) — card t1834, SPEC-INTEGRATION-STALE-BASE-001

verdict: PASS-WITH-DEBT
audited_sha: 1a16da10f6edc7b78c22bdb8460908108ab220b9

Overall: PASS-WITH-DEBT. The two iter12 blocking findings are closed at this HEAD — F1 by the `record.Tree == tipTree` identity comparison inside `factoryAdoptedTip` (factory_card.go:2889), F2 by the `git symbolic-ref --quiet HEAD` read with only `refs/heads/` stripped (factory_card.go:3047) — each with a regression test that names its reverse case, both observed green here. The required codex gate returned `pass` with no findings on this HEAD. `go build`, `go vet`, and all ten scoped test runs are green with real exit codes. Two named debts ride with the pass, neither blocking, neither this card's repair to make: the pre-existing same-name-tag defect on the merge-step path (routed separately in §E.2.8) and a precision defect in the F2 fix comment (measured below).

## Audit digest (audit_multi, project_root = tree, card_id = t1834, target = baseBranch, no gates argument, no claude_verdict — GLM main session, fresh backends)

```
overall_verdict: pass
gate_unmet: (absent)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: pass}]
disagreement_flag: null
participant_count: 1
fail_open_backends: []
audit_receipt: rcpt-5fd438dcfae9026bf8fb0732
build_commit: 8249bff8c3
```

One line: `overall_verdict=pass gate_unmet=none plan_source=config codex/required/pass receipt=rcpt-5fd438dcfae9026bf8fb0732`.
This verdict is not final until the orchestrator's `moai verify audit-plan --project-root <toplevel> --result-file` convergence check passes; the digest members above are what it reads. Plan read first: `moai verify audit-plan` printed `config_status: ok`, model `codex`, claude `off`, codex `required` (explicit), glm `off`, `cross_model_active: true`, `enforced_required: [codex]` — so audit_multi ran with no `gates` argument and the tree's own plan applied. The disagreement flag is `null`, not `false`: one participant, no observed divergence — an undetermined flag, not an agreement claim.

Receipts of THIS audit (my own calls only): rcpt-5fd438dcfae9026bf8fb0732. The audit_multi call appended one second-review record into the card evidence directory (card_id t1834, head 1a16da10f6). Stale receipts from earlier rounds (rcpt-db0ad702d7de15f1c9f12afc, rcpt-6040618efd97e27498747772) are not evidence here and are not cited.

Judging build: the moai MCP server that served the codex leg reports build_commit 8249bff8c3; `git merge-base --is-ancestor 8249bff8c3 HEAD` exits 0, so it IS a strict ancestor of the tree HEAD 1a16da10f6 — the gate's own handling code predates the repairs and could not have run them. The codex review itself reads the diff and runs overlay tests against the tree, so its pass describes this tree; a pass under build lag is not discounted, but this attribution is stated per verification-claim-integrity §2.2.

## Findings (structured defect-list)

No blocking findings. Named debts:

- D1 [P2] [optional — routed, pre-existing] internal/factory/integration_merge_step.go:950 - the merge step reads its branch with `rev-parse --abbrev-ref HEAD` (the same class as iter12 F2); a tag sharing the card branch's name makes the non-adoption complete path refuse ("is on heads/WT-alpha, not a WT- branch"). I confirmed the line by reading: the call is unchanged at this HEAD and the file is NOT among the card's 19 paths (scope discipline held — the plan-pin preserve check the lane recorded matches the diff I measured). Observed live by the lane during the F2 work, disclosed in progress.md §E.2.8, routed for separate repair; the required codex gate raised nothing on it at this HEAD. Not this card's repair to make; it rides with the pass as a debt. - Required fix (for the routing owner, not this card): `git symbolic-ref --quiet HEAD` with only `refs/heads/` stripped, mirroring factory_card.go:3047.
- D2 [P3] [optional] internal/cli/factory_card.go:3044-3046 - the F2 fix comment states a detached HEAD "like the former --abbrev-ref error path, it stays ''". Measured false in one particular: on a detached HEAD the FORMER read did not take the error path — `git rev-parse --abbrev-ref HEAD` exits 0 printing the literal `HEAD` (measured in a scratch repo: abbrev-ref → `HEAD`, rc=0; symbolic-ref --quiet → no output, rc=1) — so the former return value was `"HEAD"`, not `""`. The OUTCOME is unchanged on all three callers (factory_card.go:2598 and 2680: `"HEAD"` and `""` both fail the refs/heads lookups, so the adoption exemption does not fire and the same refusals print; factory_card.go:3020: both fall into the caller-tree refusal; factory_card_pr.go:210: both hit the not-on-a-card-branch refusal). Consequence two: the `cardBranch == "HEAD"` disjunct at factory_card_pr.go:210 is now unreachable (symbolic-ref never returns the bare `HEAD`). No reachable behavior change; comment precision and one dead guard clause. - Required fix: correct the comment sentence to say the former read returned the literal `HEAD` and the new read returns `""` (same refusals either way); the dead disjunct may be left or removed at the routing owner's discretion.

Observations carried without action (same as iter12's F3/F4 class): the other `--abbrev-ref` readers named in the brief (hook_install.go:87, session_worktree.go:768/870, session_worktree_automerge.go:311, internal/cli/integration.go:126) are untested observation-only surfaces; the explicit-`--branch` refusal behavior follows the approved 8372 reduction.

## Closure of iter12 F1 and F2 (reading + tests + gate)

- F1 (adoption identity): closed. factory_card.go:2889 requires `record.Tree == tipTree` inside the adoption condition, and the doc comment (2875-2883) states the mirror of the gate's own verdict (the gate side keeps its own `rec.Tree == res.Tree`, integration_base_freshness.go via `EvaluateBaseFreshness`; both paths call this one helper per integration_base_freshness.go:146, so they cannot drift). Regression `TestFactoryCompleteAdoptionRequiresRecordTreeIdentity` (integration_base_freshness_test.go:465) pins BOTH directions: a record stored under the tip-tree key whose `tree` field names the first parent's tree is refused (`requireFreshnessRefusal`), and a healthy record naming its key tree still completes merged-local. The pure-gate direction is pinned by `TestBaseFreshnessRefusesRecordNamingAnotherTree` (internal/factory/base_freshness_test.go:700), which ran green in the factory Freshness scope. REQ-ISB-016's "a valid re-measure record keys the tip's tree" reads consistently with the fixed predicate: a record naming another tree is by the gate's own verdict not a valid record for the tree, which is exactly what the mirror comment says.
- F2 (same-name-tag branch read): closed. factory_card.go:3047 reads `git symbolic-ref --quiet HEAD` and strips only the `refs/heads/` prefix; `gitenv.Env()` guards the child; error → `""` (best-effort, as before). Regression `TestFactoryBranchOfWorktreeWithSameNameTag` (factory_card_pr_gitenv_test.go:82) pins both directions: resolution returns the plain branch name under a same-named tag, and a same-named tag keeps the adoption-path `complete` healthy end-to-end (merged-local on the pre-made merge, no sentinel). The rescope of the end-to-end subtest from the merge-step path to the adoption path is legitimate: the merge-step path carries the pre-existing D1 defect a same-named-tag run would trip regardless of this card's fix, the file is outside the card's 19 paths, and no SPEC-artifact text promises a merge-step same-name-tag test (checked: plan.md lines 20/22/60/66 and acceptance.md name the remeasure-side AC-ISB-006a only; the full-HEAD-ref language already lives in REQ-ISB-001 and plan.md §60/66 for the gate side). The rescope reason and the disclosure are recorded in §E.2.8 and in commit 1eaf9bb649's subject.

## Scope, repairs, and text consistency

- The card's non-.moai diff against main is exactly 19 paths (iter12's 18 + `internal/cli/factory_card_pr_gitenv_test.go`), measured with `git diff main...HEAD --name-only -- . ':!.moai'`; the factory-dispatch-gates.md pair is byte-identical (`diff -q` rc=0).
- The five repair commits carry the `🗿 MoAI` marker and `card t1834`; the merge 2439da8c90 absorbed local main (`git merge-base --is-ancestor main HEAD` rc=0). Repair commits touched exactly: integration_base_freshness_test.go (F1 test), factory_card.go (F1 fix), factory_card_pr_gitenv_test.go (F2 test + rescope), factory_card.go (F2 fix) — no SPEC-artifact or CHANGELOG edits, matching §E.2.8. The CHANGELOG's adoption sentence and REQ-ISB-016's wording remain true summaries of the fixed predicate (the identity requirement is stricter than the text, never looser).
- §E.2.8 claims verified against the tree: factory_card.go:2889 is the `record.Tree == tipTree` line; factory_card.go:3047 is the symbolic-ref call; 19-path delta; commit table matches `git log`.
- One transparency item re-checked and accepted: 26dad7f9ef was re-created via `git reset --soft HEAD~1` seconds after first creation (mangled trailer repair). No `--amend`; the commit graph (26dad7f9ef RED before e9e95f4d40 GREEN; 7891703291/1eaf9bb649 RED before 1a16da10f6 GREEN) still witnesses RED-before-GREEN; the disclosure is recorded in §E.2.8 before this audit read it. Nothing cited the broken form.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. HEAD is 1a16da10f6edc7b78c22bdb8460908108ab220b9 on WT-10-17-class-8; local main is fully absorbed (ancestor rc=0). The working tree is not literally clean: `.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/progress.md` carries one uncommitted modification (+22 lines, the §E.2.8 record this audit verifies, a .moai path, outside the audited diff). No non-.moai dirt exists.
2. The card's non-.moai diff is exactly the 19 paths; the gates pair is byte-identical.
3. The repairs close iter12 F1 and F2; reverse cases are pinned by tests that ran green at this HEAD; no repair-introduced defect was found (the forged record is refused on both the pure gate and the complete path; the healthy record is adopted; same-named-tag resolution returns the plain branch and the adoption path stays healthy; detached HEAD produces the same refusals as before on every caller).
4. The required codex gate passes on this HEAD (receipt rcpt-5fd438dcfae9026bf8fb0732).
5. `go build ./...`, `go vet` on the three packages, and all ten scoped test runs are green with real exit codes; no failing test name exists in any JSONL.

### Evidence
Git and plan surface:
```
git rev-parse --show-toplevel            -> /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834
git rev-parse HEAD                       -> 1a16da10f6edc7b78c22bdb8460908108ab220b9
git status --short                       ->  M .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/progress.md  (only entry)
git merge-base --is-ancestor main HEAD   -> rc=0
git diff main...HEAD --name-only -- . ':!.moai' -> 19 paths (listed: factory-dispatch-gates.md, CHANGELOG.md,
   factory_card.go, factory_card_pr_gitenv_test.go, integration.go, integration_base_freshness.go,
   integration_base_freshness_test.go, config/{cache,defaults,types}.go, config/testdata/shipped_key_inventory.yaml,
   config/workflow_base_freshness_gate_test.go, factory/{base_freshness.go,base_freshness_target_test.go,
   base_freshness_test.go,integration_remeasure.go,integration_remeasure_gitenv_test.go},
   template/templates/{factory-dispatch-gates.md,workflow.yaml})
diff -q <gates rule> <template mirror>   -> rc=0
moai verify audit-plan --project-root <tree> -> config_status ok; model codex; claude off, codex required (explicit),
   glm off; cross_model_active true; enforced_required ["codex"]
git merge-base --is-ancestor 8249bff8c3 HEAD -> rc=0 (judging build is a strict ancestor)
```
Go, entry readings taken before each command (slot / 1-min load / mem free):
```
go build ./...    slot free, load 14.56, mem 84%  -> no output, GO_RC=0
go vet ./internal/cli/ ./internal/factory/ ./internal/hook/
                  slot free, load 14.44, mem 85%  -> no output, GO_RC=0
```
Test-list pins (each `go test -list <pattern> <pkg>`, GO_RC=0, names quoted verbatim):
- ./internal/factory/ Freshness: 16 names — TestBaseFreshnessRequestedBranchOtherThanTheTargetIsRefusedNotMeasured, ...AdvisoryWhenGateOff, ...EqualToTheTargetIsMeasuredAsBefore, ...IntegrationTreeIsNotApplicableWhenATagSharesTheBranchName, ...CandidateOffTheIntegrationBranchIsStillMeasuredWithAShadowingTag, ...RefusalRemedyNamesTheBranchPerFlow, ...KBoundary, ...RequiresValidRecordForCurrentTree, ...FailOpenCauses, ...UncertainGitAnswers, ...GateOffAdvisesOnlyPure, ...K0MatchesMergeCause2, ...GateWritesNothing, ...ExitCodeIsDistinct, ...DecisionCodeHasNoThresholdLiteral, ...RefusesRecordNamingAnotherTree.
- ./internal/factory/ Remeasure: 16 names incl. TestRemeasureKeysRecordByOwnTreeUnderInheritedGitEnv, TestRemeasureCommandMeasuresTheCandidateUnderInheritedGitEnv, TestRunRemeasureNamesTheBranchTipWhenATagSharesTheBranchName.
- ./internal/cli/ Freshness: 13 names incl. TestFactoryCompleteAdoptionNotRefusedByBaseFreshness, TestBaseFreshnessRefusalReachesTheUserThroughTheRootErrorPath, TestBaseFreshnessOutputNeverEchoesUntrustedInput.
- ./internal/cli/ Remeasure: 5 names incl. TestRemeasureAndAcquireAgreeOnIntegrationTargetAcrossLinkedRoot, TestRemeasureThenAcquireAdmitsWhenATagSharesTheIntegrationBranchName.
- ./internal/cli/ TestFactoryCompleteAdoptionRequiresRecordTreeIdentity: 1 name. TestFactoryBranchOfWorktreeWithSameNameTag: 1 name.
- ./internal/cli/ TestAcquire: 12 names. TestIntegrationAcquire: 22 names incl. TestIntegrationAcquireRefusesStaleBase.
- ./internal/config/ BaseFreshness: 4 names (GateDefaults, LoaderRoundTrip, TemplateParity, TemplateBlockSelfCheck).
- ./internal/template/ RuleTemplateMirrorDrift: 1 name (declared fork list names factory-dispatch-gates.md at rule_template_mirror_test.go:240).

Scoped runs, one plain -run pattern per invocation, -p 1 -count=1 -json, raw JSONL under /private/tmp/t1834-iter13/, real exit codes, failing names read from the JSONL (all counts of `"Action":"fail"` are 0; every package line is `ok`):
```
factory-freshness.jsonl          -run Freshness ./internal/factory/     slot free, load 12.49, mem 85% -> GO_RC=0, 0 fail, ok 8.841s
factory-remeasure.jsonl          -run Remeasure ./internal/factory/     slot free, load 12.05, mem 85% -> GO_RC=0, 0 fail, ok 9.710s
cli-freshness.jsonl              -run Freshness ./internal/cli/         slot free, load 15.05, mem 84% -> GO_RC=0, 0 fail, ok 43.173s
cli-remeasure.jsonl              -run Remeasure ./internal/cli/         slot free, load 36.38 (DEVIATION, see Gaps), mem 84% -> GO_RC=0, 0 fail, ok 7.394s
cli-f1-identity.jsonl            -run TestFactoryCompleteAdoptionRequiresRecordTreeIdentity ./internal/cli/
                                                        slot free, load 28.06, mem 84% -> GO_RC=0, 0 fail, ok 4.701s
   JSONL subtests, both pass: a_record_stored_for_the_tip_tree_but_naming_another_tree_is_not_adopted;
   a_healthy_record_naming_its_key_tree_is_still_adopted
cli-f2-samename.jsonl            -run TestFactoryBranchOfWorktreeWithSameNameTag ./internal/cli/
                                                        slot free, load 29.15, mem 85% -> GO_RC=0, 0 fail, ok 2.834s
   JSONL subtests, both pass: resolution_returns_the_plain_branch_name;
   a_same-name_tag_on_the_card_branch_keeps_the_adoption_path_healthy
cli-acquire.jsonl                -run TestAcquire ./internal/cli/       slot free, load 29.38, mem 84% -> GO_RC=0, 0 fail, ok 3.244s
cli-integration-acquire.jsonl    -run TestIntegrationAcquire ./internal/cli/  slot free, load 27.77, mem 84% -> GO_RC=0, 0 fail, ok 5.664s
config-basefreshness.jsonl       -run BaseFreshness ./internal/config/  slot free, load 25.76, mem 84% -> GO_RC=0, 0 fail, ok 0.297s
template-mirror.jsonl            -run RuleTemplateMirrorDrift ./internal/template/  slot free, load 26.18, mem 84% -> GO_RC=0, 0 fail, ok 0.384s
   JSONL includes subtest TestRuleTemplateMirrorDrift/factory-dispatch-gates.md — pass.
```
Detached-HEAD probe (read-only scratch repo under /tmp, removed after): `git symbolic-ref --quiet HEAD` → no output, rc=1; `git rev-parse --abbrev-ref HEAD` → `HEAD`, rc=0.

### Baseline-attribution
Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834 at 1a16da10f6 in this run. Judging build for the gate: 8249bff8c3 (verified strict ancestor of the tree HEAD — the gate's own handling code predates the repairs; the codex review read the tree's diff and overlays). Go toolchain commands ran directly in the tree. The `moai` CLI used for `slot status` and `verify audit-plan` is the installed 8249bff8c3 build, not a build of this tree — those two calls are lease/plan surface only, not tree measurements. The detached-HEAD probe ran against a scratch repo, not this tree's code (it measures git's own behavior, which the comment's claim rests on).

### Gaps
- Entry-condition deviation on ONE scope: the pre-command reading for `-run Remeasure ./internal/cli/` showed 1-minute load 36.38, above the 32 ceiling (transient tail of the preceding compile; conditions (a) slot free and (c) mem 84% held). The command was not delayed and completed GO_RC=0. Disclosed plainly per the brief; the run's evidence stands, the deviation is process-side.
- The two exact-name -list pins (TestFactoryCompleteAdoptionRequiresRecordTreeIdentity, TestFactoryBranchOfWorktreeWithSameNameTag) shared one reading set taken before both (one Bash call, two serial lightweight lists). Both lists are GO_RC=0 pins; the readings before them were slot free, load 12.05, mem 85%.
- Detached-HEAD behavior of `factoryBranchOfWorktree` is verified by code reading + the git probe, not by a Go test at this HEAD — no test exists for that input shape on this helper; the callers' outcome equivalence is reasoning over the three call sites, not a test observation.
- Not run at this HEAD: full-package suites, coverage measurement, `go test ./...`, Windows build, golangci-lint, and the K>0 / `--wait` / concurrent-window behaviors (the brief's scoped list only; coverage thresholds were not audited this round — the go-test-heavy discipline and the brief's scope define what ran).
- The other `--abbrev-ref` readers (hook_install.go:87, session_worktree.go:768/870, session_worktree_automerge.go:311, integration.go:126) were not reproduced or tested — observation-only surfaces from the brief.
- The §E.2.8 record is uncommitted; the commit graph does not yet carry it (it is a .moai path, outside the audited diff; the leader's landing carries it).
- The convergence check (`--result-file`) is the orchestrator's step; my verdict is not final until it passes.

### Residual-risk
- D1 (merge-step same-name-tag, pre-existing) is live on the non-adoption complete path until its routed repair lands; the F2 fix does not cover that file.
- Three `factoryBranchOfWorktree` callers were behavior-checked by reading only for the detached-HEAD value change; a caller added later could depend on the former `"HEAD"` value.
- A hand-crafted `git symbolic-ref HEAD refs/<other>/x` (not creatable via checkout) would read as a non-branch name after the prefix strip — unreachable through normal git operation.
- The codex pass ran under build lag (8249bff8c3 < HEAD); its overlay work described the tree, but the gate's own handling code is older than the repairs it judged — the same lag iter12 carried.
- The implementer's own confirming runs quoted in §E.2.8 (Freshness|Remeasure ok 12.654s, config ok 3.655s, cli family ok 109.710s) were NOT re-observed by me; my independent scoped runs above are this audit's evidence, and the two sets used different -run composites (theirs used '|' alternation in a pre-audit implementer run; mine are the brief's plain patterns).

## Iteration history

| Round | Receipt | Verdict | Audited commit |
|---|---|---|---|
| 1-3 | rcpt-c9bb6784d6ebb463b124e1d4, rcpt-60312ee19c3179eea8af7e44, rcpt-56f80c4d7a80842575127f29 | FAIL | not re-read |
| 4 | rcpt-8fdd95a2ca747abf709b8914 | PASS | ae092df3adf82decccaa2e47a35a25ad9f963426 |
| 5 | rcpt-0a5d99b910ae00b2ddd53b41 | PASS | 30dff01635b992e760d875b31a61cd9f97394eed |
| 6 | rcpt-d93ba0455b5be353b0cc3509 | PASS | a63659551504a6f24efce731a484e541f91a48b1 |
| 7-10 | see sync-audit-iter7..10.md | FAIL | F1; F2; G1/G2; H1/H2 |
| 11 | rcpt-7784a17120108f86be26360c (+2) | FAIL | c3f96a52b134eba2f13307b8afdd688a194f2a38 |
| 12 | rcpt-6040618efd97e27498747772 | FAIL | afce6d2f789b4296cd5cb3602e2502a07a61e10e |
| 13 (this) | rcpt-5fd438dcfae9026bf8fb0732 | PASS-WITH-DEBT | 1a16da10f6edc7b78c22bdb8460908108ab220b9 |

## Recommendations
- Land on this verdict; the debts D1 (merge-step same-name-tag repair — already routed in §E.2.8, needs a card/routing owner) and D2 (one comment sentence + optionally the dead `== "HEAD"` disjunct at factory_card_pr.go:210) copy into the card's binding run conditions and are the next owner's, not a re-audit trigger for this card.
- Commit the §E.2.8 progress.md record with the card's landing so the repair-round record enters the commit graph.
- For the routing owner of D1: the fix shape is factory_card.go:3047's (symbolic-ref, refs/heads/ stripped), and integration.go:126 plus the four other --abbrev-ref readers deserve the same class sweep in that routing.

AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-INTEGRATION-STALE-BASE-001 receipts=rcpt-5fd438dcfae9026bf8fb0732
