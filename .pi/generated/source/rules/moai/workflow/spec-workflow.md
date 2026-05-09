---
paths: "**/.moai/specs/**,**/.moai/config/sections/quality.yaml"
---

# SPEC Workflow

MoAI's three-phase development workflow with token budget management.

## Phase Overview

| Phase | Command | Agent | Token Budget | Purpose |
|-------|---------|-------|--------------|---------|
| Plan | /moai plan | manager-spec | 30K | Create SPEC document |
| Run | /moai run | manager-ddd/tdd (per quality.yaml) | 180K | DDD/TDD implementation |
| Sync | /moai sync | manager-docs | 40K | Documentation sync |

<!-- @MX:ANCHOR fan_in=10 - Subcommand classification single source of truth; cross-referenced by 10 workflow skills (5 multi-agent + 5 utility). Changes here affect all workflow contracts. -->

## Subcommand Classification (Pipeline vs Multi-Agent)

Source: SPEC-V3R2-WF-004. Each MoAI subcommand is classified along the
*control-flow style* axis. The classification governs which agents are spawned,
how the `--mode` flag is interpreted, and which CI guards apply.

| Subcommand   | Class          | 3-phase contract (localize → repair → validate)                | `--mode` honored? | Default mode | Valid `--mode` values | Sentinel on invalid mode | Reference                                                    |
|--------------|----------------|-----------------------------------------------------------------|-------------------|--------------|-----------------------|--------------------------|--------------------------------------------------------------|
| `/moai fix`      | Pipeline (Agentless) | Parallel Scan + Classify + MX context → Auto-Fix → Verify        | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.pi/generated/source/skills/moai/workflows/fix.md`                       |
| `/moai coverage` | Pipeline (Agentless) | Measure + Gap Analysis → Test Generation → Verify                 | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.pi/generated/source/skills/moai/workflows/coverage.md`                  |
| `/moai mx`       | Pipeline (Agentless) | Pass 1 + Pass 2 → Pass 3 → Post-edit scan                         | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.pi/generated/source/skills/moai/workflows/mx.md`                        |
| `/moai codemaps` | Pipeline (Agentless) | Explore → Analyze + Generate → Verify                             | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.pi/generated/source/skills/moai/workflows/codemaps.md`                  |
| `/moai clean`    | Pipeline (Agentless) | Static Analysis + Usage Graph → Safe Removal → Test Verification  | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.pi/generated/source/skills/moai/workflows/clean.md`                     |
| `/moai plan`     | Multi-Agent    | n/a — open-ended (mode-NA per REQ-WF003-005)                      | Yes (rejects `pipeline`) | `autopilot` | (none — `--mode` ignored) | `MODE_PIPELINE_ONLY_UTILITY` (only on `pipeline`) | `.pi/generated/source/skills/moai/workflows/plan.md`               |
| `/moai run`      | Multi-Agent    | n/a — open-ended (`autopilot` / `loop` / `team` per WF-003)        | Yes (rejects `pipeline`) | `autopilot` (harness `minimal`/`standard`); `team` (harness `thorough` + prereqs) | `autopilot`, `loop`, `team` | `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY` | `.pi/generated/source/skills/moai/workflows/run.md`                |
| `/moai sync`     | Multi-Agent    | n/a — open-ended (mode-NA per REQ-WF003-005)                      | Yes (rejects `pipeline`) | `autopilot` | (none — `--mode` ignored) | `MODE_PIPELINE_ONLY_UTILITY` (only on `pipeline`) | `.pi/generated/source/skills/moai/workflows/sync.md`               |
| `/moai design`   | Multi-Agent    | n/a — open-ended (`autopilot` / `import` / `team` per WF-003)      | Yes (rejects `pipeline`) | `autopilot` (harness `minimal`/`standard`); `team` (harness `thorough` + prereqs) | `autopilot`, `import`, `team` | `MODE_UNKNOWN`, `MODE_PIPELINE_ONLY_UTILITY` | `.pi/generated/source/skills/moai/workflows/design.md`             |
| `/moai loop`     | Multi-Agent (alias for `/moai run --mode loop`) | n/a — delegates to `/moai run` mode dispatch | Yes (alias semantics) | (inherits from `run --mode loop`) | (alias only — `--mode` resolves via `run`) | (delegates to `run` sentinels) | `.pi/generated/source/skills/moai/workflows/loop.md`               |

### Mode Dispatch Cross-Reference

Source: SPEC-V3R2-WF-003. The `--mode` axis values are valid only on multi-agent subcommands that explicitly support mode dispatch: `/moai run` and `/moai design`. Other multi-agent subcommands (`/moai plan`, `/moai sync`, `/moai project`, `/moai db`) ignore `--mode` per REQ-WF003-005, except they REJECT `--mode pipeline` with `MODE_PIPELINE_ONLY_UTILITY` per REQ-WF003-016 (shared with WF-004 REQ-WF004-014).

`/moai loop` is an alias for `/moai run --mode loop` per REQ-WF003-004. Both routes invoke the Ralph Engine identically; the alias preserves the historical entry point.

Mode precedence (REQ-WF003-018, hard-coded):

1. CLI flag `--mode <value>` — highest priority.
2. Config field `workflow.default_mode` in `.moai/config/sections/workflow.yaml`.
3. Harness auto-selection — lowest priority (per `harness.yaml` level).

Auto-selection rules (REQ-WF003-002, REQ-WF003-003):

- Harness `minimal` or `standard` → default mode = `autopilot`
- Harness `thorough` AND `workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` → default mode = `team`
- Otherwise (thorough but team prereqs missing) → fallback to `autopilot` with `[mode-auto-downgrade]` info log per REQ-WF003-012.

Sentinel error keys:

- `MODE_UNKNOWN` (REQ-WF003-010, owned by WF-003): invalid `--mode` value supplied.
- `MODE_TEAM_UNAVAILABLE` (REQ-WF003-011, owned by WF-003): explicit `--mode team` request without prerequisites.
- `MODE_PIPELINE_ONLY_UTILITY` (REQ-WF003-016 ↔ REQ-WF004-014, shared): `--mode pipeline` on a multi-agent subcommand.
- `MODE_FLAG_IGNORED_FOR_UTILITY` (REQ-WF004-011, owned by WF-004): `--mode <any>` on a utility subcommand (info log only).

See `.pi/generated/source/skills/moai/workflows/run.md` § Mode Dispatch and `.pi/generated/source/skills/moai/workflows/design.md` § Mode Dispatch for the per-skill dispatch rules.

### Pipeline Class — Contract

Pipeline-classified subcommands MUST satisfy:
- Three deterministic phases (localize → repair → validate); no LLM dispatcher selects the next phase.
- `Agent()` invocations are permitted only as **executor delegation within a phase** (e.g., `expert-testing` runs the coverage tool); never to decide phase order.
- When localize finds zero targets, exit with status `no-op` and exit code 0.
- When repair encounters an unresolvable error, fail-fast (no multi-agent fallback).
- The CI guard `internal/template/agentless_audit_test.go` enforces the no-LLM-dispatch rule via static text scan.

### Out of scope of this matrix

`/moai feedback`, `/moai review`, `/moai e2e` are *not* yet classified.
See `spec.md` §1.2 (Non-Goals) — they are deferred to a future SPEC.

### Cross-references

- `--mode` flag matrix: SPEC-V3R2-WF-003 (sibling SPEC, defines `autopilot|loop|team|pipeline`).
- Pipeline regression guard: `internal/template/agentless_audit_test.go` (REQ-WF004-013).
- Pattern source: `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6 (Agentless).
- Research source: `.moai/design/v3-redesign/research/r1-ai-harness-papers.md` §25 (Xia et al. 2024).

