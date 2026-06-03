# Progress — SPEC-WEB-CONSOLE-003

> S2a of the `web-console-v3` cohort. Tier M. FLAT/SHALLOW project-config parity (development_mode + git_convention.convention).

## Plan-phase signal

plan_complete_at: 2026-06-03
plan_status: audit-ready

## Tier & artifacts

- Tier: M (3 artifacts: spec.md + plan.md + acceptance.md)
- Justification: new project-config persistence path crossing internal/web → internal/config boundary; ~6-7 files; dual-editor (web + TUI) × validate/widget/persist/4-locale/read-on-render → 11 ACs.

## Scope decisions recorded

- Confirmed flat settings: `development_mode` ({ddd,tdd}, exported predicate `models.ValidDevelopmentModes()`), `git_convention.convention` ({auto,conventional-commits,angular,karma,custom}, `pkg/models` oneof SSOT; new exported `IsValidConvention` to be added in M1).
- NARROWED OUT: `llm.mode` (backend-switch toggle, only `""|glm`), `llm.default_model` (legacy enum-less string, no `validate` tag, no canonical enum to reuse). Recorded in spec.md §1.
- KEY design: project-config persistence via config-manager `LoadRaw`→`SetSection`→`Save` (new `app` seams), NOT `ProfilePreferences` (which has no slot for these). Bounded `SyncToProjectConfig`-pattern extension.

## REQ / AC counts

- REQs: 8 (REQ-WC3-001 .. REQ-WC3-008)
- ACs: 11 + closure (AC-WC3-001a/001b/002a/002b/003/004/005/006a/006b/007/008/009) + 6 edge cases (EC-1..EC-6)

## Next

- Phase 0.5 plan-auditor (Tier M PASS threshold 0.80) → GATE-2 → /moai run (cycle_type=tdd).

---

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: M
- scope (file count): ~7 (internal/config validation.go + internal/web app.go/validate.go/handlers.go/projectconfig.go/page.html.tmpl + internal/cli profile_setup.go/profile_setup_translations.go)
- domain count: 2 (Go source: internal/web + internal/cli + internal/config; SPEC artifacts)
- file language mix: 100% Go (+ one html/template + SPEC markdown)
- concurrency benefit: LOW (coding-heavy, sequential milestone dependency M1→M2→M3→M4→M5)
- Agent Teams prereqs status: not evaluated (single-agent run-phase delegation)

**Mode evaluation table**

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | not selected | multi-file semantic feature, not a typo |
| 2 background | not selected | write-heavy (Write/Edit), cannot background |
| 3 agent-team | not selected | single-agent delegation; not multi-domain research-heavy |
| 4 parallel | not selected | coding-heavy (Finding A4 caveat); milestones have sequential deps |
| 5 sub-agent | **selected** | coding-heavy run-phase, single manager-develop, sequential M1→M5 |
| 6 workflow | not selected | < 30 files, not a uniform mechanical transform |

**Decision: sub-agent**

**Justification**: Run-phase implementation of a coding-heavy Tier M feature with strictly sequential milestone dependencies (M1 predicate → M2 seams reuse it → M3 handler wires M2 → M4 reuses M2 write path → M5 verifies). Per Finding A4 (coding tasks involve fewer truly parallelizable subtasks than research), Mode 5 (single sequential sub-agent) is the correct default — orchestrator-decided.

**Phase 0.5 SKIP rationale**: plan-auditor verdict PASS 0.91 ≥ 0.90 AND no plan-PR commit landed since that verdict → Phase 0.5 re-execution skipped per spec-workflow.md Plan Audit Gate skip policy (CONST-V3R5-026). GATE-2 (plan-to-implement HUMAN GATE) was approved by the user independently of the 0.91 score per CLAUDE.local.md §19.1.

## §E.2 Run-phase Evidence

| AC | REQ | Status | Verification command | Actual output |
|----|-----|--------|----------------------|---------------|
| AC-WC3-001a | REQ-WC3-001 | (pending run completion) | `go test ./internal/web/ -run TestValidateProjectConfig` | — |
| AC-WC3-001b | REQ-WC3-001/005 | (pending) | web handler round-trip + quality.yaml re-read | — |
| AC-WC3-002a | REQ-WC3-002 | (pending) | validator + IsValidConvention | — |
| AC-WC3-002b | REQ-WC3-002/005 | (pending) | web handler round-trip + git-convention.yaml re-read | — |
| AC-WC3-003 | REQ-WC3-003 | (pending) | template render test (Project fieldset, no type=text) | — |
| AC-WC3-004 | REQ-WC3-004 | (pending) | handleIndex read seam pre-select + read-fail inline error | — |
| AC-WC3-005 | REQ-WC3-005/007 | (pending) | app seam grep + empty-keeps-existing + no-marshal grep | — |
| AC-WC3-006a | REQ-WC3-006 | (pending) | TestGetProfileText_AllLanguages + construction grep | — |
| AC-WC3-006b | REQ-WC3-006 | (pending) | TUI save → quality.yaml/git-convention.yaml NOT preferences.yaml | — |
| AC-WC3-007 | REQ-WC3-007 | (pending) | section-isolation test (workflow/harness/git-strategy/llm unchanged) | — |
| AC-WC3-008 | REQ-WC3-008 | (pending) | 001/002 suite + integration_test DO_NOT_TOUCH sentinels | — |
| AC-WC3-009 | all | (pending) | `go test ./internal/web/... ./internal/cli/... ./internal/config/...` | — |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: "<orchestrator-backfills>"
run_status: in-progress
ac_pass_count: 0
ac_fail_count: 0
status_transition: "draft → in-progress (M1 commit, manager-develop)"
m1_to_mN_commit_strategy: "M1 predicate + frontmatter transition; M2-M5 incremental commits per milestone"
```
