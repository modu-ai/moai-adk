auditor-model: claude-sonnet-5-5

# Card t1856 sync-phase audit (Class B, no SPEC)

verdict: FAIL
audited_sha: 772c1c0507b4846693bdb6fa7bef111e6fe610a3

Result: FAIL. The resolver change itself is correct and minimal, but the two new regression tests hard-code `/private/tmp`, which does not exist on the only CI platform (ubuntu-latest), so they fail there. The required codex backend independently returned fail on the same defect. Score 0.55.

## Scores
- Functionality 55/100 (must-pass, FAIL): the fix is right, but the new tests fail on the CI platform.
- Security 90/100 (PASS): test-only change plus one `.invalid` fixture address in a deny list; no new input surface.
- Craft 55/100: order of main vs origin/main is not pinned; /private/tmp is macOS-only.
- Consistency 75/100: other tests use `t.TempDir()` or `/tmp`; this one hard-codes a macOS path.
- Overall score 0.55. The must-pass Functionality dimension is below threshold, which forces FAIL.

## Per-backend verdicts
| backend | gate | verdict |
|---|---|---|
| claude | off | not run |
| codex (adversarial, baseBranch) | required | fail (P1 /private/tmp on Linux; P2 order not pinned) |
| glm | off | not run |

Audit receipt: rcpt-396edf8dc56f98d72d193fd7. Plan source: config.gates; enforced_required: codex. The moai binary behind the MCP server was built at ee981a932, an ancestor of HEAD 772c1c050 (build lag reported by the tool). The audit_multi digest was not produced; codex_audit was called directly, so the orchestrator has no audit_multi convergence digest to check.

## Claim
1. The diff does what the card says for the resolver: base is `merge-base HEAD main`, then `origin/main`, with develop removed (ask 1).
2. Only one non-test change exists, the deny-list address, and it is justified (ask 1).
3. The RED commit precedes the fix and the RED claim is verifiable (ask 2).
4. The new tests are sound on the CI platform (the lane implied this). This claim is false.

## Evidence
Ask 1. `git diff --stat e4b747678...HEAD`: 6 files, `internal/hook/commit_identity_guard.go | 3 ++-`, `env_literal_diff_base_test.go | 101 +`, `env_literal_diff_test.go | 20 +-`, plus progress.md, verdict.md, CHANGELOG.md. The resolver now reads `for _, ref := range []string{"main", "origin/main"}` and returns `HEAD` when neither resolves. The signature is `srlDiffBase(t, root)`. The only non-test code change adds `"env-diff-base-test@example.invalid"` and a comment to `builtinCommitIdentityDenyEmails`. The new test passes `-c user.email=env-diff-base-test@example.invalid`, so the AC-CIG-010 drift test requires the entry. Run below: `--- PASS: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (2.13s)`. The list is a deny list, so adding a `.invalid` address cannot widen anything.

Ask 2. `git log`: 899f3c0cd (RED) precedes de719326c (fix) in the graph. The RED commit changes `srlDiffBase` to take `root` and rewrites its comment, but keeps the loop `[]string{"develop", "origin/develop"}` (shown by `git show 899f3c0cd -- internal/hook/env_literal_diff_test.go`). The parameterization is a signature-only change with no behaviour change, so the RED claim is verifiable from the graph: at 899f3c0cd the tests compile and exercise the develop loop. I did not re-run 899f3c0cd myself; the lane's recorded RED came from `git revert --no-commit de719326c`, not from checking out the RED commit. The RED claim is therefore credible but not re-observed by me.

Ask 3 (mutants, reasoned against the test bodies; not executed, because the mutant copy was denied by the session guard).
- Mutant A, develop first: test 1 has a stale develop, so merge-base differs from main's. Caught.
- Mutant B, return HEAD when main is absent: test 2 expects the origin/main merge-base, and HEAD is a later commit. Caught.
- Mutant C, `{"main","develop","origin/main"}`: test 1 resolves via main and test 2 has no develop, so both pass. NOT caught. A develop path reappears between main and origin/main unnoticed.
- Mutant D, `{"origin/main","main"}`: test 1 has no origin/main and test 2 has no main, so both pass. NOT caught. Codex observed the same ("선택 순서를 origin/main → main으로 뒤집어도 둘 다 통과"). The test names and comments claim main-first ordering that no test pins.

