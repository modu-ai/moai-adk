---
description: "Phase 0.95 Mode Selection — 5-mode autonomous decision tree for the MoAI orchestrator (trivial / background / agent-team / parallel / sub-agent). Read at every /moai run entry."
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/plan.md,.claude/rules/moai/workflow/spec-workflow.md"
metadata:
  version: "1.0.0"
  status: "active"
  updated: "2026-05-25"
  tags: "orchestration, mode-selection, agent-teams, phase-0.95, anthropic-2026-alignment"
---

# Orchestration Mode Selection — Phase 0.95

Canonical 5-mode autonomous decision tree for the MoAI orchestrator. Activated at Phase 0.95 (after Phase 0.5 plan-auditor verdict, before Phase 1 implementation) per REQ-ATR-008 + REQ-ATR-013 + REQ-ATR-017 of SPEC-V3R6-AGENT-TEAM-REBUILD-001. The decision is autonomous (no `AskUserQuestion` round); the chosen mode and the selection rationale are logged to `progress.md § Mode Selection`.

> Cross-reference: `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/design.md` §B.4 documents the design-level decision tree; this rule file is the canonical runtime SSOT for the orchestrator's mode-selection behavior. `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification covers the `--mode` flag matrix (autopilot / loop / team / pipeline) which interacts with — but is separate from — the 5-mode catalog below.

---

## §A — Mode Catalog (5 modes)

The orchestrator selects exactly one of the following modes per Phase 0.95 invocation.

| # | Mode | Concurrency | Spawn Surface | When to prefer |
|---|------|-------------|---------------|----------------|
| 1 | `trivial` | None — direct execution by the orchestrator, no sub-agent spawn | n/a | Typo fix, single-line formatting, no semantic change |
| 2 | `background` | 1 concurrent sub-agent | `Agent(run_in_background: true, ...)` | Read-only analysis that can complete asynchronously without blocking the conversation |
| 3 | `agent-team` | 3-5 dynamic teammates | `TeamCreate(...)` + `Agent(subagent_type: "general-purpose", team_name: ..., name: ...)` × N | Multi-domain (≥3 domains OR ≥10 files) research-heavy work AND all REQ-ATR-013 capability-gate prerequisites met |
| 4 | `parallel` | 3-5 concurrent sub-agents (single message, multiple `Agent()` calls) | Multiple `Agent()` invocations in one assistant turn | Multi-domain research that does NOT meet Agent Teams prerequisites; or any case where Agent Teams session overhead exceeds benefit |
| 5 | `sub-agent` | 1 sequential sub-agent per milestone | Sequential `Agent(...)` spawns, one milestone at a time | Coding-heavy work (Finding A4 caveat), or any case where the simpler mode suffices |

Mode 5 is the **default fallback** when no other mode's selection criteria are unambiguously met.

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
  │   AND is the work read-only (no Write/Edit per CONST-V3R2-020)?
  │   ├── YES → Mode 2: BACKGROUND (Agent run_in_background: true)
  │   └── NO  → continue
  │
  ├── Does the task meet ALL Agent Teams capability-gate conditions?
  │   (REQ-ATR-013 — all three required):
  │     • harness level is `thorough` (`.moai/config/sections/harness.yaml`)
  │     • `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml`
  │     • environment variable `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
  │   AND is the scope multi-domain (≥3 domains OR ≥10 files)?
  │   ├── YES → Mode 3: AGENT-TEAM (TeamCreate + dynamic teammate spawn)
  │   └── NO  → continue
  │
  ├── Is the task multi-domain (≥3 domains) AND research-heavy
  │   (NOT coding-heavy per Finding A4 caveat)?
  │   ├── YES → Mode 4: PARALLEL (3-5 concurrent Agent() in single message)
  │   └── NO  → continue
  │
  └── Default → Mode 5: SUB-AGENT (single Agent() sequential spawn per milestone)
```

### §B.1 Input parameters

The orchestrator collects the following signals before traversing the decision tree:

- **tier**: SPEC tier (S / M / L) per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- **scope (file count)**: estimated count of files in the SPEC's run-phase scope
- **domain count**: number of distinct domains touched (agents, workflow skills, rules, hook scripts, template mirrors, Go source code, SPEC artifacts, etc.)
- **file language mix**: e.g., 100% markdown vs Go source vs shell vs mixed
- **concurrency benefit**: HIGH for research-heavy work (parallel reads, independent perspectives); LOW for coding-heavy work (Finding A4 caveat — most coding tasks involve fewer truly parallelizable tasks than research)
- **Agent Teams prereqs status**: each of the three REQ-ATR-013 conditions individually verified

### §B.2 Tie-breaker rules (boundary cases)

Phase 0.95 boundary cases (scope at threshold ±1, ambiguous domain count, etc.) follow these defaults:

- At threshold ±1 (9 vs 10 files; 2 vs 3 domains): default to the **simpler** mode (sub-agent over agent-team; sequential over parallel)
- **Coding-heavy + multi-domain**: prefer Mode 5 over Mode 4 (Finding A4 — coding-task parallelism caveat)
- **Markdown-heavy + multi-domain + research-heavy**: prefer Mode 4 (parallel multi-spawn)
- **Tier L + markdown / shell-script-only scope**: Mode 5 with Tier L Section A-E delegation template (per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability)
- **Tier S + minimal scope**: Mode 5 with the minimal delegation form (~500-800 tokens, Section B may be filtered)

