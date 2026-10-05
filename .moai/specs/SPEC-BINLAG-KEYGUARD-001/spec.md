---
id: SPEC-BINLAG-KEYGUARD-001
title: 허용목록 키가 실재하는지 기계가 판정한다 — 따옴표 탈락으로 조용히 무효가 된 엔트리 적발
version: "0.5.0"
status: completed
created: 2026-09-04
updated: 2026-09-06
author: manager-spec
priority: Medium
phase: "v3.1.5 target"
module: internal/cli
lifecycle: spec-anchored
tags: "binary-lag, doctor-checks, guard-test, allowlist, meta-test, test-only"
tier: M
related_specs: [SPEC-BINARY-LAG-VISIBILITY-001, SPEC-BINLAG-INVOCATION-001]
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-09-04 | manager-spec | 최초 작성(카드 **t479**, plan-phase). 워크트리 `.claude/worktrees/t479`, base `25a3212a9`에서 실측한 값 위에 세움. 수리 방식은 **(a) 허용목록 키-모양 가드**로 좁혀졌고, 기각안 (b)와 그 기각 사유를 `plan.md` §B에 병기한다. 요구 7 / 수락 8 |
| 0.2.0 | 2026-09-04 | manager-spec | plan-audit FAIL 0.86의 blocking 3건(D1/D2/D3) + optional 4건(D4/D5/D6/D8) 반영. 트리가 **로컬 `develop @ a825183dd` 흡수 병합 커밋 `93fb36344`**로 옮겨졌고, t466의 `"Hook Delivery"` 리터럴 체크와 t477의 백틱 허용목록 엔트리가 **실재**한다 — §4를 「판정 불가」에서 「측정 가능」으로 재서술. **REQ-BLKG-003을 양방향으로 확장**(따옴표 탈락뿐 아니라 덧붙임도 지목). 좌표를 `93fb36344`에서 재측정해 갱신. 요구 7 / 수락 8 (수는 유지, 성격이 바뀐 것은 REQ-BLKG-003) |
| 0.3.0 | 2026-09-04 | manager-spec | 2차 감사 FAIL 0.87의 blocking 2건(R1/R2) + optional 2건(R4/R5) 반영. **0.2.0의 수정이 심은 결함을 철회한다** — 0.2.0에서 신설한 「`go test` 종료코드는 안전하다」는 예외 조항이 실측으로 **거짓**임이 드러났다(0매치 실행이 `ok` + exit 0 + `[no tests to run]`). 그 결과 AC-BLKG-001이 가드 부재 상태에서 이미 통과로 읽히고 있었다. **REQ-BLKG-005를 「셸 종료코드 일반 금지」에서 「검사 대상 0과 통과가 같은 신호로 합쳐지는 형태 금지」로 재서술**하고, AC-BLKG-001·005(2)에 비공허성 선단언을 붙였으며, AC-BLKG-008(1)을 열거 레시피로 바꿨다. 근거 규칙: `.claude/rules/moai/development/verification-completeness.md` §1.1 · §2.1 |
| 0.4.0 | 2026-09-06 | manager-spec | **최종 수정 1회**(운영자 결정: 이후 마지막 감사 → 판정과 무관하게 run 진입). R6 + 판정식 정정. ① RED-now를 표 칸에서 **증거 원장 §C**(EV-BLKG-001/002/003)로 옮겨 §2.1 네 요소를 verbatim으로 채웠다 — 표 칸은 줄바꿈을 뭉개 verbatim을 보존하지 못한다. ② **`$` 앵커가 상시 0**임이 실측으로 드러나 판정식을 **두 매치 수**(`[no tests to run]` = 0 **먼저**, 뒤 공백형 `--- PASS: <이름> ` = 1)로 바꿨다. ③ 1단만으로는 **SKIP을 못 막음**을 실측(EV-BLKG-003)해 2단을 필수화했다. ④ **AC-BLKG-009 신설** — 판정식 자신이 실패할 수 있음을 실행으로 보이는 항목. 수락 8 → **9**. 같은 함정이 세 번(D3 → R1 → `$` 앵커) 나왔고 셋 다 실행으로만 드러났다는 것이 ④의 근거다 |
| 0.5.0 | 2026-09-06 | manager-spec | 3차(최종) 감사 **PASS 0.90** — 부채를 안고 통과. 진입 조건 2건 + 재량 2건 반영. **T1**: `spec.md` §4가 `; echo "exit=$?"`의 `0`만 읽고 「기준선 GREEN」이라 판정하던 것이 **C 위반**이었고(이전 열거가 놓쳤다), 두 매치 수 판독으로 바꿔 없앴다. 열거 좌표는 **기대값에서 seed로 강등**(두 판 연속 부패한 관측이 근거)하고 기대값은 `C = 0` 하나만 남겼다. **T2**: 초록 읽기 형제 자리 4곳에 두 매치 수를 붙였다(감사 지적 2 + 스윕 중 발견 2). **T3**: 「≠ 0은 제약 밖」을 좁혀 실패의 정체(메시지/파일:줄)를 함께 요구하도록 했다 — `≠ 0`은 빌드·setup 실패도 포함한다. **R3**: 이 카드의 뮤턴트를 3건(m1/m2/m2′)으로 정정 |

