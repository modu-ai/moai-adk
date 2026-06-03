# Design — SPEC-AUTONOMY-RUN-GOAL-001

> Design-time decisions for the three deliverables. This is the design source for the run-phase author; the canonical runtime SSOT remains `orchestration-mode-selection.md` (D1) and the run skill body (D2).

## §A — Design Goals and the GATE-2 Invariant Placement

The central invariant is **GATE-2 (plan→run human gate)**. The design places all autonomy mechanisms (Mode 6, `/goal`) strictly *downstream* of GATE-2 in the run-phase timeline:

```
/moai run SPEC-XXX
  │
  ├── Phase 0.5  Plan Audit Gate (plan-auditor verdict)
  │        verdict PASS / score (skip-eligibility is a Phase-0.5 concept ONLY)
  │
  ├── ★ GATE-2  plan→run HUMAN GATE  ──────────────────────────────┐
  │        AskUserQuestion: [run 진입 (권장) / 추가 검토 / 중단]    │  ← MANDATORY, score-independent
  │        (preferences collected here: Tier, mode pref, PR strategy)│
  │        user approval REQUIRED before anything below              │
  │        ───────────────────────────────────────────────────────┘
  │
  ├── Phase 0.95  Mode Selection (NOW with Mode 6 candidate)
  │        autonomous decision — Modes 1-6
  │        Mode 6 candidate ONLY if (GATE-2 passed ✓) AND (≥30 files mechanical parallel)
  │
  └── Run-phase autonomous region (entered ONLY after GATE-2 approval)
           ├── /goal ac_converge  (transcript-measurable, max 20 turns)
           ├── Mode 6 Workflow fan-out (mechanical migration only) — if selected
           └── /moai loop (deterministic diagnostics) — orthogonal
```

**Why this placement**: `/goal` removes per-turn STOP prompts and Mode 6 removes mid-run user input capability (Workflow agents cannot prompt). Both make mid-run user decisions impossible. Therefore GATE-2 — the one decision that MUST involve the user — is placed *before* both, and all preferences are drained at GATE-2. This is the literal realization of the strategy's "preferences collected before launch" pattern (§7.2) and the 6-condition C1+C2.

## §B — D1: Mode 6 (workflow) Entry-Condition Decision Table

Mode 6 is appended to the §A catalog of `orchestration-mode-selection.md`. The decision table the author adds:

| Signal | Mode 5 (sub-agent, default) | Mode 6 (workflow, new) |
|--------|-----------------------------|------------------------|
| Scope (file count) | any | **≥ ~30 files** (soft threshold; tie → Mode 5) |
| Transformation kind | any (incl. semantic/new code) | **mechanical only** (call-site rename, import-path bulk change, signature-stable edits) |
| Inter-file dependency | any | **none** (genuinely parallel) |
| Domain mix | coding-heavy / multi-domain → Mode 5 | mechanical-uniform (single transformation rule) |
| Concurrency benefit | LOW (Finding A4 caveat) | HIGH (independent per-file work) |
| GATE-2 status | passed | **passed (required precondition)** |
| Preferences | collected | **collected (required precondition)** |

### §B.1 Decision-tree insertion point

In `orchestration-mode-selection.md` §B, the Mode 6 branch is inserted **after** the Mode 4 (parallel) check and **before** the Mode 5 default fallback:

```
  ├── Is the task ≥ ~30 files AND mechanical (uniform transform rule)
  │   AND genuinely parallel (no inter-file dependency)
  │   AND GATE-2 already passed AND preferences collected
  │   AND Workflows available (not disabled, version ≥ v2.1.154)?
  │   ├── YES → Mode 6: WORKFLOW (orchestrator-launched fan-out, scaling not nesting)
  │   └── NO  → continue
  │
  └── Default → Mode 5: SUB-AGENT
```

### §B.2 Why a soft `~30` threshold

The `~30` is a soft boundary (tilde-prefixed) so the §B.2 tie-breaker ("default to the simpler mode") naturally resolves the boundary toward Mode 5. This avoids a hard cliff at exactly 30 files and keeps coding-heavy work — even large — on the safer sequential path unless the transformation is genuinely mechanical-uniform.

### §B.3 Mode 6 is scaling, not nesting (Finding A1)

The §C capability-gate text must state explicitly: Mode 6 Workflow is launched by the **orchestrator** (main session) as a scaling primitive. The Workflow script coordinates agents and keeps intermediate results in script variables; it does NOT represent a subagent spawning a subagent. This preserves the flat hierarchy (C3). The concurrency model (16-concurrent / 1000-total backstop) is cited from `dynamic-workflows.md` as the primitive's published cap — NOT as a MoAI-invented API.

### §B.4 Anti-patterns added to §E

- Mode 6 for coding-heavy / new-code work (violates Finding A4 — Mode 5 is correct).
- Mode 6 launch attempted before GATE-2 passes (violates C1).
- Asserting a typed named Workflow script API (`agent()`/`parallel()`/`pipeline()`/`phase()`) — the docs do not document one; describe coordinate-agents conceptually.
- Mode 6 selected without recording the GATE-2-passed + preferences-collected confirmation in `progress.md`.

