# [rc.2][E5] tag ruleset `moai-cli-v*` 생성(사용자 승인 1회, keep-set)

상위: #2157 (EP-E). 설계 근거: decisions 9; release-checklist.md:62.

## 범위
- ruleset JSON 초안 + 적용 절차, 적용은 운영자 승인 뒤 리더

## 요구(GEARS)
- REQ-E5-01: THE RULESET SHALL `moai-cli-v*` 외 tag 생성을 막지 않는다(Go tag 격리).

## 수용 기준
- AC-1: 초안 JSON
- AC-2: 적용 후 readback(승인 뒤)
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
keep-set — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
