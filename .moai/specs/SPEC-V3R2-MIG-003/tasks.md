# Task Decomposition — SPEC-V3R2-MIG-003

> Companion to `spec.md`, `plan.md`, `research.md`, `acceptance.md`.
> Task IDs: T-MIG003-01 through T-MIG003-20.
> Priority labels per agent-common-protocol.md (no time estimates).

---

## HISTORY

| Version | Date       | Author       | Description                                       |
|---------|------------|--------------|---------------------------------------------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial task decomposition: 20 tasks across M1-M4 |

---

## Milestone M1 — RED Phase (Priority: P0 Critical)

### T-MIG003-01 — Write loader_constitution_test.go (failing)

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/loader_constitution_test.go`
- `internal/config/testdata/constitution-valid/constitution.yaml` (mirror of `.moai/config/sections/constitution.yaml`)
- `internal/config/testdata/constitution-malformed/constitution.yaml` (truncated mid-map)

**Test cases**:
- `TestLoadConstitutionConfig_ValidParse` — assert all 8 sub-fields populated
- `TestLoadConstitutionConfig_MissingFile_ReturnsDefaults` — empty dir → default config returned, no error
- `TestLoadConstitutionConfig_Malformed_ReturnsErrInvalidYAML` — `errors.Is(err, ErrInvalidYAML)` check
- `TestLoadConstitutionConfig_ForbiddenLibrariesExposed` — non-empty list readable

**Exit criterion**: All 4 tests COMPILE but FAIL with `undefined: LoadConstitutionConfig` / `undefined: ConstitutionConfig`.

**Maps to**: REQ-MIG003-001/002/003/004/007/009 (AC-MIG003-01/02/03/04/07/08)

---

### T-MIG003-02 — Write loader_context_test.go (failing)

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/loader_context_test.go`
- `internal/config/testdata/context-valid/context.yaml`
- `internal/config/testdata/context-malformed/context.yaml`

**Test cases**:
- `TestLoadContextConfig_ValidParse` — assert `cfg.TokenBudget.MaxInjectionTokens == 5000`, `SkipIfUsageAbove == 150000`, `Search.DateRangeDays == 30`
- `TestLoadContextConfig_MissingFile_ReturnsDefaults`
- `TestLoadContextConfig_Malformed_ReturnsErrInvalidYAML`

**Exit criterion**: All 3 tests COMPILE but FAIL with undefined references.

**Maps to**: REQ-MIG003-001/002/003/004/007/010 (AC-MIG003-01/02/03/04/07/09)

---

### T-MIG003-03 — Write loader_interview_test.go (failing)

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/loader_interview_test.go`
- `internal/config/testdata/interview-valid/interview.yaml`
- `internal/config/testdata/interview-missing/` (empty dir)

**Test cases**:
- `TestLoadInterviewConfig_ValidParse` — assert `cfg.ClarityThreshold == 4`, `Plan.MaxRounds == 5`, `len(SkipConditions) == 3`
- `TestLoadInterviewConfig_MissingFile_ReturnsDefaults`
- `TestLoadInterviewConfig_Malformed_ReturnsErrInvalidYAML`

**Exit criterion**: All 3 tests COMPILE but FAIL.

**Maps to**: REQ-MIG003-001/002/003/004/007/011 (AC-MIG003-01/02/03/04/07/10)

---

### T-MIG003-04 — Write loader_design_test.go (failing)

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/loader_design_test.go`
- `internal/config/testdata/design-valid/design.yaml`
- `internal/config/testdata/design-pass-threshold-violation/design.yaml` (pass_threshold: 0.45)

**Test cases**:
- `TestLoadDesignConfig_ValidParse` — assert `cfg.GanLoop.PassThreshold == 0.75`, `MaxIterations == 5`, `SprintContract.Enabled == true`
- `TestLoadDesignConfig_MissingFile_ReturnsDefaults`
- `TestLoadDesignConfig_Malformed_ReturnsErrInvalidYAML`
- `TestLoadDesignConfig_PassThresholdFloorViolation` — `errors.Is(err, ErrPassThresholdFloor)` for pass_threshold < 0.60

**Exit criterion**: All 4 tests COMPILE but FAIL.

