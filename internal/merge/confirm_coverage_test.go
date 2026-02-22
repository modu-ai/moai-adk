package merge

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Tests targeting uncovered branches in confirm.go to push merge coverage above 85%.

func TestConfirmModel_Update_ToggleSelectionMode(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
				{Path: "file2.go", RiskLevel: "medium"},
			},
		},
		decision: false,
		done:     false,
	}

	// Press 's' to toggle selection mode
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	result := updatedModel.(confirmModel)

	if !result.showSelection {
		t.Error("expected showSelection to be true after pressing 's'")
	}

	if len(result.selectedFiles) != 2 {
		t.Errorf("expected selectedFiles length 2, got %d", len(result.selectedFiles))
	}

	// All should be selected by default
	for i, s := range result.selectedFiles {
		if !s {
			t.Errorf("file %d should be selected by default", i)
		}
	}

	// Toggle selection mode off
	updatedModel2, _ := result.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	result2 := updatedModel2.(confirmModel)

	if result2.showSelection {
		t.Error("expected showSelection to be false after toggling off")
	}
}

func TestConfirmModel_Update_SelectAll(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
				{Path: "file2.go", RiskLevel: "medium"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{false, false},
	}

	// Press 'a' to select all
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	result := updatedModel.(confirmModel)

	for i, s := range result.selectedFiles {
		if !s {
			t.Errorf("file %d should be selected after 'a'", i)
		}
	}
}

func TestConfirmModel_Update_DeselectAll(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
				{Path: "file2.go", RiskLevel: "medium"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{true, true},
	}

	// Press 'd' to deselect all
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updatedModel.(confirmModel)

	for i, s := range result.selectedFiles {
		if s {
			t.Errorf("file %d should be deselected after 'd'", i)
		}
	}
}

func TestConfirmModel_Update_ToggleFileSelection(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
				{Path: "file2.go", RiskLevel: "medium"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{true, true},
		cursor:        0,
	}

	// Press space to toggle current file
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	result := updatedModel.(confirmModel)

	if result.selectedFiles[0] {
		t.Error("file 0 should be deselected after space")
	}
	if !result.selectedFiles[1] {
		t.Error("file 1 should still be selected")
	}
}

func TestConfirmModel_Update_NavigateUpDown(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go"},
				{Path: "file2.go"},
				{Path: "file3.go"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{true, true, true},
		cursor:        0,
	}

	// Navigate down with 'j'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result := updatedModel.(confirmModel)
	if result.cursor != 1 {
		t.Errorf("cursor = %d after 'j', want 1", result.cursor)
	}

	// Navigate down with 'down' key
	updatedModel2, _ := result.Update(tea.KeyMsg{Type: tea.KeyDown})
	result2 := updatedModel2.(confirmModel)
	if result2.cursor != 2 {
		t.Errorf("cursor = %d after 'down', want 2", result2.cursor)
	}

	// Navigate down at bottom (should not change)
	updatedModel3, _ := result2.Update(tea.KeyMsg{Type: tea.KeyDown})
	result3 := updatedModel3.(confirmModel)
	if result3.cursor != 2 {
		t.Errorf("cursor = %d, should stay at 2", result3.cursor)
	}

	// Navigate up with 'k'
	updatedModel4, _ := result3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	result4 := updatedModel4.(confirmModel)
	if result4.cursor != 1 {
		t.Errorf("cursor = %d after 'k', want 1", result4.cursor)
	}

	// Navigate up with 'up' key
	updatedModel5, _ := result4.Update(tea.KeyMsg{Type: tea.KeyUp})
	result5 := updatedModel5.(confirmModel)
	if result5.cursor != 0 {
		t.Errorf("cursor = %d after 'up', want 0", result5.cursor)
	}

	// Navigate up at top (should not change)
	updatedModel6, _ := result5.Update(tea.KeyMsg{Type: tea.KeyUp})
	result6 := updatedModel6.(confirmModel)
	if result6.cursor != 0 {
		t.Errorf("cursor = %d, should stay at 0", result6.cursor)
	}
}

