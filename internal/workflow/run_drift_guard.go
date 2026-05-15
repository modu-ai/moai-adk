package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Task represents a single implementation task from Phase 1.5 decomposition
type Task struct {
	ID                 string   // Sequential ID (e.g., "T-001", "T-002")
	Description        string   // Clear action statement
	Requirement        string   // Mapped REQ-XXX reference
	Dependencies       []string // List of prerequisite task IDs
	PlannedFiles       []string // Files expected to be modified/created
	Status             string   // pending, in-progress, completed, skipped
	AcceptanceCriteria string   // How to verify completion
}

// DriftResult represents the result of a drift calculation
type DriftResult struct {
	Percentage     float64   // Drift percentage (unplanned / planned * 100)
	UnplannedFiles []string  // List of files not in the plan
	PlannedCount   int       // Number of planned files
	ActualCount    int       // Number of actual files
	Timestamp      time.Time // When drift was calculated
	CycleNumber    int       // DDD/TDD cycle number
}

// ExceedsThreshold checks if drift exceeds the given threshold percentage
func (d DriftResult) ExceedsThreshold(threshold float64) bool {
	return d.Percentage > threshold
}

// GenerateTasksFromSpec parses a spec.md file and generates task decomposition
// REQ-001: Extract requirements and create structured tasks
func GenerateTasksFromSpec(specPath string) ([]Task, error) {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec.md: %w", err)
	}

	specContent := string(content)

	// Extract requirements using line-by-line parsing
	var tasks []Task
	taskID := 1

	scanner := bufio.NewScanner(strings.NewReader(specContent))
	currentReqID := ""
	currentReqTitle := ""
	var reqLines []string
	inRequirement := false

	reqHeaderRegex := regexp.MustCompile(`^### (REQ-\d+):\s*(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a requirement header
		if matches := reqHeaderRegex.FindStringSubmatch(line); len(matches) > 2 {
			// Save previous requirement if exists
			if inRequirement && currentReqID != "" {
				reqDescription := strings.Join(reqLines, "\n")
				task := Task{
					ID:                 fmt.Sprintf("T-%03d", taskID),
					Description:        currentReqTitle,
					Requirement:        currentReqID,
					Dependencies:       []string{},
					PlannedFiles:       extractPlannedFiles(reqDescription),
					Status:             "pending",
					AcceptanceCriteria: fmt.Sprintf("Implement %s", currentReqTitle),
				}
				tasks = append(tasks, task)
				taskID++
			}

			// Start new requirement
			currentReqID = matches[1]
			currentReqTitle = matches[2]
			reqLines = []string{}
			inRequirement = true
		} else if inRequirement {
			// Check if we've hit the next section
			if strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "###") {
				inRequirement = false
			} else if strings.HasPrefix(line, "### ") && !strings.HasPrefix(line, "### REQ-") {
				inRequirement = false
			} else {
				reqLines = append(reqLines, line)
			}
		}
	}

	// Don't forget the last requirement
	if inRequirement && currentReqID != "" {
		reqDescription := strings.Join(reqLines, "\n")
		task := Task{
			ID:                 fmt.Sprintf("T-%03d", taskID),
			Description:        currentReqTitle,
			Requirement:        currentReqID,
			Dependencies:       []string{},
			PlannedFiles:       extractPlannedFiles(reqDescription),
			Status:             "pending",
			AcceptanceCriteria: fmt.Sprintf("Implement %s", currentReqTitle),
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		// If no structured requirements found, create default task
		task := Task{
			ID:                 "T-001",
			Description:        "Implement SPEC requirements",
			Requirement:        "REQ-ALL",
			Dependencies:       []string{},
			PlannedFiles:       []string{},
			Status:             "pending",
			AcceptanceCriteria: "Complete all SPEC requirements",
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// extractPlannedFiles extracts planned file paths from requirement description
func extractPlannedFiles(description string) []string {
	// Look for file patterns in description
	// Common patterns: "file:", "modify:", "create:", path/to/file.ext
	fileRegex := regexp.MustCompile(`(?:file|modify|create):\s*([^\s,]+(?:\.[a-z]+))`)
	matches := fileRegex.FindAllStringSubmatch(description, -1)

	var files []string
	for _, match := range matches {
		if len(match) > 1 {
			files = append(files, match[1])
		}
	}

	// If no explicit files found, look for any path-like patterns
	pathRegex := regexp.MustCompile(`[\w-]+\.\w+`)
	pathMatches := pathRegex.FindAllString(description, -1)

	for _, path := range pathMatches {
		// Filter out common non-file extensions
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".md" || ext == ".yaml" || ext == ".json" || ext == ".sh" {
			files = append(files, path)
		}
	}

	return files
}

// WriteTasksMd writes the task decomposition to tasks.md file
// REQ-001: Output .moai/specs/SPEC-{ID}/tasks.md with all required fields
func WriteTasksMd(tasksPath string, tasks []Task) error {
	var sb strings.Builder

	sb.WriteString("## Task Decomposition\n\n")
	sb.WriteString("| Task ID | Description | Requirement | Dependencies | Planned Files | Status |\n")
	sb.WriteString("|---------|-------------|-------------|--------------|---------------|--------|\n")

	for _, task := range tasks {
		// Escape pipe characters in description
		desc := strings.ReplaceAll(task.Description, "|", "\\|")
		req := strings.ReplaceAll(task.Requirement, "|", "\\|")

		// Format dependencies
		deps := "None"
		if len(task.Dependencies) > 0 {
			deps = strings.Join(task.Dependencies, ", ")
		}

		// Format planned files
		files := "None"
		if len(task.PlannedFiles) > 0 {
			files = strings.Join(task.PlannedFiles, ", ")
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			task.ID, desc, req, deps, files, task.Status))
	}

	// Add metadata section (deterministic, no timestamps)
	sb.WriteString("\n## Metadata\n\n")
	sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", len(tasks)))

	pendingCount := 0
	for _, t := range tasks {
		if t.Status == "pending" {
			pendingCount++
		}
	}
	sb.WriteString(fmt.Sprintf("Pending: %d\n", pendingCount))

	content := sb.String()

	// Write with deterministic permissions
	return os.WriteFile(tasksPath, []byte(content), 0644)
}

// CalculateDrift compares planned files against actual files
// REQ-002: Calculate drift percentage and identify unplanned files
func CalculateDrift(plannedFiles, actualFiles []string) DriftResult {
	plannedSet := make(map[string]bool)
	for _, f := range plannedFiles {
		plannedSet[f] = true
	}

	var unplanned []string
	for _, f := range actualFiles {
		if !plannedSet[f] {
			unplanned = append(unplanned, f)
		}
	}

	plannedCount := len(plannedFiles)
	actualCount := len(actualFiles)
	unplannedCount := len(unplanned)

	percentage := 0.0
	if plannedCount > 0 {
		percentage = float64(unplannedCount) / float64(plannedCount) * 100.0
	}

	return DriftResult{
		Percentage:     percentage,
		UnplannedFiles: unplanned,
		PlannedCount:   plannedCount,
		ActualCount:    actualCount,
		Timestamp:      time.Now(),
	}
}

// LogDriftToProgress appends drift information to progress.md
// REQ-002: Log drift measurement with alert thresholds
// REQ-003: Append warning entry and trigger re-planning gate
func LogDriftToProgress(progressPath string, drift DriftResult, cycleNumber int) error {
	drift.CycleNumber = cycleNumber

	var entry strings.Builder

	entry.WriteString(fmt.Sprintf("\n### Cycle %d Drift Check\n", cycleNumber))
	entry.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", drift.Timestamp.Format(time.RFC3339)))
	entry.WriteString(fmt.Sprintf("- **Planned Files**: %d\n", drift.PlannedCount))
	entry.WriteString(fmt.Sprintf("- **Actual Files**: %d\n", drift.ActualCount))
	entry.WriteString(fmt.Sprintf("- **Drift**: %.1f%%\n", drift.Percentage))

	if len(drift.UnplannedFiles) > 0 {
		entry.WriteString("- **Unplanned Files**:\n")
		for _, f := range drift.UnplannedFiles {
			entry.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}

	// REQ-003: Alert thresholds
	if drift.Percentage > 30.0 {
		entry.WriteString("- **Status**: ⚠️ **CRITICAL** - Drift exceeds 30%\n")
		entry.WriteString("- **Action**: Trigger Phase 2.7 re-planning gate for scope review\n")
	} else if drift.Percentage > 20.0 {
		entry.WriteString("- **Status**: ⚠️ **WARNING** - Drift exceeds 20%\n")
	} else if drift.Percentage > 0 {
		entry.WriteString("- **Status**: ℹ️ **INFO** - Minor drift detected\n")
	} else {
		entry.WriteString("- **Status**: ✅ **OK** - No drift\n")
	}

	// Append to progress.md
	content, err := os.ReadFile(progressPath)
	var existingContent string
	if err == nil {
		existingContent = string(content)
	}

	newContent := existingContent + "\n" + entry.String()

	return os.WriteFile(progressPath, []byte(newContent), 0644)
}

// GetCumulativeDrift calculates cumulative drift from all cycles in progress.md
// REQ-003: Track cumulative scope expansion
func GetCumulativeDrift(progressPath string) (float64, error) {
	content, err := os.ReadFile(progressPath)
	if err != nil {
		// File doesn't exist yet, no drift
		return 0.0, nil
	}

	contentStr := string(content)

	// Extract all drift percentages using regex
	driftRegex := regexp.MustCompile(`- \*\*Drift\*\*:\s*([\d.]+)%`)
	matches := driftRegex.FindAllStringSubmatch(contentStr, -1)

	var cumulative float64
	for _, match := range matches {
		if len(match) > 1 {
			var drift float64
			fmt.Sscanf(match[1], "%f", &drift)
			cumulative += drift
		}
	}

	return cumulative, nil
}

// ShouldTriggerReplanning checks if cumulative drift exceeds threshold
// REQ-003: When cumulative scope expansion exceeds 30%, trigger Phase 2.7
func ShouldTriggerReplanning(progressPath string, threshold float64) (bool, error) {
	cumulative, err := GetCumulativeDrift(progressPath)
	if err != nil {
		return false, err
	}

	return cumulative > threshold, nil
}

// getAllPlannedFiles extracts all planned files from a list of tasks
func getAllPlannedFiles(tasks []Task) []string {
	var files []string
	seen := make(map[string]bool)

	for _, task := range tasks {
		for _, file := range task.PlannedFiles {
			if !seen[file] {
				seen[file] = true
				files = append(files, file)
			}
		}
	}

	return files
}

// UpdateTaskStatus updates the status of a specific task in tasks.md
func UpdateTaskStatus(tasksPath string, taskID string, newStatus string) error {
	tasks, err := parseTasksMd(tasksPath)
	if err != nil {
		return err
	}

	// Find and update the task
	found := false
	for i := range tasks {
		if tasks[i].ID == taskID {
			tasks[i].Status = newStatus
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Write back
	return WriteTasksMd(tasksPath, tasks)
}

// parseTasksMd parses an existing tasks.md file
func parseTasksMd(tasksPath string) ([]Task, error) {
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	// Skip header lines (find table start)
	tableStart := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "|---------") {
			tableStart = i + 1
			break
		}
	}

	if tableStart == -1 {
		return nil, fmt.Errorf("no task table found in tasks.md")
	}

	var tasks []Task
	for i := tableStart; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "|") {
			continue
		}

		// Parse table row
		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}

		// Extract fields (trim spaces and unescape)
		task := Task{
			ID:                 strings.TrimSpace(parts[1]),
			Description:        strings.ReplaceAll(strings.TrimSpace(parts[2]), "\\|", "|"),
			Requirement:        strings.ReplaceAll(strings.TrimSpace(parts[3]), "\\|", "|"),
			Dependencies:       parseList(strings.TrimSpace(parts[4])),
			PlannedFiles:        parseList(strings.TrimSpace(parts[5])),
			Status:             strings.TrimSpace(parts[6]),
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// parseList parses a comma-separated list field
func parseList(field string) []string {
	if field == "None" || field == "" {
		return []string{}
	}

	items := strings.Split(field, ",")
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
