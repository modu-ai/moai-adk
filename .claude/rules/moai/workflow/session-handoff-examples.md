---
description: "Illustrative examples and full 4-locale localization table for the session-handoff doctrine"
paths: "**/session-handoff.md"
---

# Session Handoff — Examples and Full Localization Table

> This is a path-scoped reference file for `session-handoff.md`. It holds illustrative Example sections and the full 4-locale Localization Table extracted from the always-loaded doctrine file to reduce context weight. The core doctrine (6-block skeleton, cut-line markers, Field-by-Field Spec, Pre-emit self-check, Auto-Memory Integration, Post-Paste /goal Follow-up Block, Diet Constraints) remains in `session-handoff.md`.

## Localization Table (Full 4-Locale)

The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table is the SSOT for the locale renderings (the canonical skeleton uses the `<entering verb>` / `<header>` placeholders; concrete locale renderings live here). Cross-verified for consistency with `.claude/output-styles/moai/moai.md §8` (the canonical render surface).

| Element | English | Korean | Japanese | Chinese |
|---------|---------|--------|----------|---------|
| Cut-line top text | `Copy from here` | `여기부터 복사` | `ここからコピー` | `从这里复制` |
| Cut-line bottom text | `Copy to here` | `여기까지 복사` | `ここまでコピー` | `到这里复制` |
| Block 1 entering verb | `entering` | `진입` | `開始` | `进入` |
| Block 3 Preconditions header | `Preconditions:` | `전제 검증:` | `前提条件:` | `前提条件:` |
| Block 5 Run header | `Run:` | `실행:` | `実行:` | `执行:` |
| Block 6 After-merge header (PR workflow) | `After merge:` | `머지 후:` | `マージ後:` | `合并后:` |
| Block 6 Follow-up header (trunk no-PR) | `Follow-up:` | `후속:` | `後続:` | `后续:` |
| Memory heading | `## Next Session Entry Point` | `## 다음 세션 시작점` | `## 次セッション開始点` | `## 下一会话起点` |
| Post-paste /goal instruction line | Send the `/goal` line below as its own standalone message AFTER Implementation Kickoff Approval — slash commands parse only at input start, and setting a goal starts a turn immediately. | 아래 `/goal` 라인을 구현 착수 승인 후 **별도 메시지로 단독 전송** — 슬래시 커맨드는 입력 시작에서만 인식되며, goal 설정 즉시 턴이 시작됨. | 下記の `/goal` 行を実装着手承認後に**単独メッセージとして送信** — スラッシュコマンドは入力の先頭でのみ認識され、goal 設定と同時にターンが開始される。 | 在实现启动批准后，将下方 `/goal` 行**作为独立消息单独发送** — 斜杠命令仅在输入开头被识别，设定 goal 会立即开始一个回合。 |

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers) while keeping `✂` and `─` characters verbatim, and substitute the locale rendering for each Block 1/3/5/6 placeholder when emitting the paste-ready message.

**Fallback rule for locales not in the table.** The table above lists concrete renderings for en / ko / ja / zh only. When `conversation_language` is an ISO-639 code whose language column is NOT in this table (e.g. `fr`, `de`, `es`, `pt`, `vi`), English is the canonical fallback skeleton and each label translates to that locale using the naturalization principle (idiomatic phrasing a native reader expects, never literal word-by-word transliteration). In other words: locales not in the table fall back to the English column for the structural skeleton, with the label text rendered in the configured ISO-639 language — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

## Example (Illustrative; substitute project-specific values when adapting)

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-MYPROJ-001 implementation 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

전제 검증:
1) git log --oneline -1 → <commit-sha> 확인
2) ls .moai/specs/SPEC-MYPROJ-001/ → N files

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002 → SPEC-MYPROJ-003

✂──── 여기까지 복사 ────✂

