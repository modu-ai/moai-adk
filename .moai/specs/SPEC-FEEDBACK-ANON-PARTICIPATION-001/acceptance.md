# Acceptance — SPEC-FEEDBACK-ANON-PARTICIPATION-001

> Each criterion names one verification command and one observable expectation. Named test functions are deliverables of the named milestone: they are written first (RED) and judged afterwards. A `-run` selector that sweeps zero tests (`[no tests to run]`) is a failure of the criterion, never a pass. Local verification is package-scoped; the full suite is CI's.

**Criteria: 25 of the Tier L ceiling 25. Requirements: 23 of 25.**

## §D AC Matrix

| AC | REQ | Milestone | Summary |
|---|---|---|---|
| AC-ANON-001 | REQ-ANON-001 | M2 | Both keys ship `false` in template, local mirror, fixtures, inventory |
| AC-ANON-002 | REQ-ANON-002 | M3 | Participation off: capture is a no-op, zero network, zero model calls |
| AC-ANON-003 | REQ-ANON-003 | M2 | Init question: one, default no, own group after Jev, absent from shared sets, pins amended |
| AC-ANON-004 | REQ-ANON-003 | M2 | Init persistence only when the wizard ran and `CI` is empty |
| AC-ANON-005 | REQ-ANON-004 | M2 | Update asks once; never when non-interactive, CI, `--yes`, or another mode |
| AC-ANON-006 | REQ-ANON-005 | M2 | Consent text in four locales states public, GitHub-account-tied, fields, preview, web, no recall |
| AC-ANON-007 | REQ-ANON-006 | M3 | Closed kind enum, registered emit sites, every recover site reports or is allowlisted |
| AC-ANON-008 | REQ-ANON-007 | M3 | User-code, tool-failure, and user-config signals are discarded |
| AC-ANON-009 | REQ-ANON-008 | M3 | Capture is fail-open, non-blocking, bounded, network-free |
| AC-ANON-010 | REQ-ANON-009 | M3 | Deterministic ordered attribution table: every row, first match wins |
| AC-ANON-011 | REQ-ANON-009 | M6 | Ambiguous verdict under both policy values |
| AC-ANON-012 | REQ-ANON-010 | M1 | Fingerprint stable; frames filtered by module path; no paths, lines, foreign frames |
| AC-ANON-013 | REQ-ANON-011 | M1 | Payload is a closed schema; canary text never enters it |
| AC-ANON-014 | REQ-ANON-012 | M4 | Scrub and classify tripwire; withheld on block, masking, or path traversal |
| AC-ANON-015 | REQ-ANON-013 | M4 | Per-fingerprint window, global caps, queue bound, attempt limit |
| AC-ANON-016 | REQ-ANON-014 | M4 | Preview is byte-identical to what is handed to `gh`; append-only log, mode 0600 |
| AC-ANON-017 | REQ-ANON-015 | M5 | Sender re-checks consent per item, off the hook path, time-boxed, quiet when `gh` is absent |
| AC-ANON-018 | REQ-ANON-016 | M5 | Existing fingerprint issue gets an occurrence comment; zero model calls |
| AC-ANON-019 | REQ-ANON-017 | M6 | New moai issue: at most one model call, validated output, template fallback |
| AC-ANON-020 | REQ-ANON-018 | M6 | Model-call budget: zero calls in every excluded case, one in the positive control |
| AC-ANON-021 | REQ-ANON-019 | M5 | Issue title and marker contract round-trips; no body edit, no labels |
| AC-ANON-022 | REQ-ANON-020 | M2 | Web toggle (bool), four-locale i18n, guard tests, marker key not rendered |
| AC-ANON-023 | REQ-ANON-021 | M4 | Withdrawal discards unsent items; purge removes all local state; no network |
| AC-ANON-024 | REQ-ANON-022 | M7 | Template-First mirrors, skill-body copies, docs-site wording corrected |
| AC-ANON-025 | REQ-ANON-023 | M7 | No auto-repair artifact in deployed templates or the plugin tree |

