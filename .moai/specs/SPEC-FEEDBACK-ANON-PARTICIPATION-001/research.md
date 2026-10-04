# Research — SPEC-FEEDBACK-ANON-PARTICIPATION-001

Read-only research for card t1498, tree `2f492df19`. Source: two earlier read-only explorations plus the re-verification below. Anything not re-read in this tree is marked unverified. No test was run; no network call was made.

## 1. Existing feedback subsystem (reused, not duplicated)

- No Go code files an issue. The submit path is skill prose: `internal/template/templates/.claude/skills/moai/workflows/feedback.md:177` tells the orchestrator to run `gh issue create --repo <resolved-target>` under the user's own `gh` login. Go mentions of the command are comments only (`internal/feedback/queue.go:3`, `scrub.go:39`).
- `internal/feedback/`: `Scrub` (classify on raw text, then mask env values, secrets, collapse home), classifier (block threshold, English-only vocabulary), `QueueStore` (`.moai/state/feedback/queue.json`, 0600, locked atomic `Mutate`), mask log. `QueueItem.Attempts` exists (`queue.go:71`) and no non-test code increments it; there is no sender and no retry cap.
- Config: `feedback.repository` and `feedback.auto_submit` (template default false; `auto_submit` means skip the confirmation gate, not participate). The `feedback` console section is already `RouteSeam` (`internal/settings/sectionroute.go`, feedback row).
- The completed SPEC's own follow-up list already named wrapping the issue creation in Go and moving the duplicate search after scrub; this SPEC does the first for the automatic path only.

## 2. init and update

- Init runs its wizard only when `!nonInteractive && isInteractiveStdin()` (`internal/cli/init.go:648`); `isInteractiveStdin` is an injectable isatty wrapper (`internal/cli/init_update_notice.go`). Init has no CI-environment detection beyond the missing terminal.
- Quiet-wizard decision to amend: `feedback_auto_submit` was removed from init (`removedQuietInit` in `internal/cli/wizard/question_removal_test.go`); `TestInitQuestions_QuietSet` pins `conversation_language`, `user_name`, `agent_wiring`, `autonomy_tier`, `jev_enabled`. The stepper denominator is the question count (the `●/○` marks), so a sixth question moves `TestInitStepper_Denominator4` and `TestInitRegroup_SecondGroupGolden`; `internal/cli/wizard/expansion_test.go:226` asserts a count of 5.
- Jev precedent: `JevQuestionID`, `Page3Questions` slot with its own group, `saveBoolAnswer`, `applyJevFromWizard(wizardRan, res, root)` writing nothing if the wizard did not run, persisting through `settings.ApplySchemaEdits` (`internal/settings/jev.go`).
- Bare `moai update` has no prompt of its own; `--yes` is declared (`internal/cli/update.go:75`) and read only in `update_version.go:324`. The only TTY-gated interactive precedent is `runWorkflowConfigStep` (`internal/cli/init_workflow_flags.go`), reached from the reconfigure path (`update_wizard.go:125`).

## 3. moai web

Feedback rows are in `internal/settings/schema_sections.go` (a `TypeText` repository and a `TypeBool` `auto_submit`); the tab blurb is `internal/web/settings_shell.go:158`; i18n keys for `f.feedback.*` exist in four locales in `internal/web/assets/i18n.js` (lines 697-700, 1661-1664, 2493-2496, 3325-3328). Guard tests a new bool will touch: `TestShippedConfigKeysHaveReaders`, `TestFeedbackPanelRendered`, `TestFeedbackAutoSubmitI18nKeysInAllLocales`, `TestFeedbackPanelFieldsWired`, `TestI18nKeySetParity`, and the settings schema tests. Withdrawal today is the bool toggle only; no local purge exists.

## 4. Detection surface (moai-internal only)

- `cmd/moai/main.go` calls `cli.Execute()` and maps the error to an exit code; there is no top-level recover and no `debug.Stack` or `runtime.Caller` in non-test code.
- Twelve non-test `recover()` sites exist (listed in `design.md` section 2); eleven swallow a panic, one defends a benign closed-channel race.
- Hooks: `registry.Dispatch` returns a wrapped `ErrHookTimeout` (`registry.go:125`) or `handler %d for event %s: %w` (`:133`) and does not recover a panicking handler.
- Template sentinels: five in `internal/template/errors.go`; the render and validate sites are at `renderer.go:135,149` and `validator.go:56,59`; deploy path sentinels at `deployer.go:457,462,474`.
- Config sentinels usable for user attribution: `internal/config/errors.go:17-38`.
- Tool-failure events (`post_tool_failure.go`, `failure_observer.go`) describe Claude Code tool runs and are out of scope; no auto-report may read them.
- No fingerprint, dedupe, or frequency-cap primitive exists for bug reports. Version and commit come from `pkg/version/version.go` (`GetVersion`, `GetCommit`).
- Build identity and paths: `.goreleaser.yml:20-24` and the `Makefile` `LDFLAGS` carry `-s -w` and `-X` and no `-trimpath`, so frame file paths would carry the builder's absolute paths; the design never reads them.

## 5. Privacy gaps in the existing scrubber

