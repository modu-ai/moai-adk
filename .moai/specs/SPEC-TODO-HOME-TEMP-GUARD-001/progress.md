# Progress — SPEC-TODO-HOME-TEMP-GUARD-001

카드 t536 · 트리 `.claude/worktrees/t536` · 브랜치 `WT-home-fallback` · plan-phase base `412c8cb14`.

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier S — 카드 지시에 따라 `acceptance.md` 별도 작성; `design.md`/`research.md` 없음).
- SPEC ID 정규식 자가 점검: `SPEC-TODO-HOME-TEMP-GUARD-001` → `PASS` (Bash 실행, 이 트리).
- 운영자 결정 반영: 임시-디렉터 기원 **한정** 거부 (「모든 비git base」 판독은 기각 — spec.md §3, plan.md D1).
- 미결 사항: **없다.** U1(거부 시 CLI 형태)은 운영자 선택으로 **(b) 안내 후 계속**으로 결정됐다(2026-09-07, 리드 경유 — plan.md §D). run-phase는 고르지 않고 구현한다.
- 2차 감사 수리(`version: 0.1.2`): D9(대체 루트를 `base`로 정정 — 종전 값은 루트가 아니라 상태 디렉터였다), D10(AC-THG-001이 문자열 동일성이 아니라 `BacklogPathForRoot(반환 루트)`의 읽힘을 단언), D11(`/var/tmp` 근거의 미측정 단정 격하 + 재검토 트리거 교체), U1 결정 반영. spec.md §8에 기존 코드(`todo_root.go:121`)의 계층 불일치를 **관측**으로 등재 — 본 SPEC은 고치지 않는다(별도 카드 후보).

## §E.2 Run-phase Evidence

측정 트리: `.claude/worktrees/t536` · 브랜치 `WT-home-fallback` · M1 착지 `8337cdf24`. 아래 모든 실행은 **이 트리, 이 회차**의 것이며, 플랫폼은 별도 표기가 없으면 **로컬 darwin/arm64**다. `go test`는 트리에서 컴파일되므로 판정 빌드와 측정 트리가 같다(별도 설치본 개입 없음).

### M1 (착지분 요약 — 재측정 아님)

`internal/kanban/temp_origin.go`의 `TempOriginReason` + `TempRootsFn` 이음매. 호출-시점 판독 증거는 `.moai/reports/t536/m1-mutant-red.txt`, 뮤턴트 픽스처는 `.moai/reports/t536/mutants/`. M2는 이 이음매를 소비할 뿐 형태를 바꾸지 않았다.

### M2 — 배선 + 안내

**생산 변경 2파일**: `internal/kanban/todo_root.go`(술어 삽입 + `TempOriginRefusal` 노출), `internal/cli/todo.go`(`PersistentPreRun` 안내 1함수). 접촉 금지 패키지(`internal/statusline`·`config`·`hook`·`session`·`stateanchor`) 변경 0.

**술어 배치**: `primaryCheckoutRoot` 직후, `homeTodoQueueRoot`의 `ok` 검사 **앞**(D16). 두 리졸버 모두 동일 위치.

#### RED 실측 ① — AC-THG-001 (a)(b)(c)

배선 두 곳을 제거한 뮤턴트(파일 컴파일은 유지)로 M2 이전 동작을 재현해 관측. 전문: `.moai/reports/t536/m2-red-ac-thg-001.txt`.

```
$ go test ./internal/kanban/ -run 'TestTodoQueueRoot_TempOriginRefusesHomeQueue|TestTodoQueueRoot_PureGuardIsSilent' -count=1
--- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue/no_local_queue... 
    todo_root_temp_guard_test.go:64: pure resolver returned the home queue root ".../002/.moai/todo/001-4f02f775" for a temporary origin
    todo_root_temp_guard_test.go:75: pure resolver: BacklogPathForRoot(returned) = ".../002/.moai/todo/001-4f02f775/.moai/state/todo/backlog.json", want the local canonical ".../001/.moai/state/todo/backlog.json"
--- FAIL: .../local_queue_present...
    todo_root_temp_guard_test.go:92: adopting root = ".../002/.moai/todo/001-0c66f22b", want the launch base ".../001"
--- FAIL: .../temporary_origin_AND_home_unresolvable...
    todo_root_temp_guard_test.go:134: temp-origin ∧ home-unresolvable root = ".../001/.moai/state/todo", want the launch base ".../001"
```

