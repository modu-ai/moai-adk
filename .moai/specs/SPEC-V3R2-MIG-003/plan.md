---
spec_id: SPEC-V3R2-MIG-003
phase: plan
status: audit-ready
plan_complete_at: 2026-05-18
---

# Implementation Plan — SPEC-V3R2-MIG-003

> Companion to `spec.md` (REQ-MIG003-001 ~ REQ-MIG003-018, 18 REQ).
> Companion to `research.md` (38 file:line anchors).
> Companion to `acceptance.md` (15 binary AC) and `tasks.md` (T-MIG003-NN).

---

## HISTORY

| Version | Date       | Author       | Description                                                                                          |
|---------|------------|--------------|------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial plan-phase synthesis: M1-M4 milestones, REQ↔Task matrix, HRN-001 reconciliation, risks, file map |

---

## 1. Context

The original `spec.md` (authored 2026-04-23) scoped MIG-003 to add Go loaders for **5 unloaded YAML sections** in `.moai/config/sections/`: constitution, context, interview, design, harness.

Between authoring and this plan (2026-05-18), **SPEC-V3R2-HRN-001 shipped on main** (commit chain ≤ `195b469d8`). HRN-001 delivered the complete `harness.yaml` loader stack: `HarnessConfig` struct (`internal/config/types.go:398-538`), `LoadHarnessConfig()` with 4-stage validation (`internal/config/loader.go:256-311`), FROZEN enum + schema-drift detection (`detectSchemaDrift`, loader.go:318-358), and 7 extended test cases (`loader_harness_extended_test.go:17-200`).

Consequently the MIG-003 effective scope is now **4 unloaded sections** (constitution, context, interview, design) + the sunset.yaml DORMANT formalization + 2 CI guards (`YAML_SECTION_NO_LOADER`, `CONFIG_STRUCT_YAML_MISMATCH`). REQ-MIG003-002 / -004 / -008 / -012 are partially satisfied for the harness slice by HRN-001 and become **verify-only** in this plan; REQ-MIG003-001 / -003 / -005 / -006 / -007 / -009 / -010 / -011 / -013 / -014 / -015 / -016 / -017 / -018 are net-new work.

The plan turns the 4-section gap into M1-M4 milestones with explicit TDD discipline (RED → GREEN → REFACTOR), reuses the HRN-001 loader pattern (wrapper-type + `loadYAMLFile` helper + section assignment + `loadedSections[]` flag), and adds two `make build`-time audit tests that lock the contract going forward. See `research.md` §13 (inventory table) and §15 (38 anchors) for the full state-of-the-codebase reference.

## 2. Approach

TDD across M1-M4. Each milestone has an explicit RED (failing test capturing intent) → GREEN (minimal change to pass) → REFACTOR (cleanup + invariant check) loop.

We do NOT re-implement `LoadHarnessConfig` — that is HRN-001 work product. We DO add a verify-only test ensuring `loadedSections["harness"]` accounting stays internally consistent OR document HRN-001's dedicated entry as canonical (decided at M4 — see `research.md` §14 Q1).

Loader pattern is replicated 4 times from the canonical `loadUserSection` shape (`internal/config/loader.go:88-99`) plus the HRN-001 file-level separation pattern (HRN-001 chose to keep `LoadHarnessConfig` in loader.go itself; MIG-003 deviates by introducing per-section files `internal/config/loader_{constitution,context,interview,design}.go` to keep loader.go from growing unbounded). Each new file has paired `_test.go` covering the 3 baseline cases (valid / missing / malformed) plus consumption-path tests for the runtime hot path REQs.

Defaults strategy: every absent YAML returns a struct populated from `internal/config/defaults.go` helpers, mirroring the template values. This honors REQ-MIG003-004 (sensible defaults on absent file) and provides a single source of truth for the `CONFIG_STRUCT_YAML_MISMATCH` build-time guard (M3 work).

