package status

import (
	"os"
	"path/filepath"
	"testing"
)

// --- SpecInfo struct ---

func TestSpecInfoStruct(t *testing.T) {
	spec := SpecInfo{
		ID:     "SPEC-001",
		Title:  "Test Specification",
		Status: "In Progress",
	}

	if spec.ID != "SPEC-001" {
		t.Errorf("ID = %q, want %q", spec.ID, "SPEC-001")
	}
	if spec.Title != "Test Specification" {
		t.Errorf("Title = %q", spec.Title)
	}
	if spec.Status != "In Progress" {
		t.Errorf("Status = %q", spec.Status)
	}
}

// --- GetActiveSpecs ---

func TestGetActiveSpecs_NoSpecsDir(t *testing.T) {
	tmpDir := t.TempDir()

	specs, err := GetActiveSpecs(tmpDir)
	if err != nil {
		t.Fatalf("GetActiveSpecs() error = %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("expected 0 specs, got %d", len(specs))
	}
}

func TestGetActiveSpecs_EmptySpecsDir(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".moai", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	specs, err := GetActiveSpecs(tmpDir)
	if err != nil {
		t.Fatalf("GetActiveSpecs() error = %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("expected 0 specs, got %d", len(specs))
	}
}

func TestGetActiveSpecs_WithSpecs(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".moai", "specs")

	// Create SPEC-001
	spec001Dir := filepath.Join(specsDir, "SPEC-001")
	if err := os.MkdirAll(spec001Dir, 0755); err != nil {
		t.Fatal(err)
	}
	spec001Content := `# SPEC-001
Add user authentication

| **Status** | In Progress |
`
	if err := os.WriteFile(filepath.Join(spec001Dir, "spec.md"), []byte(spec001Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create SPEC-002
	spec002Dir := filepath.Join(specsDir, "SPEC-002")
	if err := os.MkdirAll(spec002Dir, 0755); err != nil {
		t.Fatal(err)
	}
	spec002Content := `# SPEC-002
Implement API endpoints

| **Status** | Completed |
`
	if err := os.WriteFile(filepath.Join(spec002Dir, "spec.md"), []byte(spec002Content), 0644); err != nil {
		t.Fatal(err)
	}

	specs, err := GetActiveSpecs(tmpDir)
	if err != nil {
		t.Fatalf("GetActiveSpecs() error = %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
}

func TestGetActiveSpecs_SkipsNonDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".moai", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (not a directory) in specs
	if err := os.WriteFile(filepath.Join(specsDir, "README.md"), []byte("specs info"), 0644); err != nil {
		t.Fatal(err)
	}

	specs, err := GetActiveSpecs(tmpDir)
	if err != nil {
		t.Fatalf("GetActiveSpecs() error = %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("expected 0 specs (files should be skipped), got %d", len(specs))
	}
}

func TestGetActiveSpecs_SkipsDirWithoutSpecMd(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".moai", "specs")

	// Create directory without spec.md
	emptySpecDir := filepath.Join(specsDir, "SPEC-EMPTY")
	if err := os.MkdirAll(emptySpecDir, 0755); err != nil {
		t.Fatal(err)
	}

	specs, err := GetActiveSpecs(tmpDir)
	if err != nil {
		t.Fatalf("GetActiveSpecs() error = %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("expected 0 specs (dir without spec.md should be skipped), got %d", len(specs))
	}
}

// --- parseSpecInfo ---

func TestParseSpecInfo(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantTitle  string
		wantStatus string
	}{
		{
			name: "standard format",
			content: `# SPEC-001
Add user authentication

| **Status** | In Progress |
`,
			wantTitle:  "Add user authentication",
			wantStatus: "In Progress",
		},
		{
			name: "completed status",
			content: `# SPEC-002
Implement API endpoints

| **Status** | Completed |
`,
			wantTitle:  "Implement API endpoints",
			wantStatus: "Completed",
		},
		{
			name:       "no title match",
			content:    "Some random content without heading",
			wantTitle:  "Unknown",
			wantStatus: "Unknown",
		},
		{
			name: "title but no status",
			content: `# SPEC-003
Build dashboard

No status table here.
`,
			wantTitle:  "Build dashboard",
			wantStatus: "Unknown",
		},
		{
			name:       "empty content",
			content:    "",
			wantTitle:  "Unknown",
			wantStatus: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, status := parseSpecInfo(tt.content)
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

// --- FormatStatusIcon ---

func TestFormatStatusIcon(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"completed", "\u2713"},
		{"Completed", "\u2713"},
		{"done", "\u2713"},
		{"Done", "\u2713"},
		{"approved", "\u2713"},
		{"Approved", "\u2713"},
		{"in progress", "\u25d0"},
		{"In Progress", "\u25d0"},
		{"active", "\u25d0"},
		{"Active", "\u25d0"},
		{"pending", "\u25d0"},
		{"Pending", "\u25d0"},
		{"draft", "\u25cb"},
		{"Draft", "\u25cb"},
		{"planned", "\u25cb"},
		{"Planned", "\u25cb"},
		{"unknown", "?"},
		{"", "?"},
		{"custom status", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := FormatStatusIcon(tt.status)
			if got != tt.want {
				t.Errorf("FormatStatusIcon(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
