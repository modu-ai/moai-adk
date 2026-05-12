---
id: SPEC-V3R4-CATALOG-002
version: "0.2.0"
status: completed
created_at: 2026-05-12
updated_at: 2026-05-12
author: GOOS행님
priority: High
labels: [catalog, distribution, slim-init, deployer, tier-filter]
issue_number: null
depends_on: [SPEC-V3R4-CATALOG-001]
related_specs: [SPEC-V3R4-CATALOG-003, SPEC-V3R4-CATALOG-004, SPEC-V3R4-CATALOG-005]
---

# SPEC-V3R4-CATALOG-002: Slim Init via Catalog Tier Filter

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-12 | GOOS행님 | Initial draft. Wave 2 Distribution 첫 SPEC. CATALOG-001 manifest 위에 `moai init` 의 default deploy 를 tier=`core` 만으로 좁히는 slim distribution layer. 디렉토리 재배치 (proposal.md 의 원안) 가 아니라 manifest-driven runtime filter — D7 lock (deployer.go no-modify) 을 indirection (SlimFS wrapper) 으로 보존. |
| 0.1.1 | 2026-05-12 | manager-spec | plan-auditor iter 1 (REVISE 0.81) 반영. 6 DEFECT 수정: (1) optional pack agent count 8→7 정정 (catalog.yaml ground truth, 20+20+17+7+1=65), (2) REQ-003 AC mapping (read-only + race-detector EC5 추가, sentinel `CATALOG_SLIM_NOT_READONLY`), (3) REQ-020 G/W/T (Scenario 8 CHANGELOG BREAKING CHANGE), (4) R3 builder-harness runtime guard (REQ-021 신설 + 5-10 LOC `AssertBuilderHarnessAvailable` helper, sentinel `CATALOG_SLIM_HARNESS_MISSING`), (5) `EmbeddedRawForInternal` 캡슐화 위반 제거 → `NewSlimDeployerWithRenderer(cat, renderer)` constructor (embeddedRaw 외부 노출 금지 invariant 유지), (6) plan.md T1.4/T1.5 prefix 로직 정정. RECs: REQ-002 wording (REC-1), catalog_doc.md unconditional (REC-3), CI jobs count 14 (REC-4), M3-T3.3 EC4 sub-assertion (REC-5). |
| 0.2.0 | 2026-05-12 | manager-docs | Sync phase 완료. Run PR #867 (commit `09e4a438f`, squash-merged `d15869bb7`) 머지로 implementation finalized. spec status `draft → completed`. Implementation Notes section 추가 (delivered artifacts / divergence / quality gates / evaluator findings / deferred items). |

## Overview

CATALOG-001 이 lock-in 한 3-tier manifest (`internal/template/catalog.yaml`, 65 entries, `core` / `optional-pack:<name>` / `harness-generated`) 위에서 **`moai init` 의 default deploy 경로를 `tier == "core"` 자산만으로 좁히는 slim distribution layer** 를 도입한다. 본 SPEC 은 디렉토리 구조 (`internal/template/templates/.claude/skills/...`) 를 그대로 두고, **deployer.go 를 수정하지 않으며** (CATALOG-001 의 D7 lock 보존), `fs.FS` 레벨의 read-only wrapper (`SlimFS`) 로 `optional-pack:*` 및 `harness-generated` entry 의 path 를 숨겨 deploy 를 차단한다.

본 SPEC 은 brain proposal.md 의 SPEC-V3R4-CATALOG-002 framing ("디렉토리 재배치") 을 다음 두 가지 이유로 **manifest-driven runtime filter 로 재정의**한다:

1. **Capability vs Implementation 분리**: proposal.md 의 사용자-facing Capability 는 "`moai init` 은 기본적으로 core 만 deploy" (Capability 1 마지막 bullet) 이다. 디렉토리 재배치는 그 capability 를 달성하기 위한 *하나의* 구현 선택지였고, CATALOG-001 의 manifest 가 이미 tier 분류를 단일 source-of-truth 로 보유하므로 디렉토리 재배치는 불필요한 churn 이다.
2. **D7 invariant 와의 상호작용**: CATALOG-001 의 D7 lock (deployer.go 미수정) 을 깨지 않고도 slim 동작이 가능하다. SlimFS 가 `fs.FS` 인터페이스 수준에서 동작하므로 deployer.go 의 walk-and-render 로직은 변경 없이 더 작은 FS 만 받게 된다.

기존 동작 (`init` 이 모든 자산을 deploy) 으로의 복귀는 두 가지 opt-out 경로로 보장된다: `MOAI_DISTRIBUTE_ALL=1` 환경변수 또는 `moai init --all` CLI flag. `moai update` 는 본 SPEC 의 영향을 받지 않는다 (full FS 유지) — slim 동기화는 SPEC-V3R4-CATALOG-004 의 책임.

