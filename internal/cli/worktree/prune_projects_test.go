package worktree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// pruneFixture builds a temporary MoAI home and returns the projects root.
func pruneFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	root := filepath.Join(home, "worktrees", "projects")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// addProject writes a project folder with an owner marker, aged by age.
func addProject(t *testing.T, root, key, owner string, age time.Duration) string {
	t.Helper()
	dir := filepath.Join(root, key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"schema": 1, "owner": owner})
	if err := os.WriteFile(filepath.Join(dir, ".moai-project.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-age)
	if err := os.Chtimes(dir, old, old); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runPrune(t *testing.T, args ...string) (stdout, stderr string) {
	t.Helper()
	cmd := newPruneProjectsCmd()
	var o, e bytes.Buffer
	cmd.SetOut(&o)
	cmd.SetErr(&e)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("prune-projects %v: %v", args, err)
	}
	return o.String(), e.String()
}

func TestPruneProjects_RemovesOnlyDeadEntries(t *testing.T) {
	root := pruneFixture(t)
	gone := filepath.Join(t.TempDir(), "vanished")
	dead := addProject(t, root, "dead-1", gone, 3*time.Hour)
	liveOwner := t.TempDir()
	live := addProject(t, root, "live-1", liveOwner, 3*time.Hour)
	recent := addProject(t, root, "recent-1", gone, 0)
	withGit := addProject(t, root, "git-1", gone, 3*time.Hour)
	if err := os.MkdirAll(filepath.Join(withGit, "tree", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-3 * time.Hour)
	if err := os.Chtimes(withGit, old, old); err != nil {
		t.Fatal(err)
	}

	out, _ := runPrune(t, "--yes")
	if !strings.Contains(out, "Removed 1 folder(s)") {
		t.Errorf("unexpected output: %s", out)
	}
	if _, err := os.Stat(dead); !os.IsNotExist(err) {
		t.Errorf("dead entry still present: %v", err)
	}
	for name, p := range map[string]string{"live-owner": live, "recent": recent, "inner .git": withGit} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s entry was removed: %v", name, err)
		}
	}
}

func TestPruneProjects_PreviewRemovesNothing(t *testing.T) {
	root := pruneFixture(t)
	dead := addProject(t, root, "dead-1", filepath.Join(t.TempDir(), "vanished"), 3*time.Hour)

	out, _ := runPrune(t)
	if _, err := os.Stat(dead); err != nil {
		t.Fatalf("preview removed the entry: %v", err)
	}
	if !strings.Contains(out, "DEAD dead-1") || !strings.Contains(out, "preview") {
		t.Errorf("preview output lacks the dead entry or the preview notice: %s", out)
	}
}

func TestPruneProjects_JSONReportsVerdicts(t *testing.T) {
	root := pruneFixture(t)
	addProject(t, root, "dead-1", filepath.Join(t.TempDir(), "vanished"), 3*time.Hour)
	addProject(t, root, "live-1", t.TempDir(), 3*time.Hour)

	out, _ := runPrune(t, "--json")
	var got []pruneProjectEntry
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	verdicts := map[string]string{}
	for _, e := range got {
		verdicts[e.Key] = e.Verdict
	}
	if verdicts["dead-1"] != pruneDead || verdicts["live-1"] != pruneKept {
		t.Errorf("verdicts = %v", verdicts)
	}
}

func TestPruneProjects_EmptyOwnerIsKept(t *testing.T) {
	root := pruneFixture(t)
	dir := addProject(t, root, "empty-owner", "", 3*time.Hour)
	runPrune(t, "--yes")
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("entry with an empty owner was removed: %v", err)
	}
}

func TestSweepNoticeOwnerlessProjects(t *testing.T) {
	root := pruneFixture(t)
	gone := filepath.Join(t.TempDir(), "vanished")
	for i := 0; i < ownerlessNoticeThreshold+1; i++ {
		addProject(t, root, fmt.Sprintf("dead-%03d", i), gone, 0)
	}
	sweepMockEnv(t, nil)

	out, err := runSweepCmd(t, nil)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	want := fmt.Sprintf("worktree sweep: %d project entries have no owner; run `moai worktree prune-projects` to remove them", ownerlessNoticeThreshold+1)
	if !strings.Contains(out, want) {
		t.Errorf("sweep output lacks the notice %q:\n%s", want, out)
	}

	// At the threshold the sweep stays silent.
	if err := os.RemoveAll(filepath.Join(root, "dead-000")); err != nil {
		t.Fatal(err)
	}
	out, err = runSweepCmd(t, nil)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if strings.Contains(out, "prune-projects") {
		t.Errorf("notice printed at the threshold:\n%s", out)
	}
}

func TestPruneProjects_RelativeOwnerIsKept(t *testing.T) {
	root := pruneFixture(t)
	dir := addProject(t, root, "relative-owner", "some/relative/repo", 3*time.Hour)
	runPrune(t, "--yes")
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("entry with a relative owner was removed: %v", err)
	}
}
