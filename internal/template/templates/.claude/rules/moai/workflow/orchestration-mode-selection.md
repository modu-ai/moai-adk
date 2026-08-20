---
description: "Phase 4 Mode Selection — 4-mode autonomous decision tree for the MoAI orchestrator, ordered by concurrent spawn count (direct / serial / fanout / sweep). Read at every /moai run entry."
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/plan.md,.claude/rules/moai/workflow/spec-workflow.md"
metadata:
  version: "2.0.0"
  status: "active"
  tags: "orchestration, mode-selection, agent-teams, workflow, phase-0.95"
---

# Orchestration Mode Selection — Phase 4

Canonical 4-mode autonomous decision tree for the MoAI orchestrator. Activated at Phase 4 (after Phase 1 plan-auditor verdict, before Phase 1 implementation). The decision is autonomous (no `AskUserQuestion` round); the chosen mode and the selection rationale are logged to `progress.md § Mode Selection`.

> **Rename note (single-axis taxonomy).** The catalog was renamed from a 6-entry numbered list onto ONE axis: the number of concurrently live spawns. The former names mixed two axes — "parallel" vs "sub-agent" both described spawn count while naming concurrency and executor separately. Legacy mapping for old logs and handoff bodies: `trivial`→`direct`, `background`→execution **option** (subagents run background by default since CC v2.1.198; no longer a mode), `agent-team`→unchanged (footnote below), `parallel`→`fanout`, `sub-agent`→`serial`, `workflow`→`sweep`. Handoff `mode:` enum tokens map likewise: `solo-sequential`→`serial`, `parallel-subagents`→`fanout`, `dynamic-workflow`→`sweep`. Legacy tokens remain READ-compatible indefinitely (old progress.md logs and pasted handoff bodies keep resolving); new writes use the new tokens only.

> **Progression-mode axis co-located with Kickoff mandatory-restoration**: the Implementation Kickoff Approval gate (which this rule keeps mandatory and score-independent) also offers an autonomous-vs-semi-autonomous progression-mode choice — a post-approval progression selection, never a relaxation of the gate. The gate remains mandatory in both modes; the axis selects only what happens after the gate passes. See `.claude/skills/moai/workflows/goal.md` § Progression Mode.

[ZONE:Frozen] [HARD] All Phase 4 execution modes are strictly downstream of Implementation Kickoff Approval (renamed from GATE-2) (the plan→run HUMAN GATE). The orchestrator reaches Phase 4 ONLY after Implementation Kickoff Approval user approval has already been obtained. Mode selection — including `sweep` — is never a substitute for Implementation Kickoff Approval and never a path that crosses the plan→run boundary ahead of the human gate. Implementation Kickoff Approval is mandatory and score-independent (a plan-auditor PASS or a high skip-eligible score never auto-bypasses it; skip-eligibility applies only to Phase 1 verdict re-execution, not to Implementation Kickoff Approval) per the Implementation Kickoff Approval mandatory-restoration policy.

> Cross-reference: `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification covers the `--mode` flag matrix (autopilot / loop / team / pipeline) which interacts with — but is separate from — the 4-mode catalog below. The run-phase `ac_converge` autonomy wiring point lives in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (ac_converge); `.claude/rules/moai/workflow/dynamic-workflows.md` is the source for the sweep primitive (16-concurrent / 1000-total cap) and the named-script-API prohibition.

---

## §A — Mode Catalog (4 modes)

The orchestrator selects exactly one of the following modes per Phase 4 invocation, ordered by concurrent spawn count:

| Mode | Concurrent spawns | Spawn surface | When to prefer |
|------|-------------------|---------------|----------------|
| `direct` | 0 — the orchestrator executes itself, no sub-agent spawn | n/a | Typo fix, single-line formatting, no semantic change |
| `serial` | 1 at a time (one sub-agent per milestone) | Sequential `Agent(...)` spawns | **DEFAULT fallback** — coding-heavy work (per Anthropic's coding-task parallelism caveat), or any case where the simpler mode suffices |
| `fanout` | N concurrent (3-5 ADVISORY band; hard bound: runtime subagent cap, default 20 — §C.2) | Multiple `Agent()` invocations in one assistant turn | Multi-domain research that does NOT meet Agent Teams prerequisites; or any case where Agent Teams session overhead exceeds benefit |
| `sweep` | Dozens–hundreds (up to 16 concurrent workflow agents, 1000-total per-run backstop, per `dynamic-workflows.md`) | Orchestrator-launched Workflow fan-out (a script the runtime executes to coordinate agents — NOT a subagent spawning subagents) | Genuinely-parallel, high-volume **mechanical** transformation (≥ ~30 files AND a single uniform transform rule AND no inter-file dependency) — call-site rename, import-path bulk change, signature-stable edits. Coding-heavy / multi-domain / new-code work stays `serial` (per Anthropic's coding-task parallelism caveat). |

**`background` is an execution option, not a mode.** Since Claude Code v2.1.198, subagents run in the background by default (the runtime chooses foreground only when it needs the result); any spawn-based mode may run foreground or background without changing which mode it is.

> **Footnote — Agent Teams (`agent-team`, experimental, re-allowed).** Not a catalog mode and never auto-selected: an explicit operator request (`--team` / `--mode team` / `Team` scale label) selects the native teammate runtime (flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` ships on). Tier L coordination auto-routing still targets `manager-lead`; multi-domain research routes to `fanout`; coding-heavy work to `serial`; mechanical bulk to `sweep`. Constraints and genealogy: §C.1.

