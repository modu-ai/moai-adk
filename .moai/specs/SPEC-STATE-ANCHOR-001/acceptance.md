# Acceptance — SPEC-STATE-ANCHOR-001

측정 기준 트리: `.claude/worktrees/t510` @ `0b1e27877` (`WT-state-write-locus`). 아래 RED 값 중 B1은 판정서 `.moai/reports/t510/verdict.md` §B-2에서 **이 트리에서 실제로 관측된 값**이며, B2/B3/B4의 RED는 plan.md §D8에 따라 run-phase가 멤버별 committed RED 테스트로 관측할 예정 형태다(채택 판정은 관측 후).

등급 용어: **blocking** = RED-now 셀 + green path 셀을 갖춘 릴리스 게이트. **regression-guard** = 기준 트리에서 이미 GREEN이라 RED-now 셀을 가질 수 없는 회귀 방어(verification-completeness §2의 undecidable disposition 적용 — release-blocking 자격 없음, 기록도 pass로 남기지 않는다). 축 A의 3건(AC-SA-009/010/011)이 regression-guard다 — 판정서 Claim 2: 축 A는 이미 수리돼 있다. 라벨 「blocking (RED 관측 전 — 채택은 run-phase)」(plan-audit D4)은 RED가 **예정형**인 AC-SA-002~005에 붙는다: 예상 RED 형태 + 코드 근거를 기록해두고 실제 RED 관측(채택)은 run-phase 마일스톤이 수행한다 — plan.md §D8(멤버별 committed RED)과 §E E8이 그 관측을 기계적으로 강제한다.

---

## §D AC 매트릭스

| AC | 요구사항 | RED (현재 트리) | GREEN (목표) | 등급 |
|---|---|---|---|---|
| AC-SA-001 | REQ-SA-001, 002, 007 | 레코드가 visited(A)에 착지, project(B) 무오염 — 판정서 §B-2 실측 | 레코드가 B에만 착지, A 무오염 | blocking |
| AC-SA-002 | REQ-SA-001, 002 | `original_cwd` 없이 서브디렉터 cd → board root가 서브디렉터 (M2 RED 관측 예정) | board root가 상태 앵커 (+ github counts 경로) | blocking (RED 관측 전 — 채택은 run-phase) |
| AC-SA-003 | REQ-SA-006 | goal 시드 + 서브디렉터 cd → GoalArmed=false (M2 RED 관측 예정) | GoalArmed=true | blocking (RED 관측 전 — 채택은 run-phase) |
| AC-SA-004 | REQ-SA-001 | configDir 유도가 cwd 좌주 (M3 RED가 유도 지점 먼저 고정) | 캐시가 프로젝트 앵커에 착지, cwd 무오염 | blocking (RED 관측 전 — 채택은 run-phase) |
| AC-SA-005 | REQ-SA-003 | 비git 디렉터에서도 상태 생성 (M1 RED 관측 예정) | 쓰기 생략 + 렌더 정상 완료 | blocking (RED 관측 전 — 채택은 run-phase) |
| AC-SA-006 | REQ-SA-004 | (불변 대상 — 변화 없음이 현재 상태) | 표시 세그먼트 출력 수리 전후 동일 | blocking |
| AC-SA-007 | REQ-SA-005 | (불변 대상 — 현재 동작) | throttle + silent-failure 유지, 기존 테스트 통과 | blocking |
| AC-SA-008 | REQ-SA-008 | (불변 대상 — diff 0이 현재 상태) | `internal/hook/`, `internal/session/` diff 0 | blocking |
| AC-SA-009 | REQ-SA-009 | (이미 GREEN — regression-guard) | 가드 활성 유지 실측 | regression-guard |
| AC-SA-010 | REQ-SA-010 | (이미 GREEN — regression-guard) | canary 스윕 0 오염 + swept count > 0 | regression-guard |
| AC-SA-011 | REQ-SA-011 | (뮤턴트 오염은 지금도 재현 가능 — 그것이 증거) | 우회 뮤턴트의 오염 관측 기록 | regression-guard |
| AC-SA-012 | REQ-SA-001 | 상태 앵커 유산이 멤버별 분산 (resolveProjectDir 사슬 공유) | 단일 리졸버 함수로 통일, current_dir 앵커 유산 제거 | blocking |

