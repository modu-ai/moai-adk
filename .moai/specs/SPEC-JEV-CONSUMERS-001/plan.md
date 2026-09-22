# SPEC-JEV-CONSUMERS-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B carries the data-model decision and the settled precedence rule that governs what M4 builds; §F carries the one live open question (N2) that blocks M5 and M6; mechanical work is last.

---

## §A. Context

Third SPEC of a four-part chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Depends on `SPEC-JEV-OPTIN-MEASURE-001` for the measurement gate, and transitively on `SPEC-JEV-CORE-001` for the call path.

Verified tree facts (re-measured in this worktree at HEAD `7e1ed63b9`; the previous revision cited these against `fd75cf692`, and three of its line numbers were wrong at both trees — corrected below):

| Fact | Source | Observed |
|---|---|---|
| Two source constants exist | `internal/kanban/backlog_store.go:118, :120` | `mechanical`, `agent` |
| `Source` has no validator | grep over write paths | free string; `todo_relate.go:76` hardcodes `agent`; `todo_analysis.go:64`, `:73`, and `:179` hardcode `mechanical` — the set is closed by its write paths, not by a check |
| Three mechanical write sites, not two | `internal/cli/todo_analysis.go:64, :73, :179` | `:64` and `:73` are in `appendAnalyzedCard` (declared `:38`, the admission path); `:179` is in `analyzeQueue` (declared `:140`, the `moai todo analyze` re-sweep) |
| The mark's predicate | `internal/kanban/backlog_store.go:409` | `HasAgentFindingForPair` selects on `Source == BacklogSourceAgent`; its doc comment states the mark records the absence of an agent-sourced record, "never that any review took place" |
| The dedup key is source-inclusive | `internal/kanban/backlog_store.go:357`, `:369` | `AppendFindingOnce` delegates solely to `HasFindingTuple`, keyed `{SubjectID, RelatedID, Relation, Source}` — ordered and source-inclusive |
| `SamePairAs` is not on the append path | `internal/kanban/backlog_store.go:159`, `:409` | one non-test caller in the repository: `HasAgentFindingForPair` |
| Render is source-conditioned | `internal/cli/todo_analysis.go:215` `todoFindingLine` | `machine-only` mark and `score N.NN` both render only for `mechanical` |
| The gate withholds on any non-live source | `internal/jevmeasure/measure.go:193` | `Report.Verdict()` first-checks `r.Source != SourceLive` and returns `VerdictWithhold` |
| Findings are inert by construction | `BacklogFinding` doc comment | no code path writes a card field as a consequence of a finding; folding/reordering/dropping would each need code that does not exist |

> **Coordinate correction, stated as a miscitation rather than drift.** The 0.1.0 revision cited `todo_relate.go:75` and `todo_analysis.go:62,71`. Those coordinates did not hold at HEAD `7e1ed63b9` and did not hold at the declared baseline `fd75cf692` either — nothing moved between the two trees. They were never correct, and the record says so rather than describing a change that did not occur.

## §B. Data-model decision and the settled precedence rule (review this first)

**B1 — A third finding source constant: `BacklogSourceJev = "jev"`.**

`BacklogFinding.Source` is a free string, but the two existing constants are **not interchangeable**. They partition by *who observed* the relation, and `HasAgentFindingForPair` reads that partition to drive the `machine-only` mark. Filing a model answer under either existing constant makes the queue assert something false: under `mechanical` it renders with a `score` that reads as a measured similarity; under `agent` it clears the `machine-only` mark from pairs nobody reviewed, claiming a review that did not happen.

Two consequences follow. The first is now a **settled decision**, not a proposal.

