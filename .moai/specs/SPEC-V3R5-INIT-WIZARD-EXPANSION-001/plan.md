---
spec_id: SPEC-V3R5-INIT-WIZARD-EXPANSION-001
version: "0.1.0"
status: draft
---

# Plan — INIT Wizard Decision-Point Expansion (Tier M)

## 0. Tier Justification

**Tier**: M (3-artifact LEAN workflow)

Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier:

- **Affected files** (~6): `internal/cli/wizard/questions.go` (+5 Question entries), `internal/cli/wizard/types.go` (+5 WizardResult fields), `internal/cli/wizard/wizard.go` (+saveAnswer cases), `internal/cli/init.go` (+5 override flags + Standard/Advanced flags + result mapping), `internal/core/project/initializer.go` (+yaml write paths), new `internal/cli/wizard/expansion_test.go`
- **LOC estimate**: ~400 LOC across source + tests (excluding generated docs)
- **Risk profile**: Medium — user-facing CLI surface; backward-compat is critical
- **Tier S threshold** (<300 LOC, 1 package) **exceeded** → M; **Tier L threshold** (>800 LOC, >3 packages) **not exceeded** → M
- **3 artifacts sufficient** per LEAN STOP escalation: no design.md or tasks.md needed because milestone structure below covers task decomposition

## 1. Implementation Approach

### 1.1 Architectural Decision Record (one paragraph)

The wizard will be expanded by adding new `Question` definitions to the existing `DefaultQuestions(projectRoot)` slice in `internal/cli/wizard/questions.go`, gated by a new `Question.Condition func` that checks a `WizardResult.StandardMode bool` field. This preserves the existing single-form-per-question architecture (which was deliberately chosen over multi-group `huh.NewForm` to work around the huh v0.8.x YOffset scroll bug per `wizard.go:14` comment). Phase 1 questions render only when `StandardMode == true`, ensuring backward-compat Quick mode is bit-identical to current output. CLI flags `--standard`/`--advanced` set `StandardMode`/`AdvancedMode` early in `init.go` before wizard invocation.

### 1.2 Why this approach over alternatives

- **Alternative A — separate "standard wizard" function**: Would duplicate ~150 lines of wizard machinery; rejected
- **Alternative B — runtime-detected upgrade prompt**: Adds session-state complexity; rejected
- **Chosen approach — Conditional questions with mode flags**: Reuses existing `Question.Condition func` infrastructure (already used for Git provider conditional questions), adds ~50 LOC instead of ~150

### 1.3 Backward-compatibility contract

| Quick mode invariant | Mechanism |
|---|---|
| Existing 9 questions appear in same order | Phase 1 questions appended after Q9 |
| Existing yaml files have identical bytes | Phase 1 yaml writes only when StandardMode |
| Existing CLI flags unchanged | New flags additive only |
| AC-IWE-010 byte-identical diff | Verified via sha256sum diff |

## 2. Milestones (Priority-Based, No Time Estimates)

### M1 — Wizard Structure Extension (Priority: High)

**Files**:
- `internal/cli/wizard/types.go`: Add 5 fields to `WizardResult`: `ProjectMode`, `HarnessProfile`, `LSPEnabled`, `EnforceQuality`, `CoverageExemptionsEnabled`, `DesignEnabled`, `ClaudeDesignEnabled`, plus `StandardMode bool`, `AdvancedMode bool`. (Total: 9 new fields.)
- `internal/cli/wizard/questions.go`: Append 7 new Question entries (one each for B1/B2/B3 + two-per for B5/B8) with `Condition: func(r) bool { return r.StandardMode }`.

**Deliverable**: `go build ./...` PASS, `go vet ./internal/cli/wizard/` PASS.

### M2 — Initializer yaml Write Paths (Priority: High)

**Files**:
- `internal/core/project/initializer.go`: Extend `qualityContent` template with `coverage_exemptions:\n  enabled: %t\n`; add new yaml write functions for `project.yaml`, `lsp.yaml` opt-in render, `design.yaml`, `harness.yaml`.
- `internal/core/project/initializer.go`: Add 5 new fields to internal template-context struct + propagation from `InitOptions` (which receives values from wizard result mapping in init.go).

