# SPEC-SESSION-CC-VERSION-002 — Acceptance Criteria

Seven criteria against seven requirements. Every criterion names a command and an expected
result; none is satisfied by reading a file and forming a judgement.

Plan-phase baselines were measured in this tree at `099250516` (the predecessor's final
DEBT-marker commit). Where a criterion is satisfied by an **absence** or is a **preservation
half** (green today, load-bearing through the rework), the baseline is stated so a green is
new information rather than a vacuous pass. RED-now cells for new behavior are captured by the
run phase before GREEN (plan.md §E8); the W2 RED was additionally measured at plan time and is
cited where it stands.

Classification: AC-SCV-012, AC-SCV-013, and AC-SCV-014 are release-blocking, with their RED
evidence captured by the run phase before GREEN (the plan.md §E8 deferral — no RED exists to
re-execute at plan time for behavior that does not exist yet); AC-SCV-011, AC-SCV-015,
AC-SCV-016, and AC-SCV-017 are regression-guards — preservation or shape-pinning criteria
whose green at baseline `099250516` was measured at plan time, with AC-SCV-016's defect RED
additionally observed at plan time (exit code 1, recorded below as its own field) and
re-captured by the run phase before GREEN.

## §A W1 — structural interpretation

**AC-SCV-011** (maps REQ-SCV-011) — Given the reworked interpreters, When the separator
semantics of the pinned suite run, Then all pass unchanged: MoAI's `--` splits the segments,
a `--` is never consumed as a value, everything after Claude's own second `--` is prompt text,
and the launcher's value flags are inert after MoAI's separator while keeping their behavior
before it. Command:
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/cli/ -run '^(TestResumeScanStopsAtClaudeSeparator|TestSeparatorWinsOverAmbiguousValue|TestSeparatorInterplaySkipsValues|TestPostSeparatorLauncherFlagsAreInert|TestPostSeparatorLauncherFlagsInertGuard|TestPreSeparatorLauncherFlagsKeepValueBehavior)$' -count=1 -v`
→ each of the 6 named functions `--- PASS` (exact swept count 6, no `[no tests to run]`, zero
`--- FAIL`). Preservation half: green at `099250516`; the load-bearing half is that the
explicit segment model reproduces the counter-based behavior on every pinned input.

**AC-SCV-012** (maps REQ-SCV-012) — Given a fixture `claude --help` text injected through the
derivation seam, When the model is derived, Then an option the fixture marks required-value
(`<value>`) classifies as required-value, one marked `[value]` as optional-value, one marked
valueless as boolean; and Given a failing derivation (seam returns an error or a timeout),
Then the model degrades to the compile-time snapshot and no error escapes. Commands:
`go test ./internal/cli/ -run '^TestClaudeOptionModelDerivation$' -v` → `--- PASS` (degradation
subtests included); and
`grep -n "exec.Command" internal/cli/lane_resume_model_test.go internal/cli/lane_resume_test.go`
→ 0 hits — no test of the derivation spawns a process (the seam carries it).

**AC-SCV-013** (maps REQ-SCV-013) — Four labeled states, each its own Given/When/Then so no
antecedent mis-attaches:

- **Given** an injected model in which the novel option `--t1515-probe` is **absent**
  (unknown), **When** the guard reads
  `["--name", "lane-3", "--", "--t1515-probe", "--resume", "<id>"]`, **Then** it FIRES
  (fail-closed — it judges the token after the unknown option).
- **Given** the injected model marks `--t1515-probe` **required-value**, **When** the guard
  reads the same argv, **Then** it does NOT fire (the model-known value is consumed).
- **Given** the injected model leaves `--t1515-probe` **unknown**, **When** the validator
  reads `["--", "--t1515-probe", "--resume"]`, **Then** it refuses the bare resume
  (unknown-must-judge — the semantics pinned by `TestPostSeparatorLauncherFlagsAreInert`).
- **Given** the injected model marks `--t1515-probe` **optional-value**, **When** the
  validator reads `["--", "--t1515-probe", "--resume", "<id>"]`, **Then** it passes silently
  (ambiguity never refuses).

Command:
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run '^TestOptionModelPolarityDefaults$' -v`
→ `--- PASS`, one subtest per cell of the class × mode matrix. This is the criterion that
encodes §A.3 of plan.md: the model class, not the token shape, decides.

