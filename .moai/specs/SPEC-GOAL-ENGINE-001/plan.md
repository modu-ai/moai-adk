# plan.md — SPEC-GOAL-ENGINE-001

> Tier L. Go + docs. Epic AGENTIC-CORE, SPEC 2 of 3. depends_on
> SPEC-ANALYZE-FIRST-ROUTING-001. Shared findings:
> `../SPEC-ANALYZE-FIRST-ROUTING-001/research.md`.

## §A — Context

- **Work location**: repo root `/Users/goos/MoAI/moai-adk-go`.
- **New Go surfaces**: `internal/goal/` (state + schema + evaluator),
  `internal/cli` hook verb `stop-goal`, `settings.json.tmpl` Stop-hook entry.
- **New docs surfaces**: `.claude/skills/moai/workflows/goal.md`; `/moai` SKILL.md
  (P1 registration + Quick Reference); `.claude/hooks/moai/handle-stop-goal.sh`
  (or fold into existing stop handler).
- **Doctrine edits**: `goal-directive.md`, `native-invocation-model.md`,
  `session-handoff.md`.
- **D8 doctrine edits (progression-mode codification)**: `CLAUDE.md` (kickoff /
  approval-gates section — template-sensitive, mirror to
  `internal/template/templates/CLAUDE.md`, no SPEC IDs per §25);
  `.claude/skills/moai/workflows/run.md` (co-located with Run-phase Autonomy);
  `.claude/rules/moai/workflow/orchestration-mode-selection.md` (co-located with
  Implementation Kickoff Approval mandatory-restoration policy). Each pinned as
  a separate reachability AC (AC-GLE-031..033).
- **PRESERVE**: `run.md § Run-phase Autonomy (/goal ac_converge)`;
  `internal/config/agentic_loop_distinctness_test.go` (must stay green);
  `internal/loop` / `internal/ralph` / `internal/cli/loop.go` (untouched — §D.3);
  the existing Stop-hook registration(s) in `settings.json.tmpl` (ADD a new entry,
  do NOT replace).

## §A.5 PRESERVE / EXTEND list

| Path | Disposition |
|------|-------------|
| `internal/goal/**` | NEW |
| `internal/cli` (stop-goal verb) | EXTEND (new hook verb) |
| `.claude/hooks/moai/handle-stop-goal.sh` | NEW (settled — new wrapper, not folded) |
| `internal/template/templates/.claude/settings.json.tmpl` | EXTEND (add Stop entry) |
| `.claude/skills/moai/workflows/goal.md` | NEW |
| `.claude/skills/moai/SKILL.md` | EXTEND (P1 + Quick Reference) |
| `run.md § Run-phase Autonomy` | PRESERVE |
| `internal/config/agentic_loop_distinctness_test.go` | PRESERVE (stays green) |
| `internal/loop`, `internal/ralph`, `internal/cli/loop.go` | PRESERVE |
| `CLAUDE.md` (kickoff section) | EXTEND (D8 progression-mode axis + kickoff-mandatory-both-modes; mirror to template) |
| `.claude/skills/moai/workflows/run.md` | EXTEND (D8 progression-mode co-located with Run-phase Autonomy) |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | EXTEND (D8 progression-mode co-located with Kickoff mandatory-restoration) |

## §B — Technical Design (folded design.md — LEAN Tier L)

### B.1 Package layout (minimal)

```
internal/goal/
  schema.go        # Goal, Condition (mechanical|model), Ceiling, ProgressEntry
  state.go         # per-session load/save (atomic temp+rename); path builder
  prune.go         # session-start orphan pruning → consumed/
  evaluate.go      # 2-tier evaluator: Tier1 mechanical, Tier2 model-judgment gate
internal/cli/
  hook_stop_goal.go  # `moai hook stop-goal` verb: read stdin JSON, evaluate, emit
```

### B.2 State schema (JSON)

```json
{
  "session_id": "<uuid|writer_pid-fallback>",
  "goal": "<condition text>",
  "conditions": [
    {"type": "mechanical", "cmd": "go test ./...", "expect_exit": 0},
    {"type": "model", "claim": "all AC rows show PASS in the transcript"}
  ],
  "ceiling": {"max_turns": 30},
  "turns_used": 0,
  "progress": [{"turn": 1, "note": "..."}],
  "progression_mode": "autonomous",
  "created_at": "2026-07-12T00:00:00Z",
  "status": "armed|satisfied|ceiling-exit|cleared"
}
```

### B.3 2-tier evaluator control flow (Tier 1 mechanical → Tier 2 model)

