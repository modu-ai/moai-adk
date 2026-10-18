---
id: SPEC-AGENT-CORE4-001
title: "Agent core four: neutral specs, two renders, verdict-only auditor guard, memory cap, retirement of three coordination agents"
version: "0.1.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: "internal/template/agentemit"
lifecycle: spec-anchored
tags: "agents, neutral-spec, codex-render, auditor-guard, memory-cap, retirement, token-budget"
tier: L
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-AGENT-MODEL-INHERIT-001, SPEC-V3R3-RETIRED-AGENT-001, SPEC-AUDIT-CEILING-001, SPEC-AGENT-MEMORY-DRAIN-001]
---

# SPEC-AGENT-CORE4-001 — Agent core four

## §A Background and motivation

Twelve agent definitions ship today. A measured plan-phase spawn pair (the planning agent and the
plan auditor) injects 196,367 bytes into a session before any work starts (`research.md` §1), and
the three copies of every definition drift apart (7 of 12 project/template Markdown pairs differ;
the project Codex mirror holds 11 TOML files against the template's 12). Three agents —
`manager-lead`, `manager-todo`, `manager-git` — are retained by name while the work they were
created for runs through other paths (a role-injected factory leader, the `moai todo` queue, the
`factory complete` landing edge).

This SPEC makes four agents the core (planner, implementer, documenter, auditor), renders every
agent definition for both harnesses from one neutral source, caps what an agent can inject and
write, adds a thin worker agent, retires the three coordination agents with their replacement paths
standing first, and generates the settings that keep the core agents' preloaded skills out of the
skill list. Design source: the 2026-10-10 agents-and-skills redesign, section 4 and 2.1 (untracked
in this tree; cited as evidence, not as authority).

## §B Definitions

- **KB** — 1,000 bytes. The redesign's "196KB" is 196,367 bytes; the same decimal convention
  governs every budget below.
- **Neutral source** — one file per agent declaring name, description, capabilities in the
  harness-neutral vocabulary (read, write with an optional path scope, shell, spawn, MCP tool
  groups), sandbox, model, effort, preloaded skills, memory policy, role hooks, and body.
- **The copies** — C1 the project Claude definitions (`.claude/agents/moai/*.md`), C2 the template
  Claude definitions (`internal/template/templates/.claude/agents/moai/*.md`), C3 the template Codex
  definitions (`internal/template/templates/.codex/agents/moai/*.toml`), and C3p the project Codex
  mirror (`.codex/agents/moai/*.toml`). The user-scope copies under `~/.claude/agents` are deployed
  output outside this repository and are not a fourth copy (see Out of Scope).
- **Spawn pair** — the two agent definitions a phase spawns: plan = planner + auditor, sync =
  documenter + auditor. Its **injected bytes** = both rendered definitions + the preloaded skill
  bodies of both + each agent's MEMORY.md counted at its cap.
- **Retired names** — `manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`,
  `sync-auditor` (absorbed into the core four) and `manager-lead`, `manager-todo`, `manager-git`
  (retired with replacement paths).
- **Resulting roster** — nine agents: planner, implementer, documenter, auditor, worker,
  builder-harness, e2e-tester, manager-design, super-advisor (plus the built-in Explore). The last
  four are carried unchanged in behavior; their move to the builder seed catalog belongs to a
  sibling card.

## §C Scope

In scope: the neutral source set and its two renderers; generation and verification of C1-C3p;
the size and memory budgets; the auditor verdict-only guard; spawn-record enrichment; the worker
agent; retirement of eight names with a refusal-with-replacement path; the reference sweep across
rules, skills, hooks, configuration, Go production code and template mirrors; roster-listing
updates; generated skill-hiding settings and their doctor check.

## §D Requirements

### D.1 Neutral source and renders

- **REQ-AC4-001** The agent definition set shall have exactly one neutral source file per agent of
  the resulting roster, and no rendered definition shall exist without one.
- **REQ-AC4-002** The Claude renderer shall produce each agent's Markdown definition from its
  neutral source deterministically, mapping every neutral capability to a tool grant, and shall
  fail emission naming the file and the token when a capability has no mapping.
- **REQ-AC4-003** The Codex renderer shall produce each agent's TOML definition from the same
  neutral source, inlining the bodies of its preloaded skills into the role instructions (per-agent
  skill grants are not observable on Codex), omitting Claude-only notation from the rendered body,
  and giving every permission-contract restriction either an enforced mapping or an explicit
  UNSUPPORTED declaration.
- **REQ-AC4-004** The copies C1, C2, C3 and C3p shall be generated artifacts: re-rendering shall
  leave zero differences, a hand edit to a rendered file shall fail the read-only check, and C1
  shall be byte-identical to C2 (neutral bodies carry no install-root path).
- **REQ-AC4-005** When the size command measures a spawn pair, the plan pair shall not exceed
  56,000 injected bytes and the sync pair shall not exceed 46,000, each per-agent preloaded skill
  shall not exceed 4,000 bytes, and the report shall name every component it summed.

### D.2 Guards and records

- **REQ-AC4-006** While an agent whose capabilities include verdict-only write is the acting agent,
  when it invokes a file-writing tool, the system shall refuse any target outside the
  `.moai/reports/` tree of the project tree the agent runs in with a sentinel-prefixed reason,
  shall allow targets inside it, shall not constrain any other agent, and shall allow the call
  (with an audit row) when the target cannot be determined.
- **REQ-AC4-007** Where the Codex runtime does not apply a role's sandbox and has no path-scoped
  write, the emitter shall render the auditor read-only, record its write-path-scope restriction
  as UNSUPPORTED rather than enforced, and leave the verdict file to be written by the launching
  parent.
- **REQ-AC4-008** The system shall refuse a write that would leave an agent's MEMORY.md above 8,000
  bytes with a sentinel-prefixed reason naming the archive remedy (lines move, none are deleted),
  the auditor shall declare no persistent memory, and the memory doctor shall report every agent
  over the cap.
- **REQ-AC4-009** When an agent is spawned, the spawn record shall carry the agent name, the phase,
  the effort and where it came from (declared, frontmatter default, or session), and the prompt
  byte length; it shall never carry the prompt text, and a record that cannot be attributed to an
  agent shall not be written.
- **REQ-AC4-010** The roster shall include a worker agent with an explicit tool list that excludes
  the spawn capability, and the run workflow's large-scope fan-out clause shall direct the
  orchestrator to spawn implementer-class leaf workers at depth one.

### D.3 Retirement

- **REQ-AC4-011** The system shall remove a retiring definition only in a commit that follows (in
  the commit graph) the commit that lands and verifies that definition's replacement path:
  the role-injected factory leader without `manager-lead`, the `moai todo` queue verbs and todo
  workflow without `manager-todo`, and the `factory complete` landing edge without `manager-git`.
- **REQ-AC4-012** When an Agent spawn names a retired name after cutover, the system shall refuse
  it with a sentinel-prefixed reason that names the replacement (a core agent with its phase
  argument, or the replacement command), shall not rewrite the spawn, and shall let every
  non-retired name through; the archived-agent rejection rule shall carry one table row per retired
  name.
- **REQ-AC4-013** Retired names shall not remain in rules, skills, hooks, configuration, Go
  production code, tests, or template mirrors outside a declared allowlist of historical and
  fixture paths; docs-site and README files are allowlisted with the card that owns their rewrite.
- **REQ-AC4-014** The roster-listing sites registered with the roster guard shall change in the same
  commit as the definition set, the guard shall stay green at every milestone commit, and the
  guard's enumeration threshold shall scale with the universe size so an undeclared listing of
  seven of ten names is still found.
- **REQ-AC4-015** The commit-trailer owner enumeration and the ownership lint shall accept the core
  agent names for the transitions their predecessors owned, and shall keep accepting the legacy
  values in existing history.
- **REQ-AC4-016** The lane spawn-authority statements (rules and the session-start notice) shall
  name planner, implementer, documenter and auditor as the specialists for plan, run, sync and the
  audits.

### D.4 Skill hiding and preload

- **REQ-AC4-017** The system shall generate, from the preloaded-skill lists of the neutral sources,
  the Claude settings entries that hide each core-agent skill by its explicit name (no wildcard),
  and the doctor shall report any such skill that is exposed to the skill list.
- **REQ-AC4-018** Before a hide mode ships, a recorded probe shall show that a skill hidden by that
  mode still reaches a subagent through its preloaded-skills field; the shipped mode shall be the
  first of the ordered candidates whose probe passes.
- **REQ-AC4-019** The Codex publication shall contain no core-agent skill under the published
  skills root and no `[[skills.config]]` entry for an absent path.

### D.5 Process

- **REQ-AC4-020** The work shall originate in the template tree and regenerate local copies,
  precede each behavior change by its failing test, introduce no skipped test and relax no
  existing assertion, and land each milestone as its own commit named for the SPEC and milestone.

## §E Acceptance

Binary-testable criteria, each with its command, expected output, and RED-now cell: `acceptance.md`.

## §F Exclusions

### Out of Scope — docs-site and README rewrite

- The four-locale docs-site pages and the four README files that name retired agents (170 and 4
  files at HEAD ae34f26831); they stay allowlisted until the documentation-rewrite cards regenerate
  the agent reference pages.

### Out of Scope — builder seed catalog and the other three carried agents

- Moving builder-harness, e2e-tester, manager-design and super-advisor into the seed catalog, the
  `moai harness create` verb, and slimming `moai-foundation-cc` (sibling card for the builder).

### Out of Scope — command-skill split and dispatcher

- Splitting `/moai` workflows into command skills, the `moai-factory` command skill, and
  renumbering workflow steps (sibling card for the skill reorganization); only the large-scope
  fan-out clause and spawn names in existing workflows change here.

### Out of Scope — user-scope installation

- Refreshing or deleting `~/.claude/agents/*` and profile-directory agent copies (the install-model
  card); this SPEC only reports their staleness.

### Out of Scope — runtime model behavior

- Changing how Claude Code or Codex resolve models, effort, or sandbox for subagents; the SPEC
  records what is measured and declares the rest UNSUPPORTED.

## §G Risks

1. **Hidden skills may not preload.** The documentation says `disable-model-invocation` blocks
   preloading and does not say whether `skillOverrides: "off"` does; the mode ladder (REQ-AC4-018)
   exists because of this.
2. **Reference sweep size.** 977 files outside the history directories name a retired name
   (`research.md` §3); the allowlist plus the semantic-rewrite class split bounds it, but M7 is the
   largest milestone.
3. **Running lanes hold old names in context.** The refusal-with-replacement path (REQ-AC4-012)
   is what keeps a lane that spawns `manager-develop` after cutover from stalling.
4. **Auditor shrink may drop protocol.** Moving rubric text out of the auditor definition must
   preserve every rubric heading (AC-AC4-009).
5. **Roster guard sensitivity.** With a ten-name universe the fixed enumeration threshold of 10
   would stop finding partial listings (REQ-AC4-014).

## §H HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Initial plan (card t1649, tree ae34f26831). |
