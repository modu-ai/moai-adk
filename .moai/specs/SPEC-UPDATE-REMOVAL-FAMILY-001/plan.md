---
id: SPEC-UPDATE-REMOVAL-FAMILY-001
title: "implementation plan"
version: "0.1.0"
created: 2026-10-17
updated: 2026-10-18
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
4. 판정 테스트 이름은 분할 기록 고정 이름 그대로: `TestMigrationRemovalGate`, `TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`, `TestMigrationRemovesSharedCopyWithoutProfileReach`. 예외: `TestIntegrityProbeCodexOnlyNoClaudeSettings` — 해결된 질문 Q3 (기본값 (가), plan-audit-iter1 확정; 이름이 만들어진 적 없고 행동은 기존 시험 둘이 판정 — M5 참조).
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

1. 측정 먼저: 현재 런 백업이 마이그레이션 제거 대상 프로젝트 자산을 덮는지 관측하고 결과를 progress.md §E.2 에 기록한다 (Gap G-02 소거 또는 확정). 같은 기록에 단계 순서 사실을 함께 남긴다(plan-audit-iter1 D3 실측): 이 트리(`76f530fd2`)에서 마이그레이션 호출은 `update_template_sync.go:1247`(`runTemplateSyncWithProgress` 안)에서 실행되고, "Backup" 인터셉트(`:443-451`)를 소유한 `runTemplateSyncWithReporter` 는 그 뒤 `:1256` 에서야 호출된다(나머지 호출부는 `:121` 단독 경로뿐) — 실행 순서는 마이그레이션이 백업보다 앞이다.
2. RED: `internal/cli/migrate_removal_backup_test.go` (신규)에 `TestMigrationBackupBeforeRemoval` — 픽스처 제거 전제를 만족시키고 `enableRemovalForTest` 로 스위치를 켠 뒤, 마이그레이션 실행 시 제거된 모든 파일의 사본이 런 백업에 존재하고 마이그레이션이 백업 뒤에 실행됨을 단언.
3. GREEN: `internal/cli/migrate_project_assets.go` — 제거 지점 (`:217`) 직전에 파일 사본을 런 백업에 기록. 백업 위치와 재사용 가능 기계는 `internal/cli/update/backup` (snapshot.go, file_snapshot.go) 재사용을 1순위로 검토하고, 부재 시 최소 추가. `migrateProjectCommonAssets` 시그니처 확장이 필요하면 호출부 `internal/cli/update_template_sync.go:1247` 와 dry-run 미리보기 `:84` 경로를 함께 점검.
4. 순서 교정 (plan-audit-iter1 D3; 매니페스트 동기화 — plan-audit-iter2 R1): 1단계가 기록한 대로 현 실행 순서는 마이그레이션이 백업보다 앞이고, 이는 REQ-RC30-004("the migration step shall run after the backup step")와 AC-RC30-03 의 Then 과 어긋난다. 이 단계는 기존 순서를 보존하는 것이 아니라 마이그레이션 호출(`update_template_sync.go:1247`)을 Backup 인터셉트(`runTemplateSyncWithReporter` 스테이지 표, `:443-451`) 실행 뒤로 옮기는 재배치를 구현하고 그 순서를 회귀로 고정한다. 재배치에는 매니페스트 인스턴스 동기화가 필수다: 리포터는 자기 매니저를 스테이지 표 앞에서 적재하고(`update_template_sync.go:277-281`) Deploy 단계에서 저장하며(`:667`), 마이그레이션은 별도 매니저 인스턴스를 적재·저장한다(`internal/cli/migrate_project_assets.go:73-74`, Save `:246`). 마이그레이션이 리포터의 적재보다 늦게 실행되면 리포터의 `:667` Save 가 마이그레이션 전의 오래된 보기를 디스크에 덮어써, 드롭된 항목이 되살아나고 `KeptBy` 표지가 지워진다(codex 재현: `KeptBy after reload = "", want "profile-reach"`). 구현은 그래서 (a) 리포터의 매니저 인스턴스를 마이그레이션과 공유하거나, 마이그레이션 직후 리포터의 인스턴스를 다시 적재해야 한다. 그리고 (b) 갱신 전체 흐름 — `runTemplateSyncWithReporter` 경로를 포함 — 이 끝난 뒤 최종 상태 검증을 둔다: 디스크 매니페스트에서 keep 표지(`KeptByProfileReach`)와 드롭된 항목의 부재가 모두 살아남았음을 확인한다. 단위 수준 마이그레이션 시험은 리포터 흐름을 우회하므로 이 덮어쓰기를 목격할 수 없다 — 이 최종 상태 검증이 그 가드다(plan-audit-iter2 residual-risk). 재배치 뒤 필수 점검(잔여 위험): 옮겨진 호출 경로에서 `removedPaths` 가 요약 집계에 그대로 올라가는지 REQ-UPM-030/031/032 불변식(§ E 제약 3)을 다시 확인한다 — plan-audit-iter1 은 재배치된 흐름의 요약 보고를 추적하지 않았다(원문: "REQ-UPM-031/030/032 summary-reporting invariants (§E-3) must be re-checked on the moved call path").