세 번째 줄이 D16이 예고한 형태를 **실측으로** 보여준다 — 술어를 뒤에 두면 반환이 `base/.moai/state/todo`가 된다.

#### RED 실측 ② — C행 새 판정식 (M2 종료 조건의 「RED 실연」)

C행 사본 1건(`..._PureFallbackWritesNothing_NonTemp`)에서 이음매 스텁만 제거하고 실행. 전문: `.moai/reports/t536/m2-red-crow-predicate.txt`.

```
=== RUN   TestResolveTodoQueueRoot_PureFallbackWritesNothing_NonTemp
    todo_root_nontemp_copy_test.go:60: the injected temp-root set was not read: base ".../001" still classifies temporary (reason "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/") — this copy would exercise the guard's refusal rather than the home-fallback branch it was written for
--- FAIL: TestResolveTodoQueueRoot_PureFallbackWritesNothing_NonTemp (0.00s)
=== RUN   TestResolveTodoQueueRoot_PureFallbackWritesNothing
--- PASS: TestResolveTodoQueueRoot_PureFallbackWritesNothing (0.05s)
```

같은 실행에서 **원본은 PASS**한다. 이것이 D19가 요구한 판별력의 직접 증거다 — 원본의 단언 집합으로는 두 상태를 구별할 수 없고, 새 판정식만이 붉어진다.

#### GREEN 실측

전문: `.moai/reports/t536/m2-green-kanban-web.txt`(23+6건 PASS), `.moai/reports/t536/m2-green-cli.txt`(10건 PASS, 1건 SKIP).

| AC | 산출 테스트 | 결과 |
|---|---|---|
| AC-THG-001 (a)(b)(c) | `TestTodoQueueRoot_TempOriginRefusesHomeQueue` (3 서브테스트) | PASS |
| AC-THG-002 | `TestTempOrigin_SymlinkSpellingEquivalence` | PASS (M1 산출, 회귀 없음) |
| AC-THG-003 | `TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback` | PASS |
| AC-THG-004 | `TestTempOrigin_FailsOpenOnUnresolvable` | PASS (M1 산출) |
| AC-THG-005 | `TestTodoQueueRoot_PureGuardIsSilent` + `TestTempOriginGuidance_NamesRootsAndContinues` + `TestTempOriginGuidance_SilentOnNonTemporaryBase` | PASS |
| AC-THG-006 | `TestTempOrigin_ComponentBoundary` | PASS (M1 산출) |
| AC-THG-007 | (M3 소관 — 뮤턴트 증명) | 미판정 |
| AC-THG-008 | `TestTodoQueueRoot_GitBranchUnreachedByGuard` + `TestResolveTodoQueueRoot_PureFallbackWritesNothing(_NonTemp)` + `find ~/.moai/todo -maxdepth 1 -type d \| wc -l` → **344** (§C 기준값과 동일) | PASS |

AC-THG-005 Then절 3항 전수 판정: ⑴ 매치된 임시 루트 이름 부름 — `strings.Contains(errOut, matched)`, ⑵ 대체 루트 이름 부름 — `strings.Contains(errOut, dir)`, ⑶ canary HOME 디렉터 0 — `os.ReadDir` 계수. 종료 코드 0 + 카드가 project-local 큐에 실제로 착지함까지 함께 단언(U1 = (b)).

#### §C.1 14행 처분 이행

