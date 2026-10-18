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

## What the wizard saves (source read, not run end to end)

`runProfileSetup` writes `PermissionMode: permissionMode` (`profile_setup.go`, the preferences literal) from the wizard result, so a user who accepts the pre-selected `auto` stores `"auto"` explicitly; the same holds for the pre-selected `sonnet[1m]` and `medium`. The launcher cells above cover what such a stored `auto` does at launch.

## Corrections made during the card (recorded so they are not re-inferred)

- A defect candidate raised earlier in the card — "the web console normalizes `acceptEdits` to empty, so a user who picks it launches in default mode" — does not exist. No non-test caller applies `Persist.Normalize` for `permission_mode` (every `.Normalize` call site in the tree was read), and `TestCharacterize_MainSessionEffortSaveWritesPreferencesOnly` already pins that the web persists `acceptEdits` as it is. Withdrawn. Only the schema function's unit contract changed (`auto` → empty, `acceptEdits` kept); it is not wired into the web save path (leader ruling).
- `TestPermissionModeNormalizeAcceptEdits` carried a source-grep guard that claimed the TUI normalizes `acceptEdits` to empty. The TUI never did; the guard matched a string used for the confirmation line. It was replaced by `TestPermissionModeNormalizeDefaultAuto`, which checks the schema function only.
- My first test for the deployed `llm.yaml` used an initializer constructed without a template deployer; that initializer never writes `llm.yaml`, so the test was wrong and was removed. The deployed file's content is covered by `TestGLMTierDefaultsTemplateMatchesGoDefaults` (embedded file equals the Go constants) plus the literal assertions on the constants.

## Gaps

- Whole-package runs observed passing: `internal/settings` and `internal/cli/wizard` only. Not run whole: `internal/config`, `internal/core/project` (its whole-package result was not observed; only the named report tests and the init test), `internal/template` and `internal/cli`, and no landing re-measure has run. Package-wide suites need the `go-test-heavy` slot; the lane has not held it for this card. `internal/cli` ran only the tests named above.
- `moai init` in a temporary directory with a temporary HOME was not run: the worktree guard refuses a `HOME=` override, and an init without it could write to the real user folders. A binary built from this tree exists in scratch (machine-local) and was not used. What a fresh project receives is inferred from: `report.yaml` in the embedded templates (read), `writeReportConfig` (`TestInit_ReportYAMLDefaultsToArtifact`, observed), and the embedded `llm.yaml` equal to the Go constants (observed).
- That `moai update` leaves an existing user's value for `llm.glm.effort.low` / `.fable` or `report.format` alone is covered by the generic update-preservation tests (5 passed); no test changes those two keys specifically and runs update.
- `moai web`'s report description text (`internal/web/schemaform.go`, `i18n.js`) was not changed or checked beyond a grep; `docs-site/content/ko` carries no `report.format` text; the en, ja and zh docs-site trees were not searched.
- The web console's empty-option labels for `model` and `effort_level` ("(project default)", "(model policy, else Claude Code default ...)") still describe an unset profile; only the TUI wizard pre-selection changed for those two keys.
- The `internal/web` suite ran only for the tests named `Permission|acceptEdits|AcceptEdits|permission_mode|MainSession|Normaliz` (3 tests).
- The built binary was not run, so the launcher change was observed through the exec seam in tests, not in a real launch.

## Residual-risk

- Users who already have `permission_mode` unset now start Claude sessions in `auto`; where Claude Code refuses `auto` for the account or model, the first launch after the change is where that would show.
- A hand-edited project `settings.json` `defaultMode` is shadowed for those users by the `settings.local.json` override.
- A `--permission-mode auto` typed on the GLM command line keeps its existing refusal; a stored `auto` on GLM is now dropped silently except for one stderr line.
