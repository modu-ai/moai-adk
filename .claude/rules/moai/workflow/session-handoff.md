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

Per-block detail — the `mode:` orchestration-seed enum and its directive couplings, the fan-out steering phrase, the two `ultracode` forms, the `source_session_id` environment fallback, and the Block 5 arm-only consequence — is in `session-handoff-examples.md` § Field-by-Field Specification (full). The binding clauses are summarized here.

- **Block 1** — `ultrathink.` opener (sets `effort: xhigh`; Adaptive Thinking is a separate axis it does not toggle). `<phase>` ∈ `plan | run | sync | mx`. [HARD] Fixed line order: opener (plus any appended keyword or steering phrase) → `mode:` → `applied lessons:` → `source_session_id:`, each conditional line omitted when its condition does not hold. A purpose-conditional `mode:` line seeds the next session's orchestration mode from the 4-token enum `solo-sequential | parallel-subagents | agent-team | dynamic-workflow`; it is **omitted** for `solo-sequential` (the default), keeping the common case byte-identical. Its couplings: `parallel-subagents` appends the locale-verbatim phrase `fan out subagents (<read-only investigation scope>)` to the opener, `agent-team` appends `--team` to the Block 5 command, `dynamic-workflow` appends a bare `ultracode` to the opener. [HARD] Every one of these is a **SEED, not a permission grant** — Implementation Kickoff Approval remains mandatory, and a steered fan-out stays within the 3-5 concurrent `Agent()` ceiling and is read-only-scoped. `mode:` values and the fan-out phrase are protocol tokens preserved verbatim in every locale; only the parenthesized scope qualifier translates.
- **Block 2** — `applied lessons:` naming the relevant memory files, plus `source_session_id: <UUID from moai session current>`. When the CLI or registry is unavailable, emit the prescribed fallback line verbatim (see the sidecar); that fallback is graceful degradation, not an anti-pattern.
- **Block 3** — separator + `Preconditions:` header (locale rendering per § Localization Table).
- **Block 4** — numbered `<N>) <action> → <expected outcome>`, each independently verifiable by a command or a file check. Maximum 4.
- **Block 5** — separator + `Run:` carrying a **single primary action**, which is always the work-starting command. [HARD] `/moai goal` is arm-only and starts no work, so it never occupies this line alone — a goal armed with nothing running spins idle turns to the ceiling. Where the next SPEC declares a machine-verifiable end-state, the goal is armed *alongside* the primary action, after Implementation Kickoff Approval.
- **Block 6** — separator + a workflow-context header carrying exactly one next action: `After merge:` for a PR-based flow, `Follow-up:` for trunk-based no-PR. Omit the block entirely on a single-SPEC close with nothing queued.

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

[ZONE:Evolvable] [HARD] Emitting a resume message means **rendering it in the response body of the turn that generates it** — not storing it. At session end the orchestrator displays all three of: (1) the main message in a fenced ```text``` block **bounded by cut-line markers** (per § Cut-line Marker Specification — marker text translated per `conversation_language`, `✂`/`─` symbols preserved verbatim) for verbatim paste, (2) the memory file path, (3) a one-sentence summary of what next session continues.

**reference-instead-of-render (named anti-pattern).** Writing the resume into the memory topic file (§ Auto-Memory Integration) and then merely *citing* that file path in the completion report is NOT emission. The memory write and the `moai handoff save` record are both persistence steps; neither reaches the user, so a report stating the resume "is saved in memory" while rendering no block leaves the user with nothing to paste. The hazard is structural rather than a lapse of attention: § Auto-Memory Integration is discharged by concrete tool calls whose results are visible, so the turn feels complete once they succeed — while the render, being ordinary response text, is the one step with no tool call to confirm it happened. Persistence without rendering is an unobserved completion claim under `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1.

**Render-surface dependency.** The per-persona banner template for this render lives in the active output style, and not every persona defines one. Where the active style carries no handoff banner, this obligation still binds unchanged — render the cut-line-bounded block from the § Canonical Format skeleton directly, styled to match that persona's other banners. A missing banner template is never a reason to skip the render.

### Pre-emit self-check (emission surface) — 3 items

- [ ] Is the cut-line-bounded block rendered in THIS response body — not only written to memory or persisted via the CLI?
- [ ] Are all three surface items present: the block, the memory file path, and the one-sentence continuation summary?
- [ ] Does the completion report avoid claiming the handoff was delivered when only the persistence steps ran?

## Anti-Patterns

> General resume-hygiene anti-pattern bullet list moved to `session-handoff-examples.md` § Anti-Patterns. See also § Diet Constraints (AP-D-001..005) and § V0 Abort Gate Doctrine (AP-V-001..004).

## Worktree-Anchored Resume Pattern

> [ZONE:Evolvable] [HARD] When the work happened inside a worktree, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure, and Block 4 gains precondition `0) git rev-parse --show-toplevel → <worktree-path>`. Block 0 uses the **canonical EnterWorktree-first forms** — `moai cc -w <name>` for a worktree under `.claude/worktrees/`, `moai cc -w <abs-path>` for one under `~/.moai/worktrees/`, or `EnterWorktree(<path>)` for current-session re-entry — NOT a bare `cd <worktree>` shell instruction. Work in the main checkout (the default) needs only the standard 6-block. Full: `session-handoff-examples.md` § Worktree-Anchored Resume Pattern.

## Diet Constraints

[ZONE:Evolvable] [HARD] A paste-ready resume message is "next session minimum executable context" — NOT an audit trail, history record, or ceremonial commitment record. Two concrete anti-patterns illustrate the hazard (full AP-D-001..005 catalogue + 9-item pre-emit checklist + V0 Abort Gate Doctrine in `session-handoff-examples.md`):

- **AP-D-002**: precondition body prose (history/lesson narrative/cumulative pattern) → keep only a one-line verifiable command + STRICT criterion (≤ 200 chars).
- **AP-D-003**: Block 5 sub-step nesting (Phase 0 + Phase 1 + Phase 1B style multi-phase 11-substep) → compress into a single primary action; sub-detail belongs in SPEC artifacts.

Full Diet Constraints catalogue (AP-D-001..005 + 9-item pre-emit checklist) and V0 Abort Gate Doctrine in `session-handoff-examples.md` § Diet Constraints (Full Catalogue) + § V0 Abort Gate Doctrine.

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
- Block 0 cwd anchoring + the single/multi-session decision rationale
- worktree isolation + --team base mismatch

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
