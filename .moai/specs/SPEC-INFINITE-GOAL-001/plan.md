# plan.md — SPEC-INFINITE-GOAL-001

> Implementation plan. Order is decision-reversibility-first: the MaxTurns contract leads (it is the smallest and most central), then the runtime block-cap ( doctrine + optional inject), then the statusline `/clear` suppression, then the real bounds (wall-clock/cost + stagnation), then the compact/clear continuity hooks, then the ac_converge correction, then verification.

## §A. Context

This SPEC is a **redesign codification**, not greenfield. The design authority is `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2 (v2) + §1.2 + the C2 finding. The C2 finding is load-bearing: it establishes that the 1M number is the auto-compact window, NOT the loop-breaking cap, so this SPEC targets the THREE real caps (`MaxTurns`, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, `/clear` doctrine) and adds real bounds (wall-clock/cost + stagnation). The HTML-dashboard aspect of §3.2 is split to `SPEC-GOAL-HTML-FLOW` (P1).

Per the report's P0 decomposition (§3.5 closeout row "SPEC-INFINITE-GOAL v2"), this is the third and final P0; it proceeds after `SPEC-STOPCHAIN-TRIM-001`. It interacts with the tier-aware hooks (sibling SPEC) but does not redefine them.

## §B. Known Issues

- **K-1**: The `> 0` guard at `evaluate.go:142` already treats `MaxTurns == 0` as "ceiling disabled" — this is the C2 finding's key insight. The implementation MUST NOT change the guard; only the value propagated. Changing the guard would break the infinite entry point.
- **K-2**: `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` is a runtime env, not MoAI-owned. Documentation alone is honest but may leave users stranded; an MoAI-side inject (when a goal is armed) is convenient but adds an env-mutation path. OQ-3 resolves whether the launcher already has an env-inject surface.
- **K-3**: Suppressing the statusline `/clear` directive when a goal is armed changes user-visible behavior. Users who relied on the marker to know when to manually `/clear` would lose that signal. Mitigation: the stage classification is still computed (informational), only the directive text is suppressed; and the goal's own per-turn output becomes the replacement signal (deferred to `SPEC-GOAL-HTML-FLOW` P1 for the rich version).
- **K-4**: Wall-clock vs cost cap (OQ-2). Wall-clock needs only `time.Since(armTime)`; cost cap needs invocation/token accounting that may not yet exist in the evaluator. If the cost-accounting surface is missing, M4 implements wall-clock as primary and leaves cost-cap as a follow-up.
- **K-5**: The strengthened stagnation guard (mechanical-condition diff over test count / file hash) introduces a file-hash computation per turn. The hash subject set MUST be bounded — it is DERIVED from the mechanical conditions' commands at evaluation time (the condition commands' target paths), NOT a new `Ceiling` schema field, and NOT the whole repo. Derivation (option a) is preferred over a new schema field (option b) because it avoids adding a field to `Ceiling` (which today carries only `MaxTurns`, schema.go:36-38) and keeps the bounded set implicit in the conditions the user already authored.
- **K-6**: REQ-6's `/clear` rearm path adds a NEW goal-state write surface: `handoff_inject.go` `claimAndInject` (`:114-181`) today writes `AdditionalContext` text ONLY and archives `pending.json` to `consumed/`; it does NOT touch goal state. REQ-6 therefore writes a new goal file under the new session-id as a NET-NEW write path (not an additive extension of the `AdditionalContext` write — the two are separate file artifacts). Session-id keying (REQ-GLE-004) is established by this new write; Option B (SPEC-id keying) is explicitly rejected (multi-session race).
- **K-7**: The `ac_converge` correction (REQ-7) is either a behavior change (option a — goal actually stops at 20) or a doc change (option b — actual 30 stated). Option (a) is a regression-risk for users who tuned their workflow around the actual 30-turn behavior. OQ-1 decides.

## §C. Pre-flight (read-only reconnaissance — before M1)

1. Read `internal/goal/schema.go:78,105` + `evaluate.go:84,106-121,142,153-161` — confirm the `> 0` guard + stagnation guard structure.
2. Read `internal/cli/goal.go` arm verb + `:27` `trailingExitClause` / `parseCondition` — confirm regex scope matches ONLY `exits <N>` (OQ-1 resolved: option (b) doc-only; no `parseCondition` change in M1).
3. Read `internal/statusline/renderer.go:529` `handoffGuideStage` (constants `:508-510`, wrapper `shouldShowHandoffGuide` `:590`, consumer `switch` at `:269`) — confirm where the goal-armed suppression clause slots in.
4. Read `internal/hook/handoff_inject.go` `claimAndInject` (`:114-181`) — confirm the `/clear` path writes `AdditionalContext` text only + archives to `consumed/`; the REQ-6 goal-file write is a NEW surface (D3).
5. Read `moai handoff save` implementation — confirm the goal-read call site.
6. Grep `moai cc` / `moai cg` launcher for `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (OQ-3 input).
7. Read `.claude/skills/moai/workflows/run.md:153-166` `ac_converge` block + `goal-directive.md` (REQ-2 doctrine surface).

