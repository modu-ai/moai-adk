# SPEC-REFLECT-001: Reflective Write Hook

## Meta

- **Status**: Draft
- **Wave**: 2 (depends on TELEMETRY-001)
- **Created**: 2026-04-11
- **Origin**: Memento-Skills paper (P2: Read-Write Reflective Learning - Write phase)
- **Blocked By**: SPEC-TELEMETRY-001 (needs usage data), SPEC-EVO-001 (storage + markers)

## Objective

Implement the "Write phase" of the Read-Write Reflective Learning framework from Memento-Skills. After each session, automatically analyze task outcomes and propose skill evolution entries. All proposals require user approval (confirmed user choice). Adopts the 5-layer safety architecture from Agency v3.2 constitution.

## Background

Currently moai-adk captures learnings only when:
- User explicitly corrects agent behavior (Lessons Protocol in moai-constitution.md)
- Agency evaluator identifies patterns (Agency learner, evolvable zone only)

The Memento-Skills paper shows that **every task completion** is a learning opportunity. The "Write phase" analyzes what worked, what didn't, and proposes skill updates. Combined with telemetry data from SPEC-TELEMETRY-001, this enables evidence-based skill evolution.

## Requirements (EARS Format)

### R1: Session-End Analysis [EVENT]

WHEN a session ends (Stop hook), the system SHALL analyze the session's skill usage and task outcomes to generate evolution proposals.

**Analysis Flow:**
```
Stop hook fires
    |
    v
Load session telemetry from SPEC-TELEMETRY-001
    |
    v
Identify patterns:
  - Skills with high error rate this session
  - Skill sequences that led to success
  - Missing skills (errors without skill coverage)
  - Rationalizations observed (skipped skills)
    |
    v
Generate proposal candidates
    |
    v
Apply 5-layer safety filters
    |
    v
Store proposals in .moai/evolution/learnings/
```

**Acceptance Criteria:**
- [ ] `internal/hook/reflective_write.go`: NEW - Session analysis logic
- [ ] Analysis runs only if session had >= 3 tool invocations (skip trivial sessions)
- [ ] Proposals stored as `LEARN-YYYYMMDD-NNN.md` in `.moai/evolution/learnings/`
- [ ] Each proposal includes: observation, evidence (telemetry references), proposed change, confidence score
- [ ] Maximum 3 proposals per session (prevent proposal flood)

### R2: 5-Layer Safety Architecture [UBIQ]

The Reflective Write system SHALL implement all 5 safety layers from Agency v3.2 constitution, adapted for moai-core.

**Layer Implementation:**

**L1 - Frozen Guard:**
- Never propose changes to: moai-constitution.md, CLAUDE.md core identity (Sections 1-4), agent-common-protocol.md
- Never modify YAML frontmatter (name, allowed-tools, paths)
- Only evolvable zones (within markers) are valid proposal targets

**L2 - Canary Check:**
- Before proposing, verify the change would not contradict existing HARD rules
- Check: does proposed rationalization/red-flag conflict with existing content?

**L3 - Contradiction Detector:**
- Scan existing evolvable zone content for contradicting proposals
- If contradiction found: flag both old and new, include in proposal for user review

**L4 - Rate Limiter:**
- Maximum 3 proposals per session
- Maximum 10 proposals per week
- Minimum 1 session between proposals for the same skill
- Cooldown: 24 hours between evolution applications to the same file

**L5 - Human Oversight (confirmed: always required):**
- Every proposal requires explicit user approval
- Proposal presented via next session's SessionStart context injection
- Format: before/after diff + evidence summary + approve/reject

**Acceptance Criteria:**
- [ ] `internal/evolution/safety.go`: NEW - 5-layer safety filter implementation
- [ ] L1: Frozen path list maintained as constant (not configurable)
- [ ] L2: Canary check reads target file, verifies no HARD rule conflict
- [ ] L3: Contradiction detector scans existing evolvable zones
- [ ] L4: Rate state persisted in `.moai/evolution/manifest.yaml`
- [ ] L5: Proposals stored as pending; presented on next SessionStart

### R3: Graduation Protocol [EVENT]

WHEN a learning reaches the "rule" tier (5+ observations with confidence >= 0.80), the system SHALL propose graduation to an evolvable zone.

**Graduation Tiers:**
| Observations | Tier | Action |
|---|---|---|
| 1x | Observation | Logged, no proposal |
| 3x | Heuristic | Promoted, may influence suggestions |
| 5x (confidence >= 0.80) | Rule | Eligible for graduation proposal |
| 10x | High-confidence | Auto-proposed (still needs L5 approval) |
| 1x (critical failure) | Anti-Pattern | Immediately flagged, blocks similar patterns |

