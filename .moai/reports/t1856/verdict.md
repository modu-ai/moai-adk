# Card t1856 — lane run-exit verdict (Class B, no SPEC)

Result: RUN-EXIT PASS-WITH-GAPS. This is the lane's own verdict at the end of the run stage; sync, the independent sync audit and the landing re-measure have NOT happened yet, and the card is not landable on this file alone.

## Claim
1. Cause: `srlDiffBase` (internal/hook/env_literal_diff_test.go) resolved its diff base only through `develop` / `origin/develop`; `origin/develop` no longer exists and local `develop` is a stale branch, so the base became an old fork point and `TestNoNewEnvLiteralsInDiff` counted every `MOAI_KANBAN_*` literal that landed on main since then as "added by this card".
2. Fix: the ref list is now `main`, then `origin/main` (commit `de719326c`); two new tests pin the resolver (`899f3c0cd` RED), and `d8428f9b6` registers the new fixture email in the commit-identity deny list.
3. No existing test was weakened, skipped or deleted by this card; the full `internal/hook` package passes on the final code tree.

## Evidence (lane-observed in this run, tree `d8428f9b6`, HEAD now `b3a78a7f4` with a docs-only delta)
- GREEN at `d8428f9b6`: `go test ./internal/hook -run 'TestSrlDiffBase|TestNoNewEnvLiteralsInDiff' -count=1 -v` -> exit 0; `--- PASS` for `TestSrlDiffBaseResolvesFromMainNotStaleDevelop`, `TestSrlDiffBaseFallsBackToOriginMain`, `TestNoNewEnvLiteralsInDiff` (`2 added lines swept ... base=e4b747678aae096b22769fdfd3efc5d89498006a, distinct literals=0`).
- Pre-fix RED, observed after `git revert --no-commit de719326c` (never committed; `git revert --abort` afterwards, `git status --porcelain` 0 lines, HEAD `d8428f9b6`): same command -> exit 1; both new tests FAIL (`srlDiffBase picked ...; want merge-base with local main ...` and `picked HEAD; want ... origin/main`); `TestNoNewEnvLiteralsInDiff` FAIL with `13273 added lines swept ... base=966f1b15852ed95d09ab32b7930142bbf1672ecb ... [MOAI_KANBAN_CARD]`.
- Full package: `go test ./internal/hook -count=1 -v` inside the go-test-heavy slot -> exit 0, `ok github.com/modu-ai/moai-adk/internal/hook 244.941s`; top-level `--- PASS` 1538, `--- FAIL` 0, `--- SKIP` 6 (counted with `grep -c` on the output file). `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration` and `TestNoNewEnvLiteralsInDiff` are among the passes.
- Deciding lines and the earlier lane's observations are recorded in `.moai/reports/t1856/progress.md` (sections 1-5, 8, 9).

## Baseline-attribution
- Code tree measured: `d8428f9b6` (base `e4b747678`, local main at lease time). `b3a78a7f4` adds only `progress.md`.
- The pre-fix 13273-line / base `966f1b158` result is the same failure the lane measured independently on card t1837 (13256 lines, same base) and the earlier t1632 A/B; those two are corroboration, not this run's measurement.

## Gaps
- Landing re-measure (b2cd scope: internal/hook plus direct importers, internal/cli per standing rule 35e9) has NOT run. Nothing here is a landing verdict.
- Sync stage and the independent sync audit (required backend codex) have not run.
- At HEAD the sweep is nearly empty (HEAD is close to its base), so a green `TestNoNewEnvLiteralsInDiff` here asserts little by itself; resolver behaviour is carried by the two new tests and the pre-fix RED.
- The earlier record showed `0` swept lines at `d8428f9b6`'s base and this run shows `2`; the difference was not investigated.
- Six package-own skips (`TestFactoryHookBenchmarkBudget`, `TestConsumerOnly_M0AndMxByteUnchanged`, `TestSessionStart_DriftCacheProbe`, `TestDriftFillBurstHelper`, `TestSessionStart_HandleStageProbe`, `TestScanWriteContent_SupportedExtension_ScannerNotAvailable`); no `-skip` flag was passed.
- The `origin/develop` survey (1903 hits) is follow-up material and untouched; this card fixes only the internal/hook resolver.
- Another writer's session (lane-13, pid 38407) holds this tree as its cwd; it showed no write activity (state `S+`, one open file under the tree) at hand-over.

## Residual-risk
- Other tests or tools may still resolve a develop-based base (survey only).
- The fix reads local `main` first; a repository without `main` and without `origin/main` falls back to `HEAD`, which sweeps nothing.
