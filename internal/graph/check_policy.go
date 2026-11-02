package graph

// Event-aware verdict classification (card t2001, user request R208, CI policy
// common to PRs outside the card's own scope).
//
// The verdict of a layer (fresh | stale | absent) is a fact about the tree and
// is never rewritten here. What this file adds is a second, separate field —
// VerdictClass — that says whether a non-fresh layer should FAIL the check
// (block) or only be announced (warn). The one case that is announced instead
// of failed is a codemaps count a pull request merely inherited from its base:
// the cumulative count is at or over the threshold, but the change under
// judgment contributed fewer described-worthy files than the threshold, and
// every other layer is fresh. Nothing about the threshold, the metric, or any
// other layer changes, and a stamp-only refresh is not offered as a remedy.

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
//     blocks — a fabricated 0 would read as the inheriting signature) and is
//     below the codemaps threshold;
//   - every other layer is fresh.
//
// A push event, an unknown or empty event, and a self-contribution at or over
// the threshold all block.
func (r *CheckResult) ClassifyVerdicts(event, inheritedPolicy string, th Thresholds) {
	nonFresh := 0
	for i := range r.Layers {
		if r.Layers[i].Verdict != VerdictFresh {
			r.Layers[i].VerdictClass = VerdictClassBlock
			nonFresh++
		}
	}
	if event != EventPullRequest || inheritedPolicy == InheritedDriftBlock || nonFresh != 1 {
		return
	}
	for i := range r.Layers {
		l := &r.Layers[i]
		if l.Layer != LayerCodemaps || l.Verdict != VerdictStale {
			continue
		}
		if l.Contribution == nil || *l.Contribution >= th.CodemapsChangedFiles {
			return
		}
		l.VerdictClass = VerdictClassWarn
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
