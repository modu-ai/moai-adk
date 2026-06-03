# Progress — SPEC-GO-TOOLCHAIN-SEC-001

## Status: draft (plan-phase)

Tier S (minimal). Hybrid Trunk 1-person OSS → main-direct, no PR.

## Phase Tracker

| Phase | Status | Owner | Commit |
|-------|--------|-------|--------|
| Plan | done | manager-spec | (this plan-phase commit) |
| Run (M1–M3) | pending | manager-develop | — |
| Sync (M4) | pending | manager-docs | — |
| Mx (close) | pending | orchestrator/manager-docs | — |

## Plan-phase notes

- SPEC ID self-check: `SPEC | GO | TOOLCHAIN | SEC | 001 → PASS` (canonical regex
  `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; digit-only end anchor satisfied).
- Ground-truth verified at plan time:
  - `go.mod:3` = `go 1.26` (no patch, no toolchain directive)
  - installed toolchain = go1.26.0
  - 9 hardcoded `go-version: "1.26"` pins (claude.yml:42, ci.yml:87/179/205/250/292,
    codeql.yml:85, release-pr-multi-os.yml:59, release.yml:31)
  - 3 `go-version-file` users (template-neutrality-check:52, spec-lint:15,
    spec-status-auto-sync:25) — no edit needed
- Design decisions (resolved in plan.md): CI pin strategy = (b) migrate 9 → `go-version-file`
  (SSOT); toolchain acquisition = explicit `toolchain go1.26.4` directive + GOTOOLCHAIN=auto.

## Run-phase evidence (to be filled by manager-develop)

- [ ] `go version` (must show ≥ go1.26.4 before trusting govulncheck)
- [ ] `govulncheck ./...` output (expect 0 affecting)
- [ ] `go build ./...` exit code
- [ ] `go test ./...` exit code
- [ ] `git diff go.mod` (directive-only)
- [ ] `grep -rn 'go-version: *"1.26"' .github/workflows/` (expect 0 matches)
