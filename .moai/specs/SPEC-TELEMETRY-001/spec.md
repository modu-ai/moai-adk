# SPEC-TELEMETRY-001: Skill Usage Telemetry

## Meta

- **Status**: Implemented
- **Wave**: 1 (parallel with SKILL-ENHANCE-001, CORE-BEHAV-001)
- **Created**: 2026-04-11
- **Origin**: Memento-Skills paper (P1: Read-Write Reflective Learning - Read phase data collection)
- **Blocked By**: SPEC-EVO-001 (storage location)
- **Blocks**: SPEC-REFLECT-001 (depends on telemetry data)

## Objective

Track which skills are loaded, how often, in what context, and with what outcome signal. This data feeds the Reflective Write Hook (SPEC-REFLECT-001) for learned skill routing and evolution proposals. Extends the existing `post_tool_metrics.go` JSONL pattern.

## Background

moai-adk already has partial telemetry:
- `internal/hook/post_tool_metrics.go` (137 LOC): Logs task metrics to `.moai/logs/task-metrics.jsonl`
- `internal/hook/session_start.go`: Initializes session context
- `internal/hook/session_end.go` (715 LOC): Persists metrics, cleans resources

The Memento-Skills paper describes a "Read phase" where a skill router selects skills based on past success rates. To enable this, we need systematic skill usage data collection.

## Requirements (EARS Format)

### R1: Skill Invocation Recording [EVENT]

WHEN a skill is invoked via the Skill tool, the system SHALL record the invocation in `.moai/evolution/telemetry/usage.jsonl`.

**Record Schema:**
```jsonl
{
  "ts": "2026-04-11T10:30:00Z",
  "session_id": "abc123",
  "skill_id": "moai-workflow-tdd",
  "trigger": "explicit",
  "context_hash": "a1b2c3d4",
  "agent_type": "manager-tdd",
  "phase": "run",
  "duration_ms": 45000,
  "outcome": "success"
}
```

**Fields:**
- `ts`: ISO-8601 timestamp
- `session_id`: Claude Code session identifier (from SessionStart)
- `skill_id`: YAML frontmatter `name` of the invoked skill
- `trigger`: `explicit` (user/agent invoked) | `auto` (progressive disclosure loaded)
- `context_hash`: SHA-256 first 8 chars of task description (NO PII)
- `agent_type`: Agent that triggered the skill (or "user" if direct)
- `phase`: SPEC phase (plan/run/sync/none)
- `duration_ms`: Time from skill load to next tool call (approximate)
- `outcome`: `success` | `partial` | `error` | `unknown`

**Acceptance Criteria:**
- [ ] New file `internal/telemetry/recorder.go` implements `RecordSkillUsage()`
- [ ] PostToolUse hook for Skill tool calls `RecordSkillUsage()`
- [ ] Records are appended atomically (write line + fsync)
- [ ] No PII: context is hashed, no user input stored
- [ ] File rotation: new file per day (`usage-YYYY-MM-DD.jsonl`)
- [ ] Storage path: `.moai/evolution/telemetry/` (protected by SPEC-EVO-001)

### R2: Outcome Signal Detection [EVENT]

WHEN a session ends (Stop hook), the system SHALL attempt to determine outcome signals for skill invocations in the current session.

**Outcome Heuristics:**
- `success`: All tests pass after skill invocation + no error tools
- `partial`: Some tests pass, some warnings
- `error`: Error-related tools used after skill invocation (debugging, fix)
- `unknown`: Cannot determine (default)

**Acceptance Criteria:**
- [ ] `internal/telemetry/outcome.go` implements `DetermineOutcome()` heuristic
- [ ] Stop hook (`internal/hook/stop.go`) calls outcome determination
- [ ] Updates existing usage records with outcome signal (in-place update of last line)
- [ ] Heuristic is conservative: `unknown` is the default, not `success`

### R3: Telemetry Report [OPT]

WHERE telemetry data exists, `moai telemetry report` SHALL generate a skill effectiveness summary.

**Report Format:**
```
Skill Usage Report (last 30 days)
=====================================
Skill                    | Uses | Success | Partial | Error
moai-workflow-tdd        |   45 |   38    |    5    |   2
moai-workflow-run        |   32 |   28    |    3    |   1
...

Top Co-occurring Skills:
  moai-workflow-tdd + moai-workflow-ddd: 18 times
  moai-workflow-spec + moai-workflow-run: 15 times

Underutilized Skills (loaded but rarely invoked):
  moai-design-craft: 2 uses in 30 days
  moai-formats-data: 1 use in 30 days
```

**Acceptance Criteria:**
- [ ] New CLI subcommand `moai telemetry report` in `internal/cli/telemetry.go`
- [ ] Reads all `.moai/evolution/telemetry/usage-*.jsonl` files
- [ ] Aggregates by skill_id, groups by outcome
- [ ] Shows co-occurrence patterns (skills used in same session)
- [ ] Identifies underutilized skills (< 3 uses in 30 days)

### R4: Privacy and Storage Limits [UBIQ]

The system SHALL NOT store personally identifiable information and SHALL enforce storage limits.

**Acceptance Criteria:**
- [ ] `context_hash` uses SHA-256 truncated to 8 chars (irreversible)
- [ ] No raw user prompts, file contents, or paths stored in telemetry
- [ ] Session IDs are Claude Code internal IDs (not user-identifiable)
- [ ] Storage limit: 90 days retention, files older than 90 days auto-deleted on session start
- [ ] `.moai/evolution/telemetry/` is in `.gitignore` (local-only per SPEC-EVO-001 R6)
- [ ] Total telemetry directory size capped at 10MB; oldest files pruned first

## Modified Files

### Go Code
- `internal/telemetry/recorder.go`: NEW - Skill usage recording
- `internal/telemetry/outcome.go`: NEW - Outcome signal heuristics
- `internal/telemetry/report.go`: NEW - Report generation
- `internal/telemetry/types.go`: NEW - Record types and constants
- `internal/cli/telemetry.go`: NEW - CLI subcommand
- `internal/hook/post_tool.go`: Add Skill tool interception
- `internal/hook/stop.go`: Add outcome determination call
- `internal/hook/session_start.go`: Add telemetry cleanup (retention enforcement)
- `cmd/moai/main.go`: Register telemetry subcommand

### Tests
- `internal/telemetry/recorder_test.go`: NEW
- `internal/telemetry/outcome_test.go`: NEW
- `internal/telemetry/report_test.go`: NEW

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Telemetry slows down hooks | Increased hook latency | Async write with buffer flush; <1ms per record |
| Outcome heuristic is inaccurate | Bad evolution proposals | Conservative default (unknown); require high confidence for non-unknown |
| File growth unbounded | Disk usage | 90-day retention + 10MB cap with LRU pruning |
| PII leakage in context_hash | Privacy violation | SHA-256 is irreversible for 8-char truncation; audit in tests |

## Dependencies

- SPEC-EVO-001: `.moai/evolution/telemetry/` directory and .gitignore entry

## Non-Goals

- Real-time telemetry dashboard (future SPEC)
- Cross-project telemetry aggregation
- Sending telemetry to external services (all local)
- Skill recommendation engine (that's SPEC-REFLECT-001)
