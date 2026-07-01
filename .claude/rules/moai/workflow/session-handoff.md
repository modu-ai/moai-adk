# Session Handoff Protocol

Long-running session continuity: clean transitions across context boundaries via paste-ready resume messages.

> **Loading scope**: Intentionally always-loaded (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context, including those without SPEC files. The ~5,900-token cost is justified by cross-cutting applicability.

## Why This Matters

Long workflows (multi-SPEC waves, multi-milestone implementation) accumulate context that exceeds the window or benefits from fresh start. Without a standardized handoff, session boundaries lose work-in-progress. This rule defines when to emit a paste-ready resume, the 6-block structure, and auto-memory integration that persists across `/clear`.

## When To Generate (5 Triggers)

[ZONE:Evolvable] [HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:

| # | Trigger | Detection |
|---|---------|-----------|
| 1 | Context usage crosses model-specific threshold (cumulative input+output) | Model-specific percentage threshold (1M-context models vs 200K-context models) — see `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets for the per-model-class threshold table (the authoritative SSOT for the numeric thresholds; this file carries no inline model-class numbers to avoid label drift). |
| 2 | SPEC phase completion (plan/run/sync) within a multi-SPEC workflow | Phase boundary in `.claude/rules/moai/workflow/spec-workflow.md` §Phase Transitions (after plan/run/sync phase finishes within a multi-SPEC SPEC ID series) |
| 3 | User explicitly requests session end ("세션 종료", "이번 세션 마무리", "next session") | Intent detection in user message |
| 4 | PR creation success when more SPECs remain in the current wave | After `gh pr create` success + memory indicates >0 pending SPECs |
| 5 | Long-running multi-milestone task reaches a stable checkpoint | After milestone Mn complete + Mn+1 not yet started |

When NONE apply (single-turn, trivial task, read-only query), emit a brief completion confirmation. The threshold in Trigger #1 reflects asymmetric stall risk: 1M models tolerate higher absolute load; 200K models hit the ceiling earlier. The `/clear` policy in `context-window-management.md` is co-anchored to the same threshold per model class.

## Canonical Format (Verbatim Spec)

[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers** (see § Cut-line Marker Specification below for the literal marker format, Unicode-preservation rules, and locale translation contract). Cut-line markers sit **inside** the fenced text block alongside the content so they are copied verbatim with the message; this provides the user an unambiguous copy boundary in long terminal scrollback:

```
✂──── 여기부터 복사 ────✂

ultrathink. <SPEC-ID> <phase> <entering verb>.
# /effort ultracode   ← emit ONLY when the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams); omit otherwise (per Field-by-Field Spec, Block 1).
# /goal <completion-condition>   ← emit ONLY when the next SPEC is run-phase AND has a machine-verifiable end-state (e.g. the SPEC's test suite passes AND lint is clean, or stop after N turns); omit otherwise (per Field-by-Field Spec, Block 1). A /goal line does NOT authorize autonomous run-phase entry — Implementation Kickoff Approval still required.
applied lessons: <memory-file-1>, <memory-file-2>, ...

전제 검증:
1) <verifiable precondition 1>
2) <verifiable precondition 2>
N) <verifiable precondition N>

실행: <command-or-action>

머지 후: <next-action-or-spec>

✂──── 여기까지 복사 ────✂
```

### Cut-line Marker Specification

- Top marker: `✂──── 여기부터 복사 ────✂` (scissors U+2702 + 4× U+2500 + space + text + space + 4× U+2500 + scissors)
- Bottom marker: `✂──── 여기까지 복사 ────✂` (same structure, text differs)
- One blank line separates each marker from adjacent block content (top → blank → Block 1; Block 6 → blank → bottom)
- `✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute
- Box-drawing characters (`─` U+2500) preserved verbatim
- Marker text translates per `conversation_language` (see Localization table below)

### Localization Table

The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table is the SSOT for the locale renderings (the canonical skeleton uses the `<entering verb>` / `<header>` placeholders; concrete locale renderings live here). Cross-verified for consistency with `.claude/output-styles/moai/moai.md §8` (the canonical render surface).

