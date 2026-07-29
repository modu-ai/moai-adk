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

[ZONE:Evolvable] [HARD] When the orchestrator emits a paste-ready resume message (any of the 5 triggers above), it MUST also persist the cut-line-bounded main block verbatim as the pending handoff record: pipe the block to `moai handoff save --stdin --spec <ID> --phase <phase> [--goal "<condition>"] [--ultrathink] [--ultracode] [--lang <conversation_language>] [--session <uuid>]` (body fed via stdin). `--goal` is recorded ONLY when the next SPEC is run-phase AND declares a machine-verifiable end-state — the same condition under which Block 5 carries a `/moai goal` directive (§ Canonical Format, Field-by-Field Block 5); `--lang` snapshots the current `conversation_language`; `--session` carries the same session id as Block 2's `source_session_id` when available.

[ZONE:Evolvable] [HARD] **Fail-open invariant**: when the `moai` CLI is absent from PATH or `moai handoff save` exits non-zero, the orchestrator emits the paste-ready surface UNCHANGED — a save failure never blocks, delays, or alters handoff emission, and no retry loop is entered. The manual paste path is fully functional without the save; the save is an additive persistence step, never a gate.

The saved record (`.moai/state/handoff/pending.json`) is consumed by the auto-injected flow when the project config sets `handoff.mode: auto` — see § Auto-Injected Resume Flow. Under the distributed default `handoff.mode: manual` the record is inert (the session-start injector never touches it, even when stale), and this save obligation still applies — flipping the mode later requires no doctrine change.

## Canonical Format (Verbatim Spec)

[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers** (see § Cut-line Marker Specification below for the literal marker format, Unicode-preservation rules, and locale translation contract). Cut-line markers sit **inside** the fenced text block alongside the content so they are copied verbatim with the message; this provides the user an unambiguous copy boundary in long terminal scrollback:

```
✂──── 여기부터 복사 ────✂

ultrathink. <SPEC-ID> <phase> <entering verb>.
mode: <value>   ← emit ONLY when the seeded orchestration mode ≠ solo-sequential; value ∈ {parallel-subagents | agent-team | dynamic-workflow} → Phase 4 Mode 4 / 3 / 6. OMIT for solo-sequential (default) → v1 byte-identical. When mode = dynamic-workflow, ALSO append bare `ultracode` to the opener line above (paste-time trigger keyword; the session-persistent `/effort ultracode` slash form is a separate variant — per Field-by-Field Spec, Block 1). When mode = agent-team, append `--team` to the Block 5 run command. When mode = parallel-subagents, append the fan-out steering phrase `fan out subagents (<read-only investigation scope>)` to the opener line above (paste-time steering phrase — per Field-by-Field Spec, Block 1).
applied lessons: <memory-file-1>, <memory-file-2>, ...

Preconditions:
1) <verifiable precondition 1>
2) <verifiable precondition 2>
N) <verifiable precondition N>

Run: <command-or-action>

After merge: <next-action-or-spec>

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

The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table carries the en / ko columns inline (the inline locales); the full 4-locale table (en / ko / ja / zh) lives in `session-handoff-examples.md` § Localization Table (Full 4-Locale). Cross-verified for consistency with `.claude/output-styles/moai/moai.md §8` (the canonical render surface).

| Element | English | Korean |
|---------|---------|--------|
| Cut-line top text | `Copy from here` | `여기부터 복사` |
| Cut-line bottom text | `Copy to here` | `여기까지 복사` |
| Block 1 entering verb | `entering` | `진입` |
| Block 3 Preconditions header | `Preconditions:` | `전제 검증:` |
| Block 5 Run header | `Run:` | `실행:` |
| Block 6 After-merge header (PR workflow) | `After merge:` | `머지 후:` |
| Block 6 Follow-up header (trunk no-PR) | `Follow-up:` | `후속:` |
| Memory heading | `## Next Session Entry Point` | `## 다음 세션 시작점` |

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers) while keeping `✂` and `─` characters verbatim, and substitute the locale rendering for each Block 1/3/5/6 placeholder and the memory heading (per § Auto-Memory Integration) when emitting the paste-ready message.

