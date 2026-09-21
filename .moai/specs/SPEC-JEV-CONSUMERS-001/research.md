# SPEC-JEV-CONSUMERS-001 — Research

Read-only findings from this worktree. The 0.1.0 pass measured at `fd75cf692`; revision 0.2.0 re-measured every cited coordinate at HEAD **`7e1ed63b9`** and every coordinate below now resolves there. Anything inherited rather than measured here is labelled as such.

---

## §1. The backlog finding contract

`internal/kanban/backlog_store.go` carries every piece Consumer C extends.

**Two source constants (`BacklogSourceMechanical` at line 118, `BacklogSourceAgent` at line 120)**, partitioned by observer: `mechanical` = the text analyser measured it; `agent` = a reader who understands the cards judged it, written through `todo relate`.

**`Source` has no validator.** It is a plain `string` field; `AppendFindingOnce` dedupes on the tuple and checks no membership. But no CLI path accepts an arbitrary value either — `todo_relate.go:76` hardcodes `BacklogSourceAgent`, and `todo_analysis.go:64`, `:73`, and `:179` hardcode `BacklogSourceMechanical`. The set is closed **in practice by its write paths, not by a check**, which means a third source needs a new write path rather than a new flag, and no validation code needs changing.

> **Coordinate correction (revision 0.2.0).** The 0.1.0 pass cited `todo_relate.go:75` and `todo_analysis.go:62,71`, and named two mechanical write sites where there are three. The cited lines hold other struct fields (`RelatedID` at `:62` and `:71`, `Relation` at `todo_relate.go:75`) — at HEAD `7e1ed63b9` and at the declared baseline `fd75cf692` alike, with no commit touching either file between them. These were miscitations from the start, not drift; the record says so rather than describing a movement that did not happen.

**Three mechanical write sites, and the third raises a scope question.** `:64` and `:73` sit in `appendAnalyzedCard` (declared `internal/cli/todo_analysis.go:38`) — the card-admission path. `:179` sits in `analyzeQueue` (declared `:140`) — the `moai todo analyze` re-sweep, a separate entry point that re-walks the queue on demand.

The inventory's original conclusion survives: the third site is also `mechanical`, so the source set is still closed by its write paths. What the third site changes is scope. `REQ-JEVN-001` triggers Consumer C on admission, and the re-sweep is exactly where precedence half (b) would be exercised most often, because it revisits pairs that may already carry a Jev finding. Leaving the question unstated would have hidden a live behavioural choice inside a settled rule, so the answer is now explicit: **Consumer C is admission-only and is not invoked from the re-sweep** (`spec.md` `REQ-JEVN-001`, `plan.md` §B3, asserted by `AC-JEVN-013`). The re-sweep's own write site is read for scope and not modified.

**`HasAgentFindingForPair` (declared line 409; doc comment from line 405)** selects on `Source == BacklogSourceAgent`. Its doc comment states it is the predicate behind the `machine-only` mark and that the mark records the *absence* of an agent-sourced record for a pair, "never that any review took place". This is the constraint forcing a third constant.

**`SamePairAs` (declared line 159)** compares two findings on the **unordered** pair, on the stated reasoning that a relation between two cards is a property of the pair rather than of the direction it was written in. The settled precedence rule operates on that unordered comparison — but **not through this function as it is wired today**. `SamePairAs` has exactly one non-test caller in the repository, `HasAgentFindingForPair` (line 409), and the append path does not reach it: `AppendFindingOnce` (line 369) delegates solely to `HasFindingTuple` (line 357), whose key `{SubjectID, RelatedID, Relation, Source}` is both ordered and source-inclusive. Suppressing an arriving Jev finding on a pair any source already names therefore needs a **new** source-agnostic unordered predicate; it cannot be obtained by composing the two existing ones (`spec.md` `REQ-JEVN-006`, `design.md` §2).

**`todoFindingLine`** is source-conditioned twice: the `machine-only` mark renders only for `mechanical` findings lacking an agent counterpart, and the score renders only for `mechanical` — with the reason stated in the source, that an agent judgement carries no measurement and `0.00` would read as measured dissimilarity. A third source therefore needs a third render form, or it silently inherits the "no score" branch and loses its probability.

**Findings are records and nothing else.** The `BacklogFinding` doc comment notes this is structural rather than conventional: folding, reordering, dropping, or editing a card in response to a finding would each need code that does not exist. Consumer C inherits the property; the obligation is not to break it.

**Relations.** `BacklogSemanticRelations` lists the four `todo relate` accepts (`contains`, `absorbs`, `replaces`, `conflicts`); the two mechanical relations (`duplicate-forced`, `near-duplicate`) are deliberately absent from it, on the stated reasoning that a caller must not be able to record a measurement it did not take. Consumer C's finding is a near-duplicate judgment, which sits on the mechanical side of that line by relation but on neither side by source — which is the whole reason for the third constant.

