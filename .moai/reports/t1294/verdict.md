# t1294 — Codex 팩토리 세션·슬롯 오배정 회귀 검증

## 주장

운영자가 제시한 rc.16 사례는 F2 팩토리 진입을 구현할 때 확인해야 할 회귀 조건이다. 이 카드의 런타임 수정·완료 판정은 아직 없다. t1240이 Codex `-f agent`를 자가 배차 역할로 도입하는 중이다.

## 근거

- 운영자 제공 사례: 설치본 `a8a9b9376`에서 `moai codex -f agent`가 런 `tm1saz`의 예상 자리 `worker-1` 대신 `worker-10`~`worker-17`을 등록했고, 세션 UUID 없이 `launch_pending`에 남았다. 이 항목은 과거 관측 전달이며 이번 턴의 실시간 재측정은 아니다.
- 이번 조사에서 `git show a8a9b9376:internal/kanban/bootstrap.go`를 읽었다. 당시 `NextFactoryWorkerNumber`는 살아 있는 최고 번호에 1을 더했다. 같은 커밋의 `internal/cli/codex_factory.go`에는 `-f agent`가 워커 합류의 옛 별칭으로 명시돼 있다.
- `d5df9457c` 기준으로 실행한 `go test ./internal/cli -run '^TestCodexFactoryEntryIsRefused$' -count=1 -v -timeout 90s`의 출력:

  ```text
  === RUN   TestCodexFactoryEntryIsRefused
  --- PASS: TestCodexFactoryEntryIsRefused (0.00s)
  PASS
  ok  github.com/modu-ai/moai-adk/internal/cli  0.822s
  ```

- 공식 SQLite 카드 t1240은 `moai cc|glm|codex -f agent`를 새 자가 배차 역할로 정의한다. 따라서 옛 워커 별칭의 번호 배정을 그대로 되살리면 안 된다. t1292의 `.moai/worktrees` 경로 변경은 `develop`의 `a7190891d`로 병합·게시됐고, CI 테스트 경로 수정은 `2dbf4321b`로 게시됐다.

## 기준 트리

위 테스트는 `WT-codex-factory-binding`이 분기한 `d5df9457c`에서 측정했다. 설치본 사례의 런 상태와 새 F2 구현 동작은 이 기준에서 검증하지 않았다. t1292의 병합·게시 사실은 이번 세션에서 Git 명령과 푸시 출력으로 확인했다.

## 미확인 항목

- t1240의 F2 실행 코드가 아직 통합되지 않아 Codex 자가 배차·워커 슬롯 제한·세션 UUID 연결을 시험할 수 없다.
- 살아 있는 `tm1saz` 런과 Codex 창은 다시 조회하거나 변경하지 않았다.
- t1292를 포함한 `develop` CI의 최종 통과 판정은 아직 없다.

## 남은 위험과 완료 조건

t1240 통합 후 `-f agent`가 워커 번호를 점유하지 않음을 확인한다. 명시적 워커 합류는 지정 런의 허용된 빈 자리만 점유하고, 범위 초과·자리 부족 시 상태를 바꾸지 않아야 한다. 첫 프롬프트 전 `launch_pending` 표시와 이후 런·역할·세션 UUID 연결을 실제 argv 및 상태 전환으로 검증한 뒤에만 t1294를 완료한다.
