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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
