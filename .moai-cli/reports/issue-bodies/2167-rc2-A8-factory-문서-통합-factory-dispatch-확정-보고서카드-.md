# [rc.2][A8] Factory 문서 통합 — factory-dispatch '확정 보고서→카드', lane loop·감독 절 정리, 임시 절 제거

상위: #2153 (EP-A). 설계 근거: AGENTS.local.md(§ Lane loop, § Assignment record and supervision); 보고서 §10.2 F-E.

## 범위
- AGENTS.local.md를 '세션·역할 / 에이전트 루프(self-claim) / 리더 감독 / 게이트' 4절로 재구성, 중복 문장 제거, 32 KiB 예산 확인
- docs/lane-loop-contract.md와 autonomy-modes-spec.md 상호 참조
- 코드 변경 0

## 요구(GEARS)
- REQ-A8-01: THE DOCS SHALL self-claim·감독·게이트 규칙을 각 한 곳에서만 정의하고 나머지는 참조한다.

## 수용 기준
- AC-1: 규칙 중복 grep(같은 문장 2곳 이상) 0
- AC-2: AGENTS.local.md+AGENTS.md 합계 < 32 KiB
- AC-3: check set exit

## 선행·차단
Blocked-By: #2160, #2161, #2162

## 사용자 결정
00:4x AGENTS.local 추가 지시 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
