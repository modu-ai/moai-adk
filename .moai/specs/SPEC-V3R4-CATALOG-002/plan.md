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

전체 task ~22개. M1 (5) + M2 (4) + M3 (5) + M4 (4) + M5 (4). development_mode 는 quality.yaml 기준 TDD 가정 — M3 audit suite 가 TDD RED → GREEN 진입점.

### M1 — SlimFS Wrapper API + Implementation

- T1.1: `internal/template/slim_fs.go` 생성. Package docstring + `SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` 시그니처 정의. 입력 검증 (nil rawFS / nil cat → error).
- T1.2: private `slimFS` struct 정의 (필드: `underlying fs.FS`, `denySet map[string]struct{}`). `fs.FS` interface (`Open(name string) (fs.File, error)`) 구현. `fs.StatFS` 와 `fs.ReadDirFS` interface 추가 구현 (audit + WalkDir 효율).
- T1.3: `computeDenySet(cat *Catalog) map[string]struct{}` 헬퍼 — `cat.AllEntries()` 순회하여 `tier != TierCore` 인 entry 의 `Path` 를 deny set 에 추가. Path 정규화 (templates/ prefix 유지, trailing slash 제거).
- T1.4: `(*slimFS) Open(name string) (fs.File, error)` — `name` (post-Sub view) 을 `templates/` prefix 와 결합하여 deny set 매칭. 매칭 시 `fs.ErrNotExist` wrap 한 error 반환. 매칭 안 되면 underlying.Open 으로 pass-through.
- T1.5: `(*slimFS) ReadDir(name string) ([]fs.DirEntry, error)` — underlying ReadDir 결과를 필터링. deny set prefix 매칭되는 entry 제거. 동시에 sub-directory 진입 차단을 위해 디렉토리 entry 도 deny 매칭 처리. 마지막으로 `fs.Sub(slimWrapper, "templates")` 로 templates/ 제거된 view 반환. 

**Risk in M1**: deny set 키 정규화 (slash 처리, 디렉토리 vs 파일 구분) 가 오타나면 leak 발생. 해결: T1.5 의 prefix matching 을 sentinel 테스트 (M3-T3.1) 가 잡는다. 또한 deny set 빌드 단계에서 catalog 의 entry.Path 가 `templates/.claude/skills/<name>/` 또는 `templates/.claude/agents/moai/<name>.md` 형태임을 단위 테스트로 확정 (slim_fs_test.go).

### M2 — CLI Integration (init.go)

- T2.1: `internal/cli/init.go:60` 의 flag 정의 블록에 `initCmd.Flags().Bool("all", false, "Deploy all catalog entries (core + optional packs + harness-generated). Bypasses slim mode (CATALOG-002).")` 추가.
- T2.2: `shouldDistributeAll(cmd *cobra.Command) bool` 헬퍼 함수 추가 (init.go 또는 init_helpers.go 신규 파일). 로직: `cmd.Flags().GetBool("all")` OR `os.Getenv("MOAI_DISTRIBUTE_ALL") == "1"` OR `strings.EqualFold(os.Getenv("MOAI_DISTRIBUTE_ALL"), "true")`.
- T2.3: `internal/cli/init.go:293-301` 블록 수정 — `embeddedFS, err := template.EmbeddedTemplates()` 호출 후 `shouldDistributeAll(cmd) == false` 시 다음을 수행:
  ```go
  cat, catErr := template.LoadEmbeddedCatalog()
  if catErr != nil {
      return fmt.Errorf("CATALOG_LOAD_FAILED: %w", catErr)
  }
  slim, slimErr := template.SlimFS(template.EmbeddedRawForInternal(), cat)
  if slimErr != nil {
      return fmt.Errorf("CATALOG_LOAD_FAILED: slim fs construction: %w", slimErr)
  }
  embeddedFS = slim
  ```
  (T2.4 의 `EmbeddedRawForInternal` 가 raw FS 노출 helper.)
- T2.4: `internal/template/embed_catalog.go` 신규 파일. exports:
  - `LoadEmbeddedCatalog() (*Catalog, error)` — `LoadCatalog(embeddedRaw)` wrap.
  - `EmbeddedRawForInternal() fs.FS` — `embeddedRaw` 그대로 노출 (raw 의 의미: `templates/` prefix 유지). 명명에 `Internal` 을 붙여 사용 범위가 internal/cli 임을 시그널링.
  - Slim 모드 활성화 시 stdout 에 1-line 안내 출력: `"Deploying core templates only (slim mode). Use --all or MOAI_DISTRIBUTE_ALL=1 for full deploy."`

