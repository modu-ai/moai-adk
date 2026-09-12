# 최신 진행 보정 — 2026-09-11 20:54 KST

이 보정은 아래 20:02 스냅샷보다 우선한다. 전체 제품 인수는 아직 미완료다.

- Windows API 저장 기준과 GPT 구독 서버 출력 정책은 사용자 승인 후 SPEC에 반영했다.
- Windows 필수 시험 목록은 기존 9개를 보존한 23개다. 실제 Windows CI는 미실행이다.
- Windows AUTH 구현 감사에서 후보 파일 identity의 지연 조회 결함이 확인되었다. 구현 담당이 handle 기반 identity로 수리하고 macOS AUTH race 및 Windows 교차 컴파일을 통과했다고 보고했다. 변경분 독립 재감사는 남아 있다.
- Factory/SSE의 두 결함은 수정 후 독립 재감사를 통과했다. 새 모델 정책·구독 출력 정책은 구현 후 독립 감사 중이다. Native SSE 이벤트별 CRLF 바이트 계수 문제는 원문 소비량 기준으로 수리했으며, 변경분 독립 재감사에서 PASS를 받았다. 실제 provider 재송신과 Windows CI는 별도 공백이다.
- Receipt core의 마지막 guard 취소 경계를 수리하고 독립 재감사에서 PASS를 받았다. Reasoning envelope와 도구 ID codec 부품 및 retained conversation family/index 부품을 구현했다. 전체 대화 귀속·launcher 연결은 아직 통합 중이다.
- 실제 GPT 네 모델의 제목 JSON 응답과 native fork의 로컬 보존 관측을 추가했다. 각각 직접 provider 및 native 모의 경로의 증거이며 제품 통합 완료를 뜻하지 않는다.
- launcher 연결 절편에서 project-root 오류 계약과 퇴역 CG fixture를 정리했으며 관련 CLI 회귀가 통과했다. Production factory·receipt/opaque HTTP 경로·실제 `moai gpt` end-to-end는 아직 활성화하지 않았다.

근거: `windows-store-functional-review.md`, `windows-store-identity-repair.md`, `factory-sse-functional-review-iter2.md`, `native-policy-verification.md`, `receipt-final-guard-repair.md`, `title-policy-runtime-observation.md`, `native-fork-runtime-observation.md`.

기준: WT-unified-gateway / 81c1d58f9의 미커밋 트리. source_session_id는 developer 제공 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이며 현재 WT CLI는 environment-fallback이다. 각 시험의 명령과 원문 출력은 해당 근거 보고서에 귀속한다. Windows 실환경·전체 인수·원격 게시 완료를 주장하지 않는다.
