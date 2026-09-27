---
id: SPEC-AC-BASELINE-REFRESH-001
title: "Plan — durable AC-count corpus baseline refresh (regeneration mode + cascade procedure)"
version: "0.1.0"
created: 2026-09-22
author: lane agent-20 (t1068)
---

# Plan — SPEC-AC-BASELINE-REFRESH-001

Tier M · cycle_type: tdd · affects `internal/spec` (test package) + tracked snapshot + one tracked procedure doc.

## §1 Approach

Three milestones, strictly ordered: the generator exists before the catch-up cascade runs, and the cascade's green state is what the mutation verification then attacks. The judgment semantics of `TestACCounterFullCorpusMatchesBaseline` are PRESERVED throughout — every requirement in §B of spec.md is additive (a mode, a message suffix, a header, a document).

Development mode: TDD on the regeneration emission (M1) — the format round-trip test is written first and must fail before the emitter exists. M2 and M3 are measurement/verification milestones, not code milestones.

## §2 File inventory (run phase)

| File | Action | Purpose |
|---|---|---|
| `internal/spec/ac_count_clause_test.go` | modify | Add the env-gated regeneration mode (`TestACCounterBaselineRegenerate` + emitter factored to a writer-taking helper), the provenance header emission, and the remedy pointer on the three recorded-file failure sites (`:455` halting problem, `:471` count/state problem, `:479` vanish). Extend the file doc comment with the cascade rule pointer. |
| `.moai/reports/t338/ac-count-baseline.txt` | regenerate (data) | Catch-up cascade: +N new entries where N is measured at generation (84 at the 2026-09-22 post-authoring measurement — all COUNT; includes SPEC-AC-GUARD-001, the four MODEL-MATRIX successors, and this SPEC's own acceptance.md), −1 removal (SPEC-MODEL-PROFILE-MATRIX-002), new provenance header. Path unchanged. |
| `.moai/docs/ac-count-baseline-refresh.md` | create (tracked) | The cascade procedure: trigger events, same-commit rule, diff-review discipline, named-cause commit-message rule, the exact command. |
| `internal/spec/CLAUDE.md` | modify (1 line) | Cross-reference the procedure doc from the package conventions file. |

NOT touched: any `acceptance.md` body; `acComparison`'s contract; the corpus glob; the snapshot path; `internal/cli` (its `todo_triage_test.go` reference is fixture data keyed to the path, which does not move).

## §3 Milestones

### M1 — regeneration mode, provenance header, remedy messages (REQ-ABR-001/002/003)

RED first: a unit test asserting the emitter's contract — given fixture measurements, the emitted text carries the four header elements (frozen glob statement verbatim, tree SHA, date, command), one line per entry in the current `COUNT`/`HALT` format, sorted by path, and round-trips through `parseACBaseline` to equal entries. GREEN: factor the corpus walk into an emitter helper taking an `io.Writer`; `TestACCounterBaselineRegenerate` gates it on `MOAI_AC_BASELINE_REGENERATE=1` and passes the real snapshot path; without the variable it performs no write (unit-tested with a temp path). Add the remedy suffix (the one-line regeneration command) to the three recorded-file failure sites.

Verification:
```
go test ./internal/spec -run 'TestACCounterBaselineRegenerate|TestACBaselineEmitter' -count=1   # exit 0
MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1   # writes the file; exit 0
unset MOAI_AC_BASELINE_REGENERATE && go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1   # no write; exit 0 (scrub and command travel in ONE invocation)
```

### M2 — catch-up cascade (REQ-ABR-004/005 executed once for the accumulated debt)

Re-read HEAD immediately before generating (staleness rule). Run the regeneration mode, then READ THE DIFF before committing. Expected shape: **+N COUNT lines, where N is measured at generation time — including this SPEC's own `acceptance.md`, which entered the glob population during plan phase (84 at the 2026-09-22 post-authoring measurement, spec.md §A.3 second block; re-measure at generation, do not carry)** — plus −1 MATRIX-002 line and the header replacement. **The only stop rule is the AC-ABR-007 attribution predicate**: every added line must resolve to a SPEC directory absent from the old snapshot, the one removal to `20cdeb6bd`'s superseded-split, and the header to the provenance format. No expected total is pinned in advance — the population moves with every authored `acceptance.md`, and a magic number would false-trip on a line whose cause is perfectly nameable. Commit the snapshot ALONE with named causes per changed-line class.

Verification:
```
git diff --stat -- .moai/reports/t338/ac-count-baseline.txt   # N additions + 1 removal + header lines, one file
go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1   # exit 0, absent-report = 0 rows
go test ./internal/spec -count=1   # affected package green (full suite is CI's)
go test ./internal/cli -run 'TestTodoTriage' -count=1   # consumer green (fixture path unchanged)
```

### M3 — procedure document + mutation verification (REQ-ABR-004/007)

Write `.moai/docs/ac-count-baseline-refresh.md` (tracked — verified: `git ls-files .moai/docs/` carries 27 files, the directory is tracked): trigger events (superseded/split removal, `_archive/` move, count-affecting rewrite), the same-commit rule, the diff-review + named-cause discipline, the command, and the t573/20cdeb6bd incident record. Reference it from the snapshot header, the test doc comment, and `internal/spec/CLAUDE.md`. Then run the mutation verification (acceptance.md AC-ABR-005) and export evidence to `.moai/reports/t1068/`.

Mutation procedure (the card's [HARD] 5, made mechanical):
```
cp .moai/specs/SPEC-MODEL-MATRIX-CORE-001/acceptance.md /tmp/abr-mutation-backup.md
rm .moai/specs/SPEC-MODEL-MATRIX-CORE-001/acceptance.md
go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1   # MUST FAIL, naming SPEC-MODEL-MATRIX-CORE-001 at :479 AND the remedy command
cp /tmp/abr-mutation-backup.md .moai/specs/SPEC-MODEL-MATRIX-CORE-001/acceptance.md
go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1   # exit 0 (restored)
go test ./internal/spec -run TestACBaselineComparisonTransitions -count=1   # exit 0 — v0.5.0 absent rows still report-not-fail
```

## §4 Risks and mitigations

- **Regeneration drifts mid-run** (develop moves between generation and commit): re-read HEAD immediately before generating and before committing; the provenance header records the SHA actually measured, so a moved HEAD is visible in the artifact itself.
- **A real regression hides inside the catch-up bless** (e.g., one of the new files HALTs): at the 2026-09-22 post-authoring measurement (cd99336bf, spec.md §A.3 second block) the absent rows were 84/84 COUNT, 0 HALT — re-measure at generation; the emitter records whatever it measures, and a HALT row in the diff is reviewed like any other line and named in the commit message, never silently normalized.
- **Parser rejects the regenerated file**: prevented by the M1 round-trip test (emitter output must parse via `parseACBaseline` to equal entries) — the format cannot drift from its consumer.
- **The env var leaks into CI** (someone exports it): the mode writes only at explicit invocation with `-run TestACCounterBaselineRegenerate`; CI never sets the variable; the unit test asserts the no-variable path writes nothing.
- **Someone later "fixes" a red by weakening the :479 loop**: AC-ABR-006's diff audit pins the error conditions; the procedure doc records that the vanish guard is the gate's purpose, not its bug.

## §5 Verification scope

Scoped per §4/§6 doctrine: `go test ./internal/spec -count=1` (affected package) and `go test ./internal/cli -run TestTodoTriage -count=1` (consumer) locally; full suite and darwin/windows matrix are CI's verdict on the pushed head. `gofmt`/`go vet` on touched files. Template-first cycle does NOT apply (no `internal/template/templates/` files touched).

## §6 Open questions

None — `[NEEDS CLARIFICATION]` count: 0. The disposition fork was resolved by intent evidence (spec.md §A.6); the mechanism home, the procedure location (tracked, `git ls-files` verified), and the mutation procedure are all measured decisions. Remaining judgment calls inside run phase (exact message wording, helper factoring shape) are within REQ-ABR constraints and do not need user input.