- **[DECIDED 2026-09-20 — N1] Dedup precedence.** Confirmed by the operator exactly as proposed. A Jev finding is appended **only when no finding of any source** already names that unordered pair with the same relation; a later mechanical or agent finding for the same pair is appended **alongside** it rather than replacing it. The asymmetry is deliberate — a model signal must never suppress a measurement or a judgement, and must never be silently suppressed by one, because both suppressions would make the queue quieter than the evidence warrants. `AC-JEVN-003` asserts all four source combinations.

  **The rule's two halves cost opposite amounts, and the requirement now says so.** Half (b) — a later mechanical or agent finding lands alongside — is the **unchanged default** of `AppendFindingOnce` (`internal/kanban/backlog_store.go:369`): its key includes `Source`, so a mechanical finding never matches an existing Jev tuple and is always appended. It needs **no code**. Half (a) — suppressing an arriving Jev finding on a pair any source already names — is **not implementable through the existing paths**: `AppendFindingOnce` delegates solely to `HasFindingTuple` (`:357`), whose source-inclusive ordered key means a Jev finding is *never* suppressed by a mechanical or agent one; and `AppendFindingOnce` never calls `SamePairAs` (`:159`), whose only non-test caller is `HasAgentFindingForPair` (`:409`). Half (a) therefore requires a **new source-agnostic unordered predicate**, provisionally `HasFindingForPairAnySource`, keyed on the unordered pair plus the relation and not on `Source`. `REQ-JEVN-006` authorises it and M4 budgets it.

- **Render.** The existing rule prints a score only for `mechanical`, on the stated reasoning that an agent judgement carries no measurement and `0.00` would read as measured dissimilarity. A Jev probability is a third thing — a calibrated model confidence — and renders with its own label, so the three are distinguishable at a glance. A third source that inherits the "no score" branch by default would silently drop its probability, which is why `REQ-JEVN-005` names the render as a requirement rather than leaving it to fall out.

**B2 — The `agent` reuse alternative, and why it was rejected.** Reusing `agent` with a note reading "written by Jev" was considered. It fails because the note is prose and `HasAgentFindingForPair` reads the constant: no reader of the predicate would ever see the note. A distinction that exists only in a free-text field is not a distinction the code makes.

**B3 — Consumer C is admission-only.** `REQ-JEVN-001` scopes the trigger to card admission (`appendAnalyzedCard`, `internal/cli/todo_analysis.go:38`) and explicitly excludes the `moai todo analyze` re-sweep (`analyzeQueue`, `:140`). The re-sweep is where precedence half (b) would otherwise be exercised most often — it re-walks pairs that may already carry a Jev finding — so leaving the scope unstated would have left a live behavioural question inside a settled rule. The exclusion is asserted by `AC-JEVN-013`, and the re-sweep's own mechanical write site (`:179`) is not modified.

## §C. Interface notes

**C1 — Write paths to model on.** `internal/cli/todo_analysis.go:58-76` (mechanical, written at admission time, finding names the NEW card as subject) and `internal/cli/todo_relate.go:72-80` (agent, written through an explicit verb with a validated relation). Consumer C follows the first: written at admission, subject is the new card, no verb.

**C2 — The three consumers share nothing but the client.** Each has its own state shape, its own question set, its own threshold, and its own measurement. There is no shared "consumer framework" to build, and building one would couple three things whose only common property is that they call the same package.

**C3 — Routing and suggestion produce values, not actions.** Neither has a write path. The verification for both (`AC-JEVN-008`, `AC-JEVN-011`) is a comparison against the capability-disabled behaviour, which is the strongest available form of "it changed nothing" — **provided the consumer's call path exists and produced an answer in the run under test**. Without that precondition the comparison is satisfied by the absence of the consumer, so both criteria now carry it explicitly.

## §D. Constraints