## Plan Phase

Create comprehensive specification using EARS format.

Sub-phases:
1. Research: Deep codebase analysis producing research.md artifact
2. Planning: SPEC document creation with EARS format requirements
3. Annotation: Iterative plan review cycle (1-6 iterations) before implementation approval

Token Strategy:
- Allocation: 30,000 tokens
- Load requirements only
- Execute /clear after completion
- Saves 45-50K tokens for implementation

Output:
- Research document at `.moai/specs/SPEC-XXX/research.md` (deep codebase analysis)
- SPEC document at `.moai/specs/SPEC-XXX/spec.md`
- EARS format requirements
- Acceptance criteria
- Technical approach

## Run Phase

Implement specification using configured development methodology.

Token Strategy:
- Allocation: 180,000 tokens
- Selective file loading
- Enables 70% larger implementations

Development Methodology (configured in quality.yaml development_mode):

### DDD Mode — ANALYZE-PRESERVE-IMPROVE

Best for existing projects with < 10% test coverage. Uses manager-ddd agent.

**ANALYZE**: Read existing code, map domain boundaries, identify side effects and implicit contracts.
**PRESERVE**: Write characterization tests capturing current behavior. Create behavior snapshots for regression detection.
**IMPROVE**: Make small incremental changes. Run characterization tests after each change. Refactor with test validation.

