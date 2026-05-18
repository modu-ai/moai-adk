# SPEC-V3R2-MIG-003 — Research / Codebase Analysis (research.md)

> Companion to `spec.md` (REQ-MIG003-001 ~ REQ-MIG003-018).
> Companion to `plan.md` (M1-M4 milestones, post-HRN-001 scope).
> Companion to `acceptance.md` (binary AC) and `tasks.md` (T-MIG003-NN).

This document is the codebase analysis artifact required by `.claude/rules/moai/workflow/spec-workflow.md` Plan Phase. All anchors are concrete `file:line` references verified at HEAD `plan/SPEC-V3R2-MIG-003` ≈ `195b469d8` (post-MIG-002 sync merge).

---

## 1. Section-by-Section Loader Inventory (current state, 2026-05-18)

Cross-reference of `.moai/config/sections/*.yaml` against Go-side loaders. Source: `ls .moai/config/sections/`, `grep ^func .* internal/config/loader.go`, `grep -rE 'loadYAMLFile|yaml.Unmarshal' internal/` (filtered to non-test).

### 1.1 LOADED via `Loader.Load()` (internal/config/loader.go)

The aggregating `Loader.Load()` at `internal/config/loader.go:31-74` invokes ten section-loaders in fixed order. Each loader follows the `loadXxxSection(dir, cfg) → wrapper → loadYAMLFile() → loadedSections["xxx"]=true` pattern.

| Section | Loader func | line | Wrapper type | Target field |
|---|---|---|---|---|
| `user.yaml` | `loadUserSection` | 88-99 | `userFileWrapper` (types.go:562-564) | `cfg.User` |
| `language.yaml` | `loadLanguageSection` | 102-113 | `languageFileWrapper` (types.go:566-568) | `cfg.Language` |
| `quality.yaml` | `loadQualitySection` | 118-129 | `qualityFileWrapper` (types.go:572-574, "constitution" YAML key for Python BC) | `cfg.Quality` |
| `git-convention.yaml` | `loadGitConventionSection` | 132-143 | `gitConventionFileWrapper` (types.go:577-579) | `cfg.GitConvention` |
| `llm.yaml` | `loadLLMSection` | 146-157 | `llmFileWrapper` (types.go:582-584) | `cfg.LLM` |
| `state.yaml` | `loadStateSection` | 160-171 | `stateFileWrapper` (types.go:587-589) | `cfg.State` |
| `statusline.yaml` | `loadStatuslineSection` | 174-185 | `statuslineFileWrapper` (types.go:592-594) | `cfg.Statusline` |
| `ralph.yaml` | `loadRalphSection` | 190-207 | `ralphFileWrapper` (types.go:604-609, inline + StaleSeconds → Session) | `cfg.Ralph` + `cfg.Session.StaleSeconds` |
| `research.yaml` | `loadResearchSection` | 210-221 | `researchFileWrapper` (types.go:597-599) | `cfg.Research` |

### 1.2 LOADED via dedicated entry-point (NOT through `Loader.Load()`)

| Section | Loader func | File:line | Consumer |
|---|---|---|---|
| `harness.yaml` | `LoadHarnessConfig(path)` | internal/config/loader.go:256-311 | HarnessRouter, CLI validate, ConfigManager.Reload — **delivered by SPEC-V3R2-HRN-001 (2026-05-18, merged main)** |
| `runtime.yaml` | `LoadRuntime(path)` | internal/runtime/config.go:81-114 | Runtime context-window / circuit-breaker — separate package, NOT in `internal/config` |

### 1.3 NOT LOADED (template-only, gap targets for MIG-003)

| Section | YAML present at | Go struct? | Loader? | Hot path? |
|---|---|---|---|---|
| `constitution.yaml` | sections/constitution.yaml:1-32 | NO | NO | Template-only; CLAUDE.md §forbidden libraries pattern |
| `context.yaml` | sections/context.yaml:1-20 | NO | NO | Template-only; CLAUDE.md §16 Context Search Protocol |
| `interview.yaml` | sections/interview.yaml:1-14 | NO | NO | Template-only; SPEC-V3R2-WF-003 discovery |
| `design.yaml` | sections/design.yaml:1-60 | NO | NO (migrate-only) | Used by `internal/cli/migrate_agency.go` Phase 4 only — see §3.4 below |
| `sunset.yaml` | sections/sunset.yaml:1-19 | YES (types.go:311-325) | NO | DORMANT — struct defined but no `LoadSunsetConfig`, no runtime consumer |

### 1.4 NOT IN SCOPE (declared out-of-scope by spec.md §1.2 / §2.2)

The following sections are explicitly excluded from MIG-003 per spec.md §1.2 Non-Goals and §2.2 Out of Scope. Listed for completeness:

- `mx.yaml` — ad-hoc parsing retained; struct neuverbalisation deferred to separate SPEC (R6 §5.3 reference)
- `workflow.yaml` — only `role_profiles` is currently consumed (types.go:148-203 WorkflowConfig partial); full unification deferred to separate SPEC per spec.md §1.2 bullet 7
- `lsp.yaml`, `gate.yaml`, `db.yaml`, `github-actions.yaml`, `observability.yaml`, `project.yaml`, `memo.yaml`, `system.yaml`, `security.yaml`, `git-strategy.yaml` — not part of MIG-003 audit scope; tracked by separate SPECs