| 행 | 테스트 | 처분 | 이행 |
|---|---|---|---|
| A1 | `kanban.TestResolveTodoQueueRoot_FallbackNoGit` | 보존 이관 | `declareNonTemporary(t)` 추가, **홈 루트 단언 원문 유지** — PASS |
| A2 | `kanban.TestResolveTodoQueueRoot_PopulatedFallbackWins` | 보존 이관 | 동일 — PASS |
| A3 | `kanban.TestResolveTodoQueueRootAdopting_AdoptsLocalQueue` (AC-WTQ-008) | 보존 이관 (필수) | 동일, adopt-not-shadow 단언 원문 유지 — PASS |
| A4 | `cli.TestResolveTodoQueueRoot_FallbackNoGit` | 보존 이관 | `declareNonTemporaryQueueBase(t)` + `assertQueueSeamHeard(t, dir)` — PASS |
| A5 | `cli.TestTodoQueue_FallbackAdoptsExistingLocalQueue` | 보존 이관 | 동일 — PASS |
| A6 | `cli.TestGuardBypassMutant_ObserveHomePollution` (AC-SA-011) | **가지 이행** | REQ-SA-011 2번째 가지로 전환: 오염 부재를 관측하고 층 경계를 보고. 공허화 방지를 위해 ⑴ base가 임시 기원임 ⑵ 반환 루트 = 대체 루트 ⑶ 카드가 project-local 큐에 실제 착지 를 **양성 단언**. `t.Logf`가 M3 경계 보고 경로를 지목 — PASS |
| B1 | `kanban.TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing` | 의도적 갱신 | 단언을 `dir/.moai/state/todo` → `dir`로 갱신. **근거는 REQ-THG-001의 교차 고정(D16)이며 편의가 아니다** — 종전 값은 소비자가 `BacklogPathForRoot`를 덧붙이는 루트보다 한 계층 아래다. 무쓰기 절반은 보존(+`dir/.moai` 부재까지 확장) — PASS |
| C1 | `kanban.TestResolveTodoQueueRoot_PureFallbackWritesNothing` | 비임시 사본 | `..._NonTemp` 신설, 원본 유지. 판별식 직접 단언 포함 — PASS |
| C2 | `kanban.TestResolveTodoQueueRoot_ReadThroughToProjectLocal` | 비임시 사본 | 동일 — PASS |
| C3 | `kanban.TestAdoptionLandsWhereConsumersRead` | 비임시 사본 | 동일 — PASS |
| C4 | `kanban.TestAdoptingAndPureResolversAgreeWhenAdoptionFails` | 비임시 사본 | 동일 — PASS |
| C5 | `web.TestTodoSectionReadsThroughToProjectLocalQueue` | 비임시 사본 | `..._NonTemp` 신설(세 번째 패키지). `kanban.TempRootsFn` exported 요건이 여기서 실제로 구속됐다 — PASS |
| C6 | `cli.TestTodoQueueRootGuard_SilentOnHomeFallbackFixture` | 비임시 사본 | `..._NonTemp` 신설. 사본은 **홈 폴백 루트임을 먼저 단언**한 뒤 침묵을 판정 — PASS |
| C7 | `cli.TestAxisACanaryHomeSweep_TodoFamily` | **수용된 손실** | 사본 구성 불가(`exec.Command` 자식 프로세스는 패키지 수준 이음매를 물려받지 않는다). 테스트는 **한 줄도 수정하지 않았다**; 기본 실행에서 SKIP. 6건에 합산하지 않았다 — 사본 처분 6, 공허화 7 |

계수 확인: A행 6(보존 5 + 가지 이행 1) · B행 1 · C행 7(사본 6 + 수용 손실 1) = **14**.

#### 형제 SPEC diff 0

```
$ git diff --stat 8337cdf24..HEAD -- .moai/specs/SPEC-WEB-TODO-QUEUE-001 .moai/specs/SPEC-STATE-ANCHOR-001
(출력 없음)
$ git diff --stat 412c8cb14 -- .moai/specs/SPEC-WEB-TODO-QUEUE-001 .moai/specs/SPEC-STATE-ANCHOR-001
(출력 없음 — 워킹 트리 포함)
```

#### 패키지 판정

```
$ go test ./internal/kanban/ ./internal/web/ -count=1
ok  github.com/modu-ai/moai-adk/internal/kanban  142.724s
ok  github.com/modu-ai/moai-adk/internal/web     3.778s
$ go vet ./internal/kanban/ ./internal/web/ ./internal/cli/   → 무출력
```

`internal/cli` 전체 스위트 1차 실행은 **1건 FAIL**했다: `TestTodoStoreClaims_NoStaleGoSourceSite`가 `todo_root.go`의 새 주석에 들어간 큐 경로 리터럴을 잡았다(비-test Go 소스에 그 문자열을 두지 않는 가드). 주석을 리터럴 없이 재서술해 해소했고, 재실행 결과는 아래 「미검증/잔여」 참조. **이 FAIL은 상속 레드가 아니라 본 카드가 만든 것이며, 가드의 판정이 옳았다.**

#### 미검증 (Gaps)

