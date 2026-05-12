# Task Decomposition — SPEC-V3R4-CATALOG-002

SPEC: SPEC-V3R4-CATALOG-002 (Wave 2 Distribution — slim init via catalog tier filter)
Total tasks: 28 (M1=6, M2=6, M3=7, M4=4, M5=5)
Development mode: TDD (RED → GREEN → REFACTOR per quality.yaml)
Methodology entry point: M3 audit suite는 TDD RED 단계로 시작

## Wave-Split Plan

Per user choice (HUMAN GATE 1): M1 → M2 → M3 → M4 → M5 sequential delegation. 각 wave 사이 drift guard + LSP gate verification.

## M1 — SlimFS Wrapper API + Implementation (6 tasks)

| Task ID | Description | REQ | Dependencies | Planned Files | Status |
|---------|-------------|-----|--------------|---------------|--------|
| T1.1 | slim_fs.go 생성. Package docstring + `SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` 시그니처. nil 입력 검증. | REQ-001 | — | internal/template/slim_fs.go | pending |
| T1.2 | private `slimFS` struct (underlying fs.FS, denySet map). `fs.FS` + `fs.StatFS` + `fs.ReadDirFS` 3종 구현. immutable invariant. | REQ-001, REQ-003 | T1.1 | internal/template/slim_fs.go | pending |
| T1.3 | `computeDenySet(cat *Catalog)` 헬퍼 — non-core entry Path를 templates/-prefixed namespace로 deny set 추가. | REQ-001, REQ-010 | T1.2 | internal/template/slim_fs.go | pending |
| T1.4 | `(*slimFS) Open(name)` — name은 templates/-prefixed (Anti-pattern: re-prepend 금지). deny set prefix match → fs.ErrNotExist. | REQ-010, REQ-014 | T1.3 | internal/template/slim_fs.go | pending |
| T1.5 | `(*slimFS) ReadDir`, `Stat` — T1.4와 동일 prefix space에서 deny match. ReadDir 결과 필터. | REQ-010, REQ-011 | T1.4 | internal/template/slim_fs.go | pending |
| T1.6 | SlimFS 마지막 단계 — `slimWrapper := &slimFS{...}; return fs.Sub(slimWrapper, "templates")`. caller view에서 templates/ prefix 제거 (REQ-002). | REQ-002 | T1.5 | internal/template/slim_fs.go | pending |

**M1 LOC target**: ~150-200 LOC. Risk R7 (fs.Sub bypass) mitigation: testing/fstest.TestFS calling convention 검증.

## M2 — CLI Integration + Encapsulated Slim Deployer Constructor (6 tasks)

| Task ID | Description | REQ | Dependencies | Planned Files | Status |
|---------|-------------|-----|--------------|---------------|--------|
| T2.1 | init.go에 `--all` Bool flag 추가 (initCmd.Flags().Bool). | REQ-013 | M1 done | internal/cli/init.go | pending |
| T2.2 | `shouldDistributeAll(cmd)` 헬퍼 — `--all` OR `MOAI_DISTRIBUTE_ALL=1` OR case-insensitive "true". 비-canonical 값 (0/yes/empty) 모두 false. | REQ-012, REQ-013 | T2.1 | internal/cli/init.go (또는 init_helpers.go) | pending |
| T2.3 | init.go:293-301 블록 수정 — LoadEmbeddedCatalog → shouldDistributeAll 분기. slim path는 NewSlimDeployerWithRenderer, full path는 기존 EmbeddedTemplates+NewDeployerWithRenderer. slim 진입 시 REQ-021 4-substring notice 출력. | REQ-007, REQ-008, REQ-021-notice | T2.2 | internal/cli/init.go | pending |
| T2.4 | embed_catalog.go 신규 — `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` 두 export. **NO EmbeddedRawForInternal** (DEFECT-5 encapsulation). | REQ-004 | T2.3 | internal/template/embed_catalog.go | pending |
| T2.5 | slim_guard.go 신규 (~5-10 LOC) — `AssertBuilderHarnessAvailable(projectFS) error`. 4 substring (CATALOG_SLIM_HARNESS_MISSING + MOAI_DISTRIBUTE_ALL=1 + `moai init --all` + SPEC-V3R4-CATALOG-005). | REQ-021 | T2.4 | internal/template/slim_guard.go | pending |
| T2.6 | slim_guard_test.go (~25 LOC) — present/missing/nil cases. missing case는 4 substring contains assert (t.Errorf 사용). | REQ-021 | T2.5 | internal/template/slim_guard_test.go | pending |

