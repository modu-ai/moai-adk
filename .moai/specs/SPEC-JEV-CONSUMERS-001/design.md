# SPEC-JEV-CONSUMERS-001 — Design

---

## §1. Three consumers, one client, nothing shared between them

```
  consumer C (near-duplicate) ─┐
  consumer A (routing)        ─┼──▶ internal/jev ──▶ POST /v1/systemone
  consumer B (skill suggest)  ─┘
```

The three share the client and nothing else: different state shapes, different question sets, different thresholds, different measurements, different gates. There is deliberately no "consumer framework" layer, because the only property they have in common is the one the client already provides, and a shared layer would couple three gates that must be able to fail independently.

Only one of the three writes anything. Consumer C appends a `BacklogFinding`; A and B produce values that a reader consumes. That asymmetry is why C carries most of this design.

## §2. The third finding source

### The problem

`BacklogFinding.Source` is a free string with two constants, and they partition the finding space by *who observed it*: `mechanical` means the text analyser measured it, `agent` means a reader who understands what the cards mean judged it. `HasAgentFindingForPair` selects on `agent` and drives the `machine-only` mark, whose documented meaning is "nothing agent-sourced was recorded here" — an honest claim precisely because the CLI cannot know who called it.

A Jev answer is neither. It is not a measurement of text similarity, and it is not a reader who understands the cards. Filing it under either existing constant makes the queue assert something false:

- Under `mechanical`, it renders with a `score` that reads as a measured similarity — the text analyser's number and a model's confidence printed in the same shape, distinguishable by nothing.
- Under `agent`, it clears the `machine-only` mark from pairs nobody reviewed. The queue then asserts a review took place. This is the more dangerous direction, because its failure is silent: a wrong "unreviewed" mark gets noticed by whoever expected it; a wrongly-cleared one is invisible by construction.

### The decision

A third constant, `jev`. Three consequences, each a design choice rather than a mechanical follow-on:

**Precedence — DECIDED 2026-09-20 (operator decision, confirmed exactly as proposed).** A Jev finding is appended only when no finding of any source already names that unordered pair with the same relation; a later mechanical or agent finding is appended alongside rather than replacing it. The asymmetry is deliberate — a model signal must never suppress a measurement or a judgement, and must never be silently suppressed by one, because both suppressions make the queue quieter than the evidence warrants. This is a settled decision, stated as `REQ-JEVN-006` halves (a) and (b) and asserted by `AC-JEVN-003` across all four source combinations. Open question N1 is closed.

**What the decision costs, which the rule's symmetry hides.** The two halves read as one rule and are two pieces of work of opposite size.

Half (b) — a later mechanical or agent finding lands alongside — is **already true and needs no code**. `AppendFindingOnce` (`internal/kanban/backlog_store.go:369`) delegates to `HasFindingTuple` (`:357`), whose key is `{SubjectID, RelatedID, Relation, Source}`. Because `Source` is part of that key, a mechanical finding never matches an existing Jev tuple and is always appended. The requirement exists so that a future change to the dedup key is recognised as breaking this behaviour rather than as an unrelated refactor.

Half (a) — suppressing an arriving Jev finding on a pair any source already names — **cannot be expressed by either existing path**, and the reason is the same property that makes (b) free. The source-inclusive key means a Jev finding is *never* suppressed by a mechanical or agent one, so `AppendFindingOnce` gives the opposite of the rule; and `AppendFindingOnce` never calls `SamePairAs` (`:159`), whose only non-test caller in this repository is `HasAgentFindingForPair` (`:409`) — so the unordered comparison is not on the append path at all. Half (a) therefore requires a **new source-agnostic unordered predicate**, provisionally `HasFindingForPairAnySource`, keyed on the unordered pair plus the relation and deliberately not on `Source`.

The failure mode if this is missed is quiet: an implementation that "reuses the existing dedup" ships a rule that suppresses nothing, and an un-suppressed append is indistinguishable from a correct one at every surface a reader looks at.