본 SPEC 의 결과물은 (1) `internal/template/slim_fs.go` `SlimFS()` wrapper, (2) `internal/template/slim_fs_test.go` 단위 테스트, (3) `internal/template/catalog_slim_audit_test.go` invariant audit, (4) `internal/template/embed_catalog.go` 의 `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` constructor (raw FS 외부 노출 없이 캡슐화), (5) `internal/template/slim_guard.go` 의 `AssertBuilderHarnessAvailable()` 런타임 가드 (~10 LOC, REQ-021), (6) `internal/cli/init.go` 의 `--all` flag + `shouldDistributeAll()` helper + `NewSlimDeployerWithRenderer` 호출 (~30 LOC delta), (7) CHANGELOG.md BREAKING CHANGE 항목이다.

[INVARIANT] 패키지 변수 `embeddedRaw` 는 `internal/template/` 패키지 외부에 노출되지 않는다 (D7 lock 정신 — encapsulation 보호). 모든 slim 동작은 `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()` 두 export 만으로 호출 가능하다.

## Background

본 SPEC 의 의사결정 근거 체인:

1. `.moai/brain/IDEA-003/proposal.md` — 7-SPEC initiative + Wave 순서 + Capability 1 ("`moai init` 은 기본적으로 core 만 deploy"). proposal 은 본 SPEC 을 "디렉토리 재배치" 로 기술했으나, capability 자체는 manifest-driven filter 로도 달성 가능.
2. `.moai/brain/IDEA-003/ideation.md` — "skill 의 존재 자체가 비용 (description 은 항상 budget 차지)" invariant (Phase 5.2 First Principles). 신규 프로젝트가 받는 skill 갯수가 적을수록 context budget 절약 + trigger 정확도 향상.
3. `.moai/brain/IDEA-003/research.md` — Anthropic skill description budget 1% 룰 (Opus 4.7 1M context → 10K tokens 예산). 사용자가 marketplace plugin 을 추가하는 순간 overflow → MoAI 자체 skill 의 description 이 drop 될 가능성. core 18-20개로 줄이면 모든 description 이 풀 텍스트로 유지됨.
4. `.moai/specs/SPEC-V3R4-CATALOG-001/spec.md` + `progress.md` — manifest 의 `tier` / `path` / `hash` 필드 lock-in. typed loader `LoadCatalog(fs.FS) (*Catalog, error)`, `Catalog.AllEntries()`, `Catalog.LookupSkill/LookupAgent`. D7 lock (`deployer.go` no-modify) 의 정당화 — 회귀 위험 격리.
5. `internal/template/deployer.go:72-185` — `Deploy()` 가 `fs.WalkDir(d.fsys, ".")` 로 일률적으로 모든 파일을 순회. fsys 가 SlimFS 이면 자동으로 hidden path 는 walk 에서 누락됨. 추가 코드 없이도 filter 효과 발생.
6. `internal/template/embed.go:28-30` — `//go:embed all:templates` + `//go:embed catalog.yaml` directive. `embeddedRaw` 가 package-private 이므로 `LoadEmbeddedCatalog()` convenience export 필요.
7. `internal/cli/init.go:293-301` — slim filter 적용 지점. `EmbeddedTemplates()` 호출 직후 SlimFS 로 wrap.

선례 참조:

- **CATALOG-001 의 D7 처리 패턴**: deployer.go 수정 없이 새 코드 path 만 추가. 회귀 위험 0 으로 입증. 본 SPEC 도 동일 원칙 — SlimFS 는 `fs.FS` 인터페이스 측에서 동작하여 deployer 의 내부 로직과 결합되지 않는다.
- **`internal/template/lang_boundary_audit_test.go`**: sentinel + parallel test 패턴. 본 SPEC 의 `catalog_slim_audit_test.go` 가 동일 패턴 차용.
- **CATALOG-001 evaluator-active 의 sentinel discipline 교훈**: hash sentinel 이 `t.Logf` 로 작성되어 advisory 였던 결함을 PR #863 에서 `t.Errorf` 로 변경. 본 SPEC 은 처음부터 모든 sentinel emission 에 `t.Errorf` 사용 명시 (plan.md M3 참조).

## EARS Requirements

### 1. SlimFS Wrapper Existence and Contract (Ubiquitous)

REQ-CATALOG-002-001: The system shall provide a function `template.SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` in `internal/template/slim_fs.go` that returns a read-only filesystem view exposing only files reachable through tier=`core` catalog entries plus all non-catalog template files.

REQ-CATALOG-002-002: The `fs.FS` returned by `SlimFS(...)` (after the internal `fs.Sub` step) shall NOT include the `templates/` prefix in its path namespace, matching the contract of `EmbeddedTemplates()` so the result is a drop-in replacement for the existing consumer chain.

REQ-CATALOG-002-003: SlimFS shall be a read-only wrapper. The implementation shall satisfy the following observable invariants verifiable at test time: (a) the `slimFS` struct contains no `sync.*` field, no channel field, and no unexported mutable field beyond the immutable allowed-path set computed at construction time; (b) parallel `fs.Stat` / `fs.ReadFile` / `fs.WalkDir` calls under `go test -race` produce no data races; (c) no goroutine is started by `SlimFS()` or by any method on the returned FS. Violations are detected by the audit suite with sentinel `CATALOG_SLIM_NOT_READONLY`.

