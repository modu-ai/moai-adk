---
description: "Run Phase 0.5~1.8 — Plan Audit Gate, environment assessment, JIT language detection, scale-based mode, analysis/planning, task decomposition, and development mode routing"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Phase 0.5-1.8: Phase Sequence and Preparation"
---

# Phase Sequence

All phases execute sequentially. Each phase receives outputs from all previous phases as context.

## Phase 0.5: Plan Audit Gate

**Purpose**: Mandatory independent audit of plan artifacts before any implementation begins.
Prevents unreviewed, incomplete, or non-compliant SPEC documents from entering Phase 1.
Source: the plan audit gate contract.

**Scope**: Every `/moai run <SPEC-ID>` invocation. Never skipped — not even in `minimal` harness.

### Step 1: Compute Plan Artifact Hash (Cache Key)

Compute a combined SHA-256 hash of all plan artifacts present in `.moai/specs/<SPEC-ID>/`:
- `spec.md` (required)
- `plan.md` (if present)
- `acceptance.md` (if present)
- `tasks.md` (if present)

Hash algorithm: SHA-256 of the UTF-8 content of each file, sorted by filename, concatenated.
Whitespace normalization: collapse all runs of whitespace to a single space before hashing (whitespace-insensitive cache).
Store hash as `plan_artifact_hash` for Step 2 cache lookup.

### Step 2: Check 24-Hour Audit Cache

