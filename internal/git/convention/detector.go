package convention

import (
	"fmt"
	"os/exec"
	"strings"
)

// Detect analyzes recent commits in the repository and returns the best
// matching built-in convention. sampleSize controls how many recent commits
// to analyze. repoPath is the git repository root.
func Detect(repoPath string, sampleSize int) (*DetectionResult, error) {
	if sampleSize <= 0 {
		sampleSize = 100
	}

	messages, err := getRecentCommitMessages(repoPath, sampleSize)
	if err != nil {
		return nil, fmt.Errorf("detect convention: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("detect convention: no commits found")
	}

	var bestResult *DetectionResult

	for _, name := range BuiltinNames() {
		conv, err := ParseBuiltin(name)
		if err != nil {
			continue
		}

		matchCount := 0
		for _, msg := range messages {
			r := Validate(msg, conv)
			if r.Valid {
				matchCount++
			}
		}

		confidence := float64(matchCount) / float64(len(messages))

		if bestResult == nil || confidence > bestResult.Confidence {
			bestResult = &DetectionResult{
				Convention: conv,
				Confidence: confidence,
				SampleSize: len(messages),
				MatchCount: matchCount,
			}
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("detect convention: no matching convention found")
	}

	return bestResult, nil
}

// Score calculates how well a set of messages matches a convention (0.0-1.0).
func Score(messages []string, conv *Convention) float64 {
	if len(messages) == 0 || conv == nil {
		return 0
	}

	matchCount := 0
	for _, msg := range messages {
		r := Validate(msg, conv)
		if r.Valid {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(messages))
}

// getRecentCommitMessages retrieves recent commit messages from git log.
func getRecentCommitMessages(repoPath string, limit int) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", fmt.Sprintf("--max-count=%d", limit), "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	// Filter empty lines.
	var messages []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			messages = append(messages, line)
		}
	}
	return messages, nil
}
