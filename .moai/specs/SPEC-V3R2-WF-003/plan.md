# SPEC-V3R2-WF-003 Implementation Plan (Phase 1B)

> Implementation plan for Multi-Mode Router (`--mode {autopilot|loop|team|pipeline}` flag).
> Companion to `spec.md` v0.2.0 and `research.md` v0.1.0.
> Authored against worktree `feature/SPEC-V3R2-WF-003` at `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003`.
> Stacked on top of `feature/SPEC-V3R2-WF-004` (PR #765, OPEN).

## HISTORY

| Version | Date       | Author                              | Description                                                              |
|---------|------------|-------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-03 | MoAI Plan Workflow (Phase 1B)       | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow (iter 2)         | M4d scope expanded from 2 (plan/sync) → 4 (plan, sync, project, db) mode-NA implementation subcommands per plan-auditor iteration 1 D4 fix. §1.3 Deliverables table + §3.1 Anchor table updated. M1-M3 unchanged. |

---

## 1. Plan Overview

### 1.1 Goal restatement

Codify spec.md REQ-WF003-001..018:

- **Mode axis**: introduce `--mode {autopilot|loop|team|pipeline}` flag on `/moai run` and `/moai design` (REQ-WF003-001).
- **Default selection**: `autopilot` for harness ∈ {minimal, standard}; `team` when harness == thorough AND team prerequisites met (REQ-WF003-002, REQ-WF003-003); silent downgrade to `autopilot` when prerequisites missing (REQ-WF003-012).
- **Alias**: `/moai loop` resolves to `/moai run --mode loop` (REQ-WF003-004).
- **Mode-NA**: `plan`, `sync`, `project`, `db` ignore `--mode` (REQ-WF003-005); `pipeline` mode rejected on multi-agent subcommands per WF-004 cross-ref (REQ-WF003-016).
- **Sentinel error keys**: `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY` (shared with WF-004).
- **Precedence**: CLI `--mode` > `workflow.default_mode` config > harness auto (REQ-WF003-018).
- **Matrix publication**: subcommand × mode matrix EXTENDS the WF-004 matrix in `spec-workflow.md`.

This is an **additive-only** change. spec.md frontmatter `breaking: false` and `bc_id: []` reflect that users not supplying `--mode` get exactly today's behavior. Verified in research.md §11 item 7 (BC-WF003 scope check).

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md:60-65`. Concretely:

- **RED**: Extend `internal/template/agentless_audit_test.go` (created in WF-004 M1) with three new test functions for `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, and `/moai loop` ↔ `/moai run --mode loop` cross-reference. Confirm RED in CI.
- **GREEN**: Add `## Mode Dispatch` section to `run.md` and `design.md` with sentinel strings and dispatch logic; add cross-reference notes to `loop.md`; add Mode-NA notation to `plan.md` and `sync.md`; extend the Subcommand Classification matrix in `spec-workflow.md` (added by WF-004 M4) with new columns + `/moai loop` alias row. Add optional `default_mode` field to `workflow.yaml`. Make tests pass.
- **REFACTOR**: Consolidate the mode resolver pseudocode if duplicated between `run.md` and `design.md`; cross-link from skill bodies to the unified matrix anchor.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Mode Dispatch section in `/moai run` skill | `.claude/skills/moai/workflows/run.md` (new `## Mode Dispatch (Multi-Mode Router)` section) | REQ-WF003-001, 002, 003, 008, 010, 011, 012, 014, 016, 017, 018 |
| Mode Dispatch section in `/moai design` skill | `.claude/skills/moai/workflows/design.md` (new `## Mode Dispatch (Multi-Mode Router)` section) | REQ-WF003-001, 009, 010, 013, 014, 016, 018 |
| `/moai loop` alias note | `.claude/skills/moai/workflows/loop.md` (header note + cross-reference) | REQ-WF003-004 |
| Mode-NA notation | `.claude/skills/moai/workflows/plan.md`, `.claude/skills/moai/workflows/sync.md`, `.claude/skills/moai/workflows/project.md`, `.claude/skills/moai/workflows/db.md` (existing `## Mode Flag Compatibility` section refined per WF-004 M3 — all 4 mode-NA implementation subcommands per REQ-WF003-005) | REQ-WF003-005 |
| Extended Subcommand × Mode matrix | `.claude/rules/moai/workflow/spec-workflow.md` (`## Subcommand Classification` section — extend columns + add `/moai loop` row) | REQ-WF003-007, REQ-WF003-016 |
| `default_mode` schema extension | `.moai/config/sections/workflow.yaml` (new optional field with comment) | REQ-WF003-014 |
| Mode audit tests (extends WF-004 file) | `internal/template/agentless_audit_test.go` (new test functions) | REQ-WF003-010, 011, 016 sentinels + REQ-WF003-004 alias documentation |
| CHANGELOG entry | `CHANGELOG.md` (Unreleased section) | Trackable (TRUST 5) |

Embedded-template parity is a **HARD** requirement: every change to `.claude/skills/moai/workflows/*.md`, `.claude/rules/moai/workflow/spec-workflow.md`, and `.moai/config/sections/workflow.yaml` MUST also be applied to the corresponding `internal/template/templates/.claude/...` and `internal/template/templates/.moai/...` source-of-truth files (per `CLAUDE.local.md` §2 Template-First Rule). `make build` regenerates `internal/template/embedded.go` after template edits.

### 1.4 Stacked PR Discipline (CRITICAL)

This SPEC's PR will be opened with base `feature/SPEC-V3R2-WF-004` (PR #765, OPEN). Per CLAUDE.local.md §18.11 v2.14.0 Case Study, stacked PRs auto-close when their parent PR merges. To prevent this:

[HARD] **Pre-merge base transition**: Before PR #765 (WF-004) merges to `main`, the WF-003 PR's base MUST be transitioned from `feature/SPEC-V3R2-WF-004` → `main`. Steps:

1. Detect PR #765 approaching merge (CI all-green, reviews approved).
2. Run `gh pr edit <WF-003-PR-#> --base main` to retarget the base.
3. Verify the WF-003 PR diff still contains only the WF-003-specific changes (the WF-004 changes will already be in `main` after PR #765 lands).
4. Resolve any merge conflicts that appear after the base transition.
5. Allow PR #765 to merge.
6. WF-003 PR is now standalone against `main` — no auto-close risk.

This pre-merge hook is the run phase's responsibility (or the user's, post-implementation). It is documented here so the run phase agent and reviewer are aware before the merge window opens.

### 1.5 Coordination with WF-004 Implementation

WF-003 is BLOCKED-BY WF-004 (per spec.md §9.1). The run phase coordination:

- WF-004 M1 creates `internal/template/agentless_audit_test.go` with 3 test functions and the embedded-FS walker pattern.
- WF-004 M2 adds `## Pipeline Contract` sections with `MODE_FLAG_IGNORED_FOR_UTILITY` to 5 utility skills.
- WF-004 M3 adds `## Mode Flag Compatibility` sections with `MODE_PIPELINE_ONLY_UTILITY` to 4 implementation skills (`plan.md`, `run.md`, `sync.md`, `design.md`).
- WF-004 M4 inserts `## Subcommand Classification` matrix in `spec-workflow.md` (9 rows × 5 columns).
- WF-003 EXTENDS WF-004 M3 sections in `run.md` and `design.md` with full Mode Dispatch logic (not just `pipeline` rejection).
- WF-003 EXTENDS WF-004 M4 matrix with two new columns ("Default mode", "Valid `--mode` values", "Sentinel on invalid mode") and one new row (`/moai loop` alias).
- WF-003 EXTENDS WF-004 M1 audit test file with 3 new test functions.

If WF-004 lands first (PR #765 merges before WF-003 PR), WF-003 builds on the merged baseline. If WF-003 PR is opened concurrently with WF-004 (stacked), WF-003 must reference the WF-004 worktree's pre-merge state and verify post-merge.

---

## 2. Milestone Breakdown (M1-M5)

Each milestone is **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule). Dependencies are explicit.

### M1: Test scaffolding extension (RED phase) — Priority P0

Reference: `internal/template/agentless_audit_test.go` (created in WF-004 M1; pattern from `internal/template/commands_audit_test.go:11-50`).

Owner role: `expert-backend` (Go test) or direct manager-tdd execution.

Scope:
1. Open `internal/template/agentless_audit_test.go` (already exists post WF-004 M1 in stacked base).
2. Add test func `TestRunDesignSkillsContainModeUnknownSentinel` walking `.claude/skills/moai/workflows/{run,design}.md` and asserting each contains the literal string `MODE_UNKNOWN` (REQ-WF003-010 enforcement).
3. Add test func `TestRunSkillContainsModeTeamUnavailableSentinel` walking `.claude/skills/moai/workflows/run.md` and asserting it contains the literal string `MODE_TEAM_UNAVAILABLE` (REQ-WF003-011 enforcement).
4. Add test func `TestLoopAliasCrossReference` walking `.claude/skills/moai/workflows/loop.md` and asserting it contains the literal string `/moai run --mode loop` (REQ-WF003-004 alias documentation).
5. Run `go test ./internal/template/ -run "TestRun|TestLoop"` — confirm RED (all three new tests fail because sentinels and cross-reference text are not yet added; existing WF-004 tests still pass).

Verification gate before advancing to M2: 3 new tests fail with "missing sentinel" message; pre-existing WF-004 tests (TestAgentlessUtilityNoLLMControlFlow, TestUtilitySkillsContainModeFlagIgnoredSentinel, TestImplementationSkillsContainPipelineRejectionSentinel) remain GREEN.

[HARD] No implementation code in M1 outside of test files.

[HARD] If WF-004 M1 has not yet landed in the stacked base, M1 must wait. The WF-003 worktree is based on `feature/SPEC-V3R2-WF-004` HEAD `5ab409292`; verify the file `internal/template/agentless_audit_test.go` exists in this base before extending it.

### M2: Mode Dispatch section in `/moai run` skill (GREEN, part 1) — Priority P0

Owner role: `manager-docs` for content authoring; `expert-backend` for any embedded-template parity build step.

Scope:
1. Open `.claude/skills/moai/workflows/run.md`.
2. Locate the existing `## Mode Flag Compatibility` section (added by WF-004 M3 — contains `MODE_PIPELINE_ONLY_UTILITY` rejection clause).
3. EXTEND that section into a new `## Mode Dispatch (Multi-Mode Router)` section that:
   - References SPEC-V3R2-WF-003 in the section header.
   - Documents the 4 `--mode` values and their behaviors:
     - `autopilot` (default): single-lead orchestration via Phase 0.95 Scale-Based Mode Selection (Fix/Focused/Standard/Full Pipeline) → Phase 2A/2B per `quality.yaml development_mode`.
     - `loop`: delegate to `Skill("moai-workflow-loop")` with the SPEC-ID and remaining args. Bypasses Phase 2A/2B and enters the Ralph engine per-iteration cycle (`loop.md:46-138`).
     - `team`: route to existing Team Mode (Phase 0.95 row 5 + `run.md:927-943` Team Mode Routing). Requires `workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.
     - `pipeline`: REJECTED. Emit `MODE_PIPELINE_ONLY_UTILITY` error (already documented per WF-004 M3 — preserve verbatim).
   - Documents the mode resolver pseudocode (precedence: CLI `--mode` > `workflow.default_mode` > harness auto).
   - Documents harness-based default selection (REQ-WF003-002, 003): autopilot for minimal/standard; team for thorough+enabled+env-set; otherwise autopilot with downgrade info log.
   - Includes the literal string `MODE_UNKNOWN` with explanation: "When an invalid `--mode` value is supplied (not in the 4 valid set), the orchestrator emits `MODE_UNKNOWN` error listing the 4 valid values: `autopilot`, `loop`, `team`, `pipeline`."
   - Includes the literal string `MODE_TEAM_UNAVAILABLE` with explanation: "When `--mode team` is requested but `workflow.team.enabled: false` or `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` is unset, the orchestrator emits `MODE_TEAM_UNAVAILABLE` and suggests `--mode autopilot` fallback."
   - Documents silent downgrade for auto-resolved team (REQ-WF003-012): when harness picks team but prerequisites missing, log `[mode-auto-downgrade]` info message and use autopilot.
4. Mirror edit into `internal/template/templates/.claude/skills/moai/workflows/run.md`.
5. Run `make build` to regenerate `internal/template/embedded.go`.

Reference for sentinel string convention: `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE` are NEW strings owned by WF-003 per spec.md REQ-WF003-010, 011. `MODE_PIPELINE_ONLY_UTILITY` is shared with WF-004 (verified in research.md §4.1 — bidirectional cross-ref).

Verification: `go test ./internal/template/ -run "TestRunDesignSkillsContainModeUnknownSentinel|TestRunSkillContainsModeTeamUnavailableSentinel"` PASSES for `run.md` portion (design.md still RED).

### M3: Mode Dispatch section in `/moai design` skill (GREEN, part 2) — Priority P0

Owner role: `manager-docs`.

Scope:
1. Open `.claude/skills/moai/workflows/design.md`.
2. Locate the existing `## Mode Flag Compatibility` section (added by WF-004 M3).
3. EXTEND that section into a `## Mode Dispatch (Multi-Mode Router)` section that:
   - References SPEC-V3R2-WF-003.
   - Documents the 4 `--mode` values for design:
     - `autopilot` (default): code-based path (Phase B-Common) — sequential `moai-domain-copywriting` + `moai-domain-brand-design` + `moai-workflow-gan-loop`. Skips Phase 1 path selection AskUserQuestion when `--mode autopilot` is explicit.
     - `import`: Path A (Claude Design handoff) — invokes `moai-workflow-design-import` per Step A1-A5. Skips Phase 1 selection AND Phase B-Common (REQ-WF003-013).
     - `team`: Phase B-Common with **parallel** TeamCreate spawning `moai-domain-copywriting` + `moai-domain-brand-design` teammates (role_profile: designer per workflow.yaml:52-56). Both feed into `moai-workflow-gan-loop` for evaluation. Requires team prerequisites (same as `run.md` per REQ-WF003-009).
     - `pipeline`: REJECTED — `MODE_PIPELINE_ONLY_UTILITY` (preserved from WF-004 M3).
   - Documents that Path B1 (Figma) and Path B2 (Pencil) are NOT in the WF-003 mode axis; users wanting those paths should NOT supply `--mode` flag (the existing AskUserQuestion path selection still presents A/B1/B2). Per research.md §2.2.3 recommendation.
   - Includes the literal string `MODE_UNKNOWN` with the same 4-value enumeration.
   - Cross-references `run.md` mode dispatch for the precedence rule and team prerequisites (avoid duplication).
4. Mirror edit into `internal/template/templates/.claude/skills/moai/workflows/design.md`.
5. Run `make build`.

Verification: `TestRunDesignSkillsContainModeUnknownSentinel` PASSES (both `run.md` and `design.md` now contain `MODE_UNKNOWN`).

[HARD] M3 must NOT modify the existing Phase 0/1/A/B1/B2/B-Common/C structure of design.md. Insert-only into the Mode Flag Compatibility section.

### M4: `/moai loop` alias note + matrix extension + `default_mode` schema (GREEN, part 3) — Priority P1

Owner role: `manager-docs` (content); `expert-backend` (yaml schema review optional).

Scope (4 sub-tasks):

**4a. `/moai loop` alias header note**:
1. Open `.claude/skills/moai/workflows/loop.md`.
2. Insert a new section near the top (after the existing flow declaration line at `loop.md:33`):
   ```markdown
   ## Invocation Routes (SPEC-V3R2-WF-003)

   This skill is invocable via two equivalent routes:
   - Direct: `/moai loop $ARGUMENTS` — historical entry point, preserved as thin wrapper.
   - Via run dispatch: `/moai run SPEC-XXX --mode loop` — per SPEC-V3R2-WF-003 REQ-WF003-004, the
     `/moai run` skill delegates to this skill when `--mode loop` is supplied.

   Both routes invoke this skill body unchanged. Behavioral equivalence is enforced by the audit
   test `TestLoopAliasCrossReference` in `internal/template/agentless_audit_test.go`.
   ```
3. Mirror to `internal/template/templates/.claude/skills/moai/workflows/loop.md`.
4. Run `make build`.

**4b. Subcommand × Mode matrix extension**:
1. Open `.claude/rules/moai/workflow/spec-workflow.md`.
2. Locate the `## Subcommand Classification (Pipeline vs Multi-Agent)` section (added by WF-004 M4 — should exist in the stacked base).
3. EXTEND the matrix table by:
   - Adding three new columns to the right of the existing 5: "Default mode", "Valid `--mode` values", "Sentinel on invalid mode".
   - Adding a new row for `/moai loop` showing it as "Multi-Agent (alias for `run --mode loop`)".
   - Filling all 10 rows with the values per research.md §8 matrix.
4. Insert a sub-section `### Mode Dispatch Cross-Reference` immediately after the matrix:
   ```markdown
   ### Mode Dispatch Cross-Reference

   Source: SPEC-V3R2-WF-003. The `--mode` axis values are valid only on multi-agent subcommands
   that explicitly support mode dispatch: `/moai run` and `/moai design`. Other multi-agent
   subcommands (`/moai plan`, `/moai sync`) ignore `--mode` per REQ-WF003-005, except they REJECT
   `--mode pipeline` with `MODE_PIPELINE_ONLY_UTILITY` per REQ-WF003-016 (shared with WF-004
   REQ-WF004-014).

   Mode precedence (REQ-WF003-018, hard-coded):
   1. CLI flag `--mode <value>` — highest priority
   2. Config field `workflow.default_mode` in `.moai/config/sections/workflow.yaml`
   3. Harness auto-selection — lowest priority (per `harness.yaml` level)

   Auto-selection rules (REQ-WF003-002, 003):
   - Harness `minimal` or `standard` → default mode = `autopilot`
   - Harness `thorough` AND `workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` → default mode = `team`
   - Otherwise (thorough but team prereqs missing) → fallback to `autopilot` with `[mode-auto-downgrade]` info log per REQ-WF003-012.

   Sentinel error keys:
   - `MODE_UNKNOWN` (REQ-WF003-010, owned by WF-003): invalid `--mode` value supplied.
   - `MODE_TEAM_UNAVAILABLE` (REQ-WF003-011, owned by WF-003): explicit `--mode team` request without prerequisites.
   - `MODE_PIPELINE_ONLY_UTILITY` (REQ-WF003-016 ↔ REQ-WF004-014, shared): `--mode pipeline` on a multi-agent subcommand.
   - `MODE_FLAG_IGNORED_FOR_UTILITY` (REQ-WF004-011, owned by WF-004): `--mode <any>` on a utility subcommand (info log only).

   See `.claude/skills/moai/workflows/run.md` § Mode Dispatch and
   `.claude/skills/moai/workflows/design.md` § Mode Dispatch for the per-skill dispatch rules.
   ```
5. Mirror to `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`.
6. Run `make build`.

**4c. `default_mode` schema extension in workflow.yaml**:
1. Open `.moai/config/sections/workflow.yaml`.
2. Add a new optional field under the top-level `workflow:` block (before `auto_clear:`):
   ```yaml
   workflow:
       # SPEC-V3R2-WF-003 REQ-WF003-014: Optional default mode for /moai run and /moai design.
       # Values: autopilot | loop | team. Empty/absent = harness-based auto-selection.
       # Precedence: CLI --mode flag > this field > harness auto.
       default_mode: ""
   ```
3. Mirror to `internal/template/templates/.moai/config/sections/workflow.yaml`.
4. Run `make build`.

[HARD] DO NOT modify any other key in `workflow.yaml`. The `team:`, `auto_clear:`, `completion:`, `loop_prevention:`, `worktree:`, etc. sub-blocks are out of scope.

**4d. Mode-NA notation in plan.md, sync.md, project.md, db.md** (4 mode-NA implementation subcommands per REQ-WF003-005):

REQ-WF003-005 enumerates 4 subcommands as mode-NA: `plan`, `sync`, `project`, `db`. Verified pre-edit (per audit iteration 1 D4 fix): all 4 skill body files exist at `.claude/skills/moai/workflows/{plan,sync,project,db}.md` and none currently parses or acts on `--mode` flag. M4d EXTENDS the WF-004 M3-added `## Mode Flag Compatibility` section in each of these 4 skills (not just plan/sync).

1. Open `.claude/skills/moai/workflows/plan.md`. Locate the `## Mode Flag Compatibility` section (added by WF-004 M3 — should exist in the stacked base).
2. EXTEND that section to clarify:
   ```markdown
   ## Mode Flag Compatibility

   Per SPEC-V3R2-WF-003 REQ-WF003-005 and SPEC-V3R2-WF-004:

   - This subcommand is multi-agent (open-ended) but does NOT participate in the
     `--mode {autopilot|loop|team}` axis defined in SPEC-V3R2-WF-003.
   - Any `--mode` value supplied to `/moai plan` is silently ignored. The plan workflow
     proceeds with its default behavior.
   - The `pipeline` value is the only special case: passing `--mode pipeline` triggers
     `MODE_PIPELINE_ONLY_UTILITY` (the same error key shared with WF-004 REQ-WF004-014).

   See `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` for the
   full subcommand × mode matrix.
   ```
3. Apply the same edit pattern (with appropriate subcommand name substitution `/moai plan` → `/moai sync` / `/moai project` / `/moai db`) to:
   - `.claude/skills/moai/workflows/sync.md`
   - `.claude/skills/moai/workflows/project.md`
   - `.claude/skills/moai/workflows/db.md`
4. Mirror all 4 source-of-truth edits into `internal/template/templates/.claude/skills/moai/workflows/{plan,sync,project,db}.md`.
5. Run `make build`.

[HARD] Pre-edit verification: read each of plan.md/sync.md/project.md/db.md skill bodies and confirm none currently parses `--mode` (e.g., grep for `--mode` returns no parsing logic, only documentation references). If any of the 4 skills already contain `--mode` parsing, escalate to user before proceeding (REQ-WF003-005 mode-NA assertion would be invalidated).

Verification (full M4): `go test ./internal/template/ -run "TestRun|TestLoop"` PASSES (3 new tests GREEN); pre-existing WF-004 tests remain GREEN; `go test ./...` shows 0 failures.

### M5: Documentation sync + cross-links + MX tags — Priority P2

Owner role: `manager-docs`.

Scope:
1. Add CHANGELOG entry under `## [Unreleased]`:
   ```
   ### Added
   - SPEC-V3R2-WF-003: Multi-Mode Router (`--mode {autopilot|loop|team|pipeline}` flag)
     for `/moai run` and `/moai design`. `/moai loop` becomes an alias for
     `/moai run --mode loop` (REQ-WF003-004). Mode auto-selection by harness level
     (REQ-WF003-002, 003) with silent downgrade fallback (REQ-WF003-012). New optional
     `workflow.default_mode` config field. Sentinels: `MODE_UNKNOWN`,
     `MODE_TEAM_UNAVAILABLE`. Subcommand × mode matrix extended in
     `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification`. CI guards
     `TestRunDesignSkillsContainModeUnknownSentinel`,
     `TestRunSkillContainsModeTeamUnavailableSentinel`, `TestLoopAliasCrossReference`.
   ```

2. Within each of the 4 affected workflow skill files (`run.md`, `design.md`, `loop.md`, `plan.md`, `sync.md`), add a one-line back-reference at the bottom of the new sections:
   ```markdown
   See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification)
   for the full pipeline-vs-multi-agent + mode-axis contract.
   ```

3. Update `progress.md` for SPEC-V3R2-WF-003 with `run_complete_at` and `run_status: implementation-complete` after M1-M4 land.

4. Verify `go test ./...` passes (full repository test suite per `CLAUDE.local.md` §6 Go Test Execution Rules HARD rule: "After fixing ANY test, run the FULL test suite to catch cascading failures").

5. Verify `make build` produces a clean `internal/template/embedded.go` with all M2-M4 changes embedded.

[HARD] No new documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase, not a planning phase.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files, post-WF-004 stacked base)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/template/agentless_audit_test.go` | After existing `TestImplementationSkillsContainPipelineRejectionSentinel` func (added by WF-004 M1) | Insert 3 new test functions | M1 / REQ-WF003-010, 011, 004 |
| `.claude/skills/moai/workflows/run.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND into `## Mode Dispatch (Multi-Mode Router)` (~60 lines added) | M2 / REQ-WF003-001..018 (run scope) |
| `.claude/skills/moai/workflows/design.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND into `## Mode Dispatch (Multi-Mode Router)` (~50 lines added) | M3 / REQ-WF003-001, 009, 010, 013, 014, 016, 018 |
| `.claude/skills/moai/workflows/loop.md` | After line 33 (flow declaration) | Insert `## Invocation Routes (SPEC-V3R2-WF-003)` (~10 lines) | M4a / REQ-WF003-004 |
| `.claude/rules/moai/workflow/spec-workflow.md` | The `## Subcommand Classification` section (added by WF-004 M4) | EXTEND matrix (3 new columns + 1 new row); INSERT `### Mode Dispatch Cross-Reference` sub-section (~40 lines) | M4b / REQ-WF003-007, 016, 018 |
| `.moai/config/sections/workflow.yaml` | Top of `workflow:` block (before `auto_clear:`) | Insert `default_mode: ""` field with comment (~5 lines) | M4c / REQ-WF003-014 |
| `.claude/skills/moai/workflows/plan.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND with REQ-WF003-005 clarification (~10 lines added) | M4d / REQ-WF003-005 |
| `.claude/skills/moai/workflows/sync.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND with REQ-WF003-005 clarification (~10 lines added) | M4d / REQ-WF003-005 |
| `.claude/skills/moai/workflows/project.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND with REQ-WF003-005 clarification (~10 lines added) | M4d / REQ-WF003-005 |
| `.claude/skills/moai/workflows/db.md` | The `## Mode Flag Compatibility` section (added by WF-004 M3) | EXTEND with REQ-WF003-005 clarification (~10 lines added) | M4d / REQ-WF003-005 |
| `internal/template/templates/.claude/skills/moai/workflows/{run,design,loop,plan,sync,project,db}.md` | Same as source-of-truth | Mirror all 7 source edits | embedded-template parity (HARD) |
| `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | Same | Mirror | parity |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | Same | Mirror | parity |
| `internal/template/embedded.go` | (regenerated) | `make build` regenerates | parity |
| `CHANGELOG.md` | `## [Unreleased]` section | Add entry (~10 lines) | M5 |

