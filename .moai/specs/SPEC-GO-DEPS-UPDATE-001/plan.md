# SPEC-GO-DEPS-UPDATE-001 — Implementation Plan (Tier S)

> Tier S minimal plan. Sibling reference: SPEC-GO-TOOLCHAIN-SEC-001 established the
> go1.26.4 baseline and the clean (0 affecting) govulncheck floor that this SPEC must
> preserve.

## A. Context

- **Location**: project root `/Users/goos/MoAI/moai-adk-go` (main checkout).
- **Tier**: S (minimal). Scope is `go.mod` + `go.sum` only — well under the < 5 files /
  < 300 LOC Tier S envelope.
- **Workflow**: Hybrid Trunk 1-person OSS → **main-direct, no PR**. Run-phase commits land
  directly on the working branch and push to `origin/main` (per
  `.claude/rules/moai/workflow/git-workflow-doctrine.md` Tier S/M policy). No
  `gh pr create` step.
- **Sibling**: SPEC-GO-TOOLCHAIN-SEC-001 (status: completed) — the immediate predecessor
  that pinned the toolchain to go1.26.4 and cleared all 19 stdlib vulnerabilities. This
  SPEC builds on that clean baseline.
- **PRESERVE**: the `go` / `toolchain` directives in `go.mod` (owned by the sibling SPEC) —
  do NOT change the Go version. EXTEND: only the third-party `require` block versions.

## B. Known Issues (filtered for Tier S)

- **B8 Working Tree Hygiene** — the working tree has unrelated pre-existing changes
  (`.claude/settings.json`, `.moai/config/sections/statusline.yaml`,
  `.moai/config/sections/user.yaml`, a deleted
  `.moai/specs/SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001/progress.md`, untracked
  `.moai/design/web-console-handoff/` + `.moai/docs/harness-delivery-strategy.md`). The
  run-phase commit MUST use **specific-path `git add`** of only `go.mod` + `go.sum` (+ the
  SPEC progress.md). These unrelated paths MUST NOT be staged.
- **B10 Scope Discipline** — touch ONLY `go.mod` / `go.sum`. Do NOT modify charmbracelet/x/*
  indirect pins manually (Phase 3 is excluded — see spec.md Out of Scope).
- **B9 main-direct commit** — Conventional Commits format; `Authored-By-Agent:
  manager-develop` trailer on run-phase commits; `--no-verify` prohibited.

## C. Pre-flight (run-phase baseline capture)

```bash
# Baseline before any bump (to distinguish NEW vs pre-existing test failures)
go build ./... && echo "host build OK"
go test ./... 2>&1 | tail -20   # capture pre-bump baseline incl. §E.1 known failures
govulncheck ./... 2>&1 | tail -5   # confirm affecting third-party count == 0 baseline
go mod edit -json | jq '.Require | length'   # require-block size baseline
```

## D. Constraints (DO NOT VIOLATE)

- PRESERVE the `go` + `toolchain` directives in `go.mod` (sibling-owned). Version stays
  go1.26.4.
- Do NOT stage the unrelated working-tree changes listed in B8.
- Do NOT manually bump any `charmbracelet/x/*` indirect module (Phase 3 excluded).
- Do NOT add `gorilla/websocket` or any new module to `require`.
- Do NOT use `--no-verify` / `--amend` / force-push.
- Diff limited to `go.mod` + `go.sum` (+ SPEC artifacts). Any `.go` edit (expected none)
  must be recorded in progress.md.

## E. Milestones (priority-ordered, no time estimates)

| Milestone | Action | Verifies |
|-----------|--------|----------|
| M1 | **Phase 1 (patch)**: `go get -u=patch ./...` then `go mod tidy`. Confirm powernap v0.1.6 / validator v10.30.3 / go-runewidth v0.0.24 in `go.mod`. | AC-GDU-001, AC-GDU-003 (partial) |
| M2 | **Phase 2 (x/* minor)**: `go get golang.org/x/sys@v0.45.0` + bump indirect `golang.org/x/crypto`, `golang.org/x/net`, `golang.org/x/tools`, `golang.org/x/mod` to latest minor; then `go mod tidy`. | AC-GDU-002, AC-GDU-003 |
| M3 | **Verify build**: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `GOOS=linux GOARCH=amd64 go build ./...` all exit 0. | AC-GDU-004 |
| M4 | **Verify tests + vuln + scope**: `go test ./...` (excluding §E.1 known failures), `govulncheck ./...` affecting == 0, `git diff --name-only` == {go.mod, go.sum}, `require`-block has no net-new module, second `go mod tidy` is a no-op. | AC-GDU-005, AC-GDU-006, AC-GDU-007, AC-GDU-008, AC-GDU-003 (final) |
| M5 | **Commit + push** (main-direct): `git add go.mod go.sum` + SPEC progress.md; Conventional Commit; push to origin. | run-phase close |

## F. Risks

- **Low**: transitive churn from `go mod tidy` may touch more `go.sum` lines than expected
  — this is normal and allowed (AC-GDU-008 only forbids net-new `require` modules, not
  `go.sum` line churn).
- **Very low**: a `golang.org/x/*` minor bump could in principle require a source edit —
  mitigated by REQ-GDU-007 (permitted + recorded). Expected: none, since the x/* libs are
  API-stable across minors and zero major bumps are in scope.

## G. Anti-Patterns to avoid

- Manually editing `charmbracelet/x/*` indirect versions (pin-skew risk — Phase 3 excluded).
- Staging the unrelated working-tree changes (B8).
- Running `go get -u ./...` (un-suffixed) which would pull minor/major bumps beyond the
  patch + explicit-x/* scope.

## H. Cross-References

- `acceptance.md` — full 8-AC matrix (SSOT for the tester gate).
- SPEC-GO-TOOLCHAIN-SEC-001 — sibling that established the go1.26.4 + 0-affecting baseline.
- `.claude/rules/moai/workflow/git-workflow-doctrine.md` — Hybrid Trunk main-direct policy.