산출물: AC-RC30-03 GREEN. 터치 파일: `internal/cli/migrate_project_assets.go`, `internal/cli/update_template_sync.go` (4단계 순서 재배치 포함), 신규 테스트 파일, `internal/cli/update/backup` (재사용 또는 최소 확장).

### M2 — 빈 디렉터리 정리 (U-10, AC-RC30-11) — Priority Medium

1. RED: `TestEmptiedSkillDirRemoved` (M1 테스트 파일 또는 인접 파일) — 마이그레이션이 스킬 디렉터리의 마지막 파일을 제거하면 그 디렉터리도 제거됨을 단언.
2. GREEN: `internal/cli/migrate_project_assets.go` — 파일 제거 뒤 부모 스킬 디렉터리가 비었으면 제거. 경계 가드는 기존 `withinRootBoundary` (`:388`) 와 심링크 해소 (`:205-208`) 패턴을 따른다. 사용자 생성 내용이 남은 디렉터리는 제거하지 않는다.
3. 매니페스트 뒤처리 (`:236-244` 의 entries-drop 루프) 와 정합성 유지.

산출물: AC-RC30-11 GREEN. 터치 파일: `internal/cli/migrate_project_assets.go`, 테스트 파일.

### M3 — 제거 반쪽 (AC-RC30-27 제거 반쪽) — Priority High

1. RED→GREEN: `internal/cli/migrate_profile_reach_test.go`에 `TestMigrationRemovesSharedCopyWithoutProfileReach` — keep 반쪽의 미러 픽스처 (`profileReachSandbox`, `mapProjectToWork` 없음 = 원장 무매핑, `enableRemovalForTest`) 로, 템플릿 동일 사본 + 대응물 확인 상태에서 마이그레이션이 사본을 제거함을 단언. 백업 선행(M1)과 정리(M2)와의 합성 단언 포함.
2. 해결 (Q1, plan-audit-iter1 확정): decision-index.md Q1 의 기본값 (가) 를 plan-audit-iter1 이 확정했다 — 원문: "Q1 → (가) CONFIRMED. The derived Given-When-Then is sound and invents no scope: the split record itself fixed the test name (`TestMigrationRemovesSharedCopyWithoutProfileReach` — F1:122), and the name asserts removal-without-reach; option (나) (switch off, copy survives) duplicates AC-RC30-02's assertion and leaves the moved criterion unjudged." 따라서 1단계 픽스처의 의미(테스트 전용 스위치 + 원장 무매핑 + 제거 단언, M1 백업·M2 정리 합성 포함)가 이 반쪽의 확정된 판정 문맥이다. 운영자가 Q1 을 뒤집으면 이 항목과 acceptance.md 해당 행은 결정 기록을 따라 다시 쓴다.

산출물: AC-RC30-27 제거 반쪽 GREEN. 터치 파일: `internal/cli/migrate_profile_reach_test.go`.

### M4 — 제거 게이트 판정 (AC-RC30-02) — Priority High

1. RED→GREEN: `internal/cli/migrate_removal_switch_test.go`에 `TestMigrationRemovalGate` — AC-RC30-01 통과 기록이 없는 상태(기본 구성)에서 제거 전제를 만족하는 사본이 그 자리에 남는지 단언. 기존 `TestMigrationDefaultKeepsConfirmedProjectCopy` ("no criterion") 는 유지하고 판정 테스트가 별도로 붙는다.
2. 게이트 문장 고정: `removalEnabled` 기본 false; 프로덕션 활성화 경로 없음 (§ D 제약 5).

산출물: AC-RC30-02 GREEN. 터치 파일: `internal/cli/migrate_removal_switch_test.go`.

