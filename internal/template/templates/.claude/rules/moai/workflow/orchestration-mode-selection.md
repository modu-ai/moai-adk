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

[ZONE:Frozen] [HARD] All Phase 0.95 execution modes are strictly downstream of Implementation Kickoff Approval (renamed from GATE-2) (the plan→run HUMAN GATE). The orchestrator reaches Phase 0.95 ONLY after Implementation Kickoff Approval user approval has already been obtained. Mode selection — including Mode 6 (workflow) — is never a substitute for Implementation Kickoff Approval and never a path that crosses the plan→run boundary ahead of the human gate. Implementation Kickoff Approval is mandatory and score-independent (a plan-auditor PASS or a high skip-eligible score never auto-bypasses it; skip-eligibility applies only to Phase 0.5 verdict re-execution, not to Implementation Kickoff Approval) per the Implementation Kickoff Approval mandatory-restoration policy.

> **Path B amendment (Intent-Gated Goal-Directed Autonomy — IGGDA)**: the Implementation Kickoff Approval gate above is AMENDED, not removed, by the safe-condition predicate documented in §H below. The `[ZONE:Frozen]` marker is PRESERVED as an audit trail; the amendment reduces the gate's per-run-blocking *framing* in safe domains (a lightweight confirmation with a clearly-recommended low-friction option) while preserving the user's veto authority and forcing explicit-gate in dangerous domains. The marker is never removed; only the gate's framing is adjusted. The user-facing `AskUserQuestion` round is ALWAYS issued and ALWAYS blocks until the user responds (auto-proceed does NOT skip it and does NOT auto-select — the user retains full veto). See §H.2 for the auto-select-after-timeout deferral note. See §H for the 4-condition compound predicate, §I for the IGGDA 4-phase pipeline it governs, and §J for the independent-audit preservation invariants.

