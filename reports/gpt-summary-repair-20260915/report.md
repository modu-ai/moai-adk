# 자동 에이전트 요약 분리 수정 및 rc.11 로컬 배포

## Claim

자동 요약 요청에 도구 목록이 있으면 일반 작업 요청으로 잘못 분류하던 조건을 수정했다. 기본 요약 및 `Previous:` 변형을 독립된 ephemeral 대화로 처리하고, 요약의 실행 도구·도구 참조·도구 결과를 제거한다. 검증된 공개 이력은 읽기 전용 데이터로 보존한다. 기존 작업의 범위 검사와 pending 도구는 변경하지 않는다.

## Evidence

장애 family `b21f7639-1751-4269-a9f3-626f5e731b4a`에서 `agent_summary` 요청의 400 `appserver_scope_mismatch`가 반복됐다. 설치된 Claude Code 2.1.270을 별도 로컬 모의 서버에 연결한 관측에서는 실제 요약 요청에 `tools: ["Agent"]`가 포함됐다. 종전 분류 함수는 `len(root.Tools) != 0`이면 false를 반환했다. 원래 실패 HTTP 요청 원문 자체를 재생한 것은 아니다.

### 수정 전 재현

`TestManagedGPTSummaryWithToolsAndPreviousIsReadOnlySnapshot`의 active/waiting × string/block × 기본/Previous 8개 조합에서 다음 오류를 관측했다.

```text
invalid App Server conversation or tool result
FAIL github.com/modu-ai/moai-adk/internal/cli 0.947s
```

### 수정 후 검사

```sh
go test ./internal/cli -run '^TestManagedGPT' -count=1 -coverprofile=/tmp/moai-summary-fix-coverage.out
```

```text
ok github.com/modu-ai/moai-adk/internal/cli 0.928s coverage: 6.2% of statements
```

분류 함수 coverage는 100.0%, 기존 큰 준비 함수는 78.6%였다. 패키지 전체 coverage 달성을 주장하지 않는다. 구현 담당자가 동일 작업 트리에서 RED/GREEN 및 `go vet ./internal/cli` exit 0을 확인했다.

주 에이전트의 동결 코드 race 검사:

```sh
unset MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE && go test -race ./internal/cli -run '^TestManagedGPT' -count=1
```

```text
ok github.com/modu-ai/moai-adk/internal/cli 2.153s
```

실제 공유 구독 App Server 검사:

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLiveSummaryPreservesPendingChild$' -count=1 -v -timeout=150s
```

```text
summary completed while child pending: elapsed=2.738s tools=0 result="Reporting echo result"
summary completed while child pending: elapsed=6.88s tools=0 result="Reporting echo function result"
child resumed successfully after two independent summaries: result="SUMMARY_CHILD_OK"
--- PASS: TestSharedGPTLiveSummaryPreservesPendingChild (14.35s)
ok github.com/modu-ai/moai-adk/internal/cli 15.082s
```

실제 sol의 echo 도구 결과를 보류한 상태에서 두 요약을 수행했다. 요약마다 별도 owner, 실행 도구·참조·결과 0개, 기존 child barrier 불변을 검사했다. 이후 실제 도구 결과를 반환해 child 완료를 확인했다. 로그: `/tmp/moai-summary-pending-live.log`.

실제 Claude Code와 로컬 App Server fixture의 연결 회귀 검사도 수행했다.

```text
actual Claude Code -> production gateway -> App Server fixture: main=4 agent=1 threads=2 toolResults=1 children=1 result=MAIN_GATEWAY_OK
--- PASS: TestClaudeCodeProductionGatewayAgentRoundTrip (0.34s)
ok github.com/modu-ai/moai-adk/internal/cli 1.232s
```

이 검사는 실제 구독 검사가 아닌 별도 로컬 통합 검사다.

## Baseline-attribution

- 작업 트리: `.claude/worktrees/gpt-session-repair`
- 브랜치: `WT-gpt-session-repair`, HEAD `b45c81349`, 미커밋 수정 포함.
- 제품 변경: `internal/cli/gpt_appserver_runtime.go`.
- 검사 변경: `internal/cli/gpt_appserver_identity_test.go`, 새 `gpt_appserver_summary_live_test.go`.
- 설치 경로: `/Users/goos/go/bin/moai`.
- 버전: `v3.2.0-rc.11`.
- Build ID: `v3.2.0-rc.11-b45c81349-summary-isolation-dirty`.
- 빌드 시각: `2026-09-14T16:05:26Z`.
- 설치 SHA-256: `f40ba4cb307e62fb1e4b48031a4e8536e78cd79e0081986a43555e36116184d1`.
- 설치본과 검증 후보의 `cmp`: exit 0. 설치본 version 출력에서 위 Build ID 확인.
- 이전 파일 백업: `/Users/goos/.moai/releases/rc11-summary-backup.xkYQ8Z/moai-before-summary-repair`.

`moai-fix`의 재현 우선 절차와 구현 분담을 적용했다. 로컬 빌드·복사 배포는 완료했으며 commit·develop 병합·push는 하지 않았다.

## Gaps

수정된 설치본의 실제 대화형 Claude 자동 타이머가 만든 요약 요청까지는 이번 배포 후 확인하지 않았다. 실제 App Server 검사는 같은 요청 구조를 준비 함수에 전달하는 방식이다. 전체 저장소 CI·장시간 Factory 부하·모든 기능 동등성 검증은 미실행이다.

## Residual-risk

실행 중인 gateway 프로세스에는 실행 파일 교체가 자동 적용되지 않는다. 기존 작업이 안전하게 끝난 뒤 종료하고 새 `moai gpt`로 확인해야 한다. 작업 중인 사용자 세션은 종료하거나 수정하지 않았다. 향후 Claude 요약 문구·구조 변경에는 별도 검증이 필요하다. 요약 실패와 무관한 모델 생성 지연의 제거를 보장하지 않는다.
