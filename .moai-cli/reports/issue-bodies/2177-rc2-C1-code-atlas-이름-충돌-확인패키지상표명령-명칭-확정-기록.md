# [rc.2][C1] Code Atlas 이름 충돌 확인(패키지·상표·명령) + 명칭 확정 기록

상위: #2155 (EP-C). 설계 근거: 설계 §9 L10 '확정 명칭으로 배포하지 않음'; decisions 00:5x.

## 범위
- crates.io·npm·PyPI·GitHub org·상표 DB(공식 URL)에서 'code-atlas'/'Code Atlas' 충돌 조사, CLI 서브커맨드명 후보 비교
- 결론 문서 → 사용자 확정(리더 중계)

## 요구(GEARS)
- REQ-C1-01: THE REPORT SHALL 조사 출처마다 URL·확인일을 적고 충돌 여부를 판정한다.

## 수용 기준
- AC-1: 출처 ≥4, 전부 URL+날짜
- AC-2: 대안 이름 ≥2
- AC-3: check set exit

## 선행·차단
Blocked-By: 없음

## 사용자 결정
00:5x Code Atlas — docs/decisions-2026-11-08.md

## 증거
PR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip 110b375c7.

## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