> Cross-reference: `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification covers the `--mode` flag matrix (autopilot / loop / team / pipeline) which interacts with — but is separate from — the 6-mode catalog below. The run-phase `/goal ac_converge` autonomy wiring point lives in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge); `.claude/rules/moai/workflow/dynamic-workflows.md` is the source for the Workflow (Mode 6) primitive (16-concurrent / 1000-total cap) and the named-script-API prohibition.

---

## §A — Mode Catalog (6 modes)

The orchestrator selects exactly one of the following modes per Phase 0.95 invocation.

| # | Mode | Concurrency | Spawn Surface | When to prefer |
|---|------|-------------|---------------|----------------|
| 1 | `trivial` | None — direct execution by the orchestrator, no sub-agent spawn | n/a | Typo fix, single-line formatting, no semantic change |
| 2 | `background` | 1 concurrent sub-agent | `Agent(run_in_background: true, ...)` | Read-only analysis that can complete asynchronously without blocking the conversation |
| 3 | `agent-team` | 3-5 dynamic teammates | `Agent(subagent_type: "general-purpose", name: ...)` × N (implicit team forms on first spawn — no setup step; `team_name` accepted but ignored per Claude Code v2.1.178) | Multi-domain (≥3 domains OR ≥10 files) research-heavy work AND all Agent Teams capability-gate prerequisites met |
| 4 | `parallel` | 3-5 concurrent sub-agents (single message, multiple `Agent()` calls) | Multiple `Agent()` invocations in one assistant turn | Multi-domain research that does NOT meet Agent Teams prerequisites; or any case where Agent Teams session overhead exceeds benefit |
| 5 | `sub-agent` | 1 sequential sub-agent per milestone | Sequential `Agent(...)` spawns, one milestone at a time | Coding-heavy work (per Anthropic's coding-task parallelism caveat), or any case where the simpler mode suffices |
| 6 | `workflow` | Up to 16 concurrent workflow agents (1000-total per-run backstop, per `dynamic-workflows.md`) | Orchestrator-launched Workflow fan-out (a script the runtime executes to coordinate agents — NOT a subagent spawning subagents) | Genuinely-parallel, high-volume **mechanical** transformation (≥ ~30 files AND a single uniform transform rule AND no inter-file dependency) — call-site rename, import-path bulk change, signature-stable edits. Coding-heavy / multi-domain / new-code work stays Mode 5 (per Anthropic's coding-task parallelism caveat). |

Mode 5 is the **default fallback** when no other mode's selection criteria are unambiguously met. Mode 6 (`workflow`) is the narrow high-volume-mechanical exception, selectable ONLY after Implementation Kickoff Approval has passed (see §C.3).

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
  │   ├── YES → Mode 3: AGENT-TEAM (implicit team — dynamic teammate spawn via Agent(name=...))
  │   └── NO  → continue
  │
  ├── Is the task multi-domain (≥3 domains) AND research-heavy
  │   (NOT coding-heavy per Anthropic's coding-task parallelism caveat)?
  │   ├── YES → Mode 4: PARALLEL (3-5 concurrent Agent() in single message)
  │   └── NO  → continue
  │
  ├── Is the task ≥ ~30 files AND mechanical (one uniform transform rule)
  │   AND genuinely parallel — no inter-file dependency
  │   AND Implementation Kickoff Approval already passed AND all preferences already collected
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

The numeric auto-select thresholds — **≥ 3 domains, ≥ 10 files, or complexity score ≥ 7** — are stated here as the single prose SSOT; the machine-readable source is `.moai/config/sections/workflow.yaml` `auto_selection` (`min_domains_for_team`, `min_files_for_team`, `min_complexity_score`). Dispatch and skill surfaces cross-reference this section instead of restating the numbers inline.

### §B.1b Auto-mode pre-launch classifier (CC 2.1.178+)

When Claude Code runs in **auto mode** (per-tool auto-approval, paired with `/goal` for unattended loops), a pre-launch classifier evaluates each subagent spawn before it is dispatched — the classifier gates whether a spawn proceeds without a per-tool approval prompt. This is a platform-level mechanism that runs ahead of the Phase 0.95 mode-selection logic documented here; Phase 0.95 selects which mode the orchestrator uses to structure work, while the auto-mode classifier gates the per-spawn approval surface underneath that choice. The two are complementary: `/goal` (see `.claude/rules/moai/workflow/goal-directive.md`) removes per-turn STOP prompts, auto mode removes per-tool approval prompts, and Phase 0.95 mode selection decides HOW the orchestrator fans out. An active auto-mode classifier does NOT relax Implementation Kickoff Approval (the plan-to-implement human gate) — the human gate is decided before any run-phase work begins, and the classifier only governs per-spawn approval latency within an already-approved run.

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
| Implementation Kickoff Approval already passed | Workflow agents cannot prompt the user mid-run (no mid-run user input). Therefore the one decision that MUST involve the user — the plan→run human gate — MUST already be cleared. A Mode 6 launch before Implementation Kickoff Approval passes is prohibited (§E anti-pattern). |
| All preferences collected | All user preferences (Tier, mode preference, PR strategy, etc.) MUST be drained at Implementation Kickoff Approval before launch, because the asymmetric boundary forbids both Workflow agents and `/goal`-turn agents from prompting the user (agent-common-protocol.md § User Interaction Boundary). |
| Scope ≥ ~30 files, mechanical, genuinely parallel | The Workflow primitive earns its overhead only on genuinely-parallel high-volume mechanical work; coding-heavy / multi-domain work stays Mode 5 (Anthropic's coding-task parallelism caveat). |
| Workflows available | `CLAUDE_CODE_DISABLE_WORKFLOWS` is not set AND runtime version ≥ v2.1.154; otherwise fall through to Mode 5. |
| Selection logged | The Mode 6 selection AND a confirmation that Implementation Kickoff Approval already passed AND that all preferences were collected MUST be recorded in `progress.md § Mode Selection` before the Workflow launches (§D). |

#### Mode 6 is scaling, not nesting

The Workflow is launched by the **orchestrator** (main session) as a scaling primitive. The Workflow script coordinates agents and keeps intermediate results in script variables; it returns only the final synthesis to the session context. This is NOT a subagent spawning a subagent — the flat hierarchy is preserved (Anthropic guidance: "Subagents cannot spawn other subagents"). The concurrency model (16 concurrent / 1000-total backstop) is the published cap of the Workflow primitive cited from `dynamic-workflows.md`, NOT a MoAI-invented API.

#### No named-script Workflow API

The official Claude Code documentation does not document a typed named-script Workflow API. This rule describes only the conceptual *coordinate-agents → intermediate results in script variables → final synthesis* model. No named Workflow-script function signatures — an `agent`-function, a `parallel`-function, a `pipeline`-function, or a `phase`-function — are asserted anywhere (§E anti-pattern; the asserted-API prohibition).

#### Mode 6 / `/goal` agents return blocker reports, never prompt the user

When a Mode 6 Workflow agent or a `/goal`-turn agent lacks a required input, that agent returns a structured blocker report; the orchestrator runs an `AskUserQuestion` round and re-delegates with the answers injected. Agents never prompt the user directly — this is the asymmetric orchestrator-subagent boundary (`.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary). The run-phase `/goal ac_converge` wiring point and its semantic-failure escalation live in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge).

