# Implementation Plan — SPEC-V3R4-CATALOG-002

## Goal

Wave 2 Distribution 의 첫 SPEC. CATALOG-001 manifest 위에 **`fs.FS` 레벨 tier 필터 (SlimFS)** 를 도입하여 `moai init` 의 default deploy 를 `tier == "core"` 자산 + 모든 non-catalog 템플릿 파일로 좁힌다. D7 lock (`deployer.go` no-modify) 을 indirection 으로 보존하고, 기존 사용자 영향을 두 가지 opt-out (`MOAI_DISTRIBUTE_ALL=1` env, `moai init --all` flag) 으로 제어한다.

## Approach

본 SPEC 은 **wrapper + minimal integration + audit-driven invariants** 의 3-layer 전략을 따른다:

1. **fs.FS Wrapper (M1)**: `SlimFS(rawFS, cat)` 를 `internal/template/slim_fs.go` 에 정의. 모든 필터 로직을 wrapper 안에 격리. deployer.go / update.go 는 일절 건드리지 않음.
2. **CLI Integration (M2)**: `internal/cli/init.go` 의 `EmbeddedTemplates()` 호출 직후에 SlimFS 적용. opt-out 두 경로 (env, flag) 의 OR 로직.
3. **Audit-driven Invariants (M3)**: leak / over-filter / walk leak / core missing 4개 invariant 를 audit suite 로 강제. 모든 sentinel 은 `t.Errorf` (CATALOG-001 evaluator-active iter 1 교훈).
4. **Backward Compat (M4)**: 기존 deployer_test.go / update_test.go / init_test.go 회귀 검증. init_test.go 의 expected deployed file count 만 slim mode 기준으로 조정.
5. **Documentation (M5)**: CHANGELOG.md 의 Unreleased 섹션에 BREAKING CHANGE 항목 + slim_fs.go godoc + init --help 텍스트.

본 SPEC 의 변경은 `moai update` 동작 영향이 0 (update 는 full FS 유지), 기존 프로젝트의 `.claude/` 영향이 0 (이전에 deploy 된 자산은 보존), 그리고 신규 init 의 deploy surface 만 좁힌다.

## Task Decomposition

전체 task ~25개. M1 (5) + M2 (6) + M3 (5) + M4 (4) + M5 (5). development_mode 는 quality.yaml 기준 TDD 가정 — M3 audit suite 가 TDD RED → GREEN 진입점.

### M1 — SlimFS Wrapper API + Implementation

