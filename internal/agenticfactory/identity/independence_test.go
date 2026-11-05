package identity

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// AC-AF-010 case (e): writer independence is scoped to the Issue. A session that
// wrote an attempt on ANOTHER Issue is still a legitimate reviewer of this one.
func TestIndependence_WriterOfAnotherIssueStillCountsAsReviewer(t *testing.T) {
	r := newRig(t)
	s := r.producer(SessionReview, CodeReview) // reviewer role on Issue 1948

	issue2, err := core.NewIssueRef(1949)
	if err != nil {
		t.Fatal(err)
	}
	// the same session is the writer of a different Issue
	if _, err := r.book.Store.Acquire(issue2, s.ID, "d2", "run-1"); err != nil {
		t.Fatal(err)
	}

	got := r.attribute(r.evidence("ev-other", CodeReview, s, false))
	if !got.Attributed || got.Role != core.RoleReviewer {
		t.Fatalf("a writer of another Issue must still attribute as %s on this one, got %+v", core.RoleReviewer, got)
	}
}
