# t528 M4 — mutant results

Expectations were declared in `expectations.md` **before** any mutant ran; nothing
in that file was edited afterwards. Two runs are recorded, not one: the first run
found two real gaps and one invalid mutant, and deleting it would leave only the
tidy second run — which is exactly the boundary-into-claim demotion
`acceptance.md` §D.11 forbids.

Judging command (both runs):

```
go test ./internal/spec/ -run TestT528 -count=1 -timeout 600s
```

Driver: `run-mutants.py`. One mutant at a time; inject → verify the injection
landed → measure → restore → verify the restore. Base tree `13fda0f6e`;
`git diff --stat -- internal/spec/parser.go` empty after both runs.

## Run 1 — as first written

| # | mutant | pre-declared | run 1 | reading |
|---|---|---|---|---|
| 1 | drop numeric-tail requirement | DETECTED | DETECTED | as expected |
| 2 | drop em dash from separator set | DETECTED | **INJECTION-FAILED** | measurement defect, not a result |
| 3 | drop paren-qualifier allowance | DETECTED | **INJECTION-FAILED** | measurement defect, not a result |
| 4 | drop closing-bold allowance | DETECTED | DETECTED | as expected |
| 5 | drop sub-id suffix recognition | DETECTED | **INJECTION-FAILED** | measurement defect, not a result |
| 6 | `findACSectionStart` → 0 **and** `##` break removed | DETECTED | **INVALID** (build failure) | non-zero exit, but not a detection |
| 7 | remove ONLY the `##` break | DETECTED | **NOT DETECTED** | real gap in the card's own test |
| 8 | relax `AC-` prefix to any uppercase token | MAY NOT BE DETECTED | NOT DETECTED | pre-declaration was accurate |

Four things went wrong, and each is worth its own line.

**Mutants 2, 3, 5 never applied.** Their anchor strings occur twice in
`parser.go` — once in the pattern and once in the doc comment that explains it.
The driver requires each old-text to occur exactly once and reported
`INJECTION-FAILED` rather than running the tests. This is the distinction the
driver exists to make: an unapplied mutant produces a green run that is
indistinguishable from a guard that failed to catch it, and reporting it as NOT
DETECTED would have been a confident figure produced by a measurement that never
happened — the sixth instance of this card's recurring failure shape. Fixed by
anchoring each on enough surrounding regex to be unique to the pattern line.

**Mutant 6 was invalid.** Rewriting `return i + 1` to `return 0` leaves the loop
variable unused, so the package did not compile:

```
internal/spec/parser.go:65:6: declared and not used: i
FAIL	github.com/modu-ai/moai-adk/internal/spec [build failed]
```

The exit code was 1 and the driver called it DETECTED — but a build failure is
not evidence that any guard noticed anything. Every mutant is now checked for
`build failed`, and mutant 6 uses `return i * 0`, which returns 0 while keeping
the variable used.

**Mutant 7 found a real gap, which is the point of running it.** The scoping
fixture in `TestT528SectionScopingInvariant` placed its out-of-section decoy
only BEFORE the AC section. Such a decoy is unreachable for a second, unrelated
reason — `findACSectionStart` returns the index AFTER the heading, so earlier
lines are never scanned at all — and the fixture therefore passed even with the
section terminator deleted. It was testing `findACSectionStart` twice and
`extractACLines` not at all. Fixed by adding a decoy in a `## Notes` section
AFTER the AC section, which only the terminator keeps out.

**Mutant 8's pre-declaration held.** Nothing in the card asserted the `AC-`
prefix, so the anchor could have been widened to collect `REQ-` and `TEST-`
bullets with every other test green. Closed with `TestT528ACPrefixRequired`.

## Run 2 — after the fixes

| # | mutant | pre-declared | run 2 | failing test |
|---|---|---|---|---|
| 1 | drop numeric-tail requirement | DETECTED | DETECTED | `TestT528NumericTailRequired` |
| 2 | drop em dash from separator set | DETECTED | DETECTED | `TestT528SeparatorSet` (7 FAIL lines) |
| 3 | drop paren-qualifier allowance | DETECTED | DETECTED | `TestT528ParenQualifier` (10 FAIL lines) |
| 4 | drop closing-bold allowance | DETECTED | DETECTED | `TestT528BoldWrapper` (14 FAIL lines) |
| 5 | drop sub-id suffix recognition | DETECTED | DETECTED | `TestT528SubIDSuffixPreserved` |
| 6 | `findACSectionStart` → 0 **and** `##` break removed | DETECTED | DETECTED | `TestT528SectionScopingInvariant` |
| 7 | remove ONLY the `##` break | DETECTED | DETECTED | `TestT528SectionScopingInvariant` |
| 8 | relax `AC-` prefix | MAY NOT BE DETECTED | DETECTED (after the gap was closed) | `TestT528ACPrefixRequired` |

Build-failure check across all eight logs: `grep -c 'build failed'` returns **0**
for every one, so each detection is a test failing rather than the package
refusing to compile.

## Boundary the mutants draw

Mutant 8's detection is narrower than it looks. `TestT528ACPrefixRequired` has
three cases and only two failed under that mutant: the mutant rewrites the prefix
to `[A-Z]+-`, which admits `REQ-` and `TEST-` but not `M1-`, because `M1`
contains a digit. So the guard covers letter-only prefixes and says nothing about
alphanumeric ones. Recorded rather than tidied away — that is where the guard's
edge actually is.

Nothing in the mutant set draws the over-acceptance axis. Every mutant above
removes capability, and over-acceptance happens while the capability works
(`plan.md` §C.1). That axis is measured separately in M6.
