# SPEC-APPJS-FIRE-GUARD-001 — 인수조건

모든 AC 는 기계적으로 확인 가능하며 판정 명령을 함께 적는다. two-cell 규율(verification-completeness §2)에 따라 각 AC 는 **RED-now 셀**(§B2 증거 장부의 4요소 항목 — 명령·원문 출력·exit 코드·트리 SHA)과 **green path 셀**(어느 마일스톤이 무엇으로 뒤집는가)을 함께 가진다. RED-now 셀을 구성할 수 없는 계열(회귀·범위·문서)은 §D 에 regression-class 로 명시 선언되며, DoD 는 이 분류를 따른다(§C).

---

## §A AC 매트릭스

| AC | 방향 | 분류 | 대응 REQ |
|---|---|---|---|
| AC-AFG-001 | 정방향 | blocking | REQ-AFG-001, REQ-AFG-003, REQ-AFG-005, REQ-AFG-006, REQ-AFG-007 |
| AC-AFG-002 | 역방향 | blocking | REQ-AFG-008 |
| AC-AFG-003 | 경계 | blocking | REQ-AFG-001 |
| AC-AFG-004 | 생존 | blocking | REQ-AFG-004 |
| AC-AFG-005 | 회귀·범위 | regression-class (§D) | REQ-AFG-002, REQ-AFG-011, REQ-AFG-013 |
| AC-AFG-006 | CI 고립 | regression-class (§D) | REQ-AFG-010 |
| AC-AFG-007 | 복원 | blocking | REQ-AFG-009 |
| AC-AFG-008 | 문서 | regression-class (§D) | REQ-AFG-013 |
| AC-AFG-009 | 영속화 안전 (효과 종별 계열) | blocking | REQ-AFG-012, REQ-AFG-014 |
| AC-AFG-010 | 도달(paint) | blocking | REQ-AFG-015 |
| AC-AFG-011 | 무쓰기 + 사본 서빙 | blocking | REQ-AFG-014 |
| AC-AFG-012 | 조건 불가분성 | blocking | REQ-AFG-014 |
| AC-AFG-013 | 폭발 반경 0 (라우팅) | blocking | REQ-AFG-014 |

AC-010~013 은 card t1106 개정분이다(REQ-AFG-014·015). AC-001~008 의 **번호·분류·대응 REQ 는 불변**이나, **문언은 세 곳이 개정됐다** — **AC-AFG-001** 의 Then (a) 는 이 사이클의 판정 집합을 「사본 서빙 표식이 없는 항목 전부」라는 **술어**로 명시하고 보고서에 운전 항목 수·선언 제외 항목 이름을 함께 요구하도록 바뀌었고(card t1106 iter-2·iter-3), **AC-AFG-008** 의 Then 은 한계 축이 넷에서 일곱으로 늘며 앵커 토큰 `paint`·`sandbox-root`·`two-surfaces` 셋이 보태졌으며(iter-2, D3 수리), 셋째가 아래 문단의 AC-AFG-009 다. **AC-AFG-009 는 문언만 개정됐다** — 리터럴 다섯 열거가 `validation-reject` 편입 뒤 AC-AFG-012 와 정면 모순되므로(둘 다 blocking, 같은 매니페스트에 exit 1 과 exit 0 을 동시 요구), 그 목록이 인코딩하던 규칙으로 바꿨다. 번호·분류·대응 REQ 의 REQ-AFG-012 는 그대로이고, 요구사항 문언은 한 글자도 바뀌지 않았다.

iter-2 수리로 매핑이 갈린 곳: REQ-AFG-007 → AC-001(c) 가 실질 검증(스왑 뒤 지표 행사), REQ-AFG-002 → AC-005(커밋된 탐침의 존재·파생 출처·go.mod 무변경), REQ-AFG-012 → 신설 AC-009.

---

## §B Given-When-Then

### AC-AFG-001 — 정방향: 통과하되 공허하지 않게

**Given** `MOAI_BROWSER_GUARD=1` 과 Chrome·python3·websockets 가 모두 있는 환경, 그리고 이 트리의 `internal/web` 패키지가 있을 때,
**When** 드라이버 테스트를 실행하면,
**Then** exit 0 이고 **그리고 동시에** (a) 이 사이클의 운전 집합이 **매니페스트에서 사본 서빙 표식이 없는 항목 전부**와 **정확히 일치**하고(술어로 판정한다 — 개정 시점의 「8」이라는 수가 아니라 「표식 없음」이라는 성질이 기준이며, 표식 없는 항목이 하나라도 선언으로 제외되면 실패다), 그 수가 0보다 크며 **그 전부가 발화**했고, 제외 집합이 **정확히 표식 계열**임을 확인할 수 있도록 보고서에 **운전한 항목 수**와 **선언으로 제외된 항목의 이름**이 함께 적혀 있어야 하며, (b) load·swap ReferenceError 가 0건이며, (c) hx-boost 스왑 **뒤에** 최소 1지표가 발화했음을 보고서가 담고 있어야 한다.

세 단언이 **함께** 걸려야 한다. 지표 단언만 있으면 빈 매니페스트가 통과하고, ReferenceError 단언만 있으면 버튼이 안 눌려도 통과한다.

> **이 사이클의 판정 집합은 실루트 계열 8항목이고, 제출은 여기 들어오지 않는다 — 결정이지 숫자 정정이 아니다(card t1106 iter-2, D1 수리).** 개정으로 매니페스트에 아홉째 항목(`validation-reject`)이 들어오지만, 이 AC 의 「전 항목」은 **그것을 포함하도록 넓어지지 않는다**. AC-AFG-001 은 **상속된 blocking 기준**이고 그 논지는 「이 매니페스트가 전부 발화한다」이다. 아홉째를 조용히 그 안에 들여보내면, 누구도 결정하지 않은 채 상속 기준의 **의미가 바뀐다** — 제출을 행사하는 사이클과 행사하지 않는 사이클이 한 이름 아래 섞이고, 나중 독자는 초록을 보고 어느 쪽이 돌았는지 알 수 없다. 그래서 두 계열을 **AC 층에서 가른다**.
>
> **가름의 형태.** 이 AC 의 실행은 REQ-AFG-014 (1) 의 **적극적 축소 선언**(실루트 계열만 운전)을 달고 돌며, 따라서 사본 base 없이도 exit 2 에 걸리지 않는다. 그 선언의 비공허성은 같은 조항이 진다 — 보고서가 운전 항목 수와 **선언으로 제외된 항목의 이름**을 함께 담고, 제외 집합이 정확히 표식 계열이며, 운전 집합이 비면 exit 1 이다. 즉 「아무것도 운전하지 않는 선언」으로 이 AC 를 통과할 수 없고, **부재가 축소를 함의하지도 않는다**(선언 없이 사본 base 가 없으면 exit 2 — 침묵 스킵은 여전히 금지다).
>
> **제출 사이클은 AC-AFG-010 이 진다.** 그 AC 의 Given 이 이미 사본 서버를 전제로 세워 두었다 — 제출 계열의 **구별되는 전제**(사본 base 가 존재할 것, 두 번째 서버가 떠 있을 것)가 거기서는 전제로 서술되고, 여기 일반 기준 안으로 밀려 들어오지 않는다. 표식 항목의 무쓰기는 AC-AFG-011, 라우팅 양방향은 AC-AFG-013 이 진다. 따라서 매니페스트의 아홉 항목 중 **어느 것도 판정 없이 남지 않는다**: 8항목은 이 AC, 아홉째는 AC-010·011·013.
>
> 폭발 반경과 상속 표면은 그대로다 — 기존 8항목의 서빙 루트는 `findRepoRoot` 로 **불변**이고(AC-AFG-013 (a)(c)), 이 AC 가 재는 표면은 개정 전후로 한 바이트도 달라지지 않는다. `plan.md` §E M7.8 이 같은 집합(실루트 계열 8항목 + 적극적 축소 선언)을 운반한다.

- **RED-now**: 증거 장부 **E1** — 가드 테스트가 존재하지 않아 기준을 통과할 주체 자체가 없다(빈 스윕, exit 0). 스위프트 토큰 `[no tests to run]` 이 그 적색의 관측 형태다(§1.1).
- 보조 baseline 관측(적색 아님 — 표면의 현재 상태): 예시 탐침 5지표 전부 발화 + ReferenceError 0건(spec.md §B.1, d726ac709 재측정).
- **green path**: M2 가 (a)(b)(c) 를 Go 단언으로, M1 이 exit 계약을 탐침에 실어 뒤집는다. 통과 형태: `--- PASS: TestAppJsHandlersFireRuntime` + `ok` 요약행.

```bash
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -v -count=1
```

### AC-AFG-002 — 역방향: 돌연변이를 거부하고 지목한다

**Given** plan.md §B 의 돌연변이 사이클(아래 판정 명령 전문)이 실행된 트리가 있을 때,
**When** 변이·재빌드 뒤 탐침을 실행하면,
**Then** 탐침이 exit **1** 이어야 하고, 보고가 최소 1개 지표의 미발화 또는 `stampRefreshed` ReferenceError 를 **이름으로** 담아야 하며, 돌연변이기 자신은 대상 줄을 못 찾으면 exit 비0 으로 실패해야 한다. 복원 뒤 같은 시퀀스의 마지막 탐침은 exit **0** 이어야 한다(0→1→0).

판정 명령(구체 시퀀스 — plan §B 의 그것이다, 자리표시자 아님):

```bash
go build -o /tmp/moai-fire-guard ./cmd/moai
python3 /tmp/t1060-probe/probe.py <port> pre          # exit 0 기대
python3 internal/web/testdata/appjs_fire_mutation.py  # exit 0 기대(변이 합성 성공)
go build -o /tmp/moai-fire-guard ./cmd/moai
python3 /tmp/t1060-probe/probe.py <port> post-mutate  # exit 1 기대 — 이 AC의 판정점
cp /tmp/t1060-probe/app.js.pristine internal/web/assets/app.js
cmp internal/web/assets/app.js /tmp/t1060-probe/app.js.pristine
```

(서버·Chrome 기동·정리 래퍼는 plan §B 스텝 시퀀스와 동일하다 — 각 행이 단일 호출이며, 판정점은 `post-mutate` 탐침의 exit 코드다.)

- **RED-now**: 증거 장부 **E5** — 같은 돌연변이 사이클을 현재 트리(d726ac709)에서 재실행하면 지표는 전부 붕괴하지만 탐침이 **exit 0** 을 낸다. 요구(1)와 관측(0)이 어긋나는, 올바른 이유의 적색이다 — exit 계약이 아직 구현되지 않았기 때문이다.
- **green path**: M1 이 3값 exit 계약을, M3 이 돌연변이기와 시퀀스를 구현해 뒤집는다. 통과 형태: `post-mutate` 탐침 exit 1 + 지명된 붕괴 지표.

### AC-AFG-003 — 전제 부재의 skip 은 사유를 이름으로 댄다

