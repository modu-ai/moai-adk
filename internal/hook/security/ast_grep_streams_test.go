package security

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// SPEC-SEC-SCAN-SURFACE-001 — stream-separation regression guard.
//
// `sg scan --json` exits 1 and writes "Error: N error(s) found in code." to
// stderr whenever an error-severity finding exists. Collecting the subprocess
// output with CombinedOutput merges that banner into the JSON body, so
// json.Unmarshal fails and no error-severity finding is ever parsed. These
// tests pin the observable consequence: an error-severity finding must reach
// ScanResult.ErrorCount, and a warning-only file must keep parsing.

// repoRootFromTest walks up to the directory carrying go.mod.
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// shippedSGConfig returns the sgconfig.yml the template distributes.
func shippedSGConfig(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRootFromTest(t), "internal", "template", "templates",
		".moai", "config", "astgrep-rules", "sgconfig.yml")
}

// stagedCorpusFile copies one scan-corpus fixture into a temp dir under the
// given name and returns its path.
func stagedCorpusFile(t *testing.T, fixture, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "scan-corpus", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write staged fixture: %v", err)
	}
	return path
}

func requireSG(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("ast-grep (sg) not on PATH")
	}
}

// TestScanReportsErrorSeverityFindings is the deny-capability guard: an
// error-severity finding must be parsed and counted. It fails whenever the
// subprocess stderr banner is allowed to contaminate the JSON body.
func TestScanReportsErrorSeverityFindings(t *testing.T) {
	requireSG(t)

	scanner := NewASTGrepScanner()
	if !scanner.IsAvailable() {
		t.Skip("ast-grep (sg) not available")
	}

	path := stagedCorpusFile(t, "go_deny_md5.go", "digest.go")
	result, err := scanner.Scan(context.Background(), path, shippedSGConfig(t))
	if err != nil {
		t.Fatalf("Scan returned an error: %v", err)
	}
	if !result.Scanned {
		t.Fatal("expected scanned=true")
	}
	if result.Error != "" {
		t.Fatalf("expected no scan error, got %q", result.Error)
	}
	if result.ErrorCount == 0 {
		t.Fatalf("expected at least one error-severity finding, got ErrorCount=0 (findings=%+v)", result.Findings)
	}
	if !result.HasErrors() {
		t.Error("expected HasErrors()=true")
	}
}

// TestScanReportsWarningSeverityFindings pins the path that already worked
// (rc=0, empty stderr), so the stream separation does not regress it.
func TestScanReportsWarningSeverityFindings(t *testing.T) {
	requireSG(t)

	scanner := NewASTGrepScanner()
	if !scanner.IsAvailable() {
		t.Skip("ast-grep (sg) not available")
	}

	path := stagedCorpusFile(t, "go_warning_only.go", "box.go")
	result, err := scanner.Scan(context.Background(), path, shippedSGConfig(t))
	if err != nil {
		t.Fatalf("Scan returned an error: %v", err)
	}
	if result.WarningCount == 0 {
		t.Fatalf("expected at least one warning-severity finding, got 0 (findings=%+v)", result.Findings)
	}
	if result.ErrorCount != 0 {
		t.Fatalf("expected no error-severity finding, got %d", result.ErrorCount)
	}
}

// TestScanMultipleReportsErrorSeverityFindings covers the fan-out entry point,
// which delegates to Scan and therefore shares the collection path.
func TestScanMultipleReportsErrorSeverityFindings(t *testing.T) {
	requireSG(t)

	scanner := NewASTGrepScanner()
	if !scanner.IsAvailable() {
		t.Skip("ast-grep (sg) not available")
	}

	denying := stagedCorpusFile(t, "go_deny_md5.go", "digest.go")
	clean := stagedCorpusFile(t, "go_clean.go", "clean.go")

	results, err := scanner.ScanMultiple(context.Background(),
		[]string{denying, clean}, shippedSGConfig(t))
	if err != nil {
		t.Fatalf("ScanMultiple returned an error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ErrorCount == 0 {
		t.Fatalf("expected error-severity finding on the denying file, got 0 (findings=%+v)", results[0].Findings)
	}
	if results[1].ErrorCount != 0 {
		t.Fatalf("expected clean file to carry no error finding, got %d", results[1].ErrorCount)
	}
}
