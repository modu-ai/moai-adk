# t1171 후속 — rosterguard 적색 수리

## 주장
develop d6992e3a0 CI(run 36225257611)의 rosterguard numeral 테스트 3건 실패는 t1171 픽스처(1a8dfa26e)가 넣은 파일 7개가 numeral 예외로 등록되지 않아서 생긴 것이다. 이 7개를 경로별 `NumeralExempt` 로 등록해 수리했다.

## 증거
- 재현 (기준 0119a8f2c, 수리 전): `go test -count=1 ./internal/harness/rosterguard/` → exit=1
  - `--- FAIL: TestNumeralResidualArithmeticCloses` / `TestNumeralAxisFindsNoUndeclaredCountClaim` / `TestNumeralBreadthSetEqualsTheDeclaredUnion`
  - `numeral_rederivation_test.go:87: 7 breadth-set path(s) are neither registered nor exempted`
  - 전체 로그: `fix-repro.log`
- 수리 후 같은 명령: exit=0, `ok github.com/modu-ai/moai-adk/internal/harness/rosterguard 26.628s` (`fix-after.log`). `go vet` 통과, `gofmt -l` 출력 없음.
- 픽스처가 인용하는 숫자 주장을 grep 으로 확인: roles/manager-spec.toml:6 "reduced 17 agents to the th…", real rollout "13 retained agents".

## 판단
t1163·t1164 는 sweep 제외(`sweepSkipPrefixes`) 방식을 썼다. 그 목록은 추적되지 않는 런타임 트리용이다. 이 픽스처는 추적되는 캡처 스냅숏이므로, 이 패키지 설계("경로별 선언, 패턴 추론 금지")를 따라 HISTORICAL CITATION 분류의 경로별 예외 7행으로 등록했다. 멤버십 sweep 은 실패하지 않았으므로 건드리지 않았다.

## 미검증
- 로컬에서는 이 패키지만 돌렸다. 판정은 CI 몫이다.
- 병합 트리(흡수 기준 b1a62fb2b) 재측정은 병합 창에서 한다.

## 잔여 위험
- 픽스처가 나중에 더 늘어나면 같은 실패가 다시 난다. 경로별 선언은 의도된 비용이다.
