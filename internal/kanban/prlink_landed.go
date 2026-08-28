// prlink_landed.go — the landed check and its regex-engine guard
// (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.9, REQ-1.10, M2;
// SPEC-TODO-LANDING-STATE-001 REQ-TLS-001..REQ-TLS-004, M1/M2).
//
// This file is separated from the resolver because its failure mode is
// SILENT. `\b` is not POSIX ERE, and git does not error on it — it returns an
// empty commit set, which is byte-for-byte indistinguishable from "this card
// has not landed". An `-E` regression would therefore report every card as
// not-landed and pass any check that only looks for a clean result.
//
// Two things stand against that: the engine flag is a named constant used by
// exactly one argv builder, and the acceptance suite carries a positive
// control plus an `-E` tripwire rather than a single negative assertion.
//
// The SAME silence has a second entrance, and it is what the landing-state
// SPEC closes: asking the question about the WRONG ref. A project that
// integrates on a branch other than the default answers "not landed" for
// every card that shipped, with no error and no empty-output warning. The ref
// is therefore resolved from configuration and threaded through as an input,
// and an unanswerable query is reported as `unknown` rather than collapsed
// into `not-landed`.
package kanban

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// LandedRegexpEngineFlag is the ONLY regular-expression engine the landed
// check may use. It is a constant rather than a literal so a change to it is
// a one-line diff a reviewer sees, and so the tripwire test can assert the
// built argv against the same symbol the implementation uses.
const LandedRegexpEngineFlag = "--perl-regexp"

// DefaultLandedRef is the ref the landed question falls back to when the
// project configures no integration branch. It is the historical constant, so
// an unconfigured project behaves exactly as it did before the ref became
// resolvable — the property that makes the resolution safe to land on its own.
const DefaultLandedRef = "origin/main"

// LandingAnswer is the landed check's THREE-valued answer.
//
// The third value is the point. Collapsing an unanswerable query into
// `not-landed` makes "the check ran and found nothing" and "the check could
// not run" the same bytes, so a guard that never executed reads exactly like
// a guard that passed. They are different facts and they are reported
// differently.
type LandingAnswer string

const (
	// LandingLanded means the resolved ref's history names the card.
	LandingLanded LandingAnswer = "landed"
	// LandingNotLanded means the query ran against the resolved ref and
	// returned an empty commit set.
	LandingNotLanded LandingAnswer = "not-landed"
	// LandingUnknown means the question could not be asked — no git, no such
	// ref, a broken remote. It is NOT evidence of not-landed.
	LandingUnknown LandingAnswer = "unknown"
)

// LandedRefFor returns the ref the landed question should be asked about for
// the project rooted at projectRoot: `origin/<worktree_base_branch>` when the
// project configures one, and DefaultLandedRef when it does not.
//
// "Has this landed?" means "is it in the branch this project integrates on".
// For a project that integrates on `develop`, asking `origin/main` answers a
// question nobody posed and answers it wrongly.
//
// Every failure path yields DefaultLandedRef, inheriting the neutral-empty
// ruling of config.LoadWorktreeBaseBranch: a project whose configuration
// cannot be read behaves as one that never configured the key.
func LandedRefFor(projectRoot string) string {
	base := strings.TrimSpace(config.LoadWorktreeBaseBranch(projectRoot))
	if base == "" {
		return DefaultLandedRef
	}
	return "origin/" + base
}

// CommandRunner runs one subprocess and returns its combined stdout. It is
// the injection seam the subprocess-census tests count invocations through.
type CommandRunner func(name string, args ...string) (string, error)

// LandedGrepArgs builds the exact argv the landed check runs against ref.
//
// The ref is an INPUT rather than a package constant so the package stays a
// pure function of what it is given (prlink.go's NFR-4 ruling). An empty ref
// falls back to DefaultLandedRef rather than emitting an empty argv element,
// which git would read as the working tree.
//
// Exported so the engine-flag tripwire can assert against the implementation's
// own construction rather than against a transcription of it — a test that
// rebuilds the argv itself proves nothing about the code.
func LandedGrepArgs(ref, cardID string) ([]string, error) {
	if !validCardToken.MatchString(cardID) {
		return nil, fmt.Errorf("prlink: card id %q is not a bare token", cardID)
	}
	if strings.TrimSpace(ref) == "" {
		ref = DefaultLandedRef
	}
	return []string{
		"log", ref,
		LandedRegexpEngineFlag,
		`--grep=\b` + cardID + `\b`,
		"--oneline",
	}, nil
}

// GitLandedQuerier answers Q2 against local git history. It performs no
// network I/O — the landed ref is an existing local ref — so it sits outside
// the one-`gh`-query budget entirely (NFR-2) and keeps working when `gh` does
// not (REQ-2.3).
type GitLandedQuerier struct {
	// Run executes git. Nil means the querier is unusable rather than
	// silently answering "not landed"; the constructor below supplies the
	// real runner.
	Run CommandRunner
	// Ref is the ref the question is asked about. Empty means
	// DefaultLandedRef, which keeps an unconfigured project byte-identical to
	// the pre-resolution behaviour. Callers resolve it with LandedRefFor.
	Ref string
}

// LandedRef reports the ref this querier actually asks about, so a caller can
// name it in a refusal or a note without re-deriving it.
func (q GitLandedQuerier) LandedRef() string {
	if strings.TrimSpace(q.Ref) == "" {
		return DefaultLandedRef
	}
	return q.Ref
}

// Landed reports whether the resolved ref's history names the card.
//
// It returns a three-valued answer and an error, and NOTHING else (REQ-1.10).
// The commit set is consumed here and discarded: no SHA and no subject escapes
// this function, because a card's first matching commit may be another card's
// report commit that merely mentions it. The answer's ARITY changed; what it
// may carry did not.
//
// An unanswerable query yields LandingUnknown alongside the error, so a caller
// that drops the error still cannot mistake it for not-landed.
func (q GitLandedQuerier) Landed(cardID string) (LandingAnswer, error) {
	ref := q.LandedRef()
	if q.Run == nil {
		return LandingUnknown, fmt.Errorf("prlink: landed querier has no command runner")
	}
	args, err := LandedGrepArgs(ref, cardID)
	if err != nil {
		return LandingUnknown, err
	}
	out, err := q.Run("git", args...)
	if err != nil {
		return LandingUnknown, fmt.Errorf("prlink: git log %s: %w", ref, err)
	}
	if strings.TrimSpace(out) == "" {
		return LandingNotLanded, nil
	}
	return LandingLanded, nil
}
