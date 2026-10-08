---
id: SPEC-CUTOVER-RESIDUE-001
title: "Implementation plan — sweep base fallback + cutover residue cleanup"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
author: t1564 card worker
priority: P1
module: "internal/cli/worktree"
tier: M
---

# SPEC-CUTOVER-RESIDUE-001 — Plan

## §A Milestones

### M1 — Sweep base fallback (RED→GREEN, tdd)

Files: `internal/cli/worktree/sweep.go` (modify), `internal/cli/worktree/sweep_base_fallback_test.go` (new).

1. **RED (live)** — with the pre-fix binary built from this tree: `./bin/moai worktree sweep --json 2>/dev/null | grep -c 'cause=fetch-failed'` against this repo (tracked config still git-flow/develop) → expect a large non-zero count (187 measured at tick 3).
2. **RED (unit)** — author `sweep_base_fallback_test.go` with seam stubs (convention: save orig var, `t.Cleanup` restore):
   - derived base `origin/develop`, probe says absent, HEAD probe says `main` → expect `origin/main` (currently FAILS — `sweepEffectiveBase` does not exist);
   - probe says present → base unchanged, symref not consulted;
   - probe error → base unchanged;
   - HEAD resolution error → base unchanged.
3. **GREEN** — implement in `sweep.go`:
   - `var sweepRemoteRefExists = func(repoRoot, remote, ref string) (bool, error)` — `git ls-remote --exit-code <remote> refs/heads/<ref>`; err==nil → true, exit 2 → false, else error.
   - `var sweepRemoteHead = func(repoRoot, remote string) (string, error)` — `git ls-remote --symref <remote> HEAD`, parse `ref: refs/heads/<name>` first line, trim after TAB/space.
   - `func sweepEffectiveBase(repoRoot, base string) (string, bool)` — split at first `/` (sweepFetchBase convention); absent ref + resolvable HEAD → `remote + "/" + head`, true; every other path → `base, false`. Guard: if resolved head equals the absent ref name, return base unchanged (honest failure).
   - Wire in `runSweep` only on the derived path (before `classifySweepVerdicts`); on fallback print one stderr line: `base <derived> absent on <remote>; falling back to the remote default branch <effective>`. Explicit `--base` skips both the derivation and the fallback.
4. Verify: `go test ./internal/cli/worktree/... -run '^TestSweep.*$' -count=1` green; GREEN (live): re-run the sweep command from step 1 → fetch-failed count collapses to 0 (fallback notice present), landings evaluated against `origin/main`.

### M2 — Config + doctrine + template residue

Files: `.moai/config/sections/git-strategy.yaml` (tracked, modify), `.moai/docs/update-local-file-survival.md` (modify), `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` (comment only).

1. (Revised at run-phase — the original github-flow flip collided with the `TestGitHubFlowSweepGuard` arming contract, see spec.md §A amendment.) Keep the tracked local config's `manual.workflow: git-flow` and its git-flow-era keys (the guard's arming switch + `develop_branch` as the IntegrationTarget source until the residue-sweep card flips the model); fix ONLY the `lead_push_threshold` counting comment to name no flow-specific branch, and keep `worktree_base_branch: main`.
2. Survival doc §2.3: supersede the `--source=develop` block — re-apply source is `main`'s committed copy (`git restore --source=main -- .moai/config/sections/git-strategy.yaml`), verification greps cover the operator keys (`worktree_base_branch: main` present, workflow value intact); replace the §0.1/§0.3 develop-era references in the re-apply context per the cutover doctrine (AGENTS.local.md §4.1 is the norm already — align the doc to it).
3. Template comment: `origin/<develop>..<develop>` → generic integration-branch phrasing. Then Template-First cycle: `make agents-emit && make build` (embed manifest + catalog hashes regenerate for the `.tmpl` edit).
4. Verify: `go test ./internal/config/... ./internal/template/... -count=1`; grep the config's threshold comment for `develop` → 0; grep the survival doc for the superseded instruction → replaced.

### M3 — Workflow branch filters

Files: 8 `branches: [main, develop]` edits + docs-i18n-check.yml trigger/comment.

1. `branches: [main, develop]` → `branches: [main]` in: test-install.yml, codeql.yml, ci.yml, graph-freshness.yml (×3 blocks), judgment-first-consistency.yaml, workflow-parse-guard.yaml, lsel-leak-guard.yaml, template-neutrality-check.yaml.
2. docs-i18n-check.yml: drop the `- develop` trigger entry; rewrite the git-flow-era comment block (lines ~27-33) to the cutover answer (PR triggers on main); keep line 72's phase-note reworded if it names develop as a push target.
3. Historical run-ID comments in ci.yml / release-pr-multi-os.yml stay (out of scope, spec.md §E.1).

### M4 — Verify, push, PR

1. `go vet` + `golangci-lint run` on changed packages; affected-package tests only (`./internal/cli/worktree/... ./internal/config/... ./internal/template/... -timeout 30m`) — no local full suite.
2. GREEN (live, post-fix binary): sweep `--json` fetch-failed = 0; verdict reasons name `origin/main`.
3. Conventional Commits, each carrying `Card: t1564`; push `WT-cutover-residue`; PR base `main`, title prefixed `t1564`. No merge (leader's).

## §B PRESERVE

- Primary checkout: read-only throughout (the deployed git-strategy.yaml there is runtime state, not ours).
- `internal/config/loader_integration_branch.go`, `loader_workflow_disposition.go` semantics — no code change.
- spec-lint.yml (already self-fixed), release-pr-multi-os.yml / ci.yml historical comments.
- Other lanes' worktrees and the shared primary tree.

## §C Risks

- `ls-remote` adds a network round trip to every sweep default run — acceptable: the sweep already fetches (network) once per run; the probe rides the same connectivity assumption, and its failure path preserves current behavior.
- The config flip changes what `moai integration acquire` / doctor resolve (target develop → main) — intended: that IS the cutover consistency fix; the leader's window flow reads main post-cutover.
- Template edit churns `embed_manifest_gen.go`/`catalog.yaml` — expected Template-First artifacts, committed together.