func TestConfirmModel_Update_AcceptWithSelectionNoSelection(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{false}, // No files selected
	}

	// Press 'y' with no selection -> proceed with all files
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := updatedModel.(confirmModel)

	if !result.decision {
		t.Error("expected decision true")
	}
	if !result.done {
		t.Error("expected done true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestConfirmModel_Update_AcceptWithSelectionSomeSelected(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", RiskLevel: "low"},
				{Path: "file2.go", RiskLevel: "high"},
				{Path: "file3.go", RiskLevel: "medium"},
			},
		},
		showSelection: true,
		selectedFiles: []bool{true, false, true}, // file1 and file3 selected
	}

	// Press 'y' with some files selected
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := updatedModel.(confirmModel)

	if !result.decision {
		t.Error("expected decision true")
	}
	if !result.done {
		t.Error("expected done true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}

	// Analysis should be filtered to only selected files
	if len(result.analysis.Files) != 2 {
		t.Errorf("filtered analysis files = %d, want 2", len(result.analysis.Files))
	}
}

func TestFilterSelectedFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []FileAnalysis
		selected      []bool
		wantCount     int
		wantRisk      string
		wantConflicts bool
	}{
		{
			name: "high risk selected",
			files: []FileAnalysis{
				{Path: "a.go", RiskLevel: "low"},
				{Path: "b.go", RiskLevel: "high"},
			},
			selected:      []bool{false, true},
			wantCount:     1,
			wantRisk:      "high",
			wantConflicts: true,
		},
		{
			name: "medium risk only",
			files: []FileAnalysis{
				{Path: "a.go", RiskLevel: "medium"},
				{Path: "b.go", RiskLevel: "low"},
			},
			selected:      []bool{true, false},
			wantCount:     1,
			wantRisk:      "medium",
			wantConflicts: false,
		},
		{
			name: "low risk only",
			files: []FileAnalysis{
				{Path: "a.go", RiskLevel: "low"},
				{Path: "b.go", RiskLevel: "low"},
			},
			selected:      []bool{true, true},
			wantCount:     2,
			wantRisk:      "low",
			wantConflicts: false,
		},
		{
			name: "all selected with mixed risk",
			files: []FileAnalysis{
				{Path: "a.go", RiskLevel: "low"},
				{Path: "b.go", RiskLevel: "medium"},
				{Path: "c.go", RiskLevel: "high"},
			},
			selected:      []bool{true, true, true},
			wantCount:     3,
			wantRisk:      "high",
			wantConflicts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := confirmModel{
				analysis: MergeAnalysis{
					Files: tt.files,
				},
				selectedFiles: tt.selected,
			}

			result := m.filterSelectedFiles()

			if len(result.Files) != tt.wantCount {
				t.Errorf("filtered file count = %d, want %d", len(result.Files), tt.wantCount)
			}
			if result.RiskLevel != tt.wantRisk {
				t.Errorf("risk level = %q, want %q", result.RiskLevel, tt.wantRisk)
			}
			if result.HasConflicts != tt.wantConflicts {
				t.Errorf("has conflicts = %v, want %v", result.HasConflicts, tt.wantConflicts)
			}
		})
	}
}

func TestFilterSelectedFiles_SummaryContainsHighRisk(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "a.go", RiskLevel: "high"},
				{Path: "b.go", RiskLevel: "high"},
			},
		},
		selectedFiles: []bool{true, true},
	}

	result := m.filterSelectedFiles()

	if !strings.Contains(result.Summary, "2 high-risk files") {
		t.Errorf("summary should mention high-risk files, got: %s", result.Summary)
	}
}

func TestTruncatePath_SkillsPath(t *testing.T) {
	formatter := NewAnalysisFormatter(MergeAnalysis{})

	tests := []struct {
		name     string
		path     string
		contains string
	}{
		{
			name:     "skills with remaining path",
			path:     ".claude/skills/moai-backend/reference.md",
			contains: "skills/moai-backend/reference.md",
		},
		{
			name:     "skills without remaining",
			path:     ".claude/skills/moai-backend",
			contains: "skills/moai-backend",
		},
		{
			name:     "agents with remaining path",
			path:     ".claude/agents/moai/expert-backend.md",
			contains: "agents/moai/expert-backend.md",
		},
		{
			name:     "agents without remaining",
			path:     ".claude/agents/moai",
			contains: "agents/moai",
		},
		{
			name:     "rules path",
			path:     ".claude/rules/moai/core/constitution.md",
			contains: "rules/moai/core/constitution.md",
		},
		{
			name:     "commands path",
			path:     ".claude/commands/moai/plan.md",
			contains: "commands/moai/plan.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.truncatePath(tt.path)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("truncatePath(%q) = %q, want containing %q", tt.path, result, tt.contains)
			}
		})
	}
}

