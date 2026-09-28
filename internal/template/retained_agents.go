package template

// retainedAgentRoster is the canonical retained-agent roster: every retained
// MoAI-custom agent plus the Anthropic built-in Explore, in display order. It
// carries names only — no model and no effort — because subagents inherit the
// main session's model and effort, so the roster is membership, not policy.
//
// @MX:ANCHOR: [AUTO] canonical retained-agent roster; internal/harness/rosterguard compares every registered roster site against it
// @MX:REASON: a second hand-typed roster is how roster drift starts; every roster-listing surface is asserted against this one literal
var retainedAgentRoster = []string{
	"manager-spec",
	"plan-auditor",
	"sync-auditor",
	"manager-develop",
	"super-advisor",
	"mission-governor",
	"manager-design",
	"manager-lead",
	"builder-harness",
	"e2e-tester",
	"manager-docs",
	"manager-git",
	"Explore",
}

// RetainedAgents returns a defensive copy of the canonical retained-agent
// roster in display order.
func RetainedAgents() []string {
	out := make([]string, len(retainedAgentRoster))
	copy(out, retainedAgentRoster)
	return out
}
