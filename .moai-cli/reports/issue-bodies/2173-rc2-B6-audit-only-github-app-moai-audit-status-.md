# [rc.2][B6] audit-only GitHub App — `moai/audit` status 출처 고정(운영자 App 등록 keep-set)

상위: #2154 (EP-B). 설계 근거: AGENTS.local.md 게이트 known gap; decisions 01:1x.

## 범위
- App 권한 최소(statuses:write, pull_requests:read) 설계 문서 + 토큰 주입 경로(환경 변수 이름만, 값 저장 0)
- status 게시 코드가 App 토큰이면 `context`에 출처 표시; 공유 계정이면 경고
- App 등록·설치는 운영자 1회(keep-set) — 카드는 준비까지

## 요구(GEARS)
- REQ-B6-01: WHEN App 토큰이 있으면 THE GATE SHALL status를 App으로 게시하고 없으면 경고 후 기존 경로.

## 수용 기준
- AC-1: 두 경로 fixture 시험
- AC-2: 비밀 값 로그 0
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:1x audit App — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
