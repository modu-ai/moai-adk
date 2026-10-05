// Package binlag holds the one comparison that decides whether the installed
// moai binary was built from a commit the current source tree has already
// moved past — the deployment-lag verdict (SPEC-BINARY-LAG-VISIBILITY-001).
//
// The comparison itself is not new: it was already correct inside the
// `moai doctor` check item. What was missing was reach — the only way to see
// the verdict was to type the diagnostic command by hand, and no automatic
// caller existed. This package exists so exactly one implementation of the
// comparison can serve two surfaces: the doctor item, and the unprompted
// session-start advisory.
//
// It lives below both callers on purpose. `internal/cli` imports
// `internal/hook`, so a seam placed in `internal/cli` could not be reached
// from the session-start handler at all. Placing it here also makes the
// single-implementation property testable rather than merely asserted: a stub
// installed in Comparer is observed by BOTH surfaces, which a per-package seam
// variable could never demonstrate.
package binlag

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Status is the lag verdict.
type Status string

const (
	// StatusNotApplicable means no comparison was possible: the binary carries
	// no commit metadata, or the directory is not inside a git working tree.
	// This is the case that keeps downstream users whole — a deployed project
	// has no repository to compare against, and treating that as a failure
	// would turn every downstream `moai doctor` run into exit 1.
	StatusNotApplicable Status = "not-applicable"
	// StatusFresh means the binary was built from this tree's HEAD.
	StatusFresh Status = "fresh"
	// StatusDivergent means the binary's commit stands in no ancestor relation
	// to HEAD in EITHER direction — a release or sibling-branch build, not a
	// stale one. A tree that has never heard of the binary's commit lands here,
	// which is every downstream installation, and the verdict stays silent.
	StatusDivergent Status = "divergent"
	// StatusAhead means the binary's commit is a strict DESCENDANT of HEAD: the
	// tree being compared against is the older of the two.
	//
	// This is not a divergence and must not be reported as one. The two commits
	// are in an ancestor relation, so "release or sibling-branch build" states
	// the wrong fact; what is actually true is that the comparison ref is behind
	// the binary, and therefore the lag verdict says nothing about whether the
	// binary is current. That is worth reporting precisely because it is the
	// case where the check answers a question nobody asked.
	StatusAhead Status = "ahead"
	// StatusBehind means the binary's commit is a strict ancestor of HEAD:
	// commits landed after the binary was built, so it is running old code.
	StatusBehind Status = "behind"
)

// Request carries everything the comparison reads.
//
// BinaryVersion is carried for reporting only and takes NO part in the
// verdict. The semantic version string cannot decide lag: a default build off
// a later commit reports a LOWER version than an earlier explicit release
// candidate, so a version comparison reaches the opposite conclusion on
// exactly the inversion this SPEC was written from.
type Request struct {
	Dir           string
	BinaryCommit  string
	BinaryVersion string
}

// Verdict is the comparison's answer.
type Verdict struct {
	Status       Status
	BinaryCommit string
	// SourceHead is the tree HEAD the binary was compared against; empty when
	// the verdict is not applicable.
	SourceHead string
	// Reason explains a not-applicable or divergent verdict in one phrase.
	Reason string
}

// Comparer is the single substitutable seam. Both the doctor check item and
// the session-start advisory reach the comparison through it, so a stub
// installed here is observed by both.
//
// @MX:ANCHOR: the one comparison implementation both lag surfaces read
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
// @MX:REASON: a second copy of the comparison in either caller lets the two
// surfaces drift apart silently; routing both through this variable is what
// makes "a single implementation" observable rather than merely asserted.
// @MX:TEST: internal/cli/binary_lag_test.go, internal/hook/session_start_binary_lag_test.go
var Comparer = gitCompare

// Evaluate returns the lag verdict for req, routed through Comparer.
//
// @MX:ANCHOR: fan_in 3 — doctor.go:520, session_start_binary_lag.go:55 and :67
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
// @MX:REASON: package API boundary; every lag verdict anywhere in the binary
// is produced here, so a change to its signature or semantics reaches both
// the diagnostic surface and the unprompted session-start surface at once.
func Evaluate(ctx context.Context, req Request) Verdict {
	return Comparer(ctx, req)
}