REQ-CATALOG-002-004: The package shall export a convenience function `template.LoadEmbeddedCatalog() (*Catalog, error)` that loads `catalog.yaml` from the package-private `embeddedRaw` FS, returning the same `*Catalog` type as `LoadCatalog`, so callers in `internal/cli/` need not access `embeddedRaw` directly. The package shall also export a `template.NewSlimDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error)` constructor that builds the slim Deployer end-to-end (internal `SlimFS` wrap + Deployer composition); external packages MUST use this constructor and MUST NOT receive `embeddedRaw` directly. The symbol `embeddedRaw` shall remain unexported.

### 2. Non-Modification Invariants (Ubiquitous — preserves CATALOG-001 D7 lock)

REQ-CATALOG-002-005: The file `internal/template/deployer.go` shall remain unmodified by this SPEC. All slim distribution behavior shall be achieved through the `fs.FS` argument passed into the existing `NewDeployer` / `NewDeployerWithRenderer` constructors.

REQ-CATALOG-002-006: The file `internal/cli/update.go` shall remain unmodified by this SPEC. The `moai update` command shall continue to deploy the full unfiltered embedded FS until SPEC-V3R4-CATALOG-004 introduces safe drift-based slim sync.

### 3. Default Slim Init (Event-Driven)

REQ-CATALOG-002-007: When a user runs `moai init` and neither the `MOAI_DISTRIBUTE_ALL` environment variable nor the `--all` flag is set, the command shall pass a `SlimFS`-wrapped filesystem to the Deployer such that only catalog entries with `tier == "core"` (plus all non-catalog template files such as rules, output-styles, .moai/config, CLAUDE.md, .gitignore) are written to the project root.

REQ-CATALOG-002-008: When `moai init` is unable to load `catalog.yaml` (e.g., missing from the binary, malformed YAML, or any error returned by `LoadEmbeddedCatalog`), the command shall fail with a wrapped error containing the sentinel substring `CATALOG_LOAD_FAILED` BEFORE invoking the Deployer or writing any files to disk.

REQ-CATALOG-002-009: When `moai update` runs, the command shall continue to use the unfiltered `EmbeddedTemplates()` filesystem; SlimFS shall NOT be applied to the update code path within the scope of this SPEC.

### 4. SlimFS Read Behavior (State-Driven)

REQ-CATALOG-002-010: While SlimFS is active, calls to `fs.Stat` or `fs.ReadFile` on a path that corresponds to a catalog entry with `tier != "core"` (i.e., `optional-pack:<name>` or `harness-generated`) shall return an error satisfying `errors.Is(err, fs.ErrNotExist)`.

REQ-CATALOG-002-011: While SlimFS is active, `fs.WalkDir` shall NOT visit any path corresponding to a hidden non-core entry; the path shall not appear in the walk function's invocations even with `entry.IsDir() == true`.

### 5. Opt-Out Mechanisms (Optional)

REQ-CATALOG-002-012: Where the environment variable `MOAI_DISTRIBUTE_ALL` is set to the literal value `"1"` or `"true"` (case-insensitive comparison for `"true"`), `moai init` shall bypass SlimFS and pass the unfiltered `EmbeddedTemplates()` filesystem to the Deployer, restoring the pre-CATALOG-002 deployment surface.

REQ-CATALOG-002-013: Where `moai init` is invoked with the `--all` boolean flag, the command shall bypass SlimFS and pass the unfiltered `EmbeddedTemplates()` filesystem to the Deployer, with the same effect as `MOAI_DISTRIBUTE_ALL=1`.

### 6. Audit Invariants (Unwanted Behavior)

REQ-CATALOG-002-014: If a path under SlimFS resolves to a file backed by a catalog entry whose `tier` is NOT `"core"`, then the audit test `TestSlimFS_HidesNonCoreEntries` shall fail with sentinel `CATALOG_SLIM_LEAK: <path> tier=<tier>`.

REQ-CATALOG-002-015: If a path under SlimFS corresponding to a catalog entry with `tier == "core"` is NOT reachable (i.e., `fs.Stat` returns an error), then the audit test `TestSlimFS_PreservesCoreEntries` shall fail with sentinel `CATALOG_SLIM_CORE_MISSING: <path>`.

REQ-CATALOG-002-016: If a non-catalog template path (e.g., `.claude/rules/`, `.claude/output-styles/`, `.moai/config/`, `CLAUDE.md`, `.gitignore` — any file under `templates/` that is NOT enumerated as a skill or agent in `catalog.yaml`) is hidden by SlimFS, then the audit test `TestSlimFS_PreservesNonCatalogFiles` shall fail with sentinel `CATALOG_SLIM_OVER_FILTER: <path>`.

REQ-CATALOG-002-017: If `fs.WalkDir` traversal of SlimFS visits a hidden non-core entry path, then the audit test `TestSlimFS_WalkDirNoLeak` shall fail with sentinel `CATALOG_SLIM_WALK_LEAK: <path>`.

### 7. Backward Compatibility (Unwanted Behavior)

