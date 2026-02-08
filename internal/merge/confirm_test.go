package merge

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmModel_Init(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test analysis",
		},
	}

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestConfirmModel_Update_AcceptWithY(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test analysis",
		},
		decision: false,
		done:     false,
	}

	// Simulate pressing 'y'
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	result := updatedModel.(confirmModel)

	if !result.decision {
		t.Error("Expected decision to be true after pressing 'y'")
	}

	if !result.done {
		t.Error("Expected done to be true after pressing 'y'")
	}

	if cmd == nil {
		t.Error("Expected Quit command after decision")
	}
}

func TestConfirmModel_Update_CancelWithN(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test analysis",
		},
		decision: false,
		done:     false,
	}

	// Simulate pressing 'n'
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	result := updatedModel.(confirmModel)

	if result.decision {
		t.Error("Expected decision to be false after pressing 'n'")
	}

	if !result.done {
		t.Error("Expected done to be true after pressing 'n'")
	}

	if cmd == nil {
		t.Error("Expected Quit command after decision")
	}
}

func TestConfirmModel_Update_CancelWithCtrlC(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test analysis",
		},
		decision: false,
		done:     false,
	}

	// Simulate pressing Ctrl+C
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	result := updatedModel.(confirmModel)

	if result.decision {
		t.Error("Expected decision to be false after pressing Ctrl+C")
	}

	if !result.done {
		t.Error("Expected done to be true after pressing Ctrl+C")
	}

	if cmd == nil {
		t.Error("Expected Quit command after Ctrl+C")
	}
}

func TestConfirmModel_View_IncludesTitle(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test summary",
		},
		done: false,
	}

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Merge Analysis Results") {
		t.Error("View should contain title")
	}
}

func TestConfirmModel_View_EmptyWhenDone(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test summary",
		},
		done: true,
	}

	view := m.View()

	if view != "" {
		t.Errorf("View should be empty when done, got: %s", view)
	}
}

func TestConfirmModel_View_IncludesSummary(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test summary message",
		},
		done: false,
	}

	view := m.View()

	if !strings.Contains(view, "Test summary message") {
		t.Error("View should contain summary")
	}
}

func TestConfirmModel_View_ShowsConflictWarning(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			HasConflicts: true,
			Files: []FileAnalysis{
				{Path: "test.md", RiskLevel: "high"},
			},
		},
		done: false,
	}

	view := m.View()

	if !strings.Contains(view, "Warning") {
		t.Error("View should contain conflict warning")
	}
}

func TestMergeAnalysis_Creation(t *testing.T) {
	analysis := MergeAnalysis{
		Files: []FileAnalysis{
			{
				Path:      "test.yaml",
				Changes:   "update",
				Strategy:  YAMLDeep,
				RiskLevel: "medium",
			},
		},
		HasConflicts: true,
		SafeToMerge:  false,
		Summary:      "Test merge",
		RiskLevel:    "high",
	}

	if len(analysis.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(analysis.Files))
	}

	if !analysis.HasConflicts {
		t.Error("Expected HasConflicts to be true")
	}

	if analysis.SafeToMerge {
		t.Error("Expected SafeToMerge to be false")
	}
}

// Security: Input Validation Tests

func TestValidateAnalysis_TooManyFiles(t *testing.T) {
	// Create analysis with too many files (DoS attack prevention)
	files := make([]FileAnalysis, 1001)
	for i := range files {
		files[i] = FileAnalysis{
			Path:      "file.go",
			RiskLevel: "low",
		}
	}

	analysis := MergeAnalysis{
		Files:     files,
		RiskLevel: "low",
	}

	err := validateAnalysis(analysis)
	if err == nil {
		t.Error("Expected error for too many files, got nil")
	}

	if !strings.Contains(err.Error(), "too many files") {
		t.Errorf("Expected 'too many files' error, got: %v", err)
	}
}

func TestValidateAnalysis_InvalidRiskLevel(t *testing.T) {
	tests := []struct {
		name      string
		riskLevel string
	}{
		{"SQL injection attempt", "'; DROP TABLE files; --"},
		{"XSS attempt", "<script>alert('xss')</script>"},
		{"Command injection", "high; rm -rf /"},
		{"Invalid value", "critical"},
		{"Empty with spaces", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := MergeAnalysis{
				Files:     []FileAnalysis{{Path: "test.go", RiskLevel: "low"}},
				RiskLevel: tt.riskLevel,
			}

			err := validateAnalysis(analysis)
			if err == nil {
				t.Errorf("Expected error for risk level '%s', got nil", tt.riskLevel)
			}

			if !strings.Contains(err.Error(), "invalid risk level") {
				t.Errorf("Expected 'invalid risk level' error, got: %v", err)
			}
		})
	}
}

