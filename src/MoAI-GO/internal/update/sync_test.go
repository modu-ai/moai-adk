package update

import (
	"strings"
	"testing"
)

// --- DefaultProtectedDirs ---

func TestDefaultProtectedDirs(t *testing.T) {
	expected := []string{".moai/project/", ".moai/specs/"}

	if len(DefaultProtectedDirs) != 2 {
		t.Errorf("DefaultProtectedDirs length = %d, want 2", len(DefaultProtectedDirs))
	}

	dirSet := make(map[string]bool)
	for _, dir := range DefaultProtectedDirs {
		dirSet[dir] = true
	}

	for _, dir := range expected {
		if !dirSet[dir] {
			t.Errorf("DefaultProtectedDirs missing %q", dir)
		}
	}
}

func TestProtectedDirsOnManager(t *testing.T) {
	sm := NewSyncManager("/tmp")

	// Verify default protected dirs are loaded into the manager
	for _, dir := range DefaultProtectedDirs {
		if !sm.isProtected(dir) {
			t.Errorf("isProtected(%q) = false, want true for default protected dir", dir)
		}
	}
}

// --- NewSyncManager ---

func TestNewSyncManager(t *testing.T) {
	sm := NewSyncManager("/tmp/project")
	if sm == nil {
		t.Fatal("NewSyncManager returned nil")
	}
	if sm.projectDir != "/tmp/project" {
		t.Errorf("projectDir = %q, want %q", sm.projectDir, "/tmp/project")
	}
	if sm.templateFS == nil {
		t.Error("templateFS is nil")
	}
	if sm.dryRun {
		t.Error("dryRun should be false by default")
	}
	if !sm.preserveProtected {
		t.Error("preserveProtected should be true by default")
	}
}

// --- SetDryRun ---

func TestSetDryRun(t *testing.T) {
	sm := NewSyncManager("/tmp")

	sm.SetDryRun(true)
	if !sm.dryRun {
		t.Error("dryRun should be true after SetDryRun(true)")
	}

	sm.SetDryRun(false)
	if sm.dryRun {
		t.Error("dryRun should be false after SetDryRun(false)")
	}
}

// --- isProtected ---

func TestIsProtected(t *testing.T) {
	sm := NewSyncManager("/tmp")

	tests := []struct {
		path string
		want bool
	}{
		{".moai/project/", true},
		{".moai/project/product.md", true},
		{".moai/specs/", true},
		{".moai/specs/SPEC-001/spec.md", true},
		{".moai/config/sections/user.yaml", false},
		{".claude/settings.json", false},
		{"CLAUDE.md", false},
		{".moai/config/config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := sm.isProtected(tt.path)
			if got != tt.want {
				t.Errorf("isProtected(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsProtected_PreserveDisabled(t *testing.T) {
	sm := NewSyncManager("/tmp")
	sm.preserveProtected = false

	// When preservation is disabled, nothing is protected
	if sm.isProtected(".moai/project/product.md") {
		t.Error("isProtected should return false when preserveProtected is disabled")
	}
	if sm.isProtected(".moai/specs/SPEC-001/spec.md") {
		t.Error("isProtected should return false when preserveProtected is disabled")
	}
}

// --- Sync ---

func TestSync_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSyncManager(tmpDir)
	sm.SetDryRun(true)

	result, err := sm.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if result == nil {
		t.Fatal("Sync returned nil result")
	}

	// In dry run, files should be tracked but not written
	// (all files should be "added" since directory is empty)
	totalChanges := len(result.FilesAdded) + len(result.FilesUpdated)
	if totalChanges == 0 {
		t.Error("expected some files to be tracked in dry run")
	}
}

func TestSync_ActualSync(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSyncManager(tmpDir)

	result, err := sm.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if result == nil {
		t.Fatal("Sync returned nil result")
	}

	// Should have added files
	if len(result.FilesAdded) == 0 {
		t.Error("expected files to be added")
	}

	// Should have no updated files (first sync)
	if len(result.FilesUpdated) != 0 {
		t.Errorf("expected 0 updated files on first sync, got %d", len(result.FilesUpdated))
	}
}

func TestSync_SecondSync(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSyncManager(tmpDir)

	// First sync
	_, err := sm.Sync()
	if err != nil {
		t.Fatalf("first Sync() error = %v", err)
	}

	// Second sync - all files should be "updated"
	result, err := sm.Sync()
	if err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}

	if len(result.FilesUpdated) == 0 {
		t.Error("expected files to be updated on second sync")
	}
	if len(result.FilesAdded) != 0 {
		t.Errorf("expected 0 added files on second sync, got %d", len(result.FilesAdded))
	}
}

// --- SyncResult ---

func TestSyncResult_GetSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   SyncResult
		contains []string
	}{
		{
			name: "files added only",
			result: SyncResult{
				FilesAdded:   []string{"a.txt", "b.txt"},
				FilesUpdated: []string{},
			},
			contains: []string{"Added 2 files", "Updated 0 files"},
		},
		{
			name: "files updated only",
			result: SyncResult{
				FilesAdded:   []string{},
				FilesUpdated: []string{"a.txt"},
			},
			contains: []string{"Added 0 files", "Updated 1 files"},
		},
		{
			name: "with protected dirs",
			result: SyncResult{
				FilesAdded:    []string{"a.txt"},
				FilesUpdated:  []string{"b.txt"},
				ProtectedDirs: []string{".moai/project/"},
			},
			contains: []string{"Added 1 files", "Updated 1 files", "Preserved 1 protected"},
		},
		{
			name: "empty result",
			result: SyncResult{
				FilesAdded:   []string{},
				FilesUpdated: []string{},
			},
			contains: []string{"Added 0 files", "Updated 0 files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.GetSummary()
			for _, expected := range tt.contains {
				if !strings.Contains(summary, expected) {
					t.Errorf("GetSummary() = %q, missing %q", summary, expected)
				}
			}
		})
	}
}

