# SPEC-V3R4-CATALOG-002 — Compact

> Auto-extracted from spec.md + acceptance.md (v0.1.1 — plan-auditor iter 1 REVISE→fix). ~30% token savings for `/moai run`.
> Wave 2 Distribution 첫 SPEC. CATALOG-001 manifest 위의 manifest-driven slim init filter.
> Ground truth: catalog.yaml 65 entries = 20 core skills + 20 core agents + 17 optional skills + 7 optional agents + 1 harness-generated agent.

## Scope (1-line)

`moai init` 의 default deploy 를 `tier == "core"` 자산 + non-catalog 템플릿으로 좁힌다. `fs.FS` 레벨 wrapper (SlimFS) 로 D7 lock (deployer.go no-modify) 을 indirection 으로 보존. opt-out 두 경로: `MOAI_DISTRIBUTE_ALL=1` 또는 `moai init --all`.

## Cross-SPEC Boundary

- **vs CATALOG-001**: 본 SPEC 은 manifest *소비자*; manifest 변경 없음.
- **vs CATALOG-003**: 본 SPEC = "core 만 자동 deploy", CATALOG-003 = "`moai pack add` 인터랙티브 install".
- **vs CATALOG-004**: 본 SPEC = init-only filter, CATALOG-004 = update-only drift sync. `update.go` 미수정 invariant 가 경계.
- **vs CATALOG-005**: 본 SPEC = user interaction 변경 0, CATALOG-005 = `/moai project` 인터뷰 + harness 부트스트랩.

## EARS Requirements (21)

### SlimFS Wrapper Contract (Ubiquitous)
- REQ-001: `template.SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` in `internal/template/slim_fs.go`.
- REQ-002: `fs.FS` returned by `SlimFS(...)` (after internal `fs.Sub`) shall NOT include `templates/` prefix — drop-in match for `EmbeddedTemplates()` contract (REC-1 reword).
- REQ-003: SlimFS read-only wrapper: (a) no `sync.*` / chan / mutable field, (b) `go test -race` clean under parallel reads, (c) no goroutines spawned. Sentinel `CATALOG_SLIM_NOT_READONLY` (EC5 verifies).
- REQ-004: Exports `template.LoadEmbeddedCatalog() (*Catalog, error)` + `template.NewSlimDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error)`. **`embeddedRaw` remains unexported** — NO `EmbeddedRawForInternal` (DEFECT-5 encapsulation).

### Non-Modification Invariants (Ubiquitous, D7 lock 보존)
- REQ-005: `internal/template/deployer.go` 미수정.
- REQ-006: `internal/cli/update.go` 미수정.

### Default Slim Init (Event-Driven)
- REQ-007: When `moai init` 실행 + env/flag opt-out 없음 → SlimFS-wrapped FS 가 Deployer 에 전달, tier=core 자산만 deploy.
- REQ-008: When `LoadEmbeddedCatalog` 실패 → init fail with `CATALOG_LOAD_FAILED` 직전, 파일 쓰기 0건.
- REQ-009: When `moai update` 실행 → SlimFS 미사용, unfiltered FS 그대로.

### SlimFS Read Behavior (State-Driven)
- REQ-010: While SlimFS active → 숨겨진 non-core entry path 에 대한 fs.Stat/fs.ReadFile 가 `fs.ErrNotExist`.
- REQ-011: While SlimFS active → fs.WalkDir 가 hidden path 미방문.

### Opt-Out Mechanisms (Optional)
- REQ-012: Where `MOAI_DISTRIBUTE_ALL=1` (or case-insensitive `"true"`) → SlimFS bypass, full deploy.
- REQ-013: Where `moai init --all` → SlimFS bypass, full deploy.

### Audit Invariants (Unwanted Behavior, all sentinels `t.Errorf`)
- REQ-014: If non-core entry path reachable through SlimFS → `TestSlimFS_HidesNonCoreEntries` fail, sentinel `CATALOG_SLIM_LEAK: <path> tier=<tier>`.
- REQ-015: If core entry path NOT reachable → `TestSlimFS_PreservesCoreEntries` fail, sentinel `CATALOG_SLIM_CORE_MISSING: <path>`.
- REQ-016: If non-catalog template path hidden by SlimFS → `TestSlimFS_PreservesNonCatalogFiles` fail, sentinel `CATALOG_SLIM_OVER_FILTER: <path>`.
- REQ-017: If fs.WalkDir visits hidden path → `TestSlimFS_WalkDirNoLeak` fail, sentinel `CATALOG_SLIM_WALK_LEAK: <path>`.

### Backward Compatibility (Unwanted Behavior)
- REQ-018: If `deployer_test.go` runs against full FS post-merge → all GREEN.
- REQ-019: If env + flag 둘 다 set → idempotent, full deploy.