- T1.1: `internal/template/slim_fs.go` 생성. Package docstring + `SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` 시그니처 정의. 입력 검증 (nil rawFS / nil cat → error).
- T1.2: private `slimFS` struct 정의 (필드: `underlying fs.FS`, `denySet map[string]struct{}`). `fs.FS` interface (`Open(name string) (fs.File, error)`) 구현. `fs.StatFS` 와 `fs.ReadDirFS` interface 추가 구현 (audit + WalkDir 효율). 필드 전부 unexported + immutable (no `sync.*`, no chan, no mutable map mutation after construction) — REQ-003 invariant.
- T1.3: `computeDenySet(cat *Catalog) map[string]struct{}` 헬퍼 — `cat.AllEntries()` 순회하여 `tier != TierCore` 인 entry 의 `Path` 를 deny set 에 추가. Deny set 엔트리 형식은 catalog.yaml 의 `path` 필드 원본 그대로 — 즉 `templates/` prefix **유지** (예: `templates/.claude/skills/moai-domain-backend/`, `templates/.claude/agents/moai/expert-mobile.md`, `templates/.claude/agents/moai/builder-harness.md`). 디렉토리 표기는 trailing slash 유지, 단일 파일은 그대로. **이 매칭 namespace 는 wrapper Open() 이 받는 `name` 과 동일한 prefix space** (T1.4 참조).
- T1.4: `(*slimFS) Open(name string) (fs.File, error)` — `slimFS` 는 raw `embeddedRaw` 를 underlying 으로 받으므로 `Open` 이 받는 `name` 은 항상 `templates/` prefix 가 포함된 path (예: `"templates/.claude/skills/foo/spec.md"`). 이 `name` 을 **추가 prefix 결합 없이 그대로 deny set 의 각 entry 와 prefix 매칭** 한다. 매칭 시 `&fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}` 반환. 매칭 없으면 `underlying.Open(name)` 으로 pass-through. (Anti-pattern: `name` 에 `"templates/"` 를 prepend 해서는 안 됨 — double-prefix bug. ground truth: `slimFS.Open` 은 raw FS layer 에서 동작.)
- T1.5: `(*slimFS) ReadDir(name string) ([]fs.DirEntry, error)` 및 `(*slimFS) Stat(name string) (fs.FileInfo, error)` — T1.4 와 동일한 prefix space (`templates/`-prefixed) 에서 deny set prefix 매칭. ReadDir 결과는 underlying 호출 후 deny prefix 와 매칭되는 child entry 제거 (sub-directory 진입 차단). Stat 은 deny match 시 즉시 `fs.ErrNotExist`.
- T1.6: `SlimFS()` 마지막 단계 — `slimWrapper := &slimFS{underlying: rawFS, denySet: computeDenySet(cat)}` 구성 후 `return fs.Sub(slimWrapper, "templates")`. 호출자 (init.go) 입장에서는 `templates/` prefix 가 제거된 view 만 보인다 (REQ-002). 단, internal 의 wrapper Open/Stat/ReadDir 는 여전히 `templates/`-prefixed name 을 받는다 (`fs.Sub` 가 prefix 를 자동으로 prepend 하여 wrapper 에 전달하기 때문). 이 invariant 가 deny set namespace 와 wrapper 입력 namespace 의 일치를 보장.

**Risk in M1**: deny set 키 정규화 (slash 처리, 디렉토리 vs 파일 구분) 가 오타나면 leak 발생. 해결: T1.5 의 prefix matching 을 sentinel 테스트 (M3-T3.1) 가 잡는다. 또한 deny set 빌드 단계에서 catalog 의 entry.Path 가 `templates/.claude/skills/<name>/` 또는 `templates/.claude/agents/moai/<name>.md` 형태임을 단위 테스트로 확정 (slim_fs_test.go). **추가 risk**: `fs.Sub` 가 wrapper Open 에 전달하는 name 이 `templates/`-prefixed 라는 가정이 잘못되면 T1.4 의 prefix matching 이 fail — `testing/fstest.TestFS` helper 로 calling convention 검증 (R7 mitigation).

### M2 — CLI Integration (init.go) + Encapsulated Slim Deployer Constructor

[ENCAPSULATION] 본 SPEC 은 `embeddedRaw` 를 `internal/template/` 패키지 외부로 노출하지 않는다. 외부 호출 surface 는 `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` 두 함수로 한정 (REQ-004).

- T2.1: `internal/cli/init.go:60` 의 flag 정의 블록에 `initCmd.Flags().Bool("all", false, "Deploy all catalog entries (core + optional packs + harness-generated). Bypasses slim mode (CATALOG-002).")` 추가.
- T2.2: `shouldDistributeAll(cmd *cobra.Command) bool` 헬퍼 함수 추가 (init.go 또는 init_helpers.go 신규 파일). 로직: `cmd.Flags().GetBool("all")` OR `os.Getenv("MOAI_DISTRIBUTE_ALL") == "1"` OR `strings.EqualFold(os.Getenv("MOAI_DISTRIBUTE_ALL"), "true")`.
- T2.3: `internal/cli/init.go:293-301` 블록 수정 — 기존 `embeddedFS, err := template.EmbeddedTemplates()` + `template.NewDeployerWithRenderer(embeddedFS, renderer)` 흐름을 다음과 같이 분기:
  ```go
  cat, catErr := template.LoadEmbeddedCatalog()
  if catErr != nil {
      return fmt.Errorf("CATALOG_LOAD_FAILED: %w", catErr)
  }
  var deployer template.Deployer
  if shouldDistributeAll(cmd) {
      embeddedFS, err := template.EmbeddedTemplates()
      if err != nil { return fmt.Errorf("embedded templates: %w", err) }
      deployer = template.NewDeployerWithRenderer(embeddedFS, renderer)
  } else {
      var slimErr error
      deployer, slimErr = template.NewSlimDeployerWithRenderer(cat, renderer)
      if slimErr != nil {
          return fmt.Errorf("CATALOG_LOAD_FAILED: slim deployer: %w", slimErr)
      }
      // REQ-021 informational notice on slim mode
      fmt.Fprintln(os.Stdout, "Deploying core templates only (slim mode). Use --all or MOAI_DISTRIBUTE_ALL=1 for full deploy. Note: builder-harness agent is omitted (see SPEC-V3R4-CATALOG-005 for bootstrap).")
  }
  // ... existing deployer.Deploy() call unchanged
  ```
  **raw FS 가 외부 코드에서 직접 보이지 않음** — `NewSlimDeployerWithRenderer` 내부에서만 `embeddedRaw` 가 소비된다.
