package evolution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/merge"
)

// @MX:NOTE: [AUTO] ApplyProposal appends to evolvable zones with backup (.bak) and atomic rename (single-writer pattern)
//
// ApplyProposal reads the skill file identified by proposal.TargetFile,
// locates the evolvable zone proposal.ZoneID, and appends proposal.Addition
// to the zone content.
//
// A backup of the original file is written to <TargetFile>.bak before any
// modification so that RevertProposal can restore it.
//
// IMPORTANT: This function is provided for use after explicit user approval.
// It MUST NOT be called autonomously.
func ApplyProposal(projectRoot string, proposal *ProposedChange) error {
	if proposal == nil {
		return fmt.Errorf("evolution: nil ProposedChange")
	}
	if err := CheckFrozenGuard(proposal.TargetFile); err != nil {
		return err
	}

	fullPath := filepath.Join(projectRoot, proposal.TargetFile)

	// Validate that the path does not escape projectRoot: check that the normalized path
	// is still under projectRoot.
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("evolution: resolve project root: %w", err)
	}
	absTarget, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("evolution: resolve target path: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("evolution: target file %q escapes project root", proposal.TargetFile)
	}
	original, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("evolution: read target file %s: %w", fullPath, err)
	}

	// Write backup.
	backupPath := fullPath + ".bak"
	if err := os.WriteFile(backupPath, original, 0o644); err != nil {
		return fmt.Errorf("evolution: write backup %s: %w", backupPath, err)
	}

	// Locate the existing zone content and append the new content.
	// Read the current zone content with ParseEvolvableZones, append proposal.Addition,
	// then replace only the zone while preserving the overall file structure (header/footer)
	// using ReplaceEvolvableZone.
	zones, parseErr := merge.ParseEvolvableZones(string(original))
	if parseErr != nil {
		return fmt.Errorf("evolution: parse evolvable zones: %w", parseErr)
	}

	// Verify the target zone exists and extract its existing content.
	var existingContent string
	found := false
	for _, z := range zones {
		if z.ID == proposal.ZoneID {
			existingContent = z.Content
			found = true
			break
		}
	}
	if !found {
		return ErrZoneNotFound
	}

	// New zone content: existing content + separator + appended content.
	newContent := existingContent
	if newContent != "" && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}
	newContent += proposal.Addition

	// ReplaceEvolvableZone replaces only the specific zone while preserving the file's header/footer.
	// (Fixes the previous misuse of MergeEvolvableZones(original, zoneID, newContent))
	updated, err := merge.ReplaceEvolvableZone(string(original), proposal.ZoneID, newContent)
	if err != nil {
		return fmt.Errorf("evolution: replace evolvable zone: %w", err)
	}

	// Write atomically.
	tmp := fullPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("evolution: write updated file: %w", err)
	}
	if err := os.Rename(tmp, fullPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evolution: rename updated file: %w", err)
	}
	return nil
}

// RevertProposal restores the original file from the .bak backup created by
// ApplyProposal.  The backup is deleted after a successful restore.
func RevertProposal(projectRoot string, proposal *ProposedChange) error {
	if proposal == nil {
		return fmt.Errorf("evolution: nil ProposedChange for revert")
	}

	fullPath := filepath.Join(projectRoot, proposal.TargetFile)
	backupPath := fullPath + ".bak"

	backup, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("evolution: read backup %s: %w", backupPath, err)
	}

	if err := os.WriteFile(fullPath, backup, 0o644); err != nil {
		return fmt.Errorf("evolution: restore from backup: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}
