# t1186 — Codex 백그라운드 handshake 제한시간

## Claim

Codex 백그라운드 작업의 자체 제한시간을 세션 생성 시점부터 적용한다. initialize handshake가 응답하지 않으면 작업을 만들지 않고 제한시간 초과로 보고하며 자식 프로세스를 종료한다. 제한시간이 턴 도중 만료될 때 EOF 경합이 발생해도 제한시간 초과로 기록한다.

## Evidence

- 실제 shell 자식이 initialize에 응답하지 않는 TestCodexTaskBackgroundHandshakeHonorsTaskBound를 추가했다. 수정 전 go test ./internal/cli -run '^TestCodexTaskBackgroundHandshakeHonorsTaskBound$' -count=1 -v -timeout=15s → FAIL, 0.70s, "background handshake outlived the 100ms task bound".
- 수정 후 같은 명령 → PASS, 0.10s. 결과는 status=failed, error에 "timed out after 100ms", job_id 없음이다. 테스트는 PID 파일로 실제 자식 실행을 확인하고 반환 뒤 해당 PID의 종료를 확인한다.
- go test ./internal/cli -run 'TestCodexTask(BackgroundHandshakeHonorsTaskBound|MCPEOFAbortsStalledHandshake|BackgroundProcessBound|_BackgroundDetachedSessionHonorsTurnBound|_BackgroundHandshakeStillBoundToRequest|_BackgroundDetachedSessionIsCancellable)$' -count=3 -timeout=45s → ok github.com/modu-ai/moai-adk/internal/cli 4.443s, exit 0.
- go test ./internal/cli -run 'TestCodexTask_(ForegroundSessionStillBoundToRequest|BackgroundSessionOutlivesRequestContext)$' -count=1 -timeout=20s → ok github.com/modu-ai/moai-adk/internal/cli 0.840s, exit 0.
- gofmt -l internal/cli/codex_task.go internal/cli/codex_task_process_context_test.go 및 git diff --check → 출력 없음, exit 0.

## Baseline-attribution

기준은 WT-codex-task-background-context 작업트리의 ee78515a9이다. RED와 GREEN, 형제 테스트는 이 작업트리에서 직접 실행했다.

## Gaps

이 검증은 실제 shell 자식 프로세스와 MCP 핸들러를 사용하지만, 운영자의 Codex 계정으로 원격 모델 턴을 실행한 것은 아니다. 원격 PR·CI 검증은 수행하지 않았다.

## Residual-risk

OS 스케줄링 지연으로 100ms 설정과 관측 시간이 정확히 일치하지 않을 수 있다. 테스트는 timeout 문자열, 자식 종료, 무응답 요청의 유한한 종료를 확인한다.
