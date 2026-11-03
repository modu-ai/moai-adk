package graph

// Event-aware verdict classification (card t2001, user request R208, CI policy
// common to PRs outside the card's own scope).
//
// The verdict of a layer (fresh | stale | absent) is a fact about the tree and
// is never rewritten here. What this file adds is a second, separate field —
// VerdictClass — that says whether a non-fresh layer should FAIL the check
// (block) or only be announced (warn). The one case that is announced instead
// of failed is a codemaps count a pull request merely inherited from its base:
// the change left the codemaps stamp and body byte-identical to the first
// parent, the base ALONE (cumulative count minus the change's own contribution)
// already meets the threshold, and every other layer is fresh. A count the
// change took over the line (crossed-by-change), one it put over the line by
// itself (originated), and one whose stamp or body moved or could not be proven
// unmoved (anchor-moved) all block. Nothing about the threshold, the metric, or
// any other layer changes, and a stamp-only refresh is not offered as a remedy.

const (
	// VerdictClassBlock marks a non-fresh layer that fails the check (exit 1).
	VerdictClassBlock = "block"
	// VerdictClassWarn marks a non-fresh layer that is announced but does not
	// fail the check (exit 0 plus a warning).
	VerdictClassWarn = "warn"
)

const (
	// EventPullRequest is the CI event on which an inherited codemaps drift may
	// be downgraded to a warning.
	EventPullRequest = "pull_request"
	// EventPush is the CI event (main, release/*) on which every non-fresh layer
	// blocks: release integrity is not negotiable.
	EventPush = "push"
)

// Where a stale codemaps count came from, stated as a fact next to the verdict.
// Base alone = Value - Contribution (the described-worthy files the change did
// not touch).
const (
	// DriftInherited: the base alone is already at or over the threshold.
	DriftInherited = "inherited"
	// DriftCrossedByChange: the base is under the threshold and this change took
	// the count over it.
	DriftCrossedByChange = "crossed-by-change"
	// DriftOriginated: this change alone is at or over the threshold.
	DriftOriginated = "originated"
	// DriftAnchorMoved: this change touched the codemaps body or its stamp (or
	// that could not be ruled out), so the count's measurement window is not the
	// base's and the count says nothing reliable about what the base carried.
	DriftAnchorMoved = "anchor-moved"
)

const (
	// InheritedDriftWarn lets a pull_request announce an inherited codemaps
	// drift instead of failing on it. It is the default.
	InheritedDriftWarn = "warn"
	// InheritedDriftBlock keeps every non-fresh layer blocking on every event —
	// the behavior before the policy existed.
	InheritedDriftBlock = "block"
)

// ClassifyVerdicts sets VerdictClass on every non-fresh layer. Every non-fresh
// layer blocks, except one: the codemaps layer when ALL of the following hold —
//
//   - the event is a pull request and the policy is not "block";
//   - the codemaps verdict is stale (an absent or unjudgeable layer blocks);
//   - the change's own contribution was measured (nil means unmeasured, which
//     blocks — a fabricated 0 would read as the inheriting signature) and the
//     BASE alone (Value - Contribution) is already at or over the threshold,
//     i.e. the origin is "inherited";
//   - the change left the codemaps stamp and body byte-identical to the first
//     parent (CodemapsInputsStable is true; a symlink, an ignored file, or any
//     input whose sameness cannot be proven leaves it nil or false);
//   - every other layer is fresh.
//
// A push event, an unknown or empty event, a count this change took over the
// line (crossed-by-change), one this change alone put over it (originated),
// and one whose measurement window moved or could not be proven unmoved
// (anchor-moved) all block.
func (r *CheckResult) ClassifyVerdicts(event, inheritedPolicy string, th Thresholds) {
	nonFresh := 0
	for i := range r.Layers {
		if r.Layers[i].Verdict != VerdictFresh {
			r.Layers[i].VerdictClass = VerdictClassBlock
			nonFresh++
		}
		// The origin is a fact about the count, stated whatever the event or policy.
		if l := &r.Layers[i]; l.Layer == LayerCodemaps && l.Verdict == VerdictStale {
			l.DriftOrigin = driftOrigin(*l, th.CodemapsChangedFiles)
		}
	}
	if event != EventPullRequest || inheritedPolicy == InheritedDriftBlock || nonFresh != 1 {
		return
	}
	for i := range r.Layers {
		l := &r.Layers[i]
		if l.Layer == LayerCodemaps && l.Verdict == VerdictStale && l.DriftOrigin == DriftInherited {
			l.VerdictClass = VerdictClassWarn
		}
	}
}

// driftOrigin classifies a stale codemaps count by where it came from. The base
// alone is Value - Contribution, but that subtraction is a statement about the
// base ONLY while the change left the codemaps stamp and body alone: Value is
// measured from an anchor derived from the tree under judgment, so a change that
// rewinds the stamp widens the window over files the base had already covered,
// and Contribution (.go files against HEAD^1) never sees that. Hence the order:
// a change over the line by itself is originated whatever the window did; a
// window that moved, or whose stability was not measured, is anchor-moved and
// never inherited; only then does the arithmetic decide. An unmeasured
// contribution states no origin and so never warns.
func driftOrigin(l LayerReport, threshold int) string {
	if l.Contribution == nil {
		return ""
	}
	switch {
	case *l.Contribution >= threshold:
		return DriftOriginated
	case l.CodemapsInputsStable == nil || !*l.CodemapsInputsStable:
		return DriftAnchorMoved
	case l.Value-*l.Contribution >= threshold:
		return DriftInherited
	default:
		return DriftCrossedByChange
	}
}

// Blocking reports whether the check must fail: some non-fresh layer is not
// classified warn. A layer nobody classified blocks, so a caller that never
// calls ClassifyVerdicts keeps the fail-closed behavior of Failed().
func (r CheckResult) Blocking() bool {
	for _, l := range r.Layers {
		if l.Verdict != VerdictFresh && l.VerdictClass != VerdictClassWarn {
			return true
		}
	}
	return false
}

// Warnings lists the layers announced instead of failed.
func (r CheckResult) Warnings() []LayerReport {
	var out []LayerReport
	for _, l := range r.Layers {
		if l.Verdict != VerdictFresh && l.VerdictClass == VerdictClassWarn {
			out = append(out, l)
		}
	}
	return out
}
