// Package astgrep provides AST-based code analysis using the ast-grep (sg) CLI tool.
// It wraps the sg CLI to perform structural code search, pattern matching,
// rule-based scanning, and code transformation.
package astgrep

import "time"

// Match represents a single AST pattern match result from sg CLI.
type Match struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine  int    `json:"endLine,omitempty"`
	EndCol   int    `json:"endColumn,omitempty"`
	Text     string `json:"text"`
	Rule     string `json:"rule"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ScanResult represents the result of scanning files for AST patterns.
type ScanResult struct {
	Matches  []Match       `json:"matches"`
	Duration time.Duration `json:"duration"`
	Files    int           `json:"files_scanned"`
	Language string        `json:"language,omitempty"`
}

// ScanConfig holds configuration for a scan operation.
type ScanConfig struct {
	RulesPath       string   `json:"rules_path,omitempty"`
	SecurityScan    bool     `json:"security_scan"`
	IncludePatterns []string `json:"include_patterns,omitempty"`
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
}

// DefaultScanConfig returns a ScanConfig with sensible defaults.
// Default exclusions: node_modules, .git, __pycache__.
func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		ExcludePatterns: []string{"node_modules", ".git", "__pycache__"},
	}
}

// FileChange represents a single code transformation applied to a file.
type FileChange struct {
	FilePath string `json:"file_path"`
	OldCode  string `json:"old_code"`
	NewCode  string `json:"new_code"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// ProjectScanResult represents the aggregated result of scanning a project.
type ProjectScanResult struct {
	TotalFiles   int            `json:"total_files"`
	TotalMatches int            `json:"total_matches"`
	Duration     time.Duration  `json:"duration"`
	BySeverity   map[string]int `json:"by_severity"`
	FileResults  []ScanResult   `json:"file_results,omitempty"`
}

// ReplaceResult represents the result of a pattern replacement operation.
type ReplaceResult struct {
	MatchesFound  int          `json:"matches_found"`
	FilesModified int          `json:"files_modified"`
	Changes       []FileChange `json:"changes"`
	DryRun        bool         `json:"dry_run"`
}

// Rule represents an ast-grep rule loaded from YAML.
type Rule struct {
	ID       string `json:"id" yaml:"id"`
	Language string `json:"language" yaml:"language"`
	Severity string `json:"severity" yaml:"severity"`
	Message  string `json:"message" yaml:"message"`
	Pattern  string `json:"pattern" yaml:"pattern"`
	Fix      string `json:"fix,omitempty" yaml:"fix,omitempty"`
}
