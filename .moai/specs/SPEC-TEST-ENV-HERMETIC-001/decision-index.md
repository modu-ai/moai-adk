# Decision Index — SPEC-TEST-ENV-HERMETIC-001

Decisions surfaced during plan assembly that the operator has not settled. Each row states what is
unresolved and why; none carries a recommendation. Labels: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED,
FOUNDER. No row is routed DECIDED or POLICY-COVERED: no committed artifact in the authority register
decides an identical question under identical conditions (the nearest, SPEC-CLI-TEST-CWD-ISOLATION-001
plan §E D3, decided guard placement for a cwd-residue guard, a different question).

### Q1: Which mechanism makes the flip-prone tests hermetic — a package-level scrub, per-test `t.Setenv` pins, or a shared seeding helper?

Label: EVIDENCE-NEEDED
Authority anchor: n/a (no anchor for this label)
Why unresolved: the choice turns on how many tests flip under the lane env and whether any test relies on an ambient value; only five tests are measured, and the whole-package pairs (plan.md M1 c1, M4) that count the rest do not exist yet. The options and their measured blast radius are in spec.md §D.
Operator verdict:

### Q2: Is the recurrence guard an ordinary test function or part of `TestMain`?

Label: FOUNDER
Authority anchor: n/a (no anchor for this label)
Why unresolved: a test function is filtered out by a narrow `-run` selector and relies on the CI full-package run; a `TestMain` check rides every selector but fails every narrow run of the package. The tradeoff is a product-posture preference, not a measurable fact (spec.md §E, §H O4).
Operator verdict:

### Q3: How is the guard's axis family defined — by constant value prefix read from `internal/config/envkeys.go`, or by an explicit list?

Label: FOUNDER
Authority anchor: n/a (no anchor for this label)
Why unresolved: a value-prefix rule picks up new axes automatically but misses a future axis with a different prefix; an explicit list never misses a listed axis but is blind to an unlisted one. Neither is settled by a committed rule (spec.md §H O1).
Operator verdict:

### Q4: Which of `MOAI_AUTONOMY_TIER`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_FACTORY_MANAGED` enter the cli scrub set and which take a reasoned exemption?

Label: EVIDENCE-NEEDED
Authority anchor: n/a (no anchor for this label)
Why unresolved: production code in `internal/cli` references each, but whether any test depends on the ambient value is unmeasured; the M1 and M4 whole-package pairs decide it (spec.md §H O2).
Operator verdict:

### Q5: Does the sweep extend to `internal/discovery` and to a cross-package registry check?

Label: EVIDENCE-NEEDED
Authority anchor: n/a (no anchor for this label)
Why unresolved: three production files in `internal/discovery` reference `MOAI_KANBAN_ID`; no test there has been measured under the lane env, so whether a flip exists is unknown (spec.md §H O3, plan.md M4).
Operator verdict:

### Q6: Is a shared non-test seeding helper package wanted enough to lift the test-files-only constraint (REQ-THE-008)?

Label: FOUNDER
Authority anchor: n/a (no anchor for this label)
Why unresolved: the card offers "standardize a gate-seeding helper" as an alternative, while the card's non-goals and this SPEC's scope keep the change to test files; whether to add a non-`_test.go` package is a scope decision (spec.md §D Option C, §H O6).
Operator verdict:

### Q7: Is Tier M the right complexity tier?

Label: FOUNDER
Authority anchor: n/a (no anchor for this label)
Why unresolved: the implementation is test-only and a handful of files, which reads as Tier S; Tier M was chosen because acceptance.md carries the load-bearing RED-now ledger and the two-cell adoption discipline (the same reasoning as SPEC-CLI-TEST-CWD-ISOLATION-001 §C). The tier judgment belongs to the operator at assembly.
Operator verdict:
