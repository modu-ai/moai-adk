package preference

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDecayScan_DoesNotWriteToRealHome guards against the memory-dir leak.
//
// resolveMemoryDirOverride joins the home directory with a slug derived from
// CLAUDE_PROJECT_DIR. A test that isolates only CLAUDE_PROJECT_DIR still resolves
// the home half to the developer's actual home, so the scan creates a permanent
// directory under the real ~/.claude/projects — and because each t.TempDir() name
// carries a fresh random suffix, every run leaves one more behind. The leak is
// unbounded and silent: no test reports it, and the only symptom is a home
// directory that accumulates entries until reading it slows to a crawl.
//
// The assertion is on the filesystem rather than on a returned path, because the
// returned path alone cannot distinguish a fixed run from a broken one on Unix:
// os.UserHomeDir already reads $HOME there (GOROOT/src/os/file.go), so both
// spellings of the home lookup agree once HOME is set. What actually separates
// the two cases is whether HOME was isolated at all — and the only way to observe
// that is to look for the directory the unisolated derivation would have created.
//
// Falsifiability: deleting the t.Setenv("HOME", tmp) line below makes this test
// fail, because the scan then creates the leaked path this test asserts absent.
func TestDecayScan_DoesNotWriteToRealHome(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	realHome, err := os.UserHomeDir() // captured BEFORE HOME is overridden
	if err != nil {
		t.Fatalf("os.UserHomeDir(): %v", err)
	}

	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	t.Setenv("HOME", tmp)

	// The exact path an unisolated derivation would have produced.
	leaked := filepath.Join(realHome, ".claude", "projects", memorySlug(tmp))

	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: "", // empty → derive from CLAUDE_PROJECT_DIR + home
		nowStr:    time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		force:     true,
	}
	if err := runDecayScan(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runDecayScan: %v (stderr=%q)", err, stderr.String())
	}

	switch _, statErr := os.Stat(leaked); {
	case statErr == nil:
		t.Errorf("decay scan wrote into the real home: %s\n"+
			"the memory-dir derivation escaped its isolation; every run of this "+
			"package leaves one such directory behind permanently", leaked)
	case !os.IsNotExist(statErr):
		t.Fatalf("stat %s: %v", leaked, statErr)
	}
}

// TestUserHomeDir_PrefersHOME pins the helper's resolution order.
//
// This is a contract test, NOT a leak guard: on Unix os.UserHomeDir also reads
// $HOME, so this cannot distinguish the two spellings there. The order is pinned
// because it does differ on Windows, where os.UserHomeDir ignores HOME in favour
// of USERPROFILE/HOMEPATH/HOMEDRIVE — so a Windows test that isolates HOME needs
// the helper to honour it. The leak guard is the test above.
func TestUserHomeDir_PrefersHOME(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	got, err := userHomeDir()
	if err != nil {
		t.Fatalf("userHomeDir(): %v", err)
	}
	if got != tmp {
		t.Errorf("userHomeDir() = %q, want %q (HOME must win over os.UserHomeDir)", got, tmp)
	}
}