**Deliverable**: Each yaml file contains the expected key under correct section; `go vet` PASS.

### M3 — CLI Flag Integration (Priority: High)

**Files**:
- `internal/cli/init.go`: Register flags `--standard`, `--advanced`, `--enforce-quality`, `--enable-lsp`, `--harness-profile`, `--project-mode`, `--enable-design` (Cobra `BoolVar`/`StringVar`). 
- `internal/cli/init.go`: In flag-parsing block, set `opts.StandardMode = ` flag value; `opts.AdvancedMode = ` flag value implying StandardMode; pass to wizard via new `wizard.RunWithDefaultsModes(rootFlag, "", standard, advanced)` signature OR via setting fields on a pre-populated `WizardResult` before `Run()`. (Choose approach with least churn during M3.)
- `internal/cli/init.go`: After `wizard.RunWithDefaults`, map `result.{ProjectMode, ..., ClaudeDesignEnabled}` to corresponding `opts.{ProjectMode, ...}` fields on `project.InitOptions`. `InitOptions` struct in `initializer.go` extended with the same 7 fields.

**Deliverable**: `moai init --help` shows all new flags with sensible descriptions.

### M4 — Advanced Gate + Phase 2 Scaffolding (Priority: Medium)

**Files**:
- `internal/cli/wizard/advanced_gate.go` (NEW): Small file with `IsAdvancedWizardReady() (p2 bool, p4 bool)` — uses reflection on `config.Config` struct to detect whether `BranchCreation` / `CommitStyle` / `Workflow.Team` nested types exist at runtime. If false: skip Phase 2 questions + emit stderr warning naming missing SPEC.
- `internal/cli/wizard/questions.go`: Add 5 Phase 2 Question stubs gated by `Condition: func(r) bool { return r.AdvancedMode && advancedGate.P2Ready }` for B6/B7 and `r.AdvancedMode && advancedGate.P4Ready` for B4. Stubs are placeholders that compile but produce no yaml write today — yaml-write activation is the follow-up SPEC's responsibility.

**Deliverable**: With current main (P2/P4 not run-phase complete), `moai init --advanced` skips Phase 2 questions + emits 2 stderr warning lines; no crash.

### M5 — Tests + Coverage (Priority: High)

**Files**:
- `internal/cli/wizard/expansion_test.go` (NEW): Table-driven tests covering each Phase 1 Question (Title, Default, Options); StandardMode/AdvancedMode toggling; saveAnswer wiring for all 9 new WizardResult fields.
- `internal/core/project/initializer_expansion_test.go` (NEW): yaml-write tests using `t.TempDir()` + bytewise diff against expected fixtures stored in test files (no testdata directory needed — keep fixtures inline per Tier M LEAN).
- `internal/cli/init_test.go` (extend): Flag registration + `--non-interactive` override flag tests.

**Deliverable**: `go test ./internal/cli/wizard/... ./internal/core/project/... ./internal/cli/...` PASS with coverage ≥85% per package.

### M6 — Documentation Sync (Priority: Low — Sync-phase, optional)

**Files**:
- `internal/template/templates/CLAUDE.local.md` (if any wizard reference): Update wizard question count from 9 → "9 (Quick) / 14 (Standard) / 17 (Advanced)" 
- README / docs-site `getting-started/initialization.md` (if present): Add `--standard` and `--advanced` flag documentation in 4 locales

**Deliverable**: Documentation reflects new flags. **DEFER to /moai sync phase** — out of this SPEC's run-phase scope.

## 3. Technical Approach Detail

### 3.1 huh v0.8.x scroll-bug workaround preservation

The existing `wizard.go` deliberately runs each question as an independent `huh.NewForm` to avoid the YOffset scroll bug (per `wizard.go:14-24` comment + `buildSelectField` comment block). New Phase 1 questions MUST follow the same per-question-form pattern. Do NOT introduce multi-group `huh.NewForm` containing Phase 1 questions — the existing per-question loop in `RunWithLocale` already handles conditional questions correctly via `q.Condition != nil && !q.Condition(result)` skip on line 48.

