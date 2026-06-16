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

- **Block 1**: `ultrathink.` triggers Adaptive Thinking xhigh effort on Opus 4.7+ (next session lacks accumulated reasoning). `<phase>` ∈ `plan | run | sync | mx`.
- **Block 2**: `applied lessons:` — relevant memory files from `~/.claude/projects/{hash}/memory/`. MUST include the most recent relevant project memory + any relevant lessons. Block 2 MUST also include a `source_session_id: <UUID from moai session current>` line carrying the Claude Code session_id of the orchestrator turn that generated this resume message per the canonical multi-session coordination policy. The session_id is the same value emitted by `moai session list --json` and stored in `.moai/state/active-sessions.json` — readers can correlate the resume back to its originating session.
  - **Environment fallback** [HARD]: the primary UUID source is `moai session current`. If `moai session current` returns the canonical fallback (runtime did not expose session.id to the CLI subprocess), OR `moai session list --json` returns error (CLI not installed in PATH), OR `.moai/state/active-sessions.json` does not exist (the multi-session coordination layer not yet deployed in this project), the orchestrator MUST emit the recognized fallback pattern verbatim: `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`. This pattern is NOT an anti-pattern; it is the prescribed graceful degradation when the CLI/registry layer is absent or the runtime does not expose session.id. The next session, upon `/moai session register` activation, MAY backfill the UUID by appending a `[backfilled: <UUID>]` annotation to the memory file's Block 2 line.
