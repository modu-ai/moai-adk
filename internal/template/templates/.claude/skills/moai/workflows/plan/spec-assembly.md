---
description: "Plan Phase 1.5/2/2.3/2.5/3/3.5/3.6/DP2/DP3/DP3.5 — Pre-creation validation, SPEC document creation, independent review, GitHub Issue creation, Git environment setup, MX tag planning, quality gate, and execution mode selection"
user-invocable: false
metadata:
  parent: moai-workflow-plan
  phase: "Phase 1.5 through Decision Point 3.5: SPEC Assembly, Review, and Environment Setup"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

### Phase 1.5: Pre-Creation Validation Gate

Purpose: Prevent common SPEC creation errors before file generation.

Step 1 - Document Type Classification:
- Detect keywords to classify as SPEC, Report, or Documentation
- Reports route to .moai/reports/, Documentation to .moai/docs/
- Only SPEC-type content proceeds to Phase 2

Step 2 - SPEC ID Validation (all checks must pass):
- ID Format: Must match SPEC-{DOMAIN}-{NUMBER} pattern (e.g., SPEC-AUTH-001)
- Domain Name: Must be from the approved domain list (AUTH, API, UI, DB, REFACTOR, FIX, UPDATE, PERF, TEST, DOCS, INFRA, DEVOPS, SECURITY, and others)
- ID Uniqueness: Search .moai/specs/ to confirm no duplicates exist
- Directory Structure: Must create directory, never flat files

Composite domain rules: Maximum 2 domains recommended, maximum 3 allowed.

### Phase 1.6: Tier Judgment Socratic Question (LEAN Workflow)

[ZONE:Evolvable] [HARD] Before artifact creation begins, the orchestrator MUST present a Tier judgment AskUserQuestion to classify the SPEC's complexity tier (S, M, or L). This drives the artifact set, the manager-develop delegation prompt template applicability, and the plan-auditor PASS threshold. Origin: the LEAN-tier workflow policy.

Skip condition: when the user explicitly provided the tier in the original request (e.g., "Tier S", "small SPEC, Tier S"), the orchestrator MAY skip the question and record the user-provided tier directly.

Tier judgment AskUserQuestion (Socratic, in conversation_language):

```
Question: "Estimated SPEC complexity tier?"
Header: "Complexity tier"
Options (4 max, recommended first):
  Option 1: "Tier S (Simple, <300 LOC, <5 files) — 2 artifacts: spec.md + plan.md (AC inline) (권장 for text-only/refactor)"
  Option 2: "Tier M (Medium, 300-1000 LOC, 5-15 files) — 3 artifacts: spec.md + plan.md + acceptance.md"
  Option 3: "Tier L (Large, >1000 LOC or constitutional, >15 files) — 5 artifacts (default, current behavior)"
  Option 4 (auto-appended by Claude Code): "Other"
```

LOC thresholds are guidance, not enforcement. The implementer's domain knowledge supplements the question (e.g., a 200-LOC SPEC touching a frozen constitutional zone may warrant Tier L).

Persistence: the chosen tier is written to `spec.md` frontmatter as `tier: S | M | L` (optional field per `.claude/rules/moai/development/spec-frontmatter-schema.md`). When the user provides "Other" or skips, the orchestrator defaults to Tier L (backward compat).

Tier-conditional artifact set (driven by user response):

- **Tier S** → Phase 2 creates only `spec.md` + `plan.md`; AC is inline in `spec.md §3`. No `acceptance.md`, no `design.md`, no `research.md`. Phase 2.3 plan-auditor uses 0.75 PASS threshold.
- **Tier M** → Phase 2 creates `spec.md` + `plan.md` + `acceptance.md`. Phase 2.3 plan-auditor uses 0.80 PASS threshold.
- **Tier L** → Phase 2 creates full 5-artifact set (spec.md + plan.md + acceptance.md + design.md + research.md). Phase 2.3 plan-auditor uses 0.85 PASS threshold (preserves pre-LEAN strict behavior).

Anti-pattern: classifying a 1000+ LOC SPEC as Tier S to skip overhead. Mitigation: plan-auditor first-pass score regression triggers tier-up suggestion to the user.

Reference: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L).

### Phase 2: SPEC Document Creation

Agent: manager-spec subagent

Input: Approved plan from Phase 1B, validated SPEC ID from Phase 1.5.

File generation (all three files created simultaneously):

- .moai/specs/SPEC-{ID}/spec.md
  - YAML frontmatter with **12 required fields** (canonical schema — see checklist below and `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields)
  - HISTORY section immediately after frontmatter
  - Complete GEARS structure with the 5 GEARS patterns (Ubiquitous, Event-driven `When`, State-driven `While`, Capability-gate `Where`, Event-detected unwanted — see `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format). EARS legacy form is accepted for pre-v3 SPECs until 2026-11-22 per the GEARS migration policy.
  - Content written in conversation_language
  - **Epic reference**: when the SPEC belongs to a multi-SPEC grouping, `plan.md §A Context` references the **Epic** (not the retired `Sprint`/`cohort`/`Wave` aliases) per `.claude/rules/moai/development/sprint-round-naming.md`. A standalone SPEC with no grouping is also valid.

#### [HARD] Pre-Write Frontmatter Checklist

[HARD] Before manager-spec calls Write/MultiEdit for spec.md, it MUST validate the frontmatter contains ALL 12 required fields AND rejects snake_case legacy aliases. This checklist prevents dual-schema drift between the plan workflow and the SPEC frontmatter lint rule. SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.

Required 12 fields (canonical order):
- [ ] `id: SPEC-{DOMAIN}-{NUM}` — matches `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`
- [ ] `title: "Human-readable title"` — quoted string
- [ ] `version: "X.Y.Z"` — quoted semver string (NOT `0.1` unquoted)
- [ ] `status: draft` — enum: draft | planned | in-progress | implemented | completed | superseded | archived | rejected
- [ ] `created: YYYY-MM-DD` — ISO date (NEVER `created_at`, NEVER `date`)
- [ ] `updated: YYYY-MM-DD` — ISO date (NEVER `updated_at`)
- [ ] `author: <name>` — string, not empty
- [ ] `priority: P1` — uppercase Pn style (P0|P1|P2|P3) or High|Medium|Low|Critical
- [ ] `phase: "vX.Y.Z target"` — release phase string
- [ ] `module: "path/to/module"` — affected module path
- [ ] `lifecycle: spec-anchored` — enum: spec-anchored | spec-lite | exploratory
- [ ] `tags: "tag1, tag2, ..."` — comma-separated string (NOT `labels:`, NOT YAML array)

Optional fields (do NOT include unless needed):
- `issue_number: null` — integer or null (omit entirely when not tracking GitHub issue)

Rejected legacy aliases (fail closed — do NOT accept):
- `created_at:` (use `created:`)
- `updated_at:` (use `updated:`)
- `labels:` (use `tags:`)
- `spec_id:` (use `id:`)

Pre-write gate behavior:
1. manager-spec generates frontmatter draft in memory.
2. manager-spec self-audits against the 12-field checklist above.
3. If any required field is missing OR any rejected alias appears: manager-spec HALTS, reports the schema violation, and re-generates. It does NOT call Write.
4. Phase 2.3 plan-auditor independently re-verifies the schema on the written file as a second line of defense.

- .moai/specs/SPEC-{ID}/plan.md
  - Implementation plan with task decomposition
  - Technology stack specifications and dependencies
  - Risk analysis and mitigation strategies

- .moai/specs/SPEC-{ID}/acceptance.md
  - Minimum 2 Given/When/Then test scenarios
  - Edge case testing scenarios
  - Performance and quality gate criteria

### Delta Markers for Brownfield Projects

When the SPEC modifies existing code (detected via research.md analysis), apply delta markers:

```
### [DELTA] {Module Name}
- [EXISTING] {description} - unchanged context, characterization tests only
- [MODIFY] {description} - existing code to change, requires characterization tests before modification
- [NEW] {description} - new code to create, full implementation + new tests
- [REMOVE] {description} - code to delete, requires dependency analysis and migration verification
```

Delta markers are OPTIONAL and only suggested for brownfield projects. Greenfield projects skip this.

### spec-compact.md Auto-Generation

After all SPEC files are created, auto-generate `.moai/specs/SPEC-{ID}/spec-compact.md`:

Extract from spec.md:
- All REQ-XXX requirements (GEARS-notation entries — EARS legacy form accepted for pre-v3 SPECs until 2026-11-22)
- All acceptance criteria (Given/When/Then scenarios)
- Files to modify list
- Exclusions (What NOT to Build) section

