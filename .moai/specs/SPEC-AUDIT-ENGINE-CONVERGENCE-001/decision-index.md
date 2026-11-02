# Decision Index — SPEC-AUDIT-ENGINE-CONVERGENCE-001

The authority register was searched: `.moai/project/product.md`, the HISTORY and `## Amendments` rows of prior completed SPECs, `.moai/config/sections/*.yaml`, and the project constitution.

The operator decision of 2026-11-03 and leader ruling R287 that relays it live in the leader's ledger, which is not a committed artifact. They are therefore cited by name, and the ruling text is quoted in each row it settles. The labels stay `FOUNDER`, because no committed anchor exists.

### Q1: Should `audit_multi` default `delta_base` from the card's newest FAIL verdict when called with `card_id`, accepting that such a call then files no second-review record?
Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree; leader ruling R287 (relaying the operator decision of 2026-11-03)
Why unresolved: it changes the default behavior of a shipped MCP tool for every `card_id` call after a FAIL, and today a delta call files no second-review record (`internal/cli/mcp_convergence.go:1059`).
Ruling quoted (R287): "Q1 = YES. audit_multi called with a card_id defaults delta_base to the card's newest FAIL verdict's audited_sha. A delta round may produce PASS only when (i) the delta verdicts since the last full review form an unbroken chain (each delta's base equals the previous verdict's SHA) AND (ii) the input of that delta includes the confirmation that every previous BLOCKING defect is closed. In that case the delta call must also file a second-review record (this changes the branch at internal/cli/mcp_convergence.go:1059 …). A broken chain means full scope. Returning to 'a final full review is always required' defeats B1's purpose."
Operator verdict: YES per leader ruling R287 (2026-11-03) — applied as REQ-AEC-001, -004, -005. Plan-audit round 1 (D1, D2) tightened the gate within this ruling: every chain record must have run every required backend, the closure set is the union over the chain's FAIL verdicts, and an explicit `delta_base` with a `card_id` goes through the same gate.

### Q2: How is a full-scope review requested once the default exists — the value `delta_base: "full"`, or a separate boolean input?
Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree
Why unresolved: both work; neither preserves current behavior and both undo with one revert.
Default: `delta_base: "full"` (rule: smaller user-visible surface — no new parameter)
Alternate: a separate boolean input
Operator verdict: DEFAULT-APPLIED 2026-11-02T17:50:12Z manager-spec+plan-author; approved by leader ruling R287

### Q3: Should the exported verdict vocabulary gain the label `INCONCLUSIVE` for an audit whose required leg could not run?
Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree; leader ruling R287
Why unresolved: it changes a template-distributed document's label set (`.moai/docs/audit-artifact-convention.md:87`) and every auditor's export.
Ruling quoted (R287): "Q3 = YES. The exported verdict vocabulary gains INCONCLUSIVE (.moai/docs/audit-artifact-convention.md:87 plus its template mirror). Consumer rules: not-performed; read as neither PASS nor FAIL; excluded from the plan-audit ceiling count (the Q8 default stands); it blocks card progress exactly as FAIL does but the blocking reason carries the error class (quota / timeout / auth); only the affected leg is re-run."
Operator verdict: YES per leader ruling R287 (2026-11-03) — applied as REQ-AEC-006, -007, -008

### Q4: How does a caller ask `audit_multi` to re-run only some legs — an input naming the legs (verdicts of the others carried from the server's own session state), or a separate tool?
Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree
Why unresolved: both carry verdicts only from server-written state; they differ in surface.
Default: an input on `audit_multi` naming the legs (rule: smaller user-visible surface — no new tool)
Alternate: a separate MCP tool
Operator verdict: DEFAULT-APPLIED 2026-11-02T17:50:12Z manager-spec+plan-author; approved by leader ruling R287

### Q5: Which binding relaxation does B3 use — an ancestry plus changed-path allowlist, or a code-tree digest that ignores the head?
Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree
Why unresolved: the dispatch delegated the choice; what each wrongly accepts is stated in `spec.md` §E.
Default: ancestry plus changed-path allowlist (rule: preserves current behavior — the existing digest and its exact-match path are kept)
Alternate: code-tree digest
Operator verdict: DEFAULT-APPLIED 2026-11-02T17:50:12Z manager-spec+plan-author; approved by leader ruling R287, with the two wrongly-accepted cases closed as R287 required (path exclusion for non-family report files; negative case for a changed line under `progress.md` §E.2). Narrowed by leader ruling R288: "A rewrite of progress.md section E.3 (the run audit-ready signal) after a PASS is the NORMAL flow and must NOT make the receipt stale; a changed line under section E.2 (the evidence body) MUST make it stale."

### Q6: Is the one approval input path for clearing a rejection a CLI verb on the human path, or an MCP tool?
Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree; leader ruling R287
Why unresolved: it decides who can clear an audit rejection, a security boundary between the operator and the audited lane.
Ruling quoted (R287): "Q6 = a CLI verb on the operator (human) path: reuse the TTY check, the agent-environment markers and the typed confirmation; NO MCP path. Include a PreToolUse deny reinforcement only if its cost is small inside this SPEC; otherwise record it as a Gap. Record the terminal-forgery residual risk."
Operator verdict: CLI verb on the human path, no MCP path, per leader ruling R287 (2026-11-03) — applied as REQ-AEC-011..013; the PreToolUse deny is recorded as a Gap (`spec.md` §E B4)

### Q7: Does the dispatch limit "no new writer" forbid B4's verb from appending an approval record and removing one mark?
Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree; leader ruling R287
Why unresolved: the dispatch text did not say which sense of "writer" it meant.
Ruling quoted (R287): "Q7 = allowed. 'No new writer' in the dispatch means: do not add a concurrently writing session or agent to a tree. The B4 verb appending one approval record and removing exactly one rejection mark is an operator invocation and is not forbidden."
Operator verdict: ALLOWED per leader ruling R287 (2026-11-03) — applied in REQ-AEC-016 and `spec.md` §D

### Q8: Do `INCONCLUSIVE` audit rounds count toward `harness.plan_audit_tier_ceilings`?
Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree
Why unresolved: this SPEC changes the label; the counting rule needed its own decision.
Default: excluded from this SPEC; the ceiling ladder is unchanged (rule: preserves current behavior)
Alternate: count only FAIL rounds, in this SPEC
Operator verdict: DEFAULT-APPLIED 2026-11-02T17:50:12Z manager-spec+plan-author; approved by leader ruling R287
Reconciliation note: R287's Q3 consumer rule says `INCONCLUSIVE` is "excluded from the plan-audit ceiling count (the Q8 default stands)". The round counter counts every iteration file regardless of label (`internal/runtime/audit_counter.go:10`). The Default above leaves the ladder unchanged, so it would keep counting `INCONCLUSIVE` files, while the Q3 consumer rule excludes them, which is this row's Alternate. The SPEC follows the explicit Q3 consumer rule (REQ-AEC-007). Leader ruling R288 confirmed this reading: Q8's "ladder unchanged" means the S1/M2/L3 numbers are untouched; what is counted is decided by Q3.
