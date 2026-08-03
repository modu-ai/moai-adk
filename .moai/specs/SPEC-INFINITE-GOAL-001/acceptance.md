# acceptance.md — SPEC-INFINITE-GOAL-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-INFINITE-GOAL-NNN); this file does NOT restate them as requirements.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-INFINITE-GOAL-001 | REQ-001 | `--max-turns 0` disables ceiling check | MUST | plan M1 |
| AC-INFINITE-GOAL-002 | REQ-001 | `--max-turns` omitted = 30 (backward compat) | MUST | plan M1 |
| AC-INFINITE-GOAL-003 | REQ-002 | block-cap doctrine surfaces the silent cap (clause-specific) | MUST | plan M2 |
| AC-INFINITE-GOAL-004 | REQ-003 | armed goal → statusline `/clear` directive suppressed | MUST | plan M3 |
| AC-INFINITE-GOAL-005 | REQ-004 | `--max-turns 0` bounded by wall-clock (OQ-2 primary) | MUST | plan M4 |
| AC-INFINITE-GOAL-006 | REQ-004 | stagnation guard fires on N-turn mechanical no-change | MUST | plan M4 |
| AC-INFINITE-GOAL-007 | REQ-005 | post-compact re-inject emits goal + SPEC + next action | MUST | plan M5 |
| AC-INFINITE-GOAL-008 | REQ-006 | `/clear` re-arms embedded goal under new session-id | MUST | plan M6 |
| AC-INFINITE-GOAL-009 | REQ-007 | ac_converge "Max 20" — OQ-1 option (b) doc-only correction | MUST | plan M7 |
| AC-INFINITE-GOAL-010 | cross-ref (deny/ask surface — cross-cuts `SPEC-STOPCHAIN-TRIM-001` REQ-007/AC-006) | armed infinite goal does NOT weaken deny/ask | MUST | plan M8 |
| AC-INFINITE-GOAL-011 | REQ-004 | `--max-turns 0` with NEITHER `--max-duration` NOR `--cost-cap` rejected at arm time | MUST | plan M1+M4 |

### §D.1 Severity model

All eleven ACs are MUST-pass. AC-005 + AC-006 + AC-011 together are the load-bearing safety trio for an infinite goal: AC-011 rejects an unbounded arm at arm time (fail-closed), AC-005 bounds the accepted arm at run time (wall-clock), and AC-006 halts a non-stagnant-but-non-converging loop (mechanical stagnation). Without all three, `MaxTurns=0` would permit unbounded cost / unbounded looping. AC-010 is the cross-cutting safety regression guard (deny/ask invariance holds even when the goal is infinite). The other seven are P0 behavior / continuity ACs.

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

**Then** the docs name `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` explicitly AND carry the SPECIFIC clause this SPEC adds — "recommend a raised value when arming a `--max-turns 0` goal" — identifying the cap as the silent terminator that pre-empts `MaxTurns` (effective bound `min(MaxTurns, cap)` today).