Exclude: Overview, technical approach, research references, annotation history.

Purpose: Run phase loads spec-compact.md (~30% token savings) instead of full spec.md.
Fallback: If generation fails, Run phase uses full spec.md.

Quality constraints:
- Requirement modules limited to 5 or fewer per SPEC
- Acceptance criteria minimum 2 Given/When/Then scenarios
- Technical terms and function names remain in English
- Exclusions section MUST contain at least 1 entry

### Phase 2.3: Independent SPEC Review (Conditional)

Purpose: Prevent confirmation bias by running an adversarial audit of the just-created SPEC before user approval and GitHub Issue creation. The reviewer sees only the final spec.md — not the author's reasoning — and is prompted to find defects, not rationalize acceptance.

Execution conditions:
- `harness.yaml` `levels.{current_level}.plan_audit.enabled` is `true`
- Current harness level is `standard` or `thorough` (default: enabled)
- SPEC files were successfully created in Phase 2

Skip conditions:
- Harness level is `minimal` (fast iteration path, plan_audit.enabled: false)
- `--no-review` flag is present in $ARGUMENTS
- spec.md was not created (Phase 2 failed)

#### Step 2.3.1: Invoke plan-auditor

Agent: plan-auditor subagent

Delegation pattern: "Use the plan-auditor subagent to audit the SPEC at .moai/specs/{SPEC-ID}/ — this is iteration 1."

Do NOT pass the author's reasoning or conversation history to plan-auditor. The agent enforces context isolation (M1) and will ignore injected reasoning. Pass only the SPEC directory path.

#### Step 2.3.2: Read Verdict

After plan-auditor completes, read the report at `.moai/reports/plan-audit/{SPEC-ID}-review-1.md`.

Extract the verdict line: `Verdict: PASS | FAIL`

#### Step 2.3.3: PASS Path

If verdict is PASS: proceed directly to Phase 2.5 (GitHub Issue Creation).

Log: "SPEC review passed (iteration 1). Proceeding to Phase 2.5."

#### Step 2.3.4: FAIL Path — Retry Loop (max 3 iterations)

If verdict is FAIL:

1. Delegate back to manager-spec: "Use the manager-spec subagent to revise .moai/specs/{SPEC-ID}/spec.md based on the review report at .moai/reports/plan-audit/{SPEC-ID}-review-{N}.md. Address all defects listed in the report. DO NOT implement any code."

2. After manager-spec revision, re-invoke plan-auditor: "Use the plan-auditor subagent to audit .moai/specs/{SPEC-ID}/ — this is iteration {N+1}. Previous review report: .moai/reports/plan-audit/{SPEC-ID}-review-{N}.md"

3. Read new verdict from `.moai/reports/plan-audit/{SPEC-ID}-review-{N+1}.md`.

4. Repeat until PASS or 3 iterations exhausted.

Iteration tracking: Display "SPEC review iteration {N}/3" after each verdict.

#### Step 2.3.5: Escalation after 3 FAIL Iterations

If all three iterations result in FAIL, do NOT proceed to Phase 2.5 automatically.

Present the full defect history to the user:
- Show `.moai/reports/plan-audit/{SPEC-ID}-review-1.md` through `-review-3.md`
- Summarize blocking defects that persisted across all iterations
- Use AskUserQuestion with options:
  - Force-accept SPEC with known defects (proceed to Phase 2.5): "Accept SPEC with known defects — I will fix them manually before /moai run"
  - Request manual SPEC revision: "I will manually edit the SPEC — re-run review after my edits"
  - Abort plan workflow: "Abort — start over with a clearer feature description"

Harness configuration reference (harness.yaml):
- `minimal`: plan_audit.enabled: false (skip this entire phase)
- `standard`: plan_audit.enabled: true, max_iterations: 3, require_must_pass: true
- `thorough`: plan_audit.enabled: true, max_iterations: 3, require_must_pass: true, cross_validate_with_evaluator_active: true

For `thorough` harness with `cross_validate_with_evaluator_active: true`: after plan-auditor PASS, additionally invoke sync-auditor in SPEC-review mode to cross-validate must-pass criteria. If sync-auditor disagrees with plan-auditor's PASS, treat as FAIL and trigger one additional iteration.

