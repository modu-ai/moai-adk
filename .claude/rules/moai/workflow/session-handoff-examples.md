---
description: "Illustrative examples and full 4-locale localization table for the session-handoff doctrine"
paths: "**/session-handoff.md"
---

# Session Handoff — Examples and Full Localization Table

> This is a path-scoped reference file for `session-handoff.md`. It holds illustrative Example sections and the full 4-locale Localization Table extracted from the always-loaded doctrine file to reduce context weight. The core doctrine (6-block skeleton, cut-line markers, the binding per-block summary, Pre-emit self-check, Auto-Memory Integration, Diet Constraints) remains in `session-handoff.md`; the full per-block detail was relocated here — see § Field-by-Field Specification (full).

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
```

> Block 5 carries the work-starting action. Where the next SPEC declares a machine-verifiable end-state, the orchestrator arms `/moai goal "<condition>"` alongside it after Implementation Kickoff Approval — arm-only, so it never replaces the `실행:` action (§ Canonical Format, Field-by-Field Block 5).

## Example with Block 0 (Illustrative)

```
✂──── 여기부터 복사 ────✂

[New Terminal — START IN WORKTREE]
$ moai cc -w ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
   # (launcher -w accepts L2 absolute paths; or moai glm -w ... | moai cg -w ...)

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

[ZONE:Evolvable] An explicit alternative single-paste form exists: the **goal-first bootstrap** — a one-line `/moai goal` message whose condition text carries both a resume pointer and the compact completion condition. Illustrative:

```text
/moai goal "resume SPEC-X run: read <handoff-file> from memory and progress.md, then continue. Completion: <machine-verifiable end-state>, or stop after N turns."
```

