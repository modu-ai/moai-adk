# [rc.2][C4] 온톨로지 노드/엣지 schema — basis·verification 분리, tombstone

상위: #2155 (EP-C). 설계 근거: 설계 §8; Core Brain L741·L752-810·L2413-2491(세 벌 통합).

## 범위
- 노드 8종·엣지 통합 목록(DR-07 부록 기준) JSON schema, `basis`(명시/추론)·`verification`(검증 상태) 분리 필드, tombstone 레코드
- JSON 정본(TOON 없음)

## 요구(GEARS)
- REQ-C4-01: THE SCHEMA SHALL 명시 link와 추론 link를 같은 필드에 섞지 않는다.

## 수용 기준
- AC-1: schema 검증 시험(유효·무효 각 ≥5)
- AC-2: tombstone 라운드트립
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:2x JSON 정본 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