REQ-CATALOG-002-018: If the pre-existing test suite `internal/template/deployer_test.go` is executed against the unfiltered `EmbeddedTemplates()` filesystem after this SPEC is merged, then every test case shall remain GREEN; no test logic shall be altered by this SPEC to accommodate slim mode.

REQ-CATALOG-002-019: If both `MOAI_DISTRIBUTE_ALL=1` and `--all` are set simultaneously, then `moai init` shall proceed with the unfiltered (full) deployment without raising an error; the two opt-outs shall be idempotent and combinable.

### 8. Documentation (Ubiquitous)

REQ-CATALOG-002-020: The system shall record the slim-init behavior change in `CHANGELOG.md` under the Unreleased section with an explicit `BREAKING CHANGE` annotation, listing both opt-out mechanisms (`MOAI_DISTRIBUTE_ALL` env var, `--all` flag).

### 9. Runtime Guard for Missing Builder-Harness (Event-Driven)

REQ-CATALOG-002-021: When `moai` (the CLI binary or any code path under `internal/`) attempts to access the `builder-harness` agent (`.claude/agents/moai/builder-harness.md`) and the file is absent on disk (i.e., omitted by slim init), the system shall return a non-nil error wrapped with the sentinel substring `CATALOG_SLIM_HARNESS_MISSING` whose error message includes BOTH (a) the opt-out instruction `MOAI_DISTRIBUTE_ALL=1` (or equivalently `moai init --all`) AND (b) the deferral reference `SPEC-V3R4-CATALOG-005`. This guard shall be implemented as `template.AssertBuilderHarnessAvailable(projectFS fs.FS) error` in `internal/template/slim_guard.go` (estimated 5-10 LOC) and shall be consumable by future callers (e.g., `moai doctor`, harness workflow loaders). Additionally, `moai init` slim path shall print a one-line informational notice on stdout containing both strings immediately after successful slim deploy, so that users discover the constraint without invoking the guard directly.

## Acceptance Criteria

See `acceptance.md` for the full Given-When-Then scenarios (**8 scenarios + 6 edge cases**, 21 REQ all mapped). High-level acceptance:

- AC-CATALOG-002-01: Default `moai init` produces a project with only core-tier skills/agents. The catalog ground truth is `20 core skills + 20 core agents + 17 optional-pack skills + 7 optional-pack agents + 1 harness-generated agent = 65 total`; slim mode deploys the 40 core entries and hides the 25 non-core entries (17 skills + 7 optional agents + 1 harness agent). All non-catalog files (rules/, output-styles/, .moai/config/, CLAUDE.md, .gitignore) deploy unchanged. (maps REQ-001, REQ-002, REQ-004, REQ-007, REQ-010, REQ-011)
- AC-CATALOG-002-02: SlimFS opt-out via `MOAI_DISTRIBUTE_ALL=1` or `--all` flag restores full deploy bit-identical to pre-CATALOG-002 behavior. (maps REQ-012, REQ-013, REQ-019)
- AC-CATALOG-002-03: SlimFS audit suite (4 sub-tests) guards leak / over-filter / walk leak / core-missing invariants. All sentinel emissions use `t.Errorf` (NOT `t.Logf`). (maps REQ-014, REQ-015, REQ-016, REQ-017)
- AC-CATALOG-002-04: D7 lock preserved — `git diff` of `internal/template/deployer.go` is empty post-merge. `internal/cli/update.go` likewise unchanged. `embeddedRaw` is NOT exported from the `template` package; external callers route through `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)`. (maps REQ-004, REQ-005, REQ-006, REQ-018)
- AC-CATALOG-002-05: Catalog load failure fails init fast with `CATALOG_LOAD_FAILED` substring BEFORE writing any files to disk. (maps REQ-008)
- AC-CATALOG-002-06: `moai update` deploys full FS (no slim filter applied). (maps REQ-009)
- AC-CATALOG-002-07: CHANGELOG documents BREAKING CHANGE with opt-out mechanisms (verified by Scenario 8 G/W/T). (maps REQ-020)
- AC-CATALOG-002-08: SlimFS is observably read-only — reflective struct inspection finds no mutable/sync/channel field, and `go test -race` over parallel `fs.Stat`/`fs.ReadFile` calls produces no race report. Sentinel `CATALOG_SLIM_NOT_READONLY` fires on violation. (maps REQ-003 via Scenario EC5)
- AC-CATALOG-002-09: Slim-mode user attempting to invoke harness workflow receives a friendly error from `AssertBuilderHarnessAvailable` containing both `MOAI_DISTRIBUTE_ALL=1` and `SPEC-V3R4-CATALOG-005`. Init slim path also prints the notice on stdout. Sentinel `CATALOG_SLIM_HARNESS_MISSING`. (maps REQ-021 via Scenario EC6)

## Files to Modify / Create

[NEW] `internal/template/slim_fs.go` — `SlimFS()` constructor + `slimFS` private type implementing `fs.FS` + `fs.ReadDirFS` + `fs.StatFS`. Computes allowed-path prefix set from `Catalog.AllEntries()` filtered by `tier == TierCore`, plus a deny-set for hidden non-core entries. Returns `fs.Sub(wrapper, "templates")` so result is drop-in compatible with `EmbeddedTemplates()`. Estimated 150-200 LOC.

