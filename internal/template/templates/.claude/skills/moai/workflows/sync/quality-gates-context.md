---
description: "Sync Phase 0 — Purpose/Scope/Input/Mode/Flags/Context Loading and Phase 0 Pre-Sync Quality Gate through Phase 0.1 Deployment Readiness Check. Contains HUMAN GATE 1 (Pre-Sync Quality)."
user-invocable: false
metadata:
  parent: moai-workflow-sync
  phase: "Phase 0: Pre-Sync Context and Deployment Readiness"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

# Sync Workflow Orchestration

## Purpose

Synchronize documentation with code changes, verify project quality, and finalize pull requests. This is the third step of the Plan-Run-Sync workflow.

## Scope

- Implements Step 4 of MoAI's 4-step workflow (Report and Commit)
- Receives implementation artifacts from /moai run
- Produces synchronized documentation, commits, and PR readiness

## Input

- $ARGUMENTS: Mode and optional path
  - Mode: auto (default), force, status, project
  - Path: Optional synchronization target path (e.g., src/auth/)
  - Flag: --merge

## Mode Flag Compatibility

Per the subcommand-classification contract:

- This subcommand is multi-agent (open-ended) but does NOT participate in the
  `--mode {autopilot|loop|team}` axis.
- Any `--mode` value supplied to `/moai sync` is silently ignored. The sync workflow
  proceeds with its default behavior.
- The `pipeline` value is the only special case: passing `--mode pipeline` triggers
  `MODE_PIPELINE_ONLY_UTILITY` (the same error key the utility subcommands share).

See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the
full subcommand × mode matrix.

## Supported Modes

- auto (default): Smart selective sync of changed files only. PR Ready conversion. Daily development workflow.
- force: Complete regeneration of all documentation. Error recovery and major refactoring use case.
- status: Read-only health check. Quick project health report with no changes.
- project: Project-wide documentation updates. Milestone completion and periodic sync use case.

### Project Mode Details (ENHANCED)

The `project` mode performs comprehensive project-wide synchronization:

**When to use:**
- After completing a milestone or major feature
- Before releasing a new version
- Periodic maintenance (weekly/monthly)
- After significant refactoring
- When `.moai/project/` documents are outdated

**What project mode does:**

1. **Full Project Scan** (vs. auto mode's selective scan):
   - Scans ALL source files (not just changed files)
   - Checks ALL SPEC documents for updates needed
   - Verifies ALL project documentation consistency
   - Validates ALL language files for MX tag coverage

2. **SPEC Document Update Detection**:
   - Compares implementation against SPEC requirements
   - Detects implemented features not documented in SPEC
   - Detects SPEC requirements not yet implemented
   - Flags SPEC documents requiring updates

3. **Project Document Updates**:
   - Updates `.moai/project/tech.md` when new dependencies/technologies added
   - Updates `.moai/project/structure.md` when architecture changes
   - Updates `.moai/project/product.md` when new features added
   - Updates `.moai/project/codemaps/` when architecture changes detected (delegates to codemaps workflow)
   - Updates README.md to reflect current project state

4. **Comprehensive Quality Verification**:
   - Runs full test suite (all languages)
   - Lint check for ALL source files
   - Type check for ALL source files
   - MX tag validation for ALL source files

**Output for project mode:**
- Complete project health report
- All SPEC documents requiring updates
- All project documents requiring updates
- Recommendations for improvements
- Full language breakdown of code quality metrics

## Supported Flags

- --pr: Push branch and create/update PR on GitHub after sync. When used, automatically returns to base branch (main/develop) after PR creation (Step 3.3.5).
- --merge: After sync, auto-merge PR and clean up branch. Worktree/branch environment is auto-detected from git context.
- --skip-mx: Skip MX tag validation and annotation during sync.

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, language settings)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/ directory listing (SPEC documents for sync)
- .moai/project/ directory listing (project documents for conditional update)
- .moai/project/codemaps/ directory listing (architecture maps for conditional update)
- README.md (current project documentation)

Pre-execution commands: git status, git diff, git branch, git log, find .moai/specs.

---

## Phase Sequence

> **Phase-number namespace note**: The Phase N / Phase N.M labels below (Phase 0, Phase 0.08, Phase 0.1, Phase 0.5, Phase 0.55, Phase 0.6, Phase 0.7, Phase 1, Phase 2, Phase 3, Phase 4) are local to the `/moai sync` workflow only. They are a distinct numbering namespace from `harness.yaml` `levels.<level>.skip_phases`, which governs which `/moai run` phases (task-decomposition.md / phase-execution.md numbering) are skipped at a given harness level. A sync-phase number and a run-phase `skip_phases` entry sharing the same numeral (e.g. both using `0.5`) do NOT refer to the same phase.

### Phase 0: Pre-Sync Quality Gate

<!-- moai:evolvable-start id="gate-sync-1" -->
### HUMAN GATE: Pre-Sync Quality

**Previous phase output:** Completed SPEC implementation
**Approval question:** Is the project in a state where documentation can be synced?
**Cannot proceed until:**
- [ ] Working tree is clean or only expected changes present
- [ ] All tests pass
- [ ] MX tags validated
- [ ] No HARD rule violations
<!-- moai:evolvable-end -->

