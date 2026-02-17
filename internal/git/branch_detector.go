// Package git provides higher-level Git utilities built on top of core/git.
package git

import "fmt"

// labelPrefixMap maps GitHub issue labels to Git branch prefixes.
// Order matters: first match wins when scanning labels.
var labelPrefixMap = map[string]string{
	"bug":           "fix/",
	"feature":       "feat/",
	"enhancement":   "feat/",
	"documentation": "docs/",
	"docs":          "docs/",
}

// DetectBranchPrefix returns the Git branch prefix for the given issue labels.
// It scans labels in order and returns the prefix for the first matching label.
// Returns "feat/" if no label matches.
func DetectBranchPrefix(labels []string) string {
	for _, label := range labels {
		if prefix, ok := labelPrefixMap[label]; ok {
			return prefix
		}
	}
	return "feat/"
}

// FormatIssueBranch returns a full branch name for a GitHub issue.
// Format: {prefix}issue-{number} (e.g., "fix/issue-123", "feat/issue-456").
func FormatIssueBranch(labels []string, issueNumber int) string {
	prefix := DetectBranchPrefix(labels)
	return fmt.Sprintf("%sissue-%d", prefix, issueNumber)
}
