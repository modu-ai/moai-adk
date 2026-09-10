# SPEC Review Report: SPEC-GRAPH-FRESHNESS-CADENCE-001 (card t322)

Iteration: 2/2 (Tier M ceiling — this is the final iteration)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.895** (Tier M PASS threshold 0.80)
Delta vs iter-1: **0.82 → 0.895 = +0.075 — monotonic improvement.** No STOP escalation.

Audited in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch
`WT-graph-freshness-cadence`, HEAD `ee74e9474`, tracked changes 0
(`git rev-parse --show-toplevel`, `git rev-parse HEAD`, `git branch --show-current`,
`git status --short` — the only entry is this report series, untracked).
Reasoning context ignored per M1 Context Isolation; the four SPEC artifacts plus the Go source they
cite are the input surface.

**The four blocking defects are genuinely closed.** Each was verified against source, not against
the HISTORY row's claim. Three new findings are blocking, all small and all measured; five are
optional. None re-opens D1-D4.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — the numeric line REQ-GFC-001..012 is complete: no gap
  (004 and 012 are retained in place as withdrawal records rather than renumbered — the gap is
  explained where it occurs), no duplicate, consistent 3-digit padding.
  Evidence: `grep -n '^\*\*REQ-GFC' spec.md` → 12 entries at spec.md:192-240.
  The inserted `REQ-GFC-003a` is an out-of-line suffix, neither a gap nor a duplicate; noted as
  optional finding O1.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-GFC-*` in
  `spec.md` §C), never against `acceptance.md`. All eleven live requirements match a GEARS pattern:
  Ubiquitous (001, 002, 003, 005, 010, 011), Unwanted `shall not` / negative (003a, 006, 009),
  Event-driven `When …` (007, 008). The Given-When-Then entries in `acceptance.md` are the
  verification layer and are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`version: "0.2.0"` quoted, `status: draft`, `created`/`updated` ISO `2026-08-27`,
  `priority: P1`, `lifecycle: spec-anchored`, `tags` comma-string). No rejected snake_case alias.
  Extras (`era`, `tier`, `related_specs`) are additive.
- **[PASS/N-A] MP-4 language neutrality** — N/A: single-language SPEC (Go tooling internal to this
  repository). Strengthened since iter-1: the template-mirror obligation was withdrawn with
  REQ-GFC-012, and `git ls-files internal/template/templates | grep -c '\.go$'` → `0` confirms the
  payload tree carries no Go surface to be neutral about.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — three referenced SPECs, all present, all
  `status: completed`; none retired/superseded/archived.
  Evidence: `grep -hEo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md | sort -u` →
  4 ids (3 external + self); `grep -m1 '^status:' …/spec.md` on each → `completed` ×3.
- **[PASS/N-A] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md plan.md
  acceptance.md progress.md` → `0` on all four. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'
  .moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/` → rc=1, no matches. (`research.md` absent — correct
  for the Tier M artifact set.) The four §G entries are recorded operator decisions, not
  clarification markers.

---

## Category Scores

| Dimension | iter-1 | iter-2 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.90 | **0.90** | 0.75-1.0 | D5 and D6 both closed and independently re-verified. Offset by N1 (§D.2's red-frequency figure) and O3 (§E exclusion vs REQ-GFC-005). |
| Completeness | 0.75 | **0.88** | 0.75-1.0 | §B.6 is a substantive new section and every figure in it reproduces exactly. Docked for N3 (stale counts in four places), O2 (`treeDirty` unnamed), O4 (the unstated premise behind the unpaired producers). |
| Testability | 0.70 | **0.88** | 0.75-1.0 | D3, D4, D7 all closed and verified against source; AC-GFC-014 kills the D1 mutant on a premise I re-measured. Docked for O5 (AC-014's undecidable half) and the surviving AC-006 self-attestation / AC-013 over-breadth. |
| Traceability | 0.92 | **0.92** | 0.75-1.0 | Coverage total in both directions after the withdrawals (checked below). Docked for N3 and O1. |

Aggregate: arithmetic mean, matching iter-1's aggregation (its 0.90/0.75/0.70/0.92 → 0.8175 ≈ 0.82).
(0.90 + 0.88 + 0.88 + 0.92) / 4 = **0.895**.

---

## Priority focus — REQ-GFC-003's producer/consumer pairing, audited as new work

D1's remedy moved the predicate out of a shared function and into call sites, which creates a
pairing obligation. I enumerated every producer and every consumer from source rather than from the
SPEC's account.

Commands: `grep -rn --include='*.go' 'aggregateFingerprint\|AggregateDescribedFingerprint' . |
grep -v _test.go`; the same for `ContentFingerprint`, `baseProvenance`, and
`SourceFingerprintsForEdges\|dirFingerprint\|compareSourceFingerprints\|EdgesSourcesMoved`.

### The described-content fingerprint (`Provenance.ContentFingerprint`)

**Producers — three, all through one function.** `baseProvenance` (`internal/mx/provenance.go:208`)
computes it at `:219` on the dirty path, and all three stamp writers reach it:

| # | Writer | Line | Artifact stamped |
|---|---|---|---|
| P1 | `StampCodemaps` | `provenance.go:237` | `.moai/project/codemaps/provenance.json` |
| P2 | `StampMXScan` | `provenance.go:253` | the mx sidecar |
| P3 | `StampEdges` | `provenance.go:264` | the edges meta sidecar |

**Consumers — exactly one comparator, plus one display reader.**

| # | Reader | Line | Kind | SPEC coverage |
|---|---|---|---|---|
| C1 | `checkCodemaps` — recompute `check.go:181`, compare `check.go:187` | the sole comparator | **covered**: REQ-GFC-003, plan M1, AC-GFC-004 |
| C2 | `Provenance.Describe()` | `provenance.go:280` | display only (`shortHash`) | **not named** — see O4 |

There are no others. `grep -rn --include='*.go' 'ContentFingerprint' . | grep -v _test.go` returns
only the field declaration, the write at `:220`, the display at `:280`, and `check.go:187`/`:194`.

**Verdict on the pairing.** P1 ↔ C1 is the codemaps pair, and the SPEC covers it correctly:
plan.md M1 requires the filtered variant at `check.go:181` **and** in the codemaps stamp path, and
AC-GFC-004 names the exact mutant (filtered checker + unfiltered writer) with an assertion that
fails immediately on it. That is the right pair and the right control.

P2 and P3 remain unpaired producers — and they are **safe**, because no consumer compares them.
I measured that; the SPEC does not. §D.1 and plan §B state the *decision* ("leaving the mx-scan and
graph-build stamps unchanged") without stating the *premise that makes it safe*, and no criterion
pins it. That is O4 below: a hole in the pairing's stated coverage, not a correctness defect today.

### The edges source-set fingerprints (`Provenance.SourceFingerprints`) — a separate family

`SourceFingerprintsForEdges` (`meta.go:37`) → `dirFingerprint` (`meta.go:57`) →
`mx.AggregateDescribedFingerprint` (`meta.go:67`), consumed by `compareSourceFingerprints`
(`meta.go:108`) via `checkEdges` (`check.go:361-362`), `EdgesSourcesMovedFor` (`meta.go:147`), and
the CLI refresh trigger (`graph_refresh_cli.go:44`). It shares the *function* with the codemaps
path but not the *field*. **Covered** by REQ-GFC-003a and AC-GFC-014.

### The mx-index inventory — not a hole

`checkMXIndex` compares `FileInventory` per-file hashes through `MXIndexDrift` (`check.go:245`),
never `ContentFingerprint`. Unaffected by any filtering decision here. Correctly untouched.

### One consumer the SPEC does not name, and it is on the codemaps stamp path

`treeDirty` (`provenance.go:201`) is scoped to the described roots and is **predicate-blind**. It
decides which anchor `StampCodemaps` uses. Consequences, read from source:

- A tree whose only uncommitted change is a `testdata` fixture is still `Dirty=true`, so the stamp
  is fingerprint-anchored. Judging stays consistent (both sides compare filtered fingerprints), so
  there is no verdict defect.
- But `StampCodemaps` **rejects** `--commit` on a dirty tree (`provenance.go:245-247`). So
  `moai graph stamp codemaps --commit "$(git merge-base …)"` — the merge-surviving anchor
  `SPEC-STAMP-REACHABILITY-001` delivered and which §D.3 cites as the orphan mitigation — is still
  refused on a tree dirty only in `testdata`, even though this SPEC's whole thesis is that such a
  tree carries no described-worthy change. The predicate does not reach the anchor-selection gate.

Leaving `treeDirty` unfiltered is the conservative and probably correct choice, but it is exactly
the "same axis, different function" divergence REQ-GFC-003 exists to prevent, and the SPEC names it
nowhere. See O2.

---

## Verification of the four closures

**D1 — closed, and it greens nothing.**
`internal/graph/meta.go:67` still calls the **unfiltered** `mx.AggregateDescribedFingerprint`, and
plan.md M1 explicitly instructs leaving `AggregateDescribedFingerprint` / `aggregateFingerprint`
unfiltered while adding a predicate-bearing variant used only at `check.go:181` and the codemaps
stamp path. The three `dirFingerprint` directories therefore keep an unfiltered route.
Evidence: `grep -rn --include='*.go' 'aggregateFingerprint\|AggregateDescribedFingerprint' .` →
`check.go:181`, `meta.go:67`, `provenance.go:103/104/107/219`.

**AC-GFC-014's failing input, stated:** a build in which the predicate is pushed down into
`aggregateFingerprint`. Under it, `.moai/project/codemaps` and `.moai/specs` — measured
`find … -name '*.go' | wc -l` → **0** for both — contribute no entries, and `hashEntries(nil)` folds
to the SHA-256 of the empty input. Verified: `printf '' | shasum -a 256` →
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`, byte-identical to the constant
the AC asserts against. The AC's inequality then fires. The criterion is genuinely failable on the
named mutant, and its stated premise is true in this tree.

