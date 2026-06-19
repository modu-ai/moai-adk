---
id: SPEC-V3R6-ORCH-IGGDA-001
title: "Intent-Gated Goal-Directed Autonomy (IGGDA) — 4-phase plan→run→sync pipeline with safe-condition kickoff reduction, bounded recursive self-diagnosis, and moai-aware Stop-hook goal driver"
version: "0.1.0"
status: completed
created: 2026-06-19
updated: 2026-06-20
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "autonomy, iggda, goal-directed, kickoff, recursive-self-diagnosis, stop-hook, frozen-amend"
era: V3R6
tier: L
depends_on: [SPEC-AUTONOMY-RUN-GOAL-001, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001]
---

# SPEC-V3R6-ORCH-IGGDA-001 — Intent-Gated Goal-Directed Autonomy (IGGDA)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-19 | manager-spec | Initial plan-phase authoring. Tier L (5-artifact). Sibling-extends SPEC-AUTONOMY-RUN-GOAL-001 (run-phase `ac_converge` autonomy) by chaining autonomy across plan→run→sync phases, adding a bounded recursive self-diagnosis loop on mechanical failure, and reducing the Implementation Kickoff Approval human gate to a safe-condition predicate (Path B). Amends the FROZEN Implementation Kickoff Approval invariant (`run.md:122,124`, CLAUDE.local.md §19.1 REQ-ATR-015) without removing the gate. |

---

## §A — Context and Motivation (WHY)

### §A.1 The vision

The user (project owner) wants `/moai`'s default plan→run→sync pipeline to run **autonomously with NO human intervention after intent collection** — self-auditing, recursive self-diagnosis on failure, a final sync audit that verifies SPEC compliance deterministically, and goal-directed so it keeps going until the objective is met (Claude Code `/goal` semantics, v2.1.139+, per `.claude/rules/moai/workflow/goal-directive.md`). Human involvement is **concentrated in Phase 0 (Socratic intent collection)**; the rest is autonomous.

This is the **Intent-Gated Goal-Directed Autonomy (IGGDA)** redesign. It is the next step beyond the run-phase autonomy delivered by `SPEC-AUTONOMY-RUN-GOAL-001` (which granted `/goal ac_converge` autonomy to Phase 2 only, while preserving the Implementation Kickoff Approval human gate verbatim). IGGDA extends autonomy across all 4 phases and introduces a **safe-condition predicate** that reduces (does NOT remove) the Implementation Kickoff Approval gate weight in safe domains.

### §A.2 The sibling lineage

`SPEC-AUTONOMY-RUN-GOAL-001` (status: completed) delivered run-phase autonomy (`run.md:116-171` Run-phase Autonomy section + Mode 6 in `orchestration-mode-selection.md`) under 6 HARD safety conditions (C1–C6), of which **C1** is the absolute invariant:

> **C1 — Implementation Kickoff Approval mandatory, score-independent**: Implementation Kickoff Approval is a mandatory `AskUserQuestion` human gate; plan-auditor PASS or score ≥ 0.90 never auto-bypasses it. Skip-eligibility applies ONLY to Phase 0.5 plan-auditor verdict re-execution, NOT to Implementation Kickoff Approval.

IGGDA is the **deliberate, bounded amendment of C1** under user-mandated Path B. It does NOT silently bypass the gate; it transforms the gate from a per-run-blocking `AskUserQuestion` into a **safe-condition predicate** that auto-proceeds ONLY when all four safe-conditions hold (see §C.1 REQ-IGGDA-004), and REMAINS a mandatory explicit `AskUserQuestion` gate whenever any safe-condition fails (Tier L, security/payment/critical keywords, or `--pr`-forcing destructive scope).

### §A.3 The FROZEN-invariant amendment (high-stakes — flag honestly)

