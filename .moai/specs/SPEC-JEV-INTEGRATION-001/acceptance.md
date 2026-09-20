# SPEC-JEV-INTEGRATION-001 — Acceptance Criteria

Each entry is binary-testable. Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

---

## §A. Display-only invariant

**AC-JEV-001** — Given the backlog queue file with a recorded SHA-256, When a Jev-consulting card admission runs to completion with the capability enabled and a credential present, Then the queue file's SHA-256 after the run differs from the before value only by the appended finding, and re-running the same admission with an identical answer produces no further change. Method: the `sha256.Sum256` before/after comparison already used in `internal/cli/todo_triage_test.go`.

**AC-JEV-002** — Given a card, When Jev returns any answer, Then no field of any `BacklogItem` differs from its pre-call value.

**AC-JEV-003** — Given the capability enabled, When a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication is reached, Then no call to `internal/jev` occurs. Method: a guard test asserting the call path is unreachable from those code paths, plus a grep over those packages for the client symbol with a positive control on a path that does call it.

**AC-JEV-004** — Given output containing a Jev answer, When a reader inspects it, Then the answer is labelled as a model-produced signal and is textually distinguishable from both a mechanical measurement and an agent judgement.

## §B. Default-off and fail-open

**AC-JEV-005** — Given a freshly initialized project from the shipped template, When `workflow.jev.enabled` is read, Then it is `false`, and the compiled default in `internal/config/defaults.go` agrees.

**AC-JEV-006** — Given `workflow.jev.enabled: false`, When each affected command runs, Then its stdout and stderr are byte-identical to the pre-SPEC baseline and no HTTP request to the TypeSafe host is constructed. Method: a transport stub asserting zero calls, plus a golden-output comparison.

**AC-JEV-007** — Given the capability enabled and `~/.moai/.env.typesafe` absent, When each affected command runs, Then it exits 0 and emits at most one notice line.

**AC-JEV-008** — Given the capability enabled, When the transport returns 401, 429, or 529, or the host is unreachable, Then the command exits 0, emits at most one notice line, and the typed unavailable result names the observed condition. One sub-case per condition.

**AC-JEV-009** — Given a call that failed ambiguously, When the client's retry policy is exercised, Then no retry is issued for a non-idempotent call, and at most the configured number for an idempotent one.

## §C. Credential handling

**AC-JEV-010** — Given a fresh credential write, When the file mode of `~/.moai/.env.typesafe` is read, Then it is 0600.

**AC-JEV-011** — Given a pre-existing credential file at mode 0644, When a new credential is written, Then the resulting mode is 0600.

**AC-JEV-012** — Given `settings.AllFields()`, When it is enumerated, Then no entry names the Jev credential. Method: a regression test in the shape of `TestGLMKeyField_AbsentFromSchema`.

**AC-JEV-013** — Given a stored credential longer than four characters, When the web view model is computed, Then `Configured` is true and the hint is exactly the final four characters.

**AC-JEV-014** — Given a stored credential of four characters or fewer, When the web view model is computed, Then `Configured` is true and the hint is empty.

**AC-JEV-015** — Given a request payload containing a credential-shaped token, When the send is attempted, Then the client refuses to send and reports the refusal. A negative control asserts an ordinary payload sends.

## §D. Opt-in surfaces

**AC-JEV-016** — Given the `moai init` wizard, When its question set is enumerated, Then exactly one question carries the Jev id, and `translations` carries an entry for it under each of `ko`, `en`, `ja`, `zh`.

**AC-JEV-017** — Given the `moai web` settings screen, When it renders, Then exactly one Jev section is present carrying an enable toggle and a credential field.

**AC-JEV-018** — Given a Jev setting changed through the wizard and the same setting changed through the console, When the write path is traced, Then both reach the same `internal/settings` seam and no second writer exists. Method: a test asserting both entrances call the shared function.

**AC-JEV-019** — Given a Jev configuration write, When the resulting `workflow.yaml` is compared to its pre-write content, Then every field other than the targeted Jev field is byte-identical.

