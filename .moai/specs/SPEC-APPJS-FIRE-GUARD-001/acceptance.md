# SPEC-APPJS-FIRE-GUARD-001 — 인수조건

모든 AC 는 기계적으로 확인 가능하다. 각 AC 는 판정 명령을 함께 적는다. RED-now 셀은 **이 plan-phase 프로토타입이 이 트리에서 이미 관측한 것**, green path 셀은 **어느 마일스톤이 무엇으로 뒤집는지**다(verification-completeness §2 two-cell 규율).

---

## §A AC 매트릭스

| AC | 방향 | 성질 | 대응 REQ |
|---|---|---|---|
| AC-AFG-001 | 정방향 | 게이트 켠 드라이버가 실트리를 통과시키되 **공허하지 않게** | REQ-AFG-001, 003, 005, 006 |
| AC-AFG-002 | 역방향 | 돌연변이를 exit 1 로 거부하고 **이름을 지목한다** | REQ-AFG-008 |
| AC-AFG-003 | 경계 | 전제 부재의 skip 이 **사유를 이름으로** 대는가 | REQ-AFG-001 |
| AC-AFG-004 | 생존 | 셀렉터 스테일이 red 로 보이는가 | REQ-AFG-004 |
| AC-AFG-005 | 회귀 | 기존 계약(정적 형제 포함)을 깨지 않는다 | REQ-AFG-011, 013 |
| AC-AFG-006 | CI 고립 | 기존 job 전제가 1바이트도 안 움직인다 | REQ-AFG-010 |
| AC-AFG-007 | 복원 | 돌연변이가 트리에 흔적을 남기지 않는다 | REQ-AFG-009 |
| AC-AFG-008 | 문서 | 한계와 직교성이 소스에 적혀 있다 | REQ-AFG-013 |

---

## §B Given-When-Then

### AC-AFG-001 — 정방향: 통과하되 공허하지 않게

**Given** `MOAI_BROWSER_GUARD=1` 과 Chrome·python3·websockets 가 모두 있는 환경, 그리고 이 트리의 `internal/web` 패키지가 있을 때,
**When** 드라이버 테스트를 실행하면,
**Then** exit 0 이고 **그리고 동시에** (a) 매니페스트 항목 수가 0보다 크며 전 항목이 발화했고, (b) load·swap ReferenceError 가 0건이며, (c) 스왑 뒤 최소 1지표가 발화했음을 보고서가 담고 있어야 한다.

세 단언이 **함께** 걸려야 한다. 지표 단언만 있으면 빈 매니페스트가 통과하고, ReferenceError 단언만 있으면 버튼이 안 눌려도 통과한다.

- RED-now(프로토타입, 이 트리, 2026-09-22): 같은 표면을 예시 탐침으로 재자 — 5 지표 전부 발화 + ReferenceError 0건(spec.md §B.1). 가드 형식(드라이버+exit 계약)은 M1/M2 가 만든다.
- green path: M2 가 (a)(b)(c) 를 Go 단언으로, M1 이 exit 계약을 탐침에 실어 뒤집는다.

```bash
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFire' -v -count=1
```

### AC-AFG-002 — 역방향: 돌연변이를 거부하고 지목한다

