# SPEC-V3R2-RT-001 Progress Tracker

> Live progress and session-handoff state for **Hook JSON-OR-ExitCode Dual Protocol**.
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
| Branch | `plan/SPEC-V3R2-RT-001` |
| Worktree | `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001` |
| Base | `origin/main` (`496595c3f`) |
| Plan-auditor | not yet run (PR will trigger) |
| Run-phase entry | pending plan-auditor PASS |
| Breaking | `true` (BC-V3R2-001) |
| Lifecycle | `spec-anchored` |

---

## Plan Phase Deliverables (this session)

- [x] `spec.md` v0.1.0 (25 EARS REQs across 6 categories — Ubiquitous 7 / Event 7 / State 3 / Optional 3 / Unwanted 3 / Complex 2; 15 ACs; 6 risks; 13 dependencies — 3 blocked-by + 5 blocks + 5 related)
- [x] `plan.md` v0.1.0 (5 milestones M1-M5, 30 file:line anchors, traceability matrix 25 REQ → 15 AC → 43 tasks, mx_plan with 7 MX tag insertions, plan-audit-ready checklist 15/15 PASS)
- [x] `research.md` v0.1.0 (12 sections, 35 file:line anchors, library evaluation, breaking-change analysis BC-V3R2-001, 27 events vs 27 wrappers reconciliation, dependency status)
- [x] `acceptance.md` v0.1.0 (15 ACs in G/W/T format with happy-path + edge cases + test mapping; ~28 new test functions across 5 new + 3 extended files)
- [x] `tasks.md` v0.1.0 (43 tasks T-RT001-01..43 across M1-M5 with owner roles + dependencies + critical path graph)
- [x] `progress.md` v0.1.0 (this file)
- [x] `issue-body.md` (GitHub issue body for tracking)

---

## Run Phase Plan (next session)

Per `plan.md` §9 Implementation Order Summary:

1. **M1 (P0)**: 18개 새 테스트 + 4개 신규 테스트 파일 추가 (T-RT001-01..18). 모두 RED 확인 + 기존 테스트 GREEN 유지.
2. **M2 (P0)**: validator/v10 의존성 추가 + HookResponse Validate() (T-RT001-19..22). AC-12 GREEN.
3. **M3 (P0)**: strict_mode + MOAI_HOOK_LEGACY env + once-per-session banner + HookConfig + system.yaml template (T-RT001-23..28). AC-04, AC-06, AC-15 GREEN.
4. **M4 (P0)**: HookSpecificOutputMismatch detection + UpdatedInput-then-Decision ordering + AdditionalContext routing + Continue:false escalation + 64 KiB truncation refactor (T-RT001-29..33). AC-02, AC-05, AC-08, AC-09, AC-10 GREEN.
5. **M5 (P1)**: api_version 2 + WatchPaths registrar + @MX routing + plugin bypass + CHANGELOG + 7 MX tags + make build + go test ./... + vet + lint + progress closure (T-RT001-34..43). AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN.

Total: 43 tasks across 5 milestones. Estimated scope: ~410 LOC new + ~150 LOC modified across 7 new Go files + 8 modified Go files + 2 YAML templates + 1 CHANGELOG. Wrappers-unchanged (26 셸 wrapper 모두 보존).

---

## 다음 세션 시작점 (paste-ready resume message)

> Per `.claude/rules/moai/workflow/session-handoff.md` canonical 6-block format. Use this verbatim after `/clear` or in the next session if plan-auditor PASSes and run phase begins.

```text
[New Terminal — START IN WORKTREE]
$ cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001
$ moai cc
   └─ Claude Code session starts here (cwd = worktree)

ultrathink. SPEC-V3R2-RT-001 run 진입.
applied lessons: project_v3_master_plan_post_v214 (RT-001 plan PR open), lessons #11 retired-agent stub chain, lessons #12 worktree isolation discipline, lessons #14 worktree paste-ready Block 0.

전제 검증:
0) git rev-parse --show-toplevel → /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001 (★ critical pre-check)
1) git -C /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001 log --oneline -1 → plan commit hash 확인
2) ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001/.moai/specs/SPEC-V3R2-RT-001/ → 6 files (spec/plan/research/acceptance/tasks/progress + issue-body)
3) gh pr view <PR-number> → MERGEABLE 또는 MERGED 상태 확인

실행: /moai run SPEC-V3R2-RT-001

머지 후: SPEC-V3R2-RT-002 (permission stack consumer) → SPEC-V3R2-RT-003 (sandbox via UpdatedInput) → /moai sync
```

---

## Session-handoff triggers detected

Per `.claude/rules/moai/workflow/session-handoff.md` §When To Generate:

- [x] Trigger #2: SPEC phase completion (plan phase complete within v3 Round-2 multi-SPEC workflow — Wave 12 breaking SPECs cluster RT-001/RT-002/RT-007/RT-005)
- [x] Trigger #4: PR creation success when more SPECs remain in the current wave (RT-001 is one of 4 v3R2 Wave 12 breaking SPECs; RT-002 follows in 2+2 sequential pair)

---

## Run-phase completion markers (to be set by run phase)

| Field | Value | Set by |
|-------|-------|--------|
| `run_started_at` | _pending_ | run-phase orchestrator (M1 start) |
| `run_complete_at` | _pending_ | run-phase orchestrator (M5 close, T-RT001-43) |
| `run_status` | _pending_ → `implementation-complete` | run-phase orchestrator |
| `acs_passed` | _pending_ → 15/15 | manager-tdd verification (T-RT001-42) |
| `tests_added` | _pending_ → ~28 | manager-tdd verification |
| `mx_tags_inserted` | _pending_ → 7 | manager-docs (T-RT001-40) |
| `wrappers_modified` | 0 (HARD constraint — wrappers-unchanged) | verified at PR review |
| `breaking_change_documented` | _pending_ → CHANGELOG entry under BC-V3R2-001 | manager-docs (T-RT001-39) |
| `pr_number` | _to be filled by manager-git_ | manager-git |
| `merged_commit` | _to be filled post-merge_ | manager-git |

---

End of progress.md.
