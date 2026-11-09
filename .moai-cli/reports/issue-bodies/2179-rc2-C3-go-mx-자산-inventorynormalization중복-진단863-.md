# [rc.2][C3] Go MX 자산 inventory·normalization·중복 진단(863 태그 기준선)

상위: #2155 (EP-C). 설계 근거: Core Brain L2852; `moai-adk-go@4fa2aa9 internal/navigator/sync/mx_bridge.go:54-63`.

## 범위
- 읽기 전용 inventory op: Go 기준선 태그 수·종류·파일 분포를 재현 명령과 함께 산출(863·28.8k·7 MB 수치 재측정)
- 중복·고아 태그 진단 보고(수정 0)

## 요구(GEARS)
- REQ-C3-01: THE INVENTORY SHALL 재현 명령과 기준 commit을 출력에 포함한다.

## 수용 기준
- AC-1: 수치 재측정 + 차이 설명
- AC-2: Go 트리 무변경(해시)
- AC-3: check set exit

## 선행·차단
Blocked-By: #2178

## 사용자 결정
Go 격리 유지 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
