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

## §F Phase 4 Mode Selection

- Inputs: tier=L, scope ~20-25 files (Go config/template/cli/web + template llm.yaml + rules docs), domains=2 (Go source + template/docs), language mix Go-dominant, concurrency benefit LOW (coding-heavy).
- Mode evaluation: trivial NO (semantic change) / background NO (write work) / agent-team RETIRED / parallel NO (coding-heavy, 2 domains) / workflow NO (not mechanical-uniform) / sub-agent YES.
- Decision: sub-agent (Mode 5, sequential manager-develop, cycle_type=tdd)
- Justification: coding-heavy multi-milestone implementation per Anthropic coding-task parallelism caveat; Plan Audit Gate skipped per 4-condition contract (PASS 0.93 ≥ 0.90, plan artifacts unchanged since c703a203c, verdict < 24h, verdict PASS). Implementation Kickoff Approval: user-pasted resume with explicit `실행: /moai run` directive.

## §E.2 Run-phase Evidence

Run-phase cycle_type=tdd, Mode 5 sequential. Milestones M1–M5, per-M Conventional Commits pushed to main (Hybrid Trunk).

### M1 — Config schema + Matrix A data model

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-001 (EffectiveProfile) | PASS | `go test ./internal/config -run TestEffectiveProfile` | ok (profile→alias→medium default) |
| AC-MPM-004/007 (agent_overrides validate) | PASS | `go test ./internal/config -run TestValidateAgentOverrides` | ok (non-catalog / out-of-enum model / effort each error) |
| AC-MPM-005 (Matrix A max fidelity) | PASS | `go test ./internal/template -run TestResolveAgentModelEffort_MatrixAFidelity` | ok (all 10 grouped agents = max column) |
| AC-MPM-006 (override precedence) | PASS | `go test ./internal/template -run TestResolveAgentModelEffort_OverridePrecedence` | ok |
| AC-MPM-007 (inherit) | PASS | `TestResolveAgentModelEffort_Inherit` | ok (Explore + user agent → inherit, hasGroup=false) |
| AC-MPM-008 (template+local profiles schema) | PASS | template + local llm.yaml carry profile+profiles+agent_overrides, no plan_type/claude_models | verified by Read |
| AC-MPM-011 (init no plan_type) | PASS | `go test ./internal/cli -run TestInitCmd_PlanTypeRetired` | ok (deployed llm.yaml has no plan_type) |
| AC-MPM-024 (no haiku) | PASS | `TestDefaultProfileMatrix_NoHaiku` | ok |

M1 build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Full `go test ./...` = 0 failures. golangci-lint baseline 0 issues (pre-M1).

New files: internal/config/profile.go, internal/config/profile_test.go, internal/template/profile_matrix.go, internal/template/profile_matrix_test.go. Edited: types.go (+Profile/Profiles/AgentOverrides), validation.go (+validateProfile/validateAgentOverrides), template+local llm.yaml, init_test.go (plan_type test inverted per AC-MPM-011).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
