---
spec_id: SPEC-V3R5-INIT-WIZARD-EXPANSION-001
version: "0.1.0"
status: draft
---

# Acceptance Criteria — INIT Wizard Decision-Point Expansion

## Conventions

- Binary PASS/FAIL — each AC has a verifiable command + expected output
- "(scoped to Phase 1)" means the AC only applies to candidates B1, B2, B3, B5, B8 in this SPEC run-phase
- Phase 2 candidates (B4/B6/B7) are NOT covered by these ACs; their own ACs ship in a follow-up post-P2/P4 SPEC

## AC-IWE-001 — project.mode wizard exposure (REQ-IWE-001)

**Verification**:
```bash
grep -n 'ID:[[:space:]]*"project_mode"' internal/cli/wizard/questions.go
grep -n 'ProjectMode' internal/cli/wizard/types.go
grep -n 'ProjectMode' internal/core/project/initializer.go
```

**Expected**:
- `questions.go` contains a Question entry with `ID: "project_mode"`, `Type: QuestionTypeSelect`, two Options `personal` (recommended) + `team`
- `types.go` `WizardResult` struct contains `ProjectMode string` field
- `initializer.go` writes `mode: <value>` to `.moai/config/sections/project.yaml` (or whichever section file owns `project.mode` per existing layout — verify via `find internal/template/templates/.moai/config -name "*.yaml"` then grep `mode:`)

## AC-IWE-002 — harness.default_profile dynamic enumeration (REQ-IWE-002)

**Verification**:
```bash
grep -n 'evaluator-profiles' internal/cli/wizard/questions.go
ls .moai/config/evaluator-profiles/ | sort
```

**Expected**:
- `questions.go` reads profile filenames at wizard-construction time from `.moai/config/evaluator-profiles/*.md` (or the embedded template path) and builds Options dynamically
- Options reflect actual filesystem state — adding a new `.md` profile file auto-extends the wizard without code change
- Selected value persisted to `harness.yaml` under key `harness.default_profile`

## AC-IWE-003 — lsp.enabled opt-in (REQ-IWE-003)

**Verification**:
```bash
grep -n 'lsp_enabled\|LspEnabled\|LSPEnabled' internal/cli/wizard/questions.go internal/cli/wizard/types.go internal/core/project/initializer.go
```

**Expected**:
- Wizard presents a Confirm question (or Select with yes/no) for LSP master switch
- Default `false` (matches `lsp.yaml.tmpl:45 enabled: false`)
- User selection persists to the rendered `lsp.yaml` under `lsp.enabled`
- If LSP is disabled (default), the per-language server matrix remains untouched

## AC-IWE-004 — quality gates double-confirm (REQ-IWE-004)

**Verification**:
```bash
grep -n 'enforce_quality\|coverage_exemptions' internal/cli/wizard/questions.go internal/cli/wizard/types.go
grep -n 'EnforceQuality\|CoverageExemptions' internal/core/project/initializer.go
cat internal/template/templates/.moai/config/sections/*quality*.yaml 2>/dev/null || true
```

**Expected**:
- Two Confirm questions: `enforce_quality` (default `true`) + `coverage_exemptions_enabled` (default `false`)
- `WizardResult` extended with `EnforceQuality bool` + `CoverageExemptionsEnabled bool`
- `initializer.go` `qualityContent` template extended with `coverage_exemptions:\n  enabled: %t\n` block under `constitution:`
- Existing `enforce_quality: %t` line preserves current format

## AC-IWE-005 — design opt-in double-confirm (REQ-IWE-005)

**Verification**:
```bash
grep -n 'design_enabled\|claude_design' internal/cli/wizard/questions.go internal/cli/wizard/types.go
grep -n 'DesignEnabled\|ClaudeDesignEnabled' internal/core/project/initializer.go
```

**Expected**:
- Two Confirm questions: `design_enabled` (default `true`) + `claude_design_enabled` (default `true`)
- Both persisted to `.moai/config/sections/design.yaml` under `design.enabled` and `design.claude_design.enabled`
- When `design_enabled=false`, `claude_design_enabled` question is conditionally skipped (using existing `Question.Condition func`) — selecting "no design" auto-implies "no claude_design"

## AC-IWE-006 — --standard flag presents Phase 1, default skips (REQ-IWE-006)

**Verification**:
```bash
grep -n '"standard"' internal/cli/init.go
grep -n 'getBoolFlag(cmd, "standard")\|StandardMode' internal/cli/init.go internal/cli/wizard/*.go
# E2E: non-interactive smoke
moai init test-quick --non-interactive 2>&1 | tee /tmp/init-quick.log
grep -E "enforce_quality:|harness.default_profile:|design.enabled:|lsp.enabled:" test-quick/.moai/config/sections/*.yaml | head -10
rm -rf test-quick
```

**Expected**:
- `init.go` registers `--standard` Bool flag with default `false`
- When flag absent, Phase 1 questions are filtered out (via existing `Question.Condition func` pattern or a `StandardMode` boolean in WizardResult)
- Quick mode (`moai init <name>` with no flag) writes yaml using existing defaults — diff from current behavior is **zero**
- `--standard` triggers all 5 Phase 1 questions in addition to current 9

## AC-IWE-007 — --advanced flag honors P2/P4 readiness (REQ-IWE-007)

**Verification**:
```bash
grep -n '"advanced"' internal/cli/init.go
grep -n 'P2_READY\|advancedReady\|advancedAvailable' internal/cli/wizard/*.go internal/cli/init.go
# Phase 2 dependency check: verify the wizard tests for nested struct presence
grep -rn 'BranchCreation\|CommitStyle' internal/config/types.go pkg/models/config.go | head -5
```

