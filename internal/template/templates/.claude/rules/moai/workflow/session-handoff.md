# Session Handoff Protocol

Long-running session continuity guidance: enables clean transitions across context boundaries via paste-ready resume messages.

## Why This Matters

Long workflows (multi-SPEC waves, multi-milestone implementation, prolonged debugging) accumulate context that may exceed the conversation window or simply benefit from a fresh start. Without a standardized handoff format, session boundaries lose work-in-progress state and force the next session to rediscover context manually.

This rule establishes:
- When the orchestrator MUST emit a paste-ready resume message
- The canonical 6-line format that the next session can execute without modification
- Auto-memory integration so the message persists across `/clear`

## When To Generate (5 Triggers)

[HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:

| # | Trigger | Detection |
|---|---------|-----------|
| 1 | Context usage crosses 75% (accumulated input+output) | Heuristic per `.claude/rules/moai/workflow/context-window-management.md` §Detection Heuristics |
| 2 | SPEC phase completion (plan/run/sync) within a multi-SPEC workflow | Workflow §Phase 4 Completion entry |
| 3 | User explicitly requests session end ("세션 종료", "이번 세션 마무리", "next session") | Intent detection in user message |
| 4 | PR creation success when more SPECs remain in the current wave | After `gh pr create` success + memory indicates >0 pending SPECs |
| 5 | Long-running multi-milestone task reaches a stable checkpoint | After milestone Mn complete + Mn+1 not yet started |

When NONE of these apply (single-turn, trivial task, read-only query), the orchestrator emits a brief completion confirmation without paste-ready format. Forcing the format on trivial tasks is anti-pattern.

## Canonical Format (Verbatim Spec)

[HARD] Resume message MUST follow this exact 6-block structure:

```
ultrathink. <SPEC-ID> <phase> 진입.
applied lessons: <memory-file-1>, <memory-file-2>, ...

전제 검증:
1) <verifiable precondition 1>
2) <verifiable precondition 2>
N) <verifiable precondition N>

실행: <command-or-action>

머지 후: <next-action-or-spec>
```

### Field-by-Field Specification

**Block 1 (Line 1)**: `ultrathink. <SPEC-ID> <phase> 진입.`
- `ultrathink.` — keyword that triggers Adaptive Thinking max effort on Opus 4.7+ (CLAUDE.md §12). Required: max effort is the safe default for handoff continuation since the next session lacks accumulated reasoning context.
- `<SPEC-ID>` — target SPEC identifier (e.g., `SPEC-V3R2-WF-004`) or workflow target (`다음 SPEC plan 작성`).
- `<phase>` — `plan` | `run` | `sync` | `loop`. Korean OK (`plan phase`, `run 진입`).

**Block 2 (Line 2)**: `applied lessons: <comma-separated memory file references>`
- Comma-separated list of relevant memory files from `~/.claude/projects/{hash}/memory/`.
- File names without `.md` extension acceptable (`project_wave6_wf002_complete`).
- MUST include the most recent relevant project memory.
- MUST include any relevant lessons (`lessons #9 wave-split`).

**Block 3 (separator + header)**: blank line, then `전제 검증:` (Korean) or `Preconditions:` (English).

**Block 4 (Lines 4..N)**: Numbered preconditions, each verifiable via a single shell or `gh` command:
- Format: `<N>) <action> → <expected outcome>`
- Each precondition MUST be independently verifiable (git command, gh command, file existence, etc.).
- Maximum 4 preconditions (cognitive-load constraint).

**Block 5 (separator)**: blank line, then `실행: <command-or-action>`.
- Single primary action the next session begins with.
- Typically a `/moai <subcommand> <args>` invocation.
- Or a structured directive (`manager-tdd 위임으로 M1→M5 순차 진행`).

**Block 6 (separator)**: blank line, then `머지 후: <next-action-or-spec>`.
- Optional but RECOMMENDED for multi-SPEC waves.
- Specifies the next SPEC or workflow after the current target completes.

### Verified Example (SPEC-V3R2-WF-002 session, 2026-05-01)

```
ultrathink. SPEC-V3R2-WF-002 implementation 진입.
applied lessons: project_wave6_wf002_plan_ready, lessons #9 wave-split.

전제 검증:
1) git log --oneline -1 → 01801c922 확인
2) ls .moai/specs/SPEC-V3R2-WF-002/ in worktree → 5 files
3) git -C worktree status → clean, base origin/main

실행: /moai run SPEC-V3R2-WF-002

머지 후: WF-004 → WF-003 → WF-005
```

This format is paste-ready: the next session reads each line and executes verification + main action without further interpretation. The format was used verbatim in the WF-002 session and proved to recover full context within the first 3 turns of the receiving session.

## Auto-Memory Integration (Mandatory)

[HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<wave>_<spec>_<status>.md` (e.g., `project_wave6_wf002_complete.md`).
2. Include the resume message verbatim in that file under a `## 다음 세션 시작점 (paste-ready resume message)` heading.
3. Update `MEMORY.md` index with a one-line entry pointing to the new memory file.
4. Mark superseded entries (if any) with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.

This ensures the message survives `/clear` and is discoverable by the next session at the start of its context.

## Output Surface (User-Facing)

When emitting the resume message at session end, the orchestrator MUST display:

1. The message inside a fenced code block (```text ... ```) so the user can paste verbatim.
2. The memory file path where the message is persisted.
3. A one-sentence summary of what the next session will continue.

Example output structure (Korean conversation_language):

````markdown
**다음 세션 paste-ready resume** (memory: `project_wave6_wf002_complete.md`)

```text
ultrathink. SPEC-V3R2-WF-004 plan phase 진입.
applied lessons: project_wave6_wf002_complete (PR #761 머지 완료 가정), lessons #9 wave-split.

전제 검증:
1) git log --oneline -2 → SPEC-V3R2-WF-002 머지 commit 확인
2) gh pr view <PR-number> → MERGED 상태 확인

실행: /moai plan SPEC-V3R2-WF-004

머지 후: WF-003 → WF-005
```

(다음 세션은 위 메시지를 그대로 붙여넣어 시작합니다.)
````

## Anti-Patterns

- Free-form prose handoff ("다음 세션에서 이어서 하시면 됩니다") — gives the next session no executable context.
- Resume message without preconditions — next session has no way to detect state drift before acting.
- Resume message without `ultrathink.` — fails to activate max effort for complex continuation; trivial-mode reasoning may miss accumulated nuance.
- Resume message saved only in chat output, not in auto-memory — lost across `/clear`, defeating the purpose.
- Multiple memory entries for the same context without `[SUPERSEDED by ...]` markers — index pollution; next session cannot identify the canonical entry.
- Forcing the format on trivial tasks (typo fix, single config edit) — pollutes memory with noise.

## Cross-references

- `.claude/rules/moai/workflow/context-window-management.md` — 75% threshold detection heuristics, broader long-horizon session continuity policy
- `output-styles/moai/moai.md` §6 (Persistence & Context Awareness) — orchestrator persistence pattern
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory write rules and `[SUPERSEDED by ...]` convention
- CLAUDE.md §11 (Error Handling) — token-limit recovery flow
- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split rationale that often precedes a session handoff

---

Source: 2026-05-01 SPEC-V3R2-WF-002 session evidence (verified 6-block format)
Status: HARD operational rule, applies to all multi-phase MoAI workflows
