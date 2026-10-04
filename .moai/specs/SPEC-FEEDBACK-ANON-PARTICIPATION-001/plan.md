# Plan — SPEC-FEEDBACK-ANON-PARTICIPATION-001

## §A Context

- Card t1498 (operator scope as corrected 2026-10-04): opt-in, real-name participation; automatic issue filing from the user's own `gh` account only for moai-adk-attributed tool errors; zero-token deterministic path with a model reached only when a new issue is about to be published; no auto-repair artifact in deployed templates.
- Standalone SPEC, Tier L (new packages, more than 15 files touched across `internal/bugreport`, `internal/feedback`, `internal/cli`, `internal/config`, `internal/settings`, `internal/web`, three skill-body copies, four docs pages). Frontmatter `depends_on: SPEC-FEEDBACK-AUTO-SUBMIT-001`.
- Branch and worktree: the card worktree branches from `develop`; the commit trail carries the card id.
- Methodology: TDD per `quality.yaml`; every named test in `acceptance.md` is written first (RED) and its failure observed before the milestone is called done.
- Version 0.3.0 is the revision after plan-audit iteration 1 (FAIL, 0.62). The consent store moved from the tracked project file to a user-scoped file; attribution became an allowlist; the milestone set (M1 to M7) and the 25-requirement and 25-criterion counts are unchanged in shape, and the old-to-new criterion map is in `acceptance.md`.

## §B Open clarifications (Kickoff inputs)

Exactly two remain. Both are operator-owned, both ship with a safe default, and both variants are specified so that answering flips a constant or a recorded scope line, not the design. This plan recommends nothing on either.

1. [NEEDS CLARIFICATION: ambiguous attribution — does an ambiguous verdict call the model, or stay local and unsent?] The card says deterministic rules come first and a model is used only when attribution is ambiguous; the leader's message says the model is called only at the moment an issue is published. Both variants are specified and tested (REQ-ANON-009, AC-008, AC-009, AC-019); the shipped constant is `local` until answered. Facts for the choice: the kinds that default to `ambiguous` are `hook_timeout`, a hook handler error without a marker, and the template tokens `unexpanded_token` and `invalid_json`; the model sees only fixed fields, so for `hook_timeout` it has no more information than the rules do. Under `adjudicate` an ambiguous item can cost two calls (one adjudication, one summary), bounded by the daily cap.
2. [NEEDS CLARIFICATION: acceptance of narrowing "deployed harness defect" to the shipped-template render and validate sentinels] The card lists "deployed harness defect" as a signal; the tree has a machine-detectable signal only for hooks and for shipped templates that fail to render or validate. A behavioural defect of an agent or skill has none, and inferring one from Claude Code tool-failure stubs is excluded (REQ-ANON-007). The shipped scope covers the render and validate tokens only, pinned by the closed-kind test (AC-007). If the operator does not accept the narrowing, this SPEC's behaviour is unchanged and a follow-up SPEC adds a deployed-asset self-check as a new signal source, extending the closed token set by exactly one member (a visible one-line change to that test).

(The numbering in earlier drafts skipped values on purpose; the former clarifications 3, 4, 7, 8 and the model-reach question are now recorded decisions below, and the consent-location and attribution-allowlist questions the audit raised are decided in `design.md` sections 3 and 9.)

### Decisions (recorded; each states the shipped default, the reason, and the one place that changes it)

