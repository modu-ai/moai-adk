package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDesignFolderSkip_NonEmpty verifies that scaffoldDesignDir skips deployment
// when .moai/design/ already contains at least one regular file (AC-10, REQ-010).
func TestDesignFolderSkip_NonEmpty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Place one regular file to make dir non-empty
	if err := os.WriteFile(filepath.Join(designDir, "stale.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale.txt: %v", err)
	}

	var warnBuf strings.Builder
	deployed, err := scaffoldDesignDir(root, &warnBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deployed {
		t.Error("scaffoldDesignDir should return deployed=false for non-empty dir")
	}

	// Warning must contain ".moai/design/" and "skip"
	warn := warnBuf.String()
	if !strings.Contains(warn, ".moai/design/") {
		t.Errorf("warning %q should contain '.moai/design/'", warn)
	}
	if !strings.Contains(warn, "skip") {
		t.Errorf("warning %q should contain 'skip'", warn)
	}

	// stale.txt must still exist
	if _, err := os.Stat(filepath.Join(designDir, "stale.txt")); os.IsNotExist(err) {
		t.Error("stale.txt must remain after skip")
	}
	// README.md must NOT be created
	if _, err := os.Stat(filepath.Join(designDir, "README.md")); !os.IsNotExist(err) {
		t.Error("README.md must not be created when skipping")
	}
}

