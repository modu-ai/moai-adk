# SPEC-V3R2-SPC-004 Task List

> Implementation task list for **@MX anchor resolver (query by SPEC ID, fan_in, danger category)**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                                    | Description |
|---------|------------|-------------------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B)     | Initial task list. 20 tasks (T-SPC004-01..20) grouped into 6 milestones (M1..M6), priorities P0/P1. Wave-split: tasks fit within ~20 entries — no wave-split needed (under feedback_large_spec_wave_split.md 30-task threshold). |

---

## 1. Task Overview

| Milestone | Phase | Tasks | Priority | REQ Coverage |
|---|---|---|---|---|
| M1: LSP `find-references` integration (G-01) | RED → GREEN | T-SPC004-01, T-SPC004-02 | P0 | REQ-003, REQ-011, REQ-020, REQ-030 |
| M2: `mx.yaml` `danger_categories:` user wire-up + validateQuery 강화 (G-02) | RED → GREEN | T-SPC004-03, T-SPC004-06 | P1 | REQ-001 (`--danger`), REQ-012, REQ-041 |
| M3: `.moai/specs/*/spec.md` `module:` 자동 로드 (G-03) | RED → GREEN | T-SPC004-04, T-SPC004-05 | P1 | REQ-006, REQ-010 |
| M4: `Resolver.ResolveAnchorCallsites()` API parity (G-04) | RED → GREEN | T-SPC004-10, T-SPC004-11 | P1 | REQ-002, REQ-003 |
| M5: test_paths glob (G-05) + stderr verify (G-06) | RED → GREEN | T-SPC004-07, T-SPC004-08, T-SPC004-09 | P1 | REQ-013, REQ-040 |
| M6: 16-언어 sweep (G-08) + benchmark (G-07) + verify | VERIFY | T-SPC004-12..20 | P0 | (cross-cutting) |

Total: **20 tasks**. Below 30-task threshold → no wave-split required (per `feedback_large_spec_wave_split.md`).

---

## 2. Tasks by Milestone

### Milestone 1: LSP `find-references` 통합 (G-01) — Priority P0

#### T-SPC004-01: Create `internal/mx/fanin_lsp.go` with LSPFanInCounter

**REQ traceback**: REQ-SPC-004-003 (LSP-first), REQ-SPC-004-020 (textual fallback annotation)

**AC traceback**: AC-02, AC-07

**Goal**: Create new `LSPFanInCounter` struct implementing the `FanInCounter` interface (existing `internal/mx/fanin.go` line 13-18). Uses powernap LSP client for `workspace/symbol` + `textDocument/references` queries. Falls back to `TextualFanInCounter` when LSP server unavailable for the target language.

**Files**:
- `internal/mx/fanin_lsp.go` (new ~150 LOC)

**Sub-tasks**:
- Define `LSPFanInCounter` struct: `Client *core.Client` (powernap), `ProjectRoot string`, `Fallback FanInCounter`
- Implement `Count(ctx, tag, projectRoot, excludeTests) (int, string, error)` per plan §4.1 pseudocode
- Add `langFor(file string) string` helper (extension → LSP language id; reuse comment_prefixes lookup pattern but for LSP language identifier strings)
- Add `isTestPath(uri string) bool` helper (URI → file path → call existing `isTestFile`; M5 추후 extension)
- Mock-friendly: `Client` is interface (or pointer to powernap struct that has interface methods); test injects `*mockLSPClient`
- Returns (count, "lsp", nil) on success; (count, "textual", nil) on Fallback path; (0, "lsp", err) on hard failure with no fallback

**Acceptance**:
- [ ] Compiles cleanly (`go build ./internal/mx/`)
- [ ] No cyclic import (verify `go vet ./...`)
- [ ] `LSPFanInCounter` satisfies `FanInCounter` interface (compile-time assertion `var _ FanInCounter = (*LSPFanInCounter)(nil)`)

**Dependencies**: powernap `core.Client` exposed (research §3.1)

---

#### T-SPC004-02: LSPFanInCounter RED tests + strictMode wire-up

**REQ traceback**: REQ-SPC-004-003, REQ-SPC-004-020, REQ-SPC-004-030

**AC traceback**: AC-02, AC-07, AC-09

