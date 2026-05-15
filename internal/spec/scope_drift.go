package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Task represents a single task from tasks.md
type Task struct {
	ID           string
	Description  string
	Requirement  string
	Dependencies string
	PlannedFiles []string
	Status       string
}

// ScopeDriftReport represents the scope drift detection result
type ScopeDriftReport struct {
	SPECID            string
	PlannedFileCount  int
	ActualFileCount   int
	UnplannedFiles    []string
	DriftPercentage   float64
	CumulativeDrift   float64
	DriftLevel        string // "informational", "warning", "critical"
	TriggerReplanning bool
}

// ParseTasksMD parses the tasks.md file and extracts task information
func ParseTasksMD(specDir string) ([]Task, error) {
	tasksPath := filepath.Join(specDir, "tasks.md")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tasks.md not found: %s", tasksPath)
	}

	file, err := os.Open(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tasks.md: %w", err)
	}
	defer file.Close()

	var tasks []Task
	var inTable bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect table start
		if strings.Contains(line, "| Task ID |") {
			inTable = true
			continue
		}

		// Skip table separator and empty lines
		if !inTable || strings.HasPrefix(line, "|--") || strings.TrimSpace(line) == "" {
			continue
		}

		// Exit table section
		if inTable && !strings.HasPrefix(line, "|") {
			break
		}

		// Parse table row
		if inTable && strings.HasPrefix(line, "|") {
			task, err := parseTaskRow(line)
			if err != nil {
				continue // Skip malformed rows
			}
			tasks = append(tasks, task)
		}
	}

	return tasks, scanner.Err()
}

// parseTaskRow parses a single markdown table row into a Task
func parseTaskRow(line string) (Task, error) {
	// Remove leading/trailing pipes and split
	parts := strings.Split(strings.Trim(line, "|"), "|")
	if len(parts) < 6 {
		return Task{}, fmt.Errorf("invalid task row format: need at least 6 columns, got %d", len(parts))
	}

	// Trim whitespace from each part
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Parse planned files (column 5, 0-indexed)
	plannedFiles := parseFileList(parts[3])

	// Get status (column 6, 0-indexed) if available, default to "pending"
	status := "pending"
	if len(parts) > 6 {
		status = parts[5]
	}

	return Task{
		ID:           parts[0],
		Description:  parts[1],
		Requirement:  parts[2],
		Dependencies: parts[3],
		PlannedFiles: plannedFiles,
		Status:       status,
	}, nil
}

// parseFileList extracts file paths from a comma-separated list
func parseFileList(fileListStr string) []string {
	if fileListStr == "-" || strings.TrimSpace(fileListStr) == "" {
		return nil
	}

	files := strings.Split(fileListStr, ",")
	var result []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}

// ExclusionPatterns defines file patterns to exclude from drift detection
var ExclusionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.git$`),
	regexp.MustCompile(`\.gitignore`),
	regexp.MustCompile(`\.gitattributes`),
	regexp.MustCompile(`_test\.go$`),
	regexp.MustCompile(`test/fixtures/`),
	regexp.MustCompile(`testdata/`),
	regexp.MustCompile(`\.md$`),
	regexp.MustCompile(`\.generated\.`),
}

// ShouldExcludeFile checks if a file should be excluded from drift detection
func ShouldExcludeFile(filePath string) bool {
	for _, pattern := range ExclusionPatterns {
		if pattern.MatchString(filePath) {
			return true
		}
	}
	return false
}

// CalculateScopeDrift calculates the drift percentage and identifies unplanned files
func CalculateScopeDrift(plannedFiles, actualFiles []string) (float64, []string) {
	// Build set of planned files (normalized)
	plannedSet := make(map[string]bool)
	for _, f := range plannedFiles {
		plannedSet[filepath.Clean(f)] = true
	}

	// Identify unplanned files (excluding patterns)
	var unplanned []string
	for _, f := range actualFiles {
		cleanPath := filepath.Clean(f)
		if !plannedSet[cleanPath] && !ShouldExcludeFile(cleanPath) {
			unplanned = append(unplanned, cleanPath)
		}
	}

	// Calculate drift percentage
	driftPercentage := 0.0
	if len(plannedFiles) > 0 {
		driftPercentage = (float64(len(unplanned)) / float64(len(plannedFiles))) * 100.0
	}

	return driftPercentage, unplanned
}

// DetectScopeDrift performs a complete scope drift analysis for a SPEC
func DetectScopeDrift(specDir string, actualFiles []string) (*ScopeDriftReport, error) {
	tasks, err := ParseTasksMD(specDir)
	if err != nil {
		return nil, err
	}

	// Collect all planned files
	var plannedFiles []string
	for _, task := range tasks {
		plannedFiles = append(plannedFiles, task.PlannedFiles...)
	}

	// Calculate drift
	driftPercentage, unplannedFiles := CalculateScopeDrift(plannedFiles, actualFiles)

	// Determine drift level
	driftLevel := "informational"
	triggerReplanning := false

	if driftPercentage > 30.0 {
		driftLevel = "critical"
		triggerReplanning = true
	} else if driftPercentage > 20.0 {
		driftLevel = "warning"
	}

	// Extract SPEC-ID from path
	specID := filepath.Base(specDir)

	return &ScopeDriftReport{
		SPECID:            specID,
		PlannedFileCount:  len(plannedFiles),
		ActualFileCount:   len(actualFiles),
		UnplannedFiles:    unplannedFiles,
		DriftPercentage:   driftPercentage,
		DriftLevel:        driftLevel,
		TriggerReplanning: triggerReplanning,
	}, nil
}