- **Block 3**: separator + `전제 검증:` (Korean) or `Preconditions:` (English).
- **Block 4**: numbered preconditions `<N>) <action> → <expected outcome>`. Each MUST be independently verifiable (git/gh command, file existence). Max 4 preconditions.
- **Block 5**: separator + `실행: <command-or-action>` — single primary action (typically `/moai <subcommand>`).
- **Block 6**: separator + `<workflow-context header>: <next-action-or-spec>` — RECOMMENDED for multi-SPEC waves or follow-up; **omit entirely** for single-SPEC close with no further actions queued.
  - **Header selection (workflow-context conditional)**:
    - **PR-based workflow** (feat/* → PR → merge): `머지 후:` (en `After merge:`)
    - **Trunk-based no-PR** (e.g., 1-person OSS, all-tier main 직진 push, no merge step): `후속:` (en `Follow-up:`)
    - **Single-SPEC close** (no further SPEC/phase queued): omit Block 6 entirely
  - **Single action principle**: `<next-action-or-spec>` MUST be one concrete SPEC ID, one command, or one phase transition — avoid vague "사이클 반복" / "iteration loop" phrasing that reads as infinite recursion.

### Example (Illustrative; substitute project-specific values when adapting)

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-MYPROJ-001 implementation 진입.
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
| AP-D-005 | ceremonial reminder ("discipline 엄수", "정확 참조") in paste-ready | § Diet Constraints / Anti-pattern catalogue |
| AP-V-001 | `ps aux` raw count `≤ 2 STRICT` as sole V0 verification | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-002 | V0 FAIL 후 "사용자 약속 누적 미이행 N회" body tracking | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-003 | V0 FAIL 시 AskUserQuestion 강행 옵션 (override + spawn) 제시 | § V0 Abort Gate Doctrine / Anti-pattern |
| AP-V-004 | V0-b 측정에 `lsof +D "$PWD" | grep -iE 'claude'` 사용 (false-positive 결함) | § V0 Abort Gate Doctrine / Anti-pattern |

- Free-form prose handoff — no executable context.
- Resume without preconditions — next session cannot detect state drift.
- Resume without `ultrathink.` — fails to activate xhigh effort.
- Resume saved only to chat, not auto-memory — lost across `/clear`.
- Duplicate memory entries without `[SUPERSEDED by ...]` markers — index pollution.
- Resume Block 2 missing `source_session_id: <UUID from moai session current>` **AND missing the environment fallback pattern** (`<not-available — environment-fallback, ...>`) — the canonical multi-session coordination policy cannot correlate the resume back to its originating session for race attribution. The environment fallback pattern itself is NOT an anti-pattern; only the complete absence of both UUID and fallback pattern is the violation.
- Forcing the format on trivial tasks — memory noise.
- Cut-line markers absent — user cannot identify exact copy boundary in long terminal scrollback (see § Cut-line Marker Specification for the literal format).
- Cut-line markers with translated `✂` symbol or `─` decorator — contrary to § Cut-line Marker Specification (only the marker text translates; the symbols are preserved verbatim).

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

ultrathink. SPEC-MYPROJ-001 Wave N 진입.
applied lessons: project_myproj_prev_sprint_complete, lessons #12 #13 #14.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

후속: Round N+1 (single-SPEC SSE split context) 또는 Sprint N+1 (multi-SPEC cohort context)

✂──── 여기까지 복사 ────✂
```

## Diet Constraints

[ZONE:Evolvable] [HARD] paste-ready resume message는 "next session minimum executable context"이다 — audit trail, history record, ceremonial commitment record가 아니다. 차수 누적 retry 진행 시 본문에 history/lesson/directive escalation prose를 append-only로 누적하는 것은 empirical 입증된 anti-pattern이다.

### Block 2 applied lessons 제약

- 최대 **4개 references** (memory file slug 또는 lesson identifier)
- 각 reference는 **1줄 identifier** (예: `L52#33`, `L_NEW_V0_ABORT_GATE` — full prose history 금지)
- 5개 이상은 anti-pattern → memory file body로 이관

### Block 4 precondition 제약

- 각 precondition **≤ 200 chars** target (실용적 가독성 한계)
- Format: `N) <verifiable command> → <expected outcome>`
- History tracking / lesson narrative / 누적 패턴 추적 prose 금지
- Multi sub-command (V0a/V0b/V0c)는 단일 precondition으로 통합 가능, STRICT criterion만 1줄로

### Block 5 실행 제약

- **단일 primary action** (typically 1줄 command, 예: `/moai run SPEC-ID`)
- Sub-detail (agent scope, AC bindings, file path line numbers)은 SPEC artifacts(plan.md / acceptance.md) 내부에 존재 — paste-ready inline 금지
- Ceremonial reminder ("정확 참조", "discipline 엄수", "self-verify") 금지 — 이는 agent body 내부 책임

### Block 6 후속 제약

- **≤ 2줄** (next concrete SPEC ID 또는 next phase command)
- Multi-step 후속 (M4→M5→M6→sync→Mx→close)는 SPEC plan.md milestone로 관리 — paste-ready inline 금지

### Doctrine reference 패턴

- N차 sustained 1st→2nd→3rd→4th→5th 같은 history는 lesson memory file에만 보관
- paste-ready에서는 `per session-handoff.md § <Doctrine Section>` 1줄 reference만 사용

### Anti-pattern catalogue

> See also: § Anti-Patterns (general resume hygiene) and § V0 Abort Gate Doctrine / Anti-pattern (abort-gate violations AP-V-001..004). This catalogue covers paste-ready budget violations (AP-D-001..005).

- **AP-D-001**: Block 2 lessons 5+ references → 4 이하로 trim, 나머지는 memory file body로 이관
- **AP-D-002**: precondition 본문 prose (history/lesson narrative/누적 패턴) → 1줄 verifiable command + STRICT criterion만 남기기
- **AP-D-003**: Block 5 sub-step nesting (Phase 0 + Phase 0.5 + Phase 1B 같은 multi-phase 11-substep) → single primary action으로 압축, sub-detail은 SPEC artifacts에
- **AP-D-004**: directive escalation 본문 임베드 (N차 "stronger directive", N+1차 "even-stronger directive", N+2차 "documentation-level codification entry-condition") → rule file로 codification, paste-ready는 reference만
- **AP-D-005**: ceremonial reminder ("B8/B15 discipline 엄수", "manager-develop은 plan.md §F.3 line 130-143 정확 참조") → SPEC artifact 내부 보관, paste-ready는 trust delegation

### Pre-emit self-check (paste-ready budget) — 8 items

- [ ] Block 2 ≤ 4 references
- [ ] Block 2 각 reference 1줄 identifier (full history 금지)
- [ ] Block 4 각 precondition ≤ 200 chars
- [ ] Block 4 precondition prose에 history 임베드 없음
- [ ] Block 5 single primary action (command + 1줄 context max)
- [ ] Block 6 ≤ 2 lines
- [ ] Doctrine history not embedded → rule file reference only
- [ ] Ceremonial reminder 없음

### 적용 범위

- 모든 신규 paste-ready resume message
- 차수 누적 retry paste-ready (다이어트 vs 본문 누적 선택 → 다이어트 default)
- Cross-line 일관 적용 (모든 SPEC line)

## V0 Abort Gate Doctrine

[ZONE:Evolvable] [HARD] paste-ready Block 4 V0 precondition은 **lsof + cwd 교차 검증**을 사용한다. `ps aux` raw count는 environmental baseline noise이며 단독 V0 검증으로 사용 시 multi-session 환경에서 STRICT ≤2 위반이 13회+ 연속 누적되는 false-positive를 발생시킨다 (empirical 입증).

### V0 검증 명령 (canonical)

```bash
# V0-a: informational baseline (blocking 아님 — multi-session 정상 환경에서 16-19 expected)
ps aux | grep -iE '\bclaude\b' | grep -v -E 'plugin|Helper|Application|antigravity|grep' | wc -l

# V0-b: critical blocking — 본 WT 내부에 file handle 보유한 claude *프로세스* 수
# 주의: `grep -iE 'claude'` 단독은 파일명에 'claude' 포함된 콘텐츠(claude-*.md 등)까지 매칭하는
#       false-positive 결함이 있다.
#       반드시 COMMAND 컬럼으로 claude *프로세스*만 필터한다 (`lsof -a -c claude`).
lsof -a -c claude +D "$PWD" 2>/dev/null | awk 'NR>1' | wc -l   # STRICT 0

# V0-c: critical blocking — cwd가 본 WT인 active claude session 수 (본 세션 + parent process만)
lsof -a -c claude -d cwd 2>/dev/null | awk 'NR>1 && $NF ~ /<본_WT_경로>/' | wc -l   # STRICT ≤2
```

### Abort 의무

V0-b ≥ 1 OR V0-c ≥ 3 시 (다른 precondition V1/V2/V3 PASS 여부 무관):
- 다음 paste-ready 차수 산출 + memory write
- **Spawn 금지** (manager-develop / manager-spec / manager-docs / 기타 implementation agents)
- **AskUserQuestion 강행 옵션 제시 금지** (override option은 doctrine 위반)
- 본 세션 종료

### Cross-pollination 이력

Cross-line provenance: retained in lesson memory; this section codifies the doctrine. (The iteration history that originally surfaced the V0 false-abort hazard is preserved in lesson memory, not in this rule body — per AP-D-002, history belongs in lessons, not in paste-ready-adjacent prose.)

### Anti-pattern

> See also: § Anti-Patterns (general resume hygiene) and § Diet Constraints / Anti-pattern catalogue (paste-ready budget violations AP-D-001..005). This catalogue covers abort-gate violations (AP-V-001..004).

- **AP-V-001**: `ps aux` raw count `≤ 2 STRICT`을 단독 V0 검증으로 사용 → environmental baseline noise (multi-session normal state에서 16-19 sessions은 정상)
- **AP-V-002**: V0 FAIL 후 "사용자 약속 누적 미이행 N회" 본문 추적 → 죄책감 부담만 부과 + 실질 행동 변화 0 + paste-ready 비대화 → 도구화 anti-pattern
- **AP-V-003**: V0 FAIL 시 AskUserQuestion에 강행 옵션 (option D "override + spawn") 제시 → doctrine 위반
- **AP-V-004**: V0-b 측정에 `lsof +D "$PWD" | grep -iE 'claude'` 사용 → 파일명에 'claude' 포함된 콘텐츠(claude-*.md 등)까지 매칭하는 false-positive 결함이 있다. COMMAND 컬럼 프로세스 필터 `lsof -a -c claude +D "$PWD"` 필수 — genuine claude race signal만 카운트해야 abort 의무가 정확히 발동한다

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
