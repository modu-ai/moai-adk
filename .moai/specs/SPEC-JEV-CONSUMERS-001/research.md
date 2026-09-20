# SPEC-JEV-CONSUMERS-001 — Research

Read-only findings from this worktree at `fd75cf692`. Anything inherited rather than measured here is labelled as such.

---

## §1. The backlog finding contract

`internal/kanban/backlog_store.go` carries every piece Consumer C extends.

**Two source constants (lines 117-120)**, partitioned by observer: `mechanical` = the text analyser measured it; `agent` = a reader who understands the cards judged it, written through `todo relate`.

**`Source` has no validator.** It is a plain `string` field; `AppendFindingOnce` dedupes on the tuple and checks no membership. But no CLI path accepts an arbitrary value either — `todo_relate.go:75` hardcodes `BacklogSourceAgent`, and `todo_analysis.go:62,71` hardcode `BacklogSourceMechanical`. The set is closed **in practice by its write paths, not by a check**, which means a third source needs a new write path rather than a new flag, and no validation code needs changing.

**`HasAgentFindingForPair` (lines 408-416)** selects on `Source == BacklogSourceAgent`. Its doc comment states it is the predicate behind the `machine-only` mark and that the mark records the *absence* of an agent-sourced record for a pair, "never that any review took place". This is the constraint forcing a third constant.

**`SamePairAs`** compares two findings on the **unordered** pair, on the stated reasoning that a relation between two cards is a property of the pair rather than of the direction it was written in. Any precedence rule (open question N1) operates on that unordered comparison.

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

`internal/cli/todo_triage_test.go` imports `crypto/sha256` (line 21) and computes `sha256.Sum256(data)` (line 649). AC-JEVN-005 reuses this.

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
| N1 | Dedup precedence between a Jev finding and a mechanical or agent finding on the same unordered pair | Carried from the pre-split SPEC as a proposal. Needs a decision before M4; AC-JEVN-003 asserts whichever rule is chosen. |
| Q3 | Labelled-set size per consumer | Owned by `SPEC-JEV-OPTIN-MEASURE-001`; depends on each consumer's base rate, which its harness measures. |
| N2 | Where routing is invoked from | The lane-question surface is prose-level today; which code path (if any) hosts Consumer A is an implementation-time question this research pass did not resolve. |