---

## §D — Logging Contract (progress.md § Mode Selection)

Per the canonical mode-logging policy, the orchestrator MUST record its mode-selection decision in `.moai/specs/SPEC-{ID}/progress.md` under a `## §F Phase 0.95 Mode Selection` section (preserving the `Mode Selection` token for the grep acceptance criterion) before spawning the first run-phase `Agent()` call. The `§F` letter is allocated by the canonical progress.md Section Map in `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map — Mode Selection MUST NOT reuse `§E` (the `§E.*` namespace is reserved for the era.go-parsed lifecycle-phase structure; overloading it would collide with era classification).

### §D.1 Required content

The Mode Selection section MUST include:

1. **Input parameters block** — values for tier, scope, domain count, file language mix, concurrency benefit, Agent Teams prereqs status
2. **Mode evaluation table** — for each of the 6 modes, a row stating "selected" or "not selected" and a one-line rationale
3. **Decision** — the chosen mode (one of: `trivial`, `background`, `agent-team`, `parallel`, `sub-agent`, `workflow`) on a single line for grep-friendly verification
4. **Justification** — a short paragraph (2-5 sentences) explaining why the chosen mode is preferable to alternatives, citing the relevant Anthropic finding(s) when applicable
5. **Mode 6 confirmation (when `workflow` is selected)** — an explicit line confirming Implementation Kickoff Approval already passed AND all preferences were collected before the Workflow launches (§C.3 logging precondition)

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
- **Launching a Mode 6 Workflow before Implementation Kickoff Approval has passed** — violates the Implementation Kickoff Approval mandatory-restoration policy; the orchestrator MUST NOT launch the Workflow before Implementation Kickoff Approval user approval and MUST return control to the Implementation Kickoff Approval `AskUserQuestion` gate. Mode 6 is strictly downstream of Implementation Kickoff Approval
- **Asserting a typed/named Workflow script API** — a named `agent`-function, `parallel`-function, `pipeline`-function, or `phase`-function signature is NOT documented by Claude Code; describe the conceptual coordinate-agents → script-variable results → final-synthesis model instead (the named-script-API prohibition)
- **Selecting Mode 6 without recording the Implementation Kickoff Approval-passed + preferences-collected confirmation in `progress.md`** — the §C.3 / §D.1 #5 logging precondition makes the autonomy decision auditable; skipping it leaves the Workflow launch unverifiable post-hoc
- **Skipping the progress.md logging step** — fails the canonical mode-logging acceptance criterion; the decision is no longer auditable post-hoc
- **Re-spawning the same mode for multiple consecutive milestones in Mode 5 without re-evaluating** — acceptable practice for a single SPEC, but when run-phase scope changes mid-flight (e.g., milestone scope-up via blocker report), the orchestrator SHOULD re-run Phase 0.95
- **Substituting an `AskUserQuestion` round for the autonomous decision** — Phase 0.95 is autonomous by contract; user intervention belongs to Phase 0.5 verdict review (when verdict is FAIL or INCONCLUSIVE) or Implementation Kickoff Approval (plan-to-implement HUMAN GATE), not Phase 0.95

---

## §F — Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification — `--mode` flag matrix and Mode Dispatch sentinels (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate — runs before Phase 0.95 and may produce `BYPASSED` / `INCONCLUSIVE` / `FAIL` verdicts that affect Phase 0.95 inputs
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S/M/L delegation template selection (interacts with Mode 5 sub-agent spawn prompts)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — sibling rule documenting the orchestrator's rejection behavior when a paste-ready resume references an archived-agent name (independent of mode selection)
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the Workflow (Mode 6) primitive: 16-concurrent / 1000-total cap, no-mid-run-user-input semantics, Implementation Kickoff Approval-is-unaffected note, and the absence of a documented named-script API
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` autonomous-continuation semantics (the run-phase `ac_converge` condition wiring lives in `run.md` § Run-phase Autonomy (/goal ac_converge))
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (/goal ac_converge) — co-located Implementation Kickoff Approval ordering reference + `/goal ac_converge` set
- The canonical agent catalog design — design-time decision tree from which this rule was derived
- Anthropic Sub-agents and Agent Teams documentation — verbatim citations grounding the Mode 3 ceiling and Mode 4-vs-Mode-5 coding-task caveat
- Anthropic Agent Teams documentation — *"Start with 3-5 teammates for most workflows."*
- Anthropic multi-agent research engineering note — *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."*

