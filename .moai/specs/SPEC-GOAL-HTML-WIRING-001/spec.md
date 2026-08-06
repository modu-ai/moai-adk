---
id: SPEC-GOAL-HTML-WIRING-001
title: "v3.2 production-caller wiring for 3 inert goal/plan HTML renderer surfaces"
version: "0.1.0"
status: in-progress
created: 2026-08-07
updated: 2026-08-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "goal, plan-html, wiring, dashboard, cli"
related_specs:
  - SPEC-GOAL-HTML-FLOW-001
  - SPEC-INFINITE-GOAL-001
depends_on:
  - SPEC-GOAL-HTML-FLOW-001
tier: M
---

# SPEC-GOAL-HTML-WIRING-001 — v3.2 production-caller wiring

## §A. Context

**Predecessor**: SPEC-GOAL-HTML-FLOW-001 (`status: completed`, 11/11 AC green) delivered three renderer surfaces — the goal dashboard (`RenderDashboard` / `RenderDashboardReArm`), the plan-HTML report (`RenderPlanHTML`), and the re-arm UI model (`ReArmContext`). That SPEC's AC scope verified the RENDERER FUNCTIONS in isolation; its `acceptance.md §D.5 "Forward-looking checks"` explicitly deferred production-caller wiring to v3.2.

This SPEC is the v3.2 follow-up. **SPEC-GOAL-HTML-FLOW-001 stays `completed` — this is a NEW SPEC, NOT an amendment** (no `amendment_of`; no transition of GOAL-HTML-FLOW-001 back to `in-progress`).

The three surfaces share a common defect today: each is fully built and unit-tested but carries **ZERO production callers**. The AUTONOMY-TIERS anti-pattern "AC pass ≠ feature works" bit the predecessor because its ACs verified renderer scope only; the production-caller gap was invisible. This SPEC's ACs are authored end-to-end (wiring → file-on-disk → DOM-verified) to close that gap.

### §A.1 The three inert surfaces (verified against live code)

| Surface | Site (file:line) | Current state | Wiring this SPEC delivers |
|---------|------------------|---------------|---------------------------|
| **1 — Verdict auto-fill (c1)** | `internal/cli/goal.go:444` `runGoalRender` calls `goal.RenderDashboard(g, nil)` | `Verdict` hardcoded `nil` → dashboard only shows the "no verdict yet" placeholder | Load the last-produced `*Verdict` for the session and pass `RenderDashboard(g, v)` so the 5 CeilingVerdict sections render |
| **2 — Plan-HTML report** | `internal/report/planhtml/renderer.go:54` `RenderPlanHTML(specDir, reviewFile string) ([]byte, error)` | 0 production callers; no Cobra wrapper. `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md:204` instructs the orchestrator (the LLM) to "Invoke the plan-HTML renderer: `RenderPlanHTML(...)`" — but an LLM cannot call a Go function, so this is a **DEAD instruction** | Add a `moai plan render-html <SPEC-ID>` Cobra subcommand; rewrite `spec-assembly.md` Step 2.3.3a to execute the CLI verb (replace dead LLM-instruction with an executable path) |
| **3 — Re-arm UI** | `internal/goal/dashboard.go:69` `ReArmContext` + `applyReArm` (`dashboard.go:111`) | Constructed **0 times in production** (tests only). `runGoalRender` routes through `RenderDashboardReArm(g, v, nil)` | Construct a `ReArmContext` from `.moai/state/handoff/pending.json` embedded goal + post-`/clear` new-session state, pass it to `RenderDashboardReArm(g, v, reArm)` so the 3 AC-GHF-007 UI states render |

### §A.2 Predecessor references

- **SPEC-GOAL-HTML-FLOW-001** — the renderer source. This SPEC consumes its signatures verbatim (`RenderDashboard` / `RenderDashboardReArm` / `RenderPlanHTML` / `ReArmContext` / `Verdict` / `CeilingVerdict`) and does NOT modify them.
- **SPEC-INFINITE-GOAL-001** — the re-arm mechanism source. The mechanical re-arm pipeline (save-embed + `rearmEmbeddedGoal` at `internal/hook/handoff_inject.go:208` + `IsUnbounded` D8 defense) ALREADY SHIPPED. Only the dashboard SURFACING is unwired.

### §A.3 Out-of-scope deferral (Surface 1 c2)

Surface 1 c2 — per-turn Stop-hook `.html` auto-refresh (the "LIVE board") — is **explicitly OUT OF SCOPE** for this SPEC and deferred to a separate follow-up SPEC. This SPEC wires only c1 (on-demand render with the real verdict). See § Out of Scope.

## §B. Goals

