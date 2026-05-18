# SPEC-V3R2-HRN-001 — Implementation Plan (plan.md)

> Companion to `spec.md` (REQ-HRN-001-001 ~ REQ-HRN-001-019, 19 REQ).
> Companion to `research.md` (codebase anchors, gap analysis).
> Companion to `acceptance.md` (10 binary AC) and `tasks.md` (T-HRN001-NN decomposition).

---

## HISTORY

| Version | Date       | Author | Description |
|---------|------------|--------|-------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial plan-phase synthesis: M1-M5 milestones, REQ↔Task matrix, risks, file map |

---

## 1. Scope (1-paragraph synthesis)

Implement the Go-side runtime for `.moai/config/sections/harness.yaml`: a fully-typed `HarnessConfig` struct, a fail-closed `LoadHarnessConfig()` loader with `validator/v10` invariants (FROZEN `pass_threshold >= 0.60`, FROZEN level enum), a `HarnessRouter.Route()` function that maps a SPEC to `minimal|standard|thorough` via Complexity Estimator signals, an `EscalationManager` with `max_escalations` cap, an `EffortForLevel()` helper that bridges to SPEC-V3R2-ORC-003's effort matrix, and a `moai harness route|validate` CLI surface. No schema changes to harness.yaml; no content authoring of evaluator-profile `.md` files; no Sprint Contract negotiation (deferred to HRN-002/003).

## 2. Approach (HOW — but at architecture-level only)

The current `HarnessConfig` struct in `internal/config/types.go:401-404` is intentionally minimal (only `DefaultProfile` + `Evaluator`) per HRN-002 substrate. HRN-001 run-phase **extends** this struct (not replaces) with five new top-level fields covering the full harness.yaml schema. Loader `LoadHarnessConfig()` at `internal/config/loader.go:227-262` is extended (not rewritten) to add `validator/v10` validation, `MOAI_CONFIG_STRICT` env-driven drift detection, and pass-threshold floor enforcement. New routing logic lives in a fresh package `internal/harness/router/` (sibling to `internal/harness/` to avoid circular deps with `EvaluatorConfig`). The CLI surface uses cobra's command-tree, registered into `internal/cli/root.go` post-deprecation-cleanup (the legacy retired `newHarnessCmd` in `internal/cli/harness.go:47-71` is **NOT** repurposed; HRN-001 introduces a separate `newHarnessRouteCmd` factory under a new file `internal/cli/cmd/harness_route.go` to avoid collision with the SPEC-V3R4-HARNESS-001 CI guard at `internal/cli/harness_retirement_test.go`).

Complexity Estimator (REQ-HRN-001-007/008) consumes the existing `spec.SPECFrontmatter` struct (`internal/spec/lint.go:255-277`) for `priority`, `tags`, and `module`; combined with body-side regex extraction for `Requirements` section to find security/payment keyword matches. `file_count` is derived from acceptance.md + plan.md path mentions (regex match `internal/...`, `.moai/...` patterns); `domain_count` is derived from `tags` field comma-split + `phase` token. Fallback rules (REQ-HRN-001-007 priority order: minimal → standard → thorough) and force-thorough overrides (REQ-HRN-001-008) are explicit in `router.Route()`.

## 3. Milestones (Priority-based, NO time estimates)

[HARD] Priority labels only — no time estimates per agent-common-protocol.

### M1 — RED Phase (Priority: P0 Critical, blocking)

**Test scaffolding before any implementation code.**

- Write test fixtures in `internal/config/testdata/harness-valid/`, `harness-invalid-threshold/`, `harness-invalid-level/`, `harness-drift-strict/`.
- Write `internal/config/loader_harness_extended_test.go`: parse fixtures, assert struct fields populated correctly, assert validator errors are typed sentinels.
- Write `internal/harness/router/router_test.go`: table-driven AC fixtures for REQ-007 (priority order), REQ-008 (force thorough), REQ-015 (spec_override).
- Write `internal/harness/router/escalation_test.go`: cap test (REQ-013), trigger test (REQ-009), cap-reached log test (REQ-018).
- Write `internal/harness/router/effort_test.go`: 3-row table (minimal→medium, standard→high, thorough→xhigh).
- Write `internal/cli/cmd/harness_route_test.go`: CLI golden tests with `--json` output schema validation (REQ-011, REQ-006).
- All tests FAIL at this milestone (no implementation yet). Verified via `go test ./internal/harness/router/... ./internal/cli/cmd/...` produces "undefined" errors.

