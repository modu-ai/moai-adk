# SPEC-ASTGREP-LANG16-001 — Acceptance Criteria

> Verification layer. The GEARS requirements live in `spec.md` §C and are not restated here.
>
> Tier L budget: 25 maximum. This document carries **19**, leaving headroom of 6.

## §0. The two-column rule this document is built on

[HARD] Every criterion below has been **executed** against the pre-implementation tree
(worktree `.claude/worktrees/t228`, HEAD `294b4b6ab`) and carries two observations, both recorded
in §H:

1. **RED NOW** — the command, its verbatim output today, and why that output is a failure.
2. **GREEN PATH** — the named milestone in *this* SPEC whose work flips that same command to
   passing, and what the output becomes.

A criterion needs **both**. One column alone admits one of two mutants:

| Missing column | Mutant admitted | Instance found in this SPEC |
|---|---|---|
| No RED NOW | **Vacuous** criterion — passes on anything, so it observes nothing. | `sg test` reported `ok` with zero tests in existence (iter-2 E1). |
| No GREEN PATH | **Impossible** criterion — red today and red forever, so the work cannot satisfy it. | The `catalog.yaml` clause: no ruleset edit produces a catalog diff (iter-2 E2). |

This is the same pairing the SPEC already requires elsewhere, applied to the criteria themselves:
an `invalid` case alone admits a rule matching everything, a `valid` case alone admits a rule
matching nothing (REQ-A16-008); an absence claim owes a citation or probe (REQ-A16-005). The
criteria were the one surface where the standard had never been applied.

[HARD] **The discipline extends to the requirement layer.** A requirement that no work in this
SPEC can make true is the same **impossible** mutant, and moving a clause up a layer must not be a
way to escape the check — REQ-A16-021's scope did exactly that for one round, because §0 as first
written bound only the criteria and a requirement had no equivalent gate. Every requirement in
`spec.md` §C is therefore scoped to artifacts this SPEC authors or modifies, and a requirement
stated over content nobody here can change is a defect to re-scope, not strictness to admire.

Path shorthand:

```
$T = internal/template/templates/.moai/config/astgrep-rules
$M = .moai/specs/SPEC-ASTGREP-LANG16-001/coverage-matrix.md
$R = the repo-side rule-tests root (outside $T, per REQ-A16-019)
```

---

## §A. Milestone 1 — harness, keying, placement, and the corpus pins

**AC-A16-001** — `sg test` runs a case per shipped rule, and all of them pass
**Given** the shipped ruleset and its rule-test root,
**When** `sg test --config $T/sgconfig.yml` is run,
**Then** its reported test count **equals the number of rules in the shipped ruleset** — 26 at M1;
or, if M1 settles keying by id alone, the number of **distinct rule ids** in the shipped ruleset
after the rename, as recorded in `design.md` §3.3 — and it reports zero failures and zero missing
snapshots.

> Both targets are derivable from the tree, neither is declarable. The earlier wording — "the
> count the M1 id-keying decision yields" — left the number open, so an implementer could name any
> observed figure as the one the decision produced. Naming distinct-rule-ids ties the alternative
> to something countable and to the record AC-A16-004 already forces M1 to write.

> The count comparison is the assertion, not the exit code. `sg test` reports a zero-test run as
> `ok` with exit 0, so "exits 0 with no failures" is satisfied by a ruleset that has no tests at
> all. Requiring count-equals-rules also discharges REQ-A16-010: a rule shipped without cases makes
> the counts differ, and `sg test` is otherwise silent about it.

**AC-A16-002** — The harness rejects a rule that matches nothing
**Given** one existing rule temporarily rewritten so its pattern matches nothing,
**When** `sg test` is re-run,
**Then** it exits non-zero and names that rule in a `[Missing]`-class failure. The rule is
reverted; the final tree carries no mutant.

**AC-A16-003** — The harness rejects a rule that matches everything
**Given** one existing rule temporarily rewritten to a shape that also matches its own `valid` case,
**When** `sg test` is re-run,
**Then** it exits non-zero and names that rule in a `[Noisy]`-class failure. Both mutants are
required: `[Missing]` alone leaves the over-broad direction unguarded, which is the defect
`sec-csrf-no-token-check` exhibits.

**AC-A16-004** — The id-keying question is answered with measured evidence
**Given** the convention of repeating one rule id across languages,
**When** `design.md` §3.3 is read,
**Then** it records which keying `sg test` uses, the command that established it, and its observed
output — and if ids had to become per-language, the renamed ids are already in the tree with no
rule carrying a duplicated id under a keying that cannot distinguish them.