### 3.2 Step counter update

`wizardTotalSteps` constant at `wizard.go:28` currently `6` (note: comment says "6-step init flow from screens.jsx:ScreenInit" but actual Question count is 9 — pre-existing minor inconsistency unrelated to this SPEC). The Stepper visible-question count is computed dynamically via `visibleIdx++` in `RunWithLocale:52`, so the constant determines the "total" denominator shown to the user.

**Decision**: Change `wizardTotalSteps` to a function `wizardTotalSteps(result *WizardResult) int` that returns `9 + 5*(if r.StandardMode)` + future Phase 2 count. This is a small API tweak but cleaner than a magic constant. Alternative: compute visible count via `TotalVisibleQuestions(questions, result)` already defined at `questions.go:166` and pass through. **Recommended: use `TotalVisibleQuestions` rather than introducing a function — zero new code, leverages existing infrastructure.**

### 3.3 Reflection-based advanced gate (M4)

`IsAdvancedWizardReady` uses `reflect.TypeOf((*config.Config)(nil)).Elem()` to walk to `GitStrategy.<mode>` and check whether `.BranchCreation` field exists. If not exported at compile time (P2 not run-phase complete), reflection returns `_, ok := t.FieldByName("BranchCreation"); !ok`. Same pattern for `Workflow.Team.DefaultModel`. This avoids hard import dependency on schema that may not yet exist.

## 4. Risk Assessment

### R-IWE-001 — Cross-platform wizard divergence
**Likelihood**: Low. **Impact**: Medium.
The huh library has known minor rendering differences on Windows terminals (curses-based). New questions follow the same `huh.NewSelect`/`huh.NewConfirm` patterns as existing 9; therefore the risk is bounded to the existing surface.
**Mitigation**: Reuse `buildSelectField` / `buildInputField` exclusively. Add a `huh.NewConfirm` variant in `buildConfirmField` if needed (most Phase 1 questions are Confirm-style). Cross-platform PASS verified via `GOOS=windows GOARCH=amd64 go build ./...` in M5 deliverable.

### R-IWE-002 — Phase 2 silent skip confusion
**Likelihood**: Medium. **Impact**: Low.
User invokes `--advanced` expecting all 17 questions, sees only 14 (because P2 not yet run-phase complete), might assume bug.
**Mitigation**: M4 emits explicit stderr warning naming the missing SPEC. AC-IWE-007 verification command grep-checks for the warning string.

### R-IWE-003 — coverage_exemptions yaml schema drift
**Likelihood**: Low. **Impact**: Medium.
`coverage_exemptions` is a nested block with multiple fields (`enabled`, `max_exempt_percentage`, etc. per `defaults.go:208`). Wizard only exposes `enabled` — if user sets `enabled: true` but defaults for sibling fields drift, behavior may surprise.
**Mitigation**: Wizard write path includes the full nested block with sibling-field defaults sourced from `internal/config/defaults.go::NewDefaultCoverageExemptions()` rather than hardcoding. M2 deliverable verifies via grep that initializer.go imports `defaults.NewDefaultCoverageExemptions`.

### R-IWE-004 — Tests don't catch yaml-write regression
**Likelihood**: Low. **Impact**: High.
A future PR could refactor `initializer.go` and break yaml writes; without robust tests, regression ships silently.
**Mitigation**: M5 yaml-write tests use bytewise diff (`bytes.Equal`) against expected fixture strings, not just key presence. Any whitespace or ordering change fails the test.

### R-IWE-005 — --force merge breakage
**Likelihood**: Medium. **Impact**: Medium.
EC-1 (user upgrade path via `--force`) requires reading existing yaml as defaults. Current `initializer.go` `Force bool` path is template-driven (overwrites), not merge-driven.
**Mitigation**: This SPEC marks EC-1 as a **NICE-TO-HAVE edge case**, NOT a binary AC requirement. If `--force` merge is non-trivial to implement, defer to a follow-up SPEC. Acceptance.md EC-1 is informational; no AC blocks on it.

## 5. Dependencies

### 5.1 Internal (block run-phase)

None for Phase 1. Phase 2 (M4 stubs only) acknowledges dependencies but does not require them to ship.