1. Load `<session-id>.json`. If none/`cleared`/`satisfied` → exit 0, no block.
2. Increment `turns_used`. If `turns_used >= ceiling.max_turns` → emit 5-section
   verdict, set `status: ceiling-exit`, exit 0 no block (REQ-GLE-013).
3. Stagnation guard: if the last N `progress` entries show no advance → 5-section
   verdict with E1/E3 note, exit 0 no block (REQ-GLE-017).
4. Native-`/goal`-active detection → yield/pass-through (REQ-GLE-016).
5. Tier 1: run each `mechanical` `cmd`; compare exit to `expect_exit`. Any FAIL →
   exit 0 stdout `{"decision":"block","reason":"<cond + tail>"}` (REQ-GLE-010).
6. All mechanical PASS AND ≥1 `model` condition → Tier 2 model judgment
   (REQ-GLE-011). Not satisfied → block; satisfied → step 7.
7. All satisfied → no block; set `status: satisfied` (REQ-GLE-012).
8. **D8 semi-autonomous checkpoint branch**: if `progression_mode ==
   "semi-autonomous"` AND the goal is NOT satisfied AND ceiling NOT reached
   (i.e., the normal "block and continue" path would fire), emit the
   checkpoint-signal block JSON (REQ-GLE-028, §B.5) INSTEAD OF the plain
   `{"decision":"block","reason":"<cond + tail>"}` — the `mode` field
   distinguishes the two; the orchestrator reads `mode == "semi-autonomous"`
   and runs AskUserQuestion. The checkpoint block carries `failed_conditions`
   (the failed-condition + tail detail from Tier 1) so the orchestrator's
   confirm AskUserQuestion surfaces WHY the goal isn't satisfied (the generic
   `reason` label alone is insufficient); this reconciles REQ-GLE-010 ↔
   REQ-GLE-028 — the failed-condition+tail mandate is present in BOTH modes
   (autonomous via plain block `reason`, semi-autonomous via the checkpoint's
   `failed_conditions` array). In `autonomous` mode (default), step 8 is
   skipped — the normal block from step 5/6 fires and the turn continues
   without asking (existing D3 behavior, REQ-GLE-027).

### B.4 Hook wiring

- Wrapper `handle-stop-goal.sh` reads stdin JSON, calls `moai hook stop-goal`,
  passes stdout through. Registered as a SEPARATE `Stop` hook entry (Stop hooks
  compose — do NOT replace the existing entry; coordinates with HARNESS-EVOLVE).
- Timeout: goal eval runs mechanical `cmd`s which may exceed the MoAI 5s policy
  default → this hook entry carries a per-hook `timeout` override of **120000ms**
  (settled). goal.md documents that goal `cmd`s SHOULD be fast (prefer
  `go test -run <pattern>` over the full suite) since the eval runs at turn-end.

### B.5 Progression-mode design (D8 — Autonomous/Semi-autonomous)

**Schema extension** (extends B.2): goal state gains `progression_mode` (string,
`"autonomous"` | `"semi-autonomous"`, default `"autonomous"`). Set at goal-arm
time from the Implementation Kickoff Approval AskUserQuestion response.

**Kickoff-time mode selection** (REQ-GLE-026): the orchestrator's Implementation
Kickoff Approval AskUserQuestion gains a progression-mode question as a DISTINCT
axis from the approve/decline decision. The user approves the kickoff (required)
AND chooses the post-approval progression mode (autonomous default, or
semi-autonomous). The selected mode flows into the `moai goal "<condition>"`
arm call → written to `progression_mode` in state. This is NOT a gate bypass —
the gate still runs; the axis selects only what happens AFTER the gate passes.

**Autonomous mode** (REQ-GLE-027): the existing D3 behavior — the evaluator
blocks each turn until conditions hold or the ceiling is reached, with NO
per-turn user prompt. No NEW behavioral surface beyond the `progression_mode`
state field. The `reason` in the block JSON carries the failed-condition /
model-claim detail as usual (no `mode` field, or `mode: "autonomous"`).

**Semi-autonomous mode** (REQ-GLE-028) — checkpoint-signal JSON contract:

```json
{
  "decision": "block",
  "reason": "semi-autonomous checkpoint: orchestrator to confirm continuation (turn 3 of 30)",
  "mode": "semi-autonomous",
  "turn": 3,
  "ceiling": 30,
  "last_progress": "M2 evaluator Tier-1 gate implemented",
  "failed_conditions": [
    {"cmd": "go test ./internal/goal/...", "exit": 1, "tail": "FAIL: TestSemiAutonomousCheckpoint ... (output tail)"}
  ]
}
```

