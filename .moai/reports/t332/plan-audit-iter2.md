# SPEC Review Report: SPEC-BACKLOG-HYGIENE-001 (card t332) — iteration 2

Iteration: 2/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`) — **the last iteration available at Tier M**
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (harmonic mean; arithmetic 0.855) against the Tier M threshold **0.80**
Score movement: **0.74 → 0.85, +0.11 — up.** No score regression, so the STOP-escalation clause does not fire.

Reasoning context ignored per M1 Context Isolation. Artifacts re-read fresh from disk at v0.2.0 —
not recalled from the iter-1 read. Same tree: worktree `.claude/worktrees/t332`, HEAD `15453140a`.

Scope: per the Retry Loop Contract, iteration 2 is scoped to the enumerated iter-1 defect delta
plus a regression check over those defects, plus the must-pass re-verification the delta touches.
It is not a from-scratch full re-audit, and the Gaps section names what that scoping left unobserved.

---

## Headline

Every iter-1 defect is fixed except one, which is **partially** fixed: D5's digest exists as a
requirement and a criterion but carries no procedure, so the criterion cannot be executed as
written (N1 below). One new minor defect was introduced by the repair (N2, a false cell in the new
traceability map). The load-bearing claim — that the 23→16 consolidation lost no obligation — is
**verified true**, mapped requirement by requirement below.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o '^\*\*REQ-BH-[0-9]\{3\}\*\*' spec.md` →
  `REQ-BH-001 … REQ-BH-016`, count **16**, `uniq -d` empty. Sequential, zero-padded, no gap, no
  duplicate after the renumbering.

- **[PASS] MP-2 GEARS/EARS format compliance.** Judged against the **requirement layer**
  (`REQ-BH-*` in `spec.md` §D); `acceptance.md`'s Given-When-Then entries are the correct
  verification-layer format and are graded under Group 4. All 16 match a GEARS pattern: Ubiquitous
  (001, 002, 004, 007, 009, 013, 014, 015, 016), Unwanted `shall not` (005, 010), Event-driven
  `When …` (003, 006, 011), State-driven `While …` (008, 012). The iter-1 D7 deviation is gone —
  the two runtime-condition requirements now read `When an overlap … is confirmed` (006) and
  `While an in-scope card has a live worktree` (012), which is the correct modality for each.
  REQ-BH-009 now names `the sweep` as its acting subject (D8).

- **[PASS] MP-3 YAML frontmatter validity.** `spec.md:12` reads `lifecycle: spec-anchored`, a member
  of the schema enum. All 12 canonical fields present under canonical names; `version: "0.2.0"` and
  `updated: 2026-08-29` were bumped with the revision. The iter-1 must-pass failure is closed.

  Carried forward as a standing caveat, not a finding: the scoped lint's `rc=0` decides this in
  neither direction. `internal/spec/lint.go:765` tests `lifecycle` for presence, never for
  membership, and it reported the same file clean at v0.1.0 while the value was invalid. The repair
  is verified against the schema SSOT's Field Reference by eye. Anyone later reading the green lint
  as confirmation of this field would be reading a check whose non-execution is indistinguishable
  from its success.

- **[N/A] MP-4 Section 22 language neutrality.** Unchanged from iter-1 — single-project scope, no
  per-language toolchain named. Auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Re-verified in iter-1 (3 referenced SPECs, statuses
  `completed` / `completed` / `in-progress`, none retired/superseded/archived). The `related_specs`
  list is unchanged at v0.2.0. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` → `0`. Auto-PASS.

- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BACKLOG-HYGIENE-001/`
  → exit 1, no matches. See § Declined items for the judgement on §F Q1/Q2.

---

## Regression Check — every iter-1 defect

