---
id: SPEC-V3R2-ORC-005
title: "Dynamic Team Generation Formalization + Mailbox Protocol v2"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 6 — Multi-Mode Workflow"
module: ".moai/config/sections/workflow.yaml, internal/cli/team_spawn.go, .claude/rules/moai/workflow/worktree-integration.md, .claude/rules/moai/workflow/team-protocol.md"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-ORC-001
  - SPEC-V3R2-ORC-004
  - SPEC-V3R2-RT-004
related_problem: []
related_theme: "Layer 4 — Orchestration + Layer 6 — Workflow (cross-layer), Master §4.4, §15 Agent Teams, SPEC-TEAM-001 formalization"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "team, dynamic-generation, role-profiles, mailbox, tasklist, sendmessage, v3r2"
---

# SPEC-V3R2-ORC-005: Dynamic Team Generation Formalization + Mailbox Protocol v2

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

MoAI v2 (SPEC-TEAM-001) introduced dynamic team generation: teammates are spawned via `Agent(subagent_type: "general-purpose")` with runtime overrides from `workflow.yaml` role profiles (researcher, analyst, architect, implementer, tester, designer, reviewer). No static team-* agent definition files exist. This pattern is lightly documented across CLAUDE.md §4 ("Dynamic Team Generation") and `.moai/config/sections/workflow.yaml`, but no SPEC formally codifies:

- the role → (mode, model, isolation) mapping contract,
- the team-scoped task-list ownership model,
- the mailbox protocol (SendMessage typed messages, not just content strings),
- the spawn-time lint that rejects mis-configured teammates (addressed partially in SPEC-V3R2-ORC-004 REQ-008).

This SPEC formalizes the dynamic-team pattern as the v3r2 canonical team model. It removes any residual static `team-*` agent files if present (no such files currently exist per repo grep, but the SPEC is belt-and-suspenders). It adds a typed message envelope (`Mailbox.Send(type, payload)`), a task-list ownership rule ("one team = one TaskList"), and a team-scoped memory scope reference (Layer 4 core type from Master §4.4).

Dependencies: depends on SPEC-V3R2-RT-004 (typed session state with file-first `.moai/state/` layout) because team tasklist and mailbox state persist under `.moai/state/team/{team-id}/`.

### 1.1 Background

Master §4.4 Layer 4 Orchestration defines the core types:

```go
type Team struct {
    ID          string
    Config      TeamConfig
    Teammates   map[string]Teammate
    Mailboxes   map[string]Mailbox
    TaskList    *TaskList
}
```

with key interfaces `AgentSpawner.Spawn`, `CommonProtocolLinter.Check`, `TeamCoordinator.Create`, `Mailbox.Send`/`Recv`, `TaskList.Claim`.

Dynamic generation is already implemented. What this SPEC formalizes:

1. **Role profile contract**: `workflow.yaml role_profiles.{name}` declares `mode`, `model`, `isolation`, `description`. The spawn wrapper MUST apply these overrides at `Agent()` call time. No teammate can override the isolation flag downward (implementer in plan mode is rejected).
2. **Task list ownership**: `TaskList` lives at `.moai/state/team/{team-id}/tasklist.md` (Magentic-ledger pattern O-3 from pattern-library.md). Every teammate writes task state via `TaskUpdate`; the ledger is append-only. Team lead coordinates claim order via TaskList visibility but does NOT reassign claimed tasks.
3. **Mailbox protocol v2**: `SendMessage` gains a typed envelope. Supported types: `message` (default), `shutdown_request`, `shutdown_response`, `blocker_report`, `task_handoff`. Each type has a documented schema; untyped messages default to `message`.
4. **Teammate ID naming**: `{role}-{ordinal}` (e.g., `implementer-1`, `tester-1`). Unique within a team; the team config at `~/.claude/teams/{team-id}/config.json` lists every teammate name.
5. **Worktree enforcement**: reaffirms SPEC-V3R2-ORC-004 REQ-008 (team spawner rejects write-heavy role profiles without `isolation: worktree`).

