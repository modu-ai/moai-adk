# SPEC-CLIFIX-HYGIENE-001 — Implementation Plan

## §A Context

P4 row of the CLI audit roadmap: reduce maintenance cost and recurrence risk. Tier L — widest file surface of the five CLIFIX SPECs, but lowest urgency; strictly last in the series so it rebases on all prior fixes. Methodology: DDD (ANALYZE-PRESERVE-IMPROVE) for the decomposition and dead-code milestones; Reproduction-First TDD for the genuine behavior defects (PAT echo, YAML misparse, rune truncation).

v0.2.0 rescope (2026-07-30): two originally-scoped defects were found already-satisfied during re-verification against the post-P0-P3 tree (see spec.md HISTORY) and are dropped from this plan: deepMerge3Way old-key preservation (now lives in `internal/merge/strategies.go:360` `deepMergeMap`, with explicit user-added-key preservation at line 388) and token-file 0600 (superseded by the F1 security redesign at `update.go:1375-1438` that does not persist tokens at all).

## §B Known Issues (findings inventory — anchors re-verified 2026-07-30 against current worktree)

| # | Anchor (symbol/content-token — re-verify at run) | Defect | Fix direction |
|---|---|---|---|
| 1 | `internal/cli/update.go` (1,905 LOC as of 2026-07-30; 8 sibling files already extracted: `update_archive.go`, `update_clean_install.go`, `update_cleanup.go`, `update_deny_migration.go`, `update_namespace_protect.go`, `update_preserve_inventory.go`, `update_tux.go`, `update_noise.go`) | `update.go` itself exceeds 1,200-line ceiling; bulk decomposition named in original audit is largely DONE | residual split to bring `update.go` under ceiling; per-file ceiling on the whole cluster |
| 2 | `buildGLMEnvVars` (`internal/cli/glm.go:917`); `ttyConfirmer` (`internal/cli/branch_protection.go:39-43`, gated on SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred pairing); `worktree_validation.go` (whole-file candidate at `internal/cli/worktree_validation.go`) | verified dead-code candidates | delete after caller-graph re-verification at run |
| 3 | `glm.go` GLM env inject/clear sites; `glm_tools.go`; `constitution.go`; `update.go`; CI `=="1"` checks | env names inlined; GLM inject/clear drift | `envkeys.go` constants + shared `glmEnvVarSet()` |
| 4 | tier-threshold literal `[]int{1, 3, 5, 10}` at THREE sites: `harness.go:150` (default), `harness.go:480` (struct default), `hook.go:1013` (`defaultTierThresholds`); dispatcher timeouts `hook.go:237` and `hook.go:361` (both `30*time.Second`); size caps; retry/circuit literals | threshold literals triplicated (original plan said "duplicated ×2" — undercounted by one) | `defaults.go` single source |
| 5 | `doctor.go`, `migration.go`, `clean.go`, `web_port*.go` | Korean UI strings vs `error_messages: en` | English literals (broader ~24-file Hangul set is out of scope — see spec.md §C) |
| 6 | constitution, tool_policy, github.go (+1 audited site) | byte-slice truncation breaks CJK | one rune-boundary truncate helper |
| 7 | `internal/cli/wizard/wizard.go:329` — the `huh.NewInput()` render dispatch for `QuestionTypeInput`; token-question IDs are `github_token` (`questions.go:225`) and `gitlab_token` (`questions.go:251`) | PAT plaintext echo (the dynamic-stepper half of the original audit finding is ALREADY satisfied by `stepperDenominator` at `wizard.go:225` and is dropped) | `.EchoMode(tty.EchoPassword)` (or equivalent) gated on token-question IDs |
| 8 | `internal/cli/worktree/tmux_integration.go:175` (`strings.Contains(trimmed, "team_mode:")`) AND `internal/cli/worktree/new.go:481` (`strings.Contains(trimmed, "tmux_preferred:")`) — TWO distinct fields, not one; tmux path quoting at `tmux_integration.go:114,124` | `strings.Contains` YAML parse (comments activate CG mode); unquoted tmux paths | `yaml.v3` parse for BOTH fields; `%q` quoting |

### Dropped from v0.1.0 (already-satisfied during P0-P3 window)

| Old # | Anchor (v0.1.0) | Why dropped (v0.2.0 evidence) |
|---|---|---|
| v0.1.0 #7 | `update.go:2396-2452` deepMerge3Way old-key drop | past EOF (file is 1,905 lines); merge machinery extracted to `internal/merge/`; `deepMergeMap` at `strategies.go:360` preserves user-added keys (line 388: "User added key - preserve") |
| v0.1.0 #8 | `update.go:2641-2651` token write 0644 | past EOF; superseded by F1 security fix at `update.go:1375-1438` — tokens intentionally NOT persisted, delegated to `gh`/`glab` CLI |

## §C Pre-flight