func TestValidateAnalysis_PathTooLong(t *testing.T) {
	// Create a path that exceeds 1024 characters
	longPath := strings.Repeat("a", 1025)

	analysis := MergeAnalysis{
		Files: []FileAnalysis{
			{
				Path:      longPath,
				RiskLevel: "low",
			},
		},
		RiskLevel: "low",
	}

	err := validateAnalysis(analysis)
	if err == nil {
		t.Error("Expected error for path too long, got nil")
	}

	if !strings.Contains(err.Error(), "path too long") {
		t.Errorf("Expected 'path too long' error, got: %v", err)
	}
}

func TestValidateAnalysis_InvalidFileRiskLevel(t *testing.T) {
	analysis := MergeAnalysis{
		Files: []FileAnalysis{
			{
				Path:      "test.go",
				RiskLevel: "invalid_risk",
			},
		},
		RiskLevel: "low",
	}

	err := validateAnalysis(analysis)
	if err == nil {
		t.Error("Expected error for invalid file risk level, got nil")
	}

	if !strings.Contains(err.Error(), "invalid file risk level") {
		t.Errorf("Expected 'invalid file risk level' error, got: %v", err)
	}
}

func TestValidateAnalysis_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		analysis MergeAnalysis
	}{
		{
			name: "Valid low risk",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: "low"},
				},
				RiskLevel: "low",
			},
		},
		{
			name: "Valid medium risk",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: "medium"},
				},
				RiskLevel: "medium",
			},
		},
		{
			name: "Valid high risk",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: "high"},
				},
				RiskLevel: "high",
			},
		},
		{
			name: "Case insensitive risk level",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: "LOW"},
				},
				RiskLevel: "HIGH",
			},
		},
		{
			name: "Empty risk level (allowed)",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: ""},
				},
				RiskLevel: "",
			},
		},
		{
			name: "Multiple files",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test1.go", RiskLevel: "low"},
					{Path: "test2.go", RiskLevel: "medium"},
					{Path: "test3.go", RiskLevel: "high"},
				},
				RiskLevel: "medium",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAnalysis(tt.analysis)
			if err != nil {
				t.Errorf("Expected no error for valid input, got: %v", err)
			}
		})
	}
}

func TestSanitizePath_PathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple relative path",
			input:    "file.go",
			expected: "file.go",
		},
		{
			name:     "Path with directory",
			input:    "internal/merge/file.go",
			expected: filepath.FromSlash("internal/merge/file.go"),
		},
		{
			name:     "Path traversal with ../",
			input:    "../../../etc/passwd",
			expected: filepath.FromSlash("etc/passwd"),
		},
		{
			name:     "Path traversal with ./",
			input:    "./file.go",
			expected: "file.go",
		},
		{
			name:     "Absolute path",
			input:    "/etc/passwd",
			expected: filepath.FromSlash("etc/passwd"),
		},
		{
			name:     "Complex path traversal",
			input:    "../.././internal/../config.yaml",
			expected: "config.yaml",
		},
		{
			name:     "Multiple leading slashes",
			input:    "///etc/passwd",
			expected: filepath.FromSlash("etc/passwd"),
		},
		{
			name:     "Nested path traversal",
			input:    "a/../../b/../c/./d.go",
			expected: filepath.FromSlash("c/d.go"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.input)
			// Normalize paths to use forward slashes for cross-platform comparison
			result = filepath.ToSlash(result)
			expected := filepath.ToSlash(tt.expected)
			if result != expected {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, result, expected)
			}
		})
	}
}

