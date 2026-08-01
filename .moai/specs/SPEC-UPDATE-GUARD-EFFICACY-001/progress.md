# SPEC-UPDATE-GUARD-EFFICACY-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase 산출물 3종(spec.md / plan.md / acceptance.md)이 작성되었다. Tier M.

**Plan-phase에서 실제로 실행해 관측한 것** (`.claude/worktrees/e2-data-survival`, PR #1275 브랜치 HEAD `457e1013f`):

- 가드 스코프 커버리지 4종 — `CleanMoaiManagedPaths 66.7%` / `MigrateLegacyMemoryDir 26.9%` / `runCleanReinstall 0.0%` / `BackupMoaiConfig 0.0%`
- update 서브시스템의 비이음매 HOME 호출부 3곳 (`update_clean_install.go` 1건, `update_template_sync.go` 2건) + 범위 밖 `glm.go` 1건
- `snapshotDir` 사용처가 `update_safety_test.go` 한 파일뿐임
- `mergeBackPreserveInventory`의 stat 실패 분기가 `chmod 0o000` + `runtime.GOOS`/`os.Geteuid` 스킵 가드에 의존함
- SPEC ID 정규식 자기 점검: `PASS`

**Plan-phase에서 정정한 선행 기록**: D1의 "`runCleanReinstall`·`BackupMoaiConfig` 0.0%"는 가드 스코프 한정 수치이며, 두 함수 모두 패키지 전체 테스트에서는 이미 구동된다. 진짜 공백은 커버리지가 아니라 사용자 소유 디렉터리 보존 어서션의 도달 범위다 (spec.md §1.3, plan.md §A.1).

**미검증 (plan-phase)**: 리베이스 후 트리의 baseline은 재측정하지 않았다 — plan.md §C2가 run-phase 의무로 지정한다. AC 판정 명령들은 설계만 되었고 실행되지 않았다.

**선행 조건**: PR #1275 머지 후 리베이스 (plan.md §C1). 미충족 시 run-phase 진입 금지.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
