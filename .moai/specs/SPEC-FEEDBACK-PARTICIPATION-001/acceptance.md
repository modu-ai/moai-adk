# Acceptance — SPEC-FEEDBACK-PARTICIPATION-001

> **Verify rule (every criterion).** A `go test` Verify line is run with `-v`. The criterion holds only when the exit code is 0 **and** a `--- PASS: <Name>` line appears for every test named in that line's `-run` selector (subtests aside; under `-count=N` at least one per name per run). A selector alternation silently skips a name that does not exist and still exits 0 with `ok`, so the missing PASS line is the failure, never the exit code. The equivalent machine form is `go test -json` with the number of distinct passed test names equal to the selector's arity. `grep` and `cmp` lines state their expected output and exit code. A `-run` selector that sweeps zero tests (`[no tests to run]`) is a failure of the criterion, never a pass. Local verification is package-scoped; the full suite is CI's. Tests that reach the user-scoped consent store set `MOAI_HOME` to a temporary directory with `t.Setenv`, so the real home directory is never touched.
>
> Verify lines are of two kinds. **New-test lines** name tests this SPEC creates (red now). **Regression-guard lines** name pins and guards that exist today and must still pass after the amendment (green now by design; their post-amendment assertions are the green path).

**Criteria: 25 of the Tier L ceiling 25. Requirements: 25 of 25.**

## §D AC Matrix

| AC | REQ | Milestone | Summary | RED-now |
|---|---|---|---|---|
| AC-001 | REQ-001, REQ-024 | M2 | User-scoped store: defaults false, tracked file ignored, repository default, no key shipped | E1, E2 |
| AC-002 | REQ-002 (capture half), REQ-008 | M3 | Participation off: capture is a no-op; capture is fail-open and bounded | E3 |
| AC-003 | REQ-002 (drain and sender half), REQ-024 | M5 | Off or tracked-only: no drain, no `gh` call, zero model calls, default repository | E4 |
| AC-004 | REQ-003 | M2 | Init question: one, default no, own group after Jev, absent from shared sets, persistence gate, pins amended | E5, E6 |
| AC-005 | REQ-004 | M2 | Update asks once on a plain template-sync run only; default no; flag inventory classified | E7a, E7b |
| AC-006 | REQ-005 | M2 | Consent text in four locales states every required fact | E5 |
| AC-007 | REQ-006, REQ-007 | M3 | Closed kind enum, registered sites and handler names, recover guard, exclusions | E3, E9, E10 |
| AC-008 | REQ-009 | M3 | Allowlist attribution table: every row, first match wins, unknown never `moai` | E3 |
| AC-009 | REQ-009 | M4 | Ambiguous verdict retained locally: nothing queued, zero model calls, retention logged | E12 |
| AC-010 | REQ-010 | M1 | Fingerprint stable; frames filtered by module path; generic frame shape pinned | E3 |
| AC-011 | REQ-011 | M1 | Payload is a closed schema; `detail` is closed-set membership in an enum type | E3 |
| AC-012 | REQ-012 | M4 | Scrub and classify tripwire; withheld on block, masking, or path traversal | E12 |
| AC-013 | REQ-013 | M4 | Per-fingerprint window, global caps, queue bound, attempt limit | E12 |
| AC-014 | REQ-014 | M4 | Preview is byte-identical to what is handed to `gh`; append-only log, mode 0600 | E13a, E13b |
| AC-015 | REQ-015 | M5 | Sender re-checks consent per item, off the hook path, time-boxed, quiet when `gh` is absent | E11 |
| AC-016 | REQ-016 | M5 | Existing fingerprint issue gets one occurrence comment, none at the cap; zero model calls | E11 |
| AC-017 | REQ-017 | M6 | New moai issue: one model call, validated output, template fallback | E11 |
| AC-018 | REQ-017, REQ-018 | M6 | Summary persisted before create; retries reuse it; per-item call bound across retries | E11 |
| AC-019 | REQ-018 | M6 | Model-call budget: zero calls in every excluded case, one in the positive control | E11 |
| AC-020 | REQ-019 | M5 | Issue title and marker contract round-trips; markers are untrusted; no body edit, no labels | E11, E23 |
| AC-021 | REQ-020 | M2 | Web toggle: existing radio pair, present-companion, user-scoped read and write, marker key not rendered | E14, E15 |
| AC-022 | REQ-021 | M4 | Withdrawal discards unsent items; purge removes all local state; no network | E12, E17 |
| AC-023 | REQ-022 | M7 | Skill-body copies, docs-site wording with a positive per-locale pattern, mirror equality, pins | E18, E19, E20 |
| AC-024 | REQ-023 | M7 | No auto-repair artifact ships; the wizard question never ships without the sender | E22a, E22b |
| AC-025 | REQ-025 | M6 | Static reachability guard: import allowlists and the injected model seam | E3, E11 |

All criteria are release-blocking. The tree pin for every RED-now observation below is the commit `bb54f2903` (measured in this run, in this worktree, on a clean tree; `git diff --stat 2f492df19 HEAD` lists only the seven SPEC files, so the code tree is identical to the earlier pin `2f492df19`).

### Renumbering map (version 0.2.0 to 0.3.0)

Merged to stay inside the 25-criterion ceiling: the capture half of old AC-002 and the capture fail-open criterion (old AC-009) became new AC-002; the init question and persistence criteria (old AC-003 and AC-004) became new AC-004; the kind enum and exclusions criteria (old AC-007 and AC-008) became new AC-007. New: AC-003 (the off-means-nothing half owned by the drain and the sender), AC-018 (retry bound), AC-025 (static guard).

| Old | New | Old | New | Old | New |
|---|---|---|---|---|---|
| AC-001 | AC-001 (rewritten) | AC-010 | AC-008 | AC-018 | AC-016 |
| AC-002 | AC-002 and AC-003 (split) | AC-011 | AC-009 | AC-019 | AC-017 |
| AC-003 | AC-004 (merged) | AC-012 | AC-010 | AC-020 | AC-019 |
| AC-004 | AC-004 (merged) | AC-013 | AC-011 | AC-021 | AC-020 |
| AC-005 | AC-005 | AC-014 | AC-012 | AC-022 | AC-021 |
| AC-006 | AC-006 | AC-015 | AC-013 | AC-023 | AC-022 |
| AC-007 | AC-007 (merged) | AC-016 | AC-014 | AC-024 | AC-023 |
| AC-008 | AC-007 (merged) | AC-017 | AC-015 | AC-025 | AC-024 |
| AC-009 | AC-002 (merged) | | | (new) | AC-018, AC-025 |

