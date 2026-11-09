# [rc.2][D7] appendix/interfaces.md OPERATIONS 재동기화 + 자동 검사 — 코드 배열 ↔ 문서 목록 일치 시험

상위: EP-D. 설계 근거: agent-2 사전 확인(tip 49 vs 문서 48/49), agent-1 [0dd806] 측정, #2230 50 보강.

## 범위
- main.rs OPERATIONS 배열과 docs/design-review/appendix/interfaces.md 목록을 비교하는 시험(개수+이름 집합) 추가
- 병합 시점 실제 개수로 문서 재동기화
- 이후 op 추가 PR은 이 시험이 막는다

## 요구(GEARS)
- REQ-D7-01: THE TEST SHALL 코드와 문서의 op 이름 집합이 다르면 실패한다.

## 수용 기준
- AC-1: 시험 추가 + 변이(문서 한 줄 삭제) FAILED 실측
- AC-2: 문서 개수·목록 일치
- AC-3: check set exit

## 선행·차단
Blocked-By: #2221·#2230 병합

## 사용자 결정
ADK_DELEGATED_CHOICES_20261109_01 — 사용자 위임에 따른 영실이 대리 판단(일상 선택). 기존 HOLD 유지.

## 항상 blocked(카드 발행으로 해제되지 않음)
실제 사용자 홈(`~/.claude`, `~/.codex`, `~/.agents`, `~/.moai-cli`) 설정 변경 · 권한/보안 설정 변경 · App 등록/키·토큰/지속 접근 · 유료 호출 · 배포·태그 push · 미승인 host 실세션 실행. 레인 산출물은 코드·문서·격리 fixture(임시 HOME/프로젝트 트리)까지. 미결정(EP-H Q-A/B/C, 프로젝트 skill 위치 P04)은 이 카드가 확정하지 않는다 — 결정 의존 부분은 '결정 대기'로 표기.

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 보고: `.moai-cli/reports/host-setup-scope-2026-11-09.md`(리더) + `host-setup-scope-verification-2026-11-09.md`(agent-1 [0dd806] 독립 재검증). 기준 tip d76a5b7ce.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
- 강제하는 한도·분기마다 변이에 걸리는 시험을 함께 넣는다(#2226 교훈). 오류 문구는 contract 분류표에 등록한다.
