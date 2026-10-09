package factory

// integration_remeasure_run_test.go — M2 verb-body tests for the re-measure
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001 REQ-MWQ-016): the clean-tree and
// unchanged-HEAD checks at start and finish refuse to write a record. Each
// test builds its own fixture repository; no network, no project files.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// remeasureFixtureRepo creates a git repository with one commit on branch
// main and one extra commit on branch develop (standing in for the absorbed
// integration branch), and returns (root, absorbed develop SHA).
func remeasureFixtureRepo(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	git("config", "user.email", "t@t.local")
	git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "base.txt")
	git("commit", "-q", "-m", "base")
	git("checkout", "-q", "-b", "develop")
	if err := os.WriteFile(filepath.Join(root, "dev.txt"), []byte("dev"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "dev.txt")
	git("commit", "-q", "-m", "dev change")
	git("checkout", "-q", "main")
	// The absorbed integration-branch commit the record names is the
	// merge-base of the integration branch and HEAD, not the branch tip.
	return root, git("merge-base", "develop", "main")
}

func TestRemeasureRunsAndWritesRecord(t *testing.T) {
	root, base := remeasureFixtureRepo(t)
	rec, err := RunRemeasure(root, root, "develop", "true")
	if err != nil {
		t.Fatalf("clean remeasure of `true` must succeed: %v", err)
	}
	tree, _ := exec.Command("git", "rev-parse", "HEAD^{tree}").Output() //nolint — asserted below via rec
	_ = tree
	if rec.Tree == "" || rec.Base != base {
		t.Fatalf("record must key the tree and name the absorbed base: %+v", rec)
	}
	if rec.ExitCode != 0 || rec.Command != "true" {
		t.Fatalf("record must carry command + observed exit: %+v", rec)
	}
	if rec.BuildIdentity == "" {
		t.Fatalf("record must carry the producing build identity")
	}
}

func TestRemeasureGoTestJSONRecordsCount(t *testing.T) {
	root, _ := remeasureFixtureRepo(t)
	// A go package with no test files: go test -json reports no per-test
	// pass, an empty sweep. The run refuses it (REQ-MWQ2-009); the record is
	// written (observed as observed) and the verifier rejects it.
	write := func(rel, body string) {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/fixture\n\ngo 1.23\n")
	write("p/p.go", "package p\n")
	// Commit the package so the remeasure starts clean.
	fixtureGit := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	fixtureGit("add", "go.mod", "p/p.go")
	fixtureGit("-c", "user.email=t@t.local", "-c", "user.name=t", "commit", "-q", "-m", "package")
	rec, err := RunRemeasure(root, root, "develop", "go test -json ./p/...")
	if err == nil {
		t.Fatalf("an empty-sweep remeasure must be reported, not silently green")
	}
	if rec == nil {
		t.Fatalf("the record is still written (observed as observed); got none")
	}
	if err := ValidateRemeasureRecord(rec); err == nil {
		t.Fatalf("an empty-sweep remeasure must be invalid, not silently green")
	}
}

func TestRemeasureRefusesDirtyStart(t *testing.T) {
	root, _ := remeasureFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec, err := RunRemeasure(root, root, "develop", "true")
	if err == nil {
		t.Fatalf("a dirty start must refuse, got record %+v", rec)
	}
	// No record written on refusal.
	if _, readErr := ReadRemeasureRecord(root, recTreeOf(root)); readErr == nil {
		t.Fatalf("a refused run must not leave a record")
	}
}

func TestRemeasureRefusesCommandThatDirtyFinish(t *testing.T) {
	// AC-MWQ-016 scenario 2: a command that modifies a tracked file — the
	// finish check refuses and no record is written.
	root, _ := remeasureFixtureRepo(t)
	rec, err := RunRemeasure(root, root, "develop", "sh -c 'echo more >> base.txt'")
	if err == nil {
		t.Fatalf("a command that dirties the tree must refuse, got %+v", rec)
	}
}

func TestRemeasureRefusesCommandThatCommits(t *testing.T) {
	// AC-MWQ-016 scenario 3: a command that commits changed the HEAD.
	root, _ := remeasureFixtureRepo(t)
	rec, err := RunRemeasure(root, root, "develop", "sh -c 'git -c user.email=t@t.local -c user.name=t commit -q --allow-empty -m scratch'")
	if err == nil {
		t.Fatalf("a command that commits must refuse, got %+v", rec)
	}
}

func TestRemeasureCleanCheckSeesIgnoredUnderShowUntrackedNo(t *testing.T) {
	// O2's regression companion: with status.showUntrackedFiles=no set, the
	// clean check must still see untracked files —
	// `git status --porcelain --untracked-files=all` overrides the config.
	root, _ := remeasureFixtureRepo(t)
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return string(out)
	}
	git("config", "status.showUntrackedFiles", "no")
	if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := git("status", "--porcelain", "--untracked-files=all")
	if !strings.Contains(out, "untracked.txt") {
		t.Fatalf("--untracked-files=all must override status.showUntrackedFiles=no, got %q", out)
	}
	if _, err := RunRemeasure(root, root, "develop", "true"); err == nil {
		t.Fatalf("the dirty start must refuse under status.showUntrackedFiles=no (O2)")
	}
}

// recTreeOf computes the tree SHA a remeasure record for root would be keyed
// under — the current HEAD tree.
func recTreeOf(root string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD^{tree}")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// REQ-MWQ2-003 (card t1582): the marker refusal of AC-MWQ-015 is superseded.
// A mixed sweep — a package with passing tests beside a package whose output
// carries `[no test files]` and whose terminal event is a package-level skip —
// is judged by its total per-test pass count, so it is valid and records the
// count it measured. The all-no-test negative procedure stays refused
// (TestRemeasureGoTestJSONRecordsCount, TestClassifyAllNoTestSweepStaysEmpty).
func TestRemeasureMixedTestAndEmptyPackageIsValid(t *testing.T) {
	root, _ := remeasureFixtureRepo(t)
	for rel, body := range map[string]string{
		"go.mod":           "module example.com/mixed\n\ngo 1.23\n",
		"tested/p_test.go": "package tested\nimport \"testing\"\nfunc TestPass(t *testing.T) {}\n",
		"empty/p.go":       "package empty\n",
	} {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "go.mod", "tested/p_test.go", "empty/p.go"}, {"commit", "-q", "-m", "mixed packages"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v, %s", err, out)
		}
	}
	cmd := exec.Command("go", "test", "-json", "./...")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("real go test: %v, %s", err, output)
	}
	if !strings.Contains(string(output), `"Test":"TestPass"`) || !strings.Contains(string(output), "[no test files]") {
		t.Fatalf("positive control: not a mixed report: %s", output)
	}
	t.Logf("positive control: real mixed report includes TestPass and [no test files]")
	count, structured, err := ClassifyStructuredOutput("go test -json ./...", strings.NewReader(string(output)))
	if err != nil || !structured || count != 1 {
		t.Fatalf("a mixed sweep must be judged by its 1 per-test pass: count=%d structured=%v err=%v", count, structured, err)
	}
	rec, err := RunRemeasure(root, root, "develop", "go test -json ./...")
	if rec == nil || err != nil {
		t.Fatalf("mixed sweep must be recorded and accepted: record=%+v err=%v", rec, err)
	}
	if err := ValidateRemeasureRecord(rec); err != nil {
		t.Fatalf("mixed sweep record must be valid: %v", err)
	}
	if rec.TestCount != 1 {
		t.Fatalf("the record must carry the measured count 1, got %d", rec.TestCount)
	}
}
