package worktreeroot

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewProjectOwnerUsesDirectoryIdentity(t *testing.T) {
	home, parent := t.TempDir(), t.TempDir()
	first, second := filepath.Join(parent, "App"), filepath.Join(parent, "app")
	for _, p := range []string{first, second} {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}
	a, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(second)
	if err != nil {
		t.Fatal(err)
	}
	key1, err := claim(home, "App", first, false, "App")
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureProjectDir(home, key1, second); (err == nil) != os.SameFile(a, b) {
		t.Errorf("recorded-key owner admission = %v with same identity=%v", err, os.SameFile(a, b))
	}
	key2, err := claim(home, "app", second, false, "app")
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(a, b) != (key1 == key2) {
		t.Fatalf("same identity=%v, keys %q/%q", os.SameFile(a, b), key1, key2)
	}
}

func TestReviewPlanClaimsUnmarkedFolderWithoutWrites(t *testing.T) {
	home := wrHome(t)
	repo := wrRepo(t, "unmarked")
	folder := filepath.Join(wrProjects(home), "unmarked")
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	planned, err := Plan(repo, "tree")
	if err != nil {
		t.Fatalf("read-only plan rejected claimable folder: %v", err)
	}
	if _, err := os.Stat(filepath.Join(folder, markerName)); !os.IsNotExist(err) {
		t.Fatalf("plan wrote marker: %v", err)
	}
	applied, err := PathFor(repo, "tree")
	if err != nil || applied != planned {
		t.Fatalf("plan/apply = %q/%q: %v", planned, applied, err)
	}
}

func TestReviewRecordedKeysCannotEscapeProjects(t *testing.T) {
	for _, key := range []string{"../../shared", ".", "..", `..\shared`} {
		t.Run(key, func(t *testing.T) {
			home := wrHome(t)
			repo := wrRepo(t, "recorded")
			common, err := gitCommonDir(repo)
			if err != nil {
				t.Fatal(err)
			}
			if err := writeRecord(common, key); err != nil {
				t.Fatal(err)
			}
			if got, err := Plan(repo, "tree"); err == nil {
				t.Errorf("Plan accepted unsafe recorded key: %q", got)
			}
			if got, err := Root(repo); err == nil {
				t.Errorf("Root accepted unsafe recorded key: %q", got)
			}
			if _, err := os.Stat(filepath.Join(home, "shared", markerName)); !os.IsNotExist(err) {
				t.Errorf("unsafe key created outside projects: %v", err)
			}
		})
	}
}

func TestReviewPublicRootEmitsRecordedPinNotice(t *testing.T) {
	wrHome(t)
	repo := wrRepo(t, "notice")
	if _, err := Root(repo); err != nil {
		t.Fatal(err)
	}
	pin := filepath.Join(repo, gitStrategyYAML)
	if err := os.MkdirAll(filepath.Dir(pin), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pin, []byte("git_strategy:\n  worktree_project_key: later\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })
	if _, err := PathFor(repo, "tree"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ignored") || !strings.Contains(buf.String(), "later") {
		t.Fatalf("public path helper discarded operator notice: %s", buf.String())
	}
}
