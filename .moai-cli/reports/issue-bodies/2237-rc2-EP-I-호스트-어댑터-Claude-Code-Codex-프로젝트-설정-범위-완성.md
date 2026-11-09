# [rc.2][EP-I] 호스트 어댑터 — Claude Code·Codex 프로젝트 설정 범위 완성

리더 조사(host-setup-scope 2026-11-09) 결론: moai-cli는 AGENTS.md managed 구간·AGENTS.local.md·호스트별 role/skill 파일·홈 `moai` skill 복사만 쓴다. 훅·settings/config 쓰기·skill 설치 완성·MCP 설정·permissions·실세션 검증은 카드가 없었다. 이 EP는 그 갭 8건을 묶는다. EP-H(H1 override·H2 Claude 어댑터·H5)는 Q-A/B/C 결정 뒤 별도.

## 카드
I1 양 호스트 훅 생성 · I2 Claude settings.json CLI · I3 TOML + Codex config.toml · I4 Codex skill 설치 완성 + openai.yaml · I5 프로젝트 단위 skill 복사(P04, 설계·fixture만) · I6 MCP 설정 생성(B9 결론 기반) · I7 permissions 설정 키 · I8 실세션 시험 계획·하네스(실행은 운영자)

## 사용자 결정
ADK_DELEGATED_CHOICES_20261109_01 — 사용자 위임에 따른 영실이 대리 판단(일상 선택). 기존 HOLD 유지.

## 항상 blocked(카드 발행으로 해제되지 않음)
실제 사용자 홈(`~/.claude`, `~/.codex`, `~/.agents`, `~/.moai-cli`) 설정 변경 · 권한/보안 설정 변경 · App 등록/키·토큰/지속 접근 · 유료 호출 · 배포·태그 push · 미승인 host 실세션 실행. 레인 산출물은 코드·문서·격리 fixture(임시 HOME/프로젝트 트리)까지. 미결정(EP-H Q-A/B/C, 프로젝트 skill 위치 P04)은 이 카드가 확정하지 않는다 — 결정 의존 부분은 '결정 대기'로 표기.

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 보고: `.moai-cli/reports/host-setup-scope-2026-11-09.md`(리더) + `host-setup-scope-verification-2026-11-09.md`(agent-1 [0dd806] 독립 재검증). 기준 tip d76a5b7ce.
