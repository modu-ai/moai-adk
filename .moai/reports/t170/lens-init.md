# t170 — lens: the `moai init` wizard surface

Read-only investigation. Every claim below carries `file:line` from the t170 worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`).

## Summary

The card's premise **holds in shape but not in the precedent it names**. `moai init` does
have a real interactive wizard the two new questions can join — a single `huh` v2 form
built from `wizard.InitQuestions(root)`, with a `QuestionTypeConfirm` kind that is exactly
the right shape for both new questions, and live examples of confirm→config wiring.

But `internal/cli/init_autonomy_wizard.go` — the file the card points at as "the closest
precedent" — **is dead code**: `applyAutonomyTierFromWizard` has no production call site
(only tests call it). Following it verbatim would produce a question that is asked, stored
into `WizardResult`, and then dropped. Two neighbouring writers are dead the same way:
`applyWorkflowBranchGuardFlags` (so the four `--branch-guard` / `--worktree-auto-*` flags
are registered but never applied) and `writeWorkflowAuditYAML` (so the audit block is never
written). Details in §1 and §4.

The **live** precedent to copy is instead:

- question shape → `worktree_auto_create` / `mcp_provision` confirms in
  `internal/cli/wizard/questions.go:367` and `:439`
- capture → `saveBoolAnswer` `internal/cli/wizard/wizard.go:459`
- wizard→opts → `applyWizardPage3ToOpts` `internal/cli/init.go:185`
- opts→yaml → `WritePhase1Configs` `internal/core/project/initializer_expansion.go:30`,
  invoked from `internal/core/project/initializer.go:246`

Second contradiction: the card requires the init prompt text be English "for 16-language
neutrality". The prompts live in **Go source, not templates**, so the template-neutrality
rule does not reach them — and the wizard is deliberately **localized into ko/ja/zh**
(`internal/cli/wizard/translations.go`), enforced by a completeness test that will FAIL if
a new question ships English-only. See §5.

---

## 1. `internal/cli/init_autonomy_wizard.go` — read in full (46 lines)

The whole file is one function:

```go
// init_autonomy_wizard.go:34-46
func applyAutonomyTierFromWizard(flagChanged bool, flagValue string, res *wizard.WizardResult, opts *project.InitOptions) {
	if flagChanged && flagValue != "" {
		opts.AutonomyTier = flagValue
		return
	}
	if res.AutonomyTier == "" {
		return
	}
	_, proofOK := config.SandboxProofKind()
	killSwitch := config.IsBypassDisabled()
	effective, _ := config.EffectiveTierWithGates(res.AutonomyTier, proofOK, killSwitch)
	opts.AutonomyTier = effective
}
```

It is only the **apply** half; the question itself is `internal/cli/wizard/questions.go:447-462`
(a `QuestionTypeSelect`, not a confirm), captured at `internal/cli/wizard/wizard.go:436-437`
(`case "autonomy_tier": result.AutonomyTier = value`).

**Prompt library**: `charmbracelet/huh` v2, one unified multi-group form. Fields are built
per question type; the confirm builder is `buildConfirmField`
(`internal/cli/wizard/wizard.go:470-503`), which binds localized Title/Description funcs
and localized Yes/No buttons, then saves through a `Validate` hook.

**Default expression**: `Question.Default` is a **string** for every type
(`internal/cli/wizard/types.go:93`); a confirm parses it with
`value := q.Default == "true"` (`wizard.go:472`). So default-OFF is `Default: "false"` and
default-ON is `Default: "true"` — precisely what the card's two questions need.

**Non-interactive**: there is no non-interactive branch inside this file. The whole wizard
is gated at `internal/cli/init.go:555` — `if !nonInteractive && isInteractiveStdin()`. When
that is false the wizard never runs and `WizardResult` stays zero, except for the five
booleans pre-seeded in `RunWithDefaults` (`wizard.go:47-53`).

**How the answer reaches config — it does not.** `applyAutonomyTierFromWizard` is never
called outside tests:

```
$ grep -rn "applyAutonomyTierFromWizard" --include='*.go' . | grep -v _test
./internal/core/project/autonomy_bundle.go:5:   (comment only)
./internal/cli/init_autonomy_wizard.go:4,19,34  (doc + definition)
```

The only non-test mentions are its own definition and two comments. `internal/cli/init.go`
never calls it, so `opts.AutonomyTier` is never set on the init path and
`ApplyAutonomyTierBundle` (`internal/core/project/autonomy_bundle.go:56`) always receives
the empty string.

## 2. `internal/cli/init.go` — where wizard questions run, and where two more go

- **Flag registration**: `init.go:69-127` (`func init()`), the workflow toggles at
  `:114-119`, autonomy at `:126`.
- **Opts seeded from flags**: `init.go:467-492`.
- **Wizard invocation**: `init.go:555-615`. The gate is `:555`; the call is
  `result, wizErr := runWizardFn(rootFlag, opts.ConvLang, opts.UserName)` at `:565`;
  cancel handling `:566-573`; identity/name/project/model/report application `:575-606`;
  and the terminal line `applyWizardPage3ToOpts(cmd, result, &opts)` at `:614`.

**Insertion points for two more questions** (three coordinated edits, all mechanical):

1. Question definitions → `internal/cli/wizard/questions.go` inside `Page3Questions`,
   alongside `worktree_auto_create` at `:365-374` (the "Quality & Workflow" group; both new
   questions share that group so they render on the same page).
2. Capture → `saveBoolAnswer` `internal/cli/wizard/wizard.go:459-468` (add two cases) plus
   two fields on `WizardResult` `internal/cli/wizard/types.go:12-72`.
3. Wizard→opts → `applyWizardPage3ToOpts` `internal/cli/init.go:185-224`, next to
   `opts.WorktreeAutoCreate = result.WorktreeAutoCreate` at `:210`.

## 3. Non-interactive path and the flag+precedence convention

**What init does when it cannot prompt**: the gate at `init.go:555` is
`!nonInteractive && isInteractiveStdin()`. `--non-interactive`, CI, or no TTY all skip the
wizard entirely; `opts` then carries only flag values plus the compiled defaults seeded at
`init.go:467-492`.

**Do wizard questions have matching flags?** Partially — and that is the convention to
follow. Three patterns exist:

| pattern | example | flag | precedence mechanism |
|---|---|---|---|
| flag + wizard, flag wins | `project_mode` | `--project-mode` `init.go:91` | `if !cmd.Flags().Changed("project-mode")` `init.go:186` |
| flag + wizard, default mismatch handled | `enable-lsp` | `--enable-lsp` `init.go:96` registered `false`, read `true` | `getBoolFlagWithDefault` `init.go:227-236` keys on `Changed()` |
| wizard-only, no flag | `worktree_auto_create`, `claude_design_enabled`, `report_format` | — | unconditional assignment `init.go:206-213` |

The load-bearing rule is documented at `init.go:181-184`:

> Explicitness is probed with `cmd.Flags().Changed(name)`, never by value:
> `getBoolFlag` / `getBoolFlagWithDefault` cannot distinguish "flag absent" from
> "flag explicitly set to the same value as the default".

The opt-in **tracker** variant (used when a zero value must not clobber a deployed template
default) is `applyWorkflowBranchGuardFlags` `internal/cli/init_workflow_flags.go:36-57`:

```go
if cmd.Flags().Changed("branch-guard") {
	v, _ := cmd.Flags().GetBool("branch-guard")
	opts.BranchGuardEnabled = v
	opts.BranchGuardSet = true
}
```

**Caveat (verified)**: that function has **no production call site** either —
`grep -rn "applyWorkflowBranchGuardFlags" --include='*.go' . | grep -v _test` returns only
its definition (`init_workflow_flags.go:36`) and a comment (`init.go:114`). So today
`moai init --branch-guard` and `--worktree-auto-create` parse and are then discarded. The
*pattern* is sound and is what the card's two questions should copy; the *wiring* is
missing and must be added, not assumed.

Pinning tests for the convention: `internal/cli/init_flag_precedence_test.go:60`
(`TestFlagBeatsWizard_Page3Settings`), `:236`, `:288`, `:312`; and
`internal/cli/init_workflow_flags_test.go:25/42/70`.

If the card wants CLI parity, it needs two flags registered in `init.go:69-127`
(e.g. `--feedback-auto-submit` default false, `--todo` default true, the latter registered
`false` and read with `getBoolFlagWithDefault(cmd, "todo", true)` following the
`--enable-lsp` idiom at `init.go:93-96` + `:487`).

## 4. How a wizard answer reaches `.moai/config/sections/*.yaml`

Two live writers, both called from `projectInitializer.Initialize` and both sitting
**outside** the deployer/fallback branch so they run on either path:

- `writeReportConfig` — `internal/core/project/initializer.go:501-515`, called at `:227`.
  Whole-file rewrite of `report.yaml`:
  ```go
  content := fmt.Sprintf("report:\n  format: %s\n", format)
  ... os.WriteFile(reportPath, []byte(content), defs.FilePerm)
  ```
- `WritePhase1Configs` — `internal/core/project/initializer_expansion.go:30-50`, called at
  `internal/core/project/initializer.go:246`. Fans out to `writeProjectModeYAML` (`:54`),
  `writeLSPYAML` (`:83`+), `writeQualityExpansionYAML`, `writeDesignYAML`. The
  **patch-don't-replace** idiom is at `:66`:
  ```go
  content = patchYAMLKey(string(existing), "project", "mode", mode)
  ```
  with the rationale at `:85-90` — an 11 KB deployed `lsp.yaml` must be patched at one leaf,
  never rewritten.

A third, comment-preserving patcher exists and is the best fit for a small boolean leaf:
`yamlpatch.PatchFile` with `[]yamlpatch.KeyEdit{{Path: []string{...}, Value: "false"}}` —
see `internal/cli/init_workflow_flags.go:97` and `buildWorkflowToggleEdits` `:106-134`.

**Dead writers (verified, both contradict the card's assumption of a working precedent):**

- `writeWorkflowAuditYAML` `internal/core/project/initializer_audit.go:37` —
  `grep -rn "writeWorkflowAuditYAML(" --include='*.go' internal | grep -v _test` returns
  only the definition. So `opts.AuditModel` / `AuditGate*` / `CodexAuditEnabled`, all set
  by `applyWizardPage3ToOpts` `init.go:216-223`, are never persisted.
- `opts.WorktreeAutoCreate` — declared `internal/core/project/initializer.go:59`, assigned
  `internal/cli/init.go:210`, and **read by nothing** in `internal/core/project`. The only
  reader of `workflow.worktree.auto_create` is `readWorktreeAutoCreate`
  `internal/cli/worktree_advisory.go:51`, which reads the deployed template value.

The one wizard confirm that *does* have a live effect is `mcp_provision`:
`opts.MCPProvision` (`init.go:222`) → `provisionMCPEntryUnlessDeclined` (`init.go:783`),
which writes `.mcp.json`, not a section yaml.

**Recommendation for the two new answers**: add a writer in `WritePhase1Configs`
(`initializer_expansion.go:30`) using `yamlpatch.PatchFile`, so it runs on both deploy
paths and preserves the template's comments.

## 5. Template-First and prompt-string language

**Template mirror exists and matches.** `internal/template/templates/.moai/config/sections/feedback.yaml`
is byte-identical to the local `.moai/config/sections/feedback.yaml` (both 356 B, 6 lines,
same content — `feedback: repository: modu-ai/moai-adk`). A new key must land in both, plus
the Go struct `FeedbackConfig` `internal/config/types.go:1310-1314`, its default
`NewDefaultFeedbackConfig` `internal/config/defaults.go:451`, and — if a new section file
is created — a loader (`l.loadFeedbackSection` is wired at `internal/config/loader.go:77`;
the registry row is `internal/config/audit_registry.go:41`). Then `make build`.

**Gap for the second question**: there is **no existing `todo` config key anywhere**.
`grep -rn "todo" .moai/config/sections/*.yaml internal/template/templates/.moai/config/sections/*.yaml`
returns only `mx.yaml:178 todo_per_file` (an MX tag limit, unrelated) and a tool-policy
audit string. So the backlog-queue toggle needs a new key, a new struct field, a default,
and a template mirror — decide its home section (`workflow.yaml` is the closest fit).

**Prompt strings live in Go source, not templates.** English source is
`internal/cli/wizard/questions.go` (e.g. `:371-372` for `worktree_auto_create`), and
ko/ja/zh translations live in `internal/cli/wizard/translations.go` (three locale blocks;
`autonomy_tier` at `:120`, `:262`, `:404`; `mcp_provision` at `:170`, `:312`, `:454`).

This **contradicts the card**: because the prompts are not under
`internal/template/templates/`, the 16-language template-neutrality guard does not apply to
them, and "keep the init prompt English" is not the local convention — the wizard is
deliberately 4-locale. Shipping English-only strings will **fail**
`TestWizardQuestionTranslationCompleteness`
(`internal/cli/wizard/translations_completeness_test.go:89-133`), whose doc comment says
verbatim: *"Adding a new question without translations will FAIL this test."* Note the two
language sets are distinct (16 programming languages for templates; 4 conversation locales
for the wizard).

## 6. Tests that pin the wizard surface

| test | file:line | what breaks when two questions are added |
|---|---|---|
| `TestWizardQuestionTranslationCompleteness` | `wizard/translations_completeness_test.go:89` | **Will fail** unless ko/ja/zh title+description are added for both new IDs. Confirms need no option translations (`q.Type != QuestionTypeSelect` skip at `:120`). |
| `TestQuestionOrder` | `wizard/questions_test.go:101` | Pins `DefaultQuestions` to exactly 5 IDs. Safe **only if** the new questions go in `Page3Questions`, not `DefaultQuestions`. |
| `TestReconfigureQuestions...` | `wizard/questions_test.go:190-210` | Pins `ReconfigureQuestions` to exactly 12 IDs. `ReconfigureQuestions` does not build on `InitQuestions` (`questions.go:294-295`), so page-3 additions do not leak — but adding to `DefaultQuestions` would break this too. |
| `TestRemovedQuestionsAbsentFromInitSet` / `...NoOrphanTranslations` | `wizard/question_removal_test.go:22`, `:37` | Unaffected by additions; will fail if a question is later added with a removed ID or a translation without a question. |
| `TestWorktreeAutoCreateQuestion` / `TestSaveBoolAnswerWorktreeAutoCreate` / `...TranslationsExist` | `wizard/worktree_test.go:8`, `:29`, `:47` | The exact template to clone for each new confirm (question exists + default + capture + translations). |
| `TestUnifiedForm_MultiGroupSinglePage`, `TestBuildFormGroups_Partition` | `wizard/unified_form_test.go:87`, `:242` | Group partitioning; new questions must carry `Group: "Quality & Workflow"` to stay on one page. |
| `TestFlagBeatsWizard_Page3Settings` | `cli/init_flag_precedence_test.go:60` | The pattern to extend if CLI flags are added — asserts `Changed()`-based precedence. |
| `TestInitCmd_WorkflowBranchGuardFlagsRegistered`, `TestApplyWorkflowBranchGuardFlags_*` | `cli/init_workflow_flags_test.go:25`, `:42`, `:70` | The opt-in-tracker pattern; note these pass while the function is unreachable in production. |
| `TestInitQuestions_HasAutonomyTierPage` | `wizard/autonomy_test.go:16` | Same shape for a select; the analogue to write for each new confirm. |
| `internal/cli/init_wizard_identity_test.go`, `init_test.go` | — | Not read in depth; may assert wizard output counts. **Gap.** |

---

## Gaps (explicitly named)

1. I ran **no tests** — every claim is from source reading, not from an observed test run.
   The "will fail" prediction for `TestWizardQuestionTranslationCompleteness` is read off
   the test body (`translations_completeness_test.go:97-118`), not from a failing run.
2. I did not read `internal/cli/init_test.go` (23 KB), `init_coverage_test.go`, or
   `wizard/wizard_test.go` (42 KB) in full — there may be further count/order assertions.
3. I did not trace `runWizardFn`'s definition/seam (only its call at `init.go:565`).
4. I did not check the `moai update --reconfigure` surface beyond noting
   `runWorkflowConfigStep` (`init_workflow_flags.go:68`) is also uncalled in production —
   whether the two new questions should also appear there is undecided.
5. I did not determine which section file the `todo` key should live in; no precedent exists.
6. The three dead-code findings are from `grep -rn ... | grep -v _test` over `*.go` in the
   repo root. A call reached through an interface value, a generated file, or a build-tagged
   file would not appear in that grep — I judge this unlikely (all three are unexported or
   package-local plain functions) but it is not proven.
