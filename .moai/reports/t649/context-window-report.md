# GPT 구독 경로의 최대 입력 길이 실증 보고서

2026-09-12 · 카드 t649 · 상태 보고 · 일반 독자용(MoAI-Easy)

## Claim — 실증 결론

현재 계정의 GPT 구독 endpoint에서 Sol과 Astra 모두 약 92만 1천 입력 토큰을 처리했다. 정상 완료 시 서버 usage를 확인했고, 입력의 처음·중간·끝에 넣은 세 표식을 모두 답변에서 회수했다. 각 모델의 최대 성공 요청을 한 차례 더 보내 재현을 확인했다.

| 모델 | 최대 성공 입력: 서버 실측 | 최소 거절 입력: 보정 추정 | 경계 구간 폭 | 최대 성공 반복 |
|---|---:|---:|---:|---:|
| gpt-5.6-sol | 921,825 | 약 921,883 | 58토큰 | 2회 성공 |
| gpt-6-astra | 921,821 | 약 921,883 | 62토큰 | 2회 성공 |

이 표는 같은 합성 입력 형식에서 관찰한 성공과 실패의 경계다. 모든 계정과 요청 형식에 적용되는 정확한 서버 최대값을 뜻하지 않는다. 실패 응답에는 usage가 없어 그 입력 길이는 실측값이 아니다.

**1,050,000토큰 전체를 입력으로 사용하는 데 성공하지 못했다.** 두 모델 모두 약 1,000,071 입력 토큰 요청을 길이 초과로 거절했다. 공개 API의 1,050,000 context 사양은 구독 endpoint에서 그만큼 입력할 수 있다는 증거가 아니다. 이번 실험의 출력은 15토큰이며, 입력과 출력의 전체 허용량이나 서버의 내부 예약량은 확인하지 않았다.

## Evidence — 명령과 관찰 출력

실제 구독 인증을 사용하는 Go overlay 테스트로 다음 명령을 실행했다. `T649_WORDS`는 토큰 수가 아니라 ` x`의 반복 횟수다. 다른 요청의 정확한 명령은 원시 summary.json의 journeys에 있다.

```sh
T649_MODEL=gpt-5.6-sol T649_WORDS=921754 T649_RESULT=<증거 디렉터리>/sol-921754-repeat.json go test -overlay .moai/reports/t649/context-window-evidence/e2e/overlay.json ./internal/cli -run '^TestT649ContextLive$' -count=1 -v
T649_MODEL=gpt-6-astra T649_WORDS=921750 T649_RESULT=<증거 디렉터리>/astra-921750-repeat.json go test -overlay .moai/reports/t649/context-window-evidence/e2e/overlay.json ./internal/cli -run '^TestT649ContextLive$' -count=1 -v
```

서버 결과 JSON에서 그대로 발췌한 값:

```json
{
  "model": "gpt-5.6-sol",
  "http_status": 200,
  "event": "response.completed",
  "needle_recall": true,
  "output": "MAPLE71,OTTER83,CEDAR29",
  "usage": {
    "input_tokens": 921825,
    "output_tokens": 15,
    "total_tokens": 921840
  }
}
{
  "model": "gpt-6-astra",
  "http_status": 200,
  "event": "response.completed",
  "needle_recall": true,
  "output": "MAPLE71,OTTER83,CEDAR29",
  "usage": {
    "input_tokens": 921821,
    "output_tokens": 15,
    "total_tokens": 921836
  }
}
{
  "model": "gpt-5.6-sol",
  "http_status": 200,
  "event": "response.failed",
  "needle_recall": false,
  "output": "",
  "server_error_code": "context_length_exceeded",
  "server_error_message": "Your input exceeds the context window of this model. Please adjust your input and try again."
}
```

HTTP 200만으로 성공을 판단하지 않았다. 길이 초과 요청도 HTTP 200으로 연결된 뒤 SSE의 `response.failed`로 끝났다. 테스트 실행기의 PASS 역시 추론 성공을 뜻하지 않으므로, `response.completed`, usage, 표식 회수를 함께 판정했다.