- **Each consumer is independently gated.** A consumer that does not beat its own constant-answer baseline is not shipped, and its absence is recorded as a decision.
- **A gate that cannot be run is a third state, not a failure.** `REQ-JEVN-015` defines what run-phase produces when no live measurement is possible: a recorded decision citing the un-runnable gate, with no measurement artifact required. `AC-JEVN-015` distinguishes it from gate-run-and-failed.
- **A gate that has not yet been run is a FOURTH state — `gate-unrun` — and it is the one M4 is actually in.** `REQ-JEVN-016` (added 2026-09-21, card t1066) defines it: gate runnable, not yet run, call path **present** in the tree but unreachable at the shipped default. It is distinct from `REQ-JEVN-015`, which requires the call path to be **absent** and is reserved for a gate that *cannot* be run. Admission conditions and the "shipped = reachable at the default" reading of `REQ-JEVO-009` are in `spec.md` `REQ-JEVN-016`; verification is `AC-JEVN-016`.
- **The CORE display-only wording is reconciled, not worked around.** `SPEC-JEV-CORE-001` v0.2.0 (card t1066) carves out the inert sibling record in REQ-JEVC-011 and defines "for" in REQ-JEVC-012. Consumer C's finding append is authorised under that carve-out **by name**, and the carve-out lapses if anything ever selects on a Jev finding. Do NOT read the carve-out as general permission for Consumers A and B.
- **Display-only**, verified against a queue-file hash before and after for Consumer C, and against disabled-behaviour equivalence for A and B.
- **Never `Source: agent`** for a Jev finding — asserted at the predicate level by `AC-JEVN-001` with a positive control, and at the write-path level by `AC-JEVN-012` with a positive control.
- **No card change.** No field, no schema, no issuance path.
- **`moai todo triage` stays model-free.**
- **Verification scope.** Affected packages only; full-suite verdict from CI.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject. Each consumer's measurement artifact is cited by path in E1, with its constant baseline alongside its accuracy — an accuracy without its baseline is not evidence. A consumer disposed under `REQ-JEVN-015` cites its recorded decision instead, and E1 states that no measurement exists for it rather than leaving the row blank.

## §F. Milestones

Each milestone is gated on its own measurement from `SPEC-JEV-OPTIN-MEASURE-001`. A milestone whose gate **fails** ends in a recorded decision, not a partial implementation. A milestone whose gate **cannot be run** ends in the `REQ-JEVN-015` recorded decision — also not a partial implementation, and not a blank. A milestone whose gate is **runnable but not yet run** ends in the `REQ-JEVN-016` **gate-unrun** decision, and MAY carry its implementation in the tree provided all five `AC-JEVN-016` conditions hold — default-off everywhere, behaviourally absent, not presented as available, and the record naming what the gate still needs.

**M4 — Consumer C (near-duplicate marking).** Priority High. **Not blocked.** Deliverables: the `BacklogSourceJev` constant; the write path modelled on `todo_analysis.go:58-76`, wired to the admission path only (`appendAnalyzedCard`, `:38`) per `REQ-JEVN-001`; **the new source-agnostic unordered predicate `HasFindingForPairAnySource`** that `REQ-JEVN-006` half (a) requires, plus its call at the Jev append site — budgeted here because no existing path can express the settled rule; the render form; the `HasAgentFindingForPair`-stays-false assertion; the write-path `Source: agent` absence assertion with its positive control; the queue-hash assertion. Half (b) of the precedence rule is delivered by changing nothing and is verified rather than built. Declaration-only updates in the seven tests that enumerate the source set (listed in `acceptance.md` §F item 5).

**M5 — Consumer A (lane-question routing).** Priority Medium. **BLOCKED on N2.** The code path that hosts routing is unresolved (`research.md` §5), and the operator's 2026-09-20 decisions did not cover it. This revision deliberately does **not** name a host: choosing one that was never measured would be exactly the unmeasured-premise defect this chain is bounded against. M5 does not enter run-phase until N2 is answered at the Implementation Kickoff Approval gate. Scope once unblocked: one batched request — a Choice naming the decision owner (lead / operator / worker, plus a no-match option) and two Nouls; operator-owned or below-threshold routes to operator. `SPEC-JEV-GOAL-DIST-001` seat (i) depends on this milestone.

**M6 — Consumer B (skill suggestion).** Priority Medium. **BLOCKED on N2.** M6's host is the `/moai` intent router, a prose-level skill surface that the §E quality gate (`go test ./internal/kanban/... ./internal/cli/...`) does not reach — so M6 has no identified verification surface, for the same unresolved reason M5 has no host. Naming one here would be unmeasured. Scope once unblocked: two requests — a wide rank over all skills batched with a needs-a-skill-at-all Noul, then a top-3 rerank under fuller text; presented as a ranked signal carrying each candidate's rank position; the intent router keeps its selection authority.

