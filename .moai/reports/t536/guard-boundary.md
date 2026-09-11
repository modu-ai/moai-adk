# 가드 경계 보고 — SPEC-TODO-HOME-TEMP-GUARD-001 (카드 t536, M3)

측정 트리: `.claude/worktrees/t536` @ `db45d8209` (브랜치 `WT-home-fallback`).
모든 수치와 출력은 **이 실행에서, 이 트리를 상대로** 관측했다.
아래 모든 인용은 §7 의 반출 파일을 지목한다 — 반출하지 않은 자료는 근거 자리에 두지 않았다.

이 파일은 `plan.md` §F M3 이 요구하는 산출물이며, `SPEC-STATE-ANCHOR-001`
REQ-SA-011 의 두 번째 가지 — 「오염을 만들지 못한 뮤턴트는 그 사실과 그것이
드러내는 가드 경계와 함께 보고하라」 — 를 이행한다. 커밋 메시지나 세션 로그가
아니라 카드 증거 경로의 파일이어야 하는 이유는 단순하다: 세션이 끝나면 로그는
읽히지 않고, 감사자가 읽는 자리는 여기다.

---

## 1. 어느 뮤턴트인가

두 개를 주입했다. 하나는 판별식을 죽이고, 다른 하나는 배선을 걷어낸다.
둘을 나눈 것이 이 보고의 핵심 결과를 만들었다(§3).

### 뮤턴트 E — 판별식 무력화

`internal/kanban/temp_origin.go` 의 `TempOriginReason` 첫 문장에 한 줄을 넣어
**항상 「임시 아님」을 보고**하게 했다. 나머지 본문은 손대지 않았다.

```go
 func TempOriginReason(base string) (reason string, isTemp bool) {
+	return "", false // MUTANT E (t536 M3): discriminant disabled — NEVER temporary
 	if strings.TrimSpace(base) == "" {
```

보존된 픽스처: `.moai/reports/t536/mutants/mutantE.go.txt`.

`-vet=off` 를 준 것은 편의가 아니다. 남은 본문이 도달 불가가 되므로, vet 실패가
테스트 실패를 대신해 서 있는 상태를 만들지 않기 위해서다 — 그러면 무엇이
잡았는지 귀속할 수 없다. (기본 `go test` vet 부분집합에 unreachable 은 없으므로
없어도 돌기는 한다. 끄는 쪽이 관측을 깨끗하게 만든다.)

**명령**

```
go test ./internal/kanban/ -count=1 -vet=off -run 'TestTodoQueueRoot_TempOriginRefusesHomeQueue'
```

**출력 전문** (`exit=1`)

```
--- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue (0.08s)
    --- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue/no_local_queue:_both_resolvers_return_the_launch_base (0.00s)
        todo_root_temp_guard_test.go:65: precondition: base "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoQueueRoot_TempOriginRefusesHomeQueueno_local_queue_bot3312424363/001" must classify temporary under the production root set (reason "")
    --- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue/local_queue_present:_the_returned_root_is_where_that_queue_is_read (0.04s)
        todo_root_temp_guard_test.go:105: adopting root = "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoQueueRoot_TempOriginRefusesHomeQueuelocal_queue_present1307054495/002/.moai/todo/001-7b9f3e87", want the launch base "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoQueueRoot_TempOriginRefusesHomeQueuelocal_queue_present1307054495/001"
    --- FAIL: TestTodoQueueRoot_TempOriginRefusesHomeQueue/temporary_origin_AND_home_unresolvable:_the_base_still_wins (0.04s)
        todo_root_temp_guard_test.go:149: temp-origin ∧ home-unresolvable root = "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoQueueRoot_TempOriginRefusesHomeQueuetemporary_origin_AN2633523154/001/.moai/state/todo", want the launch base "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoQueueRoot_TempOriginRefusesHomeQueuetemporary_origin_AN2633523154/001"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.555s
FAIL
```

AC-THG-001 세 갈래가 모두 잡았다. 갈래 (b) 의 반환 루트가 canary HOME 아래
`.moai/todo/001-7b9f3e87` 로 되돌아간 것이 오염 경로가 열렸다는 뜻이다.

