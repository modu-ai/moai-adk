# [rc.2][A2] 에이전트 self-claim 루프 — `queue-next` op(원격 미배차 재계산 → journal+lock → claim 댓글 → 최소 id → projection)

상위: #2153 (EP-A). 설계 근거: autonomy-modes-spec.md §3; `src/claim.rs:1-23`.

## 범위
- `src/main.rs` OPERATIONS에 `queue-next`(`queue.next`) 추가: milestone의 열린 Issue 중 claim 댓글 없는 것(원격 정본, D-4) → 우선순위(blocked 제외·EP 순) → 후보 1건 → 기존 `claim-acquire` 호출 → 승/패 결과 JSON
- 패배(already_owned)는 journal에 쓰지 않는 현행 유지(claim.rs:906-911); 결과 JSON에 `lost_to: <comment-id>`만 반환(상태 추가 없음)
- 승자가 label `moai:in-progress`·소유 댓글 `상태:` 줄 갱신 — projection 전용, 실패해도 claim 유효
- `assignments.json`은 읽기만(레인), 쓰기 금지

## 요구(GEARS)
- REQ-A2-01: WHEN `queue-next`가 호출되면 THE SYSTEM SHALL 원격 Issue 목록을 다시 읽어 claim 댓글 없는 Issue만 후보로 삼는다.
- REQ-A2-02: WHEN 두 세션이 같은 Issue를 집으면 THE SYSTEM SHALL 가장 낮은 live 댓글 id를 승자로 하고 패자는 다음 후보로 넘어간다(자동 takeover 0).
- REQ-A2-03: THE SYSTEM SHALL label·assignee 변경 실패를 claim 실패로 취급하지 않는다.

## 수용 기준
- AC-1: RecordedTransport fixture로 두 세션 경합 시험 — 승자 1·패자 다음 후보 이동·중복 claim 0
- AC-2: blocked 카드(본문 `Blocked-By:` 헤더)가 후보에서 제외되는 시험
- AC-3: 패자 journal 미기록 확인 시험(현행 보존)
- AC-4: check set exit, OPERATIONS 수 갱신(47→48) 문서 반영

## 선행·차단
Blocked-By: #2160

## 사용자 결정
00:3x(11-09) self-claim; D-4 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