> **The blocked milestones' disposition is not the same as the un-runnable gate's.** M5 and M6 are blocked on an unanswered design question (N2) and are not started. `REQ-JEVN-015` governs a consumer whose work was in scope and whose gate could not be run. A milestone can be in both states; each is recorded separately.

## §G. Anti-patterns

- **Writing a Jev finding as `agent`.** Produces a false review record — the queue claims a review for a pair nobody read. Caught at the predicate level by `AC-JEVN-001` and at the write-path level by `AC-JEVN-012`.
- **Enforcing precedence half (a) through `AppendFindingOnce` or `SamePairAs`.** Neither can express it: the first is keyed source-inclusively, the second is not on the append path. An implementation that "reuses the existing dedup" silently ships a rule that never suppresses anything, and the failure is invisible because an un-suppressed append looks exactly like a correct one.
- **Treating precedence halves (a) and (b) as one mechanism.** (b) is free and (a) is new code; budgeting them together hides the cost of the half that has one.
- **Letting the third source inherit the `mechanical` render branch.** The probability would print as `score N.NN` and read as a measured similarity.
- **Letting it inherit the `agent` branch instead.** The probability silently disappears.
- **Wiring Consumer C into the `moai todo analyze` re-sweep.** Out of scope per `REQ-JEVN-001` / B3; the re-sweep runs over pairs that already carry findings, so a Jev consumer there changes the finding population on a path this SPEC never measured.
- **Building a shared consumer framework.** The three share a client and nothing else; coupling them makes each one's gate harder to fail independently.
- **Making a consumer whose gate has not been run REACHABLE at the shipped default.** The gate is the committed measurement, not the intention to take one. The correct output in that state is the `REQ-JEVN-016` **gate-unrun** decision, with the consumer off by default and behaviourally absent. (Restated 2026-09-21: the earlier wording — "shipping a consumer whose gate has not been run" — was read as forbidding the presence of the code, which `REQ-JEVN-016` now explicitly permits under its five conditions. What is forbidden is reachability, not compilation.)
- **Recording a gate-unrun consumer as gate-not-runnable.** The two states call for different follow-up: an un-runnable gate is closed and cites its withholding mechanism, an unrun one is owed and cites what it still needs. Collapsing them loses the owed work, and nothing later re-raises it.
- **Recording a gate-not-run consumer as "withheld on measurement".** There is no measurement to cite; citing one that does not exist is an unobserved-verification claim. The record cites the un-runnable gate instead.
- **Naming a host path for M5 or M6 in order to unblock them.** N2 is unanswered; a host chosen to make the plan look complete is an unmeasured premise in the place a reader is least likely to check it.
- **Letting the display become the verdict.** A code path that acts on any of the three signals is the defect this whole chain is bounded against.

## §H. Cross-references

- `SPEC-JEV-OPTIN-MEASURE-001` — predecessor; owns the gate each consumer passes, the threshold-provenance rule (`AC-JEVO-013`), and the question-design rules each consumer's questions follow.
- `SPEC-JEV-CORE-001` — owns the call path, the display-only invariant, and the fail-open contract. `AC-JEVC-003` is the precedent `AC-JEVN-012` is modelled on (a grep over the producing packages with a positive control on a path that does carry the symbol).
- `SPEC-JEV-GOAL-DIST-001` — successor; its seat (i) consumes Consumer A's routing, and therefore inherits M5's N2 block.
- `.moai/reports/t1020/plan-audit-SPEC-JEV-CONSUMERS-001.md` — the iteration-1 audit verdict this revision answers.
- `internal/kanban/backlog_store.go:108-170, 350-420` — the finding contract §B1 extends and the dedup paths §B1 measures.
- `internal/cli/todo_analysis.go:38, :140, :215`, `internal/cli/todo_relate.go:72-80` — the write-path precedents, the re-sweep entry point, and the render function.
- `internal/jevmeasure/measure.go:193` — `Report.Verdict()`, the mechanism `REQ-JEVN-015` cites.
