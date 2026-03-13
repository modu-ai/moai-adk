package cli

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMergeUserFiles_PreservesCustomizations tests that mergeUserFiles
// correctly preserves user customizations during template updates.
func TestMergeUserFiles_PreservesCustomizations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	userModified := `{
  "mcpServers": {
    "user-modified": {
      "command": "user-command"
    },
    "user-added": {
      "command": "user-server"
    }
  }
}`
	newTemplate := `{
  "mcpServers": {
    "template-entry": {
      "command": "template-command"
    }
  }
}`

	// Write the "new" template (simulating post-deployment state)
	destPath := filepath.Join(tempDir, ".mcp.json")
	if err := os.WriteFile(destPath, []byte(newTemplate), 0644); err != nil {
		t.Fatalf("write new template: %v", err)
	}

	// Create backups with user's content
	backups := []fileBackup{
		{path: ".mcp.json", data: []byte(userModified)},
	}

	// Run merge
	var out bytes.Buffer
	if err := mergeUserFiles(tempDir, backups, &out); err != nil {
		t.Fatalf("mergeUserFiles: %v", err)
	}

	// Read merged result
	result, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}

	resultStr := string(result)

	// Verify user's modifications are preserved (at minimum, user content should be kept)
	// When merge engine fails or base is not found, it preserves user's version
	if !strings.Contains(resultStr, "user-modified") {
		t.Error("user's modifications should be preserved")
	}
}

// TestMergeUserFiles_NoBaseTemplate tests merge when there's no base template
// (e.g., user-created file).
func TestMergeUserFiles_NoBaseTemplate(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create parent directory
	moaiDir := filepath.Join(tempDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0755); err != nil {
		t.Fatalf("create .moai dir: %v", err)
	}

	userContent := `#!/bin/bash
# User's custom script
export CUSTOM_VAR="user-value"
`

	// Write the template file
	destPath := filepath.Join(moaiDir, "status_line.sh")
	if err := os.WriteFile(destPath, []byte("# Template version\n"), 0644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	backups := []fileBackup{
		{path: ".moai/status_line.sh", data: []byte(userContent)},
	}

	var out bytes.Buffer
	if err := mergeUserFiles(tempDir, backups, &out); err != nil {
		t.Fatalf("mergeUserFiles: %v", err)
	}

	// Read merged result
	result, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}

	resultStr := string(result)

	// User's content should be preserved when no base exists
	if !strings.Contains(resultStr, "CUSTOM_VAR") {
		t.Error("user's custom variable should be preserved when no base exists")
	}
	if !strings.Contains(resultStr, "user-value") {
		t.Error("user's value should be preserved when no base exists")
	}
}

// TestMergeUserFiles_FileRemovedInNewTemplate tests that user's file is
// preserved when it's removed from the new template.
func TestMergeUserFiles_FileRemovedInNewTemplate(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	userContent := `{"user": "content"}`

	// Don't create the destination file (simulating removal from template)

	backups := []fileBackup{
		{path: ".removed.json", data: []byte(userContent)},
	}

	var out bytes.Buffer
	if err := mergeUserFiles(tempDir, backups, &out); err != nil {
		t.Fatalf("mergeUserFiles: %v", err)
	}

	// File should be restored
	destPath := filepath.Join(tempDir, ".removed.json")
	result, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}

	if string(result) != userContent {
		t.Errorf("user's content should be restored, got %s", string(result))
	}
}

// TestMergeUserFiles_EmptyBackups is a table-driven test for edge cases.
func TestMergeUserFiles_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string) ([]fileBackup, string)
		verify  func(t *testing.T, dir string, content string, output string)
		wantErr bool
	}{
		{
			name: "empty backup list",
			setup: func(t *testing.T, dir string) ([]fileBackup, string) {
				return []fileBackup{}, ""
			},
			verify: func(t *testing.T, dir string, content string, output string) {
				if output != "" {
					t.Error("should have no output for empty backups")
				}
			},
		},
		{
			name: "identical content",
			setup: func(t *testing.T, dir string) ([]fileBackup, string) {
				content := `{"same": "content"}`
				destPath := filepath.Join(dir, ".mcp.json")
				if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return []fileBackup{{path: ".mcp.json", data: []byte(content)}}, destPath
			},
			verify: func(t *testing.T, dir string, content string, output string) {
				result, _ := os.ReadFile(content)
				if string(result) != `{"same": "content"}` {
					t.Error("content should remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			backups, targetPath := tt.setup(t, tempDir)

			var out bytes.Buffer
			err := mergeUserFiles(tempDir, backups, &out)

			if (err != nil) != tt.wantErr {
				t.Errorf("mergeUserFiles() error = %v, wantErr %v", err, tt.wantErr)
			}

			if targetPath != "" {
				content, _ := os.ReadFile(targetPath)
				tt.verify(t, tempDir, targetPath, string(content))
			} else {
				tt.verify(t, tempDir, "", out.String())
			}
		})
	}
}

// mockEmbeddedFS is a minimal mock for testing.
type mockEmbeddedFS struct {
	fs.FS
	files map[string]string
}

func (m *mockEmbeddedFS) Open(name string) (fs.File, error) {
	content, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &mockFile{name: name, content: []byte(content)}, nil
}

type mockFile struct {
	fs.File
	name    string
	content []byte
	offset  int64
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{name: m.name, size: int64(len(m.content))}, nil
}

func (m *mockFile) Read(p []byte) (int, error) {
	if m.offset >= int64(len(m.content)) {
		return 0, nil // EOF
	}
	n := copy(p, m.content[m.offset:])
	m.offset += int64(n)
	return n, nil
}

func (m *mockFile) Close() error { return nil }

type mockFileInfo struct {
	fs.FileInfo
	name string
	size int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