- **리눅스 셀 미측정.** 위 전부 로컬 macOS(darwin/arm64) 판정이다. `/tmp`가 실디렉터이고 `/private/tmp`가 없는 환경에서 AC-THG-002가 어떻게 판정되는지는 이 트리에서 재지 않았다 — CI linux 매트릭스의 결과를 인용해야 한다(AC-THG-002 리눅스 셀은 CI 몫).
- **AC-THG-007 미판정** — M3 소관.
- **`.moai/reports/t536/guard-boundary.md` 미작성** — M3 소관. A6의 `t.Logf`가 그 경로를 지목하되, 파일 존재는 단언하지 않는다(M2에서 단언하면 M3 착지 전까지 상시 적색).
- **크로스 플랫폼 빌드(`GOOS=windows`) 미실행** — M3 배치 검증 몫.
- **`internal/cli` 전체 스위트 재실행 판정** — 수리 후 재실행분의 최종 결과는 커밋 시점 기준으로 확인했으며, 판정은 CI가 소유한다.

#### 잔여 위험 (Residual risk)

- **거짓 양성 잔여**: 「임시 루트 아래에 사는 비git 진짜 프로젝트」는 여전히 홈 폴백을 잃는다. spec.md §4가 수용한 위험이며, REQ-THG-006 안내가 회복 경로를 알린다. `TestTempOrigin_LexicalResembler`가 *어휘적으로만 닮은* 경로는 오분류되지 않음을 대조군과 함께 고정했지만, **진짜로 임시 루트 안에 사는 프로젝트**는 정의상 이 방어의 대상이 아니다.
- **이중 앵커 해석**: M1의 두-앵커 대조(정규화형 + 어휘적 `Clean`)는 REQ-THG-003의 「한 규칙으로 양쪽 정규화」에 대한 **해석**이며, 리드가 2026-09-08에 채택했다. 요구사항 텍스트는 수정하지 않았다. 그 해석이 요구사항 범위 안인지는 **감사가 독립적으로 판정할 사항**으로 넘긴다.
- **`PersistentPreRun` 배치**: cobra는 체인에서 가장 가까운 `PersistentPreRun` 하나만 실행한다. 현재 루트 커맨드에 `PersistentPreRun`이 없어(`/usr/bin/grep -rn 'PersistentPreRun' internal/ --include='*.go' | grep -v _test.go` → `root.go:127`의 `WorktreeCmd.PersistentPreRunE` 1건뿐) 가려지는 것이 없다. 루트에 나중에 하나가 생기면 이 안내가 그것을 가린다 — 그때 재검토가 필요하다.
- **`internal/cli/deps.go`의 gofmt 미정렬**은 본 카드 이전부터 있던 것이다(`git diff -- internal/cli/deps.go` 무출력). 손대지 않았다.

### M3 — 뮤턴트 증명 + 배치 검증 (2026-09-08)

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536` @ `db45d8209`
(브랜치 `WT-home-fallback`). 모든 배치의 첫 명령으로 트리를 읽었고, 그 출력이 각
배치 파일 머리에 있다 — 셸 CWD 가 조용히 이동하면 **엉뚱한 트리에서 우연히
성공하는 명령**이 그럴듯한 값을 내기 때문이다.

```
$ git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536
db45d8209
WT-home-fallback
```

아래 모든 인용은 반출된 파일을 지목한다. 반출하지 않은 자료(세션 스크롤백,
셸 히스토리)는 근거 자리에 두지 않았다.

#### 뮤턴트 증명 (AC-THG-007)

전문 보고는 **`.moai/reports/t536/guard-boundary.md`** (plan.md §F M3 가 요구하는
산출물, 세 항목 충족). 요약과 반출 경로만 여기 남긴다.

**뮤턴트 E — 판별식 무력화.** `TempOriginReason` 첫 문장에 `return "", false` 한
줄을 넣어 항상 「임시 아님」을 보고하게 했다. 형태: `.moai/reports/t536/mutants/mutantE.go.txt`.

```
$ go test ./internal/kanban/ -count=1 -vet=off -run 'TestTodoQueueRoot_TempOriginRefusesHomeQueue'
--- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue (0.08s)
    --- FAIL: .../no_local_queue:_both_resolvers_return_the_launch_base (0.00s)
        todo_root_temp_guard_test.go:65: precondition: base ".../001" must classify temporary under the production root set (reason "")
    --- FAIL: .../local_queue_present:_the_returned_root_is_where_that_queue_is_read (0.04s)
        todo_root_temp_guard_test.go:105: adopting root = ".../002/.moai/todo/001-7b9f3e87", want the launch base ".../001"
    --- FAIL: .../temporary_origin_AND_home_unresolvable:_the_base_still_wins (0.04s)
        todo_root_temp_guard_test.go:149: temp-origin ∧ home-unresolvable root = ".../001/.moai/state/todo", want the launch base ".../001"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.555s