## §D. Constraints (recap from spec.md §D — binding on the plan)

1. auto-compact is the pressure-relief, not `/clear`.
2. auto-compact thrashing guard respected — each iteration's footprint small.
3. model conditions → mechanical conditions (test count / file hash / exit code).
4. 3 caps explicit (MaxTurns + block-cap + /clear doctrine); 1M number NOT modified.
5. Backward compat: `--max-turns` unspecified = 30; `/clear` suppression only when armed.
6. `tags` comma-quoted-string; self-lint before return.
7. deny/ask invariance cross-ref (sibling SPEC REQ-007).

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

- Go unit test: `--max-turns 0` arms a goal whose `Ceiling.MaxTurns == 0` AND the evaluator's `> 0` guard skips the ceiling check across >30 simulated turns (AC-001).
- Go unit test: `--max-turns` omitted produces `Ceiling.MaxTurns == 30` (AC-002).
- Grep test: the block-cap doctrine guide surfaces `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` by name (AC-003).
- Statusline fixture: armed goal → no `/clear` directive marker; unarmed → marker present (AC-004).
- Go test: `--max-turns 0` with a wall-clock bound fires a verdict at the bound (AC-005).
- Go test: stagnation guard fires after N turns of identical test-count + file-hash (AC-006).
- Go test: `--max-turns 0` with NEITHER `--max-duration` NOR `--cost-cap` is rejected at arm time (AC-011).
- Hook fixture: SessionStart(compact) emits goal-condition + SPEC-id + next-action stdout when armed; no-op when not (AC-007).
- Hook fixture: `/clear` (mode=auto) with an embedded goal writes a new goal file under the new session-id (AC-008).
- REQ-7 OQ-1 resolution fixture (AC-009): option (b) chosen — grep `run.md:153` for the corrected "30 turns" statement.
- Regression test: deny/ask rules still bind under an armed infinite goal (AC-010).
- Lint / build / test clean.

## §F. Milestones

### Milestone M1 — `--max-turns` arm flag (REQ-1 only; REQ-7 OQ-1 → option (b))

Highest reversibility: the MaxTurns contract is the SPEC's central change. OQ-1 is RESOLVED to option (b) doc-only, so M1 carries NO `parseCondition` change — the regex at `internal/cli/goal.go:27` `trailingExitClause` matches ONLY `exits <N>` and is left untouched; the doc-correct happens in M7.

**Files (expected):**
- `internal/cli/goal.go` — add `--max-turns N` flag to the arm verb (`runGoalArm`); propagate to `NewGoal`. `0` is a valid value (infinite). Default unchanged (30). Also add `--max-duration` / `--cost-cap` flags whose values propagate to the new `Ceiling` fields (M4); the arm-time reject (AC-011) is enforced here when `--max-turns 0` is supplied without either bound.
- `internal/goal/schema.go` — confirm `Ceiling.MaxTurns` accepts `0` (no schema change for the `> 0` guard at `evaluate.go:142`; the new `Ceiling.MaxDuration` / `Ceiling.CostCap` fields land in M4).
- Tests: AC-001 + AC-002 + AC-011 (arm-time reject).

**Exit:** `--max-turns 0` produces an infinite-armed goal; default unchanged; arm-time reject fires for an unbounded arm.

### Milestone M2 — Block-cap doctrine + launcher inject (REQ-2; OQ-3 → one-line addition)