| iter-1 | Status | Deciding evidence |
|---|---|---|
| **MP-3** `lifecycle: spec-first` | **RESOLVED** | `spec.md:12` → `spec-anchored`; enum membership checked against the schema SSOT |
| **D1** unfetched, unpinned refs | **RESOLVED** | `grep -c 'git fetch'` → 1 in each of spec.md / plan.md / **acceptance.md**. plan.md M1 Step 1 runs `git fetch origin develop main` + `git rev-parse origin/develop origin/main` before any card is read; REQ-BH-009 (`spec.md:254-267`) requires the refresh, the pin, and a citation of the pinned ref SHA per verdict; AC-BH-007 reds an entry that names a branch instead of a pinned SHA or omits the `--is-ancestor` exit code, with the stated mutation probe covering both. The pin also answers the moving-ref half — `verification-completeness.md` §4's discriminator confirms this is the pin-it case (a per-card claim), not the provenance carve-out |
| **D2** tier budget 23 > 16 | **RESOLVED** | 16 requirements + 16 acceptance criteria, both measured. plan.md §A restates the real ceiling table (M = 16/16, L = 25/25) and states the independence rule. Obligation preservation verified separately below — the load-bearing half |
| **D3** write-boundary contradiction | **RESOLVED** | §D.2 preamble and §E restate the invariant behaviourally ("what is forbidden is changing a card, not writing to disk"), grounded on `internal/cli/todo_relate.go:66` → `Mutate` → `backlog.json`. Both §E (`spec.md:318`) and acceptance.md §E (`:219`) now record that the state dir is gitignored so `git status` adjudicates nothing either way. A worker can now obey §E and REQ-BH-006 simultaneously |
| **D4** AC-BH-007 vacuous on path B | **RESOLVED** | Mutant re-run against the rewritten AC-BH-008 (`acceptance.md:112-130`): declare path B → skip the rebuild → cite the installed `moai todo pr`. The recorded measured value is `0`, so the first bullet fires *"whichever path was declared"* and the criterion reds. The v0.1.0 escape is closed, and the path-B branch now carries its own obligation (post-rebuild count ≥ 1 plus the reinstall command). One residual ambiguity, safe-direction, recorded as N3 |
| **D5** self-authored no-mutation observable | **PARTIALLY RESOLVED** | The third observable exists as REQ-BH-007 and AC-BH-006, and AC-BH-004/AC-BH-005 are retained deliberately with the reasoning stated in-line. But the digest has no procedure — see N1. The structural fix landed; the mechanism did not |
| **D6** five REQs with no deciding AC | **RESOLVED** | acceptance.md §B carries a 16-row REQ↔AC map; all 16 rows are populated. Spot-verified the five that were uncovered: REQ-BH-003's delta branch → AC-BH-001's third conjunct ("either 'no delta' or the delta with the re-derived in-scope set" — a sweep that never compared cannot satisfy it); REQ-BH-012's **positive** direction → AC-BH-009's second conjunct against M1's captured worktree list, with the empty-set vacuity named explicitly; REQ-BH-013 → the new AC-BH-011; REQ-BH-014's five-section obligation → AC-BH-012, widened from `falsified` to "every one of the 67 card entries — not only the `falsified` ones". Folded into existing criteria rather than added as siblings, as instructed, so the AC count landed at exactly 16 |
| **D7** `Where` for a runtime condition | **RESOLVED** | REQ-BH-006 `When …`; REQ-BH-012 `While …` |
| **D8** passive REQ-BH-010 | **RESOLVED** | REQ-BH-009 now reads "The sweep shall refresh … pin … and establish …" |
| **D9** batch sizes misstated | **RESOLVED** | plan.md M3 states "Measured sizes: B1=10, B2=11, B3=13, B4=10, B5=13, B6=10, total 67" — byte-identical to my iter-1 measurement, with the correction attributed |
| **D10** embedded-newline caveat | **RESOLVED** | §B.4 replaced with the arithmetic (5 + 91 = 96 = the card count) and an explicit retraction note |
| **D11** t256 dropped | **RESOLVED** | spec.md §C "Out of Scope — dropped cards" gains the clause; plan.md M4 states the relation is "reported as a reading only and never carried into the disposition proposal" |