**Goal**: Add RED tests for `LSPFanInCounter` covering:
1. Basic LSP-backed count (workspace/symbol → references → count)
2. No symbol match → fallback path (textual annotation)
3. Strict mode + LSP unavailable → `LSPRequiredError`
4. Test file exclusion (excludeTests=true)

Also: wire up `internal/mx/resolver_query.go` strictMode 분기 (research [E-21] line 195-202) to actually check LSP availability via the injected counter (currently returns unconditional `LSPRequired` — this is a real bug today).

**Files**:
- `internal/mx/fanin_lsp_test.go` (new ~250 LOC)
- `internal/mx/resolver_query.go` (modified +20 LOC for strictMode 강화)

**Sub-tasks**:
- Mock `core.Client`-shaped interface (or use a small internal interface `LSPClient` extracted from powernap surface used by counter)
- `TestLSPFanInCounter_BasicCount`: mock `WorkspaceSymbol` returns 1 symbol; mock `References` returns 3 locations → assert `(3, "lsp", nil)`
- `TestLSPFanInCounter_NoSymbolMatch_FallbackToTextual`: mock returns empty WorkspaceSymbol; counter has Fallback=*TextualFanInCounter → assert annotation = "textual"
- `TestLSPFanInCounter_StrictModeUnavailable`: `MOAI_MX_QUERY_STRICT=1` set via `t.Setenv` (note: parallel-safe — see CLAUDE.local.md §6 OTEL caveat is OTEL-specific, env via t.Setenv on this var is OK), Client.IsAvailable returns false → Resolve returns `*LSPRequiredError`
- `TestLSPFanInCounter_ExcludeTests`: References returns 5 locations of which 2 are `_test.go` → `excludeTests=true` returns count=3
- Modify `Resolver.Resolve` strictMode 분기 to call `lspCounter.Client.IsAnyServerAvailable(ctx)` (or equivalent) before returning `LSPRequiredError`

**Acceptance**:
- [ ] 4 LSPFanInCounter tests PASS
- [ ] AC-09 fixture (`TestResolver_AC9_StrictModeNoLSP` existing) still PASS with new strictMode logic
- [ ] No regression in existing `TestTextualFanInCounter_*` tests (research [E-07])

**Dependencies**: T-SPC004-01

---

### Milestone 2: `mx.yaml` `danger_categories:` user wire-up + validateQuery 강화 (G-02) — Priority P1

#### T-SPC004-03: Add LoadDangerConfig helper + RED tests

**REQ traceback**: REQ-SPC-004-001 (`--danger`), REQ-SPC-004-012

**AC traceback**: AC-03

**Goal**: Add `LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error)` to `internal/mx/danger_category.go` (research §6.2 pseudocode). Reads `.moai/config/sections/mx.yaml`, returns `DangerCategoryConfig{}` on file-missing (graceful), graceful skip on parse error. CLI wire-up call site: `internal/cli/mx_query.go` (T-SPC004-12).

**Files**:
- `internal/mx/danger_category.go` (modified +30 LOC)
- `internal/mx/danger_category_test.go` (extended +120 LOC)

**Sub-tasks**:
- Implement `LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error)`
- `TestLoadDangerConfig_FileMissing_DefaultUsed`: mx.yaml 부재 → returned cfg.Categories empty → `NewDangerCategoryMatcher(cfg)` falls back to DefaultDangerCategories
- `TestLoadDangerConfig_UserCustomCategories`: mx.yaml with `danger_categories: {critical: ["unwrap()", "panic()"]}` → matcher.Match("function calls unwrap()", "critical") returns true
- `TestLoadDangerConfig_ParseError_GracefulFallback`: malformed yaml → no panic, returns empty cfg
- `TestLoadDangerConfig_RespectsTestPaths`: `test_paths: ["**/integration/**"]` → cfg.TestPaths populated correctly

**Acceptance**:
- [ ] 4 sub-tests PASS
- [ ] Existing `TestDangerCategoryMatcher_*` tests still PASS
- [ ] `golangci-lint` clean

**Dependencies**: None (independent helper)

---

#### T-SPC004-06: validateQuery danger 분기 + RED test

**REQ traceback**: REQ-SPC-004-041

**AC traceback**: AC-13

