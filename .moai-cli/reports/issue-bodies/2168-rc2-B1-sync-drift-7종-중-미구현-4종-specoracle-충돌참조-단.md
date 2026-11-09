# [rc.2][B1] SYNC drift 7종 중 미구현 4종 — SPEC↔oracle 충돌·참조 단절·Issue 완료인데 증거 stale·명시 link 의미 미검토

상위: #2154 (EP-B). 설계 근거: 설계 §5 drift 7분류; `skills/moai/workflows/sync.md:20`(현 3종).

## 범위
- sync 절차에 4종 판정 규칙·출력 어휘 추가(기존 3종과 대응표), 각 종별 fixture
- 코드 op 추가 여부는 레인 판단(절차+gh 유지 가능)

## 요구(GEARS)
- REQ-B1-01: WHEN drift 7종 중 하나가 탐지되면 THE SYNC SHALL 종류·근거·제안 조치를 분리해 보고하고 승인 SPEC을 자동 약화하지 않는다.

## 수용 기준
- AC-1: 4종 fixture 각 1건 탐지
- AC-2: 상위 7종↔RC1 어휘 대응표(docs/design-review/appendix 참조)
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
RC1 결정 우선 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
