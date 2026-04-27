package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMD(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write md: %v", err)
	}
	return p
}

func TestInjectMarker_FreshFile(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "# Project\n")
	if err := InjectMarker(path, "SPEC-PROJ-INIT-001", "ios-mobile",
		[]string{".moai/config/sections/workflow.yaml", ".moai/harness/main.md"}); err != nil {
		t.Fatalf("inject: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "## Project-Specific Configuration (Harness-Generated)") {
		t.Error("heading missing")
	}
	if !strings.Contains(content, `<!-- moai:harness-start id="SPEC-PROJ-INIT-001"`) {
		t.Error("start marker missing")
	}
	if !strings.Contains(content, "<!-- moai:harness-end -->") {
		t.Error("end marker missing")
	}
	if !strings.Contains(content, "@.moai/harness/main.md") {
		t.Error("@import path missing")
	}
}

func TestInjectMarker_DifferentContent_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "# Project\n")
	if err := InjectMarker(path, "SPEC-PROJ-INIT-001", "ios-mobile",
		[]string{".moai/harness/main.md"}); err != nil {
		t.Fatal(err)
	}
	if err := InjectMarker(path, "SPEC-PROJ-INIT-002", "android-mobile",
		[]string{".moai/harness/main.md"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)

	headingCount := strings.Count(content, "## Project-Specific Configuration (Harness-Generated)")
	if headingCount != 1 {
		t.Errorf("heading appears %d times, want 1\n--- file ---\n%s", headingCount, content)
	}
	startCount := strings.Count(content, "<!-- moai:harness-start")
	if startCount != 1 {
		t.Errorf("start marker appears %d times, want 1", startCount)
	}
	endCount := strings.Count(content, "<!-- moai:harness-end -->")
	if endCount != 1 {
		t.Errorf("end marker appears %d times, want 1", endCount)
	}
	if !strings.Contains(content, "android-mobile") {
		t.Error("second-run domain not reflected")
	}
	if strings.Contains(content, "ios-mobile") {
		t.Error("first-run domain still present after replace")
	}
}

func TestInjectMarker_SameSpecID_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "# Project\n")
	for i := 0; i < 3; i++ {
		if err := InjectMarker(path, "SPEC-X", "d", nil); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	data, _ := os.ReadFile(path)
	if c := strings.Count(string(data), "## Project-Specific Configuration"); c != 1 {
		t.Errorf("heading count = %d, want 1", c)
	}
}

func TestInjectMarker_AppendToFileWithoutTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "# Project (no trailing newline)")
	if err := InjectMarker(path, "S", "d", nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "## Project-Specific Configuration") {
		t.Error("heading missing")
	}
}

func TestInjectMarker_EmptyPath(t *testing.T) {
	if err := InjectMarker("", "S", "d", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestInjectMarker_EmptySpecID(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "x")
	if err := InjectMarker(path, "", "d", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestInjectMarker_FileNotFound(t *testing.T) {
	if err := InjectMarker(filepath.Join(t.TempDir(), "nope.md"), "S", "d", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestInjectMarker_NoEnsureAllowedCall(t *testing.T) {
	// Smoke test: writing to t.TempDir() (absolute path outside any allowed
	// prefix) MUST succeed because layer3 does not invoke EnsureAllowed.
	dir := t.TempDir()
	path := writeMD(t, dir, "CLAUDE.md", "x\n")
	if err := InjectMarker(path, "S", "d", nil); err != nil {
		t.Fatalf("layer3 must not call EnsureAllowed; got: %v", err)
	}
}
