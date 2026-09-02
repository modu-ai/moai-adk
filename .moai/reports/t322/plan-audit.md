# SPEC Review Report: SPEC-GRAPH-FRESHNESS-CADENCE-001 (card t322)

Iteration: 1/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.82** (Tier M PASS threshold 0.80)

Audited in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch
`WT-graph-freshness-cadence`, HEAD `c28edd9de` (confirmed by `git rev-parse --show-toplevel`).
Reasoning context ignored per M1 Context Isolation; the four SPEC artifacts are the input surface.

Three defects are **blocking** and must be fixed before Implementation Kickoff Approval. One of them
(D1) would, if the plan is followed literally, silently disable a gate the SPEC declares out of scope.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-GFC-001..012, sequential, no gap, no duplicate,
  consistent 3-digit padding (`spec.md` §C).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-GFC-*` in
  `spec.md` §C). All twelve match a GEARS pattern: Ubiquitous (001, 002, 003, 005, 010, 011),
  Where (004, 012), Unwanted `shall not` (006, 009), Event-driven `When …` (007, 008). The
  Given-When-Then entries in `acceptance.md` are the verification layer and are graded under
  Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`id`, `title`, `version: "0.1.0"` quoted, `status: draft`, `created`/`updated` ISO,
  `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-string).
  No rejected snake_case alias. Extras (`era`, `tier`, `related_specs`) are additive.
- **[PASS/N-A] MP-4 language neutrality** — N/A: single-language SPEC (Go tooling internal to this
  repository). REQ-GFC-012 correctly carries the Template-First mirror obligation.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — three referenced SPECs all exist and all read
  `status: completed`; none is retired/superseded/archived.
  Evidence: `grep -m1 '^status:' .moai/specs/SPEC-{V3R6-GRAPH-FRESHNESS-001,V3R6-GRAPH-FRESHNESS-002,STAMP-REACHABILITY-001}/spec.md`
  → `completed` ×3.