## Baseline-attribution — 측정 기준

- 워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- 브랜치·기준 HEAD: `WT-unified-gateway` · `81c1d58f9`; 기존 변경이 있는 트리이며 HEAD만으로 전체 소스를 재현할 수 없다. overlay 소스와 해시는 원시 증거에 남겼다.
- 인증 경로: 기존 ChatGPT 구독 인증, `https://chatgpt.com/backend-api/codex/responses`. API 키 과금 경로로 전환하지 않았다.
- 요청 조건: `stream=true`, `store=false`, reasoning `medium`. `max_output_tokens` 없이 구독 서버 출력 정책을 사용했다.
- 입력: 반복 ASCII 사이에 `MAPLE71`, `OTTER83`, `CEDAR29` 세 표식을 처음·중간·끝에 배치했다. 성공 요청의 서버 usage에서 입력 토큰 수가 반복 횟수 + 71로 관찰되어 실패 길이도 이 보정식으로 추정했다.
- 요청별 제한: 150초, 요청 본문 16 MiB, 응답 2 MiB. 실제 사용자 문서나 인증 토큰을 보고서에 저장하지 않았다.
- 총 24회 요청: 정상 완료 14회, 길이 초과 종료 10회. 성공 요청 입력 usage 합계 11,167,662토큰. 이는 구독 차감량이나 비용 산정값이 아니다.

## 적용 판단 — 920,000 입력 운영 목표

두 모델에 공통으로 실증된 범위 안에서 920,000을 입력 운영 목표로 삼을 수 있다. 이는 관찰된 최대 성공점보다 Sol 1,825토큰, Astra 1,821토큰 아래다. 이 여유만으로 도구 결과·다음 턴·긴 출력의 안전성을 보장하지 못하므로 자동 압축은 목표보다 앞서 시작해야 한다.

시험용 자식 프로세스에서 `CLAUDE_CODE_MAX_CONTEXT_TOKENS=920000`과 `CLAUDE_CODE_AUTO_COMPACT_WINDOW=920000`을 넣었을 때 실제 Claude Code 2.1.269의 `/context`에 `18.7k/920k tokens (2%)`가 표시됐다. 이 시험은 표시 기능만 확인했으며 추론 요청을 보내지 않았다. 기존 실행기를 재사용한 결과의 `success:false`는 모델 선택 여정용 판정값이며, 이번 UI 판정은 `context_display_seen:true`, `observed_lines`, 빈 오류 목록을 근거로 한다. 프로세스는 정해 둔 관찰 구간 뒤 종료·정리했다.

**설치된 MoAI의 운영 설정은 이번 실증에서 변경하지 않았다.** 현재 구현에는 별도 적용 과제가 있다.

| 계층 | 현재 확인 내용 | 적용 시 필요한 처리 |
|---|---|---|
| Claude Code | 사용자 정의 GPT 모델에 920k 표시 가능 | launcher 정리 후 모델에 맞는 context 설정 주입 |
| MoAI catalog | GPT context 상수 272,000 | 실증 모델별 값과 미실증 모델 구분 |
| MoAI 요청 제한 | 1 MiB | 약 1.84 MB인 이번 성공 입력도 통과할 수 있도록 유한한 본문 상한 조정 |
| 토큰 추정 | JSON 문자 수 / 4 | 실제 서버 usage 및 길이 초과 응답과 결합; 정확한 토큰 상한 보장으로 오해하지 않기 |
| 압축·오류 복구 | 이번 시험 범위 밖 | 도구 포함 대화와 다음 턴에서 조기 압축·길이 초과 복구를 검증 |

## Gaps — 확인하지 않은 사항

