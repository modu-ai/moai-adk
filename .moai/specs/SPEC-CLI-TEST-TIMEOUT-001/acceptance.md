---
id: SPEC-CLI-TEST-TIMEOUT-001
title: "Acceptance Criteria — Explicit go test timeouts on local entry points"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
---

# acceptance.md — SPEC-CLI-TEST-TIMEOUT-001

All criteria are binary-testable. AC-001 and AC-002 are RED-now on the pre-change tree:
the same greps that verify them return no `-timeout` occurrences on any listed surface
today (verified against HEAD `b4f798dcc`; the failure observation is re-executable at
run-phase M1 entry per §C.3 of plan.md). AC-004's green path is milestone M3.

## §D AC Matrix

### AC-001 — Sanctioned local entry points carry explicit -timeout

**Given** the post-change tree **When** each COVERED surface of the §B.3 inventory is
inspected — the Makefile targets `test`, `test-verbose`, `test-race-short`,
`test-codex-live` via `make -n <target>` (or the recipe line read directly), and the
`scripts/ci-mirror/lib/go.sh` test step via the script body (reached through
`make ci-local`) **Then** every one of those `go test` invocations carries an explicit
`-timeout` flag — `60m` for the three `./...` Makefile race targets and the ci-mirror step,
`10m` for `test-codex-live` — and none relies on the go-test default. A mutant that removes
any one flag fails this AC (the flag extraction yields 4 of 5).

### AC-002 — CLAUDE.local.md §4/§6 recipes carry -timeout 30m

**Given** CLAUDE.local.md **When** the §4 Before-Commit checklist line and the §6 [HARD]
"run the AFFECTED packages" line are inspected **Then** both prescribe
`go test -timeout 30m ./internal/<pkg>/...` (flag position flexible; the
`-timeout 30m` token pair must be present), and the §6 full-suite lines
(`-count=1 ./...`, `-race ./...`) are byte-unchanged.

### AC-003 — Derivations recorded verifiably in SPEC text

**Given** spec.md §B.1 **When** the D1/D2/D3 derivations are checked arithmetically
**Then** D1: 885.287 × 3.62 ≈ 3205s and the cross-check 1118.093 × 2.868 ≈ 3207s agree at
~3.2ks, with 30m (1800s) shown insufficient and 60m (3600s) adopted at 1.12x headroom;
D2: 1800 / 1118.093 ≈ 1.61x; D3 discloses the unmeasured-subset gap explicitly. Every
source figure carries its attribution (measure-meta.txt / run 36228023389).

### AC-004 — Post-change bounded re-verification

**Given** the post-M1/M2 tree **When** an env-scrubbed
`unset <factory,kanban,GIT_* env> && go test -json -count=1 -timeout 35m ./internal/cli/`
runs (slot-serialized, single compound invocation) **Then** the run completes with exit 0,
the output stream contains no `panic: test timed out`, and all leaf tests are green
(7594 pass baseline; t1252 may lawfully shift the count — any change is reported, not
assumed). Evidence path recorded under `.moai/state/verify/`.

### AC-005 — CI workflows untouched

**Given** the working tree **When** `git diff --stat` (and `git status --porcelain`) is
read **Then** `.github/workflows/ci.yml` and `.github/workflows/release-pr-multi-os.yml`
appear in no changed-path list, and the changed set is exactly `Makefile`,
`scripts/ci-mirror/lib/go.sh`, `CLAUDE.local.md`, and
`.moai/specs/SPEC-CLI-TEST-TIMEOUT-001/**`.

### AC-006 — Coordination premises recorded

**Given** spec.md §D **When** read **Then** all three premise elements are present: t1252
named with its file-disjointness claim (`internal/cli/main_test.go` vs the three target
files) and the baseline-pre-t1252 caveat; t1219 named with the merge-order note that this
SPEC's §6 edit is the package-test recipe lines only, disjoint from t1219's full-suite
lines; and the §13 element — CLAUDE.local.md line ~532's full-suite mention with its
ownership assigned to t1219 and an explicit this-SPEC-does-not-touch statement.

