# SPEC-DOCTOR-PROMOTION-001 — Implementation Plan

> Tier S. Milestones are ordered by decision-reversibility: the decisions most likely to change under review (marker grammar, user-facing text) lead; mechanical implementation and verification steps close.

## §A Context

`moai doctor` gains one new diagnostic check ("Plugin Deployment") that detects a `plugin-deployed vX.Y.Z` marker in `.moai/config/sections/system.yaml` and suggests promotion via `moai init`. Suggestion-only; no migration, no writes. See spec.md §1 for origin (REQ-BD-007 deferral) and verified baseline facts.

## §B Known Issues / Risks

- **No config-loader support**: the loader has no `loadSystemSection`; the check must read the file directly. Risk of divergence from future loader behavior is accepted (Tier S, single call site).
- **Two key shapes**: `moai:`-rooted `version:` (what `moai init` stamps) vs top-level `version:` (defensive acceptance). A regex over raw file content covers both without a YAML dependency decision; if YAML parsing is chosen instead, both shapes must be probed explicitly.
- **Golden tests**: `internal/cli/doctor_golden_test.go` snapshots doctor output. Adding a check to `moaiChecks` may shift golden output. Mitigation: run `go test ./internal/cli` early (AC-DP-003) and update goldens only if the harness regenerates them by design — never hand-edit expectations to mask a behavior change.

## §C Pre-flight (verified recon anchors — cite, do not re-derive)

Content-token anchors in `internal/cli/doctor.go` (line numbers from recon, drift-tolerant):

- `DiagnosticCheck` struct (~:27), `checkGroup` (~:35), `runDoctor` (~:59)
- Check registration in `runGroupedChecks` (~:157); `moaiChecks` group (~:172-185); model registration example `checkMigration(cwd, verbose)` (~:184, registered pattern to mirror)
- Fix-suggestion aggregation (~:122-124)
- Path constants: `internal/defs/files.go` `SystemYAML`; `internal/defs/dirs.go` `MoAIDir`, `SectionsSubdir`
- Version source: `pkg/version` `GetVersion()` (default `v3.0.0-rc11`); `moai init` stamp site `internal/core/project/initializer.go:399-406`
- Test conventions: `internal/cli/doctor_new_test.go` (table-driven, `t.TempDir()` + fixture system.yaml, assert `Name`/`Status`/`Detail`)
- Greenfield: zero `plugin-deployed` references repo-wide (re-verified at plan time, grep exit=1)

## §D Constraints

- Files touched at run-phase: `internal/cli/doctor.go` + `internal/cli/doctor_promotion_test.go` ONLY. No loader, initializer, or template-tree changes.
- Total delta < 150 LOC (target: check function ~50-70 LOC, registration 1 line, test ~60-75 LOC).
- Cross-platform: `os.ReadFile` + `filepath.Join`; no `syscall`.
- Never Fail, never error out of doctor (REQ-DP-004); Warn is the strongest status this check emits.
- development_mode per repo quality.yaml (TDD default): test-first — RED before GREEN.

## §E Self-Verification (run-phase deliverables)

Run as a single-turn read-only batch at run-phase completion:

1. `go test -run TestCheckPluginDeployment ./internal/cli` → exit 0 (AC-DP-001)
2. `grep -c 'checkPluginDeployment' internal/cli/doctor.go` → ≥ 2 (AC-DP-002)
3. `go test ./internal/cli` → exit 0 (AC-DP-003, golden + zero-noise preserve)
4. `grep -n 'moai init' internal/cli/doctor_promotion_test.go` → ≥ 1 (AC-DP-004)
5. `go vet ./internal/cli` → exit 0
6. LOC budget: `git diff <run-base>..HEAD --numstat -- internal/cli/doctor.go internal/cli/doctor_promotion_test.go` → added lines sum < 150

## §F Milestones (decision-reversibility order)

### M1 — Marker grammar and parse strategy (highest change-likelihood)

Decide and freeze: detection regex `plugin-deployed v(\d+\.\d+\.\d+)` applied to the `version:` value; accepted key shapes = `moai:`-rooted `version:` AND top-level `version:`; parse mechanism = `os.ReadFile` + regex over raw content (no new YAML dependency, no loader change). This is the data-model-like decision a reviewer is most likely to challenge — surface it first.

### M2 — User-facing Detail text (UX flow)

Fix the Warn Detail wording: it must carry (a) deployed plugin version, (b) current binary version from `GetVersion()`, (c) the literal promotion suggestion naming `moai init` as the path to the binary-managed template tree. Fix the check `Name` as "Plugin Deployment". Ensure the fix-suggestion aggregation path (doctor.go ~:122-124) picks the suggestion up consistently with sibling checks.

### M3 — RED: `doctor_promotion_test.go`

Write the failing table-driven test first, mirroring `doctor_new_test.go` conventions (`t.TempDir()` + fixture `.moai/config/sections/system.yaml`; assert `Name`/`Status`/`Detail`). Table rows: moai-rooted marker → Warn; top-level marker → Warn; plain semver → OK; missing file → OK; malformed/unreadable content → non-Fail graceful. Confirm RED (test fails because the check does not exist yet).

### M4 — GREEN: `checkPluginDeployment` + registration (mechanical)

Implement `checkPluginDeployment(cwd string, verbose bool)` (signature mirroring `checkMigration`) in `internal/cli/doctor.go`; register one line in the `moaiChecks` group of `runGroupedChecks`. Use `internal/defs` path constants. Confirm GREEN.

### M5 — Verification batch + golden reconciliation (mechanical)

Run §E items 1-6 as a single parallel batch. If golden tests shifted due to the new check row, regenerate via the harness's sanctioned mechanism and re-run item 3.

## §G Anti-Patterns

- Adding a `loadSystemSection` to the config loader "while we're here" — explicitly out of scope.
- Making the check Fail on malformed YAML — violates REQ-DP-004.
- Writing or "fixing" system.yaml from the check — violates REQ-DP-002 (suggestion-only).
- Hand-editing golden expectations to hide an unintended output change.
- Hardcoding `.moai/config/sections/system.yaml` as a string instead of composing `internal/defs` constants.

## §H Cross-References

- spec.md §2 (REQ-DP-001..004), §3 (AC-DP-001..004), §4 (exclusions)
- Origin: SPEC-MOC-BOOTSTRAP-DESKTOP-001 REQ-BD-007 (sibling repo `claude.mo.ai.kr`)
- Conventions: `internal/cli/doctor_new_test.go`, `internal/cli/doctor_golden_test.go`
- Version source: `pkg/version/version.go` `GetVersion()`