*Source: CLAUDE.md §4 Dynamic Team Generation, §15 Agent Teams; `.moai/config/sections/workflow.yaml` role_profiles; Master §4.4 core types; pattern-library.md O-3 Magentic Ledger, O-6 Agentless; design-principles.md Principle 9 Parallelism via Explicit DAG.*

### 1.2 Non-Goals

- Adding NEW role profiles beyond the 7 already defined (researcher, analyst, architect, implementer, tester, designer, reviewer).
- Implementing CC's native TeamCreate/SendMessage/TaskList APIs (those are CC runtime primitives; this SPEC documents the wrapping pattern only).
- Changing the tmux pane cleanup behavior (SPEC-V3R2-RT-006 handles P-H02 subagentStop fix).
- CG mode (Claude + GLM) integration changes (SPEC-GLM-001 owns that).
- Pub/sub broadcast patterns (Master §13 Non-Goal: pub-sub deferred past 10 teammates).
- Federation of multiple teams (not in scope).
- Cross-team mailboxes (mailbox scope is within-team only).
- Graphql or RPC over mailbox (stays simple markdown-like typed envelopes).

---

## 2. Scope (범위)

### 2.1 In Scope

- Document the dynamic team pattern in `.claude/rules/moai/workflow/team-protocol.md` §Canonical Team Model (new subsection).
- Codify the role → (mode, model, isolation) mapping table.
- Implement typed mailbox envelope in `internal/cli/team_spawn.go` (or adjacent package):
  - `Mailbox.Send(target string, msg Message)` where `Message { Type: string, RequestID: string, Content: string, Payload map[string]any }`
  - Types: `message|shutdown_request|shutdown_response|blocker_report|task_handoff`
- Implement team-scoped task-list at `.moai/state/team/{team-id}/tasklist.md` (Magentic ledger pattern O-3); append-only semantics; `TaskUpdate` appends a new row; `TaskList.Claim()` atomically reserves the lowest-ID unblocked task.
- Implement spawn-wrapper lint: reject any Agent() spawn where role_profile is `implementer|tester|designer` and `isolation != worktree`.
- Verify and remove (if present) any static `.claude/agents/moai/team-*.md` files from both trees.
- Update `.claude/rules/moai/workflow/team-protocol.md` §Communication with the typed-envelope spec.
- Cross-reference from CLAUDE.md §4 Dynamic Team Generation and §15 Agent Teams without restating content.
- Template-first: all rule edits land in `internal/template/templates/` first; `make build` regenerates; local tree byte-identical.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Adding telemetry or metrics for team spawn duration / task completion rate.
- Wiring team mode to harness routing (SPEC-V3R2-HRN-001 decides harness → mode mapping).
- Implementing the spawn wrapper's CG (GLM) env injection (SPEC-GLM-001 owns this).
- Changing settings.json `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var (opt-in flag preserved).
- Multi-team coordination (single-team scope only).
- Pub/sub broadcast (Non-Goal per Master §13).
- Adding a new role profile beyond the 7 documented ones.
- Mailbox encryption / signing (not a security-boundary-crossing channel).
- Team shutdown retry protocol (existing tmux kill-pane flow handles per SPEC-V3R2-RT-006).
- Team lead re-election or failover (single lead per team, session-scoped).
- Remote teams (cross-machine) — Non-Goal.

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-beta.2+ (Phase 6, after RT-004 typed session state lands)
- Claude Code v2.1.50+ required for Agent Teams (coding-standards.md compatibility); v2.1.111+ recommended
- settings.json flag: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS: "1"` must be set
- workflow.yaml flag: `workflow.team.enabled: true` (default true per current config)
- Team config location: `~/.claude/teams/{team-id}/config.json` (CC-managed)
- Team state location: `.moai/state/team/{team-id}/` (moai-managed per RT-004):
  - `tasklist.md` — Magentic ledger, append-only
  - `mailbox/{teammate-id}/inbox.md` — per-teammate inbox
  - `team-config.yaml` — snapshot of role profiles at team-create time
