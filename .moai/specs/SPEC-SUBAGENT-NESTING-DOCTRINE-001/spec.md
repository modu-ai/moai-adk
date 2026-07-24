---
id: SPEC-SUBAGENT-NESTING-DOCTRINE-001
title: "Subagent-nesting doctrine correction + auditor read-only nesting pilot"
version: "0.1.0"
status: completed
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: ".claude"
lifecycle: spec-anchored
tags: "doctrine, subagent-nesting, claude-code, agent-authoring, sync-auditor"
tier: M
---

# SPEC-SUBAGENT-NESTING-DOCTRINE-001 — Subagent-nesting doctrine correction + auditor read-only nesting pilot

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Initial draft — M1 doc correction (v2.1.217 nesting facts) + M2 auditor read-only nesting pilot (opt-in, env-gated). |

## §A Context

Claude Code's subagent-nesting behavior changed materially at **v2.1.217**, and several always-loaded MoAI doctrine surfaces still assert the pre-change (v2.1.172–v2.1.216) world. Because `CLAUDE.md` is loaded into every session, the stale phrasing ("Nesting depth is fixed and not configurable … depth five") misleads on every turn.

### Orchestrator-verified ground truth (cite; do not re-derive)

Official (`code.claude.com/docs/en/sub-agents` + agent-teams + `anthropics/claude-code` CHANGELOG):

- Subagent nesting is gated by **BOTH** conditions: (1) the `Agent` tool present in the subagent's `tools` list, **AND** (2) the env var `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` set to a positive integer (default `0` = nesting OFF).
- Version history: **v2.1.172–v2.1.216** = nesting ON by default, fixed depth 5 (unchangeable). **v2.1.217** = CHANGED default to **OFF**; depth is now **CONFIGURABLE** via `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` (default 0).
- Concurrency caps (three independent): `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` default **20** (v2.1.217; ultracode sessions exempt); `CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION` default **200** (v2.1.212).
- Prevent nesting: omit `Agent` from `tools`, OR add `Agent` to `disallowedTools`, OR leave the depth env unset.
- The parenthesized `Agent(agent_type)` allowlist is a main-thread-only (`claude --agent`) feature; **inside a subagent definition the parenthesized type list is ignored** — only `Agent` presence matters, and read-only scoping of a spawned child must therefore rely on `mode: "plan"` (for `general-purpose`) or the inherently read-only `Explore`.
- Agent Teams teammates cannot nest (only the lead spawns). Official when-to-nest use case: "a reviewer subagent that dispatches a verifier per finding, so the intermediate output never reaches your main conversation."

Internal (orchestrator-verified via frontmatter reads):

- All 11 retained agents omit the `Agent`/`Task` tool → the flat hierarchy holds by configuration. With the v2.1.217 default-off, MoAI now also aligns with the runtime default (a **double guarantee**).
- `sync-auditor` frontmatter: `tools: Read, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill`; `permissionMode: plan` (ALREADY read-only) — the primary pilot target.
- `plan-auditor` frontmatter: `permissionMode: default` — NOT piloted by this SPEC. Its read-only nesting pilot is DEFERRED to a future SPEC (see §E Out of Scope — plan-auditor nesting pilot); the `default` permission mode would require an explicit `mode: "plan"` for read-only child scoping, which the future SPEC will own.
- The `AskUserQuestion` single-point-of-contact boundary is nesting-independent (a nested child is even further from the user; still barred). This invariant MUST be preserved unchanged.

### Affected surfaces (live + template mirror — all 7 confirmed to have mirrors)

