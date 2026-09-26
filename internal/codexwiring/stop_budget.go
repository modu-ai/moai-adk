package codexwiring

import "time"

// Codex Stop-chain budget declarations (SPEC-DUAL-HARNESS-HOOK-PARITY-001
// design §D3.3, §D3.5, §D3.8; REQ-HPR-018, REQ-HPR-019).
//
// On Codex the eight Claude Stop members run one after another inside the
// single rendered Stop handler, so they share its timeout T_stop. Each member
// gets an internal budget — the deadline the chain runner sets for it, not a
// host timeout. A member whose real work cannot fit (the sync gate, the codex
// review) runs outside the hook and leaves a receipt (internal/verify
// receipt.go); in the hook it only evaluates its self-gates and compares the
// receipt, so its budget here is that compare budget.
//
// Every figure below is a DECLARATION, not a measurement. The per-member
// budgets and StopChainOverhead are proposals that the M2d timing leg of
// AC-HPR-016 validates against observed costs on the golden fixtures;
// StopUnmeasuredCap and StopCapStateDir are proposals that M2d finalizes.
// Nothing here is evidence that a member fits its budget.

// StopPlacement says where a Stop member's work runs on Codex.
type StopPlacement string

const (
	// StopPlacementInHook: the member runs inside the Codex Stop handler.
	StopPlacementInHook StopPlacement = "in-hook"
	// StopPlacementReceipt: the member's check runs out of hook and the
	// handler compares its receipt (design §D3.2 receipt option).
	StopPlacementReceipt StopPlacement = "receipt"
)

// StopClass is the member's decision class (design §D3.3, last column).
type StopClass string

const (
	// StopClassAdvisory: a failure is recorded and the chain continues; never
	// recorded as passed (REQ-HPR-004).
	StopClassAdvisory StopClass = "advisory"
	// StopClassRequiredGate: once its self-gates hold, a missing or stale
	// receipt continues the turn (unmeasured) and never allows below the
	// StopUnmeasuredCap.
	StopClassRequiredGate StopClass = "required-gate"
	// StopClassGoal: lookup-only goal evaluation; a receipt miss is unmeasured.
	StopClassGoal StopClass = "goal"
	// StopClassFailOpenOnMissing: a missing persisted result allows, with a
	// discard record (member 7 only).
	StopClassFailOpenOnMissing StopClass = "fail-open-on-missing"
)

// StopMember is one declared member of the Codex Stop chain.
type StopMember struct {
	// Number is the member's position in the Claude Stop array (1-based).
	Number int
	// Name is the Claude handler the member corresponds to.
	Name      string
	Placement StopPlacement
	Class     StopClass
	// Budget is the member's internal deadline. For a receipt member it is
	// the self-gate + compare budget only; its out-of-hook cost is not here.
	Budget time.Duration
	// UncutBudget is a bounded step exempt from the Budget cut-off but still
	// counted in the aggregate: member 1's factory-continuation step, bounded
	// by its own 200 ms inspection deadline (design §D3.3).
	UncutBudget time.Duration
}

// @MX:ANCHOR: [AUTO] Codex Stop-chain budget table — the aggregate the single Codex Stop handler must fit (design §D3.5)
// @MX:REASON: read by the AC-HPR-016 sum leg here and by the M2d chain runner and timing leg in internal/cli; raising one budget without lowering another breaks Σ + chain_overhead ≤ T_stop

// StopChainMembers declares all eight Claude Stop members in Claude order.
// Proposed budgets (design §D3.5): (2 + 0.2) + 0.5 + 2 + 0.5 + 0.5 + 0.5 +
// 0.5 + 0.5 = 7.2 s. Member 8 runs only when the hook opt-in is enabled; it is
// counted anyway so the aggregate holds for either render.
var StopChainMembers = []StopMember{
	{Number: 1, Name: "moai hook stop", Placement: StopPlacementInHook, Class: StopClassAdvisory,
		Budget: 2 * time.Second, UncutBudget: 200 * time.Millisecond},
	{Number: 2, Name: "sync-phase quality gate", Placement: StopPlacementReceipt, Class: StopClassRequiredGate,
		Budget: 500 * time.Millisecond},
	{Number: 3, Name: "moai hook stop-goal", Placement: StopPlacementInHook, Class: StopClassGoal,
		Budget: 2 * time.Second},
	{Number: 4, Name: "moai hook security-turn", Placement: StopPlacementInHook, Class: StopClassAdvisory,
		Budget: 500 * time.Millisecond},
	{Number: 5, Name: "moai hook security-commit", Placement: StopPlacementInHook, Class: StopClassAdvisory,
		Budget: 500 * time.Millisecond},
	{Number: 6, Name: "moai hook codex-review-gate", Placement: StopPlacementReceipt, Class: StopClassRequiredGate,
		Budget: 500 * time.Millisecond},
	{Number: 7, Name: "moai hook multi-review-gate", Placement: StopPlacementInHook, Class: StopClassFailOpenOnMissing,
		Budget: 500 * time.Millisecond},
	{Number: 8, Name: "moai hook harness-observe-stop", Placement: StopPlacementInHook, Class: StopClassAdvisory,
		Budget: 500 * time.Millisecond},
}

const (
	// StopChainOverhead is chain_overhead (design §D3.5): input parsing,
	// attribution, dedup, and output write for the whole chain. Proposal:
	// the full 2.8 s left under T_stop = 10 s by the 7.2 s of member budgets,
	// so the runner deadline T_stop − StopChainOverhead equals the member sum
	// and any budget raise must be traded against another member.
	StopChainOverhead = 2800 * time.Millisecond

	// StopTimeoutCodexMax is T_codex_max, the largest Stop timeout Codex has
	// been measured to honour (AC-HPR-021). Zero means NOT_RUN: the probe has
	// not run (operator decision Q5), so T_stop stays at the render constant.
	StopTimeoutCodexMax = time.Duration(0)

	// StopUnmeasuredCap is N of design §D3.8: after N consecutive unmeasured
	// continuations for the same gate, HEAD, and working-tree digest, the
	// Codex Stop chain allows the stop and records the gate unverified.
	// Proposal (default 3); M2d finalizes it.
	StopUnmeasuredCap = 3

	// StopCapStateDir holds the per-session counter file
	// (<StopCapStateDir>/<session-id>.json). Proposal; M2d finalizes the name.
	StopCapStateDir = ".moai/state/codex-stop-cap"
)