---

## §C — Capability Gates

### §C.1 Mode 3 (Agent Teams) capability gate — REQ-ATR-013

Mode 3 is candidate only when all three conditions hold simultaneously. The orchestrator inspects the runtime environment + project config:

| Condition | Where to check | Required value |
|-----------|----------------|----------------|
| harness level is `thorough` | `.moai/config/sections/harness.yaml` `harness.level` | `thorough` |
| Agent Teams enabled in workflow | `.moai/config/sections/workflow.yaml` `workflow.team.enabled` | `true` |
| Experimental flag set in environment | env var `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` | `1` |

When any condition fails, the orchestrator falls through to Mode 4 evaluation. If a user explicitly requested `--mode team` and the prereqs are not met, the orchestrator emits the canonical sentinel `MODE_TEAM_UNAVAILABLE` (per `.claude/rules/moai/workflow/spec-workflow.md` § Mode Dispatch) and continues with the fallback mode plus a `[mode-auto-downgrade]` info log.

Anthropic Agent Teams documentation verbatim guidance applies: *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* The orchestrator MUST NOT spawn fewer than 3 nor more than 5 teammates in agent-team mode.

### §C.2 Mode 4 (Parallel) Compound preference — REQ-ATR-017

Mode 4 is preferred via the unified compound clause:

> `[Where the harness level is standard or thorough] [While the task scope is multi-domain (≥3 domains OR ≥10 files)] [When the orchestrator selects an execution mode in Phase 0.95]`, the orchestrator shall prefer Agent Teams mode if all REQ-ATR-013 prerequisites are met; otherwise the orchestrator shall fall back to parallel multi-spawn of retained agents (maximum 3-5 concurrent `Agent()` calls in a single message per Anthropic verbatim "Start with 3-5 teammates").

The 3-5 ceiling applies equally to Mode 3 (agent-team teammates) and Mode 4 (concurrent `Agent()` spawn calls). Exceeding the ceiling regresses to coordination overhead and contradicts Anthropic's published guidance.

---

## §D — Logging Contract (progress.md § Mode Selection)

Per REQ-ATR-008, the orchestrator MUST record its mode-selection decision in `.moai/specs/SPEC-{ID}/progress.md` under a `## §E — Phase 0.95 Mode Selection` (or analogous section name preserving the `Mode Selection` token for AC-ATR-008 grep verification) before spawning the first run-phase `Agent()` call.

### §D.1 Required content

The Mode Selection section MUST include:

1. **Input parameters block** — values for tier, scope, domain count, file language mix, concurrency benefit, Agent Teams prereqs status
2. **Mode evaluation table** — for each of the 5 modes, a row stating "selected" or "not selected" and a one-line rationale
3. **Decision** — the chosen mode (one of: `trivial`, `background`, `agent-team`, `parallel`, `sub-agent`) on a single line for grep-friendly verification
4. **Justification** — a short paragraph (2-5 sentences) explaining why the chosen mode is preferable to alternatives, citing the relevant Anthropic finding(s) when applicable

### §D.2 Token requirement (AC-ATR-008 grep verification)

The orchestrator's mode logging is verified by AC-ATR-008 via:

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
- **Selecting Mode 4 (Parallel) for coding-heavy work** — violates Finding A4 caveat; Mode 5 (Sub-Agent sequential) is the correct default for coding tasks
- **Skipping the progress.md logging step** — fails AC-ATR-008 verification; the decision is no longer auditable post-hoc
- **Re-spawning the same mode for multiple consecutive milestones in Mode 5 without re-evaluating** — acceptable practice for a single SPEC, but when run-phase scope changes mid-flight (e.g., milestone scope-up via blocker report), the orchestrator SHOULD re-run Phase 0.95
- **Substituting an `AskUserQuestion` round for the autonomous decision** — Phase 0.95 is autonomous by contract; user intervention belongs to Phase 0.5 verdict review (when verdict is FAIL or INCONCLUSIVE) or GATE-2 (plan-to-implement HUMAN GATE), not Phase 0.95

---

## §F — Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification — `--mode` flag matrix and Mode Dispatch sentinels (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate — runs before Phase 0.95 and may produce `BYPASSED` / `INCONCLUSIVE` / `FAIL` verdicts that affect Phase 0.95 inputs
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S/M/L delegation template selection (interacts with Mode 5 sub-agent spawn prompts)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — sibling rule documenting the orchestrator's rejection behavior when a paste-ready resume references an archived agent (independent of mode selection)
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/design.md` §B.4 — design-time decision tree from which this rule was derived
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` §B.1 Findings A1-A6 — Anthropic 2026 verbatim citations grounding the Mode 3 ceiling (Finding A1) and Mode 4-vs-Mode-5 coding-task caveat (Finding A4)
- Anthropic Agent Teams documentation — *"Start with 3-5 teammates for most workflows."*
- Anthropic multi-agent research engineering note — *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."*

---

Version: 1.0.0
Origin: SPEC-V3R6-AGENT-TEAM-REBUILD-001 (M5)
Status: Active — applies to all `/moai run` Phase 0.95 invocations