### Documentation (Ubiquitous)
- REQ-020: `CHANGELOG.md` Unreleased 에 BREAKING CHANGE 항목 (S8 verifies).

### Runtime Guard for Missing Builder-Harness (Event-Driven)
- REQ-021: When `builder-harness.md` agent file 부재 → `template.AssertBuilderHarnessAvailable(projectFS) error` 가 `CATALOG_SLIM_HARNESS_MISSING` sentinel + `MOAI_DISTRIBUTE_ALL=1` + `SPEC-V3R4-CATALOG-005` 세 substring 모두 포함하는 wrapped error 반환. Init slim path 도 stdout notice 출력. ~10 LOC in `slim_guard.go` (EC6 verifies).

## Acceptance Criteria (8 G/W/T + 6 Edge Cases)

### S1: Default slim init = core + non-catalog (REQ-001/002/004/007/010/011/021-notice)
- G: catalog.yaml 65 entries (20+20+17+7+1), empty target dir, no env/flag opt-out.
- W: `runInit` → `LoadEmbeddedCatalog()` → `NewSlimDeployerWithRenderer(cat, renderer)` (internally `SlimFS(embeddedRaw, cat)` + `NewDeployerWithRenderer`) → `Deploy()`. **embeddedRaw 외부 노출 0**.
- T: 20 core skills + 20 core agents deployed; 17 optional skills + 7 optional agents + 1 builder-harness 미배포 (25 hidden); non-catalog (rules/output-styles/.moai/CLAUDE.md/.gitignore) verbatim; stdout 4 substrings: `"slim mode"`, `"--all"`, `"MOAI_DISTRIBUTE_ALL=1"`, `"SPEC-V3R4-CATALOG-005"` (REQ-021 notice).

### S2: Env var opt-out = full deploy (REQ-012)
- G: `MOAI_DISTRIBUTE_ALL=1` (or `true`/`True`/`TRUE`).
- W: `shouldDistributeAll()` returns true → SlimFS wrap skip.
- T: 37 skills + 28 agents deploy bit-identical to pre-CATALOG-002. No slim message.

### S3: --all flag opt-out = full deploy (REQ-013/019)
- G: `moai init --all`.
- W: `shouldDistributeAll()` returns true.
- T: Identical to S2. `init --help` text mentions `--all`.

### S4: SlimFS hides every non-core entry (REQ-010/011/014/017)
- G: `LoadCatalog(embeddedRaw)` → `SlimFS()`.
- W: `TestSlimFS_HidesNonCoreEntries` iterates `cat.AllEntries()` non-core entries → `fs.Stat`. `TestSlimFS_WalkDirNoLeak` runs `fs.WalkDir`.
- T: Every non-core `fs.Stat` returns `fs.ErrNotExist`. fs.WalkDir never visits hidden paths. All sentinels use `t.Errorf` (NOT `t.Logf` — CATALOG-001 eval-1 EC3 lesson).

### S5: SlimFS preserves non-catalog files (REQ-015/016)
- G: catalog.yaml enumerates skills+agents only; rules/output-styles/.moai/CLAUDE.md/.gitignore 비-catalog.
- W: `TestSlimFS_PreservesNonCatalogFiles` runs `fs.Stat` for predefined non-catalog list.
- T: Every listed path exists. `TestSlimFS_PreservesCoreEntries` 또한 PASS.

### S6: D7 lock preserved (REQ-005/006/018)
- G: pre-merge HEAD `0d4bf14ef`, post-merge HEAD this SPEC merge.
- W: `git diff` for deployer.go + update.go.
- T: 두 diff 모두 empty. deployer_test.go + update_test.go 모든 케이스 변경 없이 GREEN.

### S7: moai update = full FS unchanged (REQ-009)
- G: slim-initialized project, `moai update` 호출.
- W: `runUpdate` 의 `EmbeddedTemplates()` 호출 경로.
- T: SlimFS 미호출. 65 entries + non-catalog 전체 deploy. (Drift sync 는 CATALOG-004 영역.)

### S8: CHANGELOG BREAKING CHANGE entry (REQ-020)
- G: Post-merge `CHANGELOG.md` Unreleased section.
- W: `grep -E 'BREAKING CHANGE' CHANGELOG.md`, `grep MOAI_DISTRIBUTE_ALL`, `grep -- '--all'`, `grep 'moai init'`.
- T: All grep exit 0; entry mentions BOTH opt-out mechanisms; recommended (not required) refs to CATALOG-002/003/004.

