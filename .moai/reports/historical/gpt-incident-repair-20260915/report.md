# GPT 세션 오류 수정 및 로컬 빌드 검증

작성일: 2026-09-15 KST

## Claim — 확인한 결과

**판정: 확인된 메인 400·clear 502 원인 수정 및 로컬 빌드 완료. 운영 전체 정상화는 미확정.** 설치 파일 교체, 사용자 프로세스 중단, 커밋·푸시는 하지 않았다.

| 항목 | 원인 및 변경 | 검증 범위 |
|---|---|---|
| 취소 후 새 이미지·텍스트 입력의 메인 400 | 도구 결과 전달 후 취소된 경우, 마지막 assistant 뒤의 이미 소비한 tool result를 다시 전달했다. durable accepted prefix가 정확히 일치할 때만 소비한 부분을 제외한다. | 실제 Engine + fake RPC에서 결과 1회 전달, 취소 확인, 새 턴 성공. 변조된 결과는 기존 거부 유지 |
| `/clear` 후 502 | 새 Claude session UUID를 기존 launcher UUID 검사에서 거부하고, 이 오류를 일반 transport 오류로 표시했다. private native clear 완료 증거를 확인하여 별도 owner를 허용하고, 미승인 변경은 typed 400으로 구분한다. | 실제 gateway·adapter·bridge·shared transport + fake peer에서 이전/새 UUID 모두 HTTP 200, 서로 다른 thread 2개. 미승인·충돌 UUID는 HTTP 400이며 추가 모델 실행 0 |
| review hook 실행 오류 | 실행 권한이 있는 디렉터리를 바이너리로 오인했다. 후보를 `-f && -x`로 제한한다. | 재현 테스트 RED 후 GREEN, hook 및 embedded template 반영 |
| 자동 요약 400 | 실제 실패 요청을 재현하지 못했다. 현재 classifier를 추측으로 변경하지 않았다. | 별도 실제 구독 GPT에서 pending child 상태의 요약 2회와 child 재개 성공. 기존 실패 해결의 증거는 아님 |
| 응답 지연 | native GPT가 Claude Agent를 기다리며 15/30/45초 setTimeout을 반복한 기록을 확인했다. Agent 도구가 있는 경우 불필요한 polling 대신 진행 상황을 알리고 응답을 종료하도록 지침을 보강했다. | 지침 계약 테스트 통과. 실제 지연 개선량 미측정 |

자동 요약 등 scope 오류가 재발하면 비공개 rejection 로그에 고정된 allowlist 사유 코드를 기록한다. 공개 응답은 기존 오류명을 유지하며 프롬프트·자격 증명·원본 예외를 추가로 저장하지 않는다.

## Evidence — 실행 명령과 관측 출력

환경 변수는 각 검증과 같은 호출에서 해제했다: `CLAUDECODE CLAUDE_CONFIG_DIR ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN ANTHROPIC_API_KEY MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE`. 실제 구독 검증에서만 `MOAI_GPT_LIVE=1`, Claude 통합 검증에서만 `MOAI_CLAUDE_INTEGRATION=1`을 지정했다.

### 최종 변경 영향 범위 race 검사

```sh
go test -race ./internal/cli ./internal/codexbridge ./internal/codexapp ./internal/gateway/... ./internal/hook ./internal/template -run 'TestManagedGPT|TestSharedGPT|TestAppServer|TestNative|TestCodexReviewGate|TestReviewGate|TestManifestHashFormat|Test.*Cancel|Test.*Scope|Test.*Queue|Test.*Instructions' -count=1 -timeout=180s
```

```text
ok  github.com/modu-ai/moai-adk/internal/cli 22.900s
ok  github.com/modu-ai/moai-adk/internal/codexbridge 13.255s
ok  github.com/modu-ai/moai-adk/internal/codexapp 3.646s
ok  github.com/modu-ai/moai-adk/internal/gateway 7.251s
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 3.205s
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation 2.293s
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque 1.413s [no tests to run]
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 3.921s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 3.545s
ok  github.com/modu-ai/moai-adk/internal/hook 7.456s
ok  github.com/modu-ai/moai-adk/internal/template 4.296s
```

원본 출력: [race 검사](evidence/moai-gpt-incident-final-tests.log). 위 검사는 패턴으로 선택한 검사이며 전체 저장소 검사나 모든 패키지의 전체 테스트 통과를 뜻하지 않는다.

### `/clear` production 경로

```sh
go test -race ./internal/cli -run '^TestProductionGatewayClearStartsDistinctAttestedThread$' -count=1 -v
```

```text
=== RUN   TestProductionGatewayClearStartsDistinctAttestedThread
    gpt_appserver_clear_integration_test.go:120: old+clear HTTP200; independent threads=2 turns=2; unattested+conflicting HTTP400 with zero additional model execution
--- PASS: TestProductionGatewayClearStartsDistinctAttestedThread (0.28s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 2.533s
```