**No iter-1 defect is unfixed.** One (D5) is partially fixed; two new defects were introduced by the
repair (N1, N2) and one pre-existing ambiguity surfaced under closer reading (N3).

---

## The load-bearing check: did the 23 → 16 merge lose an obligation?

*Claim*: every obligation carried by a v0.1.0 requirement survives in v0.2.0.

*Evidence*: mapped requirement by requirement against the v0.1.0 text read in iteration 1 and the
v0.2.0 text read fresh this iteration. All 23 accounted for; no v0.2.0 requirement is a stub.

| v0.1.0 | → v0.2.0 | Obligation carried across |
|---|---|---|
| 001 single captured snapshot, never live per-card | 001 | verbatim |
| 003 no truncating filter | 001 | "shall not read the queue through a live per-card invocation or through a truncating filter (`head`, `tail`, a `grep` that discards rows)" — both halves, plus the "read in full" strengthening |
| 002 capture time + tree HEAD | 002 | verbatim |
| 004 delta → re-derive + record | 003 | verbatim, plus a new in-scope-set definition (picked and dropped excluded) that v0.1.0 left implicit |
| 005 read-only verb list | 004 | same five verbs, plus an explicit pointer to the one recording verb |
| 006 shall not invoke the 7 mutating verbs | 005 | same 7 verbs, **plus** the behavioural form ("shall not drop, edit, close, reorder, unpick, or pick any card") and the overlap-confirmed carve-in |
| 020 shall not propose a merge as a decision | 005 + 015 | the prohibition lands in 005 ("including a card confirmed to overlap another"); the positive form lands in 015 ("A merge or absorption is proposed, never performed — the fold is the operator's") |
| 007 record overlap with `relate --relation --note` | 006 | verbatim, modality corrected `Where`→`When` |
| 008 evidence log carries every invocation verbatim | 007 | preserved as the first of the two observables |
| — | 007 (new) | the card-row digest, added per iter-1 D5 |
| 009 while binary lags, do not cite `moai todo pr` | 008 | verbatim |
| 010 query both refs, cite SHA + `--is-ancestor` | 009 | preserved and strengthened with the fetch, the pin, and the pinned-SHA citation |
| 011 no landing from a branch name | 010 | verbatim, t342 misattribution rationale retained |
| 012 unanswerable → `unknown`, never `not-landed` | 011 | verbatim |
| 013 live worktree → `in-flight-unlanded` + branch + tip SHA | 012 | verbatim, modality corrected `Where`→`While` |
| 014 restate premise + three-valued verdict | 013 | verbatim |
| 018 undecidable → `unverified` with reason; no promotion | 013 | both halves — "where an `unverified` verdict carries the reason it could not be decided" and "A plausible reading shall not be promoted to a verdict" |
| 016 five evidence sections on every entry | 014 | verbatim ("Every card entry shall carry the five evidence sections") |
| 015 `falsified` carries command + verbatim output | 014 | verbatim, with the VCI §2 citation retained |
| 017 absence claim names its scanned scope | 014 | verbatim, including the "an absence claim whose scanned scope is unnamed is a Gap, not a finding" clause |
| 019 compare against the other 66, name the shared artifact | 015 | verbatim |
| 021 one entry per in-scope card, count = queued − 1 | 016 | verbatim |
| 022 disposition proposal list, enum + single evidence | 016 | verbatim, same five-value enum |
| 023 state that no mutation was performed, awaiting operator | 016 | preserved with a **deliberate wording change**: "no queue mutation was performed" → "no card was mutated" |

*Baseline-attribution*: v0.1.0 text as read in iteration 1 at HEAD `15453140a`; v0.2.0 text as read
this iteration at the same HEAD (the artifacts are untracked and were rewritten in place, so the
HEAD is common to both and does not distinguish them — the version field does: `0.1.0` → `0.2.0`).