## RED-now evidence ledger

Each entry: the single-invocation command, its verbatim stdout (raw bytes as observed; the absolute path shown by `go test` is this worktree's), its exit code as its own field, and the pinned tree `bb54f2903`. An empty stdout with exit 1 is a complete observation.

**RED-now proxy rule.** For a test that does not exist yet in a package that already exists, the criterion's own `-run` command exits 0 and prints `[no tests to run]` (E1, E5, E14, E15); that reads as red only through the PASS-line rule of the preface, because no `--- PASS` line appears. For the `internal/cli` package, whose test binary is the largest in the module, the ledger records a grep-for-name proxy instead (E6, E7a, E7b, E9, E13a, E13b, E17, E22a, E22b): empty stdout with exit 1 states that the named test does not exist, which is the same fact as a missing PASS line. For a package that does not exist, `go test` itself fails with `[setup failed]` (E3, E4, E11, E12).

**Observation caveat.** The Bash tool did not surface a non-zero exit status for the empty-output `grep` commands, so every exit code below was read in a separate harness invocation that ran the same command and then printed `rc=$?`; the cited commands themselves are single invocations. Ids skip E8 and E16, which belonged to the 0.2.0 numbering and are retired; E21 is green today by design.

```
E1
tree: bb54f2903
command: go test ./internal/config/ -run '^(TestUserParticipationDefaultsOff|TestUserParticipationIgnoresTrackedProjectFile|TestUserParticipationRepositoryDefault)$' -count=1 -v
exit: 0
stdout:
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.310s [no tests to run]
why red: the three named tests do not exist; no `--- PASS` line appears, so the criterion fails by the preface rule (the exit code alone is 0, which is the vacuous-selector trap)
```

```
E2
tree: bb54f2903
command: grep -c participation internal/template/templates/.moai/config/sections/feedback.yaml .moai/config/sections/feedback.yaml
exit: 1
stdout:
internal/template/templates/.moai/config/sections/feedback.yaml:0
.moai/config/sections/feedback.yaml:0
note: an absence guard, green today by design (no consent key may ship); its mutant is a template that ships the key. Positive control E2p shows the instrument can fire.

E2p
tree: working tree atop HEAD `c34e24cab` — the v0.4.0 revision, uncommitted at observation time (re-measured in this pass after the 0.4.0 rename)
command: grep -c participation .moai/specs/SPEC-FEEDBACK-PARTICIPATION-001/spec.md
exit: 0
stdout:
21
note: re-measured on the renamed path; the same instrument observed earlier on the committed v0.2.0 spec.md at the pre-rename path printed 16 with exit 0. A non-zero count proves the same grep reports a hit when the word is present.
```

```
E3
tree: bb54f2903
command: go test ./internal/bugreport/ -run '^(TestCaptureNoopWhenParticipationOff|TestCaptureIgnoresTrackedFileConsent)$' -count=1 -v
exit: 1
stdout:
# ./internal/bugreport
stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1498/internal/bugreport: directory not found
FAIL	./internal/bugreport [setup failed]
FAIL
why red: the package does not exist; every `internal/bugreport` selector fails identically, so this entry stands for AC-002, AC-007, AC-008, AC-010, AC-011, and the bugreport half of AC-025
```

```
E4
tree: bb54f2903
command: go test ./internal/feedback/outbox/ ./internal/feedback/publish/ -run '^(TestDrainNoopWhenParticipationOff|TestSenderNoopWhenParticipationOff|TestSenderIgnoresTrackedFileConsentAndRepository)$' -count=1 -v
exit: 1
stdout:
# ./internal/feedback/outbox
stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1498/internal/feedback/outbox: directory not found
FAIL	./internal/feedback/outbox [setup failed]
# ./internal/feedback/publish
stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1498/internal/feedback/publish: directory not found
FAIL	./internal/feedback/publish [setup failed]
FAIL
why red: both packages do not exist
```

```
E5
tree: bb54f2903
command: go test ./internal/cli/wizard/ -run '^(TestParticipationQuestion|TestParticipationQuestion_FourLocalesStateFacts)$' -count=1 -v
exit: 0
stdout:
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	0.409s [no tests to run]
why red: the two named tests do not exist; no `--- PASS` line appears
```

```
E6
tree: bb54f2903
command: grep -rl TestApplyParticipationFromWizard internal/cli
exit: 1
stdout:
(empty)
why red: the init persistence test (and its seam) does not exist

E7a
tree: bb54f2903
command: grep -rl TestUpdateParticipationStep internal/cli
exit: 1
stdout:
(empty)
why red: the update step test does not exist

E7b
tree: bb54f2903
command: grep -rl TestUpdateFlagInventoryClassified internal/cli
exit: 1
stdout:
(empty)
why red: the update flag inventory test does not exist
```

```
E9
tree: bb54f2903
command: grep -rl TestEveryRegisteredHookHandlerHasBugreportName internal/cli
exit: 1
stdout:
(empty)
why red: the handler-name registry guard does not exist

E10
tree: bb54f2903
command: grep -rl internal/bugreport internal cmd
exit: 1
stdout:
(empty)
why red: no caller imports the capture package
```

```
E11
tree: bb54f2903
command: go test ./internal/feedback/publish/ -run '^(TestAmbiguousPolicyLocal|TestAmbiguousPolicyAdjudicate)$' -count=1 -v
exit: 1
stdout:
# ./internal/feedback/publish
stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1498/internal/feedback/publish: directory not found
FAIL	./internal/feedback/publish [setup failed]
FAIL
why red: the package does not exist; every `internal/feedback/publish` selector fails identically, so this entry stands for AC-015 to AC-020, and the publish half of AC-025. Note: the 0.4.0 respecification moved AC-009 to the outbox package (E12) and retired the two policy test names in this command; the red (the package is absent) is identical for any selector.

E12
tree: bb54f2903
command: go test ./internal/feedback/outbox/ -run '^(TestOutboundPayloadScrubbedAndClassified|TestBlockedOrMaskedPayloadWithheld|TestPathTraversalKindWithheld)$' -count=1 -v
exit: 1
stdout:
# ./internal/feedback/outbox
stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1498/internal/feedback/outbox: directory not found
FAIL	./internal/feedback/outbox [setup failed]
FAIL
why red: the package does not exist; stands for AC-009, AC-012, AC-013, and the outbox half of AC-022
```

```
E13a
tree: bb54f2903
command: grep -rl TestParticipationPreview internal/cli
exit: 1
stdout:
(empty)
why red: the preview test does not exist

E13b
tree: bb54f2903
command: grep -rl TestOutboxLogAppendOnly0600 internal/cli
exit: 1
stdout:
(empty)
why red: the outbox log test does not exist
```

```
E14
tree: bb54f2903
command: go test ./internal/web/ -run '^(TestFeedbackParticipationI18nKeysInAllLocales|TestFeedbackParticipationRendersAsRadioPair)$' -count=1 -v
exit: 0
stdout:
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.568s [no tests to run]
why red: the two named tests do not exist; no `--- PASS` line appears

E15
tree: bb54f2903
command: go test ./internal/settings/ -run '^(TestFeedbackParticipationFieldIsBool|TestUserScopedEditWritesHomeFileOnly|TestUserScopedValueInvariantTouchesNothing|TestUserScopedWriteSetsAsked)$' -count=1 -v
exit: 0
stdout:
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/settings	0.324s [no tests to run]
why red: the four named tests do not exist; no `--- PASS` line appears
```

```
E17
tree: bb54f2903
command: grep -rl TestParticipationPurge internal/cli
exit: 1
stdout:
(empty)
why red: the purge test does not exist
```

```
E18
tree: bb54f2903
command: grep -c -E "issue is created automatically|auto-creates a GitHub issue|GitHub 이슈로 자동 생성|GitHub Issue として自動作成|自动创建为 GitHub Issue" docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md
exit: 0
stdout:
docs-site/content/en/utility-commands/moai-feedback.md:2
docs-site/content/ko/utility-commands/moai-feedback.md:1
docs-site/content/ja/utility-commands/moai-feedback.md:1
docs-site/content/zh/utility-commands/moai-feedback.md:1
why red: the "created automatically" wording is still present in all four locales; the criterion expects four `:0` counts
```

```
E19en
tree: bb54f2903
command: grep -c -F "GitHub issue only after you confirm" docs-site/content/en/utility-commands/moai-feedback.md
exit: 1
stdout:
0

E19ko
tree: bb54f2903
command: grep -c -F "확인한 뒤에만 GitHub 이슈" docs-site/content/ko/utility-commands/moai-feedback.md
exit: 1
stdout:
0

E19ja
tree: bb54f2903
command: grep -c -F "確認した後にのみ GitHub Issue" docs-site/content/ja/utility-commands/moai-feedback.md
exit: 1
stdout:
0

E19zh
tree: bb54f2903
command: grep -c -F "确认后才会创建 GitHub Issue" docs-site/content/zh/utility-commands/moai-feedback.md
exit: 1
stdout:
0
why red (all four): the corrected wording does not exist yet
```

```
E20
tree: bb54f2903
command: grep -c participation .claude/skills/moai/workflows/feedback.md plugins/moai/skills/moai/workflows/feedback.md internal/template/templates/.claude/skills/moai/workflows/feedback.md
exit: 1
stdout:
.claude/skills/moai/workflows/feedback.md:0
plugins/moai/skills/moai/workflows/feedback.md:0
internal/template/templates/.claude/skills/moai/workflows/feedback.md:0
why red: all three skill-body copies are silent on participation; the criterion expects each count to be at least 1
```

```
E21
tree: bb54f2903
command: cmp internal/template/templates/.moai/config/sections/feedback.yaml .moai/config/sections/feedback.yaml
exit: 0
stdout:
(empty)
note: the mirror guard is green today by design; its mutant is editing one copy only. The criterion's red comes from E18 to E20, not from this line.
```

```
E22a
tree: bb54f2903
command: grep -rl TestNoAutoRepairArtifactsShipped internal/template
exit: 1
stdout:
(empty)
why red: the shipping guard does not exist

E22b
tree: bb54f2903
command: grep -rl TestParticipationQuestionRequiresSender internal/cli/wizard
exit: 1
stdout:
(empty)
why red: the coexistence guard does not exist
```

```
E23
tree: bb54f2903
command: ls internal/feedback/testdata/bugreport_issue_v1.golden
exit: 1
stdout:
(empty; stderr: ls: internal/feedback/testdata/bugreport_issue_v1.golden: No such file or directory)
why red: the golden file does not exist
```

Mutant-probe note: each criterion below states the mutant that must die where one is writable. A criterion whose mutant survives is too shallow to adopt. The mutants probed in this revision: a reader that falls back to the project file (AC-001, AC-003); a template that ships the consent key (AC-001); `moai` as a default verdict, an unmarked hook handler error attributed to moai, an `*exec.ExitError` or JSON decode error attributed to moai (AC-008); a builder that accepts a string `detail` with a loose pattern (AC-011); a scrub that never runs (AC-012); a sender that edits the body or comments past the cap (AC-016); a pipeline that calls the model on every retry or before the lookup (AC-018, AC-019); a parser that trusts marker fields (AC-020); a checkbox widget, a rendered `asked` key, or a console write that leaves `asked` false (AC-021); a model call outside the seam (AC-025).

## Scenarios

### AC-001 — User-scoped store, defaults, authority, nothing shipped
Maps REQ-ANON-001, REQ-ANON-024
- **Given** a temporary `MOAI_HOME` holding no participation file, and a project tree whose tracked `.moai/config/sections/feedback.yaml` carries `participation: true`, `participation_asked: true`, and `repository: attacker/x`,
- **When** the user-scoped reader and the sender-facing repository accessor are called, then called again with a user file holding `participation.enabled: true` and no repository key, and with a directory in place of the file and with a malformed file,
- **Then** `enabled` and `asked` read false and the repository reads `modu-ai/moai-adk` in the first case; in the second `enabled` reads true and the repository still reads `modu-ai/moai-adk`; in the directory and malformed cases every value reads false; and the shipped template `feedback.yaml` and its local mirror contain no `participation` text. Mutants that must die: a reader that falls back to the project file; a template that ships `participation: true`; a repository taken from the project tier.
- **Verify (new-test line)**: `go test ./internal/config/ -run '^(TestUserParticipationDefaultsOff|TestUserParticipationIgnoresTrackedProjectFile|TestUserParticipationRepositoryDefault)$' -count=1 -v` — expect exit 0 and a `--- PASS` line for each of the three names.
- **Verify (absence guard)**: `grep -c participation internal/template/templates/.moai/config/sections/feedback.yaml .moai/config/sections/feedback.yaml` — expect two `path:0` lines and exit 1.
- **Verify (regression guard)**: `go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -count=1 -v` — expect exit 0 and `--- PASS: TestShippedConfigKeysHaveReaders ` (the name followed by a space).
- **RED-now**: E1 (new tests absent), E2 with its positive control E2p. **Green path**: M2 adds the reader, its tests, and keeps the template files key-free.

### AC-002 — Capture is a no-op when off, and fail-open
Maps REQ-ANON-002, REQ-ANON-008 (the capture half of REQ-ANON-002; the drain and sender half is AC-003)
- **Given** a temporary `MOAI_HOME` with no file, then with a file saying false, then with only a tracked project file saying true, a recording filesystem, a recording network stub, and a counting model stub; and separately an unwritable spool directory, a full spool, and a filesystem stub that blocks,
- **When** the capture entry point is called with a panic signal, and under `-race` in the fail-open conditions,
- **Then** in the three consent states no spool file is created, the network stub records zero requests, and the model stub records zero calls; in the fail-open conditions capture returns without error and without panic, returns within the time box on the blocking stub, drops the signal when the spool is full, and the network stub records zero requests.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^(TestCaptureNoopWhenParticipationOff|TestCaptureIgnoresTrackedFileConsent)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^TestCaptureNeverPanicsOrBlocks$' -race -count=3 -v` — expect exit 0 and a PASS line for the name.
- **RED-now**: E3. **Green path**: M3.

### AC-003 — Off means no drain, no gh, no model, and the default repository
Maps REQ-ANON-002, REQ-ANON-024 (the drain and sender half of REQ-ANON-002; the capture half is AC-002)
- **Given** a queued item, a recording `gh` stub, a counting model stub, and a spool, with (a) no user-scoped file and a tracked project file saying `participation: true` and `repository: attacker/x`, (b) a user file saying false, and (c) a user file saying true with no repository key and the same tracked file,
- **When** the drain and the sender run,
- **Then** in (a) and (b) the drain reads and writes nothing, the `gh` stub records zero calls, and the model stub records zero calls; in (c) the `gh` stub's `--repo` argument is `modu-ai/moai-adk`, never `attacker/x`. Mutants that must die: a sender that trusts the tracked key; a sender that targets the project-tier repository; a drain that runs when off.
- **Verify (new-test line)**: `go test ./internal/feedback/outbox/ ./internal/feedback/publish/ -run '^(TestDrainNoopWhenParticipationOff|TestSenderNoopWhenParticipationOff|TestSenderIgnoresTrackedFileConsentAndRepository)$' -count=1 -v` — expect exit 0 and a PASS line for each of the three names.
- **RED-now**: E4. **Green path**: M5 (the drain test lands in M4; the criterion closes when the sender exists).

### AC-004 — Init question slot, persistence gate, pins
Maps REQ-ANON-003
- **Given** the init question set and a project root,
- **When** it is built, and `applyParticipationFromWizard` runs with (a) the wizard not run, (b) the wizard run and `CI` empty, (c) the wizard run and `CI` set non-empty, (d) a nil result, and with a stored user value of true,
- **Then** exactly one question has id `feedback_participation`, type confirm, its own group, positioned after `jev_enabled`, absent from `DefaultQuestions` and `ReconfigureQuestions`, defaulting to false and to true when the stored value is true; the removed-question set still contains `feedback_auto_submit` and not `feedback_participation`; (a), (c), (d) leave the user-scoped file byte-identical; (b) writes `participation.enabled` and `participation.asked` through the writer, and a decline over an existing true writes false; and every pin test the question affects passes with the new count and membership.
- **Verify (new-test line)**: `go test ./internal/cli/wizard/ -run '^TestParticipationQuestion$' -count=1 -v` — expect exit 0 and `--- PASS: TestParticipationQuestion ` (the name followed by a space, so a longer name sharing the prefix does not satisfy it).
- **Verify (new-test line)**: `go test ./internal/cli/ -run '^TestApplyParticipationFromWizard$' -count=1 -v` — expect exit 0, `--- PASS: TestApplyParticipationFromWizard ` (the name followed by a space), and four passing subtests.
- **Verify (regression guard, amended pins)**: `go test ./internal/cli/wizard/ -run '^(TestInitQuestions_QuietSet|TestRemovedQuestionsAbsentFromInitSet|TestInitStepper_Denominator4|TestInitRegroup_SecondGroupGolden|TestTotalVisibleQuestions_Page3AlwaysCounted|TestProfileWizardStepper_SameFormatAsInit|TestStepperTotal_DynamicDenominator|TestInitPages_Membership|TestInitPages_MergeIntoOneGroupPerPage)$' -count=1 -v` — expect exit 0 and a PASS line for each of the nine names.
- **Verify (regression guard, whole package)**: `go test ./internal/cli/wizard/ -count=1` — expect exit 0 and `ok`; this runs the remaining files of the 20-file inventory in plan.md section C (M2).
- **RED-now**: E5, E6. **Green path**: M2 (the amended pins land in the same commit as the question; the commit message records the amendment). The `internal/cli` tests of the inventory (`agent_model_flags_retired_test.go`, `ptycap_child_test.go`, `init_agent_wizard_test.go`, `init_update_profile_path_test.go`) run with the whole-package CI run and, locally, under the repository's slot lease.

### AC-005 — Update asks once, on a plain template-sync run only
Maps REQ-ANON-004
- **Given** an interactive-terminal seam (`isInteractiveStdin` stubbed true), `participation.asked` false in a temporary `MOAI_HOME`, and a prompt-call counter,
- **When** the update step runs, then runs again, and then runs under each of: stdin not a terminal, `CI` non-empty, and each mode flag (`--check`, `--shell-env`, `--config`, `--binary`, `--dry-run`, `--restore`, `--version`, `--templates-only`, `--yes`); and separately under each modifier flag (`--force`, `--no-hooks`, `--no-plugin`, `--verbose`, `--profile`),
- **Then** the first run prompts once, with a default of no, and persists both values; an empty answer persists `enabled` false and `asked` true; the second run does not prompt; every listed mode flag makes zero prompt calls and leaves the file unchanged; each modifier flag still prompts once; `--yes` never sets `enabled` true; the prompt text equals the init question text per locale; and a second test enumerates every flag of the update command and fails when a flag is in neither the mode list nor the modifier list. Mutants that must die: a prompt that defaults to yes; an empty answer that persists true; a step that asks under `--templates-only` or `--yes`.
- **Verify (new-test line)**: `go test ./internal/cli/ -run '^(TestUpdateParticipationStep|TestUpdateFlagInventoryClassified)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E7a, E7b. **Green path**: M2.

### AC-006 — Consent text states every required fact
Maps REQ-ANON-005
- **Given** the question translations for en, ko, ja, zh,
- **When** each locale's title and description are read,
- **Then** each states that the issue is public and tied to the user's GitHub account, that the issue and comment creation time is public, that filing is automatic with no per-report confirmation, that GitHub subscribes the filer to later comments, that the account becomes publicly associated with using moai-adk at that version and operating system, the fixed fields, `moai feedback participation preview` and `moai web`, that the tool cannot recall a filed issue, and that the summary may spend the user's own model-subscription tokens with a template text when no model is available.
- **Verify (new-test line)**: `go test ./internal/cli/wizard/ -run '^TestParticipationQuestion_FourLocalesStateFacts$' -count=1 -v` — expect exit 0 and a PASS line for the name.
- **Verify (regression guard)**: `go test ./internal/cli/wizard/ -run '^TestWizardQuestionTranslationCompleteness$' -count=1 -v` — expect exit 0 and a PASS line for the name.
- **RED-now**: E5. **Green path**: M2.

### AC-007 — Kind enum, sites, handler names, recover guard, exclusions
Maps REQ-ANON-006, REQ-ANON-007
- **Given** the emit-site register, the production hook wiring, and the source tree,
- **When** the guard walks every non-test `recover()` in `internal/`, `cmd/`, and `pkg/`, and synthetic signals from user-config errors, a tool-failure event, and a user test-run failure are offered to capture,
- **Then** each recover site either calls the capture entry point or is on the reasoned allowlist; each registered site file contains its capture call; the kind enum has exactly six members, each with a verdict, a derivation mode, and at least one registered site, and the `harness_defect` token set is exactly `missing_key`, `unexpanded_token`, `invalid_json`; every handler registered in the production wiring has its name in the registered table and the table has no extra names; none of the synthetic signals is recorded; and the files `internal/hook/post_tool_failure.go` and `internal/hook/failure_observer.go` contain no reference to the capture package.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^(TestKindEnumClosed|TestEmitSiteRegister|TestEveryRecoverSiteReportsOrIsAllowlisted|TestUserAndToolFailuresNeverCaptured|TestToolFailureHandlersDoNotImportCapture)$' -count=1 -v` — expect exit 0 and a PASS line for each of the five names.
- **Verify (new-test line)**: `go test ./internal/cli/ -run '^TestEveryRegisteredHookHandlerHasBugreportName$' -count=1 -v` — expect exit 0 and a PASS line for the name.
- **RED-now**: E3, E9, E10. **Green path**: M3.

### AC-008 — Allowlist attribution
Maps REQ-ANON-009
- **Given** a table of signals covering every row of the attribution register in `design.md` section 3: a panic; an internal-marker wrapper; a marker wrapping an `*fs.PathError`; filesystem, syscall, network, `*exec.Error`, `*exec.ExitError`, and cancellation errors; config sentinels; `*json.SyntaxError` and `*json.UnmarshalTypeError`; `*yaml.TypeError`; a plain `errors.New` hook handler error with no marker; the hook timeout; each template and harness token; and an arbitrary custom error type,
- **When** attribution runs on each,
- **Then** the verdict equals the table's verdict: panic, marker, and the tokens `not_found`, `preserve_integrity`, `path_traversal`, `missing_key` give `moai`; the marker wrapping an `*fs.PathError`, filesystem, syscall, network, exec, exec-exit, and cancellation errors give `environment`; config sentinels, JSON errors, and YAML type errors give `user`; the unmarked handler error, the hook timeout, the tokens `unexpanded_token` and `invalid_json`, and the custom error type give `ambiguous`, never `moai`; an error chain matching both an environment row and a moai row yields `environment`; every register row declares its derivation mode (derived or call-site asserted); and no row calls a model (a counting stub records zero). Mutants that must die: `moai` as the default verdict; an unmarked handler error, an `*exec.ExitError`, or a JSON decode error attributed to moai.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^(TestAttributionRulesTable|TestAttributionFirstMatchWins|TestAttributionMakesNoModelCall|TestUnknownErrorNeverAttributedToMoai|TestRegisterRowsDeclareDerivationMode)$' -count=1 -v` — expect exit 0 and a PASS line for each of the five names.
- **RED-now**: E3. **Green path**: M3.

### AC-009 — Ambiguous verdict retained locally
Maps REQ-ANON-009
- **Given** captured signals whose attribution is `ambiguous` (an unmarked hook handler error, a hook timeout, and a template token `unexpanded_token` or `invalid_json`), a counting model stub, a recording network stub, and an empty outbox log,
- **When** the drain processes each signal end to end,
- **Then** each stays local: nothing is queued, the model stub records zero calls, the network stub records zero requests, and one outbox log row names the retention reason; no code path or constant selects a model call for an ambiguous verdict. Mutants that must die: an adjudication call on an ambiguous item; an ambiguous item enqueued.
- **Verify (new-test line)**: `go test ./internal/feedback/outbox/ -run '^(TestAmbiguousRetainedLocally|TestAmbiguousMakesNoModelCall)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E12. **Green path**: M4.

### AC-010 — Fingerprint
Maps REQ-ANON-010
- **Given** captured stacks that include moai frames, standard-library frames, dependency frames, a generic-function frame, and file paths under a user home directory,
- **When** the fingerprint and frame list are computed,
- **Then** the frame list contains only function names beginning with `github.com/modu-ai/moai-adk/` with the prefix stripped, no `/`-path, line number, or argument text; identical inputs give identical fingerprints; changing version, commit, os, arch, kind, or any frame changes the fingerprint; the generic-function frame name has the shape the test records; a panic with zero moai frames is kept local.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^(TestFingerprintStable|TestFingerprintInputsAllMatter|TestFramesExcludePathsLinesAndForeignModules|TestZeroFramePanicStaysLocal|TestGenericFunctionFrameNameShape)$' -count=1 -v` — expect exit 0 and a PASS line for each of the five names.
- **RED-now**: E3. **Green path**: M1.

### AC-011 — Payload is a closed schema; `detail` is closed-set membership
Maps REQ-ANON-011
- **Given** a panic whose value is `CANARY-/Users/leak/secret-token`, an error whose text embeds a path, and the strings `CANARY-/Users/leak/secret-token`, `WorktreeCreate//srv/customer/private-project`, `/etc/passwd`, and the empty string,
- **When** the payload is built and serialised, and each string is offered directly to the `detail` registration and read-back validators,
- **Then** the bytes contain neither the canary nor any path; unknown fields and values failing an anchored allowlist are rejected; every listed string is rejected by the `detail` validators; a template token is rejected for a kind its register row does not name; a hook `detail` accepts only a registered event paired with a registered handler name; the builder's signature takes no `error`, `any`, or `string` for `detail` (checked by a type-level test), so `detail` is an enum-like type. Mutants that must die: an all-accepting validator; a builder that accepts a string `detail` with a loose character pattern.
- **Verify (new-test line)**: `go test ./internal/bugreport/ -run '^(TestPayloadSchemaClosed|TestPanicValueAndErrorTextNeverEnterPayload|TestBuilderTakesNoErrorOrAny|TestDetailIsClosedSetMembership|TestDetailRejectsCanaryAndPathShapedStrings|TestDetailTypeIsNotString)$' -count=1 -v` — expect exit 0 and a PASS line for each of the six names.
- **RED-now**: E3. **Green path**: M1.

### AC-012 — Scrub and classify tripwire
Maps REQ-ANON-012
- **Given** a payload whose frame list is mutated to contain a secret-pattern string, another containing a vulnerability phrase, and a report derived from a path-traversal token,
- **When** the pipeline renders and screens each,
- **Then** each is withheld: not queued, not handed to the publication seam, zero model calls, one `withheld` row in the outbox log; a clean payload passes with zero masking findings. Mutant that must die: a pipeline that never runs the scrubber.
- **Verify (new-test line)**: `go test ./internal/feedback/outbox/ -run '^(TestOutboundPayloadScrubbedAndClassified|TestBlockedOrMaskedPayloadWithheld|TestPathTraversalKindWithheld)$' -count=1 -v` — expect exit 0 and a PASS line for each of the three names.
- **RED-now**: E12. **Green path**: M4.

### AC-013 — Dedupe, caps, bounds
Maps REQ-ANON-013
- **Given** a controllable clock and the constants from `internal/config/defaults.go`,
- **When** the same fingerprint is offered twice inside and once outside the window, more reports than the daily and weekly caps are offered, the queue is filled past its bound, and an item fails to its attempt limit,
- **Then** the repeat inside the window is deduplicated, the cap overflow is capped, the oldest queue item is dropped at the bound, the item is dropped at the attempt limit, and each outcome is one log row with no payload text for non-queued outcomes.
- **Verify (new-test line)**: `go test ./internal/feedback/outbox/ -run '^(TestDedupeWindow|TestGlobalCaps|TestQueueBound|TestAttemptsIncrementedAndCapped)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **RED-now**: E12. **Green path**: M4.

### AC-014 — Preview and outbox log
Maps REQ-ANON-014
- **Given** two queued items,
- **When** `moai feedback participation preview` runs with a network stub and the sender then runs against a recording `gh` stub,
- **Then** the preview bytes equal, per item, the bytes the stub received; the preview made zero network requests; the outbox log is JSONL, append-only across runs, mode 0600, and holds the exact payload for queued, sent, and withheld rows.
- **Verify (new-test line)**: `go test ./internal/cli/ -run '^(TestParticipationPreview|TestOutboxLogAppendOnly0600)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E13a, E13b. **Green path**: M4.

### AC-015 — Sender discipline
Maps REQ-ANON-015
- **Given** a queued item and a `gh` runner stub,
- **When** the sender runs with participation true, then with participation flipped to false between two items, then with `gh` missing, then with `gh` unauthenticated, then from inside a hook dispatch,
- **Then** the first sends; the flip skips every remaining item; missing and unauthenticated `gh` leave the item queued with no prompt and no error to the caller; a hook-dispatch call is refused; the whole run ends within the time box on a blocking stub.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestSenderChecksConsentPerItem|TestSenderQuietWithoutGh|TestSenderNeverRunsOnHookPath|TestSenderTimeBox)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **RED-now**: E11. **Green path**: M5.

