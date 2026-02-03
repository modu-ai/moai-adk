package merge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatConflictMarkers_SingleConflict(t *testing.T) {
	t.Parallel()

	conflicts := []Conflict{
		{
			StartLine: 2,
			EndLine:   2,
			Base:      "B",
			Current:   "B_user",
			Updated:   "B_template",
		},
	}

	result := FormatConflictMarkers([]byte("A\nB_user\nC"), conflicts)

	content := string(result)
	if !strings.Contains(content, "<<<<<<< current") {
		t.Error("expected conflict marker <<<<<<< current")
	}
	if !strings.Contains(content, "B_user") {
		t.Error("expected current content B_user")
	}
	if !strings.Contains(content, "=======") {
		t.Error("expected separator =======")
	}
	if !strings.Contains(content, "B_template") {
		t.Error("expected updated content B_template")
	}
	if !strings.Contains(content, ">>>>>>> updated") {
		t.Error("expected conflict marker >>>>>>> updated")
	}
}

func TestFormatConflictMarkers_MultipleConflicts(t *testing.T) {
	t.Parallel()

	conflicts := []Conflict{
		{StartLine: 1, EndLine: 1, Base: "A", Current: "A_user", Updated: "A_tpl"},
		{StartLine: 3, EndLine: 3, Base: "C", Current: "C_user", Updated: "C_tpl"},
		{StartLine: 5, EndLine: 5, Base: "E", Current: "E_user", Updated: "E_tpl"},
	}

	result := FormatConflictMarkers([]byte("A_user\nB\nC_user\nD\nE_user"), conflicts)

	content := string(result)
	markerCount := strings.Count(content, "<<<<<<<")
	if markerCount != 3 {
		t.Errorf("expected 3 conflict marker sets, got %d", markerCount)
	}
}

func TestWriteConflictFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	originalPath := filepath.Join(dir, "test.md")

	originalContent := []byte("original content")
	if err := os.WriteFile(originalPath, originalContent, 0o644); err != nil {
		t.Fatalf("write original: %v", err)
	}

	conflicts := []Conflict{
		{StartLine: 1, EndLine: 1, Base: "old", Current: "user_ver", Updated: "tpl_ver"},
	}

	mergedContent := []byte("user_ver")
	conflictPath, err := WriteConflictFile(originalPath, mergedContent, conflicts)
	if err != nil {
		t.Fatalf("WriteConflictFile: %v", err)
	}

	// Verify conflict file was created.
	expectedPath := originalPath + ".conflict"
	if conflictPath != expectedPath {
		t.Errorf("conflict path = %q, want %q", conflictPath, expectedPath)
	}

	// Verify conflict file content has markers.
	data, err := os.ReadFile(conflictPath)
	if err != nil {
		t.Fatalf("read conflict file: %v", err)
	}
	if !strings.Contains(string(data), "<<<<<<<") {
		t.Error("conflict file should contain conflict markers")
	}

	// Verify original file is unchanged.
	origData, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(origData) != "original content" {
		t.Error("original file should not be modified")
	}
}

func TestWriteConflictFile_NoConflicts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	originalPath := filepath.Join(dir, "test.md")

	_, err := WriteConflictFile(originalPath, []byte("content"), nil)
	if err == nil {
		t.Error("expected error when no conflicts provided")
	}
}