### Phase 2.5: GitHub Issue Creation (Conditional, opt-in)

Purpose: Create a GitHub Issue linked to the SPEC document for bidirectional traceability between planning artifacts and issue tracker.

[HARD] Per the late-branch opt-in policy, this phase MUST default to a silent skip. The flag semantics are now opt-in: `--issue` activates creation; the absence of `--issue` skips the entire phase. The legacy `--no-issue` opt-out is no longer required because skipping is the default. SPEC frontmatter MUST NOT carry an `issue_number` field for new SPECs (D2 — `issue_number` field-removal prospective only; existing SPECs retain the field per EXCL-LB-008).

Execution conditions (ALL must hold):
- `--issue` flag IS set (explicit opt-in)
- GitHub CLI (`gh`) is available
- Repository has a remote origin

Skip conditions (any triggers a silent skip with no warning):
- `--issue` flag is NOT set (default — silent skip)
- `gh` CLI not available (log warning and continue)
- No remote origin configured

#### Step 2.5.1: Create GitHub Issue

Agent: manager-git subagent

[HARD] Gate: only proceed when the `--issue` flag is set. The `gh issue create` invocation below is the ONLY occurrence in this workflow and MUST be guarded by the explicit `--issue` opt-in. Default invocation of `/moai plan` (no `--issue`) silently skips Step 2.5.1/2.5.2/2.5.3.

Create a GitHub Issue from SPEC metadata (only executed when `--issue` flag is set):

```bash
# Pre-check: this block runs only when --issue flag is present.
# If --issue is absent, Phase 2.5 is silently skipped — no gh issue create occurs.
gh issue create \
  --title "[SPEC-{ID}] {SPEC title}" \
  --body "$(cat <<'EOF'
## SPEC Reference

- **SPEC ID**: SPEC-{ID}
- **Status**: draft
- **Priority**: {priority}
- **Created**: {created_date}

## Requirements Summary

{Brief summary from spec.md GEARS-notation requirements (EARS legacy form for pre-v3 SPECs)}

## Acceptance Criteria

{Summary from acceptance.md}

---

*This issue was automatically created by MoAI from SPEC-{ID}.*
*SPEC location: `.moai/specs/SPEC-{ID}/spec.md`*
EOF
)" \
  --label "spec"
```

#### Step 2.5.2: Update SPEC Metadata

After Issue creation, update the SPEC frontmatter with the issue number:

- Read the issue number from `gh issue create` output
- Update `issue_number` field in `.moai/specs/SPEC-{ID}/spec.md` YAML frontmatter
- Add cross-reference comment in the Issue: `gh issue comment {number} --body "SPEC document: .moai/specs/SPEC-{ID}/spec.md"`

#### Step 2.5.3: Bidirectional Reference

The SPEC ↔ Issue link enables:
- SPEC spec.md frontmatter contains `issue_number: {N}` for downstream workflows
- GitHub Issue body contains SPEC-ID and file path for human navigation
- run.md Phase 3 uses `issue_number` to include `Fixes #{N}` in commits/PRs
- sync.md leverages `Fixes #{N}` in PR for automatic Issue closure on merge

### Phase 3: Git Environment Setup (Conditional)

Execution conditions: Phase 2 completed successfully AND one of the following:
- --worktree flag provided
- --branch flag provided or user chose branch creation
- Configuration permits branch creation (git_strategy settings)

Skipped when: develop_direct workflow, no flags and user chooses "Use current branch".

#### Late-branch Pre-check [HARD] (the late-branch opt-in policy)

Before evaluating any of the paths below (Worktree / Branch / Current Branch), the orchestrator MUST read `team.branch_creation.auto_enabled` from `.moai/config/sections/git-strategy.yaml`.

- When `auto_enabled == false`: SKIP branch creation entirely. Cwd remains on the current branch (typically `main` — the default for Step 1 per SPEC Phase Discipline). SPEC files are committed to the current branch via the standard commit pipeline. Phase 3.5 mode-selection display MUST surface "Late-branch (main commit + late switch)" as the active mode to communicate the deferred-branch state to the user. Phase 3.0 BODP Gate STILL runs (with EntryPoint = `EntryPlanLateBranch`) — Late-branch does NOT bypass BODP, it only defers branch creation to Phase C (manual `git switch -c feat/SPEC-*` at PR time). Per the late-branch opt-in policy, no automated `git push origin main` is performed during this phase even if `team.automation.auto_push == true`.

