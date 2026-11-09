# [rc.2][C2] `@MOAI:` grammar + parser + `@MX:` dual-read 스캐너

상위: #2155 (EP-C). 설계 근거: 설계 §9 L24(dual-read→normalization→진단→opt-in migration); Core Brain L3470-3475.

## 범위
- `@MOAI:<KIND>[:<sub>] <args>` grammar 정의(상위 §9와 Core Brain 예시 통일 — DR-07 부록 참조), parser version 필드
- 스캐너: `@MX:`·`@MOAI:` 둘 다 읽어 같은 노드로 정규화, 쓰기 migration은 하지 않음

## 요구(GEARS)
- REQ-C2-01: WHEN 파일에 두 namespace가 섞여 있으면 THE SCANNER SHALL 둘 다 읽고 중복을 진단만 한다.
- REQ-C2-02: THE SCANNER SHALL 기존 파일을 자동 변경하지 않는다.

## 수용 기준
- AC-1: grammar 시험(정상·오류 각 ≥5)
- AC-2: dual-read fixture(15파일 `@MX:` 기준선) 변경 0
- AC-3: check set exit

## 선행·차단
Blocked-By: #2177

## 사용자 결정
00:5x dual-read — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