| Element | English | Korean (canonical) | Japanese | Chinese |
|---------|---------|--------------------|----------|---------|
| Cut-line top text | `Copy from here` | `여기부터 복사` | `ここからコピー` | `从这里复制` |
| Cut-line bottom text | `Copy to here` | `여기까지 복사` | `ここまでコピー` | `到这里复制` |
| Block 1 entering verb | `entering` | `진입` | `開始` | `进入` |
| Block 3 Preconditions header | `Preconditions:` | `전제 검증:` | `前提条件:` | `前提条件:` |
| Block 5 Run header | `Run:` | `실행:` | `実行:` | `执行:` |
| Block 6 After-merge header (PR workflow) | `After merge:` | `머지 후:` | `マージ後:` | `合并后:` |
| Block 6 Follow-up header (trunk no-PR) | `Follow-up:` | `후속:` | `後続:` | `后续:` |

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers) while keeping `✂` and `─` characters verbatim, and substitute the locale rendering for each Block 1/3/5/6 placeholder when emitting the paste-ready message.

**Fallback rule for locales not in the table.** The table above lists concrete renderings for en / ko / ja / zh only. When `conversation_language` is an ISO-639 code whose language column is NOT in this table (e.g. `fr`, `de`, `es`, `pt`, `vi`), English is the canonical fallback skeleton and each label translates to that locale using the naturalization principle (idiomatic phrasing a native reader expects, never literal word-by-word transliteration). In other words: locales not in the table fall back to the English column for the structural skeleton, with the label text rendered in the configured ISO-639 language — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

### Field-by-Field Specification

- **Block 1**: `ultrathink.` sets `effort: xhigh` on Opus 4.7+ (next session lacks accumulated reasoning). Adaptive Thinking is a DISTINCT axis — the thinking mode, explicitly enabled via `thinking: {type: "adaptive"}` — not something `ultrathink` toggles. `<phase>` ∈ `plan | run | sync | mx`.
  - **Purpose-conditional `/effort ultracode` re-set line** [HARD]: Block 1 also carries a purpose-conditional `/effort ultracode` re-set line, emitted ONLY when the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams). The line sits immediately after `ultrathink.`. Per `.claude/rules/moai/workflow/dynamic-workflows.md`, ultracode is NOT restored by the `ultrathink.` opener — it must be explicitly re-issued after `/clear` when the resumed session needs auto-orchestration. When the next SPEC does NOT need workflow fan-out, the ultracode line is omitted (the `ultrathink.` opener alone suffices). Default on ambiguity: omit.
  - **Purpose-conditional `/goal` re-set line** [HARD]: Block 1 also carries a purpose-conditional `/goal <completion-condition>` re-set line, authored with the same conditional-emit mechanism as the `/effort ultracode` line above. It is emitted ONLY when the next SPEC's phase is run-phase AND the next SPEC declares a mechanically verifiable completion condition (a machine-checkable end-state such as the SPEC's test suite passing, a lint-clean state, or a bounded `stop after N turns` clause). The line sits immediately after the `/effort ultracode` line. Per `.claude/rules/moai/workflow/goal-directive.md`, a `/goal` is NOT restored by the `ultrathink.` opener — `/clear` removes an active goal, so it must be explicitly re-issued after `/clear` when the resumed session needs the autonomous-continuation loop. The renderer omits the `/goal` line for a plan-phase or sync-phase next SPEC, and for any next SPEC lacking a machine-verifiable end-state. Default on ambiguity: omit — the identical default binding the `/effort ultracode` line. **Implementation Kickoff Approval invariant**: a `/goal` line does NOT authorize autonomous run-phase entry; the Implementation Kickoff Approval human gate (orchestrator `AskUserQuestion` per `.claude/rules/moai/workflow/goal-directive.md` § MoAI Integration Notes) remains required before run-phase entry, independent of whether a `/goal` line is present. The `/goal` line is a continuation-loop convenience, never a run-phase pre-authorization.
- **Block 2**: `applied lessons:` — relevant memory files from `~/.claude/projects/{hash}/memory/`. MUST include the most recent relevant project memory + any relevant lessons. Block 2 MUST also include a `source_session_id: <UUID from moai session current>` line carrying the Claude Code session_id of the orchestrator turn that generated this resume message per the canonical multi-session coordination policy. The session_id is the same value emitted by `moai session list --json` and stored in `.moai/state/active-sessions.json` — readers can correlate the resume back to its originating session.
  - **Environment fallback** [HARD]: the primary UUID source is `moai session current`. If `moai session current` returns the canonical fallback (runtime did not expose session.id to the CLI subprocess), OR `moai session list --json` returns error (CLI not installed in PATH), OR `.moai/state/active-sessions.json` does not exist (the multi-session coordination layer not yet deployed in this project), the orchestrator MUST emit the recognized fallback pattern verbatim: `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`. This pattern is NOT an anti-pattern; it is the prescribed graceful degradation when the CLI/registry layer is absent or the runtime does not expose session.id. The next session, upon `/moai session register` activation, MAY backfill the UUID by appending a `[backfilled: <UUID>]` annotation to the memory file's Block 2 line.