- **DEC-1 consent scope and storage.** Consent is per user, stored in the user-scoped file `<moai home>/config/participation.yaml`, read by a fail-closed reader that never opens a project file. Reason: consent is a property of the GitHub account and the person, not of a repository; a tracked file would let a cloned repository post from the next user's account (audit D2). Changes at: the path and reader in `internal/config` (one file).
- **DEC-2 flush trigger set.** `moai feedback participation flush` and the end of a plain `moai update`; no session-start trigger. Reason: a session-start trigger would need a detached, time-boxed process and must never run inside a hook dispatch; reports are rare by design (caps), so a slower drain is acceptable. Changes at: adding a trigger is additive in M5 (one call site).
- **DEC-3 cap values.** Per-fingerprint window 7 days; 3 reports per rolling day; 10 per rolling week; queue bound 20; attempt limit 5; daily model-call cap 6; per-issue occurrence-comment cap 50 (design section 10). Reason: conservative volume for an unattended public action. Changes at: `internal/config/defaults.go`, one constant each.
- **DEC-4 withdrawal purge scope.** Turning participation off makes the next flush discard unsent items and the spool; the sent-history log stays; the purge command removes the queue, spool, ledger, and outbox log. Reason: conservative and already specified (REQ-ANON-021); the alternative of leaving unsent items inert until purge keeps data the user withdrew consent for. Changes at: the flush discard branch (one function).
- **DEC-5 fingerprint merge granularity.** Follow the card: the fingerprint includes version, commit, operating system and architecture, kind, and the moai-internal frames, so one defect on two commits or two systems yields separate issues that the consumer merges by fingerprint family. Reason: the card defines it; a coarser key would hide which build is affected. Changes at: the canonical-input function and its golden (one place).
- **DEC-6 how Go reaches a model.** Reuse the existing headless `claude` runner through one model seam owned by `publish` and injected from `internal/cli`; the consent text names that the summary may spend the user's own subscription tokens; the deterministic template text is used whenever the model is unavailable, unauthenticated, or fails. Reason and the candidates weighed (including the GLM and Codex paths and their privacy and cost): `design.md` section 8; the earlier premise "no text-generation call exists" was false. Changes at: the production seam implementation in `internal/cli` (one file); `publish` is untouched.

## §C Milestones (ordered by decision reversibility: types and user-facing flows first, mechanical work last)

Priority High for M1 to M5; Medium for M6; Medium for M7. Complete each before starting the next; M2 and M3 may proceed in parallel worktrees only if their file sets stay disjoint, which they do except `internal/config/defaults.go` (integration serialised).

### M1 — Taxonomy, payload, fingerprint (new types, High)
Creates the new interfaces everything else binds to; most expensive to change later.
Files: `internal/bugreport/kind.go`, `verdict.go`, `payload.go`, `validate.go`, `detail.go`, `fingerprint.go`, `frames.go`, `marker.go`, and `_test.go` siblings; constants in `internal/config/defaults.go`.
First tests: a generic-function canary pinning how the runtime names a generic frame (design section 4 premise); the closed-set `detail` rejection test with the canary and path-shaped strings.
Proves: AC-010, AC-011.

