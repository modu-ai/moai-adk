---
id: SPEC-UPDATE-REMOVAL-FAMILY-001
title: "acceptance criteria"
version: "0.1.0"
created: 2026-10-17
updated: 2026-10-18
---

# SPEC-UPDATE-REMOVAL-FAMILY-001 — 인수 기준

## 예산과 상태

- Tier M 상한: 요구사항 16, 인수 기준 16. 이 파일은 **6** 개의 기준을 가진다: AC-RC30-01, -02, -03, -11, -26, 그리고 AC-RC30-27 의 제거 반쪽. 상한 이내.
- ID 는 SPEC A 에서 그대로 유지된다 (decision-index.md F1 규정). AC-RC30-27 의 keep 반쪽은 SPEC A 소유로 남는다 — 이 SPEC 은 제거 반쪽만 판정한다.
- 요구사항 커버: REQ-RC30-003 → AC-RC30-02, AC-RC30-27(제거 반쪽), 채택 전제로 AC-RC30-01 / REQ-RC30-004 → AC-RC30-03, AC-RC30-11 / REQ-RC30-021 → AC-RC30-26. 커버 못 받은 요구사항 없음.
- **RED 셀 기준**: 판정 테스트가 아직 트리에 없는 기준은 TDD RED 마일스톤에서 판정된다 (해당 마일스톤 = plan.md § F). 그때까지 회귀 가드다. 트리에 존재하는 명령의 기준은 RED 셀이 기록될 때 릴리스 게이트가 된다.
- 제로 매치 셀렉터는 통과가 아니다: `-run` 이 매치 없는 `[no tests to run]` 은 해당 기준의 실패다.

## 인수 기준

### AC-RC30-01 — 프로필 가시성 확인 (U-01 계열; REQ-RC30-001 [REF, SPEC A], REQ-RC30-003)

- **Given** 운영자 단말의 프로필 세션 (`CLAUDE_CONFIG_DIR` 가 프로필 디렉터리를 가리킴) 과, 이 SPEC 트리에서 빌드해 경로로 부른 바이너리가 준비된 상태.
- **When** 프로필 세션에서 가시 스킬 목록 명령을 실행해 `moai-plan` 이 목록에 나오는지 센다.
- **Then** 개수는 1 이상이다. 개수가 0 이면 REQ-RC30-003 은 계속 유효하고, 제거 경로는 머지되지 않는다 (게이트 유지).

- 판정 명령 (D2 수정 — plan-audit-iter1; 프로필 디렉터리 이름은 리더·운영자 입력; 미실행): `CLAUDE_CONFIG_DIR="<profile-dir>" claude -p "List every skill name you can invoke, one per line." | grep -cx 'moai-plan'` — 정확 이름 조건자(`-x`, 행 전체 일치)로 센다. SPEC A 원문의 `grep -c "moai-plan"` 은 부분 문자열을 세어 `Cannot invoke moai-plan` 이나 `moai-plan-archived` 도 1 로 세고(plan-audit-iter1 실측), 파이프라인 종료 상태는 grep 이 매치하면 claude 호출의 실패와 무관하게 0 이다.
- 운영자 기록 의무 (D2): claude 호출 자체의 종료 상태를 개수와 별도로 기록한다 — 두 단계 실행을 권장한다(목록을 파일로 받아 claude 의 종료 상태를 먼저 관측한 뒤 `grep -cx 'moai-plan' < <목록 파일>` 로 센다). claude 가 실패했으면 개수는 무효이며 채택 판정의 근거로 쓰지 않는다.
- 게이트 구분: **채택 게이트** — 제거 경로 프로덕션 활성화의 전제. 회귀 가드로 시작.
- 현재 상태: 레인 워크트리 세션은 프로필 세션을 가질 수 없어 실행 불가 (SPEC A Q8, ruling d-20261010T124923Z-2b09 2항의 가드 거부 전례). Gap G-01. 소유 마일스톤 M7 (운영자 단말).

### AC-RC30-02 — 제거 게이트 (U-01 계열; REQ-RC30-001 [REF, SPEC A], REQ-RC30-003)

- **Given** `removalEnabled` 가 기본값(false)인 트리와, 제거 전제(템플릿 해시와 바이트 일치 + 사용자 측 대응물 확인)를 만족하는 프로젝트 사본.
- **When** `TestMigrationRemovalGate` 를 실행한다.
- **Then** `--- PASS: TestMigrationRemovalGate`. AC-RC30-01 의 통과가 없는 제거는 프로젝트 사본을 그 자리에 남긴다 — 스위치는 꺼진 채고, "removal is not enabled yet" 보존 분기로 간다.

