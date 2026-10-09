# SPEC-UPDATE-MIGRATION-FIX-001 — Progress

SPEC ID: SPEC-UPDATE-MIGRATION-FIX-001
Card: t1578
Status: in-progress (run phase)
Tier: M

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: spec.md, plan.md, acceptance.md, research.md,
  decision-index.md, progress.md (this file) — all under
  `.moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/`.
- Authored on tree: 2aab5f797 (branch WT-update-migration-fixes, card
  worktree `.moai/worktrees/t1578`); the six artifacts landed as lane
  commit f569be5d8 — the plan-artifact baseline is f569be5d8, and the
  anchor-path delta 2aab5f797..f569be5d8 is empty (acceptance.md EV-7),
  so all plan-phase measurements carry over. BASELINE RE-CUT NOTICE: the
  worktree fast-forwarded from 81786284e to 2aab5f797 (main absorbed, 155
  commits) mid-research. All research anchors re-verified on 2aab5f797
  (acceptance.md EV-6): the deny-migration map, its test file, and
  internal/template/deployer_mode.go are UNCHANGED between the two bases;
  internal/cli/update.go carries a 57-line change (reconciliation preview
  rename, migrateProjectCommonAssets removal-arm relocation into the sync
  flow gated on userAssetsInstalled, participation step added to the skip
  block) — none of it touches the SPEC's conclusions; the skip-path block
  was re-read byte-identical and still carries NO integrity probe.
- Plan-phase measurements recorded: acceptance.md evidence ledger EV-1
  through EV-7 (all read-only, all run in this worktree; EV-6 is the
  re-verification batch on the re-cut baseline, including the
  normalization guard re-run — `ok ... 2.866s`, exit 0 on 2aab5f797; EV-7
  extends the anchors to f569be5d8 and records the EV-5 grep's actual
  execution). Authoring-discipline correction recorded at EV-5: one
  ledger row was initially written before its command ran; it has been
  re-measured and the correction is stated in the entry itself.
- Scope decisions: card item (3) out-of-scope with rationale (spec.md C.1);
  card item (4) in-scope as M3 (spec.md C.2); both recorded in
  decision-index.md (Q3 evidence-needed; Q1/Q2 implementation-level
  defaults applied at plan close).
- Known-issue classification: K1 (card P1) and K2 (card P2) verified as
  already repaired upstream on this tree — K1 by t1569 M2 / PR #1792
  (measured: normalization guard green), K2 by SPEC-USER-ASSET-INSTALL-001
  mechanism retirement. The only new implementation is the version-match
  integrity probe (K3). Branch contingency: if run-phase M1-b measures a
  live empty-directory producer, M2 escalates to repair per spec.md R1.
- Plan status: audit-ready.
audit_ready: true

## §E.2 Run-phase Evidence

Attribution for every entry: the command, its verbatim output, and the
baseline (this run, this tree, HEAD SHA at capture). The M1 capture HEAD is
`008d2e2a7` (branch `WT-update-migration-fixes`, worktree
`.moai/worktrees/t1578`).

### Pre-flight (SPEC §C, run 2026-10-10, HEAD 008d2e2a7)

- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578` (matches the spawn value)
- `git branch --show-current` → `WT-update-migration-fixes`
- `git rev-parse --short HEAD` → `008d2e2a7`
- `go build ./...` → exit 0, no output
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0, no output
- `golangci-lint run --timeout=2m 2>&1 | tail -5` → `0 issues.` (baseline: zero findings before any edit)
- Working tree before this run: one modified file, `progress.md`. Its only
  change was one appended §G line (ceiling-outcome record, 2026-10-09T16:19:39Z)
  written before this run. It is kept byte-identical and is committed with M1.

### M1 — measurement record (no implementation code)

**M1-a — K1 reproduction attempt (SPEC §C C3 command, re-measured on 008d2e2a7).**

```
$ go test ./internal/cli/ -run 'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.337s
```

Exit 0. Classification: **K1 repaired** by card t1569 M2 (PR #1792, commit 87da06367). The guard is GREEN on this tree, so no migration code change is authorized.

Pin-strength read (M2-a, read-only): the test-owned `legacyRootDenySpecifiers` list has 9 entries (`internal/cli/update_deny_migration_test.go:423-433`). The fixture `legacyRootDenyFixture` carries all 9 plus two user-custom rules. The test asserts that every legacy form is absent, every canonical form is present, the custom rules survive, and the `[settings] Normalized` line is printed. The pin is at least as strong as the plan describes. No code change is needed for M2-a.

**M1-b — K2 reproduction attempt (fixture measurements, temporary probe, not committed).**

The probe was a temporary `internal/cli/zz_m1b_measure_test.go`, deleted before this commit. Command: `go test ./internal/cli/ -run 'TestM1B_' -count=1 -v`.

Project side. A full file-level template sync runs on a v3 fixture through the real `runUpdate` (`runUpdateInFixture`):

```
M1B-PROJECT update output lines=45
M1B-PROJECT root-absent: .claude/skills (stat .../.claude/skills: no such file or directory)
M1B-PROJECT root-absent: .claude/agents/moai (stat .../.claude/agents/moai: no such file or directory)
M1B-PROJECT root-absent: .agents/skills (stat .../.agents/skills: no such file or directory)
M1B-PROJECT root-absent: .codex/agents/moai (stat .../.codex/agents/moai: no such file or directory)
(Elision: the temporary sandbox fixture path prefix is shown as `...` in the four stat lines above; every other character is verbatim.)
M1B-PROJECT RESULT swept=0 zero_file=0
--- PASS: TestM1B_ProjectSideSyncFixture (0.90s)
```

Installer side. The production installer (`newUserAssetInstaller(tmpHome)`, which is the same constructor `runUserAssetUpdatePhase` uses) installs L0 plus all 11 optional packs into a temp user home:

```
M1B-INSTALL packs=11 installed=686 refreshed=0 failures=0 collisions=0 divergences=0
M1B-INSTALL RESULT catalog_skill_targets_checked=116 empty_targets=0 missing_targets=0
M1B-INSTALL SWEEP user_root_dirs=208 zero_file=0
--- PASS: TestM1B_InstallerSideCatalogSelection (0.71s)
```

Classification: **K2 mechanism retired on this tree.** The project payload carries no managed skill or agent root (`swept=0` is the by-design exclusion from `isCommonAssetRoot`). The installer produced no empty directory target under the widest selection. Branch decision: **"mechanism retired; guard to be pinned in M2."** Contingency R1 (escalate M2 to repair) is NOT triggered. The REQ-UMF-006 verify-and-report is still implemented unconditionally in M2, as the plan requires.

**M1-c — K3 RED anchor (EV-4 command re-run on 008d2e2a7).**

```
$ sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go
	if syncSkipped {
		// A version-matched update runs no sync and no merge, so the retired
		// per-agent model/effort keys are stripped here, after its own backup.
		// A user-cancelled merge returns the same skipped=true; the helper
		// re-evaluates the version predicate and leaves that case untouched.
		if err := stripRetiredModelConfigOnVersionMatch(cmd, out, "."); err != nil {
			updateLedger.requiref(sevWarn, "retired model-key removal failed: %v", err)
		}
		// Card t1527 D5 + repair round: the deferred render carries the block
		// on this early return too — no explicit call here.
		// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-004): a version-matched
		// update is still a finished plain template-sync run, so the ask runs
		// here too; its own gates (mode flags, terminal, CI, asked) decide
		// whether anything prompts. A failure warns; it never fails the update.
		if err := runParticipationStep(cmd, out); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
		}
		runParticipationFlushAtUpdate(cmd.ErrOrStderr())
		return nil
	}