```

출력 전문: `.moai/reports/t536/m3/mutantE-kanban.txt`. AC-THG-001 세 갈래가 모두
잡았고, 갈래 (b) 의 반환 루트가 canary HOME 아래로 되돌아간 것이 오염 경로가
열렸다는 뜻이다.

**못 잡은 자리도 함께 보고한다(REQ-SA-011 두 번째 가지의 요구).** 같은 뮤턴트를
`internal/cli` 게이트 우회 테스트에 걸면 **오염 단언에 도달하지 못하고 전제에서
멈춘다** (`.moai/reports/t536/m3/mutantE-cli-bypass.txt`):

```
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.00s)
    todo_axisa_guard_test.go:184: precondition: ".../002" must classify as a temporary origin (reason ""); without it, an absence of pollution says nothing about the resolver-layer guard
```

결함이 아니라 전제 가드가 제 일을 한 것이다 — 부재 단언은 base 가 실제로 임시
기원으로 분류될 때만 무언가를 말한다. **귀결: 뮤턴트 E 단독으로는 오염 재현을
실연할 수 없다.** 그래서 대조군은 route (i) 다.

#### 대조군 — route (i) pre-guard 등가 (이 트리에서 재측정)

`412c8cb14` 을 체크아웃하지 않았다(브랜치 상태 변경 금지, 이 트리는 앵커됨).
대신 **M2 배선 두 곳만 되돌려** pre-guard 등가 상태를 이 트리에 만들고 같은
테스트를 돌렸다 — 선행 측정의 인용이 아니라 이번 실행의 측정이다. 형태:
`.moai/reports/t536/mutants/route-i-preguard.patch.txt`.

```
$ go test ./internal/cli/ -count=1 -vet=off -timeout 900s -run 'TestGuardBypassMutant_ObserveHomePollution'
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.12s)
    todo_axisa_guard_test.go:202: the mutant produced home pollution under /var/folders/.../001/.moai/todo (1 entr(ies)) — the resolver-layer temporary-origin guard regressed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.984s
```

**오염이 재현됐다 — canary HOME 아래 1건.** 트리 스탬프 4지점을 포함한 전문:
`.moai/reports/t536/m3/batchB-stamped-routei.txt` (1차 실행분은
`.moai/reports/t536/m3/routei-cli-bypass.txt`). 같은 상태의 AC-THG-001 RED 전문은
`.moai/reports/t536/m3/routei-kanban.txt`.

두 실행의 차이는 **가드 하나**다:

| 실행 | 판별식 | M2 배선 | 게이트 우회 뮤턴트 |
|---|---|---|---|
| 복원된 트리 `db45d8209` | 살아 있음 | 있음 | 오염 **0** + 카드가 project-local 큐에 착지 (PASS) |
| route (i) pre-guard 등가 | 살아 있음 | **없음** | 오염 **1건** under canary `$HOME/.moai/todo` (FAIL) |
| 뮤턴트 E | **죽음** | 있음 | 전제에서 정지 — 오염 단언 미도달 |

첫 두 행이 대조군이다. 테스트도 트리도 기계도 같고 배선 유무만 다르므로, 첫 행의
0 은 가드가 만든 것이지 뮤턴트가 돌지 않아서가 아니다.

#### 복귀 확인 — 뮤턴트는 트리에 남지 않았다

무출력을 증거로 쓰므로 **그 명령이 출력을 낼 수 있는 상태였음**을 대조로 남긴다.

```
# 주입 중 — 같은 명령이 실제로 diff 를 뱉었다
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go | wc -c
     731

# 복원 후 — 같은 명령, 무출력
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go | wc -c
       0