---

## §1 문제 — 허용목록 엔트리는 틀린 모양으로 써도 아무도 알려주지 않는다

좌표는 모두 트리 **`93fb36344`** 기준이다.

`internal/cli/binary_lag_test.go`의 가드 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`
(`binary_lag_test.go:199`)는 `internal/cli/doctor.go`의 세 레지스트리
(`systemChecks` / `moaiChecks` / `workspaceChecks`)에 등록된 체크 **이름 표현식**을 AST로 훑어
집합으로 만든다(`checkNamesFromSource`, `binary_lag_test.go:118`). 집합의 원소를 만드는 것은
`exprSource`(`binary_lag_test.go:162`)이며, 이것은 이름 표현식의 **원본 소스 텍스트를 그대로**
돌려준다.

그래서 이 집합에는 **두 가지 모양**이 섞인다.

| 등록 방식 | 집합에 들어가는 원소 | 따옴표 |
|---|---|---|
| 상수 식별자 (`hookWiringCheckName`) | `hookWiringCheckName` | 없음 |
| 문자열 리터럴 (`"Hook Delivery"`) | `"Hook Delivery"` | **문자로 포함** |

허용목록 `namesAddedAfterBaseline`(`binary_lag_test.go:194`)은 `binary_lag_test.go:213`에서
이 집합의 원소와 **직접 대조**된다. 따라서 허용목록의 키도 같은 모양이어야 한다 — 리터럴로 등록된
체크를 허용하려면 키를 Go **backtick raw string**으로 써서 따옴표를 글자로 살려야 한다.

결함은 여기에 있다. **따옴표를 뺀 키는 매치가 0**이다. 엔트리를 적어 넣은 사람의 눈에는 허용목록에
이름이 들어가 있으므로 고쳐진 것처럼 보이지만, 가드는 여전히 RED이고, 실패 메시지는
「이 SPEC이 체크 이름을 추가했다」라고만 말한다. 즉 실패의 원인이
**「허용목록 엔트리가 등록되지 않았다」**인데 화면에는 **「이름이 진짜로 드리프트했다」**로 표시된다.
원인과 증상이 갈리지 않는다.

### §1.1 이것은 추정이 아니라 실측이다

카드 **t477**(브랜치 `WT-binarylag-allowlist`, 커밋 `b2f98f6aa`)이 뮤턴트 2개로 확인했다.

- 뮤턴트 1 — 허용목록 엔트리 제거 → `binary_lag_test.go:215` FAIL
- 뮤턴트 2 — 따옴표를 벗긴 형태로 등록 → `binary_lag_test.go:216` FAIL

(두 줄 번호는 t477의 트리 `b2f98f6aa` 좌표다. 이 카드의 트리 `93fb36344`에서는 같은 대조 지점이
`binary_lag_test.go:213`이며, 실패 줄은 재현 시 다시 잰다.)

뮤턴트 2가 backtick 형태가 **하중을 지고 있다**는 유일한 증거다. 뮤턴트 1만으로는
「엔트리가 있으면 통과」까지만 보이고, 「엔트리의 *모양*이 틀리면 없는 것과 같다」는 보이지 않는다.

### §1.2 지금 이 규율이 사는 곳은 코드 주석 한 줄뿐이다

`namesAddedAfterBaseline` 위의 주석이 의도를 적고 있으나, **기계적 가드는 0이다**.

트리 `93fb36344`의 허용목록은 이제 **두 모양을 나란히** 담고 있다:

    var namesAddedAfterBaseline = map[string]bool{
    	"hookWiringCheckName": true,
    	`"Hook Delivery"`:     true,
    }

즉 다음 사람은 서로 다른 두 모양을 눈앞에 두고 셋째 엔트리를 적게 된다. 어느 쪽을 베끼느냐로
결과가 갈리는데, 그 선택이 옳았는지 말해 주는 것은 아무것도 없다. 실수는 양방향이다 —
리터럴 등록 이름의 따옴표를 **떨어뜨리거나**, 상수 등록 이름에 따옴표를 **덧붙이거나**.
둘 다 매치 0이고, 둘 다 같은 모양으로 빨갛다.

---

## §2 요구 (GEARS)

**REQ-BLKG-001** (Ubiquitous)
`namesAddedAfterBaseline`의 모든 키는 현재 `internal/cli/doctor.go`에서
`checkNamesFromSource`가 추출하는 이름 집합의 원소여야 한다.

**REQ-BLKG-002** (Unwanted)
어떤 키도 추출 집합에 존재하지 않으면서 조용히 허용목록에 남아 있어서는 **안 된다**(shall not).
그런 키는 어떤 이름도 허용하지 못하므로 증명 가능하게 무력(inert)하며, 가드는 그것을 실패로 알려야 한다.

**REQ-BLKG-003** (event-driven, When)
어떤 키 `k`가 추출 집합에 없고, `k`의 **따옴표 변형** 중 하나가 그 집합에 **존재할 때**,
가드의 실패 메시지는 어느 방향의 실수인지를 지목하고 그 방향에 맞는 교정을 제시해야 한다.
두 방향 모두를 구속한다.

| 관측 | 실수 | 제시할 교정 |
|---|---|---|
| `k`는 bare identifier이고 `"` + k + `"`가 집합에 있다 | 따옴표 **탈락** | backtick raw string으로 다시 쓸 것 |
| `k`가 따옴표로 감싸여 있고 그 안쪽 알맹이가 집합에 있다 | 따옴표 **덧붙임** | 따옴표 없는 평범한 문자열 키로 다시 쓸 것 |

**REQ-BLKG-004** (state-driven, While)
허용목록에 실질 키가 하나도 없는 동안, 새 가드는 통과를 주장해서는 안 된다 —
검사 대상이 0인 상태의 초록은 성취가 아니라 미착수의 부산물이다.

**REQ-BLKG-005** (Ubiquitous)
어떤 판정식도 **「검사 대상이 0이었다」와 「검사가 통과했다」가 같은 신호로 합쳐지는 형태**여서는
안 된다. 금지 대상은 「셸 종료코드 일반」이 아니라 이 합쳐짐이다.

구체적으로: 종료코드 **0을 통과로 읽는** 판정 줄은 **비공허성 단언과 짝지어질 때에만** 허용된다 —
검사 대상이 실재했음을 먼저 보이고, 그 성립을 조건으로 종료코드를 읽는다. 짝이 없는 단독 사용은 금지다.
종료코드 **≠ 0을 실패로 읽는** 판정은 공허한 경우가 실패 쪽으로 떨어지므로 이 제약 밖이다 —
**단, 실패의 정체를 함께 요구할 때에만** 그렇다. `≠ 0`은 「단언이 깨졌다」뿐 아니라 빌드 실패,
setup 실패(`[setup failed]`), 패닉도 포함하므로, `≠ 0`만 읽으면 「가드가 잡았다」와 「애초에
빌드가 안 됐다」가 합쳐진다. 그래서 이 SPEC에서 `≠ 0`을 읽는 자리는 **실패 메시지 본문 또는
파일:줄을 함께 요구**한다.

이 조항이 이렇게 쓰인 근거는 실측이다 — 러너의 종료코드도 안전하지 않다.
트리 `93fb36344`에서 새 가드가 **없는데도**:

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.704s [no tests to run]
    (exit 0)

즉 「이름이 어긋나 하나도 안 돌았다」가 「전부 통과했다」와 같은 신호를 낸다.
근거 규칙: `.claude/rules/moai/development/verification-completeness.md` §1.1
— *"A pass whose swept set is empty asserts nothing"*, 그리고 go의 `[no tests to run]` 토큰을
가장 값싼 증거로 명시한다.

**REQ-BLKG-006** (Ubiquitous)
이 SPEC의 변경 범위는 `internal/cli/binary_lag_test.go` 한 파일이다. 프로덕션 코드는 변경하지 않는다.

**REQ-BLKG-007** (Where, capability gate)
과거 blob(`lagBaselineSHA`)을 `git show`로 읽을 수 없어 기존 가드가 `Skip`하는 환경에서도,
새 가드는 현재 `doctor.go`만 읽으므로 독립적으로 판정되어야 한다.

---

## §3 수락 기준

수락 기준은 `acceptance.md`에 **AC-BLKG-001 … AC-BLKG-009**로 열거한다.
각 항목은 명령 하나로 독립 검증되며, 뮤턴트 **3건**(m1 / m2 / m2′)이 비공허성을 증명한다.
(`spec.md:59`와 `plan.md:177`의 「2건」은 카드 t477이 자기 창에서 돌린 것이며, 이 카드의 3건과는
다른 집합이다 — 고치지 않는다.)

---

## §4 기준선은 초록이지만, 이제 m2를 실제로 잴 수 있다

**전제는 착지했다.** 로컬 `develop @ a825183dd`를 흡수한 병합 커밋 **`93fb36344`**가 이 트리의 HEAD이며,
t466의 리터럴 체크와 t477의 백틱 허용목록 엔트리가 실재한다. 그 트리에서 실측:

    $ git rev-parse --short HEAD
    93fb36344

    $ grep -c 'Hook Delivery' internal/cli/doctor.go
    1

    $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.06s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.990s
    (exit 0)

**이 초록은 두 매치 수로 읽는다** — 종료코드 `0`만으로는 「기준선 GREEN」을 주장할 수 없다:

    [no tests to run] 계수 → 0      ← 빈 sweep 아님 (1단)
    뒤 공백형 PASS 계수  → 1      ← 선택·실행·통과 (2단, SKIP도 차단)

**이전 판은 여기서 `; echo "exit=$?"`의 `0`만 읽고 「기준선 GREEN」이라 판정했다.** 그러면 테스트
이름을 한 글자만 틀려도 아무것도 안 돈 채 「GREEN」이 된다 — 이 카드가 막으려는 바로 그 형태다.
AC-BLKG-008(1)의 열거가 이 줄을 놓쳤던 것이 최종 감사 T1이며, 그 누락을 여기서 닫는다.

즉 기준선은 여전히 **GREEN**이다. 달라진 것은 초록의 성격이다.

| | plan-phase 최초(`25a3212a9`) | 지금(`93fb36344`) |
|---|---|---|
| 리터럴 등록 체크 | 없음 (매치 0) | **있음** (매치 1) |
| 허용목록 키 모양 | 상수 등록 1개뿐 | **두 모양 공존** |
| 뮤턴트 m2 판정 | **불가** — 표본 부재 | **가능** |

착지 전 기준선의 GREEN은 **표본이 없어서** 초록이었다 — 따옴표가 하중을 지는 사례를 한 번도
보지 못한 채 얻는 초록이며, 그 위에서 설계한 가드는 막으려던 것을 못 막는다.
지금의 GREEN은 표본이 **있는데도** 통과한 초록이므로, 새 가드가 더해질 자리로 유효하다.

**착지 판정 ref 주의**: `origin/develop`은 `25a3212a9`에 멈춰 있고 로컬 `develop`보다 130 커밋 뒤다
(레인이 develop을 push하지 않는 구조상 정상이며 앞으로도 그렇다). 착지 판정과 흡수는
**로컬 `develop`**을 기준으로 한다 — 근거와 실측은 `plan.md` §C.

---

## §5 제약

- 테스트 파일 단일 편집. 프로덕션 동작·출력·바이너리에 영향 0.
- `lagBaselineSHA`를 범프하지 않는다 — 범프는 기준선 이후 누적된 **모든** 드리프트를 함께 침묵시킨다.
- 기존 가드 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`의 판정 의미를 바꾸지 않는다. 새 가드는 **더해지는** 것이지 대체가 아니다.
- 검증 부하는 카드 범위로 한정한다: `go test ./internal/cli/ -run <해당 테스트>`. 전체 스위트는 CI 몫이다.