All criteria are release-blocking. The tree pin for every RED-now observation below is the commit `2f492df19` (measured in this run, in this worktree).

## RED-now evidence ledger

Each entry: the single-invocation command, its verbatim stdout, its exit code, and the pinned tree. An empty stdout with exit 1 is a complete observation.

| Id | Command | Stdout (verbatim) | Exit | Why red |
|---|---|---|---|---|
| E1 | `grep -c participation internal/template/templates/.moai/config/sections/feedback.yaml` | `0` | 1 | key absent |
| E2 | `go test ./internal/bugreport/ -run '^TestCaptureNoopWhenParticipationOff$' -count=1` | `# ./internal/bugreport` / `stat .../internal/bugreport: directory not found` / `FAIL	./internal/bugreport [setup failed]` | 1 | package absent |
| E3 | `grep -c feedback_participation internal/cli/wizard/questions.go` | `0` | 1 | question absent |
| E4 | `grep -rl applyParticipationFromWizard internal/cli` | (empty) | 1 | seam absent |
| E5 | `grep -rl runParticipationStep internal/cli` | (empty) | 1 | update step absent |
| E6 | `grep -c feedback_participation internal/cli/wizard/translations.go` | `0` | 1 | translations absent |
| E8 | `grep -c -E "issue is created automatically|auto-creates a GitHub issue|GitHub 이슈로 자동 생성|GitHub Issue として自動作成|自动创建为 GitHub Issue" docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md` | `en:2`, `zh:1`, `ja:1`, `ko:1` (one `path:count` line per file) | 0 | the "created automatically" wording is still present in all four locales (the criterion expects four `0` counts) |
| E9 | `grep -c feedback_participation internal/cli/wizard/question_removal_test.go` | `0` | 1 | pin tests not yet amended |
| E10 | `grep -c participation internal/settings/schema_sections.go` | `0` | 1 | schema row absent |
| E11 | `grep -c f.feedback.participation.title internal/web/assets/i18n.js` | `0` | 1 | i18n keys absent |
| E13 | `grep -rl TestParticipationPreview internal/cli` | (empty) | 1 | test absent |
| E14 | `grep -rl TestParticipationPurge internal/cli` | (empty) | 1 | test absent |
| E15 | `ls internal/feedback/testdata/bugreport_issue_v1.golden` | `ls: ...: No such file or directory` | 1 | golden absent |
| E16 | `grep -c participation internal/template/templates/.claude/skills/moai/workflows/feedback.md` | `0` | 1 | skill body silent on participation |
| E17 | `grep -rl TestNoAutoRepairArtifactsShipped internal/template` | (empty) | 1 | guard absent |
| E18 | `go test ./internal/feedback/outbox/ -count=1` | `# ./internal/feedback/outbox` / `stat .../internal/feedback/outbox: directory not found` / `FAIL	./internal/feedback/outbox [setup failed]` | 1 | package absent |
| E19 | `go test ./internal/feedback/publish/ -count=1` | `# ./internal/feedback/publish` / `stat .../internal/feedback/publish: directory not found` / `FAIL	./internal/feedback/publish [setup failed]` | 1 | package absent |
| E20 | `grep -rl internal/bugreport internal cmd` | (empty) | 1 | no caller imports the capture package |

Observation caveat: E2, E18, E19 were observed with the exit code read from a separate invocation; stdout was captured with a bounded `head`. Ids skip E7 and E12 on purpose; none was needed after the package split.

Mutant-probe note: AC-013, AC-014, AC-018, and AC-020 each state the mutant that must die (an all-accepting validator, a scrub that never runs, a sender that edits the issue body, a stub that counts only the publish path). A criterion whose mutant survives is too shallow to adopt.

## Scenarios