- When `auto_enabled == true`: continue to the Worktree/Branch/Current Branch path evaluation below (existing behavior, unchanged).

Reference: see `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern for the 4-phase procedure (A→D) the user follows after this skill defers branch creation.

#### Phase 3.0: BODP Gate (공통)

Both Worktree Path and Branch Path execute this gate immediately before delegating worktree/branch creation. Source: the CI-autonomy policy W7-T02.

Steps:

1. **Relatedness Check** — Orchestrator calls `internal/bodp/Check()` with `CheckInput{CurrentBranch, NewSpecID, RepoRoot, EntryPoint}` (`EntryPlanBranch` for Branch Path; `EntryPlanWorktree` for Worktree Path). Result: `BODPDecision{SignalA, SignalB, SignalC, Recommended, Rationale, BaseBranch}`.

2. **AskUserQuestion Gate** — Orchestrator-only HARD (see `.claude/rules/moai/core/askuser-protocol.md`):
   - Preload: `ToolSearch(query: "select:AskUserQuestion")`.
   - Options (max 4, conversation_language=ko):
     - First option: the recommended Choice with `(권장)` suffix; description = `BODPDecision.Rationale`.
     - Remaining options: the other Choice values (e.g. when Recommended is `ChoiceMain`, present `ChoiceStacked` and `ChoiceContinue`).
   - The "Other" option is auto-appended by Claude Code.
   - User response yields the chosen Choice + base branch.

3. **Audit Trail Write** — Call `internal/bodp.WriteDecision()` with EntryPoint matching the path (`EntryPlanBranch` or `EntryPlanWorktree`), `UserChoice` from the AskUserQuestion answer, and `ExecutedCmd` describing the upcoming git operation. Failure is non-fatal.

4. **Path-Specific Delegation** — Branch Path: pass `base=<chosenBase>` parameter to `manager-git`. Worktree Path: invoke `moai worktree new <SPEC-ID> --base <chosenBase>` (or `--from-current` when chosenBase is `HEAD`).

Out of Scope (BODP Gate):
- "Other" free-form base interpretation: orchestrator parses input as a base branch name; invalid input falls back to `origin/main` with a warning.
- Concurrent invocation safety: single-session orchestrator assumed (W7-R5 follow-up).

#### Worktree Path (--worktree flag)

Prerequisite: SPEC files MUST be committed before worktree creation.
- Run **Phase 3.0: BODP Gate** above (EntryPoint = `EntryPlanWorktree`).
- Stage SPEC files: git add .moai/specs/SPEC-{ID}/
- Create commit: feat(spec): Add SPEC-{ID} - {title}
- Create worktree: `moai worktree new SPEC-{ID} --base <chosenBase>` (or `--from-current` when the user chose to continue on the current HEAD).
- Display worktree path and navigation instructions

##### Worktree-Anchored Resume Output [HARD]

When `--worktree` is used, the plan-phase output MUST include a paste-ready resume message with **Block 0 (cwd anchoring)** prepended before the standard 6-block structure. This anchors the user to start the next session inside the worktree, preventing main-cwd drift.

Block 0 format (prepended before Block 1):

```
[New Terminal — START IN WORKTREE]
$ cd <worktree-absolute-path>
$ <session-launcher>            # claude | moai cc | moai cg | moai glm
   └─ Claude Code session starts here (cwd = worktree)
