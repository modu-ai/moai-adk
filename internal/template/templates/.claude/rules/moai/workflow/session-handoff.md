# Session Handoff Protocol

Long-running session continuity: clean transitions across context boundaries via paste-ready resume messages.

> **Loading scope**: Intentionally always-loaded (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context, including those without SPEC files. The always-loaded cost is justified by cross-cutting applicability.

> **Format companion**: `session-handoff-format.md` owns the marker spec, locale tables, and activation mechanics relocated from this file — § Why This Matters · § Cut-line Marker Specification · § Localization Table · § Paste-Time Activation Matrix · § Auto-Injected Resume Flow (mode=auto) · § Pre-emit self-check (emission surface) — 3 items · § Anti-Patterns · § Cross-references (the relocated four). Sibling companion: `session-handoff-examples.md` (examples, appendices, the anti-pattern catalogue). Load a companion when rendering a handoff block, translating a label, or authoring output-style §8.

## When To Generate (5 Triggers)

[ZONE:Evolvable] [HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:

| # | Trigger | Detection |
|---|---------|-----------|
| 1 | Context usage crosses model-specific threshold (cumulative input+output) | Per-model-class percentage threshold — SSOT table: `context-window-management.md` § Context Window Targets (this file carries no inline model-class numbers to avoid label drift) |
| 2 | SPEC phase completion (plan/run/sync) within a multi-SPEC workflow | Phase boundary in `spec-workflow.md` §Phase Transitions (within a multi-SPEC SPEC ID series) |
| 3 | User explicitly requests session end ("세션 종료", "이번 세션 마무리", "next session") | Intent detection in user message |
| 4 | PR creation success when more SPECs remain in the current Epic | After `gh pr create` success + memory indicates >0 pending SPECs |
| 5 | Long-running multi-milestone task reaches a stable checkpoint | After milestone Mn complete + Mn+1 not yet started |

When NONE apply (single-turn, trivial, read-only), emit a brief completion confirmation. The `/clear` policy in `context-window-management.md` is co-anchored to Trigger #1's threshold.

### Emission-Time Save Obligation (auto-resume wiring)

[ZONE:Evolvable] [HARD] When the orchestrator emits a paste-ready resume message (any of the 5 triggers above), it MUST also persist the cut-line-bounded main block verbatim as the pending handoff record: write the block to a file with the Write tool, then redirect that file into `moai handoff save --stdin --spec <ID> --phase <phase> [--goal "<condition>"] [--ultrathink] [--ultracode] [--lang <conversation_language>] [--session <uuid>] < <file>` (body via stdin). Never pass the block as an inline heredoc: a heredoc body that names a git command is read as a git invocation — the branch guard denies it in the primary checkout and the Claude Code worktree guard refuses it in a worktree session — so the save is silently skipped under the fail-open rule below. `--goal` is recorded ONLY when the next SPEC is run-phase AND declares a machine-verifiable end-state (the same condition under which Block 5 carries a `/moai goal` directive); `--lang` snapshots the current `conversation_language`; `--session` carries Block 2's `source_session_id` when available.

[ZONE:Evolvable] [HARD] **Fail-open invariant**: when the `moai` CLI is absent from PATH or `moai handoff save` exits non-zero, the orchestrator emits the paste-ready surface UNCHANGED — a save failure never blocks, delays, or alters handoff emission, and no retry loop is entered. The manual paste path is fully functional without the save; the save is an additive persistence step, never a gate.

The saved record (`.moai/state/handoff/pending.json`) feeds the auto-injected flow under `handoff.mode: auto` (§ Auto-Injected Resume Flow). Under the distributed default `manual` it is inert — the injector never touches it, even when stale — and this save obligation still applies, so flipping the mode later needs no doctrine change.

## Canonical Format (Verbatim Spec)

[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers** (literal format: § Cut-line Marker Specification below). Cut-line markers sit **inside** the fenced text block so they are copied verbatim with the message — an unambiguous copy boundary in long terminal scrollback:

```
✂──── 여기부터 복사 ────✂

ultrathink. <SPEC-ID> <phase> <entering verb>.
mode: <value>   ← emit ONLY when the seeded orchestration mode ≠ serial; value ∈ {fanout | agent-team | sweep} (the Phase 4 catalog names) OMIT for serial (default) → v1 byte-identical. When mode = sweep, ALSO append bare `ultracode` to the opener line above (paste-time trigger keyword; the session-persistent `/effort ultracode` slash form is a separate variant — per Field-by-Field Spec, Block 1). When mode = agent-team, append `--team` to the Block 5 run command. When mode = fanout, append the fan-out steering phrase `fan out subagents (<read-only investigation scope>)` to the opener line above (paste-time steering phrase — per Field-by-Field Spec, Block 1).
applied lessons: <memory-file-1>, <memory-file-2>, ...

Preconditions:
1) <verifiable precondition 1>
2) <verifiable precondition 2>
N) <verifiable precondition N>

Run: <command-or-action>

After merge: <next-action-or-spec>

✂──── 여기까지 복사 ────✂
```

The `✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute (full marker spec: `session-handoff-format.md` § Cut-line Marker Specification).

### Field-by-Field Specification

Per-block detail — the `mode:` enum couplings, the fan-out steering phrase, the two `ultracode` forms, the `source_session_id` fallback, the Block 5 arm-only consequence — is in `session-handoff-examples.md` § Field-by-Field Specification; the binding clauses are summarized here.

- **Block 1** — `ultrathink.` opener (sets `effort: xhigh`; Adaptive Thinking is a separate axis it does not toggle). `<phase>` ∈ `plan | run | sync | mx`. [HARD] Fixed line order: opener (plus any appended keyword or steering phrase) → `mode:` → `applied lessons:` → `source_session_id:`, each conditional line omitted when its condition does not hold. A purpose-conditional `mode:` line seeds the next session's orchestration mode from the 4-token enum `serial | fanout | agent-team | sweep`; it is **omitted** for `serial` (the default). [HARD] Every mode coupling is a **SEED, not a permission grant** — Implementation Kickoff Approval remains mandatory, and a steered fan-out stays within the fanout bounds (`orchestration-mode-selection.md` §C.2) and is read-only-scoped. The per-mode couplings (`fanout`/`agent-team`/`sweep` append forms), protocol-token verbatim rule, and legacy-token mapping: `session-handoff-format.md` § Block 1 couplings.
- **Block 2** — `applied lessons:` naming the relevant memory files, plus `source_session_id: <UUID from moai session current>`. Where the CLI or registry is unavailable, emit the prescribed fallback line verbatim (the sidecar carries it) — graceful degradation, not an anti-pattern.
- **Block 3** — separator + `Preconditions:` header (locale rendering per § Localization Table).
- **Block 4** — numbered `<N>) <action> → <expected outcome>`, each verifiable by a command or a file check. Maximum 4.
- **Block 5** — separator + `Run:` carrying a **single primary action**, which is always the work-starting command. [HARD] `/moai goal` is arm-only and starts no work, so it never occupies this line alone — a goal armed with nothing running spins idle turns to the ceiling. Where the next SPEC declares a machine-verifiable end-state, the goal is armed *alongside* the primary action, after Implementation Kickoff Approval.
- **Block 6** — separator + a workflow-context header carrying exactly one next action: `After merge:` for a PR-based flow, `Follow-up:` for trunk-based no-PR. Omit the block entirely on a single-SPEC close with nothing queued.

## Auto-Injected Resume Flow (mode=auto)

[ZONE:Evolvable] Under `handoff.mode: auto` the saved pending record is consumed at the next `/clear` session start, collapsing the resume to **ONE** user message. Flow, the `/clear`-only injection boundary, and resumed-turn precondition verification: `session-handoff-format.md` § Auto-Injected Resume Flow (mode=auto) · `session-handoff-examples.md` § Auto-Injected Resume Flow (mode=auto).

### Invariants (both modes)

- **Implementation Kickoff Approval unchanged**: neither auto-injection nor a set goal pre-authorizes run-phase entry. The Implementation Kickoff Approval human gate remains required before run-phase entry in both modes.
- **Manual reversion is baseline-identical**: restoring `handoff.mode: manual` reverts runtime behavior to the pre-auto baseline — the injector's manual branch is a pure no-op that never touches the pending record, even a stale one — and the manual path documented in this file (the 6-block paste) is complete and self-sufficient without this section.
- **Fail-open everywhere**: save failures never block emission (§ Emission-Time Save Obligation); injection failures never block session start; a missing, stale, or already-claimed record degrades silently to the manual paste path.

## Auto-Memory Integration (Mandatory)

[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:

1. Save the message to a memory project entry. Filename pattern: `project_<epic>_<spec>_<status>.md` (e.g. `project_epic8_wf002_complete.md`; `<epic>` per sprint-round-naming.md — legacy `<sprint>/<wave>` tokens retired).
2. Include the resume message verbatim there under a `## Next Session Entry Point (paste-ready resume message)` heading (locale variant per the Localization Table memory-heading row; ko `## 다음 세션 시작점`).
3. Update the `MEMORY.md` index with a one-line entry pointing to it.
4. Mark superseded entries with a `[SUPERSEDED by <new-file>]` prefix per `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.
5. Annotate the index entry with `(session: <UUID-8-char-prefix>)` when the SPEC spanned multiple sessions (correlating to Block 2's `source_session_id`).
6. **Close-time pruning (auto-resume era)**: at SPEC close the consumed verbatim block in the memory topic file SHOULD shrink to a one-line summary — verbatim preservation then belongs to the `.moai/state/handoff/consumed/` audit trail. The generation-time obligation (items 1-2) is UNCHANGED; this binds only later, forward-looking, and stops double-storage growth in the always-loaded index.

The message then survives `/clear` and is discoverable at the next session's start.

## Output Surface (User-Facing)

[ZONE:Evolvable] [HARD] Emitting a resume message means **rendering it in the response body of the turn that generates it** — not storing it. At session end the orchestrator displays all three of: (1) the main message in a fenced ```text``` block **bounded by cut-line markers** (per § Cut-line Marker Specification — marker text translated per `conversation_language`, `✂`/`─` symbols preserved verbatim) for verbatim paste, (2) the memory file path, (3) a one-sentence summary of what next session continues.

**reference-instead-of-render (named anti-pattern).** Writing the resume into the memory topic file (§ Auto-Memory Integration) and merely *citing* that path in the completion report is NOT emission. The memory write and the `moai handoff save` record are persistence steps; neither reaches the user, so a report saying the resume "is saved in memory" while rendering no block leaves nothing to paste. The hazard is structural: persistence is discharged by visible tool calls and feels complete once they succeed, while the render is ordinary response text with no tool call to confirm it. Persistence without rendering is an unobserved completion claim under `verification-claim-integrity.md` §1.1 surface 1.

**Render-surface dependency.** The per-persona banner template lives in the active output style, and not every persona defines one. Where the active style carries no handoff banner this obligation still binds — render the cut-line-bounded block from the § Canonical Format skeleton, styled to that persona's banners. A missing template is never a reason to skip the render.

## Worktree-Anchored Resume Pattern

> [ZONE:Evolvable] [HARD] When the work happened inside a worktree, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure, and Block 4 gains precondition `0) git rev-parse --show-toplevel → <worktree-path>`. Block 0 uses the **canonical EnterWorktree-first forms** — `moai cc -w <name>` for a worktree under `.claude/worktrees/`, `moai cc -w <abs-path>` for one under `~/.moai/worktrees/`, or `EnterWorktree(<path>)` for current-session re-entry — NOT a bare `cd <worktree>` shell instruction. Work in the main checkout (the default) needs only the standard 6-block. Full: `session-handoff-examples.md` § Worktree-Anchored Resume Pattern.

## Diet Constraints

[ZONE:Evolvable] [HARD] A paste-ready resume message is "next session minimum executable context" — NOT an audit trail, history record, or ceremonial commitment record. Precondition body prose compresses to a one-line verifiable command + STRICT criterion (AP-D-002); Block 5 sub-step nesting compresses to a single primary action (AP-D-003). Full AP-D-001..005 catalogue + 9-item pre-emit checklist + V0 Abort Gate Doctrine: `session-handoff-examples.md`.

## V0 Abort Gate Doctrine

> [ZONE:Evolvable] [HARD] The paste-ready Block 4 V0 precondition uses **lsof + cwd cross-validation** (NOT a raw `ps aux` count). When V0-b ≥ 1 OR V0-c ≥ 3, spawning implementation agents is prohibited and the session ends (no force-through). Canonical: `session-handoff-examples.md` § V0 Abort Gate Doctrine.

## Cross-references

**Drift-mitigation self-check sentinel (SSOT → render surface).** This file is the SSOT; `.claude/output-styles/moai/moai.md §8` is the render surface. **Before committing any edit to this file**, run the parity check in `session-handoff-examples.md` § Drift-mitigation self-check sentinel — that companion is `paths:`-scoped to this file, so it is loaded whenever the check is owed.

- `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets — the per-model-class threshold SSOT for `/clear` and Trigger #1 (this file carries no inline model-class numbers to avoid label drift).
- `.claude/output-styles/moai/moai.md` §8 (Response Templates → Session Handoff) — the canonical render surface for the 6-block template + pre-emit self-check; this file is the SSOT, moai.md §8 is the render surface (bidirectional link).
- `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol — auto-memory + `[SUPERSEDED by ...]` convention
- `.moai/config/sections/handoff.yaml` — `handoff.mode` (`manual`/`auto`) + `handoff.guide` config keys consumed by § Auto-Injected Resume Flow
- `.claude/rules/moai/workflow/goal-directive.md` § Goal-Presentation Timing — the arm-only property and the Kickoff-gate timing that Block 5 implements; § MoAI Integration Notes — the auto-injected resume path
- Output-style §6 (Persistence & Context Awareness) and CLAUDE.md §11 (token-limit recovery): `session-handoff-format.md` § Cross-references

---

Status: HARD operational rule, applies to all multi-phase MoAI workflows
