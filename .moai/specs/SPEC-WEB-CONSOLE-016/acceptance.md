# SPEC-WEB-CONSOLE-016 — Acceptance Criteria

> 각 AC 는 명령으로 반증 가능하고, REQ 에 매핑되며, **현재 코드에서 무엇 때문에 실패하는지**를 함께 적는다.
> 모든 명령은 워크트리 루트에서 실행한다.
> 기준선: 트리 `WT-partial-apply` @ `d8304b49a`. 계측기 `internal/web/partial_apply_repro_test.go`(`recordingSeams`)를 확장해 쓴다 — 교체하지 않는다(REQ-WC16-010).

---

## §D.1 AC Matrix

### 축 (4) — 디스크 진실 재렌더

- **AC-WC16-001** (REQ-WC16-001) — Given `handleSave` 가 4단계(`writeProjectConfig`)에서 실패하는 요청, When 실패 페이지가 렌더되면, Then 그 페이지는 제출된 `development_mode` 값을 현재값으로 표시하지 않는다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFailurePageShowsDiskTruth' -count=1
  ```
  단언: `development_mode` 는 4단계에서 쓰이므로 4단계 실패 시 디스크에 닿은 적이 없다 — 재렌더된 폼의 선택값이 제출값(`tdd`)과 같아서는 안 되며, `readProjectConfig` 가 돌려준 값과 같아야 한다.

  **현재 무엇 때문에 실패하는가**: `renderErrorPage`(handlers.go 선언 :600) → `projectView`(선언 :589) 가 인자로 받은 **제출값** `devMode` 를 `view.CurDevelopmentMode` 에 그대로 넣는다. 디스크 재독이 없다.

- **AC-WC16-002** (REQ-WC16-002) — Given 주입 가능한 여섯 단계 각각에서 실패하는 요청, When 실패 페이지가 렌더되면, Then 그 페이지에 "어느 단계까지 반영됐고 어디서 멈췄는지"를 읽을 수 있는 표면이 존재한다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFailurePageEnumeratesLanded' -count=1
  ```
  단언: 각 실패 지점에 대해, 실패 이전 단계(`calls[:len-1]`)는 "반영됨" 쪽에, 실패 단계와 이후 단계는 "반영 안 됨" 쪽에 나타난다. 표현 형태는 plan.md §F.1 의 clarification 해소 결과를 따른다.

  **현재 무엇 때문에 실패하는가**: 페이지에 그런 표면이 없다. 기존 계측기가 이를 역으로 단언한다 — `partial_apply_repro_test.go:156-160` 은 본문이 **어떤 committed step 이름도 포함하지 않음**을 확인하고 통과한다.

- **AC-WC16-003** (REQ-WC16-003) — Given 1단계(`writePreferences`)에서 실패하는 요청, When 실패 페이지가 렌더되면, Then 제출된 `user_name` 이 현재값으로 표시되지 않고, 페이지가 "아무것도 반영되지 않았다"를 말한다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFirstStepFailureShowsNothingLanded' -count=1
  ```
  단언: 본문에 `value="REPRO-USER"` 부재.

  **현재 무엇 때문에 실패하는가**: t1049 실측 — 1단계 실패 사례에서도 `form echoes submitted user_name: true`. 아무것도 쓰이지 않았는데 폼은 저장된 것처럼 보인다.

- **AC-WC16-004** (REQ-WC16-004) — Given 영속화 실패 **그리고** 디스크 재독도 실패하는 요청, When 실패 페이지가 렌더되면, Then 해당 값은 unverified 로 표시되고 제출값으로 폴백하지 않는다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyDegradedReadNeverFallsBackToSubmitted' -count=1
  ```
  단언: `readPreferences` / `readProjectConfig` 를 오류 반환으로 주입한 상태에서, 재렌더 본문에 제출값 문자열이 현재값 자리로 나타나지 않는다.

  **현재 무엇 때문에 실패하는가**: 실패 경로에 재독 자체가 없으므로 "재독 실패 시 저하"라는 분기도 없다 — 언제나 제출값이다.

