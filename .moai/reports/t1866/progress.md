# t1866 progress record

Card t1866 (Class B, operator-directed, no SPEC): shipped defaults changed to the operator-confirmed combination. Tree `.moai/worktrees/t1866`, base `ae34f2683`.

## Decisions (leader rulings recorded in this session, 2026-10-18)

- CLAUDE settings are the profile store's launch keys (`permission_mode`, `model`, `effort_level`), not the deployed `settings.json.tmpl`. The template is untouched.
- `model` default is `sonnet[1m]`, the value the web picker stores for the label "Sonnet 5.5". The card text's `claude-sonnet-5-5` is not a picker value.
- `permission_mode`: operator chose "auto for all". The wizard pre-selects `auto`; a profile with no stored mode launches Claude sessions in `auto`; a stored value (`acceptEdits` included) is kept; the autonomy-tier mapping (`semi-auto` → `acceptEdits`) is unchanged; no migration write.
- Carve-out (leader approved after the lane's objection): the GLM backend keeps the empty mode, and a stored or settings-chosen `auto` is dropped there with a one-line note. Reason: `internal/cli/glm.go` refuses `--permission-mode auto`, and `launcher.go` appended the flag whatever the backend.
- **Permission posture change, stated as the operator asked:** a profile that never chose a permission mode moves from "follow the project `settings.json` `defaultMode`" (the template ships `"default"`) to `auto`, on the Claude backend only. The `.claude/settings.local.json` override then shadows a `defaultMode` edited by hand in the project `settings.json`.

## Claim

The five default changes are in the tree and each is pinned by a test observed red before the change and green after it; the existing update-preservation tests still pass; the full-package suites for the touched packages have NOT run (see Gaps).

## Evidence (commands run in this tree; outputs observed in this session)

RED observed before each change:

```
go test ./internal/config -run 'TestDefaultGLM|TestSessionDefaults|...'
  defaults_test.go:408: DefaultGLMEffortLow: got "low", want "max"
  defaults_test.go:408: DefaultGLMEffortFable: got "max", want "high"
go test ./internal/core/project ./internal/cli/wizard -run '...report...'
  initializer_test.go:312: report.format = "html+md", want "artifact" (default)
  questions_test.go:41: report_format default = "html+md", want "artifact"
  wizard_test.go:440: report_format default should be 'artifact', got "html+md"
go test ./internal/settings -run '^TestNormalizePermissionMode$'
  4 lines: normalize(auto) = "auto", want empty; normalize(acceptEdits) = "", want acceptEdits (x2 each)
go test ./internal/cli -run '^TestInitialProfileResult_'
  PermissionMode = "acceptEdits", want auto / Model = "", want sonnet[1m] / EffortLevel = "", want medium
go test ./internal/cli -run '^TestLaunch_(Unset|Stored)'          (before the launcher change)
  TestLaunch_UnsetPermissionModeDefaultsToAutoOnClaudeBackend: --permission-mode = "" in args [claude --model opus[1m]], want auto
  TestLaunch_StoredAutoIsNotForwardedOnGLMBackend: --permission-mode = "auto" in args [claude --permission-mode auto], want none
```

GREEN observed after:

```
go test ./internal/cli -run '^(TestLaunch_(Unset|Stored)|TestLaunchClaudeDefault_ForwardsPermissionModes|TestSyncPermissionModeToSettingsLocal|TestContainsPermissionMode|TestCharacterize_GLM_ContainsPermissionMode)' -count=1 -v
  15 x "--- PASS", 0 x "--- FAIL"; ok  github.com/modu-ai/moai-adk/internal/cli  1.654s
go test ./internal/settings -count=1                  ok  0.933s
go test ./internal/cli/wizard -count=1                ok  6.180s
go test ./internal/template -run 'GLM|Glm|Effort|TierDefault|ReportYAML|Report|SectionsSymmetry|llm' -count=1
  28 tests selected (go test -list); ok  15.217s
go test ./internal/cli -run '^(TestUpdateLLMYAML|TestUpdateGateYAMLPreservesUserValues)' -count=1 -v
  5 x "--- PASS" (the update cycle driven with the real embedded llm.yaml)
```

The four launcher cells are: unset profile + Claude backend → `--permission-mode auto` and `permissions.defaultMode: auto`; unset profile + GLM backend → neither; stored `auto` + Claude → forwarded; stored `auto` + GLM → neither (the cell that was red). The GLM-unset cell passed before and after the change: it is a guard, not a RED.

`gofmt -l` over the touched directories printed nothing; `go vet` over `internal/config`, `internal/settings`, `internal/core/project`, `internal/cli/wizard`, `internal/template` printed nothing.

## Follow-up evidence (leader's request for the init, update and web gaps)

```
go test ./internal/cli -run '^(TestInit_FreshProjectCarriesTheShippedDefaults|TestUpdate_KeepsUserValuesForTheDefaultsT1866Changed)$' -count=1 -v
--- PASS: TestInit_FreshProjectCarriesTheShippedDefaults (1.63s)
--- PASS: TestUpdate_KeepsUserValuesForTheDefaultsT1866Changed (0.51s)
ok  	github.com/modu-ai/moai-adk/internal/cli  3.465s
```

- `TestInit_FreshProjectCarriesTheShippedDefaults` runs the real init command function in-process through `quietWizardRunInit`, which installs `prepareSafeInitHome`: it redirects the home seams under a temp dir and fingerprints the REAL home before and after, failing the test if the real user folders changed. It asserts `report.yaml` `format: artifact` and, through the config loader, `llm.glm.effort` high/medium/low/fable = max/high/max/high. The wizard's own defaults are asserted by `TestReportFormatQuestion`, `TestReportFormatInWizard` and `TestInitialProfileResult_ShippedLaunchDefaults`.
- Sensitivity check: with the template `llm.yaml` `low` temporarily put back to `"low"`, the init test failed with `llm.glm.effort.low = "low", want "max"`; the template was restored (`git diff` on it empty afterwards).
- `TestUpdate_KeepsUserValuesForTheDefaultsT1866Changed` drives the real update cycle (`runTemplateSyncAt`) over a project whose `llm.yaml` is the real embedded template with `effort.low: low` and `effort.fable: max` (the pre-t1866 defaults, now the user's values; `replaceOnce` requires the new defaults to be present first, so the fixture differs from the template) plus `report.yaml` `format: md`, and a profile storing `permission_mode: acceptEdits` in an isolated store. After the update all four values are the user's. The profile store is outside what `moai update` rewrites, so that last assertion pins non-interference rather than a merge. Not mutation-checked: an update that overwrote user values was not constructed.
- Web description: the English, Korean, Japanese and Chinese `sec.report.desc` strings in `i18n.js` already named `artifact`; the stale text was the Go fallback in `internal/web/schemaform.go` ("html+md or md"), now "html+md, md, or artifact". `go test ./internal/web -run 'I18n|i18n|Governance|Baseline|Untranslated|Report|SchemaForm|Schemaform'` passed (31 i18n/baseline tests selected plus the report/schema-form ones).

## What the wizard saves (source read, not run end to end)

`runProfileSetup` writes `PermissionMode: permissionMode` (`profile_setup.go`, the preferences literal) from the wizard result, so a user who accepts the pre-selected `auto` stores `"auto"` explicitly; the same holds for the pre-selected `sonnet[1m]` and `medium`. The launcher cells above cover what such a stored `auto` does at launch.

## Corrections made during the card (recorded so they are not re-inferred)

- A defect candidate raised earlier in the card — "the web console normalizes `acceptEdits` to empty, so a user who picks it launches in default mode" — does not exist. No non-test caller applies `Persist.Normalize` for `permission_mode` (every `.Normalize` call site in the tree was read), and `TestCharacterize_MainSessionEffortSaveWritesPreferencesOnly` already pins that the web persists `acceptEdits` as it is. Withdrawn. Only the schema function's unit contract changed (`auto` → empty, `acceptEdits` kept); it is not wired into the web save path (leader ruling).
- `TestPermissionModeNormalizeAcceptEdits` carried a source-grep guard that claimed the TUI normalizes `acceptEdits` to empty. The TUI never did; the guard matched a string used for the confirmation line. It was replaced by `TestPermissionModeNormalizeDefaultAuto`, which checks the schema function only.
- My first test for the deployed `llm.yaml` used an initializer constructed without a template deployer; that initializer never writes `llm.yaml`, so the test was wrong and was removed. The deployed file's content is covered by `TestGLMTierDefaultsTemplateMatchesGoDefaults` (embedded file equals the Go constants) plus the literal assertions on the constants.

## Gaps

- Whole-package runs observed passing: `internal/settings` and `internal/cli/wizard` only. Not run whole: `internal/config`, `internal/core/project` (its whole-package result was not observed; only the named report tests and the init test), `internal/template` and `internal/cli`, and no landing re-measure has run. Package-wide suites need the `go-test-heavy` slot; the lane has not held it for this card. `internal/cli` ran only the tests named above.
- (Closed in the follow-up commit, see "Follow-up evidence" below: the in-process init run, the update-preservation test and the web description.) Open remainder: the en, ja and zh docs-site trees were not searched; `docs-site/content/ko` carries no `report.format` text.
- The web console's empty-option labels for `model` and `effort_level` ("(project default)", "(model policy, else Claude Code default ...)") still describe an unset profile; only the TUI wizard pre-selection changed for those two keys.
- The `internal/web` suite ran only for the tests named `Permission|acceptEdits|AcceptEdits|permission_mode|MainSession|Normaliz` (3 tests).
- The built binary was not run, so the launcher change was observed through the exec seam in tests, not in a real launch.

## Residual-risk

- Users who already have `permission_mode` unset now start Claude sessions in `auto`; where Claude Code refuses `auto` for the account or model, the first launch after the change is where that would show.
- A hand-edited project `settings.json` `defaultMode` is shadowed for those users by the `settings.local.json` override.
- A `--permission-mode auto` typed on the GLM command line keeps its existing refusal; a stored `auto` on GLM is now dropped silently except for one stderr line.

## Audit repair round 1

Source: `.moai/reports/t1866/sync-audit.md` (FAIL, audited_sha 2eb7daee9b). Repair cycle: each fix test-first; RED then GREEN observed in this tree. New tests: `internal/cli/launcher_permission_resume_test.go`; code: `internal/cli/launcher.go`.

| Finding | Command and observed output |
|---|---|
| F1 resume keeps stale `auto` under GLM | RED: `go test ./internal/cli -run '^TestLaunch_(GLMRemovesStaleAutoFromSettingsLocal|...)$' -count=1` -> `--- FAIL: TestLaunch_GLMRemovesStaleAutoFromSettingsLocal/GLM_launch_with_--continue`, `/GLM_launch_with_-c`, `/GLM_launch_with_DO_CLAUDE_CONTINUE=true`: `settings.local.json still carries permissions.defaultMode auto after a GLM launch`. GREEN: all four subtests PASS (the fresh-GLM subtest passed in RED too, a guard). Fix: `dropAutoDefaultModeFromSettingsLocal`, called when `cont && permGLM`; removes only a value equal to `auto`, no rewrite otherwise. `TestLaunch_GLMResumeKeepsOtherDefaultModes` (plan stays) PASS. |
| F2 failed team_mode save loses the GLM decision | RED: `--- FAIL: TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto`: `--permission-mode auto reached a GLM launch whose team_mode save failed: [claude --permission-mode auto --model opus[1m]]` plus `settings.local.json carries defaultMode auto`, with stderr `Warning: failed to persist team mode: yamlpatch: create temp: ... permission denied`. Failure inducement: the sections directory made read-only (`chmod 0555`, llm.yaml absent) as a non-root user, asserted by a precondition (`persistTeamMode` must fail; the test fails, not skips, if it does not). GREEN: PASS. Fix: `permGLM := glmBackend || os.Getenv(config.EnvMoaiLaunchProvider) == "glm"` (`launchClaudeForProvider` already sets that variable to the command's mode); the user-visible warning is unchanged. `permGLM` governs the permission decision only; model and effort selection still use the disk-resolved `glmBackend` as before. |
| F3 uncovered paths | `TestLaunch_LegacyBypassWithEmptyProfileModeKeepsBypassAndWritesNoOverride`, `TestLaunch_ExplicitPermissionModeAutoOnCommandLineIsForwardedOnClaude`, `TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal` (plan and auto hand-edits): all PASS on the unchanged launcher (characterization guards, not REDs). |
| F4 false comments | Four comments edited (`initializer.go` x2, `profile_setup.go` x2), comment-only; `gofmt -l` empty and the package tests below pass. |
| F5 settings-chosen mode persisted `auto` | RED: `--- FAIL: TestLaunch_SettingsChosenModeDoesNotPersistDefaultAuto`: `permissions.defaultMode = "auto", want the override removed`. GREEN: PASS. Fix: `syncPermMode` is not defaulted when `settings["DO_CLAUDE_PERMISSION_MODE"] != ""`. |

GREEN after all edits: `go test ./internal/cli -run '^(TestLaunch_|TestLaunchClaudeDefault_ForwardsPermissionModes|TestSyncPermissionModeToSettingsLocal|TestContainsPermissionMode|TestCharacterize_GLM_ContainsPermissionMode|TestUnifiedLaunch_GLM)' -count=1 -v` -> every `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  2.612s`; `go test ./internal/cli -run 'TestLaunch_|TestInitialProfileResult|TestShippedDefaults|TestProfileSetup|TestUnifiedLaunch_|TestLaunchClaude' -count=1` -> `ok  .../internal/cli  1.948s`; `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../settings 0.679s`, `ok .../wizard 5.003s`, `ok .../core/project 2.321s`. `gofmt -l internal/cli internal/core/project internal/settings` printed nothing; `go vet ./internal/cli/` filtered to `launcher|profile_setup` printed nothing. CHANGELOG t1866 entry gained one sentence: a GLM launch also removes a launcher-written `auto` from `settings.local.json` on resume.

F6 (cosmetic profile summary and GLM note) was out of scope and is untouched.

Gaps remaining: no whole-package run of `internal/cli`, `internal/config`, `internal/template`, `internal/web` (heavy slot not held); no landing re-measure; the en/ja/zh docs-site trees not searched; the built binary was not run and the launch was observed through the exec seam; what Claude Code does when it reads `defaultMode: auto` in a GLM-env session was not observed; the F2 failure was induced by a read-only directory, not by a real unwritable user config; the F1 tests assert the settings file and argv, not a live Claude process; the audit's delta re-audit has not run.

Note: the first fix spawn was refused with AUDIT_RECEIPT_VIOLATION and the leader removed that guard record after backing it up (board 6a81, backup directory `.moai/reports/t1866/backup-rejection-20261018T054836Z/`).

Deviation: the run-phase implementation of this card was written directly by the lane session rather than by manager-develop.

## Audit repair round 2

Source: `.moai/reports/t1866/sync-audit-iter2.md` (FAIL, audited_sha bece378100; findings F7, F8). One commit; F6 stays out of scope.

| Finding | Command and observed output |
|---|---|
| F7 failed stale-`auto` removal on a GLM resume only warned | RED: `go test ./internal/cli -run '^TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved$' -count=1` -> stderr `Warning: failed to clear stale auto permission mode: open settings lock file: open .../proj/.claude/settings.local.json.lock: permission denied`, `--- FAIL: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved`, `launcher_permission_resume_test.go:352: GLM resume returned nil although the stale auto could not be removed`. GREEN: same test PASS; `go test ./internal/cli -run '^TestLaunch_' -count=1 -v` -> all `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  1.977s`. Fix: in `runLaunchClaude`, when `cont && permGLM && settingsLocalPath != ""` a failed `dropAutoDefaultModeFromSettingsLocal` returns `GLM resume: the stale permissions.defaultMode auto in <path> could not be removed: <cause>`; the session is not started. Claude-backend resume and the fresh-launch sync stay warn-only/unchanged. Induction: `.claude` made `chmod 0555` as uid 501 (`id -u` -> `501`); a probe file creation in that directory must fail, otherwise the test fails (never skips). The test asserts the error names the file and says `auto`, `execOrSpawnClaudeFunc` was not called, and the file is byte-unchanged; permissions restored in cleanup. |
| F8 "launcher-written" wording | Comment above the drop call in `launcher.go` and the t1866 CHANGELOG sentence now say a `permissions.defaultMode: auto` is removed on a GLM resume; the CHANGELOG sentence also says a failed removal stops the GLM launch with an error. Text only. |

GREEN tail: `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../settings 0.940s`, `ok .../wizard 6.802s`, `ok .../core/project 3.706s`; `gofmt -l internal/cli internal/core/project internal/settings` printed nothing.

Gaps remaining: no whole-package run of `internal/cli`, `internal/config`, `internal/template`, `internal/web`; the failure was induced by a read-only directory (lock-file creation refused), not by a real unwritable user config; the built binary was not run; the delta re-audit has not run; F6 untouched.

Note: the second fix spawn was refused with AUDIT_RECEIPT_VIOLATION (unknown-spec) and the leader removed that guard record after backing it up (board 6a81, backup directory `.moai/reports/t1866/backup-rejection-20261018T060040Z/`).

## Audit repair round 3 (test-only exception)

Source: `.moai/reports/t1866/sync-audit-iter3.md` (FAIL, audited_sha 02bb6804fb; findings F9 blocking via the required codex gate, F10 optional). One commit; test file, this record and one CHANGELOG sentence only; no product code (`git diff -- internal/cli/launcher.go` empty).

| Finding | Command and observed output |
|---|---|
| F9 "no exec" assertion watched the wrong seam | A `--continue` resume starts its first child with `exec.Command(...).Run()`; `execOrSpawnClaudeFunc` is reached only after a resume exit-1 fallback. `TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved` now pins a stub claude (`pinMarkerClaude`: `touch <marker>` then `exit 0`, marker in its own temp dir, not in the read-only `.claude`) and asserts (1) the launch error names the file and `auto`, (2) the marker does not exist, (3) the exec seam was not called (kept), (4) the settings file is byte-unchanged. The assertions are non-fatal so each one reports. RED (working-tree-only mutation: the removal failure made warn-only, reverted afterwards): `go test ./internal/cli -run '^TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved$' -count=1 -v` -> `launcher_permission_resume_test.go:357: GLM resume returned nil although the stale auto could not be removed`, `launcher_permission_resume_test.go:362: the session was started (the claude binary ran) with a stale auto still on disk`, `--- FAIL: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved (0.23s)`, `FAIL github.com/modu-ai/moai-adk/internal/cli 1.671s`. GREEN after `git checkout -- internal/cli/launcher.go`: same test PASS; `go test ./internal/cli -run '^TestLaunch_' -count=1 -v` -> 13 top-level `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  2.293s`. Positive control `TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary` (subtests `defaultMode plan` and `no settings.local.json`): the resume returns nil and the marker EXISTS, so the stub is the child and the marker mechanism observes a `--continue` start. |
| F10 malformed settings.local.json stops a GLM resume | Recorded, not a code change. `dropAutoDefaultModeFromSettingsLocal` returns the `readSettingsMap` error, so a malformed `.claude/settings.local.json` now stops a GLM resume with an error naming the file; before the card the launch continued with a warning. One sentence added at the end of the t1866 CHANGELOG entry. |

Test-design note: the exit-0 marker stub follows the leader's instruction. The auditor suggested an exit-1 stub for the failing path; with the exit-0 stub the mutated launcher returns nil at the first child start, which both the marker and the error assertion catch (RED above), so ordering is covered without the exit-1 variant.

Gaps remaining: no whole-package run of `internal/cli`, `internal/config`, `internal/template`, `internal/web`; the mutation was a warn-only removal failure, not the move-after-child-start overlay codex described (both reach the child start before returning); the malformed-settings stop (F10) is documented, not covered by a test; the built binary was not run; the final delta re-audit has not run; F6 untouched.

Honest line: the round-2 repair instruction told the implementer to assert the exec seam for a resume, although a resume's first child start is `tryCmd.Run()`; that instruction was the lane's mistake, found by the round-3 audit.

Note: the third and fourth fix spawns were refused with AUDIT_RECEIPT_VIOLATION (unknown-spec) and the leader removed those guard records after backing them up (board 6a81; backup directories `backup-rejection-20261018T060040Z` and `backup-rejection-20261018T061236Z`).

## Landing re-measure finding: stale web test

The landing re-measure ran the whole `internal/web` package and found `TestGLMEffortTierDefaults` (`internal/web/glm_tier_test.go`) failing. Verbatim failure lines:

```
glm_tier_test.go:150: llm.glm.effort.low Default = "max", want "low"
glm_tier_test.go:150: llm.glm.effort.fable Default = "high", want "max"
glm_tier_test.go:162: llm.glm.effort.low: rendered page does not preselect the default "low" (t1278: radio checked marker)
```

Cause: the test still pinned the previous shipped defaults (low=low, fable=max). The card changed them to the operator-confirmed combination (`config.DefaultGLMEffortLow = "max"`, `config.DefaultGLMEffortFable = "high"`); the product behavior is right, the test expectation was stale.

Fix (test only): in `wantDefault`, low is now `template.GLMStateMax` and fable `template.GLMStateHigh` (assertions replaced, none removed or loosened); the doc comment now states high Max, medium High, low Max, fable High, notes glm-5.3-flash accepts max only so low=Max matches the wire. A `grep -rn` over `*_test.go` in `internal/web`, `internal/template`, `internal/settings` found no other test pinning the old low/fable effort defaults (template tests cover the reasoning-state overlay mapping, not the defaults; `glm_tier_defaults_test.go` compares against the Go constants).

Commands and observed output (each prefixed with the `unset MOAI_KANBAN_*` trio):
- RED: `go test ./internal/web -run '^TestGLMEffortTierDefaults$' -count=1` -> the three lines above, `FAIL github.com/modu-ai/moai-adk/internal/web 1.054s`.
- GREEN: same command -> `ok  github.com/modu-ai/moai-adk/internal/web 0.880s`.
- Whole package: `go test ./internal/web -p 1 -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/web 45.186s`.

Honest line: the lane's earlier scope check ran `internal/web` only by name pattern (3 to 31 tests), which is why the stale test was missed until the landing re-measure ran the whole package.

Note: the fifth fix spawn was refused with AUDIT_RECEIPT_VIOLATION (unknown-spec) and the leader removed that guard record after backing it up (board 6a81, board entry d-20261018T103229Z-1628, backup directory `backup-rejection-20261018T103222Z`).

## Audit repair round 4 (operator exception b18d: auto default for new sessions only)

Finding F14 (`.moai/reports/t1866/sync-audit-iter6.md`, P1 per codex, `internal/cli/launcher.go:670`): with an unset profile permission mode, `--continue` still appended `--permission-mode auto` to the child argv, so a hand-edited `permissions.defaultMode: plan` in `settings.local.json` stayed on disk but was overridden at run time. Operator decision (board d-20261018T104449Z-b18d): the `auto` default for a profile that never chose a mode applies to new sessions only.

Change (`internal/cli/launcher.go`, `runLaunchClaude`): the default application moved from before the argv consumption to after it, and became `if permMode == "" && !cont && !consumed.PermissionModeSet { permMode = launchDefaultPermissionMode(permGLM) }`, where `cont` is the final resume flag (settings `DO_CLAUDE_CONTINUE` or a typed `-c`/`--continue`). Unchanged: the GLM drop of a stored/settings-chosen `auto` (still before the typed-argv consumption), a typed `--permission-mode` winning, the `DO_CLAUDE_BYPASS` legacy path, `syncPermMode`, the resume drop of a stale `auto` on GLM, the failed-removal error. A stored profile mode (including `auto`) is still forwarded on a resume. CHANGELOG t1866 entry reworded to match. The extra `!consumed.PermissionModeSet` term keeps a typed (even empty) `--permission-mode` from being replaced by the default, as before.

Tests (all start the `--continue` child through `tryCmd.Run()` on a pinned stub claude that records its args to a file in its own temp dir, helper `pinArgsRecordingClaude`; they do not use the exec seam). RED run, `go test ./internal/cli -run '^(TestLaunch_ClaudeResume)' -count=1 -v`, before the fix:

```
    launcher_permission_resume_test.go:194: child args [--permission-mode auto --continue --model opus[1m]] carry --permission-mode on a resume with an unset profile, want none
--- FAIL: TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal (0.45s)
    --- FAIL: TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal/plan (0.23s)
    --- FAIL: TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal/auto (0.23s)
--- FAIL: TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag (0.74s)
    --- FAIL: TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag/--continue (0.25s)
    --- FAIL: TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag/-c (0.25s)
    --- FAIL: TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag/DO_CLAUDE_CONTINUE=true (0.23s)
    launcher_permission_resume_test.go:293: child args [--permission-mode auto --continue --model opus[1m]] carry --permission-mode, which would override the hand-written plan
--- FAIL: TestLaunch_ClaudeResumeUnsetProfileFollowsHandWrittenDefaultMode (0.23s)
--- PASS: TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode (0.47s)
    --- PASS: TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode/stored_auto_profile (0.24s)
    --- PASS: TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode/typed_plan (0.22s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.559s
```

GREEN, same command, after the fix: every test and subtest above `--- PASS` (including `TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode`), final lines `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/cli	3.600s`.

Guards (green before the fix, as intended): `TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode` (stored `auto` profile + `--continue` forwards `auto`; typed `--permission-mode plan --continue` forwards `plan`), the existing fresh-launch default test (unset profile + Claude + fresh launch gives `--permission-mode auto`, not re-run in isolation but inside the broad run below), and the existing GLM resume tests. RED for the stated reason: the new/updated resume tests failed because the child argv carried `--permission-mode auto` on a resume.

Broad run, `go test ./internal/cli -run '^TestLaunch_' -count=1 -v` after the fix: exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	4.616s`; 29 `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP` lines. `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1`: all three `ok`. `gofmt -l internal/cli`: no output.

Gaps: no whole-package run of `internal/cli`, `internal/web`, `internal/template` or `internal/config` this round; no live Claude process (the child is a stub that records argv); the broad run was a name pattern, not the full suite.

The previous round's (round 6) audit made two codex calls whose results disagreed (receipts rcpt-813169891395590312ae33f6 pass, rcpt-ef9584ba0990325e17cfe714 fail).

The leader removed the guard records created after rounds 5 and 6 after backing them up (board 6a81; backup directories `backup-rejection-20261018T103222Z` and `backup-rejection-20261018T104643Z`; the second was made by the pre-reinstall guard binary).
