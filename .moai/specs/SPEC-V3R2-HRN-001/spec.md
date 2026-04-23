---
id: SPEC-V3R2-HRN-001
title: "Harness Routing + harness.yaml Go Loader"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 5 — Harness + Evaluator"
module: "internal/config/types.go, internal/config/loader.go, internal/harness/, .moai/config/sections/harness.yaml, .moai/config/evaluator-profiles/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-ORC-003
related_problem:
  - P-H06
related_theme: "Layer 5 — Harness, Master §4.5, §7.6 Config sections"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "harness, routing, complexity-estimator, evaluator-profiles, harness-yaml, go-loader, v3r2"
---

# SPEC-V3R2-HRN-001: Harness Routing + harness.yaml Go Loader

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

MoAI v2.13.2 ships `.moai/config/sections/harness.yaml` containing a 3-level harness routing schema (minimal / standard / thorough) with auto-detection rules, escalation triggers, effort mapping, and level-specific evaluator configuration. R6 audit §5.2 reveals the CRITICAL gap: **no Go loader exists for harness.yaml**. The file is template-only. No runtime code reads it; no validation enforces its invariants; Claude Code session-time behavior does not adapt to harness level.

This SPEC delivers the Go loader, typed struct, and runtime enforcement for harness.yaml. It also:

- Implements the Complexity Estimator that maps a SPEC to minimal/standard/thorough based on `file_count`, `domain_count`, `spec_type`, `spec_priority`, and `security_keywords|payment_keywords` presence.
- Wires harness level to effort-level selection per `effort_mapping.minimal=medium|standard=high|thorough=xhigh`.
- Wires harness level to evaluator-profile selection per `levels.{level}.evaluator_profile` (e.g., thorough → strict).
- Implements escalation triggers (quality_gate_fail, review_critical, test_coverage_low) up to `max_escalations: 2`.
- Emits a `moai harness route --spec SPEC-XXX` subcommand that prints the routing decision and its rationale.

Depends on SPEC-V3R2-MIG-003 (config loader addition covers all 5 missing yaml sections; this SPEC owns harness-specific semantics) and SPEC-V3R2-ORC-003 (effort matrix must exist as the target of effort_mapping).

### 1.1 Background

Master §4.5 Layer 5 Harness defines core types:

```go
type Level string
const ( LevelMinimal = "minimal"; LevelStandard = "standard"; LevelThorough = "thorough" )

type EvaluatorProfile struct {
    Name          string
    PassThreshold float64  // floor 0.60 FROZEN
    MaxIterations int
    Escalation    int
    StrictMode    bool
    Rubrics       map[string]Rubric
}
```

with key interfaces `HarnessRouter.Route(spec) Level`, `EvaluatorRunner.Score(contract, artifact) ScoreCard`, `GanLoop.Execute(spec) Result`.

The harness.yaml schema covers:

- `default_profile`: default evaluator profile name
- `mode_defaults`: {solo, team, cg} → auto|thorough
- `auto_detection.rules`: per-level conditions
- `escalation.triggers`: [quality_gate_fail, review_critical, test_coverage_low] with `max_escalations: 2`
- `effort_mapping`: minimal→medium, standard→high, thorough→xhigh
- `levels.{level}`: {description, skip_phases, evaluator (bool), evaluator_mode, sprint_contract, playwright_testing, plan_audit.{enabled, max_iterations, require_must_pass, cross_validate_with_evaluator_active}}
- `model_upgrade_review`: manual checklist for post-upgrade review

This SPEC implements the full struct + loader + runtime usage.

*Source: r6-commands-hooks-style-rules.md §5.2 Unused sections (harness.yaml CRITICAL); problem-catalog.md P-H06; major-v3-master.md §4.5 Layer 5 Harness, §7.6 Config sections table.*

### 1.2 Non-Goals

- Changing the harness.yaml schema (this SPEC implements the loader; schema additions come through CON-002 amendment).
- Implementing evaluator profiles content (`.moai/config/evaluator-profiles/default.yaml` etc.) — this SPEC loads them if present; content authoring is out of scope.
- Implementing the Sprint Contract protocol (SPEC-V3R2-HRN-002 + HRN-003 own Sprint Contract).
- Implementing the hierarchical acceptance scoring (SPEC-V3R2-HRN-003 owns scoring).
- Migrating existing evaluator-active agent body to reference harness levels (the agent reads the loaded config; no body change required).
- Auto-detecting domain count beyond string-matching of domain keywords (more sophisticated domain detection is future work).
- CG (Claude+GLM) mode-specific routing beyond `mode_defaults.cg: thorough` (SPEC-GLM-001 territory for per-tmux overrides).
- Machine-learning-based complexity estimation (rule-based only for v3.0).

