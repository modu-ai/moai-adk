package github

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupLinker(t *testing.T) (SpecLinker, string) {
	t.Helper()
	dir := t.TempDir()
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error: %v", err)
	}
	return linker, dir
}

func TestNewSpecLinker_CreatesEmptyRegistry(t *testing.T) {
	linker, _ := setupLinker(t)

	mappings := linker.ListMappings()
	if len(mappings) != 0 {
		t.Errorf("ListMappings() len = %d, want 0", len(mappings))
	}
}

func TestLinkIssueToSpec_Success(t *testing.T) {
	linker, dir := setupLinker(t)

	if err := linker.LinkIssueToSpec(123, "SPEC-ISSUE-123"); err != nil {
		t.Fatalf("LinkIssueToSpec() error: %v", err)
	}

	// Verify mapping exists.
	specID, err := linker.GetLinkedSpec(123)
	if err != nil {
		t.Fatalf("GetLinkedSpec() error: %v", err)
	}
	if specID != "SPEC-ISSUE-123" {
		t.Errorf("GetLinkedSpec(123) = %q, want %q", specID, "SPEC-ISSUE-123")
	}

	// Verify reverse lookup.
	issueNum, err := linker.GetLinkedIssue("SPEC-ISSUE-123")
	if err != nil {
		t.Fatalf("GetLinkedIssue() error: %v", err)
	}
	if issueNum != 123 {
		t.Errorf("GetLinkedIssue(%q) = %d, want 123", "SPEC-ISSUE-123", issueNum)
	}

	// Verify file was written.
	registryPath := filepath.Join(dir, ".moai", RegistryFileName)
	if _, err := os.Stat(registryPath); err != nil {
		t.Errorf("registry file not found: %v", err)
	}
}

func TestLinkIssueToSpec_DuplicateError(t *testing.T) {
	linker, _ := setupLinker(t)

	if err := linker.LinkIssueToSpec(123, "SPEC-ISSUE-123"); err != nil {
		t.Fatalf("first LinkIssueToSpec() error: %v", err)
	}

	err := linker.LinkIssueToSpec(123, "SPEC-ISSUE-123-V2")
	if err == nil {
		t.Fatal("second LinkIssueToSpec() should error on duplicate")
	}
	if !errors.Is(err, ErrMappingExists) {
		t.Errorf("error = %v, want ErrMappingExists", err)
	}
}

func TestGetLinkedSpec_NotFound(t *testing.T) {
	linker, _ := setupLinker(t)

	_, err := linker.GetLinkedSpec(999)
	if err == nil {
		t.Fatal("GetLinkedSpec() should error for unmapped issue")
	}
	if !errors.Is(err, ErrMappingNotFound) {
		t.Errorf("error = %v, want ErrMappingNotFound", err)
	}
}

func TestGetLinkedIssue_NotFound(t *testing.T) {
	linker, _ := setupLinker(t)

	_, err := linker.GetLinkedIssue("SPEC-NONEXISTENT")
	if err == nil {
		t.Fatal("GetLinkedIssue() should error for unmapped SPEC")
	}
	if !errors.Is(err, ErrMappingNotFound) {
		t.Errorf("error = %v, want ErrMappingNotFound", err)
	}
}

func TestLinkIssueToSpec_MultipleMappings(t *testing.T) {
	linker, _ := setupLinker(t)

	if err := linker.LinkIssueToSpec(1, "SPEC-ISSUE-1"); err != nil {
		t.Fatalf("LinkIssueToSpec(1) error: %v", err)
	}
	if err := linker.LinkIssueToSpec(2, "SPEC-ISSUE-2"); err != nil {
		t.Fatalf("LinkIssueToSpec(2) error: %v", err)
	}
	if err := linker.LinkIssueToSpec(3, "SPEC-ISSUE-3"); err != nil {
		t.Fatalf("LinkIssueToSpec(3) error: %v", err)
	}

	mappings := linker.ListMappings()
	if len(mappings) != 3 {
		t.Fatalf("ListMappings() len = %d, want 3", len(mappings))
	}

	// Verify each mapping.
	for i, expected := range []struct {
		issueNum int
		specID   string
	}{
		{1, "SPEC-ISSUE-1"},
		{2, "SPEC-ISSUE-2"},
		{3, "SPEC-ISSUE-3"},
	} {
		if mappings[i].IssueNumber != expected.issueNum {
			t.Errorf("mapping[%d].IssueNumber = %d, want %d", i, mappings[i].IssueNumber, expected.issueNum)
		}
		if mappings[i].SpecID != expected.specID {
			t.Errorf("mapping[%d].SpecID = %q, want %q", i, mappings[i].SpecID, expected.specID)
		}
		if mappings[i].Status != "active" {
			t.Errorf("mapping[%d].Status = %q, want %q", i, mappings[i].Status, "active")
		}
	}
}