---

## §G — Two-Axis Confusion Warning (Mode 6 wiring author)

[ZONE:Frozen] [HARD] "Mode 6 (workflow)" added here is appended to ONE specific list: the **Phase 0.95 execution-mode catalog** in this file (`trivial` / `background` / `agent-team` / `parallel` / `sub-agent` → + `workflow`). This is a DIFFERENT axis from the `run.md` `--mode` **dispatch axis** (`autopilot` / `loop` / `team` / `pipeline`) documented in `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification and the `run.md` Mode Dispatch table. Both happen to be short lists, which is the confusion trap.

- **Execution-mode catalog** (Phase 0.95, where Mode 6 lands): governs HOW the orchestrator spawns — concurrency + spawn surface (direct / background `Agent` / implicit-team `Agent(name=...)` / parallel `Agent()` / sequential `Agent()` / **Workflow fan-out**).
- **`--mode` dispatch axis** (CLI flag, NOT touched here): governs WHICH `/moai run` workflow variant runs — `autopilot` vs `loop` (Ralph) vs `team` vs the rejected `pipeline`.

[ZONE:Frozen] [HARD] Mode 6 (`workflow`) is added ONLY to the execution-mode catalog. It is NOT a new `--mode` dispatch value; no `--mode workflow` flag is introduced; the `run.md` Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is unchanged. The header cross-reference above already notes the two axes "interact with — but are separate from" each other; this separation is preserved.

---

## §G.1 — Dispatch-Axis Crosswalk (`--mode` values + scale-table labels → catalog modes)

> **Correspondence, not merge.** This crosswalk documents how each `--mode` dispatch-axis value and each Phase 0.95 scale-table label CORRESPONDS to a catalog mode. It does NOT merge the two axes: per §G they remain separate, the `--mode` value set `{autopilot, loop, team, pipeline}` is unchanged, no `--mode workflow` value is introduced, and the Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is untouched.

### `--mode` dispatch-axis values → catalog modes

| `--mode` value | Corresponds to catalog mode | Notes |
|----------------|------------------------------|-------|
| `autopilot` | Mode 5 (`sub-agent`) | Default single-lead orchestration; the Phase 0.95 scale-based selection chooses the envelope (see scale-label rows below). |
| `loop` | Mode 5 (`sub-agent`) | Ralph-engine diagnostic fix-loop variant — sequential per-iteration delegation. The granularity differs (diagnostics, not phases) but the spawn shape is the Mode 5 sequential sub-agent. |
| `team` | Mode 3 (`agent-team`) | Resolves through the §C.1 capability gate; when the gate fails, an explicit `--mode team` emits `MODE_TEAM_UNAVAILABLE` and falls back per the Mode Resolver. |
| `pipeline` | Mode 5 (`sub-agent`) — utility subcommands only | Rejected on multi-agent subcommands (`MODE_PIPELINE_ONLY_UTILITY`); the utility subcommands are intrinsically pipeline-class, so `pipeline` names their fixed sequential direct / sub-agent execution shape — the `--mode pipeline` flag itself is ignored there (`MODE_FLAG_IGNORED_FOR_UTILITY`), not honored. |

### Phase 0.95 scale-table labels → catalog modes

| Scale label | Corresponds to catalog mode | Envelope |
|-------------|------------------------------|----------|
| Fix | Mode 5 (`sub-agent`) | Minimal envelope — single implementation agent + orchestrator verification batch. |
| Focused | Mode 5 (`sub-agent`) | Focused envelope — single implementation agent with domain context injected. |
| Standard | Mode 5 (`sub-agent`) | Standard envelope — planning + implementation + audit, sequential. |
| Full Pipeline | Mode 5 (`sub-agent`) | Full envelope — full sequential agent chain (plan → implement → audit → docs). |
| Team | Mode 3 (`agent-team`) | Resolves through the SAME §C.1 capability gate as auto-select — a forced `--team` flag and the flag-free auto-select path both pass the gate; the flag never bypasses it. |

Every `--mode` value and every scale label corresponds to exactly one catalog mode. Mode 1 (trivial), Mode 2 (background), Mode 4 (parallel), and Mode 6 (workflow) have NO dispatch-axis or scale-label counterpart — they are selectable only via the Phase 0.95 decision tree (§B). This asymmetry is expected and is further evidence the two axes are distinct.

---

## §H — IGGDA Safe-Condition Predicate (Path B Implementation Kickoff Amendment)

Intent-Gated Goal-Directed Autonomy (IGGDA) AMENDS the Implementation Kickoff Approval gate from a per-run-blocking `AskUserQuestion` (Path A — what AUTONOMY-RUN-GOAL-001 preserved) into a **safe-condition predicate** with two branches. The gate is NOT removed; its weight is redistributed. The `[ZONE:Frozen]` marker on the header paragraph above is PRESERVED as the audit trail for the original invariant; the amendment below is the behavioral change.

### §H.1 The compound predicate (4 conditions, all-must-hold for auto-proceed)

[ZONE:Evolvable] [HARD] **Where** the Socratic interview has reached 100% intent clarity, **While** the plan-auditor has issued a PASS verdict, **When** the orchestrator reaches the plan→run boundary, the Implementation Kickoff Approval gate shall evaluate the following four-condition compound predicate:

| # | Condition | FAIL verdict |
|---|-----------|--------------|
| (a) | **Intent clarity 100%** — the Socratic interview (multi-round, max 4 questions per round) is complete; every assumption surfaced; every ambiguity resolved per `.claude/rules/moai/core/askuser-protocol.md` § Socratic Interview Structure | return-to-Phase-0 (the interview must complete first) |
| (b) | **plan-auditor PASS** — the independent plan-audit verdict (Phase 0.5) is PASS (NOT FAIL, NOT INCONCLUSIVE) | surface-to-user (the independent audit halt) |
| (c) | **Tier S or Tier M** — the SPEC tier is S or M (NOT Tier L — Tier L is the complexity cutoff for "dangerous") | explicit-gate (mandatory blocking gate) |
| (d) | **No dangerous keywords AND no destructive scope** — the SPEC scope contains none of the security / payment / critical keywords enumerated in §H.3 below AND the `--pr` flag is absent AND the scope is not marked destructive | explicit-gate (mandatory blocking gate) |

The predicate is evaluated in order (a)→(b)→(c)→(d). The FIRST failing condition determines the verdict rationale (logged for audit), but the gate decision collapses to one of: **auto-proceed** (all 4 hold) / **explicit-gate** (c or d fails) / **return-to-Phase-0** (a fails) / **surface-to-user** (b fails).

### §H.2 The two branches

| Branch | Trigger | Behavior |
|--------|---------|----------|
| **Auto-proceed** | All 4 conditions hold | The `AskUserQuestion` round is STILL ISSUED (the user retains veto authority) and BLOCKS until the user responds. The gate is REDUCED in *framing* to a lightweight confirmation (the `(Recommended)` first option signals the low-friction path) but is NOT removed and does NOT auto-select. |
| **Explicit-gate** | Condition (c) OR (d) fails | The `AskUserQuestion` round REMAINS mandatory and blocks until the user responds. This is the Path A behavior preserved verbatim for dangerous domains. |

> **Auto-select-after-timeout is DEFERRED — not currently implementable.** [ZONE:Evolvable] An earlier design described the auto-proceed branch as auto-selecting the first `(Recommended)` option after a bounded 30-second timeout (`workflow.iggda.kickoff_timeout_seconds`). `AskUserQuestion` has **no timeout / auto-select mechanism** in Claude Code — a question always blocks until the user answers. The timeout auto-proceed is therefore **documentation-only, deferred, and NOT mechanically implementable via `AskUserQuestion`** (per the runtime-recovery-doctrine deferred-labeling convention; forward-link: a future runtime-layer SPEC would be required if a genuine bounded-timeout gate is ever wanted). Until then, the **canonical Implementation Kickoff Approval gate is the score-independent BLOCKING `AskUserQuestion` gate** — auto-proceed reduces the gate's *framing* (a recommended low-friction option) but never its *blocking* nature. The `run.md` skill owns the concrete skill-side gate text. The §H concept (the 4-condition safe-condition predicate and the two branches) is PRESERVED; only the auto-select-after-timeout mechanic is labeled deferred.

**User veto is never overridden.** Even in the auto-proceed branch the user CAN select "abort" or "further review" — the question blocks until they answer, so there is no timeout window that could elapse without a response. The predicate does NOT override user intent — it only reduces the *friction* of the gate in safe domains (a clearly-recommended low-friction option) where the Socratic interview has already drained intent.

### §H.3 The dangerous-domain keyword list (condition (d) determinator)

The keyword list is enumerable at design time and extensible via rule-file edit (not code change). It is intentionally OVER-INCLUSIVE in the initial release — better to force explicit-gate on a false-positive than auto-proceed on a false-negative. The orchestrator matches keywords case-insensitively against the SPEC's title, scope summary, plan.md §A context, and acceptance.md.

**Security domain** (any match → condition (d) FAILS):
- `auth`, `authentication`, `authorization`, `acl`
- `secret`, `credential`, `password`, `token`, `api_key`, `apikey`
- `crypto`, `encryption`, `decrypt`, `hash`, `salt`
- `session`, `cookie`, `jwt`, `oauth`, `saml`, `sso`
- `injection`, `xss`, `csrf`, `sqli`, `rce`
- `vulnerability`, `cve`, `owasp`

**Payment domain** (any match → condition (d) FAILS):
- `payment`, `billing`, `invoice`, `charge`
- `stripe`, `paypal`, `portone`, `iamport`, `toss`, `kakopay`, `naverpay`
- `card`, `pan`, `pci`, `dss`
- `refund`, `settlement`

**Critical infrastructure domain** (any match → condition (d) FAILS):
- `production`, `prod`, `live`
- `migration`, `rollback`, `drop_table`, `drop table`
- `force_push`, `force-push`
- `rm_rf`, `rm -rf`
- `database`, `schema` (when paired with `drop` / `alter`)

**Destructive scope markers** (any match → condition (d) FAILS):
- `--pr` flag supplied by the user
- SPEC scope marked destructive (PR creation / force-push / table-drop / external-post semantics)

**Maintenance**: this list lives in this rule file and is extended via SPEC amendment. When a new dangerous domain is identified, add it here; no code change is required (the orchestrator reads the list from this rule at gate-evaluation time).

### §H.4 Decision logging (auditable)

[ZONE:Evolvable] [HARD] Every safe-condition predicate evaluation (either branch) MUST be logged to `.moai/specs/<SPEC-ID>/progress.md` under a `## §G IGGDA Kickoff Predicate` section with: condition (a) value, condition (b) value, condition (c) value, condition (d) value + any matched keyword, the final verdict (auto-proceed / explicit-gate / return-to-Phase-0 / surface-to-user), and a timestamp. This makes every auto-proceed post-hoc reviewable; a miscaptured intent that slipped through is detectable. The `§G` letter is allocated by the canonical progress.md Section Map in `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map — the IGGDA Kickoff Predicate MUST NOT reuse `§E` (reserved for the era.go-parsed lifecycle structure).

