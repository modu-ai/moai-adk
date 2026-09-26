package gitenv

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// fixtureExec matches a test that starts the git binary by name. A package
// with such a test spawns git children, and every one of them inherits the
// test binary's environment.
var fixtureExec = regexp.MustCompile(`exec\.Command(Context)?\([^)]*"git"`)

// scrubCall is the TestMain line that removes the inherited repository.
var scrubCall = regexp.MustCompile(`gitenv\.ScrubProcess\(\)`)

// scanTestDir reports whether a directory's tests start git, and whether one
// of its test files carries a TestMain that calls ScrubProcess.
func scanTestDir(dir string) (execsGit, scrubbed bool, err error) {
	files, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
	if err != nil {
		return false, false, err
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return false, false, err
		}
		if fixtureExec.Match(b) {
			execsGit = true
		}
		if strings.Contains(string(b), "func TestMain(") && scrubCall.Match(b) {
			scrubbed = true
		}
	}
	return execsGit, scrubbed, nil
}

// moduleRoot walks up from the package directory to the directory holding
// go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the gitenv package")
		}
		dir = parent
	}
}

// Every package whose tests start git must scrub the inherited repository in
// its TestMain. Scrubbing each call site instead reaches only the sites
// someone remembered; this check is what makes a new git fixture in an
// unscrubbed package fail here rather than write into a caller's repository.
func TestFixturePackagesScrubProcess(t *testing.T) {
	root := moduleRoot(t)

	var fixturePkgs, missing []string
	for _, top := range []string{"internal", "pkg", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && path == filepath.Join(root, top) {
					return filepath.SkipDir
				}
				return err
			}
			if !d.IsDir() {
				return nil
			}
			if name := d.Name(); name == "testdata" || (strings.HasPrefix(name, ".") && path != filepath.Join(root, top)) {
				return filepath.SkipDir
			}
			execsGit, scrubbed, err := scanTestDir(path)
			if err != nil {
				return err
			}
			if !execsGit {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			fixturePkgs = append(fixturePkgs, rel)
			if !scrubbed {
				missing = append(missing, rel)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", top, err)
		}
	}

	// Premise: the scan found the packages it exists to guard. A detector that
	// matched nothing would pass this test for every tree.
	if !slices.Contains(fixturePkgs, "internal/binlag") {
		t.Fatalf("scan did not find internal/binlag among git-fixture packages (found %d: %v); the detector is blind", len(fixturePkgs), fixturePkgs)
	}
	for _, pkg := range missing {
		t.Errorf("%s: tests start git but no TestMain calls gitenv.ScrubProcess()", pkg)
	}
}

// The detector itself: a git fixture without the TestMain line is reported,
// and the same package with it is not. Without this, a regex that stopped
// matching TestMain would read every package as scrubbed.
func TestScanTestDir_Detects(t *testing.T) {
	dir := t.TempDir()
	// Split so this file does not itself read as a git fixture to the scan.
	fixture := "package p\n\nimport \"os/exec\"\n\nfunc f() { _ = exec.Command(\"g" + "it\", \"init\") }\n"
	if err := os.WriteFile(filepath.Join(dir, "a_test.go"), []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	execsGit, scrubbed, err := scanTestDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !execsGit || scrubbed {
		t.Fatalf("unscrubbed fixture: execsGit=%v scrubbed=%v, want true false", execsGit, scrubbed)
	}

	mainFile := "package p\n\nfunc TestMain(m *testing.M) { _ = gitenv.ScrubProcess() }\n"
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(mainFile), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, scrubbed, _ := scanTestDir(dir); !scrubbed {
		t.Fatal("TestMain calling gitenv.ScrubProcess() was not recognised")
	}
}