func TestSpecLinker_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Create and populate.
	linker1, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker(1) error: %v", err)
	}
	if err := linker1.LinkIssueToSpec(42, "SPEC-ISSUE-42"); err != nil {
		t.Fatalf("LinkIssueToSpec() error: %v", err)
	}

	// Reload from disk.
	linker2, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker(2) error: %v", err)
	}

	specID, err := linker2.GetLinkedSpec(42)
	if err != nil {
		t.Fatalf("GetLinkedSpec() error: %v", err)
	}
	if specID != "SPEC-ISSUE-42" {
		t.Errorf("GetLinkedSpec(42) = %q, want %q", specID, "SPEC-ISSUE-42")
	}
}

func TestSpecLinker_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	// Write corrupt JSON.
	registryPath := filepath.Join(moaiDir, RegistryFileName)
	if err := os.WriteFile(registryPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("write corrupt file error: %v", err)
	}

	// Should recover gracefully.
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error on corrupt file: %v", err)
	}

	// Should start with empty mappings.
	if len(linker.ListMappings()) != 0 {
		t.Errorf("ListMappings() len = %d, want 0 after corrupt recovery", len(linker.ListMappings()))
	}

	// Corrupt file should be backed up.
	corruptPath := registryPath + ".corrupt"
	if _, err := os.Stat(corruptPath); err != nil {
		t.Errorf("corrupt backup not found: %v", err)
	}
}

func TestSpecLinker_RegistryFormat(t *testing.T) {
	dir := t.TempDir()

	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error: %v", err)
	}
	if err := linker.LinkIssueToSpec(100, "SPEC-ISSUE-100"); err != nil {
		t.Fatalf("LinkIssueToSpec() error: %v", err)
	}

	// Read raw file and verify JSON structure.
	registryPath := filepath.Join(dir, ".moai", RegistryFileName)
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if reg.Version != RegistryVersion {
		t.Errorf("Version = %q, want %q", reg.Version, RegistryVersion)
	}
	if len(reg.Mappings) != 1 {
		t.Fatalf("Mappings len = %d, want 1", len(reg.Mappings))
	}
	if reg.Mappings[0].IssueNumber != 100 {
		t.Errorf("IssueNumber = %d, want 100", reg.Mappings[0].IssueNumber)
	}
	if reg.Mappings[0].SpecID != "SPEC-ISSUE-100" {
		t.Errorf("SpecID = %q, want %q", reg.Mappings[0].SpecID, "SPEC-ISSUE-100")
	}

	// Verify file ends with newline (POSIX convention).
	if data[len(data)-1] != '\n' {
		t.Error("registry file should end with newline")
	}
}

func TestListMappings_ReturnsCopy(t *testing.T) {
	linker, _ := setupLinker(t)

	if err := linker.LinkIssueToSpec(1, "SPEC-ISSUE-1"); err != nil {
		t.Fatalf("LinkIssueToSpec() error: %v", err)
	}

	// Modify returned slice.
	mappings := linker.ListMappings()
	mappings[0].Status = "modified"

	// Original should be unchanged.
	original := linker.ListMappings()
	if original[0].Status != "active" {
		t.Errorf("Status = %q, want %q (returned slice should be a copy)", original[0].Status, "active")
	}
}

func TestSpecLinker_Load_NilMappings(t *testing.T) {
	dir := t.TempDir()
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write valid JSON with null mappings
	registryPath := filepath.Join(moaiDir, RegistryFileName)
	reg := `{"version": "1.0.0", "mappings": null}`
	if err := os.WriteFile(registryPath, []byte(reg), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker: %v", err)
	}
	mappings := linker.ListMappings()
	if mappings == nil {
		t.Error("ListMappings should return non-nil slice")
	}
	if len(mappings) != 0 {
		t.Errorf("ListMappings len = %d, want 0", len(mappings))
	}
}

func TestSpecLinker_Load_ReadPermissionError(t *testing.T) {
	dir := t.TempDir()
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	registryPath := filepath.Join(moaiDir, RegistryFileName)
	if err := os.WriteFile(registryPath, []byte("{}"), 0o000); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := NewSpecLinker(dir)
	if err == nil {
		// On some systems, root user can still read 0o000 files.
		// Skip rather than fail.
		t.Skip("expected error for unreadable file, but got nil (possibly running as root)")
	}
}

func TestSpecLinker_Save_ReadOnlyDir(t *testing.T) {
	dir := t.TempDir()

	// Create linker first in a working directory
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker: %v", err)
	}

	// Make the .moai directory read-only
	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(moaiDir, 0o444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions for cleanup
		_ = os.Chmod(moaiDir, 0o755)
	})

	err = linker.LinkIssueToSpec(1, "SPEC-1")
	if err == nil {
		t.Skip("expected error writing to read-only dir, but succeeded (possibly running as root)")
	}
}

func TestSpecLinker_ListMappings_EmptyRegistry(t *testing.T) {
	linker, _ := setupLinker(t)
	mappings := linker.ListMappings()
	if mappings == nil {
		t.Error("ListMappings should return non-nil for empty registry")
	}
	if len(mappings) != 0 {
		t.Errorf("ListMappings len = %d, want 0", len(mappings))
	}
}