### TDD Mode — RED-GREEN-REFACTOR (default)

Best for all development work, new projects, and brownfield with 10%+ coverage. Uses manager-tdd agent.

**RED**: Write a failing test describing desired behavior. Verify it fails. One test at a time.
**GREEN**: Write simplest implementation that passes. No premature optimization.
**REFACTOR**: Clean up while keeping tests green. Extract patterns, remove duplication.

Brownfield enhancement: Pre-RED step reads existing code to understand current behavior before writing the failing test.

### Methodology Auto-Detection

| Project State | Test Coverage | Recommendation |
|--------------|---------------|----------------|
| Greenfield (new) | N/A | TDD |
| Brownfield | >= 10% | TDD |
| Brownfield | < 10% | DDD |

Manual override: `quality.development_mode` in quality.yaml, `MOAI_DEVELOPMENT_MODE` env var, or `moai init --mode <ddd|tdd>`.

### Pre-submission Self-Review

Before marking implementation complete: review full diff against SPEC acceptance criteria. Ask "Is there a simpler approach?" and "Would removing any changes still satisfy the SPEC?" Skip for single-file changes under 50 lines, bug fixes with reproduction test, or user-approved annotation cycle changes.

### Drift Guard

After each methodology cycle, compare planned files against actual modifications. Warns at <= 30% drift. Triggers re-planning (Phase 2.7) above 30%.

### Team Mode Methodology

Each teammate applies the methodology within their file ownership scope. team-validator validates compliance. team-tester exclusively owns test files.

### MX Tag Integration

| Phase | TDD Action | DDD Action |
|-------|-----------|-----------|
| Test/Analyze | RED: add `@MX:TODO` | ANALYZE: 3-Pass scan, identify targets |
| Implement/Preserve | GREEN: remove `@MX:TODO` | PRESERVE: validate tags, add `@MX:LEGACY` |
| Refactor/Improve | REFACTOR: add `@MX:NOTE` | IMPROVE: update tags, add `@MX:NOTE` |

Success Criteria:
- All SPEC requirements implemented
- Methodology-specific tests passing
- 85%+ code coverage
- TRUST 5 quality gates passed
- MX tags added for new code (NOTE, ANCHOR, WARN as appropriate)

### Re-planning Gate

Detect when implementation is stuck or diverging from SPEC and trigger re-assessment.

Triggers:
- 3+ iterations with no new SPEC acceptance criteria met
- Test coverage dropping instead of increasing across iterations
- New errors introduced exceed errors fixed in a cycle
- Agent explicitly reports inability to meet a SPEC requirement

