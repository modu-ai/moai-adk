package lifecycle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PersistentMode holds the persistent mode state for a project.
// When active, the stop hook blocks Claude from stopping so that
// the workflow can continue until a completion marker is detected
// or the max duration expires.
type PersistentMode struct {
	Active             bool      `json:"active"`
	Workflow           string    `json:"workflow"`
	SpecID             string    `json:"spec_id"`
	StartedAt          time.Time `json:"started_at"`
	MaxDurationMinutes int       `json:"max_duration_minutes"`
}

// persistentModeFile is the path (relative to projectDir) where persistent mode state is stored.
const persistentModeFile = ".moai/state/persistent-mode.json"

// ActivatePersistentMode writes a persistent-mode.json file to mark the mode as active.
// projectDir is the root of the project (CWD from the hook input).
func ActivatePersistentMode(projectDir, workflow, specID string, maxMinutes int) error {
	mode := PersistentMode{
		Active:             true,
		Workflow:           workflow,
		SpecID:             specID,
		StartedAt:          time.Now(),
		MaxDurationMinutes: maxMinutes,
	}
	filePath := filepath.Join(projectDir, persistentModeFile)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	data, err := json.MarshalIndent(mode, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// DeactivatePersistentMode sets active=false in the persistent-mode.json file.
// If the file does not exist or cannot be read, the function returns nil (no-op).
func DeactivatePersistentMode(projectDir string) error {
	mode, err := CheckPersistentMode(projectDir)
	if err != nil || mode == nil {
		return nil
	}
	mode.Active = false
	filePath := filepath.Join(projectDir, persistentModeFile)
	data, err := json.MarshalIndent(mode, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// CheckPersistentMode reads the persistent-mode.json file and returns the current state.
// Returns (nil, nil) if the file does not exist.
func CheckPersistentMode(projectDir string) (*PersistentMode, error) {
	filePath := filepath.Join(projectDir, persistentModeFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read: %w", err)
	}
	var mode PersistentMode
	if err := json.Unmarshal(data, &mode); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &mode, nil
}

// IsExpired reports whether the persistent mode session has exceeded its maximum duration.
// Returns false when MaxDurationMinutes is 0 (no limit).
func (m *PersistentMode) IsExpired() bool {
	if m.MaxDurationMinutes <= 0 {
		return false
	}
	return time.Since(m.StartedAt) > time.Duration(m.MaxDurationMinutes)*time.Minute
}
