package revoke_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/revoke"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestRevokeLeavesRepositoryUntouched checks that revoke changes no
// repository state: worktrees, refs, the queue database, the contract, and
// the SPEC documents stay byte-identical, and the only git operation it
// performs through its seam is reading HEAD (AC-GR-024).
func TestRevokeLeavesRepositoryUntouched(t *testing.T) {
	isolate(t)
	p := signtest.New(t)
	signHuman(t, p)
	p.Git("branch", "feature-a")
	p.Git("branch", "feature-b")
	p.Git("update-ref", "refs/remotes/origin/main", "HEAD")
	wt := filepath.Join(t.TempDir(), "wt")
	p.Git("worktree", "add", "-q", wt, "feature-a")

	db, err := homestate.BacklogDBPath(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(db), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(db, []byte("backlog fixture bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	type state struct{ worktrees, refs, db, contractYAML, spec, acceptance string }
	read := func() state {
		b, _ := os.ReadFile(db)
		return state{
			worktrees:    p.Git("worktree", "list", "--porcelain"),
			refs:         p.Git("for-each-ref"),
			db:           string(b),
			contractYAML: string(p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile))),
			spec:         string(p.ReadFile(signtest.SpecRel(signtest.SpecID, "spec.md"))),
			acceptance:   string(p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile))),
		}
	}
	before := read()

	var gitCalls []string
	s := seams()
	s.GitHead = func(root string) (string, error) {
		gitCalls = append(gitCalls, "rev-parse HEAD")
		return strings.TrimSpace(p.Git("rev-parse", "HEAD")), nil
	}
	res, err := revoke.Revoke(opts(p), s)
	if err != nil || res.Status != revoke.StatusRevoked {
		t.Fatalf("revoke: status %q err %v", res.Status, err)
	}
	if after := read(); after != before {
		t.Errorf("repository state changed:\nbefore %+v\nafter  %+v", before, after)
	}
	for _, c := range gitCalls {
		for _, bad := range []string{"push", "branch -d", "branch -m", "worktree remove"} {
			if strings.Contains(c, bad) {
				t.Errorf("revoke ran a git write: %s", c)
			}
		}
	}
	if len(gitCalls) != 1 {
		t.Errorf("git seam calls = %v, want exactly one HEAD read", gitCalls)
	}
}
