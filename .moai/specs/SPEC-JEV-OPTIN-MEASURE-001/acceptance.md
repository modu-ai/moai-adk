# SPEC-JEV-OPTIN-MEASURE-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

## §A. Opt-in surfaces

**AC-JEVO-001** — Given the `moai web` settings screen, When it renders, Then exactly one Jev section is present carrying an enable toggle and a credential field.

**AC-JEVO-002** — Given a Jev setting changed through the wizard and the same setting changed through the console, When the write path is traced, Then both reach the same `internal/settings` seam and no second writer exists. Method: a test asserting both entrances call the shared function.

**AC-JEVO-003** — Given a Jev configuration write, When the resulting `workflow.yaml` is compared to its pre-write content, Then every field other than the targeted Jev field is byte-identical.

**AC-JEVO-004** — Given either entrance, When its Jev text is read in each of `ko`, `en`, `ja`, and `zh`, Then it states that enabling sends card or request text to a third-party server. One sub-case per locale.

**AC-JEVO-005** — Given the locale resolver `GetLocalizedQuestion`, When the Jev question is resolved through it in each of `ko`, `en`, `ja`, and `zh`, Then every locale yields a non-empty title and description, and each non-English result is not byte-identical to the English base. The `translations` map carries locale keys `ko`, `ja`, and `zh` only — English is the source language, held on the `Question` itself, and the resolver returns before the map is consulted — so the four locales are verified at the render surface rather than by four keys in the map. Method: one assertion per locale through the resolver.

**AC-JEVO-006** — Given the console's route table, When it is enumerated, Then no init route is present — the count and membership match the pre-SPEC set.

## §B. Init-only placement

**AC-JEVO-007** — Given `InitQuestions`, When its question list is enumerated, Then it contains exactly five questions and exactly one of them carries the Jev id.

**AC-JEVO-008** — Given `DefaultQuestions` and `ReconfigureQuestions`, When each is enumerated, Then neither contains a question carrying the Jev id, and `ReconfigureQuestions` has the same membership as before this SPEC.

**AC-JEVO-009** — Given the wizard question text and the `moai init` completion output, When they are read in each of the four locales, Then at least one of the two names `moai web` as the place the setting can later be changed.

## §C. Measurement gate

**AC-JEVO-010** — Given each consumer's measurement artifact, When it is read, Then it names the pinned model id, both language arms with per-arm results and the delta, and the constant-answer baseline for that question shape.

**AC-JEVO-011** — Given a measurement artifact, When it is searched for `jev-latest`, Then there are zero matches, with a positive control confirming the search fires on the pinned id.

**AC-JEVO-012** — Given a consumer whose measured accuracy does not exceed its constant-answer baseline, When the shipped build is inspected, Then that consumer's call path is absent, and the record states it was withheld on measurement.

**AC-JEVO-013** — Given each consumer's fitted threshold, When its provenance is read, Then it cites this repository's measured data, and it is not equal-by-copy to a vendor documentation figure nor to another consumer's threshold of a different question shape.

**AC-JEVO-014** — Given a measurement result reporting an absence, When its evidence is read, Then a positive control accompanies it showing the apparatus fires on that path.

**AC-JEVO-015** — Given the card schema, the card issuance path, and the language of every card in the queue, When compared before and after this SPEC, Then they are unchanged.

## §D. Question design

**AC-JEVO-016** — Given any request the harness constructs, When its questions are inspected, Then none asks the model to count, to do arithmetic, to compare numeric proximity, to order dates, or to compare SHAs; each such value appears instead as a named JSON field in the state.

**AC-JEVO-017** — Given any Choice question, When its option set is enumerated, Then a no-match option is present.

**AC-JEVO-018** — Given any constructed state, When its fields are enumerated, Then every field is read by at least one of that request's questions.

**AC-JEVO-019** — Given a state payload containing imperative text, When the request is processed, Then that text is carried as data and no code path treats it as instruction. Method: a fixture carrying an injected instruction, asserting the answer shape is unchanged.

**AC-JEVO-020** — Given harness logic reading a Noul, When both polarities matter, Then both probabilities are read; no code derives one as `1 − other`.

---

## §E. Quality gates

- `go test ./internal/cli/wizard/... ./internal/web/... ./internal/settings/...` passes; full-suite verdict from CI.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds.
- Template-neutrality CI guard green on `internal/template/templates/**`.

## §F. Definition of Done

1. Every AC above passes.
2. Both entrances demonstrably share one writer (AC-JEVO-002) and leave sibling config fields byte-identical (AC-JEVO-003).
3. The init-only placement and its consequence are asserted mechanically (AC-JEVO-007, AC-JEVO-008) and the escape hatch is discoverable (AC-JEVO-009).
4. The measurement harness produces, for at least one trial consumer, an artifact satisfying AC-JEVO-010 — so the gate is demonstrated before `SPEC-JEV-CONSUMERS-001` depends on it.
