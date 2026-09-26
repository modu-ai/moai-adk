---
id: SPEC-CLI-TEST-TIMEOUT-001
title: "Explicit go test timeouts on every sanctioned local test entry point"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: GOOS
priority: P1
phase: "v3.2.0"
module: "Makefile"
lifecycle: spec-anchored
tags: "testing, go-test, timeout, makefile, local-verification, internal-cli"
tier: M
related_specs: [SPEC-HEAVY-TEST-SLOT-001]
---

# SPEC-CLI-TEST-TIMEOUT-001 — Explicit go test timeouts on every sanctioned local test entry point

## History

- 2026-09-26 — v0.1.0 — Initial plan-phase draft (manager-spec, card t1253). Baseline attributed to develop `b4f798dccc8b2cf3951edea5f62e2b34740f633c`, measurement record `.moai/reports/t1253/measure-meta.txt`.

## §A Background and Problem Statement

`go test` applies a default per-binary timeout of 10 minutes. `./internal/cli/` exceeds that
default on local (shared, loaded) machines, which kills lane verification and merge-window
re-measurement. CI already works around the structural half of the problem:

- `.github/workflows/ci.yml` test-race job: `go test -json -race -count=1 -timeout 20m ./...`
  with the comment "Go's 10m default kills internal/cli, whose -race runtime measures 546-1183s".
- `.github/workflows/release-pr-multi-os.yml`: `-race -timeout 25m ./...`.

The LOCAL surfaces still relying on the 10m default are the defect class this SPEC closes. The
plan-audit round (iter-1, FAIL 0.83) widened the surface set: `make ci-local` →
`scripts/ci-mirror/run.sh` → `scripts/ci-mirror/lib/go.sh:25` also runs
`go test -race -count=1 -short ./...` with NO `-timeout` — a help-listed first-class local
entry point in the same defect class. The complete audited inventory of every go-test-carrying
local surface is §B.3; this SPEC's closure claim covers every COVERED row of that table.

### §A.1 Measured baseline (attributed)

All figures attributed in `.moai/reports/t1253/measure-meta.txt` (measured on HEAD
`b4f798dccc8b2cf3951edea5f62e2b34740f633c`, branch `WT-cli-test-duration`, go1.26.8
darwin/arm64, 16 CPU, env-scrubbed in one compound call, slot-serialized via
`moai slot go-test-heavy`):

| Figure | Value |
|---|---|
| Local no-race `./internal/cli/` package elapsed | 1118.093s, exit 0 (7594 pass / 50 skip) |
| Wall including ~38s warm-cache build | 1156s |
| CI no-race `internal/cli` (ubuntu-latest, run 36228023389) | 308.601s |
| CI race `internal/cli` (ubuntu-latest, same run) | 885.287s |
| Local-vs-CI load amplification (no-race) | 3.62x |
| Top-25 slowest tests | 278.6s = 25% of package elapsed (`.moai/reports/t1253/aggregate-top25.txt`) |
| Tests ≥1s / ≥5s / ≥10s | 399 / 42 / 7 (slowest: TestHomeStateStartVsMigrateSerialized 34.18s) |
| Package scale | 662 test files, 4,246 test funcs, 184,704 test lines, 273 non-test files |
| t.Parallel adoption | 552 of 4,246 funcs (~13%), parallel factor 1.29x |

### §A.2 Incident record

- t1171 — 601s timeout panic: `./internal/cli/` run with the default timeout crossed 600s.
- t1232 — 971s and 1597s full runs under load (exit 0, no timeout flag involved).
- 2026-09-26 — fresh attributed measurement 1118.093s package elapsed (§A.1).

## §B Requirements

**Definition — sanctioned local go test entry point.** A surface is *sanctioned* when it is
(a) a help-listed Makefile target (`.PHONY` member carrying a `##` help string) whose recipe
invokes `go test`, (b) a `go test` command line prescribed by CLAUDE.local.md, or (c) a script
invoked by such a target (for example `scripts/ci-mirror/`) whose own body invokes `go test`.
Sanctioned surfaces MUST carry an explicit `-timeout` (with a recorded derivation) or appear
as an explicitly-excluded row in the §B.3 inventory with a one-line rationale. Surfaces
outside this set are inventoried for completeness but are not obligated by this SPEC.