| # | Live surface | Milestone |
|---|--------------|-----------|
| 1 | `CLAUDE.md` §4 "Watch (Claude Code 2.1.172)" note (~L64) | M1 |
| 2 | `CLAUDE.md` §14 Parallel Execution Safeguards (~L247-250) | M1 |
| 3 | `.claude/rules/moai/development/agent-authoring.md` § Agent(agent_type) Restrictions (~L94) | M1 |
| 4 | `.claude/rules/moai/development/agent-authoring.md` § Fork Subagents (~L100) | M1 |
| 5 | `.claude/rules/moai/development/agent-authoring.md` § Tool Permissions (~L226) | M1 |
| 6 | `.claude/rules/moai/development/agent-patterns.md` § Deprecated: Hierarchical Manager Chain (~L314) | M1 |
| 7 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` § Mode 6 "scaling NOT nesting" (~L66) | M1 (SHOULD-REVIEW) |
| 8 | `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-020 (L219) / CONST-V3R2-044 (L417) | M1 (conditional re-sync only) |
| 9 | `.claude/agents/moai/sync-auditor.md` (sole M2 pilot target) | M2 |

Live line numbers are indicative (2026-07-24 reads); run-phase MUST re-anchor by content token before editing.

Note on surface 8: the live CONST-V3R2-020 clause mirrors the CLAUDE.md §14 *background/concurrency* sentence ("The retained safeguard is concurrency, not backgrounding"), and CONST-V3R2-044 mirrors the `agent-common-protocol.md` § Background Agent Execution clause. **Neither clause is about nesting.** The M1 nesting corrections therefore do NOT edit these clauses; re-sync is required ONLY if the concurrency-cap sentence added in M1 point 2 falls inside the mirrored clause span.

## §B Requirements (GEARS)

### M1 — Documentation correction (low risk, no runtime behavior change)

- **REQ-SND-001** (Ubiquitous): The SPEC's M1 corrections shall replace every "nesting depth is fixed / not configurable / depth five" assertion across the identified stale surfaces with the v2.1.217 facts (default-off + depth configurable via `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`).
- **REQ-SND-002** (Event-driven): **When** a session loads `CLAUDE.md` §4, the Watch note shall state that v2.1.217 changed the default to OFF, that depth is configurable via `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`, that the MoAI flat hierarchy now holds by BOTH the runtime default-off AND `Agent`-tool omission (double guarantee), and it shall reference the M2 selective `sync-auditor` pilot exception (the one retained agent that will carry `Agent` in `tools`, kept flat at the shipped default by the env-default-off guarantee alone).
- **REQ-SND-003** (Ubiquitous): `CLAUDE.md` §14 shall document the three independent concurrency caps: `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` (default 20, ultracode exempt) and `CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION` (default 200), distinct from the depth cap.
- **REQ-SND-004** (Ubiquitous): `agent-authoring.md` § Agent(agent_type) Restrictions shall replace "nesting depth is fixed, not configurable" with "configurable via `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`, default off since v2.1.217".
- **REQ-SND-005** (Ubiquitous): `agent-authoring.md` § Fork Subagents shall mark the "fixed depth cap (depth 5) as of v2.1.187" statement as superseded by v2.1.217 (depth configurable + default-off).
- **REQ-SND-006** (Ubiquitous): `agent-authoring.md` § Tool Permissions shall note that the runtime default is now ALSO off — the MoAI flat hierarchy no longer rests on tools-omission alone.
- **REQ-SND-007** (Event-driven): **When** a reader reaches `agent-patterns.md` § Deprecated: Hierarchical Manager Chain, the version framing shall present the v2.1.217 default-off reality and shall reference the M2 selective pilot exception rather than the v2.1.172 "nesting DOES exist" framing.
- **REQ-SND-008** (Where — capability gate, SHOULD-REVIEW): **Where** `orchestration-mode-selection.md` § Mode 6 draws the "scaling NOT nesting" distinction, the correction shall keep the distinction intact and refresh only the version note to the v2.1.217 reality.
- **REQ-SND-009** (Where — conditional): **Where** an M1 edit changes the text of the CONST-V3R2-020 or CONST-V3R2-044 clause that the zone-registry mirrors, the SPEC shall re-sync the zone-registry entry to match; otherwise (the nesting facts do not touch the background/concurrency clauses) the zone-registry entries shall remain unchanged.
- **REQ-SND-010** (Ubiquitous): Every M1 live edit to `CLAUDE.md`, `.claude/rules/**`, or `.claude/agents/moai/**` shall be mirrored byte-for-byte to `internal/template/templates/` and `make build` shall be run (Template-First).
- **REQ-SND-011** (Unwanted): The template mirror content shall NOT leak this SPEC ID, an internal date, a commit SHA, an audit citation, or any macOS-bias path — only the generic, neutral v2.1.217 version facts are permitted (Template neutrality).
- **REQ-SND-012** (Ubiquitous): The M1 corrections shall introduce NO runtime behavior change — they are doctrine-prose accuracy edits only.

