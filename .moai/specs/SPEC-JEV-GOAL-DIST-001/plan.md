# SPEC-JEV-GOAL-DIST-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B carries the autonomy-boundary decisions, which are the ones a reviewer should argue with; the mechanical distribution work is last, because it is the least reversible-by-argument and the most mechanically checkable.

---

## §A. Context

Fourth and last SPEC of the chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Depends on `SPEC-JEV-CONSUMERS-001` twice, in two different modes: the MCP tool counts cannot be settled until the shipped consumer set is known (a build dependency), and the seat-(i) withdrawal is this SPEC's recorded reaction to that SPEC's own closed state — M5 blocked, N2 open — not a consumption of its output.

Verified tree facts. Iter-1 rows were measured at `fd75cf692`; iter-2 rows were measured in this worktree at `ef3ad83e2` (branch `WT-goal-dist`) on 2026-09-22. A row measured at `fd75cf692` and not re-marked has not been re-measured at `ef3ad83e2`.

| Fact | Source | Observed | Tree |
|---|---|---|---|
| Tool counts, both copies | `.claude/rules/moai/core/moai-mcp-tools.md` and its template mirror | line 3 says "30 tools", line 63 says "26 of the 30"; the two files are byte-identical (`cmp` clean) | `ef3ad83e2` (re-measured) |
| Governor scope | `.claude/agents/moai/mission-governor.md` | `tools: Read, Grep, Glob, Skill`; `permissionMode: plan`; returns one bounded decision object; never applies it | `fd75cf692` |
| Sealed-loop contract | `.claude/skills/moai/workflows/goal.md` § `/moai goal --auto` | after the single approval the loop asks no further user questions; anything exceeding sealed scope becomes a persisted blocked result and stops the loop without effects; receipts at `.moai/state/mission/governance/` are 0600 and bind mission, contract, snapshot, action, targets, expiry, issuer, HEAD, status, digest | `fd75cf692` |
| `scripts/jev/` on develop | `ls scripts/` | absent — no `jev` entry | `ef3ad83e2` (re-measured) |
| CONSUMERS M5 blocked, seat (i) forbidden | `SPEC-JEV-CONSUMERS-001/progress.md` §E.4 (:148) | *"M5 and SPEC-JEV-GOAL-DIST-001 seat (i) remain blocked and are forbidden this session"*; sync status (:106): *"no consumer ships; M4/M6 recorded gate-unrun (REQ-JEVN-016), M5 block recorded"* | `ef3ad83e2` |
| N2 OPEN, unowned | `SPEC-JEV-CONSUMERS-001/spec.md` §E (:170) | *"OPEN — blocking M5 and M6"*; carried to CONSUMERS' Kickoff gate, which directed M6 into gate-unrun but left M5 blocked; CONSUMERS is closed `completed`, so N2 has no live owner | `ef3ad83e2` |
| Three-state + "shipped" reading precedent | `SPEC-JEV-CONSUMERS-001/spec.md` REQ-JEVN-015 (:106) / REQ-JEVN-016 (:108, conditions i–iv) | gate unrun → recorded decision naming what the gate needs and who owns it, citing no measurement; "shipped" = reachable at the shipped default | `ef3ad83e2` |
| Fitness gate binding, unrun | `SPEC-JEV-OPTIN-MEASURE-001/spec.md` REQ-JEVO-009 (:76) | a consumer not beating its constant-answer baseline shall not ship; the verdict is binding; gate unrun | `ef3ad83e2` |
| Gate key and shipped default | `internal/template/templates/.moai/config/sections/workflow.yaml` (:180) / `.moai/config/sections/workflow.yaml` | template mirror carries `jev: enabled: false` (shipped default OFF); the local config carries no `jev:` block and resolves to the same default | `ef3ad83e2` |
| Declared-consumer guard | `internal/cli/doctor_jev_test.go:221` | `TestJevCallPath_HasExactlyTheDeclaredConsumers`, allowlist: `doctor_jev.go` (control), `todo_jev_finding.go`, `jev_skill_suggest.go` — the wrapper is a new importer it does not know | `ef3ad83e2` |
| Mirror-test non-enumeration | `internal/template/rule_template_mirror_test.go` | zero matches for `moai-mcp-tools` — no mechanical mirror enforcement for that file exists today | `ef3ad83e2` |
| Record destination viability | `docs/` vs `.moai/docs/` | repo-root `docs/` is tracked (`git ls-files`) with **no** `internal/template/templates/docs/` mirror → outside Template-First scope and the neutrality constraint; `.moai/docs/` is tracked too **but** `internal/template/templates/.moai/docs/` exists, so a new file there would be forced into the template mirror where REQ-JEVG-010 forbids the figures | `ef3ad83e2` |
| Lane-question host structural finding | `internal/cli/state.go` `runShowBlocker` (:156, glob at :165) | the t1066 run's recorded measurement: the only Go path holding lane questions reads `blocker-*.json`; this SPEC cites the CONSUMERS block record rather than re-deriving the finding | `ef3ad83e2` (anchor) / t1066 (measurement) |
| Mission receipt contract anchor | `internal/cli/goal.go:326,500` | `MissionContract{… RecoveryConditions: ["authoritative_readback"] …}`; `authoritative_readback_missing` error | `ef3ad83e2` |
| aitmpl ops-checklist report | `.moai/reports/aitmpl-jev-skill-suggestion-20260922/` | **absent from this tree**; cited from the plan-audit iter-1 record only | `ef3ad83e2` |

## §B. Autonomy-boundary decisions (review these first)

**B1 — Seat (i) is withdrawn; what remains is the unchanged-behaviour guarantee.** The predecessor's own record closes this decision: `SPEC-JEV-CONSUMERS-001` M5 — seat (i)'s sole producer — is recorded blocked, the host-design question N2 is OPEN with no live owner (CONSUMERS closed `completed` with the block recorded), and the t1066 run's structural finding is that the only Go path holding lane questions reads `blocker-*.json` shapes nothing in production writes as lane questions. There is no host to wire and no design to wire it to.

The alternative — a three-state conditional ("routing proceeds only if M5 unblocks", mirroring CONSUMERS REQ-JEVN-015/016) — was considered and rejected at iter-2. The mirror is superficially apt but structurally different: for the consumers, the conditional's enabling branch was a measurement gate that the SPEC's own machinery could run; here the enabling branch needs N2 answered, a host designed, and an owner appointed — none of which exists or is scheduled. A conditional whose enabling branch has no owner is a deferred promise, which is the exact "silently promise it" failure the seat-(i)/M5 constraint forbids. Option (c) (park without a named home) was rejected as strictly weaker than recording the future home explicitly.

The disposition itself is recorded in spec.md §C.1 and §F; N2 is not resolved here. What this SPEC still owes is the regression guarantee: the loop's blocked-question outcome is the pre-SPEC outcome, unchanged (REQ-JEVG-001, AC-JEVG-001 against the pinned baseline `ef3ad83e2`).

Two properties stay asserted: no user question is introduced inside the loop (REQ-JEVG-002, AC-JEVG-002), and the blocked-question path behaves exactly as the pinned baseline (AC-JEVG-001).

**B2 — Seat (ii) adds one recorded item to a snapshot and nothing else.** The governor reads a sealed snapshot and returns one bounded decision object; a deterministic executor validates it and performs any state change. A Jev Noul is one more piece of evidence in that snapshot, recorded as a separate item in the governance receipt so a reader can see what the governor was shown.

Three things stay untouched: the governor's read-only scope, its output shape, and the receipt's existing binding fields. AC-JEVG-004 verifies this by comparing the agent definition before and after.

**B3 — The hard line, and why it is drawn at inputs rather than decisions.** REQ-JEVG-005: a Jev answer never becomes an element of a completion predicate, nor of landed-ancestry or authoritative-readback evidence. Those are the irreversible judgments, and a rule forbidding the model from *deciding* them would be easy to route around — the model answers, the loop reads the answer, the loop's own reasoning carries the confidence forward, and nothing records how it got there. Forbidding it as an input closes that route and makes AC-JEVG-005 a reachability question a test can answer.

This is the clause most worth arguing with, because it is also the clause that makes seat (ii) less useful than it could be: a governor that may read a Noul but may not let it reach a completion predicate is being handed evidence it cannot fully act on. That asymmetry is deliberate and is the price of the display-only invariant surviving into an autonomous loop, where no human is present to notice it eroding.

## §C. Distribution decisions

**C1 — The MCP tool is a caller, not a layer, and it inherits the chain's gate state.** It holds no transport code of its own. A second implementation would mean a second set of size bounds, a second fail-open policy, and a second model pin — and the one most likely to drift is the one nobody is testing. It also stays inert behind the `workflow.jev.enabled` default-false gate (template default measured at `ef3ad83e2`: `jev: enabled: false`; local config carries no block and resolves to the same default): with the gate off it constructs no request and reports gated-unavailable.

