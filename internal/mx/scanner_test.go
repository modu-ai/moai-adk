package mx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestScanFile tests the basic tag extraction from a Go file.
// AC-SPC-002-01: Given a Go source file with @MX:NOTE, produce correct Tag.
func TestScanFile(t *testing.T) {
	// Create temp file with @MX tag
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// @MX:NOTE: explains why handler forks
func handler() {
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	tags, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(tags))
	}

	tag := tags[0]
	if tag.Kind != MXNote {
		t.Errorf("Expected kind MXNote, got %v", tag.Kind)
	}
	if tag.Body != "explains why handler forks" {
		t.Errorf("Expected body 'explains why handler forks', got %q", tag.Body)
	}
	if tag.Line != 3 {
		t.Errorf("Expected line 3, got %d", tag.Line)
	}
	if tag.File != testFile {
		t.Errorf("Expected file %q, got %q", testFile, tag.File)
	}
}

// TestScanAll16Languages tests tag extraction from all 16 supported languages.
// AC-SPC-002-15: All 16 languages produce tags.
func TestScanAll16Languages(t *testing.T) {
	tmpDir := t.TempDir()

	// Map of extension to sample comment with @MX tag
	testCases := map[string]string{
		".go":     "// @MX:NOTE: Go tag",
		".py":     "# @MX:NOTE: Python tag",
		".ts":     "// @MX:NOTE: TypeScript tag",
		".js":     "// @MX:NOTE: JavaScript tag",
		".rs":     "// @MX:NOTE: Rust tag",
		".java":   "// @MX:NOTE: Java tag",
		".kt":     "// @MX:NOTE: Kotlin tag",
		".cs":     "// @MX:NOTE: C# tag",
		".rb":     "# @MX:NOTE: Ruby tag",
		".php":    "// @MX:NOTE: PHP tag",
		".ex":     "# @MX:NOTE: Elixir tag",
		".exs":    "# @MX:NOTE: Elixir script tag",
		".cpp":    "// @MX:NOTE: C++ tag",
		".scala":  "// @MX:NOTE: Scala tag",
		".R":      "# @MX:NOTE: R tag",
		".dart":   "// @MX:NOTE: Dart tag",
		".swift":  "// @MX:NOTE: Swift tag",
	}

	scanner := NewScanner()

	for ext, comment := range testCases {
		t.Run(ext, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test"+ext)
			content := comment + "\n"
			if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}

			tags, err := scanner.ScanFile(testFile)
			if err != nil {
				t.Fatalf("ScanFile failed for %s: %v", ext, err)
			}

			if len(tags) != 1 {
				t.Fatalf("Expected 1 tag for %s, got %d", ext, len(tags))
			}

			tag := tags[0]
			if tag.Kind != MXNote {
				t.Errorf("Expected kind MXNote for %s, got %v", ext, tag.Kind)
			}
		})
	}
}

// TestScanFileWithWarnReason tests WARN tag with REASON sub-line.
// AC-SPC-002-05: WARN without REASON emits MissingReasonForWarn warning.
func TestScanFileWithWarnReason(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectWarn   bool
		expectReason bool
	}{
		{
			name: "WARN with REASON on next line",
			content: `// @MX:WARN: missing timeout
// @MX:REASON: HTTP requests can hang indefinitely`,
			expectWarn:   false,
			expectReason: true,
		},
		{
			name: "WARN without REASON",
			content: `// @MX:WARN: missing timeout`,
			expectWarn:   true,
			expectReason: false,
		},
		{
			name: "WARN with REASON after 3 lines",
			content: `// @MX:WARN: missing timeout
// Line 2
// Line 3
// Line 4
// @MX:REASON: Too late`,
			expectWarn:   true,
			expectReason: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			scanner := NewScanner()
			tags, err := scanner.ScanFile(testFile)
			if err != nil {
				t.Fatalf("ScanFile failed: %v", err)
			}

			if tt.expectWarn {
				warnings := scanner.GetWarnings()
				if len(warnings) == 0 {
					t.Error("Expected MissingReasonForWarn warning")
				}
				found := false
				for _, w := range warnings {
					if strings.Contains(w, "MissingReasonForWarn") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected MissingReasonForWarn warning, got: %v", warnings)
				}
			}

			if tt.expectReason && len(tags) > 0 {
				if tags[0].Reason == "" {
					t.Error("Expected WARN tag to have Reason")
				}
			}
		})
	}
}

