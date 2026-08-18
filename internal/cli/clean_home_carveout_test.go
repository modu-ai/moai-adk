package cli

// SPEC-V3R6-MOAI-CLEAN-HOME-001 REQ-MCH-005/006 + AC-MCH-006/007/009/012:
// the carve-out guard suite, releases current+N survival, MOAI_HOME redirect,
// and the project-scope clean regression.
//
// Container-granularity decision (audit residual R1), documented here as the
// binding test-side statement: when an aged backups/removed-* directory
// contains ANY carved-out file, clean --home skips the WHOLE directory — it
// never partially deletes a backup container. Whole-dir skip is the chosen
// granularity because a partial delete of a backup is exactly the
// "restorable snapshot with a hole in it" failure this SPEC exists to
// prevent; over-protection is always preferred over over-deletion.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// TestCleanHomeCarveOut_PathPredicate is the REQ-MCH-005 dedicated predicate
// test: recursive any-segment matching at every depth, carve-out names,
// credentials* prefix, and the allowlisted-but-carved negative cases.
func TestCleanHomeCarveOut_PathPredicate(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
		note string
	}{
		{"projects/foo.jsonl", true, "root projects/"},
		{"config/sections/state.yaml", true, "root config/"},
		{"state/runs/x", true, "root state/"},
		{"worktrees/feature/x", true, "root worktrees/"},
		{"mcp/cache/x", true, "root mcp/"},
		{"bin/moai-v3", true, "root bin/"},
		{"search/index/x", true, "root search/"},
		{"studio/x", true, "root studio/"},
		{"plugins/x/repo", true, "root plugins/"},
		{"credentials.yaml", true, "credentials file"},
		{"credentials-prod.yaml", true, "credentials* prefix"},
		{"launch.yaml", true, "launch.yaml"},
		{"preferences.yaml", true, "preferences.yaml"},
		{"claude-profiles/p/projects/x", true, "per-profile projects/"},
		{"claude-profiles/p/config/x", true, "per-profile config/"},
		{"claude-profiles/p/state/x", true, "per-profile state/"},
		{"claude-profiles/p/credentials.yaml", true, "per-profile credentials"},
		{"claude-profiles/p/launch.yaml", true, "per-profile launch.yaml"},
		{"claude-profiles/p/preferences.yaml", true, "per-profile preferences.yaml"},
		{"claude-profiles/p/worktrees/x", true, "per-profile worktrees/"},
		{"claude-profiles/p/mcp/x", true, "per-profile mcp/"},
		{"claude-profiles/p/bin/x", true, "per-profile bin/"},
		{"claude-profiles/p/search/x", true, "per-profile search/"},
		{"claude-profiles/p/studio/x", true, "per-profile studio/"},
		{"claude-profiles/p/plugins/x", true, "per-profile plugins/ (REQ-MCH-006)"},
		{"backups/removed-old/credentials.yaml", true, "carve-out wins inside allowlisted container"},
		{"backups/removed-old/nested/state/key.json", true, "carve-out at nested depth inside removed-*"},
		{"claude-profiles/p/debug/bin", true, "carved dir name inside allowlisted debug/"},
		{"claude-profiles/p/debug/old.log", false, "plain aged debug entry is deletable"},
		{"logs/old-run.log", false, "aged root log is deletable"},
		{"releases/moai-v1.2.3-darwin-arm64", false, "old release binary is deletable"},
		{"backups/removed-old/plain.yaml", false, "plain file inside aged removed-* is deletable"},
		{"", true, "empty relPath (root itself) is never deletable"},
		{".", true, "dot relPath (root itself) is never deletable"},
	}
	for _, tc := range cases {
		if got := isCarvedOut(tc.rel); got != tc.want {
			t.Errorf("isCarvedOut(%q) = %v, want %v (%s)", tc.rel, got, tc.want, tc.note)
		}
	}
}

