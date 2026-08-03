# acceptance.md — SPEC-INFINITE-GOAL-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-INFINITE-GOAL-NNN); this file does NOT restate them as requirements.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-INFINITE-GOAL-001 | REQ-001 | `--max-turns 0` disables ceiling check | MUST | plan M1 |
| AC-INFINITE-GOAL-002 | REQ-001 | `--max-turns` omitted = 30 (backward compat) | MUST | plan M1 |
| AC-INFINITE-GOAL-003 | REQ-002 | block-cap doctrine surfaces the silent cap | MUST | plan M2 |
| AC-INFINITE-GOAL-004 | REQ-003 | armed goal → statusline `/clear` directive suppressed | MUST | plan M3 |
| AC-INFINITE-GOAL-005 | REQ-004 | `--max-turns 0` bounded by wall-clock (or cost) | MUST | plan M4 |
| AC-INFINITE-GOAL-006 | REQ-004 | stagnation guard fires on N-turn mechanical no-change | MUST | plan M4 |
| AC-INFINITE-GOAL-007 | REQ-005 | post-compact re-inject emits goal + SPEC + next action | MUST | plan M5 |
| AC-INFINITE-GOAL-008 | REQ-006 | `/clear` re-arms embedded goal under new session-id | MUST | plan M6 |
| AC-INFINITE-GOAL-009 | REQ-007 | ac_converge "Max 20" parses OR doc corrected | MUST | plan M1/M7 |
| AC-INFINITE-GOAL-010 | cross-ref | armed infinite goal does NOT weaken deny/ask | MUST | plan M8 |

### §D.1 Severity model

All ten ACs are MUST-pass. AC-005 + AC-006 together are the load-bearing safety pair for an infinite goal: `MaxTurns=0` without a real bound (wall-clock/cost) and a strengthened stagnation guard would permit unbounded cost / unbounded looping. AC-010 is the cross-cutting safety regression guard (deny/ask invariance holds even when the goal is infinite). The other seven are P0 behavior / continuity ACs.

### §D.2 AC definitions (Given-When-Then)

#### AC-INFINITE-GOAL-001 — `--max-turns 0` disables the ceiling check (REQ-1)

**Given** the `/moai goal` arm verb invoked with `--max-turns 0`.

**When** the goal is armed and the evaluator runs.

**Then** `Ceiling.MaxTurns == 0`, AND the `evaluate.go:142` `if g.Ceiling.MaxTurns > 0 && g.TurnsUsed >= ...` guard is NOT taken across simulated turns > 30 (i.e. the ceiling check is skipped at turns 31, 50, 100), AND the goal continues evaluating.

