# [rc.2][D2] archive 자동 이동(hash/readback 필수) + restore + 검색

상위: #2156 (EP-D). 설계 근거: 설계 §12; decisions 00:5x 자동 이동 허용.

## 범위
- `archive-apply`: 이동 전 hash → 이동 → readback 일치 확인 → 인덱스 갱신; 실패 시 원위치 유지
- restore op, 아카이브 검색(비벡터)

## 요구(GEARS)
- REQ-D2-01: WHEN readback이 불일치하면 THE SYSTEM SHALL 이동을 취소하고 원본을 유지한다.

## 수용 기준
- AC-1: 성공/실패 fixture
- AC-2: restore 라운드트립
- AC-3: check set exit

## 선행·차단
Blocked-By: #2187

## 사용자 결정
00:5x — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