1. Confirm P0-P3 merged; refresh every anchor (this SPEC's anchors drift the most). v0.2.0 anchors were re-verified 2026-07-30 against worktree HEAD = origin/main, but update.go is still being touched by adjacent work — re-derive line numbers at run time, prefer symbol-name (content-token) anchors over line numbers.
2. ANALYZE: caller graph for each dead-code symbol (`go vet`, `grep -rn`, unused linter) — anything with a live caller leaves the inventory. Re-confirm `ttyConfirmer`'s SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred-pairing status before deletion.
3. PRESERVE: characterization tests over update flows (dry-run plan, settings sync) before any further split; the bulk of the original decomposition is done, so the characterization net focuses on the residual over-ceiling portion of `update.go`.
4. Inventory Hangul-bearing strings: `grep -rl '[가-힣]' internal/cli/ --include='*.go'` (expected: 30 production files; only the six in REQ-HYG-001-005 are in scope — diff actual against the in-scope set before editing).

## §D Constraints

- Behavior preservation is the default; the behavior-correcting fixes (PAT echo, YAML misparse, rune truncation) each need a RED reproduction test first.
- Decomposition commits are mechanical-move-only (no logic edits in the same commit) to keep review and git blame tractable. The path-classifier table extraction (if any residual remains — see design.md §A note) is a LOGIC refactor and MUST be a separate commit inside M5, NOT bundled into a mechanical-move commit (resolves design.md §A vs this section's tension flagged in audit D11).
- Dead-code deletion must not remove: the P0-wired lock path (`acquireUpdateLock`, `cleanStaleLock`), any symbol registered as a hook/live wrapper (cf. HOOK-DEADCODE-001 lesson: wrappers may be live via agent registration), confirmed-LIVE symbols (`scanDeprecatedPaths`, `cleanup_old_backups`), or exported symbols with external callers.
- §15 template-language-neutrality untouched: this SPEC edits Go sources only, no template files.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md (8 ACs after v0.2.0 renumber).
- E2: `go build ./... && go test ./internal/cli/... ./internal/cli/wizard/... ./internal/cli/worktree/... ./internal/merge/... -count=1` verbatim.
- E3: coverage of update cluster ≥ baseline (characterization suite counted).
- E4: dead-code grep audit — each deleted symbol returns 0 matches.
- E5: `golangci-lint run` no new findings; `wc -l` report for the update cluster files (target: every file ≤ 1,200 lines).

## §F Milestones (priority order)

- M1 — PRESERVE net: characterization tests + goldens for update flows; Hangul/threshold/env-literal inventories frozen as fixtures (in-scope Hangul set = the six audited files only).
- M2 — Correctness fixes (repro-first): rune-truncate helper + 4 sites; worktree `yaml.v3` parse for BOTH `team_mode` (`tmux_integration.go:175`) and `tmux_preferred` (`new.go:481`) + tmux path quoting (`tmux_integration.go:114,124`); wizard PAT masking (gated on `github_token`/`gitlab_token` IDs at the `huh.NewInput()` render site `wizard.go:329`). (deepMerge3Way and token-0600 dropped — see §B "Dropped from v0.1.0".)
- M3 — Hardcoding sweep: `envkeys.go` constants + `glmEnvVarSet()`; `defaults.go` threshold extraction (collapsing the THREE tier-threshold sites to one); CI env check correction fold-in where audited.
- M4 — English UI strings: six files converted; goldens updated deliberately.
- M5 — Structure: residual `update.go` decomposition to satisfy the 1,200-line ceiling (mechanical moves; path-classifier table extraction, if any residual remains, is a SEPARATE logic-refactor commit inside M5 per §D); dead-code deletion per verified inventory; final full-suite + lint + §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4 — this SPEC is last; starting it earlier guarantees merge conflicts on `update.go`/`glm.go`/`launcher.go`/`hook.go` with P0-P2 work.
- Shared-file overlap (rebase surface): `update.go`+`update_cleanup.go` (CRITICAL-001 e), `glm.go`/`launcher.go`/`glm_tools.go` (CRITICAL-001 a, CONCURRENCY-001 1-3), `hook.go` (CRITICAL-001 h, CONTRACT-001), `doctor.go` (LINTER-STALE-001 4).
- Anti-pattern: mixing mechanical moves with logic fixes in one commit — forbidden by §D (this includes the path-classifier table extraction — D11).
- Anti-pattern: deleting "dead" code by grep absence alone — reflection/registration/tag-gated callers must be checked (DDD ANALYZE names every caller).
- Anti-pattern: translating Korean strings by blind sed — each message reviewed so diagnostics stay accurate (mirrors docs-site no-blind-sed rule).
- Risk: characterization goldens over-fitting current bugs — where a golden captures a defect fixed in M2, regenerate the golden in the fixing commit with rationale.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 1/2/4/5, §4 rows 6-9, §5 P4 (re-verified 2026-07-30).
- Depends on: all four prior CLIFIX SPECs (P0-P3).
- CLAUDE.local.md §14 (hardcoding policy), §3/§6 (Go standards, test isolation), `language.yaml` `error_messages: en`.
- moai-workflow-ddd skill (ANALYZE-PRESERVE-IMPROVE governs M1/M5).
- v0.2.0 rescope driver: `.moai/reports/plan-audit/SPEC-CLIFIX-HYGIENE-001-review-1.md` (iter-1 FAIL 0.58, D1-D11).