**Given** `MOAI_BROWSER_GUARD` 가 설정되지 않은(또는 Chrome 이 없는) 평상 환경일 때,
**When** 패키지 테스트를 실행하면,
**Then** 드라이버 테스트는 SKIP 이고 그 사유에 **빠진 전제의 이름**(게이트 변수 `MOAI_BROWSER_GUARD` 또는 `chrome` / `python3` / `websockets`)이 **영어로**(error_messages: en 정책) 적혀 있어야 한다. skip 은 「이 환경에서 재지 않았다」의 기록이지 통과가 아니며, CI 판정면은 전용 job 이 유일하게 운반한다.

```bash
go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -v | grep -E 'SKIP'
# 사유 행에 게이트 변수 또는 부재 전제 이름(영어)이 보여야 한다
```

- **RED-now**: 증거 장부 **E1** (AC-001 과 동일 관측 — 사유 행이 존재할 주체가 없다).
- **green path**: M2 — 게이트 분기가 이름 붙은 영어 skip 사유를 출력한다.

### AC-AFG-004 — 셀렉터 스테일은 red 다

**Given** 매니페스트의 셀렉터 하나를 고의로 실재하지 않는 값으로 바꾼 자기검증(드라이버 서브테스트가 일회용 사본에서 수행)이 있을 때,
**When** 그 서브테스트를 실행하면,
**Then** 탐침이 exit 1 이고 보고가 **그 셀렉터 이름**을 미달 항목으로 지목했음을 서브테스트가 단언해야 한다.

```bash
go test ./internal/web/ -run 'AppJsHandlersFireSelectorMiss' -v -count=1
```

- **RED-now**: 증거 장부 **E2** — 서브테스트가 존재하지 않는다(빈 스윕, exit 0).
- **green path**: M1 이 탐침의 셀렉터 생존 검사를, M2 가 서브테스트를 만들어 뒤집는다.

### AC-AFG-005 — 기존 계약 회귀 없음 + 산출물 귀속

**Given** `internal/web` 패키지의 기존 테스트 전부와 이 SPEC 의 신규 파일들이 있을 때,
**When** 패키지 테스트를 돌리고 변경 면을 재면,
**Then** (a) 기존 테스트가 전부 통과하고 `appjs_iife_scope_test.go`·`appjs_reinit_test.go` 는 수정되지 않았으며, (b) 탐침이 **커밋된** `internal/web/testdata/appjs_fire_probe.py` 로 존재하고 상단에 t1041 예시 탐침 파생 출처가 기록돼 있으며, (c) `go.mod`/`go.sum` 이 변경되지 않았어야 한다.

```bash
go test ./internal/web/ -count=1
git diff --stat -- internal/web/appjs_iife_scope_test.go internal/web/appjs_reinit_test.go   # 기대: 출력 없음
ls internal/web/testdata/appjs_fire_probe.py                                                 # 기대: 파일 존재, exit 0
grep -c 't1041' internal/web/testdata/appjs_fire_probe.py                                    # 기대: 1 이상(파생 출처)
git diff --stat <base>..HEAD -- go.mod go.sum                                                # 기대: 출력 없음
```

> `internal/web` 는 크다. FAIL 요약이 실패 0건과 함께 나오면 타임아웃을 의심하고 `-timeout 30m` 으로 재측정한다(과거 실측 기록 있음).

- 분류: **regression-class** — RED-now 셀 없음(§D 선언). green path: M5 전체 재측정.

### AC-AFG-006 — 기존 CI 전제 무변경

**Given** 이 SPEC 의 구현 diff 가 있을 때,
**When** `ci.yml` 의 diff 를 기계로 재면,
**Then** (a) 삭제·수정 행이 **0개**이고(추가만 존재), (b) 기존 job 키 목록(detect/test/test-race/test-skip-marker/test-integration/lint/build/constitution-check)과 그 순서가 base 와 동일해야 한다.

```bash
git diff --numstat <base>..HEAD -- .github/workflows/ci.yml        # 기대: "0 <N>" — 삭제 0
git diff <base>..HEAD -- .github/workflows/ci.yml | grep -c '^-[^-]'   # 기대: 0
grep -n -E '^  [a-z0-9-]+:' .github/workflows/ci.yml              # base 측정치와 목록·순서 비교
```

범위 판정식은 흡수 ref 기준 `merge-base`으로 잡고, 귀속은 `git log <base>..HEAD -- <경로>` 커밋 열거로 답는다 — 두 끝점 diff 로 단정하지 않는다. 세 명령 모두 기계 판정이며 육안 단계는 없다(iter-2 D8 수리).

- 분류: **regression-class** — RED-now 셀 없음(§D 선언). green path: M4.

### AC-AFG-007 — 돌연변이는 트리에 흔적을 남기지 않는다

**Given** 레드 단계(또는 로컬 재현)의 돌연변이 사이클이 끝난 트리가 있을 때,
**When** 복원·검증 절차의 결과를 재면,
**Then** `internal/web/assets/app.js` 는 원본과 byte 동일(`cmp` exit 0)이고 `git status --porcelain` 에 assets 변경이 없어야 한다.

```bash
cmp internal/web/assets/app.js <원본 사본>
git status --porcelain -- internal/web/assets/
```

- **RED-now**: 증거 장부 **E3** — 복원 절차의 운반체(커밋된 돌연변이기·스크립트)가 존재하지 않는다(파일 부재, exit 1). 기준을 집행할 절차가 아직 트리에 없다는 적색이다.
- 보조 baseline 관측(적색 아님): d726ac709 재실행 사이클이 `RESTORED_BYTE_IDENTICAL` + assets porcelain 빈 출력을 냈다(E5 복원 구간).
- **green path**: M3 — 돌연변이기+복원 검증이 커밋되고 AC-002 시퀀스 안에서 매번 실행된다.