- **AC-WC16-005** (REQ-WC16-005, HARD-3) — Given GLM API key 를 실은 제출이 어느 단계에서든 실패하는 경우, When 실패 페이지와 배너가 렌더되면, Then 그 바이트 어디에도 제출된 키 값이 나타나지 않고, 실패 보고는 층 이름과 대상 파일만 지칭한다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFailureNeverLeaksSecretValue' -count=1
  ```
  단언: 식별 가능한 sentinel 키 문자열을 제출한 뒤 응답 본문 전량에 대해 `strings.Contains` == false. 9단계(`glmcred.Save`) 경로 포함(AC-WC16-008 의 탐침과 같은 하네스).

  **현재 무엇 때문에 실패하는가**: 이 단언을 하는 테스트가 없다. 위험 지점은 `handlers.go:534`/`:547` — 배너가 `err.Error()` 를 직접 이어 붙이므로, 하위 층이 값을 실어 보낸 오류를 만들면 그대로 화면에 나간다. **현재 유출을 관측한 것이 아니라, 유출을 막는 가드가 부재함을 관측했다.**

### 축 (5) — 배너 정밀도

- **AC-WC16-006** (REQ-WC16-006, REQ-WC16-007) — Given 영속화의 진부분집합만 완료한 실패, When 배너가 렌더되면, Then 배너는 전체 저장 성공 문구를 쓰지 않는다.

  ```bash
  test "$(grep -c '"settings saved, but' internal/web/handlers.go)" -eq 0 && \
  go test ./internal/web/ -run 'TestPartialApplyBannerAssertsOnlyObserved' -count=1
  ```
  단언: grep 은 과잉 주장 리터럴의 소멸을, 테스트는 각 실패 지점의 배너가 실제 완료 집합과 일치함을 본다.

  **현재 무엇 때문에 실패하는가**: `grep -c` 는 현재 **2**(handlers.go:534, :547). 8·9단계 배너가 `settings saved, but …` 로 전체 저장을 주장하는데, 그 시점에 완료된 것은 1·3·4·5·6단계뿐이다.

- **AC-WC16-007** (REQ-WC16-008) — Given 1단계에서 실패하는 요청, When 배너가 렌더되면, Then 배너는 아무것도 저장되지 않았음을 말한다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFirstStepBannerSaysNothingSaved' -count=1
  ```
  **현재 무엇 때문에 실패하는가**: 현재 문구는 `could not save profile preferences: …`(handlers.go:466) — 실패한 대상만 말하고, **나머지가 어떻게 됐는지는 말하지 않는다.** 사용자는 뒷단계가 돌았는지 알 수 없다.

### 검증 범위 · 하네스 보존 · 회귀

- **AC-WC16-008** (REQ-WC16-009) — Given 7단계(`applyPerfTierEdits`)와 9단계(`glmcred.Save`)를 파일시스템 수준으로 실패시키는 탐침, When 그 실패가 일어나면, Then 두 단계의 실패 보고도 AC-WC16-001..007 과 같은 단언을 통과한다.

  ```bash
  go test ./internal/web/ -run 'TestPartialApplyFilesystemProbe' -v -count=1
  ```
  단언: 대상 경로를 쓰기 불가로 만들어 실패를 유발하고, 관측된 실제 실패를 기록한다. **두 단계를 "재지 못했으므로 통과"로 처리하는 것은 이 AC 의 위반이다.**

  **현재 무엇 때문에 실패하는가**: 그런 탐침이 없다. 두 단계는 패키지 함수라 주입 seam 이 없고, `partial_apply_repro_test.go:55-62` 의 `injectableSteps` 에서 **측정된 이유로** 빠져 있다.

- **AC-WC16-009** (REQ-WC16-010, HARD-4) — Given 본 SPEC 의 변경이 전부 착지한 트리, When 기존 하네스의 양성 대조를 돌리면, Then GREEN 이고 `recordingSeams` 의 seam 목록은 보존돼 있다.

  ```bash
  grep -q 'func recordingSeams(a \*app, calls \*\[\]saveStep, failAt saveStep)' internal/web/partial_apply_repro_test.go && \
  test "$(grep -cE '^\ta\.(writePreferences|recordLastProfile|syncToProject|writeProjectConfig|writeProjectNestedConfig|applySchemaEdits|patchAgentFM) = ' internal/web/partial_apply_repro_test.go)" -ge 7 && \
  go test ./internal/web/ -run 'TestPartialApplyOrderPositiveControl' -count=1
  ```
  단언: 시그니처 보존 + 7개 seam 배선 보존 + 양성 대조 GREEN. t1051 이 같은 하네스에 올라탄다.

  **현재 무엇 때문에 실패하는가**: 이 AC 는 **현재 통과한다** — 보존 단언이기 때문이다. 이것은 반증력 결함이 아니라 의도된 회귀 가드다(§D.2 예외).

