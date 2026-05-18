# Acceptance Criteria — SPEC-V3R2-MIG-003

> Companion to `spec.md` (REQ-MIG003-001 ~ REQ-MIG003-018, 18 REQ).
> Companion to `plan.md` (M1-M4 milestones).
> Companion to `research.md` (38 file:line anchors).
> Companion to `tasks.md` (T-MIG003-NN).

This document enumerates the binary acceptance criteria for SPEC-V3R2-MIG-003. Each AC is independently verifiable via a single command or short script. All ACs MUST pass before `/moai sync` and PR merge.

---

## HISTORY

| Version | Date       | Author       | Description                                                                  |
|---------|------------|--------------|------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial AC synthesis: 15 binary ACs, HRN-001 reconciliation, Given/When/Then |

---

## AC Index (15 ACs, all binary)

| AC ID | REQ coverage | Phase | Status hook |
|---|---|---|---|
| AC-MIG003-01 | REQ-MIG003-001 | M2 | structs exist |
| AC-MIG003-02 | REQ-MIG003-002 (verify-only for harness; new for 4) | M2 | loaders exist |
| AC-MIG003-03 | REQ-MIG003-004, REQ-MIG003-012 | M2 | defaults |
| AC-MIG003-04 | REQ-MIG003-008 (verify-only — HRN-001 reconciled) | M2 | parse error |
| AC-MIG003-05 | REQ-MIG003-005 | M2 | godoc hot path |
| AC-MIG003-06 | REQ-MIG003-006, REQ-MIG003-015 | M2 | DORMANT marker |
| AC-MIG003-07 | REQ-MIG003-007 | M1-M2 | unit tests pass |
| AC-MIG003-08 | REQ-MIG003-009 | M2 | ForbiddenLibraries exposed |
| AC-MIG003-09 | REQ-MIG003-010 | M2 | ContextConfig fields |
| AC-MIG003-10 | REQ-MIG003-011 | M2 | InterviewConfig fields |
| AC-MIG003-11 | REQ-MIG003-014 | M2 | DesignConfig fields |
| AC-MIG003-12 | REQ-MIG003-013 | M3 | YAML_SECTION_NO_LOADER guard |
| AC-MIG003-13 | REQ-MIG003-017 | M4 | LOADER_HARDCODE_VIOLATION review |
| AC-MIG003-14 | REQ-MIG003-016 | M3 | CONFIG_STRUCT_YAML_MISMATCH guard |
| AC-MIG003-15 | REQ-MIG003-018 | M3 | SUNSET_CONFIG_DORMANT_NOTICE once |

---

## AC-MIG003-01 — 4 new structs exist in types.go

**REQ**: REQ-MIG003-001
**Phase**: M2 GREEN

**Given** the internal/config package has been built
**When** `grep -nE '^type (ConstitutionConfig|ContextConfig|InterviewConfig|DesignConfig) struct' internal/config/types.go` runs
**Then** exactly 4 matching lines are output, one per struct.

Additionally `grep -nE '^type HarnessConfig struct' internal/config/types.go` returns exactly 1 match at line ~402 (verify-only — HRN-001 delivered).

**Verify command**:
```bash
test "$(grep -cE '^type (ConstitutionConfig|ContextConfig|InterviewConfig|DesignConfig) struct' internal/config/types.go)" = "4"
```

Pass ⇔ exit 0.

---

## AC-MIG003-02 — 4 new loader functions exist + HarnessConfig loader verified

**REQ**: REQ-MIG003-002 (4 new ones in MIG-003; LoadHarnessConfig verified from HRN-001)
**Phase**: M2 GREEN

**Given** the internal/config package compiles
**When** `grep -nE '^func Load(Constitution|Context|Interview|Design)Config' internal/config/` is run across all *.go files
**Then** exactly 4 matching function declarations exist (one per section).

**Verify command**:
```bash
test "$(grep -rE '^func Load(Constitution|Context|Interview|Design)Config' internal/config/ --include='*.go' | wc -l | tr -d ' ')" = "4"
# Also verify HRN-001's LoadHarnessConfig still exists (regression check)
grep -n '^func LoadHarnessConfig' internal/config/loader.go
```

Pass ⇔ both commands exit 0 AND first grep returns 4 lines.

---

## AC-MIG003-03 — Defaults returned on absent file

**REQ**: REQ-MIG003-004, REQ-MIG003-012 (verify-only for harness)
**Phase**: M2 GREEN

**Given** a config directory that does NOT contain `constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`
**When** `Loader.Load(dir)` is invoked
**Then**:
- No error is returned
- `cfg.Constitution`, `cfg.ContextSearch`, `cfg.Interview`, `cfg.Design` are populated with non-zero default values (matching `internal/template/templates/.moai/config/sections/*.yaml`)
- `loadedSections["constitution"]`, `loadedSections["context_search"]`, `loadedSections["interview"]`, `loadedSections["design"]` are NOT set to `true` (graceful — file absent)

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoad(Constitution|Context|Interview|Design)Config_MissingFile_ReturnsDefaults' -v
```

Pass ⇔ all 4 tests exit 0.

---

## AC-MIG003-04 — Malformed YAML returns typed error

**REQ**: REQ-MIG003-008 (HRN-001 reconciliation: `ErrInvalidYAML` subsumes the `*_PARSE_ERROR` naming)
**Phase**: M2 GREEN

**Given** a `constitution.yaml` / `context.yaml` / `interview.yaml` / `design.yaml` with malformed YAML syntax
**When** the corresponding `LoadXxxConfig(path)` is invoked
**Then** the returned error satisfies `errors.Is(err, ErrInvalidYAML)` AND the error message contains the file path.

Verify-only for harness: `LoadHarnessConfig` malformed-path already returns `ErrInvalidYAML` (HRN-001 test `TestLoadHarnessConfigExtended_InvalidThreshold` covers analogous case at loader_harness_extended_test.go:86).

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoad(Constitution|Context|Interview|Design)Config_Malformed' -v
```

Pass ⇔ all 4 tests exit 0.

---

## AC-MIG003-05 — Each struct godoc names runtime hot path

**REQ**: REQ-MIG003-005
**Phase**: M2 GREEN

**Given** each new struct in `internal/config/types.go`
**When** the godoc comment block for `ConstitutionConfig`, `ContextConfig`, `InterviewConfig`, `DesignConfig` is inspected
**Then** the godoc explicitly names at least one runtime hot-path consumer:
- `ConstitutionConfig` → mentions "SPEC-V3R2-EXT-004" OR "forbidden-library policy"
- `ContextConfig` → mentions "CLAUDE.md §16 Context Search" OR "token_budget"
- `InterviewConfig` → mentions "SPEC-V3R2-WF-003 discovery" OR "clarity-threshold"
- `DesignConfig` → mentions "GAN loop" OR "sprint contract"

**Verify command**:
```bash
go doc -all ./internal/config | grep -E '^type (ConstitutionConfig|ContextConfig|InterviewConfig|DesignConfig) struct' -A 5 | grep -E 'EXT-004|CLAUDE.md §16|WF-003|GAN loop'
test "$(go doc -all ./internal/config | grep -E '^type (ConstitutionConfig|ContextConfig|InterviewConfig|DesignConfig) struct' -A 5 | grep -cE 'EXT-004|CLAUDE.md §16|WF-003|GAN loop')" -ge "4"
```

Pass ⇔ at least 4 hot-path references in the 4 struct godocs.

---

## AC-MIG003-06 — SunsetConfig DORMANT godoc marker

**REQ**: REQ-MIG003-006, REQ-MIG003-015
**Phase**: M2 GREEN

