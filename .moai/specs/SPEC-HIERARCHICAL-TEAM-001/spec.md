---
id: SPEC-HIERARCHICAL-TEAM-001
title: "Hierarchical Agent Team — manager-lead leader + Context-Folding + peer cross-validation + schema-driven fan-out"
version: "0.1.0"
status: completed
created: 2026-08-07
updated: 2026-08-07
author: manager-spec
priority: P1
phase: "v3.x target"
module: "agents/workflow-rules"
lifecycle: spec-anchored
tags: "hierarchical-team, manager-lead, context-folding, peer-validation, schema-fanout, autonomy-epic"
tier: M
depends_on: [SPEC-AUTONOMY-TIERS-001]
related_specs: [SPEC-AUTONOMY-TIERS-001, SPEC-GOAL-HTML-WIRING-001, SPEC-STOPCHAIN-TRIM-001]
---

# SPEC-HIERARCHICAL-TEAM-001 — Hierarchical Agent Team (§3.3 redesign)

## HISTORY

- 2026-08-07 — Initial draft. Codifies §3.3 of the autonomy-workflow redesign report (`moai-autonomy-workflow-redesign-20260803.html` §3.3 Hierarchical Agent Team). Second P2 of the autonomy-workflow epic. Builds on AUTONOMY-TIERS (completed) — the `MOAI_AUTONOMY_TIER` token and 3-mode selection system — by adding a NEW retained agent (`manager-lead`) that the orchestrator MAY delegate Tier L / multi-milestone work to. G6 (the leader-agent naming decision) is user-confirmed: leader is `manager-lead`, NOT `moai` — the latter collides with the main orchestrator persona. Sibling epic items (`SPEC-GOAL-HTML-WIRING-001`, `SPEC-STOPCHAIN-TRIM-001`, `SPEC-MOAI-MCP-SERVER-001`, `SPEC-AUDIT-MULTI-MODEL-001`) are cross-referenced, NOT re-scoped. The manager-lead agent definition FILE is created in run-phase by `builder-harness`; this SPEC SPECIFIES it (role / tools / constraints / AC), it does NOT author the agent file at plan-phase.

## §A. User Story

**As a** MoAI user running a Tier L / multi-milestone SPEC,
**I want** the orchestrator to optionally delegate the run-phase to a `manager-lead` agent that coordinates worktree-isolated leaf workers, folds context at each milestone boundary, drives peer cross-validation, and reduces schema-bound explorer fan-out,
**so that** Tier L runs survive beyond a single context window without `/clear`, every accepted AC is validated by a second worker (not just self-report), and parallel reconnaissance produces mechanically-mergeable results.

**Outcome hypotheses (from §3.3):**

- **Context-Folding** (arXiv:2510.11967): milestone Mn completion → raw M{n} transcript folds to a `progress.md` §E.2 summary row + evidence files under `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`. manager-lead's context is bounded by CURRENT-milestone size, not cumulative run size — so a 6-milestone Tier L run stays inside one context window where today it would force `/clear` mid-run.
- **Peer cross-validation** (AgentOrchestra arXiv:2506.12508): after `manager-develop` implements an AC, `manager-lead` spawns a SECOND worker (NOT the author) that re-runs the AC verification commands against the same tree. Structurally stronger than self-report — the author cannot mask a failing grep by mis-counting.
- **Schema-driven fan-out** (A-MapReduce arXiv:2602.01331): explorer agents return a fixed-heading markdown schema; `manager-lead`'s reduce step is a mechanical merge, not a re-derivation. Reuses the existing `plan-research-fanout` skill's fixed-heading contract.
- **Worktree re-keying**: the `worktree-integration.md` decision tree and `agent-common-protocol.md` § Background Agent currently gate worktree use on the RETIRED team-mode predicate. They are re-keyed to "parallel write workers within a hierarchical team" so worktree isolation reaches Mode-4 fan-out without depending on a retired orchestration layer.

The net-new surfaces are: (1) the `manager-lead` agent file (run-phase, builder-harness), (2) the Context-Folding procedure (a manager-lead-driven `/compact` + evidence-persist sequence, NOT a new Go mechanism), (3) the peer cross-validation orchestration step, (4) the schema-fanout reduce contract (reusing `plan-research-fanout`), (5) the re-keyed worktree decision tree.

## §B. Scope

**In scope — 5 design axes from §3.3:**