Wire the three inert renderer surfaces into production code paths so that:
- `moai goal render` displays the real 5-section CeilingVerdict when one has been produced for the session.
- `moai plan render-html <SPEC-ID>` is an executable Cobra subcommand that produces a `.html` report on disk.
- `moai goal render` surfaces the 3 re-arm UI states (re-arm indicator / re-armed-under-new-id / D8 unbounded-rejection banner) by constructing a `ReArmContext` from the already-landed mechanical re-arm state.

## §C. Non-Goals

- NOT building new renderer capabilities — all three renderers are complete and tested under SPEC-GOAL-HTML-FLOW-001. This SPEC is wiring-only.
- NOT modifying `RenderDashboard` / `RenderDashboardReArm` / `RenderPlanHTML` signatures (declared stable per GOAL-HTML-FLOW-001 §D.5).
- NOT touching the SPEC-INFINITE-GOAL-001 re-arm mechanism (save-embed + `rearmEmbeddedGoal` + D8 defense stay byte-identical).
- NOT regressing GOAL-HTML-FLOW-001's 11 ACs — they remain green.

## §D. Requirements (GEARS)

### Surface 1 — Verdict auto-fill (c1)

**REQ-WIRE-001** (Ubiquitous): The `runGoalRender` code path at `internal/cli/goal.go` shall pass a non-nil `*Verdict` to `goal.RenderDashboard(g, v)` when the session's goal state carries a last-produced Verdict, so the 5 `CeilingVerdict` sections (`Claim` / `Evidence` / `Baseline-attribution` / `Gaps` / `Residual-risk`) render instead of the "no verdict yet" placeholder.

**REQ-WIRE-002** (Event-detected): **When** the stop-goal evaluator (`moai hook stop-goal`) emits a Verdict carrying a `*CeilingVerdict`, the goal package shall persist that Verdict to a per-session sidecar file at `.moai/state/goal/<session-id>.verdict.json` so a subsequent `moai goal render` invocation can load it without re-running the evaluator.

**REQ-WIRE-003** (Ubiquitous): **While** no Verdict has been produced for the session (no `.verdict.json` sidecar exists, OR the sidecar fails to parse), the `runGoalRender` path shall degrade gracefully by passing `nil` to `RenderDashboard` so the "no verdict yet" placeholder path (AC-GHF-011) is preserved byte-identical.

### Surface 2 — Plan-HTML CLI verb

**REQ-WIRE-004** (Ubiquitous): The `internal/cli` package shall register a NEW `moai plan` Cobra parent command (distinct from the `/moai plan` slash-command skill, which routes through the skill system — NOT the CLI binary) with a `render-html` subcommand accepting exactly one positional argument `<SPEC-ID>`.

**REQ-WIRE-005** (Event-driven): **When** `moai plan render-html <SPEC-ID>` is invoked with a SPEC-ID whose `.moai/specs/<SPEC-ID>/` directory exists, the command shall call `planhtml.RenderPlanHTML(specDir, reviewFile)` with `specDir` resolved to the SPEC directory and `reviewFile` resolved to the most recent `.moai/reports/plan-audit/<SPEC-ID>-review-<N>.md`, AND write the rendered bytes to `.moai/reports/plan-html/<SPEC-ID>-plan.html` (creating the directory if absent).

**REQ-WIRE-006** (Event-detected): **When** the review file is absent or unparseable, `moai plan render-html` shall rely on `RenderPlanHTML`'s existing fail-open behavior (REQ-GHF-007) — the report is still written with the "audit verdict unavailable" placeholder, AND the command exits 0.

**REQ-WIRE-007** (Event-detected): **When** `moai plan render-html <SPEC-ID>` is invoked with a SPEC-ID whose `.moai/specs/<SPEC-ID>/` directory does NOT exist, the command shall exit non-zero AND emit a human-readable diagnostic to stderr (no `.html` file written).

### Surface 2 — Template rewrite

**REQ-WIRE-008** (Ubiquitous): `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` Step 2.3.3a shall be rewritten so the emission step instructs the orchestrator to execute `moai plan render-html <SPEC-ID>` (an executable CLI path) in place of the current dead instruction that references the Go function `RenderPlanHTML(...)` directly (an LLM cannot call a Go function).

### Surface 3 — Re-arm UI construction

**REQ-WIRE-009** (Event-driven): **When** `moai goal render` is invoked for a session whose `.moai/state/handoff/pending.json` carries an `EmbeddedGoal` (per the SPEC-INFINITE-GOAL-001 save-embed contract) AND/OR whose post-`/clear` new-session goal file exists, `runGoalRender` shall construct a `goal.ReArmContext` populated from that embedded + new-session state and pass it to `goal.RenderDashboardReArm(g, v, reArm)` so the 3 AC-GHF-007 UI states render.

