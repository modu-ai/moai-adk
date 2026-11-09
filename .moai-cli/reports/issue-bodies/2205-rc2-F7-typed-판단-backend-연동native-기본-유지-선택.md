# [rc.2][F7] typed 판단 backend 연동(native 기본 유지, 선택)

상위: #2158 (EP-F). 설계 근거: 설계 §20 F20; `src/decision.rs` P10.

## 범위
- F6 결론 뒤 연동 설계·구현; authority none 유지

## 요구(GEARS)
- REQ-F7-01: THE SYSTEM SHALL native 판단을 기본으로 하고 backend 결과에 authority를 부여하지 않는다.

## 수용 기준
- AC-1: 설계 승인
- AC-2: fixture
- AC-3: check set exit

## 선행·차단
Blocked-By: #2204

## 사용자 결정
RC1 결정 우선(00:4x) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
