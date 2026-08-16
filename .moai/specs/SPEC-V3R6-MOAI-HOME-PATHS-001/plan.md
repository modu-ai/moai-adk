# Plan — SPEC-V3R6-MOAI-HOME-PATHS-001

Tier M · Route B (PR, repo-local policy: all tiers via PR) · cycle_type tdd (brownfield, coverage ≥10%)

## 1. Milestones

| M | Scope | Owner | Verification |
|---|-------|-------|--------------|
| M1 | `internal/paths` package: `Home()`, `MoaiHome()`, sub-path accessors (REQ-MHP-001..007). TDD: RED tests for MOAI_HOME semantics FIRST | manager-develop | `go test ./internal/paths/ -count=1`; coverage ≥85% |
| M2 | `EnvHome` in `internal/config/envkeys.go` + alias in paths (REQ-MHP-004) | manager-develop | `grep -n 'EnvHome' internal/config/envkeys.go`; no inline literals (AC-6 grep) |
| M3 | Migrate 7 sites: statusline (model_cache ×2, usage, task), config/resolver ×2, glmcred | manager-develop | `go test ./internal/statusline/... ./internal/config/... ./internal/glmcred/...` |
| M4 | Migrate 10 sites: cli (launcher ×2, deps, migrate_profiles ×1, migrate_agency:184 + update.go:827 adapter), update (local, cache), kanban/status_read, hook (session_start, pre_tool) | manager-develop | `go test ./internal/cli/... ./internal/update/... ./internal/kanban/... ./internal/hook/...` |
| M5 | Coupled consumers + helper fate: pre_tool.go:275 whitelist, core/project/root.go, homedir.go delegate, preference/cmd.go duplicate deletion, glmcred routing (REQ-MHP-009/010) | manager-develop | `go test ./internal/hook/... ./internal/core/... ./internal/cli/preference/...`; duplicate-grep AC-9 |
| M6 | Full verification + docs: AC-1 predicate zero, cross-platform builds, coverage, hazard note in docs (REQ-MHP-011/012) | manager-develop → manager-docs | AC matrix in acceptance.md; `GOOS=windows GOARCH=amd64 go build ./...` |

## 2. Migration-Site Table (17 join sites)

