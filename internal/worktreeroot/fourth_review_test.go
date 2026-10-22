package worktreeroot

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// A directory replacing the record is a deterministic read error on every
// platform, including privileged test users that can still read mode 000 files.
func TestCurrentReviewRecordReadErrorsDoNotAssignProject(t *testing.T) {
	for _, mode := range []string{"plan", "resolve"} {
		t.Run(mode, func(t *testing.T) {
			home := wrHome(t)
			repo := wrRepo(t, "record-error")
			common := filepath.FromSlash(wrGit(t, repo, "rev-parse", "--path-format=absolute", "--git-common-dir"))
			if err := os.Mkdir(filepath.Join(common, recordName), 0o755); err != nil {
				t.Fatal(err)
			}
			var err error
			if mode == "plan" {
				_, err = Plan(repo, "card")
			} else {
				_, err = Resolve(repo)
			}
			if err == nil {
				t.Errorf("%s accepted an unreadable project-key record", mode)
			}
			if entries, readErr := os.ReadDir(wrProjects(home)); readErr == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						t.Errorf("%s created project folder %q after record read failure", mode, entry.Name())
					}
				}
			} else if !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
		})
	}
	t.Run("absent-record-still-assigns", func(t *testing.T) {
		wrHome(t)
		repo := wrRepo(t, "no-record")
		if got, err := Resolve(repo); err != nil || got.Key == "" {
			t.Fatalf("absent record resolution = %+v, %v", got, err)
		}
	})
}

func fourthReviewDirLink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err == nil {
		return
	} else if runtime.GOOS != "windows" {
		t.Fatal(err)
	}
	// A junction supplies a real directory alias on Windows hosts without the
	// symbolic-link privilege; it exercises the same resolved-target boundary.
	if out, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("create native directory alias: %v: %s", err, out)
	}
}

func TestCurrentReviewCandidatesRequireContainedRegisteredTrees(t *testing.T) {
	for _, kind := range []string{"outside-alias", "unregistered", "foreign-repository", "registered", "registered-alias"} {
		t.Run(kind, func(t *testing.T) {
			wrHome(t)
			repo := wrRepo(t, "candidate-owner")
			base := wrRoot(t, repo)
			candidate := filepath.Join(base, "card")
			switch kind {
			case "outside-alias":
				outside := wrLinked(t, repo, "outside")
				fourthReviewDirLink(t, outside, candidate)
			case "unregistered":
				if err := os.Mkdir(candidate, 0o755); err != nil {
					t.Fatal(err)
				}
			case "foreign-repository":
				foreign := wrRepo(t, "foreign-candidate")
				wrGit(t, foreign, "worktree", "add", "-q", "--detach", candidate, "HEAD")
			default:
				wrGit(t, repo, "worktree", "add", "-q", "--detach", candidate, "HEAD")
				if kind == "registered-alias" {
					legacy := filepath.Join(repo, ".moai", "worktrees")
					if err := os.MkdirAll(legacy, 0o755); err != nil {
						t.Fatal(err)
					}
					fourthReviewDirLink(t, candidate, filepath.Join(legacy, "card"))
				}
			}
			got, err := Candidates(repo, "card")
			if kind == "registered" || kind == "registered-alias" {
				if err != nil || len(got) != 1 || !Same(got[0].Path, candidate) {
					t.Fatalf("owned registered candidate = %v, %v; want one physical tree", got, err)
				}
			} else if err == nil || len(got) != 0 {
				t.Fatalf("unsafe candidate %s = %v, %v; want refusal", kind, got, err)
			}
		})
	}
}
