package mx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSubLineFixture writes a Go source fixture into a temp dir and returns
// its path. Used by the @MX:SPEC capture tests.
func writeSubLineFixture(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	src := "package fixture\n\n" + body
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
	return path
}

// TestScannerSpecRef_Capture verifies that an @MX:SPEC sub-line following a
// standalone tag is captured onto that tag's SpecRef field (REQ-MX-ASSOC-001).
func TestScannerSpecRef_Capture(t *testing.T) {
	dir := t.TempDir()
	path := writeSubLineFixture(t, dir, "capture.go", `// @MX:NOTE: a context note
// @MX:SPEC: SPEC-FIXTURE-001
func helper() {}
`)

	scanner := NewScanner()
	tags, err := scanner.ScanFile(path)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d (%+v)", len(tags), tags)
	}
	if tags[0].SpecRef != "SPEC-FIXTURE-001" {
		t.Errorf("SpecRef: expected SPEC-FIXTURE-001, got %q", tags[0].SpecRef)
	}
}

// TestAC006_DanglingSpecRefWarning verifies that an @MX:SPEC sub-line with no
// preceding standalone tag in the file emits a DanglingSpecRef warning, causes
// no panic, and attaches no spurious SpecRef (AC-MX-ASSOC-006).
func TestAC006_DanglingSpecRefWarning(t *testing.T) {
	dir := t.TempDir()
	path := writeSubLineFixture(t, dir, "dangling.go", `// @MX:SPEC: SPEC-X-001
func dangling() {}
`)

	scanner := NewScanner()
	tags, err := scanner.ScanFile(path)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}

	for _, tag := range tags {
		if tag.SpecRef == "SPEC-X-001" {
			t.Errorf("no tag should carry the dangling SpecRef, got %q on %+v", tag.SpecRef, tag)
		}
	}

	warnings := scanner.GetWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "DanglingSpecRef") && strings.Contains(w, "SPEC-X-001") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected DanglingSpecRef warning naming SPEC-X-001, got %v", warnings)
	}
}