**Goal**: Extend `validateQuery()` in `internal/mx/resolver_query.go` (research [E-15] line 265-275) to validate `query.Danger`. If user passes `--danger frobnicate` (not in known categories), return `*InvalidQueryError`. CLI converts to exit code 2.

**Files**:
- `internal/mx/resolver_query.go` (modified +20 LOC inside `validateQuery`)
- `internal/mx/resolver_query_test.go` (extended +60 LOC for new test)
- `internal/cli/mx_query.go` (modified +10 LOC to ensure exit code 2 on `*InvalidQueryError`)
- `internal/cli/mx_query_test.go` (extended +40 LOC for CLI exit-2 assertion)

**Sub-tasks**:
- In `validateQuery(query Query) error`: if `query.Danger != ""`, check `query.dangerMatcher.ValidateCategory(query.Danger)`; if false, return `&InvalidQueryError{Field: "danger", Value: query.Danger, Message: "allowed: " + strings.Join(matcher.KnownCategories(), ", ")}`
- Handle dangerMatcher nil: instantiate default matcher for validation purposes only
- `TestValidateQuery_UnknownDanger_InvalidQueryError` (Go API level)
- `TestMxQueryCmd_AC13_DangerInvalid_Exit2` (CLI level — ensure exit code is exactly 2, not 1)
- Update CLI `RunE` to map `*InvalidQueryError` to `cobra.SilenceErrors=true` + `os.Exit(2)` path (verify cobra mechanics)

**Acceptance**:
- [ ] Both new tests PASS
- [ ] Existing AC-13 fixture (`TestResolver_AC13_InvalidQuery_Kind` line 674) still PASS — kind validation unchanged
- [ ] CLI exit code is 2 on InvalidQuery (not 1)

**Dependencies**: T-SPC004-03 (DangerCategoryMatcher loaded via LoadDangerConfig)

---

### Milestone 3: `.moai/specs/*/spec.md` `module:` 자동 로드 (G-03) — Priority P1

#### T-SPC004-04: Create `internal/mx/spec_loader.go` with LoadSpecModules

**REQ traceback**: REQ-SPC-004-006 (a)

**AC traceback**: AC-01, AC-15

**Goal**: New file `internal/mx/spec_loader.go` per plan §4.3 pseudocode. Walks `.moai/specs/*/spec.md`, parses yaml frontmatter, extracts `id` + `module` fields, returns `map[specID][]modulePath`. Supports both string format (`module: "a, b"`) and array format (`module: [a, b]`).

**Files**:
- `internal/mx/spec_loader.go` (new ~80 LOC)

**Sub-tasks**:
- Implement `LoadSpecModules(projectRoot string) (map[string][]string, error)`
- Implement `extractFrontmatter(data []byte) []byte` helper (find `---` ... `---` block)
- Implement `parseModuleField(v interface{}) []string` (string vs []interface{} branch)
- Use `gopkg.in/yaml.v3` (verify in go.mod; fallback to v2 if needed)
- Graceful: missing dir → empty map, missing spec.md → skip, parse error → skip
- `id` fallback: if frontmatter missing `id`, use directory name

**Acceptance**:
- [ ] Compiles cleanly
- [ ] No new top-level go.mod dependency (yaml.v3 likely already present; verify)
- [ ] `go vet` clean

**Dependencies**: None

---

#### T-SPC004-05: spec_loader RED tests + SpecAssociator wire integration

**REQ traceback**: REQ-SPC-004-006

**AC traceback**: AC-01, AC-15

**Goal**: Add 5 RED tests for `LoadSpecModules` and 1 integration test verifying that `SpecAssociator` populated by loader correctly performs path-based association.

**Files**:
- `internal/mx/spec_loader_test.go` (new ~200 LOC)

**Sub-tasks**:
- `TestLoadSpecModules_StringFormat`: synthesize `t.TempDir()/.moai/specs/SPEC-X-001/spec.md` with `module: "internal/mx/, cmd/moai/"` → assert `result["SPEC-X-001"] == ["internal/mx/", "cmd/moai/"]`
- `TestLoadSpecModules_ArrayFormat`: yaml array `module: [internal/foo/, internal/bar/]` → same expected
- `TestLoadSpecModules_EmptyModule`: `module: ""` or absent → `result["SPEC-X-001"] == []`
- `TestLoadSpecModules_NoSpecsDir`: no `.moai/specs/` → empty map, no error
- `TestLoadSpecModules_MultipleSpecs`: 3 specs → 3-entry map
- `TestSpecAssociator_PathBased_FromLoader` (integration): build SpecAssociator from loader output → tag with file `internal/mx/foo.go` → Associate returns `["SPEC-X-001"]`