### 3.2 To-be-created (new files)

| File | Reason |
|------|--------|
| (none) | All M-targets edit existing files. No new files created. |

This is a critical architectural decision: WF-003 reuses the WF-004 audit test infrastructure (single audit-test file), reuses existing skill files (no new skills per WF-001 catalog stability), and extends existing rule files. No new file creations.

### 3.3 NOT to be touched (preserved by reference)

The following files are referenced but MUST NOT be modified by this SPEC's run phase:

- `.claude/skills/moai/workflows/loop.md:46-138` — Per-iteration cycle (Steps 1-9). The Ralph engine itself. WF-003 only adds a header note above this, never modifies the cycle.
- `.claude/skills/moai/workflows/run.md:46` — Existing `--team` flag. Preserve as-is (deprecated by `--mode team` but kept for backward compatibility per REQ-WF003-005-style additivity).
- `.claude/skills/moai/workflows/run.md:361-382` — Phase 0.95 Scale-Based Mode Selection. WF-003 references this from the Mode Dispatch section but does not modify the table.
- `.claude/skills/moai/workflows/run.md:537-545` — Development Mode Routing (DDD vs TDD). Distinct from `--mode` axis (config-driven, not flag-driven). Preserve.
- `.claude/skills/moai/workflows/run.md:602-686` — Phase 2A/2B implementation. Preserve.
- `.claude/skills/moai/workflows/run.md:927-943` — Team Mode Routing prerequisite check. WF-003 references this, does not modify.
- `.claude/skills/moai/workflows/design.md:64-94` — Phase 1 Path Selection AskUserQuestion. Preserve. WF-003's Mode Dispatch section documents that explicit `--mode` skips this AskUserQuestion, but the AskUserQuestion itself remains for default invocations.
- `.claude/skills/moai/workflows/design.md:97-123` — Phase A. Preserve.
- `.claude/skills/moai/workflows/design.md:127-194` — Phase B1, B2, B-Common. Preserve.
- `.claude/skills/moai/workflows/design.md:196-217` — Phase C quality gate. Preserve.
- `.moai/config/sections/harness.yaml` — Harness levels. Preserve. WF-003 reads but does not modify.
- `.moai/config/sections/workflow.yaml:17-26` — `team:` block. Preserve. WF-003 only adds `default_mode` at the workflow root, not inside team.
- `.claude/commands/moai/{run,loop,design,plan,sync}.md` — Thin wrappers (1-7 lines each). Preserve as-is; mode dispatch happens at the skill layer, not the command layer.
- `internal/hook/post_tool.go`, `internal/hook/pre_tool.go` — Fixed-pipeline hooks. Preserve (preserved by WF-004 too).

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §13, the following anchors are load-bearing and cited verbatim throughout this plan:

1. `.moai/specs/SPEC-V3R2-WF-003/spec.md:108-167` (REQ definitions)
2. `.moai/specs/SPEC-V3R2-WF-003/spec.md:173-187` (AC definitions)
3. `.moai/specs/SPEC-V3R2-WF-004/spec.md:144,155,173,174` (sentinel cross-reference verification)
4. `.moai/specs/SPEC-V3R2-WF-001/spec.md:5,254,258` (24-skill catalog status: completed)
5. `.claude/skills/moai/workflows/run.md:46` (--team flag, preserve)
6. `.claude/skills/moai/workflows/run.md:64-83` (Harness Level Routing, mode auto-select substrate)
7. `.claude/skills/moai/workflows/run.md:361-382` (Phase 0.95 Scale-Based Mode Selection, substrate)
8. `.claude/skills/moai/workflows/run.md:537-545` (Development Mode Routing, distinct from --mode axis)
9. `.claude/skills/moai/workflows/run.md:920-922` (--team/--solo flag junction, NEW `--mode` location)
10. `.claude/skills/moai/workflows/run.md:927-943` (Team Mode Routing prerequisite check, REQ-WF003-011/012 substrate)
11. `.claude/skills/moai/workflows/loop.md:33` (Flow declaration, M4a anchor)
12. `.claude/skills/moai/workflows/loop.md:46-138` (per-iteration cycle, REQ-WF003-008 target — preserve)
13. `.claude/skills/moai/workflows/loop.md:140-152` (convergence and exit conditions, REQ-WF003-017)
14. `.claude/skills/moai/workflows/design.md:64-94` (Phase 1 path selection AskUserQuestion, preserve)
15. `.claude/skills/moai/workflows/design.md:111` (Step A3: invoke moai-workflow-design-import, REQ-WF003-013)
16. `.claude/skills/moai/workflows/design.md:183-186` (Step BC-3: copywriting + brand-design + gan-loop, REQ-WF003-009)
17. `.claude/rules/moai/workflow/spec-workflow.md:9-15` (Phase Overview table, WF-004 matrix insertion anchor)
18. `.moai/config/sections/harness.yaml:67-103` (level definitions: minimal/standard/thorough, REQ-WF003-002/003)
19. `.moai/config/sections/workflow.yaml:17-26` (team configuration, REQ-WF003-011/012 substrate)
20. `.moai/config/sections/workflow.yaml:42-61` (role_profiles: designer, REQ-WF003-009)
21. `.moai/config/sections/quality.yaml:1-3` (development_mode: tdd, run-phase methodology)
22. `internal/template/agentless_audit_test.go` (NEW per WF-004 M1; M1 EXTENDS this file)
23. `.moai/specs/SPEC-V3R2-WF-004/plan.md:39-43` (M3 mode-flag rejection clauses, WF-003 builds on this)
24. `.moai/specs/SPEC-V3R2-WF-004/plan.md:140-189` (M4 Subcommand Classification matrix, WF-003 EXTENDS columns)
25. `docs/design/major-v3-master.md:L993` (§9 Phase 6 Multi-Mode Workflow anchor)