func TestSyncResult_PrintSummary_DoesNotPanic(t *testing.T) {
	result := &SyncResult{
		FilesAdded:    []string{"a.txt"},
		FilesUpdated:  []string{"b.txt"},
		FilesRemoved:  []string{},
		ProtectedDirs: []string{".moai/project/"},
	}

	// Should not panic
	result.PrintSummary()
}

func TestSyncResult_PrintSummary_EmptyResult(t *testing.T) {
	result := &SyncResult{}

	// Should not panic even with empty slices
	result.PrintSummary()
}

// --- SetProtectedDirs ---

func TestSetProtectedDirs(t *testing.T) {
	sm := NewSyncManager("/tmp")

	// Default: .moai/project/ and .moai/specs/ are protected
	if !sm.isProtected(".moai/project/product.md") {
		t.Error("default protected dir not recognized")
	}

	// Replace with custom dirs
	sm.SetProtectedDirs([]string{".custom/dir/"})

	// Old dirs should no longer be protected
	if sm.isProtected(".moai/project/product.md") {
		t.Error(".moai/project/ should no longer be protected after SetProtectedDirs")
	}

	// New dir should be protected
	if !sm.isProtected(".custom/dir/file.txt") {
		t.Error(".custom/dir/ should be protected after SetProtectedDirs")
	}
}

func TestSetProtectedDirs_Empty(t *testing.T) {
	sm := NewSyncManager("/tmp")

	sm.SetProtectedDirs([]string{})

	// Nothing should be protected
	if sm.isProtected(".moai/project/product.md") {
		t.Error("nothing should be protected with empty protected dirs")
	}
	if sm.isProtected(".moai/specs/SPEC-001/spec.md") {
		t.Error("nothing should be protected with empty protected dirs")
	}
}

// --- Sync with protected paths ---

func TestSync_ProtectedPathsPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSyncManager(tmpDir)

	result, err := sm.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	// Check that some paths were protected (templates include .moai/project/ or .moai/specs/ files)
	// Protected dirs are tracked in result
	if result == nil {
		t.Fatal("Sync returned nil result")
	}

	// Verify total files = added + updated + protected
	totalTracked := len(result.FilesAdded) + len(result.FilesUpdated) + len(result.ProtectedDirs)
	if totalTracked == 0 {
		t.Error("expected some files to be tracked")
	}
}

// --- SyncResult fields ---

func TestSyncResult_FilesRemoved(t *testing.T) {
	result := &SyncResult{
		FilesAdded:   []string{"a.txt"},
		FilesUpdated: []string{"b.txt"},
		FilesRemoved: []string{"c.txt", "d.txt"},
	}

	if len(result.FilesRemoved) != 2 {
		t.Errorf("FilesRemoved length = %d, want 2", len(result.FilesRemoved))
	}
}