### AC-ANON-001 — Keys, defaults, mirrors
Maps REQ-ANON-001
- **Given** the shipped template section file and the local mirror,
- **When** the config package loads the feedback section from a fixture with neither key,
- **Then** `Participation` and `ParticipationAsked` are both false, and the template file, the local file, the settings fixture, and `shipped_key_inventory.yaml` each carry both keys.
- **Verify**: `go test ./internal/config/ -run '^(TestFeedbackParticipationDefaultsOff|TestShippedConfigKeysHaveReaders)$' -count=1 -v` — expect exit 0 and `--- PASS: TestFeedbackParticipationDefaultsOff `.
- **RED-now**: E1. **Green path**: M2 adds the keys, accessors, mirrors, inventory rows.

### AC-ANON-002 — Off means nothing happens
Maps REQ-ANON-002
- **Given** `feedback.participation` is false or absent,
- **When** the capture entry point, the drain, and the sender are each invoked with a recording filesystem, a recording network stub, and a counting model stub,
- **Then** no spool file, queue file, or log is written, the network stub records zero requests, and the model stub records zero calls.
- **Verify**: `go test ./internal/bugreport/ -run '^TestCaptureNoopWhenParticipationOff$' -count=1 -v` — expect exit 0 and the test named in the PASS line.
- **RED-now**: E2. **Green path**: M3.

### AC-ANON-003 — Init question slot and pins
Maps REQ-ANON-003
- **Given** the init question set,
- **When** it is built,
- **Then** exactly one question has id `feedback_participation`, type confirm, default `false`, its own group, positioned after `jev_enabled`; it is absent from `DefaultQuestions` and `ReconfigureQuestions`; the removed-question set still contains `feedback_auto_submit` and not `feedback_participation`; and the amended pin tests pass with the new id list.
- **Verify**: `go test ./internal/cli/wizard/ -run '^(TestParticipationQuestion|TestInitQuestions_QuietSet|TestRemovedQuestionsAbsentFromInitSet|TestInitStepper_Denominator4|TestInitRegroup_SecondGroupGolden)$' -count=1 -v` — expect exit 0 and a PASS line for each named test.
- **RED-now**: E3, E9. **Green path**: M2 (the amended pins land in the same commit as the question; the commit message records the amendment).

### AC-ANON-004 — Init persistence gate
Maps REQ-ANON-003
- **Given** a project root with `feedback.yaml`,
- **When** `applyParticipationFromWizard` runs with (a) the wizard not run, (b) the wizard run and `CI` empty, (c) the wizard run and `CI` set to a non-empty value, (d) a nil result,
- **Then** (a), (c), (d) leave the file byte-identical; (b) writes `participation` and `participation_asked` through `settings.ApplySchemaEdits` and a decline over an existing `true` writes `false`.
- **Verify**: `go test ./internal/cli/ -run '^TestApplyParticipationFromWizard$' -count=1 -v` — expect exit 0 and four subtests passing.
- **RED-now**: E4. **Green path**: M2.

### AC-ANON-005 — Update asks once, never unattended
Maps REQ-ANON-004
- **Given** an interactive-terminal seam (`isInteractiveStdin` stubbed true) and `participation_asked: false`,
- **When** the update step runs, then runs again, and then runs under each of: stdin not a terminal, `CI` non-empty, `--yes`, `--check`, `--binary`, `--dry-run`, `--restore`, `--config`,
- **Then** the first run prompts once and persists both keys; the second run does not prompt; every listed other mode makes zero prompt calls and leaves the file unchanged; `--yes` never sets `participation` to true.
- **Verify**: `go test ./internal/cli/ -run '^TestUpdateParticipationStep$' -count=1 -v` — expect exit 0 and the table subtests passing with a prompt-call counter.
- **RED-now**: E5. **Green path**: M2.