- T2.4: `internal/template/embed_catalog.go` 신규 파일. Exports:
  - `LoadEmbeddedCatalog() (*Catalog, error)` — internally `LoadCatalog(embeddedRaw)`.
  - `NewSlimDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error)` — internally:
    ```go
    if cat == nil { return nil, fmt.Errorf("nil catalog") }
    slim, err := SlimFS(embeddedRaw, cat)
    if err != nil { return nil, fmt.Errorf("slim fs: %w", err) }
    return NewDeployerWithRenderer(slim, renderer), nil
    ```
  - **NO** `EmbeddedRawForInternal` / `EmbeddedRawForTest` exports. `embeddedRaw` 패키지 변수는 unexported 유지 (D7 lock 정신의 encapsulation 보호). 테스트는 `LoadEmbeddedCatalog()` 또는 `testing/fstest.MapFS` 합성 catalog 로 충당.
- T2.5: `internal/template/slim_guard.go` 신규 파일 (~5-10 LOC) — REQ-021 가드:
  ```go
  // AssertBuilderHarnessAvailable returns a CATALOG_SLIM_HARNESS_MISSING-tagged
  // error when the builder-harness agent is absent from the given project FS.
  func AssertBuilderHarnessAvailable(projectFS fs.FS) error {
      _, err := fs.Stat(projectFS, ".claude/agents/moai/builder-harness.md")
      if err == nil { return nil }
      return fmt.Errorf(
          "CATALOG_SLIM_HARNESS_MISSING: builder-harness omitted in slim mode. "+
              "Run `moai init --all` or set MOAI_DISTRIBUTE_ALL=1, or wait for SPEC-V3R4-CATALOG-005 auto-bootstrap. (underlying: %w)", err)
  }
  ```
  Companion `slim_guard_test.go` — present case (PASS), missing case (sentinel substring assert), nil FS (defensive return). ~25 LOC.
- T2.6: `internal/cli/init.go` 의 slim 분기 마지막 (deploy 성공 후) 에서 `template.AssertBuilderHarnessAvailable` 호출은 NOT 필수 (이미 deploy 결과로 absent 확정). 단, `moai doctor` 후속 사용을 위해 helper export 가 충분.

**Risk in M2 (resolved)**: 이전 안 (EmbeddedRawForInternal export) 의 캡슐화 위반 risk → `NewSlimDeployerWithRenderer` constructor 패턴으로 해결. R5 (former Risks table 항목) 제거. plan-auditor REC: encapsulation invariant 가 spec.md §"Overview" `[INVARIANT]` block 으로 명문화됨.

### M3 — Audit Suite (TDD RED → GREEN)

모든 sentinel emission 은 **`t.Errorf` 사용 (NOT `t.Logf`)** — CATALOG-001 evaluator-active iter 1 의 EC3 hash sentinel 교훈을 처음부터 적용. 본 M3 는 7 sub-tests (T3.1 helper + T3.2~T3.5 audit invariants + T3.6 read-only invariant + T3.7 harness guard).

