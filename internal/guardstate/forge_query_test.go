package guardstate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// The forge-backed querier's own behaviour, measured without a subprocess.
//
// What these tests do NOT cover is stated rather than implied: that the binary
// is `gh`, and that it runs with the evaluated tree as its working directory.
// Those two lines live inside runGH and are reached only by an actual
// invocation, so they are a recorded gap rather than a covered claim.

// stubGH replaces the invocation for one test.
func stubGH(t *testing.T, reply func(dir string, args []string) ([]byte, error)) {
	t.Helper()
	original := runGH
	t.Cleanup(func() { runGH = original })
	runGH = func(_ context.Context, dir string, args []string) ([]byte, error) {
		return reply(dir, args)
	}
}

// TestFailedListingIsAnErrorNotAnEmptyHistory is the distinction row 2 exists
// for. An empty history says the subject has not run; a failed query says
// nobody knows. Collapsing the second into the first hands the operator "look
// again with a longer window" for what may be an expired credential — and, one
// row over, produces a false clean for a subject nothing could read.
func TestFailedListingIsAnErrorNotAnEmptyHistory(t *testing.T) {
	stubGH(t, func(string, []string) ([]byte, error) {
		return nil, errors.New("gh: authentication required")
	})

	history, err := newGHQuerier("/tree").RunsForSubject(context.Background(), ".github/workflows/a.yml")
	if err == nil {
		t.Fatalf("a failed listing returned no error, and %d runs", len(history.Runs))
	}
	if !strings.Contains(err.Error(), "authentication required") {
		t.Fatalf("the failure does not carry what went wrong: %v", err)
	}
}

// TestUnreadableListingIsAnError covers the same boundary one layer in: a
// listing that ran and returned something undecodable is not an empty history
// either.
func TestUnreadableListingIsAnError(t *testing.T) {
	stubGH(t, func(string, []string) ([]byte, error) {
		return []byte("not json"), nil
	})

	if _, err := newGHQuerier("/tree").RunsForSubject(context.Background(), ".github/workflows/a.yml"); err == nil {
		t.Fatal("an undecodable listing was reported as a readable one")
	}
}

// TestListingDecodesConclusionAndTimestamp pins the field mapping. Both fields
// are read by the entry axis — the conclusion by the measured quantity, the
// timestamp by the window — so a mapping that dropped either would classify on
// a zero value without saying so.
func TestListingDecodesConclusionAndTimestamp(t *testing.T) {
	history, err := decodeRunListing([]byte(
		`[{"conclusion":"success","createdAt":"2026-08-30T10:00:00Z"},` +
			`{"conclusion":"skipped","createdAt":"2026-08-29T10:00:00Z"},` +
			`{"conclusion":"","createdAt":"2026-08-28T10:00:00Z"}]`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(history.Runs) != 3 {
		t.Fatalf("decoded %d runs, want 3", len(history.Runs))
	}

	want := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	if !history.Runs[0].At.Equal(want) {
		t.Fatalf("the first run's timestamp is %v, want %v", history.Runs[0].At, want)
	}
	if history.Runs[0].Conclusion != "success" {
		t.Fatalf("the first run's conclusion is %q, want %q", history.Runs[0].Conclusion, "success")
	}

	// The three measured quantities admit three different subsets of this
	// history, which is what stops one number measuring both whether a guard
	// fired and whether a firing did anything.
	for m, want := range map[Measure]int{
		MeasureFiredAtAll:      3,
		MeasureFiredWithEffect: 2,
		MeasureVerdictRendered: 1,
	} {
		if got := len(history.Qualifying(m)); got != want {
			t.Fatalf("%s admits %d of the decoded runs, want %d", m, got, want)
		}
	}
}

// TestEmptyListingIsAnEmptyHistory is the other side of the boundary: a
// successful query that found nothing is not a failure. Row 6's implied action
// — look again with a longer window — is right for this and wrong for the case
// above, which is why the two must not share a representation.
func TestEmptyListingIsAnEmptyHistory(t *testing.T) {
	history, err := decodeRunListing([]byte(`[]`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(history.Runs) != 0 {
		t.Fatalf("an empty listing decoded to %d runs", len(history.Runs))
	}
}

// TestQuerierRunsInTheEvaluatedTree records the one property of the invocation
// a stub can still speak for: which tree the query is taken against. A query
// run in the wrong directory answers about the wrong repository and looks
// exactly like a healthy one.
func TestQuerierRunsInTheEvaluatedTree(t *testing.T) {
	var sawDir string
	var sawArgs []string
	stubGH(t, func(dir string, args []string) ([]byte, error) {
		sawDir, sawArgs = dir, args
		return []byte(`[]`), nil
	})

	if _, err := newGHQuerier("/some/tree").RunsForSubject(context.Background(), ".github/workflows/a.yml"); err != nil {
		t.Fatalf("query: %v", err)
	}
	if sawDir != "/some/tree" {
		t.Fatalf("the query ran in %q, not the evaluated tree", sawDir)
	}
	if strings.Join(sawArgs, " ") != strings.Join(runListArgs(".github/workflows/a.yml"), " ") {
		t.Fatalf("the query sent %v, not the per-subject vector", sawArgs)
	}
}

// TestGlobalListingIsReachableAndUnused keeps the repository-global listing
// callable, which is what lets AC-GSM-006 (b) be a measured call count rather
// than a source grep: a mutant assembling the same request from string
// fragments defeats a grep and not a counter.
func TestGlobalListingIsReachableAndUnused(t *testing.T) {
	var sawArgs []string
	stubGH(t, func(_ string, args []string) ([]byte, error) {
		sawArgs = args
		return []byte(`[]`), nil
	})

	if _, err := newGHQuerier("/tree").AllRuns(context.Background()); err != nil {
		t.Fatalf("global listing: %v", err)
	}
	if strings.Join(sawArgs, " ") != strings.Join(allRunsArgs(), " ") {
		t.Fatalf("the global listing sent %v", sawArgs)
	}
}