---

## §6 범위 밖

이 카드가 만들지 **않는** 것들.

### Out of Scope — `exprSource` 정규화(기각안 b)

- 두 등록 방식이 같은 모양을 내도록 `exprSource`를 정규화하지 않는다.
- 기각 사유는 `plan.md` §B에 있다: `beforeNames`는 과거 blob에서 파싱되며, 오늘 존재하는 상수가 그 트리에는 아예 없을 수 있어 before/after 집합의 의미가 갈린다.

### Out of Scope — `lagBaselineSHA` 범프

- 기준선 SHA는 손대지 않는다. 범프는 이 결함을 고치지 않고 다른 드리프트를 함께 감춘다.

### Out of Scope — `doctor.go`의 체크 등록 방식 통일

- 상수 등록과 리터럴 등록 중 하나로 통일하는 작업은 프로덕션 코드 변경이며 이 카드 밖이다.
- 통일이 이뤄지더라도 이 가드는 무해하게 계속 참이 된다(무력화되지 않는다).

### Out of Scope — t466 / t477 자체의 재검증

- 두 카드의 판정은 각자의 창에서 이미 내려졌다. 이 카드는 그 결과를 **전제**로 삼을 뿐 다시 재지 않는다.

### Out of Scope — 다른 가드 파일 / CI 축

- `binary_lag_test.go` 외의 가드 테스트, GitHub Actions 워크플로, pre-commit 게이트는 건드리지 않는다.

### Out of Scope — 허용목록 항목의 정당성 심사

- 어떤 이름이 허용목록에 **들어가도 되는가**는 이 가드의 질문이 아니다. 이 가드는 오직 적힌 키가 **무엇이라도 매치하는가**만 판정한다.
