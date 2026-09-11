# t661 — code reading (no test executed)

Tree: `.claude/worktrees/t661`, branch `WT-update-test-home`, base merge `21ee20493`
(HEAD^1 `987eb7e40` origin/develop, HEAD^2 `1bea05daa` local develop). Base merge attempt 1 failed with
`fatal: Unable to write index.` (exit 128; HEAD unchanged, MERGE_HEAD absent, index.lock absent, no
manual deletion); attempt 2 succeeded. Nothing was executed; every claim below is a code-path reading
with file:line on this tree.

## Package-level isolation that exists
- `internal/cli/main_test.go:213` TestMain sandboxes ONLY `profile.BaseDirOverride` (+ a cwd `.moai`
  residue guard). It does NOT set HOME, MOAI_HOME, or `userHomeDirFn`.
- `internal/cli/glm_tools.go:124` `userHomeDirFn = userHomeDir` → `homedir.go` → `paths.Home()`
  (`internal/paths/paths.go:53`): `$HOME` if non-empty, else `os.UserHomeDir()` — the real home under
  `go test` unless a test overrides it.
- The two suspect test files contain no `Setenv("HOME")`, `userHomeDirFn`, `MOAI_HOME`, or
  `BaseDirOverride` (grep exit 1).

## Home-writing sink
`internal/cli/update.go:958` `ensureGlobalSettingsEnv()`: `userHomeDirFn()` → `os.RemoveAll` of
`~/.claude/hooks/moai` when present, then reads `~/.claude/settings.json` and rewrites it when cleanup
keys match. Its update-path caller is `internal/cli/update_template_sync.go:567`, the tail of
`runTemplateSyncWithReporter`, reached only if no earlier `return` fires (the step loop at :395-537
returns on a step error; the version-match check returns nil early).

## Per test
| Test | Path | Reaches :567 (real ~/.claude)? |
|---|---|---|
| TestReproduction_NonProjectDirectoryPollution_Issue1086 (`update_clean_install_test.go:724`) | `--yes`; fixture has no `system.yaml` → `checkProjectMarker` (`update_restore.go:21`, called from `runUpdate` before the update lock) returns an error before v2 detection and sync | **No** (code) |
| TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently (`:431`) | `--yes`; `system.yaml v3.0.0-rc2` + `.agency/` → gate passes → `runV3ResidueCleanup` → `runAgencyMigrationAdapter` (`update.go:885`, `paths.Home()`) → `runTemplateSyncWithProgress` (version mismatch, autoConfirm) → `runTemplateSyncWithReporter` | **Structurally yes, no guard**; actual reach depends on every sync step succeeding on this minimal fixture — not decidable by reading |
| TestRunUpdate_ThreeRunIdempotency_V3Project (`:824`) | same path, three runs | same as above, up to 3 times |
| TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive (`update_skip_sync_test.go`) | calls `runTemplateSyncWithProgress` directly with `--force`, `yes=false` → `confirmViaPreview` (`update_template_sync.go:694`): non-TTY → error, returns before the reporter; TTY → interactive preview | **No** under non-TTY; only after a human confirms in a TTY |

## Other candidate sinks checked
- Agency migration home use: only `checkpointPath` (`migrate_agency.go:188-193`), written only from the
  signal handler (`migrate_agency_signal.go:31`) → a normal run writes nothing under `~/.moai`.
- Shell config append (`internal/core/project/initializer.go:333-347`, `configureShellEnv`): only via
  `project.NewInitializer` in `init.go:832`; update reaches shell config only through `--shell-env`
  (`update.go:205`). Not reachable from these four tests.
- `profile.GetBaseDir()` (`profile.go:55`, `os.UserHomeDir()`): covered by the TestMain sandbox.
- `update_template_sync.go:269` and `:329`, `update_clean_install.go:443`: `userHomeDirFn` feeds
  template rendering inputs only (read).

## Sibling census (unverified reach — candidates only)
Test files that drive `updateCmd.RunE` / `runTemplateSyncWith*` with zero `Setenv("HOME")` /
`userHomeDirFn =` lines: `update_mode_test.go`, `update_hooks_guidance_test.go`, `coverage_test.go`,
`update_mirror_heal_test.go`, `update_deny_migration_test.go`, `update_llm_preserve_test.go`,
`integration_test.go` (plus the two card files). Existing repair model: `homeSeamSpy`
(`update_home_seam_test.go:39`).

## Gaps
- Whether each sync step succeeds on the two v3 fixtures (hence whether :567 runs) needs an execution —
  deferred until a seam is injected, per the lead's instruction.
- Whether `go test` hands the test binary a TTY stdin on this machine was not established.
