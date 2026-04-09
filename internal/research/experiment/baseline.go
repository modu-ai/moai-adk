package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// BaselineManager saves and loads per-target baseline evaluation results on the filesystem.
type BaselineManager struct {
	storeDir string
}

// NewBaselineManager creates a manager that manages baselines in the specified directory.
func NewBaselineManager(storeDir string) *BaselineManager {
	return &BaselineManager{storeDir: storeDir}
}

// Save persists the baseline evaluation result for a target as a JSON file.
func (m *BaselineManager) Save(target string, result *eval.EvalResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize baseline (target=%s): %w", target, err)
	}

	filePath := m.filePath(target)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create baseline directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write baseline file (target=%s): %w", target, err)
	}

	return nil
}

// Load reads the baseline evaluation result for a target from file.
func (m *BaselineManager) Load(target string) (*eval.EvalResult, error) {
	data, err := os.ReadFile(m.filePath(target))
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file (target=%s): %w", target, err)
	}

	var result eval.EvalResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to deserialize baseline (target=%s): %w", target, err)
	}

	return &result, nil
}

// Exists checks whether a baseline file exists for the given target.
func (m *BaselineManager) Exists(target string) bool {
	_, err := os.Stat(m.filePath(target))
	return err == nil
}

// filePath returns the baseline file path for the given target.
func (m *BaselineManager) filePath(target string) string {
	return filepath.Join(m.storeDir, sanitizeTarget(target)+".baseline.json")
}

// sanitizeTarget converts a target string to a filesystem-safe form.
// Replaces '/' with '_' and removes '.'.
func sanitizeTarget(target string) string {
	s := strings.ReplaceAll(target, "/", "_")
	s = strings.ReplaceAll(s, ".", "")
	return s
}
