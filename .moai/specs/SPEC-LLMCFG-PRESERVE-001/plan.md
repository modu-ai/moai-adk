---
id: SPEC-LLMCFG-PRESERVE-001
title: "Implementation plan — llm.yaml update-preservation contract pin"
version: "0.1.0"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec (card t239)
tier: M
---

# plan.md — SPEC-LLMCFG-PRESERVE-001

## §A Context

- **Card**: t239 (Factory Mode lane-15), branch `WT-llm-yaml-preserve`,
  plan authored at HEAD `b7462203a` (develop lineage).
- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t239` — all run
  phase work happens inside this tree.
- **Defect surface**: `.moai/config/sections/llm.yaml` (gitignored,
  runtime-expanded by `moai glm`/`cg`, user-editable) sits inside the
  `.moai/config` root that `CleanMoaiManagedPaths` deletes wholesale
  (`internal/cli/update/deploy/deploy.go:187-204`) before template redeploy.
- **Existing protection (verified, not to be rebuilt)**: Backup
  (`backup/backup.go:27`) → 3-way node merge (`backup/merge.go`,
  `backup/node_merge.go`; comments/order/quoting preserved per
  SPEC-UPDATE-YAML-PRESERVE-001) → Restore (`backup/restore.go`,
  wired at `update_template_sync.go:472-524` and
  `update_clean_install.go:481`).
- **The gap**: llm.yaml has NO named end-to-end preservation test on either
  update path. Sibling sections (user/language/design.yaml) have
  `TestCleanReinstall_AllSectionsYAMLPreserved`
  (`update_clean_install_config_preserve_test.go:195`); llm.yaml's only update
  test presence is a synthetic retained-advisory key
  (`update_retained_advisory_test.go:33`) plus synthetic-fixture merge units.
- **Design decision (spec.md §D.1)**: keep + pin. No new preservation
  machinery; tests + decision record; repair only what a failing contract
  names.

### §A.5 PRESERVE list (do not modify)

- `internal/cli/update/backup/**` — merge engine and restore logic: exercised
  as-is; edit ONLY if a new contract fails and names a defect.
- `internal/cli/update/merge/**` — gitignore/3-way file merge: read-only.
- `internal/cli/update/deploy/deploy.go` — the wipe is the constraint, not a
  defect to re-engineer here.
- `internal/template/templates/.moai/config/sections/llm.yaml` — template
  content unchanged (no `make build` needed unless a repair touches templates,
  which plan does not anticipate).
- `internal/cli/hook_install_precommit.go` + hook tier files — t230 surface.
- `.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/**` — provenance SPEC, read-only.
- `internal/settings/**` — runtime writer surface, out of scope.
- `.moai/lessons-inbox.jsonl` and all t280-related paths — out of scope.

## §B Known Issues (relevant subset)

- **B5 (CI 3-tier awareness)**: new tests land in `internal/cli` (heavy
  package). CI runs the full matrix; local runs MUST stay scoped (`-run
  TestUpdateLLMYAML…`, targeted packages). Baseline lint status is captured in
  §C so NEW vs pre-existing findings are separable.
- **B8 (working-tree hygiene)**: stage by explicit pathspec; never
  `git add -A`. Test temp dirs under `t.TempDir()` only.
- **B10 (scope discipline)**: §A.5 PRESERVE list binds. The tempting drive-by —
  "fix the stale `excludedDirs` comment in backup.go while reading it" — is
  NOT in scope.
- **B11 (subagent boundary)**: manager-develop returns blocker reports; no
  AskUserQuestion.
- **SPEC-specific K1**: `update_retained_advisory_test.go:33` already uses
  `llm.agent_overrides` as an advisory-render fixture — the new tests must not
  collide with its fixture shapes or package-level sink swaps
  (`SetRetainedKeySinkForTest` users must restore the sink).
- **SPEC-specific K2**: the merge base for a first-update fixture is
  `SaveTemplateDefaults` (current embedded defaults). For NEW-template-key
  assertions this is harmless (new key present in base or not, delivery holds);
  do not assert default-CHANGE adoption semantics in these fixtures — that is
  the excluded imperfect-base residual (spec.md §F).

## §C Pre-flight

```bash
git -C <worktree> rev-parse --short HEAD && git -C <worktree> branch --show-current
# expect: b7462203a (or the merged develop tip) / WT-llm-yaml-preserve

# Scoped baseline: the packages under contract (measured green at plan time)
go test -count=1 ./internal/cli/update/backup/ ./internal/cli/update/merge/

# Lint baseline (NEW vs pre-existing separation)
golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -5

# Confirm the fixture surface exists
grep -n "llm.yaml" internal/template/templates/.moai/config/sections/ -r
ls internal/cli/update_clean_install_config_preserve_test.go
```

Baseline attribution: plan-phase measured `go test -count=1
./internal/cli/update/backup/` → `ok github.com/modu-ai/moai-adk/internal/cli/
update/backup 0.626s` on HEAD `b7462203a` (this tree, this run).

## §D Constraints

- Test files only, unless a contract fails and names a production defect
  (then: minimal repair, noted in the milestone's evidence).
- New test file(s): `internal/cli/update_llm_preserve_test.go` (template-sync
  path) and an llm.yaml extension inside the existing
  `update_clean_install_config_preserve_test.go` fixture family (clean-reinstall
  parity) — or a sibling file reusing its helpers, if the fixture helper is
  file-local.
- Assertions: parsed key values (`yamlNestedString`-style), comment presence
  (`strings.Contains` on template comment sentinels), key delivery
  (template-only key present post-update). NO byte-equality on merged output.
- Every new contract gets a mutant teeth demonstration recorded in evidence
  (e.g. temporarily bypass `RestoreMoaiConfigRetained` → test RED → restore).
  The mutant run is scratch (not committed).
- Conventional Commits; card id `t239` + SPEC id in every commit subject.
- No local full-suite runs; no push from the lane (develop integration is the
  lead's window).

## §E Self-Verification

Run-phase reports per verification-claim-integrity §3 (Claim / Evidence /
Baseline-attribution / Gaps / Residual-risk), each item naming command,
verbatim output, and HEAD SHA:

- **E1** AC matrix — AC-LCP-001..006 PASS/FAIL with the exact commands from
  acceptance.md.
- **E2** Scoped packages green: `go test -count=1
  ./internal/cli/update/backup/ ./internal/cli/update/merge/` and the
  `-run TestUpdateLLMYAML` filter over `./internal/cli/`.
- **E3** Mutant teeth evidence for each new contract (RED observed under
  mutation, then GREEN restored).
- **E4** `go vet ./internal/cli/...` clean on NEW code; lint delta vs §C
  baseline = 0 NEW findings.
- **E5** Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0 and
  `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` (test files compile).
- **E6** Commit SHAs + PRESERVE-list adherence statement (§A.5 files untouched
  — `git status --short` scoped check).

## §F Milestones

Order: contract value first (decision-reversibility: the contract tests are the
deliverable most likely to force design conversation; the mechanical extension
follows; evidence last).

- **M1 (Priority High) — template-sync path contract**
  Author `internal/cli/update_llm_preserve_test.go`: fixture = `t.TempDir()`
  project with a user-edited llm.yaml (a changed `glm.models.high` value, an
  `agent_overrides` entry, a marker comment line near the template's comment
  block), real embedded template deployed via the update cycle
  (`runTemplateSyncWithReporter` seam or the step functions the existing
  characterization tests use). Assert: user values survive (parsed), marker
  comment survives, template-only key delivered (assert against a key the
  CURRENT template has that the fixture user file omits), fresh-install case
  deploys the default verbatim. Mutant: bypass restore → RED → restore → GREEN.
  Deliverable: AC-LCP-001..004 evidence.
- **M2 (Priority High) — clean-reinstall parity**
  Extend the `update_clean_install_config_preserve_test.go` fixture family
  (or a sibling file reusing its helpers) with llm.yaml: same user-edit
  pattern survives the clean-reinstall clobber + restore, matching
  AC-RIL-006's protection class. Deliverable: AC-LCP-005 evidence.
- **M3 (Priority Medium) — negative constraint + verification batch**
  No-sidecar regression guard (grep form in acceptance.md AC-LCP-006), scoped
  package runs, vet, cross-platform compile/vet, PRESERVE adherence check,
  §E evidence assembly. Deliverable: AC-LCP-006 + E1-E6.

## §G Anti-Patterns

- Do NOT "simplify" by excluding llm.yaml from deployment or relocating it —
  the design decision rejected both (spec.md §D.1); re-litigating in run phase
  without new evidence is scope drift.
- Do NOT add a provenance sidecar "for symmetry" with hooks (REQ-LCP-005).
- Do NOT broaden to t280 or workflow.yaml (spec.md §F).
- Do NOT write byte-snapshot assertions against merged YAML — they break on
  legitimate comment/ordering-preserving re-encodes and teach the suite to
  fear the engine.
- Do NOT assert default-change adoption semantics in these fixtures (K2) —
  that is the excluded imperfect-base residual, and a fixture asserting it
  would flake against future template default edits.

## §H Cross-References

- spec.md §A (verified mechanics), §D (GEARS + design decision), §F (Out of
  Scope).
- acceptance.md (AC-LCP-001..006 + verification strategy).
- `update_clean_install_config_preserve_test.go` — the fixture pattern M2 joins.
- `internal/cli/update_template_sync.go:398-524` — the pipeline steps M1 drives.
- SPEC-UPDATE-YAML-PRESERVE-001, SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001,
  SPEC-V3R6-AUDIT-MODEL-PIN-001 — provenance chain for the mechanism, the
  base, and the card-premise correction.