- Max teammates per team: 10 (existing workflow.yaml limit)
- Supported role profiles: researcher, analyst, architect, implementer, tester, designer, reviewer (7)
- Write-heavy role profiles (per SPEC-V3R2-ORC-004 classifier): implementer, tester, designer

---

## 4. Assumptions (가정)

- SPEC-V3R2-RT-004 has landed; `.moai/state/` typed session state is live; `.moai/state/team/` sub-tree is addressable.
- SPEC-V3R2-ORC-001 has landed; no static team-* agent files exist (verified at SPEC-writing time via repo grep).
- SPEC-V3R2-ORC-004 has landed; the write-heavy classifier is active; the spawn-wrapper check can defer to that SPEC's REQ-008 for implementation teammates.
- Role profiles in `.moai/config/sections/workflow.yaml` are authoritative; any change to the profile list goes through CON-002 amendment protocol (they act like constitutional invariants for team mode).
- Teammates do not rely on team lead's conversation history per team-protocol.md §Context Isolation.
- Mailbox messages are markdown strings with optional YAML frontmatter for structured payload (simple, file-first, disk-durable).
- Team shutdown uses the existing shutdown_request / shutdown_response / kill-pane flow per MEMORY.md `tmux kill-session` HARD rule.
- Claude Code TeamCreate API returns a team-id that moai stores and uses for `.moai/state/team/{team-id}/` addressing.
- Python-free implementation (Go-only, per CLAUDE.local.md §7 Shell Script Hooks Only — no Python).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-ORC-005-001 (Ubiquitous) — 동적 팀 패턴 공표**
The file `.claude/rules/moai/workflow/team-protocol.md` **shall** contain a new subsection `§Canonical Team Model` declaring: "MoAI teams are generated dynamically via `Agent(subagent_type: \"general-purpose\")` with runtime overrides from `workflow.yaml` role_profiles. No static team-* agent files exist in v3r2."

**REQ-ORC-005-002 (Ubiquitous) — 역할 매트릭스 공표**
The team-protocol.md §Canonical Team Model **shall** publish the canonical role → (mode, model, isolation, write-heavy) table for all 7 role profiles:

| Role | Mode | Model | Isolation | Write-Heavy |
|---|---|---|---|---|
| researcher | plan | haiku | none | no |
| analyst | plan | sonnet | none | no |
| architect | plan | opus | none | no |
| implementer | acceptEdits | sonnet | worktree | yes |
| tester | acceptEdits | sonnet | worktree | yes |
| designer | acceptEdits | sonnet | worktree | yes |
| reviewer | plan | sonnet | none | no |

**REQ-ORC-005-003 (Ubiquitous) — 팀 태스크리스트 위치**
A team's task ledger **shall** be stored at `.moai/state/team/{team-id}/tasklist.md` as an append-only markdown document; `TaskCreate` appends a task row, `TaskUpdate` appends a status update row referencing the original task ID.

**REQ-ORC-005-004 (Ubiquitous) — 메일박스 타입 봉투**
The `SendMessage` API wrapper **shall** accept a typed envelope with field `type: message|shutdown_request|shutdown_response|blocker_report|task_handoff`; the default type is `message` when unspecified; the envelope is serialized as markdown with optional YAML frontmatter for `payload`.

**REQ-ORC-005-005 (Ubiquitous) — 팀 범위 메모리**
The team-protocol.md **shall** declare that team-scoped memory lives at `.moai/agents/{team-id}/memory/` only when a teammate's spawn override includes `memory: project`; otherwise teammates operate without team-scoped memory (default).

### 5.2 Event-Driven (이벤트 기반)