Second reversibility tier. OQ-3 is RESOLVED: the launcher inject path is a ONE-LINE ADDITION (not a new contract, not deferred). `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` has 0 matches in `internal/` Go source today; the launcher already has `buildEnvForLaunch` (launcher.go:772) + `setGLMEnv` + `injectTmuxSessionEnv`.

**Files (expected):**
- `.claude/rules/moai/workflow/goal-directive.md` + `.claude/skills/moai/workflows/goal.md` — add an explicit subsection naming `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` as the silent terminator and recommending a raised value (e.g. 200) when arming a `--max-turns 0` goal. AC-003's grep targets the SPECIFIC clause this SPEC adds ("recommend a raised value when arming a `--max-turns 0` goal"), not merely the cap name.
- `internal/config/envkeys.go` — new const for `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`.
- `internal/cli/launcher.go` (goal-aware branch of `buildEnvForLaunch`) — one line: when a goal is armed at `--max-turns 0`, inject `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=<raised>` into the session env.
- Test: AC-003 (doctrine grep, clause-specific).

**Exit:** block-cap doctrine surfaces the silent cap; launcher inject path is a one-line addition.

### Milestone M3 — Statusline `/clear` suppression (REQ-3)

Third reversibility tier: the statusline change is mechanical but user-visible (K-3).

**Files (expected):**
- `internal/statusline/renderer.go` — the suppression clause slots into the path that consumes `handoffGuideStage`'s result (the `switch handoffGuideStage(data)` at `:269`, fed by `handoffGuideStage` at `:529` with constants `handoffStageNone/Soft/Hard` at `:508-510` and wrapper `shouldShowHandoffGuide` at `:590`). Add a "goal armed?" precondition: read `.moai/state/goal/<session-id>.json`, check `Status == "armed"`; if armed, suppress the `/clear` directive markers (the stage classification may still be computed for informational purposes).
- Test: AC-004.

**Exit:** armed goal suppresses `/clear` directive; unarmed goal retains marker.

### Milestone M4 — Real bounds (REQ-4: wall-clock primary + stagnation; cost-cap follow-up)

Depends on M1 (infinite turns needs a real bound). Highest internal complexity. OQ-2 is RESOLVED: wall-clock (`max-duration`) is the PRIMARY bound (it needs only `time.Since(armTime)`; `CreatedAt` is already persisted at schema.go:94); `cost cap` is a documented FOLLOW-UP because `Eval` and `Ceiling` carry NO invocation/token accounting today (only `MaxTurns`).

