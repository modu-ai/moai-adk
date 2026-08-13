// Package delegationmap is the L2 analyzer: it reads finalized routing-ledger
// rows, aggregates the delegation patterns actually observed per /moai
// subcommand, and emits proposals whose content is a concrete amendment to the
// delegation map under .moai/config/sections/ (mapreader.go names the file).
//
// The map itself is NEVER written here. Three independent surfaces agree that
// it requires human approval — the LSEL frozen allowlist
// (.claude/lsel/frozen-allowlist.json carries ^\.moai/config/sections/), the
// map's own `auto_apply: false` key, and the Tier-4 approval gate — so this
// package reads the map to resolve designations and stops there. Applying a
// proposal is deliberately out of scope (spec.md §G).
//
// HARD subagent boundary: this package and its CLI surface MUST NOT invoke
// AskUserQuestion. The orchestrator owns the approval gate; a proposal emitted
// here is an input to that gate, never a substitute for it.
//
// SPEC: SPEC-HARNESS-LEARNING-EVO-002 (REQ-HLA-001..016).
package delegationmap

import "github.com/modu-ai/moai-adk/internal/harness/routing"

// MinQualifyingRows is the minimum number of finalized qualifying rows a
// subcommand must carry before any proposal is emitted for it. Below it, a
// single stray row could produce a proposal.
//
// The value mirrors the existing proposalgen precedent of one declared constant
// rather than config-driven tuning. It is a guess, not a measurement — the
// ledger held 4 rows when this was written — and retuning it against real
// observations is deferred to the review window in spec.md §E.
const MinQualifyingRows = 5

// MinSupportRatio is the minimum share of a subcommand's qualifying rows in
// which an agent must appear before it can produce an undesignated-agent
// proposal. Same provenance and same caveat as MinQualifyingRows.
const MinSupportRatio = 0.60

// MaxLedgerBytes bounds the ledger the analyzer will read. routing.Reader.Read
// materializes the whole filtered row set — no streaming variant exists, and
// adding one would place a change to the producer's package inside this SPEC
// (plan.md §E D1) — so the exposure is capped rather than removed. A ledger
// above the bound is declined with a machine-readable reason, not truncated:
// a partial read would silently skew every support ratio.
const MaxLedgerBytes = 32 << 20 // 32 MiB

// Kind is the closed proposal-kind enum. There are exactly two members and
// REQ-HLA-007 admits no third: every finding either names an agent the map
// failed to designate, or a designation the pipeline never exercised.
type Kind string

const (
	// KindUndesignatedAgent names a retained-catalog agent that clears both
	// thresholds for a subcommand but is absent from its designated list.
	KindUndesignatedAgent Kind = "undesignated_agent"

	// KindDesignatedNeverSpawned names a designated agent observed in zero
	// qualifying rows for a subcommand that itself clears the row threshold.
	KindDesignatedNeverSpawned Kind = "designated_never_spawned"
)

// ValidKind reports whether k is a member of the closed enum.
func ValidKind(k Kind) bool {
	switch k {
	case KindUndesignatedAgent, KindDesignatedNeverSpawned:
		return true
	default:
		return false
	}
}

// retainedCatalog is the membership test for REQ-HLA-004: only a value that
// exactly matches one of these names is comparable against the delegation map.
//
// Source: CLAUDE.md §4 (the 12 retained agents). It is a declared constant
// rather than a derivation from the map's own agent lists, because
// deriving membership from the map would make the discrimination circular — an
// agent absent from every designation could never be found undesignated, which
// is precisely the finding this analyzer exists to produce (plan.md §E D2).
//
// The cost is a snapshot that can go stale against a catalog change; this
// comment is where that change lands.
var retainedCatalog = map[string]struct{}{
	"manager-spec":    {},
	"manager-develop": {},
	"manager-docs":    {},
	"manager-git":     {},
	"manager-design":  {},
	"manager-kanban":    {},
	"plan-auditor":    {},
	"sync-auditor":    {},
	"builder-harness": {},
	"super-advisor":   {},
	"e2e-tester":      {},
	"Explore":         {},
}

// IsRetainedAgent reports whether name exactly matches a retained-catalog agent.
//
// Everything else — a non-catalog agent type (general-purpose,
// workflow-subagent), a user-owned harness specialist (hns-*), or a spawn NAME
// recorded in place of the agent type — is a real observation but is not
// evidence that the map omitted a designation.
func IsRetainedAgent(name string) bool {
	_, ok := retainedCatalog[name]
	return ok
}

// ConditionalExclusion names one designated agent whose invocation is
// conditional by rule, together with the rule that makes it so.
//
// The citation is not decoration. This exclusion suppresses real findings by
// design, so an entry added without a rule behind it turns a justified carve-out
// into an allowlist that hides genuine drift (plan.md §G AP-6).
type ConditionalExclusion struct {
	Agent        string
	RuleCitation string
}

// ConditionalExclusions is the declared exclusion set for
// designated-never-spawned findings (REQ-HLA-008). Both members are legitimately
// absent from most runs, so without this set the analyzer would report the
// exact behavior the rules prescribe as a defect.
var ConditionalExclusions = []ConditionalExclusion{
	{
		Agent:        "sync-auditor",
		RuleCitation: "CLAUDE.md §6 + .claude/rules/moai/workflow/spec-workflow.md § Mode Dispatch — runs at harness level `thorough` only",
	},
	{
		Agent:        "plan-auditor",
		RuleCitation: ".claude/rules/moai/workflow/spec-workflow.md § Plan Audit Gate skip policy — skipped when the cached verdict is PASS at or above the tier threshold with an unchanged artifact hash",
	},
}

