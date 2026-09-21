# SPEC-JEV-GOAL-DIST-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B carries the autonomy-boundary decisions, which are the ones a reviewer should argue with; the mechanical distribution work is last, because it is the least reversible-by-argument and the most mechanically checkable.

---

## §A. Context

Fourth and last SPEC of the chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Depends on `SPEC-JEV-CONSUMERS-001` twice: seat (i) consumes Consumer A's routing, and the MCP tool counts cannot be settled until the shipped consumer set is known.

Verified tree facts (measured in this worktree at `fd75cf692`):

| Fact | Source | Observed |
|---|---|---|
| Tool counts, both copies | `.claude/rules/moai/core/moai-mcp-tools.md` and its template mirror | line 3 says "30 tools", line 63 says "26 of the 30"; the two files are byte-identical today |
| Governor scope | `.claude/agents/moai/mission-governor.md` | `tools: Read, Grep, Glob, Skill`; `permissionMode: plan`; returns one bounded decision object; never applies it |
| Sealed-loop contract | `.claude/skills/moai/workflows/goal.md` § `/moai goal --auto` | after the single approval the loop asks no further user questions; anything exceeding sealed scope becomes a persisted blocked result and stops the loop without effects; receipts at `.moai/state/mission/governance/` are 0600 and bind mission, contract, snapshot, action, targets, expiry, issuer, HEAD, status, digest |
| `scripts/jev/` on develop | `ls scripts/` | absent — 16 entries, no `jev` |

## §B. Autonomy-boundary decisions (review these first)

**B1 — Seat (i) changes which side of an existing line an item lands on, not where the line is.** The loop's contract already has two outcomes for a blocking question: the lead handles it, or it becomes a persisted blocked result and the loop stops. Routing does not add a third. It classifies, and the classification decides which of the two existing outcomes applies.

Two properties stay untouched and both are asserted: no user question is introduced inside the loop (REQ-JEVG-002, AC-JEVG-002), and an operator-owned item reaches exactly the persisted blocked result it reaches today (AC-JEVG-001). The uncertainty direction is inherited from Consumer A: an `operator` answer routes to operator, and so does a below-threshold confidence. Uncertainty escalates; it never downgrades.

**B2 — Seat (ii) adds one recorded item to a snapshot and nothing else.** The governor reads a sealed snapshot and returns one bounded decision object; a deterministic executor validates it and performs any state change. A Jev Noul is one more piece of evidence in that snapshot, recorded as a separate item in the governance receipt so a reader can see what the governor was shown.

Three things stay untouched: the governor's read-only scope, its output shape, and the receipt's existing binding fields. AC-JEVG-004 verifies this by comparing the agent definition before and after.

**B3 — The hard line, and why it is drawn at inputs rather than decisions.** REQ-JEVG-005: a Jev answer never becomes an element of a completion predicate, nor of landed-ancestry or authoritative-readback evidence. Those are the irreversible judgments, and a rule forbidding the model from *deciding* them would be easy to route around — the model answers, the loop reads the answer, the loop's own reasoning carries the confidence forward, and nothing records how it got there. Forbidding it as an input closes that route and makes AC-JEVG-005 a reachability question a test can answer.

This is the clause most worth arguing with, because it is also the clause that makes seat (ii) less useful than it could be: a governor that may read a Noul but may not let it reach a completion predicate is being handed evidence it cannot fully act on. That asymmetry is deliberate and is the price of the display-only invariant surviving into an autonomous loop, where no human is present to notice it eroding.

## §C. Distribution decisions

**C1 — The MCP tool is a caller, not a layer.** It holds no transport code of its own. A second implementation would mean a second set of size bounds, a second fail-open policy, and a second model pin — and the one most likely to drift is the one nobody is testing.

**C2 — Both copies, both figures.** `moai-mcp-tools.md` exists twice (local rule and template mirror), byte-identical today, and each copy carries two counts: "30 tools" at line 3 and "26 of the 30" at line 63. Adding a tool means four edits, and AC-JEVG-007 asserts all four agree and the two files stay byte-identical. The byte-identity assertion is what catches an edit applied to one copy only.