- **Axis 1 — `manager-lead` agent**: NEW `.claude/agents/moai/manager-lead.md`. Role: coordination-only (does NOT code itself). `tools:` includes `Agent` — sole exception among the 11→12 retained agents; depth-2 sealed (leaf workers it spawns still omit `Agent`). The flat-hierarchy principle (CLAUDE.md §4 Watch note) is explicitly opened here ONLY; every other retained agent keeps the Agent-omission flat guarantee.
- **Axis 2 — worktree writer re-keying**: the worktree-integration.md decision tree's "team mode" condition is re-keyed to "parallel write workers within hierarchical team" (removes dependency on the RETIRED team-mode layer). Same re-key applies to `agent-common-protocol.md` § Background Agent Execution's team-mode references.
- **Axis 3 — per-milestone Context-Folding**: milestone Mn completion → (a) evidence persisted to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`, (b) `progress.md` §E.2 gains a summary row pointing at those evidence files, (c) manager-lead invokes `/compact` with explicit retain-current-milestone-only instructions. Orchestrator/leader context bounded by current-milestone size, not cumulative run size.
- **Axis 4 — peer cross-validation**: after `manager-develop` reports an AC PASS for milestone Mn, `manager-lead` spawns a second `Agent(general-purpose)` (NOT the author) with read-only scope to re-run the AC verification commands. Tier S skips (trivial scope); Tier M/L mandatory.
- **Axis 5 — schema-driven fan-out**: explorer agents return the existing `plan-research-fanout` skill's fixed-heading markdown schema. `manager-lead`'s reduce step is mechanical merge of N fixed-schema returns — NOT a re-derivation. Concurrency stays within MoAI Mode 4 (3-5 concurrent `Agent()`).

**Out of scope — sibling epic items:** the mode token + 3-mode selection (`SPEC-AUTONOMY-TIERS-001`, completed); the goal-evaluator HTML dashboard / plan-HTML report / re-arm UI (`SPEC-GOAL-HTML-WIRING-001`, completed); the stateful MCP tool layer (`SPEC-MOAI-MCP-SERVER-001`); the 3-way audit model (`SPEC-AUDIT-MULTI-MODEL-001`); the stopchain trim / subagent-lifecycle dormancy (`SPEC-STOPCHAIN-TRIM-001`, completed).

### Out of Scope — The `MOAI_AUTONOMY_TIER` token + 3-mode selection

- The mode token, the Go reader, the env-key constant, the 3-value enum (`SPEC-STOPCHAIN-TRIM-001`), and the tier selector + tier→bundle renderer (`SPEC-AUTONOMY-TIERS-001`) are OWNED by sibling SPECs (both completed). This SPEC CONSUMES the tier selection to decide whether `manager-lead` is offered; it does NOT redefine the token, the selector, or the bundles.

### Out of Scope — Goal evaluator + plan-HTML surfaces

- `moai goal arm`, `moai goal status`, the goal HTML dashboard, `moai goal render`, the plan-phase HTML report, and the resume auto-re-arm UI are OWNED by `SPEC-GOAL-HTML-FLOW` / `SPEC-GOAL-HTML-WIRING-001` (completed). This SPEC's Context-Folding reuses the existing `progress.md` §E.2 row format and `.moai/state/verify/<session>/` path; it does NOT redefine them.

### Out of Scope — Native Claude Code teammate runtime

- `moai cg` GLM panes, `moai cc -w <name> --spawn` teammate windows, the `~/.claude/teams/` registry, and `teammateMode` launcher handling are the NATIVE Claude Code teammate runtime (`.claude/rules/moai/core/glm-web-tooling.md` § CG Mode). This SPEC's `manager-lead` is an in-session retained agent (`Agent()` spawn), NOT a tmux-pane teammate. The two are distinct; the native runtime is unaffected.

### Out of Scope — manager-develop / manager-docs body content

- The retained `manager-develop.md` and `manager-docs.md` agent bodies are NOT rewritten by this SPEC. `manager-lead` COORDINATES them — it does NOT redefine their cycle_type, §E self-verification, or delegation-prompt template. Only the NEW `manager-lead.md` agent file is added.

### Out of Scope — Phase 4 Mode 3 (Agent Teams) resurrection

- Mode 3 (`agent-team`) remains RETIRED (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.1). `manager-lead` is a NEW Phase 4 mode-5-style sequential delegation target — NOT a resurrection of the static team-orchestration layer. Mode 3's tombstone is preserved; this SPEC does not touch the `MODE_TEAM_UNAVAILABLE` sentinel or the `--team` dispatch-axis value.

## §C. Context — the 5 design axes restated

### §C.1 Axis 1 — manager-lead agent (the coordination-only leader)

`manager-lead` is a NEW retained agent (12th — CLAUDE.md §4 catalog grows from 11 to 12). Its role is coordination-only: it does NOT write code, it does NOT author SPEC body content, it does NOT call `AskUserQuestion` (orchestrator-subagent boundary preserved — it returns blocker reports). What it DOES:

1. Receive a Tier L / multi-milestone delegation from the orchestrator with the SPEC ID, milestone map, and AC matrix.
2. Spawn leaf workers (`Agent(general-purpose)` with domain whitelists per `archived-agent-rejection.md` §C; or `Agent(manager-develop)` for implementation milestones; or `Agent(Explore)` / read-only `Agent(general-purpose)` for reconnaissance).
3. At each milestone Mn boundary: persist evidence, fold context (Axis 3), spawn peer cross-validation (Axis 4).
4. Return the synthesized §E.2 evidence row + AC matrix to the orchestrator at run-phase completion.

**Tool surface**: `tools:` includes `Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Agent, Skill` (12 tokens). The inclusion of `Agent` is the sole exception among retained agents — the flat-hierarchy guarantee (CLAUDE.md §4 Watch note, "omitting the Agent tool from its tools list is the sole remaining flat-hierarchy guarantee") is explicitly opened here ONLY. The inclusion of `Write, Edit` is scoped to coordination artifacts ONLY (fold-row appends to `progress.md` §E.2, evidence-file writes under `.moai/state/verify/`) — the "NEVER writes implementation code" constraint (§ C.1 body prose) is preserved; Write/Edit never target source files, which remain delegated to leaf workers. **Depth-2 seal**: leaf workers spawned by manager-lead carry their own `tools:` lists that OMIT `Agent` — so depth-1 = orchestrator spawns manager-lead, depth-2 = manager-lead spawns leaf workers, and NO depth-3 recursion occurs. `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=2` is the runtime ceiling manager-lead operates under.

**Selection decision tree extension** (CLAUDE.md §4): a 13th row is added — "Multi-milestone Tier L coordination (≥3 milestones AND ≥10 files AND cross-domain fan-out)? Use the `manager-lead` subagent." This is a NEW delegation path; it does NOT displace Mode 5 (sequential `manager-develop`) for Tier S/M single-milestone work.

### §C.2 Axis 2 — worktree writer re-keying

Current state (verified):

- `worktree-integration.md` line 173: "Is this a team mode implementation with parallel agents? YES → Use Agent(isolation: worktree) for write agents" — gates worktree use on the RETIRED team-mode predicate.
- `worktree-integration.md` line 194 (HARD rule): "Implementation teammates in team mode (write-capable implementation roles) MUST use isolation: worktree when spawned via Agent()" — same retired-team-mode dependency.
- `agent-common-protocol.md` § Background Agent Execution: "MoAI does not run two write-capable agents concurrently" is the retained safeguard, but the team-mode framing around it is stale.

Re-keying: the "team mode" condition is replaced with "parallel write workers within a hierarchical team" — i.e., the worktree isolation predicate decouples from the retired static team-orchestration layer and instead keys on the manager-lead-driven parallel-write fan-out shape. This makes `isolation: "worktree"` reachable for manager-lead's leaf-worker spawns WITHOUT requiring the retired team-mode machinery.

### §C.3 Axis 3 — per-milestone Context-Folding (arXiv:2510.11967)

**What it is**: a manager-lead-driven sequence at each milestone Mn boundary that (a) persists M{n} evidence, (b) folds the raw M{n} transcript to a `progress.md` §E.2 summary row, (c) compacts the active context to retain only current-milestone + fold-row state.

**Mechanism (NOT a new Go primitive)**:

1. **Evidence persistence**: each AC verification command's verbatim output is redirected to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}` per the file-redirect contract (`agent-common-protocol-reference.md` § File-redirect contract). The evidence path is cited in the §E.2 row — the cited path MUST resolve at audit time (verification-claim-integrity §2).
2. **Fold row**: manager-lead appends a `progress.md` §E.2 row of the form `M<n>: <AC-id-1>=PASS, <AC-id-2>=PASS | evidence: .moai/state/verify/<session>/M<n>.* | fold-at: <ISO-8601>`. This is the existing §E.2 row format — no schema change.
3. **Compact**: manager-lead invokes `/compact <instructions>` where the instructions explicitly retain: (a) the SPEC ID + plan.md milestones M1..Mn summary, (b) the fold rows from §E.2, (c) the active milestone M{n+1} plan. Everything else is compact-eligible. This is the canonical `/compact` mechanism (context-window-management.md § Reduction Ladder) — manager-lead drives it explicitly rather than waiting for the runtime's auto-compact.