### M2 — GREEN Phase Part 1: Struct + Loader (Priority: P0 Critical)

**Extend `HarnessConfig` struct + `LoadHarnessConfig()` to populate full schema.**

- Modify `internal/config/types.go:401-423`: add `ModeDefaults`, `AutoDetection`, `Escalation`, `EffortMapping`, `Levels`, `ModelUpgradeReview` fields to `HarnessConfig`. Each field gets its own sub-struct (e.g., `AutoDetectionConfig`, `LevelConfig`).
- Modify `internal/config/loader.go:227-262`: extend `LoadHarnessConfig()` to validate new fields via `validator/v10` (reuse existing `validate = validator.New()` from `internal/config/validation.go:15`). Add `MOAI_CONFIG_STRICT` env probe (REQ-019).
- Add sentinel errors to `internal/config/errors.go`: `ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded`.
- M1 loader tests transition RED → GREEN.

### M3 — GREEN Phase Part 2: Routing Logic (Priority: P0 Critical, depends on M2)

**Complexity Estimator + Router.Route() + force-thorough override.**

- New file `internal/harness/router/router.go`: define `Level` string type with constants `LevelMinimal|LevelStandard|LevelThorough`, define `Rationale` struct (matched_rule, file_count, domain_count, spec_type, spec_priority, keywords), define `Router` interface + `defaultRouter` impl.
- New file `internal/harness/router/complexity.go`: signal extraction from `spec.SPECFrontmatter` + body markdown (file_count from path mentions, domain_count from tags comma-split, spec_type from priority+phase heuristic, keyword matching against canonical sets).
- New file `internal/harness/router/keywords.go`: constants `securityKeywords = []string{"auth", "crypto", "encrypt", "oauth", "jwt", "session", "password", "rbac", "acl"}` and `paymentKeywords = []string{"payment", "billing", "subscription", "invoice", "charge", "stripe", "paypal"}` (REQ-008).
- Wire force-thorough override (REQ-008): security/payment keywords OR `spec_priority == critical` OR domain in [auth, payment, migration, public_api] forces `LevelThorough`.
- Wire spec_override (REQ-015): if SPEC frontmatter has `harness_level:` field (NEW optional field in frontmatter — add to `spec.SPECFrontmatter` at `internal/spec/lint.go:268`), honor and set `matched_rule: spec_override`.
- M1 router tests transition RED → GREEN.

### M4 — GREEN Phase Part 3: Escalation + Effort + CLI (Priority: P0 Critical, depends on M3)

**EscalationManager + EffortForLevel + CLI subcommand.**

- New file `internal/harness/router/escalation.go`: `EscalationManager` struct with `MaxEscalations` field bound from config, `CheckTriggers(ctx)` method enforcing cap (REQ-013, REQ-018). Bump logic minimal → standard → thorough; cap at thorough.
- New file `internal/harness/router/effort.go`: `EffortForLevel(level, cfg)` reads `cfg.EffortMapping[string(level)]`; fallback to `cfg.EffortMapping` defaults `{minimal: medium, standard: high, thorough: xhigh}`.
- New file `internal/cli/cmd/harness_route.go`: cobra factory `newHarnessRouteCmd()` with `--spec`, `--json`, `--path` flags. Wires to `Route()`, prints rationale text or JSON document per REQ-011.
- New file `internal/cli/cmd/harness_validate.go`: cobra factory `newHarnessValidateCmd()` invokes `LoadHarnessConfig()` + reports validation errors via exit code 1 (REQ-010, FROZEN floor REQ-012).
- Register the parent `harness` cobra command in `internal/cli/root.go` (post-line 92, separate group from retired CLI). [CRITICAL] The retired `newHarnessCmd` at `internal/cli/harness.go:47` MUST remain unregistered; HRN-001's command is the NEW non-conflicting `newHarnessRouterCmd` family.
- M1 CLI tests transition RED → GREEN.
- Evaluator-profile loader integration: read `cfg.Levels[level].EvaluatorProfile` (e.g., "strict"); resolve via `internal/harness/profile_loader.go:14-19` `defaultProfilePaths`. Validate `PassThreshold >= 0.60` floor by parsing each referenced profile's `.md` via `ParseRubricMarkdown()` (`internal/harness/rubric.go`).

