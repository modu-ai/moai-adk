---
id: SPEC-GO-TOOLCHAIN-SEC-002
title: "Go toolchain security bump (go1.26.4 → go1.26.8, 8 stdlib vulns → 0)"
version: "0.1.2"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "go.mod"
lifecycle: spec-anchored
tags: "security, dependencies, toolchain, govulncheck, card-t610"
tier: S
related_specs: [SPEC-GO-TOOLCHAIN-SEC-001]
---

# SPEC-GO-TOOLCHAIN-SEC-002 — Go toolchain security bump (go1.26.4 → go1.26.8)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-09-10 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t610 (Factory lane-8), Tier S, Class C (global change). Precedent: SPEC-GO-TOOLCHAIN-SEC-001. status: draft. |
| 2026-09-10 | 0.1.1 | manager-spec | Card t610 plan-audit iteration-1 revision (`.moai/reports/t610/plan-audit.md`, FAIL: 3 blocking, 15 non-blocking). B1: AC-GTS2-003/007 re-anchored on the merge-base form `develop...HEAD`, plus a post-absorption re-measure in M4. B2: AC-GTS2-007 judged on the whole-tree complement. B3: the #80927 coverage claim withdrawn — `internal/web` tests are a net/http regression guard only. N1–N13 and N15 addressed; N14 kept as a recorded Gap. status: draft. |
| 2026-09-10 | 0.1.2 | manager-spec | Card t610 plan-audit iteration-2 residuals R1–R6 (`.moai/reports/t610/plan-audit-2.md`, PASS: 0 blocking, score 0.86). R1: E-03 exit corrected to 1, judged on stdout. R2: the pre-absorption evidence reuse clause removed; M4 always re-measures. R3: AC-GTS2-008 added to the M4 re-measure. R4: the D.0b coverage claim narrowed, and M4 compares the 5 sync paths' card-side numstat with the M5 sync commit. R5: EC-5 handling once status is `completed`. R6: AC-GTS2-006 CI verdict timing. R7 needed no change (the SPEC already names local `develop`). No requirement, decision, scope, or target-version change. status: draft. |

## A. Context and Problem Statement

With `go.mod` at `go 1.26.4`, `govulncheck ./...` reports **"Your code is affected by 8
vulnerabilities from the Go standard library"**. This is finding F01 of the Go full review
of 2026-09-10 (evidence item SEC-01). All 8 are in the standard library of go1.26.4. None
are in third-party `require` modules.

| ID | Package (found in go1.26.4) | Fixed in |
|----|-----------------------------|----------|
| GO-2026-6218 | net/url | go1.26.6 |
| GO-2026-6091 | html/template | go1.26.6 |
| GO-2026-6090 | crypto/tls | go1.26.6 |
| GO-2026-6089 | net/http | go1.26.6 |
| GO-2026-6088 | encoding/xml | go1.26.6 |
| GO-2026-5972 | encoding/asn1 | go1.26.6 |
| GO-2026-5856 | crypto/tls | go1.26.5 |
| GO-2026-5026 | net/http | go1.26.6 |

The highest fix version across the 8 findings is **go1.26.6**. The go1.26.4 scan in the
same run (`.moai/reports/t610/baseline/govulncheck-auto.log:87-88`) also reported "2
vulnerabilities in packages you import and 3 vulnerabilities in modules you require" that the
code does not appear to call. The go1.26.6 and go1.26.8 scans report 0 in packages you import
and still 3 in modules you require (`govulncheck-go1.26.6.log`, `govulncheck-go1.26.8.log`).

Baseline measured on base commit `d3b7d438d` (this tree). Evidence directory:
`.moai/reports/t610/baseline/`. The exact commands are recorded in
`.moai/reports/t610/baseline/commands.md`.

1. `go.mod:3` = `go 1.26.4`. The repo has a single `go.mod`, no `toolchain` directive, and
   no `go.work`.