### AC-AFG-008 — 한계와 직교성이 소스에 적혀 있다

**Given** 완성된 드라이버·탐침 소스가 있을 때,
**When** 상단 영어 주석(code_comments: en)에서 고정 앵커 토큰을 셈하면,
**Then** 직교성 문장과 **일곱 한계 축**(spec.md §F — 1~4 는 원판, 5~7 은 card t1106 개정분)을 가리키는 앵커 토큰 — `orthogonal`(정적 형제와의 직교성, `SPEC-APPJS-IIFE-GUARD-001` 명시와 함께), `manifest`(§F 1 — 매니페스트 밖 미측정), `scenario`(§F 2 — 단일 시나리오), `Chrome`(§F 3 — 단일 브라우저), `static-scope`(§F 4 — 원인 판정의 정적 가드 소관), `paint`(§F 5 — 칠해짐 판별식의 범위: 문구·대비·포커스는 밖), `sandbox-root`(§F 6 — 무쓰기 단언의 범위는 사본 루트뿐), `two-surfaces`(§F 7 — 두 서버 서면의 등가를 주장하지 않음) — 이 **모두** 소스에 등장해야 한다.

```bash
grep -c 'orthogonal' internal/web/appjs_fire_guard_test.go internal/web/testdata/appjs_fire_probe.py
grep -c 'SPEC-APPJS-IIFE-GUARD-001' internal/web/appjs_fire_guard_test.go
grep -c 'manifest' internal/web/testdata/appjs_fire_probe.py
grep -c 'scenario' internal/web/testdata/appjs_fire_probe.py
grep -c 'Chrome' internal/web/testdata/appjs_fire_probe.py
grep -c 'static-scope' internal/web/appjs_fire_guard_test.go
grep -c 'paint' internal/web/testdata/appjs_fire_probe.py
grep -c 'sandbox-root' internal/web/testdata/appjs_fire_probe.py
grep -c 'two-surfaces' internal/web/appjs_fire_guard_test.go
```

각 기대치는 1 이상(주석 안의 등장이면 충분하다). 토큰 집합의 소유: 원판 5토큰은 plan.md §E M1.6, 개정분 `paint`·`sandbox-root` 는 M6.6(탐침 주석), `two-surfaces` 는 M7.7(드라이버 주석)이다 — §F 의 축 수와 앵커 토큰 수는 함께 움직인다(card t1106 iter-2, D3 수리: 종전 문언은 §F 가 7축으로 늘어난 뒤에도 「네 한계 축」에 머물러 있었다).

- 분류: **regression-class** — RED-now 셀 없음(§D 선언). green path: M5.

### AC-AFG-009 — 매니페스트의 효과 종별은 두 닫힌 계열 중 하나여야 한다

> **card t1106 개정 — 왜 문언이 바뀌었나.** 종전 문언은 비영속 허용목록 다섯을 **리터럴로 열거**하고 그 밖의 항목이 하나라도 있으면 exit 1 을 요구했다. run-phase 가 `validation-reject` 항목을 커밋하는 순간 그 다섯 밖의 종별이 커밋된 매니페스트에 들어오므로, 종전 AC-AFG-009 는 **AC-AFG-012 가 exit 0 을 요구하는 바로 그 매니페스트에 exit 1 을 요구하게 된다** — 둘 다 blocking 이므로 어느 구현도 둘을 함께 초록으로 만들 수 없다. 리터럴 목록을 그 목록이 **인코딩하던 규칙**으로 바꿔 모순을 없앤다. 넓히지 않는다: 아래 두 계열은 **둘 다 닫힌 열거**이고, 조건부 계열의 오늘 원소는 정확히 하나(`validation-reject`)다.

**Given** 커밋된 탐침과 그 (페이지, 셀렉터, 효과) 매니페스트가 있을 때,
**When** 탐침의 매니페스트 자기검증 모드를 실행하면,
**Then** 모든 항목이 다음 **두 닫힌 계열 중 하나**에 속해야 exit 0 이다:

- **비영속 계열(무조건 통과)** — 되돌아가는 효과만 행사하는 종별의 닫힌 열거 {`visibility`, `label`, `clipboard`, `tab`, `swap`}. 조건 없이 통과한다.
- **조건부 영속 계열(조건 충족 시에만 통과)** — 오늘 원소는 정확히 하나, `validation-reject`. REQ-AFG-014 의 **(1)·(2) 두 조건 표식**(사본 서빙 요구 · 무쓰기 단언 요구)을 **모두** 갖췄을 때만 통과하고, 하나라도 빠지면 exit 1 이다. **조건 (3)(수명이 `t.TempDir()` 파생)은 매니페스트 표식이 아니다**(card t1106 iter-2, D4 수리) — 그것은 Go 쪽 수명 속성이라 커밋된 매니페스트가 운반할 수 없고, 그 판정은 **AC-AFG-011 (d)** 가 드라이버 측에서 진다. 여기서 셋째 표식을 요구하면 run-phase 가 의미 없는 세 번째 매니페스트 필드를 만들어야 하고, 같은 매니페스트에 exit 1 을 요구하는 AC-AFG-012 (b)(표식 **둘**)와 정면으로 어긋난다. REQ-AFG-014 의 닫는 문단이 같은 분담을 명시한다 — 매니페스트 자기검증이 1·2 의 표식을, 드라이버 측 기계 판정이 1 의 라우팅과 3 의 수명을 진다.

그리고 두 계열 **어느 쪽에도 속하지 않는** 종별(저장·제출·설정 쓰기 계열 일반)이 하나라도 있으면 exit 1 이어야 한다. 조건부 계열은 **열거이지 규칙이 아니다** — 「표식만 갖추면 어떤 영속 종별이든 통과」가 아니며, 새 원소를 더하는 것은 SPEC 개정 사항이다. 이것이 REQ-AFG-012 의 「기본값은 제외」를 문언에 유지하는 방식이다.

