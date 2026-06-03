---
id: SPEC-GO-DEPS-UPDATE-001
title: "Go third-party dependency maintenance update (patch + golang.org/x/* minor; 0 major bumps)"
version: "0.1.1"
status: in-progress
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "patch-release"
module: "go.mod"
lifecycle: spec-anchored
tags: "dependencies, maintenance, go-modules, latest-ness"
tier: S
---

# SPEC-GO-DEPS-UPDATE-001 — Go third-party dependency maintenance update

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-06-03 | 0.1.0 | manager-spec | Initial plan-phase authoring (Tier S minimal). status: draft. Phase 1 (patch) + Phase 2 (golang.org/x/* minor). Phase 3 (charmbracelet/x/*) explicitly excluded. |
| 2026-06-03 | 0.1.1 | manager-spec | plan-patch (plan-auditor PASS-WITH-DEBT 0.83, 5 defects). D1: AC-GDU-001 `grep -E` `\|`→bare `\|` (escaped pipe was literal, never matched). D2: gorilla/websocket rationale corrected (transitive via powernap→jsonrpc2 LSP, not net/http stdlib). D3: charmbracelet/x/* exclusion clarified as ALL modules (listed = update-available subset). D4: regression grep `^(FAIL\|--- FAIL)`→`^FAIL\|--- FAIL:` to catch indented sub-test failures. D5: REQ-GDU-007 orphan resolved — AC-GDU-007 now dedicated to the `.go`-edit gate; scope-guard split into AC-GDU-008, no-new-dep into AC-GDU-009 (8→9 ACs). |

## A. Context and Problem Statement

This is a **maintenance (latest-ness)** update of third-party Go dependencies, NOT a
security-driven one. `govulncheck ./...` already reports **0 affecting third-party
vulnerabilities** — the clean baseline established by the just-completed sibling
SPEC-GO-TOOLCHAIN-SEC-001 (which bumped the Go toolchain to go1.26.4 and cleared all 19
stdlib findings). The goal here is simply to keep direct + golang.org/x/* dependencies
current.

Ground-truth (verified by the orchestrator via `go list -u -m all` + `go mod edit -json`
under go1.26.4):

- **Zero major-version (vN+1) bumps are available** for any direct dependency. Every
  available update is patch- or minor-level, so there is **no API-breaking risk**.
- Four direct dependencies have updates available:
  - `github.com/charmbracelet/x/powernap` v0.1.5 → v0.1.6 (patch)
  - `github.com/go-playground/validator/v10` v10.30.2 → v10.30.3 (patch)
  - `github.com/mattn/go-runewidth` v0.0.23 → v0.0.24 (patch)
  - `golang.org/x/sys` v0.44.0 → v0.45.0 (minor)
- The other 13 direct deps are already latest (charmbracelet bubbletea / bubbles / huh /
  lipgloss / colorprofile, cobra, yaml.v3, x/sync, x/text, go-isatty, termenv, goleak,
  go-tree-sitter).
- Notable indirect (transitive) updates: `golang.org/x/crypto` v0.51 → v0.52,
  `golang.org/x/net` v0.53 → v0.55, `golang.org/x/tools` v0.44 → v0.45,
  `golang.org/x/mod` v0.35 → v0.36.

## B. Why (motivation)

- **Latest-ness hygiene**: Keeping dependencies current reduces the size of future
  bumps and keeps the project aligned with upstream bug fixes that have already shipped
  as patch/minor releases.
- **Low risk by construction**: Because zero major bumps are available, the entire scope
  is patch/minor. Patch bumps are backward-compatible by semver; the `golang.org/x/*`
  libraries follow stable API conventions where minor bumps are routinely low-risk.
- **Clean separation from security work**: The security floor was already established by
  SPEC-GO-TOOLCHAIN-SEC-001. This SPEC must NOT regress that floor — the maintenance bump
  must keep `govulncheck` affecting-count at 0.

## C. Scope — Two Phases (user-selected: Phase 1 + Phase 2, Tier S)

**Phase 1 — patch-only.** Run `go get -u=patch ./...`. This bumps patch-level direct
dependencies (powernap, validator/v10, go-runewidth) plus patch-level transitive
dependencies.

**Phase 2 — golang.org/x/* minor.** Explicit minor bumps of the API-stable
`golang.org/x/*` libraries: `golang.org/x/sys@v0.45.0` (direct) plus the indirect
`golang.org/x/crypto`, `golang.org/x/net`, `golang.org/x/tools`, `golang.org/x/mod` to
their latest minor versions. The `golang.org/x/*` libraries follow stable API
conventions; minor bumps are low-risk.

**Both phases** are followed by `go mod tidy` to keep `go.mod` / `go.sum` clean.

## D. Requirements (GEARS)

REQ-GDU-001 (Ubiquitous): The project's patch-level direct + indirect dependencies
**shall** be bumped to their latest patch via `go get -u=patch ./...` (Phase 1).

REQ-GDU-002 (Ubiquitous): The `golang.org/x/sys` direct dependency **shall** be bumped to
v0.45.0, and the indirect `golang.org/x/crypto`, `golang.org/x/net`, `golang.org/x/tools`,
`golang.org/x/mod` **shall** be bumped to their latest minor (Phase 2).

REQ-GDU-003 (Event-driven): **When** both Phase 1 and Phase 2 are applied, the project
**shall** run `go mod tidy` so that `go.mod` / `go.sum` remain a clean, minimal tree.

REQ-GDU-004 (Ubiquitous): After the bump, `go build ./...` **shall** exit 0 on the host
platform AND under `GOOS=windows GOARCH=amd64` AND under `GOOS=linux GOARCH=amd64`
(cross-platform build integrity).

REQ-GDU-005 (Ubiquitous): After the bump, `go test ./...` **shall** pass with no NEW
regressions relative to the pre-bump baseline. Known pre-existing non-regression failures
(see §E.1) are excluded from this gate.

REQ-GDU-006 (State-driven): **While** the bump is applied, `govulncheck ./...` affecting
third-party vulnerability count **shall** remain 0 (the maintenance update must not
introduce a newly-affecting third-party vulnerability).

REQ-GDU-007 (Capability gate): **Where** a `golang.org/x/*` minor bump introduces an API
change that requires a `.go` source edit, that source edit is permitted but **shall** be
explicitly noted in `progress.md`. Expected outcome: none (no major bumps are in scope).

REQ-GDU-008 (State-driven): **While** the bump is applied, the diff **shall** be limited
to `go.mod` + `go.sum` (plus any `.go` edit permitted by REQ-GDU-007); no unrelated file
is touched (scope-discipline guard).

## E. Acceptance Criteria Summary

The full AC matrix lives in `acceptance.md` (9 ACs). The anchor criteria:

- AC-GDU-001 — Phase 1 applied: `go get -u=patch ./...` ran; powernap v0.1.6 +
  validator/v10 v10.30.3 + go-runewidth v0.0.24 present in `go.mod`.
- AC-GDU-002 — Phase 2 applied: `golang.org/x/sys` ≥ v0.45.0 (direct) + the four indirect
  `golang.org/x/*` libs bumped to latest minor.
- AC-GDU-003 — `go mod tidy` leaves a clean tree (a second `go mod tidy` re-run proposes
  no further change).
- AC-GDU-004 — cross-platform build: host + windows/amd64 + linux/amd64 `go build ./...`
  all exit 0.
- AC-GDU-005 — `go test ./...` passes with no NEW regressions; the §E.1 known failures are
  excluded.
- AC-GDU-006 — `govulncheck ./...` affecting third-party count stays 0.
- AC-GDU-007 — `.go`-edit capability gate (REQ-GDU-007): IF any `.go` file appears in
  `git diff --name-only`, THEN `progress.md` records that file + the API-change reason;
  expected outcome is go.mod/go.sum-only with no `.go` edit.
- AC-GDU-008 — scope guard (REQ-GDU-008): `git diff --name-only` shows only `go.mod` +
  `go.sum` (plus any `.go` edit already accounted for by AC-GDU-007).
- AC-GDU-009 — no new dependency added (REQ-GDU-008): the `require` block gains no module
  that was not already present (transitive churn from tidy is allowed; net-new direct
  imports are not).

### E.1 Known pre-existing non-regression failures (excluded from the AC-GDU-005 gate)

These failures exist on the pre-bump baseline and are NOT caused by the dependency update.
A tester MUST NOT mistake them for update-induced breakage:

- `internal/hook` — `TestHookWrapper_MoaiBinaryFallback` and `TestHookWrapper_ValidJSON`
  (~5s timing-flaky; pass on retry).
- `internal/template` — `TestOutputStylesTemplateLiveParity` (einstein.md drift, unrelated
  to dependencies).

## Out of Scope (What NOT to Build)

### Out of Scope — Phase 3 charmbracelet/x/* indirect bumps (deferred)

- **`charmbracelet/x/*` indirect bumps** are **DEFERRED**, NOT in scope. This exclusion
  covers **ALL** transitively-pinned `charmbracelet/x/*` modules; the
  currently-update-available subset is (conpty / xpty / errors / exp / golden), but the
  exclusion is not limited to that subset (the others remain out of scope too).
  Rationale: the parent charmbracelet **direct** dependencies
  (bubbletea / bubbles / huh / lipgloss / colorprofile) are already at their latest
  versions and pin these `charmbracelet/x/*` indirect versions transitively.
  Force-bumping the indirect `charmbracelet/x/*` modules ahead of the parents risks
  **pin skew** — an indirect version newer than what the pinned parent expects. These
  indirect modules must be driven by future parent (bubbletea/huh/...) updates, not
  force-bumped here. Leaving them alone keeps the charmbracelet dependency subtree
  internally consistent.

### Out of Scope — other deliberate exclusions

- **`gorilla/websocket`** — a **pure transitive** dependency, **not directly imported by
  our code**. Provenance (`go mod why github.com/gorilla/websocket`):
  `internal/lsp/transport → charmbracelet/x/powernap → sourcegraph/jsonrpc2 →
  gorilla/websocket` (the LSP transport pulls it in via powernap). It is parent-driven —
  its version is pinned by the `charmbracelet/x/powernap` parent and must be bumped by a
  future powernap update, not independently force-bumped here. Skip.
- **Any major-version (vN+1) upgrade** — none is available anyway, and a major bump would
  carry API-breaking risk that is out of scope for a maintenance update.
- **New dependencies** — no new module is added to the `require` block. This SPEC only
  updates versions of modules already present.
- **Source-code refactors** — no `.go` file is modified unless a `golang.org/x/*` minor
  bump's API strictly requires it (REQ-GDU-007); expected: none, since no major bumps are
  in scope. Any such edit is noted in progress.md rather than expanded into a refactor.
- **Dependency-update policy docs / Dependabot / scheduled govulncheck CI** — no policy
  documentation, Dependabot configuration, or recurring CI gate is created (minimal Tier S
  scope).
