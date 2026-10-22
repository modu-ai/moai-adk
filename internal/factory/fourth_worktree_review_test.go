package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestCurrentReviewStatusRejectsCopiedProjectKey(t *testing.T) {
	repo, foreignTree := m3BaseFixture(t)
	other := filepath.Join(m3Real(t, t.TempDir()), "other-primary")
	if err := os.Mkdir(other, 0o755); err != nil {
		t.Fatal(err)
	}
	// A second real repository records the first repository's otherwise valid
	// key, reproducing a copied/stale record without changing the actual owner.
	mustGit(t, other, "init", "-q", "--template=", "-b", "other-main")
	mustGit(t, other, "fetch", "--quiet", repo, "+refs/heads/*:refs/heads/*")
	mustGit(t, other, "reset", "--hard", "main")
	key := worktreeroot.RecordedKey(repo)
	if key == "" {
		t.Fatal("fixture has no assigned key")
	}
	if err := os.WriteFile(filepath.Join(other, ".git", "moai-project-key"), []byte(key+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadCardStatus(other, "t7", DefaultWorktreeBases(other))
	if err == nil {
		t.Fatalf("copied key adopted status %+v from foreign tree %s without refusing its owner", got, foreignTree)
	}
	if got != nil {
		t.Fatalf("ownership refusal returned status %+v", got)
	}
}