```bash
python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest
```

**AC-AFG-012 와의 경계(중복 아님):** 이 AC 는 **커밋된 매니페스트가 규칙을 만족하는가**(정방향, exit 0)를 건다. AC-AFG-012 는 **규칙이 위반을 실제로 거부하는가**(역방향 — 조건이 빠진 합성 항목에서 exit 1 + 어느 조건이 빠졌는지 이름)를 건다. 정방향만 걸면 아무것도 거부하지 않는 판정식이 통과한다.

- **RED-now**: 증거 장부 **E4** — 탐침 파일 자체가 존재하지 않는다(파일 부재, exit 1). 데이터 손실 경로를 막는 이 기준이 현재 아무것도 검증하지 않는 상태가 관측이다(iter-1 감사가 지적한 REQ-AFG-012 무검증의 해소 지점).
- 보조 관측(적색 아님): 오늘의 `--lint-manifest` 는 exit 0 이지만(E9) 그 초록은 다섯 종별만 있는 상태의 것이다 — 조건부 계열에 대해서는 아무것도 주장하지 않는다.
- **green path**: M1 이 비영속 계열과 `--lint-manifest` 를 구현했고, **M6 이 조건부 계열과 그 조건 판정을 더해** 이 문언을 충족시킨다.

### AC-AFG-010 — 거부 배너는 화면에 도달한다 (card t1106)

**Given** `MOAI_BROWSER_GUARD=1` 과 전제(Chrome·python3·websockets)가 갖춰진 환경, 그리고 드라이버가 **일회용 프로젝트 사본** 위에서 서빙하는 서버가 있을 때,
**When** 탐침이 `validation-reject` 항목의 셀렉터로 검증에 실패하는 값을 담아 폼을 제출하면,
**Then** 거부 응답 이후의 문서에서 배너가 **칠해졌음**을 탐침이 보고해야 한다 — (a) 배너 노드가 레이아웃 박스를 가지고 조상 사슬 어디서도 은닉되지 않았고, (b) 텍스트가 비어 있지 않으며, (c) 같은 창에서 ReferenceError 가 0건. 노드의 **존재만** 확인하고 통과하면 이 AC 는 충족되지 않는다.

```bash
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsFireValidationRejectPaints' -v -count=1
```

> **이 AC 의 실행은 AC-AFG-001 의 실행과 별개다(card t1106 iter-2, D1 수리).** 제출 계열의 전제는 여기서 **전제로 서술된다** — 사본 base 가 공급되고(따라서 REQ-AFG-014 (1) 의 exit 2 경로에 걸리지 않는다) 두 번째 서버가 떠 있어야 한다. 판정점은 표식 항목의 paint 이며, 표식 없는 항목이 같은 실행에 동승하는지는 이 AC 의 판정에 영향이 없다 — 그 항목들의 판정은 AC-AFG-001 이 따로 진다.

- **RED-now**: 증거 장부 **E6** — 판정 주체(테스트)가 존재하지 않는다(빈 스윕, exit 0, `[no tests to run]`).
- 보조 관측(적색 아님 — 표면의 현재 상태): 이 트리에 응답 **본문** 층의 검증면이 이미 있다(`internal/web/transport400_swap_contract_test.go` 외 2건). 그 층은 도달을 주장하지 않는다(spec.md §B.6.1).
- **green path**: M6 이 탐침에 `validation-reject` 항목과 paint 술식을, M7 이 게이트된 드라이버 테스트를 실어 뒤집는다. 통과 형태: `--- PASS: TestAppJsFireValidationRejectPaints` + `ok` 요약행.
- **돌연변이 프로브**: paint 술식을 「노드 존재」로 약화하면, 배너를 `hidden` 으로 렌더하는 합성 변이에서도 통과해야 한다 — 그런 변이가 통과하면 판별식이 너무 얕아 채택 불가다(verification-completeness §2 mutant probe).

### AC-AFG-011 — 제출은 일회용 사본 위에서만, 그리고 사본은 바이트 불변이다 (card t1106)

**Given** `validation-reject` 항목을 행사하는 드라이버 테스트가 있을 때,
**When** 그 테스트를 실행하면,
**Then** (a) 그 항목을 서빙하는 서버의 `ProjectRoot` 가 테스트 수명과 함께 폐기되는 **일회용 사본**이어야 하고 `findRepoRoot` 결과가 아니어야 하며, (b) 탐침이 제출 **직전·직후** 사본 루트의 바이트 상태를 비교해 **불변**임을 단언해야 하고, (c) 사본에 한 바이트라도 쓰기가 일어나면 탐침이 **exit 1** 로 실패하며 무엇이 달라졌는지 경로를 이름으로 보고해야 한다. 비교에서 제외되는 경로가 있으면 각각 사유가 소스에 적혀 있어야 한다 그리고 (d) 사본 루트가 **`t.TempDir()` 에서 파생된 경로**임을 테스트가 단언해야 하고, 사본의 수명이 `defer` 나 함수 말미의 제거문에 걸려 있어서는 안 된다 — 정리는 프레임워크 등록(`t.TempDir()` / `t.Cleanup`)으로만 보증되며, panic·조기 실패 경로에서도 제거돼야 한다.

```bash
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsFireValidationRejectNoWrites' -v -count=1
```