### §H.5 Why this preserves the FROZEN-spirit (intent-verification-before-code)

The original Implementation Kickoff Approval invariant protects **intent verification before code is written** — a plan-auditor PASS score is NOT sufficient evidence of intent fidelity. Path B preserves this protection because:

1. **Condition (a) requires 100% Socratic clarity** — a multi-round interview (far more rigorous than a single plan→run boundary confirmation) IS intent verification, and is STRONGER than the single gate.
2. **The `AskUserQuestion` round is STILL ISSUED in the auto-proceed branch** — the user sees the confirmation and CAN veto; the question blocks until the user responds (see §H.2 for the auto-select-after-timeout deferral note). The gate is REDUCED in framing, not removed.
3. **Dangerous domains are carved out** (conditions (c) + (d)) — where the stakes are high, the gate remains fully mandatory (explicit-gate branch).
4. **The decision is auditable** (§H.4) — every auto-proceed is logged with all 4 conditions, so post-hoc verification is possible.

The amendment moves the human checkpoint's WEIGHT from a per-run-blocking gate to an intent-collection-time gate (Phase 0 Socratic), EXCEPT in dangerous domains where the per-run-blocking gate remains.

---

## §I — IGGDA 4-Phase Pipeline (the autonomous plan→run→sync chain)