**Fallback rule for locales not in the table.** The inline table above lists concrete renderings for en / ko only. When `conversation_language` is ja, zh, or any other ISO-639 code whose language column is NOT in the inline table (e.g. `fr`, `de`, `es`, `pt`, `vi`), consult the full 4-locale table in `session-handoff-examples.md` for ja / zh renderings; for all other locales, English is the canonical fallback skeleton and each label translates to that locale using the naturalization principle (idiomatic phrasing a native reader expects, never literal word-by-word transliteration). In other words: locales not in the inline table fall back to the English column for the structural skeleton, with the label text rendered in the configured ISO-639 language — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

### Field-by-Field Specification

- **Block 1**: `ultrathink.` sets `effort: xhigh` on Opus 4.7+ (next session lacks accumulated reasoning). Adaptive Thinking is a DISTINCT axis — the thinking mode, explicitly enabled via `thinking: {type: "adaptive"}` — not something `ultrathink` toggles. `<phase>` ∈ `plan | run | sync | mx`.
  - **Block 1 line-order invariant** [HARD]: the Block 1 lines emit in this fixed order — `ultrathink.` opener (with an optional appended bare `ultracode` keyword or fan-out steering phrase) → `mode:` line (when present) → `applied lessons:` → `source_session_id:`. Each conditional line is omitted when its condition does not hold; in the common solo-sequential case, only the opener + `applied lessons:` remain, byte-identical to v1. The goal-arming directive is not a Block 1 line — it rides Block 5 alongside the work-starting action (see Block 5 below).
  - **Purpose-conditional `mode:` orchestration-seed line** [HARD]: Block 1 carries a purpose-conditional `mode: <value>` line that **seeds** the next session's Phase 4 orchestration mode. It is emitted ONLY when the seeded mode is NOT `solo-sequential`; for `solo-sequential` (the default) the line is **omitted**, keeping the message byte-identical to v1 (zero-diff common case). The `mode:` line sits directly below the `ultrathink.` opener. Its value is a protocol token drawn from a fixed 4-enum that maps 1:1 onto the Phase 4 mode catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A):

    | `mode:` value (seed) | Phase 4 catalog | Emission | Directive coupling |
    |----------------------|--------------------|----------|--------------------|
    | `solo-sequential` | Mode 5 (sub-agent, default fallback) | **omitted** (default) | omission = v1 byte-identical |
    | `parallel-subagents` | Mode 4 (parallel, 3-5 concurrent `Agent()`) | emitted | append `fan out subagents (<read-only investigation scope>)` to the opener line |
    | `agent-team` | Mode 3 (agent-team, implicit team) | emitted | append `--team` to the Block 5 run command |
    | `dynamic-workflow` | Mode 6 (workflow, orchestrator fan-out) | emitted | append bare `ultracode` to the opener line |

    - **Excluded modes**: Mode 1 (trivial) and Mode 2 (background) are NOT handoff-relevant seeds — a handoff never resumes into a trivial or background mode as its primary re-entry mode, so neither is assigned a `mode:` token.
    - **Threshold reuse (no new threshold)**: the seed derives from Phase 4's existing auto-select thresholds (domains ≥ 3 / files ≥ 10 / score ≥ 7, per `orchestration-mode-selection.md` §B.1). The `mode:` seed introduces NO new threshold.
    - **SEED, not a permission grant** [HARD]: the `mode:` value is a SEED (a signal for the next session's orchestrator), NOT a permission grant. The Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory regardless of the seeded mode — a seeded `dynamic-workflow` or `agent-team` does NOT authorize autonomous run-phase entry. The seed only pre-selects the orchestration shape the user is subsequently asked to approve.
    - **Directive binding**: `ultrathink.` is emitted always (v1 invariant); bare `ultracode` is appended to the opener line ONLY when mode = `dynamic-workflow`; `--team` is appended to the Block 5 run command only when mode = `agent-team`; the fan-out steering phrase `fan out subagents (<read-only investigation scope>)` is appended to the opener line ONLY when mode = `parallel-subagents` (see the fan-out steering phrase bullet below).
    - **solo-sequential emission policy (emit-discouraged + parse-accept)**: `solo-sequential` is the emit-discouraged default — its `mode:` line is not emitted (Block 1 omits it → v1 byte-identical). An explicit `mode: solo-sequential` line, should a producer choose to write one, is parse-accepted (forward-compatible) and read as Mode 5, merely redundant with the omitted default. The framing is single: prefer omission, accept an explicit value — the doctrine does not simultaneously discourage emission and forbid parsing.
    - **`mode:` is a locale-verbatim protocol token**: like the `plan | run | sync | mx` phase tokens, the `mode:` value is preserved verbatim across all locales and is NOT added as a row to any localization / cut-line / header translation table.
    - **JSON-twin forward-compat note**: there is no JSON twin currently (this doctrine is doctrine-only, no code). Where a JSON-twin representation of the resume message is later introduced, that twin shall set `schema_version: 2` and carry the `mode` field. This note records forward-compatibility only and triggers no code change now.
  - **Purpose-conditional fan-out steering phrase (mode = parallel-subagents)** [HARD]: when the seeded mode is `parallel-subagents`, the resume message appends the natural-language fan-out steering phrase — canonical form `fan out subagents (<read-only investigation scope>)` — after the opener text on the Block 1 opener line. The paste-ready message is **user-pasted**, so the phrase counts at the runtime layer as a user-authored explicit multi-agent opt-in — the same paste-time class as the `ultrathink` / bare `ultracode` keywords. Rationale: newer models spawn fewer subagents by default, and fan-out must be instructed explicitly (per `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy Principle 4); a resumed session parsing only the `mode:` metadata line silently under-spawns without this phrase. **Locale-verbatim phrase**: `fan out subagents` is a locale-verbatim protocol phrase — preserved in English across all locales, exactly like the `mode:` values, and NOT added as a row to any localization / cut-line / header translation table; only the parenthesized scope qualifier translates per `conversation_language` (e.g. ko: `fan out subagents (read-only 코드베이스 조사)`). **Invariants**: (a) SEED-not-permission — the phrase does NOT authorize autonomous run-phase entry; the Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory, with the identical binding strength as the `mode:` / bare-`ultracode` / Block 5 goal-arming clauses; (b) concurrency ceiling — the steered fan-out respects the 3-5 concurrent `Agent()` ceiling (`orchestration-mode-selection.md` §C.2, applied equally to Mode 3 and Mode 4); (c) read-only scoping — the phrase carries a read-only investigation scope qualifier and shall NOT seed parallel WRITE fan-out (write work stays foreground-sequential per `agent-common-protocol.md` § Background Agent Execution). **Disambiguation**: the Claude Code UI tip — "Say 'fan out subagents' and Claude sends a team" — maps to **Mode 4** (parallel subagents: single-turn multi-`Agent()` spawn), NOT Mode 3 (agent-team, which requires the Agent Teams env prerequisites and carries the `--team` coupling). Default on ambiguity: omit.
  - **`ultracode` re-integration — bare opener keyword vs `/effort ultracode` session-persistence variant** [HARD]: the default opener form appends a **bare `ultracode`** keyword to the `ultrathink.` opener line (e.g. `ultrathink. ultracode`), which fires at paste time (v2.1.160+, same class as the `ultrathink` keyword), and is emitted ONLY when the seeded mode is `dynamic-workflow` (per the directive-binding table above). The **`/effort ultracode` slash form** is retained as a SEPARATE "session-persistence" variant for when ultracode must persist across the whole session rather than fire once at paste time — a `#`-commented slash line cannot execute at paste time, so it is not the opener default. Per `.claude/rules/moai/workflow/dynamic-workflows.md`, ultracode is NOT restored by the `ultrathink.` opener — it must be explicitly re-issued after `/clear` when the resumed session needs auto-orchestration. When the next SPEC does NOT declare workflow fan-out, no ultracode form is emitted (the `ultrathink.` opener alone suffices). The bare `ultracode` rides the opener line, which sits immediately after `ultrathink.` (or after the `mode:` line when present) per the line-order invariant above. Default on ambiguity: omit.
- **Block 2**: `applied lessons:` — relevant memory files from `~/.claude/projects/{hash}/memory/`. MUST include the most recent relevant project memory + any relevant lessons. Block 2 MUST also include a `source_session_id: <UUID from moai session current>` line carrying the Claude Code session_id of the orchestrator turn that generated this resume message per the canonical multi-session coordination policy. The session_id is the same value emitted by `moai session list --json` and stored in `.moai/state/active-sessions.json` — readers can correlate the resume back to its originating session.
  - **Environment fallback** [HARD]: the primary UUID source is `moai session current`. If `moai session current` returns the canonical fallback (runtime did not expose session.id to the CLI subprocess), OR `moai session list --json` returns error (CLI not installed in PATH), OR `.moai/state/active-sessions.json` does not exist (the multi-session coordination layer not yet deployed in this project), the orchestrator MUST emit the recognized fallback pattern verbatim: `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`. This pattern is NOT an anti-pattern; it is the prescribed graceful degradation when the CLI/registry layer is absent or the runtime does not expose session.id. The next session, upon `/moai session register` activation, MAY backfill the UUID by appending a `[backfilled: <UUID>]` annotation to the memory file's Block 2 line.
- **Block 3**: separator + `Preconditions:` (English) or `전제 검증:` (Korean).
- **Block 4**: numbered preconditions `<N>) <action> → <expected outcome>`. Each MUST be independently verifiable (git/gh command, file existence). Max 4 preconditions.
- **Block 5**: separator + `Run: <command-or-action>` (English) or `실행: <command-or-action>` (Korean) — single primary action (typically `/moai <subcommand>`).
  - **Arm-only consequence — the goal directive is never the primary action** [HARD]: `/moai goal` is **arm-only**. It records the completion condition and the `stop-goal` evaluator blocks turn-end until that condition holds, but it starts no work of its own. A Block 5 line carrying only a goal-arming directive would therefore arm a condition with nothing running, and each turn-end would find the condition unmet with no work advancing it — the session spins idle turns to the ceiling. Block 5's single primary action therefore stays the work-starting command (`/moai run SPEC-X`).
  - **Where the goal rides instead**: where the next SPEC declares a machine-verifiable end-state, the orchestrator arms `/moai goal "<condition>"` **alongside** that primary action rather than in place of it, after Implementation Kickoff Approval, as the autonomous-vs-semi-autonomous progression choice offered at that gate. See `.claude/skills/moai/workflows/goal.md` § Progression Mode for the axis, and `.claude/rules/moai/workflow/goal-directive.md` § Goal-Presentation Timing for the arm-only property and the timing rule. Arming does not authorize autonomous run-phase entry; the Implementation Kickoff Approval human gate remains required. A goal-first bootstrap single-paste variant is documented in `session-handoff-examples.md` § Goal-first bootstrap variant as a non-default alternative.
- **Block 6**: separator + `<workflow-context header>: <next-action-or-spec>` — RECOMMENDED for multi-SPEC Epics or follow-up; **omit entirely** for single-SPEC close with no further actions queued.
  - **Header selection (workflow-context conditional)**:
    - **PR-based workflow** (feat/* → PR → merge): `After merge:` (ko `머지 후:`)
    - **Trunk-based no-PR** (e.g., 1-person OSS, all-tier direct-to-main push, no merge step): `Follow-up:` (ko `후속:`)
    - **Single-SPEC close** (no further SPEC/phase queued): omit Block 6 entirely
  - **Single action principle**: `<next-action-or-spec>` MUST be one concrete SPEC ID, one command, or one phase transition — avoid vague "cycle-repeat" / "iteration loop" phrasing that reads as infinite recursion.

> **Example**: see `session-handoff-examples.md` § Example (Illustrative; substitute project-specific values when adapting).

## Paste-Time Activation Matrix

Handoff directives by activation mechanism: (a) paste-time keywords (`ultrathink`, bare `ultracode`) and (b) the fan-out phrase fire from a pasted body; (c) orchestrator-interpreted text (`mode:` seed, Block 5 `/moai …` including the `/moai goal` directive) routes via orchestrator reading; (d) user-only TUI commands (`/effort`, `/clear`) fire ONLY as a standalone user message. The goal-arming directive is class (c) — the orchestrator reads and routes it, so it needs no standalone user message.

> **Full classification table**: `session-handoff-examples.md` § Paste-Time Activation Matrix.

## Auto-Injected Resume Flow (mode=auto)

[ZONE:Evolvable] Where the project config `.moai/config/sections/handoff.yaml` sets `handoff.mode: auto`, the saved pending record (§ Emission-Time Save Obligation) is consumed automatically at the next `/clear` session start, collapsing the resume to **ONE** user message. This section is the SSOT for the flow; the render surface (`.claude/output-styles/moai/moai.md` §8) carries a compact emission clause + pointer only.

> **One-message flow, /clear-only injection boundary, and resumed-turn precondition verification**: moved to `session-handoff-examples.md` § Auto-Injected Resume Flow (mode=auto). In brief: the previous session saves via `moai handoff save`; at the next `/clear` (ONLY `clear` source) the handler claim-renames the pending record then injects the saved body verbatim; the user sends ONE message; injected preconditions are verified first.

### Invariants (both modes)

- **Implementation Kickoff Approval unchanged**: neither auto-injection nor a set goal pre-authorizes run-phase entry. The Implementation Kickoff Approval human gate remains required before run-phase entry in both modes.
- **Manual reversion is baseline-identical**: restoring `handoff.mode: manual` reverts runtime behavior to the pre-auto baseline — the injector's manual branch is a pure no-op that never touches the pending record, even a stale one — and the manual path documented in this file (the 6-block paste) is complete and self-sufficient without this section.
- **Fail-open everywhere**: save failures never block emission (§ Emission-Time Save Obligation); injection failures never block session start; a missing, stale, or already-claimed record degrades silently to the manual paste path.

## Auto-Memory Integration (Mandatory)

[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<epic>_<spec>_<status>.md` (e.g., `project_epic8_wf002_complete.md`). The `<epic>` token reflects the multi-SPEC grouping per sprint-round-naming.md (the legacy `<sprint>/<wave>` tokens are retired).
2. Include the resume message verbatim in that file under a `## Next Session Entry Point (paste-ready resume message)` heading (locale variant per the Localization Table memory-heading row; e.g. ko `## 다음 세션 시작점`).
3. Update `MEMORY.md` index with a one-line entry pointing to the new memory file.
4. Mark superseded entries (if any) with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.
5. Annotate the MEMORY.md index entry with a `(session: <UUID-8-char-prefix>)` parenthetical when the SPEC was worked across multiple sessions (cross-references the `source_session_id` in Block 2 — enables readers to correlate the resume back to its originating session).
6. **Close-time pruning (auto-resume era)**: on SPEC close, the consumed verbatim resume block inside the memory topic file (the next-session-start-point section) SHOULD be pruned to a one-line summary — once the record has been consumed, verbatim preservation is owned by the `.moai/state/handoff/consumed/` audit trail, not the memory file. The generation-time verbatim-persistence obligation above (items 1-2) is UNCHANGED: the resume message is still saved verbatim to memory when emitted; the pruning binds only later, at SPEC close (temporal separation). Forward-looking only — no retroactive rewrite of existing memory files is mandated. This stops double-storage growth and keeps the always-loaded memory index within the loader's line/byte cap.

This ensures the message survives `/clear` and is discoverable at the start of the next session's context.

## Output Surface (User-Facing)

At session end, the orchestrator displays: (1) the main message in a fenced ```text``` block **bounded by cut-line markers** (per § Cut-line Marker Specification — marker text translated per `conversation_language`, `✂`/`─` symbols preserved verbatim) for verbatim paste, (2) the memory file path, (3) a one-sentence summary of what next session continues.

## Anti-Patterns

> General resume-hygiene anti-pattern bullet list moved to `session-handoff-examples.md` § Anti-Patterns. See also § Diet Constraints (AP-D-001..005) and § V0 Abort Gate Doctrine (AP-V-001..004).

## Worktree-Anchored Resume Pattern

> [ZONE:Evolvable] [HARD] When the SPEC was initialized via L3 `/moai plan --worktree`, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure, and Block 4 gains precondition `0) git rev-parse --show-toplevel → <worktree-path>`. Block 0 uses the **canonical EnterWorktree-first forms** — `moai cc -w <name>` for L1 (`.claude/worktrees/`), `moai cc -w <abs-path>` for L2 (`~/.moai/worktrees/`), or `EnterWorktree(<path>)` for current-session re-entry — NOT a bare `cd <worktree>` shell instruction. Without `--worktree` (the default) the standard 6-block suffices. Full: `session-handoff-examples.md` § Worktree-Anchored Resume Pattern.

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
- **AP-D-003**: Block 5 sub-step nesting (Phase 0 + Phase 1 + Phase 1B style multi-phase 11-substep) → compress into a single primary action; sub-detail belongs in SPEC artifacts
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
- [ ] Block 1 fan-out steering phrase (`fan out subagents (<read-only investigation scope>)`) present iff mode = parallel-subagents — phrase locale-verbatim (English preserved), scope qualifier translated

### Applicable scope

- All new paste-ready resume messages
- Retry-iteration paste-ready messages (diet vs body-accumulation choice → diet is the default)
- Applied consistently across the line (all SPEC lines)

## V0 Abort Gate Doctrine

> [ZONE:Evolvable] [HARD] The paste-ready Block 4 V0 precondition uses **lsof + cwd cross-validation** (NOT a raw `ps aux` count). When V0-b ≥ 1 OR V0-c ≥ 3, spawning implementation agents is prohibited and the session ends (no force-through). Canonical: `session-handoff-examples.md` § V0 Abort Gate Doctrine.

## Cross-references

<!-- self-check sentinel — references the render surface's structural invariant by content, not line number, so it survives line drift. This is mitigation + visibility (it surfaces drift to a reading editor), NOT mechanical prevention. A future editor who changes one surface without reading the other surface's sentinel produces silent drift; the only mechanical catch is a deferred Go lint rule (see the session-handoff SSOT-align doctrine §F.6 follow-up). -->
**Drift-mitigation self-check sentinel (SSOT → render surface).** This file is the SSOT; `.claude/output-styles/moai/moai.md §8` is the render surface. Before committing any edit to the Localization Table, the 6-block skeleton, the cut-line marker spec, the Pre-emit self-check labels, § Emission-Time Save Obligation, or § Auto-Injected Resume Flow in THIS file, verify the parity check against the render surface: the moai.md §8 Localization Contract carries the full 4-locale table (en / ko / ja / zh); this file carries the en / ko subset inline with the ja / zh columns relocated to `session-handoff-examples.md`, the moai.md §8 Pre-emit self-check labels must use the same concern-name qualifiers (`paste-ready budget` / `localization render` / `session-handoff template completeness`) as this file, and the moai.md §8 emission clause (the `moai handoff save` save duty + auto-flow pointer) must remain a compact pointer consistent with § Emission-Time Save Obligation and § Auto-Injected Resume Flow here (pointer, NOT full duplication). If the two surfaces have diverged, this is the canonical surface — update the render surface to match.

- `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets — the per-model-class threshold SSOT for `/clear` and Trigger #1 (this file carries no inline model-class numbers to avoid label drift).
- `.claude/output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- `.claude/output-styles/moai/moai.md` §8 (Response Templates → Session Handoff) — the canonical render surface for the 6-block template + pre-emit self-check; this file is the SSOT, moai.md §8 is the render surface (bidirectional link).
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory + `[SUPERSEDED by ...]` convention
- `.moai/config/sections/handoff.yaml` — `handoff.mode` (`manual`/`auto`) + `handoff.guide` config keys consumed by § Auto-Injected Resume Flow
- `.claude/rules/moai/workflow/goal-directive.md` § Goal-Presentation Timing — the arm-only property and the Kickoff-gate timing that Block 5 implements; § MoAI Integration Notes — the auto-injected resume path
- CLAUDE.md §11 (Error Handling) — token-limit recovery
- large-SPEC wave-split rationale
- `--worktree` Block 0 + single/multi-session decision rationale
- worktree isolation + --team base mismatch

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
