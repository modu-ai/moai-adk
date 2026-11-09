# [rc.2][C9] Code Map / Code Graph 조회 op

상위: #2155 (EP-C). 설계 근거: 설계 §7; Core Brain L4302 context 선택.

## 범위
- `atlas-map`·`atlas-graph` 읽기 op: 파일/모듈 지도, 심볼 관계 질의, 출력 JSON(C4 schema)
- 쓰기 0

## 요구(GEARS)
- REQ-C9-01: THE OPS SHALL 읽기 전용이며 색인이 없으면 unavailable을 반환한다.

## 수용 기준
- AC-1: fixture 질의 시험
- AC-2: unavailable 경로
- AC-3: check set exit

## 선행·차단
Blocked-By: #2182, #2183

## 사용자 결정
RC1 결정 우선(00:4x) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
