# [rc.2][G1] Space 상위 페이지 적용 — fixes/top-5490c4ee.md 19항목 + 적용 후 사본 sha 확인

상위: #2159 (EP-G). 설계 근거: #2140 산출물; 결정 01:2x 리더 적용.

## 범위
- 리더: 항목별 요약 보고 → 사용자 확인 → ego-browser 적용 → 재수집 사본 sha256 기록(docs/design-review/applied.md)

## 요구(GEARS)
- REQ-G1-01: THE LEADER SHALL 적용 전 사용자 확인을 받고 적용 후 사본으로 검증한다.

## 수용 기준
- AC-1: 19항목 적용 체크표
- AC-2: 사본 sha 기록
- AC-3: 레인 작업 없음(리더)

## 선행·차단
Blocked-By: #2140

## 사용자 결정
01:2x — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
