---
id: SPEC-GO-TOOLCHAIN-SEC-001
title: "Go toolchain security bump (go1.26 → go1.26.4, 19 stdlib vulns → 0)"
version: "0.1.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "patch-release"
module: "go.mod"
lifecycle: spec-anchored
tags: "security, dependencies, toolchain, govulncheck"
tier: S
---

# SPEC-GO-TOOLCHAIN-SEC-001 — Go toolchain security bump

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-06-03 | 0.1.0 | manager-spec | Initial plan-phase authoring (Tier S minimal). status: draft. |
| 2026-06-03 | 0.1.0 | manager-spec | plan-patch D1/D2/D3: AC-GTS-005 (REQ-GTS-006 traceability), AC-GTS-006 (toolchain pre-check gate, gates AC-GTS-001), `tier: S` frontmatter. status: draft. |

## A. Context and Problem Statement

`govulncheck ./...` reports **"Your code is affected by 19 vulnerabilities from the Go
standard library"**. All 19 affecting findings are in the Go **standard library** at the
currently-declared toolchain `go1.26.0` — none are in third-party `require` modules. The
findings span `crypto/x509`, `net/http`, `crypto/tls`, `html/template`, `net/mail`,
`net/textproto`, `net`, and `archive/tar`.

Representative finding IDs (subset): GO-2026-5039, GO-2026-5037, GO-2026-4986,
GO-2026-4982, GO-2026-4980, GO-2026-4977, GO-2026-4971, GO-2026-4947, GO-2026-4946,
GO-2026-4918, GO-2026-4870, GO-2026-4869, GO-2026-4866.

The findings are fixed at toolchain releases go1.26.2 / go1.26.3 / go1.26.4. The highest
fix version required across all findings is **go1.26.4**, so bumping the project's Go
toolchain to go1.26.4 (or any later patch) clears all 19 findings.

Two artifacts pin the toolchain version:

1. `go.mod` line 3 declares `go 1.26` (no patch component, no `toolchain` directive).
2. Nine CI workflow steps hardcode `go-version: "1.26"`:
   `.github/workflows/claude.yml:42`, `.github/workflows/ci.yml:87`, `ci.yml:179`,
   `ci.yml:205`, `ci.yml:250`, `ci.yml:292`, `.github/workflows/codeql.yml:85`,
   `.github/workflows/release-pr-multi-os.yml:59`, `.github/workflows/release.yml:31`.

Three further workflows already derive the version from `go.mod` via
`go-version-file` (`template-neutrality-check.yaml:52`, `spec-lint.yml:15`,
`spec-status-auto-sync.yml:25`) and propagate a go.mod bump automatically with no edit.

## B. Why (motivation)

- **Security exposure**: The 19 stdlib vulnerabilities affect HTTP serving, TLS, x509
  certificate parsing, HTML templating, and archive extraction — all reachable surfaces
  for a CLI that performs network operations, template rendering, and file extraction.
- **Single declarative change closes all 19**: Because every finding is a stdlib
  vulnerability fixed by a patch toolchain release, no source-code change is required —
  only a toolchain version bump.
- **CI/local drift risk**: With the toolchain pinned at `go 1.26` (unpinned patch),
  different CI runners and developer machines may resolve to different patch levels.
  Pinning the minimum patch makes the security floor explicit and reproducible.

## C. Requirements (GEARS)

REQ-GTS-001 (Ubiquitous): The project Go toolchain **shall** be bumped such that
`govulncheck ./...` reports zero affecting vulnerabilities.

REQ-GTS-002 (Ubiquitous): The `go.mod` `go` directive **shall** declare a Go version at
or above go1.26.4 (the highest fix version across all 19 findings).

REQ-GTS-003 (Event-driven): **When** the toolchain version is bumped in `go.mod`, the
project **shall** apply the chosen CI-pin strategy consistently across all nine
workflow steps that currently hardcode `go-version: "1.26"`.

REQ-GTS-004 (State-driven): **While** the bump is applied, the third-party `require`
blocks in `go.mod` **shall** remain unchanged (the third-party modules are clean per
govulncheck and are out of scope).

REQ-GTS-005 (Ubiquitous): After the bump, `go build ./...` and `go test ./...` **shall**
pass under the bumped toolchain.

REQ-GTS-006 (Capability gate): **Where** an explicit `toolchain` directive is added to
`go.mod`, it **shall** name a version equal to the bumped `go` directive (so
`GOTOOLCHAIN=auto` can auto-acquire the required patch toolchain).

## D. Acceptance Criteria Summary

The full AC matrix lives in `acceptance.md` (6 ACs). The anchor criteria:

- AC-GTS-001 — `govulncheck ./...` reports 0 affecting vulnerabilities (19 → 0).
  Precondition: AC-GTS-006 (toolchain pre-check gate) MUST PASS first.
- AC-GTS-002 — `go build ./...` and `go test ./...` pass under the bumped toolchain.
- AC-GTS-003 — `go.mod` `go` directive reflects the bumped version; the chosen CI-pin
  strategy is applied consistently across all 9 workflow steps.
- AC-GTS-004 — `git diff go.mod` shows ONLY the `go`/`toolchain` directive changed; no
  `require` block line is modified (scope-discipline guard).
- AC-GTS-005 — REQ-GTS-006 traceability: if a `toolchain` directive is present, its
  version equals the `go` directive version (`awk` equality guard).
- AC-GTS-006 — toolchain pre-check gate: effective `go version` ≥ go1.26.4; gates
  AC-GTS-001 evidence acceptance (false-PASS guard, promoted from EC-1).

## E. Exclusions (What NOT to Build)

- **Third-party `go.mod` `require` version changes** — govulncheck flags zero AFFECTING
  third-party vulnerabilities. Modules appearing in govulncheck's informational "modules
  you require but your code doesn't appear to call" list are NOT affecting and are out of
  scope. No `require` line is to be touched.
- **vitest / npm / `moai-adk-ts`** — the two critical Dependabot alerts (`vitest` in
  `moai-adk-ts/package.json` + `package-lock.json`) belong to a separate TypeScript
  project whose paths do not exist in this Go repository. Deferred follow-up for the TS
  project; explicitly excluded here.
- **Recurring govulncheck CI gate** — adding a scheduled or per-PR govulncheck job is NOT
  in scope (the user selected the minimal option, not the "+ CI gate" option).
- **Dependency-update policy docs / Dependabot config** — no policy documentation or
  Dependabot configuration is created (minimal scope).
- **Source-code changes** — no Go source file is modified; the fix is purely a toolchain
  version bump. (If a future Go API deprecation surfaces during the bump, that is a
  separate SPEC.)