- **[PASS/N-A] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/`
  → rc=1, no matches. (`research.md` absent — correct for the Tier M artifact set.)

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | Judgments explicit, each with rejected alternatives (`spec.md` §D). Two imprecisions: D5 (p90 label), D6 (consumer phrasing). |
| Completeness | 0.75 | 0.75 | All sections present; §E Out of Scope carries five well-formed `### Out of Scope — <topic>` H3s with specific bullets. Missing: the cumulative-crossing cadence measurement (D2), and the `meta.go` collateral consumer (D1). |
| Testability | 0.70 | 0.50–0.75 | 13 ACs, most binary with a named deciding command; two carry named mutants. But D3 (AC-GFC-003 cannot fail on an implementation defect), D4 (AC-GFC-005 under-enumerates), D7 (AC-GFC-007's command does not decide its claim). |
| Traceability | 0.92 | 0.75–1.0 | All 12 REQs have ≥1 AC; all 13 ACs cite an existing REQ; milestone column present. REQ-GFC-011's coverage is partial in depth (D4). |

---

## Baseline reproduction (independent, this tree)

Every §B figure the SPEC states was re-measured here. All reproduce except one.

| Claim | Command | Observed | Verdict |
|---|---|---|---|
| 397 tracked files under a `testdata` segment | `git ls-files internal cmd pkg` + segment match | `397` | reproduces |
| 3,628 tracked / 998 production Go | `git ls-files …` + predicate | `3628` / `998` | reproduces |
| streak cumulative 65 | `git diff --name-only 9326b5478d… 48eb945df -- internal cmd pkg \| grep -c .` | `65` | reproduces |
| counterfactual 65 → 2 | same, + `.go` / `-_test.go` / `-/testdata/` / `-templates/` | `2` | reproduces |
| 51 described-worthy over 50 commits | `git log -50 --name-only …` + predicate + `sort -u` | `51` | reproduces |
| 12 of 30 integrations contribute 0 | 30-integration walk | `12` | reproduces |
| median 2 / mean 3.9 / max 29 | same distribution | `2` / `3.9` / `29` | reproduces |
| **p90 = 11** | same distribution | nearest-rank **9**, linear **9.2** (11 is rank 28 ≈ p93) | **does not reproduce** (D5) |
| 0 cited `testdata` paths; template tree cited once as bare dir | citation extraction over `.moai/project/codemaps/*.md` | `0`; `internal/template/templates/` once | reproduces (80 → I count 78; regex variance, not material) |
| no stamp attachment point | `grep -rln "graph stamp" .claude/`; `grep -rn … .github/` | rc=1 both | reproduces |
| all three failing integrations are true merges | `git rev-list --parents -n1` ×3 | 3 fields each | reproduces |
| `graph stamp --commit` delivered | `internal/cli/graph_stamp.go:130`, example at `:68` | flag present, example verbatim | reproduces (t292's "absent" was true of `main`, not of this lineage) |

---

## Defects Found

**D1. Predicate inside `aggregateFingerprint` silently disables the edges layer — `plan.md` §E M1 —
Severity: critical — Class: blocking**

`plan.md` M1 instructs: *"Apply the same function in `aggregateFingerprint`
(`internal/mx/provenance.go`)"*. `mx.AggregateDescribedFingerprint` is not used only by the codemaps
check. `internal/graph/meta.go:57 dirFingerprint` reuses it to fingerprint **three non-source
directories**:

```
internal/graph/meta.go:38  dirFingerprint(<root>/.moai/project/codemaps)
internal/graph/meta.go:45  dirFingerprint(<root>/.moai/specs)
internal/graph/meta.go:48  dirFingerprint(<root>/.moai/reports)
```

None contains a `.go` file. A predicate applied *inside* `aggregateFingerprint` filters all three to
an empty entry list, so all three collapse to the same constant hash regardless of content.
`compareSourceFingerprints` / `EdgesSourcesMoved` then never detect source movement, and the edges
layer reports fresh permanently.

This crosses a scope boundary the SPEC declares explicitly (`spec.md` §E, *Out of Scope — other
gated layers*: "The `mx-index` and `edges` layers … neither participates in the failure mode"), and
it produces exactly the outcome §D.3 rejects on principle — *"a permanently green gate that reports
nothing"*.

Fix: do not apply the predicate inside `aggregateFingerprint`. Add a predicate-bearing variant
(e.g. `AggregateDescribedFingerprintFiltered(projectRoot, roots, pred)`) and apply it at the
codemaps call site `internal/graph/check.go:181`, leaving `AggregateDescribedFingerprint`'s existing
contract intact for `meta.go`. Add an AC pinning `SourceFingerprintsForEdges` output byte-identical
before and after the change.

---

**D2. The threshold is derived on the per-integration axis; the axis that decides how often the gate
is red is never measured — `spec.md` §D.2 / §B.5 — Severity: major — Class: blocking**

D.2 reasons entirely per-integration (p90, mean) about a metric that is **cumulative since the
stamp**. The cumulative-crossing cadence — how many integrations elapse before a red, at the
observed integration rate — is not measured anywhere in the SPEC. Measured here:

```
git log --first-parent -30 --reverse --name-only --pretty=format:"===%h" -- internal cmd pkg
  | <predicate> | distinct-union walk
```

| integrations since a hypothetical restamp | distinct described-worthy union |
|---|---|
| 3 | 11 |
| **5** | **17**  ← crosses 15 |
| 8 | 21 |
| **10** | **49**  ← crosses 40 |
| 30 | 86 |

The window `39c677f47 … 48eb945df` spans **2026-08-25 → 2026-08-27** — 30 integrations in three
days, ≈10/day. So the adopted value 15 puts the gate into red roughly **twice per day** of factory
activity, against roughly once per day at the retained 40, and it stays red for every later
integration until someone regenerates six documents by hand — a regeneration §D.3 deliberately
refuses to automate and which two open cards (t311, t304) have not yet been able to schedule.

Two consequences the SPEC does not state:

1. **The threshold change is not load-bearing for the defect the SPEC exists to fix.** The streak's
   corrected cumulative is `2` — under 15 *and* under 40. `git diff --name-only 9326b5478d… 48eb945df … | <predicate> | grep -c .` → `2`. D.1 alone stops the reported streak; D.2 is a
   separate, sensitivity-*increasing* change bundled into the same SPEC.
2. **The "40 has become ~13× laxer" premise holds only on the commit axis.** On the integration
   axis — the one CI runs on — corrected-40 already reds within ~10 integrations ≈ one day. 40 is
   not lax there.

Fix: add the cumulative-crossing measurement above to §B.5 and make §D.2's derivation answer it
explicitly — state the intended red frequency at the observed integration rate and justify 15
against it, or retain 40 and record why. REQ-GFC-005's run-phase re-derivation must require this
axis, not only the p90/mean pair; AC-GFC-006 currently mandates only the latter.

*On the deferral itself*: REQ-GFC-005 + AC-GFC-006 + AC-GFC-007 (traceability of the value to the
derivation, and the explicit "no value justified by reference to a failing check" clause) are honest
engineering, not a way to ship an unjustified number — the SPEC does not adopt 15 on its own
authority. The defect is the missing axis, not the deferral.

---

**D3. AC-GFC-003 cannot fail on any implementation defect — `acceptance.md` §D.1 — Severity: major
— Class: blocking**

AC-GFC-003 is mapped to REQ-GFC-001 (the metric shall count only described-worthy files), but its
deciding command is a hand-written shell filter chain applied to git output, compared against the
same command unfiltered. It re-derives the baseline; it never exercises the delivered predicate or
`gitDiffNameCount`. Both figures reproduce today, at HEAD, with no implementation present — I ran
them: `65` and `2`. A shipped implementation that filtered nothing would leave this AC green.

Fix: decide AC-GFC-003 through the built artifact — run the corrected `gitDiffNameCount` (or
`./bin/moai graph check --json`) against a fixture stamped at `9326b5478d…` and assert `2`, with the
unfiltered `65` retained only as the baseline-attribution row.

---

**D4. AC-GFC-005 enumerates four absent branches; `checkCodemaps` has at least six —
`acceptance.md` §D.1 / `spec.md` REQ-GFC-011 — Severity: major — Class: blocking**

REQ-GFC-011 names five preserved conditions, including *not-comparable stamped commit*.
AC-GFC-005 enumerates only four and omits it. Read at `internal/graph/check.go:150-207`, the actual
absent branches are:

| branch | line |
|---|---|
| codemaps directory missing | (pre-`:150`) |
| provenance missing | (pre-`:150`) |
| provenance unparseable | `:151` |
| described root invalid | `:172` |
| **clean stamp carries no commit sha** | `:197` — not in the AC |
| **stamped commit not comparable** (returns report **and** error) | `:205` — not in the AC, named by the REQ |
| dirty path: described roots unreadable | `:181` — not in the AC, and this is the branch REQ-GFC-003 modifies |

Two of the uncovered branches sit exactly where the change lands (the dirty fingerprint path), and
the not-comparable branch is the one t291/t292's orphan hazard produces — the highest-value case.

Fix: extend AC-GFC-005 to all six/seven branches, and pin the not-comparable branch's
`(report, error)` pair shape, not just the verdict.

---

**D5. Reported p90 = 11 is not reproducible under any standard percentile convention — `spec.md`
§B.5 — Severity: minor — Class: optional**

Measured on the same 30-value distribution: nearest-rank p90 (`ceil(0.9·30)` = rank 27) = **9**;
linear-interpolated p90 = **9.2**. The value 11 sits at rank 28 ≈ the 93rd percentile.

The error direction is conservative — with the true p90 at 9, "15 clears the p90 by a margin" holds
*a fortiori* — so no conclusion changes. It is nonetheless a mis-stated measured figure carried into
a derivation. Fix: state the convention and correct to 9 (or keep 11 and label it as the top-decile
entry value), and require AC-GFC-006's run-phase record to name the convention it uses.

---

**D6. D.1's consumer phrasing is imprecise (the substance is stronger than stated) — `spec.md`
§D.1 — Severity: minor — Class: optional**

§D.1 says narrowing the roots *"would also silently change `mx.AggregateDescribedFingerprint` and
`internal/graph/symbol`, which consume the same list"*. `AggregateDescribedFingerprint` does not
consume `DefaultDescribedRoots`; it receives `roots` as a parameter. The actual direct consumers,
measured:

```
internal/graph/check.go:168          roots = mx.DefaultDescribedRoots
internal/graph/symbol/symbol.go:99   for _, root := range mx.DefaultDescribedRoots
internal/mx/provenance.go:237,253,264  baseProvenance(..., DefaultDescribedRoots)  ← codemaps-gen, mx-scan, graph-build
```

The placement argument survives intact and is in fact stronger — three stamp writers, not one,
would change. Fix: correct the citation to the five call sites above.

---

**D7. AC-GFC-007's deciding command does not decide its claim — `acceptance.md` §D.1 — Severity:
minor — Class: optional**

The criterion asserts a negative over an unbounded space ("no commit message, code comment, or
evidence line justifies the value by reference to a check that was failing"), decided by
`grep -n "CodemapsChangedFiles" internal/graph/check.go` plus "inspection". The grep locates the
constant; it cannot decide the negative. Fix: bound the search space explicitly — e.g. the branch's
own commit messages (`git log <base>..HEAD --format=%B`) and the diff of `check.go` plus
`progress.md` §E.2 — so the criterion is decidable rather than a judgment call.

---

**D8. The `internal/template/templates` exclusion prefix — and the entire configuration surface
built for it — excludes zero measured files — `spec.md` REQ-GFC-002/004, `plan.md` M4 — Severity:
major — Class: optional**

`git ls-files internal/template/templates | grep -c '\.go$'` → **0**. The payload tree contains no
tracked Go file, so the `.go` rule already excludes all of it. Measured both ways:

- streak counterfactual with the prefix rule: `2`; **without** it: `2`.
- across the 30-integration window, files removed by the prefix rule beyond what `.go`/`_test.go`/
  `testdata` already remove: `0`.

The 7 template rule YAMLs in t228's 55 are `.yml` and fall to the `.go` rule. So REQ-GFC-002's
exclusion clause, REQ-GFC-004's configurable override, milestone M4, AC-GFC-011 and AC-GFC-012 are
all justified by no measured file — a configuration surface with no demonstrated need
(`moai-constitution.md` § Enforce Simplicity: "does this need to be built at all?").

I classify this **optional** rather than blocking because §G Q2 records the residual honestly and
the configurability is what makes the operator's reversal cheap. But the SPEC's §B.2 presents the
payload exclusion as part of what produces 65 → 2, and it produces none of it — that is an internal
inconsistency worth correcting even if M4 is kept. Fix: either state that the prefix rule is
forward-looking and contributes 0 today (correcting §B.2's framing), or drop the clause and M4 and
let §G Q2 name it as a deferred configurability.

---

## Judgment-by-judgment adjudication (the items the dispatch named)

**1. D.2's re-derivation to 15** — the distribution is real and reproduces except for p90 (D5). The
window is **not** representative in the way the dispatch suspected it might not be: 18 of 30
integrations (60%) fall on a single day, 2026-08-27, and the maximum (29, `6786c3fa4`) is the
graph-freshness delivery itself — the SPEC's own subject, a self-referential outlier. It is
three days of peak factory activity, which is arguably the right thing to calibrate *against* (the
gate must survive peak cadence) but must be named as such rather than presented as a neutral
30-integration sample. Does 15 satisfy both intents? On the per-integration axis, yes (one of 30
integrations reds alone, at 29). On the cumulative axis, D2 above. Would it have fired on the
streak? **No** — neither 15 nor 40 fires on a corrected cumulative of 2. The deferral to run phase
is honest (REQ-GFC-005 + AC-GFC-006/007 make the number traceable and forbid failing-check
justification); the gap is the unmeasured axis, not the deferral.

**2. D.1's placement claim** — verified. The shared-consumer coupling exists and is broader than
stated (D6). A prefix-inclusion list genuinely cannot express "everything under `internal` except
any `testdata` segment": 397 tracked files sit under a `testdata` segment scattered across arbitrary
packages. The placement argument stands.

**3. D.1's predicate, both directions** — over-count direction: verified, the predicate is what
turns 65 into 2. Under-count direction: of the 78 distinct `internal|pkg|cmd` paths cited across the
six curated documents, **zero** are non-`.go` file paths and **zero** carry a `testdata` segment
(`grep -E '\.[a-z]+$' | grep -v '\.go$'` → empty). So nothing the documents currently cite is
dropped. The residual — a package composed entirely of non-Go files would be invisible to the
predicate — is disclosed in §G Q2 and is a limitation, not a defect. The one measurable
over-reach is the template-prefix clause (D8).

**4. D.3's refusal to automate** — defensible, not a declined assignment. The card itself supplies
the prohibition ("재스탬프는 레드를 지우지만 문서의 부정확은 남깁니다"), and answering "nowhere,
because attaching it automates the act you just told me is inadequate" answers the question with a
reasoned negative plus a substitute. The dispatch's sharper question — does the substitute address
inheritance or merely make it legible — has the honest answer **merely legible**, and the SPEC says
so itself ("inheritance stops being a defect and becomes the intended behaviour"). The red still
lands on innocent lanes. That is acceptable *only if* the red is rare; D2 shows it will not be, and
§G Q1 leaves the blocking policy open without measuring how often it bites. D.3 is sound; its
interaction with D.2 is what is unmeasured.

**5. REQ-GFC-003's consistency obligation** — real, not decorative. The two paths are separate
functions today: clean → `gitDiffNameCount` (`internal/graph/check.go:414`, unions diff with
`ls-files --others`, applies no filter of any kind); dirty → `mx.AggregateDescribedFingerprint` →
`aggregateFingerprint` (`internal/mx/provenance.go:107`, walks and hashes every regular file, no
predicate). The divergence the REQ describes — a `testdata` edit flipping a dirty-stamped tree stale
while a clean-stamped tree is untouched — follows directly. But see D1: the *implementation route*
the plan prescribes for closing it is unsafe.

**6. AC mechanical verifiability** — per-AC failure input:

| AC | What input makes it fail | Verdict |
|---|---|---|
| 001 | any predicate branch misclassifying one of the five named paths | failable |
| 002 | predicate wired to the diff branch only → returns 2 (mutant named in the AC) | failable, strong |
| 003 | **nothing** — decides on a shell filter, never touches the code | **vacuous (D3)** |
| 004 | fingerprint moves on a `testdata` edit, or stays on a production `.go` edit | failable (but blind to D1) |
| 005 | any of the four enumerated branches returning non-`absent` | failable, under-scoped (D4) |
| 006 | `progress.md` §E.2 missing command / output / derivation | failable (self-attested) |
| 007 | grep locates the constant but cannot decide the negative | **weak (D7)** |
| 008 | missing first parent defaulted to 0 (mutant named in the AC) | failable, strong |
| 009 | stderr carries the count only, or truncates without an overflow marker | failable |
| 010 | a pre-existing JSON field renamed, or the new fields absent | failable |
| 011 | supplied list not replacing the default, or exit-2 contract broken | failable |
| 012 | template mirror missing, or a SPEC-ID / date / SHA / host path leaking | failable |
| 013 | a `graph stamp` invocation introduced, or `provenance.json` moved | failable (over-broad: fails on any legitimate future mention) |

Eleven of thirteen are genuinely failable, two of them with named mutants — above the usual
standard. Two are defective (003, 007), and one is under-scoped (005). The `§D.2 Edge Cases`
block is a strength: the `testdatax` substring case and the trailing-separator prefix case are
exactly the two ways this predicate is normally got wrong.

**7. Scope boundary and ordering basis** — §E and §F hold. Write-surface disjointness is real
(t322 writes `internal/graph`, `internal/mx`, `internal/config`, `gate.yaml` + mirror; t311/t304
write `.moai/project/codemaps/*.md`; intersection empty). §F supplies the evidence and explicitly
declines to make the ordering decision, which is what the card asked. It correctly identifies the
one real collision (t304's scope item 3, the citation-existence axis, lands in `check.go`) and
correctly leaves the t311 ↔ t304 conflict untouched — the queue records that conflict as an
operator call, and the SPEC does not usurp it. The §F caution — that "graph check reports fresh" is
restamp-satisfiable under both metrics, so t311/t304 need a content-level AC of their own — is a
genuine contribution beyond what was asked. **One residual**: §F establishes that t322 *benefits*
from landing first on signal grounds but sets no order, and §G Q3 leaves the t304 fold-in undecided;
t322 therefore ships with an unresolved dependency on a decision nobody has made. That is the
correct division of labour (evidence supplied, decision withheld) but the lead must actually make
the call before t304 starts, or the collision lands unmanaged. **However — D1 above changes this
picture: the SPEC's write surface silently extends into `internal/graph/meta.go` and the edges
layer, which §E declares out of scope. §F's disjointness measurement is correct for the surfaces
the SPEC names; D1 is a surface it does not name.**

**8. Rejected alternatives** — five rejections, each carrying a concrete reason rather than a
preference: narrow `described_roots` (expressiveness measurement + shared consumers — verified,
D6); citation-existence axis (orthogonal, would not have prevented any of the three failures —
correct, it checks citation validity not count inflation); unconditional post-merge restamp (erases
the signal); conditional auto-restamp on codemaps touch (no behavioural change, would stamp
abandoned regenerations); purely per-change metric (eliminates the drift signal — the strongest of
the five, and correctly reasoned: N integrations of 3 each would all pass while the documents rot).
None of the five is the better option on the evidence presented. The one alternative **not**
considered is *retain 40 under the corrected metric* — which D2 shows is a live candidate, since the
corrected metric already reds within ~10 integrations. That omission is folded into D2.

---

## Recommendation

The evidence base is unusually strong — every derived figure but one reproduces exactly, and the
judgments are argued rather than asserted. The SPEC does not need re-thinking. It needs four
corrections before Implementation Kickoff Approval:

1. **D1 (critical)** — rewrite `plan.md` M1 to apply the predicate at the `check.go:181` call site
   via a predicate-bearing variant, leaving `mx.AggregateDescribedFingerprint`'s contract intact for
   `internal/graph/meta.go`. Add an AC pinning `SourceFingerprintsForEdges` unchanged. Without this,
   a faithful run phase disables the edges layer silently.
2. **D2 (major)** — add the cumulative-crossing measurement to §B.5, and make §D.2 and REQ-GFC-005 /
   AC-GFC-006 answer the integration-cadence axis. Reconsider 15 against it, with "retain 40" as an
   explicit candidate.
3. **D3 (major)** — re-point AC-GFC-003 at the built artifact.
4. **D4 (major)** — extend AC-GFC-005 to every absent branch in `checkCodemaps`, including
   not-comparable and the dirty-path unreadable-roots case.

D5–D8 are surfaced for the orchestrator's discretion (M6 finding-consumption): D5, D6 and D7 are
one-line corrections; D8 is a scope decision about whether M4 earns its place.

The three open questions in §G are correctly left open, and MP-7 is satisfied because they are
recorded as operator decisions rather than as unresolved clarification markers. §G Q1 in particular
becomes materially more urgent if D2 resolves toward 15.

---

## Gaps (explicitly NOT observed)

- **CI verdicts.** I did not query GitHub; the failure/failure/failure/success column in §A is
  consumed from the orchestrator's stated reproduction, not independently observed here.
- **Runtime behaviour.** No `moai graph check` was executed and no Go test was run (no
  implementation exists, and the dispatch forbids the full suite and `e2e/`). D1's consequence is
  read from the source at `internal/graph/meta.go:38-57` and `internal/mx/provenance.go:107-143`;
  it is a source-level determination, not an executed one.
- **The `-10`/`-50` calibration figures (21 / 198)** in §B.3 were reported reproduced by the
  orchestrator and I did not re-run them.
- **80 vs 78 cited paths.** My extraction regex differs from the SPEC's unstated one; I did not
  reconcile the two-path difference, and it changes no conclusion.
- **t311 / t304 artifacts.** I read their queue-card text, not any SPEC they may carry.

## Residual risk

- The 30-integration window is three days of peak factory cadence. If integration rate falls
  substantially, D2's cadence argument weakens and 15 becomes more defensible than measured here.
- D1's severity assumes the predicate is applied unconditionally inside `aggregateFingerprint`. An
  implementer who independently notices `meta.go` and routes around it produces a correct result and
  the defect never materializes — but nothing in the artifacts would have prompted that.
- The predicate's under-count direction is measured against what the documents **currently** cite.
  A future regeneration that legitimately describes non-Go surface inside the described roots would
  be invisible to the metric, and nothing in the SPEC would detect the divergence.
