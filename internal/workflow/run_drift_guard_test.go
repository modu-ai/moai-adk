package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTasksMdGeneration tests REQ-001: Persistent tasks.md Artifact
func TestTasksMdGeneration(t *testing.T) {
	t.Run("Phase_1_5_creates_tasks_md_with_required_fields", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		specContent := "# Spec Drift Guard\n\n## Requirements\n\n### REQ-001: Persistent tasks.md Artifact\nWhen the Run phase reaches Phase 1.5, the system shall output a tasks.md file.\n\n### REQ-002: Real-Time Drift Guard\nWhen a DDD or TDD cycle completes, the system shall compare planned vs actual files.\n"
		specPath := filepath.Join(specDir, "spec.md")
		if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
			t.Fatalf("Failed to write spec.md: %v", err)
		}

		tasks, err := GenerateTasksFromSpec(specPath)
		if err != nil {
			t.Fatalf("GenerateTasksFromSpec failed: %v", err)
		}

		tasksPath := filepath.Join(specDir, "tasks.md")
		if err := WriteTasksMd(tasksPath, tasks); err != nil {
			t.Fatalf("WriteTasksMd failed: %v", err)
		}

		content, err := os.ReadFile(tasksPath)
		if err != nil {
			t.Fatalf("Failed to read tasks.md: %v", err)
		}

		contentStr := string(content)
		requiredHeaders := []string{"Task ID", "Description", "Requirement", "Dependencies", "Planned Files", "Status"}

		for _, header := range requiredHeaders {
			if !strings.Contains(contentStr, header) {
				t.Errorf("tasks.md missing required header: %s", header)
			}
		}

		if !strings.Contains(contentStr, "|") {
			t.Error("tasks.md should use markdown table format")
		}
	})
}

// TestDriftGuardCalculation tests REQ-002: Real-Time Drift Guard
func TestDriftGuardCalculation(t *testing.T) {
	t.Run("Zero_drift_when_files_match_plan", func(t *testing.T) {
		plannedFiles := []string{"file1.go", "file2.go", "file3.go"}
		actualFiles := []string{"file1.go", "file2.go", "file3.go"}

		drift := CalculateDrift(plannedFiles, actualFiles)

		if drift.Percentage != 0.0 {
			t.Errorf("Expected 0%% drift, got %.2f%%", drift.Percentage)
		}
		if len(drift.UnplannedFiles) != 0 {
			t.Errorf("Expected no unplanned files, got %d", len(drift.UnplannedFiles))
		}
	})

	t.Run("Detects_unplanned_files", func(t *testing.T) {
		plannedFiles := []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go"}
		actualFiles := []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go", "unplanned_util.go"}

		drift := CalculateDrift(plannedFiles, actualFiles)

		expectedDrift := 10.0
		if drift.Percentage != expectedDrift {
			t.Errorf("Expected %.2f%% drift, got %.2f%%", expectedDrift, drift.Percentage)
		}
		if len(drift.UnplannedFiles) != 1 {
			t.Errorf("Expected 1 unplanned file, got %d", len(drift.UnplannedFiles))
		}
	})
}

// TestScopeAlarmIntegration tests REQ-003: Scope Alarm Integration
func TestScopeAlarmIntegration(t *testing.T) {
	t.Run("Warning_below_30_percent", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		progressPath := filepath.Join(specDir, "progress.md")

		drift := DriftResult{Percentage: 25.0, UnplannedFiles: []string{"extra.go"}, PlannedCount: 10, ActualCount: 12}

		err = LogDriftToProgress(progressPath, drift, 1)
		if err != nil {
			t.Fatalf("LogDriftToProgress failed: %v", err)
		}

		content, err := os.ReadFile(progressPath)
		if err != nil {
			t.Fatalf("Failed to read progress.md: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "drift: 25%") {
			t.Error("progress.md should contain drift percentage")
		}
		if !strings.Contains(contentStr, "WARNING") {
			t.Error("progress.md should contain warning for >20% drift")
		}
	})

	t.Run("Trigger_replanning_above_30_percent", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		progressPath := filepath.Join(specDir, "progress.md")

		drift := DriftResult{Percentage: 35.0, UnplannedFiles: []string{"extra1.go", "extra2.go", "extra3.go"}, PlannedCount: 10, ActualCount: 13}

		err = LogDriftToProgress(progressPath, drift, 1)
		if err != nil {
			t.Fatalf("LogDriftToProgress failed: %v", err)
		}

		content, err := os.ReadFile(progressPath)
		if err != nil {
			t.Fatalf("Failed to read progress.md: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "drift: 35%") {
			t.Error("progress.md should contain drift percentage")
		}
		if !strings.Contains(contentStr, "CRITICAL") {
			t.Error("progress.md should contain CRITICAL for >30% drift")
		}
		if !strings.Contains(contentStr, "Phase 2.7 re-planning") {
			t.Error("progress.md should trigger Phase 2.7 re-planning gate")
		}
	})
}

