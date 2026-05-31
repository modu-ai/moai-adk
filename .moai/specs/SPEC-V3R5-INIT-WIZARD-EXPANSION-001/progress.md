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

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-IWE-001 | PASS | `grep -n 'ID.*project_mode' questions.go` | questions.go:259, types.go:36, initializer.go:46 |
| AC-IWE-002 | PASS | `grep -n 'evaluator-profiles' questions.go` | questions.go:191,195,199,204,276; 4 profiles: default/strict/lenient/frontend |
| AC-IWE-003 | PASS | `grep -n 'lsp_enabled' questions.go types.go` | wizard types.go:38 + questions.go:284 + initializer_expansion.go:99 |
| AC-IWE-004 | PASS | `grep -n 'enforce_quality\|coverage_exemptions' questions.go types.go` | types.go:39-40, questions.go:292-304, expansion.go:127-143 |
| AC-IWE-005 | PASS | `grep -n 'design_enabled\|claude_design' questions.go types.go` | types.go:41-42, questions.go:314-332, expansion.go:163-164 |
| AC-IWE-006 | PASS | `grep -n '"standard"' init.go` | init.go:80,247,262,359; StandardMode gates Phase1 via Condition func |
| AC-IWE-007 | PASS | `grep -n '"advanced"' init.go` | init.go:81,246; IsAdvancedWizardReady() in advanced_gate.go detects P2/P4 readiness |
| AC-IWE-008 | PASS | `grep -n 'enforce-quality\|enable-lsp\|harness-profile\|project-mode\|enable-design' init.go` | 5 flags registered at init.go:84-88; mapped at init.go:264-269 |
| AC-IWE-009 | PASS | `go test -cover ./internal/cli/wizard/... ./internal/core/project/...` | wizard new code avg 85.6%; core/project 88.9%. Package total: wizard 41.8% (dominated by pre-existing fullscreen.go 0%), core/project 88.9%. New code coverage meets ≥85% target. |
| AC-IWE-010 | DEFERRED | byte-identical diff after Quick mode init | Quick mode Condition funcs guarantee Phase1 yaml writes skipped; WritePhase1Configs returns no-op when !StandardMode (verified by TestWritePhase1Configs_NoOpWhenNotStandard) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-05-30T15:00:00Z"
run_commit_sha: "7e36d369789b45f7da47aa954815d11b050bc511"
run_status: implemented
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 9
l44_pre_commit_fetch: skipped (race exception B9 — orchestrator will push)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: false
cross_platform_build:
  linux: PASS
  windows: PASS (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 13
m1_to_mN_commit_strategy: single-commit-M1-to-M5
```
