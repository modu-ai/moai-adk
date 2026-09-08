# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — Implementation Plan

Card t570 · Tier S · one milestone · `cycle_type tdd`. Base `a4855f0b2` (= `origin/develop` at
entry; dated anchor, not an AC range edge). Audit estimate: ~20 lines of test.

## §A. Context

The plan implements the guard described in `spec.md` §B. Nothing here restates the defect; §B.1's
three measurements (B-1/B-2/B-3) are the plan's precondition and are re-read, not remembered, before
M1 starts.

## §B. Decisions, most-reversible-last

The two genuinely open decisions are stated first, because both are cheap to change now and
expensive to change after the tests land.

### B-1 — The ordering guard reads the finding Detail string, not the stat count

**Decision.** REQ-CDPG-003's assertion is on the rendered Detail substring
`relative entr` (present) / `oddly-formed` (absent), with the zero-stat-count kept as a supporting,
explicitly-labelled non-discriminating check.

**Why the alternative was rejected.** The dispatch proposed zero stat calls as the discriminator.
Measured against `codexStaleSkillFinding`, both the correct and the forbidden order `continue`
before reaching `osStatFn` (`spec.md` §B.3), so the count is 0 in both and asserts nothing. Adopting
it would have produced a guard that passes under the exact defect it names — the vacuous-green shape
this project treats as a defect in its own right.

**Reversibility.** Low cost now; after the test lands, a change here rewrites the AC and its
evidence record.

### B-2 — The extended-length case is a distinct test, not a table row

**Decision.** REQ-CDPG-002 gets its own named test function, because it carries a second assertion
(the classification premise) that the ordinary-absolute case does not.

**Why.** Folding it into a table would either drop the classification assertion or force an
optional-field table that reads worse than two functions. The card is ~20 lines; two small functions
are cheaper than one parameterised one.

### B-3 — Helper naming, decided before the first line is written

`statRecorder` is **already declared** in this package at
`internal/cli/doctor_codex_stale_skill_test.go:331`. A second declaration of that name broke the
package test binary during a merge — the incident is recorded verbatim in the comment above
`pruneReadbackStatRecorder` at `internal/cli/codex_skills_prune_readback_test.go:37-41`. The plan
therefore declares **no new recorder type**: it reuses `stubStatRecording`. Should a new identifier
prove unavoidable, it carries a card-unique prefix (`doctorGuardT570…`).

## §C. Reusable facilities already in the tree

Each was located this run; line numbers are hints, symbols are the address.

| Facility | Location | What it gives M1 |
|---|---|---|
| `stubStatRecording(t) *statRecorder` | `internal/cli/doctor_codex_stale_skill_test.go:337` | records every path handed to `osStatFn`, forwards to real `os.Stat`, restores via `t.Cleanup` |
| `stubStatInject(t, map[string]error)` | `internal/cli/doctor_codex_stale_skill_test.go:354` | per-path injected stat results, if a resolving/missing verdict is needed |
| `stubCodexHome(t, home)` | `internal/cli/doctor_codex_test.go:103` | points `CODEX_HOME` at the fixture tree |
| `writeCodexHomeConfig(t, []codexSkillEntrySpec{…})` | `internal/cli/doctor_codex_test.go:122` | builds a `config.toml` the parser fully recognises |
| `overrideSeparator(t)` | `internal/cli/codex_skills_prune_readback_test.go:30` | pins `configPathSeparator` to `'\\'`, restores via `t.Cleanup` |
| `TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths` | `internal/cli/doctor_codex_stale_skill_test.go` | the nearest existing doctor-side test — the new tests sit beside it, in the same file |

