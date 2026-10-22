package worktree

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReviewMigrationUsesRequestedGitRepository(t *testing.T) {
	for _, selector := range []string{"GIT_DIR", "GIT_COMMON_DIR", "GIT_WORK_TREE"} {
		t.Run(selector, func(t *testing.T) {
			a, tree, _ := rfxLegacyWorld(t)
			b, _, _ := rfxLegacyWorld(t)
			value := filepath.Join(b, ".git")
			if selector == "GIT_WORK_TREE" {
				value = b
			}
			t.Setenv(selector, value)
			got, err := primaryRootFrom(a)
			if err != nil || !sameTree(got, a) {
				t.Errorf("requested repository %s resolved to %s with %s: %v", a, got, selector, err)
			}
			out, err := listGitOutput(a, "rev-parse", "--show-toplevel")
			if err != nil || !sameTree(strings.TrimSpace(out), a) {
				t.Errorf("migration Git command inspected ambient repository %q: %v", out, err)
			}
			registered, err := registeredWorktreePaths(a)
			if err != nil || !isRegistered(registered, tree) || isRegistered(registered, b) {
				t.Errorf("migration identity registration used ambient repository: %v, %v", registered, err)
			}
		})
	}
}

func sameTree(a, b string) bool { return canonicalTreePath(a) == canonicalTreePath(b) }

