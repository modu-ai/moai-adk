# SPEC-UPDATE-REPAIR-RC30-001 — Acceptance Criteria

## Budget and adoption status

- Tier L ceiling: 25 acceptance criteria. This file carries **26**: AC-RC30-01 to AC-RC30-26. That is one over the ceiling. Exceedance is a split signal (`spec-workflow.md` § SPEC Complexity Tier). No split is made in this revision (Gap: leader decision).
- U-22 (owned by card t1578) and U-30 (locale, blocked on decision Q1) carry **no** acceptance criterion in this SPEC. U-30 gets one after Q1 is decided.
- Adoption: under `verification-completeness.md` § 2, a release-blocking criterion is adopted only with a RED-now cell (command, verbatim stdout, exit code, tree SHA). **No criterion is adopted yet.** RED cells are captured in the first run milestone (M0) on tree `2aab5f797b75983e132af451da68f69e3426557b`. Until then every criterion is "pending RED capture" and no criterion counts as a pass.
- Each test named below is new. It is written at the milestone that owns the item (§ plan). A criterion is judged by the test's `-v` output: a `--- PASS` line for the named test with a count of at least 1. An empty swept set (`[no tests to run]`) is a failure, not a pass.
- Continued firing: a test that stops running shows as an absent `--- PASS` line in `-v` output. The CI job that runs `go test ./internal/cli/...` makes that absence visible because the named test is then missing from the run.

## Evidence ledger (observed in this run; not yet single-invocation RED cells)

These outputs were observed in this run on tree `2aab5f797`. Pipes were used, so they are not in adoptable RED-cell form. M0 recaptures each one as a single invocation.

```text
E-001  command: grep -n "func runUpdate(\|Current version\|func checkProjectMarker\|CLAUDE_CONFIG_DIR\|LookupEnv\|Getenv" internal/cli/update.go | head -30
       stdout:  140:func runUpdate(cmd *cobra.Command, _ []string) error {
                200:		_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
                765:	if os.Getenv(config.EnvSkipBinaryUpdate) == "1" {
                835:	// Card t1527 D1: the "Current version" line is dropped here — the top of
       tree:    2aab5f797b75983e132af451da68f69e3426557b

E-002  command: grep -n "func (m \*ConfigManager) Save\|DeploymentMode\|Fable\|TeamMode" internal/config/manager.go internal/config/types.go | head
       (re-read via: grep -n "DeploymentMode\|deployment_mode\|Alias\|Fable\|Opus\|Sonnet\|Haiku\|TeamMode" internal/config/types.go | head -20)
       stdout:  internal/config/types.go:284:	TeamMode string `yaml:"team_mode"`
                internal/config/types.go:383:	Fable  string `yaml:"fable,omitempty"`
                internal/config/types.go:389:// legacy backward-compat alias fields Opus/Sonnet/Haiku are removed. An
                internal/config/types.go:398:	Fable  string `yaml:"fable"`  // Fable tier (Claude Code ANTHROPIC_DEFAULT_FABLE_MODEL)
                internal/config/manager.go:225:func (m *ConfigManager) Save() error {
       tree:    2aab5f797b75983e132af451da68f69e3426557b

E-003  command: grep -n "ResolveRoots" -A 12 internal/userassets/paths.go | head -24
       stdout:  36:func ResolveRoots(home string) []Root {
                38:		{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")},
                39:		{Slug: RootClaudeAgents, Dir: filepath.Join(home, ".claude", "agents")},
                40:		{Slug: RootAgentsSkills, Dir: filepath.Join(home, ".agents", "skills")},
                41:		{Slug: RootCodexAgents, Dir: filepath.Join(home, ".codex", "agents")},
       tree:    2aab5f797b75983e132af451da68f69e3426557b

E-004  command: grep -n "customizations modified" internal/cli/update_archive.go | head -5
       stdout:  282:
                367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
       tree:    2aab5f797b75983e132af451da68f69e3426557b

E-005  command: grep -n "Save(" -c internal/config/manager.go   (not run; Gap)
```

Notes on E-002: the stdout is the `types.go` half of the grep. The `manager.go:225` line comes from a separate grep in this run. E-005 is listed only to mark a command that is planned and not yet run.

## Acceptance matrix

Columns: AC | Items | REQ | Runnable check (judge) | Expected observable result | Adoption status.