### AC-ANON-006 — Consent text
Maps REQ-ANON-005
- **Given** the question translations for en, ko, ja, zh,
- **When** each locale's title and description are read,
- **Then** each states that the issue is public and tied to the user's GitHub account, lists the fixed fields, names `moai feedback participation preview` and `moai web`, and states the tool cannot recall a filed issue; and the update prompt text equals the init text per locale.
- **Verify**: `go test ./internal/cli/wizard/ -run '^(TestParticipationQuestion_FourLocalesStateFacts|TestWizardQuestionTranslationCompleteness)$' -count=1 -v` — expect exit 0.
- **RED-now**: E6. **Green path**: M2.

### AC-ANON-007 — Kind enum, sites, recover guard
Maps REQ-ANON-006
- **Given** the emit-site register and the source tree,
- **When** the guard walks every non-test `recover()` in `internal/`, `cmd/`, and `pkg/`,
- **Then** each either calls the capture entry point or is on the reasoned allowlist; each registered site file contains its capture call; the kind enum has exactly six members, each with a verdict and at least one registered site.
- **Verify**: `go test ./internal/bugreport/ -run '^(TestKindEnumClosed|TestEmitSiteRegister|TestEveryRecoverSiteReportsOrIsAllowlisted)$' -count=1 -v` — expect exit 0.
- **RED-now**: E2, E20. **Green path**: M3.

### AC-ANON-008 — Exclusions
Maps REQ-ANON-007
- **Given** synthetic signals built from user-config errors, a tool-failure event, and a user test-run failure,
- **When** each is offered to capture,
- **Then** none is recorded; and the files `internal/hook/post_tool_failure.go` and `internal/hook/failure_observer.go` contain no reference to the capture package.
- **Verify**: `go test ./internal/bugreport/ -run '^(TestUserAndToolFailuresNeverCaptured|TestToolFailureHandlersDoNotImportCapture)$' -count=1 -v` — expect exit 0.
- **RED-now**: E2. **Green path**: M3.

### AC-ANON-009 — Capture is fail-open and bounded
Maps REQ-ANON-008
- **Given** an unwritable spool directory, a full spool, and a filesystem stub that blocks,
- **When** capture is called under `-race`,
- **Then** it returns without error and without panic, returns within the time box on the blocking stub, drops the signal when the spool is full, and the network stub records zero requests.
- **Verify**: `go test ./internal/bugreport/ -run '^TestCaptureNeverPanicsOrBlocks$' -race -count=3 -v` — expect exit 0.
- **RED-now**: E2. **Green path**: M3.

### AC-ANON-010 — Deterministic attribution
Maps REQ-ANON-009
- **Given** a table of signals covering every row of the attribution register (filesystem, syscall, network, missing-tool, cancellation errors; config sentinels; panic; hook handler error; hook timeout; explicit internal marker; template sentinels),
- **When** attribution runs on each,
- **Then** the verdict equals the table's verdict, an error chain matching both an environment row and a moai row yields `environment` (first match wins), and no row calls a model (a counting stub records zero).
- **Verify**: `go test ./internal/bugreport/ -run '^(TestAttributionRulesTable|TestAttributionFirstMatchWins|TestAttributionMakesNoModelCall)$' -count=1 -v` — expect exit 0.
- **RED-now**: E2. **Green path**: M3.

### AC-ANON-011 — Ambiguous policy, both variants
Maps REQ-ANON-009
- **Given** an `ambiguous` signal and a counting model stub,
- **When** the policy constant is `local`, then `adjudicate`,
- **Then** under `local` the signal stays local, nothing is queued, and the stub records zero calls; under `adjudicate` the stub records exactly one call whose input is the validated payload fields only, an answer other than the two permitted tokens keeps the signal local, and a `moai` answer continues the pipeline.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestAmbiguousPolicyLocal|TestAmbiguousPolicyAdjudicate)$' -count=1 -v` — expect exit 0.
- **RED-now**: E19. **Green path**: M6.