| # | File:line | Target accessor | Notes |
|---|-----------|-----------------|-------|
| 1 | internal/statusline/model_cache.go:21 (Read) | StateDir + `state/last-model.txt` | injected homeDir seam preserved; caller resolves |
| 2 | internal/statusline/model_cache.go:50 (Write) | StateDir (dir bootstrap) | same |
| 3 | internal/statusline/usage.go:77 | CacheDir | injected seam preserved |
| 4 | internal/statusline/task.go:57 | StateDir + `state/last-session-state.json` | inline resolution → accessor |
| 5 | internal/config/resolver.go:318 | UserSettingsFile | user tier |
| 6 | internal/config/resolver.go:319 | UserConfigSectionsDir | user tier |
| 7 | internal/cli/launcher.go:494 | WorktreesDir | |
| 8 | internal/cli/launcher.go:924 | WorktreesDir | |
| 9 | internal/cli/deps.go:122 | StateDir + `state/loop` | |
| 10 | internal/cli/migrate_profiles.go:64 | ProfilesDir | tracked + clean since 504797021 (#1571) — the plan-time "UNTRACKED WIP" premise was invalidated by that merge and corrected at audit remediation (2026-08-16); join-only edit as before |
| 11 | internal/update/local.go:53 | ReleasesDir | |
| 12 | internal/update/cache.go:38 | CacheDir + `cache/update_check.json` | |
| 13 | internal/kanban/status_read.go:63 | WorktreesDir | |
| 14 | internal/hook/session_start.go:1155 | GlmEnvFile | |
| 15 | internal/hook/pre_tool.go:277 | MoaiHome (whitelist — REQ-MHP-009) | security-coupled |
| 16 | internal/glmcred/glmcred.go:41 (Path) | GlmEnvFile | stdlib-only constraint: paths must stay importable by glmcred |
| 17 | internal/cli/migrate_agency.go:184 | MoaiHome + segment (`.migrate-tx-*`) | **hidden site** — `Join(r.homeDir, ".moai", ...)` receiver-prefix form missed by the original predicate; home fed from update.go:827 adapter, which also migrates |

### 2.0 Membership Verification (2026-08-16, orchestrator self-check)

Corrected predicate `grep -rn '"\.moai"' internal pkg cmd --include='*.go' | grep -v _test.go | grep -E 'Join\((home|homeDir|r\.homeDir)'` returns **exactly 17** matches across exactly the 14 files listed above (verified per-file). A broad un-anchored scan (218 `".moai"` Join matches in 133 files) was reviewed: every non-table hit is project-scope (`projectRoot`/`cwd`/`Getwd`-ancestor first argument — e.g. `hook.go`, `navigator/sync/join.go`, `statusline/memory.go:71`, `preference/cmd.go:102`), correctly out of scope. The first-argument anchor is the scope boundary.

### 2.1 Disputed-Site Classification (resolved by direct line reads, 2026-08-16)

- `internal/cli/doctor.go:551` — **OUT**: joins `.claude/.mcp.json`, not `.moai` (Claude-home territory, CLAUDE_CONFIG_DIR's domain — not MOAI_HOME's).
- `internal/cli/statusline.go:39` — **OUT**: `os.Chdir(home)` cwd-guard fallback; no `.moai` join (may later use `paths.Home()`, non-blocking nicety — NOT in AC).
- `internal/cli/update.go:827` — **IN** (feeds site #17's runner).
- `internal/cli/migrate_agency.go:711` — **IN** (same runner's home resolution).

## 3. PRESERVE List (do not touch)

.mcp.json · internal/cli/todo.go · internal/cli/todo_test.go · internal/cli/update.go **(only the :827 home-resolution migrates; file is tracked + clean at HEAD — the plan-time WIP-hunk preservation constraint is void)** · internal/config/defaults.go · .moai/config/sections/cache.yaml · internal/cli/memory.go · internal/cli/memory_test.go · internal/cli/migrate_profiles.go (join-only edit) · internal/cli/migrate_profiles_test.go · internal/hook/memo/taxonomy/linkage.go · internal/hook/memo/taxonomy/linkage_test.go · .moai/reports/** · all other SPEC dirs.

## 4. Resolved Decisions

- **[RESOLVED 2026-08-16: accessor granularity — audit iteration 1 D1 remediation]** Every migration site resolves inside the sketched 8-accessor surface (Design Sketch, spec.md §3) plus `MoaiHome()`; the per-site mapping is fixed in the §2 table, verified by direct line reads at remediation: model_cache ×2 / task / deps → StateDir(+segment); launcher ×2 / status_read → WorktreesDir; session_start / glmcred → GlmEnvFile; usage / update-cache → CacheDir(+segment); local → ReleasesDir. No accessor beyond the sketched surface is added; exact accessor signatures are pinned by the M1 RED-phase tests. (User-approved via AskUserQuestion, 2026-08-16.)
- None else — remaining decisions were settled by research evidence + user directive.

## 5. Risks

- **UNTRACKED-WIP coupling — RESOLVED 2026-08-16**: both files (migrate_profiles.go, update.go) are tracked + clean since 504797021 (#1571); the plan-time premise was invalidated by that merge and re-verified at audit remediation (iteration 1 D2). Explicit-pathspec staging discipline (never `git add -A`) still applies as general hygiene.
- **Windows HOME semantics**: unifying on HOME-first changes behavior for the 13 `os.UserHomeDir()` sites on Windows when HOME is set (test overrides) — this is the homedir.go contract already guarded by `home_isolation_test.go`; run Windows CI matrix.
- **Shell/Go split of `.env.glm`** under non-default MOAI_HOME (spec.md §4) — documented hazard, follow-up card.
