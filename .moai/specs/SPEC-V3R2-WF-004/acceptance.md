# SPEC-V3R2-WF-004 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-02 | MoAI Plan Workflow (Phase 1B)     | Initial G/W/T conversion of 13 ACs (AC-WF004-01 through AC-WF004-13)   |

---

## Scope

This document converts each of the 13 ACs from `spec.md` §6 into Given/When/Then format with happy-path + edge-case + test-mapping notation. Test-mapping uses `internal/template/agentless_audit_test.go` (new, M1) and direct skill-content audit for content-level assertions; runtime assertions for `--mode` flag rejection are anchored at the skill-body sentinel level since utility subcommands have no Go handler (per research.md §6.4).

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- AC-WF004-12 has the highest test importance (REQ-WF004-013 CI guard).

---

## AC-WF004-01 — `/moai fix` runs 3-phase pipeline without Agent() control flow

Maps to: REQ-WF004-001, REQ-WF004-003, REQ-WF004-004.

### Happy path

- **Given** the user invokes `/moai fix` with valid arguments and the worktree contains files with diagnostic issues
- **When** the orchestrator reads `.claude/skills/moai/workflows/fix.md` and executes its workflow
- **Then** the workflow follows the declared flow `Parallel Scan -> Classify -> Fix -> Verify -> Report` (fix.md:33) in deterministic order
- **And** Phase 1+2+2.5 (localize) execute before Phase 3 (repair) which executes before Phase 4 (validate)
- **And** no `Use the .* subagent to (decide|determine|orchestrate|route)` pattern appears in the body of `fix.md`
- **And** the Pipeline Contract section (added in M2) declares the 3-phase mapping explicitly

### Edge case — partial scan failure

- **Given** Scanner 2 (AST-grep) fails but Scanner 1 (LSP) and Scanner 3 (Linter) succeed
- **When** the workflow executes
- **Then** the localize phase still completes with diagnostics from successful scanners (per fix.md:121 "If any scanner fails, continue with results from successful scanners")
- **And** the failed scanner is noted in the report
- **And** the 3-phase contract is preserved (no LLM dispatch invoked to recover)

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/fix.md` in `internal/template/agentless_audit_test.go` (M1) — fails if any forbidden control-flow regex matches `fix.md` body.
- Manual verification: read `.claude/skills/moai/workflows/fix.md:33` confirms flow declaration; read `fix.md:159-163` confirms static Level-to-agent lookup table (executor delegation, not control flow).

---

## AC-WF004-02 — `/moai coverage` 3-phase pipeline (untested → tests added → re-measured)

Maps to: REQ-WF004-003.

### Happy path

- **Given** the project has files below the coverage target declared in `.moai/config/sections/quality.yaml` (`test_coverage_target: 85`)
- **When** the user invokes `/moai coverage`
- **Then** the workflow executes Phase 1 Coverage Measurement → Phase 2 Gap Analysis (localize) → Phase 3 Test Generation (repair) → Phase 4 Verification (validate)
- **And** Phase 3 is skipped when `--report` flag is supplied (per coverage.md:114), exiting after gap report
- **And** the localize → repair → validate mapping documented in the Pipeline Contract section (added in M2) matches the actual phase execution

### Edge case — `--report` flag (no repair phase)

- **Given** `/moai coverage --report` is invoked
- **When** the workflow runs
- **Then** localize completes (Phase 1 + Phase 2)
- **And** repair is skipped (no test generation)
- **And** the workflow exits with a gap report (no validate phase)
- **And** this is consistent with REQ-WF004-007 generalized: when localize finds 0 *targetable* gaps OR when the user explicitly requests a localize-only invocation, repair and validate are skipped

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/coverage.md` (M1).
- Pipeline Contract section verification: read `.claude/skills/moai/workflows/coverage.md` post-M2 confirms localize ← Phase 1+2; repair ← Phase 3; validate ← Phase 4.

---

## AC-WF004-03 — `/moai mx` 3-phase pipeline (tag-drift → annotation → re-scan)

Maps to: REQ-WF004-003.

