# t1293 — Stop 체인 간헐 실패 조사

## 주장

CI의 Race 실패는 `1ns` 예산만으로 컷오프를 강제했다고 가정한 테스트의 경쟁 상태다. Ubuntu 실패는 GPT 프로필 테스트의 임시 Git 저장소가 정리되는 동안 `.git/objects`에 쓰기가 발생한 사례다. 수정본의 최종 CI 판정은 아직 없다.

## 근거

- CI 실행 `36365223880`의 Race Test 로그: `TestStopChainGateCutOffNeverAllows`에서 멤버 3이 `deny/"unmeasured"`를 반환했지만 컷오프 오류 문자열은 비어 있었다. 같은 실행의 Ubuntu 테스트는 t1292의 옛 경로 기대치 2건으로 실패했으며 별도 수정했다.
- CI 실행 `36361758033`의 Ubuntu 로그: `TestStopChainGPTProfileNoClaudeDependency` 정리 중 `TempDir RemoveAll cleanup: unlinkat .../.git/objects: directory not empty`가 발생했다. 원인이 Git 자동 유지보수인지 남은 체인 작업인지는 로그만으로 확정할 수 없다.
- `go test ./internal/cli -run '^(TestStopChainGateCutOffNeverAllows|TestStopChainGPTProfileNoClaudeDependency)$' -count=10 -timeout 180s` 출력: `ok github.com/modu-ai/moai-adk/internal/cli 15.701s`.
- 같은 두 테스트의 `go test -race ... -count=5 -timeout 240s` 출력: `ok github.com/modu-ai/moai-adk/internal/cli 10.600s`.
- `timeout 25s yes`로 제한한 CPU 부하 아래 `go test -race ... -count=5 -timeout 120s` 출력: `ok github.com/modu-ai/moai-adk/internal/cli 13.853s`.
- `go test ./internal/cli -run '^TestStopChain' -count=1 -timeout 180s` 출력: `ok github.com/modu-ai/moai-adk/internal/cli 46.441s`.
- `go vet ./internal/cli` 종료 코드 0, 출력 없음.

## 기준 트리

위 로컬 검사는 `develop`의 `2dbf4321b`에서 분기한 `WT-stop-chain-budget` 작업 트리에서 수행했다. 이 파일 작성 시점의 변경은 아직 `develop`에 통합되거나 원격에 게시되지 않았다.

## 미확인 항목

- 임시 저장소 정리 실패의 유일한 원인은 특정하지 못했다. 테스트 종료 전 체인 작업 합류와 Git 자동 유지보수 비활성화가 재발 경로를 각각 차단하는지 새 CI에서 확인해야 한다.
- 수정 커밋의 Ubuntu 및 Race CI 결과는 아직 없다.

## 남은 위험과 완료 조건

새 CI에서 두 테스트가 통과하고, 수정이 다른 Stop 체인 결정이나 임시 저장소 정리를 깨지 않았음을 확인한 뒤 카드를 완료한다. 다른 카드의 실패가 남으면 해당 실행을 전체 통과로 보고하지 않는다.
