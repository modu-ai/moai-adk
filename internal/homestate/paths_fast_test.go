package homestate

import (
	"path/filepath"
	"testing"
)

func TestPrimaryCheckoutRootFromCommonDir(t *testing.T) {
	repo := filepath.Join(string(filepath.Separator), "repo")
	if got, ok := primaryCheckoutRootFromCommonDir(filepath.Join(repo, ".git")); !ok || got != repo {
		t.Fatalf("ordinary common dir resolved to %q ok=%v, want %q true", got, ok, repo)
	}
	for _, path := range []string{
		filepath.Join(string(filepath.Separator), "metadata", "repo.git"),
		filepath.Join(string(filepath.Separator), "metadata", "common"),
	} {
		if got, ok := primaryCheckoutRootFromCommonDir(path); ok || got != "" {
			t.Fatalf("external common dir %q resolved to %q ok=%v", path, got, ok)
		}
	}
}
