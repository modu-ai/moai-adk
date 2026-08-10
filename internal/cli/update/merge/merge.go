// Package merge contains the update-orchestration merge logic (gitignore
// pattern merge, per-file 3-way merge of user-customized files, and the
// risk-analysis summary builders) extracted from the root update command
// during M3d-B2 decomposition (SPEC-CLI-TUX-V3-003). Behavior is
// byte-identical to the pre-decomposition implementation; only the package
// location changed.
//
// The underlying 3-way merge engine and analysis types (FileAnalysis,
// MergeAnalysis, Engine) live in github.com/modu-ai/moai-adk/internal/merge
// and are imported under the alias "mrg" because this package is itself named
// "merge" (subpackage shadowing its dependency's package name).
package merge

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	mrg "github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
)

// FileBackup holds a file path and its backed-up content for merging.
type FileBackup struct {
	Path string
	Data []byte
}

// UserPatternsMarker delimits the user-owned tail of a merged .gitignore.
// Everything below it belongs to the user and is carried across a `moai update`
// verbatim — comments, blank lines, and original order included.
const UserPatternsMarker = "# User Custom Patterns (preserved by moai update)"

// MergeGitignoreFile reads the newly deployed .gitignore template and merges
// user-specific patterns from the backup. Template content is kept as-is.
//
// Two paths, chosen by whether the backup already carries UserPatternsMarker:
//
//   - marker present (the common case after the first update) — everything below
//     the marker is carried over VERBATIM. Only exact-match non-comment lines the
//     template already ships are dropped, so comments and ordering survive.
//   - marker absent (first-time upgrader) — fall back to line-set subtraction and
//     emit a fresh marker block.
//
// The verbatim path exists because the subtraction path regenerated the very
// block it advertised as preserved: it skipped comments and blank lines when
// collecting user additions, so user comments were lost and pattern order was
// destroyed. Order is semantic in gitignore — a `!negation` behaves differently
// depending on which rule precedes it (issues #1415 / #1377).
func MergeGitignoreFile(gitignorePath string, userBackup []byte) error {
	templateContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("read new .gitignore: %w", err)
	}

	// Build a set of non-blank, non-comment lines from the template
	templateLines := strings.Split(string(templateContent), "\n")
	templateSet := make(map[string]bool, len(templateLines))
	for _, line := range templateLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			templateSet[trimmed] = true
		}
	}

	if block, ok := extractUserMarkerBlock(string(userBackup)); ok {
		return writeUserMarkerBlock(gitignorePath, string(templateContent), block, templateSet)
	}

	// Collect user-specific lines that are not in the new template
	userLines := strings.Split(string(userBackup), "\n")
	var userAdditions []string
	for _, line := range userLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !templateSet[trimmed] {
			userAdditions = append(userAdditions, line)
		}
	}

	if len(userAdditions) == 0 {
		return nil // No user-specific patterns to preserve
	}

	// Append user additions to the template content
	result := string(templateContent)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	result += "\n" + UserPatternsMarker + "\n"
	for _, line := range userAdditions {
		result += line + "\n"
	}

	return os.WriteFile(gitignorePath, []byte(result), defs.FilePerm)
}

// extractUserMarkerBlock returns the lines below UserPatternsMarker in backup,
// with trailing blank lines trimmed, and reports whether the marker was found.
// Only the LAST marker occurrence is honoured, so a stray marker inside the
// user's own block cannot truncate it.
func extractUserMarkerBlock(backup string) ([]string, bool) {
	lines := strings.Split(backup, "\n")
	markerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == UserPatternsMarker {
			markerIdx = i
		}
	}
	if markerIdx < 0 {
		return nil, false
	}

	block := lines[markerIdx+1:]
	for len(block) > 0 && strings.TrimSpace(block[len(block)-1]) == "" {
		block = block[:len(block)-1]
	}
	return block, true
}

// writeUserMarkerBlock appends block to templateContent under a single
// UserPatternsMarker, preserving the block's comments, blank lines, and order.
//
// A block line is dropped only when it is a non-comment pattern the template
// already ships — re-emitting it would duplicate a template rule. Comments and
// blank lines are never dropped, so the structure around a removed duplicate
// stays intact. When nothing but blanks survives, the file is left untouched.
func writeUserMarkerBlock(gitignorePath, templateContent string, block []string, templateSet map[string]bool) error {
	kept := make([]string, 0, len(block))
	hasContent := false
	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && templateSet[trimmed] {
			continue // already shipped by the template
		}
		if trimmed != "" {
			hasContent = true
		}
		kept = append(kept, line)
	}
	// Trailing blanks can survive the filter above once a duplicate is removed.
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	if !hasContent || len(kept) == 0 {
		return nil // nothing to carry over — leave the deployed template as-is
	}

	result := templateContent
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	result += "\n" + UserPatternsMarker + "\n"
	result += strings.Join(kept, "\n") + "\n"

	return os.WriteFile(gitignorePath, []byte(result), defs.FilePerm)
}

