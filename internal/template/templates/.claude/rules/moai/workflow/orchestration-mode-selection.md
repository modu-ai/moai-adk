---
description: "Phase 0.95 Mode Selection — 6-mode autonomous decision tree for the MoAI orchestrator (trivial / background / agent-team / parallel / sub-agent / workflow). Read at every /moai run entry."
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/plan.md,.claude/rules/moai/workflow/spec-workflow.md"
metadata:
  version: "1.1.0"
  status: "active"
  tags: "orchestration, mode-selection, agent-teams, workflow, phase-0.95"
---

# Orchestration Mode Selection — Phase 0.95

Canonical 6-mode autonomous decision tree for the MoAI orchestrator. Activated at Phase 0.95 (after Phase 0.5 plan-auditor verdict, before Phase 1 implementation). The decision is autonomous (no `AskUserQuestion` round); the chosen mode and the selection rationale are logged to `progress.md § Mode Selection`.

[ZONE:Frozen] [HARD] All Phase 0.95 execution modes are strictly downstream of GATE-2 (the plan→run HUMAN GATE). The orchestrator reaches Phase 0.95 ONLY after GATE-2 user approval has already been obtained. Mode selection — including Mode 6 (workflow) — is never a substitute for GATE-2 and never a path that crosses the plan→run boundary ahead of the human gate. GATE-2 is mandatory and score-independent (a plan-auditor PASS or a high skip-eligible score never auto-bypasses it; skip-eligibility applies only to Phase 0.5 verdict re-execution, not to GATE-2) per the GATE-2 mandatory-restoration policy.

> Cross-reference: `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification covers the `--mode` flag matrix (autopilot / loop / team / pipeline) which interacts with — but is separate from — the 6-mode catalog below. The run-phase `/goal ac_converge` autonomy wiring point lives in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge); `.claude/rules/moai/workflow/dynamic-workflows.md` is the source for the Workflow (Mode 6) primitive (16-concurrent / 1000-total cap) and the named-script-API prohibition.

---

## §A — Mode Catalog (6 modes)

The orchestrator selects exactly one of the following modes per Phase 0.95 invocation.

| # | Mode | Concurrency | Spawn Surface | When to prefer |
|---|------|-------------|---------------|----------------|
| 1 | `trivial` | None — direct execution by the orchestrator, no sub-agent spawn | n/a | Typo fix, single-line formatting, no semantic change |
| 2 | `background` | 1 concurrent sub-agent | `Agent(run_in_background: true, ...)` | Read-only analysis that can complete asynchronously without blocking the conversation |
| 3 | `agent-team` | 3-5 dynamic teammates | `TeamCreate(...)` + `Agent(subagent_type: "general-purpose", team_name: ..., name: ...)` × N | Multi-domain (≥3 domains OR ≥10 files) research-heavy work AND all Agent Teams capability-gate prerequisites met |
| 4 | `parallel` | 3-5 concurrent sub-agents (single message, multiple `Agent()` calls) | Multiple `Agent()` invocations in one assistant turn | Multi-domain research that does NOT meet Agent Teams prerequisites; or any case where Agent Teams session overhead exceeds benefit |
| 5 | `sub-agent` | 1 sequential sub-agent per milestone | Sequential `Agent(...)` spawns, one milestone at a time | Coding-heavy work (per Anthropic's coding-task parallelism caveat), or any case where the simpler mode suffices |
| 6 | `workflow` | Up to 16 concurrent workflow agents (1000-total per-run backstop, per `dynamic-workflows.md`) | Orchestrator-launched Workflow fan-out (a script the runtime executes to coordinate agents — NOT a subagent spawning subagents) | Genuinely-parallel, high-volume **mechanical** transformation (≥ ~30 files AND a single uniform transform rule AND no inter-file dependency) — call-site rename, import-path bulk change, signature-stable edits. Coding-heavy / multi-domain / new-code work stays Mode 5 (per Anthropic's coding-task parallelism caveat). |

Mode 5 is the **default fallback** when no other mode's selection criteria are unambiguously met. Mode 6 (`workflow`) is the narrow high-volume-mechanical exception, selectable ONLY after GATE-2 has passed (see §C.3).

---

## §B — Decision Tree

```
START (Phase 0.95 Mode Selection)
  │
  ├── Is task trivial (typo, single-line, no semantic change)?
  │   ├── YES → Mode 1: TRIVIAL (direct execution, no Agent() spawn)
  │   └── NO  → continue
  │
  ├── Can the task complete async without blocking the user
  │   AND is the work read-only (no Write/Edit)?
  │   ├── YES → Mode 2: BACKGROUND (Agent run_in_background: true)
  │   └── NO  → continue
  │
  ├── Does the task meet ALL Agent Teams capability-gate conditions?
  │   (all three required):
  │     • harness level is `thorough` (`.moai/config/sections/harness.yaml`)
  │     • `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml`
  │     • environment variable `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
  │   AND is the scope multi-domain (≥3 domains OR ≥10 files)?
  │   ├── YES → Mode 3: AGENT-TEAM (TeamCreate + dynamic teammate spawn)
  │   └── NO  → continue
  │
  ├── Is the task multi-domain (≥3 domains) AND research-heavy
  │   (NOT coding-heavy per Anthropic's coding-task parallelism caveat)?
  │   ├── YES → Mode 4: PARALLEL (3-5 concurrent Agent() in single message)
  │   └── NO  → continue
  │
  ├── Is the task ≥ ~30 files AND mechanical (one uniform transform rule)
  │   AND genuinely parallel — no inter-file dependency
  │   AND GATE-2 already passed AND all preferences already collected
  │   AND Workflows available (not disabled, runtime version ≥ v2.1.154)?
  │   ├── YES → Mode 6: WORKFLOW (orchestrator-launched fan-out, scaling NOT nesting)
  │   └── NO  → continue
  │
  └── Default → Mode 5: SUB-AGENT (single Agent() sequential spawn per milestone)
```