The six facilities live in **four different files** (`doctor_codex_stale_skill_test.go`,
`doctor_codex_test.go`, `codex_skills_prune_readback_test.go`, and the new tests' own home). They
are all in package `cli`, so every one is directly callable from the new tests and none is
re-declared. Measured this run:
`/usr/bin/grep -rnE 'func writeCodexHomeConfig|func stubCodexHome|func stubStatRecording|type statRecorder' internal/cli/`

## §D. Constraints

- **D-1 [HARD]** Test-only. No production edit (REQ-CDPG-007). A production change discovered to be
  necessary is a blocker report to the lead, not a silent widening.
- **D-2 [HARD]** `configPathSeparator`, `osStatFn`, and `codexUserHomeDir` are package-level vars.
  Overriding tests are non-parallel and restore through `t.Cleanup` (REQ-CDPG-006).
- **D-3 [HARD]** Verification command, run **sequentially**:
  `go test ./internal/cli/... -timeout 1200s`. The package has a measured ~615 s runtime, so the
  default 600 s timeout fails it spuriously. `go test ./...` is not run locally.
- **D-4** Fixtures under `t.TempDir()`; the real `~/.codex/config.toml` is never touched.
- **D-5** Go applies only the LAST `-run` flag. A single selector per invocation.

## §E. Self-verification

Before reporting M1 complete, the implementer confirms each of the following by re-reading its
evidence, not by recall:

1. Both new tests exist, are non-parallel, and every override restores via `t.Cleanup`.
2. `go test ./internal/cli/... -timeout 1200s` passes; the rc and a bounded tail are recorded under
   `.moai/reports/t570/`.
3. The AC-CDPG-004 mutant was actually executed: the revert applied, the guard observed FAILING with
   verbatim output, the revert undone, the guard observed passing. Both outputs recorded.
4. Any mutant the guard did NOT catch is recorded — it draws the guard's boundary and is evidence.
5. `git diff --stat` against `CARD_BASE` shows changes confined to `internal/cli/*_test.go` and
   `.moai/`.
6. No new identifier collides with `statRecorder` or `pruneReadbackStatRecorder`
   (`/usr/bin/grep -rn 'statRecorder' internal/cli/` re-read after the edit).

## §F. Milestones

### M1 — The doctor-side guard (the only milestone)

One milestone, because the card is one guard on one line. Splitting it would produce milestones
whose boundaries are administrative rather than decision-bearing.

Ordered steps:

1. **RED first.** Author the REQ-CDPG-001 test beside
   `TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths`. Before the guard can be trusted, the
   production line is temporarily reverted to `statPath = e.Path` and the test observed FAILING —
   this is the AC-CDPG-004 mutant, executed at authoring time rather than bolted on afterwards.
   Restore the line; observe GREEN.
2. **REQ-CDPG-002.** Add the extended-length test, including the
   `classifyCodexSkillPath(declared) == codexPathAbsolute` assertion that verifies the reachability
   premise in the test rather than trusting `spec.md` §B.2.
3. **REQ-CDPG-003.** Add the ordering guard against the Detail string per decision B-1.
4. **Verify.** Run D-3's command; record rc and bounded tail under `.moai/reports/t570/`.
5. **Record.** Write the AC evidence rows, including any mutant the guard missed.

## §G. Anti-patterns for this card

- **Asserting zero stat calls as the ordering discriminator.** Vacuously true in both orders
  (`spec.md` §B.3).
- **Declaring a second `statRecorder`.** Breaks the package test binary; the incident is already
  recorded in-tree.
- **Running `go test ./internal/cli/...` without `-timeout 1200s`.** The default 600 s fails on a
  measured ~615 s package and reads as a regression.
- **Claiming the guard works without executing the mutant.** An unobserved RED is an unobserved
  verification claim.
- **Quietly fixing production while "adding a test".** REQ-CDPG-007 makes that a blocker report.

## §H. Cross-references

- `spec.md` §B — the three measurements establishing the gap.
- `acceptance.md` — the four ACs and their two-cell records.
- `SPEC-CODEX-SKILL-PATH-READBACK-001` AC-CSRB-006 — the darwin-identity ceiling this card lifts.