`serial` is the **default fallback** when no other mode's selection criteria are unambiguously met. `sweep` is the narrow high-volume-mechanical exception, selectable ONLY after Implementation Kickoff Approval has passed (see §C.3).

---

## §B — Decision Tree

```
START (Phase 4 Mode Selection)
  │
  ├── Is task trivial (typo, single-line, no semantic change)?
  │   ├── YES → direct (orchestrator executes, no Agent() spawn)
  │   └── NO  → continue
  │
  ├── (agent-team — experimental, re-allowed. Entered ONLY by explicit
  │    operator request (--team / --mode team / Team scale label); the tree never
  │    auto-selects it. Unrequested multi-domain work falls through to the fanout
  │    check below. See §C.1 for constraints and genealogy.)
  │
  ├── Is the task multi-domain (≥3 domains) AND research-heavy
  │   (NOT coding-heavy per Anthropic's coding-task parallelism caveat)?
  │   ├── YES → fanout (3-5-advisory concurrent Agent() in single message; hard bound is the runtime subagent cap — §C.2)
  │   └── NO  → continue
  │
  ├── Is the task ≥ ~30 files AND mechanical (one uniform transform rule)
  │   AND genuinely parallel — no inter-file dependency
  │   AND Implementation Kickoff Approval already passed AND all preferences already collected
  │   AND Workflows available (not disabled, runtime version ≥ v2.1.154)?
  │   ├── YES → sweep (orchestrator-launched fan-out, scaling NOT nesting)
  │   └── NO  → continue
  │
  └── Default → serial (single Agent() sequential spawn per milestone)
```