### M2 — Consent and setting surfaces (user-facing flow, High)
Files: the user-scoped reader `internal/config/participation_user.go`; the writer `internal/settings/participation.go`; `internal/settings/schema.go` (new `PersistUserScoped`), `sectionapply.go`, `sectionvalues.go`, `schema_sections.go`; `internal/web/schemaform.go` (editable-kind predicate), `internal/web/assets/i18n.js`, `settings_shell.go`; `internal/cli/wizard/questions.go`, `translations.go`, `types.go`, `wizard.go`, a new `participation_question_test.go`; `internal/cli/init_participation_wizard.go` (new), `init.go`; `internal/cli/update_participation.go` (new), `update.go`. No template section file and no `shipped_key_inventory.yaml` row is added: the consent keys are not shipped (REQ-ANON-001).
Proves: AC-001, AC-004, AC-005, AC-006, AC-021.
Test inventory the milestone must run in full and amend where they fail (the lists come from searches at `bb54f2903`; which tests actually change is confirmed by running the packages in full):
- Init-wizard pin tests (`grep -rln 'InitQuestions\|Page3Questions\|TotalVisibleQuestions' internal --include='*_test.go'`, 20 files): `internal/cli/`: `agent_model_flags_retired_test.go`, `ptycap_child_test.go`, `init_agent_wizard_test.go`, `init_update_profile_path_test.go`; `internal/cli/wizard/`: `question_removal_test.go`, `init_regroup_test.go`, `expansion_test.go`, `ptycap_test.go`, `jev_question_test.go`, `help_surfaces_test.go`, `translations_completeness_test.go`, `autonomy_test.go`, `profile_stepper_test.go`, `restructure_test.go`, `questions_test.go`, `questions_harness_test.go`, `wizard_test.go`, `translation_keys_test.go`, `agent_wiring_question_test.go`, `layout_alignment_test.go`. Read in this tree and known to pin a count or a membership: `TestInitQuestions_QuietSet` and `TestRemovedQuestionsAbsentFromInitSet`; `TestInitStepper_Denominator4` and `TestInitRegroup_SecondGroupGolden`; `TestTotalVisibleQuestions_Page3AlwaysCounted` (`got != 5`); `TestProfileWizardStepper_SameFormatAsInit` (init groups `{2,2,1}`, N = 5); `TestStepperTotal_DynamicDenominator` (`got != 8`, "5 + 3 page-3"); `TestInitPages_Membership` and `TestInitPages_MergeIntoOneGroupPerPage`. The axis golden under `internal/cli/wizard/testdata/axis/` moves with the group golden.
- Settings and web guard tests the new schema field touches (files that mention the feedback fields or branch on `Persist.Kind`): `internal/settings/`: `feedback_autosubmit_test.go`, `schema_sections_test.go`, `write_safety_test.go`, `schema_test.go`, `layer_atomicity_probe_test.go`, `schema_todo_test.go`, `jev_test.go`, `section_gate_test.go`; `internal/web/`: `widget_policy_test.go`, `write_safety_test.go`, `feedback_panel_test.go`, `schema_sections_test.go`, `schema_label_test.go`, `tab_layout_test.go`; `internal/cli/`: `schema_bridge_test.go`, `init_quiet_wizard_test.go`; `internal/config/`: `cache_test.go`. Expected to change by their role: `TestSchemaCurrentValuesReadsAllSections` (a `feedback.participation` case read from the user-scoped file), `TestFeedbackPanelRendered` (probes for `feedback.participation` and its `__present` companion), `TestFeedbackPanelFieldsWired` (follows `SectionFields`), `TestI18nKeySetParity` (four-locale title and description keys), `TestAbsentDefaultPolarityDeclared` (the new bool declares its polarity). Expected to need `MOAI_HOME` pointed at a temporary directory because a console save now reaches the user-scoped file: `TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical`, `TestSaveSchemaSmokeAllSections`, `TestSchemaSectionsRenderSmoke`. Must stay green unchanged: `TestBoolFieldsRenderAsRadio`, `TestSchemaTogglePresentCompanion`, `TestParseSchemaFormBoolSemanticsUnchanged`.
Amendment: the commit message records that this re-adds an init question and amends the quiet-wizard decision (REQ-ANON-022).
Release ordering: this milestone must not ship in a release before M5, otherwise users are asked to consent to a sender that does not exist (section F).

### M3 — Detection and attribution wiring (High)
Files: `internal/bugreport/capture.go`, `spool.go`, `attribution.go`, `register.go`; `cmd/moai/main.go`; one capture call in each recover site listed in `design.md` section 2 (`internal/guardliveness/evaluator.go`, `internal/cli/codex_stop_chain.go`, `internal/resilience/circuit.go`, `internal/navigator/route/run.go`, `internal/navigator/fix/request.go`, `internal/escalation/detector.go`, `internal/hook/session_start_guard_liveness.go`, `internal/hook/user_decision_capture.go`, `internal/hook/session_start_binary_lag.go`, `internal/hook/navigator_detect.go`); `internal/hook/registry.go` plus the hook event and handler name registration; `internal/cli/preference/cmd.go`; `internal/cli/update_template_sync.go`; `internal/cli/update_clean_install.go`; guard tests including the recover-site walk and the handler-name guard in `internal/cli`.
First test: measure the error type the YAML library returns for a syntax error in a user file (design section 3 row U3), and record it.
Proves: AC-002, AC-007, AC-008.

### M4 — Local pipeline, preview, log, withdrawal (High)
Files: `internal/feedback/outbox/*.go` (drain, ledger, caps, tripwire, queue wiring, log); `internal/feedback/queue.go` (optional fields, `omitempty`, backward compatible); `internal/cli/feedback_participation.go` (preview, purge, flush); tests.
Proves: AC-012, AC-013, AC-014, AC-022.