**D2 — reversal complete and coherent, with one measured residual (N1).**
No prescriptive text adopts 15. `grep -n '\b15\b' *.md` returns 11 hits and every one is
measurement, comparison, or explicit rejection (`spec.md:172,178,185,300,311,318`; `plan.md:97`;
`progress.md:14,28,29,31`). §D.2 now carries the integration-axis justification and names the
intended red frequency. §E's new "threshold sensitivity change" exclusion orphans nothing:
REQ-GFC-005 and AC-GFC-006 survive and were both *extended* to require Axis 2 — but see O3 for a
wording tension between that exclusion and REQ-GFC-005's correction clause.

I re-measured the entire §B.5 / §B.6 evidence base at HEAD `ee74e9474`. Everything reproduces:

| Claim | Observed |
|---|---|
| streak cumulative 65 | `65` |
| corrected cumulative 2 | `2` |
| 12 of 30 integrations contribute 0 | `12` |
| median 2 | `2` |
| **p90 nearest-rank = 9** (the D5 correction) | `9` — rank 27 of the sorted 30; rank 28 is `11`, confirming D5's diagnosis exactly |
| max 29 | `29` (`6786c3fa4`) |
| mean 3.9 | `3.9` (sum 117 / 30) |
| §B.6 crossings 3→11, 5→17, 8→21, 10→49, 30→86 | all five exact |
| window `39c677f47 … 48eb945df`, 2026-08-25 → 2026-08-27 | exact |