Read `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md` (today's date).

Cache HIT conditions (all must be true):
1. File exists and the most recent audit run entry has `verdict: PASS`
2. `plan_artifact_hash` in the cached entry matches `plan_artifact_hash` from Step 1
3. `audit_at` timestamp in the cached entry is within 24 hours of now (UTC)

If cache HIT:
- Log: `[plan-audit] cache hit (verdict=PASS, age=<Nh>)`
- Append to `.moai/specs/<SPEC-ID>/progress.md`: `- audit_cache_hit: true` and `- cached_audit_at: <T0>`
- Skip Step 3 and proceed to Phase 1.

If cache MISS: proceed to Step 3.

### Step 3: Invoke plan-auditor Subagent

Delegation pattern (single invocation, main session only):

> "Use the plan-auditor subagent to audit the SPEC at `.moai/specs/<SPEC-ID>/` — plan artifacts:
> spec.md, plan.md, acceptance.md, tasks.md (if present). Produce a verdict (PASS/FAIL) and
> write the report to `.moai/reports/plan-audit/<SPEC-ID>-review-<iteration>.md`."

Do NOT pass the implementation context or any prior conversation to plan-auditor.
plan-auditor enforces context isolation automatically.

Timeout: 60 seconds. On timeout, treat as INCONCLUSIVE (Step 4d).

### Step 4: Verdict Routing (4-Way Branch)

Read the verdict from the report file produced by plan-auditor.

**4a. PASS**
- Log: `[plan-audit] verdict=PASS, persisted to progress.md, proceeding to Phase 1`
- Proceed to Step 5 (persist), then continue to Phase 1.

**4b. FAIL (and grace window ACTIVE — today < merge_date + 7 days)**
- Log: `[plan-audit] verdict=FAIL [grace-window], D-<N> until auto-block`
- Grace window start: read from `.moai/state/audit-gate-merge-at.txt` (ISO-8601).
- If `MOAI_AUDIT_GATE_T0` env var is set, use it as T0 override (test injection).
- Emit warning to stdout: `[grace-window] D-<N> (auto-block activates at T0+7)`
- Record `audit_verdict: FAIL_WARNED` in progress.md.
- Proceed to Phase 1 (warn-only, not blocked).

**4c. FAIL (and grace window EXPIRED — today >= merge_date + 7 days)**
- Log: `[plan-audit] verdict=FAIL, blocking Run phase, report=<path>`
- Surface the audit report path and the list of must-pass failures to stdout.
- Do NOT proceed to Phase 1 automatically.
- [HARD] Present options to user via AskUserQuestion (orchestrator responsibility):
  - Option 1 (Recommended): Revise SPEC — fix the defects, then re-run `/moai run`
  - Option 2: Override and proceed — skip the gate (sets `--skip-audit` implicitly, records BYPASSED)
  - Option 3: Abort — exit without any implementation

**4d. INCONCLUSIVE (timeout, malformed output, error, or filesystem failure)**
- Log: `[plan-audit] verdict=INCONCLUSIVE, falling back to manual prompt`
- Record `verdict: INCONCLUSIVE` in the daily report.
- [HARD] Do NOT auto-PASS. Present options to user via AskUserQuestion:
  - Option 1 (Recommended): Retry audit — re-invoke plan-auditor (max 3 retries total)
  - Option 2: Proceed with acknowledgement — user accepts responsibility; records `inconclusive_acknowledged_by: <user.name>`
  - Option 3: Abort — exit without any implementation

### Step 5: Persist Verdict and Daily Report

**progress.md** (append to `.moai/specs/<SPEC-ID>/progress.md`):
```yaml
- audit_verdict: PASS          # or FAIL_WARNED, BYPASSED, INCONCLUSIVE
- audit_report: .moai/reports/plan-audit/<SPEC-ID>-review-<N>.md
- audit_at: <ISO-8601 UTC>
- auditor_version: plan-auditor v<version>
```

**Daily report** (append to `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md`):
Path is always `.moai/reports/plan-audit/` + SPEC-ID + `-` + current date (YYYY-MM-DD) + `.md`.
[HARD] Path must be validated with `filepath.Clean` and confirmed to reside inside the project's `.moai/reports/plan-audit/` directory (path traversal prevention).

Append format (each audit run is a numbered section):
```markdown
## Audit Run <N> of <total>

- verdict: PASS | FAIL | BYPASSED | INCONCLUSIVE | FAIL_WARNED
- report_path: .moai/reports/plan-audit/<SPEC-ID>-review-<N>.md
- audit_at: <ISO-8601 UTC>
- run_trigger: automatic | manual | bypassed | inconclusive
- plan_artifact_hash: <hash>
- auditor_version: <identifier>
```

When the same SPEC is audited multiple times in the same day, each run appends a new section.

### When --skip-audit Flag Is Provided

When the user passes `--skip-audit` flag OR sets `MOAI_SKIP_PLAN_AUDIT=1` env var:

1. Skip Steps 1–4 entirely.
2. Read user identity from `.moai/config/sections/user.yaml` → `user.name`.
3. Collect bypass rationale:
   - **Interactive** (stdin is a TTY): Collect via AskUserQuestion: "Provide bypass rationale:"
   - **Non-interactive** (stdin is not a TTY, e.g., CI): Auto-record `bypass_reason: "non-interactive"`
4. Sanitize rationale: escape Markdown special characters (`*`, `_`, `[`, `]`, etc.) before writing.
5. Append to daily report:
   ```yaml
   verdict: BYPASSED
   bypass_at: <ISO-8601 UTC>
   bypass_user: <user.name>
   bypass_reason: "<sanitized rationale>"
   ```
6. Proceed to Phase 1 under the user's explicit responsibility.

### When Plan-Auditor Fails or Times Out

Plan-auditor failure cases classified as INCONCLUSIVE:

| Failure Case | Classification | Notes |
|-------------|----------------|-------|
| Timeout (> 60s) | INCONCLUSIVE | Retry up to 3 times total |
| Malformed output / missing verdict field | INCONCLUSIVE | Log raw output for debugging |
| panic / unhandled exception | INCONCLUSIVE | Capture stack trace if available |
| Filesystem write failure (report directory) | INCONCLUSIVE | Falls back to the inconclusive-verdict handling |

[HARD] INCONCLUSIVE is never equivalent to PASS. Automatic pass-through on failure is prohibited.

Retry limit (OPEN QUESTION Q3 resolution): Maximum 3 total plan-auditor invocations per `/moai run`
call (including retries). After 3 INCONCLUSIVE results, force AskUserQuestion with proceed/abort only.

## Phase 0.6: Environment Assessment (Conditional)

Condition: Only executes when `memory_guard.enabled: true` in quality.yaml.
If memory_guard is not enabled or not present, skip to Phase 1.

Purpose: Detect available system memory and determine test execution strategy to prevent OOM.

Steps:
1. Read memory_guard configuration from quality.yaml
2. Detect available system memory:
   - Linux: `free -m | awk '/^Mem:/{print $7}'` (MemAvailable)
   - macOS: `sysctl -n hw.memsize` (total memory in bytes, divide by 1048576 for MB) and `vm_stat | awk '/Pages free/{print $3}'` (approximate available)
3. Compare available memory against thresholds:
   - Below emergency_threshold_mb: BLOCK test execution, warn user, suggest closing other applications or increasing memory
   - Below adaptive_threshold_mb: Set test_execution_strategy to memory_guard.test_split_strategy (default: "module")
   - Above adaptive_threshold_mb: Set test_execution_strategy to "full" (normal execution)
4. Pass test_execution_strategy as context to all subsequent phases via agent prompt

Output: test_execution_strategy ("full", "module", "changed") passed to Phase 1+ as binding context.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 0.6 complete: memory_guard={enabled|disabled}, available_mb={N}, strategy={full|module|changed}
```

<!-- @MX:WARN: [AUTO] Future PRs may be tempted to revert to moai-lang-* skill references here. The current rule-path mapping (post-the language-as-rules policy) MUST remain pointing to .claude/rules/moai/languages/<name>.md. Frontmatter `related-skills:` regressions fail TestRelatedSkillsNoLangReference (DEAD_LANG_FRONTMATTER_REFERENCE); body-prose regressions fail TestSkillBodyNoLangReference (DEAD_LANG_SKILL_REFERENCE). -->
<!-- @MX:REASON: High-traffic section — language detection mapping is frequently referenced by agent authors who may inadvertently reintroduce moai-lang-* skill IDs. -->
## Phase 0.9: JIT Language Skill Detection

Purpose: Detect the project's primary language and prepare the appropriate language skill reference for agent spawn prompts. Since language skills are not statically bound to agents, the orchestrator must inject them at spawn time.

Steps:
1. Check project root for language indicator files:
   - go.mod → `.claude/rules/moai/languages/go.md` (auto-loaded via paths frontmatter)
   - package.json with "typescript" in devDependencies → `.claude/rules/moai/languages/typescript.md`
   - package.json without typescript → `.claude/rules/moai/languages/javascript.md`
   - pyproject.toml or requirements.txt → `.claude/rules/moai/languages/python.md`
   - Cargo.toml → `.claude/rules/moai/languages/rust.md`
   - pom.xml or build.gradle → `.claude/rules/moai/languages/java.md`
   - build.gradle.kts → `.claude/rules/moai/languages/kotlin.md`
   - *.csproj or *.sln → `.claude/rules/moai/languages/csharp.md`
   - Gemfile → `.claude/rules/moai/languages/ruby.md`
   - mix.exs → `.claude/rules/moai/languages/elixir.md`
   - build.sbt → `.claude/rules/moai/languages/scala.md`
   - Package.swift → `.claude/rules/moai/languages/swift.md`
   - pubspec.yaml → `.claude/rules/moai/languages/flutter.md`
   - DESCRIPTION (with R content) → `.claude/rules/moai/languages/r.md`
   - CMakeLists.txt or *.cpp → `.claude/rules/moai/languages/cpp.md`
2. Store the detected language rule path(s) as context for subsequent phases
3. Language rules are auto-loaded via paths frontmatter when project files match; no explicit Skill() invocation required for language rules.
4. If multiple languages detected (e.g., monorepo), all relevant language rules are auto-loaded

Output: detected_language_skills list passed to all subsequent agent spawn prompts.

This phase always executes and does NOT require user approval.

## Phase 0.95: Scale-Based Execution Mode Selection

Purpose: Automatically select the optimal execution mode based on task scope, preventing over-engineering for simple tasks and under-resourcing for complex ones. The scale labels below are ENVELOPES of the Phase 0.95 catalog modes (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A); the label→catalog correspondence is documented in that rule's §G.1 crosswalk (correspondence, not merge).

Mode Selection Rules:

| Request Pattern | Detection Criteria | Execution Mode (catalog correspondence) | Agents |
|----------------|-------------------|---------------|--------|
| Bug fix / error fix | SPEC scope ≤ 3 files, single domain | **Fix Mode** (Mode 5 envelope) | manager-develop + orchestrator verification batch (lint + test + coverage) |
| Single endpoint / function | SPEC scope ≤ 5 files, single domain | **Focused Mode** (Mode 5 envelope) | manager-develop (domain context injected per archived-agent-rejection.md §C) |
| Feature across 1 domain | SPEC scope 5-10 files, single domain | **Standard Mode** (Mode 5 envelope) | manager-spec (planning) + manager-develop + sync-auditor |
| Multi-domain feature | SPEC scope ≥ 10 files OR ≥ 3 domains | **Full Pipeline** (Mode 5 full envelope) | manager-spec → manager-develop (per-spawn `Agent(general-purpose)` domain specialists) → sync-auditor → manager-docs |
| Large cross-cutting change | complexity score at/above the auto-select threshold (`orchestration-mode-selection.md` §B.1) | **Team Mode** (Mode 3 via the §C.1 capability gate) | 3-4 parallel teammates |

Team-row gate note: the Team Mode row resolves through the SAME §C.1 capability gate on BOTH paths — the flag-free auto-select path and a forced `--team` flag both pass the gate; the flag never substitutes for the gate, and the gate never requires the flag.

Detection Steps:
1. Count files referenced in SPEC requirements and plan
2. Identify domains touched (backend, frontend, database, infra, docs)
3. Assess complexity from SPEC priority and acceptance criteria count
4. Select mode based on the table above
5. Log selected mode: "Scale-based mode: {mode} (files: {N}, domains: {N})"

This phase auto-selects and does NOT require user approval. The user can override with --team or --solo flags.

## Phase 1: Analysis and Planning

Agent: manager-spec subagent (planning IS strategy per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 1)

Input: SPEC document content from the provided SPEC-ID. If research.md exists in the SPEC directory (.moai/specs/SPEC-{ID}/research.md), include it as additional context for deeper understanding of the codebase architecture, reference implementations, and identified risks.

Tasks for manager-spec:

- Read and fully analyze the SPEC document
- Extract requirements and success criteria
- Identify implementation phases and individual tasks
- Determine tech stack and dependencies required
- Estimate complexity and effort
- Create detailed execution strategy with phased approach

Output: Execution plan containing plan_summary, requirements list, success_criteria, and effort_estimate.

Implementation guard: [HARD] During Phase 1 (Analysis and Planning), the manager-spec subagent MUST NOT write any implementation code. The explicit instruction "DO NOT implement any code — focus exclusively on analysis and planning" MUST be included in the agent prompt. This separation of thinking and execution prevents premature implementation and ensures the plan is reviewed before any code is written.

## Decision Point 1: Plan Approval

<!-- moai:evolvable-start id="gate-run-1" -->
## HUMAN GATE: Plan Approval

**Previous phase output:** Analysis and implementation plan with task decomposition
**Approval question:** Is the implementation plan correct and complete?
**Cannot proceed until:**
- [ ] Plan covers all SPEC acceptance criteria
- [ ] Task decomposition respects Multi-File Decomposition rule (>3 files = split)
- [ ] User has approved the approach
<!-- moai:evolvable-end -->

Tool: AskUserQuestion (at orchestrator level)

Before presenting options, verify the plan against these criteria:

- Proportionality: Is the plan proportional to the requirements? Flag plans with excessive abstraction layers, unnecessary patterns, or scope creep beyond SPEC requirements.
- Code Reuse: Has the plan identified existing code, libraries, or patterns that can be reused? Flag plans that reinvent existing functionality.
- Reference Implementations: Has the plan leveraged reference implementations from research.md? Flag plans that ignore available reference code in the codebase.
- Simplicity: Does the plan follow YAGNI (You Aren't Gonna Need It)? Flag speculative features not in the SPEC.

Options:

- Proceed with plan (continue to Phase 1.5)
- Modify plan (collect feedback, re-run Phase 1)
- Postpone (exit, continue later)

If user does not select "Proceed": Exit execution.

## Phase 1.5: Task Decomposition

Agent: manager-spec subagent (continuation)

Purpose: Decompose the approved execution plan into atomic, reviewable tasks following SDD 2025 standard.

Tasks for manager-spec:

- Decompose plan into atomic implementation tasks
- Each task must be completable in a single DDD/TDD cycle
- Assign priority and dependencies for each task
- Generate task tracking entries for progress visibility
- Verify task coverage matches all SPEC requirements

Task structure for each decomposed task:

- Task ID: Sequential within SPEC (TASK-001, TASK-002, etc.)
- Description: Clear action statement
- Requirement Mapping: Which SPEC requirement it fulfills
- Dependencies: List of prerequisite tasks
- Acceptance Criteria: How to verify completion

Constraints: Decompose into atomic tasks where each task completes in a single DDD/TDD cycle. No artificial limit on task count. If the SPEC itself is too complex, consider splitting the SPEC.

Output: Task list with coverage_verified flag set to true.

### tasks.md Generation (Persistent Artifact)

After task decomposition, generate `.moai/specs/SPEC-{ID}/tasks.md` for audit trail:

```
## Task Decomposition
SPEC: {SPEC-ID}

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | {desc} | REQ-001 | - | file1.go, file2.go | pending |
| T-002 | {desc} | REQ-002 | T-001 | file3.go | pending |
```

This file is git-tracked. Update task status as implementation progresses.
The planned_files column is used by the Drift Guard (Phase 2A/2B) to detect scope drift.

## Phase 1.6: Acceptance Criteria Initialization (Failing Checklist)

Purpose: Convert all SPEC acceptance criteria into explicit pending TaskList entries. This creates a visible "failing checklist" — all items start as pending and are marked completed (passing) as implementation progresses, following the Harness Engineering pattern.

Action:
- Read spec.md acceptance criteria for SPEC-{ID}
- For each acceptance criterion, execute TaskCreate:
  - subject: `[AC-N] <acceptance criterion statement>`
  - description: Requirement reference, expected behavior, verification method
  - status: pending (starts as "failing")
- Verify all SPEC requirements are covered by at least one task

Output: TaskList populated with all acceptance criteria as pending items.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.6 complete: {N} acceptance criteria registered as pending tasks
```

## Phase 1.7: File Structure Scaffolding

Purpose: Create empty file stubs for all planned new files before implementation begins. This prevents entropy by establishing structure before coding, following the Harness Engineering "Blueprint" pattern.

Condition: Execute only for planned new files (files that do not yet exist in the codebase). Skip if all planned files already exist (modification-only SPEC).

Action:
- Identify all planned new files from Phase 1.5 task decomposition
- For each planned new file that does not yet exist:
  - Create empty stub with minimal required structure matching the project's language conventions (e.g., package declaration for Go, module header for Python, empty class for TypeScript)
  - Do NOT add any implementation logic — stubs only
- After stub creation: Capture LSP baseline (this is the clean baseline before any implementation)

Output: List of stub files created, LSP baseline diagnostics captured.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.7 complete: {N} stub files created, LSP baseline captured
```

## Phase 1.8: Pre-Implementation MX Context Scan

Purpose: Scan files that will be modified during implementation to build an MX context map for implementation agents.

**Scan Target:** All existing files listed in the task decomposition (from Phase 1.5).

**MX Context Extraction:**
- @MX:ANCHOR: Identify invariant contracts. Pass to implementation agents as "do not break" constraints with fan_in counts.
- @MX:WARN: Identify danger zones. Alert agents to approach these areas with extra caution.
- @MX:NOTE: Collect business logic context. Include in agent prompts for informed implementation.
- @MX:TODO: Match against SPEC requirements. If a TODO aligns with a task, the implementation resolves it.
- @MX:LEGACY: Identify legacy code without SPEC. Flag for careful handling during modifications.

**Output:** MX context map included in Phase 2 agent prompts. The map is structured per-file:
- file_path: list of tags with type, line, description, and constraints

**Skip Condition:** If target files do not exist (greenfield implementation), skip this phase.

See .claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions.

## Development Mode Routing

Before Phase 2, determine the development methodology by reading `.moai/config/sections/quality.yaml`:

**If development_mode is "ddd":**
- Route all tasks to manager-develop subagent
- Use ANALYZE-PRESERVE-IMPROVE cycle (see @spec-workflow.md for details)

**If development_mode is "tdd":**
- Route all tasks to manager-develop subagent
- Use RED-GREEN-REFACTOR cycle (see @spec-workflow.md for details)

## Phase 2.0: Sprint Contract Negotiation

**Condition**: Execute only when harness level = thorough.
**Skip**: When harness level = minimal or standard.

Steps:
1. Load implementation plan from Phase 1.5 task decomposition
2. Invoke sync-auditor to review the plan:
   - Identify missing edge cases in proposed test coverage
   - Flag security concerns in the implementation approach
   - Verify acceptance criteria are specific and testable
3. sync-auditor produces contract proposal with:
   - Done criteria (specific test cases that must pass)
   - Edge cases identified for coverage
   - Hard thresholds (coverage %, performance targets, security requirements)
4. Record agreed contract in `.moai/specs/SPEC-{ID}/contract.md`
5. Maximum 2 negotiation rounds. If no agreement after 2 rounds, proceed with evaluator's recommendations as the contract.

Mode-specific deployment:
- Sub-agent mode: Agent(subagent_type="sync-auditor")
- Team mode: SendMessage to reviewer teammate
- CG mode: Leader performs contract negotiation inline

**Output**: `.moai/specs/SPEC-{ID}/contract.md`

## Delta Marker Detection (Brownfield Pre-Check)

Before routing to Phase 2A or 2B, scan the loaded SPEC for `[DELTA]` section markers:

1. Check spec.md (or spec-compact.md) for any line matching `[EXISTING]`, `[MODIFY]`, `[NEW]`, or `[REMOVE]`
2. If NO delta markers found: skip this section, proceed to Phase 2A/2B normally (greenfield path)
3. If delta markers found: activate delta-aware routing as follows

**Delta-aware routing rules (applied within DDD or TDD mode):**

| Marker | Treatment | Action |
|--------|-----------|--------|
| `[EXISTING]` | Context only — do not modify | Generate characterization tests to document current behavior; no code changes |
| `[MODIFY]` | Modify with safety net | Generate characterization tests FIRST, verify they pass, THEN apply modifications |
| `[NEW]` | Full implementation | Apply complete DDD ANALYZE-PRESERVE-IMPROVE or TDD RED-GREEN-REFACTOR cycle |
| `[REMOVE]` | Safe deletion | Check all callers and dependents; confirm no active references; then remove |

**Delta processing order** (prevents regression):
1. Process all `[EXISTING]` items — characterization tests only
2. Process all `[MODIFY]` items — characterization tests → modification → verify tests still pass
3. Process all `[NEW]` items — full implementation cycle
4. Process all `[REMOVE]` items — dependency analysis → safe deletion

If no delta markers are present in the SPEC, delta processing is silently skipped and the standard implementation flow proceeds unchanged (backward compatible with greenfield SPECs).