$ git --no-optional-locks status --porcelain
?? .moai/reports/t536/guard-boundary.md
?? .moai/reports/t536/m3/
?? .moai/reports/t536/mutants/mutantE.go.txt
?? .moai/reports/t536/mutants/route-i-preguard.patch.txt
```

반출: 주입 중 diff `.moai/reports/t536/m3/routei-injected.diff`, 복원 후 diff
`.moai/reports/t536/m3/routei-restored.diff`. `??` 항목은 M3 가 의도적으로 추가한
증거이며 추적 파일 수정은 0. HEAD 는 주입 전·후·복원 전·후 모두 `db45d8209`.

**뮤테이션 창 위생**: 주입 직전 배타성을 확인했고(추적 수정 0), 주입한 파일은 같은
턴 안에서 복원했다. 이 창 안에서 컴파일된 바이너리는 귀속 불가이므로 설치·재사용
하지 않았다 — `make install` / `cp ~/go/bin/moai` 계열은 이 카드에서 한 번도
실행하지 않았다.

#### AC 매트릭스 전수 판정 (8건)

근거 출력: `.moai/reports/t536/m3/batchC1-stamped.txt` (kanban verbose),
`.moai/reports/t536/m3/batchC2-stamped.txt` (cli verbose), 뮤턴트 근거는 위 참조.

| AC | 요구사항 | 판정 | 근거 테스트 / 명령 | 관측 |
|---|---|---|---|---|
| AC-THG-001 | REQ-THG-001 | **PASS** | `TestTodoQueueRoot_TempOriginRefusesHomeQueue` (갈래 a/b/c 서브테스트 3건) | `--- PASS` ×1 + 서브 3건 전부 PASS (0.19s) |
| AC-THG-002 | REQ-THG-002·003 | **PASS (darwin) / 리눅스 셀 미측정** | `TestTempOrigin_SymlinkSpellingEquivalence` | `--- PASS` (0.00s). 리눅스 셀은 Gaps |
| AC-THG-003 | REQ-THG-005·009 | **PASS** | `TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback` + `TestTempOriginGuidance_SilentOnNonTemporaryBase` | `--- PASS` (0.09s) / `--- PASS` (0.12s) |
| AC-THG-004 | REQ-THG-004 | **PASS** | `TestTempOrigin_FailsOpenOnUnresolvable` (서브 5건) | `--- PASS` + 서브 5건 전부 PASS |
| AC-THG-005 | REQ-THG-006·007 | **PASS** | `TestTempOriginGuidance_NamesRootsAndContinues` + `TestTodoQueueRoot_PureGuardIsSilent` | `--- PASS` (0.17s) / `--- PASS` (0.08s) |
| AC-THG-006 | REQ-THG-002 | **PASS** | `TestTempOrigin_ComponentBoundary` (+ `/tmpfoo` 서브) | `--- PASS` + 서브 PASS |
| AC-THG-007 | REQ-THG-001 (판별 증거) | **PASS** | 뮤턴트 E + route (i) 대조군; 산출물 `guard-boundary.md` | 뮤턴트 E 가 AC-THG-001 3갈래 FAIL; route (i) 가 오염 1건 재현 |
| AC-THG-008 | REQ-THG-007·008 | **PASS** | `TestTodoQueueRoot_GitBranchUnreachedByGuard` + `TestResolveTodoQueueRoot_PureFallbackWritesNothing` + 셸 계수 2건 | `--- PASS` (0.12s) / `--- PASS` (0.05s); `find ~/.moai/todo` = **344** (§C 기준값과 동일); D8 카드 귀속 diff 0 |

#### 배치 검증 — 무변경·계수·빌드·정적분석

전문: `.moai/reports/t536/m3/batchA-stamped.txt`.

```
$ find ~/.moai/todo -maxdepth 1 -type d | wc -l
     344                       # §C 기준값 344 와 동일 (REQ-THG-008)

$ go vet ./internal/kanban/ ./internal/web/ ./internal/cli/
                             # 무출력, exit=0

$ GOOS=windows GOARCH=amd64 go build ./...
                             # 무출력, exit=0
```

**D7 — 형제 SPEC 산출물 diff 0**, 그리고 그 판정이 공허하지 않다는 catch-all 대조:

```
$ git diff --stat 412c8cb14..HEAD -- .moai/specs/SPEC-WEB-TODO-QUEUE-001 .moai/specs/SPEC-STATE-ANCHOR-001
                             # 무출력