---

## §D.1 AC 상세

### AC-SA-001 — 텔레메트리가 project_dir에 착지한다 (B1 · #1694 주범)

- **Given** stdin에 `workspace.current_dir`(임의 디렉터 A, `t.TempDir()`)과 `workspace.project_dir`(임의 디렉터 B, `t.TempDir()`)을 **둘 다** 준 statusline 렌더 입력이 있고,
- **When** 렌더가 완료되면,
- **Then** 세션 텔레메트리 레코드(`.moai/state/context-usage/<sid>.json`)는 B 아래에만 존재하고, A 아래에는 `.moai`가 생성되지 않는다.

판정 명령(커밋되는 테스트 — plan.md C0): `go test ./internal/statusline/ -run TestContextUsageAnchorsToProjectDir -count=1`

**RED (실측, `0b1e27877`)** — 판정서 §B-2 전문 인용:

```
=== RUN   TestT510Probe_TelemetryAnchorsToCurrentDirNotProjectDir
    t510_probe_test.go:42: probe: record=/var/folders/.../TestT510Probe_.../002/.moai/state/
                          context-usage/t510probe0001.json (visited dir);
                          project dir carries no .moai — cwd-anchored write confirmed at HEAD
--- PASS: TestT510Probe_TelemetryAnchorsToCurrentDirNotProjectDir (0.00s)
ok  github.com/modu-ai/moai-adk/internal/statusline  0.437s
```

프로브는 "방문한 디렉터에 착지했다"를 관측한 것(PASS 형태의 관측 테스트)이고, 커밋되는 C0 테스트는 같은 입력에서 "B에 착지 + A 무오염"을 단언하므로 HEAD에서 **FAIL (RED)** 이다. run-phase가 C0의 실측 RED 출력을 progress.md §E.2에 남긴다(E8).

**GREEN (M1)**: `ok github.com/modu-ai/moai-adk/internal/statusline` + 테스트 PASS.

### AC-SA-002 — landed/board root가 상태 앵커에서 해석된다 (B2)

- **Given** `worktree.original_cwd`가 없고 `workspace.current_dir`가 프로젝트 서브디렉터인 stdin 입력(프로젝트 루트는 git repo fixture)이 있고,
- **When** backlog/landed 세그먼트가 board root를 해석하면,
- **Then** board root는 상태 앵커(리졸버 결과)에서 나온다 — 서브디렉터가 아니다.
- **And** `worktree.original_cwd`가 **있는** 경우 기존 동작(그것을 우선)이 유지된다 — 이는 리졸버 체인 2단계로 흡수된다.
- **And** github counts(`builder.go:265/267`)가 같은 `boardRoot`(`builder.go:255`)에서 공급되므로, 수리 후 github 캐시(`state/github/counts.json`)도 앵커 아래에 착지한다 — github 캐시를 visited 디렉터나 별도 위치에 쓰는 수리는 실패다(B2b — 판정서 부칙: 독립 앵커 없음, B2 수리가 운반).

판정 명령: `go test ./internal/statusline/ -run TestBoardRootResolvesThroughStateAnchor -count=1`

**RED (M2 관측 예정)**: 현재 `backlog.go:28`의 폴백이 `resolveProjectDir`(current_dir 사슬)이므로 board root가 서브디렉터로 해석됨 — 테스트 FAIL. 코드 근거: `backlog.go:24-29` (직접 판독). **채택은 run-phase가 실측 RED 출력을 남긴 후다.**

**GREEN (M2)**: 테스트 PASS. original_cwd 케이스 하위 테스트도 PASS.

### AC-SA-003 — goal 읽기가 cd 내성을 갖는다 (B3 · 읽기 측 결함)

- **Given** 프로젝트 루트 `.moai/state/goal/`에 armed goal 상태가 시드돼 있고 stdin의 `current_dir`가 그 프로젝트의 서브디렉터이며,
- **When** statusline 렌더가 goal 상태를 읽으면,
- **Then** `GoalArmed == true` — 프로젝트의 goal 상태가 cd한 세션에도 보인다.

판정 명령: `go test ./internal/statusline/ -run TestGoalArmedReadsFromStateAnchor -count=1`