### M5 — REFACTOR Phase + Final Verification (Priority: P1 High, depends on M4)

**MX tags, CHANGELOG, lint, doc cross-references.**

- Add `@MX:NOTE` tags to all new exported symbols (`HarnessConfig` extension fields, `Router`, `EscalationManager`, `EffortForLevel`). MX language per `code_comments: ko` from `.moai/config/sections/language.yaml`.
- Add `@MX:ANCHOR` to high-fan_in entry points: `LoadHarnessConfig`, `Router.Route`, `EffortForLevel`, `EscalationManager.CheckTriggers`. Each requires `@MX:REASON` per protocol.
- Add `@MX:WARN` for FROZEN invariant guard sites: `validateLevelEnum`, `validatePassThresholdFloor` with `@MX:REASON: FROZEN per design-constitution §5 (SPEC-V3R2-CON-001)`.
- Update `CHANGELOG.md`: entry under `## [Unreleased] — 2026-05-XX` referencing `SPEC-V3R2-HRN-001`.
- Run `golangci-lint run ./internal/config/... ./internal/harness/... ./internal/cli/cmd/...` (zero ERROR per `make preflight`).
- Run `go test -race -count=1 ./internal/config/... ./internal/harness/router/... ./internal/cli/cmd/...` (all PASS).
- Coverage gate: `go test -cover` returns >= 85% per harness section coverage budget.
- Re-run all 10 AC via `acceptance.md` verification commands; binary 0 findings expected.

---

## 4. Technical Approach (Architecture)

### 4.1 Module Layout (NEW vs MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `internal/config/types.go:401-423` | MODIFIED | Extend `HarnessConfig` with 5 new fields; add `AutoDetectionConfig`, `LevelConfig`, `EscalationConfig`, `ModelUpgradeReviewConfig` sub-structs |
| `internal/config/loader.go:223-262` | MODIFIED | Extend `LoadHarnessConfig()`: validator/v10 invariants, MOAI_CONFIG_STRICT drift detection, profile floor validation |
| `internal/config/errors.go:14-68` | MODIFIED | Add `ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded` |
| `internal/spec/lint.go:255-277` | MODIFIED | Add optional `HarnessLevel string \`yaml:"harness_level"\`` field for REQ-015 spec_override |
| `internal/harness/router/router.go` | NEW | `Level`, `Rationale`, `Router` interface + `defaultRouter` impl |
| `internal/harness/router/complexity.go` | NEW | Signal extraction (file_count, domain_count, spec_type, keywords) |
| `internal/harness/router/keywords.go` | NEW | `securityKeywords`, `paymentKeywords` canonical sets (REQ-008) |
| `internal/harness/router/escalation.go` | NEW | `EscalationManager` with `MaxEscalations` cap |
| `internal/harness/router/effort.go` | NEW | `EffortForLevel(level, cfg) string` |
| `internal/harness/router/router_test.go` | NEW | REQ-007/008/015 table-driven tests |
| `internal/harness/router/escalation_test.go` | NEW | REQ-009/013/018 tests |
| `internal/harness/router/effort_test.go` | NEW | REQ-005 effort mapping tests |
| `internal/cli/cmd/harness_route.go` | NEW | `moai harness route --spec SPEC-XXX [--json]` |
| `internal/cli/cmd/harness_validate.go` | NEW | `moai harness validate [--path PATH]` |
| `internal/cli/cmd/harness_route_test.go` | NEW | CLI golden tests + JSON schema validation |
| `internal/cli/root.go` | MODIFIED | Register `newHarnessRouterCmd` (NOT the retired `newHarnessCmd`) |
| `internal/config/testdata/harness-valid/harness.yaml` | NEW | Loader fixture: complete valid schema |
| `internal/config/testdata/harness-invalid-threshold/` | NEW | Profile with `pass_threshold: 0.5` (FROZEN floor violation) |
| `internal/config/testdata/harness-invalid-level/` | NEW | Schema with `levels.expert:` (unknown level) |
| `internal/config/testdata/harness-drift-strict/` | NEW | Schema with `unknown_top_field: 1` (drift; only fails with `MOAI_CONFIG_STRICT=1`) |

### 4.2 Integration Points