### M5 — Publication through the user's gh, deterministic text (High)
Files: `internal/feedback/publish/ghrunner.go`, `sender.go`, `lookup.go`, `contract.go`, `template.go`, `internal/feedback/testdata/bugreport_issue_v1.golden`; the flush call at the end of `update.go`; tests with a `gh` stub.
First test item: feasibility of the title-token search (design section 7, audit D23): a recorded read-only query, against a throwaway public issue whose title carries a 16-hex token, showing whether `--search "<token> in:title"` returns it and how long the index takes; if it cannot be made reliable, the recorded decision is to accept that two issues may be created for a fingerprint (the edge already stated in `acceptance.md`), and the consumer merges them.
Proves: AC-003, AC-015, AC-016, AC-020. After M5 the whole feature works with zero model calls.

### M6 — Model step, ambiguous policy, reachability guard (Medium)
Files: `internal/feedback/publish/model.go` (the `Summarizer` interface), `adjudicate.go`, `budget.go`; `internal/cli/feedback_participation_model.go` (the production implementation over the existing headless runner, injected at the flush call); the static guard tests; tests with a counting stub. The policy value comes from clarification 1; until answered the policy is `local`. The channel is decided (DEC-6).
First test: record that the flag set `claudeAuditArgs` passes today is accepted by the installed `claude` for a summary prompt, and that an unauthenticated session yields the template fallback.
Proves: AC-009, AC-017, AC-018, AC-019, AC-025.

### M7 — Docs, skill bodies, shipping guards (mechanical, Medium)
Files: `internal/template/templates/.claude/skills/moai/workflows/feedback.md`, `plugins/moai/skills/moai/workflows/feedback.md`, `.claude/skills/moai/workflows/feedback.md`; `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md`; `internal/template/auto_repair_guard_test.go` (new); `internal/cli/wizard/participation_guard_test.go` (new, the coexistence guard); `make build`.
Proves: AC-023, AC-024.

## §D MX tag plan

