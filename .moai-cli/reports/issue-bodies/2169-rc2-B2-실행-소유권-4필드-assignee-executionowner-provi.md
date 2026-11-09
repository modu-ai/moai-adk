# [rc.2][B2] 실행 소유권 4필드 — assignee / execution_owner / provider_session / integration_owner schema + claim record

상위: #2154 (EP-B). 설계 근거: 설계 §6 L756; `src/claim.rs` record.

## 범위
- claim record·handoff record에 4필드 추가(기존 `owner`는 execution_owner로 별칭), integration_owner는 리더 게이트 식별자
- assignee는 표시 전용(lock 아님) 명시

## 요구(GEARS)
- REQ-B2-01: THE RECORD SHALL 네 필드를 분리 저장하고 어느 것도 lock으로 해석하지 않는다.

## 수용 기준
- AC-1: schema 시험 + 기존 claim 시험 불변
- AC-2: 문서(protocol-v1.md) 갱신
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
RC1 06-A 유지 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