### AC-ANON-012 — Fingerprint
Maps REQ-ANON-010
- **Given** captured stacks that include moai frames, standard-library frames, dependency frames, and file paths under a user home directory,
- **When** the fingerprint and frame list are computed,
- **Then** the frame list contains only function names beginning with `github.com/modu-ai/moai-adk/` with the prefix stripped, no `/`-path, line number, or argument text; identical inputs give identical fingerprints; changing version, commit, os, arch, kind, or any frame changes the fingerprint; a panic with zero moai frames is kept local.
- **Verify**: `go test ./internal/bugreport/ -run '^(TestFingerprintStable|TestFingerprintInputsAllMatter|TestFramesExcludePathsLinesAndForeignModules|TestZeroFramePanicStaysLocal)$' -count=1 -v` — expect exit 0.
- **RED-now**: E2. **Green path**: M1.

### AC-ANON-013 — Payload is a closed schema
Maps REQ-ANON-011
- **Given** a panic whose value is `CANARY-/Users/leak/secret-token` and an error whose text embeds a path,
- **When** the payload is built and serialised,
- **Then** the bytes contain neither the canary nor any path; unknown fields and values failing an anchored allowlist are rejected; the builder's signature takes no `error` or `any` parameter (checked by a type-level test). Mutant that must die: an all-accepting validator.
- **Verify**: `go test ./internal/bugreport/ -run '^(TestPayloadSchemaClosed|TestPanicValueAndErrorTextNeverEnterPayload|TestBuilderTakesNoErrorOrAny)$' -count=1 -v` — expect exit 0.
- **RED-now**: E2. **Green path**: M1.

### AC-ANON-014 — Scrub and classify tripwire
Maps REQ-ANON-012
- **Given** a payload whose frame list is mutated to contain a secret-pattern string, another containing a vulnerability phrase, and a report derived from a path-traversal sentinel,
- **When** the pipeline renders and screens each,
- **Then** each is withheld: not queued, not handed to the publication seam, zero model calls, one `withheld` row in the outbox log; a clean payload passes with zero masking findings. Mutant that must die: a pipeline that never runs the scrubber.
- **Verify**: `go test ./internal/feedback/outbox/ -run '^(TestOutboundPayloadScrubbedAndClassified|TestBlockedOrMaskedPayloadWithheld|TestPathTraversalKindWithheld)$' -count=1 -v` — expect exit 0.
- **RED-now**: E18. **Green path**: M4.

### AC-ANON-015 — Dedupe, caps, bounds
Maps REQ-ANON-013
- **Given** a controllable clock and the constants from `internal/config/defaults.go`,
- **When** the same fingerprint is offered twice inside and once outside the window, more reports than the daily and weekly caps are offered, the queue is filled past its bound, and an item fails to its attempt limit,
- **Then** the repeat inside the window is deduplicated, the cap overflow is capped, the oldest queue item is dropped at the bound, the item is dropped at the attempt limit, and each outcome is one log row with no payload text for non-queued outcomes.
- **Verify**: `go test ./internal/feedback/outbox/ -run '^(TestDedupeWindow|TestGlobalCaps|TestQueueBound|TestAttemptsIncrementedAndCapped)$' -count=1 -v` — expect exit 0.
- **RED-now**: E18. **Green path**: M4.

### AC-ANON-016 — Preview and outbox log
Maps REQ-ANON-014
- **Given** two queued items,
- **When** `moai feedback participation preview` runs with a network stub and the sender then runs against a recording `gh` stub,
- **Then** the preview bytes equal, per item, the bytes the stub received; the preview made zero network requests; the outbox log is JSONL, append-only across runs, mode 0600, and holds the exact payload for queued, sent, and withheld rows.
- **Verify**: `go test ./internal/cli/ -run '^(TestParticipationPreview|TestOutboxLogAppendOnly0600)$' -count=1 -v` — expect exit 0.
- **RED-now**: E13. **Green path**: M4.