### AC-016 — Existing fingerprint: one comment, none at the cap, no model
Maps REQ-ANON-016
- **Given** a `gh` runner stub whose issue search returns an open issue, then a closed issue, with the exact title key `[auto-report] panic <fp>`, once with fewer than the per-issue cap of marker comments and once with the cap reached, read from the stub's `comments` JSON,
- **When** the sender processes a queued item of that fingerprint,
- **Then** below the cap it issues exactly one comment command carrying the occurrence marker; at the cap it issues no comment command; in both it issues no create and no edit command, the counting model stub records zero calls, and a search result whose title differs from the exact key is ignored (a new issue is then created). Mutants that must die: a sender that edits the body; a sender that comments past the cap.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestExistingFingerprintGetsOccurrenceComment|TestExactTitleKeyOnly|TestClosedIssueStillCounts|TestOccurrenceCapSkipsComment)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **RED-now**: E11. **Green path**: M5.

### AC-017 — New issue: one model call, validated, with fallback
Maps REQ-ANON-017
- **Given** a moai-verdict item with no existing issue,
- **When** the model stub returns (a) a clean summary, (b) a summary containing a path, (c) an error, (d) a timeout,
- **Then** (a) the issue is created with the summary section, exactly one model call; (b), (c), (d) create the issue with the deterministic template text instead; in all four the title and marker block are Go-generated, the model input equals the payload fields only (golden), and the model output passes the scrubber, the classifier, and the length and character allowlist before use.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestPublishCallsModelOnceAndValidates|TestPublishFallsBackToTemplate|TestModelInputIsPayloadFieldsOnly)$' -count=1 -v` — expect exit 0 and a PASS line for each of the three names.
- **RED-now**: E11. **Green path**: M6.

