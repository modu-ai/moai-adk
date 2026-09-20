# SPEC-JEV-CONSUMERS-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B carries the data-model decision and the one open question that must be answered before M4 can start; mechanical work is last.

---

## §A. Context

Third SPEC of a four-part chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Depends on `SPEC-JEV-OPTIN-MEASURE-001` for the measurement gate, and transitively on `SPEC-JEV-CORE-001` for the call path.

Verified tree facts (measured in this worktree at `fd75cf692`):

| Fact | Source | Observed |
|---|---|---|
| Two source constants exist | `internal/kanban/backlog_store.go:117-120` | `mechanical`, `agent` |
| `Source` has no validator | grep over write paths | free string; `todo_relate.go:75` hardcodes `agent`, `todo_analysis.go:62,71` hardcode `mechanical` — the set is closed by its write paths, not by a check |
| The mark's predicate | `backlog_store.go:408-416` | `HasAgentFindingForPair` selects on `Source == BacklogSourceAgent`; its doc comment states the mark records the absence of an agent-sourced record, "never that any review took place" |
| Render is source-conditioned | `todo_analysis.go` `todoFindingLine` | `machine-only` mark and `score N.NN` both render only for `mechanical` |
| Findings are inert by construction | `BacklogFinding` doc comment | no code path writes a card field as a consequence of a finding; folding/reordering/dropping would each need code that does not exist |

## §B. Data-model decision and the open question (review this first)

**B1 — A third finding source constant: `BacklogSourceJev = "jev"`.**

`BacklogFinding.Source` is a free string, but the two existing constants are **not interchangeable**. They partition by *who observed* the relation, and `HasAgentFindingForPair` reads that partition to drive the `machine-only` mark. Filing a model answer under either existing constant makes the queue assert something false: under `mechanical` it renders with a `score` that reads as a measured similarity; under `agent` it clears the `machine-only` mark from pairs nobody reviewed, claiming a review that did not happen.

Two consequences follow, and **the first is an open question, not a settled decision**:

- **[OPEN — N1] Dedup precedence.** The proposal carried from the pre-split SPEC: a Jev finding is appended only when no finding of **any** source already names that unordered pair with the same relation, and a later mechanical or agent finding for the same pair is appended alongside it rather than replacing it. The asymmetry is deliberate — a model signal must never suppress a measurement or a judgement, and must never be silently suppressed by one, because both suppressions would make the queue quieter than the evidence warrants. **This remains a proposal.** It needs a decision before M4 implementation; AC-JEVN-003 asserts whatever rule is chosen, one sub-case per source combination.
- **Render.** The existing rule prints a score only for `mechanical`, on the stated reasoning that an agent judgement carries no measurement and `0.00` would read as measured dissimilarity. A Jev probability is a third thing — a calibrated model confidence — and renders with its own label, so the three are distinguishable at a glance. A third source that inherits the "no score" branch by default would silently drop its probability, which is why REQ-JEVN-005 names the render as a requirement rather than leaving it to fall out.

**B2 — The `agent` reuse alternative, and why it was rejected.** Reusing `agent` with a note reading "written by Jev" was considered. It fails because the note is prose and `HasAgentFindingForPair` reads the constant: no reader of the predicate would ever see the note. A distinction that exists only in a free-text field is not a distinction the code makes.

## §C. Interface notes

**C1 — Write paths to model on.** `internal/cli/todo_analysis.go:58-76` (mechanical, written at admission time, finding names the NEW card as subject) and `internal/cli/todo_relate.go:72-80` (agent, written through an explicit verb with a validated relation). Consumer C follows the first: written at admission, subject is the new card, no verb.

**C2 — The three consumers share nothing but the client.** Each has its own state shape, its own question set, its own threshold, and its own measurement. There is no shared "consumer framework" to build, and building one would couple three things whose only common property is that they call the same package.

**C3 — Routing and suggestion produce values, not actions.** Neither has a write path. The verification for both (AC-JEVN-008, AC-JEVN-011) is a comparison against the capability-disabled behaviour, which is the strongest available form of "it changed nothing".

## §D. Constraints

- **Each consumer is independently gated.** A consumer that does not beat its own constant-answer baseline is not shipped, and its absence is recorded as a decision.
- **Display-only**, verified against a queue-file hash before and after for Consumer C, and against disabled-behaviour equivalence for A and B.
- **Never `Source: agent`** for a Jev finding, asserted by AC-JEVN-001 with a positive control.
- **No card change.** No field, no schema, no issuance path.
- **`moai todo triage` stays model-free.**
- **Verification scope.** Affected packages only; full-suite verdict from CI.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject. Each consumer's measurement artifact is cited by path in E1, with its constant baseline alongside its accuracy — an accuracy without its baseline is not evidence.

## §F. Milestones

Each milestone is gated on its own measurement from `SPEC-JEV-OPTIN-MEASURE-001`. A milestone whose gate fails ends in a recorded decision, not a partial implementation.

**M4 — Consumer C (near-duplicate marking).** Priority High. `BacklogSourceJev` constant; the write path modelled on `todo_analysis.go:58-76`; the precedence rule once N1 is decided; the render form; the `HasAgentFindingForPair`-stays-false assertion; the queue-hash assertion. Declaration-only updates in the seven tests that enumerate the source set (listed in `acceptance.md` §E item 5).

**M5 — Consumer A (lane-question routing).** Priority Medium. One batched request: a Choice naming the decision owner (lead / operator / worker, plus a no-match option) and two Nouls. Operator-owned or below-threshold routes to operator. `SPEC-JEV-GOAL-DIST-001` seat (i) depends on this milestone.

**M6 — Consumer B (skill suggestion).** Priority Medium. Two requests: a wide rank over all skills batched with a needs-a-skill-at-all Noul, then a top-3 rerank under fuller text. Presented as a ranked signal; the intent router keeps its selection authority.

## §G. Anti-patterns

- **Writing a Jev finding as `agent`.** Produces a false review record — the queue claims a review for a pair nobody read. Mechanically caught by AC-JEVN-001.
- **Letting the third source inherit the `mechanical` render branch.** The probability would print as `score N.NN` and read as a measured similarity.
- **Letting it inherit the `agent` branch instead.** The probability silently disappears.
- **Implementing M4 before N1 is decided.** The precedence rule is a proposal; implementing it as though settled buries a decision in a diff.
- **Building a shared consumer framework.** The three share a client and nothing else; coupling them makes each one's gate harder to fail independently.
- **Shipping a consumer whose gate has not been run.** The gate is the committed measurement, not the intention to take one.
- **Letting the display become the verdict.** A code path that acts on any of the three signals is the defect this whole chain is bounded against.

## §H. Cross-references

- `SPEC-JEV-OPTIN-MEASURE-001` — predecessor; owns the gate each consumer passes and the question-design rules each consumer's questions follow.
- `SPEC-JEV-CORE-001` — owns the call path, the display-only invariant, and the fail-open contract.
- `SPEC-JEV-GOAL-DIST-001` — successor; its seat (i) consumes Consumer A's routing.
- `internal/kanban/backlog_store.go:108-150, 408-416` — the finding contract §B1 extends.
- `internal/cli/todo_analysis.go:58-76`, `todo_relate.go:72-80` — the two write-path precedents.
