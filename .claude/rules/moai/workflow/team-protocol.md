---
paths: "**/.claude/agents/**,**/.claude/teams/**,.moai/config/sections/workflow.yaml"
---

# Team Protocol

Canonical protocol for MoAI Agent Teams — dynamic generation, role profiles, typed messaging, and task management.

## Canonical Team Model

MoAI teams are generated dynamically via `Agent(subagent_type: "general-purpose")` with runtime overrides from `workflow.yaml` role_profiles. No static team-* agent files exist in v3r2. Every teammate is a general-purpose agent whose behavior is shaped by the role profile applied at spawn time.

### Role Matrix

The canonical role to (mode, model, isolation, write-heavy) mapping. Source: `.moai/config/sections/workflow.yaml` `role_profiles`.

| Role | Mode | Model | Isolation | Write-Heavy |
|------|------|-------|-----------|-------------|
| researcher | plan | haiku | none | no |
| analyst | plan | sonnet | none | no |
| architect | plan | opus | none | no |
| implementer | acceptEdits | sonnet | worktree | yes |
| tester | acceptEdits | sonnet | worktree | yes |
| designer | acceptEdits | sonnet | worktree | yes |
| reviewer | plan | sonnet | none | no |

**Constraints:**
- Write-heavy roles (implementer, tester, designer) MUST use `isolation: "worktree"` — enforced by spawn wrapper and LR-05 lint rule (SPEC-V3R2-ORC-004).
- Read-only roles (researcher, analyst, architect, reviewer) MUST NOT use `isolation: "worktree"` — enforced by LR-09 lint rule.
- Role profiles are schema-locked to these 7 names during v3.0.x. Adding a new role requires SPEC amendment through CON-002 protocol.
- No teammate can override the isolation flag downward (implementer in plan mode is rejected).

### Teammate Naming

Teammate IDs follow the pattern `{role}-{ordinal}` (e.g., `implementer-1`, `tester-1`). Unique within a team. The team config at `~/.claude/teams/{team-id}/config.json` lists every teammate name.

### Team State Layout

Team-scoped state lives under `.moai/state/team/{team-id}/`:

```
.moai/state/team/{team-id}/
  team-config.yaml       # Snapshot of role_profiles at team-create time
  tasklist.md            # Magentic ledger, append-only
  mailbox/
    {teammate-id}/
      inbox.md           # Per-teammate inbox
```

### Team-Scoped Memory

Team-scoped memory lives at `.moai/agents/{team-id}/memory/` only when a teammate's spawn override includes `memory: project`. By default, teammates operate without team-scoped memory. Memory files follow the 4-type taxonomy (user, feedback, project, reference) per SPEC-V3R2-EXT-001.

## Mailbox Protocol v2

SendMessage uses a typed envelope. Supported types: `message`, `shutdown_request`, `shutdown_response`, `blocker_report`, `task_handoff`.

### Envelope Schema

```yaml
---
type: message          # Required: message|shutdown_request|shutdown_response|blocker_report|task_handoff
request_id: ""         # Optional: correlation ID for request/response pairs
---

[message content as markdown]
```

### Type Definitions

**message** (default): Free-form communication between teammates. Used for sharing findings, asking questions, coordinating work.

**shutdown_request**: Team lead requests teammate shutdown. Contains `request_id` for correlation.

**shutdown_response**: Teammate responds to shutdown_request. Includes `approve: true/false` and optional `reason`.

**blocker_report**: Teammate reports an unresolvable blocker. Payload contains:
- `blocker_id`: Unique identifier
- `context`: What the teammate was doing
- `suggested_resolution`: What might fix it
- `required_inputs`: What the teammate needs to proceed

**task_handoff**: Teammate transfers a task to another teammate. Payload contains:
- `task_id`: The task being handed off
- `reason`: Why the handoff is needed
- `required_role`: Which role profile should pick it up

### Communication Rules

- Every SendMessage MUST specify an explicit target teammate name — no broadcast default.
- Teammates communicate via direct messages by default.
- Broadcast is prohibited unless a critical blocking issue affects ALL teammates.
- Messages persist in inbox markdown files; offline teammates read on resume.
- Untyped messages default to `message` type with a warning `MAILBOX_TYPE_UNKNOWN`.

## Task Ledger

The task ledger at `.moai/state/team/{team-id}/tasklist.md` follows the Magentic Ledger pattern (append-only):

- `TaskCreate` appends a new task row.
- `TaskUpdate` appends a status update row referencing the original task ID.
- `TaskList.Claim()` atomically reserves the lowest-ID unblocked task by appending a `CLAIMED by {teammate-id}` row.
- Concurrent claims are serialized via filesystem lock on tasklist.md.
- No teammate or team lead shall remove or reorder rows — corrections via new TaskUpdate rows.
- On TeamDelete, the ledger is archived to `.moai/state/team-archive/{team-id}-{timestamp}/`.

## Spawn Wrapper Validation

The team spawn wrapper enforces these checks before creating a teammate:

1. **Role validation**: Reject unknown role profiles with `ORC_UNKNOWN_ROLE_PROFILE`.
2. **Worktree enforcement**: Reject write-heavy roles without `isolation: "worktree"` with `ORC_WORKTREE_REQUIRED`.
3. **Roster limit**: Reject teams exceeding `workflow.yaml team.max_teammates` (default 10) with `ORC_TEAM_ROSTER_LIMIT`.
4. **Static agent check**: CI fails with `ORC_STATIC_TEAM_AGENT_PROHIBITED` if any `team-*.md` file exists in `.claude/agents/moai/`.

## Cross-References

- Worktree isolation details: `.claude/rules/moai/workflow/worktree-integration.md`
- Team communication protocol (discovery, shutdown, idle): `.claude/rules/moai/workflow/worktree-integration.md` §Team Protocol
- Agent lint rules (LR-05, LR-09, LR-10): `internal/cli/agent_lint.go`
- Role profile source of truth: `.moai/config/sections/workflow.yaml` `role_profiles`
- Team pattern cookbook: `.claude/rules/moai/workflow/team-pattern-cookbook.md`

---

Version: 1.0.0
Source: SPEC-V3R2-ORC-005