- T3.1: `internal/template/catalog_slim_audit_test.go` 골격 작성 — `package template` + imports (`io/fs`, `strings`, `testing`, `testing/fstest`). Helper `loadSlimFS(t *testing.T) (fs.FS, *Catalog)` — `LoadCatalog(embeddedRaw)` + `SlimFS()` 일괄 호출.
- T3.2: `TestSlimFS_HidesNonCoreEntries` — REQ-014 강제. `cat.AllEntries()` 순회하여 `tier != TierCore` 인 entry 의 `Path` 에 대해 `fs.Stat(slim, strings.TrimPrefix(entry.Path, "templates/"))` 호출. err 없거나 `errors.Is(err, fs.ErrNotExist)` 가 false 면 fail. Sentinel: `CATALOG_SLIM_LEAK: <path> tier=<tier>`.
- T3.3: `TestSlimFS_PreservesCoreEntries` — REQ-015 강제. tier == TierCore 인 entry 의 path 에 대해 `fs.Stat` 호출. err 가 있거나 not found 면 fail. Sentinel: `CATALOG_SLIM_CORE_MISSING: <path>`. **추가 EC4 sub-assertion (REC-5)**: 본 테스트는 `cat.AllEntries()` core entry 순회 끝에 nested path 검증 1건 — 예: `t.Run("nested_moai_workflows_plan", func(t *testing.T) { _, err := fs.Stat(slim, ".claude/skills/moai/workflows/plan.md"); if err != nil { t.Errorf("CATALOG_SLIM_CORE_MISSING: nested .claude/skills/moai/workflows/plan.md: %v", err) } })` 를 포함. (catalog entry granularity 가 top-level 이지만 sub-files 도 reachable 임을 명시적으로 보증; spec.md §Open Questions OQ4 + acceptance.md EC4 와 일치.)
- T3.4: `TestSlimFS_PreservesNonCatalogFiles` — REQ-016 강제. 미리 정의한 non-catalog path 목록 (`.claude/rules/moai/core/zone-registry.md`, `.claude/output-styles/`, `.moai/config/sections/quality.yaml`, `CLAUDE.md`, `.gitignore`) 각각에 대해 `fs.Stat`. 누락되면 fail. Sentinel: `CATALOG_SLIM_OVER_FILTER: <path>`.
- T3.5: `TestSlimFS_WalkDirNoLeak` — REQ-017 강제. `fs.WalkDir(slim, ".", walkFn)` 실행. walk 가 방문하는 모든 path 를 수집한 뒤, 그 중 어느 하나라도 deny set 의 prefix 와 매칭되면 fail. Sentinel: `CATALOG_SLIM_WALK_LEAK: <path>`. 추가로 `t.Parallel()` 사용 (lang_boundary_audit_test.go 선례).
- T3.6: `TestSlimFS_ReadOnlyInvariant` — REQ-003 강제. 두 부분으로 구성:
  - **(a) Reflective check**: `reflect.TypeOf(slimWrapper).Elem()` 의 모든 field 를 순회하여 `sync.*` 타입, chan 타입, 또는 mutable map (denySet 외) 발견 시 fail. Sentinel: `CATALOG_SLIM_NOT_READONLY: field=<name> kind=<kind>`. denySet 은 construction 후 immutable 임을 godoc 으로 명시 + 테스트가 lookup-only 호출만 사용함을 검증.
  - **(b) Race-detector check**: 32 goroutine 가 동시에 `fs.Stat`, `fs.ReadFile`, `fs.WalkDir` 를 random core/non-core path 에 대해 호출. `go test -race -run TestSlimFS_ReadOnlyInvariant` 가 race report 없이 PASS 해야 함. race detection 시 sentinel: `CATALOG_SLIM_NOT_READONLY: data race during concurrent reads`.
- T3.7: `TestAssertBuilderHarnessAvailable` (in `slim_guard_test.go`) — REQ-021 강제. 세 케이스:
  - present: `.claude/agents/moai/builder-harness.md` 가 있는 synthetic FS → `AssertBuilderHarnessAvailable` returns nil.
  - missing: builder-harness 부재 FS → returned error 가 substring `CATALOG_SLIM_HARNESS_MISSING`, `MOAI_DISTRIBUTE_ALL=1`, `SPEC-V3R4-CATALOG-005` 모두 포함.
  - nil FS: defensive `nil err`. (또는 panic 방지 — implementer 재량.)
  All sentinel emissions: `t.Errorf`.