// TestDesignFolderSkip_HiddenOnly verifies that directories containing only hidden
// files (e.g. .DS_Store) are treated as empty and deployment proceeds (AC-10).
func TestDesignFolderSkip_HiddenOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Place only a hidden file (.DS_Store)
	if err := os.WriteFile(filepath.Join(designDir, ".DS_Store"), []byte("mac"), 0o644); err != nil {
		t.Fatalf("write .DS_Store: %v", err)
	}

	var warnBuf strings.Builder
	deployed, err := scaffoldDesignDir(root, &warnBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deployed {
		t.Error("scaffoldDesignDir should return deployed=true when dir contains only hidden files")
	}
	// README.md must be created
	if _, err := os.Stat(filepath.Join(designDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md must be created when dir contains only hidden files")
	}
}

// TestDesignFolderSkip_EmptyDir verifies that an empty .moai/design/ triggers
// template deployment normally (AC-10).
func TestDesignFolderSkip_EmptyDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var warnBuf strings.Builder
	deployed, err := scaffoldDesignDir(root, &warnBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deployed {
		t.Error("scaffoldDesignDir should return deployed=true for empty dir")
	}
	if _, err := os.Stat(filepath.Join(designDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md must be created for empty dir")
	}
}

// TestDesignFolderUserEditPreserved verifies that updateDesignDir does NOT overwrite
// a file whose content differs from the canonical template hash (AC-8, REQ-005).
func TestDesignFolderUserEditPreserved(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// First, scaffold normally (deploy templates)
	var warnBuf strings.Builder
	deployed, err := scaffoldDesignDir(root, &warnBuf)
	if err != nil || !deployed {
		t.Fatalf("initial scaffold failed: err=%v deployed=%v", err, deployed)
	}

	// Simulate user editing README.md
	readmePath := filepath.Join(designDir, "README.md")
	origContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	userEdit := append(origContent, []byte("\nUSER EDIT\n")...)
	if err := os.WriteFile(readmePath, userEdit, 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}

	// Now run updateDesignDir (simulates moai update)
	var errBuf strings.Builder
	if err := updateDesignDir(root, &errBuf); err != nil {
		t.Fatalf("updateDesignDir: %v", err)
	}

	// User edit must be preserved
	after, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README.md after update: %v", err)
	}
	if !strings.Contains(string(after), "USER EDIT") {
		t.Error("user edit must be preserved after updateDesignDir")
	}
}

// TestDesignFolderUpdate_ReservedExact_WarnsButContinues verifies that updateDesignDir
// emits a warning (not an error) when a reserved exact-name file exists, preserves
// the file, and continues syncing other templates (AC-DFF-01, REQ-DFF-001/002/004).
func TestDesignFolderUpdate_ReservedExact_WarnsButContinues(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Scaffold templates first
	var warnBuf strings.Builder
	if _, err := scaffoldDesignDir(root, &warnBuf); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	// User creates a file with reserved exact name (simulates pre-v2.15 generated artifact)
	tokensContent := []byte(`{"primary": "#ff0000"}`)
	if err := os.WriteFile(filepath.Join(designDir, "tokens.json"), tokensContent, 0o644); err != nil {
		t.Fatalf("write tokens.json: %v", err)
	}

	// User modifies README.md to have a custom edit marker
	readmePath := filepath.Join(designDir, "README.md")
	origReadme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	userEdit := append(origReadme, []byte("\nUSER EDIT MARKER\n")...)
	if err := os.WriteFile(readmePath, userEdit, 0o644); err != nil {
		t.Fatalf("write user README edit: %v", err)
	}

	var errBuf strings.Builder
	err = updateDesignDir(root, &errBuf)

	// AC-DFF-01: function must return nil (not error)
	if err != nil {
		t.Fatalf("updateDesignDir must return nil for reserved filename in update path, got: %v", err)
	}

	msg := errBuf.String()
	// AC-DFF-01: warning must contain "warning" keyword
	if !strings.Contains(strings.ToLower(msg), "warning") {
		t.Errorf("errOut must contain 'warning', got: %q", msg)
	}
	// AC-DFF-01: warning must mention the filename
	if !strings.Contains(msg, "tokens.json") {
		t.Errorf("errOut must mention 'tokens.json', got: %q", msg)
	}
	// AC-DFF-01: warning must include preservation guidance
	if !strings.Contains(strings.ToLower(msg), "preserved") {
		t.Errorf("errOut must contain 'preserved', got: %q", msg)
	}

	// AC-DFF-01: tokens.json content must remain unchanged (user data preservation)
	gotTokens, readErr := os.ReadFile(filepath.Join(designDir, "tokens.json"))
	if readErr != nil {
		t.Fatalf("tokens.json must still exist: %v", readErr)
	}
	if !bytes.Equal(gotTokens, tokensContent) {
		t.Errorf("tokens.json content must not be modified, got %q want %q", gotTokens, tokensContent)
	}

	// AC-DFF-01: user edit in README.md must be preserved (REQ-005 user-edit preservation)
	afterReadme, readErr := os.ReadFile(readmePath)
	if readErr != nil {
		t.Fatalf("README.md must still exist: %v", readErr)
	}
	if !strings.Contains(string(afterReadme), "USER EDIT MARKER") {
		t.Error("README.md user edit must be preserved after updateDesignDir")
	}
}

// TestDesignFolderUpdate_ReservedGlob_WarnsButContinues verifies that updateDesignDir
// emits a warning for brief/BRIEF-*.md glob matches, preserves the file, and continues
// syncing other templates (AC-DFF-02, REQ-DFF-001/002).
func TestDesignFolderUpdate_ReservedGlob_WarnsButContinues(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Scaffold templates first
	var warnBuf strings.Builder
	if _, err := scaffoldDesignDir(root, &warnBuf); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	// User creates a file matching the brief/BRIEF-*.md glob pattern
	briefDir := filepath.Join(designDir, "brief")
	if err := os.MkdirAll(briefDir, 0o755); err != nil {
		t.Fatalf("mkdir brief: %v", err)
	}
	briefContent := []byte("local design brief notes")
	if err := os.WriteFile(filepath.Join(briefDir, "BRIEF-LOCAL.md"), briefContent, 0o644); err != nil {
		t.Fatalf("write BRIEF-LOCAL.md: %v", err)
	}

	var errBuf strings.Builder
	err := updateDesignDir(root, &errBuf)

	// AC-DFF-02: function must return nil
	if err != nil {
		t.Fatalf("updateDesignDir must return nil for reserved glob in update path, got: %v", err)
	}

	msg := errBuf.String()
	// AC-DFF-02: warning keyword present
	if !strings.Contains(strings.ToLower(msg), "warning") {
		t.Errorf("errOut must contain 'warning', got: %q", msg)
	}
	// AC-DFF-02: mentions the file (BRIEF appears in path)
	if !strings.Contains(msg, "BRIEF") {
		t.Errorf("errOut must mention 'BRIEF', got: %q", msg)
	}
	// AC-DFF-02: preservation guidance
	if !strings.Contains(strings.ToLower(msg), "preserved") {
		t.Errorf("errOut must contain 'preserved', got: %q", msg)
	}

	// AC-DFF-02: BRIEF-LOCAL.md content must remain unchanged
	gotBrief, readErr := os.ReadFile(filepath.Join(briefDir, "BRIEF-LOCAL.md"))
	if readErr != nil {
		t.Fatalf("BRIEF-LOCAL.md must still exist: %v", readErr)
	}
	if !bytes.Equal(gotBrief, briefContent) {
		t.Errorf("BRIEF-LOCAL.md content must not be modified, got %q want %q", gotBrief, briefContent)
	}

	// AC-DFF-02: other templates (research.md, system.md, spec.md) must still be accessible
	for _, name := range []string{"research.md", "system.md", "spec.md"} {
		if _, statErr := os.Stat(filepath.Join(designDir, name)); os.IsNotExist(statErr) {
			t.Errorf("template file %s must still be present after reserved glob warning", name)
		}
	}
}

// TestDesignFolderScaffold_ReservedExact_StillErrors verifies that checkReservedCollision
// in strict mode (strict=true) still returns an error for reserved exact-name files.
// This simulates the scaffold path behavior (AC-DFF-03, REQ-DFF-003/008).
func TestDesignFolderScaffold_ReservedExact_StillErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a reserved exact-name file
	tokensContent := []byte(`{"x":1}`)
	if err := os.WriteFile(filepath.Join(designDir, "tokens.json"), tokensContent, 0o644); err != nil {
		t.Fatalf("write tokens.json: %v", err)
	}

	var errBuf strings.Builder
	// AC-DFF-03: strict=true must return error (scaffold path behavior)
	err := checkReservedCollision(root, &errBuf, true)
	if err == nil {
		t.Fatal("checkReservedCollision(strict=true) must return error for reserved filename")
	}
	// Error message must identify the issue
	if !strings.Contains(err.Error(), "reserved filename") {
		t.Errorf("error must mention 'reserved filename', got: %v", err)
	}

	// AC-DFF-03: tokens.json must remain unchanged (user data preservation always honored)
	gotTokens, readErr := os.ReadFile(filepath.Join(designDir, "tokens.json"))
	if readErr != nil {
		t.Fatalf("tokens.json must still exist: %v", readErr)
	}
	if !bytes.Equal(gotTokens, tokensContent) {
		t.Errorf("tokens.json must not be modified, got %q want %q", gotTokens, tokensContent)
	}
}