**AC-A16-005** — Rule-test assets live outside the distributed tree and still resolve
**Given** the final tree,
**When** `$T` is inspected and `sg test` is run,
**Then** no rule-test fixture or snapshot exists anywhere under `$T`, `$R` resolves outside the
template tree, and `sg test` passes using the repo-side `testConfigs` wiring.

**AC-A16-006** — The deployed config stays valid where the test root is absent
**Given** `$T/sgconfig.yml` after M1,
**When** it is read, and `sg scan --config` is run against a tree in which the `testConfigs`
`testDir` path does **not** exist — the state of every user project,
**Then** `sgconfig.yml` declares `testConfigs`, `ruleDirs` lists only directories shipping at least
one `sg test`-passing rule, and the scan completes without a configuration error.

> This is the residual risk of the D13 relocation. The fixtures move out of the distributed tree,
> but `sgconfig.yml` ships with a `testDir` pointing at a path the user does not have.

**AC-A16-007** — The corpus gate's real semantics are pinned by a test, not by prose
**Given** the repo,
**When** the pinning test `internal/hook/astgrep_corpus_pin_test.go` is run — the single new file
this SPEC adds under the `internal/hook/**` PRESERVE tree, per `plan.md` §A.4's named carve-out
and M1 item 6 —
**Then** it asserts, mechanically: (a) the covered-language validity gate calls `t.Skip` rather
than failing, so `spec.md` §A.7's enforcement table cannot silently go stale; and (b) the
differential corpus run did **not** skip — output contains `--- PASS` and contains neither
`--- SKIP` nor `corpus rejected:`.

> Clause (b) applies REQ-A16-017 inside this SPEC rather than leaving it as a rule written only
> for the successor. Without it, "the corpus test passed" stays compatible with twelve disabled
> assertions.

**AC-A16-008** — The corpus baseline is pinned unmodified
**Given** the repo,
**When** the same pinning test `internal/hook/astgrep_corpus_pin_test.go` is run,
**Then** it asserts that the twelve fixtures under `internal/hook/security/testdata/scan-corpus/`
and every recorded `wantDeny` value are byte-unchanged from `294b4b6ab`; and
`git diff --stat 294b4b6ab -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'`
is empty, so every pre-existing file under the tree is unmodified.

> The baseline is the pinned SHA, not `origin/main`. Against the branch name this criterion reports
> `18 files changed, 23 insertions(+), 2825 deletions(-)` on the untouched tree — `origin/main` is
> ten commits ahead of `294b4b6ab` — so it would be red for upstream's content rather than for
> anything this SPEC did or failed to do.

> A test rather than a reviewer's `git diff`, because with the validity gate silent (§A.7) editing
> a `wantDeny` column is a live way to turn red green, and that edit is invisible at close.
>
> The diff clause and the carve-out are one decision, not two. Adding a file under a tree declared
> read-only would otherwise make `plan.md` §E's standing untouchedness proof report a violation of
> the list it exists to prove — so the pathspec excludes exactly the one new file by name, leaving
> every other path still asserted byte-unchanged. Excluding one named path is strictly narrower
> than dropping the check.

---

## §B. Milestone 2 — coverage matrix and checker

**AC-A16-009** — The matrix key set equals the Cartesian product
**Given** `$M`,
**When** the checker compares the set of cell keys against (8 families × 14 parseable languages),
**Then** the two sets are equal — 112 keys, none missing, none duplicated, none substituted — and
the checker exits 0 on that class. A count alone does not satisfy this criterion.

**AC-A16-010** — Every cell resolves to exactly one of two states
**Given** `$M`,
**When** the checker inspects each cell,
**Then** every cell carries either a rule id or a rationale, and none carries both or neither.

**AC-A16-011** — Every exemption carries evidence
**Given** each EXPLICITLY EMPTY cell,
**When** the checker inspects its rationale,
**Then** it carries at least one of: a named language/stdlib/framework reference for the absent
construct, or a recorded probe invocation with its observed output.

**AC-A16-012** — No cell names a rule that does not exist
**Given** each cell carrying a rule id,
**When** the checker resolves that id against the shipped ruleset,
**Then** every id is found. `sg test` cannot catch this — it never reads the matrix.