- `@MX:ANCHOR` on `bugreport.Capture` (called from the main recover, ten recover sites, hook registry, update paths: fan-in well above three), on `bugreport.Fingerprint` (every dedupe, cap, title, and comment keys on it), and on the user-scoped reader (capture, drain, sender, preview, console, wizard, and update all read it).
- `@MX:WARN` with `@MX:REASON` on: the capture hot path (time-boxed, fail-open, must stay network-free); the `gh` runner (public side effect from the user's account); the model seam (token spend, daily cap); the scrub tripwire stage (withholding, not masking); the consent reader (fail-closed, must never open a project file).
- `@MX:NOTE` on each recover allowlist entry (the reason is the note) and on the frame filter (why file paths are never read).
- `@MX:DEBT` with `@MX:CEILING` and `@MX:UPGRADE` on the interim `AmbiguousPolicy=local` constant, naming clarification 1 as the upgrade trigger.

## §E Risks

| Risk | Vector | Mitigation |
|---|---|---|
| Consent smuggled in by a cloned repository | a tracked project file carrying `participation: true`, or a redirected `feedback.repository` | consent lives only in the user-scoped file; the reader never opens a project file; the mutant fixture (tracked file says true, no user consent) kills a reader that falls back (AC-001, AC-003) |
| Over-attribution to moai | an input rejection, a decode error, or a subprocess exit from a user's repository state reaching a public issue | `moai` is an allowlist; unmarked handler errors are `ambiguous`; exec-exit, JSON and YAML decode errors are non-moai rows (AC-008) |
| Stack frames leak user data | frame file paths embed the builder's absolute paths because no `-trimpath` is set; closure or method names could in principle carry data | frames use `Function` only, never file, line, or args; module-prefix filter; closed-set `detail`; canary test (AC-010, AC-011) |
| Panic values or error text leak | `os.PathError` carries a full path; `internal/core/project/root.go:114` panics with a message that embeds an error | the builder takes no error, panic value, or string; attribution uses `errors.Is` and `errors.As` only; type-level signature test and canary (AC-011) |
| Handler identity leaks | `%T` of a handler | only a registered name from a closed table is used; the guard fails when a handler is wired without a table entry (AC-007) |
| Real-name exposure | the issue and comment are public, tied to the user's account, their creation time is public, and the filer is subscribed to later comments | consent text states all of it in four locales (AC-006); caps keep volume low; the per-issue comment cap bounds notification fan-out; no client timestamp is sent |
| Comment flood and forged markers | any user can post marker-shaped comments; a common defect gathers many occurrence comments | the sender skips commenting at the cap; the count is advisory; markers are documented as untrusted input and consumers re-derive fields from the title key (AC-016, AC-020) |
| Hook timeouts from machine load | a slow host looks like a bug | `hook_timeout` defaults to `ambiguous`; per-fingerprint and global caps |
| Vulnerability disclosed publicly by a crash report | a path-traversal sentinel or a security-flavoured function name | `path_traversal` reports are always withheld; classifier tripwire; a function name that trips the classifier blocks its own report (fails safe) |
| Duplicate issues from concurrent first filers | search index latency or title-token tokenisation (unverified) | accepted; the M5 first-test item records the measurement; the consumer merges by fingerprint; comments are append-only so counts do not lose updates |
| Count drift | deleted comments, repeat users, forged markers | accepted as approximate; documented in `design.md` section 7 |
| Surprise prompt at update | a user who changed the toggle in the console before being asked | resolved: a console change of `enabled` also sets `asked`, so that user is not asked again (design section 9) |
| Token spend | model called more than once per issue, or again on retry | single seam, stored summary reused on retry, per-item bound, daily cap, counting-stub budget table (AC-018, AC-019) |
| A model path outside the seam | a pipeline that shells out or posts elsewhere bypasses the counting stub | the static import guard (AC-025) |
| Console save writes the real home directory in tests | the new field persists to the user-scoped file | every test reaching it sets `MOAI_HOME` to a temporary directory (acceptance preface) |
| Settings schema blast radius | a new persist kind reaches `sectionFileFor`, `SchemaCurrentValues`, `ApplySchemaEdits`, and the web editable predicate | the M2 inventory above is run in full; the kind is added once and the value-invariant gate is kept |
| Flush slows `moai update` | network work at the end of update | 10 s time box; failure never fails the command |
| Pin-test churn | adding a question moves a stepper denominator, a membership, and a golden | the full inventory is enumerated above and amended in the same commit (AC-004) |
| Test env coupling | tests that read `CI` | every such test sets the variable with `t.Setenv` |
| Panics on other goroutines | 43 non-test files launch goroutines; a panic there bypasses `main`'s recover | recorded gap (design section 2); a goroutine launcher helper is the later repair |

## §F Release ordering

No release may carry M2 (the consent question) without M5 (the sender). The coexistence guard (AC-024, `TestParticipationQuestionRequiresSender`) fails when the participation question exists in the wizard while the sender package is absent, and its own canary test observes that red on a synthetic tree. The guard lands in M7, so it protects the branch at integration: the card branch merges as a whole, and a cherry-pick of M2 alone onto a tree without M5 trips it. Rationale: asking for consent to a feature that cannot act would record a consent the user cannot exercise.

## §G Anti-patterns to avoid

- Reading consent, or the sender's target repository, from any project-tier file.
- Treating `moai` as the default verdict; it is an allowlist.
- Passing a free string where `detail` is meant, or building `detail` from a type name at capture time.
- Building a transport abstraction for anonymity; the transport is `gh`, full stop.
- Reading `err.Error()` anywhere in the capture or payload path.
- Calling the model before the duplicate lookup, or again on a retry that has a stored summary.
- Editing an issue body to count occurrences.
- Importing `os/exec` or `net/http` into `bugreport` or the outbox package.
- Adding an init question to `DefaultQuestions` (it would reach the reconfigure path and break its pin).

## §H Cross-references

`spec.md`, `acceptance.md`, `design.md`, `research.md`; `.claude/rules/moai/development/spec-frontmatter-schema.md`; `.claude/rules/moai/development/verification-completeness.md` (RED-now ledger form).
