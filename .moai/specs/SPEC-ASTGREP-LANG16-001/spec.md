---
id: SPEC-ASTGREP-LANG16-001
title: "ast-grep ruleset contract: test harness, coverage matrix, and severity discipline"
version: "0.6.0"
status: draft
created: 2026-08-24
updated: 2026-08-25
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/template/templates/.moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "ast-grep, ruleset, security, severity, coverage-matrix, sg-test, template-first, contract"
tier: L
---

# SPEC-ASTGREP-LANG16-001 — ast-grep ruleset contract

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-24 | 0.1.0 | Initial draft (plan-phase). Card t228, Class C, Tier L. Built on `.moai/reports/t228/plan-measurements.md` (M1-M5). Operator decisions D1 (`sg test` first) and D2 (all families × 10 languages) bound as scope. |
| 2026-08-24 | 0.2.0 | Lead amendments A1-A3. A1: EXPLICITLY EMPTY cells owe a citation or probe (an absence claim is a claim). A2: `ast-grep 0.40.5` pinned at every parser-support claim. A3: measurement-citation convention (§A.0). |
| 2026-08-24 | 0.3.0 | Corrections C1-C7. C1: base SHA `a1b1ca696` → `294b4b6ab`. C7: the differential corpus is committed on `origin/main` (`a9eb896ce`, PR #1637, card t227) — two prior records reported it absent; `research.md` §8 records the cwd-drift error that produced both. t245 / PR #1643 deferral withdrawn; corpus brought into scope. |
| 2026-08-25 | 0.4.0 | **Plan-audit iteration 1 (FAIL 0.68) revision, and the split it forced.** **D8 (binding)**: re-scoped to the contract half — M1-M3 — with the breadth half (M4-M10) moved to `SPEC-ASTGREP-BREADTH-001`. Total scope unchanged; only sequencing. The forcing argument is budgetary: REQ and AC were saturated at 25/25, so the audit's own fixes could not be added without breaching the ceiling. **D1**: the corpus validity gate `t.Skip`s, it does not fail — four passages asserting the opposite corrected against `pre_tool_scan_differential_test.go:242`; REQ-A16-017 now binds every corpus-evidence claim to a no-skip assertion. **D2**: R2's mitigation withdrawn (a contrived rule passes `sg test` at exit 0); REQ-A16-011 adds the `metadata.cwe` anchor. **D3**: `SPEC-ASTG-UPGRADE-001` is archived — reconciled at every reference site; R6's re-probe owned here. **D5**: REQ-A16-012 reworded to `Where … shall`, counterexample preserved verbatim. **D7**: checker class 1 becomes set equality, plus a fourth class for a cell naming an absent rule id. **D4/D10/D11**: contiguous renumbering `001`…`023`; dangling `REQ-A16-022a` removed; tautological criterion retired. **D12/D13/D14/D15/D16**: `catalog.yaml` named; `rule-tests/` placed outside the distributed tree; neutrality widened to any human-language text; the unobservable "shall read" obligation moved to `plan.md`. |
| 2026-08-25 | 0.5.0 | **Plan-audit iteration 2 (FAIL 0.69) revision — `acceptance.md` only.** The criteria layer was rebuilt on a **two-column rule** (§0): every criterion carries both a RED-NOW observation executed against the pre-implementation tree and a GREEN-PATH milestone, which closes the vacuous-criterion class by construction. Criteria renumbered 23 → 19; §G records three obligations relocated out of the criteria layer rather than dropped; §H carries the 19-row base-red evidence table; §I adds requirement-to-criterion traceability. Recorded retroactively — this round shipped without a HISTORY row, which is itself the propagation failure v0.6.0 repairs. |
| 2026-08-25 | 0.6.0 | **Plan-audit iteration 3 (FAIL 0.71) — bounded propagation-only pass.** No re-authoring: v0.5.0's criteria-layer rewrite was never propagated into the other five artifacts, and every iteration-3 defect traces to that gap. **D1**: REQ-A16-018 inverted to the measured truth — `catalog.yaml` tracks no config path (`grep -c astgrep` → 0), so `make build` leaves it unchanged and no artifact accompanies a ruleset edit; `plan.md` §A.3 and §D corrected to match, ending a direct requirement-vs-criterion contradiction. **D3**: REQ-A16-021 and REQ-A16-001 re-scoped from `internal/template/templates/**` (22 SPEC-ID files, 90 date files, none owned here) to the ruleset + rule-test trees, matching AC-A16-019; the blind spot that let an impossible clause survive in the requirement layer is recorded under REQ-A16-021 and closed by a §0 extension. **D2**: REQ-A16-011 extended with the citation-or-probe half AC-A16-016 already asserted, and the matching work item added to `plan.md` M3. **D4**: the pinning test is given a path (`internal/hook/astgrep_corpus_pin_test.go`), an M1 work item, an explicit PRESERVE carve-out, and a restated §E diff command that stays true once it exists. **D5/D11**: `design.md` §3.1 and §4 now state the D13 placement the same way as §4's paragraph — unchanged across three iterations — plus the four-class count and three stale requirement citations. **D6**: nine un-renumbered AC references repaired across `plan.md` / `design.md` / `research.md`, two of them dangling. **D7**: four §I rows re-derived. **D8/D9**: `progress.md` §E.1 brought current; this HISTORY gap closed. **D10**: AC-A16-001's open count clause bound to a derivable number. |

---

## §A. Context

### A.0 Measurement-citation convention

[HARD] Every measured figure and every symbol citation in this SPEC names the tree it was taken
against. A symbol cited without its tree is not re-measurable: line numbers drift between trees.

| field | value |
|---|---|
| worktree | `.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab` |
| tool | `ast-grep 0.40.5` (`/opt/homebrew/bin/sg`) |

Tool-version pinning is load-bearing in the same way: parser support is a fact attached to a
version, so every parser-support claim names `ast-grep 0.40.5` explicitly.

This convention has already earned its cost twice. It caught a stale base SHA carried from the
primary checkout (v0.3.0), and `research.md` §8 records a cwd-drift measurement error it forced
into the open.

### A.1 The gap this SPEC and its successor address

The project treats sixteen programming languages as equals (`CLAUDE.local.md` §15). The
distributed ast-grep ruleset does not: of those sixteen, four carry any rule at all.

Measured inventory — `internal/template/templates/.moai/config/astgrep-rules/`, 26 rules:

| language | rules | of which security |
|---|---:|---:|
| go | 20 | 8 |
| javascript | 2 | 2 |
| python | 2 | 2 |
| typescript | 2 | 2 |
| the other twelve | 0 | 0 |

A user whose project is Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, or Swift receives
a scan surface that finds nothing while the gate reports itself enabled. That is worse than an
absent gate: a clean result reads as a clean codebase.

### A.2 Scope of THIS SPEC — the contract, not the breadth

Closing that gap takes two SPECs in sequence. **This one builds the contract**: a working
`sg test` harness over the existing 26 rules, a coverage matrix whose completeness is
mechanically checkable, an anti-contrivance anchor, and a severity discipline that decides which
rules may refuse a write. **`SPEC-ASTGREP-BREADTH-001` writes the rules** — up to 80 across ten
languages, plus the differential-corpus fixture pairs — against this contract.

The split is a sequencing decision, not a scope reduction. Card t228 requirements (2) and (4) are
delivered in full across the two SPECs; §D states which SPEC discharges each.

Contract-first is the right order for a measured reason, not an aesthetic one: two of this SPEC's
own defenses were falsified during plan-audit (§A.8). Had 80 rules already been written against
them, the correction would have invalidated 80 rules instead of one requirement.

### A.3 Which of the twelve languages are reachable — as of ast-grep 0.40.5

Probe against **ast-grep 0.40.5**, on the tree of §A.0, per language
`echo foo | sg run -p foo -l <lang> --stdin` — record at
`.moai/reports/t228/plan-measurements.md` M1, re-measured independently during plan-audit:

- **Parses under 0.40.5 (10)**: rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift
- **No parser in 0.40.5 (2)**: `r` — `error: invalid value 'r' for '--lang <LANG>': r is not supported!`;
  `flutter`/dart — `error: invalid value 'dart' for '--lang <LANG>': dart is not supported!`

The two exclusions are a property of the pinned build, not a ranking, and not permanent. A later
ast-grep shipping either parser makes those languages reachable with no change to this reasoning.

### A.4 Security families and their present coverage

Eight families exist in the shipped ruleset. Measured id / language / severity triples:

| # | family | rule id(s) | languages today | severity today |
|---|---|---|---|---|
| F1 | command injection | `sec-command-injection-shell` / `-exec` | go, python / javascript, typescript | error |
| F2 | hardcoded credential | `sec-hardcoded-credential` | go, python, javascript, typescript | error |
| F3 | weak hash | `sec-weak-hash-md5` | go | error |
| F4 | hardcoded API key | `sec-hardcoded-api-key` | go | error |
| F5 | hardcoded JWT signing key | `sec-hardcoded-jwt-signing-key` | go | error |
| F6 | CSRF token absence | `sec-csrf-no-token-check` | go | warning |
| F7 | log injection | `sec-log-injection-unsanitized` | go | warning |
| F8 | template injection / XSS | `sec-template-injection-html` | go | **error** |

Fourteen (family, language) security cells are filled. Every other cell is empty, and none of
those empties is **recorded** anywhere — an unrecorded empty and a deliberate empty are
indistinguishable today. The matrix (§C.2) is what separates them.

### A.5 What severity means mechanically

The t227 principle — `error` means a write is refused — is not a convention. It is wired, and the
chain was traced end to end during plan-audit:

`internal/hook/pre_tool.go` calls `h.scanner.ShouldAlert(result)` and returns `DecisionDeny` with
the finding report as reason; `internal/hook/security/scanner.go` delegates to
`reporter.ShouldExitWithError`; `internal/hook/security/reporter.go` is
`return result.ErrorCount > 0`. Warning-only results fall through to allow. The scanned config
resolves to the project's `.moai/config/astgrep-rules/sgconfig.yml`
(`internal/hook/security/rules.go`) — exactly where the template deploys, so a template severity
promotion refuses writes in **every user project**.

Citations are token-anchored (`ShouldAlert` → `ErrorCount > 0` → `DecisionDeny`), not
line-anchored.

This makes breadth and precision a single question. A security rule matching a construct *shape*
rather than a specific dangerous *construct* refuses writes across an entire category of correct
code. §C.4 turns that into a testable predicate.

### A.6 The differential corpus — present, and what it actually enforces

`internal/hook/security/testdata/scan-corpus/` is **committed on `origin/main`**, landed by
`a9eb896ce` (PR #1637, card t227) — the same card this SPEC inherits its severity principle from.
Four confirmations on the tree of §A.0:

| # | Command | Observed |
|---|---|---|
| 1 | `git ls-tree -r --name-only origin/main \| grep scan-corpus` | 12 paths listed |
| 2 | `git log --oneline origin/main -- '*scan-corpus*'` | `a9eb896ce` — sole commit touching the path, no deletion after |
| 3 | `git merge-base --is-ancestor a9eb896ce HEAD` | exit 0 |
| 4 | `git ls-files` → 12; `git status --porcelain` → 0 | tracked and clean here |

Two prior plan-phase records reported this path absent, for two different reasons. Both were
wrong; `research.md` §8 records the cwd-drift mechanism that produced them.

### A.7 [CORRECTED] The corpus validity gate SKIPS — it does not fail

Earlier revisions of this SPEC stated in four places that the covered-language validity gate
**fails** when a covered language has no denying fixture, and credited that failure as a forcing
function. **That was false.** Measured at `internal/hook/pre_tool_scan_differential_test.go:242`:

```go
    t.Skip(b.String())
}

for _, fx := range scanCorpus {      // line 245 — every assertion lives below the skip
```

The gate calls `t.Skip`, positioned **after** the observation loop and **before** the assertion
loop. Three consequences follow, and they are why this correction is load-bearing rather than
editorial:

1. A skip makes `go test` report `ok`. There is no red.
2. Because the skip precedes the assertion loop, it disables **all twelve** recorded differential
   assertions at once — not merely the one language that lacked a fixture.
3. `coveredCorpusLanguages` is therefore an **escape hatch**: adding a language to it without a
   denying fixture turns the entire differential test green-by-skip.

The skip's own message names the cause — `astGrepScanner.Scan` uses `CombinedOutput`, and
`sg scan --json` writes `Error: N error(s) found in code.` to stderr on any error-severity
finding, corrupting the JSON so no error finding parses: *"The pre-write gate can warn but cannot
deny."* That is a live defect in the tree. §D forbids this SPEC from fixing it.

**What the corpus test does enforce**, stated plainly so no later SPEC over-credits it:

| Enforcement | Status |
|---|---|
| Per-row `wantDeny` mismatch → `t.Errorf` | **Real.** This is the genuine forcing function. |
| Deny row with empty reason → `t.Errorf` | **Real.** |
| Allow row with non-empty reason → `t.Errorf` | **Real.** |
| Covered language lacking a denying fixture → failure | **Does not exist.** It is a skip. |

On the tree of §A.0 the test currently **runs** rather than skips — all four covered languages
have denying fixtures — and passes in ~1.1s. The hazard is prospective, and it lands on the
successor SPEC, which is where REQ-A16-017 aims it.

### A.8 Two defenses this SPEC claimed, and lost

Recorded because both were load-bearing, and because a SPEC that quietly drops a falsified claim
teaches nothing.

**Falsified — "a contrived pattern has no plausible benign counterpart to write."** This was R2's
entire mitigation. Counterexample, built in one line during plan-audit:

```yaml
id: probe-contrived
language: go
severity: error
rule:
  pattern: zzzNeverRealApi($X)
# valid:   func ok()  { zzzSafeApi(x) }
# invalid: func bad() { zzzNeverRealApi(x) }
```
```
$ sg test --config /tmp/t228sgt/sgconfig.yml
PASS probe-contrived  .U
test result: ok. 1 passed; 0 failed;    EXIT=0
```

A rule naming an API that does not exist, paired with a benign counterpart that also does not
exist, satisfies both severity clauses, qualifies for `error`, and records as IMPLEMENTED.
Nothing anchored a rule to a *real* dangerous construct. REQ-A16-011 adds that anchor.

**Falsified — the corpus validity gate as a forcing function.** §A.7 above.

Both failures share a shape: a defense that reads convincingly in prose and is never executed.
That is precisely what this SPEC's contract exists to prevent for the 80 rules to come, and it is
the strongest available argument for building the contract before the breadth.

---

## §B. Scope

**In scope**: the `sg test` harness and its fixtures; the coverage-matrix document and its
mechanical checker; the severity reclassification of the existing 26 rules; the `metadata.cwe`
anchor convention; the placement decision for rule-test assets; and the Template-First,
neutrality, and embed obligations any template edit carries.

**Not in scope**: authoring rules for the ten uncovered languages, and extending the differential
corpus. Both belong to `SPEC-ASTGREP-BREADTH-001` (§D).

---

## §C. Requirements

Notation: GEARS. `<subject>` is the named artifact or actor.

### C.1 Language coverage and equality

**REQ-A16-001** (Ubiquitous) — The ruleset shall treat all sixteen supported languages as equals.
No file under the ruleset directory `internal/template/templates/.moai/config/astgrep-rules/**` or
under the repo-side rule-test root shall label any language primary, planned, unsupported, or
otherwise ranked. The one sanctioned exception is `sgconfig.yml`'s equal-opportunity paragraph,
which REQ-A16-002 requires the exclusion record to inherit.

> Scope narrowed from `internal/template/templates/**` for the reason recorded under REQ-A16-021:
> the prohibition can only bind the trees this SPEC authors, and a prohibition stated over content
> nobody here can change is unfalsifiable rather than strict.

**REQ-A16-002** (Where) — Where a language has no parser in the pinned ast-grep version, the
coverage matrix shall record that language as tool-excluded, naming the version (`ast-grep
0.40.5`) and carrying the verbatim probe output as its rationale, phrased in the
equal-opportunity idiom already present in `sgconfig.yml`. The excluded set under 0.40.5 is `r`
and `flutter`.

### C.2 The coverage matrix — scope integrity

The matrix is the artifact that makes "we covered the languages" a statement someone can disagree
with. It is authored here and **filled** by the successor SPEC.

**REQ-A16-003** (Ubiquitous) — The coverage matrix shall carry exactly one cell for every
(security family, parseable language) pair — eight families × fourteen parseable languages = 112
cells — and shall record the ast-grep version its language axis was derived under.

**REQ-A16-004** (Ubiquitous) — Every matrix cell shall carry **either** a rule id, marking it
IMPLEMENTED, **or** a one-line rationale naming the construct that is missing from that language,
marking it EXPLICITLY EMPTY. No third state exists.

**REQ-A16-005** (Ubiquitous) — Every EXPLICITLY EMPTY cell shall additionally record how its "no
equivalent construct exists" claim was checked, as at least one of: a **citation** identifying the
language, standard-library, or framework reference relied on; or a **probe** — the invocation run
and its observed output, recorded verbatim.

**REQ-A16-006** (When) — When the matrix checker runs, and any of four conditions holds, it shall
exit non-zero and name the offending cells, reporting the four classes distinguishably:

1. **key-set mismatch** — the set of cell keys is not equal to the Cartesian product of the two
   axes. A set comparison, **not** a count.
2. **unresolved cell** — a cell carrying neither a rule id nor a rationale.
3. **unevidenced exemption** — an EXPLICITLY EMPTY cell whose rationale carries neither a citation
   nor a probe record.
4. **dangling rule id** — a cell naming a rule id not present in the shipped ruleset.

**REQ-A16-007** (Unwanted) — A cell shall not be omitted, duplicated, or substituted, and an
EXPLICITLY EMPTY cell shall not rest on a bare assertion.

> **Why class 1 is a set comparison.** A count answers "how many cells", which is not the
> question. The question is "which cells", and the two differ exactly when a cell has been
> substituted — `F1/rust` duplicated while `F1/java` disappears. That matrix passes a cardinality
> check, passes the per-cell resolution check, and reports 112/112 while a family silently loses a
> language. A substituted cell is worse than a missing one, because it presents as complete.

> **Why class 4 exists.** The matrix is an authored document, so it can drift from the ruleset it
> describes. `sg test` never reads the matrix, and no other check reads both — so a cell naming a
> rule that was renamed or never written would otherwise pass every gate in this SPEC.

### C.3 Rule quality — every rule is exercised, and anchored

**REQ-A16-008** (Ubiquitous) — Every IMPLEMENTED cell's rule shall carry both a `valid` and an
`invalid` `sg test` case. The `invalid` case shall contain the dangerous construct; the `valid`
case shall contain a benign construct of the same shape.

**REQ-A16-009** (When) — When `sg test` runs against the shipped ruleset, it shall report zero
failures and zero missing snapshots.

**REQ-A16-010** (Unwanted) — A rule shall not ship without both test cases. A rule that matches
nothing is indistinguishable from an absent rule and worse than one, because the matrix records
it as covered.

**REQ-A16-011** (Ubiquitous) — Every security-family rule shall carry a `metadata.cwe` entry
naming the weakness class it detects; its `invalid` test case shall instantiate that weakness
class in idiomatic code for the rule's language — a construct a competent practitioner of that
language would recognize as real; **and** the symbol its pattern matches at the head shall be
anchored by at least one of: a **citation** identifying the language, standard-library, or
framework reference that documents the symbol, or a **probe** — an invocation showing the pattern
matching real code from that ecosystem, recorded with its observed output.

> **This is the anchor R2 lacked** (§A.8). `metadata.cwe` is not new ceremony: every existing
> security rule already carries `metadata.owasp` and `metadata.cwe`, and `research.md` §5 already
> recommended new rules follow. This requirement elevates that convention from a recommendation to
> a checkable obligation, because naming CWE-78 forces the author to point at a weakness that
> exists in the world before writing a pattern — which `zzzNeverRealApi` cannot survive.
>
> The CWE label alone is **necessary, not sufficient**, and the third clause is why this
> requirement does not stop there. A rule matching `acmeInternalVaultFetch($X)` can carry a real
> CWE id, pass `sg test`, and match nothing that exists in the world — the label is
> reviewer-checkable, not mechanical. The citation-or-probe clause is the mechanical half, and it
> imposes on a **presence** claim exactly the standard REQ-A16-005 already imposes on an
> **absence** claim: name the reference relied on, or record the probe. A presence claim owes no
> less than an absence claim, and this SPEC was previously asymmetric on that point.

### C.4 Severity — reclassification, not relaxation

**REQ-A16-012** (Where) — **Where** a rule is promoted to severity `error`, the promotion
**shall** satisfy **both** of:

1. the rule belongs to a security family; **and**
2. a `valid` test case exists containing a benign construct of the **same shape** as the `invalid`
   case, on which the rule produces **zero** findings.

This binds every promotion — every rule, every family, every language, without exception. Clause 2
is a precondition on the promotion itself, not a caveat attached to any particular rule.
Membership in a security family is necessary and **not sufficient**: a rule satisfying clause 1
alone shall carry `warning`.

> **"Security implies error" is false, and `sec-csrf-no-token-check` is the counterexample —
> not an exception to be carved out.** Its pattern,
> `func $HANDLER(w http.ResponseWriter, r *http.Request) { $$$BODY }`, is the shape of every Go
> HTTP handler, CSRF-protected or not. It is unambiguously a security rule, so the one-clause
> rule would promote it; and because `error` returns `DecisionDeny` on the write path (§A.5),
> that promotion would refuse `Write` and `Edit` on **every Go HTTP handler in the project**. A
> gate that refuses correct code gets switched off, and then it protects nothing. Clause 2
> exists because the one-clause rule is wrong in general — this rule is where that shows up
> first, not the only place it would bite.
>
> Plan-audit confirmed the match is AST-based rather than textual: the pattern matched both a
> two-space-indented and a gofmt tab-indented benign handler, so formatting does not narrow it.
> "Every Go HTTP handler" is literal for any handler with a non-empty body.

**REQ-A16-013** (Ubiquitous) — Every rule not satisfying REQ-A16-012 — every idiom and style rule,
and every security rule failing clause 2 — shall carry severity `warning`.

**REQ-A16-014** (Unwanted) — A rule whose pattern matches a construct *shape* rather than a
specific dangerous construct shall not carry severity `error`, regardless of family membership.

**REQ-A16-015** (While) — While a security rule sits at `warning` under REQ-A16-014, the coverage
matrix shall record the precision limitation as that cell's annotation, so the demotion is visible
rather than inferred from a severity field.

> Reclassification moves the total in both directions and is never a relaxation: a rule leaves
> `error` when it lacks the evidence and enters `error` when it acquires it. What is fixed is the
> predicate, not the count. A demoted rule still reports; it simply stops refusing the write.

### C.5 Honest enforcement accounting

**REQ-A16-016** (Ubiquitous) — Every statement this SPEC or its successor makes about what a test
enforces shall match that test's measured behaviour. A test that skips shall not be described as
failing.

**REQ-A16-017** (When) — When any acceptance criterion cites the differential corpus as evidence,
it shall additionally assert that the run did **not** skip — the output contains `--- PASS` and
contains neither `--- SKIP` nor `corpus rejected:`. A skipped run is a gap, never a pass.

> This exists because the skip is silent by construction (§A.7). `go test` prints `ok` for a
> skipped package, and the skip message is visible only under `-v`. Without an explicit no-skip
> assertion, "the corpus test passed" is compatible with twelve disabled assertions. The successor
> SPEC inherits this requirement and applies it to every corpus criterion it writes.

### C.6 Template-First, neutrality, and configuration

**REQ-A16-018** (When) — The source of truth shall be
`internal/template/templates/.moai/config/astgrep-rules/`, and when it changes, `make build` shall
be run so the binary carries the edit — and that build shall leave `internal/template/catalog.yaml`
**unchanged**. No regenerated embed artifact accompanies a ruleset edit.

> **[CORRECTED] There is no artifact to carry, and the earlier requirement said there was.**
> Prior revisions mandated that the commit carry a regenerated `internal/template/catalog.yaml`.
> Measured on the tree of §A.0: `grep -c astgrep internal/template/catalog.yaml` → **0**, and
> `grep -c 'astgrep-rules' internal/template/catalog.yaml` → **0**. The catalog holds skill and
> agent entries with content hashes; it tracks no config path, so a ruleset edit produces no
> catalog diff and there is nothing for the commit to carry.
>
> The obligation that survives is the build itself. `internal/template/embed.go` uses
> `//go:embed all:templates`, which generates no file — the template tree is compiled directly
> into the binary — so `make build` is what makes a template edit take effect, and a reader who
> expects a committed artifact from it is reasoning from the general repo convention rather than
> from this path. Stating the negative explicitly is what stops that expectation from being
> re-inserted: an implementer told to "regenerate the catalog" and finding no diff has one
> obvious way to comply, and hand-editing the catalog to manufacture one is worse than the
> omission the old requirement was written to prevent.

**REQ-A16-019** (Ubiquitous) — Rule-test fixtures and their snapshots shall live **outside** the
distributed template tree, with `testConfigs` wired repo-side.

> Everything under `internal/template/templates/.moai/config/astgrep-rules/` deploys into every
> user project. Placing ~90 fixture files there would ship them to every user — and by
> REQ-A16-008 those fixtures contain deliberately dangerous constructs by construction, landing
> inside the very tree the deployed ruleset scans. The user's own pre-write gate would then fire
> on our test assets. Test assets are not part of the distributed product.

**REQ-A16-020** (Unwanted) — The local dogfood tree `.moai/astgrep-rules/` shall not be mirrored
verbatim into the template. It is experimental, carries mixed-locale messages and SPEC IDs, and
would violate REQ-A16-021 on contact.

**REQ-A16-021** (Ubiquitous) — No file under the ruleset directory
`internal/template/templates/.moai/config/astgrep-rules/**` or under the repo-side rule-test root
this SPEC creates shall contain a SPEC ID, an internal date, or a commit SHA, and **any
human-language text** in either tree — rule messages, notes, comments, and fixture prose alike —
shall be English.

> **[CORRECTED] The scope was `internal/template/templates/**` and was measured false on
> arrival.** 22 files under that tree carry a SPEC ID and 90 carry an ISO date, none of them owned
> by this SPEC and none of them fixable by any work here. Over the two trees this SPEC actually
> authors, the same scan returns **0**.
>
> **Why this correction matters beyond its own text.** `acceptance.md` §0's two-column discipline
> — every criterion owes both a red-now and a green-path — binds the **criteria** layer only. An
> obligation that is red today and red forever is caught there and nowhere else, so the same
> impossible clause becomes invisible the moment it is written as a requirement instead. That is
> exactly what happened here: the criterion was re-scoped and the requirement was not, and the
> requirement layer had no discipline to catch what the criteria layer would have rejected. §0
> now states the extension explicitly; this note records the blind spot that made it necessary.

**REQ-A16-022** (When) — When `sg scan --config sgconfig.yml <path>` runs against this repository,
it shall complete without a configuration error.

**REQ-A16-023** (Ubiquitous) — `sgconfig.yml` `ruleDirs` shall list only directories shipping at
least one rule that passes `sg test`.

---

## §D. Exclusions

Normative. Where a successor owns an item, the successor is named.

### Out of Scope — per-language rule breadth (successor: SPEC-ASTGREP-BREADTH-001)

- Authoring security rules for rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift —
  up to 80 rules across the eight families — and the per-language idiom rules.
- Filling the 98 unresolved matrix cells. This SPEC authors the matrix, its axes, and its checker;
  the successor resolves each cell to a rule id or an evidenced rationale.
- **This is sequencing, not deferral.** Card t228 requirement (2) is discharged by
  `SPEC-ASTGREP-BREADTH-001` against the contract built here. Both halves are committed work with
  no external gate between them.

### Out of Scope — differential corpus extension (successor: SPEC-ASTGREP-BREADTH-001)

- Adding fixture pairs for newly-covered languages, extending `coveredCorpusLanguages`, and
  promoting `java_uncovered.java` / `rs_uncovered.rs` to `_deny_*` rows.
- **Card t228 requirement (4) is discharged there**, under REQ-A16-017's no-skip obligation, which
  this SPEC establishes precisely so the successor cannot inherit the silent-skip hazard.
- This SPEC touches no file under `internal/hook/security/testdata/scan-corpus/`.

### Out of Scope — fixing the scanner's deny capability

- Repairing `astGrepScanner.Scan`'s `CombinedOutput` JSON corruption, which the skip message at
  `pre_tool_scan_differential_test.go:242` names as the cause of the gate's skip condition (§A.7).
- This SPEC **records** the defect and refuses to describe the gate as stronger than it is
  (REQ-A16-016). It does not fix it: that is a change to the hook's scan path, a different
  subsystem with its own risk surface. A separate card should own it, and this SPEC's honest
  accounting is what makes the case for one.

### Out of Scope — gate behaviour and hook wiring

- Changing `ast_grep_gate` defaults, the advisory/blocking mode, the PreToolUse scan path, or the
  deny decision in `internal/hook/pre_tool.go`. §A.5 reads that wiring; it does not modify it.
- Changing the `sg` CLI's `--rules-dir` default or `gate.yaml`'s `ast_grep_gate.rules_dir`.

### Out of Scope — the local dogfood tree

- Cleaning, expanding, or promoting `.moai/astgrep-rules/`, owned by
  SPEC-ASTGREP-DOGFOOD-CLEANUP-001 (completed). This SPEC touches it only to the extent of
  REQ-A16-020: it is not a mirror source.

### Out of Scope — new security families

- Inventing families beyond the eight measured in §A.4. Breadth is across the language axis; the
  family axis stays fixed so the matrix has a stable shape.

### Out of Scope — ast-grep version work

- Upgrading or re-pinning ast-grep. `SPEC-ASTG-UPGRADE-001` is **archived** and owns nothing
  active; this SPEC absorbs no part of it and pins `0.40.5` independently. R6 owns the re-probe
  obligation here, so the pinned-version question has a live owner rather than an archived
  reference.

---

## §E. Success criteria

Full enumeration in `acceptance.md`. In summary: `sg test` passes over all 26 existing rules and
demonstrably fails on both mutants; the matrix has 112 correctly-keyed cells with a checker
distinguishing four failure classes; every `error` severity is backed by a benign-shape negative
case and a `metadata.cwe` anchor; rule-test assets sit outside the distributed tree; and the
neutrality, embed, and config-mode obligations hold.

---

## §F. Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | Rule id collision under `sg test` — the shipped convention repeats one id across languages (`sec-hardcoded-credential` ×4); snapshot keying may not tolerate it. | M1 determines this empirically before any rule is authored. Fallback: per-language suffixed ids. Whichever way M1 resolves, the answer binds the successor SPEC. |
| R2 | **[MITIGATION WITHDRAWN]** A contrived rule fills a cell and passes every gate. The former mitigation — "a contrived pattern has no plausible benign counterpart to write" — was **falsified** (§A.8): the counterexample took one line and exited 0. | Replaced by REQ-A16-011's `metadata.cwe` anchor plus the idiomatic-instantiation obligation. Stated as necessary-not-sufficient rather than as a closure, because the honest reach of the anchor is a raised floor, not a proof. |
| R3 | Promoting a shape matcher to `error` refuses writes across correct code (F6 is the live instance). | REQ-A16-012's two-clause predicate and REQ-A16-014's prohibition, enforced per-rule by the negative test case. |
| R4 | The matrix reaches 112/112 on rationales nobody checked. | REQ-A16-005's citation-or-probe obligation, enforced as a distinct checker class (REQ-A16-006 class 3). |
| R5 | The matrix drifts from the ruleset it describes — a cell names a rule that does not exist. | REQ-A16-006 class 4. `sg test` cannot catch this: it never reads the matrix. |
| R6 | ast-grep version drift changes parser availability, invalidating §A.3's 0.40.5-scoped sets. **This SPEC owns the obligation**, since `SPEC-ASTG-UPGRADE-001` is archived. | The probe is re-run at M2 against the version then installed and the excluded set re-derived rather than copied; the matrix records the version it was derived under (REQ-A16-003). |
| R7 | An absence claim measured from a drifted working directory reads as a confirmed absence, silently. Not hypothetical: it invalidated two rounds of scoping on this SPEC (`research.md` §8). | Measure ref contents from the repository root or with `--full-tree`; use repo-relative pathspecs, not globs; treat rc=1-with-empty-output as unproven. Every absence claim states the directory it was run from. |
| R8 | The successor over-credits the corpus gate, as this SPEC did in three prior revisions. | REQ-A16-016 and REQ-A16-017, plus §A.7's enforcement table, are inherited by the successor as contract. |

---

## §G. Assumptions

1. `ast-grep 0.40.5` remains pinned for this SPEC's duration, so §A.3's sets hold. R6 covers drift.
2. The eight families of §A.4 are the complete family axis.
3. `sg test`'s `testConfigs` mechanism behaves on the shipped ruleset as it did in the prototype
   and in plan-audit's independent reproduction of both mutants. M1 converts this to measurement.

---

## §H. Cross-references

- `SPEC-ASTGREP-BREADTH-001` — the successor carrying rule breadth and corpus extension (§D).
- `SPEC-ASTGREP-MULTILANG-001` (completed) — landed the 26-rule curated baseline this extends.
- `SPEC-ASTGREP-DOGFOOD-CLEANUP-001` (completed) — owns the local dogfood tree.
- `SPEC-ASTG-UPGRADE-001` — **archived**; owns no active work. Recorded here only to state that
  this SPEC does not depend on it (§D, R6).
- `.moai/reports/t228/plan-measurements.md` — M1-M4 measured basis (M5 superseded).
- `.moai/reports/t228/plan-audit-iter1.md` — the audit driving v0.4.0.
- `.moai/reports/t228/plan-audit-iter2.md` — the audit driving v0.5.0's two-column criteria rule.
- `.moai/reports/t228/plan-audit-iter3.md` — the audit driving v0.6.0's propagation pass.
- `internal/hook/pre_tool.go` — the `ShouldAlert` → `ErrorCount > 0` → `DecisionDeny` chain (§A.5).
- `internal/hook/pre_tool_scan_differential_test.go:242` — the `t.Skip` corrected in §A.7.
- `a9eb896ce` (PR #1637, card t227) — landed the corpus and the deny-capability restoration.