**RED (M2 관측 예정)**: 현재 `builder.go:286`가 B1 사슬 앵커(서브디렉터)에서 읽어 armed를 못 봄 — `GoalArmed=false`, 테스트 FAIL.

**GREEN (M2)**: `GoalArmed=true`. **가시성 변화 주의** — 이 수리는 "cd한 세션이 goal을 보게 된다"는 동작 변화를 수반하며, cwd 기반 읽기에 의존하던 기존 테스트가 있으면 M2에서 함께 갱신한다(수리가 깨뜨린 것이 아니라 의도된 변화).

### AC-SA-004 — config 캐시가 cwd에 착지하지 않는다 (B4)

- **Given** config 캐시 쓰기를 트리거하는 CLI 사슬 호출이 비git 임시 디렉터를 cwd로 수행되고 프로젝트 루트(git repo fixture)가 별도로 있으며,
- **When** 캐시 쓰기가 발생하면,
- **Then** 캐시 파일은 프로젝트 앵커의 `.moai/state/config-cache.json`에 착지하고 cwd에는 `.moai`가 생성되지 않는다.

판정 명령: `go test ./internal/config/ -run TestConfigCacheAnchorsToProject -count=1`

**RED (M3 관측 예정 — 2단)**: ① 먼저 configDir 유도 지점을 함수 단위로 고정하는 RED 테스트(판정서 Gap 해소 — "정확한 유도 지점은 run-phase RED 테스트가 고정"), ② 그 위에서 cwd 착지를 단언하는 RED. lane-1 t507 관측(`fixtures/.moai/state/config-cache.json`)이 형태 대조군이다.

**GREEN (M3)**: 테스트 PASS. `go test ./internal/config/...` 무파괴.

### AC-SA-005 — 무프로젝트에서는 쓰기가 없다 (no project, no state)

- **Given** stdin에 `project_dir`도 `original_cwd`도 없고 `current_dir`가 비git 임시 디렉터(`t.TempDir()`)인 입력이 있고,
- **When** 렌더가 완료되면,
- **Then** 그 디렉터 아래에 어떤 상태 디렉터리도 생성되지 않고, 렌더는 오류·패닉 없이 정상 완료한다.

판정 명령: `go test ./internal/statusline/ -run TestNoProjectNoState -count=1`

**RED (M1 관측 예정)**: 현재 리졸버가 `current_dir`/`Getwd`를 앵커로 반환하므로 상태가 생성됨 — FAIL. **GREEN (M1)**: skip 경로 + 정상 완료 PASS. skip이 로그 소란이나 렌더 실패를 만들면 이 AC가 실패한다(spec.md §3 mutant 4).

### AC-SA-006 — 표시는 불변이다

- **Given** 동일 stdin 입력에 대해 수리 전(기준 `0b1e27877`) 렌더 출력의 디렉터 표시 세그먼트가 기록돼 있고, 골든 코퍼스에 **`project_dir`(A) ≠ `current_dir`(B)인 입력이 포함된다**(AC-SA-001과 같은 2디렉터 fixture 재사용 — 표시 소스가 실제로 갈라지는 지점이 이 입력뿐이므로, 동일-디렉터 입력만 담으면 오지시 이행도 골든을 통과한다 / plan-audit D2), 그리고,
- **When** 수리 후 코퍼스 입력으로 렌더하면,
- **Then** 디렉터 표시 세그먼트(basename 유도 포함)의 출력이 동일하다 — 표시 유도는 기존 `extractProjectDirectory`(`builder.go:415-438`, `project_dir` 1순위) 그대로이며 수리가 이 함수를 수정하지 않는다(plan-audit D1 정정 — v0.1.0의 「current_dir에서 유도」 서술은 판정서 산문 오류였다).

판정 명령: 기존 렌더 표시 테스트 전수 + 발산 입력 골든 비교 1건 추가. `go test ./internal/statusline/...`

**RED**: 없음(불변 대상 — 현재 상태가 곧 기준). **GREEN (M1 이후 유지)**: 기존 테스트 무파괴 + 발산 입력 포함 코퍼스 PASS로 입증. 표시 경로를 건드리는 mutant(spec.md §3 mutant 2)가 이 AC에서 걸린다.

### AC-SA-007 — throttle과 silent-failure가 보존된다

