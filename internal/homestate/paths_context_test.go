package homestate_test

import (
	"context"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFactoryDirContextMatchesLegacyLayouts(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	base := t.TempDir()
	primary, meta, separate, bare, linked := filepath.Join(base, "primary"), filepath.Join(base, "meta"), filepath.Join(base, "separate"), filepath.Join(base, "bare"), filepath.Join(base, "linked")
	separateLinked := filepath.Join(base, "separate-linked")
	alias := filepath.Join(base, "alias")
	for _, args := range [][]string{
		{"init", "-q", primary},
		{"-C", primary, "-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "--allow-empty", "-qm", "seed"},
		{"-C", primary, "worktree", "add", "-qb", "linked", linked},
		{"init", "-q", "--separate-git-dir", meta, separate},
		{"-C", separate, "-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "--allow-empty", "-qm", "seed"},
		{"-C", separate, "worktree", "add", "-qb", "separate-linked", separateLinked},
		{"init", "-q", "--bare", bare},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	if err := os.Symlink(primary, alias); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_DIR", bare)
	t.Setenv("GIT_WORK_TREE", separate)
	for _, home := range []string{"", t.TempDir()} {
		t.Setenv("MOAI_HOME", home)
		for _, root := range []string{primary, linked, separate, separateLinked, meta, bare, alias, t.TempDir()} {
			want, err := homestate.FactoryDir(root)
			if err != nil {
				t.Fatal(err)
			}
			got, err := homestate.FactoryDirContext(context.Background(), root)
			if err != nil {
				t.Fatalf("context path %s: %v", root, err)
			}
			if got != want {
				t.Fatalf("context path %s=%s want %s", root, got, want)
			}
		}
	}
}
