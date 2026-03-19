package dashboard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Loader reads task metrics from JSONL files.
type Loader struct {
	metricsDir string
}

// NewLoader creates a new metrics loader.
func NewLoader(metricsDir string) *Loader {
	return &Loader{metricsDir: metricsDir}
}

// LoadMetrics reads all task metrics from the metrics directory.
func (l *Loader) LoadMetrics() ([]TaskMetric, error) {
	pattern := filepath.Join(l.metricsDir, "task-metrics*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob metrics files: %w", err)
	}

	if len(files) == 0 {
		// Try single file
		singleFile := filepath.Join(l.metricsDir, "task-metrics.jsonl")
		if _, err := os.Stat(singleFile); err == nil {
			files = []string{singleFile}
		}
	}

	var metrics []TaskMetric
	for _, f := range files {
		m, err := l.loadFile(f)
		if err != nil {
			slog.Warn("failed to load metrics file", "path", f, "error", err)
			continue
		}
		metrics = append(metrics, m...)
	}
	return metrics, nil
}

func (l *Loader) loadFile(path string) ([]TaskMetric, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var metrics []TaskMetric
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var m TaskMetric
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			continue
		}
		metrics = append(metrics, m)
	}
	return metrics, scanner.Err()
}
