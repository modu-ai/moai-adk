---
id: SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
title: "Purpose-driven model+effort selection for dynamic Workflow agent() calls, Agent Teams role_profiles, and session-handoff ultracode re-set"
version: "0.2.0"
status: completed
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/workflows,.moai/config/sections,.claude/rules/moai/workflow,.claude/output-styles/moai"
lifecycle: spec-anchored
tags: "workflow, effort, model-policy, agent-teams, session-handoff, doctrine, ultracode"
depends_on: []
related_specs: [SPEC-CC2178-MODEL-POLICY-REPAIR-001]
---

# SPEC-V3R6-WORKFLOW-EFFORT-MAP-001

## §A. Problem Statement

Claude Code exposes two independent purpose-routing levers — `model` and `effort` — on three distinct MoAI surfaces that currently have **no unified SSOT** mapping "agent purpose → (model, effort)":

1. **Dynamic Workflow `agent()` calls** (`.claude/workflows/*.js`). The official `agent()` opt schema is `{model, effort, agentType, isolation, phase, schema, label}` (per `https://code.claude.com/docs/en/workflows`). Omitting `model` inherits the main-loop model; omitting `effort` inherits the session effort. The sole shipped workflow script (`codemaps-extract.js`) sets `agentType: 'Explore'` but omits both `model` and `effort`, so each per-package read-only extraction agent inherits whatever the session happens to be at — typically `xhigh` after an `ultrathink.` opener. Per the official effort guidance (`https://platform.claude.com/docs/en/build-with-claude/effort`), `low` is the **officially-recommended effort for subagents** ("Simpler tasks that need the best speed and lowest costs, such as subagents"). Running a mechanical read-only extraction at `xhigh` is a silent cost leak.

