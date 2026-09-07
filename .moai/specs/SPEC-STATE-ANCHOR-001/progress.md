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

### M4 — 축 A 검증 (AC-SA-009·010·011, 코드 수리 없음 — D10)

**AC-SA-009 가드 활성 실측** (명령: `go test ./internal/cli/ -run 'TestTodoSweepSelectorMatchesFamily|TestGuardBypassMutant_ObserveHomePollution|TestTodoQueueRootGuard' -count=1 -v`):

```
--- PASS: TestTodoSweepSelectorMatchesFamily (0.03s)
    todo_axisa_guard_test.go:48: selector TestTodo matches 137 test functions (swept-set liveness guard)
--- PASS: TestGuardBypassMutant_ObserveHomePollution (0.08s)
    todo_axisa_guard_test.go:90: mutant pollution observed: 1 entr(ies) under canary HOME/.moai/todo — runTodo's liveTodoQueueRootReason gate is what keeps the guarded family at zero
--- PASS: TestTodoQueueRootGuard_FiresOnLiveRepository (0.07s)
--- PASS: TestTodoQueueRootGuard_SilentOnFixture (0.18s)
--- PASS: TestTodoQueueRootGuard_SilentOnHomeFallbackFixture (0.07s)
ok  github.com/modu-ai/moai-adk/internal/cli	1.832s
```

가드 판별식 표면(기존 구현 판독 — 재작성 없음): `liveTodoQueueRootReason`(`todo_queue_root_test.go:226`)이 `queueRootInsideTemp`로 OS temp 트리 양철자(/var·/private/var) 안이면 침묵, 라이브 리포면 `todoFixture` 지시와 함께 발화 — 위 3형제가 세 형태를 모두 고정. 패밀리 전체가 guarded 헬퍼(`runTodo`/`runTodoWithClosedStdin` 2개) 경유임은 소스 판독 + 아래 스윕이 그 헬퍼들로 전수 통과한 것으로 입증.

**AC-SA-010 canary-HOME 스윕** — 선택자 `TestTodo`(최상위 137 함수), 판정 도구는 커밋된 계측 테스트 `TestAxisACanaryHomeSweep_TodoFamily`(플래그 게이트 `MOAI_AXIS_A_CANARY_SWEEP=1` — 자식 `go test` 재실행 비용 때문에 CI 기본 제외, on-demand 재측정 도구; 상시 활성 liveness 핀은 `TestTodoSweepSelectorMatchesFamily`, N=0이면 양쪽 모두 스스로 실패). 실행 (명령: `MOAI_AXIS_A_CANARY_SWEEP=1 go test ./internal/cli/ -run 'TestAxisACanaryHomeSweep_TodoFamily' -count=1 -v -timeout 20m`):

```
todo_axisa_guard_test.go:125: sweep verdict: 198 todo tests ran under canary HOME /var/folders/.../TestAxisACanaryHomeSweep_TodoFamily4293617800/001 — 0 directories created under .moai/todo
--- PASS: TestAxisACanaryHomeSweep_TodoFamily (36.18s)
ok  github.com/modu-ai/moai-adk/internal/cli	37.162s
```

(198 = 137 최상위 + 서브테스트의 `--- PASS` 계수. canary HOME은 `t.TempDir()` — D5/D12대로 실제 HOME 미접촉.)

**AC-SA-011 뮤턴트 증명** — `TestGuardBypassMutant_ObserveHomePollution`(커밋 유지, 이름으로 우회임을 표시): `newTodoCmd()` 직접 Execute(게이트 우회) + `CLAUDE_PROJECT_DIR`=비git temp + canary HOME(userHomeDirFn seam) → **오염 1 엔트리 관측**(`canary/.moai/todo` 아래) — AC-SA-010의 0이 가드 덕분임의 판별 증거. 못 잡은 뮤턴트: 없음(1회 시도 1회 관측). 가드의 경계: 가드는 `runTodo`/`runTodoWithClosedStdin` 2 헬퍼의 호출자만 보호한다 — 새 헬퍼가 게이트 없이 `newTodoCmd()`를 부르면 뮤턴트가 관측한 바로 그 오염이 재발한다(REQ-SA-011 보고 의무 이행).

