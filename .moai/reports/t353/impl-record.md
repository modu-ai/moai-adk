# t353 — REQ-MRG-010 (R4-form lint exclusion) implementation record

Tree: worktree `.claude/worktrees/t353`, branch `WT-r4-imperative-exempt`, base HEAD `9b89b8c5b`.
Scope touched: `internal/spec/lint_movingref.go`, `internal/spec/lint_movingref_test.go`,
`internal/spec/testdata/movingref/{r4-form,r4-korean,r4-cm1-positional,r4-cm2-command-token}/`, this file.
Doctrine files untouched: `git diff --stat` over local + template `verification-claim-integrity.md` is empty (byte-identical).
SPEC artifacts untouched: no file under `.moai/specs/` modified (`git status` shows only the two Go files tracked-modified plus new fixtures).

## Claim

The `MovingRefUnpinned` rule now carries the R4-form exclusion (REQ-MRG-010, AC-MRG-013 as retained
at acceptance.md §"DEFERRED out of this SPEC at v0.7.0"). The keying predicate is **imperative
structure, never a command token**: a flagged-shape line is exempt iff it carries BOTH
(a) an imperative measuring directive directly introducing a backticked command —
`(?i)(?:\b(?:re-?measure|measure(?:\s+with)?|run)\s*[:：]?\s*` + backtick + `|(?:측정|재측정)\s*[:：]?\s*` + backtick`) —
AND (b) a demoted dated reference — at least one parenthetical containing BOTH a calendar date
(`\d{4}-\d{2}-\d{2}`) and a demotion label (`reference|reading|기준|참조|판독`). Either conjunct alone
is NOT exempt. R4 stays named in the finding message (unchanged `remedyClause()`).

## Evidence

### RED (exclusion absent — proves the fixtures are flaggable and the exclusion does not yet exist)

Command: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/spec/ -run 'MovingRef' -count=1 -v` (exit 1). Verbatim:

```
--- FAIL: TestMovingRef_R4FormNotFlagged (0.26s)
--- FAIL: TestMovingRef_R4KoreanFormNotFlagged (0.27s)
    lint_movingref_test.go:365: expected 0 findings on the Korean R4 form, got 1: [{testdata/movingref/r4-korean/spec.md 32 warning MovingRefUnpinned `origin/develop` decides an invariant claim on this line ... }]
--- PASS: TestMovingRef_R4PositionalMutantStillFlagged (0.27s)
--- PASS: TestMovingRef_R4CommandTokenMutantStillFlagged (0.30s)
```

(The full English-form failure row is in `/tmp/t353-red.txt`, this run.) The two CM tests passing
under RED is the non-vacuity half: both counter-mutation fixtures were measured flaggable before
the exclusion existed.

### GREEN (structural exclusion implemented)

Same command (exit 0). Verbatim:

```
--- PASS: TestMovingRef_FiresOnUnpinnedAnchor (0.60s)
--- PASS: TestMovingRef_PinnedClaimNotFlagged (0.32s)
--- PASS: TestMovingRef_MarkerSuppressesOnlyWithReason (0.91s)
--- PASS: TestMovingRef_ThreeDotNotExempt (0.31s)
--- PASS: TestMovingRef_AdvisoryAndStillReported (0.64s)
--- PASS: TestMovingRef_DivergenceFigureVariant (0.32s)
--- PASS: TestMovingRef_MessageNamesAllFourBranches (0.33s)
--- PASS: TestMovingRef_ReadsSiblingArtifacts (0.35s)
--- PASS: TestMovingRef_NegativeControlOnClaimConjunct (0.34s)
--- PASS: TestMovingRef_R4FormNotFlagged (0.29s)
--- PASS: TestMovingRef_R4KoreanFormNotFlagged (0.30s)
--- PASS: TestMovingRef_R4PositionalMutantStillFlagged (0.28s)
--- PASS: TestMovingRef_R4CommandTokenMutantStillFlagged (0.33s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	5.749s
```

Re-run after the mutant probes were reverted (exit 0): `ok github.com/modu-ai/moai-adk/internal/spec 5.197s`.

### Counter-mutation CM-1 (positional bypass) — planted, red, reverted

Transient mutant: R4 check replaced with `r4ParentheticalPattern.MatchString(line)` (a parenthetical
anywhere after the ref). Verbatim (exit 1):

```
--- FAIL: TestMovingRef_FiresOnUnpinnedAnchor (0.44s)          <- AC-MRG-001 red, as specified
--- FAIL: TestMovingRef_R4PositionalMutantStillFlagged (0.28s)
    lint_movingref_test.go:385: expected 1 finding on the CM-1 positional mutant, got 0: []
--- FAIL: TestMovingRef_R4CommandTokenMutantStillFlagged (0.30s)
```

### Counter-mutation CM-2 (command-token key) — planted, red, reverted