func TestAnalysisFormatter_FormatPrompt_SelectionMode(t *testing.T) {
	formatter := NewAnalysisFormatterWithSelection(
		MergeAnalysis{},
		0,
		[]bool{true, false},
		true, // showSelection = true
	)

	result := formatter.FormatPrompt()

	if !strings.Contains(result, "[S]election Mode") {
		t.Errorf("expected selection mode prompt, got: %s", result)
	}
	if !strings.Contains(result, "[Space] Toggle") {
		t.Errorf("expected space toggle instruction, got: %s", result)
	}
}

func TestAnalysisFormatter_FormatFileTable_WithSelection(t *testing.T) {
	formatter := NewAnalysisFormatterWithSelection(
		MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go", Changes: "added", Strategy: "create", RiskLevel: "low"},
				{Path: "file2.go", Changes: "modified", Strategy: "merge", RiskLevel: "medium"},
			},
		},
		0,                  // cursor on first item
		[]bool{true, false}, // first selected, second not
		true,                // showSelection = true
	)

	result := formatter.FormatFileTable()

	if result == "" {
		t.Fatal("expected non-empty table")
	}
}

func TestAnalysisFormatter_Render_WithSelection(t *testing.T) {
	formatter := NewAnalysisFormatterWithSelection(
		MergeAnalysis{
			Summary:   "Test merge",
			RiskLevel: "medium",
			Files: []FileAnalysis{
				{Path: "file1.go", Changes: "added", Strategy: "create", RiskLevel: "low"},
			},
		},
		0,
		[]bool{true},
		true, // showSelection = true
	)

	result := formatter.Render()

	if !strings.Contains(result, "Selected: 1 / 1 files") {
		t.Errorf("expected selection summary in render, got: %s", result)
	}
}

func TestToMapInterface(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantOK  bool
		wantLen int
	}{
		{
			name:    "map[string]any",
			input:   map[string]any{"key": "value"},
			wantOK:  true,
			wantLen: 1,
		},
		{
			name:    "map[any]any (YAML style)",
			input:   map[any]any{"key": "value", 42: "number"},
			wantOK:  true,
			wantLen: 2,
		},
		{
			name:   "string (not a map)",
			input:  "hello",
			wantOK: false,
		},
		{
			name:   "nil",
			input:  nil,
			wantOK: false,
		},
		{
			name:   "int",
			input:  42,
			wantOK: false,
		},
		{
			name:   "slice",
			input:  []string{"a", "b"},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toMapInterface(tt.input)
			if ok != tt.wantOK {
				t.Errorf("toMapInterface() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && len(result) != tt.wantLen {
				t.Errorf("toMapInterface() len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestConfirmModel_Update_NoOpWhenNotInSelectionMode(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{
				{Path: "file1.go"},
			},
		},
		showSelection: false,
		cursor:        0,
	}

	// Space should do nothing when not in selection mode
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	result := updatedModel.(confirmModel)
	if result.showSelection {
		t.Error("should not enter selection mode from space")
	}

	// 'a' should do nothing when not in selection mode
	updatedModel2, _ := result.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	result2 := updatedModel2.(confirmModel)
	if result2.showSelection {
		t.Error("should not enter selection mode from 'a'")
	}

	// Navigation keys should do nothing when not in selection mode
	updatedModel3, _ := result2.Update(tea.KeyMsg{Type: tea.KeyUp})
	result3 := updatedModel3.(confirmModel)
	if result3.cursor != 0 {
		t.Error("cursor should not move when not in selection mode")
	}
}

func TestConfirmModel_Update_ToggleSelectionOnEmptyFiles(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Files: []FileAnalysis{}, // No files
		},
		showSelection: false,
	}

	// 's' on empty files should not toggle selection mode
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	result := updatedModel.(confirmModel)
	if result.showSelection {
		t.Error("should not enable selection mode with no files")
	}
}

func TestConfirmModel_Update_UnhandledKey(t *testing.T) {
	m := confirmModel{
		analysis: MergeAnalysis{
			Summary: "Test",
		},
	}

	// Pressing an unhandled key should not change state
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	result := updatedModel.(confirmModel)

	if result.done {
		t.Error("unhandled key should not set done")
	}
	if result.decision {
		t.Error("unhandled key should not change decision")
	}
	if cmd != nil {
		t.Error("unhandled key should return nil command")
	}
}
