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

### M2 — B2/B3 흡수 (AC-SA-002·003 GREEN, B2b 운반, B7 재판정)

B2 RED (명령: `go test ./internal/statusline/ -run TestBoardRootResolvesThroughStateAnchor -count=1`, 플립 전):

```
--- FAIL: TestBoardRootResolvesThroughStateAnchor (0.20s)
    state_anchor_test.go:226: board root = "/private/var/folders/.../TestBoardRootResolvesThroughStateAnchor3399397510/001/deep/sub", want the state anchor ".../001" (not the visited subdir)
    state_anchor_test.go:232: github cache path = ".../001/deep/sub/.moai/state/github/counts.json", want ".../001/.moai/state/github/counts.json" (same anchor-fed board root)
FAIL
```

B3 RED (명령: `go test ./internal/statusline/ -run TestGoalArmedReadsFromStateAnchor -count=1`, 해당 호출점만 임시 역전 상태 — 아래 순서 비고 참조):

```
--- FAIL: TestGoalArmedReadsFromStateAnchor (0.16s)
    state_anchor_test.go:285: GoalArmed = false — the cd'd session cannot see the project's armed goal at .../TestGoalArmedReadsFromStateAnchor3419912556/001/.moai/state/goal
FAIL
```

GREEN (명령: `go test ./internal/statusline/ -run 'TestBoardRootResolvesThroughStateAnchor|TestGoalArmedReadsFromStateAnchor|TestResolveBoardRoot_PrefersPrimaryCheckoutOverWorktree' -count=1 -v`):

```
--- PASS: TestResolveBoardRoot_PrefersPrimaryCheckoutOverWorktree (0.28s)
--- PASS: TestBoardRootResolvesThroughStateAnchor (0.35s)
--- PASS: TestGoalArmedReadsFromStateAnchor (0.31s)
ok  github.com/modu-ai/moai-adk/internal/statusline	1.469s
```

- **B2b 운반 확인**: `builder.go`의 `boardRoot := resolveBoardRoot(input)` 하나가 `resolveGitHubCounts`·`maybeRefreshGitHubCounts`·landed에 함께 흐르고, AC-SA-002 테스트가 `githubCachePath(boardRoot) == <anchor>/.moai/state/github/counts.json`을 단언 (PASS) — 독립 앵커 없음.
- **순서 비고 (정직 기록)**: builder.go의 goal 읽기 호출점은 M1 커밋에 실수로 함께 플립돼 있었다. §D8 이행을 위해 M2에서 해당 호출점만 `resolveSessionDir`로 임시 역전하여 위 B3 RED를 관측한 뒤 복원했다 — RED 관측은 최종 트리와 그 한 호출점만 다른 상태에서 얻어졌다.
- **goal 소비자 3 좌표 체크리스트** (plan M2): ① `builder.go:292`(생산) — 플립 완료, ② `handoff_goal_suppress_test.go` — 렌더러 수준 테스트(data.GoalArmed 직접 설정)로 앵커 무관 확인, ③ `profile_bench_test.go` — `resolveStateAnchor`로 플립 완료.

**B7 범위 재판정 (판정서 부칙 지시 이행) — 결과: 범위 밖 유지.** 근거: `internal/hook/memo/writer.go` `Write(projectDir, …)`는 projectDir을 호출자 파라미터로 받고, 호출자(hook 사슬)는 `CLAUDE_PROJECT_DIR` env 우선 → Getwd 폴백(`cwd_fallback:true` 마커)으로 해석한다 — 훅 맥락의 앵커는 env에서 오지 세션이 cd한 current_dir에서 오지 않는다. 관측 계수 7건은 프로젝트당 compact 시점 기록 형태(방문 디렉터당 1레코드 오염 아님)와 일치. R1 시접(stdin 세션 맥락)에는 훅 사슬 대응물이 없으므로 흡수는 plan 확장(REQ-SA-008) 소관이며, AC-SA-008의 diff-0 가드가 계속 지킨다.

**M2 설계 노트 — stateanchor의 프로세스-cwd 폴백 제거 (M1 착지분 수정)**: M2에서 최소-stdin Build 테스트 2건(TestBuilder_SetMode·TestBuilder_Build_NoNewline)이 **운영자 실제 리포의 백로그를 렌더링에 끌어들인 실패**(4번째 줄 `🔄 TODO: 19/39`)를 관측했다 — 원인은 M1이 시접에 넣은 nil-input Getwd 폴백이 git 워크업을 타고 워크트리의 프라이머리(실제 리포)에 도달한 것. 수리: 폴백 제거 — 「디렉터 맥락이 없으면 앵커도 없다」(REQ-SA-003 엄격 돕기)로 테스트가 구성상 헤르메틱해진다. 생산 영향: stdin에 디렉터 필드가 전혀 없는 렌더(board·goal·telemetry 모두 skip)는 REQ-SA-003 스킵 경로 그 자체다. `resolveSessionDir`도 마지막 멤버 플립과 함께 삭제(AC-SA-012 유산 제거 — M5 grep으로 재확인).

