# SPEC-FACTORY-LANE-NO-CLEAR-001 — Decision Index

Open decision rows for the operator. This index is written because `interview.decision_gate` is `on`. It is stateless on the status axis: the SPEC's lifecycle state lives only in `spec.md`. Every `Operator verdict:` line is empty at authoring, and no row is resolved by the authoring agent.

Labels use the four routes only: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. A `Default:` line appears only where the published Default rule ranks the options; the rule named in each row is the rule applied.

### Q1: Should the operator-session serial `--auto` cycle keep its per-card `/clear` guidance?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree settles this question; the requesting card does not count as a committed authority.
Why unresolved: The card scopes the lane transition and names the gtd `--auto` phrase. The same guidance text and code serve the operator-session serial cycle, which is not a factory lane. Removing it changes a shipped command's completion output.
Default: retain (rule: the option that preserves current behavior)
Alternate: remove the operator-session per-card guidance, leaving the lane transition rule as the only transition rule
Operator verdict:

### Q2: Should the `--clear-policy` launch flag and its accepted values (clear-each, clear-when-full, relaunch) be retained?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree settles this question.
Why unresolved: After this change, no lane completion output names a clear for any value (REQ-LNC-009). Retiring the flag or its values would change launch surfaces: the lane launch flag, the glm relaunch path, and the lane-resume flag list.
Default: retain (rule: the option that preserves current behavior)
Alternate: retire the flag and its values in a follow-up SPEC
Operator verdict:

### Q3: Should the sentence that only the local copy of `factory-dispatch.md` carries (around line 220) be carried into the template?

Label: POLICY-COVERED
Class: implementation-level
Authority anchor: `.claude/rules/moai/development/coding-standards.md` § Duplicate Prevention, which states "Update source file, not copies".
Why unresolved: The cited policy covers the direction: the template is the source, so the sentence belongs in the template and the local copy then matches it. The row records the divergence so the reviewer sees it.
Operator verdict:

### Q4: Should the root dogfood `.codex/agents/moai/manager-lead.toml` be synced from the regenerated template TOML?

Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The codex golden check pins only the template tree, not the root copy.
Why unresolved: The root copy differs from the template TOML in 106 lines, a difference that predates this change. The root directory also has no `manager-todo.toml`. A sync changes unrelated text in the root copy.
Default: leave the root copy untouched (rule: the option that preserves current behavior)
Alternate: sync the root copy from the regenerated template TOML, one commit with a larger diff
Operator verdict:

### Q5: Should the `.moai/policies/workflow/` factory-dispatch copies receive only the lane-transition passages, or a wholesale regeneration from the rule files?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree settles this question.
Why unresolved: The policy copies differ from the path-transformed rule files beyond the lane passages: `factory-dispatch.md` in 198 lines, `factory-dispatch-mechanics.md` in 69, `factory-dispatch-detail.md` in 4, `factory-dispatch-cards.md` in 0 (diff line counts). These copies are the text a Codex-only layout reads, so a wholesale regeneration changes what those projects read beyond this change.
Default: surgical edit of the lane-transition passages only (rule: the option that preserves current behavior)
Alternate: regenerate the five factory-dispatch policy copies from the rule files through the path transform
Operator verdict:

### Q6: Should Codex lane transitions be changed by this SPEC?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree settles this question.
Why unresolved: The Codex owned-card rule ends the lane session after completion. The root instruction file describes a new session for each Codex card. Codex lanes do not have the Claude worktree tools, so the Claude sequence cannot apply to them as written. The card names Claude Code lane transitions only.
Default: unchanged (rule: the option that preserves current behavior)
Alternate: extend the in-session transition to Codex lanes through the Codex worktree tools, with its own measurement
Operator verdict:

### Q7: Does the in-session switch (lease from the primary checkout, exit keeping the tree, enter the leased tree, verify the toplevel) succeed on the deployed Claude Code runtime?

Label: EVIDENCE-NEEDED
Class: implementation-level
Authority anchor: none. The only observation is recorded in session memory, which is not in the committed authority register.
Why unresolved: The sequence was recorded once on Claude Code 2.1.300 by a prior lane session. The run phase must observe it on the deployed runtime (AC-LNC-003). A refusal sends the lane to the blocker path of REQ-LNC-004.
Operator verdict:

### Q8: Should the policy-only kanban doctrine copies, which carry lane clear text, be updated or retired?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree settles this question.
Why unresolved: `kanban-dispatch.md` (8 literal clear lines) and `kanban-dispatch-detail.md` (7) exist only under `.moai/policies/workflow/`, with no rules-tree source. Whether this predecessor doctrine is live is not settled by a committed record checked by this plan.
Default: leave untouched (rule: the option that preserves current behavior)
Alternate: retire the orphan copies, or update their clear passages to the lane transition rule
Operator verdict:
