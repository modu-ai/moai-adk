//go:build windows

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCurrentReviewWindowsCaseAliasKeepsLiveAnchor(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "MixedCaseTree")
	if err := os.Mkdir(tree, 0o755); err != nil {
		t.Fatal(err)
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	// A recorded descendant may have disappeared since registration. Literal
	// case-folded containment must still protect the live owning process;
	// EvalSymlinks alone cannot rescue this observation.
	cwd := filepath.Join(strings.ToUpper(tree), "removed-descendant")
	writeAnchorRegistry(t, tree, []Entry{anchorTestEntry(host, cwd, os.Getpid(), 2*time.Hour)})
	if got := LiveAnchoredSessionsForProject(tree, "", time.Now().UTC()); len(got) != 1 {
		t.Fatalf("case-aliased live session was lost: %+v", got)
	}
	if pathWithinTree(filepath.Join(strings.ToUpper(tree)+"-sibling", "cwd"), tree) {
		t.Fatal("case folding crossed the path-segment boundary")
	}
}
