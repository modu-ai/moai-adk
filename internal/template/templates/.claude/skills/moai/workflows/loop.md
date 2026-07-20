---
description: >
  Project-wide improvement sweep, expressed as a goal preset built ON the goal
  engine. A scan stage builds a finite issue queue (LSP + lint + test failures +
  review lenses); the loop delegates the iterate-until-done decision to the goal
  engine (stop-goal) whose condition is "queue drained + diagnostics clean",
  bounded by an iteration ceiling. Includes memory pressure detection and
  snapshot-based resume. Use when a bounded project-wide fix sweep is needed.
user-invocable: false
metadata:
  version: "2.6.0"
  category: "workflow"
  status: "active"
  updated: "2026-07-12"
  tags: "loop, goal-preset, sweep, scan-lens, diagnostics, testing, coverage"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["loop", "iterate", "repeat", "until done", "keep fixing", "all errors"]
  agents: ["manager-develop"]
  phases: ["loop"]
---

# Workflow: Loop - Project-Wide Improvement Sweep (goal preset)

Purpose: `/moai loop` is a **goal preset** — a project-wide improvement sweep built ON the goal engine (the condition-declared loop delivered by the goal engine). A scan stage supplies a FINITE issue queue; the loop then runs as a goal preset whose completion condition is "**issue queue drained** AND **diagnostics clean** (zero errors + tests passing [+ coverage when enabled])", bounded by an iteration ceiling. The loop does NOT reimplement "iterate until done" — it delegates that mechanism to the goal engine and supplies only the preset's scan-derived queue + completion condition.

Flow: Scan -> build finite queue -> (goal preset) Check Completion -> Memory Check -> Diagnose -> Fix -> Verify -> Repeat

<!-- @MX:NOTE - alias relationship: /moai loop and /moai run --mode loop are equivalent. The alias is enforced by a CI audit that checks the cross-reference text is present. -->

## Invocation Routes

This skill is invocable via two equivalent routes:
- Direct: `/moai loop $ARGUMENTS` — historical entry point, preserved as thin wrapper.
- Via run dispatch: `/moai run --mode loop` — the `/moai run` skill delegates to this skill
  (e.g., `/moai run SPEC-XXX --mode loop`) when supplied.

Both routes invoke this skill body unchanged. Behavioral equivalence is enforced by a CI audit
that verifies this skill documents the `/moai run --mode loop` cross-reference.

See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the full pipeline-vs-multi-agent + mode-axis contract.

## Relationship to the Pipeline-Level Agentic Completion Loop

This skill is the Ralph engine — the specialized DIAGNOSTIC fix-loop (a per-iteration cycle over LSP / AST-grep / test / coverage diagnostics). It is distinct from the pipeline-level agentic completion loop (`workflows/moai.md` § Agentic Completion Loop), which iterates over PHASES (run → sync → verify) against a completion condition. The pipeline-level loop MAY invoke this engine during its verify step for mechanical convergence; the two loops are complementary, not competitors, and are NOT folded into one — the granularity differs (phases there, diagnostics here). The `/moai loop` ≡ `/moai run --mode loop` alias contract above is unchanged.

## Loop Taxonomy Position — goal engine + presets

The loop taxonomy is re-expressed as **goal engine + preset**: the four quadrants are not four independent engines but presets layered on the one goal engine. `/moai loop` occupies the **goal-based** quadrant as the **project-wide sweep preset** — a **goal-preset** whose completion condition is "issue queue drained AND diagnostics clean", bounded by the applied iteration ceiling.

- **How it starts**: a `/moai loop` (or `/moai run --mode loop`) invocation runs the scan stage, then arms the goal preset.
- **How it ends**: success-exit when the goal engine confirms "queue drained + diagnostics clean" via the independent final pass (Step 1/1.5), or a ceiling exit that emits the 5-section verdict and persists residue per § Ceiling-Exit Verdict Contract.
- **When it fits**: sweeping a project-wide finite issue queue to zero across many iterations — not a one-off turn, not a schedule.