$ git ls-files .moai/specs/SPEC-WEB-TODO-QUEUE-001 .moai/specs/SPEC-STATE-ANCHOR-001 | wc -l
       9                     # 필터가 닿은 집합이 비어 있지 않다 — 무출력이 의미를 갖는다
```

**D8 — 접촉 금지 경로. 명령을 문자 그대로 돌린 결과는 무변경이 아니다. 이것을 FINDING 으로 보고한다.**

```
$ git diff 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/ | wc -c
   18650                     # 무변경 아님
$ git diff --stat 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/
 internal/config/defaults.go                        |   1 +
 .../config/interview_recommendation_mode_test.go   | 125 +++++++++++++++++
 .../config/testdata/shipped_key_inventory.yaml     |   5 +-
 internal/config/types.go                           |  23 +++-
 internal/stateanchor/stateanchor.go                |  42 ++++--
 internal/stateanchor/stateanchor_test.go           | 150 +++++++++++++++++++--
 6 files changed, 318 insertions(+), 28 deletions(-)
```

전문 diff: `.moai/reports/t536/m3/d8-literal.diff`.

**원인 — 기준점이 낡았다, 카드가 범위를 넘은 것이 아니다.** `412c8cb14` 는 이 카드의
plan-phase 측정 기준 트리인데, 그 뒤 `6b71fdaa5` 에서 develop 을 흡수했다. 따라서
`412c8cb14..HEAD` 범위에는 **다른 카드들의 작업이 들어 있다**. 그 6개 파일을 만든
커밋을 전수 열거하면 전부 형제 카드다:

```
$ git log --oneline 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/
708e89955 Merge remote-tracking branch 'origin/develop' into WT-resolve-validate
5df939476 feat(SPEC-STATE-ANCHOR-VALIDATE-001): M2 GREEN — validate anchor candidates, fall through on failure (t537)
3b39eef8f feat(SPEC-STATE-ANCHOR-VALIDATE-001): M1 committed RED — anchor validation tests (t537)
12c7782f8 merge: absorb origin/develop 52f863f36 into WT-analysis-pull (card t401)
8b4896bcb chore(t401): raise AlwaysLoadedTokenBudget 76400 -> 78500 with justification
099c7bbe4 feat(t401): M3 — interview.recommendation_mode config key (default push)
```

**카드 귀속 측정 — 이 카드의 두 커밋(M1 `8337cdf24`, M2 `db45d8209`)은 그 경로를 한 글자도 건드리지 않았다:**

```
$ git diff 6b71fdaa5..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/ | wc -c
       0
