# SPEC-V3R2-RT-002 Progress Tracker

> Live progress and session-handoff state for **Permission Stack + Bubble Mode**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial progress tracker — plan documents written; ready for plan-auditor |

---

## Current Status

| Field | Value |
|-------|-------|
| Phase | `plan` |
| Status | `plan-complete-pending-audit` |
| Branch | `plan/SPEC-V3R2-RT-002` |
| Worktree | `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002` |
| Base | `origin/main` (`496595c3f`) |
| Plan-auditor | not yet run (PR will trigger) |
| Run-phase entry | pending plan-auditor PASS + RT-001 / RT-005 dependency status check |

---

## Plan Phase Deliverables (this session)

- [x] `spec.md` v0.1.0 (24 EARS REQs across 6 categories, 15 ACs, 6 risks, 9 dependencies, breaking=false additive)
- [x] `plan.md` v0.1.0 (5 milestones M1-M5, 30 file:line anchors, traceability matrix 27 REQ → 15 AC → 35 tasks, mx_plan with 7 MX tag insertions, plan-audit-ready checklist 15/15 PASS)
- [x] `research.md` v0.1.0 (12 sections, 48 file:line anchors, library evaluation, breaking-change analysis, dependency status, performance budget)
- [x] `acceptance.md` v0.1.0 (15 ACs in G/W/T format with happy-path + edge cases + test mapping, ~30 new test functions across 4 new files + 3 extended files)
- [x] `tasks.md` v0.1.0 (35 tasks T-RT002-01..35 across M1-M5, owner roles, dependencies, critical path graph)
- [x] `progress.md` v0.1.0 (this file)
- [x] `issue-body.md` (GitHub issue body for tracking)

---

## Run Phase Plan (next session)

Per `plan.md` §9 Implementation Order Summary:

1. **M1 (P0)**: ~10 failing tests + 1 audit lint (T-RT002-01..13). RED 모두 확인; 기존 testbase GREEN 유지.
2. **M2 (P0)**: PreAllowlist sync.Once + ValidateMode spawn-side wire (`internal/permission/spawn.go` 신규) + frontmatter strict lint + security.yaml 신규 키 + template mirror (T-RT002-14..18). AC-07, AC-09 GREEN.
3. **M3 (P0)**: hook UpdatedInput re-match 가드 + IsWriteOperation 패턴 보강 + `internal/permission/conflict.go` 신규 (specificity-then-fs-order tiebreak) (T-RT002-19..23). AC-04, AC-10, AC-12, AC-13 GREEN.
4. **M4 (P0)**: bubble.go DispatchToParent IPC contract + IsParentAvailable contract + fork depth >3 systemMessage sentinel + non-interactive fail-closed log (T-RT002-24..27). AC-08, AC-14, AC-15 GREEN.
5. **M5 (P1)**: doctor_permission.go `--all-tiers --mode --fork --format json` 보강 + `internal/permission/migration.go` (legacy bypassPermissions deprecation) + session_rules SrcSession tier 적재 contract + CHANGELOG + 7 MX tags + final `make build` + `go test ./...` (T-RT002-28..35). AC-05, AC-11 GREEN.

Total: 35 tasks across 5 milestones. Estimated scope: ~560 LOC new + ~150 LOC modified across 7 new files + 14 modified files.

[NOTE — Run-phase prerequisite] RT-002 is dependent on **SPEC-V3R2-RT-001** (Hook JSON protocol — provides `internal/hook.HookResponse` import) and **SPEC-V3R2-RT-005** (Settings reader — provides `internal/config.Source` 8-tier enum + reader). The skeleton already imports these dependencies; if either is unmerged at run-phase plan-audit gate, the run agent uses hardcoded test fixtures (resolver_test.go pattern) instead of full reader integration. Plan-audit gate verifies merge status of RT-001/RT-005 and decides whether to proceed with full integration or fixture-only mode.

---

## 다음 세션 시작점 (paste-ready resume message)

> Per `.claude/rules/moai/workflow/session-handoff.md` canonical 6-block format. Use this verbatim after `/clear` or in the next session if plan-auditor PASSes and run phase begins.

```text
ultrathink. SPEC-V3R2-RT-002 run 진입.
applied lessons: project_v3_master_plan_post_v214 (RT-002 plan PR open, batch4 Wave 12), lessons #11 retired-agent stub chain, lessons #14 worktree paste-ready Block 0.

전제 검증:
1) git -C /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002 log --oneline -1 → plan commit hash 확인
2) ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002/.moai/specs/SPEC-V3R2-RT-002/ → 6 files (spec/plan/research/acceptance/tasks/progress + issue-body)
3) gh pr view <PR-number> → MERGEABLE 또는 MERGED 상태 확인
4) gh pr list --search "SPEC-V3R2-RT-001 OR SPEC-V3R2-RT-005 in:title" --state merged → dependency 머지 확인
5) cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002 → worktree 활성

실행: /moai run SPEC-V3R2-RT-002

머지 후: SPEC-V3R2-RT-007 (다음 Wave 12 SPEC) → /moai sync
```

---

## Session-handoff triggers detected

Per `.claude/rules/moai/workflow/session-handoff.md` §When To Generate:

- [x] Trigger #2: SPEC phase completion (plan phase complete within v3 Round-2 multi-SPEC workflow Wave 12 batch)
- [x] Trigger #4: PR creation success when more SPECs remain in the current wave (RT-002 is one of RT-001/RT-002/RT-007/RT-005 batch)

---

## Run-phase completion markers (to be set by run phase)

| Field | Value | Set by |
|-------|-------|--------|
| `run_started_at` | _pending_ | run-phase orchestrator (M1 start) |
| `run_complete_at` | _pending_ | run-phase orchestrator (M5 close) |
| `run_status` | _pending_ → `implementation-complete` | run-phase orchestrator |
| `acs_passed` | _pending_ → 15/15 | manager-tdd verification |
| `tests_added` | _pending_ → ~30 | manager-tdd verification |
| `mx_tags_inserted` | _pending_ → 7 | manager-docs (T-RT002-31) |
| `pr_number` | _to be filled by manager-git_ | manager-git |
| `merged_commit` | _to be filled post-merge_ | manager-git |

---

End of progress.md.