### M5 — 무결성 프로브 판정 (AC-RC30-26, U-31) — Priority Medium

1. 프로덕션 변화 없음 (t1578 수선 착지 완료 — § B 5).
2. 해결 (Q3, plan-audit-iter1 확정): decision-index.md Q3 의 기본값 (가) 를 plan-audit-iter1 이 확정했다 — 원문: "Q3 → (가) CONFIRMED, with D4. The pinned name `TestIntegrityProbeCodexOnlyNoClaudeSettings` was never created (grep over internal/ → no match), and the two t1578 tests judge the criterion's two named sub-cases — re-executed green this run with a non-empty sweep (2 tests + 4 subtests)." 기존 시험 재지정으로 확정; 새 테스트 코드는 추가하지 않는다. 세 번째 Then 조항의 미목격 반증 입력은 acceptance.md AC-RC30-26 행의 구조적 근거 각주(D4 수정)로 명시했다. 판정 명령: `go test ./internal/cli/ -run 'TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness|TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow' -count=1 -v`.
3. codex-only 픽스처에서 "missing" 없음 + Claude 픽스처에서 같은 파일 삭제 시 "missing" 유지 — 두 하위 사례 모두 관측.

산출물: AC-RC30-26 판정 기록 (검증 전용, 회귀 가드). 터치 파일: 없음 (기본값 경로) 또는 별칭 테스트 파일.

### M6 — 번들 G (F12): 프로필 도달 레인 검증 — Priority Medium

1. 해결 (Q2, plan-audit-iter1 확정): decision-index.md Q2 의 기본값 (가) 를 plan-audit-iter1 이 확정했다 — 원문: "Q2 → (가) CONFIRMED. Bundle G is correctly NOT an AC (item ledger L56 and M6 both say so; gap G-28 is not re-declared — t1642 owns it). The lane-run procedure matches the ruling's own wording and SPEC A acceptance.md:240's committed sentence; a committed Go test would pre-empt t1642's AC-RC30-28 ownership." 형태는 레인 실행 절차이고 결과는 progress.md §E.2 에 기록한다(1회성 절차에 지속 발화 신호가 없다는 감사 지적은 t1642 착지 전 잠정 역할이라는 명시로 수용). 절차 내용 (ruling d-20261010T124923Z-2b09 1항 원문 의미): 샌드박스 프로젝트를 원장(`profile.RecordLastUsedProfileForProject`)으로 프로필에 매핑하고, `CLAUDE_CONFIG_DIR` 를 그 프로필로 둔 채 마이그레이션을 돌려, 공유 사본이 프로젝트 쪽에 남고 `KeptByProfileReach` 표지가 붙는지 확인한다.
2. 프로필 세션의 실제 적재 여부는 범위 밖 (§ G 제외 2번째 항목) — 운영자 확인 (AC-RC30-28, t1642).
3. 워크트리 가드와 충돌하면 (§ C 4번) 실행하지 않고 Gap 으로 기록한다.

산출물: progress.md §E.2 기록 항목. AC 아님. 터치 파일: 없음 (기본값 경로) 또는 통합 시험 파일.

### M7 — AC-RC30-01 운영자 단말 실행 — Priority High

1. 레인 실행 불가 (프로필 세션 필요 — § E 제약 2). 명령은 acceptance.md AC-RC30-01 행에 인용된 것을 그대로 쓴다.
2. 해결 (Q4, plan-audit-iter1 확정 — D2 조건부): decision-index.md Q4 의 기본값 (가) 를 plan-audit-iter1 이 확정했다 — 원문: "Q4 → (가) CONFIRMED, conditional on D2. The machinery ships gated off (removalEnabled=false; production activation excluded by §E-1/§G-1), so REQ-RC30-003's safety property holds regardless of AC-RC30-01's run state; holding the card on a lane-impossible action serves nothing, and Gap G-01 with its resume condition mirrors the recorded SPEC A AC-RC30-28 deferral precedent. Conditional: the resume command is currently the unsound predicate of D2 — fix D2 before the operator runs it." 종결 처분은 갭 기록 + 재개 조건으로 확정됐다(D2 조건은 이번 수리로 반영 — acceptance.md AC-RC30-01 행 수정본), 재개 명령은 그 수정본을 따른다.
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
- decision-index.md — 질문 Q1-Q4 (모두 기본값 적용 2026-10-17T11:54:29Z + plan-audit-iter1 확정).
- acceptance.md — 판정 명령과 갭 목록.
