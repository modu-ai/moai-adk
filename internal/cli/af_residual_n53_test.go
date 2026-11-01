package cli

// N53 (audit 1c05083f71, kept by the operator as an accepted residual; residual classification): an unbound ledger holds two
// leases whose worktrees are in two different repositories. Every verb run from either repository is refused, because the
// other lease disagrees. The way out exists and is found nowhere in the refusal: a release run from a directory that is no
// git repository binds nothing and is accepted. The expectation: the refusal names that way out.

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFResidual_N53TheRefusalOfTwoLeasesInTwoRepositoriesNamesTheWayOut(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()
	p, _ := afResRepo(t, filepath.Join(root, "P"))
	o, _ := afResRepo(t, filepath.Join(root, "O"))
	nonrepo := filepath.Join(root, "nonrepo")
	if err := os.MkdirAll(nonrepo, 0o755); err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(root, "central", "ledger.db")
	afResLedger(t, db, "t1", "t2")
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	c1, err := st.Claim(context.Background(), ledger.ClaimRequest{SessionID: "s", Worktree: p, PID: os.Getpid(), PIDStart: "x"})
	if err != nil {
		t.Fatal(err)
	}
	c2, err := st.Claim(context.Background(), ledger.ClaimRequest{SessionID: "s2", Worktree: o, PID: os.Getpid(), PIDStart: "y"})
	if err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	t.Setenv("CLAUDE_PROJECT_DIR", p)
	text, err := afResVerb(db, "release", c1.CardID, "--epoch", strconv.FormatInt(c1.Epoch, 10))
	t.Logf("release from P: err=%v (%.300s)", err, text)
	if err == nil {
		t.Fatal("setup: the release from P was accepted, so the two leases do not disagree")
	}
	if !strings.Contains(text, "no git repository") && !strings.Contains(text, "is no repository") && !strings.Contains(text, "not in a git repository") {
		t.Errorf("the refusal does not say that a release from a directory that is no git repository gets out: %.400s", text)
	}
	// the way out exists: from a directory that is no repository the release binds nothing and goes through
	t.Setenv("CLAUDE_PROJECT_DIR", nonrepo)
	if text, err := afResVerb(db, "release", c2.CardID, "--epoch", strconv.FormatInt(c2.Epoch, 10)); err != nil {
		t.Errorf("the release from a directory that is no repository is refused: %v (%.200s)", err, text)
	}
}
