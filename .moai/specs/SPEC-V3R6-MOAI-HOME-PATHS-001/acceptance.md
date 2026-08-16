# Acceptance — SPEC-V3R6-MOAI-HOME-PATHS-001

All ACs machine-verifiable unless marked otherwise. Predicate P1 (the core migration check — home-scope anchored; verified pre-migration baseline = exactly 17 matches, matching plan.md §2 row-for-row. An earlier un-anchored draft of this predicate matched 218 sites by sweeping in project-scope `.moai` joins — the `Join\((home|homeDir|r\.homeDir)` first-argument anchor is load-bearing and MUST NOT be dropped):

```
grep -rn '"\.moai"' internal pkg cmd --include='*.go' | grep -v _test.go | grep -E 'Join\((home|homeDir|r\.homeDir)'
```

- **AC-MHP-001** — P1 returns 0 matches after migration (all home-scope `.moai` joins route through `internal/paths`; the segment literals live only in `internal/defs/dirs.go` and `internal/paths`).
- **AC-MHP-002** — `MOAI_HOME=/tmp/x` (absolute, non-empty) → `MoaiHome()` returns `/tmp/x` (unit test).
- **AC-MHP-003** — `MOAI_HOME=""` (empty) → falls back to home resolution (empty==unset; unit test via `t.Setenv`).
- **AC-MHP-004** — `MOAI_HOME=relative/path` → disregarded, falls back (XDG; unit test).
- **AC-MHP-005** — HOME-first precedence: with `HOME` set to a temp dir, `MoaiHome()`'s fallback resolves under it even when `os.UserHomeDir()` would return a different value (characterization aligned with `home_isolation_test.go`; non-parallel test).
- **AC-MHP-006** — `internal/config/envkeys.go` defines `EnvHome = "MOAI_HOME"`; `grep -rn '"MOAI_HOME"' internal pkg cmd --include='*.go' | grep -v _test.go | grep -vE 'envkeys\.go|internal/paths/'` returns 0.
- **AC-MHP-007** — `go list -deps ./internal/paths` lists only standard-library dependencies (stdlib-only).
- **AC-MHP-008** — `internal/hook/pre_tool.go` allowed-external-paths whitelist and `internal/core/project/root.go` home special-case resolve via `internal/paths` (grep for accessor call at those sites + existing behavioral tests still pass).
- **AC-MHP-009** — the `internal/cli/preference/cmd.go:152-167` duplicate helper is deleted; `internal/cli/homedir.go` delegates to `internal/paths` (grep: one implementation remains).
- **AC-MHP-010** — `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **AC-MHP-011** — `go test ./internal/paths/ ./internal/statusline/ ./internal/config/ ./internal/cli/ ./internal/update/ ./internal/kanban/ ./internal/hook/ ./internal/glmcred/ ./internal/cli/preference/ -count=1` exit 0; no test in the diff adds `t.Parallel()` alongside `t.Setenv`.
- **AC-MHP-012** — `go test -cover ./internal/paths/` reports ≥ 85.0%.
- **AC-MHP-013** — spec.md REQ-MHP-012 supersession note present and SPEC-V3R6-SESSION-HANDOFF-AUTO-001 is listed in `related_specs` (doc check; grep).
- **AC-MHP-014** — The `.env.glm` shell/Go split hazard (spec.md §4) is recorded in the sync-phase docs surface (CHANGELOG entry or follow-up card reference; doc check — discharges spec.md §4's known-hazard documentation obligation, the requirement source for this AC).

- **AC-MHP-015** — No caching in `internal/paths`: `grep -rn 'sync\.Once\|sync\.Mutex\|sync\.Map\|var moaiHome\|cached\|memo' internal/paths --include='*.go' | grep -v _test.go` returns 0 (per-call resolution per REQ-MHP-005; pattern broadened per audit D6 for literal-shape evasion resistance).
- **AC-MHP-016** — Agency-migration adapter override reach (audit D3): `grep -rn 'os\.UserHomeDir()' internal/cli/update.go internal/cli/migrate_agency.go` returns 0 matches after migration — update.go:827 (`runAgencyMigrationAdapter`) and migrate_agency.go:711 resolve via `internal/paths`, so an overridden `MOAI_HOME` reaches the `.migrate-tx-*` checkpoint path (site #17's home source). Pre-migration baseline: exactly 1 match per file (827, 711).

Notes: AC-002..005 tests must follow the non-parallel `t.Setenv` discipline (CLAUDE.local.md §6/§13) and their assertions exercise the `(string, error)` signature of REQ-MHP-006 (no `"."`-style fallback value is ever returned). Site #10 (migrate_profiles.go — tracked + clean since 504797021/#1571; the plan-time untracked-WIP premise was corrected at audit remediation, 2026-08-16) is included in AC-001's scope. AC-013/014 are doc checks verified at sync phase. AC count: 16, at the Tier M ceiling.