The Implementation Kickoff Approval gate is a `[ZONE:Frozen] [HARD]` invariant (CLAUDE.local.md §19.1, `run.md:122,124`, `orchestration-mode-selection.md:14`). Amending a FROZEN invariant is the highest-stakes change in the MoAI rule system. design.md §F carries a dedicated FROZEN-invariant analysis section proving Path B preserves the protection the invariant provides (intent verification before code is written). The acceptance criteria (§acceptance.md) test the safe-condition predicate **exhaustively on both branches** (auto-proceed + explicit-gate).

This SPEC does NOT claim the amendment is risk-free; it claims the amendment is **narrow** (safe-condition predicate, dangerous domains never auto-proceeded) and **reversible** (the predicate is a single rule-file edit away from full restoration).

---

## §B — Scope (WHAT this SPEC delivers)

This SPEC delivers EXACTLY five deliverables (D1–D5). It is doctrine/rules + hook focused; Go-code volume is low (a Stop hook shell script + minor `internal/spec/audit.go` wiring for the deterministic final SPEC-compliance check, if needed at all — see plan.md M5).

- **D1 — IGGDA 4-phase pipeline definition** in `.claude/rules/moai/workflow/orchestration-mode-selection.md` (or a NEW sibling `iggda-pipeline.md` if the change is too large for the existing file). Phases: Phase 0 Intent (human) → Phase 1 Plan (autonomous) → Phase 2 Run (autonomous + recursive self-diagnosis) → Phase 3 Sync + final audit (autonomous). Auto-advance between phases is driven by D4's Stop hook.
- **D2 — Path B safe-condition predicate** in `.claude/rules/moai/workflow/orchestration-mode-selection.md` § Implementation Kickoff Approval + in `.claude/skills/moai/workflows/run.md:120-126`. Amends (does NOT remove) the Implementation Kickoff Approval gate. The predicate is the single decision point that determines auto-proceed vs explicit-gate. Documented in design.md §F.
- **D3 — Bounded recursive self-diagnosis loop** in `.claude/rules/moai/workflow/run.md` (NEW section) + `.claude/rules/moai/development/manager-develop-prompt-template.md` (reference injection). Mechanical failures (lint/type/import) → root-cause-analyze → patch → re-verify, **max 3 iterations** (per `ci-autofix-protocol.md` + runtime-recovery-doctrine.md §3 invariant 1 `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3`). Semantic failures (race/deadlock/panic/assertion) → MUST escalate to human, NEVER self-fixed (per `run.md:152` + runtime-recovery-doctrine.md §3). The loop complies with all 5 circuit-breaker invariants.
- **D4 — Moai-aware Stop hook goal driver** as `.claude/hooks/moai/iggda-phase-driver.sh` (NEW). A Stop hook that reads `.moai/specs/<SPEC-ID>/progress.md` + runs `moai spec audit` and deterministically judges each phase's safe-transition → emits a `/goal`-style auto-advance signal. The user does NOT write a `/goal` condition string; the goal derived from Socratic intent is auto-converted to the condition by this hook. Complies with the Recovery-Signal Carve-Out (runtime-recovery-doctrine.md §4) — on recovery turns the hook exits 0.
- **D5 — Independent-audit preservation regression** proving the independent auditors (plan-auditor, sync-auditor) REMAIN in separate contexts for bias prevention. "Self-audit" (D3's recursive loop) is first-party verification; the independent auditor is the final guarantee. Both run autonomously in IGGDA but are NOT collapsed into one (design.md §G).

### §B.1 Deferred (explicitly NOT in this SPEC)

- **EX-1**: `run.md:164` "autonomy-config follow-up" (version preflight + capability gate for `/goal` availability) — IGGDA's D4 Stop hook assumes `/goal` is available; the graceful-degradation contract (fallback to per-turn progression when `/goal` is unavailable) is inherited verbatim from `run.md:162-164`. Closing the version-preflight follow-up is a separate SPEC (`SPEC-V3R6-IGGDA-PREFLIGHT-001` candidate).
- **EX-2**: Agent Teams (Mode 3) integration with IGGDA — out of scope; IGGDA's auto-advance works in sub-agent (Mode 5) and workflow (Mode 6) modes; Agent Teams integration is a follow-up.
- **EX-3**: The FROZEN `[ZONE:Frozen]` tag on Implementation Kickoff Approval itself — this SPEC amends the **behavior** (safe-condition predicate) but does NOT remove the `[ZONE:Frozen]` marker from the rule files. The marker remains as a trail that the invariant was historically frozen and is now under controlled amendment.

---

## §C — GEARS Requirements

> Notation: GEARS (current). `<subject>` generalized. Each REQ traces to ≥1 AC in acceptance.md.

### §C.1 D2 — Path B safe-condition predicate (the FROZEN-invariant amendment)

- **REQ-IGGDA-001** (Ubiquitous): The Implementation Kickoff Approval gate shall remain a mandatory orchestrator-issued `AskUserQuestion` round, presented after Phase 0.5 (Plan Audit Gate) and before Phase 0.95 (Mode Selection), as established by `run.md:122` and CLAUDE.local.md §19.1. This SPEC does NOT remove the gate; it introduces a safe-condition predicate (REQ-IGGDA-004) that reduces the gate's per-run-blocking weight in safe domains.

- **REQ-IGGDA-002** (State-driven): **While** the Implementation Kickoff Approval `AskUserQuestion` round is presented to the user, the orchestrator shall drain ALL user preferences (Tier, mode preference, PR strategy, and any domain-specific preferences surfaced by the safe-condition predicate) at this gate, because Phase 1–3 autonomy agents (Workflow agents, `/goal`-turn agents, recursive-self-diagnosis sub-agents) cannot prompt the user mid-run (the asymmetric boundary in `agent-common-protocol.md` § User Interaction Boundary).

- **REQ-IGGDA-003** (Event-driven): **When** the Implementation Kickoff Approval `AskUserQuestion` round returns a user response that is NOT explicit approval (e.g., "further review" or "abort"), the orchestrator shall NOT advance to Phase 1 (Plan) regardless of the safe-condition predicate outcome. User intent always overrides the safe-condition auto-proceed.

- **REQ-IGGDA-004** (Capability gate — THE FROZEN-INARIANT AMENDMENT): **Where** the **safe-condition predicate** holds — ALL four of (a) intent clarity 100% (Socratic interview complete per CLAUDE.md §7 Rule 5 + askuser-protocol.md), (b) plan-auditor PASS verdict, (c) SPEC is Tier S or Tier M (NOT Tier L), AND (d) the SPEC contains NO security/payment/critical-domain keywords AND no `--pr`-forcing destructive scope — the Implementation Kickoff Approval `AskUserQuestion` gate shall be reduced to a **lightweight confirmation** that auto-proceeds to Phase 1 after a bounded timeout (the `(Recommended)` option auto-selected), EFFECTIVELY automatic but NOT silently bypassed (the `AskUserQuestion` round is still issued so the user retains veto authority).

- **REQ-IGGDA-005** (Event-detected, replaces legacy IF/THEN): **When** ANY of the four safe-conditions in REQ-IGGDA-004 FAILS — Tier L, OR security/payment/critical-domain keywords detected (the keyword list is enumerated in design.md §F.3), OR `--pr`-forcing destructive scope, OR Socratic intent clarity < 100%, OR plan-auditor verdict is FAIL/INCONCLUSIVE — the Implementation Kickoff Approval gate SHALL REMAIN a **mandatory explicit `AskUserQuestion`** that blocks until the user responds (no auto-proceed, no timeout). This is the dangerous-domain carve-out: high-stakes work never auto-proceeds.

- **REQ-IGGDA-006** (Ubiquitous): The safe-condition predicate shall be deterministic and auditable: the orchestrator shall record each of the four conditions (a)–(d) and the final auto-proceed/explicit-gate decision in `.moai/specs/<SPEC-ID>/progress.md` under a `## §E — IGGDA Kickoff Predicate` section before any Phase 1 work begins. This makes the FROZEN-invariant amendment post-hoc verifiable.

- **REQ-IGGDA-007** (Unwanted behavior): The orchestrator shall NOT treat a passing safe-condition predicate as authorization to skip Phase 0.5 (Plan Audit Gate). Phase 0.5 runs BEFORE the Implementation Kickoff Approval gate; REQ-IGGDA-004(b) requires Phase 0.5 PASS as a precondition. The predicate and Phase 0.5 are sequential, not interchangeable.

### §C.2 D1 + D4 — IGGDA 4-phase pipeline + Stop hook driver

- **REQ-IGGDA-008** (Ubiquitous): The IGGDA pipeline shall define exactly four phases in order: Phase 0 Intent (human-in-the-loop, Socratic), Phase 1 Plan (autonomous, manager-spec + plan-auditor), Phase 2 Run (autonomous, manager-develop + bounded recursive self-diagnosis), Phase 3 Sync + final audit (autonomous, manager-docs + sync-auditor + `moai spec audit` deterministic SPEC-compliance check).

- **REQ-IGGDA-009** (Event-driven): **When** a phase completes and its safe-transition predicate holds — Phase 1: plan-auditor PASS + Implementation Kickoff Approval cleared; Phase 2: all blocking ACs PASS + `go test ./...` exit 0 + no test file outside SPEC scope modified; Phase 3: sync-auditor 4-dimension score ≥ threshold AND `moai spec audit` reports 0 MUST-FIX drift for the SPEC — the moai-aware Stop hook driver (D4) shall emit a `/goal`-style auto-advance signal that transitions the pipeline to the next phase WITHOUT a per-phase human gate.

- **REQ-IGGDA-010** (State-driven): **While** the IGGDA pipeline is between phases (a safe-transition predicate is being evaluated), the Stop hook driver shall read `.moai/specs/<SPEC-ID>/progress.md` for the current phase's evidence markers (§E.2 run-phase evidence, §E.3 run-phase audit-ready signal, §E.4 sync-phase audit-ready signal) and shall invoke `moai spec audit --json --filter-spec=<SPEC-ID>` to obtain the deterministic SPEC-compliance verdict, NEVER inferring phase completion from frontmatter text alone (per `verification-claim-integrity.md` §1.1 surface 3 — defect/success claims require the dedicated tool's output).

