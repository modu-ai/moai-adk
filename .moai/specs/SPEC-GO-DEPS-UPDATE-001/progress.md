# SPEC-GO-DEPS-UPDATE-001 — Progress

## Status: implemented (sync-phase)

| Phase | State | Commit | Notes |
|-------|-------|--------|-------|
| Plan | done | (plan commit) | spec.md + plan.md + acceptance.md + progress.md authored. Tier S. status: draft. |
| Run | done | M1 (this commit) | Phase 1 (patch) + Phase 2 (x/* minor) + go mod tidy. main-direct, no PR. status: in-progress. |
| Sync | done | (sync commit) | CHANGELOG ### Changed entry + status implemented (orchestrator-direct, manager-docs trailer). |
| Mx | pending | — | 4-phase close after run. |

## Plan-phase summary

- **Phase 1 (patch)**: `go get -u=patch ./...` → powernap v0.1.6, validator/v10 v10.30.3,
  go-runewidth v0.0.24 + patch-level transitive.
- **Phase 2 (x/* minor)**: `golang.org/x/sys@v0.45.0` (direct) + indirect x/crypto v0.52,
  x/net v0.55, x/tools v0.45, x/mod v0.36.
- **Phase 3 EXCLUDED**: ALL charmbracelet/x/* indirect modules deferred (update-available
  subset = conpty/xpty/errors/exp/golden; exclusion covers the full set, not just these) —
  parents (bubbletea/huh/...) already latest and pin these; force-bump risks pin skew.
- **0 major bumps available** → no API-breaking risk → no `.go` source change expected.

## REQ-GDU-007 source-edit log (run-phase)

> If any `golang.org/x/*` minor bump requires a `.go` source edit, record it here.
> Expected: none (no major bumps in scope).

**None.** No `.go` source file was modified by either Phase 1 or Phase 2. The
`git diff --name-only` for this run is `go.mod` + `go.sum` only (AC-GDU-007 expected
outcome met; AC-GDU-008 scope guard PASS).

## §E.2 Run-phase Evidence (M1)

Run executed in the SPEC-declared location (main checkout
`/Users/goos/MoAI/moai-adk-go`, branch `docs/glm-webtool-routing-m1-m5`), go1.26.4.

### Version deltas applied (before → after)

| Module | Kind | Before | After | Driver |
|--------|------|--------|-------|--------|
| github.com/charmbracelet/x/powernap | direct | v0.1.5 | **v0.1.6** | Phase 1 patch |
| github.com/go-playground/validator/v10 | direct | v10.30.2 | **v10.30.3** | Phase 1 patch |
| github.com/mattn/go-runewidth | direct | v0.0.23 | **v0.0.24** | Phase 1 patch |
| golang.org/x/sys | direct | v0.44.0 | **v0.45.0** | Phase 2 minor (explicit @latest) |
| golang.org/x/crypto | indirect | v0.51.0 | **v0.52.0** | Phase 1 patch / graph-required |
| golang.org/x/net | indirect | v0.53.0 | v0.54.0 | tidy-resolved minimal (see §E.2.1) |
| golang.org/x/tools | indirect | v0.44.0 | v0.44.0 | tidy-resolved minimal (see §E.2.1) |
| golang.org/x/mod | indirect | v0.35.0 | v0.35.0 | tidy-resolved minimal (see §E.2.1) |

### AC PASS/FAIL matrix

| AC | Status | Actual Output |
|----|--------|---------------|
| AC-GDU-001 | PASS | go.mod shows powernap v0.1.6, validator/v10 v10.30.3, go-runewidth v0.0.24 (3 matches, bare-pipe grep) |
| AC-GDU-002 | PASS-WITH-DEBT | x/sys v0.45.0 (direct) ✅ + x/crypto v0.52.0 (indirect) ✅ kept by tidy. x/net/x/tools/x/mod could NOT reach latest minor — see §E.2.1 (tidy authoritative vs AC-GDU-003 conflict). |
| AC-GDU-003 | PASS | second `go mod tidy` is a no-op (go.mod + go.sum byte-identical) |
| AC-GDU-004 | PASS | host `go build ./...` exit 0; `GOOS=windows GOARCH=amd64` exit 0; `GOOS=linux GOARCH=amd64` exit 0 |
| AC-GDU-005 | PASS | `go test ./...` — new-regression filter (excluding the 3 §D.1 known failures) produced EMPTY output. Pre-bump baseline run also showed 0 FAIL (the 3 known failures are flaky/parity and did not fire either run). No NEW regression. |
| AC-GDU-006 | PASS | `govulncheck ./...` post-bump: "No vulnerabilities found." — affecting third-party count 0 (baseline was 0 affecting + 13 require-only-not-called; post-bump 0/0). Floor preserved. |
| AC-GDU-007 | PASS | no `.go` file in `git diff --name-only`; expected go.mod/go.sum-only outcome met |
| AC-GDU-008 | PASS | SPEC-owned diff = `go.mod` + `go.sum` only; no `.go`, no unrelated file staged (specific-path `git add`) |
| AC-GDU-009 | PASS | `go mod edit -json` Require[].Path set identical before vs after (44 paths) — no net-new module; only versions changed |

### §E.2.1 AC-GDU-002 conflict finding (PASS-WITH-DEBT rationale)

`go mod tidy` (REQ-GDU-003 / AC-GDU-003 — clean minimal tree, mandatory) reverts any
explicit indirect bump of `golang.org/x/net`, `golang.org/x/tools`, `golang.org/x/mod`
because `go mod why` reports **"main module does not need package golang.org/x/{net,tools,mod}"**
— these packages are not in the import graph. Empirically verified: `go get x/net@v0.55.0
x/tools@v0.45.0 x/mod@v0.36.0` adds them to go.mod, but the next `go mod tidy` immediately
removes them and resolves to net v0.54.0 / tools v0.44.0 / mod v0.35.0 (the minimal version
the build list demands; x/net moved 0.53→0.54 transitively, x/tools and x/mod unchanged).

Therefore AC-GDU-002's "4 indirect x/* at latest minor" and AC-GDU-003's "second tidy is
a no-op" are **mutually unsatisfiable** for x/net/x/tools/x/mod. The SPEC body already
resolves this in spec.md favor of tidy: Edge case EC-1 ("tidy-driven resolution… is
acceptable as long as it is a natural consequence of tidy, NOT a force-bump") and plan.md
Risk F ("transitive churn from `go mod tidy`… is normal and allowed"). Forcing the bumps
would inject unused indirect pins that violate the clean-minimal-tree invariant.

Resolution applied: the genuinely-needed minor bumps were kept (x/sys v0.45.0 direct,
x/crypto v0.52.0 indirect); the three unused indirect libs stay at the tidy-minimal
version. The vuln floor (AC-GDU-006) and cross-platform build (AC-GDU-004) are both
unaffected by this. The acceptance.md AC-GDU-002 "Expected" column over-asserts what is
mechanically achievable given AC-GDU-003 — flagged to the orchestrator as a debt item;
no SPEC body change made by this agent (run-phase ownership boundary).

## Mode Selection

- Decision: sub-agent (Mode 5) — coding/mechanical single-domain run-phase, Tier S minimal.
- Rationale: scope is `go.mod` + `go.sum` only; sequential single manager-develop spawn is
  the default fallback (no multi-domain / no high-volume mechanical fan-out). Per
  `.claude/rules/moai/workflow/orchestration-mode-selection.md` §B default.

## §E.3 Sync-phase Audit-Ready Signal

Orchestrator-direct sync (Tier S, manager-docs trailer) per L_orchestrator_direct_sync_tier_s — active parallel-session race (GLM-WEBTOOL-ROUTING shared-branch push absorbed M1 250c93d32 onto origin/main) + L1-worktree overhead avoidance. Sync deliverable: CHANGELOG.md `### Changed` entry + spec.md `status: in-progress → implemented`. README / docs-site NOT touched — internal dependency maintenance with no user-facing API/behavior change.

```yaml
sync_complete_at: 2026-06-03
sync_commit_sha: (backfilled in Mx commit)
sync_status: implemented
changelog_entry: "### Changed — SPEC-GO-DEPS-UPDATE-001 (1 entry under [Unreleased])"
readme_docs_site_touched: false
subagent_boundary_C_HRA_008: "n/a — orchestrator-direct doc edit, 0 AskUserQuestion in scope files"
```
