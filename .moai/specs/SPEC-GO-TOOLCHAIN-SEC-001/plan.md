# Implementation Plan — SPEC-GO-TOOLCHAIN-SEC-001

> Tier S (minimal). 4 artifacts: spec.md + plan.md + acceptance.md + progress.md.
> No design.md / research.md (Tier S does not require them).

## A. Context

Bump the Go toolchain from the currently-declared `go 1.26` (resolving to installed
go1.26.0) to **go1.26.4** so that `govulncheck ./...` reports 0 affecting vulnerabilities
(19 Go-stdlib findings → 0). The change touches two version-pin surfaces: `go.mod`
(the `go` directive, plus an optional `toolchain` directive) and the 9 CI workflow steps
that hardcode `go-version: "1.26"`. Three other workflows already use
`go-version-file: go.mod` and require no edit.

## B. Known Issues / Ground-Truth (verified by orchestrator)

- `go.mod:3` = `go 1.26` (no patch, no `toolchain` directive).
- Installed local toolchain = `go1.26.0` (below the go1.26.4 target).
- 9 hardcoded `go-version: "1.26"` pins: `claude.yml:42`, `ci.yml:87/179/205/250/292`,
  `codeql.yml:85`, `release-pr-multi-os.yml:59`, `release.yml:31`.
- 3 `go-version-file` users (auto-propagate, no edit): `template-neutrality-check.yaml:52`,
  `spec-lint.yml:15`, `spec-status-auto-sync.yml:25`.
- All 19 govulncheck findings are stdlib; third-party `require` blocks are clean.

## C. Pre-flight

- `govulncheck` must be available in the run-phase environment (`go install
  golang.org/x/vuln/cmd/govulncheck@latest`) to capture the before/after evidence.
- The run-phase toolchain must be ≥ go1.26.4 — see § Design Decision 2 (toolchain
  acquisition) for how it is obtained.

## D. Constraints

- Scope-discipline guard: `git diff go.mod` after the change MUST show ONLY the
  `go`/`toolchain` directive lines changed — no `require` block line. This is an AC
  (AC-GTS-004), enforced as a closure gate.
- This repo follows **Hybrid Trunk 1-person OSS** (CLAUDE.local.md §23): Tier S →
  **main-direct push, no PR**. There is **no manager-git PR step** in this plan; the
  run-phase and sync-phase commits land directly on `main` and are pushed.
- No source-code change. If a Go API deprecation surfaces under the bumped toolchain,
  STOP and return a blocker (it would be a separate SPEC).

## D-DESIGN. Resolved Design Decisions

### Design Decision 1 — CI pin strategy (RESOLVED: option (b), migrate to `go-version-file`)

**Decision**: Migrate the 9 hardcoded `go-version: "1.26"` pins to
`go-version-file: go.mod` (SSOT), matching the 3 workflows that already do this.

**Rationale**: With `go-version-file: go.mod`, the Go version is sourced from a single
place (the `go.mod` `go` directive). After this SPEC, the repo has 12 workflow steps all
deriving from go.mod and 0 hardcoded version literals — so every future toolchain bump
touches only `go.mod` (one line), never the workflow files. This eliminates the recurring
drift class where a go.mod bump silently leaves CI on the old version.

**Trade-off considered (option (a), bump literals to `"1.26.4"`)**: Option (a) is a smaller
textual diff this one time (replace `"1.26"` → `"1.26.4"` in 9 places) and keeps each
workflow's version visible inline without opening go.mod. Its cost is that it re-creates
the same 9-site edit on the *next* bump and preserves the two-SSOT split (go.mod +
literals) that allows them to drift apart. Because the stated motivation includes making
the security floor explicit and reproducible, the SSOT consolidation of option (b) is the
better fit. Option (b) is recommended unless a concrete reason (e.g., a workflow that must
pin a different version than go.mod) favors (a) — no such reason exists here.

**Consequence for run-phase**: The 9 steps change from `go-version: "1.26"` to
`go-version-file: go.mod` (the `actions/setup-go` `go-version-file` input reads the `go`
directive from go.mod). The go.mod `go` directive bump (Design Decision 2) is then the
single source the 12 workflows derive from.

### Design Decision 2 — Toolchain acquisition (RESOLVED: explicit `toolchain` directive + GOTOOLCHAIN=auto)