**Risk in M2**: `EmbeddedRawForInternal` 가 외부 패키지에 raw FS 를 노출하여 SlimFS bypass 가 너무 쉬워질 수 있음. 대안: `template.NewDeployerSlim(cat *Catalog) (Deployer, error)` 헬퍼로 raw FS 노출을 완전히 가린다. 그러나 init.go 의 기존 `NewDeployerWithRenderer` 호출 패턴을 보존하기 위해 fs.FS 교체 방식을 채택 (D7 indirection 의 자연스러운 결과). README + godoc 에 "external packages MUST NOT call EmbeddedRawForInternal" 명시.

### M3 — Audit Suite (TDD RED → GREEN)

모든 sentinel emission 은 **`t.Errorf` 사용 (NOT `t.Logf`)** — CATALOG-001 evaluator-active iter 1 의 EC3 hash sentinel 교훈을 처음부터 적용.

- T3.1: `internal/template/catalog_slim_audit_test.go` 골격 작성 — `package template` + imports (`io/fs`, `strings`, `testing`, `testing/fstest`). Helper `loadSlimFS(t *testing.T) (fs.FS, *Catalog)` — `LoadCatalog(embeddedRaw)` + `SlimFS()` 일괄 호출.
- T3.2: `TestSlimFS_HidesNonCoreEntries` — REQ-014 강제. `cat.AllEntries()` 순회하여 `tier != TierCore` 인 entry 의 `Path` 에 대해 `fs.Stat(slim, strings.TrimPrefix(entry.Path, "templates/"))` 호출. err 없거나 `errors.Is(err, fs.ErrNotExist)` 가 false 면 fail. Sentinel: `CATALOG_SLIM_LEAK: <path> tier=<tier>`.
- T3.3: `TestSlimFS_PreservesCoreEntries` — REQ-015 강제. tier == TierCore 인 entry 의 path 에 대해 `fs.Stat` 호출. err 가 있거나 not found 면 fail. Sentinel: `CATALOG_SLIM_CORE_MISSING: <path>`.
- T3.4: `TestSlimFS_PreservesNonCatalogFiles` — REQ-016 강제. 미리 정의한 non-catalog path 목록 (`.claude/rules/moai/core/zone-registry.md`, `.claude/output-styles/`, `.moai/config/sections/quality.yaml`, `CLAUDE.md`, `.gitignore`) 각각에 대해 `fs.Stat`. 누락되면 fail. Sentinel: `CATALOG_SLIM_OVER_FILTER: <path>`.
- T3.5: `TestSlimFS_WalkDirNoLeak` — REQ-017 강제. `fs.WalkDir(slim, ".", walkFn)` 실행. walk 가 방문하는 모든 path 를 수집한 뒤, 그 중 어느 하나라도 deny set 의 prefix 와 매칭되면 fail. Sentinel: `CATALOG_SLIM_WALK_LEAK: <path>`. 추가로 `t.Parallel()` 사용 (lang_boundary_audit_test.go 선례).

**Risk in M3**: harness-generated tier 의 `builder-harness` agent 가 hidden 인데, 만약 미래의 manifest 가 `builder-harness` 를 `tier: core` 로 옮기면 T3.3 가 fail 한다 — 의도된 회귀 (manifest 변경은 SPEC 으로 추적). plan-auditor 가 manifest 변경의 SPEC 참조를 강제하므로 (REQ-CATALOG-001-013), 본 SPEC 단독으로는 안전.

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
- T5.4: `internal/template/catalog_doc.md` (CATALOG-001 산출물) 에 SlimFS 관련 1개 paragraph 추가 — "Tier filter consumer: SlimFS()" cross-reference. 미존재 시 생략 가능 (optional).

## Risks & Mitigations

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | SlimFS deny set 의 path normalization 버그 (trailing slash, leading slash, OS path separator) 로 leak 발생 | 중 | 높음 | M3-T3.5 의 `TestSlimFS_WalkDirNoLeak` 가 fs.WalkDir 전체 walk 결과를 deny set 과 cross-check. unit test (T1.5 의 slim_fs_test.go) 가 동일 invariant 를 synthetic catalog 로도 검증. |
| R2 | `moai init` 의 기존 사용자가 slim 모드를 인지하지 못해 누락된 자산을 발견 후 confusion | 중 | 중 | M5-T5.1 CHANGELOG BREAKING CHANGE 명시 + M2-T2.4 stdout 1-line 안내 + M5-T5.3 init --help 텍스트. CATALOG-003 (`moai pack add`) 머지 전 사용자에게 "현재는 pack opt-in 명령이 없으므로 `--all` 사용 권장" 안내 가능. |
| R3 | harness-generated tier 의 builder-harness 가 hidden 되어 harness workflow 가 즉시 fail | 중 | 높음 | 본 SPEC Exclusions 에 명시 + SPEC-V3R4-CATALOG-005 의 `/moai project` 인터뷰에서 부트스트랩 책임. CATALOG-002 단독 머지 후 harness 호출 시 친절한 error 메시지 ("SPEC-V3R4-CATALOG-005 의 harness 부트스트랩 미적용") 출력은 follow-up SPEC 영역. |
| R4 | init_test.go 가 deployed file count 를 hardcode 하여 회귀 발생 | 중 | 낮 | M4-T4.1 전략 B (기존 테스트 보존 + slim 별도) 채택. 추가 작업 부담 최소. |
| R5 | `EmbeddedRawForInternal` 가 외부 패키지에서 오용되어 D7 indirection 의 보호가 우회됨 | 낮 | 중 | godoc 에 "internal use only — external packages MUST use EmbeddedTemplates" 명시. linter 룰 (`internal/` 패키지 import 제한) 검토 후속 SPEC 영역. |
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

