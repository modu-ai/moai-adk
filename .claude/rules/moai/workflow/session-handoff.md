# Session Handoff Protocol

Long-running session continuity: clean transitions across context boundaries via paste-ready resume messages.

> **Loading scope**: Intentionally always-loaded (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context, including those without SPEC files. The always-loaded cost is justified by cross-cutting applicability.

## Why This Matters

Long workflows (multi-SPEC Epics, multi-milestone implementation) accumulate context that exceeds the window or benefits from fresh start. Without a standardized handoff, session boundaries lose work-in-progress. This rule defines when to emit a paste-ready resume, the 6-block structure, and auto-memory integration that persists across `/clear`.

## When To Generate (5 Triggers)

[ZONE:Evolvable] [HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:

| # | Trigger | Detection |
|---|---------|-----------|
| 1 | Context usage crosses model-specific threshold (cumulative input+output) | Model-specific percentage threshold (1M-context models vs 200K-context models) — see `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets for the per-model-class threshold table (the authoritative SSOT for the numeric thresholds; this file carries no inline model-class numbers to avoid label drift). |
| 2 | SPEC phase completion (plan/run/sync) within a multi-SPEC workflow | Phase boundary in `.claude/rules/moai/workflow/spec-workflow.md` §Phase Transitions (after plan/run/sync phase finishes within a multi-SPEC SPEC ID series) |
| 3 | User explicitly requests session end ("세션 종료", "이번 세션 마무리", "next session") | Intent detection in user message |
| 4 | PR creation success when more SPECs remain in the current Epic | After `gh pr create` success + memory indicates >0 pending SPECs |
| 5 | Long-running multi-milestone task reaches a stable checkpoint | After milestone Mn complete + Mn+1 not yet started |

When NONE apply (single-turn, trivial task, read-only query), emit a brief completion confirmation. The threshold in Trigger #1 reflects asymmetric stall risk: 1M models tolerate higher absolute load; 200K models hit the ceiling earlier. The `/clear` policy in `context-window-management.md` is co-anchored to the same threshold per model class.

### Emission-Time Save Obligation (auto-resume wiring)

[ZONE:Evolvable] [HARD] When the orchestrator emits a paste-ready resume message (any of the 5 triggers above), it MUST also persist the cut-line-bounded main block verbatim as the pending handoff record: pipe the block to `moai handoff save --stdin --spec <ID> --phase <phase> [--goal "<condition>"] [--ultrathink] [--ultracode] [--lang <conversation_language>] [--session <uuid>]` (body fed via stdin). `--goal` is recorded ONLY when the `/goal` emission condition holds (next SPEC is run-phase AND declares a machine-verifiable end-state — the same condition as § Post-Paste /goal Follow-up Block); `--lang` snapshots the current `conversation_language`; `--session` carries the same session id as Block 2's `source_session_id` when available.

[ZONE:Evolvable] [HARD] **Fail-open invariant**: when the `moai` CLI is absent from PATH or `moai handoff save` exits non-zero, the orchestrator emits the paste-ready surface UNCHANGED — a save failure never blocks, delays, or alters handoff emission, and no retry loop is entered. The manual paste path is fully functional without the save; the save is an additive persistence step, never a gate.

The saved record (`.moai/state/handoff/pending.json`) is consumed by the auto-injected flow when the project config sets `handoff.mode: auto` — see § Auto-Injected Resume Flow. Under the distributed default `handoff.mode: manual` the record is inert (the session-start injector never touches it, even when stale), and this save obligation still applies — flipping the mode later requires no doctrine change.

## Canonical Format (Verbatim Spec)

[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers** (see § Cut-line Marker Specification below for the literal marker format, Unicode-preservation rules, and locale translation contract). Cut-line markers sit **inside** the fenced text block alongside the content so they are copied verbatim with the message; this provides the user an unambiguous copy boundary in long terminal scrollback:

```
✂──── 여기부터 복사 ────✂

ultrathink. <SPEC-ID> <phase> <entering verb>.
mode: <value>   ← emit ONLY when the seeded orchestration mode ≠ solo-sequential; value ∈ {parallel-subagents | agent-team | dynamic-workflow} → Phase 0.95 Mode 4 / 3 / 6. OMIT for solo-sequential (default) → v1 byte-identical. When mode = dynamic-workflow, ALSO append bare `ultracode` to the opener line above (paste-time trigger keyword; the session-persistent `/effort ultracode` slash form is a separate variant — per Field-by-Field Spec, Block 1). When mode = agent-team, append `--team` to the Block 5 run command. When mode = parallel-subagents, append the fan-out steering phrase `fan out subagents (<read-only investigation scope>)` to the opener line above (paste-time steering phrase — per Field-by-Field Spec, Block 1).
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
| Post-paste /goal instruction line | Send the `/goal` line below as its own standalone message AFTER Implementation Kickoff Approval — slash commands parse only at input start, and setting a goal starts a turn immediately. | 아래 `/goal` 라인을 구현 착수 승인 후 **별도 메시지로 단독 전송** — 슬래시 커맨드는 입력 시작에서만 인식되며, goal 설정 즉시 턴이 시작됨. | 下記の `/goal` 行を実装着手承認後に**単独メッセージとして送信** — スラッシュコマンドは入力の先頭でのみ認識され、goal 設定と同時にターンが開始される。 | 在实现启动批准后，将下方 `/goal` 行**作为独立消息单独发送** — 斜杠命令仅在输入开头被识别，设定 goal 会立即开始一个回合。 |

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers) while keeping `✂` and `─` characters verbatim, and substitute the locale rendering for each Block 1/3/5/6 placeholder when emitting the paste-ready message.

**Fallback rule for locales not in the table.** The table above lists concrete renderings for en / ko / ja / zh only. When `conversation_language` is an ISO-639 code whose language column is NOT in this table (e.g. `fr`, `de`, `es`, `pt`, `vi`), English is the canonical fallback skeleton and each label translates to that locale using the naturalization principle (idiomatic phrasing a native reader expects, never literal word-by-word transliteration). In other words: locales not in the table fall back to the English column for the structural skeleton, with the label text rendered in the configured ISO-639 language — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

### Field-by-Field Specification

- **Block 1**: `ultrathink.` sets `effort: xhigh` on Opus 4.7+ (next session lacks accumulated reasoning). Adaptive Thinking is a DISTINCT axis — the thinking mode, explicitly enabled via `thinking: {type: "adaptive"}` — not something `ultrathink` toggles. `<phase>` ∈ `plan | run | sync | mx`.
  - **Block 1 line-order invariant** [HARD]: the Block 1 lines emit in this fixed order — `ultrathink.` opener (with an optional appended bare `ultracode` keyword or fan-out steering phrase) → `mode:` line (when present) → `applied lessons:` → `source_session_id:`. Each conditional line is omitted when its condition does not hold; in the common solo-sequential case, only the opener + `applied lessons:` remain, byte-identical to v1. The main resume block carries NO `/goal` line — a mid-paste `#`-prefixed (or bare) slash line is inert plain text because official slash-command parsing recognizes a `/` command only at the input start of a standalone message; the `/goal` autonomous-continuation directive is delivered by the separate post-paste mechanism (see § Post-Paste /goal Follow-up Block), never inside this block.
  - **Purpose-conditional `mode:` orchestration-seed line** [HARD]: Block 1 carries a purpose-conditional `mode: <value>` line that **seeds** the next session's Phase 0.95 orchestration mode. It is emitted ONLY when the seeded mode is NOT `solo-sequential`; for `solo-sequential` (the default) the line is **omitted**, keeping the message byte-identical to v1 (zero-diff common case). The `mode:` line sits directly below the `ultrathink.` opener. Its value is a protocol token drawn from a fixed 4-enum that maps 1:1 onto the Phase 0.95 mode catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A):

    | `mode:` value (seed) | Phase 0.95 catalog | Emission | Directive coupling |
    |----------------------|--------------------|----------|--------------------|
    | `solo-sequential` | Mode 5 (sub-agent, default fallback) | **omitted** (default) | omission = v1 byte-identical |
    | `parallel-subagents` | Mode 4 (parallel, 3-5 concurrent `Agent()`) | emitted | append `fan out subagents (<read-only investigation scope>)` to the opener line |
    | `agent-team` | Mode 3 (agent-team, implicit team) | emitted | append `--team` to the Block 5 run command |
    | `dynamic-workflow` | Mode 6 (workflow, orchestrator fan-out) | emitted | append bare `ultracode` to the opener line |

    - **Excluded modes**: Mode 1 (trivial) and Mode 2 (background) are NOT handoff-relevant seeds — a handoff never resumes into a trivial or background mode as its primary re-entry mode, so neither is assigned a `mode:` token.
    - **Threshold reuse (no new threshold)**: the seed derives from Phase 0.95's existing auto-select thresholds (domains ≥ 3 / files ≥ 10 / score ≥ 7, per `orchestration-mode-selection.md` §B.1). The `mode:` seed introduces NO new threshold.
    - **SEED, not a permission grant** [HARD]: the `mode:` value is a SEED (a signal for the next session's orchestrator), NOT a permission grant. The Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory regardless of the seeded mode — a seeded `dynamic-workflow` or `agent-team` does NOT authorize autonomous run-phase entry. The seed only pre-selects the orchestration shape the user is subsequently asked to approve.
    - **Directive binding**: `ultrathink.` is emitted always (v1 invariant); bare `ultracode` is appended to the opener line ONLY when mode = `dynamic-workflow`; the post-paste `/goal` follow-up block is emitted only for a run-phase next SPEC with a machine-verifiable end-state (unchanged condition, new placement — see § Post-Paste /goal Follow-up Block); `--team` is appended to the Block 5 run command only when mode = `agent-team`; the fan-out steering phrase `fan out subagents (<read-only investigation scope>)` is appended to the opener line ONLY when mode = `parallel-subagents` (see the fan-out steering phrase bullet below).
    - **solo-sequential emission policy (emit-discouraged + parse-accept)**: `solo-sequential` is the emit-discouraged default — its `mode:` line is not emitted (Block 1 omits it → v1 byte-identical). An explicit `mode: solo-sequential` line, should a producer choose to write one, is parse-accepted (forward-compatible) and read as Mode 5, merely redundant with the omitted default. The framing is single: prefer omission, accept an explicit value — the doctrine does not simultaneously discourage emission and forbid parsing.
    - **`mode:` is a locale-verbatim protocol token**: like the `plan | run | sync | mx` phase tokens, the `mode:` value is preserved verbatim across all locales and is NOT added as a row to any localization / cut-line / header translation table.
    - **JSON-twin forward-compat note**: there is no JSON twin currently (this doctrine is doctrine-only, no code). Where a JSON-twin representation of the resume message is later introduced, that twin shall set `schema_version: 2` and carry the `mode` field. This note records forward-compatibility only and triggers no code change now.
  - **Purpose-conditional fan-out steering phrase (mode = parallel-subagents)** [HARD]: when the seeded mode is `parallel-subagents`, the resume message appends the natural-language fan-out steering phrase — canonical form `fan out subagents (<read-only investigation scope>)` — after the opener text on the Block 1 opener line. The paste-ready message is **user-pasted**, so the phrase counts at the runtime layer as a user-authored explicit multi-agent opt-in — the same paste-time class as the `ultrathink` / bare `ultracode` keywords. Rationale: newer models spawn fewer subagents by default, and fan-out must be instructed explicitly (per `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy Principle 4); a resumed session parsing only the `mode:` metadata line silently under-spawns without this phrase. **Locale-verbatim phrase**: `fan out subagents` is a locale-verbatim protocol phrase — preserved in English across all locales, exactly like the `mode:` values, and NOT added as a row to any localization / cut-line / header translation table; only the parenthesized scope qualifier translates per `conversation_language` (e.g. ko: `fan out subagents (read-only 코드베이스 조사)`). **Invariants**: (a) SEED-not-permission — the phrase does NOT authorize autonomous run-phase entry; the Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory, with the identical binding strength as the `mode:` / bare-`ultracode` / `/goal` clauses; (b) concurrency ceiling — the steered fan-out respects the 3-5 concurrent `Agent()` ceiling (`orchestration-mode-selection.md` §C.2, applied equally to Mode 3 and Mode 4); (c) read-only scoping — the phrase carries a read-only investigation scope qualifier and shall NOT seed parallel WRITE fan-out (write work stays foreground-sequential per `agent-common-protocol.md` § Background Agent Execution). **Disambiguation**: the Claude Code UI tip — "Say 'fan out subagents' and Claude sends a team" — maps to **Mode 4** (parallel subagents: single-turn multi-`Agent()` spawn), NOT Mode 3 (agent-team, which requires the Agent Teams env prerequisites and carries the `--team` coupling). Default on ambiguity: omit.
  - **`ultracode` re-integration — bare opener keyword vs `/effort ultracode` session-persistence variant** [HARD]: the default opener form appends a **bare `ultracode`** keyword to the `ultrathink.` opener line (e.g. `ultrathink. ultracode`), which fires at paste time (v2.1.160+, same class as the `ultrathink` keyword), and is emitted ONLY when the seeded mode is `dynamic-workflow` (per the directive-binding table above). The **`/effort ultracode` slash form** is retained as a SEPARATE "session-persistence" variant for when ultracode must persist across the whole session rather than fire once at paste time — a `#`-commented slash line cannot execute at paste time, so it is not the opener default. Per `.claude/rules/moai/workflow/dynamic-workflows.md`, ultracode is NOT restored by the `ultrathink.` opener — it must be explicitly re-issued after `/clear` when the resumed session needs auto-orchestration. When the next SPEC does NOT declare workflow fan-out, no ultracode form is emitted (the `ultrathink.` opener alone suffices). The bare `ultracode` rides the opener line, which sits immediately after `ultrathink.` (or after the `mode:` line when present) per the line-order invariant above. Default on ambiguity: omit.
  - **Post-paste `/goal` follow-up (NOT a Block 1 line)** [HARD]: the `/goal` autonomous-continuation directive is NOT a Block 1 line and is NOT embedded anywhere in the main resume block — a mid-paste `#`-prefixed or bare slash line is inert (parsed as plain text; official slash-command recognition is input-start-only). When the emission condition holds (the next SPEC is run-phase AND declares a machine-verifiable end-state — condition UNCHANGED from the predecessor doctrine), the orchestrator emits a separate post-paste `/goal` follow-up block; the full two-step mechanism, the standalone-message requirement, the Implementation-Kickoff-Approval timing, and the resumed-session reminder obligation live in § Post-Paste /goal Follow-up Block. Per `.claude/rules/moai/workflow/goal-directive.md`, a `/goal` is NOT restored by the `ultrathink.` opener — `/clear` removes an active goal, so it must be re-issued as its own standalone user message when the resumed session needs the autonomous-continuation loop. Default on ambiguity: omit the follow-up block. **Implementation Kickoff Approval invariant**: a `/goal` follow-up block does NOT authorize autonomous run-phase entry; the Implementation Kickoff Approval human gate remains required before run-phase entry, independent of whether a follow-up block is emitted.
- **Block 2**: `applied lessons:` — relevant memory files from `~/.claude/projects/{hash}/memory/`. MUST include the most recent relevant project memory + any relevant lessons. Block 2 MUST also include a `source_session_id: <UUID from moai session current>` line carrying the Claude Code session_id of the orchestrator turn that generated this resume message per the canonical multi-session coordination policy. The session_id is the same value emitted by `moai session list --json` and stored in `.moai/state/active-sessions.json` — readers can correlate the resume back to its originating session.
  - **Environment fallback** [HARD]: the primary UUID source is `moai session current`. If `moai session current` returns the canonical fallback (runtime did not expose session.id to the CLI subprocess), OR `moai session list --json` returns error (CLI not installed in PATH), OR `.moai/state/active-sessions.json` does not exist (the multi-session coordination layer not yet deployed in this project), the orchestrator MUST emit the recognized fallback pattern verbatim: `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`. This pattern is NOT an anti-pattern; it is the prescribed graceful degradation when the CLI/registry layer is absent or the runtime does not expose session.id. The next session, upon `/moai session register` activation, MAY backfill the UUID by appending a `[backfilled: <UUID>]` annotation to the memory file's Block 2 line.
- **Block 3**: separator + `전제 검증:` (Korean) or `Preconditions:` (English).
- **Block 4**: numbered preconditions `<N>) <action> → <expected outcome>`. Each MUST be independently verifiable (git/gh command, file existence). Max 4 preconditions.
- **Block 5**: separator + `실행: <command-or-action>` — single primary action (typically `/moai <subcommand>`).
- **Block 6**: separator + `<workflow-context header>: <next-action-or-spec>` — RECOMMENDED for multi-SPEC Epics or follow-up; **omit entirely** for single-SPEC close with no further actions queued.
  - **Header selection (workflow-context conditional)**:
    - **PR-based workflow** (feat/* → PR → merge): `머지 후:` (en `After merge:`)
    - **Trunk-based no-PR** (e.g., 1-person OSS, all-tier direct-to-main push, no merge step): `후속:` (en `Follow-up:`)
    - **Single-SPEC close** (no further SPEC/phase queued): omit Block 6 entirely
  - **Single action principle**: `<next-action-or-spec>` MUST be one concrete SPEC ID, one command, or one phase transition — avoid vague "cycle-repeat" / "iteration loop" phrasing that reads as infinite recursion.

### Example (Illustrative; substitute project-specific values when adapting)

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

## Post-Paste /goal Follow-up Block

[ZONE:Evolvable] [HARD] The `/goal` autonomous-continuation directive is delivered as a **two-step handoff**, NOT as a line inside the main resume block. The main cut-line-bounded resume block (§ Canonical Format) is `/goal`-free; a slash command pasted mid-body is inert plain text because official slash-command parsing recognizes a `/` command only at the **input start** of a standalone message (`https://code.claude.com/docs/en/interactive-mode` § Quick commands), and `/goal` is a user-only TUI command the model cannot invoke on the user's behalf (`https://code.claude.com/docs/en/goal`). The two-step form below is what actually arms the autonomous-continuation loop in the resumed session.

**Mode scope (mode=manual fallback path)**: this two-step follow-up mechanism is the delivery path for the distributed default `handoff.mode: manual`, and it remains fully functional whenever auto-injection is unavailable, disabled, or skipped. Where `handoff.mode: auto` is configured, the one-message goal-first variant in § Auto-Injected Resume Flow applies instead; nothing in this section is weakened by that flow.

### Emission condition (frozen — unchanged from the predecessor doctrine)

The follow-up block is emitted ONLY when the next SPEC is run-phase AND declares a machine-verifiable end-state (a machine-checkable end-state such as the SPEC's test suite passing, a lint-clean state, or a bounded `stop after N turns` clause). When the condition does NOT hold (plan-phase / sync-phase next SPEC, or any next SPEC lacking a machine-verifiable end-state), NO follow-up block and NO instruction line are emitted — the output is byte-identical to the pre-existing no-`/goal` form.

### Block anatomy (two parts)

**Where** the emission condition holds, **When** the orchestrator emits a paste-ready resume message, it appends — OUTSIDE and AFTER the main cut-line block — the following two parts:

1. A localized **instruction line** (prose, OUTSIDE the cut-line markers) stating the block below MUST be sent as its own standalone message (slash commands parse only at input start), RECOMMENDED after the resumed session's Implementation Kickoff Approval — setting a goal starts a turn immediately, so sending it before Kickoff Approval would begin run-phase work prematurely. The instruction line text translates per `conversation_language` (see the § Localization Table instruction-line row).
2. A second cut-line-bounded fenced block (reusing the existing Cut-line top/bottom marker rows) containing EXACTLY one line — `/goal <completion-condition>` — with no `#` prefix and no additional lines.

Skeleton (illustrative; the instruction line renders in `conversation_language`, the `/goal` token and `<completion-condition>` placeholder stay locale-verbatim):

```text
<instruction line — localized per § Localization Table>

✂──── 여기부터 복사 ────✂

/goal <completion-condition>

✂──── 여기까지 복사 ────✂
```

### Resumed-session orchestrator reminder obligation

[ZONE:Evolvable] [HARD] The resumed session's orchestrator carries a **reminder obligation**: because the model cannot set a `/goal` on the user's behalf (it is a user-only TUI command), the orchestrator MUST remind the user to send the `/goal` line as a standalone message at the recommended moment (after Implementation Kickoff Approval). The reminder is issued via **natural-language status guidance, NOT `AskUserQuestion`** (it is an announcement, not a decision). Detection path: post-retirement the pasted main block itself carries NO `/goal` reference, so the resumed orchestrator detects the pending `/goal` from the **handoff memory entry** (the resume message AND its post-paste follow-up block are persisted verbatim to auto-memory per § Auto-Memory Integration) OR, failing a memory hit, by **re-deriving the emission condition** (the resumed SPEC is run-phase AND declares a machine-verifiable end-state). This reminder obligation is also recorded in `goal-directive.md` § MoAI Integration Notes.

### Implementation Kickoff Approval invariant (carry-over)

A `/goal` follow-up block does NOT authorize autonomous run-phase entry. The Implementation Kickoff Approval human gate (orchestrator `AskUserQuestion` per `goal-directive.md` § MoAI Integration Notes) remains required before run-phase entry, independent of whether a follow-up block was emitted. The follow-up block is a continuation-loop convenience, never a run-phase pre-authorization.

### Goal-first bootstrap variant (documented alternative — NOT the default)

[ZONE:Evolvable] An explicit alternative single-paste form exists: the **goal-first bootstrap** — a standalone one-line `/goal` message whose condition text carries both a resume pointer and the compact completion condition. Illustrative:

```text
/goal SPEC-X run 재개: memory의 <handoff-file>와 progress.md를 읽고 이어서 진행. 완료 조건: <machine-verifiable end-state>, 또는 N턴 후 중단.
```

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

[ZONE:Evolvable] Where the project config `.moai/config/sections/handoff.yaml` sets `handoff.mode: auto`, the saved pending record (§ Emission-Time Save Obligation) is consumed automatically at the next `/clear` session start, collapsing the resume to **ONE** user message. This section is the SSOT for the flow; the render surface (`.claude/output-styles/moai/moai.md` §8) carries a compact emission clause + pointer only.

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

### Invariants (both modes)

- **Implementation Kickoff Approval unchanged**: neither auto-injection nor a set goal pre-authorizes run-phase entry. The Implementation Kickoff Approval human gate remains required before run-phase entry in both modes.
- **Manual reversion is baseline-identical**: restoring `handoff.mode: manual` reverts runtime behavior to the pre-auto baseline — the injector's manual branch is a pure no-op that never touches the pending record, even a stale one — and the manual path documented in this file (6-block paste + § Post-Paste /goal Follow-up Block) is complete and self-sufficient without this section.
- **Fail-open everywhere**: save failures never block emission (§ Emission-Time Save Obligation); injection failures never block session start; a missing, stale, or already-claimed record degrades silently to the manual paste path.

## Auto-Memory Integration (Mandatory)

[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<epic>_<spec>_<status>.md` (e.g., `project_epic8_wf002_complete.md`). The `<epic>` token reflects the multi-SPEC grouping per sprint-round-naming.md (the legacy `<sprint>/<wave>` tokens are retired).
2. Include the resume message verbatim in that file under a `## 다음 세션 시작점 (paste-ready resume message)` heading. When a post-paste `/goal` follow-up block was emitted (per § Post-Paste /goal Follow-up Block), include its instruction line + cut-line-bounded `/goal` block verbatim in the same memory entry, so the resumed session can detect the pending `/goal` from memory (the pasted main block itself carries no `/goal` reference).
3. Update `MEMORY.md` index with a one-line entry pointing to the new memory file.
4. Mark superseded entries (if any) with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.
5. Annotate the MEMORY.md index entry with a `(session: <UUID-8-char-prefix>)` parenthetical when the SPEC was worked across multiple sessions (cross-references the `source_session_id` in Block 2 — enables readers to correlate the resume back to its originating session).
6. **Close-time pruning (auto-resume era)**: on SPEC close, the consumed verbatim resume block inside the memory topic file (the `## 다음 세션 시작점` section) SHOULD be pruned to a one-line summary — once the record has been consumed, verbatim preservation is owned by the `.moai/state/handoff/consumed/` audit trail, not the memory file. The generation-time verbatim-persistence obligation above (items 1-2) is UNCHANGED: the resume message is still saved verbatim to memory when emitted; the pruning binds only later, at SPEC close (temporal separation). Forward-looking only — no retroactive rewrite of existing memory files is mandated. This stops double-storage growth and keeps the always-loaded memory index within the loader's line/byte cap.

This ensures the message survives `/clear` and is discoverable at the start of the next session's context.

## Output Surface (User-Facing)

At session end, the orchestrator displays: (1) the main message in a fenced ```text``` block **bounded by cut-line markers** (per § Cut-line Marker Specification — marker text translated per `conversation_language`, `✂`/`─` symbols preserved verbatim) for verbatim paste, (2) **when the `/goal` emission condition holds** (next SPEC run-phase AND machine-verifiable end-state), the localized instruction line + the separate cut-line-bounded `/goal` follow-up block (per § Post-Paste /goal Follow-up Block) — omitted entirely otherwise, (3) the memory file path, (4) a one-sentence summary of what next session continues.

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

### Example with Block 0 (Illustrative)

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

## Diet Constraints

[ZONE:Evolvable] [HARD] A paste-ready resume message is "next session minimum executable context" — it is NOT an audit trail, history record, or ceremonial commitment record. Accumulating history/lesson/directive-escalation prose in the body via append-only across retry iterations is an empirically proven anti-pattern.

### Block 2 applied-lessons constraint

- At most **4 references** (memory file slug or lesson identifier)
- Each reference is a **single-line identifier** (e.g. `<lesson-id>` — full prose history is prohibited)
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

### Pre-emit self-check (paste-ready budget) — 10 items

- [ ] Block 2 ≤ 4 references
- [ ] Block 2 each reference is a single-line identifier (full history prohibited)
- [ ] Block 4 each precondition ≤ 200 chars
- [ ] Block 4 precondition prose has no embedded history
- [ ] Block 5 single primary action (command + one-line context max)
- [ ] Block 6 ≤ 2 lines
- [ ] Doctrine history not embedded → rule-file reference only
- [ ] No ceremonial reminder
- [ ] Post-paste `/goal` follow-up block (if emitted) is a separate cut-line-bounded block outside the main message, containing exactly one `/goal` line (never inside the main resume body)
- [ ] Block 1 fan-out steering phrase (`fan out subagents (<read-only investigation scope>)`) present iff mode = parallel-subagents — phrase locale-verbatim (English preserved), scope qualifier translated

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
**Drift-mitigation self-check sentinel (SSOT → render surface).** This file is the SSOT; `.claude/output-styles/moai/moai.md §8` is the render surface. Before committing any edit to the Localization Table, the 6-block skeleton, the cut-line marker spec, the Pre-emit self-check labels, § Emission-Time Save Obligation, or § Auto-Injected Resume Flow in THIS file, verify the parity check against the render surface: the moai.md §8 Localization Contract must carry the same locale column count (en / ko / ja / zh — 4 columns) as this file's Localization Table, the moai.md §8 Pre-emit self-check labels must use the same concern-name qualifiers (`paste-ready budget` / `localization render` / `session-handoff template completeness`) as this file, and the moai.md §8 emission clause (the `moai handoff save` save duty + auto-flow pointer) must remain a compact pointer consistent with § Emission-Time Save Obligation and § Auto-Injected Resume Flow here (pointer, NOT full duplication). If the two surfaces have diverged, this is the canonical surface — update the render surface to match.

- `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets — the per-model-class threshold SSOT for `/clear` and Trigger #1 (this file carries no inline model-class numbers to avoid label drift).
- `.claude/output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- `.claude/output-styles/moai/moai.md` §8 (Response Templates → Session Handoff) — the canonical render surface for the 6-block template + pre-emit self-check; this file is the SSOT, moai.md §8 is the render surface (bidirectional link).
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory + `[SUPERSEDED by ...]` convention
- `.moai/config/sections/handoff.yaml` — `handoff.mode` (`manual`/`auto`) + `handoff.guide` config keys consumed by § Auto-Injected Resume Flow
- `.claude/rules/moai/workflow/goal-directive.md` § MoAI Integration Notes — goal-first single-message cross-reference (auto-injected path)
- CLAUDE.md §11 (Error Handling) — token-limit recovery
- large-SPEC wave-split rationale
- `--worktree` Block 0 + single/multi-session decision rationale
- worktree isolation + --team base mismatch

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
