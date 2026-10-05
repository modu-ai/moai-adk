---
id: SPEC-LLMCFG-PRESERVE-001
title: "Acceptance criteria — llm.yaml update-preservation contract"
version: "0.1.0"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec (card t239)
tier: M
---

# acceptance.md — SPEC-LLMCFG-PRESERVE-001

## §A Verification Strategy

Classification: **regression-guard contracts**. They pin behavior that exists
at plan time (verified: scoped backup-package tests green at HEAD
`b7462203a`); their RED-now cell is therefore a **mutant**, not the pre-work
tree. Two-cell adoption per verification-completeness §2:

- **GREEN path**: the named command passes on the post-M1/M2 tree.
- **RED (teeth) cell**: the named mutation makes the command FAIL, observed
  once, in scratch (not committed). A contract whose teeth were never
  observed is unadopted — it is green output of unknown sensitivity.
- Every criterion names its command verbatim; runs are scoped (`-run` filters,
  targeted packages) per the lane load discipline. Baseline attribution:
  HEAD SHA + verbatim output recorded in progress.md §E.2.

## §B AC Matrix

### AC-LCP-001 — user values survive the template-sync update cycle (e2e)

- **Given** a `t.TempDir()` project whose deployed llm.yaml carries user
  edits — `llm.glm.models.high` changed from the template default, an
  `llm.agent_overrides` entry the template default does not contain, and a
  marker comment line inserted under the template's comment block —
- **When** the template-sync update cycle runs to completion (Backup →
  Clean Managed Paths → Deploy Templates → Restore Settings) against the REAL
  embedded template,
- **Then** the post-update llm.yaml parses with the user's `glm.models.high`
  value (not the template default), contains the user's `agent_overrides`
  entry, and contains the marker comment.
- **Command**: `go test -run TestUpdateLLMYAMLPreserveTemplateSync ./internal/cli/ -count=1`
- **Mutant (teeth)**: bypass the restore step (stub `RestoreMoaiConfigRetained`
  to a no-op in a scratch edit) → the command FAILS on all three assertions →
  restore → GREEN.

### AC-LCP-002 — template-gained keys reach a preserved file (e2e)

- **Given** the same fixture class as AC-LCP-001, with the fixture user file
  derived from the real template with ONE template-carried key removed (e.g.
  the `llm.performance_tier` line),
- **When** the update cycle runs,
- **Then** the removed key is present again post-update with the template
  default value (delivery), AND the user's divergent values from AC-LCP-001's
  pattern still hold (preservation in the same pass — delivery must not cost
  preservation).
- **Command**: `go test -run TestUpdateLLMYAMLNewKeyDelivery ./internal/cli/ -count=1`
- **Mutant (teeth)**: at the merge-function level, feed `DeepMerge3Way` a
  "new" document missing the key in a scratch unit probe → delivery assertion
  FAILS (proves the e2e assertion is sensitive to the delivery path, not
  vacuously green because the key was never removed).

### AC-LCP-003 — first deploy stays calm for untouched installs (e2e)

- **Given** a `t.TempDir()` project with NO llm.yaml on disk,
- **When** the update cycle runs,
- **Then** llm.yaml exists post-update and is byte-identical to the embedded
  template llm.yaml (byte equality IS the assertion here — nothing was merged,
  so no brittleness clause applies), and the update output reports no
  preservation advisories for llm.yaml.
- **Command**: `go test -run TestUpdateLLMYAMLFirstDeployCalm ./internal/cli/ -count=1`
- **Mutant (teeth)**: scratch-edit the template deployment list to skip
  llm.yaml → the command FAILS (file absent) → restore → GREEN.

### AC-LCP-004 — comment documentation survives the merge (real-template round trip)

