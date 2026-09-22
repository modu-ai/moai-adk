# decision-index.md — SPEC-CODEX-LOCALMD-001

decision_gate: RESOLVED. 아래 R1~R4는 카드 t1078의 전체 구현·로컬 병합 요청에 따라 고정된 구현 판단이다. 각 행은 별도의 운영자 승인이나 개별 사용자 승인을 주장하지 않는다.

## R1 — 템플릿 §8 문구와 namespace 중립성

- **Detect**: 런처 계약 역전으로 `internal/template/templates/AGENTS.md.tmpl` §8의 "not forwarded to Codex" 단정은 부정확해지지만, 파일명을 다시 열거하면 namespace consolidation의 중립성을 훼손한다.
- **Resolution**: 템플릿은 공통 로컬 입력과 Codex-specific 입력의 역할을 정확히 구분한다. sync-phase의 CHANGELOG·docs-site 4-locale·추적되는 리포 `AGENTS.md` 설명은 갱신할 수 있지만 사용자 로컬 입력 두 파일은 모든 단계에서 수정하지 않는다.

## R2 — `moai codex status` readout

- **Detect**: 현재 status readout에는 local-instruction 행이 없지만, 구조 거부는 런치 오류로 관측할 수 있다.
- **Resolution**: 본 SPEC에서는 status 행을 추가하거나 출력 계약을 변경하지 않는다.

## R3 — TOCTOU 봉쇄와 실행 인자 상한

- **Detect**: Lstat→ReadFile은 check-then-reopen 경합을 허용하고, pre-quote 토큰 검사는 spawn shell-quote 팽창을 보장하지 않는다.
- **Resolution**: `CLAUDE.local.md`와 `AGENTS.local.md` 모두 symlink-follow 없는 안전 open 뒤 같은 descriptor를 fstat/read한다. direct 최종 토큰과 spawn 최종 명령 문자열은 각각 독립적으로 126,976-byte 기본 상한을 적용한다. 플랫폼이 동등한 안전 판독을 제공하지 못하면 fail closed한다.

## R4 — LIVE 실행 환경

- **Detect**: 실제 Codex 도달성은 인증 가능한 별도 세션이 필요하지만 소스 트리 로컬 파일에 nonce를 쓰면 입력 비가공 요구와 충돌한다.
- **Resolution**: Codex 레인이 독립 fixture에서 120초 hard timeout으로 exact launcher command를 실행한다. 응답과 session log 양쪽에 두 nonce/source 쌍이 있어야 PASS이며, 소스 트리 네 입력 파일은 사전·사후 해시로 불변을 증명한다. `NOT_RUN`은 전체 PASS가 아니다.