### M2 — Auditor read-only nesting pilot (behavioral, OPT-IN, env-gated)

- **REQ-SND-013** (Ubiquitous): `sync-auditor` shall add `Agent` to its `tools` list while keeping `permissionMode: plan`.
- **REQ-SND-014** (Where — capability gate): **Where** `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` is set to a positive integer, `sync-auditor` shall be permitted to spawn read-only per-dimension verifier children — one per scoring dimension (Functionality / Security / Craft / Consistency).
- **REQ-SND-015** (While — state-driven, held-out default): **While** `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` is unset (the shipped distribution default), `sync-auditor` shall behave exactly as today — flat, non-nesting, byte-identical runtime behavior.
- **REQ-SND-016** (Ubiquitous, verdict ownership): The binding 4-dimension verdict shall remain owned by the top-level `sync-auditor`; a spawned child shall NOT own or produce the binding verdict.
- **REQ-SND-017** (Ubiquitous, read-only children): A spawned verifier child shall be read-only — either `Explore` (inherently read-only) or `general-purpose` spawned with `mode: "plan"`. Because the parenthesized `Agent(agent_type)` allowlist is ignored inside a subagent, read-only enforcement rests on the `mode: "plan"` parameter (for `general-purpose`) or the `Explore` choice, NOT on a type allowlist.
- **REQ-SND-018** (Unwanted): No `sync-auditor` path, and no spawned child at any depth, shall invoke `AskUserQuestion` or `mcp__askuser` (the single-point-of-contact boundary holds at every depth).
- **REQ-SND-019** (Unwanted): The pilot-enabling env var shall NOT appear in the distributed template `settings.json`; it shall be LOCAL/dev-only (`settings.local.json` or a documented opt-in), so the shipped default distribution remains flat.
- **REQ-SND-020** (Unwanted): This SPEC shall NOT modify `plan-auditor` — `plan-auditor` shall NOT gain the `Agent` tool and shall NOT receive per-dimension / per-MUST-PASS verifier-child documentation. The M2 pilot scope is `sync-auditor` ONLY; the `plan-auditor` read-only nesting pilot is DEFERRED to a future SPEC (see §E Out of Scope — plan-auditor nesting pilot), which will own the explicit `mode: "plan"` child scoping that `plan-auditor`'s `permissionMode: default` requires.
- **REQ-SND-021** (Ubiquitous, body documentation): The `sync-auditor` body shall document the read-only per-dimension verifier pattern (one child per dimension, `Explore` or `general-purpose` + `mode: "plan"`) with the HARD constraints REQ-SND-016 / REQ-SND-017 / REQ-SND-018 stated inline.
- **REQ-SND-022** (Ubiquitous): The M2 `sync-auditor` edits shall be mirrored to `internal/template/templates/` and `make build` shall be run (Template-First), preserving template neutrality (REQ-SND-011).

## §C Non-Functional Constraints

