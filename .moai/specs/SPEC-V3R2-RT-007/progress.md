# SPEC-V3R2-RT-007 Progress Tracker

> Live progress and session-handoff state for **Hardcoded Path Fix + Versioned Migration**.
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
| Branch | `plan/SPEC-V3R2-RT-007` |
| Worktree | `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007` |
| Base | `origin/main` (`496595c3f`) |
| Plan-auditor | not yet run (PR will trigger) |
| Run-phase entry | pending plan-auditor PASS + RT-001 / RT-006 dependency confirmation |

---

## Plan Phase Deliverables (this session)

- [x] `spec.md` v0.1.0 (32 EARS REQs across 7 categories — verified by `grep -E "^- REQ-V3R2-RT-007-[0-9]+:" \| wc -l` → 32; 16 ACs — verified by `grep -E "^- AC-V3R2-RT-007-[0-9]+:" \| wc -l` → 16; 9 risks; 9 dependencies; BC-V3R2-008 with breaking=true)
- [x] `plan.md` v0.1.0 (5 milestones M1-M5, 24 file:line anchors verified via Grep/Read, traceability matrix 32 REQ → 16 AC → 40 tasks, mx_plan with 7 MX tag insertions across 7 distinct files, plan-audit-ready checklist 15/15 PASS with measured-evidence citations)
- [x] `research.md` v0.1.0 (12 sections, 29 file:line anchors, library evaluation with 0 new direct deps, breaking-change analysis, dependency status, spec-vs-reality drift acknowledged in §2.2 with 28-wrapper count + 0 hardcoded literal verification)
- [x] `acceptance.md` v0.1.0 (16 ACs in G/W/T format with happy-path + edge cases + test mapping; ~40 new test functions identified)
- [x] `tasks.md` v0.1.0 (40 tasks T-RT007-01..40 across M1-M5, owner roles per quality.yaml development_mode=tdd, dependencies with critical-path graph, cross-task constraints with 8 HARD rules)
- [x] `progress.md` v0.1.0 (this file)
- [x] `issue-body.md` (GitHub issue body for tracking, BC-V3R2-008 user-impact matrix)

### Plan-time evidence checks (verified before progress.md authoring)

| Verification | Command | Result |
|--------------|---------|--------|
| REQ count | `grep -E "^- REQ-V3R2-RT-007-[0-9]+:" .moai/specs/SPEC-V3R2-RT-007/spec.md \| wc -l` | **32** |
| AC count | `grep -E "^- AC-V3R2-RT-007-[0-9]+:" .moai/specs/SPEC-V3R2-RT-007/spec.md \| wc -l` | **16** |
| Hook wrapper count | `ls internal/template/templates/.claude/hooks/moai/*.tmpl \| wc -l` | **28** (spec.md cites 26 — drift acknowledged in research.md §2.2) |
| Hardcoded path occurrences in templates | `grep -rln "/Users/goos/go/bin/moai" internal/template/templates/.claude/hooks/moai/ \| wc -l` | **0** (already clean) |
| `internal/migration/` directory | `ls internal/migration/` | **does not exist** (new package per plan.md §3.2) |
| `detectGoBinPath` function | `grep -n "func detectGoBinPath" internal/core/project/initializer.go internal/cli/update.go` | initializer.go:286 + update.go:2515 (duplicate — to be unified per plan.md §3.1) |
| `$HOME` in passthrough tokens | `grep -n "\$HOME" internal/template/renderer.go` | line 42 — already registered (REQ-006 affirm-only) |

---

## Run Phase Plan (next session)

Per `plan.md` §9 Implementation Order Summary:

1. **M1 (P0)**: 13 tasks — ~22 RED tests across 12 new test files (T-RT007-01..13). Confirm RED for all; existing tests still GREEN.
2. **M2 (P0)**: 6 tasks — `gobin.Detect` helper + 2 call-site refactors + 3 audit lint affirms (T-RT007-14..19). AC-01, AC-02, AC-10 GREEN.
3. **M3 (P0)**: 7 tasks — `internal/migration/{runner,registry,version,log}.go` core (T-RT007-20..26). AC-04, AC-05, AC-11, AC-12, AC-14 GREEN.
4. **M4 (P0)**: 4 tasks — m001 + session-start hook + system.yaml `migrations.disabled` (T-RT007-27..30). AC-03, AC-04, AC-06, AC-13, AC-15 GREEN.
5. **M5 (P1)**: 10 tasks — `moai migration {run,status,rollback}` cobra group + `moai doctor --check migration` + 7 MX tags + CHANGELOG + final `make build` + `go test ./...` + git/PR closure (T-RT007-31..40). AC-07, AC-08, AC-09, AC-16 GREEN.

Total: 40 tasks across 5 milestones. Estimated scope: ~1,530 LOC new + ~30 LOC modified across 20 new files + 6 modified files.

Run-phase agent: `manager-cycle` (per quality.yaml `development_mode: tdd` + spec-workflow.md updated assignment).

---

## 다음 세션 시작점 (paste-ready resume message)

