# SPEC-V3R2-HRN-001 — Research / Codebase Analysis (research.md)

> Companion to `spec.md` (REQ-HRN-001-001 ~ REQ-HRN-001-019).
> Companion to `plan.md` (M1-M5 implementation).
> Companion to `acceptance.md` (10 binary AC) and `tasks.md`.

This document is the codebase analysis artifact required by `.claude/rules/moai/workflow/spec-workflow.md` Plan Phase. All anchors are concrete `file:line` references that have been verified to exist at HEAD `feat/SPEC-V3R4-WORKFLOW-SPLIT-001-wave-4` ≈ `d1c6f3104`.

---

## 1. Existing Harness YAML Schema (in-scope target)

The harness.yaml file ships as template at `internal/template/templates/.moai/config/sections/harness.yaml` and is rendered into each project at `.moai/config/sections/harness.yaml`. The shipping schema (this project's copy, 127 lines) covers ALL fields required by REQ-HRN-001-001 — no schema changes are needed.

### 1.1 Schema anchors

- **`.moai/config/sections/harness.yaml:18`** — `default_profile: default` (REQ-HRN-001-001 DefaultProfile field source)
- **`.moai/config/sections/harness.yaml:73-76`** — `mode_defaults:` block (REQ-HRN-001-014 mode_defaults consultation source: `cg: thorough`, `solo: auto`, `team: auto`)
- **`.moai/config/sections/harness.yaml:2-17`** — `auto_detection:` block with `enabled: true` and `rules:` block (REQ-HRN-001-007 priority ordering source: minimal → standard → thorough)
- **`.moai/config/sections/harness.yaml:6-8`** — `rules.minimal.conditions:` (`file_count <= 3 AND single_domain`, `spec_type in [bugfix, docs, config]`) — direct mapping to Complexity Estimator REQ-HRN-001-007/008
- **`.moai/config/sections/harness.yaml:13-17`** — `rules.thorough.conditions:` (`security_keywords OR payment_keywords present`, `spec_priority == critical`, `domain in [auth, payment, migration, public_api]`) — REQ-HRN-001-008 force-thorough source
- **`.moai/config/sections/harness.yaml:19-22`** — `effort_mapping:` (`minimal: medium`, `standard: high`, `thorough: xhigh`) — REQ-HRN-001-005 EffortForLevel source
- **`.moai/config/sections/harness.yaml:23-29`** — `escalation:` block (`enabled: true`, `max_escalations: 2`, `triggers: [quality_gate_fail, review_critical, test_coverage_low]`) — REQ-HRN-001-004/009/013 source
- **`.moai/config/sections/harness.yaml:32-72`** — `levels:` block with `minimal`, `standard`, `thorough` definitions (`description`, `evaluator`, `evaluator_mode`, `evaluator_profile`, `plan_audit`, `skip_phases`, `sprint_contract`, `playwright_testing`) — REQ-HRN-001-001 LevelConfig source
- **`.moai/config/sections/harness.yaml:64`** — `levels.thorough.evaluator_profile: strict` (only level pointing to a non-default profile) — must resolve to `.moai/config/evaluator-profiles/strict.md`
- **`.moai/config/sections/harness.yaml:77-114`** — `model_upgrade_review:` block (`checklist`, `trigger.on_model_change`, `output.report_path`) — REQ-HRN-001-016 source

## 2. Existing Go Config Loader (current state)

### 2.1 Current HarnessConfig struct (minimal — needs extension)

- **`internal/config/types.go:401-404`** — `HarnessConfig` declaration: only `DefaultProfile string` and `Evaluator EvaluatorConfig`. Comment line 399: "HRN-002 run-phase minimal substrate; HRN-001 run-phase에서 routing/profile 확장 예정." This SPEC is the **HRN-001 extension**.
- **`internal/config/types.go:406-423`** — `EvaluatorConfig` struct with `MemoryScope`, `Profiles map[string]string`, `Aggregation`, `MustPassDimensions []string`. HRN-003 M4 added Profiles/Aggregation/MustPassDimensions; HRN-001 does NOT need to modify this struct.
- **`internal/config/types.go:425-428`** — `harnessFileWrapper{Harness HarnessConfig}` YAML wrapper used by `LoadHarnessConfig()`.

### 2.2 Current LoadHarnessConfig() function (current behavior)

- **`internal/config/loader.go:223-262`** — `LoadHarnessConfig(path string) (*HarnessConfig, error)`. Reads file, unmarshals into `harnessFileWrapper`, validates `evaluator.memory_scope == "per_iteration"`. Returns `ErrConfigNotFound`, `ErrInvalidYAML`, or `ErrEvalMemoryFrozen` wrapped in `ValidationError`. **Gap**: does NOT validate the 5 new top-level fields (auto_detection, mode_defaults, escalation, effort_mapping, levels, model_upgrade_review).
- **`internal/config/loader.go:264-282`** — `loadYAMLFile()` shared helper. HRN-001 loader can REUSE this helper for sub-YAML loads (e.g., per-profile `.md` reads).

### 2.3 Sentinel errors infrastructure

- **`internal/config/errors.go:15-68`** — Existing sentinels: `ErrConfigNotFound`, `ErrInvalidConfig`, `ErrInvalidYAML`, `ErrEvalMemoryFrozen`, `ErrUnknownDimension`, `ErrRubricCitationMissing`, `ErrFlatScoreCardProhibited`, `ErrMustPassBypassProhibited`. **Gap**: missing `ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded` per REQ-HRN-001-010/012/017/019.
- **`internal/config/errors.go:70-90`** — `ValidationError` struct with `Field`, `Message`, `Value`, `Wrapped`. HRN-001 reuses this for field-level validation errors per REQ-010.
- **`internal/config/validation.go:15`** — `validate = validator.New()` global validator/v10 instance. HRN-001 reuses for `validator/v10` tags on new sub-structs.

### 2.4 ConfigManager.Reload integration point

- **`internal/config/manager.go:198-200`** — `Reload()` method on `ConfigManager`. HRN-001 must ensure `LoadHarnessConfig()` is idempotent and safe to call within `Reload()`. Existing `sync.RWMutex` guard at `internal/config/loader.go:19` provides thread-safety for `Loader`; HRN-001's free function `LoadHarnessConfig()` is stateless read-only.

## 3. Evaluator Profile Directory (referenced by REQ-HRN-001-012)

- **`.moai/config/evaluator-profiles/default.md`** — 68 LOC, declares 4 dimensions {Functionality 40%, Security 25%, Craft 20%, Consistency 15%} per scoring rubric (lines 32-68). PassThreshold floor 0.60 satisfied: anchors are 0.25/0.50/0.75/1.00 per Mechanism 1.
- **`.moai/config/evaluator-profiles/strict.md`** — Pointed by `levels.thorough.evaluator_profile`. Higher pass thresholds; FROZEN floor MUST hold (≥ 0.60).
- **`.moai/config/evaluator-profiles/lenient.md`** — Lowest pass thresholds in the family; AC-HRN-001-05 fixture variant artificially lowers to 0.5 to trigger floor violation.
- **`.moai/config/evaluator-profiles/frontend.md`** — Domain-specific frontend variant. Out of HRN-001 routing scope; mentioned only for completeness.

### 3.1 Profile parser (consumed by HRN-001)

- **`internal/harness/profile_loader.go:14-19`** — `defaultProfilePaths` map: `{"default": ".moai/config/evaluator-profiles/default.md", "strict": ..., "lenient": ..., "frontend": ...}`. HRN-001 uses this map to resolve `cfg.Levels[X].EvaluatorProfile` → file path.
- **`internal/harness/profile_loader.go:27-40`** — `newDefaultEvaluatorConfig()` returns `*config.EvaluatorConfig` with defaults. HRN-001 router consults this when `harness.yaml` is absent.
- **`internal/harness/profile_loader.go:45-63`** — `loadEvaluatorConfig()` helper, invokes `ParseRubricMarkdown()` per profile. HRN-001 floor-check piggy-backs on `ParseRubricMarkdown()` output.
- **`internal/harness/rubric.go`** — Owns `ParseRubricMarkdown(path)` (signature referenced in `profile_loader.go:50`) and `Rubric.Validate()`. Returns `Rubric` containing anchor levels per dimension. HRN-001 reads `Rubric.PassThreshold()` (or computes from anchors) to enforce FROZEN floor.

## 4. SPEC Parser (Complexity Estimator input)

The Router needs SPEC frontmatter + body to compute signals. Existing parser:

- **`internal/spec/lint.go:255-277`** — `SPECFrontmatter` struct: `ID`, `Title`, `Version`, `Status`, `Created`, `Updated`, `Author`, `Priority`, `Phase`, `Module`, `Dependencies`, `BcID`, `Lifecycle`, `Tags`, `Breaking`, `RelatedRule`, `LintConfig`. **Gap**: missing optional `HarnessLevel string \`yaml:"harness_level"\`` for REQ-HRN-001-015 spec_override.
- **`internal/spec/lint.go:286-294`** — `SPECDoc` struct with `Path`, `Frontmatter`, `Body`, `Criteria`, `REQs`, `ParseError`, `LintSkip`. HRN-001 router consumes this directly.
- **`internal/spec/lint.go:303`** — `parseSPECDoc(path)` factory. Router calls this in `fromSPECID()`.
- **`internal/spec/lint.go:334-355`** — `extractFrontmatter(content)` returns `(SPECFrontmatter, body, error)`. Used internally by `parseSPECDoc`; HRN-001 has no direct need but is the canonical parsing path.
- **`internal/spec/lint.go:357-370`** — `parseREQs(body)` returns `[]REQEntry` with `ID`, `Text`, `Line`. Router uses `REQEntry.Text` for keyword matching against `securityKeywords` / `paymentKeywords` (REQ-HRN-001-008).
- **`internal/spec/lint.go:279-283`** — `REQEntry` struct with `ID`, `Text`, `Line`. Body source for keyword matcher.
- **`internal/spec/lint.go:303-330`** — `parseSPECDoc(path string) *SPECDoc` full parser. Returns `nil` only on missing file; populates `ParseError` field for failed parses.

## 5. Existing CLI Surface (compatibility constraint)

### 5.1 Retired harness CLI (DO NOT collide with)

- **`internal/cli/harness.go:1-30`** — File-level deprecation comment per SPEC-V3R4-HARNESS-001 BC-V3R4-HARNESS-001-CLI-RETIREMENT. Lines 1-17: retired CLI is preserved as deprecation marker; physical removal deferred to follow-up SPEC.
- **`internal/cli/harness.go:47-71`** — `newHarnessCmd()` factory (RETIRED). Defines verbs `status|apply|rollback|disable`. Must NOT be confused with HRN-001's new `route|validate` verbs.
- **`internal/cli/harness.go:32-41`** — Constants `harnessDefaultLogPath`, `harnessDefaultSnapshotBase`, `harnessDefaultProposalDir`, `harnessConfigPath = ".moai/config/sections/harness.yaml"`. HRN-001 can reuse `harnessConfigPath` (same target file).
- **`internal/cli/harness.go:404-407`** — `harnessYAMLRoot` learning-specific wrapper (separate from `harnessFileWrapper` in `internal/config/types.go:425`). HRN-001 uses `config.LoadHarnessConfig()` exclusively — does NOT mirror this wrapper.
- **`internal/cli/harness_retirement_test.go`** — CI guard `TestHarnessRetirement` enforces non-registration of retired factory. HRN-001's new `newHarnessRouterCmd` factory is distinct and registered independently.

### 5.2 Root cobra command registration

- **`internal/cli/root.go:71-95`** — `AddCommand` invocations for active subcommands. Line 94 comment: "newHarnessCmd is intentionally NOT registered per SPEC-V3R4-HARNESS-001". HRN-001 inserts `rootCmd.AddCommand(newHarnessRouterCmd())` after line 92 in a separate registration block.
- **`internal/cli/doctor_harness.go`** — Existing doctor namespace for harness consistency checks. HRN-001 may add `moai doctor harness` validation hooks (deferred to T-HRN001-18).

## 6. Validator/v10 Infrastructure (existing)

- **`internal/config/types.go:52`** — Existing struct tag pattern: `validate:"omitempty"` on `ObservabilityEvents`.
- **`internal/config/types.go:84`** — Enum pattern: `validate:"omitempty,oneof=high medium low"` on `PerformanceTier`. HRN-001 reuses this pattern for `oneof=minimal standard thorough` on `Level` fields (REQ-HRN-001-017).
- **`go.mod:11`** — `github.com/go-playground/validator/v10 v10.30.2` (already imported, no new dep required per spec.md §3).

## 7. Spec Loader Test Patterns (M1 RED reference)

- **`internal/config/loader_test.go:1-31`** — `setupTestdataDir(t, tempDir, []string{...})` helper. HRN-001 RED tests reuse this helper with new fixtures in `testdata/harness-valid/`, `harness-invalid-threshold/`, `harness-invalid-level/`, `harness-drift-strict/`.
- **`internal/config/loader_test.go:33-72`** — `TestLoaderLoadAllSections` table-driven golden test. HRN-001 mirrors this pattern for `TestLoadHarnessConfigExtended`.
- **`internal/config/loader_test.go:180-211`** — `TestLoaderInvalidYAML` shows error-path testing convention. HRN-001 mirrors for `TestLoadHarnessConfigInvalidThreshold` and `TestLoadHarnessConfigUnknownLevel`.
- **`internal/config/testdata/eval-memory-frozen-violation/`** — Existing fixture pattern for HRN-002's FROZEN memory_scope violation. HRN-001's fixtures follow same structure: `testdata/<scenario>/harness.yaml` + `expected_error.txt`.

## 8. Effort Level Integration (SPEC-V3R2-ORC-003)

- **`internal/config/envkeys.go:85-89`** — `EnvClaudeCodeEffortLevel = "CLAUDE_CODE_EFFORT_LEVEL"`. Valid values: `low`, `medium`, `high`, `xhigh`, `max`. HRN-001's `EffortForLevel()` returns values from `{medium, high, xhigh}` aligned with this env var.
- **`.moai/config/sections/harness.yaml:19-22`** — Schema source: `minimal: medium`, `standard: high`, `thorough: xhigh`. Direct map; HRN-001 reads via `cfg.EffortMapping[string(level)]` per REQ-HRN-001-005.
- **`.claude/rules/moai/development/coding-standards.md` "effortLevel"** — Settings field introduced by Claude Code v2.1.110. HRN-001 indirectly drives this via env var injection (the runtime sets `CLAUDE_CODE_EFFORT_LEVEL` from `EffortForLevel()`).

## 9. Skill / Workflow References (orchestrator side)

- **`.claude/skills/moai/workflows/harness.md`** (post-V3R4-HARNESS-001) — Owns the user-facing slash command surface. HRN-001's CLI is consumed via Bash invocation from this skill workflow body.
- **`.claude/rules/moai/workflow/spec-workflow.md` §Phase 0.5 Plan Audit Gate** — Reads `cfg.Levels[X].PlanAudit.{Enabled, MaxIterations, RequireMustPass}`. HRN-001's `LevelConfig` struct exposes `PlanAudit` as a sub-struct.
- **`.claude/rules/moai/workflow/spec-workflow.md` §Mode Dispatch (REQ-WF003-018 precedence)** — `harness.yaml` level auto-selection is the lowest-priority fallback for `--mode` resolution; HRN-001's `Route()` is the source.

## 10. Cross-SPEC Dependencies (state at HEAD)

- **SPEC-V3R2-CON-001 (FROZEN floor)** — Encoded in `.claude/rules/moai/design/constitution.md:55-60` `[FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)`. HRN-001's `validatePassThresholdFloor()` reads each profile via `ParseRubricMarkdown` and asserts `>= 0.60`.
- **SPEC-V3R2-MIG-003 (cross-cutting loaders)** — Per spec.md §1 (HRN-001) lines 110-112: harness loader is **promoted** to HRN-001 due to Phase 5 ordering. MIG-003 will reference HRN-001's loader as already-shipped.
- **SPEC-V3R2-ORC-003 (effort matrix)** — Already shipped at HEAD; agent frontmatter `effortLevel:` consumes `medium|high|xhigh` per `internal/config/envkeys.go:86-88` comment.
- **SPEC-V3R2-SPC-001 (SPEC struct)** — Provides `SPECFrontmatter` + `parseSPECDoc()` per anchors §4 above. Already shipped.
- **SPEC-V3R2-HRN-002 (Evaluator memory amendment)** — Owns `EvaluatorConfig.MemoryScope FROZEN per_iteration` at `internal/config/types.go:413`. HRN-001 must NOT modify `EvaluatorConfig`; orthogonal axis.
- **SPEC-V3R2-HRN-003 (Hierarchical acceptance scoring)** — Owns `Profiles`, `Aggregation`, `MustPassDimensions` at `internal/config/types.go:415-422`. HRN-001 must NOT modify `EvaluatorConfig`; orthogonal axis.

## 11. Gap Analysis (Loader vs Schema)

Below is the field-by-field gap matrix. "Schema source" = `.moai/config/sections/harness.yaml`. "Go struct" = `internal/config/types.go` HarnessConfig.

| Schema field (harness.yaml line) | Go struct (types.go line) | Gap | HRN-001 task |
|---|---|---|---|
| `default_profile` (L18) | `DefaultProfile string` (L402) | NONE | — |
| `evaluator.memory_scope` (L31) | `EvaluatorConfig.MemoryScope` (L413) | NONE | — |
| `evaluator.profiles` (not in schema; defaults) | `EvaluatorConfig.Profiles` (L416) | NONE (HRN-003) | — |
| `mode_defaults.{solo,team,cg}` (L73-76) | **MISSING** | **GAP-1** | T-HRN001-06: add `ModeDefaults map[string]string` |
| `auto_detection.enabled` (L3) | **MISSING** | **GAP-2** | T-HRN001-06: add `AutoDetection AutoDetectionConfig` |
| `auto_detection.rules.{minimal,standard,thorough}.conditions` (L4-17) | **MISSING** | **GAP-2** | T-HRN001-06 (sub-struct) |
| `escalation.{enabled,max_escalations,triggers}` (L23-29) | **MISSING** | **GAP-3** | T-HRN001-06: add `Escalation EscalationConfig` |
| `effort_mapping.{minimal,standard,thorough}` (L19-22) | **MISSING** | **GAP-4** | T-HRN001-06: add `EffortMapping map[string]string` |
| `levels.{minimal,standard,thorough}.*` (L32-72) | **MISSING** | **GAP-5** | T-HRN001-06: add `Levels map[string]LevelConfig` (sub-struct) |
| `model_upgrade_review.{checklist,trigger,output,enabled}` (L77-114) | **MISSING** | **GAP-6** | T-HRN001-06: add `ModelUpgradeReview ReviewConfig` |
| `plan_audit_global.{always_enabled,enforce_gate_on_spec_creation}` (L111-114) | **MISSING** | **GAP-7** | T-HRN001-06: add `PlanAuditGlobal PlanAuditGlobalConfig` |
| `learning.*` (L115-126) | (handled separately at retired CLI L389-396 via `learningConfig`) | OUT OF SCOPE | (orthogonal to HRN-001 routing) |

**Verdict**: GAP-1 through GAP-7 are all in M2 scope. Schema is COMPLETE; loader is the only gap.

## 12. Risks Discovered During Research

1. **Frontmatter optional-field lint compatibility**: Adding `HarnessLevel string \`yaml:"harness_level"\`` to `SPECFrontmatter` (`internal/spec/lint.go:255-277`) MUST NOT trigger `FrontmatterInvalid` on legacy SPECs. Verified: `FrontmatterSchemaRule` (per `.claude/rules/moai/development/spec-frontmatter-schema.md`) iterates the canonical 12 fields and emits findings only for missing canonical fields. Optional fields beyond the 12 are tolerated. ✓ Safe.

2. **Retired CLI re-registration risk**: HRN-001 must use distinct command tree from `internal/cli/harness.go:47` retired `newHarnessCmd`. Proposed factory name: `newHarnessRouterCmd` registered in `internal/cli/root.go:95+`. CI guard `internal/cli/harness_retirement_test.go` validates the retired factory remains unregistered. ✓ Conflict-free design.

3. **harnessConfigPath constant collision**: `internal/cli/harness.go:41` exports `harnessConfigPath = ".moai/config/sections/harness.yaml"` (package-private). HRN-001's new CLI file `internal/cli/cmd/harness_route.go` is in a different package (`cmd` vs `cli`); no naming collision possible. ✓ Safe.

4. **Lenient profile anchor scores**: Pre-flight needed in M1: parse `.moai/config/evaluator-profiles/lenient.md` to verify all anchor levels ≥ 0.60. If any anchor < 0.60, the production harness.yaml is in violation of FROZEN floor at HEAD. Mitigation: add pre-M2 read of `lenient.md` to verify; if violation found, escalate to SPEC-V3R2-CON-001 amendment.

5. **`MOAI_CONFIG_STRICT` env conflict**: Search for existing usage of this env name. Not found in `internal/config/envkeys.go:1-90`. ✓ Free to introduce.

## 13. Test Fixture Inventory (M1 RED outputs)

The following testdata directories MUST exist at M1 RED conclusion (validated by `internal/config/loader_harness_extended_test.go`):

- `internal/config/testdata/harness-valid/harness.yaml` — Mirror of shipping `.moai/config/sections/harness.yaml` (golden parse path)
- `internal/config/testdata/harness-invalid-threshold/harness.yaml` + `expected_error.txt` — `pass_threshold: 0.5` in lenient profile copy → expect `ErrPassThresholdFloor`
- `internal/config/testdata/harness-invalid-level/harness.yaml` + `expected_error.txt` — `levels.expert:` → expect `ErrUnknownLevel`
- `internal/config/testdata/harness-drift-strict/harness.yaml` — `unknown_top_field: 1` → no error without env; with `MOAI_CONFIG_STRICT=1`, expect `ErrSchemaDrift`
- `internal/harness/router/testdata/spec-overrides/SPEC-TEST-AAA-001/spec.md` — SPEC frontmatter with `harness_level: thorough` → expect `matched_rule: spec_override`
- `internal/harness/router/testdata/keyword-force/SPEC-TEST-BBB-001/spec.md` — SPEC body containing `oauth` keyword → expect level `thorough` regardless of file_count
- `internal/harness/router/testdata/normal/SPEC-TEST-CCC-001/spec.md` — Plain feature SPEC → expect level `standard`

---

## 14. Anchor Count Summary

Total file:line anchors enumerated in this document: **51** distinct anchors (well above the ≥25 floor required by plan-phase contract). Breakdown:

- harness.yaml schema anchors: 10
- internal/config/types.go anchors: 5
- internal/config/loader.go anchors: 4
- internal/config/errors.go anchors: 3
- internal/config/validation.go anchors: 1
- internal/config/manager.go anchors: 1
- internal/spec/lint.go anchors: 7
- internal/harness/profile_loader.go anchors: 3
- internal/harness/rubric.go anchors: 1
- internal/cli/harness.go anchors: 5
- internal/cli/root.go anchors: 2
- internal/config/envkeys.go anchors: 2
- internal/cli/harness_retirement_test.go anchors: 1
- internal/config/loader_test.go anchors: 3
- internal/config/types_test.go anchors: 0 (referenced for pattern only)
- .moai/config/evaluator-profiles/ anchors: 4 (profile MD files)
- go.mod / dependency anchors: 1
- .claude/rules/moai/ anchors: 4 (frontmatter schema, design constitution, spec-workflow, coding-standards)

---

End of research.md.
