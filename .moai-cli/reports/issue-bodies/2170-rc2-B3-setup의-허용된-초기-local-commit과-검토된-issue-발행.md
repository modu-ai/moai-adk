# [rc.2][B3] SETUP의 '허용된 초기 local commit'과 '검토된 Issue 발행' 단계 — 승인 분리

상위: #2154 (EP-B). 설계 근거: 설계 §3 L294-295·L319-321; `skills/moai/workflows/setup.md`(단계 없음).

## 범위
- setup 절차에 두 단계 추가: 각각 별도 승인(commit 승인 ≠ Issue 공개 승인), dry-run 출력 먼저
- receipt에 두 효과 기록

## 요구(GEARS)
- REQ-B3-01: WHEN 사용자가 초기 commit을 승인하면 THE SETUP SHALL commit만 만들고 push·Issue 발행은 하지 않는다.
- REQ-B3-02: WHEN Issue 발행이 승인되면 THE SETUP SHALL 검토된 초안만 발행하고 readback으로 확인한다.

## 수용 기준
- AC-1: fixture로 승인 없이 commit/발행 0
- AC-2: readback 확인 시험
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
승인 분리 원칙 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