**Test shape:** Go unit test — arm with `--max-turns 0`, simulate 100 turns with a never-satisfied condition, assert the evaluator does NOT emit the MaxTurns-ceiling verdict at turn 30 or any subsequent turn. (Note: the test MUST also assert AC-005/AC-006's real bound fires before the simulation runs away — see those ACs.)

#### AC-INFINITE-GOAL-002 — `--max-turns` omitted defaults to 30 (REQ-1 backward compat)

**Given** the `/moai goal` arm verb invoked WITHOUT `--max-turns`.

**When** the goal is armed.

**Then** `Ceiling.MaxTurns == 30` (today's default), AND the evaluator's `> 0` guard fires the MaxTurns-ceiling verdict at turn 30 exactly as today (zero behavior delta).

**Test shape:** Go unit test — arm without `--max-turns`, assert `Ceiling.MaxTurns == 30`; simulate 30 turns with a never-satisfied condition, assert the MaxTurns-ceiling verdict fires at turn 30.

#### AC-INFINITE-GOAL-003 — Block-cap doctrine surfaces the silent cap (REQ-2)

**Given** the doctrine surface (`.claude/rules/moai/workflow/goal-directive.md` and/or `.claude/skills/moai/workflows/goal.md`).

**When** a user reads the goal docs to understand why their `--max-turns 0` goal stopped at turn 8.

**Then** the docs name `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` explicitly, identify it as the silent terminator that pre-empts `MaxTurns` (effective bound `min(MaxTurns, cap)` today), and recommend a raised value (e.g. 200) when arming an infinite goal.

**Test shape:** grep test — `grep -rn "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP" .claude/rules/moai/workflow/goal-directive.md .claude/skills/moai/workflows/goal.md` returns at least one match per file (or per the SSOT file) naming the cap and its bound semantics.

#### AC-INFINITE-GOAL-004 — Armed goal suppresses statusline `/clear` directive (REQ-3)

**Given** a session with an armed goal (`.moai/state/goal/<session-id>.json` exists with `Status == "armed"`), AND a context-usage level that would normally trigger the soft `(⚠️/clear)` or hard `(🛑/clear!)` marker.

**When** `computeHandoffStage` renders the statusline.

**Then** the rendered statusline does NOT contain the `/clear` directive marker text, AND the auto-compact pressure-relief handles context reclamation instead.

**Test shape:** statusline fixture — arm a goal, drive context-usage past the soft threshold, render the statusline, assert no `/clear` directive marker; clear the goal (Status != armed), re-render at the same usage, assert the marker returns.

#### AC-INFINITE-GOAL-005 — Infinite goal bounded by wall-clock (or cost) (REQ-4)

**Given** a goal armed with `--max-turns 0` AND `Ceiling.MaxDuration = <D seconds>` (per OQ-2 resolution; cost-cap alternative if OQ-2 picks cost as primary).

**When** the wall-clock time since arm time exceeds `D`.

**Then** the evaluator fires a 5-section verdict (Claim/Evidence/Baseline-attribution/Gaps/Residual-risk) indicating the wall-clock bound was hit, AND the goal loop halts, AND the verdict is indistinguishable in shape from a MaxTurns-ceiling verdict.

**Test shape:** Go unit test — arm with `--max-turns 0 --max-duration <short D>`, simulate time advancement past `D`, assert the wall-clock verdict fires; assert the MaxTurns ceiling did NOT fire (it is disabled at `0`).

#### AC-INFINITE-GOAL-006 — Stagnation guard fires on mechanical-condition no-change (REQ-4)

**Given** a goal armed with `--max-turns 0`, AND the goal's mechanical conditions target a bounded file set `F` AND a test count `T`, AND `DefaultStagnationThreshold` consecutive turns produce IDENTICAL mechanical-condition state (same test count, same pass/fail tally, same SHA of `F`).

**When** the Nth consecutive no-change turn's evaluator runs.

**Then** the stagnation guard fires a 5-section verdict indicating "N consecutive turns with no mechanical-condition change", AND the goal loop halts.

**Test shape:** Go unit test — arm with `--max-turns 0`, drive the simulation to produce identical mechanical-condition state for N consecutive turns (same test count, same file SHA), assert the stagnation verdict fires at turn N. Negative case: a single test-count change at turn N-1 resets the counter and the guard does NOT fire at turn N.

#### AC-INFINITE-GOAL-007 — Post-compact re-inject emits goal + SPEC + next action (REQ-5)

**Given** a session with an armed goal AND an active SPEC `progress.md` with a non-empty tail, AND a SessionStart(matcher: compact) hook registered.

**When** auto-compact fires and the SessionStart(compact) hook runs.

**Then** the hook's stdout contains: (a) the goal condition text, (b) the active SPEC-id, (c) the last-verified mechanical state (most recent failed-condition + observed output tail), AND (d) the single next action. The post-compact model receives this stdout as re-injection context.

**Test shape:** hook fixture — arm a goal with a known condition, write a `progress.md` tail, fire the compact event, capture the hook's stdout, assert all four elements are present. Negative case: no goal armed → hook is a no-op (empty stdout or absent).

#### AC-INFINITE-GOAL-008 — `/clear` re-arms embedded goal under new session-id (REQ-6)

**Given** a session with an armed goal (condition `C`, `Ceiling` with `MaxTurns=0`, `MaxDuration=D`), AND `moai handoff save` has run (embedding the goal into `.moai/state/handoff/pending.json`), AND a `/clear` fires with `mode=auto`.

**When** the post-`/clear` session starts and `handoff_inject.go` handles the `/clear` boundary.

**Then** a new goal state file is written under the NEW session-id, AND its condition is `C`, AND its `Ceiling` carries the embedded `MaxTurns=0` + `MaxDuration=D`, AND the new session's goal evaluator picks up the re-armed goal. Session-id keying is preserved (the new file is keyed by the new session-id, NOT by SPEC-id).

**Test shape:** hook fixture — arm a goal, run `moai handoff save`, simulate `/clear` (mode=auto) into a new session-id, assert a new goal file exists under the new session-id with the embedded condition + Ceiling. Negative case: no embedded goal in pending record → no goal file written (no spurious rearm).

#### AC-INFINITE-GOAL-009 — ac_converge "Max 20 turns" either parses or doc is corrected (REQ-7)

**Given** the OQ-1 resolution choosing option (a) [add `stop after N turns` regex to `parseCondition`] OR option (b) [correct the doc to state actual 30-turn execution].

**When** the chosen option is applied.

**Then** either (a) the `ac_converge` "Max 20 turns" clause is mechanically parsed by `parseCondition` and reflects into `Ceiling.MaxTurns == 20` (verified by arming `ac_converge` and asserting the resulting `Ceiling.MaxTurns`), OR (b) the `ac_converge` documentation states the actual 30-turn execution (verified by grep of `run.md` showing the corrected statement). Exactly one option is in force; both is contradictory.

**Test shape:** depends on OQ-1 — option (a): arm `ac_converge`, assert `Ceiling.MaxTurns == 20`; option (b): grep `run.md:153` for the corrected "30 turns" statement.

#### AC-INFINITE-GOAL-010 — Armed infinite goal does NOT weaken deny/ask rules (cross-cutting safety)

**Given** a goal armed with `--max-turns 0` (infinite), AND a tool call that matches a destructive-pattern deny rule (e.g. `git push origin main`, a secret, `rm -rf`, a deploy command).

**When** the PreToolUse gate evaluates the tool call during a goal-driven turn.

**Then** the gate DENIES the tool call exactly as it would without an armed goal — the deny decision is identical between an armed-infinite-goal session and an unarmed session.

**Test shape:** table-driven regression test — for each denylist fixture, exercise it under (1) no goal armed and (2) `--max-turns 0` goal armed; assert the gate decision is `deny` in both cells. Any cell flipping to allow is a hard FAIL. This AC cross-references `SPEC-STOPCHAIN-TRIM-001` REQ-007 / AC-006 (deny/ask invariance); the two ACs together cover the deny surface across the autonomy tier axis AND the goal-armed axis.

### §D.3 Indirect verification

- The `> 0` guard at `evaluate.go:142` is verified indirectly: AC-001 asserts the guard skips at `MaxTurns=0`; AC-002 asserts the guard fires at `MaxTurns=30`. The guard itself is NOT modified; only the propagated value changes. Any modification to the guard would surface as a regression in one of these two ACs.

### §D.4 Closure gates

- All ten MUST ACs green with attributed evidence (command + verbatim output).
- The infinite-loop safety pair (AC-005 wall-clock + AC-006 stagnation) is green — these are the load-bearing guards replacing the `MaxTurns=30` cap.
- The deny/ask regression guard (AC-010) is green under the armed-infinite-goal condition.
- Backward-compat sweep: unarmed session = zero behavior delta (AC-002 + AC-004 negative case).
- LSP gate: zero errors, zero type errors, lint clean.

### §D.5 Forward-looking checks (advisory, non-blocking for this SPEC)

- The HTML dashboard surface (deferred to `SPEC-GOAL-HTML-FLOW` P1) will consume the same `Ceiling` and verdict structure; the M4 bounds implementation should expose them in a way the P1 dashboard can render without re-derivation.
- The launcher env-inject path (REQ-2 OQ-3) — if deferred in M2, the follow-up SPEC should pick it up before the autonomy epic closes; otherwise users must manually raise the block-cap env, undermining the "arm and walk away" promise.

### §D.6 Definition of Done

This SPEC is DONE when:

1. All ten MUST ACs pass with attributed evidence.
- The infinite-loop safety pair (AC-005 + AC-006) is green — no infinite goal runs without a real bound.
- The deny/ask regression guard (AC-010) is green under the armed-infinite-goal condition.
- The backward-compat sweep is green (unarmed = zero delta).
- The project's standard quality gate is green.
- Frontmatter `status` transitions `draft → in-progress → implemented → completed` are owned by manager-develop and manager-docs; this plan-phase authoring only emits `draft`.