**Acceptance**:
- [ ] 6 sub-tests PASS
- [ ] Existing `TestSpecAssociator_*` tests (research [E-12]) still PASS

**Dependencies**: T-SPC004-04

---

### Milestone 4: `Resolver.ResolveAnchorCallsites()` API parity (G-04) — Priority P1

#### T-SPC004-10: Create `internal/mx/callsite.go` with Callsite struct

**REQ traceback**: REQ-SPC-004-002, REQ-SPC-004-003

**AC traceback**: (additive — supports AC-02 and downstream consumer needs)

**Goal**: New file `internal/mx/callsite.go` defining the `Callsite` struct (research §9 OQ-9). 30 LOC including JSON tags and small helper.

**Files**:
- `internal/mx/callsite.go` (new ~30 LOC)

**Sub-tasks**:
- Define `Callsite` struct with fields: `File string`, `Line int`, `Column int (omitempty)`, `Method string` ("lsp" or "textual")
- Add `String() string` method for log/debug output
- Add `IsTestFile() bool` method delegating to `isTestFile(c.File)`

**Acceptance**:
- [ ] Compiles cleanly
- [ ] JSON marshal produces expected schema

**Dependencies**: None

---

#### T-SPC004-11: Add Resolver.ResolveAnchorCallsites + RED tests

**REQ traceback**: REQ-SPC-004-002, REQ-SPC-004-003

**AC traceback**: (additive)

**Goal**: Add `ResolveAnchorCallsites(ctx, anchorID, projectRoot, includeTests) ([]Callsite, error)` method to `Resolver` (existing `internal/mx/resolver.go`). Existing `ResolveAnchor(anchorID) (Tag, error)` preserved unchanged.

**Files**:
- `internal/mx/resolver.go` (modified +50 LOC)
- `internal/mx/resolver_callsites_test.go` (new ~150 LOC)

**Sub-tasks**:
- Add private `fanInCounter FanInCounter` field to `Resolver` + setter `WithFanInCounter(c FanInCounter) *Resolver` (builder pattern)
- Implement `ResolveAnchorCallsites`: lookup tag via existing `ResolveAnchor` → if LSP counter available, call LSP path; else textual walk
- LSP path: invoke `LSPFanInCounter.Client.WorkspaceSymbol` + `References` → convert `Location[]` to `[]Callsite`
- Textual path: filepath.Walk + line-grep, on hit append `Callsite{File, Line, Method: "textual"}`
- `TestResolver_ResolveAnchorCallsites_LSP`: mock LSP client returns 3 locations → 3 Callsites
- `TestResolver_ResolveAnchorCallsites_TextualFallback`: no LSP → walk-based 3 callsites with method="textual"
- `TestResolver_ResolveAnchorCallsites_ExcludeTests`: includeTests=false → test files excluded
- `TestResolver_ResolveAnchor_BackwardCompat`: existing API signature unchanged (compile-time check `var _ func(string) (Tag, error) = (*Resolver).ResolveAnchor`)

**Acceptance**:
- [ ] 4 sub-tests PASS
- [ ] Existing `TestResolver_ResolveAnchor_*` (research [E-13]) still PASS
- [ ] No signature change on `ResolveAnchor` (compile-time)

**Dependencies**: T-SPC004-10 (Callsite struct), T-SPC004-01 (LSPFanInCounter for LSP path test)

---

### Milestone 5: test_paths glob (G-05) + stderr verify (G-06) — Priority P1

#### T-SPC004-07: isTestFile glob 패턴 wire-up

**REQ traceback**: REQ-SPC-004-040

**AC traceback**: AC-11

**Goal**: Extend `isTestFile()` (research [E-08]) to also accept user-defined glob patterns from `mx.yaml` `test_paths:`. Implement `isTestFileWithPatterns(filePath string, userPatterns []string) bool` helper. Existing `isTestFile` callers continue to use hard-coded fallback.

**Files**:
- `internal/mx/fanin.go` (modified +30 LOC)

