# Session Handoff Protocol

Long-running session continuity: clean transitions across context boundaries via paste-ready resume messages.

> **Loading scope**: Intentionally always-loaded (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context, including those without SPEC files. The ~1,400-token cost is justified by cross-cutting applicability.

## Why This Matters

Long workflows (multi-SPEC waves, multi-milestone implementation) accumulate context that exceeds the window or benefits from fresh start. Without a standardized handoff, session boundaries lose work-in-progress. This rule defines when to emit a paste-ready resume, the 6-block structure, and auto-memory integration that persists across `/clear`.

## When To Generate (5 Triggers)

[ZONE:Evolvable] [HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:

| # | Trigger | Detection |
|---|---------|-----------|
| 1 | Context usage crosses model-specific threshold (cumulative input+output) | **1M context model (Opus 4.7): 50%** (~500,000 tokens). **200K context model (Sonnet/Opus standard, Haiku): 90%** (~180,000 tokens). Heuristic per `.claude/rules/moai/workflow/context-window-management.md` §Detection Heuristics. |
| 2 | SPEC phase completion (plan/run/sync) within a multi-SPEC workflow | Phase boundary in `.claude/rules/moai/workflow/spec-workflow.md` §Completion Markers (after plan/run/sync phase finishes within a multi-SPEC SPEC ID series) |
| 3 | User explicitly requests session end ("세션 종료", "이번 세션 마무리", "next session") | Intent detection in user message |
| 4 | PR creation success when more SPECs remain in the current wave | After `gh pr create` success + memory indicates >0 pending SPECs |
| 5 | Long-running multi-milestone task reaches a stable checkpoint | After milestone Mn complete + Mn+1 not yet started |

When NONE apply (single-turn, trivial task, read-only query), emit a brief completion confirmation. The threshold in Trigger #1 reflects asymmetric stall risk: 1M models tolerate higher absolute load; 200K models hit the ceiling earlier. The `/clear` policy in `context-window-management.md` is co-anchored to the same threshold per model class.

## Canonical Format (Verbatim Spec)

[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure:

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

- **Block 1**: `ultrathink.` triggers Adaptive Thinking max effort on Opus 4.7+ (next session lacks accumulated reasoning). `<phase>` ∈ `plan | run | sync | loop`.
- **Block 2**: `applied lessons:` — relevant memory files from `~/.claude/projects/{hash}/memory/`. MUST include the most recent relevant project memory + any relevant lessons.
- **Block 3**: separator + `전제 검증:` (Korean) or `Preconditions:` (English).
- **Block 4**: numbered preconditions `<N>) <action> → <expected outcome>`. Each MUST be independently verifiable (git/gh command, file existence). Max 4 preconditions.
- **Block 5**: separator + `실행: <command-or-action>` — single primary action (typically `/moai <subcommand>`).
- **Block 6**: separator + `머지 후: <next-action-or-spec>` — RECOMMENDED for multi-SPEC waves.

### Example (Illustrative; substitute project-specific values when adapting)

```
ultrathink. SPEC-MYPROJ-001 implementation 진입.
applied lessons: project_wave6_myproj001_plan_ready, lessons #9 wave-split.

전제 검증:
1) git log --oneline -1 → <commit-sha> 확인
2) ls .moai/specs/SPEC-MYPROJ-001/ → N files

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002 → SPEC-MYPROJ-003
```

## Auto-Memory Integration (Mandatory)

[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<wave>_<spec>_<status>.md` (e.g., `project_wave6_wf002_complete.md`).
2. Include the resume message verbatim in that file under a `## 다음 세션 시작점 (paste-ready resume message)` heading.
3. Update `MEMORY.md` index with a one-line entry pointing to the new memory file.
4. Mark superseded entries (if any) with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.

This ensures the message survives `/clear` and is discoverable at the start of the next session's context.

## Output Surface (User-Facing)

At session end, the orchestrator displays: (1) the message in a fenced ```text``` block for verbatim paste, (2) the memory file path, (3) a one-sentence summary of what next session continues.

## Anti-Patterns

- Free-form prose handoff — no executable context.
- Resume without preconditions — next session cannot detect state drift.
- Resume without `ultrathink.` — fails to activate max effort.
- Resume saved only to chat, not auto-memory — lost across `/clear`.
- Duplicate memory entries without `[SUPERSEDED by ...]` markers — index pollution.
- Forcing the format on trivial tasks — memory noise.

## Worktree-Anchored Resume Pattern

[ZONE:Evolvable] [HARD] When the SPEC was initialized via L3 `/moai plan --worktree` (creating an L2 SPEC worktree at `~/.moai/worktrees/<project>/<spec-or-name>/`), the resume message MUST include **Block 0 (cwd anchoring)** prepended before the standard 6-block structure. Without Block 0, the next session starts in main project cwd by default, breaking L2 SPEC worktree isolation expectations.

> Per user policy 2026-05-17 (`feedback_worktree_autonomous` memory): L3 `--worktree` is **user opt-in** only. For SPECs initialized without `--worktree` (the default as of 2026-05-17), the standard 6-block structure suffices — Block 0 is NOT required.

### Why Block 0 (L3 `--worktree` opt-in only)

With L3 `--worktree`, SPEC artifacts and L1 isolation base live in a different cwd. Pasting resume into a main-cwd session causes: L1 base divergence (lessons #13), Bash commands targeting main project (lessons #12), build/test from the wrong tree. Block 0 forces a new terminal session **inside** the L2 worktree before any action.

### Block 0 Format

Block 0 is **prepended** before Block 1:

```
[New Terminal — START IN WORKTREE]
$ cd <worktree-absolute-path>
$ <launcher>     # Choose one: moai cc | moai glm | claude
   └─ Claude Code session starts here (cwd = worktree)
```

[ZONE:Evolvable] [HARD] Block 0 MUST surface the 3 primary launchers verbatim so the user can choose without consulting external docs:

1. `moai cc` — Claude Code leader with MoAI orchestration (default for normal SPEC work; supports `-p <name>` profile flag)
2. `moai glm` — cost-optimized GLM-only worker mode (no Claude Code leader, lower token cost)
3. `claude` — native Claude Code without MoAI wrapper (minimal fallback)

Advanced launchers (use only when user explicitly requests, NOT auto-surfaced in Block 0):
- `moai cc --bypass` — sandboxed-only execution (testing scenarios)
- `moai cg` — Claude leader + GLM teammates parallel mode (requires `tmux new-session -s <name>` first; pair with `--team`)

### Updated Block 4 (Preconditions)

When Block 0 is present, the **first precondition (0)** verifies compliance:

```
0) git rev-parse --show-toplevel → <worktree-path> (★ critical pre-check)
```

If verification 0) fails, stop and instruct the user to restart inside the worktree.

### Single-Session vs Multi-Session Decision

Block 0 is REQUIRED only with L3 `--worktree`. For `--branch` (or no flag — 2026-05-17 default), standard 6-block suffices because main session cwd already follows the branch.

[ZONE:Evolvable] [HARD] If L3 `--worktree` was used and the user is NOT comfortable with multi-terminal/multi-session workflow, the orchestrator SHOULD recommend `--branch` for the next SPEC. Forcing Block 0 onto a single-session user is friction without benefit. See lessons #14 for the single-session vs multi-session decision rationale.

### Example with Block 0 (Illustrative)

```
[New Terminal — START IN WORKTREE]
$ cd ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
$ moai cc        # 또는 moai glm | claude (3가지 launcher 중 선택; 본 예시는 moai cc)

ultrathink. SPEC-MYPROJ-001 Wave N 진입.
applied lessons: project_myproj_prev_wave_complete, lessons #12 #13 #14.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

머지 후: Wave N+1
```

## Cross-references

- `.claude/rules/moai/workflow/context-window-management.md` — threshold (1M = 50%, 200K = 90%) for `/clear` and Trigger #1; same table.
- `.claude/output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory + `[SUPERSEDED by ...]` convention
- CLAUDE.md §11 (Error Handling) — token-limit recovery
- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split rationale
- lessons #14 — `--worktree` Block 0 + single/multi-session decision
- lessons #12, #13 — worktree isolation + --team base mismatch

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
