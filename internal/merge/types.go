// Package merge provides a 3-way merge engine for template files.
//
// It implements ADR-008 (3-Way Merge for Template Updates) by comparing
// base (original template), current (user's file), and updated (new template)
// versions to produce a merged result that preserves user customizations.
package merge

import (
	"context"
	"errors"
)

// MergeStrategy identifies which merge algorithm to apply.
type MergeStrategy string

const (
	// LineMerge performs line-by-line diff merge. Default for .md, .txt files.
	LineMerge MergeStrategy = "line_merge"

	// YAMLDeep performs structure-preserving deep merge for YAML files.
	YAMLDeep MergeStrategy = "yaml_deep"

	// JSONMerge performs object-level merge for JSON files.
	JSONMerge MergeStrategy = "json_merge"

	// SectionMerge performs heading-based section merge for CLAUDE.md.
	SectionMerge MergeStrategy = "section_merge"

	// EntryMerge performs entry-based union merge for .gitignore-style files.
	EntryMerge MergeStrategy = "entry_merge"

	// Overwrite replaces the file entirely, backing up the original.
	Overwrite MergeStrategy = "overwrite"

	// EvolvableZoneMerge applies "user wins" strategy inside evolvable zone markers
	// and normal line-based merge outside them. Used for .md files that contain
	// <!-- moai:evolvable-start id="..." --> markers.
	EvolvableZoneMerge MergeStrategy = "evolvable_zone_merge"
)

// MergeResult holds the outcome of a merge operation.
type MergeResult struct {
	// Content is the merged file content.
	Content []byte

	// HasConflict indicates whether unresolvable conflicts were detected.
	HasConflict bool

	// Conflicts lists all conflict regions found during the merge.
	Conflicts []Conflict

	// Strategy is the merge algorithm that was applied.
	Strategy MergeStrategy
}

// Conflict describes a single conflict region within a merge result.
type Conflict struct {
	// StartLine is the 1-based line number where the conflict begins.
	StartLine int

	// EndLine is the 1-based line number where the conflict ends.
	EndLine int

	// Base is the original template content in the conflict region.
	Base string

	// Current is the user's content in the conflict region.
	Current string

	// Updated is the new template content in the conflict region.
	Updated string
}

// MergeAnalysis holds analysis results for multiple files to be merged.
//
// It is consumed outside this package by internal/cli/update.go
// (toPreviewInputs) and internal/cli/update/merge (AnalyzeMergeChanges).
type MergeAnalysis struct {
	Files        []FileAnalysis
	HasConflicts bool
	SafeToMerge  bool
	Summary      string
	RiskLevel    string
}

// FileAnalysis contains merge analysis for a single file.
//
// The field set is load-bearing: internal/cli/update/plan constructs this type
// with a named-field composite literal, so renaming or removing a field breaks
// that call site at compile time.
type FileAnalysis struct {
	Path      string
	Changes   string
	Strategy  MergeStrategy
	RiskLevel string // "low", "medium", "high"
	Note      string
}

// Engine defines the 3-way merge operations.
type Engine interface {
	// ThreeWayMerge performs a generic line-based 3-way merge on byte slices.
	ThreeWayMerge(base, current, updated []byte) (*MergeResult, error)

	// MergeFile performs a strategy-aware 3-way merge, selecting the strategy
	// based on the file path and extension.
	MergeFile(ctx context.Context, path string, base, current, updated []byte) (*MergeResult, error)
}

// StrategySelector chooses the appropriate merge strategy for a file.
type StrategySelector interface {
	// SelectStrategy returns the merge strategy for the given file path.
	SelectStrategy(path string) MergeStrategy
}

// Sentinel errors for the merge package.
var (
	// ErrMergeConflict indicates an unresolvable conflict was detected.
	ErrMergeConflict = errors.New("merge: unresolvable conflict detected")

	// ErrMergeUnsupported indicates the file type is not supported for merge.
	ErrMergeUnsupported = errors.New("merge: file type not supported for merge")
)