- **REQ-IGGDA-011** (Event-detected): **When** the Stop hook driver detects a withheld-recoverable error signal (`prompt_too_long`, `max_output_tokens`, `media_size`, `compact-failure` — per runtime-recovery-doctrine.md §1) OR a recovery-turn `stopReason`, the driver SHALL exit 0 (allow the turn to end / the recovery to proceed) rather than exit 2 (block), per the Recovery-Signal Carve-Out (runtime-recovery-doctrine.md §4). The driver MUST NOT place recovery turns into the `error → stop-hook-blocks → retry → error` death-spiral loop.

- **REQ-IGGDA-012** (Capability gate): **Where** the `/goal` primitive is unavailable (runtime version < v2.1.139, OR `disableAllHooks` / `allowManagedHooksOnly` is set), the IGGDA pipeline SHALL degrade gracefully to per-turn progression (no auto-advance between phases, the orchestrator drives each phase manually) rather than failing. The graceful-degradation contract is inherited verbatim from `run.md:162-164`.

- **REQ-IGGDA-013** (Ubiquitous): The Stop hook driver shall NOT invoke `AskUserQuestion` directly (per the orchestrator-subagent boundary in `agent-common-protocol.md` § Hook Invocation Surface — hooks return exit codes + structured JSON; the orchestrator translates block signals into `AskUserQuestion` rounds). On a phase-transition block (exit 2), the driver's JSON output carries a `ledger_note` field with the human-readable block reason, and the orchestrator surfaces it to the user.