[NEW] `internal/template/slim_fs_test.go` — Unit tests using `testing/fstest.MapFS` for synthetic catalogs and real catalog via `LoadEmbeddedCatalog()`. Covers: core entry visible, optional-pack entry hidden, harness-generated entry hidden, non-catalog file visible, nested path under core skill visible (e.g., `moai/workflows/plan.md`), `fs.Stat` returns `fs.ErrNotExist` for hidden, `fs.WalkDir` skips hidden subtree. Estimated 200-260 LOC.

[NEW] `internal/template/catalog_slim_audit_test.go` — Integration audit suite (4 parallel sub-tests) against real embedded FS + catalog. Sentinels: `CATALOG_SLIM_LEAK`, `CATALOG_SLIM_CORE_MISSING`, `CATALOG_SLIM_OVER_FILTER`, `CATALOG_SLIM_WALK_LEAK`. All emissions via `t.Errorf` (CATALOG-001 evaluator-active lesson). Estimated 180-240 LOC.

[NEW] `internal/template/embed_catalog.go` — Tiny helper exporting `LoadEmbeddedCatalog() (*Catalog, error)` and `NewSlimDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error)`. The constructor internally calls `SlimFS(embeddedRaw, cat)` and composes the standard `NewDeployerWithRenderer`. The package variable `embeddedRaw` remains unexported — NO `EmbeddedRawForInternal` / `EmbeddedRawForTest` export. External packages (`internal/cli/`) consume slim deploy through these two exports only. Estimated 35-55 LOC. (Alternative: extend `embed.go` directly; chosen separate file to keep `embed.go` minimal per CATALOG-001 SPEC scope.)

[NEW] `internal/template/slim_guard.go` — Runtime guard helper. Exports `AssertBuilderHarnessAvailable(projectFS fs.FS) error` that returns a wrapped error with sentinel `CATALOG_SLIM_HARNESS_MISSING` when `.claude/agents/moai/builder-harness.md` is absent. Error message MUST include both `MOAI_DISTRIBUTE_ALL=1` and `SPEC-V3R4-CATALOG-005` substrings. Estimated 5-10 LOC + ~25 LOC test in `slim_guard_test.go`. (REQ-021.)

[MODIFY] `internal/cli/init.go` — Add `--all` flag definition (~3 LOC). Add `shouldDistributeAll(cmd) bool` helper (~10 LOC). Replace the SlimFS wiring block: when slim mode is active, call `cat, _ := template.LoadEmbeddedCatalog()` (failure → `CATALOG_LOAD_FAILED`) then `deployer, _ := template.NewSlimDeployerWithRenderer(cat, renderer)`; when full mode is active, retain `template.EmbeddedTemplates()` + `template.NewDeployerWithRenderer(fs, renderer)`. After successful slim deploy, print the REQ-021 informational notice (~3 LOC). Total delta ~30-35 LOC. Existing tests in `init_test.go` / `init_coverage_test.go` may require updates to handle the new default; those updates are tracked under M4-T4.1.

[MODIFY] `CHANGELOG.md` — Unreleased section: BREAKING CHANGE entry documenting slim init + opt-out mechanisms. ~12 lines.

[NO MODIFY] `internal/template/deployer.go` — **D7 lock preserved** per CATALOG-001 contract. SlimFS achieves filtering at the `fs.FS` boundary; deployer internals are not aware of tier semantics.

[NO MODIFY] `internal/cli/update.go` — Slim sync deferred to SPEC-V3R4-CATALOG-004. `moai update` continues to use the unfiltered `EmbeddedTemplates()` and `NewDeployerWithForceUpdate`.

[NO MODIFY] `internal/template/embed.go` — No `//go:embed` directive changes needed. The existing `//go:embed all:templates` + `//go:embed catalog.yaml` (added in CATALOG-001 T-023) already covers everything SlimFS needs.

[NO MODIFY] `internal/template/catalog.yaml` — Manifest data is not changed. Slim filter consumes the existing tier classification.

## Exclusions (What NOT to Build)

The following are explicitly **OUT OF SCOPE** for SPEC-V3R4-CATALOG-002 and deferred to follow-up SPECs:

- **Directory relocation** (`internal/template/templates/` reorganization into `packs/<pack>/` subtrees) — proposal.md 의 original CATALOG-002 framing 이었으나 manifest-driven filter 로 capability 가 달성되므로 indefinitely deferred. No follow-up SPEC currently scheduled; if reintroduced, it would be a refactor SPEC (no functional change).
- **`moai pack add|remove|list|available` CLI** — SPEC-V3R4-CATALOG-003 영역. 본 SPEC 은 core 만 deploy 하는 distribution scope 만 다루며, optional pack 의 사용자-facing install/uninstall 명령은 정의하지 않는다.
- **Safe update synchronization** (`moai update --catalog-sync`, drift detection, 3-way merge, snapshot/rollback) — SPEC-V3R4-CATALOG-004 영역. `moai update` 는 본 SPEC 영향 밖.
- **`/moai project` interview extension** (harness opt-out AskUserQuestion) — SPEC-V3R4-CATALOG-005 영역. 본 SPEC 은 user interaction 변경 없음 (interactive wizard 영향 0).
- **`moai doctor catalog` diagnostic** — SPEC-V3R4-CATALOG-006 영역.
- **4-locale migration documentation** (docs-site ko/en/ja/zh sync, inline migration banner in `moai update` first-run) — SPEC-V3R4-CATALOG-007 영역. 본 SPEC 의 CHANGELOG 항목은 영문/한글 단일 entry 로 한정.
- **builder-harness 자동 부트스트랩** — `harness-generated` tier 의 builder-harness agent 는 slim init 에서 deploy 되지 않는다. harness workflow 가 필요 시점에 builder-harness 를 생성하는 부트스트랩 흐름은 SPEC-V3R4-CATALOG-005 (또는 별도 SPEC) 영역. 본 SPEC 은 이 동작의 **부재** 만 기술하며 부트스트랩을 도입하지 않는다.
- **`moai update` slim 회귀** — 기존 프로젝트 의 update 시 hidden 자산 제거 / drift 비교 / 사용자 확인 흐름은 CATALOG-004 영역. 본 SPEC 단독 머지 시 기존 프로젝트의 `.claude/` 디렉토리에는 영향 0 (이전에 deploy 된 optional-pack/harness-generated 자산은 보존됨).
- **catalog.yaml mutation** — manifest 데이터 자체는 수정하지 않는다. tier 분류는 CATALOG-001 의 lock-in 을 그대로 사용.
- **CATALOG-001 deferred nice-to-have 3건 흡수** — path containment audit test, REQ-011/012 pack name regex audit test, BenchmarkLoadCatalog 는 본 SPEC 의 scope 가 아니다. CATALOG-001 의 evaluator-active iter 1 보고서대로 후속 SPEC (CATALOG-003 / -004 / -005 등) 또는 별도 폴리시 SPEC 으로 처리.
- **MOAI_DISTRIBUTE_ALL 외 추가 환경변수** — `MOAI_PACK_INCLUDE`, `MOAI_TIER` 등 fine-grained 환경변수는 본 SPEC 에서 정의하지 않는다. CATALOG-003 의 pack 명령이 사용자-facing fine-grained 제어를 담당.
- **Pack 자동 권장** — backend/frontend 프로젝트 인식 후 `moai pack add backend` 등을 자동 제안하는 흐름은 CATALOG-005 의 `/moai project` 인터뷰 영역.
- **`/moai harness` 인터랙티브 흐름 변경** — `moai-meta-harness` (`tier: core` 이므로 slim 에서도 deploy 됨) 와 `builder-harness` (`tier: harness-generated`, slim 에서 숨김) 의 상호작용은 본 SPEC 의 audit 범위 외. harness workflow 자체 동작은 변경 없음.

## Dependencies

- **Depends on**: SPEC-V3R4-CATALOG-001 (manifest schema + `LoadCatalog` + tier classification 65 entries). 본 SPEC 은 CATALOG-001 의 typed API 와 `tier` 값 invariant 를 입력으로 사용.
- **Blocks**: SPEC-V3R4-CATALOG-003 (`moai pack add` 가 hidden 자산을 사용자 명령으로 다시 노출하기 위해 본 SPEC 의 SlimFS 토대 위에서 install path 구현), SPEC-V3R4-CATALOG-004 (`moai update --catalog-sync` 가 slim 모드로 deploy 된 기존 프로젝트의 drift 계산 기준선을 본 SPEC 의 deploy surface 로 정의), SPEC-V3R4-CATALOG-005 (`/moai project` 인터뷰가 hidden 자산 install 옵션을 사용자에게 제시할 때 본 SPEC 의 opt-out 매커니즘과 정합성 유지).
- **Related (non-blocking)**: SPEC-V3R4-CATALOG-006 (`moai doctor catalog` 가 slim 모드 활성화 여부 + 누락 pack 을 표시).

## References

### Internal (필수 사전 read)
- `.moai/brain/IDEA-003/proposal.md` — Capability 1 ("`moai init` 은 기본적으로 core 만 deploy") + Wave 순서.
- `.moai/brain/IDEA-003/ideation.md` — Lean Canvas 1.2 (context budget) + Phase 5.2 First Principles (skill 존재가 비용).
- `.moai/brain/IDEA-003/research.md` — Anthropic skill description budget 1% rule + Finding 2 (10K tokens 예산).
- `.moai/specs/SPEC-V3R4-CATALOG-001/spec.md` — manifest 계약 + D7 lock 정당화 + Implementation Notes (Wave 1 foundation 완료 상태).
- `.moai/specs/SPEC-V3R4-CATALOG-001/progress.md` — CATALOG-001 evaluator-active iter 1 lessons (sentinel discipline: `t.Errorf` not `t.Logf`).