Sibling presets (same **goal engine + preset** framing, different quadrant): **turn-based** one-shot fixing is `.claude/skills/moai/workflows/fix.md` (its unresolved residue persists to the same verdict schema and recommends re-entry here); **time-based** cadence recipes are `.claude/rules/moai/workflow/cadence-bridge.md`; **proactive** CI-triggered watch is the `moai-workflow-ci-loop` skill. The `goal engine + preset` framing is consistent across the `loop.md`/`fix.md` sibling quadrant notes.

## Goal-Preset Composition — how the sweep is built ON the goal engine

`/moai loop` does NOT reimplement "iterate until done". It **delegates to the goal engine** and supplies only the preset's queue + completion condition:

1. **Scan → queue**: the scan stage (§ Scan Stage below) builds a FINITE issue queue.
2. **Arm the goal preset**: the sweep is expressed as a goal condition — "issue queue drained AND diagnostics clean (zero errors + tests passing [+ coverage when enabled])" — equivalent to running `/moai goal "loop-sweep: issue queue drained + diagnostics clean, or stop after N iterations"`. The loop is a preset that pre-fills this condition rather than asking the user to author it; the orchestrator arms it via `/moai goal` on the user's behalf as the loop preset.
3. **Goal engine evaluates each turn-end**: after each iteration the **goal engine evaluates** the pre-filled condition through the `stop-goal` Stop-hook evaluator (`moai hook stop-goal`) — the same per-turn-end evaluation the goal engine performs for any `/moai goal`. When the condition holds the engine stops the sweep; otherwise the next iteration runs. The loop thus **delegates to the goal engine** for the iterate-until-done decision.
4. **Bounded**: the goal preset is additionally bounded by the iteration ceiling (§ Iteration-Ceiling Precedence) and the memory-pressure guard, so the sweep terminates even when the queue cannot fully drain.

Because the sweep **delegates to the goal engine** (`stop-goal`) for the iterate-until-done mechanism, this preset builds NO new engine code — it CONFIGURES a preset on top of the existing goal engine. The Step 1/Step 1.5 mechanical predicate + independent final pass below is the loop-preset's own mechanical confirmation that runs alongside the goal-engine evaluation.

## Scan Stage — Finite Issue Queue (no invented improvements)

Before arming the goal preset, the scan stage builds a **FINITE** issue queue from a fixed set of lenses. The queue is the sweep's entire work surface — the loop fixes exactly what the scan enqueued and does **no invented improvements** beyond it.

**Default lenses** (always scanned):
- **LSP diagnostics** — type / compile errors for the detected language.
- **lint** — the language linter (golangci-lint, ruff, eslint, …).
- **test failures** — failing tests from the detected test runner.
- **review lenses** — the loop **consumes review** lenses as queue SUPPLIERS. Two review lenses supply queue items: the **security lens** (OWASP / injection / secrets findings) and the **@MX lens** (`@MX` tag-compliance findings). Each review lens produces findings independently; the scan enqueues them as fixable issues. This is the layered relationship with `/moai review` (§ Relationship to /moai review and /moai fix below): standalone review REPORTS, the loop CONSUMES its lenses as queue suppliers.

**Opt-in lenses** (`--lens clean|simplify|coverage`, comma-separated):
- `--lens clean` — dead-code findings (workflows/clean.md scan).
- `--lens simplify` — over-engineering findings (the `/moai review --lean` 5-tag scan).
- `--lens coverage` — coverage-gap findings. This lens is opt-in and coverage-conditional: it supplies queue items only when a coverage gate is enabled; with coverage disabled the lens supplies no queue items (edge case documented).

**HARD boundary — no invented improvements**: the loop performs NO work **outside the scanned queue**. It does not add refactors, features, or cleanups that the scan lenses did not surface. Every fix maps to a scanned queue item; anything else is out of scope for the sweep.

**Empty queue → immediate exit**: when the scan yields an **empty queue** (no lens produced a finding), the loop takes an **immediate exit** — it does NOT arm the goal preset, does NOT burn an iteration, and does NOT enter the per-iteration cycle. An empty queue is success with zero work.