**Maps to**: REQ-MIG003-001/002/003/004/007/014 (AC-MIG003-01/02/03/04/07/11)

---

### T-MIG003-05 — Write SunsetConfig DORMANT marker test + Config root field tests (failing)

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files modified**:
- `internal/config/types_test.go` (extend existing file)

**Test cases**:
- `TestSunsetConfig_DORMANT_GodocMarker` — read `internal/config/types.go` raw; assert godoc block before `type SunsetConfig struct` contains `DORMANT` literal and `Activation deferred to a future SPEC` phrase
- `TestRootConfig_HasFourNewSectionFields` — using reflection on `Config{}`, assert 4 fields exist: `Constitution`, `ContextSearch`, `Interview`, `Design`

**Exit criterion**: Both tests FAIL (godoc not yet updated, fields not yet added).

**Maps to**: REQ-MIG003-001/006/015 (AC-MIG003-01/06)

---

### T-MIG003-16 — Write audit_loader_completeness_test.go (failing)

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/audit_loader_completeness_test.go`

**Test case**:
- `TestAuditLoaderCompleteness` — scan `internal/template/templates/.moai/config/sections/*.yaml`, cross-reference against loader registry + `acknowledgedUnloadedSections` allowlist. Fail with `YAML_SECTION_NO_LOADER: <name>` on uncovered.

**Allowlist initial content** (`acknowledgedUnloadedSections`):
```go
var acknowledgedUnloadedSections = []string{
    "db", "gate", "github-actions", "git-strategy", "lsp", "memo",
    "mx", "observability", "project", "security", "sunset", "system", "workflow",
}
```

**Dedicated-loader allowlist** (`acknowledgedDedicatedLoaders` — loaded outside `Loader.Load()` chain):
```go
var acknowledgedDedicatedLoaders = []string{"harness", "runtime"}
```

**Exit criterion**: Test COMPILEs but FAILs because `Loader.Load()` does not yet register the 4 new sections (constitution, context_search, interview, design will be flagged).

**Maps to**: REQ-MIG003-013 (AC-MIG003-12)

---

### T-MIG003-17 — Write audit_struct_yaml_symmetry_test.go (failing)

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M1 RED
**Files created**:
- `internal/config/audit_struct_yaml_symmetry_test.go`

**Test cases**:
- `TestStructYAMLSymmetry_Constitution` — AST-parse `ConstitutionConfig` yaml tags + parse `constitution.yaml`; assert bijection
- `TestStructYAMLSymmetry_Context`
- `TestStructYAMLSymmetry_Interview`
- `TestStructYAMLSymmetry_Design`

**Exit criterion**: All 4 tests FAIL (structs don't exist yet).

**Maps to**: REQ-MIG003-016 (AC-MIG003-14)

---

## Milestone M2 — GREEN Phase (Priority: P0 Critical)

### T-MIG003-06 — Add ConstitutionConfig struct + wrapper to types.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/types.go`

**Edits**:
- Add struct `ConstitutionConfig` adjacent to `HarnessConfig` (after line 538). Fields per `research.md` §3.1.
- Add struct `ConstitutionArchitecture`, `ConstitutionNaming`, `ConstitutionPerformance`, `ConstitutionSecurity` (sub-types).
- Add godoc with `@MX:ANCHOR: [AUTO]` mentioning "SPEC-V3R2-EXT-004 framework optional hook for forbidden-library policy enforcement".
- Add wrapper `constitutionFileWrapper{Constitution ConstitutionConfig \`yaml:"constitution"\`}` to wrapper section.
- Add field to root `Config` struct: `Constitution ConstitutionConfig \`yaml:"constitution"\``.

**Exit criterion**: `go build ./internal/config/...` succeeds; `TestRootConfig_HasFourNewSectionFields` partial pass for Constitution; `TestStructYAMLSymmetry_Constitution` PASSes if YAML keys match.

**Maps to**: REQ-MIG003-001/005/009 (AC-MIG003-01/05/08)

---

### T-MIG003-07 — Add ContextConfig struct + wrapper to types.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/types.go`

**Edits**:
- Add struct `ContextConfig` with sub-fields per `research.md` §3.2.
- Add sub-types: `ContextAutoDetect`, `ContextMemoryIntegration`, `ContextPerformance`, `ContextSearch`, `ContextTokenBudget`.
- Add godoc with hot-path reference to "CLAUDE.md §16 Context Search Protocol".
- Add wrapper `contextFileWrapper{ContextSearch ContextConfig \`yaml:"context_search"\`}` (note YAML key is `context_search`).
- Add field to root `Config`: `ContextSearch ContextConfig \`yaml:"context_search"\``.

**Exit criterion**: `go build` succeeds; symmetry test PASS.

**Maps to**: REQ-MIG003-001/005/010 (AC-MIG003-01/05/09)

---

### T-MIG003-08 — Add InterviewConfig struct + wrapper to types.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/types.go`

**Edits**:
- Add struct `InterviewConfig` per `research.md` §3.3.
- Add sub-type: `InterviewMode{MaxRounds int, QuestionsPerRound int}`.
- Add godoc with hot-path reference to "SPEC-V3R2-WF-003 discovery mode".
- Add wrapper `interviewFileWrapper{Interview InterviewConfig \`yaml:"interview"\`}`.
- Add field to root `Config`: `Interview InterviewConfig \`yaml:"interview"\``.

**Exit criterion**: `go build` succeeds; symmetry test PASS.

**Maps to**: REQ-MIG003-001/005/011 (AC-MIG003-01/05/10)

---

### T-MIG003-09 — Add DesignConfig struct + wrapper to types.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/types.go`

**Edits**:
- Add struct `DesignConfig` per `research.md` §3.4.
- Add sub-types: `DesignAdaptation`, `DesignBrandContext`, `DesignClaudeDesign`, `DesignDocs`, `DesignEvaluator`, `DesignEvolution`, `DesignFigma`, `DesignGanLoop`, `DesignSprintContract`, `DesignGraduationCriteria`.
- Add godoc with hot-path reference to "GAN loop runtime sprint contract".
- Add wrapper `designFileWrapper{Design DesignConfig \`yaml:"design"\`}`.
- Add field to root `Config`: `Design DesignConfig \`yaml:"design"\``.

**Exit criterion**: `go build` succeeds; symmetry test PASS.

**Maps to**: REQ-MIG003-001/005/014 (AC-MIG003-01/05/11)

---

### T-MIG003-10 — Create internal/config/loader_constitution.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files created**:
- `internal/config/loader_constitution.go`

**Implementation**:
- `func LoadConstitutionConfig(path string) (*ConstitutionConfig, error)` — open file via `os.ReadFile`, unmarshal into `constitutionFileWrapper`, return typed config; `ErrConfigNotFound` on missing, `ErrInvalidYAML` on parse error.
- `func (l *Loader) loadConstitutionSection(dir string, cfg *Config)` — use `loadYAMLFile(dir, "constitution.yaml", wrapper)`. On `loaded == true`: `cfg.Constitution = wrapper.Constitution; l.loadedSections["constitution"] = true`. On `loaded == false, err == nil`: no-op (defaults already in cfg via NewDefaultConfig).

**Exit criterion**: `TestLoadConstitutionConfig_*` all PASS; `TestLoadConstitutionConfig_ForbiddenLibrariesExposed` PASS.

**Maps to**: REQ-MIG003-002/003/004/009 (AC-MIG003-02/03/08)

---

### T-MIG003-11 — Create internal/config/loader_context.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files created**:
- `internal/config/loader_context.go`

**Implementation**:
- `func LoadContextConfig(path string) (*ContextConfig, error)` — same pattern as T-MIG003-10.
- `func (l *Loader) loadContextSection(dir string, cfg *Config)` — same pattern. `loadedSections["context_search"]`.

**Exit criterion**: `TestLoadContextConfig_*` all PASS.

**Maps to**: REQ-MIG003-002/003/004/010 (AC-MIG003-02/03/09)

---

### T-MIG003-12 — Create internal/config/loader_interview.go

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files created**:
- `internal/config/loader_interview.go`

**Implementation**:
- `func LoadInterviewConfig(path string) (*InterviewConfig, error)` — same pattern.
- `func (l *Loader) loadInterviewSection(dir string, cfg *Config)` — same pattern. `loadedSections["interview"]`.

**Exit criterion**: `TestLoadInterviewConfig_*` all PASS.

**Maps to**: REQ-MIG003-002/003/004/011 (AC-MIG003-02/03/10)

---

### T-MIG003-13 — Create internal/config/loader_design.go (with pass-threshold floor validation)

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files created**:
- `internal/config/loader_design.go`

**Implementation**:
- `func LoadDesignConfig(path string) (*DesignConfig, error)` — same base pattern PLUS post-unmarshal validation:
  - If `cfg.GanLoop.PassThreshold > 0` AND `cfg.GanLoop.PassThreshold < 0.60` → return `&ValidationError{Field: "gan_loop.pass_threshold", Value: cfg.GanLoop.PassThreshold, Message: "below FROZEN floor 0.60", Wrapped: ErrPassThresholdFloor}`.
- `func (l *Loader) loadDesignSection(dir string, cfg *Config)` — same pattern. `loadedSections["design"]`.

**Exit criterion**: `TestLoadDesignConfig_*` all PASS, including `TestLoadDesignConfig_PassThresholdFloorViolation`.

**Maps to**: REQ-MIG003-002/003/004/014 (AC-MIG003-02/03/11)

---

### T-MIG003-14 — Extend internal/config/defaults.go with 4 default helpers

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/defaults.go`

**Edits**:
- Add `func defaultConstitutionConfig() ConstitutionConfig` — populated from `internal/template/templates/.moai/config/sections/constitution.yaml` content (literal Go values).
- Add `func defaultContextConfig() ContextConfig`.
- Add `func defaultInterviewConfig() InterviewConfig`.
- Add `func defaultDesignConfig() DesignConfig`.
- In `NewDefaultConfig()` body, assign: `cfg.Constitution = defaultConstitutionConfig()`, `cfg.ContextSearch = defaultContextConfig()`, `cfg.Interview = defaultInterviewConfig()`, `cfg.Design = defaultDesignConfig()`.

**Exit criterion**: `TestLoad*_MissingFile_ReturnsDefaults` for all 4 sections PASS.

**Maps to**: REQ-MIG003-004 (AC-MIG003-03)

---

### T-MIG003-15 — Update SunsetConfig godoc + Config root in types.go

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/types.go`

**Edits**:
- Prepend to `SunsetConfig` godoc (line 309):
  ```
  // @MX:NOTE: [AUTO] DORMANT — struct schema defined but no runtime hot path
  //   enforces sunset conditions. Activation deferred to a future SPEC.
  //   Do NOT add LoadSunsetConfig until activation SPEC is filed.
  //   REQ-MIG003-006 ↔ REQ-MIG003-015 reference.
  ```

**Exit criterion**: `TestSunsetConfig_DORMANT_GodocMarker` PASSes.

**Maps to**: REQ-MIG003-006/015 (AC-MIG003-06)

---

### T-MIG003-18 — Wire 4 new section loaders into Loader.Load()

**Priority**: P0 Critical
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M2 GREEN
**Files modified**:
- `internal/config/loader.go`

**Edits**:
- In `Loader.Load()` body, after line 71 (`l.loadResearchSection(sectionsDir, cfg)`), append:
  ```go
  // Load constitution section (REQ-MIG003-001/002)
  l.loadConstitutionSection(sectionsDir, cfg)

  // Load context_search section (REQ-MIG003-010)
  l.loadContextSection(sectionsDir, cfg)

  // Load interview section (REQ-MIG003-011)
  l.loadInterviewSection(sectionsDir, cfg)

  // Load design section (REQ-MIG003-014)
  l.loadDesignSection(sectionsDir, cfg)
  ```

**Exit criterion**: `TestAuditLoaderCompleteness` PASS — the 4 new sections show as loaded.

**Maps to**: REQ-MIG003-002 (AC-MIG003-02/12)

---

## Milestone M3 — GREEN Phase: CI Guards + Once-per-Session Notice (Priority: P1 High)

### T-MIG003-19 — Create internal/config/sunset_notice.go + test

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M3 GREEN
**Files created**:
- `internal/config/sunset_notice.go`
- `internal/config/sunset_notice_test.go`

**Implementation (sunset_notice.go)**:
```go
package config

import (
    "log/slog"
    "os"
    "path/filepath"
    "sync"
)

var sunsetNoticeOnce sync.Once

// emitSunsetDormantNotice emits a one-shot session log per REQ-MIG003-018.
// Guarded by sync.Once so it fires at most once per process lifetime.
//
// @MX:NOTE: [AUTO] DORMANT notice helper — fires when sunset.yaml exists,
// reminds future maintainers that SunsetConfig is template-only.
func emitSunsetDormantNotice(sectionsDir string) {
    sunsetNoticeOnce.Do(func() {
        path := filepath.Join(sectionsDir, "sunset.yaml")
        if _, err := os.Stat(path); err == nil {
            slog.Info("SUNSET_CONFIG_DORMANT_NOTICE",
                "spec", "SPEC-V3R2-MIG-003 REQ-018",
                "yaml_path", path,
                "advice", "SunsetConfig has no runtime hot path; activation requires new SPEC")
        }
    })
}
```

**Test cases**:
- `TestSunsetNotice_FiresOnce` — call `Loader.Load()` twice; capture slog records via custom handler; assert exactly 1 record with key `SUNSET_CONFIG_DORMANT_NOTICE`.
- `TestSunsetNotice_AbsentFileNoNotice` — fixture dir without sunset.yaml; assert 0 records.

**Integration**: Add `emitSunsetDormantNotice(sectionsDir)` call inside `Loader.Load()` between `loadDesignSection` and `return cfg, nil` (loader.go:73).

**Exit criterion**: Both tests PASS.

**Maps to**: REQ-MIG003-018 (AC-MIG003-15)

---

### T-MIG003-20 — Verify CI guards PASS at M3 exit

**Priority**: P1 High
**Owner**: manager-develop (cycle_type=tdd)
**Phase**: M3 verification

**Verification commands**:
```bash
go test ./internal/config/... -run TestAuditLoaderCompleteness -v
go test ./internal/config/... -run TestStructYAMLSymmetry -v
go test ./internal/config/... -run TestSunsetNotice -v
make build
```

**Negative-case verification (manual, on a throwaway branch)**:
- Drop `internal/template/templates/.moai/config/sections/__sentinel.yaml` (without loader); rerun `TestAuditLoaderCompleteness`; confirm FAIL with `YAML_SECTION_NO_LOADER: __sentinel`.
- Add `NewBogusField string \`yaml:"new_bogus_field"\`` to `ConstitutionConfig` without YAML; rerun `TestStructYAMLSymmetry_Constitution`; confirm FAIL with `CONFIG_STRUCT_YAML_MISMATCH: field=ConstitutionConfig.NewBogusField, side=go-only`.

**Exit criterion**: positive cases all PASS, negative cases demonstrate FAIL with expected messages, then revert.

**Maps to**: REQ-MIG003-013/016/018 (AC-MIG003-12/14/15)

---

## Milestone M4 — REFACTOR Phase (Priority: P2 Medium)

### T-MIG003-21 — Update .claude/rules/moai/core/settings-management.md

**Priority**: P2 Medium
**Owner**: manager-docs
**Phase**: M4 REFACTOR
**Files modified**:
- `.claude/rules/moai/core/settings-management.md`

**Edits**:
- Add subsection "MoAI Configuration — Section Loaders" enumerating the 10 sections loaded via `Loader.Load()` + 2 dedicated entries (harness, runtime).
- Document 4 new MIG-003 loaders: `LoadConstitutionConfig`, `LoadContextConfig`, `LoadInterviewConfig`, `LoadDesignConfig`.
- Document `SunsetConfig` DORMANT status with REQ-MIG003-006 reference.
- Document 2 CI guards (`YAML_SECTION_NO_LOADER`, `CONFIG_STRUCT_YAML_MISMATCH`).
- Add "Adding a new YAML section" 5-step procedure:
  1. Add YAML file to `internal/template/templates/.moai/config/sections/`
  2. Add struct + wrapper to `internal/config/types.go`
  3. Add default helper to `internal/config/defaults.go`
  4. Add `LoadXxxConfig` + `loadXxxSection` to new `internal/config/loader_<name>.go`
  5. Wire `loadXxxSection` into `Loader.Load()` AND add to `audit_struct_yaml_symmetry_test.go` allowlist
- Reference HRN-001's dedicated entry pattern (research.md §14 Q1 decision: `LoadHarnessConfig` stays outside `Loader.Load()` by design).

**Exit criterion**: file edited; `git diff` shows additive changes only.

**Maps to**: REQ-MIG003 §10 traceability bullet 6

---

### T-MIG003-22 — Add @MX tags + plan-auditor PASS verification

**Priority**: P2 Medium
**Owner**: manager-develop
**Phase**: M4 REFACTOR

**Edits**:
- Add `@MX:NOTE` (and `@MX:ANCHOR` where `fan_in >= 3`) to each new `LoadXxxConfig` public function documenting its hot path consumer.
- Run `golangci-lint run ./...` — confirm zero new findings.
- Run `go test ./...` — confirm zero regressions.
- Update `progress.md`: record `plan_status: audit-ready`, `plan_complete_at: 2026-05-18`, and any acceptance run results.
- Run plan-auditor (per `.claude/rules/moai/workflow/spec-workflow.md` §Phase 0.5) on the 4 plan artifacts (spec.md / plan.md / research.md / acceptance.md / tasks.md). Target verdict: PASS.

**Exit criterion**: plan-auditor PASS, ready for `/moai run SPEC-V3R2-MIG-003`.

**Maps to**: M4 DoD checklist

---

## Task → REQ → AC Coverage Matrix

| Task ID | REQs covered | ACs verified |
|---|---|---|
| T-MIG003-01 | 001, 002, 003, 004, 007, 009 | 01, 02, 03, 04, 07, 08 |
| T-MIG003-02 | 001, 002, 003, 004, 007, 010 | 01, 02, 03, 04, 07, 09 |
| T-MIG003-03 | 001, 002, 003, 004, 007, 011 | 01, 02, 03, 04, 07, 10 |
| T-MIG003-04 | 001, 002, 003, 004, 007, 014 | 01, 02, 03, 04, 07, 11 |
| T-MIG003-05 | 001, 006, 015 | 01, 06 |
| T-MIG003-06 | 001, 005, 009 | 01, 05, 08 |
| T-MIG003-07 | 001, 005, 010 | 01, 05, 09 |
| T-MIG003-08 | 001, 005, 011 | 01, 05, 10 |
| T-MIG003-09 | 001, 005, 014 | 01, 05, 11 |
| T-MIG003-10 | 002, 003, 004, 009 | 02, 03, 08 |
| T-MIG003-11 | 002, 003, 004, 010 | 02, 03, 09 |
| T-MIG003-12 | 002, 003, 004, 011 | 02, 03, 10 |
| T-MIG003-13 | 002, 003, 004, 014 | 02, 03, 11 |
| T-MIG003-14 | 004 | 03 |
| T-MIG003-15 | 006, 015 | 06 |
| T-MIG003-16 | 013 | 12 |
| T-MIG003-17 | 016 | 14 |
| T-MIG003-18 | 002 | 02, 12 |
| T-MIG003-19 | 018 | 15 |
| T-MIG003-20 | 013, 016, 018 | 12, 14, 15 |
| T-MIG003-21 | (doc only) | DoD |
| T-MIG003-22 | (refactor + plan-audit) | DoD |

All 18 REQs covered; all 15 ACs covered. REQ-MIG003-008/012/017 covered by AC verify-only / review (no separate task — handled within T-MIG003-10..13 implementation discipline + T-MIG003-22 lint review).

## Task Ordering & Dependencies

```
M1 (P0 Critical, parallel):
  T-01 → T-02 → T-03 → T-04   (4 RED test files, can be parallel)
  T-05                         (types_test.go extension)
  T-16, T-17                   (audit guards, parallel)

M2 (P0 Critical, sequential within file groups):
  T-06 → T-10                  (Constitution: struct first, loader second)
  T-07 → T-11                  (Context)
  T-08 → T-12                  (Interview)
  T-09 → T-13                  (Design)
  T-14                         (defaults helpers — depends on T-06..09)
  T-15                         (SunsetConfig godoc)
  T-18                         (wiring — depends on T-10..13)

M3 (P1 High, sequential):
  T-19                         (sunset_notice.go + test)
  T-20                         (CI guards verification)

M4 (P2 Medium):
  T-21                         (settings-management.md docs)
  T-22                         (MX tags + plan-auditor PASS)
```

End of tasks.md
