# SPEC-CLI-COVER-SUPPRESS-001 — Implementation Plan

> Tier S investigation card (report-only). The run phase measures and reports; it changes no
> code, test, template, or configuration file. The mechanism verdict is already measured at plan
> time (spec.md §E — REFUTED); the run phase's new work is consolidating that record (M1),
> consuming the settled post-t477 sweep (M2), and composing the target-judgment verdict table
> (M3). Artifact note: this card carries spec.md + plan.md + acceptance.md per the card
> dispatch; progress.md is created by the run phase, which owns the §E.1-§E.4 skeleton and body.

## §A Context

- **Tree**: worktree `.claude/worktrees/t481`, branch `WT-cli-cover-suppression`, plan-phase
  baseline `8b391bc8c` (local develop absorbed via fast-forward; contains t477's merge
  `a825183dd` and fix `b2f98f6aa` — ancestry verified with `git merge-base --is-ancestor`).
- **Card**: t481, Class B investigation/measurement. NO code or test changes; report only.
- **SPEC artifacts**: `.moai/specs/SPEC-CLI-COVER-SUPPRESS-001/{spec,plan,acceptance}.md`
  (+ progress.md, run-phase-owned).
- **Evidence directory**: `.moai/reports/t481/` — mechanism files already present and read
  (quoted verbatim in spec.md §E); sweep output target `cli-cover-post-t477-sweep.txt` present
  at 0 bytes at plan authoring — the MAIN SESSION owns the sweep execution; the run phase
  consumes the settled file and must not assume its value.
- **The three duties**: mechanism verdict (measured at plan time, consolidated in M1),
  post-t477 sweep recording (M2), §6 target judgment table (M3).

## §B Known Issues

- **B-load (verification load discipline)**: the card's only test execution is the
  `./internal/cli/...` sweep owned by the main session. The run phase does NOT re-run it and
  NEVER runs `go test ./...` locally (CLAUDE.local.md §4/§6 — the 2026-08-15 load-413
  incident).
- **B-sweep-race**: the sweep output file may still be growing when the run phase starts —
  consume only a settled file (the main session confirms completion); an empty or partial file
  at run time is a blocker report to the lead, not an invitation to assume.
- **B-inherited-reds**: the sweep may surface OTHER reds inherited from develop (not the t477
  one). Per the refuted mechanism, a red package on go1.26.4 still prints its coverage line —
  so a missing number for a red package needs its own verbatim recording, not a restated
  suppression claim.
- **B-relay**: t237's subpackage coverage numbers (80.9%–100.0% at `bfdf833e8`) are lane-10's
  measurements on a different tree — citable as their record, never as this card's.
- **B-filter**: the 5.9% number at card base is a `-run TestBinaryLag` single-filter result —
  explicitly not a package coverage number.
- **B-artifact-statelessness**: plan.md / acceptance.md carry no `status:` field; the lifecycle
  state lives in spec.md frontmatter only.

## §C Pre-flight (run phase, before M1; from the worktree root)

```bash
git rev-parse --show-toplevel          # .claude/worktrees/t481
git rev-parse --abbrev-ref HEAD        # WT-cli-cover-suppression
git merge-base --is-ancestor b2f98f6aa HEAD && echo T477-PRESENT
go version                             # pin the toolchain row (go1.26.4 darwin/arm64 at plan time)
ls -la .moai/reports/t481/             # mechanism files + sweep file state
```

Record the outputs in progress.md §E.2 as the run-phase baseline.

## §D Constraints

- **REPORT-ONLY [HARD]**: zero code/test/template/hook/config modifications. The card's entire
  diff is `.moai/reports/t481/**` + `.moai/specs/SPEC-CLI-COVER-SUPPRESS-001/**`. A diff
  anywhere else is a scope breach (REQ-CSS-006).
- **NO ASSUMED VALUES [HARD]**: the sweep's result is read from the settled output file; its
  value is never predicted, expected, or backfilled from t237's numbers (REQ-CSS-002).