The two CI guards live in `internal/config/audit_*_test.go` files and run on every `go test ./internal/config/...`, locking the contract that (a) every YAML in `.moai/config/sections/` has a loader OR is in an acknowledged exclusion list, and (b) every struct field has a YAML key and vice versa.

## 3. Milestones (Priority-based, NO time estimates)

[HARD] Priority labels only — no time estimates per agent-common-protocol.

### M1 — RED Phase (Priority: P0 Critical, blocking)

**Test scaffolding before any implementation code.**

- Create fixture directories in `internal/config/testdata/`:
  - `constitution-valid/`, `constitution-malformed/`
  - `context-valid/`, `context-malformed/`
  - `interview-valid/`, `interview-missing/`
  - `design-valid/`, `design-pass-threshold-violation/`
- Write `internal/config/loader_constitution_test.go`:
  - `TestLoadConstitutionConfig_ValidParse` — assert all 8 sub-fields populated from `constitution-valid/constitution.yaml`.
  - `TestLoadConstitutionConfig_MissingFile_ReturnsDefaults` — assert defaults returned, no error.
  - `TestLoadConstitutionConfig_Malformed_ReturnsErrInvalidYAML` — assert `errors.Is(err, ErrInvalidYAML)`.
  - `TestLoadConstitutionConfig_ForbiddenLibrariesExposed` — REQ-MIG003-009: assert non-empty list is readable.
- Write `internal/config/loader_context_test.go`:
  - `TestLoadContextConfig_ValidParse` — assert `token_budget.max_injection_tokens`, `token_budget.skip_if_usage_above`, `search.date_range_days` populated (REQ-MIG003-010).
  - `TestLoadContextConfig_MissingFile_ReturnsDefaults`
  - `TestLoadContextConfig_Malformed_ReturnsErrInvalidYAML`
- Write `internal/config/loader_interview_test.go`:
  - `TestLoadInterviewConfig_ValidParse` — assert `clarity_threshold=4`, `plan.max_rounds=5`, `skip_conditions` length=3 (REQ-MIG003-011).
  - `TestLoadInterviewConfig_MissingFile_ReturnsDefaults`
  - `TestLoadInterviewConfig_Malformed_ReturnsErrInvalidYAML`
- Write `internal/config/loader_design_test.go`:
  - `TestLoadDesignConfig_ValidParse` — assert `gan_loop.pass_threshold=0.75`, `gan_loop.max_iterations=5`, `sprint_contract.enabled=true` (REQ-MIG003-014).
  - `TestLoadDesignConfig_MissingFile_ReturnsDefaults`
  - `TestLoadDesignConfig_Malformed_ReturnsErrInvalidYAML`
  - `TestLoadDesignConfig_PassThresholdFloorViolation` — REQ-MIG003-014 paired with HRN-001 floor: assert `errors.Is(err, ErrPassThresholdFloor)` when `pass_threshold < 0.60`.
- Write `internal/config/types_test.go` extensions:
  - `TestSunsetConfig_DORMANT_GodocMarker` — REQ-MIG003-006: assert godoc comment for `SunsetConfig` contains `DORMANT` literal (parse types.go AST or read file contents).
  - `TestRootConfig_HasFourNewSectionFields` — assert `Config` struct exposes `Constitution`, `ContextSearch`, `Interview`, `Design` fields (AC-MIG003-01).
- Write `internal/config/audit_loader_completeness_test.go`:
  - `TestAuditLoaderCompleteness` — REQ-MIG003-013: cross-reference template YAML files vs registered loaders; fail on uncovered section names with `YAML_SECTION_NO_LOADER: <name>`.
  - Maintain `acknowledgedUnloadedSections` allowlist for out-of-MIG-003-scope files (mx, gate, lsp, db, github-actions, observability, project, memo, system, security, git-strategy, workflow, sunset).
- Write `internal/config/audit_struct_yaml_symmetry_test.go`:
  - `TestStructYAMLSymmetry_Constitution` — REQ-MIG003-016: AST-parse `ConstitutionConfig` yaml tags, parse `constitution.yaml`, assert symmetric.
  - Same for Context, Interview, Design.