**AC-A16-013** — The checker distinguishes its four failure classes
**Given** four synthetic defects introduced one at a time — (a) one cell deleted, (b) one cell
duplicated while another is deleted so the count stays 112, (c) a bare-assertion rationale with no
citation and no probe, (d) a cell naming a rule id absent from the ruleset —
**When** the checker runs on each,
**Then** it exits non-zero in all four cases and names a different failure class for each. Case (b)
is the load-bearing one: correct cardinality, wrong contents, and a count-based checker reports no
failure at all.

**AC-A16-014** — Excluded languages are recorded with version and evidence
**Given** `$M`'s excluded-languages record,
**When** it is read,
**Then** it names `r` and `flutter`, carries the verbatim probe refusal for each, names the
ast-grep version the probe ran under, and phrases the exclusion in the equal-opportunity idiom used
by `$T/sgconfig.yml` — with no wording ranking the language or implying permanence.

---

## §C. Milestone 3 — severity and the anchor

**AC-A16-015** — Every `error` is backed by a benign-shape negative case
**Given** each rule carrying `severity: error` in the final tree,
**When** its `sg test` cases are run,
**Then** it has a `valid` case containing a benign construct of the same shape as its `invalid`
case, and `sg test` reports zero findings on that `valid` case.

**AC-A16-016** — Every security rule's pattern is anchored to something real
**Given** each rule in a security family,
**When** its YAML and matrix cell are read,
**Then** it carries a `metadata.cwe` entry naming a weakness class, **and** its matched head symbol
is anchored by the same evidence class an EXEMPT cell owes: a named language / stdlib / framework
reference that documents the symbol, **or** a recorded probe showing the pattern matching real code
from that ecosystem.

> The CWE label alone is a reviewer-checkable label, not a mechanical check — a rule matching
> `acmeInternalVaultFetch($X)` carries a real CWE id, passes `sg test`, and matches nothing that
> exists. The citation-or-probe half is the mechanical part, and it is exactly the standard
> REQ-A16-005 already imposes on absence claims. A presence claim owes no less.

**AC-A16-017** — Shape matchers do not carry `error`, and the demotion is visible
**Given** `sec-csrf-no-token-check`,
**When** its severity and its matrix cell are read,
**Then** its severity is `warning`, and `$M`'s corresponding cell carries a recorded
precision-limitation annotation naming why it cannot satisfy REQ-A16-012 clause 2.

**AC-A16-018** — Severity across all 26 follows from evidence, not assertion
**Given** the 26 existing rules,
**When** each rule's severity and its `sg test` cases are inspected,
**Then** every severity follows from REQ-A16-012's predicate applied to that rule's own cases, and
every security rule sitting at `warning` carries its recorded reason.

---

## §D. Close

**AC-A16-019** — Neutrality holds across every artifact this SPEC creates or modifies
**Given** `$T` and `$R` — the ruleset directory and the rule-test root this SPEC authors —
**When** they are scanned for a SPEC ID, an internal date, or a commit SHA; every human-language
string is read (`message:`, `note:`, YAML comments, fixture prose); a ranking scan is run; each
file is compared against `.moai/astgrep-rules/`; and `make build` is run —
**Then** `$R` exists and carries a case pair per shipped rule; no SPEC-ID / date / SHA match is
found in either tree; every human-language string is English; the only ranking-word match is
`sgconfig.yml`'s equal-opportunity paragraph, which REQ-A16-002 requires the exclusion record to
inherit; no file is byte-identical to a dogfood-tree file; and `make build` leaves
`internal/template/catalog.yaml` unchanged.

> **Scope is the ruleset and rule-test trees, not `internal/template/templates/**`.** The wider
> tree carries 22 SPEC-ID files and 90 date-bearing files this SPEC never touches; scanning it
> would make the criterion red for someone else's content, which no work here can fix.
>
> **The CI neutrality guard is a different check, not this one.**
> `.github/workflows/template-neutrality-check.yaml` runs `TestTemplateNeutralityAudit` over
> classes C1/C2/C4/C5/C6/C8; a SPEC-ID is not among them, and the date and hash classes belong to a
> sibling test. Both must pass; they are not the same standard.

---

## §E. Definition of Done

All 19 criteria hold on the tree that will be pushed, each with its §H row showing red-now and
green-path, and:

- `sg test`, the matrix checker, and `go test ./internal/hook/...` are green in CI, not only
  locally.
- The PR body states the matrix's authored state: 112 keys, 14 cells seeded, 98 left to the
  successor.
- The PR body names `SPEC-ASTGREP-BREADTH-001` and states that card t228 requirements (2) and (4)
  are discharged there.

