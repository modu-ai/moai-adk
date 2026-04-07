package memo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sections    []Section
		wantContain []string
		wantErr     bool
	}{
		{
			name: "single P1 section",
			sections: []Section{
				{Priority: P1Required, Title: "Context", Content: "session=abc", Budget: 200},
			},
			wantContain: []string{
				"# Session Memo",
				"## P1: Context",
				"session=abc",
			},
		},
		{
			name: "multiple sections ordered by priority",
			sections: []Section{
				{Priority: P4Low, Title: "History", Content: "decision1", Budget: 500},
				{Priority: P1Required, Title: "SPEC", Content: "SPEC-001 active", Budget: 200},
				{Priority: P2High, Title: "Tasks", Content: "3 pending tasks", Budget: 500},
			},
			wantContain: []string{
				"## P1: SPEC",
				"SPEC-001 active",
				"## P2: Tasks",
				"3 pending tasks",
				"## P4: History",
				"decision1",
			},
		},
		{
			name: "empty content sections are skipped",
			sections: []Section{
				{Priority: P1Required, Title: "Context", Content: "has content"},
				{Priority: P2High, Title: "Empty", Content: ""},
			},
			wantContain: []string{"## P1: Context"},
		},
		{
			name:    "empty projectDir returns error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var projectDir string
			if !tt.wantErr {
				projectDir = t.TempDir()
			}

			err := Write(projectDir, tt.sections)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Write() unexpected error: %v", err)
			}

			data, err := os.ReadFile(filepath.Join(projectDir, memoFileName))
			if err != nil {
				t.Fatalf("read memo file: %v", err)
			}

			content := string(data)
			for _, want := range tt.wantContain {
				if !strings.Contains(content, want) {
					t.Errorf("memo content missing %q\nfull content:\n%s", want, content)
				}
			}
		})
	}
}

func TestWrite_CreatesParentDirectories(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// Ensure the .moai/state directory does NOT pre-exist.
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if _, err := os.Stat(stateDir); err == nil {
		t.Fatal("state dir should not exist yet")
	}

	sections := []Section{
		{Priority: P1Required, Title: "Test", Content: "hello"},
	}

	if err := Write(projectDir, sections); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	memoPath := filepath.Join(projectDir, memoFileName)
	if _, err := os.Stat(memoPath); err != nil {
		t.Errorf("memo file not created: %v", err)
	}
}

func TestWrite_OverwritesExistingFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	first := []Section{{Priority: P1Required, Title: "First", Content: "original"}}
	if err := Write(projectDir, first); err != nil {
		t.Fatalf("first Write() error: %v", err)
	}

	second := []Section{{Priority: P1Required, Title: "Second", Content: "overwritten"}}
	if err := Write(projectDir, second); err != nil {
		t.Fatalf("second Write() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, memoFileName))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "original") {
		t.Error("old content should have been overwritten")
	}
	if !strings.Contains(content, "overwritten") {
		t.Error("new content should be present")
	}
}

func TestWrite_SectionsOrderedByPriority(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	sections := []Section{
		{Priority: P4Low, Title: "Fourth", Content: "p4"},
		{Priority: P2High, Title: "Second", Content: "p2"},
		{Priority: P3Medium, Title: "Third", Content: "p3"},
		{Priority: P1Required, Title: "First", Content: "p1"},
	}

	if err := Write(projectDir, sections); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(projectDir, memoFileName))
	content := string(data)

	posP1 := strings.Index(content, "## P1:")
	posP2 := strings.Index(content, "## P2:")
	posP3 := strings.Index(content, "## P3:")
	posP4 := strings.Index(content, "## P4:")

	if posP1 >= posP2 || posP2 >= posP3 || posP3 >= posP4 {
		t.Errorf("sections out of order: P1=%d P2=%d P3=%d P4=%d", posP1, posP2, posP3, posP4)
	}
}