```

Block 4 (preconditions) MUST include `0)` as the first item:

```
0) git rev-parse --show-toplevel → <worktree-path> (★ critical pre-check)
```

Recommended session-launcher per execution mode:

- `--team` → `tmux new-session -s moai-<spec> && moai cg` (teammate spawn via tmux split-window inherits worktree cwd + tmux session env)
- single-session → `moai cc` (or `claude`) directly inside worktree
- GLM-only → `moai glm`

[HARD] Single-session corollary: If the user is NOT comfortable with multi-terminal/multi-session workflow, recommend converting to `--branch` next time. `--worktree` only realizes its isolation value when the user actually starts a separate session inside the worktree path. Forcing Block 0 onto a single-session user is friction without benefit.

See `.claude/rules/moai/workflow/session-handoff.md` "Worktree-Anchored Resume Pattern" for the canonical Block 0 specification and lessons #14 for the failure-mode rationale.

#### Branch Path (--branch flag or user choice)

Agent: manager-git subagent
- Run **Phase 3.0: BODP Gate** above (EntryPoint = `EntryPlanBranch`).
- Delegate to manager-git with `base=<chosenBase>` derived from the gate answer.
- Create branch: feature/SPEC-{ID}-{description} from `<chosenBase>`.
- Set tracking upstream if remote exists.
- Switch to new branch.
- Team mode: Create draft PR via manager-git subagent.

#### Current Branch Path (no flag or user choice)

- No branch creation, no manager-git invocation
- SPEC files remain on current branch

### Phase 3.5: MX Tag Planning [MANDATORY]

Purpose: Identify code locations that will need @MX annotations during implementation. This information is passed to run workflow agents as context constraints.

Execution conditions: Always executed. Depth varies by scope:
- **Full scan**: SPEC involves modifying existing code OR creating new public APIs
- **Lightweight scan**: New feature with no existing code interaction (scan public API surface only)

Tasks:
- Scan target files for high fan_in functions (potential @MX:ANCHOR)
- Identify dangerous patterns (goroutines, complexity) for @MX:WARN
- List magic constants and business rules for @MX:NOTE
- Document MX tag strategy in `plan.md`
- Output: `mx_plan` section in SPEC document with annotation targets and priorities

### Phase 3.6: SPEC Quality Gate

<!-- moai:evolvable-start id="gate-plan-2" -->
### HUMAN GATE: SPEC Quality Validation

**Previous phase output:** Validated SPEC with quality score
**Approval question:** Is the SPEC ready for execution mode selection and implementation?
**Cannot proceed until:**
- [ ] SPEC quality gate shows PASS
- [ ] No HARD rule violations detected
- [ ] User has selected execution mode (sub-agent vs team)
<!-- moai:evolvable-end -->

Purpose: Verify SPEC document quality before proceeding to implementation. Catches incomplete or inconsistent specs early.

Tasks:
- Verify all GEARS-notation requirements (EARS legacy form for pre-v3 SPECs) have corresponding acceptance criteria
- Check that affected files list is complete (cross-reference with codebase)
- Validate that MX tag plan covers all high-risk areas (fan_in >= 3, goroutines)
- Run lightweight security check on SPEC scope (flag if auth/crypto/input-validation areas are touched)

Gate decision:
- **PASS**: All checks satisfied. Proceed to Decision Point 2.
- **WARNING**: Minor gaps found (e.g., missing acceptance criteria for edge cases). Present findings and offer fix or continue.
- **FAIL**: Critical gaps (e.g., no acceptance criteria, security-sensitive scope without security considerations). Must fix before proceeding.

Tool: AskUserQuestion (when WARNING or FAIL)
Options:
- Fix SPEC issues (Recommended): Return to SPEC editing with specific gaps highlighted
- Continue with warnings: Proceed knowing gaps exist (WARNING only, not available for FAIL)
- Abort: Exit plan workflow

### Decision Point 2: Development Environment Selection

Tool: AskUserQuestion (when prompt_always config is true and auto_branch is true)

Options:
- Create Worktree (recommended for parallel SPEC development)
- Create Branch (traditional workflow)
- Use current branch

### Decision Point 3: Next Action Selection

Tool: AskUserQuestion (after SPEC creation completes)

Options:
- Start Implementation (execute /moai run SPEC-{ID})
- Modify Plan
- Add New Feature (create additional SPEC)

### Decision Point 3.5: Execution Mode Selection Gate

Triggered when: User selects "Start Implementation" in Decision Point 3.

Purpose: After SPEC creation, detect execution environment and present optimal implementation mode.

**Step 1: Detect active LLM mode**
Read `.moai/config/sections/llm.yaml` → `llm.team_mode` field:
- `""` (empty) or `"cc"`: CC mode (Claude-only)
- `"glm"`: GLM mode (GLM-only)
- `"cg"`: CG mode (Claude Leader + GLM Workers)

**Step 2: Detect tmux availability**
Check `$TMUX` environment variable via Bash: `test -n "$TMUX" && echo "tmux" || echo "no-tmux"`

**Step 3: Present options based on detection**

When tmux IS available: AskUserQuestion with 3 options (descriptions adapt to active_mode):
- Option 1 (Recommended): Worktree + {active_mode}
  - CC: "Create MoAI worktree with tmux session. All agents use Claude. Highest quality."
  - GLM: "Create MoAI worktree with tmux session. All agents use GLM. Cost optimized."
  - CG: "Create MoAI worktree with tmux session. Leader=Claude, Workers=GLM. Balanced quality-cost."
- Option 2: Team Mode (in-process): Use Agent Teams for parallel implementation within current session. Best for multi-domain features.
- Option 3: Sub-agent Mode (sequential): Use sequential sub-agents. Best for simple, single-domain tasks.

When tmux is NOT available: AskUserQuestion with 2 options:
- Option 1 (Recommended): Sub-agent Mode: Use sequential sub-agents for implementation. Tmux is not available for session isolation.
- Option 2: Team Mode (in-process): Use Agent Teams for parallel implementation within current session.

**Step 4: Execute selected mode**
- **Worktree mode**: Execute `moai worktree new SPEC-{ID} --tmux` to create worktree with tmux session. The tmux session will:
  - CC mode: Create session, cd to worktree, run `/moai run SPEC-{ID}`
  - GLM mode: Create session, inject GLM env, cd to worktree, run `/moai run SPEC-{ID}`
  - CG mode: Create session, inject GLM env to session, clear GLM from settings.local.json, cd to worktree, run `/moai run SPEC-{ID}`
  - Display: "Implementation started in tmux session: moai-{ProjectName}-{SPEC-ID}"
- **Team mode**: Proceed to `/moai run SPEC-{ID} --team`
- **Sub-agent mode**: Proceed to `/moai run SPEC-{ID} --solo`

**Step 5: Gate result passing**
- Pass the selected execution mode to the run workflow
- If worktree mode: Run workflow executes in the tmux session (no further action needed from plan)
- If team/sub-agent mode: Continue to run workflow in current session

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-spec analyzed project and proposed SPEC candidates
- User approval obtained via AskUserQuestion before SPEC creation
- Phase 2: All SPEC files created (spec.md, plan.md, acceptance.md, spec-compact.md)
- Directory naming follows .moai/specs/SPEC-{ID}/ format
- YAML frontmatter contains all 8 required fields (including issue_number)
- GEARS structure is complete (EARS legacy form accepted for pre-v3 SPECs until 2026-11-22)
- Exclusions section present with at least 1 entry
- Delta markers applied for brownfield requirements (if applicable)
- spec-compact.md auto-generated with requirements + acceptance criteria only
- Phase 2.5: GitHub Issue created and linked (unless --no-issue)
- Phase 3: Appropriate git action taken based on flags and user choice
- If --worktree: SPEC committed before worktree creation
- Next steps presented to user
- **Audit-ready signal**: Before transitioning to `/moai run`, append to `.moai/specs/SPEC-{ID}/progress.md`:
  ```
  - plan_complete_at: {ISO-8601 timestamp}
  - plan_status: audit-ready
  ```
  This signal indicates plan artifacts (spec.md, plan.md, acceptance.md, tasks.md) are finalized
  and ready for Plan Audit Gate validation at `/moai run` Phase 0.5.

---

## Test Scenarios

### Normal Flow
**Prompt**: "/moai plan JWT authentication with refresh token rotation"
**Expected Result**:
- Phase 1A: Explore discovers existing auth files if any
- Phase 1B: manager-spec designs GEARS-notation requirements for JWT auth (canonical notation; EARS legacy form for pre-v3 SPECs only)
- Annotation cycle: 1-3 iterations refining requirements
- Phase 2: SPEC-AUTH-001 created with spec.md, plan.md, acceptance.md
- Phase 2.5: GitHub Issue created and linked to SPEC
- Phase 3: Feature branch feat/SPEC-AUTH-001-jwt-auth created (if --branch)

### Existing Assets Flow
**Prompt**: "/moai plan add payment gateway" (existing e-commerce codebase)
**Expected Result**:
- Explore discovers existing order, product, user models
- SPEC references existing models as dependencies
- plan.md identifies extension points in existing architecture
- No duplicate functionality proposed
### Error Flow
**Prompt**: "/moai plan" (no description provided)
**Expected Result**:
- AskUserQuestion prompts user for feature description
- After user provides description, normal flow continues
- If user cancels, graceful exit with no files created