// TestDuplicateAnchorID tests duplicate AnchorID detection.
// AC-SPC-002-06: Duplicate AnchorID emits error.
func TestDuplicateAnchorID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with same AnchorID
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	content := `// @MX:ANCHOR:auth-handler-v1: Authentication handler
func authHandler() {}
`

	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	_, _ = scanner.ScanFile(file1)
	_, _ = scanner.ScanFile(file2)

	errors := scanner.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected DuplicateAnchorID error")
	}

	found := false
	for _, e := range errors {
		if strings.Contains(e, "DuplicateAnchorID") {
			found = true
			if !strings.Contains(e, "auth-handler-v1") {
				t.Errorf("Error should mention anchor ID 'auth-handler-v1', got: %s", e)
			}
			break
		}
	}
	if !found {
		t.Errorf("Expected DuplicateAnchorID error, got: %v", errors)
	}
}

// TestScanDir tests directory scanning.
func TestScanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	files := map[string]string{
		"file1.go": "// @MX:NOTE: first file\n",
		"file2.go": "// @MX:NOTE: second file\n",
		"file3.py": "# @MX:NOTE: python file\n",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := NewScanner()
	tags, err := scanner.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(tags))
	}
}

// TestIgnorePatterns tests .gitignore-style pattern filtering.
// AC-SPC-002-10: ignore patterns work.
func TestIgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files that should be ignored
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.Mkdir(vendorDir, 0755); err != nil {
		t.Fatal(err)
	}

	vendorFile := filepath.Join(vendorDir, "lib.go")
	if err := os.WriteFile(vendorFile, []byte("// @MX:NOTE: vendor code\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create file that should be scanned
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte("// @MX:NOTE: main code\n"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	scanner.SetIgnorePatterns([]string{"vendor"})

	tags, err := scanner.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag (vendor ignored), got %d", len(tags))
	}

	if tags[0].File != mainFile {
		t.Errorf("Expected tag from main.go, got from %s", tags[0].File)
	}
}

// TestDetectChanges tests delta detection between old and new tags.
func TestDetectChanges(t *testing.T) {
	oldTags := []Tag{
		{Kind: MXNote, File: "file1.go", Line: 1, Body: "old note"},
		{Kind: MXWarn, File: "file1.go", Line: 2, Body: "old warn"},
		{Kind: MXAnchor, File: "file2.go", Line: 1, Body: "anchor", AnchorID: "test"},
	}

	newTags := []Tag{
		{Kind: MXNote, File: "file1.go", Line: 1, Body: "old note"}, // unchanged
		{Kind: MXWarn, File: "file1.go", Line: 2, Body: "updated warn"}, // changed
		{Kind: MXTodo, File: "file3.go", Line: 1, Body: "new todo"}, // added
		// old ANCHOR removed
	}

	added, changed, removed := DetectChanges(oldTags, newTags)

	if len(added) != 1 {
		t.Errorf("Expected 1 added tag, got %d", len(added))
	}
	if len(changed) != 1 {
		t.Errorf("Expected 1 changed tag, got %d", len(changed))
	}
	if len(removed) != 1 {
		t.Errorf("Expected 1 removed tag, got %d", len(removed))
	}
}

// TestParseTag tests tag parsing from content string.
func TestParseTag(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectKind  TagKind
		expectBody  string
		expectID    string
		expectError bool
	}{
		{
			name:        "NOTE tag",
			content:     "NOTE: simple note",
			expectKind:  MXNote,
			expectBody:  "simple note",
			expectError: false,
		},
		{
			name:        "WARN tag",
			content:     "WARN: dangerous code",
			expectKind:  MXWarn,
			expectBody:  "dangerous code",
			expectError: false,
		},
		{
			name:        "ANCHOR tag with ID",
			content:     "ANCHOR:my-id: function description",
			expectKind:  MXAnchor,
			expectBody:  "function description",
			expectID:    "my-id",
			expectError: false,
		},
		{
			name:        "ANCHOR tag without ID",
			content:     "ANCHOR: description",
			expectKind:  MXAnchor,
			expectBody:  "description",
			expectID:    "",
			expectError: false,
		},
		{
			name:        "TODO tag",
			content:     "TODO: needs implementation",
			expectKind:  MXTodo,
			expectBody:  "needs implementation",
			expectError: false,
		},
		{
			name:        "LEGACY tag",
			content:     "LEGACY: old code",
			expectKind:  MXLegacy,
			expectBody:  "old code",
			expectError: false,
		},
		{
			name:        "invalid kind",
			content:     "INVALID: some text",
			expectError: true,
		},
		{
			name:        "malformed tag",
			content:     "NOTEwithoutcolon",
			expectError: true,
		},
	}

	scanner := NewScanner()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := scanner.parseTag("/test/file.go", 1, tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tag.Kind != tt.expectKind {
				t.Errorf("Expected kind %v, got %v", tt.expectKind, tag.Kind)
			}
			if tag.Body != tt.expectBody {
				t.Errorf("Expected body %q, got %q", tt.expectBody, tag.Body)
			}
			if tt.expectID != "" && tag.AnchorID != tt.expectID {
				t.Errorf("Expected AnchorID %q, got %q", tt.expectID, tag.AnchorID)
			}
		})
	}
}

