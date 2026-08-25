# PR #1611 — override request (prepared, pending review-quota recovery)

Head at time of writing: `dd1445ab8`. Prepared during a CodeRabbit rate-limit window so that, if
the quota does not recover, the operator has the evidence rather than a summary.

## What is being asked, and what is not

The request is to merge past the auto-merge risk cap, **not** to skip review. Review ran and its
findings were acted on; what is unavailable is a *re-measurement* of the risk grade after the last
fix.

The distinction is the whole basis of this request:

> The grade did not fail to come down. **It cannot currently be re-measured.**

Those two states look identical in the checks list — both leave `Merge Risk: 🟡 Moderate` on the
PR — and they justify opposite decisions. The standing grade was computed at `0ebc24612`, before
the finding that grade most plausibly rested on was fixed. Reading it as the current grade would
be quoting a stale measurement as a live one.

If the quota recovers before a decision is made, **this request is withdrawn** and the review is
re-run instead. Re-review is the better evidence and stays the first choice.

## Findings ledger — 4 raised, 4 dispositioned

### F1 — `internal/constitution/retirement.go`, marker boundary (Major) — ACCEPTED, fixed

A bare prefix match classified `[SUPERSEDEDLY] text` and an unterminated `[SUPERSEDED …` as
retired. A retired entry skips drift, canary-gate, and source-file validation, so a live or
malformed clause could switch its own validation off.

Observed failing before the fix:

```
--- FAIL: TestIsRetiredClause/unterminated_marker
    IsRetiredClause("[SUPERSEDED text") = true, want false
--- FAIL: TestIsRetiredClause/longer_word_starting_with_the_marker_is_not_a_marker
    IsRetiredClause("[SUPERSEDEDLY] text") = true, want false
```

Non-blocking because the fix strictly narrows false positives: measured against the shipped
registry, classification changed for **0 of 100** clauses, and all four genuine `[SUPERSEDED by …]`
markers still classify as retired.

### F2 — `internal/constitution/shipped_registry_test.go`, `t.Skipf` → `t.Fatalf` (Major) — ACCEPTED, fixed

`os.Stat` failure caused a skip, so a deleted or renamed registry — precisely the failure the test
claims to guard — reported green.

Non-blocking because the change can only add failure signal to a test that previously had none for
that case. It cannot mask a defect; it can only surface one.

### F3 — `internal/constitution/retirement.go`, later bracket does not close the marker (Major) — ACCEPTED, fixed

Raised by the second review, against F1's own fix. `strings.Contains(rest, "]")` accepted
`"[SUPERSEDED live [HARD]"` — that bracket closes `[HARD`, leaving the marker itself unclosed, and
the clause then skipped validation. The finding was correct: F1 replaced an instance-shaped check
with another loose one.

Observed failing before the fix:

```
--- FAIL: TestIsRetiredClause/a_later_bracket_does_not_close_the_marker
    IsRetiredClause("[SUPERSEDED live [HARD]") = true, want false
```

Non-blocking, and the reason is measured rather than argued. The rule has now been narrowed twice,
and narrowing a classifier risks dropping a genuine retirement — which would silently re-enable
checks that were deliberately switched off, a failure with no signal of its own. So a regression
guard was landed with the fix: it asserts that every marker-carrying clause in the shipped registry
still classifies as retired.

That guard was itself observed failing, by temporarily breaking the classifier — it named all four
affected entries rather than reporting a bare count, and it is fatal when zero marked clauses are
found, so it cannot pass vacuously.

### F4 — `.moai/reports/t201/verdict.md`, MD040 fence language (Minor) — DECLINED

Measured basis for the decline, not a judgment call:

```
$ ls .markdownlint*                                          → no matches
$ grep -rn markdownlint .github/workflows/ Makefile .pre-commit-config.yaml   → 0 hits
```

The repository runs no markdownlint, so the finding cites a gate this project does not have. The
file is additionally a landing-time audit record, edited for real defects rather than for parity
with a linter that never runs here. Rationale was posted to the PR rather than kept local.

## Verification state

| Check | Result |
|---|---|
| `go test -count=1 ./internal/constitution/...` | ok |
| `go vet ./internal/constitution/...` | rc=0 |
| Shipped-registry classification regression | 0 of 100 clauses changed; 4 retirements preserved |
| CI at `dd1445ab8` | **complete: 0 in flight, 0 failures** — 21 success, 6 skipped |
| CodeRabbit at `dd1445ab8` | `state=success`, `desc=Review rate limited` — **review did not run** |

The CI row was re-read from the check-runs endpoint after the run finished, not carried over from
the in-flight snapshot this file was drafted against. Read with `status=="completed"` filtered
first and `conclusion` inspected inside it: an in-flight run carries an empty-string conclusion,
not null, so filtering on `conclusion != null` counts running jobs as failures.

If more time passes before this is submitted, re-read it again. A snapshot quoted later as a live
state is the exact error this request exists to avoid.

## Scope

Three files, all under `internal/constitution/`, plus one added test file. No template surface, no
public API, no behavior outside the retirement classifier and one test's failure mode.
