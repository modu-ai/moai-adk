# SPEC-JEV-GOAL-DIST-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

## §A. `/moai goal --auto` seats

**AC-JEVG-001** — Given a lane blocking with an operator-owned question inside the sealed-scope loop, When routing classifies it, Then the result is a persisted blocked result with the same shape as the pre-SPEC behaviour, and the loop stops without effects.

**AC-JEVG-002** — Given a lane blocking with a lead-owned, cheap-to-reverse question, When routing classifies it, Then the lead answers it and the loop continues, and no `AskUserQuestion` is issued inside the loop. Method: a guard asserting zero user-question emissions between approval and completion.

**AC-JEVG-003** — Given a governance receipt written with a Jev Noul supplied, When the receipt is parsed, Then the Jev answer appears as a separate item, every pre-existing binding field is present and unchanged, and the file mode is 0600.

**AC-JEVG-004** — Given `mission-governor`'s definition and output shape, When compared before and after this SPEC, Then its tool set, scope boundary, and decision-object shape are unchanged.

**AC-JEVG-005** — Given a completion receipt, When its completion predicates and its landed-ancestry and authoritative-readback evidence are enumerated, Then no element derives from a Jev answer.

## §B. MCP wrapper and reference skill

**AC-JEVG-006** — Given the MCP tool wrapper, When its implementation is read, Then it delegates to `internal/jev` and constructs no HTTP request of its own.

**AC-JEVG-007** — Given both copies of `moai-mcp-tools.md`, When both tool-count figures in each are read, Then all four agree with the shipped tool set, and the two files remain byte-identical.

**AC-JEVG-008** — Given the reference skill, When it is read, Then it carries question-design rules and no call-path instruction.

## §C. Template and distribution

**AC-JEVG-009** — Given every new or changed file under `.claude/` or `.moai/`, When the template tree is checked, Then a corresponding file exists under `internal/template/templates/`.

**AC-JEVG-010** — Given every template file this SPEC touches, When it is scanned, Then it contains no SPEC ID, card id, date, price, measurement figure, or local-only-file reference. Method: the existing template-neutrality CI guard, with a positive control.

**AC-JEVG-011** — Given a change to an agent definition or a command source under the template tree, When the build runs, Then `make agents-emit-check` or `make commands-emit-check` respectively reports no drift.

## §D. Preserved negative result

**AC-JEVG-012** — Given the shipped record, When the premise-death use is looked up, Then the rejection is present with its measured figures (29.2% dead-call precision against a 25% base rate; 58.9% 2-class against a 75.0% constant; 67.5% English control), and `moai todo triage` contains no model call. Method: a grep over the triage path for the client symbol, with a positive control on a path that does call it.

---

## §E. Quality gates

- `go test ./internal/cli/... ./internal/template/...` passes; full-suite verdict from CI.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds; `make agents-emit-check` and `make commands-emit-check` report no drift.
- The template-neutrality CI guard is green on `internal/template/templates/**`.

## §F. Definition of Done

1. Every AC above passes.
2. Both `moai-mcp-tools.md` copies agree with the shipped tool set and with each other (AC-JEVG-007).
3. The governor's authority and the receipt's existing bindings are demonstrably unchanged (AC-JEVG-004, AC-JEVG-005).
4. The negative result is recorded where a future reader finds it before re-litigating it (AC-JEVG-012).