### §B.1 Timeout-value derivations (normative — each value carries its derivation)

**D1 — Makefile `test`, `test-verbose`, `test-race-short`, and `ci-local` (via `scripts/ci-mirror/lib/go.sh:25`) → `-timeout 60m`.**
The four race targets run `./...` whose dominant cost is `internal/cli` under `-race`.

- CI race `internal/cli` = 885.287s (run 36228023389, ubuntu-latest).
- Local load amplification (no-race basis) = 1118.093s / 308.601s = 3.62x.
- Projected local race worst case: 885.287s × 3.62 ≈ **3205s ≈ 53.4 min**.
- Cross-check via the CI race/no-race ratio (2.868x) applied to the local no-race
  measurement: 1118.093s × 2.868 ≈ **3207s** — convergent at ~3.2ks.
- A 30m value (1800s) would NOT clear a loaded local machine (53.4 min projected).
  **Candidate adopted: 60m (3600s)** — 1.12x headroom over the projected worst case.
  The headroom is thin by construction; a future measured local race run exceeding 60m
  re-triggers this derivation (see REQ-TIMEOUT-006).

**D2 — CLAUDE.local.md §4 / §6 package recipe → `-timeout 30m`.**
The recipe runs ONE package at a time (`go test ./internal/<pkg>/...`); the worst known
package is `internal/cli` at 1118.093s measured local no-race under load.

- 30m = 1800s = **1.61x headroom** over the measured single-package worst case.
- Sufficient because the recipe excludes full-suite fan-out by design (the 2026-08-15
  load-413 incident rule, CLAUDE.local.md §6).

**D3 — Makefile `test-codex-live` → `-timeout 10m`.**
Scope-justified smaller value: the target runs a single `-run 'Live'` selector subset,
opt-in, and spends real codex/z.ai quota. **No per-subset duration measurement exists
(disclosed gap)**; the observed Live axis is a small fraction of package elapsed. The
explicit 10m pins the semantics against future default drift and bounds hang exposure on
a quota-spending run — it is an explicit pin, not a derived ceiling.

### §B.2 GEARS requirements

- REQ-TIMEOUT-001 — The Makefile `test` target shall invoke `go test` with the explicit flag `-timeout 60m` (derivation D1).
- REQ-TIMEOUT-002 — The Makefile `test-verbose` target shall invoke `go test` with the explicit flag `-timeout 60m` (derivation D1).
- REQ-TIMEOUT-003 — The Makefile `test-race-short` target shall invoke `go test` with the explicit flag `-timeout 60m` (derivation D1; `-short` skips a subset of slow tests but no short-mode duration measurement exists, so the race derivation is retained rather than narrowed).
- REQ-TIMEOUT-004 — The Makefile `test-codex-live` target shall invoke `go test` with the explicit flag `-timeout 10m` (derivation D3).
- REQ-TIMEOUT-005 — The CLAUDE.local.md §4 Before-Commit recipe and the §6 [HARD] Go Test Execution Rules recipe shall prescribe `go test -timeout 30m ./internal/<pkg>/...` (derivation D2). The §6 edit is scoped to the package-test recipe lines; the full-suite command lines in §6 belong to card t1219 (see REQ-COORD-009).
- REQ-TIMEOUT-006 — When a future measured local race run of any `./...` target exceeds its explicit `-timeout`, the timeout value shall be re-derived from a fresh attributed measurement rather than raised ad hoc.
- REQ-DOC-007 — Each explicit `-timeout` value introduced by this SPEC shall be accompanied in the modified file by a comment naming the derivation (D1/D2/D3) and the baseline record `.moai/reports/t1253/measure-meta.txt`, so no bare constant reaches the tree.
- REQ-SCOPE-008 — This SPEC shall not modify `.github/workflows/ci.yml` or `.github/workflows/release-pr-multi-os.yml`; both already carry explicit timeouts (unwanted-behavior requirement; CI is out of scope by design).
- REQ-COORD-009 — The SPEC shall record the coordination premises with t1252 (file-disjoint) and t1219 (same-file merge-order adjacency, including the §13 full-suite mention ownership) before run-phase entry (§D).
- REQ-TIMEOUT-010 — The `scripts/ci-mirror/lib/go.sh` test step (reached via `make ci-local`) shall invoke `go test` with the explicit flag `-timeout 60m` (derivation D1; the `-short`-mode unmeasured disclosure of §C.3 applies identically).
- REQ-DOC-011 — The SPEC shall maintain a complete inventory (§B.3) of every go-test-carrying local surface, where each row is either COVERED with an explicit `-timeout` value plus derivation, or EXCLUDED with a one-line rationale.
- REQ-DOC-012 — The SPEC shall record its considered-and-rejected alternatives (§C) with the quantified evidence behind each rejection and the follow-up-card disposition.

