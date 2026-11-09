# [rc.2][C10] 회귀 시험군 ID GRAPH·FRESHNESS·RETRIEVAL fixture + eval corpus(30~50 task)

상위: #2155 (EP-C). 설계 근거: 설계 §17; Core Brain L3703.

## 범위
- eval corpus 정의(위치 `tests/fixtures/atlas/`, 구성 기준), 세 시험군 fixture와 통과 기준 수치(baseline 뒤 고정)

## 요구(GEARS)
- REQ-C10-01: THE CORPUS SHALL 각 task의 ground truth와 출처를 기록한다.

## 수용 기준
- AC-1: corpus ≥30 task
- AC-2: 세 시험군 실행 결과 표
- AC-3: check set exit

## 선행·차단
Blocked-By: #2183

## 사용자 결정
RC1 결정 우선(00:4x) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
