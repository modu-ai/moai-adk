package cli

// Helpers of the residual-classification tests (the items of the final audit of a36ea8c368 that the repairs left open,
// measured again on the tree of a43cc07c18). Real ledger, real git, the real `moai af` verbs through afRunBoth. The tests
// themselves are in af_residual_<item>_test.go and begin with TestAFResidual_.

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afResRepo makes a repository at dir ("" for a fresh temporary one) with the dedicated trunk and a commit c on a branch
// that is not on the trunk, and leaves the work branch checked out.
func afResRepo(t *testing.T, dir string) (repo, c string) {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	afGit(t, dir, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("base "+filepath.Base(dir)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, dir, "add", "base.txt")
	afGit(t, dir, "commit", "-q", "-m", "base")
	afGit(t, dir, "branch", "af/trunk")
	afGit(t, dir, "checkout", "-q", "-b", "c1", "af/trunk")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, dir, "add", "a.txt")
	afGit(t, dir, "commit", "-q", "-m", "card")
	c = afGit(t, dir, "rev-parse", "HEAD")
	afGit(t, dir, "checkout", "-q", "work")
	return dir, c
}

// afResLedger makes a ledger at db holding the cards ids in READY (t1 when none is named) and leaves it unbound.
func afResLedger(t *testing.T, db string, ids ...string) {
	t.Helper()
	if len(ids) == 0 {
		ids = []string{"t1"}
	}
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	for _, id := range ids {
		if err := st.AddCard(context.Background(), ledger.Card{ID: id, Title: id, Position: core.Position{Stage: core.StageReady}, Priority: 1,
			Spec: ledger.Spec{Class: "A", Paths: []string{id + "/" + id + ".txt"}}}); err != nil { // own directory: cards that share a key cannot be claimed together
			t.Fatal(err)
		}
	}
}

// afResBound is the repository the ledger at db is bound to ("" when it is not bound).
func afResBound(t *testing.T, db string) string {
	t.Helper()
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	b, err := st.Repository(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// afResStages is "id:STAGE/epoch" of every card of the ledger at db.
func afResStages(t *testing.T, db string) string {
	t.Helper()
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	cards, err := st.Cards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var parts []string
	for _, c := range cards {
		parts = append(parts, c.ID+":"+string(c.Position.Stage)+"/"+strconv.FormatInt(c.Epoch, 10))
	}
	return strings.Join(parts, " ")
}

// afResVerb runs one verb against the ledger at db and returns what it printed, the error text included.
func afResVerb(db string, args ...string) (string, error) {
	out, errOut, err := afRunBoth(append(append([]string{}, args...), "--db", db)...)
	text := out + errOut
	if err != nil {
		text += " " + err.Error()
	}
	return text, err
}