### EC1: catalog.yaml absent/corrupt (REQ-008)
- C: 잘못된 빌드로 catalog.yaml missing, 또는 YAML 깨짐.
- B: `LoadEmbeddedCatalog()` 실패 → `CATALOG_LOAD_FAILED:` prefix error 즉시 반환. 파일 쓰기 0건.
- R: 정상 빌드 재설치.

### EC2: Env + flag 동시 set (REQ-019)
- C: `MOAI_DISTRIBUTE_ALL=1 moai init --all`.
- B: `shouldDistributeAll()` true. SlimFS bypass 1회. 에러/경고 없음.

### EC3: 비-canonical env value (REQ-012)
- C: `MOAI_DISTRIBUTE_ALL=0` / `=yes` / `=` (empty).
- B: `shouldDistributeAll()` false (좁은 매칭 — `"1"` exact + case-insensitive `"true"` 만 인정). Slim 모드.
- R: 정식 값 (`"1"` / `"true"`) 사용.

### EC4: Nested path under core skill (REQ-010/015)
- C: `moai` skill tier=core, sub-files `workflows/plan.md` 등 존재.
- B: SlimFS 가 core 디렉토리 하위 모든 path 통과. `fs.Stat(slim, ".claude/skills/moai/workflows/plan.md")` 성공.
- 검증 (REC-5): `TestSlimFS_PreservesCoreEntries` 가 명시적 sub-assertion 1개 — `t.Run("nested_moai_workflows_plan", ...)` 로 nested path stat 결과 확인. Sentinel: `CATALOG_SLIM_CORE_MISSING: nested .claude/skills/moai/workflows/plan.md`.

### EC5: SlimFS read-only invariant (REQ-003)
- C: `TestSlimFS_ReadOnlyInvariant` 가 (a) reflect 로 slimFS struct field 검사 + (b) 32 goroutine `fs.Stat`/`fs.ReadFile`/`fs.WalkDir` 동시 호출.
- B: (a) `sync.*` / chan / mutable field 발견 시 `t.Errorf("CATALOG_SLIM_NOT_READONLY: field=<name>")`. (b) `go test -race` race report 시 `t.Errorf("CATALOG_SLIM_NOT_READONLY: data race")`.
- 검증: 두 sub-assertion 모두 PASS 필수. denySet 은 construction-time immutable.

### EC6: Builder-harness guard friendly error (REQ-021)
- C: Slim-mode init 후 `template.AssertBuilderHarnessAvailable(projectFS)` 호출.
- B: builder-harness 부재 → returned error 의 `.Error()` 가 `CATALOG_SLIM_HARNESS_MISSING` + `MOAI_DISTRIBUTE_ALL=1` + `moai init --all` + `SPEC-V3R4-CATALOG-005` 4 substring 포함.
- 검증: `TestAssertBuilderHarnessAvailable` 가 `strings.Contains` 로 4 substring 모두 assert. Init slim path 도 stdout notice 출력 (S1 검증).

## Files to Modify / Create

- [NEW] `internal/template/slim_fs.go` (~150-200 LOC) — SlimFS constructor + private slimFS struct (fs.FS + fs.StatFS + fs.ReadDirFS). Wrapper Open/Stat/ReadDir 가 받는 `name` 은 항상 `templates/`-prefixed (T1.4 prefix discipline; double-prefix bug 방지).
- [NEW] `internal/template/slim_fs_test.go` (~200-260 LOC) — unit tests (synthetic + real catalog).
- [NEW] `internal/template/catalog_slim_audit_test.go` (~220-300 LOC) — 6 parallel sub-tests: 4 invariants (LEAK/CORE_MISSING/OVER_FILTER/WALK_LEAK) + 1 read-only (NOT_READONLY) + 1 harness guard (HARNESS_MISSING). All sentinels `t.Errorf`.
- [NEW] `internal/template/embed_catalog.go` (~35-55 LOC) — `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` constructor. **NO `EmbeddedRawForInternal` export** (DEFECT-5 encapsulation).
- [NEW] `internal/template/slim_guard.go` + `slim_guard_test.go` (~35 LOC total) — `AssertBuilderHarnessAvailable(projectFS) error` REQ-021 runtime guard.
- [MODIFY] `internal/cli/init.go` (~30-35 LOC delta) — `--all` flag, `shouldDistributeAll()`, `NewSlimDeployerWithRenderer` wiring (no raw FS visibility), 4-substring slim notice.
- [MODIFY] `CHANGELOG.md` (~12 lines) — BREAKING CHANGE Unreleased entry.
- [MODIFY] `internal/template/catalog_doc.md` (~5 lines, REC-3) — "Tier filter consumer: SlimFS() + NewSlimDeployerWithRenderer()" cross-reference (unconditional, file existence confirmed by catalog.yaml:13 `reserved.docs_ref`).
- [NO MODIFY] `internal/template/deployer.go` — D7 lock 보존.
- [NO MODIFY] `internal/cli/update.go` — CATALOG-004 영역.
- [NO MODIFY] `internal/template/embed.go` — 기존 directive 그대로 사용. `embeddedRaw` 외부 노출 금지.
- [NO MODIFY] `internal/template/catalog.yaml` — manifest 데이터 변경 없음.

