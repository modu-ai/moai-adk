# plan(spec): SPEC-V3R2-RT-006 — Hook Handler Completeness and 27-Event Coverage

> Plan PR for **SPEC-V3R2-RT-006** following Step 1 (plan-in-main) discipline.
> Base: `origin/main` HEAD `c810b11b7`. Branch: `plan/SPEC-V3R2-RT-006`. Merge strategy: squash.

## Summary

이 PR 은 SPEC-V3R2-RT-006 의 plan 산출물 7종을 main 에 squash-merge 합니다. 후속 run-phase 는 별도 worktree (`feat/SPEC-V3R2-RT-006`) 에서 실행됩니다.

### 산출물

| 파일 | 역할 | 크기 |
|------|------|------|
| `spec.md` | EARS 요구사항 (이미 main 에 존재; 수정 없음) | 28 KB |
| `research.md` | Phase 0.5 deep research, 12 sections, 30+ evidence anchors | ~28 KB |
| `plan.md` | Implementation plan, 10 sections, 35-task milestone breakdown (M1-M5) | ~24 KB |
| `tasks.md` | Task list (T-RT006-01..35) with REQ/AC traceability | ~10 KB |
| `acceptance.md` | 17 baseline ACs + 4 derived edge-case ACs with Given-When-Then | ~16 KB |
| `progress.md` | Phase tracker shell | ~6 KB |
| `issue-body.md` | This PR body | ~5 KB |

### 핵심 발견 (research.md §2 inventory)

- **5 critical handler upgrades 가 이미 main `c810b11b7` 에 머지됨**: subagentStop (185 LOC, P-H02 fix), configChange (105 LOC), instructionsLoaded (89 LOC), fileChanged (92 LOC), postToolUseFailure (134 LOC).
- **4 RETIRE-OBS-ONLY 헤더는 추가됨** (notification.go, elicitation.go, task_created.go) 그러나 settings.json 등록 제거 + observability_events 옵트인 메커니즘 미구현.
- **`setup.go` 본문 (0 bytes) 비워짐** 그러나 file 자체 잔존.
- **`audit_test.go` (152 LOC)** handler count + retired-not-active 만 검증; settings.json parity / per-file Resolution 헤더 / observability whitelist 미구현.

### Run-phase 잔여 작업 (35 tasks)

1. **M1 (P0)**: setup.go 제거 + 4 RED audit sub-tests + 6 RED AC-path tests + system.yaml hook 섹션 schema (4 tasks).
2. **M2 (P0)**: SystemHookConfig struct + observabilityOptIn helper + 4 retire 핸들러 gate 적용 (4 tasks).
3. **M3 (P0)**: settings.json template 4-event retire + 22 핸들러 Resolution 헤더 + 5 critical handler 의 잔여 검증 (timeout wrap, RT-005 reload, debounce, AC path tests) (12 tasks).
4. **M4 (P1)**: `moai doctor hook` 27-event 표 CLI + `--trace` (7 tasks).
5. **M5 (P0)**: full test/lint/build gates + manual tmux integration test + CHANGELOG + @MX tags (8 tasks).

## Phase contract

- Plan-phase: this PR (squash-merge into main).
- Run-phase: `moai worktree new SPEC-V3R2-RT-006 --base origin/main` after this PR merges.
- Sync-phase: same worktree as run.
- Cleanup-phase: `moai worktree done SPEC-V3R2-RT-006` after BOTH run+sync PRs merge.

## Acknowledged discrepancies (plan.md §1.2.1)

- spec.md §5.7 footer count "15 KEEP" 와 실제 row count "17 KEEP" 의 차이는 sync-phase HISTORY entry 에서 정정.
- `HookResponse` (spec) vs `HookOutput` (code) 명명 차이는 alias 로 처리; audit_test grep 가 두 이름 모두 검출.

## Plan-auditor risk areas (front-loaded mitigations per plan §9.3)