```

Exit 0. The block holds no integrity probe call. **RED anchor for AC-UMF-001 and AC-UMF-002 confirmed on 008d2e2a7.**

**B2 cross-SPEC conflict scan (run before any edit).**

`grep -rn "Retired\|superseded" internal/cli internal/userassets` returns 228 lines across 84 files. The matches in the files this run touches (`update.go`, `update_deny_migration*.go`, `update_model_key_strip.go`, `update_template_sync.go`, `user_asset_phase.go`, `install.go`, `installer_test.go`) concern the retired v2 deny-rule strip, the retired model-key strip, and superseded SPEC references. None reverses this SPEC. **No reversal to record.**

**C2 — production evidence preserved.** `cp /tmp/moaikr-force-update.log .moai/reports/t1578/moaikr-force-update.log`. Both files have sha256 `c1c17110f013d69a398270f0686083e58963464e7a7f1816f0cdd965759edc47`. The copy sits under the gitignored `.moai/reports/*` path (`.gitignore:235`), so it is a local, uncommitted copy.

**M1 findings.** K1 is repaired and guarded. K2 is retired, with the guard pinned in M2 and no repair escalation. K3's RED anchor is confirmed. M2 proceeds as the regression-guard milestone.

### M2 — regression pins and the REQ-UMF-006 report

**M2-a (K1 pin, verification only).** The test-owned `legacyRootDenySpecifiers` list (9 forms) and the `legacyRootDenyFixture` cover all nine forms, so the pin is at least as strong as the plan describes. The pin is green on 008d2e2a7 (M1-a). No code change was made for M2-a.

**M2-b(1) — REQ-UMF-006 installer empty-target report (RED first).** New test `TestInstaller_RejectsEmptyDirectoryTargets` in `internal/userassets/installer_empty_target_test.go`. The catalog carries one directory entry whose source tree is an explicit empty directory (`fstest.MapFile` with `fs.ModeDir`).

RED, observed before `install.go` was changed (verbatim):

```
=== RUN   TestInstaller_RejectsEmptyDirectoryTargets
    installer_empty_target_test.go:49: REQ-UMF-006: the empty directory target was not reported as a failure; failures=[]
--- FAIL: TestInstaller_RejectsEmptyDirectoryTargets (0.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/userassets	0.320s
FAIL
EXIT=1
```

Implementation: `installTargets` now also returns the names of directory entries that yield no file target. `Install` reports each one as a `FileOutcome` failure (`Path` = `<name>/`, reason beginning `empty install target`). Nothing is counted as installed for it.

GREEN after the change (verbatim):

```
=== RUN   TestInstaller_RejectsEmptyDirectoryTargets
--- PASS: TestInstaller_RejectsEmptyDirectoryTargets (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/userassets	0.664s
```

Full package after the change: `go test ./internal/userassets/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.566s`.

**M2-b(2) — K2 project-side guard (`TestTemplateSync_LeavesNoEmptyManagedSkillDirs`, `internal/cli/update_managed_dir_guard_test.go`).** This guard is green on this tree by design, because the healthy swept count is zero (M1-b). It cannot be observed red on this tree, so its falsifiability is shown by the sub-test in (3). Verbatim:

```
=== RUN   TestTemplateSync_LeavesNoEmptyManagedSkillDirs
    update_managed_dir_guard_test.go:81: managed skill/agent directories swept=0 zero_file=0 (healthy exclusion contract: swept=0)
--- PASS: TestTemplateSync_LeavesNoEmptyManagedSkillDirs (0.40s)
```

Anti-vacuity: the guard first asserts that the sync wrote `.claude/settings.json`, so a zero sweep cannot pass when the sync did not run.

**M2-b(3) — failure arm of the sweep (`TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir`).** A planted known-failing input: one empty managed directory and one filled one. Under the correct predicate:

```
=== RUN   TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir
--- PASS: TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.369s
```

Falsifiability: the predicate was temporarily inverted (`if dirHoldsRegularFile(p)`) and the sub-test was run. Verbatim failure (the `t.TempDir()` prefix is elided as `<tmp>`; every other character is verbatim):

```
--- FAIL: TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir (0.00s)
    update_managed_dir_guard_test.go:114: sweep must flag exactly the planted empty directory <tmp>/.claude/skills/moai-planted-empty, got zero_file=[<tmp>/.agents/skills/moai-planted-filled]
FAIL
exit 1
```

The predicate was then restored to `if !dirHoldsRegularFile(p)` before the GREEN run above.

**M2-c — scoped run and full-package attempt.**

Scoped (the update-path families, `-count=1 -v`, log `.moai/reports/t1578/m2c-scoped-cli.log`): 16 runs, all PASS, `ok  	github.com/modu-ai/moai-adk/internal/cli	2.973s`, `EXIT=0`. The swept names include the two M2 guards `TestTemplateSync_LeavesNoEmptyManagedSkillDirs` and `TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir`, and the 9-form pin `TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries`.

Full package, `go test ./internal/userassets/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.566s` (exit 0).

Full package, `internal/cli` — **NOT COMPLETED in this environment.** The run was `go test ./internal/cli/ -count=1 -timeout 25m` (log `.moai/reports/t1578/m2c-cli-suite.log`, gitignored). It was stopped by SIGQUIT after about 14 minutes blocked inside `TestFactoryCompleteNoWindowGitHubFlow`. The goroutine dump places the blocked frame at `internal/cli/factory_card_pr_test.go:157` (`ghfFixture.remoteBranchTip`, an `exec.Cmd.Output` waiting on a child's output pipe), under `factory_card_pr_test.go:580`. The package then ended with `FAIL	github.com/modu-ai/moai-adk/internal/cli	864.912s` and `EXIT=1`. The log holds 56 top-level and 18 subtest `--- FAIL` lines, all in the codex and doctor-codex families: `TestCodex*`, `TestCheckCodexWiring_*`, `TestCountCodexAgentTOMLs*`, and `TestStopChainMemberCostWithinBudget`. Their failure text includes `install the desktop app from https://example.invalid/codex-desktop`, and a codex CLI does exist at `/Users/goos/.local/bin/codex`, so the cause was not isolated further. No update-family test appears in the failure list, because the update-family tests had not started: the factory and codex test files sort before the update files. The run also overlapped another session's test processes on this machine.

Attribution: the failing tests and the blocked test sit in files this change does not touch, and they run before any file this change adds. That is an ordering and code-path argument, not a clean-baseline measurement. The full-package requirement of M2-c is therefore **not met**, and the scoped families above are the evidence for M2.

### M3 — version-match integrity probe (REQ-UMF-001..003)

Files: `internal/cli/update_integrity_probe.go` (new), `internal/cli/update_integrity_probe_test.go` (new), and `internal/cli/update.go` (one insertion of nine lines, zero deletions).

**RED (observed before any M3 production code existed; verbatim assertion lines; line numbers are from the RED state of the test file):**

```
=== RUN   TestRunUpdate_VersionMatch_RunsIntegrityProbe
    update_integrity_probe_test.go:130: REQ-UMF-001: the version-match run must print one integrity row naming .claude/settings.json; rows=[]
         ✓ Up to date · Skipping sync
--- FAIL: TestRunUpdate_VersionMatch_RunsIntegrityProbe (0.73s)
=== RUN   TestRunUpdate_UserCancelled_SkipsIntegrityProbe
--- PASS: TestRunUpdate_UserCancelled_SkipsIntegrityProbe (0.09s)
=== RUN   TestIntegrityProbe_FailOpen
    update_integrity_probe_test.go:187: REQ-UMF-002: want exactly one integrity row for the unreadable manifest, got 0
         ✓ Up to date · Skipping sync
--- FAIL: TestIntegrityProbe_FailOpen (0.38s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.349s
```

The `✓ Up to date · Skipping sync` line shows the version-match skip path was reached before the assertion failed. Output lines that are not assertions are elided with an indent-only gap. `TestRunUpdate_UserCancelled_SkipsIntegrityProbe` passes in RED because no probe exists yet. It is only a discriminating witness once the probe exists, and the mutation below shows that.

**Implementation.** The probe checks a fixed set: `.claude/settings.json` (exists and parses as JSON), `.moai/config/sections/system.yaml` (exists and is non-empty), and `.moai/manifest.json` (exists and parses as JSON). Each damaged entry produces one row through `tui.CheckLine("warn", "Integrity", "<path> (<reason>)", "", &th)`. The rendered row is `!  Integrity  <path> (<reason>)`, verbatim from the mutation run below. A deferred `recover` turns any internal fault into one warn row and never returns an error, so the probe cannot fail the update (REQ-UMF-002). Wiring, as the first statement of the `if syncSkipped {` block and before the model-key strip:

```go
if updateSkippedOnVersionMatch(cmd, ".") {
    runManagedSurfaceIntegrityProbe(out, ".")
}
```

**GREEN.** The three M3 tests pass in the final scoped batch (listed under E2 below).

**Falsifiability (mutations, each restored and re-verified by the final batch).**

- REQ-UMF-003 guard removed (the probe made unconditional inside `if syncSkipped {`): `--- FAIL: TestRunUpdate_UserCancelled_SkipsIntegrityProbe (0.11s)` with `update_integrity_probe_test.go:157: REQ-UMF-003: a user-cancelled merge printed integrity rows ["!  Integrity  .claude/settings.json (missing)" "!  Integrity  .moai/manifest.json (missing)"]`. This shows the cancel test discriminates the cancelled-merge entry from the version-match entry.
- JSON validity check disabled (`if false && !json.Valid(data)`): `--- FAIL: TestIntegrityProbeEntry_DamageReasons/unparseable_json (0.00s)` with `damageReason(.claude/settings.json) = "", want "unparseable"`.
- M2 sweep predicate inverted: recorded under M2-b(3).

**Coverage, new code (`go test ... -coverprofile`, then `go tool cover -func`):**

- `update_integrity_probe.go:59 damageReason` — **100.0%**
- `update_integrity_probe.go:91 runManagedSurfaceIntegrityProbe` — **85.7%** (the uncovered statement is the deferred `recover` branch)
- `install.go installTargets` — 85.7%; `install.go Install` — 86.7%; package `internal/userassets` — 80.1% of statements (run output `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.716s	coverage: 80.1% of statements`). `dirTargets` at 84.6% is pre-existing and unchanged by this run.
- Package-level coverage of `internal/cli` cannot be measured here, because the full package run does not complete (see M2-c).

### M4 — convergence evidence (E1–E8)

**E1 — AC matrix.**

| AC | Status | Command | Actual output (this run, tree = HEAD `6bb0d8278` + the working-tree change set) |
|----|--------|---------|-----|
| AC-UMF-001 | PASS | `sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go` plus `TestRunUpdate_VersionMatch_RunsIntegrityProbe` and `TestRunUpdate_UserCancelled_SkipsIntegrityProbe` in the scoped batch | probe call inside the block (E3 witness); both tests `--- PASS` in the final batch; cancel-path mutation witnessed as above |
| AC-UMF-002 | PASS | `TestRunUpdate_VersionMatch_RunsIntegrityProbe`, `TestIntegrityProbe_FailOpen` | a damaged path is named by one row (`.claude/settings.json`); `RunE` returns nil; the fail-open case yields exactly one row and no error |
| AC-UMF-003 | PASS | `TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries` (M1-a re-run, final batch) | M1-a: `ok  	github.com/modu-ai/moai-adk/internal/cli	1.337s`; final batch `--- PASS` |
| AC-UMF-004 | PASS | `TestTemplateSync_LeavesNoEmptyManagedSkillDirs`; `go test ./internal/userassets/ -run TestInstaller_RejectsEmptyDirectoryTargets -count=1 -v` | `swept=0 zero_file=0`; `--- PASS: TestInstaller_RejectsEmptyDirectoryTargets (0.04s)` |
| AC-UMF-005 | PASS (plan gate) | `grep -c "OUT OF SCOPE" .moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/spec.md` | `1` (spec.md §C.1) |

**E2 — cross-platform build and scoped batch.**

- `go build ./...` → exit 0, no output.
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0, no output.
- Scoped batch (final, `go test ./internal/cli/ -run 'TestRunUpdate_V3Path|TestStripRetiredV2Deny|TestTemplateSync_|TestSkipSync|TestRunUpdate_VersionMatch_RunsIntegrityProbe|TestRunUpdate_UserCancelled_SkipsIntegrityProbe|TestIntegrityProbe' -count=1 -v`; log `.moai/reports/t1578/e2-final-scoped-cli.log`): `ok  	github.com/modu-ai/moai-adk/internal/cli	3.832s`, `EXIT=0`, 28 runs. Swept test names (all `--- PASS`): `TestStripRetiredV2DenyEntries_RemovesRetiredKeepsCustom`, `_Idempotent`, `_V3CleanUntouched`, `_MissingFileNoop`, `_NoPermissionsKeyNoop`; `TestRunUpdate_V3Path_StripsRetiredDenyEntries`, `_CleanSettingsUntouched`, `_NormalizesLegacyRootDenyEntries`; `TestRunUpdate_VersionMatch_RunsIntegrityProbe`; `TestRunUpdate_UserCancelled_SkipsIntegrityProbe`; `TestIntegrityProbe_FailOpen`; `TestIntegrityProbeEntry_DamageReasons` with 8 subtests (`missing`, `directory_in_place_of_file`, `intact_json`, `unparseable_json`, `empty_system_yaml`, `intact_system_yaml`, `stat_fails_under_a_file`, `read_denied`); `TestTemplateSync_LeavesNoEmptyManagedSkillDirs`, `TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir`, `TestTemplateSync_MirrorNoticeGoesToStderrNotStdout`, `TestTemplateSync_NoFallbackEmitsNothing`, `TestTemplateSync_PlainDeployerStillDeploys`; `TestSkipSyncNoArchive` with 2 subtests.
- `go test ./internal/userassets/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.566s`.
- The `internal/cli` full-package requirement is not met (M2-c). Its scoped substitute is the batch above.

**E3 — coverage and boundary witness.** Coverage is in the M3 section above. Boundary witness, `sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go`, final state:

```
	if syncSkipped {
		// SPEC-UPDATE-MIGRATION-FIX-001 (REQ-UMF-001..003): the managed-surface
		// integrity probe runs on the version-match entry only. A user-cancelled
		// merge returns the same skipped=true, so the version predicate that
		// updateSkippedOnVersionMatch re-evaluates also gates the probe. The probe
		// runs before the strip below because it is pure observation and the strip
		// mutates configuration.
		if updateSkippedOnVersionMatch(cmd, ".") {
			runManagedSurfaceIntegrityProbe(out, ".")
		}
		// A version-matched update runs no sync and no merge, so the retired
		// per-agent model/effort keys are stripped here, after its own backup.
		// A user-cancelled merge returns the same skipped=true; the helper
		// re-evaluates the version predicate and leaves that case untouched.
		if err := stripRetiredModelConfigOnVersionMatch(cmd, out, "."); err != nil {
			updateLedger.requiref(sevWarn, "retired model-key removal failed: %v", err)
		}
		// Card t1527 D5 + repair round: the deferred render carries the block
		// on this early return too — no explicit call here.
		// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-004): a version-matched
		// update is still a finished plain template-sync run, so the ask runs
		// here too; its own gates (mode flags, terminal, CI, asked) decide
		// whether anything prompts. A failure warns; it never fails the update.
		if err := runParticipationStep(cmd, out); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
		}
		runParticipationFlushAtUpdate(cmd.ErrOrStderr())
		return nil
	}
```

**E4 — subagent boundary grep.** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli internal/userassets | grep -v _test.go | wc -l` → `113` before and after this run. This is the pre-existing baseline; no new match. The touched non-test files (`update.go`, `update_integrity_probe.go`, `install.go`) each contain 0 matches.

**E5 — lint and format.** `golangci-lint run ./internal/cli/... ./internal/userassets/...` → `0 issues.` (log `.moai/reports/t1578/lint-touched-final.log`, `LINT_FINAL_EXIT=0`). The pre-flight baseline `golangci-lint run --timeout=2m` → `0 issues.` before any edit. `gofmt -l` on the six touched Go files → no output, exit 0.

**B5 — spec-lint (CI tier).** `moai spec lint .moai/specs/SPEC-UPDATE-MIGRATION-FIX-001` → exit 0 with 0 errors and 5 `WARNING` rows, all `VacuousTestAssertion`: 2 in `plan.md` (lines 50 and 75) and 3 in `acceptance.md` (line 43 and two more). Each warning flags an unanchored `-run` selector that the SPEC itself names. None is in `spec.md` or `progress.md`. `plan.md` and `acceptance.md` are unchanged in this run (`git status` shows them clean), so these five are the plan-phase state carried forward. The raw log is `.moai/reports/t1578/spec-lint-m2.log`. This run did not capture a separate baseline at HEAD.

**E6 — branch state, no push.** HEAD `6bb0d8278` on `WT-update-migration-fixes` (`feat(SPEC-UPDATE-MIGRATION-FIX-001): M1 measurement record (card t1578)`). No M2, M3, or M4 commit existed at this E6 snapshot (blocked, see Gate). Lane update: M2 `b50dbbffb` and M3 `1a48f92da` are now committed (see §E.3). No push and no `gh pr`. Working tree: `M progress.md`, `M internal/cli/update.go`, `M internal/userassets/install.go`, and four untracked Go files (`internal/cli/update_integrity_probe.go`, `internal/cli/update_integrity_probe_test.go`, `internal/cli/update_managed_dir_guard_test.go`, `internal/userassets/installer_empty_target_test.go`).

**Scope proof.** Card base = `git merge-base HEAD main` = `2aab5f797`. `git diff --stat 2aab5f797..HEAD` covers the six plan-phase SPEC artifacts plus the M1 record, with no Go file. `git diff --stat` (working tree) gives `progress.md` (+65 −1), `internal/cli/update.go` (+9), and `internal/userassets/install.go` (33 lines changed). The anchor files are unchanged: `git diff --stat 2aab5f797..HEAD -- internal/cli/update_deny_migration.go internal/cli/update_deny_migration_test.go internal/template/deployer_mode.go` → empty. No template tree and no `.claude/rules` file is changed.

**E7 — blocker report.** See the Gate record and the Blocker section below.

**E8 — RED evidence.** M2 installer RED and M3 probe RED are verbatim above (M2-b(1), M3 RED). The M2 guard `TestTemplateSync_LeavesNoEmptyManagedSkillDirs` is green by design on this tree, because the healthy swept count is zero (M1-b). Its failure arm was observed through the planted-input sub-test and the mutation recorded under M2-b(3).

**Gate — verify rule (`moai gate`), FAIL, no commit made.**

- Installed build: `moai gate` (installed `moai-adk v3.2.0-rc.29`, build commit `4f8aba061`). That commit is neither an ancestor of HEAD nor a descendant of it (`git merge-base --is-ancestor` exits 1 in both directions), so the build is divergent. Log `.moai/reports/t1578/gate-m2-run.log`, `GATE_EXIT=1`. Verbatim: `[SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line`. Step results: `go vet: executed in 10.283s`, `typecheck: skipped — no default for this language`, `golangci-lint: executed in 24.893s`, `go test: not reached`, `Quality gate failed.`
- Tree build, to rule out the divergent installed build as the cause: `go build -o .moai/reports/t1578/moai-tree ./cmd/moai` at this worktree (HEAD `6bb0d8278` plus the working-tree change set). It reports `v3.1.3 none built unknown` because no release ldflags were passed. Invoked by path: `./.moai/reports/t1578/moai-tree gate`, log `.moai/reports/t1578/gate-m2-treebuild.log`, `GATE_EXIT=1`. The same finding appears at the same file and line. Steps: `go vet` executed in 3.049s, `golangci-lint` executed in 5.298s, `go test: not reached`.
- The flagged file is pre-existing and unmodified. `internal/astgrep/testdata/fixtures/go/suppressed.go` was last changed in `59f201e85`, and `git status --short internal/astgrep` is empty. Its header says "Scanning this file should return 0 findings". Line 14 holds `// @MX:REASON` above the `// ast-grep-ignore` marker on line 15, and the checker in both builds wants the reason adjacent on the following line. The fixture and the checker disagree. This is outside the scope of this SPEC.
- `.moai/config/sections/gate.yaml` sets `ast_grep_gate` to `warn_only_mode: true` and `block_on_error: false`, yet the gate exits 1 at this step. This is an observed inconsistency; I did not investigate its cause.
- Git hooks are disabled for this repository: `core.hooksPath` is `/dev/null` in the shared `/Users/goos/MoAI/moai-adk-go/.git/config`. No mechanical pre-commit hook blocks a commit here, but the verify rule still applies to this run.
- Decision taken: no commit for M2, M3, or M4, because the gate verdict is FAIL. I did not edit the fixture, the checker, or the gate configuration. Those belong to another owner and fall outside this SPEC's scope.

**Blocker (E7).** The following needs an operator or orchestrator decision before any commit:

1. The pre-existing ast-grep suppression mismatch (`internal/astgrep/testdata/fixtures/go/suppressed.go` against the suppression checker) must be resolved in its owning SPEC, or the operator must explicitly override the gate for this card. Either path is outside this run.
2. The full `internal/cli` package run cannot complete in this environment (factory test blocked on a git child, and codex-family failures). Full-package evidence must come from CI or from an accepted scoped substitute.
3. Once the gate is cleared, the M2, M3, and M4 change sets can be committed as separate Conventional commits that name card `t1578`, with `Authored-By-Agent: manager-develop` and the attribution line.

### M3–M4 status

M3 is implemented and verified in the working tree but not committed. M4 evidence is recorded above. Both wait on the gate decision.

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete (gate FAIL dispositioned by the leader ruling d-20261009T174939Z-a927: pre-existing base defect; the run stage is not blocked)
- run_complete_at: not set (no run-phase completion commit; the gate verdict is FAIL)
- run_commit_sha: 6bb0d8278 (M1 measurement record); M2 b50dbbffb; M3 1a48f92da; the M4 record commit carries this section (its SHA is reported to the leader, not written here)
- ac_pass_count: 5 of 5 (AC-UMF-001..005; see E1 in §E.2)
- ac_fail_count: 0
- l44_pre_commit_fetch: not performed (no push; the spawn forbids push)
- l44_post_push_fetch: not applicable (no push)
- new_warnings_or_lints_introduced: none (golangci-lint on the touched packages: 0 issues)
- cross_platform_build: go build ./... exit 0; GOOS=windows GOARCH=amd64 go build ./... exit 0
- gate_verdict: FAIL (ast-grep suppression finding in a pre-existing testdata fixture; go test not reached). Verdict unchanged; the verdict is the leader's.
- gate_finding_disposition: pre-existing base defect, not introduced by this card. Finding: `[SUPPRESSION_WITHOUT_REASON] ast-grep suppression at internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line`. Reproduced on base 2aab5f797 by running the same ast-grep gate leg on the base export: probe `TestBaseGateLegReproduction` (`go -C .moai/reports/t1578/base-export test ./internal/hook/quality/ -run '^TestBaseGateLegReproduction$' -count=1 -v`, exit 0, log `.moai/reports/t1578/gate-base-leg-probe.log`) logs `ok=false` and the same line. The probe config mirrors the CLI path (`internal/cli/gate.go:209-214`) from `gate.yaml` `ast_grep_gate` (enabled true, rules_dir `.moai/astgrep-rules`, block_on_error false, warn_only_mode true); the base export's `gate.yaml` is identical to the card tree's (`diff` exit 0).
- gate_finding_mechanism (read from code): the suppression check runs first and returns false at `internal/hook/quality/astgrep_gate.go:89`, before any WarnOnlyMode handling; it walks sources with `walkSourceFiles` (lines 77-81). The `/testdata/` path exclusion is declared for ast-grep scan findings (comment block above `astGrepExcludedPathPatterns`), and lines 77-89 do not apply it. The card does not touch `internal/astgrep` (`git diff 2aab5f797 -- internal/astgrep` is empty).
- gate_leg_reached_directly (the real run stopped before go test): `go -C <card> vet ./internal/cli/ ./internal/userassets/` exit 0 (`cond4-vet.log`, 0 bytes); `go -C <card> build ./...` exit 0 (`cond4-build.log`, 0 bytes); `GOOS=windows GOARCH=amd64 go -C <card> build ./...` exit 0 (`cond4-build-windows.log`, 0 bytes); scoped `go test ./internal/cli/ -run '^(TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries|TestRunUpdate_VersionMatch_RunsIntegrityProbe|TestRunUpdate_UserCancelled_SkipsIntegrityProbe|TestIntegrityProbe_FailOpen|TestIntegrityProbeEntry_DamageReasons|TestTemplateSync_LeavesNoEmptyManagedSkillDirs|TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir)$' -count=1 -v` exit 0, all seven `--- PASS` (`cond4-cli-scoped-test.log`); `go test ./internal/userassets/ -count=1 -v` exit 0, package `ok` (`cond4-userassets-test.log`). `TestTemplateSync_LeavesNoEmptyManagedSkillDirs` logs `swept=0` on the healthy tree; its red arm is `TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir` (PASS). Gap: the full `internal/cli` package suite was not run (scoped families only).
- lane_update (lane-2): M2 `b50dbbffb` (install.go, installer_empty_target_test.go, update_managed_dir_guard_test.go); M3 `1a48f92da` (update.go wiring, update_integrity_probe.go, update_integrity_probe_test.go); M4 is the record commit carrying this section. Nothing pushed; `.moai/config/sections/workflow.yaml` (leader-owned) not committed. The reason line above describes the state before M2 and M3 were committed; audit_ready was released by the lane after the leader ruling d-20261009T174939Z-a927 was verified on the decision board (17:55Z wake).
- full_package_internal_cli: not completed in this environment (see M2-c)
- preserve_list_post_run_count: the PRESERVE list is intact. The only edit to a PRESERVE file is the single insertion in internal/cli/update.go; the 9-form pin is unchanged; install.go changed only for REQ-UMF-006
- audit_ready: true
- reason: E6 is incomplete (no commit SHAs for M2–M4), E7 records an open blocker, and the gate verdict is FAIL. The signal is released after the operator's gate decision and the M2, M3, and M4 commits.

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Plan→run Kickoff decision record (autonomous transition, auto-semantics §9.1):

```text
decision record: decided_by=lane-27 orchestrator (card t1578) evidence_refs=.moai/reports/t1578/plan-audit-iter3.md (verdict PASS, overall 0.94 >= Tier M 0.80, blocking 0, convergence pass, codex pass, audited_sha 817b5b86f) + plan-audit-iter2.md (rcpt-5bb5f28c03f4602d0879cd46) ladder_path=autonomous-kickoff §9.1 (verdict PASS + score >= threshold + artifact-hash unchanged on the Go ComputeHash subject set + no blocker open)
```

Mode selection inputs: tier=M; scope≈6 files (probe + guard tests + install.go handler); domain count=1 (Go CLI internal/cli + internal/userassets); file language mix=Go + SPEC artifacts; concurrency benefit=LOW (coding-heavy); Agent Teams prereqs=not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic multi-file change, not a typo fix |
| serial | **yes** | coding-heavy Go implementation (Anthropic coding-task caveat) |
| fanout | no | research-heavy work only; single domain, single writer |
| sweep | no | not a mechanical-uniform ≥30-file transform |

Decision: **serial** (one manager-develop spawn, milestones M1→M4 in sequence).

Justification: the implementation is coding-heavy Go work in one subsystem family (update path + userassets installer); per Anthropic's coding-task parallelism caveat the sequential single-agent path is the safe default, and the write contract is one writer per tree (this worktree). sweep/fanout offer no concurrency benefit here; direct is below the semantic-change bar.

Boundary cases: none — all four mode criteria resolved unambiguously.


## §G Override and Refusal Record

- 2026-10-09T13:00:28Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:05:26Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:11:28Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:12:37Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T16:19:39Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