**What it is NOT**: NOT a new MoAI Go package, NOT a new hook, NOT a new settings.json field, NOT a new CLI subcommand. The mechanism is the existing `/compact` + evidence-persistence + §E.2 row pattern, composed into a manager-lead-driven sequence.

**Outcome**: manager-lead's context is bounded by CURRENT-milestone size + cumulative fold rows, not by cumulative-raw-transcript size. A 6-milestone Tier L run stays inside one context window where today it would force `/clear` mid-run (which today loses the goal-condition + §E.2 row linkage across the boundary).

### §C.4 Axis 4 — peer cross-validation (AgentOrchestra arXiv:2506.12508)

**What it is**: after `manager-develop` reports AC-X PASS for milestone Mn, manager-lead spawns a SECOND worker — `Agent(general-purpose)` scoped read-only (no Write/Edit) — that re-runs the AC-X verification commands against the same tree and reports PASS / PARTIAL / FAIL.

**Why it is structurally stronger than self-report**: the author (`manager-develop`) cannot mask a failing grep by mis-counting, cannot cite a stale baseline, cannot skip a verification command. The peer worker has zero investment in the author's claimed PASS — its only job is to falsify or confirm.

**Scope**: Tier S skips entirely (trivial scope — peer overhead exceeds value). Tier M/L mandatory. The peer worker runs the SAME acceptance.md §D GWT verification commands the author ran; it does NOT re-derive or re-author.

