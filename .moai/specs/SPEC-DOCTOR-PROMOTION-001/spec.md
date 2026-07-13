---
id: SPEC-DOCTOR-PROMOTION-001
title: "Doctor detection of plugin-deployed marker with promotion suggestion"
version: "0.1.0"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, cli, plugin, promotion, drift-detection"
tier: S
---

# SPEC-DOCTOR-PROMOTION-001 — Doctor plugin-deployed Marker Detection and Promotion Suggestion

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial draft — Tier S plan-phase artifact set (spec.md + plan.md, acceptance criteria inline in §3) |

## §1 Context and Motivation

Origin: REQ-BD-007, deferred from SPEC-MOC-BOOTSTRAP-DESKTOP-001 (sibling repo `claude.mo.ai.kr`). A plugin-based deployment can stamp a project's `.moai/config/sections/system.yaml` with a `plugin-deployed vX.Y.Z` marker instead of a plain semver. When the `moai` binary is later installed and `moai doctor` runs in such a project, doctor should surface the drift between the plugin-deployed template state and the binary-managed template tree, and suggest promotion — without performing any migration itself.

Verified baseline facts (read-only recon, cited — see plan.md §C for anchors):

- Doctor lives in `internal/cli/doctor.go` (848 lines at recon time); checks are `DiagnosticCheck` values registered in groups (`checkGroup`) inside `runGroupedChecks`, with the `moaiChecks` group as the natural home for a MoAI-config-scoped check. Fix suggestions are aggregated centrally.
- `system.yaml` path constants exist in `internal/defs` (`SystemYAML`, `MoAIDir`, `SectionsSubdir`). The config loader has NO `loadSystemSection` — a new check must read and parse the file directly.
- `moai init` stamps `system.yaml` with `moai:\n  version: %q\n  template_version: %q` from `pkg/version` (`GetVersion()`, default `v3.0.0-rc11`).
- ZERO existing references to `plugin-deployed` in the repo — this marker is greenfield (re-verified: `grep -rn "plugin-deployed" --include="*.go"` exit=1, 0 matches).
- Test convention: table-driven with `t.TempDir()` + fixture `.moai/config/sections/system.yaml`, asserting check `Name`/`Status`/`Detail` (pattern: `internal/cli/doctor_new_test.go`).

## §2 Requirements (GEARS)

### REQ-DP-001 — Marker detection (Event-driven)

**When** `moai doctor` runs in a project whose `.moai/config/sections/system.yaml` contains a version value matching `plugin-deployed v(\d+\.\d+\.\d+)` — accepting both the `moai:`-rooted `version:` key shape and a top-level `version:` key shape — the doctor **shall** emit a "Plugin Deployment" check with status Warn, reporting the deployed plugin version against the current binary version (`pkg/version` `GetVersion()`).

### REQ-DP-002 — Promotion suggestion only (Event-driven)

**When** the plugin-deployed marker is detected, the doctor **shall** suggest promotion in the check's Detail text — naming `moai init` as the promotion path to the binary-managed template tree. The check **shall not** perform any automatic migration and **shall not** write any files.

### REQ-DP-003 — Zero noise for binary-managed projects (State-driven)

**While** no plugin-deployed marker is present — the version value is a plain semver, or `system.yaml` is absent — the "Plugin Deployment" check **shall** return OK (marker absent) with no warning.

### REQ-DP-004 — Graceful degradation on unreadable input (Unwanted-behavior prevention)

**When** an unreadable or malformed `system.yaml` is encountered (read error on a present file, or content in which no version key can be recognized), the "Plugin Deployment" check **shall** degrade gracefully — returning OK or Warn with a parse note — and **shall not** return Fail and **shall not** abort or error out of the `moai doctor` run.

## §3 Acceptance Criteria (inline — Tier S)

All criteria are machine-verifiable. Commands run from the repo root `/Users/goos/moai/moai-adk-go`.

### AC-DP-001 — New check test suite passes

- Given a new table-driven test `TestCheckPluginDeployment` in `internal/cli/doctor_promotion_test.go` with fixture cases covering: (a) `moai:`-rooted marker → Warn, (b) top-level marker → Warn, (c) plain semver → OK, (d) missing `system.yaml` → OK, (e) malformed/unreadable file → non-Fail graceful result,
- When `go test -run TestCheckPluginDeployment ./internal/cli` is executed,
- Then it exits 0.

Gate: `go test -run TestCheckPluginDeployment ./internal/cli` → exit 0.

### AC-DP-002 — Check is registered in the doctor pipeline

- Given the implemented check function and its registration in the `moaiChecks` group,
- When `grep -c 'checkPluginDeployment' internal/cli/doctor.go` is executed,
- Then the count is ≥ 2 (one function definition + at least one registration site).

Gate: `grep -c 'checkPluginDeployment' internal/cli/doctor.go` → ≥ 2.

### AC-DP-003 — Zero-noise preserve (no regression)

- Given the full existing doctor check suite,
- When `go test ./internal/cli` is executed,
- Then it exits 0 (all pre-existing doctor tests, including golden tests, still pass — binary-managed projects see no new warning).

Gate: `go test ./internal/cli` → exit 0.

### AC-DP-004 — Suggestion names the promotion path

- Given the marker-detected fixture case,
- When the check result Detail text is asserted in `TestCheckPluginDeployment`,
- Then it contains the literal `moai init` (promotion suggestion) and both version strings (deployed vs binary).

Gate: `grep -n 'moai init' internal/cli/doctor_promotion_test.go` → ≥ 1 match (assertion present) AND AC-DP-001 passes.

## §4 Exclusions

The following items are explicitly out of scope for this SPEC.

### Out of Scope — Automatic migration or promotion execution

- No automatic promotion, migration, or template-tree mutation is performed by the check — REQ-DP-002 is suggestion-only.
- No file writes of any kind from the doctor check path.

### Out of Scope — Config loader extension

- No `loadSystemSection` (or equivalent) is added to the config loader; the check reads `system.yaml` directly (`os.ReadFile` + lightweight parse).

### Out of Scope — Initializer and template tree changes

- `internal/core/project/initializer.go` is untouched; `moai init` stamping behavior is unchanged.
- The binary-managed template tree (`internal/template/...`) is untouched.

### Out of Scope — Marker writing

- Nothing in this repo writes the `plugin-deployed` marker; producing the marker belongs to the plugin deployment flow in the sibling repo (`claude.mo.ai.kr`).

## §5 Constraints

- Scope boundary: exactly two code surfaces change at run-phase — `internal/cli/doctor.go` (new check function + one registration line) and `internal/cli/doctor_promotion_test.go` (new). Total delta < 150 LOC.
- Cross-platform safe: standard library file access only (`os`, `path/filepath`); no `syscall`.
- Path resolution via existing `internal/defs` constants (`MoAIDir`, `SectionsSubdir`, `SystemYAML`) — no hardcoded path strings.
- The check must never cause `moai doctor` to exit non-zero on its own (Warn is informational; Fail is prohibited per REQ-DP-004).