## §F. Out of scope for verification

- Rules for the ten uncovered languages and corpus fixtures for them — `SPEC-ASTGREP-BREADTH-001`.
- `r` / `flutter` rule behaviour — no parser at the pinned version.
- The `astGrepScanner.Scan` `CombinedOutput` defect — recorded in `spec.md` §A.7, excluded from
  repair by §D, owned by neither SPEC.
- Gate latency, `ast_grep_gate` mode defaults, the PreToolUse timeout mismatch.

## §G. Obligations relocated out of the criteria layer

Recorded rather than dropped. Each is a real obligation with no criterion that is both red now and
greenable by this SPEC's work; each is carried in `plan.md` as a build step or anti-pattern.

| Obligation | Why no criterion | Where it lives |
|---|---|---|
| `make build` regenerates a committed embed artifact | **Measured false.** `grep -c astgrep internal/template/catalog.yaml` → 0; the catalog tracks skills and agents only, and a ruleset edit produces no catalog diff. `//go:embed all:templates` emits no file. There is no artifact to assert. | Inverted into AC-A16-019's last clause (`make build` leaves `catalog.yaml` **unchanged**). `REQ-A16-018` and `plan.md` §D now state the same inverted form — both previously carried the falsified version, which made the requirement and this criterion assert opposite things. |
| Never mirror the dogfood tree | Passes today by construction and no work here changes it — a pure "don't" with no green path standing alone. | Folded into AC-A16-019 as a byte-identity clause; kept normative as REQ-A16-020 and as a `plan.md` §G anti-pattern. |
| No language is ranked | Same shape: the single sanctioned match exists today and nothing here moves it. | Folded into AC-A16-019, which is red because `$R` does not yet exist. |

---

## §H. Base-red evidence table

Every criterion, executed on the pre-implementation tree. Worktree `.claude/worktrees/t228`,
HEAD `294b4b6ab`, `ast-grep 0.40.5`, cwd = worktree root.

