# [rc.2][A6] 같은 session 값 claim 충돌 방지 — REQ-AM-04(session id에 세션 ref 포함, 동명 세션 adopt 금지)

상위: #2153 (EP-A). 설계 근거: autonomy-modes-spec.md REQ-AM-04; `src/claim.rs:829-845`.

## 범위
- claim 요청의 `owner.id`에 세션 고유 ref(예: `agent-1-e29c15`)를 필수화하거나, 동일 session 문자열의 기존 claim을 adopt하지 않고 `uncertain`으로 반환하는 옵션 중 하나(레인 판단, 근거 기록)
- 오늘 실례(agent-1 중복 세션)를 fixture로

## 요구(GEARS)
- REQ-A6-01: WHEN 같은 session 값의 claim 댓글이 이미 있고 요청자가 그 댓글의 작성 세션임을 증명할 수 없으면 THE SYSTEM SHALL adopt하지 않는다.

## 수용 기준
- AC-1: 동명 세션 fixture에서 adopt 0·uncertain 또는 ref 불일치 거절
- AC-2: 기존 claim 시험 불변, check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:0x REQ-AM-04 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