2. Every CI workflow step that sets up Go uses `actions/setup-go@v7` with
   `go-version-file: go.mod` (ci.yml ×6, release.yml, release-pr-multi-os.yml, codeql.yml,
   spec-lint.yml, lsel-leak-guard.yaml, spec-status-auto-sync.yml, graph-freshness.yml,
   template-neutrality-check.yaml). No Go version is hard-coded in CI, the Makefile, or a
   release config (evidence: `ci-go-version-refs.txt`). SPEC-GO-TOOLCHAIN-SEC-001 set up
   this go.mod single source of truth. That means `go.mod` is now the only file that pins
   the version.
3. Locally, `/opt/homebrew/bin/go` is go1.26.0 and `GOTOOLCHAIN` is `auto`. The effective
   go1.26.4 comes from GOTOOLCHAIN auto-switch reading the `go.mod` directive. It is not an
   installed Go. Evidence: `goversion-auto.txt` holds the effective go1.26.4; the go1.26.0
   reading is recorded at `commands.md:22-24` (run from `/private/tmp`), not in a saved
   output file.
4. Toolchain-only variation of govulncheck on the same tree:

   | Toolchain | Exit | Result | Evidence |
   |-----------|------|--------|----------|
   | go1.26.4 (auto) | 3 | affected by 8 vulnerabilities | `govulncheck-auto.log` |
   | go1.26.6 | 0 | affected by 0 vulnerabilities | `govulncheck-go1.26.6.log` |
   | go1.26.8 | 0 | affected by 0 vulnerabilities | `govulncheck-go1.26.8.log` |

   The go1.26.4 row is the control. It shows the scanner does report findings when they
   exist, so the two 0 results are not a silent empty scan (exits recorded in
   `govulncheck-exits.txt`).
5. Latest stable releases: go1.26.8 and go1.27.1. Evidence: `go-release-history.html`
   carries entries `id="go1.26.8"` and `id="go1.27.1"`. The download-index query is recorded
   at `commands.md:26`, but its JSON output was not saved.

## B. Why (motivation)

- **Security exposure.** The 8 findings cover URL parsing, HTML templating, TLS, HTTP
  serving, XML and ASN.1 decoding. A CLI that makes network calls, renders templates, and
  runs a local web surface can reach all of them.
- **One declarative line closes all 8.** Each finding is a standard-library defect fixed
  in a patch toolchain release. The go.mod single source of truth is already in place, so
  the change is the `go` directive alone. No source or workflow edit is needed.
- **Global blast radius.** The directive picks the build toolchain for every lane,
  every CI job, and every release artifact. That is why the change is Class C and lands
  on its own (see plan.md § D4).

## C. Requirements (GEARS)

REQ-GTS2-001 (Ubiquitous): The `go.mod` `go` directive **shall** declare `go 1.26.8`, the
target version. go1.26.6 is the proven floor.

REQ-GTS2-002 (Ubiquitous, unwanted): The bump **shall not** change any `go.mod` line other
than the `go` directive. No `toolchain` directive is added and no `require` or `replace`
line is modified.

REQ-GTS2-003 (Ubiquitous, unwanted): The bump **shall not** modify any CI workflow,
Makefile, or Go source file. CI follows `go.mod` through `go-version-file`.

REQ-GTS2-004 (Event-driven): **When** run-phase verification evidence is captured, the
effective Go toolchain in the working tree **shall** report the target version before any
other run-phase acceptance criterion is judged.

REQ-GTS2-005 (Event-driven): **When** `govulncheck ./...` runs under the bumped `go.mod`
without any `GOTOOLCHAIN` override, the scan **shall** report zero vulnerabilities
affecting the code. None of the 8 baseline IDs in § A **shall** appear in its output.

REQ-GTS2-006 (Event-driven): **When** the binary is built with `make build`, the build
**shall** succeed, and the Go version embedded in `bin/moai` **shall** be the target
version. That output is kept as evidence.

REQ-GTS2-007 (Ubiquitous): Run-phase verification evidence **shall** be written to files
under `.moai/reports/t610/`. Each judged command's exit code is recorded directly from
that command and not through a pipeline.

