# Research — SPEC-FEEDBACK-PARTICIPATION-001

Read-only research for card t1498, tree `bb54f2903` (the code tree is identical to `2f492df19`; only the SPEC files differ). Source: two earlier read-only explorations, the plan-audit iteration 1 report, and the re-verification below. Anything not re-read in this tree is marked unverified. No test of the new feature was run (none exists); the only network calls were read-only `gh` queries against `modu-ai/moai-adk`; nothing was created, commented, or edited on GitHub. Version 0.3.0 corrects the premises the audit found false (section 9 lists the corrections); version 0.4.0 renames the SPEC identifier per the operator/leader decision (audit finding D22 — the design is real-name, the former identifier said ANON).

## 1. Existing feedback subsystem (reused, not duplicated)

- No Go code files an issue. The submit path is skill prose: `internal/template/templates/.claude/skills/moai/workflows/feedback.md:177` tells the orchestrator to run `gh issue create --repo <resolved-target>` under the user's own `gh` login. Go mentions of the command are comments only (`internal/feedback/queue.go:3`, `scrub.go:39`).
- `internal/feedback/`: `Scrub` (classify on raw text, then mask env values, secrets, collapse home), classifier (block threshold, English-only vocabulary), `QueueStore` (`.moai/state/feedback/queue.json`, 0600, locked atomic `Mutate`), mask log. `QueueItem.Attempts` exists (`queue.go:71`) and no non-test code increments it; there is no sender and no retry cap.
- Config: `feedback.repository` and `feedback.auto_submit` (template default false; `auto_submit` means skip the confirmation gate, not participate). The `feedback` console section is already `RouteSeam` (`internal/settings/sectionroute.go`, feedback row). The shipped `feedback.yaml` carries `repository: modu-ai/moai-adk` and `auto_submit: false`, and equals its local mirror (`cmp` exits 0).
- The completed SPEC's own follow-up list already named wrapping the issue creation in Go and moving the duplicate search after scrub; this SPEC does the first for the automatic path only.

## 2. init and update

- Init runs its wizard only when `!nonInteractive && isInteractiveStdin()` (`internal/cli/init.go:648`); `isInteractiveStdin` is an injectable isatty wrapper (`internal/cli/init_update_notice.go`). Init has no CI-environment detection beyond the missing terminal.
- Quiet-wizard decision to amend: `feedback_auto_submit` was removed from init (`removedQuietInit` in `internal/cli/wizard/question_removal_test.go`); `TestInitQuestions_QuietSet` pins `conversation_language`, `user_name`, `agent_wiring`, `autonomy_tier`, `jev_enabled`. The stepper denominator is the question count, so a sixth question moves several pins. Enumerated by search: 20 test files mention `InitQuestions`, `Page3Questions`, or `TotalVisibleQuestions` (listed in plan.md section C, M2). Pins read in this tree: `TestInitStepper_Denominator4`, `TestInitRegroup_SecondGroupGolden` (`init_regroup_test.go`), `TestTotalVisibleQuestions_Page3AlwaysCounted` (`expansion_test.go`, `got != 5`), `TestProfileWizardStepper_SameFormatAsInit` (`profile_stepper_test.go`, init groups `{2,2,1}`, N = 5), `TestStepperTotal_DynamicDenominator` (`wizard_test.go`, `got != 8`), `TestInitPages_Membership` and `TestInitPages_MergeIntoOneGroupPerPage` (`restructure_test.go`).
- Jev precedent: `JevQuestionID`, `Page3Questions` slot with its own group, `saveBoolAnswer`, `applyJevFromWizard(wizardRan, res, root)` writing nothing if the wizard did not run, persisting through `settings.ApplySchemaEdits` (`internal/settings/jev.go`, which calls `ApplySchemaEdits` with the key `workflow.jev.enabled`).
- Bare `moai update` has no prompt of its own; `--yes` is declared (`internal/cli/update.go`) and read only in `update_version.go:324`. The only TTY-gated interactive precedent is `runWorkflowConfigStep` (`internal/cli/init_workflow_flags.go`), reached from the reconfigure path (`update_wizard.go:125`). The flag inventory of the update command, read from `update.go:71-93`: `--check`, `--shell-env`, `--config`/`-c`, `--force`, `--yes`, `--templates-only`, `--binary`, `--dry-run`, `--no-hooks`, `--no-plugin`, `--restore`, `--verbose`, `--profile` (retired), `--version`.

