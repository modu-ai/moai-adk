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
| AC-AFG-009 | 영속화 안전 | blocking | REQ-AFG-012 |

iter-2 수리로 매핑이 갈린 곳: REQ-AFG-007 → AC-001(c) 가 실질 검증(스왑 뒤 지표 행사), REQ-AFG-002 → AC-005(커밋된 탐침의 존재·파생 출처·go.mod 무변경), REQ-AFG-012 → 신설 AC-009.

---

## §B Given-When-Then

### AC-AFG-001 — 정방향: 통과하되 공허하지 않게

**Given** `MOAI_BROWSER_GUARD=1` 과 Chrome·python3·websockets 가 모두 있는 환경, 그리고 이 트리의 `internal/web` 패키지가 있을 때,
**When** 드라이버 테스트를 실행하면,
**Then** exit 0 이고 **그리고 동시에** (a) 매니페스트 항목 수가 0보다 크며 전 항목이 발화했고, (b) load·swap ReferenceError 가 0건이며, (c) hx-boost 스왑 **뒤에** 최소 1지표가 발화했음을 보고서가 담고 있어야 한다.

세 단언이 **함께** 걸려야 한다. 지표 단언만 있으면 빈 매니페스트가 통과하고, ReferenceError 단언만 있으면 버튼이 안 눌려도 통과한다.

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
**Then** 직교성 문장과 네 한계 축을 가리키는 앵커 토큰 — `orthogonal`(정적 형제와의 직교성, `SPEC-APPJS-IIFE-GUARD-001` 명시와 함께), `manifest`(매니페스트 밖 미측정), `scenario`(단일 시나리오), `Chrome`(단일 브라우저), `static-scope`(원인 판정의 정적 가드 소관) — 이 **모두** 소스에 등장해야 한다.

```bash
grep -c 'orthogonal' internal/web/appjs_fire_guard_test.go internal/web/testdata/appjs_fire_probe.py
grep -c 'SPEC-APPJS-IIFE-GUARD-001' internal/web/appjs_fire_guard_test.go
grep -c 'manifest' internal/web/testdata/appjs_fire_probe.py
grep -c 'scenario' internal/web/testdata/appjs_fire_probe.py
grep -c 'Chrome' internal/web/testdata/appjs_fire_probe.py
grep -c 'static-scope' internal/web/appjs_fire_guard_test.go
```

각 기대치는 1 이상(주석 안의 등장이면 충분하다). 토큰 집합은 plan.md §E M1 이 소유한다.

- 분류: **regression-class** — RED-now 셀 없음(§D 선언). green path: M5.

### AC-AFG-009 — 매니페스트는 영속화 부수효과를 행사하지 않는다

**Given** 커밋된 탐침과 그 (페이지, 셀렉터, 효과) 매니페스트가 있을 때,
**When** 탐침의 매니페스트 자기검증 모드를 실행하면,
**Then** 모든 항목의 효과 종별이 비영속 허용목록 {`visibility`, `label`, `clipboard`, `tab`, `swap`} 안에 있어 exit 0 이어야 하고, 허용목록 밖(저장·제출·설정 쓰기 계열)의 항목이 하나라도 있으면 exit 1 이어야 한다. 저장 계열을 행사할 필요가 생기면 드라이버는 일회용 프로젝트 사본 위에서만 서빙한다(소스의 샌드박스 루트 배선으로 검증).

```bash
python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest
```

- **RED-now**: 증거 장부 **E4** — 탐침 파일 자체가 존재하지 않는다(파일 부재, exit 1). 데이터 손실 경로를 막는 이 기준이 현재 아무것도 검증하지 않는 상태가 관측이다(iter-1 감사가 지적한 REQ-AFG-012 무검증의 해소 지점).
- **green path**: M1 — 매니페스트 효과 종별 허용목록 + `--lint-manifest` 자기검증 구현.

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

---

## §C Definition of Done

분류를 따른다 — blocking AC 의 통과는 이 SPEC 의 run-phase 완료의 차단 요건이고, regression-class AC 의 통과는 **기록되는** 산출물이지 차단 요건이 아니다(§D 선언과 정합).

blocking (AC-AFG-001, 002, 003, 004, 007, 009):

- [ ] 6건 전부 통과, 각각 판정 명령의 **출력 그대로**를 progress.md §E.2 에 기록
- [ ] AC-002 의 0→1→0 시퀀스 전 구간 출력 보존(복원 검증 포함)

regression-class (AC-AFG-005, 006, 008):

- [ ] 세 건의 판정 명령 출력을 §E.2 에 기록한다(차단 요건 아님 — §D 분류 선언대로)

공통:

- [ ] `go vet ./internal/web/` 통과; `gofmt -l internal/web/appjs_fire_guard_test.go` 출력 없음
- [ ] `git status --porcelain` — `internal/web/assets/` 변경 0건
- [ ] `go.mod`/`go.sum` 변경 0건 (탐침 의존은 `websockets` 단 하나, Go 쪽 신규 없음)
- [ ] 마일스톤 집행 순서는 plan §E (M1→M5) 를 따른다 — green path 귀속(AC-004→M1/M2, AC-001·003→M2, AC-002→M1·M3, AC-007·009→M1·M3, AC-005·006·008→M4·M5)과 충돌하는 순서 없음

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
- **E5 의 서버·Chrome 기동 단계는 장부 명령의 전제다** — 탐침 단일 호출이 단일 관측점이고, 기동 절차는 plan §B 에 고정돼 있다.