func reviewCaptureManifest(t *testing.T, tree string) string {
	t.Helper()
	digest, err := captureManifest(tree)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestReviewMigrationPlanningErrorsAreNotEmptySuccess(t *testing.T) {
	root := t.TempDir()
	for _, apply := range []bool{false, true} {
		report, err := runMigrationEngine(root, apply, "review")
		if err == nil {
			t.Errorf("apply=%v returned a successful empty plan for an unreadable Git repository: %+v", apply, report)
		}
	}
}

func TestReviewMigrationRequiresReadableManifestBeforeIntent(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	reviewMigrationStateDir(t, repo)
	ref := m6Git(t, tree, "symbolic-ref", "HEAD")
	if err := os.WriteFile(filepath.Join(repo, ".git", filepath.FromSlash(ref)), []byte("damaged-ref\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := listGitOutput(tree, "rev-parse", "HEAD"); err == nil {
		t.Fatal("fixture did not make HEAD unreadable")
	}
	dest := filepath.Join(t.TempDir(), "destination")
	j := &migrationJournal{dir: filepath.Dir(moveListPath(repo)), trigger: "review"}
	out := migrateOneTree(repo, migratePlan{source: tree, destination: dest}, moveListPath(repo), "review", j)
	if out.Outcome != "failed" || !strings.Contains(out.Reason, "manifest") {
		t.Errorf("unreadable mandatory manifest was not refused: %+v", out)
	}
	if len(j.steps) != 0 {
		t.Errorf("manifest failure recorded move intent: %+v", j.steps)
	}
	if _, err := os.Stat(tree); err != nil {
		t.Errorf("manifest failure moved the source: %v", err)
	}
}

func TestReviewMigrationRejectsEmptyInstanceToken(t *testing.T) {
	f := mlNew(t)
	tree := f.add(f.old("empty-token"))
	admin, err := gitDirPointer(tree)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(admin, tokenFileName)
	for _, body := range []string{"", " \r\n"} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if id, err := EnsureInstanceToken(tree); err == nil {
			t.Errorf("accepted interrupted token %q: %+v", body, id)
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != body {
			t.Errorf("refusal changed an existing token: %q, %v", after, err)
		}
	}
}

// The native test executable acts as a Git collaborator. It performs a real
// Git move, restores the stale administrative pointer, then returns failure:
// the exact split state reported when Git cannot update its admin gitdir.
func thirdReviewGitHelper() int {
	args := os.Args[1:]
	probeArgs := args
	if len(probeArgs) >= 2 && probeArgs[0] == "-C" {
		probeArgs = probeArgs[2:]
	}
	if want := os.Getenv("MOAI_REVIEW_FAIL_GIT_PROBE"); want != "" && strings.Join(probeArgs, "\n") == want {
		fmt.Fprintln(os.Stderr, "review: mandatory Git probe could not be read")
		return 94
	}
	if len(probeArgs) >= 2 && probeArgs[0] == "worktree" && probeArgs[1] == "repair" && os.Getenv("MOAI_REVIEW_BLOCK_REPAIR") == "1" {
		fmt.Fprintln(os.Stderr, "review: administrative pointer is still unwritable")
		return 95
	}
	cmd := exec.Command(os.Getenv("MOAI_REVIEW_REAL_GIT"), args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		return 91
	}
	for i := 0; i+3 < len(args); i++ {
		if args[i] != "worktree" || args[i+1] != "move" || args[i+2] != os.Getenv("MOAI_REVIEW_MOVE_SOURCE") {
			continue
		}
		old, dest := args[i+2], args[i+3]
		admin, err := gitDirPointer(dest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 92
		}
		if err := os.WriteFile(filepath.Join(admin, "gitdir"), []byte(filepath.Join(old, ".git")+"\n"), 0600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 93
		}
		fmt.Fprintln(os.Stderr, "review: Git moved the directory before its administrative update failed")
		return 128
	}
	return 0
}

func TestReviewEveryManifestGitProbeMustSucceed(t *testing.T) {
	probes := [][]string{
		{"rev-parse", "HEAD"}, {"rev-parse", "--abbrev-ref", "HEAD"},
		{"rev-list", "HEAD", "--not", "--remotes"}, {"status", "--porcelain"},
		{"diff", "--cached"}, {"diff"}, {"ls-files", "--others", "-z"},
		{"ls-files", "--others", "--ignored", "--exclude-standard", "-z"},
	}
	for _, probe := range probes {
		t.Run(strings.Join(probe, " "), func(t *testing.T) {
			repo, tree, _ := rfxLegacyWorld(t)
			reviewMigrationStateDir(t, repo)
			thirdReviewInstallGit(t, "")
			t.Setenv("MOAI_REVIEW_FAIL_GIT_PROBE", strings.Join(probe, "\n"))
			j := &migrationJournal{dir: filepath.Dir(moveListPath(repo)), trigger: "review"}
			out := migrateOneTree(repo, migratePlan{source: tree, destination: filepath.Join(t.TempDir(), "destination")}, moveListPath(repo), "review", j)
			if out.Outcome != "failed" || !strings.Contains(out.Reason, "manifest") || len(j.steps) != 0 {
				t.Errorf("failed probe %v was not refused before intent: %+v / %+v", probe, out, j.steps)
			}
			if _, err := os.Stat(tree); err != nil {
				t.Errorf("unmeasured tree moved: %v", err)
			}
		})
	}
}

func thirdReviewInstallGit(t *testing.T, source string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(dir, name)
	// Windows keeps every hard link to a loaded executable image locked
	// until the original test process exits. A separate copy is removable
	// after the helper child has been waited for, so TempDir cleanup retains
	// its ordinary resource-leak check.
	linked := false
	if runtime.GOOS != "windows" {
		linked = os.Link(exe, target) == nil
	}
	if !linked {
		in, err := os.Open(exe)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = in.Close() }()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
		if err != nil {
			t.Fatal(err)
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("copy native Git fixture: %v / %v", copyErr, closeErr)
		}
	}
	t.Setenv("MOAI_REVIEW_REAL_GIT", realGit)
	t.Setenv("MOAI_REVIEW_GIT_HELPER", "1")
	t.Setenv("MOAI_REVIEW_MOVE_SOURCE", source)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestReviewMigrationReversesGitMoveThatReportsFailure(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	reviewMigrationStateDir(t, repo)
	before := reviewCaptureManifest(t, tree)
	dest := filepath.Join(t.TempDir(), "destination")
	thirdReviewInstallGit(t, tree)
	j := &migrationJournal{dir: filepath.Dir(moveListPath(repo)), trigger: "review"}
	out := migrateOneTree(repo, migratePlan{source: tree, destination: dest}, moveListPath(repo), "review", j)
	if out.Outcome != "reversed" {
		t.Errorf("partial Git move did not recover to the source: %+v", out)
	}
	if _, err := os.Stat(tree); err != nil {
		t.Errorf("source was not restored: %v", err)
	}
	if _, err := os.Lstat(dest); !os.IsNotExist(err) {
		t.Errorf("destination remains after recovery: %v", err)
	}
	if !m6Registered(t, repo, tree) {
		t.Error("recovery did not restore Git registration")
	}
	if after := reviewCaptureManifest(t, tree); after != before {
		t.Errorf("partial-move recovery changed working state: %s != %s", after, before)
	}
	raw, err := os.ReadFile(moveListPath(repo))
	if err != nil || !bytes.Contains(raw, []byte(`"result":"reversed"`)) {
		t.Errorf("recovery has no reversed receipt: %v\n%s", err, raw)
	}
}

func TestReviewPartialMoveKeepsProtectionUntilRepairCanFinish(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	reviewMigrationStateDir(t, repo)
	before := reviewCaptureManifest(t, tree)
	dest := filepath.Join(t.TempDir(), "destination")
	thirdReviewInstallGit(t, tree)
	t.Setenv("MOAI_REVIEW_BLOCK_REPAIR", "1")
	j := &migrationJournal{dir: filepath.Dir(moveListPath(repo)), trigger: "review"}
	out := migrateOneTree(repo, migratePlan{source: tree, destination: dest}, moveListPath(repo), "review", j)
	if out.Outcome != "failed" || out.MLErr == nil {
		t.Errorf("unrepaired partial move did not expose its recovery failure: %+v", out)
	}
	if !MovedBy(repo, dest) {
		t.Error("partial move lost cleanup protection while its admin pointer was stale")
	}
	t.Setenv("MOAI_REVIEW_BLOCK_REPAIR", "")
	cmd := &cobra.Command{}
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := runMigrateRollback(cmd, repo, j.path, false); err != nil {
		t.Fatalf("recovery retry did not restore the registered source: %v\n%s", err, output.String())
	}
	if after := reviewCaptureManifest(t, tree); after != before {
		t.Errorf("repair retry changed working state: %s != %s", after, before)
	}
	if MovedBy(repo, tree) {
		t.Error("successful rollback did not release migration protection")
	}
}

// The functional worlds own private scratch repositories; external process
// occupancy is an explicit fixture input, independent of host lsof support.
// The planner's direct current-process and session-registry guards still run.
func m6IsolateProcessObservation(t *testing.T) {
	t.Helper()
	original := activeProcessCWDProbe
	activeProcessCWDProbe = func() ([]string, error) { return nil, nil }
	t.Cleanup(func() { activeProcessCWDProbe = original })
}

func TestReviewMigrationPreservesOnProcessInspectionFailure(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	activeProcessCWDProbe = func() ([]string, error) { return nil, fmt.Errorf("review: process observation unavailable") }
	var report *migrationReport
	err := withMigrationLock(repo, func() error { var err error; report, err = runMigrationEngine(repo, true, "review"); return err })
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range report.rows {
		if row.Source == tree {
			found = true
		}
		if row.Source == tree && (row.Outcome != "skipped" || !strings.Contains(row.SkipReason, "cannot inspect active process")) {
			t.Errorf("unknown process observation allowed move: %+v", row)
		}
	}
	if !found {
		t.Error("protected source was not inspected")
	}
	if _, err := os.Stat(tree); err != nil {
		t.Errorf("refused tree changed: %v", err)
	}
	if report.moved != 0 {
		t.Errorf("moved=%d with unknown process occupancy", report.moved)
	}
}