### 뮤턴트 E 가 **못 잡힌** 자리 — 함께 보고한다

같은 뮤턴트를 `internal/cli` 의 게이트 우회 테스트에 걸면, 그 테스트는
**오염 단언에 도달하지 못하고 전제(precondition)에서 멈춘다.**

**명령**

```
go test ./internal/cli/ -count=1 -vet=off -timeout 900s -run 'TestGuardBypassMutant_ObserveHomePollution'
```

**출력 전문** (`exit=1`)

```
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.00s)
    todo_axisa_guard_test.go:184: precondition: "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestGuardBypassMutant_ObserveHomePollution2572639914/002" must classify as a temporary origin (reason ""); without it, an absence of pollution says nothing about the resolver-layer guard
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.867s
FAIL
```

이것은 결함이 아니라 **전제 가드가 제 일을 한 것**이다. 그 테스트의 판정은
「오염이 없다」는 부재 단언이고, 부재 단언은 base 가 실제로 임시 기원으로
분류될 때만 무언가를 말한다. 판별식을 죽이면 그 전제가 무너지므로, 테스트는
공허한 초록을 내는 대신 전제 위반을 보고하고 멈춘다.

**귀결이 중요하다: 뮤턴트 E 단독으로는 오염 재현을 실연할 수 없다.** 그래서
대조군은 §3 의 route (i) 여야 한다. 이 사실을 적지 않고 「뮤턴트를 걸었더니
잡혔다」로만 남기면, 다음 사람이 뮤턴트 E 로 오염 재현을 기대하다 전제 실패를
보고 「테스트가 틀렸다」로 오진한다.

### Route (i) — 배선 되돌림 (pre-guard 등가)

판별식은 **그대로 둔 채**, M2 가 넣은 배선 두 곳만 걷어냈다.
보존된 픽스처: `.moai/reports/t536/mutants/route-i-preguard.patch.txt`.

```
@@ func ResolveTodoQueueRoot(base string) string {
-	if root, _, refused := tempOriginSubstituteRoot(base); refused {
-		return root
-	}
@@ func ResolveTodoQueueRootAdopting(base string) string {
-	if root, _, refused := tempOriginSubstituteRoot(base); refused {
-		return root
-	}
```

`tempOriginSubstituteRoot` 자체는 남긴다 — `TempOriginRefusal` 이 계속 호출하므로
패키지는 컴파일되고 `internal/cli` 안내 경로도 그대로다. **해상도(resolution)
에서만** 가드가 빠지며, 그것이 시험 대상 축이다.

---

## 2. 어느 경계를 드러내는가

`TestGuardBypassMutant_ObserveHomePollution` 은 t422 가 세운 `runTodo` 의
`liveTodoQueueRootReason` 게이트를 **의도적으로 우회**해 `newTodoCmd()` 를 직접
실행한다. 역사적으로 그 우회는 홈 루트 오염을 만들었고, 그 오염 관측이
AC-SA-010 의 0 이 우연이 아니라는 증거였다.

지금 그 우회는 **오염을 만들지 못한다.** 두 가드가 서로 다른 층에 있기 때문이다.

| 층 | 소유 | 무엇을 막는가 | 우회 가능한가 |
|---|---|---|---|
| 테스트 층 | t422 — `runTodo` 의 `liveTodoQueueRootReason` 게이트 | 격리 픽스처 없이 todo 명령이 도는 것 | **그렇다** — `newTodoCmd()` 직접 실행이 그 우회다 |
| 리졸버 층 | t536 — `tempOriginSubstituteRoot` 의 임시-기원 거부 | 임시 기원 base 가 **애초에 홈 루트로 해석되는 것** | 이 우회로는 아니다 |

**아래 층이 위 층의 우회를 덮는다.** 위 게이트를 뚫어도 홈 루트에 닿지 못하는
이유는 아래 리졸버가 그 base 에 대해 홈 루트를 **반환하지 않기** 때문이다.
t422 게이트는 생산 코드가 아니라 테스트 축의 장치이고, t536 가드는 생산 축이다.

