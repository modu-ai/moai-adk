// gitin_env_test.go — card t1632 sync-audit repair F4: gitIn scrubs the
// inherited GIT_DIR / GIT_WORK_TREE / GIT_COMMON_DIR / GIT_INDEX_FILE /
// GIT_CEILING_DIRECTORIES, so a server process launched with one of those
// variables naming another repository binds the receipt state to the tree it
// was asked about — the same scrub storeroot.go already applies to its own
// store-root inspections.
package auditreceipt

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTreeState_IgnoresInheritedGitDir pins the scrub: with GIT_DIR pointing
// at ANOTHER repository, TreeState(root) still reads root's own state — an
// inherited variable must not redirect the state read. The discriminator is
// the digest pair: scrubbed, the read returns root's digest; inherited, git
// reads the OTHER repository and returns its digest.
func TestTreeState_IgnoresInheritedGitDir(t *testing.T) {
	root := initContentRepo(t)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n\n// modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A second repository as the hostile GIT_DIR target, with a DIFFERENT state.
	other := initContentRepo(t)
	if err := os.WriteFile(filepath.Join(other, "other.go"), []byte("package other\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, rootDigest, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState(root) baseline: %v", err)
	}

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	head, hostileDigest, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState with a hostile GIT_DIR: %v", err)
	}
	if head == "" || hostileDigest == "" {
		t.Fatal("TreeState produced an empty binding with a hostile GIT_DIR")
	}
	if hostileDigest != rootDigest {
		t.Fatalf("digest under a hostile GIT_DIR = %s, want root's own %s — the inherited variable redirected the state read (card-review sync-audit F4)", hostileDigest, rootDigest)
	}
}
