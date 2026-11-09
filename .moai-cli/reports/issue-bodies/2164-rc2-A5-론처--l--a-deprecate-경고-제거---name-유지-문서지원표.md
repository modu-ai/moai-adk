# [rc.2][A5] 론처 `-l`/`-a` deprecate — 경고 → 제거, `--name` 유지, 문서·지원표·AGENTS.local 갱신

상위: #2153 (EP-A). 설계 근거: `src/main.rs` Cli/HostLaunch; 00-F #2028; #2053.

## 범위
- 1단계(이 카드): `-l`/`-a` 사용 시 stderr 경고('deprecated: use `/moai --mode leader|agent` inside the session') + notes에 기록, 동작은 유지; `--name` 독립 옵션 유지
- 문서: docs/user-entry-draft.md·host-support-matrix.md 론처 표를 '역할 없는 세션 + /moai --mode' 중심으로; AGENTS.local.md § Sessions and roles 갱신
- 2단계(후속 카드 후보): 제거 시점은 EP-A A1~A3 병합 뒤

## 요구(GEARS)
- REQ-A5-01: WHEN `-l` 또는 `-a`가 주어지면 THE LAUNCHER SHALL 경고를 출력하고 기존 동작을 유지한다.
- REQ-A5-02: THE LAUNCHER SHALL 역할 없는 호출의 동작(#2053)을 바꾸지 않는다.

## 수용 기준
- AC-1: `--dry-run` JSON notes에 deprecation 문구 시험 2건
- AC-2: 역할 없는 호출 기존 시험 불변
- AC-3: 문서 3곳 갱신, check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
D-1 폐기 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