Commands: `git diff --name-only 9326b5478d… 48eb945df -- internal cmd pkg | grep -c .` and the same
with the predicate chain; `git log --first-parent -30 --reverse --name-only
--pretty=format:"===%h" -- internal cmd pkg` piped through an awk predicate filter for the
per-integration counts and the distinct-union walk.

**D3 — closed.** AC-GFC-003 now decides through the built artifact:
`./bin/moai graph check --json | jq '.layers[] | select(.layer=="codemaps") | .value'` → `2`
against a fixture stamped at `9326b5478d…`. It exercises the delivered `gitDiffNameCount` rather
than re-deriving the baseline, and `65` is demoted to a baseline-attribution row. Both fixture
commits are reachable: `git cat-file -t` → `commit` for both, and `git merge-base --is-ancestor`
→ rc 0 for both against HEAD. Its failing input: an implementation that filters nothing, which
reports `65`.

**D4 — closed, and the enumeration matches source exactly.** AC-GFC-005 now enumerates seven
branches and pins branch 7's `(report, error)` pair shape.
Evidence: `grep -n 'VerdictAbsent' internal/graph/check.go` → 10 hits, of which exactly seven fall
inside `checkCodemaps` (lines 139, 146, 152, 172, 183, 199, 208) and the remaining three are the
constant (`:20`), `checkMXIndex` (`:229`), and `checkEdges` (`:350`, `:356`). The AC's seven rows
map one-to-one onto those seven lines, in source order, with reason strings that match verbatim.
Branch 7 (`check.go:205-208`) does return a non-nil error alongside the report, as the AC asserts.