[HARD] All RED-phase tests MUST fail before any M2 implementation code is written. Run `go test ./internal/config/... -run "TestLoadConstitution|TestLoadContext|TestLoadInterview|TestLoadDesign|TestSunsetConfig_DORMANT|TestRootConfig_HasFour|TestAuditLoader|TestStructYAMLSymmetry"` and confirm all fail with expected messages.

### M2 — GREEN Phase (Priority: P0 Critical)

**Minimal implementation to pass M1 tests.**

- Add 4 new structs to `internal/config/types.go` (adjacent to `HarnessConfig` at line 402; new lines ~540-550):
  - `ConstitutionConfig` — fields: `ApprovedFrameworks []string`, `ApprovedLanguages []string`, `Architecture ConstitutionArchitecture`, `ForbiddenPatterns []string` (aliased as `ForbiddenLibraries` in godoc for REQ-MIG003-009), `NamingConventions ConstitutionNaming`, `Performance ConstitutionPerformance`, `Security ConstitutionSecurity`.
  - `ContextConfig` — fields: `AutoDetect ContextAutoDetect`, `Enabled bool`, `MemoryIntegration ContextMemoryIntegration`, `Performance ContextPerformance`, `Search ContextSearch`, `TokenBudget ContextTokenBudget`.
  - `InterviewConfig` — fields: `ClarityThreshold int`, `Enabled bool`, `Plan InterviewMode`, `Project InterviewMode`, `SkipConditions []string`.
  - `DesignConfig` — fields: `Adaptation DesignAdaptation`, `BrandContext DesignBrandContext`, `ClaudeDesign DesignClaudeDesign`, `DefaultFramework string`, `DesignDocs DesignDocs`, `Enabled bool`, `Evaluator DesignEvaluator`, `Evolution DesignEvolution`, `Figma DesignFigma`, `GanLoop DesignGanLoop`, `Version string`.
- Add `@MX:ANCHOR: [AUTO]` godoc for each struct documenting at least one runtime hot path (REQ-MIG003-005):
  - `ConstitutionConfig` → consumed by SPEC-V3R2-EXT-004 framework optional hook for forbidden-library policy enforcement.
  - `ContextConfig` → consumed by CLAUDE.md §16 Context Search Protocol token-budget gate.
  - `InterviewConfig` → consumed by SPEC-V3R2-WF-003 discovery mode clarity-threshold + round limits.
  - `DesignConfig` → consumed by GAN loop runtime sprint contract + adaptation iteration limits.
- Add 4 wrapper types to `internal/config/types.go` (adjacent to existing wrappers, line ~610):
  - `constitutionFileWrapper{Constitution ConstitutionConfig \`yaml:"constitution"\`}`
  - `contextFileWrapper{ContextSearch ContextConfig \`yaml:"context_search"\`}`
  - `interviewFileWrapper{Interview InterviewConfig \`yaml:"interview"\`}`
  - `designFileWrapper{Design DesignConfig \`yaml:"design"\`}`
- Extend root `Config` struct (types.go:15-33) with 4 new fields:
  - `Constitution ConstitutionConfig \`yaml:"constitution"\``
  - `ContextSearch ContextConfig \`yaml:"context_search"\``
  - `Interview InterviewConfig \`yaml:"interview"\``
  - `Design DesignConfig \`yaml:"design"\``
- Create `internal/config/loader_constitution.go`:
  - `func LoadConstitutionConfig(path string) (*ConstitutionConfig, error)` — reads file, unmarshals into wrapper, returns typed config. Returns `ErrConfigNotFound` on missing, `ErrInvalidYAML` on malformed.
  - `func (l *Loader) loadConstitutionSection(dir string, cfg *Config)` — uses `loadYAMLFile` helper; on `(true, nil)` assign + flag `loadedSections["constitution"] = true`. Graceful on absent (REQ-MIG003-004).
- Create `internal/config/loader_context.go`:
  - `func LoadContextConfig(path string) (*ContextConfig, error)` — same pattern.
  - `func (l *Loader) loadContextSection(dir string, cfg *Config)` — same pattern; `loadedSections["context_search"]`.