Communication path:
- Implementation agent (manager-ddd/tdd) detects trigger condition
- Agent returns structured stagnation report to MoAI (agents cannot call AskUserQuestion)
- MoAI presents gap analysis to user via AskUserQuestion with options:
  - Continue with current approach (minor adjustments needed)
  - Revise SPEC (requirements need refinement)
  - Try alternative approach (re-delegate to manager-strategy)
  - Pause for manual intervention (user takes over)

Detection method:
- Append acceptance criteria completion count and error count delta to `.moai/specs/SPEC-{ID}/progress.md` at the end of each iteration
- Compare against previous entry to detect stagnation
- Flag stagnation when acceptance criteria completion rate is zero for 3+ consecutive entries

Integration: Referenced by run.md Phase 2.7 and loop.md iteration checks

## Sync Phase

Generate documentation and prepare for deployment.

Token Strategy:
- Allocation: 40,000 tokens
- Result caching
- 60% fewer redundant file reads

Output:
- API documentation
- Updated README
- CHANGELOG entry
- Pull request

## Completion Markers

AI uses markers to signal task completion:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion

## Context Management

/clear Strategy:
- After /moai plan completion (mandatory)
- When context exceeds 150K tokens
- Before major phase transitions

Progressive Disclosure:
- Level 1: Metadata only (~100 tokens)
- Level 2: Skill body when triggered (~5000 tokens)
- Level 3: Bundled files on-demand

## Phase Transitions

Plan to Run:
- Trigger: SPEC document approved (annotation cycle completed, user confirmed "Proceed")
- Pre-condition: plan.md records `plan_complete_at` + `plan_status: audit-ready` in progress.md
- Action: Execute /clear, then /moai run SPEC-XXX
- Gate: `/moai run` Phase 0.5 (Plan Audit Gate) executes automatically before any implementation.
  See "Phase 0.5: Plan Audit Gate" section below for details.

## Phase 0.5: Plan Audit Gate

The Plan Audit Gate is a mandatory protocol executed at the start of every `/moai run` invocation,
before any implementation phase begins. The gate invokes the plan-auditor subagent to independently
review all SPEC plan artifacts. It prevents unreviewed or incomplete SPEC artifacts from entering
the implementation phase. Source: SPEC-WF-AUDIT-GATE-001.

### Gate Entry Condition

- Triggered on every `/moai run <SPEC-ID>` invocation (REQ-WAG-001)
- Applies in both solo mode (workflows/run.md) and team mode (team/run.md)
- Cannot be skipped by harness level — gate is never disabled, not even on `minimal`

### Verdicts

| Verdict | Meaning | Action |
|---------|---------|--------|
| `PASS` | All must-pass criteria met | Persist to progress.md, proceed to Phase 1 |
| `FAIL` | One or more must-pass criteria failed | Block Phase 1, surface report, AskUserQuestion |
| `BYPASSED` | User passed `--skip-audit` or set `MOAI_SKIP_PLAN_AUDIT=1` | Record bypass in report, proceed |
| `INCONCLUSIVE` | Auditor timed out, errored, or returned malformed output | Block, AskUserQuestion (retry/proceed/abort) |

### Report Persistence

Every gate call persists a record at `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md`.
Multiple calls on the same day append to the same file. Reports are local artifacts (gitignored).

### Grace Window

7-day grace window after SPEC-WF-AUDIT-GATE-001 merge: FAIL verdicts emit warnings only (FAIL_WARNED),
not blocking. After grace window expires, FAIL verdicts block Phase 1 unconditionally.
Grace window start: `.moai/state/audit-gate-merge-at.txt` (ISO-8601 timestamp).

Run to Sync:
- Trigger: Implementation complete, tests passing
- Action: Execute /moai sync SPEC-XXX

## Agent Teams Variant