The `mode: "semi-autonomous"` field is the load-bearing signal: the orchestrator
distinguishes a checkpoint block (run AskUserQuestion) from a normal
condition-failed block (just continue). The `failed_conditions` array carries
the failed-condition + output-tail detail so the orchestrator's confirm
AskUserQuestion can surface WHY the goal isn't satisfied (the generic `reason`
label alone is insufficient for an informed continue/clear/switch decision).
When no mechanical condition is failing (e.g., the checkpoint fires because a
model condition is not yet satisfied), `failed_conditions` is empty `[]` or
absent. **REQ-GLE-010 ↔ REQ-GLE-028 reconciliation**: the failed-condition +
output-tail detail mandated by REQ-GLE-010 is present in BOTH modes — in
autonomous mode it rides the plain block `reason`; in semi-autonomous mode it
rides the checkpoint's structured `failed_conditions` array. The two REQs do
NOT conflict; the checkpoint does NOT drop the diagnostic (the v0.2.0 plan-
auditor D2-1 defect is resolved by this enrichment).

**Orchestrator-bridge worked flow** (the orchestrator-translation-responsibility
pattern from `agent-common-protocol.md` § Hook Invocation Surface — the Stop
hook CANNOT call AskUserQuestion per REQ-GLE-014, so the orchestrator bridges):

```
[turn N ends]
    ↓
stop-goal hook evaluates goal
    ↓
goal not satisfied, ceiling not reached, progression_mode == "semi-autonomous"
    ↓
hook emits exit-0 stdout: {"decision":"block","mode":"semi-autonomous","turn":N,...}
    ↓
Claude Code honors the block → next turn begins
    ↓
orchestrator reads the checkpoint JSON from the prior turn's block reason
    ↓
orchestrator runs AskUserQuestion (it CAN — it is the main session):
    ├─ "Continue to next step" → orchestrator proceeds; hook re-evaluates next turn
    ├─ "Clear goal" → orchestrator runs `moai goal clear`; loop ends
    └─ "Switch to autonomous" → orchestrator updates progression_mode; loop continues without further checkpoints
    ↓
user choice applied; turn N+1 proceeds accordingly
```

This is the SAME pattern already codified for `team-ac-verify.sh` and
`sync-phase-quality-gate.sh` (hooks emit structured JSON; the orchestrator
translates to AskUserQuestion). No NEW boundary-crossing mechanism is invented.

## §B — Known Issues (Section A-E, Tier L relevant categories)

- **B3/B11 — subagent/hook boundary (C-HRA-008)**: `stop-goal` and
  `internal/goal/` MUST NOT call `AskUserQuestion`/`mcp__askuser`. CI guard grep
  required (REQ-GLE-014).
- **B1 — cross-platform build tags**: any `syscall`/pid use in the `writer_pid`
  fallback (REQ-GLE-008) needs `//go:build` split; verify
  `GOOS=windows GOARCH=amd64 go build ./...`.
- **B4 — frontmatter schema**: this SPEC's frontmatter verified (12 fields +
  `era`/`tier`/`depends_on`).
- **B5 — CI 3-tier**: spec-lint, golangci-lint, per-OS test can fail separately.
- **B7 — capture path resolution**: `stop-goal` reads `$CLAUDE_PROJECT_DIR` for
  the `.moai/state/goal/` path; fall back to `os.Getwd()` guarded (avoid leaking
  the wrong working dir).
- **B8 — working-tree hygiene**: `.moai/state/goal/` is runtime-managed; tests use
  `t.TempDir()`, never the project `.moai/state/`.
- **B12 — sync CHANGELOG discipline** (manager-docs, later phase).

## §C — Pre-flight (run before editing)

```bash
git branch --show-current ; git rev-parse HEAD
go build ./... ; GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5
# Confirm the distinctness guard is green (must stay green)
go test ./internal/config/ -run TestAgentic 2>&1 | tail -5
# Confirm existing Stop-hook registration (add, not replace)
grep -n '"Stop"' internal/template/templates/.claude/settings.json.tmpl
# Confirm session fallback contract exists
grep -rn "session current\|writer_pid\|active-sessions" internal/ | head
# Confirm run.md autonomy section (PRESERVE target)
grep -n "Run-phase Autonomy" .claude/skills/moai/workflows/run.md
```

## §D — Constraints

- `internal/goal/` ≥ 85% coverage (REQ-GLE-024).
- No `AskUserQuestion` in hook/goal code (REQ-GLE-014, grep 0).
- Atomic state writes only (temp + rename) — REQ-GLE-006.
- Do NOT modify `run.md` autonomy section, `internal/ralph`, `internal/loop`,
  `agentic_loop_distinctness_test.go`.
