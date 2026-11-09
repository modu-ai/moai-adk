# [rc.2][B9] MCP parity 조사 — RC1 CLI op ↔ Go MCP 38 tool 대응표(구현 blocked)

상위: #2154 (EP-B). 설계 근거: 설계 §1·§13; decisions 51(MCP 다음 버전).

## 범위
- 대응표: 각 OPERATIONS op에 대응 MCP tool 유무, 입력/출력 schema 차이, structuredContent 요구
- 결론: rc.2 안 MCP 범위 제안(구현은 별도 카드)

## 요구(GEARS)
- REQ-B9-01: THE REPORT SHALL 38 tool 전수를 다룬다.

## 수용 기준
- AC-1: 38행 전수
- AC-2: 결론 1절
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
MCP 다음 버전 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