Not covered today: absolute paths other than the `$HOME/` prefix, usernames outside paths, host names, emails, repository and organisation names, project names, environment values beyond five names, session ids, and semantic code snippets. The classifier vocabulary is English-only. The existing pre-submit duplicate search sends title keywords before scrubbing. Consequence driving the design: an unattended reporter sends only machine-generated fixed fields, built by Go code, with the scrubber as a tripwire rather than the primary control.

## 6. Why the anonymity options were dropped (operator decision, 2026-10-04)

The earlier draft laid out five ways to hide the reporter (relay plus GitHub App, shipped or fetched bot token, user's own account, non-GitHub intake, pre-filled browser URL). The operator replaced the question: participation is real-name, filed from the user's own `gh` account, with no relay, bot, or credential to run. The options appendix and its decision gate were removed; nothing in this SPEC depends on a mechanism choice.

## 7. Verified premises

| # | Premise | Command or read | Observed result |
|---|---|---|---|
| P1 | No Go code creates an issue | `grep -rn -E "issue create\|IssueCreate\|CreateIssue" --include="*.go" internal cmd pkg` (non-test) | two comment-only hits: `internal/feedback/queue.go:3`, `internal/feedback/scrub.go:39` |
| P2 | The skill prose runs `gh issue create` | `grep -n "gh issue create" .claude/skills/moai/workflows/feedback.md` | lines 143 and 177 |
| P3 | `Attempts` is never incremented | `grep -rn "Attempts" --include="*.go" internal/feedback internal/cli/feedback.go` (non-test) | one hit: the field at `queue.go:71` |
| P4 | Init wizard gate | read `internal/cli/init.go` | `if !nonInteractive && isInteractiveStdin()` at line 648 |
| P5 | `--yes` has one reader | `grep -rn '"yes"' --include="*.go" internal/cli` | declared `update.go:75`, read `update_version.go:324` (other hits are other commands) |
| P6 | No top-level recover; twelve recover sites | `grep -rn "recover()" --include="*.go" internal cmd pkg` (non-test) | twelve hits, all under `internal/`, none in `cmd/` or `pkg/` |
| P7 | No stack capture today | `grep -rn -E "debug\.Stack\|runtime\.Caller" ...` | no output |
| P8 | No `-trimpath` | `grep -n trimpath .goreleaser.yml Makefile` | no output |
| P9 | Import direction allows a leaf `bugreport` | `go list -deps ./internal/config`, `./internal/template`, `./internal/hook` filtered for `hook\|template\|feedback\|cli\|config` | config: none of hook/template/feedback/cli; template: config only; hook: template and config, not feedback or cli |
| P10 | Config sentinels | read `internal/config/errors.go:17-38` | `ErrConfigNotFound`, `ErrInvalidConfig`, `ErrSectionNotFound`, `ErrInvalidDevelopmentMode`, `ErrNotInitialized`, `ErrSectionTypeMismatch`, `ErrDynamicToken`, `ErrInvalidYAML` |
| P11 | Template sentinel sites | grep of `ErrPathTraversal`, `ErrInvalidJSON`, `ErrMissingTemplateKey`, `ErrUnexpandedToken` | `deployer.go:457,462,474`; `validator.go:56,59`; `renderer.go:135,149` |
| P12 | Feedback section is already web-writable | read `internal/settings/sectionroute.go` | `"feedback": RouteSeam` |
| P13 | Hook dispatch error branches | read `internal/hook/registry.go:110-135` | timeout wrap at `:125`, handler-error wrap at `:133` |
| P14 | Only the claude availability probe launches a model binary | `grep -rln -E 'exec\.Command(Context)?\("(claude\|codex)"' --include="*.go" internal cmd` (non-test) | one file, `internal/cli/doctor.go` |
| P15 | `gh` flag surface | `gh issue list --help`, `gh issue create --help`, `gh issue comment --help`, `gh issue view --help` on `gh 2.92.0` | `--search --state --json --limit`; `--title --body-file --repo`; `--body-file --repo`; `--json` fields include `number,title,state,comments,body,url` |
| P16 | Docs wording to correct | grep per locale (ledger entry E8 in `acceptance.md`) | the "created automatically" phrasing is present in all four pages (en 2 lines, ko 1, ja 1, zh 1) |
| P17 | New-package RED-now | `go test` on the three new package paths | `directory not found`, exit 1 (ledger E2, E18, E19) |
| P18 | Wizard pin tests exist under the named identifiers | read `question_removal_test.go`, `init_regroup_test.go` | `TestInitQuestions_QuietSet`, `TestRemovedQuestionsAbsentFromInitSet`, `TestInitStepper_Denominator4` (marks equal question count, currently 5), `TestInitRegroup_SecondGroupGolden` |
| P19 | Profile-wizard guard is profile-only | read `internal/cli/init_update_profile_path_test.go` (lines around 175 and 205) | it counts profile-wizard seam calls and one needle, not generic prompts |

Unverified, listed as gaps rather than claims: a live `gh` search against GitHub (tokenisation of a 16-hex token under `in:title`, index latency); whether a non-collaborator's issue comment and issue creation succeed on the target repository (standard GitHub behaviour, not exercised); the exact YAML library error type for a syntax error in a user file; that Go prints `[...]` for generic type arguments in `runtime.Frame.Function`; the behaviour of any Claude Code runtime feature referenced only by earlier notes; the legal or policy position on bot-filed issues (not applicable once the filer is the user).

## 8. Open items

The open questions live in `plan.md` section B as marked clarifications, not here.