---

## 2. HRN-001 Reconciliation (harness.yaml is already loaded)

SPEC-V3R2-HRN-001 shipped to main on 2026-05-18 (commit ≤ `195b469d8`). It delivered the full harness.yaml loader stack that MIG-003 spec.md anticipated:

### 2.1 Delivered artifacts (HRN-001)

- **`internal/config/types.go:398-427`** — `HarnessConfig` struct with 9 fields: `DefaultProfile`, `ModeDefaults`, `AutoDetection`, `Escalation`, `EffortMapping`, `Levels`, `ModelUpgradeReview`, `PlanAuditGlobal`, `Evaluator`. `@MX:ANCHOR` REQ-HRN-001-001.
- **`internal/config/types.go:429-538`** — Sub-config types: `AutoDetectionConfig`, `AutoDetectionRule`, `EscalationConfig`, `LevelConfig`, `PlanAuditConfig`, `ModelUpgradeReviewConfig`, `ReviewChecklistItem`, `ModelUpgradeTrigger`, `ModelUpgradeOutput`, `PlanAuditGlobalConfig`, `EvaluatorConfig`.
- **`internal/config/loader.go:256-311`** — `LoadHarnessConfig(path) (*HarnessConfig, error)` with 4-stage validation: (1) struct unmarshal, (2) schema drift detection, (3) FROZEN `memory_scope == "per_iteration"` check, (4) FROZEN level enum check.
- **`internal/config/loader.go:223-244`** — `validHarnessLevels`, `knownHarnessTopLevelKeys` whitelists for drift gating.
- **`internal/config/loader.go:318-358`** — `detectSchemaDrift(data, path)` honors `MOAI_CONFIG_STRICT=1` env (returns `ErrSchemaDrift`) vs warn-only.
- **`internal/config/errors.go:69-90`** — Sentinels: `ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded` (REQ-HRN-001-013/017/018/019).
- **`internal/config/loader_harness_extended_test.go:17-200`** — 7 test cases: ValidParse, InvalidThreshold, UnknownLevel, SchemaDrift_Warn, SchemaDrift_StrictMode, MissingFile, ValidatesLevelEnum.
- **`internal/config/types.go:555-557`** — `harnessFileWrapper{Harness HarnessConfig yaml:"harness"}`.

### 2.2 What this means for MIG-003

| MIG-003 REQ | Status after HRN-001 | Action |
|---|---|---|
| REQ-MIG003-001 (5 structs) | 1 of 5 (HarnessConfig) already done; 4 remain (Constitution, Context, Interview, Design) | Implement 4 new structs |
| REQ-MIG003-002 (5 loaders) | 1 of 5 (`LoadHarnessConfig`) already done; 4 remain | Implement 4 new loaders |
| REQ-MIG003-003 (typed return + aggregated error) | Pattern established by HRN-001 (`ValidationError` wrapping) | Reuse pattern |
| REQ-MIG003-004 (defaults on absent file) | HRN-001 returns `ErrConfigNotFound`; MIG-003 needs **graceful default** semantics (different) | Use `loadYAMLFile`'s 2-return pattern; absent → defaults |
| REQ-MIG003-008 (`HARNESS_CONFIG_PARSE_ERROR`) | Existing: HRN-001 returns `ErrInvalidYAML` (not the exact sentinel name) | **Reconcile**: name harmonized to `ErrInvalidYAML` (which subsumes `_PARSE_ERROR`); AC text updated |
| REQ-MIG003-012 (default level=`standard`) | Partially: HRN-001 returns error on missing file. MIG-003 may add a `DefaultHarnessConfig()` helper for graceful path in `Loader.Load()` flow | Add `DefaultHarnessConfig()` and call from new `loadHarnessSection` wrapper |
| REQ-MIG003-013 (`YAML_SECTION_NO_LOADER` CI guard) | Not implemented by HRN-001 | NEW work in MIG-003 |
| REQ-MIG003-016 (`CONFIG_STRUCT_YAML_MISMATCH`) | Not implemented by HRN-001 | NEW work in MIG-003 |

**Net new scope post-HRN-001**: 4 unloaded sections (constitution, context, interview, design) + sunset DORMANT formalization + 2 CI guards. The harness work is **verify-only** in MIG-003.

---

## 3. YAML Schema Per Section (existing template content)

### 3.1 `constitution.yaml` (32 lines)

- **sections/constitution.yaml:1-32** — top-level key `constitution:`. Nested keys: `approved_frameworks: []` (line 2-4), `approved_languages: []` (5-6), `architecture.forbidden_dependencies: []` (7-9), `architecture.patterns: []` (10-13), `forbidden_patterns: []` (14-18), `naming_conventions{exported, files, packages}` (19-22), `performance{max_memory_mb, max_response_time_ms}` (23-25), `security.forbidden_practices: []` (26-29), `security.required_checks: []` (30-31).
- Hot path target: `ConstitutionConfig.ForbiddenLibraries` (alias of `forbidden_patterns` per spec.md REQ-MIG003-009) for runtime policy enforcement; SPEC-V3R2-EXT-004 hook consumes the list.

