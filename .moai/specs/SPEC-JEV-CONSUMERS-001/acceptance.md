# SPEC-JEV-CONSUMERS-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

## §A. Consumer C — near-duplicate marking

**AC-JEVN-001** — Given a pair of cards whose only finding is Jev-sourced, When `HasAgentFindingForPair` is called for that pair, Then it returns false. A positive control asserts it returns true for a pair carrying an agent-sourced finding.

**AC-JEVN-002** — Given a Jev finding, When its `Source` is read, Then it equals the third constant and equals neither `mechanical` nor `agent`.

**AC-JEVN-003** — Given a pair already carrying a finding of any source with the same relation, When a Jev finding for that pair is appended, Then the documented precedence rule is applied and the resulting finding list matches the rule's stated outcome. One sub-case per (existing-source, arriving-source) combination.

**AC-JEVN-004** — Given a Jev finding rendered by `todoFindingLine`, When the line is read, Then its probability is rendered in a form distinct from the mechanical `score N.NN` form, and the `machine-only` mark's meaning is unchanged for every pre-existing case.

**AC-JEVN-005** — Given the backlog queue file with a recorded SHA-256 and a recorded Jev finding, When every card in the queue is compared to its pre-finding state, Then no card field differs, and the queue file differs from its before state only by the appended finding. Method: the `sha256.Sum256` comparison already used in `internal/cli/todo_triage_test.go`.

## §B. Consumer A — lane-question routing

**AC-JEVN-006** — Given a lane's blocking question, When routing runs, Then exactly one request is issued carrying one Choice (decision owner, with a no-match option) and two Nouls (needs-a-measurement, cheap-to-reverse).

**AC-JEVN-007** — Given a routing answer of `operator`, or a Choice confidence below the fitted threshold, When the item is disposed, Then it is treated as operator-owned. One sub-case per trigger.

**AC-JEVN-008** — Given any routing answer, When the surrounding state is inspected, Then nothing was dispatched, answered, or closed as a consequence of it.

## §C. Consumer B — skill suggestion

**AC-JEVN-009** — Given a turn being routed, When suggestion runs, Then exactly two requests are issued: a wide rank batched with the needs-a-skill Noul, then a top-3 rerank.

**AC-JEVN-010** — Given a needs-a-skill Noul answering negatively above the fitted threshold, When the suggestion is presented, Then no ranked list is shown.

**AC-JEVN-011** — Given any suggestion, When the `/moai` intent router's selection is compared to its selection for the same input with the capability disabled, Then the selection and the selection path are identical.

---

## §D. Quality gates

- `go test ./internal/kanban/... ./internal/cli/...` passes; full-suite verdict from CI.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds.
- Each shipped consumer has a committed measurement artifact satisfying the gate in `SPEC-JEV-OPTIN-MEASURE-001` (AC-JEVO-010 through AC-JEVO-014).

## §E. Definition of Done

1. Every AC above passes for each shipped consumer.
2. Each consumer that ships has a committed measurement citing the pinned model id and both language arms, beating its constant-answer baseline.
3. Each consumer that does NOT ship has its absence recorded as a decision with the measurement that withheld it.
4. The `HasAgentFindingForPair` check (AC-JEVN-001) and the queue-hash check (AC-JEVN-005) pass.
5. The seven tests that enumerate the finding source set compile and pass after the third constant is added: `internal/kanban/backlog_findings_test.go`, `backlog_archive_test.go`, `todo_merge_procedure_test.go`, `todo_queue_merge_test.go`, `internal/cli/todo_analysis_test.go`, `todo_analysis_add_test.go`, `todo_relate_test.go`.