---

## Withdrawal cleanliness (D8)

`internal/template/templates` holds **0** tracked `.go` files
(`git ls-files internal/template/templates | grep -c '\.go$'`), so the D8 premise is measured
and true.

The withdrawal is clean. `grep -n 'REQ-GFC-004\|REQ-GFC-012\|\bM4\b\|AC-GFC-011\|AC-GFC-012' *.md`
returns hits only in withdrawal records and in text that describes the withdrawal — **no live
reference** to any withdrawn item anywhere in the four artifacts. Requirement numbering carries no
unexplained gap: 004 and 012 are retained in place with the reason stated at the vacated slot.

Coverage after the withdrawals is total in both directions:

| Live REQ | Covering AC | Live AC | Cited REQ | Milestone |
|---|---|---|---|---|
| 001 | 002, 003 | 001 | 002 | M1 |
| 002 | 001 | 002 | 001 | M1 |
| 003 | 004 | 003 | 001 | M1 |
| 003a | 014 | 004 | 003 | M1 |
| 005 | 006 | 005 | 011 | M1 |
| 006 | 007 | 006 | 005 | M2 |
| 007 | 008 | 007 | 006 | M2 |
| 008 | 009 | 008 | 007 | M3 |
| 009 | 013 | 009 | 008 | M3 |
| 010 | 010 | 010 | 010 | M3 |
| 011 | 005 | 013 | 009 | all |
| — | — | 014 | 003a | M1 |

Eleven live requirements, twelve live criteria. No requirement without a covering criterion, no
criterion citing a withdrawn or nonexistent requirement, no requirement without a milestone, no
surviving milestone without criteria. Both counts sit under the Tier M ceilings (16/16).

---

## Per-criterion failability

Every live criterion, with the input that makes it fail:

| AC | Input that makes it FAIL | Verdict |
|---|---|---|
| 001 | any predicate branch misclassifying one of the five named paths | failable |
| 002 | predicate wired to the diff branch only — the fixture's untracked fixture then makes `ls-files --others` return `2`, not `1` (mutant named in the AC; `gitDiffNameCount` has exactly the two branches, `check.go:417` and `:428`) | failable, strong |
| 003 | an implementation that filters nothing → the built binary reports `65`, not `2` | **failable — D3 closed** |
| 004 | a filtered checker paired with an unfiltered codemaps stamp writer → the tree reads stale against its own fresh stamp on the first assertion | failable, strong |
| 005 | any of the **seven** enumerated branches returning non-`absent`, or branch 7 returning a nil error | **failable — D4 closed** |
| 006 | `progress.md` §E.2 missing either axis, the convention, or a command's verbatim output | failable (self-attested) |
| 007 | a commit message in `d2cba5e21..HEAD`, or a comment in the `check.go` diff, justifying the value by a failing check | **failable — D7 closed**; the base is an ancestor of HEAD (`git merge-base --is-ancestor` rc 0) |
| 008 | a missing first parent defaulted to 0 (mutant named in the AC) | failable, strong |
| 009 | stderr carrying the count alone, or truncating without an overflow marker | failable |
| 010 | a pre-existing JSON field renamed, or the new fields absent | failable |
| 013 | a `graph stamp` invocation introduced into `.github/` or `.claude/`, or `provenance.json` moved. Baseline confirmed: `grep -rn "graph stamp" .github/ .claude/` → 0 hits today | failable (over-broad — fires on any future legitimate mention) |
| 014 | the predicate pushed into `aggregateFingerprint` → codemaps and specs collapse to `e3b0c442…` and the inequality fires | **failable — premise re-measured**; but see O5 |

