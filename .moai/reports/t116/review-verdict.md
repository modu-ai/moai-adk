# t116 Review Verdict — PASS (실측 발견 1건 조건부 라이더 동반)

- reviewer: review lane session (release-v311 hub), 2026-08-17
- target: WT-t116 @ 265bcba0a (+evidence 1f331cee3; base = release tip 091a42e16)
- diff scope: `internal/hook/home_isolation_test.go` (allowlist 전환) + `internal/hook/navigator_detect_nonoverlap_test.go` (baseline 정책) — test-only, production 무변경
- lenses: 기본 4-관점 + 가드 보안성 검증 (lead dispatch)

## Verdict

**PASS** — 배치 PR 통합 승인. 단, 이번 리뷰가 실측으로 발견한 **tip 전진 경쟁(R1)** 을 반드시 숙지하고 통합할 것: 배치 PR/통합 경로에는 무영향이지만, evidence의 부수 주장은 tip 전진과 함께 시효 만료됐다.

## 이번 리뷰 핵심 실측 발견 — R1: release tip 전진 경쟁

**재현 사건**: 이 리뷰 진행 중 `origin/release/v3.1.1` tip이 `091a42e16 → 0151cca71`로 전진함(0151cca71 = **t36 통합 머지** — 직전 리뷰에서 PASS 판정한 카드의 release 반영). 결과:

- t116 HEAD는 091a42e16 위 커밋이므로 새 tip(0151cca71)의 **조상이 아님** → 정책상 "not HEAD's boundary"(distance -1) → baseline이 **origin/main으로 회귀** → 가드가 다시 확정 붉음:
  ```
  AC-NS2-005a FAIL: diff vs origin/main touches mx producer path
  "internal/mx/spec_loader.go" / "internal/mx/spec_loader_test.go"
  ```
- 즉 evidence Claim 2의 부수 주장("release tip을 머지한 모든 레인 브랜치에서 확정 붉음 해소")은 **tip이 전진하지 않은 한**에서만 성립. 카드 작성 시점(tip=091a42e16)에는 참이었으나 지금 재현하면 붉음.

**배치 PR·통합 경로 무영향 입증** (카드의 선결 목표는 그대로 달성):

| 경로 | baseline | 결과 |
|---|---|---|
| 배치 PR head (= push된 release tip) | tip 자신 (is-ancestor reflexivity, distance 0) → 0 < mainDistance 치환 | diff = PR 자신 커밋분 → mx 무관 → **PASS** |
| 리드가 t116을 release에 머지한 커밋 (push 전 로컬) | origin ref = 0151cca71 = 새 머지의 조상, distance(카드분) ≪ mainDistance(172+) | diff = t116 머지분(테스트 2파일) → **PASS** |
| 구 tip만 머지한 레인 워크트리 (현재 t116) | tip 비조상 → origin/main 회귀 | **FAIL 재래** (실측) — 단 레인 push 금지 규율상 CI가 이 경로를 돌지 않음 |

**완화**: ① 통합 직전 base refresh 관행(이미 t108/t36이 "sync release tip into WT-*" 패턴으로 실행 중)이 운영 해법 — refresh하면 tip이 조상이 되어 치환 복원. ② (선택) 정책의 baseline 후보를 "origin ref가 가리키는 tip"에서 "HEAD에 머지된 release 경계 커밋"으로 일반화하면 경쟁 자체를 제거 가능 — 별도 카드 규모.

## 지정 검증 6건

### 1. ① allowlist — ✅ 완전 재현

- 집합 동등성 양방향: `unlisted`(코드 有/목록 無)·`stale`(목록 有/코드 無) 분리 검출 + 각 방향 대응 안내가 실패 메시지에 명시. 누수 양방향 재현코드 확인.
- `filepath.ToSlash` 정규화(diff에서 확인) — CI darwin/windows 매트릭스 경로 구분자 정합.
- 함수명 `TestHomeJoinSiteCountIsPinned` 유지 — 2개 파일 주석 참조 회손 없음(diff에서 참조 대상 메시지 갱신 확인). 개명 스코프 초과 사유가 evidence residual-risk에 명시.
- 풀수트(23.5s)에서도 PASS — allowlist 자체는 경쟁과 무관하게 초록.
- tokens.go 행의 "glob READ" 주석: 읽기 사이트를 쓰기 가드 클래스와 구분하되 미래의 읽기→쓰기 전환 시 재검토를 여는 설계 — 타당.

### 2. ② baseline 정책 순수함수 — ✅ 재현

`chooseConsumerOnlyBaseline` 5케이스 테이블 전부 재실행 PASS: 무 release ref→main / 비조상(distance -1) 무시 / 조상+근접 채택 / 동점 main 유지(최대 강도) / 다수 중 최근접. evidence 주장과 일치.

### 3. 핀 강도 보안 검토 (핵심) — (a) 명시 확인 ✓, (b) 미검토 → residual-risk 기록