Mode 6 is checked AFTER Mode 4 and BEFORE the Mode 5 default fallback. Coding-heavy or multi-domain or new-code work that reaches this branch falls through to Mode 5 — Mode 6 admits ONLY the genuinely-parallel high-volume mechanical case (per Anthropic's coding-task parallelism caveat: most coding tasks involve fewer truly parallelizable tasks than research, so the sequential sub-agent path is the safe default for coding work).

### §B.1 Input parameters

The orchestrator collects the following signals before traversing the decision tree:

- **tier**: SPEC tier (S / M / L) per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- **scope (file count)**: estimated count of files in the SPEC's run-phase scope
- **domain count**: number of distinct domains touched (agents, workflow skills, rules, hook scripts, template mirrors, Go source code, SPEC artifacts, etc.)
- **file language mix**: e.g., 100% markdown vs Go source vs shell vs mixed
- **concurrency benefit**: HIGH for research-heavy work (parallel reads, independent perspectives); LOW for coding-heavy work (Anthropic's coding-task parallelism caveat — most coding tasks involve fewer truly parallelizable tasks than research)
- **Agent Teams prereqs status**: each of the three Agent Teams capability-gate conditions individually verified

### §B.2 Tie-breaker rules (boundary cases)

Phase 0.95 boundary cases (scope at threshold ±1, ambiguous domain count, etc.) follow these defaults:

- At threshold ±1 (9 vs 10 files; 2 vs 3 domains): default to the **simpler** mode (sub-agent over agent-team; sequential over parallel)
- **Coding-heavy + multi-domain**: prefer Mode 5 over Mode 4 (Anthropic's coding-task parallelism caveat)
- **Markdown-heavy + multi-domain + research-heavy**: prefer Mode 4 (parallel multi-spawn)
- **Mode 6 soft `~30`-file boundary**: the `≥ ~30 files` Mode 6 entry threshold is tilde-prefixed (soft). At the boundary (exactly 30 files), the "default to the simpler mode" rule resolves toward Mode 5 — the tilde avoids a hard cliff and keeps even large work on the safer sequential path unless the transformation is genuinely mechanical-uniform.
- **Mode 6 vs Mode 5 (transformation kind)**: even at high file counts, if the work is semantic / new-code / multi-rule, prefer Mode 5; Mode 6 admits ONLY a single uniform mechanical transform rule with no inter-file dependency.
- **Workflows disabled or unavailable**: when `CLAUDE_CODE_DISABLE_WORKFLOWS=1` is set OR the runtime version is below v2.1.154, Mode 6 is not selectable and the task falls through to Mode 5 (cannot assume Workflow availability).
- **Tier L + markdown / shell-script-only scope**: Mode 5 with Tier L Section A-E delegation template (per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability)
- **Tier S + minimal scope**: Mode 5 with the minimal delegation form (~500-800 tokens, Section B may be filtered)

---

## §C — Capability Gates

### §C.1 Mode 3 (Agent Teams) capability gate

Mode 3 is candidate only when all three conditions hold simultaneously. The orchestrator inspects the runtime environment + project config:

| Condition | Where to check | Required value |
|-----------|----------------|----------------|
| harness level is `thorough` | `.moai/config/sections/harness.yaml` `harness.level` | `thorough` |
| Agent Teams enabled in workflow | `.moai/config/sections/workflow.yaml` `workflow.team.enabled` | `true` |
| Experimental flag set in environment | env var `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` | `1` |

When any condition fails, the orchestrator falls through to Mode 4 evaluation. If a user explicitly requested `--mode team` and the prereqs are not met, the orchestrator emits the canonical sentinel `MODE_TEAM_UNAVAILABLE` (per `.claude/rules/moai/workflow/spec-workflow.md` § Mode Dispatch) and continues with the fallback mode plus a `[mode-auto-downgrade]` info log.

Anthropic Agent Teams documentation verbatim guidance applies: *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* The orchestrator MUST NOT spawn fewer than 3 nor more than 5 teammates in agent-team mode.

### §C.2 Mode 4 (Parallel) Compound preference

Mode 4 is preferred via the unified compound clause:

> `[Where the harness level is standard or thorough] [While the task scope is multi-domain (≥3 domains OR ≥10 files)] [When the orchestrator selects an execution mode in Phase 0.95]`, the orchestrator shall prefer Agent Teams mode if all Agent Teams capability-gate prerequisites are met; otherwise the orchestrator shall fall back to parallel multi-spawn of retained agents (maximum 3-5 concurrent `Agent()` calls in a single message per Anthropic verbatim "Start with 3-5 teammates").

The 3-5 ceiling applies equally to Mode 3 (agent-team teammates) and Mode 4 (concurrent `Agent()` spawn calls). Exceeding the ceiling regresses to coordination overhead and contradicts Anthropic's published guidance.

### §C.3 Mode 6 (Workflow) capability gate

Mode 6 (`workflow`) is candidate ONLY when ALL of the following preconditions hold. The orchestrator MUST verify each before launching a Workflow:

| Precondition | Why it is required |
|--------------|---------------------|
| GATE-2 already passed | Workflow agents cannot prompt the user mid-run (no mid-run user input). Therefore the one decision that MUST involve the user — the plan→run human gate — MUST already be cleared. A Mode 6 launch before GATE-2 passes is prohibited (§E anti-pattern). |
| All preferences collected | All user preferences (Tier, mode preference, PR strategy, etc.) MUST be drained at GATE-2 before launch, because the asymmetric boundary forbids both Workflow agents and `/goal`-turn agents from prompting the user (agent-common-protocol.md § User Interaction Boundary). |
| Scope ≥ ~30 files, mechanical, genuinely parallel | The Workflow primitive earns its overhead only on genuinely-parallel high-volume mechanical work; coding-heavy / multi-domain work stays Mode 5 (Anthropic's coding-task parallelism caveat). |
| Workflows available | `CLAUDE_CODE_DISABLE_WORKFLOWS` is not set AND runtime version ≥ v2.1.154; otherwise fall through to Mode 5. |
| Selection logged | The Mode 6 selection AND a confirmation that GATE-2 already passed AND that all preferences were collected MUST be recorded in `progress.md § Mode Selection` before the Workflow launches (§D). |

#### Mode 6 is scaling, not nesting

The Workflow is launched by the **orchestrator** (main session) as a scaling primitive. The Workflow script coordinates agents and keeps intermediate results in script variables; it returns only the final synthesis to the session context. This is NOT a subagent spawning a subagent — the flat hierarchy is preserved (Anthropic guidance: "Subagents cannot spawn other subagents"). The concurrency model (16 concurrent / 1000-total backstop) is the published cap of the Workflow primitive cited from `dynamic-workflows.md`, NOT a MoAI-invented API.

#### No named-script Workflow API

The official Claude Code documentation does not document a typed named-script Workflow API. This rule describes only the conceptual *coordinate-agents → intermediate results in script variables → final synthesis* model. No named Workflow-script function signatures — an `agent`-function, a `parallel`-function, a `pipeline`-function, or a `phase`-function — are asserted anywhere (§E anti-pattern; the asserted-API prohibition).

#### Mode 6 / `/goal` agents return blocker reports, never prompt the user

When a Mode 6 Workflow agent or a `/goal`-turn agent lacks a required input, that agent returns a structured blocker report; the orchestrator runs an `AskUserQuestion` round and re-delegates with the answers injected. Agents never prompt the user directly — this is the asymmetric orchestrator-subagent boundary (`.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary). The run-phase `/goal ac_converge` wiring point and its semantic-failure escalation live in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge).

---

## §D — Logging Contract (progress.md § Mode Selection)

Per the canonical mode-logging policy, the orchestrator MUST record its mode-selection decision in `.moai/specs/SPEC-{ID}/progress.md` under a `## §E — Phase 0.95 Mode Selection` (or analogous section name preserving the `Mode Selection` token for the grep acceptance criterion) before spawning the first run-phase `Agent()` call.

### §D.1 Required content

The Mode Selection section MUST include:

1. **Input parameters block** — values for tier, scope, domain count, file language mix, concurrency benefit, Agent Teams prereqs status
2. **Mode evaluation table** — for each of the 6 modes, a row stating "selected" or "not selected" and a one-line rationale
3. **Decision** — the chosen mode (one of: `trivial`, `background`, `agent-team`, `parallel`, `sub-agent`, `workflow`) on a single line for grep-friendly verification
4. **Justification** — a short paragraph (2-5 sentences) explaining why the chosen mode is preferable to alternatives, citing the relevant Anthropic finding(s) when applicable
5. **Mode 6 confirmation (when `workflow` is selected)** — an explicit line confirming GATE-2 already passed AND all preferences were collected before the Workflow launches (§C.3 logging precondition)

### §D.2 Token requirement (grep verification)

The orchestrator's mode logging is verified by the canonical grep acceptance criterion via:

```bash
grep -A 5 "Mode Selection" .moai/specs/SPEC-{ID}/progress.md \
  | grep -c -i "sequential\|parallel\|agent-team\|sub-agent\|trivial\|background"
```

The grep count MUST be ≥ 1. In practice, naming the chosen mode anywhere within 5 lines of the `Mode Selection` heading satisfies this; the structured `Decision: <mode>` line accomplishes this directly.

### §D.3 When to log a boundary case

When the decision tree hit a boundary (e.g., scope = exactly 10 files, exactly 3 domains, harness = `standard` with team.enabled = true but env var unset), the orchestrator MUST additionally include a **Boundary Case** subsection documenting the tie-breaker rule that resolved the ambiguity. This enables retrospective analysis to recalibrate threshold values across SPEC cohorts.

---

## §E — Anti-Patterns

The following patterns violate the orchestration mode selection contract:

- **Spawning Mode 3 (Agent Teams) when prereqs not met** — produces runtime `MODE_TEAM_UNAVAILABLE` sentinel; orchestrator MUST verify all three capability-gate conditions in §C.1 before selecting Mode 3
- **Spawning > 5 concurrent agents in Mode 3 or Mode 4** — exceeds Anthropic-recommended 3-5 ceiling and incurs coordination overhead
- **Selecting Mode 4 (Parallel) for coding-heavy work** — violates Anthropic's coding-task parallelism caveat; Mode 5 (Sub-Agent sequential) is the correct default for coding tasks
- **Selecting Mode 6 (Workflow) for coding-heavy / multi-domain / new-code work** — violates Anthropic's coding-task parallelism caveat; Mode 6 admits ONLY genuinely-parallel high-volume mechanical work (one uniform transform rule, no inter-file dependency). Coding-heavy work belongs to Mode 5
- **Launching a Mode 6 Workflow before GATE-2 has passed** — violates the GATE-2 mandatory-restoration policy; the orchestrator MUST NOT launch the Workflow before GATE-2 user approval and MUST return control to the GATE-2 `AskUserQuestion` gate. Mode 6 is strictly downstream of GATE-2
- **Asserting a typed/named Workflow script API** — a named `agent`-function, `parallel`-function, `pipeline`-function, or `phase`-function signature is NOT documented by Claude Code; describe the conceptual coordinate-agents → script-variable results → final-synthesis model instead (the named-script-API prohibition)
- **Selecting Mode 6 without recording the GATE-2-passed + preferences-collected confirmation in `progress.md`** — the §C.3 / §D.1 #5 logging precondition makes the autonomy decision auditable; skipping it leaves the Workflow launch unverifiable post-hoc
- **Skipping the progress.md logging step** — fails the canonical mode-logging acceptance criterion; the decision is no longer auditable post-hoc
- **Re-spawning the same mode for multiple consecutive milestones in Mode 5 without re-evaluating** — acceptable practice for a single SPEC, but when run-phase scope changes mid-flight (e.g., milestone scope-up via blocker report), the orchestrator SHOULD re-run Phase 0.95
- **Substituting an `AskUserQuestion` round for the autonomous decision** — Phase 0.95 is autonomous by contract; user intervention belongs to Phase 0.5 verdict review (when verdict is FAIL or INCONCLUSIVE) or GATE-2 (plan-to-implement HUMAN GATE), not Phase 0.95

---

## §F — Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification — `--mode` flag matrix and Mode Dispatch sentinels (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate — runs before Phase 0.95 and may produce `BYPASSED` / `INCONCLUSIVE` / `FAIL` verdicts that affect Phase 0.95 inputs
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S/M/L delegation template selection (interacts with Mode 5 sub-agent spawn prompts)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — sibling rule documenting the orchestrator's rejection behavior when a paste-ready resume references an archived-agent name (independent of mode selection)
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the Workflow (Mode 6) primitive: 16-concurrent / 1000-total cap, no-mid-run-user-input semantics, GATE-2-is-unaffected note, and the absence of a documented named-script API
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` autonomous-continuation semantics (the run-phase `ac_converge` condition wiring lives in `run.md` § Run-phase Autonomy (/goal ac_converge))
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge) — co-located GATE-2 ordering reference + `/goal ac_converge` set
- The canonical agent catalog design — design-time decision tree from which this rule was derived
- Anthropic Sub-agents and Agent Teams documentation — verbatim citations grounding the Mode 3 ceiling and Mode 4-vs-Mode-5 coding-task caveat
- Anthropic Agent Teams documentation — *"Start with 3-5 teammates for most workflows."*
- Anthropic multi-agent research engineering note — *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."*

---

## §G — Two-Axis Confusion Warning (Mode 6 wiring author)

[ZONE:Frozen] [HARD] "Mode 6 (workflow)" added here is appended to ONE specific list: the **Phase 0.95 execution-mode catalog** in this file (`trivial` / `background` / `agent-team` / `parallel` / `sub-agent` → + `workflow`). This is a DIFFERENT axis from the `run.md` `--mode` **dispatch axis** (`autopilot` / `loop` / `team` / `pipeline`) documented in `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification and the `run.md` Mode Dispatch table. Both happen to be short lists, which is the confusion trap.

- **Execution-mode catalog** (Phase 0.95, where Mode 6 lands): governs HOW the orchestrator spawns — concurrency + spawn surface (direct / background `Agent` / `TeamCreate` / parallel `Agent()` / sequential `Agent()` / **Workflow fan-out**).
- **`--mode` dispatch axis** (CLI flag, NOT touched here): governs WHICH `/moai run` workflow variant runs — `autopilot` vs `loop` (Ralph) vs `team` vs the rejected `pipeline`.

[ZONE:Frozen] [HARD] Mode 6 (`workflow`) is added ONLY to the execution-mode catalog. It is NOT a new `--mode` dispatch value; no `--mode workflow` flag is introduced; the `run.md` Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is unchanged. The header cross-reference above already notes the two axes "interact with — but are separate from" each other; this separation is preserved.

---

Version: 1.1.0
Origin: canonical agent catalog policy; Mode 6 (workflow) added by the run-phase autonomy line
Status: Active — applies to all `/moai run` Phase 0.95 invocations