Ten of twelve are strongly failable, four with explicitly named mutants. The three iter-1 defects in
this table (003 vacuous, 005 under-scoped, 007 undecidable) are all repaired. The fourth vacuity
this card's lineage might have produced — AC-GFC-014 — does not materialize: its operative assertion
is the empty-hash inequality, which I verified fires on the named mutant.

---

## Defects Found

**N1. §D.2's red-frequency figure is inflated by the one integration §B.5 itself disowns —
`spec.md` §B.6 / §D.2 / §G Q4 — Severity: major — Class: blocking**

§D.2's load-bearing claim is *"corrected-40 crosses at 10 integrations, ≈one day at the observed
≈10 integrations/day. 40 is not lax on the axis that decides how often the gate is red."* The
crossing at 10 is produced by a single integration: `6786c3fa4`, contributing 29 — which §B.5 in
the same document flags as *"the SPEC-V3R6-GRAPH-FRESHNESS-001 delivery itself"*, a self-referential
outlier. The union curve is `…8:21, 9:21, 10:49…`: the jump from 21 to 49 is that one commit.

Measured counterfactual, excluding that integration entirely from the window (awk union walk with
the `===6786c3fa4` block skipped):

| Threshold | Crossing with the outlier | Crossing without it |
|---|---|---|
| 15 | integration 5 | integration 5 |
| **40** | **integration 10** | **integration 16** |

So the ex-outlier red cadence for retained-40 is roughly **one red per 1.6 days**, not one per day.
Two consequences the SPEC does not state:

1. §D.2 commits to a figure ("expected to red roughly **once per day** of factory activity") that
   overstates the gate's redness by ~60% on the same window, in the direction that *weakens* its own
   argument — the closer 40 comes to being lax, the more of v0.1.0's position survives. The reversal
   itself is **not** at risk: its stronger and independent ground (§B.6 consequence 1 — the streak's
   corrected cumulative is 2, so the threshold is not load-bearing, and moving two variables at once
   makes the outcome unattributable) holds regardless. Only the stated number is wrong.
2. §G Q4 asks the operator *"Is roughly one red per day of factory activity tolerable?"* — an
   operator decision posed on the inflated figure.

This is the same defect shape as iter-1's D5 (p90 = 11), which the remediation correctly fixed; the
outlier-sensitivity of the *new* axis was not given the same treatment.

Fix: state both crossings in §B.6 (with and without the self-referential integration), correct
§D.2's and §G Q4's frequency to the ex-outlier figure or state the range, and extend AC-GFC-006's
Axis 2 requirement to include the outlier-sensitivity note so the run-phase re-measurement does not
reproduce the same omission.

---

**N2. `baseProvenance` is cited at `internal/mx/provenance.go:196`; it is at `:208` —
`spec.md` §D.1 + HISTORY, `plan.md` §B — Severity: minor — Class: blocking**

`sed -n '190,222p' internal/mx/provenance.go` puts line 196 at `return sha, nil` inside
`ResolveCommit`. `baseProvenance`'s doc comment starts at `:206`, the function at `:208`, and the
fingerprint call this SPEC cares about at `:219`.

The wrong citation appears three times — `spec.md:25` (HISTORY), `spec.md:283` (§D.1), and
`plan.md:27` (§B Known Issues Inherited) — and it is the citation for the *newly introduced*
producer/consumer axis, in the section a run-phase implementer navigates by. The D6 remediation
corrected one citation set (the five `DefaultDescribedRoots` sites, which I re-verified as exact:
`check.go:168`, `symbol/symbol.go:99`, `provenance.go:237,253,264`) while introducing a wrong one
for the new axis.

