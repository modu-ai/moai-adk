package statusline

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/status"
)

// Formatter generates single-line status output
type Formatter struct {
	projectDir string
	version    string
}

// NewFormatter creates a new statusline formatter
func NewFormatter(projectDir, version string) *Formatter {
	return &Formatter{
		projectDir: projectDir,
		version:    version,
	}
}

// Format generates the statusline output
func (f *Formatter) Format(format string) (string, error) {
	if format == "" {
		format = "{version} | {branch} {state} | {specs} | {quality} | {status}"
	}

	result := format

	// Replace placeholders
	result = strings.ReplaceAll(result, "{version}", fmt.Sprintf("v%s", f.version))

	// Get git info
	gitInfo, _ := status.GetGitInfo()
	if gitInfo != nil {
		result = strings.ReplaceAll(result, "{branch}", gitInfo.Branch)

		stateSymbol := "✓"
		stateText := "Clean"
		if gitInfo.HasChanges {
			stateSymbol = "✗"
			stateText = "Modified"
		}
		result = strings.ReplaceAll(result, "{state}", stateSymbol)
		result = strings.ReplaceAll(result, "{statustext}", stateText)
	} else {
		result = strings.ReplaceAll(result, "{branch}", "no-git")
		result = strings.ReplaceAll(result, "{state}", "?")
		result = strings.ReplaceAll(result, "{statustext}", "Unknown")
	}

	// Get SPEC count
	specs, _ := status.GetActiveSpecs(f.projectDir) // nolint: errcheck // SPEC retrieval is optional
	specCount := len(specs)
	result = strings.ReplaceAll(result, "{specs}", fmt.Sprintf("%d SPECs", specCount))

	// Get quality status (simplified)
	quality, _ := status.GetQualityGateStatus(f.projectDir) // nolint: errcheck // Quality check is optional
	passedCount := 0
	totalCount := 5

	if quality.Tested.Passed {
		passedCount++
	}
	if quality.Readable.Passed {
		passedCount++
	}
	if quality.Unified.Passed {
		passedCount++
	}
	if quality.Secured.Passed {
		passedCount++
	}
	if quality.Trackable.Passed {
		passedCount++
	}

	qualityPercent := (passedCount * 100) / totalCount
	result = strings.ReplaceAll(result, "{quality}", fmt.Sprintf("%d%%", qualityPercent))

	// Overall status
	statusText := "Clean"
	if gitInfo != nil && gitInfo.HasChanges { // nolint: nilcheck // gitInfo was checked above
		statusText = "Modified"
	}
	result = strings.ReplaceAll(result, "{status}", statusText)

	return result, nil
}

// FormatDefault generates the default statusline format
func (f *Formatter) FormatDefault() (string, error) {
	return f.Format("")
}

// GetTestCoverage returns test coverage percentage (simplified)
func (f *Formatter) GetTestCoverage() string {
	// Try to run go test -cover
	cmd := exec.Command("go", "test", "-cover", "./...")
	cmd.Dir = f.projectDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "N/A"
	}

	// Parse coverage from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			// Extract percentage
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, "%") {
					return part
				}
			}
		}
	}

	return "N/A"
}