**AC-JEV-020** — Given either entrance, When its Jev text is read in each of the four locales, Then it states that enabling sends card or request text to a third-party server.

**AC-JEV-021** — Given `moai doctor`, When the Jev check runs, Then it reports enabled-state, credential presence, and endpoint reachability, and sends no judgment request. Method: a transport stub asserting zero judgment calls.

**AC-JEV-022** — Given the console's route table, When it is enumerated, Then no init route is present — the count and membership match the pre-SPEC set.

## §E. Measurement gate

**AC-JEV-023** — Given each of the three consumers, When its measurement artifact is read, Then it names the pinned model id, both language arms with per-arm results and the delta, and the constant-answer baseline for that question shape.

**AC-JEV-024** — Given a measurement artifact, When it is searched for `jev-latest`, Then there are zero matches, with a positive control confirming the search fires on the pinned id.

**AC-JEV-025** — Given a consumer whose measured accuracy does not exceed its constant-answer baseline, When the shipped build is inspected, Then that consumer's call path is absent, and the SPEC's record states it was withheld on measurement.

**AC-JEV-026** — Given each consumer's fitted threshold, When its provenance is read, Then it cites this repository's measured data, and it is not equal-by-copy to a vendor documentation figure nor to another consumer's threshold of a different question shape.

**AC-JEV-027** — Given a measurement result reporting an absence, When its evidence is read, Then a positive control accompanies it showing the apparatus fires on that path.

**AC-JEV-028** — Given the card schema and the card issuance path, When compared before and after this SPEC, Then they are unchanged.

## §F. Question design

**AC-JEV-029** — Given any constructed request, When its questions are inspected, Then none asks the model to count, to do arithmetic, to compare numeric proximity, to order dates, or to compare SHAs; each such value appears instead as a named JSON field in the state.

**AC-JEV-030** — Given any Choice question, When its option set is enumerated, Then a no-match option is present.

**AC-JEV-031** — Given any constructed request, When its size is computed, Then state-plus-longest-question is at or below 32k tokens and the whole request at or below 64k; a request exceeding either is refused rather than truncated.

**AC-JEV-032** — Given several independent questions over one state, When they are issued, Then exactly one HTTP request is made.

**AC-JEV-033** — Given consumer logic reading a Noul, When both polarities matter, Then both probabilities are read; no code derives one as `1 − other`.

## §G. Consumer C — near-duplicate marking

**AC-JEV-034** — Given a pair of cards whose only finding is Jev-sourced, When `HasAgentFindingForPair` is called for that pair, Then it returns false.

**AC-JEV-035** — Given a Jev finding, When its `Source` is read, Then it equals the third constant and equals neither `mechanical` nor `agent`.

**AC-JEV-036** — Given a pair already carrying a finding of any source with the same relation, When a Jev finding for that pair is appended, Then the documented precedence rule is applied and the resulting finding list matches the rule's stated outcome. One sub-case per (existing-source, arriving-source) combination.

**AC-JEV-037** — Given a Jev finding rendered by `todoFindingLine`, When the line is read, Then its probability is rendered in a form distinct from the mechanical `score N.NN` form, and the `machine-only` mark's meaning is unchanged for every pre-existing case.

**AC-JEV-038** — Given a recorded Jev finding, When every card in the queue is compared to its pre-finding state, Then no card field differs.

## §H. Consumer A — lane-question routing

**AC-JEV-039** — Given a lane's blocking question, When routing runs, Then exactly one request is issued carrying one Choice (decision owner, with a no-match option) and two Nouls (needs-a-measurement, cheap-to-reverse).

**AC-JEV-040** — Given a routing answer of `operator`, or a Choice confidence below the fitted threshold, When the item is disposed, Then it is treated as operator-owned.

**AC-JEV-041** — Given any routing answer, When the surrounding state is inspected, Then nothing was dispatched, answered, or closed as a consequence of it.

## §I. Consumer B — skill suggestion