## 3. moai web and the settings schema

- Feedback rows are in `internal/settings/schema_sections.go` (a `TypeText` repository and a `TypeBool` `auto_submit`); the tab blurb is `internal/web/settings_shell.go:158`; i18n keys for `f.feedback.*` exist in four locales in `internal/web/assets/i18n.js` (lines 697-700, 1661-1664, 2493-2496, 3325-3328).
- `settings.ApplySchemaEdits` (`internal/settings/sectionapply.go`) resolves each edited name with `Field(name)` and errors on an unknown name (`unknown schema field`), dispatches `PersistSeam` and `PersistTypedSection`, and returns an error for any other persist kind (`default:` branch). `PersistProfileStore` fields are handled elsewhere. The persist kinds are defined at `internal/settings/schema.go:124-141`. `sectionFileFor` (`sectionvalues.go:25`) returns an empty file name for any other kind, and the web editable predicate `schemaEditableField` (`internal/web/schemaform.go:362`) admits only seam and typed fields. A rendered schema field is counted by `TestFeedbackPanelFieldsWired` (the panel renders `len(SectionFields)` fields) and needs `.title` and `.desc` keys in four locales for `TestI18nKeySetParity`.
- The console renders every bool as a two-option radio pair with a hidden `<name>__present` companion and fails on any checkbox: `TestBoolFieldsRenderAsRadio` (`internal/web/widget_policy_test.go:59`), `TestSchemaTogglePresentCompanion` (`:83`), `TestParseSchemaFormBoolSemanticsUnchanged` (`:96`).
- The comment at `internal/settings/schema_sections.go:442` says the console routes the Jev field to its own panel through `isJevFieldName`; a search for that identifier across `internal` finds only that comment, so no such function exists and there is no routing precedent to follow.
- Files that mention the feedback fields or branch on `Persist.Kind` in tests are listed in plan.md section C (M2).
- Withdrawal today is the bool toggle only; no local purge exists.

## 4. Detection surface (moai-internal only)

- `cmd/moai/main.go` calls `cli.Execute()` and maps the error to an exit code; there is no top-level recover and no `debug.Stack` or `runtime.Caller` in non-test code.
- Twelve non-test `recover()` sites exist (listed in `design.md` section 2); eleven swallow a panic, one defends a benign closed-channel race.
- Goroutine launches: 43 non-test files contain a `go` statement (`grep -rEl '^[[:space:]]+go (func|[A-Za-z_.]+\()' internal cmd pkg --include='*.go' --exclude='*_test.go'`); a panic in one of those goroutines bypasses a recover in `main`. Hooks run synchronously (`registry.go:108`), so hook panics are on the main goroutine.
- Hooks: `registry.Dispatch` returns a wrapped `ErrHookTimeout` (`registry.go:125`; the sentinel is defined at `internal/hook/errors.go:8`) or `handler %d for event %s: %w` (`:133`) and does not recover a panicking handler. The `Handler` interface has `Handle` and `EventType` only (`internal/hook/types.go:624`), so a handler's identity is its Go type name. The production wiring registers about thirty handlers in `internal/cli/deps.go:233-293`, some wrapped by `hook.WithEscalationConfig(...)`.
- Template sentinels: five in `internal/template/errors.go`; the render and validate sites are at `renderer.go:135,149` and `validator.go:56,59`; deploy path sentinels at `deployer.go:457,462,474`.
- Config sentinels usable for user attribution: `internal/config/errors.go:17-38`.
- `claudeNotFoundError` is unexported in `internal/cli` (`claude_binary.go:46-47`); a package that cannot import `internal/cli` cannot name it.
- Tool-failure events (`post_tool_failure.go`, `failure_observer.go`) describe Claude Code tool runs and are out of scope; no auto-report may read them.
- No fingerprint, dedupe, or frequency-cap primitive exists for bug reports. Version and commit come from `pkg/version/version.go` (`GetVersion`, `GetCommit`).
- Build identity and paths: `.goreleaser.yml:20-24` and the `Makefile` `LDFLAGS` carry `-s -w` and `-X` and no `-trimpath`, so frame file paths would carry the builder's absolute paths; the design never reads them.

