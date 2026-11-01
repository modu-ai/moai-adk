package cli

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// Only a complete commit SHA is a clue from the rows: the resume row of a parked card is a gate "about" n/a, and nothing
// else the ledger takes as a subject says which repository it came from. Each commit is named once.
func TestAFVerbs_N51bOnlyCompleteShasAreCluesFromTheRows(t *testing.T) {
	ctx := context.Background()
	a := afN38UnboundCard(t)
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	for _, sha := range []string{"n/a", "HEAD", a.c1[:12], a.c1, a.c1} {
		if err := a.st.AddEvidence(ctx, "t1", epoch, ledger.Evidence{Kind: "gate", SubjectSHA: sha, Result: "pass"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := afRowCommits(ctx, a.st)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{a.c1}; !reflect.DeepEqual(got, want) {
		t.Errorf("afRowCommits = %v, want %v", got, want)
	}
}
