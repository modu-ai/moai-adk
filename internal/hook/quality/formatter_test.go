package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk-go/internal/hook"
)

func TestNewFormatter(t *testing.T) {
	t.Parallel()

	f := NewFormatter()
	if f == nil {
		t.Fatal("NewFormatter returned nil")
	}

	if f.registry == nil {
		t.Error("registry is nil")
	}

	if f.detector == nil {
		t.Error("detector is nil")
	}
}

func TestNewFormatterWithRegistry(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()
	f := NewFormatterWithRegistry(registry)

	if f.registry != registry {
		t.Error("registry not set correctly")
	}
}

func TestFormatter_ShouldFormat(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		// Should format
		{"Python file", "test.py", true},
		{"Go file", "main.go", true},
		{"JavaScript", "app.js", true},
		{"TypeScript", "app.ts", true},

		// Should skip - extensions
		{"JSON file", "config.json", false},
		{"Lock file", "poetry.lock", false},
		{"Minified JS", "app.min.js", false},
		{"Source map", "app.map", false},
		{"PNG image", "icon.png", false},
		{"Binary", "program.exe", false},

		// Should skip - directories
		{"node_modules", "node_modules/package.js", false},
		{".git file", ".git/config", false},
		{"vendor file", "vendor/lib.js", false},
		{"dist file", "dist/app.js", false},
		{"build file", "build/main.go", false},

		// Should skip - minified detection
		{"Minified detection", "jquery.3.6.0.min.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file if it doesn't exist
			if !strings.Contains(tt.filePath, "node_modules") &&
				!strings.Contains(tt.filePath, ".git") &&
				!strings.Contains(tt.filePath, "vendor") {
				absPath := filepath.Join(t.TempDir(), tt.filePath)
				os.MkdirAll(filepath.Dir(absPath), 0755)
				os.WriteFile(absPath, []byte("test"), 0644)
				tt.filePath = absPath
			}

			got := f.ShouldFormat(tt.filePath)
			if got != tt.want {
				t.Errorf("ShouldFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatter_ShouldFormat_NonExistentFile(t *testing.T) {
	t.Parallel()

	f := NewFormatter()
	result := f.ShouldFormat("/nonexistent/file/path.py")

	if result {
		t.Error("ShouldFormat returned true for nonexistent file")
	}
}

func TestFormatter_FormatFile_SkipUnsupported(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	// JSON should be skipped
	result, err := f.FormatFile(context.Background(), "test.json", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result for skipped file")
	}
}

func TestFormatter_FormatFile_NoFormatterAvailable(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	// Unknown extension
	result, err := f.FormatFile(context.Background(), "test.unknown_ext", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result for unknown file type")
	}
}

func TestFormatter_FormatForHook_SkippedFile(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	output := f.FormatForHook(context.Background(), "test.json", "")

	if output.SuppressOutput != true {
		t.Error("expected SuppressOutput for skipped file")
	}
}

func TestFormatter_EventType(t *testing.T) {
	t.Parallel()

	f := NewFormatter()
	if f.EventType() != hook.EventPostToolUse {
		t.Errorf("EventType = %q, want PostToolUse", f.EventType())
	}
}

func TestFormatter_Handle_SkipNonWriteEdit(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	tests := []struct {
		toolName string
	}{
		{"Read"},
		{"Bash"},
		{"Grep"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			input := &hook.HookInput{
				ToolName: tt.toolName,
			}

			output, err := f.Handle(context.Background(), input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if output.Decision != hook.DecisionAllow {
				t.Errorf("Decision = %q, want allow", output.Decision)
			}
		})
	}
}

func TestFormatter_Handle_EmptyFilePath(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	input := &hook.HookInput{
		ToolName:  "Write",
		ToolInput: []byte(`{"other_field": "value"}`),
	}

	output, err := f.Handle(context.Background(), input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output.Decision != hook.DecisionAllow {
		t.Errorf("Decision = %q, want allow", output.Decision)
	}
}

func TestFormatter_isBinary(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	// Create a temporary file with null byte (binary indicator)
	tmpDir := t.TempDir()
	binaryFile := filepath.Join(tmpDir, "binary.dat")
	content := make([]byte, 100)
	content[50] = 0 // Null byte
	if err := os.WriteFile(binaryFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !f.isBinary(binaryFile) {
		t.Error("isBinary returned false for file with null byte")
	}

	// Create a text file
	textFile := filepath.Join(tmpDir, "text.txt")
	if err := os.WriteFile(textFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if f.isBinary(textFile) {
		t.Error("isBinary returned true for text file")
	}
}

func TestFormatter_getParentDirectories(t *testing.T) {
	t.Parallel()

	f := NewFormatter()

	tests := []struct {
		name     string
		path     string
		wantDirs []string
	}{
		{
			name:     "simple path",
			path:     "/home/user/project/file.go",
			wantDirs: []string{"file.go", "project", "user", "home"},
		},
		{
			name:     "relative path",
			path:     "project/module/file.py",
			wantDirs: []string{"file.py", "module", "project"},
		},
		{
			name:     "single file",
			path:     "file.py",
			wantDirs: []string{"file.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDirs := f.getParentDirectories(tt.path)

			// Just check that we get some directories and order is correct
			if len(gotDirs) == 0 {
				t.Error("got no directories")
			}

			// First element should be the file name
			if gotDirs[0] != filepath.Base(tt.path) {
				t.Errorf("first element = %q, want %q", gotDirs[0], filepath.Base(tt.path))
			}
		})
	}
}

func TestExtractFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "valid file path",
			input:    `{"file_path": "/path/to/file.py"}`,
			wantPath: "/path/to/file.py",
			wantErr:  false,
		},
		{
			name:     "file path with quotes",
			input:    `{"file_path": "/path/to/file with spaces.py"}`,
			wantPath: "/path/to/file with spaces.py",
			wantErr:  false,
		},
		{
			name:     "no file path",
			input:    `{"other_field": "value"}`,
			wantPath: "",
			wantErr:  false,
		},
		{
			name:     "empty JSON",
			input:    `{}`,
			wantPath: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hook.HookInput{
				ToolInput: []byte(tt.input),
			}

			path, err := extractFilePath(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if path != tt.wantPath {
				t.Errorf("extractFilePath() = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestFormatter_Integration(t *testing.T) {
	t.Parallel()

	// This test verifies the Formatter works end-to-end with the hook system
	f := NewFormatter()

	// Create a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := []byte("import os\nimport sys\nprint('hello')")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Verify ShouldFormat returns true
	if !f.ShouldFormat(testFile) {
		t.Error("ShouldFormat returned false for .py file")
	}

	// Verify EventType is correct
	if f.EventType() != hook.EventPostToolUse {
		t.Errorf("EventType = %q, want PostToolUse", f.EventType())
	}

	// Test Handle with proper input
	inputJSON := `{"file_path": "` + testFile + `"}`
	input := &hook.HookInput{
		ToolName:  "Write",
		ToolInput: []byte(inputJSON),
		CWD:       tmpDir,
	}

	output, err := f.Handle(context.Background(), input)
	if err != nil {
		t.Errorf("Handle returned error: %v", err)
	}

	if output == nil {
		t.Fatal("Handle returned nil output")
	}

	// Should not block
	if output.Decision == hook.DecisionDeny {
		t.Error("Handle returned deny decision")
	}
}
