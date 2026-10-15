package cli

// t1619_integrity_probe_coverage_test.go — card t1619, bundle B2: the unexecuted changed
// lines of update_integrity_probe.go (readProbeFile's missing and non-file branches, and
// the internal-fault recovery of runManagedSurfaceIntegrityProbe). Family: TestIntegrityProbe.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// t1619PanicOnceWriter panics on its first write and records every later write. It is the
// output seam the probe's recovery path has to survive.
type t1619PanicOnceWriter struct {
	panicked bool
	buf      strings.Builder
}

func (w *t1619PanicOnceWriter) Write(p []byte) (int, error) {
	if !w.panicked {
		w.panicked = true
		panic("injected writer fault")
	}
	return w.buf.Write(p)
}

// TestIntegrityProbeReadProbeFileReportsMissingAndNonFile — an absent path reads as
// missing, and a directory where a member file belongs reads as not a file.
func TestIntegrityProbeReadProbeFileReportsMissingAndNonFile(t *testing.T) {
	root := t.TempDir()
	if data, reason := readProbeFile(filepath.Join(root, "absent.json")); data != nil || reason != "missing" {
		t.Fatalf("an absent member must read as missing, got (%q, %q)", data, reason)
	}
	if runtime.GOOS == "windows" {
		t.Skip("the non-blocking open of a directory is judged on unix hosts only")
	}
	dirPath := filepath.Join(root, "member-is-a-directory")
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if data, reason := readProbeFile(dirPath); data != nil || reason != "not a file" {
		t.Fatalf("a directory at a member path must read as not a file, got (%q, %q)", data, reason)
	}
}

// TestIntegrityProbeInternalFaultIsReportedAndDoesNotEscape — a fault while the probe writes
// its rows is recovered, named as an internal probe error, and never escapes the update.
func TestIntegrityProbeInternalFaultIsReportedAndDoesNotEscape(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	projectRoot := t.TempDir()
	manifestPath := filepath.Join(projectRoot, ".moai", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	w := &t1619PanicOnceWriter{}
	runManagedSurfaceIntegrityProbe(w, projectRoot)
	if !strings.Contains(w.buf.String(), "probe internal error") {
		t.Fatalf("a probe fault must surface as an internal-error row, got %q", w.buf.String())
	}
}