### AC-018 — Summary persisted; retries reuse it; per-item bound
Maps REQ-ANON-017, REQ-ANON-018
- **Given** a moai-verdict item with no existing issue, a counting model stub, and a `gh` stub whose `create` fails twice and then succeeds,
- **When** the sender processes the item across the three attempts,
- **Then** the model stub records exactly one summary call in total (the second and third attempts read the stored summary from the queue item); and when the remote lookup on a retry finds that an issue now exists, a comment is added and the stored summary is unused with no new model call. Mutants that must die: a sender that calls the model on every attempt; a sender that stores the summary after create instead of before.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestSummaryPersistedBeforeCreateAndReusedOnRetry|TestModelCallBoundPerQueueItem)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E11. **Green path**: M6.

### AC-019 — Model-call budget
Maps REQ-ANON-018
- **Given** a counting model stub across the whole pipeline,
- **When** the pipeline processes: participation off; a `user` verdict; an `environment` verdict; an `ambiguous` verdict; a locally deduplicated fingerprint; a locally capped fingerprint; a withheld payload; an existing remote issue; a daily-cap-exhausted state,
- **Then** every one of those records zero calls; the positive control (a new moai-verdict item, no remote issue) records exactly one; a counter that counts only the publish path would still pass the positive control, so the table asserts the stub at every pipeline entry point. Mutant that must die: a pipeline calling the model before the duplicate lookup.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestLLMBudgetZeroCalls|TestLLMBudgetPositiveControl|TestDailyModelCallCap)$' -count=1 -v` — expect exit 0 and a PASS line for each of the three names.
- **RED-now**: E11. **Green path**: M6.