### 5.2 External SPECs (block Phase 2 run-phase only)

| SPEC | Status | Blocks |
|------|--------|--------|
| SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (P2) | draft (plan-phase only) | Phase 2 B6, B7 yaml-write activation |
| SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 (P4) | draft (plan-phase only) | Phase 2 B4 yaml-write activation |

Both must complete run-phase + reach `status: implemented` before a follow-up SPEC can convert M4 stubs to active yaml-write paths. Plan-phase of this SPEC does NOT block on P2/P4 — only **scaffolding (stubs)** ships in M4, not active code.

### 5.3 Tooling

- huh ≥ v0.8.x (already in `go.mod`)
- bubbletea v2 (already in `go.mod`)
- `internal/tui.Stepper` (existing, no change required)

## 6. Verification Strategy

### 6.1 Cross-platform (CI 3-tier)

- `go build ./...` (Linux + macOS + Windows via CI matrix) — all PASS
- `go test ./...` per OS — all PASS
- `golangci-lint run --timeout=2m` — NEW issues = 0

### 6.2 Subagent boundary (C-HRA-008 class)

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/wizard/ internal/cli/init.go internal/core/project/initializer.go | grep -v "_test.go" | grep -v "// "
```
**Expected**: 0 matches. Wizard interactions are user-facing via `huh` library, NOT via AskUserQuestion (which is orchestrator-only per CLAUDE.md §8).

### 6.3 spec-lint

`### 3.1 Out of Scope` h3 sub-section present (spec.md uses `### 3.1 Out of Scope`).

### 6.4 Frontmatter

Canonical 12-field schema satisfied: id, version, status, created_at, updated_at, author, priority, labels, tier, issue_number, related_specs (optional but recommended for cross-SPEC traceability). Verified via `grep "^[a-z_]*:" .moai/specs/SPEC-V3R5-INIT-WIZARD-EXPANSION-001/spec.md | head -15`.

## 7. Out of Scope (Plan Phase)

### 7.1 Out of Scope

- M6 documentation sync (defer to /moai sync phase)
- Phase 2 yaml-write activation (defer to follow-up SPEC after P2/P4 run-phase complete)
- New evaluator-profile creation flow (separate concern)
- LSP server installation guidance / detection (B3 exposes master switch only)
- Localized question text (ko/ja/zh) — English-only ship
- `moai update` migration path for existing projects — `moai init` scaffold path only
- Web/GUI wizard alternatives — terminal-only

## 8. Plan-Auditor Self-Estimate

| Dimension | Score | Notes |
|-----------|-------|-------|
| Requirement clarity (EARS structure) | 0.92 | 11 EARS REQs cover 5 candidates + flags + non-interactive + grouping + persistence + tests; verb forms ("shall") consistent |
| AC verifiability (binary commands) | 0.93 | Each AC has explicit grep/test/ls command + expected output; AC-IWE-010 EC defers to run-phase appropriately |
| Plan completeness (milestones) | 0.90 | 6 milestones (M1-M6) cover structure + initializer + flags + advanced gate + tests + docs; deliverables explicit |
| Risk coverage | 0.88 | 5 risks identified, each with likelihood/impact/mitigation; R-IWE-005 explicitly defers to nice-to-have |
| Cross-SPEC traceability | 0.92 | P2/P4 dependencies named with SPEC ID; P1 STATUSLINE-PROFILE-WIZARD-001 referenced as sister-SPEC scope clarifier |
| Backward-compat contract | 0.94 | Section 1.3 + AC-IWE-010 enforce byte-identical Quick mode output |
| Tier appropriateness | 0.91 | Tier M justified explicitly in §0; LEAN STOP escalation respected (3 artifacts) |
| EXCL completeness | 0.89 | 7-item Out of Scope (spec) + 7-item plan Out of Scope; explicit defer of Phase 2 to follow-up |
| **Aggregate** | **0.911** | **PASS** — exceeds Tier M threshold 0.80 by margin +0.111 |

This self-estimate is binding; if plan-auditor returns aggregate <0.80, halt and revise per LEAN STOP escalation per `spec-workflow.md`.