Purpose: Run the gate workflow (workflows/gate.md) as a fast pre-check before the full deployment readiness verification. Catches lint/format/type errors early and auto-fixes them.

#### Step 0.0.1: Gate Execution

- Execute gate workflow equivalent: lint + format + type-check + test in parallel
- Auto-fix any fixable issues (lint auto-fix, format auto-fix)
- If unfixable errors remain: Present summary and offer options via AskUserQuestion
  - Fix errors (Recommended): Delegate to manager-develop subagent for targeted fixes
  - Skip gate: Proceed to Phase 0.1 (errors will be caught later but at higher cost)
  - Abort: Exit sync workflow

Output: gate_report with pass/fail per check category.

### Phase 0.08: DB Schema Doc Check (Conditional)

Purpose: Refresh `.moai/project/db/` derived documents (schema.md, erd.mmd, migrations.md) when the sync scope includes migration file changes. Replaces the per-event PostToolUse hook with a batch refresh at milestone boundary — eliminates the ~30-60ms/edit overhead the hook used to incur.

#### Step 0.08.1: Activation Gate

Evaluate all conditions in order; skip the phase if any fails:

1. `.moai/config/sections/db.yaml` exists (project opted into DB doc management)
2. `db.enabled: true` in that file
3. `db.auto_sync: true` in that file

If any condition is not met, skip Phase 0.08 silently and proceed to Phase 0.1.

#### Step 0.08.2: Migration File Diff Detection

Compute the list of migration files changed since the base branch:

- Use `git diff --name-only <base-branch>..HEAD` to collect changed files
- Filter by the glob patterns in `db.migration_patterns` (typically Prisma schema, Alembic versions, Rails migrations, raw SQL, Supabase, custom)
- Further exclude paths matching `db.excluded_patterns` (defaults: `.moai/project/db/**`, `.moai/cache/**`, `.moai/logs/**`) to prevent recursion

If the filtered list is empty, skip to Phase 0.1 with log line: "Phase 0.08: no migration files changed, skipping DB doc refresh".

#### Step 0.08.3: Refresh Invocation

Invoke the internal db-schema-sync hook subcommand directly via Bash:

```
moai hook db-schema-sync
```

- Input (stdin JSON): filtered migration file list, project language, `db.yaml` config — read from current working directory by the Go handler
- Implementation: the DB-schema-sync hook handler, registered as the `moai hook db-schema-sync` subcommand
- Output: updated `.moai/project/db/schema.md`, `erd.mmd`, `migrations.md`; refresh report
- Changes are staged for the sync commit — no separate commit is created

On refresh failure (parser error, template conflict): log the error, include in sync report under "DB doc refresh warnings", and continue to Phase 0.1. Non-blocking by contract.

The `/moai db` slash command was retired (Bundle A, 2026-05-16) — sync workflow is now the sole entry point for db doc refresh. Internal `moai hook db-schema-sync` Go subcommand remains for hook event handlers and direct invocation by sync workflow.

#### Step 0.08.4: Advisory Path

When migration files changed but `db.auto_sync: false`:

- Emit one-line advisory to the sync report: "N migration files changed but db.auto_sync is disabled — set `db.auto_sync: true` in `.moai/config/sections/db.yaml` to enable automatic db doc refresh on sync"
- Do not invoke refresh automatically — respect user opt-out

Output: phase_result with one of `skipped | refreshed | advised | failed` and the migration file count.

### Phase 0.1: Deployment Readiness Check

Purpose: Verify the implementation is deployment-ready before quality verification and documentation sync. Catches deployment-blocking issues early.

#### Step 0.1.1: Test Passage Verification

- Run full test suite for detected project language
- Verify all tests pass (zero failures required)
- If tests fail: Present failure summary and offer options via AskUserQuestion
  - Fix and retry (Recommended): Delegate to manager-develop subagent
  - Continue anyway: Proceed with warning
  - Abort: Exit sync workflow

#### Step 0.2: Migration Check

- Scan for database schema changes (new models, altered tables, migration files)
- Scan for configuration format changes (new config keys, changed defaults)
- Scan for data format changes (API request/response shape changes)
- If migrations detected: Flag as deployment prerequisite and include in sync report

#### Step 0.3: Environment and Configuration Changes

- Scan for new environment variables referenced in code but not in .env.example or documentation
- Scan for new configuration files or sections added
- Scan for changed default values in existing configuration
- If changes detected: Generate environment change summary for inclusion in PR description

#### Step 0.4: Backward Compatibility Assessment

- Identify public API changes (removed endpoints, changed signatures, removed fields)
- Identify breaking changes in exported functions or types
- Identify dependency version changes that may affect consumers
- Severity classification:
  - Breaking: Must be documented and versioned (semver major bump)
  - Deprecation: Must include migration guide
  - Compatible: No action required
- If breaking changes detected: Require explicit user acknowledgment via AskUserQuestion

Output: deployment_readiness_report with test_status, migrations_needed, env_changes, breaking_changes, and overall readiness status (READY, NEEDS_ATTENTION, or BLOCKED).

If overall status is BLOCKED: Present blocking issues to user and exit unless user overrides.