### AC-020 — Issue contract and untrusted markers
Maps REQ-ANON-019
- **Given** a payload, the golden file `internal/feedback/testdata/bugreport_issue_v1.golden`, and a forged occurrence comment whose marker carries a fingerprint, version, and `os_arch` that disagree with the issue's title key,
- **When** the issue title, body, and occurrence comment are rendered and then parsed back, and the forged comment is parsed alongside the title key,
- **Then** the rendering equals the golden bytes; parsing recovers every marker field; the same fingerprint renders the same title key; the create command carries no `--label`; no command edits an issue body; and for the forged comment the parser keeps the fields derived from the title key and the body block and ignores the disagreeing marker fields while still counting the comment toward the advisory total.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestIssueContractRoundTrip|TestSameFingerprintSameTitleKey|TestNoLabelsNoBodyEdit|TestOccurrenceMarkersAreUntrustedInput)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **RED-now**: E11, E23. **Green path**: M5.

### AC-021 — Web toggle on the user-scoped value
Maps REQ-ANON-020
- **Given** the settings schema, the console assets, and a temporary `MOAI_HOME`,
- **When** the feedback panel is rendered and a form with the toggle off then on is submitted, then submitted again unchanged, and an untouched rendered body is saved,
- **Then** `feedback.participation` renders as the console's existing two-option radio pair, exactly two inputs named `feedback.participation` with the hidden `feedback.participation__present` companion and zero checkbox inputs, with the four-locale title and description keys present and the description stating public and account-tied in all four locales; the submitted value is read from and written to the user-scoped file, not the project section file; a change of `enabled` also sets `asked` true; an unchanged submission and an untouched body write nothing; the project `feedback.yaml` stays byte-identical; and `feedback.participation_asked` is not among the rendered fields. Mutants that must die: a checkbox widget; a rendered `asked` field; a console write that leaves `asked` false; a write into the project file.
- **Verify (new-test line, web)**: `go test ./internal/web/ -run '^(TestFeedbackParticipationI18nKeysInAllLocales|TestFeedbackParticipationRendersAsRadioPair)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **Verify (new-test line, settings)**: `go test ./internal/settings/ -run '^(TestFeedbackParticipationFieldIsBool|TestUserScopedEditWritesHomeFileOnly|TestUserScopedValueInvariantTouchesNothing|TestUserScopedWriteSetsAsked)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **Verify (regression guard, web)**: `go test ./internal/web/ -run '^(TestBoolFieldsRenderAsRadio|TestSchemaTogglePresentCompanion|TestFeedbackPanelRendered|TestFeedbackPanelFieldsWired|TestI18nKeySetParity|TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical|TestSaveSchemaSmokeAllSections|TestSchemaSectionsRenderSmoke)$' -count=1 -v` — expect exit 0 and a PASS line for each of the eight names.
- **Verify (regression guard, settings)**: `go test ./internal/settings/ -run '^(TestSchemaCurrentValuesReadsAllSections|TestFeedbackAutoSubmitFieldRegistered|TestFeedbackSectionSeamWritable|TestAbsentDefaultPolarityDeclared)$' -count=1 -v` — expect exit 0 and a PASS line for each of the four names.
- **RED-now**: E14, E15. **Green path**: M2 (the amended probes of `TestFeedbackPanelRendered`, the new `feedback.participation` case of `TestSchemaCurrentValuesReadsAllSections`, and the `MOAI_HOME` isolation land in the same milestone).

