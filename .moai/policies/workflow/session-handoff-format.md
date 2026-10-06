---
description: Format companion to the session handoff protocol
paths: "**/session-handoff*.md,**/.claude/output-styles/moai/*.md,**/.moai/state/handoff/**"
---

# Session Handoff — Format and Activation Mechanics

> Owns: the cut-line marker specification, the locale tables, the activation matrix, the auto-injected-resume flow body, the emission-surface self-check, and the relocated cross-references. `session-handoff.md` is the SSOT and keeps every binding clause; sibling companion `session-handoff-examples.md` keeps the examples and appendices.

## Why This Matters

Long workflows accumulate context that exceeds the window or benefits from fresh start; without a standardized handoff, session boundaries lose work-in-progress. This rule defines when to emit a paste-ready resume, the 6-block structure, and auto-memory integration that persists across `/clear`.

### Cut-line Marker Specification

- Top marker: `✂──── 여기부터 복사 ────✂` (scissors U+2702 + 4× U+2500 + space + text + space + 4× U+2500 + scissors)
- Bottom marker: `✂──── 여기까지 복사 ────✂` (same structure, text differs)
- One blank line separates each marker from adjacent block content (top → blank → Block 1; Block 6 → blank → bottom)
- `✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute
- Box-drawing characters (`─` U+2500) preserved verbatim
- Marker text translates per `conversation_language` (see Localization table below)

### Localization Table

The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table carries the en / ko columns inline; the full 4-locale table (en / ko / ja / zh) lives in `session-handoff-examples.md` § Localization Table (Full 4-Locale). Cross-verified with `.claude/output-styles/moai/moai.md §8` (the canonical render surface).

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

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (keeping `✂` / `─` verbatim) and for each Block 1/3/5/6 placeholder and the memory heading (§ Auto-Memory Integration) when emitting the paste-ready message.

**Fallback rule for locales not in the table.** For ja / zh consult the full 4-locale table in `session-handoff-examples.md`; for any other ISO-639 code, English is the canonical fallback skeleton with each label translated to that locale via the naturalization principle (idiomatic phrasing, never literal transliteration) — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

## Paste-Time Activation Matrix

Handoff directives by activation mechanism: (a) paste-time keywords (`ultrathink`, bare `ultracode`) and (b) the fan-out phrase fire from a pasted body; (c) orchestrator-interpreted text (`mode:` seed, Block 5 `/moai …` including the `/moai goal` directive) routes via orchestrator reading, so it needs no standalone user message; (d) user-only TUI commands (`/effort`, `/clear`) fire ONLY as a standalone user message.

> **Full classification table**: `session-handoff-examples.md` § Paste-Time Activation Matrix.

### Pre-emit self-check (emission surface) — 3 items

- [ ] Is the cut-line-bounded block rendered in THIS response body — not only written to memory or persisted via the CLI?
- [ ] Are all three surface items present: the block, the memory file path, and the one-sentence continuation summary?
- [ ] Does the completion report avoid claiming the handoff was delivered when only the persistence steps ran?

## Anti-Patterns

> General resume-hygiene anti-pattern bullet list moved to `session-handoff-examples.md` § Anti-Patterns. See also § Diet Constraints (AP-D-001..005) and § V0 Abort Gate Doctrine (AP-V-001..004).

## Auto-Injected Resume Flow (mode=auto)

[ZONE:Evolvable] Where the project config `.moai/config/sections/handoff.yaml` sets `handoff.mode: auto`, the saved pending record (§ Emission-Time Save Obligation) is consumed automatically at the next `/clear` session start, collapsing the resume to **ONE** user message. `session-handoff.md` is the SSOT for the flow; the render surface (`.claude/output-styles/moai/moai.md` §8) carries a compact emission clause + pointer only.

> **One-message flow, /clear-only injection boundary, and resumed-turn precondition verification**: `session-handoff-examples.md` § Auto-Injected Resume Flow (mode=auto). In brief: at the next `/clear` (ONLY `clear` source) the handler claim-renames the pending record then injects the saved body verbatim; the user sends ONE message; injected preconditions are verified first.

## Cross-references

- `.claude/output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- CLAUDE.md §11 (Error Handling) — token-limit recovery

## Block 1 couplings

> Relocated from `session-handoff.md` § Field-by-Field Specification (t1303 always-loaded diet). The [HARD] line order and SEED rule stay in the stub.

- `mode: fanout` appends the locale-verbatim phrase `fan out subagents (<read-only investigation scope>)` to the Block 1 opener.
- `mode: agent-team` appends `--team` to the Block 5 command.
- `mode: sweep` appends a bare `ultracode` to the Block 1 opener.
- `mode:` is omitted entirely for `serial` (the default), keeping the common case byte-identical.
- `mode:` values and the fan-out phrase are protocol tokens preserved verbatim in every locale; only the parenthesized scope qualifier translates.
- Legacy pre-rename tokens (`solo-sequential`, `parallel-subagents`, `dynamic-workflow`) remain parse-accepted on read and map to `serial` / `fanout` / `sweep` — new emissions use the new tokens only.
