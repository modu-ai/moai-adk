//! #3042: the agent workflow says what happens to a lease that ended. Card C-53 S5 moved the
//! round to the card ledger: `card renew` needs a live lease (`moai card renew --help`), so an
//! ended lease goes to the leader; the `claim-acquire` behaviour stays pinned in
//! `tests/claim_cli.rs`.

// The file stays because `skills/moai/references/rule-history.md` names it (SPEC-20261122-next-flow-skills-agents D-11).

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_agent_workflow_says_what_the_command_does_with_an_expired_claim` pinned the lane's claim step, which left with the route.
