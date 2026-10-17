---
id: SPEC-UPDATE-REMOVAL-FAMILY-001
title: "implementation plan"
version: "0.1.0"
created: 2026-10-17
---

# SPEC-UPDATE-REMOVAL-FAMILY-001 — 구현 계획

## A. 컨텍스트

카드 t1668, run tn1p4e, Class C. 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668`, 브랜치 `WT-t1619-spec-b` (base `9b4922d8e`, 선행 t1619 merged-local). 이 계획은 SPEC A 분할 B "제거 가족"의 run 단계를 준비한다. 근거 문서: SPEC A `decision-index.md` F1 (분할 권위), 카드 진행 기록 `.moai/reports/t1668/progress.md`, 본 트리 코드 관측 (§ B).

개발 모드: RED-GREEN-REFACTOR (TDD). 각 마일스톤은 지정 판정 테스트의 RED 를 먼저 기록하고 GREEN 으로 닫는다. 판정 테스트 이름은 분할 기록이 고정한 이름을 그대로 쓴다.

## B. 알려진 문제와 트리 상태 (이번 실행 관측, 2026-10-17)

1. 제거 스위치는 이미 있고 기본 꺼져 있다: `removalEnabled` `internal/cli/migrate_project_assets.go:36-37` (주석이 AC-RC30-02 소유를 card t1668 로 명시), 게이트 분기 `:210-216`. SPEC A plan § B.2 3단계가 inert 로 반입한 것이다. 이 SPEC 은 스위치를 다시 만들지 않는다.
2. 판정 테스트 다섯이 모두 부재하다 (이번 실행 grep, 매치 없음): `TestMigrationRemovalGate`, `TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`, `TestIntegrityProbeCodexOnlyNoClaudeSettings`, `TestMigrationRemovesSharedCopyWithoutProfileReach`.
3. 기존 보호 테스트가 하나 있다: `TestMigrationDefaultKeepsConfirmedProjectCopy` (`internal/cli/migrate_removal_switch_test.go:14`, "no criterion" 명시) 와 픽스처 헬퍼 `enableRemovalForTest` (`:26`). 제거 기계 시험은 이 헬퍼를 쓴다.
4. keep 반쪽 픽스처가 완비돼 있다: `profileReachSandbox` (`internal/cli/migrate_profile_reach_test.go:27` — HOME·원장 샌드박스, `CLAUDE_CONFIG_DIR` 클리어), `mapProjectToWork` (`:41` — `profile.RecordLastUsedProfileForProject`). 제거 반쪽 시험은 이 패턴의 미러다.
5. U-31 코드 수선은 이미 착지했다 (t1578, `ac141b2cb`): `appliesTo` 프레디케이트 `internal/cli/update_integrity_probe.go:55-57`, claudeOnly 항목 `:71`, 스킵 루프 `:154-156`. 기존 시험 둘이 이미 거론 중: `TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness` (`update_integrity_probe_codex_test.go:43`), `TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow` (`:21`).
6. 마이그레이션 제거 경로에는 백업 쓰기가 없다 (`os.Remove(p)` `:217` 직전에 복사 없음). 런 백업의 "Backup" 단계는 인터셉트 형태다 (`internal/cli/update_template_sync.go:444-451` — config 백업·네임스페이스 백업·병합 전 스냅샷 소유). 그 스냅샷이 마이그레이션 제거 대상 프로젝트 자산 (`.claude/skills/**` 등) 을 덮는지는 이 계획이 입증하지 않았다 — M1 의 첫 단계가 측정한다 (Gap G-02).
7. 마이그레이션 호출부는 병합 확인 게이트 뒤에 있다: `internal/cli/update_template_sync.go:1237-1256`. `removedPaths` 는 리포터로 흘러 요약 삭제 집계에 올라간다 (`:1256`).
8. dry-run 미리보기의 제거 후보는 `removalEnabled` 게이트다: `internal/cli/update_dryrun_preview.go:80-84`.

## C. 사전 점검 (Pre-flight)

- [ ] 선행 카드 t1619 merged-local 유지 확인 (`git log --oneline -n 3` — base `9b4922d8e` 계승).
- [ ] `go build ./...` 성공 (변경 전 깨끗한 출발점).
- [ ] `go test ./internal/cli/ -run 'TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationKeepsSharedCopyUnderProfileReach' -count=1` 통과 — 기존 보호 테스트가 녹색인 출발점.
- [ ] 워크트리 가드 제약 확인: 레인 세션에서 인라인 HOME 교체·`CLAUDE_CONFIG_DIR` 실세션 부여는 가드 거부 전례가 있다 (ruling d-20261010T124923Z-2b09 2항). M6 번들 G 가 레인 실행 절차라면 가드 경계 안에서 실행한다; 거부되면 Gap 으로 기록하고 결정 사다리로 처분한다.

## D. 제약

1. 커밋 금지 — 레인이 검토 뒤 커밋한다 (스폰 제약). 파일 작성만.
2. 제거 경로 변화는 REQ-UPM-031 불변식 (삭제가 있었는데 삭제 없음 보고 금지) 과 REQ-UPM-030/032 를 유지해야 한다.
3. dry-run 미리보기 게이트 (`update_dryrun_preview.go:84`) 유지.
4. 판정 테스트 이름은 분할 기록 고정 이름 그대로: `TestMigrationRemovalGate`, `TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`, `TestMigrationRemovesSharedCopyWithoutProfileReach`. 예외: `TestIntegrityProbeCodexOnlyNoClaudeSettings` — 열린 질문 Q3 (이름이 만들어진 적 없고, 행동은 기존 시험 둘이 이미 판정).
5. 프로덕션 기본 동작 변화 없음: `removalEnabled` 기본 false 유지. 모든 새 제거 동작은 스위치가 켜진 테스트 픽스처 안에서만 관측된다.
6. 한 번에 한 파일씩, 의존성 순서로 (Multi-File Decomposition).

## E. 자가 검증

run 단계 종료 시 다음을 관측·기록한다 (progress.md §E.2):

- 다섯 판정 명령의 원문 출력과 종료 코드 (해당 마일스톤 것만; AC-RC30-01 제외 — 운영자 단말).
- `go test -cover` — 변경한 패키지 (기준 85% 이상, `quality.yaml` 게이트).
- `golangci-lint run internal/cli/...`, `go vet`, `gofmt -l` 결과.
- 번들 G 실행 산출물 (M6 실행 시) — 원장 매핑 → 마이그레이션 실행 → 보존·표지 관측의 세 줄 요약.
- 갭 목록의 갱신 상태.

## F. 마일스톤

순서 원칙: 되돌리기 어렵거나 합의가 덜 된 결정 (데이터 흐름 변화, 새 동작) 을 앞에 두고, 기계적·검증 항목을 뒤에 둔다.

### M1 — 제거 전 백업 (U-02, AC-RC30-03) — Priority High

1. 측정 먼저: 현재 런 백업이 마이그레이션 제거 대상 프로젝트 자산을 덮는지 관측하고 결과를 progress.md §E.2 에 기록한다 (Gap G-02 소거 또는 확정).
2. RED: `internal/cli/migrate_removal_backup_test.go` (신규)에 `TestMigrationBackupBeforeRemoval` — 픽스처 제거 전제를 만족시키고 `enableRemovalForTest` 로 스위치를 켠 뒤, 마이그레이션 실행 시 제거된 모든 파일의 사본이 런 백업에 존재하고 마이그레이션이 백업 뒤에 실행됨을 단언.
3. GREEN: `internal/cli/migrate_project_assets.go` — 제거 지점 (`:217`) 직전에 파일 사본을 런 백업에 기록. 백업 위치와 재사용 가능 기계는 `internal/cli/update/backup` (snapshot.go, file_snapshot.go) 재사용을 1순위로 검토하고, 부재 시 최소 추가. `migrateProjectCommonAssets` 시그니처 확장이 필요하면 호출부 `internal/cli/update_template_sync.go:1247` 와 dry-run 미리보기 `:84` 경로를 함께 점검.
4. 순서 보존: 마이그레이션 호출부 (`update_template_sync.go:1237-1256`) 가 "Backup" 인터셉트 (`:449`) 뒤에 있음을 회귀로 고정.

산출물: AC-RC30-03 GREEN. 터치 파일: `internal/cli/migrate_project_assets.go`, `internal/cli/update_template_sync.go` (필요 시), 신규 테스트 파일, `internal/cli/update/backup` (재사용 또는 최소 확장).

### M2 — 빈 디렉터리 정리 (U-10, AC-RC30-11) — Priority Medium

1. RED: `TestEmptiedSkillDirRemoved` (M1 테스트 파일 또는 인접 파일) — 마이그레이션이 스킬 디렉터리의 마지막 파일을 제거하면 그 디렉터리도 제거됨을 단언.
2. GREEN: `internal/cli/migrate_project_assets.go` — 파일 제거 뒤 부모 스킬 디렉터리가 비었으면 제거. 경계 가드는 기존 `withinRootBoundary` (`:388`) 와 심링크 해소 (`:205-208`) 패턴을 따른다. 사용자 생성 내용이 남은 디렉터리는 제거하지 않는다.
3. 매니페스트 뒤처리 (`:236-244` 의 entries-drop 루프) 와 정합성 유지.

산출물: AC-RC30-11 GREEN. 터치 파일: `internal/cli/migrate_project_assets.go`, 테스트 파일.

### M3 — 제거 반쪽 (AC-RC30-27 제거 반쪽) — Priority High

1. RED→GREEN: `internal/cli/migrate_profile_reach_test.go`에 `TestMigrationRemovesSharedCopyWithoutProfileReach` — keep 반쪽의 미러 픽스처 (`profileReachSandbox`, `mapProjectToWork` 없음 = 원장 무매핑, `enableRemovalForTest`) 로, 템플릿 동일 사본 + 대응물 확인 상태에서 마이그레이션이 사본을 제거함을 단언. 백업 선행(M1)과 정리(M2)와의 합성 단언 포함.
2. [NEEDS CLARIFICATION: AC-RC30-27 제거 반쪽의 픽스처 의미 — SPEC A 에 이 반쪽의 행 본문이 존재한 적 없고([RETIRED] 주석만 이름 기록), 위 의미는 keep 반쪽 대칭과 REQ-RC30-003 게이트에서 파생한 것이다. decision-index.md Q1. plan-audit 에서 확인.]

산출물: AC-RC30-27 제거 반쪽 GREEN. 터치 파일: `internal/cli/migrate_profile_reach_test.go`.

### M4 — 제거 게이트 판정 (AC-RC30-02) — Priority High

1. RED→GREEN: `internal/cli/migrate_removal_switch_test.go`에 `TestMigrationRemovalGate` — AC-RC30-01 통과 기록이 없는 상태(기본 구성)에서 제거 전제를 만족하는 사본이 그 자리에 남는지 단언. 기존 `TestMigrationDefaultKeepsConfirmedProjectCopy` ("no criterion") 는 유지하고 판정 테스트가 별도로 붙는다.
2. 게이트 문장 고정: `removalEnabled` 기본 false; 프로덕션 활성화 경로 없음 (§ D 제약 5).

산출물: AC-RC30-02 GREEN. 터치 파일: `internal/cli/migrate_removal_switch_test.go`.

### M5 — 무결성 프로브 판정 (AC-RC30-26, U-31) — Priority Medium

1. 프로덕션 변화 없음 (t1578 수선 착지 완료 — § B 5).
2. [NEEDS CLARIFICATION: 판정 명령 재지정 — SPEC A 원래 행의 `TestIntegrityProbeCodexOnlyNoClaudeSettings` 는 만들어진 적 없고, 행동은 기존 시험 둘이 이미 판정한다. 기본값은 기존 시험 재지정, 대안은 고정 이름 별칭 테스트 추가. decision-index.md Q3.] 기본값 경로의 판정 명령: `go test ./internal/cli/ -run 'TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness|TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow' -count=1 -v`.
3. codex-only 픽스처에서 "missing" 없음 + Claude 픽스처에서 같은 파일 삭제 시 "missing" 유지 — 두 하위 사례 모두 관측.

산출물: AC-RC30-26 판정 기록 (검증 전용, 회귀 가드). 터치 파일: 없음 (기본값 경로) 또는 별칭 테스트 파일.

### M6 — 번들 G (F12): 프로필 도달 레인 검증 — Priority Medium

1. [NEEDS CLARIFICATION: 번들 G 의 형태 — 커밋된 Go 통합 시험 vs 레인 실행 절차. decision-index.md Q2. 기본값은 레인 실행 절차로 progress.md §E.2 에 결과 기록.] 절차 내용 (ruling d-20261010T124923Z-2b09 1항 원문 의미): 샌드박스 프로젝트를 원장(`profile.RecordLastUsedProfileForProject`)으로 프로필에 매핑하고, `CLAUDE_CONFIG_DIR` 를 그 프로필로 둔 채 마이그레이션을 돌려, 공유 사본이 프로젝트 쪽에 남고 `KeptByProfileReach` 표지가 붙는지 확인한다.
2. 프로필 세션의 실제 적재 여부는 범위 밖 (§ G 제외 2번째 항목) — 운영자 확인 (AC-RC30-28, t1642).
3. 워크트리 가드와 충돌하면 (§ C 4번) 실행하지 않고 Gap 으로 기록한다.

산출물: progress.md §E.2 기록 항목. AC 아님. 터치 파일: 없음 (기본값 경로) 또는 통합 시험 파일.

### M7 — AC-RC30-01 운영자 단말 실행 — Priority High

1. 레인 실행 불가 (프로필 세션 필요 — § E 제약 2). 명령은 acceptance.md AC-RC30-01 행에 인용된 것을 그대로 쓴다.
2. [NEEDS CLARIFICATION: AC-RC30-01 미실행 상태에서의 SPEC 종결 처분 — 갭 기록 + 재개 조건으로 종결 허용 vs 실행까지 종결 보류. decision-index.md Q4.] 기본값은 갭 기록 종결 (SPEC A 의 AC-RC30-28 t1642 위탁 선례와 같은 형태).
3. 결과는 운영자가 progress.md §E.2 에 기록한다.

산출물: Gap G-01 기록 + 재개 조건. 터치 파일: 없음.

### M8 — 품질 측정 — Priority Medium (기계적, 마지막)

1. `go test -cover` — 터치한 패키지 전부 (85% 이상).
2. `golangci-lint run`, `go vet ./internal/cli/...`, `gofmt -l internal/cli/`.
3. 결과를 progress.md §E.2 에 기록하고 §E.3 을 작성한다.

## G. 안티 패턴

- 판정 테스트 이름 임의 변경 — 분할 기록 고정 이름 사용 (§ D 제약 4).
- 스위치 기본값 변경이나 프로덕션 제거 경로 활성화 — 금지 (§ D 제약 5).
- RED 없이 GREEN 커밋 — TDD RED 셀이 없는 판정은 릴리스 게이트가 되지 못한다 (SPEC A acceptance.md 게이트 규정 계승).
- 제로 매치 셀렉터를 통과로 보고 — `-run` 이 매치 없는 `[no tests to run]` 은 실패다.
- 백업 위치 재설계 — 기존 `internal/cli/update/backup` 기계 재사용이 1순위 (단순성 사다리).
- `git add -A`·브랜치 상태 변경 — 커밋 자체가 금지 (§ D 제약 1).

## H. 교차 참조

- SPEC A `decision-index.md` F1 — 분할 권위 기록.
- `.moai/reports/t1668/progress.md` — 카드 이력과 제약.
- `internal/cli/update/backup/` — 백업 기계 후보 (M1).
- `internal/cli/update_dryrun_preview.go:80-84` — 미리보기 게이트 (M1 점검 대상).
- decision-index.md — 열린 질문 Q1-Q4.
- acceptance.md — 판정 명령과 갭 목록.