### AC-ANON-017 — Sender discipline
Maps REQ-ANON-015
- **Given** a queued item and a `gh` runner stub,
- **When** the sender runs with participation true, then with participation flipped to false between two items, then with `gh` missing, then with `gh` unauthenticated, then from inside a hook dispatch,
- **Then** the first sends; the flip skips every remaining item; missing and unauthenticated `gh` leave the item queued with no prompt and no error to the caller; a hook-dispatch call is refused; the whole run ends within the time box on a blocking stub.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestSenderChecksConsentPerItem|TestSenderQuietWithoutGh|TestSenderNeverRunsOnHookPath|TestSenderTimeBox)$' -count=1 -v` — expect exit 0.
- **RED-now**: E19. **Green path**: M5.

### AC-ANON-018 — Existing fingerprint: comment, no model
Maps REQ-ANON-016
- **Given** a `gh` runner stub whose issue search returns an open issue, then a closed issue, with the exact title key `[auto-report] panic <fp>`,
- **When** the sender processes a queued item of that fingerprint,
- **Then** it issues exactly one comment command carrying the occurrence marker, issues no create and no edit command, the counting model stub records zero calls, and a search result whose title differs from the exact key is ignored (a new issue is then created). Mutant that must die: a sender that edits the body.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestExistingFingerprintGetsOccurrenceComment|TestExactTitleKeyOnly|TestClosedIssueStillCounts)$' -count=1 -v` — expect exit 0.
- **RED-now**: E19. **Green path**: M5.

### AC-ANON-019 — New issue: one model call, validated, with fallback
Maps REQ-ANON-017
- **Given** a moai-verdict item with no existing issue,
- **When** the model stub returns (a) a clean summary, (b) a summary containing a path, (c) an error, (d) a timeout,
- **Then** (a) the issue is created with the summary section, exactly one model call; (b), (c), (d) create the issue with the deterministic template text instead; in all four the title and marker block are Go-generated, the model input equals the payload fields only (golden), and the model output passes the scrubber, the classifier, and the length and character allowlist before use.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestPublishCallsModelOnceAndValidates|TestPublishFallsBackToTemplate|TestModelInputIsPayloadFieldsOnly)$' -count=1 -v` — expect exit 0.
- **RED-now**: E19. **Green path**: M6.

### AC-ANON-020 — Model-call budget
Maps REQ-ANON-018
- **Given** a counting model stub across the whole pipeline,
- **When** the pipeline processes: participation off; a `user` verdict; an `environment` verdict; an `ambiguous` verdict under `local`; a locally deduplicated fingerprint; a locally capped fingerprint; a withheld payload; an existing remote issue; a daily-cap-exhausted state,
- **Then** every one of those records zero calls; the positive control (a new moai-verdict item, no remote issue) records exactly one; a counter that counts only the publish path would still pass the positive control, so the table asserts the stub at every pipeline entry point. Mutant that must die: a pipeline calling the model before the duplicate lookup.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestLLMBudgetZeroCalls|TestLLMBudgetPositiveControl|TestDailyModelCallCap)$' -count=1 -v` — expect exit 0.
- **RED-now**: E19. **Green path**: M6.

### AC-ANON-021 — Issue contract
Maps REQ-ANON-019
- **Given** a payload and the golden file `internal/feedback/testdata/bugreport_issue_v1.golden`,
- **When** the issue title, body, and occurrence comment are rendered and then parsed back,
- **Then** the rendering equals the golden bytes; parsing recovers every marker field; the same fingerprint renders the same title key; the create command carries no `--label`; no command edits an issue body.
- **Verify**: `go test ./internal/feedback/publish/ -run '^(TestIssueContractRoundTrip|TestSameFingerprintSameTitleKey|TestNoLabelsNoBodyEdit)$' -count=1 -v` — expect exit 0.
- **RED-now**: E15, E19. **Green path**: M5.