- 판정 명령: `go test ./internal/cli/ -run TestMigrationRemovalGate -count=1 -v`
- 게이트 구분: **TDD RED 마일스톤 (M4) 에서의 릴리스 게이트**. 그때까지 회귀 가드.
- 현재 상태: 스위치는 본 트리에 존재 (internal/cli/migrate_project_assets.go:36-37, 게이트 :210-216 — SPEC A M1 3단계 반입). 판정 테스트 부재 (이번 실행 grep, 매치 없음). 기존 무판정 보호 테스트 `TestMigrationDefaultKeepsConfirmedProjectCopy` 는 유지된다.

### AC-RC30-03 — 제거 전 백업 (U-02; REQ-RC30-004)

- **Given** 테스트 픽스처에서 제거 스위치가 켜진 마이그레이션과, 제거될 파일이 있는 프로젝트.
- **When** 마이그레이션이 파일을 제거한다.
- **Then** `--- PASS: TestMigrationBackupBeforeRemoval`. 제거된 모든 파일은 제거 전에 런 백업에 사본을 가지며, 마이그레이션 단계는 백업 단계 뒤에 실행된다.

- 판정 명령: `go test ./internal/cli/ -run TestMigrationBackupBeforeRemoval -count=1 -v`
- 게이트 구분: **TDD RED 마일스톤 (M1) 에서의 릴리스 게이트**. 그때까지 회귀 가드.
- 현재 상태: 제거 지점 (`internal/cli/migrate_project_assets.go:217`) 경로에 백업 쓰기 없음 (이번 실행 관측). 런 백업 ("Backup" 인터셉트, `internal/cli/update_template_sync.go:444-451`) 이 마이그레이션 제거 대상 자산을 덮는지 미측정 — Gap G-02, M1 첫 단계가 소거.

### AC-RC30-11 — 빈 디렉터리 정리 (U-10; REQ-RC30-004)

- **Given** 마이그레이션이 스킬 디렉터리의 마지막 파일을 제거하는 상태.
- **When** 파일 제거가 완료된다.
- **Then** `--- PASS: TestEmptiedSkillDirRemoved`. 마이그레이션이 비운 스킬 디렉터리는 마지막 파일과 함께 제거된다. 사용자 생성 내용이 남은 디렉터리는 제거되지 않는다.

- 판정 명령: `go test ./internal/cli/ -run TestEmptiedSkillDirRemoved -count=1 -v`
- 게이트 구분: **TDD RED 마일스톤 (M2) 에서의 릴리스 게이트**. 그때까지 회귀 가드.
- 현재 상태: `os.Remove(p)` 는 호출당 파일 하나만 제거; 디렉터리 정리 단계 부재 (이번 실행 grep — 마이그레이션 경로에 prune/RemoveAll 매치 없음).

### AC-RC30-26 — 무결성 프로브 claudeOnly (U-31; REQ-RC30-021)

- **Given** codex-only 프로젝트 (`hideClaude=true` — `.claude/**` 가 배포에서 숨겨져 `.claude/settings.json` 이 없는 상태) 와, 대조 하위 사례로 `.claude/settings.json` 이 실제로 삭제된 Claude 하니스 프로젝트.
- **When** 관리 표면 무결성 프로브가 점검을 실행한다.
- **Then** codex-only 픽스처에서는 `.claude/settings.json` 의 "missing" 발견이 없다. Claude 픽스처에서는 같은 파일의 삭제가 여전히 "missing" 으로 보고된다. 세 번째 조항 "집합의 다른 모든 항목은 계속 점검된다" 는 구조적 근거로 뒷받침된다(D4 수정 — plan-audit-iter1): 프로브 루프(`internal/cli/update_integrity_probe.go:154-156`)는 `appliesTo` 를 통과하지 못하는 항목만 건너뛰고(`:157` 의 점검은 그대로 실행) 나머지 전부를 점검하므로, gpt 하니스에서 집합 전체를 건너뛰는 과잉 건너뛰기가 일어나려면 프로브 소스(루프 또는 집합 리터럴)가 바뀌어야 한다(plan-audit-iter2 R2 정정 — 집합 리터럴 `internal/cli/update_integrity_probe.go:69-72` 의 매니페스트 항목에 `claudeOnly: true` 를 더하는 변이가 루프를 건드리지 않고 같은 과잉 건너뛰기를 일으키며 판정 세트 전체를 통과시킨다, codex 실행 확인). 다만 이 조항의 반증 입력(codex-only 픽스처에서 매니페스트 손상 사례)은 판정 세트에 없으므로 판정 명령이 목격하지 않는다 — 명시적 미목격 사항이다(판정 세트 확장은 Q3 옵션 (나) 영역이며, 확정된 기본값 (가) — 새 테스트 코드 없음 — 을 따르는 한 추가하지 않는다).