- **AC-WC16-010** (REQ-WC16-011) — Given Advisory D1 결정이 유효한 상태, When 그 가드 테스트를 돌리면, Then **테스트 파일 무수정**으로 GREEN 이다.

  ```bash
  go test ./internal/web/ -run 'TestSaveSyncFailureSurfacesReadableError' -count=1 && \
  git diff --exit-code -- internal/web/handlers_test.go | head -0; \
  git diff --stat -- internal/web/handlers_test.go
  ```
  단언: 여섯 영속화 경계 중 **유일하게 선언된 근거를 가진** 경계의 거동이 보존된다. 이 테스트를 고쳐야 통과한다면, 그것은 명시된 결정을 지우는 것이므로 STOP + blocker.

  **현재 무엇 때문에 실패하는가**: 현재 통과한다 — 보존 단언(§D.2 예외).

- **AC-WC16-011** (전체 회귀) — Given 본 SPEC 의 변경이 전부 착지한 트리, When 패키지 전량 테스트를 돌리면, Then EXIT 0 이다.

  ```bash
  go test ./internal/web/ -count=1
  ```
  기준선: t1049 실측 EXIT 0, 25.2s (`.moai/reports/t1049/web_pkg.log`).

---

## §D.2 반증력 규칙 — "지금 통과하는 AC 는 AC 의 결함이다"

AC-WC16-001 .. AC-WC16-008 은 **현재 코드에서 반드시 실패해야 한다.** run-phase 는 각 신규 테스트에 대해 착수 시점 트리에서 먼저 RED 를 보이고 그 출력을 증거로 남긴다. 현재 코드에서 곧바로 통과한 AC 는 거동을 측정하지 못한 것이므로 다시 쓴다.

**예외는 둘뿐**이며 둘 다 명시적 보존 단언이다: AC-WC16-009(하네스 보존, t1051 공유), AC-WC16-010(Advisory D1 보존). 이 둘은 지금 통과하는 것이 정상이고, **끝에도 통과해야 한다.**

## §D.3 추적성 (REQ ↔ AC)

| REQ | AC |
|---|---|
| REQ-WC16-001 | AC-WC16-001 |
| REQ-WC16-002 | AC-WC16-002 |
| REQ-WC16-003 | AC-WC16-003 |
| REQ-WC16-004 | AC-WC16-004 |
| REQ-WC16-005 | AC-WC16-005 |
| REQ-WC16-006 / 007 | AC-WC16-006 |
| REQ-WC16-008 | AC-WC16-007 |
| REQ-WC16-009 | AC-WC16-008 |
| REQ-WC16-010 | AC-WC16-009 |
| REQ-WC16-011 | AC-WC16-010 |
| (전체) | AC-WC16-011 |

## §D.4 Definition of Done

- AC-WC16-001..011 전량 GREEN, 각 명령의 축자 출력이 `progress.md` §E.2 에 기록됨.
- AC-WC16-001..008 각각에 대해 **착수 시점 RED 출력**이 함께 기록됨(§D.2).
- plan.md §F.1 의 세 `[NEEDS CLARIFICATION]` 이 전부 해소되고, 해소 답이 문서에 기록됨.
- spec.md §6 의 Gaps 중 (1)은 AC-WC16-008 로 닫히고, (2)(3)(4)(5)는 **닫히지 않은 채 명시적으로 이월**된다 — 닫힌 것처럼 보고하지 않는다.
- 영속화 단계의 순서·개수·존재 무변경(HARD-1): `grep -c "a.renderErrorPage(" internal/web/handlers.go` 와 단계 순서가 계측기 양성 대조에서 보존됨.
