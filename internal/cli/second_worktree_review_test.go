package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestReviewStripWorktreePreservesTailAndRemovesAllSelections(t *testing.T) {
	for _, tc := range []struct{ args, want []string }{
		{[]string{"-w", "first", "--", "prompt", "--worktree", "first"}, []string{"--", "prompt", "--worktree", "first"}},
		{[]string{"-w", "first", "--worktree", "second", "--", "tail"}, []string{"--", "tail"}},
		{[]string{"--model", "-w", "--worktree=first", "--", "--worktree=first"}, []string{"--model", "-w", "--", "--worktree=first"}},
	} {
		if got := stripWorktreeSelection(tc.args); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("strip %q = %q, want %q", tc.args, got, tc.want)
		}
	}
}

func TestReviewLauncherRefusesUnreadableHomeOwnership(t *testing.T) {
	for _, backend := range []string{"cc", "glm"} {
		t.Run(backend, func(t *testing.T) {
			w := m3NewWorld(t)
			t.Chdir(w.repo)
			got := m2LaunchFixture(t, w.repo)
			root, err := worktreeroot.Root(w.repo)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".moai-project.json"), []byte("broken owner"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv(config.DefaultGLMEnvVar, "fixture-key")
			cmd := ccCmd
			if backend == "glm" {
				cmd = glmCmd
			}
			err = cmd.RunE(cmd, []string{"-w", "t8"})
			if err == nil || !strings.Contains(err.Error(), "marker") {
				t.Errorf("unresolvable owner must refuse: %v", err)
			}
			if *got != nil {
				t.Fatalf("backend launched after owner error: %q", *got)
			}
		})
	}
}

func TestReviewGLMPrechecksAnchoredHomeTree(t *testing.T) {
	w := m3NewWorld(t)
	t.Chdir(w.repo)
	got := m2LaunchFixture(t, w.repo)
	if err := os.RemoveAll(w.homeT8); err != nil {
		t.Fatal(err)
	}
	m3Gitf(t, w.repo, "worktree", "add", "-q", "-b", "WT-guarded-home", "--", w.homeT8)
	// No PID is deliberately undetermined, and the existing writer contract
	// refuses rather than assuming this live/unknown lock is dead.
	m3Gitf(t, w.repo, "worktree", "lock", "--reason", "operator live writer", "--", w.homeT8)
	t.Setenv(config.DefaultGLMEnvVar, "fixture-key")
	err := glmCmd.RunE(glmCmd, []string{"-w", "t8"})
	if err == nil || !strings.Contains(err.Error(), worktreeWriterAnchoredSentinel) {
		t.Errorf("GLM bypassed writer admission: %v", err)
	}
	if *got != nil {
		t.Fatalf("backend launched on anchored home tree: %q", *got)
	}
	if _, err := os.Stat(filepath.Join(w.homeT8, ".codex", "hooks.json")); !os.IsNotExist(err) {
		t.Errorf("refused tree was seeded: %v", err)
	}
}

func TestReviewPlainProjectFallbackRequiresAbsentMetadata(t *testing.T) {
	for _, tc := range []string{"plain", "broken-git-file", "git-directory", "ancestor-git", "selected-missing-git"} {
		t.Run(tc, func(t *testing.T) {
			for _, key := range []string{"GIT_DIR", "GIT_COMMON_DIR", "GIT_WORK_TREE"} {
				t.Setenv(key, "")
			}
			root := t.TempDir()
			t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
			switch tc {
			case "broken-git-file":
				if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: missing\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "git-directory":
				if err := os.Mkdir(filepath.Join(root, ".git"), 0700); err != nil {
					t.Fatal(err)
				}
			case "ancestor-git":
				if err := os.Mkdir(filepath.Join(root, ".git"), 0700); err != nil {
					t.Fatal(err)
				}
				root = filepath.Join(root, "nested")
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
			case "selected-missing-git":
				t.Setenv("GIT_DIR", filepath.Join(root, "missing"))
			}
			if got, want := repositoryMetadataAbsent(root), tc == "plain"; got != want {
				t.Fatalf("metadata absent = %v, want %v", got, want)
			}
		})
	}
}

func TestReviewPlainProjectFallbackRequiresWorkingGit(t *testing.T) {
	root := overrideProjectRoot(t)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
	// Missing Git is an inspection failure, not proof of a plain project.
	t.Setenv("PATH", t.TempDir())
	if _, err := existingWorktreeEntryDir([]string{"-w", "plain"}); err == nil {
		t.Fatal("Git execution failure silently became a no-candidate launch")
	}
}