IGGDA chains autonomy across all four phases of the pipeline. The user concentrates involvement in Phase 0 (Socratic intent collection); the rest is autonomous, with independent audits at Phase 1 and Phase 3 and a bounded recursive self-diagnosis loop at Phase 2.

### §I.1 The four phases (in execution order)

| Phase | Name | Autonomy | Human involvement |
|-------|------|----------|-------------------|
| **0** | **Intent** | Human-in-the-loop (Socratic) | Drains ALL preferences (Tier, mode, PR strategy, domain). ABSORBS the gating weight of Implementation Kickoff Approval under Path B (the human checkpoint moves HERE, except in dangerous domains where it remains at the plan→run boundary). |
| **1** | **Plan** | Autonomous | manager-spec produces SPEC artifacts → plan-auditor INDEPENDENT audit (fresh context). PASS → auto-advance to Phase 2. FAIL/INCONCLUSIVE → halt, surface to user. |
| **2** | **Run** | Autonomous + recursive self-diagnosis | manager-develop implements (DDD/TDD/autofix cycle_type). Bounded recursive self-diagnosis loop on mechanical failures (DIAGNOSE-PATCH-VERIFY, max 3 iterations). Semantic failures escalate immediately. All blocking ACs PASS + go test exit 0 + no out-of-scope modification → auto-advance to Phase 3. |
| **3** | **Sync + final independent audit** | Autonomous | manager-docs: CHANGELOG/README/frontmatter transitions → sync-auditor INDEPENDENT 4-dimension score (fresh context) → `moai spec audit` deterministic final check (0 MUST-FIX required). sync-auditor ≥ threshold + 0 MUST-FIX + git clean → IGGDA-complete (`/goal` clears). |