func TestConfirmMerge_ValidationError(t *testing.T) {
	// Test that ConfirmMerge rejects invalid input
	tests := []struct {
		name     string
		analysis MergeAnalysis
		wantErr  string
	}{
		{
			name: "Too many files",
			analysis: MergeAnalysis{
				Files:     make([]FileAnalysis, 1001),
				RiskLevel: "low",
			},
			wantErr: "invalid analysis",
		},
		{
			name: "Invalid risk level",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: "test.go", RiskLevel: "low"},
				},
				RiskLevel: "invalid",
			},
			wantErr: "invalid analysis",
		},
		{
			name: "Path too long",
			analysis: MergeAnalysis{
				Files: []FileAnalysis{
					{Path: strings.Repeat("a", 1025), RiskLevel: "low"},
				},
				RiskLevel: "low",
			},
			wantErr: "invalid analysis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConfirmMerge(tt.analysis)
			if err == nil {
				t.Error("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

// AnalysisFormatter Tests

func TestNewAnalysisFormatter(t *testing.T) {
	analysis := MergeAnalysis{
		Summary:   "Test summary",
		RiskLevel: "medium",
	}

	formatter := NewAnalysisFormatter(analysis)

	if formatter == nil {
		t.Fatal("NewAnalysisFormatter returned nil")
	}

	if formatter.analysis.Summary != analysis.Summary {
		t.Errorf("Expected summary %q, got %q", analysis.Summary, formatter.analysis.Summary)
	}

	if formatter.analysis.RiskLevel != analysis.RiskLevel {
		t.Errorf("Expected risk level %q, got %q", analysis.RiskLevel, formatter.analysis.RiskLevel)
	}
}

func TestAnalysisFormatter_FormatTitle(t *testing.T) {
	formatter := NewAnalysisFormatter(MergeAnalysis{})

	title := formatter.FormatTitle()

	if title == "" {
		t.Error("FormatTitle returned empty string")
	}

	// Note: We can't check for exact styling (ANSI codes), but we can check for content
	if !strings.Contains(title, "Merge Analysis Results") {
		t.Errorf("Expected title to contain 'Merge Analysis Results', got: %s", title)
	}
}

func TestAnalysisFormatter_FormatSummary(t *testing.T) {
	tests := []struct {
		name        string
		summary     string
		wantEmpty   bool
		wantContain string
	}{
		{
			name:        "With summary",
			summary:     "Test summary message",
			wantEmpty:   false,
			wantContain: "Test summary message",
		},
		{
			name:      "Empty summary",
			summary:   "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(MergeAnalysis{Summary: tt.summary})
			result := formatter.FormatSummary()

			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			if !tt.wantEmpty && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Expected result to contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestAnalysisFormatter_FormatRiskLevel(t *testing.T) {
	tests := []struct {
		name        string
		riskLevel   string
		wantEmpty   bool
		wantContain string
	}{
		{
			name:        "Low risk",
			riskLevel:   "low",
			wantEmpty:   false,
			wantContain: "low",
		},
		{
			name:        "Medium risk",
			riskLevel:   "medium",
			wantEmpty:   false,
			wantContain: "medium",
		},
		{
			name:        "High risk",
			riskLevel:   "high",
			wantEmpty:   false,
			wantContain: "high",
		},
		{
			name:      "Empty risk level",
			riskLevel: "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(MergeAnalysis{})
			result := formatter.FormatRiskLevel(tt.riskLevel)

			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			if !tt.wantEmpty && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Expected result to contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestAnalysisFormatter_FormatOverallRisk(t *testing.T) {
	tests := []struct {
		name        string
		riskLevel   string
		wantEmpty   bool
		wantContain string
	}{
		{
			name:        "With risk level",
			riskLevel:   "high",
			wantEmpty:   false,
			wantContain: "Risk Level: high",
		},
		{
			name:      "Empty risk level",
			riskLevel: "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(MergeAnalysis{RiskLevel: tt.riskLevel})
			result := formatter.FormatOverallRisk()

			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			if !tt.wantEmpty && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Expected result to contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestAnalysisFormatter_FormatFileTable(t *testing.T) {
	tests := []struct {
		name        string
		files       []FileAnalysis
		wantEmpty   bool
		wantContain []string
	}{
		{
			name: "Single file",
			files: []FileAnalysis{
				{
					Path:      "test.go",
					Changes:   "modified",
					Strategy:  "overwrite",
					RiskLevel: "low",
				},
			},
			wantEmpty: false,
			wantContain: []string{
				"File",
				"Changes",
				"Strategy",
				"Risk",
				"test.go",
				"modified",
				"overwrite",
				"low",
			},
		},
		{
			name: "Multiple files",
			files: []FileAnalysis{
				{Path: "file1.go", Changes: "added", Strategy: "create", RiskLevel: "low"},
				{Path: "file2.go", Changes: "modified", Strategy: "merge", RiskLevel: "medium"},
				{Path: "file3.go", Changes: "deleted", Strategy: "remove", RiskLevel: "high"},
			},
			wantEmpty: false,
			wantContain: []string{
				"file1.go",
				"file2.go",
				"file3.go",
				"added",
				"modified",
				"deleted",
			},
		},
		{
			name:      "No files",
			files:     []FileAnalysis{},
			wantEmpty: true,
		},
		{
			name: "Long file path truncation",
			files: []FileAnalysis{
				{
					Path:      "very/long/path/to/some/deeply/nested/file.go",
					Changes:   "modified",
					Strategy:  "overwrite",
					RiskLevel: "low",
				},
			},
			wantEmpty: false,
			wantContain: []string{
				"...",
				"file.go",
			},
		},
		{
			name: "Long changes truncation",
			files: []FileAnalysis{
				{
					Path:      "test.go",
					Changes:   "very long description of changes",
					Strategy:  "overwrite",
					RiskLevel: "low",
				},
			},
			wantEmpty: false,
			wantContain: []string{
				"...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(MergeAnalysis{Files: tt.files})
			result := formatter.FormatFileTable()

			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(result, want) {
					t.Errorf("Expected result to contain %q, got: %s", want, result)
				}
			}
		})
	}
}

func TestAnalysisFormatter_FormatConflictWarning(t *testing.T) {
	tests := []struct {
		name        string
		analysis    MergeAnalysis
		wantEmpty   bool
		wantContain string
	}{
		{
			name: "With conflicts",
			analysis: MergeAnalysis{
				HasConflicts: true,
				Files: []FileAnalysis{
					{Path: "file1.go", RiskLevel: "high"},
					{Path: "file2.go", RiskLevel: "high"},
					{Path: "file3.go", RiskLevel: "low"},
				},
			},
			wantEmpty:   false,
			wantContain: "2 file(s) with high risk conflicts detected",
		},
		{
			name: "No conflicts",
			analysis: MergeAnalysis{
				HasConflicts: false,
				Files: []FileAnalysis{
					{Path: "file1.go", RiskLevel: "low"},
				},
			},
			wantEmpty: true,
		},
		{
			name: "Conflicts flag true but no high risk files",
			analysis: MergeAnalysis{
				HasConflicts: true,
				Files: []FileAnalysis{
					{Path: "file1.go", RiskLevel: "low"},
					{Path: "file2.go", RiskLevel: "medium"},
				},
			},
			wantEmpty:   false,
			wantContain: "0 file(s) with high risk conflicts detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(tt.analysis)
			result := formatter.FormatConflictWarning()

			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			if !tt.wantEmpty && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Expected result to contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestAnalysisFormatter_FormatPrompt(t *testing.T) {
	formatter := NewAnalysisFormatter(MergeAnalysis{})
	result := formatter.FormatPrompt()

	if result == "" {
		t.Error("FormatPrompt returned empty string")
	}

	if !strings.Contains(result, "[S] Toggle Selection Mode") {
		t.Errorf("Expected prompt to contain '[S] Toggle Selection Mode', got: %s", result)
	}

	if !strings.Contains(result, "[Y]es to merge") {
		t.Errorf("Expected prompt to contain '[Y]es to merge', got: %s", result)
	}

	if !strings.Contains(result, "[N]o to cancel") {
		t.Errorf("Expected prompt to contain '[N]o to cancel', got: %s", result)
	}
}

func TestAnalysisFormatter_Render(t *testing.T) {
	tests := []struct {
		name        string
		analysis    MergeAnalysis
		wantContain []string
	}{
		{
			name: "Complete analysis",
			analysis: MergeAnalysis{
				Summary:      "Test merge operation",
				RiskLevel:    "medium",
				HasConflicts: true,
				Files: []FileAnalysis{
					{
						Path:      "test.go",
						Changes:   "modified",
						Strategy:  "overwrite",
						RiskLevel: "high",
					},
				},
			},
			wantContain: []string{
				"Merge Analysis Results",
				"Test merge operation",
				"Risk Level: medium",
				"test.go",
				"modified",
				"overwrite",
				"Warning",
				"[Y]es to merge",
			},
		},
		{
			name: "Minimal analysis",
			analysis: MergeAnalysis{
				Summary:   "",
				RiskLevel: "",
				Files:     []FileAnalysis{},
			},
			wantContain: []string{
				"Merge Analysis Results",
				"[Y]es to merge",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalysisFormatter(tt.analysis)
			result := formatter.Render()

			if result == "" {
				t.Error("Render returned empty string")
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(result, want) {
					t.Errorf("Expected result to contain %q, got: %s", want, result)
				}
			}
		})
	}
}

func TestAnalysisFormatter_TruncatePath(t *testing.T) {
	formatter := NewAnalysisFormatter(MergeAnalysis{})

	tests := []struct {
		name     string
		path     string
		wantLen  int
		contains string
	}{
		{
			name:     "Short path",
			path:     "test.go",
			wantLen:  7,
			contains: "test.go",
		},
		{
			name:     "Long path",
			path:     "very/long/path/to/some/deeply/nested/file.go",
			wantLen:  38,
			contains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.truncatePath(tt.path)

			// Check path length is within expected bounds
			if len(result) > 38 {
				t.Errorf("Expected path length <= 38, got %d", len(result))
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

func TestAnalysisFormatter_TruncateChanges(t *testing.T) {
	formatter := NewAnalysisFormatter(MergeAnalysis{})

	tests := []struct {
		name     string
		changes  string
		maxLen   int
		contains string
	}{
		{
			name:     "Short changes",
			changes:  "modified",
			maxLen:   13,
			contains: "modified",
		},
		{
			name:     "Long changes",
			changes:  "very long description of changes that exceeds limit",
			maxLen:   13,
			contains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.truncateChanges(tt.changes)

			if len(result) > tt.maxLen {
				t.Errorf("Expected changes length <= %d, got %d", tt.maxLen, len(result))
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}