Ask 4. Command (env scrubbed, TMPDIR/GOTMPDIR=/private/tmp/aud-gotmp):
`unset MOAI_PROJECT_DIR CLAUDE_PROJECT_DIR OTEL_EXPORTER_OTLP_ENDPOINT && TMPDIR=/private/tmp/aud-gotmp GOTMPDIR=/private/tmp/aud-gotmp go test ./internal/hook -run 'TestSrlDiffBase|TestNoNewEnvLiteralsInDiff|TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration' -count=1 -v`
```
--- PASS: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (2.13s)
--- PASS: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (1.00s)
--- PASS: TestSrlDiffBaseFallsBackToOriginMain (0.15s)
    env_literal_diff_test.go:52: env-literal sweep: 2 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=e4b747678aae096b22769fdfd3efc5d89498006a, distinct literals=0
--- PASS: TestNoNewEnvLiteralsInDiff (0.10s)
ok  	github.com/modu-ai/moai-adk/internal/hook	4.103s
```
The swept count is non-zero (2). Those 2 lines are the guard-list entry and its comment line, both added in the non-test `internal/hook/commit_identity_guard.go` by d8428f9b6 inside a swept directory. That explains the lane's 2-vs-0 difference, because the earlier 0 was measured before that commit landed; the lane marked it uninvestigated. This is my inference from the diff, not a separate measurement.

Platform defect. `env_literal_diff_base_test.go:52` calls `os.MkdirTemp("/private/tmp", "env-diff-base-")`. `.github/workflows/ci.yml` lines 104-111 run the `test` job on `os: [ubuntu-latest]` only, and line 222 runs `go test ... ./...`. `/private/tmp` exists only on macOS (`ls -d /private/tmp` here: `drwxrwxrwt 16248 root wheel ... /private/tmp`). Codex ran both new tests in a Linux container: `stat /private/tmp: no such file or directory`, both fail. Other repo tests use `t.TempDir()`, `/tmp` or `/var/tmp`; no other test uses `/private/tmp`. I did not run on Linux myself; the Linux failure is codex's observation, backed by the CI matrix and the absence of the path.

Ask 5. `git grep -n 'origin/develop' | wc -l` gives 1919 on this tree. The survey file `.moai/reports/t1856/origin-develop-survey.txt` has 1903 lines (`wc -l`). It is gitignored (`.gitignore:235:.moai/reports/*`), so it is local-only. The difference of 16 is explained by the tree now carrying the lane's own docs: the survey was written at 01:17 and progress.md and the CHANGELOG entries, which mention `origin/develop`, came later. I did not diff the two lists, so this explanation is unverified. Survey section 6 of progress.md exists and lists per-file groups. It names a few test files such as `internal/factory/prlink_landed_forms_test.go (23)` as self-built fixtures. It does not mention the tests that run `merge-base develop HEAD` against a literal local `develop`: `internal/cli/codex_review_scope_test.go:163`, `mcp_selfreview_fixture_test.go:128`, `mcp_selfreview_test.go:281`, `factory_complete_test.go:409`, `factory_merge_test.go:99`. These are the same class of develop-based base lookup the card asked to list, and the grep on `origin/develop` alone cannot find them. See F3.