// IsConditionallyInvoked reports whether agent is excluded from
// designated-never-spawned findings.
func IsConditionallyInvoked(agent string) bool {
	for _, e := range ConditionalExclusions {
		if e.Agent == agent {
			return true
		}
	}
	return false
}

// SubcommandStat is the per-subcommand aggregate. It is the only state the
// analyzer retains across rows (REQ-HLA-001).
type SubcommandStat struct {
	Subcommand string `json:"subcommand"`

	// QualifyingRows counts rows whose outcome is success or fail. Only these
	// contribute delegations (REQ-HLA-003).
	QualifyingRows int `json:"qualifying_rows"`

	// AgentCounts maps an observed agent value to the number of QUALIFYING ROWS
	// in which it appeared. It is row-presence, not entry count: an agent
	// delegated twice within one row counts once, so a support ratio stays a
	// ratio and cannot exceed 1.
	AgentCounts map[string]int `json:"agent_counts"`

	// NonCatalogAgents lists, sorted, the observed values that are real agents
	// but not retained-catalog members. They are recorded rather than dropped —
	// the classification is what suppresses their findings, and a reviewer needs
	// to see what was classified (REQ-HLA-004).
	NonCatalogAgents []string `json:"non_catalog_agents"`

	// UnattributedEntries counts delegation ENTRIES carrying the absent-identity
	// marker. Entry count rather than row-presence, because the share a reviewer
	// needs is "how much of the delegation population is unattributed"
	// (REQ-HLA-005).
	UnattributedEntries int `json:"unattributed_entries"`

	// RerouteRows and AbortRows are routing-instability context. They never
	// produce a proposal by themselves (REQ-HLA-003).
	RerouteRows int `json:"reroute_rows"`
	AbortRows   int `json:"abort_rows"`

	// EmptyDelegationRows counts qualifying rows that carried no delegation at
	// all. It exists to make one specific degradation visible rather than
	// silent: per plan.md §B6 (open as SPEC-HARNESS-LEARNING-EVO-001 R8), a
	// session's delegations can land on a different row than its outcome, so the
	// row counted as qualifying may hold none of them. That inflates this count
	// while deflating every support ratio, and both errors push toward
	// SUPPRESSING real patterns rather than inventing false ones — a failure
	// that produces no error and no odd-looking finding. A reviewer comparing
	// this count against QualifyingRows can see the split directly instead of
	// having to infer it from findings that never appeared.
	EmptyDelegationRows int `json:"empty_delegation_rows"`
}

// SupportRatio returns the share of qualifying rows in which agent appeared.
// Zero qualifying rows yields 0 rather than a division by zero.
func (s *SubcommandStat) SupportRatio(agent string) float64 {
	if s.QualifyingRows == 0 {
		return 0
	}
	return float64(s.AgentCounts[agent]) / float64(s.QualifyingRows)
}

// Finding is one proposed delegation-map amendment, carrying the evidence a
// reviewer needs to judge it without re-reading the ledger (REQ-HLA-012).
//
// There is deliberately no skill field. The schema-v1 ledger row records no
// injected-skill data, so a skill amendment could not be grounded in observation
// (spec.md §G); declaring the field and leaving it empty would invite exactly
// that ungrounded proposal.
type Finding struct {
	Kind             Kind    `json:"kind"`
	Subcommand       string  `json:"subcommand"`
	Agent            string  `json:"agent"`
	ObservationCount int     `json:"observation_count"`
	SupportRatio     float64 `json:"support_ratio"`
	QualifyingRows   int     `json:"qualifying_rows"`

	// UnattributedShare is the subcommand's unattributed delegation-entry count,
	// carried onto every finding so a reviewer can judge how much of the
	// population the finding rests on.
	UnattributedShare int `json:"unattributed_share"`
}

// Reason values are the machine-readable result diagnostics. Every non-"ok"
// value means an empty finding list, and each names WHY — an empty result with
// no reason would be indistinguishable from a clean run.
const (
	ReasonOK                  = "ok"
	ReasonLedgerAbsent        = "ledger-absent"
	ReasonLedgerEmpty         = "ledger-empty"
	ReasonLedgerOversize      = "ledger-oversize"
	ReasonAllLinesMalformed   = "all-lines-malformed"
	ReasonBelowMinRows        = "below-min-qualifying-rows"
	ReasonDelegationMapAbsent = "delegation-map-absent"
	ReasonNoFindings          = "no-findings"
)

// Result is the analyzer's structured outcome, suitable for JSON emission.
type Result struct {
	// Findings is empty (non-nil) whenever Reason is not ReasonOK.
	Findings []Finding `json:"findings"`

	// Stats carries the per-subcommand aggregate, sorted by subcommand.
	Stats []SubcommandStat `json:"stats"`

	// Reason is one of the Reason* constants above.
	Reason string `json:"reason"`

	// MalformedLines is the count of ledger lines that failed to unmarshal.
	MalformedLines int `json:"malformed_lines"`

	// EvaluatedSubcommands is the number of distinct subcommands observed.
	EvaluatedSubcommands int `json:"evaluated_subcommands"`

	// LatestTS is the newest row timestamp observed, in RFC3339. It is the
	// deterministic source for a proposal's source_ts: a wall-clock stamp would
	// make two runs over identical input produce different drafts, which
	// REQ-HLA-013 forbids.
	LatestTS string `json:"latest_ts"`
}

// isQualifying reports whether a row's outcome admits it to the aggregation.
// reroute and abort are counted separately and contribute no delegations.
func isQualifying(o routing.Outcome) bool {
	return o == routing.OutcomeSuccess || o == routing.OutcomeFail
}
