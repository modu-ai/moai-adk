// prlink_landed.go — the landed check and its regex-engine guard
// (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.9, REQ-1.10, M2).
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
package kanban

import (
	"fmt"
	"strings"
)

// LandedRegexpEngineFlag is the ONLY regular-expression engine the landed
// check may use. It is a constant rather than a literal so a change to it is
// a one-line diff a reviewer sees, and so the tripwire test can assert the
// built argv against the same symbol the implementation uses.
const LandedRegexpEngineFlag = "--perl-regexp"

// LandedRef is the ref the landed question is asked about. "Has this landed?"
// means "is it in the integration branch?", not "is it in my local history".
const LandedRef = "origin/main"

// CommandRunner runs one subprocess and returns its combined stdout. It is
// the injection seam the subprocess-census tests count invocations through.
type CommandRunner func(name string, args ...string) (string, error)

// LandedGrepArgs builds the exact argv the landed check runs.
//
// Exported so the AC-011 tripwire can assert the engine flag against the
// implementation's own construction rather than against a transcription of
// it — a test that rebuilds the argv itself proves nothing about the code.
func LandedGrepArgs(cardID string) ([]string, error) {
	if !validCardToken.MatchString(cardID) {
		return nil, fmt.Errorf("prlink: card id %q is not a bare token", cardID)
	}
	return []string{
		"log", LandedRef,
		LandedRegexpEngineFlag,
		`--grep=\b` + cardID + `\b`,
		"--oneline",
	}, nil
}

// GitLandedQuerier answers Q2 against local git history. It performs no
// network I/O — origin/main is an existing local ref — so it sits outside the
// one-`gh`-query budget entirely (NFR-2) and keeps working when `gh` does not
// (REQ-2.3).
type GitLandedQuerier struct {
	// Run executes git. Nil means the querier is unusable rather than
	// silently answering "not landed"; the constructor below supplies the
	// real runner.
	Run CommandRunner
}

// Landed reports whether origin/main history names the card.
//
// It returns a boolean and an error, and NOTHING else (REQ-1.10). The commit
// set is consumed here and discarded: no SHA and no subject escapes this
// function, because a card's first matching commit may be another card's
// report commit that merely mentions it.
func (q GitLandedQuerier) Landed(cardID string) (bool, error) {
	if q.Run == nil {
		return false, fmt.Errorf("prlink: landed querier has no command runner")
	}
	args, err := LandedGrepArgs(cardID)
	if err != nil {
		return false, err
	}
	out, err := q.Run("git", args...)
	if err != nil {
		return false, fmt.Errorf("prlink: git log %s: %w", LandedRef, err)
	}
	return strings.TrimSpace(out) != "", nil
}