**Expected**:
- `init.go` registers `--advanced` Bool flag implying `--standard`
- Wizard detects nested struct readiness via build-tag-free Go reflection or explicit feature flag (e.g., `bodp.IsAdvancedWizardReady() bool` lives in a small new file `internal/cli/wizard/advanced_gate.go`)
- If P2 struct missing: skip B6/B7 + emit `_, _ = fmt.Fprintf(cmd.OutOrStderr(), "warning: --advanced features %s skipped (depends on SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 run-phase)\n", "git-strategy.*")` 
- If P4 struct missing: skip B4 + emit equivalent warning for workflow.team
- When both prerequisites missing, `--advanced` behaves identically to `--standard` (no crash)

## AC-IWE-008 — non-interactive override flags (REQ-IWE-008)

**Verification**:
```bash
grep -n '"enforce-quality"\|"enable-lsp"\|"harness-profile"\|"project-mode"\|"enable-design"' internal/cli/init.go
# Smoke: each flag overrides yaml default
moai init test-noninteract --non-interactive --enforce-quality=false --enable-lsp=true --harness-profile strict --project-mode team --enable-design=false 2>&1
grep -E "mode:|enforce_quality:|default_profile:|enabled:" test-noninteract/.moai/config/sections/*.yaml | head -10
rm -rf test-noninteract
```

**Expected**:
- 5 new CLI flags registered: `--enforce-quality`, `--enable-lsp`, `--harness-profile <string>`, `--project-mode <string>`, `--enable-design`
- All flags honored under `--non-interactive` (no wizard shown)
- Resulting yaml reflects flag values: `enforce_quality: false`, `lsp.enabled: true`, `harness.default_profile: strict`, `project.mode: team`, `design.enabled: false`
- Bool flags accept `=true`/`=false` per Cobra convention

## AC-IWE-009 — TRUST coverage ≥85% for new code (REQ-IWE-011)

**Verification**:
```bash
go test -cover ./internal/cli/wizard/... 2>&1 | tee /tmp/cover-wizard.log
go test -cover ./internal/core/project/... 2>&1 | tee /tmp/cover-project.log
# Confirm new test file exists
ls internal/cli/wizard/expansion_test.go internal/core/project/initializer_expansion_test.go
```

**Expected**:
- `internal/cli/wizard/...` aggregate coverage ≥85% (baseline pre-SPEC + new code)
- `internal/core/project/...` aggregate coverage ≥85%
- New test file `expansion_test.go` exists with at least 10 test cases covering:
  - Each new Question rendered correctly (Title, Description, Options)
  - Each new yaml-write path produces expected file content
  - `--standard` flag inclusion/exclusion behavior
  - `--advanced` flag with both P2/P4 ready and not-ready states (mock the gate)
  - `--non-interactive` with each override flag
- Cross-platform: `GOOS=darwin go test ./internal/cli/wizard/...` AND `GOOS=windows GOARCH=amd64 go build ./...` both PASS

## AC-IWE-010 — backward compatibility / Quick mode invariance

**Verification**:
```bash
# Snapshot existing Quick mode output before changes
mkdir -p /tmp/baseline-quick && cd /tmp/baseline-quick
moai init test-baseline --non-interactive 2>&1
find test-baseline/.moai/config/sections -type f -name "*.yaml" | xargs sha256sum > /tmp/quick-baseline.sums
# After SPEC implementation:
cd $REPO && moai init /tmp/test-after --non-interactive 2>&1
find /tmp/test-after/.moai/config/sections -type f -name "*.yaml" | xargs sha256sum > /tmp/quick-after.sums
diff /tmp/quick-baseline.sums /tmp/quick-after.sums
```

**Expected**:
- Diff is empty for files unrelated to the 5 new keys
- For yaml files touching the 5 new keys: only the new keys appear; existing keys + values remain bit-identical (no whitespace or ordering changes)
- This AC SHOULD be deferred to run-phase (cannot establish baseline before code lands); plan-phase only needs assertion of intent

## Edge Cases

### EC-1 — User upgrade path
**Given** a user runs `moai init` on Quick mode (no flag), commits, then later runs `moai init --standard --force` on the same directory,
**When** the wizard executes,
**Then** existing Quick-mode yaml values are preserved as defaults in the new prompts (not overwritten), and only the 5 new keys are added.

Verification: `internal/core/project/initializer.go` `--force` path reads existing yaml as defaults and merges new keys without destroying existing values. (Existing `Force bool` flag in InitOptions; this AC verifies it correctly handles merge.)

### EC-2 — Missing evaluator-profile directory
**Given** the `.moai/config/evaluator-profiles/` directory is empty or absent at wizard-construction time,
**When** the Phase 1 wizard reaches the B2 question,
**Then** the wizard falls back to the hardcoded canonical list `[default, strict, lenient, frontend]` and emits a stderr informational message naming the missing directory.

Verification: `grep -n 'fallback\|hardcoded\|canonical' internal/cli/wizard/questions.go` shows the fallback path.

### EC-3 — --advanced without --standard
**Given** a user invokes `moai init --advanced` (without explicit `--standard`),
**When** the CLI parses flags,
**Then** `--advanced` implies `--standard` (Phase 1 is included automatically), so the user sees Quick + Phase 1 + Phase 2 (when prerequisites available).

Verification: `init.go` flag-resolution logic explicitly sets `standard=true` when `advanced=true`.
