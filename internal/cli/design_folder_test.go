package cli

import (
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

// TestDesignFolderReservedExact verifies that updateDesignDir rejects a user file
// that matches a reserved exact name (e.g. tokens.json) (AC-9, REQ-008).
func TestDesignFolderReservedExact(t *testing.T) {
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

	// User creates a file with reserved exact name
	if err := os.WriteFile(filepath.Join(designDir, "tokens.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write tokens.json: %v", err)
	}

	var errBuf strings.Builder
	err := updateDesignDir(root, &errBuf)
	if err == nil {
		t.Fatal("updateDesignDir should return error for reserved filename")
	}
	if !strings.Contains(err.Error(), "reserved filename") && !strings.Contains(errBuf.String(), "reserved filename") {
		t.Errorf("error/stderr must contain 'reserved filename', got: err=%v stderr=%q", err, errBuf.String())
	}

	// tokens.json must remain unchanged
	content, readErr := os.ReadFile(filepath.Join(designDir, "tokens.json"))
	if readErr != nil {
		t.Fatalf("tokens.json must still exist: %v", readErr)
	}
	if string(content) != "{}" {
		t.Errorf("tokens.json content must not be modified, got %q", content)
	}
}

// TestDesignFolderReservedGlob verifies that updateDesignDir rejects a user file
// matching the brief/BRIEF-*.md glob pattern (AC-9, REQ-008).
func TestDesignFolderReservedGlob(t *testing.T) {
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

	// User creates a file matching brief/BRIEF-*.md glob
	briefDir := filepath.Join(designDir, "brief")
	if err := os.MkdirAll(briefDir, 0o755); err != nil {
		t.Fatalf("mkdir brief: %v", err)
	}
	if err := os.WriteFile(filepath.Join(briefDir, "BRIEF-LOCAL.md"), []byte("local brief"), 0o644); err != nil {
		t.Fatalf("write BRIEF-LOCAL.md: %v", err)
	}

	var errBuf strings.Builder
	err := updateDesignDir(root, &errBuf)
	if err == nil {
		t.Fatal("updateDesignDir should return error for reserved glob filename")
	}
	if !strings.Contains(err.Error(), "reserved filename") && !strings.Contains(errBuf.String(), "reserved filename") {
		t.Errorf("error/stderr must contain 'reserved filename', got: err=%v stderr=%q", err, errBuf.String())
	}
}

// TestDesignFolderReservedNotModified verifies that on reserved filename error,
// the existing user file is not modified (AC-9).
func TestDesignFolderReservedNotModified(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	designDir := filepath.Join(root, ".moai", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Place reserved file with known content
	reservedContent := `{"user": "data"}`
	if err := os.WriteFile(filepath.Join(designDir, "components.json"), []byte(reservedContent), 0o644); err != nil {
		t.Fatalf("write components.json: %v", err)
	}

	var errBuf strings.Builder
	_ = updateDesignDir(root, &errBuf)

	// Content must remain unchanged
	content, err := os.ReadFile(filepath.Join(designDir, "components.json"))
	if err != nil {
		t.Fatalf("read components.json: %v", err)
	}
	if string(content) != reservedContent {
		t.Errorf("components.json was modified, got %q want %q", string(content), reservedContent)
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
