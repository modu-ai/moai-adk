# t528 M4 — mutant expectations, DECLARED BEFORE ANY MUTANT IS INJECTED

Recording the expectation first is the whole point: writing "this one was never
going to be caught" after seeing the result is rationalisation, not a verdict
(`plan.md` §H). Expectations below are copied from `plan.md` §C / `acceptance.md`
§D.11, which fixed them at plan time.

Base tree: `13fda0f6e`. Injection is one mutant at a time, each followed by
measurement and restore. The judging command is the card's own test set:

```
go test ./internal/spec/ -run TestT528 -count=1 -timeout 600s
```

| # | mutant | pre-declared expectation |
|---|---|---|
| 1 | drop the numeric-tail requirement from the id | DETECTED |
| 2 | drop `—` (U+2014) from the separator set | DETECTED |
| 3 | drop the parenthesised-qualifier allowance | DETECTED |
| 4 | drop the closing-bold allowance | DETECTED |
| 5 | drop the sub-id suffix (`.a` / `.a.i`) recognition | DETECTED |
| 6 | `findACSectionStart` returns 0 **and** the `##` break in `extractACLines` is removed | DETECTED — exercises BOTH clauses of AC-ACA-001-006 |
| 7 | remove ONLY the `##` break in `extractACLines` | DETECTED — represents "collects lines outside the section" on its own |
| 8 | relax the `AC-` prefix to an arbitrary uppercase token | **MAY NOT BE DETECTED** — recorded as a possible non-detection in advance |

Mutant 6 is split from 7 because `findACSectionStart` returning 0 alone
exercises only half of AC-ACA-001-006: `extractACLines` breaks at the first `##`
heading, so `return 0` ends collection in the document preamble rather than
making the whole document an AC section.

A mutant that is expected DETECTED and is not caught means the corresponding AC
is vacuous, and that AC is fixed before the card proceeds. A non-detection is
preserved here rather than deleted — it draws the guard's boundary, and deleting
it demotes that boundary from a measurement to a claim.