// gitCompare is the real comparison: the binary's build commit against the
// tree HEAD, decided by ancestry.
//
// Five paths report not-applicable or otherwise-fine rather than a problem —
// no commit metadata, not a git tree, HEAD match, and non-ancestor. That
// leniency is deliberate and must not be narrowed: each narrowing turns a
// perfectly healthy downstream installation into a reported fault.
//
// @MX:WARN: shells out to git; every non-lag path must stay lenient
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
// @MX:REASON: the two exec calls are the only external-system dependency in
// this package, and each early return is a deliberate not-a-problem verdict.
// Converting any of them into a fault promotes a downstream `moai doctor` run
// to a failure over a repository the user does not have.
func gitCompare(ctx context.Context, req Request) Verdict {
	binCommit := strings.TrimSpace(req.BinaryCommit)
	if binCommit == "" || binCommit == "none" || binCommit == "unknown" {
		return Verdict{Status: StatusNotApplicable, Reason: "development build (no commit metadata)"}
	}

	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		return Verdict{Status: StatusNotApplicable, BinaryCommit: binCommit, Reason: "cannot determine working directory"}
	}

	// `git rev-parse` walks upward from dir, so running from a subdirectory of
	// an applicable tree resolves the same HEAD as running from its root.
	headOut, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return Verdict{Status: StatusNotApplicable, BinaryCommit: binCommit, Reason: "not in a git source tree (skipped)"}
	}
	sourceHead := strings.TrimSpace(string(headOut))

	if strings.HasPrefix(sourceHead, binCommit) {
		return Verdict{Status: StatusFresh, BinaryCommit: binCommit, SourceHead: sourceHead}
	}

	if err := exec.CommandContext(ctx, "git", "-C", dir, "merge-base", "--is-ancestor", binCommit, sourceHead).Run(); err == nil {
		return Verdict{Status: StatusBehind, BinaryCommit: binCommit, SourceHead: sourceHead}
	}

	// The same probe in the other direction. Without it, "binary is newer than
	// the ref we compared against" is indistinguishable from "binary belongs to
	// an unrelated line", and the two want opposite treatment: the first says
	// the comparison was meaningless, the second says there was nothing to
	// compare.
	//
	// It costs one git call on a path that previously made two, and only on the
	// non-ancestor branch — the downstream-common case reaches it, fails it
	// (the commit is not in that repository at all), and lands on the same
	// divergent verdict it always did.
	if err := exec.CommandContext(ctx, "git", "-C", dir, "merge-base", "--is-ancestor", sourceHead, binCommit).Run(); err == nil {
		return Verdict{
			Status:       StatusAhead,
			BinaryCommit: binCommit,
			SourceHead:   sourceHead,
			Reason:       "source HEAD is an ancestor of the binary commit (comparison ref is behind the binary)",
		}
	}

	return Verdict{
		Status:       StatusDivergent,
		BinaryCommit: binCommit,
		SourceHead:   sourceHead,
		Reason:       "binary commit is in no ancestor relation to source HEAD (release or branch build)",
	}
}

// RemedyCommand is the one thing a reader of the advisory has to do about it.
const RemedyCommand = "make build && make install"

// Advisory renders the unprompted notice for a verdict, and the empty string
// for every verdict that is not lag.
//
// Silence on the other verdicts is the whole design. A notice on every session
// start is read once and ignored thereafter, and an ignored notice closes
// nothing — so the advisory speaks only when the binary really is running code
// the tree has moved past.
//
// StatusAhead speaks too, and says something different. There the binary is
// not stale; the ref it was compared against is, so the verdict carries no
// information about the binary's currency. Rebuilding is not the remedy — it
// would change today's number and leave the next silence exactly as invisible
// — so the notice names the condition and points at the comparison instead.
//
// @MX:NOTE: empty string for every verdict except StatusBehind and StatusAhead
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
func Advisory(v Verdict) string {
	if v.Status == StatusAhead {
		return fmt.Sprintf(
			"moai binary lag check inconclusive: the installed binary was built from commit %s, "+
				"a DESCENDANT of this tree's HEAD %s.\n"+
				"The binary is newer than the ref it was compared against, so this check cannot tell you "+
				"whether the binary is current — it may still be behind the branch it was built from.\n"+
				"Compare against that branch before trusting any moai CLI result "+
				"(git merge-base --is-ancestor %s <branch>).",
			Short(v.BinaryCommit), Short(v.SourceHead), Short(v.BinaryCommit),
		)
	}
	if v.Status != StatusBehind {
		return ""
	}
	return fmt.Sprintf(
		"moai binary lag: the installed binary was built from commit %s, an ancestor of this tree's HEAD %s.\n"+
			"Fixes committed after %s are NOT in the binary you are running, so its output describes older code.\n"+
			"Rebuild before trusting any moai CLI result: %s",
		Short(v.BinaryCommit), Short(v.SourceHead), Short(v.BinaryCommit), RemedyCommand,
	)
}

// Short returns the first 9 characters of a commit hash.
func Short(hash string) string {
	if len(hash) < 9 {
		return hash
	}
	return hash[:9]
}