2. **Agent Teams `role_profiles`** (`.moai/config/sections/workflow.yaml`). Each of the 7 profiles (analyst/architect/designer/implementer/researcher/reviewer/tester) carries `mode`, `model`, `isolation` — but **no `effort` field**. Without an `effort` declaration, role profiles cannot express that a read-only `researcher` should run at `low`/`medium` while an `architect` should run at `xhigh`. The YAML is consumed by the orchestrator (LLM) and workflow-script authors at runtime; no Go code reads it today (Go change is explicitly out of scope per decision #1).

3. **Session-handoff resume message Block 1** (`.claude/rules/moai/workflow/session-handoff.md`). The `ultrathink.` opener restores reasoning effort only. The dynamic-workflows.md rule already states (line 74) that session-wide `/effort ultracode` mode is **NOT** restored by `ultrathink.` and must be explicitly re-issued after `/clear` when the next SPEC needs workflow fan-out. But this guidance lives in the dynamic-workflows rule, not in the handoff template itself — so the orchestrator emits a resume that silently drops the ultracode re-set line even when the next SPEC is a workflow-heavy one.

These three gaps share a common root cause: **there is no canonical "agent purpose → (model, effort)" taxonomy** that workflow-script authors, role_profile authors, and the resume-message emitter can all consult. This SPEC creates that SSOT and wires it into all three surfaces.

### §A.1 Relationship to existing dual policy (do NOT duplicate)

MoAI already operates two purpose-routing policies, neither of which governs the surfaces this SPEC targets:

| Existing policy | Surface | Source file | What it routes |
|-----------------|---------|-------------|----------------|
| `agentModelMap` | **Claude Code subagent frontmatter** (`settings.json`) — the retained MoAI agents (manager-spec / manager-develop / manager-docs / manager-git / builder-harness) | `internal/template/model_policy.go` | 3-tier high/medium/low policy → model per agent |
| `agentEffortMap` | Same surface as above | `internal/template/model_policy.go` | Explicit effort per reasoning-heavy agent (manager-spec/manager-develop/plan-auditor/sync-auditor = xhigh; builder-harness = high) |
| harness → effort | CLI route surface | `internal/harness/router/effort.go` | harness level (minimal/standard/thorough) → effort (medium/high/xhigh) |

The MoAI-custom agents and the Agent Teams role profiles are **disjoint sets**: `agentModelMap` covers 5 retained named MoAI agents; `role_profiles` covers 7 generic team roles (analyst/architect/designer/implementer/researcher/reviewer/tester). The two maps do not overlap. A workflow `agent()` call with `agentType: 'Explore'` is a third surface again — it spawns a generic subagent that is neither a retained MoAI agent nor a role-profiled teammate.

This SPEC does NOT modify the existing dual policy (decision #1: Go source under `internal/` stays untouched). It creates a **declarative companion SSOT** for the three surfaces the dual policy does not reach.

## §B. Scope Decisions (binding user decisions)

Per the confirmed user decisions, this SPEC operates under the following binding constraints:

| Decision | Constraint | Rationale |
|----------|-----------|-----------|
| #1 DEPTH | Doctrine + script template only. **Go source under `internal/` MUST NOT be modified** — `internal/template/model_policy.go` and `internal/harness/router/effort.go` stay untouched. | The declarative SSOT is consumed at runtime by the orchestrator (LLM) and by workflow-script authors; no Go code reads it. Wiring it into Go is a separate future SPEC. |
| #2 MAPPING COVERAGE | Unified SSOT covering BOTH (a) Claude Code dynamic Workflow `agent()` opts AND (b) Agent Teams `role_profiles`. Add an `effort` field to all 7 `role_profiles`. Document the SAME purpose→model+effort mapping for `.claude/workflows/*.js`. | The two surfaces are disjoint from the existing dual policy and from each other; a single taxonomy prevents per-surface drift. The YAML `effort` field is declarative — the tension with #1 resolves because no Go code reads it. |
| #3 HANDOFF | Resume message Block 1 gets a purpose-conditional ultracode line. `ultrathink.` is always present; a `/effort ultracode` re-set line appears ONLY when the next SPEC needs workflow fan-out. | Formalizes (into the handoff itself) what dynamic-workflows.md already states: ultracode is NOT restored by `ultrathink.` and must be re-set after `/clear`. |
| #4 PATH | Plan-phase only → plan-auditor independent audit → Implementation Kickoff Approval (human gate) → `/moai run`. | Standard V3R6 4-phase lifecycle. |

### §B.1 SPEC ID rationale

**SPEC ID: `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`** — NOT `SPEC-CC2178-WORKFLOW-EFFORT-MAP-001`.

This SPEC is not a dependency-follow-on to `SPEC-CC2178-MODEL-POLICY-REPAIR-001`. It does not modify the Go model policy code that MODEL-POLICY-REPAIR-001 owns. It is a **sibling doctrine SPEC** creating an SSOT for a different surface set (workflow scripts + role_profiles + handoff template). The `CC2178` namespace tracks SPECs whose identity is the CC 2.1.178 alignment audit lineage; this SPEC's lineage is the V3R6 doctrine layer. CC version compatibility is *cited* (effort field since CC v2.1.110; `agent()` opts documented at the CC workflows page) but is not the SPEC's identity.

### §B.2 Tier rationale

**Tier M** (not S). Five surfaces are touched — dynamic-workflows.md (SSOT + template mirror), session-handoff.md (SSOT + template mirror), output-styles/moai/moai.md §8 (+ template mirror), workflow.yaml (local + template), codemaps-extract.js — with byte-parity invariants on three mirror pairs and a taxonomy table that must be internally consistent across all surfaces. Tier S would under-scope the parity-verification surface. Tier M is the minimum tier that covers the multi-surface doctrine + mirror-parity + YAML + JS scope.

## §C. Ground Truth (investigated — cite, do NOT re-derive)

| Claim | Verbatim evidence | Source |
|-------|-------------------|--------|
| Effort levels: low/medium/high(default)/xhigh/max | "Effort levels: low, medium, high (default), xhigh, max" | `https://platform.claude.com/docs/en/build-with-claude/effort` |
| `low` is the officially-recommended effort for subagents | "Simpler tasks that need the best speed and lowest costs, such as subagents" | same |
| `xhigh` is the coding/agentic starting point for Opus 4.7/4.8 | "xhigh for coding/agentic work" | `.claude/rules/moai/development/prompting-best-practices.md` § Thinking & Reasoning |
| `ultracode` = xhigh effort + standing permission to launch multi-agent workflows — NOT an additional API effort level, Claude Code menu only | "the session-wide `/effort ultracode` mode, which combines `xhigh` reasoning with automatic workflow orchestration" | `.claude/rules/moai/workflow/dynamic-workflows.md` line 74 |
| The `agent()` call accepts opts `{model, effort, agentType, isolation, phase, schema, label}` | (documented `agent()` signature) | `https://code.claude.com/docs/en/workflows` |
| Omitting `model` inherits the main-loop model; omitting `effort` inherits session effort | (inheritance semantics) | same |
| MoAI dual policy already exists at `internal/template/model_policy.go` | `agentModelMap` (5 retained agents) + `agentEffortMap` (manager-spec/manager-develop/plan-auditor/sync-auditor=xhigh; builder-harness=high); renders to Claude Code subagent frontmatter — does NOT govern dynamic Workflow `agent()` calls | `internal/template/model_policy.go` lines 66–93, 197–223 |
| `internal/harness/router/effort.go` maps harness → effort (minimal→medium, standard→high, thorough→xhigh) | "REQ-HRN-001-005: minimal->medium, standard->high, thorough->xhigh" | `internal/harness/router/effort.go` lines 14–18 |
| `workflow.yaml` `role_profiles` has model/mode/isolation per role but NO effort field | 7 profiles, each with `mode`/`model`/`isolation`/`description` — no `effort` key | `.moai/config/sections/workflow.yaml` lines 80–115 |
| `.claude/agents/moai/*.md` frontmatter already declares model+effort+isolation per agent | manager-git=haiku/medium, manager-docs=haiku/medium, manager-develop=inherit/xhigh/worktree, sync-auditor=inherit/xhigh, plan-auditor=inherit/xhigh, manager-spec=inherit/xhigh, builder-harness=inherit/high | `.claude/agents/moai/*.md` frontmatter |
| `.claude/workflows/codemaps-extract.js` sets `agentType:'Explore'` but does NOT set model/effort | `agent(PROMPT(pkg), { label: ..., phase: 'Extract', agentType: 'Explore' })` — no `model` or `effort` keys | `.claude/workflows/codemaps-extract.js` line 53 |
| `.claude/rules/moai/workflow/session-handoff.md` Block 1 has `ultrathink.` opener only — no ultracode line | Block 1: `ultrathink. <SPEC-ID> <phase> <entering verb>.` | `.claude/rules/moai/workflow/session-handoff.md` line 32 |
| dynamic-workflows.md already states ultracode is NOT restored by `ultrathink.` | "Because it resets on a new session, `ultracode` is **not** restored by the `ultrathink.` opener of a paste-ready resume message" | `.claude/rules/moai/workflow/dynamic-workflows.md` line 74 |
| dynamic-workflows.md / session-handoff.md / output-styles/moai.m.md are byte-parity between SSOT and template mirror (0 diff today) | `diff <SSOT> <template>` returns 0 lines for all 3 pairs | baseline measurement, 2026-06-17 |
| Mirror-parity CI enforcement is SPLIT across two mechanisms | (a) `internal/template/rule_template_mirror_test.go` allowlist covers `.claude/rules/moai/workflow/session-handoff.md` ONLY (file:10KB, allowlist lines 47/51); (b) `dynamic-workflows.md` and `output-styles/moai/moai.md` are NOT in the allowlist — their parity is AC-enforced (AC-WEM-009b), not CI-enforced | `internal/template/rule_template_mirror_test.go` lines 47, 51 (`ls`: `embedded_mirror_test.go` does NOT exist — phantom citation in the iter-0 draft; also cited pre-existing in `internal/template/CLAUDE.md` L23/L31 as a separate doc bug, out of scope — see §H Exclusions) |

## §D. The Purpose Taxonomy (heart of the SSOT)

This taxonomy is the single source that drives workflow.yaml `effort` field values, `.claude/workflows/*.js` authoring guidance, AND the codemaps-extract.js fix. Every recommended effort value is grounded in a citation to the official effort guidance.

| Purpose | Example surfaces | Recommended model | Recommended effort | Official citation |
|---------|------------------|-------------------|--------------------|-------------------|
| **read-only-extract** | codemaps per-package dep-graph + public-surface extraction; mechanical AST/grep sweeps | haiku | **low** | "`low` — Simpler tasks that need the best speed and lowest costs, such as subagents" |
| **mechanical-transform** | large migrations (call-site rename, API shape change); mechanical refactors | sonnet | **medium** | "`medium` — Balanced reasoning for general tasks" |
| **synthesize** | architectural synthesis layered on top of deterministic extraction (codemaps insight); multi-source research synthesis | sonnet | **high** | "`high` — Most tasks; good balance of quality and speed" |
| **research** | cross-checked research with adversarial voting; deep single-topic investigation | sonnet or opus | **high** or **xhigh** | `high`/`xhigh` — (project-internal rationale, NOT a verbatim citation: research effort should scale with the density of falsifiable claims the research must adjudicate; the official effort doc lists only the 5 levels and their general-purpose descriptions, it does not prescribe a per-task research-depth heuristic) |
| **verify-judge** | code review (security/perf/arch dimensions); plan-auditor independent audit; sync-auditor quality scoring | sonnet or opus | **xhigh** | "minimum `high` for intelligence-sensitive work" — reviewer/judge is intelligence-sensitive |
| **implement** | code generation (backend/frontend/full-stack); test writing | sonnet or opus | **xhigh** | "`xhigh` for coding/agentic work" |
| **design-architecture** | solution architecture decisions; system design; deep reasoning over trade-offs | opus | **xhigh** | "`xhigh` for coding/agentic work" + "minimum `high` for intelligence-sensitive work" |
| **orchestrator-main-loop** | the orchestrator itself (NEVER a workflow agent or teammate — listed for completeness) | opus (1M) | **xhigh** (via `ultrathink.`) or **ultracode** (session-wide, when workflow fan-out is the session shape) | "`ultracode` combines `xhigh` reasoning with automatic workflow orchestration" |

**Reading order.** When a workflow agent or teammate serves multiple purposes, pick the highest-effort purpose in the table. When purpose is ambiguous, prefer the cheaper effort — the cost of over-efforting a read-only extraction is a silent token leak; the cost of under-efforting a verify-judge is a missed defect.

**Exclusion: orchestrator-main-loop row.** The orchestrator row exists only to make the taxonomy exhaustive. The orchestrator is never spawned via `agent()` and is never a teammate. Its effort is controlled by `ultrathink.` (keyword) or `/effort ultracode` (session mode), NOT by any workflow-script or role_profile declaration. The session-handoff Block 1 (§F below) is the only place this SPEC touches orchestrator effort.

## §E. Requirements (GEARS)

### REQ-WEM-001 (Ubiquitous) — SSOT existence
The MoAI-ADK workflow doctrine SHALL maintain a single canonical purpose→(model,effort) taxonomy table that governs dynamic Workflow `agent()` calls, Agent Teams `role_profiles`, and the session-handoff ultracode re-set decision.

### REQ-WEM-002 (Where) — Workflow agent() explicit-purpose rule
**Where** a `.claude/workflows/*.js` script invokes `agent()`, the script author SHALL set `effort` explicitly per the §D taxonomy rather than inheriting the session default.

### REQ-WEM-003 (When) — codemaps-extract.js per-stage effort
**When** the codemaps-extract.js `Extract` phase invokes one Explore agent per package, the `agent()` call SHALL set `effort: 'low'` (read-only-extract purpose) and the authoring guidance SHALL document this as the canonical pattern for mechanical read-only fan-out.

### REQ-WEM-004 (Where) — role_profiles effort field
**Where** `workflow.yaml` declares a `role_profiles` entry, the entry SHALL carry an `effort` key whose value is drawn from the §D taxonomy and the official `{low, medium, high, xhigh, max}` enum.

### REQ-WEM-005 (Ubiquitous) — role_profiles 7-key completeness
The `role_profiles` map SHALL declare `effort` for all 7 canonical roles (analyst, architect, designer, implementer, researcher, reviewer, tester) in BOTH the local workflow.yaml and the template mirror.

### REQ-WEM-006 (Where) — declarative-only YAML
**Where** the `effort` field is added to `role_profiles`, the field SHALL be declarative metadata consumed by the orchestrator (LLM) and workflow-script authors at runtime; no Go code under `internal/` SHALL read the `effort` field (Go consumption is explicitly deferred to a future SPEC).

### REQ-WEM-007 (When) — session-handoff ultracode conditional line
**When** the orchestrator emits a paste-ready resume message and the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams), the resume message Block 1 SHALL carry a `/effort ultracode` re-set line immediately after the `ultrathink.` opener.

### REQ-WEM-008 (When) — session-handoff ultracode omission
**When** the orchestrator emits a paste-ready resume message and the next SPEC does NOT declare workflow fan-out, the resume message Block 1 SHALL omit the `/effort ultracode` line, retaining only the `ultrathink.` opener.

### REQ-WEM-009 (Where) — byte-parity invariant on mirror pairs
**Where** this SPEC edits dynamic-workflows.md, session-handoff.md, or output-styles/moai/moai.md, the corresponding template mirror at `internal/template/templates/.claude/...` SHALL be edited in the same commit so that byte-parity is preserved. Parity is enforced by a SPLIT mechanism: (a) session-handoff.md parity IS CI-enforced by `internal/template/rule_template_mirror_test.go` (the real mirror test; its allowlist covers `.claude/rules/moai/workflow/session-handoff.md` only); (b) dynamic-workflows.md and output-styles/moai/moai.md parity are NOT currently CI-enforced (the allowlist does not cover them) and MUST therefore be verified by an AC based on `diff`/`cmp` between the two trees (AC-WEM-009b). Extending the Go allowlist to cover (b) would require a Go change and is explicitly deferred to a future SPEC per decision #1.

### REQ-WEM-010 (Where) — dynamic-workflows.md taxonomy section
**Where** the dynamic-workflows.md rule already documents the `agent()` primitive, the rule SHALL add a "Purpose-driven model+effort selection" subsection that (a) cites the official effort guidance, (b) reproduces the §D taxonomy or cross-references the SSOT location, and (c) gives the codemaps-extract.js pattern as a worked example.

### REQ-WEM-011 (Where) — session-handoff Block 1 field-spec update
**Where** the session-handoff.md `Field-by-Field Specification` describes Block 1, the spec SHALL add a sub-bullet documenting the purpose-conditional `/effort ultracode` re-set line, with explicit reference to the dynamic-workflows.md doctrine that ultracode is NOT restored by `ultrathink.`.

### REQ-WEM-012 (Where) — output-styles render-surface parity
**Where** `.claude/output-styles/moai/moai.md` §8 carries the render-surface paste-ready template and the pre-emit self-check, the render surface SHALL be updated to match the session-handoff.md SSOT edit for Block 1, preserving the bidirectional parity contract documented in session-handoff.md `Cross-references`.

### REQ-WEM-013 (Ubiquitous) — template neutrality
The template-mirror edits under `internal/template/templates/` SHALL contain NO internal SPEC IDs, NO REQ tokens, NO commit SHAs, NO `feedback_` refs, and NO Audit citations — only generic doctrine prose and public-source citations (per `.moai/docs/template-internal-isolation-doctrine.md`).

### REQ-WEM-014 (When) — scope-discipline exclusion
**When** this SPEC is implemented, the implementer SHALL NOT absorb the 5 pre-existing modified `.moai/config/sections/*.yaml` files or the untracked `.moai/{design,docs,reports}/...` entries from the dirty working tree — those are unrelated pre-existing changes outside this SPEC's scope.

## §F. Constraints

1. **No Go code changes.** `internal/template/model_policy.go`, `internal/harness/router/effort.go`, and the router package stay untouched. The `effort` YAML field is declarative only (REQ-WEM-006).
2. **No `make build` required** for the doctrine/YAML/JS/MD surfaces (the surfaces are not embedded via `go:embed`; `embedded.go` is regenerated only when `internal/template/templates/` content changes — which this SPEC does touch, but the mirror parity test reads the files directly).
3. **Byte-parity on 3 mirror pairs.** dynamic-workflows.md, session-handoff.md, output-styles/moai/moai.md each exist as `.claude/...` SSOT + `internal/template/templates/.claude/...` template mirror. Both sides MUST be edited in the same commit (REQ-WEM-009).
4. **Template neutrality.** `internal/template/templates/` content is user-distributable; it MUST NOT leak MoAI-internal SPEC IDs, REQ tokens, commit SHAs, `feedback_` refs, or Audit citations (REQ-WEM-013, `.moai/docs/template-internal-isolation-doctrine.md`).
5. **Session-handoff diet constraints.** The ultracode re-set line is a SINGLE line (`/effort ultracode`) immediately after `ultrathink.` — it MUST NOT carry history/lesson/directive-escalation prose (per session-handoff.md § Diet Constraints, AP-D-002/AP-D-004).
6. **7-role canonical set locked.** role_profiles keys are schema-locked to analyst/architect/designer/implementer/researcher/reviewer/tester during v3.0.x (per team-protocol.md). This SPEC does NOT add or remove role keys.

## §G. Acceptance Criteria Matrix (summary — full scenarios in acceptance.md)

| AC ID | REQ | Severity | Description |
|-------|-----|----------|-------------|
| AC-WEM-001 | REQ-WEM-001 | MUST | SSOT taxonomy table exists in dynamic-workflows.md and is cited by all 3 downstream surfaces |
| AC-WEM-002 | REQ-WEM-002 | MUST | dynamic-workflows.md carries the "set effort explicitly" rule for workflow `agent()` calls |
| AC-WEM-003 | REQ-WEM-003 | MUST | codemaps-extract.js sets `effort: 'low'` on the per-package `agent()` call (grep-verifiable) |
| AC-WEM-004 | REQ-WEM-004, 005 | MUST | workflow.yaml (local) role_profiles carry `effort` key for all 7 roles |
| AC-WEM-005 | REQ-WEM-004, 005 | MUST | workflow.yaml (template mirror) role_profiles carry `effort` key for all 7 roles — byte-aligned values with local |
| AC-WEM-006 | REQ-WEM-006 | MUST | No Go file under `internal/` is modified by this SPEC's commits (git-diff PRIMARY; struct-field-absence is supporting evidence only — the iter-0 grep-negative was vacuous because `RoleProfileEntry` has no `Effort` field) |
| AC-WEM-007 | REQ-WEM-007, 008 | MUST | session-handoff.md Block 1 field-spec documents the purpose-conditional `/effort ultracode` line |
| AC-WEM-008 | REQ-WEM-007, 008 | MUST | output-styles/moai/moai.md §8 render surface carries the same Block 1 update (byte-parity) |
| AC-WEM-009a | REQ-WEM-009 | MUST | session-handoff.md mirror pair is byte-parity post-edit (CI-enforced via `rule_template_mirror_test.go` allowlist) |
| AC-WEM-009b | REQ-WEM-009 | MUST | dynamic-workflows.md AND output-styles/moai/moai.md mirror pairs are byte-parity post-edit (AC-enforced via `diff` — NOT in the Go allowlist; CI-enforcement deferred to a future SPEC because extending the allowlist would require a Go change, out of scope per decision #1) |
| AC-WEM-010 | REQ-WEM-010 | MUST | dynamic-workflows.md has a "Purpose-driven model+effort selection" subsection |
| AC-WEM-011 | REQ-WEM-013 | MUST | `internal/template/templates/` content under the edited paths contains no internal SPEC IDs / REQ tokens / SHAs / `feedback_` refs / Audit citations |
| AC-WEM-012 | REQ-WEM-014 | MUST | The 5 pre-existing modified `.moai/config/sections/*.yaml` files are NOT absorbed into this SPEC's commits |

## §H. Exclusions (What NOT to Build)

### Out of Scope — doctrine+config+script only (no Go, no enforcement, no absorption)

The following are explicitly out of scope for this SPEC:

- **NO Go source changes.** `internal/template/model_policy.go`, `internal/harness/router/effort.go`, the harness router package, and any Go code that would consume the new `effort` YAML field are explicitly deferred to a future SPEC. Wiring the declarative YAML into Go runtime enforcement is a separate concern.
- **NO Agent Teams runtime enforcement code.** The spawn wrapper (`internal/cli/agent_lint.go`, spawn validation) is NOT modified to enforce the `effort` field. The field is declarative metadata for the orchestrator and workflow-script authors.
- **NO new dynamic Workflow scripts.** This SPEC edits the ONE existing script (`codemaps-extract.js`) to add `effort`; it does not author new workflow scripts.
- **NO new Agent Teams role keys.** The 7-role canonical set is schema-locked (team-protocol.md); this SPEC adds an `effort` field to existing keys only.
- **NO absorption of pre-existing dirty-tree changes.** The 5 modified `.moai/config/sections/*.yaml` files and the untracked `.moai/{design,docs,reports}/...` entries are unrelated to this SPEC and MUST NOT be absorbed (REQ-WEM-014).
- **NO CHANGE to the existing dual policy.** `agentModelMap` / `agentEffortMap` / harness→effort mapping stay exactly as they are; this SPEC is a companion SSOT for a disjoint surface set.
- **NO mid-run user prompts inside workflows.** The AskUserQuestion boundary holds for workflow agents (dynamic-workflows.md already states this); the ultracode re-set line in the handoff is a paste-ready instruction to the user, not a workflow-internal prompt.
- **NO fix for the pre-existing phantom-CI-file citation in `internal/template/CLAUDE.md`.** That file's "Key Patterns" section cites `internal/template/embedded_mirror_test.go` at L23 and L31 — a file that does NOT exist (the real file is `rule_template_mirror_test.go`). This is a SEPARATE pre-existing doc bug, independent of this SPEC's iter-0 phantom citation, and is out of scope here. It should be fixed in a follow-up doc-cleanup SPEC (the fix is a 2-line edit replacing `embedded_mirror_test.go` → `rule_template_mirror_test.go` in CLAUDE.md).

## §I. History

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-06-17 | manager-spec | Initial plan-phase draft. Ground truth pre-verified (mirror parity 0/0/0; ultracode-not-restored doctrine at dynamic-workflows.md L74; codemaps-extract.js L53 omits effort; workflow.yaml role_profiles no effort key). Tier M per §B.2. SPEC ID V3R6 (not CC2178) per §B.1. |
| 0.2.0 | 2026-06-17 | manager-spec | iter-1 remediation (plan-auditor FAIL 0.76 → target ≥0.80). D1: phantom CI file `embedded_mirror_test.go` → real `rule_template_mirror_test.go` (4 citation sites + §C ground-truth row + §H exclusion for sibling CLAUDE.md bug); AC-WEM-009 split into 009a (CI-enforced, session-handoff) / 009b (AC-enforced, dynamic-workflows + output-styles). D2: §D "research" row claim-density rationale relabeled as project-internal (not verbatim). D3..D8 in plan.md/acceptance.md. |

## §J. Cross-References

- `.claude/rules/moai/workflow/dynamic-workflows.md` — workflow primitive doctrine; already states ultracode-not-restored at L74; this SPEC adds the taxonomy section + codemaps pattern.
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume SSOT; this SPEC adds Block 1 ultracode conditional line.
- `.claude/output-styles/moai/moai.md` §8 — render surface for the resume template; bidirectional parity with session-handoff.md.
- `.moai/config/sections/workflow.yaml` — `role_profiles` source of truth (7 roles); this SPEC adds `effort` field.
- `internal/template/templates/.moai/config/sections/workflow.yaml` — template mirror; `effort` field must land byte-aligned.
- `.claude/workflows/codemaps-extract.js` — sole shipped workflow script; this SPEC sets `effort: 'low'` on the per-package `agent()` call.
- `internal/template/model_policy.go` — existing `agentModelMap` + `agentEffortMap`; OUT OF SCOPE per decision #1.
- `internal/harness/router/effort.go` — existing harness→effort mapping; OUT OF SCOPE per decision #1.
- `SPEC-CC2178-MODEL-POLICY-REPAIR-001` — related (the dual policy that governs the retained-agent surface); disjoint from this SPEC's surface.
- `.moai/docs/template-internal-isolation-doctrine.md` — template neutrality constraints (§25 forbidden/allowed content classes).
- `.claude/rules/moai/workflow/team-protocol.md` — 7-role canonical set; role-key schema lock.
- `https://platform.claude.com/docs/en/build-with-claude/effort` — official effort guidance (cited in §D taxonomy).
- `https://code.claude.com/docs/en/workflows` — official `agent()` opts documentation (cited in §A).