*Gaps*: the v0.1.0 comparison is against my iteration-1 read of that file, not against a stored
copy — the artifacts are untracked and rewritten in place, so the pre-repair text no longer exists
on disk and cannot be re-read. Anyone re-checking this table has the same constraint. Two v0.1.0
framings were dropped rather than merged, and I record them so the claim is not overstated: REQ-018's
"within the card's own reading budget" qualifier, and REQ-023's "in its own words". Neither is an
obligation — the first is a rationale for when `unverified` applies, the second a stylistic hedge.

*Residual-risk*: the one substantive wording change (023 → 016) is a **correction, not a loss**. The
v0.1.0 form would have required the report to assert something false: `moai todo relate` does mutate
the queue store, so a report claiming "no queue mutation was performed" while having appended
findings would be an unobserved claim in the report the SPEC exists to produce. The v0.2.0 form
asserts what is actually true and actually checked. I flag it explicitly because a reviewer scanning
for dropped words would see a weakened claim where the opposite happened.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | iter-1 | iter-2 | Band | Evidence |
|-----------|-------:|-------:|------|----------|
| Clarity | 0.85 | **0.90** | 0.75 (upper) | Both `Where` misuses corrected; REQ-BH-009 given an acting subject. Remaining: AC-BH-008's "the recorded measured value" is ambiguous once a path-B rebuild has recorded two counts (N3), and REQ-BH-001 merges a positive and a negative modality into one entry (readable, no ambiguity introduced) |
| Completeness | 0.72 | **0.85** | 0.75 | Frontmatter valid; five `### Out of Scope — <topic>` H3 sub-headings retained at `spec.md:172,178,185,192,199`; D1's fetch/pin and D3's carve-out both landed; plan.md gains B6/B7 and two new anti-patterns. Held below 0.90 solely by N1 — the one new mechanism the revision introduces is the one it does not specify |
| Testability | 0.80 | **0.82** | 0.75 | Weasel-word scan over `acceptance.md` clean (unchanged). AC-BH-008's mutant is killed; AC-BH-009's empty-set vacuity is closed with a positive conjunct; AC-BH-012 widened to every entry; four criteria now carry stated mutation probes, two carry explicit anti-vacuity reasoning. Barely moved because N1 subtracts what the other repairs added: the criterion added to make the invariant decidable is itself not executable as written, and its probe is one-directional |
| Traceability | 0.60 | **0.85** | 0.75 | The largest single gain, and it is real: acceptance.md §B carries a 16-row REQ↔AC map, all 16 rows populated, and all 16 criteria appear somewhere in it. Milestone closure independently covers the same set — `grep -n 'Closes:' plan.md` yields 008..012 / 001..003 + 007 / 013,014 / 006,015 / 004,005,007,016, whose union is exactly 001-016 with no requirement unclosed. Held below 0.95 by N2 (one false cell) and by the AC-BH-006 cell being aspirational until N1 is fixed |

**Aggregate 0.85** (harmonic; arithmetic 0.855) vs the Tier M threshold **0.80**.

---

## Defects Found (structured defect-list)

### N1. DIGEST-HAS-NO-PROCEDURE — acceptance.md:83-98 (AC-BH-006), spec.md:243-246 (REQ-BH-007), plan.md:127 (M2) — Severity: major — Class: blocking — **NEW, introduced by the D5 repair**

*Claim*: AC-BH-006 cannot be executed as written, and its Findings-exclusion is stated as an
intention rather than as a mechanism — so the author's own flagged risk is confirmed, and it is
worse than "could accidentally include `Findings`".

*Evidence*: `grep -n 'digest'` across all three artifacts returns 14 hits describing the digest in
prose and **zero** naming a command. `grep -n 'shasum\|sha256\|jq \|backlog.json' acceptance.md`
returns exactly one hit, `:220`, which is §E's note that `git status` cannot see `backlog.json` —
not a digest procedure. Compare AC-BH-001, which names
`cut -f2 $R/queue-snapshot.tsv | grep -c '^queued$'`. AC-BH-006 is the only one of the 16 criteria
with no deciding command, and it is the one iter-1 asked for.

*Baseline-attribution*: measured this run against v0.2.0 in the t332 worktree at HEAD `15453140a`.