> Per `.claude/rules/moai/workflow/session-handoff.md` canonical 6-block format with Block 0 (worktree-anchored). Use this verbatim after `/clear` or in the next session if plan-auditor PASSes and run phase begins.

```text
[New Terminal — START IN WORKTREE]
$ cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007
$ moai cc
   └─ Claude Code session starts here (cwd = worktree)

ultrathink. SPEC-V3R2-RT-007 run 진입.
applied lessons: project_v3_master_plan_post_v214 (RT-007 plan PR open), lessons #11 retired-agent stub chain, lessons #12 worktree isolation discipline, lessons #13 --team base mismatch, lessons #14 worktree paste-ready Block 0.

전제 검증:
0) git rev-parse --show-toplevel → /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007 (★ critical)
1) git -C /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007 log --oneline -1 → plan commit hash 확인
2) ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007/.moai/specs/SPEC-V3R2-RT-007/ → 6 files (spec/plan/research/acceptance/tasks/progress + issue-body)
3) gh pr view <PR-number> → MERGEABLE 또는 MERGED 상태 확인 + RT-001 (HookResponse) + RT-006 (handler completeness) 의존성 머지 상태 확인

실행: /moai run SPEC-V3R2-RT-007

머지 후: SPEC-V3R2-EXT-004 (versioned migration general framework — depends on RT-007 MigrationRunner) → SPEC-V3R2-MIG-001 (v2→v3 user migrator) → /moai sync
```

---

## Session-handoff triggers detected

Per `.claude/rules/moai/workflow/session-handoff.md` §When To Generate:

- [x] Trigger #2: SPEC phase completion (plan phase complete within v3 Round-2 multi-SPEC workflow)
- [x] Trigger #4: PR creation success when more SPECs remain in the current wave (RT-007 is part of the v3 Master Plan Wave 12 alongside RT-001/RT-002 and RT-005)

---

## Run-phase completion markers (to be set by run phase)

| Field | Value | Set by |
|-------|-------|--------|
| `run_started_at` | _pending_ | run-phase orchestrator (M1 start) |
| `run_complete_at` | _pending_ | run-phase orchestrator (M5 close, T-RT007-38) |
| `run_status` | _pending_ → `implementation-complete` | run-phase orchestrator |
| `acs_passed` | _pending_ → 16/16 | manager-tdd verification (T-RT007-37) |
| `tests_added` | _pending_ → ~40 | manager-tdd verification |
| `mx_tags_inserted` | _pending_ → 7 | manager-docs (T-RT007-34) |
| `pr_number` | _to be filled by manager-git_ | manager-git (T-RT007-39) |
| `merged_commit` | _to be filled post-merge_ | manager-git (T-RT007-40) |

---

## Cross-SPEC coordination notes

- **RT-001 (HookResponse SystemMessage) blocker**: RT-007의 session-start hook integration이 `hookOut.SystemMessage` 필드 사용. RT-001 미머지 시 `slog.Warn` 임시 fallback (single-line swap to SystemMessage post-merge). plan-audit gate에서 RT-001 PR 상태 verify 필수.
- **RT-006 (SessionStart handler completeness) blocker**: RT-007의 핸들러 변경이 RT-006 변경 위에 들어가야 conflict-free. RT-006 머지 후 RT-007 run-phase 진입 권장.
- **CON-001 (FROZEN zone codification) blocker**: 마이그레이션 framework가 헌법적 메커니즘 — zone-registry 등록 필요. CON-001 머지 상태 verify.
- **RT-004 (Typed Session State) related**: `internal/session/lock.go` (`fileLock` interface) 재사용 가능 — RT-004 머지 후 import; 미머지 시 `internal/migration/version_unix.go` + `version_windows.go` build-tag separated 임시 자체 구현 (RT-004 머지 시 swap, ~20 LOC churn).
- **EXT-004 (versioned migration general framework) downstream**: RT-007이 base infrastructure 제공; EXT-004는 v2→v3 migration catalog 추가.

---

## Spec drift acknowledgement (research.md §2.2)

`spec.md` (drafted 2026-04-23) 의 일부 사실 진술이 코드 현실 (2026-05-10) 과 다음과 같이 다름:

1. **wrapper 개수**: spec.md "26 shell wrappers" → 실제 **28**.
2. **hardcoded literal 존재 유무**: spec.md "hardcoded `/Users/goos/go/bin/moai` in all 26 wrappers" → 실제 **0 hits in all 28 wrappers** (이미 깨끗).
3. **fallback chain 순서**: spec.md REQ-003 `PATH → $HOME → .GoBinPath` → 실제 `PATH → .GoBinPath → $HOME/go/bin → $HOME/.local/bin`.

이 차이는 본 SPEC의 *intent* 를 무효화하지 않음:

- migration framework 도입 (P-C06 closure) 은 그대로 의미 있음.
- 회귀 방지 lint 도입은 그대로 의미 있음.
- v2.x 사용자 retroactive fix는 그대로 의미 있음 (m001).

spec.md 본문 보정은 **별도 patch SPEC** 으로 분리 권장 (현 SPEC 내부 수정 시 plan-audit가 spec drift 로 reject 가능).

---

End of progress.md.
