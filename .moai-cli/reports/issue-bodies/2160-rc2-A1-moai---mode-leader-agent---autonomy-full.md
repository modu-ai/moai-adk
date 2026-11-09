# [rc.2][A1] `/moai --mode leader|agent [--autonomy full|semi]` skill 인자와 진입 절차

상위: #2153 (EP-A). 설계 근거: 보고서 §10.4; autonomy-modes-spec.md §2.

## 범위
- `skills/moai/SKILL.md`·`references/routes.json`에 `--mode` 인자(leader|agent), `--autonomy`(full|semi, leader만) 정의와 진입 workflow `workflows/mode-leader.md`·`mode-agent.md`
- 역할 없는 세션에서 입력 시: 모드 배너 출력 → 현재 상태(열린 Issue·PR·assignments 스냅샷) 요약 → 모드별 첫 행동(leader: §A4 `/goal` 안내, agent: self-claim 1회)
- Codex는 `$moai --mode agent`만(D-3). `--mode leader`를 Codex에서 입력하면 거절 메시지+근거
- source-manifest.json sha 갱신(skills/moai 편집 규칙)

## 요구(GEARS)
- REQ-A1-01: WHEN 사용자가 `/moai --mode leader` 또는 `--mode agent`를 입력하면 THE SKILL SHALL 해당 모드 workflow를 로드하고 다른 모드 지시를 무시한다.
- REQ-A1-02: WHERE host가 codex이고 `--mode leader`이면 THE SKILL SHALL 진입을 거절하고 D-3 결정을 인용한다.

## 수용 기준
- AC-1: routes.json에 두 route 추가, `tests/` 라우팅 시험 2건(leader/agent) + codex 거절 1건 passed
- AC-2: 두 workflow 문서가 self-claim 7단계·감독 루프를 autonomy-modes-spec.md 절 번호로 참조(중복 서술 0)
- AC-3: source-manifest sha 일치 시험 passed, check set exit 기록

## 선행·차단
Blocked-By: 없음

## 사용자 결정
00:0x·01:0x(11-09) D-1~D-3 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
