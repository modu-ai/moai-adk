package delegationmap

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// Options configures one analyzer run.
type Options struct {
	// LedgerPath is the finalized routing ledger to read.
	LedgerPath string

	// MapPath is the delegation map to resolve designations from. It is opened
	// read-only and never written.
	MapPath string

	// MaxLedgerBytes overrides the declared size bound. Zero selects the
	// MaxLedgerBytes constant. The override exists so a test can exercise the
	// refusal path without committing a fixture large enough to trip the real
	// bound.
	MaxLedgerBytes int64
}

// Analyze reads the ledger, aggregates delegation patterns per subcommand, and
// produces findings against the delegation map.
//
// It returns an error only for a genuine read failure. Every degraded input —
// an absent, empty, oversized, or wholly malformed ledger, or an absent map —
// yields an empty finding list with a machine-readable reason and a nil error
// (REQ-HLA-014). The distinction matters: this runs inside a fail-open harness
// path, where an error would be noise but a silent empty result would be a lie.
func Analyze(o Options) (Result, error) {
	res := Result{Findings: []Finding{}, Stats: []SubcommandStat{}, Reason: ReasonOK}

	maxBytes := o.MaxLedgerBytes
	if maxBytes <= 0 {
		maxBytes = MaxLedgerBytes
	}

	info, err := os.Stat(o.LedgerPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		res.Reason = ReasonLedgerAbsent
		return res, nil
	case err != nil:
		return res, fmt.Errorf("delegationmap: stat ledger: %w", err)
	case info.Size() == 0:
		res.Reason = ReasonLedgerEmpty
		return res, nil
	case info.Size() > maxBytes:
		// Declined, not truncated: routing.Reader materializes the whole row
		// set, and a partial read would silently skew every support ratio.
		res.Reason = ReasonLedgerOversize
		return res, nil
	}

	rows, malformed, err := routing.NewReader(o.LedgerPath).Read(routing.Filter{})
	if err != nil {
		return res, fmt.Errorf("delegationmap: read ledger: %w", err)
	}
	res.MalformedLines = malformed

	if len(rows) == 0 {
		if malformed > 0 {
			res.Reason = ReasonAllLinesMalformed
		} else {
			res.Reason = ReasonLedgerEmpty
		}
		return res, nil
	}

	stats := aggregate(rows)
	res.Stats = sortedStats(stats)
	res.EvaluatedSubcommands = len(stats)
	res.LatestTS = latestTS(rows)

	designations, err := ReadDelegationMap(o.MapPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		res.Reason = ReasonDelegationMapAbsent
		return res, nil
	case err != nil:
		return res, err
	}

	res.Findings, res.Reason = findings(stats, designations)
	return res, nil
}

// findings applies the thresholds and the conditional-invocation exclusion,
// returning the ordered finding list and the result reason.
func findings(stats map[string]*SubcommandStat, designations map[string][]string) ([]Finding, string) {
	out := []Finding{}
	anyCleared := false

	for subcommand, s := range stats {
		if s.QualifyingRows < MinQualifyingRows {
			continue
		}
		anyCleared = true

		designated := make(map[string]struct{}, len(designations[subcommand]))
		for _, a := range designations[subcommand] {
			designated[a] = struct{}{}
		}

		// undesignated_agent — a retained-catalog agent clearing both
		// thresholds that the map does not designate. Only exact catalog
		// members are comparable: a non-catalog type, a harness specialist, or
		// a spawn name recorded in place of an agent type is a real
		// observation, but it is not evidence that the map omitted a
		// designation (REQ-HLA-004).
		for agent := range s.AgentCounts {
			if !IsRetainedAgent(agent) {
				continue
			}
			if _, ok := designated[agent]; ok {
				continue
			}
			ratio := s.SupportRatio(agent)
			if ratio < MinSupportRatio {
				continue
			}
			out = append(out, Finding{
				Kind:              KindUndesignatedAgent,
				Subcommand:        subcommand,
				Agent:             agent,
				ObservationCount:  s.AgentCounts[agent],
				SupportRatio:      ratio,
				QualifyingRows:    s.QualifyingRows,
				UnattributedShare: s.UnattributedEntries,
			})
		}

		// designated_never_spawned — a designation observed zero times.
		// Conditionally-invoked designations are excluded, because reasoning
		// from absence would otherwise report the exact behavior the rules
		// prescribe as a defect (REQ-HLA-008).
		for _, agent := range designations[subcommand] {
			if IsConditionallyInvoked(agent) {
				continue
			}
			if s.AgentCounts[agent] > 0 {
				continue
			}
			out = append(out, Finding{
				Kind:              KindDesignatedNeverSpawned,
				Subcommand:        subcommand,
				Agent:             agent,
				ObservationCount:  0,
				SupportRatio:      0,
				QualifyingRows:    s.QualifyingRows,
				UnattributedShare: s.UnattributedEntries,
			})
		}
	}

	sortFindings(out)

	switch {
	case len(out) > 0:
		return out, ReasonOK
	case !anyCleared:
		return out, ReasonBelowMinRows
	default:
		return out, ReasonNoFindings
	}
}

// sortFindings orders findings deterministically so two runs over the same
// input produce the same candidate set in the same order (REQ-HLA-013).
func sortFindings(f []Finding) {
	sort.Slice(f, func(i, j int) bool {
		if f[i].Subcommand != f[j].Subcommand {
			return f[i].Subcommand < f[j].Subcommand
		}
		if f[i].Kind != f[j].Kind {
			return f[i].Kind < f[j].Kind
		}
		return f[i].Agent < f[j].Agent
	})
}

// latestTS returns the newest row timestamp as recorded, or the empty string
// when no row carries one. It is the deterministic stand-in for a wall clock:
// a proposal stamped with time.Now() would differ between two runs over
// identical input, which REQ-HLA-013 forbids.
func latestTS(rows []routing.Row) string {
	var latest string
	for _, r := range rows {
		if r.TS > latest { // RFC3339 UTC sorts lexicographically
			latest = r.TS
		}
	}
	return latest
}