No phase is skipped except by graceful degradation (when `/goal` is unavailable per `.claude/rules/moai/workflow/goal-directive.md` — the pipeline degrades to per-turn progression rather than failing).

### §I.2 Phase ordering invariant

[ZONE:Evolvable] [HARD] The four phases execute in strict order: Phase 0 (Intent) → Phase 1 (Plan) → Phase 2 (Run) → Phase 3 (Sync + audit). Phase 0.5 (plan-audit) runs WITHIN Phase 1 (it is the Phase 1 exit gate, not a separate phase). The safe-condition predicate (§H) runs at the Phase 1 → Phase 2 boundary, AFTER Phase 0.5 PASS. The predicate does NOT substitute for Phase 0.5 — it is a downstream reduction of the Phase 1→2 gate weight, evaluated only after Phase 0.5 has already passed.

### §I.3 The auto-advance mechanism (moai-aware Stop hook)

The Stop hook driver `iggda-phase-driver.sh` fires at turn-end, reads `progress.md` + invokes `moai spec audit --json --filter-spec=<SPEC-ID>`, and emits a `/goal`-style auto-advance signal when the current phase's safe-transition predicate holds. The user does NOT author a `/goal` condition string; the goal is derived from Socratic intent and auto-converted to the phase-transition signal by the driver. See `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 for the Recovery-Signal Carve-Out (the driver exits 0 on recovery turns, never blocking a recovery).