`sweep` is checked AFTER `fanout` and BEFORE the `serial` default fallback. Coding-heavy or multi-domain or new-code work that reaches this branch falls through to `serial` — `sweep` admits ONLY the genuinely-parallel high-volume mechanical case (per Anthropic's coding-task parallelism caveat: most coding tasks involve fewer truly parallelizable tasks than research, so the sequential path is the safe default for coding work).

### §B.1 Input parameters

The orchestrator collects the following signals before traversing the decision tree:

- **tier**: SPEC tier (S / M / L) per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- **scope (file count)**: estimated count of files in the SPEC's run-phase scope
- **domain count**: number of distinct domains touched (agents, workflow skills, rules, hook scripts, template mirrors, Go source code, SPEC artifacts, etc.)
- **file language mix**: e.g., 100% markdown vs Go source vs shell vs mixed
- **concurrency benefit**: HIGH for research-heavy work (parallel reads, independent perspectives); LOW for coding-heavy work (Anthropic's coding-task parallelism caveat — most coding tasks involve fewer truly parallelizable tasks than research)

The numeric auto-select thresholds — **≥ 3 domains, ≥ 10 files, or complexity score ≥ 7** — are the single **prose SSOT** for `fanout` auto-selection (the former machine-readable `workflow.yaml` team-config block was removed with the Agent Teams static layer). Dispatch and skill surfaces cross-reference this section instead of restating the numbers inline.

### §B.1b Auto-mode pre-launch classifier (CC 2.1.178+)

When Claude Code runs in **auto mode** (per-tool auto-approval, paired with an armed `/moai goal` for unattended loops), a pre-launch classifier evaluates each subagent spawn before it is dispatched — the classifier gates whether a spawn proceeds without a per-tool approval prompt. This is a platform-level mechanism that runs ahead of the Phase 4 mode-selection logic documented here; Phase 4 selects which mode the orchestrator uses to structure work, while the auto-mode classifier gates the per-spawn approval surface underneath that choice. The two are complementary: an armed `/moai goal` (see `.claude/rules/moai/workflow/goal-directive.md`) removes per-turn STOP prompts, auto mode removes per-tool approval prompts, and Phase 4 mode selection decides HOW the orchestrator fans out. An active auto-mode classifier does NOT relax Implementation Kickoff Approval (the plan-to-implement human gate) — the human gate is decided before any run-phase work begins, and the classifier only governs per-spawn approval latency within an already-approved run.

### §B.2 Tie-breaker rules (boundary cases)

Phase 4 boundary cases (scope at threshold ±1, ambiguous domain count, etc.) follow these defaults:

- At threshold ±1 (9 vs 10 files; 2 vs 3 domains): default to the **simpler** mode (serial over fanout; fewer spawns over more)
- **Coding-heavy + multi-domain**: prefer `serial` over `fanout` (Anthropic's coding-task parallelism caveat)
- **Markdown-heavy + multi-domain + research-heavy**: prefer `fanout` (parallel multi-spawn)
- **`sweep` soft `~30`-file boundary**: the `≥ ~30 files` sweep entry threshold is tilde-prefixed (soft). At the boundary (exactly 30 files), the "default to the simpler mode" rule resolves toward `serial` — the tilde avoids a hard cliff and keeps even large work on the safer sequential path unless the transformation is genuinely mechanical-uniform.
- **`sweep` vs `serial` (transformation kind)**: even at high file counts, if the work is semantic / new-code / multi-rule, prefer `serial`; `sweep` admits ONLY a single uniform mechanical transform rule with no inter-file dependency.
- **Workflows disabled or unavailable**: when `CLAUDE_CODE_DISABLE_WORKFLOWS=1` is set OR the runtime version is below v2.1.154, `sweep` is not selectable and the task falls through to `serial` (cannot assume Workflow availability).
- **Tier L + markdown / shell-script-only scope**: `serial` with Tier L Section A-E delegation template (per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability)
- **Tier S + minimal scope**: `serial` with the minimal delegation form (~500-800 tokens, Section B may be filtered)

---

## §C — Capability Gates

### §C.1 Agent Teams (`agent-team`) — footnote surface (experimental, re-allowed)

**`agent-team` — re-allowed as experimental** (operator decision). The flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` ships enabled in `.claude/settings.json` and the distributed template, making the native teammate runtime a sanctioned orchestration surface: spawn teammates with the Agent tool's `name` parameter (the team forms implicitly on first spawn — one team per session), shared TaskList coordination, `moai cg` GLM teammate panes, `moai cc -w <name> --spawn` teammate windows, `~/.claude/teams/` registry.

**Selection rule unchanged**: the Phase 4 decision tree never auto-selects `agent-team` — an explicit operator request (`--team` / `--mode team` / `Team` scale label) selects it; Tier L coordination auto-routing still targets `manager-lead`; multi-domain research routes to `fanout`; coding-heavy work to `serial`; high-volume mechanical transformation to `sweep`.

**Genealogy**: `agent-team` was previously a numbered catalog entry, then retired (tombstone; a forced `--team` emitted the canonical sentinel `MODE_TEAM_UNAVAILABLE` per `spec-workflow.md` § Mode Dispatch and fell back with a `[mode-auto-downgrade]` info log), then re-allowed as experimental. The sentinel string survives as the documented historical fallback marker (`run.md`; CI sentinel audit). Evidence for the re-allow is two-sided on the same runtime version: one session observed 5 named workers (A–E) completing normally with result returns under the enabled flag, while another observed a named spawn converting to an in-process teammate that returned no result over ~1 hour and two status pokes (resolved only by TaskStop, which exposed its `in_process_teammate` type). The discrepancy is unresolved — treat teammate result-return reliability as unproven and verify per session before relying on it.

**Constraints (conditional — apply and re-measure if the teammate conversion resurfaces on a future CC version)**:
- No nested teams; one team per session; the lead is fixed
- In-process teammates cannot spawn background subagents (request-time error)
- `/resume` does not restore in-process teammates
- Permissions are fixed at spawn time
- `/model` IS inherited from the leader by default since CC 2.1.234 (the former Default teammate model `/config` setting was removed; a spawn-named model overrides; effort inheritance unchanged since v2.1.186)
- Team state `~/.claude/teams/{name}` and `~/.claude/tasks/{name}` is runtime-managed — never hand-edit
- Defining a subagent as a teammate skips `skills:` / `mcpServers:` frontmatter (loaded from project/user settings instead)
- GLM inheritance (load-bearing for cost): whether teammates inherit the lead's `ANTHROPIC_BASE_URL`/`ANTHROPIC_AUTH_TOKEN` is not officially documented — `moai cg`'s tmux env injection is a separate path and does not answer it; measure before relying on GLM-billed teammates
- Background subagents (the non-teammate path) have the Task tool family stripped from their schema (measured: TaskCreate/TaskUpdate/TaskList/TaskGet and ToolSearch absent from an unnamed background subagent; SendMessage present) — teammates reportedly regain the Task tool, so Task-based coordination is a teammate-path capability

### §C.2 `fanout` compound preference — three concurrency limits (SSOT)

`fanout` is preferred via the unified compound clause:

> `[Where the harness level is standard or thorough] [While the task scope is multi-domain (≥3 domains OR ≥10 files)] [When the orchestrator selects an execution mode in Phase 4]`, the orchestrator shall use parallel multi-spawn of retained agents (concurrent `Agent()` calls in a single message, within the fan-out bounds below). `fanout` remains the default multi-domain parallel mode; `agent-team` is the explicit-request experimental alternative (§C.1).

**Three concurrency limits exist. They are distinct numbers grounding distinct surfaces — never quote one as the cap for another:**

| # | Limit | Binds | Value | Grounding |
|---|-------|-------|-------|-----------|
| 1 | Subagent fan-out — **HARD bound** | Concurrent subagents per turn: every `Agent()` spawn surface, incl. `fanout` and any fan-out from a team lead or factory lead | `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, default **20** per turn (env-tunable; per-session total cap removed) | Claude Code runtime default, recorded in `CLAUDE.md` §14 and `moai-constitution.md` § Parallel Execution |
| 2 | Workflow agent concurrency | Agents within ONE dynamic workflow (`sweep` only) | **16** concurrent per workflow — runtime-documented as `min(16, available CPUs − 2)`; repo-recorded as "up to 16 concurrent agents (fewer on machines with limited CPU cores)" — plus the 1,000-total per-run backstop | `.claude/rules/moai/workflow/dynamic-workflows.md` |
| 3 | Team size — **ADVISORY** | Named teammates in ONE Agent Team (`agent-team`, experimental) | **3-5** teammates | Anthropic Agent Teams guidance: *"Start with 3-5 teammates for most workflows."* — team-composition advice, not a subagent cap |

**`fanout`'s own 3-5 is an advisory band, not a hard cap — and not the team-size quote.** The fan-out band is grounded in cache/prefix economics (concurrent spawns of the same agent definition cannot read a cache entry still being written: spawn ONE first and let the remaining N−1 read its cache — `.claude/rules/moai/workflow/cache-aware-execution.md` directive 2, stagger-spawn) and in coordination overhead (every fan-out result lands back in the orchestrator's context to reconcile). Its numeric coincidence with limit 3 is just that — a coincidence; the derivations are independent, and the former text that quoted "Start with 3-5 teammates" as a `fanout` cap conflated the two. The hard bound for fan-out is limit 1 (runtime cap, default 20): exceeding the 3-5 band violates no cap — it degrades cache economics and reconciliation quality. Write fan-out is bounded separately and more strictly (never two write-capable agents concurrently, `agent-common-protocol.md` § Background Agent Execution), regardless of count.

The stagger-spawn discipline (cache-aware-execution directive 2) governs **`fanout`** — same-type `Agent()` spawns from the orchestrator in one turn. It does NOT extend to `sweep`: the workflow runtime staggers its own agent starts automatically (`CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS`, default 5000ms), which is a different mechanism than orchestrator-side spawn ordering.

### §C.3 `sweep` capability gate

`sweep` is candidate ONLY when ALL of the following preconditions hold. The orchestrator MUST verify each before launching a Workflow:

| Precondition | Why it is required |
|--------------|---------------------|
| Implementation Kickoff Approval already passed | Workflow agents cannot prompt the user mid-run (no mid-run user input). Therefore the one decision that MUST involve the user — the plan→run human gate — MUST already be cleared. A `sweep` launch before Implementation Kickoff Approval passes is prohibited (§E anti-pattern). |
| All preferences collected | All user preferences (Tier, mode preference, PR strategy, etc.) MUST be drained at Implementation Kickoff Approval before launch, because the asymmetric boundary forbids both Workflow agents and goal-loop turn agents from prompting the user (agent-common-protocol.md § User Interaction Boundary). |
| Scope ≥ ~30 files, mechanical, genuinely parallel | The Workflow primitive earns its overhead only on genuinely-parallel high-volume mechanical work; coding-heavy / multi-domain work stays `serial` (Anthropic's coding-task parallelism caveat). |
| Workflows available | `CLAUDE_CODE_DISABLE_WORKFLOWS` is not set AND runtime version ≥ v2.1.154; otherwise fall through to `serial`. |
| Selection logged | The `sweep` selection AND a confirmation that Implementation Kickoff Approval already passed AND that all preferences were collected MUST be recorded in `progress.md § Mode Selection` before the Workflow launches (§D). |

#### `sweep` is scaling, not nesting

The Workflow is launched by the **orchestrator** (main session) as a scaling primitive. The Workflow script coordinates agents and keeps intermediate results in script variables; it returns only the final synthesis to the session context. This is NOT a subagent spawning a subagent — the flat hierarchy is preserved (Anthropic guidance: "Subagents cannot spawn other subagents" — the historical default). The concurrency model (16 concurrent / 1000-total backstop) is the published cap of the Workflow primitive cited from `dynamic-workflows.md`, NOT a MoAI-invented API.

> **Version note (Claude Code v2.1.219)**: subagent *nesting* is enabled by default as of v2.1.219 (changelog: up to depth 3; `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=1` disables — the brief v2.1.217–2.1.218 default-off state was reversed). `sweep` is unaffected by that change because it is orchestrator-launched **scaling, NOT subagent nesting**: the Workflow primitive is a main-session fan-out, not a subagent recursing into further subagent spawns.

#### No named-script Workflow API

The official Claude Code documentation does not document a typed named-script Workflow API. This rule describes only the conceptual *coordinate-agents → intermediate results in script variables → final synthesis* model. No named Workflow-script function signatures — an `agent`-function, a `parallel`-function, a `pipeline`-function, or a `phase`-function — are asserted anywhere (§E anti-pattern; the asserted-API prohibition).

#### `sweep` / goal-loop agents return blocker reports, never prompt the user

When a `sweep` Workflow agent or a goal-loop turn agent lacks a required input, that agent returns a structured blocker report; the orchestrator runs an `AskUserQuestion` round and re-delegates with the answers injected. Agents never prompt the user directly — this is the asymmetric orchestrator-subagent boundary (`.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary). The run-phase `ac_converge` wiring point and its semantic-failure escalation live in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (ac_converge).

### §C.4 Factory workers (default 8) are not `fanout`

`moai glm -k <N>` factory mode runs a fleet of independent worker sessions (tmux panes) — N is an operator-side fleet size (the count-less worker entry, a bare `-k --name worker-<i>`, defaults to 8), not subagent fan-out. The count sits under the factory's own workers-registry / free-slot discipline (registry prune, live-claim probe, staggered activation), not under `fanout`'s advisory band: 8 is a legal fleet size precisely because these workers are queue-polling sessions, not `Agent()` calls inside one orchestrator turn. Where a factory lead (or any worker session) DOES invoke `glm_task` / `Agent()` fan-out from its own session, that spawn surface is ordinary subagent fan-out: limit 1 (the runtime subagent cap, default 20) is the hard bound and the stagger-spawn discipline governs same-type spawns, exactly as for `fanout` (§C.2).

---

## §D — Logging Contract (progress.md § Mode Selection)

Per the canonical mode-logging policy, the orchestrator MUST record its mode-selection decision in `.moai/specs/SPEC-{ID}/progress.md` under a `## §F Phase 4 Mode Selection` section (preserving the `Mode Selection` token for the grep acceptance criterion) before spawning the first run-phase `Agent()` call. The `§F` letter is allocated by the canonical progress.md Section Map in `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map — Mode Selection MUST NOT reuse `§E` (the `§E.*` namespace is reserved for the era.go-parsed lifecycle-phase structure; overloading it would collide with era classification).

### §D.1 Required content

The Mode Selection section MUST include:

1. **Input parameters block** — values for tier, scope, domain count, file language mix, concurrency benefit, Agent Teams prereqs status
2. **Mode evaluation table** — for each of the 4 modes, a row stating "selected" or "not selected" and a one-line rationale
3. **Decision** — the chosen mode (one of: `direct`, `serial`, `fanout`, `sweep`; plus `agent-team` when explicitly operator-selected) on a single line for grep-friendly verification
4. **Justification** — a short paragraph (2-5 sentences) explaining why the chosen mode is preferable to alternatives, citing the relevant Anthropic finding(s) when applicable
5. **`sweep` confirmation (when `sweep` is selected)** — an explicit line confirming Implementation Kickoff Approval already passed AND all preferences were collected before the Workflow launches (§C.3 logging precondition)

### §D.2 Token requirement (grep verification)

The orchestrator's mode logging is verified by the canonical grep acceptance criterion via:

```bash
grep -A 5 "Mode Selection" .moai/specs/SPEC-{ID}/progress.md \
  | grep -c -i "direct\|serial\|fanout\|sweep\|agent-team"
```

The grep count MUST be ≥ 1. In practice, naming the chosen mode anywhere within 5 lines of the `Mode Selection` heading satisfies this; the structured `Decision: <mode>` line accomplishes this directly.

### §D.3 When to log a boundary case

When the decision tree hit a boundary (e.g., scope = exactly 10 files, exactly 3 domains, harness = `standard` with team.enabled = true but env var unset), the orchestrator MUST additionally include a **Boundary Case** subsection documenting the tie-breaker rule that resolved the ambiguity. This enables retrospective analysis to recalibrate threshold values across SPECs.

---

## §E — Anti-Patterns

The following patterns violate the orchestration mode selection contract:

- **Auto-selecting `agent-team`** — `agent-team` is experimental and explicit-request-only; the orchestrator MUST NOT auto-select it. It is entered only via an explicit operator `--team` / `--mode team` / `Team` scale label (§C.1)
- **Quoting the team-size advisory as a `fanout` cap, or treating `fanout`'s 3-5 band as a hard cap** — "3-5 teammates" is Anthropic's TEAM SIZE guidance for `agent-team`; `fanout`'s 3-5 is an independent cache/coordination advisory under the runtime subagent cap (default 20) as the hard bound (§C.2). Exceeding the band degrades cache economics and reconciliation quality, and same-type fan-out that skips the stagger-spawn discipline (cache-aware-execution directive 2) forfeits the cache reads the band is grounded in
- **Selecting `fanout` for coding-heavy work** — violates Anthropic's coding-task parallelism caveat; `serial` is the correct default for coding tasks
- **Selecting `sweep` for coding-heavy / multi-domain / new-code work** — violates Anthropic's coding-task parallelism caveat; `sweep` admits ONLY genuinely-parallel high-volume mechanical work (one uniform transform rule, no inter-file dependency). Coding-heavy work belongs to `serial`
- **Launching a `sweep` Workflow before Implementation Kickoff Approval has passed** — violates the Implementation Kickoff Approval mandatory-restoration policy; the orchestrator MUST NOT launch the Workflow before Implementation Kickoff Approval user approval and MUST return control to the Implementation Kickoff Approval `AskUserQuestion` gate. `sweep` is strictly downstream of Implementation Kickoff Approval
- **Asserting a typed/named Workflow script API** — a named `agent`-function, `parallel`-function, `pipeline`-function, or `phase`-function signature is NOT documented by Claude Code; describe the conceptual coordinate-agents → script-variable results → final-synthesis model instead (the named-script-API prohibition)
- **Selecting `sweep` without recording the Implementation Kickoff Approval-passed + preferences-collected confirmation in `progress.md`** — the §C.3 / §D.1 #5 logging precondition makes the autonomy decision auditable; skipping it leaves the Workflow launch unverifiable post-hoc
- **Skipping the progress.md logging step** — fails the canonical mode-logging acceptance criterion; the decision is no longer auditable post-hoc
- **Re-spawning the same mode for multiple consecutive milestones in `serial` without re-evaluating** — acceptable practice for a single SPEC, but when run-phase scope changes mid-flight (e.g., milestone scope-up via blocker report), the orchestrator SHOULD re-run Phase 4
- **Substituting an `AskUserQuestion` round for the autonomous decision** — Phase 4 is autonomous by contract; user intervention belongs to Phase 1 verdict review (when verdict is FAIL or INCONCLUSIVE) or Implementation Kickoff Approval (plan-to-implement HUMAN GATE), not Phase 4

---

## §F — Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification — `--mode` flag matrix and Mode Dispatch sentinels (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 1 Plan Audit Gate — runs before Phase 4 and may produce `BYPASSED` / `INCONCLUSIVE` / `FAIL` verdicts that affect Phase 4 inputs
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S/M/L delegation template selection (interacts with `serial` spawn prompts)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — sibling rule documenting the orchestrator's rejection behavior when a paste-ready resume references an archived-agent name (independent of mode selection)
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the sweep primitive: 16-concurrent / 1000-total cap, no-mid-run-user-input semantics, Implementation Kickoff Approval-is-unaffected note, and the absence of a documented named-script API
- `.claude/rules/moai/workflow/goal-directive.md` — `/moai goal` autonomous-continuation semantics (the run-phase `ac_converge` condition wiring lives in `run.md` § Run-phase Autonomy (ac_converge))
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (ac_converge) — co-located Implementation Kickoff Approval ordering reference + `ac_converge` arming
- The canonical agent catalog design — design-time decision tree from which this rule was derived
- Anthropic Sub-agents and Agent Teams documentation — verbatim citations grounding the `agent-team` team-size advisory and the fanout-vs-serial coding-task caveat
- Anthropic Agent Teams documentation — *"Start with 3-5 teammates for most workflows."* (team-size advisory binding `agent-team` only — never a subagent fan-out cap; see §C.2)
- Anthropic multi-agent research engineering note — *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."*

---

## §G — Two-Axis Confusion Warning (sweep wiring author)

[ZONE:Frozen] [HARD] `sweep` lives in ONE specific list: the **Phase 4 execution-mode catalog** in this file (`direct` / `serial` / `fanout` / `sweep`). This is a DIFFERENT axis from the `run.md` `--mode` **dispatch axis** (`autopilot` / `loop` / `team` / `pipeline`) documented in `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification and the `run.md` Mode Dispatch table. Both happen to be short lists, which is the confusion trap.

- **Execution-mode catalog** (Phase 4, where `sweep` lives): governs HOW the orchestrator spawns — concurrent spawn count and spawn surface (orchestrator-executes / sequential `Agent()` / parallel `Agent()` / **Workflow fan-out**; `agent-team` as the explicit-request footnote surface).
- **`--mode` dispatch axis** (CLI flag, NOT touched here): governs WHICH `/moai run` workflow variant runs — `autopilot` vs `loop` (Ralph) vs `team` vs the rejected `pipeline`.

[ZONE:Frozen] [HARD] `sweep` is a catalog mode ONLY. It is NOT a new `--mode` dispatch value; no `--mode sweep` flag is introduced; the `run.md` Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is unchanged. The header cross-reference above already notes the two axes "interact with — but are separate from" each other; this separation is preserved.

---

## §G.1 — Dispatch-Axis Crosswalk (`--mode` values + scale-table labels → catalog modes)

> **Correspondence, not merge.** This crosswalk documents how each `--mode` dispatch-axis value and each Phase 4 scale-table label CORRESPONDS to a catalog mode. It does NOT merge the two axes: per §G they remain separate, the `--mode` value set `{autopilot, loop, team, pipeline}` is unchanged, no `--mode sweep` value is introduced, and the Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is untouched.

### `--mode` dispatch-axis values → catalog modes

| `--mode` value | Corresponds to catalog mode | Notes |
|----------------|------------------------------|-------|
| `autopilot` | `serial` | Default single-lead orchestration; the Phase 4 scale-based selection chooses the envelope (see scale-label rows below). |
| `loop` | `serial` | Ralph-engine diagnostic fix-loop variant — sequential per-iteration delegation. The granularity differs (diagnostics, not phases) but the spawn shape is the sequential sub-agent. |
| `team` | `agent-team` — experimental (re-allowed) | `--mode team` selects the Agent Teams layer (operator decision; flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` ships on). Historical: the retired era emitted `MODE_TEAM_UNAVAILABLE` and fell back per the Mode Resolver — the sentinel string is retained in `run.md` for the CI audit and as genealogy. |
| `pipeline` | `serial` — utility subcommands only | Rejected on multi-agent subcommands (`MODE_PIPELINE_ONLY_UTILITY`); the utility subcommands are intrinsically sequential, so `pipeline` names their fixed direct / sequential execution shape — the `--mode pipeline` flag itself is ignored there (`MODE_FLAG_IGNORED_FOR_UTILITY`), not honored. |

### Phase 4 scale-table labels → catalog modes

| Scale label | Corresponds to catalog mode | Envelope |
|-------------|------------------------------|----------|
| Fix | `serial` | Minimal envelope — single implementation agent + orchestrator verification batch. |
| Focused | `serial` | Focused envelope — single implementation agent with domain context injected. |
| Standard | `serial` | Standard envelope — planning + implementation + audit, sequential. |
| Full Pipeline | `serial` | Full envelope — full sequential agent chain (plan → implement → audit → docs). |
| Team | `agent-team` — experimental (re-allowed) | The `Team` scale label selects the Agent Teams layer (§C.1). Historical: the retired-era `--team` emitted `MODE_TEAM_UNAVAILABLE` and fell back; retained as genealogy. |

Every `--mode` value and every scale label corresponds to exactly one catalog mode. `direct` and `sweep` have NO dispatch-axis or scale-label counterpart — they are selectable only via the Phase 4 decision tree (§B). This asymmetry is expected and is further evidence the two axes are distinct.

---

## §G.2 — `manager-lead` as a serial-shaped delegation target (NOT a new mode)

> **Non-regression note (the hierarchical-team SPEC).** Adding `manager-lead` to the retained-agent catalog does NOT alter the Phase 4 execution-mode catalog in §A. `manager-lead` is a serial-shaped delegation target: the orchestrator spawns it sequentially (the `serial` envelope), and `manager-lead` in turn fans out write-capable leaf workers under the depth-2 seal (the sole Agent-carrier carve-out, depth-2 sealed). This is NOT a new mode. `agent-team` is experimental and explicit-request-only (§C.1; Tier L auto-routing still targets `manager-lead`), the `MODE_TEAM_UNAVAILABLE` sentinel is retained as documented history, and the `--mode` dispatch-axis values (`autopilot` / `loop` / `team` / `pipeline`) are unchanged. The entry predicate for `manager-lead` (≥3 milestones AND ≥10 files AND cross-domain fan-out) is logged in `progress.md §F Mode Selection` before the spawn, exactly as any other serial delegation; `manager-lead` does not modify the decision tree in §B — it is selectable once the orchestrator's Tier L coordination threshold is met.

---

Version: 2.0.0 (catalog renamed onto the single concurrent-spawn-count axis: `trivial`→`direct`, `sub-agent`→`serial`, `parallel`→`fanout`, `workflow`→`sweep`; `background` demoted to an execution option (subagents default background since CC v2.1.198); `agent-team` moved to the §A footnote + §C.1; legacy-token mapping added at the top; stagger-spawn scoped to `fanout` with the workflow runtime's own auto-stagger excluded; §A/§B/§C/§D/§E/§F/§G aligned. Prior 1.4.0: §C.2 rebuilt — three distinct concurrency limits; §C.4 factory-workers reconciliation; §G.2 manager-lead non-regression note)
Origin: derived from the canonical agent catalog and IGGDA policies.
Status: Active — applies to all `/moai run` Phase 4 invocations
