# reasoning 유실 탐지 후보 — 작성자·감사관 인계

이 문서는 부모의 설계 후보이며 SPEC 결정이나 제품 활성화 승인이 아니다. 코어 design §4.2·AC-MG-009가 미결로 둔
whole-carrier loss binding의 구체적 실험 입력이다. 2026-09-11 오후 7시 이후 실제 Claude Code 왕복을 측정한 뒤
manager-spec이 본문 계약을 확정하고 감사해야 한다.

## 후보

OpenAI가 반환한 공개 function call을 Anthropic tool_use로 내보낼 때, 원래 call ID와 해당 assistant 응답의 ordered
opaque envelope bytes에 대한 SHA-256을 버전 있는 tool ID 안에 가역 인코딩한다. 예시 모양은
`toolu_moai_v1_<base64url compact JSON>`이다. compact JSON의 필드는 `call_id`와 `opaque_sha256` 두 개다.
이는 서명·인증·OpenAI가 보증한 ID가 아니다. 비밀을 담지 않는 유실 탐지 표식이다.

왕복 입력에서 marker가 있는 tool_use/tool_result는 같은 assistant 묶음의 envelope와 결합한다. envelope hash와
모든 공개 tool pair의 대응이 일치할 때만 원래 OpenAI call ID를 양쪽에 복원한다. envelope 전체·일부가 없어지거나
표식이 깨지면 외부 송신 전에 명시 오류로 끝낸다. 다른 provider로 보낼 때에는 OpenAI opaque item을 제거하되 공개
tool pair의 연결과 원래 ID를 가역 복원하는 별도 계약이 필요하다.

이 선택은 request 사이 전역 mutable last-provider 값이나 사용자 저장소에 대한 임의 대화 복사를 요구하지 않는 후보다.
같은 provider resume에서도 tool ID와 envelope가 대화 기록에 남아 있으면 재구성이 가능할 것으로 예상하지만 아직
관측하지 않았다. Anthropic tool ID가 이 길이·문자열을 수용하고 도구 결과에서 그대로 돌아오는지도 미검증이다.

## 탐지 범위와 금지할 과장

- tool marker가 남은 상태의 envelope 전체 삭제 및 부분 변조가 대상이다.
- marker와 envelope를 함께 지우거나 전체 공개 대화를 다시 쓰는 변화까지 stateless하게 탐지한다고 주장하지 않는다.
- tool call이 없는 응답의 reasoning 유실 요구는 별도로 판정해야 한다. 공급자가 그 item을 다음 요청에 필수로 요구하는지
  실제 문서·응답으로 확인하고, 필요하다면 세션 lineage를 추가로 설계해야 한다. 편의상 필수 아님으로 단정하지 않는다.
- hash는 공개 consistency check이며 소유자 인증이 아니다. OpenAI encrypted item의 진위 판정은 공급자 소관이다.
- 수신 JSON의 중복 필드·버전·크기·허용 문자·hash·원래 ID를 검증하고, tool 이름과 ID의 역매핑을 혼동하지 않는다.
- 현재 AC의 공개 tool pair ID 보존을 가역 wire encoding으로 구체화할지 SPEC 작성자가 명시해야 한다.

## 준비된 시험

`picker_reasoning_after19.py`의 synthetic tool ID에 위 두 필드 형태를 적용했다. synthetic reasoning marker만 사용하며
실제 encrypted item이나 사용자 token을 사용하지 않는다. conditions.json에는 binding이 probe candidate임을 기록한다.
이미 준비된 Read 도구 결과에서 이 긴 ID와 synthetic marker가 같은 요청으로 돌아오는지를 관측한다.

변경 후 `ast.parse`는 `SYNTAX_OK`를 출력했다. 시각 차단 실행의 실제 출력은 다음과 같다(exit 0).

```json
{"status":"DEFERRED","now_utc":"2026-09-11T06:34:28.936188+00:00","not_before_utc":"2026-09-11T10:00:00+00:00","models":["gpt-6-astra","gpt-5.6-sol","gpt-5.6-terra","gpt-5.6-luna","claude-opus-5","claude-sonnet-5"],"listener_started":false,"client_started":false,"network_started":false}
```

위 JSON은 도구 출력과 동일한 값이며 문서에서는 공백만 줄였다. 실제 TUI, 실제 OpenAI reasoning, resume, foreign-provider
strip, deletion mutant의 송신 0은 아직 측정하지 않았다. 이 준비만으로 AC-MG-009의 BLOCKED 상태가 바뀌지 않는다.
