package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNavigatorEnrich_EmitsFilesWhenCapabilityMapExists is AC-NT-001: when
// capability-map.md exists, the command writes both capability-symbols.md and
// .json with non-empty rows.
func TestNavigatorEnrich_EmitsFilesWhenCapabilityMapExists(t *testing.T) {
	root := t.TempDir()
	// capability-map.md + an implementation-path with a Go source file.
	enrichWriteFile(t, filepath.Join(root, ".moai", "project", "navigator", "capability-map.md"),
		"| spec-id | title | status | implementation-path | commit-sha | captured-at |\n"+
			"|---------|-------|--------|---------------------|------------|-------------|\n"+
			"| SPEC-X-001 | X | implemented | src/x | abc123 | 2026-01-01 |\n")
	enrichWriteFile(t, filepath.Join(root, "src", "x", "a.go"),
		"package x\nfunc Foo() {}\ntype Bar struct{}\n")

	if err := runNavigatorEnrich(root, "", ""); err != nil {
		t.Fatalf("runNavigatorEnrich error: %v", err)
	}
	md := enrichReadFile(t, filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.md"))
	if !strings.Contains(string(md), "SPEC-X-001") {
		t.Errorf("md missing SPEC-X-001:\n%s", md)
	}
	if !strings.Contains(string(md), "Capability Symbols") {
		t.Errorf("md missing header")
	}
	js := enrichReadFile(t, filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.json"))
	var doc map[string]any
	if err := json.Unmarshal(js, &doc); err != nil {
		t.Fatalf("json invalid: %v\n%s", err, js)
	}
	rows, _ := doc["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
}

// TestNavigatorEnrich_AbsenceIsGraceful is AC-NT-002: when capability-map.md
// is absent, the command exits nil and writes NO output files.
func TestNavigatorEnrich_AbsenceIsGraceful(t *testing.T) {
	root := t.TempDir()
	if err := runNavigatorEnrich(root, "", ""); err != nil {
		t.Fatalf("runNavigatorEnrich error: %v", err)
	}
	outDir := filepath.Join(root, ".moai", "project", "codemaps")
	if _, err := os.Stat(filepath.Join(outDir, "capability-symbols.md")); err == nil {
		t.Errorf("capability-symbols.md should NOT exist when capability-map absent")
	}
	if _, err := os.Stat(filepath.Join(outDir, "capability-symbols.json")); err == nil {
		t.Errorf("capability-symbols.json should NOT exist when capability-map absent")
	}
}

// TestNavigatorEnrich_AtomikWriteBarrier is AC-NT-013: the
// NAVIGATOR_PRE_RENAME_BARRIER test hook blocks the rename until removed, so
// no reader observes a partial file. This exercises the barrier path itself.
func TestNavigatorEnrich_AtomicWriteBarrier(t *testing.T) {
	root := t.TempDir()
	enrichWriteFile(t, filepath.Join(root, ".moai", "project", "navigator", "capability-map.md"),
		"| spec-id | title | status | implementation-path | commit-sha | captured-at |\n"+
			"|---------|-------|--------|---------------------|------------|-------------|\n"+
			"| SPEC-X-001 | X | implemented | src/x | abc123 | 2026-01-01 |\n")
	enrichWriteFile(t, filepath.Join(root, "src", "x", "a.go"), "package x\nfunc Foo() {}\n")

	barrier := filepath.Join(root, "barrier.flag")
	t.Setenv("NAVIGATOR_PRE_RENAME_BARRIER", barrier)

	done := make(chan error, 1)
	go func() { done <- runNavigatorEnrich(root, "", "") }()

	// While the barrier exists, the .tmp file should be present and the final
	// file absent (rename blocked).
	// Wait briefly for the goroutine to reach the barrier.
	tmp := filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.md.tmp")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(tmp); err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := os.Stat(barrier); err != nil {
		t.Fatalf("barrier file not created (goroutine did not reach barrier)")
	}
	// Final file must NOT exist yet (rename blocked on barrier).
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.md")); err == nil {
		t.Errorf("final file created before barrier removed — atomic-write broken")
	}

	// Remove barrier → rename proceeds, command completes.
	if err := os.Remove(barrier); err != nil {
		t.Fatalf("remove barrier: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("runNavigatorEnrich error after barrier: %v", err)
	}
	// Now the final file exists.
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.md")); err != nil {
		t.Errorf("final file missing after barrier removal: %v", err)
	}
}

func enrichWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func enrichReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}