| AC | Command run today | Observed output | Why that is RED | GREEN PATH (milestone → output) |
|---|---|---|---|---|
| 001 | `sg test --config $T/sgconfig.yml` | `Running 0 tests` / `test result: ok. 0 passed; 0 failed;` EXIT=0 | 0 ≠ 26 shipped rules (`grep -rh '^id:'` → 26). The run is green and observes nothing. | **M1.3** authors 26 case pairs + `testConfigs` → `26 passed; 0 failed`, count equal to rule count. |
| 002 | same, after mutating one rule to match nothing | `Running 0 tests` … EXIT=0 — no `[Missing]` emitted | With no cases in existence the mutation has nothing to fail against; the assertion cannot fire. | **M1.5** — with cases present the mutant yields `[Missing]`, EXIT≠0. |
| 003 | same, after mutating one rule to match its own `valid` case | `Running 0 tests` … EXIT=0 — no `[Noisy]` emitted | Same: no `valid` case exists to be matched. | **M1.5** → `[Noisy]`, EXIT≠0. |
| 004 | `grep -c 'settled at M1\|not yet measured' design.md` | `2` | `design.md` §3.3 states the keying as an open question; no measured answer is recorded. | **M1.4** replaces it with the command and its output. |
| 005 | `find $T -name 'rule-tests' -o -name '*test*'` | *(no output)* | `$R` does not exist, so "assets live outside `$T` and `sg test` resolves them" has no subject. | **M1.2-1.3** create `$R` repo-side; `sg test` passes from it. |
| 006 | `grep -c testConfigs $T/sgconfig.yml` | `0` (rc=1) | `sgconfig.yml` declares no `testConfigs`, so the deployed-config-with-absent-testDir case does not yet exist to be checked. | **M1.2** adds `testConfigs`; the criterion then checks `sg scan` still completes where the path is absent. |
| 007 | `grep -rn 'func Test' internal/hook/pre_tool_scan_differential_test.go` | one function: `TestScanWriteContentDifferential` | No test pins the gate's skip semantics or asserts no-skip; §A.7's table is prose only and can go stale silently. | **M1 item 6** adds `internal/hook/astgrep_corpus_pin_test.go` (PRESERVE carve-out, `plan.md` §A.4) → asserts `t.Skip` semantics and `--- PASS` with no `--- SKIP` / `corpus rejected:`. |
| 008 | same | *(no pinning test)* | Nothing mechanically pins the twelve fixtures or their `wantDeny` values; a `wantDeny` edit is invisible at close. | **M1 item 6** adds the baseline assertion to `internal/hook/astgrep_corpus_pin_test.go`, and the §E diff excludes that one path so the untouchedness proof stays true. |
| 009 | `test -f $M` | `ABSENT` (rc=1) | No matrix exists, so no key set can equal the Cartesian product. | **M2.2** authors 112 keys; checker exits 0 on the key-set class. |
| 010 | `test -f $M` | `ABSENT` | No cells to resolve. | **M2.2-2.3** seeds 14, marks 98 pending; every cell carries one state. |
| 011 | `test -f $M` | `ABSENT` | No exemptions to evidence. | **M2.2** — each EXEMPT cell carries a citation or probe. |
| 012 | `test -f $M` | `ABSENT` | No cells naming rule ids; no checker to resolve them. | **M2.4** adds the dangling-id class; every named id resolves. |
| 013 | *(no checker binary/test)* | — | The checker does not exist, so none of the four classes can fire. | **M2.4-2.6** — four synthetic defects each produce a distinct class. |
| 014 | `test -f $M` | `ABSENT` | No excluded-languages record. | **M2.1-2.2** records `r` / `flutter` with verbatim probe output and the version. |
| 015 | `sg test --config $T/sgconfig.yml` | `Running 0 tests` … EXIT=0 | 12 `error`-severity rules exist and **0** have a `valid` case, so none is backed by a benign-shape negative. | **M1.3 + M3** — every `error` rule has a `valid` case producing zero findings. |
| 016 | `grep -rh 'cwe:' $T/security/*.yml \| wc -l` | `14` of 14 security rules | The CWE half **already passes** — vacuous alone. The anchor half is absent: 0 of 14 carry a citation or probe for their matched head symbol. | **M3 item 4** adds citation-or-probe per security rule, discharging REQ-A16-011's third clause; both halves then hold. |
| 017 | `grep -A2 'sec-csrf-no-token-check' $T/security/web.yml \| grep severity` | `severity: warning` | The severity half already holds — vacuous alone. The matrix annotation recording *why* does not exist (`$M` absent). | **M2.2 + M3.2** — the cell carries the precision-limitation annotation. |
| 018 | `sg test --config $T/sgconfig.yml` | `Running 0 tests` … EXIT=0 | Severity cannot "follow from that rule's own cases" when no rule has cases. | **M1.3 + M3.1** — all 26 severities justified by their own test cases. |
| 019 | `find $T -name '*test*'` ; `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}' $T/` | *(no output)* ; `0` | `$R` does not exist, so the composite's first clause ("`$R` exists and carries a case pair per shipped rule") fails. The SPEC-ID / date / ranking / dogfood / `catalog.yaml`-unchanged clauses are guards that hold today and ride this red — each is green now and must stay green, so none of them is what makes this row red. | **M1.3 + close** — `$R` exists with 26 pairs, all clauses scanned and clean, `make build` leaves `catalog.yaml` unchanged. |

### Criteria that could not be made red with a green path

Three, all resolved by relocation rather than deletion-without-record — see §G. In each case the
obligation survives; only its home changed. No requirement was dropped to fit, and no criterion was
kept that observes nothing.

The `catalog.yaml` obligation is the one worth naming twice: it was not merely unverifiable, it was
**factually wrong**. `internal/template/catalog.yaml` does not track this ruleset at all, so the
former criterion could not have been satisfied by any correct implementation — the predictable
response under deadline would have been to hand-edit the catalog to manufacture a diff.

## §I. Requirement-to-criterion traceability

Every requirement in `spec.md` §C maps to at least one criterion above. The mapping was built by
reading each criterion's body, not by matching titles, so a row here means the named criterion
actually asserts the requirement rather than merely sounding like it.

Coverage is many-to-many by design: a requirement stating a prohibition usually needs both a
positive criterion (the good state holds) and a mutant criterion (the bad state is rejected), which
is the same pairing §0 imposes on the criteria themselves.

[HARD] **A row must state any scope gap it carries.** Where a criterion's scope is narrower than
its requirement's, the row says so explicitly; an unflagged narrowing reads as full coverage and is
the same silence this table exists to end. Three rows previously carried an unflagged gap — two
scope narrowings and one criterion asserting *more* than its requirement — and all three were
closed by amending the requirement rather than by annotating the gap, which is the better fix when
the requirement is the side that is wrong.