### §B.5 Two-axis confusion warning (wiring author, D1)

"Mode 6 (workflow)" is appended to ONE specific list: the **Phase 0.95 execution-mode catalog** in `orchestration-mode-selection.md` (`trivial` / `background` / `agent-team` / `parallel` / `sub-agent` → + `workflow`). This is a DIFFERENT axis from the `run.md` `--mode` **dispatch axis** (`autopilot` / `loop` / `team` / `pipeline` / `background`) documented in `spec-workflow.md` § Subcommand Classification and the `run.md` Mode Dispatch table. Both happen to be ~5-item lists, which is the confusion trap.

- **Execution-mode catalog** (Phase 0.95, where D1 lands Mode 6): governs HOW the orchestrator spawns — concurrency + spawn surface (direct / background `Agent` / `TeamCreate` / parallel `Agent()` / sequential `Agent()` / **Workflow fan-out**).
- **`--mode` dispatch axis** (CLI flag, NOT touched by this SPEC): governs WHICH `/moai run` workflow variant runs — `autopilot` vs `loop` (Ralph) vs `team` vs rejected `pipeline`.

[HARD] The wiring author MUST add Mode 6 (`workflow`) ONLY to the execution-mode catalog. It is NOT a new `--mode` dispatch value; no `--mode workflow` flag is introduced; the `run.md` Mode Dispatch sentinel set (`MODE_UNKNOWN` / `MODE_TEAM_UNAVAILABLE` / `MODE_PIPELINE_ONLY_UTILITY`) is unchanged. `orchestration-mode-selection.md`'s own header already states the two axes "interact with — but are separate from" each other — D1 must preserve that separation.

## §C — D2: Run `/goal` `ac_converge` Wiring Point and Condition

### §C.1 Wiring location

A NEW **self-contained** `## Run-phase Autonomy (/goal ac_converge)` section is added to the run router file `.claude/skills/moai/workflows/run.md`. This section physically co-locates BOTH ordering markers in ONE file:
1. the GATE-2 `AskUserQuestion` ordering reference (the human gate that MUST precede autonomy), AND
2. the `/goal ac_converge` set.

