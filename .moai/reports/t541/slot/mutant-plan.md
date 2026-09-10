# t541 뮤턴트 계획 (internal/cli 슬롯에서 실행)

기준 트리: 워크트리 `.claude/worktrees/t541`, 브랜치 `WT-auditor-order-conflict`, 문서 커밋 `542f7dbc2` + 미커밋 테스트 파일 `internal/cli/plan_audit_order_conflict_test.go`.

절차: 사전 sha256 기록 → 초록 1회 → 뮤턴트별(백업 cp → 적용 → 실행 → 백업으로 복원 → `cmp` 동일 확인) → 복원 후 초록 1회. 뮤턴트는 커밋하지 않는다. 복원 확인 전엔 어떤 커밋도 하지 않는다.

선택자: `-run 'TestPlanAuditOrder_'` 에 `-v` 를 붙이고 `--- PASS`/`--- FAIL` 줄 수를 센다(0매치 초록 방지 — 기대 14).

| id | 대상 파일 | 변형 | 빨개져야 할 셀 |
|---|---|---|---|
| m1 | 템플릿 | 하위 제목 수준 검사 제거 (`&& RLENGTH - 1 <= mlevel` 삭제 — 모든 제목이 결속을 끊음) | SubheadingKeepsBinding (+ VerbIdenticalAcrossCopies) |
| m2 | 템플릿 | `after` 방향 CONFLICT 분기 두 줄 삭제 | AfterConflictIsReported (+ VerbIdentical) |
| m3 | 템플릿 | `(Milestone[ \t]+)?` 제거 | MilestoneHeadingFormIsRead (+ VerbIdentical) |
| m4 | 템플릿 | 절 제목 AC 대체(`else if (secac != "") subj = secac`) 제거 | SectionCriterionIsTheSubject (+ VerbIdentical) |
| m5 | 로컬 | MP-9 의 "never forces FAIL by itself" 줄 삭제 | ClausesPresent |
