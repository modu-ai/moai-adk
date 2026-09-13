# GPT 모델 전환 400 원인 및 수정 증거

## Claim

같은 계정의 GPT-5.6 응답을 가진 대화에서 GPT-6 Astra로 전환하면 기존 receipt의 모델 family hash와 대상 모델 family hash가 달라 로컬 gateway가 송신 전 거절했다. 안전한 고정 오류 분류로 재현하였다. 응답이 `Bad Request`로만 바뀌어 복구 안내도 사라졌다.

GPT 구독의 두 family 사이 암호문 호환성은 실제 Sol/Terra/Luna → Astra → 원래 모델의 여섯 방향에서 확인하였다. 생산 코드는 기존 receipt를 옮기거나 다시 쓰지 않고 각 assistant 경계를 원래 family의 hash로 검증한다. 같은 계정·세션·대화 family·공개 prefix·암호문 digest가 필요하며, 새 응답만 대상 family domain으로 게시한다. 기존 strict constructor는 유지하고 GPT subscription factory에서만 명시 호환 constructor를 선택한다. 미확인 모델은 기존 allowlist 검증에서 거절한다.

## Evidence

재현 명령은 `e2e/gateway_probe.py --binary /tmp/moai-t649-user400-diagnostic --cwd /tmp/moai-t649-e2e-project --journey answer --model gpt-5.6-sol --allow-live` 후 해당 session UUID에 `--journey resume --model gpt-6-astra`를 적용하였다. 실행별 전체 비밀값 없는 결과는 같은 폴더 JSON에 보존했다.

```text
Sol: success=true, elapsed_seconds=4.21
Astra retained history: success=false, elapsed_seconds=0.86
adapter conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation
```

격리한 합성 대화만 사용한 호환성 진단 overlay는 source Sol domain 검증을 유지하면서 Astra에 암호문을 전달했다. 이 진단 overlay는 제품에 설치하지 않았다. Sol/Terra/Luna 각각의 Astra 왕복은 `t649-user400-compatibility-astra2.json`, `t649-user400-compatibility-back-sol.json`, `t649-user400-{terra,luna}-astra.json`, `t649-user400-astra-back-{terra,luna}.json`에서 모두 success=true이다. 첫 `t649-user400-compatibility-astra.json`은 앞선 의도적 실패 후 재개 거절로 호환성을 측정하지 못한 기록이며 그대로 보존하였다. 성공한 여섯 방향의 근거를 바탕으로 family 두 개의 명시 호환 검증을 구현하였다.

RED 실제 출력은 `red-evidence.txt`에 기록하였다. 특히 실제 factory 경유 두 요청의 두 번째가 400으로 실패했고, cross-domain empty receipt가 opaque-required receipt를 숨기는 음성 시험도 실패한 뒤 보수적 거절 검사를 추가하였다.

```text
$ go test ./internal/gateway ./internal/gateway/translate ./internal/gateway/receipt -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway 6.606s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.334s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 1.325s
$ go test ./internal/gateway ./internal/gateway/translate ./internal/gateway/receipt -race -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway 9.410s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.758s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 3.920s
$ go vet ./internal/gateway ./internal/gateway/translate ./internal/gateway/receipt
(exit 0, output empty)
$ go test ./internal/cli -run TestGatewayFactory -count=1
ok  github.com/modu-ai/moai-adk/internal/cli 1.064s
```

별도 overlay 없는 제품 코드 빌드 `/tmp/moai-t649-model-switch-fixed`로 기존 Sol-domain receipt를 가진 합성 대화를 다시 Astra로 재개하고 Sol로 돌아왔다.

```text
fixed-mixed-astra.json: success=true, assistant_models=[gpt-6-astra], elapsed_seconds=4.18
fixed-mixed-back-sol.json: success=true, assistant_models=[gpt-5.6-sol], elapsed_seconds=3.07
binary_sha256=6fbcdf8848bb30521be8c5d7e9f834a5d8dfa3fd37860b08857dcf6b45ff359a
```

마지막 추가 암호문 변조 음성 시험:
```text
$ go test ./internal/gateway/translate -run TestSubscriptionReceiptModelSwitchPreservesSourceBindings -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.091s
```

## Baseline-attribution

2026-09-12 현재 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `git rev-parse --short HEAD` → `81c1d58f9`, `git branch --show-current` → `WT-unified-gateway`. 기존 dirty 작업을 보존하였다. rc8 원본을 바꾸지 않고 별도 실행 파일로 측정했다. source session `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`.

## Gaps

사용자 실제 family에서 최초 오류를 생성한 원문 request는 저장하지 않았다. 해당 로그와 같은 모델 이력 전환으로 합성 fixture에서 원인을 재현한 근거이며, 사용자 실대화 재개는 별도 검증 대상이다. 기존 실패 후 resume의 `gateway resume` 오류는 다른 작업자가 담당한다. 인터랙티브 `/model` UI 왕복, 도구가 있는 cross-model 왕복, 최종 설치 바이너리 검증은 부모 작업자가 수행한다. 전체 통합 브랜치 CI는 PENDING이고 본 작업에서 commit/push/설치하지 않았다. 1M context 설정은 이 수정에 포함하지 않는다.

## Residual-risk

호환성 관측은 현재 계정의 합성 대화와 현재 endpoint에 한정된다. 서버 정책이 바뀌면 다시 거절될 수 있다. printf 아닌 실제 PTY print 경로의 종료된 process-group cleanup에서 `signal_permission_denied`가 기록되었으며 자연 종료 exit_code=0과 구별하여 보존했다. 요청·암호문·토큰·개인 대화 원문을 증거에 저장하지 않았다.