### §B.3 Complete entry-point inventory (every go-test-carrying local surface)

Line numbers refer to the baseline tree (HEAD `b4f798dcc`). Status values: COVERED = this
SPEC adds an explicit `-timeout`; EXCLUDED = no change, rationale stated; TRANSITIVE = no
`go test` invocation of its own; EXTERNAL-EXPLICIT = already explicit, out of scope.

| # | Surface | go test invocation | Status | Value / rationale |
|---|---------|--------------------|--------|-------------------|
| 1 | Makefile `test` (L104) | `-race -coverprofile… -covermode=atomic ./...` | COVERED | 60m (D1) |
| 2 | Makefile `test-verbose` (L107) | `-race -v -coverprofile… ./...` | COVERED | 60m (D1) |
| 3 | Makefile `test-race-short` (L194) | `-race -short ./...` | COVERED | 60m (D1; `-short` unmeasured disclosure, §C.3) |
| 4 | Makefile `test-codex-live` (L110) | `./internal/cli/ -run 'Live' -v -count=1` | COVERED | 10m (D3) |
| 5 | Makefile `ci-local` → `scripts/ci-mirror/lib/go.sh:25` | `-race -count=1 -short ./...` | COVERED | 60m (D1; same `-short` disclosure as row 3) — added at plan-audit iter-1 |
| 6 | Makefile `coverage` (L112) | none — invokes `go tool cover`; depends on `test` | TRANSITIVE | inherits row 1's flag transitively |
| 7 | Makefile `tui-snapshot` (L180) | `UPDATE_GOLDEN=1 go test ./internal/tui/... ./internal/tui/golden/... -v` | EXCLUDED | golden-regen on `internal/tui`, a non-cli package with no measured incident in the timeout class; an unmeasured invented value is prohibited (REQ-TIMEOUT-006) |
| 8 | Makefile `tui-snapshot-verify` (L184) | `go test ./internal/tui/... ./internal/tui/golden/... -v -count=1` | EXCLUDED | same rationale as row 7 |
| 9 | Makefile `agents-emit` / `agents-emit-check` (L39/L48) | `go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` | EXCLUDED | single `-run` golden selector on a small non-cli template package; not in the internal/cli defect class |
| 10 | Makefile `commands-emit` / `commands-emit-check` (L52/L59) | same shape on `./internal/template/commandemit/...` | EXCLUDED | same rationale as row 9 |
| 11 | Makefile `tool-policy-drift-check` (L67) | `go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_…' -count=1` | EXCLUDED | single `-run` selector on a small config package; not in the defect class |
| 12 | `scripts/ac-baseline/check-staged.sh:26` | `MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1` | EXCLUDED | single-test selector inside the AC-snapshot commit-guard path; small package; not in the defect class |
| 13 | `.github/workflows/ci.yml` + `release-pr-multi-os.yml` | `-race -timeout 20m ./...` / `-race -timeout 25m ./...` | EXTERNAL-EXPLICIT | already explicit (REQ-SCOPE-008); CI load regime differs from local |
| 14 | CLAUDE.local.md §4 Before-Commit recipe (L265) | `go test ./internal/<pkg>/...` | COVERED | 30m (D2) |
| 15 | CLAUDE.local.md §6 [HARD] package rule (L394) | `go test ./internal/<pkg>/...` | COVERED | 30m (D2) |
| 16 | CLAUDE.local.md §6 full-suite lines (L396, L397) | `go test -count=1 ./...` / `go test -race ./...` | EXCLUDED | full-suite command lines owned by t1219's contradiction-cleanup scope (§D); not edited here |
| 17 | CLAUDE.local.md §13 GLM-testing mention (L532) | `go test ./...` | EXCLUDED | context prose inside the GLM-integration-testing ban, not a prescribed recipe; ownership assigned to t1219 (§D) |
| 18 | CLAUDE.local.md §2.0 prose rows (L161, L164) | `go test ./internal/template/agentemit/...` cited as CI-executed | EXCLUDED | descriptive prose about checks CI runs, not a local prescription; the underlying target is row 9's surface |