**M2 LOC target**: ~115-155 LOC (init.go +30~35, embed_catalog.go +35~55, slim_guard +10, slim_guard_test +25). DEFECT-5 encapsulation: `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` zero matches gate.

## M3 — Audit Suite (7 tasks — TDD RED → GREEN entry point)

모든 sentinel emission **t.Errorf 사용 (NOT t.Logf)** — CATALOG-001 eval-1 EC3 hash sentinel 교훈.

| Task ID | Description | REQ | Dependencies | Planned Files | Status |
|---------|-------------|-----|--------------|---------------|--------|
| T3.1 | catalog_slim_audit_test.go 골격 + `loadSlimFS(t)` 헬퍼 (LoadCatalog + SlimFS 일괄). | — | M2 done | internal/template/catalog_slim_audit_test.go | pending |
| T3.2 | TestSlimFS_HidesNonCoreEntries — REQ-014. non-core entry.Path 순회 fs.Stat. sentinel `CATALOG_SLIM_LEAK: <path> tier=<tier>`. | REQ-014 | T3.1 | internal/template/catalog_slim_audit_test.go | pending |
| T3.3 | TestSlimFS_PreservesCoreEntries — REQ-015 + EC4 sub-assertion (`t.Run("nested_moai_workflows_plan", ...)`). sentinel `CATALOG_SLIM_CORE_MISSING: <path>`. | REQ-015, EC4 | T3.2 | internal/template/catalog_slim_audit_test.go | pending |
| T3.4 | TestSlimFS_PreservesNonCatalogFiles — REQ-016. 5 non-catalog path 목록 fs.Stat. sentinel `CATALOG_SLIM_OVER_FILTER: <path>`. | REQ-016 | T3.3 | internal/template/catalog_slim_audit_test.go | pending |
| T3.5 | TestSlimFS_WalkDirNoLeak — REQ-017. fs.WalkDir 전체 walk + deny set cross-check. t.Parallel. sentinel `CATALOG_SLIM_WALK_LEAK: <path>`. | REQ-017 | T3.4 | internal/template/catalog_slim_audit_test.go | pending |
| T3.6 | TestSlimFS_ReadOnlyInvariant — REQ-003 두 부분: (a) reflective check (sync.*/chan/mutable field 탐지) + (b) 32-goroutine race test (`go test -race`). sentinel `CATALOG_SLIM_NOT_READONLY: ...`. | REQ-003 | T3.5 | internal/template/catalog_slim_audit_test.go | pending |
| T3.7 | TestAssertBuilderHarnessAvailable — REQ-021. present/missing/nil 3 cases. missing 시 4 substring assert. | REQ-021 | T2.6 | internal/template/slim_guard_test.go | pending |

**M3 LOC target**: ~220-300 LOC (audit_test.go). T3.6+T3.7는 +40 LOC. Risk R1 (deny set path bug) mitigation: T3.5 fs.WalkDir 전체 cross-check.

## M4 — Backward Compat & Regression (4 tasks)

| Task ID | Description | REQ | Dependencies | Planned Files | Status |
|---------|-------------|-----|--------------|---------------|--------|
| T4.1 | init_test.go + init_coverage_test.go 분석 → 전략 B 채택 (기존 보존 + slim 별도 신규 TestRunInit_SlimDefault). | REQ-018 | M3 done | internal/cli/init_test.go | pending |
| T4.2 | deployer_test.go 가 EmbeddedTemplates 직접 사용 + SlimFS 무관 확인 (grep). `git diff deployer.go` empty 확인. | REQ-005, REQ-018 | T4.1 | (verification only) | pending |
| T4.3 | update_test.go SlimFS 무관 확인 (update.go 미수정 invariant). | REQ-006, REQ-009 | T4.2 | (verification only) | pending |
| T4.4 | `go test -race -count=1 ./...` 전체 회귀 0 확인. coverage delta 측정 (slim_fs.go ≥90%, slim_guard.go ≥90%). | (all) | T4.3 | (verification only) | pending |

