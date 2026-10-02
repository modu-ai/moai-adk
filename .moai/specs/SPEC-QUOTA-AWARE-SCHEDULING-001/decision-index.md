# decision-index.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

Decisions surfaced while planning that the operator has not settled. Each row states what is unresolved and why; it carries no recommendation. The working default each decision has in `spec.md` §B is a placeholder so the SPEC reads as implementable, not an answer to the question here. Labels follow the routing in the manager-spec definition: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. No row below has a committed-tree authority that decides it, so none is DECIDED or POLICY-COVERED.

### Q1: Should the StopFailure `rate_limit` turn-end also be stamped into the session record as an exhaustion time?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the decoded hook payload carries an error type and a message but no reset time and no provider; whether the hook process inherits the session environment, and whether a 429 turn-end triggers a final 100% statusline render, are unmeasured.
Operator verdict:

### Q2: How should two Claude accounts writing one project's record directory be told apart?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the statusline stdin carries no account identifier; whether more than one account actually writes the same state directory on this machine has not been measured.
Operator verdict:

### Q3: What are the numeric hold thresholds, release margin, and maximum record age?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: no measurement exists of quota consumed per card or per lane, so any five-hour hold, seven-day hold, margin, or age value would be unmeasured.
Operator verdict:

### Q4: Should the quota gate ship enabled or disabled in the distributed template?

Label: FOUNDER
Authority anchor: none
Why unresolved: shipping on helps Claude-subscription users and does nothing for others (no data, so it never holds); shipping off follows the sibling opt-in switches (slot lease, branch guard, served-model gate). The choice is a product call about the default user experience, not a measurement.
Operator verdict:

### Q5: Which exit status should `moai factory next` return when a Claude lane is held?

Label: FOUNDER
Authority anchor: none
Why unresolved: a new status is a public CLI contract that a supervising launcher could branch on; reusing the no-card status (3) keeps the contract smaller but makes a hold indistinguishable from an empty queue for any caller that only reads the status.
Operator verdict:

### Q6: Does the gate also apply to a card already assigned to, or already started by, the held lane?

Label: FOUNDER
Authority anchor: none
Why unresolved: gating every lease arm gives one predicate and one seam but can leave an in-progress card (returned by lease expiry) unresumed during the hold; exempting it adds card-state logic to the gate. Both are consistent with the card text.
Operator verdict:

### Q7: Should `moai integration acquire` only warn, or refuse the window to a pressured Claude lane?

Label: FOUNDER
Authority anchor: none
Why unresolved: the measured stall is a lane holding the integration window through a 429; a refusal removes that hazard but is a new gate with doctrine weight in the self-served integration procedure, which this card was not scoped to change.
Operator verdict:

### Q8: Should the leader or `moai todo --auto` be steered by the quota state beyond a read-only status block?

Label: FOUNDER
Authority anchor: none
Why unresolved: `--auto` runs inside the Claude session whose quota would be measured and has no lane or backend reference today; steering it would mean pausing or a new dispatch surface, which the card's lane-routing scope does not name.
Operator verdict:

### Q9: Should a high reading older than the maximum age, whose window has not yet reset, stay binding until the reset?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: keeping it binding assumes usage never falls inside a window; the Claude Code documentation calls the five-hour window rolling, and no measurement of a usage reading falling before its reset exists.
Operator verdict:

### Q10: Which variable defines "a Claude lane" for the gate — the launch provider, the kanban backend, or both?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the token a Claude factory lane actually carries in each variable was read from code but not observed on a live lane, and the kanban backend variable's stamp site sits in kanban code that is being removed.
Operator verdict:

### Q11: Should the windows ride the existing per-session record, a separate shared account file, or the existing home-level usage cache?

Label: FOUNDER
Authority anchor: none
Why unresolved: this is the data-model decision (record schema 2 → 3 versus a second writer versus reuse of a probe-fed cache); it is the decision most likely to change later and no committed document decides it.
Operator verdict:
