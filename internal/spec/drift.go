package spec

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DriftRecord represents a single SPEC status drift entry
type DriftRecord struct {
	SPECID           string
	FrontmatterStatus string
	GitImpliedStatus  string
	Drifted          bool
}

// DriftReport represents a complete drift detection report
type DriftReport struct {
	Records []DriftRecord
	Count   int
}

// DetectDrift scans all SPECs and compares frontmatter status against git log
// Returns a report with all drift records and the total drift count
func DetectDrift(baseDir string) (*DriftReport, error) {
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return &DriftReport{Records: []DriftRecord{}, Count: 0}, nil
	}

	// Read all SPEC directories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var records []DriftRecord
	driftCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specDir := filepath.Join(specsDir, specID)

		// Parse frontmatter status
		frontmatterStatus, err := ParseStatus(specDir)
		if err != nil {
			// Skip SPECs that can't be parsed
			continue
		}

		// Get git-implied status
		gitStatus, err := getGitImpliedStatus(specID)
		if err != nil {
			// If git history is empty or unavailable, skip
			continue
		}

		// Check for drift
		drifted := frontmatterStatus != gitStatus

		record := DriftRecord{
			SPECID:           specID,
			FrontmatterStatus: frontmatterStatus,
			GitImpliedStatus:  gitStatus,
			Drifted:          drifted,
		}

		records = append(records, record)

		if drifted {
			driftCount++
		}
	}

	// Sort records by SPEC-ID for consistent output
	sort.Slice(records, func(i, j int) bool {
		return records[i].SPECID < records[j].SPECID
	})

	return &DriftReport{
		Records: records,
		Count:   driftCount,
	}, nil
}

// getGitImpliedStatus determines the status implied by git log for a SPEC
// It scans git log on main for the latest commit mentioning the SPEC-ID
// and classifies that commit to determine the implied status
func getGitImpliedStatus(specID string) (string, error) {
	// Determine main branch
	branch := "main"
	if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
		branch = "master"
	}

	// Get latest commit mentioning this SPEC-ID
	cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges", "--grep="+specID, "-1")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return "", fmt.Errorf("no git history found for %s", specID)
	}

	// Extract commit title (first line after commit hash)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if scanner.Scan() {
		line := scanner.Text()
		// Skip commit hash
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid git log format")
		}
		commitTitle := parts[1]

		// Classify the commit title to get status
		_, status, err := ClassifyPRTitle(commitTitle)
		if err != nil {
			return "", fmt.Errorf("failed to classify commit: %w", err)
		}

		if status == "" {
			// Unknown prefix - default to "in-progress" for partial work
			return "in-progress", nil
		}

		return status, nil
	}

	return "", fmt.Errorf("failed to parse git log output")
}

// DriftCount is a convenience function that returns only the drift count
func DriftCount(baseDir string) (int, error) {
	report, err := DetectDrift(baseDir)
	if err != nil {
		return 0, err
	}
	return report.Count, nil
}