- 판정 명령 (decision-index.md Q3 기본값 — 기존 시험 재지정): `go test ./internal/cli/ -run 'TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness|TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow' -count=1 -v`
- 게이트 구분: **회귀 가드 (검증 전용)** — t1578 수선 (`ac141b2cb`) 이 이 카드에 선행해 착지했으므로 SPEC A 원래 행의 RED-우선 계획은 무효다. 프로덕션 변화 없음.
- 현재 상태: 수선 착지 확인 (`internal/cli/update_integrity_probe.go:55-57` `appliesTo`, `:71` claudeOnly 항목, `:154-156` 스킵 루프; 이번 실행 관측). SPEC A 가 지정한 이름 `TestIntegrityProbeCodexOnlyNoClaudeSettings` 의 테스트는 만들어진 적 없음 — 판정 명령 재지정은 Q3 로 확정(기본값 (가), plan-audit-iter1).

### AC-RC30-27 (제거 반쪽) — 프로필 도달 부재 시 제거 (U-01 계열; REQ-RC30-002 [REF, SPEC A], REQ-RC30-003)

- **Given** 런치 원장에 매핑되지 않은 (프로필 도달이 없는) 프로젝트와, 제거 전제를 만족하는 템플릿 관리 사본, 그리고 테스트 픽스처에서 켠 제거 스위치.
- **When** 마이그레이션을 실행한다.
- **Then** `--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach`. 프로필 도달이 없는 프로젝트의 공유 사본은 제거 게이트 (REQ-RC30-003, AC-RC30-02) 뒤에서 백업 선행 순서 (REQ-RC30-004, AC-RC30-03) 를 지켜 제거된다.

- 판정 명령: `go test ./internal/cli/ -run '^TestMigrationRemovesSharedCopyWithoutProfileReach$' -count=1 -v`
- 게이트 구분: **TDD RED 마일스톤 (M3) 에서의 릴리스 게이트**. 그때까지 회귀 가드.
- 유래와 한계: SPEC A 의 acceptance.md 는 이 반쪽의 행 본문을 담은 적이 없다 ([RETIRED] 주석만 `TestMigrationRemovesSharedCopyWithoutProfileReach` 이름과 "moves with AC-RC30-01" 소유를 기록 — SPEC A acceptance.md:239). 위 Given-When-Then 의미는 keep 반쪽(SPEC A acceptance.md:239)의 대칭과 REQ-RC30-003 게이트에서 **파생한 것**으로, Q1 기본값 (가) 가 plan-audit-iter1 에 의해 확정되었다(decision-index.md Q1; plan.md § F M3 해결 항목 — 감사 원문: "Q1 → (가) CONFIRMED. The derived Given-When-Then is sound and invents no scope"). 픽스처는 keep 반쪽 미러: `profileReachSandbox` + 무매핑 + `enableRemovalForTest` (`internal/cli/migrate_profile_reach_test.go`, `internal/cli/migrate_removal_switch_test.go:26`).

## 갭 목록

- **G-01 (AC-RC30-01)**: 프로필 세션 필요 — 레인 실행 불가. 운영자 단말 항목 (M7). 재개 조건: 운영자가 프로필 세션에서 판정 명령을 실행하고 결과를 progress.md §E.2 에 기록. 종결 처분은 decision-index.md Q4.
- **G-02 (AC-RC30-03 선행 측정)**: 런 백업이 마이그레이션 제거 대상 자산을 덮는지 미측정. M1 첫 단계가 소거하거나 구현 범위를 확정한다.
- **G-28 계승 (번들 G)**: `moai doctor` 의 reach 판정은 t1642 (AC-RC30-28) 소유. 번들 G 는 그때까지의 잠정 회귀 가드다 (SPEC A acceptance.md:240). 이 카드의 진행 기록은 G-28 자체를 재선언하지 않는다 — 소유 카드의 갭이다.

## 품질 게이트 (Definition of Done)

- 이 파일의 TDD RED 마일스톤 기준 (AC-RC30-02, -03, -11, AC-RC30-27 제거 반쪽) 이 모두 GREEN 원문 출력과 함께 progress.md §E.2 에 기록됨.
- AC-RC30-26 판정 명령 실행 원문 기록 (Q3 결정에 따른 명령).
- 터치한 패키지 `go test -cover` 85% 이상; `golangci-lint`, `go vet`, `gofmt` 클린.
- REQ-UPM-031 불변식 회귀 확인: 제거 발생 런의 요약이 삭제를 보고함 (기존 테스트 계승 — 내부 삭제 집계 경로 유지).
- 갭 목록의 상태가 최신이고, 빈 Gaps 선언은 사실일 때만.
