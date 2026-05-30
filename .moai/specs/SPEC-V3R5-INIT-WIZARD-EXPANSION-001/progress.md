---
spec_id: SPEC-V3R5-INIT-WIZARD-EXPANSION-001
version: "0.1.0"
status: in-progress
---

# Progress — INIT Wizard Decision-Point Expansion

## §A — Run-Phase Status

| Milestone | Status | Notes |
|-----------|--------|-------|
| M1 — Wizard Structure Extension | in-progress | WizardResult fields + Question entries |
| M2 — Initializer yaml Write Paths | pending | quality/project/lsp/design/harness yaml writes |
| M3 — CLI Flag Integration | pending | --standard/--advanced + override flags |
| M4 — Advanced Gate + Phase 2 Scaffolding | pending | advanced_gate.go + stubs |
| M5 — Tests + Coverage | pending | expansion_test.go + initializer_expansion_test.go |

## §E — Phase 0.95 Mode Selection

**Decision**: sub-agent (Mode 5 — sequential sub-agent)

**Input parameters**:
- tier: M (3-artifact LEAN)
- scope: ~6 files, ~400 LOC
- domain count: 2 (internal/cli/wizard + internal/core/project)
- file language mix: 100% Go
- concurrency benefit: LOW — coding-heavy single-domain work
- Agent Teams prereqs: harness level = standard (not thorough) → Mode 3 not eligible

**Mode evaluation**:
| Mode | Decision | Rationale |
|------|----------|-----------|
| trivial | not selected | multi-file semantic change |
| background | not selected | requires Write/Edit |
| agent-team | not selected | harness not thorough + not multi-domain |
| parallel | not selected | coding-heavy (Finding A4 caveat) |
| **sub-agent** | **SELECTED** | coding-heavy single-domain Tier M work; sequential per milestone |

**Justification**: Per Anthropic Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), coding-heavy Tier M single-domain SPEC defaults to Mode 5 sequential sub-agent. The scope touches only wizard and initializer packages, making parallel execution counterproductive.

## §E.2 Run-phase Evidence

AC PASS/FAIL matrix populated after M5 completion.

| AC | Status | Verification | Actual Output |
|----|--------|-------------|---------------|
| AC-IWE-001 | PENDING | grep project_mode | — |
| AC-IWE-002 | PENDING | grep evaluator-profiles | — |
| AC-IWE-003 | PENDING | grep lsp_enabled | — |
| AC-IWE-004 | PENDING | grep enforce_quality | — |
| AC-IWE-005 | PENDING | grep design_enabled | — |
| AC-IWE-006 | PENDING | grep "standard" in init.go | — |
| AC-IWE-007 | PENDING | grep "advanced" in init.go | — |
| AC-IWE-008 | PENDING | grep enforce-quality flag | — |
| AC-IWE-009 | PENDING | go test -cover | — |
| AC-IWE-010 | DEFERRED | byte-identical diff | backward-compat via Condition func |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: ""
run_commit_sha: ""
run_status: in-progress
ac_pass_count: 0
ac_fail_count: 0
new_warnings_or_lints_introduced: false
cross_platform_build:
  linux: pending
  windows: pending
total_run_phase_files: 0
m1_to_mN_commit_strategy: one-commit-per-milestone
```