- **Given** write-if-changed skip과 best-effort 무소음(REQ-THRESHOLD-009/012)의 기존 테스트들이 있고,
- **When** 앵커 수리가 착지하면,
- **Then** 그 테스트들이 수정 없이(또는 경로 기대값만 갱신) 통과하고, 쓰기 실패가 렌더를 깨지 않는 의미론이 유지된다.

판정 명령: `go test ./internal/statusline/... ./internal/config/...`

**RED**: 없음(불변 대상). **GREEN (전 마일스톤 유지)**: 기존 테스트 통과 유지. throttle 제거 mutant(spec.md §3 mutant 3)가 이 AC에서 걸린다.

### AC-SA-008 — B5/B6이 불변이다

- **Given** 기준 트리 `0b1e27877`이 있고,
- **When** `git diff 0b1e27877..HEAD -- internal/hook/ internal/session/`을 보면,
- **Then** 변경 라인이 **0**이다.

판정 명령: `git diff 0b1e27877..HEAD --stat -- internal/hook/ internal/session/` (빈 출력 = 통과)

**RED**: 없음(불변 대상). **GREEN (전 마일스톤 유지)**: diff 0 유지. R1 시접이 반드시 이들을 건드려야 한다는 입증이 생기면 blocker 보고로 plan 확장 후 재판정한다(plan.md §D4).

### AC-SA-009 — 축 A 가드가 활성이다 (regression-guard)

- **Given** t422 가드(`internal/cli/todo_test.go:42` `runTodo` → `liveTodoQueueRootReason()`)가 HEAD에 존재하고,
- **When** todo 테스트 패밀리가 실행되면,
- **Then** 모든 테스트가 guarded 헬퍼(`runTodo`/`runTodoWithClosedStdin`)를 통해 실행되고, 가드가 live repository root에서 fail-loud(fatalf)함이 실측된다.

판정 명령: `go test ./internal/cli/ -run TestTodo -count=1` (전부 PASS) + 가드 활성 관측 테스트.

**RED-now**: 없다 — 기준 트리에서 이미 GREEN이다(판정서 A-6: 가드 코드 존재 + 9/3~9/7 신규 오염 0). **등급 regression-guard**: 채택 판정은 뮤턴트(AC-SA-011)가 대신한다. green path: M4가 실측 증거를 §E.2에 남긴다.

### AC-SA-010 — canary-HOME 스윕이 0 오염이다 (regression-guard)

- **Given** canary HOME(`t.TempDir()` 기반, 비-parallel `t.Setenv` 경유 — plan.md D5)에서,
- **When** `internal/cli`의 todo 테스트 패밀리 전체를 실행하면,
- **Then** canary `HOME/.moai/todo` 아래 신규 디렉터가 **0개**다.
- **And** swept count N이 명시적으로 보고되며 **N=0이면 테스트가 스스로 실패한다** — 셀렉터가 0개 테스트에 매치한 초록은 판정이 아니다(verification-completeness §1.1).

판정 명령: `go test ./internal/cli/ -run <todo-선택자> -count=1` — 선택자 표현식과 N은 M4에서 확정해 §E.2에 기록한다.

**RED-now**: 없다 — 이미 GREEN이다. **등급 regression-guard**. green path: M4가 (선택자, N, 오염 0) 삼인조를 관측·기록한다.

### AC-SA-011 — 가드의 무효화는 관측 가능하다 (뮤턴트 증명 · regression-guard의 채택 증거)

- **Given** 가드를 우회하는 뮤턴트 헬퍼(`newTodoCmd()`를 직접 Execute, `runTodo` 게이트 경유 안 함)가 있고,
- **When** 뮤턴트가 canary HOME에서 **`CLAUDE_PROJECT_DIR`가 비git temp 디렉터를 가리키는 컨텍스트**로 todo 명령을 실행하면(판정서 §A-5 기전 사슬 형태 — git 해석 실패 → HOME 폴백 → 큐 생성. git repo 맥락이면 git 해석이 fixture 리포로 가서 HOME 오염이 나지 않는다 / plan-audit D5),
- **Then** canary `HOME/.moai/todo` 아래 오염 디렉터가 **생성되는 것이 관측된다** — 이 관측이 "AC-SA-010의 0이 가드 덕분"임을 판별하는 유일한 증거다.
- **And** 오염을 만들지 못한 뮤턴트가 있으면 그 사실과 시도 형태를 함께 보고한다 — 그것이 가드의 **경계**(무엇까지 잡는가)를 그린다(REQ-SA-011).