- Add a NEW Stop-hook entry (compose), never replace the existing one.
- Template-First mirrors + §25 neutrality + `make build` (REQ-GLE-025).
- Never `--no-verify`/`--amend`/force-push (B9).

## §E — Self-Verification (plan-phase audit-ready)

Run-phase completion verified by `acceptance.md`. Plan-phase audit-ready recorded
in `progress.md` §E.1.

## §F — Milestones (priority-ordered)

- **M1 — `internal/goal/` schema + state (D2, D7)**: schema.go, state.go (atomic),
  prune.go, writer_pid fallback; TDD; ≥85% coverage. Priority High.
- **M2 — 2-tier evaluator (D3)**: evaluate.go (Tier 1 mechanical, Tier 2 gate,
  ceiling, stagnation, native-yield); TDD. Priority High.
- **M3 — `moai hook stop-goal` verb + wrapper + settings entry (D3, D7)**:
  hook_stop_goal.go; handle-stop-goal.sh; settings.json.tmpl Stop entry (compose);
  boundary grep guard. Priority High.
- **M4 — `goal.md` workflow + SKILL.md registration (D1)**: goal.md; P1 list;
  Quick Reference — BOTH registration surfaces. Priority High.
- **M5 — Safety + doctrine (D4, D5)**: safety wiring; goal-directive.md row;
  native-invocation-model.md Axis B update; session-handoff.md Block 5 note.
  Priority Medium.
- **M6 — Analyze-First integration + mirror + build (D6, D7)**: §2 stage ⑤
  reference; moai.md boundary note; distinctness guard re-verify; template mirrors;
  `make build`. Priority High (gate).
- **M7 — Progression-mode axis (D8)**: schema `progression_mode` field;
  evaluator checkpoint-signal branch (B.3 step 8, B.5); kickoff-time mode
  selection documented in `goal.md`; CLAUDE.md + `run.md` +
  `orchestration-mode-selection.md` codification (3 separate reachability ACs);
  safety invariant — kickoff mandatory in both modes (Go test); autonomous-mode
  Go test; semi-autonomous checkpoint Go test + orchestrator-bridge doc.
  Priority Medium.

Ordering: M1 → M2 → M3 → M4 → M5 → M6 → M7.

## §G — Anti-Patterns

- Single shared goal state file (REQ-GLE-004 forbids — race).
- Replacing (not composing) the Stop-hook entry (clobbers HARNESS-EVOLVE observer).
- Modifying the `run.md ac_converge` section (owned by AUTONOMY-RUN-GOAL-001).
- Arming a goal that pre-authorizes run-phase entry (REQ-GLE-015 forbids).
- Emitting `block` on exit 2 (Claude Code honors stdout JSON only on exit 0).
- Vacuous coverage claim — cite real `go test -cover` output.
- **D8 — Treating the progression-mode axis as a gate bypass** (REQ-GLE-026
  forbids): the axis is a post-approval progression CHOICE, NOT a relaxation of
  the Implementation Kickoff Approval gate. The gate stays mandatory in both
  modes.
- **D8 — Having the `stop-goal` hook call AskUserQuestion** (REQ-GLE-014
  forbids): the semi-autonomous checkpoint is a JSON signal the ORCHESTRATOR
  reads; the orchestrator runs the AskUserQuestion round, not the hook.
- **D8 — Making autonomous mode default-on without a kickoff gate pass**:
  `progression_mode` defaults to `autonomous` ONLY for an already-approved goal;
  it never pre-authorizes run-phase entry.
- **D8 — Compounding the 3 doc-surface greps into one `A|B|C ≥N` grep** (AC
  reachability lesson forbids): CLAUDE.md, run.md, orchestration-mode-selection.md
  are 3 SEPARATE ACs (AC-GLE-031..033), each with its own baseline-0 grep.

## §H — Cross-References

- Shared `research.md` §C.4 (Axis B), §C.5 (Stop hooks), §D.1 (AUTONOMY-RUN-GOAL),
  §D.2 (HARNESS-EVOLVE Stop-hook coordination).
- `goal-directive.md` (native `/goal` semantics, evaluator model, HUMAN-ONLY).
- `native-invocation-model.md` `:44`/`:62`/`:71-73` (Axis B).
- `agent-common-protocol.md` § Hook Invocation Surface (exit-code / stdout JSON
  semantics; hook boundary grep).
- `spec-frontmatter-schema.md` (frontmatter schema).

## § Dependencies (coordination detail)

- **depends_on SPEC-ANALYZE-FIRST-ROUTING-001**: REQ-GLE-021 references the §2
  pipeline stage ⑤ goal evaluator. Run-phase entry follows the Depends_on
  pre-flight (ANALYZE-FIRST must be `completed`, or `--ignore-deps` + logged
  rationale).