**REQ-WIRE-010** (State-driven): **While** neither an `EmbeddedGoal` in `pending.json` NOR a post-`/clear` new-session goal file exists, `runGoalRender` shall construct a nil `*ReArmContext` and pass `RenderDashboardReArm(g, v, nil)` so the base view renders byte-identically (the re-arm path is purely additive per AC-GHF-007).

### Cross-cutting

**REQ-WIRE-011** (Capability gate): **Where** CLI code in `internal/cli/` is touched by this SPEC, the subagent boundary (C-HRA-008) shall hold — the new `plan` command source AND the modified `goal.go` source shall contain ZERO references to `AskUserQuestion` or `mcp__askuser`, verified by a Go test mirroring `internal/cli/web_test.go::TestWeb_NoAskUserQuestion`.

**REQ-WIRE-012** (Ubiquitous): The renderer signatures declared stable by SPEC-GOAL-HTML-FLOW-001 §D.5 — `RenderDashboard(g, v)`, `RenderDashboardReArm(g, v, reArm)`, `RenderPlanHTML(specDir, reviewFile)` — shall remain byte-identical in their parameter lists and return shapes.

**REQ-WIRE-013** (Event-detected): **When** a contributor edits `spec-assembly.md` in the template source and runs `make build`, the embedded template FS shall regenerate so the binary carries the rewritten Step 2.3.3a (Template-First cycle, CLAUDE.local.md §2); AND the rewritten content shall satisfy §25 template-neutrality (no internal SPEC IDs / REQ tokens / commit SHAs leak into the distributed template — `spec-assembly.md` references `RenderPlanHTML` only, which is a permitted renderer-cross-reference; the rewrite must not ADD new internal-SPEC-ID or REQ/AC tokens).

## §E. Constraints

- **PRESERVE** (do not modify signatures or behavior): `RenderDashboard` / `RenderDashboardReArm` / `RenderPlanHTML` function signatures; all GOAL-HTML-FLOW-001 artifacts; the SPEC-INFINITE-GOAL-001 re-arm mechanism (`rearmEmbeddedGoal`, `IsUnbounded`, `pending.json` shape); AC-GHF-011 "no verdict yet" placeholder path.
- **C-HRA-008**: CLI code (`internal/cli/`) MUST NOT call `AskUserQuestion`/`mcp__askuser` — every outcome routes via exit code + stderr.
- **Template-First** (CLAUDE.local.md §2 + §25): edit `spec-assembly.md` at the template source, run `make build`, preserve §25 neutrality.
- **PR-mandatory Route B** (`enforce_admins: true`): plan-phase commits use EXPLICIT pathspecs (never `git add -A`); re-run divergence check immediately before commit.

## §F. History

- 2026-08-07 — plan-phase artifacts authored (Tier M, 3 artifacts + progress.md skeleton). manager-spec. Intent 100% drained upstream (3 user-confirmed scope decisions: c1-only, new SPEC, CLI verb + closeout rewrite).

## §G. Out of Scope — explicit deferrals

### Out of Scope — Surface 1 c2 (per-turn LIVE board auto-refresh)

- The Stop-hook-driven `.html` auto-refresh on every turn-end (the "LIVE board" variant of Surface 1) is **deferred to a separate follow-up SPEC**. This SPEC wires only c1 (on-demand render with the real verdict). The c2 variant raises cost/complexity (per-turn disk write on the Stop hook critical path) and warrants its own REQ/AC set scoped to advisory-check discipline (CLAUDE.local.md §Advisory-Check Discipline).
- A bare `pending.json` schema change or new-session goal-file detection re-architecture is out of scope — REQ-WIRE-009 CONSUMES the existing SPEC-INFINITE-GOAL-001 shapes verbatim.

### Out of Scope — renderer capability additions

- New dashboard sections, new plan-HTML contract fields, new ReArmContext UI states — all out of scope. This SPEC surfaces EXISTING renderer output through production callers; it does NOT extend what the renderers render.

### Out of Scope — re-running the evaluator at render time

- REQ-WIRE-002 persists the last-produced Verdict via a sidecar file; recomputing the Verdict by re-running the evaluator inside `moai goal render` is explicitly rejected (the evaluator runs shell conditions that may be expensive or stateful — recompute is the wrong shape for a render command).

### Out of Scope — `moai plan` slash-command skill convergence

- The NEW `moai plan` Cobra CLI verb (this SPEC) and the EXISTING `/moai plan` slash-command skill are two separate invocation surfaces that share a name. Converging them (e.g., making `/moai plan` invoke the CLI verb) is out of scope — the slash-command skill routes through the Claude Code skill system, the CLI verb is a Go binary subcommand. The naming overlap is intentional and documented, not a defect.