### Code Reference
- `internal/template/catalog.yaml` — 65-entry manifest. Tier breakdown: 20 core skills + 20 core agents in `catalog.core`; 9 optional packs in `catalog.optional_packs` containing 17 skills + 7 agents (ground truth verified by `awk 'NR>=230 && NR<406' catalog.yaml | grep -c "path: templates/.claude/<agents|skills>"`); 1 builder-harness agent in `catalog.harness_generated`. Total entries: 20+20+17+7+1 = 65.
- `internal/template/catalog_loader.go:113-125` — `LoadCatalog(fs.FS) (*Catalog, error)` typed accessor, depended on by SlimFS construction.
- `internal/template/embed.go:28-44` — `embeddedRaw` (package-private) + `EmbeddedTemplates() (fs.FS, error)` (sub-FS at `templates/`).
- `internal/template/deployer.go:18-36` (`Deployer` interface) + `:72-185` (`Deploy()` walk-and-render). **D7 lock target — no modification by this SPEC.**
- `internal/cli/init.go:119-360` (`runInit`) — entry point for slim filter wiring (lines 293-301 specifically).
- `internal/cli/update.go:445-479` — update deploy path (untouched by this SPEC).
- `internal/template/lang_boundary_audit_test.go` — audit pattern reference.

### Rules
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-005 — Template-First discipline.
- `.claude/rules/moai/development/coding-standards.md` — no hardcoded values; tier strings centralized in `catalog_loader.go` constants.
- `.moai/specs/SPEC-V3R4-CATALOG-001/spec.md` §"D7 권장 수정 반영" — deployer.go no-modify invariant.

## Open Questions (for plan-auditor / user review)

OQ1: `MOAI_DISTRIBUTE_ALL` 값 매칭 규칙 — `"1"` exact + `"true"` case-insensitive 가 권장안 (`os.Getenv` 의 일반 패턴). 추가로 `"yes"` / `"on"` 허용 여부?  
→ **권장**: 좁게 `"1"` 또는 `"true"` (case-insensitive). 다른 값은 fall-through 로 slim 모드 유지. 환경변수 동작 명확성 우선.

OQ2: `--all` flag 명칭 — `--all` vs `--full` vs `--include-optional`?  
→ **권장**: `--all`. 짧고 직관적. proposal.md 의 `moai pack add backend frontend design` 같은 명령 명령 surface 와 충돌하지 않음.

OQ3: Slim 모드 활성화 시 사용자 informational 출력 여부?  
→ **권장**: 짧은 안내 1줄 ("Deploying core templates only (slim mode). Use `--all` or `MOAI_DISTRIBUTE_ALL=1` for full deploy."). 사용자 마찰 최소화.

OQ4: harness-generated tier (`builder-harness`) 의 slim init 부재 동작 — CATALOG-002 단독으로는 builder-harness 가 deploy 되지 않으므로, `moai-meta-harness` workflow 가 호출될 때 builder-harness 부재 문제가 발생할 수 있음. 본 SPEC 에서 부트스트랩 도입할지?  
→ **권장**: **부트스트랩 도입 안 함**. 본 SPEC 의 Exclusions 에 명시. CATALOG-005 의 `/moai project` 인터뷰가 harness 활성화 시점에 builder-harness 생성을 책임진다. CATALOG-002 머지 후 CATALOG-005 머지 전 사이에 harness 호출 시 fail 가능 — 이 윈도우는 Sprint 12 timeline 내에서 좁게 유지.

OQ5: SlimFS 가 `fs.ReadDirFS` 인터페이스를 구현할지?  
→ **권장**: 구현. `fs.WalkDir` 가 효율적으로 동작하려면 `ReadDir` 구현이 필요. 추가 비용 작음 (~30 LOC).

위 Open Questions 는 plan-auditor 검토 시 lock-in 한다. M1 구현 시작 전에 사용자 confirm 권장.

## Implementation Notes

> 본 섹션은 SPEC-V3R4-CATALOG-002 의 run + sync phase 종료 시점에 작성됨 (2026-05-12, manager-docs). spec-first lifecycle level 1 — SPEC 본문은 amend 하지 않으며, 실제 구현 결과의 요약만 보존한다.

### Delivered Artifacts

| File | Type | LOC | Purpose |
|------|------|-----|---------|
| `internal/template/slim_fs.go` | New | ~150 | SlimFS read-only `fs.FS` wrapper (tier filter) |
| `internal/template/slim_fs_test.go` | New | ~230 | SlimFS unit tests (14 cases, 91.1% coverage) |
| `internal/template/catalog_slim_audit_test.go` | New | ~190 | Slim catalog audit (6 sub-tests, sentinel-driven) |
| `internal/template/embed_catalog.go` | New | ~50 | `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()` |
| `internal/template/embed_catalog_test.go` | New | ~120 | embed_catalog tests (4 cases) |
| `internal/template/slim_guard.go` | New | ~25 | `AssertBuilderHarnessAvailable()` runtime guard (REQ-021, 100% coverage) |
| `internal/template/slim_guard_test.go` | New | ~95 | slim_guard tests (3 cases) |
| `internal/template/emit_notice.go` | New | ~30 | Slim init notice emission |
| `internal/cli/init.go` | Modified | +35 / -2 | `--all` flag + `shouldDistributeAll()` + `NewSlimDeployerWithRenderer` call (100% branch coverage) |
| `internal/cli/init_slim_branch_test.go` | New | ~270 | slim branch integration tests (11 cases + emit_notice 1) |
| `internal/template/catalog_doc.md` | Modified | +20 / 0 | Slim distribution rationale unconditional |
| `CHANGELOG.md` | Modified | +18 / 0 | Wave 2 BREAKING CHANGE entry |

