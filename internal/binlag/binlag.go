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
	// to HEAD — a release or sibling-branch build, not a stale one.
	StatusDivergent Status = "divergent"
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
var Comparer = gitCompare

// Evaluate returns the lag verdict for req, routed through Comparer.
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

	return Verdict{
		Status:       StatusDivergent,
		BinaryCommit: binCommit,
		SourceHead:   sourceHead,
		Reason:       "binary commit is not an ancestor of source HEAD (release or branch build)",
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
func Advisory(v Verdict) string {
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