**Given** the `SunsetConfig` struct at `internal/config/types.go:309`
**When** its godoc comment block is inspected
**Then** the comment contains the literal string `DORMANT` AND a forward-reference note ("Activation deferred to a future SPEC. Do NOT add LoadSunsetConfig until activation SPEC is filed.").

**Verify command**:
```bash
grep -B 0 -A 8 'type SunsetConfig struct' internal/config/types.go | grep -q 'DORMANT'
grep -B 0 -A 8 'type SunsetConfig struct' internal/config/types.go | grep -q 'Activation deferred'
```

Pass ⇔ both grep exit 0.

---

## AC-MIG003-07 — Unit tests pass for all 4 loaders

**REQ**: REQ-MIG003-007
**Phase**: M1 (RED) + M2 (GREEN)

**Given** the 4 new loader test files exist and the M2 implementation is complete
**When** `go test ./internal/config/...` runs
**Then** all 4 loader test suites (constitution, context, interview, design) PASS with at minimum 3 cases each: valid file load + missing file default + malformed file error.

Additionally HRN-001's harness tests must continue to pass (regression guard).

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoad(Constitution|Context|Interview|Design|Harness)Config' -v -count=1
```

Pass ⇔ exit 0 AND zero FAIL/SKIP for the 5 prefix matches.

---

## AC-MIG003-08 — ConstitutionConfig.ForbiddenLibraries is exposed

**REQ**: REQ-MIG003-009
**Phase**: M2 GREEN

**Given** a valid `constitution.yaml` with `forbidden_patterns: ["global mutable state", "init() with side effects", ...]`
**When** `LoadConstitutionConfig(path)` returns and the result is inspected
**Then** the typed result exposes the list (via `cfg.ForbiddenLibraries` or `cfg.ForbiddenPatterns` accessor — name finalized in M2) with at least 4 entries matching the YAML, AND the field is non-private (exported, accessible from outside the package).

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoadConstitutionConfig_ForbiddenLibrariesExposed' -v
```

Pass ⇔ exit 0.

---

## AC-MIG003-09 — ContextConfig token_budget and date_range_days parsed

**REQ**: REQ-MIG003-010
**Phase**: M2 GREEN

**Given** a valid `context.yaml` with `token_budget.max_injection_tokens: 5000`, `token_budget.skip_if_usage_above: 150000`, `search.date_range_days: 30`
**When** `LoadContextConfig(path)` returns
**Then** the returned struct exposes:
- `cfg.TokenBudget.MaxInjectionTokens == 5000`
- `cfg.TokenBudget.SkipIfUsageAbove == 150000`
- `cfg.Search.DateRangeDays == 30`

(REQ-MIG003-010 mentions "staleness_window_days" but the actual YAML field is `search.date_range_days`; the struct uses the YAML name. Plan.md OQ4 documents this drift.)

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoadContextConfig_ValidParse' -v
```

Pass ⇔ exit 0.

---

## AC-MIG003-10 — InterviewConfig clarity_threshold and max_rounds parsed

**REQ**: REQ-MIG003-011
**Phase**: M2 GREEN

**Given** a valid `interview.yaml` with `clarity_threshold: 4`, `plan.max_rounds: 5`, `plan.questions_per_round: 3`, `skip_conditions: [...]`
**When** `LoadInterviewConfig(path)` returns
**Then** the returned struct exposes:
- `cfg.ClarityThreshold == 4`
- `cfg.Plan.MaxRounds == 5`
- `cfg.Plan.QuestionsPerRound == 3`
- `len(cfg.SkipConditions) == 3`

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoadInterviewConfig_ValidParse' -v
```

Pass ⇔ exit 0.

---

## AC-MIG003-11 — DesignConfig sprint_contract and gan_loop parsed

**REQ**: REQ-MIG003-014
**Phase**: M2 GREEN