**Sub-tasks**:
- Add helper function `isTestFileWithPatterns(filePath string, userPatterns []string) bool` — first check existing `isTestFile`, then iterate userPatterns calling `path/filepath.Match` (or `doublestar.PathMatch` if available in go.mod)
- Verify go.mod for `github.com/bmatcuk/doublestar` or similar — if absent, use `filepath.Match` (single-star) for v0.1.0 simplicity
- Refactor `TextualFanInCounter.Count` to accept user test_paths via constructor: `NewTextualFanInCounter(projectRoot string, testPaths []string)` — backward compat: existing `&TextualFanInCounter{}` still works (testPaths nil → fallback hard-coded only)

**Acceptance**:
- [ ] Compiles cleanly
- [ ] Existing `TestTextualFanInCounter_*` tests still PASS (backward compat)

**Dependencies**: None

---

#### T-SPC004-08: isTestFile glob RED tests + integration

**REQ traceback**: REQ-SPC-004-040

**AC traceback**: AC-11

**Goal**: Add 3 RED tests for new helper + integration with TextualFanInCounter.

**Files**:
- `internal/mx/fanin_test.go` (extended +100 LOC)

**Sub-tasks**:
- `TestIsTestFile_UserPattern_IntegrationDir`: patterns=`["**/integration/**"]` + file=`internal/foo/integration/bar.go` → true
- `TestIsTestFile_UserPattern_NoMatch_FallbackHardcoded`: patterns=`["**/integration/**"]` + file=`internal/foo/foo_test.go` → true (hard-coded `_test.go` matches)
- `TestTextualFanInCounter_RespectsUserTestPaths`: build counter with `testPaths=["**/integration/**"]`, anchor referenced in `internal/foo.go` (1) + `internal/integration/foo_int.go` (1); excludeTests=true → count=1 (integration excluded)

**Acceptance**:
- [ ] 3 sub-tests PASS
- [ ] Existing `isTestFile` tests still PASS

**Dependencies**: T-SPC004-07

---

#### T-SPC004-09: stderr format regression fixture (G-06)

**REQ traceback**: REQ-SPC-004-013

**AC traceback**: AC-04

**Goal**: Add explicit fixture asserting that `SidecarUnavailable` stderr message contains BOTH the substring `SidecarUnavailable` AND `/moai mx --full`. Closes the AC-04 verification gap (current test asserts substring only individually, not jointly).

**Files**:
- `internal/cli/mx_query_test.go` (extended +60 LOC)

**Sub-tasks**:
- `TestSidecarUnavailable_StderrFormat`: setup `t.TempDir()` project without `.moai/state/mx-index.json`; cd into it; invoke cobra cmd; capture stderr buffer; assert both substrings present in single capture
- Edge case: assert exact format `SidecarUnavailable: ... /moai mx --full` (per `internal/cli/mx_query.go:99-103` current emission); allow forward compatibility via substring match

**Acceptance**:
- [ ] Test PASS
- [ ] Existing `TestMxQueryCmd_AC4_SidecarUnavailable` (line 89) still PASS

**Dependencies**: None (CLI side)

---

### Milestone 6: 16-언어 sweep (G-08) + Performance benchmark (G-07) + verification — Priority P0

#### T-SPC004-12: CLI wire-up — LoadDangerConfig + LoadSpecModules + LSP detect

**REQ traceback**: REQ-SPC-004-001, REQ-SPC-004-006, REQ-SPC-004-012, REQ-SPC-004-013, REQ-SPC-004-030

**AC traceback**: AC-01, AC-03, AC-09 (cross-cutting)

**Goal**: Modify `internal/cli/mx_query.go` `RunE` to:
1. Call `mx.LoadDangerConfig(projectRoot)` → instantiate `DangerCategoryMatcher` and inject into Query
2. Call `mx.LoadSpecModules(projectRoot)` → instantiate `SpecAssociator` and inject into Query
3. Detect LSP availability (powernap server discovery) → choose `LSPFanInCounter` or `TextualFanInCounter`
4. Wire results into `Resolver.Resolve(query)`

**Files**:
- `internal/cli/mx_query.go` (modified +60 LOC)

