package observe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// observationsFile is the JSONL filename used to store observation data.
const observationsFile = "observations.jsonl"

// Storage manages a file-based store for observation records.
type Storage struct {
	baseDir string
}

// NewStorage creates an observation store in the specified directory.
func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

// filePath returns the full path to the observations file.
func (s *Storage) filePath() string {
	return filepath.Join(s.baseDir, observationsFile)
}

// Append adds a single observation to the JSONL file.
// Creates the file if it does not exist.
func (s *Storage) Append(obs *Observation) error {
	f, err := os.OpenFile(s.filePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("observe: failed to open file: %w", err)
	}

	data, err := json.Marshal(obs)
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("observe: failed to serialize JSON: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("observe: failed to write file: %w", err)
	}

	return f.Close()
}

// LoadAll returns all stored observations in order.
// Returns an empty slice if the file does not exist or is empty.
// Corrupted lines are skipped.
func (s *Storage) LoadAll() ([]*Observation, error) {
	f, err := os.Open(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("observe: failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var result []*Observation
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var obs Observation
		if err := json.Unmarshal([]byte(line), &obs); err != nil {
			// Skip corrupted lines
			continue
		}
		result = append(result, &obs)
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("observe: failed to scan file: %w", err)
	}

	return result, nil
}

// LoadSince returns only observations after the specified time.
// Includes observations with a Timestamp equal to or after since.
func (s *Storage) LoadSince(since time.Time) ([]*Observation, error) {
	all, err := s.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("observe: LoadSince failed: %w", err)
	}

	var filtered []*Observation
	for _, obs := range all {
		if !obs.Timestamp.Before(since) {
			filtered = append(filtered, obs)
		}
	}

	return filtered, nil
}