**Render.** The existing rule prints a score only for `mechanical`, on the reasoning stated in the source: an agent judgement carries no measurement, and `0.00` would read as measured dissimilarity. A Jev probability is a third thing again — a calibrated model confidence — so it renders with its own label. The default behaviour of a new source is to inherit one of the two existing branches, and both are wrong here, which is why the render is a requirement rather than an implementation detail.

**Mark semantics.** The `machine-only` mark's meaning is unchanged. It continues to mean "no agent-sourced record for this pair", and a Jev finding does not satisfy it. `AC-JEVN-001` pins this with a positive control, because an assertion that a predicate returns false is satisfied equally well by a predicate that always returns false.

That control proves the **predicate** behaves; it does not prove the **write path** obeys the prohibition, and the two are different claims. `REQ-JEVN-003` forbids a Jev finding ever being *written* as `agent` — a prohibition on producers, which a predicate test over a hand-built fixture cannot reach, since the fixture's `Source` is whatever the test itself set. `AC-JEVN-012` closes that gap on the chain's own precedent (`SPEC-JEV-CORE-001` `AC-JEVC-003`): a search over the producing packages for `BacklogSourceAgent`, carrying a positive control on `internal/cli/todo_relate.go:76`, which does set it. An absence measured without a control is a gap rather than a finding — the search that matches nothing and the search that never ran produce the same output.

### The alternative, and why it was rejected

Reusing `agent` with a note reading "written by Jev" fails on a structural point: the note is prose, `HasAgentFindingForPair` reads the constant, and no reader of the predicate would ever see the note. A distinction that only exists in a free-text field is not a distinction the code makes.

## §3. Why findings stay inert

The `BacklogFinding` doc comment records that a finding is a record and nothing else, and that this is a *structural* property rather than a convention: folding, reordering, dropping, and editing a card in response to a finding would each require code that does not exist, so none of them can happen by accident.

Consumer C inherits that property for free. The design obligation is only to not be the thing that breaks it — which is why REQ-JEVN-007 states it explicitly and AC-JEVN-005 verifies it against the queue file's hash rather than against a code review.

**That inertness is now load-bearing at the SPEC layer, not merely reassuring.** `SPEC-JEV-CORE-001` v0.2.0 (card t1066) makes it condition (i) of the REQ-JEVC-011 inert-sibling-record carve-out — the clause under which Consumer C's append is authorised at all. Before that amendment the defence lived only here, in a design document, while a plain reading of REQ-JEVC-011 and REQ-JEVC-012 forbade the append outright; the wording, not the design, was the defect. Condition (iii) of that carve-out — nothing selects on the record — is the one a future change can quietly remove, and removing it lapses the carve-out rather than extending it.

## §4. Routing — the classification, not the decision

> **Host unresolved — this section designs the question set, not the call site.** Which code path invokes routing is open question N2, unanswered by the 2026-09-20 operator decisions, and M5 is declared blocked on it (`plan.md` §F). Nothing below names a host, deliberately: the lane-question surface is prose-level today, and naming a call site that was never measured would put an unverified premise exactly where a reader is least likely to check it.

Consumer A asks three questions over one state: a Choice naming the decision owner, and two Nouls (needs-a-measurement-first, cheap-to-reverse). Batched into one request, both for cost — input tokens are charged, output tokens are not, so N requests over one state pay the state N times — and for consistency, since separate requests over identical input could return mutually inconsistent judgments.

Two design points carry the safety:

**Uncertainty escalates, never downgrades.** An `operator` answer routes to operator; so does a Choice confidence below the fitted threshold. The two triggers are separate sub-cases in AC-JEVN-007 because they fail differently — one is the model being confident about a hard case, the other is the model not knowing.

**The no-match option.** Every Choice carries one. A forced choice over an option set that excludes the true answer produces a confident wrong answer, and this repository has recorded that failure from its own dispatch practice independently of any model: a question offering two branches got a confident answer when the true answer was a third thing.

## §5. Skill suggestion — two requests, and the one that matters

> **Host unresolved, and additionally outside the quality gate's reach.** Consumer B's host is the `/moai` intent router, a prose-level skill surface that `acceptance.md` §E's `go test ./internal/kanban/... ./internal/cli/...` does not cover. M6 is therefore blocked on N2 alongside M5 (`plan.md` §F), and the verification surface for its criteria is part of answering N2 rather than an assumption this design may make.

