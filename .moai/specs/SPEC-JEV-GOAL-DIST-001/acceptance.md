# SPEC-JEV-GOAL-DIST-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here. Every criterion names the REQ id(s) it verifies.

**Numbering note (iter-2).** `AC-JEVG-013` and `AC-JEVG-014` were appended in the 0.2.0 revision without renumbering the existing twelve; each sits in the section of the requirement it verifies, so numeric order and document order diverge by design (§B ends at 013; §C ends at 014; §D carries 012).

**Baseline pin.** The pre-SPEC baseline for every before/after comparison in this artifact is the plan iter-2 freeze commit `ef3ad83e2` (branch `WT-goal-dist`, this worktree). If run-phase entry lands on a later commit, manager-develop re-pins the baseline to the run-entry commit and records the new SHA in `progress.md` §E.2 before any before/after criterion is judged; a criterion is then evaluated against the re-pinned SHA, never against a branch name.

## §A. `/moai goal --auto` seats

**AC-JEVG-001** — Given the pinned pre-SPEC baseline and a lane blocking with a question inside the sealed-scope loop, When the post-SPEC tree handles that block, Then the outcome matches the baseline behaviour — a persisted blocked result with the same shape as the pre-SPEC behaviour, and the loop stops without effects — with no routing consumed and no item answered by classification. *(Verifies REQ-JEVG-001.)* Method: a test asserting the loop's blocked-question path (persisted-result shape, stop semantics) against the pinned baseline, plus a diff of the blocked-question handling code between the pinned baseline and the run tip.

**AC-JEVG-002** — Given the sealed-scope loop runs from its single approval to completion, When any blocking question arises, Then no `AskUserQuestion` is issued inside the loop. *(Verifies REQ-JEVG-002.)* Method: a guard asserting zero user-question emissions between approval and completion.

**AC-JEVG-003** — Given a governance receipt written with a Jev Noul supplied, When the receipt is parsed, Then the Jev answer appears as a separate item, every pre-existing binding field is present and unchanged, and the file mode is 0600. *(Verifies REQ-JEVG-003, REQ-JEVG-004.)*

**AC-JEVG-004** — Given `mission-governor`'s definition and output shape, When `.claude/agents/moai/mission-governor.md` at the pinned baseline is diffed against the post-SPEC tree (and its template mirror likewise), Then its tool set, scope boundary, and decision-object shape are unchanged. *(Verifies REQ-JEVG-004.)*

**AC-JEVG-005** — Given a completion receipt, When its completion predicates and its landed-ancestry and authoritative-readback evidence are enumerated — the enumeration the mission contract construction assembles (the live anchor is `internal/cli/goal.go` `MissionContract`, `RecoveryConditions: ["authoritative_readback"]`) — Then no enumerated element derives from a Jev answer. *(Verifies REQ-JEVG-005.)* Method: enumerate the receipt and completion-predicate input fields in the contract writer, grep the enumerated set for the `internal/jev` client symbol, with a positive control on a path that does import it (e.g. `internal/cli/jev_skill_suggest.go`).

## §B. MCP wrapper and reference skill

**AC-JEVG-006** — Given the MCP tool wrapper, When its implementation is read, Then it delegates to `internal/jev` and constructs no HTTP request of its own. *(Verifies REQ-JEVG-006.)*

**AC-JEVG-007** — Given both copies of `moai-mcp-tools.md`, When both tool-count figures in each are read, Then all four agree with the shipped tool set, and the two files remain byte-identical. *(Verifies REQ-JEVG-007.)* Method: a Go test byte-comparing the two copies (`cmp`-equivalent) and asserting the four figure values against the registered tool count. **Mirror-test honesty note (measured 2026-09-22, this tree):** `internal/template/rule_template_mirror_test.go` does not enumerate `moai-mcp-tools.md` (grep count 0), so no mechanical enforcement for this file's mirror parity exists today; AC-JEVG-007's own test is the only mechanical check, and enrolling the file in the mirror test is a run-phase option to record in `progress.md` §E.2 (adopt or decline with a reason).