### Happy path

- **Given** the codebase has files needing @MX tag updates (drift detected via fan_in or pattern detection)
- **When** the user invokes `/moai mx --all`
- **Then** Pass 1 Full File Scan + Pass 2 Selective Deep Read (localize) → Pass 3 Batch Edit (repair) → post-edit MX scan via PostToolUse hooks (validate) execute in order
- **And** no `Agent()` invocation occurs in any pass per research.md §2.2.3 (mx is the closest existing match to Agentless)
- **And** the 3-phase mapping is documented in the Pipeline Contract section: localize ← Pass 1+2; repair ← Pass 3; validate ← post-edit MX hook scan

### Edge case — `--dry` flag (no repair)

- **Given** `/moai mx --dry` is invoked
- **When** the workflow runs
- **Then** Pass 1 + Pass 2 complete (localize)
- **And** Pass 3 is skipped (per mx.md:51 "Preview only - show tags to add without modifying files")
- **And** no validate phase runs (nothing to validate)
- **And** exit status is `no-op` analog (preview-only)

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/mx.md` (M1).
- Reference verification: read `.claude/skills/moai/workflows/mx.md:71-149` confirms the 3-Pass scan structure is unchanged by this SPEC.

---

## AC-WF004-04 — `/moai codemaps` 3-phase pipeline (stale map → regenerate → diff)

Maps to: REQ-WF004-003.

### Happy path

- **Given** `.moai/project/codemaps/` directory contains maps that are out of sync with the codebase
- **When** the user invokes `/moai codemaps --force`
- **Then** Phase 1 Codebase Exploration (localize, via `Explore` subagent) → Phase 2+3 Architecture Analysis + Map Generation (repair, via `manager-docs` subagent) → Phase 4 Verification (validate, orchestrator-side existence/consistency checks) execute in order
- **And** the agent invocations (Explore, manager-docs) are *executor* delegations within phases — no LLM picks the next phase
- **And** the Pipeline Contract section documents: localize ← Phase 1; repair ← Phase 2+3; validate ← Phase 4

### Edge case — `--area` flag scoped exploration

- **Given** `/moai codemaps --area api` is invoked
- **When** the workflow runs
- **Then** Phase 1 limits exploration to `api` and its dependencies (per codemaps.md:56)
- **And** the 3-phase contract is preserved with narrowed scope
- **And** Phase 4 verifies the area-specific maps (`.moai/project/codemaps/api/*.md`)

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/codemaps.md` (M1).
- Reference verification: read `.claude/skills/moai/workflows/codemaps.md:33` confirms flow declaration.

---

## AC-WF004-05 — `/moai clean` 3-phase pipeline (locate trash → remove → validate tree)

Maps to: REQ-WF004-003.

### Happy path

- **Given** the codebase contains dead code (unused imports, orphaned functions, etc.)
- **When** the user invokes `/moai clean`
- **Then** Phase 1+2 Static Analysis + Usage Graph (localize) → Phase 4 Safe Removal (repair, after Phase 3 user approval via AskUserQuestion) → Phase 5 + Phase 5.5 Test Verification + MX Tag Cleanup (validate) execute in order
- **And** Phase 3 user approval is orchestrator-side (AskUserQuestion is the orchestrator's tool, not a subagent's)
- **And** Phase 4-5 agent delegations (`expert-refactoring`, `expert-testing`) are executor roles per research.md §2.2.5

### Edge case — `--dry` flag (no repair, no validate)

- **Given** `/moai clean --dry` is invoked
- **When** the workflow runs
- **Then** Phase 1 + Phase 2 complete (localize)
- **And** Phase 3 displays the analysis results (per clean.md:129)
- **And** Phase 4-5.5 are skipped
- **And** exit status is `no-op` analog (analysis-only)

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/clean.md` (M1).
- Reference verification: read `.claude/skills/moai/workflows/clean.md:33,212-219` confirms the agent chain summary lists *executor* roles per phase.

---

## AC-WF004-06 — Subcommand Classification matrix exists in workflow rule file

Maps to: REQ-WF004-005.

### Happy path

- **Given** the run phase has completed M4 (Subcommand Classification matrix insertion)
- **When** an inspector reads `.claude/rules/moai/workflow/spec-workflow.md`
- **Then** a section titled `## Subcommand Classification (Pipeline vs Multi-Agent)` exists immediately after the `## Phase Overview` table (~line 18)
- **And** the matrix table contains exactly 9 rows: 5 utility (`fix`, `coverage`, `mx`, `codemaps`, `clean`) classified as "Pipeline (Agentless)" and 4 implementation (`plan`, `run`, `sync`, `design`) classified as "Multi-Agent"
- **And** each row includes the 3-phase contract description for utility rows and "n/a — open-ended" for implementation rows
- **And** the matrix explicitly notes that `/moai feedback`, `/moai review`, `/moai e2e` are out of scope (deferred per `spec.md` §1.2)

### Edge case — embedded template parity

- **Given** the matrix has been added to `.claude/rules/moai/workflow/spec-workflow.md`
- **When** `make build` regenerates `internal/template/embedded.go`
- **Then** the embedded copy of `spec-workflow.md` (read via `EmbeddedTemplates()`) contains the identical matrix
- **And** `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` has been edited to match

### Test mapping

- Manual verification: post-M4, run `grep -n "Subcommand Classification" .claude/rules/moai/workflow/spec-workflow.md` returns the heading line.
- Embedded parity: existing `internal/template/embed_test.go` exercises the embedded FS; the new audit test reads from the same FS to verify content.

---

## AC-WF004-07 — Localize finds 0 targets → no-op exit code 0

Maps to: REQ-WF004-007.

### Happy path

- **Given** `/moai fix` is invoked on a clean codebase with no diagnostic issues
- **When** Phase 1 Parallel Scan completes
- **Then** the issue list is empty
- **And** Phase 2 Classification produces 0 issues at every level
- **And** Phase 3 Auto-Fix is skipped (no targets to fix)
- **And** Phase 4 Verification is skipped (nothing was modified)
- **And** the orchestrator emits a "no-op" status report and exits with code 0

### Edge case — only false positives detected

- **Given** Phase 1 detects issues but Phase 2.5 Pre-Fix MX Context Scan reclassifies all of them as protected (e.g., all in @MX:ANCHOR functions)
- **When** the workflow proceeds
- **Then** Phase 3 finds 0 actionable fix targets
- **And** the workflow exits with status "no-op" (semantically equivalent to 0 targets)

### Edge case — `coverage` analog

- **Given** `/moai coverage` is invoked on a project already at the coverage target
- **When** Phase 1 + Phase 2 complete
- **Then** the gap list is empty
- **And** the workflow reports "target met" and skips Phase 3 (test generation)
- **And** exit code is 0

### Test mapping

- Pipeline Contract section content check (M2): the `MODE_FLAG_IGNORED_FOR_UTILITY`-bearing section explicitly documents the no-op exit per the M2 template in plan.md §2 M2.
- Manual verification: run `/moai fix` on a clean test fixture; confirm exit code 0 + "no-op" report.

---

## AC-WF004-08 — Repair encounters unresolvable error → fail-fast termination

Maps to: REQ-WF004-008.

### Happy path

- **Given** `/moai fix` is invoked and Phase 1+2 produces a Level 4 (manual) issue or a Level 3 fix that the agent reports as unresolvable
- **When** Phase 3 attempts to apply the fix and the executor agent returns an error
- **Then** the pipeline terminates immediately at Phase 3
- **And** the error is surfaced in the report with file:line context
- **And** Phase 4 (validate) is NOT entered (fail-fast)
- **And** no multi-agent fallback is attempted (consistent with REQ-WF004-008 "no multi-agent fallback")

### Edge case — partial repair success

- **Given** Phase 3 successfully repairs 3 issues but fails on the 4th
- **When** the failure is detected
- **Then** the 3 successful repairs ARE preserved (commit boundary respected)
- **And** the pipeline still terminates fail-fast on the 4th
- **And** the report enumerates: 3 fixed, 1 failed, 0 deferred

### Edge case — coverage analog

- **Given** `/moai coverage` Phase 3 generates tests but they fail to compile
- **When** test verification (Phase 4) detects the failure
- **Then** the workflow terminates with the compile error reported
- **And** no retry / re-generation loop runs (fail-fast per REQ-WF004-008)

### Test mapping

- Pipeline Contract section content check (M2): the M2 template explicitly states "When repair encounters an unresolvable error, the pipeline terminates and reports the error. There is no multi-agent fallback."
- Manual verification: contrived integration test with intentionally-broken fixture in a temp project (per `CLAUDE.local.md` §6 t.TempDir() pattern).

---

## AC-WF004-09 — Agentless subcommand main flow does not fire SubagentStart hook

Maps to: REQ-WF004-009.

### Happy path

- **Given** `/moai mx` is invoked (the purest Agentless example per research.md §2.2.3)
- **When** Pass 1, Pass 2, Pass 3 execute end-to-end
- **Then** no SubagentStart hook events originate from the mx workflow's main orchestration loop
- **And** the orchestrator uses `Glob` / `Grep` / `Read` / `Edit` tools directly (not via Agent() spawn)
- **And** any hook activity observed is from the PostToolUse fixed pipeline (MX/LSP/metrics per `internal/hook/post_tool.go`), which is *not* originating from mx itself

### Edge case — `fix` with executor agent delegation

- **Given** `/moai fix` invokes `expert-backend` for Level 1 fix execution
- **When** that agent runs
- **Then** SubagentStart fires for `expert-backend` — but this is acceptable because `expert-backend` is an *executor* delegation within Phase 3, not a control-flow decision
- **And** the SubagentStart event records the executor role, not a "decide-next-phase" role
- **And** REQ-WF004-009's intent is preserved: no subagent is dispatched to *decide what runs next*

### Test mapping

- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/mx.md` confirms `mx.md` body contains no `Use the .* subagent to` patterns at all.
- Static content audit: `TestAgentlessUtilityNoLLMControlFlow/fix.md` confirms `fix.md` Phase 3 delegations use *executor* verbs (apply, fix, refactor) — not control verbs (decide, dispatch, route).
- Manual verification: trace SubagentStart hook events during a `/moai mx --all` run; confirm 0 events from mx orchestration (per research.md §3 hook architecture).

---

## AC-WF004-10 — `/moai fix --mode team` ignores `--mode` flag

Maps to: REQ-WF004-011.

### Happy path

- **Given** the user invokes `/moai fix --mode team`
- **When** the orchestrator parses arguments
- **Then** the `--mode` flag is detected and ignored (not propagated to phase execution)
- **And** the orchestrator emits an info-level log entry containing the exact string `MODE_FLAG_IGNORED_FOR_UTILITY`
- **And** the workflow proceeds with the standard fixed pipeline (Phases 1-4) as if `--mode` were absent
- **And** the Pipeline Contract section in `fix.md` (added in M2) documents this behavior verbatim

### Edge case — all 4 invalid `--mode` values

- **Given** the user invokes any of `/moai fix --mode autopilot`, `/moai fix --mode loop`, `/moai fix --mode team`, `/moai fix --mode pipeline`
- **When** the workflow runs
- **Then** ALL invocations produce the same `MODE_FLAG_IGNORED_FOR_UTILITY` info log
- **And** ALL produce identical pipeline behavior (the flag is universally ignored on utility subcommands)
- **And** the user is not blocked or errored — the workflow is permissive on invalid mode values for utilities

### Edge case — same behavior across all 5 utility subcommands

- **Given** any of `/moai {fix,coverage,mx,codemaps,clean} --mode <any-value>` is invoked
- **When** the orchestrator parses arguments
- **Then** all 5 produce `MODE_FLAG_IGNORED_FOR_UTILITY` info log
- **And** all 5 proceed with their respective fixed pipelines

### Test mapping

- Static content audit: `TestUtilitySkillsContainModeFlagIgnoredSentinel` in `internal/template/agentless_audit_test.go` (M1) walks the 5 utility skill files and asserts each contains the literal string `MODE_FLAG_IGNORED_FOR_UTILITY`.
- Pseudocode for the Go test (do not implement here — outline only):
  - For each of {`fix.md`, `coverage.md`, `mx.md`, `codemaps.md`, `clean.md`} under `.claude/skills/moai/workflows/` in the embedded FS:
    - Read body bytes.
    - Assert `bytes.Contains(body, []byte("MODE_FLAG_IGNORED_FOR_UTILITY"))` is true.
    - On failure: emit `t.Errorf("%s missing MODE_FLAG_IGNORED_FOR_UTILITY sentinel", path)`.

---

## AC-WF004-11 — `/moai plan --mode pipeline` rejects with MODE_PIPELINE_ONLY_UTILITY

Maps to: REQ-WF004-014.

### Happy path

- **Given** the user invokes `/moai plan SPEC-V3R2-XYZ --mode pipeline`
- **When** the orchestrator parses arguments and detects `--mode pipeline` on a multi-agent subcommand
- **Then** the orchestrator emits an error-level message containing the exact string `MODE_PIPELINE_ONLY_UTILITY`
- **And** the workflow does NOT execute the plan phases (rejected at flag-validation step)
- **And** the user receives guidance pointing to the utility subcommand set (`fix`, `coverage`, `mx`, `codemaps`, `clean`) where `pipeline` mode applies

### Edge case — same rejection on all 4 implementation subcommands

- **Given** any of `/moai {plan,run,sync,design} --mode pipeline` is invoked
- **When** argument parsing runs
- **Then** all 4 produce the `MODE_PIPELINE_ONLY_UTILITY` error
- **And** none of the 4 begin their respective workflows

### Edge case — error key shared with WF-003

- **Given** the same error key `MODE_PIPELINE_ONLY_UTILITY` is referenced by SPEC-V3R2-WF-003 REQ-WF003-016
- **When** WF-003 lands its `--mode` validation logic
- **Then** the error key remains a single, shared sentinel — not two divergent strings
- **And** integration between WF-004 and WF-003 is consistent (no rename or alias needed)

### Test mapping

- Static content audit: `TestImplementationSkillsContainPipelineRejectionSentinel` in `internal/template/agentless_audit_test.go` (M1) walks the 4 implementation skill files and asserts each contains the literal string `MODE_PIPELINE_ONLY_UTILITY`.
- Pseudocode for the Go test (do not implement — outline only):
  - For each of {`plan.md`, `run.md`, `sync.md`, `design.md`} under `.claude/skills/moai/workflows/` in the embedded FS:
    - Read body bytes.
    - Assert `bytes.Contains(body, []byte("MODE_PIPELINE_ONLY_UTILITY"))` is true.
    - On failure: emit `t.Errorf("%s missing MODE_PIPELINE_ONLY_UTILITY sentinel", path)`.

---

## AC-WF004-12 — PR adding Agent() to fix flagged with AGENTLESS_CONTROL_FLOW_VIOLATION

Maps to: REQ-WF004-013. **Highest test importance — this is the regression guard.**

### Happy path (current state — green CI)

- **Given** the current state of the 5 utility skill files (no LLM-driven control flow per research.md §4)
- **When** `TestAgentlessUtilityNoLLMControlFlow` runs in CI
- **Then** all 5 utility skill subtests PASS (no forbidden patterns matched)
- **And** the test produces no output beyond the standard `--- PASS:` line per subtest
- **And** CI all-green status is achieved

### Failure scenario (regression detection)

- **Given** a future PR adds the line `Use the manager-strategy subagent to decide which fix level to apply next` to `fix.md`
- **When** that PR's CI runs `go test ./internal/template/ -run TestAgentlessUtilityNoLLMControlFlow`
- **Then** the test FAILS with output matching:
  - `t.Errorf("AGENTLESS_CONTROL_FLOW_VIOLATION: %s contains forbidden pattern %q at body line %d", path, regex, line)`
- **And** the CI run reports the violation
- **And** the PR is blocked from merge until the violation is resolved (the line is removed or the SPEC is updated)

### Edge case — false positive prevention (executor delegation allowed)

- **Given** `fix.md` contains the existing line `Level 1 (import, formatting): expert-backend or expert-frontend subagent` (fix.md:160)
- **When** the regex set evaluates this line
- **Then** the line is NOT flagged because:
  - It does not contain control verbs (decide, orchestrate, route, dispatch, choose, select, determine)
  - It is a static lookup table entry, not a "Use the X subagent to do Y" sentence
- **And** the test PASSES on this line

### Edge case — code blocks excluded from scan

- **Given** `fix.md` body contains a fenced code block with example agent-spawn pseudocode
- **When** the regex evaluation runs
- **Then** content within the fenced code block is excluded from the scan (per research.md §6.2 "excluding code blocks")
- **And** documentation-only mentions of dispatch patterns do not falsely trigger the test

### Test name (canonical)

`TestAgentlessUtilityNoLLMControlFlow` (in `internal/template/agentless_audit_test.go`).

### Subtests (one per skill file)

- `TestAgentlessUtilityNoLLMControlFlow/fix.md`
- `TestAgentlessUtilityNoLLMControlFlow/coverage.md`
- `TestAgentlessUtilityNoLLMControlFlow/mx.md`
- `TestAgentlessUtilityNoLLMControlFlow/codemaps.md`
- `TestAgentlessUtilityNoLLMControlFlow/clean.md`

### Assertion outline (Go test scaffold name only — DO NOT implement)

```
TestAgentlessUtilityNoLLMControlFlow:
  fsys ← EmbeddedTemplates()
  utilitySkills ← {fix.md, coverage.md, mx.md, codemaps.md, clean.md} under .claude/skills/moai/workflows/
  forbiddenRegex ← compiled set per research.md §6.2 (4 patterns, case-insensitive)
  for each path in utilitySkills:
    body ← ReadFile(fsys, path)
    bodyExcludingCodeBlocks ← stripFencedCodeBlocks(body)
    for each (regex, label) in forbiddenRegex:
      if regex.Match(bodyExcludingCodeBlocks):
        t.Errorf("AGENTLESS_CONTROL_FLOW_VIOLATION: %s matches forbidden pattern %q", path, label)
```

### Test mapping

- Primary: `TestAgentlessUtilityNoLLMControlFlow` (M1).
- Secondary: review `commands_audit_test.go:11-50` as the implementation reference scaffold.

---

## AC-WF004-13 — `/moai run SPEC-001 --mode loop` honors `--mode` per WF-003

Maps to: REQ-WF004-010.

### Happy path

- **Given** SPEC-V3R2-WF-003 has landed (or is sibling-merged) and the `--mode` flag is supported on `/moai run`
- **When** the user invokes `/moai run SPEC-001 --mode loop`
- **Then** `/moai run` honors the `--mode` flag per WF-003 REQ-WF003-001 (does NOT ignore it as utilities do)
- **And** the workflow invokes the Ralph Engine loop per WF-003 REQ-WF003-008
- **And** the `## Mode Flag Compatibility` section in `run.md` (added in M3) explicitly documents that `autopilot|loop|team` modes are valid here, and `pipeline` mode is rejected

### Edge case — `--mode pipeline` on `/moai run`

- **Given** `/moai run SPEC-001 --mode pipeline` is invoked
- **When** argument parsing runs
- **Then** the orchestrator emits `MODE_PIPELINE_ONLY_UTILITY` (per AC-WF004-11)
- **And** the run workflow does NOT begin
- **And** this contradicts AC-WF004-13's happy path only for the specific value `pipeline` — other valid modes are honored

### Edge case — same support across all 4 implementation subcommands

- **Given** any of `/moai {plan,run,sync,design}` with a valid `--mode` (per WF-003)
- **When** the workflow runs
- **Then** the mode is honored (or ignored for `plan`/`sync` per WF-003 REQ-WF003-005, but not rejected as `pipeline` would be)
- **And** the multi-agent classification permits open-ended mode handling

### Test mapping

- Static content audit: `TestImplementationSkillsContainPipelineRejectionSentinel` (M1) confirms `MODE_PIPELINE_ONLY_UTILITY` sentinel is present.
- Cross-SPEC integration: when WF-003 lands, its own AC-WF003-02 (`/moai run --mode loop` invokes Ralph) provides the runtime verification.
- Manual verification post-WF-003-merge: invoke `/moai run SPEC-001 --mode loop` and observe Ralph Engine activation.

---

## Quality Gate Hooks (cross-reference)

### TRUST 5 framework alignment

- **Tested**: AC-WF004-12 enforces CI guard via Go test; AC-WF004-10/11 enforce sentinel presence via static audit. All 13 ACs have explicit test mapping.
- **Readable**: Pipeline Contract sections (M2) and Mode Flag Compatibility sections (M3) use consistent template wording for clarity.
- **Unified**: Embedded-template parity (`make build` after every skill edit per `CLAUDE.local.md` §2 Template-First Rule) ensures source and embedded copies remain in sync.
- **Secured**: No new attack surface — this SPEC only adds documentation sections and a static audit test. No runtime behavior change.
- **Trackable**: CHANGELOG entry (M5) + commit messages (per `CLAUDE.local.md` §4 Conventional Commits) provide audit trail.

### LSP quality gates

Per `.moai/config/sections/quality.yaml` `lsp_quality_gates`:
- **plan phase**: `require_baseline: true` — captured at this SPEC's plan completion (current commit `a3be99e67`).
- **run phase**: `max_errors: 0`, `max_type_errors: 0`, `max_lint_errors: 0`, `allow_regression: false` — `agentless_audit_test.go` must compile cleanly; no new lint warnings introduced.
- **sync phase**: `max_warnings: 10`, `require_clean_lsp: true` — verified post-M5 by running `golangci-lint run` and `go vet ./...`.

### Pre-submission self-review checklist

Per `.claude/rules/moai/workflow/spec-workflow.md:79-80`:

- [ ] Full diff reviewed against this acceptance.md before commit.
- [ ] Asked "Is there a simpler approach?" — answer documented in plan.md §1.3 (this is the simpler approach: declarative classification + static audit, not a Go runtime refactor).
- [ ] Asked "Would removing any changes still satisfy the SPEC?" — minimum set is M1 (test) + M2 (utility headers) + M4 (matrix) + M5 (CHANGELOG); M3 (impl headers) and skill cross-links are necessary for cross-SPEC consistency.

---

## Definition of Done (DoD)

The implementation is "done" when ALL of the following are true:

1. ✅ `agentless_audit_test.go` exists with 3 test functions, and `go test ./internal/template/ -run TestAgentless` passes (3 of 3 GREEN).
2. ✅ All 5 utility skill files contain a `## Pipeline Contract (Agentless Classification)` section with `MODE_FLAG_IGNORED_FOR_UTILITY` sentinel string.
3. ✅ All 4 implementation skill files contain a `## Mode Flag Compatibility` section with `MODE_PIPELINE_ONLY_UTILITY` sentinel string.
4. ✅ `.claude/rules/moai/workflow/spec-workflow.md` contains a `## Subcommand Classification (Pipeline vs Multi-Agent)` section with the 9-row matrix.
5. ✅ `internal/template/templates/.claude/...` mirrors all 11 skill/rule edits (embedded template parity).
6. ✅ `make build` regenerates `internal/template/embedded.go` cleanly.
7. ✅ Full repository test suite passes: `go test ./...` returns 0 (per `CLAUDE.local.md` §6 HARD rule).
8. ✅ CHANGELOG `## [Unreleased]` section has the SPEC-V3R2-WF-004 entry.
9. ✅ MX tags per plan.md §6 inserted in all 8 target locations.
10. ✅ `progress.md` updated with `run_complete_at` and `run_status: implementation-complete`.

---

End of acceptance.md.