**M2 의도된 변경 갱신 테스트 목록**:
1. `session_identity_test.go` TestResolveBoardRoot_PrefersPrimaryCheckoutOverWorktree — primary 케이스를 리터럴 경로 fixture에서 실제 git fixture 서브디렉터 워크업으로 교체 (리터럴 "/repo"는 git 해석에 응답 불가).
2. `stateanchor_test.go` TestResolve_EmptySessionResolvesViaProcessCWD → TestResolve_EmptySessionIsEmpty 재작성 (위 설계 노트의 폴백 제거 반영).

전 패키지 무파괴 (명령: `go test ./internal/statusline/ -count=1`): `ok github.com/modu-ai/moai-adk/internal/statusline 20.575s` (위 2건의 갱신분 제외 미갱신 무실패). `go test ./internal/stateanchor/ -count=1`: `ok ... 1.544s`.

### M3 — B4 config 캐시 (AC-SA-004, 2단 RED→GREEN)

**Stage-① 유도 지점 확정 (판정서 Gap 해소)**: B4의 cwd 연결은 config 패키지 내부에 없다(`grep -rn Getwd internal/config/ internal/paths/` = 무출력, 비테스트). 실제 유도 사슬: **`internal/cli/deps.go` `InitDependencies` — `cwd, _ := os.Getwd()` → `deps.Config.Load(cwd)`** → `ConfigManager.Load`가 `configDir = <projectRoot>/.moai`(+`MOAI_CONFIG_DIR` override)로 유도 → `LoadWithCache` → `cacheFilePath(configDir)` = `<configDir>/state/config-cache.json`. 캐시 착지의 전제: `<cwd>/.moai`가 존재할 것(cache.go의 config-dir-exists 가드) — 즉 B4 오염은 B1이 만든 stray `.moai`를 **뒤따르는** 형태였고, lane-1 t507 관측(`fixtures/.moai/state/config-cache.json`)과 정확히 같은 모양이다.

**Stage-② RED** (명령: `go test ./internal/config/ -run TestConfigCacheAnchorsToProject -count=1`, cwd형 유도 상태 — git fixture 루트 + stray `.moai`를 가진 서브디렉터 + 별도 비git 디렉터):

```
--- FAIL: TestConfigCacheAnchorsToProject (0.17s)
    state_anchor_test.go:81: config cache not anchored to the project root (want .../001/.moai/state/config-cache.json): stat .../001/.moai/state/config-cache.json: no such file or directory
    state_anchor_test.go:85: config cache mis-landed in the stray .moai of the visited dir (the B4 pollution shape): stat err = <nil>
FAIL
```

(`stat err = <nil>` = 캐시가 stray 안에 **실제로 쓰였다** — 부재가 아니라 오착지의 관측.)

GREEN (동일 명령, 유도 라인을 `stateanchor.FromDirectory(visited)`로 플립): `ok github.com/modu-ai/moai-adk/internal/config 0.766s`.

**수리**: deps.go가 `stateanchor.FromDirectory(cwd)`를 config 루트로 넘기고, git 맥락이 없을 때만 원시 cwd로 폴백(비git MoAI 프로젝트 보존 — 그 경우에도 캐시는 config-dir-exists 가드가 쓰기 화생성을 막는다). statusline→config 역방향 의존 없음: 해석은 호출 사슬(deps.go)이 stateanchor 시접을 직접 사용(plan M3 적용 지점 결정).

전 패키지 무파괴 (명령: `go test ./internal/config/ -count=1`): `ok github.com/modu-ai/moai-adk/internal/config 5.292s`. cli 접촉(deps.go 4줄)에 대해 `go vet ./internal/cli/...` = 통과, `go build ./...` = 통과, `go test ./internal/cli/ -run 'TestInitDependencies|TestDeps|TestGetDeps' -count=1` = `ok ... 1.076s`, `golangci-lint run ./internal/config/... ./internal/cli/...` = `0 issues.`

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M5에서 확정>_


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관. sync_commit_sha 기입 + GH #1694 회신 착지 기록(plan.md §F M5 인계)>_
