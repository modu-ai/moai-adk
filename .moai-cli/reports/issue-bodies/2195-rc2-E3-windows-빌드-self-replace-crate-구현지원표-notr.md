# [rc.2][E3] Windows 빌드 + self-replace crate 구현(지원표 not_run)

상위: #2157 (EP-E). 설계 근거: decisions 12 self-replace 1.5.1; Cargo.toml 의존성 없음.

## 범위
- self-replace 의존성 추가(버전 고정), Windows 실행 중 교체 경로 구현, CI windows build job(테스트는 not_run 표기 가능)

## 요구(GEARS)
- REQ-E3-01: WHEN Windows에서 UPDATE가 실행되면 THE SYSTEM SHALL self-replace로 교체하고 실패 시 known-good을 유지한다.

## 수용 기준
- AC-1: 컴파일·단위 시험(모의)
- AC-2: 지원표 not_run 표기
- AC-3: check set exit

## 선행·차단
Blocked-By: #2193

## 사용자 결정
RC1 결정 우선(00:4x) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