**Asymmetry vs sync-auditor**: `sync-auditor` runs at sync-phase (post-implementation) and scores 4 dimensions. Peer cross-validation runs at run-phase (per-AC, mid-implementation) and is a binary PASS/PARTIAL/FAIL on the AC verification command. The two are complementary — sync-auditor is the final skeptical read; peer cross-validation is the per-AC mid-loop check.

### §C.5 Axis 5 — schema-driven fan-out (A-MapReduce arXiv:2602.01331)

**What it is**: explorer / reconnaissance agents spawned by manager-lead return a FIXED-HEADING markdown schema (the existing `plan-research-fanout` skill's contract). manager-lead's reduce step is a mechanical merge of N fixed-schema returns — NOT a re-derivation.

**Why schema-bound**: a free-form prose return from N parallel explorers forces manager-lead to re-derive structure per spawn — linear cost growth in explorer count. A fixed-schema return makes the reduce step O(N) mechanical concatenation + conflict annotation (cross-explorer contradictions surface as a named section).

**Concurrency**: stays within MoAI Mode 4 (3-5 concurrent `Agent()`). The `plan-research-fanout` skill is the canonical schema source — this SPEC does NOT redefine it.

## §D. Requirements (GEARS)

### REQ-LEAD-001 — `manager-lead` agent file (coordination-only, depth-2 sealed)

**The** `manager-lead` agent file (`.claude/agents/moai/manager-lead.md`) **shall** declare a coordination-only role: the agent does NOT write implementation code (Write/Edit are scoped to coordination artifacts only — fold-row appends to `progress.md` §E.2 and evidence-file writes under `.moai/state/verify/`), does NOT author SPEC body content, and does NOT invoke `AskUserQuestion`. **The** agent's `tools:` list **shall** include `Agent` as the sole exception among retained agents — the flat-hierarchy guarantee is explicitly opened here ONLY. **The** agent's `tools:` list **shall** also include `Write, Edit` for the coordination-artifact scope above (12 tokens total: `Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Agent, Skill`); these two tokens MUST NOT be used to write source files (delegated to leaf workers). **The** leaf workers it spawns (`Agent(general-purpose)`, `Agent(manager-develop)`, `Agent(Explore)`) **shall** omit `Agent` from their own `tools:` lists, sealing the hierarchy at depth 2 (no depth-3 recursion). **When** spawned at depth > 2 (e.g., a misconfigured `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` env), the runtime **shall** reject the nested spawn per Claude Code's own depth cap.

### REQ-LEAD-002 — `manager-lead` selection decision tree extension

**The** CLAUDE.md §4 Selection Decision Tree **shall** gain a 13th row: "Multi-milestone Tier L coordination (≥3 milestones AND ≥10 files AND cross-domain fan-out)? Use the `manager-lead` subagent." **When** a SPEC does not satisfy that predicate, the orchestrator **shall** continue to use Mode 5 (sequential `manager-develop` / `manager-docs`) — `manager-lead` is an opt-in delegation target for the multi-milestone case, NOT a replacement for the standard run-phase path.

### REQ-LEAD-003 — `manager-lead` catalog entry + CLAUDE.md §4 Watch note update

**The** CLAUDE.md §4 Retained Agents table **shall** grow from 11 to 12 entries with `manager-lead` added. **The** §4 Watch note's claim "every retained MoAI agent omits Agent, so the flat hierarchy holds by tool omission" **shall** be amended to: "every retained MoAI agent except `manager-lead` omits `Agent`; `manager-lead` is the sole Agent-carrying retained agent, and the flat hierarchy holds by depth-2 sealing (leaf workers it spawns omit Agent)." **The** `amendment_of:` / supersession note for `SPEC-SUBAGENT-NESTING-DOCTRINE-001` **shall** be extended to reference this SPEC's depth-2 seal as the active flat-hierarchy guarantee.

### REQ-DEPTH-001 — depth-2 seal CI guard (defense-in-depth)

**The** repository **shall** contain a CI test under `internal/template/` that mirrors the `subagent_boundary_test.go` pattern and greps every `manager-lead`-spawned leaf-worker agent file for the literal token `Agent` in its `tools:` list, failing the build on match. **The** test is defense-in-depth on the REQ-LEAD-001 depth-2 seal: it mechanically catches a future SPEC that amends a leaf-worker agent file to add `Agent` to `tools:` — catching it at lint time rather than at runtime (where the depth-3 spawn attempt would be rejected by `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`). **The** test is OBLIGATORY (OQ-4 RESOLVED — user confirmed); it is NOT gated on a follow-up decision.

### REQ-WORKTREE-001 — worktree decision tree re-key

**Where** the `worktree-integration.md` § Worktree Selection Rules decision tree currently gates on "team mode implementation with parallel agents", the predicate **shall** be re-keyed to "parallel write workers within a hierarchical team (e.g., `manager-lead` fan-out)". **Where** the same file's HARD rule line 194 currently says "Implementation teammates in team mode MUST use isolation: worktree when spawned via Agent()", the rule **shall** be re-keyed to "Implementation leaf workers spawned in parallel by `manager-lead` (or any parallel-write fan-out) MUST use isolation: worktree when spawned via Agent()". **The** re-keying is a predicate substitution only — the underlying `isolation: "worktree"` mechanism, the L1/L2 layer distinction, and the read-only-agent exemption are unchanged.

### REQ-WORKTREE-002 — `agent-common-protocol.md` § Background Agent consistency

**The** `agent-common-protocol.md` § Background Agent Execution paragraph that references "MoAI does not run two write-capable agents concurrently" **shall** be retained verbatim (the concurrency safeguard is unchanged). **Where** the same section's stale team-mode framing appears, it **shall** be re-keyed to the same "parallel write workers within a hierarchical team" predicate as REQ-WORKTREE-001. **The** re-keying does NOT alter the concurrency ceiling (3-5 concurrent `Agent()`), the background-default alignment, or the read-only-vs-write tool-restriction guarantee.

### REQ-FOLD-001 — per-milestone Context-Folding procedure

**When** `manager-lead` detects milestone Mn completion (all M{n} AC rows PASS per the peer cross-validation of REQ-PEER-001), the agent **shall** execute the Context-Folding procedure: (a) persist each M{n} AC verification command's verbatim output to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`, (b) append a `progress.md` §E.2 fold row of the form `M<n>: <AC-id-1>=PASS, ... | evidence: .moai/state/verify/<session>/M<n>.* | fold-at: <ISO-8601>` citing those evidence paths, (c) invoke `/compact` with explicit retain-current-milestone + retain-fold-rows instructions. **The** fold row format **shall** be the existing §E.2 row format — no schema change to progress.md.

### REQ-FOLD-002 — Context-Folding evidence-persistence obligation

**The** evidence paths cited in the §E.2 fold row **shall** resolve at audit time per the verification-claim-integrity §2 attribution requirement. **The** evidence **shall** be persisted under `.moai/state/verify/<session>/` (NOT `/tmp` — the OS clears `/tmp` and the cited path would dangle). **When** an evidence path cannot be populated (command failed, output redirected elsewhere), the fold row **shall** mark that AC `GAP` rather than `PASS`, and manager-lead **shall** NOT proceed to M{n+1} without resolving the gap or returning a blocker report.

### REQ-FOLD-003 — Context-Folding bounded-context invariant

**After** the `/compact` of REQ-FOLD-001 fires, manager-lead's active context **shall** be bounded by (current-milestone plan size) + (cumulative fold rows) + (always-loaded rule prefix) — NOT by (cumulative raw transcript). **The** invariant is observable: `moai status` (or equivalent) after fold reports a token usage approximately proportional to current-milestone size, not to Mn-cumulative. **When** the post-fold context exceeds the model-specific handoff threshold (context-window-management.md § Context Window Targets), manager-lead **shall** emit a paste-ready resume + `/clear` per session-handoff.md — the fold procedure does NOT bypass the handoff gate.

### REQ-PEER-001 — peer cross-validation spawn

**When** `manager-develop` reports an AC PASS for milestone Mn at Tier M or Tier L, `manager-lead` **shall** spawn a second `Agent(general-purpose)` (NOT the author) scoped read-only (`tools:` list omitting Write/Edit/NotebookEdit). **The** peer worker **shall** re-run the acceptance.md §D GWT verification commands for that AC against the same tree. **The** peer worker's return **shall** be one of `PASS` (the verification command reproduces), `PARTIAL` (command runs but output differs from the author's claim), or `FAIL` (command does not run or contradicts the author's PASS). **Where** the SPEC is Tier S, peer cross-validation **shall** be skipped (trivial scope — peer overhead exceeds value).

### REQ-PEER-002 — peer cross-validation blocker on FAIL

**When** the peer worker returns `FAIL` or `PARTIAL` for an AC the author marked PASS, `manager-lead` **shall** NOT advance to the next milestone. **The** agent **shall** return a structured blocker report to the orchestrator containing: the AC ID, the author's claimed PASS evidence, the peer's FAIL/PARTIAL evidence, and the divergence point. **The** orchestrator (NOT manager-lead — orchestrator-subagent boundary preserved) runs the `AskUserQuestion` round to resolve: re-spawn author / accept the divergence as a documented debt / abort the milestone.

### REQ-FANOUT-001 — schema-driven explorer fan-out

**When** `manager-lead` fans out reconnaissance (read-only exploration across N≥3 domains), each explorer agent **shall** return the fixed-heading markdown schema defined by the existing `plan-research-fanout` skill. **The** reduce step **shall** be a mechanical merge of the N fixed-schema returns — manager-lead does NOT re-derive structure per spawn. **Where** two explorers return contradictory findings on the same signal, manager-lead **shall** annotate the contradiction as a named section in the merged result (NOT silently pick one).

### REQ-FANOUT-002 — fan-out concurrency ceiling

**The** explorer fan-out concurrency **shall** stay within MoAI Mode 4 (3-5 concurrent `Agent()`). **The** manager-lead agent **shall** NOT exceed 5 concurrent leaf-worker spawns; where > 5 workers are warranted, it **shall** sequence them in batches. **This** ceiling composes with — does NOT replace — the runtime cap `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` (default 20).

### REQ-CLOSE-001 — `manager-lead` non-regression on Phase 4 mode taxonomy

**The** addition of `manager-lead` as a delegation target **shall not** alter the Phase 4 execution-mode catalog (orchestration-mode-selection.md §A): Mode 1 (trivial), Mode 2 (background), Mode 3 (agent-team — RETIRED), Mode 4 (parallel), Mode 5 (sub-agent), Mode 6 (workflow) — unchanged. **The** `manager-lead` delegation path is a Mode-5-shaped sequential delegation (orchestrator → manager-lead → leaf workers); it does NOT introduce a Mode 7, does NOT resurrect Mode 3, and does NOT modify the `--mode` dispatch axis values (`autopilot` / `loop` / `team` / `pipeline`).

## §E. Assumptions

1. The `plan-research-fanout` skill's fixed-heading markdown schema is stable enough to reuse as the fan-out return contract (AXIS 5). If the skill schema changes in flight, this SPEC's REQ-FANOUT-001 cross-reference tracks the skill (NOT a parallel schema).
2. The existing `progress.md` §E.2 row format is permissive enough to carry the fold row's `<AC-id>=PASS | evidence: .moai/state/verify/.../M<n>.* | fold-at: <ISO-8601>` structure. If era.go's §E.2 heading match is sensitive to row-prefix changes, the fold row prefix `M<n>:` is the canonical form — verified in plan.md §C Pre-flight §3.
3. The `/compact` slash-command is available from inside a subagent context (manager-lead). Claude Code's subagent tool surface for `/compact` is changelog-sourced and needs run-phase verification (OQ-1).
4. The leaf-worker `tools:` list omission of `Agent` is mechanically enforceable via the agent-body declaration (builder-harness authors the leaf-worker agent files with explicit `tools:` lists omitting `Agent`). The runtime honors the `tools:` list (CLAUDE.md §4 Watch note — verified for the 11 retained agents; manager-lead's spawned leaf workers inherit the same guarantee).
5. CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH defaults to 3 on Claude Code v2.1.219+ (changelog-sourced). manager-lead's depth-2 seal sits under this ceiling; a user setting `=1` disables manager-lead's leaf-worker spawns (the agent returns a blocker report — does NOT silently regress).
6. `.moai/state/verify/<session>/` is the canonical evidence-persistence path per `agent-common-protocol-reference.md` § Evidence persistence (already cited in the verification-batch-pattern.md cross-references). The fold procedure reuses it; no NEW path is introduced.

## §F. Open Questions (for plan-auditor + Implementation Kickoff)

- **OQ-1 (G-CTX-FOLD)** — Context-Folding trigger policy: automatic at every Tier L milestone boundary, OR opt-in via a config flag (`workflow.context_folding: auto | opt-in | off`, default `auto` for Tier L, `opt-in` for Tier M)? The cost/benefit profile differs by Tier: Tier L runs genuinely need folding to survive; Tier M runs may finish inside one context window. **Recommended**: `auto` at Tier L, `opt-in` at Tier M. **Decision DEFERRED to Implementation Kickoff** (this is the load-bearing fold-policy question).
- **OQ-2 (G-PEER-SCOPE)** — peer cross-validation AC scope: all ACs at Tier M/L (the spec's REQ-PEER-001 framing), OR only "critical-path" ACs (those tied to REQs whose `priority` ∈ `{P0, P1}`)? Peer overhead scales linearly in AC count; a 25-AC Tier L SPEC pays 25 peer spawns. **Recommended**: all ACs at Tier M/L (structural strength is the point); Tier S skips entirely. The 25-spawn cost is the price of falsifiable ACs.
- **OQ-3 (G-LEAD-TRIGGER) — RESOLVED**: conditionally spawned via the conjunctive heuristic. The `manager-lead` entry predicate is conjunctive: the run MUST satisfy ALL three dimensions (≥3 milestones AND ≥10 files AND cross-domain fan-out). User confirmed at audit-entry. Disjunctive OR was rejected as too permissive — it would admit runs satisfying only one dimension (e.g. a 10-file single-milestone refactor with no cross-domain fan-out). Tier L runs below the heuristic stay on the standard Mode 5 path.
- **OQ-4 (G-DEPTH-VERIFICATION) — RESOLVED**: add the CI guard (defense-in-depth on the flat-hierarchy carve-out). User confirmed. The guard is codified as REQ-DEPTH-001 + AC-DEPTH-001: a CI test under `internal/template/` mirroring the `subagent_boundary_test.go` pattern that greps every `manager-lead`-spawned leaf-worker agent file for `Agent` in `tools:` and fails on match.

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.3 (Hierarchical Agent Team), §6 risk row 3 (depth-2 seal + flat-hierarchy carve-out), §6 open question 2 (G-CTX-FOLD).
- Academic grounding (per the redesign report §3.3):
  - Context-Folding: arXiv:2510.11967 (the fold-at-milestone-boundary technique this SPEC's Axis 3 operationalizes).
  - Peer cross-validation: AgentOrchestra arXiv:2506.12508 (the second-worker-validates pattern this SPEC's Axis 4 operationalizes).
  - Schema-driven fan-out: A-MapReduce arXiv:2602.01331 (the fixed-schema-return + mechanical-reduce pattern this SPEC's Axis 5 operationalizes).
- Sibling SPEC (completed, CONSUMED): `SPEC-AUTONOMY-TIERS-001` — the `MOAI_AUTONOMY_TIER` token + 3-mode selection + tier→bundle renderer. This SPEC's `manager-lead` delegation predicate keys off Tier L scope; it does NOT redefine the tier system.
- Sibling SPEC (completed, cross-referenced): `SPEC-GOAL-HTML-WIRING-001` — the goal/plan-HTML surfaces this SPEC's Context-Folding reuses (`progress.md` §E.2 row + `.moai/state/verify/` path).
- Retained-agent catalog: CLAUDE.md §4 (grows from 11 to 12 with `manager-lead`); `.claude/agents/moai/` (manager-lead.md to be authored at run-phase by builder-harness).
- Worktree layer: `.claude/rules/moai/workflow/worktree-integration.md` § Worktree Selection Rules (re-keyed by REQ-WORKTREE-001); `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution (re-keyed by REQ-WORKTREE-002).
- Context-Folding surfaces: `.claude/rules/moai/workflow/context-window-management.md` § Reduction Ladder (`/compact` rung); `agent-common-protocol-reference.md` § File-redirect contract + § Evidence persistence; `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution).
- Fan-out schema source: `.claude/skills/plan-research-fanout/` (existing skill — REQ-FANOUT-001 consumes its schema).
- Phase 4 mode taxonomy: `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A (Mode 1-6 unchanged per REQ-CLOSE-001).

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-LEAD-001 (REQ-LEAD-001): `manager-lead.md` exists, declares coordination-only role, carries `Agent` in `tools:`, leaf-worker spawns omit `Agent`.
- AC-LEAD-002 (REQ-LEAD-002): CLAUDE.md §4 Selection Decision Tree carries the 13th `manager-lead` row; non-matching SPECs still use Mode 5.
- AC-LEAD-003 (REQ-LEAD-003): CLAUDE.md §4 Retained Agents table shows 12 entries; Watch note amended to name `manager-lead` as sole Agent-carrier; supersession note for SUBAGENT-NESTING-DOCTRINE references depth-2 seal.
- AC-DEPTH-001 (REQ-DEPTH-001): a CI test under `internal/template/` greps leaf-worker agent files for `Agent` in `tools:` and fails on match; the regression is caught at lint time.
- AC-WORKTREE-001 (REQ-WORKTREE-001): worktree-integration.md decision tree + HARD rule re-keyed to "parallel write workers within hierarchical team"; no residual "team mode" gating language in the decision tree.
- AC-WORKTREE-002 (REQ-WORKTREE-002): agent-common-protocol.md § Background Agent's stale team-mode framing re-keyed; concurrency safeguard retained verbatim.
- AC-FOLD-001 (REQ-FOLD-001): manager-lead executes the 3-step fold procedure at milestone Mn completion (evidence → fold row → `/compact`); fold row format matches existing §E.2 row.
- AC-FOLD-002 (REQ-FOLD-002): evidence paths under `.moai/state/verify/<session>/M<n>.*` resolve at audit time; GAP-marked ACs do not advance.
- AC-FOLD-003 (REQ-FOLD-003): post-fold manager-lead context bounded by current-milestone + fold rows + rule prefix; handoff gate still fires at model-specific threshold.
- AC-PEER-001 (REQ-PEER-001): Tier M/L ACs get a second-worker re-run; Tier S skips.
- AC-PEER-002 (REQ-PEER-002): FAIL/PARTIAL peer reports produce a blocker (no silent advance to M{n+1}); orchestrator runs AskUserQuestion.
- AC-FANOUT-001 (REQ-FANOUT-001): explorers return plan-research-fanout's fixed-heading schema; reduce is mechanical merge; contradictions annotated.
- AC-FANOUT-002 (REQ-FANOUT-002): fan-out concurrency ≤ 5 (MoAI Mode 4 ceiling preserved).
- AC-CLOSE-001 (REQ-CLOSE-001): Phase 4 mode catalog unchanged (Modes 1-6 verbatim); `--mode` dispatch axis values unchanged; no Mode 7 introduced.
- AC-REGRESS-001 (cross-cutting): spec-lint strict run on this SPEC's own directory returns 0 errors (no `LegacyEARSKeyword`, no `FrontmatterInvalid`, no `MissingExclusions`).
