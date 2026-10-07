# SPEC-CUTOVER-RESIDUE-001 — Progress

Card: t1564 (factory run tmhxo0, self-tree mode). Branch: `WT-cutover-residue` (cut from `origin/main` `cb2a011d0`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (self-tree mode — the dispatch authorizes in-lane-style plan→run→sync without a separate plan-auditor round; the plan-phase evidence below is the audit-cross input the leader reads)

plan_complete_at: 2026-10-07

### Plan-phase measurements (this run, tree `cb2a011d0`)

- Remote probe: `git ls-remote --symref origin HEAD` → `ref: refs/heads/main	HEAD` (exit 0); `git ls-remote --exit-code origin refs/heads/develop` → exit 2 (absent); `git ls-remote --exit-code origin refs/heads/main` → exit 0.
- Tracked config residue: `.moai/config/sections/git-strategy.yaml` carries `manual.workflow: git-flow`, `develop_branch: develop`, `worktree_base_branch: main` (mixed state), threshold comment naming `origin/develop..develop`.
- Survival doc residue: §2.3 instructs `git restore --source=develop` and lists `worktree_base_branch: develop` as a key to re-apply.
- Workflow residue: `branches: [main, develop]` in 8 files (10 filter lines) + docs-i18n-check.yml develop trigger + git-flow-era comment block.
- Template residue: `git-strategy.yaml.tmpl` line ~25 comment names `origin/<develop>..<develop>`.

## Run-phase notes

(appended by M1–M4 below)

### M1 — sweep base fallback (done)

- Unit RED observed (compile failure, `undefined: sweepEffectiveBase` + seams) before implementation; GREEN: `go test ./internal/cli/worktree/ -run Sweep -count=1` → ok (123.667s).
- Live RED→GREEN pair on the same tree: pre-fix binary 185/187 `cause=fetch-failed`; post-fix binary 0 fetch-failed (69 DISPOSE / 118 PRESERVE, preview only), stderr fallback notice naming both bases. Explicit `--base origin/develop` keeps 185 fetch-failed (honest failure, no silent switch).

### M2 — config/doctrine/template (done, one run-phase amendment)

- **Design-conflict discovery**: the plan's github-flow flip armed `TestGitHubFlowSweepGuard` (SPEC-GITHUB-FLOW-DEFAULT-001 M4 — arms only on `workflow == github-flow`) and its armed live-text assertions went red (`--- FAIL: TestGitHubFlowSweepGuard` in `internal/template`, observed on the flipped tree). The value's flip is that card's own M5 act, sequenced WITH its residue sweep. Resolution: REQ-CR-005 revised (spec §A amendment); the local config keeps `workflow: git-flow` + git-flow-era keys, gains only the branch-neutral threshold comment; `worktree_base_branch: main` unchanged. The sweep is value-independent via the M1 fallback (already measured GREEN with the git-flow config).
- Survival doc §2.3 re-keyed to `--source=main` with the two verification greps; §0.1 provenance line updated to the main tree.
- Template `lead_push_threshold` comment neutralized; `make build` green.

### M3 — workflow filters (done)

- 8 workflows `branches: [main, develop]` → `[main]`; docs-i18n-check push trigger + comment aligned. `grep -rn 'branches:.*develop' .github/workflows/` → 0. All 9 edited YAMLs parse.

### M4 — verify, push, PR (done)

- Build unblock: main-tip embed manifest defect repaired (`make embed-manifest`) — measured, committed as `a54c6ab60`.
- vet 0 · lint `0 issues.` (worktree, config, template) · affected suites green · spec-lint baseline gate exit 0 · guard PASS disarmed after the revert.
- Final live sweep: 71 DISPOSE / 120 PRESERVE, 0 fetch-failed, fallback notice (preview only).
- Commits `a54c6ab60` `642a662d9` `723b0272b` `ed7b7c6e4` `8302697f7` on `WT-cutover-residue`, pushed; **PR #1786** (base main). Card-review run (raw fail with base anomaly — advisory disposition, `card-review.md`).

## §E.2 Run-phase Evidence

- RED: unit compile-failure output captured pre-implementation (`undefined: sweepEffectiveBase`); live pre-fix sweep `--json` → 185/187 `cause=fetch-failed; fetch origin develop: exit status 128` (evidence: `.moai/reports/t1564/t1564-sweep-red.json`).
- GREEN: post-fix `--json` → 0 fetch-failed (69/118 then 71/120 across two runs), stderr fallback notice; explicit `--base origin/develop` → 185 fetch-failed preserved (evidence: `t1564-sweep-green.json`, `t1564-sweep-final.json`).
- AC matrix + baselines: `.moai/reports/t1564/verdict.md` (AC roll-up PASS ×8, one criterion revised per the §A amendment).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready (self-tree mode; all 8 ACs pass, RED evidence recorded, card-review filed with attribution analysis)

run_complete_at: 2026-10-07