Ask 6. CHANGELOG (85487b706 then 772c1c050): the final text claims the resolver tries main then origin/main (true), two new tests pin that order (partly false: see mutants C and D, the order is not pinned), the fixture email is registered in the deny list (true), and "the one non-test change adds a fixture address ... so no other runtime behaviour changes" (true). The first version said "Test code only", which was wrong and was corrected. It also says "failing on main for reasons unrelated to any card" (supported by the t1632 A/B and the lane's 13273-line pre-fix measurement, which I did not reproduce).

Ask 7. verdict.md Gaps correctly state: no landing re-measure, no sync audit, a nearly empty sweep at HEAD, the uninvestigated 2-vs-0 difference, the six package skips, the untouched survey, and a foreign session holding the tree. progress.md section 5 carries slot receipts. Unsupported or overstated claims: verdict.md claim 3 and "RUN-EXIT PASS" never mention platform portability, and the lane never ran the tests on Linux although CI runs only Linux. The statement "two new tests pin that order" in the CHANGELOG is not borne out. I did not re-verify the 1538-pass full-package count (not re-run, by instruction).

Ask 8. `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1856` printed `config_status: ok`, codex gate `required`, claude/glm `off`, `cross_model_required: true`, `enforced_required: ["codex"]`. codex_audit (adversarial, baseBranch, project_root set to this tree) returned `verdict: fail` with receipt rcpt-396edf8dc56f98d72d193fd7. A failed required backend is not a pass.

## Baseline-attribution
Tree measured: HEAD 772c1c0507b4846693bdb6fa7bef111e6fe610a3 (base e4b747678, local main at lease time). Judging moai build: `moai version` printed `v3.2.0-rc.36-t1822c`, and the MCP server and codex tool report build commit ee981a932, a strict ancestor of the tree HEAD (build lag). The audit-plan output and the codex result are therefore from an older build than the tree; they were not rebuilt before judging. The go test run used the Go toolchain on this tree, not the moai binary, so the lag does not affect it.

## Gaps
- The mutants (ask 3) were reasoned against the test bodies, not executed: the copy-the-tree command was refused by the session guard and denied. Codex independently ran mutant D, which agrees with my reasoning; mutant C is reasoning only.
- 899f3c0cd was not re-run to observe RED.
- The Linux failure was observed by codex, not by me.
- The 16-hit difference in the origin/develop count was not diffed.
- The landing re-measure and the full `internal/hook` package were not run (by instruction), and the 1538-pass claim is the lane's, not mine.
- No audit_multi convergence digest exists; the orchestrator's `--result-file` check cannot be run on a digest I did not produce. The codex gate verdict is fail regardless.
- A refused command was recorded above (the mutant copy); the fallback was reasoning, stated as such.

## Residual-risk
- Even after the portability fix, the resolver depends on local `main` existing in the card's tree; a shallow CI checkout falls back to HEAD and sweeps nothing (known, stated by the lane).
- Other tests still pin a literal local `develop` for their base (F3). They are not in scope to repair, but the card asked for them to be listed.
- `git merge-base HEAD main` on a lane whose local main is behind origin/main would still pick the older base; the origin/main-versus-main preference is untested (F2).

## Findings
- F1 [high] [blocking] internal/hook/env_literal_diff_base_test.go:52 - `os.MkdirTemp("/private/tmp", ...)` fails on Linux, and CI runs only ubuntu-latest, so `TestSrlDiffBaseResolvesFromMainNotStaleDevelop` and `TestSrlDiffBaseFallsBackToOriginMain` fail in CI (codex observed `stat /private/tmp: no such file or directory`). Required fix: use `t.TempDir()` (or `os.MkdirTemp("", ...)`) in `envDiffTempRepo` and drop the manual cleanup. The guard against `$HOME` or repo placement is the operator's local TMPDIR discipline, not a reason to hard-code a macOS path.
- F2 [medium] [blocking for the AC "RED first / pins main-first"; the CHANGELOG claims the order is pinned] internal/hook/env_literal_diff_base_test.go:70-100 - no test has both local main and origin/main present, so reversing the order to `origin/main, main`, or inserting `develop` between them, passes both new tests (mutants C and D). Required fix: add a case with a stale `origin/main` ref and an advanced local main, asserting local main's merge-base is chosen, and a case with a stale develop plus origin/main only. Alternatively soften the CHANGELOG sentence to what the tests prove.
- F3 [medium] [optional, card scope] .moai/reports/t1856/progress.md:180-210 - the survey greps only `origin/develop` and so misses the same develop-based base lookups that use a literal local `develop`: `internal/cli/codex_review_scope_test.go:163`, `internal/cli/mcp_selfreview_fixture_test.go:128`, `internal/cli/mcp_selfreview_test.go:281`, `internal/cli/factory_complete_test.go:409`, `internal/cli/factory_merge_test.go:99`, plus `internal/cli/todo_issuance_probe_test.go:5` (comment). Required fix: add this list to the card's survey. This is a listing duty only; the repair stays limited to the test base resolver.
- F4 [low] [optional] .moai/reports/t1856/progress.md section 6 - the survey count was 1903 at 01:17 and is 1919 on HEAD now; the difference is probably the lane's own docs but is not stated. Required fix: record the recount.
- F5 [low] [optional] CHANGELOG.md - "two new tests pin that order" overstates (see F2). Required fix: reword unless F2 is repaired.

## Recommendations
- Repair F1 and F2 in one fix commit, re-run the three-pattern command above, then re-audit only the F1 and F2 delta; the resolver itself needs no change.
- Extend the survey with the literal `develop` base lookups (F3) in the progress record.
- Re-run codex after the fix; the required gate is codex.