판정 명령: 뮤턴트 관측 테스트 1건 + 결과 기록. 뮤턴트 헬퍼는 테스트 인프라로 남기지 않고 관측 후 제거하거나, 남길 경우 가드 우회임을 이름으로 드러낸다(`...GuardBypassMutant`).

**RED-now**: 해당 없음 — 뮤턴트 오염은 지금도 재현 가능하며, **그것이 본 AC의 GREEN 내용**이다. **등급 regression-guard이자 AC-SA-009/010의 채택 증거**(plan.md D9).

### AC-SA-012 — 상태 앵커는 단일 시접이다

- **Given** 수리가 착지한 트리가 있고,
- **When** 4멤버(B1 쓰기 / B2 board root / B3 goal 읽기 / B4 캐시 경로 소유자)의 앵커 획득 경로를 코드로 검사하면,
- **Then** 모두 단일 리졸버 함수(또는 그 반환값)를 거치고, 상태 쓰기·읽기 경로에서 `current_dir` 기반 앵커 유산(`input.CWD`/`os.Getwd` 앵커 후보)이 제거됐다 — `current_dir`는 기존 표시 유도(`extractProjectDirectory` — 불변)에만 남는다.

판정 명령: `grep -n "ProjectDir\|resolveStateAnchor" internal/statusline/context_usage.go` 등 멤버별 판독 + 표시 경로에 current_dir 잔존 확인. run-phase M5가 관측 방법(구체 grep 집합)을 확정해 §E.2에 기록한다.

**RED (현재)**: `resolveProjectDir`(`context_usage.go:278-291`)의 사슬이 `current_dir` → `input.CWD` → `os.Getwd()`로 상태 앵커를 결정 — 유산 존재. **GREEN (M1~M3 누적)**: 단일 시접 통일.

---

## §D.2 AC ↔ 판정서 대응표 (traceability)

| AC | 판정서 근거 |
|---|---|
| AC-SA-001 | §B-1 B1 행 + §B-2 (HEAD 격리 재현 — 실측 RED) + §B-3 (#1694 정합성: 226 stray = 렌더당 1레코드 기전의 직접 귀결) |
| AC-SA-002 | §B-1 B2 행 (`original_cwd` 있으면 옳음, 없으면 B1 사슬) |
| AC-SA-003 | §B-1 B3 행 (읽기 측 결함) + Residual-risk 3 (가시성 변화 경고) |
| AC-SA-004 | §B-1 B4 행 + Gaps (configDir 유도 지점 미확정 → M3 2단 RED) + t507 관측 형태 일치 |
| AC-SA-005 | 수리 방향 1 ("끝까지 못 찾으면 상태 쓰기 보류 — 무프로젝트=무상태") |
| AC-SA-006 | 수리 방향 1 ("표시용 basename은 종전대로") + Residual-risk 3 (읽기·쓰기·표시 관심 분리) |
| AC-SA-007 | `writeContextUsage` 주석 REQ-THRESHOLD-009/012 보존 + 수리 방향 1 (기존 writer 의미론 유지) |
| AC-SA-008 | §B-1 B5/B6 행 (경계 밖 분류) + 카드 scope discipline |
| AC-SA-009 | §A-6 (가드 코드 존재 `todo_test.go:42`) + Claim 2 (이미 수리) |
| AC-SA-010 | §A-6 (활발한 트리거 + 부재 = 고쳐짐 요건) + Claim 2 |
| AC-SA-011 | §A-5 (가드 기전 사슬) + Residual-risk 1 (runTodo/newTodoCmd 우회 표면) |
| AC-SA-012 | 수리 방향 1 (단일 시접 — "B1·B2·B3·B4를 이 시접으로 통일") |

범위 밖 항목의 판정서 근거: 생산 홈 폴백 가드 = Claim 4 + §A-4 표본 (spec.md §5 운영자 미결 결정), 오염 디렉터 343개 보존 = §"관련 자료" + 카드 [HARD], GH #1694 회신 = 수리 방향 3 (plan.md M5 sync-phase 태스크).