## §C Documented Rejections (considered and rejected, with evidence)

1. **Package split of `internal/cli`** — REJECTED for this SPEC. Cost is Tier L (662 test
   files / 4,246 funcs / 184,704 test lines / 273 source files, 13 subpackages already
   extracted); urgency is low because CI no-race measures 308.601s — about half the 600s
   default ceiling, structural but not emergency. Follow-up-card candidate. Out of scope here.
2. **Slow-test repair** — REJECTED for this SPEC. The top-25 cap is ~25% of package elapsed
   (278.6s of 1118.093s) and the distribution is broad (399 tests ≥1s), so repair cannot fix
   the timeout class and yields at most ~25%. Follow-up-card candidate. Out of scope here.
3. **Narrowing `test-race-short` below the D1 derivation** — REJECTED. No `-short`-mode
   duration measurement exists; adopting a smaller unmeasured value would trade a measured
   derivation for a guess.

## §D Coordination Premises

- **t1252 (in flight at plan-phase authoring)** — fixes `internal/cli` TestMain (`internal/cli/main_test.go`).
  No file overlap with this SPEC (Makefile + CLAUDE.local.md only). The measurement baseline is
  pre-t1252 develop `b4f798dcc`; t1252 may shift absolute durations but not the structural
  conclusion.
- **t1219 (queued)** — covers CLAUDE.local.md contradiction cleanup including the §6
  full-suite command lines. This SPEC's §6 edit touches the `[HARD]` package-test recipe lines
  only (line "run the AFFECTED packages (`go test ./internal/<pkg>/...`)"). Merge-order note:
  whoever merges second re-reads the §6 region and absorbs the other card's edits; no semantic
  conflict is expected because the line sets are disjoint.
- **t1219 / CLAUDE.local.md §13 (added at plan-audit iter-1)** — §13 carries a further
  full-suite mention (line ~532: "Unit tests: dev project (`go test ./...`)") whose ownership
  is unstated. This SPEC does NOT touch §13: that line is context prose inside the
  GLM-integration-testing ban, not a sanctioned recipe (it falls outside the §B sanctioned
  definition). Ownership of its wording is assigned to t1219's contradiction-cleanup scope;
  recorded here so the same-file merge-order hazard has a named owner for every full-suite
  mention.

## §E Out of Scope

### Out of Scope — CI workflow files

- `.github/workflows/ci.yml` and `.github/workflows/release-pr-multi-os.yml` are NOT modified; they already carry explicit `-timeout 20m` / `-timeout 25m` (REQ-SCOPE-008).

### Out of Scope — internal/cli structural change

- No package split, subpackage extraction, or test-file relocation of `internal/cli` (Tier L cost; follow-up-card candidate).

### Out of Scope — slow-test optimization

- No per-test duration repair, `t.Parallel` adoption campaign, or test rewriting (top-25 caps at ~25%; follow-up-card candidate).

### Out of Scope — template distribution

- No change under `internal/template/templates/**`: the Makefile and CLAUDE.local.md touched here are repo-local developer surfaces, not template-managed files distributed by `moai update`.

## §F Non-Functional Constraints

- Values are plain `make`/shell flags; no new dependency, no Go code change, no runtime behavior change.
- Comments added to the Makefile must not break `make help` extraction (keep `##` help-text convention intact).

## §G Acceptance Criteria

The AC enumeration lives in `acceptance.md` (§D AC Matrix). Summary: 9 criteria, all
binary-testable; AC-001/AC-002 are RED-now on the pre-change tree (no `-timeout` present on
any listed surface today — verified by the same grep the AC specifies).