**C3 — The reference skill carries rules, not a call path.** Question-design guidance (compute in Go, no-match option, small state, both Noul polarities) belongs where an author reads it. The call path belongs in `internal/jev` and is reached through the tool. A skill that carried both would become a second place to describe the call, and skills drift faster than code.

**C4 — `scripts/jev/`: supersede without porting.** The scripts exist only as an uncommitted working copy in the primary checkout, absent from `develop`, with no template mirror, so they reach no user project. After the chain lands, their behaviour is available through the binary and the MCP tool. They are not committed, not distributed, and not maintained; the Go package is canonical. Nothing in this SPEC deletes anything from anyone's working tree.

## §D. Constraints

- **No authority moves.** The governor's scope, output shape, and receipt bindings are unchanged; no Jev answer reaches a completion predicate.
- **No new user question** inside the sealed-scope loop.
- **Thin wrapper only.** No transport code outside `internal/jev`.
- **Both copies, both figures** for the MCP tool counts.
- **Template-First.** Every `.claude/` or `.moai/` file lands in `internal/template/templates/` first; `make agents-emit` if an agent definition changed, `make commands-emit` if a command source changed.
- **Template neutrality.** No SPEC ID, card id, date, price, measurement figure, or local-only-file reference in template content — which means the negative result's *figures* live in the SPEC and the rule file, not in the shipped template.
- **Verification scope.** Affected packages only; full-suite verdict from CI.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject. AC-JEVG-012's grep is reported with its positive control, since a zero-hit and a broken search are indistinguishable without one.

## §F. Milestones

**M7a — Seat (i).** Routing consumed at the lane-block point inside the sealed-scope loop; lead-owned and cheap-to-reverse items answered; operator-owned items unchanged. Depends on `SPEC-JEV-CONSUMERS-001` M5.

**M7b — Seat (ii).** Noul recorded as a separate item in the governance receipt; governor definition unchanged; completion-predicate exclusion asserted.

**M8a — MCP wrapper.** Thin tool over `internal/jev`; both tool-count figures updated in both copies of `moai-mcp-tools.md`; byte-identity between the copies preserved.

**M8b — Reference skill.** Question-design rules only; template mirror; emit-checks if an agent or command source moved.

**M8c — Records.** The `scripts/jev/` disposition and the preserved negative result, written where a reader finds them before re-litigating. Priority Low in sequence, not in importance — it is last because the surfaces it documents are only final after M8a and M8b.

## §G. Anti-patterns

- **Letting a Jev answer into a completion predicate.** The single most consequential failure available in this SPEC, and the one an autonomous loop has no human present to catch.
- **Adding a user question inside the sealed loop** to resolve a routing ambiguity. The ambiguity's answer is "operator-owned", which is already a defined outcome.
- **Treating an idle or absent routing answer as a classification.** An unavailable result is no signal, not a lead-owned signal.
- **Editing one copy of `moai-mcp-tools.md`.** The byte-identity assertion exists because this has a history of happening.
- **Updating one count and not the other.** Each copy carries two figures; three of four correct reads as done.
- **Putting the negative result's figures in template content.** Neutrality forbids measurement figures there; the record lives in the SPEC and the rule file.
- **Re-running the premise-death experiment** because a new gate or prompt suggests itself. The sweep was flat and the English arm measured; a new run needs a new reason.

## §H. Cross-references

- `SPEC-JEV-CONSUMERS-001` — predecessor; seat (i) consumes its Consumer A routing, and the tool counts depend on its shipped set.
- `SPEC-JEV-CORE-001` — owns the call path the MCP tool wraps and the display-only invariant both seats preserve.
- `.claude/skills/moai/workflows/goal.md` § `/moai goal --auto` — the sealed-scope loop seat (i) enters.
- `.claude/agents/moai/mission-governor.md` — the read-only decision agent seat (ii) supplies an auxiliary input to.
- `.claude/rules/moai/core/moai-mcp-tools.md` + its template mirror — the two files C2 updates.