- **Task count**: 22 (M1: 5, M2: 4, M3: 5, M4: 4, M5: 4).
- **LOC delta**:
  - `slim_fs.go`: 150-200 LOC.
  - `slim_fs_test.go`: 200-260 LOC.
  - `catalog_slim_audit_test.go`: 180-240 LOC.
  - `embed_catalog.go`: 30-50 LOC.
  - `init.go` modifications: ~30 LOC.
  - `init_test.go` 신규 sub-test (M4-T4.1 전략 B): ~60 LOC.
  - `CHANGELOG.md` entry: ~12 lines.
  - godoc / help text: ~20 lines.
  - **Total: ~700-900 LOC** (Go + Markdown).
- **Risk level**: **Medium** — D7 indirection + Go fs.FS interface 가 상대적으로 신선한 패턴. R7 (fs.Sub interface bypass) 이 mitigation 강도가 가장 큼.
- **사용자 검토 필요 지점**: Open Decisions 5건 confirm (pre-M1), `init_test.go` 전략 선택 (M4-T4.1).

## plan-in-main + plan-auditor PASS Criteria

본 SPEC 은 `plan/SPEC-V3R4-CATALOG-002` 브랜치에서 작성되며 (PR #822 doctrine), 단일 commit 으로 묶어 plan-auditor 가 단일 diff 로 검토 가능하도록 한다.

plan-auditor PASS 기준 (예상 dimensions):

- Functionality (40%): EARS 20 REQ 가 verifiable + AC mapping 완전 (REQ 누락 0). Filter semantics 가 명확하고 회귀 테스트 (M4) 가 포함.
- Security (25%): D7 lock 보존 (deployer.go no-modify). Read-only wrapper. Path traversal 등 보안 회귀 0. catalog.yaml 미존재 시 fail-fast (REQ-008).
- Craft (20%): EARS 5 분류 (Ubiquitous / Event-Driven / State-Driven / Optional / Unwanted Behavior) 모두 표현. Sentinel discipline (`t.Errorf` not `t.Logf`) 처음부터 명시.
- Consistency (15%): CATALOG-001 의 D7 lock + tier semantics + LoadCatalog API 와 일관. proposal.md scope 명시적 재정의 (directory relocation → manifest-driven filter) 가 Background 와 Exclusions 양쪽에 정합.

목표 overall_score: **≥ 0.88** (CATALOG-001 plan PASS 0.94 대비 약간 낮음 — 새 패턴 도입 risk premium).

## Cross-SPEC Boundary Definitions

본 SPEC 과 인접 SPEC 의 경계를 1-line 으로 명시:

- **vs CATALOG-001**: 본 SPEC 은 manifest 의 *소비자*. manifest 자체는 변경 없음.
- **vs CATALOG-003 (moai pack CLI)**: 본 SPEC 은 "core 만 deploy" (자동), CATALOG-003 은 "사용자 명령으로 pack 추가" (인터랙티브). 본 SPEC 머지 후 CATALOG-003 머지 전 윈도우에서는 optional pack 사용 시 사용자가 `MOAI_DISTRIBUTE_ALL=1` 로 대응.
- **vs CATALOG-004 (safe update sync)**: 본 SPEC 은 init-only filter, CATALOG-004 는 update-only drift sync. update.go 미수정 invariant 가 두 SPEC 의 경계 보장.
- **vs CATALOG-005 (/moai project interview)**: 본 SPEC 은 user interaction 변경 0, CATALOG-005 가 harness opt-out / pack 추천 책임.
- **vs CATALOG-006 (moai doctor catalog)**: 본 SPEC 은 distribution, CATALOG-006 은 진단/리포팅. 직접 의존성 없음.
- **vs CATALOG-007 (migration docs)**: 본 SPEC 의 BREAKING CHANGE 항목이 CATALOG-007 의 4-locale docs-site sync 의 입력 데이터.
