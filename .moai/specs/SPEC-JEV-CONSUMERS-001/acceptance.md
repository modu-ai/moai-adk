# SPEC-JEV-CONSUMERS-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here. Each criterion names the requirement ids it covers, so the REQ↔AC mapping does not rest on section-order correspondence.

> **Numbering note.** `AC-JEVN-012`…`015` were added by revision 0.2.0 and are placed in the section each belongs to, so the ids are not monotonic in document order. The set is `001`…`015`, complete, with no gaps and no duplicates.

## §A. Consumer C — near-duplicate marking

**AC-JEVN-001** *(covers `REQ-JEVN-004`, and the predicate half of `REQ-JEVN-003`)* — Given a pair of cards whose only finding is Jev-sourced, When `HasAgentFindingForPair` is called for that pair, Then it returns false. A positive control asserts it returns true for a pair carrying an agent-sourced finding.

**AC-JEVN-002** *(covers `REQ-JEVN-002`)* — Given a Jev finding, When its `Source` is read, Then it equals the third constant and equals neither `mechanical` nor `agent`. This criterion verifies the constant only; the production write path is `AC-JEVN-012`'s subject.

**AC-JEVN-003** *(covers `REQ-JEVN-006` halves (a) and (b))* — Given a pair already carrying a finding, When a further finding naming the same unordered pair with the same relation is appended, Then the settled precedence rule in `spec.md` `REQ-JEVN-006` is applied and the resulting finding list matches that rule's stated outcome. **All four (existing-source, arriving-source) combinations are asserted, in both arrival directions:**

| # | Existing | Arriving | Expected outcome | Rule half |
|---|---|---|---|---|
| 1 | mechanical | jev | the Jev finding is **not** appended; the list is unchanged | (a) |
| 2 | agent | jev | the Jev finding is **not** appended; the list is unchanged | (a) |
| 3 | jev | mechanical | the mechanical finding **is** appended; the Jev finding survives alongside it | (b) |
| 4 | jev | agent | the agent finding **is** appended; the Jev finding survives alongside it | (b) |

Sub-cases 3 and 4 are the half whose failure is silent — a wrongly-suppressed measurement or judgement leaves no trace — so each asserts both that the arriving finding is present and that the pre-existing Jev finding is still present. Method: append against a `BacklogRecord` fixture and compare the resulting `Findings` slice length and contents; sub-cases 1 and 2 additionally assert the append call's own return value reports no append.

**AC-JEVN-004** *(covers `REQ-JEVN-005`)* — Given a Jev finding rendered by `todoFindingLine` (`internal/cli/todo_analysis.go:215`), When the line is read, Then its probability is rendered in a form distinct from the mechanical `score N.NN` form, and the `machine-only` mark's meaning is unchanged for every pre-existing case. Method: a `todoFindingLine` unit test over one finding of each source, plus a regression assertion over the pre-existing mechanical and agent cases.

**AC-JEVN-005** *(covers `REQ-JEVN-007`)* — Given the backlog queue file with a recorded SHA-256 and a recorded Jev finding, When every card in the queue is compared to its pre-finding state, Then no card field differs, and the queue file differs from its before state only by the appended finding. Method: the `sha256.Sum256` comparison already used in `internal/cli/todo_triage_test.go` (import at `:21`, call at `:649`).

**AC-JEVN-012** *(covers the write-path half of `REQ-JEVN-003`)* — Given the packages that produce Jev findings, When they are searched for `BacklogSourceAgent`, Then there are zero matches on any Jev production path. **A positive control accompanies the absence**: the same search over `internal/cli/todo_relate.go` returns its known occurrence at `:76`, confirming the search fires. Method modelled on `SPEC-JEV-CORE-001` `AC-JEVC-003` — a grep over the producing packages for the symbol, with a positive control on a path that does carry it. An absence assertion without the control is a gap, not a finding, and does not satisfy this criterion.

**AC-JEVN-013** *(covers the admission-only scope clause of `REQ-JEVN-001`)* — Given the capability enabled and a recording client stub that counts requests, When `analyzeQueue` (`internal/cli/todo_analysis.go:140`) runs the `moai todo analyze` re-sweep over a queue containing candidate pairs, Then the stub records zero Jev requests. A positive control runs the admission path (`appendAnalyzedCard`, `:38`) against the same stub and records at least one, confirming the stub counts.

## §B. Consumer A — lane-question routing

**AC-JEVN-006** *(covers `REQ-JEVN-008`, `REQ-JEVN-009`)* — Given a lane's blocking question, When routing runs, Then exactly one request is issued carrying one Choice (decision owner, with a no-match option) and two Nouls (needs-a-measurement, cheap-to-reverse). Method: a recording transport stub that counts requests and captures each request's question set.

