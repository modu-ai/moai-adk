# [rc.2][C8] retrieval budget — 기본 256 KiB, profile 재정의, host별 tokenizer 측정

상위: #2155 (EP-C). 설계 근거: 설계 §11; decisions 01:1x budget; `src/context/mod.rs` bytes_div_4.

## 범위
- 기본값 상수 + `project.yaml` profile 키, 초과 시 omitted 기록(현행)
- tokenizer 측정 카드: Claude/Codex 실제 토큰 수와 bytes/4 오차 표(실측은 운영자 환경 — not_run 허용, 표 틀만)

## 요구(GEARS)
- REQ-C8-01: WHEN budget이 없으면 THE SYSTEM SHALL 256 KiB 기본을 쓰고 출처를 기록한다.

## 수용 기준
- AC-1: 기본/재정의 시험
- AC-2: 오차 표 틀 + 측정 절차
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:1x budget — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
