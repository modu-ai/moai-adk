# [rc.2][A7] lease 기본 4h 상수·만료 표시·수동 인계 절차 문서

상위: #2153 (EP-A). 설계 근거: decisions 00:5x(11-09) P08 잔여; `src/claim.rs:25-28`.

## 범위
- `lease_until` 미지정 시 기본 now+4h(상수 `DEFAULT_LEASE`), 명시값 우선; 만료는 `claim-show`에 `expired: true`로 표시만, 자동 takeover 0(현행)
- handoff 절차 문서: 만료 확인 → 리더가 epoch+1 인계 → 기록

## 요구(GEARS)
- REQ-A7-01: WHEN lease_until이 없으면 THE SYSTEM SHALL 4시간 기본값을 쓴다.
- REQ-A7-02: THE SYSTEM SHALL 만료만으로 소유권을 옮기지 않는다.

## 수용 기준
- AC-1: 기본값 시험·명시값 우선 시험
- AC-2: 만료 표시 시험 + takeover 0 기존 시험 불변
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
00:5x lease 4h — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