- **RED-now**: 증거 장부 **E7** — 판정 주체가 존재하지 않는다(빈 스윕, exit 0).
- 보조 관측(적색 아님): 오늘의 드라이버는 실저장소 루트를 서빙한다 — `grep -n 'ProjectRoot:' internal/web/appjs_fire_guard_test.go` → `204:		ProjectRoot:    findRepoRoot(t),`(exit 0, spec.md §B.6.3). 이 관측이 (a) 가 지금 성립하지 않음의 근거다.
- **green path**: M7 이 사본 서빙 배선·무쓰기 단언·프레임워크 등록 수명을 붙여 뒤집는다. (c) 방향은 합성 쓰기를 **쓰기 이음매가 도달 가능한 경로**에 넣어 exit 1 을 **관측**하는 것으로 확인한다 — 구체적으로 사본 루트 안 `.moai/config/sections/` 아래 파일 1바이트 변경이다(`handleSave` 의 `SyncToProjectConfig`·`writeProjectConfig` 가 닿는 자리, spec.md §B.6.3). 「임의 파일」이 아닌 이유는 card t1106 iter-2 의 D2 수리다: 제외 목록이 덮을 수 없는 자리에서 재야 반대 방향이 공허해지지 않는다 — 임의 파일에 넣으면 이음매가 닿는 하위 트리를 통째로 제외한 구현에서도 이 방향이 초록으로 나온다. 양방향 모두 보이지 않으면 이 AC 는 반쪽이다. (d) 는 사본 경로가 `t.TempDir()` 파생임을 단언하고, 드라이버 소스에 사본 경로를 운반하는 `defer`/말미 제거문이 없음을 함께 확인한다.

### AC-AFG-012 — 두 조건 없이는 `validation-reject` 를 주장할 수 없다 (card t1106)

**Given** 탐침의 매니페스트 자기검증(`--lint-manifest`)과 그 검증을 강제하는 Go 테스트가 있을 때,
**When** 자기검증을 실행하면,
**Then** (a) 커밋된 매니페스트는 통과(exit 0)해야 하고, (b) `validation-reject` 를 선언하면서 일회용 사본 서빙 표식 또는 무쓰기 단언 표식 중 하나라도 빠진 **합성 항목**은 exit 1 로 거부되며 어느 조건이 빠졌는지 이름으로 보고돼야 한다. 즉 두 조건과 효과 종별의 불가분성이 기계 판정이어야 한다.

```bash
python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest
go test ./internal/web/ -run 'AppJsFireSandboxPairing' -v -count=1
```

- **RED-now**: 증거 장부 **E8** — 불가분성을 판정하는 테스트가 존재하지 않는다(빈 스윕, exit 0). 같은 트리에서 `--lint-manifest` 는 exit 0 이지만(E9) 그 초록은 이 기준에 대해 **아무것도 주장하지 않는다** — 오늘의 허용목록에 `validation-reject` 가 없기 때문이다(`grep -c 'validation-reject' …` → `0`, exit 1).
- **green path**: M6 이 종별과 쌍 검증 규칙을, M7 이 Go 강제 테스트를 만들어 뒤집는다.
- **돌연변이 프로브**: 커밋된 매니페스트만 통과시키는 규칙(즉 (a) 만 보는 규칙)은 (b) 의 합성 항목을 놓치므로 채택 불가다 — 한 방향 판정식은 반대 방향의 변이를 통과시킨다(verification-completeness §2).

### AC-AFG-013 — 사본은 신설 항목만 서빙한다: 폭발 반경 0 (card t1106)

**Given** `startFireGuardServer` 가 서버를 하나 띄우고 매니페스트 전 항목이 그 하나를 쓰는 오늘의 구조와, 사본 서빙을 요구하는 신설 항목이 있을 때,
**When** 드라이버의 라우팅을 기계로 재면,
**Then** (a) 사본 서빙 표식이 **없는** 모든 항목의 서빙 루트가 여전히 `findRepoRoot` 결과여야 하고, (b) 표식이 **있는** 항목만 사본 서버로 라우팅돼야 하며(양방향 — 표식 없는 항목이 사본으로 가도, 표식 있는 항목이 실루트로 가도 적색), (c) 기존 전체 사이클(AC-AFG-001)이 서빙 루트 변경 없이 그대로 통과해야 한다.

이 AC 가 막는 것은 구체적이다: `ProjectRoot` 를 사본으로 통째 바꾸면 `glm_reveal`(시드된 자격증명 상태 의존), `swap_todo_nav`·`popover_after_swap`(`/todo`), `copy_button`(`/specs`) 이 전부 다른 트리를 보게 된다. 그때의 초록은 **틀린 트리를 잰 초록**이다.

```bash
go test ./internal/web/ -run 'AppJsFireSandboxRouting' -v -count=1
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -v -count=1
```

- **RED-now**: 증거 장부 **E10** — 라우팅 판정 주체가 존재하지 않는다(빈 스윕, exit 0).
- 보조 관측(적색 아님): 오늘은 서버가 하나이고 전 항목이 `findRepoRoot` 를 본다 — `grep -n 'ProjectRoot:' internal/web/appjs_fire_guard_test.go` → `204:		ProjectRoot:    findRepoRoot(t),` (단 1행, exit 0). 이 「1행」이 폭발 반경 문제의 근거다.
- **green path**: M7 이 두 번째 서버 인스턴스와 표식 기반 라우팅을 붙여 뒤집는다.
- **돌연변이 프로브**: 「사본 서버가 존재한다」만 보는 판정식은, 전 항목이 사본으로 가는 변이(= 안 (b) 를 몰래 취한 상태)를 통과시킨다 — 따라서 (a) 와 (b) 를 **양방향으로** 걸지 않은 판정식은 채택 불가다.

---