## 5. Privacy gaps in the existing scrubber

Not covered today: absolute paths other than the `$HOME/` prefix, usernames outside paths, host names, emails, repository and organisation names, project names, environment values beyond five names, session ids, and semantic code snippets. The classifier vocabulary is English-only. The existing pre-submit duplicate search sends title keywords before scrubbing. Consequence driving the design: an unattended reporter sends only machine-generated fixed fields, built by Go code, with the scrubber as a tripwire rather than the primary control.

## 6. Why the anonymity options were dropped (operator decision, 2026-10-04)

The earlier draft laid out five ways to hide the reporter (relay plus GitHub App, shipped or fetched bot token, user's own account, non-GitHub intake, pre-filled browser URL). The operator replaced the question: participation is real-name, filed from the user's own `gh` account, with no relay, bot, or credential to run. The options appendix and its decision gate were removed; nothing in this SPEC depends on a mechanism choice.

## 7. Where consent can live (added in 0.3.0)

The audit found that consent in a tracked project file lets a cloned repository post from the next user's account. Facts read in this tree:

- `.moai/config/sections/feedback.yaml` is tracked (`git ls-files` lists it) and carries `feedback.repository`.
- This repository's own `.gitignore` ignores `.moai/config/local/` (line 282), `.moai/config/sections/llm.yaml` (line 287), `.moai/state/` (line 400), and `.moai/logs/` (line 399); `git check-ignore -v` confirms each. The shipped template `.gitignore` (`internal/template/templates/.gitignore`) lists `.moai/state/` (line 241) and `.moai/logs/*` (line 252) but a search finds no `.moai/config/local` line in it. The audit's statement that the template ignores `.moai/config/local/` was not reproduced; an untracked project-local consent file therefore cannot be relied on in a user's project, which is one reason the consent lives under the user's home directory instead.
- The config resolver has a local tier (`.moai/config/local/*.yaml`, `internal/config/resolver.go:389-413`) and a user tier (`~/.moai/config/sections/*.yaml`, `resolver.go:312-335`), merged by tier priority, where the enum documents that higher-priority sources override lower ones (`internal/config/source.go`). Which of the user, project, and local tiers wins was not traced here; the point is that a merged read lets a value from a project-tier file decide the result. A consent value must not take part in that merge, so the SPEC reads it directly from `<moai home>/config/participation.yaml`, beside and not inside `config/sections/`.
- `paths.MoaiHome()` (`internal/paths/paths.go:68`) honours the `MOAI_HOME` environment variable (`EnvHome`, line 30), and `EnsureHomeLayout` creates `<moai home>/config` with mode 0700 (`internal/homestate/paths.go:265-295`). Precedent for a direct home-tier file read with its own wrapper struct and a compiled default for an absent file: `loadHomeRetentionDays` (`internal/cli/clean_home.go:97-120`).
- `internal/config` depends on `internal/config/atomicfile`, `internal/defs`, `internal/paths`, `pkg/models`, and the YAML library only (`go list -deps ./internal/config`), so a reader there is importable by a leaf `bugreport`.
- `internal/settings` already imports `yamlpatch`, used by `WriteSectionViaSeam` (`sectionwrite.go:99-121`, which patches `<project>/.moai/config/sections/<section>.yaml` only).

## 8. Code paths that reach a model (corrects the earlier premise)

The earlier research (P14) searched only for the literal `exec.Command("claude"|"codex")` and concluded that no text-generation call exists in the Go tree. The variable-aware search `grep -rn 'exec.CommandContext\|http.NewRequest' internal/cli --include='*.go' --exclude='*_test.go'` finds model-reaching code:

- The headless `claude -p` path: `claudeAuditArgs` builds the arguments (`internal/cli/mcp_claude.go:212-218`: `-p`, `--input-format text`, `--output-format json`, `--json-schema`, `--safe-mode`, `--restricted`, `--tools ""`, `--strict-mcp-config`, `--permission-mode dontAsk`, `--no-session-persistence`, `--model`, `--effort`) and `runClaudeCommand` executes it (`internal/cli/mcp_claude_runner.go:48-54`, `exec.CommandContext(ctx, binary, args...)`, prompt on stdin, bounded output writers, `WaitDelay` 2 s). `RunAuthStatus` runs `claude auth status --json` (`mcp_claude_runner.go`).
- GLM clients: `internal/cli/todo_classify_llm.go:108` (`http.NewRequest` POST to the GLM endpoint, injectable doer), `internal/cli/glm_task.go:318` and `internal/cli/mcp_glm.go:332` (`http.NewRequestWithContext` POST).
- Codex runners: `internal/cli/mcp_codex.go:443,526` and `internal/cli/codex_audit_launch.go:666` (`exec.CommandContext` with the codex binary).

The weighing of these candidates, including privacy and cost, is in `design.md` section 8 and the decision in plan.md (DEC-6).

## 9. Verified premises

| # | Premise | Command or read | Observed result |
|---|---|---|---|
| P1 | No Go code creates an issue | `grep -rn -E "issue create\|IssueCreate\|CreateIssue" --include="*.go" internal cmd pkg` (non-test) | two comment-only hits: `internal/feedback/queue.go:3`, `internal/feedback/scrub.go:39` |
| P2 | The skill prose runs `gh issue create` | `grep -n "gh issue create" .claude/skills/moai/workflows/feedback.md` | lines 143 and 177 |
| P3 | `Attempts` is never incremented | `grep -rn "Attempts" --include="*.go" internal/feedback internal/cli/feedback.go` (non-test) | one hit: the field at `queue.go:71` |
| P4 | Init wizard gate | read `internal/cli/init.go` | `if !nonInteractive && isInteractiveStdin()` at line 648 |
| P5 | `--yes` has one reader | `grep -rn '"yes"' --include="*.go" internal/cli` | declared in `update.go`, read `update_version.go:324` (other hits are other commands) |
| P6 | No top-level recover; twelve recover sites | `grep -rn "recover()" --include="*.go" internal cmd pkg` (non-test) | twelve hits, all under `internal/`, none in `cmd/` or `pkg/` |
| P7 | No stack capture today | `grep -rn -E "debug\.Stack\|runtime\.Caller" ...` | no output |
| P8 | No `-trimpath` | `grep -n trimpath .goreleaser.yml Makefile` | no output |
| P9 | Import direction allows a leaf `bugreport` | `go list -deps ./internal/config`, `./internal/template`, `./internal/hook`, `./internal/feedback` filtered for this module | config: `atomicfile`, `defs`, `paths`, `pkg/models`; template: `config` and `manifest`; hook: `cli/preference`, `template`, `config`, `resilience`, `escalation`, `guardliveness`, no `cli`, no `feedback`; feedback: `hook`, `cli/preference`, `escalation`, `resilience`, `guardliveness`, no bare `cli` |
| P10 | Config sentinels | read `internal/config/errors.go:17-38` | `ErrConfigNotFound`, `ErrInvalidConfig`, `ErrSectionNotFound`, `ErrInvalidDevelopmentMode`, `ErrNotInitialized`, `ErrSectionTypeMismatch`, `ErrDynamicToken`, `ErrInvalidYAML` |
| P11 | Template sentinel sites | grep of `ErrPathTraversal`, `ErrInvalidJSON`, `ErrMissingTemplateKey`, `ErrUnexpandedToken` | `deployer.go:457,462,474`; `validator.go:56,59`; `renderer.go:135,149` |
| P12 | Feedback section is already web-writable | read `internal/settings/sectionroute.go` | `"feedback": RouteSeam` |
| P13 | Hook dispatch error branches | read `internal/hook/registry.go:95-135` | timeout wrap at `:125`, handler-error wrap at `:133` |
| P14 | **RETRACTED (false premise).** The earlier row claimed only the claude availability probe launches a model binary | the earlier literal-name grep | superseded by section 8: the variable-aware search finds the headless runner, the GLM clients, and the Codex runners |
| P15 | `gh` flag surface | `gh issue list --help`, `gh issue create --help`, `gh issue comment --help`, `gh issue view --help` on `gh 2.92.0` | `--search --state --json --limit --repo`; `--title --body-file --repo`; `--body-file --repo`; `--json` fields include `number,title,state,comments,body,url` |
| P16 | Docs wording to correct | grep per locale (ledger entry E18 in `acceptance.md`) | the "created automatically" phrasing is present in all four pages (en 2 lines, ko 1, ja 1, zh 1) |
| P17 | New-package RED-now | `go test` on the three new package paths | `directory not found`, exit 1 (ledger E3, E4, E11, E12) |
| P18 | Wizard pin tests exist under the named identifiers | read `question_removal_test.go`, `init_regroup_test.go`, `expansion_test.go`, `profile_stepper_test.go`, `wizard_test.go`, `restructure_test.go` | see section 2 |
| P19 | Profile-wizard guard is profile-only | read `internal/cli/init_update_profile_path_test.go` (lines around 175 and 205) | it counts profile-wizard seam calls and one needle, not generic prompts |
| P20 | `gh issue list --json comments` returns comment bodies | `gh issue list --repo modu-ai/moai-adk --state all --limit 3 --json number,comments --jq '...'` (read-only) | each `comments` value is an array; each element has keys `author`, `authorAssociation`, `body`, `createdAt`, `id`, `includesCreatedEdit`, `isMinimized`, `minimizedReason`, `reactionGroups`, `url`, `viewerDidAuthor` |
| P21 | The update flag inventory | read `internal/cli/update.go:71-93` | fourteen flags listed in section 2 |
| P22 | The user-scoped home directories exist | read `internal/homestate/paths.go` `EnsureHomeLayout`, `internal/paths/paths.go` | `<moai home>/config` created 0700; `MOAI_HOME` overrides the home |
| P23 | Hook handler registration site | read `internal/cli/deps.go:233-293` | about thirty `deps.HookRegistry.Register(...)` calls, some wrapped |
| P24 | Goroutine launch inventory | `grep -rEl` (section 4) | 43 non-test files |
| P25 | `claudeNotFoundError` is unexported | read `internal/cli/claude_binary.go:46-47` | `type claudeNotFoundError struct{}` |
| P26 | `isJevFieldName` does not exist | search `internal` for the identifier | one hit: the comment at `schema_sections.go:442` |
| P27 | The shipped template ignores `.moai/config/local/` | read `internal/template/templates/.gitignore` (search for `config/local`) | no such line; only this repository's own `.gitignore` carries it |

Unverified, listed as gaps rather than claims: a live `gh` search against GitHub (tokenisation of a 16-hex token under `in:title`, index latency); whether `gh issue list --json comments` truncates the array for an issue with very many comments; whether a non-collaborator's issue comment and issue creation succeed on the target repository (standard GitHub behaviour, not exercised); the exact YAML library error type for a syntax error in a user file; that Go prints `[...]` for generic type arguments in `runtime.Frame.Function`; that the installed `claude` accepts the audit flag set for a summary prompt and how an unauthenticated session fails; the exact set of settings and web tests that change when the new schema field lands (the lists in plan.md are searches; the run phase runs the packages in full); the legal or policy position on bot-filed issues (not applicable once the filer is the user).

## 10. Open items

Zero open clarifications remain: every former question is a recorded decision in `plan.md` section B. The last two — ambiguous attribution and the harness-defect narrowing — were resolved by operator decisions recorded there as DEC-7 and DEC-8.