**축 A 코드 diff**: 생산 코드 변경 0 — M4 신규 파일은 테스트 1개(`internal/cli/todo_axisa_guard_test.go`)뿐. D10의 「코드 변경 금지」는 가드 행위의 변경 금지로 읽었다(AC-SA-009의 「테스트로 고정」·AC-SA-011의 뮤턴트 보관 지시가 테스트 추가를 요구하므로); 가드 코드(e7a078970 착지분)는 무접촉.

### M5 — 배치 검증 (§C 재측정 + AC 전수 판정)

**§C 전 행 재측정** (2026-09-07, 최종 HEAD `e0196a517` — 아래 M5 커밋 직전 기준):
- `git rev-parse --short HEAD` = `e0196a517`, `git branch --show-current` = `WT-state-write-locus`; `git merge-base --is-ancestor 0b1e27877 HEAD` = 참 (진입 시 흡수 상태 유지).
- B1 committed RED probe → C0 테스트로 상시화, GREEN (아래 AC-SA-001).
- `grep -n "\.ProjectDir" internal/statusline/context_usage.go` → 무출력 → **수리 후** `resolveStateAnchor` 경유로 교체됨 (아래 grep 집합).
- `grep -n "liveTodoQueueRootReason" internal/cli/todo_test.go` → 존재 (가드 무접촉).
- canary baseline: `/bin/ls ~/.moai/todo | wc -l` = **343 (작업 전) → 343 (전 마일스톤 종료 후)** — 신규 디렉터 0.
- 기존 테스트 무파괴 baseline → 최종 전수 GREEN (아래).

**AC-SA-012 단일 시접 관측 grep 집합** (M5 확정분):
1. `grep -rn "resolveProjectDir\|resolveSessionDir" internal/statusline/` → 코드 매치 0 (유일 매치는 state_anchor_test.go:29의 RED 이력 기록 주석). 전 멤버의 current_dir 앵커 유산 제거 완료.
2. `grep -n "resolveStateAnchor" internal/statusline/builder.go internal/statusline/backlog.go internal/statusline/state_anchor.go` → 정확히 4매치: 정의(`state_anchor.go:22`) + B1 쓰기(`builder.go:181`) + B3 goal 읽기(`builder.go:292`) + B2 board root(`backlog.go:31`).
3. `grep -rn "Getwd" internal/statusline/*.go | grep -v _test` → `builder.go:439`(extractProjectDirectory — 표시, 불변 D2)·`memory.go:73`(LLM yaml 워크업, 상태 무관)·`version.go:113`(버전 수집, 상태 무관) — 상태 쓰기·읽기 경로의 Getwd 앵커 0.
4. `grep -rn "CurrentDir" internal/statusline/*.go | grep -v _test` → 표시 유도(`builder.go:428-429`)와 시접 입력(`state_anchor.go:27`)뿐 — REQ-SA-002의 「git 해석의 기점」으로만 사용.
5. `grep -n "stateanchor\." internal/cli/deps.go` → `:178 FromDirectory(cwd)` — B4도 동일 시접.
6. 보너스: `internal/stateanchor/stateanchor.go` 자체에 `Getwd` 없음(M2에서 폴백 제거) — 시접 내부에도 cwd 앵커 경로 없음.

**AC-SA-008 B5/B6 불변** (명령: `git diff 0b1e27877..HEAD --stat -- internal/hook/ internal/session/`): **빈 출력** — SPEC 기준 트리 대비로도, 흡수 기점 `36b1aff8f` 대비로도 0. `internal/hook/`·`internal/session/` 무접촉 확인.

