---
id: SPEC-CLI-COVER-SUPPRESS-001
title: "internal/cli coverage suppression claim — measured mechanism verdict, post-t477 coverage sweep, and CLAUDE.local.md §6 target judgment (report-only investigation)"
version: "0.1.0"
status: completed
created: 2026-09-04
updated: 2026-09-04
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "cli, coverage, go-test-cover, investigation, measurement, report-only"
related_specs: [SPEC-PRECOMMIT-VET-MONOREPO-001, SPEC-UPDATE-HOOK-DELIVERY-001]
---

# SPEC-CLI-COVER-SUPPRESS-001

## §A Problem / Motivation

During card t237's run phase (SPEC-PRECOMMIT-VET-MONOREPO-001), the main `internal/cli` package
coverage cell could not be recorded — the package was red on an inherited failure
(`TestBinaryLag_DoctorCheckNameSetIsUnchanged`), and the run-phase record explains the missing
number with a mechanism claim:

> "**Unmeasured** — the package FAILs on the inherited red (see AC-PVM-010), and `go test -cover`
> prints no coverage line for a failing package. Named gap, not a claim."
> — `.moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md` §E.2 coverage table (lane-10's
> authoring tree `32319621b` per the lead's dispatch; sentence read directly on this branch at
> HEAD `8b391bc8c`)

That sentence asserts a toolchain behavior no one had measured. If it is wrong, the "Unmeasured"
cell was not forced by the toolchain — it was lost on the aggregation side — and every future
card that hits a red package would inherit the same wrong expectation. This card (t481, Class B
investigation) exists to settle it with measurements, in three duties:

1. **Mechanism verdict** — verify or refute "a failing package prints no coverage line" with a
   minimal two-package experiment observing BOTH output shapes (plain `-cover` stdout and
   `-coverprofile` data) on a pinned go version. *(Measured at plan time — REFUTED; §E.)*
2. **Post-repair coverage** — after t477's allowlist repair landed on develop, measure whether
   the main `internal/cli` package coverage number actually appears in a full
   `./internal/cli/...` `-cover` sweep. *(PENDING — the run-phase subject; value never assumed.)*
3. **Target judgment** — judge each measured package number against the CLAUDE.local.md §6
   coverage targets (85% per-package minimum; 90%+ for the critical cli/template/hook set).
   *(Report only.)*

Inherited-red provenance chain: t466 (commit `2b4582a43`, SPEC-UPDATE-HOOK-DELIVERY-001)
registered the "Hook Delivery" doctor check without extending the `namesAddedAfterBaseline`
allowlist in `binary_lag_test.go`; that made `TestBinaryLag_DoctorCheckNameSetIsUnchanged` fail
on every tree descending from t466; t477 repaired it (commit `b2f98f6aa`, merged into develop as
`a825183dd` — both verified ancestors of this branch's HEAD `8b391bc8c` via
`git merge-base --is-ancestor`).

**Report-only constraint [HARD]**: no code, test, template, or configuration change is permitted
in this card. Its entire write surface is evidence under `.moai/reports/t481/` and this SPEC
directory. Coverage *improvement* work belongs to other cards.

## §B History

| Date | Version | Change |
|------|---------|--------|
| 2026-09-04 | 0.1.0 | Initial plan-phase artifacts (card t481, Tier S, report-only investigation). Mechanism experiment measured this session (§E — REFUTED); post-t477 sweep marked PENDING. |

## §C Requirements (GEARS)

### §C.1 Measurement requirements

