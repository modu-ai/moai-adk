# [rc.2][E1] cargo-dist 도입 — sha256·SBOM 산출(결정 A, 미구현)

상위: #2157 (EP-E). 설계 근거: decisions 7; docs/release-checklist.md:29.

## 범위
- `dist-workspace.toml` 설정, 로컬 `cargo dist build` 산출물 검증, SBOM 포함 설정; 실제 release 게시는 카드 범위 밖

## 요구(GEARS)
- REQ-E1-01: THE BUILD SHALL 바이너리마다 sha256과 SBOM을 산출한다.

## 수용 기준
- AC-1: 로컬 dist build 산출물 목록
- AC-2: 체크리스트 갱신
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:1x SBOM=cargo-dist — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
