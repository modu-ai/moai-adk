# [rc.2][A4] `/goal` 안내 생성 — Issue·PR 분석 → 완료 조건 문장 → 붙여넣기 블록(Claude 내장 /goal·Codex /goal)

상위: #2153 (EP-A). 설계 근거: autonomy-modes-spec.md §7; #2145 지원표 goal 행.

## 범위
- 리더 모드 진입 시: milestone 열린 Issue·PR을 읽어 완료 조건 문장 생성(예: 'milestone X의 Issue·PR 모두 병합·종료, #2016 제외') → 사용자에게 `/goal <문장>` 붙여넣기 블록 출력(모델이 대신 입력 불가 — 확인됨)
- 에이전트 모드: 자기 카드 범위 goal 문장(예: '#<n> PR 병합 또는 blocked 보고') 출력; Codex는 `/goal` 존재 확인됨(learn.chatgpt.com) — 같은 블록 제공
- 제외 규칙: 운영자 보유·keep-set Issue는 조건에서 빼고 사유를 적는다

## 요구(GEARS)
- REQ-A4-01: WHEN 리더 모드가 시작되면 THE SKILL SHALL 현재 열린 Issue·PR로부터 완료 조건 문장을 만들고 붙여넣기 블록으로 제시한다.
- REQ-A4-02: THE SKILL SHALL `/goal`을 대신 실행했다고 주장하지 않는다.

## 수용 기준
- AC-1: fixture 3종(열린 Issue만/PR만/keep-set 포함)에서 생성 문장 snapshot 시험
- AC-2: 문서에 '모델 입력 불가, 붙여넣기' 명시 + 지원표 goal 행 참조
- AC-3: check set exit

## 선행·차단
Blocked-By: #2160, #2145

## 사용자 결정
00:1x /goal 붙여넣기 안내 — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