// TestCleanHomeCarveOut_ForcePreservesCarvedSegments (AC-MCH-006): after
// --force on a fixture planted with every carved-out segment — root level,
// per-profile, and a credentials* file inside an aged backups/removed-*
// container — all of them still exist.
func TestCleanHomeCarveOut_ForcePreservesCarvedSegments(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")

	carvedDirs := []string{
		"projects", "config", "state", "worktrees", "mcp", "bin", "search", "studio", "plugins",
	}
	carvedFiles := []string{"credentials.yaml", "launch.yaml", "preferences.yaml"}

	// Root-level matrix.
	for _, d := range carvedDirs {
		writeHomeFixtureFile(t, filepath.Join(root, d, "keep.dat"), 10, agedTime(t))
	}
	for _, f := range carvedFiles {
		writeHomeFixtureFile(t, filepath.Join(root, f), 10, agedTime(t))
	}
	// Per-profile matrix.
	for _, d := range carvedDirs {
		writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "p1", d, "keep.dat"), 10, agedTime(t))
	}
	for _, f := range carvedFiles {
		writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "p1", f), 10, agedTime(t))
	}
	// Carved dir name nested inside the allowlisted debug/ container.
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "p1", "debug", "bin", "inner.log"), 10, agedTime(t))
	// R1: credentials* inside an aged removed-* — the WHOLE dir must survive.
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "removed-mixed", "credentials.yaml"), 10, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "removed-mixed", "plain-sibling.yaml"), 10, agedTime(t))
	if err := os.Chtimes(filepath.Join(root, "backups", "removed-mixed"), agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes removed-mixed: %v", err)
	}
	// A plain aged removed-* proves --force actually ran and deleted.
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "removed-plain", "x.yaml"), 10, agedTime(t))
	if err := os.Chtimes(filepath.Join(root, "backups", "removed-plain"), agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes removed-plain: %v", err)
	}

	p, _, _ := newHomeTestPrinter()
	if err := runCleanHome(p, true); err != nil {
		t.Fatalf("runCleanHome(force): %v", err)
	}

	// Control: the plain aged removed-* was deleted, so --force demonstrably ran.
	if _, err := os.Stat(filepath.Join(root, "backups", "removed-plain")); !os.IsNotExist(err) {
		t.Fatalf("control assertion failed: removed-plain should be deleted (stat err=%v)", err)
	}

	// Every carved-out segment survives.
	mustExist := []string{}
	for _, d := range carvedDirs {
		mustExist = append(mustExist, filepath.Join(root, d, "keep.dat"))
		mustExist = append(mustExist, filepath.Join(root, "claude-profiles", "p1", d, "keep.dat"))
	}
	for _, f := range carvedFiles {
		mustExist = append(mustExist, filepath.Join(root, f))
		mustExist = append(mustExist, filepath.Join(root, "claude-profiles", "p1", f))
	}
	mustExist = append(mustExist,
		filepath.Join(root, "claude-profiles", "p1", "debug", "bin", "inner.log"),
		filepath.Join(root, "backups", "removed-mixed", "credentials.yaml"),
		filepath.Join(root, "backups", "removed-mixed", "plain-sibling.yaml"),
	)
	for _, path := range mustExist {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("carved-out entry must survive --force: %s (%v)", path, err)
		}
	}
}

// TestCleanHomeCarveOut_ReleasesKeepCurrentPlusNewest (AC-MCH-007): the
// current-version binary (and its rc builds) plus the DefaultReleaseKeep
// newest of the rest survive; only older ones go.
func TestCleanHomeCarveOut_ReleasesKeepCurrentPlusNewest(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	current := version.GetVersion()

	// Current binary + an rc of current — always protected regardless of age.
	writeHomeFixtureFile(t, filepath.Join(root, "releases", "moai-"+current+"-darwin-arm64"), 40, time.Now().AddDate(0, 0, -90))
	writeHomeFixtureFile(t, filepath.Join(root, "releases", "moai-"+current+"-rc.1-darwin-arm64"), 40, time.Now().AddDate(0, 0, -89))
	// Five older, non-current binaries: v0.0.5 newest ... v0.0.1 oldest.
	for i := 1; i <= 5; i++ {
		writeHomeFixtureFile(t,
			filepath.Join(root, "releases", fmt.Sprintf("moai-v0.0.%d-darwin-arm64", i)),
			40, time.Now().AddDate(0, 0, -(6-i)))
	}
	writeHomeFixtureFile(t, filepath.Join(root, "releases", "version.json"), 60, time.Time{})

	p, _, _ := newHomeTestPrinter()
	if err := runCleanHome(p, true); err != nil {
		t.Fatalf("runCleanHome(force): %v", err)
	}

	survive := []string{
		"moai-" + current + "-darwin-arm64",
		"moai-" + current + "-rc.1-darwin-arm64",
		"moai-v0.0.5-darwin-arm64",
		"moai-v0.0.4-darwin-arm64",
		"moai-v0.0.3-darwin-arm64",
		"version.json",
	}
	for _, name := range survive {
		if _, err := os.Stat(filepath.Join(root, "releases", name)); err != nil {
			t.Errorf("release must survive --force: %s (%v)", name, err)
		}
	}
	deleted := []string{"moai-v0.0.1-darwin-arm64", "moai-v0.0.2-darwin-arm64"}
	for _, name := range deleted {
		if _, err := os.Stat(filepath.Join(root, "releases", name)); !os.IsNotExist(err) {
			t.Errorf("release beyond current+keep must be deleted: %s (stat err=%v)", name, err)
		}
	}
}