- Create `internal/config/loader_interview.go`:
  - `func LoadInterviewConfig(path string) (*InterviewConfig, error)` — same pattern.
  - `func (l *Loader) loadInterviewSection(dir string, cfg *Config)` — same; `loadedSections["interview"]`.
- Create `internal/config/loader_design.go`:
  - `func LoadDesignConfig(path string) (*DesignConfig, error)` — same pattern + ADD pass-threshold floor validation: if `gan_loop.pass_threshold < 0.60` return `&ValidationError{Field: "gan_loop.pass_threshold", Wrapped: ErrPassThresholdFloor}` reusing HRN-001 sentinel (errors.go:79).
  - `func (l *Loader) loadDesignSection(dir string, cfg *Config)` — same pattern; `loadedSections["design"]`.
- Wire 4 new calls into `Loader.Load()` (`internal/config/loader.go:31-74`, after line 71 `loadResearchSection`):
  ```
  l.loadConstitutionSection(sectionsDir, cfg)
  l.loadContextSection(sectionsDir, cfg)
  l.loadInterviewSection(sectionsDir, cfg)
  l.loadDesignSection(sectionsDir, cfg)
  ```
- Extend `internal/config/defaults.go` `NewDefaultConfig()` with 4 default helpers (`defaultConstitutionConfig`, `defaultContextConfig`, `defaultInterviewConfig`, `defaultDesignConfig`) returning struct populated from template values (mirrors `.moai/config/sections/*.yaml` content). Assign in `NewDefaultConfig()` body.
- Update `SunsetConfig` godoc in `internal/config/types.go:309` to prepend `// @MX:NOTE: [AUTO] DORMANT — struct schema defined but no runtime hot path enforces sunset conditions. Activation deferred to a future SPEC. Do NOT add LoadSunsetConfig until activation SPEC is filed.` (REQ-MIG003-006).
- Run M1 tests: `go test ./internal/config/...` — ALL 4 loader test files + symmetry tests + completeness audit MUST pass.

[HARD] Drift-guard at M2 exit: `git diff --stat internal/config/` must show only the files enumerated above (no surprise edits). M2 exit criterion: 100% of M1 RED tests green.

### M3 — GREEN Phase (CI Guards + Once-per-Session Notice) (Priority: P1 High)

**Lock the contract via build-time guards.**

- Verify `internal/config/audit_loader_completeness_test.go` (from M1) passes after M2 wiring. The `acknowledgedUnloadedSections` allowlist excludes out-of-scope files; the 4 new sections (constitution, context_search, interview, design) MUST be detected as loaded. Sunset.yaml stays in the allowlist (DORMANT — never loaded).
- Verify `internal/config/audit_struct_yaml_symmetry_test.go` passes for all 4 new structs.
- Create `internal/config/sunset_notice.go`:
  - Package-level `var sunsetNoticeOnce sync.Once`.
  - `func emitSunsetDormantNotice(sectionsDir string)` — on first call: stat `<sectionsDir>/sunset.yaml`; if exists, `slog.Info("SUNSET_CONFIG_DORMANT_NOTICE", "spec", "SPEC-V3R2-MIG-003 REQ-018", "yaml_path", path)`. Wrapped in `sync.Once.Do`.
- Add `emitSunsetDormantNotice(sectionsDir)` call inside `Loader.Load()` between `loadDesignSection` and the `return` (loader.go:73) so it fires once per process lifetime when `Loader.Load()` is first invoked with a sections dir containing sunset.yaml.
- Write `internal/config/sunset_notice_test.go`:
  - `TestSunsetNotice_FiresOnce` — call `Loader.Load()` twice; capture `slog` handler; assert exactly 1 `SUNSET_CONFIG_DORMANT_NOTICE` record (REQ-MIG003-018, AC-MIG003-15).
  - `TestSunsetNotice_AbsentFileNoNotice` — fixture without sunset.yaml; assert zero notice records.
  - Use a custom `slog.Logger` with a buffered handler for capture.