**Given** a valid `design.yaml` with `gan_loop.pass_threshold: 0.75`, `gan_loop.max_iterations: 5`, `gan_loop.sprint_contract.enabled: true`, `adaptation.iteration_limits.builder: 3`
**When** `LoadDesignConfig(path)` returns
**Then** the returned struct exposes:
- `cfg.GanLoop.PassThreshold == 0.75`
- `cfg.GanLoop.MaxIterations == 5`
- `cfg.GanLoop.SprintContract.Enabled == true`
- `cfg.Adaptation.IterationLimits["builder"] == 3` (or struct field equivalent)

(Plan.md OQ4 documents that REQ-MIG003-014's `phase_weights` is conceptually subsumed by the actual YAML `iteration_limits`.)

**Verify command**:
```bash
go test ./internal/config/... -run 'TestLoadDesignConfig_ValidParse' -v
```

Pass ⇔ exit 0.

Bonus: `TestLoadDesignConfig_PassThresholdFloorViolation` PASSes (pass_threshold < 0.60 rejected with `ErrPassThresholdFloor`).

---

## AC-MIG003-12 — YAML_SECTION_NO_LOADER CI guard active

**REQ**: REQ-MIG003-013
**Phase**: M3 (CI Guards)

**Given** the audit test `internal/config/audit_loader_completeness_test.go` exists
**When** `go test ./internal/config/... -run TestAuditLoaderCompleteness` runs
**Then**:
- Current state: PASS (4 new sections wired + DORMANT sunset + acknowledged-out-of-scope allowlist)
- Simulated state: if a new YAML file `internal/template/templates/.moai/config/sections/__test_new.yaml` is temporarily added with no corresponding loader and not in the allowlist, the test FAILs with `YAML_SECTION_NO_LOADER: __test_new`

**Verify command**:
```bash
# Positive case
go test ./internal/config/... -run TestAuditLoaderCompleteness -v
# Negative case (manual or scripted): drop a sentinel yaml, expect FAIL
```

Pass ⇔ positive-case exit 0 AND (during run-phase verification) negative-case demonstrates FAIL with the expected message.

---

## AC-MIG003-13 — LOADER_HARDCODE_VIOLATION review checklist

**REQ**: REQ-MIG003-017
**Phase**: M4 REFACTOR

**Given** the 4 new loader files (`loader_constitution.go`, `loader_context.go`, `loader_interview.go`, `loader_design.go`)
**When** code review (or automated grep audit) inspects them
**Then** zero inline literal values are found that override YAML field values. Specifically, no `const maxRounds = 4` style hardcodes overriding `interview.clarity_threshold`. Defaults are sourced exclusively from `internal/config/defaults.go` helpers.

**Verify command**:
```bash
# Grep audit: any numeric/string literal in loader_*.go that matches a YAML value field
# (heuristic — manual review supplements)
! grep -nE '^\s*(const|var)\s+\w+\s*=\s*[0-9]+\s*$' internal/config/loader_constitution.go internal/config/loader_context.go internal/config/loader_interview.go internal/config/loader_design.go
```

Pass ⇔ no numeric `const` / `var` declarations at top-level in the 4 loader files. Note: graceful-defaults sourced from `defaults.go` are NOT in scope of this AC (they live in defaults.go, intentionally).

---

## AC-MIG003-14 — CONFIG_STRUCT_YAML_MISMATCH `make build` guard

**REQ**: REQ-MIG003-016
**Phase**: M3 (CI Guards)

**Given** the audit test `internal/config/audit_struct_yaml_symmetry_test.go` exists
**When** `make build` runs (which executes `go test ./internal/config/...` as part of the build target)
**Then**:
- Current state: PASS — for each of `ConstitutionConfig`, `ContextConfig`, `InterviewConfig`, `DesignConfig`, every yaml-tagged Go field has a corresponding key in the template YAML, and every YAML key has a yaml-tagged Go field.
- Simulated state: if a Go field with `yaml:"new_field"` is added without a matching YAML key, test FAILs with `CONFIG_STRUCT_YAML_MISMATCH: field=ConstitutionConfig.NewField, side=go-only`

**Verify command**:
```bash
# Build-time invocation
make build
# Direct invocation
go test ./internal/config/... -run TestStructYAMLSymmetry -v
```

Pass ⇔ both commands exit 0.

---

## AC-MIG003-15 — SUNSET_CONFIG_DORMANT_NOTICE logged once per session

**REQ**: REQ-MIG003-018
**Phase**: M3 (Once-per-Session Notice)

**Given** the `Loader.Load()` function is called from a process where `.moai/config/sections/sunset.yaml` exists
**When** `Loader.Load()` is invoked twice (or more) within the same process lifetime
**Then** exactly ONE `SUNSET_CONFIG_DORMANT_NOTICE` slog record is emitted, with attributes `spec="SPEC-V3R2-MIG-003 REQ-018"` and `yaml_path` populated.

Additionally: if `sunset.yaml` does NOT exist in the sections directory, zero notices are emitted.

**Verify command**:
```bash
go test ./internal/config/... -run 'TestSunsetNotice_FiresOnce|TestSunsetNotice_AbsentFileNoNotice' -v
```

Pass ⇔ both tests exit 0.

---

## Definition of Done (DoD)

All of the following must hold before this SPEC is considered Done and a sync PR is created:

- [ ] AC-MIG003-01 through AC-MIG003-15: all PASS (exit code 0 from all verify commands)
- [ ] `go test ./...`: 100% green across the entire repo (no regressions outside internal/config)
- [ ] `golangci-lint run ./...`: zero new findings
- [ ] `make build`: success (includes CI guard AC-MIG003-14 verification)
- [ ] Test coverage for `internal/config/` (incremental): ≥85% line coverage; new files ≥90%
- [ ] `.claude/rules/moai/core/settings-management.md` updated to document 4 new loaders + DORMANT sunset + 2 CI guards (M4 task)
- [ ] `progress.md` records `plan_status: audit-ready` (already set in this plan-phase artifact frontmatter)
- [ ] Plan-auditor verdict: PASS (run before /moai run handoff)
- [ ] No new `direct` dependencies added to `go.mod` (REQ-MIG003 §7 9-direct-dep policy)
- [ ] All commits follow Conventional Commits format
- [ ] HRN-001 regression: `TestLoadHarnessConfigExtended_*` all PASS (verify-only baseline)

## Quality Gates

| Gate | Threshold | Source |
|---|---|---|
| Line coverage (internal/config) | ≥85% | quality.yaml |
| New file coverage | ≥90% | plan.md §8.2 |
| Lint findings | 0 | quality.yaml |
| TRUST 5 Tested | All tests green | `.claude/rules/moai/core/moai-constitution.md` |
| TRUST 5 Readable | golangci-lint pass | same |
| TRUST 5 Unified | gofmt + goimports pass | same |
| TRUST 5 Secured | No hardcoded credentials, no panic in library code | same |
| TRUST 5 Trackable | Conventional commits | same |
| Plan-Auditor | PASS verdict before /moai run | `.claude/rules/moai/workflow/spec-workflow.md` §Phase 0.5 |

## Note on Verify-Only ACs

The following ACs are partially **verify-only** because SPEC-V3R2-HRN-001 already delivered the harness slice:

- **AC-MIG003-02**: harness loader portion verified (HRN-001 work product); 4 new loaders are net-new
- **AC-MIG003-03**: harness defaults portion is dedicated (`LoadHarnessConfig` returns `ErrConfigNotFound`, NOT defaults — by HRN-001 design); the 4 new loaders ARE graceful-default
- **AC-MIG003-04**: harness parse error returns `ErrInvalidYAML` (HRN-001 sentinel); the 4 new loaders mirror

This is the explicit HRN-001 reconciliation point (plan.md §1, research.md §2). The spec.md was authored before HRN-001 shipped; this acceptance.md reflects the post-HRN-001 reality.

End of acceptance.md