REQ-GTS2-008 (Event-driven): **When** the sync phase runs, the 5 project-document mentions
of `1.26.4` listed in plan.md § F M5 and acceptance.md E-08 (evidence `doc-1264-refs.txt`)
**shall** name the target version.

## D. Acceptance Criteria Summary

The full matrix and the RED-now evidence ledger are in `acceptance.md` (8 ACs):

- AC-GTS2-001: toolchain pre-check gate. `go -C <worktree-root> version` reports go1.26.8.
  It gates AC-GTS2-002..007.
- AC-GTS2-002: `go.mod:3` is `go 1.26.8` and there is no `toolchain` line.
- AC-GTS2-003: the committed `go.mod` diff against the merge-base with `develop` is exactly
  one line removed and one added.
- AC-GTS2-004: govulncheck without an override exits 0, reports "affected by 0
  vulnerabilities", and none of the 8 IDs appears.
- AC-GTS2-005: `make build` exits 0 and `go version -m bin/moai` reports go1.26.8. The
  output is kept.
- AC-GTS2-006: regression guard. `go vet ./...` exits 0, and the tests of packages the run
  phase chooses exit 0 with a non-empty swept set. The selection includes `internal/web` as
  a net/http regression guard. The full-suite verdict comes from the CI run on the lead's
  `origin/develop` push after M4.
- AC-GTS2-007: regression guard. No committed tracked file outside `go.mod`, this SPEC
  directory, and `.moai/reports/t610/` differs from the merge-base with `develop`. The M4
  re-measure also admits the M5 sync paths.
- AC-GTS2-008 (sync): 0 `1.26.4` mentions and 5 target mentions across the 4 project
  documents.

AC-GTS2-001..008 are re-measured on the absorbed tree at M4. M4 also checks that the 5 sync
paths carry only the M5 sync commit's changes (acceptance.md § D.0b).

Design decisions D1–D4 (acquisition, target version, directive form, global landing) are
recorded with their rationale and alternatives in `plan.md` § D-DESIGN. The Implementation
Kickoff Approval gate can override any of them.

## E. Exclusions (What NOT to Build)

This section lists what is out of scope for this SPEC.

### Out of Scope — non-affecting findings

- The 3 vulnerabilities in "modules you require" that govulncheck reports the code does not
  call. They are not affecting, and no `require` line is touched. The "2 in packages you
  import" reported at go1.26.4 are not excluded work: they are standard-library findings the
  bump itself clears (0 at go1.26.6 and go1.26.8).

### Out of Scope — Go minor version bump

- go1.27.x, a minor bump with its own compatibility review. This SPEC stays on the 1.26
  patch line.

### Out of Scope — already-shipped binaries

- The vulnerability status of released binaries built with go1.26.4. Symbol reachability
  in govulncheck is not exploit reproduction, and re-assessing or re-releasing past
  artifacts is not part of this SPEC.

### Out of Scope — frozen capture goldens

- The three `go1.26.4` mentions in
  `internal/cli/testdata/tuxiu/postm4/update.{nocolor,notty,tty}.stdout.golden:2`. They are
  frozen captures compared only against other goldens, and the line they sit on is the
  "◆ MoAI-ADK" identity band, which `tuxIsPresentation` drops before comparison
  (`internal/cli/tuxiu_characterization_test.go:83-85`). The bump cannot break them, and
  rewriting them is not part of this SPEC.

### Out of Scope — other review findings and code changes

- Review finding SEC-02 (web Host-header check). It is a separate card.
- Any Go source change. If the bumped toolchain surfaces a compile or behavior break, the
  run phase stops and returns a blocker. A fix would be a separate SPEC.
- Any CI workflow edit, a recurring govulncheck CI gate, or Dependabot and dependency-policy
  configuration.
- Changing how environments acquire toolchains, for example making `GOTOOLCHAIN=local` or
  offline setups work. This is recorded as a residual risk in plan.md § D1.
