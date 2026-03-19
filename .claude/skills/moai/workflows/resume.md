---
name: moai-workflow-resume
description: >
  Resume interrupted SPEC work from journal checkpoint. Automatically detects
  the last interrupted session and synthesizes recovery context.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-19"
  tags: "resume, recovery, checkpoint, session, context"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

triggers:
  keywords: ["resume", "이어서", "continue", "recover", "복원"]
  agents: []
  phases: ["resume"]
---

# Resume Workflow Orchestration

## Purpose

Resume interrupted SPEC work by reading journal checkpoints and progress data.
Provides zero-friction context recovery after token exhaustion, terminal crashes,
or context compaction.

## Input

- $ARGUMENTS: Optional SPEC-ID (e.g., SPEC-AUTH-001)
- No arguments: Auto-detect most recent interrupted SPEC

## Phase Sequence

### Phase 1: Detect Resumable Work

If SPEC-ID provided:
- Load .moai/specs/SPEC-{ID}/journal.jsonl
- Load .moai/specs/SPEC-{ID}/progress.md
- Verify resumable state exists

If no SPEC-ID:
- Scan .moai/specs/*/journal.jsonl for interrupted sessions
- Sort by most recent, present top 3 to user via AskUserQuestion
- Options: each resumable SPEC with last phase and end reason

### Phase 2: Build Resume Context

From journal.jsonl (most recent entries):
- Last session ID, end reason, tokens used
- Last phase and checkpoint status
- Files modified, next planned action

From progress.md:
- Phase completion history
- Acceptance criteria status

Synthesize into a structured resume summary.

### Phase 3: Present and Confirm

Tool: AskUserQuestion

Display resume context to user:
```
Resume SPEC-AUTH-001
- Last session: 2026-03-19 14:30 (ended: token exhaustion)
- Progress: Phase 2B, 3/5 AC completed
- Next: TASK-003 (Auth middleware error handling)
- Files: internal/auth/middleware.go, middleware_test.go
```

Options:
- **Resume from checkpoint** (Recommended): Continue with /moai run SPEC-{ID}
- **Resume from phase start**: Restart the current phase from scratch
- **View full journal**: Display complete session history
- **Cancel**: Exit without resuming

### Phase 4: Execute Resume

If "Resume from checkpoint":
- Execute: /moai run SPEC-{ID} with resume context injected
- Journal records new session_start with "resumed_from" context

If "Resume from phase start":
- Reset current phase in progress.md
- Execute: /moai run SPEC-{ID}

---

## Completion Criteria

- Resumable SPEC detected and context synthesized
- User confirmed resume action
- /moai run dispatched with recovery context

---

Version: 1.0.0
Updated: 2026-03-19