**Files (expected):**
- `internal/goal/schema.go` — add `Ceiling.MaxDuration` (wall-clock seconds) field. `Ceiling.CostCap` is left as a documented follow-up (accounting surface absent) — DO NOT add a `CostCap` field in M4.
- `internal/goal/evaluate.go` — check `time.Since(g.CreatedAt)` against `Ceiling.MaxDuration` each turn; fire the existing 5-section verdict when the wall-clock bound is hit.
- `internal/goal/evaluate.go:84,106-121,153-161` stagnation guard — replace "same note 3 times" with mechanical-condition diff: test count, test pass/fail tally, bounded file-set SHA. The bounded file set is DERIVED at evaluation time from the mechanical conditions' commands' target paths (option a derivation — NO new `Ceiling` field; K-5). This avoids a schema change and keeps the set implicit in the conditions the user authored.
- Tests: AC-005 (wall-clock) + AC-006 (stagnation) + AC-011 (arm-time reject, enforced in M1's arm verb but verified here end-to-end).

**Exit:** `--max-turns 0` goals bounded by wall-clock; stagnation guard mechanical; cost-cap is a documented follow-up.

### Milestone M5 — SessionStart(compact) re-inject (REQ-5)

**Files (expected):**
- `.claude/hooks/moai/handle-session-start-compact.sh` (or analogous wrapper) — register a SessionStart(matcher: compact) hook that reads the armed goal + the active SPEC's `progress.md` tail and emits the re-injection to stdout.
- `moai hook session-start-compact` (or extension of existing) — Go-side handler producing the stdout text.
- Test: AC-007.

**Exit:** post-compact turn sees the goal re-injection.

### Milestone M6 — `/clear` auto-rearm (REQ-6, Option A — net-new goal-state write surface)

**Files (expected):**
- `moai handoff save` — read any live armed goal, embed verbatim condition + arm-time `Ceiling` into the pending record at `.moai/state/handoff/pending.json`.
- `internal/hook/handoff_inject.go` `claimAndInject` (`:114-181`) — on `/clear` (`source=clear ∧ mode=auto`), when the pending record carries an embedded goal, write a NEW goal state file under the NEW session-id. This is a NET-NEW write surface (D3): `claimAndInject` today writes `AdditionalContext` text only and archives `pending.json` to `consumed/`; it does NOT touch goal state. The new goal-file write is a separate file artifact from the `AdditionalContext` write. Session-id keying is established by this new write; Option B (SPEC-id keying) is rejected (K-6).
- Tests: AC-008.

**Exit:** `/clear` mechanically re-arms the goal via a net-new goal-state write.

### Milestone M7 — ac_converge correction (REQ-7 option (b) doc-only — OQ-1 RESOLVED)

OQ-1 is RESOLVED to option (b) doc-only. This milestone updates `.claude/skills/moai/workflows/run.md:153` to state the actual 30-turn execution. (Option (a) — adding a `stop after N turns` regex to `parseCondition` — is NOT taken: it would silently change 30→20 for every user and contradict AC-002.)

**Files (expected):**
- `.claude/skills/moai/workflows/run.md:153-166` — option (b) doc-correction: state that the `ac_converge` goal runs at `MaxTurns=30` because the `trailingExitClause` regex at `internal/cli/goal.go:27` matches ONLY `exits <N>` and the "Max 20 turns" clause is not parseable.
- Test: AC-009 (option (b) chosen).

**Exit:** `ac_converge` documentation states the actual 30-turn execution.

### Milestone M8 — Verify

Lowest reversibility.

- Full `go test ./...` + race (`internal/goal/...`, `internal/statusline/...`, `internal/hook/...`).
- All AC matrix green with attributed evidence.
- Deny/ask regression test under armed infinite goal green (AC-010).
- Backward-compat sweep: unarmed session = zero behavior delta.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-1**: Modifying the `> 0` guard at `evaluate.go:142` to special-case `0` differently — the guard already handles `0` correctly; touching it breaks the infinite entry point (K-1).
- **AP-2**: Raising `Default1MContextTokens` instead of touching the 3 real caps — the constraint-4 hazard; users would believe the cap was raised while `MaxTurns=30` still terminates the loop.
- **AP-3**: Implementing Option B (SPEC-id keying) for `/clear` rearm — multi-session race (K-6); session-id keying is the only safe path.
- **AP-4**: Strengthening the stagnation guard with an UNBOUNDED file-hash subject set (whole repo) — per-turn hash tax defeats the SPEC's purpose (K-5); the subject set MUST be the goal's target files.
- **AP-5**: Suppressing the statusline `/clear` directive unconditionally (not just when armed) — breaks the existing `/clear` doctrine for every session, not just armed-goal sessions (K-3 / backward-compat).
- **AP-6**: Letting the `stop-goal` hook decide auto-compact (block/continue ⟶ force `/clear`) — constraint-1 violation; auto-compact is the runtime's pressure-relief, not the hook's.
- **AP-7**: Implementing the HTML dashboard surface here — split off to `SPEC-GOAL-HTML-FLOW` P1; this SPEC is the loop-continuation mechanics only.

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-INFINITE-GOAL-001/spec.md`.
- acceptance.md: `.moai/specs/SPEC-INFINITE-GOAL-001/acceptance.md`.
- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2 (v2) + §1.2 + C2 finding.
- Sibling P0s: `SPEC-AUDIT-SNAPSHOT-001`, `SPEC-STOPCHAIN-TRIM-001`.
- Sibling P1: `SPEC-GOAL-HTML-FLOW` (planned) — owns the HTML dashboard surface split out of §3.2.
- Integrity invariants: `verification-claim-integrity.md` §1.1; `SPEC-STOPCHAIN-TRIM-001` REQ-007 (deny/ask invariance — this SPEC's AC-010 cross-references).
- run.md ac_converge: `.claude/skills/moai/workflows/run.md:153-166`.
