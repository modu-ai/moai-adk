# [rc.2][F6] 판단 backend 조사 — Jev·OpenAI Decisions·Claude/Codex 구독 credit(URL+확인일, 비용 호출 0)

상위: #2158 (EP-F). 설계 근거: 설계 §18 '남은 결정' 8항; [e361] L147-149; decisions 01:1x.

## 범위
- 표: backend별 현행 API·요금·구독 credit 자격·확인일·URL; 미확인은 미확인
- Jev는 Go `internal/jev` 기준선 사실만

## 요구(GEARS)
- REQ-F6-01: THE REPORT SHALL 과거 문구만으로 혜택을 확정하지 않는다.

## 수용 기준
- AC-1: 표 전수 URL+날짜 또는 미확인
- AC-2: 비용 호출 0 선언
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:1x 조사 카드 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
