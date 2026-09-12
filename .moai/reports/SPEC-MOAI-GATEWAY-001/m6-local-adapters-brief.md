# M6 로컬 adapter 구현 준비

이 문서는 실행 지시 준비 기록이며 구현·검증 완료 보고가 아니다.

## 범위

현재 plan.md M6는 GLM Messages 중계와 Anthropic API key 중계를 허용한다. Anthropic 구독 OAuth passthrough는 M0 양성 결과 전 구체 타입·catalog 항목·전달 코드를 만들지 않는 계약을 유지한다. OpenAI adapter는 별도 작성자가 맡고 있다.

후속 작성자는 spec/design/acceptance의 현재 계약 및 기존 glmcred 저장소 API를 먼저 읽는다. GLM은 저장소 credential 참조를 사용하고 beta 헤더를 제거한다. Anthropic API key 경로는 명시적으로 공급된 credential만 사용한다. 고정 provider endpoint 외 주소, redirect, 다른 provider로의 fallback은 허용하지 않는다. 클라이언트의 인증·임의 헤더를 그대로 복사하지 않는다.

## 선행 경계

- Claude·Codex·provider 실제 시험은 2026-09-11 19:00 Asia/Seoul 이후에 실행한다. 그전에는 로컬 HTTP fixture만 사용한다.
- 타 provider 이력의 opaque 처리 계약은 별도 reasoning 보존 게이트와 함께 확정한다. ordinary Messages 중계 성공을 opaque 이력 호환성으로 보고하지 않는다.
- credential 세대 확인과 실제 송신 경계는 기존 AUTH 감사·수리 결과를 소비한다. Generation을 읽은 것만으로 원자적 송신을 증명했다고 하지 않는다.
- SSE 성공 종료·중간 오류·EOF·취소에서 body와 worker가 정리되는지 시험한다. upstream 401/429와 Retry-After는 민감 본문 노출 없이 전달한다.
- root/CLI 제품 factory 연결과 구독 passthrough 활성화는 이 지시 범위 밖이다.

## 준비 근거

부모는 현재 WT의 plan.md 226–234 및 design.md의 adapter 경계·GLM beta 제거 계약을 읽었다. 아직 GLM/Anthropic adapter 구현이나 해당 시험은 실행하지 않았다. baseline HEAD는 81c1d58f9이며 다른 작성자의 미커밋 변경이 함께 존재한다.