**REQ-CSS-001** — **When** the suppression claim ("`go test -cover` prints no coverage line for a
failing package") is evaluated, the verdict shall rest on a minimal-module experiment containing
exactly one passing package and one deliberately failing package, measured on a pinned go
version, observing both the plain `-cover` stdout shape and the `-coverprofile` profile-data
shape, with the raw outputs persisted under `.moai/reports/t481/`.

**REQ-CSS-002** — **When** the t477 allowlist repair is present on the measurement tree, the run
phase shall observe the full sweep `go test -count=1 -cover -timeout 900s ./internal/cli/...`
(output at `.moai/reports/t481/cli-cover-post-t477-sweep.txt`) and shall record the main
`internal/cli` package coverage number — or shall record that number's absence as an observed
fact accompanied by the verbatim output lines around the package result. The sweep's value shall
not be assumed at plan time.

**REQ-CSS-003** — **When** a package coverage number is measured, the verdict table shall judge
it against the CLAUDE.local.md §6 target for its class — 85% minimum per package, 90%+ for the
critical set (cli, template, hook) — with each row naming the measured number, the applicable
target class, and the judgment.

**REQ-CSS-004** — **When** the measured mechanism contradicts the t237 record's claim, the
correction record shall carry the original sentence verbatim with its file and tree, placed
directly next to the measured counter-evidence and the pinned go version.

### §C.2 Evidence discipline

**REQ-CSS-005** — Every fact this SPEC reports shall be attributed to a command plus its verbatim
observed output plus an evidence path; a number carried over from another card's measurement is
a relayed value, not this card's measurement, and shall be labeled as such.

### §C.3 Scope fence

**REQ-CSS-006** — The card shall not modify any Go source, test, template, hook, or configuration
file; its only write surface is `.moai/reports/t481/**` and
`.moai/specs/SPEC-CLI-COVER-SUPPRESS-001/**`.

## §D Success Criteria

- All four acceptance criteria (AC-CSS-001..004) discharged with cited evidence.
- The suppression claim carries a measured verdict — stdout shape AND coverprofile shape, go
  version pinned — not a restated belief.
- The main `internal/cli` coverage number (or its observed absence) is recorded from the settled
  post-t477 sweep output, never assumed.
- The verdict table judges every measured package number against its §6 target class.
- Zero code/test/template/config diff attributable to this card (report-only holds).

## §E Measured Facts vs Unmeasured

Evidence grounding in the SPEC-CODEX-DUAL-AGENTS-001 §E shape. Files under
`.moai/reports/t481/` were read directly on this branch (HEAD `8b391bc8c`); tree pins noted per
row. Nothing in this table carries a value for the pending sweep.

| Fact | Status | Evidence |
|---|---|---|
| Toolchain: `go1.26.4 darwin/arm64` | **MEASURED** | this session, worktree t481 |
| Minimal module at `.moai/cache/t481-mech/` (package `ok` passes, package `bad` fails deliberately; module path `t481mech`) | **MEASURED** | fixture dir present; outputs below |
| Plain run `go -C .moai/cache/t481-mech test -count=1 -cover ./...` → exit 1; the FAILING package prints `coverage: 100.0% of statements` on its OWN line above its package line (`FAIL	t481mech/bad	0.414s`), while the passing package's number is glued to its `ok` line (`ok 	t481mech/ok	0.727s	coverage: 100.0% of statements`) | **MEASURED** | `.moai/reports/t481/mechanism-plain.txt` (8 lines, verbatim) |
| With `-coverprofile`: the failing package's profile lines ARE merged into the profile (`t481mech/bad/bad.go:4.24,6.2 1 1`) | **MEASURED** | `.moai/reports/t481/mechanism-coverprofile.txt` + `mechanism-coverprofile-data.txt` |
| Process note (honesty record): an initial profile grep used the wrong pattern `t481-mech` (hyphen) and read 0; corrected to `t481mech` | **MEASURED** | lead's dispatch + the data file |
| **CONCLUSION**: on go1.26.4, NEITHER the stdout percentage NOR the profile data is suppressed for a failing package — the t237 mechanism claim is **REFUTED** at the toolchain level | **MEASURED** (derived from the two rows above) | — |
| t237's subpackage cells quote exactly the glued `ok <pkg> <time> coverage: X%` shape (e.g. `ok .../internal/cli/worktree 10.857s coverage: 87.1%`); the failing package's number sits on a separate line of a DIFFERENT shape. Hypothesis (labeled, not proven): an aggregation reading only the glued `ok`-line shape loses the failing package's number — the aggregation, not the toolchain, produced the "Unmeasured" cell | **MEASURED observation + labeled hypothesis** | t237 `progress.md` AC-PVM-010 row / §E.2 (this branch HEAD `8b391bc8c`; lane-10 authoring tree `32319621b`) |
| Inherited-red chain: `2b4582a43` (t466) registered "Hook Delivery" without the allowlist extension; t477 fix `b2f98f6aa` merged as `a825183dd`; both ancestors of this HEAD | **MEASURED** | `git merge-base --is-ancestor b2f98f6aa HEAD` → exit 0; `a825183dd` in `git log`; t477 verdict `.moai/reports/t477/verdict.md` (commit `82222b332`) |
| At card base `25a3212a9` (pre-t466): `go test -count=1 -cover -run TestBinaryLag -timeout 480s ./internal/cli/` → ok, `coverage: 5.9% of statements` — a single `-run`-filter number, NOT the package number | **MEASURED** | `.moai/reports/t481/cli-binarylag-at-base.txt`; base verified ancestor of HEAD |
| Full sweep `go test -count=1 -cover -timeout 900s ./internal/cli/...` on the absorbed develop tree; main `internal/cli` package coverage number | **PENDING** | output target `.moai/reports/t481/cli-cover-post-t477-sweep.txt` — file present at **0 bytes** at plan authoring (sweep in flight, main-session-owned); value NEVER assumed; the AC-CSS-002 subject |

## §F Traceability Matrix

| Requirement | Acceptance criteria |
|---|---|
| REQ-CSS-001 | AC-CSS-001 |
| REQ-CSS-002 | AC-CSS-002 |
| REQ-CSS-003 | AC-CSS-003 |
| REQ-CSS-004 | AC-CSS-004 |
| REQ-CSS-005 | AC-CSS-001..004 (attribution discipline across all rows) |
| REQ-CSS-006 | AC-CSS-003 |

## Out of Scope

The following are out of scope for this SPEC:

### Out of Scope — any code, test, or coverage work

- No coverage *improvement*: writing tests to raise a judged-below-target number belongs to a
  separate card; this card measures and judges only.
- No repair work: the BinaryLag allowlist repair already landed (t477 `b2f98f6aa`); this card
  does not touch `binary_lag_test.go`, `doctor.go`, or any other source file.
- No test authoring, deletion, or modification of any kind.

### Out of Scope — coverage-gate and aggregation machinery

- Adding, changing, or enforcing any CI coverage gate or threshold automation.
- Fixing the hypothesized glued-line-only aggregation: if the run-phase report confirms that
  hypothesis, the fix is a follow-up card's subject — this card records the finding only.

### Out of Scope — re-judging the t237 SPEC's verdict

- t237's AC-PVM-010 FAIL attribution (inherited red authored by `2b4582a43`, repair owned by the
  t466/BinaryLag axis) stands unchanged; this card corrects only the mechanism *sentence* in the
  §E.2 coverage row (AC-CSS-004) — it does not reopen t237's AC matrix.

### Out of Scope — packages outside the sweep

- Coverage judgment for packages other than those the `./internal/cli/...` sweep observes is not
  attempted; no `go test ./...` full-suite run happens in this card (load discipline,
  CLAUDE.local.md §4/§6).

## §G Cross-References

- SPEC-PRECOMMIT-VET-MONOREPO-001 — the t237 SPEC whose progress.md carries the claim under
  test (§E.2 coverage row, AC-PVM-010 row).
- SPEC-UPDATE-HOOK-DELIVERY-001 — t466, origin of the inherited red (commit `2b4582a43`).
- `.moai/reports/t477/verdict.md` — t477 repair verdict (RED/GREEN cells + two mutants proving
  the backtick allowlist key load-bearing), commit `82222b332`.
- CLAUDE.local.md §6 — the coverage targets this card judges against.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1/§2 — the defect class this card
  polices (unmeasured mechanism claims) and the attribution discipline its rows follow.
- plan.md — run-phase milestones M1-M3; acceptance.md — AC-CSS-001..004.