**Given** §B.2 의 돌연변이(727행 등록의 IIFE#0 이동)가 적용되고 재빌드된 바이너리가 서빙될 때,
**When** 탐침을 실행하면,
**Then** exit **1** 이어야 하고, 보고가 최소 1개 지표의 미발화 또는 `stampRefreshed` ReferenceError 를 **이름으로** 담아야 하며, 돌연변이기 자신은 대상 줄을 못 찾으면 exit 비0 으로 실패해야 한다.

- RED-now(프로토타입): **이 방향은 plan-phase에서 이미 이 트리에서 관측됐다** — glm·copy 지표 붕괴 + 3Phase ReferenceError `app.js:552`(spec.md §B.2). 단, 관측 당시 탐침은 보고만 하고 exit 는 0 이었다 — exit 1 계약은 M1 이 추가한다.
- green path: M3 가 돌연변이기+스텝 시퀀스로, M4 가 CI 레드 단계로 기계화한다.

```bash
bash <돌연변이 스텝 시퀀스>   # probe exit: 0 → 1 → (복원) → 0 을 출력으로 남긴다
```

### AC-AFG-003 — 전제 부재의 skip 은 사유를 이름으로 댄다

**Given** `MOAI_BROWSER_GUARD` 가 설정되지 않은(또는 Chrome 이 없는) 평상 환경일 때,
**When** 패키지 테스트를 실행하면,
**Then** 드라이버 테스트는 SKIP 이고 그 사유에 **빠진 전제의 이름**(게이트 변수명 또는 chrome/python3/websockets)이 적혀 있어야 한다.

```bash
go test ./internal/web/ -run 'AppJsHandlersFire' -v | grep -E 'SKIP|skip'
# 사유 행에 게이트 변수 또는 부재 전제 이름이 보여야 한다
```

- RED-now: 없음(테스트 미존재) — **Gap 으로 기록한다**, 프로토타입으로 대체 관측 불가.
- green path: M2.

### AC-AFG-004 — 셀렉터 스테일은 red 다

**Given** 매니페스트의 셀렉터 하나를 고의로 실재하지 않는 값으로 바꾼 실행(일회용 복사본에 대해)이 있을 때,
**When** 탐침을 돌리면,
**Then** exit 1 이고 보고가 **그 셀렉터 이름**을 미달 항목으로 지목해야 한다.

- RED-now: 없음 — Gap(exit 계약 자체가 미구현).
- green path: M1 이 셀렉터 생존 검사를, 본 AC 의 판정은 M1 셀프검증으로 실시한다.

### AC-AFG-005 — 기존 계약 회귀 없음

**Given** `internal/web` 패키지의 기존 테스트 전부가 있을 때,
**When** 이 SPEC 의 신규 파일을 더한 뒤 패키지 테스트를 돌리면,
**Then** 전부 통과하고, `appjs_iife_scope_test.go`·`appjs_reinit_test.go` 는 **수정되지 않은 채로** 통과해야 한다.

```bash
go test ./internal/web/ -count=1
git diff --stat -- internal/web/appjs_iife_scope_test.go internal/web/appjs_reinit_test.go   # 기대: 출력 없음
```

> `internal/web` 는 크다. FAIL 요약이 실패 0건과 함께 나오면 타임아웃을 의심하고 `-timeout 30m` 으로 재측정한다(과거 실측 기록 있음).

### AC-AFG-006 — 기존 CI 전제 무변경

**Given** 이 SPEC 의 구현 diff 가 있을 때,
**When** 변경 파일 목록을 재면,
**Then** `.github/workflows/` 아래 변경은 `ci.yml` 의 **job 추가 블록뿐**이고, 기존 job 키(detect/test/test-race/test-skip-marker/test-integration/lint/build/constitution-check)의 본문은 불변이어야 한다.

```bash
git diff <base>..HEAD -- .github/workflows/ci.yml
# 기존 job 키 위치의 행이 바뀌지 않았는지 눈으로 + 신설 job 블록만 추가됐는지 확인
```

범위 판정식은 흡수 ref 기준 `merge-base`로 잡고(배차 규율), 귀속은 `git log <base>..HEAD -- <경로>` 커밋 열거로 답는다 — 두 끝점 diff 로 단정하지 않는다.

### AC-AFG-007 — 돌연변이는 트리에 흔적을 남기지 않는다

**Given** 레드 단계가 끝난 job(또는 로컬 재현)이 있을 때,
**When** 트리 상태를 재면,
**Then** `internal/web/assets/app.js` 는 원본과 byte 동일(`cmp` 종료 0)이고 `git status --porcelain` 에 assets 변경이 없어야 한다.

```bash
cmp internal/web/assets/app.js <원본 사본> && echo IDENTICAL
git status --porcelain -- internal/web/assets/
```

- RED-now(절차 실측): 프로토타입이 이미 `RESTORED_BYTE_IDENTICAL` 을 관측했다(spec.md §B.2).

### AC-AFG-008 — 한계와 직교성이 소스에 적혀 있다

**Given** 완성된 드라이버·탐침 소스가 있을 때,
**When** 상단 주석을 읽으면,
**Then** spec.md §F 의 네 한계와 「정적 가드와 직교한다 — 어느 쪽 초록도 다른 쪽을 함의하지 않는다」라는 문장이 모두 서술돼 있어야 한다.

---

## §C Definition of Done

- [ ] AC-AFG-001 ~ 008 전부 통과, 각각 판정 명령의 **출력 그대로**와 함께 기록
- [ ] `go test ./internal/web/ -count=1` 통과 (종료 코드 0, 요약 행 `ok`)
- [ ] `go vet ./internal/web/` 통과; `gofmt -l internal/web/appjs_fire_guard_test.go` 출력 없음
- [ ] `git status --porcelain` — `internal/web/assets/` 변경 0건
- [ ] `go.mod`/`go.sum` 변경 0건 (탐침 의존은 `websockets` 단 하나, Go 쪽 신규 없음)
- [ ] progress.md §E.2 에 위 명령들의 출력과 baseline 귀속(HEAD SHA) 기록

---

## §D 미검증으로 남는 것 (Gaps)

이 인수조건이 **주장하지 않는 것**을 명시한다.

- **시각적 회귀·접근성·타 브라우저는 재지 않는다** — §E 제외 범위다.
- **매니페스트에 없는 상호작용**은 재지 않는다. 제외 사유의 품질이 그대로 가드의 품질이다.
- **돌연변이가 지표를 못 뒤집는 다른 결함 계열**(예: 스왑 뒤에만 죽는 재결합 결함)에 대해 이 가드가 잡는다는 보장은, 매니페스트가 그 경로를 행사할 때만 성립한다.
- **AC-AFG-003/004 의 RED-now 셀은 없다** — 테스트 미존재로 관측 불가. two-cell 규율상 이 둘은 green path(M1/M2)만으로 채택됐음을 여기에 기록한다.
- **러너(ubuntu-latest)의 Chrome 사전 설치 여부는 검증하지 않았다** — 그래서 REQ-AFG-010 이 고정 다운로드를 요구한다. 검증되지 않은 전제 위에 설계를 세우지 않는다.