**Acceptance Criteria:**
- [ ] `internal/evolution/graduation.go`: NEW - Graduation logic
- [ ] Observation count tracked in `.moai/evolution/manifest.yaml` per learning
- [ ] Confidence score calculated from: outcome consistency (0.4), frequency (0.3), recency (0.3)
- [ ] Graduation proposal includes: target file, target evolvable zone id, before/after diff
- [ ] Anti-pattern detection on critical failure (outcome=error AND task required full restart)

### R4: Learning Entry Schema [UBIQ]

Each learning entry SHALL follow a structured markdown format.

**Schema:**
```markdown
# LEARN-20260411-001

- **Status**: observation | heuristic | rule | graduated | anti-pattern
- **Skill**: moai-workflow-tdd
- **Zone**: rationalizations
- **Observations**: 3
- **Confidence**: 0.65
- **Created**: 2026-04-11
- **Last Seen**: 2026-04-11

## Observation

[Description of the pattern observed]

## Evidence

- Session abc123 (2026-04-11): TDD skill loaded but RED phase skipped, outcome=error
- Session def456 (2026-04-10): Same pattern, outcome=partial
- Session ghi789 (2026-04-09): TDD with full RED phase, outcome=success

## Proposed Change

**Target**: moai-workflow-tdd SKILL.md
**Zone**: rationalizations
**Addition**:
| "The existing tests cover this case" | Existing tests verify old behavior. New behavior needs its own test to prove it works. |
```

**Acceptance Criteria:**
- [ ] Template for learning entries in `internal/evolution/templates/learning.md`
- [ ] Parser validates entry schema before storage
- [ ] Maximum 50 active learnings (archive older entries per Lessons Protocol)
- [ ] Duplicate detection: check if observation matches existing learning before creating new

### R5: Proposal Presentation [EVENT]

WHEN a session starts and pending proposals exist, the SessionStart hook SHALL present them to the user.

**Acceptance Criteria:**
- [ ] SessionStart hook checks `.moai/evolution/learnings/` for status=rule or status=high-confidence
- [ ] Presents top 3 pending proposals via `additionalContext` in SessionStart response
- [ ] Format: skill name, observation summary, proposed change (before/after)
- [ ] User can approve via natural language ("approve LEARN-001") or reject ("reject LEARN-001")
- [ ] Approved proposals: apply to evolvable zone, update manifest, mark as graduated
- [ ] Rejected proposals: mark as rejected, do not re-propose for 30 days

## Modified Files

### Go Code
- `internal/hook/reflective_write.go`: NEW - Session analysis + proposal generation
- `internal/hook/stop.go`: Add reflective write call
- `internal/hook/session_start.go`: Add proposal presentation
- `internal/evolution/safety.go`: NEW - 5-layer safety filters
- `internal/evolution/graduation.go`: NEW - Graduation protocol
- `internal/evolution/learning.go`: NEW - Learning entry CRUD
- `internal/evolution/types.go`: NEW - Types and constants
- `internal/evolution/apply.go`: NEW - Apply graduated proposals to evolvable zones

### Templates
- `internal/template/templates/.moai/evolution/learnings/.gitkeep`: Placeholder

### Tests
- `internal/evolution/safety_test.go`: NEW - Safety filter tests
- `internal/evolution/graduation_test.go`: NEW - Graduation logic tests
- `internal/evolution/learning_test.go`: NEW - Learning CRUD tests
- `internal/hook/reflective_write_test.go`: NEW - Analysis pipeline tests

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Proposal quality is low | User fatigue from bad proposals | Conservative heuristics; 3/session cap; high confidence threshold |
| Safety filters too strict | No proposals ever generated | Start with L1+L4+L5; add L2+L3 incrementally |
| Session analysis adds latency to Stop hook | Slow session exit | Async processing; <500ms target |
| Anti-pattern false positive | Valid pattern blocked | Anti-pattern requires critical failure + full restart (very conservative) |

## Dependencies

- SPEC-EVO-001: `.moai/evolution/` directory, manifest.yaml, evolvable zone merge
- SPEC-TELEMETRY-001: Skill usage data for session analysis

## Non-Goals

- Automatic proposal application without user approval
- Cross-project learning sharing
- LLM-based analysis (pure heuristic for v1; LLM analysis is future enhancement)
- Modifying non-evolvable-zone content