**REQ-ORC-005-006 (Event-Driven) — 팀 생성 시 상태 초기화**
**When** the orchestrator invokes `TeamCreate` with a roster list, the team spawner **shall** create the directory `.moai/state/team/{team-id}/` and write `team-config.yaml` snapshot of role_profiles at create time.

**REQ-ORC-005-007 (Event-Driven) — 쓰기집중 역할 격리 강제**
**When** the spawn wrapper invokes `Agent(subagent_type: "general-purpose")` for a role profile in {implementer, tester, designer}, the wrapper **shall** include `isolation: "worktree"` in the override parameters; absence **shall** cause spawn failure with error `ORC_WORKTREE_REQUIRED` (cross-reference SPEC-V3R2-ORC-004 REQ-008).

**REQ-ORC-005-008 (Event-Driven) — 타입 봉투 역직렬화**
**When** a teammate receives a message via `Recv()`, the mailbox parser **shall** extract the envelope type from the YAML frontmatter; if type is unrecognized, the parser **shall** default to `message` and emit a warning `MAILBOX_TYPE_UNKNOWN`.

**REQ-ORC-005-009 (Event-Driven) — 태스크 크레임 원자성**
**When** a teammate invokes `TaskList.Claim()`, the operation **shall** atomically reserve the lowest-ID unblocked task by appending a `CLAIMED by {teammate-id}` row; concurrent claims **shall** be serialized via filesystem lock on tasklist.md.

**REQ-ORC-005-010 (Event-Driven) — 팀 종료 시 아카이브**
**When** the orchestrator invokes `TeamDelete` after successful shutdown (all teammates responded shutdown_response:true), the team spawner **shall** rename `.moai/state/team/{team-id}/` to `.moai/state/team-archive/{team-id}-{timestamp}/` for audit retention.

### 5.3 State-Driven (상태 기반)

**REQ-ORC-005-011 (State-Driven) — 동시성 상한**
**While** `workflow.yaml team.max_teammates` is set (default 10), the team spawner **shall** reject any TeamCreate request with roster size exceeding the limit; error `ORC_TEAM_ROSTER_LIMIT`.

**REQ-ORC-005-012 (State-Driven) — 역할 스키마 고정**
**While** the v3.0.0 minor cycle is active, only the 7 role profiles listed in REQ-002 **shall** be valid `workflow.yaml` keys; adding a new role profile requires SPEC amendment through the CON-002 protocol.

**REQ-ORC-005-013 (State-Driven) — 태스크 리스트 불변성**
**While** a team is active, the task ledger **shall** be append-only; no teammate or team lead **shall** remove or reorder rows; corrections **shall** be applied via new `TaskUpdate` rows referencing the original task ID.

### 5.4 Optional (선택)

**REQ-ORC-005-014 (Optional) — 블로커 리포트 전용 타입**
**Where** a teammate encounters a hard blocker it cannot resolve, it **may** send a `blocker_report` typed message to the team lead containing fields `{blocker_id, context, suggested_resolution, required_inputs}` in the YAML frontmatter payload.

**REQ-ORC-005-015 (Optional) — 태스크 핸드오프**
**Where** a task's scope exceeds a single teammate's role profile, the teammate **may** send a `task_handoff` typed message to another teammate with payload `{task_id, reason, required_role}`; the receiving teammate **may** claim the handed-off task via TaskList.Claim.

### 5.5 Unwanted Behavior

**REQ-ORC-005-016 (Unwanted Behavior) — 정적 team-* 에이전트 금지**
**If** a file matching `.claude/agents/moai/team-*.md` is introduced to the repository, **then** CI (via extension of SPEC-V3R2-ORC-002 lint) **shall** fail with error `ORC_STATIC_TEAM_AGENT_PROHIBITED` because v3r2 uses exclusively dynamic team generation.