그 테스트는 부재만 단언하지 않는다 — **긍정 귀속**을 함께 건다: 반환 루트가
가드의 대체 루트와 같은지 확인하고, 그 자리의 큐를 읽어 카드가 실제로 착지했는지
센다. 아무것도 하지 않은 명령도 「오염 0」은 똑같이 만족시키므로, 이 긍정 단언이
없으면 0 은 아무것도 말하지 못한다.

복원된 트리에서 그 테스트는 PASS 하며 다음을 남긴다:

```
=== RUN   TestGuardBypassMutant_ObserveHomePollution
    todo_axisa_guard_test.go:223: REQ-SA-011 second branch: the bypass mutant produced NO pollution under /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestGuardBypassMutant_ObserveHomePollution629961695/001/.moai/todo. Boundary revealed: t422's runTodo gate (test layer) is bypassed, but SPEC-TODO-HOME-TEMP-GUARD-001's temporary-origin refusal (resolver layer) never resolves to a home root for this base, so the card landed in the project-local queue at /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestGuardBypassMutant_ObserveHomePollution629961695/002/.moai/state/todo/backlog.json instead. Full report with the pre-guard control run: .moai/reports/t536/guard-boundary.md
--- PASS: TestGuardBypassMutant_ObserveHomePollution (0.13s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.972s
```

---

## 3. 오염 부재가 왜 가드의 작동이지 뮤턴트의 실패가 아닌가 — 대조군

부재는 그 자체로 아무것도 세우지 못한다. 「가드가 막았다」와 「뮤턴트가 그냥 안
돌았다」는 **같은 출력**(오염 0)을 낸다. 둘을 가르는 것은 대조군뿐이고, 대조군의
요건은 **가드 하나만 다른 두 실행**이다.

**경로는 route (i) 다 — 이 트리 안에서 재측정했다.** `412c8cb14` 을 체크아웃하지
않았다(공유 트리의 브랜치 상태 변경은 금지이며, 이 트리는 앵커돼 있다). 대신 M2
배선 두 곳을 되돌려 pre-guard **등가** 상태를 이 트리에 만들고 같은 테스트를
돌렸다. 선행 측정을 인용한 것이 아니라 이번 실행의 측정이다.

**명령** (배선 되돌린 상태)

```
go test ./internal/cli/ -count=1 -vet=off -timeout 900s -run 'TestGuardBypassMutant_ObserveHomePollution'
```

**출력 전문** (`exit=1`)

```
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.09s)
    todo_axisa_guard_test.go:202: the mutant produced home pollution under /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestGuardBypassMutant_ObserveHomePollution690346063/001/.moai/todo (1 entr(ies)) — the resolver-layer temporary-origin guard regressed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.943s
FAIL
```

**오염이 재현됐다 — canary HOME 아래 1건.**

같은 배선 되돌림 상태에서 AC-THG-001 도 세 갈래 전부 RED 이며, 그 출력은
`.moai/reports/t536/m3/routei-kanban.txt` 에 있다. 갈래 (a) 의 실패 메시지가
M2 가 남긴 pre-guard RED 기록(`m2-red-ac-thg-001.txt`)과 같은 형태다 — 줄번호만
테스트 파일이 자란 만큼 이동했다(64/67/75 → 77/80/88).

### 두 실행의 차이는 가드 하나다

| 실행 | 판별식 | M2 배선 | 게이트 우회 뮤턴트의 결과 |
|---|---|---|---|
| 복원된 트리 (`db45d8209`) | 살아 있음 | 있음 | **오염 0** + 카드가 project-local 큐에 착지 (PASS) |
| route (i) pre-guard 등가 | 살아 있음 | **없음** | **오염 1건** under canary `$HOME/.moai/todo` (FAIL) |
| 뮤턴트 E | **죽음** | 있음 | 전제에서 정지 — 오염 단언에 미도달 |

첫 두 행이 대조군이다. 테스트도, 트리도, 기계도 같고 **M2 배선의 유무만**
다르다. 따라서 첫 행의 0 은 가드가 만든 것이지 뮤턴트가 돌지 않아서가 아니다 —
같은 뮤턴트가 둘째 행에서 실제로 오염을 만들었기 때문이다.