- **FIXED COMMAND FORM**: the sweep under judgment is exactly
  `go test -count=1 -cover -timeout 900s ./internal/cli/...`, exit code observed unpiped.
- **NO LOCAL FULL SUITE [HARD]**: no `go test ./...` — lane-local scope only; CI owns the full
  suite (CLAUDE.local.md §4).
- **NO SCOPE CREEP INTO REPAIR**: no allowlist edits, no aggregation-tooling fixes, no coverage
  work — findings are recorded, not acted on (spec.md Out of Scope).

## §E Self-Verification

E1 AC matrix (4 rows, acceptance.md) · E2 n/a — no build obligation (zero source changes; the
sweep compiles the package as its own side effect) · E3 IS the deliverable — the coverage
judgment table vs CLAUDE.local.md §6 (85% package / 90% critical cli) · E4 n/a · E5 n/a (no
lint surface changed) · E6 artifact + commit list (commits owned by the run/lead per the card
dispatch) · E7 blockers (sweep not settled → blocker report, not an assumption) · E8 n/a (no
TDD surface). Every E-row names command + verbatim output + evidence path (VCI §2).

## §F Milestones

### M1 — Mechanism-verdict consolidation (Priority High)

Consolidate the plan-phase mechanism evidence into the run-phase record: quote the failing
package's stdout shape (`coverage: 100.0% of statements` on its own line above
`FAIL	t481mech/bad`) and the profile line (`t481mech/bad/bad.go:4.24,6.2 1 1`), pin
`go1.26.4 darwin/arm64`, and record the verbatim t237 sentence with its file and tree directly
next to the counter-evidence (AC-CSS-001, AC-CSS-004). The cells already exist under
`.moai/reports/t481/`; M1 formalizes them into progress.md §E.2 with per-row attribution.

### M2 — Post-t477 sweep consumption (Priority High)

Read the SETTLED sweep output (`cli-cover-post-t477-sweep.txt`, main-session-owned): record the
main `internal/cli` package coverage number — or, where no number appears, record the absence as
an observed fact with the verbatim surrounding lines (REQ-CSS-002 / AC-CSS-002). Attribute any
red in the sweep to its commit; a red is NOT a reason to skip recording — the mechanism verdict
says red packages still print coverage lines on this toolchain, so a missing number for a red
package is a counter-observation to record verbatim, not to explain away. If the file is
unsettled at run-phase time, return a blocker report to the lead.

### M3 — Target-judgment verdict table + card close (Priority Medium)

Compose the verdict table: every measured package number judged against its §6 target class
(`internal/cli` = critical, 90%+; other packages 85%+), each row naming number / target class /
judgment (REQ-CSS-003 / AC-CSS-003). Zero-diff check (report-only held). Findings for follow-up
cards (the aggregation-shape hypothesis if confirmed; any below-target package) are recorded in
the report — not acted on. The closing report to the lead names: the mechanism verdict, the
main-package number (or its observed absence), the verdict table, and the zero-diff proof.

## §G Anti-Patterns (this SPEC specifically)

- Do NOT assume the sweep's value because t477 landed — an expected value written before the
  output settles is exactly the unmeasured-claim defect this card exists to police.
- Do NOT answer a below-target judgment by writing tests inside this card.
- Do NOT re-run the sweep or any full suite locally "to be sure" — the main session owns the
  single sweep; CI owns the full suite.
- Do NOT cite t237's subpackage numbers as this card's measurements (carry-over ≠ baseline).
- Do NOT present the 5.9% `-run`-filtered number as the package coverage number.
- Do NOT reopen t237's AC-PVM-010 FAIL attribution — only the mechanism sentence is corrected.

## §H Cross-References

- spec.md §E — the measured-facts table M1 consolidates and M2 extends.
- acceptance.md — AC-CSS-001..004.
- `.moai/reports/t477/verdict.md` — the repair whose landing makes M2 meaningful.
- CLAUDE.local.md §6 (coverage targets) / §4 (integration chain; no local full suite).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — attribution for every row.