(The condition text follows the user's `conversation_language`; shown above in English-canonical form. The `/moai goal` token itself is locale-verbatim.)

Normative content:

- **(a) Selection criterion**: choose goal-first bootstrap when the user wants one-paste + autonomous continuation; the standard 6-block paste (§ Canonical Format) remains the DEFAULT.
- **(b) Caveats**: effort keywords (`ultrathink` / `ultracode`) placed inside a command argument are NOT documented to fire — the session may run at default effort; and precondition verification shifts from paste-time structure (the Block 4 verifiable commands) to **model discretion** via the directive text.
- **(c) Invariants preserved**: the condition must stay compact (one measurable end state); the Implementation Kickoff Approval gate is unaffected — arming never authorizes autonomous run-phase entry; the `/moai goal` token stays locale-verbatim (never translated).

## Paste-Time Activation Matrix

[ZONE:Evolvable] The following normative table classifies every handoff directive by its activation mechanism, so an author never places a directive where it cannot fire. Ground truth: `https://code.claude.com/docs/en/interactive-mode` (slash commands recognized only at input start).

| Class | Directives | Mechanism | Fires from pasted body? |
|-------|-----------|-----------|------------------------|
| (a) Paste-time keyword | `ultrathink`, bare `ultracode` | Runtime keyword, position-independent in message text | YES |
| (b) Paste-time natural-language phrase | `fan out subagents (<scope>)` | Explicit multi-agent opt-in phrase — same opt-in class as (a) | YES |
| (c) Orchestrator-interpreted text | `mode:` seed, Block 5 `실행: /moai <subcommand>`, the `/moai goal` directive | The orchestrator reads the text and routes (`/moai` via the Skill tool); NOT auto-executed as a slash command | YES (via orchestrator interpretation) |
| (d) User-only TUI command | `/effort`, `/clear` | Slash command parsed ONLY at input start; not model-invocable; cannot be set by pasted body text NOR by the model | NO — requires a standalone user message |

Consequence: the goal-arming directive belongs to class (c), not class (d) — the orchestrator reads and routes it, so it needs no standalone user message and may ride the pasted body. The class-(d) constraint still binds `/effort` and `/clear`.

The same classification governs the auto-injected body (§ Auto-Injected Resume Flow): content delivered as session-start context injection is inert context — it cannot fire class (a)/(b) paste-time keywords on its own and cannot execute class (d) commands. That is why the auto flow's ONE user message carries the class (a) `ultrathink` keyword in the user's own message, and why the class (c) goal directive can instead be armed by the orchestrator once the resumed turn begins.

## Auto-Injected Resume Flow (mode=auto)

### One-message flow

1. The previous session emits the paste-ready resume AND persists it via `moai handoff save` (§ Emission-Time Save Obligation). The paste-ready surface is still displayed — the user can always fall back to the manual paste path.
2. The user runs `/clear`.
3. The session-start handler, in the single consume cell (session source is `clear` AND `handoff.mode: auto` AND a live pending record exists), **claim-renames** the pending record into a `consumed/` audit-trail copy FIRST, then injects the saved content as session-start additional context. The claim-then-inject atomic rename means exactly one of two racing sessions injects — the loser's rename fails and it skips injection fail-open. A record older than the stale TTL is cleaned up instead of injected.
4. What the injection actually contains: a localized header; a disclaimer stating the injection only delivers context and does NOT automatically enable any extended-reasoning mode; restoration-guidance lines for the recorded directives (`ultrathink` / `/effort ultracode` / the recorded goal condition — each rendered as manual-input guidance the user may type, never as an executed command); and the saved body **verbatim** (no re-localization — `--lang` snapshots the language at save time). The injected context cannot start a turn and cannot claim effort restoration; the platform caps session-start injected context at 10,000 characters, and the § Diet Constraints budget keeps the 6-block body far below that cap.
5. The user sends **ONE** message:
   - **Goal-first variant** — Where the next SPEC is run-phase AND declares a machine-verifiable end-state, the one message MAY be a single `/moai goal "<condition>"` line, which the orchestrator interprets and routes (class (c) in § Paste-Time Activation Matrix). Because the directive is class (c) rather than class (d), the orchestrator may equally arm the condition itself once the resumed turn begins — the user is not obliged to type it.
   - **Approval variant** — otherwise, the one message is a short approval/continue message. Keep recommending that the user include the `ultrathink` keyword in this first message: the injected context cannot restore effort, but a paste-time keyword in the user's own message can.
   - **Effort caveat (goal-first)**: effort keywords placed inside a command argument are NOT documented to fire — a goal line carrying `ultrathink` inside its condition text may leave the session at default effort. The doctrine does not claim the goal-first variant restores extended reasoning.

### /clear-only injection boundary

Injection happens ONLY when the session-start source is `clear`. All other session-start sources — `startup`, `resume`, `compact` — are **notice-only**: the pending record is never consumed there, and with the `handoff.guide` key at its default `false` the notice is silent (no visible output; when `guide: true`, a best-effort stderr hint mentions the waiting record). Consequences:

- A terminal restart (new session process, source `startup`) does NOT auto-inject — the manual paste path applies.
- An L3 worktree Block 0 resume (new terminal inside the worktree, source `startup`) falls OUTSIDE auto-inject — Block 0 + the manual paste path remain the mechanism (§ Worktree-Anchored Resume Pattern).
- Only the in-place `/clear` boundary gets the one-message flow.

### Precondition verification at resumed-turn start

The injected Block 4 preconditions MUST be verified at the start of the resumed session's first working turn — injection delivers the TEXT of the preconditions, not their truth. This is most acute in the goal-first variant, where arming begins a turn immediately: the orchestrator verifies the injected preconditions FIRST, before acting on the goal condition.

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
- Putting a bare `/moai goal` directive in Block 5 as the single primary action — the directive is arm-only, so the session would arm a condition with no work running and spin idle turns to the ceiling. Block 5 keeps the work-starting command; the goal is armed alongside it (see `session-handoff.md` § Canonical Format, Field-by-Field Block 5).
- Omitting the fan-out steering phrase (`fan out subagents (<read-only investigation scope>)`) when `mode: parallel-subagents` is seeded — the resumed session silently under-spawns: fewer subagents are spawned by default unless fan-out is explicitly instructed (per `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy Principle 4; the fan-out steering phrase is NOT restored by the `ultrathink.` opener).

## Worktree-Anchored Resume Pattern

[ZONE:Evolvable] [HARD] When the work happened inside a worktree, the resume message MUST include **Block 0 (cwd anchoring)** prepended before the standard 6-block structure. Without Block 0, the next session starts in the main project cwd by default, breaking the worktree isolation the work relied on.

> Working inside a worktree is **user opt-in** only. For work done in the main checkout (the default), the standard 6-block structure suffices — Block 0 is NOT required.

### Why Block 0 (worktree work only)

When the work lives in a worktree, the SPEC artifacts and the L1 isolation base sit in a different cwd. Pasting resume into a main-cwd session causes: L1 base divergence per the worktree isolation guidance, Bash commands targeting main project per the worktree isolation guidance, build/test from the wrong tree. Block 0 forces a new terminal session **inside** the L2 worktree before any action.

### Block 0 Format

Block 0 is **prepended** before Block 1. Two forms exist; pick by **where the worktree lives**.

**Form A — `.claude/worktrees/<name>/` worktree (single command, preferred where it applies):**

```
[New Terminal — START IN WORKTREE]
$ moai cc -w <worktree-name>     # or: moai glm -w <name> | moai cg -w <name>
   └─ Claude Code session starts here (cwd = .claude/worktrees/<name>/)
```

`-w <name>` takes the **worktree name**, not a branch name and not a SPEC ID; it resolves to `.claude/worktrees/<name>/`. An existing worktree of that name is **reused, not recreated**, which is what makes this a valid re-entry path. Naming the worktree after the SPEC ID at creation time (`git worktree add -b feat/SPEC-X-001 .claude/worktrees/SPEC-X-001 origin/main`) lets the resume line read `moai cc -w SPEC-X-001`.

**Form B — L2 worktree at `~/.moai/worktrees/<project>/<spec>/` (cross-session launch via extended `-w`):**

```
[New Terminal — START IN WORKTREE]
$ moai cc -w <worktree-absolute-path>     # or: moai glm -w <abs-path> | moai cg -w <abs-path>
   └─ Claude Code session starts here (cwd = the L2 worktree at the given absolute path)
```

The launcher's `-w` flag accepts BOTH short names (Form A — resolved against `.claude/worktrees/<name>/`) AND absolute paths under `~/.moai/worktrees/<project>/...` (Form B — L2 persistent worktrees). This makes `moai cc -w <abs-path>` a valid re-entry path for L2 worktrees, replacing the legacy bare-`cd` shell form. An absolute path NOT under `~/.moai/worktrees/` or `.claude/worktrees/` is REJECTED with a clear error so the launcher does not silently create a new worktree under the wrong prefix.

For **current-session re-entry** into an L2 worktree (no `/clear`, same session continuing), prefer the Claude Code runtime tool `EnterWorktree(<worktree-absolute-path>)` over the launcher form — see `.claude/rules/moai/workflow/worktree-integration.md` § `EnterWorktree` / `ExitWorktree` Tools for the EnterWorktree-first policy.

> **DEPRECATED — Legacy shell form (do NOT emit in orchestrator Block 0 guidance):**
>
> ```
> $ cd <worktree-absolute-path> && <launcher>
> ```
>
> The bare-`cd` form breaks `Agent(isolation: "worktree")` CWD isolation and was the root cause of prior incidents where a sub-agent used `git -C` instead of `EnterWorktree`. It remains valid ONLY for human-typed, manual-shell contexts. Orchestrator-emitted Block 0 guidance SHALL use Form A, Form B, or `EnterWorktree(<path>)` — never a bare `cd`.

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

Block 0 is REQUIRED only for worktree work. For `--branch` (or no flag — the default), the standard 6-block suffices because the main session cwd already follows the branch.

[ZONE:Evolvable] [HARD] If a worktree was used and the user is NOT comfortable with a multi-terminal/multi-session workflow, the orchestrator SHOULD recommend `--branch` in the main checkout for the next SPEC. Forcing Block 0 onto a single-session user is friction without benefit. See the single-session vs multi-session decision rationale below.

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

---

## Field-by-Field Specification (full)

> Relocated from `session-handoff.md` § Canonical Format when the block-detail was
> split out to shrink the always-loaded rule footprint. `session-handoff.md` retains
> the binding summary and points here for the per-block detail. Content unchanged.

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

---

## Diet Constraints (Full Catalogue)

> Relocated from `session-handoff.md` § Diet Constraints to shrink the always-loaded footprint. The always-loaded file retains a 2-example inline summary + 1-line pointer; this section holds the full AP-D-001..005 catalogue, the 9-item pre-emit checklist, and the applicable-scope statement. Content unchanged.

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

### Anti-pattern catalogue (AP-D-001..005)

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

