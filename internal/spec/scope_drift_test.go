package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTasksMD(t *testing.T) {
	tests := []struct {
		name           string
		tasksMDContent string
		wantTaskCount  int
		wantErr        bool
	}{
		{
			name: "Valid tasks.md with multiple tasks",
			tasksMDContent: `# Task Decomposition

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | Create main handler | REQ-001 | - | internal/handler.go | pending |
| T-002 | Add tests | REQ-002 | T-001 | internal/handler_test.go | pending |
`,
			wantTaskCount: 2,
			wantErr:      false,
		},
		{
			name: "Tasks.md with no planned files",
			tasksMDContent: `# Task Decomposition

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | Review code | REQ-001 | - | - | pending |
`,
			wantTaskCount: 1,
			wantErr:      false,
		},
		{
			name:           "Empty tasks.md",
			tasksMDContent: `# Task Decomposition\n\nNo tasks yet.\n`,
			wantTaskCount:  0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			tasksPath := filepath.Join(tmpDir, "tasks.md")
			
			// Write test content
			if err := os.WriteFile(tasksPath, []byte(tt.tasksMDContent), 0644); err != nil {
				t.Fatalf("failed to write tasks.md: %v", err)
			}

			// Parse tasks
			tasks, err := ParseTasksMD(tmpDir)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTasksMD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(tasks) != tt.wantTaskCount {
				t.Errorf("ParseTasksMD() got %d tasks, want %d", len(tasks), tt.wantTaskCount)
			}
		})
	}
}

func TestCalculateScopeDrift(t *testing.T) {
	tests := []struct {
		name             string
		plannedFiles     []string
		actualFiles      []string
		wantDriftPct     float64
		wantUnplannedMin int // Minimum unplanned files (may vary due to exclusions)
	}{
		{
			name:         "No drift - perfect match",
			plannedFiles: []string{"file1.go", "file2.go"},
			actualFiles:  []string{"file1.go", "file2.go"},
			wantDriftPct: 0.0,
		},
		{
			name:         "Minor drift - 10% (1 unplanned out of 10 planned)",
			plannedFiles: []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go"},
			actualFiles:  []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go", "bonus.go"},
			wantDriftPct: 10.0,
		},
		{
			name:         "Significant drift - 40% (4 unplanned out of 10 planned)",
			plannedFiles: []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go"},
			actualFiles:  []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "extra1.go", "extra2.go", "extra3.go", "extra4.go", "file7.go", "file8.go", "file9.go", "file10.go"},
			wantDriftPct: 40.0,
		},
		{
			name:         "Test files excluded from drift calculation",
			plannedFiles: []string{"handler.go"},
			actualFiles:  []string{"handler.go", "handler_test.go", "testdata/mock.json"},
			wantDriftPct: 0.0, // Test files are excluded
		},
		{
			name:         "Markdown files excluded",
			plannedFiles: []string{"code.go"},
			actualFiles:  []string{"code.go", "README.md", "docs/api.md"},
			wantDriftPct: 0.0, // Markdown files excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driftPct, unplanned := CalculateScopeDrift(tt.plannedFiles, tt.actualFiles)
			
			if driftPct != tt.wantDriftPct {
				t.Errorf("CalculateScopeDrift() drift = %v%%, want %v%%", driftPct, tt.wantDriftPct)
			}
			
			if tt.wantUnplannedMin > 0 && len(unplanned) < tt.wantUnplannedMin {
				t.Errorf("CalculateScopeDrift() unplanned count = %d, want >= %d", len(unplanned), tt.wantUnplannedMin)
			}
		})
	}
}

func TestShouldExcludeFile(t *testing.T) {
	tests := []struct {
		filePath string
		want     bool
	}{
		{"internal/handler.go", false},
		{"internal/handler_test.go", true},
		{"testdata/mock.json", true},
		{".gitignore", true},
		{"README.md", true},
		{"docs/api.md", true},
		{"generated/code.go", true},
		{"src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			if got := ShouldExcludeFile(tt.filePath); got != tt.want {
				t.Errorf("ShouldExcludeFile(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestDetectScopeDrift(t *testing.T) {
	t.Run("Complete drift detection workflow", func(t *testing.T) {
		// Create temporary SPEC directory
		tmpDir := t.TempDir()
		
		// Create tasks.md with planned files
		tasksContent := `# Task Decomposition — TEST-SPEC

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | Create handler | REQ-001 | - | internal/handler.go | pending |
| T-002 | Create service | REQ-002 | T-001 | internal/service.go | pending |
`
		tasksPath := filepath.Join(tmpDir, "tasks.md")
		if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
			t.Fatalf("failed to write tasks.md: %v", err)
		}

		// Test scenario 1: No drift
		actualFiles := []string{"internal/handler.go", "internal/service.go"}
		report, err := DetectScopeDrift(tmpDir, actualFiles)
		if err != nil {
			t.Fatalf("DetectScopeDrift() error = %v", err)
		}

		if report.DriftPercentage != 0.0 {
			t.Errorf("Expected 0%% drift, got %v%%", report.DriftPercentage)
		}
		if report.DriftLevel != "informational" {
			t.Errorf("Expected 'informational' level, got '%s'", report.DriftLevel)
		}
		if report.TriggerReplanning {
			t.Error("Expected no re-planning trigger for 0%% drift")
		}

		// Test scenario 2: Significant drift (40%)
		actualFilesWithDrift := []string{
			"internal/handler.go", 
			"internal/service.go",
			"internal/utils.go",  // Unplanned
			"internal/config.go", // Unplanned
		}
		report, err = DetectScopeDrift(tmpDir, actualFilesWithDrift)
		if err != nil {
			t.Fatalf("DetectScopeDrift() error = %v", err)
		}

		if report.DriftPercentage != 100.0 { // 2 planned, 2 unplanned = 100% drift
			t.Errorf("Expected 100%% drift, got %v%%", report.DriftPercentage)
		}
		if report.DriftLevel != "critical" {
			t.Errorf("Expected 'critical' level, got '%s'", report.DriftLevel)
		}
		if !report.TriggerReplanning {
			t.Error("Expected re-planning trigger for >30%% drift")
		}
	})
}

func TestParseFileList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
	}{
		{
			name:    "Single file",
			input:   "file1.go",
			wantLen: 1,
		},
		{
			name:    "Multiple files comma-separated",
			input:   "file1.go, file2.go, file3.go",
			wantLen: 3,
		},
		{
			name:    "Empty/dash placeholder",
			input:   "-",
			wantLen: 0,
		},
		{
			name:    "Empty string",
			input:   "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFileList(tt.input)
			if len(result) != tt.wantLen {
				t.Errorf("parseFileList(%q) = %v, want length %d", tt.input, result, tt.wantLen)
			}
		})
	}
}
