---
paths: "**/.claude/workflows/**,**/.moai/config/sections/workflow.yaml,**/dynamic-workflows.md"
---

# Dynamic Workflows (Claude Code Orchestration Primitive)

Guidance for the Claude Code **dynamic workflow** primitive — a script the runtime executes to orchestrate subagents at scale. Distinct from MoAI's SPEC plan/run/sync workflow (which is a development lifecycle, not a runtime primitive).

> **Loading scope**: orchestration-level guidance. Read when deciding how to fan out a large task across many agents, or when a user asks for a "workflow".

## What It Is

A dynamic workflow is a JavaScript script that the Claude Code runtime executes in the background to orchestrate [subagents](https://code.claude.com/docs/en/sub-agents). Claude writes the script for the task; the runtime runs it while the session stays responsive. Intermediate results live in script variables instead of the conversation context, so only the final answer returns to the session.

Requires Claude Code v2.1.154 or later (research preview). Reference: https://code.claude.com/docs/en/workflows

## The Three Orchestration Primitives

MoAI now recognizes three runtime primitives for multi-step work. The difference is **who holds the plan**:

| Primitive | Who decides next step | Intermediate results | Scale | Repeatable unit |
|-----------|----------------------|----------------------|-------|-----------------|
| **Subagents** (`Agent()`) | Claude, turn by turn | Claude's context window | A few delegated tasks per turn | The agent definition |
| **Agent Teams** | Claude + teammates via shared TaskList | Each teammate's own context | 3-5 teammates (Anthropic recommendation) | The team composition |
| **Dynamic Workflows** | The script | Script variables | Dozens to hundreds of agents per run | The orchestration script |

Subagents and skills keep the plan in Claude's context (it decides turn by turn). A workflow moves the plan into code: the script holds the loop, the branching, and the intermediate results, so the session context holds only the final answer. This also lets a workflow apply a repeatable quality pattern — e.g. independent agents adversarially reviewing each other's findings before reporting.

## When to Use a Dynamic Workflow

Reach for a workflow when a task needs **more agents than one conversation can coordinate**, or when the orchestration should be codified as a script you can read and rerun:

- Codebase-wide sweeps (bug hunt across every file, audit every endpoint for missing auth)
- Large migrations (hundreds of call sites transformed independently)
- Research questions where sources must be cross-checked against each other
- A hard plan worth drafting from several independent angles before committing to one

### When NOT to Use a Workflow

- A task one conversation can coordinate with a handful of subagents → use `Agent()` directly
- Interactive, iterative work needing user sign-off between stages → workflows take no mid-run user input
- Work that must call MoAI's interactive surfaces (`AskUserQuestion`) mid-run → not available inside workflow agents
- Routine single-file edits → direct execution

The Anthropic guidance is explicit that most coding tasks involve fewer truly parallelizable subtasks than research, so the default for coding-heavy work remains sequential subagents; reserve workflow-scale fan-out for genuinely parallel, large-volume work.

### Routing Heuristic (which primitive to pick)

When choosing among the three runtime primitives, route by the **shape and volume** of the work (this heuristic reuses, and does not contradict, the three-primitive table above):

- **Dynamic workflow** — when the work fans out over **dozens-to-hundreds** of mostly read-only, independent items (a codebase-wide sweep, a large mechanical migration, cross-checked research). The script holds the plan and the intermediate results, so the session context stays small even at high agent counts.
- **Agent Teams** — when a **small number** of long-running peers must coordinate through a shared task list (cross-layer work where teammates hand off and review each other). Start with 3-5 teammates; coordination cost rises sharply beyond that.
- **Sequential subagents** — the **default** for coding-heavy run-phase work. One subagent per milestone, each result landing back in Claude's context. Prefer this whenever the task is not genuinely high-volume parallel, because coding tasks rarely decompose into many truly independent subtasks.

The deciding question is **who should hold the plan**: the script (workflow), a coordinating peer set (Agent Teams), or Claude turn-by-turn (sequential subagents).

## How a Workflow Runs

- The runtime executes the script in an isolated environment, separate from the conversation.
- Up to **16 concurrent agents** (fewer on machines with limited CPU cores); **1,000 agents total per run** as a runaway-loop backstop.
- **Workflow size is user-tunable** — the `/config` **Dynamic workflow size** setting (`small` / `medium` / `large`, v2.1.202+) sets a guideline for how many agents a workflow targets, scaling the effective agent count within the 16-concurrent / 1,000-total ceilings above. MoAI does not pin a size in the deployed template — the choice is left to the user/org, so a size guideline surfaced in a session (e.g. "keep workflows under 15 agents") is user configuration, not a MoAI default.
- **No mid-run user input** — only agent permission prompts can pause a run. For sign-off between stages, run each stage as its own workflow.
- The workflow script itself has **no direct filesystem or shell access** — its agents read, write, and run commands; the script only coordinates them.
- Runs are **resumable within the same session**: completed agents return cached results, the rest run live. Exiting Claude Code restarts a running workflow fresh in the next session.
- Workflow subagents always run in `acceptEdits` mode and inherit the session tool allowlist regardless of the session's permission mode. Add the commands agents need to the allowlist before a long run to avoid mid-run prompts.
- **The script body must be deterministic** — it must not call wall-clock or random-number functions. Resume caching keys on the script's deterministic outputs, so a clock read or a random draw produces a different result on resume and silently breaks the cache. Any timestamp or random value the workflow needs must be injected through the script's input arguments, or stamped onto the results after the run returns — never generated inside the script body. (Note: *mentioning* `Date.now()` / `Math.random()` inside a prompt string or comment is fine — Claude Code 2.1.172 fixed an over-strict validation that previously rejected scripts merely referencing these functions; only an actual *call* in the script body breaks resume caching.)
- **Per-run approval depends on the permission mode**: under Default or accept-edits permission modes the runtime prompts for approval on every workflow run; under Auto mode it prompts only on the first launch; under Bypass mode, headless `-p`, and the SDK it never prompts. This per-run gate is an execution-level approval and is separate from MoAI's Implementation Kickoff Approval plan-to-implement human gate (see § MoAI Integration Notes).

### Manage runs

While a workflow run is active, the `/workflows` TUI lets you manage it: list active and recent runs, watch a run's live progress, pause a run, resume a paused run, and save a finished run's script as a reusable command. The default key bindings inside the TUI are `p` (pause), `x` (cancel/stop), `s` (save), and `r` (resume).

## MoAI Integration Notes

- **AskUserQuestion boundary still holds**: workflow agents cannot prompt the user (same asymmetric boundary as subagents per `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary). The MoAI orchestrator collects all preferences via `AskUserQuestion` BEFORE launching a workflow, never inside it.
- **Implementation Kickoff Approval is unaffected**: a workflow is a run-phase execution mechanism. The plan-to-implement human gate is decided by the orchestrator before any workflow launches, not by the workflow.
- **Cost awareness**: a single workflow run can spend meaningfully more tokens than the same task in conversation. It counts toward the session's usage and the context-window thresholds in `.claude/rules/moai/workflow/context-window-management.md`. Surface the cost trade-off to the user before launching a large fan-out.
- **Bundled `/deep-research`**: Claude Code ships a built-in research workflow (`/deep-research <question>`) that fans out web searches, cross-checks sources, votes on claims, and returns a cited report. It requires the WebSearch tool. This complements MoAI's WebSearch + Explore exploration pattern for research-heavy questions.
- **`ultracode` per-prompt trigger vs session effort**: the `ultracode` trigger keyword (or asking to "use a workflow") is a **per-prompt** trigger — it launches a workflow for that one request. This is distinct from the **session-wide** `/effort ultracode` mode, which combines `xhigh` reasoning with automatic workflow orchestration so Claude plans a workflow for each substantive task across the whole session. Use the session mode deliberately; every task then uses more tokens. Session mode reverts on a new session; step back with `/effort high` for routine work. Because it resets on a new session, `ultracode` is **not** restored by the `ultrathink.` opener of a paste-ready resume message — that opener restores reasoning effort only. A resumed session that needs auto-orchestration must explicitly re-issue `/effort ultracode`, parallel to how a `/goal` must be re-set after a session boundary.
- **Saved workflows**: a run's script can be saved as a `/command` in `.claude/workflows/` (project, shared) or `~/.claude/workflows/` (personal). A project workflow with the same name wins over a personal one. A saved workflow accepts an `args` global input — the arguments string passed when the workflow command is invoked. MoAI does not ship any saved workflows by default; the user-owned `.claude/workflows/` directory is not template-managed.
- **Plan / provider availability**: dynamic workflows require a paid plan and are available on the Claude API, Amazon Bedrock, Google Vertex AI, and Microsoft Foundry; on the Pro plan the feature is enabled via `/config`.

## Purpose-driven model+effort selection

The dynamic workflow `agent()` primitive accepts an opts object `{model, effort, agentType, isolation, phase, schema, label}` (per `https://code.claude.com/docs/en/workflows`). Omitting `model` inherits the main-loop model; omitting `effort` inherits the session effort. Because a paste-ready resume message's `ultrathink.` opener commonly leaves the session at `xhigh`, a workflow `agent()` call that omits `effort` silently runs every spawned agent at `xhigh` — including mechanical read-only extraction, which the official guidance recommends at `low`. That silent inheritance is a cost leak.

[ZONE:Evolvable] [HARD] When a `.claude/workflows/*.js` script invokes `agent()`, the script author SHALL set `effort` explicitly per the purpose taxonomy below rather than inheriting the session default. Set `model` explicitly only when the purpose demands a specific tier (haiku for mechanical extraction; opus for deep architectural reasoning); otherwise omit it to inherit the main-loop model.

The official effort levels are `low`, `medium`, `high` (default), `xhigh`, `max` (`https://platform.claude.com/docs/en/build-with-claude/effort`). The taxonomy below maps each workflow-agent purpose to a recommended `(model, effort)`.

> **Config surface.** The `workflow_agents:` block in `.moai/config/sections/workflow.yaml` is the SSOT for these per-purpose `(model, effort)` DEFAULTS — the web console and tooling read/write that block, and per-script literals in `.claude/workflows/*.js` remain overrides that win over the config defaults. Values are validated against the closed sets above (model: inherit/haiku/sonnet/opus; effort: low/medium/high/xhigh/max).

| Purpose | Example surfaces | Recommended model | Recommended effort | Official citation |
|---------|------------------|-------------------|--------------------|-------------------|
| **read-only-extract** | per-package dep-graph + public-surface extraction; mechanical AST/grep sweeps | haiku | **low** | "`low` — Simpler tasks that need the best speed and lowest costs, such as subagents" |
| **mechanical-transform** | large migrations (call-site rename, API shape change); mechanical refactors | sonnet | **medium** | "`medium` — Balanced reasoning for general tasks" |
| **synthesize** | architectural synthesis layered on deterministic extraction; multi-source research synthesis | sonnet | **high** | "`high` — Most tasks; good balance of quality and speed" |
| **research** | cross-checked research with adversarial voting; deep single-topic investigation | sonnet or opus | **high** or **xhigh** | research effort should scale with claim density the research must adjudicate (project-internal heuristic, not a verbatim prescription) |
| **verify-judge** | code review (security/perf/arch dimensions); independent plan/spec audit; quality scoring | sonnet or opus | **xhigh** | "minimum `high` for intelligence-sensitive work" |
| **implement** | code generation (backend/frontend/full-stack); test writing | sonnet or opus | **xhigh** | "`xhigh` for coding/agentic work" |
| **design-architecture** | solution architecture decisions; system design; deep reasoning over trade-offs | opus | **xhigh** | "`xhigh` for coding/agentic work" + "minimum `high` for intelligence-sensitive work" |

**Reading order.** When a workflow agent serves multiple purposes, pick the highest-effort purpose in the table. When purpose is ambiguous, prefer the cheaper effort — the cost of over-efforting a read-only extraction is a silent token leak; the cost of under-efforting a verify-judge is a missed defect.

**Worked example — codemaps-extract.js.** The bundled `.claude/workflows/codemaps-extract.js` fans out one `Explore` agent per source package for read-only dep-graph + public-surface extraction plus an architectural-synthesis layer. Each per-package `agent()` call carries `agentType: 'Explore'` and `effort: 'low'` — the read-only-extract purpose, per the official "`low` — such as subagents" guidance. The synthesis layer's architecture-insight value comes from the prompt, not from raising effort; raising effort on the extraction step would multiply token cost without improving the mechanical baseline (see the script's VERDICT SCOPING header). This is the canonical pattern for mechanical read-only fan-out.

## Disabling Workflows

Workflows can be turned off per-user (`/config` Dynamic workflows toggle, `"disableWorkflows": true` in `~/.claude/settings.json`, or `CLAUDE_CODE_DISABLE_WORKFLOWS=1`) or org-wide via the `workflowKeywordTriggerEnabled` managed setting (v2.1.157+; org admins set it to `false` to disable the keyword trigger fleet-wide). When disabled, the bundled workflow commands are unavailable, the `ultracode` trigger keyword no longer triggers a run, and `ultracode` is removed from the `/effort` menu. (`ultracode` is the current trigger keyword as of v2.1.160; `workflow` was the pre-v2.1.160 keyword — a plain natural-language request still routes to a workflow run on both versions.) MoAI does not enable or disable workflows in the deployed template — the decision is left to the user/org.

## Pattern Catalog

Validated patterns from MoAI dynamic-workflow pilots — each entry records the pattern shape, the primitive mechanics, and the falsification verdict that justified (or scoped) it.

### Per-Package Codemaps Extraction Fan-Out

**Pattern**: one read-only agent per source package extracts that package's dependency graph + public surface + an architectural synthesis; results aggregate in script variables; only the final synthesis returns to the session.

**Primitive mechanics**:
- The script fans out N read-only agents — a parallel call maps the package list to one agent per package, each typed read-only (Explore). The script array holds the intermediate results, so the session context stays small.
- Determinism: the package list is injected via the `args` global; the script body reads no wall-clock and draws no random value; any timestamp is stamped after the run returns (keeps the resume cache valid).
- Read-only is enforced by the Explore agent type. A forced output schema is omitted to avoid rate-limit brittleness — agents return markdown, and the orchestrator parses it and applies a synthesis-vs-restatement reduction test.

**Falsification verdict — value proven, with three scoping caveats**:
1. **Augmentation, not extraction.** Dependency-graph + public-surface extraction is mechanically complete from `go list -deps -json` + `go doc`; the per-package LLM synthesis adds zero to extraction. Its surviving value is architecture-REVIEW insight (coupling risk, latent contracts, layering judgments, negative-space gaps) layered on top of the deterministic baseline.
2. **Not primitive-specific.** An identical synthesis is obtainable from a sequential sub-agent or a single Explore agent; the fan-out's only marginal benefit is parallel wall-clock speed.
3. **High-count justification only.** The fan-out earns its token cost ONLY at high package count (near the full codebase) where parallel speed offsets per-agent cost; at small scale a single sub-agent is cheaper.

**When to use**: high-count codebase codemaps where architecture INSIGHT beyond mechanical extraction is wanted AND parallel speed matters. **When NOT to use**: pure dependency-graph / public-surface extraction (use the deterministic `go list -deps -json` + `go doc` path), or small scale (use a single sub-agent).

**Artifact**: the validated script lives in the local, user-owned `.claude/workflows/` directory (not template-managed, per the statement above).

## Cross-references

- https://code.claude.com/docs/en/workflows — canonical Claude Code workflows documentation
- `.claude/rules/moai/core/moai-constitution.md` § Parallel Execution — orchestration primitive selection
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.1 — Agent Teams static layer (RETIRED; Mode 3 `agent-team` tombstone)
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — AskUserQuestion asymmetry (applies to workflow agents)
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` autonomous-continuation primitive (complementary)

---

Version: 1.0.0
Classification: Evolvable orchestration guidance — applies when fanning out large tasks across many agents
