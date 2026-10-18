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
