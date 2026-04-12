package evolution

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/merge"
)

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
	original, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("evolution: read target file %s: %w", fullPath, err)
	}

	// Write backup.
	backupPath := fullPath + ".bak"
	if err := os.WriteFile(backupPath, original, 0o644); err != nil {
		return fmt.Errorf("evolution: write backup %s: %w", backupPath, err)
	}

	// Find the existing zone content and append to it.
	zones, _ := merge.ParseEvolvableZones(string(original))
	var existingContent string
	for _, z := range zones {
		if z.ID == proposal.ZoneID {
			existingContent = z.Content
			break
		}
	}
	if existingContent == "" && len(zones) > 0 {
		// Zone exists but is empty — that's fine, we'll just add the new content.
		for _, z := range zones {
			if z.ID == proposal.ZoneID {
				existingContent = ""
				break
			}
		}
	}

	// Check that the zone actually exists.
	found := false
	for _, z := range zones {
		if z.ID == proposal.ZoneID {
			found = true
			break
		}
	}
	if !found {
		return ErrZoneNotFound
	}

	// Build new zone content: existing + separator + addition.
	newContent := existingContent
	if newContent != "" && len(newContent) > 0 && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}
	newContent += proposal.Addition

	updated, err := merge.MergeEvolvableZones(string(original), proposal.ZoneID, newContent)
	if err != nil {
		return fmt.Errorf("evolution: merge evolvable zones: %w", err)
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