When team mode is enabled (workflow.team.enabled and AGENT_TEAMS env), phases can execute with Agent Teams instead of sub-agents.

### Team Mode Phase Overview

| Phase | Sub-agent Mode | Team Mode | Condition |
|-------|---------------|-----------|-----------|
| Plan | manager-spec (single) | Dynamic teammates: researcher + analyst + architect (parallel, general-purpose) | Complexity >= threshold |
| Run | manager-ddd/tdd (sequential) | Dynamic teammates: backend-dev + frontend-dev + tester (parallel, general-purpose) | Domains >= 3 or files >= 10 |
| Sync | manager-docs (single) | manager-docs (always sub-agent) | N/A |

All teammates are spawned dynamically via `Agent(subagent_type: "general-purpose")` with runtime overrides from `workflow.yaml` role profiles. No static team agent definitions are used. See `.pi/generated/source/skills/moai/team/run.md` for complete orchestration.

### Team Mode Plan Phase
- TeamCreate for parallel research team
- Spawn general-purpose teammates with mode: "plan" (read-only)
- researcher teammate produces research.md with deep codebase analysis
- analyst teammate validates requirements against research findings
- architect teammate designs solution using reference implementations found in research
- MoAI runs annotation cycle with user for plan refinement (1-6 iterations)
- MoAI synthesizes into SPEC document
- Shutdown team, /clear before Run phase

### Team Mode Run Phase
- TeamCreate for implementation team
- Task decomposition with file ownership boundaries
- [HARD] Implementation teammates (role_profiles: implementer, tester) MUST use `isolation: "worktree"` for parallel file safety
- [HARD] Read-only teammates (role_profiles: reviewer) MUST NOT use isolation — mode: "plan" is sufficient
- Teammates self-claim tasks from shared list
- Quality validation after all implementation completes
- Worktree cleanup via `git worktree prune` after team shutdown
- Shutdown team

### Token Cost Awareness

Agent teams use significantly more tokens than a single session. Each teammate has its own independent context window, so token usage scales linearly with the number of active teammates.

Estimated token multipliers by team pattern:
- plan_research (3 teammates): ~3x plan phase tokens
- implementation (3 teammates): ~3x run phase tokens
- design_implementation (4 teammates): ~4x run phase tokens
- investigation (3 teammates): ~2x (haiku model reduces cost)
- review (3 teammates): ~2x (read-only, shorter sessions)

When to prefer team mode over sub-agent mode:
- Research and review tasks where parallel exploration adds real value
- Cross-layer features (frontend + backend + tests)
- Complex debugging with multiple potential root causes
- Tasks where teammates need to communicate and coordinate

When to prefer sub-agent mode:
- Sequential tasks with heavy dependencies
- Same-file edits or tightly coupled changes
- Routine tasks with clear single-domain scope
- Token budget is a concern

### Team Workflow References

Detailed team orchestration steps are defined in dedicated workflow files:

- Plan phase: .pi/generated/source/skills/moai/team/plan.md
- Run phase: .pi/generated/source/skills/moai/team/run.md
- Fix phase: .pi/generated/source/skills/moai/team/debug.md
- Review: .pi/generated/source/skills/moai/team/review.md

### Known Limitations

For complete limitations list, see .pi/generated/source/CLAUDE.md Section 15.

### Prerequisites

Both conditions must be met:
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in settings.json env
- `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml`

See .pi/generated/source/CLAUDE.md Section 15 for details.

### Mode Selection
- --team flag: Force team mode
- --solo flag: Force sub-agent mode
- No flag (default): Complexity-based selection
- See workflow.yaml team.auto_selection for thresholds

### Fallback
If team mode fails or prerequisites are not met:
- Graceful fallback to sub-agent mode
- Continue from last completed task
- No data loss or state corruption
- Trigger conditions: AGENT_TEAMS env not set, workflow.team.enabled false, TeamCreate failure, teammate spawn failure
