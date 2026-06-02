---
spec_id: SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001
created: 2026-06-02
updated: 2026-06-03
---

# Progress — SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001

## §A.3 Version

- v0.1.0 (plan-phase, 2026-05-22)
- v0.1.1 (plan-audit iter-1 D1/D2/D3/D4 patch, 2026-06-02, commit 9902b1fbe)

## §D — Phase 0.5 Plan Audit Gate

- plan-auditor iter-1 verdict: **FAIL 0.71** (Tier M threshold 0.80, -0.09)
- MP-2 Testability FAIL: AC-WSE-007 default oracle drift
- Defects:
  - D1 (BLOCKING): AC-WSE-007 `Team.DefaultModel == "opus[1m]"` (production) → should be template SSOT `"sonnet"`. **RESOLVED** (9902b1fbe)
  - D2 (BLOCKING): AC-WSE-007 + plan.md M2 `Worktree.AutoCreate == true` → both template+production say `false`. **RESOLVED** (9902b1fbe)
  - D3 (SHOULD-FIX): `workflow.yaml.tmpl` referenced 4× but file does not exist (template is plain `workflow.yaml`). **RESOLVED** (.tmpl→.yaml, 0 matches)
  - D4 (SHOULD-FIX): stale line numbers (WorkflowConfig types.go:152-159→300-306; audit exception :28→27). **RESOLVED**
  - D5 (MINOR): Edge-WSE-004 dual behavior — left as-is (yaml:"-" tag yields deterministic "silently ignored", acceptable)
- post-amendment cross-check: 34 of 36 AC-WSE-007 assertions already matched template; only D1/D2 diverged
- plan-auditor iter-2 PASS 0.87 (2026-06-03)
- GATE-2: user-approved 2026-06-02 (Option A — amend-first then run, no re-audit)

## §E — Phase 0.95 Mode Selection

### Input parameters
- tier: M
- scope (file count): ~12 files — `internal/config/` (types.go, defaults.go, workflow_accessors.go NEW, types_test.go, defaults_test.go, workflow_nested_test.go NEW, audit_loader_completeness_test.go, audit_registry.go, loader.go) + `internal/cli/` (team_spawn.go, team_spawn_test.go, workflow_lint.go, workflow_lint_test.go)
- domain count: 2 (`internal/config` + `internal/cli` Go packages)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy; hard milestone dependency M1 struct → M2 defaults/accessor → M3 test → M4 ad-hoc parser migration → M5 audit cleanup)
- Agent Teams prereqs: not evaluated (domain count 2 < 3 threshold; coding-heavy resolves to Mode 5)

### Mode evaluation table
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | Multi-milestone struct refactor + ad-hoc parser migration + new test file |
| 2 background | no | Write/Edit-heavy; background subagents auto-deny Write per CONST-V3R2-020 |
| 3 agent-team | no | Domain count 2 (<3 threshold); coding-heavy; Agent Teams ceiling unjustified |
| 4 parallel | no | Coding-heavy; Finding A4 caveat; sequential milestone dependency |
| 5 sub-agent | YES | Coding-heavy, 2-package, hard milestone dependency (M1 → M2-M4 → M5 audit cleanup LAST) |

### Decision: sub-agent

### Justification
Coding-heavy 2-package (`internal/config` + `internal/cli`) brownfield refactor with a hard milestone dependency: the struct skeleton (M1) must land before defaults/accessor (M2), test fixture (M3), ad-hoc parser migration (M4), and the audit-registry cleanup (M5, gated last). Per Finding A4 and orchestration-mode-selection.md §B.2, Mode 5 (sequential sub-agent — manager-develop, cycle_type=tdd, Tier M Section A-E) is correct. plan-auditor FAIL 0.71 resolved via amendment (9902b1fbe), GATE-2 approved Option A (no re-audit).

## §Run-phase Evidence

Run-phase executed by manager-develop (cycle_type=tdd, Mode 5). Note: this run-phase
landed inside an L1 worktree; the worktree branch was rebased onto origin/main
(absorbing the disjoint parallel-session commit 822f2e550 — `update_clean_install.go`)
before any edits. Implementation tracked against a clean origin/main baseline.

### Milestone evidence

| Milestone | REQs | Deliverables | Verification |
|-----------|------|--------------|--------------|
| M1 — Struct skeleton + sub-structs | REQ-WSE-001/002/004 | `types.go`: WorkflowConfig nested (9 fields + 9 sub-structs); FLAT rename AutoClear→AutoClearLegacy, AutoSelection→AutoSelectionLegacy; PlanTokens/RunTokens/SyncTokens `// Deprecated:` + yaml:"-"; `workflowFileWrapper` added | `TestWorkflowConfigNestedFieldReachability` PASS (24 paths); `TestTeamConfigStructShape` PASS (no Patterns); `TestWorkflowConfigLegacyFieldsRenamed` PASS |
| M2 — Defaults + accessors | REQ-WSE-005/007 | `defaults.go`: NewDefaultWorkflowConfig nested defaults mirroring template SSOT (Team.DefaultModel="sonnet", Worktree.AutoCreate=false, 7-key RoleProfiles); `workflow_accessors.go` NEW (5 accessors); `loader.go`: loadWorkflowSection added to Load() chain | `TestNewDefaultWorkflowConfigNestedDefaults` PASS (36 assertions); `TestWorkflowAccessorMethods` PASS (5 sub-cases) |
| M3 — yaml.Unmarshal + nested test | REQ-WSE-003 | `workflow_nested_test.go` NEW: production fixture (20 assertions via Loader.Load) + Edge-WSE-002/003/004 | `TestWorkflowYAMLUnmarshalProductionFixture` PASS (AC-WSE-003 20 assertions); 3 Edge tests PASS |
| M4 — Ad-hoc parser migration | REQ-WSE-006 | `team_spawn.go`: LoadRoleProfiles migrated to typed loader (ad-hoc string-parser removed); `workflow_lint.go`: internal workflowConfig/workflowRoleProfile types removed, uses config.WorkflowConfig | `TestLoadRoleProfiles` PASS; `TestWorkflowLint_*` PASS; ad-hoc parser grep = 0 |
| M5 — Audit registry cleanup | REQ-WSE-008 | `audit_loader_completeness_test.go`: "workflow" removed from acknowledgedUnloadedSections; `audit_registry.go`: comment updated | `TestAuditLoaderCompleteness` PASS (workflow now counted as loaded); exception grep = 0 |

(Detailed E1-E7 self-verification matrix returned to orchestrator in the run-phase report.)

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha:

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: (this commit)