- **SPEC-V3R2-ORC-003 effort matrix**: `EffortForLevel()` output (`medium|high|xhigh`) MUST align with agent frontmatter `effortLevel:` values consumed by ORC-003. Verified via AC-HRN-001-006.
- **SPEC-V3R2-RT-005 ConfigManager.Reload**: `LoadHarnessConfig()` is called by `ConfigManager.Reload()` (`internal/config/manager.go:200`) at runtime to refresh harness config without restart. Verified via AC-HRN-001-007.
- **SPEC-V3R2-CON-001 FROZEN floor**: `validatePassThresholdFloor()` in `internal/harness/router/router.go` reads each `cfg.Levels[X].EvaluatorProfile` reference, resolves to `.moai/config/evaluator-profiles/{name}.md` path, parses via `harness.ParseRubricMarkdown()`, and rejects any `PassThreshold < 0.60`. Verified via AC-HRN-001-005.
- **SPEC-V3R2-SPC-001 SPEC struct**: `spec.SPECFrontmatter` provides `Priority`, `Tags`, `Module` fields needed by Complexity Estimator. `parseSPECDoc()` (`internal/spec/lint.go:303`) is the parser used by `router.fromSPECID()`.
- **SPEC-V3R2-MIG-003 cross-cutting loader**: HRN-001 owns `LoadHarnessConfig()`; MIG-003 lists harness loader as already-shipped in its scope when MIG-003 enters plan-phase.

### 4.3 Configuration Precedence (REQ-014 mode_defaults)

1. SPEC `harness_level:` frontmatter override (REQ-015) — HIGHEST
2. Force-thorough overrides (REQ-008) — security/payment keywords / critical / sensitive domain
3. Escalation bumps (REQ-009) — cumulative
4. Auto-detection rules (REQ-007) — minimal → standard → thorough priority order
5. Mode defaults (REQ-014) — `cfg.ModeDefaults.cg = "thorough"` (FROZEN), solo/team `auto` (uses auto-detection)

Force-thorough (REQ-008) wins over spec_override only when explicitly documented (per spec.md §1 line 270 risk table); current implementation: spec_override > force-thorough (less surprising). [DECISION: per design-constitution §3 brand context primacy analog, spec_override wins; security/payment keyword force-thorough is a *suggestion* via rationale field annotation. AC-HRN-001-09 verifies.] (Open question for run-phase; document in `progress.md`.)

---

## 5. REQ ↔ Task ↔ AC Traceability Matrix

| REQ | Description (short) | Milestone | Task(s) | AC |
|-----|---------------------|-----------|---------|-----|
| REQ-HRN-001-001 | HarnessConfig struct types | M2 | T-HRN001-06 | AC-HRN-001-01 |
| REQ-HRN-001-002 | LoadHarnessConfig() function | M2 | T-HRN001-07, T-HRN001-08 | AC-HRN-001-01, AC-HRN-001-04 |
| REQ-HRN-001-003 | Router.Route() function | M3 | T-HRN001-10, T-HRN001-11 | AC-HRN-001-02, AC-HRN-001-03 |
| REQ-HRN-001-004 | EscalationManager.CheckTriggers() | M4 | T-HRN001-14 | AC-HRN-001-08 |
| REQ-HRN-001-005 | EffortForLevel() | M4 | T-HRN001-15 | AC-HRN-001-10 |
| REQ-HRN-001-006 | moai harness route/validate CLI | M4 | T-HRN001-16, T-HRN001-17 | AC-HRN-001-02, AC-HRN-001-04, AC-HRN-001-06 |
| REQ-HRN-001-007 | Auto-detection priority order | M3 | T-HRN001-12 | AC-HRN-001-02, AC-HRN-001-03 |
| REQ-HRN-001-008 | Force-thorough on security/payment keywords | M3 | T-HRN001-13 | AC-HRN-001-03 |
| REQ-HRN-001-009 | Escalation triggers fire on phase failure | M4 | T-HRN001-14 | AC-HRN-001-08 |
| REQ-HRN-001-010 | Validation error wrapping | M2 | T-HRN001-08 | AC-HRN-001-05 |
| REQ-HRN-001-011 | JSON output for CLI | M4 | T-HRN001-17 | AC-HRN-001-06 |
| REQ-HRN-001-012 | FROZEN pass_threshold floor | M2 | T-HRN001-09 | AC-HRN-001-05 |
| REQ-HRN-001-013 | max_escalations cap (default 2, max 3) | M4 | T-HRN001-14 | AC-HRN-001-08 |
| REQ-HRN-001-014 | mode_defaults consultation | M3 | T-HRN001-12 | (regression: cg=thorough) |
| REQ-HRN-001-015 | SPEC harness_level: override | M3 | T-HRN001-13 | AC-HRN-001-09 |
| REQ-HRN-001-016 | Model upgrade review reminder | M4 | T-HRN001-18 | (regression: model-change reminder) |
| REQ-HRN-001-017 | Unknown level rejection | M2 | T-HRN001-08 | (regression: ErrUnknownLevel) |
| REQ-HRN-001-018 | Escalation cap log on overflow | M4 | T-HRN001-14 | AC-HRN-001-08 |
| REQ-HRN-001-019 | MOAI_CONFIG_STRICT schema drift detection | M2 | T-HRN001-08 | AC-HRN-001-07 |

