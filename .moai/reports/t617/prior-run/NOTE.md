# t617 이전 실행 기록 — 출처와 한계

첫 워커(agentId a5130b13e9e45a187)가 2026-09-10 07:12:12Z 에 429 세션 한도로 종료됐다. 판정서를 쓰기 전이었고 커밋도 없었다. 이 디렉터리의 15개 파일은 그 워커가 세션 스크래치에 남긴 출력을 lane-6 이 그대로 복사한 것이다. 내용은 고치지 않았다.

**근거로 쓸 수 없는 부분.** 워커 트랜스크립트는 마지막 27레코드만 남아 있다. 그래서 M1 부터 M4 까지 각 뮤턴트가 `session_start.go` 를 어떻게 바꿨는지, SKIP 을 넣은 이유를 git 이력에서 무엇으로 확인했는지는 복구되지 않는다. 뮤턴트 출력 파일에도 변형 내용이 적혀 있지 않다. 따라서 이 파일들은 다시 잴 곳을 가리키는 참고 자료이지 판정 근거가 아니다. 뮤턴트와 SKIP 경위 조사는 변형 diff 를 함께 기록하며 다시 한다.

**복사 시점에 확인한 것.**
- 워커 트리의 `internal/hook/session_start.go` sha256 `939547cc…b247c8` 는 develop 의 같은 파일과 같다. 뮤턴트는 원복된 상태다.
- 트리에 남은 변경은 `internal/hook/session_start_migration_test.go` 하나다. 임시 탐침 파일 `zz_t617_probe_test.go` 는 없다.
- 워커 브랜치의 기준 `d3b7d438d` 이후 develop 은 `internal/hook` 을 바꾸지 않았다(`git diff --stat d3b7d438d develop -- internal/hook` 결과가 비어 있음).
- `12-final-hook-suite.txt` 의 `internal/hook` FAIL 은 `TestSessionStart_DeferredScanDoesNotBlockReturn`(`session_start_parallel_test.go:97`, 585ms 블로킹) 한 건이다. 이 카드의 파일과는 다른 파일이다. 원래 불안정한 테스트인지는 아직 재지 않았다.