## §B2 RED-now 증거 장부 (evidence ledger)

§2.1의 4요소 — 명령(단일 호출), 원문 출력, exit 코드, 트리 SHA — 을 전부 갖춘 항목들. 전부 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1060`, branch `WT-appjs-handler-guard`, HEAD **`d726ac709`** 에서 2026-09-22 이번 iter-2 실행 중 측정됐다.

**E1** — AC-001·AC-003 공용. 명령:

```
go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.495s [no tests to run]
```

exit 코드: `0`. 트리: `d726ac709`. 적색의 성질: 기준의 판정 주체(드라이버 테스트)가 존재하지 않는다 — 러너 자신의 `[no tests to run]` 토큰이 그 빈 스윕을 대신 증언한다(§1.1: 빈 스윑의 초록은 통과가 아니라 미측정이다).

**E2** — AC-004. 명령:

```
go test ./internal/web/ -run 'AppJsHandlersFireSelectorMiss' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.330s [no tests to run]
```

exit 코드: `0`. 트리: `d726ac709`.

**E3** — AC-007. 명령:

```
ls internal/web/testdata/appjs_fire_mutation.py
```

원문 출력:

```
ls: internal/web/testdata/appjs_fire_mutation.py: No such file or directory
```

exit 코드: `1`. 트리: `d726ac709`.

**E4** — AC-009. 명령:

```
ls internal/web/testdata/appjs_fire_probe.py
```

원문 출력:

```
ls: internal/web/testdata/appjs_fire_probe.py: No such file or directory
```

exit 코드: `1`. 트리: `d726ac709`.

**E5** — AC-002 (적색) + AC-007 보조 baseline. 명령(탐침 실행 — 서버·Chrome 기동과 전제는 plan §B 스텝 시퀀스, 돌연변이 합성 `MUTATED removed@727 inserted@552` 확인 뒤):

```
python3 /tmp/t1060-probe/probe.py 18442 t1060-mutation
```

원문 출력:

```json
{
  "label": "t1060-mutation",
  "port": "18442",
  "p1_load_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": true,
  "p2_glm_handler_fired": false,
  "p3_swap_clicked": true,
  "p3_swap_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p3_url_after_swap": "/todo",
  "p4_load_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p3_has_copy_btn": true,
  "p4_label_after_click": "Copy",
  "p4_label_before_click": "Copy",
  "p4_copy_handler_fired": false
}
```

exit 코드: `0` — **적색이다**: AC-AFG-002 는 exit 1 을 요구하며, 지표는 전부 붕괴했는데도 탐침이 통과 코드를 낸다. 붕괴는 실측이고 exit 미달은 그 원인(exit 계약 미구현)이 분명한 올바른-이유 적색이다. 트리: `d726ac709`. 같은 실행의 복원 구간: `RESTORED_BYTE_IDENTICAL` (`cmp` exit 0), `git status --porcelain -- internal/web/assets/` 빈 출력 — AC-007 의 보조 baseline 관측.

### §B2.1 개정분(card t1106) RED-now 항목

E6~E9 는 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1106`, branch `WT-fireguard-reject-submit`, HEAD **`176d8b658`** 에서 2026-09-23 개정 작성 중에 측정했다. E1~E5 의 트리(`d726ac709`)와 다른 트리이므로 서로 섞어 인용하지 않는다.

**E6** — AC-010. 명령:

```
go test ./internal/web/ -run 'AppJsFireValidationRejectPaints' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.424s [no tests to run]
```

exit 코드: `0`. 트리: `176d8b658`. 적색의 성질: 기준의 판정 주체가 존재하지 않는다 — 러너 자신의 `[no tests to run]` 토큰이 빈 스윕을 증언한다. 빈 스윕의 초록은 통과가 아니라 미측정이다(verification-completeness §1.1).

**E7** — AC-011. 명령:

```
go test ./internal/web/ -run 'AppJsFireValidationRejectNoWrites' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.429s [no tests to run]
```

exit 코드: `0`. 트리: `176d8b658`.

**E8** — AC-012. 명령:

```
go test ./internal/web/ -run 'AppJsFireSandboxPairing' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.428s [no tests to run]
```

exit 코드: `0`. 트리: `176d8b658`.

**E9** — AC-012 보조(적색 아님 — 오늘의 초록이 이 기준에 대해 아무것도 주장하지 않음의 근거). 명령:

```
python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest
```

원문 출력:

```
LINT OK: 8 entries + 7 exclusions cover 13 inventory groups; all effects within ['clipboard', 'label', 'swap', 'tab', 'visibility']; post-swap entry present
```

exit 코드: `0`. 트리: `176d8b658`. 이 초록은 허용목록에 `validation-reject` 가 없는 상태에서 나온 것이므로(`grep -c 'validation-reject' internal/web/testdata/appjs_fire_probe.py` → `0`, exit `1`, 같은 트리), 불가분성 기준에 대해서는 미측정이다.

**E10** — AC-013. 명령:

```
go test ./internal/web/ -run 'AppJsFireSandboxRouting' -v -count=1
```

원문 출력:

```
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.646s [no tests to run]
```

exit 코드: `0`. 트리: `176d8b658`.

---

## §C Definition of Done

분류를 따른다 — blocking AC 의 통과는 이 SPEC 의 run-phase 완료의 차단 요건이고, regression-class AC 의 통과는 **기록되는** 산출물이지 차단 요건이 아니다(§D 선언과 정합).

blocking (AC-AFG-001, 002, 003, 004, 007, 009, 010, 011, 012, 013):

