package guardstate

import (
	"os"
	"regexp"
	"testing"
)

// The value-blind partition harness lives in partition_blind_test.go. THIS file
// is its auditor, and the two are separate files for a reason that is the whole
// point of the pair: the audit has to NAME the vocabulary in order to forbid
// it, so an auditor living inside the harness would be the very match it
// searches for.

// TestThisHarnessNamesNoValue is what makes the blindness a MEASUREMENT rather
// than a claim. It greps this file for the producer's vocabulary — both the
// wire tokens and the Go identifiers that carry them — and requires no match.
//
// Stated over this file's own bytes rather than over a copy, so it cannot drift
// from what it audits.
func TestThisHarnessNamesNoValue(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("partition_blind_test.go")
	if err != nil {
		t.Fatalf("read the harness: %v", err)
	}

	// The wire tokens, matched as whole words. The two-letter one is matched
	// case-sensitively because a case-insensitive whole-word form of it is a
	// common English word; the rest are matched case-insensitively, so a
	// lower-cased mention in a comment is caught too.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\bOK\b`),
		regexp.MustCompile(`(?i)\b(stale|unknown|undeclared|unreadable|unresolved|orphaned)\b`),
		// The Go identifiers that carry them, and the fold's own surface names.
		regexp.MustCompile(`\bClass[A-Z]`),
		regexp.MustCompile(`\bSurface[A-Z]`),
	}

	for _, p := range patterns {
		if m := p.FindAll(source, -1); m != nil {
			t.Fatalf("the value-blind harness names the vocabulary: %q matched %d time(s), first %q",
				p, len(m), m[0])
		}
	}
}