아래 /goal 라인을 구현 착수 승인 후 별도 메시지로 단독 전송 — 슬래시 커맨드는 입력 시작에서만 인식됨 (run-phase + machine-verifiable end-state일 때만 방출; 아니면 생략):

✂──── 여기부터 복사 ────✂

/goal the SPEC's test suite passes AND lint is clean, or stop after 20 turns

✂──── 여기까지 복사 ────✂
```

## Example with Block 0 (Illustrative)

```
✂──── 여기부터 복사 ────✂

[New Terminal — START IN WORKTREE]
$ cd ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
$ moai cc        # 또는 moai glm | claude (3가지 launcher 중 선택; 본 예시는 moai cc)

ultrathink. SPEC-MYPROJ-001 Epic N 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

후속: Milestone M<N+1> (single-SPEC next step) 또는 Epic N+1 (multi-SPEC next grouping)

✂──── 여기까지 복사 ────✂
```

---

# Moved Sections (extracted from session-handoff.md for context diet)

## Goal-first bootstrap variant (documented alternative — NOT the default)

[ZONE:Evolvable] An explicit alternative single-paste form exists: the **goal-first bootstrap** — a standalone one-line `/goal` message whose condition text carries both a resume pointer and the compact completion condition. Illustrative:

```text
/goal resume SPEC-X run: read <handoff-file> from memory and progress.md, then continue. Completion: <machine-verifiable end-state>, or stop after N turns.
```

(The condition text follows the user's `conversation_language`; shown above in English-canonical form. The `/goal` token itself is locale-verbatim.)

Grounding: official goal doc — "Setting a goal starts a turn immediately, with the condition itself as the directive" (`https://code.claude.com/docs/en/goal`). Normative content:

- **(a) Selection criterion**: choose goal-first bootstrap when the user wants one-paste + autonomous continuation; the two-step handoff (§ Block anatomy) remains the DEFAULT.
- **(b) Caveats**: effort keywords (`ultrathink` / `ultracode`) placed inside a slash-command argument are NOT documented to fire — the session may run at default effort; and precondition verification shifts from paste-time structure (the Block 4 verifiable commands) to **model discretion** via the directive text.
- **(c) Invariants preserved**: the condition must stay compact (official guidance: one measurable end state); the Implementation Kickoff Approval gate is unaffected; the `/goal` token stays locale-verbatim (never translated).

## Paste-Time Activation Matrix

[ZONE:Evolvable] The following normative table classifies every handoff directive by its activation mechanism, so an author never places a directive where it cannot fire. Ground truth: `https://code.claude.com/docs/en/interactive-mode` (slash commands recognized only at input start) and `https://code.claude.com/docs/en/goal` (`/goal` is a user-typed TUI command, not model-invocable).

| Class | Directives | Mechanism | Fires from pasted body? |
|-------|-----------|-----------|------------------------|
| (a) Paste-time keyword | `ultrathink`, bare `ultracode` | Runtime keyword, position-independent in message text | YES |
| (b) Paste-time natural-language phrase | `fan out subagents (<scope>)` | Explicit multi-agent opt-in phrase — same opt-in class as (a) | YES |
| (c) Orchestrator-interpreted text | `mode:` seed, Block 5 `실행: /moai <subcommand>` | The orchestrator reads the text and routes (`/moai` via the Skill tool); NOT auto-executed as a slash command | YES (via orchestrator interpretation) |
| (d) User-only TUI command | `/goal`, `/effort`, `/clear` | Slash command parsed ONLY at input start; not model-invocable; cannot be set by pasted body text NOR by the model | NO — requires a standalone user message |

Consequence: a `/goal` line belongs to class (d) — it MUST arrive as its own standalone user message (§ Post-Paste /goal Follow-up Block), never inside the pasted resume body where it would be inert class-(d) plain text.

The same classification governs the auto-injected body (§ Auto-Injected Resume Flow): content delivered as session-start context injection is inert context — it cannot fire class (a)/(b) paste-time keywords on its own and cannot execute class (d) commands. That is why the auto flow's ONE user message carries the class (d) `/goal` line (goal-first variant) or the class (a) `ultrathink` keyword (approval variant) in the user's own message.