**AC-SCV-014** (maps REQ-SCV-014) — Given the completed rework, When the structural greps run,
Then the three hand-synced map identifiers are gone from the walk and the discharged DEBT
markers are gone. Commands:
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && grep -rn "claudeValueTakingOptions\|launcherValueTakingOptions\|ambiguousValueOptions" internal/cli/ --include="*.go"`
→ 0 hits (baseline: 9 references in `lane_resume.go`, measured at `099250516`); and
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && grep -rn "@MX:UPGRADE: t1515" internal/cli/lane_resume.go internal/session/ccversion.go`
→ 0 hits (baseline: 2, measured at `099250516`); and
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && grep -n "source-version provenance" internal/cli/lane_resume_model.go`
→ the snapshot's provenance comment, whose `X.Y.Z` version token equals the version token the
run phase's M4 live re-measure observes and records in `progress.md` §E.2 — an
observation-based, version-agnostic equality across the two grep-extracted tokens (no version
literal is embedded in this criterion, so any honestly-measured installed version satisfies
it, and a provenance token the M4 measure never observed — a stale unmeasured string — fails
it). Plus, in the same test binary
as AC-SCV-013's safety subtests, both sides of the named residual are measured against a model
whose snapshot is deliberately stripped of one required-value entry: the guard reading
`["--name", "lane-3", "--", "<stripped>", "--resume", "<id>"]` FIRES (it judges the stripped
option's next token — no leak), and the validator reading `["--", "<stripped>", "--resume"]`
refuses — the compound-condition false-refusal of plan.md §A.4, observed and asserted as
exactly that outcome rather than left silent.

**AC-SCV-015** (maps REQ-SCV-015) — Given the whole rework, When the r1-r5 regression suite
runs, Then all 19 test functions pass unchanged and the test file is byte-identical.
Commands: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/cli/ -run '^(TestLaneJoinChildArgv|TestResumeRequiresValue|TestResumeShortAliasRequiresValue|TestResumeAliasExactTokenOnly|TestGuardRecognizesAttachedShortForm|TestValidatorSkipsOptionValues|TestValidatorStillRefusesValueless|TestGuardSkipsOptionValues|TestSeparatorInterplaySkipsValues|TestGuardFiresOnShortCluster|TestGuardFiresOnAmbiguousValueOption|TestAmbiguityNeverRefusesInValidator|TestPostSeparatorLauncherFlagsAreInert|TestPostSeparatorLauncherFlagsInertGuard|TestPreSeparatorLauncherFlagsKeepValueBehavior|TestSeparatorWinsOverAmbiguousValue|TestRelaunchRefusesResumeAlias|TestResumeScanStopsAtClaudeSeparator|TestRelaunchRefusesResumeToken)$' -count=1 -v`
→ zero `--- FAIL`, every one of the 19 function names present with `--- PASS`, exact swept
count 19 (subtest `--- PASS` lines are additional and expected): TestLaneJoinChildArgv,
TestResumeRequiresValue, TestResumeShortAliasRequiresValue, TestResumeAliasExactTokenOnly,
TestGuardRecognizesAttachedShortForm, TestValidatorSkipsOptionValues,
TestValidatorStillRefusesValueless, TestGuardSkipsOptionValues,
TestSeparatorInterplaySkipsValues, TestGuardFiresOnShortCluster,
TestGuardFiresOnAmbiguousValueOption, TestAmbiguityNeverRefusesInValidator,
TestPostSeparatorLauncherFlagsAreInert, TestPostSeparatorLauncherFlagsInertGuard,
TestPreSeparatorLauncherFlagsKeepValueBehavior, TestSeparatorWinsOverAmbiguousValue,
TestRelaunchRefusesResumeAlias, TestResumeScanStopsAtClaudeSeparator,
TestRelaunchRefusesResumeToken); and `git diff --name-only 099250516 -- internal/cli/lane_resume_test.go`
→ empty (base-pinned form — the bare form compares working tree to index, so a staged
mutation would report empty). Plus the r5 instance: `go test ./internal/cli/ -run '^TestRemoteControlPrefixValue$' -v`
→ `--- PASS` — `--remote-control-session-name-prefix --resume` (a prefix value that literally
reads `--resume`) passes the validator.

## §B W2 — install-root-anchored extraction

**AC-SCV-016** (maps REQ-SCV-016) — Given the anchored extractor, When the table of path
shapes runs, Then `/opt/versions/9/tools/claude/versions/2.1.281` reads `2.1.281` (the overlay
repro), `~/.local/share/claude/versions/2.1.287` reads `2.1.287`, an npm-style
`…/claude-code/2.1.284` reads `2.1.284`, a trailing `…/2.1.281/claude` binary-name shape
reads `2.1.281`, and `/opt/versions/9/tools/other/versions/3.0.0` (no claude product
directory) reads "". Command:
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/session/ -run '^TestVersionSegmentAnchoredToInstallRoot$' -v`
→ `--- PASS`, one subtest per shape. RED-now (measured at plan time, tree `099250516`): a
`go test -overlay` probe of the first shape returned `--- FAIL` with
`versionSegmentFromPath = "9", want 2.1.281`, exit code **1** (recorded as its own field; the
plan-time invocation piped through `tail`, so the run phase re-captures this RED verbatim on
the pre-implementation tree — single invocation, no pipes — before GREEN).

**AC-SCV-017** (maps REQ-SCV-017) — Given each degradation input — a resolved path with no
anchored claude product-directory segment, a mapping line stripped of ` (deleted)` still
naming the claude binary, and the existing dead-pid / unreadable-mapping fixtures — When the
reads run, Then every degraded value renders `unknown`, never an inferred value, and the
existing degradation and mapping tests stay green. Command:
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/session/ -run '^(TestVersionDegradationRendersUnknown|TestRunningVersionFromDeletedBinary|TestRunningVersionFromInjectedMapping|TestInstalledVersionFromResolvedPath)$' -count=1 -v`
→ all 4 `--- PASS`, zero `--- FAIL`. Preservation half: green at `099250516`; the load-bearing
half is that the anchor change does not degrade any previously-satisfying read.

## §C Quality gates and Definition of Done

- All seven criteria PASS in the run-phase matrix (`progress.md` §E.2), each with command +
  verbatim output + tree SHA.
- `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/cli/... ./internal/session/...` → exit 0.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go vet ./internal/cli/... ./internal/session/...` → clean.
- `golangci-lint run ./internal/cli/... ./internal/session/...` → 0 new issues vs the §C
  pre-flight baseline.
- New/changed files ≥ 85% covered (`go test -cover ./internal/cli/ ./internal/session/`).
- PRESERVE list untouched: `git diff --name-only 099250516` over the run shows no file
  outside `internal/cli/`, `internal/session/`, and
  `.moai/specs/SPEC-SESSION-CC-VERSION-002/` (excluding `.moai/reports/` evidence exports) —
  the base-pinned form, since the bare form compares working tree to index and a staged
  mutation would report empty; and `git diff --name-only 099250516 -- internal/cli/lane_resume_test.go`
  → empty (E6).
- The §F run-phase measurements of `plan.md` are either measured or recorded as explicit gaps.
