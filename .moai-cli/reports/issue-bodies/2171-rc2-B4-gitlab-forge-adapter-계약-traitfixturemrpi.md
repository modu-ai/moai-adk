# [rc.2][B4] GitLab Forge adapter 계약 — trait+fixture(MR·pipeline·note), 실제 구현은 blocked

상위: #2154 (EP-B). 설계 근거: 설계 §6 Forge Adapter; `docs/rust-product-path.md:238`.

## 범위
- Forge trait에 GitLab capability 선언(unknown/documented), fixture 스키마만
- 지원표에 GitLab 행 'documented, not_run'

## 요구(GEARS)
- REQ-B4-01: THE ADAPTER SHALL GitLab을 capability 선언으로만 노출하고 실호출 경로를 만들지 않는다.

## 수용 기준
- AC-1: trait 컴파일·fixture 로드 시험
- AC-2: 지원표 행
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
01:1x GitLab 계약만(계정 없음) — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
