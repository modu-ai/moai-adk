# progress.md — SPEC-WEB-CONSOLE-CONFIG-DIET-001

3-phase 라이프사이클(plan → run → sync) 추적. §E가 감사-준비 신호(audit-ready signal)를 담는다.

## §E.1 Plan-phase Audit-Ready Signal

- **작성자**: manager-spec (plan-phase).
- **상태**: draft. plan-phase 산출물 4종 + 본 progress.md 작성 완료.
- **산출물**:
  - `spec.md` — 12-필드 frontmatter(status: draft), §A 개요(+A.4 "dead"의 3결 N1/N2/N3), §B GEARS
    요구사항(REQ-CD-001..050, Tier 1/2/3 + 결함 4 + SHOULD 1), §C Tier 분류(M), §D Out of Scope(4개
    `### Out of Scope —` H3), §E 검증 pointer, §F 교차참조.
  - `plan.md` — §A Context, §B Known Issues(B1 N3 오삭제, B4 team.auto_selection 미확인), §C Pre-flight
    6-command(앵커 재검증), §D Constraints, §E Self-Verification pointer, §F Milestones M0→M1→M2→M3→M4(+M5
    선택) + §F.4 결함 수정 방향 + §F.5 그룹별 결정 매트릭스(은닉/read-only/제거), §G Anti-Patterns AP-1..6,
    §H Cross-References.
  - `acceptance.md` — §A 검증 헬퍼 규약, §B DoD, §C Given-When-Then 3, §D AC 매트릭스(AC-CD-001..091,
    grep/go-test 검증), §D.7 회귀 방지.
  - `research.md` — §R 감사 증거 file:line 인용(clean worktree 97723664c 재-grep), §R.5 미확인 항목 정직
    표기(team.auto_selection "consumed" 미확인), §R.6 요약 카운트.
- **검증 근거(plan-phase, 관측)**: research.md의 모든 사강 주장은 `preview-wc011`에서 grep 재검증됨.
  주요 확인: `internal/research/` 부재, `.Research.`/`.WorkflowAgents` read 빈 출력, `NewTrustGate` CLI
  호출 0, `defaultRetentionDays = 30`(observer.go:147), `LoadRoleProfiles` production 호출 0.
- **SPEC ID self-check**: `SPEC-WEB-CONSOLE-CONFIG-DIET-001` — decomposition PASS
  (SPEC | WEB | CONSOLE | CONFIG | DIET | 001, 마지막 세그먼트 `\d{3}` 순수 숫자).
- **미해결 결정(착수 승인에서 확정)**: §F.5 그룹별 은닉/read-only 선택, §F.4 결함 수정 방향(A/B).
- **주의(정직 표기)**: team.auto_selection "consumed" 주장 미확인 → REQ-CD-050은 소비자 재확인 전제.

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