**M4 LOC target**: init_test.go +60 LOC (전략 B 신규 sub-test). Risk R4 mitigation: hardcoded file count 회귀 회피.

## M5 — Documentation (5 tasks)

| Task ID | Description | REQ | Dependencies | Planned Files | Status |
|---------|-------------|-----|--------------|---------------|--------|
| T5.1 | CHANGELOG.md `## [Unreleased]` BREAKING CHANGE entry — `moai init` slim default + `--all` + MOAI_DISTRIBUTE_ALL=1 + CATALOG-003/004 mention. | REQ-020 | M4 done | CHANGELOG.md | pending |
| T5.2 | slim_fs.go godoc — D7 lock indirection 정당성 + init.go 호출 패턴. | — | T5.1 | internal/template/slim_fs.go | pending |
| T5.3 | init.go `initCmd.Long` 텍스트에 slim mode 안내 1줄 추가. | — | T5.2 | internal/cli/init.go | pending |
| T5.4 | catalog_doc.md (catalog.yaml:13 reserved.docs_ref 참조) — "Tier filter consumer: SlimFS() + NewSlimDeployerWithRenderer()" cross-reference. CATALOG-001 머지로 파일 존재 확정 (unconditional). | REC-3 | T5.3 | internal/template/catalog_doc.md | pending |
| T5.5 | slim_guard.go godoc — REQ-021 sentinel + 두 안내 substring + 사용 예제 (moai doctor 호출 패턴). | REQ-021 | T5.4 | internal/template/slim_guard.go | pending |

## Acceptance Criteria Coverage

전체 21 REQ → 8 Scenarios (S1-S8) + 6 Edge Cases (EC1-EC6) 매핑 (spec-compact.md §Acceptance Criteria 참조). 각 task 완료 시 해당 AC sub-checklist 갱신.

| AC | Verifies REQ | Owning Task(s) |
|----|-------------|----------------|
| S1 | REQ-001/002/004/007/010/011/021-notice | T1.1-T1.6, T2.3, T2.4 |
| S2 | REQ-012 | T2.2, T2.3 |
| S3 | REQ-013/019 | T2.1, T2.2, T2.3, T5.3 |
| S4 | REQ-010/011/014/017 | T3.2, T3.5 |
| S5 | REQ-015/016 | T3.3, T3.4 |
| S6 | REQ-005/006/018 | T4.2, T4.3 |
| S7 | REQ-009 | T4.3 |
| S8 | REQ-020 | T5.1 |
| EC1 | REQ-008 | T2.3 |
| EC2 | REQ-019 | T2.2 |
| EC3 | REQ-012 | T2.2 |
| EC4 | REQ-010/015 | T3.3 |
| EC5 | REQ-003 | T3.6 |
| EC6 | REQ-021 | T2.5, T3.7 |

## Quality Gates Targets

- `go test -race -count=1 ./internal/template/... ./internal/cli/...` PASS
- slim_fs.go coverage ≥ 90%; slim_guard.go coverage ≥ 90%
- All 8 scenarios + 6 EC verifiable
- CI 14 jobs GREEN (Test×3 OS + Build×5 + Lint + Constitution + Integration×3 + CodeQL)
- D7 lock: `git diff deployer.go update.go` empty
- Encapsulation gate: `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` zero matches
- CHANGELOG BREAKING CHANGE entry present

## P3 Findings (plan-auditor review 1, non-blocking)

| ID | Description | Fix in Phase |
|----|-------------|--------------|
| P3-1 | spec-compact.md L57 "3 substring" vs L127 "4 substring" wording inconsistency (REQ-021) | M3-T3.7 implementation에서 "4" 채택, sync-phase에서 spec-compact.md sync |
| P3-2 | S6 git baseline `0d4bf14ef` staleness (사용자 측정 시점) | M4-T4.2 verification 시 최신 HEAD 기준 적용 |
| P3-3 | EC5(a) denySet immutability godoc-only invariant | M5-T5.2 godoc에 명시 |
| P3-4 | EC5(b) 32-goroutine 통계적 근거 | M3-T3.6 godoc comment에 race 검증 의도 명시 |

## Sync Status (2026-05-12)

All tasks completed via PR #867 (run phase, squash-merged `d15869bb7`). Sync PR adds spec.md status + Implementation Notes only; no new task content.