- 1,050,000 총 context 지원 여부, 최대 출력과 입력의 동시 한도, 서버 내부 예약량.
- 자연어·코드·이미지·대규모 도구 스키마의 경계 및 장문 추론 품질. 표식 세 개 회수는 모든 토큰을 유효하게 활용한다는 증거가 아니다.
- 실제 Claude Code 대화가 MoAI의 모든 제한을 통과하여 920k 입력으로 답하는 통합 E2E. 서버 실험은 로컬 272k·1 MiB 제한을 우회한 시험 전용 transport다.
- 다른 계정·구독 등급·다른 GPT 모델·높은 reasoning 설정에서 같은 한도인지 여부.
- 구독 잔여량·차감률 및 과금 상세. 구독 경로 성공을 확인했지만 요금제 전체에 대한 보장은 아니다.

## Residual-risk — 남는 위험

서버 정책이나 모델 배포가 바뀌면 한도가 달라질 수 있다. 관찰 경계 가까이 설정하면 요청의 부가 필드나 도구 결과 때문에 다음 턴이 실패할 수 있다. 920k UI 표시는 서버 용량 확장의 증거가 아니며, 운영 반영은 요청 크기·모델별 설정·압축·오류 복구의 통합 검증을 거쳐야 한다.

## 근거 파일과 문서

- [실험별 원시 결과와 명령](context-window-evidence/e2e/summary.json)
- [독립 재집계 JSON](context-window-measurement.json)
- [920k UI 관찰](context-window-ui-probe.json)
- [실험 소스](context-window-evidence/e2e/context_probe_test.go)
- [Sol 공식 API 사양](https://developers.openai.com/api/docs/models/gpt-5.6-sol)
- [Astra 공식 API 사양](https://developers.openai.com/api/docs/models/gpt-6-astra)
- [Claude Code 사용자 정의 모델 context 설정](https://code.claude.com/docs/en/model-config#correct-the-window-for-a-gateway-or-custom-model-id)

## 전체 요청 판정

| 모델 | 파일 | 입력: 성공 실측 / 실패 추정 | 종료 이벤트 | 표식 회수 |
|---|---|---:|---|---|
| gpt-6-astra | astra-1000000.json | 약 1000071 | response.failed | 해당 없음 |
| gpt-6-astra | astra-800000.json | 800071 | response.completed | 성공 |
| gpt-6-astra | astra-921000.json | 921071 | response.completed | 성공 |
| gpt-6-astra | astra-921500.json | 921571 | response.completed | 성공 |
| gpt-6-astra | astra-921750-repeat.json | 921821 | response.completed | 성공 |
| gpt-6-astra | astra-921750.json | 921821 | response.completed | 성공 |
| gpt-6-astra | astra-921812.json | 약 921883 | response.failed | 해당 없음 |
| gpt-6-astra | astra-921875.json | 약 921946 | response.failed | 해당 없음 |
| gpt-6-astra | astra-922000.json | 약 922071 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-1000.json | 1071 | response.completed | 성공 |
| gpt-5.6-sol | sol-1000000.json | 약 1000071 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-400000.json | 400071 | response.completed | 성공 |
| gpt-5.6-sol | sol-800000.json | 800071 | response.completed | 성공 |
| gpt-5.6-sol | sol-872000.json | 872071 | response.completed | 성공 |
| gpt-5.6-sol | sol-921000.json | 921071 | response.completed | 성공 |
| gpt-5.6-sol | sol-921464.json | 921535 | response.completed | 성공 |
| gpt-5.6-sol | sol-921696.json | 921767 | response.completed | 성공 |
| gpt-5.6-sol | sol-921754-repeat.json | 921825 | response.completed | 성공 |
| gpt-5.6-sol | sol-921754.json | 921825 | response.completed | 성공 |
| gpt-5.6-sol | sol-921812.json | 약 921883 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-921928.json | 약 921999 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-921929.json | 약 922000 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-922000.json | 약 922071 | response.failed | 해당 없음 |
| gpt-5.6-sol | sol-950000.json | 약 950071 | response.failed | 해당 없음 |