// TestExtractReason tests REASON extraction from tag lines.
func TestExtractReason(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{
			line:     "// @MX:REASON: This is the reason",
			expected: "This is the reason",
		},
		{
			line:     "# @MX:REASON: Python reason",
			expected: "Python reason",
		},
		{
			line:     "// @MX:REASON:reason with spaces",
			expected: "reason with spaces",
		},
		{
			line:     "// @MX:NOTE something else",
			expected: "",
		},
		{
			line:     "// regular comment",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := extractReason(tt.line)
			if got != tt.expected {
				t.Errorf("extractReason(%q) = %q, want %q", tt.line, got, tt.expected)
			}
		})
	}
}

// TestExtractTagContent tests extraction of @MX tag content from comment lines.
func TestExtractTagContent(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		prefix       string
		expectOk     bool
		expectContent string
	}{
		{
			name:         "valid Go comment",
			line:         "// @MX:NOTE some text",
			prefix:       "//",
			expectOk:     true,
			expectContent: "NOTE some text",
		},
		{
			name:         "valid Python comment",
			line:         "# @MX:WARN: danger",
			prefix:       "#",
			expectOk:     true,
			expectContent: "WARN: danger",
		},
		{
			name:         "comment without @MX",
			line:         "// regular comment",
			prefix:       "//",
			expectOk:     false,
		},
		{
			name:         "comment before @MX prefix",
			line:         "// text before @MX:NOTE: tag",
			prefix:       "//",
			expectOk:     true,
			expectContent: "NOTE: tag",
		},
		{
			name:         "wrong comment prefix",
			line:         "# @MX:NOTE tag",
			prefix:       "//",
			expectOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, ok := extractTagContent(tt.line, tt.prefix)
			if ok != tt.expectOk {
				t.Errorf("extractTagContent ok = %v, want %v", ok, tt.expectOk)
			}
			if ok && content != tt.expectContent {
				t.Errorf("extractTagContent content = %q, want %q", content, tt.expectContent)
			}
		})
	}
}

// TestScanFileTimestamps tests that scanned tags have current timestamps.
func TestScanFileTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := "// @MX:NOTE: timestamp test\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	before := time.Now()
	scanner := NewScanner()
	tags, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now()

	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(tags))
	}

	tagTime := tags[0].LastSeenAt
	if tagTime.Before(before) || tagTime.After(after) {
		t.Errorf("Tag timestamp %v is outside scan window [%v, %v]", tagTime, before, after)
	}
}