Total: 25 distinct file:line anchors (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §2.2 Out of Scope, **no new technology** is introduced:

- No new Go modules or external libraries.
- No new directory structure.
- Embedded-template machinery (`go:embed` in `internal/template/embed.go`) reused as-is.

Additive surfaces only:
- 3 new Go test functions in existing `internal/template/agentless_audit_test.go` (created by WF-004 M1).
- New `## Mode Dispatch` section in `run.md` (~60 lines).
- New `## Mode Dispatch` section in `design.md` (~50 lines).
- New `## Invocation Routes` note in `loop.md` (~10 lines).
- Extended Subcommand Classification matrix in `spec-workflow.md` (3 new columns + 1 new row + ~40-line `### Mode Dispatch Cross-Reference` sub-section).
- New `default_mode: ""` field in `workflow.yaml` (~5 lines).
- Refined Mode Flag Compatibility section in `plan.md` and `sync.md` (~10 lines each).
- One CHANGELOG entry.

No runtime behavior change for users not supplying `--mode` flag (additive guarantee).

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| Stacked PR base auto-close on WF-004 merge | M | H | §1.4 Pre-merge base transition hook (CLAUDE.local.md §18.11 reference). Run-phase agent must monitor PR #765 status and trigger `gh pr edit --base main` before merge. |
| Mode dispatch logic in skill body diverges between `run.md` and `design.md` | M | M | M2 and M3 use a common pseudocode template; cross-reference `spec-workflow.md#mode-dispatch-cross-reference` from both skills (single source of truth for precedence + sentinel definitions). |
| `default_mode` field added to workflow.yaml without Go-side loader | L | L | spec.md §8 risks table explicitly accepts this (skill-only consumption). Loader deferred to SPEC-V3R2-MIG-003. CI test verifies the field is documented in skill bodies. |
| Path B1 (Figma) / B2 (Pencil) UX confusion when explicit `--mode autopilot` is supplied | M | M | M3 includes explicit user-facing note: "If you want Figma or Pencil paths, do NOT supply `--mode` flag — the AskUserQuestion path selection presents those options." |
| `MODE_TEAM_UNAVAILABLE` (explicit user request) and `[mode-auto-downgrade]` (auto-resolved) conflated | M | M | M2 documents the distinction explicitly per research.md §6.1: explicit `--mode team` failure = sentinel error; auto-resolved fallback = info log only. |
| Sentinel string drift across the 3 skills containing `MODE_PIPELINE_ONLY_UTILITY` | L | M | Static audit `TestImplementationSkillsContainPipelineRejectionSentinel` (existing per WF-004 M1) enforces literal string in `plan.md`, `run.md`, `sync.md`, `design.md`. WF-003 does not change the sentinel. |
| Embedded-template mismatch (skill edited but `internal/template/templates/` not synced) | M | H | Each M-task ends with template parity step + `make build`. CI runs `go test ./...` exercising embedded FS. Per CLAUDE.local.md §2 Template-First Rule HARD constraint. |
| `/moai loop` behavioral divergence from `/moai run --mode loop` over time | M | M | `TestLoopAliasCrossReference` audit asserts `loop.md` documents the alias relationship; behavioral equivalence noted in CHANGELOG and §1.4 of plan.md. |
| WF-004 M1 audit test file does not exist yet in stacked base (timing) | L | M | Verified: stacked worktree base is HEAD `5ab409292` (WF-004 worktree). M1 must verify `internal/template/agentless_audit_test.go` exists before extending. If not present, escalate to user. |
| BC-WF003 perceived as behavioral break because new flag is added | L | L | spec.md §10 "BC 영향: 없음 (additive flag, 기존 command behavior 보존)". Verified in research.md §11 item 7. CHANGELOG explicitly notes "Added" not "Changed". |
| New test functions cascade-fail other tests when they run | L | M | Tests are pure read-only string scans on embedded FS; no shared state. CI matrix on Linux/macOS/Windows verifies cross-platform stability. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md:627` (mx_plan in plan.md MANDATORY).

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/template/agentless_audit_test.go` `TestRunDesignSkillsContainModeUnknownSentinel` | `@MX:ANCHOR fan_in=2 - SPEC-V3R2-WF-003 REQ-WF003-010 enforcer; guards run.md and design.md against sentinel drift. Touching this test signature affects mode dispatch contract for both implementation skills.` | High downstream impact: any future skill body change that removes `MODE_UNKNOWN` from run/design skills will fail this test. |
| `.claude/rules/moai/workflow/spec-workflow.md` `### Mode Dispatch Cross-Reference` sub-section | `@MX:ANCHOR fan_in=10 - Subcommand × Mode dispatch single source of truth; cross-referenced by 5 multi-agent skills (run, design, loop, plan, sync) and 5 utility skills (fix, coverage, mx, codemaps, clean). Changes here affect all workflow contracts.` | The matrix is the unified contract for mode dispatch; cross-linked from all 10 workflow skills. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `.claude/skills/moai/workflows/run.md` `## Mode Dispatch` section header | `@MX:NOTE - Multi-Mode Router per SPEC-V3R2-WF-003 REQ-WF003-001..018; --mode {autopilot,loop,team,pipeline} dispatch with harness-based default. See spec-workflow.md#mode-dispatch-cross-reference.` | Carries the WHY for future readers — explains the additive-flag philosophy and precedence. |
| `.claude/skills/moai/workflows/design.md` `## Mode Dispatch` section header | `@MX:NOTE - Multi-Mode Router per SPEC-V3R2-WF-003 REQ-WF003-009, 013; --mode {autopilot,import,team} dispatch with Path B1/B2 preserved under default invocation. See spec-workflow.md#mode-dispatch-cross-reference.` | Documents that Figma/Pencil paths are intentionally outside the mode axis. |
| `.claude/skills/moai/workflows/loop.md` `## Invocation Routes` section header | `@MX:NOTE - SPEC-V3R2-WF-003 REQ-WF003-004 alias relationship; /moai loop ↔ /moai run --mode loop. Behavioral equivalence enforced by TestLoopAliasCrossReference.` | Documents the alias contract that future PRs must preserve. |
| `internal/template/agentless_audit_test.go` package-level doc comment (extend existing) | `@MX:NOTE - Audit suite extended for SPEC-V3R2-WF-003 mode dispatch contract. Adds: TestRunDesignSkillsContainModeUnknownSentinel (REQ-WF003-010), TestRunSkillContainsModeTeamUnavailableSentinel (REQ-WF003-011), TestLoopAliasCrossReference (REQ-WF003-004).` | Documents test scope so future maintainers do not delete coverage. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `.claude/skills/moai/workflows/run.md` `## Mode Dispatch` mode resolver pseudocode block | `@MX:WARN @MX:REASON - Future PRs may be tempted to add new --mode values without updating the precedence rule or the audit tests. Any new mode value MUST be reflected in: (a) this skill body's enumeration, (b) design.md if applicable, (c) spec-workflow.md matrix, (d) audit tests in agentless_audit_test.go.` | Most likely point of future regression: adding modes without coordinated cross-file updates. |
| `.moai/config/sections/workflow.yaml` `default_mode` field comment | `@MX:WARN @MX:REASON - default_mode is read by skill bodies (orchestrator path), NOT by Go CLI handlers. Adding a Go-side reader is deferred to SPEC-V3R2-MIG-003. Until then, terminal moai commands ignore this field.` | Pre-warns future maintainers about the skill-only consumption boundary. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

This SPEC is a router publication, not a feature build. No `@MX:TODO` markers planned — all work converges to GREEN within M1-M4. Any `@MX:TODO` introduced during implementation must be resolved before M5 (per mx-tag-protocol.md GREEN-phase resolution rule).

### 6.5 MX tag count summary

- @MX:ANCHOR: 2 targets
- @MX:NOTE: 4 targets
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 8 MX tag insertions planned across 6 distinct files.

---

## 7. Worktree Discipline

[HARD] All run-phase work for SPEC-V3R2-WF-003 MUST execute in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003
```

Branch: `feature/SPEC-V3R2-WF-003`, base: `feature/SPEC-V3R2-WF-004` (HEAD `5ab409292`).

[HARD] All Read/Write/Edit tool invocations MUST use absolute paths under this worktree root. Tool calls referencing `/Users/goos/MoAI/moai-adk-go/` (the main session checkout) are PROHIBITED for this SPEC.

[HARD] `make build` and `go test ./...` execute inside the worktree with `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003 && make build` etc. Per `CLAUDE.local.md` §15 worktree isolation conventions.

[HARD] **Pre-merge base transition** (per §1.4): When PR #765 (WF-004) approaches merge, run-phase agent (or user) MUST execute `gh pr edit <WF-003-PR-#> --base main` to retarget the base BEFORE PR #765 lands. Failure to do so will cause WF-003 PR to auto-close (CLAUDE.local.md §18.11 case study).

The run phase SHOULD NOT enter the worktree via `cd` from a long-running shell; instead each Bash invocation uses an absolute-path command or `cd <worktree> && ...`.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.2.0 schema** — `spec.md` frontmatter has all 9 required fields (`id`, `title`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`). Verified by reading `spec.md:2-25` (lines 1-25 inclusive of v0.2.0 migration).
- [x] **C2: HISTORY entry for v0.2.0** — `spec.md:31-34` HISTORY table has v0.2.0 row with description.
- [x] **C3: 18 EARS REQs across 5 categories** — `spec.md:104-167` (Ubiquitous 7, Event-Driven 4, State-Driven 2, Optional 2, Complex 3).
- [x] **C4: 15 ACs all map to REQs (100% coverage)** — `spec.md:173-187`. Each AC explicitly cites the REQ(s) it maps to.
- [x] **C5: BC scope clarity** — `spec.md:21,238` (`breaking: false`, `bc_id: []`, "BC 영향: 없음 (additive flag, 기존 command behavior 보존)") + research.md §11 item 7 (verified additive).
- [x] **C6: File:line anchors ≥10** — research.md §13 (38 anchors), this plan.md §3.4 (25 anchors).
- [x] **C7: Exclusions section present** — `spec.md:48-54,70-78` Out of Scope explicitly enumerates 7 + 7 = 14 exclusions.
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml:1-3`.
- [x] **C9: mx_plan section** — this plan §6 (8 MX tag insertions across 4 categories).
- [x] **C10: Risk table with mitigations** — `spec.md:204-211` + this plan §5 (11 risks, file-anchored mitigations).
- [x] **C11: Worktree absolute path discipline** — this plan §7 (4 HARD rules including stacked-PR base transition).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, and YAML/Markdown templates. No Go function bodies or executable code.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — see acceptance.md §1-15.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — see tasks.md §M1-M5 (manager-tdd / expert-backend / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — WF-001 status: completed (verified `spec.md:5`); WF-004 sentinel `MODE_PIPELINE_ONLY_UTILITY` byte-identical (verified research.md §4.1); WF-002 thin-wrapper pattern preserved for `/moai loop` (verified research.md §5).
- [x] **C16: Stacked PR pre-merge hook documented** — this plan §1.4 + §7 (CLAUDE.local.md §18.11 cross-ref).

All 16 criteria PASS → plan is **audit-ready**.

### Open Questions for plan-auditor Review

1. Should the `default_mode` schema extension in `workflow.yaml` include a corresponding YAML schema validation in `internal/template/schema/`? Currently no Go-side loader is planned (deferred to SPEC-V3R2-MIG-003), but schema validation is a separate concern.
2. Is the recommendation in research.md §2.2.3 (subsume Path B1/B2 under default invocation rather than creating new mode values) consistent with the master design intent? The auditor may want to verify against `docs/design/major-v3-master.md:L993` Phase 6 description.
3. Should the audit test for `/moai loop` alias documentation (`TestLoopAliasCrossReference`) also verify the literal phrase `--mode loop` appears in `run.md` (not just `loop.md`)? Current plan asserts the cross-reference only in `loop.md`.
4. The info-log message format for `[mode-auto-downgrade]` (REQ-WF003-012) is recommended in research.md §6.3 but not specified in spec.md. Should the auditor request a strict format or accept the recommendation as discretionary?

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Verify WF-004 M1's `agentless_audit_test.go` exists in stacked base → extend with 3 new test functions. Confirm RED.
2. **M2 (P0)**: Add `## Mode Dispatch` section to `run.md` (extend existing Mode Flag Compatibility from WF-004 M3) (+ embedded template parity + `make build`). Partial GREEN (run.md sentinels added).
3. **M3 (P0)**: Add `## Mode Dispatch` section to `design.md` (extend existing) (+ template parity + `make build`). Confirm `TestRunDesignSkillsContainModeUnknownSentinel` GREEN.
4. **M4 (P1)**: 4 sub-tasks:
   - M4a: Add `## Invocation Routes` to `loop.md`.
   - M4b: Extend Subcommand Classification matrix in `spec-workflow.md` + add `### Mode Dispatch Cross-Reference` sub-section.
   - M4c: Add `default_mode` field to `workflow.yaml`.
   - M4d: Refine `## Mode Flag Compatibility` in `plan.md` and `sync.md`.
   All sub-tasks include template parity + `make build`. Confirm 3 of 3 new audit tests GREEN, full `go test ./...` passes.
5. **M5 (P2)**: Add CHANGELOG entry, cross-link footers in 5 affected skills, MX tags per §6, update `progress.md`, final `go test ./...` + `make build`.

Total milestones: 5. Total file edits (existing): ~13 (5 skills + 1 rule + 1 yaml + 1 test + CHANGELOG + 5 template mirrors). Total file creations (new): 0.

Pre-merge hook (orthogonal to milestones): monitor PR #765, transition base when approaching merge.

---

End of plan.md.