### §C.3 D3 — Bounded recursive self-diagnosis loop (runtime-recovery-doctrine compliance)

- **REQ-IGGDA-014** (Ubiquitous): The recursive self-diagnosis loop shall be **bounded** — maximum 3 iterations per mechanical-failure episode, after which the loop MUST halt and escalate to the orchestrator for an `AskUserQuestion` human decision (continue manual / revert + re-plan / abort). The max-3 bound aligns with `ci-autofix-protocol.md` and runtime-recovery-doctrine.md §3 invariant 1 (`MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3`).

- **REQ-IGGDA-015** (Capability gate): **Where** a failure is **mechanical** (lint rule violation, type error, build error, missing import, formatting drift), the recursive self-diagnosis sub-agent (`Agent(general-purpose)` with backend whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row #7) shall execute the three-step DIAGNOSE-PATCH-VERIFY pattern (read failing output → root-cause → minimal patch → re-verify) and, on VERIFY exit 0, advance; on VERIFY still failing, increment the iteration counter and repeat from DIAGNOSE.

- **REQ-IGGDA-016** (Event-detected — HARD): **When** a failure is **semantic** (data race, deadlock, panic, test assertion failure, concurrency hazard), the recursive self-diagnosis loop SHALL halt IMMEDIATELY (no patch attempt) and escalate to the orchestrator for an `AskUserQuestion` human decision. Semantic failures are NEVER auto-fixed by the recursive loop (per `run.md:152` semantic-failure-handling + runtime-recovery-doctrine.md §3 — these require human judgment).

- **REQ-IGGDA-017** (State-driven): **While** the recursive self-diagnosis loop is active, the orchestrator shall comply with all 5 circuit-breaker invariants in runtime-recovery-doctrine.md §3 — (1) max-3 same-rung failures → escalate rung; (2) `hasAttemptedReactiveCompact` no-self-loop (never re-attempt the same DIAGNOSE-PATCH-VERIFY within one turn); (3) compact-can-PTL last-resort escape (if the loop itself PTLs, fall to rung-4 abort + preserve, not another patch); (4) abort-closes-ledger (persist state to `progress.md` before session end); (5) narrative-consistency (5-Section Evidence-Bearing Report across the compact/recovery boundary per `verification-claim-integrity.md`).

- **REQ-IGGDA-018** (Ubiquitous): Every recursive self-diagnosis iteration SHALL be logged to `.moai/specs/<SPEC-ID>/progress.md` under a `## §E — Recursive Self-Diagnosis Log` section with: iteration number, failure classification (mechanical/semantic), root-cause summary, patch summary, VERIFY result, and (on halt) escalation reason. This makes the bounded loop post-hoc auditable.

- **REQ-IGGDA-019** (Unwanted behavior): The recursive self-diagnosis loop SHALL NOT modify `.env`, `.env.*`, credentials files, secrets, `scripts/ci-watch/run.sh`, or any Wave 2 infrastructure scripts (per CONST-V3R5-011 + CONST-V3R5-013, inherited from `manager-develop-prompt-template.md` autofix escalation contract). The loop's patch surface is the SPEC scope only.

### §C.4 D5 — Independent-audit preservation

- **REQ-IGGDA-020** (Ubiquitous): The independent auditors (plan-auditor, sync-auditor) SHALL remain in separate agent contexts from the implementation agents (manager-spec, manager-develop, manager-docs) for bias prevention. IGGDA's autonomous execution does NOT collapse the independent auditor into the implementer.

- **REQ-IGGDA-021** (Capability gate): **Where** the IGGDA pipeline reaches Phase 1 plan-audit or Phase 3 sync-audit, the independent auditor (plan-auditor / sync-auditor) SHALL be spawned via `Agent(subagent_type: ...)` in a fresh isolated context — NOT as a continuation of the implementer's turn. The auditor's verdict is the FINAL guarantee that the phase is safe to advance.

- **REQ-IGGDA-022** (Ubiquitous): "Self-audit" in IGGDA terminology means the recursive self-diagnosis loop (D3) performing first-party verification of mechanical failures during Phase 2. "Self-audit" does NOT replace the independent auditor — the independent auditor runs in Phase 1 (plan-auditor) and Phase 3 (sync-auditor), and the self-audit runs in Phase 2 (recursive loop). The two are complementary, not interchangeable.

- **REQ-IGGDA-023** (Event-detected): **When** the independent auditor's verdict is FAIL or INCONCLUSIVE, the IGGDA pipeline SHALL halt auto-advance and surface the verdict to the user via `AskUserQuestion` (orchestrator-mediated, never auditor-direct). A FAIL/INCONCLUSIVE verdict is a hard stop on autonomy, regardless of any prior phase's PASS.

### §C.5 Cross-cutting HARD constraints (inherited from SPEC-AUTONOMY-RUN-GOAL-001 C1–C6 + new IGGDA additions)

- **REQ-IGGDA-024** (Ubiquitous, inherited C2): All user preferences SHALL be collected by the orchestrator via `AskUserQuestion` in Phase 0 (Intent) BEFORE any Phase 1–3 autonomy launch. Autonomous agents (Workflow agents, `/goal`-turn agents, recursive-self-diagnosis sub-agents) cannot prompt the user mid-run.

- **REQ-IGGDA-025** (Ubiquitous, inherited C3): Only the orchestrator (main session) SHALL spawn agents. No autonomous agent spawns another autonomous agent (the flat hierarchy per Anthropic Finding A1 "Subagents cannot spawn other subagents"). The recursive self-diagnosis loop is spawned BY the orchestrator, not by a run-phase agent.

- **REQ-IGGDA-026** (Ubiquitous, inherited C4): Background agents (`Agent(run_in_background: true)`) SHALL NOT perform Write/Edit operations. Implementation writes are foreground. The recursive self-diagnosis sub-agent runs in FOREGROUND (it patches code; background write is prohibited).

- **REQ-IGGDA-027** (State-driven, new IGGDA addition): **While** the IGGDA pipeline is executing autonomously across phases, the orchestrator SHALL NOT treat autonomy as authorization to perform destructive operations (force-push, `rm -rf`, dropping tables, posting externally) or PR creation. These remain explicit gates surfaced separately after Phase 3 convergence, per `run.md:154-156` non-substitution clause (inherited).

- **REQ-IGGDA-028** (Ubiquitous, new IGGDA addition): The IGGDA pipeline's final SPEC-compliance check (`moai spec audit` deterministic drift detection) SHALL be the terminal gate of Phase 3. A SPEC is "IGGDA-complete" only when `moai spec audit --json --filter-spec=<SPEC-ID>` reports `drift_findings: []` (zero MUST-FIX) AND the sync-auditor 4-dimension score ≥ threshold AND `git status` is clean. Until then, the `/goal` does not clear.

---

## §D — Constraints

### §D.1 Template-First Rule (CLAUDE.local.md §2)

Every template change introduced by this SPEC goes into `internal/template/templates/` FIRST, then `make build` regenerates `internal/template/embedded.go`, then `moai update` syncs to the local project. The template files this SPEC will touch (run-phase, NOT in this plan-phase — only enumerated for the run-phase plan):

- `internal/template/templates/.claude/skills/moai/workflows/run.md` (D2 amendment + D3 NEW section)
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` (D1 pipeline + D2 predicate)
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` (D3 reference injection)
- `internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` (D4 NEW hook)
- (Possibly) `internal/template/templates/.claude/rules/moai/workflow/iggda-pipeline.md` (D1 NEW sibling if the change is too large for orchestration-mode-selection.md — design decision deferred to plan.md M1)

This plan-phase authors the SPEC artifacts ONLY. No template files are modified in this phase.

### §D.2 Neutrality CI guard

No internal SPEC IDs / commit SHAs / macOS-bias paths / `feedback_` / `CLAUDE.local` references leak into template content. The safe-condition keyword list (REQ-IGGDA-005) is generic prose (security / payment / critical-domain keywords), not moai-adk-internal identifiers.

### §D.3 runtime-recovery-doctrine compliance

The recursive self-diagnosis loop (D3) honors ALL 5 circuit-breaker invariants (runtime-recovery-doctrine.md §3). The Stop hook driver (D4) honors the Recovery-Signal Carve-Out (§4). No Go runtime creep — this SPEC is policy-layer + shell-hook + minor Go wiring only (AP-RR-001 compliance).

### §D.4 Independent auditor separation preserved

The independent auditors (plan-auditor, sync-auditor) remain in separate contexts. "Self-audit" (D3) is first-party; the independent auditor is the final guarantee (REQ-IGGDA-020 through REQ-IGGDA-023).

### §D.5 AskUserQuestion boundary preserved

All autonomous agents return blocker reports; the orchestrator runs `AskUserQuestion`. The Stop hook driver returns exit codes + JSON, never `AskUserQuestion` directly (REQ-IGGDA-013).

---

## §E — Assumptions

1. **`/goal` primitive is available** at the runtime version this SPEC targets (v3.0.0 → Claude Code v2.1.139+). Graceful degradation (REQ-IGGDA-012) covers the unavailable case.
2. **`moai spec audit --json --filter-spec=<SPEC-ID>`** returns structured drift findings consumable by the Stop hook driver. If the flag does not exist, M5 of plan.md adds it (minor Go wiring in `internal/spec/audit.go`).
3. **The safe-condition keyword list** (security/payment/critical) is enumerable at design time (design.md §F.3 carries the initial list; it is extensible via rule-file edit, not code change).
4. **The user (project owner) has explicitly chosen Path B** — this assumption is grounded in the orchestrator's pre-delegation analysis and the user's confirmation. This SPEC does NOT re-litigate Path A vs Path B; it encodes Path B.
5. **The FROZEN-invariant amendment is narrow and reversible** — the safe-condition predicate is a single rule-file edit away from full restoration of the original Implementation Kickoff Approval behavior. This is the reversibility guarantee.

---

## §F — Open Questions / Forward Gaps

1. **Q1 — Keyword-list maintenance**: who owns the safe-condition keyword list (security/payment/critical) over time? Proposed: the list lives in `.claude/rules/moai/workflow/orchestration-mode-selection.md` §F.3 and is maintained via SPEC amendment. Deferred to plan.md M2.
2. **Q2 — Tier L boundary**: is Tier L the correct cutoff for "dangerous domain", or should Tier M with security keywords also force explicit-gate? REQ-IGGDA-005 currently forces explicit-gate for BOTH Tier L AND any-tier + security keywords (the OR is deliberate). Confirm in plan-auditor review.
3. **Q3 — Auto-proceed timeout**: REQ-IGGDA-004 says the lightweight confirmation "auto-proceeds after a bounded timeout". What is the timeout value? Proposed: 30 seconds (long enough for a human to veto, short enough to not stall autonomy). Deferred to design.md §F.4.
4. **Q4 — `moai spec audit --filter-spec` flag existence**: does this flag exist today, or does M5 add it? Need to grep `internal/spec/audit.go` in research.md.

---

## §G — Out of Scope

### Out of Scope — Go runtime-layer hook evolution

This SPEC is policy-layer + shell-hook + minor Go wiring. It does NOT introduce `internal/recovery/` or any Go-side recursive-self-diagnosis engine. The recursive loop is agent-driven (`Agent(general-purpose)` DIAGNOSE-PATCH-VERIFY), not Go-runtime. Mechanical enforcement of the Recovery-Signal Carve-Out via `stopReason` parsing is deferred to a future runtime-layer SPEC (forward-link: `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` candidate), per AP-RR-001 + AP-RR-006.

- The current `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` hooks do NOT parse `stopReason`; this SPEC's D4 hook (`iggda-phase-driver.sh`) likewise does NOT parse `stopReason` in v0.1.0. Mechanical recovery-signal detection is out of scope.

### Out of Scope — Agent Teams (Mode 3) integration

IGGDA's auto-advance works in sub-agent (Mode 5) and workflow (Mode 6) modes. Agent Teams (Mode 3) integration — where teammates coordinate via shared TaskList across IGGDA phases — is a follow-up SPEC. This SPEC's acceptance criteria verify Mode 5 + Mode 6 only.

### Out of Scope — Version preflight + capability gate closure

`run.md:164` documents the graceful-degradation contract for when `/goal` is unavailable (version < v2.1.139, or hooks disabled). The actual version-preflight implementation (detecting the runtime version and emitting a structured signal) is deferred to a follow-up SPEC (`SPEC-V3R6-IGGDA-PREFLIGHT-001` candidate, EX-1). This SPEC inherits the degradation contract verbatim and assumes `/goal` is available.

### Out of Scope — `[ZONE:Frozen]` marker removal

This SPEC amends the **behavior** of Implementation Kickoff Approval (safe-condition predicate) but does NOT remove the `[ZONE:Frozen]` marker from `run.md:122`, `orchestration-mode-selection.md:14`, or CLAUDE.local.md §19.1. The marker remains as an audit trail that the invariant was historically frozen and is now under controlled amendment. Marker-removal is a separate governance decision, out of scope.

### Out of Scope — Cross-SPEC IGGDA rollout

This SPEC delivers IGGDA for a single SPEC's plan→run→sync pipeline. Applying IGGDA across multiple concurrent SPECs (Epic-level IGGDA) — where the Stop hook driver manages phase transitions for N SPECs in parallel — is a follow-up. This SPEC's acceptance criteria verify single-SPEC IGGDA only.

---

## §H — Cross-References

- `.claude/skills/moai/workflows/run.md:116-171` — existing `/goal ac_converge` run-phase autonomy (sibling; IGGDA extends it with cross-phase chaining + recursive loop).
- `.claude/skills/moai/workflows/run.md:122,124` — Implementation Kickoff Approval FROZEN invariant (the invariant this SPEC amends under Path B).
- `.claude/skills/moai/workflows/run.md:152` — semantic-failure escalation (inherited verbatim by REQ-IGGDA-016).
- `.claude/skills/moai/workflows/run.md:162-164` — graceful-degradation contract for `/goal` unavailability (inherited by REQ-IGGDA-012).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md:14` — Implementation Kickoff Approval mandatory-restoration (the FROZEN rule).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md §C.3` — Mode 6 (Workflow) capability gate (IGGDA Phase 2 heavy work).
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` semantics (IGGDA's D4 driver is the moai-native equivalent).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — ultracode session mode + Workflow primitive (Phase 2 heavy work).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md §2,§3,§4` — cheapest-first ladder + 5 circuit-breaker invariants + Recovery-Signal Carve-Out (IGGDA recursive loop + Stop hook driver MUST comply).
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume (IGGDA long-running autonomy crosses `/clear`).
- `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3` — defect/success claims require the dedicated tool's output (REQ-IGGDA-010 `moai spec audit` compliance).
- `.claude/rules/moai/core/agent-common-protocol.md § Hook Invocation Surface` — hooks return exit codes + JSON; orchestrator translates to `AskUserQuestion` (REQ-IGGDA-013).
- `.claude/rules/moai/development/manager-develop-prompt-template.md § cycle_type=autofix` — DIAGNOSE-PATCH-VERIFY pattern (inherited by REQ-IGGDA-015).
- `.moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/` — direct predecessor/sibling (run-phase autonomy; IGGDA extends to cross-phase).
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/` — workflow `agent()` purpose taxonomy (Phase 2 fan-out effort selection).
- `.moai/specs/SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001/` — orchestrator interrupt ledger closure (Stop hook driver ledger_note field lineage).
- `CLAUDE.local.md §19.1` — Implementation Kickoff Approval mandatory-restoration (REQ-ATR-015, the FROZEN invariant owner).

---

Version: 0.1.0 (plan-phase authoring)
Status: Active — plan-phase draft, awaits plan-auditor independent audit