- **Given** the AC-LCP-001 fixture,
- **When** the update cycle runs,
- **Then** the merged llm.yaml retains the template's comment documentation:
  at minimum the comment-block sentinels present in the shipped template
  (e.g. the `# Profile matrix` block header and the GLM reasoning-effort
  collapse comment) — issue #1243's class, asserted against REAL template
  bytes rather than a synthetic fixture.
- **Command**: `go test -run TestUpdateLLMYAMLCommentsSurvive ./internal/cli/ -count=1`
  (may live inside the AC-LCP-001 test function as additional assertions; the
  command names the selector that runs them).
- **Mutant (teeth)**: at the merge-function level, run `MergeYAML3Way` on the
  real template bytes through a `map[string]any` round trip in a scratch probe
  → the comment assertions FAIL (the pre-#1243 defect class reproduces),
  proving the assertions detect comment loss.

### AC-LCP-005 — clean-reinstall path provides the same protection class

- **Given** the `update_clean_install_config_preserve_test.go` fixture family
  extended with the same llm.yaml user-edit pattern (value change +
  `agent_overrides` entry + marker comment),
- **When** the clean-reinstall flow runs (force-deploy clobber + restore),
- **Then** the user values and marker comment survive — llm.yaml joins the
  protection class AC-RIL-006 already grants user/language/design.yaml.
- **Command**: `go test -run TestCleanReinstallLLMYAMLPreserved ./internal/cli/ -count=1`
- **Mutant (teeth)**: bypass the clean-reinstall restore call in scratch →
  FAIL → restore → GREEN.

### AC-LCP-006 — no provenance sidecar enters the update path (negative guard)

- **Given** the update-path source tree (`internal/cli/update/**`, excluding
  `_test.go`),
- **When** scanned for a pre-commit-style provenance sidecar name pattern,
- **Then** zero matches exist (the pattern `\.moai-[a-z-]+\.sha256` appears
  only under the hook tier, `internal/cli/hook_install*`).
- **Command**: `grep -rnE "\.moai-[a-z-]+\.sha256" internal/cli/update/ --include="*.go" | grep -v _test | wc -l` → `0`
  and `grep -rlnE "\.moai-[a-z-]+\.sha256" internal/cli/ --include="*.go" | grep -v _test` → hook tier files only.
- **Teeth note**: this guard's failure mode is a future addition; its teeth are
  demonstrated by the pattern MATCHING the hook tier's own constant
  (`moaiPreCommitProvenanceName = ".moai-pre-commit.sha256"`,
  `internal/cli/hook_install_precommit.go:19`) — the same scan that returns 0
  in scope returns ≥1 at the hook tier, proving the pattern is not vacuous.

## §C Quality gates

- Scoped packages green: `go test -count=1 ./internal/cli/update/backup/
  ./internal/cli/update/merge/`.
- `go vet ./internal/cli/...` clean on new code; lint delta vs plan §C
  baseline: 0 NEW findings.
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` (test files compile
  cross-platform).
- Coverage: the new test functions cover the fixture→cycle→assert path they
  own; no production-code coverage regression is expected (test-only change).
- PRESERVE list adherence: `git status --short` over §A.5 paths shows no
  modifications (evidence in progress.md §E.2).

## §D Definition of Done

- AC-LCP-001..006 all PASS with recorded commands + verbatim outputs + teeth
  evidence (or an explicitly logged SKIP per the absent-credential SKIP rule —
  not applicable here; no external backends are involved).
- No production code changed unless a contract failed and named a defect; any
  such repair is named in progress.md §E.2 with its motivating RED.
- Spec/plan/acceptance consistency: AC ↔ REQ mapping — AC-LCP-001/004 →
  REQ-LCP-002 + REQ-LCP-006; AC-LCP-002 → REQ-LCP-003; AC-LCP-003 →
  REQ-LCP-004; AC-LCP-005 → REQ-LCP-006; AC-LCP-006 → REQ-LCP-001/005.
- Commit subjects carry `t239` + `SPEC-LLMCFG-PRESERVE-001`.
