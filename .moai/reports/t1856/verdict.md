# Card t1856 — lane verdict (Class B, no SPEC)

verdict: PASS
audited_sha: b234210fcd1f3e912a974032b8a5aa4b8e7bd3d4

Result: PASS at the audited head b234210fc (after absorbing main; the card's own code is byte-identical to the substantively audited c4a7a79af). Basis: the independent sync audit, round 5 (a read-only delta after the absorb), PASS 0.93 with the required codex backend pass, see `.moai/reports/t1856/sync-audit.md`; rounds 1 and 2 were FAIL, are preserved in git (6b97e64bc, 5e8c07915) and were closed by repairs the lane observed itself (mutants C, D, F executed by the lane). A first landing re-measure was valid on tree e711d0186175 before main moved; the re-measure on the final post-absorb tree runs right after this file is committed and is recorded in the landing record, not here, so this file alone is not the landing verdict.

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
- Landing re-measure (b2cd scope: internal/hook plus its five direct importers in full, internal/cli by rule 35e9 as `-run '^TestVersionStamp'`): valid once on tree e711d0186175 (both calls exit 0, test counts not shown by the verb), invalidated by main moving to f4bcd4f2c; the final-tree run has not happened at the time of writing. The rest of internal/cli is unobserved locally and carried to the next batch CI run.
- The sync audit's Ubuntu CI run is not observed: codex observed a Linux/arm64 run of the six resolver tests only; the full internal/hook package and the CI matrix are not part of that observation.
- The stage record's evidence SHAs predate the absorb (c2e633946 for sync, a574521ea for the audit) and cannot be re-stamped; the merge gates read this file, the audit file and the re-measure record against the current tree instead.
- A receipt-refusal record (`sync-auditor--unknown-spec.json`, agent a0a03c32473b3e55c) was left in place by the leader because round 3 was PASS; cause "verdict line missing" is unexplained (each auditor's final message did end with a verdict line).
- At HEAD the sweep is nearly empty (HEAD is close to its base), so a green `TestNoNewEnvLiteralsInDiff` here asserts little by itself; resolver behaviour is carried by the two new tests and the pre-fix RED.
- The earlier record showed `0` swept lines at `d8428f9b6`'s base and this run shows `2`; the difference was not investigated.
- Six package-own skips (`TestFactoryHookBenchmarkBudget`, `TestConsumerOnly_M0AndMxByteUnchanged`, `TestSessionStart_DriftCacheProbe`, `TestDriftFillBurstHelper`, `TestSessionStart_HandleStageProbe`, `TestScanWriteContent_SupportedExtension_ScannerNotAvailable`); no `-skip` flag was passed.
- The `origin/develop` survey (1903 hits) is follow-up material and untouched; this card fixes only the internal/hook resolver.
- Another writer's session (lane-13, pid 38407) holds this tree as its cwd; it showed no write activity (state `S+`, one open file under the tree) at hand-over.

## Residual-risk
- Other tests or tools may still resolve a develop-based base (survey only).
- The fix reads local `main` first; a repository without `main` and without `origin/main` falls back to `HEAD`, which sweeps nothing.