## Implementation Notes (for /moai run)

- M1: SlimFS API + impl (6 tasks T1.1-T1.6). T1.4-T1.6 의 `fs.Sub` prefix convention (DEFECT-6 fix): wrapper Open 이 받는 `name` 은 `templates/`-prefixed; deny set 도 동일 namespace; final `fs.Sub(wrapper, "templates")` 가 호출자 view 에서 prefix 제거.
- M2: init.go integration + encapsulated constructor (6 tasks T2.1-T2.6). `NewSlimDeployerWithRenderer(cat, renderer)` 가 raw FS 외부 노출 차단 (DEFECT-5). `slim_guard.go` + REQ-021 notice 포함.
- M3: Audit suite (7 tasks T3.1-T3.7). 모든 sentinel emission `t.Errorf`. T3.6 `TestSlimFS_ReadOnlyInvariant` (REQ-003 DEFECT-2 fix) + T3.7 `TestAssertBuilderHarnessAvailable` (REQ-021 DEFECT-4) 추가.
- M4: Backward compat (4 tasks). init_test.go 전략 B (기존 보존 + slim 별도 신규) 권장.
- M5: Documentation (5 tasks T5.1-T5.5). CHANGELOG BREAKING CHANGE + slim_fs godoc + init --help 텍스트 + catalog_doc.md cross-ref (REC-3, unconditional) + slim_guard.go godoc.

## Risk Highlights

- R3 (downgraded high→low): builder-harness hidden 시 친절 에러 보장. REQ-021 `AssertBuilderHarnessAvailable` (5-10 LOC) + init slim 4-substring notice 로 in-SPEC mitigation. CATALOG-005 부트스트랩까지의 윈도우 동안 사용자 confusion 최소화.
- R7 (high impact): `fs.Sub` 의 wrapped FS interface bypass 가능성. M1-T1.6 의 `fs.Sub(wrapper, "templates")` 호출 convention + wrapper 가 fs.FS + fs.StatFS + fs.ReadDirFS 3종 구현으로 mitigation. `testing/fstest.TestFS` helper.
- R1 (high impact): deny set path 정규화 버그 → leak. M3-T3.5 fs.WalkDir 전체 cross-check 로 mitigation. T1.4 prefix discipline 가 single source of truth.
- R5 (resolved): `EmbeddedRawForInternal` 미도입. 외부 surface = `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()` 두 함수 (DEFECT-5).

## Exclusions (defer to follow-up SPECs)

- Directory relocation (proposal.md 의 원래 CATALOG-002 framing — indefinitely deferred).
- `moai pack add|remove|list|available` CLI → CATALOG-003.
- `moai update --catalog-sync` drift sync → CATALOG-004.
- `/moai project` 인터뷰 + harness 부트스트랩 → CATALOG-005.
- `moai doctor catalog` 진단 → CATALOG-006.
- 4-locale migration docs → CATALOG-007.
- builder-harness 자동 부트스트랩 (CATALOG-005 영역).
- CATALOG-001 deferred nice-to-have 3건 (path containment, pack regex test, BenchmarkLoadCatalog) 미흡수.
- `MOAI_DISTRIBUTE_ALL` 외 추가 env var 미정의.
- pack 자동 권장 (CATALOG-005 영역).
- `moai update` slim 회귀 (CATALOG-004 영역).
- catalog.yaml mutation 없음.

## Quality Gates Targets

- `go test -race -count=1 ./internal/template/... ./internal/cli/...` PASS (race detector clean on TestSlimFS_ReadOnlyInvariant 32-goroutine sub-test).
- `internal/template/slim_fs.go` 커버리지 ≥ 90%; `slim_guard.go` 커버리지 ≥ 90%.
- All 8 scenarios (S1-S8) + 6 edge cases (EC1-EC6) verifiable.
- CI 14 jobs GREEN (REC-4: Test×3 OS + Build×5 + Lint + Constitution + Integration×3 + CodeQL).
- D7 lock: `git diff deployer.go update.go` empty.
- Encapsulation gate: `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` zero matches (DEFECT-5).
- Cross-platform: Linux/macOS/Windows hash + slim 정상.
- CHANGELOG BREAKING CHANGE entry present (S8 G/W/T verifies).
- plan-auditor target: **≥ 0.88** (iter 1: 0.81 → iter 2 projection 0.89-0.92).