Transient mutant: R4 check replaced with `strings.Contains(line, "git fetch")`. Verbatim (exit 1):

```
--- PASS: TestMovingRef_FiresOnUnpinnedAnchor (0.44s)          <- AC-MRG-001 stays green
--- PASS: TestMovingRef_DivergenceFigureVariant (0.36s)        <- AC-MRG-006 stays green
--- PASS: TestMovingRef_R4PositionalMutantStillFlagged (0.30s) <- CM-1 stays green
--- FAIL: TestMovingRef_R4CommandTokenMutantStillFlagged (0.31s)
    lint_movingref_test.go:409: expected 1 finding on the CM-2 command-token mutant, got 0: []
```

Exactly the DoD pattern: CM-2 goes green-to-red under a command-token key while AC-MRG-001, -006
and CM-1 all stay passing.

### Corpus over-exemption check (whole `.moai/specs/**`, this worktree)

- BEFORE: `go build -o /tmp/moai-t353-base ./cmd/moai && /tmp/moai-t353-base spec lint --json` →
  `jq -r '[.[] | select(.code=="MovingRefUnpinned")] | length'` → **115**
- AFTER: `go build -o /tmp/moai-t353-after ./cmd/moai && /tmp/moai-t353-after spec lint --json` →
  same jq → **113**
- `diff` of the sorted `file:line` coordinate lists → exactly two lines removed, zero added:

```
< .../t353/.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:357
< .../t353/.moai/specs/SPEC-MOVING-REF-GUARD-001/progress.md:158
```

Both newly exempted lines were read back: each is verbatim AC-MRG-013's own retained fixture line
(`- verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty)` —
progress.md:158 is its quotation in the §E.2 record), i.e. genuinely-R4 lines and precisely the
line the deferred criterion was written around. No other corpus finding changed.

### Static checks

- `go vet ./internal/spec/` → exit 0.
- `go test ./internal/spec/ -count=1` (full package) → exit 1 with the single failure
  `TestCatalogHashParity` (CATALOG_HASH_DRIFT between `internal/template/templates/.claude/skills/*/SKILL.md`
  and `catalog.yaml`). Pre-existing at base HEAD: `git status internal/template/` is clean — both
  sides of that comparison are at their committed state in this worktree, untouched by this card.
  Regenerating catalog hashes is outside this card's declared scope. Every other test in the
  package passes.

## Baseline-attribution

All commands above ran in this run, in this worktree (`.claude/worktrees/t353`), against the
working tree as it stood at each phase: RED on base HEAD `9b89b8c5b` plus fixtures/tests only;
GREEN/mutants/corpus-after on the same tree plus the implementation in `internal/spec/lint_movingref.go`.
The corpus BEFORE binary was built from the pre-implementation tree; the AFTER binary from the
post-implementation tree; both lint runs read the identical, unmodified `.moai/specs/` corpus.

## Gaps

- The corpus measurement is local to this worktree's corpus snapshot; `origin/develop` CI is the
  full-suite verdict surface and has not run this change (lane-local verification only, per
  gitflow-lane-protocol §8).
- `TestCatalogHashParity` failure was NOT repaired (pre-existing, out of scope); it was verified
  pre-existing only by the committed-state argument plus its subject matter being untouched — the
  test was not separately re-run on a stashed tree.
- The "remove the exclusion" mutation is evidenced by the RED run itself (the pre-implementation
  tree IS that mutant); it was not re-planted after GREEN.
- R4 fixture properties (moving-ref token count, SHA absence, etc.) were inherited from the
  acceptance.md §AC-MRG-013 measured table rather than re-measured with fresh greps in this run,
  except where the RED run itself demonstrates flaggability (which is the load-bearing row).
- golangci-lint was not run (mission's verification list named go vet, which is clean).

## Residual-risk

- The imperative-directive alternation includes the common word "run" followed by a backtick; the
  demoted-dated-reference conjunct bounds the false-exemption risk (both conjuncts required), and
  the corpus sweep measured zero unintended exemptions on today's 115-line finding set — but a
  future line that happens to carry "run `<cmd>`" AND a parenthetical with a date plus one of the
  five demotion labels would be silently exempted. The corpus sweep is the only guard against that
  drift; it is not wired as a standing test.
- The demotion-label list (`reference|reading|기준|참조|판독`) is vocabulary, not pure structure; a
  corpus that demotes dated references with new labels (e.g. "as of", "당시") would fall outside
  the exclusion and be flagged despite being R4-form — an under-exemption, which the deferral's
  trade-off analysis treats as the safe direction.
- Detection limit L2 (shape, never subject) still holds: the exclusion reads line structure and
  can be satisfied by a forged line combining both conjuncts around a genuinely unpinned claim;
  the R3 marker with its mandatory reason remains the author-declared path and R4's message
  placement is unchanged.