*Gaps*: I did not attempt to write the extraction myself, so I have not established that a
`Findings`-excluding extraction is straightforward against this `backlog.json` schema. I read
`internal/cli/todo_relate.go` far enough to confirm `relate` appends to `rec.Findings` (iteration 1),
not far enough to enumerate the record's card-row fields.

*Residual-risk*: three consequences, in increasing order of how quietly they fail. (a) Two
independently-improvised extractions at M2 and M5 make "byte-identical" a comparison of
incomparable things — and the comparison would red, sending the run phase hunting for a mutation
that never happened. (b) An extraction that reads `backlog.json` wholesale reds on a legitimate
`relate`, which REQ-BH-006 mandates — the criterion built to be the reliable observable becomes the
one that cries wolf. (c) The stated mutation probe covers only the true-positive direction
("run `moai todo edit` … → the digests differ"). Nothing probes the other direction, and per
`verification-completeness.md` §2's rule-pairing corollary a single-direction criterion admits a
mutant on the direction it does not test. Here that mutant is concrete: an extraction including
`Findings` passes the stated probe perfectly while being wrong.

*Required fix* (one sentence plus one probe, no ceiling cost): name the extraction in AC-BH-006 —
e.g. "the digest is `<extraction over the card rows of the queue store, excluding Findings> | shasum
-a 256`, the same invocation at M2 and M5, recorded verbatim in `01-scope.md` beside its output" —
and add the negative control: "Control: run the M4 `moai todo relate` between the two captures →
the digests are unchanged." That second probe is what proves the exclusion works rather than
asserting it, and it costs nothing: M4 already runs `relate` between M2 and M5, so the control is
observed by the sweep's own ordering rather than staged.

### N2. AC-BH-015 IS MIS-TRACED — acceptance.md:31 (map row), :192-197 (the criterion) — Severity: minor — Class: optional — **NEW, introduced by the D6 repair**

*Claim*: the map's row `REQ-BH-007 | AC-BH-004, AC-BH-006, AC-BH-015` asserts a relation that does
not hold, so the map's own header claim ("every criterion decides at least one requirement") is
formally satisfied and substantively false in one cell.

*Evidence*: REQ-BH-007 requires "two independent observables of the no-mutation invariant: a
verbatim log … and a digest". AC-BH-015 tests that the per-batch card-id sets are pairwise disjoint
and that their union equals the in-scope set. Nothing in it decides anything about either
observable. What it actually decides is `spec.md` §E's constraint "Read-only fan-out workers write
only their own file; two workers never share an output path" — a constraint, and no requirement
carries it.

*Baseline-attribution*: both texts read this run at v0.2.0.

*Gaps*: I did not check the other 15 map cells to the same depth. Spot-checks on the five D6 cells
and on 002 / 010 / 016 were correct.

*Residual-risk*: a map that is right in 15 cells and wrong in one is more dangerous than no map,
because the next reader trusts the row rather than re-deriving it. The declined-item reasoning
(don't spend a ceiling slot on a new requirement for it) is **correct and should stand** — the
error is not the decision to leave AC-BH-015 without a requirement, it is recording that absence as
a presence.

*Required fix*: drop `AC-BH-015` from REQ-BH-007's cell and add a footnote under the map: "AC-BH-015
decides `spec.md` §E's fan-out write-isolation constraint; no requirement carries it, deliberately —
the constraint is a plan-phase property of the fan-out, not an obligation on the sweep's output."
Truthful, one line, no slot spent.

### N3. "THE RECORDED MEASURED VALUE" IS AMBIGUOUS AFTER A PATH-B REBUILD — acceptance.md:119-122 — Severity: minor — Class: optional

*Claim*: on the path-B branch AC-BH-008 records two `strings` counts (pre-rebuild `0`, post-rebuild
`≥ 1`) and its first bullet does not say which one "the recorded measured value" denotes.

*Evidence*: the two bullets at `:119` and `:121` are conjunctive; the first fires on `0`, the second
requires the post-rebuild count. A reader taking "the recorded measured value" as the pre-rebuild
`0` would forbid citing `moai todo pr` even after a valid rebuild.

*Baseline-attribution*: read this run at v0.2.0.

*Gaps*: none — this is a reading of the criterion's text, not a claim about behaviour.

*Residual-risk*: **low, and it errs safe.** The ambiguity over-blocks rather than under-blocks, and
REQ-BH-008 disambiguates it correctly on its own terms: the prohibition holds *"While the installed
binary lacks the `worktree_base_branch` string"*, which a successful rebuild ends. The requirement
governs; the criterion is merely less precise than the requirement it decides. I record it because
the whole point of the D4 repair was to remove a reading under which the guard does not bind.

*Required fix*: in the first bullet, say "the **operative** measured value (the post-rebuild count
where path B was taken, otherwise the measured count)".

---

## Declined items — judgement

**§F Q1/Q2 left as prose Open Questions rather than `[NEEDS CLARIFICATION]` markers — ACCEPTABLE.**
The marker convention exists so the orchestrator resolves an *unresolved* topic via
`AskUserQuestion` before the plan→run gate. Neither question is unresolved in that sense. Q1 is
decided by M1 by measurement, has a stated default (path A), and is bound on **both** branches by
REQ-BH-008 and AC-BH-008 — a decision with a recorded procedure, not an open question. Q2 is an
operator confirmation of a Tier that is now inside its budget with the grounds stated in plan.md §A;
its default is safe and the operator can move it at the kickoff gate, which they reach regardless.
Marking either would route a decided item into a clarification round. MP-7 passes on the current
form, and it would still pass on the marked form — this is not a case where the gate is being
dodged by choosing the unmarked wording.

**AC-BH-015 traced into REQ-BH-007's row rather than given its own requirement — the DECISION is
acceptable, the RECORDING is not.** Declining to spend a ceiling slot on a requirement whose only
purpose is to host an existing criterion is right: it would inflate the count for a bookkeeping
reason, which is exactly what D2 was about. But the map cell asserts a relation that does not exist.
Fix the record, keep the decision — see N2.

---

## Tier judgement

**Tier M, confirmed.** iter-1 said "Tier M is the right classification but the SPEC does not fit
inside it"; the second clause no longer holds. 16 requirements and 16 acceptance criteria, both at
the ceiling and neither over it, with the ceilings applied independently as the SSOT requires. The
complexity signals are unchanged and still read M: no source file lands, no schema moves, the
artifact set is the Tier M three plus `progress.md`, and the one genuine decision is M1's landing
method. plan.md §A's refusal to tier up to L "to legalize the count" is the correct call and is now
argued from the real ceiling table rather than from a misread row.

Sitting exactly at 16/16 is worth one caution: any further requirement or criterion — including the
fix for N1 if it were written as a new sibling — breaches the budget. Both required fixes above are
deliberately shaped to fit inside existing entries.

---

## Recommendation

**PASS-WITH-DEBT at 0.85.** All seven must-pass criteria pass, the aggregate clears the Tier M
threshold by 0.05, no iter-1 defect is unfixed, and the consolidation claim is verified rather than
accepted. The debt is **N1**, and it should be discharged before M2 runs rather than by a third plan
iteration:

1. **Before M2 captures the opening digest**, name the extraction command in AC-BH-006 and add the
   `relate`-direction control probe (N1). M2 cannot capture a digest without choosing an extraction
   anyway — writing it into the criterion first is what makes the choice reviewable instead of
   improvised, and it is the difference between a criterion that decides REQ-BH-007 and one that
   merely mentions it.
2. Fix the map cell and add the footnote (N2) — one line, no ceiling cost.
3. Optional: disambiguate AC-BH-008's first bullet (N3).

**I am naming the judgement call rather than burying it.** A FAIL is defensible here: N1 is
classified blocking, and it sits on the one criterion iteration 1 created. I did not choose FAIL
because the M5 firewall is clean, the aggregate clears the threshold, and the defect is a missing
sentence in a criterion whose requirement, milestone placement, and purpose are all correctly
specified — the run phase is forced to confront it at M2 by construction. If the operator would
rather the extraction be named before run-entry than at M2, treating this verdict as FAIL and
returning it for a one-line fix is a reasonable override, and iteration 2 is the Tier M ceiling, so
that fix would land as a direct revision rather than as an iteration 3.

---

## Gaps — what this audit did NOT observe

- **Scope.** This is a delta-scoped re-audit per the Retry Loop Contract. I did not re-derive the
  full Group 1-8 checklist from scratch; MP-4 and MP-5 are carried from iteration 1 on the grounds
  that the repair touched neither surface (`related_specs` unchanged; no per-language tooling
  introduced). A defect outside the delta and outside those two would not have been caught here.
- **No lint signal of my own.** I ran no `moai spec lint` in any form, scoped or corpus. The
  coordinator's report of `rc=0` under `--strict` is consumed as stated evidence, not observed —
  and, as recorded under MP-3, it decides the one must-pass this iteration turned on in neither
  direction.
- **No git verification.** This worktree's guard refuses compound commands containing `git`, and I
  did not split out the ancestry checks because nothing in the delta depends on them: D1's repair is
  a property of the SPEC text (does it require the fetch and the pin?), not of the refs' current
  state. So §B.2, §B.3's `260ea5369` attribution, and the t342 control queries remain consumed as
  the author's evidence. The one measurement I re-took independently in iteration 1 —
  `strings ~/go/bin/moai | grep -c 'worktree_base_branch'` → `0` — I did not re-take this iteration;
  a rebuild between the two would invalidate it, and I have no evidence either way.
- **`git worktree list` still not run.** AC-BH-009's second conjunct and plan.md M1 Step 3 both rest
  on a live worktree list. I verified that M1 now *captures* it (closing the input gap iteration 1
  would otherwise have left in AC-BH-009), but not that the list contains what §B.5 says it does.
- **The v0.1.0 text is gone.** The obligation-preservation table compares v0.2.0 on disk against my
  iteration-1 read of v0.1.0. The artifacts are untracked and were rewritten in place, so no stored
  copy exists to diff against. My read is the only surviving record of the pre-repair text, and it
  is a read, not a diff.
- **11 of 16 map cells spot-checked, not all 16.** N2 was found in one of the five I checked closely;
  I did not audit the remainder to the same depth.
- **No mutation testing executed.** Four criteria now carry stated probes; all four were reasoned
  about on paper. The artifacts they would mutate still do not exist.
- **`harness.plan_audit_tier_ceilings` not read**; the "2/2" label assumes the documented Tier M
  ceiling of 2.

## Residual risk

**The score moved up, and that is the finding most likely to be over-read.** 0.74 → 0.85 measures
the repair of the defects iteration 1 enumerated; it does not measure the SPEC against defects
neither iteration looked for, and a delta-scoped re-audit cannot. The clean regression table means
"every named defect was addressed", not "no defect remains".

**N1 is the specific way this SPEC can still produce a confidently-wrong result.** The three-observable
design is now correct on paper, and the observable that carries the design — the only one not
authored by the run phase about itself — is the one that has no method. If M2 improvises an
extraction that includes `Findings`, AC-BH-006 reds on the sweep's own mandated `relate` and the run
phase spends its budget chasing a mutation that never occurred; if it improvises differently at M5,
the comparison is meaningless in the other direction. Both failures present as digest trouble rather
than as a specification gap, which is what makes the missing sentence expensive.

**And the standing caveat from MP-3 has not gone away just because the field is now valid.** The
linter reported this file clean while `lifecycle` was wrong, and it reports it clean now. Nothing in
the gating signal changed between those two states. Any later reader taking `rc=0` as confirmation of
this SPEC's frontmatter is reading a check whose non-execution is indistinguishable from its success
— which is the defect class this SPEC was written to hunt, sitting in the tooling that grades it.