**AC-JEVN-007** *(covers `REQ-JEVN-011`)* — **Gated on the fitted threshold's existence.** Given this consumer's fitted threshold exists and satisfies `SPEC-JEV-OPTIN-MEASURE-001` `AC-JEVO-013`, When the item is disposed, Then it is treated as operator-owned — the routing surface labels the item `operator` and the consumer offers no other disposition for it. One sub-case per trigger: (i) a routing answer of `operator`, and (ii) a Choice confidence below the fitted threshold. Method: a stubbed answer at each trigger, asserting the emitted label.

**Disposition when no fitted threshold exists**: sub-case (ii) is **not evaluable** and this criterion is recorded as **blocked**, not failed and not silently passed. Sub-case (i) remains evaluable independently of any threshold and is asserted regardless. A run in which no threshold exists disposes this consumer under `REQ-JEVN-015` and reports AC-JEVN-007 as blocked-on-threshold in the E1 matrix, naming the missing threshold.

**AC-JEVN-008** *(covers `REQ-JEVN-010`)* — **Precondition: the consumer's call path is present in the build under test and produced a routing answer during the run.** Given that precondition holds and a routing answer was produced, When the surrounding state is inspected, Then nothing was dispatched, answered, or closed as a consequence of it. Method: compare the surrounding state against the same run with the capability disabled, and assert the recording stub logged at least one answer — without that assertion the comparison is satisfied by the consumer's absence and proves nothing.

## §C. Consumer B — skill suggestion

**AC-JEVN-009** *(covers `REQ-JEVN-012`)* — Given a turn being routed, When suggestion runs, Then exactly two requests are issued: a wide rank batched with the needs-a-skill Noul, then a top-3 rerank. Method: a recording transport stub that counts requests and captures each request's shape.

**AC-JEVN-010** *(covers `REQ-JEVN-014`)* — **Gated on the fitted threshold's existence.** Given this consumer's fitted threshold exists and satisfies `AC-JEVO-013`, and a needs-a-skill Noul answering negatively above that threshold, When the suggestion is presented, Then no ranked list is shown. Method: a stubbed negative Noul answer above threshold, asserting the presented output carries no ranked list.

**Disposition when no fitted threshold exists**: this criterion is **not evaluable** and is recorded as **blocked**, not failed and not silently passed. A run in which no threshold exists disposes this consumer under `REQ-JEVN-015` and reports AC-JEVN-010 as blocked-on-threshold in the E1 matrix, naming the missing threshold.

**AC-JEVN-011** *(covers the authority half of `REQ-JEVN-013`)* — **Precondition: the consumer's call path is present in the build under test and produced a suggestion during the run.** Given that precondition holds, When the `/moai` intent router's selection is compared to its selection for the same input with the capability disabled, Then the selection and the selection path are identical. Method: paired capability-enabled / capability-disabled runs over the same input, plus an assertion that the recording stub logged at least one suggestion in the enabled run — without it, enabled and disabled behaviour are identical by construction and the comparison asserts nothing.

**AC-JEVN-014** *(covers the presentation half of `REQ-JEVN-013`)* — Given a suggestion that was not suppressed, When the object handed to the orchestrator is inspected, Then it is a ranked signal carrying each candidate and its rank position, and it carries no selection, no dispatch, and no instruction to act. Method: inspect the presented value's shape against the rank-signal contract; a positive control asserts the same inspection rejects a value carrying a selection field.

## §D. Cross-consumer — disposition when the gate cannot be run

**AC-JEVN-015** *(covers `REQ-JEVN-015`)* — Given a consumer whose measurement gate could not be run, When the run-phase output for that consumer is read, Then a recorded decision exists that (i) names the consumer, (ii) states that the gate could not be run and why, (iii) cites the withholding mechanism (`Report.Verdict()`, `internal/jevmeasure/measure.go:193`, returning `VerdictWithhold` for any `Source` other than `SourceLive`), and (iv) cites **no** measurement artifact — because in this state none exists.

**The two states are distinguished, and the distinction is the criterion's point.** The record states which of the following applies, and the two are never interchangeable:

| State | What the record cites | What it must not do |
|---|---|---|
| gate **run and failed** | the measurement artifact, its accuracy, and its constant-answer baseline | omit the artifact |
| gate **could not be run** | the reason the gate is un-runnable and the withholding mechanism | cite a measurement that does not exist |
| gate **unrun** (`AC-JEVN-016`) | what the gate still needs in order to run, and who owns it | cite a measurement, or claim the gate was withheld |

A record that describes a gate-not-run consumer as "withheld on measurement" fails this criterion: it claims an observation that was never made. A record that describes an **unrun** gate as **un-runnable** fails it for the mirror reason: it claims an impossibility that was never established, and the two call for different follow-up — an un-runnable gate is closed, an unrun one is owed.