| AC | Items | REQ | Runnable check (judging test or command) | Expected observable result | Status |
|----|-------|-----|------------------------------------------|----------------------------|--------|
| AC-RC30-01 | U-01 | REQ-002 | Profile-visibility check: `CLAUDE_CONFIG_DIR=<profile-dir> <non-interactive Claude Code skill listing>` — the exact form is **a Gap**, measured in M0 before it is written here. | The listing contains at least one shared skill (for example `moai-plan`) that exists only under `~/.claude/skills` today. Until this passes, project-side copies are not deleted (REQ-RC30-003). | Pending RED capture; command form Gap |
| AC-RC30-02 | U-01 | REQ-001, REQ-003 | `go test ./internal/cli/ -run TestMigrationRemovalGate -count=1 -v` | `--- PASS: TestMigrationRemovalGate`. Project-side `.claude/skills/moai/**` is present after update while AC-RC30-01 has not passed. | Pending RED capture |
| AC-RC30-03 | U-02 | REQ-004 | `go test ./internal/cli/ -run TestMigrationBackupBeforeRemoval -count=1 -v` | `--- PASS`. Every removed file has a copy in the run backup before removal. Migration runs after the backup step. | Pending RED capture |
| AC-RC30-04 | U-03 | REQ-005 | `go test ./internal/cli/ -run TestBundleAnnounceForPluginMode -count=1 -v` | `--- PASS`. Output contains the exact `moai bundle add <name>` line for each optional pack not installed. | Pending RED capture |
| AC-RC30-05 | U-04 | REQ-006 | `go test ./internal/cli/ -run TestBundleList -count=1 -v` | `--- PASS`. Output names each catalog bundle with skill and agent counts. `moai bundle add --help` names the valid bundles. | Pending RED capture |
| AC-RC30-06 | U-05 | REQ-007 | `go test ./internal/config/ -run TestConfigSaveRoundTripLLM -count=1 -v` | `--- PASS`. A fixture `llm.yaml` with `deployment_mode`, an unmodelled key, and comments is unchanged by Save except for the intended field. | Pending RED capture |
| AC-RC30-07 | U-06 | REQ-008 | `go test ./internal/cli/ -run TestRecordlessArchiveWarning -count=1 -v` | `--- PASS`. A record-less project with a migration archive prints a visible warning and does not re-stamp `local` silently. | Pending RED capture |
| AC-RC30-08 | U-07 | REQ-009, REQ-010 | `go test ./internal/cli/ -run TestRestoreSettingsAndRetention -count=1 -v` | `--- PASS`. Restore applies the backed-up `settings.json`. Restore output lists uncovered paths. Retention keeps N directories per run. | Pending RED capture |
| AC-RC30-09 | U-08 | REQ-011, REQ-017 | `grep -n "0 user customizations modified" internal/cli/update_archive.go` | After the fix: no match, exit 1. Before the fix (E-004): one match at line 367, exit 0. | Pending RED capture (RED observed as E-004, pipe form) |
| AC-RC30-10 | U-09 | REQ-001 (scope) | `go test ./internal/cli/ -run TestNonCatalogSkillMessage -count=1 -v` | `--- PASS`. A managed skill with no catalog entry gets a distinct message, not "opt in via moai bundle add". | Pending RED capture |
| AC-RC30-11 | U-10 | REQ-001 (scope) | `go test ./internal/cli/ -run TestEmptiedSkillDirRemoved -count=1 -v` | `--- PASS`. A skill directory emptied by the migration is removed. | Pending RED capture |
| AC-RC30-12 | U-11 | REQ-012 | `go test ./internal/userassets/ -run TestUserRootDowngradeRefused -count=1 -v` | `--- PASS`. A binary older than the recorded user-root version refuses to refresh the user root. | Pending RED capture; blocked on t1578 and t1591 |
| AC-RC30-13 | U-12 | REQ-001 | `go test ./internal/cli/ -run TestKeptOptionalAssetRefresh -count=1 -v` | `--- PASS`. A kept optional agent or `.agents/skills` copy is refreshed in place or gets a notice. | Pending RED capture; blocked on t1591 |
| AC-RC30-14 | U-13 | — | `go test ./internal/cli/ -run TestSkipDecisionUsesTreeHash -count=1 -v` | `--- PASS`. A same-version run with a different template tree does not skip. | Pending RED capture |
| AC-RC30-15 | U-14 | — | `go test ./internal/cli/ -run TestStampSkewWarning -count=1 -v` | `--- PASS`. A stamp that disagrees with the deployed bytes produces a warning. The stamp `cutover/github-flow` is reported as invalid. | Pending RED capture |
| AC-RC30-16 | U-15 | — | `go test ./internal/cli/ -run TestPermissionsAskPreserved -count=1 -v` (test to be written first, from the fixture in U-15) | `--- PASS` after the cause is established. Before that, the test is the RED cell. | Gap: cause not established |
| AC-RC30-17 | U-16 | REQ-015 | `go test ./internal/cli/update/report/ -run TestPreservedLinesCountOnly -count=1 -v` | `--- PASS`. Default output has one count line. `--verbose` lists the paths. The `--verbose` meaning of SPEC-V3R6-UPDATE-NOISE-001 (REQ-UN-005, REQ-UN-010) is unchanged. | Pending RED capture |
| AC-RC30-18 | U-17 | REQ-013 | `go test ./internal/cli/ -run TestUpdateSingleBanner -count=1 -v` | `--- PASS`. Exactly one version line appears on a run. The band is the one decided by REQ-TUXIU-015. | Pending RED capture; blocked on t1578 |
| AC-RC30-19 | U-18 | REQ-011 | `go test ./internal/cli/ -run TestPreviewCountEqualsRunCount -count=1 -v` | `--- PASS`. The preview count equals the run count for the same fixture (266 against 423 is the RED case). | Pending RED capture |
| AC-RC30-20 | U-19, U-20 | — | Investigation, recorded in research.md: `grep -rln "<moai skill path>" internal/template/templates` and `ls docs-site/content/ko/cli-reference/update.md` | Counts and the docs path are recorded with their exact commands and output. No count is quoted without its command. | GAP-INVESTIGATE |
| AC-RC30-21 | U-21 | REQ-002 (R1 profile) | `go test ./internal/cli/ -run TestProfileSyncResolution -count=1 -v` | `--- PASS`. Profile sync resolves the project ledger, not the environment, so the result does not depend on where update runs. | Pending RED capture; blocked on t1578 |
| AC-RC30-22 | U-23 | — | `go test ./internal/config/ -run TestGateYAMLCommentsSurviveRestore -count=1 -v` (fixture: 12 comment lines) | `--- PASS`. All comment lines survive backup, wipe, deploy, and restore. The test is RED first; the cause is then named in research.md. | GAP-INVESTIGATE |
| AC-RC30-23 | U-24 | — | `go test ./internal/cli/ -run TestLLMYamlRemarshalCallers -count=1 -v` (one case per `ConfigManager.Save` caller) | `--- PASS` for each caller. Each caller that drops a key is named. The `team_mode` reset at `launcher.go:397` is reported as a caller result. | GAP-INVESTIGATE |
| AC-RC30-24 | U-25 | — | Investigation: `config-change-audit.log` at `2026-10-09T02:26Z` and operator shell history, recorded in research.md with exact commands and output | The untraced removal is either attributed to a command with evidence, or recorded as unresolved in research.md with the evidence still missing. | GAP-INVESTIGATE |
| AC-RC30-25 | U-26, U-27, U-28, U-29 | REQ-014, REQ-016, REQ-017, REQ-018, REQ-020 | `make tui-snapshot` (target existence Gap, to be confirmed in M0), then `go test ./internal/cli/ -run 'TestUpdateOutputGolden|TestUpdatePlainFallback' -count=1 -v` and `NO_COLOR=1 go test ./internal/cli/ -run TestUpdatePlainFallback -count=1 -v` | `--- PASS` for each. TTY golden shows N of M, bar, one ✓/!/✗ line per step, a one-screen summary, and conflicts above the summary. Plain fallback carries the same counts. | Pending RED capture |
| AC-RC30-26 | U-31 | REQ-021 | `go test ./internal/cli/ -run TestIntegrityProbeCodexOnlyNoClaudeSettings -count=1 -v` (codex-only fixture with `hideClaude=true` first, RED; then the harness-conditioned fix, GREEN; a Claude fixture with `.claude/settings.json` deleted must still report "missing") | RED before the fix: `--- FAIL` with a `.claude/settings.json` "missing" finding on the codex-only fixture. GREEN after the fix: `--- PASS`; the Claude sub-case still reports "missing". The probe is on branch WT-update-migration-fixes only, so the RED cell is captured after M1 absorption. | Pending RED capture (after M1 absorption) |

## Mutant probes (verification-completeness § 2)

- AC-RC30-09 (grep for the literal): a relocated literal in another file passes a single-file grep. The judge must be `grep -rn` over `internal/cli` (repo-wide), not one file.
- AC-RC30-18 (one banner): a mutant that removes the identity band and keeps `update.go:200` passes "exactly one line". The judge must also assert the band is present (REQ-TUXIU-015).
- AC-RC30-25 (golden): a mutant that regenerates the golden from the current output passes. Golden updates must be made only through `make tui-snapshot` on a reviewed diff.

## Definition of Done (for the run phase)

- Every criterion has a RED-now cell captured at M0 on tree `2aab5f797` (or the then-current tree after absorption, with the new tip recorded in progress.md § G).
- Every criterion passes at its owning milestone, with the `-v` output recorded in progress.md.
- The coverage target of 85% or more per touched package is measured by `go test -cover` on the packages named in plan.md (measurement Gap until M0).
