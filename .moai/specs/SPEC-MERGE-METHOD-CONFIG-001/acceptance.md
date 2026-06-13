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
| AC-MMC-010 | REQ-MMC-010/011 | grep + TestRuleTemplateMirrorDrift (FROZEN amend) | MUST |
| AC-MMC-011 | REQ-MMC-002 | grep (template key default) | MUST |
| AC-MMC-012 | cross-cutting | Go full suite no-regression | MUST |
| AC-MMC-013 | REQ-MMC-007/012 | grep (ref-git-workflow skill.md config-driven) | MUST |

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
- Verify (two-part — [STRENGTHENED per plan-audit D5]):
  ```bash
  # PART 1 (positive): the prose now reads the config value — merge_method is referenced.
  grep -n 'merge_method' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
  grep -n 'merge_method' internal/template/templates/.claude/agents/moai/manager-git.md
  # Expected: ≥1 match each (the prose now reads the config value)

  # PART 2 (negative — this is the part D5 added, and is the load-bearing assertion):
  # Prove no BARE unconditional `gh pr merge --squash` literal survives. Every surviving
  # `gh pr merge --squash` token (if any) MUST be inside an explicit default/conditional
  # formulation (e.g. "when merge_method=squash (default) → gh pr merge --squash ...").
  # Pre-SPEC baseline: 5 bare sites (delivery.md:313,325 + manager-git.md:140,176,204).
  grep -rn 'gh pr merge --squash' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md internal/template/templates/.claude/agents/moai/manager-git.md
  # For EACH line returned, the manual-inspection criterion is: the line is part of a
  # method-selection block that names `merge_method` within the same step/paragraph
  # (i.e. it is the documented squash-DEFAULT branch, not a hardcoded unconditional command).
  # FAIL if any returned line is a bare imperative "Execute `gh pr merge --squash --delete-branch`"
  # with no surrounding merge_method-driven conditional.
  ```
- Manual-inspection checklist (the 5 baseline sites, each must transition from bare → conditional):
  - `sync/delivery.md:313` (Step 3.4 first auto-merge) → conditional on `merge_method`
  - `sync/delivery.md:325` (Step 3.4 second auto-merge) → conditional on `merge_method`
  - `manager-git.md:140` (merge command block) → conditional on `merge_method`
  - `manager-git.md:176` (auto-merge with `--auto-merge` flag) → conditional on `merge_method`
  - `manager-git.md:204` (numbered merge step 4) → conditional on `merge_method`
- Note: line numbers are pre-SPEC baseline anchors for the implementer; the rewiring will shift them. The binary assertion is "zero bare unconditional `gh pr merge --squash` imperatives remain; every squash command is the documented default branch of a `merge_method` selection."

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
  # [CORRECTED per plan-audit D7] real mirror-drift test is TestRuleTemplateMirrorDrift
  # (internal/template/rule_template_mirror_test.go:99), NOT the nonexistent TestEmbeddedMirror.
  go test ./internal/template/... -run TestRuleTemplateMirrorDrift   # byte-identity of SSOT vs template mirror
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

### AC-MMC-013 — ref-git-workflow skill.md references the config-driven method [ADDED per plan-audit D4]
- **Given** `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md` (in-scope per spec.md §A.4 + plan.md M5; this is the file whose squash/merge/rebase guidance table — `:133-135` baseline — the issue cites as contradicted by the unconditional squash)
- **When** the squash/merge/rebase guidance section is read after M5 rewiring
- **Then** the guidance no longer implies the method is hardcoded — it explicitly states (or links) that the active method is governed by `git_strategy.<mode>.merge_method` config (default `squash`), so the doc and the runtime behavior agree
- Verify:
  ```bash
  # PART 1 (positive): skill.md now references the config field.
  grep -n 'merge_method' internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md
  # Expected: ≥1 match (the guidance points to the configurable field)

  # PART 2 (neutrality — skill.md is template content): no internal SPEC ID / date / SHA leaked.
  go test ./internal/template/... -run TestTemplateNeutralityAudit
  ```
- Note: the existing `:133-135` rows (Squash merge / Merge commit / Rebase) already document all 3 methods correctly; M5's job is to add the pointer that selection is config-driven, NOT to rewrite the rows. Resolves the D4 traceability gap (file was in-scope with no AC).

## D.2 Edge cases (must be covered by tests)

- EC-1: `merge_method: ""` (empty) → resolves to `squash`, no validation error (AC-MMC-006).
- EC-2: `merge_method: rebase` with a `main_direct` workflow (no PR) → valid config, field unused, no error.
- EC-3: case sensitivity — `merge_method: Squash` (capital S) → validator decision: REJECT (enum is lowercase `{squash, merge, rebase}`). Documented so run-phase implements lowercase-only matching consistent with existing hook-severity enum handling.

## D.3 Definition of Done

- [ ] All 13 ACs pass (AC-MMC-001..013).
- [ ] All 3 edge cases (EC-1..3) covered by tests.
- [ ] `go test ./...` exits 0, no coverage regression in `internal/config` / `internal/template`.
- [ ] `golangci-lint run` clean.
- [ ] `make build` run (or `go:embed all:templates` confirmed to make it unnecessary) after template edits.
- [ ] `TestRuleTemplateMirrorDrift` passes (spec-workflow.md SSOT/mirror byte-identity — corrected from the nonexistent `TestEmbeddedMirror` per plan-audit D7).
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