- Run `make build` — REQ-MIG003-016 audit symmetry test must pass during build.

[HARD] M3 exit: `go test ./internal/config/...` 100% green; `make build` 100% green.

### M4 — REFACTOR Phase (verification + docs) (Priority: P2 Medium)

**Documentation + verification gates.**

- Update `.claude/rules/moai/core/settings-management.md` to document:
  - 4 new loaders (`LoadConstitutionConfig`, `LoadContextConfig`, `LoadInterviewConfig`, `LoadDesignConfig`).
  - `SunsetConfig` DORMANT status + REQ-MIG003-006 reference.
  - 2 CI guards (`YAML_SECTION_NO_LOADER`, `CONFIG_STRUCT_YAML_MISMATCH`) and how to add a new section in the future (5-step procedure).
  - `harness.yaml` dedicated entry (`LoadHarnessConfig` is NOT in `Loader.Load()` chain by design — HRN-001 enforces stricter semantics) — research.md §14 Q1 decision.
- Refactor pass on `internal/config/loader.go`:
  - Verify no inline literal values overriding YAML defaults (REQ-MIG003-017). Grep `loader_*.go` for numeric/string literals matching YAML field values; any match is `LOADER_HARDCODE_VIOLATION` (covered by code review).
  - Confirm all new loaders use `loadYAMLFile` helper (no copy-paste of read-and-unmarshal logic).
- Add `@MX:NOTE` (and `@MX:ANCHOR` where `fan_in >= 3`) tags per `.claude/rules/moai/workflow/mx-tag-protocol.md`:
  - `loadYAMLFile` helper now has fan_in = 9 (was 9, no change) — already tagged.
  - Each new `LoadXxxConfig` public function: add `@MX:NOTE` documenting hot-path consumer.
- Run full test suite: `go test ./...` — confirm zero regressions outside internal/config.
- Run `golangci-lint run` — confirm zero new findings.
- Verify acceptance.md AC-MIG003-01 through AC-MIG003-15 (some are verify-only for HRN-001 deliverables): execute the verification commands listed in acceptance.md and check all pass.
- Plan-auditor PASS verification: confirm `progress.md` records `plan_status: audit-ready` and `plan_complete_at: 2026-05-18`.

[HARD] M4 exit: 100% AC pass, 100% test green, zero lint findings, settings-management.md updated.

## 4. REQ ↔ Task Matrix