세 번째 행은 대조군이 아니라 §1 이 기록한 별개의 관측이다: 판별식과 배선은
**각각** 필수이며, 어느 하나만 죽여도 가드는 무너진다. 다만 판별식을 죽이면
게이트 우회 테스트는 전제에서 멈추므로 그 경로로는 오염을 볼 수 없다.

---

## 4. 복귀 확인 — 뮤턴트는 트리에 남지 않았다

주입은 작업트리의 임시 상태였고, 관측 직후 같은 턴에서 되돌렸다. 무출력을 증거로
쓰므로, **그 명령이 출력을 낼 수 있는 상태였음**을 대조로 함께 남긴다.

**주입 중** (route (i) 적용 상태) — 같은 명령이 실제로 diff 를 뱉었다:

```
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go | wc -c
     731
```

(그 731 바이트 전문은 `.moai/reports/t536/m3/restore-positive-control.diff`.)

**복원 후** — 같은 명령, 무출력:

```
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go
$ git diff -- internal/kanban/todo_root.go internal/cli/todo.go internal/kanban/temp_origin.go | wc -c
       0

$ git --no-optional-locks status --porcelain
?? .moai/reports/t536/m3/
```

`??` 한 줄은 M3 가 의도적으로 추가한 증거 디렉터리다. 추적 파일 수정은 0.
`git rev-parse --short HEAD` 는 주입 전후 모두 `db45d8209` — HEAD 는 움직이지
않았다.

**뮤테이션 창 위생**: 주입 직전 `git --no-optional-locks status --porcelain` 으로
배타성을 확인했고(추적 수정 0), 주입한 파일은 같은 턴 안에서 복원했다. 이 창
안에서 컴파일된 바이너리는 귀속 불가이므로 설치하거나 재사용하지 않았다 —
`make install` / `cp ~/go/bin/moai` 계열은 이 카드에서 한 번도 실행하지 않았다.

---

## 5. 미검증 (Gaps)

- **리눅스/윈도우 셀에서의 오염 재현은 관측하지 않았다.** 위 대조군은 darwin
  단일 셀이다. `/tmp` 이 실디렉터리이고 `/private/tmp` 이 없는 환경에서 같은
  대조가 어떻게 나오는지는 이 기계에서 잴 수 없다 — CI 판정에 맡긴다.
- **`GOOS=windows` 크로스빌드는 테스트 파일을 컴파일하지 않는다.** `go build` 는
  non-test 패키지만 본다. 따라서 windows 빌드 통과는 이 카드의 테스트가 windows
  에서 컴파일된다는 뜻이 **아니다**.
- 뮤턴트를 더 넓게 훑지 않았다. 판별식 내부의 개별 분기(정규화 규칙,
  `pathWithin` 의 구성요소 비교)에 대한 뮤턴트는 M1 이 A/C/D 로 이미 다뤘고
  (`m1-mutant-red.txt`), M3 은 REQ-THG-001 축의 뮤턴트만 다룬다.
- `SPEC-STATE-ANCHOR-001` M4 의 기록(route (ii))은 **인용하지 않았다.** route (i)
  가 성립했으므로 선행 측정을 대조군으로 재사용할 이유가 없다.

## 6. 잔여 위험

- route (i) 는 pre-guard **등가**이지 `412c8cb14` 그 자체가 아니다. 등가성의
  근거는 「M2 가 추가한 것이 정확히 이 두 블록」이라는 커밋 사실(`db45d8209` 의
  `todo_root.go` diff)이다. 그 사이에 제3의 경로로 가드가 들어왔다면 등가성이
  깨지지만, 같은 diff 가 그것을 배제한다.
- 이 보고의 대조군은 `TestGuardBypassMutant_ObserveHomePollution` 한 건에
  걸려 있다. 그 테스트가 나중에 약화되면(예: 긍정 귀속 단언이 빠지면) 이 문서가
  인용하는 0 의 의미도 함께 약해진다.
