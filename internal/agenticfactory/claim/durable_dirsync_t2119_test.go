package claim

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"testing"
)

// Card t2119 (#1969, 326 Windows failures): Windows cannot open a directory to
// fsync it — the open or the Sync answers "Access is denied" — so the claim
// store's directory fsync after a rename must be best effort there. It stays a
// hard failure everywhere else, where a directory fsync is the durability step.
func TestDirSyncFailureIsFatal(t *testing.T) {
	denied := &fs.PathError{Op: "sync", Path: `C:\work\docs\checkpoint`, Err: fs.ErrPermission}
	other := errors.New("input/output error")
	cases := []struct {
		name  string
		goos  string
		err   error
		fatal bool
	}{
		{"windows access denied is tolerated", "windows", denied, false},
		{"windows other failures are still fatal", "windows", other, true},
		{"linux access denied is fatal", "linux", denied, true},
		{"darwin access denied is fatal", "darwin", denied, true},
		{"no error is never fatal", "linux", nil, false},
	}
	for _, c := range cases {
		if got := dirSyncFailureIsFatal(c.goos, c.err); got != c.fatal {
			t.Errorf("%s: dirSyncFailureIsFatal(%q, %v) = %v, want %v", c.name, c.goos, c.err, got, c.fatal)
		}
	}
}

// The real directory fsync on this platform is acceptable for a directory that
// exists (on Windows it answers "Access is denied", which is exactly the
// tolerated refusal, so the check is the decision function and not a bare nil),
// and still fails for one that does not (the tolerance is for the Windows
// refusal only, never for a missing directory).
func TestSyncDirDefaultPath(t *testing.T) {
	var y syncer
	if err := y.syncDir(t.TempDir()); dirSyncFailureIsFatal(runtime.GOOS, err) {
		t.Fatalf("syncDir on an existing directory: %v", err)
	}
	if err := y.syncDir(t.TempDir() + string(os.PathSeparator) + "missing"); err == nil {
		t.Fatal("syncDir on a missing directory reported success")
	}
}
