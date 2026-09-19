---
id: SPEC-GOBIN-GOTOOLCHAIN-001
title: "Pin GOTOOLCHAIN=local on the gobin resolver's go env subprocesses"
version: "0.1.0"
status: in-progress
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/runtime/gobin"
lifecycle: spec-anchored
tags: "gobin, gotoolchain, subprocess, init, resolver"
tier: S
---

# SPEC-GOBIN-GOTOOLCHAIN-001 — Pin GOTOOLCHAIN=local on the gobin resolver's `go env` subprocesses

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-09-20 | 0.1.0 | Initial plan-phase authoring (card t969), from the defect measured in card t964 | manager-spec |

## 1. Context

`internal/runtime/gobin/resolver.go` is the single source of truth for the user's Go bin
directory (`@MX:ANCHOR fan_in=2`). Its fallback chain begins with two subprocess calls,
`go env GOBIN` and `go env GOPATH`, spawned via `exec.Command` with no `Env` override — so
each child inherits the parent environment, `GOTOOLCHAIN` included. Both call sites are in
the unexported helpers `goEnvGOBIN` and `goEnvGOPATHBin`; both verified present at this HEAD.

Its two callers are `detectGoBinPath` (`internal/core/project/initializer.go`, the
`moai init` path) and `detectGoBinPathForUpdate` (`internal/cli/update.go`, the
`moai update` path) — both verified present at this HEAD. Every `moai init` therefore
spawns `go env` twice.

Where the PATH `go` is older than the module's `go` directive, that child `go` resolves a
toolchain itself and downloads a full toolchain into whatever `HOME` points at. Card t964
measured the consequences (`.moai/reports/t964/verdict.md` — prior measurement, not
re-measured here): the module cache written that way is read-only (`dr-xr-xr-x`), which
defeats `t.TempDir` cleanup and fails the test with
`TempDir RemoveAll cleanup: unlinkat .../go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64/LICENSE: permission denied`;
each firing leaves 231 MB that nothing removes; and in a user's environment `moai init`
silently downloads a toolchain.

The firing condition is the **invocation form**, not the environment. Under `go test` the
runtime prepends the resolved toolchain's `bin` to the test process's PATH, so the child
`go` is already resolved and nothing downloads. Under a precompiled test binary run
directly, that PATH injection is absent, the child `go` is the PATH `go`, and it resolves
and downloads. Card t964's control pair, same tree and same test: precompiled binary
`--- FAIL (8.07s)`, `go test` `--- PASS (0.36s)`. Every test invocation in
`.github/workflows/` is `go test`, so CI structurally cannot observe this defect — which is
why it surfaces first in a user's environment, exactly where `moai init` runs.

## 2. Requirements (GEARS)

- **REQ-GGT-001** (Ubiquitous) — The gobin resolver shall execute its `go env GOBIN` and
  `go env GOPATH` subprocesses with `GOTOOLCHAIN=local` in the child environment.

- **REQ-GGT-002** (Ubiquitous) — The gobin resolver shall preserve the rest of the parent
  process environment when overriding `GOTOOLCHAIN`, so that `HOME`, `PATH`, `GOPATH`, and
  `GOBIN` continue to reach the child unchanged.

- **REQ-GGT-003** (Event-driven) — When `gobin.Detect` is invoked, the resolver shall not
  cause any Go toolchain to be downloaded into the caller's `HOME`.

- **REQ-GGT-004** (Ubiquitous) — The gobin resolver shall leave its four-step fallback chain
  (`go env GOBIN` → `go env GOPATH`/bin → `$HOME/go/bin` → platform last resort) unchanged in
  order, in return values, and in error handling.

- **REQ-GGT-005** (Ubiquitous, unwanted) — The gobin package shall not acquire a
  characterization test whose guarding assertion has never been observed to fail against the
  unfixed resolver.

- **REQ-GGT-006** (Where — capability gate) — Where a toolchain-download assertion is observed, the `gobin` package shall be exercised through a precompiled test binary run directly, because the `go test` invocation form cannot produce the condition (see §1).

## 3. Acceptance Criteria

Acceptance criteria are enumerated in `acceptance.md` (AC-GGT-001 .. AC-GGT-006) as
Given-When-Then scenarios. They are the binding verification layer for §2.

## 4. Exclusions — what is out of scope

This SPEC's change is two lines plus one import. Everything below is deliberately excluded.

### Out of Scope — removing the subprocess entirely (R2)

- Deriving GOBIN/GOPATH from the `GOBIN` / `GOPATH` environment variables plus the
  `$HOME/go` default — which the resolver already carries as its step-3 fallback — instead of
  shelling out at all. Recorded here as a **follow-up candidate only**; it is NOT approved,
  and this SPEC carries no requirement and no acceptance criterion for it.
- Any reasoning about whether R2 subsumes R1. R1 lands on its own terms.

### Out of Scope — behavior of the fallback chain

- Any change to the order, the return values, or the error handling of the four-step chain.
- Any change to what `gobin.Detect` returns for a given environment.
- Any change to its two callers (`detectGoBinPath`, `detectGoBinPathForUpdate`).

### Out of Scope — other subprocess sites

- Any other `exec.Command` site in the repository, whether or not it shows the same
  `GOTOOLCHAIN` exposure. This SPEC binds `internal/runtime/gobin/resolver.go` only.

### Out of Scope — test surface beyond the characterization test

- Any test outside the single new characterization test in `internal/runtime/gobin/`.
- Any change to CI workflow files, including adding a precompiled-binary invocation to CI.
  The precompiled form is an acceptance-verification form for this SPEC, not a CI change.

## 5. Constraints

- **C-1** — The change is confined to `internal/runtime/gobin/`. No implementation file
  outside that directory is modified.
- **C-2** — `go vet` and `golangci-lint` must be clean on the changed package.
- **C-3** — The new test must bind `HOME` to `t.TempDir()`; it must not write to the real
  `HOME`, the repository tree, or the shared module cache.
- **C-4** — RED must be observed before GREEN (REQ-GGT-005), and observed through the
  precompiled-binary form (REQ-GGT-006).