### AC-022 — Withdrawal and purge
Maps REQ-ANON-021
- **Given** a queue, a spool, a ledger, and an outbox log with content,
- **When** a flush starts with participation false, and separately the purge command runs,
- **Then** the flush discards unsent items and the spool and appends `discarded` rows while keeping the sent history; purge removes all four stores; neither makes a network request or a model call.
- **Verify (new-test line)**: `go test ./internal/feedback/outbox/ ./internal/cli/ -run '^(TestWithdrawalDiscardsQueueOnFlush|TestParticipationPurge)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E12, E17. **Green path**: M4.

### AC-023 — Skill bodies, docs wording, mirror
Maps REQ-ANON-022
- **Given** the template tree, the three skill-body copies, the four docs-site pages, and the local feedback section file,
- **When** the build and the guard tests run,
- **Then** each of the three skill-body copies mentions participation; the template and local `feedback.yaml` are byte-identical; none of the four docs-site pages still says the issue is created automatically; each page positively states the corrected wording, which is for each locale the substring in the Verify lines below; and `make build` exits 0.
- **Verify (skill bodies)**: `grep -c participation .claude/skills/moai/workflows/feedback.md plugins/moai/skills/moai/workflows/feedback.md internal/template/templates/.claude/skills/moai/workflows/feedback.md` — expect three `path:N` lines, each with N at least 1, and exit 0.
- **Verify (mirror)**: `cmp internal/template/templates/.moai/config/sections/feedback.yaml .moai/config/sections/feedback.yaml` — expect empty stdout and exit 0.
- **Verify (old wording gone)**: the E18 command — expect four `path:0` lines.
- **Verify (positive wording, en)**: `grep -c -F "GitHub issue only after you confirm" docs-site/content/en/utility-commands/moai-feedback.md` — expect `1` or more, exit 0.
- **Verify (positive wording, ko)**: `grep -c -F "확인한 뒤에만 GitHub 이슈" docs-site/content/ko/utility-commands/moai-feedback.md` — expect `1` or more, exit 0.
- **Verify (positive wording, ja)**: `grep -c -F "確認した後にのみ GitHub Issue" docs-site/content/ja/utility-commands/moai-feedback.md` — expect `1` or more, exit 0.
- **Verify (positive wording, zh)**: `grep -c -F "确认后才会创建 GitHub Issue" docs-site/content/zh/utility-commands/moai-feedback.md` — expect `1` or more, exit 0.
- **Verify (neutrality)**: `go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1 -v` — expect exit 0 and `--- PASS: TestTemplateNeutralityAudit ` (the name followed by a space).
- **Verify (build)**: `make build` — expect exit 0.
- **RED-now**: E18, E19 (four entries), E20; the mirror line E21 is green today by design. **Green path**: M7. Mutants that must die: skipping a skill-body copy; deleting the old sentence without writing the new one (the positive patterns); editing one `feedback.yaml` copy only.

### AC-024 — No auto-repair artifact ships; the question never ships without the sender
Maps REQ-ANON-023
- **Given** the embedded template tree and `plugins/moai/`, and the wizard question set with the sender package directory,
- **When** the guards walk them, with a seeded canary file named `auto-repair-canary.md` added in a temporary overlay, and with a synthetic tree that carries the question but lacks the sender,
- **Then** the real trees pass and the canary run fails; no file path contains `auto-repair`, `autorepair`, or `auto_repair`, and no file content contains `moai-bugreport`; the coexistence guard passes on the real tree and fails on the synthetic tree (the red is observed on the synthetic tree before the guard is adopted).
- **Verify (new-test line)**: `go test ./internal/template/ -run '^(TestNoAutoRepairArtifactsShipped|TestNoAutoRepairGuardCatchesCanary)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **Verify (new-test line)**: `go test ./internal/cli/wizard/ -run '^(TestParticipationQuestionRequiresSender|TestParticipationQuestionGuardCatchesMissingSender)$' -count=1 -v` — expect exit 0 and a PASS line for each name.
- **RED-now**: E22a, E22b. **Green path**: M7.

