package receipt_test

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// TestStoreDirMatchesEscalation pins the store resolver to the escalation
// detector's store directory, for a primary checkout and a linked worktree.
func TestStoreDirMatchesEscalation(t *testing.T) {
	p := signtest.New(t)
	wt := filepath.Join(t.TempDir(), "linked")
	p.Git("branch", "side")
	p.Git("worktree", "add", "-q", wt, "side")
	for _, root := range []string{p.Root, wt} {
		got, err := receipt.StoreDir(root)
		if err != nil {
			t.Fatal(err)
		}
		want, err := escalation.StoreDir(root)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s: receipt.StoreDir = %s, escalation.StoreDir = %s", root, got, want)
		}
	}
	a, _ := receipt.StoreDir(p.Root)
	b, _ := receipt.StoreDir(wt)
	if a != b {
		t.Errorf("a linked worktree resolves a different store: %s vs %s", a, b)
	}
}