**Risk in M3**: harness-generated tier 의 `builder-harness` agent 가 hidden 인데, 만약 미래의 manifest 가 `builder-harness` 를 `tier: core` 로 옮기면 T3.3 가 fail 한다 — 의도된 회귀 (manifest 변경은 SPEC 으로 추적). plan-auditor 가 manifest 변경의 SPEC 참조를 강제하므로 (REQ-CATALOG-001-013), 본 SPEC 단독으로는 안전. **REQ-003 race-detector test 의 비용**: 32 goroutine × 약간의 random path = ms 단위, CI 회귀 부담 무시 가능.

### M4 — Backward Compat & Regression

- T4.1: `internal/cli/init_test.go` + `init_coverage_test.go` 분석 → 새 default (slim) 가 깨뜨릴 가능성 있는 expected deployed file count 항목 식별. **두 가지 전략 중 선택**:
  - 전략 A: 기존 테스트의 expected count 를 slim mode 기준 (core only) 으로 업데이트.
  - 전략 B: 기존 테스트는 `--all` flag 또는 `t.Setenv("MOAI_DISTRIBUTE_ALL", "1")` 로 full deploy 보장 후 그대로 유지하고, 별도 새 테스트 `TestRunInit_SlimDefault` 를 추가하여 slim 경로 검증.
  - **권장**: 전략 B (기존 테스트 보존 + slim 별도). 회귀 위험 최소.
- T4.2: `internal/template/deployer_test.go` 의 모든 케이스가 `EmbeddedTemplates()` (full FS) 를 직접 사용함을 grep 으로 확인. SlimFS 미사용 → 변경 영향 0 (read-only verification). `git diff internal/template/deployer.go` 가 empty 임도 확인.
- T4.3: `internal/cli/update_test.go` 의 모든 케이스가 SlimFS 무관임을 확인 (update.go 는 본 SPEC 미수정). full test suite 회귀 0.
- T4.4: `go test -race -count=1 ./...` 전체 suite 회귀 0 확인. coverage delta 측정 (`internal/template` 의 LoadCatalog 100% + 신규 SlimFS 90%+ 목표).

**Risk in M4**: init_test.go 가 deployed file count 를 hardcode 한 케이스가 있다면 T4.1 의 전략 B 로 대응. 전략 A 채택 시 expected count 는 catalog.yaml 의 core 자산 수 (20 skills + 20 agents) + non-catalog 템플릿 파일 수 (rules / output-styles / .moai / 루트 파일들) 의 합으로 계산하되, 정확한 수치는 implementation 시 측정.

### M5 — Documentation

- T5.1: `CHANGELOG.md` 의 `## [Unreleased]` 섹션 (또는 신규 생성) 에 다음 항목 추가:
  ```markdown
  ### BREAKING CHANGE
  - `moai init` now deploys core-tier catalog entries only by default (SPEC-V3R4-CATALOG-002, ~50% reduction in skills/agents deployed).
    - Restore previous behavior with `--all` flag: `moai init --all`
    - Or set `MOAI_DISTRIBUTE_ALL=1` environment variable
    - Optional packs (backend/frontend/mobile/etc.) will be installable via `moai pack add <name>` once SPEC-V3R4-CATALOG-003 is released.
    - Existing projects: `moai update` is unchanged in this release. Drift-aware sync arrives in SPEC-V3R4-CATALOG-004.
  ```
