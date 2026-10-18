auditor-model: claude-sonnet-5-5
verdict: PASS
audited_sha: ed6c8180b0b71f298df138478bf9b2849eccbc89

# Sync audit round 7 (delta + HEAD confirmation) - SPEC-CARD-ELIGIBLE-PULL-001, card t1839

Overall: PASS. F14 is repaired and the repair is faithful; the re-sweep is sound by an independent axis (one harmless omission, no test at risk); F15/F16 are accepted as load flakes. No BLOCKING finding.

audit_multi result digest (the verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes): overall_verdict=pass, gate_unmet=absent, plan_source=config, per_backend_verdicts=[backend codex, gate required, verdict pass], audit_receipt=rcpt-6290ea1b86c3e9dca14dfab9.

## Claim
1. The delta `954e25f96..HEAD` is exactly the five named files; the four test edits only add arming (`cepSelfDispatchYAML(t, root, true)`) and spec ids (`cepSetSpecID`) and change no assertion; the new comments are accurate.
2. Arming is faithful: the subject of each of the four tests now runs and is no longer vacuous, and it does not over-test the gate. Mutation evidence in progress.md reproduced by an independent mutant.
3. The re-sweep is complete by an independent axis; every lease-entry test form is in the union or provably unreachable from the gate.
4. F15/F16 are load flakes, not card defects (accepted).
5. Final head: card diff against the merge-base is exactly the card set; `factory_card.go` / `mcp_factory_card.go` carry main's t1680/t1748 changes plus the card gate.
6. Health at HEAD: all 160 scoped tests pass; build, vet, config and template pass.

## Evidence
- `git rev-parse HEAD` -> `ed6c8180b0b71f298df138478bf9b2849eccbc89` (unchanged at the end of the audit; `git status --short` empty before this report).
- `git diff 954e25f96..HEAD --stat`: progress.md (+92), sweep-union.txt (82), factory_bundle_test.go (+20), factory_m3_test.go (+9), factory_serial_slot_stale_test.go (+8); "5 files changed, 200 insertions(+), 11 deletions(-)".
- `git diff 954e25f96 HEAD -- internal/cli` read in full: every hunk adds a comment block, one `cepSelfDispatchYAML(t, root, true)` and one to two `cepSetSpecID(...)` calls; zero `-`/`+` lines touch an `if`/`t.Fatalf`/assertion. Placement: arming sits after `fcQueue`/`fcClassify`, before `sdRegisterLane`; `cepSetSpecID` after `fcQueue` (the ordering the progress record names).
- Faithfulness, subject runs. My own mutant (scratch `go test -overlay`; `factory_card.go` copy with line 1006 `elig.disqualify(c.CardID, it, inQueue, !selectionSkips(c), ...)` changed to pass `true` for depsMet; worktree untouched):
  - armed HEAD tests against the mutant: `--- FAIL: TestFactoryNextBundleSerialLane  factory_bundle_test.go:136: lane lane-1 next: factory next: predecessor card not merged: t2 has not reached merged-local ...` and `--- FAIL: TestFactoryBundleKeepsSerialSlot  factory_bundle_test.go:196: ... t3 has not reached merged-local ...`, `FAIL internal/cli 7.071s` (the mutant is killed; the bundle skip is what the tests now exercise).
  - pre-arming `factory_bundle_test.go` (`git show 954e25f96:...`) against the same mutant: `ok github.com/modu-ai/moai-adk/internal/cli 10.111s` (the old tests were vacuous; the repair is what makes them see the gate).
  - Over-testing check: the armed tests still assert only the original negatives/positives; arming adds no gate-specific assertion. Scenario C of BundleSerialLane leases through arm (b-priority), which the flag never touches (factory_card.go:964-986, ungated), so it correctly stays unarmed.