| REQ | Tasks | Verified by AC |
|---|---|---|
| REQ-MIG003-001 (5 structs) | T-MIG003-06..09 (4 new structs in M2; HarnessConfig from HRN-001 unchanged) | AC-MIG003-01 |
| REQ-MIG003-002 (5 loaders) | T-MIG003-10..13 (4 new loaders; LoadHarnessConfig from HRN-001) | AC-MIG003-02 (verify-only for harness) |
| REQ-MIG003-003 (typed return + aggregated error) | T-MIG003-10..13 | AC-MIG003-02..04 |
| REQ-MIG003-004 (sensible defaults on absent file) | T-MIG003-14 (defaults.go extensions) | AC-MIG003-03 |
| REQ-MIG003-005 (godoc hot path) | T-MIG003-06..09 (godoc per struct) | AC-MIG003-05 |
| REQ-MIG003-006 (SunsetConfig DORMANT marker) | T-MIG003-15 | AC-MIG003-06 |
| REQ-MIG003-007 (unit tests valid/missing/malformed) | T-MIG003-01..04 (M1 RED tests) | AC-MIG003-07 |
| REQ-MIG003-008 (HARNESS_CONFIG_PARSE_ERROR) | T-MIG003-05 (verify HRN-001 ErrInvalidYAML satisfies) | AC-MIG003-04 (verify-only) |
| REQ-MIG003-009 (ConstitutionConfig.ForbiddenLibraries exposed) | T-MIG003-06, T-MIG003-10 | AC-MIG003-08 |
| REQ-MIG003-010 (ContextConfig token_budget + date_range_days) | T-MIG003-07, T-MIG003-11 | AC-MIG003-09 |
| REQ-MIG003-011 (InterviewConfig clarity_threshold + max_rounds) | T-MIG003-08, T-MIG003-12 | AC-MIG003-10 |
| REQ-MIG003-012 (default level=standard) | T-MIG003-14 (verify HRN-001 harness defaults handled via dedicated entry) | AC-MIG003-03 (verify-only for harness) |
| REQ-MIG003-013 (YAML_SECTION_NO_LOADER CI guard) | T-MIG003-16 (audit_loader_completeness_test.go) | AC-MIG003-12 |
| REQ-MIG003-014 (DesignConfig sprint_contract + adaptation) | T-MIG003-09, T-MIG003-13 | AC-MIG003-11 |
| REQ-MIG003-015 (sunset activation future-proofing) | T-MIG003-15 (godoc note) | AC-MIG003-06 (forward-looking) |
| REQ-MIG003-016 (CONFIG_STRUCT_YAML_MISMATCH `make build`) | T-MIG003-17 (audit_struct_yaml_symmetry_test.go) | AC-MIG003-14 |
| REQ-MIG003-017 (LOADER_HARDCODE_VIOLATION) | T-MIG003-19 (code review checklist + grep audit) | AC-MIG003-13 |
| REQ-MIG003-018 (SUNSET_CONFIG_DORMANT_NOTICE once) | T-MIG003-18 (sunset_notice.go + test) | AC-MIG003-15 |

## 5. File Map

### 5.1 Files CREATED (12)