**Sub-tasks**:
- Add helpers: `loadResolverDeps(projectRoot string) (*mx.DangerCategoryMatcher, *mx.SpecAssociator, mx.FanInCounter, error)`
- LSP detect: call `core.NewClientFromEnv()` (or whatever powernap exposes); on error → fallback `&mx.TextualFanInCounter{ProjectRoot: projectRoot}`
- Inject all three into `mx.Query` (extend Query struct exposure or use builder methods on Resolver)
- Verify CLI tests still pass

**Acceptance**:
- [ ] CLI compiles cleanly
- [ ] All 4 P0 ACs (01, 03, 04, 05, 08, 12) still PASS
- [ ] Manual end-to-end: `moai mx query --kind anchor` against this project returns ≥1 entry

**Dependencies**: T-SPC004-01 (LSPFanInCounter), T-SPC004-03 (LoadDangerConfig), T-SPC004-04 (LoadSpecModules)

---

#### T-SPC004-13: Performance benchmark fixtures (G-07)

**REQ traceback**: spec §7 Constraints (advisory)

**AC traceback**: AC-08 (pagination)

**Goal**: Add benchmark functions for performance regression detection. Advisory only (CI does not enforce thresholds).

**Files**:
- `internal/mx/resolver_query_bench_test.go` (new ~80 LOC)

**Sub-tasks**:
- `BenchmarkResolver_Resolve_1KTags`: generate 1000-tag sidecar fixture, b.N iterations of `Resolve(Query{Limit: 100})`. Spec §7 target: <100ms per op (advisory)
- `BenchmarkResolver_Resolve_50AnchorsLSP`: 50 ANCHOR + LSP mock. Spec §7 target: <2s per op (advisory)
- `BenchmarkSpecAssociator_Associate_1KSpecs`: scaling check on path-based association (informational)
- Use `testing.B.ReportMetric` for ns/op + custom metrics

**Acceptance**:
- [ ] Benchmarks compile cleanly (`go test -bench Benchmark ./internal/mx/ -run='^$'`)
- [ ] Benchmarks complete without error (no assertions; advisory only)

**Dependencies**: T-SPC004-01 (LSP mock for second benchmark)

---

#### T-SPC004-15: 16-language sweep test (G-08)

**REQ traceback**: REQ-SPC-004-001, spec §1 (16-language neutrality)

**AC traceback**: AC-15 (extends to all 16 languages)

**Goal**: Add `TestResolver_AllSixteenLanguages` per plan §M6 description. Verifies resolver layer (filepath.Walk + SPEC association + fan_in textual mode) works across all 16 supported languages.

**Files**:
- `internal/mx/resolver_query_test.go` (extended +200 LOC)

**Sub-tasks**:
- `t.TempDir()` with 16 source files (one per language: go, py, ts, js, rs, java, kt, cs, rb, php, ex, cpp, scala, R, dart, swift)
- Each file contains 1 `@MX:NOTE` + 1 `@MX:ANCHOR` (anchor_id unique per language: `anchor-go-001`, `anchor-py-001`, ...)
- Each anchor's `tag.Body` contains explicit `SPEC-LANG-001` body reference (16 different SPEC IDs)
- Build sidecar via Scanner → Manager.Write
- Run `Resolver.Resolve(Query{Kind: MXAnchor})` → assert 16 entries
- Run `Resolver.Resolve(Query{Kind: MXAnchor, FanInMin: 1})` with TextualFanInCounter → assert all 16 have fan_in≥0 (textual scan finds the anchor_id reference in own file at minimum 0; if cross-references seeded, ≥1)
- Verify each entry's `spec_associations` contains the body-referenced SPEC ID

**Acceptance**:
- [ ] Test PASS
- [ ] All 16 languages produce results
- [ ] Existing AC-15 fixture (`TestResolver_AC15_BodyBasedSpecAssociation` line 449) still PASS

**Dependencies**: SPC-002 Scanner + comment_prefixes (already merged)

---

#### T-SPC004-14: Full test suite + race detector

**REQ traceback**: TRUST 5 Tested

**AC traceback**: All 15 ACs

**Goal**: Run full test suite under race detector with count=1 to disable caching. Zero regressions allowed.

**Sub-tasks**:
- `go test -race -count=1 ./...` → exit 0
- If failure: diagnose root cause, fix, re-run

**Acceptance**:
- [ ] All tests PASS under `-race -count=1`
- [ ] No flaky tests introduced

**Dependencies**: All other tasks complete