**REQ-ORC-005-017 (Unwanted Behavior) — 역할 없는 팀원 금지**
**If** a TeamCreate request includes a teammate with a role_profile not matching any `workflow.yaml` role_profiles key, **then** the team spawner **shall** reject the request with error `ORC_UNKNOWN_ROLE_PROFILE` listing the unknown role name.

**REQ-ORC-005-018 (Unwanted Behavior) — 브로드캐스트 기본값 금지**
**If** a teammate sends a message without specifying a target teammate name, **then** the mailbox wrapper **shall** reject the send with error `ORC_BROADCAST_NOT_PERMITTED` per team-protocol.md §Communication "NEVER broadcast unless a critical blocking issue affects ALL teammates".

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-ORC-005-01**: `team-protocol.md` contains §Canonical Team Model with the 7-row role matrix (REQ-002).
- **AC-ORC-005-02**: `find .claude/agents/moai/ -name 'team-*.md'` returns empty on both trees.
- **AC-ORC-005-03**: Invoking TeamCreate with a 3-teammate roster creates `.moai/state/team/{team-id}/` with `team-config.yaml` and empty `tasklist.md`.
- **AC-ORC-005-04**: Sending a message with `type: blocker_report` produces a parseable markdown file in the inbox with YAML frontmatter carrying the payload.
- **AC-ORC-005-05**: Two teammates invoking `TaskList.Claim()` concurrently result in distinct task IDs reserved (atomic claim test).
- **AC-ORC-005-06**: Attempting to spawn an implementer teammate without `isolation: worktree` fails with `ORC_WORKTREE_REQUIRED`.
- **AC-ORC-005-07**: Attempting to spawn with role_profile `coordinator` fails with `ORC_UNKNOWN_ROLE_PROFILE`.
- **AC-ORC-005-08**: Sending a message with no target fails with `ORC_BROADCAST_NOT_PERMITTED`.
- **AC-ORC-005-09**: Creating a team with 11 teammates fails with `ORC_TEAM_ROSTER_LIMIT`.
- **AC-ORC-005-10**: TeamDelete after all shutdown_responses archives state to `.moai/state/team-archive/{team-id}-{timestamp}/`.
- **AC-ORC-005-11**: Attempting to delete a row from `tasklist.md` fails validation (append-only check).
- **AC-ORC-005-12**: Creating file `.claude/agents/moai/team-custom.md` triggers CI failure with `ORC_STATIC_TEAM_AGENT_PROHIBITED`.

---

## 7. Constraints (제약)

- [HARD] Team-protocol rule update is EVOLVABLE under CON-001; canonical team model clause lands via graduation protocol.
- [HARD] Template-First (CLAUDE.local.md §2).
- [HARD] Role profiles are schema-locked to the 7 documented names (REQ-012); new roles require SPEC amendment.
- [HARD] Task ledger append-only (REQ-013) — no modifications or deletes.
- [HARD] No broadcast default; every SendMessage has an explicit target (REQ-018).
- [HARD] Worktree enforcement for write-heavy roles preserved from SPEC-V3R2-ORC-004 REQ-008.
- [HARD] tmux kill-pane sequence preserved per MEMORY.md HARD rule; this SPEC does not bypass it.
- [HARD] Max 10 teammates per team (REQ-011).
- [HARD] Go-only implementation; no Python dependencies.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                              | Impact | Mitigation                                                                                          |
|-------------------------------------------------------------------|--------|-----------------------------------------------------------------------------------------------------|
| Append-only tasklist grows unbounded across long teams            | MEDIUM | Archive on TeamDelete (REQ-010); clear policy at 1000-row threshold (future optimization)           |
| Typed envelope parsing fails on legacy v2 untyped messages        | MEDIUM | REQ-008 default `message` on missing type; backward compatible                                      |
| Filesystem lock for claim serialization slow on NFS / Windows     | LOW    | Go's `flock` abstraction; Windows uses LockFileEx; document tested platforms                        |
| New role profile needed mid-cycle blocks feature work             | LOW    | CON-002 amendment is bounded; canary + human approval; add when needed                              |
| Team-archive/ directory unbounded growth                          | LOW    | Document retention policy; users clean with `moai clean team-archive` (future CLI subcommand)       |
| SendMessage target teammate offline                               | MEDIUM | Message persists in inbox markdown; offline teammate reads on resume                                |
| Task handoff cycles (A→B→A)                                       | LOW    | REQ-015 is optional; team lead intervention via blocker_report if cycle detected                    |
| Team lead assumes work complete while teammate idle               | MEDIUM | team-protocol.md §Shutdown Handling + TeammateIdle hook + TaskCompleted hook (existing primitives)  |
| Cross-team mailbox requested by users                             | LOW    | Non-Goal per §1.2; redirect to federation pattern post-v3.0                                         |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification)
- **SPEC-V3R2-ORC-001** (17-agent roster; no static team files)
- **SPEC-V3R2-ORC-004** (Worktree MUST for write-heavy roles, spawn-wrapper check)
- **SPEC-V3R2-RT-004** (Typed session state at `.moai/state/`)

