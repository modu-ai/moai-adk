# Acceptance Criteria — SPEC-MERGE-METHOD-CONFIG-001

> Plan-phase artifact. Each AC is observable (test output, grep result, byte-equivalence). REQ traceability noted per AC.

## D. AC Matrix

| AC | REQ | Verification kind | Severity |
|----|-----|-------------------|----------|
| AC-MMC-001 | REQ-MMC-001/002 | Go unit (defaults) | MUST |
| AC-MMC-002 | REQ-MMC-003 | Go unit (loader absent) | MUST |
| AC-MMC-003 | REQ-MMC-004 | Go unit (loader present) | MUST |
| AC-MMC-004 | REQ-MMC-003/004 | Go unit (partial override) | MUST |
| AC-MMC-005 | REQ-MMC-005 | Go unit (validation error) | MUST |
| AC-MMC-006 | REQ-MMC-006 | Go unit (validation empty=ok) | MUST |
| AC-MMC-007 | REQ-MMC-007/008 | grep (no unconditional --squash) | MUST |
| AC-MMC-008 | REQ-MMC-009 | grep (squash-default byte-equiv) | MUST |
| AC-MMC-009 | REQ-MMC-012 | Go test (neutrality audit) | MUST |
| AC-MMC-010 | REQ-MMC-010/011 | grep + mirror test (FROZEN amend) | MUST |
| AC-MMC-011 | REQ-MMC-002 | grep (template key default) | MUST |
| AC-MMC-012 | cross-cutting | Go full suite no-regression | MUST |

## D.1 Given-When-Then scenarios

### AC-MMC-001 — Default merge_method is squash in all modes
- **Given** a freshly constructed `GitStrategyConfig` from `NewDefaultGitStrategyConfig()`
- **When** `Manual.MergeMethod`, `Personal.MergeMethod`, `Team.MergeMethod` are read
- **Then** each equals `"squash"`
- Verify: `go test ./internal/config/ -run TestNewDefaultGitStrategyConfig` (or extended defaults test) shows all 3 fields == `squash`.

### AC-MMC-002 — Loader retains default when merge_method absent
- **Given** a `git-strategy.yaml` that defines `team:` but omits `merge_method`
- **When** the loader runs `loadGitStrategySection`
- **Then** `cfg.GitStrategy.Team.MergeMethod == "squash"` (compiled default, not empty)
- Verify: extended `git_strategy_loader_test.go` (mirrors `TestLoader_GitStrategy_Absent_KeepsDefaultsAndFlagUnset` for the new field).

### AC-MMC-003 — Loader populates file value when present
- **Given** a `git-strategy.yaml` with `git_strategy.team.merge_method: merge`
- **When** the loader runs
- **Then** `cfg.GitStrategy.Team.MergeMethod == "merge"` (file value, not compiled default)
- Verify: extended `git_strategy_loader_test.go` (mirrors `TestLoader_GitStrategy_Present_LoadsValuesAndFlagsSection`).

### AC-MMC-004 — Partial override preserves sibling defaults
- **Given** a `git-strategy.yaml` setting `git_strategy.team.merge_method: rebase` but omitting it for `manual`/`personal`
- **When** the loader runs
- **Then** `Team.MergeMethod == "rebase"` AND `Manual.MergeMethod == "squash"` AND `Personal.MergeMethod == "squash"`
- Verify: extended `git_strategy_loader_test.go` partial-override test.

### AC-MMC-005 — Invalid merge_method fails validation with field path
- **Given** a config with `git_strategy.team.merge_method: rocket`
- **When** validation runs
- **Then** a validation error is returned naming `git_strategy.team.merge_method`
- Verify: validation test asserts error contains the field path and rejects the value.

### AC-MMC-006 — Empty merge_method passes validation (fail-safe)
- **Given** a config where `merge_method` is the empty string (no user value, no default applied in the test fixture)
- **When** validation runs
- **Then** no validation error is emitted for `merge_method` (treated as default `squash`)
- Verify: validation test with empty field asserts no error.

### AC-MMC-007 — No unconditional hardcoded --squash in sync prose
- **Given** the rewired `sync/delivery.md` and `manager-git.md`
- **When** scanning for hardcoded merge commands
- **Then** every `gh pr merge` occurrence is inside a config-driven/default formulation (no bare unconditional `--squash`)
- Verify:
  ```bash
  # Each gh pr merge line must reference the config-driven method, not a bare hardcoded --squash.
  # Manual inspection of context + grep that the method-selection prose references merge_method:
  grep -n 'merge_method' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
  grep -n 'merge_method' internal/template/templates/.claude/agents/moai/manager-git.md
  # Expected: ≥1 match each (the prose now reads the config value)
  ```