**AC-JEV-042** — Given a turn being routed, When suggestion runs, Then exactly two requests are issued: a wide rank batched with the needs-a-skill Noul, then a top-3 rerank.

**AC-JEV-043** — Given a needs-a-skill Noul answering negatively above the fitted threshold, When the suggestion is presented, Then no ranked list is shown.

**AC-JEV-044** — Given any suggestion, When the `/moai` intent router's selection is compared to its pre-SPEC behaviour for the same input with the capability disabled, Then the router's authority and its selection path are unchanged.

## §J. `/moai goal --auto` seats

**AC-JEV-045** — Given a lane blocking with an operator-owned question inside the sealed-scope loop, When routing classifies it, Then the result is a persisted blocked result with the same shape as the pre-SPEC behaviour, and the loop stops without effects.

**AC-JEV-046** — Given a lane blocking with a lead-owned, cheap-to-reverse question, When routing classifies it, Then the lead answers it and the loop continues, and no `AskUserQuestion` is issued inside the loop. Method: a guard asserting zero user-question emissions between approval and completion.

**AC-JEV-047** — Given a governance receipt written with a Jev Noul supplied, When the receipt is parsed, Then the Jev answer appears as a separate item, every pre-existing binding field is present and unchanged, and the file mode is 0600.

**AC-JEV-048** — Given `mission-governor`'s definition and output shape, When compared before and after this SPEC, Then its tool set, scope boundary, and decision-object shape are unchanged.

**AC-JEV-049** — Given a completion receipt, When its completion predicates and its landed-ancestry and authoritative-readback evidence are enumerated, Then no element derives from a Jev answer.

## §K. Surfaces, template, and the preserved negative result

**AC-JEV-050** — Given the MCP tool wrapper, When its implementation is read, Then it delegates to `internal/jev` and constructs no HTTP request of its own.

**AC-JEV-051** — Given both copies of `moai-mcp-tools.md`, When both tool-count figures in each are read, Then all four agree with the shipped tool set, and the two files remain byte-identical.

**AC-JEV-052** — Given the reference skill, When it is read, Then it carries question-design rules and no call-path instruction.

**AC-JEV-053** — Given every new or changed file under `.claude/` or `.moai/`, When the template tree is checked, Then a corresponding file exists under `internal/template/templates/`.

**AC-JEV-054** — Given every template file this SPEC touches, When it is scanned, Then it contains no SPEC ID, card id, date, price, measurement figure, or local-only-file reference. Method: the existing template-neutrality CI guard, with a positive control.

**AC-JEV-055** — Given a change to an agent definition or a command source under the template tree, When the build runs, Then `make agents-emit-check` or `make commands-emit-check` respectively reports no drift.

**AC-JEV-056** — Given the shipped record, When the premise-death use is looked up, Then the rejection is present with its measured figures (29.2% dead-call precision against a 25% base rate; 58.9% 2-class against a 75.0% constant; 67.5% English control), and `moai todo triage` contains no model call. Method: a grep over the triage path for the client symbol, with a positive control.

---

## §L. Quality gates

- Affected-package tests pass: `go test ./internal/jev/... ./internal/kanban/... ./internal/cli/... ./internal/web/... ./internal/config/...`. Full-suite verdict from CI.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds; `make agents-emit-check` and `make commands-emit-check` report no drift.
- Coverage on `internal/jev` at or above the project's 85% package-level target.
- Template-neutrality CI guard green on `internal/template/templates/**`.

## §M. Definition of Done

1. Every AC above passes, or is recorded as withheld with the measurement that withheld it (AC-JEV-025).
2. Each shipped consumer has a committed measurement citing the pinned model id and both language arms.
3. Each unshipped consumer's absence is recorded as a decision, not an omission.
4. The default-off byte-identity check (AC-JEV-006) passes.
5. The queue-hash check (AC-JEV-001) and the `HasAgentFindingForPair` check (AC-JEV-034) pass.
6. The credential anti-leak check (AC-JEV-012) passes.
7. Both `moai-mcp-tools.md` copies agree with the shipped tool set and with each other.
