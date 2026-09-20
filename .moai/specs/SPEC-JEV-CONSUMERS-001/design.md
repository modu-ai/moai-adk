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

**Precedence — OPEN.** The proposal: a Jev finding is appended only when no finding of any source already names that unordered pair with the same relation; a later mechanical or agent finding is appended alongside rather than replacing it. The asymmetry is deliberate — a model signal must never suppress a measurement or a judgement, and must never be silently suppressed by one, because both suppressions make the queue quieter than the evidence warrants. **This remains a proposal** (open question N1) and is not settled by this document.

**Render.** The existing rule prints a score only for `mechanical`, on the reasoning stated in the source: an agent judgement carries no measurement, and `0.00` would read as measured dissimilarity. A Jev probability is a third thing again — a calibrated model confidence — so it renders with its own label. The default behaviour of a new source is to inherit one of the two existing branches, and both are wrong here, which is why the render is a requirement rather than an implementation detail.

**Mark semantics.** The `machine-only` mark's meaning is unchanged. It continues to mean "no agent-sourced record for this pair", and a Jev finding does not satisfy it. AC-JEVN-001 pins this with a positive control, because an assertion that a predicate returns false is satisfied equally well by a predicate that always returns false.

### The alternative, and why it was rejected

Reusing `agent` with a note reading "written by Jev" fails on a structural point: the note is prose, `HasAgentFindingForPair` reads the constant, and no reader of the predicate would ever see the note. A distinction that only exists in a free-text field is not a distinction the code makes.

## §3. Why findings stay inert

The `BacklogFinding` doc comment records that a finding is a record and nothing else, and that this is a *structural* property rather than a convention: folding, reordering, dropping, and editing a card in response to a finding would each require code that does not exist, so none of them can happen by accident.

Consumer C inherits that property for free. The design obligation is only to not be the thing that breaks it — which is why REQ-JEVN-007 states it explicitly and AC-JEVN-005 verifies it against the queue file's hash rather than against a code review.

## §4. Routing — the classification, not the decision

Consumer A asks three questions over one state: a Choice naming the decision owner, and two Nouls (needs-a-measurement-first, cheap-to-reverse). Batched into one request, both for cost — input tokens are charged, output tokens are not, so N requests over one state pay the state N times — and for consistency, since separate requests over identical input could return mutually inconsistent judgments.

Two design points carry the safety:

**Uncertainty escalates, never downgrades.** An `operator` answer routes to operator; so does a Choice confidence below the fitted threshold. The two triggers are separate sub-cases in AC-JEVN-007 because they fail differently — one is the model being confident about a hard case, the other is the model not knowing.

**The no-match option.** Every Choice carries one. A forced choice over an option set that excludes the true answer produces a confident wrong answer, and this repository has recorded that failure from its own dispatch practice independently of any model: a question offering two branches got a confident answer when the true answer was a third thing.

## §5. Skill suggestion — two requests, and the one that matters

Consumer B issues a wide rank over all skills batched with a Noul asking whether the turn needs a skill at all, then reranks the top three under fuller text.

The Noul is the load-bearing half, not the ranking. A ranker over a fixed skill set always returns a best skill, including for turns that need none — and a presented "best" is read as a recommendation regardless of how weak it was. Suppressing the list when the Noul answers negatively above threshold (REQ-JEVN-014) is what stops the consumer manufacturing suggestions out of turns that had no use for one.

The two-request shape exists because the two stages want different state sizes: the wide rank needs every skill's short description and would blow the state budget with full text, while the rerank needs fuller text on three candidates and can afford it. Since accuracy falls as irrelevant state grows, one large request over everything would be worse than two targeted ones.

The intent router's authority is untouched. AC-JEVN-011 verifies this by comparing the router's selection against its selection with the capability disabled — the strongest available form of "it changed nothing".

## §6. Why each consumer is independently gated

The three could plausibly share one gate. They do not, for the reason the whole chain exists: the rejected premise-death task demonstrated that this model's usefulness is per-task, not global. It scored 58.9% where a constant answer scored 75.0% — on one task. That says nothing about near-duplicate marking, and near-duplicate marking's result will say nothing about routing.

A shared gate would also create a failure mode with no good resolution: two consumers pass, one fails, and the shipped set is now a judgement call made under pressure. Independent gates make the answer mechanical — the failing consumer does not ship, the other two do, and the absence is recorded as a decision.