**AC-SA-006/007 불변 유지 증거**: AC-SA-006 = `TestDisplaySegmentUnchangedByAnchorRepair` PASS(M1 기록, 발산 입력 코퍼스) + `extractProjectDirectory`(builder.go:415-438) 무변경. AC-SA-007 = throttle·silent-failure 기존 테스트(`context_usage_test.go` TestWriteContextUsage_*·`session_telemetry_payload_test.go` throttle군) 무수정 통과 — 아래 전 패키지 run에 포함.

**최종 스코프 게이트** (명령: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/stateanchor/ ./internal/statusline/ ./internal/config/ -count=1` + `go test ./internal/cli/ -run TestTodo -count=1` + `go vet ./internal/statusline/... ./internal/stateanchor/... ./internal/config/... ./internal/cli/...` + `golangci-lint run` 동일 스코프 + `GOOS=windows GOARCH=amd64 go build ./internal/...`):

```
ok  github.com/modu-ai/moai-adk/internal/stateanchor	1.144s
ok  github.com/modu-ai/moai-adk/internal/statusline	15.674s
ok  github.com/modu-ai/moai-adk/internal/config	2.702s
ok  github.com/modu-ai/moai-adk/internal/cli	69.115s   ( -run TestTodo )
VET_OK
0 issues.
GOOS_WINDOWS_OK
```

**AC 판정 매트릭스 (최종)**:

| AC | 판정 | 1차 증거 |
|---|---|---|
| AC-SA-001 | **PASS** | C0 RED → M1 GREEN (`TestContextUsageAnchorsToProjectDir`) |
| AC-SA-002 | **PASS** | M2 RED → GREEN (`TestBoardRootResolvesThroughStateAnchor`, B2b 경로 단언 포함) |
| AC-SA-003 | **PASS** | M2 RED → GREEN (`TestGoalArmedReadsFromStateAnchor`) |
| AC-SA-004 | **PASS** | M3 2단 RED → GREEN (`TestConfigCacheAnchorsToProject`) |
| AC-SA-005 | **PASS** | M1 RED → GREEN (`TestNoProjectNoState`) |
| AC-SA-006 | **PASS** | `TestDisplaySegmentUnchangedByAnchorRepair` + `extractProjectDirectory` 무변경 |
| AC-SA-007 | **PASS** | throttle 기존 테스트 무수정 통과 (전 패키지 run) |
| AC-SA-008 | **PASS** | `git diff 0b1e27877..HEAD -- internal/hook/ internal/session/` 빈 출력 |
| AC-SA-009 | **PASS** | 가드 3형제 PASS + 판별식 표면 고정 |
| AC-SA-010 | **PASS** | canary 스윕: 198 pass entries, 오염 0, N=0 자기실패 도구 커밋 |
| AC-SA-011 | **PASS** | 뮤턴트 오염 1 엔트리 관측 (`TestGuardBypassMutant_ObserveHomePollution`) |
| AC-SA-012 | **PASS** | 위 grep 집합 1-6 |

**12/12 PASS. 미해결 결함·차단 0.**

**GH #1694 회신 초안 재료 (sync-phase 인계)**: (a) 원인 — statusline이 세션 원격청구 텔레메트리를 stdin `current_dir`에 기록(세션이 cd한 디렉터마다 stray `.moai` 1개, 제보 226개와 직접 일치; config-cache 150건은 그 stray를 뒤따른 B4); (b) 수리 — 단일 상태-앵커 시접(`internal/stateanchor`, REQ-SA-002 체인)으로 B1/B2(+B2b)/B3/B4 전부를 프로젝트 루트에 고정, 무프로젝트는 쓰기 생략; (c) 제보자 환경의 226개 stray `.moai`는 사용자 측 정리 대상(Out of Scope) — 안내는 sync 단계에서 확정.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-07
run_commit_sha: "15c6f5857" (D3 백필 — run 최종 커밋 M5; sync 진입 시 HEAD로 재확인)
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 0 (PRESERVE 대상 — extractProjectDirectory·B5/B6·throttle·축 A 가드 — 전부 무변경, AC-SA-006/007/008/009로 입증)
l44_pre_commit_fetch: n/a (레인 로컬 브랜치 — push는 리드 일괄 소관, repo-local git-flow)
l44_post_push_fetch: n/a (동일)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues, go vet clean — 4개 접촉 패키지)
cross_platform_build.darwin: pass (로컬 전 패키지 빌드+테스트)
cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./internal/...)
cross_platform_build.linux: pending-ci (CI 매트릭스 판정 몫)
total_run_phase_files: 17 (신규 4: stateanchor.go/.go 테스트, statusline state_anchor.go/테스트, config state_anchor_test.go, cli todo_axisa_guard_test.go + 수정 10 + SPEC 산출물 3)
m1_to_mN_commit_strategy: per-milestone commits (M1 6e0c6625a / M2 ee680220d / M3 1150d1f14 / M4 e0196a517 / M5 본 커밋), card id t510 전 커밋 명기


## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-07
sync_commit_sha: "pending-backfill-sync" (본 커밋은 자신의 SHA를 인용할 수 없음 — D3 백필 창, 다음 커밋에서 확정값 기입)
sync_status: complete
changelog_entry: CHANGELOG.md [Unreleased] `### Fixed` 첫 항목 — 외부 제보 링크는 [#1694](https://github.com/modu-ai/moai-adk/issues/1694), 선례 형식(#1632/#1640) 준수
b12_self_test_a: pass (pre-emission grep — `grep -c 'SPEC-STATE-ANCHOR-001' CHANGELOG.md` = 0, 병렬 BATCH-SYNC 중복 없음)
b12_self_test_b: pass (AC count match — acceptance.md distinct AC = `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' … | sort -u | wc -l` = 12; 엔트리 기재 "All 12 acceptance criteria (AC-SA-001..012)"와 동일)
b12_self_test_c: pass (file path verification — 엔트리 인용 경로 전부 `ls` 실존 확인: internal/stateanchor/, internal/statusline/context_usage.go, internal/statusline/backlog.go, internal/cli/deps.go, .moai/reports/t510/gh-1694-reply-draft.md)
frontmatter_status_transitions:
  spec.md: "in-progress → completed" (본 sync 커밋에 병합 — 3-phase close; status + updated만 변경)
  plan.md / acceptance.md: 해당 없음 — status 필드 미보유 (Artifact Statelessness, spec-frontmatter-schema.md)
canary_compliance_check:
  sync_phase_home_contact: none (sync 페이즈는 테스트 실행·측정 0건 — 문서·frontmatter 변경뿐이므로 D5/D12 canary 규율과 실접촉 없음)
  run_phase_canary_baseline: 343 → 343 (§E.2 M5 재측정 인용 — sync에서 재측정하지 않음, run 기록 그대로)
gh_1694_reply: 초안 sync-phase 확정 — .moai/reports/t510/gh-1694-reply-draft.md. **게시는 리드의 릴리스-타이밍 확인 게이트 대기**(수리의 develop 착지 + 릴리스 일정 확인 후). pre-run 초안 대비 변경 3건: (1) session-memo 7건의 결함 귀속 철회 — B7 재판정(§E.2 M2)에 따라 프로젝트당 compact 기록으로 분리, (2) 수리 시접 명명(internal/stateanchor 고정 우선순위 체인) + 수리 4가족 명시, (3) 무프로젝트 쓰기 생략 + cd 세션 goal 가시성 수리 반영. 정리 레시피·display 불변 문단은 pre-run에서 불변.
docs_site_readme_judgment: no user-facing doc change required — README 4로케일·docs-site content의 `.moai/state/...` 서술은 전부 프로젝트 루트 기준 상대 경로 기술이며(예: `.moai/state/goal/<session-id>.json`, `.moai/state/context-usage/<session-id>.json`), cwd-앵커(방문 디렉터 착지)를 서술하는 문서 0건. 수리는 문서가 이미 기술한 경로로 현실을 맞춘 것 — 경로·스키마 체계는 불변(D13). 측정: `grep` over README*.md + docs-site/content/ (`\.moai/state|context-usage|config-cache|state/goal|landed/counts`).
user_facing_doc_change_required: false