## Auto-Injected Resume Flow (mode=auto)

### One-message flow

1. The previous session emits the paste-ready resume AND persists it via `moai handoff save` (§ Emission-Time Save Obligation). The paste-ready surface is still displayed — the user can always fall back to the manual paste path.
2. The user runs `/clear`.
3. The session-start handler, in the single consume cell (session source is `clear` AND `handoff.mode: auto` AND a live pending record exists), **claim-renames** the pending record into a `consumed/` audit-trail copy FIRST, then injects the saved content as session-start additional context. The claim-then-inject atomic rename means exactly one of two racing sessions injects — the loser's rename fails and it skips injection fail-open. A record older than the stale TTL is cleaned up instead of injected.
4. What the injection actually contains: a localized header; a disclaimer stating the injection only delivers context and does NOT automatically enable any extended-reasoning mode; restoration-guidance lines for the recorded directives (`ultrathink` / `/effort ultracode` / `/goal <condition>` — each rendered as manual-input guidance the user may type, never as an executed command); and the saved body **verbatim** (no re-localization — `--lang` snapshots the language at save time). The injected context cannot start a turn and cannot claim effort restoration; the platform caps session-start injected context at 10,000 characters, and the § Diet Constraints budget keeps the 6-block body far below that cap.
5. The user sends **ONE** message:
   - **Goal-first variant** — Where the next SPEC is run-phase AND declares a machine-verifiable end-state, the one message is the single standalone `/goal <condition>` line (slash commands parse only at input start, so a standalone message satisfies the class (d) activation constraint in § Paste-Time Activation Matrix).
   - **Approval variant** — otherwise, the one message is a short approval/continue message. Keep recommending that the user include the `ultrathink` keyword in this first message: the injected context cannot restore effort, but a paste-time keyword in the user's own message can.
   - **Effort caveat (goal-first)**: effort keywords placed inside a slash-command argument are NOT documented to fire — a `/goal ... ultrathink ...` line may leave the session at default effort. The doctrine does not claim the goal-first variant restores extended reasoning.

### /clear-only injection boundary

Injection happens ONLY when the session-start source is `clear`. All other session-start sources — `startup`, `resume`, `compact` — are **notice-only**: the pending record is never consumed there, and with the `handoff.guide` key at its default `false` the notice is silent (no visible output; when `guide: true`, a best-effort stderr hint mentions the waiting record). Consequences:

- A terminal restart (new session process, source `startup`) does NOT auto-inject — the manual paste path applies.
- An L3 worktree Block 0 resume (new terminal inside the worktree, source `startup`) falls OUTSIDE auto-inject — Block 0 + the manual paste path remain the mechanism (§ Worktree-Anchored Resume Pattern).
- Only the in-place `/clear` boundary gets the one-message flow.

### Precondition verification at resumed-turn start

The injected Block 4 preconditions MUST be verified at the start of the resumed session's first working turn — injection delivers the TEXT of the preconditions, not their truth. This is most acute in the goal-first variant, where `/goal` starts a turn immediately: the orchestrator verifies the injected preconditions FIRST, before acting on the goal condition.

## Anti-Patterns

> See also: § Diet Constraints / Anti-pattern catalogue (paste-ready budget violations AP-D-001..005) and § V0 Abort Gate Doctrine / Anti-pattern (abort-gate violations AP-V-001..004). This list covers general resume-hygiene patterns; the Diet and V0 lists cover their respective specialized domains.

See the general-hygiene bullet list and the §Diet Constraints and §V0 Abort Gate Doctrine anti-pattern catalogues below for the full catalogue.