| Path | Purpose | Approx LOC |
|---|---|---|
| internal/config/loader_constitution.go | ConstitutionConfig loader + section helper | ~80 |
| internal/config/loader_context.go | ContextConfig loader + section helper | ~80 |
| internal/config/loader_interview.go | InterviewConfig loader + section helper | ~70 |
| internal/config/loader_design.go | DesignConfig loader + section helper + pass-threshold validation | ~120 |
| internal/config/sunset_notice.go | Once-per-session DORMANT notice (REQ-018) | ~40 |
| internal/config/loader_constitution_test.go | 4 tests (valid/missing/malformed/forbidden-list-exposed) | ~150 |
| internal/config/loader_context_test.go | 3 tests | ~120 |
| internal/config/loader_interview_test.go | 3 tests | ~120 |
| internal/config/loader_design_test.go | 4 tests (incl. pass-threshold violation) | ~180 |
| internal/config/sunset_notice_test.go | 2 tests (fires-once / absent-file-no-notice) | ~80 |
| internal/config/audit_loader_completeness_test.go | CI guard YAML_SECTION_NO_LOADER (REQ-013) | ~120 |
| internal/config/audit_struct_yaml_symmetry_test.go | CI guard CONFIG_STRUCT_YAML_MISMATCH (REQ-016) | ~180 |
| internal/config/testdata/{constitution,context,interview,design}-{valid,malformed/missing}/*.yaml | Test fixtures | 10 yaml files |

### 5.2 Files EDITED (4)

| Path | Change | LOC delta |
|---|---|---|
| internal/config/types.go | 4 new structs + 4 wrapper types + 4 root Config fields + SunsetConfig DORMANT godoc | +200 / -2 |
| internal/config/loader.go | Wire 4 new loadXxxSection calls + emitSunsetDormantNotice into Loader.Load() | +6 |
| internal/config/defaults.go | 4 new default helpers + wiring into NewDefaultConfig() | +100 |
| .claude/rules/moai/core/settings-management.md | Document 4 new loaders + DORMANT sunset + CI guards | +60 |

### 5.3 Files NOT MODIFIED (deliberately)

| Path | Reason |
|---|---|
| internal/template/templates/.moai/config/sections/*.yaml | Templates are READ, not modified (spec.md §1.2) |
| internal/runtime/config.go | LoadRuntime is separate package, runtime.yaml unaffected |
| internal/cli/migrate_agency.go | Existing design.yaml consumer; runtime loader coexists |
| internal/config/errors.go | Reuse existing ErrInvalidYAML / ErrConfigNotFound / ErrPassThresholdFloor; no new sentinels |
| internal/config/loader.go LoadHarnessConfig | HRN-001 deliverable; verify-only |

## 6. Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Wrapper YAML-key collision (constitution.yaml ↔ quality.yaml both use `constitution:` top-level key) | Cross-loader pollution | Disambiguate by **filename** in `loadYAMLFile(dir, "<file>.yaml", wrapper)` — each loader targets a distinct file. Test: load both and assert no cross-pollution (`Loader.Load()` integration test). |
| `context.yaml` top-level key is `context_search:` not `context:` | Silent unmarshal-into-empty if wrapper uses wrong tag | wrapper field tag `yaml:"context_search"` exactly; covered by `TestLoadContextConfig_ValidParse` (M1) |
| design.yaml runtime loader vs migrate_agency.go double-read | None — both flows orthogonal | Verify in `TestLoadDesignConfig_ValidParse` that loader is read-only |
| pass_threshold floor in design.yaml not currently enforced | Spec drift if user sets <0.60 | Add validator in `LoadDesignConfig` reusing `ErrPassThresholdFloor` sentinel (HRN-001) |
| Plan-auditor expects 12-field canonical frontmatter | plan-audit FAIL | spec.md uses 14-field snake_case schema; this is the canonical SDD format used by SPEC-V3R2-MIG-001..003 series. SPEC-V3R2 series predates the unified 12-field schema (per `.claude/rules/moai/development/spec-frontmatter-schema.md` Version History 2026-05-16). Plan-auditor expects 12-field schema; spec.md frontmatter normalization is out-of-scope for MIG-003 itself (would be SPECLINT-DEBT-002 cleanup territory). Flag via `lint.skip: [FrontmatterInvalid]` if blocking. |
| Test fixture maintenance overhead | High | Use file content identical to `.moai/config/sections/*.yaml` (single source of truth) — fixtures are byte-for-byte mirrors of shipping templates. |
| 4 new loaders × 4 wrappers × 4 default helpers = 16 components to review | Review burden | Per-file separation (`loader_<section>.go`) keeps PR diff readable; each file is independently reviewable. |
| Schema drift not detected at runtime for new sections (only at build time via symmetry audit) | Acceptable: build-time guard catches drift; runtime drift detection deferred | Document in settings-management.md that new sections do NOT have `MOAI_CONFIG_STRICT` runtime mode (yet); future SPEC can add. |

## 7. Open Questions (to resolve mid-plan)

OQ1. **Should `harness.yaml` be additionally wired into `Loader.Load()`?**
- Decision: **NO**. Keep HRN-001's dedicated `LoadHarnessConfig` entry. Document in settings-management.md. Audit completeness test treats `harness.yaml` as "intentionally outside Loader.Load()" via the acknowledgedDedicatedLoaders allowlist (separate from acknowledgedUnloadedSections).

OQ2. **Should `LoadDesignConfig` enforce `gan_loop.pass_threshold >= 0.60`?**
- Decision: **YES**. Symmetric with `LoadHarnessConfig`'s pass-threshold floor. Reuse `ErrPassThresholdFloor` sentinel (HRN-001 errors.go:79).

OQ3. **Should `LoadDesignConfig` enforce `evaluator.memory_scope == "per_iteration"`?**
- Decision: **NO**. design.yaml's memory_scope mirrors harness.yaml's FROZEN field structurally, but the FROZEN enforcement lives in `LoadHarnessConfig` (per HRN-002 substrate decision). Design loader treats `memory_scope` as a free string. If future SPEC ties them, refactor then.

OQ4. **`spec.md` REQ-MIG003-014 mentions `adaptation.phase_weights` but YAML has `adaptation.iteration_limits`.**
- Decision: implement `IterationLimits` per actual YAML schema. Add struct godoc clarifying that `phase_weights` in REQ-MIG003-014 is a conceptual term subsuming `iteration_limits`. Update `acceptance.md` AC-MIG003-11 wording accordingly.

OQ5. **Should `audit_loader_completeness_test.go` parse template YAML files or project files?**
- Decision: **template files** (`internal/template/templates/.moai/config/sections/`). The shipping template is the authoritative section list; user-project YAML can differ (deletions, overrides). Test pins behavior to template surface.

OQ6. **What is the `acknowledgedUnloadedSections` allowlist content?**
- Decision: maintained as a sorted `[]string` literal in `audit_loader_completeness_test.go`. Initial entries: `["db", "gate", "github-actions", "git-strategy", "lsp", "memo", "mx", "observability", "project", "security", "sunset", "system", "workflow"]`. Each entry has an inline `// out-of-scope: <reason>` comment referencing the deferring SPEC (where known).

## 8. Test Plan

### 8.1 Test execution sequence

```
# M1 RED phase verification (must FAIL before M2 implementation)
go test ./internal/config/... -run "TestLoadConstitution|TestLoadContext|TestLoadInterview|TestLoadDesign|TestSunsetConfig_DORMANT|TestRootConfig|TestAuditLoaderCompleteness|TestStructYAMLSymmetry|TestSunsetNotice"
# Expected: all FAIL with not-found / not-defined / undefined-field errors

# M2 GREEN phase verification (all PASS)
go test ./internal/config/... -run "TestLoadConstitution|TestLoadContext|TestLoadInterview|TestLoadDesign|TestSunsetConfig_DORMANT|TestRootConfig"
# Expected: all PASS

# M3 GREEN phase verification (CI guards PASS)
go test ./internal/config/... -run "TestAuditLoaderCompleteness|TestStructYAMLSymmetry|TestSunsetNotice"
# Expected: all PASS

# M4 REFACTOR verification (full suite, no regressions)
go test ./...
golangci-lint run ./...
make build
# Expected: 100% green, 0 lint findings, build success
```

### 8.2 Coverage target

- Per-package: 85% (existing project bar)
- New files specifically: target 90%+ as they are small and focused
- Critical paths: 100% — defaults handling, malformed parse, pass-threshold floor violation

## 9. Dependencies

### 9.1 Blocked by (must merge before MIG-003 run-phase)

- SPEC-V3R2-EXT-004 (migration framework) — spec.md §9.1. Status: independent; this plan does NOT block on EXT-004 for the LOADER work itself; it blocks for the SPEC-V3R2-MIG-001 migrator that invokes loader completeness step.

### 9.2 Blocks (MIG-003 unblocks these once merged)

- SPEC-V3R2-MIG-001 (v2→v3 migrator)
- SPEC-V3R2-WF-003 (multi-mode router consumes HarnessConfig + InterviewConfig)

### 9.3 Related (informational)

- SPEC-V3R2-HRN-001 (harness.yaml loader) — merged main; reconciliation in §1
- SPEC-V3R2-MIG-002 (hook registration cleanup) — merged main; established the audit_test.go style pattern
- SPEC-V3R2-EXT-001 (memory subsystem) — InterviewConfig.skip_conditions relates to memory staleness; no API contract between SPECs

## 10. Plan-Audit Readiness Checklist

- [x] All 18 REQs covered by tasks (§4 matrix)
- [x] All 15 ACs verifiable by automation (acceptance.md ≥10 binary)
- [x] HRN-001 reconciliation documented (§1, research.md §2)
- [x] File map exhaustive (§5)
- [x] Risks itemized with mitigations (§6)
- [x] OQs resolved with decisions (§7)
- [x] Test plan with sequence + coverage targets (§8)
- [x] No time estimates (priority labels only)
- [x] Drift-guard at each milestone exit
- [x] research.md anchor count ≥25 (actual: 38)
- [x] acceptance.md AC count ≥10 (actual: 15)

## 11. Plan Status

`plan_status: audit-ready` — plan ready for plan-auditor PASS gate before /moai run.

End of plan.md