### AC-MMC-008 — squash-default rendered command is byte-equivalent
- **Given** `merge_method` resolves to `squash` (the default)
- **When** the agent renders the merge command
- **Then** the command is exactly `gh pr merge --squash --delete-branch` (no behavior change for non-opt-in users)
- Verify: the rewired prose explicitly states the squash case produces `gh pr merge --squash --delete-branch` verbatim.

### AC-MMC-009 — Template content remains neutral
- **Given** the edited template files under `internal/template/templates/`
- **When** the neutrality audit runs
- **Then** no SPEC ID / internal date / commit SHA / CLAUDE.local reference is present
- Verify: `go test ./internal/template/... -run TestTemplateNeutralityAudit` passes; `TestInternalContentLeak` (if present) passes.

### AC-MMC-010 — FROZEN lifecycle table amended in both SSOT and mirror
- **Given** `spec-workflow.md` (SSOT) and its template mirror
- **When** the lifecycle table is read
- **Then** the PR-strategy column reads "configured `merge_method` (default `squash`)" (or equivalent) for all 3 phase rows, the FROZEN rationale prose is retained, AND both copies are byte-identical
- Verify:
  ```bash
  go test ./internal/template/... -run TestEmbeddedMirror   # byte-identity of SSOT vs template mirror
  grep -n 'merge_method' .claude/rules/moai/workflow/spec-workflow.md
  grep -n 'merge_method' internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
  ```

### AC-MMC-011 — Template git-strategy.yaml ships merge_method default
- **Given** `git-strategy.yaml.tmpl`
- **When** scanning each mode profile
- **Then** each mode (manual/personal/team) has `merge_method: squash`
- Verify:
  ```bash
  grep -c 'merge_method: squash' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
  # Expected: 3
  ```

### AC-MMC-012 — Full suite no-regression
- **Given** the complete change set
- **When** `go test ./...` runs
- **Then** all tests pass; `internal/config` and `internal/template` coverage is not lower than baseline
- Verify: `go test ./...` exits 0; `golangci-lint run` clean.

## D.2 Edge cases (must be covered by tests)

- EC-1: `merge_method: ""` (empty) → resolves to `squash`, no validation error (AC-MMC-006).
- EC-2: `merge_method: rebase` with a `main_direct` workflow (no PR) → valid config, field unused, no error.
- EC-3: case sensitivity — `merge_method: Squash` (capital S) → validator decision: REJECT (enum is lowercase `{squash, merge, rebase}`). Documented so run-phase implements lowercase-only matching consistent with existing hook-severity enum handling.

## D.3 Definition of Done

- [ ] All 12 ACs pass (AC-MMC-001..012).
- [ ] All 3 edge cases (EC-1..3) covered by tests.
- [ ] `go test ./...` exits 0, no coverage regression in `internal/config` / `internal/template`.
- [ ] `golangci-lint run` clean.
- [ ] `make build` run (or `go:embed all:templates` confirmed to make it unnecessary) after template edits.
- [ ] `embedded_mirror_test.go` passes (spec-workflow.md SSOT/mirror byte-identity).
- [ ] Neutrality CI guard (`template-neutrality-check.yaml` / `TestTemplateNeutralityAudit`) passes.
- [ ] Default-squash behavior is byte-equivalent to pre-SPEC behavior (AC-MMC-008).
- [ ] FROZEN-zone edit (M6) was performed only after GATE-2 approval.
- [ ] Frontmatter `status` transitioned `draft → in-progress` (M1) → `in-progress → implemented` (sync) per ownership matrix.

## D.4 Quality gate criteria

- TRUST-Tested: loader/validation/defaults unit tests (AC-MMC-001..006), grep/structural assertions (AC-MMC-007..011).
- TRUST-Readable: new field godoc mirrors `HooksConfig` style.
- TRUST-Unified: follows existing `git_strategy.<mode>.hooks.*` field convention exactly.
- TRUST-Secured: enum validation prevents arbitrary shell flag injection via `merge_method`.
- TRUST-Trackable: conventional commits, Issue #1061 reference, REQ-AC traceability table above.

## D.5 Forward-looking checks (non-blocking, candidate follow-ups)

- The `internal/github` `PRMerger` default (`merge`) remains unreconciled with this SPEC's `squash` default; flagged for any future PRMerger-wiring SPEC (spec.md C.3, EX-1).
- Per-branch-type gitflow override (`release/* → merge`) remains a candidate follow-up (EX-2).
