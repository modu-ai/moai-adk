---
id: SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001
title: Implementation Plan — skills_audit_test.go solo run.md path update
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
phase: v3.0.0
module: internal/template
lifecycle: spec-anchored
tags: "template-mirror-drift, test-fix, workflow-split"
tier: S
---

# Implementation Plan — SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001

## 1. Approach Summary

Single test-file edit (`internal/template/skills_audit_test.go`) updating the `solo run.md` test entry to point at the post-SPEC-V3R4-WORKFLOW-SPLIT-001 sub-skill location (`workflows/run/phase-execution.md`). Zero production code changes. Zero template content changes. Verification via existing test suite re-execution.

## 2. Milestones

Tier S — single milestone.

### M1 — Test path update + verification

**Scope**: 2-4 line edits in `internal/template/skills_audit_test.go` (lines 36-48 region).

**Edit map** (conforming to spec.md §C decision rule — `filePath` is canonical, `name` is diagnostic):

| Location | Before | After |
|----------|--------|-------|
| `skills_audit_test.go:39` | `name:     "solo run.md — plan audit gate markers",` | `name:     "solo run/phase-execution.md — plan audit gate markers",` |
| `skills_audit_test.go:40` | `filePath: ".claude/skills/moai/workflows/run.md",` | `filePath: ".claude/skills/moai/workflows/run/phase-execution.md",` |

**Optional** (per REQ-SARM-007): update comment block at lines 37-38 to note the SPEC-V3R4-WORKFLOW-SPLIT-001 sub-skill migration. Decision deferred to manager-develop (aesthetic, not assertion-critical).

**Verification commands**:

```bash
# 1. Target sub-test PASS
go test ./internal/template/ -run "TestSkillsContainPlanAuditGateMarkers/solo" -v

# 2. Full test in file PASS (no regression in 3 other sub-tests + TestReportsDirGitkeepExists)
go test ./internal/template/ -run "TestSkillsContainPlanAuditGateMarkers|TestReportsDirGitkeepExists" -v

# 3. Full package PASS
go test ./internal/template/... -count=1

# 4. Baseline quality
go vet ./internal/template/...
golangci-lint run --timeout=2m ./internal/template/...
```

## 3. Technical Approach

### 3.1 Why path update, not test deletion or test split

- **Deletion rejected**: The test asserts a useful invariant — Phase 0.5 documentation MUST exist somewhere in the run-phase skill tree. Deleting loses this invariant.
- **Split rejected**: Adding a 5th sub-test for `phase-execution.md` while keeping the broken `run.md` entry would require deciding what to assert against the thin router. Per spec.md REQ-SARM-004, router content is NOT in scope for this SPEC.
- **Path update accepted**: Minimal-change principle. The invariant being asserted (Phase 0.5 documentation exists in the run-phase skill subsystem) is preserved; only the file location tracked.

### 3.2 Why update display name

Per spec.md REQ-SARM-005, the sub-test display name in `go test -v` output should be unambiguous. Current `solo run.md` would mislead future debuggers into reading the router instead of the sub-skill.

## 4. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Sub-skill `phase-execution.md` may itself drift in future SPEC splits | Out of scope for this SPEC. Future drift would surface via this same test failing — fix-forward pattern preserved. |
| `team/run.md` may have similar drift latent | Verified explicitly: `team/run.md` sub-test currently PASSES per problem reproduction. No latent drift in team variant. |
| Touching test file may trigger unrelated lint warning | Pre-edit baseline 0 findings confirmed by `golangci-lint run --timeout=2m ./internal/template/...`. Post-edit verification required. |
| Concurrent edits from parallel session (L9) | Pre-spawn fetch + post-edit `git rev-list --count --left-right origin/main...HEAD` = `0 0` check before commit. |

## 5. Dependencies

- **Predecessors**: SPEC-V3R6-IVB-001 (Sprint 2 P4.1, merged at `d3ed4727d`) — clean baseline.
- **No blocking dependencies**: Test failure is independent of any other in-flight SPEC.
- **No code dependencies**: Edit isolated to single test file.

## 6. Definition of Done (Plan-level)

- M1 complete with all 4 verification commands PASS
- 5/5 ACs verified (see acceptance.md)
- spec-lint clean
- Frontmatter canonical 12 fields valid
- progress.md updated with `audit-ready` signal (this plan-phase) + run/sync/mx phase rows ready for downstream agents
- No commit yet (plan-phase deliverable; orchestrator will spawn plan-auditor first, then commit + push after audit PASS)
