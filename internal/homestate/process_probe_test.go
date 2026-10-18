package homestate

// process_probe_test.go — card t1779 card-review P2-1: the probe must
// SURFACE the ps process's stderr. In a sandbox, ps exits 1 with
// "Operation not permitted" on stderr; a bare exec error string ("exit
// status 1") hides exactly the text the cli-side sandbox-signature match
// needs. A fake ps on PATH drives the real exec → ExitError → stderr path.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessListProbeErrorSurfacesStderr(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "ps")
	body := "#!/bin/sh\necho 'ps: Operation not permitted' >&2\nexit 1\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	err := ProcessListProbeError(os.Getpid())
	if err == nil {
		t.Fatal("fake ps exited 1: want an error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
		t.Fatalf("probe error = %q, want the ps stderr surfaced in the error text", err)
	}
}