```

전문(빈 파일): `.moai/reports/t536/m3/d8-card.diff`. D8 이 방어하려는 성질 —
「이 카드가 접촉 금지 경로를 건드리지 않았다」 — 는 **성립한다**. 문자 그대로의
명령이 실패한 것은 그 명령이 **이동하는 기준점 대신 고정 SHA 를 쓰되, 그 고정 SHA
가 develop 흡수 이전이라서** 범위에 남의 작업이 섞였기 때문이다. plan.md §F 의
D8 문구는 흡수를 예견하지 못했다 — 후속 정정 대상으로 §E.2 잔여 위험에 남긴다.

#### 패키지 판정 (재측정, 복원된 트리)

```
$ go test ./internal/kanban/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	144.092s
$ go test ./internal/web/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	3.896s
$ go test ./internal/cli/ -count=1 -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	459.946s
```

전문: `.moai/reports/t536/m3/batchC1-stamped.txt`, `.moai/reports/t536/m3/batchC2-stamped.txt`.
M2 가 만들었던 `TestTodoStoreClaims_NoStaleGoSourceSite` FAIL 은 해소된 채로 유지된다
(위 `internal/cli` 판정이 그 재실행분이다). **`go test ./...` 는 돌리지 않았다** —
저장소 규율(전체 스위트 로컬 실행 금지), 전 패키지 판정은 CI 몫이다.

#### 미검증 (Gaps) — M3

- **리눅스/윈도우 셀 미측정.** 위 전부 darwin/arm64 단일 셀이다. AC-THG-002 의
  리눅스 셀(`/tmp` 이 실디렉터리이고 `/private/tmp` 이 없는 환경)은 이 기계에서 잴
  수 없다. **로컬 macOS PASS 를 리눅스 판정으로 재사용하지 않는다** — CI 몫이며,
  레인은 push 하지 않으므로 이 카드 안에서는 인용 불가다.
- **`GOOS=windows` 크로스빌드는 테스트 파일을 컴파일하지 않는다.** `go build` 는
  non-test 패키지만 본다. 따라서 windows 빌드 exit=0 은 이 카드의 **테스트**가
  windows 에서 컴파일된다는 근거가 **아니다**.
- **전 패키지 스위트 미실행** — `internal/kanban` / `internal/web` / `internal/cli`
  3개만 돌렸다. 나머지 패키지 판정은 CI 몫이다.
- **`golangci-lint` 미실행** — `go vet` 만 돌렸다. 린트 판정은 CI 몫이다.
- **`SPEC-STATE-ANCHOR-001` M4 기록(route (ii))은 인용하지 않았다.** route (i) 가
  성립했으므로 선행 측정을 대조군으로 재사용할 이유가 없다.
- **D8 문자 그대로의 명령은 무변경이 아니다** — 위 FINDING 참조. 카드 귀속 범위로
  재측정해 성질은 확인했으나, plan.md §F 의 D8 문구 자체는 정정하지 않았다(본
  카드는 plan.md 본문을 수정할 소관이 아니다).

#### 잔여 위험 (Residual risk) — M3

- **대조군이 테스트 한 건에 걸려 있다.** `TestGuardBypassMutant_ObserveHomePollution`
  이 나중에 약화되면(특히 긍정 귀속 단언 — 카드가 실제로 착지했는지 세는 부분 —
  이 빠지면) 이 문서가 인용하는 「오염 0」의 의미도 함께 약해진다.
- **route (i) 는 pre-guard 등가이지 `412c8cb14` 그 자체가 아니다.** 등가성의 근거는
  「M2 가 추가한 것이 정확히 그 두 블록」이라는 `db45d8209` 의 diff 다. 그 사이 제3의
  경로로 가드가 들어왔다면 등가성이 깨지지만, 같은 diff 가 그것을 배제한다.
- **뮤턴트 E 는 게이트 우회 축에서 아무것도 말하지 못한다**(전제에서 정지). 판별식
  축의 회귀 방어는 AC-THG-002/004/006 과 M1 뮤턴트 A/C/D 가 담당한다.
- **D8 기준점 노후화는 재발한다.** 고정 SHA 를 기준으로 쓴 무변경 판정은 그 SHA 가
  흡수 이전이면 범위에 남의 작업이 섞인다. 후속 카드에서 D8 류 판정식은 **카드
  귀속 범위**(흡수 병합 이후)로 쓰는 편이 옳다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: pending-backfill        # 이 커밋은 자기 해시를 인용할 수 없다 (D3 백필 창)
run_status: audit-ready

ac_pass_count: 8
ac_fail_count: 0
ac_pass_with_debt: 1                    # AC-THG-002 — darwin PASS, 리눅스 셀은 CI 몫 (Gaps)

preserve_list_post_run_count: 14        # plan.md §C.1 영향 테스트 14건 전수, 처분 이행 후 전부 PASS

l44_pre_commit_fetch: "git fetch origin develop → exit 0; git rev-list --count --left-right origin/develop...HEAD → 48\t2"
l44_post_push_fetch: "n/a — 레인은 push 하지 않는다 (develop push 는 리드 일괄 소관)"

new_warnings_or_lints_introduced: 0     # go vet 3패키지 무출력; golangci-lint 는 미실행 (CI 몫)

cross_platform_build:
  darwin_arm64: "go build ./... → exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... → exit 0"
  caveat: "go build 는 테스트 파일을 컴파일하지 않는다 — 테스트의 windows 컴파일 가능성은 미검증"

total_run_phase_files: 43               # M1+M2 커밋 21 + M3 증거·산출물 21 + progress.md 1
m1_to_mN_commit_strategy: "마일스톤별 개별 커밋 (M1 8337cdf24 · M2 db45d8209 · M3 이 커밋); squash 없음, --amend 없음, force-push 없음"

evidence_root: .moai/reports/t536/      # 모든 인용은 이 추적 경로의 반출 파일을 지목한다
guard_boundary_report: .moai/reports/t536/guard-boundary.md
measurement_tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536
measurement_head: db45d8209
measurement_branch: WT-home-fallback
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