---

#### T-SPC004-16: Lint clean

**REQ traceback**: TRUST 5 Readable

**AC traceback**: (cross-cutting)

**Goal**: `golangci-lint run` produces no warnings on touched files.

**Sub-tasks**:
- Run `golangci-lint run`
- Fix any new warnings (especially `errcheck`, `gosec`, `unused`)
- Ensure new functions have godoc comments per `.claude/rules/moai/languages/go.md` MUST list

**Acceptance**:
- [ ] `golangci-lint run` exit 0
- [ ] All exported new symbols have godoc

**Dependencies**: All other tasks complete

---

#### T-SPC004-17: Build + embedded.go regenerate

**REQ traceback**: CLAUDE.local.md §2 Template-First

**AC traceback**: (cross-cutting)

**Goal**: `make build` regenerates `internal/template/embedded.go`. Verify `diff -r .claude/ internal/template/templates/.claude/` is byte-identical (no template change expected for this SPEC since changes are in `internal/`).

**Sub-tasks**:
- `make build` → exit 0
- `diff -r .claude/ internal/template/templates/.claude/` → empty output (or only intentional differences from settings.local.json)
- Verify CLAUDE.local.md §2 Template-First rule respected (no new files in `.claude/` without template counterpart)

**Acceptance**:
- [ ] `make build` succeeds
- [ ] Template parity verified

**Dependencies**: All other tasks complete

---

#### T-SPC004-18: CHANGELOG entry

**REQ traceback**: TRUST 5 Trackable

**AC traceback**: (cross-cutting)

**Goal**: Add 4 entries to `CHANGELOG.md` under `## Unreleased` section per plan §M6.

**Sub-tasks**:
- Append:
  - `feat(mx/SPEC-V3R2-SPC-004): LSP find-references integration via powernap (LSPFanInCounter)`
  - `feat(mx/SPEC-V3R2-SPC-004): mx.yaml danger_categories: + test_paths: user wire-up`
  - `feat(mx/SPEC-V3R2-SPC-004): .moai/specs/*/spec.md module: frontmatter auto-load + SpecAssociator injection`
  - `feat(mx/SPEC-V3R2-SPC-004): Resolver.ResolveAnchorCallsites() API parity (additive)`

**Acceptance**:
- [ ] CHANGELOG diff visible in PR
- [ ] Entries follow Conventional Commits style

**Dependencies**: None

---

#### T-SPC004-19: @MX tags applied

**REQ traceback**: spec §10 Traceability + plan §6

**AC traceback**: (cross-cutting)

**Goal**: Apply 6 @MX tags per plan §6 (mx_plan): 1 ANCHOR + 2 WARN + 3 NOTE.

**Sub-tasks**:
- Add tags listed in plan §6 to:
  - `internal/mx/fanin_lsp.go:LSPFanInCounter.Count` → `@MX:ANCHOR` + `@MX:REASON`
  - `internal/mx/fanin_lsp.go:LSPFanInCounter` (struct) → `@MX:NOTE`
  - `internal/mx/spec_loader.go:LoadSpecModules` → `@MX:NOTE`
  - `internal/mx/danger_category.go:LoadDangerConfig` → `@MX:NOTE`
  - `internal/mx/resolver_query.go:validateQuery` → `@MX:WARN` + `@MX:REASON`
  - `internal/mx/resolver.go:ResolveAnchorCallsites` → `@MX:WARN` + `@MX:REASON`
- Verify via `moai mx query --kind anchor --file-prefix internal/mx/fanin_lsp.go` shows ≥1 entry post-implementation

**Acceptance**:
- [ ] 6 tags present in source
- [ ] Each WARN tag has sibling `@MX:REASON` (per mx-tag-protocol.md)
- [ ] ANCHOR tag has fan_in justification

**Dependencies**: All implementation tasks complete

---

#### T-SPC004-20: Manual end-to-end verification

**REQ traceback**: spec §6 (Acceptance) + plan §9.1

**AC traceback**: (smoke test)

**Goal**: Run `moai mx query` against the moai-adk-go project itself and verify results are sensible.

