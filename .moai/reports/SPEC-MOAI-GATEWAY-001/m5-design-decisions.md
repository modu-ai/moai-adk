# M5 변환 결정 검토안 — SPEC 0.8.0

## Claim

문서 결정과 미결 후보를 분리했다. system text/order, optional schema, stop mapping, 요청별 도구 이름 복원은
구현 대상 결정이며 opaque reasoning 운반은 PROBE ONLY다. 이 보고서는 구현·호환성·계획 감사 PASS가 아니다.

## 결정

1. 최상위 system text는 두 줄바꿈으로 연결한 instructions, messages 내부 system은 같은 위치 Responses 메시지.
2. function strict:false, max_tokens→max_output_tokens, 충돌 없는 요청별 tool name 역매핑.
3. completed text=end_turn, completed tool=tool_use, 유효 공개 완료 블록이 있는 max_output_tokens 제한=max_tokens.
   나머지 incomplete/error/failed/truncated EOF는 오류이며 성공 terminal을 합성하지 않는다.
4. 공개 text/tool pair history 보존은 opaque reasoning과 독립 검증.

## 미결 게이트

redacted_thinking.data의 버전 있는 MoAI/OpenAI envelope는 후보다. Anthropic 서명으로 주장하지 않으며 다른 provider로
전달하지 않는다. Claude TUI의 보존·도구 왕복·same-provider resume과 전체 carrier 유실 탐지 계약이 없다면 제품 비활성이다.
전체 envelope가 사라지면 stateless decoder는 최초부터 부재인 요청과 구분하지 못한다. tool call ID hash binding 또는
resume 가능한 세션 한정 lineage가 검토 대안이며 아직 어느 것도 채택하지 않았다. 전역 previous_response_id 표는 금지한다.

## Evidence와 Baseline-attribution

작성 기준: 지정 moai-proxy-unified WT의 SPEC 0.8.0을 읽고 좁게 보강했다. 공식 문서 판독·실패한 스키마 열기의 범위는
research §18에 기록했다. 코드 실행·실제 TUI·실제 provider 시험은 이 작성 작업에서 수행하지 않았다.

작성 뒤 지정 WT에서 실행한 구조 검사:

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

Exit code 0. 기존 HEAD 81c1d58f9 빌드 바이너리를 사용했으며 이번 작성자는 재빌드하지 않았다.

## Gaps 및 Residual-risk

문서 변환 규칙은 실제 provider 수용의 증명이 아니다. strict/schema·종료 이벤트·도구 이름·system 위치는 golden과
실행으로 검증해야 한다. opaque 원문이 전부 유실되는 경우를 잡는 계약 없이 decoder 검증만 통과시키면 잘못된 성공이 된다.
후속 probe는 2026-09-11 19:00 Asia/Seoul 이후 수행하며 후보 실패를 이유로 사용자 GPT 전체 목표를 축소하지 않는다.