- spec §5.7 footer drift → §1.2.1 명시.
- HookResponse/HookOutput 명명 → audit grep 두 이름 모두 매칭.
- SPC-002 미머지 시 → T-RT006-18 stub interface + mock test.
- settings.json template branching breaks `moai update` → default empty map + Go template `{{- if }}`.
- 4 retire handler 코드 path 혼란 → §3.1 file-level map + per-file Resolution 헤더 + doctor_hook --observability.
- main HEAD drift → run-phase explicitly rebases on `origin/main`.

## Breaking change

- **BC-V3R2-018**: 4 retire 이벤트 (Notification, Elicitation, ElicitationResult, TaskCreated) 가 settings.json 에서 기본 제거. v2.x 사용자가 이 이벤트들을 hook 으로 사용 중이면, `system.yaml` 의 `hook.observability_events: ["notification"]` 등으로 옵트인 필요. CHANGELOG breaking-changes 섹션에 명시 (M5-T32).

## Dependencies

### Consumed (all merged)

- ✅ SPEC-V3R2-RT-001 (HookResponse / HookOutput protocol)
- ✅ SPEC-V3R2-RT-002 (PreToolUse PermissionDecision)
- ✅ SPEC-V3R2-RT-004 (SessionState checkpointing)
- ✅ SPEC-V3R2-RT-005 (8-tier resolver) — required for SystemHookConfig + Manager.Reload
- ⚠ SPEC-V3R2-SPC-002 (MX TagScanner) — TBD; T-RT006-18 has stub fallback

### Blocks (downstream)

- SPEC-V3R2-HRN-002 (evaluator memory per-iteration)
- SPEC-V3R2-WF-003 (Multi-mode router)
- SPEC-V3R2-MIG-002 (hook registration cleanup)

## Test plan

- [x] research.md cited 30+ file:line evidence anchors
- [x] plan.md REQ↔AC↔Task traceability matrix (§1.4) covers all 22 unique REQs and 17 ACs
- [x] tasks.md groups 35 tasks across 5 milestones, P0 first
- [x] acceptance.md provides Given-When-Then for each AC
- [x] No time estimates anywhere (P0/P1 priority labels only)
- [x] 16-language neutrality preserved (file_changed.go covers 21 extensions across all 16 languages)
- [x] @MX tag plan (plan.md §6) covers ANCHOR + WARN + NOTE
- [x] BC-V3R2-018 retire mechanism design (Option A, research §3.3) justified
- [x] Worktree-base alignment per Step 2 called out (plan §10)
- [ ] plan-auditor verdict: PASS ≥ 0.85 (target on first iteration)

## Files touched (plan-phase only)

- `.moai/specs/SPEC-V3R2-RT-006/research.md` (new)
- `.moai/specs/SPEC-V3R2-RT-006/plan.md` (new)
- `.moai/specs/SPEC-V3R2-RT-006/tasks.md` (new)
- `.moai/specs/SPEC-V3R2-RT-006/acceptance.md` (new)
- `.moai/specs/SPEC-V3R2-RT-006/progress.md` (new)
- `.moai/specs/SPEC-V3R2-RT-006/issue-body.md` (new — this file)

`spec.md` (existing 28 KB) 은 본 PR 에서 수정하지 않음. spec.md §5.7 footer count 정정은 sync-phase 에서 처리.

## Reference

- spec.md §1: P-H02 P-H03 P-H15 P-H16 P-H17 P-H19 P-R01 problem references
- master §7.3 + §8 BC-V3R2-018: retire mechanism authoritative source
- r6-commands-hooks-style-rules.md §A: 27-event coverage matrix authoritative input
- MEMORY.md `feedback_team_tmux_cleanup.md`: P-H02 root cause
- MEMORY.md `teammate_mode_regression.md`: stale binary regression warning
- `.moai/specs/SPEC-V3R2-RT-005/{plan,research,tasks,acceptance,progress,issue-body}.md`: reference plan structure adopted here

🗿 MoAI <email@mo.ai.kr>
