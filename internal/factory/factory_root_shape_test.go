package factory

import (
	"os"
	"path/filepath"
	"testing"
)

// Invalid project roots must fail before the home-keyed admission/roster can
// create state. In particular Windows may report a file ancestor as absent,
// so a later optional workers.json read cannot establish this precondition.
func TestClaimFactoryLaneRejectsFileRootBeforeHomeMutation(t *testing.T) {
	for _, child := range []bool{false, true} {
		name := "file root"
		if child {
			name = "below file root"
		}
		t.Run(name, func(t *testing.T) {
			blocker := filepath.Join(t.TempDir(), "regular-file")
			if err := os.WriteFile(blocker, []byte("original blocker\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			root := blocker
			if child {
				root = filepath.Join(blocker, "missing-child")
			}
			home := filepath.Join(t.TempDir(), "uncreated-home-state")
			t.Setenv("MOAI_HOME", home)
			claim, err := ClaimFactoryLane(root, "lane-7", false, os.Getpid(), "root-shape", func(int) bool { return true })
			if err == nil {
				t.Errorf("non-directory project root admitted a claim: %+v", claim)
			}
			if _, err := os.Lstat(home); !os.IsNotExist(err) {
				t.Errorf("invalid project root changed home state before refusal: %v", err)
			}
			if data, err := os.ReadFile(blocker); err != nil || string(data) != "original blocker\n" {
				t.Errorf("invalid-root blocker changed: %q, err=%v", data, err)
			}
		})
	}
}

func TestClaimFactoryLaneKeepsValidNonGitRootShapes(t *testing.T) {
	for _, shape := range []string{"directory", "dot alias", "directory alias", "empty means cwd"} {
		t.Run(shape, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home-state"))
			switch shape {
			case "dot alias":
				root += string(os.PathSeparator) + "."
			case "directory alias":
				root = factoryRootAlias(t, root)
			case "empty means cwd":
				t.Chdir(root)
				root = ""
			}
			claim, err := ClaimFactoryLane(root, "lane-7", false, os.Getpid(), "root-shape", func(int) bool { return true })
			if err != nil || claim.Label != "lane-7" {
				t.Fatalf("valid non-Git project root refused: claim=%+v err=%v", claim, err)
			}
		})
	}
}
