package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixture is the shared test helper for the sync-package fixtures.
func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestScanDec_EmitsRecordPerOccurrence exercises REQ-NS-003 / AC-003:
// a single `@NAV:DEC-AUTH-STRATEGY` line in a design doc yields one binding
// record with the five required fields populated.
func TestScanDec_EmitsRecordPerOccurrence(t *testing.T) {
	root := t.TempDir()
	tech := filepath.Join(root, ".moai", "project", "tech.md")
	writeFixture(t, tech,
		"# Tech\n\nDecision @NAV:DEC-AUTH-STRATEGY: adopt OAuth2.\n")
	pathLog := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	recs, diags, err := ScanDec(root, "abc123", pathLog)
	if err != nil {
		t.Fatalf("ScanDec error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	r := recs[0]
	if r.TokenFamily != FamilyNavDec {
		t.Errorf("TokenFamily = %q, want %q", r.TokenFamily, FamilyNavDec)
	}
	if r.Identifier != "AUTH-STRATEGY" {
		t.Errorf("Identifier = %q, want AUTH-STRATEGY", r.Identifier)
	}
	if !strings.HasSuffix(r.SourcePath, "tech.md") {
		t.Errorf("SourcePath = %q, want suffix tech.md", r.SourcePath)
	}
	if r.LineNumber != 3 {
		t.Errorf("LineNumber = %d, want 3", r.LineNumber)
	}
	if r.CommitSHA != "abc123" {
		t.Errorf("CommitSHA = %q, want abc123", r.CommitSHA)
	}
	if len(diags) != 0 {
		t.Errorf("diagnostics = %v, want empty", diags)
	}
}

// TestScanDec_MalformedEmitsDiagnosticSkipsRecord exercises REQ-NS-017 /
// AC-017 for the @NAV:DEC scanner: `@NAV:DEC-` with empty id is skipped with
// a diagnostic warning, not aborted.
func TestScanDec_MalformedEmitsDiagnosticSkipsRecord(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, ".moai", "project", "tech.md"),
		"# Tech\nBad @NAV:DEC- here\nGood @NAV:DEC-FOO: ok\n")
	pathLog := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	recs, diags, err := ScanDec(root, "abc123", pathLog)
	if err != nil {
		t.Fatalf("ScanDec error: %v", err)
	}
	if len(recs) != 1 || recs[0].Identifier != "FOO" {
		t.Fatalf("records = %+v, want one FOO", recs)
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %v, want 1", diags)
	}
}

// TestScanSym_EmitsFromCodeAndDesign exercises REQ-NS-004 / AC-004:
// @NAV:SYM tokens surface from both Go code and Markdown design docs.
func TestScanSym_EmitsFromCodeAndDesign(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "src", "a", "foo.go"),
		"package foo\n\n// @NAV:SYM:pkg.ParseHeader discusses header parse.\nfunc ParseHeader() {}\n")
	writeFixture(t, filepath.Join(root, ".moai", "project", "structure.md"),
		"# Structure\n\nUses @NAV:SYM:pkg.WriteAtomic for writes.\n")
	pathLog := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	recs, diags, err := ScanSym(root, "deadbeef", pathLog)
	if err != nil {
		t.Fatalf("ScanSym error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %v, want empty", diags)
	}
	var sawGo, sawMd bool
	for _, r := range recs {
		if strings.HasSuffix(r.SourcePath, ".go") && r.Identifier == "pkg.ParseHeader" {
			sawGo = true
		}
		if strings.HasSuffix(r.SourcePath, ".md") && r.Identifier == "pkg.WriteAtomic" {
			sawMd = true
		}
	}
	if !sawGo {
		t.Errorf("missing code-side record; records=%+v", recs)
	}
	if !sawMd {
		t.Errorf("missing design-side record; records=%+v", recs)
	}
}

// TestScanSym_MalformedEmitsDiagnostic exercises REQ-NS-017 / AC-017 for the
// @NAV:SYM scanner.
func TestScanSym_MalformedEmitsDiagnostic(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, ".moai", "project", "structure.md"),
		"# Structure\n\nBad @NAV:SYM: here\nGood @NAV:SYM:pkg.X here\n")
	pathLog := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	recs, diags, err := ScanSym(root, "deadbeef", pathLog)
	if err != nil {
		t.Fatalf("ScanSym error: %v", err)
	}
	if len(recs) != 1 || recs[0].Identifier != "pkg.X" {
		t.Fatalf("records = %+v, want one pkg.X", recs)
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %v, want 1", diags)
	}
}

// TestScanSym_SkipsTestAndVendorFiles exercises REQ-NS-004's exclusion list:
// *_test.go and vendor paths MUST be skipped.
func TestScanSym_SkipsTestAndVendorFiles(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "src", "foo_test.go"),
		"package foo\n// @NAV:SYM:ShouldNotSurface\n")
	writeFixture(t, filepath.Join(root, "vendor", "ext", "ext.go"),
		"package ext\n// @NAV:SYM:ShouldNotSurfaceEither\n")
	writeFixture(t, filepath.Join(root, "src", "real.go"),
		"package src\n// @NAV:SYM:ShouldSurface\n")
	pathLog := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	recs, _, err := ScanSym(root, "deadbeef", pathLog)
	if err != nil {
		t.Fatalf("ScanSym error: %v", err)
	}
	for _, r := range recs {
		if r.Identifier == "ShouldNotSurface" || r.Identifier == "ShouldNotSurfaceEither" {
			t.Errorf("excluded-path record surfaced: %+v", r)
		}
	}
}