- (a) evidence Gaps에 이미 명시: "직접 커밋이 release에 들어오면 이 가드의 관할 밖 — 배치 PR 시점엔 tip 이후 커밋은 전부 잡힘". **무리뷰 push가 tip이 되면 baseline으로 흡수** 경로가 문서화돼 있음.
- (b) 완화 여지(치환 조건에 머지 커밋 `--no-ff`+review 표기 요건)는 evidence/주석에 검토 흔적 없음 → 아래 residual-risk로 기록.
- 리뷰어 덧붙임: 머지 커밋 요건도 부분 완화일 뿐이다 — (i) 무리뷰 머지 push도 가능하고 (ii) tip refresh sync 머지(`merge: sync release/vX.Y.Z tip … into WT-*`)와 review 탑재 머지(`merge: tN — … review-PASS`)가 메시지 구조로 완전 구분되지 않는다. 기계적 완화의 한계가 있으므로 "push = 리뷰 동반" 운영 규율 의존을 유지하고 이 한계를 문서로 남기는 것이 실질적인 최선 — 현 카드가 취한 입장과 동일.

### 4. 배치 PR 시뮬레이션 — 부분 재현 (신가드 PASS는 경쟁으로 재현 불가)

- 구가드 붉음 입력 ✓: `origin/main...WT-t116` = 172커밋, mx 2파일(`spec_loader{,_test}.go`) grep 적중.
- 신가드 치환→PASS: evidence 시점(tip=091a42e16)엔 성립했으나 **지금은 tip 전진으로 재현 불가** (R1). 치환 산술 자체는 probe로 실측 지지: is-ancestor rc=0일 때 distance 2 < mainDistance 172 → 치환 (tip이 조상인 조건에서 정책이 정확히 작동함을 단계별 관측).

### 5. 프로브(tip 위 mx 변경 FAIL) — 완전 재현 불가, 코드 경로 관찰로 대체

tip이 baseline으로 선택된 상태(t116이 새 tip을 머지한 상태)를 리뷰 세션에서 만들 수 없어 evidence 프로브의 FAIL을 그대로 재현하지 못함. 단, `diff <sha>..HEAD → internal/mx/ prefix → FAIL(chosen.name 명시)` 코드 경로는 현재 FAIL 출력(동일 라인 navigator_detect_nonoverlap_test.go:317)이 산출한 메시지 형식과 동일해 경로가 살아있음을 관측. 갭 기록.

### 6. 재실행 — 타깃 3·풀수트·vet·lint

| 검증 | 결과 |
|---|---|
| TestHomeJoinSiteCountIsPinned | ✅ PASS |
| TestChooseConsumerOnlyBaseline (5 서브케이스) | ✅ PASS |
| TestConsumerOnly_M0AndMxByteUnchanged | ❌ FAIL — **R1 경쟁 재래** (evidence 시점 PASS와 다른 이유 = tip 전진, 카드 코드 결함 아님) |
| hook 풀수트 `-count=1 -timeout 300s` | 23.5s, FAIL 함수 정확히 1개(위 가드) — evidence "ok"와 상이한 이유 동일 |
| `go vet ./internal/hook/` | ✅ rc=0 |
| `golangci-lint run ./internal/hook/...` | ✅ 0 issues |

## 4-관점 평가

| 관점 | 판정 | 근거 |
|---|---|---|
| Functionality | PASS | ① 완달성(풀수트 포함). ② 배치 PR 선결 목표 달성 — head=tip에서 치환 산술·정책 테스트로 입증. 부수 주장(레인 완전 해소)은 R1 시효 만료 — 목표 외 |
| Security | PASS | 무리뷰 push 흡수 경로 evidence 명시; 미푸시분 보수 취급·동점 main(최대 강도) 원칙 유지; baseline은 명시 sha 2-dot이라 우연 일치 없음. 기계 완화 한계는 residual-risk로 남김 |
| Craft | PASS | 순수 정책 함수 + 테이블 테스트 분리, 실패 메시지에 baseline 출처(`chosen.name`) 명시로 관찰 가능, allowlist 동기(개수→집합) 문서화 충실. 니트: `containsSite` 수제 헬퍼 → `slices.Contains` 표준 대체 여지; mainDistance(3-dot symmetric)와 release distance(2-dot one-sided) 측정 단위 비대칭 — 편향 방향은 치환 촉진이나 오류는 false-red(보수) 방향 |
| Consistency | PASS | 주석·에러 전부 영어, 기존 skip 경로(환경변수·origin/main 부재) 보존, 원본 AC verbatim 형태를 default로 유지하는 확장 구조 |

## Gaps (이번 리뷰 미재현)

- 신가드 baseline 치환→PASS의 라이브 재현(4)·tip 위 mx 프로브(5) — tip 전진으로 재현 불가. 배치 PR CI가 최종 확인.
- 평범한 피처 브랜치(release 미포함)에서의 라이브 실행 — evidence와 동일 갭, 정책 단위테스트로 대체 입증.
- CI 매트릭스(배치 PR) 실측 — 카드 범위 밖.

## 라이더

- **R1 (중요 — 이번 리뷰 발견, 위 본문)**: tip 전진 경쟁. 통합 전 base refresh(관행)로 해소되고 배치 PR엔 무영향이나, 리드가 t116 워크트리에서 재검증 시 확정 붉음을 보게 됨 — 카드 결함 아님을 인지할 것. 근본 제거(경계 커밋 일반화)는 후속 카드 후보.
- **R2 (residual-risk, 리드 지시 기록)**: 무리뷰 push의 baseline 흡수 — 기계 완화(머지 커밋 요건)는 부분적이며 sync/review 머지 구분 불가로 한계 존재. 운영 규율 의존 유지를 권고.
- **R3 (사소)**: `containsSite` → `slices.Contains`; distance 측정 단위 비대칭 주석화.