### 3.2 `context.yaml` (20 lines)

- **sections/context.yaml:1-19** — top-level key `context_search:` (note: NOT `context:`). Nested keys: `auto_detect.enabled` (2-3), `enabled` (4), `memory_integration{enabled, include_in_context, priority_over_search}` (5-8), `performance{cache_ttl_seconds, timeout_seconds}` (9-11), `search{date_range_days, max_results, max_tokens_per_result, project_scope_only}` (12-16), `token_budget{max_injection_tokens, skip_if_usage_above}` (17-19).
- Hot path target: CLAUDE.md §16 Context Search Protocol — `token_budget.max_injection_tokens` (5000) and `token_budget.skip_if_usage_above` (150000); also `search.date_range_days` (30) for staleness window.
- **Naming note**: spec.md uses `staleness_window_days` but YAML uses `search.date_range_days` — research will align field name to YAML (`DateRangeDays`) and document the alias in struct godoc.

### 3.3 `interview.yaml` (14 lines)

- **sections/interview.yaml:1-13** — top-level key `interview:`. Nested keys: `clarity_threshold` (2 — integer 4, NOT max_rounds — clarification of spec.md REQ-MIG003-011), `enabled` (3), `plan{max_rounds, questions_per_round}` (4-6), `project{max_rounds, questions_per_round}` (7-9), `skip_conditions: []` (10-13: `resume_spec_id_present`, `skip_interview_flag`, `technical_keywords_gte_5`).
- Hot path target: SPEC-V3R2-WF-003 discovery mode — `interview.plan.max_rounds` (5), `interview.plan.questions_per_round` (3), `interview.clarity_threshold` (4), `interview.skip_conditions` list.

### 3.4 `design.yaml` (60 lines)

- **sections/design.yaml:1-59** — top-level key `design:`. Nested keys: `adaptation{confidence_threshold, enabled, iteration_limits{builder, copywriter, designer}, min_projects_for_adaptation}` (2-9), `brand_context{dir, interview_on_first_run}` (10-12), `claude_design{enabled, fallback_path, supported_bundle_versions[]}` (13-17), `default_framework` (18), `design_docs{auto_load_on_design_command, dir, priority[], token_budget}` (19-27), `enabled` (28), `evaluator.memory_scope` (29-30 — note this is `per_iteration`, mirroring HarnessConfig FROZEN constraint), `evolution{archive_after_evolve, auto_evolve_threshold, cooldown_hours, graduation_criteria{...}, max_active_learnings, max_evolution_rate_per_week, require_approval}` (31-42), `figma.enabled` (43-44), `gan_loop{escalation_after, improvement_threshold, max_iterations, pass_threshold, sprint_contract{...}, strict_mode}` (45-58), `version` (59).
- Current consumer: **only** `internal/cli/migrate_agency.go` Phase 4 (lines containing `design.yaml` → 5 occurrences; reads-and-renders during `.agency/` migration). No runtime consumer outside migrate flow.
- Hot path target post-MIG-003: GAN loop runtime — `gan_loop.pass_threshold` (FROZEN floor 0.60 per design constitution §11), `gan_loop.sprint_contract.enabled`, `adaptation.phase_weights` (per spec.md REQ-MIG003-014; though current YAML uses `iteration_limits` per language naming).

### 3.5 `sunset.yaml` (19 lines, dormant)

- **sections/sunset.yaml:1-18** — top-level key `sunset:`. Nested keys: `conditions: []` (2-17) with three entries `{action, description, gate, metric, threshold}` (gates: vet@50, lint@30, test@20), `enabled: true` (18).
- Existing Go struct: **`internal/config/types.go:309-325`** — `SunsetConfig{Enabled bool, Conditions []SunsetCondition}` + `SunsetCondition{Gate, Metric, Threshold, Action, Description}`.
- **Gap**: struct exists but no `LoadSunsetConfig()` and no `cfg.Sunset` field on root `Config`. No hot path enforces sunset conditions — they are template-only.

---

## 4. Existing Loader Patterns (HRN-001 + Loader.Load reference)

### 4.1 Wrapper pattern (types.go)

Each loaded section has a `xxxFileWrapper` struct wrapping the target Config with the YAML top-level key as the yaml tag.

- **types.go:555-557** — `harnessFileWrapper{Harness HarnessConfig \`yaml:"harness"\`}`
- **types.go:562-564** — `userFileWrapper{User models.UserConfig \`yaml:"user"\`}`
- **types.go:572-574** — `qualityFileWrapper{Constitution models.QualityConfig \`yaml:"constitution"\`}` (BC: YAML key differs from struct semantics)
- **types.go:604-609** — `ralphFileWrapper` (composed: inline + extra field promoted to different Config field)