| Requirement | Criteria | What carries the assertion |
|---|---|---|
| REQ-A16-001 all sixteen languages equal, within the ruleset + rule-test trees | AC-014, AC-019 | equal-opportunity idiom in the exclusion record; ranking scan at close. **Scopes now match**: the requirement was re-scoped from `internal/template/templates/**` to the two trees AC-019 scans. |
| REQ-A16-002 unparseable language recorded | AC-014 | names `r` / `flutter`, verbatim refusal, ast-grep version |
| REQ-A16-003 one cell per pair | AC-009 | set equality against 8 × 14, not a count |
| REQ-A16-004 cell is rule id XOR rationale | AC-010 | neither-nor and both-and are each rejected |
| REQ-A16-005 exemption records how it was checked | AC-011 | citation or recorded probe |
| REQ-A16-006 checker's four conditions | AC-013 | four synthetic defects, four distinct classes |
| REQ-A16-007 no omission, duplication, substitution, dangling id | AC-009, AC-012 | AC-013 case (b) is the load-bearing one: right count, wrong contents |
| REQ-A16-008 both test cases per rule | AC-001, AC-002, AC-003 | count-equals-rules, plus both mutant directions |
| REQ-A16-009 `sg test` reports zero failures | AC-001 | count comparison is the assertion, not the exit code |
| REQ-A16-010 no rule ships without cases | AC-001, AC-002 | discharged by count divergence — `sg test` is otherwise silent |
| REQ-A16-011 `metadata.cwe` anchor **and** citation-or-probe for the matched head symbol | AC-016 | CWE label plus the mechanical citation-or-probe half. **The requirement now carries both halves**; the criterion previously asserted more than any requirement stated, and `plan.md` M3 item 4 builds the half it asserts. |
| REQ-A16-012 two-clause `error` promotion | AC-015, AC-018 | benign same-shape `valid` producing zero findings |
| REQ-A16-013 everything else is `warning` | AC-018 | severity follows the predicate applied to the rule's own cases |
| REQ-A16-014 shape matchers never `error` | AC-017, AC-018 | AC-017 carries it for the one named rule (`sec-csrf-no-token-check` stays `warning`); AC-018 carries the general prohibition, since severity there must follow the predicate applied to each rule's own cases. **AC-003 was cited here and did not hold**: `[Noisy]` is a property of the *harness*, and asserts nothing about any rule's severity field. |
| REQ-A16-015 demoted security rule annotated | AC-017, AC-018 | the matrix cell records why clause 2 cannot be met |
| REQ-A16-016 claims about test enforcement are true | AC-007 (a) | pins that the validity gate skips rather than fails |
| REQ-A16-017 corpus evidence asserts no-skip | AC-007 (b), AC-008 | `--- PASS` with neither `--- SKIP` nor `corpus rejected:` |
| REQ-A16-018 template tree is the source of truth; `make build` leaves `catalog.yaml` unchanged | AC-019 | AC-019's last clause asserts exactly that — `make build` run, `internal/template/catalog.yaml` unchanged. **The requirement previously asserted the negation** (that the commit carry a *regenerated* catalog); it was inverted to the measured truth (`grep -c astgrep …` → 0), so requirement and criterion now agree instead of contradicting. |
| REQ-A16-019 fixtures outside the distributed tree | AC-005, AC-006 | relocation holds and `sg test` still resolves |
| REQ-A16-020 dogfood tree never mirrored | AC-019 | byte-identity comparison against `.moai/astgrep-rules/` |
| REQ-A16-021 no SPEC ID, date, or SHA in the ruleset or rule-test trees | AC-019 | scan over `$T` and `$R`. **Scopes now match**: the requirement said `internal/template/templates/**`, which is measured false on arrival (22 SPEC-ID files, 90 date files owned by others) and unfixable here; both layers are now the two trees this SPEC authors. |
| REQ-A16-022 `sg scan --config` completes without configuration error | AC-006 | run against a tree lacking the `testDir` path |
| REQ-A16-023 `ruleDirs` lists only vetted directories | AC-006 | asserted in the same criterion |

**One criterion has no direct requirement**, recorded rather than back-filled: AC-A16-004 (the
id-keying question) verifies a design decision that REQ-A16-008 and REQ-A16-009 depend on — whether
`sg test` can distinguish per-language rules sharing one id determines whether a per-rule case count
is even well defined. It is a precondition of those two requirements, not a requirement of its own,
and inventing a requirement to give it a row would be the reverse of the defect this table exists to
prevent.