- T5.2: `internal/template/slim_fs.go` 의 godoc 작성. SlimFS 가 D7 lock 을 indirection 으로 보존하는 정당성 명시. 사용 예제 (init.go 에서의 호출 패턴).
- T5.3: `internal/cli/init.go` 의 `initCmd.Long` 텍스트에 slim mode 안내 1줄 추가. 예: `"By default, only core templates are deployed (~50% lighter). Use --all or set MOAI_DISTRIBUTE_ALL=1 for full deploy."`
- T5.4: `internal/template/catalog_doc.md` (CATALOG-001 산출물 — `internal/template/catalog.yaml:13` 의 `reserved.docs_ref` 필드가 참조) 에 SlimFS 관련 1개 paragraph 추가 — "Tier filter consumer: `SlimFS()` + `NewSlimDeployerWithRenderer()` 가 `tier == TierCore` filter 를 어떻게 적용하는지" cross-reference. CATALOG-001 머지 시 파일 존재가 확정되므로 unconditional 작업.
- T5.5: `internal/template/slim_guard.go` 의 `AssertBuilderHarnessAvailable` godoc — REQ-021 의 sentinel + 두 가지 안내 substring (MOAI_DISTRIBUTE_ALL=1, SPEC-V3R4-CATALOG-005) 명시. 사용 예제 (moai doctor 가 호출하는 패턴).

## Risks & Mitigations

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | SlimFS deny set 의 path normalization 버그 (trailing slash, leading slash, OS path separator) 로 leak 발생 | 중 | 높음 | M3-T3.5 의 `TestSlimFS_WalkDirNoLeak` 가 fs.WalkDir 전체 walk 결과를 deny set 과 cross-check. unit test (T1.5 의 slim_fs_test.go) 가 동일 invariant 를 synthetic catalog 로도 검증. |
| R2 | `moai init` 의 기존 사용자가 slim 모드를 인지하지 못해 누락된 자산을 발견 후 confusion | 중 | 중 | M5-T5.1 CHANGELOG BREAKING CHANGE 명시 + M2-T2.4 stdout 1-line 안내 + M5-T5.3 init --help 텍스트. CATALOG-003 (`moai pack add`) 머지 전 사용자에게 "현재는 pack opt-in 명령이 없으므로 `--all` 사용 권장" 안내 가능. |
| R3 | harness-generated tier 의 builder-harness 가 hidden 되어 harness workflow 가 즉시 fail | 중 | 낮 (downgraded) | REQ-021 + `AssertBuilderHarnessAvailable` (5-10 LOC) 런타임 가드 + init.go slim path stdout notice. 두 substring (`MOAI_DISTRIBUTE_ALL=1`, `SPEC-V3R4-CATALOG-005`) 가 모든 error message 에 포함되므로 사용자 confusion 최소. CATALOG-005 의 부트스트랩이 머지되면 가드가 silent 로 PASS. (이전 평가: 높음 — but mitigated by in-SPEC runtime guard, not deferred.) |
| R4 | init_test.go 가 deployed file count 를 hardcode 하여 회귀 발생 | 중 | 낮 | M4-T4.1 전략 B (기존 테스트 보존 + slim 별도) 채택. 추가 작업 부담 최소. |
| R5 | ~~`EmbeddedRawForInternal` 가 외부 패키지에서 오용되어 D7 indirection 의 보호가 우회됨~~ — **RESOLVED**: `EmbeddedRawForInternal` 미도입. 외부 surface 는 `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()` 두 함수로 한정 (DEFECT-5 fix). 캡슐화 invariant 가 spec.md `[INVARIANT]` block 으로 명문화. |
| R6 | SlimFS 성능 회귀 — Open / ReadDir 호출마다 deny set lookup | 낮 | 낮 | deny set 은 map[string]struct{} 로 O(1) lookup. 65 entries 규모에서 무시 가능. M4-T4.4 의 `go test -race` 가 race condition 없음을 확인. |
| R7 | `fs.Sub` 의 sub-FS 가 wrapped FS 의 Open/Stat/ReadDir 를 호출하지 않을 가능성 (Go stdlib 동작 불확실) | 중 | 높음 | M1-T1.5 가 wrapper 의 인터페이스 구현을 `fs.FS` + `fs.StatFS` + `fs.ReadDirFS` 3종 모두 보장. `testing/fstest` 의 `TestFS` helper 로 calling convention 검증. 만약 `fs.Sub` 가 일부 인터페이스를 bypass 하면 wrapper 안에서 `templates/` prefix 매칭을 직접 수행하는 두 번째 전략으로 fallback. |
| R8 | catalog.yaml 의 entry path 가 trailing slash 유무 불일치 (디렉토리 vs 파일) | 낮 | 중 | T1.3 의 `computeDenySet` 가 정규화 로직 (trailing slash 제거, leading "templates/" 보장) 을 포함. unit test (slim_fs_test.go) 가 양쪽 케이스 모두 커버. |
| R9 | `MOAI_DISTRIBUTE_ALL` 값이 빈 문자열 또는 명시되지 않은 경우 slim 모드 진입 — 사용자 의도와 다를 가능성 | 낮 | 낮 | M2-T2.2 의 매칭 규칙을 "값이 `\"1\"` 이거나 case-insensitive `\"true\"`" 로 좁힘. 기타 값 (예: `\"0\"`, `\"\"`) 은 slim 모드 유지. README 문서에 명시. |

## Open Decisions (Pre-implementation Confirm)

본 SPEC 의 spec.md §"Open Questions" 5건은 plan-auditor / 사용자 검토에서 lock-in 한다:

1. **OQ1 (env var matching rule)**: `"1"` exact + `"true"` case-insensitive 권장.
2. **OQ2 (flag 명칭)**: `--all` 권장.
3. **OQ3 (slim 모드 안내 출력)**: 1-line stdout 안내 권장.
4. **OQ4 (builder-harness 부트스트랩)**: 본 SPEC 비도입, CATALOG-005 위임.
5. **OQ5 (fs.ReadDirFS 구현)**: 구현 권장 (WalkDir 효율).

implementation 시작 (M1) 전에 사용자가 위 결정에 confirm.

## MX Tag Plan

본 SPEC 의 산출물 중 fan-in 이 높은 신규 함수:

- `template.SlimFS()` — `@MX:ANCHOR` 부착. fan_in 예상 ≥ 2 (init.go + 미래 CATALOG-003/004). 단, 본 SPEC 머지 시점에서는 fan_in = 1 (init.go 만). CATALOG-001 의 LoadCatalog ANCHOR 와 동일 정책: 미래 fan_in 을 근거로 ANCHOR 미리 부착 (CATALOG-001 progress.md §"Phase 2.9 MX Tag Update").
- `template.LoadEmbeddedCatalog()` — `@MX:NOTE` 부착. wrapper convenience function 의 intent 명시.
- `internal/cli/init.go` 의 `shouldDistributeAll()` — `@MX:NOTE` 부착. opt-out 정책의 single decision point 임을 표시.
- `internal/template/catalog_slim_audit_test.go` — `@MX:NOTE` 부착. audit suite intent 명시 (lang_boundary_audit_test.go 선례 + CATALOG-001 의 catalog_tier_audit_test.go 선례).

P0 / P1 / P2 위반 0 을 목표로 한다.

## Estimated Complexity

- **Task count**: 27 (M1: 6 [T1.1~T1.6], M2: 6 [T2.1~T2.6], M3: 7 [T3.1~T3.7], M4: 4, M5: 5 [T5.1~T5.5]). +5 from baseline due to DEFECT-4 (T2.5/T2.6/T3.7 harness guard), DEFECT-2 (T3.6 read-only audit), DEFECT-6 (T1.6 split), REC-3 (T5.4 unconditional), REC-5 (T3.3 EC4 sub-assertion within existing task).
- **LOC delta**:
  - `slim_fs.go`: 150-200 LOC.
  - `slim_fs_test.go`: 200-260 LOC.
  - `catalog_slim_audit_test.go`: 220-300 LOC (+40 LOC for T3.6 read-only audit + EC4 sub-assertion).
  - `embed_catalog.go`: 35-55 LOC (no EmbeddedRawForInternal; +5 LOC for `NewSlimDeployerWithRenderer`).
  - `slim_guard.go` + `slim_guard_test.go`: 10 + 25 = ~35 LOC (REQ-021).
  - `init.go` modifications: ~30-35 LOC (slim branch + REQ-021 notice).
  - `init_test.go` 신규 sub-test (M4-T4.1 전략 B): ~60 LOC.
  - `CHANGELOG.md` entry: ~12 lines.
  - godoc / help text: ~25 lines (+5 for slim_guard.go).
  - **Total: ~770-1000 LOC** (Go + Markdown).
- **Risk level**: **Low-to-Medium** (downgraded from Medium) — R3 builder-harness 가 in-SPEC guard 로 mitigation. R5 (EmbeddedRawForInternal misuse) eliminated. R7 (fs.Sub interface bypass) 이 mitigation 강도가 가장 큼.
- **사용자 검토 필요 지점**: Open Decisions 5건 confirm (pre-M1), `init_test.go` 전략 선택 (M4-T4.1).

## plan-in-main + plan-auditor PASS Criteria

본 SPEC 은 `plan/SPEC-V3R4-CATALOG-002` 브랜치에서 작성되며 (PR #822 doctrine), 단일 commit 으로 묶어 plan-auditor 가 단일 diff 로 검토 가능하도록 한다.

plan-auditor PASS 기준 (예상 dimensions):

- Functionality (40%): EARS 21 REQ 가 verifiable + AC mapping 완전 (REQ 누락 0; REQ-003은 EC5 read-only audit, REQ-020 은 Scenario 8 CHANGELOG G/W/T, REQ-021 은 EC6 builder-harness guard 로 매핑). Filter semantics 가 명확하고 회귀 테스트 (M4) 가 포함.
- Security (25%): D7 lock 보존 (deployer.go no-modify). Read-only wrapper (REQ-003 가 reflective + race-detector 두 가지로 검증). Path traversal 등 보안 회귀 0. catalog.yaml 미존재 시 fail-fast (REQ-008). `embeddedRaw` encapsulation invariant 유지 (DEFECT-5 fix).
- Craft (20%): EARS 6 분류 (Ubiquitous / Event-Driven / State-Driven / Optional / Unwanted Behavior + Runtime Guard) 모두 표현. Sentinel discipline (`t.Errorf` not `t.Logf`) 처음부터 명시. Plan tasks 가 prefix logic 명확화 (T1.4/T1.5/T1.6 fs.Sub calling convention) 와 encapsulated constructor 패턴 (T2.4) 으로 implementer ambiguity 제거.
- Consistency (15%): CATALOG-001 의 D7 lock + tier semantics + LoadCatalog API 와 일관. proposal.md scope 명시적 재정의 (directory relocation → manifest-driven filter) 가 Background 와 Exclusions 양쪽에 정합. catalog.yaml ground truth (20+20+17+7+1=65) 가 spec.md / acceptance.md / spec-compact.md 모두에서 일치.

목표 overall_score: **≥ 0.88** (iter 1: 0.81 with 6 unresolved defects → iter 2 projection: 0.89-0.92 after full fix).

## Cross-SPEC Boundary Definitions

본 SPEC 과 인접 SPEC 의 경계를 1-line 으로 명시:

- **vs CATALOG-001**: 본 SPEC 은 manifest 의 *소비자*. manifest 자체는 변경 없음.
- **vs CATALOG-003 (moai pack CLI)**: 본 SPEC 은 "core 만 deploy" (자동), CATALOG-003 은 "사용자 명령으로 pack 추가" (인터랙티브). 본 SPEC 머지 후 CATALOG-003 머지 전 윈도우에서는 optional pack 사용 시 사용자가 `MOAI_DISTRIBUTE_ALL=1` 로 대응.
- **vs CATALOG-004 (safe update sync)**: 본 SPEC 은 init-only filter, CATALOG-004 는 update-only drift sync. update.go 미수정 invariant 가 두 SPEC 의 경계 보장.
- **vs CATALOG-005 (/moai project interview)**: 본 SPEC 은 user interaction 변경 0, CATALOG-005 가 harness opt-out / pack 추천 책임.
- **vs CATALOG-006 (moai doctor catalog)**: 본 SPEC 은 distribution, CATALOG-006 은 진단/리포팅. 직접 의존성 없음.
- **vs CATALOG-007 (migration docs)**: 본 SPEC 의 BREAKING CHANGE 항목이 CATALOG-007 의 4-locale docs-site sync 의 입력 데이터.