**C1b — While the fitness gate stands unrun, the wrapper is present but unpresented.** `SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009 is binding and unrun; the consumers sit in the CONSUMERS REQ-JEVN-016 gate-unrun state. An MCP tool that made the unmeasured capability reachable from any agent session — with a catalogue entry presenting it as available — would be a reachability channel around exactly the boundary the chain drew. So: "shipped" carries CONSUMERS' declared reading (reachable at the shipped default), the wrapper's catalogue presentation in both copies of `moai-mcp-tools.md` states gated-unavailable, its disposition record names what the fitness gate still needs and who owns it and cites no measurement, and the declared-consumer guard `TestJevCallPath_HasExactlyTheDeclaredConsumers` — whose allowlist (measured: three entries) does not know the wrapper — is extended for the wrapper's registration or its refusal to extend is justified in the run-phase record.

**C2 — Both copies, both figures.** `moai-mcp-tools.md` exists twice (local rule and template mirror), byte-identical (re-measured at `ef3ad83e2` via `cmp`), and each copy carries two counts: "30 tools" at line 3 and "26 of the 30" at line 63. Adding a tool means four edits, and AC-JEVG-007 asserts all four agree and the two files stay byte-identical. The byte-identity assertion is what catches an edit applied to one copy only. Honesty note: `rule_template_mirror_test.go` does not enumerate this file (measured, zero matches), so AC-JEVG-007's own test is the only mechanical enforcement; enrolling the file in the mirror test is a run-phase option, adopted or declined with a recorded reason.

**C3 — The reference skill carries rules, not a call path.** Question-design guidance (compute in Go, no-match option, small state, both Noul polarities) belongs where an author reads it. The call path belongs in `internal/jev` and is reached through the tool. A skill that carried both would become a second place to describe the call, and skills drift faster than code.

**C4 — `scripts/jev/`: supersede without porting, and write the disposition down.** The scripts exist only as an uncommitted working copy in the primary checkout, absent from `develop` (re-measured at `ef3ad83e2`), with no template mirror, so they reach no user project. After the chain lands, their behaviour is available through the binary and the MCP tool. They are not committed, not distributed, and not maintained; the Go package is canonical. Nothing in this SPEC deletes anything from anyone's working tree.

**C5 — Record destinations, named.** Both records — the preserved negative result with its figures (REQ-JEVG-013) and the scripts/jev disposition (REQ-JEVG-012) — live in **`docs/jev-negative-results.md`**, a new file in the repo-root `docs/` directory. That surface was chosen by measurement, not preference: it is tracked (`git ls-files` lists `docs/design/…` precedents), it has **no** `internal/template/templates/docs/` mirror, so it sits outside both the Template-First obligation (REQ-JEVG-009 scopes to `.claude/`/`.moai/`) and the template-neutrality constraint (REQ-JEVG-010) that forbids measurement figures; and it is where a reader about to re-litigate the premise-death use will look. `.moai/docs/` was considered and rejected: it is tracked too, but `internal/template/templates/.moai/docs/` exists, so a new file there would be forced into the template mirror — where the figures the record exists to carry are neutrality-forbidden. That collision is exactly the defect D3 named.

## §D. Constraints

- **No authority moves.** The governor's scope, output shape, and receipt bindings are unchanged; no Jev answer reaches a completion predicate.
- **No new user question** inside the sealed-scope loop.
- **Thin wrapper only.** No transport code outside `internal/jev`.
- **Both copies, both figures** for the MCP tool counts.
- **Template-First.** Every `.claude/` or `.moai/` file lands in `internal/template/templates/` first; `make agents-emit` if an agent definition changed, `make commands-emit` if a command source changed.
- **Template neutrality.** No SPEC ID, card id, date, price, measurement figure, or local-only-file reference in template content — which means the negative result's *figures* live in this SPEC's artifacts and in `docs/jev-negative-results.md`, never in shipped or template content (destination reasoning: §C5).
- **Wrapper gate discipline.** The MCP wrapper is inert at the shipped default, is presented as gated-unavailable while the fitness gate stands unrun, and is known to (or justifiedly absent from) the declared-consumer guard.
- **Verification scope.** Affected packages only; full-suite verdict from CI.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject. AC-JEVG-012's grep is reported with its positive control, since a zero-hit and a broken search are indistinguishable without one. AC-JEVG-013's gate-off invocation and the guard-allowlist disposition (extended or justifiedly refused) are reported explicitly. Any before/after criterion is judged against the baseline pinned in `acceptance.md` (`ef3ad83e2`, or the run-entry SHA re-pinned in `progress.md` §E.2).

## §F. Milestones

**M7a — Seat (i) disposition.** No routing is built. The milestone's output is the recorded disposition (spec.md §C.1, §F) and the regression proof that the blocked-question path behaves exactly as the pinned baseline (AC-JEVG-001). Its once-awaited producer, `SPEC-JEV-CONSUMERS-001` M5, is recorded blocked with N2 open — which is the reason for the withdrawal, not a wait. Any run-phase discovery that a host path for lane questions has appeared in the tree is a blocker report to the orchestrator (re-plan), not a licence to build the seat here.

**M7b — Seat (ii).** Noul recorded as a separate item in the governance receipt; governor definition unchanged; completion-predicate exclusion asserted.

**M8a — MCP wrapper.** Thin tool over `internal/jev`, inert behind the `workflow.jev.enabled` default-false gate; both tool-count figures updated in both copies of `moai-mcp-tools.md` with the wrapper presented as gated-unavailable while the fitness gate stands unrun; byte-identity between the copies preserved; the wrapper's gate-unrun disposition recorded; the declared-consumer guard extended or its refusal justified.

**M8b — Reference skill.** Question-design rules only; template mirror; emit-checks if an agent or command source moved.

**M8c — Records.** `docs/jev-negative-results.md` written: §1 the preserved negative result with its measured figures, §2 the `scripts/jev/` disposition with the Go package as canonical. Priority Low in sequence, not in importance — it is last because the surfaces it documents are only final after M8a and M8b.

## §G. Anti-patterns

- **Letting a Jev answer into a completion predicate.** The single most consequential failure available in this SPEC, and the one an autonomous loop has no human present to catch.
- **Adding a user question inside the sealed loop** to resolve a routing ambiguity. The ambiguity's answer is "operator-owned", which is already a defined outcome.
- **Treating an idle or absent routing answer as a classification.** An unavailable result is no signal, not a lead-owned signal.
- **Editing one copy of `moai-mcp-tools.md`.** The byte-identity assertion exists because this has a history of happening.
- **Updating one count and not the other.** Each copy carries two figures; three of four correct reads as done.
- **Putting the negative result's figures in template content.** Neutrality forbids measurement figures there; the record lives in this SPEC's artifacts and `docs/jev-negative-results.md` (§C5).
- **Putting the records in `.moai/docs/`.** That directory has a template mirror, so a new file there gets pulled into the neutrality constraint the figures cannot satisfy. The destination is the mirrorless `docs/` (§C5).
- **Re-running the premise-death experiment** because a new gate or prompt suggests itself. The sweep was flat and the English arm measured; a new run needs a new reason.
- **Reading the withdrawn seat's producer as merely "not yet done".** M5 is recorded blocked with its design question unowned; treating it as a wait re-imports the deferred promise the withdrawal removed.
- **Presenting the MCP wrapper as available while the fitness gate stands unrun.** The catalogue entry is the reachability channel this SPEC's gate coupling exists to close.

## §H. Cross-references

- `SPEC-JEV-CONSUMERS-001` — predecessor; its closed state (M5 block recorded, N2 open) is the provenance of the seat-(i) withdrawal, and its shipped consumer set fixes the tool counts.
- `SPEC-JEV-CONSUMERS-001` REQ-JEVN-015/016 — the three-state gate discipline and the declared "shipped = reachable at the shipped default" reading this SPEC adopts for the wrapper.
- `SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009 — the binding fitness gate the wrapper's presentation restraint tracks.
- `SPEC-JEV-CORE-001` — owns the call path the MCP tool wraps and the display-only invariant both the seat and the wrapper preserve.
- `.claude/skills/moai/workflows/goal.md` § `/moai goal --auto` — the sealed-scope loop whose blocked-question behaviour REQ-JEVG-001 pins.
- `.claude/agents/moai/mission-governor.md` — the read-only decision agent seat (ii) supplies an auxiliary input to.
- `.claude/rules/moai/core/moai-mcp-tools.md` + its template mirror — the two files C2 updates; `internal/template/rule_template_mirror_test.go` does not enumerate them (C2 honesty note).
- `internal/cli/doctor_jev_test.go` `TestJevCallPath_HasExactlyTheDeclaredConsumers` — the declared-consumer guard C1b reconciles.
- `docs/jev-negative-results.md` — the record destination C5 names (created by M8c; absent from the tree until then).
- `internal/cli/goal.go` `MissionContract` (:326) — the receipt-completion anchor AC-JEVG-005 enumerates.