### 실제 구독 GPT 및 Claude Code 검사

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLiveSummaryPreservesPendingChild$' -count=1 -v -timeout=150s
MOAI_CLAUDE_INTEGRATION=1 go test ./internal/cli -run '^TestClaudeCodeProductionGatewayAgentRoundTrip$' -count=1 -v -timeout=120s
```

원본 출력: [구독 GPT 요약·재개](evidence/moai-gpt-incident-summary-live.log), [Claude Code Agent 왕복](evidence/moai-gpt-incident-claude-integration.log).

구독 GPT 검사는 별도 테스트 대화를 사용했다. Claude Code 왕복 검사의 upstream은 fake App Server이며 실제 구독 종단 간 검사와 구별해야 한다.

### 정적 검사

`go vet`의 cli·codexbridge·codexapp·gateway·hook 영향 범위 검사: exit 0.

`golangci-lint run ./internal/codexbridge/... ./internal/codexapp/... ./internal/gateway/... ./internal/cli/...`: exit 1.

```text
60 issues:
* errcheck: 53
* ineffassign: 1
* staticcheck: 5
* unused: 1
```

[최종 lint 원본](evidence/moai-gpt-incident-lint-final.txt). 초기 관측 총계도 60이었으나 항목별 무회귀 증명을 완료하지 않았으므로 lint PASS로 판정하지 않는다.

`ast-grep scan --config .moai/astgrep-rules/sgconfig.yml internal/codexbridge internal/codexapp internal/gateway internal/cli --json=compact`: 최종 exit 1. 초기 2,495건, 최종 2,499건. rule/file/text 비교 신규 4건은 error 반환, allowlist 검사 뒤 map 조회, 테스트 채널 전송, 오류를 별도로 검사하는 다중 반환 대입 패턴이다. 패턴 경고 자체를 실제 결함으로 단정하지 않으며 전체 정적 분석 통과도 주장하지 않는다.

`git diff --check`: exit 0, 출력 없음.

### 로컬 빌드

```text
[v3.2.0-rc.11] [v3.2.0-rc.11-b45c81349-scope-clear-dirty] [built 2026-09-14T16:49:41Z]
3c1290e8e79b230c3f1826279088d85844f60d6a6d440ddfada844e5de904151  /tmp/moai-rc11-incident.wytxwx/moai
f40ba4cb307e62fb1e4b48031a4e8536e78cd79e0081986a43555e36116184d1  /Users/goos/go/bin/moai
```

`go build`로 `./cmd/moai`를 빌드하고 Version/Commit/Date/BuildID를 ldflags로 지정했다. 빌드 및 version 실행 exit 0. 설치 파일 해시는 기존 관측값과 같으며 교체하지 않았다. 임시 경로의 파일은 영구 배포 자산이 아니다.

## Baseline-attribution — 측정 기준

- 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`
- 브랜치: `WT-gpt-session-repair`, HEAD `b45c81349`
- 기존 미커밋 수정을 보존한 dirty tree를 빌드했다. HEAD만으로 해당 바이너리를 재현할 수 없다.
- 조사 대상 family: `f0579c71-d3e9-4f9b-a83c-8a5f6ddaef77`, `b702a146-4196-4e47-b15c-34ac6169c6a6`.
- clear 이후 native UUID: `cee85e72-d4ee-4030-9521-71748453c4fa`. 원래 debug 파일만 보면 clear 이후 오류를 놓치므로 새 debug 파일과 native rollout을 함께 분석했다.
- 메인 400: 도구 결과 수신 16:28:25.263 UTC → 취소 16:28:33.468 → 새 입력 16:28:40.846 → 오류 16:28:41.008.
- 분석·구현·검증은 incident-response와 moai-fix 절차를 적용했다. 안전 검사 유지, 재현 후 수정, 영향 범위 검증으로 범위를 제한했다.

## Gaps — 관측하지 못한 것

1. 반복된 운영 자동 요약 400의 정확한 내부 거부 분기는 미확정이다. 새 진단을 포함한 바이너리를 실제 사용해야 추가 근거를 확보할 수 있다.
2. 수정 바이너리로 사용자의 실제 GPT 세션에서 취소→재입력과 `/clear`를 실행하지 않았다.
3. `/clear` 이후 세션 검색·재개 인덱스 등록까지 완료했다고 주장하지 않는다.
4. Claude Code 모든 기능, 모든 모델·effort 조합, Factory 모든 lane 조합은 전수 검증하지 않았다.
5. 순수 디코딩 tokens/s, 수정 전후 동일 조건 성능 비교, 수정 파일 85% coverage 및 CI는 미측정이다.
6. 설치 바이너리는 기존 버전이다. 현재 사용자 세션에 새 수정이 적용되었다고 볼 수 없다.

## Residual-risk — 남은 위험과 다음 검증

- accepted-prefix 입증 실패나 64개 후보 초과는 기존 안전 거부를 유지한다. 임의 이력 변경을 복구하는 기능은 아니다.
- native clear 증거 확인은 제한된 크기·기록 수를 읽으며 symlink와 파일 동일성을 검사한다. 지원하지 않는 기록 형태는 fail-closed다.
- Agent 대기 지침은 모델 행동을 유도할 뿐 지연 상한을 보장하지 않는다. native 요청 완료의 관측 표본은 약 25–54 output tokens/s였지만 입력 처리·추론 시간이 포함된 요청 단위 처리율이며 순수 생성 속도가 아니다. 긴 응답 공백의 개선 여부는 별도 실사용 측정이 필요하다.
- 다음 단계는 승인 후 로컬 설치 파일 교체, 새 세션에서 취소·이미지·clear·자동 요약 재검증, 비공개 reason으로 잔여 400의 분기 확정이다. 사용자 세션을 임의 종료하지 않는다.
