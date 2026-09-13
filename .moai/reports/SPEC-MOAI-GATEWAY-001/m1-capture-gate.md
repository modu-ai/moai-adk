# M1 — 실제 TUI 원본 요청 캡처 게이트

Verdict: BLOCKED
Blocker: M1-B1 — 검증 요청에 `stream` 키가 없음

## Claim

Claude Code 2.1.268의 실제 TUI에서 `/model gpt-5.6-sol`을 실행했을 때 수신한 검증 요청 원문에는 `stream` 키가 없다. 현재 가정 A-VAL-2.1.267은 `stream` 키의 존재와 JSON `false`를 모두 요구하므로 이 실제 검증 요청을 인정하지 않는다. `plan.md` M1 진입 게이트와 `design.md` §4.1이 명시한 중단 조건에 해당한다.

첫 turn과 변경 뒤 turn은 같은 TUI에서 로컬 mock 응답을 받았다. 이 결과는 두 모델 ID의 client 수준 전환 관측이며 제품 gateway의 3사 라우팅 PASS는 아니다.

## Evidence

실행 명령:

```sh
timeout 140 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m1_capture.py
```

두 번째 시도 exit 0. 첫 시도는 임시 프로젝트의 신뢰 화면에서 기본 선택 `No, exit`를 Enter로 선택하여 종료되었다. 화면을 읽은 뒤 직접 만든 임시 프로젝트에 대한 `Yes` 선택으로 시험 조작을 수정했다. 첫 시도는 제품이나 인식기 실패로 세지 않는다.

최종 TUI 화면에 남은 실제 출력:

```text
❯ Reply only CAPTURE_FIRST.
⏺ MOCK_OK:claude-sonnet-4-5
❯ /model gpt-5.6-sol
  ⎿  Set model to gpt-5.6-sol and saved as your default for new sessions
❯ Reply only CAPTURE_LATER.
⏺ MOCK_OK:gpt-5.6-sol
```

원문 파일을 Python으로 다시 파싱한 검증 요청 판정 출력:

```json
{"file": "request-003.json", "sha256": "0c65416916579d77028e64d1fbbec322055f07b3464e9b3aa1ee0ebe06c133e9", "model": "gpt-5.6-sol", "conditions": {"stream_explicit_false": false, "max_tokens_one": true, "one_user": true, "tools_empty": true}, "recognizer_match": false, "title_instruction_present": false, "local_path_present": false}
```

캡처 코드의 요약 출력 중 해당 항목:

```json
{
  "file": "request-003.json",
  "method": "POST",
  "path_query": "/v1/messages?beta=true",
  "model": "gpt-5.6-sol",
  "stream_present": false,
  "stream": null,
  "max_tokens": 1,
  "messages": 1,
  "roles": ["user"],
  "tools": 0
}
```

위 요약의 `stream: null`은 `dict.get`의 파생 표현이다. 원문에는 `stream: null`도 `stream: false`도 **없다**. 원문 파일을 별도로 읽어 키 부재를 확인했다.

전체 POST 다섯 건은 아래와 같다. 모두 경로·질의 문자열은 `/v1/messages?beta=true`였다.

| 원문 | 종류 | model | stream | max_tokens | messages / tools |
|---|---|---|---|---:|---:|
| request-001.json | 첫 prompt의 제목 생성 | claude-sonnet-4-5 | true | 32000 | 1 / 0 |
| request-002.json | 첫 turn | claude-sonnet-4-5 | true | 32000 | 1 / 29 |
| request-003.json | `/model` 검증 | gpt-5.6-sol | **키 없음** | 1 | 1 / 0 |
| request-004.json | 이후 prompt의 제목 생성 | gpt-5.6-sol | true | 32000 | 1 / 0 |
| request-005.json | 이후 turn | gpt-5.6-sol | true | 32000 | 5 / 25 |

001·004의 user 본문에 제목 작성 지시 `Write the title in the predominant language of the session`이 있음을 직접 읽었다. 캡처한 다섯 건 가운데 다른 `max_tokens: 1` 요청은 없었다. 관측 범위 밖 요청의 부재는 주장하지 않는다.

## Baseline-attribution

- 작업 트리 `WT-unified-gateway`, HEAD `81c1d58f9`, SPEC 0.7.0.
- `m1-conditions.json`: `version: 2.1.268 (Claude Code)`, `suppression_flags_present: []`.
- 실제 TUI를 별도 tmux 서버와 임시 config·프로젝트에서 실행했다. `env -i`로 환경을 구성하여 네 억제 플래그를 모두 전달하지 않았다.
- 응답은 로컬 mock만 생성했다. 기존 mock의 프로토콜 응답을 재사용하되 POST 본문 바이트는 파생 변환 전에 그대로 별도 저장하고 경로·질의 문자열을 함께 기록했다.
- 스크립트 finally에서 이 시험 전용 tmux 서버와 HTTP 서버를 종료했다. Claude 실행에는 별도 timeout도 걸었다.
- 증거 디렉터리: `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/`. `m1-raw/`, `m1-index.json`, `m1-summary.json`, `m1-conditions.json`, `m1-screen-final.txt`를 보존했다.

## Gaps

- 현재 SPEC이 캡처 불일치 시 픽스처 고정을 금지하므로 `internal/gateway/testdata/`에 복사하지 않았다.
- turn 원문 002·005에는 임시 기계 경로가 있다. 추적 픽스처로 옮기기 전 경로·metadata 처리와 원본 해시의 연결을 명시해야 한다. 원본은 기계 로컬 증거로 보존한다.
- 이 캡처만으로 모든 가능한 비스트리밍 요청이 검증 요청과 구분된다고 증명하지 않는다.
- Claude→GPT 두 turn만 관측했다. GLM→Claude까지의 네 turn, 제품 gateway, picker `s`, tool 왕복, provider 인증은 수행하지 않았다.
- 억제 플래그를 전달하지 않았다는 조건을 확인했으며, client 내부의 부가 트래픽 정책 전체를 역공학하지 않았다.

## Residual-risk 및 다음 결정

`design.md` §4.1의 인식 조건과 `acceptance.md` AC-MG-003(c)의 `stream` 누락 변형 기대값은 실제 캡처와 충돌한다. 단순히 bool 변환을 되살리면 `null`·문자열 등도 섞일 수 있으므로 구현에서 임의 완화하지 않는다.

다음 수정 후보는 명시 `false`와 키 부재의 허용 여부를 원본 캡처에 근거하여 결정하고, 나머지 형태·실제 turn·제목 요청 대조를 함께 고정하는 것이다. SPEC의 해당 조건과 변형 기대값을 먼저 고치고 변경분 감사를 거친 뒤 인식기 구현을 재개한다. 별도 M0의 T09 양성 판정은 아직 INCONCLUSIVE다.