Fix: `:196` → `:208` in all three places (or `:219` if the intent is the fingerprint call site
specifically — the surrounding prose says "computes the dirty `ContentFingerprint`", so `:219` is
the better target).

---

**N3. Four counts are stale after the withdrawals and the §G Q4 addition — `acceptance.md` §D.3,
`progress.md` §E.1 / §E.1b — Severity: minor — Class: blocking**

| Location | States | Actual |
|---|---|---|
| `acceptance.md:174` (Definition of Done) | "All **thirteen** criteria decided" | twelve live (011, 012 withdrawn) |
| `acceptance.md:180` | "The **three** open questions in `spec.md` §G" | four (Q4 added at v0.2.0) |
| `progress.md:9-10` | "**12** requirements … **13** acceptance criteria" | 11 live / 12 live |
| `progress.md:35` | "Open questions … `spec.md` §G (**three**)" | four |

Evidence: the §C requirement list (11 live), the §D matrix (12 live), and
`grep -n '^[0-9]\. \*\*' spec.md` over §G → 4 entries.

One of the four is a Definition-of-Done gate, so an implementer checking it off will look for a
thirteenth criterion that does not exist. This is the cross-layer sweep the revision missed: the
withdrawal was applied where the items live and not where they are counted.

Fix: four number edits. (Cosmetic sibling, fold in while there: `acceptance.md` §D.1 renders
AC-GFC-014 before AC-GFC-013, out of the order the §D matrix uses.)

---

**O1. `REQ-GFC-003a` is an out-of-line identifier while `REQ-GFC-004`'s slot sits vacated —
`spec.md` §C — Severity: minor — Class: optional**

The suffix form deviates from the 3-digit convention used by every other requirement in the SPEC and
by the repository's `REQ-[A-Z]{2,5}-\d{3}` shape. Appending `REQ-GFC-013` would preserve identifier
stability (the stated reason for not renumbering) without the deviation. Nothing breaks as written —
the SPEC's prose requirement format is not parsed by `internal/spec/lint.go`'s `reqLinePattern`
either way (see Gaps).

---

**O2. `treeDirty` is a predicate-blind described-roots consumer on the codemaps stamp path, and the
SPEC names it nowhere — `provenance.go:201` vs `spec.md` §D.1 / §E — Severity: major —
Class: optional**

Full analysis in the priority-focus section above. The practical edge: a tree dirty only in
`testdata` is still refused the `--commit` merge-base anchor that `SPEC-STAMP-REACHABILITY-001`
delivered and that §D.3 cites as the orphan mitigation — even though this SPEC's thesis is that such
a tree carries no described-worthy change.

Classified optional because nothing instructs the implementer to touch `treeDirty` and the default
(leave it) is the conservative, correct behaviour. But it is an unnamed consumer of the same axis
REQ-GFC-003 governs, and leaving it undiscussed means the next reader has to rediscover it.

Fix: one sentence in §D.1 or an §E bullet naming `treeDirty` as deliberately unfiltered, with the
`--commit` interaction stated as a known residual.

---

**O3. The §E threshold exclusion and REQ-GFC-005's correction clause say different things —
`spec.md` §E vs §C / §D.2 — Severity: major — Class: optional**