- The `AskUserQuestion` orchestrator-only single-point-of-contact boundary is preserved unchanged (a nested child is further from the user, still barred).
- The shipped default distribution is flat: nesting OFF, no env in template `settings.json`, other 10 agents carry no `Agent` in `tools`.
- Doctrine facts written into template content stay generic/neutral (16-language distribution).
- No Go production code changes are required by the requirements; the only possible Go touch is a mirror-parity test allowlist entry (test data, not production code).

## §D Acceptance Criteria

The full held-in / held-out / boundary-guard / doc-accuracy AC matrix is enumerated in `acceptance.md` (AC-SND-001 … AC-SND-016).

## §E Out of Scope (exclusions)

The following are explicitly **out of scope** for this SPEC.

### Out of Scope — Agent model/effort tuning

- The `model-policy.md` medium-profile resolver inconsistency (a separately observed issue) is NOT part of this SPEC.
- No agent `model:` or `effort:` field is changed.

### Out of Scope — Enabling nesting in the shipped distribution

- The distributed template `settings.json` shall NOT set `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`; the shipped default remains flat (nesting OFF).
- No user project receives nesting on by default.

### Out of Scope — Nesting for write-capable or non-auditor agents

- `manager-develop`, `manager-docs`, `manager-git`, `builder-harness`, `manager-design`, `e2e-tester`, `super-advisor`, `manager-spec` retain flat configuration; none gains the `Agent` tool.
- No write-capable child spawning is introduced.

### Out of Scope — plan-auditor nesting pilot

- `plan-auditor` is NOT part of the M2 pilot. This SPEC pilots read-only nesting on `sync-auditor` ONLY.
- `plan-auditor` does NOT gain the `Agent` tool in this SPEC, and its body is not modified.
- Rationale: `plan-auditor` has `permissionMode: default` (not `plan`), so read-only child scoping would require an explicit `mode: "plan"` — a distinct design deferred to a future SPEC. This SPEC records the deferral here rather than as an open question.

### Out of Scope — Claude Code runtime internals

- MoAI consumes the runtime nesting/concurrency mechanism; it does NOT implement or modify Claude Code's depth/concurrency enforcement.
- Agent Teams teammate nesting (upstream disallows) is not addressed.

### Out of Scope — Live nested-execution end-to-end test

- A full runtime execution of a nested `sync-auditor` child (requires a dev session with the env set) is NOT a mechanical AC of this SPEC; M2 evidence is design + grep-observable (tools list, body documentation, boundary greps). Runtime exercise is deferred to a dev-session manual check recorded as residual risk.

## §F @MX Tag Targets (plan-phase identification)

- **No @MX code-annotation targets.** This SPEC edits doctrine prose (`CLAUDE.md`, `.claude/rules/**`) and agent frontmatter/body (`.claude/agents/moai/**`), not Go production code. No exported function, high-`fan_in` symbol, goroutine, or complexity-≥15 block is created or modified, so there is no `@MX:NOTE`/`@MX:ANCHOR`/`@MX:WARN`/`@MX:TODO` target.
- If the mirror-parity test allowlist (`internal/template/.../rule_template_mirror_test.go` or equivalent) requires a new entry, that is a test-data change, not a production-code fan_in change — still no @MX target.

## §G Cross-References

- `code.claude.com/docs/en/sub-agents` § Spawn nested subagents (official ground truth).
- `.claude/rules/moai/development/agent-authoring.md` (M1 surfaces 3-5).
- `.claude/rules/moai/development/agent-patterns.md` § Deprecated Hierarchical Manager Chain (M1 surface 6).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` § Mode 6 (M1 surface 7).
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-020 / CONST-V3R2-044 (M1 surface 8, conditional).
- `.claude/agents/moai/sync-auditor.md` (M2 surface 9 — sole pilot target). `plan-auditor.md` nesting pilot is deferred (spec.md §E Out of Scope — plan-auditor nesting pilot).
- `CLAUDE.local.md` §2 (Template-First) + §25 (Template internal-content isolation) + §15 (language neutrality).