// TestCleanHomeCarveOut_MOAIHomeRedirect (AC-MCH-012, audit D7): an absolute
// MOAI_HOME pointing at a second fixture tree redirects clean --home --force
// at that tree, and the default-HOME fixture stays untouched — the dependency
// SPEC's core promise.
func TestCleanHomeCarveOut_MOAIHomeRedirect(t *testing.T) {
	defaultHome := hermeticHomeEnv(t)
	defaultRoot := filepath.Join(defaultHome, ".moai")
	writeHomeFixtureFile(t, filepath.Join(defaultRoot, "logs", "old.log"), 10, agedTime(t))

	// Second fixture tree under an absolute MOAI_HOME override. MOAI_HOME is a
	// full REPLACEMENT of the ~/.moai root (paths.MoaiHome returns it
	// verbatim), so the override root is the tree itself — no .moai segment.
	overrideRoot := t.TempDir()
	writeHomeFixtureFile(t, filepath.Join(overrideRoot, "logs", "old.log"), 20, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(overrideRoot, "logs", "fresh.log"), 20, time.Time{})
	t.Setenv("MOAI_HOME", overrideRoot)

	cmd := newCleanCmd()
	cmd.SetArgs([]string{"--home", "--force"})
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("clean --home --force under MOAI_HOME: %v", err)
	}

	// The override tree was cleaned (aged gone, fresh survives).
	if _, err := os.Stat(filepath.Join(overrideRoot, "logs", "old.log")); !os.IsNotExist(err) {
		t.Errorf("aged entry under the MOAI_HOME override should be deleted (stat err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(overrideRoot, "logs", "fresh.log")); err != nil {
		t.Errorf("fresh entry under the override should survive: %v", err)
	}
	// The default-HOME tree is untouched.
	if _, err := os.Stat(filepath.Join(defaultRoot, "logs", "old.log")); err != nil {
		t.Errorf("default-HOME fixture must stay untouched under MOAI_HOME redirect: %v", err)
	}
	if !strings.Contains(errBuf.String(), "Deleted") {
		t.Errorf("force run under override should report deletions, got %q", errBuf.String())
	}
}

// TestClean_ProjectScopeRegression (AC-MCH-009): the project-scope `moai
// clean` (no --home) keeps its SPEC-V3R2-RT-004 contract — runs/ retention
// from the PROJECT state.yaml, dry-run default, --force deletes.
func TestClean_ProjectScopeRegression(t *testing.T) {
	project := t.TempDir()
	writeHomeFixtureFile(t, filepath.Join(project, ".moai", "state", "runs", "run-old", "trace.json"), 10, agedTime(t))
	// runClean ages a run by the run DIR's own mtime — pin that too.
	if err := os.Chtimes(filepath.Join(project, ".moai", "state", "runs", "run-old"), agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes run-old: %v", err)
	}
	writeHomeFixtureFile(t, filepath.Join(project, ".moai", "state", "runs", "run-fresh", "trace.json"), 10, time.Time{})
	projectStateDir := filepath.Join(project, ".moai", "config", "sections")
	if err := os.MkdirAll(projectStateDir, 0o755); err != nil {
		t.Fatalf("mkdir project sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectStateDir, "state.yaml"),
		[]byte("state:\n  retention_days: 5\n"), 0o644); err != nil {
		t.Fatalf("write project state.yaml: %v", err)
	}
	t.Chdir(project)

	// Dry-run default: nothing deleted, candidate listed.
	p, _, errBuf := newHomeTestPrinter()
	if err := runClean(p, false); err != nil {
		t.Fatalf("runClean(dry-run): %v", err)
	}
	if !strings.Contains(errBuf.String(), "[dry-run]") {
		t.Errorf("project-scope dry-run output should carry [dry-run], got %q", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(project, ".moai", "state", "runs", "run-old")); err != nil {
		t.Fatalf("project-scope dry-run must not delete: %v", err)
	}

	// --force: aged run deleted, fresh survives.
	p2, _, _ := newHomeTestPrinter()
	if err := runClean(p2, true); err != nil {
		t.Fatalf("runClean(force): %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".moai", "state", "runs", "run-old")); !os.IsNotExist(err) {
		t.Errorf("aged project run should be deleted (stat err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".moai", "state", "runs", "run-fresh")); err != nil {
		t.Errorf("fresh project run should survive: %v", err)
	}
}
