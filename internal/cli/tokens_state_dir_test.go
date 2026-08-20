package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustMkdirAll is a t.Fatal-on-error helper so the cases below stay readable.
func mustMkdirAll(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
}

// TestFindStateDirFromWalksUp pins the walk-up contract that decides where
// `moai tokens record` writes its ledger when no state dir is supplied.
//
// This is the logic behind a windows-only CI failure: the walk climbs from the
// working directory to the filesystem root, and on Windows the temp directory
// lives INSIDE the user profile (%USERPROFILE%\AppData\Local\Temp), so the
// climb passes through the home directory and picks up any ~/.moai/state on
// the machine. On Linux and macOS the temp root (/tmp, /var/folders) is not
// under $HOME, so the same climb never sees it — hence a test that leans on
// this resolution passes on two platforms and fails on the third.
//
// The layout below is windows-shaped on purpose; the assertions hold on every
// platform because they only compare paths inside a directory the test owns.
func TestFindStateDirFromWalksUp(t *testing.T) {
	t.Run("an ancestor state dir wins over the starting directory", func(t *testing.T) {
		root := t.TempDir()
		ancestor := mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
		start := mustMkdirAll(t, filepath.Join(root, "AppData", "Local", "Temp", "TestSomething", "001"))

		got, err := findStateDirFrom(start)
		if err != nil {
			t.Fatalf("findStateDirFrom: %v", err)
		}
		if got != ancestor {
			t.Errorf("got %q, want the ancestor %q", got, ancestor)
		}
		if got == filepath.Join(start, ".moai", "state") {
			t.Errorf("resolution stayed at the starting directory; the walk-up did not happen")
		}
	})

	t.Run("the starting directory wins over an ancestor", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
		start := mustMkdirAll(t, filepath.Join(root, "AppData", "Local", "Temp", "TestSomething", "001"))
		own := mustMkdirAll(t, filepath.Join(start, ".moai", "state"))

		got, err := findStateDirFrom(start)
		if err != nil {
			t.Fatalf("findStateDirFrom: %v", err)
		}
		if got != own {
			t.Errorf("got %q, want the starting directory's own %q", got, own)
		}
	})

	t.Run("nothing inside the tree is claimed when the tree holds no state dir", func(t *testing.T) {
		root := t.TempDir()
		start := mustMkdirAll(t, filepath.Join(root, "AppData", "Local", "Temp", "TestSomething", "001"))

		// The walk continues past the test's own directory into whatever the
		// machine has above it, so the result is not asserted directly — only
		// that nothing inside `root` was claimed. Asserting an error here
		// would be asserting the absence of ~/.moai/state on the CI runner,
		// which is exactly the environment dependency this test documents.
		got, err := findStateDirFrom(start)
		if err == nil && strings.HasPrefix(got, root) {
			t.Errorf("got %q inside the test tree, which holds no .moai/state", got)
		}
	})
}

// TestRunTokensRecordHonoursInjectedStateDir pins the seam the ledger tests
// depend on: an explicit StateDir is used verbatim, with no walk-up and no
// working-directory lookup. Without this, where the ledger lands is a property
// of the machine rather than of the call.
func TestRunTokensRecordHonoursInjectedStateDir(t *testing.T) {
	root := t.TempDir()

	// A state dir an unbounded walk-up would find, sitting above the working
	// directory — the shape that hijacked the ledger on windows.
	decoy := mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
	workdir := mustMkdirAll(t, filepath.Join(root, "AppData", "Local", "Temp", "TestSomething", "001"))
	t.Chdir(workdir)

	want := filepath.Join(root, "elsewhere", ".moai", "state")
	path := writeTranscriptFixture(t, defaultFixtureLines())

	if _, err := runTokensRecord(tokensRecordOpts{
		Transcript: path,
		StateDir:   want,
	}); err != nil {
		t.Fatalf("runTokensRecord: %v", err)
	}

	if _, err := os.Stat(filepath.Join(want, tokensLedgerFilename)); err != nil {
		t.Errorf("ledger not at the injected state dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(decoy, tokensLedgerFilename)); !os.IsNotExist(err) {
		t.Errorf("ledger leaked to the ancestor state dir; stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, ".moai", "state", tokensLedgerFilename)); !os.IsNotExist(err) {
		t.Errorf("ledger leaked to the working directory; stat err = %v", err)
	}
}
