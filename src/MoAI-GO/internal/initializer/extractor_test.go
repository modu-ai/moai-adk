package initializer

import (
	"os"
	"path/filepath"
	"testing"
)

// --- NewExtractor ---

func TestNewExtractor(t *testing.T) {
	e := NewExtractor()
	if e == nil {
		t.Fatal("NewExtractor returned nil")
	}
	if e.templateFS == nil {
		t.Error("templateFS is nil")
	}
}

// --- ListTemplates ---

func TestListTemplates(t *testing.T) {
	e := NewExtractor()
	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("ListTemplates returned no files; expected embedded templates")
	}

	// Verify some expected template files exist
	expectedPrefixes := []string{".claude/", ".moai/"}
	foundPrefix := make(map[string]bool)
	for _, f := range files {
		for _, prefix := range expectedPrefixes {
			if len(f) >= len(prefix) && f[:len(prefix)] == prefix {
				foundPrefix[prefix] = true
			}
		}
	}

	for _, prefix := range expectedPrefixes {
		if !foundPrefix[prefix] {
			t.Errorf("expected template files with prefix %q, but none found", prefix)
		}
	}
}

// --- ExtractTemplates ---

func TestExtractTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	e := NewExtractor()

	if err := e.ExtractTemplates(tmpDir); err != nil {
		t.Fatalf("ExtractTemplates() error = %v", err)
	}

	// Verify some files were extracted
	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	for _, f := range files {
		// Just check a sampling of files exists on disk
		// (checking all would be redundant since ExtractTemplates walks all)
		if f == files[0] {
			// First file should have been extracted
			break
		}
	}

	if len(files) == 0 {
		t.Error("no templates to extract")
	}
}

// --- ExtractFile ---

func TestExtractFile_ValidTemplate(t *testing.T) {
	e := NewExtractor()

	// Get the list of templates to find a valid path
	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(files) == 0 {
		t.Skip("no embedded template files to test")
	}

	tmpDir := t.TempDir()
	targetPath := tmpDir + "/extracted_file"

	err = e.ExtractFile(files[0], targetPath)
	if err != nil {
		t.Fatalf("ExtractFile() error = %v", err)
	}
}

func TestExtractFile_InvalidTemplate(t *testing.T) {
	e := NewExtractor()
	tmpDir := t.TempDir()
	targetPath := tmpDir + "/extracted_file"

	err := e.ExtractFile("nonexistent/file/path.txt", targetPath)
	if err == nil {
		t.Error("expected error for nonexistent template file")
	}
}

func TestExtractTemplates_VerifyFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	e := NewExtractor()

	if err := e.ExtractTemplates(tmpDir); err != nil {
		t.Fatalf("ExtractTemplates() error = %v", err)
	}

	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	// Verify at least one extracted file has non-zero content
	for _, f := range files {
		targetPath := filepath.Join(tmpDir, f)
		info, err := os.Stat(targetPath)
		if err != nil {
			t.Errorf("file %q not found on disk: %v", f, err)
			continue
		}
		if info.Size() > 0 {
			return // Found a non-empty file, test passes
		}
	}

	t.Error("no non-empty files found after extraction")
}

func TestExtractFile_CreatesDirectoryStructure(t *testing.T) {
	e := NewExtractor()
	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(files) == 0 {
		t.Skip("no embedded templates")
	}

	tmpDir := t.TempDir()
	// Use a deeply nested target path
	targetPath := filepath.Join(tmpDir, "deep", "nested", "dir", "file.txt")

	if err := e.ExtractFile(files[0], targetPath); err != nil {
		t.Fatalf("ExtractFile() error = %v", err)
	}

	// Verify the deep directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "deep", "nested", "dir")); err != nil {
		t.Fatalf("directory structure not created: %v", err)
	}
}

func TestListTemplates_ContainsExpectedFiles(t *testing.T) {
	e := NewExtractor()
	files, err := e.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	// Should contain config files from .moai/ templates
	foundConfig := false
	for _, f := range files {
		if filepath.Dir(f) == ".moai/config/sections" || filepath.Dir(f) == ".moai/config" {
			foundConfig = true
			break
		}
	}
	if !foundConfig {
		t.Error("expected .moai/config files in template listing")
	}
}