---

## 6. Risks & Mitigations (synthesis with spec.md §8)

| Risk | Severity | Mitigation in plan |
|------|----------|--------------------|
| Keyword matcher false positives on docs prose | MEDIUM | Restrict matching to SPEC frontmatter `tags` + EARS Requirements section parsed via `parseREQs()` (`internal/spec/lint.go:357`); skip narrative prose |
| file_count / domain_count estimator unreliable for legacy SPECs lacking acceptance.md | MEDIUM | Fallback: file_count = 0 → minimal candidate; absence of acceptance.md is normal for legacy. `harness_level:` frontmatter override gives author opt-out |
| FROZEN floor (0.60) enforcement breaks lenient profile (`.moai/config/evaluator-profiles/lenient.md`) | HIGH | Verify `lenient.md` per-anchor scores ≥ 0.60 BEFORE M2 RED→GREEN transition; add fixture covering lenient profile to AC-HRN-001-05 |
| Collision with retired `newHarnessCmd` at `internal/cli/harness.go:47` | HIGH | Use distinct factory name `newHarnessRouterCmd` + register only the NEW factory in `root.go`. CI guard `internal/cli/harness_retirement_test.go` MUST continue to pass (the retired factory remains unregistered) |
| Adding `harness_level:` field to `SPECFrontmatter` triggers `FrontmatterInvalid` lint findings on existing SPECs | MEDIUM | New field is **optional** (no `validate:"required"` tag); lint rule `FrontmatterSchemaRule` ignores optional fields per `internal/spec/lint.go` §SSOT |
| `ConfigManager.Reload()` integration race with concurrent SPEC loads | LOW | `LoadHarnessConfig()` is read-only on disk; `ConfigManager` already uses `sync.RWMutex`. No new lock needed |
| Schema drift warning floods stderr when running on legacy projects | LOW | REQ-019 explicit: warning is non-blocking unless `MOAI_CONFIG_STRICT=1`; mirror existing `slog.Warn` pattern from `loader.go:42` |

---

## 7. Quality Gates (M5 verification)

- [ ] `go test -race ./internal/config/... ./internal/harness/router/... ./internal/cli/cmd/...` → 0 failures
- [ ] `go test -cover ./internal/harness/router/...` → ≥ 85% per spec.md §3 coverage target
- [ ] `golangci-lint run ./internal/config/... ./internal/harness/router/...` → 0 ERROR
- [ ] `moai spec lint --strict` → 0 ERROR (frontmatter `harness_level:` optional field accepted)
- [ ] `moai harness validate` on shipping `.moai/config/sections/harness.yaml` → exit 0
- [ ] CI guard `internal/cli/harness_retirement_test.go` → still PASS (retired factory unregistered)
- [ ] CHANGELOG.md updated with SPEC-V3R2-HRN-001 entry
- [ ] MX tags added to all new exported symbols per `mx-tag-protocol.md`

---

## 8. Out-of-Scope (re-affirmation, mirrors spec.md §1.2)

- Schema authoring of evaluator-profile `.md` files (lenient/strict/frontend) — content unchanged
- Sprint Contract negotiation (HRN-002 + HRN-003)
- Hierarchical acceptance scoring (HRN-003)
- harness.yaml schema CHANGE (HRN-001 implements loader only; schema FROZEN for v3.0)
- evaluator-active agent body modification (agent reads runtime config; no body change)
- Telemetry export (beta.1 follow-up)
- Cross-SPEC harness coordination (per-SPEC scope)
- ML-based complexity estimation (rule-based only for v3.0)

---

End of plan.md.