- `go test -list` of the four group patterns (`TestSD_`; `EligiblePull|SelfDispatch`; `TestFactoryCard|TestFactoryNext|TestFactoryBundle|TestFactoryNominate`; `TestFactoryQuota|TestFactoryLease|TestFactoryM3|TestFactoryT1533|TestFactoryTwoHub|TestReviewObservation|TestMCPFactory`) in ./internal/cli -> 128 unique names; `sweep-union.txt` -> 123 unique names, all resolve in `go test -list '.'` (5980 tests; 0 unresolved); overlap 91; fixed scope = 160 (123 + 128 - 91).
- Independent enumeration, different axis (not the commit's greps): (a) every line calling `factoryNextLeaseOnce(` / `factoryNextLeaseOnceGated(` / `handleFactoryNext` / `factoryNextNominate(` in any file under internal/, mapped to its enclosing `func Test` by awk -> 26 names, `comm -23` against the union: empty (all in the union); (b) every `"next"` literal in `internal/cli/*_test.go` mapped to its enclosing test -> 74 names, 36 outside the fixed scope, all `moai todo next` / `todo pick` tests (`runTodo(t, "next", ...)`, allowlist tables), not `factory next`; spot-read TestSD_AC015_LaneQueueAllowlistWalk (todo allowlist), TestOrdinaryCardSelectionSkipsBinding (`runTodo`), TestSemanticRelationsChangeNothing (todo); (c) helper axis: functions in test files containing `runFactory(t, "next"` / `handleFactoryNext` / `factoryNextLeaseOnce` / `qasRunNext` that are not `func Test`: fbLeasedCard, flRaceLanes, cepNoCard, nmRunCase, qasRunNext, qasLaneEnv - callers of each enumerated by `awk`; all callers are in the union except `TestFactoryNextNominateRefusalLeavesStateUnchanged` (caller of nmRunCase, which drives `--card` nomination only, never arms (b)/(b2)); it is matched by the `TestFactoryNext` group pattern, so it ran in this audit and passed; (d) production-caller axis: `grep` of non-test callers of `factoryNextLeaseOnceEligibility` -> CLI verb (factory_card.go:1914), MCP handler (mcp_factory_card.go:181), the gated wrapper (factory_card.go:382); the todo `--auto` path (todo_auto_lane.go:155,166) calls the gated wrapper with `noNewCards=true`, and factory_card.go:961-963 returns before arms (b-priority)/(b)/(b2) when `noNewCards` is true, so no `todo --auto` test can reach the gated arms; direct callers of `factoryNextClaim` / `factoryNextRecordAndClaim` in tests (factory_lease_bound_test, factory_classify_test, factory_nominate_test, factory_record_before_commit_test, factory_dispatch_ordering_test, factory_lease_current_dispatch_test) call below the gate.
- Arm gating at the code: factory_card.go:995-1014 (arm b: `if !elig.active() { continue }` then `elig.disqualify(...)`), :1022-1044 (arm b2, same gate); `elig` is created once per invocation at :403 from `factorySelfDispatchEnabled(root)`.
- Shards (the fixed 160-name scope split into 6, `split -l 27` of the sorted deduped list; each shard run once with `go test -json ./internal/cli -count=1 -run '^(...)$'`, JSONL kept under `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/94352e8a-6862-42bf-bba5-b9cfa0b1a4b4/scratchpad/t1839-iter7/shard-a{a..f}.jsonl`; package-level event read from the JSONL):

  | Shard | Names | Pass | Fail | Skip | Package elapsed | Tool exit | Load at launch |
  |---|---|---|---|---|---|---|---|
  | aa | 27 | 27 | 0 | 0 | 76.592s | 0 | 31.93 |
  | ab | 27 | 26 | 0 | 1 (`TestFactoryLeaseSerialCrossProcessHelper`, helper process only) | 157.321s | 0 | 28-30 |
  | ac | 27 | 27 | 0 | 0 | 148.064s | 0 | 29.73 |
  | ad | 27 | 27 | 0 | 0 | 107.974s | 0 | 31.71 |
  | ae | 27 | 27 | 0 | 0 | 141.905s | 0 | 29.96 |
  | af | 25 | 25 | 0 | 0 | 56.833s | 0 | 30.95 |
  | total | 160 | 159 | 0 | 1 | | | |

  Union check: `cat shard-a?.jsonl | jq` of top-level test names -> 160 lines, 160 unique, `comm -3` against the fixed scope list -> 0 differing; `"Action":"fail"` events across all shards -> 0 (subtests included). The two timing tests ran and passed in this run: `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` (shard aa) and `TestFactoryLeaseQueueLockStallBounded` (shard ab).
- `go build ./...` -> exit 0, output 0 bytes. `go vet ./internal/cli ./internal/config ./internal/template` -> exit 0, output 0 bytes.
- `go test -json ./internal/config -count=1` -> package `pass 4.196` s, 575 top-level pass events, 0 fail. `go test -json ./internal/template -run 'TestTemplateNoInternalContentLeak|ShippedKey|Workflow' -count=1` -> package `pass 23.47` s, 15 pass, 0 fail.
- F15/F16 judgment: the progress record's A/B (card vs `git archive main`, 8 runs for F15: 1 fail on each side, 19 ms vs 43 ms over a 500 ms headroom bound; 6/6 pass for F16) is consistent with a load flake; the failing F15 cell was the nominated lease, which the card diff does not touch (`factoryNextNominate` and the lock wait are unchanged). In this audit both tests passed at loads 28-32 on the card tree, which agrees. I did not run my own card-vs-main A/B pair (no failure to explain). Accepted, with the stated limit that this rests on the record's A/B plus one passing run of mine, not on a fresh A/B.
- Final-head diff (rule 6). `git diff main HEAD --stat` now lists 35 files because main has moved: `git log --oneline 03faddb705..main` shows 11 commits (t1818, t1873, tip `2bcce5d9dc`) that this tree does not contain; the extra files (audit_delta_persist_test.go, mcp_convergence.go, binding_ledger.json, two factory-dispatch rule files and their template mirrors) are main-side changes read in reverse. Recheck against the merge-base instead: `git merge-base main HEAD` -> `03faddb705a43a7d1757deb797d7e621d115a4d9`; `git diff --stat 03faddb705 HEAD` -> 28 files = SPEC artifacts, workflow.yaml (repo + template), CHANGELOG.md (+1), factory_card.go, mcp_factory_card.go (7 lines), the card tests, config defaults/defaults_test/types, shipped_key_inventory.yaml, plus `.moai/reports/t1839/sync-audit-iter6.md` (see F18). `git diff --name-only 03faddb705 main` -> 8 files; intersection with the card set: CHANGELOG.md only. `git diff --stat 03faddb705 main -- internal/cli/factory_card.go internal/cli/mcp_factory_card.go internal/config` -> empty (main did not touch those since the absorb), so the t1680/t1748 content in the card tree is still current main content; `git log -3 -- internal/cli/mcp_factory_card.go` shows the t1680 commits `a3094207c6`, `1e768f7b04` and the card's merge `7616a604b6`.

## Baseline-attribution
- Tree measured: HEAD `ed6c8180b0b71f298df138478bf9b2849eccbc89`, branch `WT-10-17-c`, worktree clean at start and at end of the test runs; this report is the only file written (it is gitignored by `.gitignore:235 .moai/reports/*`; the lane commits it).
- Judging build of the moai MCP tool: server build `ee981a932` (built 2026-10-15T21:56:00Z), reported by `audit_multi` as `build_lag`: an ancestor of the tree's HEAD, so the codex leg describes the diff reviewed, not code behaviour. All Go test/build/vet results come from the local Go toolchain over this tree (no moai CLI binary was used for any measurement).
- Mutants and the pre-arming test copy: scratch overlay files under the session scratchpad, deleted from nothing (never in the tree).
- Codex receipt: `rcpt-6290ea1b86c3e9dca14dfab9`, single call, `project_root` = this worktree, `target: baseBranch`, minted during this audit at this HEAD.

## Gaps
- Process exit codes: the guard refused no redirects to the scratchpad, but a compound `; echo $?` form was not used (guard refuses compounds); the "tool exit 0" cells are the Bash tool's success result (a non-zero exit is reported as an error), cross-checked by the JSONL package-level `pass` event. The raw `go` exit code was not echoed separately.
- Load: the 1-minute load was above 32 at three decision points and I waited (re-read) each time; launches were at 28-32. The admission rule was observed, but shards ran in the 28-32 band, so timing-sensitive tests were exercised at high load (both passed).
- F15/F16: no own card-vs-main A/B pair; relied on the record plus one passing run each.
- The whole-package `internal/cli` suite, the go-test-heavy slot, and CI on the integration branch were not run (per brief); the lane's landing re-measure owns them.
- The re-sweep completeness remains non-mechanical for a test reaching the lease entry through a helper of an unlisted shape; my four axes (callers, `"next"` literal, helpers, production callers) bound it but are greps over source, not a runtime probe. I did not re-run the commit's behaviour-neutral probe.
- Subtest-level and `-v` raw output of earlier rounds' batches 1-4 (progress record) were not re-observed; this audit's shards supersede them for HEAD.
- Main has moved (11 commits) since the absorb; this audit judged the tree as it stands, not a re-absorbed tree.

## Residual-risk
- Re-absorbing main before landing will conflict textually in CHANGELOG.md (both sides add lines) and could change test behaviour through main's t1818/t1873 changes (rules, `mcp_convergence.go`, `audit_delta_persist_test.go`); the landing re-measure must run at the post-absorb head, not at `ed6c8180b0`.
- Timing tests (F15/F16) remain load-sensitive on both trees; a red there at landing is a flake candidate to be A/B'd against main, not a card defect.
- The codex leg is an older server build's diff review; it found nothing, but in round 6 it also found nothing while AC010 was red, so it carries little weight against the test evidence.

## Findings
- F17 [MINOR] [optional] `.moai/specs/SPEC-CARD-ELIGIBLE-PULL-001/sweep-union.txt` - `TestFactoryNextNominateRefusalLeavesStateUnchanged` (a caller of the helper `nmRunCase`) is not among the 123 names; it drives only `--card` nominations, never arms (b)/(b2), is matched by the `TestFactoryNext` group pattern and passed in shard ac. No fix needed for the verdict; optionally list it on the next regeneration so the header's helper-axis sentence ("callers NominateBackendSkip / NominateQuotaHold") is complete (it names two of three-plus callers).
- F18 [MINOR] [optional] `.moai/reports/t1839/sync-audit-iter6.md` - this verdict file is tracked in the branch (present in `git ls-files .moai/reports/t1839` and in the merge-base diff) although `.gitignore:235` ignores `.moai/reports/*` and the audit-artifact convention keeps verdicts local; it will ride into the landing diff. Fix, if the lane agrees: untrack it (`git rm --cached`) in a follow-up commit instead of landing a verdict file; not a correctness matter.
- F19 [SHOULD-FIX for landing, not a verdict blocker] main moved 11 commits ahead of the audited absorb (tip `2bcce5d9dc`); only CHANGELOG.md overlaps. Fix: absorb main, resolve CHANGELOG.md, then run the landing re-measure at the new head; this verdict covers `ed6c8180b0` only.
- F14 (round 6, BLOCKING): resolved at this HEAD (AC010 passes in shard af; test armed; mutant evidence reproduced).
- F15 (round 6, SHOULD-FIX): accepted as a load flake (see Evidence). F16 (round 6, MINOR): accepted likewise.
- Carried optional items F4/F5/F7/F8: not re-examined.
- Binding run conditions: `grep -n -i "binding run conditions" progress.md` -> no match (exit 1, no output); no binding condition to dispose.

AUDIT-VERDICT: PASS spec=SPEC-CARD-ELIGIBLE-PULL-001 receipts=rcpt-6290ea1b86c3e9dca14dfab9