**AC-JEVG-008** — Given the reference skill, When it is read, Then it carries question-design rules and no call-path instruction. "No call-path instruction" means the forbidden token set is absent: any import path or package qualifier of `internal/jev`, any invocation of the MCP wrapper tool by its registered name, and any shell command example invoking a `moai` jev surface. *(Verifies REQ-JEVG-008.)* Method: grep the skill, both copies, for the forbidden token set, with a positive control on a file that does contain one.

**AC-JEVG-013** — Given the MCP tool wrapper on the shipped default (`workflow.jev.enabled: false` in the template default; the block absent from the local config, which resolves to the same default), When the tree is examined and the wrapper invoked with the gate off, Then the wrapper constructs no request and reports gated-unavailable; both copies of `moai-mcp-tools.md` present the tool as gated-unavailable rather than available; the wrapper's gate-unrun disposition is recorded in the run-phase record naming what the fitness gate still needs and who owns it, citing no measurement; and the declared-consumer guard `TestJevCallPath_HasExactlyTheDeclaredConsumers` (`internal/cli/doctor_jev_test.go`) knows the wrapper's registration — its allowlist extended for the wrapper, or the refusal to extend it justified in the run-phase record. *(Verifies REQ-JEVG-006, REQ-JEVG-007.)* Method: a test invoking the wrapper with the gate off asserting the unavailable response and zero constructed requests; inspection of the guard allowlist for the wrapper entry or the recorded justification.

## §C. Template and distribution

**AC-JEVG-009** — Given every file this SPEC's branch adds or modifies under `.claude/` or `.moai/` — the diff of this branch against the pinned baseline, template-tree counterparts included — When the template tree is checked, Then a corresponding file exists under `internal/template/templates/`. *(Verifies REQ-JEVG-009.)*

**AC-JEVG-010** — Given every template file this SPEC touches, When it is scanned, Then it contains no SPEC ID, card id, date, price, measurement figure, or local-only-file reference. *(Verifies REQ-JEVG-010.)* Method: the existing template-neutrality CI guard, with a positive control.

**AC-JEVG-011** — Given a change to an agent definition or a command source under the template tree, When the build runs, Then `make agents-emit-check` or `make commands-emit-check` respectively reports no drift. *(Verifies REQ-JEVG-011.)*

**AC-JEVG-014** — Given the scripts/jev disposition section of `docs/jev-negative-results.md`, When it is read, Then it states the disposition of the uncommitted `scripts/jev/` working copy with the Go package as the canonical implementation, and records that the scripts are not committed, not distributed, and not maintained. *(Verifies REQ-JEVG-012.)*

## §D. Preserved negative result

**AC-JEVG-012** — Given the preserved-negative-result record at `docs/jev-negative-results.md` — a tracked, non-template surface (the repo-root `docs/` directory is tracked and has no `internal/template/templates/docs/` mirror, so it is outside both the Template-First scope and the neutrality constraint) — When the premise-death use is looked up there, Then the rejection is present with its measured figures (29.2% dead-call precision against a 25% base rate; 58.9% 2-class against a 75.0% constant; 67.5% English control on 40 cards), and `moai todo triage` contains no model call. *(Verifies REQ-JEVG-013, REQ-JEVG-014.)* Method: the record file exists at the named path and carries the figures; a grep over the triage path for the client symbol, with a positive control on a path that does call it.

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
4. The negative result and the scripts/jev disposition are recorded at `docs/jev-negative-results.md`, where a future reader finds them before re-litigating them (AC-JEVG-012, AC-JEVG-014).
5. The MCP wrapper is inert and unpresented-as-available at the shipped default, with its gate-unrun disposition recorded and the declared-consumer guard reconciled (AC-JEVG-013).
6. The sealed-scope loop's blocked-question behaviour is unchanged from the pinned baseline (AC-JEVG-001, AC-JEVG-002).
