# decision-index.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

Decisions surfaced while planning. Each row states what was unresolved and why, carries no recommendation, and records the verdict that closed it. Resolutions: the lane resolved DO-1..DO-11 with the decision oracle (Jev; confidence quoted per row), the leader's verdict then settled the three escalations (DO-3, DO-7, DO-8), and DO-12 was resolved by the oracle last, after its premise was verified in code. Labels follow the routing in the manager-spec definition: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. No row had a committed-tree authority that decides it, so none is DECIDED or POLICY-COVERED.

### Q1: Should the StopFailure `rate_limit` turn-end also be stamped into the session record as an exhaustion time? (DO-1)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the decoded hook payload carries an error type and a message but no reset time and no provider; whether the hook process inherits the session environment, and whether a 429 turn-end triggers a final 100% statusline render, are unmeasured.
Operator verdict: Resolved by decision oracle — not now (confidence 1.00).

### Q2: How should two Claude accounts writing one project's record directory be told apart? (DO-2)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the statusline stdin carries no account identifier; whether more than one account actually writes the same state directory on this machine has not been measured.
Operator verdict: Resolved by decision oracle — reset time only (confidence 0.97).

### Q3: What are the numeric hold thresholds, release margin, and maximum record age? (DO-3)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: no measurement exists of quota consumed per card or per lane, so any five-hour hold, seven-day hold, margin, or age value would be unmeasured.
Operator verdict: Oracle chose 90 / 95 / 5 / 30m at confidence 0.32 (LOW); the leader ACCEPTED them as provisional. They are unmeasured defaults, all four are configuration keys, and re-setting them after the first real measurement is a listed follow-up.

### Q4: Should the quota gate ship enabled or disabled in the distributed template? (DO-4)

Label: FOUNDER
Authority anchor: none
Why unresolved: shipping on helps Claude-subscription users and does nothing for others (no data, so it never holds); shipping off follows the sibling opt-in switches (slot lease, branch guard, served-model gate). The choice is a product call about the default user experience, not a measurement.
Operator verdict: Resolved by decision oracle — false in the template, true in the local `workflow.yaml` (confidence 0.99).

### Q5: Which exit status should `moai factory next` return when a Claude lane is held? (DO-5)

Label: FOUNDER
Authority anchor: none
Why unresolved: a new status is a public CLI contract that a supervising launcher could branch on; reusing the no-card status (3) keeps the contract smaller but makes a hold indistinguishable from an empty queue for any caller that only reads the status.
Operator verdict: Resolved by decision oracle — reuse status 3 (confidence 0.81); the stderr hold line is the discriminator, and a supervising launcher treating 3 as "back off" is the intended behaviour.

### Q6: Does the gate also apply to a card already assigned to, or already started by, the held lane? (DO-6)

Label: FOUNDER
Authority anchor: none
Why unresolved: gating every lease arm gives one predicate and one seam but can leave an in-progress card (returned by lease expiry) unresumed during the hold; exempting it adds card-state logic to the gate. Both are consistent with the card text.
Operator verdict: Resolved by decision oracle — exempt cards the lane already holds or started (confidence 0.99); the gate applies to leasing new cards only.

### Q7: Should `moai integration acquire` only warn, or refuse the window to a pressured Claude lane? (DO-7)

Label: FOUNDER
Authority anchor: none
Why unresolved: the measured stall is a lane holding the integration window through a 429; a refusal removes that hazard but is a new gate with doctrine weight in the self-served integration procedure, which this card was not scoped to change.
Operator verdict: Oracle: warn-only at confidence 0.07 (a coin flip, 0.54 / 0.46). FINAL by leader verdict: warn-only, never blocks taking the integration window.

### Q8: Should the leader or `moai todo --auto` be steered by the quota state beyond a read-only status block? (DO-8)

Label: FOUNDER
Authority anchor: none
Why unresolved: `--auto` runs inside the Claude session whose quota would be measured and has no lane or backend reference today; steering it means a new printed surface and a way to know which lanes are live and on which backend.
Operator verdict: Oracle leaned toward steering (probability 0.66, confidence 0.32, LOW); the lane's `status block only` override was REJECTED by the leader. LEADER VERDICT: steering is in scope in its minimal form — with quota pressure on, `moai todo --auto` and the `moai factory status` leader surface recommend non-Claude lanes for the next card; Claude lanes get the status-3 hold; with no live non-Claude lane only a warning is printed; no card is dispatched or reassigned; the card-class question stays out of scope.

### Q9: Should a high reading older than the maximum age, whose window has not yet reset, stay binding until the reset? (DO-9)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: keeping it binding assumes usage never falls inside a window; the Claude Code documentation calls the five-hour window rolling, and no measurement of a usage reading falling before its reset exists.
Operator verdict: Resolved by decision oracle — treat as unknown (confidence 0.86).

### Q10: Which variable defines "a Claude lane" for the gate — the launch provider, the kanban backend, or both? (DO-10)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the token a Claude factory lane actually carries in each variable was read from code but not observed on a live lane, and the kanban backend variable's stamp site sits in kanban code that is being removed.
Operator verdict: Resolved by decision oracle — launch provider first, kanban backend as fallback (confidence 0.96).

### Q11: Should the windows ride the existing per-session record, a separate shared account file, or the existing home-level usage cache? (DO-11)

Label: FOUNDER
Authority anchor: none
Why unresolved: this is the data-model decision (record schema 2 → 3 versus a second writer versus reuse of a probe-fed cache); it is the decision most likely to change later and no committed document decides it.
Operator verdict: Resolved by decision oracle — per-session record (confidence 1.00).

### Q12: Where should the steering read a lane's backend from — the lane's session record, or a registry column written when the lane claims its label? (DO-12)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the registry's `backend` column exists but is never written (the claim inserts omit it), while the web console already joins registry pid to session to session record; whether Codex lane sessions write a session record at all, and whether that record store survives the removal of kanban mode, were not observed. Writing the column at claim time changes the claim path and the launchers.
Operator verdict: Resolved by decision oracle — write the backend at claim (probability 0.78, confidence 0.55, above the 0.5 gate). Verified before applying: every claim site holds its backend as a launcher-local value, no kanban-mode code is needed, the column needs no migration, and old rows read as empty (unknown). The only kanban-package file edited is the factory registry's claim cluster. The session-record join is dropped.
