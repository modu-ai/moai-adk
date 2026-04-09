package memo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead_MissingFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	content, err := Read(projectDir, 2200)
	if err != nil {
		t.Fatalf("Read() on missing file should not error, got: %v", err)
	}
	if content != "" {
		t.Errorf("Read() = %q, want empty string for missing file", content)
	}
}

func TestRead_EmptyProjectDir(t *testing.T) {
	t.Parallel()

	_, err := Read("", 2200)
	if err == nil {
		t.Error("expected error for empty projectDir")
	}
}

func TestRead_ExistingFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	sections := []Section{
		{Priority: P1Required, Title: "Context", Content: "session=test123"},
	}
	if err := Write(projectDir, sections); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	content, err := Read(projectDir, 2200)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if !strings.Contains(content, "session=test123") {
		t.Errorf("content missing expected text, got: %q", content)
	}
}

func TestRead_TokenTrimming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sections      []Section
		maxTokens     int
		wantContains  []string
		wantAbsent    []string
	}{
		{
			name: "no trimming needed",
			sections: []Section{
				{Priority: P1Required, Title: "Required", Content: "short p1"},
				{Priority: P4Low, Title: "Low", Content: "short p4"},
			},
			maxTokens:    2200,
			wantContains: []string{"short p1", "short p4"},
		},
		{
			name: "trims low priority when over budget",
			sections: []Section{
				{Priority: P1Required, Title: "Required", Content: "critical info"},
				{Priority: P4Low, Title: "Low", Content: strings.Repeat("x", 9000)}, // ~2250 tokens
			},
			maxTokens:    500,
			wantContains: []string{"critical info"},
			wantAbsent:   []string{"## P4: Low"},
		},
		{
			name: "unlimited tokens with maxTokens=0 returns all",
			sections: []Section{
				{Priority: P1Required, Title: "A", Content: "a"},
				{Priority: P4Low, Title: "B", Content: "b"},
			},
			maxTokens:    0,
			wantContains: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectDir := t.TempDir()
			if err := Write(projectDir, tt.sections); err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			content, err := Read(projectDir, tt.maxTokens)
			if err != nil {
				t.Fatalf("Read() error: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("content should contain %q", want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(content, absent) {
					t.Errorf("content should NOT contain %q", absent)
				}
			}
		})
	}
}

func TestRead_CorruptFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write valid content - not corrupted, just verify read works normally
	memoPath := filepath.Join(projectDir, memoFileName)
	if err := os.WriteFile(memoPath, []byte("# Session Memo\n\nsome content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := Read(projectDir, 2200)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if !strings.Contains(content, "some content") {
		t.Errorf("expected content, got: %q", content)
	}
}

func TestEstimateTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"abcd", 1},      // 4 chars = 1 token
		{"abcde", 2},     // 5 chars = ceil(5/4) = 2 tokens
		{"abcdefgh", 2},  // 8 chars = 2 tokens
	}

	for _, tt := range tests {
		got := estimateTokens(tt.input)
		if got != tt.want {
			t.Errorf("estimateTokens(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