---

## 2. Scope (범위)

### 2.1 In Scope

- Implement `internal/config/types.go` struct `HarnessConfig` covering the full harness.yaml schema (default_profile, mode_defaults, auto_detection, escalation, effort_mapping, levels, model_upgrade_review).
- Implement `internal/config/loader.go` function `LoadHarnessConfig(path) (*HarnessConfig, error)` with YAML unmarshal + validator/v10 validation.
- Implement `internal/harness/router.go`:
  - `HarnessRouter.Route(spec *SPEC, cfg *HarnessConfig) (Level, Rationale, error)` applies auto_detection rules in priority order (minimal check first; if fail, standard; if fail, thorough).
  - Complexity Estimator signals: `file_count`, `domain_count`, `spec_type` (bugfix|docs|config|feature|refactor|other), `spec_priority` (derived from frontmatter), presence of security/payment keywords in title + requirements.
- Implement `internal/harness/escalation.go`:
  - `EscalationManager.CheckTriggers(phaseResult)` with counter bounded by `max_escalations`.
- Implement `internal/harness/effort.go`:
  - `EffortForLevel(level Level, cfg *HarnessConfig) string` returns `medium|high|xhigh` per `effort_mapping`.
- CLI subcommand `moai harness route --spec SPEC-XXX [--json]` prints the routing decision.
- CLI subcommand `moai harness validate` validates harness.yaml against the struct + invariants (pass_threshold >= 0.60 floor, max_escalations <= 3, etc.).
- Loader test in `internal/config/loader_test.go` covering successful parse, invalid enum, missing required fields, schema-drift detection.
- Template-first: no harness.yaml content changes required (schema unchanged); Go-side implementation only under `internal/`.
- Includes `harness.yaml` Go loader (originally scoped to MIG-003; promoted here due to phase order — Phase 5 cannot depend on Phase 8). MIG-003 will reference this loader via `internal/config/loader.go` and list the remaining 4 loaders (constitution, context, interview, design).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Changing harness.yaml schema values (schema is FROZEN for v3.0; amendments via CON-002).
- Implementing evaluator-profile YAML files (`default.yaml`, `strict.yaml`, `lenient.yaml`, `frontend.yaml`) — those are authored separately; this SPEC's loader validates them if present.
- Sprint Contract negotiation (SPEC-V3R2-HRN-002 + HRN-003).
- Fresh-memory evaluator flow (SPEC-V3R2-HRN-002 amendment).
- Hierarchical acceptance scoring (SPEC-V3R2-HRN-003).
- Harness telemetry export (beta.1 feature per Master §12 Open Question #3).
- Cross-SPEC harness coordination (per-SPEC scope only).
- Per-team harness routing (team mode inherits route from the parent SPEC).
- User-defined custom levels beyond minimal/standard/thorough (enum is FROZEN for v3.0).
- Changing Claude Code `effortLevel` setting directly (runtime code injects via env; CC reads).

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-beta.1+ (Phase 5)
- Go version: 1.23+
- YAML library: existing project dependency (go-yaml v3)
- Validator library: `github.com/go-playground/validator/v10` (from SPEC-V3-SCH-001 infrastructure)
- Target file: `.moai/config/sections/harness.yaml` (exists in template; unchanged by this SPEC)
- Evaluator profile directory: `.moai/config/evaluator-profiles/` (referenced but content out of scope)
- Subcommand registration: `internal/cli/cmd/harness.go` (new)
- Test coverage target: ≥ 85% on `internal/harness/`
- Lint: `moai agent lint` (ORC-002) unaffected; this SPEC adds `moai harness validate` instead
- FROZEN invariant: `PassThreshold >= 0.60` per design-constitution §5 and Master §1.3

---

## 4. Assumptions (가정)

- SPEC-V3R2-CON-001 has landed; the FROZEN invariant `evaluator.pass_threshold >= 0.60` is encoded.
- SPEC-V3R2-MIG-003 delivers the cross-cutting loader scaffolding (constitution.yaml / context.yaml / interview.yaml / design.yaml / harness.yaml); this SPEC implements harness-specific logic.
- SPEC-V3R2-ORC-003 has landed; the effort matrix is stable; `effort_mapping.*` values align with agent frontmatter.
- The SPEC struct (from SPEC-V3R2-SPC-001) provides enough metadata for the router (frontmatter `priority`, `tags`, body `Requirements` section) to compute complexity signals.
- Auto-detection rules in harness.yaml are authoritative; rule precedence is top-to-bottom (minimal first; fallthrough to standard; fallthrough to thorough).
- Keyword matching is case-insensitive; `security_keywords` include {auth, crypto, encrypt, oauth, jwt, session, password, rbac, acl}; `payment_keywords` include {payment, billing, subscription, invoice, charge, stripe, paypal}.
- domain_count is estimated from comma-separated `domain:` or `related_theme:` frontmatter tokens; fallback 1 if absent.
- file_count is estimated from acceptance.md + plan.md mentioned file paths; fallback 0 if absent.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-HRN-001-001 (Ubiquitous) — 타입 정의**
The file `internal/config/types.go` **shall** define the `HarnessConfig` struct covering at minimum: `DefaultProfile string`, `ModeDefaults map[string]string`, `AutoDetection AutoDetectionConfig`, `Escalation EscalationConfig`, `EffortMapping map[string]string`, `Levels map[string]LevelConfig`, `ModelUpgradeReview ReviewConfig`.

**REQ-HRN-001-002 (Ubiquitous) — 로더 함수**
The file `internal/config/loader.go` **shall** export `LoadHarnessConfig(path string) (*HarnessConfig, error)` that reads the YAML file, unmarshals into the struct, and runs validator/v10.

**REQ-HRN-001-003 (Ubiquitous) — 라우팅 함수**
The file `internal/harness/router.go` **shall** export `Route(spec *SPEC, cfg *HarnessConfig) (Level, Rationale, error)` where `Rationale` is a struct with fields `{matched_rule, file_count, domain_count, spec_type, spec_priority, keywords}`.

**REQ-HRN-001-004 (Ubiquitous) — 에스컬레이션 함수**
The file `internal/harness/escalation.go` **shall** export `EscalationManager.CheckTriggers(ctx EscalationContext) (newLevel Level, escalated bool)` bounded by `cfg.Escalation.MaxEscalations` (default 2).

**REQ-HRN-001-005 (Ubiquitous) — 효과 매핑 함수**
The file `internal/harness/effort.go` **shall** export `EffortForLevel(level Level, cfg *HarnessConfig) string` returning the effort value from `cfg.EffortMapping[string(level)]` (e.g., `"xhigh"` for `LevelThorough`).

**REQ-HRN-001-006 (Ubiquitous) — CLI 서브커맨드**
The `moai` binary **shall** expose two new subcommands: `moai harness route --spec SPEC-XXX [--json]` and `moai harness validate [--path PATH]`.

### 5.2 Event-Driven (이벤트 기반)

**REQ-HRN-001-007 (Event-Driven) — 자동 감지 우선순위**
**When** `Route()` is invoked, the router **shall** evaluate `auto_detection.rules` in the order minimal → standard → thorough; the first matching rule set determines the level.

**REQ-HRN-001-008 (Event-Driven) — 보안/결제 키워드 강제 thorough**
**When** the SPEC title OR any requirement body contains a case-insensitive match for `security_keywords` OR `payment_keywords` OR `spec_priority == critical` OR `domain in [auth, payment, migration, public_api]`, the router **shall** force `LevelThorough` regardless of file/domain counts.

**REQ-HRN-001-009 (Event-Driven) — 에스컬레이션 발동**
**When** a harness phase result indicates `quality_gate_fail` OR `review_critical` OR `test_coverage_low` (per escalation.triggers), the EscalationManager **shall** bump the level by one tier (minimal → standard → thorough) subject to `MaxEscalations` cap.

**REQ-HRN-001-010 (Event-Driven) — 검증 실패 처리**
**When** `LoadHarnessConfig()` encounters a validation error (missing required field, invalid enum, pass_threshold < 0.60), the function **shall** return an error wrapping the validator/v10 error with the field name and illegal value.

**REQ-HRN-001-011 (Event-Driven) — JSON 출력**
**When** `moai harness route` is invoked with `--json`, the CLI **shall** print a single JSON document with fields `{level, rationale: {...}, effort, evaluator_profile, sprint_contract (bool), plan_audit (bool)}`.

### 5.3 State-Driven (상태 기반)

**REQ-HRN-001-012 (State-Driven) — FROZEN 문턱값 보존**
**While** the v3.0.0 minor cycle is active, the loader **shall** reject any `HarnessConfig` with `levels.{any}.evaluator_profile` pointing to a profile file whose `pass_threshold < 0.60` (FROZEN floor per design-constitution §5).

**REQ-HRN-001-013 (State-Driven) — max_escalations 상한**
**While** the runtime is active, the EscalationManager **shall** cap total escalations per phase at `cfg.Escalation.MaxEscalations` (default 2, hard ceiling 3).

**REQ-HRN-001-014 (State-Driven) — 모드 기본값 준수**
**While** `workflow.yaml execution_mode` is `auto`, the router **shall** consult `cfg.ModeDefaults.solo|team|cg` when no SPEC override is set; `cg` mode **shall** force thorough (per schema default).

### 5.4 Optional (선택)

**REQ-HRN-001-015 (Optional) — 하니스 오버라이드**
**Where** a SPEC frontmatter declares `harness_level: minimal|standard|thorough`, the router **may** honor the override instead of running auto-detection; the rationale field **shall** record `matched_rule: spec_override`.

**REQ-HRN-001-016 (Optional) — 모델 업그레이드 검토**
**Where** `cfg.ModelUpgradeReview.Trigger.OnModelChange == true` and a model version change is detected, the CLI **may** emit a reminder pointing at the checklist in `cfg.ModelUpgradeReview.Checklist`.

### 5.5 Unwanted Behavior

**REQ-HRN-001-017 (Unwanted Behavior) — 알 수 없는 레벨 거부**
**If** any harness.yaml field references a level name outside `{minimal, standard, thorough}`, **then** `LoadHarnessConfig()` **shall** return error `HRN_UNKNOWN_LEVEL` naming the offending key.

**REQ-HRN-001-018 (Unwanted Behavior) — 에스컬레이션 초과 금지**
**If** EscalationManager is invoked after `max_escalations` has been reached, **then** it **shall** return `escalated: false` and emit log `HRN_ESCALATION_CAP_REACHED`; the workflow **shall** fall back to user AskUserQuestion via the orchestrator.

**REQ-HRN-001-019 (Unwanted Behavior) — 스키마 드리프트 감지**
**If** the harness.yaml file contains keys not present in the `HarnessConfig` struct, **then** `LoadHarnessConfig()` **shall** emit a warning `HRN_SCHEMA_DRIFT` listing the unknown keys; the warning is non-blocking unless `MOAI_CONFIG_STRICT=1` (in which case it becomes an error).

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-HRN-001-01**: `go test ./internal/config/... ./internal/harness/...` passes with ≥ 85% coverage.
- **AC-HRN-001-02**: `moai harness route --spec SPEC-V3R2-ORC-001` returns `level: standard` with rationale listing file_count > 3 AND multi_domain (per auto_detection.rules.standard).
- **AC-HRN-001-03**: `moai harness route --spec SPEC-V3R2-HRN-002` (contains `evaluator` security/amendment keywords) returns `level: thorough`.
- **AC-HRN-001-04**: `moai harness validate` on the shipping harness.yaml exits 0.
- **AC-HRN-001-05**: Editing harness.yaml to add `pass_threshold: 0.5` in a profile file causes `moai harness validate` to exit 1 with error `HRN_PASS_THRESHOLD_FLOOR` (0.60 FROZEN).
- **AC-HRN-001-06**: `moai harness route --json` emits valid JSON with the documented schema.
- **AC-HRN-001-07**: Setting `MOAI_CONFIG_STRICT=1` and introducing an unknown key causes loader to error with `HRN_SCHEMA_DRIFT`.
- **AC-HRN-001-08**: Escalation test: simulating 3 consecutive `quality_gate_fail` at minimal level bumps level by 2 (minimal → standard → thorough) and caps at thorough.
- **AC-HRN-001-09**: SPEC frontmatter with `harness_level: thorough` overrides auto-detection; rationale shows `matched_rule: spec_override`.
- **AC-HRN-001-10**: Effort mapping returns `medium` for minimal, `high` for standard, `xhigh` for thorough.

---

## 7. Constraints (제약)

- [HARD] FROZEN `pass_threshold >= 0.60` per design-constitution §5; loader enforces (REQ-012).
- [HARD] FROZEN level enum `{minimal, standard, thorough}` for v3.0 (REQ-017).
- [HARD] Max 3 escalations per phase (hard ceiling; configurable default 2 per REQ-013).
- [HARD] No network calls during loader run (reproducible CI).
- [HARD] Go-only implementation; validator/v10 dependency already approved in SPEC-V3-SCH-001.
- [HARD] Template-First (CLAUDE.local.md §2) — no content changes to harness.yaml itself.
- [HARD] 16-language neutrality — routing logic language-agnostic; no language-specific assertions.
- [HARD] Backward compatibility — existing evaluator-active agent and GAN-loop skill continue to function; this SPEC does not change their bodies, only wires config.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                               | Impact | Mitigation                                                                                                  |
|--------------------------------------------------------------------|--------|-------------------------------------------------------------------------------------------------------------|
| Keyword matcher false-positive on unrelated docs (e.g., "payment" in README example) | MEDIUM | Scope matching to SPEC title + requirements section only, not body prose                                    |
| file_count and domain_count estimation unreliable for sparse SPECs | MEDIUM | Fallback 0 → minimal; contributors can `harness_level:` override                                            |
| Escalation cycle A→B→A livelock                                    | MEDIUM | Cap at max_escalations (REQ-013); beyond cap, orchestrator surfaces AskUserQuestion (REQ-018)               |
| Evaluator profile file missing at runtime                          | MEDIUM | Loader emits warning and falls back to `default_profile`; test fixture covers                              |
| FROZEN pass_threshold floor misinterpreted by lenient-profile author | HIGH   | REQ-012 explicit runtime check; lenient profile can go as low as 0.60 but not below                         |
| Schema drift from future harness.yaml changes                      | LOW    | REQ-019 warning + strict-mode error; CI check for struct-drift                                              |
| Harness override abused to skip thorough on critical SPECs          | MEDIUM | REQ-008 security/payment keyword force-thorough overrides spec_override; document precedence                |
| Existing v2 SPECs lacking frontmatter fields cause router to default to minimal | LOW    | Default fallthrough is standard per REQ-007 fallthrough order                                               |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN floor invariant)
- **SPEC-V3R2-MIG-003** (Config loader scaffolding)
- **SPEC-V3R2-ORC-003** (Effort matrix)
- **SPEC-V3R2-SPC-001** (SPEC struct for complexity signals)