// MergeUserFiles performs 3-way merge for user-customized files after template deployment.
// It uses the manifest's TemplateHash as the base, user's backed-up content as current,
// and the newly deployed template as updated. This preserves user customizations while
// incorporating template changes.
func MergeUserFiles(projectRoot string, backups []FileBackup, out io.Writer) error {
	// Load embedded templates to get original template content for base version
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Load manifest to get template hashes for base version
	mgr := manifest.NewManager()
	if _, loadErr := mgr.Load(projectRoot); loadErr != nil {
		return fmt.Errorf("load manifest: %w", loadErr)
	}

	// Create merge engine
	engine := mrg.NewEngine()

	var mergedCount int
	for _, fb := range backups {
		destPath := filepath.Join(projectRoot, fb.Path)

		// Read newly deployed file (updated version)
		updatedContent, err := os.ReadFile(destPath)
		if err != nil {
			// File might not exist in new template version - keep user's version
			if writeErr := os.WriteFile(destPath, fb.Data, defs.FilePerm); writeErr != nil {
				return fmt.Errorf("restore removed file %s: %w", fb.Path, writeErr)
			}
			_, _ = fmt.Fprintf(out, "  %s %s preserved (removed in new template)\n", uikit.SymSuccess(), fb.Path)
			mergedCount++
			continue
		}

		// Get original template content from embedded filesystem for base version
		// Try both with and without leading dot
		possiblePaths := []string{fb.Path, strings.TrimPrefix(fb.Path, ".")}
		var baseContent []byte
		for _, p := range possiblePaths {
			if data, readErr := fs.ReadFile(embedded, p); readErr == nil {
				baseContent = data
				break
			}
		}

		// Perform 3-way merge: base (original template), current (user's backup), updated (new template)
		// If base is not available, treat as new file - preserve user content
		if baseContent == nil {
			// No base available - this might be a user-created file
			// Prefer user's content but merge if compatible
			if string(fb.Data) == string(updatedContent) {
				continue // No change needed
			}
			// Keep user's version as-is
			if err := os.WriteFile(destPath, fb.Data, defs.FilePerm); err != nil {
				return fmt.Errorf("restore user file %s: %w", fb.Path, err)
			}
			_, _ = fmt.Fprintf(out, "  %s %s user content preserved\n", uikit.SymSuccess(), fb.Path)
			mergedCount++
			continue
		}

		// Use merge engine for proper 3-way merge
		result, mergeErr := engine.MergeFile(context.Background(), fb.Path, baseContent, fb.Data, updatedContent)
		if mergeErr != nil {
			// Merge failed - preserve user's version
			_, _ = fmt.Fprintf(out, "  %s %s merge failed, preserving user version: %v\n", uikit.SymWarning(), fb.Path, mergeErr)
			if err := os.WriteFile(destPath, fb.Data, defs.FilePerm); err != nil {
				return fmt.Errorf("preserve user file %s: %w", fb.Path, err)
			}
			mergedCount++
			continue
		}

		// Write merged result
		if err := os.WriteFile(destPath, result.Content, defs.FilePerm); err != nil {
			return fmt.Errorf("write merged file %s: %w", fb.Path, err)
		}

		// Report merge status
		if result.HasConflict {
			_, _ = fmt.Fprintf(out, "  %s %s merged with conflicts (user version preferred)\n", uikit.SymWarning(), fb.Path)
		} else {
			_, _ = fmt.Fprintf(out, "  %s %s user customizations preserved\n", uikit.SymSuccess(), fb.Path)
		}
		mergedCount++
	}

	if mergedCount > 0 {
		_, _ = fmt.Fprintf(out, "  %s Merged %d file(s) with 3-way merge engine\n", uikit.SymSuccess(), mergedCount)
	}

	return nil
}

// BuildMergeAnalysis creates a summary from individual file analysis results.
// It counts high/medium/low risk files, determines overall risk level,
// identifies conflicts, and generates a human-readable summary.
func BuildMergeAnalysis(files []mrg.FileAnalysis) mrg.MergeAnalysis {
	var highRisk, medRisk, lowRisk int
	for _, f := range files {
		switch f.RiskLevel {
		case "high":
			highRisk++
		case "medium":
			medRisk++
		case "low":
			lowRisk++
		}
	}

	overallRisk := "low"
	hasConflicts := false
	if highRisk > 0 {
		overallRisk = "high"
		hasConflicts = true
	} else if medRisk > 0 {
		overallRisk = "medium"
	}

	summary := fmt.Sprintf("Found %d files to sync", len(files))
	if highRisk > 0 {
		summary += fmt.Sprintf(" (%d high-risk files)", highRisk)
	}

	return mrg.MergeAnalysis{
		Files:        files,
		HasConflicts: hasConflicts,
		SafeToMerge:  highRisk == 0,
		Summary:      summary,
		RiskLevel:    overallRisk,
	}
}

// AnalyzeMergeChanges performs a quick analysis of template files that will be modified.
// It evaluates risk levels based on file types and existing content:
//   - High risk: CLAUDE.md, settings.json (core config files)
//   - Medium risk: Existing files being updated
//   - Low risk: New files being created
//
// Returns a MergeAnalysis with file-by-file risk assessment and recommended strategies.
func AnalyzeMergeChanges(deployer template.Deployer, projectRoot string) mrg.MergeAnalysis {
	templates := deployer.ListTemplates()
	files := plan.AnalyzeFiles(templates, projectRoot)
	return BuildMergeAnalysis(files)
}