### AC-007 — Documented rejections present with evidence

**Given** spec.md §C **When** read **Then** both rejections (package split; slow-test
repair) are recorded with their quantified evidence (Tier L scale figures; top-25 = 25%
cap) and each names its follow-up-card disposition, plus the third rejection
(narrowing `test-race-short` below D1 for lack of a `-short` measurement).

### AC-008 — No bare timeout constant

**Given** the modified Makefile and CLAUDE.local.md **When** each introduced `-timeout`
occurrence is inspected **Then** a derivation pointer (D1/D2/D3 and/or
`.moai/reports/t1253/measure-meta.txt`) is present adjacent to it in the same file
(comment or inline note). A mutant that keeps the flag but strips the pointer fails
this AC.

### AC-009 — Inventory completeness

**Given** spec.md §B.3 **When** the inventory table is cross-checked against every
`go test` invocation discoverable in the Makefile, the scripts its help-listed targets
invoke (`scripts/ci-mirror/`, `scripts/ac-baseline/check-staged.sh`), every CLAUDE.local.md
go-test line (sanctioned-definition clause (b) surface), and the two CI workflow files
**Then** each invocation appears in exactly one row, and each row is either
COVERED with an explicit `-timeout` value plus a derivation pointer, or EXCLUDED /
TRANSITIVE / EXTERNAL-EXPLICIT with its stated rationale. A mutant that adds an unlisted
go-test-carrying surface (or drops a row's rationale) fails this AC. Baseline discovery
commands: `grep -n 'go test' Makefile`, `grep -rn 'go test' scripts/ci-mirror/
scripts/ac-baseline/`, `grep -n 'go test' CLAUDE.local.md`.

## §D.1 Severity

- Release-blocking: AC-001, AC-002, AC-003, AC-004, AC-005, AC-009.
- Non-blocking (documentation integrity): AC-006, AC-007, AC-008.

## §D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-001 | REQ-TIMEOUT-001, REQ-TIMEOUT-002, REQ-TIMEOUT-003, REQ-TIMEOUT-004, REQ-TIMEOUT-010, REQ-DOC-007 | M1 |
| AC-002 | REQ-TIMEOUT-005 | M2 |
| AC-003 | REQ-TIMEOUT-006 | M1 (authored) / M4 (verified) |
| AC-004 | REQ-TIMEOUT-001, REQ-TIMEOUT-002, REQ-TIMEOUT-003, REQ-TIMEOUT-004, REQ-TIMEOUT-010 | M3 |
| AC-005 | REQ-SCOPE-008 | M4 |
| AC-006 | REQ-COORD-009 | M2 |
| AC-007 | REQ-DOC-012 | M4 |
| AC-008 | REQ-DOC-007 | M4 |
| AC-009 | REQ-DOC-011, REQ-DOC-007 | M4 |

## §D.3 Edge cases

- `test-race-short` and the ci-mirror step (both `-short` race runs) — heavy tests skip,
  duration drops, but the 60m value stays on both — uniformity with the measured race
  derivation beats per-target tuning on unmeasured subsets (spec.md §C.3).
- A loaded-machine run slower than 60m: REQ-TIMEOUT-006 triggers re-derivation, not ad-hoc
  flag raising.
- t1252 lands first and shifts internal/cli duration: AC-004's expected leaf count is
  reported as observed, not asserted against the stale baseline number.

## §D.4 Quality gates

- `make -n` dry-runs show the flags (no execution needed for AC-001).
- M3 re-verification is the only execution-gated AC; it is bounded (one package, one run,
  slot-serialized) per CLAUDE.local.md §6 load discipline.
- No lint/format impact expected (Makefile recipes and markdown prose only); `make help`
  extraction intact (plan.md §E5).

## §D.5 Definition of Done

All §D.1 release-blocking ACs PASS with verbatim evidence; non-blocking ACs PASS or carry
an explicit recorded gap; `git diff` path set exactly as AC-005; `progress.md` §E.1
populated and §E.2–§E.4 left as placeholders for run/sync phases.