- 뮤턴트 E 가 게이트 우회 테스트의 전제에서 멈춘다는 사실은, 그 테스트가
  **판별식 회귀에 대해서는 오염 축으로 아무것도 말하지 못한다**는 뜻이기도 하다.
  판별식 축의 회귀 방어는 AC-THG-002/004/006 과 M1 뮤턴트가 담당한다.

---

## 7. 증거 반출 목록 + 트리 확인

판정 근거로 쓴 자료는 전부 추적 경로 `.moai/reports/t536/` 아래로 반출했고,
본문의 인용은 그 파일을 지목한다. 세션 스크롤백·셸 히스토리·`/tmp` 에만 있는
자료는 근거로 쓰지 않았다.

| 인용 위치 | 반출 파일 |
|---|---|
| §1 뮤턴트 E — 정확한 형태 | `.moai/reports/t536/mutants/mutantE.go.txt` |
| §1 뮤턴트 E — AC-THG-001 출력 | `.moai/reports/t536/m3/mutantE-kanban.txt` |
| §1 뮤턴트 E — 게이트 우회 전제 정지 | `.moai/reports/t536/m3/mutantE-cli-bypass.txt` |
| §1 route (i) — 정확한 형태 | `.moai/reports/t536/mutants/route-i-preguard.patch.txt` |
| §3 route (i) 대조군 — 오염 재현(1차) | `.moai/reports/t536/m3/routei-cli-bypass.txt` |
| §3 route (i) 대조군 — 오염 재현(재현행, 트리 스탬프 포함) | `.moai/reports/t536/m3/batchB-stamped-routei.txt` |
| §3 route (i) — AC-THG-001 RED | `.moai/reports/t536/m3/routei-kanban.txt` |
| §4 복귀 확인 — **주입 중** diff (731 bytes) | `.moai/reports/t536/m3/routei-injected.diff` |
| §4 복귀 확인 — **복원 후** diff (무출력) | `.moai/reports/t536/m3/routei-restored.diff` |
| §2 복원 트리의 PASS + 경계 로그 | `.moai/reports/t536/m3/batchC2-stamped.txt` |
| 배치 검증 — D8/D7/계수/vet/windows 빌드 | `.moai/reports/t536/m3/batchA-stamped.txt` |
| 배치 검증 — AC 전수 verbose + kanban/web 판정 | `.moai/reports/t536/m3/batchC1-stamped.txt` |
| 배치 검증 — cli 가드 테스트 + 패키지 판정 | `.moai/reports/t536/m3/batchC2-stamped.txt` |
| M1 뮤턴트 A/C/D (선행 마일스톤) | `.moai/reports/t536/m1-mutant-red.txt` |
| M2 pre-guard RED 기록 (선행 마일스톤) | `.moai/reports/t536/m2-red-ac-thg-001.txt` |

### 트리 확인

측정 배치마다 첫 명령으로 트리를 읽었고, 그 출력이 각 배치 파일의 머리에 있다.
뮤테이션 창은 주입 전·주입 후·복원 전·복원 후 **네 지점** 모두에서 재확인했다
(`batchB-stamped-routei.txt`) — 트리 판독은 HEAD 판독처럼 낡기 때문이다.

모든 지점에서 동일하게 관측된 값:

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536
$ git rev-parse --short HEAD
db45d8209
$ git branch --show-current
WT-home-fallback
```

이 확인이 왜 필요한가: 셸 CWD 가 조용히 이동하면 **엉뚱한 트리에서 우연히
성공하는 명령**이 그럴듯한 값을 낸다. 특히 소문자 `/Users/goos/moai/moai-adk-go`
는 자매 워크트리가 아니라 primary 체크아웃의 **두 번째 철자**이므로, 그리로 새면
워크트리 전용 파일이 없는데도 내용이 그럴듯한 출력이 나온다. `--show-toplevel`
이 대문자 `MoAI` 경로의 워크트리를 가리키는 것이 그 오염을 배제한다.

pre-guard 등가 재현(§3)은 트리 상태에 특히 민감하므로, 그 배치는 `git -C
<절대경로>` 와 절대경로 `go test` 대상만 사용했다 — 상대경로에 의존하지 않는다.