- Free-form prose handoff — no executable context.
- Resume without preconditions — next session cannot detect state drift.
- Resume without `ultrathink.` — fails to activate xhigh effort.
- Resume saved only to chat, not auto-memory — lost across `/clear`.
- Duplicate memory entries without `[SUPERSEDED by ...]` markers — index pollution.
- Resume Block 2 missing `source_session_id: <UUID from moai session current>` **AND missing the environment fallback pattern** (`<not-available — environment-fallback, ...>`) — the canonical multi-session coordination policy cannot correlate the resume back to its originating session for race attribution. The environment fallback pattern itself is NOT an anti-pattern; only the complete absence of both UUID and fallback pattern is the violation.
- Forcing the format on trivial tasks — memory noise.
- Cut-line markers absent — user cannot identify exact copy boundary in long terminal scrollback (see § Cut-line Marker Specification for the literal format).
- Cut-line markers with translated `✂` symbol or `─` decorator — contrary to § Cut-line Marker Specification (only the marker text translates; the symbols are preserved verbatim).
- Omitting the bare `ultracode` opener keyword (or the `/effort ultracode` session-persistence variant) when the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams) — the resumed session silently drops to non-ultracode effort and loses auto-orchestration (ultracode is NOT restored by `ultrathink.` per `.claude/rules/moai/workflow/dynamic-workflows.md`).
- Omitting the post-paste `/goal` follow-up block when the next SPEC has a verifiable run-phase completion condition — the resumed session silently loses the autonomous-continuation loop (a `/goal` is NOT restored by `ultrathink.`; `/clear` removes an active goal, per `.claude/rules/moai/workflow/goal-directive.md`; the follow-up block + resumed-session reminder obligation is the two-step delivery mechanism, per § Post-Paste /goal Follow-up Block).
- Embedding a `/goal` (or any slash command) line inside the main resume body — slash commands parse only at input start of a standalone message; a mid-paste slash line is inert plain text and never arms the goal loop (see § Paste-Time Activation Matrix). Deliver `/goal` via the separate post-paste follow-up block sent as its own standalone message.
- Omitting the fan-out steering phrase (`fan out subagents (<read-only investigation scope>)`) when `mode: parallel-subagents` is seeded — the resumed session silently under-spawns: fewer subagents are spawned by default unless fan-out is explicitly instructed (per `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy Principle 4; the fan-out steering phrase is NOT restored by the `ultrathink.` opener).

## Worktree-Anchored Resume Pattern

[ZONE:Evolvable] [HARD] When the SPEC was initialized via L3 `/moai plan --worktree` (creating an L2 SPEC worktree at `~/.moai/worktrees/<project>/<spec-or-name>/`), the resume message MUST include **Block 0 (cwd anchoring)** prepended before the standard 6-block structure. Without Block 0, the next session starts in main project cwd by default, breaking L2 SPEC worktree isolation expectations.

> L3 `--worktree` is **user opt-in** only. For SPECs initialized without `--worktree` (the default), the standard 6-block structure suffices — Block 0 is NOT required.

### Why Block 0 (L3 `--worktree` opt-in only)

With L3 `--worktree`, SPEC artifacts and L1 isolation base live in a different cwd. Pasting resume into a main-cwd session causes: L1 base divergence per the worktree isolation guidance, Bash commands targeting main project per the worktree isolation guidance, build/test from the wrong tree. Block 0 forces a new terminal session **inside** the L2 worktree before any action.

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

Block 0 is REQUIRED only with L3 `--worktree`. For `--branch` (or no flag — the opt-in default), standard 6-block suffices because main session cwd already follows the branch.

[ZONE:Evolvable] [HARD] If L3 `--worktree` was used and the user is NOT comfortable with multi-terminal/multi-session workflow, the orchestrator SHOULD recommend `--branch` for the next SPEC. Forcing Block 0 onto a single-session user is friction without benefit. See the single-session vs multi-session decision rationale below.

> **Example with Block 0**: see `session-handoff-examples.md` § Example with Block 0 (Illustrative).

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
