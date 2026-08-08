---
description: "Run Mode Routing — Execution mode gate integration, mode dispatch routing, context propagation, completion criteria, the verify exit gate of run-phase, test scenarios, and custom harness extension"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Mode Routing: Execution Mode Gate, Mode Dispatch, Completion, and Scenarios"
---

# Execution Mode Gate Integration

When the run phase is invoked from plan.md Decision Point 3.5 or moai.md step 11.5, the gate passes these parameters:
- `execution_mode`: worktree | team | sub-agent
- `active_mode`: cc | glm | cg
- `tmux_available`: true | false

**If execution_mode == "worktree":**
This run invocation is already inside the isolated tmux session and worktree.
Proceed with standard sub-agent run phase in the current environment.
No additional routing needed — CC/GLM/CG env is already configured by the Gate.

**If execution_mode == "team":**
The `team` execution mode is RETIRED (Agent Teams static layer). Emit the
canonical sentinel `MODE_TEAM_UNAVAILABLE` and fall back to the standard
sub-agent run phase (Phase 5 Strategy). The `active_mode` (cc / glm / cg) still
selects the backend for the native `moai cg` teammate runtime, which is
unaffected by this retirement.

**If execution_mode == "sub-agent":**
Proceed directly to Phase 5 (Strategy).

**If no execution_mode provided (direct `/moai run` invocation):**
Standard sub-agent run phase. A forced `--mode team` emits
`MODE_TEAM_UNAVAILABLE` and falls back to `autopilot`; `--solo` is the explicit
sub-agent selector.

---

# Mode Dispatch (team dispatch retired)

The `--mode team` dispatch value is RETIRED: Mode 3 (`agent-team`) of the Phase
0.95 catalog was retired with the Agent Teams static layer
(`.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.1). A forced
`--mode team` emits the canonical sentinel `MODE_TEAM_UNAVAILABLE` (per
`.claude/rules/moai/workflow/spec-workflow.md` § Mode Dispatch) and the
orchestrator falls back to `autopilot` with a `[mode-auto-downgrade]` info log.
The native Claude Code teammate runtime (`moai cg` GLM panes, `moai cc -w <name>
--spawn` teammate windows) is unaffected — only MoAI's static team-orchestration layer is retired.

All worktree path rules from context-loading.md "Worktree Path Rules [HARD] (All
Modes)" continue to apply to every execution mode.

---

# Context Propagation

Context flows forward through every phase:

- Phase 5 to Phase 11: Execution plan with architecture decisions guides implementation
- Phase 11 to Phase 13: Implementation code plus planning context enables context-aware validation
- Phase 13 to Phase 19: Quality findings enable semantically meaningful commit messages
- Phase 11 to /moai sync: Implementation divergence report enables accurate SPEC and project document updates

---

# Completion Criteria

All of the following must be verified:

- Phase 5: manager-spec returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 11 until user confirmed
- Phase 6: Tasks decomposed with requirement traceability
- Phase 9: MX context map built for target files (skipped for greenfield)
- Phase 11: Implementation completed according to development_mode (with MX context)
- Phase 13: sync-auditor (or orchestrator verification batch) completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 19 if status was CRITICAL
- Phase 19: manager-git created commits (branch or direct) only if quality permitted
- Phase 20: Next step honors the pipeline contract — `full-pipeline` auto-chains into `/moai sync` (announced in the transcript); `single-phase` presents sync as the "(Recommended)" first next-step option (never a silent chain)

---

## Verify Exit Gate (factory contract)

The `factory` pipeline contract (`workflows/moai.md` § run→sync chaining policy) adds exactly one stage to run-phase: a security verify stage that is the **exit gate of run-phase**. It is not a sync-phase stage and it is not a new subcommand. Ordering: the gate fires after acceptance-criterion convergence and BEFORE the inherited run→sync auto-chain, and the whole of run-phase — this gate included — is downstream of Implementation Kickoff Approval.

Invocation, verbatim:

```text
/moai review --security --deep --repo
```

### Severity partition — three cases, mutually exclusive and jointly exhaustive

Every verify stage lands in exactly one of the three cases below, and in no more than one. Readability is what separates S3 from S2: a stage with no readable result must never be absorbed by the S2 "no confirmed findings" wording, because a gate whose failure mode is *proceed* is not a gate.

| Case | Condition | Route |
|---|---|---|
| **S1** | a readable result carrying one or more CONFIRMED findings at `critical` or `high` | the chain **shall not proceed to sync**; the findings **re-enter run-phase scoped to the changed surface** |
| **S2** | a readable result carrying findings only at `medium` or `low`, or a readable result carrying no confirmed findings at all | proceed to sync, carrying the findings forward as **inherited sync-phase evidence** |
| **S3** | **no readable result** — the invocation errored, the pipeline aborted, the recorded results directory is absent, or its findings artifact is absent or carries a line that does not parse | **HALT** the chain; it does not proceed to sync |

S1 is the single human gate this contract adds — the operator decides whether to re-enter. S3 **HALT**s, emits the 5-section verdict (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per `verification-claim-integrity.md` §3), and escalates; an S3 attempt does **not** count against the re-entry ceiling, because a stage that produced nothing consumed none of the re-entry budget.

The chain permits **at most two verify re-entries**. When a third would be required, halt and emit that same 5-section verdict, then escalate.

### Rung attribute — orthogonal to the severity partition

The rigor rung (`PRIMARY`, `FALLBACK`, or `DEGRADED`, as self-labelled by the review degradation ladder) is an **attribute of** an S1 or S2 result, recorded on the factory state record. It is not a fourth case standing beside S1/S2/S3: a readable result with no confirmed findings at the `DEGRADED` rung is S2 *and* `DEGRADED` simultaneously. S3 produced no result and therefore carries no rung at all.

A `DEGRADED` rung (single-pass, no voter panel) is surfaced both in the chain transcript and in the sync report, and forces the sync-phase security-analysis suppression OFF, so the independent adversarial analysis of the same surface still runs.

**Precedence, stated in both directions.** The severity case **governs routing** and the rung never changes it; the rung **governs suppression** and the severity case never relaxes it. An S2 result at the `DEGRADED` rung therefore proceeds to sync with suppression forced off.

---

# Test Scenarios

## Normal Flow
**Prompt**: "/moai run SPEC-AUTH-001"
**Expected Result**:
- Phase 3: Detects Go project (go.mod) → references `.claude/rules/moai/languages/go.md`
- Phase 4: SPEC has 8 files, 2 domains → Standard Mode selected
- Phase 5: manager-spec creates execution plan with 5 tasks
- Decision Point: User approves plan
- Phase 11: Implementation via manager-develop (DDD mode)
- Phase 13: TRUST 5 validation passes
- Phase 19: Commits created on feature branch

## Fix Mode Flow
**Prompt**: "/moai run SPEC-BUG-042" (bug fix SPEC, 2 files affected)
**Expected Result**:
- Phase 4: SPEC has 2 files, 1 domain → Fix Mode selected
- Directly spawns manager-develop + orchestrator verification batch (lint + test + coverage)
- Minimal overhead, fast execution
- Quality validation still runs

## Error Flow
**Prompt**: "/moai run SPEC-NONEXISTENT"
**Expected Result**:
- SPEC directory not found in .moai/specs/
- AskUserQuestion: "SPEC not found. Create it with /moai plan?"
- If user confirms, redirect to plan workflow

---

Version: 2.11.0
Changes: Added Phase 3 JIT Language Detection, Phase 4 Scale-Based Mode Selection, test scenarios.

---

# Custom Harness Extension (Optional)

@.moai/harness/run-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
