package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileChangedHandler_EventType(t *testing.T) {
	h := NewFileChangedHandler()
	if h.EventType() != EventFileChanged {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventFileChanged)
	}
}

func TestFileChangedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *HookInput
		createFile    bool
		fileContent   string
		expectMessage bool
	}{
		{
			name: "deleted file - skip",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "/tmp/deleted.go",
				ChangeType:    "deleted",
				HookEventName: "FileChanged",
			},
			createFile:    false,
			expectMessage: false,
		},
		{
			name: "unsupported extension",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "/tmp/file.txt",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:    false,
			expectMessage: false,
		},
		{
			name: "supported Go file without tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.go",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:    true,
			fileContent:   "package main\n\nfunc main() {}\n",
			expectMessage: false,
		},
		{
			name: "supported Go file with tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.go",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:    true,
			fileContent:   "// @MX:NOTE: This is a note\npackage main\n",
			expectMessage: true,
		},
		{
			name: "Python file with tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.py",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:    true,
			fileContent:   "# @MX:NOTE: Python note\nprint('hello')\n",
			expectMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewFileChangedHandler()

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.FilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filepath.Base(filePath))
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.FilePath = filePath
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			if tt.expectMessage && out.SystemMessage == "" {
				t.Error("expected SystemMessage for MX tag delta")
			}
			if !tt.expectMessage && out.SystemMessage != "" {
				t.Errorf("unexpected SystemMessage: %v", out.SystemMessage)
			}
		})
	}
}

func TestFileChangedHandler_SupportedExtensions(t *testing.T) {
	tests := []struct {
		ext      string
		supported bool
	}{
		{".go", true},
		{".py", true},
		{".ts", true},
		{".js", true},
		{".rs", true},
		{".java", true},
		{".kt", true},
		{".cs", true},
		{".rb", true},
		{".php", true},
		{".ex", true},
		{".exs", true},
		{".cpp", true},
		{".cc", true},
		{".cxx", true},
		{".h", true},
		{".hpp", true},
		{".scala", true},
		{".r", true},
		{".dart", true},
		{".swift", true},
		{".txt", false},
		{".md", false},
		{".json", false},
		{".yaml", false},
		{".yml", false},
		{".toml", false},
		{".xml", false},
		{".html", false},
		{".css", false},
		{".sh", false},
		{".bash", false},
		{".zsh", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			if supportedExtensions[tt.ext] != tt.supported {
				t.Errorf("Extension %v: supported=%v, want %v", tt.ext, supportedExtensions[tt.ext], tt.supported)
			}
		})
	}
}