§E declares out of scope: *"Moving `CodemapsChangedFiles` away from 40."* REQ-GFC-005 requires the
threshold to be *"confirmed **or corrected**"* from implementation-time measurement; §D.2 and
`plan.md` §D both repeat the conditional (*"a correction is admissible if the measurement supports
one"*). An implementer whose M2 measurement contradicts 40 faces an absolute exclusion and a
conditional permission at once.

Optional because both readings converge on the same action in the expected case (hold 40), the
divergent case requires operator escalation regardless, and §E's own second sentence cites §D.2,
which carries the conditional. Still, an exclusion stated absolutely against a requirement stated
conditionally is the CN-2 shape.

Fix: qualify the §E bullet — "Moving `CodemapsChangedFiles` away from 40 **other than as the
recorded outcome of REQ-GFC-005's measurement**".

---

**O4. The unpaired producers are safe, but the SPEC states the decision without its premise —
`spec.md` §D.1, `plan.md` §B — Severity: minor — Class: optional**

`StampMXScan` and `StampEdges` keep writing an unfiltered `ContentFingerprint`. That is safe
**because no consumer compares those two layers' `ContentFingerprint`** — I measured it
(`grep -rn --include='*.go' 'ContentFingerprint' . | grep -v _test.go` yields exactly one
comparator, `check.go:187`, plus one display reader, `provenance.go:280`). The SPEC asserts the
decision and never the enabling fact, and no criterion pins it. A future layer that added a
dirty-fingerprint comparison for mx-index or edges would make the asymmetry live with nothing
guarding it.

Fix: state the measured premise in §D.1 ("codemaps is the only layer that compares
`ContentFingerprint`") and, optionally, extend AC-GFC-004 to assert it, so the premise is a control
rather than a footnote.

---

**O5. AC-GFC-014's byte-identity half is not decidable by the command that decides it —
`acceptance.md` §D.1 — Severity: minor — Class: optional**

The criterion asserts *"every one of its four source-set fingerprints is byte-identical across the
two [trees, before and after the change]"*, decided by
`go test ./internal/graph/ -run TestSourceFingerprintsForEdges_Unchanged`. A Go test running on one
tree cannot compute the pre-change tree's fingerprints — the pre-change code no longer exists at
test time. What actually carries the criterion is its second half, the inequality against
`e3b0c442…`, and the AC says so explicitly (*"which is why the assertion is stated as an inequality
against it"*). So this is an honest criterion with one inert clause, not a vacuous one.

Fix: either drop the byte-identity clause, or make it decidable by pinning golden fingerprint values
for a fixture project so "unchanged" becomes "equals these four constants".

---

## Regression check against iteration 1

| iter-1 defect | Severity | Status | Evidence |
|---|---|---|---|
| D1 — predicate inside `aggregateFingerprint` greens the edges layer | critical | **RESOLVED** | `meta.go:67` routes through the unfiltered `AggregateDescribedFingerprint`; plan M1 mandates a separate variant; REQ-GFC-003a + AC-GFC-014 pin it, and the AC's premise (0 `.go` files; empty-hash constant) re-measured true |
| D2 — cumulative-crossing axis unmeasured | major | **RESOLVED** (with N1 residual) | §B.6 added; all five crossings reproduce exactly; §D.2 reversed to retain 40; REQ-GFC-005 + AC-GFC-006 extended to both axes |
| D3 — AC-GFC-003 cannot fail | major | **RESOLVED** | now decided by the built binary against a reachable fixture; `65` demoted to baseline-attribution |
| D4 — AC-GFC-005 under-enumerates | major | **RESOLVED** | seven branches, matching the seven `VerdictAbsent` sites in `checkCodemaps` one-to-one in source order; branch 7's error pair pinned |
| D5 — p90 = 11 not reproducible | minor | **RESOLVED** | corrected to 9 with the nearest-rank convention named; re-measured: rank 27 = 9, rank 28 = 11 |
| D6 — consumer citation imprecise | minor | **RESOLVED** | five call sites, all five verified exact |
| D7 — AC-GFC-007 undecidable | minor | **RESOLVED** | bounded to three named commands over a stated finite space |
| D8 — payload-prefix surface excludes zero files | major | **RESOLVED by withdrawal** | premise re-measured (`0` tracked `.go`); clause, REQ-004, REQ-012, M4, AC-011, AC-012 all withdrawn with no dangling live reference |

No defect appears unchanged across both iterations. No stagnation.

---

## Recommendation

The four blocking defects are closed, and each closure was verified against source rather than
accepted from the HISTORY row. The remediation is unusually disciplined: it reversed a judgment
rather than defending it (D2), withdrew a requirement family rather than rationalising it (D8), and
the one genuinely new axis it introduced — the producer/consumer pairing — is covered where it
matters, with the right mutant named in AC-GFC-004. Every one of the fifteen measured figures in
§B.2, §B.5 and §B.6 reproduces exactly at HEAD `ee74e9474`, including the two corrections the
remediation made.

This is the Tier M ceiling, so there is no iteration 3. The three blocking findings are all small
and mechanical; my recommendation is to apply them as direct edits before Implementation Kickoff
Approval rather than to re-enter plan-phase:

1. **N1** — re-measure the §B.6 crossing with and without `6786c3fa4`, correct §D.2's and §G Q4's
   red-frequency figure, and extend AC-GFC-006 to require the outlier-sensitivity note at run time.
   This is the one finding with substance; the reversal to retain 40 survives it, but the number the
   operator is asked to judge should be the right one.
2. **N2** — `provenance.go:196` → `:208` (or `:219`) in three places.
3. **N3** — four stale counts.

O1-O5 are the orchestrator's discretion (M6 finding-consumption). O2 and O3 are each one sentence
and would improve the SPEC materially; O1, O4 and O5 are refinements.

§G's four open questions are correctly left open, and MP-7 holds because they are recorded as
operator decisions rather than clarification markers. §G Q4 becomes answerable only after N1.

---

## Gaps (explicitly NOT observed)

- **No Go test or build was executed.** No implementation exists, and the dispatch forbids
  `go test ./...` and anything under `e2e/`. Every source-level determination in this report —
  D1's closure, the producer/consumer enumeration, the seven absent branches, the `treeDirty`
  interaction — is read from source at HEAD `ee74e9474`, not executed.
- **CI verdicts.** The failure/failure/failure/success column in §A is consumed from the
  orchestrator's stated reproduction; I did not query GitHub.
- **AC-GFC-003's fixture was not constructed.** I verified both commits are reachable
  (`git cat-file -t`, `git merge-base --is-ancestor`) but did not build the fixture or run
  `./bin/moai graph check --json` against it. Neither `plan.md` M1 nor §D.3 describes constructing
  that fixture; the run phase will have to.
- **The `-10`/`-50` calibration figures (21 / 198)** in §B.3 were reported reproduced by iter-1's
  orchestrator; I did not re-run them in this iteration.
- **80 vs 78 cited paths** (§B.2) — unreconciled from iter-1; changes no conclusion.
- **SPEC lint was not run.** I note without measuring its consequences that this SPEC's prose
  requirement form (`**REQ-GFC-001 (Ubiquitous).**`) does not match `internal/spec/lint.go`'s
  `reqLinePattern` (`-\s+(REQ-[A-Z]{2,5}-\d{3}-\d{3})\s*:\s*(.+)`), so `parseREQs` will find zero
  requirements and the EARS/traceability lint rules are inert on it. This is a repository-wide
  convention divergence, not a t322 defect, and iter-1 did not raise it either — recorded here so
  it is not mistaken for a finding.
- **t311 / t304 artifacts.** Not read; §F's claims about them are taken from queue-card text.

## Residual risk

- **N1's counterfactual is itself a single-window measurement.** Excluding one integration from a
  30-integration sample is a large intervention on a small sample; the true ex-outlier cadence is
  bounded loosely, and REQ-GFC-005's run-phase re-measurement is the right place to settle it.
- **The 30-integration window remains three days of peak factory cadence.** If the integration rate
  falls, both crossings stretch and the retain-40 argument strengthens further.
- **D1's closure depends on the implementer following plan M1's two-mechanism instruction.** M1
  offers two ways to pair producer and consumer (a fingerprint-function parameter on
  `baseProvenance`, or a recompute inside `StampCodemaps`). AC-GFC-004 catches an unpaired result
  either way, so the choice is safe — but the first mechanism changes `baseProvenance`'s signature
  for all three writers, and nothing in the artifacts flags that as a review point.
- **The predicate's under-count direction is measured against what the codemaps documents currently
  cite.** A future regeneration legitimately describing non-Go surface inside the described roots
  would be invisible to the metric. Disclosed in §G Q2; unchanged from iter-1.