**Installation status (honest reference).** The driver ships via the template tree and is installed into a project's `.claude/hooks/moai/` on `moai init` / `moai update`. It is therefore present and active in any project that has run `moai update` since the driver shipped, and absent in a checkout that has not yet synced. The driver is neither phantom (it exists and ships) nor unconditionally active in every session (a given checkout may not have it installed yet); its presence depends on whether the project has synced the template. When it is absent, the auto-advance degrades gracefully to per-turn progression (§I.1: "No phase is skipped except by graceful degradation").


---

## §J — IGGDA Independent-Audit Preservation

IGGDA's autonomous execution does NOT collapse the independent auditors (plan-auditor, sync-auditor) into the implementer. The independent-audit guarantee (bias prevention via fresh-context skeptical evaluation) is preserved verbatim.

### §J.1 Fresh-context spawn (NOT continuation of implementer turn)

[ZONE:Evolvable] [HARD] The plan-auditor (Phase 1) and sync-auditor (Phase 3) are spawned via `Agent(subagent_type: "plan-auditor")` and `Agent(subagent_type: "sync-auditor")` respectively, each in a FRESH isolated context. They are NOT continuations of the implementer's turn (manager-spec / manager-docs). The implementer's context carries the assumptions that produced the implementation; the auditor's fresh context catches what the implementer rationalized away. This is the "skeptical evaluation stance" in `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance.

### §J.2 FAIL/INCONCLUSIVE verdict halts auto-advance

[ZONE:Evolvable] [HARD] When plan-auditor OR sync-auditor returns FAIL or INCONCLUSIVE, the Stop hook driver halts auto-advance (no transition to the next phase). The orchestrator surfaces the verdict to the user via `AskUserQuestion`. The FAIL/INCONCLUSIVE is a hard stop regardless of prior phase PASS — a Phase 1 PASS does not authorize proceeding past a Phase 3 sync-auditor FAIL.

### §J.3 "Self-audit" vs "Independent audit" disambiguation

The IGGDA documentation uses both terms; they are COMPLEMENTARY, NOT interchangeable:

- **"Self-audit"** = the Phase 2 bounded recursive self-diagnosis loop (D3) performing first-party verification of MECHANICAL failures. It is a code-quality loop, NOT a SPEC-quality audit. No bias-prevention guarantee (it is the implementer checking their own mechanical work — acceptable for lint/type/import errors).
- **"Independent audit"** = plan-auditor (Phase 1) + sync-auditor (Phase 3) in fresh contexts. These are the SPEC-quality audits with bias-prevention guarantees.

IGGDA does NOT trade one for the other. Self-audit handles mechanical code failures fast (bounded loop, no human in the loop for easy cases); independent audit handles SPEC-quality assurance (human-grade skeptical evaluation in a fresh context).

---

Version: 1.2.0 (IGGDA: §H safe-condition predicate + §I 4-phase pipeline + §J independent-audit preservation added — Path B Implementation Kickoff Amendment)
Origin: derived from the canonical agent catalog and IGGDA policies.
Status: Active — applies to all `/moai run` Phase 0.95 invocations
