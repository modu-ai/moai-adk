# Progress — SPEC-STATE-ANCHOR-001

카드: t510 · Tier M · 근거 보고서 `.moai/reports/t510/verdict.md`

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-audit
plan_complete_at: 2026-09-07 (plan-phase artifacts authored by manager-spec — plan-auditor verdict 대기)
plan_artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M 3 + progress)

## §E.2 Run-phase Evidence

> 귀속 삼인조 규율: 각 항목 = (a) 명령 + (b) 관측 출력 전문 + (c) baseline attribution `(this run, this tree)` + HEAD SHA.
> 진입 시 §C 재측정 (2026-09-07, HEAD `592e2470e` = base `0b1e27877` + docs 3커밋 + develop 흡수 `36b1aff8f`; `git merge-base --is-ancestor 0b1e27877 HEAD` = 참):
> - `grep -n "\.ProjectDir" internal/statusline/context_usage.go` → 무출력 (exit 1) — project_dir 후보 부재 재확인 (plan §C 행의 "ProjectDir 0매치"는 필드 접근 기준; 함수명 자체는 `resolveProjectDir`로 존재)
> - `grep -n "liveTodoQueueRootReason" internal/cli/todo_test.go internal/cli/todo_queue_root_test.go` → 7매치 (t422 가드 존재)
> - `/bin/ls ~/.moai/todo | wc -l` → **343** (작업 전 canary baseline; 검증 중 오염 금지 D5/D12)

### C0 — B1 committed RED (AC-SA-001의 RED-now 셀)

명령: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/statusline/ -run TestContextUsageAnchorsToProjectDir -count=1`

관측 출력 전문 (RED — 수리 전 체인에서):

```
--- FAIL: TestContextUsageAnchorsToProjectDir (0.00s)
    state_anchor_test.go:70: telemetry record not anchored to project_dir (want /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestContextUsageAnchorsToProjectDir401269422/002/.moai/state/context-usage/sess-anchor-001.json): open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestContextUsageAnchorsToProjectDir401269422/002/.moai/state/context-usage/sess-anchor-001.json: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.457s
```

RED right-reason 입증 (임시 프로브, 관측 후 삭제 — verdict §B-2 동일 형태, 같은 트리에서 재관측):

```
zz_t510_c0_probe_test.go:47: probe: visited-dir record exists=true (err=<nil>); project-dir record exists=false (err=stat .../TestZZC0ProbeWhereRecordLands1459067315/002/.moai/state/context-usage/sess-c0probe.json: no such file or directory)
```

attribution: (this run, this tree) @ `592e2470e` — B1 수리 전 상태.

### M1 — 시접 착지 + B1 GREEN

신설: `internal/stateanchor`(단일 시접 — REQ-SA-002 체인: project_dir → original_cwd → `gitcore.ResolveGitDirs` common-dir 부모 → ""), `internal/statusline/state_anchor.go`(statusline 어댑터 + 수리 전 체인 보류 `resolveSessionDir` — M2/M3 플립 대기분). 삭제: `resolveProjectDir`(context_usage.go). B1 호출점(builder.go:178)을 리졸버로 교체. B2(backlog.go)/B3(builder.go:286)은 §D8에 따라 각자의 RED 관측 전까지 동작 불변(`resolveSessionDir` 경유).

AC-SA-005 RED (명령: `go test ./internal/statusline/ -run TestNoProjectNoState -count=1`, B1 플립 전 임시 배선 상태):

```
--- FAIL: TestNoProjectNoState (4.37s)
    state_anchor_test.go:124: no-project render created .moai under the visited dir (must skip, REQ-SA-003): stat err = <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/statusline	4.833s
```

GREEN (AC-SA-001·005·006·007 판정 명령 — 명령: `go test ./internal/statusline/ -run 'TestContextUsageAnchorsToProjectDir|TestNoProjectNoState|TestDisplaySegmentUnchangedByAnchorRepair|TestResolveStateAnchor_Chain' -count=1 -v`):

```
--- PASS: TestContextUsageAnchorsToProjectDir (0.00s)
--- PASS: TestNoProjectNoState (0.13s)
--- PASS: TestDisplaySegmentUnchangedByAnchorRepair (0.00s)
--- PASS: TestResolveStateAnchor_Chain (0.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.573s
```

시접 단위 테스트 (명령: `go test ./internal/stateanchor/ -count=1`): `ok github.com/modu-ai/moai-adk/internal/stateanchor 1.225s` — 7 테스트(체인 1·2·3단, 비git → "", 프로세스 cwd 폴백, 워크트리=프라이머리 단일 루트, 빈 디렉터 가드).

기존 테스트 무파괴 (명령: `go test ./internal/statusline/ -count=1` — 전 패키지): `ok github.com/modu-ai/moai-adk/internal/statusline 22.540s`. baseline(수리 전, §C 마지막 행): `ok ... 20.810s`.

**의도된 변경으로 갱신된 테스트 목록** (plan §F M2 무파괴 정의 적용 — 미갱신 테스트 무실패는 위 전 패키지 run이 입증):
1. `builder_test.go` TestBuild_WritesContextUsageWithSessionID — fixture에 `ProjectDir: proj` 추가 (체인 1단 → 앵커=proj, 테스트 의도·단언 불변).
2. `session_telemetry_test.go` `buildForSession` — 동일 사유로 `ProjectDir: projDir` 추가 (TestPerSessionRecordPath·TestRecordKeyIsThePayloadSessionID 커버).
3. `session_telemetry_payload_test.go` `buildWith` — 동일 (TestRecordedModelIsBackendResolved·TestModelAndEffortRoundTripAndOmit 커버).
4. `forge_spawn_gate_test.go` `builderStdin` — stdin JSON에 `"project_dir":root` 추가 (boardRoot가 앵커에서 해석되도록 — 미갱신 시 "7/3" 캐시 단언이 깨지는 것을 수리로 예방).
5. `context_usage_test.go` TestResolveProjectDir_Chain → TestResolveStateAnchor_Chain 재작성 (피검 함수 삭제·체인 시맨틱 교체 — stateanchor 단위 테스트와 분업).
6. `profile_bench_test.go` — `resolveProjectDir` 참조 4건 교체: writeContextUsage 2건 → `resolveStateAnchor`(B1), resolveGoalArmed 2건 → `resolveSessionDir`(M2에서 플립).

**알려진 환경 플레이크 (본 SPEC 변경 무관)**: `TestResolveBacklogCounts_LatencyBudget`가 기계 부하 시 sqlite p95 스파이크(2.7s)로 실패 가능 — 격리 재실행 2회 모두 `ok`(중앙값 1.85ms < 10ms 예산). `git status --porcelain`으로 backlog_sqlite/kanban/store 무접촉 확인. M5에서 재측정.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M5에서 확정>_


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관. sync_commit_sha 기입 + GH #1694 회신 착지 기록(plan.md §F M5 인계)>_
