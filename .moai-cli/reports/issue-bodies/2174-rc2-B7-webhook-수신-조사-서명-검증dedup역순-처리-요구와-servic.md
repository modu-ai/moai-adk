# [rc.2][B7] webhook 수신 조사 — 서명 검증·dedup·역순 처리 요구와 service-free 전제의 양립 조건(구현 blocked)

상위: #2154 (EP-B). 설계 근거: 설계 §6 L164; [0d45] L164.

## 범위
- 조사 문서: 수신자 없는 설계에서 webhook이 성립하는 조건(사용자 기존 수신 환경), GitHub 재전송 정책(공식 URL+확인일), polling과의 역할 분담
- 구현 카드는 조사 결론 뒤

## 요구(GEARS)
- REQ-B7-01: THE REPORT SHALL 외부 사실마다 URL과 확인일을 붙이고 확인 못 한 것은 미확인으로 둔다.

## 수용 기준
- AC-1: 외부 사실 전부 URL+날짜 또는 미확인
- AC-2: 결론 1절(채택/보류)
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
조사 우선 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
