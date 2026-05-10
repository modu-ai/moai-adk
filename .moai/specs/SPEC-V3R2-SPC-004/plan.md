# SPEC-V3R2-SPC-004 Implementation Plan

> Implementation plan for **@MX anchor resolver (query by SPEC ID, fan_in, danger category)**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored on branch `plan/SPEC-V3R2-SPC-004` (Step 1 plan-in-main; base `origin/main` HEAD `73742e3ee`).
> Run phase will execute on a fresh worktree `feat/SPEC-V3R2-SPC-004` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Step 2.

## HISTORY

| Version | Date       | Author                                            | Description |
|---------|------------|---------------------------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B)             | Initial implementation plan. Scope: LSP `find-references` integration via powernap (G-01), `mx.yaml` `danger_categories:` 사용자 wire-up (G-02), `.moai/specs/*/spec.md` `module:` frontmatter 자동 로드 + `SpecAssociator` 주입 (G-03), `Resolver.ResolveAnchorCallsites()` API parity (G-04), `mx.yaml test_paths:` 패턴 wire-up (G-05), stderr format verification (G-06), performance benchmark (G-07), 16-언어 sweep (G-08). Existing `internal/mx/{resolver_query,fanin,danger_category,spec_association}.go` (PR #746 commit `68795dbe3`) provides 90%+ of the code surface. |

---

## 1. Plan Overview

### 1.1 Goal restatement

`spec.md` §1 의 핵심 목표를 milestone 분해:

> Expose a structured query API over the @MX TAG sidecar index (SPEC-V3R2-SPC-002) as an ACI-shaped tool: `moai mx query`. Callers can filter tags by SPEC ID, by Kind (NOTE/WARN/ANCHOR/TODO/LEGACY), by fan_in threshold, by danger category, and by file path prefix. Output format is JSON (primary) for tool consumption and a human-readable table (secondary) for terminal inspection. This same API is consumed by codemaps generation, evaluator-active scoring hints, and the /moai review phase.

### 1.2 Current State Audit (research.md §2 cross-reference)

`internal/mx/` 패키지의 SPC-004 표면은 이미 main 에 머지되어 있다 (research.md [E-02], [E-03] commit `68795dbe3`, PR #746). 본 SPEC 의 구현 진척도:

| 영역 | 상태 | Evidence |
|------|------|----------|
| `Query` struct + 10 filter 필드 | ✅ 구현 | research [E-04] |
| `Resolver.Resolve(query) (QueryResult, error)` | ✅ 구현 | research [E-05] |
| `Resolver.ResolveAnchor(anchorID) (Tag, error)` | ⚠ Signature 부분 일치 | research [E-13] (단일 Tag — spec verbatim 은 `[]Callsite`) |
| `TextualFanInCounter` + `isTestFile` | ✅ 구현 | research [E-07], [E-08] |
| `DangerCategoryMatcher` + 4 default 카테고리 | ✅ 구현 | research [E-09], [E-10] |
| `SpecAssociator` (path + body 합성) | ✅ 구현 | research [E-11], [E-12] |
| `validateQuery()` + 3 sentinel error | ✅ 구현 | research [E-15] |
| Pagination (default 100, max 10000) + offset | ✅ 구현 | research [E-06], [E-16] |
| JSON / table / markdown output | ✅ 구현 | research [E-17], [E-18], [E-19] |
| CLI `moai mx query` 10-flag | ✅ 구현 | research [E-20], [E-31] |
| `MOAI_MX_QUERY_STRICT=1` env 분기 | ✅ 구현 | research [E-21] |
| `--include-tests` flag | ✅ 구현 | research [E-22] |
| AND-composed multi-filter | ✅ 구현 | research [E-23] |
| **LSP-backed `LSPFanInCounter`** (G-01) | ❌ 격차 | research §2.3 G-01 |
| **`mx.yaml` `danger_categories:` user wire-up** (G-02) | ⚠ 부분 (yaml 태그 정의; load 로직 부재) | research [E-24] |
| **`.moai/specs/*/spec.md` `module:` 자동 로드 → `SpecAssociator`** (G-03) | ❌ 격차 | research §5.1 |
| **`Resolver.ResolveAnchorCallsites()` API parity** (G-04) | ❌ 격차 (additive, backward-compat) | research §2.3 G-04 |
| **`mx.yaml test_paths:` glob 패턴 wire-up** (G-05) | ⚠ 부분 (TestPaths field 정의; `isTestFile` 미사용) | research [E-09] line 18, [E-08] |
| **stderr message format 회귀 방지 fixture** (G-06) | ⚠ 부분 (substring 일치; explicit fixture 부재) | research [E-25] |
| **Performance benchmark fixture** (G-07) | ❌ 격차 | research §10 row 7 |
| **16-언어 sweep test** (G-08) | ❌ 격차 | research §2.3 G-08 |

Net actionable scope (Run-phase): **8 격차 해소** (G-01..G-08) + 정합성 검증 (template parity unaffected, race detector clean, coverage ≥ 85%).

### 1.2.1 Acknowledged Discrepancies

본 plan 이 spec.md 와 의도적으로 다르게 처리하는 부분 (research evidence 기반):

- **Spec §1 은 본 SPEC 이 "Expose a structured query API ..." 하는 것처럼 서술하나, 실제로는 90%+ 가 이미 main 에 존재** — SPC-004 PR #746 commit `68795dbe3` (2026-04-30) 가 spec.md §2.1 "In Scope" 의 모든 bullet (CLI subcommand, Go API, fan_in via fallback, danger category, JSON output, pagination, SPEC association) 을 구현 완료. 본 plan 의 milestone 은 따라서 "build" 가 아니라 **"complete API to spec parity + integrate LSP + wire user config"** 로 reframe 됨. Sync-phase HISTORY 에서 spec.md §1 "Expose ... API" 표현을 "complete the API to spec parity (LSP integration + user config wire-up + 16-language coverage)" 로 reconcile 권장.

- **`Resolver.ResolveAnchor(anchorID) []Callsite` signature 격차** — 현 `ResolveAnchor` 는 `(Tag, error)` 반환 (research [E-13]). spec §1 + §5 의 verbatim 은 `[]Callsite` (multi-callsite enumeration). 두 API 는 의미가 다르다 — 전자는 "anchor 위치 조회", 후자는 "anchor 의 모든 caller 위치 enumeration". plan-phase 결정 (research §9 OQ-9): existing `ResolveAnchor` 보존 + 신규 `ResolveAnchorCallsites(ctx, anchorID, projectRoot, includeTests) ([]Callsite, error)` 추가 (additive). spec.md sync HISTORY 에서 두 메서드 의도 차이 명시 권장.

- **JSON 출력 envelope 차이** — spec §5.3 REQ-021 "TruncationNotice in the output header" 는 stdout envelope 으로 해석 가능하나, 현 CLI 는 stdout = `[]TagResult` slice + stderr = TruncationNotice 메시지 (research [E-39], [E-40]). plan-phase 결정 (research §9 OQ-5): backward-compat 위해 현 형식 유지. Go API 의 `QueryResult` 는 envelope (Tags / TruncationNotice / TotalCount) 노출 — programmatic caller 충족.

- **`--danger` 미정의 카테고리 검증 격차** — 현 `validateQuery()` 는 `Kind` 만 enum-check (research [E-15]); `--danger` 임의 string 허용. plan-phase 결정 (research §9 OQ-4): G-02 의 일부로 `validateQuery()` 에 `DangerCategoryMatcher.ValidateCategory()` 호출 추가. Empty `--danger` (no filter) 는 검증 skip. spec §5.5 REQ-041 의 invalid query 범주에 포함.

- **LSP availability detection** — 현재 (research [E-21]) strictMode 분기는 무조건 `LSPRequired` 반환 (todo "(no LSP client)"). G-01 의 일부로 powernap server discovery 결과를 검출 로직으로 대체. strictMode 에서 LSP available 시 `LSPFanInCounter` 사용; unavailable 시 `LSPRequiredError`.

- **Performance budget verification 형식** — spec §7 의 <100ms / <2s 는 machine-dependent; CI 강제 X. plan-phase 결정 (research §9 OQ-7): benchmark fixture (`go test -bench`) 추가, 회귀 detect 용도. 절대값 assertion 은 advisory.

### 1.3 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

- **RED**: 8개 격차 영역마다 새 RED test fixture 추가:
  - `TestLSPFanInCounter_Count` (G-01) — powernap mock 으로 LSP textDocument/references 응답 → fan_in N. textual fallback 비교.
  - `TestLoadDangerConfig_UserCustomCategories` (G-02) — mx.yaml 에 사용자 정의 카테고리 → matcher 가 user pattern 매칭.
  - `TestValidateQuery_UnknownDanger_InvalidQueryError` (G-02) — `--danger frobnicate` (미정의) → `InvalidQueryError`.
  - `TestLoadSpecModules_FromFrontmatter` (G-03) — `.moai/specs/SPEC-X/spec.md` frontmatter 의 `module:` → map[specID][]modulePath.
  - `TestSpecAssociator_PathBased_FromLoader` (G-03) — loader 결과 → SpecAssociator 가 path-prefix 매칭으로 SPEC ID 반환.
  - `TestResolver_ResolveAnchorCallsites` (G-04) — anchor_id → []Callsite (file/line/method).
  - `TestIsTestFile_RespectsTestPathsConfig` (G-05) — mx.yaml test_paths: `["**/integration/**"]` → `tag.File = "integration/foo.go"` 가 test 로 분류.
  - `TestSidecarUnavailable_StderrFormat` (G-06) — sidecar 부재 시 stderr 가 정확히 "SidecarUnavailable" + "/moai mx --full" 둘 다 포함.
  - `BenchmarkResolver_Resolve_1KTags` (G-07) — 1K tag fixture 에 대한 b.N iteration 측정.
  - `BenchmarkResolver_Resolve_50AnchorsLSP` (G-07) — 50 ANCHOR + LSP-backed counter (mock) 측정.
  - `TestResolver_AllSixteenLanguages` (G-08) — 16개 언어 fixture에서 fan_in / SPEC association 동작.

- **GREEN**: 각 RED test 를 GREEN 으로 전환하기 위한 production code 추가/수정:
  - `internal/mx/fanin_lsp.go` 신규: `LSPFanInCounter` struct 구현 (powernap client 의존). FanInCounter 인터페이스 (research [E-07]) 구현.
  - `internal/mx/danger_category.go` 보강: `LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error)` 헬퍼 추가.
  - `internal/mx/resolver_query.go` 수정: `validateQuery()` 에 danger 검증 분기 추가.
  - `internal/mx/spec_loader.go` 신규: `LoadSpecModules(projectRoot string) (map[string][]string, error)` 헬퍼.
  - `internal/mx/resolver.go` 수정: `ResolveAnchorCallsites()` 메서드 추가 (additive). `Callsite` struct 정의 (`internal/mx/callsite.go` 또는 같은 파일).
  - `internal/mx/fanin.go` 수정: `isTestFile()` 가 user-configured glob patterns 을 받도록 변경 (또는 `isTestFileWithPatterns(file, patterns []string) bool` helper 신설).
  - `internal/cli/mx_query.go` 확장: `LoadDangerConfig` + `LoadSpecModules` 호출 → `query.dangerMatcher` / `query.specAssociator` 주입. LSP availability detect → `query.fanInCounter` 분기.

- **REFACTOR**: 공유 logic 추출:
  - `internal/mx/config.go` (SPC-002 의 잠재 신규 파일과 충돌 우려 — research §11 cross-ref 참조; SPC-002 plan §1.4 의 `internal/mx/config.go` 와 별도 파일로 분리. 본 SPEC 은 `internal/mx/spec_loader.go` 와 `internal/mx/danger_category.go` 의 `LoadDangerConfig()` 만 추가).
  - `@MX:NOTE` 태그로 결정 사유 명시 (plan §6).

### 1.4 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| LSPFanInCounter (powernap-backed) | `internal/mx/fanin_lsp.go` (신규 ~150 LOC) | REQ-003, REQ-020, REQ-030 |
| `LoadDangerConfig` helper | `internal/mx/danger_category.go` (+30 LOC) | REQ-001 (`--danger`), REQ-012 |
| `LoadSpecModules` helper | `internal/mx/spec_loader.go` (신규 ~80 LOC) | REQ-006 (a) |
| `Callsite` struct + `Resolver.ResolveAnchorCallsites()` | `internal/mx/callsite.go` (신규 ~30 LOC) + `internal/mx/resolver.go` (+50 LOC) | REQ-002, REQ-003 |
| `isTestFile` glob 패턴 wire-up | `internal/mx/fanin.go` (+30 LOC) | REQ-040 |
| `validateQuery` danger 검증 분기 | `internal/mx/resolver_query.go` (+20 LOC) | REQ-041 |
| CLI wire-up (config + LSP detect) | `internal/cli/mx_query.go` (+60 LOC) | REQ-001, REQ-006, REQ-012, REQ-013, REQ-030 |
| 11 신규 RED tests | `internal/mx/fanin_lsp_test.go`, `internal/mx/spec_loader_test.go`, `internal/mx/resolver_test.go` (확장), `internal/mx/danger_category_test.go` (확장), `internal/mx/fanin_test.go` (확장) (+~700 LOC) | T-SPC004-01..11 |
| 16-언어 sweep fixture | `internal/mx/resolver_query_test.go` (확장 ~150 LOC) | REQ-001 + 16-lang neutrality |
| Benchmark fixtures (advisory) | `internal/mx/resolver_query_bench_test.go` (신규 ~80 LOC) | spec §7 budget |
| Template parity check | `make build` regenerates `internal/template/embedded.go` | TRUST 5 Trackable |
| CHANGELOG entry | `CHANGELOG.md` Unreleased | TRUST 5 Trackable |
| MX tags per §6 | 6 tags (per §6 below) | mx_plan |

Embedded-template parity는 **applicable** (`make build` 후 `internal/template/embedded.go` 재생성). 그러나 본 SPEC 의 변경은 모두 `internal/` Go 코드이고 template 자산 변경 없음. 검증: `diff -r .claude/ internal/template/templates/.claude/` 변경 0 line 기대.

### 1.5 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-SPC004-NN IDs.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-SPC-004-001 | Ubiquitous (CLI subcommand + 10 flags) | AC-01..AC-15 (cross-cutting) | T-SPC004-12 (CLI wire-up); existing CLI tests |
| REQ-SPC-004-002 | Ubiquitous (Go API `Resolver.Resolve()`) | AC-05 | T-SPC004-12 (existing Resolve covers; G-04 adds Callsites) |
| REQ-SPC-004-003 | Ubiquitous (LSP-first + ast-grep/textual fallback) | AC-07 | T-SPC004-01, T-SPC004-02 (LSPFanInCounter) |
| REQ-SPC-004-004 | Ubiquitous (default JSON + `--format table`) | AC-05, AC-06 | (existing); AC-15 16-lang sweep regression |
| REQ-SPC-004-005 | Ubiquitous (JSON 11-field schema) | AC-05 | (existing); T-SPC004-15 sweep verifies schema |
| REQ-SPC-004-006 | Ubiquitous (SPEC association: path + body) | AC-01, AC-15 | T-SPC004-04, T-SPC004-05 (LoadSpecModules + SpecAssociator wire) |
| REQ-SPC-004-007 | Ubiquitous (pagination default 100/max 10000) | AC-08 | (existing); T-SPC004-13 benchmark verifies |
| REQ-SPC-004-010 | Event-driven (`--spec X --kind anchor`) | AC-01 | T-SPC004-04, T-SPC004-05 + existing |
| REQ-SPC-004-011 | Event-driven (`--fan-in-min 3 --kind anchor`) | AC-02 | T-SPC004-01, T-SPC004-02 + existing |
| REQ-SPC-004-012 | Event-driven (`--danger concurrency`) | AC-03 | T-SPC004-03 (LoadDangerConfig wire) + existing |
| REQ-SPC-004-013 | Event-driven (sidecar absent → SidecarUnavailable) | AC-04 | T-SPC004-09 (stderr format fixture) + existing |
| REQ-SPC-004-020 | State-driven (no LSP → textual + annotation) | AC-07 | T-SPC004-01, T-SPC004-02 |
| REQ-SPC-004-021 | State-driven (10K+ tags → auto limit + TruncationNotice) | AC-08 | (existing); T-SPC004-13 benchmark |
| REQ-SPC-004-030 | Optional (`MOAI_MX_QUERY_STRICT=1` → fail no LSP) | AC-09 | T-SPC004-02 (LSP-backed strictMode 분기) |
| REQ-SPC-004-031 | Optional (`--format markdown`) | AC-10 | (existing); T-SPC004-15 16-lang regression |
| REQ-SPC-004-040 | Complex (test fixture excluded; `--include-tests` override) | AC-11 | T-SPC004-07, T-SPC004-08 (test_paths glob wire) + existing |
| REQ-SPC-004-041 | Complex (zero match → `[]` exit 0; invalid filter → InvalidQuery exit 2) | AC-12, AC-13 | T-SPC004-06 (validateQuery danger 분기) + existing |
| REQ-SPC-004-042 | Complex (multi-filter AND-composed) | AC-14 | (existing covers; T-SPC004-15 16-lang verifies) |

Coverage: **18 unique REQs (001..007, 010..013, 020..021, 030..031, 040..042) → 15 ACs (AC-01..AC-15) → 15 tasks (T-SPC004-01..15)**. AC-01..AC-15 from spec §6 are mapped.

→ All REQ IDs from spec §5 are mapped.

---

## 2. Milestone Breakdown (M1-M6)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule).

### M1: LSP `find-references` 통합 (G-01) — Priority P0

Owner role: `expert-backend` (Go LSP client + concurrent dispatch).

Tasks:
- T-SPC004-01: 신규 `internal/mx/fanin_lsp.go` 생성. `LSPFanInCounter` struct: `Client` (powernap interface), `ProjectRoot` field. `Count(ctx, tag, projectRoot, excludeTests)` 구현 — workspace/symbol query (anchor_id) → first match → textDocument/references → result count (excludeTests 적용).
- T-SPC004-02: 신규 `internal/mx/fanin_lsp_test.go`. powernap client mock (`type mockLSPClient struct{}`) 으로 RED tests:
  - `TestLSPFanInCounter_BasicCount` — symbol 매칭 + 3 references → fan_in = 3, method = "lsp".
  - `TestLSPFanInCounter_NoSymbolMatch` — workspace/symbol 0 결과 → fan_in = 0, method = "lsp" (또는 textual fallback decision; OQ-2 결정 = textual fallback + annotation).
  - `TestLSPFanInCounter_StrictModeUnavailable` — `MOAI_MX_QUERY_STRICT=1` + LSP unavailable → `LSPRequiredError`.
  - `TestLSPFanInCounter_ExcludeTests` — references 결과 중 _test.go → 제외.

RED 검증: `TestLSPFanInCounter_*` FAIL on `go test ./internal/mx/` (LSPFanInCounter 미구현).

GREEN 검증: 4 test PASS. existing `TestTextualFanInCounter_*` (research [E-07]) still PASS — interface compat.

### M2: `mx.yaml` `danger_categories:` user wire-up + validateQuery 강화 (G-02) — Priority P1

Owner role: `expert-backend` (Go yaml unmarshal + CLI wire).

Tasks:
- T-SPC004-03: `internal/mx/danger_category.go` 보강. `LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error)` 추가:
  - mx.yaml 부재 → `DangerCategoryConfig{}` (graceful → DefaultDangerCategories).
  - parse error → graceful + log warning.
  - 정상 → user-defined Categories + TestPaths 반환.
- T-SPC004-06: `internal/mx/resolver_query.go` 의 `validateQuery()` 에 분기 추가:
  - `query.Danger != ""` 일 때 `dangerMatcher.ValidateCategory(query.Danger)` false → `InvalidQueryError{Field: "danger", Value: query.Danger, Message: "...allowed: " + KnownCategories()}`.
  - 단 `validateQuery` 시점에 dangerMatcher 가 nil 일 수 있음 — 그 경우 default matcher 로 검증.

신규 RED tests:
- `TestLoadDangerConfig_UserCustomCategories` (T-SPC004-03 RED) — mx.yaml 에 `danger_categories: { critical: ["unwrap()", "panic()"] }` → matcher 가 "unwrap()" 매칭.
- `TestLoadDangerConfig_FileMissing_DefaultUsed` (T-SPC004-03 RED) — mx.yaml 부재 → DefaultDangerCategories 사용.
- `TestValidateQuery_UnknownDanger_InvalidQueryError` (T-SPC004-06 RED) — `--danger frobnicate` → exit 2.

RED 검증: 3 test FAIL.

GREEN 검증: 3 PASS. existing `validateQuery` test (kind validation) still PASS.

### M3: `.moai/specs/*/spec.md` `module:` 자동 로드 + SpecAssociator 주입 (G-03) — Priority P1

Owner role: `expert-backend` (yaml frontmatter parse + filesystem walk).

Tasks:
- T-SPC004-04: 신규 `internal/mx/spec_loader.go`. `LoadSpecModules(projectRoot string) (map[string][]string, error)` 구현:
  - `.moai/specs/*/spec.md` glob 순회 (sort 보장).
  - 각 file 의 yaml frontmatter (`---` ... `---` 사이) parse.
  - `id` field → SPEC ID, `module` field → string (쉼표 split + TrimSpace) 또는 `[]string` (yaml array — type assertion 분기).
  - 결과 map 반환. spec.md 부재 시 빈 map.
- T-SPC004-05: 신규 `internal/mx/spec_loader_test.go`. RED tests:
  - `TestLoadSpecModules_StringFormat` — `module: "internal/mx/, cmd/moai/"` → `map["SPEC-X"] = ["internal/mx/", "cmd/moai/"]`.
  - `TestLoadSpecModules_ArrayFormat` — `module: [internal/foo/, internal/bar/]` → 동일 결과.
  - `TestLoadSpecModules_EmptyModule` — `module: ""` → `map["SPEC-X"] = []`.
  - `TestLoadSpecModules_NoSpecsDir` — `.moai/specs/` 부재 → empty map.
  - `TestSpecAssociator_PathBased_FromLoader` — loader 결과 → SpecAssociator 의 path-prefix 매칭 동작.

RED 검증: 5 test FAIL.

GREEN 검증: 5 PASS. existing `SpecAssociator_BodyBased` test (research [E-12]) still PASS.

### M4: `Resolver.ResolveAnchorCallsites()` API parity (G-04) — Priority P1

Owner role: `expert-backend` (additive API).

Tasks:
- T-SPC004-10: 신규 `internal/mx/callsite.go`. `Callsite` struct 정의:
  ```go
  type Callsite struct {
      File   string `json:"file"`
      Line   int    `json:"line"`
      Column int    `json:"column,omitempty"`
      Method string `json:"method"` // "lsp" or "textual"
  }
  ```
- T-SPC004-11: `internal/mx/resolver.go` 수정. `Resolver.ResolveAnchorCallsites(ctx, anchorID, projectRoot, includeTests) ([]Callsite, error)` 메서드 추가:
  - LSP available 시 LSPFanInCounter 의 underlying client 호출 → `Location[]` → `Callsite[]` 변환.
  - LSP unavailable 시 TextualFanInCounter 와 같은 walk + line-grep 하되 hit 마다 `Callsite{File, Line, Method: "textual"}` 추가.
  - existing `ResolveAnchor` 보존 (no signature change).
- 신규 RED tests:
  - `TestResolver_ResolveAnchorCallsites_LSP` — LSP mock → 3 callsites 반환.
  - `TestResolver_ResolveAnchorCallsites_TextualFallback` — LSP unavailable → walk-based 결과.
  - `TestResolver_ResolveAnchor_BackwardCompat` — 기존 `ResolveAnchor(anchorID)` (Tag, error) 변경 없음.

RED 검증: 3 test FAIL (Callsite struct + ResolveAnchorCallsites 미구현).

GREEN 검증: 3 PASS. existing `TestResolver_ResolveAnchor*` test (research [E-13]) still PASS.

### M5: `mx.yaml test_paths:` glob 패턴 wire-up (G-05) + stderr verification (G-06) — Priority P1

Owner role: `expert-backend`.

Tasks:
- T-SPC004-07: `internal/mx/fanin.go` 의 `isTestFile()` 변경 또는 `isTestFileWithPatterns(file string, userPatterns []string) bool` 신설. user pattern (glob) 가 매칭하면 true; hard-coded fallback (`_test.go`, `tests`, `fixtures`, `testdata`) 도 유지.
- T-SPC004-08: 신규 RED tests:
  - `TestIsTestFile_UserPattern_IntegrationDir` — patterns `["**/integration/**"]` + file `internal/foo/integration/bar.go` → true.
  - `TestIsTestFile_UserPattern_NoMatch_FallbackHardcoded` — patterns `["**/integration/**"]` + file `internal/foo/foo_test.go` → true (hard-coded `_test.go` 매칭).
  - `TestTextualFanInCounter_RespectsUserTestPaths` — counter 가 LoadDangerConfig 의 TestPaths 받아 적용.
- T-SPC004-09: 신규 `TestSidecarUnavailable_StderrFormat` (G-06):
  - sidecar 부재 + `moai mx query` 실행 → stderr 가 substring `SidecarUnavailable` AND `/moai mx --full` 둘 다 포함.
  - `internal/cli/mx_query_test.go` 확장.

RED 검증: 4 test FAIL (사용자 pattern 미적용).

GREEN 검증: 4 PASS. existing `isTestFile` test (research [E-08]) still PASS.

### M6: 16-언어 sweep (G-08) + Performance benchmark (G-07) + Verification — Priority P0

Owner role: `manager-cycle` + `manager-quality`.

Tasks:
- T-SPC004-15: 신규 `TestResolver_AllSixteenLanguages` (G-08):
  - `t.TempDir()` 안에 16개 source file 생성 (go / python / typescript / javascript / rust / java / kotlin / csharp / ruby / php / elixir / cpp / scala / r / flutter / swift), 각 1개 `@MX:NOTE` + 1개 `@MX:ANCHOR` (anchor_id 고유, e.g. `anchor-go-001`, `anchor-py-001` ...).
  - sidecar 생성 후 resolver 가 16 language 모든 tag 추출.
  - `--kind anchor` filter → 16 anchors 반환.
  - fan_in textual mode + cross-file callsite (anchor_id 가 다른 파일에서 참조됨) → fan_in = N (configurable).
  - SPEC association: tag body 에 `SPEC-V3R2-X-001` 명시 → `spec_associations` 에 포함.
- T-SPC004-13: 신규 `internal/mx/resolver_query_bench_test.go`:
  - `BenchmarkResolver_Resolve_1KTags` — 1000 tag fixture, no fan_in computation. spec §7: <100ms.
  - `BenchmarkResolver_Resolve_50AnchorsLSP` — 50 ANCHOR + LSP mock. spec §7: <2s.
  - 두 benchmark 모두 advisory; CI 강제 X.
- T-SPC004-14: 전체 테스트 스위트 — `go test -race -count=1 ./...` PASS, 0 회귀.
- T-SPC004-16: `golangci-lint run` clean.
- T-SPC004-17: `make build` exits 0; `internal/template/embedded.go` 재생성.
- T-SPC004-18: `CHANGELOG.md` Unreleased 업데이트:
  - "feat(mx/SPEC-V3R2-SPC-004): LSP `find-references` integration via powernap (LSP-backed `LSPFanInCounter`)"
  - "feat(mx/SPEC-V3R2-SPC-004): `mx.yaml` `danger_categories:` + `test_paths:` user wire-up"
  - "feat(mx/SPEC-V3R2-SPC-004): `.moai/specs/*/spec.md` `module:` frontmatter 자동 로드 + `SpecAssociator` 주입"
  - "feat(mx/SPEC-V3R2-SPC-004): `Resolver.ResolveAnchorCallsites()` API parity (additive)"
- T-SPC004-19: @MX 태그 적용 per §6.
- T-SPC004-20: 수동 검증 — 임시 Go file 에 `@MX:ANCHOR` 추가 후 `moai mx query --kind anchor --fan-in-min 1` 실행, LSP-backed result 와 textual result 비교.

Verification gate: All AC-SPC-004-01 through AC-SPC-004-15 verified per acceptance.md.

---

## 3. File-Level Modification Map

### 3.1 Files modified (existing)

| File | Lines added/changed | Purpose |
|------|---------------------|---------|
| `internal/mx/danger_category.go` | +30 LOC | `LoadDangerConfig()` helper |
| `internal/mx/resolver_query.go` | +20 LOC | `validateQuery()` danger 분기 |
| `internal/mx/resolver.go` | +50 LOC | `ResolveAnchorCallsites()` 신규 메서드 |
| `internal/mx/fanin.go` | +30 LOC | `isTestFileWithPatterns()` user glob 적용 |
| `internal/cli/mx_query.go` | +60 LOC | `LoadDangerConfig` + `LoadSpecModules` 호출 + LSP detect 분기 |
| `CHANGELOG.md` | +4 lines | Unreleased entry |

### 3.2 Files created (new)

| File | LOC (approx) | Purpose |
|------|--------------|---------|
| `internal/mx/fanin_lsp.go` | ~150 | `LSPFanInCounter` (powernap-backed) |
| `internal/mx/fanin_lsp_test.go` | ~250 | T-SPC004-01, T-SPC004-02 RED tests |
| `internal/mx/spec_loader.go` | ~80 | `LoadSpecModules()` helper |
| `internal/mx/spec_loader_test.go` | ~200 | T-SPC004-04, T-SPC004-05 RED tests (5 cases) |
| `internal/mx/callsite.go` | ~30 | `Callsite` struct + JSON tags |
| `internal/mx/resolver_callsites_test.go` | ~150 | T-SPC004-10, T-SPC004-11 RED tests (3 cases) |
| `internal/mx/resolver_query_bench_test.go` | ~80 | T-SPC004-13 advisory benchmarks |

### 3.3 Files extended (existing test files)

| File | LOC (added) | Purpose |
|------|-------------|---------|
| `internal/mx/danger_category_test.go` | +120 | T-SPC004-03 RED tests (LoadDangerConfig + UnknownDanger) |
| `internal/mx/fanin_test.go` | +100 | T-SPC004-07, T-SPC004-08 RED tests (user test_paths) |
| `internal/cli/mx_query_test.go` | +60 | T-SPC004-09 RED test (stderr format) |
| `internal/mx/resolver_query_test.go` | +200 | T-SPC004-15 16-language sweep + T-SPC004-06 validateQuery danger |

### 3.4 Files removed

None.

### 3.5 Files NOT modified (out-of-scope)

- `.claude/rules/moai/workflow/mx-tag-protocol.md` — FROZEN per CONST-V3R2-003.
- `internal/mx/sidecar.go` — SPC-002 territory (read-only consume from this SPEC).
- `internal/mx/scanner.go` — SPC-002 territory (no scanner change needed).
- `internal/mx/tag.go` — schema 변경 금지 (SPC-002 schema_version: 2 invariant).
- `internal/mx/comment_prefixes.go` — already complete (16-lang lookup).
- `internal/hook/post_tool_*.go` — SPC-002 territory.
- `internal/cli/mx.go` — parent command, SPC-002 의 G-03 격차 영역.
- `cmd/moai/main.go` — root command, 변경 불필요.

---

## 4. Technical Approach

### 4.1 LSPFanInCounter 의 powernap 호출 흐름

```go
// internal/mx/fanin_lsp.go (신규)
package mx

import (
    "context"
    "strings"

    "github.com/modu-ai/moai-adk/internal/lsp/core"
)

// LSPFanInCounter is a FanInCounter that uses LSP textDocument/references
// to count anchor references. Falls back to TextualFanInCounter when LSP
// is unavailable for the target language.
//
// @MX:ANCHOR LSPFanInCounter — LSP-first fan_in measurement
// @MX:REASON: 본 SPEC 의 spec §5.2 REQ-003 의 LSP-first 정책 단일 진입점;
//             callers: CLI mx_query.go, ResolveAnchorCallsites, codemaps generator
type LSPFanInCounter struct {
    Client      *core.Client // powernap LSP client
    ProjectRoot string
    Fallback    FanInCounter // typically *TextualFanInCounter
}

func (c *LSPFanInCounter) Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
    if tag.AnchorID == "" {
        return 0, "lsp", nil
    }
    if c.Client == nil || !c.Client.IsAvailable(ctx, langFor(tag.File)) {
        if c.Fallback != nil {
            return c.Fallback.Count(ctx, tag, projectRoot, excludeTests)
        }
        return 0, "textual", nil
    }
    // workspace/symbol query
    syms, err := c.Client.WorkspaceSymbol(ctx, tag.AnchorID)
    if err != nil || len(syms) == 0 {
        if c.Fallback != nil {
            return c.Fallback.Count(ctx, tag, projectRoot, excludeTests)
        }
        return 0, "lsp", nil
    }
    // textDocument/references on first match
    locs, err := c.Client.References(ctx, syms[0].Location, false /* includeDeclaration */)
    if err != nil {
        if c.Fallback != nil {
            return c.Fallback.Count(ctx, tag, projectRoot, excludeTests)
        }
        return 0, "lsp", err
    }
    count := 0
    for _, loc := range locs {
        if excludeTests && isTestPath(loc.URI) {
            continue
        }
        count++
    }
    return count, "lsp", nil
}

func langFor(file string) string {
    // mapping ext → LSP language id
    // delegated to internal/mx/comment_prefixes.go pattern (read-only)
    ext := strings.ToLower(filepath.Ext(file))
    return langIDForExt(ext) // helper in same package
}
```

powernap API (`core.Client.WorkspaceSymbol` / `core.Client.References`) 가 안정 표면이라 가정. 실 구현 시 powernap signature 에 맞게 조정 (research §3.1 cross-ref).

### 4.2 strictMode 분기

```go
// internal/mx/resolver_query.go (수정 — research [E-21] 강화)
strictMode := os.Getenv("MOAI_MX_QUERY_STRICT") == "1"
needsFanIn := query.FanInMin > 0
if strictMode && needsFanIn {
    if lspCounter, ok := fanInCounter.(*LSPFanInCounter); ok {
        if lspCounter.Client == nil || !lspCounter.Client.IsAnyServerAvailable(context.Background()) {
            return QueryResult{}, &LSPRequiredError{Language: "any"}
        }
    } else {
        return QueryResult{}, &LSPRequiredError{Language: "any"}
    }
}
```

기존 (E-21) 의 unconditional `LSPRequired` 반환을 LSP availability 검출 결과로 대체. Backward-compat: strictMode unset 일 때는 동작 변경 없음.

### 4.3 LoadSpecModules helper

```go
// internal/mx/spec_loader.go (신규)
package mx

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

// LoadSpecModules walks .moai/specs/*/spec.md and extracts the `module:`
// frontmatter field for each SPEC. Used by the SpecAssociator to perform
// path-based association (REQ-SPC-004-006 (a)).
//
// Supported `module:` formats:
//   - String:    module: "internal/mx/, cmd/moai/"
//   - Array:     module: [internal/mx/, cmd/moai/]
//
// Empty / missing `module:` → empty []string for that SPEC ID.
// Missing .moai/specs/ → empty map (graceful).
func LoadSpecModules(projectRoot string) (map[string][]string, error) {
    specsDir := filepath.Join(projectRoot, ".moai/specs")
    entries, err := os.ReadDir(specsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return map[string][]string{}, nil
        }
        return nil, fmt.Errorf("read specs dir: %w", err)
    }
    result := map[string][]string{}
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
        data, err := os.ReadFile(specPath)
        if err != nil {
            continue // graceful skip
        }
        front := extractFrontmatter(data)
        if front == nil {
            continue
        }
        var meta struct {
            ID     string      `yaml:"id"`
            Module interface{} `yaml:"module"` // string or []string
        }
        if err := yaml.Unmarshal(front, &meta); err != nil {
            continue
        }
        if meta.ID == "" {
            meta.ID = entry.Name() // fallback to dir name
        }
        result[meta.ID] = parseModuleField(meta.Module)
    }
    return result, nil
}

func parseModuleField(v interface{}) []string {
    switch m := v.(type) {
    case string:
        if m == "" {
            return []string{}
        }
        parts := strings.Split(m, ",")
        out := make([]string, 0, len(parts))
        for _, p := range parts {
            if t := strings.TrimSpace(p); t != "" {
                out = append(out, t)
            }
        }
        return out
    case []interface{}:
        out := make([]string, 0, len(m))
        for _, item := range m {
            if s, ok := item.(string); ok {
                if t := strings.TrimSpace(s); t != "" {
                    out = append(out, t)
                }
            }
        }
        return out
    default:
        return []string{}
    }
}

// extractFrontmatter returns the bytes between the first two "---" lines,
// or nil if no frontmatter is found.
func extractFrontmatter(data []byte) []byte {
    // implementation: first --- on line 1, find second ---, return between
    // ... (omitted for brevity)
}
```

### 4.4 LoadDangerConfig helper

```go
// internal/mx/danger_category.go (수정 — append helper)
func LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error) {
    path := filepath.Join(projectRoot, ".moai/config/sections/mx.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return DangerCategoryConfig{}, nil // graceful → DefaultDangerCategories
        }
        return DangerCategoryConfig{}, err
    }
    var cfg DangerCategoryConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return DangerCategoryConfig{}, nil // graceful (log warning)
    }
    return cfg, nil
}
```

### 4.5 isTestFile glob 패턴 wire-up

```go
// internal/mx/fanin.go (수정)
func isTestFileWithPatterns(filePath string, userPatterns []string) bool {
    // hard-coded fallback (preserve existing behavior)
    if isTestFile(filePath) {
        return true
    }
    // user-defined glob patterns (gitignore-style)
    for _, pattern := range userPatterns {
        if match, _ := doublestar.PathMatch(pattern, filePath); match {
            return true
        }
    }
    return false
}
```

doublestar 라이브러리는 이미 repo 의 go.mod 에 있을 가능성 — 검증 후 부재 시 `path/filepath.Match` 의 single-star 패턴만 지원 (간소화).

### 4.6 Callsite struct + ResolveAnchorCallsites

```go
// internal/mx/callsite.go (신규)
package mx

type Callsite struct {
    File   string `json:"file"`
    Line   int    `json:"line"`
    Column int    `json:"column,omitempty"`
    Method string `json:"method"` // "lsp" or "textual"
}

// internal/mx/resolver.go (수정 — additive method)
func (r *Resolver) ResolveAnchorCallsites(ctx context.Context, anchorID, projectRoot string, includeTests bool) ([]Callsite, error) {
    tag, err := r.ResolveAnchor(anchorID)
    if err != nil {
        return nil, err
    }
    counter := r.fanInCounter()
    if lspCounter, ok := counter.(*LSPFanInCounter); ok && lspCounter.Client != nil {
        // LSP path
        return resolveCallsitesViaLSP(ctx, lspCounter, tag, projectRoot, includeTests)
    }
    // textual fallback
    return resolveCallsitesViaTextual(tag, projectRoot, includeTests)
}
```

`r.fanInCounter()` 는 Resolver 에 새 필드 (`fanInCounter FanInCounter` private field) + getter 추가.

### 4.7 Performance budget verification 형식

```go
// internal/mx/resolver_query_bench_test.go (신규)
func BenchmarkResolver_Resolve_1KTags(b *testing.B) {
    tmpDir := b.TempDir()
    mgr := NewManager(tmpDir)
    sidecar := generateBenchmarkSidecar(1000, 0 /* no anchors with fan_in */)
    if err := mgr.Write(sidecar); err != nil {
        b.Fatal(err)
    }
    resolver := NewResolver(mgr)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := resolver.Resolve(Query{Limit: 100})
        if err != nil {
            b.Fatal(err)
        }
    }
    // Advisory: <100ms per op (spec §7)
}

func BenchmarkResolver_Resolve_50AnchorsLSP(b *testing.B) {
    // 50 ANCHOR fixture + LSP mock counter
    // Advisory: <2s per op
}
```

### 4.8 Backward compatibility

- `Resolver.ResolveAnchor(anchorID) (Tag, error)` — signature 보존 (G-04 는 additive `ResolveAnchorCallsites`).
- `TextualFanInCounter` — 변경 없음. `LSPFanInCounter` 가 신규 인터페이스 구현체로 추가.
- `DangerCategoryMatcher` — `NewDangerCategoryMatcher(cfg)` 가 cfg.Categories 비어있으면 DefaultDangerCategories 사용 (existing 동작 유지).
- `validateQuery()` — danger 분기 추가 (existing kind 분기 보존).
- CLI 의 stdout 출력 형식 (slice-only JSON) — 변경 없음 (OQ-5 결정).
- sidecar schema_version: 2 — 변경 금지 (SPC-002 invariant).

### 4.9 Cross-platform behavior

- `os.ReadDir` / `os.ReadFile` — cross-platform.
- `filepath.Match` / `doublestar.PathMatch` — case-sensitive on Linux/Mac, case-insensitive on Windows (acceptable; user 가 패턴 작성 시 인지).
- `MOAI_MX_QUERY_STRICT` env — cross-platform.
- powernap LSP client — Windows 에서 일부 LSP server (예: rust-analyzer) 가 다른 binary path 가질 수 있음 (powernap 측 책임).

---

## 5. Quality Gates

Per `.moai/config/sections/quality.yaml`:

| Gate | Requirement | Verification command |
|------|-------------|-----------------------|
| Coverage | ≥ 85% per modified file | `go test -cover ./internal/mx/ ./internal/cli/` |
| Race | `go test -race ./...` clean | `go test -race -count=1 ./...` |
| Lint | golangci-lint clean | `golangci-lint run` |
| Build | embedded.go regenerated | `make build` |
| Template parity | `diff -r` byte-identical | `diff -r .claude/ internal/template/templates/.claude/` |
| Sidecar contract | schema_version: 2 보존 (SPC-002 invariant) | T-SPC004-15 16-lang sweep verifies sidecar Load OK |
| Performance budget | benchmark advisory | `go test -bench BenchmarkResolver_Resolve ./internal/mx/` |
| MX | @MX tags applied per §6 | manual (post-implementation review) |

---

## 6. @MX Tag Plan (mx_plan)

Apply per `.claude/rules/moai/workflow/mx-tag-protocol.md`. Language: en (per `code_comments: ko` — 한국어, 단 MX tag 본문은 영문 per `feedback_mx_tag_language.md`). 본 plan 자체는 plan markdown 이므로 @MX 태그를 추가하지 않음 (이는 source code 대상). 아래 표는 run-phase 에서 Go 코드에 추가될 태그 (`code_comments: ko` per language.yaml — single-line MX tag is in English by convention).

| File | Tag | Reason |
|------|-----|--------|
| `internal/mx/fanin_lsp.go:LSPFanInCounter.Count` | `@MX:ANCHOR @MX:REASON: SPEC-V3R2-SPC-004 의 LSP-first fan_in 단일 진입점; fan_in ≥ 3 (CLI mx_query.go + ResolveAnchorCallsites + future codemaps generator)` | invariant contract + high fan_in |
| `internal/mx/fanin_lsp.go:LSPFanInCounter` (struct) | `@MX:NOTE: powernap client 의존; `Fallback` 필드 nil 시 LSP 실패가 silent 0-count 으로 transitionm — caller 가 strictMode 분기로 명시 처리해야 함` | non-obvious behavior |
| `internal/mx/spec_loader.go:LoadSpecModules` | `@MX:NOTE: spec.md frontmatter 의 module 필드는 string 또는 []string 양쪽 형식 허용; runtime type assertion 으로 분기. 추가 형식 (예: nested map) 은 silent ignore — graceful degradation 정책` | parsing 분기 사유 |
| `internal/mx/danger_category.go:LoadDangerConfig` | `@MX:NOTE: mx.yaml 부재 시 graceful — DefaultDangerCategories 사용; CI 환경 isolation 보장 (config 부재가 lint break 가 아님)` | graceful fallback rationale |
| `internal/mx/resolver_query.go:validateQuery` (수정 영역) | `@MX:WARN @MX:REASON: --danger 검증 시 dangerMatcher nil 가능성 — default matcher 로 fallback. 호출자가 dangerMatcher 를 항상 주입한다고 가정하는 invariant 위반 시 silent allow` | mutability hazard |
| `internal/mx/resolver.go:ResolveAnchorCallsites` | `@MX:WARN @MX:REASON: LSP 실패 시 textual fallback 으로 단일 anchor 가 두 가지 fan_in_method 결과를 반환 가능 — caller 가 method 필드를 inspect 하지 않으면 결과 의미가 모호. 명시적 method 필드로 해결` | API surface hazard |

---

## 7. Risk Mitigation Plan (spec §8 risks → run-phase tasks)

| spec §8 risk | Mitigation in run-phase |
|--------------|--------------------------|
| Row 1 — Fan_in false positives (textual matches) | T-SPC004-01 LSP-first; T-SPC004-02 method annotation; doc 명시 |
| Row 2 — `danger_categories:` patterns over-match | T-SPC004-03 user customizable; T-SPC004-06 InvalidQuery 검증; defaults conservative (4 카테고리만) |
| Row 3 — SPEC association heuristic misses tags | T-SPC004-04, T-SPC004-05 path-based wire; body-based always works (research [E-12]); `--spec none` 잠재 추가 (out-of-scope, advisory) |
| Row 4 — Output size blows up | already mitigated (--limit default 100, max 10000 — research [E-06]); T-SPC004-13 benchmark verifies pagination |
| Row 5 — Resolver bypasses SPC-002 freshness | sidecar age check 잠재 추가 (out-of-scope; SPC-002 plan §7 OQ 와 cross-ref) |

추가 mitigation:
- **OOQ — LSP server 가 16-언어 중 일부에서 부재** (R 등): T-SPC004-01 의 fallback 분기로 silent textual transition + annotation.
- **OOQ — workspace/symbol query 가 anchor_id 매칭 실패**: G-01 의 fallback 분기 (LSP → textual).
- **OOQ — spec.md frontmatter 의 module 필드가 nested 형식**: T-SPC004-04 의 graceful skip (`default: []string{}`).
- **OOQ — TextualFanInCounter 의 filepath.Walk 가 50 anchor × 10K-file 시 O(N²)**: T-SPC004-13 benchmark 가 회귀 detect; advisory cache 도입은 v3.1.

---

## 8. Dependencies (status as of `73742e3ee`)

### 8.1 Blocking (consumed)

- ✅ **SPEC-V3R2-SPC-002** (sidecar contract) — plan PR #836 머지; sidecar 구현은 PR #741 commit `3f0933550` 에서 완료. schema_version: 2 invariant 가정.
- ✅ **SPEC-LSP-CORE-002** (LSP client via powernap) — assumed merged; `internal/lsp/core/client.go` 가 powernap-based client 노출.

### 8.2 Blocked by (none active)

All blockers are merged or assumed merged. powernap API 변경 시 plan 재조정 필요.

### 8.3 Blocks (downstream consumers)

- **codemaps generation tools** — per-module anchor count 를 본 SPEC 의 `Resolver.Resolve(query)` 로 조회 가능.
- **SPEC-V3R2-HRN-003** (evaluator MX scoring) — danger_category distribution 을 harness signal 로 사용 가능.
- **`/moai review` phase** — 잠재 consumer.

---

## 9. Verification Plan

### 9.1 Pre-merge verification (run-phase end)

- [ ] All 20 tasks (T-SPC004-01..20) complete per tasks.md
- [ ] All 15 ACs (AC-SPC-004-01..15) verified per acceptance.md
- [ ] `go test -race -count=1 ./...` PASS (no regressions)
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] Coverage `go test -cover ./internal/mx/ ./internal/cli/` ≥ 85% per package
- [ ] `diff -r .claude/ internal/template/templates/.claude/` byte-identical (no template change expected)
- [ ] CHANGELOG entry written in Unreleased section
- [ ] @MX tags applied per §6 (6 tags)
- [ ] Manual verification: `moai mx query --kind anchor --fan-in-min 1` against real project, verify LSP path 동작 (gopls available)

### 9.2 Plan-auditor target

- [ ] All 18 unique REQs mapped in §1.5 traceability matrix
- [ ] All 15 ACs mapped to ≥1 task
- [ ] No orphan tasks (every task supports ≥1 REQ)
- [ ] research.md evidence anchors cited (≥30 per plan-audit mandate; current: 45)
- [ ] §1.2.1 explicitly addresses spec/plan discrepancies (90%+ already implemented; ResolveAnchor signature; JSON envelope; danger 검증 격차; LSP detect; benchmark 형식)
- [ ] FROZEN clause CONST-V3R2-003 explicitly preserved (mx-tag-protocol.md untouched)
- [ ] Sidecar schema_version: 2 invariant 보존 명시
- [ ] Worktree-base alignment per Step 2 (run-phase) called out (§10 below)
- [ ] §6 mx_plan covers ≥3 of {ANCHOR, WARN, NOTE} types (covered: 1 ANCHOR + 2 WARN + 3 NOTE = 모두 3)
- [ ] No time estimates (P0/P1 priority labels only)
- [ ] Parallel SPEC isolation: only `internal/mx/`, `internal/cli/`, `CHANGELOG.md`. .claude/ tree 미변경.

### 9.3 Plan-auditor risk areas (front-loaded mitigations)

- **Risk: SPC-004 가 이미 90%+ 구현 → REQ-001..007 이 redundant 처럼 보임** → §1.2.1 acknowledged; sync-phase HISTORY 에서 spec.md "Expose ... API" → "complete the API to spec parity" reconcile. plan 의 task 는 격차 8개 (G-01..G-08) 에 집중.
- **Risk: SPC-002 가 schema 를 변경하면 본 SPEC break 가능** → SPC-002 plan §1.4 의 schema_version: 2 invariant 보존 서약 cross-ref. sidecar Load 호환성은 본 SPEC 의 책임.
- **Risk: powernap API 변경** → assumption (research §8.2). 변경 시 `LSPFanInCounter` 의 client signature 만 조정.
- **Risk: 16-언어 LSP server 가 사용자 환경에서 일부 부재** → fallback 정책 (G-01 OQ-1). silent textual transition + annotation으로 graceful.
- **Risk: spec.md `Resolver.ResolveAnchor(anchorID) []Callsite` verbatim 과 현 구현 mismatch** → §1.2.1 acknowledged; G-04 가 additive 로 해결.
- **Risk: `73742e3ee` baseline drift if main advances during plan PR review** → run-phase 가 명시적으로 `origin/main` 에서 rebase (Step 2 `moai worktree new --base origin/main`).
- **Risk: G-01 LSP 통합이 powernap 의 안정 API 가 없는 경우 stuck** → research §3.1 의 powernap reference 가 가정. plan 진입 전 powernap signature 검증 필요. 위험 LOW (이미 다른 hook 에서 사용 중).
- **Risk: T-SPC004-15 의 16-language fixture 가 LSP 비의존이라 G-01 검증 부족** → T-SPC004-01 + T-SPC004-02 의 LSP mock test 가 직교 검증.

### 9.4 Rollback plan (if LSP integration causes regressions)

만약 run-phase 후 LSP 통합 (G-01) 이 unforeseen regressions 을 일으키면:
1. CLI 의 LSP detect 분기를 unconditional `TextualFanInCounter` 로 fallback (한 줄 변경).
2. `LSPFanInCounter` 코드는 보존 (`internal/mx/fanin_lsp.go` deleted X — 단 사용 site 비활성화).
3. Hotfix SPEC 으로 LSP path 점진 재활성화.

만약 G-03 의 SpecAssociator wire-up 이 false-positive (잘못된 SPEC 매칭) 을 양산하면:
1. CLI 가 `LoadSpecModules` 호출을 skip → empty map 으로 SpecAssociator 초기화.
2. body-based association (REQ-006 (b)) 만 동작 — existing 동작과 동일 (research [E-12]).

---

## 10. Run-Phase Entry Conditions

After plan PR squash-merged into main:

1. `git checkout main && git pull` (host checkout).
2. `moai worktree new SPEC-V3R2-SPC-004 --base origin/main` per Step 2 spec-workflow.md.
3. `cd ~/.moai/worktrees/moai-adk/SPEC-V3R2-SPC-004`.
4. `git rev-parse --show-toplevel` should output the worktree path (Block 0 verification per session-handoff.md).
5. `git rev-parse HEAD` should match plan-merge commit SHA on main.
6. Verify SPC-002 sidecar contract 무결성: `go test ./internal/mx/...` — 기존 SPC-002 + SPC-004 tests still PASS.
7. Verify powernap LSP client 가용: `go test ./internal/lsp/core/...` PASS.
8. `/moai run SPEC-V3R2-SPC-004` invokes Phase 0.5 plan-audit gate, then proceeds to M1.

---

Version: 0.1.0
Status: Plan artifact for SPEC-V3R2-SPC-004
Run-phase methodology: TDD (per `.moai/config/sections/quality.yaml` `development_mode: tdd`)
Estimated artifacts: 7 new Go source files + extensions to 4 existing test files + 5 modified Go files + CHANGELOG = ~1,400 LOC delta (production ~480 + test ~920)