### 9.2 Blocks

- **SPEC-V3R2-HRN-002** (Evaluator memory amendment) — uses the typed harness config to declare `memory_scope: per_iteration`.
- **SPEC-V3R2-HRN-003** (Hierarchical scoring) — reads evaluator profiles per level.
- **SPEC-V3R2-WF-003** (Multi-mode router) — consults harness level when selecting autopilot/loop/team.

### 9.3 Related

- **SPEC-V3R2-EVAL-001** (v3-legacy evaluator profile) — provides the profile YAML schema consumed by this loader.
- **SPEC-V3-SCH-001** (validator/v10 infrastructure).
- **SPEC-GLM-001** (GLM compatibility) — mode_defaults.cg reference.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01 (struct compiles); REQ-002 → AC-01, AC-04; REQ-003 → AC-02, AC-03; REQ-004 → AC-08; REQ-005 → AC-10; REQ-006 → AC-02, AC-04, AC-06; REQ-007 → AC-02, AC-03; REQ-008 → AC-03; REQ-009 → AC-08; REQ-010 → AC-05; REQ-011 → AC-06; REQ-012 → AC-05; REQ-013 → AC-08; REQ-014 → cg-mode regression; REQ-015 → AC-09; REQ-016 → model-upgrade regression; REQ-017 → unknown-level regression; REQ-018 → AC-08; REQ-019 → AC-07.
- Total REQ count: 19 (Ubiquitous 6, Event-Driven 5, State-Driven 3, Optional 2, Unwanted 3)
- Expected AC count: 10
- Wave 1/2 sources:
  - `r6-commands-hooks-style-rules.md` §5.2 Unused config sections (harness.yaml CRITICAL gap)
  - `problem-catalog.md` P-H06 (CRITICAL)
  - `major-v3-master.md` §4.5 Layer 5 Harness (core types), §7.6 Config sections (loader addition required)
  - `.moai/config/sections/harness.yaml` (existing schema, unchanged)
  - `design-principles.md` Principle 4 (Evaluator Fresh Judgments) — harness routing enables the fresh-memory flow per HRN-002
  - `pattern-library.md` E-1 Agent-as-a-Judge, E-3 Rubric-Anchored Independent Re-eval
- Code-side paths:
  - `internal/config/types.go` (modified, REQ-001)
  - `internal/config/loader.go` (modified, REQ-002)
  - `internal/config/loader_test.go` (modified, REQ-002 tests)
  - `internal/harness/router.go` (new, REQ-003, REQ-007, REQ-008, REQ-015)
  - `internal/harness/escalation.go` (new, REQ-004, REQ-009, REQ-013, REQ-018)
  - `internal/harness/effort.go` (new, REQ-005)
  - `internal/cli/cmd/harness.go` (new, REQ-006, REQ-011)
  - `internal/harness/router_test.go` (new, AC regression fixtures)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1060` (§11.5 HRN-001 definition)
  - `docs/design/major-v3-master.md:L992` (§9 Phase 5 Harness + Evaluator)
  - `docs/design/major-v3-master.md:L972` (§8 BC-V3R2-013 — config loaders context, referenced by HRN-001's harness.yaml loader)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-H06)

---

End of SPEC.