The GATE-2 `AskUserQuestion` reference MUST appear textually before the `/goal` token (D3's ordering check, §D.1, depends on both markers existing in this one file).

**Empty-router fact (why a NEW section, not an edit)**: today `.claude/skills/moai/workflows/run.md` is a thin Phase-routing router with ZERO GATE-2 / `/goal` content — a tree-wide grep for `GATE-2` over the run skill tree (`run.md` + `run/*.md`) returns 0 matches; the actual Phase-0.5 gate + Decision-Point logic lives in the `run/phase-execution.md` sub-skill. So the markers D3 searches for do not exist anywhere yet. D2 is the deliverable that INTRODUCES both markers into `run.md`. The run-phase author ADDS this section; there is nothing pre-existing to edit, and the author must not be surprised by the absence of markers.

**EX-5 reconciliation**: EX-5 excludes *rewriting the legacy agent chain / Phase 0.5 gate in `phase-execution.md`* (which carries archived-agent drift + the stale 5-mode table + the legacy Decision-Point gate). Adding the NEW autonomy section to `run.md` (a DIFFERENT file — the router) is additive new content, NOT a modification of the legacy `phase-execution.md` gate, and is therefore permitted. Placing the autonomy section in `run.md` (rather than `phase-execution.md`) is exactly what keeps this SPEC clear of the EX-5 legacy-drift surface while still giving D3 a single self-contained file with both ordering markers.

### §C.2 The `ac_converge` condition (verbatim, from strategy §5.2)

```text
Every blocking acceptance criterion in
.moai/specs/SPEC-{ID}/acceptance.md has its PASS evidence surfaced in
the conversation (test output, build exit 0, or explicit AC-id: PASS
line); AND `go test ./...` exit 0 is surfaced; AND no test file outside
the SPEC scope was modified (surfaced via git status). Stop when all
hold. Max 20 turns.
On any semantic failure (data race, deadlock, panic, test assertion
failure), clear this goal and escalate via AskUserQuestion — do NOT
auto-fix semantic failures.
[PRECONDITION: GATE-2 user approval already obtained; this goal does
 NOT substitute for or bypass GATE-2.]
```

### §C.3 Transcript-measurability note (C5 / REQ-ARG-007)

The reference to `acceptance.md` in the condition is the *naming* of where the AC list lives — NOT an instruction for the `/goal` evaluator to read the file. The measured predicate is "PASS evidence **surfaced in the conversation**". The Haiku evaluator judges only what Claude has surfaced; the orchestrator is responsible for surfacing the per-AC PASS lines + `go test` exit + `git status` into the transcript. The wiring section must make this explicit so the condition is not mistaken for a file-path predicate (anti-pattern AP-2 in the strategy).

### §C.4 Non-substitution + escalation clauses

The section documents two HARD clauses:
- **Non-substitution** (REQ-ARG-010): the goal never authorizes bypassing GATE-2, creating a PR, or any destructive operation; those remain separate explicit gates surfaced after convergence.
- **Semantic-failure escalation** (REQ-ARG-009): data race / deadlock / panic / test assertion failure → clear goal + `AskUserQuestion`, never auto-fix (CONST-V3R5-010 alignment).

## §D — D3: GATE-2 Preservation Regression Guard Design

Two complementary checks. Both are lightweight (read the run skill body string; no run-phase orchestration simulation needed).

### §D.1 Check A — ordering invariant

A Go test in `internal/template/` (e.g., `TestGate2PreservedBeforeGoal`) reads the run skill body and asserts:
1. a GATE-2 `AskUserQuestion` human-gate marker exists, AND
2. its first occurrence's byte offset precedes the first `/goal` token (the `ac_converge` set) AND any Mode-6 launch reference.

This proves `/goal`/Mode-6 cannot cross the plan→run boundary ahead of the human gate (the body literally orders the gate first).

### §D.2 Check B — score-independence + cross-reference

The same test (or a sibling assertion) asserts the run body contains:
1. a statement that GATE-2 is emitted **regardless of plan-auditor score** (incl. ≥ 0.90 skip-eligible) and that skip-eligibility applies only to Phase 0.5 verdict re-execution, NOT GATE-2; AND
2. a cross-reference token `§19.1` and/or `REQ-ATR-015`.

### §D.3 Why a body-assertion rather than a runtime simulation

The orchestrator's GATE-2 behavior is doctrine, not Go code — there is no Go function whose unit test would observe the gate. The verifiable artifact is the skill body that instructs the orchestrator. Asserting the body's ordering + score-independence statement + doctrine cross-reference is the strongest mechanically-checkable guarantee at this layer. (A future PostToolUse/runtime hook enforcement is possible but is explicitly deferred — consistent with the schema's "forward-looking enforcement is optional defense-in-depth" stance.)

## §E — Template Mirror Design (CLAUDE.local.md §2 + §25)

- The Mode 6 doctrine and the run autonomy section are mirrored into `internal/template/templates/.claude/...` so downstream `moai init`/`moai update` projects receive the same catalog + wiring.
- The mirrored copy must be **internal-content-neutral** (CLAUDE.local.md §25): the template copy uses generic prose for the doctrine (Mode 6 entry conditions, `/goal` condition shape, GATE-2 ordering) but strips internal SPEC IDs / REQ tokens / audit citations. The `SPEC-{ID}` placeholder in the `ac_converge` condition is already generic (it is a runtime substitution token, not an internal SPEC ID), so it is acceptable in the template.
- The D3 regression test lives under `internal/template/` (Go test) — it is project-internal tooling, not a template artifact, so it is NOT mirrored.

## §F — Decision Log

| # | Decision | Rationale |
|---|----------|-----------|
| DD-1 | Add a NEW self-contained `## Run-phase Autonomy (/goal ac_converge)` section to `run.md` router (co-locating BOTH the GATE-2 AskUserQuestion ordering reference AND the `/goal` set), NOT editing `phase-execution.md` | `run.md` is an empty router today (0 GATE-2/`/goal` markers tree-wide), so D3's guard needs markers that D2 introduces. Co-locating both in one `run.md` section gives D3 a single self-contained file for the ordering check (AC-ARG-004/008a). `phase-execution.md` carries EX-5 legacy drift (archived agents, stale 5-mode table, legacy Phase 0.5 gate) and is excluded. Adding to `run.md` is additive — reconciles with EX-5 (which forbids only the legacy `phase-execution.md` rewrite). |
| DD-2 | Mode 6 added to `orchestration-mode-selection.md` (the canonical SSOT), not duplicated in the run skill | The rule file is the runtime SSOT for mode selection; the skill references it. Avoids dual-source drift. |
| DD-3 | Hard-code the `run` `ac_converge` condition inline rather than reference a registry | The condition registry is EX-3 (sibling SPEC). Inlining keeps this SPEC standalone (no cross-SPEC blocker). |
| DD-4 | D3 guard = body string assertion, not runtime simulation | GATE-2 is doctrine instructing the orchestrator; the checkable artifact is the body. |
| DD-5 | Soft `~30` files threshold | Avoids a hard cliff; the existing tie-breaker resolves the boundary toward the safer Mode 5. |
| DD-6 | `autonomy.enabled` default-off untouched | Research-preview safety; flipping defaults is EX-7 / the config SPEC's responsibility. |

## §G — Cross-References

- `.moai/docs/autonomous-workflow-strategy.md` §4.2-B (run group), §5.2 (`ac_converge`), §6.2 (run Workflow shape), §7.1–7.2 (invariants + preferences-before-launch), §8.1 (run autonomy profile).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A/§B/§C/§E (D1 target).
- `.claude/rules/moai/workflow/goal-directive.md` (D2 — `/goal` semantics).
- `.claude/rules/moai/workflow/dynamic-workflows.md` (Mode 6 primitive, 16/1000 cap, named-API prohibition).
- CLAUDE.local.md §19.1 (REQ-ATR-015), §2 (Template-First), §25 (internal-content isolation).
