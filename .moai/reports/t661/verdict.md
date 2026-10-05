# t661 verdict — internal/cli update tests must not touch the real home

Card: t661 (Class B, test isolation) · branch `WT-update-test-home` · measured tree HEAD `1694ff8c0`
Base merge `21ee20493` (^1 `987eb7e40` origin/develop, ^2 `1bea05daa` local develop). Nothing pushed.
Card diff vs base under `internal/`: test files only, 3 files, +222 (`git diff --stat 21ee20493 HEAD -- internal/`).

## Claim
1. Before the repair, `TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently` and
   `TestRunUpdate_ThreeRunIdempotency_V3Project` could structurally reach `ensureGlobalSettingsEnv()`
   (`internal/cli/update.go:958`, via `update_template_sync.go:567`), which removes `~/.claude/hooks/moai` and
   rewrites `~/.claude/settings.json` through `userHomeDirFn` with no test-side guard. Issue1086 and the
   SkipSync force subtest cannot reach it (code reading).
2. Repair A+C (lead decision) is in place: A — `homeSeamSpy` + a not-real-home pre-assertion on the two
   reachable tests, pre-assertion only on the two unreachable tests; C — TestMain wraps `userHomeDirFn` so only
   the captured real home is redirected to a sandbox, with guard test `TestUserHomeDirFnSandboxesRealHome`.
3. Both guard branches are non-vacuous: each mutant fails exactly the predicted subtest.
4. Across every slot run, the real-home fingerprint (7 files + hooks dir absence) did not change.
5. Among the tests that reported a result in the full package run, the only failures are the four known reds.

## Evidence
- Code reading: `.moai/reports/t661/code-reading.md` (commit `c3c8f867f`).
- Repair: commit `7de194e23` (`main_test.go`, `update_clean_install_test.go`, `update_skip_sync_test.go`).
- Pre-scan of C side effects: `.moai/reports/t661/c-wrapper-prescan.md`.
- Slot (commit `f84538b14`): `go vet ./internal/cli/` exit 0 (`slot-vet.txt`); the 8 named tests all PASS, exit 0,
  redirect line present (`slot-run-tests.txt`).
- RED prediction pinned before injection: `red-prediction.md` (commit `f84538b14`).
- Mutants (commit `8861553bd`), selector `-run '^TestUserHomeDirFnSandboxesRealHome$'`:
  - M1 (TestMain sandbox call removed): `real_home_redirects_to_sandbox` FAIL, `overridden_home_passes_through`
    PASS, exit 1, redirect lines 0 (`red-m1.txt`); revert diffstat 0 bytes.
  - M2 (`if err != nil {`): `real_home_redirects_to_sandbox` PASS, `overridden_home_passes_through` FAIL, exit 1
    (`red-m2.txt`); revert diffstat 0 bytes.
  - Pre-checks before each run: `red-m1-precheck.txt`, `red-m2-precheck.txt` (no other internal/cli compile).
- Home fingerprint around the targeted slot: sha256 diff exit 0; stat lines 1-7 diff exit 0 (`home-compare.md`).
- Full package run (commit `1694ff8c0`, `full-run/summary.md`): `go test ./internal/cli/ -count=1 -timeout 1200s -v`
  exit 1 with `panic: test timed out after 20m0s` (`cli-full.txt:11673`). `--- FAIL` top-level exactly:
  TestHomeStateChangedSurfaceCoverageConsumesFreshProfile, TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite,
  TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition, TestAuditLagUsesBinlagSeam. SKIP 15, all
  env opt-in or environment conditions. C redirect log lines: 1 (`cli-full.txt:149`). Home fingerprint before/after
  with the same command set and separate stderr: `cmp` exit 0 for all 8 streams.
- Lead read the full-run evidence independently and ruled: no local re-run; the unfinished part is a Gap covered by
  CI after merge and push.

## Baseline-attribution
All runs above were measured in this lane on branch `WT-update-test-home`: the targeted slot and mutants on the
tree at `f84538b14`/`8861553bd` (test code identical to `7de194e23`), the full run on HEAD `8861553bd` clean
(`full-run/tree-head.txt`). Go toolchain go1.26.8 darwin-arm64 (from the panic trace). The known-red list is the
lead's; it was not re-derived on a tree without the card.

## Gaps
- Full package run incomplete: of 3614 defined top-level tests, 2992 started and 2739 finished; 622 never started and
  253 (parallel, paused) have no result. Load before `11.78`, after `18.30` on 16 cores. Lead ruling: CI's full run
  after the batch push is the verdict basis for these; CI runners also have HOME equal to the real home, so C takes
  the same branch there.
- Whether a tree without the card also exceeds 1200s under this load was not measured.
- Whether every sync step succeeds on the two v3 fixtures (actual reach of the sink before the repair) was not
  executed on an unrepaired tree; claim 1 is a code-path reading.
- Sibling candidates listed in `code-reading.md` (`update_mode_test.go`, `update_hooks_guidance_test.go`,
  `coverage_test.go`, `update_mirror_heal_test.go`, `update_deny_migration_test.go`, `update_llm_preserve_test.go`,
  `integration_test.go`) are covered by C by reading only; no per-test measurement.
- TTY stdin under `go test` on this machine was not established (SkipSync force path is unreachable only under
  non-TTY).
- Integration-window re-measurement on the merged tree is still to come.

## Residual-risk
- C covers only the `userHomeDirFn` seam (and the glmcred/kanban closures over it). Production sites that call the
  home resolver directly are not redirected (HEAD `1694ff8c0`): `update.go:886` and `migrate_agency.go:720`
  (`paths.Home()`), `doctor_disk.go:220` (`paths.Home()`), `doctor.go:602`, `statusline.go:75`, `tokens.go:223`
  (`os.UserHomeDir()`).
- `TestAuditPinLive_GLMDifferential` with `MOAI_AUDIT_PIN_LIVE=1` would now read the empty sandbox credential path
  and take its SKIP branch instead of the live call (default runs unaffected; observed skip reason in the full run
  was the env gate).
- A timeout panic skips TestMain cleanup: the sandbox dir survives. Observed once
  (`/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-1490004665`); removed with `rmdir` on lead
  approval, exit 0. The piped `ls -A | wc -l` pre-count read 3 because the profile's `ls` alias prints the
  `total 0`, `.`, `..` lines (the earlier listing showed only those); `rmdir` succeeding is the mechanical witness
  that it was empty.
- Instrument defect in the targeted-slot fingerprint: before/after hooks absence was captured with different tools
  (`bfs`/`stat` vs `stat`) and stderr mixed into the stat stdout file, which produced a literal `diff` exit 1 with no
  content change. Fixed for the full run (same command set, separate stderr).
- Base merge attempt 1 failed with `fatal: Unable to write index.` (exit 128; HEAD unchanged, MERGE_HEAD absent,
  index.lock absent, no manual deletion); attempt 2 succeeded. Cause not established.