**Decision**: Add an explicit `toolchain go1.26.4` directive to `go.mod` alongside the
`go 1.26.4` directive, and rely on `GOTOOLCHAIN=auto` (the Go default) to auto-download
the go1.26.4 toolchain when the local/CI toolchain is older.

**Rationale**: The local installed toolchain is go1.26.0 (below go1.26.4). With
`GOTOOLCHAIN=auto` (Go's default since 1.21) and a `toolchain go1.26.4` directive in
go.mod, `go build` / `go test` automatically download and switch to go1.26.4 — no manual
local toolchain upgrade is required for the run-phase to produce valid evidence. CI
`actions/setup-go` with `go-version-file: go.mod` likewise honors the `toolchain`
directive and provisions go1.26.4.

**Run-phase acceptance-evidence implication**: The run-phase MUST capture
`govulncheck ./...` and `go test ./...` output produced **under go1.26.4**, not go1.26.0.
The run-phase verifies the effective toolchain with `go version` (expect go1.26.4+) before
trusting the govulncheck "0 affecting" result. If `GOTOOLCHAIN=auto` is disabled in the
environment (`GOTOOLCHAIN=local`), the run-phase MUST either re-enable auto or upgrade the
local toolchain to ≥ go1.26.4 before capturing evidence — running govulncheck under
go1.26.0 would falsely still report the 19 findings.

**Trade-off considered (no `toolchain` directive, require local upgrade)**: Omitting the
`toolchain` directive and instead requiring every developer/runner to have ≥ go1.26.4
installed is simpler in go.mod but pushes an out-of-band setup burden onto each
environment and re-introduces drift (a machine on go1.26.0 silently builds against the
older stdlib). The explicit `toolchain` directive makes the required patch self-acquiring
and reproducible, which matches the security-floor motivation.

## E. Self-Verification (run-phase entry checks)

1. `go version` confirms the effective toolchain is ≥ go1.26.4 (auto-acquired) before any
   govulncheck capture.
2. `git diff go.mod` shows only the `go` (and new `toolchain`) directive lines changed.
3. `grep -rn 'go-version: *"1.26"' .github/workflows/` returns zero matches after the
   migration (all 9 converted to `go-version-file`).

## F. Milestones (priority-ordered, no time estimates)

- **M1 — go.mod toolchain bump**: Change `go 1.26` → `go 1.26.4` and add
  `toolchain go1.26.4`. Verify `git diff go.mod` touches only those directive lines
  (no `require` change). Run `go version` to confirm auto-acquired go1.26.4.
- **M2 — CI pin migration**: Convert the 9 hardcoded `go-version: "1.26"` steps to
  `go-version-file: go.mod` (claude.yml, ci.yml ×5, codeql.yml, release-pr-multi-os.yml,
  release.yml). Confirm zero residual `go-version: "1.26"` literals.
- **M3 — Verification evidence**: Capture `govulncheck ./...` (expect 0 affecting),
  `go build ./...` (pass), `go test ./...` (pass) — all under go1.26.4. Confirm
  scope-discipline guard (AC-GTS-004).
- **M4 — sync**: CHANGELOG entry (security bump), frontmatter status transition
  (draft → in-progress on M1 run commit; in-progress → implemented on sync). No
  README/docs-site change expected (internal toolchain bump). No PR (Tier S main-direct).

## G. Anti-Patterns to Avoid

- Bumping a third-party `require` version "while we're here" — scope violation
  (AC-GTS-004 guards against this).
- Capturing govulncheck "0 affecting" while the effective toolchain is still go1.26.0 —
  a false PASS. Always verify `go version` ≥ go1.26.4 first.
- Leaving any of the 9 workflows on a hardcoded literal after migration — partial
  application violates REQ-GTS-003.
- Editing the 3 workflows that already use `go-version-file` — they propagate
  automatically and need no change.
- Creating a govulncheck CI gate or Dependabot config — out of scope (minimal option).

## H. Cross-References

- spec.md § C (GEARS requirements), § E (Exclusions)
- acceptance.md (full AC matrix + closure gates)
- CLAUDE.local.md §23 (Hybrid Trunk 1-person OSS, Tier S → main-direct, no PR)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition
  Ownership Matrix (draft → in-progress = manager-develop; in-progress → implemented =
  manager-docs)
