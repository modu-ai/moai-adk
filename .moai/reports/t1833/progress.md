# t1833 progress (run phase, Class B, no SPEC)

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833`, branch `WT-10-17-class-7`, base `fb749d372` (local main). Commits: `c78e02e68` (bisect verb), `44bff97a6` (release CI gate), `774c166ab` (runbook). Nothing pushed. No remote contacted, no tag touched, `scripts/release.sh` never run for real (tests use `--dry-run` against scratch repositories with a stub `gh`).

## 1. Cause-establishing evidence (Class B record)

Card (a): is a landing one `--no-ff` merge carrying the card id?
Code: `internal/factory/integration_merge_step.go:399` `mergeMsg := fmt.Sprintf("Merge %s into %s (card %s, integration merge)", cardBranch, in.IntegrationBranch, in.CardID)`; line 609 `git("merge", "--no-ff", "--no-autostash", "--no-overwrite-ignore", "-q", "-m", mergeMsg, pinned)`.
Observed (command `git log --first-parent --format='%P|%s' main -n N`, classified by parent count and `\(card t[0-9]+` in the subject, tree fb749d372):

| window | two parents + card id | two parents, no card id | one parent |
|---|---|---|---|
| 40 | 40 | 0 | 0 |
| 100 (back to 2026-10-07) | 64 | 4 | 32 |

Examples, two parents + card id: `fb749d372 d2d9ac08f c3d2a5d4c | Merge WT-10-15-release-3 into main (card t1760, integration merge)`; `7faf75fdf e806cdd6c e6ca1d560 | Merge WT-audit-gate-revert into main (card t1845, integration merge)`.
Two parents, no card id: `ad002d3a0 ... | Merge WT-codex-moai-home-add-dir into main (codex --add-dir moai home, integration merge)`; `2aab5f797 ... | Merge pull request #1816 from modu-ai/WT-always-loaded-char-budget`.
One parent: `842160157 763193861 | chore(config): auditor and planner models per operator role table (opus high)` (3 one-parent `chore` commits: 2 `chore(config)` and 1 `chore: commit the primary checkout's pending changes onto main`); `db0c514d3 d3f51d0d9 | fix(factorymsg): set synchronous(NORMAL) on broker open paths (t1592) (#1815)` (PR-era single-parent commits, 28 of the 32 carry a `tNNNN` token).
First-parent positions 1-42 (2026-10-14 .. 2026-10-17) are all card merges. Older than the 100 window: 0 further card merges (`awk` over the rest of first-parent history; 138 of 3738 first-parent commits are two-parent).

Batch shape (origin/main, read-only local ref): `M=6dde1c1c1` "Merge pull request #1826 from modu-ai/WT-batch-landing-1017": `git merge-base --is-ancestor M^1 M^2` true; `M^{tree}` equals `M^2^{tree}`; `git log --first-parent M^1..M^2` yields 63 units, including two non-card units (`merge: absorb origin/main into the batch landing branch`, `fix: repair three main-side CI failures blocking the batch landing PR`). So origin/main first-parent = one step per batch; card granularity lives on the `M^2` chain.

Result: the claim "each landing carries the card id" holds for the code path and for the observed log since 2026-10-14; it does not hold for direct commits and batch-assembly commits, which a bisect reports as `card: none`.

Card (b), pre-change defect (RED, scratch repo, stub gh): `TestReleaseScriptCIGate` all 7 rows FAIL on the unchanged script; `TestReleaseScriptCIGateWithoutGh` on the unchanged script: exit 0, output `[WARN] gh CLI not available — skipping CI status check` (fail-open), override run `[WARN] CI check skipped (--skip-ci-check)` with no summary record. Old read: `gh api /repos/modu-ai/moai-adk/commits/$LOCAL_SHA/status` (combined status), `pending` accepted via y/N prompt, unknown only warned.

## 2. Design decision

- Verb vs recipe: a recipe alone is `git bisect start --first-parent <bad> <good>` + `git bisect run sh -c '<cmd>'` + `git bisect reset`. A verb is warranted only for what the recipe cannot enforce: bisect moves HEAD, so the primary checkout must be refused (shared by several sessions); a dirty tree, an in-progress bisect and an unresolvable ref fail late and confusingly; a forgotten `reset` leaves a tree mid-bisect; and the useful output is the first-parent merge subject with the card id, which `git bisect run` prints only as a SHA. Verb kept minimal (one file, ~130 lines, reuses `runScrubbedGit`, `identifyPrimaryCheckout`, `factorySameTree`, `shellJoinArgs`, `factory.ShortSHA`); `--dry-run` prints the three git commands. Not built: endpoint verification, a batch-aware `--batch <M>` shorthand, an inner-card mode (documented as 3 git lines in the runbook section 2.4).
- Release gate: in `scripts/release.sh` validation 9 (single tag-placement point). Check runs, not combined status. Strict (all check runs) rather than required-only: the tag is irreversible. Override stays `--skip-ci-check`, recorded in output and summary. GoReleaser: `.goreleaser.yml` consumes an already-pushed tag and has no point at which a tag is placed or refused; `release.yml` runs after the tag exists. Not changed; named as a follow-up candidate in the runbook.
- Template mirror: none. The runbook is repo-local (`scripts/release.sh`, `AGENTS.local.md`); `internal/template/templates/.moai/docs/` mirrors 6 generic docs only, and no test lists `.moai/docs` membership (moai-docs sweep floor 33, repo has 41 files).

## 3. Tests (all with TMPDIR=GOTMPDIR=/private/tmp/lane3-bisect, no -race, one package per run)

| command | exit | result |
|---|---|---|
| `go test ./internal/cli -run '^TestBisect' -count=1 -v` | 0 | 8 tests PASS, `ok ... 6.876s` (prefix selector, 8 tests, 7 s; broader than a single name) |
| `go test ./internal/template -run '^TestReleaseScriptCIGate$' -count=1` | 0 | ok (7 rows) |
| `go test ./internal/template -run '^TestReleaseScriptCIGateWithoutGh$' -count=1` | 0 | ok (2 rows) |
| `go test ./internal/template -run '^TestReleaseScriptCharacterization$' -count=1` | 0 | ok, also ok before the change |
| `go test ./internal/template -run '^TestReleaseScriptMatrixGate$' -count=1` | 0 | ok |
| `go test ./internal/template -run '^TestReleaseScriptHeadAdmission$' -count=1` | 0 | ok |
| `go test ./internal/template -run '^TestReleaseScriptRcChangelogSkip$' -count=1` | 0 | ok |
| `bash -n scripts/release.sh` | 0 | syntax ok |
| `go vet ./internal/cli`, `go vet ./internal/template` | 0, 0 | clean |
| `gofmt -l` on the 5 Go files touched | 0 | no output |

Mutants (each applied by `sed`, run, then restored; `cmp` against the saved original confirmed restoration):
- bisect verb: no primary refusal -> `TestBisectRefusesThePrimaryCheckout` FAIL; no reset -> `TestBisectNamesTheCulpritCard` and `TestBisectReportsAnAbortedSearchAndStillResets` FAIL; no `--first-parent` -> `TestBisectNamesTheCulpritCard` and `TestBisectDryRunPrintsTheGitCommandsAndMovesNothing` FAIL; no dirty check -> `TestBisectRefusesADirtyTree` FAIL (first attempt did not compile, reapplied with `&& false`).
- release.sh: pending branch disabled -> `.../pending_is_refused` FAIL; gh-missing made non-fatal -> `.../no_override` FAIL; failed branch disabled -> `.../failed_is_refused` and `.../failed_beats_pending` FAIL.

## 4. Files changed

`internal/cli/integration_bisect.go` (new), `internal/cli/integration_bisect_test.go` (new), `internal/cli/integration.go` (register verb), `scripts/release.sh`, `internal/template/release_fixture_test.go` (stub answers check-runs, default green), `internal/template/release_ci_gate_test.go` (new), `.moai/docs/batch-failure-bisect-runbook.md` (new), this file.

## 5. Gaps

- The verb was not run against the real repository history and `moai` was not rebuilt (`make build` barred); its behavior is evidenced on synthetic repositories with real git only.
- The jq expression and the `check-runs` response shape were exercised against the stub only; no live GitHub call (remote contact barred). The claim that the combined-status endpoint omits Actions check runs is GitHub's documented definition, not observed here.
- `golangci-lint` was not run (package-wide lint judged heavy); `go vet` and `gofmt` were.
- The sweep guard (`github_flow_sweep_guard_test.go`) is env-armed and was not run against the new doc.
- Whole-package `go test ./internal/cli` and `./internal/template` not run (barred); only the named tests above.
- `.github/workflows/release.yml` (post-tag provenance gate) deliberately untouched.
- Local main and origin/main currently diverge (`git rev-list --count --left-right origin/main...main` gave `10 131`); noted only, not acted on.

## 6. Residual risk

A strict all-check-runs gate can block a release over an auxiliary failing check; the override exists and is recorded. A hand-made tag bypasses `release.sh` entirely.

## WAIT STATE: sync-audit iteration 1 FAIL; the fix spawn is denied by the audit receipt guard (lane-3, 2026-10-17)
- Verdict: .moai/reports/t1833/sync-audit-iter1.md, FAIL 0.62 at audited_sha 8f1873c58d32e8f09fd9655b0b835399b2ff8a48 (committed 2ca38cdcf); codex required gate `fail`, receipts rcpt-12da61c41fa93963600ea2a4 and rcpt-5abefcca08d731b67723a1c5. Blocking: F1-F5 (bisect culprit read from mixed output; GIT_* env bypasses the primary-checkout refusal; failed `git bisect reset` returns success; failed `bisect start` leaves HEAD moved; `git tag -a` names no commit). Optional but P1 by codex: F6 (gate passes on any single green check run). Low: F7-F9.
- What was attempted: `Agent(subagent_type: manager-develop)` for fix round 1 (Reproduction-First, one test per defect). Denied by the PreToolUse hook, verbatim: "AUDIT_RECEIPT_VIOLATION: manager-develop cannot be spawned while an audit PASS stands unproven in /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833. Outstanding: sync-auditor / unknown-spec / verdict line missing. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt."
- Observed record (read-only, `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json`, gitignored machine-local state): agent_type sync-auditor, spec_id unknown-spec, agent_id a0c5e4ec206274aa7 (the t1833 sync auditor), cause "verdict line missing", rejected_at 2026-10-17T14:34:51Z, reentry_warned true. It was written when the auditor stopped once without a parseable verdict line and was blocked, then continued; the auditor's final message did end with `AUDIT-VERDICT: FAIL spec=t1833 receipts=...`.
- Read from internal/hook/audit_receipt_guard.go (not tested by running it): a FAIL verdict line returns without clearing rejections ("A FAIL needs no receipt"); only a proven PASS by the same role in this tree clears receipt-kind rejections (ClearRejectionsForRoleInTreeKind); phase-entry spawns (manager-develop, manager-docs, manager-git) are denied while any such record exists; no CLI verb or rule documents a manual clear (grep over .claude/rules, .moai/docs, AGENTS.local.md found no mention of the sentinel). Consequence: the only sanctioned way to clear it is a PASS, a PASS needs F1-F5 fixed, and the fixes are routed to the specialist the guard denies.
- Not done on purpose: the record was not deleted or edited (it is the guard's record, not the lane's); no substitute agent type was spawned to make the fixes (that routes around the deny); the lane did not edit the Go code or scripts itself (delegation duty). No push, no tag.
- Who is waited on: the leader (operator-delegated), who decides how to proceed; the lane asked by message.
- Resume condition: a decision that lets the fix round run. Then: fix round 1 as scoped above -> commit -> `moai factory stage t1833 sync <sha>` is not needed (stage is already sync) -> sync-audit iteration 2 on the new HEAD with the codex gate -> T11 `sync-audit` with `<audited_sha>:<tracked file>` -> T13 `merge-ready` -> slot, re-measure, window, `factory complete`.
- State now: factory record t1833 = sync v6 (run tn1p4e), lease kept alive by the session-local renew cron; card branch WT-10-17-class-7 tip is this record's commit on top of 2ca38cdcf.

## Fix round 1 (after sync-audit iteration 1 FAIL, lane-3, 2026-10-17/18)

Tree `.moai/worktrees/t1833`, branch `WT-10-17-class-7`, start HEAD `1a1f42b23`. Commits: `e86799369` (F1-F4 + F7 error text: `integration_bisect.go` and its tests), `315fc6dc9` (F5, F6: `scripts/release.sh`, `release_ci_gate_test.go`), then the docs/evidence commit that carries this section. Nothing pushed, no tag created or moved, `scripts/release.sh` never run against the real repository (every release run was in a scratch clone with a stub `gh` and a `git` shim that blocks the push). All go commands: `env -u MOAI_FACTORY_WORKER -u CLAUDE_PROJECT_DIR TMPDIR=/private/tmp/lane3-fix GOTMPDIR=/private/tmp/lane3-fix go ...`, no `-race`, one package per invocation.

### Red then green, per defect (tests written first, observed red on the unchanged code)

| Defect | Test | RED (before the fix) | GREEN (after the fix) |
|---|---|---|---|
| F1 culprit from test output | `TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand` | exit 1: `integration_bisect_test.go:193: a decoy line printed by the judged command must not change the culprit; want card t1003` (verb printed `card: t1001`) | exit 0 |
| F2 GIT_DIR split | `TestBisectActsOnTheCheckedTreeWhenGitDirIsSet` | exit 1: `:223: bisect with GIT_DIR set: integration bisect: the search did not converge` | exit 0 |
| F3 reset failure swallowed | `TestBisectFailsWhenTheResetFails` | exit 1: `:247: a failed git bisect reset must be an error that names the reset, got <nil>` (output carried `WARNING git bisect reset failed (exit status 1)`) | exit 0 |
| F4 failed start leaves HEAD moved | `TestBisectRestoresHeadWhenStartFailsPartway` | exit 1: `:269: a failed start must leave HEAD at c0171b68..., is 6ea47b46...` | exit 0 |
| F5 tag names no commit | `TestReleaseScriptTagsTheVerifiedCommit` | exit 1: `tag v9.9.9-rc.1 names 000c51b7..., but the CI gate verified 1bcab3f7... (HEAD had moved to 000c51b7...)` | exit 0 |
| F6 unrelated green check run | `TestReleaseScriptCIGate/unrelated_green_check_run_alone_is_refused`, `/expected_workflow_missing_one_check_is_refused` | exit 1: `exit code = 0, want 1` (both rows), `output lacks "CI gate: expected check run(s) not reported"` | exit 0 |

Commands (RED, run on the unchanged code with the new tests):
- `go test ./internal/cli -run '^(TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand|TestBisectActsOnTheCheckedTreeWhenGitDirIsSet|TestBisectFailsWhenTheResetFails|TestBisectRestoresHeadWhenStartFailsPartway)$' -count=1` -> exit 1, `--- FAIL` for all four, `FAIL github.com/modu-ai/moai-adk/internal/cli 9.657s`.
- `go test ./internal/template -run '^(TestReleaseScriptCIGate|TestReleaseScriptTagsTheVerifiedCommit)$' -count=1 -v` -> exit 1; `--- FAIL: TestReleaseScriptCIGate` with the two F6 rows FAIL and the other 5 PASS; `--- FAIL: TestReleaseScriptTagsTheVerifiedCommit` (the first RED attempt of the F5 test was invalid: it died at the harness-provenance check, then at a PATH that reached the real `gh`; both test-harness errors were fixed before the RED above was recorded, see Gaps).

Commands (GREEN, after the fixes):
- `go test ./internal/cli -run '^TestBisect' -count=1` -> exit 0, `ok github.com/modu-ai/moai-adk/internal/cli 15.989s` (12 tests: the 8 old plus the 4 new).
- `go test ./internal/template -run '^(TestReleaseScriptCIGate|TestReleaseScriptCIGateWithoutGh|TestReleaseScriptTagsTheVerifiedCommit)$' -count=1 -v` -> exit 0, `--- PASS` for all 11 `TestReleaseScriptCIGate` rows, 2 `WithoutGh` rows and `TestReleaseScriptTagsTheVerifiedCommit`; `ok ... 9.789s`.
- `go test ./internal/template -run '^(TestReleaseScriptCharacterization|TestReleaseScriptMatrixGate|TestReleaseScriptHeadAdmission|TestReleaseScriptRcChangelogSkip|TestGitHubFlowSweepGuard)$' -count=1` -> exit 0, `ok ... 22.656s`.
- `bash -n scripts/release.sh` -> exit 0. `gofmt -l` on the four Go files -> exit 0, no output. `go vet ./internal/cli` -> exit 0. `go vet ./internal/template` -> exit 0 (separate invocations).

### What changed, by defect

- F1: the culprit is the `# first bad commit: [<sha>]` line in `git bisect log`, read after `git bisect run` returns 0 and before the reset removes it. The log is git's state; the judged command cannot write to it. `git bisect run` output now goes straight to the caller (no longer captured for matching).
- F2: `bisectGit` drops `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, `GIT_CEILING_DIRECTORIES` (the set `internal/auditreceipt` scrubs for the pre-checks) and keeps the rest of the environment, which the judged command inherits through `git bisect run`.
- F3, F4: the starting HEAD and branch are recorded before `git bisect start`. `bisectRestore` runs on every path (failed start, failed run, converged, not converged): `git bisect reset` when `BISECT_LOG` exists, a `git checkout` of the starting branch/commit when HEAD differs, then verifies HEAD sha, branch and absence of `BISECT_LOG`. Any reset error or failed verification returns an error that names it and the manual recovery; the command exits 1. When a culprit was found the report lines are still printed before the error. The claim "always resets" in the CHANGELOG and runbook was reworded to what is now true (reset on every path, verified, failure exits 1) rather than left as "always".
- F5: `git tag -a "$VERSION" -F "$TMP_NOTES" "$LOCAL_SHA"`; the HEAD re-check is kept as an early refusal. The test moves HEAD with a `git` shim at the moment `git tag -a` runs and blocks `git push`; the tag must name the verified commit.
- F6: `CI_REQUIRED_CHECKS=("Lint" "Test (ubuntu-latest)" "Build (linux/amd64)")` at the top of validation 9; the gate reads one `name<TAB>status<TAB>conclusion` line per check run, applies the old rules (none / failed / pending refuse first), then refuses unless each expected name is present. `--skip-ci-check` remains the only override and is still recorded.
- F7: runbook section 2.2 exit-code sentence corrected (126, 127 and 128+ abort the search; only 1-124 mark bad; observed `error: bogus exit code 127 for good revision`), and the verb's error text no longer says "exited >127". F8: `동사` -> `하위 명령`, `끝점` -> `양 끝 ref` (a rewording of 10 occurrences, no change of meaning). F9: see below.

### F6 decision: how the workflow name was determined

- `.github/workflows/ci.yml` line 1 is `name: CI`; it is the workflow that gates main (triggers: push to `main` and `ci/**`, pull_request to `main`, workflow_dispatch). A check run is named after its JOB, not the workflow. Job names read from the file: `Detect (paths-filter)` (line 62), `Test (${{ matrix.os }})` (103, matrix `[ubuntu-latest]`), its docs-only complement `test-skip-marker` with the identical name (667), `Guard Bundle`, `Race Test (...)` x6, `Integration Tests (...)`, `Lint` (773), `Build (${{ matrix.goos }}/${{ matrix.goarch }})` (820, matrix linux/darwin/windows x amd64/arm64), `Constitution Check`, `Test (browser fire guard)`.
- `.github/required-checks.yml` (the SSoT for branch protection, loaded by `internal/config/required_checks.go`) lists for `main`: Lint, `Test (ubuntu-latest)`, five `Build (...)` names, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`, `Guard Bundle`; the `ci/**` key lists "the contexts a ci/** push ACTUALLY publishes": Lint, `Test (ubuntu-latest)`, the five Build names. A tag is cut from a commit already on main, so the check runs on that commit come from the push event; `Analyze` and `Release PR Multi-OS Gate` are pull-request names.
- Chosen: the three names that `ci.yml` itself publishes on a push and that are also in main's required set: `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`. Not required by name (but still blocking if red, because every check run must be green): `Guard Bundle`, the other four Build legs, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`. Reason: a smaller name list has fewer ways to be wrong against GitHub's real naming; if any of the three names is wrong the gate refuses (fail closed) instead of passing falsely.
- NOT verified against live GitHub: that GitHub reports these three exact names on a push-to-main commit, that the `Test (ubuntu-latest)` skip-marker path publishes the same name on a docs-only commit (stated in `ci.yml` comments, not observed), that `gh api ... --jq '.check_runs[] | ...'` behaves under real `--paginate` as it does on the stub (jq runs once per page; the output lines are concatenated), and that `filter=latest` (the endpoint default) hides re-run duplicates. The first real `release.sh <tag> --dry-run` is the measurement; a wrong name makes it refuse with `expected check run(s) not reported`, naming what is missing.

### F9 recount

`git log --first-parent --format='%P|%s' main -n 100`, main = `fb749d372` (local main did not advance), classified by parent count: 32 one-parent, 68 two-parent, 64 of the two-parent carry `(card tNNNN`. One-parent subjects starting with `chore`: 3 (`chore(config): auditor and planner models ...`, `chore(config): consume auditor model overrides ...`, `chore: commit the primary checkout's pending changes onto main`); of those, `chore(config)`: 2. Section 1 of this file now says "3 one-parent `chore` commits: 2 `chore(config)` and 1 `chore: commit ...`"; the runbook sentence that said "`chore(config)` ... 3" was corrected the same way.

### Mutation checks (apply the revert, run the matching test, restore with `git checkout -- <file>`, compare sha256)

| Fix reverted | Test | Result | File after restore |
|---|---|---|---|
| F1: culprit by regex `is the first bad commit` over the mixed output | `TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand` | exit 1, FAIL (`want card t1003`) | sha256 `f0363c4d...7660ff` = committed |
| F2: `bisectGit` keeps GIT_* | `TestBisectActsOnTheCheckedTreeWhenGitDirIsSet` | exit 1, FAIL (`did not converge`) | same hash |
| F3: restore error swallowed (`_ = restoreErr; return nil`) | `TestBisectFailsWhenTheResetFails` | exit 1, FAIL (`got <nil>`) | same hash |
| F4: no restore after a failed start | `TestBisectRestoresHeadWhenStartFailsPartway` | exit 1, FAIL (`is 1ae8ac73...`) | same hash |
| F5: `git tag -a` without `"$LOCAL_SHA"` | `TestReleaseScriptTagsTheVerifiedCommit` | exit 1, FAIL (tag on the moved HEAD) | sha256 `da57cda6...199974` = committed |
| F6: expected-name check disabled (`if false`) | `TestReleaseScriptCIGate` | exit 1, the two F6 rows FAIL | same hash |

`git status --short` was empty after each restore.

### Gaps

- Live GitHub was never read on purpose. One unintended read-only call happened: my first run of `TestReleaseScriptTagsTheVerifiedCommit` replaced the PATH that held the `gh` stub, so the script ran the real `gh api repos/modu-ai/moai-adk/commits/<scratch sha>/check-runs` once and got `gh: No commit found for SHA ... (HTTP 422)` (a GET for a sha that exists only in a scratch repo; nothing was written, no tag or push). The test now builds the PATH from the stub's PATH and cannot reach the real `gh`; the invalid run is not counted as evidence.
- The real `gh --paginate` page boundaries, the check-run names GitHub reports, and the skip-marker name on a docs-only commit were not observed (see F6 decision).
- Whole-package test runs, `golangci-lint`, and `make build` were not run (leader's load rule). The CHANGELOG sentences changed because they would otherwise be false; the markdown edits were not linted.
- `git bisect log`'s `# first bad commit:` line was observed on git 2.54.0 (Apple Git-157) only. On a git that does not write it the verb reports "did not converge" (safe failure), not a wrong card.
- The F3 test forces the reset failure with a post-checkout hook; a failure of another kind (a locked index) takes the same code path but was not separately exercised.
- Residual risk: a judged command that itself runs `git bisect` commands in the tree could alter the log; not guarded. The strict all-check-runs gate can still block a tag over an auxiliary failing check; a hand-made tag skips `release.sh` entirely.

## WAIT STATE 2: sync-audit iteration 2 FAIL; the fix spawn is denied again by the audit receipt guard (lane-3, 2026-10-18)
- Verdict: .moai/reports/t1833/sync-audit-iter2.md, FAIL 0.76 at audited_sha 505749a927a31338e3a5cb625e91af6825b87957 (committed d858e84d2); receipts rcpt-e45273c3e30caaae5dc3d6d8 and rcpt-57a278085d6222449dcf34c6; codex required gate `fail`. Blocking N1 (SIGINT/SIGTERM leaves the card worktree mid-bisect, so "resets on every path" is false); codex P1 N2 (check names matched without the app slug) and N3 (a bisect checkout can overwrite an ignored local file a visited commit tracks); codex P2 (--dry-run prints an unquoted command; the gh stub now needs jq). Iteration 1 was FAIL 0.62 (F1-F6 now confirmed fixed by this audit).
- Attempt: `Agent(subagent_type: manager-develop)` for fix round 2 (scope: N1, N2, N3, dry-run quoting, jq skip; N4/N5/N6b recorded as runbook limits). Denied by the PreToolUse hook, verbatim: "AUDIT_RECEIPT_VIOLATION: manager-develop cannot be spawned while an audit PASS stands unproven in /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833. Outstanding: sync-auditor / unknown-spec / verdict line missing. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt."
- Observed record (read-only), `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json`: agent_type sync-auditor, spec_id unknown-spec, agent_id adb7b3c57ae283384 (the iteration-2 auditor), cause "verdict line missing", rejected_at 2026-10-17T15:19:52Z, reentry_warned true. The iteration-1 record (agent a0c5e4ec206274aa7) was removed on the operator's decision relayed by the leader (board d-20261017T144753Z-0a97, backup .moai/reports/t1833/backup-rejection-20261017T144746Z/); this is a NEW record from a second auditor instance. The auditor's final message did end with `AUDIT-VERDICT: FAIL spec=t1833 receipts=...` (the spawn prompt required the verdict line as the last line of every turn). Also observed: a stray `.claude/agent-memory/sync-auditor` directory (empty) inside the rejections directory, created 2026-10-18T00:03Z.
- Pattern, from reading internal/hook/audit_receipt_guard.go (not reproduced by a test): a record is written when the auditor stops once without a parseable verdict line and is blocked; a later FAIL verdict line returns without clearing rejections; only a proven PASS by the same role in this tree clears them. Both t1833 sync audits therefore left a record that the next FAIL could not clear, and each blocks the fix round. Unverified: why both auditors stopped once without the line.
- Not done on purpose: the record was not deleted or edited; no substitute agent type was spawned; the lane did not edit the Go code or scripts. No push, no tag. The earlier operator approval covered one named record and is not treated as covering this one.
- Who is waited on: the leader (operator-delegated), asked by message.
- Resume condition: a decision that lets the fix round run. Then: fix round 2 as scoped above -> sync-audit iteration 3 on the new HEAD with the codex gate -> T11 `sync-audit` with `<audited_sha>:<tracked file>` -> T13 `merge-ready` -> slot, re-measure, window, `factory complete`.
- State now: factory record t1833 = sync v6 (run tn1p4e), lease kept alive by the session-local renew cron; card branch WT-10-17-class-7.