### AC-025 — Static reachability guard and the injected seam
Maps REQ-ANON-025
- **Given** the source of `internal/bugreport`, `internal/feedback/outbox`, `internal/feedback/publish`, and the CLI wiring,
- **When** a go/parser walk reads each non-test file's imports and the CLI flush wiring is exercised with a counting model stub,
- **Then** `internal/bugreport` and `internal/feedback/outbox` import neither `os/exec` nor `net/http` nor any package under `internal/cli`, and `internal/bugreport` imports none of `internal/hook`, `internal/template`, `internal/feedback`, or `internal/cli`; `internal/feedback/publish` imports `net/http` in no file and `os/exec` only in `ghrunner.go`, and imports no package under `internal/cli`; the model seam is an interface owned by `publish`; and the CLI flush call injects the production implementation (a counting stub substituted at that call receives the calls). Mutants that must die: a model call or an HTTP post placed in a helper outside the seam; an `os/exec` import added to the outbox package.
- **Verify (new-test line)**: `go test ./internal/feedback/publish/ -run '^(TestBugreportAndOutboxImportAllowlist|TestPublishImportAllowlist|TestModelSeamIsInjectedNotImported)$' -count=1 -v` — expect exit 0 and a PASS line for each of the three names.
- **Verify (new-test line)**: `go test ./internal/cli/ -run '^TestFlushWiresClaudeSummarizer$' -count=1 -v` — expect exit 0 and a PASS line for the name.
- **RED-now**: E3, E11. **Green path**: M6 (the bugreport half is green from M1; the criterion closes when the seam exists).

## Edge cases

- A user runs `moai update` twice in one terminal session: the second run does not ask (AC-005).
- Two users hit the same fingerprint at the same moment before the search index reflects the first issue: two issues are created; the consumer merges by fingerprint and sums comments. This approximate behaviour is accepted; the M5 first-test item measures the title-token search and records the decision (plan.md).
- `gh` is installed but the user is logged into a different host: the sender leaves the item queued (AC-015).
- The participation toggle is turned off while items are queued: the next flush discards them (AC-022).
- The moai home directory is absent or unreadable: every consent value reads false and nothing is captured (AC-001, AC-002).
- A cloned repository ships `participation: true` and a different `repository`: both are ignored (AC-001, AC-003).
- Someone posts a forged occurrence marker on a public issue: it counts toward the advisory total, its disagreeing fields are ignored (AC-020).

## Definition of Done

All 25 criteria pass with their named commands, each Verify line satisfying the preface rule (exit 0 and a `--- PASS` line per named test); `go vet` and `golangci-lint` clean on touched packages; the guard families (shipped-key readers, i18n parity, template neutrality, radio widget, present-companion) pass; the amendment of the quiet-wizard pins is recorded in the commit message; progress.md §E.2 carries each criterion's observed output.