// TestDriftGuardIntegration tests full integration scenarios from SPEC
func TestDriftGuardIntegration(t *testing.T) {
	t.Run("Clean_implementation_no_drift", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		tasks := []Task{
			{ID: "T-001", PlannedFiles: []string{"file1.go", "file2.go", "file3.go"}},
			{ID: "T-002", PlannedFiles: []string{"file4.go", "file5.go", "file6.go"}},
			{ID: "T-003", PlannedFiles: []string{"file7.go", "file8.go", "file9.go"}},
			{ID: "T-004", PlannedFiles: []string{"file10.go"}},
			{ID: "T-005", PlannedFiles: []string{"file11.go", "file12.go"}},
		}

		tasksPath := filepath.Join(specDir, "tasks.md")
		if err := WriteTasksMd(tasksPath, tasks); err != nil {
			t.Fatalf("WriteTasksMd failed: %v", err)
		}

		plannedFiles := getAllPlannedFiles(tasks)
		actualFiles := plannedFiles

		drift := CalculateDrift(plannedFiles, actualFiles)

		if drift.Percentage != 0.0 {
			t.Errorf("Scenario 1: Expected 0%% drift, got %.2f%%", drift.Percentage)
		}

		progressPath := filepath.Join(specDir, "progress.md")
		if err := LogDriftToProgress(progressPath, drift, 1); err != nil {
			t.Fatalf("LogDriftToProgress failed: %v", err)
		}

		content, _ := os.ReadFile(progressPath)
		if strings.Contains(string(content), "WARNING") || strings.Contains(string(content), "ALERT") {
			t.Error("Scenario 1: Should not generate alerts for clean implementation")
		}
	})

	t.Run("Minor_drift_below_threshold", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		plannedFiles := []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go"}
		actualFiles := append(plannedFiles, "utility_helper.go")

		drift := CalculateDrift(plannedFiles, actualFiles)

		if drift.Percentage != 10.0 {
			t.Errorf("Scenario 2: Expected 10%% drift, got %.2f%%", drift.Percentage)
		}

		progressPath := filepath.Join(specDir, "progress.md")
		if err := LogDriftToProgress(progressPath, drift, 1); err != nil {
			t.Fatalf("LogDriftToProgress failed: %v", err)
		}

		content, _ := os.ReadFile(progressPath)
		contentStr := string(content)
		if !strings.Contains(contentStr, "drift: 10%") {
			t.Error("Scenario 2: Should record drift percentage")
		}
		if strings.Contains(contentStr, "WARNING") {
			t.Error("Scenario 2: Should not generate warning for 10% drift (below 20% threshold)")
		}
	})

	t.Run("Significant_drift_above_threshold", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		plannedFiles := []string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go", "file7.go", "file8.go", "file9.go", "file10.go"}
		actualFiles := append(plannedFiles, "extra1.go", "extra2.go", "extra3.go", "extra4.go")

		drift := CalculateDrift(plannedFiles, actualFiles)

		if drift.Percentage != 40.0 {
			t.Errorf("Scenario 3: Expected 40%% drift, got %.2f%%", drift.Percentage)
		}

		progressPath := filepath.Join(specDir, "progress.md")
		if err := LogDriftToProgress(progressPath, drift, 1); err != nil {
			t.Fatalf("LogDriftToProgress failed: %v", err)
		}

		content, _ := os.ReadFile(progressPath)
		contentStr := string(content)
		if !strings.Contains(contentStr, "drift: 40%") {
			t.Error("Scenario 3: Should record drift percentage")
		}
		if !strings.Contains(contentStr, "CRITICAL") {
			t.Error("Scenario 3: Should generate CRITICAL alert for >30% drift")
		}
		if !strings.Contains(contentStr, "Phase 2.7 re-planning") {
			t.Error("Scenario 3: Should trigger Phase 2.7 re-planning gate")
		}
	})
}

// TestGitTrackedTasksMd tests that tasks.md is suitable for git tracking
func TestGitTrackedTasksMd(t *testing.T) {
	t.Run("Deterministic_output", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-DRIFT-001")
		err := os.MkdirAll(specDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create spec dir: %v", err)
		}

		specContent := "# Test SPEC\n\n## Requirements\n\n### REQ-001: Test requirement\nTest requirement description.\n"
		specPath := filepath.Join(specDir, "spec.md")
		if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
			t.Fatalf("Failed to write spec.md: %v", err)
		}

		tasksPath := filepath.Join(specDir, "tasks.md")

		tasks1, err := GenerateTasksFromSpec(specPath)
		if err != nil {
			t.Fatalf("First GenerateTasksFromSpec failed: %v", err)
		}
		if err := WriteTasksMd(tasksPath, tasks1); err != nil {
			t.Fatalf("First WriteTasksMd failed: %v", err)
		}
		content1, _ := os.ReadFile(tasksPath)

		time.Sleep(10 * time.Millisecond)

		tasks2, err := GenerateTasksFromSpec(specPath)
		if err != nil {
			t.Fatalf("Second GenerateTasksFromSpec failed: %v", err)
		}
		if err := WriteTasksMd(tasksPath, tasks2); err != nil {
			t.Fatalf("Second WriteTasksMd failed: %v", err)
		}
		content2, _ := os.ReadFile(tasksPath)

		if string(content1) != string(content2) {
			t.Error("tasks.md should have deterministic output (no timestamps)")
		}
	})
}
