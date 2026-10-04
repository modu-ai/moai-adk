# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`; `interview.recommendation_mode: pull`). Rows state what is unresolved and why; no row carries a recommendation or a preferred answer. `Operator verdict:` lines are empty at authoring. The SPEC states its working choice for each in `spec.md` §C so run-phase is not blocked; a verdict here would confirm or replace that choice.

### Q1: Which spawn names, beyond `harness-learner`, count as self-improvement identities?

Label: EVIDENCE-NEEDED
Authority anchor: (none — the data does not exist: no agent definition named `harness-learner` exists under `.claude/agents/`, the learner skill runs in the orchestrator session where `agent_type` is empty, and no observation of which `agent_type` values autonomous flows carry has been taken.)
Why unresolved: the guard keys on a spawn-name equality, so its reach depends entirely on which names real self-improvement flows spawn under. Candidates exist (`builder-harness`, any curator or applier spawn), but adding one is a judgment about whether that agent is a self-improver, and the observation that would ground it has not been made.
Operator verdict:

### Q2: Should a present-but-invalid manifest fail closed for the identity set, or fail open like the opt-in guard family?

Label: FOUNDER
Authority anchor: (none — no committed document or operator setting covers fail-closed versus fail-open for this guard. The branch-guard rule's fail-open norm governs that guard's own uncertainty handling, not this question as written.)
Why unresolved: fail closed means a corrupt manifest stops autonomous self-improvement writes until a human repairs or restores the file, loudly; fail open means the guard disables itself silently on a parse error. The blast radius of closing is confined to the identity set, but it is a behaviour the operator may weigh differently.
Operator verdict:

### Q3: Should the PreToolUse matcher be widened to deliver `MultiEdit` and `NotebookEdit` to the guard?

Label: FOUNDER
Authority anchor: (none — an existing template test pins the matcher string `Write|Edit|Bash` with the message "must not be renamed or widened"; that records a constraint set by another change, not a decision about this question.)
Why unresolved: without widening, those two tools never reach any PreToolUse guard, so the protected zone cannot cover them. Widening changes a pinned shared surface and its tests.
Operator verdict:

### Q4: Who keeps the manifest in step with the apparatus files the later audit-overhaul card will add or move?

Label: FOUNDER
Authority anchor: (none — no committed process assigns manifest upkeep.)
Why unresolved: the liveness check catches an entry that stops matching anything, not a new apparatus file nobody listed. Whether that card carries its own manifest edit, or a separate sweep follows it, is a process choice for the operator.
Operator verdict:

### Q5: Should the four existing frozen lists be unified into one source derived from the manifest, and should `HARNESS_FROZEN_CONFIG_VIOLATION` be wired or retired?

Label: FOUNDER
Authority anchor: (none — no committed decision covers consolidating the hook, safety-pipeline, meta-harness and LSEL lists, or the unused config sentinel.)
Why unresolved: the SPEC adds a drift test instead of unifying, to stay inside the stated scope; unification touches three Go packages and a shell script and would change their tests. The unused sentinel is a pre-existing defect that the manifest approach routes around.
Operator verdict:

### Q7: Keep the JSONL audit log (REQ-SIPZ-014), or drop it to shrink scope?

Label: FOUNDER
Authority anchor: (none — the card names declaration, blocking and human routing; it does not name an audit log, and no committed policy requires one for this guard.)
Why unresolved: the plan audit named the audit log as the first candidate to drop at the Tier M requirement ceiling. It is kept in this draft because two other requirements lean on it: the degraded-manifest disclosure (an absent manifest allows through the compiled floor, and the audit row is the only place that state is recorded) and the human route (a denial reaches the subagent, not the operator, so the row is the one record that does not depend on the subagent behaving — spec §F G8). Dropping it removes the log, the degraded-state record and one acceptance criterion (AC-SIPZ-008), and leaves the absent-manifest state undisclosed outside the CI check.
Operator verdict:

### Q6: Is manifest membership an adequate reading of "routed to a human regardless of zone", or is content-level recognition of removal, lowering and deletion wanted outside the manifest?

Label: FOUNDER
Authority anchor: (none — the card text states the intent; no committed document fixes the recognition mechanism.)
Why unresolved: the SPEC reads "regardless of zone" as manifest-over-tag precedence and routes any modification of a listed path, because direction and content cannot be judged soundly by a hook. A literal reading would also catch a removed check in an unlisted, evolvable file, which needs a content classifier the SPEC deliberately does not build.
Operator verdict:
