# SPEC-CLIFIX-HYGIENE-001 — Design

## §A update.go Decomposition Seam Map (v0.2.0 — residual scope)

Basis: live inventory against the current tree (2026-07-30, worktree HEAD = origin/main). `internal/cli/update.go` is 1,905 lines with 8 already-extracted sibling files in the cluster. The bulk decomposition named in the v0.1.0 audit (split into update_sync/update_merge/update_wizard/update_settings seams) is largely DONE; the merge machinery has been extracted further into a dedicated `internal/merge/` package (see design note below).

| Already-extracted sibling (current tree) | Concern |
|---|---|
| `update_archive.go` | archive extraction |
| `update_clean_install.go` | clean-install path |
| `update_cleanup.go` | deprecated-path scan/cleanup + the P0-wired lock pair (`acquireUpdateLock`, `cleanStaleLock`) |
| `update_deny_migration.go` | deny-list migration |
| `update_namespace_protect.go` | namespace-preservation logic |
| `update_preserve_inventory.go` | user-owned preserve inventory |
| `update_tux.go` | TUX3 path-classifier integration (`update.Classify` with shared `isUserOwnedNamespace` predicate — see update.go:1131-1150) |
| `update_noise.go` | noise/deprecation telemetry |

Residual decomposition target: bring `update.go` itself (1,905 lines) under the 1,200-line per-file ceiling. The residual over-ceiling portion (~705 lines) is the only split work remaining; run-phase re-derives the live function inventory (`grep -n '^func ' internal/cli/update.go`) at execution time to choose the seam, since `update.go` is still touched by adjacent work.

Move rules:

- One seam per commit, mechanical move only (no logic edits in the move commit — plan.md §D).
- Characterization suite runs after every move; per-file ceiling 1,200 lines (AC-HYG-001-001).
- **D11 resolution — path-classifier table extraction is a SEPARATE commit inside M5.** If the residual split reveals that the shared path-classifier prefix table (the v0.1.0 audit's cluster-1 Major row 8) still needs extraction into `update_sync.go` or a sibling, that extraction is a LOGIC refactor (not a mechanical move) and MUST land as its own commit within M5, distinct from any mechanical-move commit. This resolves the tension between this section (originally implying a logic refactor) and plan.md §D (mechanical-move-only). The path-classifier integration already partly lives in `update_tux.go` via `update.Classify` + the `isUserOwnedNamespace` predicate (update.go:1131-1150), so the residual extraction scope may be smaller than the v0.1.0 audit estimated — re-derive at run time.
- Sequencing note: this SPEC runs last (P0→P4); the P0 lock wiring (`runUpdate` ← `acquireUpdateLock` at update.go:264) and the F1 security redesign (no-token-persist at update.go:1375-1438) are already merged before any further split starts.

### Design note — merge-package extraction (DONE, out of scope)

The v0.1.0 design carried a `update_merge.go` seam for `deepMerge3Way` and related functions. That work is overtaken: the merge machinery now lives in `internal/merge/` (`strategies.go`, `three_way.go`, `differ.go`, `conflict.go`, `evolvable_zone.go`, `types.go`). The `deepMergeMap` function at `internal/merge/strategies.go:360` explicitly classifies keys against the 3-way inputs and preserves user-added keys (line 388: `// User added key - preserve.`). The v0.1.0 REQ-HYG-001-007 / AC-HYG-001-007 targeting this defect is DROPPED — see spec.md HISTORY v0.2.0.

### Design note — token persistence (DONE, out of scope)

The v0.1.0 design carried an `update_settings.go` concern that included token-file permissions. That work is overtaken by the F1 security redesign: `internal/cli/update.go:1375-1438` intentionally does NOT persist `github_token`/`gitlab_token` to `user.yaml` (credentials are delegated to the `gh`/`glab` CLI store). The v0.1.0 REQ-HYG-001-008 / AC-HYG-001-008 targeting file-mode 0600 is DROPPED — see spec.md HISTORY v0.2.0.

## §B PAT-Mask Render-Site Design (REQ-HYG-001-007)

The token questions (`github_token`, `gitlab_token`) flow through the wizard's generic `QuestionTypeInput` render dispatch. The render site is `internal/cli/wizard/wizard.go:329`, which constructs `huh.NewInput()` without any echo masking. The fix gates password-echo mode on the question ID:

- Token-question IDs to gate: `github_token` (`questions.go:225`) and `gitlab_token` (`questions.go:251`).
- Render-site change: at `wizard.go:329`, when `q.ID` is one of the token IDs, chain `.EchoMode(tty.EchoPassword)` (or the `huh`-equivalent password-echo API) onto the `huh.NewInput()` builder.
- Non-interactive path: the wizard's `-p`/CI flow skips interactive prompts entirely, so the gating does not hang headless runs.
- The wizard question descriptions already advertise "A token pasted here is NOT saved to any file" — that contract is preserved; the fix only masks the echo, it does not change persistence (which remains none).

(The v0.1.0 §B deepMerge3Way Key-Retirement Design section is removed — that defect is already fixed in `internal/merge/`. See design note above.)

## §C Cross-References

- research.md — dead-code caller-graph inventory feeding REQ-HYG-001-002 (M5 deletion milestone input).
- acceptance.md AC-HYG-001-001 / AC-HYG-001-007 / AC-HYG-001-008 — the machine checks over this design.
- Audit SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 cluster 1 (update.go structure rows), §4 row 1 (closed-struct round-trip family).
- v0.2.0 rescope driver: `.moai/reports/plan-audit/SPEC-CLIFIX-HYGIENE-001-review-1.md` (iter-1 FAIL 0.58, D1-D11).