**Sub-tasks**:
- Build local binary: `make install`
- `moai mx query --kind anchor --fan-in-min 1 --format json | jq 'length'` → expect ≥1 result
- Verify at least one entry has `fan_in_method: "lsp"` (gopls running)
- `moai mx query --danger concurrency --format table` → expect human-readable table
- `moai mx query --spec SPEC-V3R2-SPC-004 --kind anchor` → expect entries from `internal/mx/` if @MX tags applied (T-SPC004-19)
- Verify `MOAI_MX_QUERY_STRICT=1 moai mx query --fan-in-min 1` succeeds when LSP is available

**Acceptance**:
- [ ] All 4 manual verifications pass
- [ ] Performance is subjectively responsive (<2s for project's tag set)

**Dependencies**: All other tasks complete

---

## 3. Task Dependency Graph

```
T-SPC004-01 (LSPFanInCounter)
  ├── T-SPC004-02 (LSP RED tests + strictMode)
  ├── T-SPC004-11 (ResolveAnchorCallsites — LSP path)
  └── T-SPC004-12 (CLI wire — LSP detect)

T-SPC004-03 (LoadDangerConfig)
  └── T-SPC004-06 (validateQuery danger 분기)
      └── T-SPC004-12 (CLI wire — danger config)

T-SPC004-04 (LoadSpecModules)
  ├── T-SPC004-05 (spec_loader RED tests + integration)
  └── T-SPC004-12 (CLI wire — SpecAssociator)

T-SPC004-07 (isTestFile glob helper)
  └── T-SPC004-08 (isTestFile RED tests)

T-SPC004-10 (Callsite struct)
  └── T-SPC004-11 (ResolveAnchorCallsites)

T-SPC004-09 (stderr fixture) — independent

T-SPC004-13 (benchmarks) — depends on T-SPC004-01 for LSP mock
T-SPC004-15 (16-lang sweep) — depends on T-SPC004-04, T-SPC004-05

Verification (no inter-deps among these but order matters):
  T-SPC004-14 (full suite -race) → T-SPC004-16 (lint) → T-SPC004-17 (build) → T-SPC004-18 (CHANGELOG) → T-SPC004-19 (MX tags) → T-SPC004-20 (manual verify)
```

Critical path: T-SPC004-01 → T-SPC004-02 → T-SPC004-12 → T-SPC004-15 → T-SPC004-14 → T-SPC004-20.

Parallelizable: M2 (T-03/06), M3 (T-04/05), M5 (T-07/08/09) can run in parallel after M1 LSP scaffolding. M4 (T-10/11) depends on M1 LSP for one test path.

---

## 4. SPC-002 Reuse Plan

본 SPEC 은 SPC-002 의 다음 표면을 read-only consume:

| SPC-002 표면 | SPC-004 사용 site |
|-------------|-------------------|
| `mx.Manager` (sidecar I/O) | `Resolver.Resolve` 가 `r.manager.Load()` 호출 (existing) |
| `mx.Sidecar` struct + schema_version 2 | JSON unmarshal 시 schema check (existing) |
| `mx.Tag` struct (8 fields) | `applyFilters` + `TagResult` 생성 (existing) |
| `mx.SidecarFileName` 상수 | os.Stat for SidecarUnavailable detection (existing) |
| `mx.NewScanner()` + `comment_prefixes.go` | T-SPC004-15 16-lang sweep 가 source file 스캔 시 사용 |
| `mx.MXNote/MXWarn/MXAnchor/MXTodo/MXLegacy` 상수 | `validKinds` map (existing) + new tests |

본 SPEC 의 sidecar 변경은 0건. SPC-002 의 schema_version: 2 invariant 보존 의무 (plan §1.2.1).

---

## 5. Wave-Split Decision

**Total tasks**: 20.
**Threshold**: 30 (per `feedback_large_spec_wave_split.md`).
**Decision**: **No wave-split**. Single PR for run-phase covering all 6 milestones.

Rationale: Total task count is 20 (well below 30-task threshold). Critical path involves M1 → M3 → M6, with parallelizable branches in M2/M3/M5. Estimated ~1,400 LOC delta is within single-PR review capacity.

Risk if wave-split forced: M4 (Callsites API) artificial split would orphan G-04 across PRs. M1 (LSP) scaffolding split would orphan G-01 LSP-backed counter from its consumers (G-04, M6 sweep).

---

End of tasks.

Version: 0.1.0
Status: Task list for SPEC-V3R2-SPC-004