## Relationship to /moai review and /moai fix

**`/moai review` (read-only, report-only)**: standalone `/moai review` REMAINS read-only and report-only — it produces findings and modifies nothing. This SPEC does not change its behavior. `/moai loop` **consumes review** lenses (security, `@MX`) as queue SUPPLIERS: the review lens produces findings, the loop enqueues them and drives the queue to drain. Run a review to SEE findings; run a loop to FIX the finite set the scan (including review lenses) found. The layering is documented from the review side in `.claude/skills/moai/workflows/review.md`.

**`/moai fix` (turn-based sibling preset)**: `/moai fix` is UNCHANGED (single-pass Agentless pipeline). Its residue-handoff persists to the same `loop-verdict-<id>.json` schema, and that residue **enters the loop queue** when the user re-enters `/moai loop`.

## Supported Flags

- --max N (alias --max-iterations): Maximum iteration count. When absent, the effective default is ralph.yaml `loop.max_iterations` (shipped 10) per the Iteration-Ceiling Precedence rule (see § Ceiling-Exit Verdict Contract) — not a freestanding 100 default.
- --auto-fix: Enable auto-fix (default Level 1)
- --sequential (alias --seq): Sequential diagnostics instead of parallel
- --errors (alias --errors-only): Fix errors only, skip warnings
- --coverage (alias --include-coverage): Include coverage threshold (default 85%)
- --memory-check: Enable memory pressure detection
- --resume ID (alias --resume-from): Restore from snapshot

## Per-Iteration Cycle

Each iteration executes the following steps in order:

Step 1 - Completion Predicate Check (mechanical, not sentinel-based):
- Re-evaluate the previous iteration's PARSED Step-3 diagnostics (exit codes, error count, test pass/fail, coverage percentage) against ralph.yaml `loop.completion` (zero_errors, tests_pass, coverage_threshold, zero_warnings). The mechanical read surface is the shared diagnostic snapshot under `.moai/state/verify/` (written by Step 3 via `moai verify record` — the shared-schema formalization of the former Step-8-only persistence); the Step 8 iteration snapshot under `.moai/cache/loop-snapshots/` remains the resume artifact.
- The completion sentence "All loop completion conditions satisfied; exiting loop." is DISPLAY-ONLY (emitted by Step 4 as a report string) — it carries no exit authority. Do NOT exit on detecting this sentence in the previous response; the mechanical predicate re-evaluation above is the only exit-eligible signal.
- If the mechanical predicate holds (all loop.completion conditions satisfied on the previous iteration's parsed diagnostics): proceed to Step 1.5 (Independent Final Pass) before declaring success-exit.
- If the predicate does not hold: continue the loop (Step 2 onward) as a normal iteration.

Step 1.5 - Independent Final Pass (entered only when Step 1's predicate holds):
- Execute an independent verification pass that is NOT the loop executor's own claim — a fresh-context re-run of the diagnostic gate, distinct from this loop's own Step 3 measurement.
- Independence carve-out: Step 1.5 shall NOT consume a snapshot produced by the same loop run — directly OR transitively through a consuming layer. The independent pass exists to keep success-exit evidence non-self-referential; reusing the loop's own Step-3 snapshot (whose key is still fresh at success-exit, the tree being unchanged) would make the confirmation circular.
- Primary vehicle: re-run the gate in force-fresh mode — `/moai gate --fresh` — a fresh mechanical gate invocation, independent of this loop's iteration state. Because the gate is itself a snapshot consumer, the `--fresh` flag is what closes the gate-mediated path: it disables ALL snapshot consumption for that invocation, so the same-run Step-3 snapshot cannot flow back through the gate layer. Step 1.5's own force-fresh gate results MAY be recorded for downstream consumers (e.g. sync Phase 0).
- Fallback vehicle (when `/moai gate` is unavailable in the environment): spawn a read-only verifier `Agent()` (no Write/Edit tools) to re-run the same diagnostic commands and report results. The verifier never prompts the user (subagent boundary — blocker reports only).
- Divergence rule: if the independent pass's observed diagnostics (error count, test result, coverage) diverge from the builder-observed (Step 3) diagnostics, the loop does NOT exit with success — it continues to the next iteration, or escalates per the existing no-progress escalation rule if the same divergence repeats across consecutive checks.
- Degradation (independent pass unavailable): do NOT silently claim full verification. Record the gap explicitly (Gaps section of the eventual exit report — see § Completion Conditions) and continue to the next iteration rather than exit with an unconfirmed success.
- Only after the independent pass CONFIRMS the predicate: declare success-exit. Proceed to the pre-exit clean sweep and final report.

Step 2 - Memory Pressure Check (if --memory-check enabled):
- Calculate session duration from start time
- Monitor iteration time for GC pressure signs (doubling iteration time)
- If session duration exceeds 25 minutes OR iteration time doubling:
  - Save proactive checkpoint to $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/memory-pressure.json
  - Warn user about memory pressure
  - Suggest resuming with /moai:loop --resume memory-pressure
- Physical memory check (if memory_guard.enabled in quality.yaml):
  - Linux: Read MemAvailable from `free -m`
  - macOS: Estimate from `vm_stat` pages free * page_size
  - If available_mb < emergency_threshold_mb:
    - Save checkpoint to $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/memory-emergency.json
    - Exit loop with message: "System memory critically low ({available_mb}MB < {emergency_threshold_mb}MB). Checkpoint saved."
    - Suggest: /moai:loop --resume memory-emergency
  - If available_mb < adaptive_threshold_mb:
    - Switch Step 3 diagnostics to module-split execution for remaining iterations
    - Log: "Low memory detected ({available_mb}MB). Switching to module-split test execution."
- If memory-safe limit reached (50 iterations): Exit with checkpoint

Step 3 - Parallel Diagnostics:
- Launch four diagnostic tools simultaneously using Bash with run_in_background
- Tool 1: LSP diagnostics for detected language
- Tool 2: AST-grep scan with sgconfig.yml rules
- Tool 3: Test runner for detected language (pytest, jest, go test, cargo test)
- Tool 4: Coverage measurement (coverage.py, c8, go test -cover, cargo tarpaulin)
- Collect results using Read on each background task's output file path
- Aggregate into unified diagnostic report with metrics: error count, warning count, test pass rate, coverage percentage
- Record the parsed diagnostics into the shared diagnostic snapshot via `moai verify record` (one entry per diagnostic command: exact command string, exit code, parsed counts) — the mechanical writer for the shared schema. Step 1's completion predicate re-reads this snapshot; downstream consumers (the gate, sync Phase 0) may also reuse it while it stays fresh (key equality AND TTL). Every reuse cites the snapshot path, key, original command, and recorded exit code as its evidence, and a stale snapshot is never cited (per `.claude/rules/moai/core/verification-claim-integrity.md` §2 — the snapshot is the observed evidence; the freshness rule keeps the attribution valid).

If --sequential flag: Run LSP, then AST-grep, then Tests, then Coverage sequentially.

Step 4 - Completion Condition Check:
- Conditions: Zero errors AND all tests passing AND coverage meets threshold (ralph.yaml `loop.completion`)
- If all conditions met: Persist the parsed diagnostics with the Step 8 iteration snapshot and display the completion sentence "All loop completion conditions satisfied; exiting loop." as a REPORT-ONLY string — it documents the result for the transcript but does NOT itself terminate the loop. Step 1 of the next iteration re-evaluates the persisted diagnostics mechanically (per Step 1) and, if satisfied, runs the independent final pass (Step 1.5) before any success-exit.
- If only coverage below target (zero errors + tests passing): route coverage-gap handling through `go test -cover` gap analysis + `/moai gate` (the documented coverage replacement path) for intelligent gap analysis and test generation instead of blind looping. Identify P1-P4 priority gaps and generate targeted tests.

Step 5 - Task Generation:
- [HARD] Batched bookkeeping — at most ONE batched turn per iteration: consolidate ALL Task-tool bookkeeping for the iteration into a single turn of parallel Task-tool calls, with aggregate tasks per file/group. The batch covers TaskCreate for all newly discovered issues (pending status), the TaskUpdate `in_progress` transitions for the fixes this iteration starts, and the TaskUpdate `completed` transitions for the fixes it verifies — replacing the former 3-calls-per-issue cadence. The same task entries and status transitions are produced; only the call cadence changes (no record is dropped).

Step 5.5 - Pre-Fix MX Context Scan:
- Scan files with newly discovered issues for existing @MX tags
- @MX:ANCHOR functions: Pass as "do not break" constraints to fix agents
- @MX:WARN zones: Pass danger context; ensure fix does not worsen the warned condition
- @MX:NOTE context: Provide business logic understanding before modification
- @MX:TODO items: Match against current issues for resolution tracking
- Output: MX context map included in Step 6 fix agent prompts
- Skip if no @MX tags found in target files
- See .claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions

Step 6 - Fix Execution:
- Status bookkeeping for started fixes rides the Step-5 per-iteration batch (`in_progress` transitions — no per-fix Task-tool call)
- [HARD] Agent delegation mandate (Level 2+): ALL Level 2 and above fix tasks MUST be delegated to specialized agents. NEVER execute Level 2+ fixes directly. Level 1 (import sorting, whitespace, formatting) is exempt: the orchestrator runs the language's deterministic formatter command directly, without an Agent() spawn (the same Level-1 orchestrator-direct exception as fix.md Phase 3).

Agent selection by issue type (domain expertise injected per-spawn per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C):
- Type errors, logic bugs: manager-develop subagent (or orchestrator verification batch)
- Import/module issues: manager-develop (or per-spawn `Agent(general-purpose)` backend/frontend specialist) — inject `At start, invoke Skill("moai-ref-api-patterns")` for backend import/module work
- Test failures: manager-develop subagent — inject `At start, invoke Skill("moai-ref-testing-pyramid")` for test-suite structure
- Security issues: per-spawn `Agent(general-purpose)` security reviewer — inject `At start, invoke Skill("moai-ref-owasp-checklist") for the OWASP Top 10 baseline.`
- Performance issues: per-spawn `Agent(general-purpose)` performance specialist

Skill injection: per skill-routing.md §1, before each spawn inject 0-3 `At start, invoke Skill("<name>") for <reason>` lines from the delegation map (`.moai/config/sections/delegation.yaml`), matched to the issue type's domain (security → moai-ref-owasp-checklist, test → moai-ref-testing-pyramid, backend import/module → moai-ref-api-patterns).

Fix levels applied per --auto setting:
- Level 1 (Immediate): No approval. Import sorting, whitespace
- Level 2 (Safe): Log only. Rename variable, add type
- Level 3 (Approval): AskUserQuestion required. Logic change, API modify
- Level 4 (Manual): Not auto-fixed. Security, architecture

Step 7 - Verification:
- Status bookkeeping for verified fixes rides the Step-5 per-iteration batch (`completed` transitions — no per-fix Task-tool call)

Step 7.5 - MX Tag Check:
- After fixes applied, scan modified files for MX tag requirements
- Add missing tags for modified functions:
  - New exported functions: Add @MX:NOTE or @MX:ANCHOR if fan_in >= 3
  - Dangerous patterns introduced: Add @MX:WARN with @MX:REASON
  - Unresolved issues: Keep @MX:TODO
- Remove resolved @MX:TODO tags for fixed issues
- Per-iteration tag ADD/REMOVE/UPDATE actions still apply here; the MX_TAG_REPORT itself is emitted ONCE at sweep exit, aggregated across all iterations (see § MX Tag Report below) — not per iteration
- See .claude/rules/moai/workflow/mx-tag-protocol.md for tag rules

Step 8 - Snapshot Save:
- Save iteration snapshot to $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/
- Increment iteration counter

Step 9 - Repeat or Exit:
- If max iterations reached: emit the ceiling-exit 5-section verdict per § Ceiling-Exit Verdict Contract (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), persist remaining issues to `.moai/state/loop-verdict-<id>.json` (or mirror into TaskList when a team task ledger is active), and propose a lesson-capture entry (per moai-constitution.md § Lessons Protocol) before ending the session.
- Otherwise: Return to Step 1

## Completion Conditions

The loop exits when any of these conditions are met:
- Mechanical predicate confirmed twice: the previous iteration's parsed diagnostics satisfy zero errors + tests passing + coverage threshold (Step 1) AND the independent final pass (Step 1.5) confirms the same result — this is the ONLY success-exit path. The completion sentence is a display-only report string and carries no exit authority on its own.
- Max iterations reached: emits the ceiling-exit 5-section verdict (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per verification-claim-integrity.md §3) and persists remaining issues (see § Ceiling-Exit Verdict Contract)
- Memory pressure threshold exceeded (saves checkpoint)
- User interruption (state auto-saved)

## Ceiling-Exit Verdict Contract

When the loop exits at the iteration ceiling (Step 9) — or exits with Level-4 manual items still outstanding — the loop workflow emits a structured 5-section evidence report per `verification-claim-integrity.md` §3, using the section names verbatim:

- **Claim**: the loop did not reach the mechanical completion predicate within the applied ceiling.
- **Evidence**: the final iteration's parsed diagnostics (error count, test result, coverage percentage) — verbatim Step-3 output, not a summary.
- **Baseline-attribution**: the ceiling that was applied and its source (CLI `--max` flag / ralph.yaml `loop.max_iterations` / workflow.yaml `loop_prevention.max_iterations` — see § Iteration-Ceiling Precedence), plus the iteration count consumed.
- **Gaps**: conditions not yet satisfied (e.g. "3 errors remaining", "coverage 78% < 85% threshold"), and any independent-pass degradation recorded during the run (per Step 1.5).
- **Residual-risk**: unresolved Level-4 manual items, flaky-test risk, or environment-specific findings that could still be wrong.

This same 5-section report applies to `workflows/moai.md` § Agentic Completion Loop termination cause 2 (iteration ceiling) — closing its protocol gap relative to causes 3 (escalation) and 4 (context-threshold suspension), which already carry structured protocols.

### Remaining-Issue Persistence

The loop workflow persists remaining issues to `.moai/state/loop-verdict-<id>.json` (`<id>` = session- or timestamp-derived identifier), or mirrors into TaskList when a team task ledger is active (TaskList mirroring is additive; the state file is the always-on floor). Transcript-only residue is prohibited — every ceiling exit MUST leave an auditable, resumable artifact on disk.

Minimum JSON schema (doctrine-defined; orchestrator-written at exit time — no Go loader in this contract):

```
{
  "spec_or_scope": "<SPEC-ID or free-form scope label>",
  "exit_kind": "ceiling | manual-residue | one-shot-residue | sweep-residue",
  "iterations_used": <int>,
  "ceiling_applied": <int>,
  "ceiling_source": "flag | ralph | loop_prevention",
  "conditions": {
    "zero_errors": <bool>, "error_count": <int>,
    "tests_pass": <bool>,
    "coverage_threshold": <int>, "coverage_actual": <number>,
    "zero_warnings": <bool>
  },
  "remaining_issues": [
    {"severity": "P1|P2|P3|P4", "description": "...", "file": "path:line", "suggested_action": "..."}
  ],
  "vci_report_ref": "<pointer to the 5-section report in the transcript or a saved file>",
  "created_at": "<ISO-8601 timestamp>"
}
```

> **`exit_kind` value ownership (additive)**: `loop.md` owns the base `ceiling | manual-residue`; `fix.md` added `one-shot-residue` (its one-shot pipeline exit path); the project-wide **sweep preset** adds `sweep-residue` ADDITIVELY as a fourth value — without reassigning the base enum owner. **When** `/moai loop` (the sweep preset) exits at the iteration ceiling with scanned-queue items still outstanding, it persists `"exit_kind": "sweep-residue"` (the settled value alongside the base `ceiling | manual-residue | one-shot-residue` enum), so a downstream backlog re-discovery reader can distinguish a sweep-residue exit from a plain ceiling exit.

When `.moai/state/` is unwritable at exit, surface the verdict content in-conversation AND name the write failure explicitly in the Residual-risk section — the loop fails open on persistence, never on honesty.

### Lesson-Capture Proposal (unsuccessful exit)

When the loop exits unsuccessfully (ceiling reached with conditions unmet), propose a lesson-capture entry per `moai-constitution.md` § Lessons Protocol before session close — the failure pattern (which condition stalled, why) is offered for memory capture rather than silently dropped.

## Iteration-Ceiling Precedence

Iteration-ceiling precedence: CLI `--max` flag > ralph.yaml `loop.max_iterations` > workflow.yaml `loop_prevention.max_iterations`. The memory-safe 50-iteration checkpoint (Step 2) is an orthogonal memory-pressure safeguard, not a fourth ceiling.

Pre-exit clean sweep (when exiting with success):
- Before final report, run clean workflow (workflows/clean.md) scan on all modified files
- Remove dead code exposed by fixes (unused imports, orphaned functions)
- Skip if no dead code detected or if --errors flag was set

## MX Tag Integration

Each iteration includes MX tag management:

**Tag Updates During Loop:**
- Fix resolves an issue: Remove corresponding @MX:TODO
- Fix introduces new code: Add appropriate @MX tags
- Fix changes function signature: Re-evaluate @MX:ANCHOR
- Fix adds complexity: Add @MX:WARN if threshold exceeded

**Tag Types for Fixes:**
| Fix Type | MX Action |
|----------|-----------|
| Bug fix (resolved) | Remove @MX:TODO |
| New function added | Add @MX:NOTE or @MX:ANCHOR |
| Refactoring | Update @MX:NOTE, check ANCHOR |
| Security fix | Add @MX:NOTE with security context |

**MX Tag Report (aggregate, once at sweep exit):**
When the sweep exits (success, ceiling, or interruption), emit the MX_TAG_REPORT ONCE, aggregated across all iterations. Per-iteration tag ADD/REMOVE/UPDATE actions still apply during Step 7.5 — only the report cadence changes, and no record is dropped:
- Tags Added: List new tags with file:line (all iterations)
- Tags Removed: List resolved TODOs (all iterations)
- Tags Updated: List modified tags (all iterations)
- Attention Required: WARN tags requiring review

## Snapshot Management

> **Best-effort**: snapshot persistence and `--resume` re-entrancy are best-effort orchestrator behaviors — the `.moai/cache/loop-snapshots/` directory has no mechanical writer guarantee and may never materialize. Do not rely on it for persistent state.

Snapshot location: $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/

Files:
- iteration-001.json, iteration-002.json, etc. (per-iteration snapshots)
- latest.json (symlink to most recent)
- memory-pressure.json (proactive checkpoint on memory pressure)
- memory-emergency.json (emergency checkpoint when physical memory critically low)

Loop state file: $CLAUDE_PROJECT_DIR/.moai/cache/.moai_loop_state.json

Resume commands:
- /moai:loop --resume latest
- /moai:loop --resume iteration-002
- /moai:loop --resume memory-pressure

## Language-Specific Commands

Test runner and coverage tool selection is based on auto-detected project language:

All 16 MoAI-supported languages are listed alphabetically with equal
priority. Per `.claude/rules/moai/development/coding-standards.md` § Language Policy
(16-language neutrality contract), no language receives PRIMARY
or SECONDARY classification.

| Language   | Indicator File           | Test Command                                          | Coverage Command                                  |
|------------|--------------------------|-------------------------------------------------------|---------------------------------------------------|
| C#         | `*.csproj`               | `dotnet test`                                         | `dotnet test --collect:"XPlat Code Coverage"`     |
| C++        | `CMakeLists.txt`         | `ctest --test-dir build`                              | `gcov` / `lcov` (if configured)                   |
| Elixir     | `mix.exs`                | `mix test`                                            | `mix test --cover`                                |
| Flutter    | `pubspec.yaml`           | `flutter test` or `dart test`                         | `flutter test --coverage`                         |
| Go         | `go.mod`                 | `go test ./...`                                       | `go test -cover ./...`                            |
| Java       | `pom.xml` / `build.gradle` | `mvn test -q` (Maven) / `gradle test -q` (Gradle)  | `mvn jacoco:report` / `gradle jacocoTestReport`   |
| JavaScript | `package.json`           | `npm test` or `jest`                                  | `npm run coverage` or `c8`                        |
| Kotlin     | `build.gradle.kts`       | `gradle test -q`                                      | `gradle jacocoTestReport`                         |
| PHP        | `composer.json`          | `vendor/bin/phpunit`                                  | `vendor/bin/phpunit --coverage-text`              |
| Python     | `pyproject.toml` / `setup.py` | `pytest --tb=short`                              | `coverage run -m pytest`                          |
| R          | `DESCRIPTION`            | `Rscript -e 'testthat::test_package(".")'`            | `covr::package_coverage()`                        |
| Ruby       | `Gemfile`                | `bundle exec rspec` or `bundle exec rake test`        | `simplecov` (via `.simplecov` config)             |
| Rust       | `Cargo.toml`             | `cargo test`                                          | `cargo tarpaulin`                                 |
| Scala      | `build.sbt`              | `sbt test`                                            | `sbt coverage test coverageReport`                |
| Swift      | `Package.swift`          | `swift test`                                          | `swift test --enable-code-coverage`               |
| TypeScript | `tsconfig.json` / `package.json` | `npm test` or `jest`                          | `npm run coverage` or `c8`                        |

Language detection priority: Check for indicator files in project root. If multiple present, prefer the one with the most associated source files. If detection fails, prompt user to specify language.

## Cancellation

Send any message to interrupt the loop. State is automatically saved via session_end hook.

## Safe Development Protocol

All fixes within the loop follow CLAUDE.md Section 7 Safe Development Protocol:
- Reproduction-first: Write failing tests before fixing bugs
- Post-fix review: List potential side effects after each fix cycle
- Maximum 3 retries per individual operation (per CLAUDE.md constitution)

## Execution Summary

1. Parse arguments (extract flags: --max, --auto-fix, --sequential, --errors, --coverage, --memory-check, --resume)
2. If --resume: Load state from specified snapshot and continue
3. Detect project language from indicator files
4. Initialize iteration counter and memory tracking (start time)
5. Loop: Execute per-iteration cycle (Steps 1-9, including Step 1.5 Independent Final Pass and Step 5.5 MX Context Scan)
6. On exit: Report final summary with evidence
7. If memory checkpoint created: Display resume instructions

---

Version: 2.6.0
Updated: 2026-07-12. Redefined `/moai loop` as a **goal preset** — a project-wide improvement sweep built ON the goal engine. Added the Goal-Preset Composition section (delegates the iterate-until-done decision to the goal engine via `stop-goal`), the Scan Stage finite-issue-queue section (default LSP + lint + test + review lenses [security, @MX], opt-in `--lens clean|simplify|coverage`, no-invented-improvements HARD boundary, empty-queue immediate exit), the /moai review + /moai fix layering section, and the additive `sweep-residue` exit_kind value. PRESERVED: the mechanical predicate + independent final pass (Step 1/1.5), the ceiling-exit 5-section verdict contract with `.moai/state/loop-verdict-<id>.json` persistence, the iteration-ceiling precedence rule, and the memory-pressure guard. Previous: 2.3.0 (2026-07-09) replaced sentinel-string success-exit with mechanical predicate; 2.2.0 (2026-03-02) expanded Language-Specific Commands to 16 languages.
