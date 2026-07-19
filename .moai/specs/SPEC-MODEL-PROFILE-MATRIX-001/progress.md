# Progress — SPEC-MODEL-PROFILE-MATRIX-001

- **plan_complete_at**: 2026-07-20
- **plan_status**: audit-ready

## §E.1 Plan-phase Audit-Ready Signal

**Plan iteration:** v0.1.1 — plan-audit iter-1 FAIL (0.69, Tier L threshold 0.85) revised for iter-2. Tier L 5-file set now complete: spec.md (40 REQ-MPM), plan.md (M1–M5, D1–D7, 2 DECISIONs), acceptance.md (25 AC-MPM), design.md (architecture §A–§G), research.md (verified investigation §A–§F). Status: draft.

**iter-1 defects addressed:**
- **D1 (effort-injection over-claim):** rewrote REQ-MPM-025/026/027 + AC-MPM-016/017 — the Agent tool accepts per-spawn `model` only, NOT per-spawn `effort` for named subagents; profile injects MODEL at spawn, effort = documented intent (frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt). Removed the "effort overrides frontmatter at spawn" claim. → DECISION-001.
- **D2 (ApplyTierProfile enumeration + web save-path):** corrected spec §A.4 + plan §B to enumerate all 4 production call sites (`initializer.go:195`, `update.go:486`, `update.go:1447`, `web/agentfm.go:108`); corrected the wrong "web is display-only" premise (the web settings-save path mutates frontmatter today via `applyPerfTierEdits`). Added REQ-MPM-040 + AC-MPM-025 for the web save-path mutation retirement.
- **D3 (Tier L artifacts):** authored design.md + research.md, grounded in the verified investigation (call-site grep, Agent-tool capability, current schema state, GLM overlay mechanics).
- **D4 (default profile discrepancy):** default profile = `medium` (confirmed); reconciled `DefaultModelPolicy = high (→max)` vs config/template `medium` in REQ-MPM-002. Both `[NEEDS CLARIFICATION]` markers resolved into DECISION-001/002 (plan.md §E, dated 2026-07-20) — no open clarification markers remain.

Investigation findings (no agentlint LR-12; statusline not a reader; plan_type UI-selector removed but web save-path still mutates; ApplyTierProfile 4 call sites; template source `model: inherit` vs mutated local `model: opus`; `moai model profile` accessor does not exist yet) are recorded in research.md and plan.md §B and shape scope.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
