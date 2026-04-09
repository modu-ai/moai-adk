package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResultStore manages per-target experiment results and change history on the filesystem.
type ResultStore struct {
	baseDir string
}

// NewResultStore creates a store that manages experiment results in the specified directory.
func NewResultStore(baseDir string) *ResultStore {
	return &ResultStore{baseDir: baseDir}
}

// SaveExperiment saves an experiment result as a JSON file.
// Filenames use the format exp-{NNN}.json with zero-padded 3-digit sequence numbers.
func (s *ResultStore) SaveExperiment(target string, exp *Experiment) error {
	targetDir := s.targetDir(target)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create experiment directory (target=%s): %w", target, err)
	}

	count := s.ExperimentCount(target)
	filename := fmt.Sprintf("exp-%03d.json", count+1)

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize experiment (id=%s): %w", exp.ID, err)
	}

	filePath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write experiment file (%s): %w", filePath, err)
	}

	return nil
}

// LoadExperiments loads all experiment results for a target in filename order.
// Returns an empty slice if the directory does not exist.
func (s *ResultStore) LoadExperiments(target string) ([]*Experiment, error) {
	targetDir := s.targetDir(target)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Experiment{}, nil
		}
		return nil, fmt.Errorf("failed to read experiment directory (target=%s): %w", target, err)
	}

	// Filter and sort exp-*.json files only
	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "exp-") && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e.Name())
		}
	}
	sort.Strings(jsonFiles)

	experiments := make([]*Experiment, 0, len(jsonFiles))
	for _, name := range jsonFiles {
		data, err := os.ReadFile(filepath.Join(targetDir, name))
		if err != nil {
			return nil, fmt.Errorf("failed to read experiment file (%s): %w", name, err)
		}

		var exp Experiment
		if err := json.Unmarshal(data, &exp); err != nil {
			return nil, fmt.Errorf("failed to deserialize experiment (%s): %w", name, err)
		}
		experiments = append(experiments, &exp)
	}

	return experiments, nil
}

// AppendChangelog appends a markdown entry to the target's changelog.md.
func (s *ResultStore) AppendChangelog(target string, entry ChangelogEntry) error {
	targetDir := s.targetDir(target)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create changelog directory: %w", err)
	}

	changelogPath := filepath.Join(targetDir, "changelog.md")

	// Build markdown entry
	md := fmt.Sprintf(
		"### %s (score: %.2f, decision: %s)\n\n- **Change**: %s\n- **Reason**: %s\n\n",
		entry.ExperimentID,
		entry.Score,
		entry.Decision,
		entry.Change,
		entry.Reasoning,
	)

	f, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open changelog file: %w", err)
	}

	if _, err := f.WriteString(md); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return f.Close()
}

// ExperimentCount returns the number of exp-*.json files in the target directory.
func (s *ResultStore) ExperimentCount(target string) int {
	targetDir := s.targetDir(target)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "exp-") && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}

	return count
}

// targetDir returns the directory path for the given target.
func (s *ResultStore) targetDir(target string) string {
	return filepath.Join(s.baseDir, sanitizeTarget(target))
}
