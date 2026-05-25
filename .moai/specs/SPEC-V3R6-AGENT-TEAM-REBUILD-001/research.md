---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
artifact: research
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: orchestrator (synthesizing 3 prior parallel research agents + 3 deep audit reports)
sources:
  - "Anthropic official docs: claude.com/docs/en/{agents,skills,best-practices,hooks,sub-agents,agent-teams}, claude.com/blog, anthropic.com/engineering"
  - "Context7 MCP: /anthropics/claude-code, /zebbern/claude-code-guide, /anthropics/anthropic-sdk-{python,typescript}, /nothflare/claude-agent-sdk-docs"
  - "GitHub: revfactory/harness (Apache 2.0), forrestchang/andrej-karpathy-skills"
  - "Internal: .moai/research/anthropic-best-practices-2026-05-24.md (13 findings prior audit) + .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md (base prior research, 2026-05-25)"
  - "Deep audit reports 2026-05-25 (prior session turn): Audit 1 (17 agent SRP) + Audit 2 (workflow phase ownership) + Audit 3 (Anthropic 2026 verbatim)"
research_method: "Inherits the 3 parallel general-purpose subagents output from the superseded SPEC's research.md (§A through §G below, verbatim preserved) AND appends §H Audit 3 synthesis specific to this SPEC's architectural-pivot rationale"
sync_commit_sha: "f0f222fa3"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | orchestrator (synthesizing) | Initial research.md for SPEC-V3R6-AGENT-TEAM-REBUILD-001. §A through §G inherited from `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md` (the predecessor SPEC's research is preserved as the base since the same 3 parallel research agents inform both SPECs). §H NEW — Audit 3 Synthesis (2026-05-25) appends the architectural-pivot rationale specific to this SPEC's 17→8 consolidation strategy. |

---

## §A — Executive Summary

**Mission**: Audit the MoAI agent catalog and workflow orchestration architecture against Anthropic 2026 published guidance, and propose remediation that aligns with the published best-practice ceiling (3-5 teammates) and architectural constraints ("subagents cannot spawn subagents") while preserving the 6-month GEARS backward-compatibility window for existing 88 pre-v3 SPECs.

**Verdict (revised by Audit 3)**: The MoAI orchestrator's current behavior **violates the spirit and letter of Anthropic 2026 published guidance**:
- Catalog inflation 2-5× over the 3-built-in-subagent + 3-5-teammate ceiling (Finding A1)
- Hierarchical fiction (`manager-strategy → manager-develop` chain) violates verbatim Anthropic "subagents cannot spawn other subagents" (Finding A2)
- 12 of 17 agents fail Anthropic's "define when keep spawning the same kind of worker" criterion via 0-invocation empirical evidence (Finding A3)
- Coding-task parallel multi-spawn violates Anthropic's verbatim caveat about LLM real-time coordination weakness (Finding A4)
- Static domain-expert agent files duplicate the Anthropic-recommended per-spawn-prompt specialization pattern (Finding A5)
- Hook-based mechanical enforcement (PostToolUse, Stop, TaskCompleted) is available today and unused (Finding A6)

The predecessor SPEC `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001`'s remediation strategy (restoring 17 phases) is invalidated by Findings A1-A4. **Pivot to consolidation 17→8 + hook enforcement is the correct remediation path.**

**6 Core Findings** (from prior parallel research agents — preserved):

| # | Finding | Severity | Source corroboration |
|---|---------|----------|----------------------|
| 1 | sync-phase quality/security/coverage specialists never invoked | CRITICAL | Research 1 §3 + Research 2 §5 + Research 3 §1 |
| 2 | run-phase manager-strategy → manager-develop hierarchical chain collapsed | HIGH | Research 1 §3 + Research 3 §1 + §2 |
| 3 | Sub-skill on-demand loading bypassed (router skill not actually invoked via Skill()) | HIGH | Research 1 §1 + §6 + Research 2 §2 |
| 4 | 5 HUMAN GATE decision points silently skipped | HIGH | Research 1 §4 + Research 2 §5 + Research 3 §2 |
| 5 | plan-phase Explore subagent + research.md + GitHub Issue + BODP audit trail all missing | MEDIUM | Research 1 §4 + Research 1 §2 |
| 6 | **Autonomous execution mode selection not implemented — defaults to sequential single-spawn always** | HIGH (user-flagged) | All 3 research agents |

**Audit 3 supersedence note (NEW for this SPEC)**: Findings 1-6 above informed the predecessor SPEC's REQ-WOF-001..015 (restoration strategy). Audit 3's Findings A1-A6 (in §H below) **invalidate the restoration strategy** for findings 1, 2, 4 and recommend a different remediation (consolidation + hook enforcement) for those three. Findings 3, 5, 6 are preserved by this SPEC (REQ-ATR-007 Skill router discipline, REQ-ATR-008 Mode Selection logging, REQ-ATR-007 Skill router for paste-ready resume).

---

## §B — Detailed Findings by Source (prior research, preserved verbatim from predecessor)

### B.1 Research Agent 1 — Anthropic Official Docs (HIGH confidence verbatim)

**Skill() invocation pattern** ([claude.com/docs/en/skills](https://code.claude.com/docs/en/skills)):

> "Skills extend what Claude can do. Create a `SKILL.md` file with instructions, and Claude adds it to its toolkit. **Claude uses skills when relevant, or you can invoke one directly with `/skill-name`.**"

> "**full skill content only loads when invoked.** [...] If those instructions reference other files (like FORMS.md or a database schema), Claude reads those files too using additional bash commands."

**Multi-agent parallelism** ([anthropic.com/engineering/built-multi-agent-research-system](https://www.anthropic.com/engineering/built-multi-agent-research-system)):

> "**lead agent spins up 3–5 subagents in parallel rather than serially** ... cut research time by up to 90% for complex queries."

> "Simple fact-finding requires 1 agent with 3–10 tool calls; direct comparisons need 2–4 subagents with 10–15 calls each; complex research uses 10+ subagents with divided responsibilities."

**Built-in subagents**: `Explore`, `Plan`, `general-purpose` are the only officially-named subagents. MoAI's expansive 17-agent catalog is **MoAI-original**; Anthropic prefers minimal taxonomy.

**Critical constraint** ([claude.com/docs/en/sub-agents](https://code.claude.com/docs/en/sub-agents)):

> "**Subagents cannot spawn other subagents.** If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."

**Sequential chain pattern**:

> "For multi-step workflows, ask Claude to use subagents in sequence. **Each subagent completes its task and returns results to Claude, which then passes relevant context to the next subagent.**"

**4-phase canonical pattern** ([claude.com/docs/en/best-practices](https://code.claude.com/docs/en/best-practices)):

1. Explore — read files, no changes (use `Agent(Explore)`)
2. Plan — detailed plan, **press Ctrl+G to open in editor for HUMAN approval** (HUMAN GATE)
3. Implement — code, verify against plan
4. Commit — descriptive message + PR

### B.2 Research Agent 2 — Context7 MCP (HIGH confidence on API, MEDIUM on patterns)

**Task tool API contract** (verbatim from Claude Agent SDK):

```
POST /tools/task
{
  "description": str,     // 3-5 word task description
  "prompt": str,          // The task for the agent
  "subagent_type": str    // Specialized agent type
}
→ { "result": str, "usage": dict, "total_cost_usd": float, "duration_ms": int }
```

**Subagent configuration options**:
- `tools`: tool subset restriction
- `disallowedTools`: explicit denylist
- `model`: per-subagent model override
- `permissionMode`: `default` | `bypassPermissions`
- `isolation: "worktree"`: filesystem isolation for parallel writes
- `background: true`: non-blocking execution

**8 orchestration patterns enumerated**:
1. Sequential execution
2. Parallel processing
3. Pipeline patterns
4. Map-reduce workflows
5. Event-driven coordination
6. Hierarchical delegation (manager → specialist)
7. Consensus mechanisms
8. Failover strategies

**Hook-based enforcement** (verbatim):

> Stop hook test enforcement: *"Review transcript. If code was modified (Write/Edit tools used), verify tests were executed. If no tests were run, block with reason 'Tests must be run after code changes'."*

### B.3 Research Agent 3 — revfactory/harness + Karpathy (verbatim from upstream)

**6 harness canonical orchestration patterns**:
1. Pipeline — Sequential A → B → C → D
2. Fan-out/Fan-in — Parallel dispatch + synthesize
3. Expert Pool — Router → {Expert A | Expert B | Expert C}
4. Producer-Reviewer — Iterative loop with bounded iteration count (`max_iterations: 3`)
5. Supervisor — Central coordinator → workers
6. Hierarchical Delegation — Multi-level (max 2-3); anti-pattern: ">2-3 levels causes coordination overhead"

**harness mandatory HUMAN-IN-LOOP triggers**:
> Phase 0 Audit & Plan Confirmation: "사용자에게 요약 보고하고, 실행 계획을 확인받는다"
> **"Users are harness architects, not passive observers."**

**3 orchestrator templates**: Team-orchestrator, Sub-orchestrator, Hybrid-orchestrator.

**Karpathy 4 Coding Principles**:
1. Think Before Coding
2. Simplicity First
3. Surgical Changes
4. Goal-Driven Execution

**Karpathy 8 anti-patterns mapping to MoAI orchestrator behavior**:

| # | Anti-pattern | MoAI exhibits? |
|---|--------------|----------------|
| 1 | Premature Abstraction | No |
| 2 | Over-Engineering | YES (17-agent catalog mostly unused) |
| 3 | Drive-By Refactoring | YES |
| 4 | Style Drift | No |
| 5 | Silent Assumption | YES |
| 6 | Guessing Over Clarifying | YES |
| 7 | Sycophantic Agreement | YES |
| 8 | Claiming Without Evidence | YES |

---

## §C — User-Flagged Finding #6 (synthesized from Research 1+2+3) — preserved verbatim

5-mode autonomous selection corroboration matrix:

| Mode | Anthropic doc | Context7 doc | harness doc | MoAI usage |
|------|--------------|--------------|-------------|-----------|
| Sequential single-spawn | Sub-agents chain pattern | Sequential execution (#1) | Pipeline pattern (#1) | **100% of recent SPECs** |
| Parallel multi-spawn | "3-5 parallel cut research time 90%" | Parallel processing (#2) | Fan-out/Fan-in (#2) | 0% — used only for research |
| Background non-blocking | — | `background: true` config | — | 0% |
| Single sub-agent | Sub-agents docs | Task tool API | Sub-orchestrator | Used (manager-develop) |
| Agent Teams | — | TeamCreate + SendMessage | Team-orchestrator default | 0% — feature flag exists but never activated |

Decision tree (preserved):

```
Task complexity → mode:
├─ Trivial (typo, single-line) → No agent, direct execution
├─ Single domain, sequential dependency → Sub-agent (1× Agent())
├─ Multi-domain, independent investigations → Parallel multi-spawn (3-5× Agent() single message)
├─ Long-running, can complete async → Background (Agent run_in_background: true)
└─ Real-time coordination needed → Agent Teams (TeamCreate + SendMessage)
```

---

## §D — Unified Recommendations from Prior Research (priority ranked, preserved)

### D.1 CRITICAL (must-fix) — predecessor SPEC referenced

**R1**, **R2**, **R3** — see predecessor research §D.1. **Audit 3 invalidates R2's restoration strategy** for sync-phase 3 specialists; the replacement is hook-based enforcement (REQ-ATR-009 + REQ-ATR-014 in this SPEC).

### D.2 HIGH

**R4** — manager-strategy chain restoration. **Audit 3 INVALIDATES R4** per Finding A2 "subagents cannot spawn subagents". Replaced by: archive manager-strategy + absorb planning role into manager-spec (REQ-ATR-001 + retention matrix §B.3 of this SPEC's spec.md).

**R5** — Skill router invocation discipline. **PRESERVED** as REQ-ATR-007 of this SPEC.

**R6** — Producer-Reviewer cycle with evaluator-active. **PRESERVED** as part of this SPEC's 8-retain matrix (evaluator-active retained, harness-thorough invocation surface clarified).

### D.3 MEDIUM

**R7** — plan-phase Explore. **PRESERVED** — Anthropic built-in `Explore` is retained as the 8th agent in this SPEC's catalog.

**R8** — GitHub Issue + BODP audit trail. **Not in this SPEC's scope** (deferred); existing CONST-V3R5-030..036 BODP HARD rules remain in force.

**R9** — manager-git PR vs Hybrid Trunk doctrine. **PRESERVED** as REQ-ATR-020 of this SPEC.

### D.4 LOW

**R10** — Hook-based enforcement. **PROMOTED to CRITICAL by Audit 3** (per Finding A6). This SPEC implements 3 NEW hook scripts per REQ-ATR-009 + REQ-ATR-014 + Status Transition Ownership PostToolUse.

**R11** — A/B test methodology. **Not in this SPEC's scope** (deferred; future SPEC).

**R12** — Status Transition Ownership PostToolUse hook. **PROMOTED to in-scope** (status-transition-ownership.sh per design.md §D.1).

---

## §E — Limitations of This Research (preserved)

1. **Context7 prompt caching for sub-agent loops**: Low confidence.
2. **Karpathy 8 anti-patterns**: MoAI-derivative expansion, not verbatim from Karpathy.
3. **MoAI vs Anthropic gap on "harness"**: Same word, different concept.
4. **Research time**: 3 parallel agents wall-clock ~5-10 min for prior research.

**NEW limitation (Audit 3)**: Audit 3 verbatim quotes are from 2026-mid Anthropic published guidance. If Anthropic ships new guidance reversing positions, this SPEC's conclusions may need revision (per spec.md §G G8 risk).

---

## §F — Cross-References (preserved)

- Research source 1 (Anthropic official, agentId `a1222604845ca190c`)
- Research source 2 (Context7 MCP, agentId `ac6f961ebbfa62257`)
- Research source 3 (GitHub repos, agentId `a3a149d78fb82a962`)
- Internal prior audit: `.moai/research/anthropic-best-practices-2026-05-24.md`
- User concern: 5-mode autonomous selection (Finding #6)

---

## §G — Next Steps (preserved from predecessor; replaced for this SPEC by §H below)

This section is preserved from the predecessor SPEC research.md for historical continuity. The actual next-steps for this SPEC (SPEC-V3R6-AGENT-TEAM-REBUILD-001) are documented in §H below.

---

## §H — Audit 3 Synthesis (NEW for SPEC-V3R6-AGENT-TEAM-REBUILD-001, 2026-05-25)

### §H.1 Three deep audits conducted in prior session turn

In the immediate prior session turn (2026-05-25), three deep audits were conducted to assess the MoAI agent catalog and workflow orchestration architecture against Anthropic 2026 published guidance:

**Audit 1 — 17 agent SRP audit (Single Responsibility Principle)**: 19 agents reviewed (including 2 dormant/legacy). Per-agent SRP scoring (4 dimensions: scope clarity, tool whitelist precision, NOT-for declaration, single-domain focus). Verdict: 13 of 19 score ≥ 0.85 SRP; 12 of 17 active agents recorded **0 invocations across the recent 4-SPEC cohort** (`SPEC-V3R6-SKILL-GEARS-ALIGN-001`, `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`, `SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001`, `SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001`). Per-agent SRP healthy, system-level utilization broken.

**Audit 2 — Workflow phase ownership audit**: 23 workflow router skills + 10 sub-skill modules audited (13K LOC total). Result: **0/45 phases have explicit owner declarations**; 0/6 HUMAN GATE decision points mechanically enforced; team mode (`workflow.team.enabled` + `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`) is a dead variant (never invoked). AMBER assessment.

**Audit 3 — Anthropic 2026 verbatim audit**: comparison of MoAI agent catalog + workflow architecture against published Anthropic guidance as of 2026-mid. 6 critical findings A1-A6:

### §H.2 Audit 3 Finding A1 — Catalog inflation 2-5× over Anthropic ceiling

Anthropic ships exactly **3 built-in subagents**: `Explore`, `Plan`, `general-purpose`.

Anthropic Agent Teams documentation states verbatim:

> *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."*
>
> — `claude.com/docs/en/agent-teams`

**Quantitative gap**: MoAI ships 17 custom agents (8 manager-* + 6 expert-* + 1 builder-* + 2 evaluator/audit). Vs Anthropic 3-5 teammate ceiling: **2-5× over recommendation**.

**Empirical cost**: 12 of 17 agents recorded 0 invocations across recent 4-SPEC cohort (Audit 1 data).

**Implication for this SPEC**: REQ-ATR-001 — consolidate to exactly 8 retained agents (7 MoAI-custom + 1 Anthropic built-in `Explore`).

### §H.3 Audit 3 Finding A2 — Hierarchical fiction architecturally impossible

Anthropic Sub-agents documentation states verbatim:

> *"Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."*
>
> — `claude.com/docs/en/sub-agents`

**Implication**: the MoAI `manager-strategy → manager-develop` chain pattern documented in `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` REQ-WOF-003 + REQ-WOF-006 is **architecturally impossible to execute at runtime**. Restoration is not a viable remediation; the chain cannot exist.

**Implication for this SPEC**: REQ-ATR-001 (archive `manager-strategy`) + spec.md §B.3 retention matrix absorbs planning role into `manager-spec`. The "chain" is replaced by orchestrator-mediated context passing between two L1-spawned sub-agents (manager-spec → manager-develop via orchestrator).

### §H.4 Audit 3 Finding A3 — Phantom agent criterion failure

Anthropic best-practices guidance states verbatim:

> *"Define a custom subagent when you keep spawning the same kind of worker with the same instructions."*
>
> — `claude.com/docs/en/best-practices`

**Implication**: the criterion is empirical — frequent spawning of the same worker. 12 of MoAI's 17 agents fail this criterion outright with 0 invocations across recent 4 SPEC sessions:

- `manager-strategy` — 0 invocations
- `manager-quality` — 0 invocations
- `manager-brain` — 0 invocations
- `manager-project` — 0 invocations
- `claude-code-guide` — 0 invocations
- `researcher` — 0 invocations
- `expert-backend` — 0 invocations
- `expert-frontend` — 0 invocations
- `expert-security` — 0 invocations
- `expert-devops` — 0 invocations
- `expert-performance` — 0 invocations
- `expert-refactoring` — 0 invocations

The retained 5 (`manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`, `manager-git`) plus 2 dormant-but-justified (`evaluator-active`, `builder-harness`) plus Anthropic built-in (`Explore`) form the consolidation target of 8 retained.

**Implication for this SPEC**: REQ-ATR-005 — archive the 12 phantom agents to `.moai/backups/agent-archive-2026-05-25/` (preserve git history via `git mv`); not delete.

### §H.5 Audit 3 Finding A4 — Coding-task parallelism caveat

Anthropic verbatim:

> *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."*
>
> — `anthropic.com/engineering/built-multi-agent-research-system`

**Implication**: MoAI workflow's run-phase coding tasks SHOULD remain single-agent (`manager-develop` standalone with cycle_type=ddd/tdd/autofix). Parallel multi-spawn applies primarily to research phase (Explore + analyst + architect role profiles), not to implementation.

**Implication for this SPEC**: design.md §C.2 Run-phase pattern map confirms Phase 1 implementation = Mode 5 SUB-AGENT (single manager-develop). This SPEC's own Phase 0.95 Mode Selection (plan.md §D.3) chooses Mode 5 per this caveat.

### §H.6 Audit 3 Finding A5 — Domain expertise via spawn-prompt, not via agent file

Anthropic's recommended pattern for domain expertise is per-spawn parameter injection at delegation time:

```python
Agent(
    subagent_type: "general-purpose",
    model: "opus",
    tools: ["Read", "Write", "Edit", "Bash", "Grep"],  # domain whitelist
    prompt: "<domain-specific instructions injected at spawn time>"
)
```

**Implication**: the 6 `expert-*` agents (backend, frontend, security, devops, performance, refactoring) duplicate this pattern as static agent files in `.claude/agents/expert/`. This violates Anthropic's recommended dynamic pattern in 2 ways:

(a) **Static knowledge vs active context**: domain knowledge embedded in static agent definition files cannot adapt to the current conversation's specific scope; per-spawn prompt injection adapts naturally.

(b) **0-invocation empirical evidence**: combined with Finding A3, the 6 `expert-*` agents fail "define when keep spawning" because they ARE NOT being spawned. Active spawning would adopt the per-spawn pattern, not the static file.

**Implication for this SPEC**: REQ-ATR-005 archive the 6 expert-* files; design.md §B.5 ARCHIVED_AGENT_REJECTED Migration Table documents the per-spawn-prompt replacement pattern with concrete tool whitelists per domain.

### §H.7 Audit 3 Finding A6 — Hook-based enforcement available

Anthropic hook documentation (`claude.com/docs/en/hooks`) describes the following canonical hook types:

- **PostToolUse**: fires after every tool invocation; can inspect tool name + input + output
- **Stop**: fires at session end or commit completion; can block via exit 2
- **SubagentStop**: fires after a sub-agent invocation completes
- **TaskCompleted**: fires when team-mode TaskList task is marked complete

**Implication**: orchestrator-discipline phantom-agent spawn patterns (e.g., "manager-quality MUST run sync-phase lint+test+coverage") can be replaced with **mechanically-unbypassable hook-based enforcement**:

- `manager-quality` sync-phase lint+test+coverage → **Stop hook** (`sync-phase-quality-gate.sh`)
- `expert-security` dependency manifest audit → **Stop hook** (same script, REQ-ATR-014 conditional logic on `go.mod`/`package-lock.json` changes)
- Status Transition Ownership Matrix enforcement → **PostToolUse hook** (`status-transition-ownership.sh`)
- Team-mode per-AC PASS evidence → **TaskCompleted hook** (`team-ac-verify.sh`, dormant default)

**Implication for this SPEC**: REQ-ATR-009 + REQ-ATR-014 + Status Transition Ownership PostToolUse hook (3 NEW hook scripts in plan.md §D.7). Design.md §D Hook Architecture documents the implementation.

### §H.8 Synthesis — pivot rationale

The three audits combined establish that:

1. **Catalog over-engineering** (Audit 1: 12/17 phantom agents) + **architectural fiction** (Audit 3 Finding A2: hierarchical chain impossible) + **published-guidance gap** (Audit 3 Finding A1: 2-5× ceiling overage) → the predecessor SPEC's "restore 17 phases" remediation is invalid.

2. **The Anthropic-aligned remediation** is consolidation 17→8 retained + 12 archived + hook-based enforcement for the 4 highest-value mechanical enforcement points (Stop sync-phase quality + Stop dependency manifest + PostToolUse Status Transition Ownership + TaskCompleted team-mode AC verify).

3. **The pivot supersedes** the predecessor SPEC `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` because the original SPEC's foundational assumption (restoration is the path forward) is invalidated by Audit 3 Findings A1-A4. Per Status Transition Ownership Matrix, the `* → superseded` transition is owned by manager-spec when authoring the new superseding SPEC.

4. **Backward-compatibility preserved**: the 6-month GEARS backward-compatibility window for 88 pre-v3 SPECs is unchanged. This SPEC modifies agent catalog + workflow router declarations + hook scripts only; SPEC body content of pre-v3 SPECs is not touched.

### §H.9 Cross-references to spec.md / plan.md / acceptance.md / design.md

- **spec.md §B Background** — cites Findings A1-A6 verbatim
- **spec.md §B.3 Retain-vs-archive matrix** — 8 retain + 12 archive with per-agent rationale
- **spec.md §D 20 REQ-ATR-XXX** — consolidation + hook + supersedence + doctrine reconciliation REQs
- **plan.md §D 8 milestones M1-M8** — implementation decomposition
- **acceptance.md §A 22 AC-ATR-XXX** — binary observable evidence per REQ
- **design.md §B Target Architecture** — 8-agent delegation graph + 2-level Anthropic constraint + 4-layer enforcement + 5-mode decision tree + ARCHIVED_AGENT_REJECTED migration table
- **design.md §D Hook Architecture** — 3 NEW hook scripts implementation guidance
- **design.md §E Karpathy 8 anti-pattern remediation matrix** — 7/8 anti-patterns addressed

---

Version: 0.1.0 (initial synthesis — §A-G preserved from predecessor research.md verbatim; §H NEW Audit 3 synthesis specific to SPEC-V3R6-AGENT-TEAM-REBUILD-001 architectural-pivot rationale)
