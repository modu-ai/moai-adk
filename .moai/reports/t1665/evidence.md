# t1665 — doctor profile-cap read path (Class B, run → sync)

Lane: lane-17 · run tmvgvn · tree `.moai/worktrees/t1665` (branch WT-t1631-d-10, base a4130a809) · leased 2026-10-17.
Source: t1631 split proposal row D / blocker D10 (commit 26e2f2bc4, `.moai/reports/t1631/split-proposal.md` lines 19 and 32). Prerequisite t1631 (A) landed: `459d17e7c Merge WT-capacity-reclaim into main (card t1631, integration merge)`.

## Cause (established by reading the tree at a4130a809)

Two sources of truth for the per-profile size cap:

- `internal/cli/doctor_disk.go:155` — `if size > config.DefaultProfileMaxBytes` (compiled 5 GiB, `internal/config/defaults.go:471`)
- `internal/cli/doctor_disk.go:261` — the detail line prints `formatDiskBytes(config.DefaultProfileMaxBytes)`
- The SessionStart advisory (`internal/hook/session_start_hygiene.go:108`) reads `state.profile_max_bytes` through `hygiene.LoadProfileCapMaxBytes(projectDir)` — HOME tier, then PROJECT tier, **no compiled fallback** (`internal/hygiene/profile_cap.go:36-66`; its `@MX:NOTE` forbids substituting `config.DefaultProfileMaxBytes`).
- The shipped template sets `state.profile_max_bytes: 8589934592` (8 GiB, `internal/template/templates/.moai/config/sections/state.yaml:21`).

With the shipped config, doctor flags profiles above 5 GiB while the advisory warns above 8 GiB, and changing the key moves only the advisory.

## Decision record (lane judgment, auto-semantics §6 terminal step)

- **Scope:** the profile cap only (the two lines the card names). The other compiled values in `doctor_disk.go` (`DefaultHomeDiskWarnBytes`, `DefaultProfileUnusedDays`, `DefaultHomeCleanRetentionDays`, `DefaultReleaseKeep`) have no config key today; giving them one is new surface, not this card.
- **Absent key:** doctor follows the loader contract — no compiled fallback, so no oversize list and no "profiles above" line. A loader error degrades the same way and is named in the detail, never fatal (doctor is advisory).
- **Project dir:** `resolveProjectDir()` (`internal/cli/session.go:305`), the resolver the CLI already uses.
- **Constant:** `config.DefaultProfileMaxBytes` may be removed only if no reference remains after the test change; otherwise it stays.

## Run record (2026-10-17)

- Stage: `moai factory stage t1665 run .moai/reports/t1665/evidence.md` → `t1665 run v5 lease renewed`.
- Implementer: `manager-develop` (cycle tdd), spawned by lane-17 without an explicit `effort` (deviation from the class-B `medium` rule in factory-dispatch.md § Lane spawn authority — it ran at the session effort).
- RED `ec8c2bd58` — `TestHomeDiskReportOversizeUsesConfiguredProfileCap` (3 cases) failed on the pre-fix code: `oversize profiles = [], want exactly [above] under a 4096-byte cap` / `oversize profiles = [huge], want none when state.profile_max_bytes is absent` / `... want none when the cap cannot be loaded`.
- GREEN `27ea20974` — `gatherHomeDiskReport` loads the cap once via `hygiene.LoadProfileCapMaxBytes(resolveProjectDir())`; `homeDiskReport.ProfileCapBytes` (0 = no cap) + `ProfileCapErr`; loader error adds one detail line and skips the check. `config.DefaultProfileMaxBytes` removed (no code reference left; comments only).
- Diff vs a4130a809: 4 files, +116/−6 (`internal/cli/doctor_disk.go`, `internal/cli/doctor_disk_test.go`, `internal/config/defaults.go`, `internal/hygiene/profile_cap.go`).
- Lane-local verification, re-run by the lane at `27ea20974` (TMPDIR=GOTMPDIR=/private/tmp/dcap-tmp):
  - `go test -count=1 -run 'HomeDisk|ProfileCap|DoctorDisk' ./internal/cli/ ./internal/hygiene/` → `ok internal/cli 3.747s`; `internal/hygiene [no tests to run]` (hygiene is covered by the implementer's full-package run: `ok internal/hygiene 0.779s`).
  - `golangci-lint run --new-from-rev a4130a809 ./internal/cli/ ./internal/config/ ./internal/hygiene/` → exit 0, `0 issues.`
  - Implementer-reported: `go build ./...` exit 0; `go vet` (3 pkgs) exit 0; `gofmt -l` empty; `go test ./internal/config/...` ok.
- Card review: `card-review.md` (codex scope=card, `pass`, 0 findings, advisory).
- Docs decision (lane judgment): no CHANGELOG line in this card — the CHANGELOG is composed at release time by the release harness; the behaviour change (doctor follows `state.profile_max_bytes`) is recorded here and in the fix commit body.

Gaps: full `internal/cli` suite and repo-wide lint not run (CI owns them; repo-wide lint on this base already fails on `internal/cli/factory_run_adopt.go` errcheck ×3, outside this card); darwin only.

## Completion assertion

Done when: a test writes `state.profile_max_bytes` = X into a fixture HOME tier and `gatherHomeDiskReport` flags exactly the profiles above X (and the detail line prints X); with the key absent, no profile is flagged.