### AC-ANON-022 — Web toggle
Maps REQ-ANON-020
- **Given** the settings schema and the console assets,
- **When** the feedback panel is rendered and a form with the toggle off then on is submitted,
- **Then** `feedback.participation` renders as a checkbox (never a text input) with the four-locale title and description keys present, the description states public and account-tied in all four locales, the value reaches `feedback.yaml`, a value-identical save rewrites nothing, and `participation_asked` is not in the rendered fields.
- **Verify**: `go test ./internal/web/ ./internal/settings/ -run '^(TestFeedbackPanelRendered|TestFeedbackParticipationI18nKeysInAllLocales|TestFeedbackPanelFieldsWired|TestI18nKeySetParity|TestSchemaCurrentValuesReadsAllSections|TestFeedbackParticipationFieldIsBool)$' -count=1 -v` — expect exit 0.
- **RED-now**: E10, E11. **Green path**: M2.

### AC-ANON-023 — Withdrawal and purge
Maps REQ-ANON-021
- **Given** a queue, a spool, a ledger, and an outbox log with content,
- **When** a flush starts with participation false, and separately the purge command runs,
- **Then** the flush discards unsent items and the spool and appends `discarded` rows while keeping the sent history; purge removes all four stores; neither makes a network request or a model call.
- **Verify**: `go test ./internal/feedback/outbox/ ./internal/cli/ -run '^(TestWithdrawalDiscardsQueueOnFlush|TestParticipationPurge)$' -count=1 -v` — expect exit 0.
- **RED-now**: E14, E18. **Green path**: M4.

### AC-ANON-024 — Template-First mirrors and docs
Maps REQ-ANON-022
- **Given** the template tree, the three skill-body copies, and the four docs-site pages,
- **When** the build and the guard tests run,
- **Then** the template and local `feedback.yaml` match, `make build` exits 0, each skill-body copy mentions participation, and none of the four docs-site pages still says the issue is created automatically (the old phrasing per locale is the pattern in ledger entry E8); each page instead states that the issue is filed after the user's confirmation, or, for participation, automatically from the user's own account only when they opted in.
- **Verify**: the E8 command — expect four `path:0` lines and exit 1; plus `go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1` expect exit 0; plus `make build` expect exit 0.
- **RED-now**: E8, E16. **Green path**: M7.

### AC-ANON-025 — No auto-repair artifact ships
Maps REQ-ANON-023
- **Given** the embedded template tree and `plugins/moai/`,
- **When** the guard walks them, with a seeded canary file named `auto-repair-canary.md` added in a temporary overlay,
- **Then** the real trees pass and the canary run fails (the red is observed on the canary before the guard is adopted); no file path contains `auto-repair`, `autorepair`, or `auto_repair`, and no file content contains `moai-bugreport`.
- **Verify**: `go test ./internal/template/ -run '^(TestNoAutoRepairArtifactsShipped|TestNoAutoRepairGuardCatchesCanary)$' -count=1 -v` — expect exit 0.
- **RED-now**: E17. **Green path**: M7.

## Edge cases

- A user runs `moai update` twice in one terminal session: the second run does not ask (AC-ANON-005).
- Two users hit the same fingerprint at the same moment before the search index reflects the first issue: two issues are created; the consumer merges by fingerprint and sums comments. This approximate behaviour is accepted and recorded in plan.md risks.
- `gh` is installed but the user is logged into a different host: the sender leaves the item queued (AC-ANON-017).
- The participation toggle is turned off while items are queued: the next flush discards them (AC-ANON-023).

## Definition of Done

All 25 criteria pass with their named commands; `go vet` and `golangci-lint` clean on touched packages; the three guard families (shipped-key readers, i18n parity, template neutrality) pass; the amendment of the quiet-wizard pins is recorded in the commit message; progress.md §E.2 carries each criterion's observed output.
