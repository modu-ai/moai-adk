// mcp_server_staleness_test.go — card t1632 card-review repair P1-3: the
// staleness check compares the installed binary against the identity of the
// executable CAPTURED AT SERVER START, not a fresh stat of the same path —
// an in-place reinstall makes both end-of-call stats read the NEW file, so
// the stale server judged itself fresh.
package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withStalenessSeams(t *testing.T, captured binaryIdentity, installedPath string) {
	t.Helper()
	if resolved, err := filepath.EvalSymlinks(installedPath); err == nil {
		installedPath = resolved
	}
	captured.Path = installedPath
	prevCap, prevLook := servingBinaryIdentity, moaiLookPath
	servingBinaryIdentity = captured
	moaiLookPath = func(string) (string, error) { return installedPath, nil }
	t.Cleanup(func() { servingBinaryIdentity, moaiLookPath = prevCap, prevLook })
}

func writeFileOfSize(t *testing.T, path string, size int, modTime time.Time) {
	t.Helper()
	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
}

// TestServerStaleness_InPlaceReinstallReadsStale pins the defect fix: the
// serving identity captured AT START (the old build) differs from the
// currently-installed file (the reinstall overwrote the same path), so the
// result is stale — even though a fresh stat of the serving path would now
// read the new file and have called it fresh.
func TestServerStaleness_InPlaceReinstallReadsStale(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "moai")
	startedAt := time.Now().Add(-time.Hour)
	writeFileOfSize(t, installed, 4096, startedAt.Add(time.Minute)) // the reinstall

	captured := binaryIdentity{Path: installed, Size: 2048, ModTime: startedAt, Captured: true} // what the server started with
	withStalenessSeams(t, captured, installed)

	got := serverStalenessInfo()
	if !got.Stale {
		t.Fatal("an in-place reinstall judged the stale server fresh — the start-time identity was not used (card-review P1-3)")
	}
}

// TestServerStaleness_UnchangedInstallReadsFresh pins the positive space: an
// installed binary identical to the start-time capture is not stale.
func TestServerStaleness_UnchangedInstallReadsFresh(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "moai")
	startedAt := time.Now().Add(-time.Hour)
	writeFileOfSize(t, installed, 4096, startedAt) // untouched since start

	captured := binaryIdentity{Path: installed, Size: 4096, ModTime: startedAt, Captured: true}
	withStalenessSeams(t, captured, installed)

	got := serverStalenessInfo()
	if got.Stale {
		t.Fatalf("an unchanged install read as stale: %+v", got)
	}
}

// TestServerStaleness_SameBuildAtDifferentPathReadsFresh pins the identity
// rule (t1632): staleness judges BUILD identity — size and mtime, the data
// the start-time capture already carries — never the executable's PATH. The
// same build bytes installed at a second PATH location are the same build;
// judging them by path alone refused every write-capable call on a server
// that was in fact current, and reconnecting with the same config reproduced
// the refusal.
func TestServerStaleness_SameBuildAtDifferentPathReadsFresh(t *testing.T) {
	dir := t.TempDir()
	installed := filepath.Join(dir, "moai-installed") // what LookPath resolves now
	serving := filepath.Join(dir, "moai-serving")     // the build this process started with
	startedAt := time.Now().Add(-time.Hour)
	writeFileOfSize(t, installed, 4096, startedAt) // same build bytes, second location
	writeFileOfSize(t, serving, 4096, startedAt)

	prevCap, prevLook := servingBinaryIdentity, moaiLookPath
	servingBinaryIdentity = binaryIdentity{Path: serving, Size: 4096, ModTime: startedAt, Captured: true}
	moaiLookPath = func(string) (string, error) { return installed, nil }
	t.Cleanup(func() { servingBinaryIdentity, moaiLookPath = prevCap, prevLook })

	got := serverStalenessInfo()
	if got.Stale {
		t.Fatalf("an identical build at a different PATH location judged the current server stale: %+v", got)
	}
}