### 9.2 Blocks

- **SPEC-V3R2-WF-003** (Multi-mode router) — `--mode team` depends on this SPEC's canonical team model.
- **SPEC-V3R2-HRN-001** (Harness routing) — references role profiles when selecting implementer effort level.

### 9.3 Related

- **SPEC-TEAM-001** (v2 legacy) — v3r2 supersedes the v2 implementation but preserves the dynamic-generation approach.
- **SPEC-V3R2-RT-006** (Hook handler completeness) — owns the SubagentStop tmux cleanup (P-H02 fix).
- **SPEC-GLM-001** (GLM compatibility) — CG mode teammate spawning is complementary; this SPEC applies equally in cc/glm/cg.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-01; REQ-003 → AC-03, AC-11; REQ-004 → AC-04, AC-08; REQ-005 → memory-test regression; REQ-006 → AC-03; REQ-007 → AC-06; REQ-008 → AC-04 parsing test; REQ-009 → AC-05; REQ-010 → AC-10; REQ-011 → AC-09; REQ-012 → AC-07; REQ-013 → AC-11; REQ-014 → AC-04; REQ-015 → handoff regression; REQ-016 → AC-12; REQ-017 → AC-07; REQ-018 → AC-08.
- Total REQ count: 18 (Ubiquitous 5, Event-Driven 5, State-Driven 3, Optional 2, Unwanted 3)
- Expected AC count: 12
- Wave 1/2 sources:
  - `major-v3-master.md` §4.4 Layer 4 Orchestration (Team core type), §15 Agent Teams
  - `pattern-library.md` O-3 Magentic Ledger (task-list pattern), O-6 Agentless (non-multi-agent contrast), O-4 Multi-Mode Router
  - `design-principles.md` Principle 9 (Parallelism via Explicit DAG), Principle 10 (Agent Count Matches Task Structure)
  - CLAUDE.md §4 Dynamic Team Generation, §15 Agent Teams
  - `.moai/config/sections/workflow.yaml` role_profiles (authoritative)
  - `.claude/rules/moai/workflow/team-protocol.md` (existing rule file, extended here)
  - SPEC-TEAM-001 (legacy v2 precedent)
- Code-side paths:
  - `.claude/rules/moai/workflow/team-protocol.md` (modified, REQ-001, REQ-002, REQ-005)
  - `internal/cli/team_spawn.go` (new or modified, REQ-004, REQ-006, REQ-007, REQ-008, REQ-009)
  - `internal/cli/agent_lint.go` (extended, REQ-016)
  - `.moai/state/team/{team-id}/` (runtime layout, REQ-003, REQ-006, REQ-010)
  - `.moai/config/sections/workflow.yaml` (verified, REQ-002, REQ-012)
  - `internal/template/templates/.claude/rules/...` (template-first mirrors)

---

End of SPEC.