**AC-JEVN-016** *(covers `REQ-JEVN-016`)* — **[NEW 2026-09-21 — v0.3.0]** Given a consumer whose measurement gate is runnable but has not yet been run, and whose implementation is present in the tree, When the shipped default and the run-phase record are inspected, Then all five hold:

1. `internal/config/defaults.go` sets `Jev.Enabled` to `false`, and no shipped template under `internal/template/templates/` sets `enabled: true` under a `jev:` key. Method: read the compiled default; `grep -rn 'jev:' -A 2 internal/template/templates/` and assert every `enabled:` in that block reads `false`. **Positive control**: the same grep over the local (non-template) `.moai/config/sections/workflow.yaml` resolves the same key, confirming the pattern matches a `jev:` block at all.
2. With the gate off, the consumer constructs no request. Method: a recording client stub counting requests, asserting zero across the consumer's entry path with `enabled: false`. **Positive control**: the same stub with `enabled: true` records at least one — without it, zero is satisfied equally by a stub that counts nothing.
3. With the gate off, the surrounding output is identical to its pre-Jev behaviour apart from at most one notice line. Method: byte-compare the command output against the same run with the consumer's call site absent.
4. No documentation, release note, or CHANGELOG entry presents the consumer as available. Method: a grep over the published surfaces for the consumer's user-facing name, expecting zero; **positive control** on a surface that does name a shipped capability, confirming the search fires.
5. The run-phase record names the state as **gate-unrun**, names what the gate still needs and who owns it, and cites no measurement artifact.

Failing any of the five means the consumer is not in the gate-unrun state this criterion describes; it is either reachable at the default (and therefore shipped without its gate, which `REQ-JEVO-009` forbids) or un-runnable (and therefore governed by `AC-JEVN-015` instead).

---

## §E. Quality gates

- `go test ./internal/kanban/... ./internal/cli/...` passes; full-suite verdict from CI. This gate covers Consumer C's packages. It does **not** reach Consumer A's or Consumer B's host surfaces, which are unresolved (`plan.md` §F, N2); a verification surface for M5 and M6 is part of answering N2, not an assumption this gate may make.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds.
- Each **shipped** consumer has a committed measurement artifact satisfying the gate in `SPEC-JEV-OPTIN-MEASURE-001` (`AC-JEVO-010` through `AC-JEVO-014`). A consumer disposed under `REQ-JEVN-015` or `REQ-JEVN-016` has no such artifact by construction and is covered by `AC-JEVN-015` / `AC-JEVN-016` instead. **"Shipped" here means reachable at the shipped default**, per the reading stated and defended in `spec.md` `REQ-JEVN-016` — a consumer present in the tree but off by default is not shipped, and is covered by `AC-JEVN-016`.

## §F. Definition of Done

1. **Per consumer, scoped to that consumer's own criteria.** For each consumer that ships, every AC in that consumer's own section passes — §A for Consumer C, §B for Consumer A, §C for Consumer B — plus §D `AC-JEVN-015` where that consumer did not ship. A build shipping only Consumer C is **not** required to pass §B or §C; the three consumers must be able to fail independently, and a Definition of Done that couples them undoes the independence `plan.md` §C2 / §D and `design.md` §6 establish.
2. Each consumer that ships has a committed measurement citing the pinned model id and both language arms, beating its constant-answer baseline.
3. Each consumer that does NOT ship has its absence recorded as a decision, in exactly one of **three** forms: **gate run and failed** — the decision cites the measurement that withheld it (`AC-JEVN-015`); **gate could not be run** — the decision cites the un-runnable gate and the withholding mechanism, and cites no measurement (`AC-JEVN-015`); **gate unrun** — the decision names what the gate still needs and who owns it, cites no measurement, and the consumer's code, if present, satisfies all five conditions of `AC-JEVN-016`.
4. For Consumer C specifically: the `HasAgentFindingForPair` check (`AC-JEVN-001`), the write-path absence check with its positive control (`AC-JEVN-012`), the four-combination precedence check (`AC-JEVN-003`), and the queue-hash check (`AC-JEVN-005`) all pass.
5. The seven tests that enumerate the finding source set compile and pass after the third constant is added: `internal/kanban/backlog_findings_test.go`, `backlog_archive_test.go`, `todo_merge_procedure_test.go`, `todo_queue_merge_test.go`, `internal/cli/todo_analysis_test.go`, `todo_analysis_add_test.go`, `todo_relate_test.go`.
6. Any milestone declared blocked in `plan.md` §F (currently M5 and M6, on N2) is either unblocked by an answer recorded at the Implementation Kickoff Approval gate, or remains out of the shipped set with its block recorded. A blocked milestone is not the same state as a gate-not-run consumer and is recorded separately.