// TestDesignFolderUpdate_WarningIncludesGuidance verifies that the warning message
// contains all required guidance keywords for the user to understand the issue
// (AC-DFF-04, REQ-DFF-005/007).
func TestDesignFolderUpdate_WarningIncludesGuidance(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a reserved exact-name file (components.json)
	if err := os.WriteFile(filepath.Join(designDir, "components.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write components.json: %v", err)
	}

	var errBuf strings.Builder
	err := updateDesignDir(root, &errBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := errBuf.String()
	// AC-DFF-04: all required keywords must be present
	requiredKeywords := []string{"warning", "components.json", "preserved", "rename"}
	for _, kw := range requiredKeywords {
		if !strings.Contains(strings.ToLower(msg), strings.ToLower(kw)) {
			t.Errorf("warning must contain %q, got: %s", kw, msg)
		}
	}
	// AC-DFF-04: message must not be empty (REQ-DFF-007 no silent failure)
	if strings.TrimSpace(msg) == "" {
		t.Error("warning message must not be empty (silent skip is prohibited)")
	}
}

// TestDesignFolderUpdate_ReservedNotModified verifies that reserved files are
// never modified by updateDesignDir — byte-identical preservation required
// (AC-DFF-05, REQ-DFF-006/004).
func TestDesignFolderUpdate_ReservedNotModified(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create two reserved files with known content
	components := []byte(`{"user": "data"}`)
	imports := []byte(`{"warning": "test"}`)
	if err := os.WriteFile(filepath.Join(designDir, "components.json"), components, 0o644); err != nil {
		t.Fatalf("write components.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(designDir, "import-warnings.json"), imports, 0o644); err != nil {
		t.Fatalf("write import-warnings.json: %v", err)
	}

	var errBuf strings.Builder
	// AC-DFF-05: function must return nil
	if err := updateDesignDir(root, &errBuf); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// AC-DFF-05: both files must be byte-identical (no modification)
	got1, err := os.ReadFile(filepath.Join(designDir, "components.json"))
	if err != nil {
		t.Fatalf("read components.json: %v", err)
	}
	if !bytes.Equal(got1, components) {
		t.Error("components.json was modified — must remain byte-identical")
	}

	got2, err := os.ReadFile(filepath.Join(designDir, "import-warnings.json"))
	if err != nil {
		t.Fatalf("read import-warnings.json: %v", err)
	}
	if !bytes.Equal(got2, imports) {
		t.Error("import-warnings.json was modified — must remain byte-identical")
	}
}

// TestDesignFolderUpdate_MultipleReservedConflicts verifies that updateDesignDir
// emits warnings for all reserved files simultaneously, preserves all files, and
// continues syncing non-reserved templates (AC-DFF-06, REQ-DFF-001/002).
func TestDesignFolderUpdate_MultipleReservedConflicts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	briefDir := filepath.Join(designDir, "brief")
	if err := os.MkdirAll(briefDir, 0o755); err != nil {
		t.Fatalf("mkdir brief: %v", err)
	}

	// Create three reserved files simultaneously
	reservedFiles := map[string][]byte{
		filepath.Join(designDir, "tokens.json"):     []byte(`{"a":1}`),
		filepath.Join(briefDir, "BRIEF-X.md"):       []byte("brief X"),
		filepath.Join(designDir, "components.json"): []byte(`[]`),
	}
	for path, content := range reservedFiles {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	var errBuf strings.Builder
	// AC-DFF-06: function must return nil despite multiple conflicts
	if err := updateDesignDir(root, &errBuf); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	msg := errBuf.String()
	// AC-DFF-06: all three filenames must appear in warnings
	for _, kw := range []string{"tokens.json", "BRIEF-X", "components.json"} {
		if !strings.Contains(msg, kw) {
			t.Errorf("warning must mention %q, got: %s", kw, msg)
		}
	}

	// AC-DFF-06: all three files must be byte-identical (preserved)
	for path, want := range reservedFiles {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s was modified — must remain byte-identical", path)
		}
	}
}

// TestDesignFolderSubdirs verifies that scaffoldDesignDir creates wireframes/ and screenshots/
// subdirectories (AC-4, REQ-003).
func TestDesignFolderSubdirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	var warnBuf strings.Builder
	deployed, err := scaffoldDesignDir(root, &warnBuf)
	if err != nil || !deployed {
		t.Fatalf("scaffold failed: err=%v deployed=%v", err, deployed)
	}

	for _, subdir := range []string{"wireframes", "screenshots"} {
		dirPath := filepath.Join(root, ".moai", "design", subdir)
		info, statErr := os.Stat(dirPath)
		if statErr != nil {
			t.Errorf("subdir %q must exist: %v", subdir, statErr)
			continue
		}
		if !info.IsDir() {
			t.Errorf("subdir %q must be a directory", subdir)
		}
	}
}
