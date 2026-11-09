# [rc.2][D6] 문서/소스/template/live 의미 parity 검사

상위: #2156 (EP-D). 설계 근거: 설계 §18 6단계 '의미 parity'; 00-A baseline 분류.

## 범위
- parity op: 템플릿↔생성물↔문서의 키·값 차이 보고(00-A 분류 재사용)

## 요구(GEARS)
- REQ-D6-01: THE CHECK SHALL 차이를 종류별(누락/추가/값 상이)로 분리한다.

## 수용 기준
- AC-1: fixture 차이 탐지
- AC-2: 보고 형식
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
RC1 결정 우선(00:4x) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