Total: 10 new source files + 3 modified files + 2 SPEC artifacts (spec/plan/acceptance/progress/tasks updates from plan phase). +1878 / -12 LOC across 13 commit files (commit `09e4a438f`, squash-merged as `d15869bb7`).

### Divergence Summary

Implementation aligned with SPEC intent. No scope expansion or unplanned additions beyond what plan.md anticipated. Supplementary artifacts were declared in plan.md ahead of time:

- `emit_notice.go` + `emit_notice_test.go`: Slim init user-facing notice (REQ-020 G/W/T, pre-declared in plan.md M3-T3.4)
- `slim_guard.go` + `slim_guard_test.go`: REQ-021 builder-harness runtime guard (added via plan-auditor iter 1 REVISE 0.81 DEFECT-4)

### D7 Lock Compliance

CATALOG-001 D7 invariant preserved verbatim. The following files were **NOT touched** by this SPEC:

- `internal/template/deployer.go` (walk-and-render logic unchanged)
- `internal/template/update.go`
- `internal/template/embed.go` (`embeddedRaw` remains package-private)
- `internal/template/catalog.yaml` (3-tier manifest unchanged)
- `internal/template/catalog_loader.go` (typed loader unchanged)

Encapsulation invariant: `embeddedRaw` is never exported. All slim flow accessible only via `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` constructor pair. `git grep "EmbeddedRawForInternal"` returns zero matches (DEFECT-5 closure, verified by audit test).

### Quality Gates (run phase, PR #867 verified)

| Gate | Verdict | Detail |
|------|---------|--------|
| `go vet ./...` | PASS | clean |
| `go test -race ./internal/template/...` | PASS | all suites green |
| `go test -race ./internal/cli/...` | PASS | including init_slim_branch_test |
| `golangci-lint run` | PASS | 0 issues |
| `internal/template` slim_fs.go coverage | PASS | 91.1% |
| `internal/template` slim_guard.go coverage | PASS | 100% |
| `internal/cli` shouldDistributeAll coverage | PASS | 100% |
| MX tags | PASS | 1 NOTE added (init.go:133 `shouldDistributeAll`), 3 existing ANCHOR tags preserved (runInit fan_in=3, EmbeddedTemplates fan_in=6, LoadCatalog fan_in≥3) |
| MX P1/P2 violations | PASS | 0 (no new goroutines/async, no fan_in regression) |
| plan-auditor (iter 2) | PASS | 0.91 |
| evaluator-active (iter 1) | PASS | 0.916 (no P0/P1) |
| CI 15 required checks | PASS | Lint / Test ubuntu+macos+windows / Build 5 platforms / Integration Tests 3 OS / CodeQL / Constitution Check / Labeler |

### Deferred / Follow-up Items

- **P2-2**: One evaluator finding deferred — architecturally unreachable error path in slim FS. Documented as defensive coding (no fix).
- **P3 findings (3)**: All P3 issues deferred to post-merge follow-up or absorbed into successor SPECs (CATALOG-003: builder-harness skill fetcher; CATALOG-004: slim sync `moai update` opt-in; CATALOG-005: pack metadata enrichment).
- **Optional pack count documentation**: catalog.yaml ground truth is 8 optional packs; SPEC v0.1.0 originally claimed 7 (off-by-one). Fixed in spec v0.1.1 (plan-auditor iter 1 DEFECT-1 acknowledged).

### Sentinel & Sentinel Discipline

CATALOG-002 introduced 2 new sentinels (per plan v0.1.1):

- `CATALOG_SLIM_NOT_READONLY` — emit on attempt to mutate SlimFS (`slim_fs.go` runtime guard + audit test)
- `CATALOG_SLIM_HARNESS_MISSING` — emit when `AssertBuilderHarnessAvailable()` cannot resolve builder-harness skill (`slim_guard.go`)

All sentinel emissions use `t.Errorf` in audit tests (no `t.Logf` advisory regression — lesson learned from CATALOG-001 PR #863 sentinel fix). Verified in `catalog_slim_audit_test.go`.

### Successor SPEC References

- **CATALOG-003** (Wave 2 next): builder-harness skill fetcher — fetches harness-generated tier skills on-demand during `/moai project` interview
- **CATALOG-004** (Wave 3): slim sync `moai update` opt-in (current PR keeps `moai update` full FS — out of scope per SPEC)
- **CATALOG-005** (Wave 4): pack metadata enrichment + marketplace prep