- [ ] 10건 전부 통과, 각각 판정 명령의 **출력 그대로**를 progress.md §E.2 에 기록
- [ ] AC-002 의 0→1→0 시퀀스 전 구간 출력 보존(복원 검증 포함)
- [ ] AC-011 의 (c) 방향 — 사본 루트 안 `.moai/config/sections/` 아래(쓰기 이음매가 닿는 자리) 파일 1바이트 합성 변경에서 탐침 exit 1 을 **관측**한 출력 보존(양방향 없이는 반쪽이다)
- [ ] AC-009 정방향(커밋된 매니페스트 exit 0)과 AC-012 역방향(조건이 빠진 합성 항목 exit 1 + 빠진 조건 이름) **둘 다** 관측한 출력 보존 — 정방향만으로는 아무것도 거부하지 않는 판정식이 통과한다
- [ ] AC-013 의 양방향 — 표식 없는 항목이 사본으로 가는 변이와 표식 있는 항목이 실루트로 가는 변이 **둘 다** 적색임을 관측한 출력 보존
- [ ] AC-011 (d) — 사본 경로가 `t.TempDir()` 파생임의 단언 출력과, 드라이버에 사본 수명을 운반하는 `defer`/말미 제거문이 없음의 확인

regression-class (AC-AFG-005, 006, 008):

- [ ] 세 건의 판정 명령 출력을 §E.2 에 기록한다(차단 요건 아님 — §D 분류 선언대로)

공통:

- [ ] `go vet ./internal/web/` 통과; `gofmt -l internal/web/appjs_fire_guard_test.go` 출력 없음
- [ ] `git status --porcelain` — `internal/web/assets/` 변경 0건
- [ ] `go.mod`/`go.sum` 변경 0건 (탐침 의존은 `websockets` 단 하나, Go 쪽 신규 없음)
- [ ] 마일스톤 집행 순서는 plan §E (M1→M7) 를 따른다 — green path 귀속(AC-004→M1/M2, AC-001·003→M2, AC-002→M1·M3, AC-007·009→M1·M3, AC-005·006·008→M4·M5, AC-010·012→M6·M7, AC-011·013→M7)과 충돌하는 순서 없음

---

## §D 미검증으로 남는 것 (Gaps) + 분류 선언

이 인수조건이 **주장하지 않는 것**과, two-cell 규율상 셀 없이 채택되는 기준의 분류를 명시한다.

**regression-class 선언 (RED-now 셀 없음 — 구조적 이유와 함께):**

- **AC-AFG-005(회귀·범위), AC-AFG-006(CI 고립), AC-AFG-008(문서)** — 셋의 위반 방향은 구현이 존재한 뒤에만 성립한다(회귀는 구현 뒤에야 회귀이고, CI diff 는 job 이 추가된 뒤에야 재고, 문서는 소스가 쓰인 뒤에야 검사된다). plan-phase 트리에서 만들 적색 관측이 없으므로 셋은 RED-now 셀 없이 채택되며, §2.1의 귀결에 따라 **release-blocking 자격을 갖지 않는다** — 통과는 기록되지만 차단하지 않는다. 이 선언은 iter-1 에서 누락됐다가(iter-1 감사 D2) iter-2 에서 보태졌다.
- iter-1 의 §D 가 AC-003/004의 RED-now 를 「관측 불가」로 기록한 것은 **틀린 주장이었다** — 빈 스윕은 관측 가능한 상태다. iter-2 에서 E1/E2 로 직접 재측정해 셀을 저작했고, 해당 문장을 이 선언으로 교체했다.

**여전히 열려 있는 간극:**

- **시각적 회귀·접근성·타 브라우저는 재지 않는다** — spec.md §E 제외 범위다.
- **매니페스트에 없는 상호작용**은 재지 않는다. 제외 사유의 품질이 그대로 가드의 품질이다.
- **돌연변이가 지표를 못 뒤집는 다른 결함 계열**(예: 스왑 뒤에만 죽는 재결합 결함)에 대해 이 가드가 잡는다는 보장은, 매니페스트가 그 경로를 행사할 때만 성립한다.
- **러너(ubuntu-latest)의 Chrome 사전 설치 여부는 검증하지 않았다** — 그래서 REQ-AFG-010 이 고정 다운로드를 요구한다.
- **AC-AFG-002 의 green 통과 형태(exit 1)는 아직 관측되지 않았다** — E5 는 그 적색일 뿐이고, 뒤집는 것은 M1(exit 계약)·M3(돌연변이기)의 몫이다.
- **배너의 문구·대비·포커스 이동은 재지 않는다**(card t1106 개정분) — AC-010 은 「보이는 자리에 비지 않은 텍스트로 도달했는가」까지다(spec.md §F 5).
- **무쓰기 단언은 일회용 사본 루트 밖을 보지 않는다** — 프로필 스토어·임시 디렉터리·프로세스 환경에 남는 흔적은 AC-011 의 범위 밖이다(spec.md §F 6).
- **두 서버 서면의 표면 등가는 어느 AC 도 주장하지 않는다** — 사본 서버는 거부 경로 한 갈래만 운반한다(spec.md §F 7). 안 (b)(전 항목 사본 서빙)를 택했다면 기존 8항목의 사본 위 무회귀가 인수조건이 됐겠지만, §C 개정 결정에서 (a) 를 택했으므로 그 측정은 필요하지 않고 **하지도 않았다**.
- **성공 저장 경로는 어느 AC 도 행사하지 않는다** — 개정분의 `validation-reject` 항목은 거부 경로 한 갈래만 운반한다.
- **E5 의 서버·Chrome 기동 단계는 장부 명령의 전제다** — 탐침 단일 호출이 단일 관측점이고, 기동 절차는 plan §B 에 고정돼 있다.
