//go:build afred

package ledger

import (
	"errors"
	"testing"
	"time"
)

func TestEvidenceIsStoredWithTheEpochAndTheDatabaseClock(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	in := Evidence{Kind: KindGate, TreeSHA: "tree1", SubjectSHA: "commit1", BaseSHA: "base", Command: "go test ./...", Result: ResultPass, Path: ".moai/reports/a/gate.log"}
	if err := s.AddEvidence(bg, "a", c.Epoch, in); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindRed, TreeSHA: "tree0", Result: ResultFail, Guards: []string{"ignored"}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvidenceFor(bg, "a")
	if err != nil || len(got) != 2 {
		t.Fatalf("%+v %v", got, err)
	}
	if got[0].Kind != KindGate || got[0].TreeSHA != "tree1" || got[0].SubjectSHA != "commit1" || got[0].BaseSHA != "base" ||
		got[0].Command != "go test ./..." || got[0].Result != ResultPass || got[0].Path != ".moai/reports/a/gate.log" {
		t.Errorf("round trip: %+v", got[0])
	}
	if got[0].Epoch != c.Epoch || got[0].CardID != "a" || time.Since(got[0].At) > time.Minute || got[0].ID >= got[1].ID {
		t.Errorf("stamps: %+v", got[0])
	}
	if got[1].Kind != KindRed || got[1].SubjectSHA != "" || got[1].Guards != nil {
		t.Errorf("a red row: %+v", got[1])
	}
	if other, _ := s.EvidenceFor(bg, "nobody"); len(other) != 0 {
		t.Errorf("another card's evidence: %+v", other)
	}
}

func TestEvidenceNeedsAKnownKindAResultAndATree(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for name, e := range map[string]Evidence{
		"kind":   {Kind: "vibes", TreeSHA: "t", Result: ResultPass},
		"result": {Kind: KindGate, TreeSHA: "t", Result: "maybe"},
		"tree":   {Kind: KindGate, Result: ResultPass},
	} {
		if err := s.AddEvidence(bg, "a", c.Epoch, e); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if got, _ := s.EvidenceFor(bg, "a"); len(got) != 0 {
		t.Errorf("a refused row was stored: %+v", got)
	}
}
