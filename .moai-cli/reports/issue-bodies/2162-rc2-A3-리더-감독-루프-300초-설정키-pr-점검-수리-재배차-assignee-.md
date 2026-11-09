# [rc.2][A3] 리더 감독 루프 — 300초 설정키, PR 점검, 수리 재배차 `assignee-hint`, assignments 스냅샷 커밋

상위: #2153 (EP-A). 설계 근거: autonomy-modes-spec.md §4; AGENTS.local.md § Assignment record and supervision(#2144).

## 범위
- 설정 키 `leader.pr_check_interval_secs`(기본 300, 문서에 '사용자 지정값·측정값 아님') — `.moai-cli/project.yaml` profile 관리 키 추가
- `leader-tick` op: 열린 PR 상태(CI·moai/audit·충돌·BEHIND)·claim 댓글·assignments 대조 → 변경 요약 JSON; 변화 없는 PR은 보고 생략(폴링 증식 금지)
- 수리 필요 PR: Issue 댓글에 `assignee-hint: agent-N`(원 작성자 아닌 레인) + 수리 Issue 본문 초안 생성(발행은 리더 승인)
- `assignments.json` 스냅샷 재생성(리더만 커밋)

## 요구(GEARS)
- REQ-A3-01: WHEN 간격이 지나면 THE LEADER SHALL 열린 PR과 claim 댓글을 다시 읽고 변화가 있을 때만 보고한다.
- REQ-A3-02: WHEN PR이 FAIL verdict 또는 CI 실패이면 THE LEADER SHALL 원 작성자 아닌 레인에 assignee-hint를 남긴다(소유권은 claim 댓글로만).

## 수용 기준
- AC-1: fixture로 PR 상태 변화 유무에 따른 보고/생략 시험 2건
- AC-2: assignee-hint 생성 시험(원 작성자 제외)
- AC-3: profile 키 추가 후 `profile.rs` 관리 키 시험 갱신, check set exit

## 선행·차단
Blocked-By: #2161

## 사용자 결정
D-2 300초; 00:3x 리더 감독 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
