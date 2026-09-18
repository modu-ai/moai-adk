// Package rosterguard is the declared-axis guard over agent-roster listings.
//
// Roster listings across this tree go stale because an addition is propagated
// FORWARD ONLY, by whoever made it, to the sites they happened to know about;
// earlier gaps are never backfilled. Two measured instances:
//
//   - `mission-governor` was registered in internal/config/profile.go by its own
//     creation commit 5ec516165, but `manager-lead` (which arrived via the
//     rename 310d75dd2) never was.
//   - internal/template/profile_matrix.go records the SAME agent falling
//     through the SAME crack once before: "manager-lead was absent from this
//     list until t205 and therefore resolved to the unmapped-agent `inherit`
//     sentinel".
//
// # Why this guard compares SETS and never counts
//
// At the time this package was written the agent-definition file count was 12
// (11 MoAI-custom + mission-governor; the built-in Explore has no file) and
// delegationmap.retainedCatalog was ALSO 12 (11 MoAI-custom + Explore; no
// mission-governor). Two different 12s whose intersection is 11. A guard that
// compared numbers would have passed vacuously on exactly the drift it exists
// to catch. Every membership assertion here is therefore a set comparison, and
// a count assertion is only ever an ADDITIONAL claim on a named axis — never
// the membership evidence.
//
// # Why there is more than one axis
//
// Not every roster listing claims the same thing. "12 agent definitions" is a
// TRUE statement about the file count and a FALSE statement about the retained
// roster; a guard that conflated the two would break a correct site. Each
// registered site therefore declares WHICH axis it speaks about, and the guard
// asserts only what that axis supports.
package rosterguard

// Axis names one measurable roster population. A site declares the axis it
// speaks about; the guard asserts only what that axis supports.
type Axis string

const (
	// AxisRetainedRoster is the CLAUDE.md section 4 retained roster — every
	// retained agent including the Anthropic built-in Explore. The canonical
	// membership is template.ProfileMatrixAgents().
	AxisRetainedRoster Axis = "retained-roster"

	// AxisDefinitionFiles is the population of .claude/agents/moai/*.md
	// definition files. It is strictly smaller than AxisRetainedRoster because
	// the built-in Explore has no definition file. A site declaring a count on
	// THIS axis is correct when it states the file count, and conflating it
	// with AxisRetainedRoster would break a correct site.
	AxisDefinitionFiles Axis = "definition-files"

	// AxisSubsetByDesign marks a site that legitimately lists only a subset of
	// the roster — a per-subcommand designation list, for instance. Such a site
	// gets NO membership assertion, because asserting complete membership on it
	// would be wrong: manager-design, manager-lead, super-advisor and
	// mission-governor are all legitimately absent from every delegation
	// designation.
	AxisSubsetByDesign Axis = "subset-by-design"
)

// ClaimKind is a bitset of what a site claims. A site may declare a count,
// enumerate a membership, or do both — the two are independent, and a site that
// declares a stale count while enumerating a correct membership (or the
// reverse) is a real and observed shape.
type ClaimKind uint8

const (
	// ClaimCount marks a site that states HOW MANY agents its axis has.
	ClaimCount ClaimKind = 1 << iota
	// ClaimMembership marks a site that enumerates WHICH agents its axis has.
	ClaimMembership
)

// Has reports whether k includes want.
func (k ClaimKind) Has(want ClaimKind) bool { return k&want != 0 }

// Staleness records a site whose claim is KNOWN to disagree with its axis at
// the time of registration, together with the exact shape of the disagreement.
//
// It is deliberately NOT a mute. The guard asserts that the observed
// disagreement matches the DECLARED disagreement exactly, so the marker is
// self-expiring: when the site is repaired the assertion fails with an
// instruction to delete the marker, and when the site drifts FURTHER the
// assertion fails with the new gap. A marker that merely silenced the site
// would decay into a permanent exemption, which is the failure mode that
// produced the drift this package guards.
type Staleness struct {
	// Reason is the one-line why, stated so a reader can judge it.
	Reason string
	// FollowUp names the card that owns the repair, or a placeholder.
	FollowUp string
	// MissingNames are the axis members the site omits. Required non-empty
	// when the site carries ClaimMembership — a "stale" site that omits
	// nothing is not stale, and the guard says so.
	MissingNames []string
	// ExtraNames are names the site lists that the axis does not carry.
	ExtraNames []string
	// DeclaredCount is the stale number the site currently states. Required
	// (non-zero) when the site carries ClaimCount.
	DeclaredCount int
}

// Site is one registered roster-listing location.
//
// A path may carry SEVERAL sites: a file can declare a count about one axis in
// its prose while enumerating a subset on another axis in its body, and
// collapsing those into one row would force one of the two claims to go
// unasserted.
type Site struct {
	// ID is the stable identifier used in failure messages and in the
	// sweep's registered-path set.
	ID string
	// Path is repo-root-relative, forward-slashed.
	Path string
	// Axis is the axis THIS claim speaks about.
	Axis Axis
	// Claims is what this row asserts.
	Claims ClaimKind
	// BlockStart is a literal substring identifying the first line of the
	// region the claim occupies. When empty the block is the WHOLE file —
	// correct only where the file has no roster mention outside its listing,
	// which is a property of the file and is recorded per row in Note.
	BlockStart string
	// BlockEnd is a literal substring identifying the terminating line. When
	// empty (and BlockStart is not) the block is the single line containing
	// BlockStart — the right scope for a prose sentence that enumerates
	// inline, where taking the whole file would let an unrelated later mention
	// satisfy the assertion.
	BlockEnd string
	// CountPattern is a regexp with exactly one capture group holding the
	// declared number. It MUST match the body exactly once: a pattern matching
	// twice is ambiguous about which claim is being asserted, and a pattern
	// matching zero times has drifted off its anchor. Required when Claims has
	// ClaimCount.
	//
	// A regexp rather than a line anchor because several count claims share a
	// line with another integer — delegationmap/types.go cites "CLAUDE.md §4
	// (the 12 retained agents)", where a first-integer-on-the-line rule reads
	// the section number 4 as the roster count.
	CountPattern string
	// KnownStale, when non-nil, records a measured disagreement (see
	// Staleness). Repair is out of scope for the card that created this
	// package; the marker keeps the staleness enumerable instead of silent.
	KnownStale *Staleness
	// Note carries any context a reader needs to judge the row.
	Note string
}