**Test shape (D6 vacuous-pass fix):** the grep is CLAUSE-SPECIFIC, not merely cap-name-specific. The cap name alone (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`) is already named in `goal-directive.md:11` and `goal.md` today, so a name-only grep passes vacuously. The AC therefore targets the clause this SPEC adds. Test command:
`grep -rn "max-turns 0" .claude/rules/moai/workflow/goal-directive.md .claude/skills/moai/workflows/goal.md | grep -i "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP"`
returns at least one match — i.e. a single line that names BOTH the cap AND the `--max-turns 0` arming context, proving the raised-value recommendation is scoped to the infinite-goal case this SPEC delivers. (The launcher inject path is verified in M2 via a Go unit test on `buildEnvForLaunch`, not via this grep.)

#### AC-INFINITE-GOAL-004 — Armed goal suppresses statusline `/clear` directive (REQ-3)

**Given** a session with an armed goal (`.moai/state/goal/<session-id>.json` exists with `Status == "armed"`), AND a context-usage level that would normally trigger the soft `(⚠️/clear)` or hard `(🛑/clear!)` marker.

**When** the renderer path consuming `handoffGuideStage` (`internal/statusline/renderer.go:529`; consumer `switch` at `:269`) renders the statusline.

**Then** the rendered statusline does NOT contain the `/clear` directive marker text, AND the auto-compact pressure-relief handles context reclamation instead.

**Test shape:** statusline fixture — arm a goal, drive context-usage past the soft threshold, render the statusline, assert no `/clear` directive marker; clear the goal (Status != armed), re-render at the same usage, assert the marker returns.

#### AC-INFINITE-GOAL-005 — Infinite goal bounded by wall-clock (REQ-4; OQ-2 → wall-clock primary)

**Given** a goal armed with `--max-turns 0 --max-duration <D seconds>` (OQ-2 RESOLVED: wall-clock is the primary bound; `cost-cap` is a documented follow-up because `Eval`/`Ceiling` carry no invocation/token accounting today).

**When** the wall-clock time since `CreatedAt` (schema.go:94) exceeds `D`.

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

#### AC-INFINITE-GOAL-009 — ac_converge "Max 20 turns" doc corrected to actual 30-turn execution (REQ-7; OQ-1 → option (b))

**Given** OQ-1 RESOLVED to option (b) doc-only correction (option (a) — adding a `stop after N turns` regex — is NOT taken because it would silently change 30→20 for every user and contradict AC-002).

**When** the doc-only correction is applied to `.claude/skills/moai/workflows/run.md:153`.

**Then** the `ac_converge` documentation states the actual 30-turn execution: the goal runs at `DefaultMaxTurns=30` because the `trailingExitClause` regex at `internal/cli/goal.go:27` matches ONLY `exits <N>` and the "Max 20 turns" clause is not parseable. Exactly option (b) is in force; option (a) is explicitly NOT in force.

**Test shape (option (b) chosen):** grep `run.md:153` region for a corrected statement carrying BOTH "30" (the actual execution) AND a reference to the unparseable clause (e.g. "Max 20 turns" quoted as the literal that is not parsed). A bare "30" without the explanatory reference is insufficient — the AC proves the doc explains WHY the actual execution is 30, not merely asserts it.

#### AC-INFINITE-GOAL-010 — Armed infinite goal does NOT weaken deny/ask rules (cross-cutting safety)

**Given** a goal armed with `--max-turns 0` (infinite), AND a tool call that matches a destructive-pattern deny rule (e.g. `git push origin main`, a secret, `rm -rf`, a deploy command).

**When** the PreToolUse gate evaluates the tool call during a goal-driven turn.

**Then** the gate DENIES the tool call exactly as it would without an armed goal — the deny decision is identical between an armed-infinite-goal session and an unarmed session.

**Traceability (D5):** this AC is a CROSS-CUTTING regression guard for the deny/ask surface this SPEC preserves (it does not modify the deny/ask rules — it asserts they remain binding under an armed infinite goal). It cross-references `SPEC-STOPCHAIN-TRIM-001` REQ-007 / AC-006 (deny/ask invariance across the autonomy-tier axis); this AC-010 covers the same surface across the goal-armed axis. The two ACs together cover the deny surface across both axes.

**Fixture source (D5):** if `SPEC-STOPCHAIN-TRIM-001` ships its deny fixtures (the deny-rule corpus its AC-006 exercises), this SPEC REUSES them (no duplication); otherwise this SPEC ships a MINIMAL local deny fixture covering at least: `git push origin main`, `rm -rf`, a secret literal. The choice is decided at M8 by checking whether the sibling SPEC's deny fixtures exist on disk at run-phase start; the AC text is updated then to name the chosen source. Reuse is preferred to avoid fixture drift.

**Test shape:** table-driven regression test — for each denylist fixture (reused or minimal-local), exercise it under (1) no goal armed and (2) `--max-turns 0` goal armed; assert the gate decision is `deny` in both cells. Any cell flipping to allow is a hard FAIL.

#### AC-INFINITE-GOAL-011 — Arm-time reject for unbounded infinite arm (REQ-4 arm-time enforcement, D1)

**Given** the `/moai goal` arm verb (`internal/cli/goal.go` `runGoalArm`) invoked with `--max-turns 0` AND NEITHER `--max-duration` NOR `--cost-cap` supplied.

**When** the arm command runs.

**Then** the arm command REJECTS the invocation (fail-closed): non-zero exit code AND a stderr message naming the missing bound (e.g. "`--max-turns 0` requires at least one real bound: `--max-duration <seconds>` or `--cost-cap <N>`"). No goal state file is written. **While** `--max-turns 0` is supplied WITH at least one real bound, the arm succeeds (covered by AC-005).

**Test shape:** Go unit test — invoke `runGoalArm` with `--max-turns 0` alone, assert non-zero exit + stderr contains the missing-bound message + no goal file written. Negative case: invoke with `--max-turns 0 --max-duration 3600`, assert exit 0 + goal file written (this is the AC-005 positive path, cross-referenced).

### §D.3 Indirect verification

- The `> 0` guard at `evaluate.go:142` is verified indirectly: AC-001 asserts the guard skips at `MaxTurns=0`; AC-002 asserts the guard fires at `MaxTurns=30`. The guard itself is NOT modified; only the propagated value changes. Any modification to the guard would surface as a regression in one of these two ACs.

### §D.4 Closure gates

- All eleven MUST ACs green with attributed evidence (command + verbatim output).
- The infinite-loop safety trio (AC-005 wall-clock + AC-006 stagnation + AC-011 arm-time reject) is green — these are the load-bearing guards replacing the `MaxTurns=30` cap. AC-011 is fail-closed at arm time; AC-005 bounds the accepted arm at run time; AC-006 halts a non-stagnant-but-non-converging loop.
- The deny/ask regression guard (AC-010) is green under the armed-infinite-goal condition (reused or minimal-local deny fixtures per the D5 fixture-source decision).
- Backward-compat sweep: unarmed session = zero behavior delta (AC-002 + AC-004 negative case).
- LSP gate: zero errors, zero type errors, lint clean.

### §D.5 Forward-looking checks (advisory, non-blocking for this SPEC)

- The HTML dashboard surface (deferred to `SPEC-GOAL-HTML-FLOW` P1) will consume the same `Ceiling` and verdict structure; the M4 bounds implementation should expose them in a way the P1 dashboard can render without re-derivation.
- The cost-cap bound (REQ-4 OQ-2 follow-up): M4 implements wall-clock as primary. The cost-cap (max invocations / token-spend) is a documented FOLLOW-UP because `Eval`/`Ceiling` carry no invocation/token accounting today. A follow-up SPEC should add the accounting surface before the autonomy epic closes; until then, an infinite goal is bounded only by wall-clock + stagnation (AC-005 + AC-006), which is sufficient for the "arm and walk away" promise but not for cost-sensitive deployments.
- The launcher env-inject path (REQ-2 OQ-3) is NOT deferred — it is a one-line addition in M2 (per the OQ-3 resolution). This note is retained only to flag that the M2 inject covers the `--max-turns 0` case; a future SPEC MAY generalize it to non-infinite goals if warranted.

### §D.6 Definition of Done

This SPEC is DONE when:

1. All eleven MUST ACs pass with attributed evidence.
- The infinite-loop safety trio (AC-005 + AC-006 + AC-011) is green — no infinite goal runs without a real bound, and no unbounded arm is accepted.
- The deny/ask regression guard (AC-010) is green under the armed-infinite-goal condition.
- The backward-compat sweep is green (unarmed = zero delta).
- The project's standard quality gate is green.
- Frontmatter `status` transitions `draft → in-progress → implemented → completed` are owned by manager-develop and manager-docs; this plan-phase authoring only emits `draft`.