- **Block 3**: separator + `전제 검증:` (Korean) or `Preconditions:` (English).
- **Block 4**: numbered preconditions `<N>) <action> → <expected outcome>`. Each MUST be independently verifiable (git/gh command, file existence). Max 4 preconditions.
- **Block 5**: separator + `실행: <command-or-action>` — single primary action (typically `/moai <subcommand>`).
- **Block 6**: separator + `<workflow-context header>: <next-action-or-spec>` — RECOMMENDED for multi-SPEC waves or follow-up; **omit entirely** for single-SPEC close with no further actions queued.
  - **Header selection (workflow-context conditional)**:
    - **PR-based workflow** (feat/* → PR → merge): `머지 후:` (en `After merge:`)
    - **Trunk-based no-PR** (e.g., 1-person OSS, all-tier direct-to-main push, no merge step): `후속:` (en `Follow-up:`)
    - **Single-SPEC close** (no further SPEC/phase queued): omit Block 6 entirely
  - **Single action principle**: `<next-action-or-spec>` MUST be one concrete SPEC ID, one command, or one phase transition — avoid vague "cycle-repeat" / "iteration loop" phrasing that reads as infinite recursion.

### Example (Illustrative; substitute project-specific values when adapting)

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-MYPROJ-001 implementation 진입.
# /goal the SPEC's test suite passes AND lint is clean, or stop after 20 turns   ← run-phase + machine-verifiable; omit for plan/sync or non-verifiable.
applied lessons: project_sprint6_myproj001_plan_ready, lessons #9 wave-split.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

전제 검증:
1) git log --oneline -1 → <commit-sha> 확인
2) ls .moai/specs/SPEC-MYPROJ-001/ → N files

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002 → SPEC-MYPROJ-003

✂──── 여기까지 복사 ────✂
```

## Auto-Memory Integration (Mandatory)

[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<sprint>_<spec>_<status>.md` (e.g., `project_sprint6_wf002_complete.md`). The `<sprint>` token reflects the multi-SPEC time-unit grouping per `.claude/rules/moai/development/sprint-round-naming.md` (the legacy `<wave>` token is retired per AP-SRN-004).
2. Include the resume message verbatim in that file under a `## 다음 세션 시작점 (paste-ready resume message)` heading.
3. Update `MEMORY.md` index with a one-line entry pointing to the new memory file.
4. Mark superseded entries (if any) with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.
5. Annotate the MEMORY.md index entry with a `(session: <UUID-8-char-prefix>)` parenthetical when the SPEC was worked across multiple sessions (cross-references the `source_session_id` in Block 2 — enables readers to correlate the resume back to its originating session).

This ensures the message survives `/clear` and is discoverable at the start of the next session's context.

## Output Surface (User-Facing)

At session end, the orchestrator displays: (1) the message in a fenced ```text``` block **bounded by cut-line markers** (per § Cut-line Marker Specification — marker text translated per `conversation_language`, `✂`/`─` symbols preserved verbatim) for verbatim paste, (2) the memory file path, (3) a one-sentence summary of what next session continues.

## Anti-Patterns

> See also: § Diet Constraints / Anti-pattern catalogue (paste-ready budget violations AP-D-001..005) and § V0 Abort Gate Doctrine / Anti-pattern (abort-gate violations AP-V-001..004). This list covers general resume-hygiene patterns; the Diet and V0 lists cover their respective specialized domains.

### Anti-Pattern Index (consolidated)

The table below is the single navigational index for every anti-pattern code defined in this file. Each row links forward to the detail section that carries the domain context; the index does NOT duplicate the prose. This is the canonical single-source entry point — when a code is referenced elsewhere, link to this index, not to the detail section directly.

| Code | Concern | Detail section |
|------|---------|----------------|
| (general hygiene) | Free-form prose handoff — no executable context | § Anti-Patterns (general list below this index) |
| (general hygiene) | Resume without preconditions — next session cannot detect state drift | § Anti-Patterns |
| (general hygiene) | Resume without `ultrathink.` — fails to activate xhigh effort | § Anti-Patterns |
| (general hygiene) | Resume saved only to chat, not auto-memory — lost across `/clear` | § Anti-Patterns |
| (general hygiene) | Duplicate memory entries without `[SUPERSEDED by ...]` markers | § Anti-Patterns |
| (general hygiene) | Resume Block 2 missing `source_session_id` AND the environment fallback | § Anti-Patterns |
| (general hygiene) | Forcing the format on trivial tasks — memory noise | § Anti-Patterns |
| (general hygiene) | Cut-line markers absent — user cannot identify copy boundary | § Anti-Patterns |
| (general hygiene) | Cut-line markers with translated `✂` symbol or `─` decorator | § Anti-Patterns |
| AP-D-001 | Block 2 lessons 5+ references | § Diet Constraints / Anti-pattern catalogue |
| AP-D-002 | precondition body prose (history/lesson narrative) | § Diet Constraints / Anti-pattern catalogue |
| AP-D-003 | Block 5 sub-step nesting (multi-phase 11-substep) | § Diet Constraints / Anti-pattern catalogue |
| AP-D-004 | directive escalation embedded in body (N-th "stronger directive") | § Diet Constraints / Anti-pattern catalogue |
| AP-D-005 | ceremonial reminder ("observe discipline", "exact reference") in paste-ready | § Diet Constraints / Anti-pattern catalogue |
| AP-V-001 | `ps aux` raw count `≤ 2 STRICT` as sole V0 verification | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-002 | "user promise accumulated non-fulfillment N times" body tracking after V0 FAIL | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-003 | offering a force-through option (override + spawn) in AskUserQuestion on V0 FAIL | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-004 | measuring V0-b with `lsof +D "$PWD" | grep -iE 'claude'` (false-positive defect) | § V0 Abort Gate Doctrine / Anti-pattern |

- Free-form prose handoff — no executable context.
- Resume without preconditions — next session cannot detect state drift.
- Resume without `ultrathink.` — fails to activate xhigh effort.
- Resume saved only to chat, not auto-memory — lost across `/clear`.
- Duplicate memory entries without `[SUPERSEDED by ...]` markers — index pollution.
- Resume Block 2 missing `source_session_id: <UUID from moai session current>` **AND missing the environment fallback pattern** (`<not-available — environment-fallback, ...>`) — the canonical multi-session coordination policy cannot correlate the resume back to its originating session for race attribution. The environment fallback pattern itself is NOT an anti-pattern; only the complete absence of both UUID and fallback pattern is the violation.
- Forcing the format on trivial tasks — memory noise.
- Cut-line markers absent — user cannot identify exact copy boundary in long terminal scrollback (see § Cut-line Marker Specification for the literal format).
- Cut-line markers with translated `✂` symbol or `─` decorator — contrary to § Cut-line Marker Specification (only the marker text translates; the symbols are preserved verbatim).
- Omitting the `/effort ultracode` re-set line when the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams) — the resumed session silently drops to non-ultracode effort and loses auto-orchestration (ultracode is NOT restored by `ultrathink.` per `.claude/rules/moai/workflow/dynamic-workflows.md`).
- Omitting the `/goal` re-set line when the next SPEC has a verifiable run-phase completion condition — the resumed session silently loses the autonomous-continuation loop (a `/goal` is NOT restored by `ultrathink.`; `/clear` removes an active goal, per `.claude/rules/moai/workflow/goal-directive.md`).

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

### `/cd` cache-preserving alternative (CC 2.1.169+)

The new-terminal Block 0 above is a cold-start path: it opens a fresh Claude Code session inside the L2 worktree, which re-reads skills/rules from scratch. Claude Code 2.1.169+ ships a `/cd` command that changes the session's working directory **while preserving the prompt cache** — so the in-flight reasoning context survives the cwd switch instead of being rebuilt. For an L2 worktree resume where you want to keep the current session's accumulated context (rather than cold-starting), `/cd <worktree-absolute-path>` is a cache-preserving complement to the new-terminal Block 0. This note does NOT replace Block 0 — the new-terminal path remains the default for clean isolation; `/cd` is the lower-friction option when cache preservation matters more than a fresh tree.

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
✂──── 여기부터 복사 ────✂

[New Terminal — START IN WORKTREE]
$ cd ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
$ moai cc        # 또는 moai glm | claude (3가지 launcher 중 선택; 본 예시는 moai cc)

ultrathink. SPEC-MYPROJ-001 Epic N 진입.
applied lessons: project_myproj_prev_sprint_complete, lessons #12 #13 #14.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

후속: Milestone M<N+1> (single-SPEC next step) 또는 Epic N+1 (multi-SPEC next grouping)

✂──── 여기까지 복사 ────✂
```

## Diet Constraints

[ZONE:Evolvable] [HARD] A paste-ready resume message is "next session minimum executable context" — it is NOT an audit trail, history record, or ceremonial commitment record. Accumulating history/lesson/directive-escalation prose in the body via append-only across retry iterations is an empirically proven anti-pattern.

### Block 2 applied-lessons constraint

- At most **4 references** (memory file slug or lesson identifier)
- Each reference is a **single-line identifier** (e.g. `L52#33`, `L_NEW_V0_ABORT_GATE` — full prose history is prohibited)
- Five or more is an anti-pattern → move the surplus into the memory file body

### Block 4 precondition constraint

- Each precondition targets **≤ 200 chars** (practical readability limit)
- Format: `N) <verifiable command> → <expected outcome>`
- History tracking / lesson narrative / cumulative-pattern prose is prohibited
- Multi sub-command (V0a/V0b/V0c) may be folded into a single precondition, keeping only the STRICT criterion on one line

### Block 5 run constraint

- **Single primary action** (typically a one-line command, e.g. `/moai run SPEC-ID`)
- Sub-detail (agent scope, AC bindings, file path line numbers) lives inside SPEC artifacts (plan.md / acceptance.md) — inline in the paste-ready is prohibited
- Ceremonial reminders ("exact reference", "observe discipline", "self-verify") are prohibited — those belong inside the agent body

### Block 6 follow-up constraint

- **≤ 2 lines** (next concrete SPEC ID or next phase command)
- Multi-step follow-ups (M4→M5→M6→sync→Mx→close) are managed via the SPEC plan.md milestones — inline in the paste-ready is prohibited

### Doctrine reference pattern

- N-th-iteration sustained 1st→2nd→3rd→4th→5th style history belongs ONLY in lesson memory files
- In the paste-ready, use a single one-line reference: `per session-handoff.md § <Doctrine Section>`

### Anti-pattern catalogue

> See also: § Anti-Patterns (general resume hygiene) and § V0 Abort Gate Doctrine / Anti-pattern (abort-gate violations AP-V-001..004). This catalogue covers paste-ready budget violations (AP-D-001..005).

- **AP-D-001**: Block 2 lessons 5+ references → trim to 4 or fewer, move the rest into the memory file body
- **AP-D-002**: precondition body prose (history/lesson narrative/cumulative pattern) → keep only a one-line verifiable command + STRICT criterion
- **AP-D-003**: Block 5 sub-step nesting (Phase 0 + Phase 0.5 + Phase 1B style multi-phase 11-substep) → compress into a single primary action; sub-detail belongs in SPEC artifacts
- **AP-D-004**: directive escalation embedded in body (N-th "stronger directive", N+1-th "even-stronger directive", N+2-th "documentation-level codification entry-condition") → codify in a rule file; the paste-ready keeps only the reference
- **AP-D-005**: ceremonial reminder ("B8/B15 observe discipline", "manager-develop must exactly reference plan.md §F.3 line 130-143") → keep inside SPEC artifacts; the paste-ready relies on trust delegation

### Pre-emit self-check (paste-ready budget) — 9 items

- [ ] Block 2 ≤ 4 references
- [ ] Block 2 each reference is a single-line identifier (full history prohibited)
- [ ] Block 4 each precondition ≤ 200 chars
- [ ] Block 4 precondition prose has no embedded history
- [ ] Block 5 single primary action (command + one-line context max)
- [ ] Block 6 ≤ 2 lines
- [ ] Doctrine history not embedded → rule-file reference only
- [ ] No ceremonial reminder
- [ ] Block 1 `/goal` re-set line (if emitted) is a single conditional line, not a multi-line block

### Applicable scope

- All new paste-ready resume messages
- Retry-iteration paste-ready messages (diet vs body-accumulation choice → diet is the default)
- Applied consistently across the line (all SPEC lines)

## V0 Abort Gate Doctrine

[ZONE:Evolvable] [HARD] The paste-ready Block 4 V0 precondition uses **lsof + cwd cross-validation**. A raw `ps aux` count is environmental baseline noise; used as the sole V0 check it produces false-positives where the STRICT ≤2 violation accumulates 13+ consecutive times in a multi-session environment (empirically proven).

### V0 verification commands (canonical)

```bash
# V0-a: informational baseline (NOT blocking — 16-19 sessions are normal in a healthy multi-session env)
ps aux | grep -iE '\bclaude\b' | grep -v -E 'plugin|Helper|Application|antigravity|grep' | wc -l

# V0-b: critical blocking — count of claude *processes* holding a file handle inside this WT
# Note: bare `grep -iE 'claude'` has a false-positive defect — it also matches content whose
#       filename contains 'claude' (claude-*.md etc.).
#       Always filter by the COMMAND column to keep only claude *processes* (`lsof -a -c claude`).
lsof -a -c claude +D "$PWD" 2>/dev/null | awk 'NR>1' | wc -l   # STRICT 0

# V0-c: critical blocking — count of active claude sessions whose cwd is this WT (this session + parent process only)
lsof -a -c claude -d cwd 2>/dev/null | awk 'NR>1 && $NF ~ /<this-WT-path>/' | wc -l   # STRICT ≤2
```

### Abort obligation

When V0-b ≥ 1 OR V0-c ≥ 3 (regardless of whether the other preconditions V1/V2/V3 PASS):
- Produce the next paste-ready iteration + write it to memory
- **Spawn prohibited** (manager-develop / manager-spec / manager-docs / any other implementation agents)
- **AskUserQuestion force-through options prohibited** (an override option violates the doctrine)
- End this session

### Cross-pollination history

Cross-line provenance: retained in lesson memory; this section codifies the doctrine. (The iteration history that originally surfaced the V0 false-abort hazard is preserved in lesson memory, not in this rule body — per AP-D-002, history belongs in lessons, not in paste-ready-adjacent prose.)

### Anti-pattern

> See also: § Anti-Patterns (general resume hygiene) and § Diet Constraints / Anti-pattern catalogue (paste-ready budget violations AP-D-001..005). This catalogue covers abort-gate violations (AP-V-001..004).

- **AP-V-001**: using `ps aux` raw count `≤ 2 STRICT` as the sole V0 check → environmental baseline noise (16-19 sessions are normal in a healthy multi-session state)
- **AP-V-002**: tracking "user promise accumulated non-fulfillment N times" in the body after a V0 FAIL → imposes only guilt, produces zero real behavior change, and bloats the paste-ready → instrumentalization anti-pattern
- **AP-V-003**: offering a force-through option (option D "override + spawn") in AskUserQuestion on a V0 FAIL → doctrine violation
- **AP-V-004**: measuring V0-b with `lsof +D "$PWD" | grep -iE 'claude'` → has a false-positive defect that also matches content whose filename contains 'claude' (claude-*.md etc.). The COMMAND-column process filter `lsof -a -c claude +D "$PWD"` is mandatory — only a genuine claude race signal may be counted so the abort obligation fires accurately

## Cross-references

<!-- self-check sentinel — references the render surface's structural invariant by content, not line number, so it survives line drift. This is mitigation + visibility (it surfaces drift to a reading editor), NOT mechanical prevention. A future editor who changes one surface without reading the other surface's sentinel produces silent drift; the only mechanical catch is a deferred Go lint rule (see the session-handoff SSOT-align doctrine §F.6 follow-up). -->
**Drift-mitigation self-check sentinel (SSOT → render surface).** This file is the SSOT; `.claude/output-styles/moai/moai.md §8` is the render surface. Before committing any edit to the Localization Table, the 6-block skeleton, the cut-line marker spec, or the Pre-emit self-check labels in THIS file, verify the parity check against the render surface: the moai.md §8 Localization Contract must carry the same locale column count (en / ko / ja / zh — 4 columns) as this file's Localization Table, and the moai.md §8 Pre-emit self-check labels must use the same concern-name qualifiers (`paste-ready budget` / `localization render` / `session-handoff template completeness`) as this file. If the two surfaces have diverged, this is the canonical surface — update the render surface to match.

- `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets — the per-model-class threshold SSOT for `/clear` and Trigger #1 (this file carries no inline model-class numbers to avoid label drift).
- `.claude/output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- `.claude/output-styles/moai/moai.md` §8 (Response Templates → Session Handoff) — the canonical render surface for the 6-block template + pre-emit self-check; this file is the SSOT, moai.md §8 is the render surface (bidirectional link).
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory + `[SUPERSEDED by ...]` convention
- CLAUDE.md §11 (Error Handling) — token-limit recovery
- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split rationale
- lessons #14 — `--worktree` Block 0 + single/multi-session decision
- lessons #12, #13 — worktree isolation + --team base mismatch

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
