# Acceptance — SPEC-V3R6-MOAI-CLEAN-HOME-001

- **AC-MCH-001** — `moai doctor` output contains a `Home Disk Usage` check (grep on output; fixture home with populated `claude-profiles/`).
- **AC-MCH-002** — On a fixture with two profiles carrying byte-equal `plugins/` dirs, the check reports a duplicate cluster naming both profiles (unit test on the detector function).
- **AC-MCH-003** — Estimated cleanable bytes above `DefaultHomeDiskWarnBytes` yields WARN status (fixture); below yields OK.
- **AC-MCH-004** — `moai clean --home` without `--force` mutates nothing: fixture tree file count + mtimes identical before/after (dry-run test).
- **AC-MCH-005** — With `--force`, only allowlisted categories are removed: aged `debug/`, `releases/` beyond current+N, aged `logs/`, aged `backups/removed-*` (fixture with aged + fresh entries; fresh survive).
- **AC-MCH-006** — Carve-out guard (REQ-MCH-005): after `--force` on a fixture, `projects/` (root + per-profile), `config/` (root + per-profile), `state/`, `credentials.yaml` — including a `credentials*`-named file planted inside an aged `backups/removed-*` directory (carve-out wins inside allowlisted containers) — `launch.yaml`, `preferences.yaml`, `worktrees/`, `mcp/`, `bin/`, `search/`, `studio/`, and every `plugins/` dir still exist.
- **AC-MCH-007** — `releases/` retains the current-version binary plus `DefaultReleaseKeep` (3) newest; older removed only under `--force`.
- **AC-MCH-008** — `state.home_retention_days` honored (aged-vs-fresh cutoff), read from the home tier (`~/.moai/config/sections/state.yaml` via `paths.UserConfigSectionsDir()`); absent key defaults to 30; explicit 0 disables (three sub-tests).
- **AC-MCH-009** — Project-scope `moai clean` (no `--home`) behavior unchanged: existing runs/ retention test still passes unmodified.
- **AC-MCH-010** — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test ./internal/cli/ ./internal/config/ -count=1` exit 0.
- **AC-MCH-011** — No inline threshold literals: `grep -rn '500\s*\*\|524288000' internal/cli/doctor_disk.go internal/cli/clean_home.go` returns 0 (constants live in defaults.go).
- **AC-MCH-012** — MOAI_HOME redirect (REQ-MCH-008, audit D7): with an absolute `MOAI_HOME` pointing at a second fixture tree, `clean --home --force` deletes aged entries under the override root and leaves the default-`HOME` fixture untouched (the dependency SPEC's core promise).

Notes: all fixture tests use a hermetic home — `t.Setenv("HOME", t.TempDir())` + `t.Setenv("MOAI_HOME", "")` (scrubbed so ambient overrides cannot leak in), non-parallel. `~/.claude` appears only as a doctor report line — no AC mutates it by construction (AC-004/006 scope `~/.moai` fixture). AC count: 12 (Tier M ceiling 16).