MIG-003 will add four such wrappers:
- `constitutionFileWrapper{Constitution ConstitutionConfig \`yaml:"constitution"\`}` — collision with `qualityFileWrapper.Constitution` (BC alias for `quality` field). Section file IS `.moai/config/sections/constitution.yaml`, distinct from `quality.yaml`. The YAML top-level key in both files is the literal word `constitution:` — disambiguation requires loading by **filename**, not by top-level key alone. The two wrappers have different Go types so unmarshalling is safe.
- `contextFileWrapper{ContextSearch ContextConfig \`yaml:"context_search"\`}` — note YAML key is `context_search`, not `context`.
- `interviewFileWrapper{Interview InterviewConfig \`yaml:"interview"\`}`
- `designFileWrapper{Design DesignConfig \`yaml:"design"\`}`

### 4.2 Loader function pattern (loader.go)

Each `loadXxxSection(dir, cfg)`:
1. Initialize wrapper with current default config field
2. Call `loadYAMLFile(dir, "xxx.yaml", wrapper)` — returns `(loaded bool, err error)`
3. On `err != nil`: `slog.Warn(...)` and `return` (graceful degradation)
4. On `loaded == true`: assign wrapper field back to `cfg.Xxx` and set `l.loadedSections["xxx"] = true`

Anchor: **internal/config/loader.go:88-99 (`loadUserSection`)** — canonical pattern. MIG-003 will replicate 4 times.

### 4.3 `loadYAMLFile` helper (loader.go:363-378)

Returns three-state semantics:
- `(true, nil)` — file exists and parsed cleanly → caller applies the wrapper
- `(false, nil)` — file does not exist → caller uses defaults (no warning) — **this is the REQ-MIG003-004 sensible-defaults path**
- `(false, err)` — file exists but malformed → caller logs warning (does not panic)

MIG-003 reuses this helper without modification.

### 4.4 NewDefaultConfig anchor (defaults.go)

- **internal/config/defaults.go** (10949 bytes total) — `NewDefaultConfig()` returns `*Config` with all field defaults pre-populated. MIG-003 will extend this function with `cfg.Constitution = defaultConstitutionConfig()`, `cfg.ContextSearch = defaultContextConfig()`, `cfg.Interview = defaultInterviewConfig()`, `cfg.Design = defaultDesignConfig()`. Each default helper returns a struct initialized from the values shipped in `internal/template/templates/.moai/config/sections/*.yaml` (consistency with template defaults — REQ-MIG003-004).

### 4.5 LoadedSections() map (loader.go:78-85)

Public accessor `Loader.LoadedSections() map[string]bool` returns a copy of which sections were loaded successfully. MIG-003 will add 4 new keys: `constitution`, `context_search`, `interview`, `design`. This is the surface that AC-MIG003-12 (CI guard) inspects.

---

## 5. Defaults Strategy (REQ-MIG003-004 enforcement)

Spec.md REQ-MIG003-004 mandates **sensible defaults when YAML file is absent**. Spec.md REQ-MIG003-012 hardcodes `level: standard` as default for harness. For the 4 new loaders MIG-003 introduces:

| Section | Default source | Default values (from template) |
|---|---|---|
| Constitution | `internal/template/templates/.moai/config/sections/constitution.yaml` (same content as project's `.moai/config/sections/constitution.yaml:1-32`) | `approved_languages: ["go"]`, `forbidden_patterns: ["global mutable state", ...]`, etc. |
| Context | template constitution.yaml | `token_budget.max_injection_tokens: 5000`, `skip_if_usage_above: 150000`, `search.date_range_days: 30`, etc. |
| Interview | template interview.yaml | `clarity_threshold: 4`, `enabled: true`, `plan.max_rounds: 5`, etc. |
| Design | template design.yaml | `gan_loop.pass_threshold: 0.75` (note: FROZEN floor 0.60), `gan_loop.max_iterations: 5`, etc. |

Defaults are encoded as Go literals in `internal/config/defaults.go` (REQ-MIG003-017 forbids hardcoded values overriding YAML, but defaults applied **when YAML is absent** are explicitly permitted by REQ-MIG003-004 — these are two distinct cases).

---

## 6. CI Guard Surface (REQ-MIG003-013 / REQ-MIG003-016)

### 6.1 `YAML_SECTION_NO_LOADER` (REQ-MIG003-013, AC-MIG003-12)

Audit test that detects YAML files in `.moai/config/sections/` lacking a corresponding loader. Pattern:

- Scan `internal/template/templates/.moai/config/sections/*.yaml` (or `os.ReadDir(sectionsDir)`)
- Strip `.yaml` extension → section name
- Cross-reference against the union of:
  - Known loaders in `internal/config/loader.go` (parsed via AST or static list)
  - Known dedicated loaders (e.g., `LoadHarnessConfig` at loader.go:256, `LoadRuntime` at internal/runtime/config.go:81)
  - Out-of-scope sections (declared in `internal/config/loader.go` package-level `var unloadedAcknowledgedSections = []string{...}` or similar registry)
- For each new YAML file not in any list → fail test with `YAML_SECTION_NO_LOADER: <file>` (REQ-MIG003-013)

Implementation file (proposed): `internal/config/audit_loader_completeness_test.go`. Cross-reference test pattern: **internal/template/agentless_audit_test.go** (existing CI guard model).

### 6.2 `CONFIG_STRUCT_YAML_MISMATCH` (REQ-MIG003-016, AC-MIG003-14)

`make build`-time validation: Go struct field defined in `ConstitutionConfig` MUST have a matching key in the template YAML (and vice versa). Detection:

- AST-walk Go struct fields with yaml tags (e.g., `ApprovedFrameworks []string \`yaml:"approved_frameworks"\``)
- Parse template YAML file into `map[string]any`
- Recursive compare: every yaml-tagged Go field has a key in the YAML map, and every YAML key has a yaml-tagged Go field (modulo `omitempty` and `,inline`)
- On mismatch: emit `CONFIG_STRUCT_YAML_MISMATCH: field=<path>, side=<go-only|yaml-only>` and fail

Implementation file (proposed): `internal/config/audit_struct_yaml_symmetry_test.go`. Run as `go test ./internal/config/...` and gated by `make build`.

---

## 7. SunsetConfig DORMANT Strategy (REQ-MIG003-006, REQ-MIG003-018)

Per spec.md §2.1 and §1.2, sunset.yaml stays dormant. Required deliverables:

### 7.1 DORMANT godoc marker (REQ-MIG003-006, AC-MIG003-06)

Modify `internal/config/types.go:309` block to prepend explicit DORMANT marker:

```
// SunsetConfig defines the Build-to-Delete framework configuration.
// Quality gates that consistently pass can be relaxed over time.
//
// @MX:NOTE: [AUTO] DORMANT — struct schema defined but no runtime hot path enforces
//                  sunset conditions. Activation deferred to a future SPEC.
//                  Do NOT add LoadSunsetConfig until activation SPEC is filed.
type SunsetConfig struct {
```

### 7.2 Once-per-session log (REQ-MIG003-018, AC-MIG003-15)

When `Loader.Load()` encounters `sunset.yaml` (existence check, not parse), emit `slog.Info("SUNSET_CONFIG_DORMANT_NOTICE", "spec", "SPEC-V3R2-MIG-003 REQ-018", "yaml_path", path)`. Guarded by a `sync.Once` so it fires at most once per process lifetime. Implementation lives in a new helper `internal/config/sunset_notice.go` (small file, no test ambiguity).

---

## 8. Related SPECs (dependency / overlap analysis)

| Related SPEC | Status | Relationship |
|---|---|---|
| SPEC-V3R2-HRN-001 | merged 2026-05-18 | **Delivers harness.yaml loader** — MIG-003 reconciliation point (§2 above). Effective scope reduced from 5 → 4 sections. |
| SPEC-V3R2-MIG-002 | merged 2026-05-18 | Hook registration cleanup. Established `audit_test.go` style 3-way sync test pattern that MIG-003 §6 CI guard mirrors. |
| SPEC-V3R2-EXT-004 | blocked-by | Migration framework — MIG-003 loader completeness step runs after EXT-004's migrate command. Reference: spec.md §9.1. |
| SPEC-V3R2-WF-003 | downstream | Multi-mode router consumes `InterviewConfig` and `HarnessConfig`. MIG-003 unblocks this consumption. |
| SPEC-V3R2-EXT-001 | related | Memory subsystem; `InterviewConfig.skip_conditions[resume_spec_id_present]` relates to memory staleness. |

---

## 9. Test Fixture Plan (for M1 RED phase)

Each new loader requires fixtures parallel to HRN-001's `internal/config/testdata/`:

```
internal/config/testdata/
├── constitution-valid/constitution.yaml            (mirror sections/constitution.yaml)
├── constitution-malformed/constitution.yaml        (truncated mid-map for parse error path)
├── context-valid/context.yaml                       (mirror sections/context.yaml)
├── context-malformed/context.yaml                   (invalid YAML, top-level key mistyped)
├── interview-valid/interview.yaml                   (mirror sections/interview.yaml)
├── interview-missing/                               (empty dir for default-path test)
├── design-valid/design.yaml                         (mirror sections/design.yaml)
├── design-pass-threshold-violation/design.yaml      (pass_threshold: 0.45 — below FROZEN 0.60 floor; loader should reject if validation enforces; otherwise loader accepts and downstream WF enforces)
```

Existing fixture directory pattern reference: `ls internal/config/testdata/` shows the testdata convention is in active use by HRN-001 tests (loader_harness_extended_test.go:17-200).

---

## 10. File Map (write/edit targets for plan.md)

### 10.1 Files to CREATE (new)

| File | Purpose |
|---|---|
| `internal/config/loader_constitution.go` | `ConstitutionConfig` consumers + `LoadConstitutionConfig` + `loadConstitutionSection` |
| `internal/config/loader_context.go` | `ContextConfig` + `LoadContextConfig` + `loadContextSection` |
| `internal/config/loader_interview.go` | `InterviewConfig` + `LoadInterviewConfig` + `loadInterviewSection` |
| `internal/config/loader_design.go` | `DesignConfig` + `LoadDesignConfig` + `loadDesignSection` |
| `internal/config/sunset_notice.go` | One-shot session log for REQ-MIG003-018 |
| `internal/config/loader_constitution_test.go` | Unit tests: valid / missing / malformed |
| `internal/config/loader_context_test.go` | Unit tests |
| `internal/config/loader_interview_test.go` | Unit tests |
| `internal/config/loader_design_test.go` | Unit tests |
| `internal/config/audit_loader_completeness_test.go` | CI guard `YAML_SECTION_NO_LOADER` (REQ-013) |
| `internal/config/audit_struct_yaml_symmetry_test.go` | CI guard `CONFIG_STRUCT_YAML_MISMATCH` (REQ-016) |
| `internal/config/testdata/{constitution,context,interview,design}-{valid,malformed}/...` | Fixtures (10 yaml files) |

### 10.2 Files to EDIT (existing)

| File | Change |
|---|---|
| `internal/config/types.go` | Add 4 new structs (`ConstitutionConfig`, `ContextConfig`, `InterviewConfig`, `DesignConfig`) — placed adjacent to `HarnessConfig` (types.go:402). Add 4 fields to root `Config` struct (types.go:15-33). Prepend DORMANT godoc to `SunsetConfig` (types.go:309). Add 4 wrapper types (types.go:560 area). |
| `internal/config/loader.go` | Wire 4 new `loadXxxSection(sectionsDir, cfg)` calls into `Loader.Load()` (loader.go:31-74) — append after `loadResearchSection` call at line 71. Optionally also invoke `sunsetDormantNoticeOnce(sectionsDir)`. |
| `internal/config/defaults.go` | Add 4 default helpers (`defaultConstitutionConfig`, etc.) and wire into `NewDefaultConfig()`. |
| `internal/config/errors.go` | Reuse existing `ErrInvalidYAML`, `ErrConfigNotFound`. May add `ErrYAMLSectionNoLoader` for REQ-013 CI guard signal. |
| `.claude/rules/moai/core/settings-management.md` | Document the 4 new loaders + DORMANT sunset.yaml + CI guards (per spec.md §10 Traceability bullet 6). |

### 10.3 Files NOT in scope

- `internal/template/templates/.moai/config/sections/*.yaml` — template defaults are READ, not modified. (Out-of-scope per spec.md §1.2.)
- `internal/runtime/config.go` — separate loader for `runtime.yaml`; not affected.
- `internal/cli/migrate_agency.go` — current `design.yaml` consumer; remains as-is per spec.md §2.1 design.yaml row ("ADD LOADER for runtime; migrate_agency.go stays").

---

## 11. Risk Analysis (deep)

### 11.1 Wrapper-key collision (constitution.yaml vs quality.yaml)

Both YAML files use the literal top-level key `constitution:`. The existing `qualityFileWrapper.Constitution` (types.go:573) targets `quality.yaml`'s `constitution:` key as a Python-BC alias for the QualityConfig data. The new `constitutionFileWrapper.Constitution` targets `constitution.yaml`'s `constitution:` key with ConstitutionConfig schema. **Resolution**: `loadYAMLFile(dir, "<filename>", wrapper)` keys off filename (`"quality.yaml"` vs `"constitution.yaml"`) — the YAML key collision is harmless because each loader unmarshals into its own typed wrapper. Verify via test that loading both does not cross-pollute fields.

### 11.2 Context YAML key naming mismatch

`context.yaml` uses top-level key `context_search:` (not `context:`). The wrapper must use `\`yaml:"context_search"\`` to match. Naming the wrapper field `ContextSearch` (not `Context`) avoids surprise; the public `ContextConfig` struct stays clean.

### 11.3 `design.yaml` ↔ migrate_agency.go double-consumption

`migrate_agency.go` Phase 4 reads `design.yaml` during a one-time migration (.agency/ → .moai/). Adding `DesignConfig` runtime loader does NOT change migration behavior — migrate phase writes the file, runtime reads it. Both flows can coexist.

### 11.4 FROZEN constraints in `design.yaml`

design.yaml line 30 declares `evaluator.memory_scope: per_iteration`, mirroring the FROZEN constraint validated by `LoadHarnessConfig` (loader.go:281-295). The `LoadDesignConfig` loader **must NOT** enforce this constraint (out of scope); enforcement stays in harness loader. Design loader treats `memory_scope` as a free string field.

### 11.5 Pass-threshold floor (FROZEN 0.60) in `design.yaml.gan_loop.pass_threshold`

design.yaml line 49 has `pass_threshold: 0.75` (above FROZEN floor 0.60). `LoadDesignConfig` SHOULD validate `gan_loop.pass_threshold >= 0.60` and return `ErrPassThresholdFloor` (reusing the existing sentinel from HRN-001 errors.go:79) on violation. This is a runtime-enforcement parallel to HRN-001's harness floor.

### 11.6 Schema drift for new loaders

HRN-001's `detectSchemaDrift` (loader.go:318-358) is hardcoded for `harness:` top-level. MIG-003 will NOT replicate `MOAI_CONFIG_STRICT` drift detection for the 4 new sections in the initial scope — risk-mitigated via the `CONFIG_STRUCT_YAML_MISMATCH` build-time guard (§6.2). Runtime drift detection is a future enhancement, tracked separately.

### 11.7 sunset.yaml deletion edge case (REQ-MIG003-006 mitigation)

If user removes sunset.yaml, the `sunsetDormantNoticeOnce` helper must NOT panic. Implementation uses `os.Stat` graceful path — file absent → log nothing (no notice fires) — file present → fire once. No `LoadSunsetConfig` exists, so deletion has no other failure surface.

---

## 12. Code Quality Compliance

### 12.1 9-direct-dep policy

spec.md §7 mandates using only `yaml.v3` + stdlib. Inventory:

- `gopkg.in/yaml.v3` — already in `go.mod` (used by loader.go:11)
- `github.com/go-playground/validator/v10` — already in `go.mod` (used by HRN-001). MIG-003 MAY use struct tags for `validator:"required,gte=0,lte=1"` style invariants, consistent with HRN-001 precedent.

No new direct deps required.

### 12.2 Thread-safety (spec.md §7 line 198)

The existing `Loader` has `mu sync.RWMutex` (loader.go:18-21) and `Load()` acquires write lock (loader.go:32-33). New `loadXxxSection` helpers inherit the lock from `Load()` caller — no extra mutex needed in section helpers (mirrors HRN-001 pattern).

Public `LoadConstitutionConfig(path)` etc. are pure functions (no shared state) — thread-safe by design.

### 12.3 No-hardcoded-values (CLAUDE.local.md §14 / REQ-MIG003-017)

Loaders MUST read defaults from `defaults.go` helpers, not inline literals. `internal/config/envkeys.go:1-100` is the env-var registry — no MIG-003 changes needed there. Acceptance test `audit_loader_completeness_test.go` will grep loader bodies for literal numeric constants matching YAML field values; any match is a `LOADER_HARDCODE_VIOLATION` (AC-MIG003-13).

---

## 13. Section Inventory Summary Table

| # | YAML file | Struct exists? | Loader exists? | Loaded by Load()? | MIG-003 action |
|---|---|---|---|---|---|
| 1 | user.yaml | ✓ (models) | ✓ (loadUserSection) | ✓ | unchanged |
| 2 | language.yaml | ✓ (models) | ✓ | ✓ | unchanged |
| 3 | quality.yaml | ✓ (models) | ✓ | ✓ | unchanged |
| 4 | git-convention.yaml | ✓ (models) | ✓ | ✓ | unchanged |
| 5 | llm.yaml | ✓ | ✓ | ✓ | unchanged |
| 6 | state.yaml | ✓ | ✓ | ✓ | unchanged |
| 7 | statusline.yaml | ✓ (models) | ✓ | ✓ | unchanged |
| 8 | ralph.yaml | ✓ | ✓ | ✓ | unchanged |
| 9 | research.yaml | ✓ | ✓ | ✓ | unchanged |
| 10 | harness.yaml | ✓ (HRN-001) | ✓ LoadHarnessConfig (HRN-001) | NO (dedicated entry) | **MIG-003 wire `loadHarnessSection` into Load()** OR document HRN-001's dedicated entry as intentional — TBD M4 decision |
| 11 | runtime.yaml | ✓ (separate pkg) | ✓ LoadRuntime | NO (separate pkg) | unchanged (out of internal/config) |
| 12 | constitution.yaml | ✗ | ✗ | ✗ | **ADD** |
| 13 | context.yaml | ✗ | ✗ | ✗ | **ADD** |
| 14 | interview.yaml | ✗ | ✗ | ✗ | **ADD** |
| 15 | design.yaml | ✗ | ✗ (migrate-only) | ✗ | **ADD** runtime loader |
| 16 | sunset.yaml | ✓ (types.go:311) | ✗ | ✗ | **DORMANT marker + once-per-session log** |
| 17 | workflow.yaml | partial | partial | NO | out-of-scope (spec.md §1.2) |
| 18 | mx.yaml | ✗ | ✗ | ✗ | out-of-scope (spec.md §2.2) |
| 19 | lsp.yaml | ✗ | ✗ | ✗ | out-of-scope |
| 20 | gate.yaml | ✓ (GateConfig types.go:252) | ✗ | ✗ | out-of-scope (separate SPEC) |
| 21 | db.yaml | ✗ | ✗ | ✗ | out-of-scope |
| 22 | github-actions.yaml | ✗ | ✗ | ✗ | out-of-scope |
| 23 | observability.yaml | ✗ | ✗ | ✗ | out-of-scope |
| 24 | project.yaml | ✓ (models) | ✗ (via separate loader) | ✗ | out-of-scope |
| 25 | memo.yaml | ✗ | ✗ | ✗ | out-of-scope |
| 26 | system.yaml | ✓ (SystemConfig types.go:58) | partial | partial | out-of-scope |
| 27 | security.yaml | ✓ (SecuritySandbox types.go:174) | partial | partial | out-of-scope |
| 28 | git-strategy.yaml | ✓ (types.go:36) | partial | partial | out-of-scope |

**Net MIG-003 gap**: 4 new loaders (rows 12-15) + 1 DORMANT formalization (row 16) + 2 CI guards (covering rows 12-16). Harness.yaml (row 10) is already loaded via dedicated entry; MIG-003 may decide in M4 whether to additionally wire it through `Loader.Load()` for uniformity, or leave HRN-001's dedicated entry as canonical.

---

## 14. Open Questions (resolve in M2/M3 of plan)

1. **Should `harness.yaml` be additionally wired into `Loader.Load()` for uniformity?**
   - Pro: Uniform `loadedSections["harness"] = true` accounting; one entry point for `Loader.Load()` users.
   - Con: HRN-001 deliberately keeps `LoadHarnessConfig` as a dedicated function with stricter semantics (`ErrConfigNotFound` instead of graceful-default). Wiring through `Load()` would require a wrapper that calls `LoadHarnessConfig` and downgrades errors. Possibly more friction than value.
   - **Recommendation**: Leave HRN-001's dedicated entry. Document in settings-management.md that `harness.yaml` uses dedicated `LoadHarnessConfig` due to FROZEN validation, and `Loader.Load()` consumers should call `LoadHarnessConfig` separately. AC-MIG003-12 audit test treats `harness.yaml` as "intentionally outside Loader.Load() — loader exists, just not in main Load() chain."

2. **Should `ContextConfig.ForbiddenLibraries` (or `ConstitutionConfig.ForbiddenLibraries`) actually expose for enforcement at MIG-003, or stub it?**
   - spec.md REQ-MIG003-009 says "expose the list for runtime policy enforcement (consumed by SPEC-V3R2-EXT-004 framework optional hook)". MIG-003 implements the LOADER and EXPOSURE; actual enforcement is downstream (EXT-004). Acceptable to ship just the typed field + godoc reference to SPEC-V3R2-EXT-004.

3. **Pass-threshold floor in `design.yaml`?**
   - design.yaml line 49: `pass_threshold: 0.75`. Floor is FROZEN at 0.60 (HRN-001 errors.go:79 `ErrPassThresholdFloor`). Should `LoadDesignConfig` enforce this? **Recommended yes** — reuse the existing sentinel, validate at load time. Symmetrical with harness loader.

4. **`design.yaml.adaptation.phase_weights` field — spec.md REQ-MIG003-014 mentions it, but the actual YAML has `adaptation.iteration_limits` instead.**
   - This is a spec.md drift (REQ wrote `phase_weights` but YAML has `iteration_limits`). Resolution: implement `IterationLimits` struct field matching YAML; document in struct godoc that `phase_weights` is the conceptual term spec.md uses; add an `IterationLimits` getter returning the data structure.

---

## 15. Anchors Inventory (≥25 file:line — total 38)

Counted from §§1-13 above:

1. internal/config/loader.go:31-74 (Loader.Load entry)
2. internal/config/loader.go:88-99 (loadUserSection canonical pattern)
3. internal/config/loader.go:102-113 (loadLanguageSection)
4. internal/config/loader.go:118-129 (loadQualitySection BC alias)
5. internal/config/loader.go:132-143 (loadGitConventionSection)
6. internal/config/loader.go:146-157 (loadLLMSection)
7. internal/config/loader.go:160-171 (loadStateSection)
8. internal/config/loader.go:174-185 (loadStatuslineSection)
9. internal/config/loader.go:190-207 (loadRalphSection)
10. internal/config/loader.go:210-221 (loadResearchSection)
11. internal/config/loader.go:223-244 (validHarnessLevels + knownHarnessTopLevelKeys)
12. internal/config/loader.go:256-311 (LoadHarnessConfig — HRN-001 reference)
13. internal/config/loader.go:318-358 (detectSchemaDrift — HRN-001 drift gate)
14. internal/config/loader.go:363-378 (loadYAMLFile helper)
15. internal/config/types.go:15-33 (root Config struct — extension target)
16. internal/config/types.go:309-325 (SunsetConfig DORMANT target)
17. internal/config/types.go:398-427 (HarnessConfig — HRN-001 delivered)
18. internal/config/types.go:429-538 (Harness sub-types — HRN-001 delivered)
19. internal/config/types.go:555-557 (harnessFileWrapper)
20. internal/config/types.go:562-609 (wrapper pattern reference)
21. internal/config/errors.go:14-90 (sentinel error patterns)
22. internal/config/errors.go:69-90 (HRN-001 harness sentinels)
23. internal/config/defaults.go (NewDefaultConfig — extend target)
24. internal/config/loader_harness_extended_test.go:17-200 (HRN-001 test pattern reference)
25. internal/runtime/config.go:78-114 (LoadRuntime — alternate pattern, NOT extended)
26. internal/cli/migrate_agency.go (design.yaml current sole consumer)
27. .moai/config/sections/constitution.yaml:1-32 (in-scope schema source)
28. .moai/config/sections/context.yaml:1-20 (in-scope schema source)
29. .moai/config/sections/interview.yaml:1-14 (in-scope schema source)
30. .moai/config/sections/design.yaml:1-60 (in-scope schema source)
31. .moai/config/sections/sunset.yaml:1-19 (DORMANT formalization target)
32. internal/template/agentless_audit_test.go (CI guard pattern reference for §6.1)
33. internal/config/envkeys.go:1-100 (env-var registry — no MIG-003 changes)
34. .claude/rules/moai/core/settings-management.md (doc target — spec.md §10)
35. go.mod (validator/v10 v10.30.2 + yaml.v3 — 9-direct-dep compliance)
36. internal/config/manager.go (ConfigManager.Reload — downstream consumer awareness)
37. internal/config/audit_registry.go (existing audit registry pattern reference)
38. internal/config/required_checks.go (alternate YAML consumer reference)

---

End of research.md
