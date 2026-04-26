package runtime

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PersistProgress writes an auto-save section to .moai/specs/<specID>/progress.md
// and returns a paste-ready resume message.
//
// If the SPEC directory does not exist, the function silently skips (returns "", nil).
// Existing progress.md content is preserved; the new section is appended.
// Atomic write is used via os.Rename to avoid partial writes (SPEC-V3R3-ARCH-007 §2.4).
func (t *Tracker) PersistProgress(specID, waveLabel, approach, nextStep string) (string, error) {
	t.mu.RLock()
	cfg := t.config
	projectRoot := t.projectRoot
	t.mu.RUnlock()

	specDir := filepath.Join(projectRoot, ".moai", "specs", specID)
	if _, err := os.Stat(specDir); os.IsNotExist(err) {
		// Silent skip per REQ-ARCH007-005 constraint C6
		slog.Debug("PersistProgress: SPEC directory not found, skipping",
			"spec_id", specID,
			"path", specDir,
		)
		return "", nil
	}

	progressPath := filepath.Join(specDir, "progress.md")

	// Build the appended section
	timestamp := time.Now().UTC().Format(time.RFC3339)
	section := fmt.Sprintf("\n## Auto-saved at %s (75%% threshold)\n\n- Wave: %s\n- Approach: %s\n- Next step: %s\n",
		timestamp, waveLabel, approach, nextStep)

	// Read existing content (if any)
	var existing []byte
	if data, err := os.ReadFile(progressPath); err == nil {
		existing = data
	}

	// Build new content = existing + new section
	newContent := append(existing, []byte(section)...)

	// Atomic write via temp file + rename
	if err := atomicWrite(progressPath, newContent); err != nil {
		return "", fmt.Errorf("PersistProgress: atomic write failed: %w", err)
	}

	slog.Info("PersistProgress: progress.md saved", "spec_id", specID, "path", progressPath)

	// Generate resume message
	resumeMsg := buildResumeMessage(cfg.ResumeMessageFormat, specID, waveLabel, approach, progressPath, nextStep)
	return resumeMsg, nil
}

// atomicWrite writes data to path atomically using a temp file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".progress_tmp_*")
	if err != nil {
		return fmt.Errorf("atomicWrite: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: rename: %w", err)
	}
	return nil
}

// buildResumeMessage generates the paste-ready resume message from the format template.
// Replacements follow context-window-management.md §Resume message format.
func buildResumeMessage(format, specID, waveLabel, approach, progressPath, nextStep string) string {
	msg := format
	msg = strings.ReplaceAll(msg, "{spec_id}", specID)
	msg = strings.ReplaceAll(msg, "{SPEC_ID}", specID)
	msg = strings.ReplaceAll(msg, "{wave_label}", waveLabel)
	msg = strings.ReplaceAll(msg, "{approach_summary}", approach)
	msg = strings.ReplaceAll(msg, "{progress_path}", progressPath)
	msg = strings.ReplaceAll(msg, "{next_step}", nextStep)
	return msg
}
