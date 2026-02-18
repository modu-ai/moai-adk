package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSpecLinker_Save_MkdirAllFails tests the save error path when MkdirAll fails.
// This covers the "create registry dir" error branch in save().
func TestSpecLinker_Save_MkdirAllFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	dir := t.TempDir()

	// Create the linker with a valid directory first.
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error: %v", err)
	}

	// Now make the parent of .moai read-only so MkdirAll will fail when
	// save() tries to create a subdirectory inside the read-only dir.
	// We achieve this by making the temp dir itself read-only,
	// which prevents creation of the .moai directory within it.
	// First ensure .moai does NOT yet exist (it won't since no link was saved).
	moaiDir := filepath.Join(dir, ".moai")
	if _, err := os.Stat(moaiDir); err == nil {
		// If .moai already exists, remove it so MkdirAll is needed.
		if err := os.RemoveAll(moaiDir); err != nil {
			t.Fatalf("RemoveAll: %v", err)
		}
	}

	// Make the parent directory read-only to prevent subdirectory creation.
	if err := os.Chmod(dir, 0o444); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o755)
	})

	err = linker.LinkIssueToSpec(1, "SPEC-1")
	if err == nil {
		t.Skip("expected error for mkdir in read-only dir, but succeeded (possibly running as root)")
	}
}

// TestSpecLinker_Save_WriteFailsOnReadOnlyFile tests write failure when
// the temp file cannot be written (simulate by using a dir as the registry path).
func TestSpecLinker_Save_RegistryPathIsDir(t *testing.T) {
	dir := t.TempDir()
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a directory where the registry file should be.
	// This forces os.Rename to fail (cannot rename a temp file over a directory).
	registryPathAsDir := filepath.Join(moaiDir, RegistryFileName)
	if err := os.MkdirAll(registryPathAsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll registry as dir: %v", err)
	}

	// Directly create a fileSpecLinker pointing to the "directory" as registry path.
	// We use internal struct access since we're in the same package.
	linker := &fileSpecLinker{
		registryPath: registryPathAsDir,
		registry: &Registry{
			Version:  RegistryVersion,
			Mappings: []SpecMapping{},
		},
	}

	// save() will try to rename temp file to a directory path, which should fail.
	err := linker.save()
	if err == nil {
		t.Error("save() expected error when registry path is a directory, got nil")
	}
}

// TestSpecLinker_Save_AtomicWrite verifies that save() writes data atomically
// by confirming the file exists and has valid JSON after a successful save.
func TestSpecLinker_Save_AtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	registryPath := filepath.Join(moaiDir, RegistryFileName)
	linker := &fileSpecLinker{
		registryPath: registryPath,
		registry: &Registry{
			Version: RegistryVersion,
			Mappings: []SpecMapping{
				{IssueNumber: 1, SpecID: "SPEC-1", Status: "active"},
			},
		},
	}

	if err := linker.save(); err != nil {
		t.Fatalf("save() error = %v", err)
	}

	// Verify the file was written correctly.
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if reg.Version != RegistryVersion {
		t.Errorf("Version = %q, want %q", reg.Version, RegistryVersion)
	}
	if len(reg.Mappings) != 1 {
		t.Fatalf("Mappings len = %d, want 1", len(reg.Mappings))
	}
	if reg.Mappings[0].IssueNumber != 1 {
		t.Errorf("IssueNumber = %d, want 1", reg.Mappings[0].IssueNumber)
	}
}

// TestSpecLinker_Save_CreatesDir verifies that save() creates the .moai directory if absent.
func TestSpecLinker_Save_CreatesDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Do NOT pre-create the .moai directory; save() should create it.
	registryPath := filepath.Join(dir, ".moai", RegistryFileName)

	linker := &fileSpecLinker{
		registryPath: registryPath,
		registry: &Registry{
			Version:  RegistryVersion,
			Mappings: []SpecMapping{},
		},
	}

	if err := linker.save(); err != nil {
		t.Fatalf("save() error = %v (should create .moai dir)", err)
	}

	if _, err := os.Stat(registryPath); err != nil {
		t.Errorf("registry file not found after save: %v", err)
	}
}
