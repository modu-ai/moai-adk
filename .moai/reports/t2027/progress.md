# t2027 progress record (lane-4, run tnwnit)

- 2026-11-04 intake: leased via `moai factory next --card t2027 --run tnwnit`; tree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2027, branch WT-internal-agenticfactory-; toplevel confirmed.
- base: merged release/main-batch f9a4efd964 into the card tree -> HEAD 9fc89e246e (merge commit, card t2027); tree clean after the merge.
- class: B (three static review findings, defect unknown; no SPEC). Stage: run via manager-develop (effort high), then card-review, STOP at merge-ready; no push/PR/paid audit/complete.
- lease: `moai factory renew t2027 --run tnwnit` only with `--run`. Observed outputs: renewed until 00:14:09Z, 00:19:07Z, 00:28:39Z (2026-11-04), taken while the writer ran and after it finished.
- no guard refusal occurred in this card (every command ran as issued; the writer reports none either).

## Result per finding (reproduction input, result, disposition)
| # | Finding | Reproduction input | Result | Disposition |
|---|---|---|---|---|
| 1 | P1 mergeq.go:284 | a gate that rewrites tracked `base.txt` in the check worktree, then `Land` | `outcome=landed ... trunkAfter == pre-gate result`; `status after gate: "M base.txt"`; `base.txt on the trunk: "base"`; `check worktree still exists: false`; test fails with `REPRODUCED: landed ... although the gate ran on a tree that differs from it` | reproduced; NOT fixed: each fix changes what the merge queue means (see below) |
| 2 | P2 location_memo.go:149 | resolve through a symlink, re-point it to another repository, resolve again in one process | plain case: not reproduced (config mtimes differ, fingerprint already changes); with repoB `.git/config` given repoA's size and mtime: reproduced (memo answered repoA, cold lookup repoB) | fixed: commit 8aed1a8293 |
| 3 | P2 af.go:219 | `moai af claim --db` with `ledger.db`, `af/ledger.db`, `./ledger.db`, then look for the heartbeat binding | every form fails in `ledger.Open` with `invalid uri authority` (`ledger.db?_pragma=...`, `af`, `.`); no binding file is written | not reproduced; latent only (if `ledger.Open` ever accepted relative paths, af.go:219 would store them as given); no change |

Details and the writer's own record: `.moai/reports/t2027/repro.md`; preserved reproduction tests (as .txt, not compiled): `repro-finding1-mergeq.go.txt`, `repro-finding3-af.go.txt`.

## Lane observations (this run, plain commands, tree HEAD 8aed1a8293)
- Diff read: `git show --stat 8aed1a8293` -> internal/homestate/location_memo.go (+10/-1) and internal/homestate/location_memo_symlink_test.go (+76). `locationFingerprint` now begins with `filepath.EvalSymlinks(dir)` and writes `=>resolved` into the fingerprint; when the call fails nothing is added.
- `uptime` -> load averages 71.00 60.94 61.05
- `go test ./internal/homestate/ -count=1 -run 'TestLocationMemo|TestCanonicalProjectRoot|TestOutsideRepository|TestProjectDir'` -> `ok  github.com/modu-ai/moai-adk/internal/homestate  2.009s`
- `go test ./internal/agenticfactory/mergeq/ -count=1 -v -run TestLand_GateThatModifiesATrackedFile_Reproduction` -> FAIL with the `REPRODUCED` line and the values above (the test was untracked; kept as .txt copy, original removed)
- `go test ./internal/cli -count=1 -v -run '^TestAFClaim_RelativeDBReachesNoBinding$'` -> PASS; the three `--db` forms log `invalid uri authority` errors (name filter only; `internal/cli` full suite NOT run)
- writer's own runs (not re-observed by the lane): vet clean on homestate and mergeq, gofmt clean; RED before the fix for `TestLocationMemo_SymlinkRepointedBetweenRepositoriesWithEqualConfigStamps` ("CanonicalProjectRoot = .../001/repo, want .../002/repo (a stale remembered answer)"), GREEN after; the writer's load readings 56.81 57.30 61.58 to 65.94 58.32 60.44.

## Decision record (auto-semantics section 10) — finding 1 left unfixed
- Judgment: lane's own judgment through the ladder; no contract/Jev/advisor consulted (ordering gap noted: not run).
- Options the writer identified: (a) refuse the landing when the tree is dirty after the gate (needs a new Git interface method; gates such as `go mod tidy` or `go test` may legitimately touch `go.sum`, so some gates would start failing), (b) commit the gate's changes into the merge result, (c) accept as designed.
- Chosen: reproduce and report, do not guess a fix. Reason: the card asks for a minimal fix only when the fix is unambiguous; each option changes queue semantics (landing behavior for gates), which is a design decision for the card owner. This is the lane's call, not an operator-held or irreversible action; the leader may re-dispatch with the chosen option.

## Gaps
- Finding 2 needs two repositories whose `.git/config` have the same size and mtime (an archive restore, a coarse-timestamp filesystem); its practical likelihood is low. The fix covers symlinked paths only.
- `go vet` was not run on `internal/cli` (the card gained no committed change there); the `internal/cli` test used a name filter only. Full homestate and mergeq suites were not run; only name-filtered runs.
- Findings 1 and 3 evidence rests on one run each at load ~55-70; the failing/passing outcomes are deterministic logic, not timing, but were not repeated.
- The writer's Grep/Glob tools were unavailable (used `grep`/`find` via Bash); no command was refused.
- Load-bearing: the `uptime` readings above are machine load at that time, not a controlled baseline.

## Task state
- run DONE: one fix commit 8aed1a8293 (finding 2); findings 1 and 3 dispositioned without product change.
- card-review DONE: .moai/reports/t2027/card-review.md. Tool verdict `fail` (advisory, static): the same 3 batch-code findings as t2022 (branch_guard.go:463 P1, factory_release_route.go:252 and :280 P2), outside this card's diff; not adopted, not recorded as a pass.
- state: merge-ready stop at the evidence commit on top of 8aed1a8293; no push, no PR, no paid audit; `moai factory complete` NOT run. Open for the leader: finding 1 design decision.