- **SPEC-HARNESS-EVOLVE-001 Stop-hook coordination**: both SPECs add a `Stop` hook
  entry to `settings.json.tmpl`. Stop hooks COMPOSE (multiple entries all run).
  Register `stop-goal` as an ADDITIONAL entry; do NOT overwrite the routing-ledger
  observer entry. If HARNESS-EVOLVE lands first, append after its entry; if this
  lands first, HARNESS-EVOLVE appends after this. Both wrappers read the same
  stdin JSON independently.

## § Deferred (NOT run-phase scope)

- **docs-site 4-locale** `/moai goal` documentation. Follow-up SPEC.
- **Go `moai loop` / goal-engine unification** (research.md §C.1). Follow-up SPEC.
- **`/moai goal` full condition-template registry** across subcommands (the
  AUTONOMY roadmap's `SPEC-AUTONOMY-GOAL-CONDITIONS`). This SPEC hard-codes only
  the generic mechanical+model condition shape.

## § Settled Decisions (iteration-2 — clarifications resolved via AskUserQuestion)

- **DECISION (Tier-2 model-condition evaluation mechanism)** — RESOLVED:
  **Option B — orchestrator self-eval**. Once all mechanical (Tier 1) conditions
  PASS, the Stop hook surfaces the model claim in the block `reason`; the
  orchestrator evaluates it against conversation-surfaced evidence (provider-agnostic,
  incl. GLM). `stop-goal` itself does NOT run a model call. REQ-GLE-011 is reworded
  from "the hook shall perform Tier-2 judgment" to "shall gate Tier-2 evaluation so
  it is reached only after all mechanical conditions pass, surfacing the model claim
  in the block reason for orchestrator evaluation" (resolves the auditor's D-minor
  contradiction).
- **DECISION (stop-goal wrapper placement)** — RESOLVED: **new
  `handle-stop-goal.sh` wrapper + new `moai hook stop-goal` Go verb** (clean
  composition with the HARNESS-EVOLVE observer; the existing `handle-stop.sh` stays
  single-purpose).
- **DECISION (goal-eval hook timeout)** — RESOLVED: **120000ms** (2 min). goal.md
  MUST carry a documented note that goal `cmd`s SHOULD be fast (prefer
  `go test -run <pattern>` over the full suite) because the eval runs at turn-end.
- **DECISION (native-/goal-active detection)** — RESOLVED: **degrade + document
  DEBT**. When the runtime does not expose a native-goal-active signal, `stop-goal`
  degrades to "always evaluate the MoAI goal". This is recorded as accepted **DEBT**:
  REQ-GLE-016's double-block-avoidance guarantee is INERT in the degrade path — in
  the rare concurrent-native-`/goal` case the turn may be evaluated by both the
  native evaluator and the MoAI `stop-goal` (possible double evaluation only; no
  correctness hazard, both only continue-or-stop the turn). AC-GLE-016 verifies the
  yield behavior when the signal IS exposed; the degrade path is the documented
  fallback.
- **DECISION (D8 progression-mode axis = kickoff-time choice, NOT gate bypass)**
  — RESOLVED (user mid-turn directive 2026-07-12): the autonomous-vs-semi-
  autonomous axis is a CHOICE offered AT the Implementation Kickoff Approval
  gate, NOT a bypass OF the gate. The gate remains mandatory in both modes (C1
  preserved, REQ-GLE-015 preserved). The axis selects ONLY post-approval
  progression behavior (continue autonomously vs. checkpoint-confirm each step).
  §D.4 is reconciled: autonomous MODE is opt-in per-goal at the gate, NOT a
  default-on switch.
- **DECISION (D8 semi-autonomous confirm via orchestrator bridge, NOT hook
  prompt)** — RESOLVED: the `stop-goal` hook CANNOT call AskUserQuestion
  (REQ-GLE-014 binding). The semi-autonomous per-turn confirm flow uses the
  orchestrator-bridge pattern (B.5): the hook emits a checkpoint-signal block
  JSON with `mode: "semi-autonomous"`; the orchestrator (which CAN call
  AskUserQuestion) reads the checkpoint and runs the confirm round (continue /
  clear / switch-to-autonomous). This reuses the existing
  `agent-common-protocol.md` § User Interaction Boundary / Hook Invocation
  Surface orchestrator-translation-responsibility pattern (same shape as
  `team-ac-verify.sh` and `sync-phase-quality-gate.sh`). No NEW boundary-crossing
  mechanism is invented.
