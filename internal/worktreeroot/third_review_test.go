package worktreeroot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReviewOwnerMarkerRequiresAuthenticatedOwner(t *testing.T) {
	for _, body := range []string{`{}`, `{"schema":1,"owner":""}`, `{"schema":1,"owner":17}`, `{"schema":1,"owner":"relative"}`} {
		t.Run(body, func(t *testing.T) {
			home := wrHome(t)
			repo := wrRepo(t, "owner-required")
			res, err := Resolve(repo)
			if err != nil {
				t.Fatal(err)
			}
			marker := wrMarker(home, res.Key)
			if err := os.WriteFile(marker, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(repo); err == nil {
				t.Error("recorded assignment accepted an unauthenticated owner")
			}
			if _, ok := OwnerMarkerOf(res.Root); ok {
				t.Error("invalid owner was exposed as a classifier witness")
			}
			after, err := os.ReadFile(marker)
			if err != nil || string(after) != body {
				t.Fatalf("refusal rewrote the existing marker: %q, %v", after, err)
			}
		})
	}
}

func TestReviewSeparateGitDirsDoNotShareProjectOwnership(t *testing.T) {
	wrHome(t)
	parent := wrReal(t, t.TempDir())
	var results []Resolution
	for _, name := range []string{"first", "second"} {
		repo, common := filepath.Join(parent, name), filepath.Join(parent, "metadata", name+".git")
		if err := os.MkdirAll(filepath.Dir(common), 0755); err != nil {
			t.Fatal(err)
		}
		wrGit(t, parent, "init", "-q", "--template=", "-b", "main", "--separate-git-dir", common, repo)
		wrGit(t, repo, "-c", "user.name=Review", "-c", "user.email=separate@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "seed")
		linked := wrLinked(t, repo, name+"-linked")
		primary, err := Resolve(repo)
		if err != nil {
			t.Fatal(err)
		}
		fromLinked, err := Resolve(linked)
		if err != nil {
			t.Fatal(err)
		}
		if primary.Key != fromLinked.Key || !Same(primary.Root, fromLinked.Root) {
			t.Errorf("one repository split its assignment: %+v / %+v", primary, fromLinked)
		}
		owner, ok := OwnerMarkerOf(primary.Root)
		if !ok || Same(owner, filepath.Dir(common)) {
			t.Errorf("ownership still names the shared metadata parent: %q, valid=%v", owner, ok)
		}
		results = append(results, primary)
	}
	if results[0].Key == results[1].Key || Same(results[0].Root, results[1].Root) {
		t.Fatalf("separate repositories share a writable project root: %+v", results)
	}
}