Consumer B issues a wide rank over all skills batched with a Noul asking whether the turn needs a skill at all, then reranks the top three under fuller text.

The Noul is the load-bearing half, not the ranking. A ranker over a fixed skill set always returns a best skill, including for turns that need none — and a presented "best" is read as a recommendation regardless of how weak it was. Suppressing the list when the Noul answers negatively above threshold (REQ-JEVN-014) is what stops the consumer manufacturing suggestions out of turns that had no use for one.

The two-request shape exists because the two stages want different state sizes: the wide rank needs every skill's short description and would blow the state budget with full text, while the rerank needs fuller text on three candidates and can afford it. Since accuracy falls as irrelevant state grows, one large request over everything would be worse than two targeted ones.

The intent router's authority is untouched. AC-JEVN-011 verifies this by comparing the router's selection against its selection with the capability disabled — the strongest available form of "it changed nothing".

## §6. Why each consumer is independently gated

The three could plausibly share one gate. They do not, for the reason the whole chain exists: the rejected premise-death task demonstrated that this model's usefulness is per-task, not global. It scored 58.9% where a constant answer scored 75.0% — on one task. That says nothing about near-duplicate marking, and near-duplicate marking's result will say nothing about routing.

A shared gate would also create a failure mode with no good resolution: two consumers pass, one fails, and the shipped set is now a judgement call made under pressure. Independent gates make the answer mechanical — the failing consumer does not ship, the other two do, and the absence is recorded as a decision.

### The third and fourth states: a gate that cannot be run, and one that has not been run

"Passed" and "failed" are not the only outcomes, and treating them as such leaves run-phase with nothing to produce in the state it is most likely to be in. `Report.Verdict()` (`internal/jevmeasure/measure.go:193`) first-checks the measurement's source and returns `VerdictWithhold` for anything other than `SourceLive`, so where no live measurement is permitted, **no consumer can reach ship** — and the gate has not failed, it has not run.

The distinction is not pedantic, because the two states call for different records. A failed gate is evidence: the record cites the measurement and the baseline it did not beat, and a later reader can disagree with the verdict by reading the numbers. An un-runnable gate has no measurement to cite, so a record written in the failed-gate shape has to invent one — which is precisely the unobserved-verification claim this chain exists to avoid. `REQ-JEVN-015` gives the second state its own output: a recorded decision naming why the gate could not be run and citing the withholding mechanism, with no measurement artifact required. `AC-JEVN-015` asserts that the two records stay distinguishable.

This is separate again from a milestone **blocked** on an unanswered design question (M5 and M6 on N2, `plan.md` §F). Blocked work was never started; an un-runnable gate belongs to work that was in scope and could not be judged. A milestone can be in both states at once, and each is recorded on its own terms.

**A fourth state exists, and M4 is in it.** `Report.Verdict()` withholds where no *live* measurement is permitted; it says nothing about the case where a live measurement is permitted and simply has not been taken. That case is real and common: the credential exists, the endpoint is reachable, and what is missing is the labelled set — whose size is open question Q3, owned by the predecessor. The gate has not failed and it is not un-runnable; it is **unrun**, and the work that would run it is owed rather than closed.

The distinction matters for the same reason the third state does, and in the same direction: the record shapes differ. A gate-not-runnable record cites a withholding mechanism and closes; a gate-unrun record names what is still missing and who owns it, so the owed work survives the milestone. Writing the second in the first's shape retires work nobody decided to retire.

It matters a second way that the third state does not: a gate that cannot be run has nothing to guard, so `REQ-JEVN-015` can require the call path to be **absent** and lose nothing. An unrun gate will be run, so requiring absence would mean deleting and re-landing the implementation between the decision and the measurement — cost with no safety return, since `REQ-JEVC-017` already guarantees a default-off consumer constructs nothing and `REQ-JEVC-014` already guarantees its output is unchanged. `REQ-JEVN-016` therefore permits presence and buys back the safety with four conditions that are each mechanically checkable (`AC-JEVN-016`), rather than with a deletion that is not.