**Tests enumerating the source set** (likely to need declaration-only updates): `internal/kanban/backlog_findings_test.go`, `backlog_archive_test.go`, `todo_merge_procedure_test.go`, `todo_queue_merge_test.go`, `internal/cli/todo_analysis_test.go`, `todo_analysis_add_test.go`, `todo_relate_test.go`. **Inherited from the orchestrator's measurement; not independently re-run in this pass.**

---

## §2. Write-path precedents

**Mechanical** (`internal/cli/todo_analysis.go:58-76`): written at card-admission time inside the store mutation, with the source comment noting the finding names the NEW card as the subject — "the card whose admission the finding explains" — and that neither branch touches a field of the card it names. Consumer C follows this shape.

**Agent** (`internal/cli/todo_relate.go:72-80`): written through an explicit verb, after relation validation and self-reference rejection, with `AppendFindingOnce` returning false on an already-recorded tuple. Consumer C does not follow this shape — there is no Jev verb, and a judgment is recorded as a side effect of admission rather than on request.

---

## §3. Queue-hash verification method

`internal/cli/todo_triage_test.go` imports `crypto/sha256` (line 21) and computes `sha256.Sum256(data)` (line 649). `AC-JEVN-005` reuses this. Both coordinates re-verified at HEAD `7e1ed63b9`.

---

## §3b. The gate's withhold mechanism — why "gate not run" is a third state

`internal/jevmeasure/measure.go` declares `func (r Report) Verdict() (Verdict, string)` at line 193, and its **first** check is the measurement's provenance: any `Source` other than `SourceLive` (declared `:134`) returns `VerdictWithhold` (`:149`) before the baseline is even consulted. `VerdictShip` (`:147`) is therefore unreachable without a live measurement.

The consequence for this SPEC is structural rather than incidental. Where no live measurement is permitted, every consumer here is withheld — and the gate has not *failed*, it has not *run*. Those are different states with different evidence: a failed gate cites a measurement and the baseline it did not beat; an un-runnable gate has no measurement to cite at all. The 0.1.0 artifacts had no output defined for the second state, which meant the Definition of Done asked for "the measurement that withheld it" in exactly the situation where none exists. `REQ-JEVN-015` and `AC-JEVN-015` close that.

---

## §4. Inherited measurement lessons that bind this SPEC's gates

**Labelled as inherited, not re-measured.** Figures originate in `.moai/reports/t943/verdict.md`, gitignored local evidence in the primary checkout that may be absent from any given tree.

The rejected premise-death task scored 58.9% 2-class against a 75.0% constant baseline, with `premise_dead` call precision at 29.2% against a 25% base rate. Three consequences bind here:

1. **Per-task, not global.** That result says nothing about near-duplicate marking, routing, or skill suggestion — which is precisely why each consumer here is gated on its own measurement rather than on a shared one.
2. **High confidence does not imply correctness.** 67 of 85 errors sat above 0.30, peaking at 0.93. A fitted threshold is a cost/coverage lever, not a correctness filter — relevant to REQ-JEVN-011 and REQ-JEVN-014, both of which use a threshold to route *away* from the model rather than to trust it more.
3. **Marker-not-judgment is the usable mode.** The same tool used as a marker narrowed 124 cards to 48 to inspect, yielding 14 — a 2.6× narrowing. Read as a judgment it skipped 34 live cards wrongly. All three consumers here are markers by construction.

---

## §5. Open items

| # | Item | Why it is not settled here |
|---|---|---|
| N1 | Dedup precedence between a Jev finding and a mechanical or agent finding on the same unordered pair | **CLOSED 2026-09-20 — settled by operator decision, exactly as proposed.** Recorded in `plan.md` §B1 and `design.md` §2; stated as `REQ-JEVN-006` halves (a) and (b); asserted by `AC-JEVN-003` across all four source combinations. Research note: the two halves have opposite implementation costs — see the `SamePairAs` entry in §1. |
| Q3 | Labelled-set size per consumer | OPEN. Owned by `SPEC-JEV-OPTIN-MEASURE-001`; depends on each consumer's base rate, which its harness measures. |
| N2 | Where routing is invoked from | **OPEN, and blocking.** The lane-question surface is prose-level today; which code path (if any) hosts Consumer A is an implementation-time question this research pass did not resolve, and the 2026-09-20 operator decisions did not cover it. M5 and M6 are declared blocked on it (`plan.md` §F) and N2 is carried to the Implementation Kickoff Approval gate. Consumer B compounds it: its host is the `/moai` intent router, a prose-level skill surface that `acceptance.md` §E's Go quality gate does not reach, so answering N2 must also name a verification surface for M6. |
